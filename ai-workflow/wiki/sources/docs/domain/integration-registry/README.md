---
title: README
type: source
tags: [domain, README.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/integration-registry/README.md]
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

# integration-registry 도메인

- 문서 목적: `integration-registry` 도메인의 SDLC 진입점.
- 범위: SCM/비-SCM (Jira, Confluence, Homelab 등) 연동 공급자와 바인딩 통합 관리, ProviderModal UI 공통 컴포넌트 소유.
- 상태: draft (Phase 1 골격)
- 최종 수정일: 2026-05-29
- 관련 문서: [code-taxonomy.md §2.1.9](../../governance/code-taxonomy.md)

## 1. 도메인 정의

> SCM(Gitea) 및 비-SCM(Jira, Confluence, Homelab 등) 연동 공급자와 바인딩을 통합 관리하며, **ProviderModal** UI 공통 컴포넌트의 소유권을 가집니다. ([code-taxonomy.md §2.1.9](../../governance/code-taxonomy.md))

`integrationcaps.ProviderHasCapability` 공용 helper (OR semantics) 가 도메인 간 capability 게이트의 single source of truth.

## 2. 4 계층 모듈 매핑

| 계층 | Backend | Frontend |
|---|---|---|
| view | `backend-core/internal/domain/integration-registry/view/` (`integration_registry.go`, `integrations.go`, `external_task_handler.go`, `handler.go`) | `frontend/app/admin/settings/{integrations,integration-bindings}/`, `frontend/domain/integration-registry/view/{ProviderModal,ProviderTable,BindingsTable,CreateBindingModal,EditBindingModal}.tsx` |
| service | 외부 Preset 매핑, Sync 작업 스케줄링 큐, Task Ingestion 라우팅 | `frontend/domain/integration-registry/service/{infra,integration-provider-presets}.ts` |
| repository | `backend-core/internal/domain/integration-registry/repository/` (`integration_registry.go`, `external_task_store.go`) | — |
| schema | DB: `integration_providers` (000028), `integration_bindings` (000040), `external_task_items` (000046) | (frontend 내장) |

## 3. SDLC 문서 link

| 단계 | 위치 | 상태 |
|---|---|---|
| REQ | [`./requirements.md`](./requirements.md) | active (Phase 3 split, 2026-05-29) |
| REQ (Task Ingestion) | [`./task_requirements.md`](./task_requirements.md) | active (Phase 3) |
| ARCH | [`./architecture.md`](./architecture.md) | active (Phase 3) |
| ARCH (Task Ingestion) | [`./task_architecture.md`](./task_architecture.md) | active (Phase 3) |
| API | [`./api.md`](./api.md) | active (Phase 3) |
| API (Task Ingestion) | [`./task_api.md`](./task_api.md) | active (Phase 3) |
| TC | `./test_cases.md` | planned (Phase 2 — `docs/domain/integration-registry/test_cases.md`) |
| Concept | `./external_system_integration_concept.md` | planned (Phase 2 — `docs/planning/`) |
| Spec | `./external_integration_capability_matrix.md` | planned (Phase 2 — `docs/planning/`) |
| Concept | `./task_ingestion_concept.md` | planned (Phase 2 — `docs/domain/integration-registry/task_ingestion_concept.md`) |

## 4. 관련 ADR

- ADR-0015 (HomeLab pull strategy)

## 5. cross-cutting 참조

- `docs/backend_api_contract.md` §15 (Integration API)
- `docs/architecture.md` §8 (External Integration)
- `backend-core/internal/shared/integrationcaps/` (공용 helper, PR #409)

## 6. E2E spec

- `frontend/tests/e2e/admin-integrations.spec.ts`
- `frontend/tests/e2e/admin-integration-bindings.spec.ts`
