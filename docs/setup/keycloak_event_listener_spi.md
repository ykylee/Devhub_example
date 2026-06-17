# Keycloak Event Listener SPI 정공법 SOP (P2-6 + X-8)

문서 목적: P2-6 (Keycloak SPI provider JAR) + X-8 (P3-5 audit event listener push 전환) 의 정공법 SOP. Keycloak 의 user/admin event 를 polling (30s) → push (<1s) 로 전환하여 backend audit-ops latency 개선. ADR-0019 §5.3 (9) + ADR-0022 §4.1 + sprint `feat/260617-x8-keycloak-event-listener-spi-sop` (PR #641+).

관련 문서:
- [`docs/setup/keycloak_operations.md`](keycloak_operations.md) — Keycloak admin SOP (§4.3 group + §4.4 verify-keycloak-groups.sh, P1-3 / X-6 정공법)
- [`infra/idp/keycloak-event-listener-spi/`](../../infra/idp/keycloak-event-listener-spi/) — SPI source (Java 21 + Maven 3.13)
- [`backend-core/internal/domain/audit-ops/service/keycloak_event_puller.go`](../../backend-core/internal/domain/audit-ops/service/keycloak_event_puller.go) — backend polling cron (residual — push 우선, polling fallback)
- [`backend-core/internal/domain/audit-ops/view/keycloak_events_webhook.go`](../../backend-core/internal/domain/audit-ops/view/keycloak_events_webhook.go) — backend webhook handler (`POST /api/v0-1/internal/keycloak-events`)
- [`backend-core/internal/domain/audit-ops/repository/event_cursors.go`](../../backend-core/internal/domain/audit-ops/repository/event_cursors.go) — dedup state (polling cursor = at-least-once)
- ADR-0019 §5.3 (9) — audit event listener SPI push 전환 (latency < 1s)
- ADR-0022 §4.1 — Keycloak 단일 IdP + SPI 도입

## 1. SPI architecture

```
Keycloak (user/admin event)
  ↓ onEvent/onEvent(AdminEvent) [Keycloak EventListenerProvider SPI]
DevHubEventListenerProvider (Java) — sendAsync (non-blocking)
  ↓ HTTP POST + X-Webhook-Secret header
  ↓ env: DEVHUB_BACKEND_SPI_WEBHOOK_URL = http://backend-core:8080/api/v0-1/internal/keycloak-events
  ↓ env: DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET = <vault-managed>
backend audit-ops /keycloak_events_webhook handler
  ↓ X-Webhook-Secret 검증 + dedup (event_cursors) + audit_logs INSERT
  ↓ latency < 1s (push)
  ↓ residual: polling cron 30s fallback (keycloak_event_puller.go) — push 가 도착하면 cursor 가 그보다 작게 유지
```

## 2. SPI build (sabun convention)

Prerequisites:
- Java 21+
- Maven 3.13+

```bash
cd infra/idp/keycloak-event-listener-spi
mvn clean package
# 산출물: target/devhub-keycloak-event-listener.jar
# 검증: mvn test (unit test 0건 — 본 SPI 가 외부 HTTP 만 의존, mock 으로 unit test 미작성)
```

## 3. compose volume mount (정공법 = 2026-06-17)

### 3.1 colima/local (정합 ✓, PR #641+)

[`docker-compose.colima.yml`](../setup/keycloak_event_listener_spi.md) 의 `keycloak` service:
- `volumes`:
  - `./infra/idp/keycloak-event-listener-spi/target/devhub-keycloak-event-listener.jar:/opt/keycloak/providers/devhub-keycloak-event-listener.jar:ro`
- `environment`:
  - `DEVHUB_BACKEND_SPI_WEBHOOK_URL: "http://backend-core:8080/api/v0-1/internal/keycloak-events"`
  - `DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET: "${DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET:-devhub-spi-secret}"` (colima default = `devhub-spi-secret` for test)

### 3.2 deploy (정합 ✓, PR #641+)

[`docker-compose.deploy.yml`](../setup/keycloak_event_listener_spi.md) 의 `keycloak` service:
- `volumes`: colima 와 동일
- `environment`:
  - `DEVHUB_BACKEND_SPI_WEBHOOK_URL: "${DEVHUB_BACKEND_SPI_WEBHOOK_URL:-http://backend-core:8080/api/v0-1/internal/keycloak-events}"`
  - `DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET: "${DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET:?set DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET}"` (사내 운영 vault mandatory)

## 4. realm config (정합 ✓, keycloak-realm.{dev,prod}.json)

[`infra/idp/keycloak-realm.dev.json`](../../infra/idp/keycloak-realm.dev.json) line 6 + [`infra/idp/keycloak-realm.prod.json`](../../infra/idp/keycloak-realm.prod.json) line 24:
- `eventsListeners: ["jboss-logging", "devhub-event-listener"]`
- 즉 SPI 가 realm import 시 자동 등록. Keycloak boot 후 admin console Events > Config 에서 "devhub-event-listener" 항목 존재 확인.

[`infra/idp/keycloak-realm.ci.json`](../../infra/idp/keycloak-realm.ci.json): CI test 환경 = SPI 의도적 미적용. `eventsListeners: []` (CI test = polling cron 만 검증, SPI webhook 미적용).

## 5. SPI 동작 flow (Java)

`DevHubEventListenerProvider` ([`infra/idp/keycloak-event-listener-spi/src/main/java/com/devhub/keycloak/spi/DevHubEventListenerProvider.java`](../../infra/idp/keycloak-event-listener-spi/src/main/java/com/devhub/keycloak/spi/DevHubEventListenerProvider.java)):
- `onEvent(Event userEvent)`: `Event` payload → JSON → `sendAsync POST`
- `onEvent(AdminEvent adminEvent, boolean includeRepresentation)`: `AdminEvent` payload → JSON (with `authDetails`) → `sendAsync POST`
- **noisy event skip**: `REFRESH_TOKEN`, `CODE_TO_TOKEN`, `INTROSPECT_TOKEN` (token rotation 부하 방지)
- **async**: `HttpClient.sendAsync` (Keycloak transaction blocking 방지)
- **timeout**: 3s (`HttpClient.connectTimeout` + `HttpRequest.timeout`)
- **secret header**: `X-Webhook-Secret` (env `DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET`)
- **error handling**: HTTP 200 외 warning log, exception error log

`DevHubEventListenerProviderFactory` 의 `init()` 가 env var `DEVHUB_BACKEND_SPI_WEBHOOK_URL` + `DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET` 사용 (cloud-native). env var 미설정 시 SPI 가 warning log + skip push (polling cron fallback).

## 6. 자동 검증 (scripts/verify-keycloak-spi.sh)

P2-6 + X-8 의 verify script (PR #641+ follow-up, 별도 sprint 가능):

```bash
KEYCLOAK_URL=https://kc.staging.internal/auth \
KC_BOOTSTRAP_ADMIN_USERNAME=admin \
KC_BOOTSTRAP_ADMIN_PASSWORD='<vault-managed>' \
DEVHUB_REALM=devhub \
DEVHUB_BACKEND_API_URL=http://backend-core:8080 \
DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET='<vault-managed>' \
  ./scripts/verify-keycloak-spi.sh
```

검증 항목 (4):
1. **SPI JAR classpath**: Keycloak container 내부 `/opt/keycloak/providers/devhub-keycloak-event-listener.jar` 존재 + `META-INF/services/org.keycloak.events.EventListenerProviderFactory` 에 `com.devhub.keycloak.spi.DevHubEventListenerProviderFactory` 등록
2. **Realm event listener 등록**: `GET /admin/realms/$REALM/events/config` → `eventsListeners: ["jboss-logging", "devhub-event-listener"]`
3. **Env var**: Keycloak container 의 `DEVHUB_BACKEND_SPI_WEBHOOK_URL` + `DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET` set
4. **Webhook push smoke**: Keycloak admin console 에서 test user login → `audit_logs` 의 `keycloak_event` source 1건 추가 (push latency < 1s 검증)

**exit code**:
- `0` — 4 항목 모두 OK → P2-6 + X-8 acceptance 충족
- `1` — 1건 이상 FAIL → FAIL detail stderr

## 7. 운영 SOP

### 7.1 1차 setup (P2-6 SPI JAR 처음 빌드 + deploy)

1. **빌드**: §2 의 `mvn clean package` 실행 → `infra/idp/keycloak-event-listener-spi/target/devhub-keycloak-event-listener.jar` 산출
2. **compose volume mount 확인**: §3.1/§3.2 의 mount entry 정합 (이미 PR #641+ 적용)
3. **Keycloak 재시작**: `docker compose restart keycloak` (or `docker compose up -d keycloak` 1차)
4. **realm import 확인**: Keycloak admin console → Realm `devhub` → Events > Config → "devhub-event-listener" listed
5. **verify**: §6 의 `scripts/verify-keycloak-spi.sh` 실행 → exit 0

### 7.2 일상 운영

- **SPI JAR rebuild 시**: §2 build → `docker compose restart keycloak` (volume mount = :ro 이므로 Keycloak container 가 새 JAR 자동 read 후 SPI class reload). **단**, SPI 가 active 한 상태에서 process 가 reload 안 될 수 있으므로 `restart` 권장.
- **secret rotation**: `DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET` (사내 vault) rotate 시 backend 의 동일 env 동기 갱신. webhook handler 가 `X-Webhook-Secret` header 검증하므로 rotate 후 양쪽 즉시 일치 필수.
- **webhook URL 변경**: nginx upstream 변경 시 `DEVHUB_BACKEND_SPI_WEBHOOK_URL` 갱신 + `docker compose restart keycloak`.
- **push fail** (HTTP 503/timeout): SPI 가 warning log + audit event 손실 → polling cron fallback (30s) 가 dedup cursor 갱신하므로 **audit event 보존**. 단 latency 30s+.

### 7.3 사내 운영 분리

- **로컬/colima 환경**: `DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET=devhub-spi-secret` (test default). 본 SOP §6 의 verify script 가 colima 의 container name 기준 검증.
- **운영 환경**: `DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET` 사내 vault 별도 관리. backend 의 동일 secret 환경변수 mandatory sync.

## 8. 잔여 carve

- **CI 환경의 SPI 미적용 명시** (keycloak-realm.ci.json 의 `eventsListeners: []`) — CI test = polling cron only 검증, SPI push webhook = e2e smoke 별도. follow-up sprint: CI 의 SPI e2e smoke spec 추가 (push latency < 1s 검증, 2026-07 정공법 결정).
- **residual polling cron** (`keycloak_event_puller.go`): SPI 가 push 안 됐거나 fail 한 경우 fallback. dedup cursor 가 push 시점까지의 cursor 만 polling → 30s latency 의 audit event 보존. polling cron 자체는 **살아있어야 함** (at-least-once deliver 보장). P3-5 의 polling-cron 제거 결정 시 ADR-0019 §5.3 (9) amendment + risk assessment 필수.
- **SPI JAR 의 unit test** (현재 0건): Java 의 HTTP client mock 으로 작성 가능 (sprint follow-up). e2e smoke 가 본 SOP §6 의 4번에서 통합 검증.

## 9. 정합

- P2-6 (SPI provider JAR, CodeReview #9 + ADR-0019 §5.3 carve P) → 본 SOP §2-§4 + compose mount
- X-8 (P3-5 audit event listener SPI push 전환) → 본 SOP §5-§7 + latency < 1s
- backend 변경 0 (webhook handler + polling cron + dedup 모두 main 정합)
- frontend 변경 0
- Tier: **공용** (Java + compose + SOP 만 — 사내 한정 정보 0)
