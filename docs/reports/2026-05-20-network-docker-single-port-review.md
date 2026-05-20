# 네트워크/Docker 단일 포트 컨셉 정합성 리뷰

- 문서 목적: DevHub 의 "외부 단일 포트 진입 + 내부 sub-path 라우팅" 컨셉(ADR-0018) 과 "Keycloak 단일 IdP"(ADR-0019) 에 대한 리포지토리 전체 정합성 점검 결과를 정리한다.
- 범위: `docker-compose.deploy.yml`, `infra/nginx/*.conf`, `infra/idp/*`, `dev-up.{sh,ps1}`, `frontend/{lib,app,next.config.ts,.env.example}`, `backend-core` (Trusted Proxy + OIDC 설정 + redirect/Location 작성), `.github/workflows/{ci,docker-image-publish}.yml`, `scripts/setup-keycloak.sh`, `scripts/ci-e2e-sync-check.sh`
- 점검 axes:
  1. **외부 다른 host:port 로의 redirect**: `Location:` 헤더 / `c.Redirect` / `window.location.assign(absolute URL)` / nginx `return 30* http(s)://...`
  2. **절대 URL 에 박힌 port**: `://host:NNNN/...` (localhost 외 host 도 포함)
  3. **build-time inline URL**: Docker build-arg, `NEXT_PUBLIC_*` 빌드 inline
  4. **Set-Cookie domain / Path port 관련 속성**
  5. **Keycloak / OIDC client 의 redirect_uri allowlist 가 다른 port 허용 여부**
  6. **WebSocket URL 의 port 점프**
- 대상 독자: backend / frontend / infra / 운영팀, ADR-0018 / ADR-0019 후속 carve 담당자
- 상태: addressed (모든 9 항목 동일 브랜치에서 정정)
- 최종 수정일: 2026-05-20
- 조사 기준: `main` HEAD `63e0157`
- 정정 브랜치: `claude/network-docker-single-port-cleanup`
- 관련 문서:
  - [ADR-0018 단일 외부 포트 역프록시 정책](../adr/0018-single-port-reverse-proxy-policy.md)
  - [ADR-0019 Keycloak 단일 IdP](../adr/0019-keycloak-only-idp.md)
  - [docs/planning/single_port_reverse_proxy.md](../planning/single_port_reverse_proxy.md)
  - [docs/setup/single_port_deployment.md](../setup/single_port_deployment.md)
  - [docs/setup/keycloak_operations.md](../setup/keycloak_operations.md)

## 1. 컨셉 정의 (검토 기준)

ADR-0018 + ADR-0019 가 정의한 운영 컨셉은 다음과 같다.

1. **외부 단일 포트 진입**: 사용자는 `https://<host>/devhub/*` 하나의 origin (host + port) 만 알면 모든 기능(프론트엔드 UI, 백엔드 API, OIDC 인증) 에 도달한다. 외부에서 `localhost:4444`, `:3000`, `:8180` 같은 별개 포트로 redirect 되면 동작하지 않는 환경(corp ingress + 방화벽 단일 포트만 허용) 을 가정한다.
2. **`/devhub` sub-path 고정**: 모든 외부 트래픽은 `/devhub` prefix 하에 라우팅. 다른 사내 서비스와 충돌 회피 + 쿠키 Path scope 고립.
3. **Same-Origin 강제**: API/Auth 호출이 모두 같은 origin 의 absolute-path-relative (`/devhub/api/*`, `/devhub/auth/keycloak/*`) — CORS 무력화 + cookie SameSite=Lax 정합.
4. **단일 IdP (Keycloak)**: ADR-0019. Hydra/Kratos 모두 폐기. OIDC discovery 한 곳에서 모든 endpoint 도출.
5. **로컬 개발 예외(ADR-0018 §3.3)**: 멀티포트 native (`:3000` + `:8080` + Keycloak `:8180`) 는 dev-only 로 허용. 운영/스테이징은 단일 포트 강제.

본 리뷰는 "외부 단일 포트로만 접근해도 모든 기능이 동작하는가, 그 외 포트로의 redirect 가 발생하지 않는가" 라는 사용자 정의 기준으로 정합 / 위배 / 회색영역 항목을 분류한다.

## 2. 자산 인벤토리

