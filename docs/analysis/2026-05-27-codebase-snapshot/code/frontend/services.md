# Frontend 분석 — Services (lib/services)

- 문서 목적: `frontend/lib/services/` 의 18 service + types + api-client/realtime/websocket 레이어를 백엔드 API 호출·메서드·에러 처리 기준으로 인벤토리화한다.
- 범위: `frontend/lib/services/`. wire/UI 경계 규약(`wire.ts` ↔ `types.ts`): service 가 snake_case wire shape 를 UI shape 으로 변환, UI 컴포넌트는 `types.ts` 만 import.
- 최종 작성일: 2026-05-27 (main `cf19c94`)
- 관련 문서: `pages.md`, `components.md`, `frontend_platform.md`

## 1. 전송 코어 — api-client.ts

`apiClient<T>(method, path, body)` 가 모든 service 의 표준 fetch wrapper.

- **Bearer 주입**: `tokenStore.getAccessToken()` 이 있으면 `Authorization: Bearer`.
- **basePath 해석**: `path` 가 `/api/` 로 시작하면 `${API_BASE_URL}${path}` (reverse-proxy/basePath 대응), 절대 URL 은 그대로.
- **401 refresh-then-retry (ADR-0024 §6 carve 3)**: 응답이 401 이고 (access token OR refresh token 존재) 시 `attemptTokenRefresh()` 1회 → 성공하면 새 토큰으로 재요청. hard-refresh 직후 access header 가 비어도 refresh token 으로 복구.
- **parallel refresh storm 가드**: `inflightRefresh` 단일 promise 공유 — N개 동시 401 이 IdP 토큰 endpoint 를 1회만 친다.
- **token endpoint 해석**: `OIDC_ISSUER_URL` (build-time) 우선, 없으면 `/api/runtime-config` (Docker runtime) fetch → `cachedTokenEndpoint` 캐시.
- **에러 표면화**: non-2xx 시 `ApiError(status, payload, message)` throw. message 는 응답 body 의 `error` 필드 우선, 없으면 `HTTP {status}`.

`error-message.ts` `toUserErrorMessage(error, fallback)`: ApiError status별 한국어 표준 메시지(500/404/401·403) + 그 외 message/fallback. 페이지·모달이 toast/배너 문안 통일에 사용.

## 2. Service 인벤토리

