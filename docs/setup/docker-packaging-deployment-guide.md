# Docker 패키징/배포 가이드

> ⚠ 2026-05-20 정정: 본 문서는 Keycloak-only 배포 기준이다. Hydra/Kratos 변수/절차는 사용하지 않는다.

- 문서 목적: DevHub Example에서 Docker 패키징 오류를 줄이기 위한 표준 빌드 절차와 배포 방식(이미지 배포 vs compose 배포) 선택 기준을 정의한다.
- 범위: 이미지 태깅 규칙, 빌드/푸시 절차, compose 사용 범위, 운영 권장안
- 대상 독자: 개발자, 릴리즈 담당자, 운영자
- 상태: draft
- 최종 수정일: 2026-05-26 (§1.1 host 사전조건 신규)
- 관련 문서: [개발 환경 구성 가이드](./environment-setup.md), [테스트 서버 배포 가이드](./test-server-deployment.md), [ADR-0003](../adr/0003-no-docker-policy-ci-scope.md)

## 1. 현재 저장소 기준 Docker 자산

현재 저장소에는 다음 Docker 자산이 존재한다.

- `backend-core/Dockerfile`
- `backend-ai/Dockerfile`
- `frontend/Dockerfile`
- `docker-compose.yml`
- `docker-compose.local.yml`

compose는 서비스 간 연결/개발 실행에 유용하지만, 배포 산출물의 재현성과 추적성은 이미지 중심으로 관리하는 편이 안정적이다.

### 1.1 host 사전조건

`scripts/deploy-from-env.sh` + `scripts/setup-keycloak.sh` + `scripts/build-artifacts.sh` 가 deploy host 에 의존하는 도구는 다음과 같다. **미설치 시 silent fail (script 중간 단계에서 깨짐)** 가능하므로 deploy 진입 전 확인한다.

| 도구 | 용도 | 검증 명령 |
| --- | --- | --- |
| `python3` | `deploy-from-env.sh:generate_local_realm_import()` (heredoc, realm.dev.json → generated realm import) + `setup-keycloak.sh` (admin token 파싱 / role JSON 추출 / mapper 존재 체크 등 다수 위치) | `python3 --version` (3.8+ 권장, json 모듈 표준 라이브러리만 사용) |
| `curl` | `setup-keycloak.sh` 의 Keycloak Admin REST API 호출 + readiness wait | `curl --version` |
| `docker` + `docker compose` (v2 plugin) | runtime image build + `scripts/deploy-up.sh` 의 compose 호출 | `docker --version` + `docker compose version` |
| `bash` 4+ | `scripts/*.sh` 의 array/heredoc/`[[` syntax | `bash --version` |
| `jq` (선택) | runtime-config endpoint 응답 확인 등 docs 예시에서 사용. script 자체는 의존 X. | `jq --version` |

**python3 미설치 host 에서 deploy 진입 시 증상**: `deploy-from-env.sh:2` 의 `set -euo pipefail` + `:202` 의 `KEYCLOAK_REALM_IMPORT_PATH="$(generate_local_realm_import ...)"` 가 `python3: command not found` 로 fail 시 함수 안에서 **script 전체 즉시 종료** (`emit_env_line` 단계까지 도달 못 함). 자동 fallback 아님.

