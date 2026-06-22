---
title: README
type: source
tags: [domain, README.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/realtime/README.md]
git_commit: fb3894f7
git_branch: chore/260622-wiki-drift-cleanup-3
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T06:03:34Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# realtime 도메인

- 문서 목적: `realtime` 도메인의 SDLC 진입점.
- 범위: WebSocket 기반 백그라운드 이벤트 전송 및 단일-사용 WS 인증 티켓 발행/검증 생명주기.
- 상태: draft (Phase 1 골격)
- 최종 수정일: 2026-05-29
- 관련 문서: [code-taxonomy.md §2.1.10](../../governance/code-taxonomy.md)

## 1. 도메인 정의

> WebSocket을 통한 백그라운드 이벤트 전송 및 단일-사용(single-use) WebSocket 인증 티켓 발행/검증 생명주기를 다룬다. ([code-taxonomy.md §2.1.10](../../governance/code-taxonomy.md))

ticket-only 컷오버 (ADR-0024 §6 carve 5) 후 `?access_token=` query fallback 제거, ticket pattern 으로 일원화.

## 2. 4 계층 모듈 매핑

| 계층 | Backend | Frontend |
|---|---|---|
| view | `backend-core/internal/domain/realtime/view/` (`realtime.go`, `realtime_ticket.go`, `handler.go`) | (frontend 직접 페이지 없음 — 다른 도메인이 subscribe) |
| service | WS 브로드캐스트 이벤트 필터링 (RBAC 재검사), ticket 60s TTL 만료 검증 | `frontend/domain/realtime/service/{realtime,websocket}.service.ts` |
| repository | `backend-core/internal/domain/realtime/repository/realtime_tickets.go` | — |
| schema | DB: `realtime_tickets` (000035) | (frontend 내장) |

의존 도메인: [auth-session](../auth-session/)

## 3. SDLC 문서 link

| 단계 | 위치 | 상태 |
|---|---|---|
| REQ | [`./requirements.md`](./requirements.md) | active (Phase 3 split, 2026-05-29) |
| ARCH | [`./architecture.md`](./architecture.md) | active (Phase 3) |
| API | [`./api.md`](./api.md) | active (Phase 3) |
| TC | `./test_cases.md` | planned (Phase 2, 현재 cross-domain) |

## 4. 관련 ADR

- ADR-0024 (WS `?access_token=` query + ticket 패턴)

## 5. cross-cutting 참조

- `docs/backend_api_contract.md` §3 (Realtime API)
- `docs/architecture.md` §5 (Realtime)
