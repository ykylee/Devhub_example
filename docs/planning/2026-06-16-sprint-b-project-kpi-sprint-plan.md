# Sprint B — Project 상세 KPI / Tests sub-section (가중치 정공법)

- 문서 목적: [`kpi-tests-per-domain-scope.md`](./kpi-tests-per-domain-scope.md) 의 Sprint B — Project sub-section 의 1차 진입. Sprint A (Repository raw, weight=1) 의 후속으로, **가중치 적용 rollup (N:M linked repository)** 의 정공법 확립.
- 범위: backend 2 endpoint (`GET /api/v1/projects/:id/kpi` + `/test-results`, 가중치 적용) + frontend 2 component (`ProjectKPISection` + `ProjectTestsSection`) + `ManagerView` 또는 `ProjectView` 통합 + `routePermissionTable` 등록 + 회귀 가드.
- 상태: planned → in_progress (branch `chore/260616-sprint-b-project-kpi`)
- 작성일: 2026-06-16
- 결정 근거: [`kpi-tests-per-domain-scope.md` §2.2](./kpi-tests-per-domain-scope.md#22-project-상세-가중치-적용-rollup-nm-repo) + §6.2 + 기존 `contribution_weight` + `completionRate` 정공법 활용
- 관련 문서:
  - [`project_concept.md`](../domain/platform-lifecycle/project_concept.md) §2 운영 계층 모델
  - [`kpi-tests-per-domain-scope.md` §4.2](./kpi-tests-per-domain-scope.md#42-project--backend-api-가중치-적용) 데이터 흐름
  - [`docs/domain/platform-lifecycle/requirements.md`](../domain/platform-lifecycle/requirements.md) REQ-FR-PROJ-* + REQ-FR-APPDASH-*
  - [`docs/traceability/report.md`](../traceability/report.md) §3 application-lifecycle row + repository-integration row
  - PR #597 (Sprint A 정공법), PR #625 (Sprint A follow-up), PR #626 (Sprint D)

## 0. 컨텍스트

### 0.1 Sprint A 정공법 차용

- backend: `repositoryKPI` + `repositoryTestResults` handler 의 `parseTestResultsWindow` + `BuildRunListOptions.WindowFrom/To` SQL filter + `routePermissionTable` deny-by-default
- frontend: `RepositoryKPISection` + `RepositoryTestsSection` + `fetchRepositoryKPI` + `fetchRepositoryTestResults` service 의 `apiClient<T>` + 4 options (`windowDays` / `from`+`to` / `limit` / `from-to`)
- 정합법: raw metric, 가중치 없음, weight=1

### 0.2 Sprint B 의 차이점

- **가중치 적용**: `Σ(metric_i × contribution_weight_i) / Σ(contribution_weight_i)` 정공법 (Sprint A 의 단순 합산과 다름)
- **다중 repository 종합**: project 하위 N개 linked repository 의 metric 종합 (Sprint A 는 단일 repository)
- **completionRate 정공법 활용**: 기존 `frontend/app/(dashboard)/projects/[id]/page.tsx` 의 `completionRate` 계산 정공법 + `project_repositories.contribution_weight` 활용

### 0.3 기존 정공법 활용

- `ListRepositoryActivity` (PostgresStore) — 단일 repository 의 metric 종합
- `ListRepositoryBuildRuns` — build_runs window-bounded
- `CountOpenAndMergedPRs` (PostgresStore, Sprint A 신규) — open/merged PR 카운트
- `project_repositories.contribution_weight` 컬럼 (migration 000034) — 가중치 (default 1.0)
- `projectService.getProjectRepositories(projectId)` — linked repo + weight

## 1. 결정

### 1.1 Backend 2 endpoint

#### 1.1.1 `GET /api/v1/projects/:id/kpi?window=30d`

- **목적**: 단일 project 의 N개 linked repository 의 가중치 적용 KPI 종합
- **Response** (`200 OK`):
  ```json
  {
    "status": "ok",
    "data": {
      "project_id": "p-001",
      "window_from": "2026-05-16T00:00:00Z",
      "window_to": "2026-06-15T00:00:00Z",
      "weighted_quality_score": 87.3,
      "weighted_build_success_rate": 0.94,
      "total_build_run_count": 156,
      "open_pr_count": 7,
      "merged_pr_count": 23,
      "active_contributor_count": 12,
      "linked_repository_count": 3,
      "weighted_at": "2026-06-15T00:00:00Z"
    }
  }
  ```
- **권한**: `routePermissionTable` — `Resource: domain.ResourcePlatformProjects, Action: domain.ActionView`
- **가중치 정공법**: `metric_i = repository_i 의 raw metric`, `weighted = Σ(metric_i × contribution_weight_i) / Σ(contribution_weight_i)` (linked repo 0건 → null/0)
- **window 파라미터**: `7d` / `30d` / `90d` / `1y` / RFC3339 from/to. default 30d
- **에러**:
  - `400 project_id required` (id parse 실패)
  - `400 window must be 7d|30d|90d|1y or from/to RFC3339`
  - `403 auth.policy_unmapped` (routePermissionTable 미매핑 — deny-by-default)
  - `500 project.kpi.weighted_quality` 등

#### 1.1.2 `GET /api/v1/projects/:id/test-results?window=30d&limit=20`

- **목적**: 단일 project 의 가중치 적용 test results 종합 + recent runs 통합
- **Response** (`200 OK`):
  ```json
  {
    "status": "ok",
    "data": {
      "project_id": "p-001",
      "window_from": "2026-05-16T00:00:00Z",
      "window_to": "2026-06-15T00:00:00Z",
      "weighted_pass_rate": 0.93,
      "totals": {
        "success": 145,
        "failed": 8,
        "running": 1,
        "cancelled": 2,
        "skipped": 0,
        "queued": 0,
        "unknown": 0
      },
      "recent": [
        { "id": 100, "repository_id": 1, "repository_full_name": "org/repo-a", "run_external_id": "ext-100", "commit_sha": "feedface", "status": "success", "branch": "main", "started_at": "...", "finished_at": "..." }
      ]
    },
    "meta": { "total": 156, "limit": 20 }
  }
  ```
- **권한**: 동일
- **weighted_pass_rate 정공법**: `Σ(success_i × weight_i) / Σ((success_i + failed_i) × weight_i)` (denom 0 → null)
- **totals**: N개 repo 의 build_runs status 합산 (가중치 무관 — 단순 카운트)
- **recent**: 모든 linked repo 의 build_runs 의 최신순 limit 개

### 1.2 Backend 가중치 SQL 정공법

- `ListProjectRepositories` (기존) — linked repo + weight
- `ListProjectKPI` (NEW) — `weighted_quality_score` + `weighted_build_success_rate`:
  ```sql
  SELECT
    COALESCE(SUM(qs.score * pr.contribution_weight) / NULLIF(SUM(pr.contribution_weight), 0), 0) AS weighted_quality_score,
    COALESCE(SUM(activity.build_success_rate * pr.contribution_weight) / NULLIF(SUM(pr.contribution_weight), 0), 0) AS weighted_build_success_rate,
    SUM(activity.build_run_count) AS total_build_run_count,
    COALESCE(SUM(activity.active_contributors), 0) AS active_contributor_count
  FROM project_repositories pr
  JOIN repositories r ON r.id = pr.repository_id
  LEFT JOIN LATERAL (
    SELECT score, measured_at FROM quality_snapshots qs
    WHERE qs.repository_id = r.id ORDER BY measured_at DESC LIMIT 1
  ) qs ON true
  LEFT JOIN LATERAL (
    SELECT build_success_rate, build_run_count, active_contributors
    FROM repository_activity_window(...)
  ) activity ON true
  WHERE pr.project_id = $1
  ```
- `CountProjectOpenAndMergedPRs` (NEW) — `Σ(count × weight)` 가중치 적용:
  ```sql
  SELECT
    COALESCE(SUM(open_count * pr.contribution_weight), 0)::int AS weighted_open,
    COALESCE(SUM(merged_count * pr.contribution_weight), 0)::int AS weighted_merged
  FROM project_repositories pr
  LEFT JOIN LATERAL (
    SELECT
      COUNT(*) FILTER (WHERE event_type = 'opened') AS open_count,
      COUNT(*) FILTER (WHERE event_type = 'merged') AS merged_count
    FROM pr_activities pa
    WHERE pa.repository_id = pr.repository_id
      AND pa.occurred_at >= $2 AND pa.occurred_at < $3
  ) pr_stats ON true
  WHERE pr.project_id = $1
  ```

### 1.3 Frontend 2 component

#### 1.3.1 `ProjectKPISection`

- **위치**: `frontend/domain/platform-lifecycle/view/ProjectKPISection.tsx` (NEW)
- **props**: `projectId: string`
- **내부**: `useEffect` 로 `GET /api/v1/projects/:id/kpi?window=30d` fetch. `apiClient` 정공법
- **표시**:
  - Weighted Quality Score (큰 카드, 색상 코드)
  - Weighted Build Success Rate
  - Total Build Run Count
  - Weighted Open PR Count + Weighted Merged PR Count
  - Active Contributors
  - Linked Repository Count
  - Window selector (7d/30d/90d)
  - **가중치 라벨**: 모든 metric 옆에 "(weighted)" 텍스트
- **shadcn/ui + Recharts** 정공법 (기존 `RepositoryKPISection` 와 동일)

#### 1.3.2 `ProjectTestsSection`

- **위치**: `frontend/domain/platform-lifecycle/view/ProjectTestsSection.tsx` (NEW)
- **props**: `projectId: string`
- **내부**: `useEffect` 로 `GET /api/v1/projects/:id/test-results?window=30d&limit=20` fetch
- **표시**:
  - Weighted Pass Rate (도넛)
  - Status Distribution (5 status 합산)
  - Recent Runs (table) — repository_full_name 컬럼 추가 (multi-repo)
  - Window selector + limit
- **Recharts 도넛** 정공법 (기존 `/tests/page.tsx` 와 동일)

### 1.4 ProjectView 통합

`frontend/app/(dashboard)/projects/[id]/page.tsx` (719 line) 의 메인 콘텐츠에 KPI/Tests sub-section 추가. 기존 `editingWeights` / `projectRepositories` / `completionRate` 옆에 배치.

### 1.5 신규 ID (3 row)

- `REQ-FR-PROJKPI-001` (Project KPI/Tests sub-section) — 1
- `ARCH-PROJ-01` (Project sub-section layout) — 1
- `API-XXX` (2 endpoint) — 2 (`/kpi` + `/test-results`)
- `IMPL-XXX` (frontend component 2 + service 2 + 가중치 SQL) — 5
- `UT-XXX` (unit test 8+) — 2
- `TC-XXX` (e2e 1 + Vitest component 4) — 3
- **합계 ~14 ID**

## 2. 변경 범위 (PR 1개, ~1200 line)

### 2.1 backend (5 file)

1. `backend-core/internal/httpapi/projects.go` (MODIFY, +200 line) — `projectKPI` + `projectTestResults` handler
2. `backend-core/internal/store/repository_ops.go` (MODIFY, +80 line) — `ListProjectRepositories` 의 `ListProjectWeightedKPI` SQL + `CountProjectOpenAndMergedPRs` SQL
3. `backend-core/internal/httpapi/router.go` (MODIFY, +3 line) — 2 path 등록
4. `backend-core/internal/httpapi/projects_test.go` (MODIFY, +80 line) — handler test 2 case
5. `backend-core/internal/httpapi/swaggerui/asset/openapi.yaml` (MODIFY, +200 line) — 2 path + schemas 2

### 2.2 frontend (8 file)

6. `frontend/domain/platform-lifecycle/schema/project-kpi.types.ts` (NEW, ~40 line)
7. `frontend/domain/platform-lifecycle/schema/project-tests.types.ts` (NEW, ~50 line)
8. `frontend/domain/platform-lifecycle/service/project-kpi.service.ts` (NEW, ~50 line)
9. `frontend/domain/platform-lifecycle/service/project-tests.service.ts` (NEW, ~50 line)
10. `frontend/domain/platform-lifecycle/view/ProjectKPISection.tsx` (NEW, ~200 line)
11. `frontend/domain/platform-lifecycle/view/ProjectTestsSection.tsx` (NEW, ~250 line)
12. `frontend/app/(dashboard)/projects/[id]/page.tsx` (MODIFY, +30 line) — KPI/Tests sub-section 배치
13. `frontend/shared/ui-foundation/components/DomainPicker.tsx` (MODIFY, +5 line) — projects fetch 활성화
14. `frontend/app/(dashboard)/kpis/page.tsx` (MODIFY, +20 line) — projects fetch useEffect
15. `frontend/app/(dashboard)/tests/page.tsx` (MODIFY, +20 line) — projects fetch useEffect

### 2.3 test (5 file)

16. `frontend/domain/platform-lifecycle/service/__tests__/project-kpi.service.test.ts` (NEW, ~120 line)
17. `frontend/domain/platform-lifecycle/service/__tests__/project-tests.service.test.ts` (NEW, ~120 line)
18. `frontend/domain/platform-lifecycle/view/__tests__/ProjectKPISection.test.tsx` (NEW, ~150 line)
19. `frontend/domain/platform-lifecycle/view/__tests__/ProjectTestsSection.test.tsx` (NEW, ~170 line)
20. `frontend/tests/e2e/project-kpi-tests-section.spec.ts` (NEW, ~120 line)

### 2.4 docs (3 file)

21. `docs/planning/2026-06-16-sprint-b-project-kpi-sprint-plan.md` (NEW, 본 문서)
22. `docs/planning/kpi-tests-per-domain-scope.md` (MODIFY) — §6.2 status "planned" → "in_progress" + 변경 이력 row
23. `docs/traceability/report.md` (MODIFY, +1 row) — §6

## 3. 검증

- `go build ./...` PASS
- `go test ./internal/httpapi/...` 2 신규 handler test PASS
- `go test ./...` 회귀 0
- `bash scripts/check-openapi-yaml-lint.sh` PASS
- `cd frontend && npm run test` — 100+ unit test (8+ 신규) PASS
- `cd frontend && npm run build` — silent PASS
- `npx playwright test project-kpi-tests-section.spec.ts` PASS
- e2e shard 1/2/3 path-detect: frontend 변경 → trigger
- CI 4/4 PASS 예상

## 4. Tier

- **공용** (코드 + openapi + docs 모두 사내 한정 정보 미포함)

## 5. 잔여 (Sprint C / D / E)

- Sprint C: Platform sub-section (sub-project rollup)
- Sprint D 보완: DomainPicker 의 platforms fetch 활성화
- Sprint E: 글로벌 페이지 옵션 A/B/C 결정 + legacy 본문 처리

## 6. 다음 세션 directive

- PR 머지 후 main flat memory 3 file finalize
- 위키 mirror 갱신 (`bash scripts/wiki-sync-devhub.sh`)
- Sprint C 진입 시 본 Sprint B 의 가중치 SQL 정공법 차용 + project 의 sub-project rollup 으로 확장

## 7. 변경 이력

| 일자 | 변경 | sprint |
|---|---|---|
| 2026-06-16 | 본 sprint plan 초안 (Sprint B — Project sub-section, 가중치 적용) | `chore/260616-sprint-b-project-kpi` |
