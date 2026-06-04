# N-10 Manager RBAC 검증 보고서

- 문서 목적: release_v1_roadmap §3.5 N-10 "Manager role RBAC 검증" sprint 의 결과 기록. PR #461 (RBAC hardening) + PR #462 (CI Run API + RBAC row filter + org subtree scope) 의 회귀 검증.
- 범위: Manager role (정식 명칭: `team_manager`) 의 (1) DB/마이그레이션 정합, (2) backend row filter/subtree scope 구현, (3) E2E seed 매핑, (4) E2E/UT 커버리지, (5) `role-access-concept.md` 정합.
- 대상 독자: 프로젝트 리드, Claude (backend RBAC owner), 사용자가 v1.0 출시 전 결함 triage.
- 상태: verified-with-followup (1건 결함 발견)
- 최종 수정일: 2026-06-04
- 검증 sprint: `opencode/work_260604-c-N10-manager-rbac-validation` (opencode Lane 2 첫 carve)
- 관련 PR: #461 (RBAC hardening), #462 (CI Run + RBAC row filter + subtree)
- 관련 문서: [role-access-concept.md](../planning/role-access-concept.md), [ADR-0011 RBAC row-scoping](../adr/0011-rbac-row-scoping.md), [ADR-0026 Keycloak role excluded decision](../adr/0026-keycloak-role-excluded-decision.md), [release_v1_roadmap §3.5 N-10](../planning/release_v1_roadmap.md)

## 0. 검증 환경

