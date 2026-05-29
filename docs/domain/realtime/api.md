# realtime 도메인 API

- 문서 목적: realtime 도메인 (`/api/v1/realtime/*`) 의 API 계약을 정의한다.
- 범위: WebSocket upgrade, ticket 발급/소비. envelope·공통 enum 은 master `docs/backend_api_contract.md` §1–§2 또는 `docs/api/conventions.md` 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: draft (Phase 3 split)
- 최종 수정일: 2026-05-29
- 관련 문서: [도메인 README](./README.md), [requirements.md](./requirements.md), [architecture.md](./architecture.md), [master API](../../backend_api_contract.md), [ADR-0024](../../adr/0024-websocket-auth-query-token.md)

## 1. API ID 인덱스

| ID | endpoint | 비고 |
| --- | --- | --- |
| API-14 | `GET /api/v1/realtime/ws` | WebSocket upgrade |
| API-37 | `POST /api/v1/realtime/ticket` | ticket 발급 (ADR-0024 §6.6) |

## 2. `GET /api/v1/realtime/ws` (API-14)

REST snapshot 조회 이후 변경 이벤트를 수신하는 WebSocket endpoint다. 브라우저 프론트엔드는 gRPC stream에 직접 연결하지 않는다.

RBAC enabled 환경에서는 `types` query로 필요한 event type을 명시한다. `actor`/`role` query 는 `DEVHUB_AUTH_DEV_FALLBACK=true`인 개발 환경에서만 허용한다. 운영 환경에서는 ticket(API-37) 기반 actor context를 사용한다 ([ADR-0024](../../adr/0024-websocket-auth-query-token.md) — `?access_token=` query fallback 은 ticket-only 컷오버 후 제거됐고 `X-Devhub-Actor` fallback 헤더는 [ADR-0004](../../adr/0004-x-devhub-actor-removal.md) 로 폐기됐다).

### 메시지 envelope

```json
{
  "schema_version": "1",
  "type": "ci.run.updated",
  "event_id": "evt_20260502100000.000000000",
  "occurred_at": "2026-05-02T10:00:00Z",
  "data": {
    "id": "101",
    "repository_name": "devhub-core",
    "status": "success"
  }
}
```

### 초기 event type

```text
infra.node.updated
infra.edge.updated
ci.run.updated
ci.log.appended
risk.critical.created
risk.updated
command.status.updated
notification.created
```

프론트는 알 수 없는 `type`을 무시하고, 지원하지 않는 `schema_version`은 사용자 화면을 깨뜨리지 않는 방식으로 로깅한다.

### 구현 상태

- 2026-05-06 기준 `/api/v1/realtime/ws` endpoint와 in-process WebSocket hub가 구현되어 있다.
- 현재 publish 대상은 dry-run command worker가 발생시키는 `command.status.updated`다.
- 인증, 구독 필터, 마지막 event replay는 ticket 인증 (API-37) 도입과 함께 ADR-0024 §6 carve 들에서 다뤄졌다.

### `command.status.updated` data 예시

```json
{
  "command_id": "cmd_1f2a3b4c5d6e",
  "command_type": "service_action",
  "target_type": "service",
  "target_id": "runner-asia-01",
  "action_type": "restart",
  "status": "succeeded",
  "actor_login": "yklee",
  "result_payload": {
    "executor": "dry_run",
    "message": "Dry-run command accepted without external side effects."
  },
  "updated_at": "2026-05-06T10:00:00Z"
}
```

## 3. `POST /api/v1/realtime/ticket` (API-37)

ADR-0024 §6.6 — 인증 actor 가 단일-사용(single-use, 60s TTL) WebSocket ticket 을 발급받는다. 발급된 ticket 은 `GET /api/v1/realtime/ws?ticket=...` 의 query 로 1회 소비된다.

요청: body 없음. 인증된 session 또는 Bearer token 필수.

응답 (200):

```json
{
  "ticket": "rt_4b2c1f...",
  "expires_at": "2026-05-29T01:23:45Z"
}
```

소비 시 store fault 는 503 으로 응답해 정상 사용자 오거부를 회피한다.

## 4. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/backend_api_contract.md` §8 (Realtime WebSocket 계약) 을 도메인 sub-document 로 이관. ID(API-14/API-37) 보존, 신규 발급 없음. ADR-0024 §6.6 의 ticket 발급 endpoint 본문 통합. |
