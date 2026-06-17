# X-8 Staging Hand-off SOP (사내 SCM tier-2, 2026-06-17)

문서 목적: P2-6 (Keycloak SPI provider JAR) + X-8 (P3-5 audit event listener push 전환) 의 **staging/prod 사내 SCM (tier-2)** hand-off 절차. main (sabun, tier-1) 에서 SPI 의 정공법 + compose mount + verify script + SOP 가 이미 정합 (PR #641 + 본 PR). 본 SOP 는 **staging/prod 환경에 SPI 를 적용하기 위한 단계별 절차**.

관련 문서:
- [`docs/setup/keycloak_event_listener_spi.md`](keycloak_event_listener_spi.md) — SPI 정공법 (architecture, build, compose mount, realm config, verify)
- [`infra/idp/keycloak-event-listener-spi/`](../../infra/idp/keycloak-event-listener-spi/) — SPI source
- [`scripts/verify-keycloak-spi.sh`](../../scripts/verify-keycloak-spi.sh) — 4 항목 자동 검증
- [`scripts/build-keycloak-spi.sh`](../../scripts/build-keycloak-spi.sh) — JAR build + artifact push
- ADR-0019 §5.3 (9) + ADR-0022 §4.1
- sprint `feat/260617-x8-staging-handoff-e2e-smoke` (PR #642+)

## 1. Hand-off 전제 (sabun main = tier-1)

sabun main (2026-06-17 PR #641 merge) 의 정합:
- **Java SPI source**: [`infra/idp/keycloak-event-listener-spi/`](../../infra/idp/keycloak-event-listener-spi/) (Keycloak 26.0 + Java 21)
- **Compose mount**:
  - `docker-compose.colima.yml` (colima/local) — volume mount + env (default `devhub-spi-secret`)
  - `docker-compose.deploy.yml` (deploy/sabun) — volume mount + env (vault mandatory)
- **Realm config**: `keycloak-realm.{dev,prod}.json` 의 `eventsListeners: ["jboss-logging", "devhub-event-listener"]`
- **Verify script**: [`scripts/verify-keycloak-spi.sh`](../../scripts/verify-keycloak-spi.sh) 4 항목 (SPI JAR classpath / Realm eventsListeners / Env var / Webhook push smoke)
- **Dockerfile.keycloak** (multi-stage): Stage 1 Maven build (spi-builder) → Stage 2 Keycloak + SPI JAR copy + `kc.sh build` (auto SPI feature)

## 2. Hand-off 단계

### 단계 1: 사전 준비 (사용자 결정)

- [ ] **Java 21 + Maven 3.13+ 환경**: 사내 staging VM 또는 빌드 컨테이너
- [ ] **Docker + Docker Compose**: 사내 staging VM
- [ ] **Keycloak admin token**: staging Keycloak admin console 접근
- [ ] **Vault 접근**: `DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET` (colima default `devhub-spi-secret` 와 다른 staging-specific secret 권장)
- [ ] **backend webhook URL**: staging backend container 의 `/api/v0-1/internal/keycloak-events` endpoint (또는 nginx upstream)
- [ ] **staging compose file**: staging 사내 SCM 의 `docker-compose.deploy.yml` 또는 staging-specific override

### 단계 2: SPI JAR build + push (사용자 실행)

```bash
# option A: 표준 maven build
cd infra/idp/keycloak-event-listener-spi
mvn clean package
ls -la target/devhub-keycloak-event-listener.jar
# → 산출물: target/devhub-keycloak-event-listener.jar (~20KB)

# option B: build script (sabun 의 scripts/build-keycloak-spi.sh)
KEYCLOAK_SPI_VERSION=1.0.0 \
KEYCLOAK_SPI_REGISTRY=harbor.internal/devhub \
bash scripts/build-keycloak-spi.sh
# → 도커 이미지 build + push (harbor.internal/devhub/devhub-keycloak-spi:1.0.0)
```

### 단계 3: compose mount + env 갱신 (사용자 결정)

sabun 의 `docker-compose.deploy.yml` 의 keycloak service 적용 (PR #641 의 keycloak service 정합 참고):

```yaml
keycloak:
  profiles: ["local-idp"]
  build:
    context: ./infra/idp
    dockerfile: Dockerfile.keycloak
  environment:
    # ... 기존 ...
    # P2-6 + X-8 정공법 — SPI webhook env (PR #641 정합)
    DEVHUB_BACKEND_SPI_WEBHOOK_URL: "${DEVHUB_BACKEND_SPI_WEBHOOK_URL:-http://backend-core:8080/api/v0-1/internal/keycloak-events}"
    DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET: "${DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET:?set DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET}"
  volumes:
    # P2-6 + X-8 — SPI JAR mount
    - ./infra/idp/keycloak-event-listener-spi/target/devhub-keycloak-event-listener.jar:/opt/keycloak/providers/devhub-keycloak-event-listener.jar:ro
```

staging 사내 SCM 의 compose file 이 sabun 과 다를 경우 같은 pattern 적용:
- **JAR path**: `./infra/idp/keycloak-event-listener-spi/target/devhub-keycloak-event-listener.jar` 또는 CI build artifact
- **env**: staging backend container name 기준 + staging secret (vault)

### 단계 4: Keycloak 재시작 (사용자 실행)

```bash
docker compose -f docker-compose.deploy.yml up -d --force-recreate keycloak
# 또는 staging 사내 SCM 의 staging-specific compose file
docker compose -f docker-compose.staging.yml up -d --force-recreate keycloak

# healthcheck 확인 (P2-6 + X-8 acceptance 의 1차 gate)
docker compose -f docker-compose.deploy.yml ps keycloak
# STATUS: Up (healthy) — Keycloak 가 boot + SPI auto-load + realm import 완료
```

### 단계 5: verify 자동 실행 (사용자 실행)

```bash
KEYCLOAK_URL=https://kc.staging.internal/auth \
KC_BOOTSTRAP_ADMIN_USERNAME=admin \
KC_BOOTSTRAP_ADMIN_PASSWORD='<vault-managed>' \
DEVHUB_REALM=devhub \
DEVHUB_BACKEND_API_URL=http://backend-core:8080 \
DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET='<vault-managed>' \
TEST_USER_USERNAME=staging-test-user \
TEST_USER_PASSWORD='<vault-managed>' \
  ./scripts/verify-keycloak-spi.sh
```

검증 항목 (4):
1. **SPI JAR classpath**: Keycloak container 내부 `/opt/keycloak/providers/devhub-keycloak-event-listener.jar` 존재 + `META-INF/services/org.keycloak.events.EventListenerProviderFactory` 에 `DevHubEventListenerProviderFactory` 등록
2. **Realm event listener 등록**: `GET /admin/realms/$REALM/events/config` → `eventsListeners: ["jboss-logging", "devhub-event-listener"]`
3. **Env var**: `DEVHUB_BACKEND_SPI_WEBHOOK_URL` + `DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET` set
4. **Webhook push smoke**: test user login → `audit_logs` 의 `keycloak_event` source 1건 추가 (push latency < 1s)

**exit code 0** = P2-6 + X-8 acceptance 충족. **exit 1** = §6 의 trouble-shooting.

### 단계 6: 결과 보고 + sabun main (tier-1) feedback

sabun 의 `docs/traceability/report.md` §6 의 본 sprint row 갱신 (staging 1차 적용 결과 + verify 4 항목 PASS + latency < 1s). tier-1 의 **정합 evidence** 가 됨.

## 3. Prod 환경 hand-off

staging 에서 verify 4 항목 PASS + push latency < 1s 확인 후 **prod 환경** hand-off:

```bash
# 동일한 단계 1-5 반복 (sabun main 의 정공법 그대로)
KEYCLOAK_URL=https://kc.prod.internal/auth \
# ...
```

staging → prod 의 차이 = **secret rotation** (vault 별도 secret) + **compose file 의 backend container name 다름** (prod 의 internal DNS). `DEVHUB_BACKEND_SPI_WEBHOOK_URL` 의 prod backend URL 로 갱신.

## 4. 일상 운영 (post hand-off)

- **SPI JAR rebuild 시**: §2 의 `mvn clean package` + `docker compose up -d --force-recreate keycloak` (volume mount `:ro` 이므로 Keycloak container 가 새 JAR 자동 read 후 SPI class reload)
- **secret rotation**: vault 의 `DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET` rotate 시 backend 의 동일 env 동기 갱신. webhook handler 가 `X-Webhook-Secret` header 검증하므로 rotate 후 양쪽 즉시 일치 필수
- **webhook URL 변경**: nginx upstream 변경 시 `DEVHUB_BACKEND_SPI_WEBHOOK_URL` 갱신 + `docker compose restart keycloak`
- **push fail (HTTP 503/timeout)**: SPI 가 warning log + audit event 손실 → polling cron fallback (30s) 가 dedup cursor 갱신하므로 **audit event 보존**. 단 latency 30s+

## 5. Trouble-shooting

| 증상 | 원인 | 해결 |
|---|---|---|
| `verify-keycloak-spi.sh [1/4] FAIL: SPI JAR missing` | 빌드 산출물 부재 (`target/devhub-keycloak-event-listener.jar` 없음) | §2 의 `mvn clean package` 재실행 |
| `verify-keycloak-spi.sh [1/4] FAIL: SPI service registration missing` | JAR 의 META-INF/services 부재 (build 오류) | `mvn clean package` + JAR 의 `META-INF/services/org.keycloak.events.EventListenerProviderFactory` 내용 확인 |
| `verify-keycloak-spi.sh [2/4] FAIL: Realm eventsListeners missing` | `keycloak-realm.{staging,prod}.json` 의 eventsListeners 미포함 | `eventsListeners: ["jboss-logging", "devhub-event-listener"]` 추가 + realm 재import |
| `verify-keycloak-spi.sh [3/4] FAIL: env var not set` | compose 의 `DEVHUB_BACKEND_SPI_WEBHOOK_URL` / `DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET` 미주입 | compose 갱신 + Keycloak 재시작 |
| `verify-keycloak-spi.sh [4/4] FAIL: backend audit_logs did not receive push event` | webhook URL / secret 불일치 | `DEVHUB_BACKEND_SPI_WEBHOOK_URL` 의 backend URL + `DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET` 의 vault 동기화 + backend 의 `keycloak_events_webhook.go` 의 secret 검증 코드 확인 |
| `keycloak:healthcheck fail after compose up` | SPI JAR 의 classpath conflict | `docker logs keycloak` 의 SPI init error 확인 → SPI factory 의 `init()` 가 `webhookUrl == null` 시 warning log 만 = 정상 |

## 6. 잔여 carve

- **CI 환경의 SPI e2e smoke** (의도적 미적용): `docker-compose.test.yml` 의 keycloak service 는 polling cron only 검증. SPI push 검증 = 별도 e2e smoke spec (`backend-e2e/specs/keycloak-event-listener.spec.ts` + SPI mount in compose.test + push latency < 1s 검증). **본 PR 의 scope 외** — follow-up sprint 결정 (sabun 의 tier-1 정합 후).
- **SPI JAR unit test** (현재 0건): Java 의 HTTP client mock 으로 작성 가능. e2e smoke 가 본 SOP §2 의 4번에서 통합 검증.
- **residual polling cron** (`keycloak_event_puller.go`): SPI 가 push 안 됐거나 fail 한 경우 fallback. dedup cursor 가 push 시점까지의 cursor 만 polling → 30s latency 의 audit event 보존. polling cron 자체는 **살아있어야 함** (at-least-once deliver 보장). P3-5 의 polling-cron 제거 결정 시 ADR-0019 §5.3 (9) amendment + risk assessment 필수.

## 7. 정합

- P2-6 (SPI provider JAR) + X-8 (P3-5 push 전환) 의 staging/prod 환경 hand-off
- backend 변경 0 (webhook handler + polling cron + dedup 모두 main 정합)
- frontend 변경 0
- **Tier**: **사내** (staging/prod hand-off SOP, 사내 vault secret + 사내 staging URL 포함)
- **sabun main (tier-1) 변경 0** (본 SOP 는 docs only, scripts/build-keycloak-spi.sh + scripts/verify-keycloak-spi.sh 는 script only)