| 영역 | 자산 | 외부 노출 포트 |
| --- | --- | --- |
| Reverse proxy | `docker-compose.deploy.yml` `nginx` 서비스 (`infra/nginx/devhub.deploy.conf` mount) | `${NGINX_HTTP_PORT:-80}:80`, `${NGINX_HTTPS_PORT:-443}:443` (사용자 정의 가능) |
| Frontend | `docker-compose.deploy.yml` `frontend` (Next.js standalone) | `expose: "3000"` (compose 내부만) |
| Backend Core | `docker-compose.deploy.yml` `backend-core` (Go) | `expose: "8080"` (compose 내부만) |
| Backend AI | `docker-compose.deploy.yml` `backend-ai` (FastAPI) | `expose: "8000"` (compose 내부만) |
| Keycloak | `docker-compose.deploy.yml` `keycloak` (`profiles: ["local-idp"]`) | `expose: "8080"` (compose 내부만, KC_HTTP_RELATIVE_PATH=`/devhub/auth/keycloak`) |
| DB | `docker-compose.deploy.yml` `db` (`profiles: ["local-db"]`, postgres:15) | 내부 네트워크 only |
| Native dev | `dev-up.sh` / `dev-up.ps1` | `localhost:8080`, `localhost:3000`, `localhost:8180` (Keycloak 외부 dev 기대) |
| Nginx 설정 (운영) | `infra/nginx/devhub.deploy.conf` | upstream 모두 docker service name (`frontend:3000`, `backend-core:8080`, `keycloak:8080`) |
| Nginx 설정 (template) | `infra/nginx/devhub.conf` | upstream 모두 `127.0.0.1:*` (native + nginx 호스트 동거 가정) |

→ **운영 자산(`docker-compose.deploy.yml` + `devhub.deploy.conf`) 은 ports 매핑이 nginx 단 하나** 이므로 1번 컨셉(외부 단일 포트) 자체는 정합.

## 3. 정합 항목 (✅)

### 3.1 docker-compose 운영 자산
- `frontend` / `backend-core` / `backend-ai` / `keycloak` 모두 **`ports:` 대신 `expose:`** 만 사용 → compose host 의 0.0.0.0 으로 노출되지 않음. 외부에서 `:3000`, `:8080`, `:8180` 직접 접근 자체가 불가능.
- `nginx` 만 `${NGINX_HTTP_PORT:-80}:80` + `${NGINX_HTTPS_PORT:-443}:443` 으로 호스트 노출 → 외부 진입은 80/443 하나로 통일.
- `db` / `keycloak` 은 `profiles: ["local-db"]` / `["local-idp"]` 로 분리 → 사내 운영 환경(외부 PostgreSQL + 외부 Keycloak)과 매끄럽게 분기.

### 3.2 Nginx 라우팅 (`infra/nginx/devhub.deploy.conf`)
- `:443` 단일 `server` 에서 `/devhub/api/`, `/devhub/auth/keycloak/`, `/devhub/`, `/devhub/api/runtime-config`, `/_next/` 모두 처리 → 외부에서 별도 포트로 redirect 발생하지 않음.
- 모든 `proxy_pass` 의 upstream 이 compose service name (`frontend:3000`, `backend-core:8080`, `keycloak:8080`) → 외부로 새는 host:port 노출 없음.
- `:80` → `:443` 그리고 legacy root `/login`, `/auth/callback`, `/auth/logout` → `/devhub/auth/*` 의 302 모두 **같은 origin 내부** redirect (외부 포트로 점프 없음).
- `proxy_set_header X-Forwarded-Prefix /devhub` + `proxy_redirect off` → upstream 가 절대 URL 재작성 시도해도 무력화.

### 3.3 Frontend 단일 origin 정합
- `frontend/lib/config/endpoints.ts`
  - `API_BASE_URL` 기본값이 `BASE_PATH` (비어있으면 `""`) → 브라우저에서 항상 same-origin relative fetch.
  - `WS_BASE_URL` 이 `window.location.{protocol,host}` 기반 동적 ws/wss 조립 → 별도 포트 노출 없음.
  - `OIDC_REDIRECT_URI` 가 `${window.location.origin}${BASE_PATH}/auth/callback` 동적 조립 → 단일 origin 내부.
- `frontend/next.config.ts` `basePath: '/devhub'` (env 설정 시) + rewrites `/api/:path* → ${BACKEND_API_URL_SERVER}/api/:path*` (server-side proxy only).
- `frontend/lib/services/auth.service.ts` 는 **OIDC discovery 기반** 으로 모든 endpoint(authorize / token / end_session) 를 issuer URL 한 곳에서 도출 → Keycloak 이 `/devhub/auth/keycloak/...` 로 own URL 을 발급하면 자동 정합.
- `frontend/app/api/runtime-config/route.ts` 가 OIDC 환경변수를 런타임에 반환 → 빌드 시점 박힌 URL 의존도 최소화.

