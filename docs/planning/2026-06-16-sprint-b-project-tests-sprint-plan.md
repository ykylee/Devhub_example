# Sprint B-Tests — Project 상세 Tests sub-section (가중치 종합)

- 문서 목적: [`kpi-tests-per-domain-scope.md`](./kpi-tests-per-domain-scope.md) 의 Sprint B §6.2 follow-up. Sprint B 1차 PR #627 (projectKPI + ProjectKPISection) 의 후속으로, **가중치 적용 Tests sub-section** 의 정공법 확립.
- 범위: backend 1 endpoint (`GET /api/v1/projects/:id/test-results`) + frontend 1 component (`ProjectTestsSection`) + `ManagerView` 또는 `ProjectView` 통합 + `routePermissionTable` 등록 + 회귀 가드 + e2e 1 case.
- 상태: planned → in_progress (branch `chore/260616-sprint-b-project-tests`)
- 작성일: 2026-06-16
- 결정 근거: [`kpi-tests-per-domain-scope.md` §2.2](./kpi-tests-per-domain-scope.md#22-project-상세-가중치-적용-rollup-nm-repo) + §6.2 + Sprint A 의 `repositoryTestResults` + `RepositoryTestsSection` 정공법 차용
- 관련 문서:
  - [`project_concept.md`](../domain/platform-lifecycle/project_concept.md) §2 운영 계층 모델
  - [`kpi-tests-per-domain-scope.md` §4.2](./kpi-tests-per-domain-scope.md#42-project--backend-api-가중치-적용) 데이터 흐름
  - [`docs/domain/platform-lifecycle/requirements.md`](../domain/platform-lifecycle/requirements.md) REQ-FR-PROJ-*
  - [`docs/traceability/report.md`](../traceability/report.md) §3 application-lifecycle row + repository-integration row
  - [PR #627 sprint plan](./2026-06-16-sprint-b-project-kpi-sprint-plan.md) (Sprint B 1차)
  - PR #597 (Sprint A 정공법), PR #625 (Sprint A follow-up), PR #626 (Sprint D), PR #627 (Sprint B 1차)

## 0. 컨텍스트

### 0.1 Sprint B 1차 PR #627 의 follow-up

- PR #627 가 `projectKPI` endpoint + `ProjectKPISection` 만 작업. 본 sprint 는 follow-up 인 Sprint B-Tests:
  - `projectTestResults` handler (build_runs status 가중치 종합)
  - `ProjectTestsSection` component (Recharts 도넛 + recent + window)
  - e2e 1 case

### 0.2 Sprint A 정공법 차용

- backend: `repositoryTestResults` handler 의 `parseTestResultsWindow` + `BuildRunListOptions.WindowFrom/WindowTo` SQL filter + `routePermissionTable` deny-by-default
- frontend: `RepositoryTestsSection` + `fetchRepositoryTestResults` service + `repository-tests.types` schema + Recharts 도넛 + 4 status distribution + recent runs table
- 정합법: build_runs 기반 분포, raw metric, 가중치 없음 (단일 repo)

### 0.3 Sprint B-Tests 의 차이점

- **가중치 적용**: `weighted_pass_rate = Σ(success_i × weight_i) / Σ((success_i + failed_i) × weight_i)` 정공법
- **다중 repository 종합**: project 하위 N개 linked repository 의 build_runs 통합 (Sprint A 는 단일 repository)
- **totals 가중치 무관**: N개 repo 의 build_runs status 합산 (가중치 무관 — 단순 카운트, totals 의 정의 자체가 raw count)
- **recent 의 repository_full_name 표시**: 모든 linked repo 의 build_runs 최신순 limit + repository_full_name 컬럼 추가 (multi-repo)
- **denom=0 케이스**: linked repo 의 모든 build 가 success/failed 0 → pass_rate = null (가중치 denom 도 0)

## 1. 결정

### 1.1 Backend 1 endpoint

#### 1.1.1 `GET /api/v1/projects/:project_id/test-results?window=30d&limit=20`

- **목적**: 단일 project 의 가중치 적용 test pass rate + N개 linked repository 의 build_runs status 합산 + recent runs
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
        {
          "id": 100,
          "repository_id": 1,
          "repository_full_name": "org/repo-a",
          "run_external_id": "ext-100",
          "commit_sha": "feedface",
          "status": "success",
          "branch": "main",
          "started_at": "2026-06-15T01:00:00Z",
          "finished_at": "2026-06-15T01:02:00Z"
        }
      ]
    },
    "meta": { "total": 156, "limit": 20 }
  }
  ```
- **권한**: `routePermissionTable` — `Resource: domain.ResourceProjects, Action: domain.ActionView` (projectKPI 와 동일)
- **weighted_pass_rate 정공법**:
  - denom = Σ((success_i + failed_i) × contribution_weight_i)
  - num = Σ(success_i × contribution_weight_i)
  - denom=0 → null
  - linked_repository_count=0 → null (Σ weight=0)
- **totals**: N개 repo 의 build_runs status 합산 (가중치 무관 — 단순 카운트, 7 status 모두 0 초기화)
- **recent**: 모든 linked repo 의 build_runs 의 최신순 limit 개 (multi-repo, repository_full_name 표시)
- **window 파라미터**: `7d` / `30d` / `90d` / `1y` / RFC3339 from/to. default 30d
- **limit**: 1..50, default 20
- **에러**:
  - `400 project_id required` (id parse 실패)
  - `400 window must be 7d|30d|90d|1y or from/to RFC3339`
  - `400 limit must be 1..50`
  - `403 auth.policy_unmapped` (routePermissionTable 미매핑)
  - `500 project.tests.weighted` 등

### 1.2 Backend 가중치 SQL 정공법

- `ListProjectTestResults` (NEW) — N개 linked repository 의 build_runs 통합 + 가중치 pass_rate + recent:
  ```sql
  WITH linked_repos AS (
    SELECT pr.repository_id, pr.contribution_weight, r.full_name
    FROM project_repositories pr
    JOIN repositories r ON r.id = pr.repository_id
    WHERE pr.project_id = $1::uuid
  ),
  -- 1) weighted_pass_rate: Σ(success × weight) / Σ((success+failed) × weight)
  weighted AS (
    SELECT
      COALESCE(SUM(success_count * lr.contribution_weight), 0) AS num,
      COALESCE(SUM((success_count + failed_count) * lr.contribution_weight), 0) AS denom
    FROM linked_repos lr
    LEFT JOIN LATERAL (
      SELECT
        COUNT(*) FILTER (WHERE status = 'success')::int AS success_count,
        COUNT(*) FILTER (WHERE status = 'failed')::int AS failed_count
      FROM build_runs br
      WHERE br.repository_id = lr.repository_id
        AND COALESCE(br.started_at, br.created_at) >= $2
        AND COALESCE(br.started_at, br.created_at) < $3
    ) stats ON true
  ),
  -- 2) totals: N개 repo 의 build_runs status 합산
  totals AS (
    SELECT status, COUNT(*)::int AS cnt
    FROM build_runs br
    WHERE br.repository_id IN (SELECT repository_id FROM linked_repos)
      AND COALESCE(br.started_at, br.created_at) >= $2
      AND COALESCE(br.started_at, br.created_at) < $3
    GROUP BY status
  ),
  -- 3) recent: 모든 linked repo 의 build_runs 최신순 limit
  recent AS (
    SELECT
      br.id, br.repository_id, r.full_name AS repository_full_name,
      br.external_id AS run_external_id, br.branch, br.commit_sha,
      br.status, br.duration_seconds, br.started_at, br.finished_at
    FROM build_runs br
    JOIN repositories r ON r.id = br.repository_id
    WHERE br.repository_id IN (SELECT repository_id FROM linked_repos)
      AND COALESCE(br.started_at, br.created_at) >= $2
      AND COALESCE(br.started_at, br.created_at) < $3
    ORDER BY COALESCE(br.started_at, br.created_at) DESC
    LIMIT $4
  )
  SELECT
    (SELECT num FROM weighted) AS weighted_num,
    (SELECT denom FROM weighted) AS weighted_denom,
    (SELECT json_agg(json_build_object('status', status, 'count', cnt)) FROM totals) AS totals_json,
    (SELECT json_agg(json_build_object(
      'id', id, 'repository_id', repository_id, 'repository_full_name', repository_full_name,
      'run_external_id', run_external_id, 'branch', branch, 'commit_sha', commit_sha,
      'status', status, 'duration_seconds', duration_seconds,
      'started_at', started_at, 'finished_at', finished_at
    )) FROM recent) AS recent_json
  ```
  - linked repo 0 → num/denom = 0, totals = 빈 array, recent = 빈 array
  - `denom = 0` 인 경우 store 는 `pass_rate = nil` 반환, handler 가 null 직렬화
  - handler 가 totals/denom/num/recent 를 조합해 `domain.ProjectWeightedTestResults` build

### 1.3 Frontend 1 component

#### 1.3.1 `ProjectTestsSection`

- **위치**: `frontend/domain/platform-lifecycle/view/ProjectTestsSection.tsx` (NEW)
- **props**: `projectId: string`
- **내부**: `useEffect` 로 `GET /api/v1/projects/:id/test-results?window=30d&limit=20` fetch. `apiClient` 정공법
- **표시**:
  - Weighted Pass Rate (큰 카드, Recharts 도넛)
  - Status Distribution (7 status 합산, 가중치 무관 count)
  - Recent Runs (table) — `repository_full_name` 컬럼 추가 (multi-repo)
  - Window selector (7d/30d/90d/1y) + Refresh
  - **가중치 라벨**: pass rate 옆에 "(weighted)" 텍스트
  - **multi-repo hint**: "across N linked repository" hint
- **shadcn/ui + Recharts** 정공법 (기존 `RepositoryTestsSection` 와 동일)

### 1.4 ProjectView 통합

`frontend/app/(dashboard)/projects/[id]/page.tsx` 의 메인 콘텐츠의 left column `ProjectKPISection` 다음 자리에 `ProjectTestsSection` 추가. 기존 `editingWeights` / `projectRepositories` / `completionRate` 와 공존.

### 1.5 신규 ID (3 row)

- `REQ-FR-PROJTEST-001` (Project Tests sub-section, 가중치 정합) — 1
- `ARCH-PROJ-TEST-01` (Project Tests sub-section layout + multi-repo 표시) — 1
- `API-XX` (1 endpoint) — 1 (`/test-results`)
- `IMPL-XX` (frontend component 1 + service 1 + 가중치 SQL) — 3
- `UT-XX` (unit test 2+ e2e 1) — 1 (component)
- `TC-XX` (e2e 1 + Vitest component 5) — 1
- **합계 ~8 ID** (PR #627 의 14 ID 와 정합 — 중복 ID 발급 ❌)

## 2. 변경 범위 (PR 1개, ~700 line)

### 2.1 backend (4 file)

1. `backend-core/internal/httpapi/project_tests.go` (NEW, ~80 line) — `projectTestResults` handler
2. `backend-core/internal/store/project_test_results.go` (NEW, ~150 line) — `ListProjectTestResults` method (LATERAL join + 가중치 종합)
3. `backend-core/internal/domain/application.go` (MODIFY, +35) — `domain.ProjectWeightedTestResults` struct + `domain.ProjectBuildRun` struct
4. `backend-core/internal/httpapi/router.go` (MODIFY, +1) — 1 path 등록
5. `backend-core/internal/httpapi/swaggerui/asset/openapi.yaml` (MODIFY, +90) — 1 path + inline schema
6. `backend-core/internal/httpapi/project_tests_test.go` (NEW, ~60 line) — handler test 2 case
7. `backend-core/internal/domain/rbac-permissions/view/permissions.go` (MODIFY, +4) — routePermissionTable 의 `/projects/:project_id/test-results` row 등록
8. `backend-core/internal/domain/application-lifecycle/view/handler.go` (MODIFY, +3) — `PlatformStore` interface 에 `ListProjectTestResults` 추가
9. `backend-core/internal/httpapi/applications_test.go` (MODIFY, +30) — `memoryPlatformStore` 에 `ListProjectTestResults` stub 추가 (projectKPI 패턴 정합)
10. `backend-core/internal/domain/application-lifecycle/view/fake_store_test.go` (MODIFY, +30) — `fakeViewPlatformStore` 에 `ListProjectTestResults` stub 추가

### 2.2 frontend (4 file)

11. `frontend/domain/platform-lifecycle/schema/project-tests.types.ts` (NEW, ~50 line)
12. `frontend/domain/platform-lifecycle/service/project-tests.service.ts` (NEW, ~50 line)
13. `frontend/domain/platform-lifecycle/view/ProjectTestsSection.tsx` (NEW, ~280 line)
14. `frontend/app/(dashboard)/projects/[id]/page.tsx` (MODIFY, +6) — `ProjectTestsSection` import + left column 2번째 child

### 2.3 test (2 file)

15. `frontend/domain/platform-lifecycle/service/__tests__/project-tests.service.test.ts` (NEW, ~120 line) — 4 case
16. `frontend/domain/platform-lifecycle/view/__tests__/ProjectTestsSection.test.tsx` (NEW, ~150 line) — 5 case
17. `frontend/tests/e2e/project-kpi-tests-section.spec.ts` (NEW, ~120 line) — 1 case

### 2.4 docs (2 file)

18. `docs/planning/2026-06-16-sprint-b-project-tests-sprint-plan.md` (NEW, 본 문서)
19. `docs/planning/kpi-tests-per-domain-scope.md` (MODIFY, +1) — §6.2 변경 이력 row

## 3. 검증

- `go build ./...` PASS
- `go test ./internal/httpapi/...` 2 신규 handler test PASS
- `go test ./...` 회귀 0
- `bash scripts/check-openapi-yaml-lint.sh` PASS
- `cd frontend && npm run test` — 197+ → 197+11 = 208+ unit test PASS
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
| 2026-06-16 | 본 sprint plan 초안 (Sprint B-Tests — Project 가중치 적용 test results + Recharts 도넛) | `chore/260616-sprint-b-project-tests` |
