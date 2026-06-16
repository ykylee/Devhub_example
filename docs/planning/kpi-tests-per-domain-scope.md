# KPI / Tests 대시보드 위치 정공법 — 도메인 scope 별 sub-section 통합

- 문서 목적: 신규 추가된 글로벌 `/kpis` 및 `/tests` 라우트의 위치를 **각 도메인 페이지 (Platform/Project/Repository 상세) 의 sub-section** 으로 재배치하는 컨셉 정리. 운영자가 "어디서 봐야 하나" 가 직관적으로 Platform > Project > Repository 계층 정합되도록 함.
- 범위: (1) 위치 정공법 컨셉 — Repository/Project/Platform 별 KPI 와 Tests 의 정보 범위와 가중치 적용 차이 (2) 기존 글로벌 `/kpis` `/tests` 의 처리 방향 (3) 후속 작업의 후크 — sub-section component 분리, 가중치 적용 rollup API 정합, 권한/탐색 정합.
- 상태: draft (사용자 confirm 후 별도 sprint 진입)
- 작성일: 2026-06-15
- 대상 독자: Frontend / Backend 트랙 담당자, AI agent, 리뷰어
- 관련 문서:
  - [`gemini/work_260607-a-dashboard-improvements` sprint memory](../../ai-workflow/memory/gemini/work_260607-a-dashboard-improvements/session_handoff.md) — 신규 `/kpis` `/tests` 의 1차 본문 + 1025 unit test PASS
  - [`docs/domain/platform-lifecycle/project_concept.md`](../domain/platform-lifecycle/project_concept.md) §2 운영 계층 모델 (Platform > Repository > Project)
  - [`docs/domain/platform-lifecycle/dashboard_concept.md`](../domain/platform-lifecycle/dashboard_concept.md) — dashboard metric 정공법
  - [`docs/planning/release_v0-1_roadmap.md` §3.5 v0.1.1 milestone](../../docs/planning/release_v0-1_roadmap.md) (X-1 100% 완료, X-2 100% 완료, X-4 100% 완료)
  - `frontend/app/(dashboard)/platforms/[id]/page.tsx` (557 line) — platform 상세 + sub-section 자리 있음 (`metrics_overview.quality_score` + `quality_metrics` + `history_trend`)
  - `frontend/app/(dashboard)/projects/[id]/page.tsx` (719 line) — project 상세 + `editingWeights` + `projectRepositories` + `completionRate` 이미 구현
  - `frontend/app/(dashboard)/repositories/[id]/page.tsx` (13 line) — `RepositoryDashboardView` thin shell
  - `frontend/shared/ui-foundation/layout/Sidebar.tsx` line 24-25 — 현재 KPIs / Tests 메뉴 위치 (baseMenu 의 4-5번째)

## 1. 컨셉 정리 배경

### 1.1 기존 상태

- `gemini/work_260607-a-dashboard-improvements` (2026-06-07) 가 `frontend/app/(dashboard)/kpis/page.tsx` (26.6KB) 와 `frontend/app/(dashboard)/tests/page.tsx` (45.5KB) 신규 추가.
  - `/kpis`: Python 가상 스크립트 실행 엔진 (`executePythonMetric`) + 코드 에디터 UI + 클릭 삽입형 변수 뱃지 + 백분율 매핑 수식
  - `/tests`: 캘린더/간트 타임라인 + 테스트케이스 카탈로그 + 결과 조율 큐 + Recharts 도넛 차트 + Git/CI 자동화 vs 수동 테스트 이원화 탭
- `frontend/shared/ui-foundation/layout/Sidebar.tsx` 의 `baseMenu` 의 4-5번째 칸에 `KPIs` / `Tests` 메뉴 단일 entry 로 표시.

### 1.2 문제 (사용자 지시)

