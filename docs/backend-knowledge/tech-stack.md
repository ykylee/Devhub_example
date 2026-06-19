# backend-knowledge v0.2.0+ — Tech Stack

- 문서 목적: v0.2.0 PoC `backend-knowledge` standalone backend 의 확정한 runtime + test 기술 스택 + 의존성 버전 + 선택 근거.
- 범위: Python 3.13+ / FastAPI 0.115.6 / Pydantic v2 / structlog / cryptography + 테스트 스택. 운영 배포 절차 + network 정책은 [`docs/setup/environment-setup.md`](../setup/environment-setup.md) + `architecture.md` §6 (Operation) 정합.
- 대상 독자: M-v0.2.0+ sprint 진입자, 후속 contributor, PR reviewer, 운영자 onboarding.
- 상태: **accepted** (2026-06-19, M-v0.2.0 PoC release post-impl retrospective, retro-design recovery)
- 최종 수정일: 2026-06-19
- 관련 문서: [`architecture.md`](./architecture.md) (main design) / [`docs/planning/release_v0-2_roadmap.md`](../planning/release_v0-2_roadmap.md) §3.3 / [ADR-0035 backend-knowledge 신설](../adr/0035-backend-knowledge-creation.md)

## 1. Runtime 의존성

| 의존성 | 버전 | 역할 | 선택 근거 |
|---|---|---|---|
| **Python** | 3.13+ | 언어 | umbrella doc §3.3 정합 (2026-06-17 결정) |
| **FastAPI** | 0.115.6 | ASGI web framework | type hint 기반 dependency injection + Pydantic v2 통합 + OpenAPI 자동 생성 (umbrella §3.1 API 매트릭스) |
| **Pydantic** | 2.9.2 | data validation + serialization | OKF v0.1 frontmatter 12 field + API request/response model + §3.6.2 governance field |
| **pydantic-settings** | 2.x | env var → Settings | `config.py` 의 11 env var (PATH_Y_MAX_AGE_SECONDS / VAR_DIR / GITEA_URL / RAW_ENCRYPTION_KEY / ENABLE_METRICS 등) |
| **uvicorn** | 0.32.1 | ASGI server | `python -m backend_knowledge.main` entry point |
| **structlog** | 24.4.0 | structured JSON logging | FastAPI middleware + 5 metric 기반 + §11.3 monitoring 정합 |
| **cryptography** | 43.0.0 | AES-256-GCM 봉투 암호화 v2 | ADR-0025 + Codex P2 review 4 fix (per-raw DEK + KEK wrap) |
| **httpx** | (transitive, 0.x) | async HTTP client | 4 Gitea sub-plugin 의 `httpx.AsyncClient` + Bearer token auth + 자동 mock mode (`GITEA_URL`/`GITEA_TOKEN` 미설정 시 in-memory) |
| **PyYAML** | 6.x | YAML parser | OKF frontmatter (12 field, YAML) ↔ Pydantic model 변환 |
| **python-multipart** | 0.x | multipart parser | FastAPI form data (PoC 미사용, dependency 명시) |

**명시적 NOT 사용** (의도적 제외):
- ❌ **DB driver (psycopg / asyncpg / sqlite3)** — PoC = file mode only. M-v0.2.3+ production 시 PostgreSQL option (`§10.1`).
- ❌ **ORM (SQLAlchemy / Tortoise)** — M-v0.2.3+ 에서 평가.
- ❌ **External config (Dynaconf / Hydra)** — pydantic-settings 단일. M-v0.2.3+ 까지 유지.
- ❌ **OpenAI SDK** — Pi (pi.dev) SDK 미설치 (M-v0.2.3+ scope, `§3.5.7`).
- ❌ **Celery / Dramatiq** — PoC = inline async + manual cron. M-v0.2.3+ production 시 cron worker 검토.

## 2. Test 의존성

| 의존성 | 버전 | 역할 |
|---|---|---|
| **pytest** | 8.3.3 | test runner (166 UT + 11 E2E step = 177 test) |
| **pytest-asyncio** | 0.x | async test (httpx async + FastAPI endpoint async) |
| **anyio** | 4.x | httpx async 백엔드 + FastAPI TestClient async 지원 |
| **httpx** | (transitive) | Gitea mock mode test 의 httpx MockTransport |
| **FastAPI TestClient** | (built-in) | sync API endpoint test |

