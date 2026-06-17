# DevHub v0.2.0 릴리즈 로드맵 — 외부 연동 + AI Agent Library (OKF) 통합 백엔드

- 문서 목적: DevHub v0.2.0 의 **단일 source-of-truth umbrella 로드맵**. 외부 시스템 연동 + 데이터 취합을 별도 백엔드로 모으고, 이를 **Google Open Knowledge Format (OKF) 기반 AI Agent Library** 로 발전시키는 컨셉 + 3가지 기본 기능 + 1차 raw 데이터의 API 정책 + 마일스톤 + 기존 `external-integrations-agentic-rag-roadmap.md` child 문서로의 진입 경로.
- 범위: v0.2.0 의 (1) 외부 시스템 연동 분리 (기존 `backend-ai/` 폐기 흡수) (2) OKF 형 knowledge bundle 생성/관리 (3) AI agent + 사용자 query 응답. 1차 외부 연동 (Gitea, HomeLab) + OKF reference PoC + 핵심 3 endpoint.
- 대상 독자: 프로젝트 리드, 모든 contributor (사람 + AI agent), 후속 sprint 작업자, owner.
- 상태: accepted (2026-06-17 publish, 2026-06-18 cross-section 정합 fix 추가, §9 변경 이력 + ADR-0034/0035 publish 완료 + Q&A 11/11 결정 완료)
- 최종 수정일: 2026-06-18 (cross-section 정합 fix — A/A 결정 (Gitea 4 sub-plugin + x_devhub_category + 5 카테고리) 의 9 위치 cross-section 정합 일괄 반영, §2.1 sources/var/bundles 트리 + §3.2/§3.3 예시 + §6.1 mock count + §1.2 G3/G7 + §2.3 + §4.2/§4.3 + §5.2 P2 + §6.2 + §7 Q9).
- 결정 근거: 사용자 2026-06-17 결정 + 사용자 2026-06-10 결정 (외부 연동 = agentic RAG 와 발전) + Google Cloud `Open Knowledge Format v0.1` (2026-06-12 발표, Apache 2.0).
- 관련 문서:
  - [v0.1.0 릴리즈 로드맵](./release_v0-1_roadmap.md) (직전)
  - [외부 연동 + Agentic RAG 통합 child roadmap](./external-integrations-agentic-rag-roadmap.md) (외부 연동 분리 detail — 본 umbrella 의 §3/§4 가 가리킴)
  - [외부 시스템 연동 컨셉](./external_system_integration_concept.md)
  - [외부 연동 capability matrix](./external_integration_capability_matrix.md)
  - [ADR-0019 Keycloak 단일화](../adr/0019-keycloak-only-idp.md) (Keycloak = 사내 IdP, v0.2.0 scope 제외 근거)
  - [ADR-0025 봉투 암호화](../adr/0025-envelope-encryption-key-management.md)
  - [Code Taxonomy](../governance/code-taxonomy.md) (3 레이어 + 4 계층, 본 v0.2.0 의 신규 백엔드 위치 결정 근거)
  - [Backend API 공통 규약](../api/conventions.md) (envelope + enum, 본 v0.2.0 API 정합)
  - Google OKF reference: [`GoogleCloudPlatform/knowledge-catalog/okf/SPEC.md`](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md), [`README.md`](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/README.md)

## 0. 사용 가이드

본 문서는 **v0.2.0 의 umbrella**. 신규 sprint 진입 시 §3 (3가지 기본 기능) + §4 (마일스톤) + §5 (1차 raw API 정책) + §6 (독립 → 연동 단계) 확인 후 진입.

1. **신규 sprint 진입 전** §3 / §4 / §5 확인
2. **외부 연동 분리 detail** → [external-integrations-agentic-rag-roadmap.md](./external-integrations-agentic-rag-roadmap.md) 의 §3 (Phase 1 adapter pattern) / §4 (Phase 2 agentic RAG)
3. **OKF 스펙 정확값** → Google `SPEC.md` 직접 참조 (본 문서는 요약 + 본 프로젝트 적용)
4. **결정 변경** 발생 시 §8 변경 이력에 row 추가

## 1. v0.2.0 컨셉 — Why, What, How

### 1.1 Why v0.2.0 (문제 인식)

**현재 v0.1.x 의 한계** 3가지 (각 한계는 concrete DevHub 시나리오로 짚음):

1. **외부 시스템 연동 분산**: `backend-core/internal/infrastructure/` (gitea, ci, commandworker, hrdb, serviceaction) + `backend-core/internal/integrations/adapters/` (homelab, task_item_puller, metrics) + `backend-ai/` (placeholder, 사실상 미사용) — 동일 카테고리 ("외부 시스템 연동") 가 3 디렉터리에 분산. **시나리오**: Gitea sync 가 느려졌을 때 `infrastructure/gitea/` 의 어느 file 을 봐야 할지 1눈에 안 보임. homelab / gitea / hrdb / task_item_puller / metrics 5종 adapter 가 각각 다른 위치에 흩어져 있어 운영자 onboarding 비용이 크고, 신규 adapter 추가 시 어느 layer 에 둘지 매번 결정 필요.
2. **AI agent / LLM context 부재**: 운영자가 "최근 Gitea sync 가 느려졌나?" 같은 자연어 쿼리 불가능. integration 의 운영 history + monitoring + spec 이 context 로 retrieve 되지 않음. **시나리오**: 사내 incident 발생 시 "어제 homelab pull 이 몇 건 실패했어?" / "gitea webhook 이 마지막으로 변경된 게 언제야?" 같은 운영 query 를 LLM agent 에 던져서 metric + runbook + recent event 를 한 번에 retrieve 할 방법이 없음 — 사람이 여러 dashboard + Slack + wiki 를 직접 cross-reference.
3. **`backend-ai/` 의 dead state**: `main.py` 1 endpoint (`/health`) + TODO 2개 ("gRPC Server for AnalysisService" / "AI Logic for Log Analysis") + FastAPI + grpcio skeleton 만. v0.1.0 roadmap §1.2 "제외 기능 (v2 P3)" 분류 — AI Gardener gRPC + Suggestion Feed 미구현. **시나리오**: v0.1.0 release 시점부터 placeholder 상태로 6+ release 누적, "어차피 안 쓰이는 코드" 가 되어감. 이 dead state 를 정식으로 처리할 시점.

### 1.2 What v0.2.0 (목표)

> **외부 시스템 연동 + 데이터 취합을 별도의 백엔드(`backend-knowledge`)로 모으고, Google OKF 형 AI Agent Library 로 통합.**

| # | 목표 | 산출물 |
| --- | --- | --- |
| G1 | 외부 시스템 연동을 `backend-knowledge` 단일 백엔드로 모으기 | 신규 `backend-knowledge/` (Python + FastAPI). bundle 단위로 외부 시스템 데이터 취합 — 예: `devhub-homelab/`, `devhub-gitea/` (Gitea 4 sub-plugin 통합 단일 bundle), `devhub-metrics/`, `devhub-hrdb/` 등 per-source bundle 디렉터리 (M-v0.2.3 운영 기준 7종 source = 4 bundle, §1.2 G3 / §2.1 / §6.4 정합) |
| G2 | 기존 `backend-ai/` **폐기** (placeholder 상태, 흡수할 코드 0) | `backend-ai/` 디렉터리 제거, Makefile / docker-compose / docs 의 backend-ai reference 일괄 정리 (M-v0.2.2). 실제 production wiring 없음 → 이전 코드 0 |
| G3 | **외부 연동 책임의 단일화** — backend-core 의 `integrations/adapters/` + `infrastructure/` 의 외부 연동 코드를 새 백엔드로 이전 (**외부 시스템 API 만 참조, backend-core 코드 참조 ❌**) | source plugin **7종** (Gitea 4 sub-plugin: gitea_repo_pull / gitea_issue / gitea_wiki / gitea_action + homelab + metrics + hrdb, M-v0.2.3 운영 기준) 을 **외부 시스템 공식 API spec 만 보고 0에서 Python 으로 작성**. 기존 Go adapter 의 로직 / interface / naming / repository / model 참조 ❌. backend-core 측의 Go adapter 제거는 별도 PR (backend-knowledge 책임 아님) |
| G4 | OKF 형 knowledge bundle **rule-based + LLM-optional** 생성/관리 (1차 rule-based) | rule-based enricher (1차) + bundle 디렉터리 + index.md 자동 생성 + cross-link 자동 추출. LLM enrichment agent (Pi `pi-coding-agent` SDK / RPC mode) 는 v0.2.3+ 부터 활성화 (Q3 결정) |
| G5 | 3가지 기본 기능 API 노출 (internal-only, no auth — §2.3) | Ingest / Curate / Query endpoint (envelope 정합). OIDC / Keycloak / backend-core 인증 위임 ❌ (본 시스템 scope 외) |
| G6 | 1차 raw 데이터의 API 조회/추가 (외부 시스템 raw, internal-only) | `GET/POST /api/v0-2/raw/{type}/{name}`. 1차 raw 데이터 = **외부 시스템 데이터** (homelab JSON, gitea API response, prometheus scrape 등). backend-core 의 Repository / Domain 데이터 ❌ (본 시스템 scope 외) |
| G7 | **완전 standalone 운영** — 다른 backend (backend-core / 다른 백엔드 / 다른 시스템) 와의 연결 ❌, OIDC ❌, **외부 시스템만 단방향** (2026-06-17 결정, §4 self-review 강화) | M-v0.2.0/v0.2.1: standalone (mock source + no auth + 별도 docker network) → M-v0.2.2: 외부 시스템 **6종** source wire (Gitea 4 + homelab + metrics, backend-core 와 무관) + backend-ai 폐기 (단독 결정) → M-v0.2.3: 외부 시스템 **7종** source wire (+ hrdb) + Pi (pi.dev) LLM enrich 활성화 |

### 1.3 How — Google OKF 구조 차용