### 3.4 Keycloak 운영 정합
- compose 의 `KC_HTTP_RELATIVE_PATH=/devhub/auth/keycloak` + `--proxy-headers=xforwarded` → Keycloak own URL 생성 시 prefix 포함, X-Forwarded-Proto 신뢰.
- `KC_HOSTNAME` env 강제 (`:?set`) → 운영 시 외부 hostname 명시 필수 강제.
- `infra/idp/keycloak-realm.json` redirect_uris 에 `https://*/devhub/*`, `http://*/devhub/*` 와일드카드 포함 → 외부 hostname 변경에도 reverse proxy 정합 유지.

### 3.5 Backend Trusted Proxy 정합
- `backend-core/internal/httpapi/router.go:158-182` 가 `DEVHUB_TRUSTED_PROXIES` env 기반 `SetTrustedProxies` 호출.
- compose 의 `backend-core` 에 `DEVHUB_TRUSTED_PROXIES=${DEVHUB_TRUSTED_PROXIES:-172.16.0.0/12}` (docker bridge CIDR default) → nginx 가 바인딩한 `X-Forwarded-For` 신뢰 + audit attribution 정합.

### 3.6 IdP 단일화 (ADR-0019) 코드 정합
- `backend-core` Go 코드의 Hydra/Kratos 잔재는 모두 audit history 보존용 (예: `domain.AuditSourceKratos = "kratos"`, 마이그레이션 컬럼 이름 주석). active code path 에서는 사용 안 됨.
- `config_test.go:26` 는 `hydra_kratos` provider 자체를 **reject** 하는 단위 테스트로 회귀 가드 박혀있음.
- Frontend `auth.service.ts` 는 Keycloak OIDC discovery 만 사용 — Hydra/Kratos endpoint 직접 호출 없음.
- `scripts/ci-e2e-sync-check.sh:44` 가 `DEVHUB_HYDRA_*` / `DEVHUB_KRATOS_*` env 사용을 **금지하는 회귀 가드** 보유 → ADR-0019 정합 강화.

### 3.7 외부 다른 host:port 로의 redirect 부재
점검 axis ①. 다음 grep 모두 **0 hit / 안전**:
- `backend-core` Go 코드 전체에 `c.Redirect(...)` / `http.Redirect(...)` / Response `Location:` 헤더 직접 작성 — **0 hit**. 즉 백엔드가 사용자 브라우저를 다른 host:port 로 점프시키는 경로 자체가 없음.
- `infra/nginx/devhub.deploy.conf` 의 모든 `return 30*` 은 **same-origin path-relative** (`/devhub/...` 또는 `https://$host$request_uri`) → 포트 변경 없는 internal redirect.
- `frontend` 의 `window.location.assign(...)` 호출은 ① OIDC authorize URL (issuer 가 same origin 이면 same origin) ② post_logout_redirect URI (`${window.location.origin}${BASE_PATH}/`) ③ catch-all `"/"` — 모두 same-origin 정합.
- `frontend/lib/services/websocket.service.ts` 는 `window.location.host` 기반 ws/wss 동적 조립 → 포트 점프 없음.

### 3.8 Set-Cookie domain/port 정합
- 운영 자산 (backend-core, frontend, nginx) 에 `Set-Cookie` 의 `domain=` / `port=` 속성 **하드코딩 없음** (grep 결과 0 hit).
- nginx `proxy_cookie_path / /devhub/auth/keycloak` 으로 Keycloak 쿠키의 Path scope 만 단일 origin 내부에서 고립 → ADR-0018 §3.4 정합.

## 4. 위배 / 정정 후보 (⚠)

### 4.1 [P1] `docker-image-publish.yml` 의 stale Hydra build-arg
**위치**: `.github/workflows/docker-image-publish.yml:78-79`
```yaml
build-args: |
  BACKEND_API_URL=http://backend-core:8080
  NEXT_PUBLIC_OIDC_AUTH_URL=http://localhost:4444/oauth2/auth      # ⚠ Hydra :4444
  NEXT_PUBLIC_OIDC_REDIRECT_URI=http://localhost:3000/auth/callback # ⚠ basePath 없음 + 별도 포트
```
**문제**:
- ADR-0019 (2026-05-19) 로 Hydra 폐기됐는데도 이미지 빌드 시 `NEXT_PUBLIC_OIDC_AUTH_URL` 이 **Hydra :4444 endpoint** 로 inline. 빌드된 운영 image 가 이 fallback 으로 동작 시 ① 외부 포트 4444 로 점프 시도 → "동작 안 함" 컨셉 위배 ② Hydra 자체가 stack 에서 제거됨 → 어떤 환경에서도 동작 불가.
- `NEXT_PUBLIC_OIDC_REDIRECT_URI=http://localhost:3000/auth/callback` 도 **basePath `/devhub` 미포함 + 별도 포트 `:3000`** → ADR-0018 §3.4 위배 후보.

