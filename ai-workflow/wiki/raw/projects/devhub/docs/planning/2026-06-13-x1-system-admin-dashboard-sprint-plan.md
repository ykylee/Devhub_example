# X-1 System Admin 운영 대시보드 sprint plan (2026-06-13, v0.1.1 milestone sprint)

- 문서 목적: v0.1.1 milestone 의 X-1 (System Admin 운영 대시보드, RM-M4-07) 의 구현 scope/단계/ID 발급/추적성/검증 전략을 단일 sprint plan 으로 합본 정의한다.
- 1차 출처 합본:
  - `docs/planning/system_admin_catalog_plan_2026-05-27.md` §2 옵션 B (Admin Catalog) + §3 IA + §4 MVP + §8 테스트 전략
  - `docs/planning/release_v0-1_roadmap.md` line 198 의 X-1 정의 (Gitea sync job 큐/상태 + provider health, FE+BE)
  - `docs/planning/release_v0-1_roadmap.md` line 42 정합 (P3-9 RM-M4-07, v2 P3) — 본 sprint = **v0.1.1 milestone X-1 carve (P3 → v0.1.1 sprint scope 전환)**
- 범위: backend (admin 운영 endpoint 3 + IntegrationRepository method 보강 3) + frontend (admin 운영 도구 진입 + X-1 widget 4) + e2e (admin catalog + X-1 widget)
- 대상 독자: PMO, 시스템 관리자, FE/BE 개발자, QA
- 상태: planned
- 최종 수정일: 2026-06-13
- 관련 문서: `docs/planning/system_admin_catalog_plan_2026-05-27.md`, `docs/planning/release_v0-1_roadmap.md` §3.5 X-1 + §4.1 M-v0.1.0 + §4.2 M-v0.1.1, `docs/traceability/conventions.md`, `docs/traceability/report.md`

---

## 1) 배경 및 정합

### 1.1 시스템 운영자 가시성 부족

v0.1.0-alpha 출시 (2026-06-11) + v0.1.1-alpha 격하 (commit `3e2701b3`) 이후, 시스템 관리자 (`system_admin`) 가 운영 중 다음 정보에 빠르게 접근할 수 없다:

1. **Gitea sync job 큐/상태** — `integration_sync_jobs` table 의 `queued` / `running` / `failed` 분포 + 최근 50개 status 흐름
2. **Provider health** — Gitea/ALM/HomeLab 등 provider 별 health (RuntimeSnapshotProvider 의 degraded/healthy) summary
3. **운영 dashboard summary** — sync job count (by status) + provider health count (by health) + degraded count

### 1.2 기존 코드 정합

- `backend-core/internal/domain/integration-registry/repository/integration_registry.go` line 444 `CreateIntegrationSyncJob` + line 659 `UpdateIntegrationSyncJobStatus` — **이미 존재**
- `backend-core/internal/store/integration_sync_jobs_integration_test.go` — `integration_sync_jobs` table DDL/insert/select **이미 존재**
- `backend-core/internal/httpapi/infra_integrations.go` — `infraTopologyV2` / `listInfraServices` / `ingestInfraServicesSnapshot` admin endpoint **이미 존재**
- `backend-core/internal/httpapi/runtime_snapshot_provider.go` — `RuntimeSnapshotProvider` (provider health snapshot) **이미 존재**
- **`ListIntegrationSyncJobs` / `GetIntegrationSyncJob` / `GetProviderHealthSummary` 메서드 = 미존재** — 본 sprint 의 backend 보강 대상

### 1.3 frontend 정합

- `frontend/app/(dashboard)/admin/catalog/page.tsx` (30K) — `system_admin_catalog_plan` Phase 1+2 의 **3 도메인 자산 UI 이미 구현**
- `frontend/app/(dashboard)/admin/page.tsx` — admin landing ("아카이브" 표시) — 본 sprint 의 **운영 도구 진입 강화 대상** (X-1 widget 4 추가)
- `frontend/app/(dashboard)/admin/settings/` — 9 sub-page (audit / dev-request-tokens / dev-requests / integration-bindings / integrations / organization / permissions / platforms / users) — system_admin 만 접근

**scope 재확정** (release_v0-1_roadmap.md line 198 + admin page 정합):
- **admin catalog (3 도메인 자산 UI) = 이미 구현 (별도 carve 없음)**
- **X-1 (RM-M4-07) = Gitea sync job 큐/상태 + provider health + dashboard summary 신규**

