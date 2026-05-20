# Docker 패키징/배포 가이드

> ⚠ 2026-05-20 정정: 본 문서는 Keycloak-only 배포 기준이다. Hydra/Kratos 변수/절차는 사용하지 않는다.

- 문서 목적: DevHub Example에서 Docker 패키징 오류를 줄이기 위한 표준 빌드 절차와 배포 방식(이미지 배포 vs compose 배포) 선택 기준을 정의한다.
- 범위: 이미지 태깅 규칙, 빌드/푸시 절차, compose 사용 범위, 운영 권장안
- 대상 독자: 개발자, 릴리즈 담당자, 운영자
- 상태: draft
- 최종 수정일: 2026-05-20
- 관련 문서: [개발 환경 구성 가이드](./environment-setup.md), [테스트 서버 배포 가이드](./test-server-deployment.md), [ADR-0003](../adr/0003-no-docker-policy-ci-scope.md)

## 1. 현재 저장소 기준 Docker 자산

현재 저장소에는 다음 Docker 자산이 존재한다.

- `backend-core/Dockerfile`
- `backend-ai/Dockerfile`
- `frontend/Dockerfile`
- `docker-compose.yml`
- `docker-compose.local.yml`

compose는 서비스 간 연결/개발 실행에 유용하지만, 배포 산출물의 재현성과 추적성은 이미지 중심으로 관리하는 편이 안정적이다.

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
# backend-core
docker build -f backend-core/Dockerfile -t devhub/backend-core:${GIT_SHA} backend-core

# backend-ai
docker build -f backend-ai/Dockerfile -t devhub/backend-ai:${GIT_SHA} backend-ai

# frontend
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

### 8.1 배포용 compose 실행 예시

```sh
export IMAGE_TAG=<git-sha-or-release-tag>
export IMAGE_REPO_PREFIX=ghcr.io/<owner>/<repo>   # 로컬 검증 시 devhub
export PUBLIC_BASE_URL=https://<host>/devhub
export DB_URL='postgres://<user>:<pw>@<db-host>:5432/<db>?sslmode=disable'
export DEVHUB_IDP_PROVIDER=keycloak
export DEVHUB_OIDC_ISSUER_URL=https://<host>/devhub/auth/keycloak/realms/devhub
export DEVHUB_OIDC_CLIENT_ID=devhub-web
export DEVHUB_OIDC_CLIENT_SECRET='<oidc-client-secret>'
export DEVHUB_KEYCLOAK_ADMIN_URL=http://keycloak:8080
export DEVHUB_KEYCLOAK_ADMIN_REALM=devhub
export DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID=devhub-admin
export DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET='<keycloak-admin-secret>'
export NEXT_PUBLIC_IDP_PROVIDER=keycloak
export NEXT_PUBLIC_OIDC_ISSUER_URL=https://<host>/devhub/auth/keycloak/realms/devhub
export NEXT_PUBLIC_BASE_PATH=devhub
export NEXT_PUBLIC_OIDC_REDIRECT_URI=https://<host>/devhub/auth/callback
export NGINX_HTTP_PORT=80
export NGINX_HTTPS_PORT=443
docker compose -f docker-compose.deploy.yml pull
docker compose -f docker-compose.deploy.yml up -d
```

로컬 빌드 이미지를 그대로 사용할 때:

```sh
export IMAGE_TAG=$(git rev-parse --short HEAD)
export IMAGE_REPO_PREFIX=devhub
docker compose -f docker-compose.deploy.yml up -d
```

`IMAGE_TAG`와 Keycloak/OIDC 관련 URL은 필수다. 미지정 시 compose가 오류로 중단되도록 설정되어 있다.
`DB_URL`도 필수다. 미지정 시 compose가 오류로 중단된다.
`DEVHUB_OIDC_CLIENT_SECRET`, `DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET`은 운영 배포 필수값이며, 예시 기본값(`dev-token` 류) 사용은 금지한다.
`DEVHUB_AUTH_DEV_FALLBACK` 기본값은 `0`(비활성)이며, 배포 환경에서 `1`로 켜지지 않도록 유지한다.

### 8.1.1 변수 스키마 (권장)

- Public (브라우저가 직접 접근): `PUBLIC_BASE_URL`, `NEXT_PUBLIC_IDP_PROVIDER`, `NEXT_PUBLIC_OIDC_ISSUER_URL`
- Internal (서비스 간 통신): `DEVHUB_OIDC_*`, `DEVHUB_KEYCLOAK_ADMIN_*`, `BACKEND_API_URL`
- DB: `DB_URL`

`localhost`는 fallback일 뿐 표준값이 아니다. 서버를 분리 배치하는 경우에는 위 3축을 환경별로 명시 주입한다.

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

