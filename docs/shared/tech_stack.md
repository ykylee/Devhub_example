# DevHub 기술 스택 및 환경 설정 가이드

- 문서 목적: DevHub 의 확정 기술 스택 (Frontend / Go Core / Python AI / IdP / DB) 과 환경 설정 기준을 정의한다.
- 범위: 런타임 버전, 핵심 라이브러리 선택, 통신 프로토콜, DB 선택. 운영 절차는 `docs/setup/test-server-deployment.md`, 아키텍처 결정은 `docs/architecture.md` 가 source-of-truth.
- 대상 독자: 모든 개발자, DevOps, AI agent, 신규 환경 부트스트랩 담당.
- 상태: accepted
- 작성일: 2026-04-29
- 최종 수정일: 2026-05-21 (Go 1.25.9 기준으로 정합, sprint `codex/work_260521-c-db-docker-option`)
- 관련 문서: [아키텍처 설계서](./architecture.md), [요구사항 정의서](./requirements.md), [환경 구성 가이드](./setup/environment-setup.md), [테스트 서버 배포 가이드](./setup/test-server-deployment.md), [ADR-0019 Keycloak 단일화 (현재 IdP)](./adr/0019-keycloak-only-idp.md), [ADR-0001 IdP (Hydra+Kratos, superseded)](./adr/0001-idp-selection.md), [ADR-0003 No-Docker CI scope](./adr/0003-no-docker-policy-ci-scope.md), [ADR-0018 Single-Port Reverse Proxy](./adr/0018-single-port-reverse-proxy-policy.md).

## 1. 확정 기술 스택 (Technology Stack)

DevHub은 Gitea 연동, AI 분석, 실시간 대시보드 제공을 위해 다음과 같은 하이브리드 스택을 사용합니다.

### 1.1 Backend
- **Core Service (Main):** **Go (Gin)**
    - 역할: Gitea API/Webhook 연동, 시스템 제어 로직, 권한 관리.
    - 특징: 고성능 비동기 처리, Gitea와의 언어적 정합성.
- **AI/Analysis Module:** **Python (FastAPI)**
    - 역할: 빌드 로그 분석(AI 가드너), 리스크 탐지 모델 구동.
    - 특징: 풍부한 AI/데이터 분석 생태계 활용. 현재 스캐폴딩은 FastAPI health endpoint와 gRPC 의존성 준비까지 완료되었으며, 실제 gRPC 서버 구현은 후속 작업 범위입니다.
- **Internal Communication:** **gRPC** (Go ↔ Python, 확정 기본 계약)
    - REST/HTTP는 외부 API 및 상태 확인용 endpoint에 사용하며, Go Core와 Python AI 간 분석 요청/응답의 기본 계약은 gRPC로 둡니다.

### 1.1.b 신규 백엔드 backend-knowledge (v0.2.0+, ADR-0035)