- **단일 글로벌 페이지** 는 운영자의 mental model 과 어긋남. KPI 와 Tests 는 운영 entity (Repository / Project / Platform) 에 종속된 정보.
- **위치 정공법 위반**: Platform > Project > Repository 계층의 운영 모델 ([`project_concept.md §2`](../domain/platform-lifecycle/project_concept.md)) 과 KPI/Tests 의 scope 가 일치하지 않음.
- **가중치 적용 시점 모호**: `/kpis` 페이지의 Python 수식은 global 데이터 소스 (`DATA_SOURCES`) 만 사용 — repository/project 의 `contribution_weight` 가중치 적용 부재.

### 1.3 해결 방향

KPI / Tests 위치를 **세 가지 도메인 scope 별 sub-section** 으로 재배치:

| 위치 | 정보 범위 | 가중치 적용 | 표시 |
|---|---|---|---|
| **Repository 상세** (`/repositories/[id]`) | 1개 repository 의 raw 데이터 | 가중치 미적용 (single repo = weight=1) | raw metric, 단일 commit SHA / PR / build |
| **Project 상세** (`/projects/[id]`) | 1개 project 의 N개 linked repository | `project_repositories.contribution_weight` 가중치 적용 (default 100%) | rollup metric, 1:N repo 가중 평균 |
| **Platform 상세** (`/platforms/[id]`) | 1개 platform 의 N개 하위 project (sub-rollup) | sub-project 들의 metric 를 platform 가중치 (균등 or custom) 적용 | 종합 metric, N:M project rollup |

기존 글로벌 `/kpis` `/tests` 라우트는 **삭제하지 않고**, sidebar 에서는 "Cross-Reference" 또는 "Analytics" 별도 그룹으로 격하 (또는 fully deprecated, 사용자 결정 영역).

## 2. 위치 정공법 결정

### 2.1 Repository 상세 (raw, weight=1)

**위치**: `frontend/app/(dashboard)/repositories/[id]/page.tsx` 의 `RepositoryDashboardView` 안 **KPI sub-section** + **Tests sub-section**.

**정보 범위**: 단일 repository 의 `quality_snapshots`, `build_runs`, `pr_activities`, `repository_tests`, `commits`. 

**KPI 표시 (raw, 가중치 없음)**:
- Quality Score (단일 repo 의 평균)
- Test Pass Rate (단일 repo 의 마지막 N 회 결과)
- Open PR Count
- Recent Build Success Rate (last 7d)
- Test Coverage (단일 repo 의 마지막 측정값)

**Tests 표시**:
- 자동화 테스트 결과 (GitHub Actions / CI 워크플로 결과)
- 수동 테스트 케이스 (단일 repo scope)

**컴포넌트 분리**: `<RepositoryKPISection />` + `<RepositoryTestsSection />` — `frontend/domain/repository-integration/view/` 에 위치.

### 2.2 Project 상세 (가중치 적용 rollup, N:M repo)

**위치**: `frontend/app/(dashboard)/projects/[id]/page.tsx` 의 메인 콘텐츠에 **KPI sub-section** + **Tests sub-section**. 기존 `editingWeights` / `projectRepositories` (with `contribution_weight`) 의 가중치 정공법을 활용.

**정보 범위**: project 하위 N개 linked repository 의 metric.

**KPI 표시 (가중치 적용 rollup)**:
- Weighted Quality Score = Σ(quality_i × weight_i) / Σ(weight_i)
- Weighted Test Pass Rate
- Weighted Open PR Count
- completionRate (이미 구현됨, line 41 `pullRequests` 등 활용)
- **Project 전체 progress** (raw task completion + repo rollup)

**Tests 표시**:
- 모든 하위 repository 의 test 결과 통합 (가중치 적용)
- 수동 테스트 케이스 (project scope)

**컴포넌트 분리**: `<ProjectKPISection />` + `<ProjectTestsSection />` — `frontend/domain/platform-lifecycle/view/` 에 위치.

**기존 정공법 활용**: sprint `work_260607-a-dashboard-improvements` 가 이미 `contribution_weight` 가중치 + `completionRate` 구현. 본 컨셉은 그 정공법을 KPI/Tests sub-section 으로 확장.