## 11. frontend Dockerfile build args — `NEXT_PUBLIC_*` build-time inline (issue #238 P2-2)

Next.js 의 `NEXT_PUBLIC_*` 환경변수는 **build time** 에 client bundle 에 inline 된다. docker-compose 의 `environment:` 는 runtime env 만 set 하므로 frontend 이미지를 빌드할 때 `--build-arg` 로 전달해야 client 코드에 반영된다.

### 11.1 build args 매트릭스

| build arg | 용도 | 운영 예시 |
| --- | --- | --- |
| `NEXT_PUBLIC_BASE_PATH` | Next.js basePath (`next.config.ts:9`) | `devhub` |
| `NEXT_PUBLIC_OIDC_ISSUER_URL` | client OIDC issuer | `https://devhub.example.com/devhub/auth/keycloak/realms/devhub` |
| `NEXT_PUBLIC_OIDC_CLIENT_ID` | client id | `devhub-frontend` |
| `NEXT_PUBLIC_OIDC_REDIRECT_URI` | callback URL | `https://devhub.example.com/devhub/auth/callback` |
| `NEXT_PUBLIC_OIDC_SCOPE` | OIDC scope | `openid offline_access email profile` |
| `NEXT_OUTPUT` | standalone server build | `standalone` |

### 11.2 build 명령 예시

```bash
docker build -f frontend/Dockerfile \
  --build-arg NEXT_PUBLIC_BASE_PATH=devhub \
  --build-arg NEXT_PUBLIC_OIDC_ISSUER_URL=https://devhub.example.com/devhub/auth/keycloak/realms/devhub \
  --build-arg NEXT_PUBLIC_OIDC_CLIENT_ID=devhub-frontend \
  --build-arg NEXT_PUBLIC_OIDC_REDIRECT_URI=https://devhub.example.com/devhub/auth/callback \
  --build-arg NEXT_OUTPUT=standalone \
  -t devhub/frontend:${GIT_SHA} frontend
```

Dockerfile 내부 (`frontend/Dockerfile`, 운영자 책임 자산이라 git 추적 외) 에서는:

```dockerfile
FROM node:20-alpine AS builder
ARG NEXT_PUBLIC_BASE_PATH
ARG NEXT_PUBLIC_OIDC_ISSUER_URL
ARG NEXT_PUBLIC_OIDC_CLIENT_ID
ARG NEXT_PUBLIC_OIDC_REDIRECT_URI
ARG NEXT_PUBLIC_OIDC_SCOPE
ARG NEXT_OUTPUT=standalone

ENV NEXT_PUBLIC_BASE_PATH=$NEXT_PUBLIC_BASE_PATH
ENV NEXT_PUBLIC_OIDC_ISSUER_URL=$NEXT_PUBLIC_OIDC_ISSUER_URL
# ... (나머지 ARG → ENV mapping)
ENV NEXT_OUTPUT=$NEXT_OUTPUT

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build
```

### 11.3 빌드 ↔ 런타임 환경변수 구분

| 변수 | build-time 인라인 | runtime env |
| --- | --- | --- |
| `NEXT_PUBLIC_*` | **YES** — bundle 에 박힘 | (no-op — bundle 이 이미 build 된 값 가짐) |
| `BACKEND_API_URL` | NO | YES — Next.js standalone server 가 동적 사용 |
| `OIDC_ISSUER_URL` / `OIDC_REDIRECT_URI` | NO | YES — runtime-config route 가 동적 응답 |

운영자가 환경 별로 frontend image 를 별도 빌드 (`https://stage.devhub.example.com` vs `https://devhub.example.com`) 또는 빌드 시점에 placeholder 사용 후 runtime config endpoint (`/api/runtime-config`) 로 동적 override 하는 패턴 (codex PR #245 도입) 채택.

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
    "post.logout.redirect.uris": "https://devhub.example.com/devhub/##https://devhub.example.com/devhub/auth/login"
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
   - `https://<운영도메인>/devhub/##https://<운영도메인>/devhub/auth/login` 등 명시
4. **Save** + 운영 검증:
   - 브라우저로 `https://<운영도메인>/devhub` 진입 → OIDC login → callback → logout 전체 흐름 정상 종료 확인
   - Keycloak admin event log 에 `INVALID_REDIRECT_URI` error 없음

### 12.4 추가 운영 강화 (CSRF + hostname strict)

- `KC_HOSTNAME_STRICT=true` + `KC_HOSTNAME_STRICT_HTTPS=true` (docker-compose env). issuer URL hijack 방어.
- Keycloak realm 의 `bruteForceProtected: true` 활성 (admin console → Realm settings → Security defenses → Brute Force Detection).

자세한 Keycloak 운영 SOP 는 [keycloak_operations.md](./keycloak_operations.md) 참조.
