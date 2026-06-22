# Dogfood 환경 셋업 가이드

- 문서 목적: `colima` 기반 DevHub dogfood 전용 시뮬레이션 환경을 분리된 포트와 컨테이너 이름으로 구성하는 절차를 제공한다.
- 범위: `PostgreSQL + Keycloak` dogfood 컨테이너 기동, native 앱 연결, 환경 변수, 기본 검증, 트러블슈팅
- 대상 독자: 개발자, QA, AI 에이전트, 로컬 환경 운영자
- 상태: draft
- 최종 수정일: 2026-06-05
- 관련 문서: [Dogfood 환경 문서](./README.md), [개발 환경 구성 가이드](../setup/environment-setup.md), [Keycloak 운영 SOP](../setup/keycloak_operations.md), [E2E Test Guide](../setup/e2e-test-guide.md)

## 1. 목적과 구성

이 가이드는 **개발용 컨테이너와 충돌하지 않는 별도 dogfood 스택**을 전제로 한다.

- Docker runtime: `colima`
- dogfood compose project: `devhub-dogfood`
- 인프라 컨테이너: `PostgreSQL 15`, `Keycloak 26`
- 애플리케이션 실행 방식: `backend-core`, `frontend` 는 우선 native 실행 (2026-06-22 M-v0.2.2 backend-ai 폐기 반영)
- 외부 연동: `Gitea` 는 `https://homelab.ddn777.synology.me/gitea` 사용

## 2. 파일과 자산

### 2.1 Dogfood 자산

다음 파일은 dogfood 용으로 사용한다.

| 파일 | 용도 | 비고 |
| --- | --- | --- |
| `docker-compose.colima.yml` | dogfood DB/Keycloak 컨테이너 구성 | 저장소 추적 파일, compose project 명 `devhub-dogfood` |
| `.env.dogfood` | dogfood 포트, DSN, OIDC, Gitea 토큰 | 로컬 전용, `.gitignore` 대상 |

### 2.2 비밀값 정책

- `.env.dogfood` 에는 `GITEA_TOKEN`, `DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET`, `DEVHUB_E2E_KEYCLOAK_ADMIN_CLIENT_SECRET`, `GITEA_WEBHOOK_SECRET` 가 들어간다.
- 이 파일은 git 에 커밋하지 않는다.
- Gitea PAT 가 외부로 노출됐다고 판단되면 즉시 폐기하고 새 토큰으로 교체한다.

## 3. 포트와 네이밍

### 3.1 포트 맵

| 구성 요소 | dogfood 포트 | 기본 개발 포트와의 관계 |
| --- | --- | --- |
| Frontend | `13000` | `3000` 회피 |
| Backend Core | `18080` | `8080` 회피 |
| Backend AI | `18000` | `8000` 회피 |
| Keycloak | `18180` | `8180` 회피 |
| PostgreSQL | `15433` | `5433` 회피 |

### 3.2 컨테이너 이름

dogfood compose project 가 `devhub-dogfood` 이므로 컨테이너는 다음 형태로 생성된다.

- `devhub-dogfood-db-1`
- `devhub-dogfood-db-init-1`
- `devhub-dogfood-keycloak-1`

기존 개발용 컨테이너(`devhub-db`, `devhub-keycloak`) 와 동시에 떠 있어도 충돌하지 않는다.

## 4. 사전 조건

### 4.1 host runtime

| 항목 | 권장 |
| --- | --- |
| macOS + Apple Silicon | 지원 |
| `colima` | 설치 및 실행 |
| `docker` CLI | `colima` context 연결 |
| `go` | 1.22+ |
| `python3` | 3.12+ 권장 |
| `node` / `npm` | Node 20 LTS 권장 |
| `migrate` | 설치 |

### 4.2 `colima` 확인

```sh
colima status
docker context ls
```

정상 기준:

- `colima is running`
- `docker context` 의 현재 항목이 `colima`

## 5. dogfood 인프라 컨테이너 기동

### 5.1 권장 진입점: `scripts/dogfood.sh`

가장 간단한 방법은 전용 wrapper 스크립트를 쓰는 것이다.

```sh
./scripts/dogfood.sh up
```

이 스크립트는 아래 작업을 순서대로 수행한다.

1. dogfood `PostgreSQL + Keycloak` 컨테이너 기동
2. Keycloak discovery readiness 확인
3. DB migration 실행
4. `scripts/setup-keycloak.sh` 로 realm/client 정합
5. `backend-core`, `frontend` native 실행 (2026-06-22 M-v0.2.2 backend-ai 폐기 반영)

기본 `up` 은 기존 이미지를 재사용한다. Keycloak Dockerfile 또는 SPI 소스를 바꿨을 때만 명시적으로 재빌드한다.

