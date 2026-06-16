# Sprint C — Platform 상세 KPI / Tests sub-section (sub-project rollup)

- 문서 목적: [`kpi-tests-per-domain-scope.md`](./kpi-tests-per-domain-scope.md) 의 Sprint C — Platform sub-section 의 1차 진입. Sprint B (Project 가중치, N:M repo) 의 후속으로, **sub-project rollup (N:M project)** 의 정공법 확립.
- 범위: backend 2 endpoint (`GET /api/v1/platforms/:id/kpi` + `/test-results`, sub-project 가중치) + frontend 2 component (`PlatformKPISection` + `PlatformTestsSection`) + `ManagerView` 또는 `PlatformView` 통합 + `routePermissionTable` 등록 + 회귀 가드 + e2e 1 case.
- 상태: planned → in_progress (branch `chore/260616-sprint-c-platform-kpi-tests`)
- 작성일: 2026-06-16
- 결정 근거: [`kpi-tests-per-domain-scope.md` §2.3](./kpi-tests-per-domain-scope.md#23-platform-상세-sub-project-rollup-nm-project) + §6.3
- 관련 문서:
  - [`project_concept.md`](../domain/platform-lifecycle/project_concept.md) §2 운영 계층 모델
  - [`kpi-tests-per-domain-scope.md` §4.3](./kpi-tests-per-domain-scope.md#43-platform--backend-api-sub-project-rollup) 데이터 흐름
  - [`docs/domain/platform-lifecycle/requirements.md`](../domain/platform-lifecycle/requirements.md) REQ-FR-APP-* + REQ-FR-APPDASH-*
  - [`docs/traceability/report.md`](../traceability/report.md) §3 application-lifecycle row
  - PR #597 (Sprint A 정공법), PR #625 (Sprint A follow-up), PR #626 (Sprint D), PR #627 (Sprint B 1차), PR #628 (Sprint B-Tests), PR #629 (Sprint B-Projects Picker)

## 0. 컨텍스트

### 0.1 Sprint B 정공법 차용 + 확장

- backend: `ComputeProjectWeightedKPI` + `CountProjectOpenAndMergedPRs` + `ListProjectTestResults` (CTE 3개, 단일 round-trip)
- frontend: `ProjectKPISection` + `ProjectTestsSection` (Recharts 도넛 + recent + window + (weighted) 라벨 + multi-repo)
- 정합법: N개 linked **repository** 의 raw metric × contribution_weight 가중 평균

### 0.2 Sprint C 의 차이점

- **sub-project rollup**: N개 linked **project** 의 가중치 적용 metric 을 통합 (Sprint B 는 repo)
- **2-depth 가중치**: Sprint B 가 이미 contribution_weight × raw metric 인데, Sprint C 는 그 결과를 sub-project 균등 평균 (`AVG(per_project_metric)`)
- **기존 `ComputePlatformRollup` 와 별개**: 기존 `ComputePlatformRollup` 는 repository 단위 weight_policy (equal / repo_role / custom) — Sprint C 는 sub-project 단위 equal average (정합법 분리)
- **multi-project recent**: Sprint B-Tests 의 multi-repo recent 와 정합 + project_name 컬럼 추가

### 0.3 기존 정공법 활용

- `ListPlatformProjects` (기존) — platform 의 linked projects
- `project_repositories` table — project 의 linked repositories (Sprint B 의 1차 사용처)
- `quality_snapshots` + `build_runs` + `pr_activities` (Sprint A/B 의 1차 사용처)
- `WindowFrom/WindowTo` (Sprint A 의 BuildRunListOptions 정공법)

## 1. 결정

### 1.1 Backend 2 endpoint

#### 1.1.1 `GET /api/v1/platforms/:platform_id/kpi?window=30d`

- **목적**: 단일 platform 의 N개 sub-project 의 가중치 적용 KPI 종합 (sub-project equal avg)
- **Response** (`200 OK`):
  ```json
  {
    "status": "ok",
    "data": {
      "platform_id": "pl-001",
      "window_from": "2026-05-16T00:00:00Z",
      "window_to": "2026-06-15T00:00:00Z",
      "weighted_quality_score": 87.3,
      "weighted_build_success_rate": 0.94,
      "total_build_run_count": 156,
      "open_pr_count": 7,
      "merged_pr_count": 23,
      "active_contributor_count": 12,
      "linked_project_count": 3,
      "weighted_at": "2026-06-15T00:00:00Z"
    }
  }
  ```
- **권한**: `routePermissionTable` — `Resource: domain.ResourcePlatforms, Action: domain.ActionView` (기존 `/platforms/:id/rollup` 와 동일)
- **가중치 정공법**: sub-project 의 `ProjectWeightedKPI` 를 1차 계산 후 equal avg:
  - per_project_quality = latest_quality_score (project 의 linked repo 중 가장 최근 1건)
  - per_project_build_success = (success_i / total_i) (project 의 모든 linked repo)
  - per_project_build_count = Σ (모든 linked repo 의 build_run_count)
  - per_project_active_contributors = Σ (모든 linked repo 의 distinct commit_author)
  - **platform weighted = AVG(per_project_X)** (sub-project 균등)
  - linked_project_count = 0 → weighted metric 0
- **window 파라미터**: `7d` / `30d` / `90d` / `1y` / RFC3339 from/to. default 30d
- **에러**:
  - `400 platform_id required`
  - `400 window must be 7d|30d|90d|1y or from/to RFC3339`
  - `403 auth.policy_unmapped`
  - `500 platform.kpi.weighted`

#### 1.1.2 `GET /api/v1/platforms/:platform_id/test-results?window=30d&limit=20`

- **목적**: 단일 platform 의 sub-project 가중치 적용 test pass rate + N개 sub-project 의 build_runs status 합산 + recent runs
- **Response** (`200 OK`):
  ```json
  {
    "status": "ok",
    "data": {
      "platform_id": "pl-001",
      "window_from": "...",
      "window_to": "...",
      "weighted_pass_rate": 0.93,
      "totals": {
        "success": 145, "failed": 8, "running": 1, "cancelled": 2,
        "skipped": 0, "queued": 0, "unknown": 0
      },
      "recent": [
        {
          "id": 100,
          "project_id": "p-001",
          "project_full_name": "API Modernization",
          "repository_id": 1,
          "repository_full_name": "org/repo-a",
          "run_external_id": "ext-100",
          "commit_sha": "feedface",
          "status": "success",
          "branch": "main",
          "started_at": "...",
          "finished_at": "..."
        }
      ]
    },
    "meta": { "total": 156, "limit": 20 }
  }
  ```
- **권한**: 동일
- **weighted_pass_rate 정공법**:
  - per_project_pass_rate = (success / (success+failed)) per project
  - **platform weighted = AVG(per_project_pass_rate)** (sub-project 균등)
  - denom=0 인 sub-project 모두 0 → weighted 0 또는 null
- **totals**: N개 sub-project 의 build_runs status 합산 (가중치 무관, 단순 count, 7 status 모두 0 초기화)
- **recent**: 모든 sub-project 의 build_runs 최신순 limit (multi-project, `project_full_name` + `repository_full_name` 표시)

### 1.2 Backend SQL 정공법

- `ComputePlatformWeightedKPI` (NEW) — sub-project equal average, 단일 query:
  ```sql
  WITH linked_projects AS (
    SELECT p.id, p.contribution_weight
    FROM projects p WHERE p.platform_id = $1::uuid
  ),
  per_project_metric AS (
    SELECT
      lp.id, lp.contribution_weight,
      COALESCE(latest_quality.score, 0) AS quality_score,
      COALESCE(agg.build_success_rate, 0) AS build_success_rate,
      COALESCE(agg.build_run_count, 0) AS build_run_count,
      COALESCE(agg.active_contributor_count, 0) AS active_contributor_count,
      COALESCE(pr_stats.open_count, 0) AS open_count,
      COALESCE(pr_stats.merged_count, 0) AS merged_count
    FROM linked_projects lp
    LEFT JOIN LATERAL (
      SELECT score FROM quality_snapshots qs
      WHERE qs.repository_id IN (
        SELECT pr.repository_id FROM project_repositories pr WHERE pr.project_id = lp.id
      )
      ORDER BY measured_at DESC LIMIT 1
    ) latest_quality ON true
    LEFT JOIN LATERAL (
      SELECT
        COALESCE(COUNT(*) FILTER (WHERE br.status = 'success')::float / NULLIF(COUNT(*), 0), 0) AS build_success_rate,
        COUNT(*)::int AS build_run_count,
        COUNT(DISTINCT br.commit_author)::int AS active_contributor_count
      FROM ci_runs br
      WHERE br.repository_id IN (
        SELECT pr.repository_id FROM project_repositories pr WHERE pr.project_id = lp.id
      )
        AND ($2::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) >= $2)
        AND ($3::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) < $3)
    ) agg ON true
    LEFT JOIN LATERAL (
      SELECT
        COUNT(DISTINCT pa.number) FILTER (WHERE pa.event_type = 'opened')::int AS open_count,
        COUNT(DISTINCT pa.number) FILTER (WHERE pa.event_type = 'merged')::int AS merged_count
      FROM pr_activities pa
      WHERE pa.repository_id IN (
        SELECT pr.repository_id FROM project_repositories pr WHERE pr.project_id = lp.id
      )
        AND pa.occurred_at >= $2 AND pa.occurred_at < $3
    ) pr_stats ON true
  )
  SELECT
    COUNT(*)::int AS linked_project_count,
    COALESCE(AVG(quality_score), 0)::float8 AS weighted_quality_score,
    COALESCE(AVG(build_success_rate), 0)::float8 AS weighted_build_success_rate,
    COALESCE(SUM(build_run_count), 0)::int AS total_build_run_count,
    COALESCE(SUM(active_contributor_count), 0)::int AS total_active_contributors,
    COALESCE(SUM(open_count), 0)::int AS total_open_pr,
    COALESCE(SUM(merged_count), 0)::int AS total_merged_pr
  FROM per_project_metric
  ```

- `ListPlatformTestResults` (NEW) — sub-project equal avg weighted_pass_rate + multi-project recent, 단일 query:
  ```sql
  WITH linked_projects AS (
    SELECT p.id FROM projects p WHERE p.platform_id = $1::uuid
  ),
  -- 1) per-project pass_rate
  per_project AS (
    SELECT
      lp.id,
      COALESCE(stats.success_count::float / NULLIF(stats.success_count + stats.failed_count, 0), 0) AS pass_rate
    FROM linked_projects lp
    LEFT JOIN LATERAL (
      SELECT
        COUNT(*) FILTER (WHERE br.status = 'success')::int AS success_count,
        COUNT(*) FILTER (WHERE br.status = 'failed')::int AS failed_count
      FROM ci_runs br
      WHERE br.repository_id IN (
        SELECT pr.repository_id FROM project_repositories pr WHERE pr.project_id = lp.id
      )
        AND ($4::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) >= $4)
        AND ($5::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) < $5)
    ) stats ON true
  ),
  -- 2) totals
  totals AS (
    SELECT br.status, COUNT(*)::int AS cnt
    FROM ci_runs br
    WHERE br.repository_id IN (
      SELECT pr.repository_id FROM project_repositories pr
      WHERE pr.project_id IN (SELECT id FROM linked_projects)
    )
      AND ($4::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) >= $4)
      AND ($5::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) < $5)
    GROUP BY br.status
  ),
  -- 3) recent (multi-project)
  recent AS (
    SELECT
      br.id, br.repository_id, r.full_name AS repository_full_name,
      br.external_id AS run_external_id, br.branch,
      COALESCE(br.commit_sha, '') AS commit_sha,
      br.status, br.duration_seconds,
      COALESCE(br.started_at, br.created_at) AS started_at, br.finished_at,
      p.id AS project_id, p.name AS project_full_name
    FROM ci_runs br
    JOIN repositories r ON r.id = br.repository_id
    JOIN project_repositories pr ON pr.repository_id = br.repository_id
    JOIN projects p ON p.id = pr.project_id
    WHERE p.id IN (SELECT id FROM linked_projects)
      AND ($4::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) >= $4)
      AND ($5::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) < $5)
    ORDER BY COALESCE(br.started_at, br.created_at) DESC
    LIMIT $3
  ),
  total_count AS (
    SELECT COUNT(*)::int AS cnt
    FROM ci_runs br
    WHERE br.repository_id IN (
      SELECT pr.repository_id FROM project_repositories pr
      WHERE pr.project_id IN (SELECT id FROM linked_projects)
    )
      AND ($4::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) >= $4)
      AND ($5::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) < $5)
  )
  SELECT
    (SELECT COALESCE(AVG(pass_rate), 0)::float8 FROM per_project) AS weighted_pass_rate,
    (SELECT json_object_agg(status, cnt) FROM totals) AS totals_json,
    (SELECT json_agg(json_build_object(
      'id', id, 'project_id', project_id, 'project_full_name', project_full_name,
      'repository_id', repository_id, 'repository_full_name', repository_full_name,
      'run_external_id', run_external_id, 'branch', branch, 'commit_sha', commit_sha,
      'status', status, 'duration_seconds', duration_seconds,
      'started_at', started_at, 'finished_at', finished_at
    )) FROM recent) AS recent_json,
    (SELECT cnt FROM total_count) AS total_count
  ```

  - linked_projects 0 → AVG(pass_rate) = NULL (no rows) → weighted_pass_rate = 0 fallback or null. 결정: null (linked project 0 시 명시적 null).
  - totals 정규화: 7 status 모두 0 초기화 + store 의 raw status merge (Sprint B-Tests 정공법 정합).

### 1.3 Frontend 2 component

#### 1.3.1 `PlatformKPISection`

- **위치**: `frontend/domain/platform-lifecycle/view/PlatformKPISection.tsx` (NEW)
- **props**: `platformId: string`
- **내부**: `useEffect` 로 `GET /api/v1/platforms/:id/kpi?window=30d` fetch. `apiClient` 정공법
- **표시**:
  - Weighted Quality Score (큰 카드)
  - Weighted Build Success Rate
  - Total Build Run Count
  - Open PR Count + Merged PR Count (정수)
  - Active Contributors
  - Linked Project Count
  - Window selector (7d/30d/90d/1y) + Refresh + "(weighted)" 라벨
- **shadcn/ui** 정공법 (기존 `ProjectKPISection` 와 동일)

#### 1.3.2 `PlatformTestsSection`

- **위치**: `frontend/domain/platform-lifecycle/view/PlatformTestsSection.tsx` (NEW)
- **props**: `platformId: string`
- **내부**: `useEffect` 로 `GET /api/v1/platforms/:id/test-results?window=30d&limit=20` fetch
- **표시**:
  - Weighted Pass Rate (도넛)
  - Status Distribution (7 status)
  - Recent Runs (table) — `project_full_name` + `repository_full_name` 컬럼 (multi-project)
  - Window selector + limit
- **Recharts 도넛** 정공법 (기존 `ProjectTestsSection` 와 동일)

### 1.4 PlatformView 통합

`frontend/app/(dashboard)/platforms/[id]/page.tsx` 의 메인 콘텐츠에 KPI/Tests sub-section 추가. 기존 `metrics_overview` + `quality_metrics` + `history_trend` 와 공존.

### 1.5 신규 ID

- `REQ-FR-PLATKPI-001` (Platform KPI/Tests sub-section) — 1
- `ARCH-PLAT-01` (Platform sub-section layout) — 1
- `API-2XX` (2 endpoint) — 2
- `IMPL-2XX` (frontend component 2 + service 2 + 가중치 SQL) — 4
- `UT-2XX` (unit test) — 1
- `TC-2XX` (e2e + Vitest) — 1
- **합계 ~10 ID** (기존 가중치 정공법 확장 — 신규 ID 0 가능)

## 2. 변경 범위 (PR 1개, ~1500 line)

### 2.1 backend (8 file)

1. `backend-core/internal/httpapi/platform_kpi.go` (NEW, ~80 line) — `platformKPI` handler
2. `backend-core/internal/httpapi/platform_kpi_test.go` (NEW, ~60 line) — handler test 2 case
3. `backend-core/internal/httpapi/platform_tests.go` (NEW, ~80 line) — `platformTestResults` handler
4. `backend-core/internal/httpapi/platform_tests_test.go` (NEW, ~60 line) — handler test 2 case
5. `backend-core/internal/store/platform_kpi_tests.go` (NEW, ~200 line) — `ComputePlatformWeightedKPI` + `ListPlatformTestResults` method
6. `backend-core/internal/domain/application.go` (MODIFY, +50) — `domain.PlatformWeightedKPI` + `domain.PlatformWeightedTestResults` + `domain.PlatformBuildRun` struct
7. `backend-core/internal/httpapi/router.go` (MODIFY, +4) — 2 path 등록
8. `backend-core/internal/httpapi/swaggerui/asset/openapi.yaml` (MODIFY, +200) — 2 path + inline schema
9. `backend-core/internal/domain/rbac-permissions/view/permissions.go` (MODIFY, +6) — routePermissionTable 의 2 row 등록
10. `backend-core/internal/domain/application-lifecycle/view/handler.go` (MODIFY, +4) — `PlatformStore` interface 에 2 method 추가
11. `backend-core/internal/httpapi/applications_test.go` (MODIFY, +60) — `memoryPlatformStore` 에 2 stub 추가
12. `backend-core/internal/domain/application-lifecycle/view/fake_store_test.go` (MODIFY, +60) — `fakeViewPlatformStore` 에 2 stub 추가

### 2.2 frontend (6 file)

13. `frontend/domain/platform-lifecycle/schema/platform-kpi.types.ts` (NEW, ~40 line)
14. `frontend/domain/platform-lifecycle/schema/platform-tests.types.ts` (NEW, ~50 line)
15. `frontend/domain/platform-lifecycle/service/platform-kpi.service.ts` (NEW, ~50 line)
16. `frontend/domain/platform-lifecycle/service/platform-tests.service.ts` (NEW, ~50 line)
17. `frontend/domain/platform-lifecycle/view/PlatformKPISection.tsx` (NEW, ~220 line)
18. `frontend/domain/platform-lifecycle/view/PlatformTestsSection.tsx` (NEW, ~290 line)
19. `frontend/app/(dashboard)/platforms/[id]/page.tsx` (MODIFY, +8) — `PlatformKPISection` + `PlatformTestsSection` import + sub-section 배치
20. `frontend/shared/ui-foundation/components/DomainPicker.tsx` (MODIFY, +2) — Platform scope `ready: true` + projects fetch (Sprint C 와 함께)
21. `frontend/app/(dashboard)/kpis/page.tsx` (MODIFY, +10) + `tests/page.tsx` (MODIFY, +10) — platforms fetch (Sprint C 와 함께)

### 2.3 test (4 file)

22. `frontend/domain/platform-lifecycle/service/__tests__/platform-kpi.service.test.ts` (NEW, ~120 line) — 4 case
23. `frontend/domain/platform-lifecycle/service/__tests__/platform-tests.service.test.ts` (NEW, ~120 line) — 4 case
24. `frontend/domain/platform-lifecycle/view/__tests__/PlatformKPISection.test.tsx` (NEW, ~150 line) — 5 case
25. `frontend/domain/platform-lifecycle/view/__tests__/PlatformTestsSection.test.tsx` (NEW, ~150 line) — 5 case
26. `frontend/tests/e2e/platform-kpi-tests-section.spec.ts` (NEW, ~120 line) — 1 case

### 2.4 docs (2 file)

27. `docs/planning/2026-06-16-sprint-c-platform-kpi-tests-sprint-plan.md` (NEW, 본 문서)
28. `docs/planning/kpi-tests-per-domain-scope.md` (MODIFY, +1) — §6.3 status + 변경 이력 row

## 3. 검증

- `go build ./...` PASS
- `go test ./internal/httpapi/...` 4 신규 handler test PASS
- `go test ./...` 회귀 0
- `bash scripts/check-openapi-yaml-lint.sh` PASS
- `cd frontend && npm run test` — 1100+ unit test PASS (회귀 0)
- `cd frontend && npm run build` — silent PASS
- `npx playwright test platform-kpi-tests-section.spec.ts` PASS
- e2e shard 1/2/3 path-detect: frontend 변경 → trigger
- CI 4/4 PASS 예상

## 4. Tier

- **공용** (코드 + openapi + docs 모두 사내 한정 정보 미포함)

## 5. 잔여 (Sprint D 보완 / E)

- Sprint D 보완 완료: DomainPicker 의 platforms fetch 활성화 (본 PR 에서)
- Sprint E: 글로벌 페이지 옵션 A/B/C 결정 + legacy 본문 처리

## 6. 다음 세션 directive

- PR 머지 후 main flat memory 3 file finalize
- 위키 mirror 갱신 (`bash scripts/wiki-sync-devhub.sh`)
- Sprint E 진입 시 옵션 A/B/C 결정

## 7. 변경 이력

| 일자 | 변경 | sprint |
|---|---|---|
| 2026-06-16 | 본 sprint plan 초안 (Sprint C — Platform sub-project rollup, sub-project equal avg) | `chore/260616-sprint-c-platform-kpi-tests` |
