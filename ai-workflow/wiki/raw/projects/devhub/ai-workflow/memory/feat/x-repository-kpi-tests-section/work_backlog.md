# Work Backlog — feat/x-repository-kpi-tests-section (Sprint A)

- 문서 목적: KPI/Tests 위치 정공법 컨셉 ([kpi-tests-per-domain-scope.md](../../../docs/planning/kpi-tests-per-domain-scope.md)) 의 Sprint A — Repository 단위 sub-section.
- 범위: backend 2 endpoint + openapi 2 path + frontend 2 component + 2 service + 2 schema + ManagerView 통합. Tier: **공용**.
- 상태: 작업 완료, commit + push + PR 발행 pending
- 최종 수정일: 2026-06-15

## 0. 현재 status

**main HEAD**: `7f0b5ae2` (PR #595 X-4 100% 완료 시점)
**branch**: `feat/x-repository-kpi-tests-section` (worktree `.worktrees/x-repository-kpi-tests-section/`)
**sprint scope**: Sprint A — Repository KPI/Tests sub-section (kpi-tests-per-domain-scope.md §6.1)

## 1. 본 sprint 진행 상황

| Step | 작업 | 결과 |
|---|---|---|
| 1 | 컨셉 문서 (`docs/planning/kpi-tests-per-domain-scope.md`) | ✅ (main) |
| 2 | worktree + branch 생성 | ✅ |
| 3 | sprint plan 작성 (`docs/planning/2026-06-15-x-repository-kpi-tests-section-sprint-plan.md`) | ✅ |
| 4 | backend: `PlatformStore.CountOpenAndMergedPRs` interface method 추가 | ✅ |
| 5 | backend: `PostgresStore.CountOpenAndMergedPRs` SQL + Go | ✅ |
| 6 | backend: `memoryPlatformStore.CountOpenAndMergedPRs` mock | ✅ |
| 7 | backend: `repositoryKPI` handler | ✅ |
| 8 | backend: `repositoryTestResults` handler | ✅ |
| 9 | backend: router 2 path 등록 | ✅ |
| 10 | openapi.yaml 2 path 추가 (paths 84 → 86) | ✅ |
| 11 | openapi lint PASS | ✅ |
| 12 | frontend: `repository-kpi.types.ts` + `repository-tests.types.ts` | ✅ |
| 13 | frontend: `repository-kpi.service.ts` + `repository-tests.service.ts` | ✅ |
| 14 | frontend: `RepositoryKPISection.tsx` (4 card + window selector) | ✅ |
| 15 | frontend: `RepositoryTestsSection.tsx` (도넛 차트 + status 분포 + recent table) | ✅ |
| 16 | frontend: `ManagerView.tsx` import + 배치 | ✅ |
| 17 | `go build ./...` PASS | ✅ |
| 18 | `bash scripts/check-openapi-yaml-lint.sh` PASS | ✅ |
| 19 | frontend `npm install` + build/test — **CI 검증 위임** (본 세션 node_modules 부재) | ⏳ CI |
| 20 | kpi-tests-per-domain-scope.md §6.1 status 갱신 | ⏳ |
| 21 | traceability report.md §6 row | ⏳ |
| 22 | mirror-list.md §1.7.1 frontend file 추가 | ⏳ |
| 23 | CHANGELOG.md status | ⏳ |
| 24 | memory 4 file (state.json + session_handoff + work_backlog + backlog) | ✅ (방금 작성) |
| 25 | commit + push + PR 발행 | ⏳ pending |

## 2. 잔여 follow-up (Sprint B/C/D/E)

| ID | 아이템 | 영역 | effort |
|---|---|---|---|
| Sprint B | Project 단위 KPI/Tests sub-section (contribution_weight 가중치) | BE/사외 | 1~2 ses |
| Sprint C | Platform 단위 KPI/Tests sub-section (sub-project rollup) | BE/사외 | 1~2 ses |
| Sprint D | Sidebar 의 `analyticsMenu` 분리 + 글로벌 `/kpis` `/tests` picker | FE/사외 | 0.5 ses |
| Sprint E | 글로벌 페이지 옵션 A/B/C 결정 + legacy 정리 | FE/docs | 0.3 ses |
| e2e spec | `repository-kpi-tests-section.spec.ts` + smoke manifest 등록 | FE | 0.3 ses |
| 별도 repository_tests table | build_runs 분포 한계, Sprint F 또는 후속 | BE | 별도 |

## 3. 다음 sprint (가장 자연스러운 것: Sprint B)

Project 단위 sub-section. `work_260607-a-dashboard-improvements` 가 이미 `contribution_weight` 가중치 + `completionRate` 구현 — 본 sprint 는 그 정공법 활용. backend 2 endpoint (`GET /api/v1/projects/:id/kpi/quality` + `/kpi/test-pass-rate`), frontend 2 component (`ProjectKPISection` + `ProjectTestsSection`).

## 4. 변경 이력

| 일자 | 변경 | 비고 |
|---|---|---|
| 2026-06-15 | 컨셉 doc (`kpi-tests-per-domain-scope.md`) main 신규 | 본 sprint 의 기반 |
| 2026-06-15 | Sprint A branch 신규 (~800 line, Tier: 공용) | 본 sprint |
