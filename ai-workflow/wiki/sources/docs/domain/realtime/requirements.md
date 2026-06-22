---
title: requirements
type: source
tags: [domain, requirements.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/realtime/requirements.md]
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

# realtime 도메인 요구사항

- 문서 목적: WebSocket 기반 실시간 이벤트 전송 + 단일-사용(single-use) WS 인증 ticket 발행/검증의 기능·비기능 요구사항을 정의한다.
- 범위: realtime 도메인의 REQ-FR / REQ-NFR. 인증 일반 정책은 `auth-session` 도메인, RBAC 정책은 `rbac-permissions` 도메인 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: draft (Phase 3 split)
- 최종 수정일: 2026-05-29
- 관련 문서: [도메인 README](./README.md), [architecture.md §5](./architecture.md) (도메인 ARCH 신규), [API](./api.md) (도메인 API 신규), [master requirements](../../requirements.md), [ADR-0024](../../adr/0024-websocket-auth-query-token.md)

## 1. 개요

본 도메인은 인증된 사용자에게 backend 서비스 상태/이벤트를 실시간으로 push 하기 위한 채널을 정의한다. 인증은 단일-사용 ticket 패턴으로, 권한 검사는 event type 별 RBAC matrix 재검사로 보장한다.

본 도메인의 ID 본문은 `docs/requirements.md` master 의 § "기술 스택 결정 사항" / "상세 시스템 아키텍처 설계" 항목 중 실시간 통신 관련 결정(`WebSocket을 통한 프론트엔드 실시간 상태 전송`)을 출발점으로 한다. 추적성 매트릭스 측면의 REQ-FR ID 매핑은 `docs/traceability/report.md` §3 에서 유지한다.

## 2. 기능 요구사항 (REQ-FR — realtime)

- **WebSocket 채널 (확정):** Backend 는 인증된 사용자 세션에 대해 WebSocket 채널을 제공하고, Gitea Actions 빌드 상태 실시간 업데이트 / 긴급 리스크 알림 / 실시간 이슈 액티비티 피드를 전송한다.
- **Ticket 인증 (확정, [ADR-0024](../../adr/0024-websocket-auth-query-token.md)):** WebSocket 인증은 단일-사용 ticket 패턴을 사용한다. 인증된 actor 가 `POST /api/v1/realtime/ticket` 으로 60초 TTL ticket 을 발급받고 `GET /api/v1/realtime/ws?ticket=...` 로 업그레이드 시 ticket 을 소비한다.
- **Ticket Store (확정):** ticket store 는 in-memory(single-instance) 또는 PostgreSQL `realtime_tickets`(`DELETE ... RETURNING` 으로 multi-instance 원자 소비, migration 000035) 다.
- **`?access_token=` query fallback 폐기 (확정, ADR-0024 §6 carve 5):** 초기에 검토된 `?access_token=` query 직접 전달 방식은 access token 노출 위험 때문에 ticket-only 컷오버로 제거됐다.
- **Event RBAC 재검사 (확정):** WS subscribe 이후 각 event type 별로 RBAC matrix(`PermissionCache.Allows(role, resource, action)`)를 재검사해 권한 없는 event 구독을 거부한다.

## 3. 비기능 요구사항 (REQ-NFR — realtime)

- **장애 격리 (MVP):** ticket consume 시 store fault 는 401 이 아니라 503 으로 응답해 정상 사용자 오거부를 회피한다.
- **단일 포트 정합 (MVP):** ticket 발급/WS upgrade 는 same-origin 내부 path-relative 경로만 사용한다(ADR-0018, single-port). 외부 host:port redirect 금지.
- **SSE 대체 정책 (확정):** SSE 는 초기 구현 범위에서 제외하며, 프록시/운영 환경 제약으로 WebSocket 유지가 어렵다고 확인될 때 별도 fallback 으로 재검토한다.
- **알려진 표면 (Hardening 후속):** `CheckOrigin` 이 현재 모든 origin 을 허용한다 (ticket 인증으로 보호되나 CSWSH 표면 잔존). origin 검증 강화는 후속 hardening carve.

## 4. 범위 경계 (Out of Scope)

- SSE / Long-polling fallback (위 §3 정책).
- 비인증 public 채널 (모든 채널은 인증 후 ticket 발급 필요).
- WebSocket 위에서의 application-level QoS / acknowledgement / replay buffer 등 메시징 보강 — 후속 carve.

## 5. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — `docs/requirements.md` 의 실시간 통신 관련 결정 + `docs/architecture.md` §3.2 / §6.5.3 (보강) 의 ticket 인증 정책을 도메인 sub-document 로 이관. ID 보존, 신규 ID 미발급. |
