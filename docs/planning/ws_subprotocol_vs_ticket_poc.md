# WebSocket 인증 — Subprotocol negotiation vs Ticket PoC 비교 (ADR-0024 §6 carve 4)

- 문서 목적: 브라우저 WebSocket 인증에서 `Sec-WebSocket-Protocol` (subprotocol) 로 Bearer 를 전달하는 방식을 ticket 패턴과 비교 분석하고, DevHub 가 ticket 패턴을 채택한 근거를 명문화한다.
- 범위: 브라우저 WS API 제약 + subprotocol/ticket 메커니즘 + 보안/운영 비교 + 결론. 코드 PoC 스케치 포함.
- 대상 독자: backend/frontend 담당자, 리뷰어, 후속 ADR 작성자.
- 상태: accepted (ADR-0024 §6 carve 4 resolved 근거)
- 최종 수정일: 2026-05-27
- 관련 문서: [ADR-0024](../adr/0024-websocket-auth-query-token.md) (§3.3 reject 대안 + §6 carve), [`backend-core/internal/httpapi/auth.go`](../../backend-core/internal/httpapi/auth.go) (`/api/v1/realtime/ws` 인증), [`backend-core/internal/httpapi/realtime_ticket.go`](../../backend-core/internal/httpapi/realtime_ticket.go), [`frontend/lib/services/realtime.service.ts`](../../frontend/lib/services/realtime.service.ts)

## 1. 배경

브라우저 `new WebSocket(url, protocols?)` 는 `Authorization` 헤더를 설정할 수 없다 (W3C/WHATWG WS API). ADR-0024 §2.3 의 4 우회 패턴 중 query string 은 단기 채택 (PR #335) → ticket 패턴으로 정공법 확장 (carve 1, done) → 본 문서가 남은 subprotocol 대안 (carve 4) 을 비교한다.

## 2. Subprotocol negotiation 메커니즘

```js
// frontend
new WebSocket(url, ["bearer", accessToken]);
// → handshake 요청 헤더: Sec-WebSocket-Protocol: bearer, <accessToken>
```

```go
// backend (gorilla/websocket)
var upgrader = websocket.Upgrader{
    Subprotocols: []string{"bearer"}, // 서버가 negotiation 에 참여
}
// handshake 에서 r.Header["Sec-Websocket-Protocol"] 파싱 → 2번째 토큰 = bearer 값
// 검증 성공 시 upgrader.Upgrade 가 응답에 Sec-WebSocket-Protocol: bearer 에코
```

핵심: 브라우저는 token 을 **요청 헤더** (`Sec-WebSocket-Protocol`) 에 싣고, 서버는 negotiation 으로 받아 검증한다. URL query 가 아니라는 점이 query string 대비 차이.

## 3. 비교

| 축 | Query string (`?access_token=`) | Subprotocol (`Sec-WebSocket-Protocol`) | **Ticket (`?ticket=`, 채택)** |
| --- | --- | --- | --- |
| token 노출 | URL → access_log / proxy log leak (장기 JWT) | **handshake 헤더 → 여전히 access/proxy log 에 기록될 수 있음** (장기 JWT) | **single-use + 60s TTL** — leak 돼도 즉시 만료/소진 |
| 브라우저 지원 | 전부 | 전부 (`protocols` 인자) | 전부 (query) |
| 서버 구현 | query 파싱 1줄 | upgrader Subprotocols + 헤더 파싱 + negotiation 에코 (분량 증가) | ticket 발급 endpoint + store (구현됨) |
| revocation | 불가 (JWT 만료까지 유효) | 불가 | **즉시** (single-use consume) |
| replay 내성 | 낮음 (JWT 유효기간 내 재사용) | 낮음 | **높음** (1회 소비 후 무효) |
| 운영 redact | nginx query redact 필요 | 헤더 redact 필요 (log_format 조정) | ticket 은 단기라 redact 부담 ↓ |

## 4. 결론

**Subprotocol 은 채택하지 않는다.** query string 대비 보안 이점이 marginal 하다 — 핵심 위협(장기 JWT 가 log 에 남는 것)이 subprotocol 에서도 그대로 남고 (handshake 헤더도 access log/디버그 로그에 기록될 수 있음), revocation/replay 내성도 없다. 반면 서버 구현 분량은 증가한다 (ADR-0024 §3.3 reject 사유와 일치).

ticket 패턴이 우월하다: single-use + 60s TTL 로 leak/replay 위협을 구조적으로 차단하고, 발급 endpoint 가 Bearer 인증을 그대로 재사용하므로 RBAC 정합도 유지된다. DevHub 는 ticket-only 로 컷오버 완료 (carve 5) — query/subprotocol 모두 미사용.

향후 ticket store 가 multi-instance 부담이 되거나 (carve 6, PG 백킹으로 해소됨) 추가 요구가 생기면 재검토하되, 현재 결론은 **ticket 단일화**다.
