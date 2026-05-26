# ADR-0024: WebSocket 인증 패턴 — `?access_token=` query string (browser 제약 우회)

## 1. 상태
- **상태**: Accepted
- **작성일**: 2026-05-26
- **수정일**: 2026-05-26
- **결정 근거 sprint**: `claude/work_260526-organization-followups`
- **관련 PR/문서**: [PR #335](https://github.com/ykylee/Devhub_example/pull/335) (frontend WS auth fix, 사용자 보고 root cause), [`backend-core/internal/httpapi/auth.go:79-83`](../../backend-core/internal/httpapi/auth.go), [`frontend/lib/services/realtime.service.ts:128-145`](../../frontend/lib/services/realtime.service.ts), [ADR-0019 Keycloak 단일 IdP](./0019-keycloak-only-idp.md).

## 2. 컨텍스트

### 2.1 사용자 보고 회귀 (2026-05-26)

사내 테스트 환경에서 사용자가 "WebSocket 에서 계속 연결 오류 발생" 보고. 진단 결과:

- `frontend/lib/services/realtime.service.ts:buildURL()` 가 WS handshake URL 에 token 미첨부
- `backend-core/internal/httpapi/auth.go:75-84` 의 `authenticateActor` middleware 가 `/api/v1/realtime/ws` 경로에 한해 `?access_token=<bearer>` query parameter 를 인식 (Authorization header 부재 시 fallback)
- frontend 가 query token 미첨부 → backend 401 → frontend `handleReconnect` 3s 후 재시도 → 무한 401 cycle

PR #335 (sprint `claude/work_260526-realtime-ws-auth-fix`) 가 `tokenStore.getAccessToken()` 첨부로 즉시 hotfix.

### 2.2 ADR 발급 동인

본 패턴 (query string Bearer) 은 결정/risk 명문화 없이 backend 에 잠재 — 사용자 보고 회귀 root cause 가 frontend/backend 양쪽 정합 누락. ADR 미발급 = 향후 회귀 위험. 본 ADR 이 패턴 + risk + 후속 carve 명문화.

### 2.3 브라우저 WebSocket API 제약

`new WebSocket(url, protocols?)` 의 제약:

- **Authorization header 설정 불가** (W3C / WHATWG WebSocket API spec) — fetch/XHR 와 달리 arbitrary header set 미지원
- 표준 우회 패턴 4 가지:
  1. **Subprotocol field** (`new WebSocket(url, ["bearer", token])`) — Sec-WebSocket-Protocol header 에 token 전달
  2. **Query string** (`ws://...?access_token=<token>`) — URL 에 token 첨부
  3. **Cookie 기반** (Same-Origin) — 브라우저가 자동 첨부
  4. **Ticket 패턴** — REST 로 short-lived ticket 발급 → ticket 으로 handshake

## 3. 결정

**현재 구현 (query string `?access_token=`) 을 단기 채택하되, 중기 carve out 으로 ticket 패턴 검토.**

### 3.1 단기 결정 (query string 유지)

backend `auth.go:79-83` + frontend `realtime.service.ts:buildURL` 의 query string 패턴을 active code 로 유지. 본 PR #335 가 frontend hotfix 완료 후 사내 환경 검증 완료 시점.

근거:
- 즉시 사용 가능 (배포 lag 최소)
- backend 가 이미 query parsing 구현 (`auth.go:79-83`)
- 마이너 보안 risk 는 §4 운영 권장으로 완화

### 3.2 중기 carve (ticket 패턴 검토)

[OWASP WebSocket Security guide](https://cheatsheetseries.owasp.org/cheatsheets/WebSocket_Security_Cheat_Sheet.html) 권장 = ticket pattern. 후속 carve:

- REST endpoint `POST /api/v1/realtime/ticket` — short-lived (60s) one-time ticket 발급, Bearer JWT 인증 필요
- frontend 가 connect 직전 ticket 발급 → `wss://...?ticket=<ticket>` 으로 handshake
- backend 가 ticket 검증 (single-use + TTL) 후 actor role 부착

본 ADR 의 §6 carve 1번 항목.

### 3.3 reject 한 대안

| 대안 | reject 사유 |
| --- | --- |
| Subprotocol field | gorilla/websocket 의 Subprotocols negotiation 이 token 으로 사용 가능하나, browser 의 Sec-WebSocket-Protocol negotiation 이 token 노출 (handshake log 등). 보안 차이 minimal vs query string. backend 구현 분량 증가. |
| Cookie 기반 | Same-Origin 가정. SSO 시나리오 (Keycloak issuer 가 다른 origin) 에서 CORS preflight 우회 어려움. ADR-0019 Keycloak 단일 IdP 의 cross-origin token flow 와 부합 안 함. |

## 4. 결과 + 운영 권장

### 4.1 즉시 적용 (PR #335)

- frontend `realtime.service.ts:buildURL` 가 `tokenStore.getAccessToken()` 첨부
- backend 무변경 (이미 query parsing 구현)

### 4.2 보안 risk 완화

query string token 의 일반적 risk + 완화:

| risk | 완화 |
| --- | --- |
| URL 이 access log / proxy log 에 leak | nginx access_log 의 query string redact (`if ($args ~* "access_token=([^&]+)") { ... }`) — 사내 nginx 동반 carve |
| browser history 에 leak | WebSocket URL 은 history 에 안 들어감 (a tag/location.href 가 아님) — 해당 없음 |
| referrer leak | WebSocket 은 Referer header 안 보냄 — 해당 없음 |
| URL 가시성 (어깨너머 보기) | 일반 페이지 가시성과 동일. risk minimal |

### 4.3 사내 동반 carve (§6 1번)

- nginx access_log 에서 `access_token=` query parameter redact 설정 추가 (사내 운영자 영역)
- 또는 ticket pattern 으로 마이그레이션 (claude 영역, 별도 sprint)

## 5. 변경 이력

| 일자 | 변경 | sprint/PR |
| --- | --- | --- |
| 2026-05-26 | 본 ADR 신규 발급 (사용자 보고 + PR #335 client-side hotfix 명문화) | sprint `claude/work_260526-organization-followups`, PR pending |
| 2026-05-26 | **§6 carve 1, 3 closed + carve 2 sample 작성** — 사용자 명시 override ("영역 무시") 로 ticket pattern + refresh-then-reconnect 본 sprint 추가 흡수. backend `realtime_ticket.go` 신규 (in-memory store + 60s TTL + single-use) + `POST /api/v1/realtime/ticket` endpoint + `auth.go` 의 `?ticket=` query 인식. frontend `realtime.service.ts:buildURL` async + ticket fetch + 401 시 refresh-then-retry. carve 2 sample 은 `infra/nginx/README.md §6` 에 http block log_format 권장 안 (사내 nginx 운영자 영역). access_token query 는 backward-compat fallback 유지 (deprecated, removal 차기 sprint). | 본 sprint |

## 6. 잔여 carve

| ID | 항목 | 영역 | 우선순위 | 상태 |
| --- | --- | --- | --- | --- |
| 1 | ticket pattern 마이그레이션 (`POST /api/v1/realtime/ticket` + short-lived TTL + single-use) | claude (backend + frontend) | P2 | ✅ resolved (본 sprint) |
| 2 | nginx access_log 의 `access_token=` / `ticket=` redact 설정 | 사내 nginx 운영자 | P2 | sample 작성 ([`infra/nginx/README.md §6`](../../infra/nginx/README.md#6-websocket-auth-query-token-redact-adr-0024-43-6-carve-2)), 적용 사내 |
| 3 | WS handshake 401 시 frontend refresh-then-reconnect 패턴 | claude | P3 | ✅ resolved (본 sprint, ticket fetch 의 401 → `authService.refreshTokens()` → ticket retry 1회) |
| 4 | WS subprotocol negotiation 으로 Bearer 전달 비교 PoC | claude (architecture 검토) | P3 | open |
| 5 | access_token query backward-compat fallback 제거 (모든 client 가 ticket 사용 확인 후) | claude (backend + frontend) | P3 | open (carve 1 의 자연 후속) |
| 6 | multi-instance backend 환경에서 in-memory ticket store 가 sticky 미사용 시 깨짐 → Redis/PG 백킹 store | claude (backend) | P2 | open (현재 single-instance 가정) |
