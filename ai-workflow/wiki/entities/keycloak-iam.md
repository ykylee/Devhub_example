---
type: entity
status: active
last_ingested_from: docs/adr/0019-keycloak-only-idp.md + docs/setup/keycloak_operations.md + docs/infrastructure/keycloak-idp/
related_pages: [concepts/devhub-overview, patterns/in-repo-redirect, entities/gitea-scm, entities/postgres-store]
created: 2026-06-15
updated: 2026-06-15
active_since: 2026-06-15
active_reason: "DevHub v0.1.0 의 1차 IdP. v0.7.17 vendor import 후에도 본 저장소 의 *in-repo* 운영 정합"
git_commit: 046e0c81
git_branch: chore/260622-wiki-drift-cleanup-4
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T06:22:35Z
mirror_dirty: |
---

# Keycloak IAM (L1 entity, in-repo)

## TL;DR

DevHub 의 1차 IdP (Identity Provider) = Keycloak OIDC. `sso-integrations/keycloak/saovae_stub.go` (4 port stub) + `integration/ports.go` (4 port interface + 4 type alias + 3 sentinel error) + `audit-ops/view/keycloak_events_webhook.go` (event webhook handler) + `audit-ops/service/keycloak_event_puller.go` (1분 polling) 의 4-layer 통합. 사내 1차 IdP 강제 (ADR-0019) + RBAC 4 role + OIDC federation + service account minimum role.

## 1. 통합 layer (4 layer)

### 1.1 Runtime injection (`backend-core/main.go` + `sso-integrations/keycloak/saovae_stub.go`)
- `DEVHUB_BUILD_TIER` env var: `=internal` (real Keycloak) | `=saovae` (4 port stub) | `=ci` (test double)
- saovae stub = 4 port + webhook handler + default wiring (sprint -a follow-up PR #539)
- 4 port interface (`integration/ports.go`): BearerTokenVerifier / OIDCLogoutClient / IdentityAdmin / KeycloakEventListener
- 4 type alias: BearerToken / OIDCSession / KeycloakIdentity / KeycloakEvent
- 3 sentinel error: ErrTokenInvalid / ErrSessionExpired / ErrIdentityNotFound

### 1.2 OIDC auth flow (`auth-session/view/auth.go` + `auth-session/view/handler.go`)
- BearerTokenVerifier interface (deprecation) — canonical = `integration/ports.go`
- IdentityAdmin + OIDCLogoutClient interface — canonical = `integration/ports.go`
- `/api/v0-1/auth/*` (OIDC login + logout + token refresh)

### 1.3 RBAC (rbac.go + rbac-permissions/view/rbac.go)
- 4 role: system_admin / developer / manager / team_manager
- rbac_policies seeded (initial migration)
- PermissionCache LISTEN/NOTIFY (future ADR-0007)
- 4 role-conditional landing + pathRequiresSystemAdmin

### 1.4 Event audit (audit-ops/view/keycloak_events_webhook.go + audit-ops/service/keycloak_event_puller.go)
- Keycloak event webhook handler (event listener type assertion + webhook routing)
- Keycloak event puller (1분 polling + audit_logs 통합, PR #189~#193 + #241)
- audit actor enrichment (source_ip + request_id + source_type + X-Request-ID)
- audit_logs X-Request-ID propagation (repository audit_logs.go)

## 2. ADR (Architecture Decision Record)

- **ADR-0019**: Keycloak only IdP (1차 IdP 강제) — 본 저장소 의 1차 source-of-truth
- **ADR-0020**: 계정/사용자 책임 경계 (sub-carve Phase 1/2/3)
- **ADR-0030**: SSO integrations and auth session port (4 port + 4 type + 3 error)
- **ADR-0031**: Keycloak event audit (event puller + webhook 통합)

## 3. 사내 vs 사외 (2-tier)

- **사외 (in-repo)**: ADR/governance/setup 문서, e2e/frontend code, saovae stub
- **사내 (별도 SCM)**: 
  - `infra/idp/keycloak-realm.ci.json` (CI 용 Keycloak realm JSON)
  - `infra/idp/keycloak-realm.prod.json` (prod 용 realm)
  - `docker-compose.{local,test,deploy,colima}.yml` (Keycloak 컨테이너 셋업)
  - `scripts/setup-keycloak.sh` (sprint setup)
  - `.env.deploy` / `.env.prod` / `.env.test` 의 `DEVHUB_KEYCLOAK_*` / `KEYCLOAK_URL` / `KC_INTERNAL_URL` 등
- **Tier 분리 검증**: `bash scripts/check-tier-separation.sh` PASS 예상 (PR 작성 시 self-check)

## 4. 운영 정공법 (operations)

- **Service account minimum role** (ADR-0031 follow-up): `infra/idp/keycloak-realm.json` 의 `service_account_min_role` = `realm-management: view-events` (event puller 용) + `view-users` (IdentityAdmin 용) — minimum privilege
- **OIDC federation**: `sso_federation.md` 의 SAML/OIDC bridge 패턴 — IdP 간 user migration
- **Offboarding immediacy** (`offboarding_immediacy.md`): Keycloak user disable event → audit log + session invalidate within 1분
- **Failover** (`failover.md`): 1차 Keycloak 장애 시 saovae_stub 의 graceful degradation + 503 → 200 fallback (auto_route.go 의 Route() 의 graceful pattern)

## 5. follow-up (별도 PR)

- ADR-0007 (PermissionCache LISTEN/NOTIFY) 의 backend 구현 (현재 docs only)
- sub-carve B (account/user management 폐기) 의 v0.1.0 release 시점
- X-2 inbound source routing (Gitea/Jira 4 provider) 의 RBAC 통합 (PR #586 5차)
- ProviderHealthWidget (X-1, API-107 별도 carve) 의 v0.1.1 후속 sprint