| Service | 인스턴스 | 주요 메서드 → 백엔드 API | 에러 처리 |
|---|---|---|---|
| `auth.service.ts` | singleton | `getAuthorizeURL`(PKCE+discovery), `exchangeCode`→token_endpoint, `refreshTokens`, `logout`(end_session), `resolveIdentity`→whoAmI, `getAccountConsoleURL` | discovery fetch 실패 시 issuer-derived endpoint fallback; runtime-config fetch 실패 시 env fallback; resolveIdentity 실패 시 logout() |
| `identity.service.ts` | singleton | `whoAmI`→`GET /me`, `getUsers`→`/users`, CRUD `/users`·`/organization/units`, `getOrgHierarchy`/`updateOrgHierarchy`→`/organization/hierarchy`, `getUnitMembers`/`replaceUnitMembers`, `lookupHR`→`/hr/lookup` | apiClient 위임. `result.data` 누락 시 `ApiError(500)`. role wire↔UI 매핑(ROLE_BACKEND_TO_UI 기본 Developer) |
| `application.service.ts` | new | `listApplications`(status/q/include_archived), `getApplication`, `getApplicationRollup`→`/applications/{id}/rollup` | apiClient 위임 (throw 전파) |
| `project.service.ts` | new | SCM providers, applications CRUD, `/applications/{id}/repositories` connect/disconnect, `/repositories/{id}/projects`, `createProjectStandalone`→`POST /projects`, `createApplicationProject`(v2 primary + legacy fallback), project tasks/activity 정규화 | `createApplicationProject` 404/405 시 legacy `createProject` fallback; `listAllProjects`/`getApplicationProjects` per-repo catch→[]; activity/tasks 방어적 normalize |
| `repository.service.ts` | new | `listRepositories`, `getRepository`(목록 find), `createRepositoryDraft`→`POST /repositories`, `requestRepositoryPublish`→`/repositories/{id}/publish`, `getRepositoryActivity`, `getRepositoryBuildRuns` | apiClient 위임. draft/publish 는 최신 #368/#373 기능 |
| `dev_request.service.ts` | new | `list`(status 배열 join), `get`, `register`→`/{id}/register`, `reject`, `reassign`(PATCH), `close`(DELETE), `getMyPending` | apiClient 위임 |
| `dev_request_token.service.ts` | new | `list`, `issue`(plain 1회), `revoke`(DELETE), `update`(PATCH allowed_ips 등), `updateIPs` | apiClient 위임 |
| `integration.service.ts` | new | providers CRUD, `syncProvider`→`/{id}/sync`, `testConnection`→`/test-connection`, `listScmRepositories`(API-88), `importScmRepositories`(API-89), `createScmRepository`(API-90), bindings CRUD(API-74/75/80) | 409 code(`integration_provider_has_bindings`/`integration_binding_conflict`)·422(`integration_policy_violation`)는 caller 가 `ApiError.payload.code` 로 분기 |
| `infra.service.ts` | singleton | `getMetrics`→`/dashboard/metrics?role=`, `getNodes`, `getTopology`, `getTopologyV2`(API-76/78), `controlService`→`/admin/service-actions`(dry_run) | **getNodes/getTopology 실패 시 mock fallback** (하드코딩 노드 배열). getMetrics/getTopologyV2 는 throw |
| `audit.service.ts` | singleton | `getLogs`→`/audit-logs` (필터 + limit/offset, meta 반환) | apiClient 위임. data/meta null-safe |
| `rbac.service.ts` | new | `listPolicies`/`createPolicy`/`updatePolicies`/`deletePolicy`→`/rbac/policies` | **`RbacError`(status+code)로 래핑** — contract §12 error code(role_in_use 등) 보존 |
| `onboarding.service.ts` | singleton | `submit`→`/me/onboarding`, `patchMe`→`PATCH /me`, `searchOrganizations`→`/organizations/search`, `confirmUserReview`→`/admin/users/{id}/review` | apiClient 위임, data 누락 시 ApiError(500) |
| `gardener.service.ts` | singleton | `getSuggestions`→`/gardener/suggestions`, `applySuggestion`→`/{id}/apply` | **getSuggestions 실패 시 mock fallback** (Phase 4 prototyping 주석). applySuggestion throw |
| `risk.service.ts` | singleton | `getCriticalRisks`→`/risks/critical`, `applyMitigation`→`/risks/{id}/mitigations` | **raw `fetch` 직접 사용** (apiClient 미경유 → Bearer/401 refresh 없음). 실패 시 mock fallback. `X-Devhub-Actor: yklee` **하드코딩** |
| `dashboard.service.ts` | new | developer stream/builds, manager velocity/team-load/decisions→`/dashboard/...` | 방어적 normalize(다양한 키 fallback). 호출 페이지가 모두 아카이브됨 → **dead-path** |
| `integration-provider-presets.ts` | (순수함수) | `composeCredentialsRef`/`parseCredentialsRef`/`getVendorPreset` + VENDOR_PRESETS(7 vendor) | 순수 함수, vitest 14 케이스 |

## 3. 타입 레이어

- `wire.ts` — 백엔드 JSON envelope(`ApiResponse<T>`{status,data,meta}, `ApiErrorResponse`, `WSEvent`(5 key), `ApiUserRole`). UI 미참조 규약.
- `types.ts` — UI shape(`UserRole` 라벨, `Metric`, `Risk`, `ServiceNode/Edge`, `ServiceActionCommand`, RBAC shape). wire envelope 를 re-export 하여 legacy import 호환.
- 도메인 types: `dev_request.types.ts`, `dev_request_token.types.ts`, `integration.types.ts`(auth_mode 5종 + write-only secret 필드), `project.types.ts`, `rbac.types.ts`(defaultRoles), `audit.types.ts`.

## 4. 실시간 레이어 — realtime vs websocket (중복)

두 WS 구현이 공존한다.

