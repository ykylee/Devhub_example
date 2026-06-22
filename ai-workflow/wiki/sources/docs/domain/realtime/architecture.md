---
title: architecture
type: source
tags: [domain, architecture.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/realtime/architecture.md]
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

# realtime 도메인 아키텍처

- 문서 목적: realtime 도메인의 컴포넌트·통신·인증 경계·이벤트 카탈로그를 정의한다.
- 범위: WebSocket 채널, ticket store, event RBAC 재검사, single-port 정합. cross-cutting 3대 레이어 아키텍처 + 호출 규칙은 master `docs/architecture.md` §1–§4 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, 아키텍처 검토자.
- 상태: draft (Phase 3 split)
- 최종 수정일: 2026-05-29
- 관련 문서: [도메인 README](./README.md), [requirements.md](./requirements.md), [API](./api.md), [master architecture](../../architecture.md), [ADR-0024](../../adr/0024-websocket-auth-query-token.md).

## 1. 컴포넌트 (ARCH-RT-01)

- `backend-core/internal/domain/realtime/view/` — `realtime.go` (WS upgrade handler), `realtime_ticket.go` (ticket 발급/소비 handler), `handler.go` (router 등록).
- `backend-core/internal/domain/realtime/repository/realtime_tickets.go` — PostgreSQL `realtime_tickets` store + in-memory fallback.
- in-process WebSocket hub (publish bus) — command worker / domain service 가 hub 를 통해 event 를 broadcast.
- frontend `frontend/domain/realtime/service/{realtime,websocket}.service.ts` — client-side WS 연결, ticket 재발급, event subscription.

## 2. WebSocket 인증 경계 (ARCH-RT-02)

cross-cutting 보안 정책은 master `docs/architecture.md` §6.5.3 (ADR-0024) 에서 확정한다. 본 도메인 관점 요약:

- `POST /api/v1/realtime/ticket` — 인증 actor 면 RBAC bypass, 60s TTL single-use ticket 발급.
- `GET /api/v1/realtime/ws?ticket=...` — ticket single-use consume.
- ticket store fault → 503 (정상 사용자 오거부 회피).
- subscribe 후 각 event type 별 RBAC matrix 재검사.
- ticket-only 컷오버 (ADR-0024 §6 carve 5) 후 `?access_token=` query fallback 제거.

## 3. Ticket Store 모델 (ARCH-RT-03)

```text
realtime_tickets (migration 000035)
  ticket_id    text       PK
  user_id      text       NOT NULL  REFERENCES users(user_id)
  expires_at   timestamptz NOT NULL  -- 발급 + 60s
  created_at   timestamptz NOT NULL DEFAULT NOW()
```

원자 소비는 `DELETE FROM realtime_tickets WHERE ticket_id=$1 AND expires_at>NOW() RETURNING user_id` 단일 쿼리로 보장한다(multi-instance 환경). in-memory fallback 은 single-instance 개발/테스트 환경 전용.

## 4. Event Catalog (ARCH-RT-04)

초기 event type (master `docs/architecture.md` §3.2 + `docs/backend_api_contract.md` §8 동기):

- `infra.node.updated`
- `infra.edge.updated`
- `ci.run.updated`
- `ci.log.appended`
- `risk.critical.created`
- `risk.updated`
- `command.status.updated`
- `notification.created`

알 수 없는 `type` 은 frontend 가 무시하고, 지원하지 않는 `schema_version` 은 사용자 화면을 깨뜨리지 않는 방식으로 로깅한다.

## 5. Single-Port 정합 (ARCH-RT-05)

ADR-0018 (single-port) 정합 — ticket 발급 / WS upgrade 는 same-origin 내부 path-relative 만 사용. 외부 host:port 로의 `c.Redirect` / `Location:` 헤더 직접 작성 = 0 hit 가드 유지.

## 6. 알려진 표면 + Hardening 후속

- `CheckOrigin` 이 현재 모든 origin 을 허용한다 — ticket 인증으로 보호되나 CSWSH 표면 잔존.
- origin 검증 강화는 별도 hardening carve.

## 7. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/architecture.md` §3.2 (보강) + §6.5.3 (WS ticket 인증 경계) + `docs/backend_api_contract.md` §8 (Realtime WebSocket 계약) 의 realtime 컴포넌트·인증·이벤트 카탈로그를 도메인 sub-document 로 이관. ID 보존, 신규 ARCH-RT-01..05 발급은 본 sprint scope. |
