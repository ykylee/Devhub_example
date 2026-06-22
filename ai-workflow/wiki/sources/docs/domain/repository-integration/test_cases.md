---
title: test_cases
type: source
tags: [domain, test_cases.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/repository-integration/test_cases.md]
git_commit: fb3894f7
git_branch: chore/260622-wiki-drift-cleanup-3
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T06:03:34Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# repository-integration 도메인 — Test Cases 카탈로그 (Sprint A 후속, 2026-06-16)

- 문서 목적: `repository-integration` 도메인의 현행 E2E/UT 검증 범위를 정의하고, Sprint A (PR #597) 의 KPI/Tests sub-section 의 backend 2 endpoint + frontend 2 component 의 회귀 가드를 명시한다.
- 범위:
  - Sprint A 의 신규 endpoint 2종 (`GET /api/v1/repositories/:id/kpi` + `/test-results`) 의 backend handler / store / openapi 정합
  - Sprint A 의 frontend component 2종 (`RepositoryKPISection` + `RepositoryTestsSection`) 의 fetch / display / window selector / error fallback
  - Sprint A 의 routePermissionTable 등록 정합 (deny-by-default 회귀)
  - Sprint A 의 window filter 회귀 (`?window=Nd` short string + `from/to` RFC3339 + build_runs SQL filter)
  - repository-integration 도메인의 기존 E2E (repositories-ui / repositories-detail-negative / repositories-publish / admin-x1 / admin-x2 / rbac-routes) 의 회귀 가드
- 대상 독자: Backend/Frontend 개발자, QA, AI 에이전트, 리뷰어
- 상태: draft (Sprint A 의 e2e spec 별도 PR + Sprint B/C/D/E 후속)
- 최종 수정일: 2026-06-16
- 관련 문서:
  - [`requirements.md`](./requirements.md) (active, Phase 3 split)
  - [`api.md`](./api.md) (active, Phase 3)
  - [`architecture.md`](./architecture.md) (active, Phase 3)
  - [`docs/planning/kpi-tests-per-domain-scope.md`](../../planning/kpi-tests-per-domain-scope.md) §2.1 (Sprint A — Repository sub-section, raw weight=1)
  - [`docs/planning/2026-06-15-x-repository-kpi-tests-section-sprint-plan.md`](../../planning/2026-06-15-x-repository-kpi-tests-section-sprint-plan.md) (sprint plan, PR #597 의 source)
  - [`backend-core/internal/httpapi/repository_ops.go`](../../../backend-core/internal/httpapi/repository_ops.go) `repositoryKPI` + `repositoryTestResults` handler
  - [`backend-core/internal/httpapi/router.go`](../../../backend-core/internal/httpapi/router.go) `routePermissionTable` (2 route 등록)
  - [`backend-core/internal/store/repository_ops.go`](../../../backend-core/internal/store/repository_ops.go) `PostgresStore.CountOpenAndMergedPRs` + `BuildRunListOptions.WindowFrom/To` + `ListRepositoryBuildRuns` SQL filter
  - [`frontend/domain/repository-integration/view/RepositoryKPISection.tsx`](../../../frontend/domain/repository-integration/view/RepositoryKPISection.tsx)
  - [`frontend/domain/repository-integration/view/RepositoryTestsSection.tsx`](../../../frontend/domain/repository-integration/view/RepositoryTestsSection.tsx)
  - [`frontend/domain/repository-integration/service/repository-kpi.service.ts`](../../../frontend/domain/repository-integration/service/repository-kpi.service.ts)
  - [`frontend/domain/repository-integration/service/repository-tests.service.ts`](../../../frontend/domain/repository-integration/service/repository-tests.service.ts)

## 1. 기준

- KPI/Tests 위치 정공법: [`kpi-tests-per-domain-scope.md`](../../planning/kpi-tests-per-domain-scope.md) §1.3 + §2.1 — 단일 repository 의 raw metric (가중치 미적용, weight=1). Sprint B (Project 가중치 rollup) + Sprint C (Platform sub-project rollup) 의 기반.
- 인증/세션: `apiClient<T>` 자동 token refresh + session death 정공법 (기존 `repository.service.ts` 와 동일).
- RBAC: `routePermissionTable` 의 deny-by-default 회귀 (PR #597 P1 #2 fix). 신규 2 route 는 `Resource: domain.ResourcePlatformRepositories, Action: domain.ActionView` 패턴.
- Window 옵션: `7d` / `30d` / `90d` / `1y` / RFC3339 `from`+`to`. default 30d. 기존 `ListRepositoryActivity` 의 `WindowFrom` / `WindowTo` 정합.
- build_runs SQL filter: `COALESCE(started_at, created_at) >= $N AND < $N` 양방향 (count + list 양쪽). zero time → `nil` cast (filter 미적용 정합).

## 2. 유효 E2E 시나리오

| 영역 | spec 파일 | 핵심 TC | sprint |
| --- | --- | --- | --- |
| Sprint A KPI/Tests sub-section UI | `frontend/tests/e2e/repository-kpi-tests-section.spec.ts` (2026-06-16 본 follow-up PR 작성 완료) | `TC-REPO-KPI-TESTS-01` — Repository 상세 진입 → KPI/Tests sub-section 탭 클릭 → metric 정상 표시 + window selector 7d/30d/90d 전환 시 data refetch | `chore/260616-sprint-a-tests-followup` (PR #597 follow-up 2차) |
| 기존 Repository 통합 UI | `frontend/tests/e2e/repositories-ui.spec.ts` | Repository list + detail 진입 + 연동 표시 | 기존 |
| Repository detail 음수 경로 | `frontend/tests/e2e/repositories-detail-negative.spec.ts` | 404 / 권한없음 / 잘못된 id 형식 처리 | 기존 |
| Repository publish flow | `frontend/tests/e2e/repositories-publish.spec.ts` | publish 요청 + 동기화 상태 표시 | 기존 |
| RBAC route 권한 (신규 2 route) | `frontend/tests/e2e/rbac-routes.spec.ts` (확장 필요) | `TC-REPO-RBAC-KPI-01` — 비인가 사용자 401/403, 시스템 admin 허용 | `feat/x-repository-kpi-tests-section` follow-up |
| Admin X-1 / X-2 (시스템 admin) | `frontend/tests/e2e/admin-x1.spec.ts`, `admin-x2.spec.ts` | 시스템 admin 의 repository 운영 | 기존 |

## 3. Backend 유효 단위/통합 테스트 시나리오

| 영역 | test 파일 | 핵심 TC |
| --- | --- | --- |
| handler smoke (memoryPlatformStore) | `backend-core/internal/httpapi/applications_test.go` | `TC-REPO-KPI-HANDLER-01` — `repositoryKPI` 200 OK + RepositoryKPIResponse 스키마 정합 (quality_score / build_success_rate / open_pr_count / merged_pr_count / active_contributor_count / build_run_count / window_from / window_to) |
| handler smoke (memoryPlatformStore) | `backend-core/internal/httpapi/applications_test.go` | `TC-REPO-TEST-RESULTS-HANDLER-01` — `repositoryTestResults` 200 OK + totals 5 status (success/failed/running/cancelled/skipped) + pass_rate (0.0~1.0) + recent 1~20 row + limit 1~50 강제 |
| routePermissionTable 등록 | `backend-core/internal/httpapi/router_test.go` (또는 동등) | `TC-REPO-RBAC-ROUTE-01` — `/repositories/:repository_id/kpi` + `/repositories/:repository_id/test-results` route 가 `routePermissionTable` 에 `ResourcePlatformRepositories/ActionView` 로 매핑 |
| window parsing | `backend-core/internal/httpapi/repository_ops.go` 의 `parseTestResultsWindow` helper | `TC-REPO-WINDOW-PARSE-01` — `?window=7d/30d/90d/365d` 모두 정상 RFC3339 변환, `?from=...&to=...` RFC3339 그대로 사용, 잘못된 값 400 |
| build_runs filter | `backend-core/internal/store/repository_ops.go` `ListRepositoryBuildRuns` | `TC-REPO-BUILDRUN-FILTER-01` — window 외 build_run 제외, zero time → filter 미적용 정합 |
| CountOpenAndMergedPRs | `backend-core/internal/store/repository_ops.go` | `TC-REPO-COUNT-PR-01` — `occurred_at >= $N AND < $N` 정상 카운트, window zero → 0 또는 정합 (P2 #1 fix) |

## 4. Frontend 유효 단위 테스트 시나리오 (Vitest)

| 영역 | test 파일 | 핵심 TC |
| --- | --- | --- |
| `RepositoryKPISection` | `frontend/domain/repository-integration/view/__tests__/RepositoryKPISection.test.tsx` (2026-06-16 본 follow-up PR 작성 완료) | `TC-REPO-KPI-UI-01` — quality_score 색상 코드 (≥80 emerald / ≥60 amber / <60 red / null muted) + build_success_rate ≥0.9 emerald / ≥0.7 amber / <0.7 red + window selector 4 옵션 + refresh |
| `RepositoryTestsSection` | `frontend/domain/repository-integration/view/__tests__/RepositoryTestsSection.test.tsx` (2026-06-16 본 follow-up PR 작성 완료) | `TC-REPO-TESTS-UI-01` — Recharts 도넛 + Status Distribution 5 + Recent Runs table + window selector |
| `fetchRepositoryKPI` | `frontend/domain/repository-integration/service/__tests__/repository-kpi.service.test.ts` (2026-06-16 본 follow-up PR 작성 완료) | `TC-REPO-KPI-SVC-01` — `?window=30d` default, `?from&to` 우선, 404 → `null`, 그 외 error throw |
| `fetchRepositoryTests` | `frontend/domain/repository-integration/service/__tests__/repository-tests.service.test.ts` (2026-06-16 본 follow-up PR 작성 완료) | `TC-REPO-TESTS-SVC-01` — `?limit=20` default, 1~50 강제, 404 → `null` |

## 5. 실행

```sh
cd backend-core
go test ./internal/httpapi/... -run TestRepositoryKPI -v
go test ./internal/httpapi/... -run TestRepositoryTestResults -v
go test ./internal/store/... -run TestRepositoryBuildRuns -v
go test ./internal/store/... -run TestCountOpenAndMergedPRs -v
cd ..
go test ./...
```

```sh
cd frontend
npm run test -- --run repository-kpi.repository-tests
npm run build
cd ..
npx playwright test repository-kpi-tests-section.spec.ts
npx playwright test rbac-routes.spec.ts
```

## 6. 핵심 합격 기준

1. **Endpoint 스키마 정합**: `GET /api/v1/repositories/:id/kpi` 와 `/test-results` 가 `openapi.yaml` 의 `RepositoryKPIResponse` / `RepositoryTestResultsResponse` 스키마와 byte-identical.
2. **RBAC 강제**: `routePermissionTable` 의 2 신규 route 가 deny-by-default 회귀 0. 미매핑 → 403 `auth.policy_unmapped` (PR #597 P1 #2 fix 회귀).
3. **Window filter 정합**: `?window=Nd` 와 `?from&to` 양쪽 parser 정상, 잘못된 값 400, build_runs SQL 양방향 filter 정상.
4. **Frontend display**: quality_score 색상 코드 + build_success_rate 색상 + window selector 4 옵션 + 에러 fallback 정상.
5. **E2E 정상**: `TC-REPO-KPI-TESTS-01` 1 case PASS (sprint plan §2.5 별도 PR 후 smoke 등록).
6. **회귀 0**: 기존 `repositories-ui.spec.ts` + `repositories-detail-negative.spec.ts` + `repositories-publish.spec.ts` + `rbac-routes.spec.ts` 의 기존 TC 모두 PASS 유지.

## 7. Sprint A 의 미완 / 후속 항목

PR #597 본문 명시 + sprint plan §2.5/§3 의 후속:

- [x] **`repository-kpi-tests-section.spec.ts` e2e** — `TC-REPO-KPI-TESTS-01` 1 case 작성 (2026-06-16)
- [x] **`__tests__/RepositoryKPISection.test.tsx`** — Vitest unit 작성 (2026-06-16)
- [x] **`__tests__/RepositoryTestsSection.test.tsx`** — Vitest unit 작성 (2026-06-16)
- [x] **`__tests__/repository-kpi.service.test.ts`** + **`repository-tests.service.test.ts`** — service unit 작성 (2026-06-16)
- [ ] **`rbac-routes.spec.ts` 확장** — 2 신규 route 의 RBAC 강제 TC 추가 (후속)

## 8. 후속 Sprint (참조)

본 TC 카탈로그는 **Sprint A (Repository sub-section)** 한정. 후속 sprint 의 TC 는 별도 문서/추가:

- **Sprint B**: Project sub-section + 가중치 rollup → `docs/domain/platform-lifecycle/test_cases.md` (245 line 의 기존 test_cases.md) 의 §추가
- **Sprint C**: Platform sub-section + sub-project rollup → 동 문서
- **Sprint D**: Sidebar `analyticsMenu` 분리 + 글로벌 `/kpis` `/tests` 도메인 picker → frontend E2E 추가
- **Sprint E**: 글로벌 페이지 옵션 A/B/C 결정 → legacy TC 정리

## 9. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-06-16 | 본 TC 카탈로그 초안 — Sprint A (PR #597) 의 backend 2 endpoint + frontend 2 component + window filter + routePermissionTable 의 회귀 가드 정의. e2e + Vitest unit 의 후속 4 file 명시. | `feat/x-repository-kpi-tests-section` (PR #597 follow-up) |
| 2026-06-16 | 본 follow-up PR 작성 — 후속 4 file (`repository-kpi-tests-section.spec.ts` e2e + `RepositoryKPISection.test.tsx` + `RepositoryTestsSection.test.tsx` + `repository-kpi.service.test.ts` + `repository-tests.service.test.ts`) 완료. `rbac-routes.spec.ts` 확장 잔여. | `chore/260616-sprint-a-tests-followup` (PR #597 follow-up 2차) |
