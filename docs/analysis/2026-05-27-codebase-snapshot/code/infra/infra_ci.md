# DevHub 인프라 / CI / 운영 자산 분석

- 문서 목적: DevHub 저장소의 인프라(`infra/`), CI 파이프라인(`.github/workflows/`), 운영 스크립트(`scripts/`), 관측/배포 자산을 코드 스냅샷 기준으로 정리한다.
- 범위: GitHub Actions CI/CD job 구성 + 인증 인프라(Keycloak SPI / nginx 단일포트) + 관측(Prometheus/Alertmanager/Grafana) + 배포 모드(native default / docker env-specific) + gRPC proto 상태.
- 대상 독자: 인프라/운영자, 신규 합류자, CI·배포 carve 담당자.
- 상태: analysis snapshot
- 기준 시점: 2026-05-27 (main HEAD `cf19c94`, post-#374)
- 분석 방식: 읽기 전용. 코드/설정 파일 직접 인용 (근거 경로:라인 표기).

---

## 1. CI 파이프라인 (`.github/workflows/ci.yml`)

### 1.1 트리거 + 동시성

- 트리거: `push` (branch `main`) + `pull_request` (base `main`). (`ci.yml:3-7`)
- 동시성: `ci-${{ github.workflow }}-${{ github.ref }}` 그룹 + `cancel-in-progress: true` — 같은 ref 의 이전 run 자동 취소. (`ci.yml:9-11`)

### 1.2 Job 표

| Job | name | 실행 조건 (`if`) | 핵심 동작 | 근거 |
| --- | --- | --- | --- | --- |
| `changed-paths` | Detect Changed Paths | (무조건) | `dorny/paths-filter@v3` 로 backend/frontend/e2e/workflow 4 output 산출 | `ci.yml:14-46` |
| `workflow-lint` | Workflow Lint (actionlint) | (무조건) | `raven-actions/actionlint@v2` — workflow YAML 정합 (ADR-0005) | `ci.yml:48-56` |
| `migration-prefix-lint` | Migration Prefix Uniqueness | `backend OR workflow` | `backend-core/migrations/*.up.sql` prefix(version) `uniq -d` 중복 검사 | `ci.yml:58-90` |
| `backend-unit` | Backend Unit Tests | `backend OR workflow` | Go 1.25 + `make test` (`go test ./...`) + Go 모듈/빌드 캐시 | `ci.yml:92-116` |
| `backend-integration` | Backend Integration Tests | `backend OR workflow` | native PG 15 (pgdg apt) + migrate up + `TestIntegration_` (`DEVHUB_TEST_DB_URL` gate) | `ci.yml:118-207` |
| `frontend-unit` | Frontend Unit Tests | `frontend OR workflow` | Node 20 + `npm ci` + `make test-frontend` (Vitest) | `ci.yml:209-225` |
| `e2e` | E2E Tests (Playwright, shard 1·2/2) | `e2e OR workflow` | matrix shard 2분할. native PG15 + Keycloak 컨테이너 + 실 OIDC 연동 | `ci.yml:227-568` |

### 1.3 paths-filter 정의 (`ci.yml:28-46`)

- `backend`: `backend-core/**`, `backend-ai/**`, `proto/**`, `Makefile`
- `frontend`: `frontend/**`, `proto/**`, `Makefile`
- `e2e`: `backend-core/**`, `backend-ai/**`, `frontend/**`, `infra/idp/**`, `tests/**`, `Makefile`
- `workflow`: `.github/workflows/**`

모든 test job 은 `needs: [changed-paths]` + `outputs.<filter> == 'true' || outputs.workflow == 'true'` 로 gate — workflow 자체 변경 시 전체 강제 실행.

### 1.4 backend-integration / e2e 의 native PostgreSQL 셋업 (ADR-0003 no-docker)

- ubuntu 러너 기본 PG14 cluster 를 `pg_dropcluster` 로 제거 → pgdg apt repo 로 PG15 설치 → 5432 재생성. (`ci.yml:142-166`, `253-292`)
- 5432 점유로 PG15 가 5433 에 생성되던 회귀(run 25774911487) 관찰 후 명시적 drop+recreate. (`ci.yml:277-283`)
- loopback `trust` auth 는 CI 전용. migrate 도구는 `golang-migrate v4.19.1` 을 `sudo GOBIN=/usr/local/bin go install`. (`ci.yml:168-175`, `482-486`)
- 통합 테스트 로그를 `tee` 로 `GITHUB_STEP_SUMMARY` 에도 적재 — raw log endpoint 차단 환경 우회. (`ci.yml:184-207`)

### 1.5 e2e job 의 Keycloak 실 연동 (`ci.yml:332-539`)

- Keycloak 컨테이너: `quay.io/keycloak/keycloak:26.0` `start-dev`, port 8180, `KC_HTTP_RELATIVE_PATH=/devhub/auth/keycloak`. (`ci.yml:341-351`)
- Keycloak 25/26 admin env 양쪽 동시 주입(`KC_BOOTSTRAP_ADMIN_*` + legacy `KEYCLOAK_ADMIN*`) — ADR-0022 25.0 retreat 잔재 호환. (`ci.yml:335-349`)
- realm `devhub` + role(developer/manager/system_admin) + client(`devhub-frontend` public, `devhub-backend` confidential service-account) 부트스트랩. (`ci.yml:353-480`)
- `devhub-backend` service account 에 `realm-management` 의 `view-users`/`query-users`/`manage-users`/`view-realm` 매핑(e2e seed 용) → 발급 secret 을 `GITHUB_ENV` 로 전파. (`ci.yml:438-480`)
- backend/frontend 를 native 기동 후 `/health`(60s) + `/`(120s) readiness wait. (`ci.yml:498-528`)
- `scripts/ci-e2e-sync-check.sh` 로 E2E-CI env 토큰 계약 검증. (`ci.yml:488-489`)
- artifact 3종: Playwright report / UI screenshot(18 페이지, 디자인 review source) / 실패 시 service log. (`ci.yml:541-567`)
- SPI webhook env 를 컨테이너에 주입(`DEVHUB_BACKEND_SPI_WEBHOOK_URL`, `DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET`)하나, 컨테이너 이미지는 stock(SPI JAR 미포함)이라 webhook push 는 실제 동작하지 않음 — §6.2 참조. (`ci.yml:343-345`)

### 1.6 Docker 이미지 발행 워크플로 (`.github/workflows/docker-image-publish.yml`)

- 트리거: `workflow_dispatch`(tag override input) + `push` tag `v*`. (`docker-image-publish.yml:3-12`)
- GHCR 로그인 후 backend-core / backend-ai / frontend 3 이미지를 `ghcr.io/<repo>` prefix 로 build+push. (`docker-image-publish.yml:52-84`)
- frontend 는 `BACKEND_API_URL` build-arg 만 inject. OIDC URL 은 build 에 박지 않고 런타임 `/api/runtime-config` + OIDC discovery 로 해결(ADR-0019 §4.2 / ADR-0018 §3.3). 과거 `NEXT_PUBLIC_OIDC_AUTH_URL`/`REDIRECT_URI` build-arg 는 2026-05-20 리뷰로 제거. (`docker-image-publish.yml:70-84`)

---

## 2. 인증 인프라 (Keycloak 단일 IdP — ADR-0019)

### 2.1 Keycloak Event Listener SPI (`infra/idp/keycloak-event-listener-spi/`)

- Java 21 Maven 모듈. provider id `devhub-event-listener`, `keycloak.version=26.0.0` provided scope. finalName `devhub-keycloak-event-listener`. (`pom.xml:7-53`)
- `DevHubEventListenerProvider` — user `onEvent` + admin `onEvent` 를 JSON payload 로 직렬화, `DEVHUB_BACKEND_SPI_WEBHOOK_URL` 로 비동기 POST(`sendAsync`, 3s timeout), `X-Webhook-Secret` 헤더 부착. `REFRESH_TOKEN`/`CODE_TO_TOKEN`/`INTROSPECT_TOKEN` 은 noise 로 skip. (`DevHubEventListenerProvider.java:30-136`)
- `DevHubEventListenerProviderFactory.init()` 가 env(`DEVHUB_BACKEND_SPI_WEBHOOK_URL`, `DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET`)에서 설정 resolve. URL 미설정 시 push disable(warning). (`DevHubEventListenerProviderFactory.java:24-35`)
- ServiceLoader 등록: `META-INF/services/org.keycloak.events.EventListenerProviderFactory`.
- `infra/idp/Dockerfile.keycloak` — 2-stage(maven build → `quay.io/keycloak/keycloak:26.0` providers 로 JAR COPY + `kc.sh build`). (`Dockerfile.keycloak:1-17`)
- backend 수신측: `POST /api/v1/internal/keycloak-events` (router.go:212, `keycloak_events_webhook.go`). secret 미설정 시 503 거부. config key `DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET` (`config.go:165`).
- realm import 파일이 listener 를 활성화: `eventsListeners: ["jboss-logging", "devhub-event-listener"]` (`keycloak-realm.dev.json:6`, `keycloak-realm.prod.json:24`).

> ADR-0020 의 SPI push 전환(polling 30s → realtime <1s) 의 사내 동반 자산. 현재 push 미동작 — §6.2 참조.

### 2.2 Keycloak 버전 pin (ADR-0022 / ADR-0023)

- 현행 운영 이미지 pin: `quay.io/keycloak/keycloak:26.0` (`docker-compose.deploy.yml:106`). 주석상 ADR-0023(2026-05-26) 이 ADR-0022 의 25.0 retreat 를 reversal — 사내 재확인으로 26.x 사용 가능 확정. (`docker-compose.deploy.yml:104-105`)
- CI e2e 도 26.0 사용. 단 25.0 호환 잔재로 legacy admin env 동시 주입 유지(§1.5). (`ci.yml:335-349`)
- 26.x management interface(`:9000/health/ready`, `KC_HEALTH_ENABLED=true`) 기반 healthcheck — distroless 라 bash `/dev/tcp` 사용. (`docker-compose.deploy.yml:144-152`)

### 2.3 단일 포트 reverse proxy (ADR-0018) — nginx

두 모드:

| 파일 | 용도 | 처리 |
| --- | --- | --- |
| `infra/nginx/devhub.deploy.conf.template` | compose 운영용 | nginx 공식 이미지 envsubst entrypoint 가 `${KEYCLOAK_UPSTREAM}`/`${KEYCLOAK_ADMIN_ALLOW_CIDR}` 치환 |
| `infra/nginx/devhub.native.conf` | host nginx + native loopback(`127.0.0.1:*`) dev/staging 정합 검증 | host 직접 설치 |
| `infra/nginx/devhub.deploy.conf` | template 의 렌더 사본(reference only, 비-docker용) | `scripts/nginx-conf-sync.sh` 가 생성 — DO NOT EDIT |
| `infra/nginx/default.conf` | stock default server 비활성화 stub | (`default.conf:1` 한 줄) |

- 단일 origin `/devhub/*` 로 frontend(`:3000`) / backend(`/devhub/api/*`→`/api/*`) / Keycloak(`/devhub/auth/keycloak/*`) 통합. (`devhub.deploy.conf.template:100-178`)
- Keycloak admin path(`^~ /devhub/auth/keycloak/admin/`) 는 `${KEYCLOAK_ADMIN_ALLOW_CIDR}` allow + `deny all` 로 internal-only. backend admin client 는 internal `keycloak:8080` 직접 통신. (`devhub.deploy.conf.template:120-142`, native: `devhub.native.conf:44-56`)
- **X-Forwarded-Host = `$http_host`** 명시 필수 — Next.js `request.nextUrl.origin` 이 이 헤더를 우선 사용. 누락 시 OIDC `redirect_uri` 가 외부 ingress port(예 :13000)가 아닌 :80 으로 잘못 생성되는 회귀(2026-05-26 사용자 보고). 모든 proxy location 에 적용. (`devhub.deploy.conf.template:66-70` 외 다수)
- legacy root auth route(`/login`, `/auth/callback`, `/auth/logout`)는 canonical `/devhub/*` 로 302. (`devhub.deploy.conf.template:47-56`)
- TLS 미사용 — compose/native 모두 HTTP `:80` 만. (`infra/nginx/README.md:74-77`)
- Keycloak 로컬/외부 upstream 분기는 `KEYCLOAK_UPSTREAM` env 만 변경(`keycloak:8080` 로컬 vs `kc.internal.example.com:8443` 외부). (`infra/nginx/README.md:24-49`)

### 2.4 Keycloak 셋업 스크립트 (`scripts/setup-keycloak.sh`)

- realm/role/client(`devhub-frontend` public + PKCE S256 + audience mapper, `devhub-backend` confidential service-account, `devhub-e2e-seeder`) idempotent 부트스트랩. `test` user 시드(`system_admin` role). (`setup-keycloak.sh:98-346`)
- redirect_uri wildcard 금지 + `DEVHUB_FRONTEND_ORIGIN` 미설정 시 fail-fast(단일 포트 컨셉 가드). (`setup-keycloak.sh:58-81`)
- `devhub-backend` service account role 을 `view-users`/`query-users`/`view-events`/`view-realm` 으로 최소화(`manage-users` 제거) — ADR-0020 sub-carve E 정합. (`setup-keycloak.sh:218-235`)
- `SETUP_KEYCLOAK_QUIET=1` 로 매 deploy sync 시 secret 평문 stdout 누적 회피(issue #302). (`setup-keycloak.sh:360-367`)
- `scripts/verify-keycloak-groups.sh` — group 4종 + composite role 1:1 + Default Groups 비어있음 read-only 검증(issue #214 acceptance, ADR-0019 §5.3). (`verify-keycloak-groups.sh:16-206`)

### 2.5 IdP 자산 인벤토리 (`infra/idp/`)

- `keycloak-realm.dev.json` — 로컬/CI/smoke realm import(localhost wildcard, 운영 금지). `keycloak-realm.prod.json` — 외부 운영팀 reference 템플릿. (`infra/idp/README.md:18-24`)
- `identity.schema.json` — Kratos 시기 legacy schema(보존). `sql/001`,`sql/003` — IdP schema + test admin seed.
- `_archive_hydra_kratos/` — ADR-0001 시기 Hydra/Kratos 자산(deprecated, ADR-0019 supersedes).

---

## 3. 관측 (Observability)

### 3.1 backend-core `/metrics` 엔드포인트

- `router.go:210` — `GET /metrics` 가 `promhttp.Handler()` wrap. 모든 도메인 metric 자동 노출.

### 3.2 Prometheus metric 인벤토리 (코드 등록 기준)

| 패키지 | metric | type | 근거 |
| --- | --- | --- | --- |
| auth | `devhub_jwks_stale_while_error_total{result}` | CounterVec | `internal/auth/metrics.go:27` |
| auth | `devhub_jwks_stale_age_seconds` | Histogram | `metrics.go:34` |
| audit | `devhub_keycloak_events_processed_total{kind,action}` | CounterVec | `internal/audit/metrics.go:33` |
| audit | `devhub_keycloak_event_cursor_lag_seconds{cursor_key}` | GaugeVec | `metrics.go:40` |
| audit | `devhub_keycloak_event_pull_errors_total{kind}` | CounterVec | `metrics.go:47` |
| audit | `devhub_keycloak_user_sync_total{action}` / `_errors_total{action}` / `_lag_seconds` | CounterVec×2 + Histogram | `metrics.go:59-79` |
| devrequest | `devhub_intake_token_expiring_soon` / `_stale` / `_auto_revoked_total` | Gauge×2 + Counter | `internal/devrequest/metrics.go:28-45` |
| integrations/adapters | `devhub_homelab_pull_runs_total{result}` / `_pull_duration_seconds` / `_snapshot_services` / `_degraded_providers` / `_last_success_unixtime` | Counter+Histogram+Gauge×3 | `internal/integrations/adapters/metrics.go:23-54` |
| httpapi(onboarding) | `devhub_onboarding_gate_blocked_total{reason}` / `_submit_total{status}` / `_submit_duration_seconds` / `_review_confirm_total{status}` / `_pending_review_count` | CounterVec×3 + Histogram + Gauge | `internal/httpapi/onboarding_metrics.go:46-79` |

- label cardinality bounded 패턴 일관 적용(`normalizeMetricAction` 등 unknown → unified). (`audit/metrics.go:107-112`)

### 3.3 Prometheus / Alertmanager (`docs/setup/prometheus_alertmanager_setup.md`)

- 책임 분리: devhub repo = 의도/임계 source-of-truth(ADR-0016 + 본 가이드 + raw YAML reference). 실 운영 자산(`prometheus.yml`/`alertmanager.yml`/rules)은 **별도 ops git/vault** — env-specific 정책 정합. (`prometheus_alertmanager_setup.md:14-37`)
- scrape job reference: `devhub-backend-core`, `metrics_path: /metrics`, 30s interval, `environment` label.
- alert rule(prod/stage 분리): `DevhubHomeLabPullFailing` / `NoRecentSuccess` / `DegradedProvidersDetected` + DREQ 토큰 3종(`ExpiringSoon`/`StaleDetected`/`AutoRevokeBurst`).
- codex hotfix #8: 모든 expr 에 `{environment="prod|stage"}` matcher(cross-env contamination 회피) + `max by(provider)` aggregation(multi-instance staleness). (`prometheus_alertmanager_setup.md:64-66`)
- Alertmanager 라우팅: prod-critical→PagerDuty, prod→ops-slack, stage→stage-slack(노이즈 감소). secret 은 vault 주입.
- ADR-0016 §6 잔여 carve: (3) pull latency p95 alert / (4) push 경로(API-73 webhook) 알림 / (5) stage→prod 임계 확정.

### 3.4 Grafana (`docs/setup/grafana/homelab_dashboard.json`)

- HomeLab dashboard JSON 모델 1건 존재(import 대상, ops git 사본). Prometheus 가이드가 dashboard 를 별도 자산으로 분리 참조. (`prometheus_alertmanager_setup.md:4`)

---

## 4. 배포 모드

### 4.1 native default (ADR-0003 no-docker)

- 설치 `make setup`(go mod tidy + pip + npm install). 빌드/실행은 환경별이라 `make build`/`run` 은 안내 echo 만. (`Makefile:19-45`)
- 테스트 target: `test`(go), `test-race`, `test-coverage`, `test-frontend`(Vitest), `e2e`(Playwright). (`Makefile:56-70`)
- migrate target: `migrate-up`/`down`/`create`/`version`(golang-migrate v4.19.1). (`Makefile:12-35`)
- `dev-up.sh`/`dev-down.sh`(+ `.ps1`) — native 로 backend(8080)/frontend(3000)/Keycloak(8180) 기동. (root)
- CI 도 ADR-0003 정합으로 PG/서비스 모두 native(컨테이너 sidecar 제거). (`ci.yml:236-239`)

### 4.2 docker (환경 특화, git 추적 외)

- `docker-compose.deploy.yml` 는 사내 운영팀 reference(env 만 변경). 서비스: db/db-init/db-migrate/keycloak/backend-ai/backend-core/frontend/nginx. profile `local-db`/`local-idp`. (`docker-compose.deploy.yml:30-295`)
- compose 자산(`docker-compose.yml` 등 각 서비스 Dockerfile)은 `.gitignore` DEV ENVIRONMENT 섹션으로 git 추적 외(환경별 관리). 단 `docker-compose.deploy.yml` + `Dockerfile.keycloak` 은 추적 포함(reference).
- `scripts/build-artifacts.sh` — **host build pattern**(docker network 의존 제거). go1.25 정적 binary + python3.12 정확 ABI + Next.js standalone. 사내 proxy 가 container 안으로 전파 안 되는 시나리오 회피. Dockerfile 은 COPY-only. (`build-artifacts.sh:1-180`)
- `scripts/deploy-from-env.sh` — one-shot deploy 오케스트레이터(env 파일 생성 → build-artifacts → docker build → deploy-up → Keycloak redirect sync). DB_MODE external|docker, 단일포트 issuer/JWKS 분리 처리. (`deploy-from-env.sh:1-417`)
- `scripts/deploy-preflight.sh` — deploy 전 검증: nginx conf sync(`--fix`+`--check`) / compose render / OIDC issuer+JWKS reachability / CIDR / redirect_uri 일치 / `DEVHUB_AUTH_DEV_FALLBACK=0` 강제. (`deploy-preflight.sh:151-213`)
- `scripts/deploy-up.sh` — preflight 호출 후 compose pull+up+ps. local image 면 pull skip. (`deploy-up.sh:15-54`)
- `scripts/nginx-conf-sync.sh` — template→.conf envsubst 렌더 + drift 검사(`--check` exit1). 렌더본은 reference only(compose 는 template 직접 mount). (`nginx-conf-sync.sh:1-95`)
- 외부 ingress port 기본 13000(ADR-0022 §3.4 정합). (`deploy-from-env.sh:24`)

---

## 5. proto / gRPC

- `proto/analysis.proto` — `AnalysisService`(`AnalyzeBuildLog`, `DetectRisks`) proto3 정의 1건. `go_package = github.com/devhub/backend-core/proto/analysis`. (`analysis.proto:1-32`)
- `Makefile:15-17` `proto` target 이 protoc(go + python grpc_tools) 생성 명령 보유. proto-tools install target 존재. (`Makefile:8-17`)
- backend-ai `requirements.txt` 에 `grpcio`, `grpcio-tools` 선언. (line 3-4)
- **그러나 gRPC 서버/스텁 미구현** — §6.3 참조.

---

## 6. 발견 사항 (불일치 / stale / 부채)

### 6.1 CI migration prefix guard 가 CI bypass 머지를 못 잡음 (이력)

- guard 자체는 `ci.yml:58-90` 에 정상 존재(`uniq -d` prefix 중복 검사, `if backend || workflow`).
- 그러나 2026-05-27 PR #363↔#368 이 동시에 migration `000042` 를 잡았고, **#368 이 CI bypass(merge-commit + check 우회)로 머지되어 guard 가 평가되지 못함** → main 에 prefix 중복 유입. 사후 #371 전수점검에서 적발, #368 분을 `000044` 로 재번호. (근거: MEMORY.md post-#371 항목 + `ci.yml` guard 가 PR check 로만 작동 / `feedback_concurrent_migration_prefix_collision`)
- 부채 성격: guard 는 PR check 에 의존하므로 CI bypass(admin merge / required check 미설정) 시 무력. branch protection 의 required status check 강제가 동반되어야 guard 가 실효.

### 6.2 Keycloak SPI realm events 사내 미등록 (P3-5 / P2-6) — push 미동작

- realm import JSON 은 `eventsListeners` 에 `devhub-event-listener` 를 선언(`keycloak-realm.dev.json:6`, `.prod.json:24`)하나, **운영 compose 가 stock 이미지(`quay.io/keycloak/keycloak:26.0`)를 사용**하고 `Dockerfile.keycloak`(SPI JAR 빌드 이미지)를 쓰지 않음. (`docker-compose.deploy.yml:106`)
- 따라서 SPI provider(`devhub-event-listener`)가 컨테이너에 부재 → realm 이 존재하지 않는 listener 를 참조하는 상태. SPI webhook env(`DEVHUB_BACKEND_SPI_WEBHOOK_URL`/`_SECRET`)도 compose keycloak 서비스에 미주입.
- CI e2e 는 webhook env 를 주입(`ci.yml:343-345`)하나 컨테이너 이미지가 stock 이라 역시 push 미동작 — backend 수신 핸들러(`/api/v1/internal/keycloak-events`)만 wire.
- 미해결 carve 로 등재: release_v1_roadmap.md P2-6(`devhub-event-listener` Maven 빌드 + compose volume mount + 운영 SOP, "Codex(infra) + 사용자(Java 빌드 환경)"), session_handoff "Keycloak SPI realm events 등록 + `DEVHUB_BACKEND_SPI_WEBHOOK_URL` wire(#340 codex P1×2)". (근거: `release_v1_roadmap.md:136`, `account_user_management_redesign.md:470`)
- 영향: ADR-0020 SPI push 전환(realtime <1s)이 미완 — 현재는 backend cron polling(30s) 경로만 실 동작.

### 6.3 proto 미사용 — backend-ai gRPC 미구현

- `backend-ai/main.py` 는 FastAPI `/health` 만 노출하고 gRPC 서버는 TODO 주석뿐: `# TODO: gRPC Server for AnalysisService` / `# TODO: AI Logic for Log Analysis`. (`backend-ai/main.py:10-11`)
- 생성 코드 부재 확인: `backend-core/**/*.pb.go` 없음, `backend-ai/**/*_pb2*.py` 없음(Glob 결과 No files found). `make proto` 는 정의되어 있으나 산출물이 commit 되지 않음.
- backend-core↔backend-ai 연동은 `BACKEND_AI_URL`(HTTP `:8000`) 기반(compose env). 즉 proto/gRPC 는 미래 방향 placeholder 이며 현재 런타임 경로에 미사용.

### 6.4 hrdb_etl_sync.sh deprecated 잔재

- `scripts/hrdb_etl_sync.sh` 헤더가 명시 DEPRECATED(2026-05-20, issue #215): DevHub 가 외부 Keycloak 시나리오 채택 → HR↔Keycloak sync 는 사내 IdP 팀 책임. DevHub 자체 cron 미등록. historical reference 로만 보존. (`hrdb_etl_sync.sh:4-21`)
- 대체 경로: Keycloak event listener(backend cron, PR #241) → `user_sync.go::SyncUserProfile` → `users.status='deactivated'` 자동 sync.
- 부채 성격: 실행 코드(full ETL loop)가 deprecated 명시만 된 채 저장소에 잔존. `scripts/hrdb_etl_seed.sql`(ADR-0008 PoC seed loader) 도 "operators expected to maintain a private fork" 안내 상태로 PoC 잔재. (`hrdb_etl_seed.sql:11-17`)

### 6.5 nginx WebSocket auth query token redact 사내 미적용 (ADR-0024 §6 carve 2)

- `/devhub/api/v1/realtime/ws` 인증 토큰이 query string(`?ticket=` 우선, deprecated `?access_token=`)으로 전달되어 nginx access_log 기본 format($request)에 leak. (`infra/nginx/README.md:90-94`)
- README §6.1 이 권장 `log_format devhub_safe`(query redact) / §6.2 location `access_log off` 대안을 문서화하나, **실제 nginx conf(`devhub.deploy.conf.template`/`devhub.native.conf`)에는 redact 미반영** — http block 변경 + 재기동이 사내 nginx 운영자 영역. (`infra/nginx/README.md:139-141`)
- 완화: ticket 은 single-use + 60s TTL 이라 capture-replay risk 낮음. 단 deprecated access_token query 는 만료 전까지 valid Bearer 동등 — redact 필요. 적용 시점은 사내 SLA/log retention 정책 의존(미적용 부채).

### 6.6 ci-setup.sh 참조 stale

- `ci.yml:171`, `ci.yml:258`, `scripts/setup-test-db.sh:20` 가 `scripts/ci-setup.sh` 를 정합 근거로 주석 참조하나 **해당 파일이 저장소에 부재**(Glob `scripts/**` 결과에 없음). 과거 분리되어 있던 setup 스크립트가 ci.yml inline step 으로 흡수된 뒤 주석만 잔존한 stale reference. 동작 영향은 없음(주석일 뿐 invoke 안 함).

### 6.7 frontend `NEXT_PUBLIC_APP_ORIGIN` dead set (문서화된 의도적 부채)

- `docker-compose.deploy.yml:237` 의 `NEXT_PUBLIC_APP_ORIGIN` runtime set 은 dead set — `NEXT_PUBLIC_*` 는 build-time inline 이라 runtime env 무효. client 는 `window.location.origin` 으로 fallback. 주석에 의도(image 환경간 재사용) 명시됨(PR #278 review P2 #1). 실제 부채는 아니나 신규 운영자 오독 risk. (`docker-compose.deploy.yml:227-237`)

---

## 7. 요약

- CI 는 8 job(changed-paths gate + actionlint + migration prefix + backend unit/integration + frontend unit + e2e shard×2)으로 paths-filter 기반 선택 실행. e2e 는 native PG15 + Keycloak 26.0 실 OIDC 연동까지 포함.
- 인증 인프라는 Keycloak 단일 IdP(ADR-0019) + 단일포트 nginx(ADR-0018, X-Forwarded-Host 정합) + 버전 26.0 pin(ADR-0023). SPI 자산은 빌드 가능 상태이나 운영 compose 미배선.
- 관측은 backend `/metrics` 에 5개 도메인 metric 노출 + Prometheus/Alertmanager 규칙은 의도(devhub) ↔ 운영자산(ops git) 분리. Grafana dashboard 1건.
- 배포는 native default(ADR-0003) + docker 환경특화(git 추적 외), host build pattern 으로 사내 proxy 회피.
- 주요 부채 6.1~6.7: migration guard 의 CI-bypass 사각 / SPI push 미배선 / gRPC 미구현 / hrdb ETL deprecated 잔재 / nginx token redact 미적용 / ci-setup.sh stale 참조.
