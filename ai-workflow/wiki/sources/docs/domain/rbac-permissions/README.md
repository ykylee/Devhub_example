---
title: README
type: source
tags: [domain, README.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/rbac-permissions/README.md]
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

# rbac-permissions 도메인

- 문서 목적: `rbac-permissions` 도메인의 SDLC 진입점.
- 범위: 역할(Role), 자원(Resource), 액션(Action) 매트릭스를 기반으로 한 다차원 접근 제어.
- 상태: draft (Phase 1 골격)
- 최종 수정일: 2026-05-29
- 관련 문서: [code-taxonomy.md §2.1.3](../../governance/code-taxonomy.md)

## 1. 도메인 정의

> 역할(Role), 자원(Resource), 액션(Action) 매트릭스를 기반으로 한 다차원 접근 제어를 수행한다. ([code-taxonomy.md §2.1.3](../../governance/code-taxonomy.md))

PermissionCache 메모리 캐시 + DB-backed RBAC 정책 + Row-scoping (ADR-0011) 으로 다층 접근 제어를 구현한다.

## 2. 4 계층 모듈 매핑

| 계층 | Backend | Frontend |
|---|---|---|
| view | `backend-core/internal/domain/rbac-permissions/view/` (`permissions.go`, `rbac.go`, `authz.go`, `handler.go`) | `frontend/app/admin/settings/permissions/`, `frontend/domain/rbac-permissions/view/PermissionEditor.tsx`, `PermissionMatrix.tsx` |
| service | `backend-core/internal/domain/rbac-permissions/view/permissions.go` 내 `PermissionCache` | `frontend/domain/rbac-permissions/service/rbac.service.ts` |
| repository | `backend-core/internal/domain/rbac-permissions/repository/postgres_rbac.go` | — |
| schema | `domain/rbac.go` (역할 + 권한 매핑), DB: `rbac_policies` (000005/000018/000021/000024/000026) | `frontend/lib/services/rbac.types.ts` |

## 3. SDLC 문서 link

| 단계 | 위치 | 상태 |
|---|---|---|
| REQ | [`./requirements.md`](./requirements.md) | active (Phase 3 split, 2026-05-29) |
| ARCH | [`./architecture.md`](./architecture.md) | active (Phase 3) |
| API | [`./api.md`](./api.md) | active (Phase 3) |
| TC | `./test_cases.md` | active (2026-06-02 RBAC 고도화 TC 카탈로그) |
| Design | `./keycloak_groups_mapping.md` | planned (Phase 2 — `docs/domain/rbac-permissions/keycloak_groups_mapping.md`) |

## 4. 관련 ADR

- ADR-0002 (RBAC policy edit API)
- ADR-0007 (RBAC enforcement)
- ADR-0011 (Row-scoping)

## 5. cross-cutting 참조

- `docs/backend_api_contract.md` §12 (RBAC API)
- `docs/architecture.md` §13 (RBAC architecture)

## 6. E2E spec

- `frontend/tests/e2e/admin-permissions.spec.ts`
- `frontend/tests/e2e/rbac-routes.spec.ts`