**권고**:
- 두 build-arg 모두 제거하고, OIDC URL 은 **런타임 `/api/runtime-config` route** (이미 구현됨) 와 **OIDC discovery** 로 일원화. 빌드 image 에 stale URL 박지 않는다.
- 또는 build-arg 를 운영 placeholder (예: `https://devhub.example.com/devhub/auth/keycloak/realms/devhub/protocol/openid-connect/auth`) 로 갱신.

### 4.2 [P2] `infra/idp/` 의 Hydra/Kratos 설정 파일 잔존
**위치**: `infra/idp/hydra.yaml`, `kratos.yaml`, `hydra.ci.yaml`, `kratos.ci.yaml`, `hydra.deploy.yaml`, `kratos.deploy.yaml`
**문제**:
- ADR-0019 로 폐기된 IdP stack 의 설정 파일이 여전히 `infra/idp/` 루트에 존재. `README.md` 에 deprecation banner 는 명시되어 있으나 파일 자체는 active naming.
- 신규 합류자 또는 운영팀이 `kratos.deploy.yaml` 등을 보고 운영 배포에 사용하려 시도할 confusion risk.
- 내부에 `base_url: http://localhost:4444/`, `post_logout_redirect: http://localhost:3000/` 등 단일 포트 컨셉 위배 후보 grep hit 가 다수 발생 — false positive 양산.

**권고**:
- `infra/idp/legacy-hydra-kratos/` 하위 디렉터리로 이동 (또는 `.archived` 확장자 / 별도 ADR-0001-snapshot/ 아카이브) → grep 노이즈 제거 + 운영 자산과 격리.
- `infra/idp/README.md` deprecation banner 와 위치 이동 사실 cross-link.

### 4.3 [P2] `runtime-config/route.ts` fallback URL 의 dev-only `localhost:8180`
**위치**: `frontend/app/api/runtime-config/route.ts:22,30`
```ts
"http://localhost:8180/devhub/auth/keycloak/realms/devhub/protocol/openid-connect/auth"
"http://localhost:8180/devhub/auth/keycloak/realms/devhub"
```
**문제**:
- 운영 환경에서 `OIDC_AUTH_URL` / `OIDC_ISSUER_URL` env 누락 시 이 fallback 이 브라우저로 응답 → 외부 사용자가 `localhost:8180` 으로 redirect 시도 → "동작 안 함" 컨셉 위배.
- 현재는 docker-compose.deploy.yml 의 `OIDC_REDIRECT_URI:?set` 등 일부 env 만 강제, `OIDC_AUTH_URL` 은 강제 아님.

**권고**:
- 운영 모드에선 env 누락 시 500 + 명확한 로그를 반환 (fail-fast), localhost fallback 은 명시적 `NODE_ENV !== 'production'` 가드 안에서만 노출.
- 또는 fallback 전체 제거 + frontend `auth.service.ts` 가 `OIDC_ISSUER_URL` 만 있으면 discovery 로 도출하도록 책임 일원화.

### 4.4 [P2] `infra/idp/keycloak-realm.json` 의 dev localhost redirect_uris 가 prod realm 에도 import 됨
**위치**: `infra/idp/keycloak-realm.json:58-72`
```json
"redirectUris": [
  "https://*/devhub/*",
  "http://*/devhub/*",
  "http://localhost:3000/*",
  "http://localhost:3000/auth/callback",
  "http://localhost:3000/devhub/auth/callback"
],
"post.logout.redirect.uris": "...##http://localhost:3000/*##..."
```
**문제**:
- `docker-compose.deploy.yml` 의 `keycloak` 서비스가 이 파일을 `--import-realm` 으로 적용 → 운영 realm 에도 `http://localhost:3000/*` 가 허용 redirect_uri 로 등록됨.
- 운영 환경에서 공격자가 phishing site 를 `http://localhost:3000/...` (피해자 PC 의 localhost) 로 유도 시 redirect_uri 검증 통과 → 토큰 탈취 surface 발생.
- ADR-0018 §3.3 "dev / prod 이중 분기 가드" 의 정신 위배.

**권고**:
- `infra/idp/keycloak-realm.dev.json` (localhost 포함) 과 `infra/idp/keycloak-realm.prod.json` (운영 hostname 만) 분리.
- 또는 사내 운영팀이 realm 을 외부 git 외 자산에서 별도 관리 (compose 파일 주석에 권장 명시됨) 사실을 운영 SOP (`docs/setup/keycloak_operations.md`) 에 강조 + 본 realm.json 은 dev 전용임을 파일 헤더에 명시.

