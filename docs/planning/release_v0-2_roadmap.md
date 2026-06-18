# DevHub v0.2.0 릴리즈 로드맵 — 외부 연동 + AI Agent Library (OKF) 통합 백엔드

- 문서 목적: DevHub v0.2.0 의 **단일 source-of-truth umbrella 로드맵**. 외부 시스템 연동 + 데이터 취합을 별도 백엔드로 모으고, 이를 **Google Open Knowledge Format (OKF) 기반 AI Agent Library** 로 발전시키는 컨셉 + 3가지 기본 기능 + 1차 raw 데이터의 API 정책 + 마일스톤 + 기존 `external-integrations-agentic-rag-roadmap.md` child 문서로의 진입 경로.
- 범위: v0.2.0 의 (1) 외부 시스템 연동 분리 (기존 `backend-ai/` 폐기 흡수) (2) OKF 형 knowledge bundle 생성/관리 (3) AI agent + 사용자 query 응답. 1차 외부 연동 (Gitea, HomeLab) + OKF reference PoC + 핵심 3 endpoint.
- 대상 독자: 프로젝트 리드, 모든 contributor (사람 + AI agent), 후속 sprint 작업자, owner.
- 상태: accepted (2026-06-17 publish, 2026-06-18 cross-section 정합 fix 추가, §9 변경 이력 + ADR-0034/0035 publish 완료 + Q&A 11/11 결정 완료)
- 최종 수정일: 2026-06-18 (5 카테고리 정합 + Path Y caller-provided user context + data normalization pipeline + source plugin 작성 정공법 + OKF concept 운영 lifecycle + §4 1차 raw API 심화 + §5 마일스톤 상세화 + §10 DB-based raw + Pi periodic ingest pipeline + §11 운영 runbook + §6.5~§6.7 Phase 1/2/3 운영 정공법 상세 + §7 Q&A 확장 (Q12~Q18) + §12 frontend page 상세화 (M-v0.2.1+) + **§13 cross-cutting 종합 (12 commit 후 umbrella doc 전체 cross-reference 정합성 최종 검토)** + **§1.1 한계 4~7 추가 (2026-06-18 결정에서 식별된 4가지 한계 = caller-provided user context 신뢰 / dual storage mode 운영 복잡도 / backup DR transactional 정합성 / frontend standalone 유지보수) + §1.3 How 정당화 강화 (한계 7개 → §3~§12 해결책 cross-reference 표)** — §3.2.1 보강 + 신규 §3.5~§3.9 + §4.4~§4.7 + §5.4~§5.7 + §10 + §11 + §6.5~§6.7 + §7 Q12~Q18 + §12 (viz.html + 5 page) + §13 (cross-reference matrix 20 row + gap 6 row + post-sprint follow-up 6 row + 정합 검증 12 row ✅) + **§1.1 한계 4~7 (Path Y trust model / dual mode 운영 / backup DR / frontend lifecycle) + §1.3 한계 7개 → 해결책 cross-reference 표**. cross-section 정합 fix: §1.2 G7 / §1.3 producer 다중 row / §2.1 sources/ tree + var/raw 트리 / §2.3 3 row / §3.1 API 매트릭스 / §3.2 type enum / §3.3 frontmatter spec / §3.5.3 bundle 디렉터리 / §3.6.1 endpoint 표 / §3.6.2 curation governance / §3.7 normalization pipeline + §10 storage_mode / §3.7.2 per-source mapping / §3.8.1 SourceMeta + §3.8.4 Step 2 / §4.1 정책 정의 표 / §4.7 raw 정합성 검증 / §5.1 M-v0.2.1 scope / §5.6 cutover checklist §11 cross-reference / §6.1 Phase 1 viz.html / §6.3 Phase 3 Pi 3 역할 / ADR-0034 §4.3 + ADR-0035 §3.4 + §3.6 + §3.8 + §4.2/§4.3 갱신 + §8 timeline Q12~Q18 결정 row 추가 + **§13 자체는 종합 review 이므로 fix 0 row + post-sprint 6 row 명시** + **§1.1 한계 4~7 추가 + §1.3 한계 → 해결책 cross-reference 표 + §1.2 의 "1차 raw 데이터" 정합 (한계 4~7 의 §3~§12 해결책 모두 umbrella 본문 §1.2 G1~G7 + §3~§12 과 정합, 추가 fix 불필요)**).
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

**v0.2.0 설계 과정 (2026-06-18 결정) 에서 추가로 식별된 한계** 4가지:

4. **caller-provided user context 신뢰 문제 (Path Y 의 trust model 한계)**: §3.6.1 의 Path Y caller-provided user context 정책 = backend-knowledge 는 caller (gateway / 별도 agent) 가 전달한 `X-DevHub-User-Context` header 를 신뢰하고 format 검증 (JSON parse + schema check + 만료시간) 만 수행. **문제**: caller 가 잘못된 user context (e.g., 권한 없는 user 의 context 를 위조) 를 전달하면 backend-knowledge 는 그대로 받아들임 (visibility check 만). 권한 우회 가능 (caller 의 신뢰성에 의존). **시나리오**: 악의적 gateway 가 `user_id: "admin"` + `roles: ["system_admin"]` 로 위조하여 backend-knowledge 의 모든 concept 에 접근 + §3.6.2 curation ownership 의 system_admin branch 실행 가능. **영향**: §3.6 의 governance 가 caller 의 신뢰성 만큼만 보장됨. **해결 방향** (본 v0.2.0 scope 외, M-v0.3.0+ 검토): (a) gateway 의 audit 로그 + caller 인증 강화 (Keycloak 의 client_credentials grant + IP allowlist) (b) backend-knowledge 측 anomaly detection (user context 의 issued_at 이 정상 범위 내 인지, request_id 일관성) (c) signature verification (caller 가 user context 에 HMAC signature 첨부, backend-knowledge 가 검증). **현재 한계 식별**: §3.6 의 정책은 caller 신뢰 가정 기반 (gateway 가 올바른 user context 만 전달한다고 가정). §11.3 의 monitoring 으로 audit log 추적은 가능하나, 능동적 위조 탐지는 M-v0.3.0+.

5. **DB-based raw + Pi-driven source 의 운영 복잡도 (dual storage mode 한계)**: §10 의 file + db dual storage mode per source 정책 = 운영자가 per source 별 `storage_mode: file|db` + `normalize_mode: rule-based|pi-sdk|pi-rpc` 관리. **문제**: dual mode 의 운영 복잡도 (per source 별 다른 mode, retention/quota/backup 정책 별도 적용 + §10.4 default mapping 의 운영자 override 시 검증 부담 + storage_mode 변경 시 bundle/index.md 재구성 + Pi ingest pipeline 검증). **시나리오**: 운영자가 source_meta 의 storage_mode 를 file → db 로 변경하면 (a) 기존 var/raw/ 의 file 들이 DB 로 migrate 되어야 하고 (b) Pi ingest pipeline 의 cron 활성화 + normalize_mode 검증 (c) bundle/index.md 의 concept path 변경 + cross-link 검증. **영향**: §10.4 default mapping (gitea 4 = file, homelab/hrdb = db) 외 override 시 운영 부담 가중. **해결 방향** (본 v0.2.0 scope 외, M-v0.2.1+ 검토): (a) storage_mode 변경 CLI tool (mode 변경 + 자동 migration + 검증) (b) 운영자 dashboard 에 mode 변경 영향 preview (c) §11.3 monitoring 의 mode 변경 event audit. **현재 한계 식별**: §10 의 dual mode 는 운영 복잡도를 증가시키지만, 단순화 시 (file only 또는 db only) §10.3 의 Pi-driven 정규화 또는 §3.7 의 rule-based 의 flexibility 가 줄어듦. trade-off.

6. **backup + DR 복잡도 (dual storage mode 의 transactional 정합성 한계)**: §11.2 의 5 backup 대상 (DB / var/bundles/ / var/raw/ / .env-KEK / governance field) + per storage mode 별 backup 방법 + restore RTO 5 target. **문제**: dual storage mode (file + db) 의 backup 일관성 (raw 변경 → bundle 변경 의 transactional 정합성). **시나리오**: 운영자가 file mode source 의 raw 변경 후 backup → bundle 자동 생성 직전에 backup 실패 → restore 시 raw 만 복원, bundle 미복원 (orphan concept 발생) 또는 bundle 만 복원, raw 미복원 (raw_ref dangling). **영향**: §11.2 의 backup 단계가 §11.1 incident runbook 의 §11.1.5 integrity violation 와 연계 (raw 변경 → bundle 변경 → backup 의 transactional 보장 ❌). **해결 방향** (본 v0.2.0 scope 외, M-v0.2.3+ 검토): (a) transactional backup (raw + bundle 동시 backup, 단일 transaction) (b) §11.1.5 integrity violation 자동 trigger backup re-sync (c) §11.2 backup drill (분기 1회, 실 운영 환경 test data 로 partial restore 검증). **현재 한계 식별**: §11.2 backup 은 best-effort (eventual consistency), strong consistency 는 M-v0.2.3+ 검토.

7. **§12 frontend standalone 유지보수 부담 (lifecycle 분리 한계)**: §12 의 `backend-knowledge/web/` 별도 standalone frontend + devhub frontend 와 코드 공유 ❌ 정책 (standalone 정합, §1.2 G7). **문제**: frontend 의 유지보수 (e.g., backend-knowledge API 변경 시 frontend 동시 업데이트 + 별도 deploy + frontend 의 자체 test suite 필요 + API 정합성 검증). **시나리오**: backend-knowledge 의 §3.1 endpoint 변경 시 (e.g., 새 query param 추가) → frontend 의 5 page (concept list / detail / ingest / bundles / raw inspector) 동시 업데이트 필요. frontend 가 backend-knowledge 의 unit test 와 별도 deploy 주기를 가지므로 정합성 drift 발생 가능. **영향**: §5.7 의 parallel sprint PR 전략은 PR 단위 분리이지만, frontend-backend 동기화 보장은 ❌ (CI 자동 검증 권장). **해결 방향** (본 v0.2.0 scope 외, M-v0.2.1+ 검토): (a) PR template 의 cross-component 영향 명시 (e.g., `affects-frontend: yes/no`) (b) frontend build 시 backend-knowledge 의 OpenAPI (`/openapi.json`) 자동 fetch + frontend type generation (TypeScript type or OpenAPI Generator) (c) CI 에서 frontend contract test (backend-knowledge OpenAPI 와 frontend 의 API client 가 정합하는지 자동 검증). **현재 한계 식별**: §12 의 frontend standalone 정책은 코드 중복 (e.g., type 정의, error handling) 을 허용하나, 동기화 보장은 운영자 책임.

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
| G7 | **완전 standalone 운영** — 다른 backend (backend-core / 다른 백엔드 / 다른 시스템) 와의 연결 ❌, OIDC ❌, **외부 시스템만 단방향** (2026-06-17 결정, §4 self-review 강화) + **caller-provided user context** 로 governance 수행 (2026-06-18 Path Y 결정, §3.6.1 정합 — backend-knowledge 는 auth 자체 안 함, caller 가 X-DevHub-User-Context header 로 user/org/project/roles 전달 시 filter/curation ownership 만 backend-knowledge 책임) | M-v0.2.0/v0.2.1: standalone (mock source + no auth + 별도 docker network) → M-v0.2.2: 외부 시스템 **6종** source wire (Gitea 4 + homelab + metrics, backend-core 와 무관) + backend-ai 폐기 (단독 결정) → M-v0.2.3: 외부 시스템 **7종** source wire (+ hrdb) + Pi (pi.dev) LLM enrich 활성화 |

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
| 1 concept = 1 .md | `backend-knowledge/var/bundles/{bundle}/{category}/{slug}.md` (예: `scm/integration_gitea_repo_puller.md` — slug prefix = type, `x_devhub_category` = dir 이름, `type` 은 frontmatter + slug prefix 에 동시 명시, §3.5.3 정합) |
| frontmatter `type` 필수 | 본 프로젝트 `type` enum = `dataset` / `metric` / `api_endpoint` / `runbook` / `integration` / `event` / `reference` / `decision` (8종, §3.2) |
| vendor-neutral | OKF 그대로 채택 (자체 dialect 만들지 않음) |
| producer 다중 | **1차 (M-v0.2.0~v0.2.2)**: human curator + **rule-based enricher** (`curate/enricher.py`, 외부 system adapter 5종 PoC, Q6 정합, §3.7.1 / §3.7.4 정공법). **후속 (M-v0.2.3)**: + LLM enrichment agent (Pi `pi-coding-agent` SDK or RPC mode, 1 vendor, §6.3 정합). **장기 (M-v0.3.0+)**: + multi-vendor LLM (vendor-neutral) |
| consumer 다중 | 신규 API (Query) + frontend 위키 viewer + Obsidian 호환 + LLM context |
| progressive disclosure | `index.md` 자동 생성 (**per-bundle** + **per-type** + **per-category**, §3.5.4 정합) |
| graph | cross-link 자동 추출 (**intra-bundle** / **cross-bundle** / **source-external** 4종 rule, §3.5.5 정합) + **reverse index** (`okf/link_graph.py`) + `viz.html` 자가 viewer (Cytoscape.js) |

**§1.1 의 7가지 한계 → 본 v0.2.0 의 §3~§12 해결책 정렬** (How 의 정당화):

| 한계 | 해결책 위치 | 핵심 메커니즘 |
| --- | --- | --- |
| **1. 외부 시스템 연동 분산** (3 디렉터리) | §2.1 / §3.5 / §6.4 | `backend-knowledge/` 단일 백엔드 + **per-source bundle 디렉터리** (5종 source = 4 bundle, Gitea 4 sub-plugin 통합) + 5종 카테고리 정합 (이슈 트래커 / 위키 / SCM / CI-CD / 코드 품질) |
| **2. AI agent / LLM context 부재** | §3.1 / §3.7 / §3.9 / §6.3 | 3가지 기본 기능 (Ingest / Curate / Query) + OKF 형 knowledge bundle + rule-based enricher (1차) + LLM enrichment (M-v0.2.3+, Pi) + Query API 의 1차 raw context 표시 |
| **3. `backend-ai/` 의 dead state** | §1.2 G2 / §6.6 | `backend-ai/` **폐기** (M-v0.2.2) + placeholder 정리 + 이전 코드 0 (production wiring 없음) |
| **4. caller-provided user context 신뢰 문제** (Path Y) | §3.6.1 / §3.6.2 / §11.3 | Path Y caller-provided user context + format 검증 (JSON parse + schema check + 만료시간) + §11.3 monitoring 의 audit log 추적. **능동적 위조 탐지** (HMAC signature / anomaly detection) 는 M-v0.3.0+ 로 scope 외 (한계 식별 유지, §1.1 한계 4) |
| **5. DB-based raw + Pi-driven source 의 운영 복잡도** (dual mode) | §10.1 / §10.4 / §10.6 / §11.3 | file + db dual storage mode + §10.4 default mapping (gitea 4 = file, homelab/hrdb = db) + §10.6 운영 가이드 (per mode retention / quota / backup 별도 정책) + §11.3 mode 변경 event audit. **storage_mode 변경 CLI tool / dashboard** 는 M-v0.2.1+ 검토 (한계 식별 유지, §1.1 한계 5) |
| **6. backup + DR 복잡도** (dual mode transactional 정합성) | §11.1.5 / §11.2 / §11.3 | 5 backup 대상 (DB / var/bundles/ / var/raw/ / .env-KEK / governance field) + per storage mode 별 backup 방법 + RTO 5 target + §11.1.5 integrity violation 자동 trigger backup re-sync + §11.2 backup drill (분기 1회). **transactional backup (raw + bundle 동일 transaction)** 은 M-v0.2.3+ 검토 (한계 식별 유지, §1.1 한계 6) |
| **7. §12 frontend standalone 유지보수 부담** | §12.1 / §12.2 / §5.7 | `backend-knowledge/web/` 별도 standalone frontend (devhub frontend 와 import ❌) + §12.2 의 API client 가 OpenAPI 자동 fetch + §5.7 parallel sprint PR 전략 (PR 단위 분리). **CI 자동 contract test (OpenAPI ↔ frontend API client 정합 검증)** 은 M-v0.2.1+ 검토 (한계 식별 유지, §1.1 한계 7) |

**정당화 원칙**: v0.2.0 은 **7가지 한계 모두** 를 풀지만, **한계 4~7 은 v0.2.0 PoC 의 trade-off (trust model / dual mode / eventual consistency / lifecycle 분리) 로 받아들이고 능동적 강화는 후속 milestone (M-v0.2.1~M-v0.3.0+) 으로 scope 분리**. §1.1 에 4~7 을 명시적으로 식별한 이유: (a) 한계 자체를 알지 못하면 후속 sprint 진입 시 "왜 HMAC signature 가 없지?" / "왜 transactional backup 이 없지?" 같은 의문 발생, 운영자가 v0.2.0 PoC 의 의도적 한계임을 인지하지 못함 (b) 후속 milestone 의 scope 결정 (M-v0.2.1+ 가 HMAC signature / CLI tool / contract test 어느 것을 우선할지) 시 baseline 으로 사용. **본 v0.2.0 PoC 운영 중에는 한계 4~7 이 자연 노출 (audit log / mode 변경 event / partial restore / frontend drift) 되나, §11 운영 runbook 으로 mitigation 가능**. §13.2 의 known gaps 와 정합.

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
├── curate/                                # raw → OKF concept 정형화 (§3.7.1 정공법)
│   ├── enricher.py                        # rule-based enrich (LLM-optional, 1차 rule-based 만) — source plugin 의 normalize() 와 함께 5 step normalization 의 Step 3+5 담당 (§3.7.1 / §3.7.4 정합)
│   ├── index_builder.py                   # per-bundle/per-type/per-category index.md + viz.html 동시 생성 (§3.5.4 정합)
│   └── link_resolver.py                   # unresolved cross-link 추적 (1차 rule-based, M-v0.2.3+ Pi LLM cross-link 자동 resolution, §3.5.5 정합)
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
│   ├── raw/                               # 1차 raw 데이터 (§4 API 정책 + §4.4 raw 운영 정책) — 봉투 암호화 (ADR-0025) + .gitignore (민감 source, §4.4 표) + retention default 90일 + storage quota default 1GB/bundle
│   │   └── {source}/{slug}.json           # 예: homelab/2026-06-17-foo.json (sha256 + received_at + visibility)
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
| **다른 backend 연결 (general)** | ❌ **전면 금지**. `backend-knowledge` 는 **완전 standalone 시스템** (2026-06-17 결정, §4 self-review 강화). 다른 backend (backend-core / 다른 백엔드 / 다른 시스템) 의 Go/Python 코드 / API / domain model / database / cache / envelope / repository / 어떤 layer 든 import / 호출 / 공유 ❌. **외부 시스템 7종 source 만 단방향** (Gitea 4 + homelab + metrics + hrdb, M-v0.2.3 운영 기준, §1.2 G3 정합). **단, caller-provided user context (X-DevHub-User-Context header) 의 schema 는 DevHub 의 user/org/project 모델과 format 호환** (cross-reference 만, import ❌, §3.6.1 정합) |
| **OIDC / Keycloak** | ❌ **OIDC 자체를 본 시스템에서 제외**. Keycloak 인증은 backend-core 의 책임 (변경 없음). `backend-knowledge` 는 bearer token / API key / session 어떤 인증 scheme 도 자체적으로 검증 안 함. caller (gateway / 별도 agent) 가 Keycloak 인증 후 user context 만 전달 (§3.6.1) |
| **외부 시스템** | ✅ **유일한 통신 대상**. Gitea 1 instance (4 sub-plugin) + homelab + metrics + hrdb 등 source plugin **7종** (M-v0.2.3 운영 기준) 의 외부 시스템 API 만 호출. 단방향 pull (외부 → backend-knowledge) |
| **API 인증** | **internal-only, no auth** + **caller-provided user context (Path Y, §3.6.1, 2026-06-18 결정)**. `/api/v0-2/*` endpoint 는 bearer/API key 자체 검증 안 함. **단, caller 가 `X-DevHub-User-Context` header (base64url(json) 의 user/org/project/roles) 를 전달하면, backend-knowledge 는 그 context 로 filter / curation ownership check 만 수행** (gateway/firewall/IP allowlist 가 호출 자체 보호, Phase 1~3 의 운영 책임). 운영자 또는 별도 agent 가 호출 (**backend-core 의 어떤 layer 든 호출 ❌**, 2026-06-17 결정, §1.2 G7 / §7 Q9 정합). query / concept 직접 조회 / concept manual edit endpoint 는 user context 필수 (§3.6.1 표 정합) |
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

> **인증 정책 (모든 endpoint 공통)**: **internal-only, no auth** + **Path Y caller-provided user context (2026-06-18 결정, §3.6.1 정합)**. `/api/v0-2/*` 전체가 인증 없이 호출 가능. 별도 gateway / firewall / IP allowlist 가 호출 자체 보호 (Phase 1~3 의 운영 책임, §2.3 참조). 운영자 또는 별도 agent 가 호출. OIDC / Keycloak / backend-core 인증 위임 ❌. **단, caller 가 `X-DevHub-User-Context` header (base64url(json)) 로 user/org/project/roles 를 전달하면, backend-knowledge 는 그 context 로 filter / curation ownership check 만 수행 (§3.6.1 endpoint 별 필수/권장 표 + OpenAPI security scheme 정합)**. 운영자 또는 별도 agent 가 호출.

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

> **어떤 source 가 어떤 type emit** = §3.7.2 정합. 7 source (Gitea 4 sub-plugin + homelab + metrics + hrdb) 별 emit_types 는 source plugin 의 `SourceMeta.emit_types` 필드로 명시 (§3.7.4 정공법).

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

#### 3.2.1.1 5 카테고리별 대표 concept 예시 (frontmatter 발췌, 2026-06-18 신규)

5 카테고리별 1개 concept 의 frontmatter 발췌 (5 카테고리 = 5 concept 1:1 mapping 으로 카테고리 자체의 의미를 명확히 함). 전체 concept 예시 + 5×8 matrix + bundle 디렉터리 구조는 §3.5 정합.

**1. 이슈 트래커 (`issue_tracker`)** — `gitea_issue` source:
```yaml
---
type: integration
title: "Gitea issue puller"
description: "Pulls Gitea issue list + comments + labels into backend-knowledge via REST"
tags: ["gitea", "issue_tracker", "integration"]
x_devhub_source: "gitea_issue"
x_devhub_bundle: "devhub-gitea"
x_devhub_category: "issue_tracker"
---
```

**2. 위키 (`wiki`)** — `gitea_wiki` source:
```yaml
---
type: reference
title: "Gitea wiki page mirror"
description: "Mirror of Gitea wiki page content for offline query"
tags: ["gitea", "wiki", "reference"]
x_devhub_source: "gitea_wiki"
x_devhub_bundle: "devhub-gitea"
x_devhub_category: "wiki"
---
```

**3. SCM (`scm`)** — `gitea_repo_pull` source:
```yaml
---
type: integration
title: "Gitea repository puller"
description: "Pulls Gitea repository list + metadata into backend-knowledge via REST + git HTTP"
tags: ["gitea", "scm", "integration"]
x_devhub_source: "gitea_repo_pull"
x_devhub_bundle: "devhub-gitea"
x_devhub_category: "scm"
---
```

**4. CI-CD (`cicd`)** — `gitea_action` source:
```yaml
---
type: event
title: "Gitea Actions workflow run event"
description: "Webhook payload schema for Gitea Actions workflow_run events (queued/in_progress/completed)"
tags: ["gitea", "cicd", "event", "webhook"]
x_devhub_source: "gitea_action"
x_devhub_bundle: "devhub-gitea"
x_devhub_category: "cicd"
---
```

**5. 코드 품질 (`code_quality`)** — (1차 scope 외, 2차 wire, M-v0.2.4+ 검토):
- 2차 wire 후보: SonarQube, Snyk, Codecov, Dependabot
- M-v0.2.0~v0.2.3 scope 외 (5 카테고리 5번째 = 2차 wire 결정, 2026-06-17 §3.2.1 정합)
- 2차 PoC 시 `code_quality` category 의 concept 예시 추가 (예: SonarQube project analysis metric = `metric_sonarqube_code_coverage_pct.md`)

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
> **5 카테고리 × 8 type 의 valid combination**: §3.5.2 정합 (5×8 matrix + 5종 PoC source plugin mapping). 모든 조합이 가능하나 일부는 uncommon (△, §3.5.2 범례 정합).
> **Path Y governance 필드 (2026-06-18 신규, §3.6.4 정합)**: 5 field 추가 — `x_devhub_owner_org_id` (string, FK → org_units.unit_id) / `x_devhub_owner_user_id` (string, FK → users.user_id, nullable) / `x_devhub_owner_org_unit_ids` (array of string, recursive subtree) / `x_devhub_owner_project_ids` (array of string) / `x_devhub_visibility` (enum: `org` \| `personal` \| `project` \| `public`). caller-provided user context (§3.6.1) 와 함께 backend-knowledge 의 curation governance + query scope priority (org > personal > project > public, §3.6.3) 의 기반. default: source plugin 자동 emit = `org` (bundle owner org 기준), human 작성 = `org` (manual override 가능).

### 3.4 Envelope (독립 정의)

- **envelope = `backend-knowledge` 자체 정의** (backend-core 의 `docs/api/conventions.md` 와 format 호환 유지, **import ❌, cross-reference 만**):
  ```json
  { "envelope": { "version": "v0", "trace_id": "..." }, "data": { ... } }
  ```
- **common enum**: `error.code` (`E_OK`, `E_NOT_FOUND`, `E_CONFLICT`, `E_VALIDATION`, `E_UNAUTHORIZED`, `E_FORBIDDEN`, `E_RATE_LIMIT`, `E_INTERNAL`) — backend-core 와 format 호환 (cross-backend client 가 envelope parser 재사용 가능)
- **OpenAPI**: `/openapi.json` (FastAPI 자동 생성) — `docs/openapi.yaml` 와는 별도 (`x-internal: backend-knowledge` 표기)

### 3.5 Concept organization (5 카테고리 + 8 type + index.md + cross-link) — 2026-06-18 신규

OKF 형 concept 를 실제로 디렉터리에 배치하고 navigation 가능하게 만드는 **운영 정공법**. §1.3 의 "progressive disclosure + graph" 원칙 + §3.2.1 의 5 카테고리 결정 + §3.3 의 frontmatter spec 의 **실제 적용 규칙**. 본 §3.5 가 ADR-0034 (OKF 채택) 의 §4.3 영향 section 의 운영 정합 기준.

#### 3.5.1 원칙 (orthogonal axes)

| 축 | 정의 | 값 |
| --- | --- | --- |
| **type** (8종, §3.2) | concept 의 **본질적 종류** — 무엇에 대한 knowledge 인가 | `dataset` / `metric` / `api_endpoint` / `runbook` / `integration` / `event` / `reference` / `decision` |
| **`x_devhub_category`** (5종, §3.2.1) | concept 의 **출처 분류** — 어느 외부 시스템 영역의 knowledge 인가 | `issue_tracker` / `wiki` / `scm` / `cicd` / `code_quality` (5 카테고리 외 시스템은 `x_devhub_category` 미설정 또는 별도 tag) |

- **두 축은 orthogonal**. 모든 type × category 조합이 가능 (일부는 uncommon, §3.5.2 정합).
- **bundle** = 1 외부 시스템 단위 (예: `devhub-gitea`) 또는 cross-cutting 주제 단위. 5종 PoC source plugin 은 **Gitea 1 instance = 1 bundle (`devhub-gitea`, Gitea 4 sub-plugin 통합)** + homelab_mock = 1 bundle (`devhub-homelab`). **M-v0.2.0 5종 PoC = 2 bundle** (Q6 결정 정합, §6.4 / §3.5.3 정합).
- **분류 (categorization) 의 1차**: rule-based (frontmatter `x_devhub_*` 직접 read, `curate/enricher.py`). **2차**: M-v0.2.3+ Pi LLM 분류 가능 (frontmatter 가 비어있을 때 LLM 추천, §6.3).
- **§1.1 의 "내부 knowledge 가 산만하여 AI agent 가 참조할 표준 부재"** 한계의 해법 = 5 카테고리 + 8 type 의 orthogonal 분류 + 일관된 bundle 구조 (§1.1 정합).

#### 3.5.2 5×8 matrix (valid combinations)

**범례**: ○ = common (해당 카테고리에서 자주 등장) / △ = uncommon (특수 상황에서만) / ✗ = 비권장 (이해관계자 혼선 유발)

| type \\ category | 이슈 트래커 | 위키 | SCM | CI-CD | 코드 품질 |
| --- | --- | --- | --- | --- | --- |
| `dataset` | ○ (gitea.issues table) | ○ (gitea.wiki_pages table) | ○ (gitea.repositories table) | △ (gitea.action_runs table) | ○ (analysis_results table) |
| `metric` | △ (issue_throughput) | △ (wiki_search_duration) | ○ (repo_sync_duration_seconds) | ○ (workflow_duration_seconds) | ○ (code_coverage_pct) |
| `api_endpoint` | ○ (gitea_issue_api_list) | ○ (gitea_wiki_api_list) | ○ (gitea_repo_pull_api) | ○ (gitea_action_api) | ○ (sonarqube_api) |
| `runbook` | ○ (issue_triage_runbook) | △ (wiki_maintenance_runbook) | ○ (repo_pull_failure_recovery) | ○ (cicd_failure_recovery) | ○ (code_quality_alert_runbook) |
| `integration` | ○ (gitea_issue_puller) | △ (gitea_wiki_puller) | ○ (gitea_repo_puller) | ○ (gitea_action_runner) | △ (sonarqube_scanner) |
| `event` | ○ (gitea_issue_event) | △ (gitea_wiki_page_event) | ○ (gitea_push_event) | ○ (gitea_action_run_event) | ✗ |
| `reference` | ○ (gitea_issue_api_doc) | ○ (gitea_wiki_doc_mirror) | ○ (gitea_repo_api_doc) | ○ (gitea_action_doc_mirror) | ○ (sonarqube_doc_mirror) |
| `decision` | △ (in-bundle ADR) | △ (in-bundle ADR) | △ (in-bundle ADR) | △ (in-bundle ADR) | △ (in-bundle ADR) |

**5종 PoC source plugin 의 category × type mapping** (M-v0.2.0 PoC, Q6 결정 정합):

| Source plugin | Bundle | Category | 주로 등장하는 type (representative) |
| --- | --- | --- | --- |
| `gitea_repo_pull` | `devhub-gitea` | `scm` | `integration`, `api_endpoint`, `metric`, `runbook`, `event`, `dataset`, `reference` |
| `gitea_issue` | `devhub-gitea` | `issue_tracker` | `integration`, `api_endpoint`, `event`, `dataset`, `runbook`, `metric`, `reference` |
| `gitea_wiki` | `devhub-gitea` | `wiki` | `integration`, `api_endpoint`, `reference`, `dataset`, `metric` |
| `gitea_action` | `devhub-gitea` | `cicd` | `integration`, `api_endpoint`, `event`, `metric`, `runbook`, `dataset`, `reference` |
| `homelab_mock` | `devhub-homelab` | (5 카테고리 외, `x_devhub_category` 미설정 또는 `homelab_internal` tag) | `reference`, `metric`, `runbook`, `api_endpoint` |

> **Gitea 1 instance = 1 bundle** (Q6 결정 정합): Gitea 의 4 sub-plugin (`gitea_repo_pull` / `gitea_issue` / `gitea_wiki` / `gitea_action`) 은 **모두 `devhub-gitea` bundle** 안에서 4개 category (`scm` / `issue_tracker` / `wiki` / `cicd`) directory 로 분리. **bundle = 1 외부 시스템 단위** 원칙 정합. 운영 시 Gitea 1 instance 의 인증/연결 config 를 1번만 관리.

#### 3.5.3 Bundle 디렉터리 구조 + concept 예시

**Bundle 디렉터리 layout** (§2.1 `var/bundles/` 정합 + §3.7.1 5 step normalization 정합, M-v0.2.0 PoC 기준):

