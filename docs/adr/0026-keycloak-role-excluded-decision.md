# ADR-0026: Keycloak Role 무시 — DevHub 내부 Role 단독 사용

- **상태**: accepted (2026-06-02)
- **영향 도메인**: `rbac-permissions`, `auth-session`, `audit-ops`, `infra-idp`
- **결정 근거 브랜치**: `deepseek/work_260602`
- **관련 ADR**: [ADR-0019 (Keycloak-only IdP)](0019-keycloak-only-idp.md), [ADR-0020 (Account/User Management Boundary)](0020-account-user-management-boundary.md), [ADR-0021 (Onboarding Self-Service)](0021-onboarding-self-service-unit-selection.md)
- **관련 문서**: [RBAC 요구사항](../domain/rbac-permissions/requirements.md), [Keycloak Groups Mapping](../domain/rbac-permissions/keycloak_groups_mapping.md), [통합 테스트 리포트](../planning/integrated_test_report_20260601.md)
- **대체 방안**: Keycloak Admin REST API polling + group composite role → `users.role` 동기화 (PR #460 BUG-02, P1-1 carve)

---

## Context

PR #460 통합 테스트(2026-06-01)에서 **BUG-02** 발견:

> Keycloak Admin API로 사용자에게 `devhub_role` attribute를 지정하거나 Keycloak realm role을 할당해도 DevHub DB `users.role`에 반영되지 않는다. JWT access token에도 role claim이 누락되어 RBAC `enforceRoutePermission`이 항상 부재 처리 — RBAC 무력화.

초기 설계에서는 ADR-0020 sub-carve C (sprint -k)로 Keycloak Admin REST API event polling + `DEVHUB_KEYCLOAK_EVENT_LISTENER_ENABLED` 환경변수 기반 `SyncUserMembership`이 이 role sync를 담당할 예정이었다. 그러나 이 접근은 세 가지 근본 문제가 있었다:

1. **Race condition**: Keycloak Admin UI에서 role 변경 직후 event poller가 sync하기 전까지 DevHub는 stale role로 동작. Drift fail-closed(PR #461)가 요청을 403 차단하지만, 실제로는 사용자 경험 저하.
2. **Keycloak을 이중 source of truth로 만듦**: 운영자가 role 변경 시 Keycloak Admin UI + DevHub Admin UI 두 곳을 모두 조작해야 하는 혼란. 어느 한쪽만 변경하면 drift.
3. **Group → role 매핑의 불완전성**: Keycloak group name 기반 role 매핑(`devhub-managers` → `team_manager`)은 naming convention에 의존하며, custom role 시나리오에서 확장 불가.

## Decision

**Keycloak이 가진 모든 role 정보(realm role, group membership, `devhub_role` attribute, JWT role claim)를 DevHub RBAC의 source of truth로 사용하지 않는다.**

대신:
- **유일한 source of truth**: DevHub DB `users.role` 컬럼 (FK → `rbac_policies.role_id`)
- **Role 변경 경로**: DevHub Admin UI (`/admin/settings/users`) 또는 DevHub API (`PATCH /api/v1/users/:id` / `PUT /api/v1/organization/units/:id/members`)
- **Keycloak의 역할**: 오직 **OIDC 인증(login)**과 **사용자 프로필 정보(email, display_name)** 제공으로 한정

### 세부 규칙

| 영역 | 규칙 | 근거 |
|------|------|------|
| JWT role claim | 추출하지만 **DB `users.role`이 존재하면 DB 값으로 override** | DB miss 시 fallback (pre-onboarding 계정) |
| Keycloak realm role | 완전히 **무시** | DevHub는 realm role을 읽지 않음 |
| Keycloak group membership | 완전히 **무시** | `user_sync.go` `SyncUserMembership` 삭제 완료 |
| `devhub_role` attribute | **어떤 Go 코드에서도 읽지 않음** | 해당 claim 존재 자체가 0건 |
| Keycloak event listener | **audit_logs 기록용으로만 유지** | Role sync callback 제거 완료 |

## Consequences

### 즉시 적용된 변경 (commit `4060f90`)

| 파일 | 조치 |
|------|------|
| `user_sync.go` | **삭제** — `SyncUserProfile`, `SyncUserMembership`, `MarkUserDeactivated`, `pickHighestPriorityRole`, `groupNameToRole`, `composeDisplayName`, `ParseIdentityIDFromResourcePath` |
| `user_sync_test.go` | **삭제** |
| `user_sync_integration_test.go` | **삭제** |
| `keycloak_event_puller.go` | `UserSyncCallback` 타입, `KeycloakEventPullerOptions.UserSync` 필드, `classifyAdminEventForSync` 함수 제거 |
| `keycloak_event_puller_test.go` | user_sync 관련 3개 테스트 함수 제거 |
| `metrics.go` | `ObserveUserSync`, `ObserveUserSyncError`, `ObserveUserSyncLag` + 관련 메트릭 변수 제거 |
| `main.go` | user_sync callback wiring 제거 |

### 유지되는 코드

| 코드 | 사유 |
|------|------|
| `extractKeycloakRole` (`keycloak_verifier.go:428`) | DB miss 시 fallback. `authenticateActor`가 DB row를 찾으면 즉시 override하므로 안전 |
| `authenticateActor` drift 감지 (`auth.go:197-210`) | Token role ≠ DB role 감지 → 403 `role_sync_required`. 역할 불일치를 명시적으로 차단 |
| `DEVHUB_KEYCLOAK_EVENT_LISTENER_ENABLED` | Keycloak audit event → `audit_logs` 기록용. Role sync와 무관 |
| `POST /api/v1/auth/logout` | 세션 정리 전용. Role 변경 없음 |

### 향후 Role 관리 프로세스

```
사용자 Role 변경 시나리오
┌─────────────────────────────────────────────────────┐
│  1. 운영자, DevHub Admin UI 접속                      │
│     (/admin/settings/users)                          │
│  2. 대상 사용자 선택 → Role 필드 변경                  │
│  3. PATCH /api/v1/users/:id → users.role UPDATE      │
│  4. 다음 요청 시 authenticateActor가 새 role 감지      │
│  5. EnforceRoutePermission이 새 role로 RBAC 적용      │
└─────────────────────────────────────────────────────┘
```

변경은 **DevHub Admin UI** 또는 **DevHub API**를 통해서만 이루어지며, Keycloak Admin UI에서의 role 조작은 DevHub에 **영향을 주지 않는다**.