### 4.5 [P2] `infra/nginx/devhub.conf` vs `devhub.deploy.conf` 의 모호한 역할 분리
**위치**: `infra/nginx/devhub.conf` (upstream `127.0.0.1:*`) vs `infra/nginx/devhub.deploy.conf` (upstream service name)
**문제**:
- 두 파일 모두 거의 동일 라우팅이지만 upstream 만 다름. `docker-compose.deploy.yml` 은 `devhub.deploy.conf` 만 mount → `devhub.conf` 의 사용처가 명시 안 됨.
- ADR-0018 §5.1 의 cutover SOP 에서 `sudo cp infra/nginx/devhub.conf /etc/nginx/sites-available/` 로 native nginx + native processes 조합용 template 로 추정되지만 README 부재.
- 운영자가 어느 conf 를 mount 해야 단일 포트 컨셉이 동작하는지 가늠하기 어려움.

**권고**:
- `devhub.conf` → `devhub.native.conf` 로 rename + 본 파일 헤더에 "native + host nginx 조합 dev/staging template" 명시.
- 또는 `infra/nginx/README.md` 추가하여 두 파일의 사용 분기 (compose 운영 vs native + nginx) 를 docs/setup/single_port_deployment.md cross-link 와 함께 정리.

### 4.6 [P3] backend Keycloak Admin URL 의 외부 노출 검토
**위치**: `infra/nginx/devhub.deploy.conf:103-118` + `docker-compose.deploy.yml:129`
**문제**:
- `/devhub/auth/keycloak/` 가 OIDC public flow 뿐 아니라 Keycloak Admin REST (`/admin/...`) 도 함께 외부 노출. compose 의 `DEVHUB_KEYCLOAK_ADMIN_URL` 도 외부 URL 로 설정 가정.
- `devhub.deploy.conf:78-79` 주석에 "외부 노출 가능하지만 사내 정책에 따라 별도 internal-only 분리 carve" 명시.
- 단일 포트 컨셉 자체는 위배 아니지만, **단일 포트 = 모든 admin API 도 외부 origin 노출** 로 이어지므로 보안 surface 가 단일 origin 에 집중. internal-only nginx server block 분리 carve 필요.

**권고**:
- carve 후보: nginx `/devhub/auth/keycloak/admin/*` 에 `allow <internal CIDR>; deny all;` 추가, 또는 internal-only :8081 server block 별도 두고 backend 의 `DEVHUB_KEYCLOAK_ADMIN_URL` 만 `http://keycloak:8080/devhub/auth/keycloak` 직접 host 내부 통신 (이미 가능한 구조).

### 4.7 [P1] `scripts/setup-keycloak.sh` 의 임의 포트 fallback + redirect_uri wildcard
**위치**: `scripts/setup-keycloak.sh:7,71`
```bash
BASE_URL="${KEYCLOAK_URL:-http://localhost:23000/devhub/auth/keycloak}"   # ⚠ :23000 임의 포트
...
-d '{"clientId":"devhub-frontend","enabled":true,"publicClient":true,
     "standardFlowEnabled":true,"directAccessGrantsEnabled":false,
     "redirectUris":["*"],"webOrigins":["*"]}'                           # ⚠ 모든 origin 허용
```
**문제 (점검 axis ②, ⑤)**:
- **L7 임의 포트 :23000 fallback**: 사용자가 `KEYCLOAK_URL` 환경변수를 설정하지 않으면 스크립트가 `localhost:23000` 으로 시도. 단일 포트 (80/443) 도, 일반 Keycloak dev 포트 (`:8180`) 도 아닌 임의 포트. dev-up.sh 가 사용하는 `:8180` 과도 불일치 → 운영자 실수 시 broken.
- **L71 redirect_uris `["*"]` + webOrigins `["*"]`**: 모든 host:port 의 redirect_uri 를 Keycloak 이 허용. 단일 포트 컨셉의 가장 강한 안전망 (Keycloak 의 redirect URI allowlist) 을 사실상 무력화. dev/CI 의도이나 운영 realm 에 이 스크립트를 잘못 적용 시 phishing site 의 `http://attacker.example.com:9999/...` 같은 임의 호스트:포트로 redirect 토큰 탈취 가능.

