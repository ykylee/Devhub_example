# Docker 패키징/배포 가이드

- 문서 목적: DevHub Example에서 Docker 패키징 오류를 줄이기 위한 표준 빌드 절차와 배포 방식(이미지 배포 vs compose 배포) 선택 기준을 정의한다.
- 범위: 이미지 태깅 규칙, 빌드/푸시 절차, compose 사용 범위, 운영 권장안
- 대상 독자: 개발자, 릴리즈 담당자, 운영자
- 상태: draft
- 최종 수정일: 2026-05-15
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
export IMAGE_NAMESPACE=<owner>/<repo>
export PUBLIC_BASE_URL=http://<host>:23000
export DEVHUB_HYDRA_ADMIN_URL=http://<host>:14445
export DEVHUB_HYDRA_PUBLIC_URL=http://<host>:14444
export DEVHUB_KRATOS_PUBLIC_URL=http://<host>:14433
export DEVHUB_KRATOS_ADMIN_URL=http://<host>:14434
export HYDRA_DSN='postgres://<user>:<pw>@<db-host>:5432/<db>?sslmode=disable&search_path=hydra'
export KRATOS_DSN='postgres://<user>:<pw>@<db-host>:5432/<db>?sslmode=disable&search_path=kratos'
export DB_URL='postgres://<user>:<pw>@<db-host>:5432/<db>?sslmode=disable'
export OIDC_AUTH_URL=http://<host>:14444/oauth2/auth
export OIDC_REDIRECT_URI=${PUBLIC_BASE_URL}/auth/callback
export NEXT_PUBLIC_OIDC_AUTH_URL=http://<host>:14444/oauth2/auth
export NEXT_PUBLIC_OIDC_REDIRECT_URI=${PUBLIC_BASE_URL}/auth/callback
export FRONTEND_PUBLIC_BASE_URL=${PUBLIC_BASE_URL}
export KRATOS_PUBLIC_EXTERNAL_URL=http://<host>:14433
export KRATOS_ADMIN_EXTERNAL_URL=http://<host>:14434
export HYDRA_PUBLIC_EXTERNAL_URL=http://<host>:14444
export NGINX_PORT=23000
docker compose -f docker-compose.deploy.yml pull
docker compose -f docker-compose.deploy.yml up -d
```

`IMAGE_TAG`와 IdP/Hydra/Kratos 관련 URL은 필수다. 미지정 시 compose가 오류로 중단되도록 설정되어 있다.

### 8.1.1 변수 스키마 (권장)

- Public (브라우저가 직접 접근): `PUBLIC_BASE_URL`, `OIDC_AUTH_URL`, `OIDC_REDIRECT_URI`, `NEXT_PUBLIC_OIDC_*`
- Internal (서비스 간 통신): `DEVHUB_HYDRA_ADMIN_URL`, `DEVHUB_HYDRA_PUBLIC_URL`, `DEVHUB_KRATOS_PUBLIC_URL`, `DEVHUB_KRATOS_ADMIN_URL`, `BACKEND_API_URL`, `KRATOS_HOOK_TARGET_URL`
- DB: `DB_URL`, `HYDRA_DSN`, `KRATOS_DSN`

`localhost`는 fallback일 뿐 표준값이 아니다. 서버를 분리 배치하는 경우에는 위 3축을 환경별로 명시 주입한다.

### 8.1.2 DB 모드 선택

- 번들 DB 모드 (`local-db` profile): compose 내부 `db`(postgres:15) 포함 기동
- 외부 DB 모드 (default): compose 내부 `db` 미기동, `DB_URL`/`HYDRA_DSN`/`KRATOS_DSN`을 외부 DSN으로 지정

예시:

```sh
# 1) 번들 DB 모드
docker compose -f docker-compose.deploy.yml --profile local-db up -d

# 2) 외부 DB 모드
docker compose -f docker-compose.deploy.yml up -d
```

`docker-compose.deploy.yml`은 `nginx`를 포함한다. 외부 진입은 `http://<host>:${NGINX_PORT}` 하나로 통일하고, `frontend`/`backend-core` 포트는 운영망에서 필요 시에만 노출한다.

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
- docker 전용 이슈 2: Hydra `urls.consent`를 내부 DNS(`http://backend-core:8080/...`)로 두면 외부 브라우저가 `backend-core`를 해석하지 못한다.
  - 증상: 로그인 후 `DNS_PROBE_FINISHED_NXDOMAIN` (`backend-core`)
  - 대응: deploy Hydra 설정에서 `consent`를 외부 진입 URL(`__FRONTEND_PUBLIC_BASE_URL__/api/v1/auth/consent`)로 설정한다.
- 공통 설정 이슈: OIDC redirect URI를 `/api/auth/callback`로 주면 라우트 불일치가 발생한다.
  - 기준 라우트: `/auth/callback`
  - 대응: `OIDC_REDIRECT_URI`, `NEXT_PUBLIC_OIDC_REDIRECT_URI`, Hydra client `redirect_uris`를 동일하게 `/auth/callback`으로 맞춘다.

최소 검증 순서:

1. `curl http://<host>:<nginx-port>/api/runtime-config`에서 OIDC URL/redirect 값 확인
2. `curl http://<host>:<hydra-public-port>/.well-known/openid-configuration` 확인
3. Playwright 단건 검증  
   `PLAYWRIGHT_BASE_URL=http://<host>:<nginx-port> KRATOS_ADMIN_URL=http://<host>:<kratos-admin-port> npm run e2e -- tests/e2e/auth.spec.ts --grep "developer lands on /developer"`
