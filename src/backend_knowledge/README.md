# DevHub v0.2.0 — backend-knowledge

> **v0.2.0 PoC = standalone 도구** (umbrella doc §1.2 G7). 다른 backend (`backend-core`) 와의 연결 ❌, OIDC ❌, 외부 시스템 7종 source 만 단방향 (`backend-knowledge` 가 외부 시스템 API 호출, `backend-core` 의 어떤 layer 도 호출 ❌). Path Y caller-provided user context (gateway 3-step orchestration, §12.3) — backend-knowledge 는 auth 자체 안 함, `X-DevHub-User-Context` header 로 user/org/project/roles 전달 시 filter / curation ownership check 만 수행.

## 1. 빠른 시작 (Quick Start)

### 1.1 의존성 설치

```bash
cd backend-knowledge
python3.13 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

### 1.2 환경 변수 설정

```bash
cp .env.example .env
# .env 수정: GITEA_URL, GITEA_TOKEN 등 (없어도 mock mode 동작)
```

### 1.3 실행

```bash
# 개발 모드 (auto-reload)
uvicorn backend_knowledge.main:app --reload --port 8000

# 또는
python -m backend_knowledge.main
```

### 1.4 검증

```bash
# Health check
curl http://localhost:8000/health

# Path Y 사용자 context 로 query (X-DevHub-User-Context header)
curl -H "X-DevHub-User-Context: <base64url(json)>" http://localhost:8000/api/v0-2/health/protected
```

## 2. 아키텍처 (Architecture)

### 2.1 디렉터리 구조 (umbrella doc §2.1 정합)

```
backend-knowledge/
├── okf/                       # OKF concept format library
│   ├── concept.py             # Concept reader/writer (Markdown + frontmatter)
│   ├── frontmatter.py         # Pydantic v2 model (7+5=12 field)
│   └── cross_link.py          # Cross-link extractor (markdown-link + wiki-link)
├── sources/                   # Source plugin (외부 시스템 5종)
│   ├── _base.py               # ABC: connect() / fetch() / normalize()
│   ├── registry.py            # 5 source 등록 + dispatch
│   ├── gitea_repo_pull.py     # Gitea repo pull
│   ├── gitea_issue.py         # Gitea issue
│   ├── gitea_wiki.py          # Gitea wiki
│   ├── gitea_action.py        # Gitea Actions / CI
│   └── homelab_mock.py        # Homelab mock (PoC, M-v0.2.0+ 확장)
├── auth/                      # 인증 (Path Y 만, 자체 인증 ❌)
│   └── path_y.py              # X-DevHub-User-Context header 검증
├── storage/                   # 저장소
│   └── raw_store.py           # 봉투 암호화 (AES-256-GCM) + file mode (var/raw/)
├── api/                       # FastAPI routers
│   └── ingest.py              # 6 Ingest endpoint (PR 1)
├── tests/                     # pytest
├── config.py                  # pydantic-settings (env var)
├── logger.py                  # structlog JSON log
├── main.py                    # FastAPI app entry
├── var/                       # Runtime data (gitignore)
│   ├── raw/                   # 봉투 암호화 raw file mode
│   ├── bundles/               # OKF bundle Markdown
│   ├── audit/                 # Audit log JSON Lines
│   └── log/                   # Application log
├── pyproject.toml             # Python project metadata
├── requirements.txt           # Pinned dependencies
└── .env.example               # Environment variable template
```

### 2.2 Layer 격리 (umbrella doc §2.3 정합)

| Layer | Import | 다른 layer 호출 |
| --- | --- | --- |
| **API** (FastAPI routers) | storage, sources, auth, okf, config, logger | ❌ (router 끼리 호출 ❌) |
| **Sources** | okf, config, logger, storage (read), storage.raw_store (write) | ❌ (source 끼리 호출 ❌) |
| **Storage** | config, logger, okf (frontmatter), cryptography | ❌ (raw_store 단독) |
| **Auth (Path Y)** | config, logger | ❌ (path_y 단독) |
| **OKF** | config, logger, pydantic, markdown-it-py | ❌ (독립 library) |
| **Config / Logger** | (utility) | (utility 끼리 호출만) |

## 3. 5 source plugin (per umbrella doc §3.8.1)

### 3.1 Source plugin ABC

```python
from abc import ABC, abstractmethod

class SourcePlugin(ABC):
    name: str  # "gitea_issue" | "gitea_wiki" | ...

    @abstractmethod
    async def connect(self, credential: dict) -> None: ...

    @abstractmethod
    async def fetch(self, since: datetime | None) -> list[dict]: ...

    @abstractmethod
    async def normalize(self, raw: dict) -> "Concept": ...