---

## 2) 목표 (in-scope)

### 2.1 backend (scope = 4~5 method + 3 endpoint + 1~2 migration)

1. `IntegrationRepository` 보강
   - `ListIntegrationSyncJobs(ctx, status, limit, offset)` — 큐 (status) + 최근 50개
   - `GetIntegrationSyncJob(ctx, jobID)` — 단건 조회
   - `GetSyncJobStatusCounts(ctx)` — `queued` / `running` / `succeeded` / `failed` count (dashboard summary)
2. httpapi admin 운영 endpoint 3종 (system_admin 권한 가드)
  - `GET /api/v1/admin/integrations/sync-jobs?status=&limit=&offset=` — sync job 큐/상태
  - `GET /api/v1/admin/integrations/sync-jobs/:jobID` — 단건 조회
  - `GET /api/v1/admin/integrations/summary` — sync job status count + provider health count
3. (선택) provider health endpoint — `RuntimeSnapshotProvider.GetProviderHealthSummary(ctx)` 메서드 보강 + `GET /api/v1/admin/integrations/provider-health`

### 2.2 frontend (scope = 1 page + 1 widget set + 1 e2e)

1. `frontend/app/(dashboard)/admin/page.tsx` 강화
   - "운영 도구" 섹션 추가 (Topology v2 / Settings / X-1 widget 4)
   - X-1 widget 4종 (sync job queue / sync job status counts / provider health summary / dashboard summary)
2. widget 컴포넌트 4종 (`frontend/components/admin/x1-widgets/`)
   - `SyncJobQueueWidget.tsx` (status: queued/running) — 큐 depth + 최근 10개 row
   - `SyncJobStatusWidget.tsx` (status counts) — 4 status 별 count + chart
   - `ProviderHealthWidget.tsx` (provider health) — provider별 healthy/degraded count
   - `DashboardSummaryWidget.tsx` (종합) — sync job count + provider health count + degraded count

### 2.3 e2e (scope = 1 spec + 3 test)

1. `frontend/tests/e2e/admin-x1.spec.ts` (신규)
   - system_admin login → `/admin` 진입 → X-1 widget 4 표시
   - sync job status API mock → widget 값 검증
   - non-admin login → `/admin` 진입 차단 (settings redirect)

---

## 3) 정보구조(IA) 및 라우팅

### 3.1 라우트

- `/admin` (admin landing, 본 sprint 강화 — X-1 widget 4 추가)
- `/admin/catalog` (Phase 1+2, 이미 구현)
- `/admin/settings` + 9 sub-page (system_admin 만, 이미 구현)
- `/admin/topology-v2` (이미 구현)

### 3.2 권한

- `/admin/*` = `system_admin` 만 접근
- 비-admin 접근 시 기존 정책 `defaultLandingFor(role)` redirect

---

## 4) 구현 단계

### Phase 1. backend 보강 (IntegrationRepository method 3 + httpapi admin endpoint 3)

1. `IntegrationRepository.ListIntegrationSyncJobs` — DDL SELECT + status filter + limit/offset + order by created_at desc
2. `IntegrationRepository.GetIntegrationSyncJob` — DDL SELECT WHERE job_id = $1
3. `IntegrationRepository.GetSyncJobStatusCounts` — DDL SELECT COUNT(*) GROUP BY status
4. httpapi admin handler 3종 + route 등록
5. (선택) `RuntimeSnapshotProvider.GetProviderHealthSummary` 보강

### Phase 2. frontend widget 4 (X-1 widget 신규 + admin page 강화)

1. `frontend/components/admin/x1-widgets/SyncJobQueueWidget.tsx`
2. `frontend/components/admin/x1-widgets/SyncJobStatusWidget.tsx`
3. `frontend/components/admin/x1-widgets/ProviderHealthWidget.tsx`
4. `frontend/components/admin/x1-widgets/DashboardSummaryWidget.tsx`
5. `frontend/app/(dashboard)/admin/page.tsx` 강화 (위 widget 4 import + grid 배치)

### Phase 3. e2e (admin-x1.spec.ts)

1. system_admin login flow
2. X-1 widget 4 렌더링 검증
3. sync job status API mock (fixture)
4. non-admin redirect 검증

### Phase 4. CI 검증 + 추적성 정합

