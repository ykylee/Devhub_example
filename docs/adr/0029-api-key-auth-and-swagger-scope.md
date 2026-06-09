# ADR-0029: API key 인증 (Keycloak-independent) + OpenAPI scope 확장

- **문서 목적**: 공개 API (public endpoints) 의 Keycloak 종속을 끊고 정적 API key (`Authorization: Bearer <key>`) 로도 호출 가능하도록 인증 경로를 보강하는 결정을 명문화한다. 동시에 1차 bootstrap (ADR-0027) 의 4 paths / 5.6% 커버리지를 도메인 P0/P1 endpoint 30+ 까지 확장하는 결정도 본 ADR 에 포함한다.
- **범위**: `backend-core/internal/shared/config/config.go` 의 `APIKey` / `APIKeyAdminOnly` env + `backend-core/internal/domain/auth-session/view/handler.go` + `auth.go` 의 API key 분기 (JWT vs 정적 키) + `RouterConfig.APIKey` / `RouterConfig.APIKeyAdminOnly` + `backend-core/main.go` env wire + `backend-core/internal/httpapi/swaggerui/asset/openapi.yaml` 의 components/schemas 확장 (~30 schema 추가) + `paths` 섹션 P0/P1 30+ endpoint 추가 + `securitySchemes` 에 `staticTokenAuth` 추가.
- **대상 독자**: Backend / 프론트엔드 개발자, AI agent, API consumer, QA, 운영자, DevOps.
- **상태**: accepted
- **최종 수정일**: 2026-06-09
- **결정 근거 sprint**: `feat/work_260609-a-swagger-apikey-expand` (v1.0 출시 직전 swagger 보강 1차)
- **관련 문서**: [ADR-0019](./0019-keycloak-only-idp.md) (Keycloak-only IdP — 본 ADR 의 supersession 추가), [ADR-0027](./0027-openapi-hand-maintained.md) (1차 bootstrap, 본 ADR 의 §3 scope 확장 정합), [docs/openapi.yaml](../openapi.yaml) (1차 527 lines → ~1500 lines 확장 대상), [docs/backend_api_contract.md §0/§1/§3](../backend_api_contract.md) (envelope/enum cross-link), [docs/traceability/report.md §2.4 IMPL-auth-02 / IMPL-swagger-02](../traceability/report.md), [docs/governance/sync-checklist.md §3.2](../governance/sync-checklist.md) (backend_api_contract.md ↔ openapi.yaml cross-link step), [AGENTS.md §v1.0 릴리즈 로드맵](../../AGENTS.md) (워커 분업 전면 취소 결정 정합).

## 1. 배경

### 1.1 Keycloak-only 종속이 swagger UI + 공개 API 에 미치는 영향

v1.0 출시 직전 시점에 DevHub 백엔드 API 가 70+ endpoint 에 도달했고, ADR-0027 의 swagger UI 1차 bootstrap 으로 4 path (health, me, logout, metrics) 만 노출되었다. 그러나 두 가지 한계가 운영 시 다음과 같은 마찰을 만들었다:

1. **Keycloak outage 시 swagger UI 자체 사용 불가** — Keycloak JWKS endpoint 가 다운되거나 network unreachable 상태가 되면 `authenticateActor` 미들웨어가 모든 `/api/v1/*` 호출에 401 을 반환. swagger UI 페이지 자체는 열리지만 `Authorize` 후 모든 endpoint 가 401.
2. **공개 API (read-only 조회성 endpoint) 호출에 Keycloak 로그인 필수** — `/api/v1/repositories`, `/api/v1/issues`, `/api/v1/risks` 같은 read-only 공개 API 도 Keycloak OIDC 로그인 후 발급받은 access token 이 필요. CI / 스크립트 / 외부 consumer 통합 부담.

### 1.2 swagger spec coverage 부족 (5.6%)

1차 bootstrap 의 openapi.yaml 은 4 paths / 7 schemas 만 정의 (527 lines). router.go 가 노출하는 70+ endpoint 중 64+ 가 spec 에 미정의. 운영자 / QA / 외부 consumer 가 API contract 를 사전에 확인할 수 없음. ADR-0027 §6 carve (d) 의 "별도 sprint 에서 도메인 endpoint 100+ spec 흡수" 가 본 ADR 에서 1차 처리됨.

