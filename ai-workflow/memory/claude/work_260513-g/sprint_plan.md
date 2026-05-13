# Sprint Plan — claude/work_260513-g (B1 auth 도메인 확장)

- 문서 목적: B 묶음 RBAC 1차 (PR #92) 다음의 auth 도메인 본문 ID 노출 + IMPL-auth-XX 정밀 매핑.
- 범위: docs only. `backend_api_contract.md` §11 + `docs/traceability/report.md` 갱신. 코드 변경 0.
- 진입 base: main HEAD `a73dba1` (PR #92 직후).
- 최종 수정일: 2026-05-13
- 상태: in_progress

## 1. ID 매핑 (본 sprint 결정)

### API-XX (auth 도메인)

| API ID | 본문 위치 | 항목 |
| --- | --- | --- |
| `API-19` | §11.3 | Go Core Bearer token 경계 (`Authorization: Bearer ...` verifier interface + actor context) |
| `API-20` | §11.5 | `POST /api/v1/auth/login` (login_challenge → Kratos api-mode login + Hydra accept) |
| `API-21` | §11.5 | `POST /api/v1/auth/logout` (logout_challenge → Hydra revoke + accept) |
| `API-22` | §11.5 | `POST /api/v1/auth/token` (authorization_code → Hydra /oauth2/token) |
| `API-23` | §11.5 | `POST /api/v1/auth/signup` (HRDB lookup + Kratos identity 생성) |
| `API-24` | §11.5 | `GET /api/v1/auth/consent` (Hydra consent flow auto-accept) |
| `API-35` | §11.5.1 | `POST /api/v1/account/password` (self-service password change) |

§11.2 Hydra 표준 endpoint 는 외부 의존성 (conventions.md §5.2) — DevHub 재정의 0, API-XX 미부여.
§11.4 admin identity wrapper 는 planned (M3), 진입 시 발급.

### IMPL-auth-XX

| IMPL ID | 코드 위치 | 책임 |
| --- | --- | --- |
| `IMPL-auth-01` | `backend-core/internal/auth/hydra_introspection.go` + `internal/httpapi/auth.go::BearerTokenVerifier` | Bearer token verifier interface + Hydra introspection 구현 (§11.3) |
| `IMPL-auth-02` | `internal/httpapi/auth.go::authenticateActor` middleware + `AuthenticatedActor` context propagation | Bearer 검증 결과를 request context 의 actor 로 (§11.3) |
| `IMPL-auth-03` | `internal/httpapi/auth_login.go` | `POST /api/v1/auth/login` handler (API-20) |
| `IMPL-auth-04` | `internal/httpapi/auth_logout.go` | `POST /api/v1/auth/logout` handler (API-21) |
| `IMPL-auth-05` | `internal/httpapi/auth_token.go` | `POST /api/v1/auth/token` handler (API-22) |
| `IMPL-auth-06` | `internal/httpapi/auth_signup.go` | `POST /api/v1/auth/signup` handler (API-23) |
| `IMPL-auth-07` | `internal/httpapi/auth_consent.go` | `GET /api/v1/auth/consent` handler (API-24) |

API-35 `POST /api/v1/account/password` 의 IMPL 은 별도 도메인 (account) — 본 sprint 스코프 밖이지만 매트릭스 §3 의 cross-cut (인증 + 계정 관리 두 도메인 행) 으로 ID 노출.

## 2. 작업 항목

| 항목 | 위치 | 결과 |
| --- | --- | --- |
| §11.3 헤더 (API-19) | `backend_api_contract.md` §11.3 | 본문 ID 노출 |
| §11.5 endpoint 표에 API ID 컬럼 추가 | `backend_api_contract.md` §11.5 | 6 endpoint (API-20..24 + API-35 별도) |
| §11.5.1 헤더 (API-35) | `backend_api_contract.md` §11.5.1 | 본문 ID 노출 |
| §2.2 끝에 auth API 매핑 서브 표 | `docs/traceability/report.md` §2.2 | 패턴 정착 — RBAC 와 동일 |
| §2.4 끝에 IMPL-auth-XX 책임 정의 서브 표 | `docs/traceability/report.md` §2.4 | 7 IMPL |
| §3 인증 행 정리 | `docs/traceability/report.md` §3 | ID 범위 + §2 서브 표 참조 패턴 |
| §3 회원가입 행 — API-23 참조 | `docs/traceability/report.md` §3 | 동일 패턴 |
| §3 계정 관리 행 — API-35 cross-cut | `docs/traceability/report.md` §3 | 동일 패턴 |
| §6 변경 이력 + main flat sync (PR #92 흡수) | `docs/traceability/report.md` §6 + `ai-workflow/memory/*` | |

## 3. 검증

- 코드 변경 0 — backend/frontend test 회귀 0.
- `go test ./internal/httpapi/...` sanity 만.

## 4. 미진입 / 다음 sprint 후보

본 sprint 의 직속 후속:
- B1 다른 도메인 (account / org / command / audit / infra) 본문 ID 노출 — 점진
- B2 deprecated 문서 식별 + 마킹
- B4 X-Devhub-Actor 폐기 ADR
- C1 frontend 컴포넌트 Vitest
- C2 E2E 신규 TC
- D5 actionlint
- M3 / M4 진입
