## 카테고리 · 모듈

- 카테고리: `repository-integration` + `kpi/tests 위치 정공법` (컨셉 `kpi-tests-per-domain-scope.md` 의 Sprint A)
- 모듈: `internal/store/repository_ops` (PostgresStore.CountOpenAndMergedPRs) + `internal/domain/application-lifecycle/view/handler` (PlatformStore interface) + `internal/httpapi/repository_ops` (2 handler) + `internal/httpapi/router` (2 path) + `internal/httpapi/swaggerui/asset/openapi` (2 path) + frontend `domain/repository-integration` (2 schema + 2 service + 2 component + ManagerView 통합)

## Tier

- [x] 사외 (GitHub main)
- [ ] 사내 (사내 SCM)
- [ ] 공용 (양쪽 동기화)

## 사내 한정 정보 self-check (사외 PR 인 경우)

- [x] 사내 env var 미포함 (DEVHUB_* env var 미사용, 표준 backend API)
- [x] 사내 호스트명/IP 미포함
- [x] 사내 한정 경로 변경 없음
- [x] 사내 한정 env 파일 변경 없음

## 요약

**KPI/Tests 위치 정공법 컨셉** ([`kpi-tests-per-domain-scope.md`](docs/planning/kpi-tests-per-domain-scope.md)) 의 **Sprint A** — 신규 추가된 글로벌 `/kpis` `/tests` 대시보드의 위치를 **각 도메인 페이지 (Platform/Project/Repository) 의 sub-section** 으로 재배치하는 컨셉의 **1차 진입 (Repository)**. 단일 PR, ~800 line, Tier: **공용**.

**3 도메인 scope**:
- **Repository** (`/repositories/[id]`) — raw metric, weight=1 (본 PR)
- **Project** (`/projects/[id]`) — `project_repositories.contribution_weight` 가중치 적용 rollup (Sprint B, 후속)
- **Platform** (`/platforms/[id]`) — sub-project rollup, 균등 or custom 가중치 (Sprint C, 후속)

본 PR 은 **가장 단순 (가중치 없음)** 의 Repository 단위 1차 진입. Sprint B/C/D/E 는 후속 PR.

## 변경 상세

### 컨셉 문서 (신규, 11.3KB)

1. `docs/planning/kpi-tests-per-domain-scope.md` (NEW) — KPI/Tests 위치 정공법 컨셉. §1 배경 + §2 3 도메인 scope + §3 sidebar 재구성 + §4 데이터 흐름 + §5 권한/탐색 + §6 5 sprint 진입 hook + §7 DoD + §8 Tier.

### backend (6 file, +427 line)

2. `backend-core/internal/domain/application-lifecycle/view/handler.go` (MODIFY, +5 line) — `PlatformStore` interface 의 `Repository 운영 지표` 그룹에 `CountOpenAndMergedPRs(context.Context, int64, time.Time, time.Time) (int, int, error)` method 1개 추가.
3. `backend-core/internal/store/repository_ops.go` (MODIFY, +19 line) — `PostgresStore.CountOpenAndMergedPRs` SQL: `pr_activities` 의 (event_type='opened'|'merged') 별 distinct number count. migration X-5 follow-up 의 `stateToEventType` 정공법 정합 (event_type enum).
4. `backend-core/internal/httpapi/repository_ops.go` (MODIFY, +234 line) — `repositoryKPI` + `repositoryTestResults` handler 2개 + `parseTestResultsWindow` + `parseWindowShort` helper (Nd/Nw/Nm/Ny parse).
5. `backend-core/internal/httpapi/router.go` (MODIFY, +3 line) — 2 path 등록.
6. `backend-core/internal/httpapi/applications_test.go` (MODIFY, +5 line) — `memoryPlatformStore.CountOpenAndMergedPRs` mock (1차 메모리 store 정합).
7. `backend-core/internal/httpapi/swaggerui/asset/openapi.yaml` (MODIFY, +161 line) — 2 path (paths 84 → 86) + schemas 0 (Envelope.data inline).

### frontend (7 file, +500 line)

8. `frontend/domain/repository-integration/schema/repository-kpi.types.ts` (NEW, ~25 line) — `RepositoryKPI` type (quality_score + build_success_rate + open_pr_count + merged_pr_count + active_contributor_count).
9. `frontend/domain/repository-integration/schema/repository-tests.types.ts` (NEW, ~40 line) — `RepositoryTestResults` type (totals 7 status + pass_rate + recent).
10. `frontend/domain/repository-integration/service/repository-kpi.service.ts` (NEW, ~45 line) — `fetchRepositoryKPI` (windowDays option). apiClient 정공법.
11. `frontend/domain/repository-integration/service/repository-tests.service.ts` (NEW, ~40 line) — `fetchRepositoryTestResults` (window + limit option).
12. `frontend/domain/repository-integration/view/RepositoryKPISection.tsx` (NEW, ~190 line) — 4 card (Quality Score / Build Success Rate / Pull Requests / Active Contributors) + window selector + refresh. Color code: emerald (>=80%) / amber (>=60%) / red (<60%).
13. `frontend/domain/repository-integration/view/RepositoryTestsSection.tsx` (NEW, ~250 line) — Pass Rate 큰 카드 + Recharts 도넛 차트 + Status Distribution 7 status + Recent Runs table + window selector.
14. `frontend/domain/repository-integration/view/ManagerView.tsx` (MODIFY, +9 line) — 2 component import + SCM Connection Log 뒤에 inline 배치.

