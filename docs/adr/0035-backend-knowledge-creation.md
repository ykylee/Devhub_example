# ADR-0035: `backend-knowledge` 백엔드 신설 (외부 시스템 연동 + OKF concept library)

## 1. 상태

- **상태**: Accepted
- **작성일**: 2026-06-17
- **수정일**: 2026-06-18 (umbrella doc §3.6 Path Y caller-provided user context + §4.4~§4.7 raw 운영/API/정합성 정책 신규에 따른 §3.4 1차 raw API 정책 row 갱신 — internal-only no auth + caller-provided user context + raw 운영 정책 + endpoint 권한 + 정합성 검증 명시)
- **결정 근거 sprint**: `docs/work_260617-v0-2-umbrella-concept`
- **supersedes**: 없음 (신규)
- **Tier**: 사외 (vendor-neutral 정책, 다른 backend 연결 ❌)
- **관련 문서**:
  - [release_v0-2_roadmap.md 전체](../planning/release_v0-2_roadmap.md) (umbrella 컨셉)
  - [release_v0-2_roadmap.md §3.6 Data governance & query scoping](../planning/release_v0-2_roadmap.md) (caller-provided user context + curation governance + query scope priority)
  - [external-integrations-agentic-rag-roadmap.md](../planning/external-integrations-agentic-rag-roadmap.md) (외부 연동 분리 detail)
  - [ADR-0034 OKF 채택](./0034-okf-adoption.md) (knowledge bundle 형식)
  - [ADR-0025 봉투 암호화](../adr/0025-envelope-encryption-key-management.md) (credential 관리)
  - [ADR-0019 Keycloak 단일화](../adr/0019-keycloak-only-idp.md) (Keycloak 분류 재확인)

## 2. 컨텍스트

### 2.1 문제 (v0.1.x 의 한계)

1. **외부 시스템 연동 분산**: `backend-core/internal/infrastructure/` (gitea, ci, commandworker, hrdb, serviceaction) + `backend-core/internal/integrations/adapters/` (homelab, task_item_puller, metrics) + `backend-ai/` (placeholder) — 동일 카테고리 ("외부 시스템 연동") 가 3 디렉터리에 분산
2. **AI agent / LLM context 부재**: 운영자가 자연어 query 못 던짐, integration 의 운영 history + monitoring + spec 이 context 로 retrieve 안 됨
3. **`backend-ai/` 의 dead state**: `main.py` 1 endpoint (`/health`) + TODO 2개, FastAPI + grpcio skeleton, v0.1.0 roadmap §1.2 "제외 기능 (v2 P3)" 분류 — 6+ release 누적
4. **AI agent 가 참조할 context 의 표준 부재**: 내부 knowledge (테이블 정의, metric 의미, API 변경 이력, runbook) 가 코멘트/구두/문서 산만

### 2.2 사용자 결정 (2026-06-17)

> "외부 시스템 연동 및 데이터 취합을 별도의 백엔드로 모을거야. 기존의 backend ai는 폐기하고 이걸로 통합하는거야. 기존 백엔드에서 외부 시스템 연동 관련 기능을 하던 부분도 여기에서 흡수해야해. 외부 시스템 연동과 데이터를 모두 취합해서 llm wiki 형태의 ai agent를 위한 도서관을 만들거야. google의 okf에 대해 조사하고 구조를 참조하도록 해. 기존 시스템과는 별도로 완전히 독립된 형태로 1차 개발을 진행하고, 개발 완료 후 기존 시스템과 연동하도록 하자."

> "이 시스템은 우리 기존 시스템의 백엔드를 참조하지 않고 외부 시스템만 바라보는 구조로 되어있어야해."

> "pi coding agent를 사용해볼거야."

> "인증은 외부 시스템 연결 시 설정할 수 있도록 하자. id, pw든, 토큰 방식이든."

### 2.3 결정 옵션 비교