1. backend: `go test ./internal/domain/integration-registry/...` + `go test ./internal/httpapi/...`
2. frontend: `npm run lint` + `npm run typecheck` + `npm run test` (Vitest)
3. e2e: `npm run e2e:smoke -- admin-x1`
4. CI 4종 lint: `openapi yaml / migration prefix / changed paths / workflow`
5. 추적성: `docs/traceability/report.md` §2.4/§2.5/§2.6 + §3 + §4 + §6 정합

---

## 5) 추적성 ID 발급 (예정)

### 5.1 backend

- `API-104` — `GET /api/v1/admin/integrations/sync-jobs?status=&limit=&offset=`
- `API-105` — `GET /api/v1/admin/integrations/sync-jobs/:jobID`
- `API-106` — `GET /api/v1/admin/integrations/summary`
- (선택) `API-107` — `GET /api/v1/admin/integrations/provider-health`
- `IMPL-admin-x1-01` — `IntegrationRepository.ListIntegrationSyncJobs/GetIntegrationSyncJob/GetSyncJobStatusCounts`
- `IMPL-admin-x1-02` — httpapi admin handler 3종 + route
- (선택) `IMPL-admin-x1-03` — `RuntimeSnapshotProvider.GetProviderHealthSummary` 보강
- `UT-admin-x1-01` — backend unit test (List/Status/Health)
- `UT-admin-x1-02` — httpapi admin endpoint test (system_admin 권한 가드)

### 5.2 frontend

- `IMPL-frontend-admin-x1-01` — X-1 widget 4 (SyncJobQueue/SyncJobStatus/ProviderHealth/DashboardSummary)
- `IMPL-frontend-admin-x1-02` — admin landing page 강화 (widget 4 import + grid)
- `UT-admin-x1-03` — frontend unit test (widget 4 render + props)

### 5.3 e2e

- `TC-ADMIN-X1-01` — system_admin login + widget 4 표시
- `TC-ADMIN-X1-02` — sync job status API mock + widget 값 검증
- `TC-ADMIN-X1-03` — non-admin redirect 검증

### 5.4 ADR

- `ADR-0032` — System Admin 운영 대시보드 (RM-M4-07) — 본 sprint 의 핵심 정책 (admin endpoint 권한, widget 표시 정책, RBAC 재확인)

### 5.5 REQ/ARCH/RM

- `REQ-FR-114` — system_admin 운영 대시보드 (Gitea sync job + provider health)
- `UC-ADMIN-08` — use case (운영 대시보드 진입 + widget 조회)
- `ARCH-24` — admin 운영 대시보드 컴포넌트 구조
- `RM-ADMIN-08` — RM (X-1 widget 4 + admin page 강화)

---

## 6) 검증 전략

### 6.1 unit

- backend: `ListIntegrationSyncJobs` (status filter / limit / offset / order) + `GetIntegrationSyncJob` (not found) + `GetSyncJobStatusCounts` (empty / 4 status) + httpapi admin endpoint 3 (system_admin 권한 가드)
- frontend: widget 4 render + props (loading / error / empty) + admin page grid layout

### 6.2 integration (UI)

- 위젯 로딩/에러/빈 상태 (regression 1차)
- 각 위젯 4종 render
- admin page grid layout (1x2 + 1x2)

### 6.3 e2e

- `system_admin`: `/admin` 진입 → widget 4 표시 → sync job status API mock → 값 검증
- `non-admin`: `/admin` 진입 차단 (settings redirect)

### 6.4 회귀

- 기존 admin settings 9 sub-page 흐름 영향 점검
- 기존 admin catalog (Phase 1+2) 흐름 영향 점검
- 기존 `/admin/topology-v2` 흐름 영향 점검

---

## 7) 리스크 및 대응

1. **scope 폭주** — backend method 3 + endpoint 3 + widget 4 + e2e 1 = 1차 PR scope 폭주 가능
   - 대응: 5~8 commit 단위 분할, chunk 4~5 (1300~1700줄)
2. **admin catalog (Phase 1+2) 와의 정합** — 이미 구현된 catalog page 와 본 sprint 의 admin page 강화 (widget 4 추가) 의 layout 정합
   - 대응: admin page 의 "운영 도구" 섹션에 widget 4 grid 배치 (catalog / settings / topology-v2 link + X-1 widget 4)
