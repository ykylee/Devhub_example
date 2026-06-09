# Session Handoff — feat/work_260609-a-swagger-apikey-expand

- **문서 목적**: 다음 세션이 본 sprint 의 작업 상태를 빠르게 복원하도록 핵심 사실만 기록한다.
- **sprint branch**: `feat/work_260609-a-swagger-apikey-expand`
- **최종 갱신일**: 2026-06-09 (session 완료 — commit/push/PR 직전)
- **관련 문서**: [`work_backlog.md`](./work_backlog.md), [`backlog/2026-06-09.md`](./backlog/2026-06-09.md), [`state.json`](./state.json), [ADR-0029](../../../docs/adr/0029-api-key-auth-and-swagger-scope.md), [sprint plan](../../../docs/planning/api-key-management-sprint-plan.md) (후속 multi-key sprint)

## 1. sprint 목표 (완료)

사용자 요청: "swagger 개발 이어서 하자. api 호출은 keycloak 과 무관하게 key만 있으면 사용할 수 있도록 할거야. 그리고 지금 swagger에 항목이 부족하니 보강하자." 의 1차 PR.

- (a) 공개 read-only API 의 Keycloak 종속을 끊고 정적 API key (`Authorization: Bearer <DEVHUB_API_KEY>`) 로 호출 가능. ✓
- (b) swagger openapi.yaml 의 4 paths / 5.6% 커버리지를 P0/P1 30+ endpoint 까지 확장. ✓

## 2. 사용자 결정 사항 (in-session)

- **인증 범위**: 공개 API = API key, admin / write API = Keycloak 유지.
- **전달 방식**: `Authorization: Bearer <key>` (Bearer scheme 재사용, JWT 와 동일 헤더).
- **API key 형식**: 단순 정적 문자열 (JWT 아님). 분기: 3-part dot-separated base64url 이면 JWT, 아니면 static key.
- **security scheme 표기**: openapi `securitySchemes` 에 `bearerAuth` + `staticTokenAuth` 두 scheme 병존 (openapi OR semantics).
- **PR 범위**: 단일 PR (auth 변경 + P0/P1 swagger 보강 + ADR + 테스트 일괄).
- **openapi.yaml SoT**: embedded asset 유지 (mirror 동기화 없음).
- **후속 결정 (follow-up 메시지)**: multi-key management (Backend + Frontend) + DB hashed key storage.

## 3. 완료된 작업 (모두 done)

### 3.1 인증 미들웨어 ✓
- `backend-core/internal/shared/config/config.go`: `APIKey` + `APIKeyAdminOnly` + env wire
- `backend-core/internal/domain/auth-session/view/handler.go`: `AuthConfig.APIKey` + `APIKeyAdminOnly` 추가
- `backend-core/internal/domain/auth-session/view/auth.go`: `AuthenticateActor` 에 API key 분기 + `looksLikeJWT` + `subtleEqual`
- `backend-core/internal/httpapi/router.go`: `RouterConfig.APIKey` + `APIKeyAdminOnly` + `NewAuthHandler` wire
- `backend-core/main.go`: env wire

### 3.2 테스트 ✓
- `backend-core/internal/httpapi/auth_test.go`: 5 신규 unit test PASS
- `backend-core/internal/httpapi/api_key_e2e_test.go` (신규 파일): 4 E2E test PASS
  - `TestAPIKeyEndToEnd_SwaggerServes` (openapi.yaml path/schema/securityScheme marker 회귀 가드)
  - `TestAPIKeyEndToEnd_PublicReadEnvelopes` (8 endpoint 통합 인증)
  - `TestAPIKeyEndToEnd_KeycloakJWTStillWorks` (JWT path 회귀)
  - `TestAPIKeyEndToEnd_NoAuthKeycloakStillLocked` (security invariant)
  - `TestAPIKeyEndToEnd_StaticKeyVerifierError` (verifier 우선순위)