```
backend-knowledge/var/bundles/
├── devhub-gitea/                          # Gitea 1 instance 의 bundle (M-v0.2.0 PoC)
│   ├── index.md                           # per-bundle index (§3.5.4)
│   ├── viz.html                           # self-contained Cytoscape viewer (§1.3 / §3.1 정합)
│   ├── scm/                               # x_devhub_category = scm
│   │   ├── index.md                       # per-category index (§3.5.4)
│   │   ├── integration_gitea_repo_puller.md
│   │   ├── api_endpoint_gitea_repo_pull.md
│   │   ├── metric_repo_kpi_sync_duration_seconds.md
│   │   ├── runbook_gitea_repo_pull_failure_recovery.md
│   │   ├── event_gitea_push_event.md
│   │   ├── dataset_gitea_repositories.md
│   │   └── reference_gitea_repo_api_doc.md
│   ├── issue_tracker/                     # x_devhub_category = issue_tracker
│   │   ├── index.md
│   │   ├── integration_gitea_issue_puller.md
│   │   ├── api_endpoint_gitea_issue_api_list.md
│   │   ├── event_gitea_issue_event.md
│   │   ├── runbook_issue_triage.md
│   │   ├── metric_issue_throughput.md
│   │   └── dataset_gitea_issues.md
│   ├── wiki/                              # x_devhub_category = wiki
│   │   ├── index.md
│   │   ├── integration_gitea_wiki_puller.md
│   │   ├── api_endpoint_gitea_wiki_api_list.md
│   │   └── reference_gitea_wiki_page_doc.md
│   ├── cicd/                              # x_devhub_category = cicd
│   │   ├── index.md
│   │   ├── integration_gitea_action_runner.md
│   │   ├── api_endpoint_gitea_action_api_list.md
│   │   ├── event_gitea_action_run_event.md
│   │   ├── metric_workflow_duration_seconds.md
│   │   ├── runbook_cicd_failure_recovery.md
│   │   └── dataset_gitea_action_runs.md
│   └── decision/                          # in-bundle ADR
│       ├── index.md
│       └── decision_2026_06_18_gitea_pull_strategy.md
└── devhub-homelab/                        # homelab 1 instance 의 bundle (M-v0.2.0 PoC, mock)
    ├── index.md
    ├── viz.html
    └── homelab_internal/                  # 5 카테고리 외 — `x_devhub_category` 미설정 또는 tag "homelab_internal" 으로 grouping
        ├── index.md
        ├── reference_homelab_node_inventory.md
        ├── metric_homelab_pull_duration.md
        ├── runbook_homelab_recovery.md
        └── api_endpoint_homelab_node_api.md
```

**Representative concept frontmatter 예시** (5 카테고리 + 8 type 중 대표, §3.2.1.1 보강 + 추가):

```yaml
---
# integration (scm) — gitea_repo_pull
type: integration
title: "Gitea repository puller"
description: "Pulls Gitea repository list + metadata into backend-knowledge via REST + git HTTP"
resource: "https://docs.gitea.io/api-1.20/"
tags: ["gitea", "scm", "integration", "pull"]
timestamp: "2026-06-18T10:00:00+09:00"
x_devhub_source: "gitea_repo_pull"
x_devhub_raw_ref: "raw://gitea/2026-06-18-repo-pull.json"
x_devhub_bundle: "devhub-gitea"
x_devhub_version: 1
x_devhub_curator: "rule-based"
x_devhub_category: "scm"
---
```

```yaml
---
# event (issue_tracker) — gitea_issue
type: event
title: "Gitea issue event payload"
description: "Webhook payload schema for Gitea issue events (opened/closed/labeled)"
resource: "https://docs.gitea.io/api-1.20/#tag-issue"
tags: ["gitea", "issue_tracker", "event", "webhook"]
timestamp: "2026-06-18T10:00:00+09:00"
x_devhub_source: "gitea_issue"
x_devhub_raw_ref: "raw://gitea/2026-06-18-issue-event.json"
x_devhub_bundle: "devhub-gitea"
x_devhub_version: 1
x_devhub_curator: "rule-based"
x_devhub_category: "issue_tracker"
---
```

```yaml
---
# metric (cicd) — gitea_action
type: metric
title: "Workflow run duration (Gitea Actions)"
description: "Histogram of Gitea Actions workflow run duration in seconds, labeled by repo + workflow"
tags: ["gitea", "cicd", "metric", "duration", "kpi"]
timestamp: "2026-06-18T10:00:00+09:00"
x_devhub_source: "gitea_action"
x_devhub_raw_ref: "raw://metrics/2026-06-18-workflow-duration.json"
x_devhub_bundle: "devhub-gitea"
x_devhub_version: 1
x_devhub_curator: "rule-based"
x_devhub_category: "cicd"
---
```

```yaml
---
# runbook (wiki) — gitea_wiki
type: runbook
title: "Wiki page sync failure recovery"
description: "Steps to recover from gitea_wiki sync failure: retry policy, fallback, manual trigger"
tags: ["gitea", "wiki", "runbook", "recovery"]
timestamp: "2026-06-18T10:00:00+09:00"
x_devhub_source: "gitea_wiki"
x_devhub_raw_ref: "raw://gitea/2026-06-18-wiki-recovery.txt"
x_devhub_bundle: "devhub-gitea"
x_devhub_version: 1
x_devhub_curator: "rule-based"
x_devhub_category: "wiki"
---
```

```yaml
---
# decision (in-bundle ADR) — cross-cutting
type: decision
title: "Gitea 1 instance = 1 bundle (4 sub-plugin 통합)"
description: "ADR-style decision: Gitea 의 4 sub-plugin (gitea_repo_pull / gitea_issue / gitea_wiki / gitea_action) 은 1 bundle 로 통합. 인증/연결 config 1번만 관리. 4 카테고리 (scm/issue_tracker/wiki/cicd) 별 directory 분리."
tags: ["decision", "bundle", "gitea", "2026-06-18"]
timestamp: "2026-06-18T10:00:00+09:00"
x_devhub_source: "rule-based-curator"
x_devhub_bundle: "devhub-gitea"
x_devhub_version: 1
x_devhub_curator: "human"
x_devhub_category: "scm"  # primary category; or omit for cross-cutting decisions
---
```

#### 3.5.4 index.md 자동 생성 규칙

**3종 index.md** — bundle rebuild 시 `curate/index_builder.py` 가 자동 생성:

| 종류 | 위치 | 생성 시점 | 내용 | 활성화 시점 |
| --- | --- | --- | --- | --- |
| **per-bundle** | `{bundle}/index.md` | 모든 rebuild | bundle 의 전체 concept 를 **category → type → slug** 순으로 grouping. 각 concept = title + 1-line description + path link. footer: `Last rebuild: {ts} \| Sources: {N} \| Categories: {M} \| Total: {K} concepts` | M-v0.2.0 (PoC) |
| **per-category** | `{bundle}/{category}/index.md` | 모든 rebuild | category 내 모든 concept. **type → slug** 순. footer: `Category: {category} \| Total: {K}` | M-v0.2.0 (PoC, category directory 가 곧 grouping 단위) |
| **per-type** | 생략 가능 (per-bundle index 가 type 으로 grouping 함) 또는 `{bundle}/{type}/index.md` 별도 | 선택적 | type 내 모든 concept. alphabetical by slug. | M-v0.2.1+ (type 단독 navigation 필요 시) |

**Per-bundle index.md 예시 발췌** (`devhub-gitea/index.md`):

```markdown
# devhub-gitea bundle

- Last rebuild: 2026-06-18T10:00:00+09:00
- Sources: 4 (gitea_repo_pull, gitea_issue, gitea_wiki, gitea_action)
- Categories: 4 (scm, issue_tracker, wiki, cicd)
- Total: 25 concepts

## scm (8 concepts)

### integration (1)
- [Gitea repository puller](scm/integration_gitea_repo_puller.md) — Pulls Gitea repository list + metadata into backend-knowledge via REST + git HTTP

### api_endpoint (1)
- [Gitea repo pull API](scm/api_endpoint_gitea_repo_pull.md) — REST + git HTTP endpoints for repository pull

### metric (1)
- [Repo KPI sync duration seconds](scm/metric_repo_kpi_sync_duration_seconds.md) — Histogram of repository KPI sync duration

... (이하 scm 의 나머지 type 별 listing 생략)

## issue_tracker (6 concepts)
...

## wiki (3 concepts)
...

## cicd (6 concepts)
...

## decision (1)
- [Gitea 1 instance = 1 bundle decision](decision/decision_2026_06_18_gitea_pull_strategy.md) — ADR-style decision
```

**구현 정공법** (`curate/index_builder.py`):

1. bundle directory 의 모든 `.md` file 을 scan (recursive glob `{bundle}/**/*.md`, 단 `index.md` / `viz.html` 제외)
2. 각 file 의 frontmatter parse (PyYAML + python-frontmatter, §2.2 정합)
3. (concept_path, type, category, source, bundle, title, description, timestamp) tuple 추출
4. sort: category → type → slug (안정 sort)
5. group by category → within group, group by type
6. emit Markdown table-of-contents 형식 (per-bundle index)
7. category directory 의 `index.md` 도 동시 emit (per-category index)

**Timing**: M-v0.2.0 = rule-based 만 (frontmatter 직접 read). M-v0.2.3+ = Pi LLM cross-link 자동 resolution (M-v0.2.0~v0.2.2 = rule-based 만, §3.1 `POST /bundles/{bundle}/rebuild` 정합, §6.3).

#### 3.5.5 cross-link 규칙

**4종 link 종류**:

| Link 종류 | Syntax | 언제 사용 | 예시 |
| --- | --- | --- | --- |
| **intra-bundle** (가장 일반적) | 같은 `{category}` dir: `[title]({slug}.md)` / 다른 `{category}` dir: `[title](../{category}/{slug}.md)` | 같은 bundle 내 concept 간 | 같은 dir: `[Gitea repository puller](integration_gitea_repo_puller.md)` from `devhub-gitea/scm/api_endpoint_gitea_repo_pull.md`. 다른 dir: `[Gitea issue event payload](../issue_tracker/event_gitea_issue_event.md)` from `devhub-gitea/scm/api_endpoint_gitea_repo_pull.md` |
| **cross-bundle** (드묾) | `[title](../../{other-bundle}/{category}/{slug}.md)` | 다른 bundle 의 concept 참조 시 (명시적 의미만) | devhub-gitea 의 metric → devhub-metrics 의 metric |
| **source/external** (vendor docs) | `[title](https://...)` 또는 frontmatter `resource:` | vendor 공식 docs / OpenAPI / schema 참조 | `<https://docs.gitea.io/api-1.20/>` |
| **reverse index** (incoming link 추적) | 자동 생성 (`okf/link_graph.py`) | 모든 link 의 target 추적 | `{concept_path: [incoming_link_1, incoming_link_2, ...]}` |

**Cross-link 정책 (정공법)**:

1. **Intra-bundle link 가 default**. 같은 bundle 내 concept 간 link 는 자유롭게 (단, 의미 있는 link 만).
2. **Cross-bundle link 는 최소화**. cross-bundle link 가 많은 concept → bundle 통합 검토. "bundle = 1 외부 시스템 단위" 원칙 정합.
3. **Source/external link 는 vendor docs**. OKF 의 `reference` type concept 는 본문에 vendor docs 발췌 포함 가능 (mirror).
4. **Reverse index 는 자동**. `okf/link_graph.py` 가 bundle rebuild 시 전체 link scan → reverse index 생성. Incoming link 0 = orphan, unresolved link = target 부재. `curate/link_resolver.py` 가 보고.
5. **Resolution timing**: M-v0.2.0~v0.2.2 = rule-based 보고 (orphan list + unresolved link list). M-v0.2.3+ = Pi LLM 이 unresolved link 에 대해 "가장 유사한 concept 추천" (cross-link 자동 resolution, §3.1 정합).

### 3.6 Data governance & query scoping (Path Y caller-provided user context, 2026-06-18 신규)

**문제 (Motivation)**: §1.1 의 4번째 한계 "내부 knowledge 가 코멘트/구두/문서 산만하여 AI agent 가 참조할 표준 부재" + §3.5 의 5 카테고리 결정 이후 자연스러운 후속. 데이터가 (category, system) 단위로 수집된 후, **(a) 누가 어떤 데이터의 curation 책임지고**, **(b) 누가 어떤 데이터를 조회 우선순위로 보는지** governance 가 필요. 사용자 2026-06-18 결정.

**Tension 인식**: §1.2 G7 + §2.3 standalone 정책은 backend-knowledge 를 "internal-only no auth" 로 명시. 이 standalone 정책과 user/org 기반 governance 사이 tension 을 다음 **Path Y** 로 해결 (2026-06-18 사용자 결정):

- **backend-knowledge 는 auth 자체를 수행 안 함** (Keycloak / OIDC / bearer token 자체 검증 ❌, §2.3 정합).
- **대신 caller (gateway / 별도 agent) 가 authenticated user context 를 전달**하면, backend-knowledge 는 그 context 로 **filter + governance 만 수행**.
- 즉, **auth 책임 = caller** (DevHub backend-core 의 Keycloak federation 정합), **governance 책임 = backend-knowledge** (curation ownership + query scope filter).
- backend-core 의 어떤 layer 도 호출 안 함 (standalone 정합, §1.2 G7). 단지 context schema 가 DevHub 의 user/org/project 모델과 format 호환 (cross-reference).

**독립 backend-core RBAC 와의 정합**: DevHub backend-core 의 user/org/project 모델 + 2차원 RBAC + `enforceRowOwnership` 패턴 ([ADR-0011](../adr/0011-rbac-row-scoping.md) / [ADR-0013](../adr/0013-dreq-rbac-row-scoping.md) / [role-access-concept.md §2~4](./role-access-concept.md)) 의 **데이터 모델은 caller 가 채워서 전달**. backend-knowledge 는 이 context 로 filter / governance 만 수행. backend-core 코드/API 호출 ❌, format-only 정합.

#### 3.6.1 Caller-provided user context (schema + trust model)

**HTTP header 규약** (M-v0.2.0 PoC):

```
X-DevHub-User-Context: <base64url(json)>
```

**JSON schema** (M-v0.2.0 PoC, DevHub backend-core `AppUser` + RBAC 모델과 정합):

```json
{
  "version": "v0",
  "user_id": "u_abc123",
  "org_id": "ou_root_dept_a",
  "org_unit_ids": ["ou_root_dept_a", "ou_dept_b1", "ou_dept_b2"],
  "project_ids": ["prj_x", "prj_y"],
  "roles": ["developer", "project_leader:prj_x"],
  "request_id": "req_20260618_xxx",
  "issued_at": "2026-06-18T10:00:00+09:00"
}
```

| Field | 정의 | 출처 (DevHub backend-core) |
| --- | --- | --- |
| `user_id` | DevHub internal user PK | `users.user_id` |
| `org_id` | primary organization | `users.primary_unit_id` |
| `org_unit_ids` | org_head scope 의 subtree (recursive CTE 결과) | `org_units` recursive CTE from `OrgUnit.LeaderUserID` |
| `project_ids` | user 의 project memberships | `project_members.user_id` |
| `roles` | system role + resource role 목록 | `users.role` + `project_members.project_role` |
| `request_id` | trace id (envelope trace_id 와 동일값 권장) | cross-cutting |
| `issued_at` | context 생성 시각 (caller 측) | caller 책임 |

**Trust model**:

- backend-knowledge 는 **context 의 진위를 검증하지 않음** (caller 신뢰). caller (gateway / 별도 agent) 가 authentication + context 구성 책임 (§2.3 "운영자 또는 별도 agent 가 호출" 정합).
- **format 검증** 만 backend-knowledge 책임: JSON parse + schema check (`version`, `user_id`, `org_id` 필수) + `issued_at` 만료 (e.g., 5분) → 만료 시 401 `E_UNAUTHORIZED`.
- 이상 context 탐지 (e.g., 존재하지 않는 org_id) 는 **out-of-scope for v0.2.0** (caller 책임). M-v0.3.0+ anomaly detection 검토 (rule-based audit log 기반).
- backend-knowledge 측 audit log 에 `X-DevHub-User-Context` 의 hash + `request_id` 기록 (cross-trace 가능).

**Required vs Optional (endpoint 별)**:

| Endpoint | user context | 응답 (없을 때) |
| --- | --- | --- |
| `POST /api/v0-2/query` | **필수** (user-scoped query) | 400 `E_VALIDATION` ("X-DevHub-User-Context required") |
| `GET /api/v0-2/concepts/{type}/{name}` | **필수** (visibility check) | 400 `E_VALIDATION` |
| `GET /api/v0-2/search` | **필수** (full-text search 의 visibility filter) | 400 `E_VALIDATION` |
| `POST /api/v0-2/concepts/{id}` (PUT, manual edit) | **필수** (curation ownership check, §3.6.2) | 400 `E_VALIDATION` |
| `POST /api/v0-2/concepts/{id}/enrich` | **권장** (LLM enrich usage attribution) | 200 (anonymous curator) |
| `POST /api/v0-2/bundles/{bundle}/rebuild` | **권장** (audit attribution) | 200 (system rebuild) |
| `POST /api/v0-2/ingest/{source}/sync` | **권장** (audit attribution) | 200 (system sync) |
| `POST /api/v0-2/raw` (1차 raw 등록) | **필수** (provenance attribution + 권한 check, §4.5 정합) | 400 `E_VALIDATION` (X-DevHub-User-Context missing) or 403 `E_FORBIDDEN` (raw.register_denied) |
| `GET /api/v0-2/raw/{type}/{name}` | **필수** (raw 의 visibility 정합, §4.5 정합) | 400 `E_VALIDATION` or 403 `E_FORBIDDEN` (visibility mismatch) or 404 `E_NOT_FOUND` |
| `GET /api/v0-2/raw?source=...&since=...` (list) | **필수** (caller scope filter, §4.5 정합) | 400 `E_VALIDATION` (filtered list 반환) |
| `DELETE /api/v0-2/raw/{id}` | **필수** (raw 삭제 권한, §4.5 정합) | 400 `E_VALIDATION` or 403 `E_FORBIDDEN` (raw.delete_denied) |
| `GET /api/v0-2/bundles/{bundle}/index.md` | **권장** (bundle-level visibility check) | 200 (visibility default 적용) |
| `GET /api/v0-2/bundles/{bundle}/viz.html` | **권장** | 200 |
| `GET /api/v0-2/bundles` (list) | **권장** | 200 (public bundle 만) |
| `POST /api/v0-2/bundles` (create) | **필수** (bundle owner 설정) | 400 `E_VALIDATION` |

**OpenAPI security scheme**:

```yaml
components:
  securitySchemes:
    DevHubUserContext:
      type: apiKey
      in: header
      name: X-DevHub-User-Context
      description: |
        Caller-provided user context. backend-knowledge 는 auth 자체 수행 안 함 (§2.3,
        §3.6.1). caller (gateway / agent) 가 Keycloak 인증 후 user/org/project/roles 를
        base64url(json) 으로 전달. 권한 검증 (curation ownership, query scope filter) 만
        backend-knowledge 책임.
```

#### 3.6.2 Curation governance model

**Concept ownership 표현** (frontmatter extension, §3.3 + §3.6.4 정합):

```yaml
---
type: metric
title: "Workflow run duration (Gitea Actions)"
description: "Histogram of Gitea Actions workflow run duration in seconds, labeled by repo + workflow"
tags: ["gitea", "cicd", "metric", "duration", "kpi"]
timestamp: "2026-06-18T10:00:00+09:00"
x_devhub_source: "gitea_action"
x_devhub_bundle: "devhub-gitea"
x_devhub_version: 1
x_devhub_curator: "rule-based"

# Path Y curation governance (2026-06-18 신규, §3.6.2 정합)
x_devhub_owner_org_id: "ou_root_devhub"
x_devhub_owner_user_id: null
x_devhub_owner_org_unit_ids: ["ou_root_devhub", "ou_devhub_gitea_team"]
x_devhub_owner_project_ids: []
x_devhub_visibility: "org"
---
```

**Visibility enum** (4종, M-v0.2.0 PoC):

| visibility | 정의 | 조회 가능 caller 조건 |
| --- | --- | --- |
| `public` | owner_org / owner_user 무관. 모든 user 조회 가능 | 모든 caller (user context 의 user_id 와 무관) |
| `org` | owner_org 의 subtree 내 caller 만 조회 가능 | `caller.org_id ∈ caller.org_unit_ids` ∧ `caller.org_id` 가 `concept.x_devhub_owner_org_id` 의 subtree (`concept.x_devhub_owner_org_unit_ids`) 에 포함 |
| `personal` | owner_user 본인 만 조회 가능 | `caller.user_id == concept.x_devhub_owner_user_id` |
| `project` | owner_project 의 member 만 조회 가능 | `caller.user_id ∈ users JOIN project_members ON user_id WHERE project_id ∈ concept.x_devhub_owner_project_ids` |

**Default visibility**:
- **source plugin 자동 emit concept** (rule-based curator): `org` (bundle owner org 기준 자동 expand)
- **human 작성 concept** (manual edit): `org` default, manual override 가능

**Curation permission** (write/edit, `PUT /concepts/{id}`):

> **§3.9 lifecycle 정합**: 본 §3.6.2 의 curation permission 은 `x_devhub_status: reviewed` 또는 `published` state 의 concept 에만 적용 (§3.9.1 lifecycle 5 단계 정합). `created` state 의 concept 는 자동 normalize 의 부산물로 curator 직접 control 불가. `archived` state 의 concept 는 write 불가 (superseded 또는 obsolete).

```python
def check_curation_permission(concept, caller_context):
    curator = concept.frontmatter.get("x_devhub_curator", "rule-based")

    if curator == "rule-based":
        # rule-based 자동 생성, manual edit 불가 (source plugin 재실행 시 overwrite 됨)
        return deny("E_FORBIDDEN", "rule-based curator 는 manual edit 불가")

    if curator == "llm":
        # LLM 자동 생성 (M-v0.2.3+), manual edit 시 curator="human" 으로 승격
        if "system_admin" in caller_context.roles:
            return allow_with_promotion("human")
        return deny("E_FORBIDDEN", "llm curator manual edit 은 system_admin 만")

    if curator == "human":
        # manual edit 가능, caller 의 ownership check
        owner_user = concept.frontmatter.get("x_devhub_owner_user_id")
        owner_org_units = concept.frontmatter.get("x_devhub_owner_org_unit_ids", [])

        if owner_user and caller_context.user_id == owner_user:
            return allow()
        if any(uid in caller_context.org_unit_ids for uid in owner_org_units):
            return allow()  # org_head scope
        if "system_admin" in caller_context.roles:
            return allow()
        return deny("E_FORBIDDEN", "auth.curation_denied")

    return deny("E_INTERNAL", f"unknown curator type: {curator}")
```

#### 3.6.3 Query scope priority (org > personal > project > public)

**4-tier priority** (사용자 결정, 2026-06-18, "사용자 조직 > 개인 데이터 > 소속 프로젝트" 의 M-v0.2.0 PoC 정합):

```
Priority 1: org scope      (caller.org_id 가 concept.x_devhub_owner_org_id 의 subtree)
Priority 2: personal scope (caller.user_id == concept.x_devhub_owner_user_id)
Priority 3: project scope  (caller.user_id ∈ concept.x_devhub_owner_project_ids 의 members)
Priority 4: public scope   (모든 caller)
```

**Filter algorithm** (`POST /query`):

```python
def filter_concepts(query_result, caller_context):
    visible = []  # list of (concept, priority)
    for concept in query_result:
        visibility = concept.frontmatter.get("x_devhub_visibility", "public")

        if visibility == "public":
            visible.append((concept, priority=4))
        elif visibility == "org":
            owner_org_units = set(concept.frontmatter.get("x_devhub_owner_org_unit_ids", []))
            if owner_org_units.intersection(set(caller_context.org_unit_ids)):
                visible.append((concept, priority=1))
            # else: hidden
        elif visibility == "personal":
            owner_user = concept.frontmatter.get("x_devhub_owner_user_id")
            if owner_user and owner_user == caller_context.user_id:
                visible.append((concept, priority=2))
            # else: hidden
        elif visibility == "project":
            owner_projects = set(concept.frontmatter.get("x_devhub_owner_project_ids", []))
            if owner_projects.intersection(set(caller_context.project_ids)):
                visible.append((concept, priority=3))
            # else: hidden

    # Sort by priority (1 = highest), stable sort
    visible.sort(key=lambda x: x[1])
    return [c for c, _ in visible]
```

**Aggregation policy**: 같은 concept (e.g., `gitea_repo_puller`) 가 여러 bundle 에 존재하는 경우:
- 각 bundle 의 instance 별로 (visibility, priority) 가 다를 수 있음.
- **가장 높은 priority 의 instance 만 노출** (priority 1 > 2 > 3 > 4).
- 예: user 가 `gitea_repo_puller` query 시:
  - (a) devhub-gitea 의 `org` visibility instance (priority 1, owner_org_unit_ids 에 caller 포함) → 노출
  - (b) 다른 외부 시스템 mirror 의 `public` instance (priority 4) → hidden
  - 결과: caller 는 1개 instance 만 봄 (priority 1).

**Edge cases**:
- Caller 가 priority 1 instance 의 owner_org_unit_ids 에 속하지 않으나 priority 2 (personal) instance 가 자기 것인 경우: 둘 다 노출 (priority 별로), highest 먼저.
- 같은 priority 의 여러 instance: bundle 이름 alphabetical 순.

#### 3.6.4 Frontmatter extension (5개 field 추가)

§3.3 frontmatter spec 의 DevHub 확장 prefix `x_devhub_*` 에 5 field 추가:

| Field | Type | 정의 | Default (자동) |
| --- | --- | --- | --- |
| `x_devhub_owner_org_id` | string (FK → org_units.unit_id) | concept 의 primary owner org | bundle owner org (source plugin 자동 emit) |
| `x_devhub_owner_user_id` | string (FK → users.user_id, nullable) | human curator 의 경우 명시 | `null` (rule-based / llm curator) |
| `x_devhub_owner_org_unit_ids` | array of string | org_head scope 의 subtree (recursive expand from owner_org_id) | `[x_devhub_owner_org_id]` |
| `x_devhub_owner_project_ids` | array of string | project visibility 의 경우 project list | `[]` |
| `x_devhub_visibility` | enum: `org` \| `personal` \| `project` \| `public` | 조회 scope 등급 | `org` (source plugin 자동) / `org` (human 작성 default) |

**§3.3 정책 노트** 에 다음 추가 (cross-section 정합 fix §3.6.5 의 일환):

> **Path Y governance 필드 (2026-06-18 신규)**: `x_devhub_owner_org_id` / `x_devhub_owner_user_id` / `x_devhub_owner_org_unit_ids` / `x_devhub_owner_project_ids` / `x_devhub_visibility` 5 field 추가 (§3.6.2 / §3.6.4 정합). OKF spec 의 "extra keys 자유" 원칙 정합, vendor-neutral 유지. caller-provided user context (§3.6.1) 와 함께 backend-knowledge 의 governance 모델의 핵심.

#### 3.6.5 Cross-section 정합 fix

본 §3.6 신규에 따른 정합 fix 위치:

1. **§1.2 G7 standalone 정책** — "backend-knowledge 는 auth 자체 안 함 + caller 가 user context 전달하면 그 context 로 governance 만 수행 (§3.6.1 정합)" 1줄 추가.
2. **§2.3 시스템 경계 표의 "API 인증" row** — "internal-only, no auth + **caller 가 X-DevHub-User-Context header 로 user context 전달 시 governance 수행 (§3.6.1)**".
3. **§2.3 정책 표의 "다른 backend 연결 (general)" row** — "외부 시스템 + caller-provided user context 만 처리. backend-core 의 users/org_units/projects 어떤 layer 도 호출 ❌ (caller 가 context 구성해서 전달, §3.6.1)".
4. **§3.1 API 매트릭스** — 인증 정책 노트에 Path Y 추가 + §3.6.1 표 cross-reference.
5. **§3.3 frontmatter spec 정책 노트** — Path Y governance 필드 5종 추가 (§3.6.4 정합).
6. **§4.1 정책 정의 표 "인증" row** — Path Y caller-provided user context 명시 + §3.6.1 cross-reference.
7. **ADR-0034 §4.3 영향** — §3.6 row 추가.
8. **ADR-0035 §3.4 1차 raw API 정책 row** — caller-provided user context (gateway 책임) 명시 + §3.6.1 cross-reference + ADR-0035 §4.2/§4.3 갱신.

### 3.7 Data normalization pipeline (category × system → OKF concept, 2026-06-18 신규)

**Motivation**: §3.5 의 5 카테고리 + 8 type 결정 + §3.6 의 data governance 이후 자연스러운 후속. 외부 시스템 데이터를 (category, system) 단위로 수집한 후, 이를 OKF concept 형으로 **정규화**하여 **cross-source aggregation 가능**하게 하는 파이프라인 정의. 사용자 2026-06-18 결정.

**DevHub 기존 패턴 참고**: backend-core 의 Go adapter 패턴이 본 §3.7 의 reference:
- `HomeLabRawSnapshot` → `HomeLabSnapshot` (transport → normalized, [`backend-core/internal/integrations/adapters/contract.go`](../../backend-core/internal/integrations/adapters/contract.go))
- `stateToEventType()` (Gitea PR state → enum, [`gitea_pull.go`](../../backend-core/internal/integrations/adapters/gitea_pull.go))
- `collectDegradedProviders()` (health → degraded flag)
- `REQ-FR-INT-004/005/008/009` (provider → 공통 model 정규화 요구사항, [`integration-registry/requirements.md`](../domain/integration-registry/requirements.md))

**독립 backend-knowledge 구현 차이**: Python `sources/{source}.py` 가 source plugin ABC (`sources/_base.py` §2.1 정합) 의 subclass 로 구현. backend-core 의 Go adapter 코드/로직 import ❌ (standalone 정합, §1.2 G7). **외부 시스템 공식 API spec 만 보고 0에서 Python 작성** (§1.2 G3 정합, 외부 시스템 API spec 만 참조).

#### 3.7.1 Normalization principles

**단일 shape = OKF concept** (1 .md = 1 concept, ADR-0034 §3.1 정합). 모든 외부 시스템 데이터는 source plugin 을 통해 OKF concept 형으로 정규화.

**책임 분리** (5 module):

| Module | 책임 | 위치 |
| --- | --- | --- |
| **source plugin (adapter)** | 외부 시스템 API 호출 + raw JSON 파싱 + OKF concept emit | `backend-knowledge/sources/{source}.py` |
| **1차 raw storage** | raw JSON 봉투 암호화 저장 (ADR-0025 정합) | `backend-knowledge/var/raw/{source}/{slug}.json` |
| **OKF bundle storage** | concept .md 저장 (git 가능, Markdown + frontmatter) | `backend-knowledge/var/bundles/{bundle}/{category}/{slug}.md` |
| **rule-based enricher** | frontmatter 직접 read + raw → concept 변환 (1차 rule-based) | `backend-knowledge/curate/enricher.py` |
| **index_builder** | per-bundle/per-category/per-type index.md 자동 생성 (§3.5.4 정공법) | `backend-knowledge/curate/index_builder.py` |
| **link_resolver** | cross-link 추적 + reverse index (§3.5.5 정공법) | `backend-knowledge/curate/link_resolver.py` |

**5 step normalization** (POST /ingest/{source}/sync 호출 시 1 cycle):

```
Step 1: 외부 시스템 API 호출 (source plugin 이 credential config 로 연결, §2.2 봉투 암호화 정합)
  ↓
Step 2: raw JSON 응답 저장 (var/raw/{source}/{slug}.json, 봉투 암호화, ADR-0025)
  ↓
Step 3: raw → OKF concept 변환 (sources/{source}.py 의 normalize() method, §3.7.4 정공법)
  ↓
Step 4: concept .md 파일 emit (var/bundles/{bundle}/{category}/{slug}.md)
  ↓
Step 5: index.md 자동 생성 (curate/index_builder.py 가 per-bundle/per-category/per-type)
```

**M-v0.2.0 PoC**: 1차 rule-based enrich (LLM 미사용). **M-v0.2.3+**: Pi LLM enrich 활성화 (curate/enricher.py 의 `llm_enrich(concept)` method 추가, §6.3 정합).

#### 3.7.2 Per-source type mapping (7 source × types emitted)

§3.5.2 5×8 matrix + §1.2 G3 source plugin 7종 (Gitea 4 sub-plugin + homelab + metrics + hrdb) 의 cross-reference. 각 source 가 emit 하는 type:

