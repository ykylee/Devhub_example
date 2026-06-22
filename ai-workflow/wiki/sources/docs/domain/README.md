---
title: README
type: source
tags: [domain, core, project-devhub]
sources: [raw/projects/devhub/docs/domain/README.md]
git_commit: e91115f0
git_branch: chore/260622-wiki-drift-cleanup-2
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T04:24:49Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# Domain Layer SDLC 문서 (10 도메인)

- 문서 목적: `docs/governance/code-taxonomy.md` §2.1 의 10 도메인 SDLC 문서 진입점 index.
- 범위: Domain 레이어 10 도메인의 README 진입점 + SDLC 단계별 (REQ / ARCH / API / TC) 후속 Phase plan.
- 상태: draft (Phase 1 골격)
- 최종 수정일: 2026-05-29
- 관련 문서: [code-taxonomy.md §2.1](../governance/code-taxonomy.md), [traceability/report.md §3](../traceability/report.md)

## 1. 도메인 index

| # | 도메인 | 진입점 | 핵심 책임 |
|---|---|---|---|
| 1 | auth-session | [`auth-session/`](./auth-session/README.md) | Keycloak OIDC + 토큰 라이프사이클 + 로그인 플로우 |
| 2 | audit-ops | [`audit-ops/`](./audit-ops/README.md) | 감사 로그 발행 + Keycloak event puller + 메트릭 |
| 3 | rbac-permissions | [`rbac-permissions/`](./rbac-permissions/README.md) | RBAC 매트릭스 + PermissionCache + Row-scoping |
| 4 | organization-management | [`organization-management/`](./organization-management/README.md) | 조직 트리 + Appointments + 사용자 마스터 |
| 5 | onboarding | [`onboarding/`](./onboarding/README.md) | 신규 사용자 승인 + 온보딩 게이트 |
| 6 | platform-lifecycle | [`platform-lifecycle/`](./platform-lifecycle/README.md) | Platform/Project CRUD + 상태 머신 + 롤업 |
| 7 | repository-integration | [`repository-integration/`](./repository-integration/README.md) | SCM 저장소-프로젝트 연결 + Import |
| 8 | dev-request | [`dev-request/`](./dev-request/README.md) | DREQ 인입/검토/promote 트랜잭션 |
| 9 | integration-registry | [`integration-registry/`](./integration-registry/README.md) | SCM/비-SCM 공급자 + 바인딩 + ProviderModal |
| 10 | realtime | [`realtime/`](./realtime/README.md) | WebSocket 이벤트 + ticket 인증 |

## 2. 4 계층 매핑 (공통)

모든 Domain 도메인은 코드와 1:1 mirror 관계로 4 계층을 갖는다 ([code-taxonomy.md §2.1](../governance/code-taxonomy.md)):

| 계층 | Backend 위치 | Frontend 위치 |
|---|---|---|
| view | `backend-core/internal/domain/<도메인>/view/` | `frontend/domain/<도메인>/view/` |
| service | `backend-core/internal/domain/<도메인>/service/` | `frontend/domain/<도메인>/service/` |
| repository | `backend-core/internal/domain/<도메인>/repository/` | (해당 없음, backend mirror) |
| schema | `backend-core/internal/domain.*.go` | `frontend/domain/<도메인>/schema/` |

각 도메인 README 의 §2 표 가 도메인별 정확한 모듈 file 경로를 명시한다.

## 3. SDLC 단계별 위치 (Phase 별 작업)

| 단계 | Phase 1 (현재) | Phase 2 (이관 예정) | Phase 3 (split 예정) |
|---|---|---|---|
| README 진입점 | ✅ 본 PR | — | — |
| Concept / Design (이관) | — | ✅ `docs/planning/` → `docs/domain/<도메인>/` | — |
| TC (rename) | — | ✅ `docs/tests/test_cases_m*.md` → `docs/domain/<도메인>/test_cases.md` | — |
| REQ (split) | — | — | ✅ `docs/requirements.md` §4-5 → `docs/domain/<도메인>/requirements.md` |
| ARCH (split) | — | — | ✅ `docs/architecture.md` §5-12 → `docs/domain/<도메인>/architecture.md` |
| API (split) | — | — | ✅ `docs/backend_api_contract.md` §11-17 → `docs/domain/<도메인>/api.md` |

후속 Phase 4: `docs/traceability/report.md` §3 매트릭스 10 도메인 row 재구성.
후속 Phase 5: `docs/governance/document-standards.md` §4 단계별 문서 위치 가이드 갱신.

## 4. cross-domain 참조

- Shared 레이어 — [`docs/shared/`](../shared/README.md)
- Infrastructure 레이어 — [`docs/infrastructure/`](../infrastructure/README.md)
- ADR index — [`docs/adr/`](../adr/)
- 추적성 매트릭스 — [`docs/traceability/report.md`](../traceability/report.md) §3
- 거버넌스 SoT — [`docs/governance/code-taxonomy.md`](../governance/code-taxonomy.md)
