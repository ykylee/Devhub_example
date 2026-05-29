# rbac-permissions 도메인 아키텍처

- 문서 목적: RBAC PermissionCache + route → (resource, action) 매핑 + row-scoping + Keycloak group 매핑의 아키텍처를 정의한다.
- 범위: master `docs/architecture.md` §6.3 (RBAC 단계화) 의 상세화. cross-cutting 보안 baseline 은 master §6.1/§6.4 참조, Keycloak event sync 는 audit-ops 도메인 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, 아키텍처 검토자.
- 상태: draft (Phase 3 split)
- 최종 수정일: 2026-05-29
- 관련 문서: [도메인 README](./README.md), [requirements.md](./requirements.md), [api.md](./api.md), [keycloak_groups_mapping.md](./keycloak_groups_mapping.md), [master architecture](../../architecture.md), [ADR-0002](../../adr/0002-rbac-policy-edit-api.md), [ADR-0007](../../adr/0007-rbac-enforcement.md), [ADR-0011](../../adr/0011-rbac-row-scoping.md), [ADR-0020](../../adr/0020-account-user-management-boundary.md)

## 1. 컴포넌트 (ARCH-RBAC-01)

```
┌──────────────────────────────────────────────┐
│  Go Core HTTP 미들웨어 체인                  │
│  ├── authenticateActor                       │
│  ├── onboardingGate (onboarding 도메인)       │
│  └── requirePermission                       │
│      ├── routePermissionTable lookup         │
│      ├── PermissionCache.Allows(role, R, A)  │
│      ├── deny-by-default (매핑 누락)         │
│      └── row-level helper enforceRowOwnership│
└──────────┬───────────────────────────────────┘
           │
           ▼
┌──────────────────────────────────────────────┐
│  PermissionCache (in-process)                │
│  ├── rbac_policies store hydrate             │
│  ├── PUT 시 reload                            │
│  └── multi-instance reload pending           │
└──────────┬───────────────────────────────────┘
           ▼
┌──────────────────────────────────────────────┐
│  PostgreSQL                                  │
│  └── rbac_policies (000005/000018/000021/    │
│                     000024/000026)            │
└──────────────────────────────────────────────┘
```

## 2. 매트릭스 모델 (ARCH-RBAC-02)

- **Role**: 시스템 정의 3종 (immutable) + 사용자 정의(`custom-{slug}` prefix). seed 시 `system_admin`/`manager`/`developer` 영속화.
- **Resource**: 11종 (`infrastructure`, `pipelines`, `organization`, `security`, `audit`, `applications`, `application_repositories`, `projects`, `scm_providers`, `dev_requests`, `dev_request_intake_tokens`).
- **Action**: `view | create | edit | delete` 4축.
- **Audit append-only invariant**: `audit.{create,edit,delete}` 는 모든 role 에서 false 강제 (store 검증).

## 3. Enforcement (ARCH-RBAC-03)

- `requirePermission` 미들웨어가 `routePermissionTable` 을 source-of-truth 로 enforcement.
- 매핑 누락 라우트는 deny-by-default — `403 Forbidden` + `auth.policy_unmapped` audit + payload `{actor_role, method}` + monitoring.
- enforcement 결과 거부 audit 는 `auth.role_denied` (matrix false) + `auth.row_denied` (row helper false) + `auth.policy_unmapped` (매핑 자체 없음) 3종.

## 4. Row-scoping (ARCH-RBAC-04, ADR-0011)

- `enforceRowOwnership(c *gin.Context, ownerUserID string, allowedRoles ...string) bool` helper.
- allow 순서:
  1. actor.role == `system_admin` → allow.
  2. actor.role ∈ allowedRoles → allow.
  3. actor.login == ownerUserID → allow.
  4. otherwise → 403 + `code=auth_row_denied` + audit `auth.row_denied`.
- 본 helper 는 dev-request 도메인의 row-level 권한 enforcement 1순위 진입점이며, application-lifecycle/onboarding 등에서도 재사용된다.

## 5. PermissionCache (ARCH-RBAC-05)

- in-process LRU + key=role → matrix snapshot.
- write API(PUT /rbac/policies, POST /rbac/policies, DELETE /rbac/policies/:id) 머지 시 same-process reload.
- multi-instance 환경 일관성은 pub/sub 또는 polling 으로 보강 carve out — 운영 phase 진입 시.

## 6. Keycloak group → RBAC role 매핑 (ARCH-RBAC-06)

- ADR-0020 결정 D: DevHub 가 직접 노출하던 subject role 변경 endpoint (API-30/API-31) 는 **폐기**. role 변경 source 는 Keycloak admin console (group membership).
- DevHub `users.role` 컬럼은 audit-ops 도메인의 Keycloak event listener (push SPI + poll cron) 가 자동 sync 한다.
- group → role 매핑 표 + 운영 SOP 는 [`./keycloak_groups_mapping.md`](./keycloak_groups_mapping.md) 참조.

## 7. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/architecture.md` §6.3 (RBAC 단계화) 및 `docs/backend_api_contract.md` §12.0/§12.10 의 아키텍처-관련 내용을 도메인 sub-document 로 재집합. ID는 ARCH-RBAC-01..06 도메인 발급 (master 의 RBAC 전용 ARCH ID 없음). |