| Source plugin | Bundle | Category | Emit types (representative) | Source 비고 |
| --- | --- | --- | --- | --- |
| `gitea_repo_pull` | `devhub-gitea` | `scm` | `integration`, `api_endpoint`, `metric`, `runbook`, `event`, `dataset`, `reference` | Gitea REST + git HTTP, 5종 PoC 의 1차 |
| `gitea_issue` | `devhub-gitea` | `issue_tracker` | `integration`, `api_endpoint`, `event`, `runbook`, `metric`, `dataset`, `reference` | Gitea Issue API, 5종 PoC 의 1차 |
| `gitea_wiki` | `devhub-gitea` | `wiki` | `integration`, `api_endpoint`, `reference`, `dataset`, `metric` | Gitea Wiki API, 5종 PoC 의 1차 |
| `gitea_action` | `devhub-gitea` | `cicd` | `integration`, `api_endpoint`, `event`, `metric`, `runbook`, `dataset`, `reference` | Gitea Actions API, 5종 PoC 의 1차 |
| `homelab_mock` | `devhub-homelab` | (5 카테고리 외) | `reference`, `metric`, `runbook`, `api_endpoint` | M-v0.2.0 PoC, filesystem fixture (`var/fixtures/homelab/*.json`) |
| `metrics` | `devhub-metrics` | (5 카테고리 외) | `metric`, `reference`, `api_endpoint` | M-v0.2.2 운영 wire, Prometheus scrape API |
| `hrdb` | `devhub-hrdb` | (5 카테고리 외) | `dataset`, `reference`, `metric` | M-v0.2.3 운영 wire, 사내 HR DB PostgreSQL |

**5종 PoC source plugin** (M-v0.2.0) = Gitea 4 sub-plugin + homelab_mock = **5 source × 평균 7 type** = 약 **35개 concept** 자동 emit (cross-section §3.5.3 의 devhub-gitea 25 + devhub-homelab 4 정합).

**Type emit 여부 결정**: source plugin 의 `SourceMeta.emit_types: List[ConceptType]` 필드. 어떤 type 을 emit 하는지는 source 의 외부 API 응답 schema 에 따라 결정 (e.g., Gitea Issue API 가 webhook payload 를 제공하면 `event` type emit, runbook 이 없으면 `runbook` type 미emit).

**작성 정공법**: source plugin 작성 시 §3.8 정공법 따름 (10 step 절차 + §3.8.1 SourcePlugin ABC + §3.8.5 3 tier 검증).

#### 3.7.3 Cross-source 동질화 (same type across multiple sources)

**같은 type 의 다른 source 의 concept 는 동일 shape** (5×8 matrix §3.5.2 정합 + 8 type enum §3.2 정합). 즉, **Jira issue / Gitea issue / GitHub issue 모두 `integration_*_issue_puller.md` 형으로 정규화**.

**Cross-source concept 예시** (`integration` type):

| Source | Bundle | Category | Concept slug | Concept frontmatter |
| --- | --- | --- | --- | --- |
| `gitea_issue` | `devhub-gitea` | `issue_tracker` | `integration_gitea_issue_puller.md` | `type: integration, x_devhub_source: gitea_issue, x_devhub_category: issue_tracker` |
| (향후) `jira_issue` | `devhub-jira` | `issue_tracker` | `integration_jira_issue_puller.md` | `type: integration, x_devhub_source: jira_issue, x_devhub_category: issue_tracker` |
| (향후) `github_issue` | `devhub-github` | `issue_tracker` | `integration_github_issue_puller.md` | `type: integration, x_devhub_source: github_issue, x_devhub_category: issue_tracker` |

**Query 시 cross-source aggregation**:
- `GET /search?q=type:integration` → 모든 source 의 integration concept 동시 노출 (Gitea 1 + Jira 1 + GitHub 1 = 3+ instances)
- 같은 `integration` type 이지만 source 가 다름 → viz.html 에서 cross-source cluster 시각화 (e.g., `issue_tracker` category 의 integration sub-cluster 에 Gitea/Jira/GitHub 모두 포함)
- Cross-source link: `integration_gitea_issue_puller.md` 본문에서 `integration_jira_issue_puller.md` 로 cross-link 가능 (M-v0.2.3+ Pi LLM, §6.3 정합)

**Priority 기반 deduplication** (§3.6.3 aggregation policy 정합): 같은 `integration_*_issue_puller` 가 여러 source 에 있을 때, caller 의 user context 에 따라 highest priority instance 만 노출. priority = source 의 caller org_unit 정합도 (caller 가 owner_org 일치하면 priority 1, public 일 때 priority 4).

#### 3.7.4 Normalize algorithm (raw JSON → OKF concept .md)

```python
# sources/{source}.py
class SourceMeta:
    bundle: str
    category: str
    source_plugin: str
    concept_type: ConceptType  # 8종 enum
    title_field: str  # raw JSON 의 어떤 field 가 title
    description_field: str
    slug_field: str  # raw JSON 의 unique id (e.g., gitea 의 "id")
    body_template: Callable[[dict], str]
    link_fields: List[str]  # cross-link 추출할 raw field
    tags: List[str]
    bundle_owner_org_id: str
    bundle_owner_org_unit_ids: List[str]
    emit_types: List[ConceptType]

def normalize(raw_response: dict, source_meta: SourceMeta) -> Concept:
    """Step 1: Parse raw JSON"""
    parsed = parse(raw_response, source_meta.response_schema)
    # - missing field → degraded flag + audit log (§3.7.5)
    # - schema drift → KeyError catch + skip

    """Step 2: Extract frontmatter"""
    frontmatter = {
        "type": source_meta.concept_type.value,  # 8종 enum
        "title": parsed.get(source_meta.title_field, parsed.get("name", "Untitled"))[:200],
        "description": parsed.get(source_meta.description_field, "")[:500],
        "resource": parsed.get("url") or parsed.get("html_url", ""),
        "tags": list(set(source_meta.tags + [source_meta.bundle, source_meta.category])),
        "timestamp": parsed.get("updated_at", parsed.get("created_at", "")),
        # DevHub 확장 (§3.3)
        "x_devhub_source": source_meta.source_plugin,
        "x_devhub_raw_ref": f"raw://{source_meta.bundle}/{parsed[source_meta.slug_field]}",
        "x_devhub_bundle": source_meta.bundle,
        "x_devhub_version": 1,
        "x_devhub_curator": "rule-based",
        "x_devhub_category": source_meta.category,
        # §3.6.4 governance fields
        "x_devhub_owner_org_id": source_meta.bundle_owner_org_id,
        "x_devhub_owner_org_unit_ids": source_meta.bundle_owner_org_unit_ids,
        "x_devhub_owner_user_id": None,
        "x_devhub_owner_project_ids": [],
        "x_devhub_visibility": "org",
    }

    """Step 3: Emit body (Markdown 변환)"""
    body = source_meta.body_template(parsed)
    # body_template 예시 (per type):
    # - dataset:  ## Columns\n| name | type | description |\n
    # - runbook:  ## Steps\n1. ...\n
    # - metric:   ## Labels\n- key=value\n## Value\n- ...
    # - api_endpoint: ## Method\n`GET /api/v1/...`\n## Parameters\n- ...
    # - event:    ## Payload: <json payload object>\n

    """Step 4: Attach cross-links (resource URL, related concept)"""
    cross_links = []
    for field in source_meta.link_fields:
        if field in parsed:
            cross_links.append(extract_link(parsed[field], source_meta))
    for link in cross_links:
        body += f"\n- Related: [{link.title}]({link.path})\n"

    """Return OKF concept"""
    slug = f"{source_meta.concept_type.value}_{parsed[source_meta.slug_field]}"
    return Concept(
        frontmatter=frontmatter,
        body=body,
        slug=slug,
        path=f"var/bundles/{source_meta.bundle}/{source_meta.category}/{slug}.md",
    )
```

**Timing**: M-v0.2.0 = rule-based 만 (frontmatter 직접 read, LLM 미사용). M-v0.2.3+ = `curate/enricher.py` 의 `llm_enrich(concept)` method 추가 (Pi LLM 으로 body 보강 + cross-link 자동 resolution, §6.3 정합).

#### 3.7.5 Edge cases + degraded handling

| Edge case | 처리 | 코드 위치 | 운영 impact |
| --- | --- | --- | --- |
| **Partial failure** (일부 field normalize 실패, e.g., description 누락) | degraded flag + degraded_field list. concept emit 은 진행 (frontmatter 의 optional field 는 null). `x_devhub_degraded_fields` array 에 누락 field 기록 | `sources/{source}.py` normalize() | viz.html 에서 degraded 표시 + bundle 별 degraded metric 수집 |
| **Schema drift** (외부 시스템 API schema 변경, e.g., Gitea API v2 → v3) | normalize() 가 KeyError 발생 시 skip + audit log + 빈 degraded_field. **M-v0.3.0+ schema migration procedure 검토** (현재 out-of-scope) | `sources/{source}.py` normalize() | source plugin update 필요 + M-v0.3.0+ migration tool |
| **Source-specific custom transform** | source plugin 의 `_custom_transform(raw, parsed)` hook (override 가능). 각 source 의 특수 처리 (e.g., Gitea 의 nested `user.login` → flat, hrdb 의 joined `full_name` → split into first/last) | `sources/{source}.py` 의 normalize() override | source 별 code review 필요 |
| **Duplicate concept** (같은 slug 가 다른 source 에서 emit) | priority-based deduplication (§3.6.3 aggregation policy 정합). source_meta 의 `priority` 필드 (default = bundle 생성 순). viz.html 에서 dedup 표시 | `curate/index_builder.py` | M-v0.2.3+ multi-source 1차 적용 |
| **Large raw** (e.g., homelab 의 1MB JSON) | frontmatter 에는 summary (max 500 char), body 는 truncated + `x_devhub_raw_ref` link. `body_max_length: int = 10000` default | normalize() 의 emit_body 단계 | bundle size 관리 |
| **Auth failure** (credential expired, network timeout) | source plugin raise `SourceAuthError` → `POST /ingest/{source}/sync` returns 502 `E_FORBIDDEN` + audit log | `sources/_base.py` SourcePlugin ABC | M-v0.2.1+ e2e smoke alert |

**M-v0.2.0 PoC 범위**: Partial failure + Source-specific custom transform 만 구현. Schema drift + Duplicate + Large raw + Auth failure 는 M-v0.2.1+ 검토.

### 3.8 Source plugin 작성 정공법 (How to write a source plugin, 2026-06-18 신규)

**Motivation**: §3.7 의 abstract 5 step normalization pipeline + [ADR-0035 §3.2](../adr/0035-backend-knowledge-creation.md) + §6.4 의 high-level 결정만 있고, **실제 M-v0.2.0 PoC 의 5종 source plugin (Gitea 4 sub-plugin + homelab_mock) 작성 시 따라야 할 정공법 + 신규 source 추가 절차 + source plugin 검증 절차** 부재. 본 §3.8 가 그 정공법을 정의.

**독립 backend-core 정합**: backend-core 의 Go adapter 패턴 (port + adapter, [`external-integrations-agentic-rag-roadmap.md` §1.2](./external-integrations-agentic-rag-roadmap.md)) 과 같은 **port/adapter 구조**. 차이점: Python `SourcePlugin ABC` (단순 상속) vs Go interface (composition). **로직 import ❌, 외부 시스템 공식 API spec 만 참조** (§1.2 G3 정합).

#### 3.8.1 Source plugin ABC (sources/_base.py) 인터페이스 명세

**파일 위치**: `backend-knowledge/sources/_base.py`

```python
from abc import ABC, abstractmethod
from pydantic import BaseModel, Field
from typing import List, Dict, Any, Optional
from datetime import datetime

# Credential schema (§2.2 봉투 암호화 정합, type-agnostic string)
class Credential(BaseModel):
    type: str = Field(..., description="credential type: 'basic' | 'bearer' | 'api_key' | 'oauth2'")
    value: str = Field(..., description="type-agnostic string (id/pw or token/bearer/api_key)")

# Source meta (§3.7.4 + §10.4 정합)
class SourceMeta(BaseModel):
    bundle: str
    category: str
    source_plugin: str
    emit_types: List[str]
    title_field: str
    description_field: str
    slug_field: str
    body_template: Optional[str] = None
    link_fields: List[str] = []
    tags: List[str] = []
    bundle_owner_org_id: str
    bundle_owner_org_unit_ids: List[str] = []
    # §10.4 storage_mode + normalize_mode (2026-06-18 신규)
    storage_mode: Literal["file", "db"] = "file"
    normalize_mode: Literal["rule-based", "pi-sdk", "pi-rpc"] = "rule-based"
    ingest_schedule: Optional[str] = None  # cron expression, file mode 는 미설정

# Connection state
class Connection(BaseModel):
    base_url: str
    authenticated: bool
    connected_at: datetime
    credential_type: str  # never store raw credential

# Raw response wrapper
class RawResponse(BaseModel):
    data: Dict[str, Any]
    received_at: datetime
    source_plugin: str

# Concept (OKF format, §3.3 / §3.6.4 정합)
class Concept(BaseModel):
    frontmatter: Dict[str, Any]
    body: str
    slug: str
    path: str

# Fetch query
class FetchQuery(BaseModel):
    since: Optional[datetime] = None
    limit: int = 100
    filter: Dict[str, Any] = {}

# Health status
class HealthStatus(BaseModel):
    healthy: bool
    last_sync: Optional[datetime] = None
    last_error: Optional[str] = None
    source_plugin: str

# Source plugin ABC
class SourcePlugin(ABC):
    meta: SourceMeta

    @abstractmethod
    def connect(self, credential: Credential) -> Connection: ...
    """Step 1 정합: 외부 시스템 연결. credential 은 type-agnostic string."""

    @abstractmethod
    def fetch(self, query: FetchQuery) -> RawResponse: ...
    """외부 시스템 API 호출. M-v0.2.0 = real wire (Gitea 1 instance) 또는 filesystem fixture (homelab_mock)."""

    @abstractmethod
    def normalize(self, raw: RawResponse) -> List[Concept]: ...
    """Step 3 정합: raw → OKF concept 변환. §3.7.4 pseudocode 의 normalize() 4 step."""

    @abstractmethod
    def emit_concept(self, concept: Concept) -> None: ...
    """Step 4 정합: concept .md 파일 emit. var/bundles/{bundle}/{category}/{slug}.md."""

    @abstractmethod
    def health_check(self) -> HealthStatus: ...
    """source plugin 상태 확인. GET /api/v0-2/ingest/{source}/status 의 응답."""

# Plugin registry
_REGISTRY: Dict[str, Type[SourcePlugin]] = {}

def register(source_name: str, plugin_class: Type[SourcePlugin]) -> None:
    """신규 source plugin 등록. plugin_class 는 SourcePlugin ABC 의 subclass."""
    _REGISTRY[source_name] = plugin_class

def get_plugin(source_name: str) -> SourcePlugin:
    """registry 에서 plugin instance 반환."""
    return _REGISTRY[source_name]()

def list_plugins() -> List[str]:
    return list(_REGISTRY.keys())
```

**§2.1 `sources/` 트리 정합**:
- `_base.py` = SourcePlugin ABC + Credential + SourceMeta + Connection + RawResponse + Concept + FetchQuery + HealthStatus + registry (12 type)
- `{source}.py` = SourcePlugin ABC 의 subclass (per source)

#### 3.8.2 Gitea 4 sub-plugin 작성 정공법

Gitea 1 instance 의 4 sub-plugin (gitea_repo_pull / gitea_issue / gitea_wiki / gitea_action) 작성 정공법. **모두 `devhub-gitea` bundle + 4개 category directory** (§3.5.3 정합). **M-v0.2.0 부터 real wire** (mock 없이 실제 Gitea 1 instance PoC, §6.1 / §6.4 정합).

| Sub-plugin | SourceMeta | 외부 API | Emit types | Credential | 비고 |
| --- | --- | --- | --- | --- | --- |
| `gitea_repo_pull` | bundle=devhub-gitea, category=scm | Gitea REST `/api/v1/repos/{owner}/{repo}` + git HTTP `/info/refs` | 7 type (integration / api_endpoint / metric / runbook / event / dataset / reference) | `type=bearer, value=<Gitea access token>` | repo list + metadata + git refs, 5종 PoC 의 1차 |
| `gitea_issue` | bundle=devhub-gitea, category=issue_tracker | Gitea REST `/api/v1/repos/{owner}/{repo}/issues` | 7 type (integration / api_endpoint / event / runbook / metric / dataset / reference) | `type=bearer, value=<Gitea access token>` | issue list + comment + label, 5종 PoC 의 1차 |
| `gitea_wiki` | bundle=devhub-gitea, category=wiki | Gitea REST `/api/v1/repos/{owner}/{repo}/wiki` | 5 type (integration / api_endpoint / reference / dataset / metric) | `type=bearer, value=<Gitea access token>` | wiki page + history, 5종 PoC 의 1차 |
| `gitea_action` | bundle=devhub-gitea, category=cicd | Gitea REST `/api/v1/repos/{owner}/{repo}/actions` | 7 type (integration / api_endpoint / event / metric / runbook / dataset / reference) | `type=bearer, value=<Gitea access token>` | action run + workflow, 5종 PoC 의 1차 |

**Gitea access token 권한**: 각 sub-plugin 별 필요한 read scope:
- `gitea_repo_pull`: `read:repository` + `read:user`
- `gitea_issue`: `read:issue` + `read:user`
- `gitea_wiki`: `read:repository` (wiki 는 repo 의 sub-resource)
- `gitea_action`: `read:repository` (actions 도 repo 의 sub-resource)

**공통 인증 endpoint**: `POST /api/v0-2/ingest/{source}/sync` 호출 시 caller (gateway / agent) 가 Gitea URL + access token 을 source plugin 의 `connect(credential)` method 에 전달. credential 은 메모리에서만 사용, **저장 ❌** (봉투 암호화 후 sqlite metadata `source_credentials` table 에 저장 시 ADR-0025 정합).

**공통 normalize() 패턴** (4 sub-plugin 모두):

```python
class GiteaSubPlugin(SourcePlugin):
    def connect(self, credential: Credential) -> Connection:
        # Gitea access token 으로 /api/v1/user endpoint 호출하여 인증 확인
        response = httpx.get(
            f"{credential.value}/api/v1/user",  # credential.value 가 base_url + token
            headers={"Authorization": f"token {token}"}
        )
        response.raise_for_status()
        return Connection(
            base_url=credential.value,
            authenticated=True,
            connected_at=datetime.now(),
            credential_type="bearer",
        )

    def fetch(self, query: FetchQuery) -> RawResponse:
        # Gitea REST API 호출 (per sub-plugin)
        ...

    def normalize(self, raw: RawResponse) -> List[Concept]:
        # §3.7.4 pseudocode 정합
        # - parse: raw.data 의 list 를 iterate
        # - extract frontmatter: per-type field mapping
        # - emit body: per-type body_template
        # - attach cross-links: resource URL + related concept
        ...

    def emit_concept(self, concept: Concept) -> None:
        # var/bundles/{meta.bundle}/{meta.category}/{concept.slug}.md 파일 쓰기
        ...

    def health_check(self) -> HealthStatus:
        # 마지막 sync 시각 + last error
        ...
```

#### 3.8.3 homelab_mock 작성 정공법

`homelab_mock` 은 M-v0.2.0 PoC = filesystem fixture 기반 mock source plugin. M-v0.2.1 = real wire (real homelab agent 호출) 로 교체.

**SourceMeta**:
- bundle=devhub-homelab
- category=(5 카테고리 외, x_devhub_category 미설정)
- source_plugin=homelab_mock
- emit_types=[reference, metric, runbook, api_endpoint]
- credential_type=`mock` (no real credential, M-v0.2.0 PoC)

**Filesystem fixture**:
- 위치: `backend-knowledge/var/fixtures/homelab/*.json` (per fixture)
- 예시: `node_inventory.json`, `pull_metrics.json`, `recovery_runbook.json`, `node_api.json`
- 각 fixture = 1 raw JSON, mock 의 normalize() 가 1 concept emit

**normalize() 정공법**:
```python
class HomelabMockPlugin(SourcePlugin):
    def fetch(self, query: FetchQuery) -> RawResponse:
        # var/fixtures/homelab/*.json 파일들을 glob 으로 read
        fixtures = glob("var/fixtures/homelab/*.json")
        return RawResponse(
            data={"fixtures": [json.load(open(f)) for f in fixtures]},
            received_at=datetime.now(),
            source_plugin="homelab_mock",
        )

    def normalize(self, raw: RawResponse) -> List[Concept]:
        concepts = []
        for fixture in raw.data["fixtures"]:
            # fixture 별로 1 concept emit (M-v0.2.0 PoC 단순화)
            concept = self._fixture_to_concept(fixture)
            concepts.append(concept)
        return concepts
```

**M-v0.2.0 PoC 범위**: filesystem fixture 만. M-v0.2.1 = `homelab.py` 로 교체 (real wire, HTTP API 호출 + 5종 PoC 의 5번째 운영 source).

#### 3.8.4 신규 source 추가 절차 (10 step)

**외부 시스템 5 카테고리 외 시스템 추가** (e.g., Grafana, Slack, PagerDuty) 또는 **5 카테고리 내 multi-vendor** (e.g., Jira, GitHub Issues, GitLab) 시 다음 10 step 절차:

```
Step 1: 외부 시스템 API spec 1차 정독
  - vendor 공식 docs / OpenAPI / GraphQL schema
  - 인증 방식 (OAuth2 / API key / basic auth / service account)
  - rate limit + pagination 정책
  - webhook 지원 여부 (event type source 시)

Step 2: SourceMeta 정의
  - bundle: {bundle-name} (소문자 + kebab-case)
  - category: 5 enum 중 1 (or 5 카테고리 외)
  - source_plugin: {bundle-name}_{vendor} or {vendor}
  - emit_types: 8 type enum 중 subset
  - credential_schema: Pydantic v2 model
  - **§10.4 storage_mode 결정**: `file` (default) / `db` (Pi-driven 정규화 필요 시)
  - **§10.4 normalize_mode 결정**: `rule-based` (default) / `pi-sdk` (M-v0.2.0 PoC DB path) / `pi-rpc` (M-v0.2.3+ 옵션)
  - **§10.4 ingest_schedule 결정**: cron expression (db mode 일 때만, default `*/5 * * * *`)

Step 3: SourcePlugin ABC 의 5 method 구현
  - connect(credential) → Connection (외부 시스템 인증 + base_url)
  - fetch(query) → RawResponse (외부 시스템 API 호출)
  - normalize(raw) → List[Concept] (§3.7.4 pseudocode 정합)
  - emit_concept(concept) → None (var/bundles/{bundle}/{category}/{slug}.md)
  - health_check() → HealthStatus

Step 4: credential schema Pydantic 모델 작성
  - source 의 sources/{source}.py 내 _credential_schema: Type[BaseModel]
  - id/pw or token type-agnostic string (§2.2 정합)

Step 5: normalize() method 의 body_template 작성 (per emit_type)
  - §3.7.4 의 5 type 별 예시 (dataset/runbook/metric/api_endpoint/event)
  - source 별 특수 처리 (custom transform hook)

Step 6: 단위 테스트 작성 (pytest, mock external API)
  - test_connect: mock HTTP response (responses or httpx_mock)
  - test_fetch: mock API response → RawResponse
  - test_normalize: RawResponse → Concept list (frontmatter + body 검증)
  - test_emit_concept: file write 검증 (tmpdir)
  - test_health_check: 정상/실패 케이스

Step 7: e2e smoke (ingest → curate → query happy path)
  - 실제 외부 시스템 연결 또는 mock instance
  - POST /api/v0-2/ingest/{source}/sync → raw + concept emit
  - GET /api/v0-2/bundles/{bundle}/index.md → 자동 갱신 확인
  - GET /api/v0-2/concepts/{type}/{name} → query 검증

Step 8: bundle 디렉터리 layout 결정 + §3.7.2 per-source mapping table 업데이트
  - bundle = 1 외부 시스템 단위 (or cross-cutting 주제)
  - category directory per x_devhub_category
  - umbrella doc §3.7.2 표 에 row 추가

Step 9: representative concept .md 발췌 작성
  - §3.5.3 / §3.6.2 / **§3.9 lifecycle** 정합: 5 카테고리별 1+ concept frontmatter 예시
  - bundle owner org + governance field 채움
  - **§3.9.2 frontmatter template (per 8 type) 의 권장 field 모두 채움** (e.g., `dataset` type → `x_devhub_table_name`, `x_devhub_columns` 등)
  - **§3.9.3 review checklist 의 1~4 항목 자동 validate 통과** (5번 cross-link 은 optional)

Step 10: ADR-0034 / ADR-0035 영향 section 갱신 (선택, 영향 시)
  - 새 type 추가 시: ADR-0034 §3.2 type enum 표 갱신
  - 새 governance 정책 시: ADR-0035 §3.3 / §4 갱신
  - 영향 없는 경우 skip (default)
```

**Quality gate** (Step 10 까지 완료 후):
- [ ] Linter (ruff, mypy) 통과
- [ ] 단위 테스트 100% (또는 명시적 skip 사유)
- [ ] e2e smoke 통과 (M-v0.2.0 = Gitea 1 instance)
- [ ] §3.7.2 per-source mapping table 업데이트
- [ ] umbrella doc §9 변경 이력 row 추가

#### 3.8.5 Source plugin 검증

**3 tier 검증**:

| Tier | 범위 | 도구 | 시점 |
| --- | --- | --- | --- |
| **단위 테스트** | SourcePlugin ABC 의 5 method + normalize() 의 4 step (§3.7.4) + emit_concept 의 file write | pytest + responses (mock HTTP) + tmpdir | M-v0.2.0 PoC PR 마다 |
| **통합 테스트** | 실제 외부 시스템 연결 (Gitea 1 instance) | httpx + e2e fixture | M-v0.2.0 sprint 진입 시 + 이후 주요 변경 시 |
| **e2e smoke** | ingest → curate → query happy path (5종 PoC source, 1 Gitea instance) | pytest + FastAPI TestClient | M-v0.2.1+ CI e2e lane (release pipeline) |

**Source plugin health check** (실 운영):
- `GET /api/v0-2/ingest/{source}/status` (§3.1 API 매트릭스 정합) — `health_check()` method 결과 반환
- 응답: `{envelope, data: {healthy, last_sync, last_error, source_plugin}}`
- 비정상 source = M-v0.2.1+ alert (caller 의 audit + notify)

**M-v0.2.0 PoC 범위**:
- 5종 PoC source plugin = Gitea 4 sub-plugin + homelab_mock
- 모두 §3.8.2 + §3.8.3 정공법 따름
- §3.8.4 의 10 step 중 Step 1~7 필수, Step 8~9 는 §3.5.3 / §3.7.2 정합, Step 10 은 optional
- §3.8.5 의 3 tier 검증 중 단위 테스트 + e2e smoke (M-v0.2.0 = 5종 source, 1 Gitea instance)

### 3.9 OKF concept 운영 lifecycle (Created → Reviewed → Published → Active → Archived, 2026-06-18 신규)

**Motivation**: §3.5 (concept organization) + §3.6 (governance) + §3.7 (normalization) + §3.8 (source plugin 작성) 완료 후, **concept 의 lifecycle 운영 정공법** 부재. 1 concept .md 가 어떤 단계를 거쳐 운영되는지, 각 단계의 책임자/정책/audit, archive 정책 등 정의. **M-v0.2.0 PoC = rule-based 자동 publish (frontend 0 page)**, **M-v0.2.1+ = human 작성 + review workflow** (frontend 관리/조회 page).

#### 3.9.1 Lifecycle 5 단계 state machine

```
        created
           ↓
       [Reviewed] ←─────┐
           ↓             │
       Published         │
           ↓             │ (rejected)
        Active ──────────┘
           ↓
       Archived
```

| State | 정의 | 진입 조건 | 책임자 | 정합 section |
| --- | --- | --- | --- | --- |
| **created** | concept .md 파일이 처음 disk 에 쓰여진 상태 (frontmatter + body 존재, validate 전) | rule-based: source plugin 의 normalize() + emit_concept() 완료 / human: frontend "new concept" form 제출 | rule-based: source plugin / human: frontend user | §3.7.4 + §3.8.4 Step 9 |
| **reviewed** | frontmatter + body 검증 통과 (또는 skip) | rule-based: source plugin 자동 emit 시 즉시 reviewed (validate 통과 시) / human: reviewer 승인 (M-v0.2.1+) | rule-based: source plugin (auto) / human: reviewer (system_admin or org_head) | §3.9.3 review checklist |
| **published** | bundle/index.md + viz.html graph 에 노출 | reviewed 통과 시 자동 publish (rule-based) OR 수동 publish 버튼 (human, M-v0.2.1+) | 자동 | §3.5.4 index.md + §3.5.5 viz.html |
| **active** | published 후 정상 운영 상태 (조회 가능) | published 직후 자동 진입 | 자동 | §3.6.3 query scope priority + §3.6 visibility 정합 |
| **archived** | superseded or obsolete. 조회 시 hidden (viz.html 에서도 hidden). raw 는 유지 (§4.6 soft_archive 정합) | superseded (raw 변경 시 자동, §4.6 정합) OR 운영자 수동 결정 (M-v0.2.1+) | 자동 (raw 변경) + 수동 (operator) | §4.6 raw 삭제 시 concept 처리 |

**frontmatter status field** (lifecycle state 명시):
- M-v0.2.0 PoC = status field 미사용 (5 state 가 sqlite 의 `concept_index.status` column 으로만 추적, file 변경 없음)
- M-v0.2.1+ = `x_devhub_status: created|reviewed|published|active|archived` frontmatter field 추가 (5 state 명시 + viz.html / frontend 표시용)

**Default lifecycle 단축** (rule-based 자동화):
- rule-based source plugin: created → reviewed → published → active 가 source plugin 의 `POST /api/v0-2/ingest/{source}/sync` 완료 시 1 cycle 내 자동 진행
- human 작성: created → reviewed (수동) → published (수동) → active (자동)
- archive: superseded 자동 OR 운영자 수동 결정

#### 3.9.2 Frontmatter template (per 8 type)

§3.3 frontmatter spec 의 `type` field 별 권장 frontmatter template. concept 작성 시 (rule-based: normalize() / human: frontend form) 다음 template 참고.

| Type | 필수 field | 권장 field | 예시 |
| --- | --- | --- | --- |
| **`dataset`** | `type: dataset, title, x_devhub_source, x_devhub_bundle, x_devhub_category` | `x_devhub_table_name` (외부 DB table), `x_devhub_columns: [{name, type, description}]`, `x_devhub_primary_key` | `hrdb.persons` |
| **`metric`** | `type: metric, title, x_devhub_source, x_devhub_bundle, x_devhub_category` | `x_devhub_metric_type: counter\|gauge\|histogram\|summary`, `x_devhub_unit` (seconds/bytes/count), `x_devhub_labels: [key=value]` | `repo_kpi_sync_duration_seconds` |
| **`api_endpoint`** | `type: api_endpoint, title, x_devhub_source, x_devhub_bundle, x_devhub_category` | `x_devhub_method: GET\|POST\|...`, `x_devhub_path`, `x_devhub_params: [{name, type, required}]`, `x_devhub_response_schema: {ref}` | `gitea_api_v1_repos_list` |
| **`runbook`** | `type: runbook, title, x_devhub_source, x_devhub_bundle, x_devhub_category` | `x_devhub_trigger: condition`, `x_devhub_steps: [step]`, `x_devhub_rollback: [step]`, `x_devhub_owner_org_unit_ids` (runbook 책임 org) | `gitea_repo_pull_failure_recovery` |
| **`integration`** | `type: integration, title, x_devhub_source, x_devhub_bundle, x_devhub_category` | `x_devhub_external_system` (예: "Gitea v1.20"), `x_devhub_auth_type: bearer\|basic\|oauth2\|api_key`, `x_devhub_connection_config_ref` (raw link) | `homelab_file_puller` |
| **`event`** | `type: event, title, x_devhub_source, x_devhub_bundle, x_devhub_category` | `x_devhub_source_event` (예: "gitea.push"), `x_devhub_payload_schema: {ref}`, `x_devhub_trigger_event` | `gitea_push_event` |
| **`reference`** | `type: reference, title, x_devhub_source, x_devhub_bundle, x_devhub_category` | `x_devhub_vendor` (예: "Gitea"), `x_devhub_url` (vendor docs URL), `x_devhub_mirror_version` (raw link 의 mirror version) | `keycloak_admin_rest_api_v1` |
| **`decision`** | `type: decision, title, x_devhub_source, x_devhub_bundle, x_devhub_category` | `x_devhub_decision_status: proposed\|accepted\|deprecated`, `x_devhub_supersedes` (deprecated 시 이전 decision ID), `x_devhub_decision_date` | `decision_2026_06_18_gitea_pull_strategy` |