### 2.3 Platform 상세 (sub-project rollup, N:M project)

**위치**: `frontend/app/(dashboard)/platforms/[id]/page.tsx` 의 sub-section 자리 (`metrics_overview` + `quality_metrics` + `history_trend` + `linked_dev_requests` 와 공존).

**정보 범위**: platform 하위 N개 project 의 종합 metric.

**KPI 표시 (sub-project rollup)**:
- 평균 KPI = 각 project 의 Weighted KPI 를 project 가중치 (균등 default or custom) 적용
- **Platform 전체 Trend** (history_trend 그래프)
- VOC 현황 (이미 카드 존재, line 16 `MessageSquare` icon)
- Critical Risk (기존 sub-section)

**Tests 표시**:
- 모든 하위 project 의 test 결과 종합 (sub-rollup)
- Project 들의 progress + test pass rate 통합

**컴포넌트 분리**: `<PlatformKPISection />` + `<PlatformTestsSection />` — `frontend/domain/platform-lifecycle/view/` 에 위치.

### 2.4 글로벌 `/kpis` `/tests` 의 처리 방향

세 가지 옵션:

| 옵션 | 설명 | 권장 |
|---|---|---|
| A. **완전 deprecated (삭제)** | `/kpis` `/tests` 라우트 제거 + sidebar 메뉴 제거. 모든 KPI/Tests 는 도메인 sub-section 으로만. | 사용자 결정 영역 |
| B. **Cross-Reference / Analytics 통합 페이지** | `/kpis` `/tests` 를 도메인 picker (Repository/Project/Platform 선택) 가 있는 통합 analytics 페이지로 격하. sidebar 에는 "Analytics" 그룹 (KPIs, Tests) 으로 별도. | 권장 (사용자 의도 — 모든 위치에서 접근 가능) |
| C. **현상 유지 (legacy)** | 글로벌 페이지 + 도메인 sub-section 공존. | 비권장 (중복) |

**권장**: **옵션 B** — 사용자가 "각 위치에서 참조하는 정보의 범위에 따라 표현하는 방식이 달라질 수 있다" 고 명시했으므로, 글로벌 페이지는 "위치 picker" 역할로 격하 + sidebar 의 "Analytics" 그룹으로 별도. 도메인 상세 페이지의 sub-section 이 1차 진입점.

## 3. Sidebar 위치 정공법

`frontend/shared/ui-foundation/layout/Sidebar.tsx` 의 메뉴 재구성 (제안):

```ts
// 1. baseMenu (모든 사용자)
const baseMenu: MenuItem[] = [
  { href: "/platforms", icon: Zap, label: "Platforms", color: "..." },
  { href: "/repositories", icon: Server, label: "Repositories", color: "..." },
  { href: "/projects", icon: Settings, label: "Projects", color: "..." },
];

// 2. analyticsMenu (모든 사용자, KPI/Tests global view — 도메인 picker 포함)
const analyticsMenu: MenuItem[] = [
  { href: "/kpis", icon: LayoutDashboard, label: "KPI Analytics", color: "..." },
  { href: "/tests", icon: ShieldCheck, label: "Test Analytics", color: "..." },
];

// 3. systemMenu (system_admin 한정, 기존)
const systemMenu: MenuItem[] = [
  { href: "/admin/catalog", icon: Boxes, label: "Admin Catalog", color: "..." },
  { href: "/admin/reception-test", icon: TestTube, label: "Reception Test", color: "..." },
];
```

**변경**: 기존 `baseMenu` 의 KPIs/Tests 를 별도 `analyticsMenu` 그룹으로 이동. 모든 사용자가 접근 가능 (admin-only 아님) — KPI 와 Tests 는 운영의 핵심 정보이므로 일반 사용자도 조회 가능.

## 4. 데이터 흐름 정공법

### 4.1 Repository → Backend API

