# N-9 잔여 build-runs polish — TanStack Query hook + dashboard widget (frontend only)

- 문서 목적: PR #555 (`chore(memory,docs): N-9 (P1-7 Repository build-runs) 기본 구현 완료 정합 + 잔여 4건 sub-issue 분리`) 가 명시한 잔여 4건 sub-issue 중 **1, 2번 (backend RBAC 404 + Histogram metric) 은 main 에 이미 구현 완료** 상태. 본 sprint 는 **3, 4번 (frontend) 만** 다룬다.
- 범위: (1) `frontend/domain/repository-integration/hook/useRepositoryBuildRuns.ts` (NEW) — TanStack Query hook + status filter dropdown + skeleton + 무한 스크롤 cursor pagination. (2) `frontend/domain/repository-integration/view/RecentRepositoryActivityWidget.tsx` (NEW) — manager/admin dashboard widget. (3) `frontend/tests/e2e/repository-build-runs.spec.ts` (NEW) — TC-E2E-BUILD-RUNS-01. (4) `frontend/tests/e2e/recent-repository-activity-widget.spec.ts` (NEW) — widget 표시 + RBAC filter.
- 상태: planned → in_progress (branch `chore/260617-n9-residual-build-runs-polish`)
- 작성일: 2026-06-17
- 결정 근거: 사용자 결정 (2026-06-17, "잔여 2건 (frontend only) 로 sprint 범위 축소")
- 관련 문서: [PR #555 본문](https://github.com/ykylee/Devhub_example/pull/555) (잔여 4건 sub-issue 원본), [`docs/planning/release_v0-1_roadmap.md` §3.5 N-9 row](./release_v0-1_roadmap.md) (status ✅ resolved), [`docs/traceability/report.md`](../traceability/report.md) §6 (N-9 정합 row).

## 1. 컨텍스트 (2026-06-17 main 정합)

PR #555 의 잔여 4건 sub-issue 가 분리된 시점 (2026-06-11) 대비, 2026-06-11 ~ 2026-06-17 사이 **1, 2번 (backend) 은 main 에 이미 구현 완료**:

| Sub-issue | 상태 (2026-06-17) | 근거 |
| --- | --- | --- |
| 1. RBAC 403/404 가드 | ✅ **resolved** (2026-06-11 이후) | `backend-core/internal/httpapi/repository_ops.go:144-152` 의 `GetRepositoryByID` 404 가드 + `errRepositoryNotFound` 정의 (line 15) + `repository_ops_test.go:90` 검증. |
| 2. Histogram metric | ✅ **resolved** (2026-06-11 이후) | `backend-core/internal/httpapi/repository_ops.go:177` 의 `observeBuildRunsQueryDuration(opts.Status, queryDuration)` + `metrics.go:29` 의 `BuildRunsQueryDurationSeconds` HistogramVec 정의. |
| 3. TanStack Query hook | 🟡 **in_progress** (본 sprint) | `frontend/domain/repository-integration/hook/useRepositoryBuildRuns.ts` (NEW). |
| 4. Dashboard widget + e2e | 🟡 **in_progress** (본 sprint) | `RecentRepositoryActivityWidget` + `repository-build-runs.spec.ts` (NEW). |

## 2. Sprint scope — frontend 2건

### 2.1 `useRepositoryBuildRuns` TanStack Query hook

**신규**: `frontend/domain/repository-integration/hook/useRepositoryBuildRuns.ts`

- 입력: `repositoryId: string | number` + `options?: { statusFilter?: string; pageSize?: number; enabled?: boolean }`
- TanStack Query: `useInfiniteQuery` (cursor pagination)
- API: `repositoryService.getRepositoryBuildRuns` (이미 존재, PR #462 의 frontend 통합)
- 응답 정규화: `data: { items: BuildRun[]; nextCursor: number | null; total: number }`
- React Query key: `['repositoryBuildRuns', repositoryId, statusFilter]`
- 자동 refetch: `staleTime: 30s` + `refetchOnWindowFocus: false`
- error state: `{ code: 'not_found' | 'unauthorized' | 'network' | 'unknown', message: string }`
- status filter dropdown: 8 enum (queued/running/success/failed/cancelled/skipped/unknown + all)
- skeleton: 첫 page loading 시 5 row skeleton
### 2.1 `useRepositoryBuildRuns` custom hook (useState + useEffect pattern)

**신규**: `frontend/domain/repository-integration/hook/useRepositoryBuildRuns.ts`

> **2026-06-17 결정 (사용자 + Claude)**: 본 sprint 의 frontend hook 은 기존 `useState` + `useEffect` + 직접 fetch 패턴을 따른다 (TanStack Query 미도입). TanStack Query 도입은 별도 architectural sprint (react-query integration) 에서 다룬다. 본 sprint 는 N-9 잔여 4건 sub-issue 의 "TanStack Query hook" 의도를 "custom hook" 으로 약화 적용 — skeleton + status filter + 무한 스크롤 + 에러 정규화는 동일하게 제공하되 query/cache layer 는 custom.

- 입력: `repositoryId: string | number` + `options?: { statusFilter?: string; pageSize?: number; enabled?: boolean }`
SWAP 47.=47:
- **신규 4 file**: hook + widget + 2 e2e spec.
- **수정 가능**: `frontend/domain/repository-integration/service/repository.service.ts` (이미 `getRepositoryBuildRuns` 존재) — 추가 가공 불필요.
- **dashboard 통합 위치**: `frontend/app/(dashboard)/manager/page.tsx` 또는 `frontend/app/(dashboard)/admin/page.tsx` (manager/admin 도메인의 기존 widget slot).

> **2026-06-17 결정**: TanStack Query 도입 안함. 기존 useState + useEffect 패턴 유지. architectural sprint 분리.

### 2.2 `RepositoryBuildRunsSection` sub-section (Repository 상세 view 통합)

> **2026-06-17 결정 (사용자 + Claude)**: widget 의 명칭은 `RecentRepositoryActivityWidget` → `RepositoryBuildRunsSection` 으로 변경. 기존 `RepositoryKPISection` / `RepositoryTestsSection` 와 같은 suffix pattern 정합. 통합 위치는 `ManagerView` (`frontend/domain/repository-integration/view/ManagerView.tsx:329`) 의 sub-section sibling. `/manager` dashboard (archived, line 8 "Quality Status Archived") 가 아닌 단일 repository 상세 view 의 sub-section.

**신규**: `frontend/domain/repository-integration/view/RepositoryBuildRunsSection.tsx`

- `useRepositoryBuildRuns` hook 사용 (status filter dropdown + 무한 스크롤 + skeleton)
- 표시: branch + commit SHA short 7자 + status badge + duration_seconds + started_at relative time
- 정렬: started_at DESC
- click row → 동일 row highlight (deep link 없음 — RepositoryKPISection 과 정합)
- empty state: "No build activity for this repository" + "View all repositories" link

### 2.3 E2E 회귀 가드

**신규 1**: `frontend/tests/e2e/repository-build-runs.spec.ts`
- TC-E2E-BUILD-RUNS-01 — Manager 진입 후 `useRepositoryBuildRuns` hook 의 status filter dropdown 전환 시 data refetch + cursor pagination 동작 검증.
- mock 정공법: `page.route(/\/api\/v1\/repositories\/\d+\/build-runs.*/)` + `?status=success` / `?status=failed` 별 mock 분기.
- 정합 대상: `backend-core/internal/httpapi/repository_ops.go:144-152` RBAC 404 가드 + `metrics.go:29` Histogram metric (backend 검증은 기존 UT 3건 + IT 1건에 위임).

**신규 2**: `frontend/tests/e2e/repository-build-runs-section.spec.ts`

- TC-E2E-BUILD-RUNS-SECTION-01 — Repository 상세 view 진입 시 `RepositoryBuildRunsSection` 표시 + 1 status filter dropdown + 적어도 1 row + skeleton 5 row 초기 로드.
- mock 정공법: `page.route(/\/api\/v1\/repositories\/\d+\/build-runs.*/)` (initial 20 row) + `page.route(/\/api\/v1\/repositories\/\d+\/kpi.*/)` + `page.route(/\/api\/v1\/repositories\/\d+\/test-results.*/)` (RepositoryKPISection / RepositoryTestsSection 통합 정합).
- 정합 대상: `backend-core/internal/httpapi/repository_ops.go:144-152` RBAC 404 가드 + `metrics.go:29` Histogram metric (backend 검증은 기존 UT 3건 + IT 1건에 위임).

**신규 2**: `frontend/tests/e2e/recent-repository-activity-widget.spec.ts`

- TC-E2E-WIDGET-01 — Manager dashboard 진입 시 `RecentRepositoryActivityWidget` 표시 + 10 row 이내 build run + click 시 repository 상세 deep link 검증.
- mock 정공법: `page.route(/\/api\/v1\/repositories(\?.*)?$/)` (전체 repo list) + per-repo `/build-runs` mock.

## 3. 영향 분석

### 3.1 Backend

- **변경 0** (1, 2번은 이미 main 에서 active).
- `metrics.go:29` 의 `observeBuildRunsQueryDuration` 의 metric name = `devhub_repository_build_runs_query_duration_seconds` + `status_filter` label 은 추가 변경 없이 그대로 사용.

### 3.2 Frontend

- **신규 4 file**: hook + widget + 2 e2e spec.
- **수정 가능**: `frontend/domain/repository-integration/service/repository.service.ts` (이미 `getRepositoryBuildRuns` 존재) — 추가 가공 불필요.
- **dashboard 통합 위치**: `frontend/app/(dashboard)/manager/page.tsx` 또는 `frontend/app/(dashboard)/admin/page.tsx` (manager/admin 도메인의 기존 widget slot).

### 3.3 Docs

- 본 sprint plan (NEW).
- `docs/planning/release_v0-1_roadmap.md` §3.5 N-9 row 의 "잔여 4건 sub-issue" status 갱신 (3, 4 → done, 1, 2 → done).
- `docs/traceability/report.md` §6 본 sprint 종합 row 추가.

## 4. DoD

1. ✅ `useRepositoryBuildRuns` hook 의 status filter dropdown 전환 + cursor pagination 동작
2. ✅ `RecentRepositoryActivityWidget` 의 10 row 이내 표시 + deep link
3. ✅ E2E 2 spec PASS (CI)
4. ✅ Vitest hook unit test 1 case PASS (react-query Testing Library)
5. ✅ `release_v0-1_roadmap.md` §3.5 N-9 row "잔여 4건 sub-issue" 1, 2, 3, 4 모두 done 갱신
6. ✅ `docs/traceability/report.md` §6 본 sprint 종합 row 추가
7. ✅ Tier self-check: 사외 (frontend only + docs)
8. ✅ 사내 한정 정보 미포함 self-check 4/4

## 5. Tier

- **사외** (GitHub main)
- frontend only + docs
- backend 변경 0 (1, 2번 main 정합 활용)

## 6. 변경 이력

| 일자 | 변경 | sprint |
|---|---|---|
| 2026-06-17 | 본 sprint plan 작성. 사용자 결정 (2026-06-17, "잔여 2건 frontend only 로 축소") 반영. branch `chore/260617-n9-residual-build-runs-polish`. | (본 sprint) |
| 2026-06-17 | PR #633 의 codex P2 review 코멘트 2건 정공법 fix (동일 branch follow-up commit). (1) backend `store.ListRepositoryBuildRuns` 의 3-tuple total 을 frontend 가 합성하지 않고 `service.getRepositoryBuildRunsWithMeta` (legacy `getRepositoryBuildRuns` caller 무손상 보존) 로 expose, hook 이 `meta.total` 사용 (fallback = `items.length >= pageSize`) — `hasMore` 정확. (2) hook 의 cleanup boolean + ref-reset race → `AbortController` per-request token + `isCurrent` 가드 (ref 일치 + `signal.aborted`) + service 옵션 `signal` → `apiClient` plumb. `apiClient` 4번째 인자 `options.signal?: AbortSignal` 추가 (backward-compatible) + refresh 후 2nd fetch 동일 signal 유지. 검증: hook 9/9 + section 7/7 + service 13/13 + apiClient 17/17 + RepositoryDashboardView 8/8 PASS, `tsc --noEmit` PR 영향 7 file 모두 0. | (본 sprint) |
