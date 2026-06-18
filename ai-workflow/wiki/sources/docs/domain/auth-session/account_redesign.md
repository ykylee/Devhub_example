---
title: account_redesign
type: source
tags: [domain, account_redesign.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/auth-session/account_redesign.md]
git_commit: 01f1969c
git_branch: main
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-18T07:11:15Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# 계정/사용자 관리 리팩토링 — 외부 Keycloak 가정 (redesign)

- 문서 목적: 외부 Keycloak 을 단일 IdP 로 사용한다는 전제로, DevHub 의 **계정 관리** + **사용자 관리** 의 책임 경계를 재정의하고 리팩토링한다. 본 문서 §1~§4 는 **Phase 1 — 현황 파악**, §5+ 는 후속 sprint 의 **Phase 2 — 책임 분리 design** + **Phase 3 — 리팩토링 실행 계획**.
- 범위: backend `/api/v1/accounts/*` + `/api/v1/users/*` + `/api/v1/organization/*` + `/api/v1/rbac/*` + `/api/v1/me` endpoint, frontend `/account` + `/admin/settings/{users,organization,permissions}` page, DB schema (users + organization_units + rbac_policies + rbac_subject_roles), Keycloak realm 의 user/group/role + token claim 정합.
- 대상 독자: 백엔드/프론트엔드/IdP 담당, 운영 (SRE), 보안.
- 상태: draft
- 최종 수정일: 2026-05-20
- 결정 근거 sprint: `claude/work_260520-a` (Phase 1 현황 파악 + Phase 2 책임 분리 design)
- 관련 문서: [ADR-0019 Keycloak 단일화](../adr/0019-keycloak-only-idp.md), [keycloak_operations.md (§8.5b self-service 비밀번호 변경 Account Console 위임)](../setup/keycloak_operations.md), [ADR-0011 RBAC row-scoping](../adr/0011-rbac-row-scoping.md), [keycloak_groups_rbac_mapping.md (group composite role)](./keycloak_groups_rbac_mapping.md), [keycloak_offboarding_immediacy.md (off-boarding chain)](./keycloak_offboarding_immediacy.md), [ADR-0008 HRDB production adapter](../adr/0008-hrdb-production-adapter.md), [traceability/report.md (REQ/API/IMPL 매트릭스)](../traceability/report.md).

## 0. 배경 + 문제 정의

DevHub 는 [ADR-0019](../adr/0019-keycloak-only-idp.md) 채택으로 Keycloak 을 **단일 IdP** 로 운영한다. 그러나 **계정 관리** + **사용자 관리** 의 코드 + UI + DB 책임이 다음 4 경로로 분산되어 있고, 외부 Keycloak (사내 IdP 팀이 별도 운영) 시나리오에서 책임이 모호하다.

1. **DevHub backend `/api/v1/accounts/*`** — Keycloak Admin REST 의 proxy. service account 로 Keycloak Admin API 호출
2. **DevHub backend `/api/v1/users/*`** — 조직 메타데이터 (status / role / unit assignment) CRUD
3. **DevHub backend `/api/v1/organization/*`** — 조직 단위 (units / hierarchy)
4. **DevHub backend `/api/v1/rbac/*`** — RBAC policy + subject-role assignment

본 문서의 Phase 1 (§1~§4) 은 위 4 경로의 현황을 매트릭스로 정리한다. **Phase 2 (별도 sprint)** 에서 외부 Keycloak 시나리오의 책임 재정의 옵션 (A~D 후보) + 결정 + ADR 발급을 진행한다.

## 1. 책임 분리 매트릭스 (현재 상태)

