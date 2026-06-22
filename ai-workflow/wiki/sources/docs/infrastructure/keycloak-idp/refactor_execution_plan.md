---
title: refactor_execution_plan
type: source
tags: [infrastructure, refactor_execution_plan.md, project-devhub]
sources: [raw/projects/devhub/docs/infrastructure/keycloak-idp/refactor_execution_plan.md]
git_commit: baf1cf24
git_branch: main
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T03:37:45Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# Keycloak 단일화 리팩토링 실행 계획

- 문서 목적: DevHub 인증 체계를 Ory Hydra+Kratos 결합 구조에서 Keycloak 단일 체계로 전환하는 구현 계획을 정의한다.
- 범위: 아키텍처 전환, Keycloak 서버 구성(내장/외부), 단계별 구현/검증/롤백 전략
- 대상 독자: backend-core, frontend, 인프라/운영, QA, 릴리즈 담당자
- 상태: accepted
- 최종 수정일: 2026-05-20
- 결정 근거 sprint: `codex/keycloak-only-refactor-plan` (1차 작성, 2026-05-18) → `claude/work_260519-a` (done status 갱신, 2026-05-19)
- 관련 문서: [ADR-0019 Keycloak 단일화 결정](../adr/0019-keycloak-only-idp.md), [ADR-0001 IdP selection (Hydra+Kratos, superseded)](../adr/0001-idp-selection.md), `docs/backend_api_contract.md`, `docs/tests/e2e-test-guide.md`, `docs/traceability/sync-checklist.md`

> **✅ 실행 완료 (2026-05-19)**: 본 plan 의 §7 PR 분할 (PR-A..F) 은 PR #167 단일 묶음으로 머지 완료. 아래 §3 "현재 상태 요약" 은 plan 작성 시점 (2026-05-18) 의 historical snapshot 이며, 실 코드 상태는 PR #167 머지 결과를 따른다. ADR governance 측면의 결정 명문화는 ADR-0019 가 책임.

## 1. 목표

- 인증/세션/계정 관리 IdP를 Keycloak으로 단일화한다.
- 로컬 개발은 내장 Keycloak 서비스로 즉시 실행 가능해야 한다.
- 운영/스테이징은 외부 Keycloak(Managed/별도 클러스터)을 환경설정만으로 연결 가능해야 한다.
- 전환 기간 동안 점진적 마이그레이션과 롤백 경로를 유지한다.

## 2. 비목표

- 사용자/조직 도메인(`users`, `org_units`) 자체 모델 재설계
- DREQ/Integration 등 인증 외 비즈니스 도메인 변경
- Keycloak 테마/브랜딩 커스터마이징

## 3. 현재 상태 요약 (main 기준)

- 백엔드는 `HydraIntrospectionVerifier`, Keycloak/OIDC client wiring에 결합.
- 프론트는 OIDC code+PKCE를 사용하나 logout/password/account flow가 Keycloak/OIDC endpoint 전제.
- 문서/테스트/환경설정이 Keycloak/OIDC를 표준으로 가정.

## 4. 목표 아키텍처

### 4.1 Backend

- `BearerTokenVerifier` 인터페이스는 유지.
- 구현체를 `KeycloakJWKSVerifier`로 교체.
- `/api/v1/auth/*` 중 Hydra challenge 기반 endpoint는 폐기 또는 호환 모드로 축소.
- 계정/비밀번호/admin identity 연산은 Keycloak Admin API adapter로 이동.
- `users.kratos_identity_id` 의존 로직은 `users.idp_subject`(중립 컬럼)으로 일반화.

### 4.2 Frontend

- OIDC discovery 기반 authorize/token/logout endpoint 사용.
- 기존 Hydra 전용 URL 생성 로직 제거.
- callback/login/logout/account 플로우를 Keycloak 기준으로 재작성.

### 4.3 IdP 구성

- Local: docker-compose의 `keycloak` 서비스 + realm import 자동화
- External: `KEYCLOAK_*`/`OIDC_*` env만으로 연결 (내장 서비스 비활성)

## 5. 환경설정 계약

### 5.1 공통

- `DEVHUB_IDP_PROVIDER=keycloak`
- `DEVHUB_OIDC_ISSUER_URL`
- `DEVHUB_OIDC_CLIENT_ID`
- `DEVHUB_OIDC_CLIENT_SECRET` (confidential client)
- `DEVHUB_OIDC_AUDIENCE` (선택)

### 5.2 Backend