| 옵션 | 변경 범위 | 비고 |
| --- | --- | --- |
| A. `backend-ai/` 확장 | 적음 — 기존 placeholder 확장 | 외부 연동 5종 흡수 어려움 (Go + Python 혼재, backend-ai 자체가 placeholder), OKF 형 concept library 불가 |
| B. `backend-core/internal/` 에 `knowledge/` 추가 | 중간 — 기존 backend-core 확장 | 다른 backend 연결 ❌ (standalone 정책) 과 모순. vendor-neutral 정책 정합 어려움 (backend-core 는 Go) |
| C. **신규 `backend-knowledge/` 디렉터리** | 큼 — 신규 + backend-ai 폐기 + source plugin 7종 재구현 (Gitea 4 sub-plugin + homelab + metrics + hrdb, 2026-06-17 A/A 결정) | vendor-neutral + standalone + OKF 정합 + Pi 정합 ✅ |

## 3. 결정

**신규 `backend-knowledge/` 디렉터리 신설. 외부 시스템 연동 + 데이터 취합 + OKF 형 concept library 의 단일 백엔드.**

### 3.1 위치 + tier

- 위치: `backend-knowledge/` (DevHub repo root, `backend-core/` / `backend-ai/` 옆)
- **Tier: 사외** (vendor-neutral + 다른 backend 연결 ❌ + 외부 시스템 only)
- **완전 standalone 시스템** (다른 backend (backend-core / 다른 백엔드 / 다른 시스템) 와의 어떤 layer 든 import / 호출 / 공유 ❌)
- **외부 시스템 7종 source 만 단방향** (M-v0.2.3 운영 기준, 2026-06-17 A/A 결정 + §6 supersession 정합): Gitea 4 sub-plugin (gitea_repo_pull / gitea_issue / gitea_wiki / gitea_action) + homelab + metrics + hrdb

### 3.2 기술 스택 (1차)