### 1.3 검토 요구

- (a) Keycloak 도달 불필요한 정적 인증 수단 제공 (공개 read-only endpoint 한정, 관리자 endpoint 는 Keycloak 유지)
- (b) JWT 와의 정적 키 구분 (호환성 — Keycloak JWT 가 기존 bearerAuth scheme 그대로 동작)
- (c) API key 가 노출되어도 admin endpoints 접근 불가 (RBAC + 인증 source 분리)
- (d) openapi.yaml 을 30+ endpoint 까지 확장 (P0/P1) — 도메인 endpoint 64+ 의 1차 절반 흡수
- (e) timing attack 회피 (constant-time string comparison)
- (f) 기존 Keycloak 종속 회귀 0

## 2. 후보 옵션

### 2.1 인증 수단 (3종)

| # | 옵션 | 호환성 | 보안 | 결정 |
| --- | --- | --- | --- | --- |
| 1 | `X-Api-Key: <key>` 별도 헤더 | swagger-ui `Authorize` 에서 별도 버튼 필요 | 명확 (header 단일) | ❌ (swagger-ui 멀티 scheme UX 복잡) |
| 2 | **API key = `Authorization: Bearer <static-string>`** (Bearer scheme 재사용) | swagger-ui `Authorize` 가 한 scheme 안에서 입력 | Keycloak JWT 와 동일 자리 (충돌 없음 — JWT 포맷 검증으로 분기) | ⭐ **채택** |
| 3 | Basic auth (`Authorization: Basic base64(id:key)`) | HTTP 표준 | credentials base64 인코딩 (보안 약점) | ❌ (가독성↓) |

### 2.2 API key 적용 범위 (3종)

| # | 옵션 | 공개 API 호출 | admin API 호출 | 결정 |
| --- | --- | --- | --- | --- |
| A | 모든 endpoint API key 허용 | ✅ | ✅ | ❌ (admin API key 노출 시 위험) |
| B | **공개 API 만 API key 허용 + admin API 는 Keycloak 강제** (본 결정) | ✅ | ❌ | ⭐ **채택** |
| C | API key 자체 미지원 | ❌ | ❌ | ❌ (요구사항 미충족) |

옵션 B 의 admin API gate 는 향후 §6 carve (a) 의 `enforceRoutePermission` 에서 `auth_source != "api_key"` 가드를 추가하여 enforce. 본 1차 PR 에서는 AuthHandler 가 `devhub_actor_login=api-key` + `devhub_auth_source=api_key` 로 식별 가능하게 set 하고, RBAC 매트릭스 정합은 후속 sprint 에서 별도 처리. **1차 PR 의 보안 trade-off** 는 §5.1 에 명시.

### 2.3 openapi.yaml scope 확장 (2종)

| # | 옵션 | 작업량 | 결정 |
| --- | --- | --- | --- |
| 1 | **P0/P1 endpoint 30+ 1차 흡수** (본 결정) | ~30 endpoints | ⭐ **채택** |
| 2 | P0/P1/P2 전체 endpoint 60+ 흡수 | ~60 endpoints | ❌ (PR 크기 과다, review 부담) |

## 3. 결정

**옵션 2-1-1 (API key = Bearer + 옵션 B 범위 + openapi P0/P1 1차) 채택**.

### 3.1 인증 미들웨어

`auth.go::AuthenticateActor` 의 분기 로직:

1. `Authorization: Bearer <token>` 파싱 (기존)
2. **API key 분기 (신규)**: `cfg.APIKey != ""` 이고 `!looksLikeJWT(token)` 이면 `subtleEqual(token, cfg.APIKey)` 비교. 일치 시 `devhub_actor_login=api-key`, `devhub_actor_role=system_admin`, `devhub_auth_source=api_key`, `X-Devhub-Auth: api_key` header 부착 + 다음 미들웨어 호출.
3. JWT 분기 (기존, 변경 없음): Keycloak JWKS verifier 호출.

`looksLikeJWT(token)` 은 `header.payload.signature` 형태 (3-part, base64url charset) 인지 검사. 정확성보다 단순성 우선 — false negative (JWT 가 API key 분기로 잘못 분류) 가 발생해도 `cfg.APIKey == ""` 시 분기 자체가 skip 되어 영향 없음.

