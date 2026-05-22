# 단일 포트 배포 운영 가이드 (ADR-0018, Issue #238)

- 문서 목적: DevHub 를 외부에서 단일 포트 (80/443) 로 노출하는 운영 배포 시나리오 + TLS 인증서 발급 옵션 + dev local 모드 분기 + 운영 진입 절차를 정리한다.
- 범위: nginx reverse proxy + `/devhub` URL context + TLS 인증서 + healthcheck + dev/prod 운영 분기.
- 대상 독자: 사내 운영팀, dev 운영자, infra reviewer.
- 상태: draft (issue #238 머지 동반)
- 최종 수정일: 2026-05-20 (sprint claude/work_260520-o-238-augment)
- 관련 문서: [ADR-0018 단일 포트 reverse proxy](../adr/0018-single-port-reverse-proxy.md), [ADR-0019 Keycloak 단일화](../adr/0019-keycloak-only-idp.md), [keycloak_operations.md](./keycloak_operations.md), [docker-packaging-deployment-guide.md](./docker-packaging-deployment-guide.md), [test-server-deployment.md](./test-server-deployment.md), [e2e-test-guide.md](./e2e-test-guide.md)

## 1. 목적

issue #238 의 요구 사항을 정의한다:

- **외부 접근 = 단일 포트** (80 HTTP redirect → 443 HTTPS)
- **URL context = `/devhub` basePath** — frontend (Next.js basePath) + backend API (`/devhub/api/*` → backend `/api/*`) + Keycloak (`/devhub/auth/keycloak/*`) 모두 동일 prefix
- **Redirect 정합** — OIDC login/callback/logout 흐름이 prefix 중복 없이 정상 종료
- **컨테이너 외부 노출** — nginx 만. backend/frontend/Keycloak 은 internal network 전용

## 2. 사전 조건

| 항목 | 요건 |
| --- | --- |
| Docker | 20.10+ (compose v2 + healthcheck 지원) |
| Docker Compose | v2.20+ (depends_on 의 `condition: service_healthy` 지원) |
| 외부 도메인 | (운영) `devhub.example.com` (DNS A record + 443 inbound 허용) |
| TLS 인증서 | `.crt` + `.key` PEM 파일 (자세한 발급 옵션은 §3) |
| PostgreSQL 15 | 사내 운영 DB (`KC_DB_URL` + `DB_URL`) |
| Keycloak | (운영) 별도 사내 instance + DevHub realm + clients 사전 setup. (dev) `local-idp` profile 로 docker-compose 안에서 |

## 3. TLS 인증서 발급 옵션

본 운영 가이드는 nginx 가 `infra/nginx/certs/tls.crt` + `infra/nginx/certs/tls.key` 두 PEM 파일을 mount 한다고 가정한다. 실제 발급은 운영 환경 정책 따른다.

### 3.1 옵션 A — Let's Encrypt (운영 권장)

운영 도메인이 인터넷에 노출돼 있는 경우. 자동 갱신 + 무료.

```bash
# 1. certbot 설치
sudo apt install certbot

# 2. webroot 또는 standalone 모드로 발급
sudo certbot certonly --standalone -d devhub.example.com

# 3. nginx 가 mount 할 위치로 복사 (또는 symlink)
sudo cp /etc/letsencrypt/live/devhub.example.com/fullchain.pem ./infra/nginx/certs/tls.crt
sudo cp /etc/letsencrypt/live/devhub.example.com/privkey.pem ./infra/nginx/certs/tls.key
sudo chown $(id -u):$(id -g) ./infra/nginx/certs/tls.{crt,key}
sudo chmod 644 ./infra/nginx/certs/tls.crt
sudo chmod 600 ./infra/nginx/certs/tls.key
```

자동 갱신은 `certbot renew --post-hook "docker compose -f docker-compose.deploy.yml exec nginx nginx -s reload"` cron 또는 systemd timer 로.

### 3.2 옵션 B — mkcert (사내망 / 로컬 dev CA)

사내 polished CA + 사내 도메인. 브라우저 신뢰 인증서가 필요한 dev 환경.

```bash
# 1. mkcert 설치 (Linux/macOS)
brew install mkcert    # macOS
# 또는 https://github.com/FiloSottile/mkcert/releases/ 에서 binary

# 2. 로컬 CA 설치 (브라우저 trust store 등록)
mkcert -install

# 3. 인증서 발급
mkcert -cert-file ./infra/nginx/certs/tls.crt -key-file ./infra/nginx/certs/tls.key devhub.local localhost 127.0.0.1
```

mkcert 의 CA 인증서 (`~/.local/share/mkcert/rootCA.pem`) 를 다른 개발자 머신에 import 하면 같은 사내 도메인 공유 가능.

### 3.3 옵션 C — Self-signed (단순 dev only)

CA 없이 빠른 테스트. 브라우저 경고 발생 — production 금지.

```bash
# 1. 키 + cert 생성 (10년 유효, SAN 포함)
openssl req -x509 -nodes -days 3650 -newkey rsa:4096 \
  -keyout ./infra/nginx/certs/tls.key \
  -out ./infra/nginx/certs/tls.crt \
  -subj "/CN=devhub.local" \
  -addext "subjectAltName=DNS:devhub.local,DNS:localhost,IP:127.0.0.1"
```

브라우저 / curl 사용 시 verify 우회 필요 (`curl -k`). e2e Playwright 는 `ignoreHTTPSErrors: true` (codex PR #245 가 적용).

### 3.4 인증서 파일 위치 + 권한

```
infra/nginx/certs/
  tls.crt       (chmod 644, group/world readable)
  tls.key       (chmod 600, owner only — Docker daemon UID 가 읽을 수 있어야)
```

`.gitignore` 의 DEV ENVIRONMENT 섹션에 `infra/nginx/certs/*.{crt,key}` 가 포함되어 있어 git 추적 외. 사내 secret manager (Vault / AWS Secrets Manager) 에서 prod 배포 시 inject 권장.

## 4. 환경 변수

```bash
# === 단일 포트 외부 노출 ===
NGINX_HTTP_PORT=80           # 80 → 443 redirect
NGINX_HTTPS_PORT=443
NGINX_TLS_CERT_PATH=./infra/nginx/certs/tls.crt
NGINX_TLS_KEY_PATH=./infra/nginx/certs/tls.key

# === 외부 도메인 ===
KEYCLOAK_HOSTNAME=devhub.example.com    # Keycloak --hostname (issuer URL 도메인)

# === Keycloak ===
KC_BOOTSTRAP_ADMIN_USERNAME=admin
KC_BOOTSTRAP_ADMIN_PASSWORD=<강력한 secret>
KC_DB_URL=jdbc:postgresql://db:5432/devhub
KC_DB_USERNAME=user
KC_DB_PASSWORD=<DB password>
KC_DB_SCHEMA=keycloak

# === Backend ===
DB_URL=postgres://user:pass@db:5432/devhub?sslmode=disable
DEVHUB_OIDC_ISSUER_URL=https://devhub.example.com/devhub/auth/keycloak/realms/devhub
DEVHUB_OIDC_CLIENT_SECRET=<생성된 client secret>
DEVHUB_KEYCLOAK_ADMIN_URL=http://keycloak:8080
DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET=<생성된 service account client secret>
DEVHUB_TRUSTED_PROXIES=172.16.0.0/12    # docker bridge CIDR

# === Frontend ===
NEXT_PUBLIC_BASE_PATH=devhub
OIDC_ISSUER_URL=https://devhub.example.com/devhub/auth/keycloak/realms/devhub
OIDC_REDIRECT_URI=https://devhub.example.com/devhub/auth/callback
NEXT_PUBLIC_OIDC_ISSUER_URL=https://devhub.example.com/devhub/auth/keycloak/realms/devhub
NEXT_PUBLIC_OIDC_REDIRECT_URI=https://devhub.example.com/devhub/auth/callback

# === Images ===
IMAGE_TAG=<git sha or release tag>
IMAGE_REPO_PREFIX=ghcr.io/ykylee/devhub_example
```

## 5. 운영 부트 절차

### 5.1 cold-start (영구 운영 환경)

```bash
# 1. 인증서 발급 (§3 옵션 선택)
# 2. .env 파일 작성 (§4 위 변수 모두)
# 3. db init (PostgreSQL schema 생성)
docker compose -f docker-compose.deploy.yml --profile local-db up -d db db-init

# 4. Keycloak start (1차 realm import)
docker compose -f docker-compose.deploy.yml --profile local-idp up -d keycloak

# 5. backend + frontend + nginx 시작 (healthcheck 따라 순차 진행)
docker compose -f docker-compose.deploy.yml up -d backend-ai backend-core frontend nginx
```

### 5.2 healthcheck 진행 순서

docker-compose 의 `depends_on: condition: service_healthy` 체인:

```
db (15s × 5)
  ↓
db-init (1회, profile local-db)
  ↓
keycloak (90s start_period + 15s × 12, DB migration + realm import)
  ↓
backend-ai (30s start_period + 15s × 6)
  ↓
backend-core (45s start_period + 15s × 6, DB migration + idp-apply-schemas + admin token fetch)
  ↓
frontend (30s start_period + 15s × 6)
  ↓
nginx (15s × 10, depends on frontend + backend-core healthy)
```

전체 cold-start = 약 2~3분. `docker compose ps` 로 STATUS=healthy 모두 확인 후 외부 접속.

### 5.3 운영 검증 (smoke test)

```bash
# 1. nginx 자체 health
curl -k https://devhub.example.com/nginx-health
# expected: ok

# 2. backend health (nginx 우회 안 됨 — internal network)
docker compose exec backend-core wget -q -O- http://localhost:8080/health
# expected: {"status":"ok"}

# 3. frontend → backend rewrite
curl -k https://devhub.example.com/devhub/api/v1/health
# expected: 200

# 4. OIDC discovery (Keycloak)
curl -k https://devhub.example.com/devhub/auth/keycloak/realms/devhub/.well-known/openid-configuration
# expected: 200 + JSON issuer + authorization_endpoint 등

# 5. 브라우저 진입
# https://devhub.example.com → 302 /devhub → 302 /devhub/developer
# → AuthGuard /devhub/login → Keycloak login → /devhub/auth/callback → /devhub/developer
```

## 6. Dev local mode (no TLS, no nginx, native dev)

native dev 모드는 nginx + TLS 우회 + frontend basePath off 권장:

```bash
# .env (local dev)
NEXT_PUBLIC_BASE_PATH=      # empty — basePath 비활성
NEXT_PUBLIC_OIDC_ISSUER_URL=http://localhost:8180/realms/devhub
NEXT_PUBLIC_OIDC_REDIRECT_URI=http://localhost:3000/auth/callback

# backend
unset DEVHUB_TRUSTED_PROXIES   # nginx 없으므로 X-Forwarded-* 신뢰 영역 없음
make setup
make dev    # backend-core :8080 + backend-ai :8000 + frontend :3000 native

# Keycloak (local-idp profile only)
docker compose -f docker-compose.deploy.yml --profile local-idp up -d keycloak
# 또는 사내 dev Keycloak 사용
```

native dev 모드는:
- 외부 80/443 노출 없음 — `localhost:3000` 직접 접속
- nginx 의 `/devhub/*` prefix 강제 없음 — frontend 가 root path 운영
- TLS 없음 — HTTP only

운영 모드 (§5) 와 dev 모드 (§6) 의 frontend bundle 이 다르다 (basePath 인 vs 아웃). e2e CI 는 운영 모드 정합 (basePath 활성 + nginx + TLS — codex PR #245 의 `playwright.config.ts ignoreHTTPSErrors` 동반).

## 7. 운영 troubleshooting

### 7.1 nginx 가 backend 503 응답

원인: backend-core healthcheck 실패 (start_period 45s 안에 ready 안 됨) — DB migration 대기 또는 Keycloak admin token fetch 실패.

```bash
docker compose logs backend-core --tail 100
docker compose ps    # backend-core STATUS=unhealthy?
```

해결: `docker compose restart backend-core` 또는 Keycloak 준비 확인 후 재시작.

### 7.2 Keycloak healthcheck timeout

원인: realm import 가 90초 안에 끝나지 않음 (사내 realm-export.json 이 큰 경우).

해결: `docker-compose.deploy.yml` 의 keycloak healthcheck `start_period` 를 180s 로 늘림 + retries 24 로.

### 7.3 OIDC callback redirect loop

원인: `NEXT_PUBLIC_OIDC_REDIRECT_URI` 와 Keycloak realm 의 `redirectUris` 불일치, 또는 frontend basePath 와 OIDC redirect path 불일치.

확인:
- Keycloak admin console → Clients → `devhub-frontend` → Valid Redirect URIs 에 `https://devhub.example.com/devhub/auth/callback` 포함?
- frontend env `NEXT_PUBLIC_OIDC_REDIRECT_URI` 도 동일?
- nginx 가 `/devhub/auth/callback` 을 frontend 로 proxy?

### 7.4 인증서 만료

증상: 브라우저 `ERR_CERT_DATE_INVALID`.

확인:
```bash
openssl x509 -in infra/nginx/certs/tls.crt -noout -dates
```

해결: `certbot renew` (Let's Encrypt) 또는 mkcert 재발급 + `docker compose exec nginx nginx -s reload`.

## 8. 보안 강화 (운영 권장)

본 가이드는 운영 진입 기준. 운영 후 보안 강화 carve:

- **Keycloak `KC_HOSTNAME_STRICT=true`** + `KC_HOSTNAME_STRICT_HTTPS=true` (issuer URL hijack 방지)
- **Keycloak realm wildcard 좁히기** — `redirectUris` / `webOrigins` / `post.logout.redirect.uris` 의 `*` wildcard → specific 도메인 (`keycloak_operations.md` 참조)
- **CSP / HSTS / X-Frame-Options** — nginx 추가 보안 헤더 carve
- **rate limit** — nginx `limit_req_zone` 으로 brute force 방어
- **secrets rotation SOP** — `DEVHUB_OIDC_CLIENT_SECRET` + `DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET` 분기별 회전

## 9. 변경 이력

| 일자 | sprint | 변경 |
| --- | --- | --- |
| 2026-05-20 | claude/work_260520-o-238-augment (PR #245 보강) | 본 가이드 신규 — TLS cert 발급 옵션 3종 + 운영 부트 절차 + healthcheck 순서 + dev local mode 분기 + 운영 troubleshooting. codex `#238` 작업 (단일 포트 nginx + `/devhub` basePath) 위에 healthcheck / TLS / dev mode 누락 보강. |
