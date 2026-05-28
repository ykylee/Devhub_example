# backend-core/internal/httpapi 패키지 분석

- 문서 목적: DevHub backend-core 의 HTTP API 계층(`backend-core/internal/httpapi/`) 을 핸들러·라우트·인증/권한 미들웨어 단위로 전수 분석하고, 발견된 불일치/부채/보안 주의점을 근거와 함께 기록한다.
- 범위: 비-test Go 파일 42개 + `router.go` + `permissions.go` + `auth.go` + `request_context.go` + 미들웨어 체인.
- 분석 기준 시점: 2026-05-27 (main `cf19c94`).
- 코드 경로/심볼은 원문 그대로 유지한다.

---

## 1. 패키지 개요

### 역할
`httpapi` 는 gin 기반의 단일 HTTP 진입점이다. `NewRouter(cfg RouterConfig) *gin.Engine` 이 모든 라우트를 등록하고, `Handler{cfg}` 의 메서드들이 각 endpoint 를 구현한다. 패키지는 **store layer 에 대한 의존을 인터페이스로 분리**(`DomainStore`, `ApplicationStore`, `OrganizationStore`, `DevRequestStore`, `IntakeTokenStore`, `RBACStore`, `CommandStore`, `AuditStore`, `WebhookEventStore`, `HealthStore`, `IdentityAdmin`, `HRDBClient`, `SnapshotProvider`, `realtimeTicketStore`)해 두어, 프로덕션은 `*store.PostgresStore` 를, 테스트는 in-memory fake 를 주입한다.

### 주요 의존
- `github.com/gin-gonic/gin` — 라우터/컨텍스트
- `github.com/devhub/backend-core/internal/domain` — 도메인 타입·RBAC matrix·상태전이 규칙
- `github.com/devhub/backend-core/internal/store` — 영속 계약 + sentinel error (`ErrNotFound`, `ErrConflict`, `ErrDuplicateEvent`, `ErrAuditInvariantViolation`, `ErrSystemRoleImmutable`, `ErrRoleInUse`)
- `github.com/devhub/backend-core/internal/gitea` — SCM 어댑터(HMAC 서명 검증 `VerifySignature`, REST client)
- `github.com/devhub/backend-core/internal/integrations/adapters` — HomeLab infra snapshot 정규화
- `github.com/prometheus/client_golang` — `/metrics` + onboarding metric
- `github.com/gorilla/websocket` — realtime hub

### 응답 컨벤션
모든 핸들러가 `gin.H{"status": ...}` 봉투를 쓴다. `status` 값: `ok` / `created` / `deleted` / `accepted` / `rejected` / `conflict` / `not_found` / `gone` / `unavailable` / `unauthenticated` / `unauthorized` / `forbidden` / `failed`. list 응답은 `{"status","data","meta"}`, 일부 검증 실패는 머신 파싱용 `code` 필드를 추가한다. 내부 서버 오류는 `writeServerError`(`errors.go`)가 SQL/스키마 누설 없이 일괄 500 `{"status":"failed","error":"internal error"}` 로 변환.

---

## 2. 파일별 1줄 요약 (비-test 42개 + 핵심 인프라)

| 파일 | 1줄 요약 |
| --- | --- |
| `router.go` | `NewRouter` + 전 라우트 등록 + `RouterConfig`(store 주입) + store 인터페이스 정의 + `/health` + trusted-proxy 설정 |
| `permissions.go` | `PermissionCache`(rbac matrix in-memory) + `routePermissionTable`(route↔resource:action) + `enforceRoutePermission` + `enforceRowOwnership` |
| `auth.go` | `authenticateActor` 미들웨어 — Bearer 검증, ticket/access-token, 동적 role lookup, lazy idp_subject backfill, onboarding flag set |
| `request_context.go` | `requireRequestID` 미들웨어 + request-id 생성/검증 + `logRequest`/`logRequestCtx` + `sourceTypeFrom`/`clientIPFrom` |
| `onboarding_gate.go` | `onboardingGate` 미들웨어 — 미완료 actor 를 allowlist 외 endpoint 에서 403 차단 (feature flag gated) |
| `onboarding_feature_flag.go` | `requireOnboardingFlag` — flag OFF 시 onboarding endpoint 4종을 404 처리 (handler-level conditional registration) |
| `onboarding_roles.go` | `onboardingValidRole` — fallback role 화이트리스트 검증 helper |
| `onboarding_metrics.go` | Prometheus metric 4종(gate_blocked/submit_total/submit_duration/review_confirm) + pending_review gauge 등록 |
| `onboarding_pending_gauge.go` | `RunOnboardingPendingReviewGauge` cron worker — pending_review 카운트 주기 갱신 |
| `dev_request_intake_auth.go` | `requireIntakeToken` 미들웨어 — DREQ 외부수신 API 토큰 + IP allowlist 인증 (`IntakeTokenStore`) |
| `me.go` | `getMe`(API-32) + `patchMe`(API-85) — self profile, onboarding state hydrate |
| `me_onboarding.go` | `submitOnboarding`(API-83) — onboarding 제출 + row INSERT/UPDATE + audit |
| `organizations_search.go` | `searchOrganizations`(API-84) — org_unit typeahead |
| `users_admin_review.go` | `confirmUserReview`(API-86) — admin pending_review → reviewed 전이 |
| `organization.go` | `OrganizationStore` 정의 + users/org-units/hierarchy/unit-members CRUD + cycle detection |
| `applications.go` | SCM provider(API-41/42) + Application CRUD(API-43..47) + Application-Repository link(API-48..50) + 상태전이 가드 |
| `application_rollup.go` | `applicationRollup`(API-57) — weight_policy 기반 롤업 집계 |
| `projects.go` | Project CRUD(API-55/56) + standalone/application-centric project + legacy/v2/hybrid 모드 게이트 |
| `repository_ops.go` | Repository 운영지표 read(API-51..54) — activity/PR/build-run/quality-snapshot |
| `domain.go` | repositories list + `createRepositoryDraft`/`requestRepositoryPublish`(draft→publish) + issues/pull-requests/risks list + list 봉투 |
| `integration_registry.go` | Integration provider/binding CRUD(API-69..75,80) + sync(API-72) + webhook ingest(API-73) + test-connection(API-87) + outbound webhook 서명 검증 |
| `integration_scm_repositories.go` | SCM repo 연동 — list(API-88)/import(API-89)/create(API-90) + capability gate + gitea-compat 검사 |
| `integrations.go` | project_integrations CRUD(API-58) — scope polymorphism(application/project) |
| `infra_integrations.go` | infra snapshot ingest(`InfraAgentToken` Bearer) + services/topology-v2 + HomeLab adapter |
| `dev_requests.go` | DREQ CRUD(API-59..65) — intake/list/get/register(promote)/reject/reassign/close + 단일-tx promote |
| `dev_request_intake_tokens_admin.go` | DREQ intake token admin(API-66..68 + IP/expiry update) — plain token 1회 노출, SHA-256 보관 |
| `rbac.go` | `RBACStore` 정의 + rbac/policies CRUD(legacy 410 + list/create/bulk-update/delete) + cache invalidate |
| `audit.go` | `listAuditLogs` + `recordAudit`/`recordAuditBestEffort` (actor/source_ip/request_id/source_type stamp) |
| `commands.go` | service-action/risk-mitigation command 생성 + approve/reject + `requestActor`/`actorLogin`/`devFallbackEnabled` helper |
| `gitea_webhook.go` | `receiveGiteaWebhook` — HMAC 서명 검증 + dedupe + EventProcessor 위임 |
| `keycloak_events_webhook.go` | `receiveKeycloakEventWebhook` — SPI push, `X-Webhook-Secret` 검증, audit row 변환 |
| `realtime.go` | `RealtimeHub` + `handleRealtimeWebSocket` — WS subscribe + per-event RBAC 검사 |
| `realtime_ticket.go` | ticket store(in-memory + PG) + `issueRealtimeTicket` (ADR-0024 §3.2) |
| `snapshot.go` | dashboard metrics + infra nodes/edges/topology + ci-runs/ci-logs + critical-risks (DB → static fallback) |
| `snapshot_provider.go` | `SnapshotProvider` 인터페이스 + `StaticSnapshotProvider`(하드코딩 mock 데이터) |
| `runtime_snapshot_provider.go` | `RuntimeSnapshotProvider` — static 위에 live health probe 덧씌우는 decorator |
| `events.go` | `listWebhookEvents` + `parseBoundedInt` 공용 helper |
| `hr_lookup.go` | `hrLookup` — HRDB 조회 (`HRDBClient`) |
| `authz.go` | `requireMinRole` 미들웨어 + role rank (**현재 라우터에 wire 안 됨, 아래 발견사항 참조**) |
| `identity_resolver.go` | `resolveIdPSubject` helper (**테스트만 사용, 핸들러 미참조**) |
| `keycloak_admin_client.go` | `KeycloakAdminClient`(`IdentityAdmin` 구현) — read-only (FindIdentityByUserID) |
| `errors.go` | `writeServerError` — 일괄 500 + 내부정보 누설 차단 |