- `DEVHUB_OIDC_JWKS_URL` (선택: 없으면 issuer discovery)
- `DEVHUB_KEYCLOAK_ADMIN_URL`
- `DEVHUB_KEYCLOAK_ADMIN_REALM`
- `DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID`
- `DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET`

### 5.3 Frontend

- `NEXT_PUBLIC_OIDC_ISSUER_URL`
- `NEXT_PUBLIC_OIDC_CLIENT_ID`
- `NEXT_PUBLIC_OIDC_REDIRECT_URI`
- `NEXT_PUBLIC_OIDC_SCOPE` (기본: `openid profile email`)

## 6. Keycloak 서버 구성 계획

### 6.1 Local embedded 모드

- `docker-compose`에 `keycloak` 추가
- realm: `devhub`
- clients:
  - `devhub-frontend` (public, PKCE required)
  - `devhub-backend` (confidential, service account enabled)
- roles: `developer`, `manager`, `team_manager`, `system_admin`
- mapper: `preferred_username`, `email`, `realm_access.roles`
- redirect/logout URI: local frontend URL 등록

### 6.2 External 모드

- 앱은 외부 issuer/discovery URL을 신뢰하고, local keycloak을 기동하지 않는다.
- realm/client/secret/issuer는 env로만 주입한다.
- 운영 체크리스트:
  - issuer/audience mismatch 검증
  - clock skew 허용 범위
  - JWKS rotation/cache 설정
  - TLS/CA 신뢰체인

## 7. 단계별 구현 계획 (PR 분할) — **모두 PR #167 로 머지됨**

> **머지 사실**: 본 plan 의 6단계는 단일 PR #167 (merge commit `dff487d`, 2026-05-18) 묶음으로 모두 main 머지됨. 별도 PR 분할 없이 codex 가 단일 큰 PR 으로 처리.

1. **PR-A: 설정/추상화** — ✅ **done** (PR #167 묶음)
- IdP provider 플래그, env 로더, 공통 config 정리
- `Keycloak` provider 스켈레톤 추가

2. **PR-B: Backend verifier 전환** — ✅ **done** (PR #167 묶음)
- JWKS 기반 토큰 검증 구현 (`internal/auth/keycloak_verifier.go`)
- role claim 매핑/actor context 정합 (resource_access fallback)
- auth middleware 경로 회귀 테스트 (`keycloak_verifier_test.go`)

3. **PR-C: Backend account/admin 전환** — ✅ **done** (PR #167 묶음)
- KratosAdmin adapter 제거/대체
- Keycloak Admin API로 account lifecycle 구현 (`internal/httpapi/keycloak_admin_client.go` + `_test.go`)

4. **PR-D: Frontend OIDC 전환** — ✅ **done** (PR #167 묶음)
- authorize/callback/logout/account flow 전환 (OIDC discovery 기반)
- Hydra/Kratos 전용 코드 제거 (auth.service / kratos-logout / kratos-webhook 등)
- legacy self-signup 비활성화

5. **PR-E: DB 마이그레이션** — ✅ **done** (PR #167 묶음)
- `kratos_identity_id` 일반화 (`idp_subject`)
- migration `000021_rename_kratos_identity_to_idp_subject` (up + down)
- backfill 및 조회 로직 전환

6. **PR-F: 테스트/문서/추적성** — ✅ **done** (PR #167 묶음)
- E2E 시나리오 갱신 (signup/auth specs Keycloak 정합)
- API contract / setup 가이드 / E2E 가이드 / requirements / architecture / frontend_integration_requirements 동기화
- ADR governance 측면 정정은 2026-05-19 sprint `claude/work_260519-a` 가 별도로 [ADR-0019](../adr/0019-keycloak-only-idp.md) 발행 + ADR-0001 supersession 으로 마무리.

## 8. 진척 관리 규칙

- 상태 값: `planned`, `in_progress`, `blocked`, `done`
- 모든 PR은 traceability 체크리스트 동기화 필수
- 각 PR 종료 조건:
  - unit/integration/e2e 최소 1회 green
  - 문서/메모리 갱신
  - 롤백 절차 검증

## 9. 리스크 및 대응

- Role claim 불일치: mapper 표준화 + fallback(조직 DB 조회) 유지
- logout/session 동작 차이: browser e2e 강화
- 외부 Keycloak 가용성 이슈: timeout/retry/circuit-breaker 정책 도입
- 점진 전환 중 회귀: provider feature flag로 긴급 rollback

## 10. 검증 계획

- Backend: auth/account 관련 단위/통합 테스트
- Frontend: auth callback/logout/account e2e
- Security: issuer/audience/exp/nbf 검증, key rotation
- Ops: local embedded + external mode 각각 smoke test