`subtleEqual` 은 constant-time string comparison (timing attack 회피). `crypto/subtle` 미사용 — 본 비교는 32~128 byte 정적 키 수준에서 Go 의 `crypto/subtle.ConstantTimeCompare` 와 동등.

### 3.2 openapi.yaml scope 확장

`backend-core/internal/httpapi/swaggerui/asset/openapi.yaml` 의 `components/schemas` 에 약 30 schema 추가 (Platform, Project, Repository, Issue, PullRequest, CIRun, Risk, DevRequest, IntegrationProvider, IntegrationBinding, AuditLog, RBACPolicy, Organization, OrgUnit, ExternalTask, UserSummaryExtended, PaginatedResponse 등). `paths` 섹션에 P0/P1 30+ endpoint 추가.

`securitySchemes` 에 `staticTokenAuth` 추가:

```yaml
securitySchemes:
  bearerAuth:
    type: http
    scheme: bearer
    bearerFormat: JWT
  staticTokenAuth:
    type: http
    scheme: bearer
    bearerFormat: Static
    description: DEVHUB_API_KEY 정적 키. Keycloak-independent.
```

각 endpoint 의 security:
- `/health`, `/metrics` → `security: []` (no auth, public)
- `/api/v1/internal/keycloak-events` → `security: []` + X-Webhook-Secret header
- 공개 read-only API (예: `/api/v1/repositories`, `/api/v1/issues`, `/api/v1/risks`) → `security: [{staticTokenAuth: []}]` (API key 가능) 또는 `[{bearerAuth: []}]` (Keycloak 가능) — openapi 의 OR 의미
- admin / write API (예: `/api/v1/admin/*`, `/api/v1/rbac/*`, `/api/v1/dev-request-tokens/*`, `/api/v1/users/*` PATCH/DELETE) → `security: [{bearerAuth: []}]` (Keycloak 강제)

### 3.3 결정 근거 6 항목

1. **Keycloak-independent 운영 가용성** — Keycloak JWKS outage 시 swagger UI + 공개 read-only API 도 호출 가능 (sprint 의 핵심 동기).
2. **JWT 와의 정적 키 구분 명확** — 3-part dot-separated base64url 검사로 분기. false positive (JWT 가 API key 로) 0%, false negative (API key 가 JWT 로) 발생해도 `cfg.APIKey` 미설정 시 무영향.
3. **Timing attack 회피** — `subtleEqual` constant-time comparison.
4. **기존 Keycloak 회귀 0** — 분기 추가가 기존 Keycloak path 의 미변경. 신규 테스트 `TestAPIKeyAuthentication_*` 5건 + 기존 `TestBearerTokenActor*` 회귀 가드.
5. **swagger scope 5.6% → 40%+ 1차 흡수** — P0/P1 30+ endpoint + ~30 schema. P2/P3 endpoint 30+ 는 후속 sprint (`feat/openapi-domain-extend-p2` 후보, §6 carve (d)).
6. **PR 크기 적정** — 인증 미들웨어 + openapi 보강 + ADR + 신규 테스트 = 단일 PR 검토 가능.

### 3.4 trade-off 인정

- API key 가 system_admin role 로 들어옴 → §5.1 의 admin API 가드 (1차 PR 에서는 식별만, gate 는 후속) + §6 carve (a) 의 `enforceRoutePermission` 가드 추가
- API key rotation 정책 미정 → §5.2 + §6 carve (b) 운영 SOP 필요
- swagger spec hand-maintained 부담 → ADR-0027 §5.4 의 trade-off 와 동일 (확장 부담은 본 ADR 이 추가)

## 4. 결과

### 4.1 코드 / 자산 변경 요약

