# 배포 Preflight 체크리스트 (실수 방지 SOP)

- 문서 목적: 배포 시 반복되는 설정 실수(특히 OIDC/JWKS, base URL, redirect)를 사전에 차단한다.
- 범위: `docker-compose.deploy.yml` 기반 배포 전/중/후 점검 절차
- 대상 독자: DevHub 운영자, 릴리즈 담당자
- 상태: active
- 최종 수정일: 2026-05-22
- 관련 문서: [docker-packaging-deployment-guide.md](./docker-packaging-deployment-guide.md), [single_port_deployment.md](./single_port_deployment.md), [deploy.env.example](./deploy.env.example)

## 1. 핵심 원칙 (반드시 준수)
- `DEVHUB_OIDC_ISSUER_URL`은 **브라우저/토큰 claim 기준 public URL** 이다.
- `DEVHUB_OIDC_JWKS_URL`은 **backend 컨테이너가 실제 접근 가능한 URL** 이다.
- 위 두 값은 같을 수도 있지만, 로컬/프록시 환경에서는 **서로 다를 수 있다**.
- 컨테이너 내부에서 `localhost`는 자기 자신이다. backend 설정에 `localhost`를 넣으면 오동작 가능성이 높다.

## 2. 가장 자주 나는 실수와 정답
- 실수: `DEVHUB_OIDC_JWKS_URL=http://localhost:13000/...`
  - 문제: backend 컨테이너 안에서 `localhost:13000`은 nginx/keycloak이 아니라 backend 자신을 보게 될 수 있음.
  - 정답(로컬 compose): `http://keycloak:8080/devhub/auth/keycloak/realms/devhub/protocol/openid-connect/certs`
- 실수: `NEXT_PUBLIC_*`만 맞추고 backend OIDC 값을 다르게 둠
  - 문제: 프론트 로그인은 되지만 `/api/v1/me`가 401 루프
  - 정답: issuer/public 경로는 프론트/백엔드가 동일한 realm을 가리키게 맞춤
- 실수: basePath(`/devhub`) 누락
  - 문제: callback/login/logout 404 또는 redirect loop
  - 정답: `NEXT_PUBLIC_BASE_PATH=devhub`, redirect URI 포함 경로 일치 확인

## 3. 배포 전 체크 (Preflight)
1. env 파일 점검
- `DEVHUB_PUBLIC_BASE_URL`, `DEVHUB_OIDC_ISSUER_URL`, `NEXT_PUBLIC_OIDC_ISSUER_URL`, `OIDC_REDIRECT_URI`, `NEXT_PUBLIC_OIDC_REDIRECT_URI`가 동일 origin + `/devhub` 경로 정합인지 확인
- `DEVHUB_OIDC_JWKS_URL`이 backend에서 reachable 한 internal 주소인지 확인 (`keycloak:8080` 권장)

2. 구성 렌더 검증
- `docker compose --env-file <env> -f docker-compose.deploy.yml config`

3. reachability 검증
- 호스트에서 issuer discovery 확인
- backend 컨테이너에서 JWKS URL HTTP 200 확인

## 4. 표준 배포 순서
1. `ENV_FILE=<env> scripts/deploy-up.sh`
2. 상태 확인
- `docker compose --env-file <env> -f docker-compose.deploy.yml ps`
3. 인증 스모크 테스트
- 브라우저 로그인 후 `/devhub/admin` 또는 `/devhub/account` 진입
- API 401 루프 여부 확인 (`/api/v1/me`)

## 5. 장애 발생 시 즉시 점검
- 증상: 로그인 후 `session_expired`, `/login` 반복, `/api/v1/me` 401
- 우선 점검 순서:
1. `DEVHUB_OIDC_ISSUER_URL` vs 토큰 `iss` 일치 여부
2. `DEVHUB_OIDC_JWKS_URL` 대상이 backend 컨테이너에서 reachable 인지
3. Keycloak realm/client redirect URI에 `/devhub/auth/callback` 등록 여부
4. 브라우저 쿠키/세션 초기화 후 재시도

## 6. 로컬 compose 권장 값 (참조)
- `DEVHUB_OIDC_ISSUER_URL=http://localhost:13000/devhub/auth/keycloak/realms/devhub`
- `DEVHUB_OIDC_JWKS_URL=http://keycloak:8080/devhub/auth/keycloak/realms/devhub/protocol/openid-connect/certs`
- `NEXT_PUBLIC_BASE_PATH=devhub`
- `NEXT_PUBLIC_OIDC_REDIRECT_URI=http://localhost:13000/devhub/auth/callback`

