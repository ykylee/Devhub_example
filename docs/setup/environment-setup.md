# 개발 환경 구성 가이드

- 문서 목적: DevHub Example 의 개발/검증 환경을 구성하는 표준 절차를 docker 모드와 native(no-docker) 모드 두 갈래로 제공한다.
- 범위: 사전 준비, 데이터베이스 기동, 백엔드 2종(backend-core/backend-ai) 실행, frontend 실행, 마이그레이션, 검증
- 대상 독자: 신규 개발자, 환경 트러블슈팅 담당자
- 상태: draft
- 최종 수정일: 2026-05-08
- 관련 문서: [Makefile](../../Makefile), [tech_stack.md](../tech_stack.md), [ai-workflow/memory/environments/homelab-postgresql.md](../../ai-workflow/memory/environments/homelab-postgresql.md)

## 0. 자산 관리 원칙

`docker-compose.yml`, 각 서비스의 `Dockerfile`, `.dockerignore` 등 컨테이너 자산은 작업 환경마다 제약(사내 네트워크/SSL inspection/이미지 mirror 정책 등)이 다르므로 **git 추적에서 제외**한다 (`.gitignore` 의 `DEV ENVIRONMENT` 섹션 참조). 본 가이드는 두 모드 어디서나 동일한 결과(API 8080, AI 8000, DB 5432, Frontend 3000)를 만들 수 있도록 절차를 정의한다.

- 각 환경에서 사용한 docker 자산은 사내 위키 / 팀 노션 / 개인 작업 디렉터리 등 git 외부에 보관한다.
- 정책 갱신/예외가 필요해지면 본 문서와 `ai-workflow/memory/feedback_no_docker.md`(claude memory) 를 함께 갱신한다.

## 1. 공통 사전 준비

| 항목 | 권장 버전 | 비고 |
| --- | --- | --- |
| Go | 1.22+ | backend-core |
| Python | 3.11+ | backend-ai |
| Node.js | 20 LTS | frontend (Next.js 15) |
| PostgreSQL | 15 | DB |
| `migrate` CLI | v4.19.1 | `make migrate-tools` 로 설치 |

`make setup` 으로 backend-core / backend-ai / frontend 의존을 한 번에 설치할 수 있다 (docker 와 무관, native 셋업).

## 2. Native (no-docker) 모드 — default

### 2.1 PostgreSQL 준비

호스트에 PostgreSQL 15 를 직접 설치하거나 시스템 서비스로 기동한다 (Windows Service / systemd / launchd / Homebrew 등). 기본 접속 정보:

- DSN: `postgres://user:pass@localhost:5432/devhub?sslmode=disable`
- 다른 값을 쓰려면 `DB_URL`, `MIGRATE_DB_URL` 환경변수로 override

DB 기동 후 마이그레이션:

```sh
make migrate-tools   # 최초 1회
make migrate-up
```

### 2.2 backend-core (Go, :8080)

```sh
cd backend-core
go run .
```

필요 환경변수 (예시):

```sh
export DB_URL="postgres://user:pass@localhost:5432/devhub?sslmode=disable"
export BACKEND_AI_URL="http://localhost:8000"
export GITEA_URL="..."
export GITEA_TOKEN="..."
export GITEA_WEBHOOK_SECRET="..."
```

### 2.3 backend-ai (Python/FastAPI, :8000)

```sh
cd backend-ai
python main.py
# 또는
uvicorn main:app --host 0.0.0.0 --port 8000
```

### 2.4 frontend (Next.js, :3000)

```sh
cd frontend
npm run dev
```

`BACKEND_API_URL` 환경변수를 별도로 지정하지 않으면 `next.config.ts` 의 native default(`http://localhost:8080`) 를 사용해 backend-core 로 프록시한다. compose 환경처럼 다른 호스트명을 쓰려면 `BACKEND_API_URL` 을 export 하면 된다.

#### Auth env (PR #18 이후 적용)

frontend 의 `/login` 화면이 OIDC redirect 흐름으로 진입하므로 다음 env 를 설정한다 (default 는 PoC 용).

| 변수 | default | 용도 |
| --- | --- | --- |
| `NEXT_PUBLIC_OIDC_LOGIN_URL` | `http://127.0.0.1:4444/oauth2/auth` | Hydra public authorize endpoint |
| `NEXT_PUBLIC_OIDC_CLIENT_ID` | `devhub-frontend` | Hydra OIDC client id (Phase 13 PoC 에서 등록) |
| `NEXT_PUBLIC_OIDC_REDIRECT_URI` | `http://127.0.0.1:3000/login/callback` | Hydra → frontend callback URL |
| `NEXT_PUBLIC_OIDC_SCOPE` | `openid offline` | 요청 scope |

또한 dev 모드에서 backend `/api/v1/me` 가 401 을 반환하지 않도록 backend 측에서 `DEVHUB_AUTH_DEV_FALLBACK=1` + `X-Devhub-Actor` 헤더를 사용하거나, Hydra/Kratos PoC 가 가동 중이어야 한다.

### 2.5 검증

```sh
curl http://localhost:8080/health
curl http://localhost:8000/health
curl -I http://localhost:3000/
```

각 서비스 헬스 응답이 정상이면 이후 통합 시나리오 검증으로 진행한다.

## 3. Docker (옵션) 모드

회사/개인 환경에서 docker compose 로 환경을 묶는 편이 편한 경우, 본 저장소는 다음 자산을 **로컬에서 추가**하여 사용한다 (git 추적 안 됨).

```
backend-core/Dockerfile
backend-ai/Dockerfile
frontend/Dockerfile
docker-compose.yml
```

표준 골격은 사내 위키의 "DevHub docker compose 표준 자산" 페이지(또는 팀 공유 위치)를 참고한다. 프로젝트 루트 README 에는 별도 진입점을 두지 않는다.

기동/검증 명령은 사용자 로컬 자산이 있을 때만 의미가 있다.

```sh
docker-compose up -d
docker-compose ps
```

frontend 컨테이너에서 backend-core 컨테이너로 프록시하려면 `BACKEND_API_URL=http://backend-core:8080` 을 frontend 서비스 environment 에 명시 주입한다 — `next.config.ts` 의 native default 와는 별개로 동작한다.

## 4. 마이그레이션

| 명령 | 동작 |
| --- | --- |
| `make migrate-tools` | `migrate` CLI 설치 (1회) |
| `make migrate-create NAME=foo` | 신규 마이그레이션 파일 생성 |
| `make migrate-up` | 최신 스키마까지 적용 |
| `make migrate-down` | 한 단계 롤백 |
| `make migrate-version` | 현재 적용된 버전 확인 |

`MIGRATE_DB_URL` 을 override 하면 native / docker DB 모두 동일 명령으로 처리된다.

## 5. 트러블슈팅 단서

- **frontend 가 `/api/*` 호출 시 ECONNREFUSED**: native 모드면 backend-core 가 8080 에서 떠 있는지, docker 모드면 `BACKEND_API_URL` 이 실제 도달 가능한지 확인.
- **DNS resolution failure for `backend-core`**: docker compose 외부에서 frontend 만 단독 실행 중일 때 발생. `BACKEND_API_URL` 을 unset 하거나 `http://localhost:8080` 으로 export.
- **DB 연결 실패**: `MIGRATE_DB_URL` 과 `DB_URL` 이 같은 인스턴스/포트를 가리키는지, 사용자/비밀번호가 일치하는지 확인.
- **사내 SSL inspection 환경**: docker mirror 가 아니라 release binary mirror 또는 소스 빌드 경로를 사용한다 (정책 메모 참조).