### 3.3 ADR ✓
- `docs/adr/0029-api-key-auth-and-swagger-scope.md` (신규, 182 lines). 결정 6 근거 + 7 carve out + 3 옵션 비교.

### 3.4 openapi.yaml 보강 ✓
- 527 lines → 5,999 lines (+5,472)
- paths: 4 → 52 (P0/P1 48 신규)
- schemas: 7 → 59 (52 신규)
- securitySchemes: `bearerAuth` + `staticTokenAuth`
- tags: 4 → 18 (14 신규)
- yaml safe_load VALID

### 3.5 회귀 검증 ✓
- `go test ./...` 30+ packages PASS
- `go build ./...` PASS
- 신규 테스트 9건 PASS, 회귀 0

### 3.6 Traceability 갱신 ✓
- `docs/traceability/report.md` §2.4 IMPL-auth-04 + IMPL-swagger-02 추가
- §6 변경 이력에 본 sprint entry 추가 (2026-06-09)

## 4. 잔여 / 후속 작업

1. **commit + push + PR (gh CLI)**: 본 1차 sprint 의 모든 코드/문서/테스트가 staging area 에 있음. 사용자가 결정한 의도 = commit + push + PR 자동.
2. **multi-key management sprint** (사용자 follow-up 결정):
   - `feat/work_260609-b-api-key-management-backend` (Phase 1: Backend) + `feat/work_260609-b-api-key-management-frontend` (Phase 2: Frontend)
   - Backend: `api_keys` table (DB hashed) + CRUD endpoints + auth middleware DB lookup 확장
   - Frontend: `/admin/api-keys` page (list / create dialog / revoke button) + frontend 도메인 mirror
   - detail: [`docs/planning/api-key-management-sprint-plan.md`](../../../docs/planning/api-key-management-sprint-plan.md)
3. **ADR-0029 §6 carve**: (a) RBAC 가드, (b) rotation SOP, (c) P2/P3 endpoint 30+ 확장, (d) CI lint, (e) swagger-ui 가드, (g) audit 강화 — multi-key sprint 와 별개 후속

## 5. 핵심 파일 / 라인 참조

- `backend-core/internal/domain/auth-session/view/auth.go:155-180` — API key 분기
- `backend-core/internal/domain/auth-session/view/auth.go:320-374` — looksLikeJWT + subtleEqual
- `backend-core/internal/httpapi/auth_test.go` 끝 — TestAPIKeyAuthentication_* 5건
- `backend-core/internal/httpapi/api_key_e2e_test.go` (신규) — TestAPIKeyEndToEnd_* 4건
- `docs/adr/0029-api-key-auth-and-swagger-scope.md` — ADR 본문
- `backend-core/internal/httpapi/swaggerui/asset/openapi.yaml` — 5,999 lines
- `docs/traceability/report.md` §2.4 line 240+ — IMPL-auth-04 + IMPL-swagger-02

## 6. 알아둘 trade-off (의도적, ADR-0029 §5/§6 명시)

- **API key caller 의 admin endpoint RBAC 가드 미구현** — §6 carve (a) 의 후속 sprint. 1차 PR 의 mitigation: 운영 SOP (DEVHUB_API_KEY 는 staging/dev 에서만 발급).
- **DEVHUB_API_KEY 단일 정적 키** — multi-key management 는 후속 sprint `feat/work_260609-b-api-key-management-backend`.
- **API key caller 의 audit 강화** — `devhub_actor_login="api-key"` + `X-Devhub-Auth: api_key` header 만 부착. §6 carve (g).

## 7. 다음 세션이 가장 먼저 할 일

1. `git status` / `git diff --staged` / `git log --oneline -5` 확인.
2. multi-key sprint branch 생성: `git checkout -b feat/work_260609-b-api-key-management-backend main` (main 머지 후).
3. `docs/planning/api-key-management-sprint-plan.md` 의 Phase 1 부터 진행.
4. 또는 multi-key 결정 변경 / 추가 요구사항 반영.
