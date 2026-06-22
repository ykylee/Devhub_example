---
title: README
type: source
tags: [domain, README.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/repository-integration/README.md]
git_commit: baf1cf24
git_branch: main
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T03:37:45Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# repository-integration 도메인

- 문서 목적: `repository-integration` 도메인의 SDLC 진입점.
- 범위: SCM 저장소와 프로젝트 연결, 가져오기(Import), 코드 자산 매핑.
- 상태: draft (Phase 1 골격)
- 최종 수정일: 2026-05-29
- 관련 문서: [code-taxonomy.md §2.1.7](../../governance/code-taxonomy.md)

## 1. 도메인 정의

> SCM 저장소를 프로젝트와 연결하고, 가져오기(Import) 및 코드 자산 매핑을 수행한다. ([code-taxonomy.md §2.1.7](../../governance/code-taxonomy.md))

## 2. 4 계층 모듈 매핑

| 계층 | Backend | Frontend |
|---|---|---|
| view | `backend-core/internal/domain/repository-integration/view/` (`integration_scm_repositories.go`, `handler.go`) | `frontend/app/repositories/`, `frontend/domain/repository-integration/view/{RepositoryLinkModal,RepositoryTable}.tsx` |
| service | SCM↔DevHub 프로젝트 맵핑 검증, 강제 동기화 규칙 | `frontend/domain/repository-integration/service/repository.service.ts` |
| repository | `platform-lifecycle/repository/applications.go` 내 ListRepositories 등 (cross-domain, 후속 carve out 권장) | — |
| schema | SCM Repository 도메인 모델, DB: `repositories` (000002/000042) | (frontend 내장) |

의존 도메인: [integration-registry](../integration-registry/)

## 3. SDLC 문서 link

| 단계 | 위치 | 상태 |
|---|---|---|
| REQ | [`./requirements.md`](./requirements.md) | active (Phase 3 split, 2026-05-29) |
| ARCH | [`./architecture.md`](./architecture.md) | active (Phase 3) |
| API | [`./api.md`](./api.md) | active (Phase 3) |
| TC | [`./test_cases.md`](./test_cases.md) | active (Sprint A follow-up, 2026-06-16 — PR #597 의 repository-kpi-tests-section) |

## 4. 관련 ADR

- (도메인 전용 ADR 없음; integration-registry ADR-0015 참조)

## 5. cross-cutting 참조

- `docs/architecture.md` §10 (Repository)
- `docs/backend_api_contract.md` §13 (Platform/Repository/Project)

## 6. E2E spec

- `frontend/tests/e2e/repositories-ui.spec.ts`
- `frontend/tests/e2e/repositories-detail-negative.spec.ts`
