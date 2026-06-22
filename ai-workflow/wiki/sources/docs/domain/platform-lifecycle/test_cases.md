---
title: test_cases
type: source
tags: [domain, test_cases.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/platform-lifecycle/test_cases.md]
git_commit: 71c0d2cd
git_branch: chore/260622-wiki-drift-cleanup
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T03:47:55Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# Platform/Project Role-Based View Scope TC

- 문서 목적: 새 역할 모델 (two-dimensional RBAC) 기준 Platform/Project view scope 의 test case 를 정의한다.
- 범위: 3개 system role (developer / team_manager / system_admin) × 4개 resource role (project_member / project_leader / application_leader / org_head) 의 view scope 조합. Application LIST / Project LIST / 상세 조회 / 관리 정보 접근을 검증.
- 대상 독자: Backend 개발자, QA, AI agent.
- 상태: draft
- 최종 수정일: 2026-06-01
- 관련 문서: [role-access-concept.md](../../planning/role-access-concept.md), [project_concept.md](./project_concept.md), [requirements.md](./requirements.md), [architecture.md](./architecture.md), [api.md](./api.md)
- traceability: 본 문서 TC-088 ~ TC-097 은 `docs/traceability/report.md` §3.6 Platform/Project 행의 TC 컬럼에서 참조.

---

## 1. 테스트 환경 구성 (공통)

### 1.1 시드 데이터

```
조직 구조:
  company-root
    └── dept-eng (org_head: user_org_head)
          ├── team-frontend (team_manager: user_team_mgr)
          └── team-backend

Platform:
  app-alpha (LeaderUserID: user_app_leader, DevelopmentUnitID: dept-eng)

Project:
  proj-front (app-alpha 소속, OwnerUserID: user_owner)
    Members:
      - user_dev1   (contributor)
      - user_leader (lead)     ← project leader
  proj-back  (app-alpha 소속)
    Members:
      - user_dev2   (contributor)
  proj-external (다른 부서 app 소속, user_dev1 만 member)
```

### 1.2 테스트 사용자

| Actor | System Role | Resource Role |
|---|---|---|
| `user_dev1` | developer | proj-front member(contributor) + proj-external member |
| `user_dev2` | developer | proj-back member(contributor) |
| `user_leader` | developer | proj-front member(lead) ← project leader |
| `user_app_leader` | developer | app-alpha LeaderUserID ← platform leader |
| `user_org_head` | developer | dept-eng OrgUnit LeaderUserID ← org head |
| `user_team_mgr` | team_manager | team-frontend 소속 |
| `user_sysadmin` | system_admin | — (global) |
| `user_nobody` | developer | project_members 에 없음, org_head 아님 |

---

## 2. Baseline: 멤버십 기반 접근

### TC-088: Developer — 자신이 member 인 project 만 LIST 조회 가능

- **Given**: `user_dev1` (developer, proj-front member)
- **When**: `ListProjects` 호출
- **Then**: 응답에 `proj-front` 포함, `proj-back` 미포함
- **Expected**: 200 OK, data 에 member 인 project 만 반환
- **Trace**: REQ-FR-ROLE-001

### TC-089: Developer — member 가 아닌 project 상세 조회 시 403

- **Given**: `user_dev1` (developer, proj-front member)
- **When**: `GetProject(proj-back)` 호출
- **Then**: 403 Forbidden
- **Expected**: `code: "auth_row_denied"`, `denied_reason: "not_project_member"`
- **Trace**: REQ-FR-ROLE-002

### TC-090: Developer — member 인 project 의 Platform 은 LIST 에 포함

- **Given**: `user_dev1` (developer, proj-front member)
- **When**: `ListPlatforms` 호출
- **Then**: `app-alpha` (proj-front 의 부모 application) 가 응답에 포함
- **Expected**: 200 OK, data 에 app-alpha 포함
- **Trace**: REQ-FR-ROLE-003

---

## 3. Project Leader: 관리 정보 접근

### TC-091: Project Leader — lead 인 project 의 management info 접근 가능

- **Given**: `user_leader` (developer, proj-front 의 project_role=lead)
- **When**: `GetProjectRollup(proj-front)` 호출
- **Then**: rollup 데이터 정상 반환
- **Expected**: 200 OK, rollup/metrics 데이터 포함
- **Trace**: REQ-FR-ROLE-004

### TC-092: Project Leader — member 만 있는 project 의 management info 는 403

- **Given**: `user_leader` 가 proj-back 에 contributor 로 추가됨
- **When**: `GetProjectRollup(proj-back)` 호출 (proj-back 에서는 member, leader 아님)
- **Then**: 403 Forbidden
- **Expected**: `code: "auth_row_denied"`, `denied_reason: "not_project_leader"`
- **Trace**: REQ-FR-ROLE-005

---

## 4. Platform Leader: Platform 관리 정보 접근

### TC-093: Platform Leader — leader 로 지정된 application 의 관리 정보 접근

- **Given**: `user_app_leader` (LeaderUserID = app-alpha)
- **When**: `GetApplicationDashboard(app-alpha)` 호출
- **Then**: dashboard 데이터 정상 반환
- **Expected**: 200 OK, dashboard 데이터 포함
- **Trace**: REQ-FR-ROLE-006

### TC-094: Platform Leader — leader 가 아닌 application 의 관리 정보 403

- **Given**: `user_app_leader` (다른 app 의 leader 아님)
- **When**: 다른 application 의 dashboard 호출
- **Then**: 403 Forbidden
- **Expected**: `code: "auth_row_denied"`
- **Trace**: REQ-FR-ROLE-007

---

## 5. Org Head: 부서 전체 조회

### TC-095: Org Head — 소속 org subtree 전체 project LIST 조회

