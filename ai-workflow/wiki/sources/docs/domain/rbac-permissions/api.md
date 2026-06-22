---
title: api
type: source
tags: [domain, api.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/rbac-permissions/api.md]
git_commit: baf1cf24
git_branch: main
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T03:37:45Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# rbac-permissions 도메인 API

- 문서 목적: RBAC 정책 조회·갱신·권한 강제 계약(API-26..29, 38..40)과 연계 정책을 정의한다.
- 범위: API-26..29 (정책 CRUD), API-30/API-31 (subject role, **폐기**), API-38 (route 매핑 표), API-39 (deny-by-default), API-40 (cache).
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: accepted
- 최종 수정일: 2026-06-02 (Two-Dimensional RBAC 정합화 + role drift fail-closed/read-scope 규칙 반영)
- 관련 문서: [도메인 README](./README.md), [requirements.md](./requirements.md), [architecture.md](./architecture.md), [master API](../../backend_api_contract.md), [ADR-0002](../../adr/0002-rbac-policy-edit-api.md), [ADR-0007](../../adr/0007-rbac-enforcement.md), [ADR-0011](../../adr/0011-rbac-row-scoping.md), [ADR-0020](../../adr/0020-account-user-management-boundary.md)

## 개요

ADR-0002 채택 (2026-05-08) 으로 *DB-backed RBAC matrix + write API + per-resource 4-boolean 모델* 을 source-of-truth 로 한다. 본 §의 spec 은 M1 PR-G2~G6 의 구현 대상이다.

## 1. 모델

### 1.1 Role

- 시스템 정의 role 3종 (immutable id, name 변경 불가): `developer`, `team_manager`, `system_admin`.
- 사용자 정의 role: `system_admin` 만 생성 가능. id 는 `custom-{slug}` 패턴 권장. 시스템 role 의 권한을 *상회* 하는 매트릭스 설정 가능 (단 본 단계에서는 enforcement 가 *최상위 1개 role 단일 평가* 라 multi-role 합산 정책은 미해결).
- 응답 wire 형식: `id` (snake_case 또는 `custom-*`), `name` (display, 자유 문자열), `description`, `permissions`.

### 1.2 Resource (11종)

| resource | 의미 |
| --- | --- |
| `infrastructure` | 인프라 토폴로지, 노드/엣지, dashboard 메트릭, system action command, command lifecycle |
| `pipelines` | repository, issue, pull request, CI run, CI log |
| `organization` | users, org units, hierarchy, unit members, subject role assignment |
| `security` | risks, risk mitigation command, RBAC policy 자체 (조회·편집) |
| `audit` | audit-logs 조회 (audit 생성은 시스템 전용) |
| `applications` | application 관리 |
| `platform_repositories` | application-repository 매핑 관리 |
| `projects` | project 관리 |
| `scm_providers` | SCM provider 활성화/설정 관리 |
| `dev_requests` | 개발 의뢰(DREQ) 조회/관리 |
| `dev_request_intake_tokens` | DREQ 외부 수신 토큰 발급/관리 |

> 본 5종은 ADR-0002 §4.2 채택. 신규 자원이 추가되면 본 contract 갱신 + 매핑 표 (§7) 갱신.

### 1.3 Action (4축)

| action | 의미 |
| --- | --- |
| `view` | GET / 조회 |
| `create` | POST / 생성 또는 명령 발행 |
| `edit` | PATCH/PUT / 수정 또는 멤버 갱신 |
| `delete` | DELETE / 삭제 또는 회수 |

각 (resource, action) 좌표는 boolean. 예: `{ "security": { "view": true, "create": true, "edit": false, "delete": false } }` 는 risk 조회·mitigation 발행은 가능하지만 risk 수정/삭제는 불가.

### 1.4 Audit append-only invariant

`audit` resource 의 `create`, `edit`, `delete` 는 **모든 role 에 대해 false 를 강제** (UI 노출도 readOnly). audit 항목은 시스템 코드가 작성하며 사용자 API 로 직접 mutation 하지 않는다. seed 또는 PUT 으로 true 를 설정하려 해도 store 가 invariant 검증으로 거부한다 (PR-G3 도메인 규칙).

## 2. 시스템 default policy 매트릭스

시스템 role 3종의 *기본* 매트릭스는 다음과 같으며, M0 sprint 의 `requireMinRole` enforcement 와 *완전히 호환* 된다 (PR-G5 의 `requirePermission` 마이그레이션 시 거동 보존). seed 시 store 에 시스템 role row 로 영속화 (PR-G3).

| role \ resource | infrastructure | pipelines | organization | security | audit |
| --- | --- | --- | --- | --- | --- |
| `developer` | view | view | view | view | — |
| `team_manager` | view | view | view, edit | view, create | view |
| `system_admin` | view, create, edit, delete | view, create, edit, delete | view, create, edit, delete | view, create, edit, delete | view |