```

### 3.2 5 source (M-v0.2.0 PoC)

| # | Source | Gitea API | Mock mode | 비고 |
| --- | --- | --- | --- | --- |
| 1 | `gitea_repo_pull` | `GET /api/v1/repos/{owner}/{repo}/pulls` | ✅ | homelab_mock 으로 PoC 운영 가능 |
| 2 | `gitea_issue` | `GET /api/v1/repos/{owner}/{repo}/issues` | ✅ | 동일 |
| 3 | `gitea_wiki` | `GET /api/v1/repos/{owner}/{repo}/wiki/pages` | ✅ | 동일 |
| 4 | `gitea_action` | `GET /api/v1/repos/{owner}/{repo}/actions/runs` | ✅ | 동일 |
| 5 | `homelab_mock` | (없음, in-memory mock) | (default) | M-v0.2.0+ PoC default |

### 3.3 Mock mode (PoC default)

- `GITEA_URL` 또는 `GITEA_TOKEN` 미설정 시 → 자동으로 mock data 반환
- `homelab_mock` 은 항상 in-memory mock (외부 호출 ❌)
- Real Gitea 연동 시 `.env` 에 `GITEA_URL=http://localhost:3000` + `GITEA_TOKEN=<token>` 설정

## 4. Path Y caller-provided user context (umbrella doc §3.6.1 정합)

### 4.1 8 field schema

```json
{
  "version": "v0",
  "user_id": "u_abc123",
  "org_id": "ou_root_dept_a",
  "org_unit_ids": ["ou_root_dept_a", "ou_dept_b1"],
  "project_ids": ["prj_x"],
  "roles": ["developer", "project_leader:prj_x"],
  "request_id": "req_20260618_xxx",
  "issued_at": "2026-06-18T10:00:00+09:00"
}
```

### 4.2 X-DevHub-User-Context header

`base64url(json)` 으로 인코딩. 만료 5분 (`issued_at + PATH_Y_MAX_AGE_SECONDS = 300`).

## 5. 봉투 암호화 (umbrella doc §3.6 / ADR-0025 정합)

**scope = raw + .env/KEK 만**. bundle / concept (.md) 는 git-pushable + wiki review flow 위해 봉투 암호화 ❌ (per §3 NFR NFR-S-001 AC5).

- DEK per raw (per-message random)
- KEK 1Password / HashiCorp Vault 보관 (별도 secure storage)
- AES-256-GCM, nonce 96 bit, auth tag 128 bit

## 6. 환경 변수 (config.py)

| 변수 | default | 설명 |
| --- | --- | --- |
| `PATH_Y_MAX_AGE_SECONDS` | `300` | Path Y header 만료 (5분) |
| `VAR_DIR` | `./var` | Runtime data dir |
| `LOG_LEVEL` | `INFO` | structlog log level |
| `LOG_FORMAT` | `json` | `json` | `console` |
| `GITEA_URL` | (없음) | Gitea instance URL (mock mode 면 미설정) |
| `GITEA_TOKEN` | (없음) | Gitea access token (mock mode 면 미설정) |
| `GITEA_DEFAULT_OWNER` | `devhub` | Gitea repo owner (default) |
| `GITEA_DEFAULT_REPO` | `example` | Gitea repo name (default) |
| `RAW_ENCRYPTION_KEY` | (없음) | 봉투 암호화 KEK base64 (없으면 plaintext mode) |
| `AUDIT_LOG_RETENTION_DAYS` | `7` | Audit log retention |
| `ENABLE_METRICS` | `false` | `/metrics` endpoint 활성화 (M-v0.2.0+ Prometheus) |

## 7. M-v0.2.0 PoC 운영 검증

```bash
# Health check
curl http://localhost:8000/health
# {"status": "ok", "version": "0.2.0", "timestamp": "..."}

# Path Y 검증 (X-DevHub-User-Context header)
curl -H "X-DevHub-User-Context: <base64url(json)>" http://localhost:8000/api/v0-2/health/protected
# {"status": "ok", "user_id": "u_abc123", "roles": ["developer"]}

# Source sync trigger
curl -X POST -H "X-DevHub-User-Context: <base64url(json)>" \
  http://localhost:8000/api/v0-2/ingest/homelab_mock/sync
# {"synced": 3, "failed": 0, "raw_ids": ["abc1234", "def5678", "ghi9012"]}
```

## 8. 후속 PR (M-v0.2.0+ 구현 sprint)

- **PR 2** (v0.2.0 PoC 운영 검증 후): Curate 5 + Query 5 + Graph 4 + viz.html viewer
- **PR 3**: Audit + Monitoring + Operational + E2E smoke test

## 9. 관련 문서 (umbrella doc / ADR)

- umbrella doc: `docs/planning/release_v0-2_roadmap.md` §2 + §3 + §3.6 + §3.8 + §10 + §11
- ADR-0037: `docs/adr/0037-okf-adoption.md` (OKF v0.1 채택)
- ADR-0038: `docs/adr/0038-backend-knowledge-creation.md` (backend-knowledge 신설)
- §1 도메인 분류: `docs/requirements/v0.2.0-domain-classification.md`
- §2 FR: `docs/requirements/v0.2.0-functional-requirements.md`
- §3 NFR: `docs/requirements/v0.2.0-non-functional-requirements.md`
- §4 UC: `docs/requirements/v0.2.0-usecases.md`
- §5 TM: `docs/requirements/v0.2.0-traceability-matrix.md`