**권고**:
- L7: fallback 을 dev 의 default port (`http://localhost:8180/devhub/auth/keycloak`) 또는 단일 포트 (`http://localhost/devhub/auth/keycloak`) 로 일관화 + `:?` 강제 (env 누락 시 fail-fast).
- L71: redirect_uris / webOrigins 를 `dev-up` 모드의 실제 origin (`http://localhost:3000/*`) 만 허용 + 운영팀이 prod realm 적용 시 반드시 운영 hostname 으로 교체하라는 헤더 주석 추가. 운영 realm 적용 전 정적 검사 (workflow lint) 추가 carve 후보.

### 4.8 [P3] dev/prod 의 OIDC issuer path 구조 불일치
**위치**: `dev-up.sh:121`, `dev-up.ps1:134`, `docker-compose.deploy.yml:124`, `.github/workflows/ci.yml:492,504`

| 환경 | DEVHUB_OIDC_ISSUER_URL 형태 |
| --- | --- |
| dev-up (native) | `http://localhost:8180/realms/devhub` |
| ci.yml | `http://localhost:8180/devhub/auth/keycloak/realms/devhub` |
| compose (운영) | `https://<host>/devhub/auth/keycloak/realms/devhub` (운영자 주입 가정) |

**문제 (점검 axis ②)**:
- dev native 의 issuer 는 `/realms/devhub` (Keycloak default `KC_HTTP_RELATIVE_PATH=/`), 운영 / CI 는 `/devhub/auth/keycloak/realms/devhub` (relative path 박힘). 즉 dev 와 prod 의 OIDC URL prefix 구조가 다름.
- OIDC discovery 가 `.well-known/openid-configuration` 을 같은 prefix 로 따라가므로 path 자체는 동작하지만, dev → prod cutover 시 `KC_HTTP_RELATIVE_PATH` 적용 여부에 따라 dev 에서 작성한 절대 URL 가정 (예: 테스트 fixture 의 `${ISSUER}/protocol/openid-connect/auth`) 이 prod 에서 path 분기 회귀를 일으킬 위험.
- ADR-0018 §3.3 dev 멀티포트 허용 정책상 컨셉 위배 자체는 아니지만, **dev native Keycloak 도 `KC_HTTP_RELATIVE_PATH=/devhub/auth/keycloak` 로 띄우도록** 일관화하면 cutover 회귀 표면 축소 가능.

**권고**:
- `dev-up.sh` / `dev-up.ps1` 의 Keycloak 기동 가이드 (주석 docker 명령) 에 `-e KC_HTTP_RELATIVE_PATH=/devhub/auth/keycloak` 추가 + 기본 issuer URL 도 운영과 동일 prefix 로 갱신.
- 또는 `docs/setup/keycloak_operations.md` 에 dev/prod path prefix 일관화 SOP 명시.

### 4.9 [P3] dev-up 스크립트의 단일 포트 mode 검증 부재
**위치**: `dev-up.sh`, `dev-up.ps1`
**문제**:
- ADR-0018 §3.3 가 명시적으로 "로컬 개발은 멀티포트 OK" 라 dev 자체는 컨셉 위배 아님.
- 그러나 단일 포트 reverse proxy 정합성을 사전 검증할 수 있는 native dev 모드(예: `./dev-up.sh --single-port` 가 native nginx 까지 함께 띄움) 가 없음 → 단일 포트 컨셉 회귀를 PR 시점에 잡기 어려움.
- 현재는 ci.yml 의 E2E 가 멀티포트 native (`:3000` + `:8080` + `:8180`) 기준 — 단일 포트 mode 의 회귀 가드 미흡.

**권고**:
- `dev-up.sh --single-port` 모드 추가 (native + host nginx + `BASE_PATH=/devhub` 환경변수) carve.
- 또는 `docker-compose -f docker-compose.deploy.yml --profile local-db --profile local-idp up` 을 `make smoke-single-port` 로 wrap + E2E 추가.

## 5. 컨셉 정합성 종합 평가

