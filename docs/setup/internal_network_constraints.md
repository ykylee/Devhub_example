# 사내 네트워크 제약 운영 가이드

- 문서 목적: DevHub 사내 운영 환경의 network 제약 3 항목을 단일 docs 로 정리. 각 제약별 자산 매핑 + 분기 매트릭스 + cross-link.
- 범위: build / image / deploy 의 host build pattern 강제 / 외부 ingress port forward / db + Keycloak 내부·외부 분기.
- 대상 독자: 사내 운영자 (DevHub SRE / Infra), 신규 진입 개발자, 외부 환경 reference 운영자.
- 상태: active
- 최종 수정일: 2026-05-26 (sprint `claude/work_260526-internal-network-constraints-docs`)
- 관련 문서: [docker-packaging-deployment-guide.md](./docker-packaging-deployment-guide.md) (메인 가이드), [ADR-0022 §3.4 port 13000 정합](../adr/0022-keycloak-version-pin-25-0.md#34-외부-ingress-포트-13000-정합), [ADR-0023 Keycloak 26.0 forward pin](../adr/0023-keycloak-version-pin-26-0.md), [infra/idp/README.md](../../infra/idp/README.md), [keycloak_operations.md](./keycloak_operations.md), [deploy.env.example](./deploy.env.example).

## 1. 제약 1 — docker 내 빌드 불가 (host build pattern 강제)

### 1.1 제약

사내 환경의 docker container 안에서 외부 PyPI / npm registry / GoProxy 접근 불가 — host 의 HTTP proxy 가 docker container 안으로 전파 안 됨. multi-stage Dockerfile (RUN go build / RUN npm ci / RUN pip install) 사용 불가.

### 1.2 적용 자산

| 자산 | 적용 |
| --- | --- |
| `scripts/build-artifacts.sh` | host 에서 모든 build artifact 산출 (Go binary / Python deps / Next.js standalone). `verify_prerequisites()` 사전 검증 (go 1.25+ / python3.12 정확히 / node 20+ / npm). dockerized fallback 제거 (PR #310, 2026-05-26). |
| `backend-core/Dockerfile` | COPY-only — `FROM ${BACKEND_CORE_BASE:-alpine:3.21}` + `COPY bin/main` + `COPY migrations/` |
| `backend-ai/Dockerfile` | COPY-only — `FROM ${BACKEND_AI_BASE:-python:3.12-slim}` + `COPY .build/site-packages` + `COPY main.py` |
| `frontend/Dockerfile` | COPY-only — `FROM ${FRONTEND_BASE:-node:20-alpine}` + `COPY .next/standalone` + `COPY .next/static` |
| Dockerfile `ARG <SERVICE>_BASE` | 사내 mirror registry tag override 가능 (PR #316). `--build-arg BACKEND_CORE_BASE=internal-registry.example.com/alpine:3.21` 등. 자세한 절차: [docker-packaging-deployment-guide.md §5.1](./docker-packaging-deployment-guide.md#51-사내-mirror-registry-사용-base-image-pull-차단-우회). |

### 1.3 proxy 환경 동작

| 단계 | proxy 의존성 |
| --- | --- |
| 1. host build (`build-artifacts.sh`) | host 의 `HTTP_PROXY` / `HTTPS_PROXY` / `NO_PROXY` env 가 go / npm / pip child process 로 **자연 전파**. docker 미사용. |
| 2. docker image build | base image pull 만 — `~/.docker/config.json` 또는 `/etc/systemd/system/docker.service.d/http-proxy.conf` 의 **docker daemon proxy 필요**. 또는 §1.2 의 ARG override 로 사내 mirror tag 사용 시 docker daemon proxy 불필요. |
| 3. deploy (`docker compose up`) | image pull (remote registry) 외 외부 접근 없음. local image 사용 시 `SKIP_PULL=1` 또는 `IMAGE_REPO_PREFIX=local/*`. |

### 1.4 fail-mode 진단

`build-artifacts.sh` 의 `verify_prerequisites()` 가 missing 도구 발견 시 친절한 에러 + 설치 안내 + `exit 1`. silent fail 없음. 자세한 troubleshooting: [docker-packaging-deployment-guide.md §13](./docker-packaging-deployment-guide.md#13-build--deploy-troubleshooting-matrix) (10 시나리오).

## 2. 제약 2 — 외부 ingress port forward (외부 13000 → 내부 3000 → 도커)

### 2.1 제약

사내 외부 client (사내 PC) 는 **호스트 머신의 13000 port** 로 접근. 호스트의 13000 → VM 내부의 nginx **3000 port** 로 forward. VM 내부 nginx 는 docker network 의 `frontend` / `backend-core` / `keycloak` 으로 reverse proxy.

```
[사내 PC]
    ↓ http://<host-ip>:13000/devhub/...
[호스트 머신 :13000]                          ← 사내 인프라 설정 (SSH tunnel / NAT / iptables)
    ↓ port forward
[VM 내부 nginx :3000]                         ← compose NGINX_HTTP_PORT=3000
    ↓ reverse proxy
[docker network: frontend:3000 / backend-core:8080 / keycloak:8080]
```

### 2.2 적용 자산

| 자산 | 적용 |
| --- | --- |
| ADR-0022 §3.4 (immutable, [ADR-0023](../adr/0023-keycloak-version-pin-26-0.md) supersession 후 보존) | port 13000 의 코드 매핑 표 — realm.dev.json 6 위치 + deploy-from-env.sh PUBLIC_ACCESS_PORT default + NGINX_HTTP_PORT 3000 default + host:VM forward 외부 의존성 명시 |
| `scripts/deploy-from-env.sh:14` | `PUBLIC_ACCESS_PORT:=13000` default — **외부 client 의 진입 port** |
| `scripts/deploy-from-env.sh:34` | `NGINX_HTTP_PORT:=3000` default — **VM 내부 nginx 가 host 에 bind 하는 port** |
| `infra/idp/keycloak-realm.dev.json` | `devhub-frontend` client 에 `http://localhost:13000` 6 entry — 사내 ingress 시뮬레이션 (redirectUris 3 + webOrigins 1 + post.logout 2) |
| `scripts/sync_keycloak_redirects()` | 매 deploy 후 Keycloak 의 redirectUris / webOrigins 를 `PUBLIC_ACCESS_*` env 기반으로 동기화 |

### 2.3 사내 인프라 측 설정 (compose 외부 의존)

`docker-compose.deploy.yml` 의 nginx service 는 `NGINX_HTTP_PORT` 의 **VM 내 host bind** 만 처리한다. 호스트의 13000 ↔ VM 내부 3000 의 forward 는 **사내 인프라 측 별도 설정** 에 의존:

- **SSH tunnel**: `ssh -L 3000:localhost:3000 user@vm-host` (사내 PC 에서 호스트로) + `-L 13000:vm-host:3000` (호스트에서 VM 으로)
- **NAT / iptables**: 호스트의 13000 → VM 의 3000 redirect rule
- **virtualization platform port forward**: VMware / VirtualBox / Hyper-V 의 NAT port forward 설정

→ 본 가이드의 scope 외. 사내 인프라 운영팀이 별도 매뉴얼로 관리.

### 2.4 다른 host:port 재배치

`PUBLIC_ACCESS_HOST` + `PUBLIC_ACCESS_PORT` env 만 변경하면 다른 host:port 로 재배치 가능. realm.dev.json 의 13000 entry 는 사내 dev/smoke 외 환경에서는 무해 (해당 origin 으로 접속 자체가 안 됨).

## 3. 제약 3 — db + Keycloak 내부/외부 분기

### 3.1 제약

DB (PostgreSQL) 와 Keycloak 은 두 가지 모드 운영 가능:

- **내부 모드** (테스트용): compose 가 자체적으로 PostgreSQL + Keycloak 컨테이너 가동
- **외부 모드** (운영용): 사내 별도 운영의 PostgreSQL + Keycloak instance 연결

### 3.2 분기 매트릭스

| 자산 | 내부 모드 | 외부 모드 |
| --- | --- | --- |
| `DB_MODE` env | `docker` | `external` (default) |
| `COMPOSE_PROFILES` env | `local-db,local-idp` | (빈 값 또는 unset) |
| `docker-compose.deploy.yml` db service | 가동 (postgres:15) | 미가동 (profiles: ["local-db"]) |
| `docker-compose.deploy.yml` keycloak service | 가동 (quay.io/keycloak/keycloak:26.0, [ADR-0023](../adr/0023-keycloak-version-pin-26-0.md)) | 미가동 (profiles: ["local-idp"]) |
| `DB_URL` env | 자동 생성 (`postgres://user:pass@db:5432/...`) | **운영자가 명시 설정** (외부 DB host) |
| `KEYCLOAK_UPSTREAM` env (nginx forward 대상) | `keycloak:8080` (docker service name) | `kc.internal.example.com:8443` (외부 host:port) |
| `KEYCLOAK_REALM_IMPORT_PATH` env | emit (dev.json fallback, PR #304 emit gate) | **미 emit** (외부 Keycloak 은 사내 운영팀이 realm 자체 발급) |
| `DEVHUB_OIDC_JWKS_URL` env | `http://nginx/...` (compose internal DNS, [PR #312 Linux 호환](https://github.com/ykylee/Devhub_example/pull/312)) | `https://kc.internal.example.com/...` (외부 Keycloak) |

### 3.3 자산 위치

| 위치 | 적용 |
| --- | --- |
| `scripts/deploy-from-env.sh:206-220` | `DB_MODE` 분기 (`docker` → COMPOSE_PROFILES 설정 + DB_URL 자동 생성 + `KEYCLOAK_HOSTNAME` 자동 설정 / `external` → DB_URL 필수 검증) |
| `scripts/deploy-from-env.sh:223-238` | `COMPOSE_PROFILES` 안 `local-idp` 포함 여부로 OIDC_ISSUER_URL / KEYCLOAK_ADMIN_URL / JWKS URL 분기 (내부 = container DNS / 외부 = public hostname) |
| `scripts/deploy-from-env.sh:268-274` | `KEYCLOAK_REALM_IMPORT_PATH` emit gate (PR #304) — local-idp profile 만 emit (external mode 운영자 오독 회피) |
| `infra/idp/README.md` §2 | 두 가지 모드 SOP — 로컬 모드 / 외부 모드 |
| [keycloak_operations.md](./keycloak_operations.md) | Keycloak 운영 SOP (realm / client / role / event listener / group setup) — 모드 무관 |
| `docs/setup/deploy.env.example` | 내부/외부 분기 example 2 case 분리 (PR `claude/work_260526-internal-network-constraints-docs`) |
| `docs/setup/deploy.stage.env.example` | stage 환경 — 외부 모드 가정 |
| `docs/setup/deploy.prod.env.example` | prod 환경 — 외부 모드 가정 (사내 db + Keycloak 운영팀 instance) |

## 4. 통합 환경 매트릭스 (모드 조합)

| 시나리오 | DB_MODE | COMPOSE_PROFILES | KEYCLOAK_UPSTREAM | env example |
| --- | --- | --- | --- | --- |
| **개발자 로컬 (전체 docker)** | `docker` | `local-db,local-idp` | `keycloak:8080` | `deploy.env.example` § 내부 모드 |
| **개발자 로컬 (외부 db + 내부 Keycloak)** | `external` | `local-idp` | `keycloak:8080` | (custom, 운영자 직접 작성) |
| **개발자 로컬 (내부 db + 외부 Keycloak)** | `docker` | `local-db` | `kc.internal.example.com:8443` | (drift case — 일반적이지 않음) |
| **staging** | `external` | (unset) | `kc-stage.internal.example.com:8443` | `deploy.stage.env.example` |
| **prod** | `external` | (unset) | `kc-prod.internal.example.com:8443` | `deploy.prod.env.example` |

## 5. 자주 만나는 fail 시나리오

[docker-packaging-deployment-guide.md §13 troubleshooting matrix](./docker-packaging-deployment-guide.md#13-build--deploy-troubleshooting-matrix) 의 10 시나리오 + 본 가이드의 분기 별 추가 case:

| 시나리오 | 원인 | 해결 |
| --- | --- | --- |
| 외부 mode 인데 `db` 컨테이너 가동 | `COMPOSE_PROFILES` 에 `local-db` 잔존 | unset 또는 명시 제거 |
| 내부 mode 인데 `DB_URL` external 형식 | `DB_MODE=external` 잔존 | `DB_MODE=docker` 변경 후 `DB_URL` 자동 생성 의존 |
| 13000 으로 접근 안 됨 | 사내 인프라 측 port forward 미설정 | §2.3 의 사내 인프라 설정 점검 (SSH tunnel / NAT / VM platform) |
| JWKS URL resolve fail | 내부 모드 + Linux host (host.docker.internal 미지원) | 본 PR 의 PR #312 변경으로 `http://nginx/...` 자동 사용 (compose internal DNS). docker-compose update 필요. |
| Keycloak admin token 401 | 26.x vs 25.x admin bootstrap env mismatch | docker-compose.deploy.yml:117-121 의 양쪽 env 주입 패턴 (`KC_BOOTSTRAP_ADMIN_*` + `KEYCLOAK_ADMIN`). `feedback_keycloak_25_26_admin_env` 정합. |
| realm INVALID_REDIRECT_URI | 외부 mode 에서 realm 의 redirectUris 와 deploy 환경의 `OIDC_REDIRECT_URI` 불일치 | 외부 Keycloak 의 realm 운영팀에 redirect URI 갱신 요청, 또는 `sync_keycloak_redirects()` 의 admin REST 호출 권한 확인 |

## 6. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-26 | 1차 발행 — 3 제약사항 통합 (host build / port forward / db+Keycloak 분기) + 분기 매트릭스 + cross-link 5+ 위치. 사용자 사내 네트워크 제약 전달 후 정리. | `claude/work_260526-internal-network-constraints-docs` |
