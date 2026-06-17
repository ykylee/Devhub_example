# Sprint E — 글로벌 페이지 legacy 결정 + Cross-Reference Picker 활성화 (옵션 B)

- 문서 목적: [`kpi-tests-per-domain-scope.md`](./kpi-tests-per-domain-scope.md) 의 Sprint E — 글로벌 `/kpis` `/tests` 페이지의 legacy 본문 결정 + 정공법 확립. 2026-06-16 사용자 결정 = **옵션 B (Cross-Reference / Analytics 통합 페이지)**: legacy 본문 유지 + `DomainPicker` 가 cross-reference 진입점.
- 범위: (1) legacy 본문 영역 deprecation hint banner (2) `DomainPicker` 의 default scope 결정 (3) `routePermissionTable` 의 sub-section 3 endpoint 권한 정합 (4) e2e 회귀 가드 (5) master plan §6.5 status 정합 (6) `docs/traceability/report.md` §6 종합 row.
- 상태: planned → in_progress (branch `chore/260616-sprint-e-legacy-picker-active`)
- 작성일: 2026-06-16
- 결정 근거: [`kpi-tests-per-domain-scope.md` §2.4 옵션 결정 매트릭스](./kpi-tests-per-domain-scope.md#24-글로벌-kpis-tests-의-처리-방향) + 사용자 결정 (2026-06-16, 옵션 B 확정)
- 관련 문서: [`kpi-tests-per-domain-scope.md`](./kpi-tests-per-domain-scope.md) (master), [`release_v0-1_roadmap.md`](./release_v0-1_roadmap.md) (마일스톤), [`docs/traceability/sync-checklist.md`](../traceability/sync-checklist.md) (PR 절차).

## 1. 결정 컨텍스트

`kpi-tests-per-domain-scope.md` §1.3 / §2.4 의 옵션 A/B/C 중 **옵션 B (Cross-Reference / Analytics 통합 페이지)** 채택. 권장 이유:

- 사용자가 "각 위치에서 참조하는 정보의 범위에 따라 표현하는 방식이 달라질 수 있다" 고 명시. 글로벌 페이지는 "위치 picker" 역할로 격하 + sidebar 의 "Analytics" 그룹으로 별도.
- 도메인 상세 페이지의 sub-section (Sprint A/B/C 가 main에 이미 머지) 이 1차 진입점. 글로벌은 cross-reference.
- 옵션 A (deprecated) 는 운영자가 통합 cross-platform view 를 보고 싶을 때 진입점 없음.
- 옵션 C (legacy 공존) 는 sub-section 활성화 후 중복 가능성.

## 2. Sprint E scope — 옵션 B 의 정공법

### 2.1 Legacy 본문 유지 (deprecate 표시만)

**변경**: `/kpis` `/tests` 페이지의 기존 legacy 본문 (KPIItem Python 스크립트 / 캘린더 타임라인 / Recharts 도넛 / 테스트케이스 카탈로그 / 결과 조율 큐) **삭제하지 않음**. 단 페이지 상단 또는 DomainPicker 영역에 **명시적 deprecation hint banner** 추가.

**Banner 내용** (안):
- "이 페이지는 cross-reference picker 역할입니다. 각 도메인 (Repository/Project/Platform) 의 상세 페이지의 KPI/Tests sub-section 이 1차 진입점입니다. 본문의 legacy 위젯은 v0.1.0 이전 구현의 잔재이며 추후 deprecate 예정."
- 시각: 기존 `Info` icon + 흐린 톤. 

**위치 결정** (안): `DomainPicker` 의 helper hint footer (`DomainPicker.tsx` line 128-135) 의 본문을 더 명시적으로 확장. 또는 페이지 자체의 상단 (DashboardHeader 위/아래) 에 별도 banner component.

### 2.2 DomainPicker 의 default scope 결정

**현재**: `defaultScope="repository"` (DomainPicker.tsx line 47).

**검토 옵션**:
- **repository** (Sprint A, weight=1, 가장 raw) — 신규 사용자 onboarding 시 직관적
- **project** (Sprint B, 가중치 적용) — 운영자가 가장 많이 봄
- **platform** (Sprint C, sub-project rollup) — manager/admin view 의 종합

**사용자 결정 영역**. 본 sprint 에서 결정하지 않을 경우 default = "repository" 유지 (회귀 0). Sprint E 의 핵심 scope 아님.

### 2.3 Sub-section 3 endpoint 의 routePermissionTable 정합

**이미 등록됨** (PR #597 Sprint A, #627 Sprint B, #630 Sprint C):
- `GET /repositories/:repository_id/kpi`
- `GET /repositories/:repository_id/test-results`
- `GET /projects/:project_id/kpi`
- `GET /projects/:project_id/test-results`
- `GET /platforms/:platform_id/kpi`
- `GET /platforms/:platform_id/test-results`

**Sprint E 검증**: `routePermissionTable` 의 resource 매트릭스 + handler test 의 권한 케이스 (403 unauthorized role, 404 entity 없음) 가 모두 active. **회귀 가드 강화** — 본 sprint 에서 deny-by-default 회귀 test 1 case 추가.

### 2.4 E2E 회귀 가드 — `analytics-picker.spec.ts`

**신규**: `frontend/tests/e2e/analytics-picker.spec.ts`

**Test 1**: `/kpis` 진입 → DomainPicker 의 3 scope tab (Repository/Project/Platform) 모두 표시 + 각 scope 의 entity list 가 fetch 결과로 표시 + entity 클릭 시 해당 도메인 상세 페이지로 redirect.

**Test 2**: `/tests` 동일 검증.

기존 `repository-kpi-tests-section.spec.ts` 의 `page.route()` mock 정공법 차용. Backend mock 대상:
- `GET /api/v1/repositories` (entity list)
- `GET /api/v1/projects` (entity list)
- `GET /api/v1/platforms` (entity list)

### 2.5 Master plan status 정합

**`docs/planning/kpi-tests-per-domain-scope.md`**:
- §2.4 옵션 결정 row 추가: **"옵션 B 확정 (2026-06-16, 사용자 결정)"**
- §6.5 status row 갱신: `in_progress` → `done` (Sprint E 머지 후)
- §6 follow-up hook: "Sprint B-Projects Picker follow-up" row status = "TBD" (clean up 시 deleted, 본 sprint 에서 미재개; 사용자 결정 영역)

**`docs/traceability/report.md`**:
- §6 종합 row 추가: Sprint E 종합 (legacy 결정 + deprecation hint + e2e 가드 + master plan 정합)

### 2.6 OpenAPI 정합 (선택)

`/kpis` `/tests` 자체는 frontend-only 페이지 (backend endpoint 없음). openapi.yaml 에 신규 path 추가 불필요. 단 **README 또는 docstring** 에 "글로벌 페이지는 cross-reference picker, sub-section 의 1차 endpoint 는 §repositories/:id/kpi, §projects/:id/kpi, §platforms/:id/kpi" 명시 가능.

## 3. Sub-task 분해

| ID | Sub-task | 영역 | DoD |
| --- | --- | --- | --- |
| **E-1** | `/kpis` 페이지 상단 deprecation hint banner | FE | DomainPicker 위 또는 helper hint footer 확장. 회귀 0. |
| **E-2** | `/tests` 페이지 상단 deprecation hint banner | FE | 동일. |
| **E-3** | `DomainPicker` default scope 결정 (사용자) | FE | 결정 또는 "repository" 유지 명시. |
| **E-4** | sub-section 3 endpoint 의 deny-by-default 회귀 test 1 case | BE | `handler_test.go` 1 case 추가 (unauthorized role = 403). |
| **E-5** | E2E `analytics-picker.spec.ts` (2 test) | FE | `frontend/tests/e2e/analytics-picker.spec.ts` 신규. 3 scope tab + entity list + redirect. |
| **E-6** | master plan `kpi-tests-per-domain-scope.md` §2.4 / §6.5 정합 | docs | 옵션 B 결정 + status row 갱신. |
| **E-7** | `docs/traceability/report.md` §6 row | docs | Sprint E 종합 row. |
| **E-8** | OpenAPI / docstring cross-reference note | docs | sub-section 1차 endpoint 명시. |

## 4. 변경 영향 분석

### 4.1 Frontend (legacy 본문 + DomainPicker)

- **회귀 위험**: legacy 본문 유지 + banner 추가는 **페이지 본문 변경 0**. DomainPicker 의 helper hint footer 만 확장. **회귀 0**.
- **default scope 결정**: 변경 시 (예: "project") entity list 의 default 가 바뀜. e2e `analytics-picker.spec.ts` 의 assertion 이 그에 맞춰 조정.
- **banner component**: `frontend/shared/ui-foundation/components/` 에 신규. 기존 `Info` icon 재사용.

### 4.2 Backend (routePermissionTable)

- **변경 0**: sub-section 3 endpoint 의 권한 row 가 모두 active. 회귀 test 1 case 추가만.

### 4.3 Docs

- `kpi-tests-per-domain-scope.md` §2.4 / §6.5 갱신
- `docs/traceability/report.md` §6 row 추가

## 5. DoD (Definition of Done)

1. ✅ `/kpis`, `/tests` 페이지 상단 deprecation hint banner 표시
2. ✅ E2E `analytics-picker.spec.ts` 2 test PASS (CI)
3. ✅ Backend sub-section 3 endpoint 의 deny-by-default 회귀 test 1 case PASS (`go test ./...`)
4. ✅ Vitest `DomainPicker` 3 scope tab + entity link unit test PASS
5. ✅ `docs/planning/kpi-tests-per-domain-scope.md` §2.4 옵션 B 결정 + §6.5 status = `done` 갱신
6. ✅ `docs/traceability/report.md` §6 Sprint E 종합 row 추가
7. ✅ Tier self-check: 사외 (사내 한정 정보 미포함)
8. ✅ OpenAPI 변경 0 (frontend-only 페이지) — 단 docstring cross-reference note 는 README 또는 주석

## 6. Tier

- **사외** (GitHub main)
- frontend only + docs + backend 회귀 test 1 case (사내 한정 정보 미포함)
- sub-section 3 endpoint 의 routePermissionTable 권한 검증은 backend + frontend 모두 사외 가능 (RBAC 표준 로직)

## 7. 변경 이력

| 일자 | 변경 | sprint |
|---|---|---|
| 2026-06-16 | 본 sprint plan 작성 — 옵션 B 확정 (사용자 결정, 2026-06-16), Sprint E 1차 진입. branch `chore/260616-sprint-e-legacy-picker-active`. | (본 sprint) |