### docs (3 file)

15. `docs/planning/2026-06-15-x-repository-kpi-tests-section-sprint-plan.md` (NEW, ~8.4KB) — 본 sprint 의 sprint plan (Sprint A scope + 결정 + 변경 범위 + 검증 + 잔여 follow-up).
16. `docs/planning/kpi-tests-per-domain-scope.md` (MODIFY, +1 line) — §9 변경 이력 row 추가 (Sprint A in_progress).
17. `docs/traceability/report.md` (MODIFY, +1 row) — §6 Sprint A row.
18. `docs/llm-wiki/mirror-list.md` (MODIFY) — §1.7.1 의 6 frontend file 추가 + count 19 → 25 + §1.7 title 55 → 61.
19. `CHANGELOG.md` (MODIFY) — KPI/Tests Sprint A row 신규.

### 메모리 (4 file)

20-23. `ai-workflow/memory/feat/x-repository-kpi-tests-section/{state.json, session_handoff.md, work_backlog.md, backlog/2026-06-15.md}` (NEW)

## 추적성 영향

- 추가: **신규 ID 0건** (e2e spec 별도 PR + Sprint B/C/D/E 정합 시 일괄 발권 — cross-project lesson §1 의 scope 폭주 방지)
- 갱신: `docs/traceability/report.md` §6 + `docs/planning/kpi-tests-per-domain-scope.md` §9 + `CHANGELOG.md` KPI/Tests row
- Deprecate: 없음
- 매트릭스: 본 row + §1.7.1 의 6 frontend file 추가 (mirror scope 갱신)

## 테스트

- [x] 로컬 backend `go build ./...` — silent PASS
- [x] 로컬 backend `bash scripts/check-openapi-yaml-lint.sh` — PASS (paths 84 → 86)
- [x] 로컬 backend `go build ./internal/httpapi/...` — silent PASS (2 신규 handler + interface + store + mock)
- [ ] CI: backend-unit / backend-integration / e2e shard 1/2/3 (frontend 변경 trigger) / openapi-yaml-lint
- [ ] frontend `npm run build` + `npm run test` — **CI 검증 위임** (Sprint A 의 frontend unit test 미작성, e2e spec 별도 PR)
- [ ] 수동 검증: production 환경 (postgres + 1h cycle) 에서 KPI/Tests sub-section 정상 표시 — staging Gitea SOP 별도

## 관련 컨셉 / ADR

- 컨셉: [`kpi-tests-per-domain-scope.md`](docs/planning/kpi-tests-per-domain-scope.md) (Sprint A 1차 진입, B/C/D/E 5 sprint hook)
- 컨셉 §2.1 Repository sub-section (raw, weight=1) — 본 PR scope
- 컨셉 §4.1 데이터 흐름: `GET /api/v1/repositories/:id/quality?window=N` + `GET /api/v1/repositories/:id/test-results?window=N` (본 PR 의 backend 2 endpoint 정합)
- 관련 sprint: `feat/x5-gitea-hourly-pull` (X-5) 의 `stateToEventType` 정합 (event_type enum)
- 관련 sprint: `work_260607-a-dashboard-improvements` (gemini) 의 contribution_weight 가중치 정공법 (Sprint B 정합 기반)

## 잔여 follow-up (본 PR scope 외, 사용자 결정 영역)

- **Sprint B** — Project 단위 KPI/Tests sub-section (contribution_weight 가중치 적용 rollup). backend: `GET /api/v1/projects/:id/kpi/quality` + `/kpi/test-pass-rate`. (1~2 ses)
- **Sprint C** — Platform 단위 KPI/Tests sub-section (sub-project rollup). backend: `GET /api/v1/platforms/:id/kpi/quality` + `/kpi/test-pass-rate` + `/progress`. (1~2 ses)
- **Sprint D** — Sidebar 의 `analyticsMenu` 분리 + 글로벌 `/kpis` `/tests` 에 도메인 picker 추가. (0.5 ses)
- **Sprint E** — 글로벌 페이지 옵션 A (deprecated) / B (cross-reference picker) / C (legacy) 결정 + legacy 정리. (0.3 ses)
- **e2e spec** — `frontend/tests/e2e/repository-kpi-tests-section.spec.ts` + smoke manifest 등록 (TC-REPO-KPI-TESTS-01). (0.3 ses, 별도 PR)
- **별도 repository_tests table** 도입 — Sprint A 한정 표현 (build_runs 분포) 의 한계. Sprint F 또는 후속. (별도)