**§1.1 의 3가지 한계 + 업계 표준 부재 (내부 knowledge 가 코멘트/구두/문서 산만하여 AI agent 가 참조할 표준 부재)** 를 함께 풀기 위해, **Google Open Knowledge Format (OKF) v0.1** (2026-06-12 Google Cloud 발표, Apache 2.0, [`SPEC.md`](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md) 직접 참조 권장) 를 채택. OKF 는 "Markdown + YAML frontmatter + cross-link" 표준으로, vendor-neutral + git 가능 + agent-friendly 한 지식 표현을 제공 — 본 §1.3 에서 OKF 핵심 + 본 프로젝트 적용 매핑을 정리한다.

**OKF v0.1 핵심** (위 reference 의 `SPEC.md` 가 1차 출처, 본 bullet 은 요약):

- **형식**: plain Markdown 파일 + YAML frontmatter, 디렉터리 트리
- **1 concept = 1 `.md` 파일** (테이블 / metric / API / runbook / event 등)
- **frontmatter 최소**: `type` 1개 필수. 권장: `title`, `description`, `resource`, `tags`, `timestamp` — 모두 옵션
- **cross-link**: Markdown `[title](/path/to/concept.md)` 로 concept 간 graph 형성
- **progressive disclosure**: 디렉터리별 `index.md` 자동 생성 (agent / human 이 한 level 씩 navigation)
- **vendor-neutral**: 특정 cloud / DB / agent framework / model provider 종속 ❌
- **producer**: human hand-author / agent (Google ADK, LangChain, custom) / export pipeline (Dataplex, Unity Catalog, Collibra) / DB walking script
- **consumer**: static file server / Obsidian / Notion / MkDocs / LLM context / search index / graph viewer (Cytoscape.js 기반 self-contained `viz.html`)
- **reference impl**: [`enrichment_agent`](https://github.com/GoogleCloudPlatform/knowledge-catalog/tree/main/okf) (Google ADK + Gemini + BigQuery source)

**본 프로젝트 적용** (v0.2.0):

| OKF 원리 | DevHub v0.2.0 적용 |
| --- | --- |
| 1 concept = 1 .md | `backend-knowledge/var/bundles/{bundle}/{type}/{slug}.md` (예: `metric/repo_kpi_sync_duration.md`) |
| frontmatter `type` 필수 | 본 프로젝트 `type` enum = `dataset` / `metric` / `api_endpoint` / `runbook` / `integration` / `event` / `reference` / `decision` (8종, §3.2) |
| vendor-neutral | OKF 그대로 채택 (자체 dialect 만들지 않음) |
| producer 다중 | **1차 (M-v0.2.0~v0.2.2)**: human curator + **rule-based enricher** + 외부 system adapter 1~2종 (source plugin, Q6). **후속 (M-v0.2.3)**: + LLM enrichment agent (1 vendor, 시점 결정). **장기 (M-v0.3.0+)**: + multi-vendor LLM (vendor-neutral) |
| consumer 다중 | 신규 API (Query) + frontend 위키 viewer + Obsidian 호환 + LLM context |
| progressive disclosure | `index.md` 자동 생성 (per bundle + per type) |
| graph | cross-link 자동 추출 + `viz.html` 자가 viewer (Cytoscape.js) |

**자주 묻는 차이점**:

- ❌ "OKF 그대로 백엔드" — OKF 는 **파일 시스템 표준**이지 서버 표준이 아님. 본 프로젝트는 OKF bundle 을 **저장/관리/노출하는 백엔드**를 만든다. bundle 자체의 file format 은 OKF 그대로.
- ❌ "Google ADK 종속" — OKF reference impl 이 ADK + Gemini 인 것일 뿐, 본 프로젝트는 vendor-neutral. **자체 enrich 로직 (rule-based + LLM-optional)** 으로 시작.
- ❌ "RAG 시스템 전체" — v0.2.0 의 Query API 는 **단계적 확장**. **1차 (M-v0.2.0~v0.2.2)**: 단순 retrieval (OKF concept match + raw context 표시, **LLM answer 합성 ❌**). **2차 (M-v0.2.3)**: LLM answer 합성 추가 (rule-based fallback 유지). **3차 (M-v0.3.0+)**: 풀 RAG (chunking + embedding + retrieval + reranking).

## 2. 신규 백엔드 — `backend-knowledge`

### 2.1 위치 + tier

```
backend-knowledge/                         # tier=사외 (외부 인프라 무관, OKF 형식 자체가 vendor-neutral)
├── Dockerfile
├── pyproject.toml                         # Python 3.13+ / FastAPI / Pydantic v2
├── main.py                                # FastAPI app 정의 + router wiring. uvicorn 실행은 Dockerfile/Compose 책임, dev-only `if __name__ == "__main__"` (로컬 `python main.py` 용)
├── okf/                                   # OKF spec model (Pydantic + frontmatter parser)
│   ├── spec.py                            # Concept, Frontmatter, Bundle dataclass
│   ├── frontmatter.py                     # YAML frontmatter parse/emit
│   └── link_graph.py                      # cross-link extract + reverse index
├── pi_bridge/                             # Pi (pi.dev) integration — M-v0.2.3+ 부터 활성화 (1차 미사용)
│   ├── __init__.py
│   ├── rpc_client.py                      # Pi `pi-coding-agent` **RPC mode** client (JSON over stdin/stdout, non-Node integration)
│   ├── sdk_client.py                      # Pi `pi-coding-agent` **SDK mode** client (Node subprocess via @earendil-works/pi-coding-agent npm pkg, alternative to RPC)
│   └── tools.py                           # Pi `pi-agent-core` tool definition (backend-knowledge 의 enrich tool — raw → OKF concept)
├── sources/                               # source plugin (외부 시스템 → raw concept). credential = id/pw or token (type-agnostic string), 연결 시 config 주입
│   ├── _base.py                           # SourcePlugin ABC + Credential config schema (Pydantic v2) + plugin registry
│   ├── homelab.py                         # §3 source plugin — 사내 HomeLab agent, M-v0.2.1 정식 wire (M-v0.2.0 은 homelab_mock PoC)
│   ├── homelab_mock.py                    # §3 source plugin mock — filesystem fixture 기반, M-v0.2.0 PoC
│   ├── gitea_repo_pull.py                 # Gitea 4 sub-plugin 1 — Gitea Repository (REST + git HTTP), **SCM 카테고리 (3)**, M-v0.2.0 PoC / v0.2.1 정식
│   ├── gitea_issue.py                     # Gitea 4 sub-plugin 2 — Gitea Issue API, **이슈 트래커 카테고리 (1)**, M-v0.2.0 PoC / v0.2.1 정식
│   ├── gitea_wiki.py                      # Gitea 4 sub-plugin 3 — Gitea Wiki API, **위키 카테고리 (2)**, M-v0.2.0 PoC / v0.2.1 정식
│   ├── gitea_action.py                    # Gitea 4 sub-plugin 4 — Gitea Actions API, **CI/CD 카테고리 (4)**, M-v0.2.0 PoC / v0.2.1 정식
│   ├── metrics.py                         # §3 source plugin — Prometheus scrape API, **모니터링 (5 카테고리 외)**, M-v0.2.2 운영 wire
│   └── hrdb.py                            # §3 source plugin — 사내 HR DB PostgreSQL, **HR/조직 (5 카테고리 외)**, M-v0.2.3 운영 wire
├── curate/                                # raw → OKF concept 정형화
│   ├── enricher.py                        # rule-based enrich (LLM-optional, 1차 rule-based 만)
│   ├── index_builder.py                   # bundle 별 index.md + viz.html 동시 생성
│   └── link_resolver.py                   # unresolved cross-link 추적
├── api/                                   # §3 3가지 기본 기능 + §4 1차 raw API + Bundle management (envelope 정합)
│   ├── ingest.py                          # §3.1 Ingest 3 endpoint (POST /ingest/{source}/sync, GET /ingest/{source}/status, ...)
│   ├── curate.py                          # §3.1 Curate 3 endpoint (POST /concepts/{id}/enrich, PUT /concepts/{id}, POST /bundles/{bundle}/rebuild)
│   ├── query.py                           # §3.1 Query 5 endpoint (POST /query, GET /concepts/{type}/{name}, GET /search, GET /bundles/{bundle}/index.md, GET /bundles/{bundle}/viz.html)
│   ├── raw.py                             # §4 1차 raw 데이터 (POST /raw, GET /raw/{type}/{name}, GET /raw, DELETE /raw/{id})
│   └── bundles.py                         # Bundle CRUD (GET /bundles list, POST /bundles create, GET /bundles/{name}/viz.html)
├── web/                                   # M-v0.2.1+ 부터 (frontend 관리/조회 page, §5.1 / §5.2 P1 / §6.1 정합). 별도 standalone frontend, devhub frontend 와 분리
│   ├── index.html                         # bundle list + concept list
│   ├── concept/                           # concept 조회 page (read-only, viz.html 자가 viewer 와 별도)
│   ├── admin/                             # 관리 page (raw 등록 / ingest trigger / rebuild)
│   └── ...
├── var/
│   ├── raw/                               # 1차 raw 데이터 (§4 API 정책) — 봉투 암호화 (ADR-0025), .gitignore 권장
│   │   └── {source}/{slug}.json           # 예: homelab/2026-06-17-foo.json
│   └── bundles/                           # OKF bundle 저장 (git 가능, Markdown + frontmatter)
│       ├── devhub-homelab/                # homelab source 의 OKF bundle (M-v0.2.1 정식 운영)
│       ├── devhub-gitea/                  # gitea 4 sub-plugin 의 OKF bundle (M-v0.2.0 PoC / v0.2.1 정식, Gitea 1 instance 단일 bundle)
│       ├── devhub-metrics/                # prometheus metrics 의 OKF bundle (M-v0.2.2)
│       └── devhub-hrdb/                   # hrdb 의 OKF bundle (M-v0.2.3)
└── tests/
    ├── unit/                              # OKF spec / enricher / link_graph
    └── e2e/                               # ingest → curate → query happy path
```

### 2.2 기술 스택 (1차)

| 항목 | 선택 | 근거 |
| --- | --- | --- |
| 언어 | **Python 3.13+** | OKF reference impl (Google ADK) 정합 + FastAPI 생태계 + LLM SDK 호환 |
| HTTP framework | **FastAPI** | OpenAPI 자동 생성 + Pydantic v2 + async |
| 영속성 | **file system (OKF bundle)** + **sqlite (metadata index)** | OKF 원리 ("just files, just git") 정합. RDB 는 metadata (sync 상태, link index) 만 |
| frontmatter | **PyYAML + python-frontmatter** | OKF 의 YAML frontmatter 표준 |
| markdown parse | **markdown-it-py** + custom link extractor | cross-link graph 구축 |
| LLM 호출 (optional) | **Pi (pi.dev, [github.com/earendil-works/pi](https://github.com/earendil-works/pi), MIT, v0.79.6, Mario Zechner / badlogic, 회사: Earendil Inc.)** — 메인 4 package: `pi-ai` (multi-provider LLM API, **15+ provider**: Anthropic / OpenAI / Google / Azure / Bedrock / Mistral / Groq / Cerebras / xAI / Hugging Face / Kimi For Coding / MiniMax / OpenRouter / Ollama + more, "hundreds of models") + `pi-agent-core` (agent runtime, tool calling + state mgmt) + `pi-coding-agent` (interactive coding agent CLI, **4 mode**: Interactive / Print·JSON / **RPC** / **SDK**; RPC = JSON over stdin/stdout, "non-Node integrations"; SDK = 임베드) + `pi-tui` (terminal UI, differential rendering). **"Primitives, not features"** + **"minimal agent harness"** + **"minimal system prompt"** + Steer/Follow-up 메시지 + `AGENTS.md` / `SYSTEM.md` per-project + OpenClaw (real-world integration). backend-knowledge integration 의 1차 target = **`pi-coding-agent` 의 SDK mode (Node subprocess, npm pkg) or RPC mode** | 1차 (M-v0.2.0~v0.2.2) = **rule-based 만** (Pi 의존성 ❌). 후속 (M-v0.2.3) = + `pi-coding-agent` SDK or RPC mode 로 LLM enrich 활성화 (1 vendor, 시점 결정). 장기 (M-v0.3.0+) = + `pi-ai` 의 multi-provider. Q3 결정 정합 |
| observability | **structlog + OpenTelemetry** | 표준 observability stack. tracing backend (Tempo / Jaeger / Datadog 등) 결정은 Phase 2 이후 |
| **credential 관리** | **연결 시 source plugin config 로 주입 + 봉투 암호화 저장 (ADR-0025 정합)** | 외부 시스템 credential (id/pw or token/bearer/api_key, **type-agnostic string**) 은 source plugin 의 connection config 로 **연결 시점에 주입**, 저장 시 봉투 암호화 (DEK + KMS). 평문 저장 / 로그 노출 ❌ |

### 2.3 시스템 경계 — OIDC 제외 + 외부 시스템 only (2026-06-17 결정)

**2026-06-17 결정**: **`backend-knowledge` 는 backend-core 와의 wire 안 함. OIDC / Keycloak / 기존 시스템 백엔드 참조 전부 ❌. 외부 시스템만 단방향.**

| 항목 | 정책 |
| --- | --- |
| **다른 backend 연결 (general)** | ❌ **전면 금지**. `backend-knowledge` 는 **완전 standalone 시스템** (2026-06-17 결정, §4 self-review 강화). 다른 backend (backend-core / 다른 백엔드 / 다른 시스템) 의 Go/Python 코드 / API / domain model / database / cache / envelope / repository / 어떤 layer 든 import / 호출 / 공유 ❌. **외부 시스템 7종 source 만 단방향** (Gitea 4 + homelab + metrics + hrdb, M-v0.2.3 운영 기준, §1.2 G3 정합) |
| **OIDC / Keycloak** | ❌ **OIDC 자체를 본 시스템에서 제외**. Keycloak 인증은 backend-core 의 책임 (변경 없음). `backend-knowledge` 는 bearer token / API key / session 어떤 인증 scheme 도 자체적으로 검증 안 함 |
| **외부 시스템** | ✅ **유일한 통신 대상**. Gitea 1 instance (4 sub-plugin) + homelab + metrics + hrdb 등 source plugin **7종** (M-v0.2.3 운영 기준) 의 외부 시스템 API 만 호출. 단방향 pull (외부 → backend-knowledge) |
| **API 인증** | **internal-only, no auth**. `/api/v0-2/*` endpoint 는 인증 없이 호출 가능. 별도 gateway / firewall / IP allowlist 가 보호 (Phase 1~3 의 운영 책임). 운영자 또는 별도 agent 가 호출 (**backend-core 의 어떤 layer 든 호출 ❌**, 2026-06-17 결정, §1.2 G7 / §7 Q9 정합) |
| **Keycloak 분류 (재확인)** | [external-integrations-agentic-rag-roadmap.md §0.4](./external-integrations-agentic-rag-roadmap.md) 정합 — Keycloak 은 사내 IdP, 외부 시스템 아님, backend-core 의 `domain/auth-session/` 책임. 본 시스템 scope 외 |
| **Phase 2 의 의미 변경** | ~~backend-core 와 wire~~ → **외부 시스템 6종 source plugin wire (M-v0.2.2: Gitea 4 + homelab + metrics, backend-core 와 무관)** + 7종 (M-v0.2.3: + hrdb). Phase 2 자체는 backend-core 와 분리되어 진행 |

## 3. 3가지 기본 기능 (API)

### 3.1 기능 ↔ API 매트릭스

| 기능 | API | method + path | 응답 |
| --- | --- | --- | --- |
| **(1) 수집 Ingest** | 외부 시스템 → 1차 raw concept **(M-v0.2.3+ 부터 OKF enrich 동시 가능)** | `POST /api/v0-2/ingest/{source}/sync` | `{envelope, data: {synced: N, failed: M, raw_ids: [...]}}` |
| | sync 상태 조회 | `GET /api/v0-2/ingest/{source}/status` | `{envelope, data: {last_sync, next_sync, state}}` |
| | 1차 raw concept 등록 (manual) | `POST /api/v0-2/raw` | `{envelope, data: {raw_id}}` |
| | 1차 raw concept 조회 | `GET /api/v0-2/raw/{type}/{name}` | `{envelope, data: {frontmatter, body, raw_refs}}` |
| | 1차 raw concept list (filter: source/since) | `GET /api/v0-2/raw?source=...&since=...` | `{envelope, data: {items: [...]}}` |
| | 1차 raw concept 삭제 | `DELETE /api/v0-2/raw/{id}` | `{envelope, data: {deleted: true}}` |
| **(2) 정리 Curate** | raw → OKF concept 변환 trigger **(M-v0.2.3+ 부터 Pi LLM enrich 활성화, 1차 M-v0.2.0~v0.2.2 = rule-based 만)** | `POST /api/v0-2/concepts/{id}/enrich` | `{envelope, data: {concept_id, version}}` |
| | concept manual edit | `PUT /api/v0-2/concepts/{id}` | `{envelope, data: {concept_id, version}}` |
| | bundle list | `GET /api/v0-2/bundles` | `{envelope, data: {items: [{name, concepts, links, last_rebuild}]}}` |
| | bundle 생성 | `POST /api/v0-2/bundles` | `{envelope, data: {name}}` |
| | bundle index + cross-link 재생성 **(M-v0.2.3+ 부터 LLM cross-link 자동 resolution, 1차 M-v0.2.0~v0.2.2 = rule-based 만)** | `POST /api/v0-2/bundles/{bundle}/rebuild` | `{envelope, data: {bundle, concepts, links}}` |
| **(3) 조회 Query** | 자연어 query → context + answer **(M-v0.2.3+ 부터 LLM answer 합성, 1차 M-v0.2.0~v0.2.2 = 단순 retrieval + raw context 표시)** | `POST /api/v0-2/query` | `{envelope, data: {answer, contexts: [concept_id, ...]}}` |
| | concept 직접 조회 | `GET /api/v0-2/concepts/{type}/{name}` | `{envelope, data: {frontmatter, body}}` |
| | full-text search | `GET /api/v0-2/search?q=...` | `{envelope, data: {hits: [...]}}` |
| | bundle index (progressive disclosure) | `GET /api/v0-2/bundles/{bundle}/index.md` | raw markdown |
| | self-contained graph viewer | `GET /api/v0-2/bundles/{bundle}/viz.html` | raw HTML (Cytoscape.js CDN embed) |

> **인증 정책 (모든 endpoint 공통)**: **internal-only, no auth**. `/api/v0-2/*` 전체가 인증 없이 호출 가능. 별도 gateway / firewall / IP allowlist 가 보호 (Phase 1~3 의 운영 책임, §2.3 참조). 운영자 또는 별도 agent 가 호출. OIDC / Keycloak / backend-core 인증 위임 ❌

### 3.2 Concept `type` enum (8종, 1차)

| type | 정의 | 예시 |
| --- | --- | --- |
| `dataset` | 외부 DB / table 정의 | `hrdb.persons`, `gitea.repositories` |
| `metric` | 운영 metric 정의 (Prometheus / KPI) | `repo_kpi_sync_duration_seconds` |
| `api_endpoint` | 외부 API endpoint 정의 | `gitea_api_v1_repos_list` |
| `runbook` | 운영 매뉴얼 | `gitea_repo_pull_failure_recovery` |
| `integration` | `backend-knowledge` ↔ 외부 시스템 1쌍 정의 | `homelab_file_puller` |
| `event` | webhook payload 정의 | `gitea_push_event` |
| `reference` | 외부 문서 mirror (1차 raw) | `keycloak_admin_rest_api_v1` |
| `decision` | ADR-style concept (in-bundle ADR) | `decision_2026_06_17_backend_knowledge_creation` |

### 3.2.1 5 카테고리 결정 (2026-06-17, 1차 wire 정공법)

**외부 시스템 5 카테고리 결정** (사용자 결정 기반, Gitea 1 instance 의 4 sub-plugin 으로 1차 wire):

| # | 카테고리 | 1차 wire 후보 | Gitea 통합 1차 wire (M-v0.2.0 PoC) |
| --- | --- | --- | --- |
| 1 | **이슈 트래커** | Jira, GitHub Issues, GitLab Issues, Linear | `gitea_issue` |
| 2 | **위키 / 문서** | Confluence, Notion, Docusaurus, GitBook | `gitea_wiki` |
| 3 | **형상관리 (SCM)** | Bitbucket, GitHub, GitLab, Gitea | `gitea_repo_pull` |
| 4 | **CI/CD** | Bamboo, Jenkins, GitHub Actions, GitLab CI, CircleCI, Argo CD | `gitea_action` |
| 5 | **코드 품질** (있으면 좋음) | SonarQube, Snyk, Codecov, Dependabot | (1차 scope 외, 2차 wire) |

**제외 카테고리** (5 카테고리 외, 별도 결정): 협업/메시징, 인증/SSO (Keycloak 은 backend-core 책임 — ADR-0019 정합), 컨테이너/오케스트레이션, 모니터링, HR, 결제 등.

**5 카테고리 표시 방식**: OKF `type` enum 8종 유지 + `x_devhub_category` 필드 추가 (§3.3 정합). OKF spec 의 "extra keys 자유" 원칙 정합, vendor-neutral 유지. type enum 자체에 `issue_tracker` / `wiki` / `scm` / `cicd` / `code_quality` 추가 안 함 (1차 wire 부담 회피, §3.3 `x_devhub_category` 단일 enum 으로 카테고리 표현).

### 3.3 Frontmatter spec (본 프로젝트 적용)

```yaml
---
# OKF 표준 (필수)
type: metric                          # 8종 enum (§3.2)

# OKF 표준 (권장)
title: "Repository KPI sync duration"  # human-readable title
description: "..."                    # 1-2 sentence 요약
resource: "https://..."               # 외부 resource URL (있으면)
tags: ["kpi", "sync", "gitea"]        # 검색/필터용
timestamp: "2026-06-17T20:00:00+09:00"  # 마지막 갱신

# DevHub 확장 (옵션, vendor-neutral 유지 위해 prefix 1글자)
x_devhub_source: "gitea_repo_pull"    # 어느 source plugin 으로 부터 생성됐는지
x_devhub_raw_ref: "raw://..."         # 1차 raw concept 참조
x_devhub_bundle: "devhub-gitea"       # 소속 bundle
x_devhub_version: 3                   # 갱신 회차
x_devhub_curator: "rule-based"        # "rule-based" | "llm" | "human"
x_devhub_category: "scm"              # 5 카테고리 enum (§3.2.1): issue_tracker / wiki / scm / cicd / code_quality (5 카테고리 결정, 2026-06-17). 5 카테고리 외 시스템은 x_devhub_category 미설정 또는 별도 tag
---
```

> **정책**: `x_devhub_*` prefix 로 본 프로젝트 확장을 명시. OKF spec 의 "extra keys 자유" 원칙 정합. consumer 가 OKF 만 보면 unknown key 무시 가능.

### 3.4 Envelope (독립 정의)

- **envelope = `backend-knowledge` 자체 정의** (backend-core 의 `docs/api/conventions.md` 와 format 호환 유지, **import ❌, cross-reference 만**):
  ```json
  { "envelope": { "version": "v0", "trace_id": "..." }, "data": { ... } }
  ```
- **common enum**: `error.code` (`E_OK`, `E_NOT_FOUND`, `E_CONFLICT`, `E_VALIDATION`, `E_UNAUTHORIZED`, `E_FORBIDDEN`, `E_RATE_LIMIT`, `E_INTERNAL`) — backend-core 와 format 호환 (cross-backend client 가 envelope parser 재사용 가능)
- **OpenAPI**: `/openapi.json` (FastAPI 자동 생성) — `docs/openapi.yaml` 와는 별도 (`x-internal: backend-knowledge` 표기)

## 4. 1차 raw 데이터의 API 정책 (사용자 강조)

> **"1차 raw 데이터는 여타 백엔드의 데이터들과 동일하게 api를 통해서 조회하고 추가할 수 있어야 해."**

### 4.1 정책 정의

| 항목 | 정책 |
| --- | --- |
| **저장 위치** | `backend-knowledge/var/raw/{source}/{slug}.json` (file system, **봉투 암호화 후 git 가능**, ADR-0025 정합. raw 자체는 민감 정보일 수 있어 **민감 source 의 경우 .gitignore 권장**) |
| **메타** | sqlite `raw_index` table (id, source, slug, path, ingested_at, byte_size) |
| **API** | `POST /api/v0-2/raw` + `GET /api/v0-2/raw/{type}/{name}` + `GET /api/v0-2/raw?source=...&since=...` (list) + `DELETE /api/v0-2/raw/{id}` |
| **envelope** | **독립 정의** (자체, backend-core 의 `docs/api/conventions.md` 와 format 호환 유지, **import ❌**, cross-reference 만, §3.4 정합) |
| **인증** | **internal-only, no auth** (gateway / firewall / IP allowlist 별도 보호, §2.3). OIDC ❌, Keycloak ❌, backend-core 인증 위임 ❌ |
| **동기** | 동기 응답 (단일 raw concept 추가/조회). 비동기 sync 는 `/ingest/{source}/sync` 의 별도 endpoint |
| **idempotency** | `(source, slug)` unique. 중복 POST → 기존 id 반환 (201 대신 200) |

### 4.2 다른 backend 와의 정합 — ❌ standalone 정책 (2026-06-17 결정)

**`backend-knowledge` 는 완전 standalone 시스템. 다른 backend (backend-core / 다른 백엔드 / 다른 시스템) 와의 연결 / API 호출 / envelope / repository / 어떤 layer 든 공유 / import ❌ (2026-06-17 결정, §4 self-review 강화).**

- path prefix `/api/v0-2/` 은 major version 명시 (standalone 의 자체 정책, 다른 backend 와의 namespace 충돌 회피용)
- envelope format 자체 정의 (§3.4, §4.1 정합) — 다른 backend 와 format 호환 유지하지만 **import / wire / 호출 안 함**
- backend-core 의 `repository-integration` / `integration-registry` / 어떤 도메인 / 어떤 backend 가 신규 API 호출 ❌ (§1.2 G7, §2.3, §7 Q9 정합)
- **§4.2 의 "다른 backend 와의 정합" 표** (이전 version) 는 **standalone 정책 결정 (2026-06-17) 으로 무효**. 외부 시스템 **7종** source (M-v0.2.3 운영 기준, §1.2 G3 / §6.4) 만 단방향

### 4.3 Phase 별 backend-knowledge 의 위치 (standalone 유지)

| Phase | 시점 | backend-knowledge 의 위치 |
| --- | --- | --- |
| **Phase 1 — 독립 1차** | M-v0.2.0~v0.2.1 | **standalone**. docker-compose 로 단독 기동. mock source (filesystem fixture). **internal-only no auth (gateway 별도)**. **기존 backend-core 와 네트워크 분리**. |
| **Phase 2 — 외부 시스템 6종 wire + backend-ai 폐기** | M-v0.2.2 | 외부 시스템 **6종** source plugin wire (**Gitea 4 sub-plugin** gitea_repo_pull / gitea_issue / gitea_wiki / gitea_action + homelab + metrics, §6.4 정합). backend-ai/ 폐기 (단독 결정, placeholder). backend-core 와 wire ❌ |
| **Phase 3 — LLM enrich (Pi) + hrdb 운영** | M-v0.2.3 | **+ 외부 시스템 7종 wire** (+ hrdb) + Pi `pi-coding-agent` SDK or RPC mode 로 LLM enrich 활성화 (1 vendor, §2.2). 장기 multi-vendor (M-v0.3.0+). 풀 RAG 는 M-v0.3.0 |

> **standalone 유지 정책 (2026-06-17 결정)**: 모든 Phase 에서 `backend-knowledge` 는 **완전 standalone 시스템**. 다른 backend (backend-core / 다른 백엔드 / 다른 시스템) 와의 연결 / API 호출 / envelope / repository / 어떤 layer 든 공유 / import ❌. 외부 시스템 **7종** source 만 단방향 (M-v0.2.3 운영 기준, §1.2 G3 / G7, §2.3, §7 Q9 정합)

## 5. 마일스톤 + 우선순위 (P0~P3)

### 5.1 마일스톤 표

| ID | 마일스톤 | scope | 의존 | status |
| --- | --- | --- | --- | --- |
| **M-v0.2.0-alpha** | 컨셉 umbrella + child doc 정합 + PoC 진입 | 본 문서 publish + `external-integrations-agentic-rag-roadmap.md` cross-link | (없음) | ⏳ planned (v0.2.0-alpha) |
| **M-v0.2.0** | 1차 standalone 구현 (Gitea 통합 4 sub-plugin PoC, backend 단독, frontend 0) | `backend-knowledge/` skeleton + OKF spec model (frontmatter `x_devhub_category` 필드 추가) + **Gitea 통합 4종** (`gitea_repo_pull` / `gitea_issue` / `gitea_wiki` / `gitea_action`, Gitea 1 instance 의 4 sub-plugin, 5 카테고리 중 4: 이슈/위키/SCM/CI-CD) + `homelab_mock` 1종 = **5종 PoC (5 카테고리 결정 기반, 2026-06-17, §3.2.1 / §6.4 정합)** + Ingest 1 endpoint + Query 1 endpoint (concept 직접 조회) + 1차 raw API + OpenAPI. **frontend 0 page** (M-v0.2.0 만, viz.html 자가 viewer 만 SSR) | M-v0.2.0-alpha | ⏳ planned (v0.2.0) |
| **M-v0.2.1** | 1차 완성 + Gitea 통합 정식 + 사내 시스템 wire + Curate + frontend 관리/조회 page 1 | Gitea 통합 4종 정식 wire (1차 PoC → 정식, 5 카테고리 중 4) + `homelab_mock` → `homelab` (real wire, 5 카테고리 외 사내 시스템) = 5종 운영 + Curate 3 endpoint (enrich / edit / rebuild) + 1차 viz.html (자가 viewer) + **frontend 관리/조회 page 1** (`backend-knowledge/web/`, 별도 standalone frontend, **devhub frontend 와 분리**, standalone 정책 정합) + e2e smoke | M-v0.2.0 | ⏳ planned (v0.2.1) |
| **M-v0.2.2** | 5 카테고리 외 추가 wire + backend-ai 폐기 | M-v0.2.1 의 5종 + `metrics` 정식 wire (모니터링, 5 카테고리 외) = 6종 운영 + `backend-ai/` 디렉터리 제거 (단독 결정) | M-v0.2.1 | ⏳ planned (v0.2.2) |
| **M-v0.2.3** | Pi LLM enrich + cross-link 자동 resolution | + Pi `pi-coding-agent` SDK or RPC mode 로 LLM enrich 활성화 (1 vendor) + cross-link 자동 resolution | M-v0.2.2 | ⏳ planned (v0.2.3) |
| **M-v0.3.0** | 풀 RAG (chunking + embedding + retrieval) | sentence-transformers or 외부 embedding + vector index (sqlite-vss or pgvector) + reranking + LLM answer | M-v0.2.3 | ⏳ planned (v0.3.0) |

### 5.2 P0~P3 우선순위 (1차)

| 우선순위 | 항목 | 마일스톤 |
| --- | --- | --- |
| **P0** | OKF spec model (frontmatter / link_graph) | M-v0.2.0 |
| **P0** | 1 source plugin (PoC) | M-v0.2.0 |
| **P0** | Ingest + Query 핵심 1 endpoint | M-v0.2.0 |
| **P0** | 1차 raw API (envelope 정합) | M-v0.2.0 |
| **P0** | OpenAPI 자동 생성 | M-v0.2.0 |
| **P1** | 2 source plugin + Curate 3 endpoint | M-v0.2.1 |
| **P1** | viz.html (자가 viewer) + **frontend 관리/조회 page 1** (`backend-knowledge/web/`, 별도 standalone frontend, devhub frontend 와 분리) | M-v0.2.1 |
| **P1** | e2e smoke (ingest → curate → query) | M-v0.2.1 |
| **P2** | 외부 시스템 **6종** source wire (Gitea 4 + homelab + metrics) + backend-ai 폐기 (단독) | M-v0.2.2 |
| **P3** | Pi LLM enrich (1 vendor) | M-v0.2.3 |
| **P3** | 풀 RAG (embedding + vector index) | M-v0.3.0 |

### 5.3 1차 sprint 진입 (M-v0.2.0) 체크리스트

sprint 진입 시 다음 6 항목 확인:

1. [ ] 본 문서 (`release_v0-2_roadmap.md`) publish + cross-link 정합
2. [ ] `external-integrations-agentic-rag-roadmap.md` §8 변경 이력에 row 추가 (umbrella doc publish)
3. [ ] `ai-workflow/memory/state.json` M-v0.2.0 row 발급 (또는 v0.1.x status update)
4. [ ] `backend-knowledge/` 디렉터리 skeleton (Dockerfile, pyproject.toml, main.py, okf/spec.py)
5. [ ] OKF `SPEC.md` 1차 정독 (vendor-neutral 정책 + frontmatter 정확한 spec 확인)
6. [ ] 신규 GitHub milestone `v0.2.0` 생성 + 본 문서 link 첨부

## 6. 1차 독립 개발 → 연동 단계 (Phase 1 / 2 / 3)

### 6.1 Phase 1 — 1차 standalone (M-v0.2.0 + M-v0.2.1)

- **독립 기동**: **`backend-knowledge/dev-up.sh` (별도 스크립트, backend-knowledge 만 단독 기동)** + **`backend-knowledge/docker-compose.yml` (별도, backend-knowledge 서비스만)**. **devhub 의 root `dev-up.sh` / `docker-compose.colima.yml` / `docker-compose.deploy.yml` 사용 ❌** (다른 backend 연결 방지, §1.2 G7 / §2.3 / §4.2 standalone 정책 정합). 별도 docker network.
- **mock source**: filesystem fixture 1개 (`backend-knowledge/var/fixtures/homelab/*.json`, `homelab_mock.py` 의 입력, 5 카테고리 외 사내 시스템). 외부 시스템 실제 호출 ❌. **Gitea 통합 4 sub-plugin 은 M-v0.2.0 부터 real Gitea 1 instance 에 PoC wire** (mock 없이 실제 API, 1차 wire 정공법 — §3.2.1 / §6.4 정합).
- **인증**: **internal-only, no auth** (gateway / firewall / IP allowlist 별도 보호, §2.3). OIDC / Keycloak / backend-core 인증 위임 ❌.
- **테스트**: unit (OKF spec / enricher / link_graph) + e2e (ingest → curate → query) 의 **신규 백엔드 단독**.
- **frontend**: **M-v0.2.0 만 frontend 0 page** (1차 backend 단독 구현). **M-v0.2.1 부터 frontend 관리/조회 page 1 추가** (§5.1 M-v0.2.1 / §5.2 P1 정합, `backend-knowledge/web/` 별도 standalone frontend, devhub frontend 와 분리, §1.2 G7 standalone 정책 정합). viz.html 자체 viewer (자가 graph viewer) 는 backend-knowledge 가 SSR (모든 Phase 공통)

### 6.2 Phase 2 — 외부 시스템 6종 source wire + backend-ai 폐기 (M-v0.2.2)

- **외부 시스템 6종 source wire** (M-v0.2.2 운영 기준, §1.2 G7 / §4.3 / §5.1 정합): **Gitea 4 sub-plugin** (gitea_repo_pull / gitea_issue / gitea_wiki / gitea_action) + homelab (real) + metrics. docker-compose wire (별도 docker network 유지, backend-core 와 공유 ❌).
- **M-v0.2.3 추가 wire**: + hrdb = 7종 운영 + Pi (pi.dev) LLM enrich 활성화 (§4.3 Phase 3 정합).
- **backend-ai/ 폐기** (단독 결정):
  - `backend-ai/` 디렉터리 제거 (placeholder 상태, 이전 코드 0)
  - `docker-compose.deploy.yml` 의 `backend-ai` service 제거 (root level, 사내 한정 tier, `backend-knowledge/docker-compose.yml` 와 무관)
  - `docs/` 의 backend-ai 관련 reference 제거 (있다면)
  - `Makefile` / devhub 의 root `dev-up.sh` 의 backend-ai target 제거 (root level 정리, `backend-knowledge/dev-up.sh` 와 무관)
- **cross-backend 호출**: ❌. `backend-core/internal/integrations/adapters/` 의 source plugin 호출이 `backend-knowledge` API 로 전환되지 않음 (backend-core 참조 전면 금지). backend-knowledge 는 외부 시스템만 단방향
- **e2e 통합 테스트**: `backend-knowledge/tests/e2e/` 에 `e2e-knowledge-*` spec 추가. backend-knowledge 단독 e2e (외부 시스템 6종 source → curate → query happy path, M-v0.2.2 기준)

### 6.3 Phase 3 — LLM enrich (Pi) (M-v0.2.3 / M-v0.3.0)

- **Pi `pi-coding-agent` SDK or RPC mode**: M-v0.2.2 까지 rule-based enrich 만, M-v0.2.3 부터 Pi 의 SDK mode (Node subprocess via @earendil-works/pi-coding-agent npm pkg) or RPC mode (JSON over stdin/stdout) 로 LLM enrich 활성화 (§2.2)
- **1차 vendor**: 1 vendor 선택 (운영자 결정 시점). `pi-ai` 의 15+ provider 중 (Anthropic / OpenAI / Google / Azure / Bedrock / Mistral / Groq / Cerebras / xAI / Hugging Face / Kimi For Coding / MiniMax / OpenRouter / Ollama). 장기 multi-vendor (M-v0.3.0+)
- **cross-link 자동 resolution**: link_graph 가 unresolved link 발견 시 Pi LLM 에 "가장 유사한 concept 추천" 요청
- **풀 RAG** (M-v0.3.0): chunking + embedding + vector index + retrieval + reranking

### 6.4 source plugin 작성 표 (외부 시스템 API spec 기반, 0에서 Python 작성)

| 외부 시스템 API | 카테고리 | 신규 위치 | 작성 시점 | 비고 |
| --- | --- | --- | --- | --- |
| **Gitea 통합 4 sub-plugin** (Gitea 1 instance, 5 카테고리 중 4 wire) | | | | Gitea 1 instance 가 4 카테고리 (이슈/위키/SCM/CI-CD) 의 4 sub-plugin 으로 wire. 테스트/개발 환경 단일 외부 시스템. |
| Gitea Repository (git HTTP + REST) — 공식 docs | **SCM (3)** | `backend-knowledge/sources/gitea_repo_pull.py` | M-v0.2.0 (PoC) / M-v0.2.1 (정식) | Gitea REST API (`/api/v1/repos/{owner}/{repo}`) + git HTTP |
| Gitea Issue (`/api/v1/repos/{owner}/{repo}/issues`) — 공식 docs | **이슈 트래커 (1)** | `backend-knowledge/sources/gitea_issue.py` | M-v0.2.0 (PoC) / M-v0.2.1 (정식) | Gitea Issue API |
| Gitea Wiki (`/api/v1/repos/{owner}/{repo}/wiki`) — 공식 docs | **위키 (2)** | `backend-knowledge/sources/gitea_wiki.py` | M-v0.2.0 (PoC) / M-v0.2.1 (정식) | Gitea Wiki API |
| Gitea Action (Gitea Actions, `/api/v1/repos/{owner}/{repo}/actions`) — 공식 docs | **CI/CD (4)** | `backend-knowledge/sources/gitea_action.py` | M-v0.2.0 (PoC) / M-v0.2.1 (정식) | Gitea Actions (GitHub Actions 호환) |
| **사내 시스템** (5 카테고리 외) | | | | |
| homelab (사내 HomeLab agent, file/HTTP) — 공식 API spec | 사내 시스템 (5 카테고리 외) | `backend-knowledge/sources/homelab.py` | M-v0.2.0 (PoC, `homelab_mock.py` filesystem fixture) / M-v0.2.1 (정식, real wire) | M-v0.2.0 = `homelab_mock` (filesystem fixture), M-v0.2.1 = real wire |
| **5 카테고리 외 시스템** (선택 wire) | | | | |
| prometheus (Prometheus scrape API) — 공식 docs | 모니터링 (5 카테고리 외) | `backend-knowledge/sources/metrics.py` | M-v0.2.2 | 사용자 결정 (2026-06-17): 5 카테고리 (이슈/위키/SCM/CI-CD/코드 품질) 외. 별도 카테고리. |
| hrdb (사내 HR DB, PostgreSQL) — 사내 schema spec | HR/조직 (5 카테고리 외) | `backend-knowledge/sources/hrdb.py` | M-v0.2.3 | 동일. |

> **정책**: **외부 시스템 공식 API spec 만 참조** (vendor docs, OpenAPI, GraphQL schema 등). backend-core 의 Go adapter / repository / model / 어떤 layer 든 참조 ❌. backend-knowledge 의 source plugin 은 외부 시스템 API 와 1:1 대응. OKF 형 concept emit + source plugin interface 정합 (§2.1 `sources/_base.py`) 이 우선. backend-core 측의 기존 Go adapter 제거는 별도 PR (backend-knowledge 책임 아님)

## 7. Risks + Q&A 결정 (2026-06-17 11/11 결정 완료)

| # | 항목 | 옵션 | 추천 |
| --- | --- | --- | --- |
| Q2 | **언어** (2026-06-17 결정) | Python 3.13+ (OKF reference 정합) / Go (backend-core 정합) | **Python 3.13+** (OKF reference 가 Python, LLM SDK 호환, vendor-neutral 정책, §2.2 정합) |
| Q3 | **1차 LLM vendor** (2026-06-17 결정) | OpenAI / Anthropic / Gemini / 1차 LLM 없이 (rule-based 만) | **1차 rule-based 만** (M-v0.2.0~v0.2.2, §2.2 / §1.3 / §3.1 정합). v0.2.3+ 부터 Pi (pi.dev) SDK or RPC mode 로 LLM enrich 활성화 (1 vendor, 시점 결정) |
| Q4 | **신규 백엔드 의 frontend page** (2026-06-17 결정) | viz.html 만 / frontend wiki viewer page / 둘 다 | **viz.html (자가 viewer, 모든 Phase 공통) + frontend 관리/조회 page 1 (별도 standalone frontend, `backend-knowledge/web/`, devhub frontend 와 분리)** — M-v0.2.0 frontend 0 page, M-v0.2.1 부터 frontend 1 page 추가 (§5.1 / §5.2 P1 / §6.1 정합) |
| Q5 | **bundle 저장소** (2026-06-17 결정) | file system only (OKF 원칙) / git LFS / sqlite blob | **file system only** (1차, §2.1 `var/bundles/`, OKF 원리 정합, git 가능). v0.3.0+ 에서 git LFS 검토 |
| Q6 | **5 source plugin 의 1차 채택** (2026-06-17 결정, 5 카테고리 결정 기반) | homelab / gitea_pull 중 1 + mock / 둘 다 + mock / 5개 전부 | **Gitea 통합 4종 (gitea_repo_pull / gitea_issue / gitea_wiki / gitea_action, Gitea 1 instance 의 4 sub-plugin, 5 카테고리 중 4: 이슈/위키/SCM/CI-CD) + homelab_mock 1종 = 5종 PoC** (M-v0.2.0, §2.1 / §3.2.1 / §5.1 / §6.4 정합). 5 카테고리 5번째 (코드 품질) = 2차 wire (SonarQube/Snyk). M-v0.2.1 = Gitea 4 정식 + homelab (real) = 5종 운영. M-v0.2.2 = + metrics = 6종 운영 + backend-ai 폐기. M-v0.2.3 = + hrdb = 7종 운영 + Pi LLM enrich |
| Q7 | **기존 `external-integrations-agentic-rag-roadmap.md` 의 status** (2026-06-17 결정) | draft 유지 (v0.2.0 umbrella publish 후 active 전환) / active 전환 (umbrella 후 child) | **umbrella publish 시점 → active 전환** (본 문서가 umbrella 확정 signal, §0.4 Keycloak 분류 정합) |
| Q8 | **ADR 신규** (2026-06-17 결정) | `ADR-0034 OKF 채택` + `ADR-0035 backend-knowledge 신설` 2건 / 1건 통합 / ADR 없이 본 문서로 충분 | **ADR-0034 OKF 채택 + ADR-0035 backend-knowledge 신설 2건 분리** (각각 1개 결정 = 1 ADR 원칙, 1 ADR = 1 결정) |
| Q9 | **다른 backend 와의 관계 (general, standalone 정책)** (2026-06-17 결정) | backend-core 와 wire / 다른 backend 와 wire / 어떤 backend 의 source plugin 호출 / 어떤 backend 의 어떤 layer 든 import | **❌ 전면 금지** (2026-06-17 결정, §4 self-review 강화). `backend-knowledge` 는 **완전 standalone 시스템**. 다른 backend (backend-core / 다른 백엔드 / 다른 시스템) 의 어떤 코드 / API / 데이터도 import / 호출 / 공유 ❌. **외부 시스템 7종 source 만 단방향** (Gitea 4 sub-plugin + homelab + metrics + hrdb, M-v0.2.3 운영 기준, §1.2 G3 / G7 정합) |
| Q10 | **API 인증 정책** (2026-06-17 결정) | OIDC bearer token / API key / session / no auth (gateway 별도) | **internal-only, no auth** (2026-06-17 결정). `/api/v0-2/*` 전체 인증 없이 호출 가능. 별도 gateway / firewall / IP allowlist 가 보호 (운영 책임, §2.3) |
| Q11 | **source plugin 의 source of truth** (2026-06-17 결정) | backend-core 의 Go adapter / 외부 시스템 공식 API spec / 둘 다 | **외부 시스템 공식 API spec 만** (vendor docs, OpenAPI, GraphQL schema 등, 2026-06-17 결정). backend-core 의 Go adapter / repository / model 참조 ❌ (§1.2 G3 정합) |

## 8. 결정 timeline (장기)

| 시점 | 결정 |
| --- | --- |
| 2026-06-10 | 외부 시스템 연동 = agentic RAG 와 함께 발전 (사용자 결정) → `external-integrations-agentic-rag-roadmap.md` draft |
| 2026-06-12 | Google OKF v0.1 발표 (외부事实, 사용자 결정 무관) |
| 2026-06-17 | v0.2.0 umbrella 컨셉 + OKF 차용 + `backend-knowledge` 통합 결정 (본 문서) |
| 2026-06-17 | **Q1 결정** — 신규 백엔드 이름 = `backend-knowledge` (사용자 결정, "지식 도서관" 의미) |
| 2026-06-17 | **Q9 결정** — 다른 backend 와의 관계 (general, standalone 정책): ❌ 전면 금지 (§1.2 G7 / §2.3 / §4.2 정합) |
| 2026-06-17 | **Q10 결정** — API 인증 정책: internal-only, no auth (§2.3 / §3.1 정합) |
| 2026-06-17 | **Q11 결정** — source plugin 의 source of truth: 외부 시스템 공식 API spec 만 (§1.2 G3 / §6.4 정합) |
| 2026-06-17 | **Q4 결정** — frontend page: viz.html (자가 viewer) + frontend 관리/조회 page 1 (별도 standalone frontend, devhub frontend 와 분리, §5.1 / §5.2 P1 / §6.1 정합) |
| (예정) | **Q2·Q3·Q5~Q8 결정 완료** (2026-06-17 결정, §7 정합 — **Q2 (언어 = Python 3.13+) / Q3 (1차 LLM = rule-based 만) / Q5 (bundle = file system only) / Q6 (1차 source = homelab + mock) / Q7 (child doc = active 전환) / Q8 (ADR-0034 + ADR-0035 2건 분리)**) |
| 2026-06-17 | **ADR-0034 + ADR-0035 publish 완료** — [ADR-0034 OKF 채택](../adr/0034-okf-adoption.md) + [ADR-0035 backend-knowledge 신설](../adr/0035-backend-knowledge-creation.md) (Q8 결정 정합) |
| 2026-06-17 | **B 결정 (OIDC 제외 + 기존 시스템 백엔드 참조 ❌ + 외부 시스템 only)** + **Pi (pi.dev) 채택 결정** — cross-section 정합 round 2 진행 (§3/§4/§5/§6/§7/§8 영향 section 일괄 fix) | 사용자 2026-06-17 결정 ("이 시스템은 우리 기존 시스템의 백엔드를 참조하지 않고 외부 시스템만 바라보는 구조로 되어있어야해" + "pi coding agent를 사용해볼거야") |
| (예정) | M-v0.2.0 sprint 진입 (6 checklist 통과 후) |

## 9. 변경 이력

| 일자 | 변경 | 근거 |
| --- | --- | --- |
| 2026-06-17 | v0.2.0 umbrella 컨셉 1차 — OKF 차용 + 3가지 기능 + 1차 raw API + 6 마일스톤 + 8 Q&A | 사용자 2026-06-17 결정 ("외부 시스템 연동 및 데이터 취합을 별도의 백엔드로 모을거야… google의 okf에 대해 조사하고 구조를 참조") |
| 2026-06-17 | **Q1 결정 반영** — 신규 백엔드 = `backend-knowledge` 확정. §1.2 G1 / §2 제목 / §7 Q1 row / §8 timeline / 본문 "(가칭)" 표기 일괄 제거. status: draft | 사용자 결정 (AskUser 응답 `backend-knowledge`) |
| 2026-06-17 | **§1 self-review 6 항목 fix** — (1) §1.1 4번 한계 (§1.1 4번 → §1.3 motivation 이관) + 한계 1/2/3 concrete DevHub 시나리오 보강 (2) §1.2 G2 "폐기 + 흡수" → "폐기" (placeholder 정합) (3) §1.2 G4 "자동" → "rule-based + LLM-optional" (Q3 정합) (4) §1.3 "본 프로젝트 적용" producer 다중 모순 정정 — 1차 rule-based / M-v0.2.3+ LLM / M-v0.3.0+ multi-vendor LLM 단계 명시 (5) §1.3 "자주 묻는 차이점" 3번 — v0.2.0 Query 의 RAG 범위를 1차/2차/3차 단계로 명시 (6) §1.2 G1/G6/G7 concrete 예시 (bundle 이름 + raw 데이터 종류 + Phase 명시) | self-review (사용자 "섹션 1부터 상세 리뷰하자" 지시 + "발견된 사항은 다 수정해줘" 후속 지시) |
| 2026-06-17 | **§2 self-review 5+1 항목 fix (사용자 결정 5개 반영)** — (1) §2.2 LLM row 모순 정정 — "OpenAI / Anthropic / Gemini 3 vendor" → **"pi 와 같은 하네스" 패턴** (vendor-agnostic LLM abstraction: provider-agnostic interface + tool calling + agent loop + context/memory management) + 1차 rule-based / v0.2.3+ 1 vendor / v0.3.0+ multi-vendor (Q3 정합) (2) §2.1 sources/ 에 `sources/_base.py` (SourcePlugin ABC + Credential config schema) + source plugin 파일별 1차/후속 분류 명시 (Q6 정합) + §2.2 에 "credential 관리" row 추가 (id/pw or token type-agnostic string, 연결 시 source plugin config 주입, 봉투 암호화 저장 ADR-0025) (3) §2.1 디렉터리 tree 에 `var/raw/{source}/{slug}.json` 추가 (§4 cross-ref 정합) + bundle 5종 전체 명시 (homelab/gitea/hrdb/metrics/task-item) (4) sources/ 1차/후속 분류 명시 — 1차 PoC = homelab + homelab_mock, v0.2.1 = gitea_pull, v0.2.3 = task_item_puller + metrics + hrdb (Q6 정합) (5) api/ endpoint 분포 매핑 명시 (5 module ↔ §3.1 endpoint) + main.py uvicorn dev-only 분리 | self-review (사용자 "섹션2 넘어가자" 지시 + 5개 결정 응답: pi 하네스 / credential type-agnostic / var/raw 정합 / sources 분류 / api+main.py 수정) |
| 2026-06-17 | **컨셉 재구성 round 1 (Pi / pi.dev 조사 + OIDC 제외 + 외부 시스템 only 결정 반영)** — (1) **Pi (pi.dev, github.com/earendil-works/pi, MIT, v0.79.6, Mario Zechner / badlogic, 회사: Earendil Inc.)** 메인 4 package 확인: `pi-ai` (multi-provider LLM API, 15+ provider) + `pi-agent-core` (agent runtime) + `pi-coding-agent` (interactive coding agent CLI, 4 mode: Interactive / Print·JSON / **RPC** / **SDK**) + `pi-tui` (terminal UI). §2.2 LLM row 에 정확한 4 mode + 15+ provider 명시 (2) **§2.1 에 `pi_bridge/`** 디렉터리 추가 (`rpc_client.py` RPC mode + `sdk_client.py` SDK mode + `tools.py`, M-v0.2.3+ 부터 활성화) (3) **§1.2 G3** — "기존 backend-core 코드 참조 ❌, 외부 시스템 공식 API spec 만 보고 0에서 source plugin 작성" 으로 의미 변경 (4) **§1.2 G7** — "기존 시스템과 wire ❌, OIDC ❌, 외부 시스템 only" 로 의미 변경 (5) **§2.3 전면 재작성** — OIDC 제외 + backend-core 참조 전면 금지 + 외부 시스템 only + internal-only no auth + Phase 2 의미 변경 (backend-core wire → 외부 시스템 5종 source wire) (6) **round 1.1 정합 fix**: web_search snippet 기반의 추정 표현들 ("7-package monorepo" / "20+ model" / "vendor-neutral" / "< 1000 token system prompt" / "pi-mono" naming) 을 pi.dev / github README 1차 출처로 정정 | self-review (사용자 "pi.dev 조사해봐 pi coding agent를 사용해볼거야 / 이 시스템은 OIDC 제외 / 기존 백엔드 참조 안 함 / 외부 시스템만" 결정) + "pi-mono는 뭐지? https://pi.dev 여기 참조한거 맞니?" 검증 후 round 1.1 정합 fix |
| 2026-06-17 | **컨셉 재구성 round 2 (cross-section 정합 일괄 fix)** — (1) §3.1 API 매트릭스에 "인증 = internal-only, no auth" 코멘트 추가 + §3.4 envelope 자체 정의 (cross-reference 만, import ❌) (2) §4.1 인증 row "OIDC 위임" → "internal-only, no auth, gateway 별도" (3) §4.2 다른 backend API 와의 정합 — backend-core 의 어떤 도메인도 신규 API 호출 ❌ 명시 (4) §4.3 Phase 1/2/3 표 — Phase 2 의미 변경 (backend-core wire → 외부 시스템 5종 source wire), Phase 3 = Pi LLM enrich (5) §5.1 M-v0.2.2 = "외부 시스템 5종 source wire + backend-ai 폐기" + M-v0.2.3 = "Pi LLM enrich" (6) §5.2 P2/P3 정합 + row 합치기 (7) §6.1 Phase 1 "OIDC 검증 stub" → "internal-only no auth" + §6.2 Phase 2 전면 재작성 (backend-core wire 전면 ❌, 외부 시스템 5종 source wire + backend-ai 폐기만 잔류) + §6.3 Phase 3 Pi 정합 + §6.4 source plugin 작성 표 의미 변경 (외부 시스템 API spec 기반, backend-core Go adapter 참조 ❌) (8) §7 Q&A Q9 (backend-core 와의 관계) / Q10 (API 인증) / Q11 (source plugin source of truth) 추가 (9) §8 timeline B 결정 row 추가 | self-review (사용자 "round 2 진행하자" 지시) |
| 2026-06-17 | **§3 self-review 5 fix (사용자 A 옵션)** — (1) §2.3 line 159 "API 인증" row 의 "backend-core 의 integration-registry 또는 cron worker 가 호출" → "운영자 또는 별도 agent 가 호출 (**backend-core 의 어떤 layer 든 호출 ❌**)" 로 정정 (§1.2 G7 / §7 Q9 정합, **모순 fix**) (2) §3.1 표 raw API 의 GET /raw (list, filter) + DELETE /raw/{id} 추가 (§2.1 raw.py 의 4 endpoint 와 정합) (3) §3.1 표 bundles API 의 GET /bundles (list) + POST /bundles (create) 추가 (§2.1 bundles.py 의 3 endpoint 와 정합) (4) §2.1 query.py 코멘트 "Query 4 endpoint" → "Query 5 endpoint" (§3.1 표 기준 5 endpoint 와 정합) (5) §3.2 `integration` type 정의 "DevHub ↔ 외부 시스템" → "`backend-knowledge` ↔ 외부 시스템" (DevHub = backend-core alias 모호성 해소, 2026-06-17 backend-core 참조 ❌ 정합) | self-review (사용자 "섹션3 진행하자" + "A" 옵션 선택) |
| 2026-06-17 | **§3 self-review 5번 결정 (LLM enrich timing endpoint 별 inline)** — §3.1 표 4 endpoint 의 API 설명 cell 에 timing inline 추가 (A 옵션). 1) `POST /ingest/{source}/sync` = "M-v0.2.3+ 부터 OKF enrich 동시 가능" 2) `POST /concepts/{id}/enrich` = "M-v0.2.3+ 부터 Pi LLM enrich 활성화, 1차 = rule-based 만" 3) `POST /query` = "M-v0.2.3+ 부터 LLM answer 합성, 1차 = 단순 retrieval" 4) `POST /bundles/{bundle}/rebuild` = "M-v0.2.3+ 부터 LLM cross-link 자동 resolution, 1차 = rule-based 만" | self-review (사용자 "4. 결정해보자" + AskUser 응답 "A — endpoint 별 inline") |
| 2026-06-17 | **§5 self-review 2 fix (B 옵션, 사용자 strong guidance: 관리/조회용 frontend 필요)** — (1) §5.1 M-v0.2.0 scope = "1 source plugin (homelab or gitea_pull 중 1)" → "homelab 1 source + homelab_mock 1 source (Q6 정합) + frontend 0 page 명시" (모순 fix) (2) §5.1 M-v0.2.1 scope = "+ frontend 위키 viewer 1 page" → "+ frontend **관리/조회 page 1** (viz.html 자가 viewer 별도, `backend-knowledge/web/` 별도 standalone frontend, **devhub frontend 와 분리**)" (사용자 strong guidance) (3) §5.2 P1 = "viz.html + frontend 위키 viewer 1 page" → "viz.html (자가 viewer) + **frontend 관리/조회 page 1 (별도 standalone frontend)**" (4) §6.1 Phase 1 frontend = "별도 frontend page 0개 (1차 scope 외)" → "M-v0.2.0 만 frontend 0 page, M-v0.2.1 부터 frontend 관리/조회 page 1 추가 (viz.html 자가 viewer 는 모든 Phase 공통)" (의도 명확화) (5) §2.1 디렉터리 tree 에 `web/` 추가 (M-v0.2.1+ frontend 관리/조회 page, 별도 standalone frontend, devhub frontend 와 분리) | self-review (사용자 "좋아 다음 섹션" + "B 명확화하자. 이 시스템도 관리 및 조회용 프론트엔드는 필요할거 같긴 해" strong guidance) |
| 2026-06-17 | **§6 self-review 1 fix (docker-compose / dev-up.sh standalone 정합)** — (1) §6.1 "독립 기동" 의 `dev-up.sh --knowledge` / `docker-compose.knowledge.yml` → "`backend-knowledge/dev-up.sh` (별도 스크립트, backend-knowledge 만 단독) + `backend-knowledge/docker-compose.yml` (별도, backend-knowledge 서비스만). **devhub 의 root `dev-up.sh` / `docker-compose.colima.yml` / `docker-compose.deploy.yml` 사용 ❌**" (standalone 정책 정합, §1.2 G7 / §2.3 / §4.2) (2) §6.2 backend-ai 폐기 부분의 docker-compose.deploy.yml / Makefile / dev-up.sh → root level 정리 (backend-knowledge 의 별도 docker-compose.yml / dev-up.sh 와 무관) 명시 | self-review (사용자 "섹션 6 가자" + docker-compose / dev-up.sh standalone 정합 fix) |
| 2026-06-17 | **§7 self-review 1 fix (Q4 outdated 정합)** — §7 Q4 추천값 "viz.html 만 (v0.2.0 1차). frontend wiki viewer 는 v0.2.1" → "**viz.html (자가 viewer, 모든 Phase 공통) + frontend 관리/조회 page 1 (별도 standalone frontend, `backend-knowledge/web/`, devhub frontend 와 분리)** — M-v0.2.0 frontend 0 page, M-v0.2.1 부터 frontend 1 page 추가" (round 5 strong guidance + §5.1 / §5.2 P1 / §6.1 정합) | self-review (사용자 "섹션 7 가자" + Q4 outdated 정합 fix) |
| 2026-06-17 | **§8 self-review 1 fix (Q2~Q8 outdated 정합 + Q4·Q9·Q10·Q11 결정 row 추가)** — §8 timeline 의 (예정) "Q2~Q8 결정 sprint" → "(예정) **Q2·Q3·Q5~Q8 결정 sprint** (Q1·Q4·Q9·Q10·Q11 결정 완료, **Q2·Q3·Q5·Q6·Q7·Q8 미결정, 추천값 default**)" 로 outdated 정합. Q1 결정 row 다음에 Q9 / Q10 / Q11 / Q4 결정 row 추가 (round 2 / 4 / 5 / 7 결정) | self-review (사용자 "좋아 다음 섹션 8" + §8 timeline 정합 fix) |
| 2026-06-17 | **Q&A 명시 결정 (Q2·Q3·Q5~Q8) — 추천값 default 일괄 진행** — Q2 (언어 = Python 3.13+), Q3 (1차 LLM = rule-based 만, M-v0.2.3+ Pi SDK or RPC), Q5 (bundle = file system only), Q6 (1차 source = homelab + mock), Q7 (child doc = active 전환), Q8 (ADR-0034 + ADR-0035 2건 분리) — §7 Q&A row 에 (2026-06-17 결정) 추가 + 추천값 정정 + §7 제목 "Open questions (결정 필요)" → "Q&A 결정 (2026-06-17 11/11 결정 완료)" + §8 timeline (예정) Q2·Q3·Q5~Q8 결정 sprint row → 결정 완료 정정. **Q&A 11개 모두 결정 완료 (5 + 6 = 11)** | self-review (사용자 "2 가자" + Q&A 명시 결정) |
| 2026-06-17 | **ADR-0034 + ADR-0035 publish** (Q8 결정 기반) — [ADR-0034 OKF v0.1 채택](../adr/0034-okf-adoption.md) (1차 출처: Google SPEC.md / README.md, Apache 2.0, 1 concept = 1 .md, frontmatter `type` 1개 필수, 8종 type enum, `x_devhub_*` prefix 확장) + [ADR-0035 backend-knowledge 신설](../adr/0035-backend-knowledge-creation.md) (Python 3.13+ / FastAPI / OKF bundle + sqlite metadata / Pi (pi.dev) v0.79.6 / standalone 정책 / 다른 backend 연결 ❌ / 외부 시스템 5종 source 만 단방향 / M-v0.2.0~v0.3.0 6 마일스톤). §8 timeline (예정) ADR-0034 + ADR-0035 publish row → publish 완료 정합 | self-review (사용자 "2 가자" + ADR publish) |
| 2026-06-17 | **5 카테고리 결정 + Gitea 통합 1차 wire + `x_devhub_category` 필드 추가** (사용자 결정 A/A) — 5 카테고리 (이슈 트래커 / 위키 / 형상관리 / CI-CD / 코드 품질) 확정. Gitea 1 instance 의 4 sub-plugin (gitea_repo_pull / gitea_issue / gitea_wiki / gitea_action) 으로 1차 PoC (M-v0.2.0) + `homelab_mock` 1종 = 5종. §3.2.1 신규 subsection (5 카테고리 결정) + §3.3 frontmatter spec 에 `x_devhub_category` 필드 추가 (5 enum) + §5.1 M-v0.2.0 scope 정합 (Gitea 통합 4종 + homelab_mock) + §5.1 M-v0.2.1/M-v0.2.2 정합 (Gitea 4 정식 + homelab real + metrics) + §6.4 source plugin 표 8 row 정합 (Gitea 4 + homelab 2 + metrics/hrdb) + §7 Q6 정합 (Gitea 통합 4종 + homelab_mock 1종 = 5종 PoC) + state.json M-v0.2.0 row 의 external_system_5_categories + m_v0_2_0_5_plugins 추가 | self-review (사용자 "A/A" + 5 카테고리 결정 + Gitea 통합 1차 wire + x_devhub_category 필드 추가) |
| 2026-06-17 | **§4 self-review 3 fix (모두 수정) + standalone 정책 강화 (사용자 strong guidance)** — (1) §4.1 "envelope" row `docs/api/conventions.md 정합` → "독립 정의 (import ❌, cross-reference 만, §3.4 정합)" (모순 fix) (2) §4.1 "저장 위치" row "git 가능" → "봉투 암호화 후 git 가능, ADR-0025 정합, 민감 source .gitignore 권장" (3) §4.2 "다른 backend 와의 정합" → "다른 backend 와의 정합 — ❌ standalone 정책" 으로 전면 재작성 (다른 backend 연결 전면 금지, standalone 명시) (4) §4.3 표 제목 "다른 backend 와의 관계 (Phase 1 vs Phase 2)" → "Phase 별 backend-knowledge 의 위치 (standalone 유지)" + 표 아래 standalone 유지 정책 노트 추가 (5) **standalone 정책 cross-section 강화**: §1.2 G7 "완전 standalone 운영" 으로 일반화, §2.3 정책 표 "backend-core 참조" → "다른 backend 연결 (general)" 로 일반화, §7 Q9 "기존 backend-core 와의 관계" → "다른 backend 와의 관계 (general, standalone 정책)" 으로 일반화 | self-review (사용자 "일단 섹션 4로 넘어가자" + "모두 수정해주고, 참고해야할 사항은 이 backend-knowledge는 standalone이야. 다른 백엔드와는 연결이 없어야해." strong guidance) |
| 2026-06-18 | **A/A 결정 cross-section 정합 일괄 fix (9 위치)** — §2.1 `sources/` 트리: 구 `gitea_pull.py` 단일 (v0.2.1) → Gitea 4 sub-plugin (`gitea_repo_pull` / `gitea_issue` / `gitea_wiki` / `gitea_action`, M-v0.2.0 PoC) + `homelab.py` / `homelab_mock.py` / `metrics.py` / `hrdb.py` 정합 + `task_item_puller.py` 제거 (A/A 결정에서 gitea_issue 가 대체) + §2.1 `var/bundles/` 트리에서 `devhub-task-item/` 제거 + §3.2 `runbook` 예시 `gitea_pull_failure_recovery` → `gitea_repo_pull_failure_recovery` + §3.3 `x_devhub_source` 예시 `gitea_pull` → `gitea_repo_pull` + §6.1 mock source "1~2개" → "1개 (homelab_mock)" + Gitea 4 sub-plugin M-v0.2.0 부터 real Gitea 1 instance PoC wire (mock 없이) 명시 + §1.2 G3 "source plugin 5종" → "7종 (Gitea 4 + homelab + metrics + hrdb, M-v0.2.3 운영 기준)" + §1.2 G7 M-v0.2.2 "5종" → "6종", M-v0.2.3 "Pi enrich" → "+ hrdb = 7종 + Pi enrich" + §2.3 "외부 시스템 5종" → "7종 (M-v0.2.3 운영 기준)" + Phase 2 의미 변경 5종 → 6종 + §4.2 "5종" → "7종" + §4.3 Phase 2 "5종 wire" → "6종 wire" + Phase 2 list `homelab + gitea_pull + metrics + task_item_puller + hrdb` → `Gitea 4 + homelab + metrics` + Phase 3 정합 (+hrdb 7종 + Pi enrich) + standalone 유지 정책 5종 → 7종 + §5.2 P2 "5종" → "6종" + §6.2 "5종" → "6종" + list 정합 + e2e 5종 → 6종 + §7 Q9 "5종" → "7종 (M-v0.2.3 운영 기준)" | self-review (사용자 "1 진행하고 3은 별도의 gitea 서버를 연결할거야" + A/A 결정 cross-section 정합 follow-up) |
