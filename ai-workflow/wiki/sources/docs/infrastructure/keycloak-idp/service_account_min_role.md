---
title: service_account_min_role
type: source
tags: [infrastructure, service_account_min_role.md, project-devhub]
sources: [raw/projects/devhub/docs/infrastructure/keycloak-idp/service_account_min_role.md]
git_commit: fb3894f7
git_branch: chore/260622-wiki-drift-cleanup-3
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T06:03:34Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# Keycloak Service Account 최소 권한 설계 (ADR-0020 sub-carve E)

- 문서 목적: DevHub backend 의 Keycloak service account (`devhub-backend` client) 에 부여된 `manage-users` realm role 을 제거하기 위한 호출처 매트릭스, 옵션 비교, 권장 안, 운영 SOP 를 정의한다.
- 범위: backend `KeycloakAdminClient` 호출처 5건 + Keycloak realm role 요구사항 매트릭스 + 단계적 권한 축소 옵션 + service account 권한 축소 운영 SOP.
- 대상 독자: backend 개발자, Keycloak admin (사내 운영팀), 보안 검토자, 후속 sprint 진입자.
- 상태: draft
- 최종 수정일: 2026-05-20 (sprint claude/work_260520-n-214-service-account-min-role)
- 관련 문서: [ADR-0020 계정/사용자 관리 책임 경계](../adr/0020-account-user-management-boundary.md), [ADR-0019 Keycloak 단일화](../adr/0019-keycloak-only-idp.md), [keycloak_operations.md](../setup/keycloak_operations.md), [account_user_management_redesign.md](./account_user_management_redesign.md).

## 1. 배경

ADR-0020 (계정/사용자 관리 책임 경계) 의 결정 A 는 "Keycloak admin = 별도 운영팀, DevHub 는 admin write 미관여" 정공법이다.

- **결정 A** (`/api/v1/accounts/*` 4 endpoint 폐기) — PR #239 sprint -i 가 완료.
- **결정 C** (event listener 확장 + DevHub `users` sync) — PR #241 sprint -k 가 완료.
- **결정 D** (JWKS stale-while-error expiry case 확장) — PR #242 sprint -l 가 완료.

본 sub-carve E 의 목적은 ADR-0020 결정 A 의 자연 연속 작업이다. service account 가 보유한 `manage-users` realm role 을 제거하여 **최소 권한 원칙 (PoLP, Principle of Least Privilege)** 에 정합시킨다.

다만 PR #239 이후에도 backend 가 `manage-users` 를 요구하는 호출처가 **2건 잔존** 함을 본 sprint 가 식별. design 결정 후 호출처 정리.

## 2. 현황 매트릭스

### 2.1 `KeycloakAdminClient` method 별 Keycloak realm role 요구사항

| Method | Keycloak API | Required realm role |
| --- | --- | --- |
| `CreateIdentity` | `POST /admin/realms/{realm}/users` | **manage-users** |
| `UpdateIdentityPassword` | `PUT /admin/realms/{realm}/users/{id}/reset-password` | **manage-users** |
| `SetIdentityState` | `PUT /admin/realms/{realm}/users/{id}` (enabled) | **manage-users** |
| `DeleteIdentity` | `DELETE /admin/realms/{realm}/users/{id}` | **manage-users** |
| `FindIdentityByUserID` | `GET /admin/realms/{realm}/users?username=` | view-users |
| `GetUserDetails` | `GET /admin/realms/{realm}/users/{id}` | view-users |
| `GetUserGroups` | `GET /admin/realms/{realm}/users/{id}/groups` | view-users |
| `ListUserEvents` | `GET /admin/realms/{realm}/events` | view-events |
| `ListAdminEvents` | `GET /admin/realms/{realm}/admin-events` | view-events |

### 2.2 호출처 매트릭스

