# 역할 기반 접근 권한 (Role-Based Access) 컨셉

- 문서 목적: Application/Project 도메인에 대한 **새로운 역할 모델**과 **view scope 규칙**을 정의한다. 기존 4축 RBAC 매트릭스 (route-level) + row-level read scoping 의 2계층 구조로 전환하기 위한 컨셉 기준.
- 범위: `developer` / `team_manager` / `system_admin` 3개 system role + `project_member` / `project_leader` / `application_leader` / `org_head` 4개 resource role 의 정의, view scope 규칙, enforcement architecture.
- 대상 독자: Backend/Frontend 개발자, 정책 stakeholder, AI agent.
- 상태: reviewed (Phase 1~3 구현 완료, 2026-06-02)
- 작성일: 2026-06-01
- 최종 수정일: 2026-06-02
- 관련 PR: [#461](https://github.com/ykylee/Devhub_example/pull/461), `deepseek/work_260602` (6-P2, 6-P3)
- 관련 문서: [ADR-0011 (Row scoping)](../adr/0011-rbac-row-scoping.md), [project_concept.md](../domain/application-lifecycle/project_concept.md), [backend_api_contract.md](../backend_api_contract.md), [rbac.go](../../backend-core/internal/domain/rbac.go), [permissions.go](../../backend-core/internal/domain/rbac-permissions/view/permissions.go)

---

## 1. 현재 모델과 문제점

### 1.1 현재 역할 구조

| System Role | applications:view | projects:view | 비고 |
|---|---|---|---|
| `developer` | ✅ view | ✅ view | project_members 기반 row filter 적용 (6-P2) |
| `team_manager` | ✅ view+edit | ✅ full CRUD | org unit subtree scope 적용 (6-P3) |
| `system_admin` | ✅ full | ✅ full | 모든 row 노출 (bypass) |

### 1.2 현재 RBAC enforcement 계층

```
Request → authenticateActor (actor context) → enforceRoutePermission (matrix) 
        → handler (enforceRowOwnership: write only) → store (no filter)
```

- **Route-level**: `routePermissionTable` 로 endpoint 접근 제어 — matrix 기반, row 구분 없음
- **Handler-level (write)**: `enforceRowOwnership` 로 update/delete/archive 시 owner 검증
- **Handler-level (read)**: **없음** — ListApplications / ListProjects / GetProject 모두 store 직접 호출
- **Store-level**: SQL WHERE 절에 member/org-unit filter 없음

### 1.3 핵심 문제

1. ~~developer/manager 가 applications/projects 를 전혀 볼 수 없음~~ → **✅ 6-P2 해결**: developer 에게 `applications:view` / `projects:view` 부여 + project_members 기반 row filter
2. ~~team_manager/sysadmin 은 모든 row 가 노출됨~~ → **✅ 6-P3 해결**: team_manager 에게 org unit subtree scope 적용
3. **project_leader / org_head / team_manager 개념 부재** — resource-level 역할이 system role 과 분리되어 있지 않음
4. **org unit 계층과 권한이 연동되지 않음** — `org_units.LeaderUserID` / `Application.DevelopmentUnitID` / `project_members` 가 read scope 에 미활용

---

## 2. 새 역할 모델 (Two-Dimensional RBAC)

권한을 **2개 축**으로 분리한다:

> **Dimension 1**: System Role (전역, route-level matrix)  
> **Dimension 2**: Resource Role (맥락적, resource membership/ownership 기반)

### 2.1 System Role (Dimension 1)

| System Role | 기존 대응 | 설명 |
|---|---|---|---|
| `developer` | developer 대체 | apps/projects:view **ON** (row-scoped). 멤버십 기반 접근. |
| `team_manager` | **신규** (manager + team_manager 통합) | team-scoped view + management. 기존 manager/team_manager 역할을 통합. |
| `system_admin` | system_admin 유지 | global unrestricted. 기존과 동일. |

**`manager` / `team_manager`** → `team_manager` 로 통합하여 제거. 3개 system role 로 단순화.

### 2.2 Resource Role (Dimension 2)

Resource role 은 system role 과 독립적이다 — developer 든 team_manager 든 아래 resource role 을 가질 수 있다:

| Resource Role | 식별 기준 | 부여되는 권한 |
|---|---|---|
| `project_member` | `project_members` row 존재 | 해당 project + 연결 application **조회** |
| `project_leader` | `project_members.project_role = 'lead'` | 위 + 해당 project **관리 정보** (롤업/메트릭/리스크) |
| `application_leader` | `Application.LeaderUserID` | 해당 application **관리 정보** + 멤버 관리 |
| `org_head` | `org_units.LeaderUserID` | 소속 org unit **전체 subtree 조회** |

### 2.3 역할 계층 (적용 우선순위)

```
모든 사용자
  ├── project_member   ← project_members 에 포함
  │     └── project_leader  ← project_role = 'lead' (member 의 상위)
  ├── application_leader   ← Application.LeaderUserID
  ├── org_head            ← OrgUnit.LeaderUserID
  └── system role scope   ← developer < team_manager < system_admin
```

실행 시 **가장 넓은 scope** 가 적용된다:
- member 면서 org_head 면 → org_head scope (더 넓음)
- developer 이면서 project_leader 면 → project_leader scope (project 한정 관리)

---

## 3. 역할별 View Scope 상세

### 3.1 일반 개발자 (Developer)

**System Role**: `developer`  
**Resource Role 조건**: `project_members` 에 포함된 project

| 리소스 | Scope 규칙 |
|---|---|
| Project 목록 | WHERE user_id IN (SELECT user_id FROM project_members WHERE user_id = $actor) |
| Application 목록 | WHERE id IN (SELECT application_id FROM projects WHERE id IN (project_members scope)) |
| Project 상세 | IF member OR owner THEN return ELSE 403 |
| Application 상세 | IF 연결된 project 중 member 인 것이 있으면 return ELSE 403 |
| ProjectMembers 정보 | 자신이 속한 project 의 member list (가시성) |

**제한**: management 정보 (롤업, 리스크, 메트릭) 는 볼 수 없음.

### 3.2 Project Leader (과제 리더)

**System Role**: `developer` (동일)  
**Resource Role 조건**: `project_members.project_role = 'lead'`

| 리소스 | Scope 규칙 |
|---|---|
| Project 목록 | 동일 (member 와 동일) |
| Project 상세 | 동일 |
| **Project 관리 정보** | ✅ **리더인 project 에 한해** rollup/metrics/risks 접근 |
| ProjectMember 관리 | ❌ (member role 변경은 team_manager 이상) |
| Application 관리 | ❌ |

**개념**: 일반 개발자와 system role 은 같지만, lead 로 지정된 project 에서 **관리자 뷰**를 제공받음.

### 3.3 Application Leader (Application 책임자)

**Resource Role 조건**: `Application.LeaderUserID = $actor`

| 리소스 | Scope 규칙 |
|---|---|
| Application 상세 | 동일 |
| **Application 관리 정보** | ✅ 해당 application rollup/dashboard/metrics |
| Application metadata 수정 | ✅ (owner 위양 범위 내, ADR-0011 §4.2) |
| ApplicationMember 관리 | ✅ (leader 권한 범위 내) |
| 하위 project 목록 | ✅ leader application 의 모든 project |

### 3.4 조직장 (Org Head)

**Resource Role 조건**: `org_units.LeaderUserID = $actor`  
**System Role**: 모든 system role 가능 (developer/team_manager/system_admin)

| 리소스 | Scope 규칙 |
|---|---|
| Project 목록 | WHERE DevelopmentUnitID IN (org_unit + descendants) |
| Application 목록 | WHERE DevelopmentUnitID IN (org_unit + descendants) |
| Project 상세 | org scope 내 project 는 접근 허용 |
| Application 상세 | org scope 내 application 은 접근 허용 |
| 조직 정보 | GetHierarchy / ListUnitMembers (본인 조직 + 하위) |
| **제한** | org scope 밖은 member 인 project 만 접근 |

**조직장 scope 계산**:
```sql
-- 재귀 CTE 로 하위 org_unit 전부 조회
WITH RECURSIVE subtree AS (
  SELECT unit_id FROM org_units WHERE leader_user_id = $actor
  UNION ALL
  SELECT ou.unit_id FROM org_units ou 
  JOIN subtree st ON ou.parent_unit_id = st.unit_id
)
SELECT * FROM applications WHERE development_unit_id IN (SELECT unit_id FROM subtree)
```

### 3.5 팀 관리자 (Team Manager)

**System Role**: `team_manager` (신규)  
**Scope 기준**: 자신이 속한 org unit (ResolvePrimaryUnit or CurrentUnitID)

| 리소스 | Scope 규칙 |
|---|---|
| Project 목록 | team scope 내 전체 (membership 불필요) |
| Application 목록 | team scope 내 전체 |
| Project 상세 | team scope 내 project 전체 접근 |
| Application 상세 | team scope 내 application 전체 접근 |
| Project 관리 | ✅ team scope 내 project metadata 수정 |
| ProjectMember 관리 | ✅ team scope 내 member role 변경 |
| Application 관리 | ✅ team scope 내 application metadata 수정 |
| **제한** | team scope 밖은 member 인 project 만 접근 |

**team_manager DefaultPermissionMatrix** (신규 추가):
```go
case string(AppRoleTeamManager):
    return PermissionMatrix{
        ResourceApplications:            {View: true, Edit: true},
        ResourceProjects:                {View: true, Create: true, Edit: true, Delete: true},
        ResourceOrganization:            {View: true, Edit: true},
        // ... 기타 resource 는 developer 와 동등
    }, true
```

### 3.6 총괄 관리자 (System Admin)

**System Role**: `system_admin` (기존 유지)

| 리소스 | Scope 규칙 |
|---|---|
| 전체 | Unrestricted (기존과 동일) |
| Row filter | 적용하지 않음 |

---

## 4. View Scope 통합 규칙 (Execution Order)

```
function getViewScope(actor):
    // 1. system_admin → global
    if actor.role == system_admin: return UNRESTRICTED
    
    // 2. org_head scope 확장
    scopes = []
    if actor is org_head:
        scope_units = getSubtree(actor.leader_unit_ids)
        scopes.append({type: ORG_UNIT, unit_ids: scope_units})
    
    // 3. team_manager scope 확장
    if actor.role == team_manager:
        scope_units = getTeamScope(actor.primary_unit_id)
        scopes.append({type: TEAM_UNIT, unit_ids: scope_units})
    
    // 4. membership baseline (모든 사용자)
    member_project_ids = getMemberProjects(actor.user_id)
    scopes.append({type: MEMBERSHIP, project_ids: member_project_ids})
    
    // 5. project_leader scope (해당 project 관리 정보)
    lead_project_ids = getLeadProjects(actor.user_id)
    scopes.append({type: LEADERSHIP, project_ids: lead_project_ids})
    
    // 6. application_leader scope
    lead_app_ids = getLeadApplications(actor.user_id)
    scopes.append({type: APPLICATION_LEADER, app_ids: lead_app_ids})
    
    return mergeScopes(scopes)  // OR 조건으로 통합
```

**멤버십은 모든 역할의 baseline**: developer/team_manager/system_admin 구분 없이, 멤버로 포함된 요소는 항상 접근 가능.

---

## 5. Enforcement Architecture

### 5.1 계층별 변경 사항

```
Request → authenticateActor (unchanged) 
        → enforceRoutePermission (matrix update: developer 에 apps/projects view ON)
        → enforceRowReadScope (NEW: list 상세에 row filter) 
        → enforceRowOwnership (unchanged: write)
```

### 5.2 Route-Level Matrix 변경

`DefaultPermissionMatrix` 의 `developer` role 에 `applications:view` / `projects:view` 추가:
```go
case string(AppRoleDeveloper):
    return PermissionMatrix{
        ResourceApplications:            {View: true},       // ← 변경: false → true
        ResourceProjects:                {View: true},       // ← 변경: false → true
        // ... 나머지는 동일
    }, true
```

### 5.3 Store-Level Row Filter (신규)

`ListProjects` / `ListApplications` / `GetProject` / `GetApplication` 에 **actor context 기반 WHERE 조건** 추가:

```go
// projects.ListProjects
func (s *PostgresStore) ListProjects(ctx context.Context, opts ProjectListOptions, actor ProjectViewScope) ([]domain.Project, int, error) {
    query := `SELECT ... FROM projects WHERE 1=1`
    args := []interface{}{}
    
    if actor.Scope == SCOPE_GLOBAL {
        // no filter (system_admin)
    } else {
        query += ` AND (id = ANY($N) OR project_id IN (SELECT project_id FROM project_members WHERE user_id = $M))`
        if actor.OrgUnitIDs != nil {
            query += ` OR development_unit_id = ANY($K)`
        }
    }
    // ... 기존 status/query filter 추가
}
```

Actor scope 정보는 `authenticateActor` 이후 context 에 주입된 정보로부터 **request-scoped view scope 객체**를 구성:
```go
type ProjectViewScope struct {
    Mode        ScopeMode  // GLOBAL | ORG_UNIT | TEAM | MEMBERSHIP
    UserID      string
    OrgUnitIDs  []string   // org_head / team_manager 용
    ProjectIDs  []string   // member 용
}
```

### 5.4 Management Info Gating

Project leader / Application leader 에게만 노출되어야 하는 management 정보는 **handler 레벨**에서 gate:

```go
func (h *ApplicationHandler) GetProjectRollup(c *gin.Context) {
    projectID := c.Param("project_id")
    actor := extractActor(c)
    
    // project_leader 또는 team_manager 또는 system_admin 만 접근 가능
    if !h.canViewManagementInfo(c, projectID, actor) {
        c.JSON(http.StatusForbidden, gin.H{"status": "rejected", "code": "auth_row_denied"})
        return
    }
    // ... rollup 계산
}
```

---

## 6. 기존 Entity 와의 정합성

### 6.1 새 역할 ↔ 기존 DB 스키마

| 새 개념 | 기존 스키마 | 상태 |
|---|---|---|
| `project_member` | `project_members` 테이블 | ✅ 이미 존재 |
| `project_leader` | `project_members.project_role = 'lead'` | ✅ 이미 존재 |
| `application_leader` | `Application.LeaderUserID` | ✅ 이미 존재 |
| `org_head` | `org_units.LeaderUserID` | ✅ 이미 존재 |
| `team_manager` | 신규 system role | ❌ 추가 필요 |
| `developer` w/ row scope | matrix 확장만 필요 | 변경 필요 |

### 6.2 System Role 변화

| Current | New | Migration |
|---|---|---|
| `developer` | `developer` (matrix 확장) | matrix 변경 + row filter 추가 |
| `manager` | → **제거** (`team_manager` 로 통합) | 기존 manager 사용자 team_manager 로 migration |
| `team_manager` | → **제거** (`team_manager` 로 통합) | 기존 team_manager 사용자 team_manager 로 migration |
| `system_admin` | `system_admin` (변화 없음) | 없음 |
| **신규** | `team_manager` | migration + matrix 추가 |

### 6.3 Frontend 영향

- `SYSTEM_ROLE_IDS` 에 `team_manager` 추가 ([rbac.types.ts](../../frontend/domain/rbac-permissions/schema/rbac.types.ts))
- `defaultRoles` 에 `team_manager` matrix 추가
- `role-routing.ts` 의 `defaultLandingFor` 에 `team_manager` 경로 추가
- Project 상세 화면: project_leader 일 때 management tab 노출
- Application 상세 화면: leader 일 때 management tab 노출

---

## 7. 구현 접근법 (점진적)

### ~~Phase 1: Matrix 확장 + Row Filter (MVP)~~ ✅ 완료 (PR #461 + 6-P2)

1. `developer` role 의 matrix 에 `ResourceApplications.View = true`, `ResourceProjects.View = true` 적용
2. `ListProjects` / `ListApplications` 에 project_members 기반 row filter 추가
3. 기존 테스트 통과 확인

### ~~Phase 2: Resource Role Enforcement~~ ✅ 완료 (6-P2)

1. `enforceRowReadScope` helper 구현 (actor → view scope 객체 변환)
2. `GetProject` / `GetApplication` 에 read scope gate 추가
3. Project leader management info gate 구현
4. Application leader management info gate 구현

### ~~Phase 3: Team Manager + Org Head~~ ✅ 완료 (6-P3)

1. `team_manager` system role 추가 (migration + matrix seed)
2. Org head scope 구현 (org_units subtree 기반)
3. Team manager scope 구현 (primary_unit_id 기반)
4. 통합 scope merge 로직 구현

### Phase 4: 정책 안정화 (본 문서 — 6-P4)

1. 기존 `manager` / `team_manager` → `team_manager` 마이그레이션 경로 (또는 병렬 유지)
2. Frontend role-based UI 구현
3. E2E 테스트 추가

---

## 8. 오픈 이슈

| 이슈 | 결정 | 근거 |
|---|---|---|
| 기존 `manager` / `team_manager` | ✅ **제거** → `team_manager` 로 통합 | 2026-06-01 결정 |
| team scope 의 정확한 범위 | `primary_unit_id` 기준 하위 subtree 포함 | (TC-095 참조) |
| Org head scope 의 깊이 | `OrgUnit.LeaderUserID` 기준 전체 subtree | (TC-096 참조) |
| Application 접근과 Project 접근 관계 | baseline: project membership 통한 간접 접근 + application_leader 직접 접근 | (TC-090, TC-093 참조) |
| 멤버십 확장 | `project_members.user_id = actor` (1차). 후속 확장 가능 | Phase 1 baseline |

---

## 9. 변경 이력

| 일자 | 변경 | 메모 |
|---|---|---|
| 2026-06-01 | 초안 — Two-Dimensional RBAC 컨셉 정의 | -- |