**공통 OKF 표준 field** (§3.3 정합):
- `type` (필수, 8 enum)
- `title` (human-readable)
- `description` (1-2 sentence 요약)
- `resource` (외부 resource URL, 있으면)
- `tags` (검색/필터)
- `timestamp` (마지막 갱신)

**공통 DevHub 확장 field** (§3.3 + §3.6.4 정합):
- `x_devhub_source`, `x_devhub_raw_ref`, `x_devhub_bundle`, `x_devhub_version`, `x_devhub_curator`, `x_devhub_category`
- `x_devhub_owner_org_id`, `x_devhub_owner_user_id`, `x_devhub_owner_org_unit_ids`, `x_devhub_owner_project_ids`, `x_devhub_visibility`

**§3.7.4 normalize() 정합**: source plugin 의 `SourceMeta.body_template` 가 본 §3.9.2 type 별 권장 field 의 일부분을 body 에 자동 emit (e.g., `dataset` type 의 columns → Markdown table).

#### 3.9.3 Review checklist

concept 가 reviewed state 로 진입 시 (M-v0.2.1+ human 작성 review OR rule-based 자동 validate) 다음 5 항목 체크리스트:

**1. Frontmatter validation** (필수):
- [ ] `type` ∈ 8 enum (§3.2)
- [ ] `x_devhub_category` ∈ 5 enum (또는 미설정, §3.2.1)
- [ ] `x_devhub_curator` ∈ {rule-based, llm, human} (§3.6.2)
- [ ] `x_devhub_visibility` ∈ {org, personal, project, public} (§3.6.4)
- [ ] `x_devhub_owner_org_id` / `x_devhub_owner_org_unit_ids` 일치 (org_unit_ids 가 owner_org_id 의 subtree 인지 validate)
- [ ] `x_devhub_version` ≥ 1
- [ ] `timestamp` 가 ISO 8601 형식

**2. Body validation** (권장):
- [ ] body 최소 길이 100자 (rule-based 자동 emit 의 경우 보통 만족)
- [ ] body 가 Markdown 형식 (heading / list / table 등)
- [ ] cross-link 의 target 이 실제 존재 (`okf/link_resolver.py` 의 reverse index, §3.5.5 정합)

**3. Governance validation** (필수, M-v0.2.1+):
- [ ] owner_user 가 source plugin 자동 emit 의 경우 null (rule-based) / human 작성의 경우 명시
- [ ] visibility 가 source plugin 자동 emit 의 경우 `org` / human 작성의 경우 명시 (default `org`)

**4. Bundle validation** (필수):
- [ ] bundle 디렉터리 layout 정합 (`var/bundles/{bundle}/{category}/{slug}.md`, §3.5.3 정합)
- [ ] slug 가 `{type}_{name}` 형식 (e.g., `integration_gitea_repo_puller`, §3.5.3 정합)
- [ ] file path 가 bundle owner_org 의 subtree 에 속함

**5. Cross-link validation** (권장):
- [ ] intra-bundle link 의 target 이 같은 bundle 내 존재
- [ ] cross-bundle link 가 명시적 의미가 있는 경우만 (§3.5.5 정공법)
- [ ] unresolved link 가 0개 (또는 명시적 사유)

**자동 validate (rule-based)**: source plugin 의 normalize() + emit_concept() 가 본 §3.9.3 의 1~4 항목 자동 validate. 실패 시 `x_devhub_status: created` 로 멈춤 (reviewed 진입 안 함). M-v0.2.0 PoC = 5번 (cross-link) 만 optional.

**수동 review (human 작성)**: frontend 의 "submit for review" 버튼 클릭 시 자동 validate + reviewer (system_admin OR org_head scope) 의 수동 승인. M-v0.2.1+ frontend 관리 page.

#### 3.9.4 Publish + archive 절차 + 운영 정책

**Publish 절차**:

| Trigger | 동작 | 시점 |
| --- | --- | --- |
| **rule-based 자동** | reviewed 통과 시 `curate/index_builder.py` 가 자동 publish → bundle/index.md + viz.html 갱신 | M-v0.2.0 PoC |
| **human 수동** | frontend "publish" 버튼 → reviewed → published state 전이 + index.md 갱신 | M-v0.2.1+ |
| **system_admin override** | admin 이 reviewed skip 후 publish 강제 (긴급 patch 시) | M-v0.2.0+ (API endpoint: `POST /api/v0-2/concepts/{id}/publish` with `?skip_review=true`) |

**Publish 시 side effect**:
- `curate/index_builder.py` 가 `{bundle}/index.md` + per-category `{bundle}/{category}/index.md` 갱신 (§3.5.4 정합)
- `okf/link_graph.py` 가 reverse index 갱신 (§3.5.5 정합)
- `sqlite concept_index` table 의 `published_at` timestamp 갱신
- viz.html 자동 SSR (M-v0.2.0 = viz.html 만, frontend page 없음)

**Archive 절차**:

| Trigger | 동작 | 시점 |
| --- | --- | --- |
| **superseded (raw 변경)** | source plugin 의 sync 가 새 version 의 concept emit 시 이전 version 의 concept `x_devhub_status: archived` 로 자동 archive (§4.6 raw 삭제 시 concept 처리 3 mode 정합) | M-v0.2.0 PoC (overwrite) / M-v0.2.1+ (superseded 명시) / M-v0.2.3+ (.md.prev history) |
| **obsolete (operator 결정)** | frontend "archive" 버튼 → 운영자 (system_admin OR owner_org_unit_ids 의 org_head) 의 수동 결정 | M-v0.2.1+ |
| **orphan (raw 삭제, §4.6)** | `x_devhub_status: orphaned` 자동 설정 + audit + bundle owner_org_unit_ids 내 caller 에 notify | M-v0.2.0+ |

**Archive 시 side effect**:
- `curate/index_builder.py` 가 archived concept 를 bundle/index.md 에서 제거 (active concept 만 표시)
- viz.html 에서 archived concept 의 node 색상 변경 (gray) 또는 hidden (default)
- raw 의 `concept_ids` 에서 archived concept ID 제거 (raw 는 유지, §4.4 retention 정합)

**운영 정책 (M-v0.2.0~v0.2.3+)**:

| Milestone | Lifecycle 지원 범위 |
| --- | --- |
| **M-v0.2.0 PoC** | rule-based 자동 publish 만. human 작성 없음 (frontend 0 page). archive 는 superseded 자동 (overwrite). viz.html 자가 viewer 만 SSR. |
| **M-v0.2.1** | human 작성 지원 (frontend 관리/조회 page 1 추가, §5.1 정합). review workflow (system_admin 또는 org_head scope). manual archive 버튼. |
| **M-v0.2.2** | soft_archive mode default (raw 변경 시 superseded 명시). review checklist 자동화 강화. `x_devhub_status` frontmatter field 추가. |
| **M-v0.2.3** | .md.prev history 보존. cross-bundle cross-link 자동화 강화. anomaly detection (이상 상태 concept 자동 flag). |
| **M-v0.3.0+** | AI-assisted review (Pi LLM enrichment, §6.3). multi-reviewer 승인 (e.g., system_admin + org_head scope 둘 다). |

**§3.7.5 edge cases 정합**:
- normalize() 실패 시 concept 는 `created` state 에서 멈춤 (reviewed 진입 안 함). `x_devhub_degraded_fields` + audit log.
- cross-link unresolved 시 concept `created` state 가능 (orphan link 허용), `reviewed` 진입 시 §3.9.3 의 5번 cross-link validation 으로 reject.

## 4. 1차 raw 데이터의 API 정책 (사용자 강조)

> **"1차 raw 데이터는 여타 백엔드의 데이터들과 동일하게 api를 통해서 조회하고 추가할 수 있어야 해."**

### 4.1 정책 정의

| 항목 | 정책 |
| --- | --- |
| **저장 위치** | `backend-knowledge/var/raw/{source}/{slug}.json` (file system, **봉투 암호화 후 git 가능**, ADR-0025 정합. raw 자체는 민감 정보일 수 있어 **민감 source 의 경우 .gitignore 권장**, §4.4 raw 운영 정책 + retention + storage quota 정합) **또는** `raw_records` table (DB, sqlite M-v0.2.0 PoC / PostgreSQL M-v0.2.3+, **per source `storage_mode: file\|db` field 로 분기**, §10.1 정합) |
| **메타** | sqlite `raw_index` table (id, source, slug, path, ingested_at, byte_size, sha256, visibility, retention_days, registered_by, concept_ids, last_verified_at) (§4.7 정합성 검증 field 추가) |
| **API** | `POST /api/v0-2/raw` + `GET /api/v0-2/raw/{type}/{name}` + `GET /api/v0-2/raw?source=...&since=...` (list) + `DELETE /api/v0-2/raw/{id}` (§4.5 endpoint 별 권한 + visibility 정합) **또는** `/api/v0-2/db/raw` 8 endpoint (DB path, §10.2 정합 — POST/GET/PATCH/DELETE/list/aggregate/search/ingest-status, SQL sort/filter/aggregate/search, Path Y caller-provided user context 필수) |
| **envelope** | **독립 정의** (자체, backend-core 의 `docs/api/conventions.md` 와 format 호환 유지, **import ❌**, cross-reference 만, §3.4 정합) |
| **인증** | **internal-only, no auth** + **Path Y caller-provided user context (2026-06-18 결정, §3.6.1 정합)** (gateway / firewall / IP allowlist 별도 보호, §2.3). OIDC ❌, Keycloak ❌, backend-core 인증 위임 ❌. caller 가 `X-DevHub-User-Context` header 전달 시 backend-knowledge 가 filter / curation ownership check 수행 (§3.6.1). **raw API 4 endpoint 모두 user context 필수 (§4.5 정합)** |
| **동기** | 동기 응답 (단일 raw concept 추가/조회). 비동기 sync 는 `/ingest/{source}/sync` 의 별도 endpoint |
| **idempotency** | `(source, slug)` unique. 중복 POST → 기존 id 반환 (201 대신 200) |
| **정합성** | sha256 hash 저장 + 매 조회 시 재검증 (§4.7). 불일치 시 `E_INTERNAL` ("raw.integrity_violation") + audit |
| **lifecycle** | raw 삭제 시 concept 처리 정책 (§4.6): M-v0.2.0 = hard_delete, M-v0.2.1+ = soft_archive (default). 1 raw → N concepts 관계 추적 (sqlite `raw_index.concept_ids`) |

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

> **standalone 유지 정책 (2026-06-17 결정)**: 모든 Phase 에서 `backend-knowledge` 는 **완전 standalone 시스템**. 다른 backend (backend-core / 다른 백엔드 / 다른 시스템) 와의 연결 / API 호출 / envelope / repository / 어떤 layer 든 공유 / import ❌. 외부 시스템 **7종** source 만 단방향 (M-v0.2.3 운영 기준, §1.2 G3 / G7, §2.3, §7 Q9 정합). **day-2 운영 (incident / backup / monitoring) 도 standalone — §11 의 운영 runbook 다른 backend 모니터링 도구 공유 ❌**.

### 4.4 raw 운영 정책 (저장 · 암호화 · gitignore · retention · quota, 2026-06-18 신규)

**봉투 암호화** ([ADR-0025 §3](../adr/0025-envelope-encryption-key-management.md) 정합):
- format: `$env$v0.1$<wrapped_dek_b64>$<nonce_b64>$<ciphertext_b64>`
- KEK(32B master key, `DEVHUB_ENCRYPTION_KEY` env) + DEK(AES-GCM-256)
- **raw file on disk = ciphertext**. 평문은 메모리에서만 사용
- **민감 source** (homelab / hrdb / metrics) 는 default 암호화 + .gitignore 권장
- **일반 source** (gitea_public_repo) 는 default 암호화 (raw 자체는 API response 가 잠재적으로 민감)

**.gitignore 정책** (per source, source_meta 의 `gitignore: bool` 필드):
| Source | gitignore 권장 | 근거 |
| --- | --- | --- |
| `gitea_repo_pull` (public repo) | ❌ (commit 가능) | Gitea public repo 의 metadata 는 비공개 정보 아님. 단, internal Gitea instance 의 경우 ✅ |
| `gitea_issue` | ⚠️ (사내 Gitea 한정 ✅, public ✅) | 사내 Gitea 의 issue 는 잠재적 민감 (assignee/label) |
| `gitea_wiki` | ⚠️ (사내 Gitea 한정 ✅, public ✅) | 위키 본문이 민감할 수 있음 |
| `gitea_action` | ⚠️ (사내 Gitea 한정 ✅, public ✅) | CI workflow 가 사내 시스템 정보 포함 |
| `homelab_mock` | ✅ | 사내 시스템 inventory |
| `homelab` (M-v0.2.1+ real wire) | ✅ | 사내 시스템 inventory |
| `metrics` | ⚠️ (Prometheus scrape URL 의 secret 포함 시 ✅, otherwise ❌) | scrape config 의 basic auth credential |
| `hrdb` | ✅ | 사내 인사 데이터 (전형적 PII) |

**Retention 정책** (per source, default 90일):
- raw file 의 `received_at` 으로 N일 경과 시 자동 삭제
- source_meta 의 `retention_days: int = 90` 필드로 override 가능
- **예외**: `metrics` raw = retention 30일 (운영 metric 은 1개월 보관 후 압축 archive)
- **예외**: `hrdb` raw = retention 365일 (인사 데이터 보존 의무, 사내 정책)
- **자동 삭제 시점**: 매일 03:00 UTC cron (M-v0.2.1+ 스케줄러)
- **삭제 전 audit log**: `audit.raw_deleted` event + raw 의 hash + 삭제 시각 + 사유

**Storage quota** (per bundle, default 1GB):
- bundle 별 raw_size 합계가 quota 초과 시 oldest raw 부터 삭제 (LRU policy)
- `bundle_quota_bytes: int = 1GB` (default)
- quota 초과 시 audit log + alert (caller 의 caller 가 notify)
- M-v0.2.0 PoC: quota 미적용 (filesystem 충분). M-v0.2.1+ quota 적용

### 4.5 raw API 권한 + visibility (§3.6 governance 정합, 2026-06-18 신규)

**endpoint 별 권한 matrix** (§3.6.1 endpoint 표 정합 + §3.6.2 curation ownership 정합):

| Endpoint | 인증 | 권한 | 응답 (권한 부족 시) |
| --- | --- | --- | --- |
| `POST /api/v0-2/raw` | user context **필수** (§3.6.1) | caller 가 raw 등록 대상 bundle 의 owner_org_unit_ids 에 속하거나 system_admin | 403 `E_FORBIDDEN` ("raw.register_denied") |
| `GET /api/v0-2/raw/{type}/{name}` | user context **필수** | concept 의 `x_devhub_visibility` 와 동일 (§3.6.4): `org` → caller.org_id ∈ owner_org_unit_ids / `personal` → caller.user_id == owner_user_id / `project` → caller.project_ids ∩ owner_project_ids / `public` → all | 403 `E_FORBIDDEN` (visibility mismatch) or 404 `E_NOT_FOUND` |
| `GET /api/v0-2/raw?source=...&since=...` (list) | user context **필수** | caller 가 조회 가능한 raw 전체 (visibility 정합) + filter by source / since | 200 with caller-visible only (filtered) |
| `DELETE /api/v0-2/raw/{id}` | user context **필수** | system_admin OR raw 등록자 (caller.user_id == raw.registered_by) OR caller 가 raw 의 bundle owner_org_unit_ids 에 속함 | 403 `E_FORBIDDEN` ("raw.delete_denied") |

**Visibility enum 재사용** (§3.6.2 의 4 enum: `org` / `personal` / `project` / `public`):
- raw file 의 frontmatter (또는 sqlite `raw_index` table 의 `visibility` column) 에 저장
- default: source plugin 자동 emit raw = `org` (bundle owner org 기준)
- manual 등록 raw = `org` (caller 등록 시 명시, override 가능)

**caller-provided user context** (§3.6.1) 가 raw API 의 모든 endpoint 에 필수. **missing 시 400 `E_VALIDATION`** ("X-DevHub-User-Context required"). 이 정책은 §3.6 의 caller-provided user context 패턴의 raw 도메인 적용.

### 4.6 raw → concept 정합성 (1:N 관계 · orphan 처리 · update 정책, 2026-06-18 신규)

**1 raw → N concepts 관계** (§3.7.4 normalize() 정합):
- 1 raw file 이 여러 concept emit 가능. 예: homelab 의 `node_inventory.json` 1개 → `dataset_homelab_nodes`, `metric_homelab_node_count`, `reference_homelab_node_specs` 등 3+ concepts.
- sqlite `raw_index` table 의 `concept_ids: str` (comma-separated) 로 추적
- raw 삭제 시 해당 concept 들도 archive 또는 delete (§4.6 orphan 처리)

**raw 삭제 시 concept 처리 정책** (3 mode):

| Mode | 동작 | 사용 시점 |
| --- | --- | --- |
| **hard_delete** | raw + 연관 concept 모두 삭제 | M-v0.2.0 PoC default. file system 직접 삭제 + sqlite `raw_index` + `concept_index` row 삭제 |
| **soft_archive** | raw 삭제 + 연관 concept 는 `x_devhub_status: archived` 로 archive (concept file 유지) | M-v0.2.1+ default. 감사 추적 가능. bundle 별 archive directory 로 이동 or frontmatter status 변경 |
| **retain_concept** | raw 만 삭제, concept 는 유지 (raw_ref dangling) | M-v0.2.3+ 옵션. raw 가 일시적 미러링 (e.g., homelab fixture) 일 때 |

**M-v0.2.0 PoC**: hard_delete (단순화). **M-v0.2.1+**: soft_archive (default), source_meta 의 `on_raw_delete: Enum[hard_delete|soft_archive|retain_concept]` 필드로 설정.

**Orphan concept** (raw 삭제됐는데 concept 만 남은 상태):
- hard_delete mode: 발생 안 함 (raw 와 함께 삭제)
- soft_archive mode: 발생 안 함 (concept 도 archive)
- retain_concept mode: **발생 가능** — `curate/link_resolver.py` 가 `x_devhub_raw_ref: "raw://..."` unresolved link 로 보고 (§3.5.5 정공법)
- 운영 policy: orphan concept 의 `x_devhub_status: orphaned` 자동 설정 + audit + bundle owner_org_unit_ids 내 caller 에 notify

**raw 변경 시 concept update** (raw file 의 source 가 변경된 경우, 예: Gitea 의 issue update):
- source plugin 의 `POST /api/v0-2/ingest/{source}/sync` 가 새 raw emit
- normalize() 가 새 version 의 concept emit (`x_devhub_version` increment, §3.3 정합)
- 이전 version 의 concept 처리:
  - M-v0.2.0 PoC: overwrite (같은 slug 의 .md 파일 덮어쓰기)
  - M-v0.2.1+: `x_devhub_status: superseded` 로 변경 + 새 concept 만 active (viz.html 에서 active 만 표시)
  - M-v0.2.3+: history 보존 (`.md.prev` suffix 로 archived)

### 4.7 raw 정합성 검증 (hash · timestamp · audit, 2026-06-18 신규)

**Hash 검증** (저장 시 + 조회 시):
- 저장 시: sha256 hash 계산 후 sqlite `raw_index.sha256` column 저장
- 조회 시: 매번 hash 재계산 → 불일치 시 `E_INTERNAL` ("raw.integrity_violation") + audit log
- **source-of-truth**: var/raw/ 의 file system vs sqlite metadata → file system 우선 (file 이 source of truth, sqlite 는 index)

**Timestamp 검증**:
- raw 의 `received_at` (caller 측 or source plugin 측) vs `x_devhub_timestamp` (frontmatter) 일치 확인
- 불일치 시 degraded flag (§3.7.5 정합) + audit

**Source timestamp 검증** (외부 시스템 source_timestamp vs raw received_at):
- 외부 시스템 응답의 `updated_at` (또는 equivalent) vs `received_at` 의 차이 = **ingestion lag**
- lag > threshold (default 5분): `ingestion_stale` flag + audit
- M-v0.2.0 PoC: timestamp 기록만 (verification 없음). M-v0.2.1+: threshold 검증 + alert

**Audit log** (raw lifecycle 전체):
| Event | 발생 시점 | Audit log field |
| --- | --- | --- |
| `raw.received` | POST /api/v0-2/raw 성공 or source plugin sync | `{raw_id, source, sha256, size, caller_user_id}` |
| `raw.read` | GET /api-v0-2/raw/{type}/{name} | `{raw_id, caller_user_id, visibility_scope}` |
| `raw.deleted` | DELETE /api-v0-2/raw/{id} | `{raw_id, mode: hard_delete\|soft_archive\|retain_concept, caller_user_id}` |
| `raw.integrity_violation` | hash 재검증 실패 | `{raw_id, expected_sha256, actual_sha256, severity: high}` |
| `raw.ingestion_stale` | timestamp lag > threshold | `{raw_id, source, lag_seconds, threshold_seconds}` |
| `raw.retention_deleted` | 자동 cron retention 만료 | `{raw_id, source, retention_days, age_days}` |
| `raw.quota_evicted` | LRU eviction | `{raw_id, bundle, bundle_size_bytes, quota_bytes}` |