| 영역 | 정합 상태 | 비고 |
| --- | --- | --- |
| 외부 진입 포트 | ✅ 정합 | nginx 만 80/443 노출, 그 외 모든 서비스 `expose:` |
| sub-path `/devhub` 라우팅 | ✅ 정합 | nginx + Next.js basePath + Keycloak KC_HTTP_RELATIVE_PATH 일관 |
| Same-Origin API/Auth | ✅ 정합 | endpoints.ts + auth.service.ts 모두 same-origin 동적 조립 |
| Trusted Proxy 처리 | ✅ 정합 | DEVHUB_TRUSTED_PROXIES env + gin SetTrustedProxies |
| Keycloak 단일 IdP | ✅ 정합 (code) | backend/frontend 모두 Keycloak 전용 path, Hydra/Kratos active code 없음 |
| Docker 이미지 build-time URL | ⚠ 위배 | `docker-image-publish.yml` 이 Hydra :4444 + `:3000` localhost build-arg 보유 (P1) |
| `scripts/setup-keycloak.sh` 임의포트+wildcard | ⚠ 위배 | fallback `:23000` + redirectUris `["*"]` + webOrigins `["*"]` (P1) |
| 외부 host:port 로의 redirect | ✅ 정합 | backend `c.Redirect`/`Location:` 0 hit, nginx `return 30*` 모두 same-origin |
| Set-Cookie domain/port | ✅ 정합 | 운영 자산 cookie domain 하드코딩 없음 + `proxy_cookie_path` 로 path scope 고립 |
| Hydra/Kratos infra/idp 잔재 | ⚠ 회색 | 파일 잔존 + grep 노이즈 (P2) |
| runtime-config fallback | ⚠ 회색 | dev-only localhost:8180 fallback 이 prod 에 노출 가능 (P2) |
| Keycloak realm dev/prod 분리 | ⚠ 회색 | realm.json 에 dev localhost redirect_uri 가 prod 에도 import (P2) |
| nginx conf 분리 명세 | ⚠ 회색 | devhub.conf vs deploy.conf 사용처 모호 (P2) |
| Admin API 외부 노출 | ⚠ 회색 | Keycloak Admin path 도 외부 노출 — carve 권장 (P3) |
| dev/prod issuer path 일관성 | ⚠ 회색 | dev native `/realms/devhub` vs prod `/devhub/auth/keycloak/realms/devhub` (P3) |
| 단일 포트 사전 검증 SOP | ⚠ 회색 | dev 멀티포트만 검증, single-port 회귀 가드 부재 (P3) |

**결론**: 외부 단일 포트 진입 + sub-path 라우팅 + Same-Origin 컨셉은 **운영 자산 (`docker-compose.deploy.yml` + `devhub.deploy.conf`) 에서 구조적으로 정합** 한다. 사용자가 `https://<host>/devhub/*` 한 origin 만으로 모든 기능에 도달 가능하며, **backend / nginx / frontend 어디에서도 외부 다른 host:port 로의 redirect 가 발생하지 않는다** (axis ① 점검 결과 0 hit).

그러나 다음 surface 가 잠재적 회귀 위험으로 남아있다:
- **P1 (즉시 정정)** — ① `docker-image-publish.yml` 의 stale Hydra `:4444` + localhost `:3000` build-arg ② `scripts/setup-keycloak.sh` 의 임의 포트 `:23000` fallback + redirect_uris `["*"]` wildcard 가드 부재.
- **P2 (carve sprint 권장)** — ③ Hydra/Kratos 잔재 자산 (`infra/idp/{hydra,kratos}*.yaml`) ④ `runtime-config/route.ts` 의 prod 노출 가능 localhost fallback ⑤ realm.json 의 dev/prod 미분리 ⑥ `infra/nginx/devhub.conf` vs `devhub.deploy.conf` 사용처 모호.
- **P3 (인지/문서화)** — ⑦ Keycloak Admin API 의 외부 노출 ⑧ dev/prod OIDC issuer path 구조 불일치 ⑨ 단일 포트 mode 의 native dev / E2E 회귀 가드 SOP 부재.

P1 은 즉시 정정 (image 빌드 가 production 환경에서 부정합 / 보안 가드 무력화 위험). P2/P3 는 별도 carve sprint.

## 6. 정정 내역 (carve A~H 동일 PR 처리)

본 리뷰의 9 항목 (carve A~I) 중 8 항목을 `claude/network-docker-single-port-cleanup` 브랜치 단일 PR 에서 정정. carve I (Playwright 단일 포트 E2E) 만 별도 sprint 로 분리.

