# 계정/사용자 관리 리팩토링 — 외부 Keycloak 가정 (redesign)

- 문서 목적: 외부 Keycloak 을 단일 IdP 로 사용한다는 전제로, DevHub 의 **계정 관리** + **사용자 관리** 의 책임 경계를 재정의하고 리팩토링한다. 본 문서 §1~§4 는 **Phase 1 — 현황 파악**, §5+ 는 후속 sprint 의 **Phase 2 — 책임 분리 design** + **Phase 3 — 리팩토링 실행 계획**.
- 범위: backend `/api/v1/accounts/*` + `/api/v1/users/*` + `/api/v1/organization/*` + `/api/v1/rbac/*` + `/api/v1/me` endpoint, frontend `/account` + `/admin/settings/{users,organization,permissions}` page, DB schema (users + organization_units + rbac_policies + rbac_subject_roles), Keycloak realm 의 user/group/role + token claim 정합.
- 대상 독자: 백엔드/프론트엔드/IdP 담당, 운영 (SRE), 보안.
- 상태: draft (Phase 1 only)
- 최종 수정일: 2026-05-20
- 결정 근거 sprint: `claude/work_260520-a` (Phase 1 현황 파악)
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
| **role** (developer/manager/pmo_manager/system_admin) | **이중 source** — Keycloak `realm_access.roles` (group composite) + DevHub `users.role` (cache) | `users.role` | token 검증 시 `realm_access.roles` 가 권한 wire format 우선 (`keycloak_verifier.go:260`). `users.role` 은 화면 표시 + audit 용 cache (sprint -j PR #185 multi-role priority filter) |
| **조직 단위 (unit)** | DevHub | `organization_units` + `users.{primary_unit_id, current_unit_id}` | DevHub 자체 (Keycloak 무관) |
| **조직 계층 (hierarchy)** | DevHub | `organization_hierarchy` (parent_unit_id) | DevHub 자체 |
| **RBAC policy** (Resource × Action) | DevHub | `rbac_policies` | DevHub 자체 (`/api/v1/rbac/policies`) |
| **RBAC subject-role assignment** | DevHub (admin 직접) **또는** Keycloak group composite (간접) | `rbac_subject_roles` (`/api/v1/rbac/subjects/:subject_id/roles`) | endpoint 정의됨, frontend UI 미구현 (carve). 실 권한 평가는 token claim 의 `realm_access.roles` 우선 |
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
| `role` | TEXT (CHECK developer/manager/system_admin/pmo_manager) | `Role` | `realm_access.roles` (group composite 매핑) | **cache** — backend 권한 평가는 token claim 우선 |
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

### 4.3 `rbac_policies` + `rbac_subject_roles`

| 테이블 | 핵심 컬럼 | 역할 |
| --- | --- | --- |
| `rbac_policies` | `role_id` (TEXT), `resource` + `action`, `allow` | role × resource × action 매트릭스. system_admin / developer / manager / pmo_manager seed (migration 000004 + 000005) |
| `rbac_subject_roles` | `subject_id` + `role_id` | user 별 role override (간접 — Keycloak group composite 가 1차) |

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

각 옵션의 trade-off + ADR 발급은 Phase 2 sprint.

[^A-lazy]: 현재 (sprint -j PR #188) 의 자동 sync 는 `authenticateActor` 가 **`users.idp_subject` 컬럼만** lazy backfill (이미 존재하는 row 의 idp_subject 가 비어 있으면 Keycloak FindIdentityByUserID 로 조회 후 세팅). `users` row 자체가 없는 경우의 자동 생성은 **현재 mechanism 없음** — token 검증 시 GetUser miss → ErrIdentityNotFound. 옵션 A 채택 시 신규 mechanism 으로 lazy auto-create 추가 필요.

## 5. (TBD — Phase 2) 책임 분리 design

Phase 2 sprint 에서 작성:

- §5.1 옵션 A/B/C/D 의 trade-off 표 (운영 부담 / divergence risk / 사용자 경험 / 마이그레이션 비용)
- §5.2 권장 옵션 (잠정 — Option A 또는 D 가 ADR-0019 정공법 정합)
- §5.3 provisioning 흐름 design (lazy / HRDB ETL push / SCIM bridge)
- §5.4 role/status sync 정합 design (token 검증 시 cache invalidate / event listener 활용)
- §5.5 ADR-0020 후보 — "Account/User management boundary with external Keycloak"

## 6. (TBD — Phase 3) 리팩토링 실행 계획

Phase 3 sprint 에서 작성:

- §6.1 endpoint 폐기/유지 결정에 따른 backend migration (route 제거 + handler 삭제)
- §6.2 frontend UI 변경 (admin/settings/users 의 admin endpoint 호출 → read-only mirror 또는 제거)
- §6.3 DB schema 정리 (필요 시 `users.role` deprecation 또는 cache-only 표시)
- §6.4 traceability 영향 (REQ/API/IMPL/TC row 갱신)
- §6.5 마이그레이션 단계 (Strangler Fig — 기존 endpoint deprecation banner → 새 흐름 도입 → 폐기)

## 7. 잔여 carve (Phase 1 시점 식별)

- **(carve)** `users.role` 의 source-of-truth 정합 mechanism — token 검증 시 cache 자동 sync 또는 Keycloak event listener 가 group change 감지 → DevHub users.role update
- **(carve)** account.service.ts 의 dead UI 메서드 (`unlockAccount`, `deleteAccount`) — Phase 2 결정 따라 UI 추가 또는 service 제거
- **(carve)** `PUT /api/v1/rbac/subjects/:subject_id/roles` UI 또는 endpoint 제거 결정 — Phase 2 design 입력
- **(carve)** `POST /api/v1/accounts` vs `POST /api/v1/users` atomic 단일 작업 통합 또는 명확한 책임 분리 (one-shot create with unit assignment)
- **(carve)** email / display_name 변경 시 Keycloak → DevHub 동기화 (현재 미동기화 — account.create 시 1회만)

## 8. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-20 | sprint `claude/work_260520-a` Phase 1 — 현황 파악 1차 작성. §1 책임 분리 매트릭스 + §2 backend 23 endpoint (18 row) + §3 frontend 4 page + service 매트릭스 + §4 DB schema (users 14 컬럼) + Keycloak attribute 정합 + §4.5 Phase 2 입력 옵션 A~D 후보 + §7 잔여 carve 5건. §5/§6 는 Phase 2/3 별도 sprint. |
| 2026-05-20 | Self-review Stage 3 보강 — P1×2 + P2×1 흡수. (P1-1) §2 header endpoint count 17→23 + 묶음 row 안내 추가. (P1-2) §4.1 컬럼 count 14 (commit msg 정정). (P2-1) §4.5 옵션 A "lazy users row 자동 생성" wording → footnote 로 현재 mechanism (idp_subject 만 backfill) 과 신규 mechanism (row auto-create) 명확화. |