- **언어**: Python 3.13+ (OKF reference + LLM SDK 호환 + vendor-neutral)
- **HTTP framework**: FastAPI + Pydantic v2
- **영속성**: file system (OKF bundle) + sqlite (metadata index)
- **frontmatter**: PyYAML + python-frontmatter
- **markdown parse**: markdown-it-py + custom link extractor
- **LLM 호출 (M-v0.2.3+)**: Pi (pi.dev, [github.com/earendil-works/pi](https://github.com/earendil-works/pi), MIT, v0.79.6) — `pi-coding-agent` SDK or RPC mode (vendor-agnostic, 15+ provider via `pi-ai`)
- **observability**: structlog + OpenTelemetry
- **credential 관리**: 봉투 암호화 저장 (ADR-0025 정합), 연결 시 source plugin config 주입 (id/pw or token type-agnostic string)

### 3.3 3가지 기본 기능 (1차)

- **Ingest**: 외부 시스템 → 1차 raw concept (`POST /api/v0-2/ingest/{source}/sync`, `GET /ingest/{source}/status`)
- **Curate**: raw → OKF concept 정형화 (`POST /concepts/{id}/enrich` (M-v0.2.3+ Pi LLM), `PUT /concepts/{id}`, `POST /bundles/{bundle}/rebuild`)
- **Query**: 사용자 query → context + answer (`POST /query` (M-v0.2.3+ LLM 합성), `GET /concepts/{type}/{name}`, `GET /search`, `GET /bundles/{bundle}/index.md`, `GET /bundles/{bundle}/viz.html`)

### 3.4 1차 raw 데이터의 API 정책 (사용자 강조)

> "1차 raw 데이터는 여타 백엔드의 데이터들과 동일하게 api를 통해서 조회하고 추가할 수 있어야 해."

- `POST /api/v0-2/raw` + `GET /api/v0-2/raw/{type}/{name}` + `GET /api/v0-2/raw?source=...&since=...` (list) + `DELETE /api/v0-2/raw/{id}`
- 봉투 암호화 후 git 가능, 민감 source .gitignore 권장 (ADR-0025 정합)
- envelope: **자체 정의** (backend-core 의 `docs/api/conventions.md` 와 format 호환, **import ❌**)
- **internal-only, no auth** + **Path Y caller-provided user context (2026-06-18 결정, [release_v0-2_roadmap.md §3.6.1](../planning/release_v0-2_roadmap.md) 정합)**: bearer/API key 자체 검증 ❌. caller (gateway / 별도 agent) 가 `X-DevHub-User-Context` header 로 user/org/project/roles 전달 시, backend-knowledge 는 filter / curation ownership check 만 수행. **auth 책임 = caller (DevHub backend-core Keycloak federation 정합), governance 책임 = backend-knowledge**.
- **raw 운영 정책 (2026-06-18 신규, [release_v0-2_roadmap.md §4.4~§4.7](../planning/release_v0-2_roadmap.md) 정합)**: 봉투 암호화 format (`$env$v0.1$...` ADR-0025) + .gitignore per source 정책 + retention default 90일 + storage quota 1GB/bundle + endpoint 별 권한 (POST = bundle owner_org member / GET = visibility 정합 / DELETE = system_admin OR 등록자 OR owner_org member) + 1 raw → N concepts 관계 + raw 삭제 시 concept 처리 (M-v0.2.0 = hard_delete / M-v0.2.1+ = soft_archive) + sha256 정합성 검증 + audit log 7 event.

### 3.5 운영 환경 (standalone 정합)

- **별도 dev script**: `backend-knowledge/dev-up.sh` (backend-knowledge 만 단독)
- **별도 docker-compose**: `backend-knowledge/docker-compose.yml` (backend-knowledge 서비스만)
- devhub root 의 `dev-up.sh` / `docker-compose.colima.yml` / `docker-compose.deploy.yml` 사용 ❌
- root level 의 `backend-ai/` 디렉터리 / `backend-ai` service / `backend-ai` reference 정리 (M-v0.2.2)

### 3.6 frontend 정책

- **viz.html (자가 viewer)**: 모든 Phase 공통 (backend-knowledge 가 SSR, **§12.1 상세 정합**: Cytoscape.js + marked.js CDN embed + 4 edge type + 8 type node 색상 + static HTML)
- **frontend 관리/조회 page 1**: M-v0.2.1+ 부터, `backend-knowledge/web/` (별도 standalone frontend, devhub frontend 와 분리, §1.2 G7 standalone 정책 정합, **§12.2 의 5 page 상세 정합**: concept list / concept detail / ingest trigger / bundle management / raw inspector + §12.3 user flow 3 role + §12.4 API integration matrix + §12.5 cutover 정책)
- **frontend 0 page**: M-v0.2.0 (1차 backend 단독)
- **§13 cross-cutting 종합 영향**: umbrella doc 전체 cross-reference 정합성 최종 검토 (§13.1 matrix 20 row + §13.2 gap 6 row + §13.3 post-sprint follow-up 6 row + §13.4 정합 검증 결과) 의 frontend 영향 — frontend 단독 frontend 기술 선택 (vanilla JS / Next.js / Vue.js / Svelte, §12.2) + frontend cutover 정책 (§12.5) + frontend update 주기 (per release) + standalone frontend 운영 부담 (§1.2 G7 정합)

### 3.7 Keycloak 분류 (재확인)

- Keycloak = 사내 IdP (DevHub 가 authenticate 받는 곳), 외부 시스템 아님
- `backend-core/internal/domain/auth-session/` 의 책임 (ADR-0019 정합)
- 본 시스템 scope 외 — OIDC ❌, Keycloak ❌, backend-core 인증 위임 ❌

### 3.8 마일스톤

| ID | 마일스톤 | scope | 의존 | status |
| --- | --- | --- | --- | --- |
| **M-v0.2.0-alpha** | 컨셉 umbrella + child doc 정합 + PoC 진입 | 본 문서 publish + `external-integrations-agentic-rag-roadmap.md` cross-link | (없음) | ⏳ planned (v0.2.0-alpha) |
| **M-v0.2.0** | 1차 standalone 구현 (Gitea 통합 4 sub-plugin PoC, backend 단독, frontend 0, 2026-06-17 A/A 결정) | `backend-knowledge/` skeleton + OKF spec model (frontmatter `x_devhub_category` 필드 추가) + **Gitea 통합 4 sub-plugin** (gitea_repo_pull / gitea_issue / gitea_wiki / gitea_action, Gitea 1 instance 의 4 카테고리 = 이슈/위키/SCM/CI-CD) + `homelab_mock` 1종 = **5종 PoC** + Ingest 1 endpoint + Query 1 endpoint + 1차 raw API + OpenAPI | M-v0.2.0-alpha | ⏳ planned (v0.2.0) |
| **M-v0.2.1** | 1차 완성 + Gitea 4 정식 + homelab real + Curate + frontend 관리/조회 page 1 | Gitea 통합 4 sub-plugin 정식 wire (1차 PoC → 정식) + `homelab_mock` → `homelab` (real wire, 5 카테고리 외 사내 시스템) = **5종 운영** + Curate 3 endpoint + 1차 viz.html (자가 viewer) + frontend 관리/조회 page 1 (`backend-knowledge/web/`, devhub frontend 와 분리) + e2e smoke | M-v0.2.0 | ⏳ planned (v0.2.1) |
| **M-v0.2.2** | 외부 시스템 6종 source wire + backend-ai 폐기 | 외부 시스템 **6종** source plugin wire (**Gitea 4 sub-plugin** + homelab + metrics) + `backend-ai/` 디렉터리 제거 | M-v0.2.1 | ⏳ planned (v0.2.2) |
| **M-v0.2.3** | hrdb 운영 + Pi LLM enrich + cross-link 자동 resolution | + hrdb = **7종 운영** + Pi `pi-coding-agent` SDK or RPC mode 로 LLM enrich 활성화 + cross-link 자동 resolution | M-v0.2.2 | ⏳ planned (v0.2.3) |
| **M-v0.3.0** | 풀 RAG | chunking + embedding + vector index + retrieval + reranking | M-v0.2.3 | ⏳ planned (v0.3.0) |

**§5 정합** (2026-06-18 추가, [release_v0-2_roadmap.md §5.4~§5.7](../planning/release_v0-2_roadmap.md)): 본 §3.8 표의 각 마일스톤 scope 가 §5.5 per-milestone DoD 의 (a) 코드/문서 column 과 1:1 정합. M-v0.2.0 의 (b) 검증 = §5.5 의 pytest + e2e smoke. M-v0.2.1 의 frontend page 1 = §5.7 parallel sprint PR (3). M-v0.2.3 의 Pi LLM enrich = §5.7 PR (2).

## 4. 결과

### 4.1 positive

- **외부 시스템 데이터의 AI Agent Library** (Markdown + YAML + cross-link, vendor-neutral)
- **7 source plugin 의 1차 5종 PoC wire** (M-v0.2.0 Gitea 4 sub-plugin + homelab_mock) → **7종 전부 wire** (M-v0.2.3, Gitea 4 + homelab + metrics + hrdb)
- **3가지 기본 기능** (Ingest / Curate / Query) + 1차 raw API
- **standalone 운영** (다른 backend 연결 ❌, 외부 시스템 only)
- **vendor-neutral** (Pi, OpenAI, Anthropic, Gemini 등 15+ provider 정합)
- **`backend-ai/` 폐기** (placeholder 정리, M-v0.2.2)
- **기존 외부 연동 5종 source 의 Go adapter 와 무관하게 0에서 Python 작성** (외부 시스템 API spec 만 참조, backend-core 의 어떤 layer 든 참조 ❌)

### 4.2 negative / trade-off

- **신규 디렉터리** + Python 추가 (backend-core 는 Go) — 운영 부담 (별도 dev script, 별도 docker-compose)
- **M-v0.2.0~v0.2.2 frontend 0 page** (관리/조회 page 는 M-v0.2.1+ 부터, viz.html 자가 viewer 만 1차) — 운영자 직접 backend-knowledge 의 API 호출 필요
- **7 source plugin 의 Python 재구현** (Go 코드 이전 ❌, 로직 재작성) — 작업 부담
- **standalone 정책** 으로 backend-core 와 직접 wire ❌ — gateway / cron / 별도 agent 가 cross-repository 호출 필요 (운영 부담)
- **OIDC 제외** — `/api/v0-2/*` 전체 인증 없이 호출 가능. 별도 gateway / firewall / IP allowlist 가 보호 (운영 책임)
- **Path Y caller-provided user context (2026-06-18 결정, [release_v0-2_roadmap.md §3.6](../planning/release_v0-2_roadmap.md) 추가)** — auth 자체 ❌, caller 가 user context 전달 시 governance 만 수행. **governance 책임 = backend-knowledge** (curation ownership + query scope filter) + **auth 책임 = caller** (DevHub backend-core Keycloak federation 정합). 운영 부담: caller (gateway / agent) 가 Keycloak 인증 + user context 구성 + backend-knowledge 호출 의 3-step orchestration 필요
- **§7 Q12~Q18 결정 (2026-06-18, [release_v0-2_roadmap.md §7](../planning/release_v0-2_roadmap.md))** — Q12 raw storage_mode dual mode (file|db per source_meta.storage_mode) + Q13 Pi SDK mode M-v0.2.0+ timing + Q14 DB type (sqlite M-v0.2.0~v0.2.2 + PostgreSQL M-v0.2.3+) + Q15 per source default mapping + Q16 cron interval `*/5 * * * *` + Q17 Pi LLM fallback to rule-based + Q18 backend-ai 폐기 M-v0.2.2 동시. §11 운영 runbook 영향: Q16 cron interval = §11.3 monitoring #4 Pi ingest success rate 측정 주기 정합 / Q17 fallback = §11.1.3 Pi ingest pipeline timeout-degraded runbook 정합 / Q18 backend-ai 폐기 = §6.6.2 폐기 절차 10 단계 정합. **운영 부담: Q12 dual mode + Q14 DB 추가 = 운영 복잡도 증가 (per source storage_mode 별 관리) + DB backup 추가 (§11.2)**

### 4.3 영향

- `backend-ai/` 디렉터리 제거 (M-v0.2.2, placeholder 정리)
- `backend-core/internal/integrations/adapters/` 의 5종 source 의 backend-knowledge 의 source plugin 으로 점진적 흡수 (M-v0.2.2~v0.2.3) — backend-core 측의 Go adapter 제거는 별도 PR (backend-knowledge 책임 아님)
- **Path Y caller-provided user context (2026-06-18 신규)** — [release_v0-2_roadmap.md §3.6](../planning/release_v0-2_roadmap.md) 에 5 subsection 정의 (§3.6.1 schema+trust model / §3.6.2 curation governance / §3.6.3 query scope priority 4-tier / §3.6.4 frontmatter 5 governance field / §3.6.5 cross-section 정합 fix 7 위치). concept organization 의 §3.5 와 §3.6 의 조합으로 backend-knowledge 의 운영 정공법 완성 (1 concept = 1 .md + 5 카테고리 + 8 type + 5 governance field + 4-tier query scope priority)
- 사내 한정 tier (`Makefile`, `docker-compose.deploy.yml`, `docker-compose.colima.yml`, root `dev-up.sh`) 의 backend-ai reference 정리
- 사외 tier (이 ADR + release_v0-2_roadmap.md + 신규 `backend-knowledge/` + 신규 `backend-knowledge/web/`) 신규
- `external-integrations-agentic-rag-roadmap.md` status draft → active 전환 (Q7 결정)

## 5. 후속 작업 (M-v0.2.0 sprint 진입 checklist)

- [ ] 본 문서 (release_v0-2_roadmap.md) publish + cross-link 정합
- [ ] `external-integrations-agentic-rag-roadmap.md` §8 변경 이력에 row 추가 (umbrella doc publish) + status draft → active 전환
- [ ] `ai-workflow/memory/state.json` M-v0.2.0 row 발급 (또는 v0.1.x status update)
- [ ] `backend-knowledge/` 디렉터리 skeleton (Dockerfile, pyproject.toml, main.py, okf/spec.py)
- [ ] OKF `SPEC.md` 1차 정독 (vendor-neutral 정책 + frontmatter 정확한 spec 확인) + [ADR-0034](./0034-okf-adoption.md) §5 정합
- [ ] 신규 GitHub milestone `v0.2.0` 생성 + 본 ADR-0035 + release_v0-2_roadmap.md link 첨부

## 6. Supersession / 변경 이력 (2026-06-18)

**2026-06-17 A/A 결정 (5 카테고리 + Gitea 통합 1차 wire + `x_devhub_category` 필드)** 으로 본 ADR-0035 의 §3.1 / §3.8 / §4.1 / §4.2 가 정합 수정. 결정 자체 (신규 `backend-knowledge/` 디렉터리 신설, standalone 정책, 다른 backend 연결 ❌, OIDC 제외, Pi 채택) 는 변경 없음. 변경된 부분만:

| 위치 | 변경 전 | 변경 후 |
| --- | --- | --- |
| §2.3 옵션 C | source plugin 5종 재구현 | source plugin **7종** 재구현 (Gitea 4 sub-plugin + homelab + metrics + hrdb) |
| §3.1 외부 시스템 단방향 | 외부 시스템 5종 source (homelab / gitea / hrdb / prometheus / task_item_puller) | 외부 시스템 **7종** source (Gitea 4 sub-plugin gitea_repo_pull / gitea_issue / gitea_wiki / gitea_action + homelab + metrics + hrdb, M-v0.2.3 운영 기준) |
| §3.8 M-v0.2.0 | 1 source plugin (homelab + homelab_mock) | **Gitea 통합 4 sub-plugin** (gitea_repo_pull / gitea_issue / gitea_wiki / gitea_action) + `homelab_mock` 1종 = **5종 PoC** + frontmatter `x_devhub_category` 필드 추가 |
| §3.8 M-v0.2.1 | + 1 source plugin 추가 (총 2) | Gitea 4 정식 + `homelab_mock` → `homelab` (real) = **5종 운영** + frontend 관리/조회 page 1 |
| §3.8 M-v0.2.2 | 외부 시스템 5종 source wire (homelab + gitea_pull + metrics + task_item_puller + hrdb) | 외부 시스템 **6종** source wire (**Gitea 4 sub-plugin** + homelab + metrics) + `backend-ai/` 폐기 |
| §3.8 M-v0.2.3 | + Pi LLM enrich + cross-link 자동 resolution | + hrdb = **7종 운영** + Pi LLM enrich + cross-link 자동 resolution |
| §4.1 positive | 5 source plugin 의 1차 1~2종 wire (M-v0.2.0 homelab + homelab_mock) → 5종 전부 wire (M-v0.2.3) | **7 source plugin 의 1차 5종 PoC wire** (M-v0.2.0 Gitea 4 sub-plugin + homelab_mock) → **7종 전부 wire** (M-v0.2.3) |
| §4.2 negative | 5 source plugin 의 Python 재구현 | **7 source plugin** 의 Python 재구현 |

**supersession 정공법**: 본 ADR-0035 의 결정 자체 (신규 `backend-knowledge/` 디렉터리 신설 + standalone 정책 + OIDC 제외 + Pi 채택) 는 변경 ❌. source plugin 의 구체적 구성 (5종 → 7종) + M-v0.2.0 scope (homelab + homelab_mock → Gitea 4 + homelab_mock) 만 정합. [`release_v0-2_roadmap.md` §9 변경 이력 2026-06-18 row](../planning/release_v0-2_roadmap.md) 가 cross-section 정합 fix 의 source of truth.