---

## 3. 등록된 HTTP 라우트 전수 표

`NewRouter`(router.go) 의 등록 순서 기준. RBAC 컬럼은 `permissions.go::routePermissionTable` 의 값. **Bypass** = 인증은 거치되 RBAC matrix lookup 생략. v1 group 미들웨어 체인은 §4 참조.

### group 외부 (인증 미들웨어 비적용)
| 메서드 | 경로 | 핸들러 | 인증 |
| --- | --- | --- | --- |
| GET | `/health` | `health` | 없음 (public) |
| GET | `/metrics` | `promhttp.Handler()` | 없음 (public) |
| POST | `/api/v1/internal/keycloak-events` | `receiveKeycloakEventWebhook` | `X-Webhook-Secret` 헤더(fail-closed) |

### /api/v1 group (requireRequestID → authenticateActor → onboardingGate → enforceRoutePermission)
| 메서드 | 경로 | 핸들러 | RBAC resource:action / 인증 |
| --- | --- | --- | --- |
| GET | `/api/v1/me` | `getMe` | Bypass (self) |
| PATCH | `/api/v1/me` | `patchMe` | Bypass |
| POST | `/api/v1/me/onboarding` | `submitOnboarding` | Bypass |
| GET | `/api/v1/organizations/search` | `searchOrganizations` | Bypass |
| POST | `/api/v1/admin/users/:user_id/review` | `confirmUserReview` | organization:edit |
| GET | `/api/v1/dashboard/metrics` | `dashboardMetrics` | infrastructure:view |
| GET | `/api/v1/events` | `listWebhookEvents` | infrastructure:view |
| GET | `/api/v1/infra/edges` | `infraEdges` | infrastructure:view |
| GET | `/api/v1/infra/nodes` | `infraNodes` | infrastructure:view |
| GET | `/api/v1/infra/topology` | `infraTopology` | infrastructure:view |
| GET | `/api/v1/infra/services` | `listInfraServices` | infrastructure:view |
| POST | `/api/v1/infra/services/snapshot` | `ingestInfraServicesSnapshot` | Bypass (public path) + `InfraAgentToken` Bearer |
| GET | `/api/v1/infra/topology/v2` | `infraTopologyV2` | infrastructure:view |
| GET | `/api/v1/repositories` | `repositories` | pipelines:view |
| POST | `/api/v1/repositories` | `createRepositoryDraft` | application_repositories:create |
| POST | `/api/v1/repositories/:repository_id/publish` | `requestRepositoryPublish` | application_repositories:edit |
| GET | `/api/v1/issues` | `issues` | pipelines:view |
| GET | `/api/v1/pull-requests` | `pullRequests` | pipelines:view |
| GET | `/api/v1/ci-runs` | `ciRuns` | pipelines:view |
| GET | `/api/v1/ci-runs/:ci_run_id/logs` | `ciRunLogs` | pipelines:view |
| GET | `/api/v1/risks` | `risks` | security:view |
| GET | `/api/v1/risks/critical` | `criticalRisks` | security:view |
| POST | `/api/v1/risks/:risk_id/mitigations` | `createRiskMitigation` | security:create |
| GET | `/api/v1/audit-logs` | `listAuditLogs` | audit:view |
| GET | `/api/v1/rbac/policy` | `getRBACPolicyLegacyGone` | security:view (항상 410) |
| GET | `/api/v1/rbac/policies` | `listRBACPolicies` | security:view |
| POST | `/api/v1/rbac/policies` | `createRBACPolicy` | security:edit |
| PUT | `/api/v1/rbac/policies` | `updateRBACPolicies` | security:edit |
| DELETE | `/api/v1/rbac/policies/:role_id` | `deleteRBACPolicy` | security:edit |
| POST | `/api/v1/admin/service-actions` | `createServiceAction` | infrastructure:create |
| GET | `/api/v1/commands/:command_id` | `getCommand` | infrastructure:view |
| POST | `/api/v1/commands/:command_id/approve` | `approveCommand` | infrastructure:edit |
| POST | `/api/v1/commands/:command_id/reject` | `rejectCommand` | infrastructure:edit |
| GET | `/api/v1/users` | `listUsers` | organization:view |
| POST | `/api/v1/users` | `createUser` | organization:create |
| GET | `/api/v1/users/:user_id` | `getUser` | organization:view |
| PATCH | `/api/v1/users/:user_id` | `updateUser` | organization:edit |
| DELETE | `/api/v1/users/:user_id` | `deleteUser` | organization:delete |
| GET | `/api/v1/organization/hierarchy` | `getHierarchy` | organization:view |
| PUT | `/api/v1/organization/hierarchy` | `updateHierarchy` | organization:edit |
| POST | `/api/v1/organization/units` | `createOrgUnit` | organization:create |
| GET | `/api/v1/organization/units/:unit_id` | `getOrgUnit` | organization:view |
| PATCH | `/api/v1/organization/units/:unit_id` | `updateOrgUnit` | organization:edit |
| DELETE | `/api/v1/organization/units/:unit_id` | `deleteOrgUnit` | organization:delete |
| GET | `/api/v1/organization/units/:unit_id/members` | `listUnitMembers` | organization:view |
| PUT | `/api/v1/organization/units/:unit_id/members` | `replaceUnitMembers` | organization:edit |
| GET | `/api/v1/scm/providers` | `listSCMProviders` | scm_providers:view |
| PATCH | `/api/v1/scm/providers/:provider_key` | `updateSCMProvider` | scm_providers:edit |
| GET | `/api/v1/applications` | `listApplications` | applications:view |
| POST | `/api/v1/applications` | `createApplication` | applications:create |
| GET | `/api/v1/applications/:application_id` | `getApplication` | applications:view |
| PATCH | `/api/v1/applications/:application_id` | `updateApplication` | applications:edit (+ row-owner) |
| DELETE | `/api/v1/applications/:application_id` | `archiveApplication` | applications:delete (+ row-owner) |
| GET | `/api/v1/applications/:application_id/repositories` | `listApplicationRepositories` | application_repositories:view |
| POST | `/api/v1/applications/:application_id/repositories` | `createApplicationRepository` | application_repositories:create |
| DELETE | `/api/v1/applications/:application_id/repositories/*repo_key` | `deleteApplicationRepository` | application_repositories:delete |
| GET | `/api/v1/repositories/:repository_id/activity` | `repositoryActivity` | application_repositories:view |
| GET | `/api/v1/repositories/:repository_id/pull-requests` | `repositoryPullRequests` | application_repositories:view |
| GET | `/api/v1/repositories/:repository_id/build-runs` | `repositoryBuildRuns` | application_repositories:view |
| GET | `/api/v1/repositories/:repository_id/quality-snapshots` | `repositoryQualitySnapshots` | application_repositories:view |
| GET | `/api/v1/repositories/:repository_id/projects` | `listProjects` | projects:view (legacy/hybrid) |
| POST | `/api/v1/repositories/:repository_id/projects` | `createProject` | projects:create (legacy/hybrid) |
| POST | `/api/v1/projects` | `createProjectStandalone` | projects:create (v2/hybrid) |
| GET | `/api/v1/applications/:application_id/projects` | `listApplicationProjects` | projects:view (v2/hybrid) |
| POST | `/api/v1/applications/:application_id/projects` | `createApplicationProject` | projects:create (v2/hybrid) |
| GET | `/api/v1/projects/:project_id` | `getProject` | projects:view |
| PATCH | `/api/v1/projects/:project_id` | `updateProject` | projects:edit (+ row-owner) |
| DELETE | `/api/v1/projects/:project_id` | `archiveProject` | projects:delete (+ row-owner) |
| GET | `/api/v1/projects/:project_id/repositories` | `listProjectRepositories` | projects:view (v2/hybrid) |
| POST | `/api/v1/projects/:project_id/repositories` | `createProjectRepository` | projects:edit (v2/hybrid) |
| DELETE | `/api/v1/projects/:project_id/repositories/:repository_id` | `deleteProjectRepository` | projects:delete (v2/hybrid) |
| GET | `/api/v1/applications/:application_id/rollup` | `applicationRollup` | applications:view |
| GET | `/api/v1/integrations` | `listIntegrations` | applications:view |
| POST | `/api/v1/integrations` | `createIntegration` | applications:edit |
| PATCH | `/api/v1/integrations/:integration_id` | `updateIntegration` | applications:edit |
| DELETE | `/api/v1/integrations/:integration_id` | `deleteIntegration` | applications:edit |
| POST | `/api/v1/integrations/gitea/webhooks` | `receiveGiteaWebhook` | Bypass (public path) + HMAC |
| GET | `/api/v1/integration/providers` | `listIntegrationProviders` | infrastructure:view |
| POST | `/api/v1/integration/providers` | `createIntegrationProvider` | infrastructure:edit |
| PATCH | `/api/v1/integration/providers/:provider_id` | `updateIntegrationProvider` | infrastructure:edit |
| DELETE | `/api/v1/integration/providers/:provider_id` | `deleteIntegrationProvider` | infrastructure:delete |
| POST | `/api/v1/integration/providers/:provider_id/sync` | `syncIntegrationProvider` | infrastructure:edit |
| GET | `/api/v1/integration/providers/:provider_id/scm-repositories` | `listSCMRepositories` | infrastructure:view |
| POST | `/api/v1/integration/providers/:provider_id/import-repositories` | `importSCMRepositories` | infrastructure:edit |
| POST | `/api/v1/integration/providers/:provider_id/create-repository` | `createSCMRepository` | infrastructure:edit |
| POST | `/api/v1/integration/providers/:provider_id/webhook` | `ingestIntegrationProviderWebhook` | Bypass (public path) + per-provider 서명 |
| POST | `/api/v1/integration/test-connection` | `testIntegrationConnection` | infrastructure:edit |
| GET | `/api/v1/integration/bindings` | `listIntegrationBindings` | infrastructure:view |
| POST | `/api/v1/integration/bindings` | `createIntegrationBinding` | infrastructure:edit |
| PATCH | `/api/v1/integration/bindings/:binding_id` | `updateIntegrationBinding` | infrastructure:edit |
| DELETE | `/api/v1/integration/bindings/:binding_id` | `deleteIntegrationBinding` | infrastructure:delete |
| POST | `/api/v1/dev-requests` | `intakeDevRequest` | **별도 intakeGroup** (Bypass + `requireIntakeToken`) |
| GET | `/api/v1/dev-requests` | `listDevRequests` | dev_requests:view (+ row filter) |
| GET | `/api/v1/dev-requests/:dev_request_id` | `getDevRequest` | dev_requests:view (+ row-owner) |
| POST | `/api/v1/dev-requests/:dev_request_id/register` | `registerDevRequest` | dev_requests:edit (+ row-owner) |
| POST | `/api/v1/dev-requests/:dev_request_id/reject` | `rejectDevRequest` | dev_requests:edit (+ row-owner) |
| PATCH | `/api/v1/dev-requests/:dev_request_id` | `patchDevRequest` | dev_requests:edit (+ system_admin only) |
| DELETE | `/api/v1/dev-requests/:dev_request_id` | `closeDevRequest` | dev_requests:delete (+ system_admin only) |
| POST | `/api/v1/dev-request-tokens` | `createDevRequestIntakeToken` | dev_request_intake_tokens:create |
| GET | `/api/v1/dev-request-tokens` | `listDevRequestIntakeTokens` | dev_request_intake_tokens:view |
| DELETE | `/api/v1/dev-request-tokens/:token_id` | `revokeDevRequestIntakeToken` | dev_request_intake_tokens:delete |
| PATCH | `/api/v1/dev-request-tokens/:token_id` | `updateDevRequestIntakeTokenIPs` | dev_request_intake_tokens:edit |
| GET | `/api/v1/hr/lookup` | `hrLookup` | organization:view |
| POST | `/api/v1/realtime/ticket` | `issueRealtimeTicket` | Bypass (인증 actor 면 발급 가능) |
| GET | `/api/v1/realtime/ws` | `handleRealtimeWebSocket` | Bypass + ticket 인증 (RealtimeHub != nil 일 때만 등록) |