**Test marker 정책**:
- `unit` — 기본값, `tests/test_*.py` 10 file
- `e2e` — `tests/e2e/test_*.py` 1 file (FastAPI TestClient 기반 happy path)
- `integration` — PoC = 사용 안 함. M-v0.2.3+ production 시 real Gitea instance + DB 통합 test.

## 3. 선택 근거 (Decision rationale)

### 3.1 Python 3.13+
- umbrella doc §3.3 정합 (2026-06-17 결정)
- 신기능 활용: `dict[str, Any]` PEP 585 generic + improved error message

### 3.2 FastAPI 0.115.6
- **선택 이유**: type hint 기반 dependency injection (Path Y 8 field validation) + Pydantic v2 native 통합 + OpenAPI 자동 생성 (외부 developer onboarding)
- **대안 평가**:
  - Flask: dependency injection 약함 + Pydantic v2 통합 manual
  - Django REST: ORM 강결합 + DRF overhead
  - Litestar / Starlette 직접: ecosystem 부족
- **트레이드오프**: FastAPI 의 `on_event` deprecation (PR 1~3 에서 4 warning) → M-v0.2.3+ 에서 lifespan event 로 전환

### 3.3 Pydantic v2
- **선택 이유**: OKF v0.1 frontmatter 의 12 field + type 검증 + JSON Lines audit log + Path Y 8 field schema
- **대안**: dataclasses + manual validation — boilerplate 폭증
- **v1 → v2 migration**: M-v0.2.0 부터 v2 만 사용 (umbrella doc §3.3 정합)

### 3.4 cryptography 43.0.0
- **선택 이유**: AES-256-GCM 봉투 암호화 v2 (per-raw DEK + KEK wrap) — ADR-0025 + Codex P2 review fix
- **대안**: PyNaCl (libsodium) — API surface 큼 + 의존성 무거움
- **트레이드오프**: cryptography 43 의 빌드 도구 (libssl-dev) 필요. M-v0.2.3+ production 시 Docker image multi-stage build 정합

### 3.5 structlog 24.4.0
- **선택 이유**: JSON Lines audit log + FastAPI middleware 통합 + 5 metric 기반 (request count / latency / error count)
- **대안**: stdlib logging + custom JSON formatter — boilerplate

## 4. 호환성 매트릭스 (M-v0.2.0+ 진화)

| 시점 | 추가 의존성 | 제거 의존성 | 비고 |
|---|---|---|---|
| **M-v0.2.0 (PoC, 현재)** | 위 §1 표 그대로 | (없음) | umbrella doc §3.3 1:1 정합 |
| **M-v0.2.1** | + frontend 의 `requests` (homelab real wire) | (없음) | 5 source plugin 정식 wire |
| **M-v0.2.2** | + `apscheduler` 또는 `celery` (cron worker) | (없음) | §10.3 Pi periodic ingest |
| **M-v0.2.3** | + `asyncpg` 또는 `psycopg[binary]` (PostgreSQL option, §10.1) | (없음) | DB-based raw mode |
| **M-v0.2.3+** | + `pi-coding-agent` SDK (pi.dev) | (없음) | §3.5.7 Pi LLM cross-link resolution |
| **M-v0.3.0+** | + 임베딩 vendor (OpenAI / Anthropic / Cohere) | (없음) | §3.5.7 Pi LLM retrieval + reranking |

## 5. 검증 (Verification)

```bash
# 의존성 정확 설치 확인
pip install -r src/backend_knowledge/requirements.txt --dry-run

# 버전 일치 확인 (umbrella §3.3 + 본 §1 표)
pip show fastapi pydantic uvicorn structlog cryptography | grep -E "Name|Version"

# import 검증
cd src && python -c "from backend_knowledge.main import app; print(f'{len(app.routes)} routes')"
# 예상: 30+ routes
```

## 6. 변경 이력

| 일자 | 변경 |
|---|---|
| 2026-06-19 | 1차 작성 (M-v0.2.0 PoC release, retro-design recovery, PR #657 follow-up). umbrella doc §3.3 1:1 정합 + 선택 근거 3.1-3.5 + 호환성 매트릭스 5 row. |
