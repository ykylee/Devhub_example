# ADR-0033: Multi-Provider Webhook Architecture (X-2, v0.1.1 sprint)

- **문서 목적**: X-2 inbound webhook 정규화 깊이 (multi-provider sync 일반화) 의 **정책 결정 + architecture** 을 단일 ADR 로 명문화한다. N-13 의 InboundSourceRoutingConfig (1차 PR #586) + WebhookAdapter interface (2차 PR #587) + Jira/Generic adapter (3차 PR #588) + openapi 정합 (4차 PR #589) + frontend 운영 UI (5차 PR #590) 의 5 chunk 정공법 종합.
- **범위**: backend 의 multi-provider webhook dispatch architecture + frontend 의 운영 UI + openapi 정합 + e2e 검증 + supersession trigger.
- **대상 독자**: Backend / Frontend 개발자, AI agent, QA, 운영자, product owner.
- **상태**: accepted (2026-06-13, sprint `feat/work_260614-x2-frontend-e2e` 의 5차 PR #590, v0.1.1 milestone X-2)
- **최종 수정일**: 2026-06-13
- **결정 근거 sprint**: `feat/work_260614-x2-system-admin-dashboard` (X-1 sprint 의 multi-provider 일반화 선결정) + `feat/work_260612-7-v0-1-1-inbound-source-impl` (N-13 의 backend foundation + PR A-2) + `feat/work_260614-x2-frontend-e2e` (X-2 의 1~5차 sprint 종합)
- **Tier**: **사외** (backend code + frontend code + openapi + ADR + e2e, 사내 한정 정보 미포함)
- **관련 문서**: [release_v0-1_roadmap.md §3.5 X-2](../planning/release_v0-1_roadmap.md) (X-2 의 scope), [system_admin_catalog_plan_2026-05-27.md](../planning/system_admin_catalog_plan_2026-05-27.md) (system_admin 운영 동선), [2026-06-12-inbound-source-routing-sprint-plan.md](../planning/2026-06-12-inbound-source-routing-sprint-plan.md) (N-13 의 inbound_source routing 1차 출처), [ADR-0028 §6 (a)](./0028-dev-requests-voc-external-ref.md) (voc → dev-request 자동 routing 의 source ADR), [PR #586](https://github.com/ykylee/Devhub_example/pull/586) (1차 PR backend depth), [PR #587](https://github.com/ykylee/Devhub_example/pull/587) (2차 PR WebhookAdapter + Gitea adapter), [PR #588](https://github.com/ykylee/Devhub_example/pull/588) (3차 PR Jira + Generic adapter), [PR #589](https://github.com/ykylee/Devhub_example/pull/589) (4차 PR openapi 정합).

## 1. 배경

### 1.1 N-13 의 foundation + X-1 의 system_admin 운영 가시성

- **N-13 backend foundation** (1차 출처: `2026-06-12-inbound-source-routing-sprint-plan.md`):
  - migration 000007 (applications.inbound_source_type + inbound_source_config JSONB 컬럼 추가)
  - `applications.inbound_source_type` enum (gitea | jira | other | "")
  - `applications.inbound_source_config` JSONB-serialized text
  - 1차 PR #586 의 `auto_route.go` 의 InboundSourceRoutingConfig struct + Gitea/jira/github/gitlab pattern matcher
- **X-1** (1차 PR #583 + 2차 PR #584) — system_admin 운영 대시보드 (Gitea sync job 큐/상태 + provider health 의 운영 view)
- **현시점 문제**: webhook ingest 가 Gitea-only (`gitea_webhook.go` 의 legacy + `IngestIntegrationProviderWebhook` 의 newer Gitea-agnostic foundation) — multi-provider 일반화 (Jira, GitHub, GitLab, custom) 의 architecture 미정

### 1.2 본 ADR 의 정공법 (X-2 의 5 chunk 종합)

| Chunk | PR | 정공법 |
|---|---|---|
| 1차 | PR #586 | backend `auto_route.go` multi-provider pattern matcher depth (Gitea/jira/github/gitlab regex + InboundSourceRoutingConfig) |
| 2차 | PR #587 | `internal/infrastructure/webhook/` 디렉터리 신규 + WebhookAdapter interface + Gitea adapter + adapter registry dispatcher |
| 3차 | PR #588 | JiraWebhookAdapter (provider_type='alm') + GenericWebhookAdapter (provider_type='other') + init() 등록 dispatcher |
| 4차 | PR #589 | openapi.yaml 의 X-2 schema 정합 (WebhookEvent + WebhookAdapterType + InboundSourceRoutingConfig) |
| 5차 | PR #590 (본 sprint) | frontend multi-provider 운영 UI + e2e + 본 ADR + traceability/CHANGELOG/mirror-list |

## 2. 결정 (X-2 의 5 chunk)

### 2.1 backend 1차 (PR #586): multi-provider pattern matcher depth

`auto_route.go` 의 case 1 을 multi-provider 일반화:
- **Gitea provider-specific**: `^GITEA-\d+$` (GiteaWebhookAdapter 의 external_ref 와 1:1 매핑)
- **Jira provider-specific**: `^([A-Z][A-Z0-9_]{1,9})-\d+$` (Jira issue key 형식)
- **GitHub provider-specific**: `^#\d+$` (GitHub PR/issue number)
- **GitLab provider-specific**: `^!\d+$` (GitLab MR/issue number)
- **Custom (other) provider**: `InboundSourceRoutingConfig.CustomExternalRefPattern` (사용자 정의 regex)
- 4-tier priority: source-system-specific provider → 'other' custom pattern → InboundSourceType-empty custom pattern → no_match
- `AutoRouteDecision.ProviderHint` field 신규 (reason + provider 정밀 식별)

### 2.2 backend 2차 (PR #587): WebhookAdapter interface + Gitea adapter

- `package webhook` (102 line):
  - `WebhookEvent` struct (provider-agnostic normalized event: ProviderType, ProviderKey, EventType, DeliveryID, ExternalRef, ActorLogin, PayloadHash, RawPayload)
  - `WebhookAdapter` interface (`ProviderType() string` + `ExtractEvent(providerKey, payload, headers) (WebhookEvent, error)`)
  - `RegisterAdapter` / `GetAdapterForProviderType` (registry dispatcher)
  - helper: `payloadHashHex`, `firstHeader` (multi-alias 지원)
- `GiteaWebhookAdapter` (provider_type='scm'):
  - Gitea/Forgejo (X-Gitea-*) / Gogs (X-Gogs-*) / DevHub-native (X-Integration-*) header alias
  - external_ref = `GITEA-<number>` (auto_route.go 의 giteaExternalRefPattern 와 cross-ref)
- IngestIntegrationProviderWebhook handler 통합 = **후속 follow-up** (3~5차 commit dispatcher 활용)

### 2.3 backend 3차 (PR #588): Jira + Generic adapter + init dispatcher

- `JiraWebhookAdapter` (provider_type='alm'):
  - Atlassian Jira webhook envelope (webhookEvent + issue + user)
  - event type 4-tier priority (payload.webhookEvent > X-Atlassian-Webhook-Identifier > X-Integration-Event > empty)
  - external_ref = Jira issue key (e.g., `DEV-456`, `PROJ-123`)
  - jiraEventTypeAllowed whitelist (jira:issue_* / jira:project_* / jira:sprint_* / jira:comment_* / jira:worklog_*)
- `GenericWebhookAdapter` (provider_type='other'):
  - generic envelope 의 3-field minimal schema: `{event, external_ref, actor}`
  - event type: payload.event > X-Integration-Event > X-Webhook-Event > X-Event-Type > empty
  - `MatchExternalRefPattern(externalRef, customPattern)`: custom regex (InboundSourceRoutingConfig 와 cross-ref)
  - 10 case 검증 (CUSTOM/JIRA/#/! pattern + empty/invalid edge)
- `init()` 자동 등록 dispatcher (3 adapter): 별도 main.go 변경 불요

### 2.4 backend 4차 (PR #589): openapi schema 정합

- `backend-core/internal/httpapi/swaggerui/asset/openapi.yaml` (CI lint canonical) 정합
- 신규 schema 3종:
  - `WebhookEvent`: 7 required field (provider_type, provider_key, event_type, external_ref, payload_hash, raw_payload) + 2 optional (delivery_id, actor_login)
  - `WebhookAdapterType`: provider_type enum (scm | alm | other)
  - `InboundSourceRoutingConfig`: custom_external_ref_pattern + custom_requester_pattern + custom_department_pattern (3 field, 모두 optional)
- `docs/openapi.yaml` byte-identical sync (swaggerui 와 diff 0)
- schema reference 정합: docs 12 + swaggerui 12

### 2.5 frontend 5차 (PR #590, 본 sprint): multi-provider 운영 UI

- `frontend/components/admin/inbound-source-config/` (4 widget + index):
  - `InboundSourceTypeSelector` (raw `<select>` + 4 옵션 + hint 표시)
  - `InboundSourceConfigEditor` (JSONB textarea editor + parse error + save 버튼)
  - `PatternPreview` (provider-specific 패턴 / custom regex 검증 + MATCH/NO MATCH 표시)
  - `InboundSourceManager` (3 widget 통합 view + platform selector + 저장 + audit)
- `frontend/domain/integration-registry/schema/integration.types.ts`:
  - `IntegrationProviderType` (7종) + `InboundSourceType` (4종) + `WebhookProviderHint` (7종)
  - `AutoRouteDecision` (1차 PR #586 의 Go struct 와 1:1 매핑)
  - `InboundSourceRoutingConfig` (TypeScript interface, 1차 PR #586 의 Go struct 와 1:1 매핑)
  - `WebhookEvent` (TypeScript interface, 2~3차 PR 의 Go struct 와 1:1 매핑)
  - `PlatformInboundSourceView` (frontend 운영 UI 의 SSOT)
- `frontend/app/(dashboard)/admin/inbound-source/page.tsx` (신규 page, mock data + handleSave callback)
- `frontend/app/(dashboard)/admin/page.tsx` 강화: 운영 도구 link 에 Inbound Source 진입 추가
- `frontend/components/admin/inbound-source-config/inbound-source-config.test.tsx` (Vitest 4+3+1+2 = 10 case)
- `frontend/tests/e2e/admin-x2.spec.ts` (Playwright 4 case: TC-ADMIN-X2-01/02/03/04)

## 3. 정공법 정합 (1:1 cross-ref)

| Layer | Go (backend) | TypeScript (frontend) | OpenAPI (schema) | 1:1 |
|---|---|---|---|---|
| Adapter | `WebhookAdapter` (PR #587) | `IntegrationProviderType` (PR #590) | `WebhookAdapterType` (PR #589) | ✅ |
| Event | `WebhookEvent` struct (PR #587) | `WebhookEvent` interface (PR #590) | `WebhookEvent` schema (PR #589) | ✅ |
| Routing | `InboundSourceRoutingConfig` (PR #586) | `InboundSourceRoutingConfig` (PR #590) | `InboundSourceRoutingConfig` schema (PR #589) | ✅ |
| Pattern | `giteaExternalRefPattern` etc. (PR #586) | provider-specific default (PR #590) | `enum` 7종 (PR #589) | ✅ |

## 4. 권한 / RBAC

- X-2 의 frontend 운영 UI (`/admin/inbound-source`) = system_admin 일임 (Sidebar 의 isSystemAdmin(actor?.role) gate 자동)
- backend 의 IngestIntegrationProviderWebhook handler = integration_registry scope + system_admin 의 path-level 강제 (routePermissionTable 의 `integration_providers` resource)
- 비-admin (developer / team_manager) 진입 시 defaultLandingFor(role) redirect
- e2e TC-ADMIN-X2-03 검증: developer 가 /admin/inbound-source 진입 시 /developer landing

## 5. 구현 단계 (5 chunk 정공법)

### 5.1 1차 (commit `PR #586`, merge `7075dbd`)

- backend `auto_route.go` multi-provider pattern matcher depth
- 2 file, 245 line
- go test 12/12 PASS

### 5.2 2차 (commit `PR #587`, merge `57fda7e`)

- backend `internal/infrastructure/webhook/` + WebhookAdapter + Gitea adapter
- 3 file 신규, 408 line
- go test 9/9 PASS

### 5.3 3차 (commit `PR #588`, merge `b026a53`)

- backend Jira + Generic adapter + init() dispatcher
- 4 file 신규, 469 line
- go test 22/22 PASS

### 5.4 4차 (commit `PR #589`, merge `e26c58a`)

- openapi.yaml X-2 schema block (WebhookEvent + WebhookAdapterType + InboundSourceRoutingConfig)
- 2 file, 179 line
- openapi lint PASS + schema reference 12/12

### 5.5 5차 (commit `PR #590`, merge pending)

- frontend multi-provider 운영 UI + e2e + 본 ADR + traceability/CHANGELOG/mirror-list

## 6. 검증

### 6.1 backend (1~3차 PR)

- `go build ./...` PASS (5 PR 모두)
- `go test -v ./internal/infrastructure/webhook/` **22/22 PASS** (9 Gitea + 6 Jira + 6 Generic + 1 init dispatcher)
- `go test -v ./internal/domain/application-lifecycle/routing/` **12/12 PASS** (6 N-13 + 6 X-2)

### 6.2 openapi (4차 PR)

- `bash scripts/check-openapi-yaml-lint.sh` PASS
- schema reference 정합: docs 12 + swaggerui 12
- docs↔swaggerui byte-identical diff = 0

### 6.3 frontend (5차 PR)

- `npm run test` (Vitest) — 10 case (4 widget + 통합 view) PASS
- `npm run lint` (eslint) — clean
- `npm run typecheck` (tsc --noEmit) — clean
- `npm run e2e -- admin-x2` (Playwright) — 4 case (TC-ADMIN-X2-01/02/03/04) PASS

### 6.4 회귀

- 기존 admin settings 9 sub-page 흐름 영향 0 (InboundSourceManager 는 별도 page)
- admin catalog (Phase 1+2) 흐름 영향 0 (X-2 의 운영 UI 는 별도 page)
- /admin/topology-v2 흐름 영향 0

## 7. 추적성 (X-2 의 5 chunk 종합)

| 단계 | ID | 정의 |
| --- | --- | --- |
| REQ | REQ-FR-115 | X-2 multi-provider webhook architecture (Gitea/Jira/Custom) |
| UC | UC-ADMIN-09 | system_admin 운영: inbound_source_type + config 관리 (X-2) |
| ARCH | ARCH-25 | WebhookAdapter + dispatcher registry architecture (X-2) |
| API | API-108 | `POST /api/v1/integration/providers/{provider_id}/webhook` 의 multi-provider dispatcher (기존 정의 활용) |
| RM | RM-ADMIN-09 | X-2 frontend 운영 UI (InboundSourceManager) |
| IMPL | IMPL-inbound-source-multi-provider-01 (1차 PR) + IMPL-webhook-adapter-01 (2차 PR) + IMPL-webhook-adapter-02 (3차 PR) + IMPL-frontend-inbound-source-01 (5차 PR) | 4 IMPL |
| UT | UT-inbound-source-multi-provider-01 (1차 PR) + UT-webhook-adapter-01 (2차 PR) + UT-webhook-adapter-02 (3차 PR) + UT-frontend-inbound-source-01 (5차 PR) | 4 UT |
| TC | TC-ADMIN-X2-01/02/03/04 | 4 e2e |
| ADR | ADR-0033 | 본 문서 |

**신규 ID 발급 12 row** (REQ-1 + UC-1 + ARCH-1 + API-1 + RM-1 + IMPL-4 + UT-4 = 12).

## 8. 대안 검토

### 8.1 webhook 의 per-provider 별도 HTTP endpoint

- 장점: 단순, per-provider 별도 signature 검증
- 단점: 운영자가 provider 별도 URL 관리 부담, multi-provider 일반화 미흡
- 결정: **기각**. 단일 `/api/v1/integration/providers/{provider_id}/webhook` + WebhookAdapter dispatcher (X-2 정공법)

### 8.2 gitea_webhook.go 의 legacy 코드 유지 + multi-provider 별도 추가

- 장점: legacy 코드 회귀 없음
- 단점: 두 가지 webhook ingest 코드 공존, 운영자 audit trail 분산
- 결정: **기각**. legacy gitea_webhook.go 는 DevHub 운영 event source (자체 Gitea instance) — 본 ADR 의 integration-registry 의 multi-provider 와 별도 route. cross-cut 미흡.

### 8.3 WebhookAdapter 의 method signature 단순화 (single ExtractEvent)

- 장점: interface 단순
- 단점: 4-tier priority 의 provider-specific vs custom 정공법 미흡
- 결정: **채택**. 본 ADR 의 interface 가 ExtractEvent(providerKey, payload, headers) 만 — adapter 내부에서 4-tier priority 처리. **이유**: Go interface 의 작은 surface area + 각 adapter 의 자율성 + 테스트 용이성.

## 9. 후속 (forward path)

1. **IngestIntegrationProviderWebhook handler 의 adapter dispatch 통합** (별도 follow-up) — 본 ADR 의 adapter + init dispatcher 활용, handler 가 GetAdapterForProviderType 호출 후 ExtractEvent 호출하는 1 commit. 본 sprint scope 외.
2. **CiCd/Doc/Infra/TaskTracker provider adapter 추가** (post-X-2) — 본 ADR 의 4-tier priority 정공법 그대로 활용, 신규 adapter 구현 후 `RegisterAdapter` 호출.
3. **webhook retry 정책** (별도 follow-up) — webhook ingest 실패 시 재시도 (exponential backoff) + dead letter queue.
4. **webhook signature verification 통합** (X-2 의 후속) — 본 ADR 의 adapter 에 HMAC verification method 추가, provider-native signature 와 DevHub-native X-Integration-Signature 의 통합 검증.
5. **e2e admin-x2 의 CI 정합** (CI e2e-shard 1 추가) — 본 ADR 의 e2e 4 case 가 e2e-shard 1 의 smoke 테스트로 정합.

## 10. supersession trigger

- **CiCd/Doc/Infra/TaskTracker provider adapter 추가** 시: 본 ADR 의 WebhookAdapter interface 의 변경 가능성 (event_type 의 일반화) — 별도 ADR 발행.
- **webhook signature verification 변경** 시: 본 ADR 의 provider-native header (X-Gitea-Event / X-Atlassian-Event) 가 1개 provider 만의 변종이 아닌 generic spec 으로 통합될 시 — 본 ADR 의 IngestIntegrationProviderWebhook handler 의 signature verify 정공법 갱신.
- **외부 SSO federation** (X-6, P1-3, issue #214) 진행 시: Keycloak identity broker 의 webhook signature 정공법이 본 ADR 의 adapter pattern 으로 흡수 가능 — 별도 ADR 발행.
- **v0.1.2 / v0.2.0** 의 운영 dashboard 확장 (build run queue, ci-run status, audit log streaming) 시: 본 ADR 의 widget 4 패턴 (2x2 grid + widget 4종 + audit) 을 standard pattern 으로 승격 검토.

## 11. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-06-13 | 1차 작성 (sprint `feat/work_260614-x2-frontend-e2e` 의 5차 PR) — X-2 의 5 chunk 종합 결정. backend (1~3차 PR #586/587/588) + openapi (4차 PR #589) + frontend 운영 UI (5차 PR #590) 의 architecture + RBAC + e2e + 추적성 + supersession trigger 명문화. 신규 ID 발급 12 row. 본 ADR 의 §1 (배경) + §2 (결정 5 chunk) + §3 (정공법 1:1 cross-ref) + §4 (권한) + §5 (구현 단계) + §6 (검증) + §7 (추적성) + §8 (대안 검토) + §9 (후속) + §10 (supersession trigger). Tier: 사외. |