**§11 monitoring 5 지표 정합** (2026-06-18 신규):
- raw 정합성 violation rate (per day) = `audit.raw.integrity_violation` event count / day (§11.3 monitoring #3)
- raw retention_deleted 정상 작동 = `audit.raw.retention_deleted` event count / day (§11.3 monitoring #5 일부)
- §11.1.5 integrity violation runbook → audit log + monitoring alert 자동 trigger (§11.3 routing)
- §11.2 backup + restore drill 시 raw 정합성 검증 자동 수행 (§11.2 Step 4)

**M-v0.2.0 PoC 범위**: hash 검증 (sha256) + audit log (raw.received / raw.read / raw.deleted) 만. timestamp verification + retention cron + quota eviction + integrity_violation alert 는 M-v0.2.1+.

## 5. 마일스톤 + 우선순위 (P0~P3)

### 5.1 마일스톤 표

| ID | 마일스톤 | scope | 의존 | status |
| --- | --- | --- | --- | --- |
| **M-v0.2.0-alpha** | 컨셉 umbrella + child doc 정합 + PoC 진입 | 본 문서 publish + `external-integrations-agentic-rag-roadmap.md` cross-link | (없음) | ⏳ planned (v0.2.0-alpha) |
| **M-v0.2.0** | 1차 standalone 구현 (Gitea 통합 4 sub-plugin PoC, backend 단독, frontend 0) | `backend-knowledge/` skeleton + OKF spec model (frontmatter `x_devhub_category` 필드 추가) + **Gitea 통합 4종** (`gitea_repo_pull` / `gitea_issue` / `gitea_wiki` / `gitea_action`, Gitea 1 instance 의 4 sub-plugin, 5 카테고리 중 4: 이슈/위키/SCM/CI-CD, §3.8.2 정공법) + `homelab_mock` 1종 (§3.8.3 정공법) = **5종 PoC (5 카테고리 결정 기반, 2026-06-17, §3.2.1 / §3.7.2 / §3.8 / §6.4 정합)** + Ingest 1 endpoint + Query 1 endpoint (concept 직접 조회) + 1차 raw API + OpenAPI. **frontend 0 page** (M-v0.2.0 만, viz.html 자가 viewer 만 SSR) | M-v0.2.0-alpha | ⏳ planned (v0.2.0) |
| **M-v0.2.1** | 1차 완성 + Gitea 통합 정식 + 사내 시스템 wire + Curate + frontend 관리/조회 page 1 | Gitea 통합 4종 정식 wire (1차 PoC → 정식, 5 카테고리 중 4) + `homelab_mock` → `homelab` (real wire, 5 카테고리 외 사내 시스템) = 5종 운영 + Curate 3 endpoint (enrich / edit / rebuild) + 1차 viz.html (자가 viewer) + **frontend 관리/조회 page 1** (`backend-knowledge/web/`, 별도 standalone frontend, **devhub frontend 와 분리**, standalone 정책 정합, **§12.2 의 5 page 상세 정합**: concept list / concept detail / ingest trigger / bundle management / raw inspector) + e2e smoke | M-v0.2.0 | ⏳ planned (v0.2.1) |
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

### 5.4 Milestone dependency graph + critical path (2026-06-18 신규)

**6 마일스톤 의존 관계** (linear, 단 §5.7 parallel sprint 가능):

```
M-v0.2.0-alpha
    │
    ▼ (의존)
M-v0.2.0
    │
    ▼ (의존)
M-v0.2.1
    │
    ▼ (의존)
M-v0.2.2
    │
    ▼ (의존)
M-v0.2.3
    │
    ▼ (의존)
M-v0.3.0
```

| 마일스톤 | 의존 (직전) | 의존 (병렬 가능) | 산출물 | 상태 |
| --- | --- | --- | --- | --- |
| **M-v0.2.0-alpha** | (없음) | ADR-0034/0035 publish + state.json M-v0.2.0 row | umbrella doc + child doc active 전환 | planned |
| **M-v0.2.0** | alpha | OKF SPEC.md 1차 정독 (M-v0.2.0 이전) | skeleton + OKF spec model + Gitea 4 + homelab_mock + Ingest/Query + raw API + OpenAPI | planned |
| **M-v0.2.1** | v0.2.0 | frontend page 1 design + Pi LLM vendor 결정 (M-v0.2.3 이전) | Gitea 정식 + homelab real + Curate + viz.html + frontend 관리 page 1 + e2e smoke | planned |
| **M-v0.2.2** | v0.2.1 | backend-ai/ 디렉터리 제거 PR | metrics + 6종 운영 + backend-ai 폐기 | planned |
| **M-v0.2.3** | v0.2.2 | hrdb schema spec + Pi SDK/RPC mode 결정 | + hrdb + Pi LLM enrich + cross-link 자동 resolution + 7종 운영 | planned |
| **M-v0.3.0** | v0.2.3 | embedding model 결정 (sentence-transformers or 외부) | chunking + embedding + vector index + reranking + multi-vendor LLM | planned |

**Critical path** (linear, parallel 불가능):
```
M-v0.2.0 (skeleton + 5종 source)
  → M-v0.2.1 (Gitea 정식 + frontend page 1)
    → M-v0.2.2 (metrics + backend-ai 폐기)
      → M-v0.2.3 (hrdb + Pi LLM enrich)
        → M-v0.3.0 (RAG)
```

**병렬 가능 sprint** (§5.7 parallel sprint + PR 전략):
- `backend-knowledge/` skeleton PR + umbrella doc PR = 별도 PR (동시 가능)
- ADR-0034 + ADR-0035 publish = 별도 PR (동시 가능, ADR-0035 가 ADR-0034 영향 받지만 publish 는 parallel)
- 5종 source plugin (Gitea 4 + homelab_mock) = 1 batch PR (M-v0.2.0) 또는 5개 PR (M-v0.2.1 의 정식 wire 시)
- frontend page 1 = 별도 PR (M-v0.2.1)
- backend-ai/ 폐기 = 별도 PR (M-v0.2.2)
- Pi LLM enrich = 별도 PR (M-v0.2.3)

**Risk analysis** (병렬 시 의존성 충돌):
- (a) skeleton PR + ADR PR 동시 진행: ADR 이 skeleton 의 okf/ 디렉터리 영향 받으므로, skeleton 의 okf/spec.py commit 후 ADR 영향 section 갱신 필요 (sequential within PR chain, but PR 자체는 parallel)
- (b) frontend page 1 + 5종 source plugin 동시 진행: frontend 가 source plugin 의 normalize() 결과 (viz.html) 사용 → frontend 는 source plugin PR merge 이후 진행 권장
- (c) Pi LLM enrich + hrdb source 동시 진행: hrdb source 의 normalize() 결과가 Pi LLM 의 enrich input → hrdb 먼저 commit 후 Pi 진행

### 5.5 마일스톤별 DoD (Definition of Done, 2026-06-18 신규)

**M-v0.2.0-alpha DoD**:
- (a) umbrella doc (`release_v0-2_roadmap.md`) status = `accepted` + cross-link 정합
- (b) child doc (`external-integrations-agentic-rag-roadmap.md`) status = `active` + §8 변경 이력 row 추가
- (c) `ai-workflow/memory/state.json` M-v0.2.0 row 발급
- (d) ADR-0034 + ADR-0035 publish
- (e) GitHub milestone `v0.2.0` 생성 + umbrella doc link 첨부

**M-v0.2.0 DoD** (PoC, 1차 standalone):
- (a) **코드/문서**:
  - `backend-knowledge/` 디렉터리 skeleton: Dockerfile + pyproject.toml + main.py + `okf/spec.py` + `okf/frontmatter.py` + `okf/link_graph.py` + `curate/enricher.py` + `curate/index_builder.py` + `curate/link_resolver.py` + `sources/_base.py`
  - 5종 PoC source plugin: `gitea_repo_pull.py` + `gitea_issue.py` + `gitea_wiki.py` + `gitea_action.py` + `homelab_mock.py`
  - 4 API module: `api/ingest.py` + `api/curate.py` + `api/query.py` + `api/raw.py` + `api/bundles.py`
  - dev script: `backend-knowledge/dev-up.sh` + `backend-knowledge/docker-compose.yml` (standalone)
- (b) **검증** (per §3.8.5):
  - 단위 테스트: pytest, 5종 source plugin + curate modules, ≥80% coverage
  - e2e smoke: POST /ingest/{source}/sync → real Gitea instance 또는 homelab fixture → GET /concepts/{type}/{name} → 정상 응답
- (c) **ADR 영향**:
  - ADR-0034 §4.3 영향 section §3.5~§3.9 row 모두 정합
  - ADR-0035 §3.8 마일스톤 표 와 §5 정합
- (d) **운영**:
  - docker-compose.yml 단독 기동 (backend-core 와 별도 docker network)
  - viz.html 자가 viewer SSR (frontend 0 page, viz.html 만)
- (e) **cross-section 정합**:
  - umbrella doc §9 변경 이력 row 추가
  - 본 §5.3 checklist 6/6 통과

**M-v0.2.1 DoD** (1차 완성 + Gitea 정식 + frontend):
- (a) **코드/문서**:
  - `gitea_repo_pull.py` / `gitea_issue.py` / `gitea_wiki.py` / `gitea_action.py` 정식 wire (PoC → real)
  - `homelab.py` (real wire, `homelab_mock.py` 대체)
  - 3 Curate endpoint: `POST /concepts/{id}/enrich` + `PUT /concepts/{id}` + `POST /bundles/{bundle}/rebuild`
  - frontend 관리/조회 page 1: `backend-knowledge/web/` (별도 standalone frontend, devhub frontend 와 분리)
- (b) **검증**:
  - 5종 source plugin 정식 wire e2e smoke
  - frontend 페이지 관리/조회 happy path e2e
  - 검토 workflow (human 작성) e2e (M-v0.2.1 PoC 의 frontend 1 page)
- (c) **ADR 영향**:
  - ADR-0035 §4.1 positive 갱신 (frontend page 1 추가)
- (d) **운영**:
  - standalone 유지 (gateway/firewall 보호)
  - viz.html 정식 viewer + per-bundle/per-category/index.md 자동 생성
- (e) **cross-section 정합**:
  - §3.9 lifecycle reviewed/published state 정합 (human 작성 review workflow)
  - §6.1 Phase 1 운영 정합

**M-v0.2.2 DoD** (5 카테고리 외 추가 wire + backend-ai 폐기):
- (a) **코드/문서**:
  - `metrics.py` (Prometheus scrape API)
  - 6종 운영 (Gitea 4 + homelab + metrics)
  - `backend-ai/` 디렉터리 제거 (placeholder 정리)
  - root `docker-compose.deploy.yml` 의 `backend-ai` service 제거
  - root `Makefile` / `dev-up.sh` 의 backend-ai target 제거
- (b) **검증**:
  - 6종 source 운영 e2e smoke
  - backend-ai reference 0 (grep / find 결과)
- (c) **ADR 영향**:
  - ADR-0035 §6 supersession 갱신 (backend-ai 폐기 결정 명시)
- (d) **운영**:
  - root level 정리 + backend-knowledge 독립 유지
- (e) **cross-section 정합**:
  - §6.2 Phase 2 운영 정합
  - §3.7.2 per-source mapping 6 row 정합

**M-v0.2.3 DoD** (hrdb + Pi LLM enrich + cross-link 자동):
- (a) **코드/문서**:
  - `hrdb.py` (사내 HR DB PostgreSQL)
  - 7종 운영 (Gitea 4 + homelab + metrics + hrdb)
  - Pi LLM enrich (`pi_bridge/rpc_client.py` 또는 `sdk_client.py`)
  - cross-link 자동 resolution (`curate/link_resolver.py` Pi LLM 연동)
- (b) **검증**:
  - 7종 source 운영 e2e smoke
  - Pi LLM enrich 단위 테스트 (mock Pi response)
  - cross-link 자동 resolution e2e smoke (orphan link 해소)
- (c) **ADR 영향**:
  - ADR-0034 §3 (Pi 의 RPC mode / SDK mode 결정) 갱신
- (d) **운영**:
  - 1 vendor Pi 정합 (vendor-agnostic, Q3 결정)
  - long-running connection (RPC mode) or subprocess (SDK mode) 운영
- (e) **cross-section 정합**:
  - §6.3 Phase 3 운영 정합
  - §3.7.5 edge cases (Pi LLM 다운시 rule-based fallback) 정합

**M-v0.3.0 DoD** (풀 RAG):
- (a) **코드/문서**:
  - chunking module (`backend-knowledge/curate/chunker.py`)
  - embedding module (`backend-knowledge/curate/embedder.py`, sentence-transformers or 외부)
  - vector index (sqlite-vss or pgvector)
  - reranker module
  - multi-vendor LLM abstraction (Pi `pi-ai` 의 15+ provider)
- (b) **검증**:
  - RAG happy path e2e smoke (query → retrieval → reranking → answer)
  - benchmark (precision/recall on sample corpus)
- (c) **ADR 영향**:
  - ADR-0034 §3 OKF + vector index reference 갱신
  - ADR-0035 §3.3 (Query API 의 LLM answer 합성) 갱신
- (d) **운영**:
  - vector index 백업 + restore procedure
  - multi-vendor fallback (1 vendor 다운시 다른 vendor 사용)
- (e) **cross-section 정합**:
  - §3.7.5 edge cases (vector index 손상 시 fallback) 정합

### 5.6 Cutover 절차 + rollback plan (per milestone, 2026-06-18 신규)

**Cutover 절차** (per milestone 의 deployment 전환):

```
Step 1: 이전 milestone 의 docker-compose 가동 상태 확인 (health check)
Step 2: 새 milestone 의 docker 이미지 빌드 + push (CI/CD)
Step 3: 새 milestone 의 docker-compose 가동 (별도 docker network, parallel 가동)
Step 4: smoke test (새 milestone 의 핵심 endpoint 1~2개)
Step 5: 이전 milestone 의 docker-compose 종료
Step 6: 새 milestone 단독 가동 + monitoring dashboard 확인
```

**Cutover trigger**:
- M-v0.2.0 → M-v0.2.1: frontend page 1 deploy + viz.html 정식 viewer 활성화
- M-v0.2.1 → M-v0.2.2: metrics source 추가 + backend-ai 제거
- M-v0.2.2 → M-v0.2.3: hrdb source 추가 + Pi LLM enrich 활성화
- M-v0.2.3 → M-v0.3.0: chunking + embedding + vector index 추가

**Rollback plan** (cutover 실패 시):

| Trigger | Rollback | RTO (Recovery Time Objective) |
| --- | --- | --- |
| 새 milestone 의 health check 실패 (5분 내) | `docker-compose down` + 이전 milestone 의 `docker-compose up` | < 10분 |
| smoke test 실패 (1 endpoint 이상) | 새 milestone 의 docker-compose 종료 + 이전 milestone 복원 | < 30분 |
| 운영 중 data corruption (raw integrity violation, §4.7) | 이전 milestone 로 rollback + raw 백업본으로 복원 | < 1시간 |
| 보안 incident (credential leak 등) | 모든 milestone 즉시 정지 + credential 회전 + ADR emergency §5 amendment | < 24시간 |

**Cutover 후 monitoring** (per milestone):
- (a) source plugin sync 성공률 (per source): ≥99%
- (b) Query API p95 latency: ≤500ms (M-v0.2.0 PoC), ≤200ms (M-v0.2.1+)
- (c) raw 정합성 violation: ≤0.01% (per day)
- (d) Concept archive 자동 trigger 정상 작동 (raw 변경 → concept superseded, §4.6 정합)
- (e) Audit log 정상 기록 (7 event type, §4.7 정합)

**Cutover checklist** (per milestone, 8 항목):
- [ ] 이전 milestone 의 docker image backup 완료
- [ ] 새 milestone 의 docker image CI build 성공
- [ ] 새 milestone 의 unit test 100% (or 명시적 skip 사유)
- [ ] 새 milestone 의 e2e smoke 통과
- [ ] ADR 영향 section 갱신 (해당 milestone 의 DoD 정합)
- [ ] umbrella doc §9 변경 이력 row 추가
- [ ] 운영자 notification (cutover 시작/완료)
- [ ] Rollback plan 검증 (dry-run)

**§11 운영 runbook 과의 정합** (2026-06-18 신규):
- cutover rollback trigger 발동 시 §11.1 의 incident runbook 중 해당 type 즉시 활성화 (예: cutover 후 source plugin sync 실패 → §11.1.1)
- on-call operator (§11.4) 가 cutover rollback + incident 대응 동시 수행
- cutover 후 monitoring 지표 5 항목 중 1 이상 임계 초과 시 incident 등록 (§11.3 routing)
- §11.2 backup = cutover 직전 backup 필수 (8 번째 cutover checklist 와 정합)

### 5.7 Parallel sprint + PR 전략 (2026-06-18 신규)

**PR 단위 분리 전략** (per 마일스톤):

| 마일스톤 | 권장 PR 수 | PR 목록 (예시) |
| --- | --- | --- |
| M-v0.2.0-alpha | 3 PR | (1) umbrella doc publish + ADR-0034 + ADR-0035 / (2) state.json M-v0.2.0 row + GitHub milestone 생성 / (3) child doc status active 전환 |
| M-v0.2.0 | 4 PR | (1) backend-knowledge/ skeleton (Dockerfile + pyproject + main + okf/) / (2) 5종 PoC source plugin (sources/) / (3) 4 API module + OpenAPI / (4) dev-up.sh + docker-compose |
| M-v0.2.1 | 3 PR | (1) Gitea 4 정식 wire + homelab real / (2) 3 Curate endpoint + lifecycle state / (3) frontend 관리 page 1 (`backend-knowledge/web/`) |
| M-v0.2.2 | 2 PR | (1) metrics source + 6종 운영 / (2) backend-ai/ 폐기 + root level 정리 |
| M-v0.2.3 | 3 PR | (1) hrdb source / (2) Pi LLM enrich (RPC mode 또는 SDK mode) / (3) cross-link 자동 resolution |
| M-v0.3.0 | 4 PR | (1) chunking / (2) embedding + vector index / (3) reranker / (4) multi-vendor LLM abstraction |

**PR 의존성 정합** (sequential merge 필요):
- dependency 가 있는 PR 은 순차 merge (rebase on main + merge)
- dependency 가 없는 PR 은 parallel 가능 (각각 별도 branch + squash merge)

**Branch prefix 전략** (per AGENTS.md 2026-06-09 결정의 free prefix 허용):
- `chore/<scope>` — doc / ADR / config 변경 (umbrella doc, ADR)
- `feat/<scope>` — 신규 기능 (source plugin, endpoint)
- `fix/<scope>` — bug fix
- 예: `chore/v0-2-umbrella`, `feat/backend-knowledge-skeleton`, `feat/gitea-source-plugin`, `fix/raw-integrity-bug`

**PR template** (per AGENTS.md 추적성):
- PR body: "추적성 영향" 섹션 (REQ/ARCH/API/RM/IMPL/UT/TC ID)
- label: tier (사외/사내/공용) + milestone (M-v0.2.0 등) + scope (v0.2-umbrella / backend-knowledge 등)
- reviewer: AGENTS.md §0.4 Owner 권한 기준

**PR 머지 후 처리** (per milestone release):
- GitHub milestone close
- umbrella doc §9 변경 이력 row 추가 (release commit)
- ADR 영향 section 갱신 (해당 milestone 정합)
- state.json status update (`planned` → `in_progress` → `done`)

## 6. 1차 독립 개발 → 연동 단계 (Phase 1 / 2 / 3)

### 6.1 Phase 1 — 1차 standalone (M-v0.2.0 + M-v0.2.1)

- **독립 기동**: **`backend-knowledge/dev-up.sh` (별도 스크립트, backend-knowledge 만 단독 기동)** + **`backend-knowledge/docker-compose.yml` (별도, backend-knowledge 서비스만)**. **devhub 의 root `dev-up.sh` / `docker-compose.colima.yml` / `docker-compose.deploy.yml` 사용 ❌** (다른 backend 연결 방지, §1.2 G7 / §2.3 / §4.2 standalone 정책 정합). 별도 docker network.
- **mock source**: filesystem fixture 1개 (`backend-knowledge/var/fixtures/homelab/*.json`, `homelab_mock.py` 의 입력, 5 카테고리 외 사내 시스템). 외부 시스템 실제 호출 ❌. **Gitea 통합 4 sub-plugin 은 M-v0.2.0 부터 real Gitea 1 instance 에 PoC wire** (mock 없이 실제 API, 1차 wire 정공법 — §3.2.1 / §6.4 정합).
- **인증**: **internal-only, no auth** (gateway / firewall / IP allowlist 별도 보호, §2.3). OIDC / Keycloak / backend-core 인증 위임 ❌.
- **테스트**: unit (OKF spec / enricher / link_graph) + e2e (ingest → curate → query) 의 **신규 백엔드 단독**.
- **frontend**: **M-v0.2.0 만 frontend 0 page** (1차 backend 단독 구현). **M-v0.2.1 부터 frontend 관리/조회 page 1 추가** (§5.1 M-v0.2.1 / §5.2 P1 / **§12 정합**, `backend-knowledge/web/` 별도 standalone frontend, devhub frontend 와 분리, §1.2 G7 standalone 정책 정합). viz.html 자체 viewer (자가 graph viewer) 는 backend-knowledge 가 SSR (모든 Phase 공통, **§12.1 상세 정합**: Cytoscape.js + marked.js CDN embed + 4 edge type + 8 type node 색상)
- **day-2 운영**: §11 운영 runbook 정공법 적용 (incident 대응 + backup + monitoring, M-v0.2.0 PoC = 1 operator, §11.4)

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

### 6.3 Phase 3 — LLM enrich (Pi) + Pi periodic ingest pipeline (2026-06-18 신규)

- **M-v0.2.0 PoC** (Pi 의 periodic ingest pipeline, §10.3 정합): Pi `pi-coding-agent` **SDK mode** (Node subprocess per call) 가 db mode source 의 raw 데이터를 주기적 ingest. 1차 source = `homelab_mock` (§10.4 default). source plugin 의 rule-based normalize 와 **병렬 운영** (file mode source 는 §3.7 정공법, db mode source 는 §10).
- **M-v0.2.1+**: scheduler 고도화 (APScheduler), cron interval per source config
- **M-v0.2.3+**: Pi **RPC mode** option 추가 (long-running JSON over stdin/stdout, 동일 process 내 multi-ingest 처리), cross-link 자동 resolution 활성화 (link_graph 가 unresolved link 발견 시 Pi LLM 호출). **추가로** LLM enrich (`POST /concepts/{id}/enrich`) 활성화 — §3.1 Curate 의 M-v0.2.3+ timing 정합.
- **1차 vendor**: 1 vendor 선택 (운영자 결정 시점). `pi-ai` 의 15+ provider 중 (Anthropic / OpenAI / Google / Azure / Bedrock / Mistral / Groq / Cerebras / xAI / Hugging Face / Kimi For Coding / MiniMax / OpenRouter / Ollama). 장기 multi-vendor (M-v0.3.0+)
- **§10 cross-section**: Pi 의 역할이 (a) db mode source 의 periodic ingest (M-v0.2.0+) + (b) LLM enrich (M-v0.2.3+) + (c) cross-link 자동 resolution (M-v0.2.3+) 의 3 가지로 확장. (a) 가 가장 먼저 activate.
- **풀 RAG** (M-v0.3.0): chunking + embedding + vector index + retrieval + reranking

### 6.5 Phase 1 운영 정공법 상세 (M-v0.2.0 + M-v0.2.1, 2026-06-18 신규)

**§6.1 의 high-level 정공법 보강** — Phase 1 의 docker-compose standalone 정합 + gateway/firewall 정책 + e2e smoke pipeline 상세.

#### 6.5.1 docker-compose standalone 정합 (§2.1 + §6.1 보강)

**`backend-knowledge/docker-compose.yml` (별도, root 와 무관)**:
- services:
  - `backend-knowledge`: FastAPI app (port 8000), image: `backend-knowledge:v0.2.0` (PoC)
  - `db` (M-v0.2.0 PoC = sqlite, M-v0.2.3+ = postgres): §10.1 정합
- networks:
  - `backend-knowledge-net` (별도, backend-core-net 와 격리, §1.2 G7 + §2.3 standalone 정합)
- volumes:
  - `./var/bundles:/app/var/bundles` (git 가능)
  - `./var/raw:/app/var/raw` (file mode 일 때, .gitignore 정합 §4.4)
  - `./var/raw_index.db:/app/var/raw_index.db` (db mode 일 때, §10.1)
  - `./var/fixtures:/app/var/fixtures` (homelab_mock PoC)

**`backend-knowledge/dev-up.sh` (별도 스크립트)**:
```bash
#!/bin/bash
# dev-up.sh: backend-knowledge 만 단독 기동 (root level 무관)
set -euo pipefail
docker compose -f backend-knowledge/docker-compose.yml up -d
echo "backend-knowledge started at http://localhost:8000"
echo "viz.html: http://localhost:8000/viz.html"
echo "API docs: http://localhost:8000/docs"
```

**Standalone 정합 검증** (per release):
- (a) `docker compose up -d` 시 backend-core 와의 네트워크 격리 확인 (`docker network ls`)
- (b) root level 의 `docker-compose.colima.yml` / `docker-compose.deploy.yml` / `dev-up.sh` 사용 ❌ (no docker compose from root)
- (c) port 충돌 없음 (backend-core 가 8001+ 사용 가정)

#### 6.5.2 Mock source + real wire transition (M-v0.2.0 → M-v0.2.1)

**M-v0.2.0 PoC = 5종 source + 1 mock**:
- `gitea_repo_pull.py` / `gitea_issue.py` / `gitea_wiki.py` / `gitea_action.py` = **real wire** (Gitea 1 instance, PoC 단계)
- `homelab_mock.py` = **filesystem fixture** (`var/fixtures/homelab/*.json`)

**M-v0.2.1** = `homelab_mock.py` → `homelab.py` (real wire) 교체:
- homelab.py 의 SourceMeta.storage_mode = `db` (Pi-driven normalize, §10.4 default)
- homelab.py 의 SourceMeta.normalize_mode = `pi-sdk` (M-v0.2.0 PoC SDK mode, §10.3)
- mock.py 는 deprecated (M-v0.2.1+ source plugin registry 에서 제거, but backward compat 유지)
- `homelab` 의 source plugin ABC 의 `connect()` method 가 HTTP API 호출 (homelab agent 의 /api/v1/nodes, /api/v1/services)

**§10.4 storage_mode 전환 정책**:
- M-v0.2.0 PoC: `homelab_mock` = file (fixture 기반), `homelab` 도입 시 db 모드 (§10.4 default mapping)
- M-v0.2.1+: `homelab` = db (Pi-driven), `homelab_mock` = 운영 종료 (단, fallback 용도 유지, M-v0.3.0+ deprecated)

#### 6.5.3 Gateway + firewall + IP allowlist 정책

**gateway 책임** (caller, §3.6.1 Path Y 정합):
- (a) Keycloak 인증 + user context 추출 (DevHub backend-core 의 Keycloak federation 정합)
- (b) `X-DevHub-User-Context` header 생성 (base64url(json) format)
- (c) backend-knowledge 호출 (HTTP)
- (d) backend-knowledge 응답의 envelope filter 결과를 user 에게 반환

**firewall / IP allowlist 정책** (per environment):

| Environment | 정책 | IP allowlist | TLS |
| --- | --- | --- | --- |
| **dev** (local) | localhost 만 | 127.0.0.1, ::1 | ❌ (plaintext) |
| **staging** (사내) | VPN or bastion host | 10.0.0.0/8 (사내) | ✅ (self-signed or 사내 CA) |
| **production** (외부) | WAF + IP allowlist | 화이트리스트 (운영자 결정) | ✅ (Let's Encrypt or 사외 CA) |

**§2.3 의 "API 인증: internal-only, no auth + gateway/firewall/IP allowlist"** 정합. backend-knowledge 자체는 firewall 보호 + gateway caller 책임 인증 (Path Y, §3.6.1).

#### 6.5.4 E2E smoke pipeline (M-v0.2.0 PoC, 5종)

**5 step e2e smoke** (per source):
```
Step 1: POST /api/v0-2/ingest/gitea_repo_pull/sync (real Gitea instance)
Step 2: raw_records / var/raw/ 확인 (storage_mode 별)
Step 3: bundle/index.md 자동 생성 확인
Step 4: GET /api/v0-2/concepts/integration_gitea_repo_puller 정상 응답
Step 5: viz.html SSR 확인 (cross-link node 표시)
```

**5종 PoC source plugin** e2e smoke 1 cycle = ~5분 (real Gitea + fixture). CI e2e lane = PR merge 시 자동 실행 (M-v0.2.1+).

### 6.6 Phase 2 운영 정공법 상세 (M-v0.2.2, 2026-06-18 신규)

**§6.2 의 high-level 정공법 보강** — Phase 2 의 6종 source wire cutover + backend-ai 폐기 절차 + e2e 6종 smoke 상세.

#### 6.6.1 6종 source wire cutover (5 → 6종, metrics 추가)

**metrics.py 추가** (Prometheus scrape API):
- SourceMeta:
  - bundle = `devhub-metrics`
  - category = `(5 카테고리 외)`
  - storage_mode = `file` (§10.4 default, operational metric 단순)
  - normalize_mode = `rule-based`
- Endpoint: `/api/v1/query?query=...` (Prometheus HTTP API)
- Emit types: `metric`, `reference`, `api_endpoint` (§3.7.2 정합)
- Credential: bearer token (optional, M-v0.2.2+ Prometheus scrape config 정합)

**Cutover 절차** (M-v0.2.1 → M-v0.2.2):
```
Step 1: metrics.py 작성 + 단위 테스트 (§3.8.4 10 step 절차)
Step 2: e2e smoke (5종 → 5종 + metrics = 6종)
Step 3: §5.6 cutover checklist 8 항목 통과
Step 4: metrics.py merge + docker-compose 재기동
Step 5: 운영 검증 (1주 monitoring 지표 정상)
```

#### 6.6.2 backend-ai 폐기 절차 (단계별, M-v0.2.2 정공법)

**폐기 단계** (per file/directory):

| 단계 | 대상 | 작업 | 영향 |
| --- | --- | --- | --- |
| 1 | `backend-ai/main.py` | ADR-0035 §6 supersession 에 따라 제거 | (no backend reference) |
| 2 | `backend-ai/Dockerfile` | 제거 | docker image 빌드 제외 |
| 3 | `backend-ai/.venv` / `.build` | git rm | repo size 감소 |
| 4 | `docker-compose.deploy.yml` 의 backend-ai service | 제거 (root level) | root `dev-up.sh` 영향 |
| 5 | root `Makefile` 의 backend-ai target | 제거 | root Makefile 영향 |
| 6 | root `dev-up.sh` 의 backend-ai target | 제거 | root dev script 영향 |
| 7 | `docs/` 의 backend-ai reference (있는 경우) | 제거 | docs 영향 |
| 8 | `infra/` 의 backend-ai config (있는 경우) | 제거 | infra 영향 |
| 9 | ci workflow (`ci-internal.yml`) 의 backend-ai job | 제거 | CI 영향 |
| 10 | `ai-workflow/memory/state.json` 의 backend-ai row | 제거 (archive) | state 정합 |

**PR 분리** (per AGENTS.md / §5.7):
- PR 1: backend-ai/ 디렉터리 + Dockerfile + main.py 제거
- PR 2: root level 정리 (docker-compose, Makefile, dev-up.sh)
- PR 3: docs + infra + ci 정리
- PR 4: state.json archive

**각 PR 의 DoD**:
- (a) grep `backend-ai` 결과 = 0 (단, archive reference 제외)
- (b) docker compose 가동 정상 (backend-knowledge 만)
- (c) ADR-0035 §6 supersession row 추가

#### 6.6.3 E2E 6종 smoke + alert routing 검증

**6 step e2e smoke** (per source, M-v0.2.2):
```
Step 1: POST /api/v0-2/ingest/{gitea_repo_pull|gitea_issue|gitea_wiki|gitea_action|homelab|metrics}/sync
Step 2: raw_records / var/raw/ 확인 (storage_mode 별, §10.4)
Step 3: bundle/index.md 자동 생성 (per-bundle/per-category 6 bundle)
Step 4: GET /api/v0-2/concepts/{type}/{name} 정상 응답 (per source 1+ concept)
Step 5: viz.html SSR 확인 (6 source 의 concept 표시)
Step 6: alert routing 검증 (Slack #backend-knowledge-critical 채널, 1 test alert 발송)
```

**6종 smoke 1 cycle = ~10분**. CI e2e lane = PR merge 시 자동 실행.

### 6.7 Phase 3 운영 정공법 상세 (M-v0.2.3, 2026-06-18 신규)

**§6.3 의 high-level 정공법 보강** — Phase 3 의 hrdb 추가 + Pi 운영 상세 + LLM enrich 운영 상세.

#### 6.7.1 7종 source wire cutover (6 → 7종, hrdb 추가)

**hrdb.py 추가** (사내 HR DB PostgreSQL):
- SourceMeta:
  - bundle = `devhub-hrdb`
  - category = `(5 카테고리 외, x_devhub_category 미설정)`
  - storage_mode = `db` (§10.4 default, PII + 복잡)
  - normalize_mode = `pi-sdk` (Pi-driven normalize, §10.3)
- Endpoint: PostgreSQL connection string + SELECT query (사내 schema)
- Emit types: `dataset`, `reference`, `metric` (§3.7.2 정합, PII → Pi 가 semantic 처리)
- Credential: PostgreSQL connection (봉투 암호화, ADR-0025 정합)

**Cutover 절차** (M-v0.2.2 → M-v0.2.3):
```
Step 1: hrdb.py 작성 + §3.8.4 10 step 절차 (storage_mode=db 정합)
Step 2: §10.3 Pi ingest pipeline 검증 (homelab_mock → hrdb 전환)
Step 3: §5.6 cutover checklist 8 항목 통과
Step 4: hrdb.py + §10.3 변경사항 merge + docker-compose 재기동
Step 5: 운영 검증 (1주 monitoring 지표 정상 + Pi ingest 정상)
```

#### 6.7.2 Pi 운영 상세 (SDK mode M-v0.2.0~v0.2.2 + RPC mode M-v0.2.3+)

**SDK mode** (M-v0.2.0~v0.2.2, default):
- `@earendil-works/pi-coding-agent` npm pkg (Node.js subprocess)
- Python `subprocess.Popen` 으로 `node` 실행
- stdin/stdout JSON message passing
- 매 ingest job 마다 subprocess 시작/종료 (latency 1~3초 overhead)
- 장점: 단순, stateless, 실패 isolation
- 단점: latency, resource (Node 매번 시작)

**RPC mode** (M-v0.2.3+, option):
- long-running JSON-RPC connection (Node server 시작 후 stdin/stdout 으로 message passing)
- 한 번 Node 시작 후 여러 ingest job 처리
- 장점: latency 감소 (~100ms), resource 효율
- 단점: stateful (실패 시 connection 관리 필요)

**Mode 선택 기준**:
- M-v0.2.0 PoC: SDK mode (단순)
- M-v0.2.1+: SDK mode 유지 (1 source 만 Pi-driven = homelab_mock)
- M-v0.2.3+: RPC mode option 추가 (hrdb 추가 + 2 source Pi-driven)
- §11.1.3 Pi ingest pipeline timeout/degraded runbook 정합

#### 6.7.3 LLM enrich + cross-link 자동 resolution 운영 (M-v0.2.3+)

**POST /api/v0-2/concepts/{id}/enrich** 운영 (§3.1 + §6.3 정합):
- Trigger: 운영자 또는 frontend 관리자가 concept ID 지정
- Pi SDK mode (default) 또는 RPC mode 호출
- Input: 기존 concept 의 frontmatter + body
- Output: Pi 가 LLM enrich 한 새 concept (audit log)
- 새 version 의 concept emit (`x_devhub_version` increment, §3.9 정합)

**cross-link 자동 resolution** 운영 (§3.5.5 + §6.3 정합):
- `curate/link_resolver.py` 가 unresolved link 발견 시 Pi LLM 호출
- Pi 가 "가장 유사한 concept 추천" (vector similarity + LLM semantic)
- 추천 결과 운영자 승인 시 cross-link 자동 추가 (M-v0.2.3+ frontend)

**§11 운영 runbook 정합**:
- §11.1.3 Pi ingest pipeline timeout/degraded runbook = 본 §6.7.2 정합
- §11.1.6 archive trigger 실패 = 본 §6.7.3 LLM enrich 의 superseded trigger 정합
- §11.3 monitoring 5 지표 = 본 §6.7 운영 지표 (sync / Query p95 / integrity / Pi ingest success / archive trigger) 정합

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

## 7. Risks + Q&A 결정 (2026-06-18 18/18 결정 완료)

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
| Q12 | **raw storage_mode 결정** (2026-06-18 결정, §10 신규) | file only / db only / **dual mode per source** (file + db, source_meta.storage_mode field) | **dual mode per source** (2026-06-18 결정, §10.4). gitea 4 = file + homelab = db + metrics = file + hrdb = db (default mapping). 운영자 override 가능. **단순 external API (gitea, metrics) 는 file + rule-based**, **complex 사내 시스템 (homelab, hrdb) 는 db + pi-sdk**. §4 + §10 정합 |
| Q13 | **Pi SDK mode timing** (2026-06-18 변경, §6.3 + §10 신규) | M-v0.2.3+ enrich only (original) / **M-v0.2.0+ DB path periodic ingest + M-v0.2.3+ LLM enrich + M-v0.2.3+ cross-link 자동 resolution (변경)** | **변경** (2026-06-18 결정). Pi 의 역할 3 가지로 확장: (a) M-v0.2.0+ db mode source 의 periodic ingest (가장 먼저 activate) + (b) M-v0.2.3+ LLM enrich (`POST /concepts/{id}/enrich`) + (c) M-v0.2.3+ cross-link 자동 resolution (`curate/link_resolver.py` unresolved link → Pi LLM 추천). §6.3 / §10.3 정합 |
| Q14 | **DB type 결정 (sqlite vs PostgreSQL)** (2026-06-18 결정, §10 신규) | M-v0.2.0 부터 PostgreSQL / **M-v0.2.0~v0.2.2 sqlite + M-v0.2.3+ PostgreSQL option** / always sqlite | **M-v0.2.0~v0.2.2 sqlite + M-v0.2.3+ PostgreSQL option** (2026-06-18 결정, §10.1). M-v0.2.0 PoC = sqlite (단순, file-based, git 가능 시 .gitignore). M-v0.2.3+ = PostgreSQL option (production-grade, 동시성 + backup/restore 정합). **backend-core 의 PostgreSQL 와 공유 ❌** (standalone 정합, §1.2 G7) |
| Q15 | **per source storage_mode default mapping** (2026-06-18 결정, §10 신규) | gitea 4 = file + homelab = db + metrics = file + hrdb = db (decided) / 모두 file / 모두 db | **decided mapping** (2026-06-18 결정, §10.4). (1) `gitea_repo_pull` / `gitea_issue` / `gitea_wiki` / `gitea_action` = file (rule-based normalize 적합, simple REST schema) / (2) `homelab_mock` (M-v0.2.0 PoC) / `homelab` (M-v0.2.1+ real wire) = db (Pi-driven 정규화 적합, 복잡한 nested data) / (3) `metrics` = file (Prometheus scrape = 단순 time-series) / (4) `hrdb` = db (PII + 복잡, Pi-driven). **근거: simple schema source 는 file+rule-based (빠르고 deterministic), complex 사내 시스템 source 는 db+pi-sdk (Pi LLM 이 더 적합)** |
| Q16 | **Pi ingest cron interval default** (2026-06-18 결정, §10.3) | `*/1 *` (1분) / **`*/5 *` (5분)** / `*/15 *` (15분) / `*/30 *` (30분) / 매시간 | **`*/5 *` (5분)** (2026-06-18 결정, §10.3). db mode source 의 raw_records.ingested_at < NOW() - 5 minutes 자동 SELECT + Pi SDK mode ingest. M-v0.2.1+ APScheduler 사용 시 per source custom 가능 (source_meta.ingest_schedule). 5분 간격의 근거: (a) Pi SDK mode 의 latency 1~3초 × 100 raw batch = ~3분 (5분 내 처리 가능) + (b) 외부 시스템 sync 주기 정합 (Gitea 의 push event 가 보통 1~5분) |
| Q17 | **Pi LLM 실패 시 fallback 정책** (2026-06-18 결정, §10.3 + §11.1.3) | 실패 시 skip (next cycle 대기) / 실패 시 queue retry / **실패 시 rule-based 자동 fallback** | **rule-based 자동 fallback** (2026-06-18 결정, §10.3 + §11.1.3). Pi LLM timeout 30초 초과 OR 출력 invalid (§3.9.3 5번 cross-link validation 실패) → source_meta.normalize_mode = "pi-sdk" 일 때 rule-based normalize() method 로 자동 fallback (source plugin 의 `_rule_based_normalize(raw)` override). fallback 발동 시 `audit.pi_ingest_fallback` event + degrade flag. Pi 복귀 시 다음 cycle 부터 자동 정상. **장점: 운영 중지 없이 운영 가능** (LLM 일시 장애 영향 최소화). **단점: rule-based normalize 의 정확도가 LLM 보다 낮음 (인수인계 품질 trade-off)** |
| Q18 | **backend-ai 폐기 timing** (2026-06-18 결정, §5.5 + §6.6.2) | M-v0.2.0 release 직후 / **M-v0.2.2 와 동시** / M-v0.3.0 와 동시 | **M-v0.2.2 와 동시** (2026-06-18 결정, §5.5 M-v0.2.2 DoD + §6.6.2 폐기 절차 10 단계 정합). M-v0.2.2 = 6종 source wire + backend-ai 폐기 동시 (placeholder 정리 + code-removed). **근거**: (a) M-v0.2.0 PoC 에서 backend-knowledge 의 5종 source plugin 정합 검증 후 (b) M-v0.2.1 의 Gitea 4 정식 wire 안정화 후 (c) M-v0.2.2 에서 backend-ai 제거 + metrics 추가 동시 (deployment atomic). M-v0.3.0+ 에서 backend-ai reference 0 검증 (grep `backend-ai` = 0) |

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
| 2026-06-18 | **Q12 결정 (raw storage_mode dual mode)** — §10 신규로 raw 데이터를 file + db dual mode 로 운영. per source `storage_mode` field (source_meta) + default mapping (gitea 4 = file + homelab/hrdb = db + metrics = file) |
| 2026-06-18 | **Q13 결정 (Pi SDK mode timing 변경)** — §6.3 + §10 신규로 Pi 의 역할 3 가지로 확장. (a) M-v0.2.0+ db mode source 의 periodic ingest + (b) M-v0.2.3+ LLM enrich + (c) M-v0.2.3+ cross-link 자동 resolution. original (M-v0.2.3+ enrich only) → 변경 (M-v0.2.0+ periodic ingest) |
| 2026-06-18 | **Q14 결정 (DB type)** — §10.1 로 M-v0.2.0~v0.2.2 sqlite + M-v0.2.3+ PostgreSQL option. backend-core PostgreSQL 공유 ❌ (standalone 정합) |
| 2026-06-18 | **Q15 결정 (per source storage_mode mapping)** — §10.4 default mapping 결정. simple schema source = file+rule-based, complex 사내 시스템 source = db+pi-sdk |
| 2026-06-18 | **Q16 결정 (Pi ingest cron interval default)** — §10.3 로 `*/5 * * * *` (5분) default. M-v0.2.1+ per source custom 가능 (source_meta.ingest_schedule) |
| 2026-06-18 | **Q17 결정 (Pi LLM fallback to rule-based)** — §10.3 + §11.1.3 로 Pi LLM timeout 30초 초과 OR 출력 invalid → source_meta.normalize_mode = "pi-sdk" 일 때 rule-based normalize() 자동 fallback. audit.pi_ingest_fallback event + degrade flag |
| 2026-06-18 | **Q18 결정 (backend-ai 폐기 timing)** — §5.5 M-v0.2.2 DoD + §6.6.2 폐기 절차 10 단계로 M-v0.2.2 와 동시 폐기. M-v0.2.0 PoC 정합 검증 → M-v0.2.1 정식 wire 안정화 → M-v0.2.2 폐기 + metrics 추가 동시 (deployment atomic) |

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
| 2026-06-18 | **5 카테고리 정합 — §3.2.1 보강 + 신규 §3.5 (concept organization) + cross-section 정합 fix 5 위치** — (1) §3.2.1 에 신규 subsection §3.2.1.1 "5 카테고리별 대표 concept frontmatter 예시" 추가 (5 카테고리 = 5 concept 1:1 mapping: issue_tracker/gitea_issue integration / wiki/gitea_wiki reference / scm/gitea_repo_pull integration / cicd/gitea_action event / code_quality 2차 wire placeholder) (2) **신규 §3.5 "Concept organization"** 5 subsection: §3.5.1 원칙 (orthogonal axes — type 8종 + x_devhub_category 5종) / §3.5.2 5×8 matrix (○/△/✗ valid combinations + 5종 PoC source plugin 의 category × type mapping) / §3.5.3 Bundle 디렉터리 구조 (devhub-gitea = 4 category directory + devhub-homelab 5 카테고리 외) + 5개 representative concept frontmatter 예시 (integration/event/metric/runbook/decision) / §3.5.4 index.md 자동 생성 규칙 (per-bundle/per-category/per-type 3종 + `curate/index_builder.py` 구현 정공법 + per-bundle index.md 발췌) / §3.5.5 cross-link 4종 rule (intra-bundle / cross-bundle / source-external / reverse index) + 5개 정책 정공법 (3) cross-section 정합 fix 5 위치: §1.3 progressive disclosure row 보강 (per-bundle/per-type/per-category 명시) + graph row 보강 (4종 cross-link rule + reverse index 명시) + path pattern row `{bundle}/{type}/{slug}.md` → `{bundle}/{category}/{slug}.md` (slug prefix = type, §3.5.3 정합) / §2.1 `curate/index_builder.py` 코멘트 "bundle 별 index.md" → "per-bundle/per-type/per-category index.md (§3.5.4 정합)" + `link_resolver.py` 코멘트 보강 (1차 rule-based / M-v0.2.3+ Pi LLM cross-link 자동 resolution, §3.5.5 정합) / §3.3 frontmatter spec 마지막에 "> **5 카테고리 × 8 type 의 valid combination**: §3.5.2 정합" 1줄 추가 / §3.5.5 intra-bundle link syntax 정합 (현재 dir vs sibling dir 케이스 분리 + 동일 예시 source/target 일관성) / ADR-0034 §4.3 영향 section 에 §3.2.1 + §3.5 row 추가 + §3.3 row 갱신 (`x_devhub_category` 5 enum mention) + ADR-0034 frontmatter 수정일 2026-06-18 갱신 (4) frontmatter 갱신 (umbrella doc 최종 수정일) | self-review (사용자 "1 진행하자. 컨셉 정리를 계속 할거야" + "5 카테고리 정합" 결정 (C 옵션) + "좋아 일단 초안은 이렇게 가져가자" 후속 + path pattern 정합 fix follow-up) |
| 2026-06-18 | **Path Y caller-provided user context — 신규 §3.6 "Data governance & query scoping" + cross-section 정합 fix 8 위치** — (1) **신규 §3.6** 5 subsection: §3.6.1 caller-provided user context (X-DevHub-User-Context HTTP header, base64url(json) 의 user_id/org_id/org_unit_ids/project_ids/roles/request_id/issued_at 7 field schema + DevHub `AppUser`/RBAC 모델 format 호환 + trust model: caller 책임 인증 + backend-knowledge 책임 format 검증만 + endpoint 별 필수/권장/없음 13개 endpoint 표 + OpenAPI security scheme) / §3.6.2 curation governance model (`x_devhub_curator` 별 manual edit permission: rule-based ❌ / llm system_admin 만 with curator=human 승격 / human owner-user self or org_head scope or system_admin) / §3.6.3 query scope priority 4-tier (org > personal > project > public, priority 1 = highest, same concept multiple instances → highest priority 만 노출) / §3.6.4 frontmatter extension 5 field (`x_devhub_owner_org_id` / `_user_id` / `_org_unit_ids` / `_project_ids` / `x_devhub_visibility` 4 enum) / §3.6.5 cross-section 정합 fix 8 위치 (2) cross-section 정합 fix: §1.2 G7 standalone 정책 + caller-provided user context 1줄 추가 / §2.3 시스템 경계 표 "다른 backend 연결 (general)" row + "OIDC / Keycloak" row + "API 인증" row 3 row 갱신 / §3.1 API 매트릭스 인증 정책 노트 갱신 (Path Y 추가) / §3.3 frontmatter spec 정책 노트 마지막에 "> **Path Y governance 필드 (2026-06-18 신규, §3.6.4 정합)**" 1줄 추가 / §4.1 정책 정의 표 "인증" row 갱신 (Path Y 추가) / ADR-0034 §4.3 영향 section 에 §3.6 row 추가 + ADR-0034 frontmatter 수정일 2026-06-18 갱신 / ADR-0035 §3.4 1차 raw API 정책 row caller-provided user context (gateway 책임) 명시 + §3.6.1 cross-reference + ADR-0035 frontmatter 수정일 2026-06-18 갱신 + §4.2 negative/trade-off row 추가 + §4.3 영향 row 추가 (3) frontmatter 갱신 (umbrella doc 최종 수정일) | self-review (사용자 "다음 컨셉 정리 이어서 하자. 카테고리 나눴고 카테고리 별로 데이터 관리... 데이터 정리는 조직에 따라, 사람에 따라 관리... 조회 우선순위는 사용자 조직 > 개인 > 프로젝트" 결정 + explore agent 2 병렬 검색 (user/org/project RBAC + 데이터 정규화/envelope/governance) + Path Y 권장 (caller-provided user context) + 사용자 "Path Y" 결정 + "1 진행하자" 후속) |
| 2026-06-18 | **Data normalization pipeline — 신규 §3.7 "Data normalization pipeline" + cross-section 정합 fix 6 위치** — (1) **신규 §3.7** 5 subsection: §3.7.1 5 step normalization 원칙 (Step 1 외부 시스템 API 호출 + Step 2 raw JSON 봉투 암호화 저장 + Step 3 raw → OKF concept 변환 + Step 4 concept .md emit + Step 5 index.md 자동 생성, 책임 분리 표 6 module: source plugin / 1차 raw storage / OKF bundle storage / rule-based enricher / index_builder / link_resolver) / §3.7.2 per-source type mapping (7 source × types emitted: Gitea 4 sub-plugin + homelab + metrics + hrdb, 5종 PoC = 약 35개 concept 자동 emit) / §3.7.3 cross-source 동질화 (Jira/Gitea/GitHub 모두 `integration_*_issue_puller.md` 형, query 시 cross-source aggregation, viz.html cross-source cluster) / §3.7.4 normalize algorithm pseudocode (`sources/{source}.py` 의 normalize() method 4 step: parse → extract frontmatter → emit body → attach cross-links, `SourceMeta` 11 field) / §3.7.5 edge cases + degraded handling (6 case 표: Partial failure / Schema drift / Source-specific custom transform / Duplicate concept / Large raw / Auth failure, M-v0.2.0 PoC 범위 명시) (2) cross-section 정합 fix 6 위치: §1.3 producer 다중 row 갱신 (rule-based enricher §3.7.1/§3.7.4 명시) / §2.1 sources/ 트리 `sources/{source}.py` + `curate/enricher.py` 코멘트 보강 (§3.7.1 5 step 정합) / §3.2 concept type enum 하단 cross-reference 추가 (§3.7.2 per-source mapping) / §3.5.3 bundle 디렉터리 layout 의 §3.7.1 reference 추가 / ADR-0034 §4.3 영향 section §3.7 row 추가 + ADR-0034 frontmatter 수정일 갱신 (3) frontmatter 갱신 (umbrella doc 최종 수정일) | self-review (사용자 "1 진행하자" (다음 concept organization) + §3.7 "data normalization pipeline (category × system → OKF concept)" 자연스러운 다음 영역 (사용자 명시 "(b) 카테고리별 정규화" 잔여) + DevHub backend-core 의 NormalizeSnapshot() / stateToEventType() / REQ-FR-INT-004/005 패턴 참고 + 사용자 "진행하자" 승인 후속) |
| 2026-06-18 | **Source plugin 작성 정공법 — 신규 §3.8 "Source plugin 작성 정공법" + cross-section 정합 fix 4 위치** — (1) **신규 §3.8** 5 subsection: §3.8.1 SourcePlugin ABC 인터페이스 명세 (Pydantic v2 + 12 type: Credential/SourceMeta/Connection/RawResponse/Concept/FetchQuery/HealthStatus + ABC 의 5 abstract method: connect/fetch/normalize/emit_concept/health_check + registry register/get/list 3 function, `sources/_base.py` 1차 작성 정공법) / §3.8.2 Gitea 4 sub-plugin 정공법 (real wire, M-v0.2.0 PoC 부터, 4 sub-plugin × 5~7 type = 약 26 concept emit, Gitea access token `type=bearer, value=<token>` credential schema, REST API 호출 + normalize() 4 step 공통 패턴) / §3.8.3 homelab_mock 정공법 (filesystem fixture `var/fixtures/homelab/*.json` 기반, 4 type, M-v0.2.0 PoC 단순화, M-v0.2.1+ real wire 교체) / §3.8.4 신규 source 추가 10 step 절차 (Step 1 외부 API spec 정독 → Step 2 SourceMeta 정의 → Step 3 5 method 구현 → Step 4 credential schema Pydantic 모델 → Step 5 body_template per type → Step 6 단위 테스트 → Step 7 e2e smoke → Step 8 bundle layout 결정 + §3.7.2 갱신 → Step 9 representative concept .md 발췌 → Step 10 ADR 영향 section 갱신 + Quality gate 5 항목) / §3.8.5 3 tier 검증 (단위 pytest + 통합 real Gitea instance + e2e smoke pytest + FastAPI TestClient, M-v0.2.0 PoC = 단위 + e2e smoke) (2) cross-section 정합 fix 4 위치: §3.7.2 per-source mapping 표 하단 "**작성 정공법**: source plugin 작성 시 §3.8 정공법 따름" 1줄 추가 / §5.1 M-v0.2.0 scope row 에 "Gitea 통합 4종 (§3.8.2 정공법) + homelab_mock (§3.8.3 정공법) = 5종 PoC, §3.7.2 / §3.8 정합" 갱신 / ADR-0034 §4.3 영향 section §3.8 row 추가 + ADR-0034 frontmatter 수정일 갱신 (3) frontmatter 갱신 (umbrella doc 최종 수정일) | self-review (사용자 "1" (다음 concept organization) + 4-option 질문 응답 "(a) §3.8 Source plugin 작성 정공법" + §3.7 의 abstract 5 step normalization pipeline + ADR-0035 §3.2/§6.4 의 high-level 결정을 구현 가능 정공법으로 구체화 + 5종 PoC source plugin (Gitea 4 sub-plugin + homelab_mock) 의 작성 절차 + 신규 source 추가 절차 + 3 tier 검증 절차 정의) |
| 2026-06-18 | **§4 1차 raw 데이터 API 심화 — 신규 §4.4~§4.7 "raw 운영/API/정합성 정책" + cross-section 정합 fix 4 위치** — (1) **신규 §4.4** raw 운영 정책 (봉투 암호화 `$env$v0.1$...` ADR-0025 + .gitignore per source 8 row 표 + retention default 90일, 예외: metrics 30일/hrdb 365일, 매일 03:00 UTC cron 자동 삭제 + LRU storage quota default 1GB/bundle) / **§4.5** raw API 권한 + visibility (endpoint 별 권한 matrix 4 row: POST = bundle owner_org member / GET = visibility 정합 / list = caller scope filter / DELETE = system_admin OR 등록자 OR owner_org member, visibility 4 enum 재사용 §3.6.2) / **§4.6** raw → concept 정합성 (1 raw → N concepts 관계, sqlite `raw_index.concept_ids` 추적, raw 삭제 시 concept 처리 3 mode: M-v0.2.0 hard_delete / M-v0.2.1+ soft_archive default / M-v0.2.3+ retain_concept 옵션, orphan concept = `x_devhub_status: orphaned` + audit, raw 변경 시 concept update: M-v0.2.0 overwrite / M-v0.2.1+ superseded / M-v0.2.3+ .md.prev history) / **§4.7** raw 정합성 검증 (sha256 hash 저장 + 매 조회 시 재검증, file system source-of-truth, timestamp lag threshold default 5분, audit log 7 event 표) (2) cross-section 정합 fix 4 위치: §2.1 `var/raw/` 트리 코멘트 보강 (봉투 암호화 + .gitignore + retention 90일 + quota 1GB 정합) / §3.6.1 endpoint 표 raw 4 row 갱신 (권장 → 필수, §4.5 정합) / §4.1 정책 정의 표 9 row 보강 (저장 위치 + 메타 + API + 인증 + 정합성 + lifecycle 6 row 신규, sqlite `raw_index` field 6개 추가 sha256/visibility/retention_days/registered_by/concept_ids/last_verified_at) / ADR-0035 §3.4 1차 raw API 정책 row 갱신 (§4.4~§4.7 reference 추가) + ADR-0035 frontmatter 수정일 갱신 (3) frontmatter 갱신 (umbrella doc 최종 수정일) | self-review (사용자 "1" (다음 concept organization) + 4-option 질문 응답 "(a) §4 1차 raw API 심화" + §4.1 정책 정의 표 의 high-level 정의 + ADR-0025 봉투 암호화 + 사용자 명시 'API 로 조회/추가 가능' 정책 + §3.6 governance 정합 raw visibility + 1 raw → N concepts 관계 추적 필요성 + sha256 정합성 검증 + audit log 운영 정책) |
| 2026-06-18 | **OKF concept 운영 lifecycle — 신규 §3.9 "OKF concept 운영 lifecycle" + cross-section 정합 fix 3 위치** — (1) **신규 §3.9** 4 subsection: §3.9.1 lifecycle 5 단계 state machine (created → reviewed → published → active → archived, transition + 책임자 + 정합 section 표, frontmatter status field M-v0.2.1+ 추가) / §3.9.2 frontmatter template per 8 type (dataset/metric/api_endpoint/runbook/integration/event/reference/decision 의 필수 + 권장 field + 예시 표, §3.7.4 normalize() 정합) / §3.9.3 review checklist 5 항목 (frontmatter validation 7 sub / body validation 3 sub / governance validation 2 sub / bundle validation 3 sub / cross-link validation 3 sub, rule-based 자동 + human 수동 review M-v0.2.1+) / §3.9.4 publish + archive 절차 + 운영 정책 (publish trigger 3 mode: rule-based 자동 / human 수동 / system_admin override, archive trigger 3 mode: superseded 자동 / obsolete 수동 / orphan 자동, M-v0.2.0~v0.3.0+ 별 lifecycle 지원 범위 표 5 row) (2) cross-section 정합 fix 3 위치: §3.6.2 curation permission 코드 블록 상단에 §3.9 lifecycle cross-reference note 추가 (created state 의 concept 는 curator 직접 control 불가, archived state 는 write 불가) / §3.8.4 신규 source 추가 10 step 절차의 Step 9 "representative concept .md 발췌 작성" 의 §3.9 cross-reference 추가 (§3.9.2 frontmatter template per 8 type 권장 field 채움 + §3.9.3 review checklist 1~4 항목 자동 validate) / ADR-0034 §4.3 영향 section §3.9 row 추가 + ADR-0034 frontmatter 수정일 갱신 (3) frontmatter 갱신 (umbrella doc 최종 수정일) | self-review (사용자 "1 진행" (다음 concept organization) + 4-option 질문 응답 "(a) §3.9 OKF concept 운영 lifecycle" + §3.5/§3.6/§3.7/§3.8 완료 후 lifecycle 운영 정공법 부재 + M-v0.2.0 PoC = rule-based 자동 publish (frontend 0 page) / M-v0.2.1+ = human 작성 + review workflow (frontend 관리 page) + created/reviewed/published/active/archived 5 단계 state machine + frontmatter template + review checklist + publish/archive trigger 정책) |
| 2026-06-18 | **§5 마일스톤 상세화 — 신규 §5.4~§5.7 + cross-section 정합 fix 3 위치** — (1) **신규 §5.4** dependency graph + critical path (6 마일스톤 linear 의존 표 + critical path linear 표시 + 병렬 가능 sprint 4 case + risk analysis 3 case) / **§5.5** 마일스톤별 DoD (M-v0.2.0-alpha/M-v0.2.0/M-v0.2.1/M-v0.2.2/M-v0.2.3/M-v0.3.0 별 5 항목 DoD: 코드/문서 / 검증 / ADR 영향 / 운영 / cross-section 정합) / **§5.6** cutover 절차 + rollback plan (6 step cutover + 4 trigger rollback 표 + RTO + 5 monitoring 지표 + 8 항목 cutover checklist) / **§5.7** parallel sprint + PR 전략 (per 마일스톤 권장 PR 수 6 row + PR 의존성 정합 + branch prefix 전략 + PR template + 머지 후 처리) (2) cross-section 정합 fix 3 위치: §5.1 마일스톤 표 cross-reference (각 마일스톤 scope 가 §5.5 DoD 와 1:1 정합) / §5.3 sprint 진입 checklist 6 항목 cross-reference (§5.5 M-v0.2.0 DoD 의 (e) cross-section 정합) / ADR-0035 §3.8 마일스톤 표 cross-reference note 추가 (§5 정합, §5.5 DoD 의 (a) 코드/문서 + (b) 검증 정합) + ADR-0035 frontmatter 수정일 갱신 (3) frontmatter 갱신 (umbrella doc 최종 수정일) | self-review (사용자 "1" (다음 concept organization) + 4-option 질문 응답 "(a) §5 마일스톤 상세화" + §5.1/§5.2/§5.3 의 high-level 정의 + M-v0.2.0~v0.3.0 6 마일스톤 별 DoD + cutover + rollback + parallel sprint PR 전략 정의) |
| 2026-06-18 | **§10 DB-based raw + Pi periodic ingest pipeline 신규 + cross-section 정합 fix 5 위치** — (1) **신규 §10** 4 subsection: **§10.1** DB storage + schema (`raw_records` table 14 field + sqlite M-v0.2.0 PoC / PostgreSQL M-v0.2.3+ + 봉투 암호화 ADR-0025 + 4 index idx_raw_records_source_received/visibility/ingested/ingest_lock + `storage_mode` per source_meta 정합) / **§10.2** DB CRUD + 데이터 처리 API 8 endpoint (POST/GET/PATCH/DELETE + list(filter+sort+pagination) + aggregate(group_by/count/sum/avg) + full-text search + ingest-status, Path Y caller-provided user context 필수, OpenAPI security scheme) / **§10.3** Periodic Pi ingest pipeline (SDK mode M-v0.2.0 PoC + 8 step pipeline: scheduler → SELECT raw_records → set ingest_lock → decrypt → Pi LLM normalize → validate → emit OKF concept → update raw_records + trigger lifecycle §3.9 / Pi prompt template j2 / failure handling: timeout 30초 + degraded flag + Pi LLM unreachable → rule-based fallback) / **§10.4** Source path vs DB path 분기 (per source `storage_mode: file|db` + `normalize_mode: rule-based|pi-sdk|pi-rpc` + default mapping 표 7 row: gitea_repo_pull = file/rule-based, gitea_issue = file/rule-based, gitea_wiki = file/rule-based, gitea_action = file/rule-based, homelab_mock = db/pi-sdk, metrics = file/rule-based, hrdb = db/pi-sdk, 운영자 override 가능) (2) cross-section 정합 fix 5 위치: §4.1 정책 정의 표 "저장 위치" + "API" row 갱신 (file path + DB path dual mode 명시) / §3.7 normalization pipeline 에 DB path + Pi driver 추가 (§10.4 storage_mode 분기 명시) / §3.8.1 SourceMeta 에 `storage_mode` + `normalize_mode` + `ingest_schedule` 3 field 추가 + §3.8.4 Step 2 SourceMeta 정의 의 storage_mode 결정 단계 추가 / §6.3 Phase 3 LLM enrich 의 Pi 의 역할 갱신 (Pi 의 3 역할: db mode source periodic ingest M-v0.2.0+ / LLM enrich M-v0.2.3+ / cross-link 자동 resolution M-v0.2.3+) / ADR-0034 §4.3 영향 section §10 row 추가 + ADR-0034 frontmatter 수정일 갱신 (3) frontmatter 갱신 (umbrella doc 최종 수정일) | self-review (사용자 "추가 컨셉 정리해보자. raw에 해당하는 1차 데이터는 종전의 시스템들과 동일하게 db에 데이터를 수집하고 crud를 비롯한 데이터 처리 용 api를 제공할거야. db에 저장된 데이터는 주기적으로 ai (여기서는 pi 경유)를 통하여 ingest하도록 구성해보자." 결정 + 4-option 질문 응답 "(A) default 추천값" (gitea 4 = file/rule-based + homelab/hrdb = db/pi-sdk + metrics = file, Pi SDK mode M-v0.2.0 PoC + sqlite M-v0.2.0 / PostgreSQL M-v0.2.3+ + source_meta 의 storage_mode + normalize_mode field 추가 + 8 DB CRUD endpoint + 8 step Pi ingest pipeline + per source default mapping) |
| 2026-06-18 | **§11 운영 runbook (day-2 운영 정공법) 신규 + cross-section 정합 fix 4 위치** — (1) **신규 §11** 4 subsection: **§11.1** Incident 대응 runbook 6 type (source plugin sync 실패 §11.1.1 / credential 만료 §11.1.2 / Pi ingest pipeline timeout-degraded §11.1.3 / retention cron 실패 §11.1.4 / integrity violation §11.1.5 / archive trigger 실패 §11.1.6 — per trigger / detection / triage / mitigation / recovery 구조, RTO < 30분/< 1시간/< 4시간/< 15분/< 1시간) / **§11.2** Backup + restore 절차 (5 backup 대상: DB / var/bundles/ / var/raw/ / .env-KEK / governance field + per storage mode backup 방법 + retention 정책 일별 7일 / 주별 4주 / 월별 12개월 + restore 5 step 절차 + RTO 5 target + 분기 1회 restore drill) / **§11.3** Monitoring + alert routing (5 monitoring 지표: source plugin sync 성공률 / Query API p95 latency / raw 정합성 violation rate / Pi ingest pipeline success rate / concept archive trigger 정상 작동 + 3 tier alert routing info/warning/critical + alert message template + alert deduplication 5분) / **§11.4** On-call 운영 + role 정의 (4 role: backend-knowledge operator / source plugin developer / Pi LLM curator / security auditor + M-v0.2.0 1 person / M-v0.2.1+ 1주 rotation + Operator training per release) (2) cross-section 정합 fix 4 위치: §1.2 G7 standalone 유지 정책 노트 보강 ("day-2 운영도 standalone — §11 의 운영 runbook 다른 backend 모니터링 도구 공유 ❌") / §4.7 raw 정합성 검증 의 audit log 7 event 표 + §11.3 monitoring 5 지표 cross-reference / §5.6 cutover checklist 8 항목 + §11 운영 runbook 정합 (cutover rollback trigger → §11.1 incident runbook 자동 활성화) / §6.1 Phase 1 운영 정공법 에 "day-2 운영 = §11 운영 runbook 정공법 적용 (M-v0.2.0 PoC = 1 operator, §11.4)" 추가 / ADR-0034 §4.3 영향 section §11 row 추가 + ADR-0034 frontmatter 수정일 갱신 (3) frontmatter 갱신 (umbrella doc 최종 수정일) | self-review (사용자 "1" (다음 concept organization) + 4-option 질문 응답 "(a) §11 운영 runbook" + §4.7 raw 정합성 검증 + §5.6 cutover/rollback 의 운영 연속 + §6 Phase 운영 정공법 + day-2 운영 정공법 (incident 대응 + backup/restore + monitoring/alert + on-call) 의 umbrella doc 본문 정공법 부재) |
| 2026-06-18 | **§6 Phase 1/2/3 운영 정공법 상세 — 신규 §6.5~§6.7 + cross-section 정합 fix 0 위치** — (1) **신규 §6.5** Phase 1 (M-v0.2.0+M-v0.2.1) 운영 정공법 상세 4 subsection: §6.5.1 docker-compose standalone 정합 (`backend-knowledge/docker-compose.yml` 별도, `backend-knowledge/dev-up.sh` 별도, `backend-knowledge-net` 격리, 4 volumes mount + standalone 검증 3 항목) / §6.5.2 mock source + real wire transition (M-v0.2.0 = 5종 source + 1 mock + Gitea 4 real + homelab_mock fixture, M-v0.2.1 = homelab_mock → homelab.py real wire 교체 + storage_mode=db + normalize_mode=pi-sdk) / §6.5.3 gateway + firewall + IP allowlist 정책 (dev = localhost / staging = VPN+사내 CA / production = WAF+allowlist+외부 CA, §2.3 "API 인증 internal-only" 정합) / §6.5.4 5 step e2e smoke pipeline (ingest→raw 확인→index.md 자동 생성→concept 응답→viz.html SSR) (2) **신규 §6.6** Phase 2 (M-v0.2.2) 운영 정공법 상세 3 subsection: §6.6.1 6종 source wire cutover (metrics.py 추가, bundle=devhub-metrics, storage_mode=file, normalize_mode=rule-based, 5 step cutover) / §6.6.2 backend-ai 폐기 절차 10 단계 (디렉터리 + Dockerfile + docker-compose.deploy.yml + Makefile + dev-up.sh + docs + infra + ci workflow + state.json, PR 4 분리) / §6.6.3 6 step e2e 6종 smoke + alert routing 검증 (3) **신규 §6.7** Phase 3 (M-v0.2.3) 운영 정공법 상세 3 subsection: §6.7.1 7종 source wire cutover (hrdb.py 추가, bundle=devhub-hrdb, storage_mode=db, normalize_mode=pi-sdk, PII + complex → §10 Pi-driven) / §6.7.2 Pi 운영 상세 (SDK mode M-v0.2.0~v0.2.2 default / RPC mode M-v0.2.3+ option, mode 선택 기준 표) / §6.7.3 LLM enrich + cross-link 자동 resolution 운영 (`POST /concepts/{id}/enrich` 운영 / `curate/link_resolver.py` 의 unresolved link → Pi LLM 추천, §11 runbook 정합) / §5.6 cutover + §11 monitoring 5 지표 정합 (4) ADR-0034 §4.3 영향 section §6.5~§6.7 row 추가 + ADR-0034 frontmatter 수정일 갱신 (5) frontmatter 갱신 (umbrella doc 최종 수정일) | self-review (사용자 "1" (다음 concept organization) + 4-option 질문 응답 "(a) §6 Phase 1/2/3 상세화" + §6.1/§6.2/§6.3 의 high-level 정공법 + §5.6 cutover 와 §10 DB path 운영 정합 + §11 runbook 의 cutover/rollback 단계와 의미 중복 조정 (서로 다른 초점: §6 = deployment/transition, §11 = incident response)) |
| 2026-06-18 | **§7 Q&A 확장 (Q12~Q18, 11/11 → 18/18 결정 완료) + cross-section 정합 fix 3 위치** — (1) **Q12~Q18 row 추가** (7 row): **Q12** raw storage_mode 결정 (dual mode per source, gitea 4 = file + homelab/hrdb = db + metrics = file, §10.4 정합) / **Q13** Pi SDK mode timing 변경 (M-v0.2.0+ periodic ingest + M-v0.2.3+ LLM enrich + M-v0.2.3+ cross-link 자동 resolution, §6.3 / §10.3 정합) / **Q14** DB type 결정 (M-v0.2.0~v0.2.2 sqlite + M-v0.2.3+ PostgreSQL option, §10.1 정합) / **Q15** per source storage_mode default mapping 결정 (simple schema = file+rule-based, complex 사내 시스템 = db+pi-sdk, §10.4 정합) / **Q16** Pi ingest cron interval default 결정 (`*/5 * * * *` 5분 default, §10.3 정합) / **Q17** Pi LLM fallback 결정 (timeout 30초 초과 OR 출력 invalid → rule-based normalize() 자동 fallback, §10.3 + §11.1.3 정합) / **Q18** backend-ai 폐기 timing 결정 (M-v0.2.2 와 동시, §5.5 + §6.6.2 정합) (2) **§7 제목 갱신** "11/11 결정 완료" → "18/18 결정 완료" (3) cross-section 정합 fix 3 위치: §8 timeline 결정 row 7 row 추가 (Q12~Q18 결정 timeline) / ADR-0035 §4.2 negative/trade-off row 갱신 (§7 Q12~Q18 결정 + 운영 부담 영향) / frontmatter 갱신 (umbrella doc 최종 수정일 + §7 Q&A 확장 cross-section fix 명시) | self-review (사용자 "1" (다음 concept organization) + 4-option 질문 응답 "(a) §7 Q&A 확장" + §10/§11/§6.5~§6.7 추가로 인한 새 Q 7 row 식별 + §7 Q&A 표 = 18/18 결정 완료 + ADR-0035 §4.2 trade-off 영향) |
| 2026-06-18 | **§12 frontend page 상세화 (M-v0.2.1+ 관리/조회 page 1 + viz.html 자가 viewer) 신규 + cross-section 정합 fix 6 위치** — (1) **신규 §12** 5 subsection: **§12.1** M-v0.2.0 viz.html 자가 viewer 상세 (Cytoscape.js v3.x + marked.js v5.x CDN embed + inline style + SVG fallback + viz.html component 5 element + Cytoscape node 7 field + edge 4 type 정합) / **§12.2** M-v0.2.1 frontend 관리/조회 page 1 5 page 상세 (concept list / concept detail / ingest trigger / bundle management / raw inspector + routing 구조 + frontend 기술 선택 vanilla JS/Next.js/Vue.js/Svelte 옵션) / **§12.3** User flow + 권한 매트릭스 3 role (visitor/operator/admin + Path Y caller-provided user context 흐름 + gateway 의 3-step orchestration) / **§12.4** API integration matrix 14 row (per frontend page → backend-knowledge API 1:1 mapping, Path Y user context 필수 정합) / **§12.5** frontend cutover 정책 (7 step M-v0.2.0→v0.2.1 cutover + frontend update 주기 per release + viz.html 단독 vs frontend 통합 운영 + §5.6 cutover 정합) (2) cross-section 정합 fix 6 위치: **§5.1 M-v0.2.1 DoD row** "frontend 관리/조회 page 1" 의 5 page detail (§12.2) cross-reference / **§6.1 Phase 1** "M-v0.2.0 만 frontend 0 page, viz.html 자가 viewer 만 SSR" 의 viz.html 상세 (§12.1) cross-reference / ADR-0035 §3.6 frontend 정책 row + §12.2 5 page + §12.3 3 role + §12.4 API matrix + §12.5 cutover 정책 cross-reference 추가 / ADR-0034 §4.3 영향 section §12 row 추가 + ADR-0034 frontmatter 수정일 갱신 / frontmatter 갱신 (umbrella doc 최종 수정일 + §12 frontend page 상세화 cross-section fix 명시) | self-review (사용자 "1" (다음 concept organization) + 4-option 질문 응답 "(a) §12 frontend page 상세화" + §5.1 M-v0.2.1 DoD + §6.1 Phase 1 frontend 의 viz.html + frontend page 1 의 high-level 정의 → 5 page / 3 role / 14 row API matrix / 7 step cutover 정책 의 concrete 정공법 부재) |
| 2026-06-18 | **§13 cross-cutting 종합 신규 (umbrella doc 전체 cross-reference 정합성 최종 검토) + cross-section 정합 fix 0 row (본 §13 자체가 종합 review)** — (1) **신규 §13** 4 subsection: **§13.1** cross-reference matrix (12 umbrella sections × ADR-0034 / ADR-0035 / state.json / external-integrations-agentic-rag-roadmap.md / docs/llm-wiki mirror 20 row, high/medium/low/none 정합성 검증 결과) / **§13.2** 미해결 cross-section gap 6 row 식별 (cross-link reverse index / Pi prompt template / incident runbook tuning / M-v0.2.0 sprint 진입 checklist 4/6 + 잔여 2 / Pi SDK mode npm dependency / backup schedule cron 등록) / **§13.3** 후속 결정 항목 (post-sprint follow-up) 6 row (GitHub milestone v0.2.0 / state.json M-v0.2.0 row / external-integrations-agentic-rag-roadmap.md status active / docs/llm-wiki mirror / M-v0.2.0 release notes / docs/DOCUMENT_INDEX.md + docs/planning/README.md 갱신) / **§13.4** Cross-cutting 영향 종합 + 정합 검증 결과 (12 항목 ✅, post-sprint 6 row 📋, known gaps 6 row 자연 해소) (2) cross-section 정합 fix 0 row: 본 §13 자체가 종합 review 이므로 fix 불필요. 단, ADR-0034 §4.3 + ADR-0035 §3.6 + frontmatter + §9 변경 이력 row 갱신 | self-review (사용자 "1" (다음 concept organization) + 4-option 질문 응답 "(a) §13 cross-cutting 종합" + 12 commit 후 umbrella doc 전체 cross-reference 정합성 최종 검토 + post-sprint follow-up 항목 종합 (GitHub milestone / state.json / external-integrations-agentic-rag-roadmap.md / docs/llm-wiki mirror / release notes / DOCUMENT_INDEX.md) + known gaps 식별 (cross-link reverse index / Pi prompt template / incident runbook tuning / M-v0.2.0 sprint 진입 checklist 4/6 + 잔여 2 / Pi SDK mode npm dependency / backup schedule cron 등록)) |
| 2026-06-18 | **§1.1 한계 4~7 추가 (2026-06-18 결정에서 식별된 trade-off 한계) + §1.3 How 정당화 강화 (한계 7개 → §3~§12 해결책 cross-reference 표) + cross-section 정합 fix** — (1) **§1.1 한계 4** caller-provided user context 신뢰 문제 (Path Y 의 trust model 한계, §3.6.1 mitigation = format 검증 + §11.3 monitoring audit log, 능동적 위조 탐지 HMAC signature / anomaly detection 은 M-v0.3.0+ scope 외) (2) **§1.1 한계 5** DB-based raw + Pi-driven source 의 운영 복잡도 (dual storage mode 한계, §10.4 default mapping + §10.6 운영 가이드 + §11.3 mode 변경 event audit, storage_mode 변경 CLI tool / dashboard 는 M-v0.2.1+ scope 외) (3) **§1.1 한계 6** backup + DR 복잡도 (dual storage mode 의 transactional 정합성 한계, §11.1.5 integrity violation 자동 trigger backup re-sync + §11.2 backup drill 분기 1회, transactional backup strong consistency 는 M-v0.2.3+ scope 외) (4) **§1.1 한계 7** §12 frontend standalone 유지보수 부담 (lifecycle 분리 한계, §12.1 viz.html 자가 viewer + §12.4 API integration matrix + §5.7 parallel sprint PR 전략, CI 자동 contract test 는 M-v0.2.1+ scope 외) (5) **§1.3 How 정당화 강화** — 한계 7개 (1~3 = v0.1.x 한계 + 4~7 = 2026-06-18 식별 trade-off) → §3~§12 해결책 cross-reference 표 7 row + "정당화 원칙" 1줄 (v0.2.0 PoC 의 trade-off 로 받아들이고 능동적 강화는 후속 milestone scope 분리) + "본 v0.2.0 PoC 운영 중에는 한계 4~7 이 자연 노출, §11 운영 runbook 으로 mitigation 가능" (6) cross-section 정합 fix: umbrella doc frontmatter (최종 수정일 + §1.1/§1.3 영향 명시) + ADR-0034 §4.3 영향 section row 1 추가 (한계 4~7 의 §1.1 명시 + §1.3 정당화 cross-reference + 한계 4~7 의 §3~§12 mitigation 위치 7 row) + ADR-0034 frontmatter 갱신 + ADR-0035 §4.3 영향 section row 1 추가 (한계 4~7 의 §1.1 명시 + §1.3 정당화 cross-reference + §12 frontend 정책이 한계 7 의 baseline mitigation) + ADR-0035 frontmatter 갱신 (7) §9 변경 이력 row 신규 추가 (현재 row) | self-review (사용자 "1" (다음 concept organization) + 4-option 질문 응답 "(a) §1.1 한계 추가 (2026-06-18 결정 반영)" + §13 cross-cutting 종합 후 자연스러운 후속 — 2026-06-18 결정 (Path Y / DB-based raw + Pi / 운영 runbook / frontend standalone) 과정에서 식별된 4가지 trade-off 한계 (caller 신뢰 / dual mode 운영 / backup DR / frontend lifecycle) 의 명시적 식별 + §1.3 의 "How" 정당화를 한계 7개 → 해결책 cross-reference 표로 강화 + 한계 4~7 의 능동적 강화 (HMAC signature / CLI tool / transactional backup / contract test) 는 M-v0.2.1+~M-v0.3.0+ scope 외로 분리 + §11 runbook 으로 mitigation 가능 명시) |

## 10. DB-based raw + Pi periodic ingest pipeline (2026-06-18 신규)

**Motivation**: 기존 §4 (file-based raw) + §3.7 (rule-based normalization pipeline) 의 한계 보완. 사용자 결정 (2026-06-18): "raw 에 해당하는 1차 데이터는 종전의 시스템들과 동일하게 db 에 데이터를 수집하고 crud 를 비롯한 데이터 처리 용 api 를 제공할거야. db 에 저장된 데이터는 주기적으로 ai (여기서는 pi 경유) 를 통하여 ingest 하도록 구성해보자."

**핵심 변화**:
1. **raw storage = file system (기존 §4) + DB (신규 §10) dual mode** — per source `storage_mode: file|db` field
2. **DB CRUD + 데이터 처리 API** — SQL CRUD (POST/GET/PATCH/DELETE) + sort/filter/aggregate/search (사용자 명시 "종전의 시스템들과 동일하게")
3. **Periodic Pi ingest pipeline** — DB 의 raw 데이터를 Pi SDK mode (subprocess per call, M-v0.2.0 PoC simple) 로 **주기적** read → LLM normalize → OKF concept emit

**독립 backend-core 정합**: backend-core 의 DB 사용 패턴 참고 (e.g., `backend-core/internal/store/`, sqlc/Pgx 구조). 단, **DB schema import / sqlc generated code 공유 ❌** (§1.2 G7 standalone 정합). backend-knowledge 자체 DB schema 작성 + SQL 직접 작성 (또는 자체 sqlc generated).

**Source path vs DB path 분기** (per source `storage_mode`):
- **gitea 4 sub-plugin**: `storage_mode=file` (rule-based normalize 적합, simple REST schema)
- **homelab**: `storage_mode=db` (Pi-driven 적합, 복잡 + 사내 시스템)
- **metrics**: `storage_mode=file` (operational metric, 단순)
- **hrdb**: `storage_mode=db` (PII, complex, Pi-driven 정규화 필요)
- 운영자 override 가능 (source_meta 의 `storage_mode` field)

### 10.1 DB-based raw storage + schema

**DB 선택** (per milestone):
- **M-v0.2.0 PoC**: **sqlite** (단순, file-based, `var/raw_index.db` 위치). 단일 file 이므로 git 가능 시 `.gitignore` 권장 (§4.4 정합)
- **M-v0.2.1+**: sqlite 유지 (운영 부담 적음)
- **M-v0.2.3+**: **PostgreSQL** option 추가 (production-grade, 동시성, backup/restore 정합). 단, **backend-core 의 PostgreSQL 와 공유 ❌** (standalone 정합)

**`raw_records` table schema** (sqlite / PostgreSQL 호환):

```sql
CREATE TABLE raw_records (
    id BIGSERIAL PRIMARY KEY,                          -- internal PK
    source TEXT NOT NULL,                              -- e.g., 'gitea_repo_pull'
    slug TEXT NOT NULL,                                -- unique within source
    bundle TEXT,                                       -- e.g., 'devhub-gitea'
    category TEXT,                                     -- e.g., 'scm' (5 enum)
    data_json_encrypted BLOB NOT NULL,                 -- 봉투 암호화된 raw JSON (ADR-0025 정합)
    data_json_hash_sha256 TEXT NOT NULL,               -- sha256 hash (정합성 검증)
    received_at TIMESTAMP NOT NULL,                    -- ingest 시각 (DB write 시점)
    source_timestamp TIMESTAMP,                        -- 외부 시스템 응답의 updated_at (lag 계산용)
    registered_by_user_id TEXT,                        -- POST /db/raw 시 caller
    visibility TEXT NOT NULL DEFAULT 'org',            -- 4 enum (§3.6.4)
    retention_days INT NOT NULL DEFAULT 90,            -- §4.4 retention 정합
    -- DB path specific fields
    storage_mode TEXT NOT NULL DEFAULT 'db',           -- 'file' or 'db' (per source_meta)
    ingested_at TIMESTAMP,                             -- 마지막 Pi ingest 시각 (§10.3)
    last_concept_id TEXT,                              -- 마지막 emit 된 concept ID (§10.3)
    ingested_count INT NOT NULL DEFAULT 0,             -- Pi ingest 횟수
    ingest_locked_until TIMESTAMP,                     -- 동시 ingest 방지 lock
    UNIQUE (source, slug)
);

CREATE INDEX idx_raw_records_source_received ON raw_records (source, received_at DESC);
CREATE INDEX idx_raw_records_visibility ON raw_records (visibility, source);
CREATE INDEX idx_raw_records_ingested ON raw_records (ingested_at) WHERE ingested_at IS NULL;
CREATE INDEX idx_raw_records_ingest_lock ON raw_records (ingest_locked_until) WHERE ingest_locked_until > NOW();
```

**§4 file-based raw 와의 정합**: `storage_mode=file` 인 source 는 §4 의 `var/raw/{source}/{slug}.json` + `sqlite raw_index` (메타만) 사용. `storage_mode=db` 인 source 는 본 §10.1 의 `raw_records` table 사용. **두 모드 동시 운영 가능** (per source 분기).

**§3.6 governance 정합**: `visibility` 4 enum + `registered_by_user_id` = §3.6.2 curation governance 와 동일. caller-provided user context (§3.6.1) 필수.

### 10.2 DB CRUD + 데이터 처리 API

**8 endpoint** (Path Y caller-provided user context 필수, §3.6.1 + §4.5 정합):

| Endpoint | Method | 권한 (§3.6.2 정합) | 응답 |
| --- | --- | --- | --- |
| `/api/v0-2/db/raw` | POST | caller 가 source 의 bundle owner_org member OR system_admin | 201 + raw_id |
| `/api/v0-2/db/raw/{id}` | GET | visibility 정합 (§4.5) | 200 + raw record (data_json decrypted) |
| `/api/v0-2/db/raw/{id}` | PATCH | registered_by_user_id == caller.user_id OR system_admin | 200 + updated raw_id |
| `/api/v0-2/db/raw/{id}` | DELETE | system_admin OR registered_by_user_id == caller.user_id | 200 + `{deleted: true}` |
| `/api/v0-2/db/raw?source=...&since=...&sort=...&limit=...&offset=...` | GET (list) | visibility 정합 (caller scope filter) | 200 + items (paginated) |
| `/api/v0-2/db/raw/aggregate?group_by=...&agg=count\|sum\|avg&field=...` | GET (aggregate) | visibility 정합 | 200 + aggregate result |
| `/api/v0-2/db/raw/search?q=...&field=data_json.key\|source\|slug` | GET (full-text search) | visibility 정합 | 200 + matching items |
| `/api/v0-2/db/raw/ingest-status?source=...&since=...` | GET (Pi ingest status) | visibility 정합 | 200 + `{raw_id, ingested_at, last_concept_id, ingested_count}` |

**SQL 데이터 처리 API** (사용자 명시 "데이터 처리용 API"):
- **sort**: `?sort=received_at:desc` / `?sort=data_json_hash_sha256:asc` 등 multi-column
- **filter**: `?filter=source.eq(gitea_repo_pull)&filter=visibility.eq(org)` (jsonpath 또는 SQL where 절)
- **pagination**: `?limit=100&offset=0` (default limit 100, max 1000)
- **aggregate**: `?group_by=source&agg=count` (count / sum / avg / min / max)
- **full-text search**: sqlite FTS5 (M-v0.2.0 PoC) / PostgreSQL `tsvector` (M-v0.2.3+)

**OpenAPI extension**:
```yaml
/api/v0-2/db/raw:
  post:
    summary: "Insert raw record into DB"
    requestBody:
      required: true
      content:
        application/json:
          schema:
            type: object
            required: [source, slug, data_json]
            properties:
              source: {type: string}
              slug: {type: string}
              bundle: {type: string}
              category: {type: string}
              data_json: {type: object}
              visibility: {type: string, enum: [org, personal, project, public]}
              retention_days: {type: integer, default: 90}
    responses:
      '201':
        description: "raw record inserted"
      '403':
        description: "E_FORBIDDEN (caller 가 owner_org member 아님)"
```

**§4 file-based raw API 와의 정합**: §4 의 4 endpoint (POST/GET/list/DELETE raw) 는 file-based 한정. 본 §10 의 8 endpoint 는 DB-based 한정. **API path prefix 로 구분** (`/api/v0-2/raw` vs `/api/v0-2/db/raw`).

### 10.3 Periodic Pi ingest pipeline

**Pi integration mode** (2026-06-18 결정):
- **M-v0.2.0 PoC**: **SDK mode** (subprocess per call, `@earendil-works/pi-coding-agent` npm pkg via Node subprocess, §2.2 정합). 단순, stateless, 매 ingest job 마다 subprocess 시작/종료.
- **M-v0.2.3+**: **RPC mode** option 추가 (long-running JSON over stdin/stdout, §2.2 정합). 동일 process 내에서 여러 ingest job 동시 처리, latency 감소.

**Ingest pipeline** (`backend-knowledge/ingest/pipeline.py`):

```
Step 1: Scheduler cron (매 N 분, per source config)
  ↓
Step 2: SELECT raw_records WHERE ingested_at IS NULL OR ingested_at < threshold
        AND ingest_locked_until IS NULL OR ingest_locked_until < NOW()
        LIMIT batch_size (default 100)
  ↓
Step 3: Per raw record:
  - Set ingest_locked_until = NOW() + 5 minutes (atomic, 동시 ingest 방지)
  - Decrypt data_json (ADR-0025 봉투 암호화)
  ↓
Step 4: Pi LLM normalize (SDK mode: subprocess 호출, RPC mode: stdin/stdout)
  - Prompt: "다음 raw JSON 을 OKF concept (type 8종 중 1, x_devhub_category 5종 중 1, x_devhub_* 5 governance field 포함) 으로 변환해주세요"
  - Input: {raw_json, source_meta (bundle/category/type/credential/owner_org)}
  - Output: {frontmatter, body, suggested_type, suggested_category}
  ↓
Step 5: Validate Pi output (§3.9.3 review checklist 의 1~4 자동)
  - 8 type enum 정합
  - x_devhub_category 5 enum 정합
  - x_devhub_visibility 4 enum 정합
  - x_devhub_owner_org_id 존재
  ↓
Step 6: Emit OKF concept (§3.5.3 path pattern)
  - var/bundles/{bundle}/{category}/{type}_{slug}.md
  - x_devhub_curator: "llm" (M-v0.2.3+ 부터, §3.6.2 정합)
  ↓
Step 7: Update raw_records
  - ingested_at = NOW()
  - last_concept_id = emitted concept slug
  - ingested_count += 1
  - ingest_locked_until = NULL
  ↓
Step 8: Trigger concept lifecycle (§3.9 정합)
  - new concept = `x_devhub_status: active`
  - supersede old concept (있으면) = `x_devhub_status: archived`
```

**Pi prompt template** (`backend-knowledge/ingest/prompts/{source}.j2`):

```jinja2
You are a backend-knowledge curator. Convert the following raw external system data into an OKF concept.

## Source metadata
- source: {{ source }}
- bundle: {{ bundle }}
- category: {{ category }}
- expected type (one of 8 enums): {{ expected_type }}
- owner_org_id: {{ owner_org_id }}

## Raw JSON
```json
{{ raw_json }}
```

## Output format (Markdown frontmatter + body)
```yaml
---
type: {{ expected_type }}  # or another from 8 enums if better
title: "..."
description: "..."
# ... (full frontmatter spec per §3.3 + §3.6.4)
---

# Body in Markdown
```

## Constraints
- Type MUST be one of: dataset, metric, api_endpoint, runbook, integration, event, reference, decision
- x_devhub_category MUST be one of: issue_tracker, wiki, scm, cicd, code_quality
- x_devhub_visibility MUST be one of: org, personal, project, public
- x_devhub_curator MUST be "llm" (this is LLM-generated)
- Body MUST be in Markdown with proper headings
- Cross-links to related concepts MUST be valid path

Return ONLY the OKF concept (frontmatter + body), no extra explanation.
```

**Scheduler** (per source `ingest_schedule` field):
- default: 매 5분 cron (`*/5 * * * *`)
- M-v0.2.0 PoC: simple cron (Python `schedule` lib, in-process)
- M-v0.2.1+ : APScheduler (async, multi-worker)
- M-v0.2.3+ : external scheduler (Celery + Redis, 또는 kubernetes cron job)

**Failure handling** (Pi ingest 실패 시):
- Pi subprocess timeout (default 30초) → `ingest_locked_until = NOW() + 5 minutes` (retry 후 다음 cycle)
- Pi LLM 출력 invalid (§3.9.3 5번 cross-link validation 실패) → `x_devhub_degraded_fields` array 에 사유 기록, ingest_count 는 증가시키지 않음
- Pi LLM unreachable → §3.7.5 fallback (rule-based normalize 로 자동 전환, audit log)

### 10.4 Source path vs DB path 분기 + 운영 정책

**Per source `storage_mode` 결정** (2026-06-18 default):

| Source plugin | storage_mode | normalize mode | rationale |
| --- | --- | --- | --- |
| `gitea_repo_pull` | `file` | rule-based | Gitea REST API 가 well-documented, simple schema. rule-based 가 빠르고 deterministic |
| `gitea_issue` | `file` | rule-based | 동일 |
| `gitea_wiki` | `file` | rule-based | 동일 |
| `gitea_action` | `file` | rule-based | 동일 |
| `homelab_mock` (M-v0.2.0) / `homelab` (M-v0.2.1+ real) | `db` | Pi SDK | 사내 시스템 + 복잡한 nested data. Pi LLM 이 더 적합 |
| `metrics` | `file` | rule-based | Prometheus scrape = 단순 time-series data |
| `hrdb` | `db` | Pi SDK | PII + 복잡한 schema. Pi LLM 이 semantic context 풍부 |

**Per source config** (`SourceMeta.storage_mode` + `SourceMeta.ingest_schedule`):

```python
class SourceMeta(BaseModel):
    # ... (§3.8.1 정합)
    storage_mode: Literal["file", "db"] = "file"
    ingest_schedule: Optional[str] = None  # cron expression, file mode 는 미설정
    normalize_mode: Literal["rule-based", "pi-sdk", "pi-rpc"] = "rule-based"
```

**M-v0.2.0 PoC 범위**:
- `gitea_repo_pull` / `gitea_issue` / `gitea_wiki` / `gitea_action` = file + rule-based (기존 §3.7 / §3.8 정합)
- `homelab_mock` = **db + pi-sdk** (DB path 의 첫 PoC source)
- `metrics` / `hrdb` = M-v0.2.2 / v0.2.3 까지 (DB path 정합 후 추가)

**운영 정책** (per source storage_mode):

| Mode | 파일 | DB | Cron | 정규화 |
| --- | --- | --- | --- | --- |
| **file + rule-based** (§3.7) | `var/raw/{source}/{slug}.json` | (사용 안 함) | POST /ingest/{source}/sync (수동) | source plugin 의 normalize() (§3.7.4) |
| **db + pi-sdk** (§10) | (사용 안 함) | `raw_records` table | cron 자동 (매 5분 default) | Pi subprocess (§10.3) |
| **db + rule-based** (M-v0.2.1+ 옵션) | (사용 안 함) | `raw_records` table | cron 자동 | source plugin 의 normalize() + DB read |

**M-v0.2.0 PoC 운영 검증**:
- (a) §3.7 의 5종 PoC source plugin (file + rule-based) 정상 작동 (4 Gitea sub-plugin + homelab_mock, **단 homelab_mock 만 db + pi-sdk 으로 분리**)
- (b) §10 의 DB path + Pi SDK mode 정상 작동 (homelab 의 raw 를 Pi 가 ingest)
- (c) 2 path 동시 운영 검증 (file path 와 DB path 의 concept 모두 var/bundles/ 에 emit)

**Cross-section 정합 fix** (5 위치):
1. **§4.1 정책 정의 표** — "저장 위치" row 갱신: "file path (file mode) OR DB `raw_records` table (db mode, §10.1 정합)"
2. **§3.7 normalization pipeline** — §3.7.1 5 step 에 step 2 의 "raw JSON 봉투 암호화 저장" 보강: "file mode = var/raw/, db mode = raw_records table (봉투 암호화)"
3. **§3.8 source plugin 작성 정공법** — §3.8.1 SourcePlugin ABC 의 SourceMeta 에 `storage_mode` + `normalize_mode` field 추가 (§10.4 정합). 신규 source 추가 10 step 절차의 Step 2 SourceMeta 정의 에 storage_mode 결정 단계 추가
4. **§6.3 Phase 3 LLM enrich** — Pi 의 역할 갱신: "M-v0.2.3+ LLM enrich (rule-based enricher 의 보강) + M-v0.2.0+ Pi periodic ingest pipeline (DB path source 의 정규화)". §6.3 의 M-v0.2.3 timing 이 M-v0.2.0 으로 option 이동
5. **ADR-0034 §4.3 영향** — §10 row 추가 + §3.7/§3.8/§6.3 cross-reference 명시
6. **ADR-0035 §3 영향** — §10 의 DB-based raw + Pi periodic ingest 가 ADR-0035 의 "외부 시스템 7종 source 만 단방향" 정책과 정합 (§10 의 DB path 는 외부 시스템 데이터 + Pi LLM 처리, backend-core wire ❌) 명시

## 11. 운영 runbook (Day-2 운영 정공법, 2026-06-18 신규)

**Motivation**: §4.7 raw 정합성 검증 (audit log 7 event) + §5.6 cutover + rollback plan (4 trigger rollback) + §6 Phase 운영 정공법 의 정공법이 incident 발생 시 **운영자가 즉시 대응 가능한 runbook 형태**로 정리되지 않음. 본 §11 이 day-2 운영 정공법 (incident 대응 + backup/restore + monitoring/alert + on-call) 을 정의.

**독립 backend-knowledge 운영 정합** (§1.2 G7 + §2.3 standalone 정합):
- backend-knowledge 자체 monitoring + alerting 만 정의 (다른 backend 의 monitoring 도구 공유 ❌)
- 운영자 또는 별도 agent 가 대응 (backend-core 의 operator 호출 ❌, §2.3 정합)

### 11.1 Incident 대응 runbook (6 type, per trigger/detection/triage/mitigation/recovery)

**6 incident type** (severity + RTO 정의):

| # | Incident type | Severity | Detection | RTO (Recovery Time Objective) | Runbook section |
| --- | --- | --- | --- | --- | --- |
| **1** | **Source plugin sync 실패** (HTTP 4xx/5xx, network timeout) | warning → critical (반복 시) | §4.7 `raw.received` audit log 의 `last_error` + source plugin health check endpoint (§3.1) | < 30분 | §11.1.1 |
| **2** | **Credential 만료** (401/403 from external system) | critical | source plugin health check 의 401/403 + `raw.deleted` audit log | < 1시간 | §11.1.2 |
| **3** | **Pi ingest pipeline timeout/degraded** (subprocess hang or LLM invalid) | warning | §10.3 `ingest_locked_until` 만료 + degraded flag + `ingest_locked_until > NOW()` index | < 30분 | §11.1.3 |
| **4** | **Retention cron 실패** (DB write 실패 or storage quota 초과) | warning | 매일 03:00 UTC cron 의 audit log + storage quota 90% 임계 | < 4시간 | §11.1.4 |
| **5** | **Integrity violation** (sha256 mismatch on raw query) | critical | §4.7 `raw.integrity_violation` audit log + 매 조회 시 hash 재계산 | < 15분 | §11.1.5 |
| **6** | **Archive trigger 실패** (raw 변경 시 concept superseded 실패) | warning | §3.9 lifecycle `x_devhub_status` 자동 archive 실패 audit + 시각 비교 | < 1시간 | §11.1.6 |

#### 11.1.1 Source plugin sync 실패

- **Trigger**: `POST /api/v0-2/ingest/{source}/sync` 호출 시 HTTP 4xx/5xx 또는 network timeout
- **Detection**:
  - source plugin `health_check()` method 의 `last_error` 필드 (§3.8.1)
  - `audit.raw.received` log 의 `last_error` 컬럼 (§4.7)
  - GET `/api/v0-2/ingest/{source}/status` endpoint (§3.1)
- **Triage**:
  1. `last_error` 메세지 확인 (HTTP 401 → §11.1.2 credential 만료 / 429 → rate limit / 5xx → 외부 시스템 장애)
  2. 외부 시스템 status page 확인 (Gitea / homelab / metrics / hrdb)
  3. network connectivity 확인 (`curl`, `ping`)
- **Mitigation**:
  - 401/403 → §11.1.2 credential rotation
  - 429 → source plugin 의 retry policy 확인 (`§3.7.5 partial failure` 의 degraded flag 활용)
  - 5xx → external system incident → 일시적 정지 + 다음 cron cycle 대기
- **Recovery**:
  - sync 성공 시 audit log 에 `raw.received` success event
  - 5회 연속 실패 시 critical alert (§11.3 routing) + on-call page

#### 11.1.2 Credential 만료

- **Trigger**: source plugin health check 의 401/403 응답
- **Detection**:
  - `audit.raw.received.last_error` 에 "401 Unauthorized" 또는 "403 Forbidden"
  - health_check endpoint 의 `healthy: false`
- **Triage**:
  1. 외부 시스템의 credential 유효성 확인 (Gitea access token / homelab agent token / Prometheus scrape credential / hrdb DB password)
  2. credential 만료 / 회수 여부 확인 (운영자 또는 외부 시스템 admin)
- **Mitigation**:
  - **M-v0.2.0 PoC**: credential rotation 절차 (운영자 수동, 외부 시스템 admin 의 새 credential 발급 + backend-knowledge 의 source plugin 의 credential config 업데이트)
  - **M-v0.2.1+**: credential 자동 rotation (외부 시스템 API 지원 시)
- **Recovery**:
  - 새 credential 로 sync 성공 시 audit log success
  - 1시간 이내 미복구 시 escalation (§11.4)

#### 11.1.3 Pi ingest pipeline timeout/degraded

- **Trigger**: Pi subprocess hang (>30초 timeout) or Pi LLM 출력 invalid (§10.3 step 5 validation 실패)
- **Detection**:
  - `raw_records.ingest_locked_until > NOW()` index → lock 풀리지 않은 raw
  - degraded flag (`x_devhub_degraded_fields`) 증가
  - §10.3 fallback 발동 (Pi LLM unreachable → rule-based fallback)
- **Triage**:
  1. Pi subprocess 상태 확인 (`pgrep`, `top`)
  2. Pi LLM 출력 샘플 확인 (마지막 5개 ingest 의 output)
  3. Pi vendor 의 status page 확인 (`pi.dev` / vendor API)
- **Mitigation**:
  - Pi subprocess timeout → `ingest_locked_until = NULL` (수동, 다음 cycle 에서 retry)
  - Pi LLM invalid → §10.3 fallback 으로 rule-based normalize 자동 전환
  - Pi vendor outage → §10.3 fallback 으로 rule-based 일시 운영
- **Recovery**:
  - Pi 복귀 시 자동 detect → fallback 해제 + rule-based 로 이미 처리된 raw 는 Pi 로 재처리 옵션

#### 11.1.4 Retention cron 실패

- **Trigger**: 매일 03:00 UTC cron 의 retention_days 초과 raw 자동 삭제 실패
- **Detection**:
  - cron audit log 미기록 (cron 자체 fail)
  - storage quota 90% 임계 초과 alert (§11.3 monitoring)
- **Triage**:
  1. cron daemon 상태 확인 (`crontab -l`, `systemctl status cron`)
  2. retention_days 컬럼 / WHERE 조건 검증
  3. raw_records 의 received_at + retention_days < NOW() row count
- **Mitigation**:
  - cron daemon 재시작
  - retention_days 값 잘못 설정된 raw 수동 삭제 (`DELETE FROM raw_records WHERE received_at + retention_days < NOW()`)
- **Recovery**:
  - cron 정상 작동 시 다음 03:00 UTC 부터 자동 retention 재개
  - storage quota 90% 이하 회복 시 alert 해제

#### 11.1.5 Integrity violation

- **Trigger**: §4.7 raw 정합성 검증 의 sha256 hash mismatch
- **Detection**:
  - `audit.raw.integrity_violation` event (severity: high)
  - GET `/api/v0-2/raw/{type}/{name}` 응답의 `E_INTERNAL` ("raw.integrity_violation")
- **Triage**:
  1. source_timestamp vs received_at 비교 (외부 시스템 응답 시점 vs DB write 시점)
  2. fs vs sqlite 의 mtime 비교 (file mode 의 경우)
  3. raw_records 의 sha256 vs file 재계산 비교 (db mode 의 경우)
- **Mitigation**:
  - source plugin 의 sync 재실행 (`POST /api/v0-2/ingest/{source}/sync` 로 raw 재-emit)
  - DB mode 의 경우: `UPDATE raw_records SET data_json_encrypted = ?, data_json_hash_sha256 = ? WHERE id = ?` (운영자 수동, source_plugin 의 fetch 결과로)
- **Recovery**:
  - hash 일치 후 audit log success
  - integrity violation alert 해제
  - 1시간 이내 미복구 시 critical alert

#### 11.1.6 Archive trigger 실패

- **Trigger**: §3.9 lifecycle 의 superseded 자동 archive 실패 (raw 변경 시)
- **Detection**:
  - `x_devhub_status: superseded` 로 변경 안 된 previous concept 발견 (audit log + sqlite `concept_index.status` 비교)
- **Triage**:
  1. source plugin 의 sync 결과 + normalized concept 비교
  2. `x_devhub_version` 비교 (이전 version 이 active 상태로 남았는지)
  3. bundle/index.md 의 archive 표시 정합
- **Mitigation**:
  - 운영자 수동 archive (`POST /api/v0-2/concepts/{id}/enrich` + curator="human" 승격)
  - bundle/index.md 재생성 (`POST /api/v0-2/bundles/{bundle}/rebuild`)
- **Recovery**:
  - archive 정상 적용 후 audit log success
  - viz.html 에서 superseded concept 가 archived 표시

### 11.2 Backup + restore 절차

**Backup 대상** (per storage mode, 2026-06-18 결정):

| 대상 | file mode | db mode | Backup 방법 | Schedule |
| --- | --- | --- | --- | --- |
| **DB (raw_records)** | (사용 안 함) | sqlite `.db` file (M-v0.2.0~v0.2.2) / PostgreSQL `pg_dump` (M-v0.2.3+) | 파일 copy / pg_dump | 일별 (cron 02:00 UTC) + 매 cutover 직전 |
| **var/bundles/** | git push 가능 (Markdown + frontmatter) | 동일 | git push | 매 commit (CI 자동) |
| **var/raw/** (file mode) | 봉투 암호화 후 git push 가능 OR .gitignore + 별도 backup | (사용 안 함) | git push OR tar + S3 upload | 일별 + 매 ingest 직후 (snapshot) |
| **.env** / **KEK** | 봉투 암호화 키 / 외부 시스템 credential | 동일 | 별도 secure storage (1Password / HashiCorp Vault) | 수동 (즉시) |
| **bundle_owner_org_id / governance field** | git push (var/bundles/ 안) | 동일 | git push | 매 commit |

**Backup retention**:
- 일별 backup: 7일 보관 (rolling)
- 주별 backup: 4주 보관
- 월별 backup: 12개월 보관
- KEK / .env: 별도 secure storage 의 backup 정책 따름

**Restore 절차** (per target):

```
Step 1: 최신 backup 파일 확인
  - DB: ls -lt backups/db/ | head -5
  - bundles: git log --oneline var/bundles/ | head -5
  - raw (file mode): ls -lt backups/raw/ | head -5

Step 2: 백업 시점의 dependencies 정합 확인
  - .env / KEK 가 백업 시점과 동일한지 (credential 회전 후 잘못된 credential 사용 방지)
  - bundle layout / source plugin metadata 가 백업 시점과 일치하는지

Step 3: Restore 실행
  - DB (sqlite): cp backup.sqlite var/raw_index.db
  - DB (PostgreSQL): psql -f backup.sql
  - bundles: git checkout <commit> -- var/bundles/
  - raw (file mode): tar -xzf backup.tar.gz -C var/raw/

Step 4: Restore 검증
  - sha256 hash 재계산 (§4.7 정합)
  - source plugin health check
  - query API 의 smoke test (3개 endpoint)

Step 5: Audit log 기록
  - audit.restore.executed event (timestamp + backup_file + operator)
  - audit.restore.verified event
```

**Restore RTO** (per target):
- DB (sqlite): < 5분 (file copy + verify)
- DB (PostgreSQL): < 30분 (pg_dump + psql)
- bundles (git): < 5분 (git checkout)
- raw (file mode tar): < 15분 (untar + verify sha256)
- KEK / .env: < 1시간 (별도 secure storage 에서 retrieve)

**Restore drill** (per quarter):
- 월 1회 dry-run restore (운영 staging 환경에서)
- 분기 1회 실 restore drill (production 환경 test data 로)

### 11.3 Monitoring + alert routing

**5 monitoring 지표** (§5.6 cutover monitoring 과 정합):

| # | 지표 | 측정 방법 | Threshold (warning) | Threshold (critical) |
| --- | --- | --- | --- | --- |
| **1** | **Source plugin sync 성공률** (per source) | `audit.raw.received` event 의 success/failure ratio (24h sliding window) | < 99% | < 95% |
| **2** | **Query API p95 latency** | FastAPI middleware 의 response time histogram | > 500ms (M-v0.2.0 PoC), > 200ms (M-v0.2.1+) | > 1s |
| **3** | **Raw 정합성 violation rate** (per day) | `audit.raw.integrity_violation` event count | > 0.01% | > 0.1% |
| **4** | **Pi ingest pipeline success rate** | `audit.pi_ingest` event 의 success/failure ratio (1h sliding window, §10.3) | < 95% | < 80% |
| **5** | **Concept archive trigger 정상 작동** | §3.9 superseded archive 실패 audit count | > 0/day | > 5/day |

**Monitoring 도구** (M-v0.2.0 PoC default):
- FastAPI middleware + structlog (JSON log)
- Prometheus exporter (`/metrics` endpoint, M-v0.2.1+)
- Grafana dashboard (5 지표 panel, M-v0.2.1+)
- 외부: Datadog / Sentry (M-v0.2.3+ 옵션)

**Alert routing** (3 tier):

| Severity | Channel | Response time | Escalation |
| --- | --- | --- | --- |
| **info** | Slack `#backend-knowledge-info` (M-v0.2.1+) | 1 business day | (없음) |
| **warning** | Slack `#backend-knowledge-alerts` | 1시간 | on-call responder (§11.4) |
| **critical** | Slack `#backend-knowledge-critical` + on-call page (PagerDuty / Opsgenie) | 15분 | on-call responder → 30분 미대응 시 team lead → 1시간 미대응 시 director |

**Alert message template**:
```
[SEVERITY] backend-knowledge incident

Incident: {incident_type}
Source: {source_plugin or N/A}
Trigger: {trigger_description}
Detection time: {timestamp}
Affected scope: {raw_count affected, concept_count affected}
Suggested runbook: §11.1.{N}
On-call: @{responder}

#backend-knowledge #v0.2.0
```

**Alert deduplication**: 같은 incident type + 같은 source 가 5분 이내 반복 발생 시 1개 alert 만 발송 (노이즈 방지)

### 11.4 On-call 운영 + role 정의

**4 role** (M-v0.2.0~v0.2.1+):

| Role | 책임 | 권한 | On-call schedule |
| --- | --- | --- | --- |
| **backend-knowledge operator** | backend-knowledge 의 day-2 운영 전체 (incident 대응, restore, monitoring 확인) | system_admin role + §11.3 의 alert 수신 + restore 권한 | M-v0.2.0 = 1 person (project lead), M-v0.2.1+ = 1주 rotation (4 person team) |
| **source plugin developer** | per source 의 fetch / normalize / sync 구현 (incident 발생 시 code fix) | source_meta 의 bundle owner_org_unit_ids 의 org_head scope + source plugin 코드 read/write | (없음, 필요 시 on-call) |
| **Pi LLM curator** (M-v0.2.3+) | Pi ingest pipeline 의 prompt + vendor 관리 | system_admin role + Pi 의 vendor API key | M-v0.2.3+ = 1 person |
| **security auditor** | incident 중 보안 관련 (credential leak, supply chain) 의 escalation | `audit.integrity_violation` + `audit.security` event 수신 | (없음, 필요 시 page) |

**On-call rotation 정책** (M-v0.2.1+):
- 1주 rotation (월요일 09:00 ~ 다음주 월요일 09:00 KST)
- backup on-call: rotation 의 다음 사람 (즉시 응답 불가 시)
- handoff 절차: rotation 시작 시 §11.3 의 dashboard snapshot + 진행 중인 incident list 공유

**Operator training** (per release):
- M-v0.2.0 release 직전: §11.1 incident runbook walkthrough (1 hour)
- M-v0.2.1 release 직전: §11.2 backup/restore drill + §11.3 monitoring dashboard 사용법
- M-v0.2.3 release 직전: §10.3 Pi ingest pipeline 운영 + fallback 절차

**Communication channel** (M-v0.2.0 PoC):
- Slack `#backend-knowledge` (운영 channel)
- email `backend-knowledge-alerts@example.com` (alert routing, M-v0.2.1+)
- GitHub Issues (incident tracking)

**§5.6 cutover 와의 정합**:
- §5.6 의 cutover 절차 중 rollback trigger 발동 시 본 §11.1 의 incident runbook 중 해당 type 의 runbook 즉시 활성화
- on-call operator 가 cutover rollback + incident 대응 동시 수행
- cutover 후 monitoring 지표 5 항목 중 1 이상 임계 초과 시 incident 등록

## 12. Frontend page 상세화 (M-v0.2.1+ 관리/조회 page 1 + viz.html 자가 viewer, 2026-06-18 신규)

**Motivation**: §5.1 M-v0.2.1 DoD 의 "frontend 관리/조회 page 1" + §6.1 Phase 1 의 "M-v0.2.0 만 frontend 0 page, viz.html 자가 viewer 만 SSR" 가 구체적 정공법 없이 high-level 정의만 존재. 본 §12 가 frontend component 구조 / page list / routing / user flow / API integration / 운영 정책 상세 정의.

**독립 backend-knowledge frontend 정합** (§1.2 G7 + §2.3 standalone 정합):
- `backend-knowledge/web/` 별도 standalone frontend (devhub frontend 와 분리)
- frontend 기술 선택 = 운영자 결정 (Next.js / Vue.js / Svelte / vanilla JS — M-v0.2.0 PoC 는 vanilla JS + CDN 권장, M-v0.2.1+ 운영자 결정)
- backend-knowledge 의 API 와 HTTP 통신 (envelope + endpoint, §3.1 / §10.2 정합)
- gateway 가 frontend ↔ backend-knowledge 사이 인증 + user context 처리 (Path Y, §3.6.1 정합)

### 12.1 M-v0.2.0 frontend 0 page + viz.html 자가 viewer 상세

**viz.html 자가 viewer** (M-v0.2.0 PoC = frontend 의 전부):

**위치**: `backend-knowledge/var/bundles/{bundle}/viz.html` (per-bundle, §3.5.4 정합)

**기술 스택** (M-v0.2.0 PoC 권장):
- **Cytoscape.js** v3.x (graph visualization, CDN embed)
- **marked.js** v5.x (Markdown → HTML 변환, CDN embed)
- **HTML/CSS/JS CDN**: jsdelivr / unpkg (외부 CDN, M-v0.2.0 PoC simple)
- **inline style** (외부 CSS 의존 ❌, M-v0.2.0 PoC self-contained)
- **SVG fallback** (CDN 부재 시)

**viz.html component 구조**:

```
<viz.html>
├── <div id="cy-container">          ← Cytoscape.js canvas (full screen)
├── <div id="concept-detail">       ← concept detail panel (right side, hidden default)
├── <div id="bundle-info">          ← bundle metadata (top, sticky)
├── <div id="filter-bar">           ← category + type filter (left side, collapsible)
└── <script src="cytoscape.min.js"> ← CDN embed
    <script src="marked.min.js">    ← CDN embed
    <script>
      // 1. load bundle's index.md (parent dir)
      // 2. parse per-category sections → build Cytoscape nodes
      // 3. parse intra-bundle cross-link → build edges
      // 4. node click → show concept-detail (frontmatter + body)
    </script>
```

**Cytoscape nodes** (concept .md):
- node.id = `{type}_{slug}.md` (filename)
- node.label = `concept.title` (frontmatter)
- node.type = concept.type (8 enum)
- node.category = `concept.x_devhub_category` (5 enum)
- node.curator = `concept.x_devhub_curator`
- node.lifecycle = `concept.x_devhub_status` (M-v0.2.1+ 추가)
- node.style.color = type 별 색상 (e.g., `dataset` = blue, `metric` = green, `runbook` = orange, ...)

**Cytoscape edges** (cross-link, §3.5.5 정합):
- edge.id = `{from_node}__to__{to_node}`
- edge.source = from concept
- edge.target = to concept
- edge.type = `intra-bundle` / `cross-bundle` / `source-external` / `reverse-index`
- edge.style.dashArray = type 별 점선 패턴

**viz.html 의 자가 viewer 특성** (M-v0.2.0 PoC 의 frontend 0 page 정책 정합):
- **외부 backend 호출 ❌**: viz.html 은 정적 HTML 파일 (curl 가능)
- **graph data = index.md 의 frontmatter + body** (parse at load time)
- **CDN 의존성**: M-v0.2.0 PoC = jsdelivr/unpkg CDN 허용 (offline 시 SVG fallback)
- **browser-only**: JavaScript 만 사용 (Node.js / server-side rendering ❌)

**viz.html 자동 생성**: `curate/index_builder.py` 가 per-bundle rebuild 시 자동 emit (M-v0.2.0 PoC, §3.5.4 정합)

### 12.2 M-v0.2.1 frontend 관리/조회 page 1 상세

**`backend-knowledge/web/` 별도 standalone frontend** (M-v0.2.1+ 추가, devhub frontend 와 분리, §5.1 / §5.5 정합):

**5 page 구성** (per §5.5 M-v0.2.1 DoD "frontend 관리/조회 page 1" 의 page list):

| Page | path | 기능 | backend-knowledge API | 권한 |
| --- | --- | --- | --- | --- |
| **(1) Concept list page** | `/` | bundle/category/type filter + full-text search + pagination | GET `/api/v0-2/search?q=...&type=...&category=...&limit=...&offset=...` (§3.1) | visitor + (caller scope filter, §3.6.3 정합) |
| **(2) Concept detail page** | `/concept/{type}/{name}` | frontmatter + body markdown + cross-link 표시 + lifecycle state + lifecycle transition UI (review/publish/archive) | GET `/api/v0-2/concepts/{type}/{name}` (§3.1) + PUT `/api/v0-2/concepts/{id}` (manual edit, §3.6.2 정합) | visitor + operator (edit) + admin (lifecycle state change, §3.9 정합) |
| **(3) Ingest trigger page** | `/ingest` | source 별 POST trigger + status + cron interval | POST `/api/v0-2/ingest/{source}/sync` (§3.1) + GET `/api/v0-2/ingest/{source}/status` (§3.1) | operator (system_admin OR bundle owner_org member) |
| **(4) Bundle management page** | `/bundles` | bundle list + create + rebuild + viz.html preview + index.md download | GET `/api/v0-2/bundles` + POST `/api/v0-2/bundles` + POST `/api/v0-2/bundles/{bundle}/rebuild` (§3.1) | admin (system_admin) |
| **(5) Raw inspector page** | `/db/raw` (M-v0.2.0+ DB path 추가) | raw_records 검색/필터 + sha256 검증 + Pi ingest status | GET `/api/v0-2/db/raw?source=...&since=...` + GET `/api/v0-2/db/raw/{id}` + GET `/api/v0-2/db/raw/ingest-status?source=...&since=...` (§10.2 정합) | operator (storage_mode=db source 한정, §10.5 visibility 정합) |

**Frontend 기술 선택** (운영자 결정, M-v0.2.1+ sprint 진입 시):
- **vanilla JS + CDN** (M-v0.2.0 PoC 권장 — simple, backend-knowledge 의 standalone 정합)
- **Next.js** (React 기반, SSR 가능, M-v0.2.1+ 옵션)
- **Vue.js** (Composition API, M-v0.2.1+ 옵션)
- **Svelte** (compile-time reactive, M-v0.2.1+ 옵션)
- 운영자 결정 (별도 ADR 권장, M-v0.2.1 sprint 진입 시)

**Routing 구조** (vanilla JS 예시):
```
backend-knowledge/web/
├── index.html                  # redirect → /concept (default page)
├── concept/
│   └── index.html              # (1) concept list page (M-v0.2.1+)
├── concept-detail/
│   └── index.html              # (2) concept detail page (M-v0.2.1+)
├── ingest/
│   └── index.html              # (3) ingest trigger page (M-v0.2.1+)
├── bundles/
│   └── index.html              # (4) bundle management page (M-v0.2.1+)
└── db/
    └── raw/
        └── index.html          # (5) raw inspector page (M-v0.2.0+)
```

**Standalone 정합**:
- `backend-knowledge/web/` 는 backend-knowledge repo 내부 (별도 standalone repo ❌)
- frontend build 결과물 (정적 HTML/JS/CSS) 만 deploy (SSR / dynamic rendering ❌, M-v0.2.0 PoC = static)
- devhub frontend 와 코드 공유 ❌ (import ❌, §1.2 G7 standalone 정합)

### 12.3 User flow + 권한 매트릭스 (3 user role)

**3 user role** (M-v0.2.1+ frontend 정책, §3.6 governance 정합):

| Role | 정의 | 권한 범위 | frontend UI |
| --- | --- | --- | --- |
| **visitor** | 모든 인증된 user (Path Y caller-provided user context 정합) | GET only, `x_devhub_visibility = public` 인 concept + bundle 만 | concept list + detail (read only) |
| **operator** | bundle owner_org member + caller user_id 가 owner_user_id 또는 org_head scope | GET + POST `/ingest/{source}/sync` + GET raw inspector (visibility 정합) | visitor + ingest trigger + raw inspector |
| **admin** | system_admin (or org_head scope 의 critical action) | GET + POST + PUT + DELETE + lifecycle state change | operator + bundle management + lifecycle transition UI |

**Path Y caller-provided user context 흐름** (§3.6.1 정합):
```
[User browser] → [gateway (Keycloak 인증 + user context 추출)]
                  ↓ X-DevHub-User-Context header
              [backend-knowledge API]
                  ↓ 4-tier query scope priority (§3.6.3) 적용
              [Filtered response]
                  ↓ JSON envelope (§3.4)
              [frontend rendering]
```

**Frontend ↔ gateway 통신**:
- frontend 는 backend-knowledge 직접 호출 ❌ (gateway 통과 필수, §2.3 "운영자 또는 별도 agent 가 호출")
- gateway 가 Keycloak 인증 + user context 구성 + backend-knowledge 호출 의 3-step orchestration
- frontend → gateway: OIDC code flow + PKCE (Keycloak standard)
- gateway → backend-knowledge: X-DevHub-User-Context header

### 12.4 API integration matrix (frontend ↔ backend-knowledge)

**Per page → API mapping** (M-v0.2.1+ frontend 운영):

| Frontend page | HTTP method | API endpoint | §3.1 / §10.2 cross-ref | Path Y user context |
| --- | --- | --- | --- | --- |
| **(1) Concept list page** | GET | `/api/v0-2/search` + `/api/v0-2/concepts/{type}/{name}` (개별 상세 link) | §3.1 Query | 필수 |
| **(1)** | GET | `/api/v0-2/bundles/{bundle}/index.md` (per-bundle index for viz.html) | §3.1 Query | 필수 (caller scope filter) |
| **(2) Concept detail page** | GET | `/api/v0-2/concepts/{type}/{name}` | §3.1 Query | 필수 (visibility 정합) |
| **(2)** | PUT | `/api/v0-2/concepts/{id}` (manual edit) | §3.6.2 curation permission | 필수 |
| **(2)** | POST | `/api/v0-2/concepts/{id}/enrich` (LLM enrich, M-v0.2.3+) | §3.1 Curate | 필수 |
| **(2)** | POST | `/api/v0-2/concepts/{id}/publish` (system_admin override) | §3.9.4 publish 절차 | 필수 |
| **(2)** | POST | `/api/v0-2/concepts/{id}/archive` (obsolete, M-v0.2.1+) | §3.9.4 archive 절차 | 필수 (operator 권한) |
| **(3) Ingest trigger page** | POST | `/api/v0-2/ingest/{source}/sync` | §3.1 Ingest | 필수 (caller 권한 check) |
| (3) | GET | `/api/v0-2/ingest/{source}/status` | §3.1 Ingest | 필수 |
| **(4) Bundle management page** | GET | `/api/v0-2/bundles` | §3.1 Query | 필수 |
| (4) | POST | `/api/v0-2/bundles` (create) | §3.1 Query | 필수 (admin only) |
| (4) | POST | `/api/v0-2/bundles/{bundle}/rebuild` | §3.1 Query | 필수 (admin only) |
| **(5) Raw inspector page** | GET | `/api/v0-2/db/raw?source=...&since=...&sort=...&limit=...&offset=...` | §10.2 | 필수 (operator, storage_mode=db source 한정) |
| (5) | GET | `/api/v0-2/db/raw/{id}` | §10.2 | 필수 |
| (5) | GET | `/api/v0-2/db/raw/ingest-status?source=...&since=...` | §10.2 | 필수 |

**API path prefix 구분**:
- §3.1 endpoints = `/api/v0-2/{ingest|curate|query|raw|bundles|concepts}` (기존 raw storage = file, §4 정합)
- §10.2 endpoints = `/api/v0-2/db/raw` (db storage, §10 정합)
- §3.6.1 endpoints = `/api/v0-2/{bundles|concepts|raw|db/raw}` 의 caller scope filter 정합

### 12.5 Frontend cutover 정책 (M-v0.2.1+ 운영)

**M-v0.2.0 → M-v0.2.1 frontend cutover** (viz.html → frontend 관리/조회 page 1 추가):

```
Step 1: `backend-knowledge/web/` 디렉터리 작성 (5 page + routing)
Step 2: backend-knowledge API integration (§12.4 매트릭스 정합)
Step 3: Path Y caller-provided user context 구현 (frontend → gateway → backend-knowledge)
Step 4: e2e smoke (frontend 5 page 정상 응답)
Step 5: §5.6 cutover checklist 8 항목 통과
Step 6: frontend deploy (별도 정적 파일 deploy, backend-knowledge 와 독립)
Step 7: 운영 검증 (1주 frontend 5 page 정상 운영)
```

**Frontend update 주기**:
- per release (M-v0.2.1 release, M-v0.2.2 release, ...)
- frontend 는 backend-knowledge 와 별도 deploy 가능 (정적 파일 deploy, no coupling)
- 긴급 patch: §5.6 cutover 절차 + §11.1 incident runbook §11.1.6 (archive trigger 실패) 와 연계

**viz.html 단독 운영 vs frontend 통합 운영**:
- M-v0.2.0 PoC = viz.html 단독 (frontend 0 page)
- M-v0.2.1+ = viz.html + frontend 관리/조회 page 1 양립
  - viz.html = public visibility 의 concept + bundle 만 표시 (visitor 용)
  - frontend page = operator/admin 기능 (ingest / bundle management / raw inspector / lifecycle transition)

**§5.6 cutover 와 정합**:
- frontend cutover 도 §5.6 의 6 step cutover 절차 적용 (이전 viz.html → 새 frontend)
- §5.6 의 8 항목 cutover checklist + frontend-specific item (browser compatibility check, CDN fallback 검증)
- cutover 후 monitoring: §11.3 monitoring 5 지표 + frontend-specific (page load time, JS error rate) 추가

**Cross-section 정합 fix 6 위치** (2026-06-18 신규):
1. **§5.1 M-v0.2.1 DoD** — "frontend 관리/조회 page 1" 의 5 page detail (§12.2 정합) cross-reference
2. **§6.1 Phase 1** — "M-v0.2.0 만 frontend 0 page, viz.html 자가 viewer 만 SSR" 의 viz.html 상세 (§12.1 정합) cross-reference
3. **§3.5.4 index.md 자동 생성** — viz.html 자동 emit 정합 (§12.1 정합)
4. **§3.9.4 publish + archive 절차** — frontend lifecycle transition UI 정합 (§12.2 정합)
5. **§10 DB path raw_records** — frontend raw inspector page 정합 (§12.2 정합)
6. **ADR-0035 §3.6 frontend 정책** — §12 frontend cutover 정책 cross-reference + frontend 독립 frontend 기술 선택 노트

## 13. Cross-cutting 종합 (umbrella doc 전체 cross-reference 정합성 최종 검토 + post-sprint follow-up 종합, 2026-06-18 신규)

**Motivation**: 12 commit 으로 추가된 v0.2.0 umbrella doc (§1~§12 + 12 신규 subsection) + ADR-0034 / ADR-0035 영향 section 의 cross-reference 정합성 최종 검토 + post-sprint follow-up 항목 종합. 본 §13 이 umbrella doc 의 마지막 comprehensive review.

**독립 backend-knowledge 정합**: 본 §13 의 모든 항목이 standalone 정책 (§1.2 G7) + Path Y (§3.6) + backend-ai 폐기 (§6.6.2) 와 정합.

### 13.1 Cross-reference matrix (umbrella doc + ADRs + state.json)

**12 umbrella sections vs cross-cutting artifacts**:

| Section | ADR-0034 (OKF) | ADR-0035 (backend-knowledge) | state.json | external-integrations-agentic-rag-roadmap.md | docs/llm-wiki mirror | 비고 |
| --- | --- | --- | --- | --- | --- | --- |
| §1 v0.2.0 컨셉 | low (motivation) | low (motivation) | low (status) | low (background) | ✅ (mirror scope) | umbrella 전체의 motivation |
| §2 backend-knowledge | low (참조) | **high** (위치/tier/외부 시스템 단방향) | low | low | ✅ | §2.3 standalone 정책 핵심 |
| §3.1 API 매트릭스 | low | high (8 endpoint) | low | low | ✅ | query API + ingest + curate |
| §3.2 type enum (8종) | **high** (ADR-0034 §3.2 정합) | low | low | low | ✅ | OKF concept type 핵심 |
| §3.2.1 5 카테고리 결정 | low (참조) | low | low | low | ✅ | Gitea 통합 1차 wire |
| §3.5 Concept organization | high (5×8 matrix, index.md, cross-link) | low | low | low | ✅ | OKF 적용 핵심 |
| §3.6 Data governance | medium (frontmatter spec) | **high** (Path Y caller-provided user context) | low | low | ✅ | Path Y 핵심 결정 |
| §3.7 Data normalization | high (source plugin ABC) | low | low | low | ✅ | 5 step normalization |
| §3.8 Source plugin 정공법 | high (SourcePlugin ABC + 10 step) | low | low | low | ✅ | source plugin 작성 정공법 |
| §3.9 OKF lifecycle | medium (lifecycle 5 단계) | low | low | low | ✅ | concept lifecycle 운영 |
| §4 raw API | medium (envelope format) | high (raw API 정책) | low | low | ✅ | §4.4~§4.7 운영 정책 |
| §5 마일스톤 | low | medium (6 마일스톤 표) | **high** (M-v0.2.0 row 발급, §13.3 정합) | low | ✅ | M-v0.2.0~v0.3.0 |
| §6 Phase 1/2/3 | low | high (Phase 운영 정공법) | low | low | ✅ | docker-compose + cutover |
| §7 Q&A (18/18) | low | medium (Q12~Q18 ADR 영향) | low | low | ✅ | 결정 완료 |
| §8 timeline | low | medium (Q12~Q18 결정 row) | low | low | ✅ | 2026-06-10~18 결정 |
| §9 변경 이력 | low (12 row) | medium (ADR 영향 row) | low | low | ✅ | commit 기록 |
| §10 DB-based raw + Pi | medium (frontmatter spec) | **high** (file|db dual storage mode + Pi periodic ingest) | low | low | ✅ | DB path + Pi SDK mode 핵심 |
| §11 운영 runbook | low | medium (incident + backup + monitoring) | low | low | ✅ | day-2 운영 정공법 |
| §12 frontend page | medium (viz.html) | high (frontend 관리/조회 page) | low | low | ✅ | M-v0.2.1+ frontend |

**Cross-reference 정합성 검증 결과** (2026-06-18):
- umbrella doc 본문: ✅ 모든 §1~§12 가 §13.1 의 matrix 와 정합
- ADR-0034 / ADR-0035 영향 section: ✅ 모든 신규 section (§3.5~§3.9 / §4.4~§4.7 / §6.5~§6.7 / §10 / §11 / §12) row 추가
- state.json: 📋 post-sprint 항목 (§13.3 정합)
- external-integrations-agentic-rag-roadmap.md: 📋 status active 전환 (Q7 결정, §13.3 정합)
- docs/llm-wiki mirror: ✅ 자동 mirror (per release, `scripts/wiki-mass-ingest.sh --apply`)

### 13.2 미해결 cross-section gap 식별 (2026-06-18)

**Known gaps** (umbrella doc 본문 정의 vs 실제 구현 시점 차이):

| Gap | 정의 (umbrella doc) | 실제 구현 (M-v0.2.0 PoC 진입 시) | 영향 |
| --- | --- | --- | --- |
| **§3.5.5 cross-link reverse index** | `okf/link_graph.py` 가 자동 갱신 (§3.5.5) | M-v0.2.0 PoC = simple in-memory, M-v0.2.1+ = sqlite persisted | mid (perf 영향, M-v0.2.0 PoC 시 OK) |
| **§10.3 Pi prompt template** | 단순 prompt + raw JSON (§10.3 j2 template) | M-v0.2.0 PoC = 단순 prompt, M-v0.2.1+ = 진보된 prompt engineering (few-shot examples + chain-of-thought) | low (PoC 작동, accuracy 는 M-v0.2.1+ 개선) |
| **§11.1 incident runbook 6 type trigger 조건** | RTO + mitigation 정의 (§11.1) | M-v0.2.0 PoC 운영 1주 후 실제 trigger 조건 tuning 필요 (false positive / false negative) | mid (PoC 검증 후 §11.3 alert threshold 재조정) |
| **§5.3 M-v0.2.0 sprint 진입 checklist 6 항목** | 본 §13 commit 시점에 **4/6 완료** (umbrella doc publish / external-integrations-agentic-rag-roadmap.md status / state.json M-v0.2.0 row / OKF SPEC.md 1차 정독), 잔여 **2 항목** (backend-knowledge/ 디렉터리 skeleton / GitHub milestone v0.2.0) | §13.3 후속 결정 항목으로 처리 | high (sprint 진입 전 필수) |
| **§10.3 Pi SDK mode 의 npm dependency** | `@earendil-works/pi-coding-agent` npm pkg (§2.2 / §10.3) | M-v0.2.0 PoC = Node.js 설치 + npm install 필수 | low (1 회 설치, 이후 cache) |
| **§11.2 backup schedule 의 cron 등록** | 매일 02:00 UTC (§11.2) | M-v0.2.0 PoC 운영 환경 cron daemon 설정 필요 (Docker container 내부 cron vs host cron) | mid (PoC 운영 환경 setup) |

### 13.3 후속 결정 항목 (post-sprint follow-up, 6 row)

**M-v0.2.0 sprint 진입 전 반드시 처리**:

| # | 항목 | 위치 | 책임자 | 정합 section |
| --- | --- | --- | --- | --- |
| 1 | **GitHub milestone `v0.2.0` 생성** + 본 문서 link 첨부 | GitHub repo | project lead | §5.3 checklist 6번 |
| 2 | **`ai-workflow/memory/state.json` M-v0.2.0 row 발급** (status: planned → in_progress) | `ai-workflow/memory/state.json` | project lead | §5.3 checklist 3번 |
| 3 | **`external-integrations-agentic-rag-roadmap.md` status active 전환** (Q7 결정, umbrella publish signal) | `docs/planning/external-integrations-agentic-rag-roadmap.md` | project lead | §0.4 + §7 Q7 |
| 4 | **`docs/llm-wiki` mirror scope 갱신** (12 commit content mirror, ~+1900 줄) | `~/wiki/raw/projects/devhub/` | (CI 자동, `scripts/wiki-mass-ingest.sh --apply`) | AGENTS.md §문서 작업 기준 |
| 5 | **M-v0.2.0 release notes draft** (12 commit summary + 18/18 결정 + 5 카테고리 정합) | `docs/release-notes/v0.2.0.md` (M-v0.2.0 release 시) | project lead | §5.5 M-v0.2.0 DoD |
| 6 | **`docs/DOCUMENT_INDEX.md` + `docs/planning/README.md` 갱신** (umbrella doc + ADR-0034/0035 인덱스 추가) | `docs/DOCUMENT_INDEX.md` + `docs/planning/README.md` | project lead | docs governance |

**Post-sprint follow-up workflow**:
1. 항목 1~6 중 sprint 진입 시점에 처리 (M-v0.2.0 release 직전)
2. 항목 4 는 자동 (CI), 5~6 는 project lead 책임
3. 항목 1, 2, 3 은 본 umbrella doc 본 §13 commit 시점에 아직 미완료 (의도적, sprint 진입 trigger)

### 13.4 Cross-cutting 영향 종합 + 정합 검증 결과 (2026-06-18)

**최종 정합 검증 결과**:

| 검증 항목 | 결과 | 비고 |
| --- | --- | --- |
| umbrella doc 본문 (§1~§12) cross-reference 정합성 | ✅ | §13.1 matrix 정합 |
| §7 Q&A 18/18 결정 완료 | ✅ | Q12~Q18 §7 추가 완료 |
| §8 timeline 결정 row 정합 (Q12~Q18) | ✅ | 7 row 추가 완료 |
| §9 변경 이력 12 commit row | ✅ | 각 commit 별 영향 section 정합 |
| ADR-0034 §4.3 영향 (10 row) | ✅ | §3.2.1 / §3.3 / §3.5 / §3.6 / §3.7 / §3.8 / §3.9 / §6.5~§6.7 / §10 / §11 / §12 |
| ADR-0035 §3 영향 (frontend 정책 / Q12~Q18 trade-off / §3.8 마일스톤 표) | ✅ | 4 row 갱신 |
| §11 운영 runbook 6 incident type + 5 monitoring 지표 + 4 role | ✅ | §13.1 matrix 정합 |
| §12 frontend page 5 page + 3 role + 14 row API matrix | ✅ | §13.1 matrix 정합 |
| standalone 정책 (§1.2 G7) 일관성 | ✅ | 모든 section 에서 다른 backend 연결 ❌ 유지 |
| Path Y caller-provided user context (§3.6) 일관성 | ✅ | §3.6.1 endpoint 표 + §12.4 API matrix 정합 |
| DB-based raw (§10) + Pi periodic ingest pipeline 정합 | ✅ | §10.4 storage_mode 분기 + §10.3 Pi SDK mode scheduler |

**📋 미완료 (post-sprint follow-up)** — §13.3 의 6 row (GitHub milestone / state.json / external-integrations-agentic-rag-roadmap.md status / docs/llm-wiki mirror / release notes / DOCUMENT_INDEX.md). 이 항목들은 **M-v0.2.0 sprint 진입 시점에 처리** (umbrella doc 본 §13 commit 시점에서는 의도적 미완료).

**§13.2 의 6 row Known gaps** — M-v0.2.0 PoC 운영 시 자연스럽게 해소 (cross-link reverse index / Pi prompt template / incident runbook tuning / cron daemon setup). umbrella doc 본문 변경 불필요.

**Cross-section 정합 fix** (2026-06-18 신규, 본 §13 자체는 종합 review 이므로 cross-section fix 0 row):
- ADR-0034 §4.3 영향 + frontmatter 갱신 (§13 row 추가)
- ADR-0035 영향 + frontmatter 갱신 (§13 cross-cutting 종합 영향 추가)
- umbrella doc frontmatter 갱신 (§13 cross-cutting 종합 cross-section fix 명시)
- §9 변경 이력 row 추가 (§13)
