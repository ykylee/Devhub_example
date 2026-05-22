# JWKS rotation 직후 backend cache flush 운영 SOP

- 문서 목적: Keycloak JWKS key rotation (정상 또는 비상) 후 DevHub backend 의 JWKS cache 를 강제 flush 하는 운영 절차를 정의한다. [ADR-0020 §6.3](../adr/0020-account-user-management-boundary.md) 의 "JWKS rotation 직후 backend cache flush SOP" 사내 동반 carve.
- 범위: backend 의 JWKS cache (TTL 5분) + stale-while-error fallback (MaxStaleDuration default 24h, PR #242) 의 강제 invalidation 절차. revoked key (유출 의심) 시나리오의 backend 측면 대응. [Keycloak 운영 SOP §6.5](./keycloak_operations.md#65-비상-rotation-key-유출-의심-시) 의 Keycloak 측면 SOP 와 짝.
- 대상 독자: DevHub 운영자 (SRE), 사내 IdP 팀, security, on-call.
- 상태: draft (1차)
- 최종 수정일: 2026-05-22
- 결정 근거 sprint: `claude/work_260522-internal-coordinated-carve-docs`
- 관련 문서: [Keycloak 운영 SOP §6](./keycloak_operations.md#6-jwks-rotation-운영-sop), [ADR-0019 §5.3](../adr/0019-keycloak-only-idp.md), [ADR-0020 §6.3](../adr/0020-account-user-management-boundary.md), [account/user redesign §5.6.3](../planning/account_user_management_redesign.md), [keycloak_admin_responsibility](../governance/keycloak_admin_responsibility.md).

## 1. 배경

[ADR-0020 sub-carve D PR #242](../adr/0020-account-user-management-boundary.md) (sprint `-l`) 가 도입한 **JWKS stale-while-error fallback** 은 Keycloak unreachable 시 cachedUntil + MaxStaleDuration (default 24h, env `DEVHUB_OIDC_JWKS_MAX_STALE_DURATION`) 안의 stale cache 사용을 허용한다. 이는 Keycloak 90일 key rotation 주기까지의 uptime 보장에 정합.

그러나 본 fallback 은 **revoked key 보호 측면 위협**을 동반:

- Keycloak 이 의도적 rotation (security incident — key 유출 의심) 으로 옛 key 를 disable 했을 때
- backend 의 stale cache 가 그 옛 key 를 (MaxStaleDuration 24h 안) 그대로 보관
- 그 옛 key 로 발급된 (또는 forge 된) token 이 stale fallback 경로에서 검증 통과
- → revoked key 보호 깨짐

본 SOP 가 이 시나리오의 **운영 측면 대응** — Keycloak 비상 rotation 시 backend cache 도 동반 강제 flush.

## 2. backend JWKS cache 동작 요약

[`backend-core/internal/auth/keycloak_verifier.go`](../../backend-core/internal/auth/keycloak_verifier.go) 의 cache 동작:

| 동작 | 값 | 출처 |
| --- | --- | --- |
| cache TTL | 5분 (`defaultJWKSTTL`) | sprint `claude/work_260513-q` 1차 도입 |
| stale-while-error MaxStaleDuration | 24h default (env `DEVHUB_OIDC_JWKS_MAX_STALE_DURATION` override) | PR #242 sub-carve D |
| `kid` mismatch retry | 자동 — 1회 invalidate + refetch + 1회 retry | sprint `-r` PR #186 |
| signature/expired/issuer/audience error | retry 안 함 | PR #186 |
| cache flush endpoint | **미구현** — backend 재기동만 가능 | 본 SOP scope |

## 3. Trigger 분류 — 정상 rotation vs 비상

### 3.1 정상 rotation (90일 주기, key 유출 미의심)

- Keycloak 측 절차: [Keycloak 운영 SOP §6.4](./keycloak_operations.md#64-rotation-운영-sop-d-day)
- backend 측 동작: **자동 kid mismatch fallback** — PR #186 의 invalidate + refetch + 1회 retry 가 graceful window 0 처리
- **운영자 액션**: 기본적으로 불필요. 단 §6.4 의 검증 step 에서 backend log refetch 확인.

### 3.2 비상 rotation (key 유출 의심)

- Keycloak 측 절차: [Keycloak 운영 SOP §6.5](./keycloak_operations.md#65-비상-rotation-key-유출-의심-시)
- backend 측 위협: **stale cache 에 잔존 옛 key 가 forge token 검증 통과 위험**
- **운영자 액션**: 본 SOP §4 (강제 cache flush) **즉시 수행**.

### 3.3 passive cleanup (D+14)

- Keycloak 측 절차: §6.4 의 step 4 (이전 key provider disable, passive 종료)
- backend 측 위협: 옛 key 로 발급된 잔여 token 검증 실패 → 강제 재로그인
- **운영자 액션**: backend 측 변경 불필요. token 만료 자연 처리.

## 4. 강제 cache flush 절차 (비상 rotation 시)

### 4.1 사전 조건

- Keycloak 측 비상 rotation 이 [§6.5](./keycloak_operations.md#65-비상-rotation-key-유출-의심-시) step 1~3 (의심 key disable + Revoke all sessions + 신규 key) 완료
- IdP 팀이 DevHub 운영자에게 통보 (사내 운영 채널)

### 4.2 backend 강제 재기동 (정공법)

**deployment 환경별**:

| 환경 | 명령 |
| --- | --- |
| docker compose | `docker compose restart backend-core` |
| systemd (native) | `sudo systemctl restart devhub-backend` |
| kubernetes | `kubectl rollout restart deployment/devhub-backend -n <namespace>` |
| supervisor | `sudo supervisorctl restart devhub-backend` |

**재기동 효과**:
- JWKS in-memory cache 완전 비움 (struct field reset)
- 다음 token 검증 요청 시 새 JWKS fetch → 새 active key 만 cache 진입
- 옛 (revoked) key 의 cache 잔존 0건

### 4.3 검증 (재기동 후 5분 이내)

1. **health check**: `curl http://localhost:8080/health` → 200 OK
2. **1 test login**: 운영자 본인 계정으로 DevHub `/login` 진입 → `/auth/callback` → `/developer` (또는 admin landing) 정상 도달
3. **JWKS fetch log 확인**:
   ```bash
   # backend stdout 또는 journalctl
   grep '\[jwks\]' /var/log/devhub/backend.log | tail -10
   # expected: 재기동 직후 1건의 `[jwks] fetched ... kid=...` log
   ```
4. **Prometheus metric**:
   - `devhub_jwks_fetch_total{status="success"}` 가 1 증가
   - `devhub_jwks_cache_age_seconds` 가 0 으로 리셋 (혹은 cache hit count 가 첫 hit 이후 정합)

### 4.4 영향

| 영역 | 영향 |
| --- | --- |
| 사용자 세션 | 재기동 시 backend HTTP 연결 일시 끊김 (~5초). 이미 발급된 valid token 으로 재요청 시 정상 처리. Keycloak 측 "Revoke all sessions" 실행 시 모든 사용자 재로그인 강제. |
| 활성 WebSocket 연결 | 재기동 시 모든 WS 끊김. frontend `websocketService` 가 자동 재연결 (5초 backoff). |
| 진행 중 트랜잭션 | gin graceful shutdown timeout (사내 정책) — 진행 중 요청은 마감. 새 요청은 새 backend instance 가 받음. |
| audit_logs | 재기동 자체는 emit 없음. JWKS fetch 의 stdout log 만. |

### 4.5 후속 점검 (재기동 후 1시간)

- `devhub_jwks_stale_while_error_total{result="success"}` 가 비상 rotation 후 0 (정상 fresh fetch)
- backend log 의 `[jwks] kid mismatch retry` 패턴 0건 (모든 token 이 새 active key 정합)
- DevHub `/admin/settings/audit` 에서 비상 rotation 시점 전후의 `auth.login` action 비교 — 비정상 spike 없음 (revoked key forge token 시도 검출)

## 5. Cache flush endpoint 도입 carve (P3)

본 SOP 는 backend 재기동을 정공법으로 정의한다. 사내 정책상 재기동 부담이 크면 별도 cache flush endpoint 도입 검토:

- 후보 endpoint: `POST /api/v1/internal/jwks/flush` (system_admin only, RBAC + 내부 IP allowlist)
- 동작: `keycloakVerifier.invalidateCache()` 호출 + 0 downtime
- carve 영역: backend handler + RBAC policy + audit action `internal.jwks_flushed` + Prometheus metric

본 SOP 의 정공법 (재기동) 으로 충분한지 1주 monitoring + 사내 정책 검토 후 carve 진입 결정.

## 6. 트러블슈팅

| 증상 | 1차 의심 | 검증 / 대응 |
| --- | --- | --- |
| 재기동 후 backend `/health` 401 | 재기동 직후 첫 JWKS fetch 실패 (Keycloak 측 신규 key 가 활성 안 됐을 수 있음) | Keycloak `/realms/devhub/protocol/openid-connect/certs` curl 직접 확인 → 신규 key 응답 여부. 재기동 timing 이 너무 빨랐다면 1분 후 재시도 |
| 재기동 후에도 옛 token 정상 통과 | Keycloak 측 "Revoke all sessions" 미실행 또는 옛 key 가 still active | Keycloak admin → Realm settings → Keys 의 옛 key state 확인 (disabled / passive 검증) |
| Prometheus `devhub_jwks_cache_age_seconds` 가 재기동 후에도 0 으로 안 떨어짐 | metric 수집 lag 또는 backend 가 재기동 안 됨 | `systemctl status devhub-backend` (또는 docker ps) 로 process uptime 검증 |
| `devhub_jwks_stale_while_error_total{result="success"}` 가 재기동 후에도 증가 | Keycloak unreachable 또는 issuer URL 오기재 | backend log 의 `[jwks] stale fallback` 메시지 grep + `DEVHUB_OIDC_JWKS_URL` env 확인 |

## 7. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-22 | 1차 draft — §1 배경 (PR #242 stale-while-error fallback 후 revoked key 위협) + §2 backend cache 동작 표 + §3 trigger 3 분류 (정상 / 비상 / passive cleanup) + §4 강제 재기동 절차 (환경별 4가지 + 검증 4 step + 영향 + 후속 점검) + §5 cache flush endpoint carve P3 + §6 트러블슈팅 4 case. [Keycloak 운영 SOP §6.5](./keycloak_operations.md#65-비상-rotation-key-유출-의심-시) Keycloak 측 SOP 와 짝. | `claude/work_260522-internal-coordinated-carve-docs` |