- `backend-core/internal/shared/config/config.go`. `APIKey string` (env `DEVHUB_API_KEY`) + `APIKeyAdminOnly bool` (env `DEVHUB_API_KEY_ADMIN_ONLY`) 필드 추가. `Load()` env wire.
- `backend-core/internal/domain/auth-session/view/handler.go`. `AuthConfig.APIKey` + `AuthConfig.APIKeyAdminOnly` 필드 추가.
- `backend-core/internal/domain/auth-session/view/auth.go`. `AuthenticateActor` 에 API key 분기 추가 (bearerToken 파싱 후). `looksLikeJWT` + `subtleEqual` helper 추가.
- `backend-core/internal/httpapi/router.go`. `RouterConfig.APIKey` + `RouterConfig.APIKeyAdminOnly` 필드 추가 + `NewAuthHandler` 호출 시 wire.
- `backend-core/main.go`. `cfg.APIKey` + `cfg.APIKeyAdminOnly` env wire.
- `backend-core/internal/httpapi/auth_test.go`. `TestAPIKeyAuthentication_ValidKeyPassesThrough` / `_InvalidKeyReturns401` / `_JWTFormatGoesToKeycloakVerifier` / `_EmptyKeyDoesNotActivate` / `TestLooksLikeJWT` 5 신규 테스트 케이스 추가.
- `backend-core/internal/httpapi/swaggerui/asset/openapi.yaml`. components/schemas ~30 schema 추가 + paths P0/P1 30+ endpoint 추가 + securitySchemes 에 `staticTokenAuth` 추가. 기존 4 path + 7 schema 회귀 0.
- 신규 ID: `IMPL-auth-02` (API key middleware) + `IMPL-swagger-02` (openapi scope 1차 확장).

### 4.2 라우트 / 미들웨어

변경 없음. `authenticateActor` 미들웨어 내부에 API key 분기 추가. 기존 70+ endpoint 모두 영향 없음 (단, API key caller 가 admin endpoints 도 호출 가능한 상태 — §5.1 trade-off).

### 4.3 검증

- 5 신규 테스트 PASS (`TestAPIKeyAuthentication_*` + `TestLooksLikeJWT`)
- **2차 commit** (P0 §6 (a)): 2 신규 테스트 PASS (`TestEnforceRoutePermission_APIKeyCallerWriteEndpointsForbidden` 6 sub-test + `TestEnforceRoutePermission_APIKeyCallerReadOnlyAllowed` 5 sub-test)
- **3차 commit** (P1 §6 (c) P2 batch 1): `python3 -c "import yaml; yaml.safe_load(open('...'))"` PASS, 5 신규 path (user CRUD 4 + risks/critical 1 + risks/{id}/mitigations 1 + commands/{id}/{approve,reject} 2) + 5 신규 schema (UserCreateRequest / UserUpdateRequest / RiskMitigationRequest / RiskMitigation / CommandDecisionRequest) + commands/{id} GET 의 staticTokenAuth 허용. path 50 → 55.
- **4차 commit** (P1 §6 (c) P2 batch 2): yaml safe_load PASS, 5 신규 path (org units CRUD 4 + /units/{id}/members GET|PUT 2 + hierarchy PUT 1 + rbac/policy legacy 410 1 + rbac/policies GET|POST|PUT 3 + rbac/policies/{id} DELETE 1 = 11 신규 endpoint) + 6 신규 schema (OrgUnitCreateRequest / OrgUnitUpdateRequest / OrgUnitMemberReplaceRequest / OrganizationHierarchyRequest / RBACPolicyCreateRequest / RBACPolicyBulkUpdateRequest). 1차 PR 50 path → 4차 commit 후 62 path, schema 59 → 70.
- 기존 `TestBearerToken*` 회귀 0 (`TestRouter_*` / `TestSwagger*` 포함)
- backend `go test ./internal/httpapi/ -run 'TestAPIKey|TestBearerToken|TestRouter_|TestSwagger|TestEnforceRoutePermission'` PASS
- backend `go test ./...` PASS (회귀 0, 30+ packages)
- openapi.yaml syntax 검증: `python3 -c "import yaml; yaml.safe_load(open('...'))"` PASS
- 운영 가드: `DEVHUB_API_KEY` 미설정 시 분기 미작동 (회귀 0) — `TestAPIKeyAuthentication_EmptyKeyDoesNotActivate` 가드
- **2차 commit 검증**: API key caller 의 mutation endpoint (POST /api/v1/users / PATCH /api/v1/users/:id / DELETE /api/v1/users/:id / POST /api/v1/platforms / POST /api/v1/dev-requests/:id / POST /api/v1/risks/:id/mitigations) 6 endpoint 모두 403 + `auth_api_key_denied` envelope + `auth.api_key_denied` audit row 6 row 검증. read-only 5 endpoint (GET /api/v1/{repositories,issues,risks,dashboard/metrics,audit-logs}) 정상 통과.

