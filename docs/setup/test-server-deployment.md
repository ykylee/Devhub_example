# 테스트 서버 배포 가이드 (Native, Keycloak OIDC)

- 문서 목적: DevHub Example 을 단일 테스트 서버에 native binary 로 빌드·배포·기동하는 표준 절차를 정의한다.
- 범위: 사전 준비, 빌드, 환경변수, 기동 순서, OIDC client 등록, 시드 사용자, 헬스체크, 로그인 검증.
- 대상 독자: 테스트 서버 운영자, QA, 신규 환경 부트스트랩 담당.
- 상태: active
- 최종 수정일: 2026-05-18
- 관련 문서: [개발 환경 구성](./environment-setup.md), [백엔드 API 계약](../backend_api_contract.md), [아키텍처](../architecture.md), [E2E 가이드](./e2e-test-guide.md)

## 0. 원칙

- Docker 미사용. 모든 프로세스는 host OS 에서 native 로 실행한다.
- 인증 source-of-truth 는 Keycloak OIDC 이다.
- `/api/v1/auth/*`, `logout_challenge` 전제 절차는 사용하지 않는다.

## 1. 사전 준비

### 1.1 런타임

| 항목 | 권장 | 확인 |
| --- | --- | --- |
| Go | 1.22+ | `go version` |
| Node.js | 20 LTS | `node --version` |
| npm | 10+ | `npm --version` |
| PostgreSQL | 15+ | `psql --version` |
| Keycloak | 26+ | `keycloak --version` 또는 배포 방식별 확인 |

### 1.2 데이터베이스

```sh
createdb -U postgres devhub
```

DevHub backend 마이그레이션:

```sh
MIGRATE_DB_URL="postgres://devhub:<pw>@<host>:5432/devhub?sslmode=disable" make migrate-up
```

## 2. 빌드

### 2.1 backend-core

```sh
cd backend-core
go build -o bin/devhub-backend .
```

### 2.2 frontend

```sh
cd frontend
npm ci
npm run build
```

## 3. 환경변수

### 3.1 backend-core

| 변수 | 예시 | 설명 |
| --- | --- | --- |
| `PORT` | `8080` | API 포트 |
| `DEVHUB_ENV` | `prod` | 운영 모드 |
| `DB_URL` | `postgres://devhub:<pw>@localhost:5432/devhub?sslmode=disable` | DevHub DB |
| `DEVHUB_OIDC_ISSUER_URL` | `http://localhost:8081/realms/devhub` | OIDC issuer |
| `DEVHUB_OIDC_CLIENT_ID` | `devhub-backend` | backend OIDC client id |
| `DEVHUB_OIDC_CLIENT_SECRET` | `<secret>` | backend OIDC client secret |
| `DEVHUB_OIDC_JWKS_URL` | (선택) | 생략 시 issuer discovery 사용 |
| `DEVHUB_KEYCLOAK_ADMIN_URL` | `http://localhost:8081` | Keycloak admin base |
| `DEVHUB_KEYCLOAK_ADMIN_REALM` | `devhub` | admin realm |
| `DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID` | `devhub-admin` | admin client id |
| `DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET` | `<secret>` | admin client secret |
| `DEVHUB_AUTH_DEV_FALLBACK` | (미설정) | prod 에서 사용 금지 |

### 3.2 frontend

| 변수 | 예시 | 설명 |
| --- | --- | --- |
| `BACKEND_API_URL` | `http://localhost:8080` | Next rewrite 대상 |
| `NEXT_PUBLIC_IDP_PROVIDER` | `keycloak` | IdP provider |
| `NEXT_PUBLIC_OIDC_ISSUER_URL` | `http://localhost:8081/realms/devhub` | OIDC issuer |
| `NEXT_PUBLIC_OIDC_AUTH_URL` | `http://localhost:8081/realms/devhub/protocol/openid-connect/auth` | authorize endpoint |
| `NEXT_PUBLIC_OIDC_CLIENT_ID` | `devhub-frontend` | frontend client id |
| `NEXT_PUBLIC_OIDC_REDIRECT_URI` | `http://localhost:3000/auth/callback` | callback URI |
| `NEXT_PUBLIC_OIDC_SCOPE` | `openid offline_access email profile` | 요청 scope |

## 4. Keycloak 설정

최소 필요 항목:

1. realm 생성: `devhub`
2. client 생성: `devhub-frontend` (public)
3. client 생성: `devhub-backend` (confidential)
4. client 생성: `devhub-admin` (service account)
5. redirect URI 등록: `http://localhost:3000/auth/callback`
6. post logout redirect URI 등록: `http://localhost:3000/`

## 5. 시드 사용자

테스트용 사용자 3명(예: alice/bob/charlie)을 Keycloak과 DevHub `users`에 정합하게 준비한다.

권장: `frontend/tests/e2e/global-setup.ts`가 사용하는 시드 정책(`docs/setup/e2e-test-guide.md`)을 따른다.

## 6. 기동 순서

1. PostgreSQL
2. Keycloak
3. backend-core (`bin/devhub-backend`)
4. frontend (`npm run start` 또는 `npm run dev`)

## 7. 헬스체크

```sh
curl http://localhost:8080/health
curl http://localhost:3000/api/runtime-config
curl http://localhost:8081/realms/devhub/.well-known/openid-configuration
```

정상 기준:
- `/health` 200
- runtime-config 에 OIDC issuer/auth/redirect 값 노출
- OIDC discovery JSON 응답

## 8. 로그인 스모크 테스트

1. `http://localhost:3000/login` 진입
2. OIDC authorize redirect 확인
3. 인증 완료 후 `/auth/callback` 경유
4. 역할별 기본 랜딩(`/developer`, `/manager`, `/admin`) 확인
5. Sign Out 후 `/login` 재진입 시 자격증명 재요청 확인

## 9. 트러블슈팅

| 증상 | 점검 |
| --- | --- |
| `/api/v1/me` 401 | `DEVHUB_OIDC_ISSUER_URL`, audience/claim mapping, 서버 시간 오차 확인 |
| 로그인 후 잘못된 랜딩 | DevHub `users.role` 값과 subject 매핑 확인 |
| Sign Out 후 자동 재로그인 | end-session endpoint, `id_token` 저장/전달 여부 확인 |
| callback 오류 | client redirect URI / post logout URI 불일치 확인 |

## 10. 보안 체크리스트

- 운영에서는 HTTPS 필수
- OIDC client secret 을 git 에 커밋하지 않음
- Keycloak admin endpoint 접근 제어
- 테스트용 공용 계정/기본 비밀번호 제거

## 11. Docker 단일 포트 배포 (issue #238)

`docker-compose.deploy.yml` 기준 운영 모드에서는 외부 노출 포트를 `nginx` 하나로 제한한다.

- 외부 노출: `80` / `443` (`nginx`)
- 내부 통신 전용: `frontend:3000`, `backend-core:8080`, `keycloak:8080`, `backend-ai:8000` (`devhub-internal` 네트워크)
- URL 기준:
  - `https://<host>/devhub/` → frontend
  - `https://<host>/devhub/api/*` → backend
  - `https://<host>/devhub/auth/keycloak/*` → keycloak

필수 환경값:

- `NEXT_PUBLIC_BASE_PATH=devhub`
- `NEXT_PUBLIC_OIDC_REDIRECT_URI=https://<host>/devhub/auth/callback`
- `DEVHUB_OIDC_ISSUER_URL=https://<host>/devhub/auth/keycloak/realms/devhub`
- `DEVHUB_TRUSTED_PROXIES=172.16.0.0/12` (또는 운영 네트워크 CIDR)
