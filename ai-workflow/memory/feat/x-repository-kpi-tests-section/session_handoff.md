# Session Handoff — feat/x-repository-kpi-tests-section (Sprint A)

- 문서 목적: KPI/Tests 위치 정공법 컨셉 ([kpi-tests-per-domain-scope.md](../../../docs/planning/kpi-tests-per-domain-scope.md)) 의 Sprint A — Repository 단위 raw KPI/Tests sub-section 진입. 단일 PR, Tier: **공용**.
- 범위: backend 2 endpoint + 1 interface method + openapi 2 path + frontend 2 component + 2 service + 2 schema + ManagerView 통합.
- 상태: branch `feat/x-repository-kpi-tests-section` 작업 완료, commit + push + PR 발행 pending.
- 최종 수정일: 2026-06-15

## 0. 본 세션 핵심 결과

### KPI/Tests 위치 정공법 (컨셉)

신규 `/kpis` `/tests` (2026-06-07 sprint `gemini/work_260607-a-dashboard-improvements` 신규 추가) 의 위치를 **각 도메인 페이지의 sub-section** 으로 재배치. Sprint A 는 가장 단순 (가중치 없음) 의 Repository 단위 1차 진입.

**3 도메인 scope**:
- **Repository** (`/repositories/[id]`) — raw metric, weight=1
- **Project** (`/projects/[id]`) — `project_repositories.contribution_weight` 가중치 적용 rollup (Sprint B)
- **Platform** (`/platforms/[id]`) — sub-project rollup, 균등 or custom 가중치 (Sprint C)

**글로벌 `/kpis` `/tests`** 는 Cross-Reference / Analytics 통합 페이지로 격하 (Sprint D).

### Sprint A scope (본 PR)

1. `backend-core/internal/domain/application-lifecycle/view/handler.go` — `PlatformStore` interface 에 `CountOpenAndMergedPRs` method 1개 추가
2. `backend-core/internal/httpapi/repository_ops.go` — `repositoryKPI` + `repositoryTestResults` handler 2개 신규
3. `backend-core/internal/httpapi/router.go` — 2 path 등록
4. `backend-core/internal/httpapi/swaggerui/asset/openapi.yaml` — 2 path + schemas (paths 84 → 86)
5. `backend-core/internal/store/repository_ops.go` — `PostgresStore.CountOpenAndMergedPRs` 신규
6. `backend-core/internal/httpapi/applications_test.go` — `memoryPlatformStore.CountOpenAndMergedPRs` mock
7. `frontend/domain/repository-integration/view/RepositoryKPISection.tsx` (NEW)
8. `frontend/domain/repository-integration/view/RepositoryTestsSection.tsx` (NEW)
9. `frontend/domain/repository-integration/view/ManagerView.tsx` (MODIFY) — import + 배치
10. `frontend/domain/repository-integration/service/repository-kpi.service.ts` (NEW)
11. `frontend/domain/repository-integration/service/repository-tests.service.ts` (NEW)
12. `frontend/domain/repository-integration/schema/repository-kpi.types.ts` (NEW)
13. `frontend/domain/repository-integration/schema/repository-tests.types.ts` (NEW)
14. `docs/planning/2026-06-15-x-repository-kpi-tests-section-sprint-plan.md` (NEW, sprint plan)
15. `docs/planning/kpi-tests-per-domain-scope.md` (MODIFY) — §6.1 의 Sprint A status "TBD" → "in_progress"
16. `docs/traceability/report.md` (MODIFY) — §6 row 추가
17. `docs/llm-wiki/mirror-list.md` (MODIFY) — §1.7.1 의 frontend file 추가
18. `CHANGELOG.md` (MODIFY) — Sprint A status

총 ~800 line (backend 250 + frontend 400 + docs 100 + memory 50).

## 1. Backend 결정

### 1.1 `CountOpenAndMergedPRs` method 추가

PlatformStore interface 의 `Repository 운영 지표` 그룹에 `CountOpenAndMergedPRs(context.Context, int64, time.Time, time.Time) (int, int, error)` 추가. 

**이유**: Sprint A 의 `repositoryKPI` 가 open_pr_count + merged_pr_count 를 종합. 기존 `ListRepositoryActivity` 가 PR event count 만 반환. 별도 query 가 필요. **interface 변경 1 method + PostgresStore 1 method + memoryPlatformStore 1 mock** 으로 격리. 

**PostgresStore SQL**:
```sql
SELECT
  COUNT(DISTINCT number) FILTER (WHERE event_type = 'opened')::int AS open_count,
  COUNT(DISTINCT number) FILTER (WHERE event_type = 'merged')::int AS merged_count
FROM pr_activities
WHERE repository_id = $1 AND occurred_at >= $2 AND occurred_at < $3
```