| KPI | API | 비고 |
|---|---|---|
| Quality Score | `GET /api/v1/repositories/:id/quality?window=30d` | migration 000001 quality_snapshots |
| Test Pass Rate | `GET /api/v1/repositories/:id/test-results?window=30d` | 신규 API |
| Open PR Count | `GET /api/v1/repositories/:id/pull-requests?state=open` (이미 존재, dev-request pattern) | PR list API |
| Build Success Rate | `GET /api/v1/repositories/:id/build-runs?status=success&window=7d` | build_runs |

### 4.2 Project → Backend API (가중치 적용)

| KPI | API | 비고 |
|---|---|---|
| Weighted Quality Score | `GET /api/v1/projects/:id/kpi/quality` | `project_repositories.contribution_weight` 가중치 적용 |
| Weighted Test Pass Rate | `GET /api/v1/projects/:id/kpi/test-pass-rate` | 가중치 적용 rollup |
| completionRate | `GET /api/v1/projects/:id/completion` (이미 존재 가능) | task completion |

### 4.3 Platform → Backend API (sub-project rollup)

| KPI | API | 비고 |
|---|---|---|
| 평균 Quality Score | `GET /api/v1/platforms/:id/kpi/quality` | project rollup (균등 or custom) |
| 평균 Test Pass Rate | `GET /api/v1/platforms/:id/kpi/test-pass-rate` | project rollup |
| 종합 progress | `GET /api/v1/platforms/:id/progress` | project 별 progress 의 평균 |

## 5. 권한 / 탐색 정공법

- KPI / Tests sub-section 은 **도메인 entity 의 RBAC 정합** 사용 (이미 `enforceRoutePermission` 으로 보호).
- **Analytics 페이지** (글로벌 `/kpis` `/tests`): 일반 사용자는 자기 소속 project / repository 만. system_admin 은 전사.
- **외부 사용자 / restricted role**: KPI / Tests sub-section 조회 가능 (raw metric), 단 권한이 없는 entity 의 detail API 는 403.

## 6. 후속 sprint 진입 hook

본 컨셉이 confirm 되면 다음 sprint 들로 분리 가능:

### 6.1 Sprint A: Repository sub-section (1차)

- `frontend/domain/repository-integration/view/RepositoryKPISection.tsx` (NEW)
- `frontend/domain/repository-integration/view/RepositoryTestsSection.tsx` (NEW)
- `frontend/app/(dashboard)/repositories/[id]/page.tsx` 에 import + 배치
- backend: `GET /api/v1/repositories/:id/quality?window=N` + `GET /api/v1/repositories/:id/test-results?window=N` (신규)
- Vitest 12+ unit test + e2e 1 case
- 신규 ID: REQ-FR-XXX + ARCH-XXX + API-XXX + IMPL-XXX + UT-XXX + TC-XXX

### 6.2 Sprint B: Project sub-section (2차, 가중치 정공법)

- `frontend/domain/platform-lifecycle/view/ProjectKPISection.tsx` (NEW)
- `frontend/domain/platform-lifecycle/view/ProjectTestsSection.tsx` (NEW)
- backend: `GET /api/v1/projects/:id/kpi/quality` + `/kpi/test-pass-rate` (가중치 적용 rollup query)
- 기존 `contribution_weight` + `completionRate` 활용

**Sprint B 1차 PR status (2026-06-16)**: **in_progress** (branch `chore/260616-sprint-b-project-kpi`, PR TBD). projectKPI endpoint (1) + ProjectKPISection component (1) + service 1 + schema 1 + Vitest 2 + backend handler test 2 + openapi.yaml + routePermissionTable 등록 + ProjectView 통합 + DomainPicker 의 Project scope `ready: true` 활성화. projectTestResults + ProjectTestsSection + e2e 1 은 follow-up PR (Sprint B-Tests).

### 6.3 Sprint C: Platform sub-section (3차, sub-project rollup)