3. **권한 누락** — admin endpoint 권한 가드 미흡 시 일반 사용자 노출
   - 대응: backend `system_admin` 권한 미들웨어 정합 + e2e RBAC 1 case + ADR-0032 §3 권한 정공법
4. **provider health 미구현** — `RuntimeSnapshotProvider.GetProviderHealthSummary` 가 별도 구현 필요한 경우 scope 폭주
   - 대응: sprint 진입 시 1차 PoC — `RuntimeSnapshotProvider.GetProviderHealthSummary` 가 기존 snapshot 으로 health count 계산 가능하면 in-scope, 별도 provider health endpoint 가 필요하면 후속 sprint carve

---

## 8) 테스트 전략

### 8.1 단위

- backend: `go test ./internal/domain/integration-registry/repository/...` + `go test ./internal/httpapi/...`
- frontend: `npm run test` (Vitest) — widget 4 + admin page

### 8.2 통합 (UI)

- 위젯 4종 render + 로딩/에러/빈 상태

### 8.3 E2E

- `frontend/tests/e2e/admin-x1.spec.ts` (신규)
- `system_admin` login + widget 4 + sync job status API mock
- `non-admin` redirect

---

## 9) 산출물 목록

### 9.1 backend

1. `backend-core/internal/domain/integration-registry/repository/integration_registry.go` (List/Status method 추가)
2. `backend-core/internal/httpapi/admin_x1.go` (신규, 3 endpoint)
3. `backend-core/internal/httpapi/admin_x1_test.go` (신규, unit test)
4. `backend-core/internal/httpapi/routes.go` (route 등록)
5. `backend-core/internal/httpapi/runtime_snapshot_provider.go` (선택, GetProviderHealthSummary 보강)
6. `backend-core/openapi/openapi.yaml` (신규 3 endpoint 정의)

### 9.2 frontend

1. `frontend/app/(dashboard)/admin/page.tsx` (강화, widget 4 import + grid)
2. `frontend/components/admin/x1-widgets/SyncJobQueueWidget.tsx` (신규)
3. `frontend/components/admin/x1-widgets/SyncJobStatusWidget.tsx` (신규)
4. `frontend/components/admin/x1-widgets/ProviderHealthWidget.tsx` (신규)
5. `frontend/components/admin/x1-widgets/DashboardSummaryWidget.tsx` (신규)
6. `frontend/components/admin/x1-widgets/index.ts` (신규)
7. `frontend/lib/api/admin-x1.ts` (신규, fetch helper)
8. `frontend/tests/e2e/admin-x1.spec.ts` (신규)

### 9.3 docs

1. `docs/adr/0032-system-admin-x1-dashboard.md` (신규)
2. `docs/traceability/report.md` (§2.4/§2.5/§2.6/§3/§4/§6 갱신)
3. `docs/llm-wiki/mirror-list.md` (§1.9 X-1 추가, scope 갱신)
4. `docs/llm-wiki/lint-config.toml` (X-1 file pattern 추가)
5. `scripts/wiki-sync-devhub.sh` (X-1 pattern 화이트리스트 갱신)
6. `CHANGELOG.md` (X-1 row)

---

## 10) 의사결정 요약

- **scope**: X-1 (RM-M4-07) backend 3 endpoint + frontend widget 4 + e2e 1 spec
- **기존 정공법 활용**: `IntegrationRepository` 의 Create/Update + `integration_sync_jobs` table + `infra_integrations.go` + `RuntimeSnapshotProvider` — List/Status method + httpapi endpoint 만 신규
- **admin catalog (3 도메인 UI)** = 이미 구현 (별도 carve, 본 sprint scope 외)
- **admin landing page 강화** = X-1 widget 4 + 기존 "운영 도구" link 통합
- **구현 원칙**: 기존 infrastructure 최대 재사용 → List/Status method 신규 → httpapi endpoint 신규 → frontend widget 4 → e2e 1 spec
- **chunk 수**: 4~5 (1300~1700줄)
- **commit 단위**: 5~8 commit, 1차 PR 로 합본

---

## 11) 일정

- **2026-06-13 (Day 1)**: sprint plan + ADR-0032 + backend IntegrationRepository method 3 + httpapi admin_x1.go + route 등록
- **2026-06-14 (Day 2)**: backend unit test + openapi.yaml + frontend widget 4 + admin page 강화
- **2026-06-15 (Day 3)**: e2e admin-x1.spec.ts + CI 4종 lint + traceability/report.md + PR 머지