## 5. trade-off + carve out

### 5.1 API key caller 의 admin endpoint 접근 (1차 PR 의 의도적 trade-off)

**문제**: API key 가 `devhub_actor_role=system_admin` 으로 식별되어, `enforceRoutePermission` 이 admin endpoints 에서도 통과시킬 수 있음. Keycloak 사용자보다 넓은 권한.

**1차 PR 의 의도**: openapi.yaml 의 security scheme 표기는 그대로 `bearerAuth` 만 admin endpoints 에 적용 (openapi spec level 의 가드). 실제 백엔드 RBAC 가드는 §6 carve (a) 의 후속 sprint 에서 `enforceRoutePermission` 에 `auth_source != "api_key"` 조건 추가.

**완화 (1차)**: 운영 환경에서 `DEVHUB_API_KEY_ADMIN_ONLY=1` 환경변수 + AuthHandler 의 `cfg.APIKeyAdminOnly` 분기 추가 (구현만, 미적용). 후속 sprint 에서 가드 활성화. 운영 SOP: API key 는 staging/dev 환경에서만 발급, prod 에서는 비권장.

### 5.2 API key rotation 정책

**문제**: `DEVHUB_API_KEY` 가 단일 정적 키. 키 유출 시 rotate 가 환경변수 교체 + backend 재시작 필요.

**완화**: §6 carve (b). 운영 SOP 에서 90일 rotation + 다중 키 동시 활성 (rolling) 패턴 결정. 1차 PR 에서는 단일 키만 지원.

### 5.3 swagger spec hand-maintained 부담

ADR-0027 §5.4 와 동일한 trade-off. 본 ADR 이 흡수하는 ~30 endpoint + ~30 schema 추가 = hand-maintained 부담 +1 sprint.

### 5.4 OpenAPI 의 security OR semantics

openapi 3.0 의 security array 가 `[{a:[]}, {b:[]}]` 일 때 swagger-ui 가 양쪽 scheme 모두 "Authorize" 버튼으로 노출. 본 ADR 의 `staticTokenAuth` + `bearerAuth` OR 적용은 swagger-ui UX 일관성 측면에서 자연스럽고, backend 가 `Authorization: Bearer <key-or-jwt>` 단일 헤더로 양쪽을 수신.

## 6. carve out (후속 sprint)

| # | 항목 | 우선순위 | 비고 |
| --- | --- | --- | --- |
| (a) | API key caller 의 admin endpoint RBAC 가드 (`enforceRoutePermission` 에 `auth_source != "api_key"` 추가) | ✅ done (2차 commit) | 본 PR 2차 commit 으로 즉시 흡수. `devhub_auth_source == "api_key"` && `policy.Action != ActionView` (Create/Edit/Delete) 시 `auth.api_key_denied` audit + 403 + `auth_api_key_denied` envelope. 신규 테스트 2건. |
| (b) | API key rotation 정책 SOP (90일 + rolling pattern) | P1 | 운영 SOP 문서 (`docs/setup/`) |
| (c) | openapi.yaml P2/P3 endpoint 30+ 확장 | ✅ done (P2 batch 1 + batch 2) → 3차 PR (P3 잔여) | 본 PR 2차 commit 의 P2 batch 1 = 7 신규 path + 5 신규 schema. 3차 commit 의 P2 batch 2 = 5 신규 path (org units CRUD 4 + /units/{id}/members GET|PUT 2 + hierarchy PUT 1 + rbac/policy legacy 410 1 + rbac/policies GET|POST|PUT 3 + rbac/policies/{id} DELETE 1 = 11 신규) + 6 신규 schema. 1차 50 path → 3차 PR 후 62 path, schema 59 → 70. P3 잔여 (integration/scm tail 14 + realtime/ws + infra/services/snapshot + infra/topology/v2 = 17 path) → 4차 PR. |
| (d) | CI lint gate (openapi.yaml schema validity check + cross-link `backend_api_contract.md`) | P2 | v1.1 milestone |
| (e) | swagger-ui system_admin 가드 미들웨어 (현재 public) | P2 | v1.1 milestone (ADR-0027 §6 carve (c)) |
| (f) | API key 다중 활성 (rolling) 지원 | P3 | v1.1+ |
| (g) | API key 사용 audit 강화 (callsite 별 X-Request-ID 부착) | P2 | v1.1 |

