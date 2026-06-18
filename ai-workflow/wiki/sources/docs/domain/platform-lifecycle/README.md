---
title: README
type: source
tags: [domain, README.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/platform-lifecycle/README.md]
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

# platform-lifecycle 도메인

- 문서 목적: `platform-lifecycle` 도메인의 SDLC 진입점.
- 범위: 핵심 비즈니스 엔티티 Platform + Project 의 CRUD, 상태 전이 머신, 롤업 요약.
- 상태: draft (Phase 1 골격)
- 최종 수정일: 2026-05-29
- 관련 문서: [code-taxonomy.md §2.1.6](../../governance/code-taxonomy.md)

## 1. 도메인 정의

> 핵심 비즈니스 엔티티인 Platform과 Project의 CRUD, 상태 전이 머신, 그리고 롤업 요약 데이터 생성을 담당한다. ([code-taxonomy.md §2.1.6](../../governance/code-taxonomy.md))

## 2. 4 계층 모듈 매핑

| 계층 | Backend | Frontend |
|---|---|---|
| view | `backend-core/internal/domain/platform-lifecycle/view/` (`applications.go`, `projects.go`, `application_rollup.go`, `handler.go`) | `frontend/app/platforms/`, `frontend/app/projects/`, `frontend/domain/platform-lifecycle/view/{ApplicationCreationModal,ApplicationTable,ProjectCreationModal,ProjectTable}.tsx` |
| service | Platform/Project 생명주기 상태 머신 (활성/유휴/폐기), 롤업 계산 | `frontend/domain/platform-lifecycle/service/{application,project}.service.ts` |
| repository | `backend-core/internal/domain/platform-lifecycle/repository/` (`applications.go`, `projects.go`, `repository.go`) | — |
| schema | `domain/application.go`, DB: `applications` (000013), `projects` (000015), `project_members`, `repo_ops_snapshots` (000017) | (frontend 내장) |

## 3. SDLC 문서 link

| 단계 | 위치 | 상태 |
|---|---|---|
| REQ | [`./requirements.md`](./requirements.md) | active (Phase 3 split, 2026-05-29) |
| ARCH | [`./architecture.md`](./architecture.md) | active (Phase 3) |
| API | [`./api.md`](./api.md) | active (Phase 3) |
| TC | `./test_cases.md` | planned (Phase 2) |
| Concept | `./project_concept.md` | planned (Phase 2 — `docs/domain/platform-lifecycle/project_concept.md`) |
| Concept | `./dashboard_concept.md` | planned (Phase 2 — `docs/domain/platform-lifecycle/dashboard_concept.md`) |

## 4. 관련 ADR

- ADR-0011 (Row-scoping)
- ADR-0014 (Platform/Project lifecycle)

## 5. cross-cutting 참조

- `docs/backend_api_contract.md` §13 (Platform/Repository/Project API)
- `docs/architecture.md` §11 (APPDASH)

## 6. E2E spec

- `frontend/tests/e2e/admin-applications.spec.ts`
- `frontend/tests/e2e/admin-projects.spec.ts`