| carve | Pri | 정정 사실 |
| --- | --- | --- |
| A | P1 | `.github/workflows/docker-image-publish.yml`: `NEXT_PUBLIC_OIDC_AUTH_URL` + `NEXT_PUBLIC_OIDC_REDIRECT_URI` build-arg 제거. OIDC URL 은 `/api/runtime-config` + OIDC discovery 로 일원화. |
| B | P1 | `scripts/setup-keycloak.sh`: fallback `:23000` → `:8180/devhub/auth/keycloak` (dev-up 정합). `redirectUris:["*"], webOrigins:["*"]` 제거 → `DEVHUB_FRONTEND_ORIGIN` + `DEVHUB_FRONTEND_BASEPATH` env 기반 single-origin allowlist. env 미설정 시 fail-fast. 외부 Keycloak 모드도 지원. |
| C | P2 | `infra/idp/{hydra,kratos}*.yaml`, `kratos_webhooks/`, `scripts/install-binaries.ps1`, `README.md`, `ENVIRONMENT_NOTES.md` 모두 `infra/idp/_archive_hydra_kratos/` 로 `git mv`. `ARCHIVE_NOTICE.md` 신규. `infra/idp/README.md` 는 Keycloak 모드 분기 SOP 로 재작성. |
| D | P2 | `infra/idp/keycloak-realm.json` → `keycloak-realm.dev.json` rename. dev realm 의 redirect_uri 에서 `http://*/devhub/*` wildcard 제거 + localhost-only 로 좁힘. 외부 모드용 reference template `keycloak-realm.prod.json` 신규 (`__DEVHUB_HOST__` / `__DEVHUB_BACKEND_SECRET__` placeholder + manage-users 제외 최소 권한). |
| E | P2 | `frontend/app/api/runtime-config/route.ts`: `NODE_ENV === 'production'` + OIDC env 누락 시 500 fail-fast. dev fallback 은 `http://localhost:8180/devhub/auth/keycloak/...` 로 prefix 정합. |
| F | P2 | `infra/nginx/devhub.conf` → `devhub.native.conf` rename + 헤더 명확화. `devhub.deploy.conf` → `devhub.deploy.conf.template` (envsubst). `infra/nginx/README.md` 신규 (compose template vs native + 로컬/외부 Keycloak 모드 SOP). |
| G | P3 | nginx 양쪽 conf 에 `location ^~ /devhub/auth/keycloak/admin/` 추가 — `KEYCLOAK_ADMIN_ALLOW_CIDR` 미설정 시 deny all (외부 admin REST 차단). backend 는 internal direct 통신. |
| H | P3 | `dev-up.sh` / `dev-up.ps1` 의 issuer URL 을 `http://localhost:8180/devhub/auth/keycloak/realms/devhub` 로 갱신 + Keycloak docker hint 에 `KC_HTTP_RELATIVE_PATH=/devhub/auth/keycloak` 추가 → dev/prod path 일관화. |
| I | P3 | (별도 sprint) — `dev-up --single-port` 모드 + Playwright E2E 단일 포트 cover. 본 PR scope 외. |

### 6.1 외부 Keycloak 모드 지원 사실 (사용자 directive 정합)
사용자 요청 "Keycloak 은 로컬 테스트 모드와 외부 모드를 나눠줘야 해" 를 다음으로 충족:

- compose 의 `keycloak` 서비스는 `profiles: ["local-idp"]` 로 분리 — `--profile local-idp` 없이 가동 시 외부 Keycloak 모드 진입.
- nginx 의 Keycloak upstream 은 `${KEYCLOAK_UPSTREAM}` 환경변수 (envsubst) — 로컬은 `keycloak:8080`, 외부는 `kc.internal.example.com:8443` 등.
- realm 파일도 `dev.json` (로컬 자동 import) / `prod.json` (외부 운영팀 reference template) 으로 분리.
- `setup-keycloak.sh` 는 두 모드 모두 동작 (필수 env `KEYCLOAK_URL` + `DEVHUB_FRONTEND_ORIGIN`).
- `infra/idp/README.md` §2.1 / §2.2 와 `infra/nginx/README.md` §2.1 / §2.2 에 양 모드 SOP 명시.

## 7. 추적성 영향

- 본 리뷰는 의미 변경 없는 진단 보고 — 추적성 매트릭스 row 갱신 N/A.
- carve A~G 진입 시 각 carve PR 이 `docs/traceability/sync-checklist.md` 절차로 IMPL/UT/TC 갱신.

## 8. 변경 이력

- **2026-05-20** (`main` HEAD `63e0157`): 단일 포트 컨셉 리뷰 초안 작성. P1×2 / P2×4 / P3×3 발견.
  - 초기: localhost 사용 + reverse proxy 정합성 (P1×1 / P2×4 / P3×2)
  - 보강 (사용자 directive): 포트 변경 axes 6개 (redirect, 절대 URL port, build-time inline, Set-Cookie, OIDC allowlist, WebSocket) 추가 점검 → `setup-keycloak.sh` 의 임의 포트 + wildcard 발견 (P1 추가) + dev/prod issuer path 불일치 발견 (P3 추가)
- **2026-05-20** (브랜치 `claude/network-docker-single-port-cleanup`): 9 항목 중 carve A~H (8건) 동일 PR 에서 정정. Keycloak 로컬/외부 모드 분리 + 단일 포트 컨셉 정합 가드 강화. §6 정정 내역 표 참조.