- `frontend/domain/platform-lifecycle/view/PlatformKPISection.tsx` (NEW)
- `frontend/domain/platform-lifecycle/view/PlatformTestsSection.tsx` (NEW)
- backend: `GET /api/v1/platforms/:id/kpi/quality` + `/kpi/test-pass-rate` + `/progress`

### 6.4 Sprint D: Sidebar 재구성 + 글로벌 페이지 picker

- `frontend/shared/ui-foundation/layout/Sidebar.tsx` 의 `analyticsMenu` 분리
- `frontend/app/(dashboard)/kpis/page.tsx` 에 도메인 picker (Repository/Project/Platform select) 추가
- `frontend/app/(dashboard)/tests/page.tsx` 동일

**Sprint D status (2026-06-16)**: **in_progress** (branch `chore/260616-sprint-d-sidebar-picker`, PR TBD). Sprint A 정합법 + 1차 진입 완료. 후속 (Sprint B/C sub-section + Sprint E legacy 결정) 는 follow-up PR.

### 6.5 Sprint E: 글로벌 페이지 옵션 B/C 결정 + legacy 정리

- 옵션 A (deprecated) vs B (cross-reference picker) vs C (legacy) 결정 후 처리

## 7. 정합 검증 (권장 DoD)

- (1) Repository / Project / Platform 각각의 sub-section 에서 KPI / Tests 정상 표시
- (2) 가중치 적용 시 project 의 weighted metric 이 contribution_weight 변화에 따라 동적 변경
- (3) sidebar 의 "Analytics" 그룹 표시 (모든 사용자)
- (4) Vitest 12+ unit test + e2e 3 case (각 도메인 1 case) PASS
- (5) docs/traceability/report.md §6 에 4 sprint 종합 row 추가
- (6) `docs/domain/platform-lifecycle/dashboard_concept.md` § 갱신 (KPI/Tests sub-section 정합)

## 8. Tier

- **공용** (docs only, 사내 한정 정보 미포함)
- 후속 backend API (4.1~4.3) — 사내 한정 정보 미포함 시 사외 가능 (가중치 / rollup 은 표준 로직)

