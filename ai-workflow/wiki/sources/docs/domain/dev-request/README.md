---
title: README
type: source
tags: [domain, README.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/dev-request/README.md]
git_commit: 71c0d2cd
git_branch: chore/260622-wiki-drift-cleanup
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T03:47:55Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# dev-request 도메인

- 문서 목적: `dev-request` (DREQ) 도메인의 SDLC 진입점.
- 범위: 외부/사내 채널을 통해 인입된 신규 시스템 개발 의뢰의 인입, 검토, promote (Platform/Project 자동 생성) 프로세스.
- 상태: draft (Phase 1 골격)
- 최종 수정일: 2026-05-29
- 관련 문서: [code-taxonomy.md §2.1.8](../../governance/code-taxonomy.md)

## 1. 도메인 정의

> 외부 연동 채널 또는 사내 채널을 통해 들어온 신규 시스템 개발 의뢰(DREQ)의 인입, 검토, promote(Platform/Project 자동 생성) 프로세스를 관리한다. ([code-taxonomy.md §2.1.8](../../governance/code-taxonomy.md))

DREQ 6단계 상태 머신 + intake token 만료 처리 + promote 트랜잭션으로 개발 의뢰 전체 lifecycle 을 관리한다.

## 2. 4 계층 모듈 매핑

| 계층 | Backend | Frontend |
|---|---|---|
| view | `backend-core/internal/domain/dev-request/view/` (`dev_requests.go`, `dev_request_intake_auth.go`, `dev_request_intake_tokens_admin.go`, `handler.go`) | `frontend/app/dev-requests/`, `frontend/domain/dev-request/view/{DevRequestDetailModal,DevRequestTable,IntakeTokenTable,MyPendingDevRequestsWidget}.tsx` |
| service | DREQ 6단계 상태 머신 전이 규칙, promote 트랜잭션, intake token 만료 처리 | `frontend/domain/dev-request/service/dev_request.service.ts` |
| repository | `backend-core/internal/domain/dev-request/repository/` (`dev_requests.go`, `dev_request_intake_tokens.go`, `dev_requests_promote.go`) | — |
| schema | DB: `dev_requests` (000022), `dev_request_intake_tokens` (000023) | `frontend/domain/dev-request/schema/dev_request.types.ts` |

## 3. SDLC 문서 link

| 단계 | 위치 | 상태 |
|---|---|---|
| REQ | [`./requirements.md`](./requirements.md) | active (Phase 3 split, 2026-05-29) |
| ARCH | [`./architecture.md`](./architecture.md) | active (Phase 3) |
| API | [`./api.md`](./api.md) | active (Phase 3) |
| TC | `./test_cases.md` | planned (Phase 2 — `docs/domain/dev-request/test_cases.md`) |
| Concept | `./concept.md` | planned (Phase 2 — `docs/domain/dev-request/concept.md`) |

## 4. 관련 ADR

- ADR-0012 (DREQ intake auth)
- ADR-0013 (DREQ RBAC row-scoping)
- ADR-0014 (Platform/Project lifecycle)
- ADR-0017 (DREQ intake token operational hardening)

## 5. cross-cutting 참조

- `docs/backend_api_contract.md` §14 (DREQ API)
- `docs/architecture.md` §7 (DREQ)

## 6. E2E spec

- `frontend/tests/e2e/dev-requests.spec.ts`