> 주: `POST /api/v1/dev-requests` 는 `intakeGroup`(별도 `router.Group("/api/v1")`)에 등록되어 v1 group 의 `authenticateActor`/`enforceRoutePermission` 가 적용되지 않는다. routePermissionTable 에는 `Bypass: true` 로만 들어 있는데, 이는 `TestRoutePermissionTable_CoversAllProtectedV1Routes` 통과용이며 실제 인증은 `requireIntakeToken` 미들웨어가 책임진다.

---

## 4. 인증/권한 미들웨어 체인 분석

### v1 group 표준 체인 (router.go:215-222)
```
v1.Use(requireRequestID)       // request_context.go
v1.Use(authenticateActor)      // auth.go
v1.Use(onboardingGate)         // onboarding_gate.go
v1.Use(enforceRoutePermission) // permissions.go
```

#### (1) `requireRequestID` (request_context.go:82)
- 인바운드 `X-Request-ID`(1..128 chars of `[A-Za-z0-9_-]`)를 검증·승계, 없거나 malformed 면 `req_<24hex>` 생성.
- gin.Context + `c.Request.Context()`(typed key `requestIDCtxKey{}`)에 저장 → 백그라운드 worker/store 로그 상관용.
- 응답 `X-Request-ID` 헤더 노출.