> `audit` 의 create/edit/delete 는 §1.4 invariant 로 모든 role 에서 false. system_admin 도 view 만 true.
> 레거시 `manager`/`team_manager` 는 신규 부여 금지이며 migration alias 로만 허용.

## 3. `GET /api/v1/rbac/policies` (API-26)

시스템 정의 + 사용자 정의 role 전체와 각 role 의 4축 매트릭스를 조회. 매핑 누락 resource (시스템 자원 추가됐는데 role 매트릭스가 미갱신) 는 응답에 `view=false, create=false, edit=false, delete=false` 로 채워 반환.

### 3.1 권한

`security:view` (모든 시스템 role 이 보유 — 자기 권한 가시성).

### 3.2 응답 예시

```json
{
  "status": "ok",
  "data": [
    {
      "id": "developer",
      "name": "Developer",
      "description": "개발자 대시보드, 본인 관련 repository/CI/risk 조회 권한",
      "system": true,
      "permissions": {
        "infrastructure": { "view": true,  "create": false, "edit": false, "delete": false },
        "pipelines":      { "view": true,  "create": false, "edit": false, "delete": false },
        "organization":   { "view": true,  "create": false, "edit": false, "delete": false },
        "security":       { "view": true,  "create": false, "edit": false, "delete": false },
        "audit":          { "view": false, "create": false, "edit": false, "delete": false }
      }
    },
    {
      "id": "team_manager",
      "name": "Team Manager",
      "description": "팀 범위 운영 및 권한 관리",
      "system": true,
      "permissions": {
        "infrastructure": { "view": true, "create": false, "edit": false, "delete": false },
        "pipelines":      { "view": true, "create": false, "edit": false, "delete": false },
        "organization":   { "view": true, "create": false, "edit": true, "delete": false },
        "security":       { "view": true, "create": true,  "edit": false, "delete": false },
        "audit":          { "view": true, "create": false, "edit": false, "delete": false }
      }
    },
    {
      "id": "system_admin",
      "name": "System Admin",
      "description": "시스템 설정, 조직/사용자 관리, 운영 command 관리 권한",
      "system": true,
      "permissions": {
        "infrastructure": { "view": true, "create": true, "edit": true, "delete": true },
        "pipelines":      { "view": true, "create": true, "edit": true, "delete": true },
        "organization":   { "view": true, "create": true, "edit": true, "delete": true },
        "security":       { "view": true, "create": true, "edit": true, "delete": true },
        "audit":          { "view": true, "create": false, "edit": false, "delete": false }
      }
    }
  ],
  "meta": {
    "policy_version": "2026-05-08.adr-0002.v1",
    "source": "rbac_policies_store",
    "editable": true,
    "system_roles": ["developer", "team_manager", "system_admin"]
  }
}
```

## 4. `PUT /api/v1/rbac/policies` (API-27)

전체 role 또는 특정 role 의 매트릭스를 갱신한다. 시스템 role 의 *id, name, system flag* 는 변경 불가 (store invariant). 시스템 role 의 *permissions* 만 변경 가능.

### 4.1 권한

`security:edit` (default policy 상 `system_admin` 단독).

### 4.2 요청 예시

```json
{
  "roles": [
    {
      "id": "team_manager",
      "permissions": {
        "pipelines": { "view": true, "create": true, "edit": false, "delete": false }
      }
    }
  ]
}
```

- 부분 갱신 — 응답에서는 partial diff 가 적용된 *전체 매트릭스* 를 반환 (§3 응답과 동일 shape).
- 매트릭스 갱신 시 `audit` resource 의 view 외 다른 action 을 true 로 설정하면 422 + `audit_invariant_violation` 거부 (§1.4).
- `auth.policy_unmapped` audit (§8) 와 `rbac.policy.updated` audit 가 동일 트랜잭션에 기록.

### 4.3 응답

200 + 전체 role 응답 (§3). audit log 에 `rbac.policy.updated` 항목 1건 (target_type=`rbac_role`, target_id=role id). payload 에 `before`/`after` 매트릭스 diff.

## 5. `POST /api/v1/rbac/policies` (사용자 정의 role 생성) (API-28)

사용자 정의 role 신규 생성. id 는 `custom-{slug}` 검증 (필수 prefix), name 은 자유 문자열, permissions 는 §1.3 의 4축 boolean 매트릭스.

### 5.1 권한

`security:edit` (system_admin).

## 6. `DELETE /api/v1/rbac/policies/:role_id` (사용자 정의 role 삭제) (API-29)