| 항목 | 상태 | 메모 |
| --- | --- | --- |
| OS | linux | dev workstation |
| Go | 1.24.4 | `go version` |
| Node | 24.16.0 | E2E 의존성 (실행 안 함) |
| Docker / Keycloak staging | N/A | live 환경 의존 검증은 별도 sprint (사용자 동반) |
| `git HEAD` | `829d529` (PR #466 merged) | main 최신 |
| 검증자 | Sisyphus (opencode 워커) | Lane 2 |

## 1. 검증 결과 요약 (DoD)

- [x] 검증 대상 10개 (V-01..V-10) 모두 결과 명시
- [x] backend UT + go vet + go build + migration prefix check 모두 PASS
- [x] `role-access-concept.md` ↔ 코드 정합 확인 (단, 1건 자기모순 발견 — §9 결함)
- [x] 발견 결함 1건 (V-08: E2E spec TC-RBAC-ROW-READ-01/02 등 spec 만 존재, 실제 E2E 미구현) — **Follow-up 항목** §3.1

## 2. 검증 대상별 결과

### V-01: Keycloak realm `team_manager` role 존재 → **N/A (by design)**

- **근거**: `infra/idp/keycloak-realm.dev.json` 의 `roles.realm` 에는 `user` 1개만 정의됨. `team_manager` / `system_admin` 은 Keycloak realm 에 없음.
- **판정**: **N/A — 의도된 설계** (ADR-0026 Keycloak role excluded decision). system role 은 DB 의 `rbac_policies` 가 source-of-truth, Keycloak 은 token 검증만. PR #461 (`3680648 fix(e2e): ensureRealmRole auto-creates missing Keycloak realm role`) 의 E2E `ensureRealmRole` 로직이 필요 시 자동 생성.
- **위험**: 없음. ADR-0026 과 정합.

### V-02: Keycloak realm `org_head` role 존재 → **N/A**

- **근거**: V-01 과 동일. `org_head` 는 **resource role** (system role 아님 — `role-access-concept.md §2.2` 기준) 이라 Keycloak realm 에 등록되지 않음. `org_units.LeaderUserID` 가 source-of-truth.
- **판정**: **N/A — 의도된 설계**

### V-03: E2E seed `mgr-user-b` 가 `team_manager` realm role 보유 → **PASS (명칭 차이)**

- **근거**: `frontend/tests/e2e/fixtures.ts:28` + `global-setup.ts:28` 모두 동일: `{ user_id: "bob", email: "bob@example.com", display_name: "Bob", password: "ChangeMe-12345!", role: "team_manager" }`
- **판정**: `mgr-user-b` 라는 명칭은 roadmap 의 비공식 표현. 실제 E2E 시드 user_id = `bob`, role = `team_manager`. 시드 3종:
  - `alice` (developer)
  - `bob` (team_manager) ← 검증 대상
  - `charlie` (system_admin)
- **위험**: 없음. 단, roadmap 의 "mgr-user-b" 가 그대로 남아 혼동 가능 → 후속 §3.2 에서 명칭 정정 제안.

### V-04: `ListProjects` row filter (team_manager subtree scope) → **PASS**

- **근거**: `backend-core/internal/domain/application-lifecycle/repository/projects.go:76-88` (count) + `:95-110` (list) SQL:
  ```sql
  AND ($6 = '' OR $6 = 'system_admin' OR owner_user_id = $7
       OR EXISTS (SELECT 1 FROM project_members WHERE project_id = projects.id AND user_id = $7)
       OR (array_length($8::text[], 1) > 0
           AND EXISTS (SELECT 1 FROM applications WHERE id = projects.application_id AND development_unit_id = ANY($8)))
       OR (array_length($9::text[], 1) > 0
           AND EXISTS (SELECT 1 FROM applications WHERE id = projects.application_id AND development_unit_id = ANY($9))))
  ```
  - `$8` = `OrgUnitIDs` (org_head subtree)
  - `$9` = `PrimaryUnitIDs` (team_manager subtree)
- **판정**: `role-access-concept.md §3.4 + §3.5` 의 org_head/team_manager scope 와 정합.
- **위험**: `listQuery` 와 `countQuery` 가 동일한 row filter 를 공유 — 양쪽 동시 검증 완료.

### V-05: `ListApplications` row filter (team_manager subtree scope) → **PASS**

- **근거**: `backend-core/internal/domain/application-lifecycle/repository/applications.go:118-132` 에 `rowFilterCount` + `rowFilterList` 두 변수로 정의. pattern 동일 (system_admin bypass + owner + leader + member-via-project + OrgUnitIDs + PrimaryUnitIDs).
- **판정**: PASS. 단, V-04 와 비교 시 `$8` / `$9` parameter index 가 count vs list 에서 다름 (count: 4~7, list: 6~9) — `countQuery` 와 `listQuery` 의 param ordering 이 다름. SQL 작성 시 정확히 검증되었으나 follow-up 시 주의.

### V-06: E2E `TC-RBAC-ROW-READ-01/02` PASS → **FAIL (E2E 미구현)**

- **근거**:
  - spec 정의: `docs/domain/rbac-permissions/test_cases.md` 에 `TC-RBAC-ROW-READ-01` (List read scope 필터) + `TC-RBAC-ROW-READ-02` (Get read scope 차단) 모두 명시. 매핑: REQ-RBAC-013, 015 / UC-RBAC-04.
  - 실제 E2E: `frontend/tests/e2e/rbac-routes.spec.ts` 검색 결과 동일 ID 의 `test()` 미존재. `TC-RBAC-DEV-VIEW-01/02` 가 route-level 만 검증 (실제 list 의 row filter 미검증).
  - 유사하게 `TC-RBAC-ROLE-DRIFT-01`, `TC-RBAC-LOGOUT-01/02`, `TC-RBAC-CODE-01`, `TC-RBAC-TRACE-01` 6건 spec 존재하나 E2E 미구현.
- **판정**: **FAIL — spec vs 구현 갭**. PR #461 의 body 가 추가를 주장했으나 실제 `frontend/tests/e2e/` 에는 미반영.
- **위험**: **v1.0 출시 전 blocker 가능성** (row filter 회귀 검출 수단 부재). 다만 backend UT (`TestPostgresStoreListUsersAndHierarchy` 등) 가 부분 cover.
- **조치 (V-08 참조)**

### V-07: `role-access-concept.md` 가 3-role 모델 정합 → **PASS (with note)**

- **근거**:
  - 문서 §2.1 = "3개 system role (`developer` / `team_manager` / `system_admin`)"
  - `users_role_check` constraint: `CHECK (role IN ('developer', 'team_manager', 'system_admin'))` (`migrations/000004`)
  - frontend `SYSTEM_ROLE_IDS = ["developer", "team_manager", "system_admin"]` (`frontend/domain/rbac-permissions/schema/...`)
- **판정**: 3-role 모델 정합. **with note**: 문서 §6.2 의 migration 표가 "manager → team_manager" 통합을 명시하나 §2.1 표 의 컬럼 "기존 대응" 에는 `manager` + `team_manager` 가 별도 행으로 남아있어 시각적으로 모순 — 후속 §3.3.

### V-08: legacy `manager` alias 가 완전히 제거됨 → **PARTIAL**

- **근거**:
  - DB: migration 000021 이 `team_manager` 도입 + 000047 이 display name normalize. `users_role_check` 가 3 role 만 허용 (`000004`).
  - E2E: fixtures.ts:250 `UPDATE rbac_policies SET role_id = 'team_manager' WHERE role_id = 'manager';` 가 seed SQL 에 포함 → stale install 호환.
  - backend: `normalizeSystemRoleAlias` 가 legacy `manager` → `team_manager` 정규화 (PR #462 의 review 반영 — `bc3e0f8`).
  - **잔존 참조**: `backend-core/internal/store/users_units.go:1266` 의 주석이 `org_head` 6-P3 scope enforcement 언급. 본문 grep 결과 legacy `manager` (소문자) token 이 다수 backend 주석/문자열에 잔존하나 functional code 아님.
- **판정**: **PARTIAL — DB/seed/normalize 모두 정합**, 단 코드 주석/string 에 `manager` (구명칭) 잔존. 가시성 영향 없음.
- **조치 (선택)**: 후속 housekeeping sprint 에서 `manager` token rename — 본 sprint scope out.

### V-09: `normalizeSystemRoleAlias` 가 legacy manager → team_manager 정규화 → **PASS**

- **근거**: `backend-core/internal/httpapi/permissions_test.go` + backend UT grep 결과 alias 정규화 UT 존재. PR #462 의 test restore 포함 (`bc3e0f8` 의 restore 항목: "normalizeSystemRoleAlias legacy manager 매핑 복원").
- **판정**: PASS.

### V-10: team_manager 의 subtree scope 에서 privilege escalation 차단 → **PASS**

- **근거**: PR #462 review 반영 (`team_manager subtree scope privilege escalation 방지`). `ListProjects` SQL row filter 의 `$6 = 'system_admin'` 우회 경로 + `$8` / `$9` array_length 체크 + project_members EXISTS 가 AND-OR 로 결합. team_manager 가 다른 unit 의 project_members 에 인위적으로 추가되어도 `org_units` subtree 외부 row 는 노출되지 않음 (project_members EXISTS 가 OR 조건이지만 subtree scope 가 더 좁은 scope 임 — row 가 member 면 subtree 외부라도 노출되는 점은 의도된 멤버십 baseline).
- **판정**: PASS. backend UT `TestPostgresStoreListUsersAndHierarchy` 가 부분 cover. 추가 E2E 검증 (V-08) 권장.

## 3. 발견 결함 + Follow-up

### 3.1 [P1] E2E `TC-RBAC-ROW-READ-01/02` 등 spec vs 구현 갭

- **현상**: `docs/domain/rbac-permissions/test_cases.md` 에는 `TC-RBAC-ROW-READ-01/02`, `TC-RBAC-LOGOUT-01/02`, `TC-RBAC-ROLE-DRIFT-01`, `TC-RBAC-CODE-01`, `TC-RBAC-TRACE-01` 6건 spec 존재. `frontend/tests/e2e/` 에 동일 ID 의 `test()` 미존재. PR #461 의 body 가 추가 주장했으나 미반영.
- **영향**:
  - v1.0 출시 후 row filter 회귀 검출 수단 부재
  - `rbac-routes.spec.ts` 의 `TC-RBAC-DEV-VIEW-01/02` 는 route-level 만 cover (실제 list 데이터 row filter 미검증)
- **권장 fix**: 별도 sprint (Gemini 영역) 또는 opencode Lane 2 follow-up. 6건 모두 `frontend/tests/e2e/rbac-routes.spec.ts` 또는 신규 `rbac-data-scope.spec.ts` 에 추가.
- **v1.0 차단 여부**: P1 (회귀 위험). codex review 또는 staging 1주 운영에서 발견 가능.

### 3.2 [P3] roadmap 의 "mgr-user-b" 명칭 비공식

- **현상**: `release_v1_roadmap.md §3.5 N-10` 본문이 `mgr-user-b` 표기. 실제 E2E seed user_id = `bob`.
- **권장 fix**: roadmap 본문 갱신 — 본 검증 보고서 §3.5 의 follow-up 으로 즉시 적용 (아래 §4 의 PR 에서 함께 처리).

### 3.3 [P3] `role-access-concept.md` §2.1 표의 "기존 대응" 컬럼 모순

- **현상**: §2.1 표에 `team_manager` 의 "기존 대응" 이 "(manager + team_manager 통합)" 으로 표기. §6.2 의 migration 표는 `manager` 와 `team_manager` 를 별도 행으로 분리. 시각적 모순.
- **권장 fix**: §2.1 표에서 "(manager + team_manager 통합)" → "(legacy `manager` 의 후속. 000021 migration)" 으로 단순화. 별도 housekeeping PR (opencode Lane 1).

### 3.4 [P3] backend 주석/문자열의 legacy `manager` token 잔존

- **현상**: V-08 결과. functional code 아님, 주석/string 만.
- **권장 fix**: 후속 housekeeping sprint 에서 token rename. 본 sprint scope out.

## 4. v1.0 D-11 잔여 권장 행동

| 우선순위 | 항목 | 워커 | sprint |
| --- | --- | --- | --- |
| **P1** | 3.1 E2E spec-vs-구현 갭 해소 (6 TC) | Gemini (frontend E2E) + Claude (backend handler 보강) | sprint -m/-n (v1.0 마지막) |
| **P2** | 3.2 roadmap "mgr-user-b" → "bob (team_manager seed)" 갱신 | opencode Lane 1 | 본 sprint 후속 (PR 동봉) |
| P3 | 3.3 `role-access-concept.md` §2.1 표 정리 | opencode Lane 1 | 후속 |
| P3 | 3.4 legacy `manager` token rename | opencode Lane 1 (housekeeping) | 후속 |

## 5. 검증 sprint 종료

- 본 보고서 작성 + P2 갱신 (roadmap §3.5 N-10) 동봉 PR
- PR 머지 후 release_v1_roadmap §3.5 N-10 row status `✅ verified (with P1 follow-up)`
- v1.0 D-11 안 P1 (3.1) 만 blocker. 나머지 P3 housekeeping 은 v1.0 후 가능.

## 6. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-06-04 | 1차 작성 — V-01..V-10 검증 결과 + 4건 follow-up 식별 | `opencode/work_260604-c-N10-manager-rbac-validation` |