| # | 호출처 | 메서드 | 권한 요구 | 분류 | ADR-0020 충돌 |
| --- | --- | --- | --- | --- | --- |
| 1 | `backend-core/main.go:499` `seedLocalAdmin` | `CreateIdentity` + `FindIdentityByUserID` | **manage-users** + view-users | dev/test PoC seed (`test/test`) | ⚠️ |
| 2 | `backend-core/internal/httpapi/organization.go:435` `POST /api/v1/users` password 분기 | `CreateIdentity` | **manage-users** | DevHub `/users` API password handling | ⛔ 결정 A 우회 경로 |
| 3 | `backend-core/internal/httpapi/lazy_auto_create.go` `authenticateActor` lazy auto-create (PR #239) | — (DB 만 사용) | 없음 | ADR-0020 정공법 (결정 A) | ✅ |
| 4 | `backend-core/internal/audit/keycloak_event_puller.go` (PR #189..193) | `ListUserEvents` + `ListAdminEvents` | view-events | event listener (결정 C) | ✅ |
| 5 | `backend-core/internal/audit/user_sync.go` (PR #241) | `GetUserDetails` + `GetUserGroups` | view-users | event listener users sync (결정 C) | ✅ |

**핵심 발견**:
- 호출처 1 (`seedLocalAdmin`): dev/test 부트스트랩 전용. 운영 환경에서는 `keycloak-realm.json` realm import 가 admin seed 책임을 가져야 함. 운영자 환경 (사내 Keycloak) 은 realm-export.json + admin console reset password 가 정공법.
- 호출처 2 (`POST /api/v1/users` password 분기): `req.Password != ""` 시 Keycloak identity 도 함께 생성. ADR-0020 결정 A 와 직접 충돌 (DevHub 가 Keycloak admin 의 manage-users 역할 우회). 호출 plumbing 자체는 PR #239 의 `/api/v1/accounts/*` 폐기로 줄었지만, `/api/v1/users` 가 동일 우회 경로를 유지.
- 호출처 3-5: 모두 read-only 또는 DB only → `manage-users` 미요구.

### 2.3 Keycloak realm role 부여 현황 (사내 운영자가 확인 동반)

- 가정: 사내 Keycloak `devhub-backend` client 의 service account 가 `realm-management` client 의 `manage-users` + `view-users` + `view-events` composite role 보유.
- 목표: `manage-users` 제거 + `view-users` + `view-events` 유지 (sub-carve C event listener 동작 보장).

## 3. 옵션 비교

### 옵션 A — 호출처 전면 제거 (정공법, 단일 PR)

**작업 범위**:
- `seedLocalAdmin` — Keycloak `CreateIdentity` 호출 제거. 사용자 부트스트랩은 `keycloak-realm.json` realm import 에 `test` admin user pre-defined 으로 대체. e2e/dev 가이드 갱신.
- `POST /api/v1/users` password 분기 — 제거. `req.Password` 필드 폐기 또는 `400 Bad Request` 응답. frontend `account.service.ts` (sub-carve B-frontend) 동반 carve 필요.
- `KeycloakAdminClient.CreateIdentity` / `UpdateIdentityPassword` / `SetIdentityState` / `DeleteIdentity` 4 method 제거 (dead code).
- `IdentityAdmin` interface 의 write method 4건 제거.
- `idpAdminFake` test mock 정리.
- `keycloak_admin_client_test.go` 의 write method test 제거.

**위험**:
- dev/e2e 환경 부트스트랩 부서질 수 있음 — realm-export.json + Playwright global-setup 갱신 필요 (codex 가 PR #237/#202 에서 e2e 부분 정합 중).
- `POST /api/v1/users` password 분기 제거가 backward compat 위반 — 외부 caller 없으면 즉시 가능, frontend 만 caller 이므로 같은 PR 에서 frontend 동반 정리 권장.
- 변경 분량 큼 (단일 PR ~10-15 file 변경 + e2e/dev 가이드 정합).

**장점**:
- ADR-0020 결정 A 정공법 — `/api/v1/users` 우회 경로 완전 제거.
- 단계 분리 없음 — 한 번에 종결.

### 옵션 B — ENV gate (운영 모드 분기, 단계적 폐기, 2 sprint 분할)

**Phase 1 (본 sprint, backend only)**:
- `DEVHUB_ALLOW_USER_PASSWORD_PROVISIONING` env 신규 (default=`false`).
- `seedLocalAdmin` + `POST /api/v1/users` password 분기 — env=true 시만 활성화.
- env=false 시 (prod default): `seedLocalAdmin` skip + `POST /api/v1/users` password 분기 `503 Service Unavailable` ("user password provisioning disabled — use Keycloak admin console").
- 운영자가 service account 에서 `manage-users` 제거 가능 (env=false 환경).

**Phase 2 (사내 운영 동반, 별도 carve)**:
- 옵션 A 의 전면 호출처 제거.
- frontend `account.service.ts` 정리 (sub-carve B-frontend 와 결합).
- `KeycloakAdminClient` write method 4건 제거.

**위험**:
- 코드 복잡도 증가 (env 분기). 다만 backward compat 보장.
- 운영자가 env 설정을 잘못하면 dev 모드에서도 부트스트랩 실패.

**장점**:
- backward compat — dev 환경 즉시 영향 없음.
- 단계적 폐기 — 사내 운영팀이 Phase 2 진입 전 manage-users 제거 → 운영 검증 → 호출처 제거.
- frontend cleanup (Gemini sub-carve B-frontend) 와 자연 합류.

### 옵션 C — manage-users 유지 + Phase 2 carve (docs only)

**작업 범위**:
- design doc + 운영 SOP 만 작성. 코드 변경 없음.
- service account 권한 축소는 Phase 2 (별도 sprint) 로 완전 연기.
- ADR-0020 §4.1 sub-carve E "carve out continued" 로 표시.

**위험**:
- ADR-0020 sub-carve E 의 정신 (실제 권한 축소) 미달성 — 단순 docs only.
- 호출처 매트릭스 식별만 가치 있고, 실 효과 0.

**장점**:
- codex `#238` 와 즉시 분리 가능 (docs 만 작성).
- 분량 최소.

## 4. 권장 — 옵션 A (전면 호출처 제거, 단일 PR) **— 채택**

**최종 결정** (사용자 확인, 2026-05-20 sprint -n):
- 옵션 A 정공법 채택. 단계 분리 없이 본 sprint 에서 호출처 전면 제거.
- realm-export.json 갱신 미동반 — codex `#238` 와 충돌 회피. dev 운영자가 keycloak admin console 1회 시드.
- frontend `account.service.ts` 폐기 (sub-carve B-frontend) 는 Gemini 별도 sprint 자연 follow-up.

**작업 단위** (본 sprint, 5 commit):
1. **Commit 1**: `backend-core/internal/httpapi/organization.go` 의 `POST /api/v1/users` `req.Password` 분기 제거 + `createUserRequest.Password` field 폐기 + `audit_logs.details` 의 dead key `kratos_id` 제거.
2. **Commit 2**: `backend-core/main.go` 의 `seedLocalAdmin` 함수 + 호출 + `seedOrgStore` interface 완전 제거. `main_test.go` 전체 삭제 (seedLocalAdmin 전용 test 3건 + `idpAdminFake` + `orgStoreFake`).
3. **Commit 3**: `KeycloakAdminClient` 의 4 write method (`CreateIdentity` / `UpdateIdentityPassword` / `SetIdentityState` / `DeleteIdentity`) 제거 + dead helper `keycloakIDFromLocation` 제거. `IdentityAdmin` interface 의 write method 4건 제거 (`FindIdentityByUserID` 만 유지, view-users role 만 요구). `MockIdentityAdmin` + `keycloak_admin_client_test.go` 정리.
4. **Commit 4** (본): design doc 옵션 A 채택 갱신 + ADR-0020 §4.1 sub-carve E done + traceability §4 row.
5. **Commit 5**: memory state + sprint -n state.

**Phase 2 (사내 운영 동반)**:
- 운영자: service account 에서 `manage-users` realm role 제거 → 운영 1~2주 검증 (§5 SOP).
- frontend (Gemini sub-carve B-frontend): `account.service.ts` 폐기 + `/admin/settings/users` page 정리 + e2e TC-ACC-* 갱신.
- realm-export.json (codex `#238` 머지 후 별도 carve): `test` admin user pre-defined 추가 → seedLocalAdmin 폐기 후 dev 부트스트랩 자동화.

## 5. 운영 SOP — service account 권한 축소 절차

### 5.1 사전 확인

- DevHub backend 버전이 sprint -n (PR TBD, 본 문서) 이후 — `KeycloakAdminClient` write methods + `seedLocalAdmin` 모두 제거된 상태.
- backend log 에 `[seedLocalAdmin]` 키워드 0건 (함수 자체가 제거됨).
- `POST /api/v1/users` 호출 시 `password` field 가 응답 schema 에 없음 (silently ignored — 또는 strict 모드면 400).
- backend metric `devhub_keycloak_user_sync_total` + `devhub_keycloak_events_processed_total` (PR #241, PR #189..193) 정상 emit — read API 만 호출되는지 확인.

### 5.2 Keycloak 운영자 작업

1. Keycloak admin console 진입 → `Clients` → `devhub-backend` → `Service Account Roles` 탭.
2. **Assigned Roles** 에서 `realm-management` client 의 `manage-users` composite role 제거.
3. **유지해야 할 role**:
   - `realm-management` / `view-users` (event listener users sync — PR #241)
   - `realm-management` / `view-events` (event listener — PR #189..193)
   - `realm-management` / `view-realm` (issuer + JWKS lookup)
4. `View Effective Roles` 에서 `manage-users` 가 더 이상 없음을 확인.

### 5.3 검증 (운영 1~2주)

- backend metric `devhub_keycloak_events_processed_total` + `devhub_keycloak_user_sync_total` 정상 증가 — event listener 동작 확인.
- backend log 에 `Keycloak admin API` 4xx/5xx 에러 없음.
- 사용자 로그인 정상 (lazy auto-create + JWKS verification).
- `GET /api/v1/users` 정상 — DevHub `users` 테이블 read.

### 5.4 회복 절차 (롤백)

옵션 A 정공법 채택 후 backend 코드 자체가 manage-users 호출처를 갖지 않으므로 운영자가 manage-users 를 service account 에 재부여해도 backend 가 사용할 일이 없다. 만약 dev 환경에서 `seedLocalAdmin` 동작이 필요하면 git revert 또는 별도 dev 부트 script 로 복구.

## 6. ADR-0020 §4.1 sub-carve 표 갱신 (본 sprint -n PR 후)

| sub-carve | 결정 | 상태 |
| --- | --- | --- |
| E (호출처 전면 제거) | 옵션 A 채택 | ✅ done (sprint -n, PR TBD) |
| (사내 운영자 후속) | service account 에서 `manage-users` realm role 제거 | carve (운영팀 책임) |
| B-frontend (Gemini) | `account.service.ts` 폐기 + admin/settings/users page 정리 | carve (Gemini 별도 sprint) |
| (codex #238 머지 후 carve) | realm-export.json `test` admin user pre-defined 추가 | carve (codex `#238` 머지 후) |

## 7. 다음 sprint 권장 진입

- **사내 운영자**: §5.2 절차에 따라 `devhub-backend` service account 에서 `manage-users` realm role 제거 → §5.3 검증 1~2주.
- **Gemini sub-carve B-frontend**: `frontend/lib/services/account.service.ts` 폐기 + `/admin/settings/users` page admin actions 제거 + e2e TC-ACC-* 갱신.
- **codex `#238` 머지 후 carve**: `infra/idp/keycloak-realm.json` 의 `users` 배열에 `test` admin user pre-defined 추가 → dev 부트스트랩 자동화 (실 운영자가 keycloak admin console 수동 시드 부담 제거).

## 8. 변경 이력

| 일자 | sprint | 변경 |
| --- | --- | --- |
| 2026-05-20 | claude/work_260520-n-214-service-account-min-role | 본 design doc 신규 (현황 매트릭스 + 옵션 A/B/C 비교 + 옵션 A 정공법 채택 + 운영 SOP + Phase 2 carve). 본 sprint 의 5 commit (organization.go password 분기 제거 + seedLocalAdmin 완전 제거 + KeycloakAdminClient write methods + IdentityAdmin interface 정리 + docs/memory) 가 호출처 전면 제거 완료. |