```sh
./scripts/dogfood.sh up --build
```

종료:

```sh
./scripts/dogfood.sh down
```

상태 확인:

```sh
./scripts/dogfood.sh status
```

빠른 smoke:

```sh
./scripts/dogfood.sh smoke
```

온보딩 smoke:

```sh
./scripts/dogfood.sh test-onboarding
```

관리자 self dogfooding 시나리오:

```sh
./scripts/dogfood.sh test-self-dogfood
```

초기화:

```sh
./scripts/dogfood.sh reset-db
```

설명:

- dogfood compose 볼륨을 제거해 `PostgreSQL` 과 `Keycloak` 데이터를 처음 상태로 되돌린다
- 다음 `./scripts/dogfood.sh up` 에서 migration, realm import, client setup, seed 준비가 다시 수행된다

완전 초기화:

```sh
./scripts/dogfood.sh reset-all
```

설명:

- `reset-db` 동작 포함
- `artifacts/dogfood/logs/`, `.pids/dogfood/` 까지 함께 삭제
- 테스트 중 남은 런타임 흔적까지 비우고 싶을 때 사용

로그 확인:

```sh
./scripts/dogfood.sh logs
./scripts/dogfood.sh logs backend
./scripts/dogfood.sh logs frontend
./scripts/dogfood.sh logs ai
```

native 앱 로그는 `artifacts/dogfood/logs/` 에 저장되고, PID 파일은 `.pids/dogfood/` 에 저장된다.

### 5.2 수동 compose 기동

```sh
docker compose --env-file .env.dogfood -f docker-compose.colima.yml -p devhub-dogfood up -d --build
```

설명:

- `db` 는 `15433`
- `keycloak` 은 `18180`
- Keycloak 은 `infra/idp/Dockerfile.keycloak` 로 커스텀 SPI 포함 이미지를 빌드한다
- 현재는 build context 를 `infra/idp/` 로 제한해 저장소 루트 전체 전송을 피한다
- 이 커스텀 build 는 dogfood 전용이다. 현재 CI 와 `docker-compose.deploy.yml` 의 local-idp/deploy 경로는 stock `quay.io/keycloak/keycloak:26.0` 이미지를 유지한다

### 5.3 상태 확인

```sh
docker compose --env-file .env.dogfood -f docker-compose.colima.yml -p devhub-dogfood ps
```

정상 예시:

- `db` → `healthy`
- `keycloak` → 초기 수십 초 동안 `starting`, 이후 `healthy`

### 5.4 헬스 체크

```sh
curl http://localhost:18180/devhub/auth/keycloak/realms/devhub/.well-known/openid-configuration
pg_isready -h localhost -p 15433 -U postgres -d devhub
```

## 6. DB 마이그레이션

`.env.dogfood` 기준으로 실행한다.

```sh
set -a
source .env.dogfood
set +a
migrate -path backend-core/migrations -database "$MIGRATE_DB_URL" up
```

정상 기준:

- 최초 실행: migration 적용
- 이후 재실행: `no change`

## 7. Keycloak client / seeder 정합

dogfood Keycloak 을 띄운 직후 `scripts/setup-keycloak.sh` 로 realm/client 를 정합한다.

```sh
KEYCLOAK_URL=http://localhost:18180/devhub/auth/keycloak \
KC_BOOTSTRAP_ADMIN_USERNAME=admin \
KC_BOOTSTRAP_ADMIN_PASSWORD=admin \
DEVHUB_FRONTEND_ORIGIN=http://localhost:13000 \
DEVHUB_FRONTEND_BASEPATH= \
./scripts/setup-keycloak.sh
```

설명:

- native frontend 는 basePath 없이 `13000` 에 붙이므로 `DEVHUB_FRONTEND_BASEPATH=` 로 빈 값을 준다
- 이 스크립트는 `devhub-frontend`, `devhub-backend`, `devhub-e2e-seeder` client 를 upsert 한다
- macOS 에서도 동작하도록 readiness wait 가 내장돼 있다

## 8. native 앱 실행

### 8.1 backend-core

```sh
set -a
source .env.dogfood
set +a
cd backend-core
PORT=18080 go run .
```

### 8.2 ~~backend-ai (Python, :18000)~~ — **2026-06-22 M-v0.2.2 폐기 (PR #663)**