## 7. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-06-09 | 1차 발행. API key 인증 미들웨어 + openapi.yaml P0/P1 30+ endpoint + ~30 schema + 5 신규 테스트. 결정 근거 6 항목 (Keycloak-independent / JWT 분기 / timing attack / 회귀 0 / scope 5.6%→40%+ / PR 크기 적정). carve out 7 항목. 신규 ID: `IMPL-auth-02` + `IMPL-swagger-02`. | `feat/work_260609-a-swagger-apikey-expand` |
| 2026-06-09 | **2차 commit (PR #518 후속, codex P1 + 사용자 결정 정합)**. §6 (a) P0 admin gate 실 구현 — `EnforceRoutePermission` 본체에 `devhub_auth_source == "api_key"` && `policy.Action != ActionView` 가드. `auth.api_key_denied` audit + 403 + `auth_api_key_denied` envelope. 신규 테스트 2건 (write 6 endpoint 거부 + read-only 5 endpoint 통과 회귀). `routePermissionTable` 변경 0 — Action != View 한 줄 분기. `IMPL-auth-04` row 갱신 (permissions.go 가드 + permissions_test.go 2 신규). `go test ./...` 30+ packages PASS, 회귀 0. ADR-0029 §6 (a) P0 → done. 잔여 carve = 6 항목 (b)~(g). | `feat/work_260609-a-swagger-apikey-expand` |
| 2026-06-09 | **3차 commit (PR #519, §6 (c) P1 P2 batch 1)**. openapi.yaml scope 1차 P2 batch 흡수 — 5 신규 path (user CRUD 4 + risks/critical 1 + risks/{id}/mitigations 1 + commands/{id}/{approve,reject} 2) + 5 신규 schema (UserCreateRequest / UserUpdateRequest / RiskMitigationRequest / RiskMitigation / CommandDecisionRequest) + commands/{id} GET 의 staticTokenAuth 허용 (infrastructure:view 정합). 1차 PR 의 50 path → 본 PR 후 55 path + schema 59 → 64. backend 변경 0 — openapi.yaml + ADR + trace docs only. yaml safe_load VALID. §6 (c) P1 P2 batch 1 → done, 잔여 = P2 batch 2 (org-units CRUD 7 + rbac/policies 5) + P3 (integration/scm tail 14 + realtime/ws + infra/services/snapshot + infra/topology/v2) = 29 path → 3차 PR. | `feat/work_260609-b-openapi-p2p3-extend` |
| 2026-06-09 | **4차 commit (PR #520, §6 (c) P1 P2 batch 2)**. openapi.yaml scope 2차 P2 batch 흡수 — 5 신규 path (org units CRUD 4 + /units/{id}/members GET|PUT 2 + hierarchy PUT 1 + rbac/policy legacy 410 1 + rbac/policies GET|POST|PUT 3 + rbac/policies/{id} DELETE 1 = 11 신규 endpoint) + 6 신규 schema (OrgUnitCreateRequest / OrgUnitUpdateRequest / OrgUnitMemberReplaceRequest / OrganizationHierarchyRequest / RBACPolicyCreateRequest / RBACPolicyBulkUpdateRequest). 기존 1차 PR GET (/api/v1/organization/{hierarchy,units/{id},units/{id}/members,organizations/search}) 보존. 1차 50 path → 4차 commit 후 62 path, schema 59 → 70. info.version "0.2.0-p2-extension" → "0.3.0-p2-batch2" 정합. backend 변경 0 — openapi.yaml + ADR + trace docs only. yaml safe_load VALID. §6 (c) P1 P2 batch 2 → done, 잔여 = P3 (integration/scm tail 14 + realtime/ws + infra/services/snapshot + infra/topology/v2 = 17 path) → 4차 PR. | `feat/work_260609-c-openapi-p2-batch2` |
