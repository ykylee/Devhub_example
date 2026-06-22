---
type: entity
status: active
last_ingested_from: ai-workflow/wiki/entities/keycloak-iam.md
related_pages: [sources/keycloak-iam]
created: 2026-06-15
updated: 2026-06-15
last_touched: 2026-06-22T06:03:34Z
git_commit: fb3894f7
git_branch: chore/260622-wiki-drift-cleanup-3
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
mirror_dirty: |
---

# Keycloak IAM (L2 dense, in-repo)

> L1 SSOT: `ai-workflow/wiki/entities/keycloak-iam.md`
> 본 L2 derived view 는 in-repo retrieval 용 압축 요약.

## TL;DR

DevHub 의 1차 IdP = Keycloak OIDC. 4-layer 통합: (1) runtime injection + (2) OIDC auth flow + (3) RBAC 4 role + (4) event audit. ADR-0019 (Keycloak only IdP) + ADR-0020 (계정/사용자 책임 경계) + ADR-0030 (port interface) + ADR-0031 (event audit) 의 4 ADR 정합.

## 1. 4 layer

- **Runtime injection** (`main.go` + `saovae_stub.go`): `DEVHUB_BUILD_TIER` = `=internal|saovae|ci`. 4 port + webhook handler + default wiring.
- **OIDC auth flow** (`auth-session/view/{auth,handler}.go`): BearerTokenVerifier + IdentityAdmin + OIDCLogoutClient. Canonical = `integration/ports.go`.
- **RBAC** (`rbac.go` + `rbac-permissions/view/rbac.go`): 4 role (system_admin/developer/manager/team_manager) + rbac_policies seeded + PermissionCache LISTEN/NOTIFY.
- **Event audit** (`audit-ops/view/keycloak_events_webhook.go` + `service/keycloak_event_puller.go`): webhook + 1분 polling + audit_logs X-Request-ID propagation.

## 2. ADR

- **ADR-0019**: Keycloak only IdP (1차 IdP 강제)
- **ADR-0020**: 계정/사용자 책임 경계 (sub-carve Phase 1/2/3)
- **ADR-0030**: SSO integrations + auth session port (4 port + 4 type + 3 error)
- **ADR-0031**: Keycloak event audit (puller + webhook)

## 3. Tier 분리

- **사외 (in-repo)**: ADR/governance/setup 문서, e2e/frontend code, saovae stub
- **사내 (별도 SCM)**: `infra/idp/keycloak-realm.{ci,prod}.json`, `docker-compose.{local,test,deploy,colima}.yml`, `scripts/setup-keycloak.sh`, `.env.{deploy,prod,test}` 의 `DEVHUB_KEYCLOAK_*` / `KEYCLOAK_URL` / `KC_INTERNAL_URL`

## 4. 운영 정공법

- **Service account minimum role** (ADR-0031 follow-up): `realm-management: view-events` + `view-users`
- **OIDC federation** (`sso_federation.md`): SAML/OIDC bridge, IdP 간 user migration
- **Offboarding immediacy** (`offboarding_immediacy.md`): Keycloak user disable → audit + session invalidate within 1분
- **Failover** (`failover.md`): 1차 Keycloak 장애 → saovae_stub graceful degradation

## 5. Follow-up

- ADR-0007 (PermissionCache LISTEN/NOTIFY) backend 구현
- sub-carve B (account/user management 폐기) v0.1.0 release
- X-2 inbound source routing 의 RBAC 통합
- ProviderHealthWidget (X-1, API-107) v0.1.1 후속