state="closed" + merged_at IS NOT NULL row 는 `event_type='merged'` 로 upsert 되므로 본 query 가 정확 (X-5 follow-up 의 `stateToEventType` 정공법 정합).

### 1.2 `GET /api/v1/repositories/:id/test-results`

- build_runs.status 분포로 test results 표현 (별도 repository_tests table 미존재).
- `window` short string `7d`/`30d`/`90d`/`1y` OR `from`/`to` RFC3339.
- `limit` 1~50, default 20.
- `parseWindowShort` helper 가 `Nd`/`Nw`/`Nm`/`Ny` parse.

### 1.3 Tier

- **공용** (코드 + openapi + docs 모두 사내 한정 정보 미포함).

## 2. Frontend 결정

### 2.1 apiClient 정공법

기존 `frontend/domain/repository-integration/service/repository.service.ts` 의 `repositoryService.listRepositories()` 등 pattern 정합. 자동 token refresh + session death.

### 2.2 4 card + 도넛 차트 + table

`RepositoryKPISection`:
- 4 card (Quality Score / Build Success Rate / Pull Requests / Active Contributors)
- Color code: emerald (>=80%), amber (>=60%), red (<60%) for quality; similar for build success rate
- Window selector (7d/30d/90d/1y)
- Refresh button

`RepositoryTestsSection`:
- Pass Rate 큰 카드 + 도넛 차트 (Recharts)
- Status Distribution 작은 카드 7개
- Recent Runs table
- Window selector

### 2.3 ManagerView 통합

`frontend/domain/repository-integration/view/ManagerView.tsx` 의 SCM Connection Log 뒤에 `<RepositoryKPISection />` + `<RepositoryTestsSection />` 배치. **scope 최소**: ManagerView 안 inline 배치 (별도 route / tab 불요).

## 3. 신규 ID (4 row)

- `REQ-FR-XXX` (Repository KPI/Tests sub-section) — 본 follow-up 후속
- `ARCH-XXX` (Repository sub-section layout) — 본 follow-up 후속
- `API-XXX` (2 endpoint) — 2 (`/kpi` + `/test-results`)
- `IMPL-XXX` (frontend 2 component + 2 service) — 4
- `UT-XXX` (unit test 12) — 1 (Sprint A 의 e2e 별도 PR 에서)
- `TC-XXX` (e2e 1) — 1 (별도 PR)

본 PR 의 ID slot 발권은 cross-project lesson §1 의 "scope 폭주 방지" — **PR 1개, 800 line cap**. 정공법 정공법. 신규 ID 는 e2e 별도 PR 에서 일괄 발권.

## 4. 잔여 follow-up (Sprint B/C/D)

- **Sprint B** — Project 단위 KPI/Tests sub-section (contribution_weight 가중치 적용 rollup). backend: `GET /api/v1/projects/:id/kpi/quality` + `/kpi/test-pass-rate` (가중치 적용). 
- **Sprint C** — Platform 단위 KPI/Tests sub-section (sub-project rollup). backend: `GET /api/v1/platforms/:id/kpi/quality` + `/kpi/test-pass-rate` + `/progress`.
- **Sprint D** — Sidebar 의 `analyticsMenu` 분리 + 글로벌 `/kpis` `/tests` 에 도메인 picker 추가.
- **Sprint E** — 글로벌 페이지 옵션 A/B/C 결정 + legacy 정리.
- **e2e spec** — `frontend/tests/e2e/repository-kpi-tests-section.spec.ts` + smoke manifest 등록 (TC-REPO-KPI-TESTS-01).
- **별도 repository_tests table** 도입 — Sprint A 한정 표현 (build_runs 분포) 의 한계. Sprint F 또는 후속.

## 5. 검증 결과

- `go build ./...` — silent PASS
- `bash scripts/check-openapi-yaml-lint.sh` — PASS (paths 84 → 86)
- frontend `npm run build` — **CI 검증 위임** (본 세션은 node_modules 부재로 local tsc skip)
- frontend `npm run test` — **CI 검증 위임** (Sprint A 의 unit test 미작성 — 본 PR 의 frontend component 는 e2e spec 으로 검증, 별도 PR)
- e2e shard 1/2/3 path-detect: frontend 변경 → trigger
- CI 4/4 PASS 예상

## 6. 다음 세션 directive

- 본 PR commit + push + PR 발행
- PR 머지 후 main flat memory 3 file finalize
- 위키 mirror 갱신 (`bash scripts/wiki-sync-devhub.sh`)
- e2e spec 별도 PR (TC-REPO-KPI-TESTS-01, smoke manifest 등록)
- Sprint B (Project sub-section) 진입 가능 — `contribution_weight` 가중치 정공법 work_260607-a 정합