**수동 우회 절차** (python3 설치 불가 host): deploy 진입 전 `KEYCLOAK_REALM_IMPORT_PATH=$ROOT/infra/idp/keycloak-realm.dev.json` env 를 사전 export → `build_env_file()` 의 `KEYCLOAK_REALM_IMPORT_PATH="${KEYCLOAK_REALM_IMPORT_PATH:-...}"` ([line 250](../../scripts/deploy-from-env.sh#L250)) 가 사전 export 값을 그대로 사용. 단 이 경로는 `generate_local_realm_import()` 의 dynamic redirect URI 주입 (PUBLIC_ACCESS_HOST 기준 callback URL append) 효과를 잃으므로 사내 ingress (`localhost:13000` 외 host) 사용 시 redirect URI 가 dev.json 의 hardcoded entries 로 제한된다. 권고는 python3 설치.

## 2. 결론 (권장 운영안)

이 저장소 환경에서는 다음 모델을 권장한다.

- 배포 단위: `versioned image` (권장)
- 실행 오케스트레이션: 환경별 `compose` 또는 플랫폼 설정
- 금지/비권장: 배포 시점마다 대상 서버에서 `compose build`로 새 이미지를 직접 생성

이유:

- 서버마다 빌드 캐시, 네트워크, 빌더 버전이 달라 재현성 문제가 발생한다.
- 빌드 오류가 배포 단계로 전이되어 복구 시간이 늘어난다.
- 동일 태그 이미지로 dev/stage/prod 승격이 가능해 릴리즈 추적이 단순해진다.

## 3. 배포 방식 비교

### 3.1 방식 A: 이미지 만들어서 배포 (권장)

흐름:

1. CI/빌드 서버에서 이미지 빌드
2. 레지스트리에 푸시 (`<service>:<git-sha>` + 선택적 릴리즈 태그)
3. 운영 서버는 `docker compose pull` + `up -d`만 수행

장점:

- 재현성 높음 (같은 digest 재사용)
- 실패 지점이 "빌드"와 "배포"로 분리됨
- 롤백 단순 (`이전 태그` 재적용)

단점:

- 레지스트리 운영 필요
- 태그 정책/정리 정책 필요

### 3.2 방식 B: compose 파일만 배포해서 서버에서 build

흐름:

1. compose + 소스 전달
2. 대상 서버에서 `docker compose build && up -d`

장점:

- 초기에 설정이 단순해 보임

단점:

- 서버 환경차로 빌드 오류가 반복 발생
- 릴리즈 검증 전, 운영 서버에서 첫 빌드가 실행됨
- 동일 버전 보장 어려움 (의존성/캐시 영향)

## 4. 권장 태깅/이미지 규칙

- 태그 기본: `devhub/<service>:<git-sha>`
- 릴리즈 태그(선택): `devhub/<service>:vX.Y.Z`
- latest 단독 운영 금지: `latest`는 보조 태그로만 사용
- 배포 기준: 태그보다 `image digest` 우선

서비스 예시:

- `devhub/backend-core:<git-sha>`
- `devhub/backend-ai:<git-sha>`
- `devhub/frontend:<git-sha>`

## 5. 표준 빌드 명령 예시

```sh
# host build (build artifacts only)
ENV_FILE=./.env.deploy ./scripts/build-artifacts.sh

# runtime image packaging
docker build -f backend-core/Dockerfile -t devhub/backend-core:${GIT_SHA} backend-core
docker build -f backend-ai/Dockerfile -t devhub/backend-ai:${GIT_SHA} backend-ai
docker build -f frontend/Dockerfile -t devhub/frontend:${GIT_SHA} frontend
```

푸시:

```sh
docker push devhub/backend-core:${GIT_SHA}
docker push devhub/backend-ai:${GIT_SHA}
docker push devhub/frontend:${GIT_SHA}
```

## 6. compose 배포 시 운영 규칙

배포용 compose는 build가 아니라 image 참조를 사용한다.

예시 원칙:

- `build:` 대신 `image: devhub/backend-core:${IMAGE_TAG}`
- 서버에서는 `docker compose pull && docker compose up -d`
- `docker compose build`는 로컬 개발/긴급 점검에서만 사용

## 7. 체크리스트

배포 전:

- 이미지 3종 빌드 성공
- 태그/다이제스트 기록
- 앱 설정값(`DB_URL`, `BACKEND_AI_URL`, OIDC 관련 env) 검증

배포 시:

- `pull` 후 재기동
- 헬스체크 확인 (`/health`, frontend 응답)

배포 후:

- 로그 에러율 확인
- 필요 시 이전 태그로 즉시 롤백

## 8. 배포 템플릿/CI 자동화

저장소에 다음 자산을 추가했다.

- `docker-compose.deploy.yml`: `build` 없이 `image`만 참조하는 배포용 compose 템플릿
- `.github/workflows/docker-image-publish.yml`: backend-core/backend-ai/frontend 이미지 빌드+GHCR 푸시
- `docs/setup/deploy.env.example`: 배포용 필수 env 템플릿
- `docs/setup/deploy.stage.env.example`: stage 환경 템플릿
- `docs/setup/deploy.prod.env.example`: production 환경 템플릿
- `scripts/deploy-preflight.sh`: 배포 전 필수 env/compose 렌더/OIDC reachability 검증
- `scripts/deploy-up.sh`: preflight + pull + up 일괄 실행

### 8.1 배포용 compose 실행 예시

```sh
export IMAGE_TAG=<git-sha-or-release-tag>
export IMAGE_REPO_PREFIX=local/devhub             # 원격 배포 시 ghcr.io/<owner>/<repo> 로 덮어쓰기
export PUBLIC_ACCESS_SCHEME=http
export PUBLIC_ACCESS_HOST=<host-ip-or-dns>
export PUBLIC_ACCESS_PORT=13000
export NGINX_HTTP_PORT=3000                       # VM ingress port
export DB_URL='postgres://<user>:<pw>@<db-host>:5432/<db>?sslmode=disable'
export DEVHUB_IDP_PROVIDER=keycloak
export DEVHUB_OIDC_ISSUER_URL=http://<host-ip-or-dns>:13000/devhub/auth/keycloak/realms/devhub
export DEVHUB_OIDC_CLIENT_ID=devhub-web
export DEVHUB_OIDC_CLIENT_SECRET='<oidc-client-secret>'
export DEVHUB_KEYCLOAK_ADMIN_URL=http://keycloak:8080
export DEVHUB_KEYCLOAK_ADMIN_REALM=devhub
export DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID=devhub-admin
export DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET='<keycloak-admin-secret>'
export NEXT_PUBLIC_IDP_PROVIDER=keycloak
export NEXT_PUBLIC_OIDC_ISSUER_URL=http://<host-ip-or-dns>:13000/devhub/auth/keycloak/realms/devhub
export NEXT_PUBLIC_BASE_PATH=devhub
export NEXT_PUBLIC_OIDC_REDIRECT_URI=http://<host-ip-or-dns>:13000/devhub/auth/callback
./scripts/deploy-from-env.sh
```

또는 저장소 표준 스크립트 사용:

```sh
cp docs/setup/deploy.env.example ./.env.deploy
# .env.deploy 값 수정
ENV_FILE=./.env.deploy ./scripts/deploy-up.sh
```

로컬 빌드 이미지를 그대로 사용할 때:

```sh
export IMAGE_TAG=$(git rev-parse --short HEAD)
export IMAGE_REPO_PREFIX=local/devhub
./scripts/build-artifacts.sh
docker build -f backend-core/Dockerfile -t "${IMAGE_REPO_PREFIX}/backend-core:${IMAGE_TAG}" backend-core
docker build -f backend-ai/Dockerfile -t "${IMAGE_REPO_PREFIX}/backend-ai:${IMAGE_TAG}" backend-ai
docker build -f frontend/Dockerfile -t "${IMAGE_REPO_PREFIX}/frontend:${IMAGE_TAG}" frontend
docker compose -f docker-compose.deploy.yml up -d
```

원격 레지스트리 이미지를 쓸 때만 `docker compose pull` 을 실행한다. `IMAGE_REPO_PREFIX=local/*`
계열은 로컬 빌드 산출물을 바로 사용하므로 pull 이 필요 없다.

`IMAGE_TAG`와 Keycloak/OIDC 관련 URL은 필수다. 미지정 시 compose가 오류로 중단되도록 설정되어 있다.
`DB_URL`도 필수다. 미지정 시 compose가 오류로 중단된다.
`DEVHUB_OIDC_CLIENT_SECRET`, `DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET`은 운영 배포 필수값이며, 예시 기본값(`dev-token` 류) 사용은 금지한다.
`DEVHUB_AUTH_DEV_FALLBACK` 기본값은 `0`(비활성)이며, 배포 환경에서 `1`로 켜지지 않도록 유지한다.
`db-migrate` 서비스가 항상 선행되어 신규 DB 는 1회 초기 마이그레이션, 기존 DB 는 증분 마이그레이션만 수행한다.

### 8.1.1 변수 스키마 (권장)

- Public (브라우저가 직접 접근): `PUBLIC_ACCESS_SCHEME`, `PUBLIC_ACCESS_HOST`, `PUBLIC_ACCESS_PORT`, `DEVHUB_PUBLIC_BASE_URL`, `NEXT_PUBLIC_IDP_PROVIDER`, `NEXT_PUBLIC_OIDC_ISSUER_URL`
- Internal (서비스 간 통신): `DEVHUB_OIDC_*`, `DEVHUB_KEYCLOAK_ADMIN_*`, `BACKEND_API_URL`
- DB: `DB_URL`

`localhost`는 fallback일 뿐 표준값이 아니다. 서버를 분리 배치하는 경우에는 위 3축을 환경별로 명시 주입한다.

issuer/JWKS 분리 권장:
- `DEVHUB_OIDC_ISSUER_URL`: 브라우저/토큰 claim 과 일치하는 public issuer
- `DEVHUB_OIDC_JWKS_URL`: backend 가 실제로 접근 가능한 internal JWKS URL (필요 시)

### 8.1.2 DB 모드 선택

- 번들 DB 모드 (`local-db` profile): compose 내부 `db`(postgres:15) 포함 기동
- 번들 Keycloak 모드 (`local-idp` profile): compose 내부 `keycloak` 포함 기동
- 외부 DB/외부 Keycloak 모드(기본): profile 없이 기동하고 `DB_URL`, `DEVHUB_OIDC_ISSUER_URL`, `DEVHUB_KEYCLOAK_ADMIN_URL`, `OIDC_ISSUER_URL`, `OIDC_REDIRECT_URI`를 외부 주소로 지정
- 외부 DB 모드 (default): compose 내부 `db` 미기동, `DB_URL`은 외부 DSN 지정

예시:

```sh
# 1) 번들 DB 모드
export DB_URL='postgres://<user>:<pw>@db:5432/<db>?sslmode=disable'
docker compose -f docker-compose.deploy.yml --profile local-db --profile local-idp up -d

# 2) 외부 DB 모드
docker compose -f docker-compose.deploy.yml up -d
```

`docker-compose.deploy.yml`은 `nginx`를 포함한다. 외부 진입은 `https://<host>/devhub` 기준으로 통일하고, `frontend`/`backend-core`/`keycloak`/`backend-ai`는 host에 직접 노출하지 않는다.
`local-db` 프로필에서는 `db-init` 단계가 keycloak schema를 준비한다.

### 8.1.3 DB 마이그레이션 정책 (자동)

- `db-migrate` 서비스가 `migrate ... up`을 실행한 뒤 `backend-core`가 기동된다.
- 빈 DB(first deploy): 전체 초기 스키마 자동 생성.
- 기존 DB(redeploy): 미적용 버전만 증분 적용, 기존 데이터 보존.
- 실패 시: `backend-core`가 시작되지 않으므로 partial rollout을 조기에 차단.

주의:

- 로컬 재검증 시 realm/secret 변경 후 불일치가 발생하면 아래처럼 볼륨 초기화 후 재기동한다.

```sh
docker compose -f docker-compose.deploy.yml --profile local-db down -v
docker compose -f docker-compose.deploy.yml --profile local-db up -d
```

### 8.2 GitHub Actions 이미지 퍼블리시

- 수동 실행: Actions > `Docker Image Publish` > `Run workflow`
- 태그 실행: `v*` 태그 푸시 시 자동 실행
- 결과 이미지:
  - `ghcr.io/<owner>/<repo>/backend-core:<tag>`
  - `ghcr.io/<owner>/<repo>/backend-ai:<tag>`
  - `ghcr.io/<owner>/<repo>/frontend:<tag>`

## 9. 이 저장소에서의 실무 권장

- 개발 편의: 현재 `docker-compose.local.yml` 유지
- 배포 안정성: 이미지 선빌드 + compose는 image pinning만 수행
- 문서/스크립트 개선 과제:
  - 릴리즈 노트에 태그+digest 자동 기록

## 10. 장애 분리 기준 (docker vs native)

아래 항목은 실제 재현에서 분리된 이슈다.

- docker 전용 이슈 1: `frontend`의 `/api` 프록시 대상이 이미지 빌드 시점 기본값(`http://localhost:8080`)으로 굳어질 수 있다.
  - 증상: 로그인 흐름에서 `Failed to proxy http://localhost:8080/...`
  - 대응: 배포 패키지의 `nginx`에서 `/api/v1/*`를 `backend-core:8080`으로 직접 프록시한다.
- docker 전용 이슈 2: OIDC issuer/redirect/logout URL 을 내부 DNS(`http://backend-core:8080` 등)로 두면 외부 브라우저가 해당 호스트를 해석하지 못한다.
  - 증상: 로그인/로그아웃 후 `DNS_PROBE_FINISHED_NXDOMAIN` 또는 callback 실패
  - 대응: IdP client 설정의 `redirect_uris`/`post_logout_redirect_uris` 와 app env 의 public URL을 외부 접근 가능한 host 로 일치시킨다.
- 공통 설정 이슈: OIDC redirect URI를 `/api/auth/callback`로 주면 라우트 불일치가 발생한다.
  - 기준 라우트: `/auth/callback`
  - 대응: `OIDC_REDIRECT_URI`, `NEXT_PUBLIC_OIDC_REDIRECT_URI`, IdP client `redirect_uris`를 동일하게 `/auth/callback`으로 맞춘다.

최소 검증 순서:

1. `curl -k https://<host>/devhub/api/runtime-config`에서 OIDC URL/redirect 값 확인
2. `curl <OIDC_ISSUER_URL>/.well-known/openid-configuration` 확인
3. Playwright 단건 검증  
   `PLAYWRIGHT_BASE_URL=https://<host>/devhub npm run e2e -- tests/e2e/auth.spec.ts --grep "developer lands on /developer"`

## 11. frontend 패키징 정책 — runtime-config 우선 + 최소 build arg

### 11.1 build-time env 매트릭스

| env | 용도 | 운영 예시 |
| --- | --- | --- |
| `BACKEND_API_URL` | 서버측 rewrite 대상 (`next.config.ts`) | `http://backend-core:8080` |
| `NEXT_PUBLIC_BASE_PATH` | basePath / callback 경로 | `devhub` |
| `NEXT_PUBLIC_APP_ORIGIN` | client fallback origin | `https://devhub.example.com` |
| `NEXT_PUBLIC_OIDC_ISSUER_URL` | client fallback issuer | `https://devhub.example.com/devhub/auth/keycloak/realms/devhub` |
| `NEXT_PUBLIC_OIDC_REDIRECT_URI` | client fallback callback | `https://devhub.example.com/devhub/auth/callback` |

### 11.2 build 명령 예시

```bash
ENV_FILE=./.env.deploy ./scripts/build-artifacts.sh
docker build -f frontend/Dockerfile -t devhub/frontend:${GIT_SHA} frontend
```

OIDC 관련 URL(`OIDC_ISSUER_URL`, `OIDC_REDIRECT_URI`, `NEXT_PUBLIC_OIDC_ISSUER_URL` 등)은
런타임 env + `/api/runtime-config` 경로를 우선 사용하고, host build 는 fallback 값만
bundle 에 넣는다.

Dockerfile 핵심 형태:

```dockerfile
FROM node:20-alpine
COPY .next/standalone ./
COPY .next/static ./.next/static
COPY public ./public
CMD ["node", "server.js"]
```

### 11.3 빌드 ↔ 런타임 환경변수 구분

| 변수 | build-time 인라인 | runtime env |
| --- | --- | --- |
| `BACKEND_API_URL` | YES (Dockerfile ARG/ENV) | YES (컨테이너 env) |
| `OIDC_ISSUER_URL` / `OIDC_REDIRECT_URI` | NO | YES — runtime-config route 가 동적 응답 |
| `NEXT_PUBLIC_OIDC_ISSUER_URL` | fallback 용도 | YES — runtime-config 미사용/실패 시 fallback |

운영 기본 원칙: frontend 이미지는 환경간 재사용하고, OIDC endpoint 는 런타임 주입으로 분리한다.
host build 는 `scripts/build-artifacts.sh` 가 담당하고, Dockerfile 은 결과물만 복사한다.

## 12. Keycloak realm 운영 보안 — wildcard 좁히기 SOP (issue #238 P3-1)

`infra/idp/keycloak-realm.json` 의 client 설정에 dev 친화 wildcard 가 포함되어 있어 운영 배포 전 specific 도메인으로 좁혀야 한다.

### 12.1 현재 (dev 친화) 설정

```json
{
  "redirectUris": [
    "https://*/devhub/*",
    "http://*/devhub/*",
    "http://localhost:3000/*",
    "http://localhost:3000/auth/callback",
    "http://localhost:3000/devhub/auth/callback"
  ],
  "webOrigins": [
    "+",
    "http://localhost:3000"
  ],
  "attributes": {
    "post.logout.redirect.uris": "https://*/devhub/*##http://*/devhub/*##http://localhost:3000/*##http://localhost:3000/##http://localhost:3000/devhub/"
  }
}
```

### 12.2 운영 (specific 도메인) 설정

```json
{
  "redirectUris": [
    "https://devhub.example.com/devhub/auth/callback"
  ],
  "webOrigins": [
    "https://devhub.example.com"
  ],
  "attributes": {
    "post.logout.redirect.uris": "https://devhub.example.com/devhub/##https://devhub.example.com/devhub/login"
  }
}
```

### 12.3 변경 절차 (Keycloak admin console)

1. Clients → `devhub-frontend` → Settings 탭 → **Valid Redirect URIs**:
   - wildcard 5건 제거 → `https://<운영도메인>/devhub/auth/callback` 만 유지
2. **Web Origins**:
   - `+` 제거 → `https://<운영도메인>` 명시
3. **Logout Settings** → **Valid Post Logout Redirect URIs**:
   - wildcard 패턴 (`https://*/devhub/*##http://*/devhub/*`) 제거
   - `https://<운영도메인>/devhub/##https://<운영도메인>/devhub/login` 등 명시
4. **Save** + 운영 검증:
   - 브라우저로 `https://<운영도메인>/devhub` 진입 → OIDC login → callback → logout 전체 흐름 정상 종료 확인
   - Keycloak admin event log 에 `INVALID_REDIRECT_URI` error 없음

### 12.4 추가 운영 강화 (CSRF + hostname strict)

- `KC_HOSTNAME_STRICT=true` + `KC_HOSTNAME_STRICT_HTTPS=true` (docker-compose env). issuer URL hijack 방어.
- Keycloak realm 의 `bruteForceProtected: true` 활성 (admin console → Realm settings → Security defenses → Brute Force Detection).

자세한 Keycloak 운영 SOP 는 [keycloak_operations.md](./keycloak_operations.md) 참조.