| | `realtime.service.ts` (신규, ADR-0024) | `websocket.service.ts` (레거시) |
|---|---|---|
| 인증 | **ticket pattern** — `POST /realtime/ticket` (single-use, 60s TTL) → `&ticket=` query. 401 시 `authService.refreshTokens()` 후 재발급 | **`?access_token=` query** (URL/log 토큰 노출 위협) |
| reconnect | identity 변화 구독, code≠1000 시 재시도(max 5, 3s) | exponential backoff(max 5) |
| 소비처 | **Header.tsx**(`status.changed`, `dev_request.created`), **admin/topology-v2**(`infra.node.updated`, `infra.service.updated`) | **AuthGuard.tsx**(`notification.created`, `risk.critical.created`) |
| 특이 | `DEFAULT_EVENT_TYPES` 에 `command.status.updated` 포함하나 소비처 없음 | `startMockEvents()` 데드코드(10s mock 이벤트 generator, 주석 처리되어 호출 안 됨) |

ADR-0024 §6 carve 5 가 ticket-only 컷오버를 명시했으나 **AuthGuard 는 아직 레거시 websocketService 사용** → `?access_token=` 노출 경로가 완전히 제거되지 않음.

## 발견 사항 (불일치/stale/부채)

- **단위테스트 밀도 낮음**: 18 service 중 vitest 가 있는 건 `project.service.test.ts`(3 케이스) + `integration-provider-presets.test.ts`(순수함수 14)뿐. auth.service(PKCE/discovery/refresh fallback), identity.service(role 매핑/방어 normalize), integration.service(409/422 code 분기) 등 분기 많은 service 는 무테스트 → 회귀 위험.
- **legacy websocket.service ↔ realtime.service 중복**: ADR-0024 ticket-only 컷오버 의도에도 `AuthGuard.tsx:96` 가 여전히 `websocketService.connect()` 사용 — `?access_token=` query(`websocket.service.ts:50`)가 살아있음. 두 reconnect 로직·구독 dispatch 가 별개로 유지되는 중복 부채.
- **websocket.service 데드코드**: `startMockEvents()`(`websocket.service.ts:140`)는 어디서도 호출 안 됨(onopen 주석 처리 `:58`). `mockTimer` 관련 코드 전체가 죽은 Phase 3 verification 잔재.
- **risk.service 가 apiClient 미경유**: `risk.service.ts:19,58` 이 raw `fetch` 직접 사용 → Bearer 주입·401 refresh-retry·basePath 해석 모두 누락. `X-Devhub-Actor: 'yklee'` 하드코딩(`risk.service.ts:64`, 주석 "Phase 4" / "Hardcoded for now"). 호출 페이지(manager)가 아카이브돼 현재 미사용이나 활성화 시 인증 깨짐.
- **mock fallback 잔재**: `infra.service.getNodes/getTopology`(`infra.service.ts:46,71`), `gardener.service.getSuggestions`(`gardener.service.ts:38`), `risk.service.getCriticalRisks`(`risk.service.ts:36`)가 에러 시 하드코딩 mock 데이터 반환 → 백엔드 장애를 정상 데이터로 가린다(운영 UI 전환 원칙 위배). archive(`lib/archive/mock-ui-legacy.ts`)는 분리됐으나 service 내부 fallback 은 미정리.
- **dashboard.service 전체 dead-path**: developer/manager dashboard 페이지 아카이브 후 `dashboard.service.ts`(5 메서드)를 호출하는 곳이 없다. 코드만 잔존.
- **command.status WS UI 미완 (Phase 4 잔여)**: `realtime.service.ts:11` 가 `command.status.updated` 를 구독 대상에 넣었으나 소비처 0. service-action 비동기 진행 표시 미구현.
- **Header 의 stale 구독**: `Header.tsx:81` 가 `dev_request.created` 구독하나 이 타입은 `DEFAULT_EVENT_TYPES`(realtime.service.ts:10-16)에 없어 백엔드가 push 하지 않음 → 영구 미발화 구독.
- **최신 SCM 기능 happy-path E2E 후행**: integration.service 의 `importScmRepositories`(API-89)/`createScmRepository`(API-90)/repository draft·publish 는 무테스트(서비스·E2E 모두). admin-integrations.spec.ts 는 provider lifecycle 만 cover.