| 일자 | 변경 | sprint |
|---|---|---|
| 2026-06-15 | 본 컨셉 초안 — KPI/Tests 위치 정공법 (도메인 scope 별 sub-section) + sidebar 재구성 + 5 sprint 진입 hook | (TBD) |
| 2026-06-15 | **Sprint A (Repository sub-section) implemented** — PR #597 (squash `25f2262e`) 머지 완료. 24 file, +1946/-23 line. Backend 2 endpoint (`GET /api/v1/repositories/:id/kpi` + `/test-results`) + `routePermissionTable` 등록 + parseWindowShort helper + `BuildRunListOptions.WindowFrom/To` filter + frontend 2 component (RepositoryKPISection + RepositoryTestsSection) + `ManagerView.tsx` inline 배치 + openapi.yaml +86 path (84 → 86) + test_cases.md (예정) + memory 4 file. 4 follow-up P1/P2 (lucide-react import 닫기, routePermissionTable 등록, ?window=Nd parsing, build-runs window filter) 동시 squash. | `feat/x-repository-kpi-tests-section` (PR #597) |
| 2026-06-16 | **Sprint A follow-up 회귀 가드** — PR #625 (squash `71a227b6`) 머지 완료. 6 file, +755/-10 line. 4 unit test (fetchRepositoryKPI/fetchRepositoryTestResults + RepositoryKPISection/RepositoryTestsSection) + 1 e2e (`repository-kpi-tests-section.spec.ts`, `TC-REPO-KPI-TESTS-01`) + test_cases.md 4 row `[x]` + 변경 이력 row. | `chore/260616-sprint-a-tests-followup` (PR #625) |
| 2026-06-16 | **Sprint D 1차 진입 — Sidebar `analyticsMenu` 분리 + 글로벌 페이지 도메인 picker** — branch `chore/260616-sprint-d-sidebar-picker` 작업 중. (1) `frontend/shared/ui-foundation/layout/Sidebar.tsx` — `baseMenu` 에서 KPIs/Tests 제거 + `analyticsMenu` 별도 그룹 (Analytics 섹션 헤더 + 모든 사용자 접근). (2) `frontend/shared/ui-foundation/components/DomainPicker.tsx` (NEW) — scope tab (Platform/Project/Repository) + entity list + 미준비 scope 'sub-section 예정' 배지. Repository scope 만 ready=true (Sprint A 활성화), Project/Platform 은 Sprint B/C 와 함께 fetch 추가. (3) `/kpis` + `/tests` 페이지 상단에 picker 통합 (legacy 본문은 그대로 유지, 회귀 0). (4) DomainPicker unit test 8 case. (5) Sprint D status row 추가. 본 Sprint D 는 picker + Sidebar 의 1차 진입. Sprint B (Project sub-section) + Sprint C (Platform sub-section) + Sprint E (legacy 결정) 는 follow-up. | `chore/260616-sprint-d-sidebar-picker` (TBD) |
| 2026-06-16 | **Sprint B 1차 진입 — Project 가중치 적용 rollup** — branch `chore/260616-sprint-b-project-kpi` 작업 중. (1) backend `GET /api/v1/projects/:project_id/kpi` (NEW) + `ComputeProjectWeightedKPI` + `CountProjectOpenAndMergedPRs` store method 2개 (PostgresStore, application.go 의 `domain.ProjectWeightedKPI` struct). (2) `routePermissionTable` 의 `/projects/:project_id/kpi` row 등록 (deny-by-default 회귀). (3) openapi.yaml +1 path (87 → 87, inline schema). (4) PlatformStore interface + 2 method + memoryPlatformStore + fakeViewPlatformStore override (PR #597 P1 #2 fix 회귀 정합). (5) backend handler test 2 case (TestProjectKPI_Happy / TestProjectKPI_InvalidWindow). (6) frontend `ProjectKPISection` component + `fetchProjectKPI` service + `project-kpi.types` schema. (7) ProjectView 통합 (`frontend/app/(dashboard)/projects/[id]/page.tsx` 의 left column 첫 child). (8) DomainPicker 의 Project scope `ready: true` 활성화. (9) Vitest 2 (service + component). projectTestResults + ProjectTestsSection + e2e 1 은 follow-up PR (Sprint B-Tests). | `chore/260616-sprint-b-project-kpi` (TBD) |
| 2026-06-16 | **Sprint B-Tests 진입 — Project 가중치 적용 test results** — branch `chore/260616-sprint-b-project-tests` 작업 중. PR #627 (Sprint B 1차) 의 follow-up. (1) backend `GET /api/v1/projects/:project_id/test-results` (NEW) + `ListProjectTestResults` store method (PostgresStore) + `domain.ProjectWeightedTestResults` + `domain.ProjectBuildRun` struct (application.go) — CTE 3개 (linked_repos / weighted / totals_rows / recent_rows) 단일 round-trip + weighted_pass_rate + multi-repo recent. (2) `routePermissionTable` 의 `/projects/:project_id/test-results` row 등록 (deny-by-default 회귀). (3) openapi.yaml +1 path (87 → 88, inline schema). (4) PlatformStore interface + ListProjectTestResults + memoryPlatformStore + fakeViewPlatformStore override. (5) backend handler test 3 case (Happy / InvalidWindow / InvalidLimit). (6) frontend `ProjectTestsSection` component (Recharts 도넛 + status distribution + recent runs + window selector + (weighted) 라벨 + multi-repo `repository_full_name` 컬럼) + `fetchProjectTestResults` service + `project-tests.types` schema. (7) ProjectView 통합 (`ProjectKPISection` 다음 child). (8) Vitest 2 (service 5 case + component 5 case). (9) E2E 1 case (TC-PROJ-KPI-TESTS-01, `project-kpi-tests-section.spec.ts`). | `chore/260616-sprint-b-project-tests` (TBD) |