- **Given**: `user_org_head` (developer, dept-eng 의 org_head)
  - dept-eng subtree = dept-eng + team-frontend + team-backend
  - proj-front (team-frontend) + proj-back (team-backend) 모두 dept-eng subtree 에 포함
- **When**: `ListProjects` 호출
- **Then**: proj-front + proj-back 모두 응답에 포함 (member 가 아니어도 조회 가능)
- **Expected**: 200 OK, subtree 내 모든 project 반환
- **Trace**: REQ-FR-ROLE-008

### TC-096: Org Head — member 가 아닌 project 상세도 subtree 내면 조회 가능

- **Given**: `user_org_head` (dept-eng org_head, proj-back 의 member 아님)
- **When**: `GetProject(proj-back)` 호출
- **Then**: 200 OK (subtree 내 project 이므로)
- **Expected**: 정상 project 상세 반환
- **Trace**: REQ-FR-ROLE-009

---

## 6. Team Manager: 팀 전체 관리

### TC-097: Team Manager — 소속 team 전체 project LIST + 관리 가능

- **Given**: `user_team_mgr` (team_manager, team-frontend 소속)
  - proj-front (team-frontend 소속) + proj-back (team-backend 소속)
- **When**: `ListProjects` 호출
- **Then**: proj-front 만 응답에 포함 (team-frontend scope)
- **Expected**: 200 OK
- **Trace**: REQ-FR-ROLE-010

### TC-098: Team Manager — 소속 team scope 내 project 수정 가능

- **Given**: `user_team_mgr` (team_manager)
- **When**: `UpdateProject(proj-front)` — proj-front 의 metadata 수정
- **Then**: 200 OK
- **Expected**: 정상 수정
- **Trace**: REQ-FR-ROLE-011

### TC-099: Team Manager — team scope 밖 project 수정 시 403

- **Given**: `user_team_mgr` (team_manager, team-frontend 소속)
- **When**: `UpdateProject(proj-back)` — team-backend 소속 project
- **Then**: 403 Forbidden
- **Expected**: `code: "auth_row_denied"`
- **Trace**: REQ-FR-ROLE-012

---

## 7. System Admin: Global Unrestricted

### TC-100: System Admin — 전체 project LIST 조회 (member 아님도 포함)

- **Given**: `user_sysadmin` (system_admin)
- **When**: `ListProjects` 호출
- **Then**: 모든 project 반환 (member 여부 무관)
- **Expected**: 200 OK, data 에 모든 project 포함
- **Trace**: REQ-FR-ROLE-013

---

## 8. Scope Merging (통합 시나리오)

### TC-101: Member + Org Head — 가장 넓은 scope 가 적용됨

- **Given**: `user_dev1` 이면서 동시에 `조직장`으로 승격됨 (developer + org_head)
  - member scope: proj-front + proj-external
  - org_head scope: dept-eng subtree (proj-front + proj-back)
- **When**: `ListProjects` 호출
- **Then**: org_head scope 이 더 넓으므로 proj-front + proj-back + proj-external 반환
- **Expected**: 200 OK, 합집합 결과
- **Trace**: REQ-FR-ROLE-014

---

## 9. Negative Tests

### TC-102: Nobody — member 도 org_head 도 아닌 사용자는 project LIST empty

- **Given**: `user_nobody` (developer, project_members 에 없음, org_head 아님)
- **When**: `ListProjects` 호출
- **Then**: 빈 목록 반환
- **Expected**: 200 OK, `data: []`
- **Trace**: REQ-FR-ROLE-015

### TC-103: Nobody — project 상세 조회 403

- **Given**: `user_nobody` (developer, 권한 없음)
- **When**: `GetProject(proj-front)` 호출
- **Then**: 403 Forbidden
- **Expected**: `code: "auth_row_denied"`
- **Trace**: REQ-FR-ROLE-016

---

## 10. 추적성 매트릭스

| TC ID | REQ-FR | 검증 대상 | 우선순위 | Phase |
|---|---|---|---|---|
| TC-088 | ROLE-001 | Developer project LIST scoping | P0 | Phase 1 |
| TC-089 | ROLE-002 | Developer project detail 403 | P0 | Phase 1 |
| TC-090 | ROLE-003 | Developer application LIST scoping | P0 | Phase 1 |
| TC-091 | ROLE-004 | Project leader management info | P1 | Phase 2 |
| TC-092 | ROLE-005 | Project leader non-lead 403 | P1 | Phase 2 |
| TC-093 | ROLE-006 | Platform leader dashboard | P1 | Phase 2 |
| TC-094 | ROLE-007 | Platform leader non-lead 403 | P1 | Phase 2 |
| TC-095 | ROLE-008 | Org head subtree LIST | P2 | Phase 3 |
| TC-096 | ROLE-009 | Org head detail 조회 | P2 | Phase 3 |
| TC-097 | ROLE-010 | Team manager team LIST | P2 | Phase 3 |
| TC-098 | ROLE-011 | Team manager team-scope update | P2 | Phase 3 |
| TC-099 | ROLE-012 | Team manager outside-scope 403 | P2 | Phase 3 |
| TC-100 | ROLE-013 | System admin unrestricted | P0 | 기존 |
| TC-101 | ROLE-014 | Scope merging (union) | P2 | Phase 3 |
| TC-102 | ROLE-015 | Nobody empty LIST | P0 | Phase 1 |
| TC-103 | ROLE-016 | Nobody 403 detail | P0 | Phase 1 |

---

## 11. 변경 이력

| 일자 | 변경 | 메모 |
|---|---|---|
| 2026-06-01 | 초안 — TC-088 ~ TC-103 (16개) | role-access-concept.md 기준 |