사용자 정의 role 만 삭제 가능. 시스템 role (`developer`, `team_manager`, `system_admin`) 은 422 + `system_role_not_deletable`. 삭제 직전 해당 role 이 할당된 subject 가 있으면 store 가 cascade 또는 422 거부 — *cascade 거부* 채택 (subject 가 있으면 422 + `role_in_use`). 호출자가 Keycloak admin console 의 group membership 변경 (ADR-0020 결정 D — sprint -d 의 `PUT /rbac/subjects/.../roles` 폐기 이후 단일 경로) 으로 재할당한 뒤 event listener sync 가 DevHub `users.role` 컬럼 갱신을 기다린 후 삭제.

### 6.1 권한

`security:edit` (system_admin).

## 7. 폐기 — `GET/PUT /api/v1/rbac/subjects/{subject_id}/roles` (API-30, API-31)

**상태**: 폐기 (sprint `claude/work_260520-d`, [ADR-0020](../../adr/0020-account-user-management-boundary.md) 결정 D, 2026-05-20).

**폐기 이유**: backend-only dead-end (frontend UI 미구현). 실제로 `users.role` 컬럼 직접 read/write 였고 Keycloak group composite 가 실 권한 source. ADR-0020 결정 C 의 event listener 확장 (sprint -f 후속) 이 `users.role` 자동 sync 를 담당 — DevHub UI 가 직접 role assignment 를 노출할 필요 없음. role assignment 는 Keycloak admin console 에서 group membership 으로 처리 + DevHub `users.role` 컬럼은 event listener 가 자동 sync. 본 endpoint 가 `users.role` 직접 write 였던 동작은 결정 C 와 충돌 (event listener 가 곧 덮어쓰기).

호출자는 `GET /api/v1/users/{user_id}` (organization-management 도메인 API-33) 의 응답에서 `role` 필드 확인.

## 8. 라우트 → (resource, action) 매핑 표 (API-38)

`requirePermission` 미들웨어 (PR-G5) 가 본 표를 source-of-truth 로 enforcement 한다. 표에 *없는* 보호 라우트는 §9 의 deny-by-default 정책에 따라 거부된다.

### 8.1 매핑이 *불필요* 한 라우트 (별도 정책)

| method/path | 정책 |
| --- | --- |
| `GET /health` | public, 인증 미부착 |
| `POST /api/v1/integrations/gitea/webhooks` | HMAC 시그니처 검증 (M0 SEC-2 화이트리스트), RBAC 미적용 |
| `GET /api/v1/me` | 인증된 모든 사용자 (자기 정보) |
| `GET /api/v1/realtime/ws` | 인증된 모든 사용자, 권한 필터링은 메시지 수준 (M3 publish 분류 의존) |

### 8.2 (resource, action) 매핑

| method/path | resource | action |
| --- | --- | --- |
| `GET /api/v1/dashboard/metrics` | infrastructure | view |
| `GET /api/v1/events` | infrastructure | view |
| `GET /api/v1/infra/edges` | infrastructure | view |
| `GET /api/v1/infra/nodes` | infrastructure | view |
| `GET /api/v1/infra/topology` | infrastructure | view |
| `GET /api/v1/repositories` | pipelines | view |
| `GET /api/v1/issues` | pipelines | view |
| `GET /api/v1/pull-requests` | pipelines | view |
| `GET /api/v1/ci-runs` | pipelines | view |
| `GET /api/v1/ci-runs/:ci_run_id/logs` | pipelines | view |
| `GET /api/v1/risks` | security | view |
| `GET /api/v1/risks/critical` | security | view |
| `POST /api/v1/risks/:risk_id/mitigations` | security | create |
| `GET /api/v1/audit-logs` | audit | view |
| `GET /api/v1/rbac/policy` *(legacy)* | security | view |
| `GET /api/v1/rbac/policies` | security | view |
| `POST /api/v1/rbac/policies` | security | edit |
| `PUT /api/v1/rbac/policies` | security | edit |
| `DELETE /api/v1/rbac/policies/:role_id` | security | edit |
| ~~`GET /api/v1/rbac/subjects/:subject_id/roles`~~ | ~~organization~~ | ~~view~~ — **폐기 (ADR-0020, sprint -d)** |
| ~~`PUT /api/v1/rbac/subjects/:subject_id/roles`~~ | ~~organization~~ | ~~edit~~ — **폐기 (ADR-0020, sprint -d)** |
| `POST /api/v1/admin/service-actions` | infrastructure | create |
| `GET /api/v1/commands/:command_id` | infrastructure | view |
| `GET /api/v1/users` | organization | view |
| `POST /api/v1/users` | organization | create |
| `GET /api/v1/users/:user_id` | organization | view |
| `PATCH /api/v1/users/:user_id` | organization | edit |
| `DELETE /api/v1/users/:user_id` | organization | delete |
| `GET /api/v1/organization/hierarchy` | organization | view |
| `GET /api/v1/organization/units/:unit_id` | organization | view |
| `POST /api/v1/organization/units` | organization | create |
| `PATCH /api/v1/organization/units/:unit_id` | organization | edit |
| `DELETE /api/v1/organization/units/:unit_id` | organization | delete |
| `GET /api/v1/organization/units/:unit_id/members` | organization | view |
| `PUT /api/v1/organization/units/:unit_id/members` | organization | edit |
| `GET /api/v1/dev-requests` | dev_requests | view |
| `POST /api/v1/dev-requests` | dev_requests | (intake token auth) |
| `GET /api/v1/dev-requests/:id` | dev_requests | view |
| `POST /api/v1/dev-requests/:id/register` | dev_requests | edit |
| `POST /api/v1/dev-requests/:id/reject` | dev_requests | edit |
| `PATCH /api/v1/dev-requests/:id` | dev_requests | edit |
| `DELETE /api/v1/dev-requests/:id` | dev_requests | delete |
| `POST /api/v1/dev-request-tokens` | dev_request_intake_tokens | create |
| `GET /api/v1/dev-request-tokens` | dev_request_intake_tokens | view |
| `DELETE /api/v1/dev-request-tokens/:token_id` | dev_request_intake_tokens | delete |
| `PATCH /api/v1/dev-request-tokens/:token_id` | dev_request_intake_tokens | edit |