> **폐기 사유**: backend-ai/ 디렉터리 (placeholder Python AI service + Dockerfile + dev-up.sh 정합) 가 production wiring 없이 v0.1.x ~ v0.2.0 PoC 기간 dead state 였음. v0.2.0+ AI/ML scope 은 `backend-knowledge/` 의 §3.7 Pi LLM enrich + §3.5.7 cross-link 자동 resolution 으로 대체. dogfood 환경은 2-tier (backend-core + frontend) 로 단순화.
>
> 이전 절차 (보존 reference, v0.2.2 이전):
> ```sh
> set -a
> source .env.dogfood
> set +a
> cd backend-ai
> uvicorn main:app --host 0.0.0.0 --port 18000
> ```

### 8.3 frontend

```sh
set -a
source .env.dogfood
set +a
cd frontend
PORT=13000 npm run dev
```

## 9. 기본 검증 순서

### 9.1 서비스 헬스

```sh
curl http://localhost:18080/health
curl http://localhost:18000/health
curl -I http://localhost:13000/
curl http://localhost:18180/devhub/auth/keycloak/realms/devhub/.well-known/openid-configuration
```

### 9.2 Gitea 연결성

```sh
curl -H "Authorization: token $GITEA_TOKEN" \
  https://homelab.ddn777.synology.me/gitea/api/v1/version
```

### 9.3 E2E seed 준비

`frontend/tests/e2e/global-setup.ts` 가 자동 시드를 수행하므로 아래 값이 `.env.dogfood` 에 있어야 한다.

- `DEVHUB_KEYCLOAK_ADMIN_URL`
- `DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID`
- `DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET`
- `DEVHUB_E2E_KEYCLOAK_ADMIN_CLIENT_ID`
- `DEVHUB_E2E_KEYCLOAK_ADMIN_CLIENT_SECRET`
- `DB_URL`

### 9.4 신규 계정 온보딩 smoke

dogfood 에서 신규 사용자 1명을 만들어 온보딩까지 검증하려면 다음 자산을 사용한다.

- 계정 생성: `./scripts/dogfood-create-user.sh`
- 자동 검증: `frontend/tests/e2e/dogfood-onboarding-smoke.spec.ts`
- 실행 가이드: [onboarding_smoke.md](./onboarding_smoke.md)

수동 계정 생성 예시:

```sh
./scripts/dogfood-create-user.sh dogfood-manual@example.com 'ChangeMe-12345!' 'Dogfood Manual'
```

자동 smoke 실행 예시:

```sh
./scripts/dogfood.sh test-onboarding
```

## 10. 운영 팁

### 10.1 shell export

zsh/bash 에서는 다음 패턴이 가장 간단하다.

```sh
set -a
source .env.dogfood
set +a
```

### 10.2 종료

dogfood 컨테이너만 내릴 때:

```sh
docker compose --env-file .env.dogfood -f docker-compose.colima.yml -p devhub-dogfood down
```

볼륨까지 삭제할 때:

```sh
docker compose --env-file .env.dogfood -f docker-compose.colima.yml -p devhub-dogfood down -v
```

wrapper 스크립트를 사용할 경우에는 다음 명령이 전체 종료 기본값이다.

```sh
./scripts/dogfood.sh down
```

DB 와 IdP 데이터를 완전히 초기화할 때:

```sh
./scripts/dogfood.sh reset-db
```

로그/PID 흔적까지 포함한 완전 초기화:

```sh
./scripts/dogfood.sh reset-all
```

## 11. 트러블슈팅

| 증상 | 원인 후보 | 대응 |
| --- | --- | --- |
| Keycloak 이 오래 `starting` 상태 | realm import 직후 healthcheck 지연 | `docker logs devhub-dogfood-keycloak-1` 확인 후 30~60초 추가 대기 |
| `setup-keycloak.sh` 실패 | bootstrap admin / frontend origin 불일치 | `KEYCLOAK_URL`, `DEVHUB_FRONTEND_ORIGIN=http://localhost:13000` 재확인 |
| `token has invalid issuer` | backend env 와 Keycloak port 불일치 | `DEVHUB_OIDC_ISSUER_URL=http://localhost:18180/...` 확인 |
| 로그인 후 `/api/v1/me` 401 | JWKS URL / issuer mismatch | `DEVHUB_OIDC_JWKS_URL`, `DEVHUB_OIDC_ISSUER_URL` 동시 점검 |
| Gitea provider 등록 실패 | PAT scope 부족 또는 base URL 오입력 | `https://homelab.ddn777.synology.me/gitea` + 새 PAT 재확인 |
| 개발용 컨테이너와 헷갈림 | compose project 분리 미인지 | `docker ps` 에서 `devhub-dogfood-*` prefix 확인 |

## 12. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-06-05 | dogfood 전용 `colima` 분리 환경 기준 초판 작성. 포트 `13000/18080/18000/18180/15433`, compose project `devhub-dogfood`, 외부 Gitea 연동, native 앱 연결 절차 반영. |
