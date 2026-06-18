---
title: README
type: source
tags: [domain, README.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/audit-ops/README.md]
git_commit: 6c434887
git_branch: main
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-18T12:08:55Z
mirror_dirty: (dirty: uncommitted changes) |
related: [none]
status: draft
contradictions: [none]
---

# audit-ops 도메인

- 문서 목적: `audit-ops` 도메인의 SDLC 진입점.
- 범위: 사용자 및 시스템에 의해 트리거된 주요 변경 사항을 캡처하여 감사용 영속 로그를 발행하고 메트릭을 수집한다.
- 상태: draft (Phase 1 골격)
- 최종 수정일: 2026-05-29
- 관련 문서: [code-taxonomy.md §2.1.2](../../governance/code-taxonomy.md), [traceability/report.md §3](../../traceability/report.md)

## 1. 도메인 정의

> 사용자 및 시스템에 의해 트리거된 주요 변경 사항을 캡처하여 감사용 영속 로그를 발행하고 메트릭을 수집한다. ([code-taxonomy.md §2.1.2](../../governance/code-taxonomy.md))

Keycloak event listener cron + audit log persistence + Prometheus metric 발행을 통해 시스템 변경 가시성을 확보한다.

## 2. 4 계층 모듈 매핑

| 계층 | Backend | Frontend |
|---|---|---|
| view | `backend-core/internal/domain/audit-ops/view/` (`audit.go`, `keycloak_events_webhook.go`, `handler.go`) | (admin 페이지 — settings 통합) |
| service | `backend-core/internal/domain/audit-ops/service/` (`keycloak_event_puller.go`, `user_sync.go`, `metrics.go`, `keycloak_admin_adapter.go`) | `frontend/domain/audit-ops/service/audit.service.ts` |
| repository | `backend-core/internal/domain/audit-ops/repository/` (`audit_logs.go`, `event_cursors.go`) | — |
| schema | `domain/audit.go` (AuditLog 엔티티), DB: `audit_logs` (000003), `event_cursors` (000031) | `frontend/domain/audit-ops/schema/audit.types.ts` |

의존 도메인: [auth-session](../auth-session/) (요청 actor 식별)

## 3. SDLC 문서 link

| 단계 | 위치 | 상태 |
|---|---|---|
| REQ | [`./requirements.md`](./requirements.md) | active (Phase 3 split, 2026-05-29) |
| ARCH | [`./architecture.md`](./architecture.md) | active (Phase 3) |
| API | [`./api.md`](./api.md) | active (Phase 3) |
| TC | `./test_cases.md` | planned (Phase 2 이관) |

## 4. 관련 ADR

- ADR-0020 sub-carve E (audit event listener)

## 5. cross-cutting 참조

- `docs/setup/keycloak_operations.md` §8.6 (event listener 운영 SOP)
- `docs/governance/code-taxonomy.md` §2.1.2