> 신규 v1 라우트가 추가되면 본 표에 행 추가가 *필수*. 누락 시 §9 deny-by-default 가 발동해 모든 사용자 거부 + audit 알림.

## 9. 매핑 누락 정책 (deny-by-default) (API-39)

`requirePermission` 미들웨어가 라우트 처리 시점에 (resource, action) 매핑을 찾지 못하면:

1. 응답: `403 Forbidden` + `{"status":"forbidden","error":"route is not mapped to an RBAC permission"}`.
2. audit: `auth.policy_unmapped`, target_type=`route`, target_id=`c.FullPath()`, payload=`{"actor_role": "...", "method": "..."}`.
3. 운영 알림: 본 audit action 은 별도 monitoring (sprint M3 publish 확장 대상) 으로 *모든 발생을 즉시 인지* 가능하게 한다.

본 정책은 라우트 추가 시점의 매핑 표 갱신 누락을 *런타임 거부* 로 강제하기 위한 안전장치다.

## 9.1 표준 거부 코드

- `auth.policy_unmapped`: 라우트 매핑 누락 거부
- `auth.row_denied`: row scope/ownership 거부
- `auth.role_sync_required`: role drift 감지로 fail-closed 거부

## 9.2 Read scope 계약 (연계 정책)

- `applications:view`, `projects:view` 는 route-level 허용만으로 충분하지 않으며 read scope가 결합되어야 한다.
- `List*` 계열: scope 밖 데이터는 필터링(빈 목록 허용)
- `Get*` 계열: scope 밖 데이터는 `403` + `code=auth_row_denied`
- 상세 scope 병합 규칙은 `docs/planning/role-access-concept.md` 및 `requirements.md`(REQ-RBAC-013/015) 기준.

## 10. Cache 와 무효화 (API-40)

- store 적중 비용 회피를 위해 `requirePermission` 은 in-memory matrix cache (per process) 를 유지한다.
- `PUT/POST/DELETE /api/v1/rbac/policies` 머지 시 동일 프로세스 내 cache reload. (`PUT /api/v1/rbac/subjects/.../roles` 는 ADR-0020 결정 D 로 sprint -d 폐기 — cache reload trigger 도 자연 제거.)
- 다중 인스턴스 환경의 cache 일관성은 미해결 — 운영 phase 진입 시 pub/sub 또는 polling 으로 보강.

## 10.1 Sign-out 연계 (cross-domain)

- FE `/devhub/auth/signout` 는 `POST /api/v1/auth/logout` orchestration route 로 동작해야 한다.
- logout의 세션 정리/revoke/audit 계약은 auth-session 도메인 API 문서를 source-of-truth로 하며, RBAC 도메인은 `auth.logout` 감사 추적과 권한 경계 정합을 보장한다.

## 11. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-06-02 | Two-Dimensional RBAC 정합화: 시스템 role을 `developer/team_manager/system_admin`으로 갱신. 표준 거부 코드(`auth.policy_unmapped`, `auth.row_denied`, `auth.role_sync_required`), read scope 결합 계약(List filter/Get 403), FE signout → API logout 연계 정책을 추가. |
| 2026-05-29 | Phase 3 split — master `docs/backend_api_contract.md` §12 (RBAC) 본문 그대로 이관. ID(API-26..31, API-38..40) 보존, 신규 발급/삭제 없음. |
