---
title: api
type: source
tags: [domain, api.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/organization-management/api.md]
git_commit: 6c434887
git_branch: main
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-18T12:08:55Z
mirror_dirty: (dirty: uncommitted changes) |
related: [none]
status: draft
contradictions: [none]
---

# organization-management 도메인 API

- 문서 목적: `users` CRUD + `/api/v1/organization/*` (조직 계층/units/members) API 계약을 정의한다.
- 범위: API-33 (users CRUD — 기본 의미), API-34 (organization endpoint set). onboarding 의 `POST /api/v1/me/onboarding`, `/api/v1/organizations/search`, `POST /api/v1/users` 사전등록 의미 확장은 `docs/domain/onboarding/api.md` 참조. 사용자 master row 의 인증 부분은 `docs/domain/auth-session/api.md` 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master `docs/backend_api_contract.md` §10.3 + §10.4 본문 이관)
- 관련 문서: [도메인 README](./README.md), [requirements.md](./requirements.md), [architecture.md](./architecture.md), [master API](../../backend_api_contract.md), [기존 organizational_hierarchy_spec](./organizational_hierarchy_spec.md), [기존 org_chart_ux_spec](./org_chart_ux_spec.md), [ADR-0008](../../adr/0008-organization-model.md)

## 개요

조직 master data (users + org_units + members) 의 endpoint set.

## 1. `/api/v1/users` CRUD (API-33)

조직 master data 의 user 차원 endpoint set.

| method/path | 용도 | audit action |
| --- | --- | --- |
| `GET /api/v1/users` | 사용자 목록 조회 (query: `unit_id`, `role`, `status`, `q`) | (조회, audit 미작성) |
| `POST /api/v1/users` | DevHub user master data 생성 (identity lifecycle 은 외부 IdP 운영 경로 담당) | `user.created` |
| `GET /api/v1/users/{user_id}` | 개별 사용자 조회 | (조회, audit 미작성) |
| `PATCH /api/v1/users/{user_id}` | user 정보 수정 (display_name, role, status, idp_subject 등) | `user.updated` |
| `DELETE /api/v1/users/{user_id}` | DevHub user soft-delete (identity lifecycle 은 외부 IdP 운영 경로 담당) | `user.deleted` |

> `POST /api/v1/users` 의 onboarding 사전등록 의미는 `docs/domain/onboarding/api.md` §6 참조.

### 1.1 권한

- `organization:view` (GET), `organization:create` (POST), `organization:edit` (PATCH), `organization:delete` (DELETE).

### 1.2 응답 envelope

```json
{
  "status": "ok",
  "data": { "user_id": "...", "email": "...", "display_name": "...", "role": "...", "status": "active", "system_id": "...", "idp_subject": "..." },
  "meta": { "audit_log_id": "..." }
}
```

`audit_log_id` 는 mutation endpoint (POST/PATCH/DELETE) 응답에만 포함된다.

## 2. `/api/v1/organization/*` (API-34)

조직 계층 (hierarchy + units + unit members).

| method/path | 용도 | audit action |
| --- | --- | --- |
| `GET /api/v1/organization/hierarchy` | 전체 조직 계층 트리 조회 (parent-child 그래프) | (조회) |
| `PUT /api/v1/organization/hierarchy` | hierarchy bulk replace (parent/order 일괄 갱신) | `organization.hierarchy_updated` |
| `POST /api/v1/organization/units` | 새 organizational unit (부서) 생성 | `org_unit.created` |
| `GET /api/v1/organization/units/{unit_id}` | unit 단건 조회 (members 포함) | (조회) |
| `PATCH /api/v1/organization/units/{unit_id}` | unit 정보 수정 (name, type, parent, leader_id 등) | `org_unit.updated` |
| `DELETE /api/v1/organization/units/{unit_id}` | unit 삭제 (cascade 또는 422 — 데이터 정합성 정책은 `./organizational_hierarchy_spec.md` 참조) | `org_unit.deleted` |
| `GET /api/v1/organization/units/{unit_id}/members` | unit 의 members 조회 | (조회) |
| `PUT /api/v1/organization/units/{unit_id}/members` | unit members bulk replace | `org_unit.members_replaced` |

### 2.1 권한

- `organization:view` (GET), `organization:create` (POST units), `organization:edit` (PUT/PATCH), `organization:delete` (DELETE).

### 2.2 응답 envelope

```json
{
  "status": "ok",
  "data": { "unit_id": "...", "name": "...", "type": "team", "parent_id": "...", "leader_id": "...", "members": [...] },
  "meta": { "audit_log_id": "..." }
}
```

> 자세한 schema 는 [`./organizational_hierarchy_spec.md`](./organizational_hierarchy_spec.md) + [`./org_chart_ux_spec.md`](./org_chart_ux_spec.md) 참조. 본 절은 rbac-permissions API §8.2 의 RBAC enforcement 매핑과 cross-link 1차. 하위 mutation endpoint 의 1차 schema 보강은 §2.3~§2.6 (sprint `claude/work_260513-l`, RM-M3-03).

### 2.3 `POST /api/v1/organization/units`

새 organizational unit 을 생성한다.

요청 body:

```json
{
  "unit_id": "team-frontend",
  "parent_unit_id": "dept-engineering",
  "unit_type": "team",
  "label": "Frontend Team",
  "leader_user_id": "yklee",
  "position_x": 200,
  "position_y": 120
}
```

- `unit_id`: 필수, 신규 식별자. 충돌 시 409.
- `parent_unit_id`: optional. root 단위는 빈 문자열.
- `unit_type`: 자유 텍스트 (`team`, `department`, `org` 등). 정합 규칙은 후속 spec.
- `leader_user_id`: optional. 비어 있으면 leader 미배정.
- `position_x`, `position_y`: 조직도 캔버스 좌표. UI drag 위치 영속화에 사용.

응답 (`201 Created`):

```json
{ "status": "created", "data": { /* orgUnitResponse */ }, "meta": { "audit_log_id": "..." } }
```

에러 매트릭스:

| status | code | 의미 |
| --- | --- | --- |
| 400 | (없음) | body parse 실패, `unit_id` 누락 |
| 400 | `x_devhub_actor_removed` | inbound `X-Devhub-Actor` 헤더 ([ADR-0006](../../adr/0006-x-devhub-actor-reject-inbound.md)) |
| 401 | `unauthenticated` | Bearer token 부재 / 실패 |
| 403 | `forbidden` | RBAC `organization:create` 미보유 |
| 404 | `parent_not_found` | `parent_unit_id` 가 존재하지 않음 |
| 409 | `conflict` | `unit_id` 중복 |
| 500 | (없음) | 저장 실패 |

### 2.4 `PATCH /api/v1/organization/units/{unit_id}`

unit 의 일부 필드를 갱신. 모든 필드 optional (pointer wire) — 명시된 필드만 변경.

요청 body:

```json
{
  "parent_unit_id": "dept-platform",
  "label": "Frontend Platform Team",
  "leader_user_id": "akim"
}
```

응답 (`200 OK`):

```json
{ "status": "ok", "data": { /* updated orgUnitResponse */ }, "meta": { "audit_log_id": "..." } }
```

에러 매트릭스:

| status | code | 의미 |
| --- | --- | --- |
| 400 | (없음) | body parse 실패 |
| 401 / 403 / 400 (`x_devhub_actor_removed`) | (auth/RBAC 공통) | (위와 동일) |
| 404 | `not_found` | `unit_id` 가 존재하지 않음 |
| 409 | `cycle_detected` | `parent_unit_id` 변경이 cycle 을 유발 (구현은 carve out, 본 spec 은 의도 기록) |

> `parent_unit_id` 변경 시 cycle 방지 검증은 carve out — [`./organizational_hierarchy_spec.md`](./organizational_hierarchy_spec.md) §3 의 결정 항목. 본 sprint 는 spec 의도만 노출.

### 2.5 `DELETE /api/v1/organization/units/{unit_id}`

unit 삭제. cascade 정책 (자식 unit / member 처리) 은 [`./organizational_hierarchy_spec.md`](./organizational_hierarchy_spec.md) 참조.

응답 (`200 OK`):

```json
{ "status": "deleted", "data": { "unit_id": "team-frontend" }, "meta": { "audit_log_id": "..." } }
```

에러 매트릭스:

| status | code | 의미 |
| --- | --- | --- |
| 401 / 403 / 400 (`x_devhub_actor_removed`) | (auth/RBAC 공통) | |
| 404 | `not_found` | `unit_id` 가 존재하지 않음 |
| 422 | `has_children` | 자식 unit 이 존재 (cascade 미지원 정책 시) |
| 422 | `has_members` | members 가 비어 있지 않음 (cascade 미지원 정책 시) |

### 2.6 `PUT /api/v1/organization/units/{unit_id}/members`

unit 의 member 목록을 bulk replace. 누락된 user 는 unit 에서 제거, 신규 user 는 추가.

요청 body:

```json
{
  "user_ids": ["yklee", "akim", "sjones"]
}
```

응답 (`200 OK`):

```json
{
  "status": "ok",
  "data": {
    "unit_id": "team-frontend",
    "members": [ /* user array */ ]
  },
  "meta": { "audit_log_id": "..." }
}
```

에러 매트릭스:

| status | code | 의미 |
| --- | --- | --- |
| 400 | (없음) | body parse 실패 |
| 401 / 403 / 400 (`x_devhub_actor_removed`) | (auth/RBAC 공통) | |
| 404 | `not_found` | `unit_id` 가 존재하지 않음 |
| 422 | `unknown_user_ids` | `user_ids` 중 DevHub `users` 에 없는 항목 — 응답 detail 에 목록 포함 |

> primary_dept 자동 판정 (겸임 우선순위, 동급 시 자식 노드 수) 은 본 endpoint 의 후속 결정 — [`./backend_requirements_org_hierarchy.md`](./backend_requirements_org_hierarchy.md) §1·2 의 미해결 항목. 본 sprint 는 spec 의도만 노출.

## 3. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/backend_api_contract.md` §10.3 (users CRUD) + §10.4 (organization endpoint set) 본문 그대로 이관. ID(API-33, API-34) 보존, 신규 발급/삭제 없음. onboarding-specific 의미 확장(API-33 사전등록)은 onboarding api 로 분리. |
