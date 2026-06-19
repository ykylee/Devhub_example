# ADR-0035: `backend-knowledge` 백엔드 신설 (외부 시스템 연동 + OKF concept library)

## 1. 상태

- **상태**: Accepted
- **작성일**: 2026-06-17
- **수정일**: 2026-06-18 (umbrella doc §3.6 Path Y caller-provided user context + §4.4~§4.7 raw 운영/API/정합성 정책 + **§1.1 한계 4~7 추가 (2026-06-18 결정에서 식별된 4가지 trade-off 한계 = Path Y trust model / dual mode 운영 / backup DR / frontend lifecycle) + §1.3 How 정당화 강화 (한계 7개 → §3~§12 해결책 cross-reference 표 7 row) + §3.5.6 cross-link reverse index 정공법 (M-v0.2.0 PoC 부터 능동적 강화, §13.2 known gap 1 ✅ resolved) + §2.4 standalone 검증 매트릭스 (10 row 검증 항목 + 운영자 onboarding SOP + 자동화 tool) + §14 M-v0.2.0 release notes draft (§13.3 #5 ✅ partial resolved) + §8 timeline 보강 (§8.1 17 commit 결정 timeline + §8.2 cross-reference 매트릭스 + §8.3 향후 결정 row 10 row + §8.4 4 layer 정합) + §2.6 backend-knowledge network 정책 (5 subsection + dev/staging/production 3 단계 + docker-compose networks + iptables + WAF 10 rule + 8 row 자동화 tool) + §15 ADR supersession 정공법 (M-v0.2.3+ 부터, 5 step + deprecation policy + release notes 정합)** 신규에 따른 §3.4 1차 raw API 정책 row 갱신 — internal-only no auth + caller-provided user context + raw 운영 정책 + endpoint 권한 + 정합성 검증 명시 + 한계 4~7 의 §1.1 명시 + §1.3 정당화 cross-reference 추가, §4.3 영향 row 갱신 + §3.5.6 cross-link reverse index 정공법 row 추가 (3 graph endpoint + `okf/link_graph.py reverse_index()` + impact-based archive 거부 정책 + viz.html incoming edge visualization + 7 cross-section fix 위치) — backend-knowledge 의 4번째 핵심 기능 (Ingest/Curate/Query/**Graph**) + §2.4 standalone 검증 매트릭스 row 추가 (10 row 검증 항목: network 격리 / port expose / env var / import / API 호출 / DB / cron worker / monitoring / log / artifact + 운영자 onboarding SOP + 자동화 tool `scripts/check_standalone_drift.sh`, §3.5 운영 환경 (standalone 정합) 의 구체적 검증 정공법) + §14 M-v0.2.0 release notes draft row 추가 (7 subsection: highlight / 16 commit summary / breaking change 4 row / per-source plugin 7종 / per-milestone 5 M / §13 정합 / template / contributor, §13.3 #5 ✅ partial resolved, M-v0.2.0 release 시점에 `docs/release-notes/v0.2.0.md` 로 post-process) + §8 timeline 보강 row 추가 (4 subsection: §8.1 17 commit 결정 timeline + §8.2 cross-reference 매트릭스 17 commit × 5 artifacts + §8.3 향후 결정 row 10 row + §8.4 4 layer 정합 L1~L4, 10/17 commit 영향 = 59%) + §2.6 backend-knowledge network 정책 row 추가 (5 subsection: §2.6.1 3 단계 network 정책 + §2.6.2 docker-compose networks + §2.6.3 iptables + §2.6.4 WAF 10 rule + §2.6.5 검증 절차 정밀화, §2.4 item 1 + §6.5.3 + §11.1.1 정합, 4 cross-section fix 위치, 사외/사내 2-tier 정합) + **§15 ADR supersession 정공법 row 추가 (6 subsection: §15.1 정의 + 사용 시나리오 4 종 + §15.2 5 step 정공법 + §15.3 row format + §15.4 cross-reference 4~5 file + §15.5 deprecation policy 12개월 + §15.6 umbrella doc §13~§15 cross-cutting 정공법 3 종 정합, M-v0.2.3+ 부터 supersession 가능, 본 ADR-0035 §6 Supersession section row + ADR-0034 §6 Supersession section 신규 추가 + docs/governance/worker_division.md §4.2 1:1 정합)** + **§3.5.7 Pi LLM cross-link 자동 resolution 정공법 row 추가 (5 subsection: §3.5.7.1 목적 + §3.5.7.2 j2 prompt template + §3.5.7.3 SDK/RPC mode + §3.5.7.4 3 mode confirm workflow + §3.5.7.5 audit log + 5 metrics, M-v0.2.3+ 부터 활성화, §3.5.6.4 auto-fix strategy 구현 + §13.2 known gap 2 ✅ resolved, 5 cross-section fix 위치)** + **§16 API versioning 정책 row 추가 (6 subsection: §16.1 API versioning 정의 + /api/v0-2/ prefix 의의 + §16.2 deprecation policy 12개월 + dual endpoint + §16.3 Sunset/Deprecation header + §16.4 monitoring 2개 버전 4 metrics + §16.5 breaking change 정의 5 종 + release notes + §16.6 §3.1 API 매트릭스 future deprecation timing, M-v0.3.0+ 부터 v0-3 도입, 5 cross-section fix 위치)** + **§3.5.8 Pi LLM resolution false positive rollback 정공법 row 추가 (4 subsection: §3.5.8.1 false positive 정의 + 4 종 시나리오 + 발생 빈도 / §3.5.8.2 rollback trigger 3 종 + 5분 pending undo + audit log + impact analysis snapshot / §3.5.8.3 operator notification 4 channel + alert routing M3 threshold 5%/10%/20% / §3.5.8.4 recovery workflow 5 step ≤ 5분, M-v0.2.3+ 부터 활성화, §3.5.7.4 auto-apply safety net + §3.5.7.5 M3 능동적 강화, 4 cross-section fix 위치)** + **§3.6.6 governance audit log 정공법 row 추가 (5 sub-subsection: §3.6.6.1 audit log event 7 row + §3.6.6.2 curation 권한 추적 5 curator_type + §3.6.6.3 governance dashboard 13 metric + §3.6.6.4 data lineage 4 단계 + §3.6.6.5 GDPR/PII compliance, Path Y audit + curation 권한 추적 + governance dashboard, 2026-06-18 신규, M-v0.2.0 PoC = audit log + M-v0.2.1+ dashboard + M-v0.2.3+ GDPR/PII + data lineage, 5 cross-section fix 위치)** + **§3.7.6 data normalization pipeline 자동 검증 정공법 row 추가 (5 sub-subsection: §3.7.6.1 Partial failure detection + §3.7.6.2 Schema drift detection + §3.7.6.3 Duplicate concept detection + §3.7.6.4 Large raw handling + §3.7.6.5 Auth failure monitoring, M-v0.2.0 PoC 운영 시점 활성화, 2026-06-18 신규, §3.7.5 edge cases 6 종 자동 detection, 9 cross-section fix 위치)** + **§17 cross-cutting 종합 review row 추가 (9 sub-subsection: §17.0 high-level 종합 + §17.1 25 commit × cross-section 매트릭스 25 row + §17.2 known gaps 4 row 자연 해소 시점 + §17.3 post-sprint 5 row 우선순위 + §17.4 18/18 Q&A 정합 + §17.5 28 metrics 정합 + §17.6 4 cross-cutting 정공법 검증 + §17.7 19 main section cross-ref 매트릭스 + §17.8 stats, umbrella doc 25 commit 후 cross-cutting 최종 종합, 2026-06-18 신규, 9 cross-section fix 위치)** + **§17.8 stats 정정 row 추가 (4 row 갱신: §17.0 line 5050 + §17.8 line 5288 + §17.8 line 5306 + §13.4 §17 row line 4471, §13.4 row 15 → §13.4 row 23 = 12 prefix + 11 non-prefix, 2026-06-18 minor issue 즉시 fix, 6 cross-section fix 위치)** )
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
- **상세 검증 정공법**: [release_v0-2_roadmap.md §2.4](../planning/release_v0-2_roadmap.md) standalone 검증 매트릭스 (10 row 검증 항목: network 격리 / port expose / env var / import / API 호출 / DB / cron worker / monitoring / log / artifact) + 운영자 onboarding SOP (sprint 진입 시 10 row PASS 검증 + 결과 문서 작성) + 자동화 tool `scripts/check_standalone_drift.sh` (M-v0.2.1+ CI pre-merge)

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
- **§1.1 한계 4~7 추가 + §1.3 How 정당화 강화 (2026-06-18 신규)** — 한계 4 (caller-provided user context 신뢰, Path Y 의 trust model) / 한계 5 (dual storage mode 운영 복잡도) / 한계 6 (backup DR transactional 정합성) / 한계 7 (frontend standalone 유지보수 부담) 의 4가지를 v0.2.0 PoC 의 trade-off 로 명시적 식별. §1.3 에 한계 7개 (1~3 = v0.1.x 한계 + 4~7 = 2026-06-18 식별 trade-off) → §3~§12 해결책 cross-reference 표 7 row. 본 backend-knowledge 의 정책 (§3.6 / §10 / §11 / §12) 이 한계 4~7 의 mitigation 기반. 능동적 강화 (HMAC signature / CLI tool / transactional backup / CI contract test) 는 M-v0.2.1+~M-v0.3.0+ scope 외, §1.1 한계 식별 + §13.2 known gaps 와 정합. §12 frontend 정책 (코드 공유 ❌, viz.html 자가 viewer + M-v0.2.1+ frontend 5 page) 이 한계 7 의 baseline mitigation, 후속 sprint 의 contract test 추가 가 본 한계의 능동적 강화.
- **§3.5.6 cross-link reverse index 정공법 (M-v0.2.0 PoC 부터 능동적 강화, 2026-06-18 신규 — §13.2 known gap 1 ✅ resolved)** — §3.5.5 cross-link 4종 rule 의 4번째 (reverse index) 의 구현 정공법. 5 subsection + 3 graph endpoint (`GET /api/v0-2/graph/reverse/{path}` + `GET /api/v0-2/graph/impact/{path}` + `POST /api/v0-2/graph/reindex`) + `var/bundles/.index/reverse_index.json` schema + `okf/link_graph.py reverse_index()` implementation + 3 strategy stale handling + impact analysis 기반 archive 거부 정책 (inlink_count >= 1 → 409 Conflict + soft archive 권장) + viz.html incoming edge visualization + 4 CLI tool M-v0.2.1+. §3.3.4 의 신규 backend-knowledge 핵심 기능 4종 (Ingest / Curate / Query / **Graph**) 중 4번째 기능. cross-section 정합 fix 7 위치 (§3.5.5 / §2.1 / §3.1 / §3.9.4 / §6.5.4 / §11.1.7 / §13.2-§13.4). 본 §3.5.6 정공법은 backend-knowledge 의 day-2 운영 (§11.1.7 stale link runbook) + frontend 운영 (§12.1 viz.html incoming edge + §12.4 API matrix 3 row 추가) 의 기반.
- **§2.4 standalone 검증 매트릭스 (2026-06-18 신규)** — §1.2 G7 + §3.5 의 standalone 정책 (다른 backend 연결 ❌, OIDC ❌, 외부 시스템 only 단방향, caller-provided user context) 의 **구체적 검증 정공법**. 10 row 검증 매트릭스 (network 격리 / port expose / env var / import / API 호출 / DB / cron worker / monitoring / log / artifact) + per 항목 검증 방법 + PASS 기준 + FAIL 시 mitigation + 자동화 (M-v0.2.0 = CI grep, M-v0.2.1+ = `scripts/check_standalone_drift.sh` pre-merge) + 운영자 onboarding SOP (sprint 진입 시 10 row PASS 검증 + 결과 문서 작성). cross-section 정합 fix 3 위치 (§1.2 G7 cross-reference / §6.5.1 docker-compose standalone 정합 검증 cross-reference / §11.4 on-call Operator training cross-reference). §3.5 운영 환경 (standalone 정합) row 에 상세 검증 정공법 cross-reference 추가. 본 §2.4 매트릭스는 운영자가 sprint 진입 시점에 PASS 검증 필수 + PR review 시 contributor self-check (PR template 의 `affects-standalone` field 정합 권장, M-v0.2.1+ 도입 검토) + M-v0.2.3+ production 분기 1회 audit.
- **§14 M-v0.2.0 release notes draft (2026-06-18 신규 — §13.3 #5 ✅ partial resolved)** — umbrella doc 본문 release notes draft (M-v0.2.0 release 시점에 `docs/release-notes/v0.2.0.md` 로 copy + post-process). 7 subsection: §14.1 highlight 7-10 bullet (신규 백엔드 / OKF v0.1 / 5 카테고리 / Path Y / DB-based raw / 운영 runbook / 한계 7개 / backend-ai 폐기 / 18/18 결정 / 10 row 매트릭스) / §14.2 16 commit summary (highlight + 영향 section) / §14.3 breaking change 4 row (`backend-ai/` 폐기 M-v0.2.2 + Go adapter 흡수 M-v0.2.2~v0.2.3 + Tier 분리 정책 + `x_devhub_curator` curation governance 정책) / §14.4 per-source plugin 7종 (Gitea 4 + homelab + homelab_mock + metrics + hrdb, per M-v0.2.0~v0.2.3 + per storage_mode × normalize_mode) / §14.5 per-milestone 5 M (M-v0.2.0 PoC = 본 release, M-v0.2.1/2.2/2.3/M-v0.3.0+) / §14.6 §13 cross-cutting 정합 + post-sprint follow-up 6 row (1 resolved + 5 자연 해소, ⚠️ 5 row release 직전 처리) / §14.7 release notes template per backend-knowledge (frontmatter + 7 section + post-process 10 step) / §14.8 contributor placeholder (Sisyphus + 사용자, release 시점에 자동화). cross-section 정합 fix 3 위치: §13.3 #5 partial resolved 갱신 + §13.4 정합 검증 row 추가 (release notes) + umbrella doc frontmatter 갱신.
- **§8 timeline 보강 (2026-06-18 신규 — §13.1 cross-reference matrix + §14.2 16 commit summary 정합)** — §8 의 high-level 결정 19 row (Q1~Q18) → §8.1 17 commit 결정 timeline (per commit 의 concept change 1줄 + 영향 section + cross-reference) → §8.2 cross-reference 매트릭스 (17 commit × 5 artifacts: ADR-0034 14/17, **ADR-0035 10/17**, state.json 3/17, external-integrations-agentic-rag-roadmap.md 2/17, docs/llm-wiki mirror 17/17 = 100%) → §8.3 향후 결정 row 10 row (Q-N1~Q-N6 sprint 진입 시점 2026-06-19 + Q-F1~Q-F4 후속 sprint 2026-07-01~09-01) → §8.4 결정 timeline 의 4 layer 정합 (L1 high-level 결정 / L2 commit 결정 / L3 cross-reference / L4 향후 결정). 본 backend-knowledge 의 정책 (§3.6 / §10 / §11 / §12 / §3.5.6 / §2.4 / §14) 이 §8.1 의 17 commit 결정 row 의 영향 section 과 1:1 정합. cross-section 정합 fix 3 위치: §13.4 정합 검증 row 추가 (timeline 보강) + §13.1 cross-reference matrix 정합 + umbrella doc frontmatter 갱신.
- **§2.6 backend-knowledge 운영 환경의 network 정책 (2026-06-18 신규 — §2.4 item 1 + §6.5.3 + §11.1.1 정합)** — §2.4 매트릭스 item 1 (network 격리) 의 **구체적 정공법**. 5 subsection: §2.6.1 3 단계 network 정책 (dev = localhost / staging = VPN+사내 CA / production = WAF+외부 CA) / §2.6.2 docker-compose.yml networks 설정 정공법 (3 단계 별 YAML, `internal: true` flag 활용) / §2.6.3 firewall iptables rule 예시 (production, INPUT/OUTPUT chain + source plugin source_url rate limit + Docker iptables chain interaction) / §2.6.4 WAF 설정 (Cloudflare / AWS WAF / nginx mod_security 3 option + 10 row WAF rules) / §2.6.5 §2.4 item 1 검증 절차 정밀화 (8 row 자동화 tool + 운영자 manual SOP + per release audit + incident runbook 정합). 본 backend-knowledge 의 §3.5 운영 환경 standalone 정합 (다른 backend 연결 ❌) 의 **구체적 network 정공법** = §2.6. 본 §2.6 의 사외/사내 2-tier 정합 (AGENTS.md §사외/사내 2-tier 형상관리 분리, 2026-06-10 결정) = dev = 사외 / staging + production = 사내. cross-section 정합 fix 4 위치: §2.4 item 1 "상세 정공법" cross-reference + §6.5.3 "상세 정공법" cross-reference + §11.1.1 "Network 진단" 4 row + §13.4 정합 검증 row 추가.
- **§3.5.7 Pi LLM cross-link 자동 resolution 정공법 (M-v0.2.3+ 부터, 2026-06-18 신규 — §3.5.6.4 auto-fix strategy 구현 + §13.2 known gap 2 ✅ resolved)** — backend-knowledge 의 day-2 운영 정공법의 cross-link 자동 resolution 부분. 5 subsection: §3.5.7.1 목적 (unresolved link 자동 recommend + operator confirm + §3.5.6.4 auto-fix strategy 구현) / §3.5.7.2 j2 prompt template design (input unresolved link context ±2 lines + output 3 row recommendation + reason + confidence 0.5) / §3.5.7.3 SDK/RPC mode 선택 §10.3 정합 (M-v0.2.3+ default SDK mode + production RPC mode option) / §3.5.7.4 3 mode confirm workflow (dry-run/confirm/auto-apply ≥ 0.9) + `POST /api/v0-2/concepts/{id}/resolve-links?mode={dry-run|confirm|auto-apply}&selected_rank={1|2|3}&confidence_threshold=0.9` endpoint / §3.5.7.5 audit log + 5 metrics (MTTR < 30분 / accuracy ≥ 70% / false positive ≤ 5% / pi_sdk_timeout ≤ 1% / pi_llm_recommendation_count 일 ≤ 50) + `cli/fix_unresolved.py` 4 CLI tool. 본 backend-knowledge 의 정책 (§3.5.5 cross-link 4종 rule + §3.5.6.4 stale handling 3 strategy + §3.9.4 archive 거부 정책) 이 §3.5.7 정공법 기반. cross-section 정합 fix 5 위치: §3.5.6.4 auto-fix row + §3.1 API 매트릭스 endpoint + §6.7.3 LLM enrich 운영 + §10.3 Pi prompt template row 갱신 + §13.2 known gap 2 ✅ resolved (1/6 → 2/6 resolved, residual 4/6).
- **§16 API versioning 정책 (M-v0.3.0+ 부터 v0-3 도입, 2026-06-18 신규 — §3.1 API 매트릭스 향후 호환성 + §14.7 release notes template breaking change + §15.5 deprecation policy 정합)** — backend-knowledge 의 API 향후 호환성 정공법. `/api/v0-2/` prefix 의 의도 + M-v0.3.0+ 부터 `/api/v0-3/` prefix 도입 시 마이그레이션. 6 subsection: §16.1 API versioning 정의 (URL prefix 기반 semver + v0.x pre-1.0 + 12개월 deprecation) / §16.2 deprecation policy 12개월 + dual endpoint support (M-v0.3.0 release 시 /api/v0-2/ + /api/v0-3/ 동시 운영 + 6개월 warning + 12개월 제거 + client migration SOP 5 step) / §16.3 API gateway deprecation header (Sunset RFC 8594 + Deprecation + Link successor-version) / §16.4 monitoring 2개 버전 동시 운영 4 metrics (per endpoint request count + error rate + client identification + migration progress) + §11.3 monitoring 5 + 4 = 9 metrics / §16.5 breaking change 정의 5 종 (a) path 변경 / (b) method 변경 / (c) schema 변경 / (d) auth 변경 / (e) default 변경 + release notes 정합 §14.7 + §14.3 / §16.6 §3.1 API 매트릭스 versioning 영향 + future deprecation timing (M-v0.2.0~v0.3.0 / M-v0.3.0+ deprecation / M-v0.3.0+ 제거) + 운영 runbook 영향 (§11.1 incident + §11.3 monitoring + §11.4 on-call role API curator 6번째). 본 backend-knowledge 의 정책 (§3.1 API 매트릭스 14 endpoint + §6.5.3 gateway middleware + §11.3 monitoring 9 metrics) 이 §16 정공법 기반. cross-section 정합 fix 5 위치: §13.4 정합 검증 row 추가 (API versioning) / §3.1 API 매트릭스 future deprecation timing / §14.7 release notes template breaking change / §15.5 deprecation policy 12개월 1:1 정합 / §11.3 monitoring 9 metrics (5+4).
- **§3.5.8 Pi LLM resolution false positive rollback 정공법 (M-v0.2.3+ 부터, 2026-06-18 신규 — §3.5.7.4 auto-apply safety net + §3.5.7.5 M3 능동적 강화)** — backend-knowledge 의 day-2 운영 정공법의 cross-link 자동 resolution safety net. 4 subsection: §3.5.8.1 false positive 정의 + 4 종 시나리오 (typo 매칭 / renamed target / self-reference / cycle) + 발생 빈도 target ≤ 5% (M-v0.2.3+ PoC 예상 ~3.6%, production ~5.6% estimated) / §3.5.8.2 rollback trigger 3 종 (operator manual undo / 5분 내 impact analysis detection / 24시간 monitoring flag) + 5분 pending undo 상태 4 종 (`x_devhub_status`: `applied_pending` / `applied` / `rolled_back` / `rolled_back_late`) + audit log `pi_link_resolve.rolled_back` event + impact analysis snapshot / §3.5.8.3 operator notification 4 channel (Slack / email / dashboard banner / §11.3 monitoring alert) + alert routing M3 threshold 5%/10%/20% / §3.5.8.4 recovery workflow 5 step ≤ 5분 (`cli/revert_unresolved.py` + `POST /graph/reindex` + `POST /bundles/{bundle}/rebuild` + viz.html + audit log) + §3.5.6.4 stale handling 3 strategy 정합. 본 backend-knowledge 의 정책 (§3.5.5 cross-link 4종 rule + §3.5.6.4 stale handling 3 strategy + §3.5.7.4 auto-apply + §3.9.4 archive 거부 정책) 이 §3.5.8 정공법 기반. cross-section 정합 fix 4 위치: §3.5.7.4 auto-apply row + §3.5.7.5 M3 row + §13.4 정합 검증 row 추가 (false positive rollback) / §11.3 monitoring 10 metrics (5 + §3.5.7.5 5) / §12 frontend page admin dashboard banner.
- **§3.6.6 governance audit log 정공법 (Path Y audit + curation 권한 추적 + governance dashboard, 2026-06-18 신규 — §3.6.1~§3.6.5 + §3.9 + §11.3 + §16 정합)** — backend-knowledge 의 §3.6 Path Y data governance 의 audit log + curation 권한 추적 + governance dashboard 정공법. 5 sub-subsection: §3.6.6.1 audit log event 7 row (user login / concept access / curation edit / query / archive / publish / config change + JSON Lines format + daily file rotation) / §3.6.6.2 curation 권한 추적 5 curator_type (rule-based / llm-system_admin / human-self-user / human-org-head / human-system-admin) + change history per frontmatter field + curation ownership verification / §3.6.6.3 governance dashboard 13 metric per user/org/project/event type + audit log viewer 4 endpoint + alert routing 5 row / §3.6.6.4 data lineage 4 단계 source plugin → raw → concept → cross-link + lineage graph query + MTTR < 30분 / §3.6.6.5 GDPR/PII compliance M-v0.2.3+ hrdb source + 5 PII field type 자동 detection + PII access log 별도 storage + GDPR right-to-be-forgotten 정공법. M-v0.2.0 PoC = audit log + curation 이력 + dashboard M-v0.2.1+ 부터 활성화. M-v0.2.3+ = GDPR/PII + data lineage. cross-section 정합 fix 5 위치 (§11.3 monitoring 28 metrics 정합 + §3.9 OKF concept 운영 lifecycle 5 row audit log 정합 + §3.5.7.4 auto-apply + §3.5.8.2 rollback trigger 의 curator_type `llm-system_admin` audit log 정합 + §16 API versioning 정책 의 audit log 의 user_id 별 마이그레이션 추적 정합 + ADR-0034/0035 영향 row + frontmatter). 본 backend-knowledge 의 정책 (§3.6 Path Y caller-provided user context + §3.6.2 curation governance model + §3.6.3 query scope priority 4-tier + §3.6.4 frontmatter extension 5 fields + §3.6.5 cross-section 정합 fix) 이 §3.6.6 정공법 기반.
- **§3.7.6 data normalization pipeline 자동 검증 정공법 (M-v0.2.0 PoC 운영 시점 활성화, 2026-06-18 신규 — §3.7.5 edge cases 6 종 자동 detection)** — backend-knowledge 의 §3.7.5 edge cases 6 종 의 **자동 detection 정공법**. 5 sub-subsection: §3.7.6.1 Partial failure detection (failed concept count > 10% warning, > 30% critical) / §3.7.6.2 Schema drift detection (Pydantic v2 schema validation + Pi LLM 추천으로 schema normalization 업데이트) / §3.7.6.3 Duplicate concept detection (similarity ≥ 0.95 + priority-based deduplication) / §3.7.6.4 Large raw handling (> 10MB raw 자동 청크 1MB per chunk + 4 worker parallel processing) / §3.7.6.5 Auth failure monitoring (401/403 count ≥ 5/source/1시간 alert + audit log). 본 backend-knowledge 의 정책 (§3.7.5 edge cases 6 종 + §3.8.1 SourceMeta + §3.5.7 Pi LLM + §3.6.6 governance audit log) 이 §3.7.6 정공법 기반. cross-section 정합 fix 9 위치 (§3.7.5 + §3.5.7 + §3.6.6 + §11.1 + §11.3 monitoring 23 metrics + §3.5.6 + §3.8.1 + ADR-0034/0035 + frontmatter). M-v0.2.0 PoC 운영 시점 활성화.
- **§17 cross-cutting 종합 review (umbrella doc 25 commit 후 cross-cutting 최종 종합, 9 sub-subsection, 2026-06-18 신규)** — backend-knowledge 의 umbrella doc 25 commit 후 cross-cutting 최종 종합. 9 sub-subsection: §17.0 high-level 종합 (25 commit + 18/18 Q&A + 28 metrics + 4 cross-cutting 정공법 + stats) / §17.1 25 commit × cross-section 매트릭스 25 row / §17.2 §13.2 known gaps 4 row 자연 해소 시점 (M-v0.2.0 PoC 운영 +1주 이내 100% 해소) / §17.3 §13.3 post-sprint follow-up 5 row 해결 우선순위 (P0 2 row 즉시 + P1 2 row 1주 + P2 1 row release 직전) / §17.4 18/18 Q&A 정합 18 row / §17.5 28 metrics M-v0.2.3+ production 정합 28 row / §17.6 4 cross-cutting 정공법 검증 4 row / §17.7 umbrella doc cross-reference 최종 매트릭스 19 row / §17.8 umbrella doc stats 16 row. 본 backend-knowledge 의 정책 (§0~§16 모든 section + §13.1 cross-reference matrix + §13.2 known gaps + §13.3 post-sprint follow-up + §13.4 정합 검증 + §14 release notes + §15 ADR supersession + §16 API versioning + §17 종합 review) 정합. cross-section 정합 fix 9 위치 (§17 본문 + umbrella frontmatter + §9 row + ADR-0034/0035 영향 + frontmatter + §13.4 정합 검증 row).
- **§17.8 stats 정정 (minor issue 즉시 fix, 2026-06-18)** — §17.8 stats 의 "§13.4 row 15" 표현이 실제 카운트와 불일치. 실제 = 23 row (12 row `| **§` prefix + 11 row non-prefix). 4 row 정정: §17.0 line 5050 "**14 row ✅**" → "**23 row ✅**" / §17.8 line 5288 "**15 row**" → "**23 row**" / §17.8 line 5306 `grep` 검증 정공법 "= 15 row" → "= 23 row" / §13.4 §17 row line 4471 description "§13.4 row 15" → "§13.4 row 23 (12 prefix + 11 non-prefix)". cross-section 정합 fix 6 위치 (§17.0 + §17.8 + §13.4 §17 row 4 위치 + ADR-0034/0035 영향 row + frontmatter). self-review comment 의 minor issue 정공법.

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
| 2026-06-19 | §4.3 영향 row +1 (PR #661 v0.2.0 frontend PoC merge, §1.2 G7 frontend cell 정합) | PR 661 squash merge (#14560ed1), backend-knowledge 의 frontend (SvelteKit 2 + Svelte 5) 1차 release, §1.2 G7 standalone 정공법 frontend cell 영향 (다른 backend 호출 ❌ 유지, frontend → backend-knowledge 만 호출), umbrella doc §1.2 G7 + §3 (frontend-design.md) + §12 frontend page 영향 row, M-v0.2.1+ scope 4 row (backend 3 endpoint + svelte-check strict + listSources 5 source loop) |
| 2026-06-19 | §4.3 영향 row +1 (M-v0.2.1+ backend endpoint 3개 추가, PR #661 placeholder 정합) | #c683c05e main push, 3 endpoint 추가 (POST /ingest/{source}/pull, GET /ingest/statuses, GET /bundles/{name}), §1.2 G7 standalone 정공법 backend cell 정합 (다른 backend 호출 ❌ 유지, frontend → backend-knowledge 만 호출), backend-knowledge 의 API 14 endpoint → 17 endpoint 정합 (§3.1 API 매트릭스 row 추가), M-v0.2.1+ 후속: unit test + openapi + frontend wrapper 갱신 |
| 2026-06-19 | §4.3 영향 row +1 (M-v0.2.1+ frontend svelte-check 0 error + vitest config 분리) | #f49078eb main push, frontend 7 page + types.ts + api.ts + vitest.config.ts 정합, svelte-check 0 error + build OK, §1.2 G7 standalone 정공법 frontend-backend 정합 (frontend → backend-knowledge 만 호출), backend-knowledge 의 frontend layer (3 group + 1 utility, routes/lib/components/config) + 7 page 정합, M-v0.2.1+ 후속 4 row (audit next_offset + offset query, bundles/{name}/concepts, vitest unit test, openapi.yaml sync) |

**supersession 정공법**: 본 ADR-0035 의 결정 자체 (신규 `backend-knowledge/` 디렉터리 신설 + standalone 정책 + OIDC 제외 + Pi 채택) 는 변경 ❌. source plugin 의 구체적 구성 (5종 → 7종) + M-v0.2.0 scope (homelab + homelab_mock → Gitea 4 + homelab_mock) 만 정합. [`release_v0-2_roadmap.md` §9 변경 이력 2026-06-18 row](../planning/release_v0-2_roadmap.md) 가 cross-section 정합 fix 의 source of truth.

**M-v0.2.3+ 부터 supersession 가능** (2026-06-18 신규 결정, §15 ADR supersession 정공법 정합): 본 ADR-0035 가 후속 ADR (e.g., ADR-0036 backend-knowledge 의 §1.2 G7 standalone 정책 변경 / §3.5 운영 환경 변경 / §2.2 LLM technology 변경) 에 의해 supersede 될 경우, [`release_v0-2_roadmap.md` §15](../planning/release_v0-2_roadmap.md) 의 5 step 정공법 (New ADR 작성 → 본 ADR frontmatter `superseded-by` 추가 → 본 ADR §6 row 추가 → cross-reference 4~5 file 갱신 → state.json `adrs` field 갱신) + 12개월 deprecation policy + release notes 정합. supersession 발생 시 본 §6 의 row 가 supersession 결정 row 로 갱신.


