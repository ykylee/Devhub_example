# ADR-0032: System Admin 운영 대시보드 (X-1, RM-M4-07, v0.1.1 sprint)

- **문서 목적**: `release_v0-1_roadmap.md` line 198 의 X-1 (System Admin 운영 대시보드, RM-M4-07) 의 정공법 (Gitea sync job 큐/상태 + provider health 의 운영 view + admin catalog 와의 정합) 을 단일 ADR 로 명문화한다.
- **범위**: X-1 의 backend 3 endpoint + frontend 4 widget + admin landing page 강화 + system_admin 권한 + openapi 정합 + provider health endpoint 의 별도 carve + 기존 admin catalog (옵션 B) 와의 경계. **코드 변경은 §6 PR 목록 참조**, 본 ADR 은 정책 결정만.
- **대상 독자**: PMO, 시스템 관리자 (`system_admin`), FE/BE 개발자, QA, 후속 sprint 작업자.
- **상태**: accepted (2026-06-13, sprint `feat/work_260614-x1-system-admin-dashboard` 의 1차 PR + sprint `feat/work_260614-x1-frontend-e2e` 의 2차 PR 종합 결정)
- **최종 수정일**: 2026-06-13
- **결정 근거 sprint**: `feat/work_260614-x1-system-admin-dashboard` (1차 PR #583, backend + openapi, main HEAD `77f0c76`), `feat/work_260614-x1-frontend-e2e` (2차 PR 본 ADR 동반, frontend widget 4 + e2e).
- **Tier**: **공용** (ADR + frontend code + backend code + openapi, 사내 한정 정보 미포함)
- **관련 문서**: [release_v0-1_roadmap.md §3.5 X-1 + §4.2 M-v0.1.1](../planning/release_v0-1_roadmap.md), [`system_admin_catalog_plan_2026-05-27.md` §2 옵션 B](../planning/system_admin_catalog_plan_2026-05-27.md) (admin catalog 의 1차 출처), [`2026-06-13-x1-system-admin-dashboard-sprint-plan.md`](../planning/2026-06-13-x1-system-admin-dashboard-sprint-plan.md) (X-1 sprint plan), [openapi.yaml §X-1 schema](#) (swaggerui/asset 의 IntegrationSyncJob / IntegrationSyncJobStatusCounts schema), [PR #583](https://github.com/ykylee/Devhub_example/pull/583) (1차 PR backend + openapi), 2차 PR (frontend widget 4 + e2e, sprint `feat/work_260614-x1-frontend-e2e`).

## 1. 배경

### 1.1 시스템 운영자 가시성 부족

v0.1.0-alpha 출시 (2026-06-11) + v0.1.1-alpha 격하 (commit `3e2701b3`, 잔여 5 + X-1~8) 이후, 시스템 관리자 (`system_admin`) 가 운영 중 다음 정보에 빠르게 접근 불가:

1. **Gitea sync job 큐/상태** — `integration_sync_jobs` table 의 `queued` / `running` / `failed` 분포 + 최근 50개 흐름
2. **Provider health** — Gitea / ALM / HomeLab 등 provider 별 health (RuntimeSnapshotProvider 의 degraded / healthy) summary
3. **운영 dashboard summary** — sync job count (by status) + provider health count (by health) + degraded count

### 1.2 기존 코드 정공법 (reuse-first)

- `integration_sync_jobs` table (DDL migration 000001 line 316) — schema 이미 존재
- `IntegrationRepository.CreateIntegrationSyncJob` (line 444) + `UpdateIntegrationSyncJobStatus` (line 659) + `AcquireNextQueuedSyncJob` (line 640) — **이미 존재**
- `infra_integrations.go` (`infraTopologyV2` / `listInfraServices` / `ingestInfraServicesSnapshot` admin endpoint) — **이미 존재**
- `RuntimeSnapshotProvider` (provider health snapshot) — **이미 존재**
- `IntegrationStore` interface — `ListIntegrationProviders` / `ListIntegrationBindings` 등 method 존재, **List/Get sync job 미존재** (본 ADR 의 보강 대상)

### 1.3 frontend 정공법

- `frontend/app/(dashboard)/admin/catalog/page.tsx` (30K, 5 tab: applications / repositories / projects) — **`system_admin_catalog_plan_2026-05-27.md` §2 옵션 B 의 admin catalog Phase 1+2 이미 구현** (별도 carve)
- `frontend/app/(dashboard)/admin/page.tsx` — admin landing ("Sys Admin Dashboard Archived" 표시) — **본 ADR 의 X-1 widget 4 추가 강화 대상**
- `frontend/app/(dashboard)/admin/settings/` (9 sub-page: audit / dev-request-tokens / dev-requests / integration-bindings / integrations / organization / permissions / platforms / users) — system_admin 만 접근 (routePermissionTable)
- `frontend/shared/ui-foundation/layout/Sidebar.tsx` — `isSystemAdmin(actor?.role)` gate + admin catalog / reception-test / settings link 노출

## 2. 결정

### 2.1 scope

**본 ADR 의 정공법 = X-1 (RM-M4-07) backend 3 endpoint + frontend 4 widget + admin landing page 강화**. **admin catalog (3 도메인 자산 UI) = 이미 구현 (별도 carve, 본 ADR scope 외)**.

### 2.2 backend 3 endpoint (API-104/105/106)

1. **`GET /api/v1/admin/integrations/sync-jobs?status=&limit=&offset=`** (API-104) — sync job 큐/상태 조회
   - status validation: `queued | running | succeeded | failed` 4 status 만 허용, 그 외 400
   - limit 1~100 (default 50), offset 0+ (default 0)
   - response: `{ items: [...], total, limit, offset }`
   - system_admin 일임 (`routePermissionTable` 의 `integration_sync_jobs` resource gate 자동)

2. **`GET /api/v1/admin/integrations/sync-jobs/:jobID`** (API-105) — sync job 단건 조회
   - not found 시 404 (`store.ErrNotFound`)
   - response: `{ job_id, provider_id, requested_by, status, created_at }`
   - system_admin 일임

3. **`GET /api/v1/admin/integrations/summary`** (API-106) — dashboard summary
   - response: `{ sync_job_status_counts: { queued, running, succeeded, failed } }`
   - 4 status 별 count (SUM CASE WHEN)
   - system_admin 일임

### 2.3 frontend 4 widget

1. **SyncJobQueueWidget** — queued + running status 의 sync job 큐 + 최근 10개 row + status dot (amber/sky) + refresh button
2. **SyncJobStatusWidget** — 4 status 별 count 의 2x2 grid + status 별 색 border
3. **DashboardSummaryWidget** — 종합 summary (totalJobs / queueDepth / failedCount / successRate) + tone (good/warn/danger/neutral) 별 text color
4. **ProviderHealthWidget** — **placeholder** (provider health endpoint 미구현, §3 carve 참조)

`/admin` landing page 강화 — "Sys Admin Dashboard Archived" 메시지 제거 + X-1 widget 4 의 2x2 grid 배치 + "운영 도구" 섹션 (Topology v2 / Settings link) 유지.

### 2.4 권한

- X-1 endpoint 3종 + widget 4 = `system_admin` 만 접근
- 비-admin (developer / team_manager) 진입 시 `defaultLandingFor(role)` redirect (예: developer → `/developer`)
- 권한 가드 = `enforceRoutePermission` middleware 의 `routePermissionTable` 의 `integration_sync_jobs` resource gate 자동 처리
- 별도 handler 내부 role check 불필요 (middleware 위임)
- Sidebar 의 `isSystemAdmin(actor?.role)` gate = admin landing page (`/admin`) 진입 자동 권한 가드

### 2.5 openapi 정합

- `backend-core/internal/httpapi/swaggerui/asset/openapi.yaml` (CI lint canonical) 정합
- paths +3: sync-jobs, sync-jobs/{jobID}, summary
- schemas +2: IntegrationSyncJob, IntegrationSyncJobStatusCounts
- `docs/openapi.yaml` (별도, line 9455) ↔ `swaggerui/asset/openapi.yaml` (line 9360) 의 drift 95 line — **별도 sprint 에서 sync** (CI lint target 인 `swaggerui/asset/` 만 본 ADR 의 정공법 갱신 대상)

## 3. Provider Health Endpoint 의 별도 carve

### 3.1 미구현 사유 (현시점)

`RuntimeSnapshotProvider` (provider health snapshot) 가 **이미 존재**하지만, "provider health summary (provider 별 healthy/degraded count)" 의 admin endpoint 는 **미존재**. 본 X-1 sprint 의 scope 폭주 방지 + 별도 carve 결정:

1. **본 sprint (X-1 1차 PR + 2차 PR)**: §2.2 의 sync job endpoint 3종 + §2.3 의 widget 4 (단, ProviderHealthWidget 은 placeholder)
2. **별도 follow-up sprint (post-MVP)**: `GET /api/v1/admin/integrations/provider-health` (API-107) + `RuntimeSnapshotProvider.GetProviderHealthSummary` 메서드 보강 + ProviderHealthWidget 의 placeholder → 실제 rendering 으로 교체

### 3.2 trigger / SLA

- **trigger**: 1주일 staging 운영 후 사용자 결정 시점 (v0.1.0 GA 전후)
- **SLA**: 별도 sprint 진입 시점 (2~3 session, ~5 commit)
- **scope**: backend method 1~2 + endpoint 1 + schema 1 + widget 1 + e2e 1

## 4. 기존 admin catalog 와의 정합

### 4.1 admin catalog 의 scope (system_admin_catalog_plan_2026-05-27.md §2 옵션 B)

- `/admin/catalog?tab=applications|repositories|projects` — 3 도메인 자산 (Application / Repository / Project) 의 **조회/관리 UI**
- **이미 구현** (frontend/app/(dashboard)/admin/catalog/page.tsx, 30K)
- Phase 1+2 (3 tab 목록 MVP) 완료, Phase 3 (관리 액션: 보관/복원, 상태 변경, repository 연결 관리) 는 별도 carve

### 4.2 X-1 의 scope 차이

| 측면 | admin catalog | X-1 (RM-M4-07) |
| --- | --- | --- |
| **대상** | 3 도메인 자산 (App/Repo/Project) | 운영 메트릭 (sync job / provider health) |
| **UI** | 3 tab 목록/상세 | 4 widget (큐 / status / summary / health) |
| **endpoint** | 기존 `/api/v1/platforms` + `/repositories` + `/platforms/:id/projects` 재사용 | 신규 3 endpoint (sync-jobs list/get/summary) |
| **권한** | system_admin | system_admin |
| **현 status** | ✅ 구현 (별도 carve) | 본 ADR 의 X-1 sprint |

### 4.3 정합 (둘 다 유지)

- admin catalog = 운영자 동선 (3 도메인 자산 관리)
- X-1 = 운영 메트릭 (sync job / provider health 가시성)
- 두 menu 가 Sidebar 의 "Admin" 그룹에 동시 노출 (`/admin/catalog` + `/admin` 의 widget 4 + `/admin/settings` 의 9 sub-page)
- 사용자 동선: `/admin` 진입 → X-1 widget 4 로 메트릭 glance → 필요 시 `/admin/catalog` 로 drill-down → 필요 시 `/admin/settings` 로 운영 액션

## 5. 구현 단계 (5 chunk)

### 5.1 1차 (commit `04257e3`, PR #583, main HEAD `77f0c76`)

- `domain.IntegrationSyncJob` struct 3 + `IntegrationRepository.List/Get/StatusCounts` method 3
- `TestIntegration_IntegrationSyncJob_CRUD` (backend-integration, 8 step)
- sprint plan 신규

### 5.2 2차 (commit `1639af1`, PR #583)

- `store.IntegrationSyncJobListOptions` type + `IntegrationStore` interface 3 method
- `ListIntegrationSyncJobs` / `GetIntegrationSyncJob` / `GetIntegrationSyncJobStatusSummary` handler 3
- `router.go` route 3개 등록 (system_admin 가드 = enforceRoutePermission middleware)
- fake store 3 method + handler test 4

### 5.3 3차 (commit `53bd8d3`, PR #583)

- `swaggerui/asset/openapi.yaml` paths +3 + schemas +2
- `bash scripts/check-openapi-yaml-lint.sh` PASS

### 5.4 4차 (commit 본 2차 PR, sprint `feat/work_260614-x1-frontend-e2e`)

- `frontend/components/admin/x1-widgets/SyncJobQueueWidget` (114 line)
- `frontend/components/admin/x1-widgets/SyncJobStatusWidget` (83 line)
- `frontend/components/admin/x1-widgets/ProviderHealthWidget` (94 line, placeholder)
- `frontend/components/admin/x1-widgets/DashboardSummaryWidget` (98 line)
- `frontend/components/admin/x1-widgets/index.ts` (barrel)
- `frontend/components/admin/x1-widgets/x1-widgets.test.tsx` (Vitest, 5 case)
- `frontend/domain/integration-registry/service/admin-x1.service.ts` (48 line)
- `frontend/domain/integration-registry/schema/integration.types.ts` (X-1 types +39 line)
- `frontend/app/(dashboard)/admin/page.tsx` 강화 (X-1 widget 4 + grid + 운영 도구 link)

### 5.5 5차 (commit 본 2차 PR, sprint `feat/work_260614-x1-frontend-e2e`)

- `frontend/tests/e2e/admin-x1.spec.ts` (3 case: TC-ADMIN-X1-01/02/03)
- ADR-0032 (본 문서)
- traceability/report.md + CHANGELOG.md + mirror-list.md

## 6. 검증

### 6.1 backend (1차 PR #583)

- `go build ./...` — PASS
- `go test ./internal/domain/integration-registry/view/` — PASS (4/4)
- `go test -c -o /tmp/test_x1.o ./internal/domain/integration-registry/repository/` — compile PASS
- `bash scripts/check-openapi-yaml-lint.sh` — PASS (yaml valid + semver + paths=87 >= 84 + cross-link ok)
- backend-integration CI job: `DEVHUB_TEST_DB_URL` 설정 시 `TestIntegration_IntegrationSyncJob_CRUD` 실행 (8 step)

### 6.2 frontend (2차 PR, sprint `feat/work_260614-x1-frontend-e2e`)

- `npm run lint` (eslint) — CI frontend-unit job
- `npm run test` (vitest) — CI frontend-unit job (5 case: SyncJobQueueWidget Success/Empty + SyncJobStatusWidget Success + DashboardSummaryWidget Success + ProviderHealthWidget Placeholder)
- `npm run typecheck` (tsc --noEmit) — CI frontend-unit job
- `npm run e2e -- admin-x1` (Playwright) — CI e2e-shard 1 (3 case: TC-ADMIN-X1-01 system_admin 진입 + widget 4 / TC-ADMIN-X1-02 sync job status API mock + DashboardSummary 위젯 값 / TC-ADMIN-X1-03 non-admin redirect)

## 7. 추적성 (본 ADR + 1차/2차 PR 종합)

| 단계 | ID | 정의 |
| --- | --- | --- |
| **REQ** | REQ-FR-114 | system_admin 운영 대시보드 (Gitea sync job + provider health) |
| **UC** | UC-ADMIN-08 | use case (운영 대시보드 진입 + widget 조회) |
| **ARCH** | ARCH-24 | admin 운영 대시보드 컴포넌트 구조 (X-1 widget 4 + admin page 강화) |
| **API** | API-104 | GET /api/v1/admin/integrations/sync-jobs |
| **API** | API-105 | GET /api/v1/admin/integrations/sync-jobs/:jobID |
| **API** | API-106 | GET /api/v1/admin/integrations/summary |
| **(API)** | API-107 (별도) | GET /api/v1/admin/integrations/provider-health (ProviderHealthWidget placeholder target) |
| **RM** | RM-ADMIN-08 | X-1 widget 4 + admin page 강화 |
| **IMPL** | IMPL-admin-x1-01 | IntegrationRepository.List/Get/StatusCounts method 3 |
| **IMPL** | IMPL-admin-x1-02 | httpapi admin handler 3 + route 3 |
| **IMPL** | IMPL-frontend-admin-x1-01 | X-1 widget 4 (SyncJobQueue/SyncJobStatus/ProviderHealth/DashboardSummary) |
| **IMPL** | IMPL-frontend-admin-x1-02 | admin landing page 강화 (widget 4 import + grid + 운영 도구 link) |
| **UT** | UT-admin-x1-01 | backend-integration test 1 (TestIntegration_IntegrationSyncJob_CRUD) |
| **UT** | UT-admin-x1-02 | httpapi handler test 4 (List/Get/StatusSummary + InvalidStatus) |
| **UT** | UT-admin-x1-03 | frontend widget test 5 (Vitest) |
| **TC** | TC-ADMIN-X1-01 | system_admin /admin 진입 + widget 4 + API-106 fetch |
| **TC** | TC-ADMIN-X1-02 | sync job status API mock + DashboardSummary 위젯 total/queue/failed/rate |
| **TC** | TC-ADMIN-X1-03 | non-admin redirect → defaultLandingFor(role) |
| **ADR** | ADR-0032 | 본 문서 |

**신규 ID 발급 17 row** (REQ-1 + UC-1 + ARCH-1 + API-3 + API-1(별도) + RM-1 + IMPL-4 + UT-3 + TC-3 + ADR-1 = 17).

## 8. 대안 검토

### 8.1 기존 admin settings 의 별도 tab 추가

- 장점: 추가 menu 없음
- 단점: Settings(설정) vs Operations(운영) 성격 혼재 (admin_catalog_plan §2 옵션 A 의 단점)
- 결정: **기각**. 옵션 B (admin catalog + 본 ADR 의 X-1 widget) 정공법 유지

### 8.2 admin catalog 의 3 tab 에 X-1 widget 통합

- 장점: 단일 페이지 (catalog)
- 단점: catalog 의 "자산 관리" 와 X-1 의 "운영 메트릭" 성격 혼재
- 결정: **기각**. admin landing page (`/admin`) 의 widget 4 + 별도 admin catalog page 의 정공법 유지

### 8.3 admin settings 의 9 sub-page 중 하나로 X-1 추가

- 장점: 기존 menu tree 활용
- 단점: Settings 의 "구성/정책" vs X-1 의 "운영 view" 성격 혼재
- 결정: **기각**. admin landing page 의 widget 4 + 운영 도구 link (Topology v2, Settings) 의 정공법 유지

## 9. 후속 (forward path)

1. **Provider Health Endpoint (API-107)** — 별도 sprint (post-MVP, 1주일 staging 운영 후 사용자 결정 시점). scope = backend method 1~2 + endpoint 1 + schema 1 + widget 1 + e2e 1 (ProviderHealthWidget placeholder → 실제 rendering).
2. **admin catalog Phase 3 (관리 액션)** — 별도 carve (P1~P2 backlog). 보관/복원 / 상태 변경 / repository 연결 관리. 본 ADR scope 외.
3. **e2e seed 정합 (TC-ADMIN-X1-02 의 mock data)** — backend-integration 의 `integration_sync_jobs` table 의 seed 가 e2e 환경에 정합되어 있으면 mock 불필요, 불일치 시 Playwright 의 page.route() mock 사용. 정합 확인 후 mock 제거 가능.

## 10. supersession trigger

- 1주일 staging 운영 후 사용자 결정 (Provider Health Endpoint 별도 sprint 진입 / 보류)
- admin catalog Phase 3 진입 시 본 ADR 의 X-1 widget 4 와 admin catalog 의 관계 재평가 (현재는 둘 다 유지 정공법)
- v0.1.1 GA 후 v0.1.2 / v0.2 의 운영 dashboard 확장 (예: build run queue / ci-run status / audit log streaming) 시 본 ADR 의 widget 4 패턴 (2x2 grid + widget 4종) 을 standard pattern 으로 승격 검토