> **독립 standalone backend** — 다른 backend (backend-core / 다른 시스템) 의 어떤 layer 든 import / 호출 / 공유 ❌. 외부 시스템 5종 (Gitea 4 sub-plugin + homelab_mock, M-v0.2.0 PoC) / 7종 (M-v0.2.3 운영 기준) source 만 단방향. **M-v0.2.0 PoC release (PR #654/#655/#656/#657 MERGED, 30 endpoint, 166 UT, 11 E2E step, tag v0.2.0)**.

자세한 design: [`docs/backend-knowledge/architecture.md`](../backend-knowledge/architecture.md) + [`docs/backend-knowledge/tech-stack.md`](../backend-knowledge/tech-stack.md) + [ADR-0035 backend-knowledge 신설](../adr/0035-backend-knowledge-creation.md).

- **언어 / runtime**: Python 3.13+ / uvicorn 0.32.1
- **web framework**: FastAPI 0.115.6 (Pydantic v2 native 통합 + OpenAPI 자동)
- **validation**: Pydantic 2.9.2 (OKF v0.1 frontmatter 12 field + Path Y 8 field schema)
- **config**: pydantic-settings 2.x (11 env var)
- **logging**: structlog 24.4.0 (JSON Lines audit log + FastAPI middleware)
- **cryptography**: cryptography 43.0.0 (AES-256-GCM 봉투 암호화 v2 = per-raw DEK + KEK wrap, ADR-0025)
- **http client**: httpx (transitive, 4 Gitea sub-plugin 의 async + Bearer token + 자동 mock mode)
- **yaml**: PyYAML 6.x (OKF frontmatter)
- **test**: pytest 8.3.3 / pytest-asyncio / anyio / FastAPI TestClient (166 UT + 11 E2E step)
- **명시적 NOT 사용** (의도적): DB driver / ORM / Celery / OpenAI SDK / external config (Dynaconf/Hydra) — M-v0.2.3+ production 시 PostgreSQL option + Pi (pi.dev) SDK 추가

**layer 격리** (architecture.md §3): 7 module group (API / Sources / Storage / OKF / Auth / Audit / Monitoring) + 2 utility (Config / Logger). **API cross-router call ❌** + **Source cross-source ❌** 정공법.

**Tier**: 사외 (단일 env, mock + standalone + GitHub main push, 사용자 2026-06-19 결정). 사내 한정 정보 0 row (DEVHUB_KEYCLOAK_* / GITEA_URL / 172.16.0.0/12 pattern 0).

### 1.2 Frontend
- **Framework:** **Next.js 16 (React 19.2, App Router)**
    - 역할: 역할별 기본 진입 우선순위 대시보드, 실시간 상태 시각화.
    - 현재 버전: `next@16.2.x`, `react@19.2.x` (`frontend/package.json`). 빌드 산출물 26 routes (static + dynamic 혼합).
- **Styling:** **Tailwind CSS** + semantic theme (CSS 변수 기반 다크/라이트 모드).
- **Data Fetching:** 현재는 `lib/services/*.service.ts` + `apiClient` (커스텀 wrapper) 사용. TanStack Query 도입은 carve out 상태.
- **Interactive UI:** **React Flow** (`@xyflow/react`) — Organization hierarchy editor + Topology v2 (Environment grouping + WebSocket 실시간 갱신) 에 active 사용.
- **Auth flow:** Keycloak OIDC code flow with PKCE — `lib/auth/pkce.ts` + `lib/auth/token-store.ts` + `app/auth/login,callback,logout,signup` routes ([ADR-0019](./adr/0019-keycloak-only-idp.md)).

### 1.3 Database
- **Main DB:** **PostgreSQL (v15+)**
    - 역할: 정형 데이터 및 JSONB 기반 비정형 분석 결과 저장.

---

## 2. 개발 환경 설정 (Environment Setup)

본 저장소는 **native (no-docker) 모드를 default** 로 한다 ([ADR-0003](./adr/0003-no-docker-policy-ci-scope.md)). 컨테이너 자산 중 `docker-compose.yml` 과 로컬 오버라이드는 환경별 제약이 달라 git 추적에서 제외되지만, 각 서비스의 `Dockerfile` 은 빌드 계약이므로 git 추적 대상이다. 절차는 [`docs/setup/environment-setup.md`](./setup/environment-setup.md) 가 source-of-truth.

### 2.1 사전 요구 사항 (Prerequisites)
- **Go**: 기준 v1.25.9 (`backend-core/go.mod`). 로컬 `go` 도 같은 버전을 권장.
- **Python**: 기준 v3.12 (`backend-ai/Dockerfile`). host build (`scripts/build-artifacts.sh`) 가 `pip install --target` 으로 site-packages 를 만들어 컨테이너에 복사하므로, **host 의 `python3` 도 v3.12 권장** (메이저/마이너 mismatch 시 컴파일 확장 모듈의 ABI 충돌 가능). `make setup` 과 `make proto` 는 로컬 `python3` 를 그대로 사용한다.
- **Node.js**: v20 LTS+ (Next.js 16 빌드 요구). 로컬 `npm install` 과 `npm run` 계열 명령은 로컬 Node.js 를 사용한다.
- **PostgreSQL**: v15 (호스트 native 또는 시스템 서비스로 기동).
- **protoc** (gRPC 컴파일러)
- **Go protoc plugins:** `protoc-gen-go`, `protoc-gen-go-grpc` (`make proto-tools` 로 설치)
- **Python gRPC tools:** `grpcio`, `grpcio-tools` (`make setup` 으로 `backend-ai/requirements.txt` 설치)
- **DB migration tool:** `golang-migrate/migrate` v4.19.1 (PostgreSQL driver 포함, `make migrate-tools` 로 설치)
- **Keycloak (인증)**: v26+ — 로컬에서는 native binary 또는 별도 docker 자산으로 기동. `docs/setup/keycloak_operations.md` 참고.
- **Docker / Docker Compose (optional)**: 사용자 환경에서 컨테이너 모드를 쓸 때만 필요. git 추적 외 자산으로 사내 위키 등에서 환경별로 관리.

### 2.2 프로젝트 초기화 (Initialization)
루트 디렉토리에서 제공된 `Makefile`을 사용하여 의존성을 설치합니다.

```bash
# 각 프로젝트 의존성, proto 생성 도구, migration 도구 설치 후 proto 파일 컴파일
make init
```

단계별로 실행할 경우 다음 순서를 사용합니다.

```bash
make setup
make proto-tools
make proto
```

### 2.3 로컬 실행 (Running Locally)

`make build` / `make run` 은 환경별 절차 안내만 출력한다. 실제 기동 절차는 [`docs/setup/environment-setup.md`](./setup/environment-setup.md) 의 두 갈래 가이드를 따른다.

```bash
# Native (default):
(cd backend-core && go run .)        # :8080
python backend-ai/main.py            # :8000   (또는 uvicorn main:app)
(cd frontend && npm run dev)         # :3000

# Docker (optional, 사용자 로컬 자산):
docker-compose up -d                 # docker-compose.yml 은 환경별로 git 외부 보관
```

### 2.4 주요 서비스 포트 정보
- **Frontend:** `http://localhost:3000`
- **Backend Core:** `http://localhost:8080`
- **Backend AI:** `http://localhost:8000`
- **Backend AI gRPC:** `localhost:50051` (예약 노출, 현재 서버 구현 전)
- **PostgreSQL:** `localhost:5432`

### 2.5 데이터베이스 마이그레이션

PostgreSQL 스키마 변경은 **golang-migrate/migrate**를 사용합니다. Go Core가 Gitea Webhook 원본 이벤트, 정규화 테이블, command/audit lifecycle, 권한/프로젝트 매핑을 관리하므로 migration 파일은 `backend-core/migrations/`에 둡니다.

```bash
# migrate CLI 설치(PostgreSQL driver 포함)
make migrate-tools

# 새 migration 파일 생성
make migrate-create NAME=create_webhook_events

# 로컬 DB에 migration 적용
make migrate-up

# 현재 적용 버전 확인
make migrate-version
```

기본 로컬 접속 문자열은 `postgres://user:pass@localhost:5432/devhub?sslmode=disable`입니다. 다른 환경에서는 `MIGRATE_DB_URL`로 override합니다.

---

## 3. 개발 가이드라인

### 3.1 Gitea 연동 설정
`.env` 파일(또는 환경 변수)에 Gitea 서버 정보를 설정해야 합니다.
```env
GITEA_URL=http://your-gitea-server.com
GITEA_TOKEN=your-access-token
GITEA_WEBHOOK_SECRET=your-webhook-secret
BACKEND_AI_URL=http://localhost:8000
```

### 3.2 아키텍처 원칙
- 모든 Gitea 이벤트는 **Go Core**를 통해 먼저 수신되며, 필요한 경우에만 **Python AI**로 전달됩니다.
- 프론트엔드와 백엔드 간의 실시간 연동은 **WebSocket**을 우선적으로 사용합니다.
- 데이터 보존 정책(1개월)을 준수하기 위해 DB 파티셔닝 또는 스케줄링된 삭제 로직을 사용합니다.