| 영역 | source-of-truth | DevHub 측 표상 | sync 방향 / 시점 |
| --- | --- | --- | --- |
| **인증 credentials** (password, MFA, session) | Keycloak | (없음 — DevHub 가 보관하지 않음) | OIDC token 발급 시 (Keycloak → DevHub OIDC discovery) |
| **identity_id (Keycloak user UUID)** | Keycloak `sub` claim | `users.idp_subject` (UNIQUE) | token 검증 시 eager / lazy backfill (sprint -j PR #188 자동 sync) |
| **username (login id)** | Keycloak `preferred_username` claim | `users.user_id` (UNIQUE) | account.create 또는 token 검증 시 |
| **email** | Keycloak `email` claim | `users.email` | account.create 시점 1회. 변경은 Keycloak admin 측 → DevHub 미동기화 (carve) |
| **display name** | Keycloak `name` claim | `users.display_name` | 동일 — account.create 시 1회 |
| **employee_id** (HRDB 매핑) | Keycloak custom attribute `employee_id` | (claim 으로만 사용, DevHub DB 미저장) | Keycloak admin 측 user attribute 입력 (HRDB ETL push, [§8.2 keycloak_operations.md](../setup/keycloak_operations.md)) |
| **계정 status** (active/disabled) | Keycloak `enabled` boolean | `users.status` (active/pending/deactivated) | DevHub `/api/v1/accounts/:user_id` PATCH → Keycloak SetIdentityState + DevHub users.status update (atomic, rollback 가능) |
| **role** (developer/manager/team_manager/system_admin) | **이중 source** — Keycloak `realm_access.roles` (group composite) + DevHub `users.role` (cache) | `users.role` | token 검증 시 `realm_access.roles` 가 권한 wire format 우선 (`keycloak_verifier.go:260`). `users.role` 은 화면 표시 + audit 용 cache (sprint -j PR #185 multi-role priority filter) |
| **조직 단위 (unit)** | DevHub | `organization_units` + `users.{primary_unit_id, current_unit_id}` | DevHub 자체 (Keycloak 무관) |
| **조직 계층 (hierarchy)** | DevHub | `organization_hierarchy` (parent_unit_id) | DevHub 자체 |
| **RBAC policy** (Resource × Action) | DevHub | `rbac_policies` | DevHub 자체 (`/api/v1/rbac/policies`) |
| **RBAC subject-role assignment** | Keycloak group composite (간접) | `users.role` 컬럼 직접 write (별도 테이블 없음 — Phase 1 §1 매트릭스 작성 시 `rbac_subject_roles` 표기 오류, [§5.9 참조](#59-phase-1-매트릭스-오류-정정)). endpoint 자체 결정 D 로 완전 제거 | endpoint 정의됐었으나 frontend UI 미구현. 결정 D 후 코드 자체 폐기 |
| **audit log** | DevHub (인입 source 다양) | `audit_logs` | source_type=oidc/webhook/kratos(deprecated)/system/keycloak_event. Keycloak admin event 는 §5.3 (9) listener (sprint -u~-y) 가 polling-emit |

### 1.1 source-of-truth 이중화 issue (current)

- **role 이중 source** — Keycloak group composite 가 backend 권한 평가 의 1차 source, `users.role` 는 화면 표시 cache. 두 값이 divergent (예: Keycloak 에서 group 이동 후 DevHub `users.role` 미갱신) 하면 화면에 옛 role 표시 + 실 권한은 새 role. **현재 명시 sync mechanism 없음** — 본 sprint 의 발견 사항 #1.
- **status 이중 source** — Keycloak `enabled` 와 DevHub `users.status` 둘 다 보유. DevHub `/api/v1/accounts/:user_id` PATCH 가 atomic update (SetIdentityState + UpdateUser, rollback 보호). 그러나 Keycloak admin console 직접 변경 (DevHub 우회) 시 DevHub `users.status` 가 stale.
- **identity_id (idp_subject) lazy backfill** — sprint -j 의 자동 sync 가 `authenticateActor` 에서 처리. 첫 로그인 시 DB write 발생 (perf 영향 microscopic).

## 2. Backend endpoint 매트릭스 (23 endpoint, 18 row)

> 표의 마지막 두 row 는 묶음 — `rbac/policies` 는 4 endpoint (GET/POST/PUT/DELETE), `rbac/subjects/:subject_id/roles` 는 2 endpoint (GET/PUT).

| Method | Endpoint | Handler | DB 작용 (R/W) | Keycloak 작용 | Audit action | RBAC 권한 |
| --- | --- | --- | --- | --- | --- | --- |
| POST | `/api/v1/accounts` | createAccount | users(W), users.idp_subject(W) | CreateIdentity | account.issued + account.issue.kratos_failed (error) | security:create |
| PUT | `/api/v1/accounts/:user_id/password` | resetAccountPassword | users.idp_subject(R) | UpdateIdentityPassword | account.password_force_reset | security:edit |
| PATCH | `/api/v1/accounts/:user_id` | updateAccountStatus | users.status(W) + users.idp_subject(R) | SetIdentityState (atomic + rollback) | account.enabled / account.disabled | organization:edit |
| DELETE | `/api/v1/accounts/:user_id` | deleteAccount | users(D), users.idp_subject(R) | DeleteIdentity (non-fatal if missing) | account.deleted + account.delete.kratos (error) | organization:delete |
| GET | `/api/v1/users` | listUsers | users(R), appointments(R) | — | — | organization:view |
| POST | `/api/v1/users` | createUser | users(W), users.idp_subject(W if password) | CreateIdentity (optional, password 있을 때만) | user.created | organization:create |
| GET | `/api/v1/users/:user_id` | getUser | users(R), appointments(R) | — | — | organization:view |
| PATCH | `/api/v1/users/:user_id` | updateUser | users(W) | — | user.updated | organization:edit |
| DELETE | `/api/v1/users/:user_id` | deleteUser | users(D), appointments(D) | — | user.deleted | organization:delete |
| GET | `/api/v1/me` | getMe | (none — request context only) | — | — | Bypass (auth-only) |
| GET | `/api/v1/organization/hierarchy` | getHierarchy | organization_units(R), organization_hierarchy(R) | — | — | organization:view |
| PUT | `/api/v1/organization/hierarchy` | updateHierarchy | organization_units(U), organization_hierarchy(U) | — | org_unit.hierarchy_updated | organization:edit |
| POST | `/api/v1/organization/units` | createOrgUnit | organization_units(W) | — | org_unit.created | organization:create |
| GET | `/api/v1/organization/units/:unit_id` | getOrgUnit | organization_units(R), users(R) | — | — | organization:view |
| PATCH | `/api/v1/organization/units/:unit_id` | updateOrgUnit | organization_units(W) + hierarchy 순환 검증 | — | org_unit.updated | organization:edit |
| DELETE | `/api/v1/organization/units/:unit_id` | deleteOrgUnit | organization_units(D) | — | org_unit.deleted | organization:delete |
| PUT | `/api/v1/organization/units/:unit_id/members` | replaceUnitMembers | users(W: primary_unit_id), appointments(W) | — | org_unit.members_replaced | organization:edit |
| GET/POST/PUT/DELETE | `/api/v1/rbac/policies` (4 endpoint) | listRBACPolicies / createRBACPolicy / updateRBACPolicies / deleteRBACPolicy | rbac_policies(R/W/D) | — | rbac.policy.{created,updated,deleted} | security:edit |
| GET/PUT | `/api/v1/rbac/subjects/:subject_id/roles` (2 endpoint) | getSubjectRoles / setSubjectRoles | rbac_subject_roles(R/W) | — | rbac.role.assigned | organization:edit |

### 2.1 `IdentityAdmin` interface (router.go) → Keycloak Admin REST 매핑

| 메서드 | Keycloak Admin REST endpoint | 용도 |
| --- | --- | --- |
| `CreateIdentity(ctx, email, name, userID, password)` | `POST /admin/realms/{realm}/users` + credentials sub-resource | 신규 Keycloak user + 임시 password (temporary=true) |
| `FindIdentityByUserID(ctx, userID)` | `GET /admin/realms/{realm}/users?username={userID}` | DevHub user_id 로 Keycloak UUID lookup (lazy backfill 경로) |
| `UpdateIdentityPassword(ctx, identityID, password)` | `PUT /admin/realms/{realm}/users/{id}/reset-password` | 비밀번호 강제 reset (admin 발급 temp password) |
| `SetIdentityState(ctx, identityID, active)` | `PUT /admin/realms/{realm}/users/{id}` (`enabled: bool`) | 계정 활성/비활성 |
| `DeleteIdentity(ctx, identityID)` | `DELETE /admin/realms/{realm}/users/{id}` | Keycloak user 삭제 |

`KeycloakAdminClient` (backend-core/internal/httpapi/keycloak_admin_client.go) 가 위 interface 구현. service account (`devhub-backend` client_credentials grant) 가 `realm-management.{view-users, manage-users, view-events}` role 보유 필요 (keycloak_operations.md §3.2).

### 2.2 endpoint 중복 / 모호 (current state 발견)

1. **POST `/api/v1/accounts` vs POST `/api/v1/users`** — 둘 다 신규 user 생성. 전자는 Keycloak IdP 정공법 (CreateIdentity 필수), 후자는 password 옵션 (있으면 IdP 생성, 없으면 DevHub-only). **atomic 단일 작업이 아님** — accounts 로 IdP 만들고 별도 PATCH users 로 unit assignment 가 필요.
2. **DELETE `/api/v1/accounts/:id` vs DELETE `/api/v1/users/:id`** — 전자는 Keycloak DeleteIdentity + DevHub user 삭제, 후자는 DevHub user 만 삭제 (Keycloak 잔존). 운영자가 어느 걸 써야 할지 명세 없음.
3. **`PUT /api/v1/rbac/subjects/:subject_id/roles`** — endpoint 정의됐으나 frontend UI 미구현 (dead-end from frontend). 실 권한 평가는 token `realm_access.roles` 우선이므로 본 endpoint 의 의미가 약함.

## 3. Frontend UI + service 매트릭스

### 3.1 Page 매트릭스 (4 page)

| Page 경로 | RBAC gate | 호출 endpoint | CRUD 작용 | UX flow |
| --- | --- | --- | --- | --- |
| `/account` | 모든 인증 user | `/api/runtime-config` (read-only) | (self-service only) | **Keycloak Account Console redirect** (`${issuer}/account/`, sprint -ad) |
| `/admin/settings/users` | system_admin | `GET /api/v1/users`, `POST /api/v1/users`, `PATCH /api/v1/users/:id` | List + Create + Update(role) | FilterBar + MemberTable + Modal |
| `/admin/settings/organization` | system_admin | `GET/PUT /api/v1/organization/hierarchy` + `GET/POST/PATCH/DELETE /api/v1/organization/units/:unit_id` + `PUT /api/v1/organization/units/:unit_id/members` | List + Create + Update + Delete(unit) | 3-view (list/grid/chart) + Modal |
| `/admin/settings/permissions` | system_admin | `GET/PUT /api/v1/rbac/policies` + `POST/DELETE /api/v1/rbac/policies/:role_id` | List + Create + Update + Delete(role) | PermissionEditor + matrix toggle |

### 3.2 Service 매트릭스

| Service file | 메서드 | 호출 endpoint | 응답 |
| --- | --- | --- | --- |
| `account.service.ts` | `issueAccount(userId, loginId, forceReset, options?)` | POST `/api/v1/accounts` | `{tempPassword, identityId?}` |
| | `forceResetPassword(userId)` | PUT `/api/v1/accounts/{userId}/password` | `{tempPassword}` |
| | `disableAccount(userId, reason)` | PATCH `/api/v1/accounts/{userId}` (status: disabled) | `void` |
| | `unlockAccount(userId)` | PATCH `/api/v1/accounts/{userId}` (status: active) | `void` (**UI 미사용**) |
| | `deleteAccount(userId)` | DELETE `/api/v1/accounts/{userId}` | `void` (**UI 미사용**) |
| `auth.service.ts` | `getAccountConsoleURL()` | `/api/runtime-config` | `string` (`${issuer}/account/`, sprint -ad) |
| `identity.service.ts` | `getUsers()` / `createUser(...)` / `updateUser(...)` | `GET/POST/PATCH /api/v1/users[/:id]` | `OrgMember[]` |
| | `getOrgHierarchy()` / `updateOrgHierarchy(...)` | `GET/PUT /api/v1/organization/hierarchy` | `{nodes: OrgNode[], edges: OrgEdge[]}` |
| | `createUnit / updateUnit / deleteUnit / replaceUnitMembers` | `/api/v1/organization/units[/:id]` | `OrgUnit \| OrgMember[]` |
| `rbac.service.ts` | `listPolicies / createPolicy / updatePolicies / deletePolicy` | `/api/v1/rbac/policies[/:role_id]` | `{roles: Role[], meta}` |

### 3.3 role-routing 규칙 (`lib/auth/role-routing.ts`)

| Role | 접근 가능 경로 | 불가 경로 |
| --- | --- | --- |
| **system_admin** | `/admin/*`, `/account`, `/organization` (deprecated → redirect) | `/manager`, `/developer` |
| **manager** | `/account`, `/manager` | `/admin/*`, `/organization` |
| **developer** | `/account`, `/developer` | `/admin/*`, `/organization`, `/manager` |

- `pathRequiresSystemAdmin()` : `/admin*` + `/organization*` → system_admin 전용
- `defaultLandingFor()` : role 별 기본 landing (admin → `/admin`, manager → `/manager`, dev → `/developer`)
- AuthGuard (layout): 진입 시 `whoAmI()` (`GET /api/v1/me`) → role 검증 → 비허가 경로 접근 시 default landing 으로 redirect
- AdminSettingsLayout: defense-in-depth (AuthGuard 재검증)

### 3.4 Store / Actor 구조

| State | Type | 역할 |
| --- | --- | --- |
| `actor` | `AuthenticatedActor` (login, subject, role, source) | 인증된 사용자 정보 |
| `role` | `UserRole \| null` | 현재 role (setActor 에서 자동 sync) |

`actor.subject` = Keycloak `sub` claim (= `users.idp_subject`). `actor.login` = `users.user_id` (= Keycloak `preferred_username`).

### 3.5 frontend 발견 사항

1. **account.service.ts 의 `unlockAccount` + `deleteAccount` 메서드 — UI 미사용**. MemberTable.tsx 에서 `issueAccount / forceResetPassword / disableAccount` 3건만 호출. backend 는 준비됐으나 frontend dead.
2. **`PUT /api/v1/rbac/subjects/:subject_id/roles` UI 미구현** — subject-role assignment 가 backend-only. Permission editor 는 role × resource matrix 만.
3. **organization 페이지 deprecated** — 구 `/organization` → 신 `/admin/settings/{users,organization,permissions}` redirect (PR-S2 시점).
4. **AuthGuard 가 `whoAmI()` 동기 의존** — 매 page 진입 시 1회 호출. `actor` 가 store 에 cache 되지만 SPA refresh 시 매번 호출.
5. **account console redirect 외 self-service 흐름 없음** — MFA enrollment / profile 수정 / 활성 세션 관리 모두 Keycloak Account Console 위임.

## 4. DB schema + Keycloak attribute 정합 매트릭스

### 4.1 `users` 테이블 (migration 000004 base + 후속)

| 컬럼 | 타입 | domain.AppUser | Keycloak 대응 | 비고 |
| --- | --- | --- | --- | --- |
| `id` | BIGSERIAL PK | `ID` | (없음) | 내부 DB ID |
| `user_id` | TEXT UNIQUE | `UserID` | `preferred_username` claim | DevHub primary identifier |
| `email` | TEXT UNIQUE | `Email` | `email` claim | account.create 시 1회 sync |
| `display_name` | TEXT | `DisplayName` | `name` claim (firstName + lastName) | 동일 |
| `role` | TEXT (CHECK developer/manager/system_admin/team_manager) | `Role` | `realm_access.roles` (group composite 매핑) | **cache** — backend 권한 평가는 token claim 우선 |
| `status` | TEXT (CHECK active/pending/deactivated) | `Status` | `enabled` boolean | atomic update + rollback (`/api/v1/accounts/:id` PATCH) |
| `user_type` | TEXT (CHECK human/system) | `Type` | (없음 — DevHub-only) | migration 000007 추가 |
| `idp_subject` | TEXT UNIQUE | `IdPSubject` | `sub` claim (Keycloak user UUID) | migration 000009 추가 (`kratos_identity_id`), 000030 rename |
| `primary_unit_id` | TEXT FK organization_units(unit_id) | `PrimaryUnitID` | (없음 — DevHub-only) | 조직 배치 |
| `current_unit_id` | TEXT FK organization_units(unit_id) | `CurrentUnitID` | (없음 — DevHub-only) | 파견 (is_seconded=true) 시 |
| `is_seconded` | BOOLEAN | `IsSeconded` | (없음 — DevHub-only) | 파견 여부 |
| `joined_at` | DATE | `JoinedAt` | (없음 — DevHub-only) | 입사일 |
| `created_at` / `updated_at` | TIMESTAMPTZ | `CreatedAt` / `UpdatedAt` | (없음) | DB 메타 |

### 4.2 `organization_units` + `organization_hierarchy` (DevHub 자체)

| 테이블 | 핵심 컬럼 | 역할 |
| --- | --- | --- |
| `organization_units` | `unit_id` (TEXT PK), `unit_type` (department/team), `display_name`, `leader_user_id` (FK users.user_id) | 조직 단위 |
| `organization_hierarchy` | `parent_unit_id` + `child_unit_id` (FK organization_units) | DAG 계층 |
| `appointments` | `user_id` + `unit_id` + `appointment_type` (primary/seconded) | user ↔ unit 다대다 관계 |

Keycloak 무관. DevHub 자체 도메인.

### 4.3 `rbac_policies` + (~~`rbac_subject_roles`~~)

| 테이블 | 핵심 컬럼 | 역할 |
| --- | --- | --- |
| `rbac_policies` | `role_id` (TEXT), `resource` + `action`, `allow` | role × resource × action 매트릭스. system_admin / developer / manager / team_manager seed (migration 000004 + 000005) |
| ~~`rbac_subject_roles`~~ | — | **Phase 1 매트릭스 작성 시 표기 오류** — 별도 테이블 없음. backend `GetSubjectRoles` / `SetSubjectRole` 은 `users.role` 컬럼 직접 read/write ([§5.9 정정](#59-phase-1-매트릭스-오류-정정), [결정 D](#51-명시-결정-6건-종합)) |

### 4.4 Keycloak realm 자산 (DevHub 외부)

| 자산 | DevHub 매핑 | source-of-truth |
| --- | --- | --- |
| Keycloak `users` (user UUID + username + email + enabled + credentials + groups + attributes) | `users.idp_subject` + `users.user_id` + `users.email` + `users.status` | Keycloak |
| Keycloak `groups` (devhub-developers/managers/pmo-managers/system-admins) | `users.role` (cache via token `realm_access.roles`) | Keycloak (group composite role assigned) |
| Keycloak custom attribute `employee_id` | HRDB primary key (DevHub DB 미저장, token claim 으로만 사용) | Keycloak admin (HRDB ETL push) |
| Keycloak `realm_access.roles` token claim | backend `keycloak_verifier.go` 의 role extraction | Keycloak (group composite) |

## 4.5 외부 Keycloak 시나리오 — 책임 분리 후보 (Phase 2 입력)

Phase 2 (별도 sprint) 에서 결정할 책임 재정의 후보:

| 옵션 | DevHub backend `/api/v1/accounts/*` | DevHub backend `/api/v1/users/*` | provisioning 흐름 | role/status sync |
| --- | --- | --- | --- | --- |
| **A — DevHub admin endpoint 전면 폐기** | 제거 (404) | 조직 메타데이터만 (PATCH role/unit) | Keycloak admin 직접 + HRDB ETL push + **token 검증 시 lazy users row 자동 생성 (Phase 2 신규 mechanism, 현재 미구현)** [^A-lazy] | Keycloak group composite 가 유일 source. `users.role` 은 cache 만 (token 검증 시마다 sync) |
| **B — 현재 상태 유지 (Admin Client proxy)** | 유지 (DevHub admin UI 운영 편의) | 유지 (조직 메타데이터) | DevHub `/accounts` POST 또는 Keycloak admin 직접 (2 경로 동시) | 이중 source — 본 문서 §1.1 issue 재발 |
| **C — Hybrid (write 일부만)** | 일부 (password reset / disable) 만 유지, create/delete 는 폐기 | 유지 | 생성/삭제 = Keycloak admin 직접, 일시 disable / password reset = DevHub UI 가 편의 | 동일 — 부분 정합 |
| **D — Read-only DevHub admin** | 제거 + read-only mirror (`GET /api/v1/accounts/:id` 만 유지) | 유지 (read+update unit/role cache) | Keycloak admin 직접 + SCIM bridge | Keycloak group composite 유일 source |

**Phase 2 결정 (sprint -a, 2026-05-20): 옵션 A 전면 폐기 확정** — 사용자 조건 (IdP 팀 별도 운영, DevHub 운영자 manage-users 권한 없음, 자주 쓰는 동작은 `/api/v1/users` 계열) 정합. 상세는 [§5.1 명시 결정 6건](#51-명시-결정-6건-종합) + [§5.2 lazy auto-create](#52-lazy-auto-create-mechanism).

[^A-lazy]: 현재 (sprint -j PR #188) 의 자동 sync 는 `authenticateActor` 가 **`users.idp_subject` 컬럼만** lazy backfill (이미 존재하는 row 의 idp_subject 가 비어 있으면 Keycloak FindIdentityByUserID 로 조회 후 세팅). `users` row 자체가 없는 경우의 자동 생성은 **현재 mechanism 없음** — token 검증 시 GetUser miss → ErrIdentityNotFound. 옵션 A 채택 시 신규 mechanism 으로 lazy auto-create 추가 필요.

## 5. Phase 2 책임 분리 design (sprint -a, 2026-05-20 확정)

본 sprint 의 명시 결정 토론 (Q&A 6 round) 결과 — 6건 결정 + Phase 2 동반 작업 6건 + Phase 1 매트릭스 오류 정정 1건.

### 5.1 명시 결정 6건 종합

| # | 영역 | 결정 | 근거 |
| --- | --- | --- | --- |
| **A** | `/api/v1/accounts/*` 4 endpoint 향방 | **전면 폐기** (옵션 A) | 사용자 조건 — IdP 팀 별도 운영, DevHub 운영자 manage-users 권한 없음. 자주 쓰는 동작 (user 목록 view + 조직 unit/role assignment) 은 `/api/v1/users` 계열 (Keycloak 무관). divergence 원천 차단 + service account 권한 축소 + dead UI 메서드 자연 정리 |
| **B** | `/login` page 향방 | **entry minimal page 유지** | DevHub brand 첫인상 + 명시 로그인 step + error message 표시 + `/login` ↔ `/auth/login` 중복 정리 |
| **C** | role/status sync mechanism | **event listener 확장 (sprint -u~-y 자연 확장)** + lazy backfill/auto-create hot path 1회 | (a) token 검증 write-through, (b) stale 비교 모두 token claim 한계 (group_membership / status change 감지 불가). event listener 만 정확. hot path 영향 0, latency 30s 는 access_token 5분 stale 안에 묻힘 |
| **D** | `rbac_subject_roles` (endpoint + store + interface) | **완전 제거** (옵션 a) | 발견: 별도 테이블 없음, `users.role` 컬럼 직접 write. 명시 결정 C 와 충돌 (event listener 가 곧 덮어쓰기). `PATCH /api/v1/users/:id` 와 기능 중복. ADR-0011 row-scoping 과 무관 |
| **E** | read-only 모드 carve (Keycloak down 시 GET grace period) | **도입 안 함** (self-reverse) | signature 검증 skip 은 OIDC 표준 위반 + token forgery 위험 + revoked / audit forgery. 명시 결정 F 가 진짜 정공법 |
| **F** | JWKS stale-while-error 확장 (expiry mismatch) | **확장 도입** (옵션 a) | sprint -r (PR #186) kid mismatch fallback 의 자연 확장. signature 검증 유지 (token forgery 위험 없음) + uptime → key rotation period (90일) 까지. revoked key 시나리오는 별도 mitigation |

### 5.2 lazy auto-create mechanism (결정 A 동반 #1)

신규 user 가 Keycloak admin console 에서 생성된 후 첫 DevHub 진입 시 `users` row 자동 생성.

**현재 상태** (sprint -j PR #188): `authenticateActor` 가 **`users.idp_subject` 컬럼만** lazy backfill — `users` row 자체가 없으면 ErrIdentityNotFound → 401.

**Phase 2 신규**: `users` row 자체가 없을 때 자동 INSERT.

#### 5.2.1 흐름

```text
GET /api/v1/me 요청 (신규 user 의 첫 진입)
  ↓
authenticateActor
  ↓ token 검증 통과 (Keycloak 의 user)
  ↓
GetUser(ctx, userID) → ErrNotFound (DevHub 에 row 없음)
  ↓ (현재) 401 + "DevHub user 등록 안 됨" 안내
  ↓ (Phase 2 신규) lazy auto-create
  ↓
CreateUser(ctx, CreateUserInput{
  UserID:       token.preferred_username,
  Email:        token.email,
  DisplayName:  token.name,
  Role:         extractRole(token.realm_access.roles),  // event listener 와 동일 로직
  Status:       UserStatusActive,
  Type:         UserTypeHuman,
  IdPSubject:   token.sub,
  // PrimaryUnitID, CurrentUnitID 는 default 또는 unassigned 상태로 둠
  JoinedAt:     now(),
})
  ↓ audit: account.lazy_provisioned (action 이름 신규)
  ↓
정상 응답
```

#### 5.2.2 결정 항목

| 항목 | 결정 |
| --- | --- |
| **조직 unit 자동 매핑** | **unassigned** — `primary_unit_id` NULL. admin 이 후속 `/admin/settings/users` 에서 unit 배치. (HRDB ETL push 가 사전에 unit 매핑 정보를 stage 하면 자동 배치 가능 — sprint -p `hrdb_etl_sync.sh` 확장 carve) |
| **role 매핑** | token `realm_access.roles` 추출 — Keycloak group composite (`developer/manager/team_manager/system_admin`). 명시 결정 C 의 event listener 매핑 로직과 동일 함수 공유 (`extractKeycloakRole`, sprint -j PR #185 multi-role priority filter) |
| **role 매핑 fallback** (sprint -b Stage 3 P1-3 결정) | token `realm_access.roles` 가 비어 있거나 매핑 가능한 role 없을 때 → **default `developer`** 부여 + audit `user.role_default_assigned`. backend `keycloak_verifier.go` 의 현재 fallback 동작과 정합. 별도 신규 role state (`unassigned`) 도입 안 함 — `rbac_policies` 의 4 role enum 유지 |
| **status 초기값** | `active` — Keycloak `enabled=true` 인 token 발급 시점 정합 |
| **audit action 이름** | `account.lazy_provisioned` 신규 (DB row 정합 영향 없음, 신규 row 만 emit) |
| **HRDB pre-stage 와의 race** | HRDB ETL push 가 먼저 row 생성한 경우 → lazy auto-create 는 `GetUser` 가 row 발견 → noop. 두 경로 idempotent |

#### 5.2.3 보안 검토

- **token 검증 성공한 user 만 lazy create** — Keycloak 이 발급한 valid token signature 가 통과한 user 만. 즉 Keycloak 의 user lifecycle 정책 (enabled, group, etc) 이 1차 필터
- **enumeration 위협 없음** — DevHub backend 가 자체 user list 노출 안 함. lazy create 자체가 token 인증 필수
- **audit 추적** — `account.lazy_provisioned` row 가 모든 lazy create event 기록 (actor / source_ip / token claim summary)

### 5.3 Keycloak event listener 확장 (결정 A 동반 #2 + 결정 C)

sprint -u~-y 의 audit event listener 가 `audit_logs` 만 emit. Phase 2 확장 — DevHub `users` 컬럼 sync.

#### 5.3.1 매핑 표 (sprint -u~-y §8.6.2 매핑 표 확장)

| Keycloak Admin Event Type | Operation Type | DevHub 작용 | audit_logs action (기존 sprint -u~-y) |
| --- | --- | --- | --- |
| `USER` | `UPDATE` | `users.email` / `users.display_name` / `users.status` (enabled boolean → active/deactivated) sync | `user.profile_updated` |
| `USER` | `CREATE` | (lazy auto-create 정합 — DevHub 의 첫 진입 시점에 handler 가 처리. event 는 audit log 만) | `user.created_external` |
| `USER` | `DELETE` | **`users.status = deactivated` (soft delete 채택, sprint -b Stage 3 P1-2 결정)** — DevHub `users` row 의 historical 정합 보존 (audit_logs 의 actor reference 깨짐 회피). archive 또는 hard delete 는 사내 보존 정책 따른 별도 carve | `user.deleted_external` |
| `GROUP_MEMBERSHIP` | `CREATE` / `DELETE` | `users.role` 재계산 (token `realm_access.roles` 와 동일 추출 로직, group composite role 매핑 후 update) | `user.group_membership_changed` |
| `USER` | `RESET_PASSWORD` (admin) | (audit only — DevHub `users` 영향 없음) | `account.password_reset_external` |
| `USER` | `DISABLE_CREDENTIALS` | (audit only — token revocation 효과는 다음 token 만료 시점에) | `account.credentials_disabled_external` |

#### 5.3.2 store-level write 정합

- **`users.role` write** = `extractRole(realm_access.roles)` 와 동일 로직. event payload 의 group 정보로 role 재계산
- **`users.status` write** = `enabled=true` → `UserStatusActive`, `enabled=false` → `UserStatusDeactivated`
- **`users.email` / `users.display_name` write** = event payload 의 새 값
- **race condition** — 같은 ms 에 user 가 token refresh + event listener tick 동시 → token claim 기반 read 와 event 기반 write 가 1 tick 안에 정합. write 순서가 last-write-wins 이지만 둘 다 Keycloak 이 source 라 정합

#### 5.3.3 metric 확장 (sprint -u~-y `audit/metrics.go` 정합)

신규 metric 3종:
- `devhub_keycloak_user_sync_total` Counter — label `action` ∈ {`profile`, `membership`, `status`}. event listener 가 DevHub `users` 컬럼 write 한 회수
- `devhub_keycloak_user_sync_errors_total` Counter — write 실패 (DB error 등)
- `devhub_keycloak_user_sync_lag_seconds` Gauge — event timestamp vs DevHub write timestamp 차이

### 5.4 frontend cleanup 매트릭스 (결정 A 동반 #4)

| 파일 | 변경 |
| --- | --- |
| `app/(dashboard)/admin/settings/users/page.tsx` | "Issue Account" 버튼 제거 + Keycloak admin console 안내 link (`${OIDC_ISSUER_URL}/admin/master/console/#/realms/{realm}/users` 또는 사내 운영 문서 link). modal 'Issue / Reset / Disable' action 모두 제거. user list view + role assignment + unit assignment 만 남김 |
| `components/organization/MemberTable.tsx` | `accountService.issueAccount` / `forceResetPassword` / `disableAccount` 호출 제거. PATCH `/api/v1/users/:id` 호출만 유지 |
| `lib/services/account.service.ts` | **파일 자체 제거** — 5 메서드 모두 폐기. (또는 admin-action helper 없는 빈 module 로 archive — 운영 결정) |
| `lib/services/identity.service.ts` | 변경 없음 (`/api/v1/users` + `/api/v1/organization/*` 그대로) |
| `app/(dashboard)/account/page.tsx` | 변경 없음 (sprint -ad Keycloak Account Console redirect 그대로) |

### 5.5 service account 권한 축소 + governance 협약 SOP (결정 A 동반 #5+#6)

#### 5.5.1 service account 권한

| 기존 (sprint -ad 이전) | Phase 2 |
| --- | --- |
| `realm-management.view-users` | **유지** (사실상 사용 안 함 — 결정 A 후 dead. 단 view-events 만 필요한 case 분리 위해 별도 carve) |
| `realm-management.manage-users` | **제거** (account/* endpoint 폐기로 불필요) |
| `realm-management.view-events` | **유지** (audit event listener — sprint -u~-y) |

Phase 2 정합:
- `manage-users` 제거 → service account 가 user create / update / delete 불가
- `view-events` + (선택) `view-users` 만 → read-only + event polling 만
- DevHub backend 가 Keycloak Admin API write 호출 자체 안 함 (코드 변경 — `KeycloakAdminClient.CreateIdentity` / `UpdateIdentityPassword` / `SetIdentityState` / `DeleteIdentity` 메서드 호출처 모두 제거)

#### 5.5.2 governance 협약 SOP

`docs/setup/keycloak_operations.md §8.5c` (신규) 에 협약 문서:

| 운영 동작 | 책임 주체 | 도구 |
| --- | --- | --- |
| user 생성 (account.create) | **Keycloak admin** (IdP 팀) | Keycloak admin console **또는** HRDB ETL push (`scripts/hrdb_etl_sync.sh`) |
| user 비밀번호 reset | **Keycloak admin** (IdP 팀) | Keycloak admin console (Credentials 탭 — Reset Password) |
| user status disable / enable | **Keycloak admin** (IdP 팀) | Keycloak admin console (User detail — Enabled toggle) **또는** HRDB ETL push |
| user 삭제 | **Keycloak admin** (IdP 팀) | Keycloak admin console (Users — Delete) |
| group membership (role 변경) | **Keycloak admin** (IdP 팀) | Keycloak admin console (User detail — Groups 탭) |
| **user 조직 unit assignment** | **DevHub admin** | DevHub `/admin/settings/users` (PATCH `/api/v1/users/:id`) |
| **신규 user 의 unit 초기 배치** (sprint -b Stage 3 P2-2) | **DevHub admin** (lazy auto-create 직후 후속 작업) | (a) HRDB ETL pre-stage 가 unit 정보 동반 시 자동 매핑, (b) 미동반 시 unit 미할당 (`primary_unit_id=NULL`) 으로 lazy create 후 admin 이 `/admin/settings/users` filter `unit_id=null` 로 식별 → 배치. 첫 진입 후 API call 차단 안 함 (단순 unit 미할당 상태) |
| **DevHub `users.role` 직접 수정** | **금지** (event listener 가 자동 sync) | — |
| 조직 unit (department/team) CRUD | **DevHub admin** | DevHub `/admin/settings/organization` (Keycloak 무관) |
| RBAC policy (role × resource × action matrix) 편집 | **DevHub admin** | DevHub `/admin/settings/permissions` |

### 5.6 JWKS stale-while-error 확장 (결정 F)

sprint -r (PR #186) 의 kid mismatch fallback 의 자연 확장. expiry mismatch case 까지.

#### 5.6.1 현재 (sprint -r)

```text
token 검증 시 cache lookup
  ↓ cache hit → 정상 검증
  ↓ cache miss (TTL 만료) → JWKS fetch 시도
  ↓ fetch 성공 → cache update + 정상 검증
  ↓ fetch 실패 → 401
  ↓ kid mismatch 시 (current sprint -r) → cache forced refetch → 1회 retry
```

#### 5.6.2 Phase 2 (확장)

```text
token 검증 시 cache lookup
  ↓ cache hit → 정상 검증
  ↓ cache miss (TTL 만료) → JWKS fetch 시도
  ↓ fetch 성공 → cache update + 정상 검증
  ↓ fetch 실패 (Keycloak unreachable) → **stale key 로 검증 시도** (Phase 2 신규)
  ↓ stale key 로 통과 → 정상 응답 + log 마킹 ("stale-while-error" 표시)
  ↓ stale key 로도 실패 (kid mismatch + Keycloak unreachable) → 401
  ↓ kid mismatch (Keycloak 도달 가능) → cache forced refetch → retry (sprint -r 정합)
```

#### 5.6.3 보안 검토 + mitigation

| 위험 | mitigation |
| --- | --- |
| **revoked key 보호 깨짐** — Keycloak 이 의도적 rotation (security incident) 한 옛 key 가 stale 로 통과 | rotation 직후 운영 SOP — backend 강제 재시작 (JWKS cache 초기화) 또는 SIGHUP 으로 cache flush endpoint |
| stale key TTL 무한 확장 | 별도 max stale duration (예: 24h) 설정. 그 후엔 stale-while-error 도 fail |
| log noise — stale 검증 매번 log emit | log level WARN + structured (token.sub + stale_age_seconds) |

#### 5.6.4 metric

- `devhub_jwks_stale_while_error_total{result="ok|fail"}` Counter
- `devhub_jwks_stale_age_seconds` Histogram — cache 만료 후 stale 사용 시간 분포

### 5.7 `/login` page 정리 (결정 B)

| 파일 | 변경 |
| --- | --- |
| `app/login/page.tsx` | "Sign in with Keycloak" 버튼 + error message 표시 영역. `?error=...` query param 처리 (예: `session_expired`, `login_failed`, `unauthorized`) |
| `app/auth/login/page.tsx` | **제거** (2026-05-22 sub-carve F 옵션 B 채택 — 사용자 결정). 외부 bookmark 호환은 Keycloak realm + nginx + setup-keycloak.sh 의 allowlist URI 정합으로 대체. |
| `app/auth/callback/page.tsx` | error 발생 시 `/login?error=...` 로 redirect (현재 `/auth/error` 와 정합 확인 후 결정) |
| `app/auth/error/page.tsx` | 보존 — `/auth/callback` 의 critical error (예: invalid state) 처리. 일반 error 는 `/login?error=...` |

### 5.8 `rbac_subject_roles` 폐기 (결정 D)

| 파일 | 변경 |
| --- | --- |
| `backend-core/internal/httpapi/rbac.go` | `getSubjectRoles` + `setSubjectRoles` handler 제거 |
| `backend-core/internal/httpapi/router.go` | `v1.GET("/rbac/subjects/:subject_id/roles", ...)` + `v1.PUT(...)` 2 route 제거 |
| `backend-core/internal/httpapi/rbac.go` (RBACStore interface) | `GetSubjectRoles` + `SetSubjectRole` method signature 제거 |
| `backend-core/internal/store/postgres_rbac.go` | `GetSubjectRoles` + `SetSubjectRole` impl 제거 |
| `backend-core/internal/store/postgres_rbac_test.go` | 해당 test 제거 (TestRBAC_SetSubjectRole_* + TestRBAC_GetSubjectRoles_*) |
| `backend-core/internal/httpapi/rbac_test.go` | `fakeRBACStore.GetSubjectRoles` + `SetSubjectRole` mock 제거 |
| `backend-core/internal/httpapi/permissions.go` | `/api/v1/rbac/subjects/:subject_id/roles` 의 routePermissionTable entry 제거 |
| `frontend/lib/services/rbac.service.ts` | dead method 정리 — `getSubjectRoles` + `setSubjectRole` + `SubjectRolesEnvelope` interface 제거 (호출처 0건이지만 dead 코드 정리. sprint -d Stage 3 보강) |
| `docs/backend_api_contract.md` | §12.6 (API-30) + §12.7 (API-31) 본문 spec 폐기 마킹 + §12.8 routing table 2 row strikethrough + §12.10 cache reload trigger 참조 정정 + §12.5 (DELETE policy) 본문 "재할당 안내" 정정 (Keycloak admin console + event listener 경로로) |
| migration | **불필요** — `rbac_subject_roles` 테이블 자체가 없음 (`users.role` 컬럼 직접 write 였음). DB schema 변경 없음 |

### 5.9 Phase 1 매트릭스 오류 정정

§1 책임 분리 매트릭스의 "RBAC subject-role assignment" row 의 "DevHub 측 표상 = `rbac_subject_roles`" 표기 **오류**:
- 실제로 `rbac_subject_roles` 테이블 자체가 없음
- `GetSubjectRoles` 는 `SELECT role FROM users WHERE user_id = $1` (= `users.role` 컬럼 read)
- `SetSubjectRole` 은 `UPDATE users SET role = $2 WHERE user_id = $1` (= `users.role` 컬럼 write)

→ **결정 D 의 (a) 완전 제거로 자연 해소** — endpoint + 메서드 폐기 후 본 row 자체 매트릭스에서 제거.

### 5.10 ADR-0020 후보 outline

본 sprint 의 명시 결정 6건 종합을 ADR-0020 로 승격. draft outline:

```markdown
# ADR-0020: Account/User Management Boundary with External Keycloak

## 1. 컨텍스트
- ADR-0019 채택 후 Keycloak 단일 IdP 운영
- 외부 Keycloak 시나리오 (IdP 팀 별도 운영) 에서 DevHub 의 계정/사용자 관리 책임 경계 재정의 필요
- Phase 1 §1.1 의 source-of-truth 이중화 issue 3건 (role / status / idp_subject) 해결

## 2. 결정 동인
- divergence 원천 차단 (single source-of-truth)
- 운영 거버넌스 명확화 (IdP 팀 vs DevHub admin 책임 분리)
- 보안 권한 최소화 (service account permission)
- OIDC 표준 정합 (signature 검증 의무)

## 3. 검토 옵션 (Phase 1 §4.5 옵션 A~D)
- A 전면 폐기 / B 현재 유지 / C Hybrid / D Read-only mirror

## 4. 결정
- **옵션 A 전면 폐기** 채택 (Phase 2 동반 6건 + 명시 결정 B/C/D/E/F 포함)

## 5. 결과 / 영향
- ... (sprint -a 의 종합 결정 매트릭스 인용)

## 6. 변경 이력
| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-20 | draft (sprint -a 의 명시 결정 6건 종합) | `claude/work_260520-a` |
```

ADR-0020 draft 작성 + 사내 검토 → accepted 처리는 Phase 3 진입 시 동반.

## 6. Phase 3 리팩토링 실행 계획 (sprint `-d`, 2026-05-20)

[ADR-0020](../adr/0020-account-user-management-boundary.md) 가 Phase 2 결정 6건 명문화 + sub-carve 6건 분리. 본 sprint `-d` 가 Phase 3 진입 (sub-carve A 흡수 + 후속 sub-carve B~F 분리).

### 6.1 sub-carve 분담

| sub-carve | 영역 | 영향 파일 (요약) | 위험 | sprint |
| --- | --- | --- | --- | --- |
| **A** — ADR-0020 + design §6 + `rbac_subject_roles` 폐기 (결정 D) | `docs/adr/0020-*` + design doc §6 + backend rbac.go/router.go/permissions.go/postgres_rbac.go + test | docs + 격리된 dead code 제거 | 낮음 | ✅ **`-d` (PR #205 `f2a389a`)** |
| **B (backend)** — `/api/v1/accounts/*` 4 endpoint 제거 + `authenticateActor` lazy auto-create 실 구현 + audit action 2종 신규 | backend `accounts_admin.go` 삭제 + router + permissions + `lazy_auto_create.go` 신규 + `AuthenticatedActor` 확장 | backend 만, e2e 영향 없음 (frontend admin actions 는 별도 sprint) | 낮음~중간 | ✅ `-i` (sprint `claude/work_260520-i-209-accounts-deprecation`, PR TBD) |
| **B (frontend)** — `account.service.ts` 폐기 + admin/settings/users page 의 admin actions 제거 + e2e TC-ACC-* 갱신 | frontend `account.service.ts` 삭제 + `app/(dashboard)/admin/settings/users/page.tsx` + e2e | frontend cleanup + e2e | 중간 (e2e 회귀) | Gemini 별도 sprint (sprint label 미배정) |
| **C** — event listener 확장 (USER:UPDATE / GROUP_MEMBERSHIP / USER:DELETE) | backend `keycloak_event_puller.go` + `audit_logs` action 매핑 + `users` write + metric 3종 (`audit/metrics.go`) + `user_sync.go` 신규 + `KeycloakAdminClient.GetUserDetails`/`GetUserGroups` 신규 + `main.go` wire | sprint -u~-y 자연 확장. event handler 가 DevHub `users` write 추가 | 중간 (event listener 회귀 위험) | ✅ `-k` (sprint `claude/work_260520-k-212-event-listener-users-sync`, PR TBD) |
| **D** — JWKS stale-while-error expiry case 확장 | backend `keycloak_verifier.go` + JWKS cache + metric 2종 + `internal/auth/metrics.go` 신규 + config OIDCJWKSMaxStaleDuration | sprint -r kid mismatch fallback 자연 확장 | 낮음 | ✅ `-l` (sprint `claude/work_260520-l-213-jwks-stale-expiry`, PR TBD) |
| **E** — service account 권한 축소 + governance SOP | `keycloak_operations.md §8.5c` 신규 + §3.2 SOP 갱신 | docs only | 낮음 | `-i` (후속) |
| **F** — `/login` page 정리 (결정 B) | frontend `app/login/page.tsx` 본문 swap + `app/auth/login/` **삭제** + callback/error/signup/AuthGuard/onboarding + role-routing.test + vitest 신규 + infra (realm.prod.json/nginx/setup-keycloak.sh) + 5 docs allowlist URI 정합 | frontend UX + infra/docs 일괄. 사내 운영자 Keycloak admin allowlist 1회 갱신 동반 | 낮음 (frontend) + 중간 (운영 동반) | ✅ `claude/work_260522-adr-0020-subcarve-f-login` (PR TBD) — 옵션 B (사용자 결정) |
| (신규) — **Keycloak SPI provider JAR** (PR #203 codex P2 후속) | `infra/idp/devhub-event-listener/` (신규 SPI module 빌드) + `docker-compose.deploy.yml` (volume mount) + `keycloak_operations.md` 운영 SOP | 사내 인프라 동반 — Keycloak SPI Java 빌드 + Maven/Gradle 자산 | 중간 (사내 인프라 결정 동반) | TBD (사내 인프라 진입 시) |

#### 6.1.1 Phase 3 closing status (2026-05-22)

본 sprint `-d` 진입 이후 누적 결과 — **8 carve 중 6 closed + 1 사내 동반 + 1 잔여 (P3)**.

| 진행 그룹 | 항목 |
| --- | --- |
| ✅ **closed (7/8)** | A (PR #205) + B-backend (#239, lazy 폐기는 #290) + B-frontend (gemini #246) + C (#241) + D (#242) + E (#244) + **F (sprint `claude/work_260522-adr-0020-subcarve-f-login`)** |
| 🟡 **잔여 — 사내 동반** | **SPI provider JAR** (P2) — Keycloak SPI Java 빌드 + Maven/Gradle 자산 |

→ ADR-0020 핵심 결정 + frontend UX 정리 모두 적용 완료. 잔여는 SPI JAR (사내 인프라 동반) 1건만. ADR-0020 §4.1.1 cross-link.

### 6.2 sub-carve A — `rbac_subject_roles` 완전 제거 (결정 D, 본 sprint 흡수 범위)

§5.8 의 8 파일 변경 — backend 의 dead-end endpoint 제거 (frontend UI 미구현). 위험 낮음 (Keycloak group composite 가 실 권한 source, 본 endpoint 는 `users.role` 컬럼 직접 write 였음).

#### 변경 매트릭스

| 파일 | 변경 |
| --- | --- |
| `backend-core/internal/httpapi/rbac.go` | `getSubjectRoles` + `setSubjectRoles` handler 제거 + `rbacAuditActionAssigned` const 제거 + `rbacSubjectRolesRequest` wire struct 제거 + `RBACStore` interface 의 `GetSubjectRoles` + `SetSubjectRole` 메서드 제거 |
| `backend-core/internal/httpapi/router.go` | `v1.GET("/rbac/subjects/:subject_id/roles", ...)` + `v1.PUT(...)` 2 route 제거 |
| `backend-core/internal/httpapi/permissions.go` | `routePermissionTable` 의 `/rbac/subjects/:subject_id/roles` 2 entry 제거 |
| `backend-core/internal/store/postgres_rbac.go` | `GetSubjectRoles` + `SetSubjectRole` impl 제거 |
| `backend-core/internal/store/postgres_rbac_test.go` | `TestRBAC_SetSubjectRole_*` + `TestRBAC_GetSubjectRoles_*` 제거 |
| `backend-core/internal/httpapi/rbac_test.go` | `fakeRBACStore.GetSubjectRoles` + `SetSubjectRole` mock + 관련 handler test 제거 |
| `frontend/lib/services/rbac.service.ts` | dead method 정리 (sprint -d Stage 3 보강) — `getSubjectRoles` + `setSubjectRole` + `SubjectRolesEnvelope` interface 제거 |
| `docs/backend_api_contract.md` | §12.6/§12.7 본문 spec 폐기 마킹 + §12.8 routing table 2 row + §12.10 cache reload trigger + §12.5 본문 정정 (sprint -d Stage 3 보강) |
| migration | **불필요** — `rbac_subject_roles` 테이블 자체가 없음 (`users.role` 컬럼 직접 write 였음). DB schema 변경 없음 |

#### 검증

- backend `go build ./...` PASS
- backend `go test ./internal/httpapi/... ./internal/store/...` PASS (제거된 test 외 회귀 없음)
- traceability `docs/traceability/report.md` §2 API-26..40 매트릭스의 `/rbac/subjects/:subject_id/roles` row 제거 + §6 변경 이력 row

### 6.3 sub-carve B~F 진입 순서 권장

1. **B (accounts/* 제거 + lazy auto-create)** — Phase 3 의 가장 큰 carve. Keycloak admin 책임 이관의 실질적 출발점
2. **C (event listener 확장)** — B 와 의존성 있음 (lazy auto-create 의 role 추출 로직과 event listener 의 role 매핑 로직 공유 → `extractKeycloakRole` 공유 함수)
3. **E (governance SOP)** — B + C 완료 후 운영 협약 SOP 작성. 사내 IdP 팀과의 협약 동반
4. **D (JWKS expiry case 확장)** — 독립적, B~C 와 무관. 언제든 진행 가능
5. **F (`/login` 정리)** — 우선순위 가장 낮음. frontend UX 정리만

### 6.4 traceability 영향 (sub-carve A 본 sprint)

| 단계 | 영향 |
| --- | --- |
| REQ | 없음 (endpoint 제거는 REQ 변경 아님) |
| ARCH | 없음 |
| API | API-26..40 (RBAC) 매트릭스의 `/rbac/subjects/:subject_id/roles` row 2개 제거 |
| RM | 없음 |
| IMPL | `IMPL-rbac-01` (handler — getSubjectRoles/setSubjectRoles 2개 제거 후 4 endpoint) + `IMPL-rbac-02` (store — GetSubjectRoles/SetSubjectRole 2 method 제거 후 6 method) 갱신. `IMPL-rbac-03` (permissions.go route table) 도 2 entry 제거. |
| UT | `TestRBAC_GetSubjectRoles_*` + `TestRBAC_SetSubjectRole_*` (postgres_rbac_test.go + rbac_test.go) 제거. `TestRBAC_DeleteCustomRoleInUse` 는 `CreateUser` 의 `Role: domain.AppRole(roleID)` 로 재구성 (`users.role` FK to `rbac_policies.role_id` 활용, migration 000006). |
| TC | 없음 (e2e 미구현이라 TC-RBAC-SUBJECT-* 없었음) |

### 6.5 Strangler Fig 패턴 (sub-carve B 진입 시 적용)

sub-carve B (accounts/* 제거) 는 frontend UI 호출과 backend handler 가 함께 변경되므로 명시 단계 분리 권장:

1. **deprecation banner** — `/api/v1/accounts/*` 4 endpoint 에 `X-Devhub-Deprecation: 410-after-2026-XX-XX` 헤더 추가 + audit `account.deprecated_call_warning` (1 sprint)
2. **frontend cleanup** — admin/settings/users 의 admin actions 제거 + `account.service.ts` 폐기 (다음 sprint)
3. **backend route 제거** — handler 4개 삭제 + service account `manage-users` permission 제거 SOP 안내 (다음 sprint)
4. **e2e spec 정리** — TC-ACC-* 의 admin issue/disable 시나리오 제거 + Keycloak admin console flow 안내 link 검증으로 대체

본 sprint `-d` 는 sub-carve A 만. Strangler Fig 의 1단계 deprecation 은 sub-carve B 진입 시 다시 검토.

## 7. 잔여 carve

### 7.1 Phase 2 결정으로 resolved

- ✅ `users.role` source-of-truth — **결정 C** (event listener 확장) 로 해소. §5.3 매핑 표 참조
- ✅ `account.service.ts` dead UI 메서드 — **결정 A** 의 frontend cleanup 으로 모듈 제거. §5.4 참조
- ✅ `PUT /api/v1/rbac/subjects/:subject_id/roles` 처리 — **결정 D** 완전 제거. §5.8 참조
- ✅ `POST /api/v1/accounts` vs `/api/v1/users` 중복 — **결정 A** 로 accounts POST 폐기. `/api/v1/users` POST 만 유지 (또는 lazy auto-create 자동 흐름)
- ✅ email / display_name sync — **결정 C** event listener 확장 (`USER:UPDATE` 매핑) 으로 해소

### 7.2 Phase 3 (실 구현) sprint 영역

**Closed (7/8, 2026-05-22 기준)** — §6.1.1 closing status 참조:
- ✅ resolved (sprint `-d`, PR #205 `f2a389a`) — ADR-0020 발급 + `rbac_subject_roles` 완전 제거 (sub-carve A)
- ✅ resolved (sprint `-i`, PR #239 `d21e801`) — backend `/api/v1/accounts/*` 4 endpoint 제거 + lazy auto-create (sub-carve B backend). lazy auto-create 자체는 PR #290 `fa042c5` 가 ADR-0021 §3.3 따라 폐기 (token-only actor + onboarding 제출 시점 INSERT)
- ✅ resolved (gemini `work_260520-a-209-accounts-cleanup`, PR #246 `b1e34bd`) — frontend `account.service.ts` 폐기 + admin/settings/users page 의 admin actions 제거 (sub-carve B frontend)
- ✅ resolved (sprint `-k`, PR #241 `9ea7e1c`) — Keycloak event listener 확장 (USER:UPDATE / GROUP_MEMBERSHIP / USER:DELETE 매핑) + `users` write + metric 3종 (sub-carve C)
- ✅ resolved (sprint `-l`, PR #242 `cb6646d`) — JWKS stale-while-error expiry case 확장 (sub-carve D)
- ✅ resolved (sprint `-n`, PR #244 `6810384`) — service account 권한 축소 정공법 + governance SOP (sub-carve E, `keycloak_service_account_min_role.md` 신규)
- ✅ resolved (sprint `claude/work_260522-adr-0020-subcarve-f-login`, PR TBD) — `/login` canonical page swap (`/auth/login` 96 LoC → `/login` + `?error=` 처리) + **`/auth/login` 완전 제거** (사용자 결정 옵션 B) + AuthGuard 401 fallback `/login?error=session_expired` + 호출처 8 위치 sync + vitest 회귀 가드 8건 + infra/scripts (realm.prod.json / nginx template / setup-keycloak.sh) 3건 + 5 docs allowlist URI 정합. **사내 운영자 1회 작업 동반** — Keycloak admin console 의 `devhub-frontend` client 의 Valid Post Logout Redirect URIs allowlist 갱신 (sub-carve F)

**잔여 사내 동반 (1/8)**:
- **(carve, SPI provider JAR, P2)** — Keycloak SPI Java 빌드 + Maven/Gradle 자산 + docker-compose volume mount. event listener push 전환 (현재 polling 이 정공법).

### 7.3 사내 동반 carve

> **[2026-05-22 docs 초안 resolved]** sprint `claude/work_260522-internal-coordinated-carve-docs` 가 3 carve 의 docs 초안 신규. 사내 실 적용 (IdP 팀 / HRDB 팀 / Security sign-off) 은 별도.

- ✅ resolved (docs 초안) — [HRDB ETL push 의 unit 매핑 정보 stage](../setup/hrdb_unit_pre_stage.md). 3 옵션 (A Keycloak user attribute 권장 / B DevHub hrdb.persons 보조 / C self-service only) + ADR-0021 §6.2 cross-check 후속 carve 결정 옵션
- ✅ resolved (docs 초안) — [Keycloak admin 운영 SOP 승격](../governance/keycloak_admin_responsibility.md). §5.5.2 governance 표를 사내 정책 문서로 승격 (IdP 팀 vs DevHub 운영자 책임 매트릭스 5 sub-section + escalation 4 level + 명시 금지 5건)
- ✅ resolved (docs 초안) — [JWKS rotation 직후 backend cache flush SOP](../setup/jwks_rotation_cache_flush.md). §5.6.3 mitigation 운영 SOP — backend 강제 재기동 4 환경별 + 검증 4 step + cache flush endpoint carve P3

## 8. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-20 | sprint `claude/work_260520-a` Phase 1 — 현황 파악 1차 작성. §1 책임 분리 매트릭스 + §2 backend 23 endpoint (18 row) + §3 frontend 4 page + service 매트릭스 + §4 DB schema (users 14 컬럼) + Keycloak attribute 정합 + §4.5 Phase 2 입력 옵션 A~D 후보 + §7 잔여 carve 5건. §5/§6 는 Phase 2/3 별도 sprint. |
| 2026-05-20 | Self-review Stage 3 보강 — P1×2 + P2×1 흡수. (P1-1) §2 header endpoint count 17→23 + 묶음 row 안내 추가. (P1-2) §4.1 컬럼 count 14 (commit msg 정정). (P2-1) §4.5 옵션 A "lazy users row 자동 생성" wording → footnote 로 현재 mechanism (idp_subject 만 backfill) 과 신규 mechanism (row auto-create) 명확화. |
| 2026-05-20 | sprint `claude/work_260520-a` Phase 2 design 작성 — 명시 결정 6건 토론 결과 (Q&A 6 round) 흡수. (1) §0/§4.5 결정 표기, (2) §5 신규 10 sub-section — §5.1 명시 결정 6건 종합 표 + §5.2 lazy auto-create mechanism (3 흐름 + 결정 항목 5 + 보안 검토) + §5.3 event listener 확장 (매핑 표 + store-level write + metric 3) + §5.4 frontend cleanup 매트릭스 (5 파일) + §5.5 service account 권한 축소 (`manage-users` 제거) + governance 협약 SOP 표 + §5.6 JWKS stale-while-error 확장 (sprint -r 자연 확장 + mitigation) + §5.7 `/login` 정리 + §5.8 `rbac_subject_roles` 폐기 (8 파일) + §5.9 Phase 1 §1 매트릭스 오류 정정 (rbac_subject_roles 테이블 없음, `users.role` 직접 write) + §5.10 ADR-0020 후보 outline. (3) §7 잔여 carve 갱신 — Phase 2 결정으로 resolved 5건 + Phase 3 실 구현 carve 7건 + 사내 동반 carve 3건. **명시 결정 A~F 모두 확정** — A 전면 폐기 / B entry minimal / C event listener 확장 / D rbac_subject_roles 완전 제거 / E read-only carve 도입 안 함 (self-reverse) / F JWKS stale-while-error 확장. Phase 3 (실 구현) 은 별도 sprint. |
| 2026-05-20 | sprint `claude/work_260520-l-213-jwks-stale-expiry` — **sub-carve D 실 구현 완료**. 5 commit. (1) `keycloak_verifier.go` cache struct 확장 — `cachedAt` 필드 + `MaxStaleDuration` 필드 + `defaultJWKSMaxStale = 24h` const + `readStaleCachedKeys` helper. (2) `fetchJWKS` 흐름 — network fetch fail 시 readStaleCachedKeys fallback + log WARN + observe metric. `fetchAndCacheJWKS` 별도 함수로 분리. (3) `internal/auth/metrics.go` 신규 — sync.Once init pattern + 2 metric (devhub_jwks_stale_while_error_total{result} CounterVec + devhub_jwks_stale_age_seconds Histogram). (4) `Config.OIDCJWKSMaxStaleDuration` env + main.go wire (jwksVerifier 변수 분리 + time.ParseDuration + invalid fallback). (5) 회귀 test 4건. backward compatible — MaxStaleDuration 0 시 defaultJWKSMaxStale (24h) 적용. errKidMismatch 흐름 (sprint -r) 그대로 유지. backend `go test ./...` PASS 전 패키지. |
| 2026-05-20 | sprint `claude/work_260520-k-212-event-listener-users-sync` — **sub-carve C 실 구현 완료**. 5 commit 분리: (1) `KeycloakAdminClient.GetUserDetails` + `GetUserGroups` 신규 (admin event 시 user state fetch). (2) `audit/user_sync.go` 신규 (207 lines) — `SyncUserProfile`/`Membership`/`MarkUserDeactivated` + 4 helper (composeDisplayName/pickHighestPriorityRole/groupNameToRole/ParseIdentityIDFromResourcePath) + narrow interface 2종 (UserSyncOrgStore + UserSyncAdminClient) + `SyncUserAction` enum (profile/membership/status). (3) `keycloak_event_puller.go` 확장 — `KeycloakEventPullerOptions.UserSync UserSyncCallback` 필드 + `classifyAdminEventForSync` helper + `mapAdminEventToAudit` 에 `GROUP_MEMBERSHIP:CREATE/DELETE` 2 row 추가 (8 row → 10 row) + admin event loop 의 sync callback 호출 분기. (4) `audit/metrics.go` 확장 — 신규 metric 3종 + 3 observe helper + 회귀 test 4건 (TestMapAdminEventToAudit GROUP_MEMBERSHIP 추가 + TestClassifyAdminEventForSync 5 case + TestPullAdminEvents_InvokesUserSyncCallback + TestPullAdminEvents_NilUserSync_BackwardCompatible). (5) `main.go` wire — `UserSync` callback dispatcher (action 별 SyncUserProfile/Membership/MarkUserDeactivated + metric observe + error log + backward compatible nil-safe). ADR-0020 §5.3 결정 정합 — group composite role priority filter (system_admin > team_manager > manager > developer, sprint -j PR #185 정합) + soft delete (users.status=deactivated, §5.3.1 P1-2). backward compat — UserSync nil 인 경우 sprint -u~-y 동작 동등. backend `go test ./...` PASS 전 패키지. |
| 2026-05-20 | sprint `claude/work_260520-i-209-accounts-deprecation` — **sub-carve B (backend) 실 구현 완료**. (1) `backend-core/internal/httpapi/accounts_admin.go` (338 lines) + `accounts_admin_test.go` 삭제. (2) router.go 4 route 제거 + permissions.go 4 entry 제거 + ADR-0020 reference 주석. (3) `lazy_auto_create.go` 신규 — `lazyAutoCreateUser` + `lazyAutoCreateDefaultRole` (developer) + `lazyAutoCreateDefaultStatus` (active) + `lazyAutoCreateAuditAction` (account.lazy_provisioned) + `lazyRoleDefaultAuditAction` (user.role_default_assigned) + `isValidLazyRole` + `roleSourceLabel` helper. (4) `AuthenticatedActor` 에 Email/DisplayName 필드 추가 + `keycloak_verifier.go::extractDisplayName` 신규 (name → given+family → empty 우선순위). (5) `authenticateActor::ErrNotFound` 분기 — `lazyAutoCreateUser` 호출로 확장 (이전엔 token role claim fall-through 만). (6) 회귀 test 5건 (Happy / DefaultsRole / PMOManager / FallbackDisplayName / SkippedWhenStoreNil) + 기존 test 3건 admin pre-seed fix (audit count drift 회피). (7) testhelpers_test.go 신규 — doJSON helper 보존. (8) identity_resolver_test.go 의 `TestCreateAccount_EagerBackfillsIdPSubject` 제거. backend `go test ./...` PASS 전 패키지. **frontend cleanup (account.service.ts + admin/settings/users + e2e) 은 Gemini 별도 sprint**. ADR-0020 §4.1 sub-carve 표 sprint label shift 동반 (C/D/E/F → `-g/-h/-i/-j`). |
| 2026-05-20 | sprint `claude/work_260520-d` Stage 3 보강 #2 — PR #205 codex review P1 응답. `validAppRoles` 에 `team_manager` 추가 (`backend-core/internal/httpapi/organization.go`) + error message 정정 + 회귀 test 2건 (`TestCreateUserAcceptsPMOManager` + `TestUpdateUserAcceptsPMOManagerRole`). ADR-0020 §5.5 신규 hotfix 섹션 (sprint -f 의 event listener sync 가 정공법, sprint -d hotfix 는 backward compat 임시) + §7 변경 이력 row. **codex P1 회귀 응답** — sprint -d 의 `/rbac/subjects/:id/roles` 폐기 후 `team_manager` 가 API 로 할당 불가능했던 회귀 해소. custom role 임의 할당은 결정 C 의 event listener (sprint -f) 자연 흡수 + sub-carve E (sprint -h) governance SOP 동반. |
| 2026-05-20 | sprint `claude/work_260520-d` Phase 3 진입 (sub-carve A) — (1) ADR-0020 발급 ([docs/adr/0020-account-user-management-boundary.md](../adr/0020-account-user-management-boundary.md) 신규) — Phase 2 명시 결정 6건 종합 + sub-carve B~F 분리 plan. (2) §6 Phase 3 실행 계획 신규 — sub-carve 분담 표 + sub-carve A 변경 매트릭스 + sub-carve B~F 진입 순서 권장 + Strangler Fig 패턴 (sub-carve B 진입 시 적용). (3) §5.8 따라 `rbac_subject_roles` 완전 제거 — backend `rbac.go` handler 2개 + interface method 2개 + audit action const + wire struct 제거, `router.go` 2 route 제거, `permissions.go` 2 routePermissionTable entry 제거, `postgres_rbac.go` impl 2개 제거, `postgres_rbac_test.go` 테스트 3건 제거 (Delete-in-use 테스트는 `CreateUser` 의 `Role: domain.AppRole(roleID)` 로 재구성), `rbac_test.go` fake mock 2개 + handler test 3건 제거. (4) §7.2 Phase 3 carve list 갱신 — sub-carve A resolved 2건 + sub-carve B~F (5건 carve) 분담. (5) traceability §2.2 API-30/31 strikethrough + §2.4 IMPL-rbac-01/02 책임 갱신 + §4 ADR-0020 row 추가 + §6 변경 이력 row. backend go build + go test (httpapi + store) PASS. |
| 2026-05-20 | sprint `claude/work_260520-b` Self-review Stage 3 보강 — P1×3 + P2×2 일괄 흡수. (P1-1) §4.3 + §1 매트릭스의 `rbac_subject_roles` 표기 정정 (취소선 + §5.9 link). (P1-2) §5.3.1 `USER:DELETE` 정책 결정 명시 — soft delete (`status=deactivated`) 채택, audit_logs actor reference 깨짐 회피. (P1-3) §5.2.2 role 매핑 fallback 결정 명시 — token `realm_access.roles` 비어 있을 때 default `developer` 부여 + audit `user.role_default_assigned`. (P2-1) §5.3.3 metric label 표기 정정 `action="profile|membership|status"` → `label action ∈ {profile, membership, status}`. (P2-2) §5.5.2 governance 표 row 추가 — 신규 user 의 unit 초기 배치 (HRDB ETL pre-stage 자동 또는 admin filter 후속 배치, 첫 API call 차단 안 함). |
| 2026-05-22 | sprint `claude/work_260522-adr-0020-phase3-closing-housekeeping` — **Phase 3 closing status 명문화**. §6.1.1 신규 sub-section (8 carve 중 6 closed + 1 잔여 P3 [F] + 1 사내 동반 [SPI JAR] 표 + memory directive 정정 안내) + §7.2 carve list 의 closed 6건 SHA 표기 (PR #205/#239/#246/#241/#242/#244) + 잔여 active 2건 (F + SPI) 그룹 분리. ADR-0020 §4.1.1 cross-link. main flat memory directive 의 misleading "8 carve" 표현 정정. |
| 2026-05-22 | sprint `claude/work_260522-adr-0020-subcarve-f-login` — **sub-carve F resolved** (`/login` canonical page). `/auth/login` 96 LoC → `/login/page.tsx` swap (DevHub brand + `?error=` query 처리 + 자동 OIDC redirect 분기 — error 진입 시 자동 redirect 차단). `/auth/login` 14 LoC stub (외부 bookmark 호환). AuthGuard 401 fallback `/login?error=session_expired` + 비-401 `/login?error=login_failed`. 호출처 8 위치 sync — AuthGuard.tsx 2 + onboarding 1 + auth/callback 1 (error_description propagate) + auth/error 1 + auth/signup 1 + role-routing.test 1. vitest 신규 — `app/login/page.test.tsx` (resolveErrorMessage 6 unit) + AuthGuard.test.tsx fallback 회귀 2건. §6.1.1 표 F → done + §7.2 closed 7/8 + §6.1 sub-carve 분담 표 F sprint label resolved. |
| 2026-05-22 | sprint `claude/work_260522-internal-coordinated-carve-docs` — **§7.3 사내 동반 carve 3건 docs 초안 resolved**. (1) `docs/governance/keycloak_admin_responsibility.md` 신규 (IdP 팀 vs DevHub 운영자 책임 매트릭스 5 sub-section + escalation 4 level + 명시 금지 5건 + 변경 절차). (2) `docs/setup/jwks_rotation_cache_flush.md` 신규 (revoked key 위협 대응 — backend 강제 재기동 4 환경 + 검증 4 step + cache flush endpoint carve P3). (3) `docs/setup/hrdb_unit_pre_stage.md` 신규 (3 옵션 + onboarding cross-check 후속 carve 3 결정 옵션). ADR-0020 §6.3 + redesign §7.3 표의 3 carve 모두 docs 초안 mark. 사내 실 적용은 별도. |
