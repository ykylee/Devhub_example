# Sprint A — Repository 상세 KPI / Tests sub-section (2026-06-15)

- 문서 목적: KPI/Tests 위치 정공법 컨셉 ([`kpi-tests-per-domain-scope.md`](./kpi-tests-per-domain-scope.md)) 의 Sprint A — Repository 단위 raw KPI/Tests sub-section 진입. 단일 PR, Tier: 공용.
- 범위: backend 2 endpoint (`GET /api/v1/repositories/:id/kpi` + `GET /api/v1/repositories/:id/test-results`) + frontend 2 component (`RepositoryKPISection` + `RepositoryTestsSection`) + RepositoryDashboardView import + 통합 정합.
- 상태: in_progress (branch `feat/x-repository-kpi-tests-section` 의 worktree 작업 중)
- 결정 근거 컨셉: [`kpi-tests-per-domain-scope.md` §2.1](./kpi-tests-per-domain-scope.md#21-repository-상세-raw-weight1) + [`kpi-tests-per-domain-scope.md` §6.1](./kpi-tests-per-domain-scope.md#61-sprint-a-repository-sub-section-1차)
- 관련 문서: [`project_concept.md` §2](../domain/platform-lifecycle/project_concept.md) 운영 계층 모델, [`dashboard_concept.md`](../domain/platform-lifecycle/dashboard_concept.md), [`kpi-tests-per-domain-scope.md` §4.1](./kpi-tests-per-domain-scope.md#41-repository--backend-api) 데이터 흐름

## 0. 컨텍스트

### 0.1 컨셉 정공법 위치

Sprint A 는 `kpi-tests-per-domain-scope.md` 의 **3 단계 sub-section** 중 가장 단순 (가중치 없음, 단일 repo) 의 1차 진입. Sprint B (Project 가중치) + Sprint C (Platform sub-rollup) 의 기반이 됨.

### 0.2 기존 정공법 활용

- `ListRepositoryActivity` (PostgresStore) — quality_snapshots + build_runs + pr_activities 종합. handler 측 1 query 추가로 KPI 산출.
- `ListRepositoryBuildRuns` — test-results API 의 build_runs 분포 source.
- `RepositoryDashboardView` (frontend thin shell) — sub-section 추가 자리가 명확.
- `frontend/domain/repository-integration/view/` — 신규 component 위치.

## 1. 결정

### 1.1 Backend 2 endpoint

#### 1.1.1 `GET /api/v1/repositories/:id/kpi?window=30d`

- **목적**: 단일 repository 의 raw KPI 종합 (가중치 없음, weight=1).
- **Response** (`200 OK`):
  ```json
  {
    "status": "ok",
    "data": {
      "repository_id": 101,
      "window_from": "2026-05-16T00:00:00Z",
      "window_to": "2026-06-15T00:00:00Z",
      "quality_score": 87.3,
      "quality_score_measured_at": "2026-06-14T22:00:00Z",
      "build_success_rate": 0.94,
      "build_run_count": 47,
      "open_pr_count": 3,
      "merged_pr_count": 12,
      "active_contributor_count": 4
    }
  }
  ```
- **권한**: RBAC 정합 (기존 `repositoryActivity` 와 동일) — `enforceRoutePermission` 으로 보호.
- **에러**:
  - `400 repository_id must be an integer`
  - `503 platform_store_or_unavailable` (기존 패턴 정합)
  - `500 repository.kpi` (writeServerError)
- **window 파라미터**: `7d` / `30d` / `90d` / `1y` / RFC3339 from/to. default 30d. 기존 `ListRepositoryActivity` 의 `WindowFrom` / `WindowTo` 정합.

#### 1.1.2 `GET /api/v1/repositories/:id/test-results?window=30d`

- **목적**: 단일 repository 의 test/build 결과 분포 + 최근 N 회 상세.
- **Response** (`200 OK`):
  ```json
  {
    "status": "ok",
    "data": {
      "repository_id": 101,
      "window_from": "...",
      "window_to": "...",
      "totals": {
        "success": 42,
        "failed": 3,
        "running": 1,
        "cancelled": 1,
        "skipped": 0
      },
      "pass_rate": 0.875,
      "recent": [
        { "id": 100, "run_external_id": "...", "commit_sha": "feedface", "status": "success", "branch": "main", "started_at": "...", "finished_at": "..." }
      ]
    },
    "meta": { "total": 47, "limit": 20 }
  }
  ```
- **권한**: 동일.
- **limit 파라미터**: 1~50, default 20.

### 1.2 Frontend 2 component

#### 1.2.1 `RepositoryKPISection`

- **위치**: `frontend/domain/repository-integration/view/RepositoryKPISection.tsx` (NEW)
- **props**: `repoId: number`
- **내부**: `useEffect` 로 `GET /api/v1/repositories/:id/kpi?window=30d` fetch. **apiClient** 정공법 (자동 token refresh + session death).
- **표시**:
  - Quality Score (큰 카드, 색상 코드)
  - Build Success Rate (작은 카드 + progress bar)
  - Open PR Count (작은 카드)
  - Merged PR Count (작은 카드)
  - Active Contributors (작은 카드)
  - Window selector (7d/30d/90d) — local state
- **shadcn/ui** 정공법 (기존 `frontend/shared/ui-foundation/components/`).
- **로딩/에러**: 기존 `PageError` / `PageLoading` 정공법.
- **mock fallback**: `mockData` 가 있으면 mockData, 없으면 빈 상태.

#### 1.2.2 `RepositoryTestsSection`

- **위치**: `frontend/domain/repository-integration/view/RepositoryTestsSection.tsx` (NEW)
- **props**: `repoId: number`
- **내부**: `useEffect` 로 `GET /api/v1/repositories/:id/test-results?window=30d` fetch.
- **표시**:
  - Pass Rate (큰 도넛 차트, Recharts)
  - Status 분포 (작은 카드 4-5개)
  - Recent Runs 리스트 (table)
  - Window selector
- **shadcn/ui** + Recharts 정공법 (기존 `/tests/page.tsx` 와 동일).

### 1.3 RepositoryDashboardView import

`frontend/domain/repository-integration/view/RepositoryDashboardView.tsx` 의 기존 dashboard metric + PR list + commit list 사이에 **KPI/Tests sub-section 탭 or inline 배치** — **탭 방식** (KPI / Tests / Overview) 권장. inline 배치는 페이지 길이 폭주.

### 1.4 신규 ID (4 row)

- `REQ-FR-XXX` (Repository KPI/Tests sub-section) — 1
- `ARCH-XXX` (Repository sub-section layout) — 1
- `API-XXX` (2 endpoint) — 2 (`/kpi` + `/test-results`)
- `IMPL-XXX` (frontend component 2 + service 2) — 4
- `UT-XXX` (unit test 12) — 1
- `TC-XXX` (e2e 1) — 1
- **합계 ~10 ID** (각각 1씩, 기존 1차 PR 의 X-5 follow-up 의 4 ID 와 동일 pattern).

## 2. 변경 범위 (PR 1개, ~700 line)

### 2.1 backend (1 file modify)
1. `backend-core/internal/httpapi/repository_ops.go` (MODIFY, +130 line)
   - `repositoryKPI` handler (line 248 이후)
   - `repositoryTestResults` handler

### 2.2 openapi (1 file modify)
2. `backend-core/internal/httpapi/swaggerui/asset/openapi.yaml` (MODIFY, +80 line)
   - `/api/v1/repositories/{repository_id}/kpi` GET
   - `/api/v1/repositories/{repository_id}/test-results` GET
   - schemas 2 (RepositoryKPIResponse, RepositoryTestResultsResponse)

### 2.3 backend test (1 file modify)
3. `backend-core/internal/httpapi/applications_test.go` (MODIFY, +30 line)
   - Test handler 2 case (memoryPlatformStore mock)

### 2.4 frontend (6 file)
4. `frontend/domain/repository-integration/view/RepositoryKPISection.tsx` (NEW, ~150 line)
5. `frontend/domain/repository-integration/view/RepositoryTestsSection.tsx` (NEW, ~150 line)
6. `frontend/domain/repository-integration/view/RepositoryDashboardView.tsx` (MODIFY, +30 line) — tab 추가 + import
7. `frontend/domain/repository-integration/service/repository-kpi.service.ts` (NEW, ~40 line) — API client
8. `frontend/domain/repository-integration/service/repository-tests.service.ts` (NEW, ~40 line)
9. `frontend/domain/repository-integration/schema/repository-kpi.types.ts` (NEW, ~30 line)
10. `frontend/domain/repository-integration/schema/repository-tests.types.ts` (NEW, ~30 line)
11. `frontend/domain/repository-integration/view/__tests__/RepositoryKPISection.test.tsx` (NEW, ~50 line)
12. `frontend/domain/repository-integration/view/__tests__/RepositoryTestsSection.test.tsx` (NEW, ~50 line)

### 2.5 e2e (1 file NEW)
13. `frontend/tests/e2e/repository-kpi-tests-section.spec.ts` (NEW, ~80 line)
    - TC-REPO-KPI-TESTS-01: Repository 상세 진입 → KPI/Tests 탭 클릭 → metric 정상 표시

### 2.6 docs (4 file)
14. `docs/planning/2026-06-15-x-repository-kpi-tests-section-sprint-plan.md` (NEW, 본 문서)
15. `docs/planning/kpi-tests-per-domain-scope.md` (MODIFY) — §6.1 의 Sprint A status "TBD" → "in_progress"
16. `docs/traceability/report.md` (MODIFY, +1 row) — §6
17. `docs/llm-wiki/mirror-list.md` (MODIFY) — §1.7.1 의 신규 frontend file 추가
18. `CHANGELOG.md` (MODIFY) — Sprint A status

### 2.7 메모리 (4 file)
19-22. `ai-workflow/memory/feat/x-repository-kpi-tests-section/{state.json, session_handoff.md, work_backlog.md, backlog/2026-06-15.md}` (NEW)

## 3. 검증

- `go build ./...` PASS
- `go test ./internal/httpapi/...` 2 신규 handler test PASS
- `go test ./...` 회귀 0
- `bash scripts/check-openapi-yaml-lint.sh` PASS
- `cd frontend && npm run test` — 100+ unit test (12+ 신규) PASS
- `cd frontend && npm run build` — silent PASS
- `npx playwright test repository-kpi-tests-section.spec.ts` PASS
- e2e shard 1/2/3 path-detect: frontend 변경 → trigger
- CI 4/4 PASS 예상

## 4. Tier

- **공용** (코드 + openapi + migration 0 + docs 모두 사내 한정 정보 미포함)

## 5. 잔여 (Sprint B/C/D/E)

본 PR 은 Sprint A 만. Sprint B (Project 가중치) / C (Platform sub-rollup) / D (Sidebar 재구성) / E (legacy 결정) 는 후속 PR.

## 6. 다음 세션 directive

- PR 머지 후 main flat memory 3 file finalize
- 위키 mirror 갱신 (`bash scripts/wiki-sync-devhub.sh`)
- 다음 sprint 진입 시 main HEAD rebase
- Sprint B (Project sub-section) 진입 가능 (가중치 정공법 work_260607-a 정공법 활용)

## 7. 변경 이력

| 일자 | 변경 | sprint |
|---|---|---|
| 2026-06-15 | 본 sprint plan 초안 (Sprint A — Repository KPI/Tests sub-section) | `feat/x-repository-kpi-tests-section` |