#### (2) `authenticateActor` (auth.go:41)
인증 변형이 한 미들웨어에 집약돼 있다:
1. **legacy 헤더 거부**: `X-Devhub-Actor` 가 있으면 400 `x_devhub_actor_removed` (ADR-0004/0006).
2. **public webhook path** (`publicAPIPaths`: gitea webhooks / integration provider webhook / infra snapshot): Bearer 없이 통과, `ctxKeySourceType` 만 set 후 `c.Next()`. (실제 인증은 각 핸들러의 HMAC/토큰 검사.)
3. **realtime WS ticket**(`/api/v1/realtime/ws` + Authorization 없음): `?ticket=` query 를 `RealtimeTickets.consume` 으로 단일 소비. store fault 는 401 이 아니라 **503**(codex PR #344 — 정상 사용자 오거부 회피). hit 시 actorLogin/role/sourceType set. **legacy `?access_token=` query fallback 은 ticket-only 컷오버(ADR-0024 §6 carve 5)로 제거됨.**
4. **빈 Authorization**: `AuthDevFallback=true` 면 통과(actor="system"), 아니면 401.
5. **Bearer 검증**: `bearerToken` 으로 scheme 파싱 → `BearerTokenVerifier.VerifyBearerToken`. verifier 미설정 시 dev fallback 또는 401.
6. **동적 role lookup**: token 의 role claim 을 신뢰하지 않고 `OrganizationStore.GetUser` 로 DB role 재조회(실시간 권한 변경 반영). DB hit 시:
   - `idp_subject` 비어 있고 token sub 있으면 `SetIdPSubject` lazy backfill (best-effort).
   - `OnboardingCompletedAt == nil` 이면 `devhub_onboarding_required=true`.
   - `ErrNotFound`(token-only actor): email/display_name claim set + `onboarding_required=true` (ADR-0021 §3.3, lazy 폐기 후 unconditional).
   - 기타 에러(schema drift): token role claim 으로 폴백 + log (e2e 회귀 마스킹 방지).

#### (3) `onboardingGate` (onboarding_gate.go:35)
- `OnboardingGateEnabled=false` → no-op.
- `devhub_onboarding_required != true` → 통과.
- `onboardingGateAllowlist`(path-only): `/api/v1/me`, `/api/v1/me/onboarding`, `/api/v1/organizations/search`, `/api/v1/organization/hierarchy` 멤버면 통과.
- 그 외 → 403 `onboarding_required` + `observeOnboardingGateBlocked` metric.

#### (4) `enforceRoutePermission` (permissions.go:375)
- `devFallbackEnabled(c)` → 통과 (핸들러 단위 테스트 보호).
- `lookupRoutePolicy(method, FullPath())` 미스 → **section 12.9 deny-by-default**: 403 `auth_policy_unmapped` + audit `auth.policy_unmapped`. (라우트를 등록하고 표 등록을 빠뜨리면 silent allow 가 아니라 visible deny.)
- `policy.Bypass` → 통과.
- 그 외: `PermissionCache.Allows(role, resource, action)`. cache 미스/deny → 403 `auth.role_denied` audit + 403 응답. cache 에러 → 500.

### `PermissionCache` (permissions.go:22)
rbac_policies matrix 를 in-memory 보관. `store==nil` 이면 `domain.SystemRoles()`(§12.1 기본 matrix) 폴백. rbac.go 의 mutation 핸들러가 `Invalidate()` 호출로 다음 요청 재로드. role 미존재 시 `(false, nil)` = deny.

### row-level 위양: `enforceRowOwnership` (permissions.go:329)
route-level RBAC 통과 후 핸들러가 호출하는 2차 가드 (ADR-0011 §4.2). 통과 규칙: `system_admin` || `allowedRoles` 멤버(예 `pmo_manager`) || `actorLogin == ownerUserID`(owner-self, ownerUserID 비어있으면 비활성). 실패 시 audit `auth.row_denied` + 403 `auth_row_denied`. `devFallbackEnabled` 면 bypass. 사용처: applications update/archive, projects update/archive, dev-requests get/register/reject.

### 인증 변형 종합
| 변형 | 적용 endpoint | 검증 메커니즘 | 위치 |
| --- | --- | --- | --- |
| OIDC Bearer | 대부분의 v1 | `BearerTokenVerifier.VerifyBearerToken` + DB role lookup | auth.go:151 |
| Realtime ticket | `/realtime/ws` | `RealtimeTickets.consume` (단일-사용, 60s TTL) | auth.go:82, realtime_ticket.go |
| Intake token | `POST /dev-requests` | SHA-256(token) lookup + revoke/expiry + IP allowlist | dev_request_intake_auth.go:64 |
| Gitea webhook HMAC | `/integrations/gitea/webhooks` | `gitea.VerifySignature`(`X-Gitea-Signature`/`X-Gogs-Signature`) | gitea_webhook.go:55 |
| Integration webhook 서명 | `/integration/providers/:id/webhook` | provider.CredentialsRef 전략별(hmac_sha256/provider_sdk/shared) | integration_registry.go:93 |
| Keycloak SPI secret | `/internal/keycloak-events` | `X-Webhook-Secret` 상수 비교(fail-closed) | keycloak_events_webhook.go:42 |
| Infra agent token | `/infra/services/snapshot` | `Authorization: Bearer <InfraAgentToken>` constant-time | infra_integrations.go:69 |
| Dev fallback | 전부 | `AuthDevFallback=true` → actor="system" 통과 | auth.go:114, commands.go:401 |

---

## 5. 도메인별 핸들러 그룹

### 5.1 me / onboarding (me.go, me_onboarding.go, organizations_search.go, users_admin_review.go, onboarding_*.go)
- `getMe`: actor 가 "" 또는 "system" 이면 401. onboarding state(`onboarding_required`/`completed_at`/`review_status`)는 gate flag 와 분리해 항상 hydrate (frontend AuthGuard 가 정확히 판단). DB miss / schema drift 모두 `onboarding_required=true` fail-safe.
- `patchMe`(API-85): `requireOnboardingFlag` gate. display_name(1~100) / primary_unit_id 변경. **primary_unit_id 변경 시 `review_status=pending_review` 자동 reset**(REQ-FR-ONBOARD-007) + `account.unit_changed` audit.
- `submitOnboarding`(API-83): `requireOnboardingFlag` gate. role 필드 미수용(fallback `developer`). `SubmitOnboarding` 단일 tx → INSERT/UPDATE. 결과별 metric(`observeOnboardingSubmit`) + duration histogram.
- `searchOrganizations`(API-84): q≥2, limit≤20. 권한 가드 없음(typeahead). 응답은 unit_id+name 만.
- `confirmUserReview`(API-86): `ConfirmUserReview` 시도 → ErrNotFound 면 `GetUser` 로 재확인해 404(없음)/422(미완료)/409(이미 reviewed) 정밀 분기.

### 5.2 applications / projects / repositories (applications.go, projects.go, repository_ops.go, application_rollup.go, domain.go)
- **Application 상태전이**(`allowedStatusTransitions`): planning→active 는 활성 repo ≥1(`CountActiveApplicationRepositories`), active→on_hold 는 hold_reason, on_hold→active 는 resume_reason, →archived 는 archived_reason, active→closed 는 critical_warning_count=0(`CountApplicationCriticalWarnings`) 가드. key immutable(PATCH 거부), key 패턴 `^[A-Za-z0-9]{1,10}$`. update/archive 는 `enforceRowOwnership(..., pmo_manager)`.
- **Project 모드 게이트**(`projectModel`: legacy|hybrid|v2, default hybrid): repository-centric 라우트는 `allowLegacyProjectRoutes`, application-centric/standalone/project-repo 라우트는 `allowV2ProjectRoutes`. 비활성 시 **410 gone**. standalone/application-centric 생성은 `CreateProjectWithRepositoryPayload`(repo 동반 생성 단일 tx, codex #349 P2).
- **Repository 운영지표**(API-51..54): read-only. repository_id int 파싱, RFC3339 window, limit 1..200.
- **Application rollup**(API-57): weight_policy(equal/repo_role/custom), custom_weights JSON. store 의 "invalid weight policy" 에러는 `containsAny` 로 감지해 422.
- **Repository draft→publish**(domain.go): `createRepositoryDraft`(provider_key → provider_id FK 해석, migration 000045), `requestRepositoryPublish`(draft only → provider SCM/push/gitea-compat 검사 → `gitea.CreateRepo` → `UpsertRepository`). DomainStore 가 `repositoryDraftStore` 인터페이스 미충족 시 503.

### 5.3 integration registry / scm (integration_registry.go, integration_scm_repositories.go, integrations.go)
- **provider CRUD**(API-69..71, 80): provider_type(alm/scm/ci_cd/doc/infra), auth_mode(token/basic/oauth2/app_password/agent). base_url/auth_token_url 은 `validBaseURL`(http(s)+host). `api_token`/`auth_secret` 은 write-only(응답엔 `*_set` bool). DELETE 는 binding 존재 시 409 FK guard.
- **sync**(API-72): SCM type 만(fast-fail, codex #345 P2), pull|sync capability gate. provider disabled → 409.
- **webhook ingest**(API-73): `X-Integration-*`→`X-Gitea-*`→`X-Gogs-*` fallback. 서명 검증 후 dedupe SaveWebhookEvent + sync state best-effort.
- **test-connection**(API-87): GET + 5s timeout + redirect 미추적. **SSRF 의도적 수용**(사내 internal endpoint 합법).
- **bindings CRUD**(API-74/75/81/82): scope(application/project), policy(summary_only/execution_system).
- **SCM repo 연동**: list(API-88, imported 플래그)/import(API-89, SCM 재조회 값으로 upsert source=scm)/create(API-90, push capability, source=system). 공통 게이트 `scmProviderForCapability`: exists + enabled(codex #366 P2) + scm type + capability + gitea-compat(`isGiteaCompatibleProvider`, codex #363 P2).
- **project_integrations**(API-58): scope-target 정합(application↔application_id), type jira/confluence.

### 5.4 dev-requests (dev_requests.go, dev_request_intake_tokens_admin.go)
- `intakeDevRequest`(API-59): source_system 은 body 무시·intake token 매핑값 사용(spoofing 방지). (source_system, external_ref) idempotency. 검증 실패 시 `rejected`(invalid_intake) row 저장(audit 보존). assignee FK violation 시 NULL assignee + rejected 재저장.
- `listDevRequests`(API-60): system_admin/pmo_manager 외엔 본인 assignee 만(row filter).
- `registerDevRequest`(API-62, **promote**): target_id(legacy) | application_payload | project_payload 중 정확히 1개. 신규 생성은 store 의 단일-tx(`RegisterDevRequestWithNew*`). status pending/in_review 만 허용. promote 시 SCM provider enabled gate(codex #4 P2) + repo role CHECK 가드(codex #4 P1).
- `rejectDevRequest`(API-63)/`patchDevRequest`(API-64, reassign, **system_admin only**)/`closeDevRequest`(API-65, **system_admin only**). 전이는 `domain.IsValidDevRequestTransition`.
- intake token admin(dev_request_intake_tokens_admin.go): 발급 시 plain token **1회 노출**, store 는 `hashIntakeToken`(SHA-256). allowed_ips deny-by-default(빈 리스트 거부). PATCH 는 raw map 파싱으로 allowed_ips/expires_at 부분 갱신 구분.

### 5.5 organization / users (organization.go)
- users CRUD: role(`validAppRoles`: developer/manager/pmo_manager/system_admin), status, type. Keycloak identity write 는 제거(ADR-0020 E) — DevHub `users` row 만.
- org-units CRUD + hierarchy: `updateOrgUnit` 의 parent 변경 시 **cycle detection**(GetHierarchy edges ancestor chain, RM-M3-03) → 422 `cycle_detected`.
- 모든 mutation 이 `recordAuditBestEffort` + `addAuditMeta`(audit_log_id 응답 노출).

### 5.6 rbac (rbac.go)
- `getRBACPolicyLegacyGone`: 항상 410(ADR-0002, GET /rbac/policy 폐기).
- list/create/bulk-update/delete: system role 보호(metadata immutable, 삭제 불가, id 예약), audit invariant(audit resource 에 create/edit/delete 금지) 핸들러+store 이중 검증. bulk-update 는 **전체 검증 후 일괄 write**(부분 변경 방지). 성공 시 `PermissionCache.Invalidate()`.

### 5.7 audit (audit.go)
- `recordAudit`: actor/source(actor_source) + request-scoped(source_ip via `clientIPFrom`, request_id, source_type) stamp. `recordAuditBestEffort` 는 에러 swallow(주 mutation 이미 commit → 500 → 중복 retry 회피).

### 5.8 realtime (realtime.go, realtime_ticket.go)
- `handleRealtimeWebSocket`: types query 필수, 각 event type 을 `realtimeEventPermission` → `PermissionCache.Allows` 로 검사(미인증 role 거부). `RealtimeHub` 가 subscribe/publish 관리.
- ticket: in-memory(`RealtimeTicketStore`, single-instance) + PG(`DBRealtimeTicketStore`, multi-instance, DELETE...RETURNING 원자성). `issueRealtimeTicket` 은 인증 actor 면 발급(RBAC bypass).

### 5.9 infra (infra_integrations.go, snapshot.go, snapshot_provider.go, runtime_snapshot_provider.go)
- snapshot ingest 는 `InfraAgentToken` Bearer constant-time. HomeLab adapter 가 있으면 정규화·persist, 없으면 best-effort persist. runtime 상태는 process-global `runtimeInfraSnapshots`(sync.RWMutex).
- dashboard/infra/ci/risk read 는 DB(`DomainStore`) → 없으면 `StaticSnapshotProvider`(하드코딩 mock) 폴백. `RuntimeSnapshotProvider` 는 static 위에 live health probe.

---

## 6. API ID 매핑 (이 패키지가 구현하는 것)

| API ID | endpoint(들) | 핸들러 |
| --- | --- | --- |
| API-32 | GET /me | `getMe` |
| API-41/42 | GET/PATCH /scm/providers | `listSCMProviders`/`updateSCMProvider` |
| API-43..47 | applications CRUD | `listApplications`/`createApplication`/`getApplication`/`updateApplication`/`archiveApplication` |
| API-48..50 | application repositories | `listApplicationRepositories`/`createApplicationRepository`/`deleteApplicationRepository` |
| API-51..54 | repository 운영지표 | `repositoryActivity`/`repositoryPullRequests`/`repositoryBuildRuns`/`repositoryQualitySnapshots` |
| API-55/56 | project CRUD | `listProjects`/`createProject`/`getProject`/`updateProject`/`archiveProject` (+ v2 변형) |
| API-57 | application rollup | `applicationRollup` |
| API-58 | project_integrations CRUD | `listIntegrations`/`createIntegration`/`updateIntegration`/`deleteIntegration` |
| API-59..65 | dev-requests | `intakeDevRequest`/`listDevRequests`/`getDevRequest`/`registerDevRequest`/`rejectDevRequest`/`patchDevRequest`/`closeDevRequest` |
| API-66..68 | dev-request-tokens admin | `createDevRequestIntakeToken`/`listDevRequestIntakeTokens`/`revokeDevRequestIntakeToken` (+ PATCH update) |
| API-69..75 | integration providers/bindings | `listIntegrationProviders`/`createIntegrationProvider`/`updateIntegrationProvider`/`syncIntegrationProvider`/`ingestIntegrationProviderWebhook`/`listIntegrationBindings`/`createIntegrationBinding` |
| API-80 | DELETE integration provider | `deleteIntegrationProvider` |
| API-81/82 | PATCH/DELETE binding | `updateIntegrationBinding`/`deleteIntegrationBinding` |
| API-83..86 | onboarding | `submitOnboarding`/`searchOrganizations`/`patchMe`/`confirmUserReview` |
| API-87 | test-connection | `testIntegrationConnection` |
| API-88..90 | SCM repo list/import/create | `listSCMRepositories`/`importSCMRepositories`/`createSCMRepository` |
| (legacy) | rbac/commands/audit/infra/risks/ci | rbac.go / commands.go / audit.go / snapshot.go / domain.go (M0~M3 contract) |

> repository draft→publish(`POST /repositories`, `POST /repositories/:id/publish`)는 #368/#373 으로 도입됐으나 본 분석 시점 API ID 표기(예 API-91/92)가 코드 주석에 없다 — contract 미발급 추정(아래 발견사항 참조).

---

## 발견 사항 (불일치 / stale / 부채)

### F-1. 무테스트로 머지된 repository draft→publish 핸들러 (HIGH)
`createRepositoryDraft`/`requestRepositoryPublish`(`domain.go:141-287`)와 그 의존 store 인터페이스 `repositoryDraftStore`(`domain.go:35-39`: `CreateRepositoryDraft`/`MarkRepositoryDraftPublishRequested`/`GetRepositoryByID`)에 대한 단위/통합 테스트가 **전무**하다. `*_test.go` 전체에서 `createRepositoryDraft`/`requestRepositoryPublish`/`CreateRepositoryDraft`/`MarkRepositoryDraftPublishRequested` 식별자 검색 결과 0건. #368(draft→publish lifecycle, codex 머지)이 무테스트로 들어왔고 #373(provider_id 단일화)이 그 위를 수정. `requestRepositoryPublish` 는 SCM 생성 실패 시 `MarkRepositoryDraftPublishRequested` 만 호출하고 BadGateway 반환하는 부분 실패 경로(`domain.go:260-263`)가 있어 검증 공백이 특히 위험.
- 근거: `backend-core/internal/httpapi/domain.go:141-287`; 테스트 부재(`*_test.go` grep 0건).

### F-2. credentials_ref / 평문 secret 노출 경로 (HIGH, 알려진 gap)
`integrationProviderResponse`(`integration_registry.go:36`)가 `"credentials_ref": p.CredentialsRef` 를 **raw 그대로 응답에 노출**한다. credentials_ref 는 webhook 서명 검증용 시크릿(`hmac_sha256:<secret>` / `provider_sdk:<vendor>:<secret>` / shared token)을 담을 수 있어, GET `/integration/providers` 권한(infrastructure:view)을 가진 사용자에게 시크릿이 그대로 노출된다. `api_token`/`auth_secret` 은 write-only(`api_token_set`/`auth_secret_set` bool)로 보호되는데 `credentials_ref` 만 예외. 메모리 노트의 "#6 평문 secret 저장 carve" 가 이 gap.
- 근거: `backend-core/internal/httpapi/integration_registry.go:36` (`"credentials_ref": p.CredentialsRef`), 대비 `:42`/`:48` 의 `api_token_set`/`auth_secret_set`.

### F-3. SSRF 의도적 수용 (MEDIUM, 문서화된 결정)
`testIntegrationConnection`(`integration_registry.go:430`)이 임의 `base_url` 로 서버측 GET 을 보낸다. 주석(`:425-429`)이 명시하듯 사내 internal endpoint(Gitea/Jenkins)가 합법 대상이라 internal IP 차단을 하지 않음 — admin 신뢰 경계 + 5s timeout + 응답 본문 미반환으로 표면 최소화. 동일하게 `requestRepositoryPublish`/import/create 의 `scmProviderClient` 도 provider.BaseURL 로 outbound. infrastructure:edit 권한자에게만 노출되지만 SSRF 표면이 존재함을 명시.
- 근거: `integration_registry.go:425-460`; `integration_scm_repositories.go:88-103`.

### F-4. dead / 미연결 코드 (MEDIUM)
- **`requireMinRole`(authz.go:12)**: 어떤 라우트에도 wire 되지 않음. `router.go` 내 유일 언급은 주석(`router.go:162`)뿐. 실제 권한은 `enforceRoutePermission` 이 담당. 테스트(`authz_test.go`)만 사용.
- **`resolveIdPSubject`(identity_resolver.go:19)**: 어떤 핸들러도 호출하지 않음. grep 결과 production 호출처 0, `identity_resolver_test.go` 만 사용. ADR-0020 으로 account-admin write 경로가 제거되며 호출처가 사라진 잔재.
- 근거: `authz.go:12`; grep `requireMinRole` → router.go(주석만)+authz.go; `identity_resolver.go:19`; grep `resolveIdPSubject` → 테스트만.

### F-5. onboardingGate allowlist 가 method-agnostic (MEDIUM)
`onboardingGate`(`onboarding_gate.go:47`)는 `c.FullPath()`(method 무관 path 패턴)로 allowlist 를 검사한다. `/api/v1/me` 가 allowlist 멤버라 **GET /me 뿐 아니라 PATCH /me 도 미완료 사용자에게 gate 를 통과**시킨다. `patchMe`(me.go:137)는 onboarding 완료 여부를 검사하지 않고 `requireOnboardingFlag`(flag on/off)만 본다. 즉 onboarding 미완료 사용자가 `submitOnboarding` 을 거치지 않고 PATCH /me 로 display_name/primary_unit_id 를 바꿀 수 있는 경로가 열려 있다. 주석(`onboarding_gate.go:29`)은 GET 용 "API-32 onboarding_required flag 응답"으로만 의도를 적었으나 path 키라 PATCH 도 포함됨.
- 근거: `onboarding_gate.go:28-33, 47`; `me.go:137-156`(완료 검사 없음).

### F-6. `requireOnboardingFlag` 의 flag 의미가 코드와 운영 default 불일치 (MEDIUM)
`requireOnboardingFlag`(`onboarding_feature_flag.go:19`)는 `OnboardingGateEnabled` 가 false 면 onboarding endpoint(patchMe/submitOnboarding/searchOrganizations/confirmUserReview)를 **404 `onboarding_feature_disabled`** 로 막는다. 그러나 동일 flag(`OnboardingGateEnabled`)는 onboardingGate 의 차단 동작도 제어한다(router.go:164-173 주석은 default **true**). 즉 단일 flag 가 "endpoint 활성화"와 "gate 차단"이라는 두 책임을 겸한다. flag 를 rollback(=false)으로 내리면 gate 만 끄려 했는데 onboarding endpoint 자체가 404 가 되어버린다(코드 주석은 default true 라 정상 운영엔 영향 없으나, rollback 경로의 의미가 과도하게 결합).
- 근거: `onboarding_feature_flag.go:19-29`; `router.go:164-173`; `onboarding_gate.go:36-39`.

### F-7. project standalone 의 `scm_provider` placeholder 필드 (LOW, 문서화된 부채)
`createRepositoryPayload`(`projects.go:132-136`)의 `SCMProvider` 필드와 store 의 `RepositoryCreatePayload.SCMProvider` 는 #373 메모리 노트상 "placeholder(별개라 유지)"로 남아있다. draft→publish 흐름의 provider_id FK 해석(`domain.go`)과 의미 중복 소지. 현재 `createProjectStandalone`/`createApplicationProject` 가 이 payload 를 그대로 store 에 전달(`projects.go:281-291, 405-415`).
- 근거: `projects.go:132-136`; 메모리 노트 post-#373.

### F-8. 라우트 ↔ permissions 표 일치성 (정상, 단 주의점 1건)
`router.go` 의 모든 v1 라우트가 `routePermissionTable` 에 대응 엔트리를 가지며(`TestRoutePermissionTable_CoversAllProtectedV1Routes`, permissions_test.go 가 가드), 미스 시 deny-by-default(403 `auth_policy_unmapped`)로 fail-loud 처리되어 silent 불일치는 구조적으로 차단됨. **단 주의**: `POST /api/v1/dev-requests` 는 intakeGroup 소속이라 표상 `Bypass: true` 지만 실제 v1 미들웨어를 안 탄다 — 표 엔트리는 순전히 coverage 테스트 통과용(permissions.go:275-280 주석 명시). 인증 누락이 아니라 `requireIntakeToken` 이 책임지므로 정상이나, 표만 보면 오해 소지.
- 근거: `permissions.go:280`; `router.go:332-334`.

### F-9. env 자격증명 fallback 금지 정책 (MEDIUM, 정상 동작이나 표면 명시)
`scmProviderClient`(`integration_scm_repositories.go:88`)는 `provider.ResolveOutboundAuth()` 만 사용하고 worker-global env 토큰 fallback 을 쓰지 않는다(codex #358 P1 / #359 — 잘못된 계정 토큰 유출 방지). 미설정 시 422 `integration_outbound_credentials_missing`. 다만 `scmProviderResponse`(`applications.go:76-90`)의 `has_credentials` 는 여전히 `gitea` provider_key 한정으로 `GITEA_URL`/`GITEA_TOKEN` env 존재 여부를 본다 — SCM provider(legacy) catalog 와 integration provider 의 자격증명 모델이 이원화돼 있음(혼동 소지).
- 근거: `integration_scm_repositories.go:88-103`; `applications.go:78-80`.

### F-10. stale 주석 / 진화 흔적 (LOW)
- `router.go:218-220`: onboardingGate 주석이 "Feature flag default OFF (no-op)" 라고 적었으나, `RouterConfig.OnboardingGateEnabled` doc(`:164-167`)은 default **true**(2026-05-21 lazy 폐기 이후). 같은 파일 내 default 기술이 상충.
- `applications.go:18-25`: "Handler bodies are 501 stubs" / "501 stubs ... store body 는 후속 sprint carve out"(`router.go:275-276`) 주석은 이미 정식 구현 완료된 현재 stale.
- `dev_requests.go:498`: 에러 메시지가 `^[A-Za-z0-9]{10}$`(정확히 10자) 라고 하는데 실제 검증은 `applicationKeyPattern`=`^[A-Za-z0-9]{1,10}$`(1~10자, applications.go:27). 사용자에게 보이는 에러 텍스트가 실제 규칙과 불일치.
- 근거: `router.go:218-220` vs `:164-167`; `applications.go:18`, `router.go:275-276`; `dev_requests.go:499` vs `applications.go:27`.

### F-11. realtime WS CheckOrigin 무조건 허용 (LOW)
`websocketUpgrader.CheckOrigin`(`realtime.go:173-177`)이 `return true` 로 모든 origin 을 허용. ticket 인증으로 보호되긴 하나 CSWSH(cross-site WebSocket hijacking) 표면이 origin 검증 부재로 남아있음.
- 근거: `realtime.go:173-177`.

### F-12. infra snapshot 상태가 process-global 가변 전역 (LOW)
`runtimeInfraSnapshots`(`infra_integrations.go:52`)는 package-level `var` (sync.RWMutex 보호). multi-instance 배포 시 인스턴스별로 상태가 갈리고(어느 인스턴스가 ingest 를 받았는지에 따라 topology-v2 응답이 달라짐), 테스트 간 상태 누수 위험. ticket store 는 PG 백킹으로 multi-instance 안전화했으나 infra snapshot 은 미적용.
- 근거: `infra_integrations.go:44-52, 140-148`.

---

## 부록: 핸들러 receiver 혼용 (관찰)
일부 핸들러는 `func (h Handler)`(값 receiver), 일부는 `func (h *Handler)`(포인터 receiver)로 정의돼 있다(예: `domain.go` 는 값, `applications.go`/`projects.go`/`integration_*.go` 는 포인터). `Handler` 가 `cfg RouterConfig` 단일 필드라 동작 차이는 없으나(불변), 코드 일관성 부채. gin 의 `MethodExpr` 등록은 양쪽 모두 정상.
