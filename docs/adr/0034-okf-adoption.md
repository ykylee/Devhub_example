# ADR-0034: Open Knowledge Format (OKF) v0.1 채택

## 1. 상태

- **상태**: Accepted
- **작성일**: 2026-06-17
- **수정일**: 2026-06-18 (umbrella doc §3.5~§3.9 + §6.5~§6.7 + §10 + §11 + §12 + §13 + **§1.1 한계 4~7 추가 + §1.3 How 정당화 강화 (한계 7개 → §3~§12 해결책 cross-reference 표) + §3.5.6 cross-link reverse index 정공법 (M-v0.2.0 PoC 능동적 강화, §13.2 known gap 1 ✅ resolved) + §2.4 standalone 검증 매트릭스 (10 row 검증 항목 + 운영자 onboarding SOP + 자동화 tool) + §14 M-v0.2.0 release notes draft (§13.3 #5 ✅ partial resolved) + §8 timeline 보강 (§8.1 17 commit 결정 timeline + §8.2 cross-reference 매트릭스 + §8.3 향후 결정 row 10 row + §8.4 4 layer 정합) + §2.6 backend-knowledge network 정책 (5 subsection + dev/staging/production 3 단계 + docker-compose networks + iptables + WAF 10 rule + 8 row 자동화 tool) + §15 ADR supersession 정공법 (M-v0.2.3+ 부터, 5 step + deprecation policy + release notes 정합)** 신규 + path pattern 정합 fix 에 따른 §4.3 영향 section 갱신 — §3.2.1 5 카테고리 결정 + §3.3 `x_devhub_category` + §3.5 concept organization + §3.6 data governance & query scoping + §3.7 data normalization pipeline + §3.8 source plugin 작성 정공법 + §3.9 OKF concept 운영 lifecycle + §10 DB-based raw + Pi periodic ingest pipeline + §11 운영 runbook (day-2 운영 정공법) + §6.5~§6.7 Phase 1/2/3 운영 정공법 상세 + §12 frontend page 상세화 (M-v0.2.1+) + §13 cross-cutting 종합 (12 commit 후 umbrella doc 전체 cross-reference 정합성 최종 검토 + post-sprint follow-up 6 row) + §1.1 한계 4~7 (Path Y trust model / dual mode 운영 / backup DR / frontend lifecycle) + §1.3 한계 7개 → 해결책 cross-reference 표 7 row + §3.5.6 cross-link reverse index 정공법 (5 subsection + 7 cross-section fix 위치 + §13.2 known gap 1 ✅ resolved) + §2.4 standalone 검증 매트릭스 (10 row 검증 항목 + 운영자 onboarding SOP + 자동화 tool, §1.2 G7 + §3.5 의 standalone 정책의 구체적 검증 정공법) + §14 M-v0.2.0 release notes draft (7 subsection: highlight / 16 commit / breaking change / per-source / per-milestone / §13 정합 / template / contributor, §13.3 #5 ✅ partial resolved) + §8 timeline 보강 (4 subsection: §8.1 17 commit 결정 timeline + §8.2 cross-reference 매트릭스 17 commit × 5 artifacts + §8.3 향후 결정 row 10 row + §8.4 4 layer 정합 L1~L4, 14/17 commit 영향 = 82%) + §2.6 backend-knowledge network 정책 (5 subsection: §2.6.1 3 단계 + §2.6.2 docker-compose networks + §2.6.3 iptables + §2.6.4 WAF 10 rule + §2.6.5 검증 절차 정밀화, §2.4 item 1 + §6.5.3 + §11.1.1 정합, 4 cross-section fix 위치) + **§15 ADR supersession 정공법 (6 subsection: §15.1 정의 + 사용 시나리오 4 종 + §15.2 5 step 정공법 + §15.3 row format + §15.4 cross-reference 4~5 file + §15.5 deprecation policy 12개월 + §15.6 umbrella doc §13~§15 cross-cutting 정공법 3 종 정합, M-v0.2.3+ 부터 supersession 가능, ADR-0035 §6 Supersession section row + 본 ADR-0034 §6 Supersession section 신규 추가)** 20 row 추가 (§3.5.7 Pi LLM cross-link 자동 resolution 정공법 M-v0.2.3+ 부터, 5 subsection + §3.5.6.4 auto-fix strategy 구현 + §13.2 known gap 2 ✅ resolved) + §16 API versioning 정책 (M-v0.3.0+ 부터 v0-3 도입, 6 subsection + deprecation policy 12개월 + dual endpoint + Sunset/Deprecation header + 4 metrics + breaking change 5 종 + 5 cross-section fix 위치) + §3.5.8 Pi LLM resolution false positive rollback 정공법 (M-v0.2.3+ 부터, 4 subsection + 5분 pending undo + 4 channel notification + 5 step recovery workflow, §3.5.7.4 auto-apply safety net + §3.5.7.5 M3 능동적 강화))
- **결정 근거 sprint**: `docs/work_260617-v0-2-umbrella-concept`
- **supersedes**: 없음 (신규)
- **Tier**: 사외 (vendor-neutral 정책)
- **관련 문서**:
  - [release_v0-2_roadmap.md §1.3 / §2.2 / §3.1 / §3.2 / §3.3 / §6.4 / §7 Q11](../planning/release_v0-2_roadmap.md)
  - [Google OKF SPEC.md](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md) (1차 출처)
  - [Google OKF README.md](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/README.md) (1차 출처)
  - [Pi (pi.dev) — github.com/earendil-works/pi](https://github.com/earendil-works/pi) (LLM 호출 시 정합)
  - [external-integrations-agentic-rag-roadmap.md §4](../planning/external-integrations-agentic-rag-roadmap.md) (외부 연동 분리 detail)

## 2. 컨텍스트

### 2.1 문제

`backend-knowledge` 의 knowledge bundle (외부 시스템 데이터 취합) 형식이 필요. 다음 옵션 비교:

| 옵션 | 표준 | vendor-neutral | LLM-friendly | cross-link | 비고 |
| --- | --- | --- | --- | --- | --- |
| A. 자체 schema (JSON / YAML) | ❌ | ❌ | partial | ❌ | 우리만의 schema, 매번 재구성 |
| B. JSON-LD / RDF | partial (W3C) | ✅ | ❌ | ✅ | LLM 친화도 낮음, 복잡 |
| C. **Open Knowledge Format (OKF) v0.1** | ✅ | ✅ | ✅ (Markdown) | ✅ (Markdown link) | Google Cloud 2026-06-12 발표, Apache 2.0, Pi 정합, frontend 정합 |
| D. Obsidian / Notion 형 | ❌ (vendor-specific) | ❌ | ✅ | partial | vendor 종속 |

### 2.2 OKF v0.1 핵심 (1차 출처)

- **형식**: plain Markdown 파일 + YAML frontmatter, 디렉터리 트리
- **1 concept = 1 `.md` 파일** (테이블 / metric / API / runbook / event 등)
- **frontmatter 최소**: `type` 1개 필수. 권장: `title`, `description`, `resource`, `tags`, `timestamp`
- **cross-link**: Markdown `[title](/path/to/concept.md)` 로 concept 간 graph 형성
- **progressive disclosure**: 디렉터리별 `index.md` 자동 생성 (agent / human 이 한 level 씩 navigation)
- **vendor-neutral**: 특정 cloud / DB / agent framework / model provider 종속 ❌
- **producer**: human hand-author / agent (Google ADK, LangChain, custom) / export pipeline (Dataplex, Unity Catalog, Collibra) / DB walking script
- **consumer**: static file server / Obsidian / Notion / MkDocs / LLM context / search index / graph viewer (Cytoscape.js 기반 self-contained `viz.html`)
- **reference impl**: [`enrichment_agent`](https://github.com/GoogleCloudPlatform/knowledge-catalog/tree/main/okf) (Google ADK + Gemini + BigQuery source)
- **라이센스**: Apache 2.0
- **공식 사이트**: https://pi.dev (단, OKF 는 Google Cloud 발표, pi.dev 와 무관)

## 3. 결정

**`backend-knowledge` 의 knowledge bundle 형식으로 OKF v0.1 채택.**

### 3.1 적용 범위

- `backend-knowledge/var/bundles/{bundle}/{type}/{slug}.md` 구조 (1 concept = 1 .md)
- frontmatter: OKF 표준 (type 1개 필수) + `x_devhub_*` prefix 확장 (vendor-neutral 유지)
- cross-link 자동 추출 (`okf/link_graph.py`)
- `index.md` 자동 생성 (per bundle + per type, `curate/index_builder.py`)
- `viz.html` 자가 viewer (Cytoscape.js + marked, 모든 Phase 공통)

### 3.2 type enum (8종, 1차)

| type | 정의 | 예시 |
| --- | --- | --- |
| `dataset` | 외부 DB / table 정의 | `hrdb.persons`, `gitea.repositories` |
| `metric` | 운영 metric 정의 (Prometheus / KPI) | `repo_kpi_sync_duration_seconds` |
| `api_endpoint` | 외부 API endpoint 정의 | `gitea_api_v1_repos_list` |
| `runbook` | 운영 매뉴얼 | `gitea_pull_failure_recovery` |
| `integration` | `backend-knowledge` ↔ 외부 시스템 1쌍 정의 | `homelab_file_puller` |
| `event` | webhook payload 정의 | `gitea_push_event` |
| `reference` | 외부 문서 mirror (1차 raw) | `keycloak_admin_rest_api_v1` |
| `decision` | ADR-style concept (in-bundle ADR) | `decision_2026_06_17_backend_knowledge_creation` |

### 3.3 정책

- **자체 dialect 만들지 않음** (OKF spec 의 "extra keys 자유" 원칙 정합, `x_devhub_*` prefix 로 확장)
- **consumer 가 OKF 만 보면 unknown key 무시 가능**
- **vendor-neutral** (Google ADK, LangChain, custom agent, Obsidian, Notion, MkDocs 모두 정합)

## 4. 결과

### 4.1 positive

- 외부 시스템 데이터의 LLM-friendly 형식 제공 (Markdown + YAML)
- vendor-neutral 로 multi-provider 정합 (Pi, OpenAI, Anthropic, Gemini 모두 정합)
- git 가능 (Markdown 텍스트), cross-link graph 자동 추출
- `viz.html` 자가 viewer (브라우저에서 즉시 조회 가능)
- frontend 별도 개발 불필요 (viz.html 로 1차 viewer 충분)
- Apache 2.0 라이센스로 우리 코드에 통합 부담 없음

### 4.2 negative / trade-off

- 1차 PoC 시 8종 type enum 결정 (추후 custom type 가능하나 1차 scope 8종)
- OKF spec 의 1차 출처 (Google SPEC.md) 와 100% 정합은 sprint 진입 시 1차 정독 필요 (vendor-neutral 정책 + frontmatter 정확한 spec 확인)

### 4.3 영향

- §1.3 본 프로젝트 적용 (OKF 핵심 + 본 프로젝트 매핑)
- §2.2 LLM row (Pi 의 RPC mode / SDK mode 와 정합)
- §2.1 `okf/` (spec.py / frontmatter.py / link_graph.py)
- §3.2 concept type enum (8종)
- §3.2.1 5 카테고리 결정 (이슈 트래커 / 위키 / SCM / CI-CD / 코드 품질, 2026-06-17)
- §3.3 frontmatter spec (`x_devhub_*` prefix + `x_devhub_category` 5 enum)
- §3.4 envelope (자체 정의, OKF 와 cross-reference 만)
- **§3.5 concept organization (5 카테고리 + 8 type orthogonal axes + 5×8 matrix + per-bundle/per-category index.md 3종 자동 생성 + cross-link 4종 rule + bundle 디렉터리 구조 + representative concept frontmatter 예시, 2026-06-18 신규)**
- **§3.6 data governance & query scoping (Path Y caller-provided user context, 2026-06-18 신규) — 5 subsection: §3.6.1 caller-provided user context schema+trust model / §3.6.2 curation governance model (`x_devhub_curator` 별 manual edit permission) / §3.6.3 query scope priority 4-tier (org > personal > project > public) / §3.6.4 frontmatter 5 governance field extension (`x_devhub_owner_org_id` / `_user_id` / `_org_unit_ids` / `_project_ids` / `x_devhub_visibility`) / §3.6.5 cross-section 정합 fix 7 위치**
- **§3.7 data normalization pipeline (category × system → OKF concept, 2026-06-18 신규) — 5 subsection: §3.7.1 5 step normalization 원칙 (raw → concept + 책임 분리) / §3.7.2 per-source type mapping (7 source × types emitted) / §3.7.3 cross-source 동질화 (same type across multiple sources, Jira/Gitea/GitHub 모두 integration_*_issue_puller.md) / §3.7.4 normalize algorithm pseudocode (`sources/{source}.py` 의 normalize() method 4 step) / §3.7.5 edge cases + degraded handling (Partial failure / Schema drift / Source-specific custom transform / Duplicate concept / Large raw / Auth failure)**
- **§3.8 source plugin 작성 정공법 (How to write a source plugin, 2026-06-18 신규) — 5 subsection: §3.8.1 SourcePlugin ABC 인터페이스 (Pydantic v2 + Credential/SourceMeta/Connection/RawResponse/Concept/FetchQuery/HealthStatus 12 type + registry) / §3.8.2 Gitea 4 sub-plugin 정공법 (real wire, M-v0.2.0 PoC 부터) / §3.8.3 homelab_mock 정공법 (filesystem fixture) / §3.8.4 신규 source 추가 10 step 절차 (외부 시스템 API spec 정독 → SourceMeta 정의 → 5 method 구현 → credential schema → body_template → 단위 테스트 → e2e smoke → bundle layout → concept 발췌 → ADR 영향 section) / §3.8.5 source plugin 검증 3 tier (단위 + 통합 + e2e smoke)**
- **§3.9 OKF concept 운영 lifecycle (Created → Reviewed → Published → Active → Archived, 2026-06-18 신규) — 4 subsection: §3.9.1 lifecycle 5 단계 state machine (created/reviewed/published/active/archived + transition + 책임자 + 정합 section) / §3.9.2 frontmatter template per 8 type (dataset/metric/api_endpoint/runbook/integration/event/reference/decision 의 필수 + 권장 field + 예시) / §3.9.3 review checklist 5 항목 (frontmatter validation 7 / body validation 3 / governance validation 2 / bundle validation 3 / cross-link validation 3) / §3.9.4 publish + archive 절차 + 운영 정책 (M-v0.2.0~v0.2.3+ 별 lifecycle 지원 범위 표)**
- **§10 DB-based raw + Pi periodic ingest pipeline (2026-06-18 신규) — 4 subsection: §10.1 DB storage + schema (`raw_records` table + sqlite M-v0.2.0 / PostgreSQL M-v0.2.3+ + 봉투 암호화 ADR-0025) / §10.2 DB CRUD + 데이터 처리 API 8 endpoint (POST/GET/PATCH/DELETE/list/aggregate/search/ingest-status + SQL sort/filter/aggregate/search) / §10.3 Periodic Pi ingest pipeline (SDK mode M-v0.2.0 PoC + 8 step pipeline + Pi prompt template j2 + scheduler cron default `*/5 * * * *` + failure handling) / §10.4 Source path vs DB path 분기 (per source `storage_mode: file|db` + `normalize_mode: rule-based|pi-sdk|pi-rpc` + default mapping: gitea 4 = file/rule-based, homelab = db/pi-sdk, metrics = file, hrdb = db)**
- **§11 운영 runbook (day-2 운영 정공법, 2026-06-18 신규) — 4 subsection: §11.1 Incident 대응 runbook 6 type (source plugin sync 실패 / credential 만료 / Pi ingest timeout-degraded / retention cron 실패 / integrity violation / archive trigger 실패 — per trigger / detection / triage / mitigation / recovery) / §11.2 Backup + restore 절차 (5 backup 대상: DB / var/bundles/ / var/raw/ / .env-KEK / governance field + per storage mode backup 방법 + retention 정책 + restore RTO 5 target + 분기 1회 restore drill) / §11.3 Monitoring + alert routing (5 monitoring 지표: sync 성공률 / Query p95 / integrity violation rate / Pi ingest success / archive trigger + 3 tier alert routing info/warning/critical + alert deduplication 5분) / §11.4 On-call 운영 + role 정의 (4 role: operator / source plugin developer / Pi LLM curator / security auditor + M-v0.2.0 1 person / M-v0.2.1+ 1주 rotation)**
- **§6.5~§6.7 Phase 1/2/3 운영 정공법 상세 (2026-06-18 신규) — 3 subsection: §6.5 Phase 1 (M-v0.2.0+M-v0.2.1) docker-compose standalone 정합 + mock-real wire transition + gateway+firewall+IP allowlist 정책 + 5 step e2e smoke pipeline / §6.6 Phase 2 (M-v0.2.2) 6종 source wire cutover (metrics.py 추가) + backend-ai 폐기 절차 10 단계 + 6 step e2e 6종 smoke + alert routing 검증 / §6.7 Phase 3 (M-v0.2.3) 7종 wire cutover (hrdb.py 추가, db mode + Pi-driven) + Pi 운영 상세 (SDK mode M-v0.2.0~v0.2.2 / RPC mode M-v0.2.3+ option) + LLM enrich + cross-link 자동 resolution 운영**
- **§12 frontend page 상세화 (M-v0.2.1+ 관리/조회 page 1 + viz.html 자가 viewer, 2026-06-18 신규) — 5 subsection: §12.1 M-v0.2.0 viz.html 자가 viewer (Cytoscape.js + marked.js CDN embed + 4 edge type + 8 type node 색상) / §12.2 M-v0.2.1 frontend 관리/조회 page 1 5 page (concept list / concept detail / ingest trigger / bundle management / raw inspector + routing 5 path) / §12.3 User flow + 권한 매트릭스 3 role (visitor/operator/admin) / §12.4 API integration matrix 14 row (per frontend page → backend-knowledge API 1:1 mapping) / §12.5 frontend cutover 정책 (7 step cutover + frontend update 주기 + viz.html 단독 vs frontend 통합 운영 + §5.6 cutover 정합)**
- **§13 cross-cutting 종합 (umbrella doc 전체 cross-reference 정합성 최종 검토 + post-sprint follow-up 종합, 2026-06-18 신규) — 4 subsection: §13.1 cross-reference matrix (12 umbrella sections × ADR-0034 / ADR-0035 / state.json / external-integrations-agentic-rag-roadmap.md / docs/llm-wiki mirror 20 row) / §13.2 미해결 cross-section gap 6 row 식별 (cross-link reverse index / Pi prompt template / incident runbook tuning / M-v0.2.0 sprint 진입 checklist 4/6 + 잔여 2 / Pi SDK mode npm dependency / backup schedule cron 등록) / §13.3 후속 결정 항목 (post-sprint follow-up) 6 row (GitHub milestone v0.2.0 / state.json M-v0.2.0 row / external-integrations-agentic-rag-roadmap.md status active / docs/llm-wiki mirror / M-v0.2.0 release notes / docs/DOCUMENT_INDEX.md + docs/planning/README.md 갱신) / §13.4 Cross-cutting 영향 종합 + 정합 검증 결과 (12 항목 ✅, post-sprint 6 row 📋, known gaps 6 row 자연 해소)**
- **§1.1 한계 4~7 추가 (2026-06-18 결정에서 식별된 4가지 trade-off 한계, 2026-06-18 신규) — 한계 4 caller-provided user context 신뢰 (Path Y 의 trust model, §3.6.1 mitigation) / 한계 5 dual storage mode 운영 복잡도 (§10.4 default mapping + §10.6 운영 가이드 + §11.3 mode 변경 audit) / 한계 6 backup DR transactional 정합성 (§11.1.5 integrity violation + §11.2 backup drill) / 한계 7 frontend standalone 유지보수 부담 (§12.1 viz.html 자가 viewer + §12.4 API integration matrix). §1.3 How 정당화 강화 — 한계 7개 → §3~§12 해결책 cross-reference 표 7 row (한계 1~3 = §1.2 G1~G3 + §6.6 / 한계 4 = §3.6.1 + §11.3 / 한계 5 = §10.4 + §10.6 + §11.3 / 한계 6 = §11.1.5 + §11.2 / 한계 7 = §12.1 + §12.2 + §5.7). 한계 4~7 의 능동적 강화 (HMAC signature / CLI tool / transactional backup / contract test) 는 M-v0.2.1+~M-v0.3.0+ scope 외, §1.1 한계 식별 + §13.2 known gaps 와 정합**
- **§3.5.6 cross-link reverse index 정공법 (M-v0.2.0 PoC 부터 능동적 강화, 2026-06-18 신규 — §13.2 known gap 1 ✅ resolved) — 5 subsection: §3.5.6.1 reverse index 목적 (4 use case: impact 분석 / importance rank / viz.html visualization / archive 거부 정책) / §3.5.6.2 reverse index schema + layout (`var/bundles/.index/reverse_index.json` + schema_version=1 + stats 5 field + per concept inlink list with source/type/section/context + regen timing 4 trigger) / §3.5.6.3 `okf/link_graph.py reverse_index()` implementation (3 step: scan + extract + reverse map + 4 type classification, full pseudocode, full scan default M-v0.2.0 PoC, in-memory dict, atomic write) / §3.5.6.4 stale handling + source-external link 검증 (3 strategy: tolerate/warn/auto-fix per M-v0.2.0/M-v0.2.1+/M-v0.2.3+ + 별도 `external_link_index.json` HTTP HEAD 검증 M-v0.2.1+) / §3.5.6.5 Query API + impact 분석 (3 graph endpoint: GET `/api/v0-2/graph/reverse/{path}` + GET `/api/v0-2/graph/impact/{path}` + POST `/api/v0-2/graph/reindex` + impact analysis JSON 예시 + archive 거부 정책 `inlink_count >= 1` → 409 Conflict + soft archive 권장 + viz.html incoming edge visualization + 4 CLI tool M-v0.2.1+). Cross-section 정합 fix 7 위치: §3.5.5 row 4 reverse index 보강 (§3.5.6 cross-reference 추가) / §2.1 `okf/link_graph.py` 코멘트 갱신 / §3.1 API 매트릭스 row 4 (Graph) + 3 endpoint / §3.9.4 archive 거부 정책 신규 / §6.5.4 E2E smoke Step 6 reverse index PoC 검증 / §11.1.7 stale link runbook 신규 / §13.2 known gap 1 ✅ resolved + §13.4 정합 검증 row 추가**
- **§2.4 standalone 검증 매트릭스 (2026-06-18 신규)** — §1.2 G7 standalone 정책 + ADR-0035 §3.5 의 standalone 정합의 **구체적 검증 정공법**. 10 row 검증 매트릭스 (network 격리 / port expose / env var / import / API 호출 / DB / cron worker / monitoring / log / artifact) + per 항목 검증 방법 + PASS 기준 + FAIL 시 mitigation. OKF 형식 자체는 vendor-neutral 이지만, 본 시스템의 OKF bundle 저장 위치 + `okf/` 디렉터리는 backend-knowledge 내부에 격리 (item 4 import 격리 정합). ADR-0034 §1 의 vendor-neutral 정책 + 본 §2.4 의 item 5 (API 호출 격리 = 외부 시스템 source plugin 호출만 허용) 가 §3.1 API 매트릭스 의 모든 endpoint 의 "내부 호출" 정합 + §6.4 source plugin 작성 정공법 정합. cross-section 정합 fix 3 위치: §1.2 G7 cross-reference (umbrella doc) / §6.5.1 docker-compose standalone 정합 검증 cross-reference (umbrella doc) / §11.4 on-call Operator training cross-reference (umbrella doc). §2.4 매트릭스의 item 4 (import 격리) 가 본 ADR-0034 의 OKF reference (GoogleCloudPlatform/knowledge-catalog) 의 import 정합 + `google-cloud-knowledge-catalog` external dependency 와 정합 (외부 표준 = ✅, 다른 backend Python module = ❌).
- **§14 M-v0.2.0 release notes draft (2026-06-18 신규 — §13.3 #5 ✅ partial resolved)** — umbrella doc 본문 release notes draft (M-v0.2.0 release 시점에 `docs/release-notes/v0.2.0.md` 로 copy + post-process). 7 subsection (§14.1 highlight 7-10 bullet / §14.2 16 commit summary / §14.3 breaking change 4 row: `backend-ai/` 폐기 + Go adapter 흡수 + Tier 분리 정책 + curation governance 정책 / §14.4 per-source plugin 7종 per storage_mode × normalize_mode / §14.5 per-milestone 5 M / §14.6 §13 cross-cutting 정합 + post-sprint follow-up 6 row / §14.7 release notes template per backend-knowledge / §14.8 contributor placeholder). OKF 정합: §14.1 highlight 의 "Google OKF v0.1 채택" 1차 출처 (GoogleCloudPlatform/knowledge-catalog, Apache 2.0) + §14.3 breaking change 의 Tier 분리 정책 + §14.4 per-source plugin 의 bundle 디렉터리 구조 (`{bundle}/{category}/{slug}.md`, §3.5.3 정합) + §14.7 release notes template 의 frontmatter `type` 1개 필수 (8종 type enum 정합, §3.2). cross-section 정합 fix 3 위치: §13.3 #5 partial resolved 갱신 + §13.4 정합 검증 row 추가 (release notes) + umbrella doc frontmatter 갱신.
- **§8 timeline 보강 (2026-06-18 신규 — §13.1 cross-reference matrix + §14.2 16 commit summary 정합)** — §8 의 high-level 결정 19 row (Q1~Q18) → §8.1 17 commit 결정 timeline (per commit 의 concept change 1줄 + 영향 section + cross-reference) → §8.2 cross-reference 매트릭스 (17 commit × 5 artifacts: ADR-0034 14/17, ADR-0035 10/17, state.json 3/17, external-integrations-agentic-rag-roadmap.md 2/17, docs/llm-wiki mirror 17/17 = 100%) → §8.3 향후 결정 row 10 row (Q-N1~Q-N6 sprint 진입 시점 2026-06-19 + Q-F1~Q-F4 후속 sprint 2026-07-01~09-01) → §8.4 결정 timeline 의 4 layer 정합 (L1 high-level 결정 / L2 commit 결정 / L3 cross-reference / L4 향후 결정). 본 §8 보강은 운영자 / contributor 가 §8 의 어느 layer 를 봐도 결정 timeline + 영향 + 향후 결정 row 파악 가능 의 umbrella doc 본문 SoT 역할. OKF 정합: §8.1 의 17 commit 의 영향 section 의 §3.5.3 / §3.6 / §3.7.2 / §3.8 / §3.9 / §6.5~§6.7 / §10 / §11 / §12 / §13 / §1.1 / §1.3 / §3.5.6 / §2.4 / §14 row 의 OKF 형식 (per frontmatter `type` 1개 필수 + 8종 type enum) + bundle 디렉터리 구조 (`{bundle}/{category}/{slug}.md`) 정합. cross-section 정합 fix 3 위치: §13.4 정합 검증 row 추가 (timeline 보강) + §13.1 cross-reference matrix 정합 (12 umbrella sections × 5 artifacts, 본 §8.2 의 commit × 5 artifacts 와 cross-reference) + umbrella doc frontmatter 갱신.
- **§2.6 backend-knowledge 운영 환경의 network 정책 (2026-06-18 신규 — §2.4 item 1 + §6.5.3 + §11.1.1 정합)** — §2.4 매트릭스 item 1 (network 격리) 의 **구체적 정공법**. 5 subsection: §2.6.1 3 단계 network 정책 (dev = localhost + port 8000 / staging = VPN+사내 CA+iptables basic+gateway IP allowlist / production = WAF+외부 CA+iptables strict+gateway+IP allowlist) / §2.6.2 docker-compose.yml networks 설정 정공법 (3 단계 별 YAML 예시: dev = default bridge / staging = internal bridge + egress-internal / production = internal bridge + egress-allowlist + `internal: true` flag) / §2.6.3 firewall iptables rule 예시 (production, INPUT chain SSH+gateway → 8000 ACCEPT, OUTPUT chain source plugin source_url + rate limit, FORWARD default DROP, Docker iptables chain interaction 주의) / §2.6.4 WAF 설정 (Cloudflare / AWS WAF / nginx mod_security 3 option + 10 row WAF rules: R1 Path Y header / R2 HTTP method / R3 SQL injection / R4 XSS / R5 rate limit / R6 request size / R7 IP allowlist / R8 user agent / R9 geolocation / R10 bot detection) / §2.6.5 §2.4 item 1 검증 절차 정밀화 (8 row 자동화 tool `scripts/check_network_isolation.sh` + 운영자 manual SOP + per release audit + incident runbook 정합). OKF 정합: §2.6.4 WAF rule R1 (Path Y header 검증) 가 §3.6.1 caller-provided user context 정합 + ADR-0035 §3.5 운영 환경 standalone 정합 (다른 backend 연결 ❌ = WAF + firewall + IP allowlist = 사외/사내 2-tier). cross-section 정합 fix 4 위치: §2.4 item 1 "상세 정공법" cross-reference + §6.5.3 "상세 정공법" cross-reference + §11.1.1 "Network 진단" 4 row 진단 절차 + §13.4 정합 검증 row 추가 (network 정책).
- §6.4 source plugin 작성 (외부 시스템 API spec 만 참조, OKF 형 concept emit)
- **§3.5.7 Pi LLM cross-link 자동 resolution 정공법 (M-v0.2.3+ 부터, 2026-06-18 신규 — §3.5.6.4 auto-fix strategy 구현 + §13.2 known gap 2 ✅ resolved)** — §3.5.5 cross-link 4종 rule 의 unresolved link 자동 recommend + operator confirm + 3 mode confirm workflow (dry-run/confirm/auto-apply) 의 구현 정공법. 5 subsection: §3.5.7.1 목적 (unresolved link 자동 recommend + operator confirm + §3.5.6.4 auto-fix strategy 구현) / §3.5.7.2 j2 prompt template design (input unresolved link context ±2 lines + output 3 row recommendation + reason + confidence 0~1) / §3.5.7.3 SDK/RPC mode 선택 §10.3 정합 (M-v0.2.3+ default SDK mode + production RPC mode option) / §3.5.7.4 3 mode confirm workflow (dry-run/confirm/auto-apply ≥ 0.9) + `POST /api/v0-2/concepts/{id}/resolve-links?mode={dry-run|confirm|auto-apply}&selected_rank={1|2|3}&confidence_threshold=0.9` endpoint / §3.5.7.5 audit log + 5 metrics (MTTR < 30분 / accuracy ≥ 70% / false positive ≤ 5% / pi_sdk_timeout ≤ 1% / pi_llm_recommendation_count 일 ≤ 50) + `cli/fix_unresolved.py` 4 CLI tool. OKF 정합: §3.5.7.2 의 prompt template 의 input/output schema 가 OKF `type: reference` type enum + bundle 디렉터리 구조 (`{bundle}/{category}/{slug}.md`, §3.5.3 정합) + `x_devhub_*` prefix (confidence score field) 정합. cross-section 정합 fix 5 위치: §3.5.6.4 auto-fix row + §3.1 API 매트릭스 endpoint + §6.7.3 LLM enrich 운영 + §10.3 Pi prompt template row 갱신 + §13.2 known gap 2 ✅ resolved (1/6 → 2/6 resolved, residual 4/6).
- **§16 API versioning 정책 (M-v0.3.0+ 부터 v0-3 도입, 2026-06-18 신규 — §3.1 API 매트릭스 향후 호환성 + §14.7 release notes template breaking change + §15.5 deprecation policy 정합)** — `/api/v0-2/` prefix 의 의도 + M-v0.3.0+ 부터 `/api/v0-3/` prefix 도입 시 마이그레이션. 6 subsection: §16.1 API versioning 정의 (URL prefix 기반 semver + v0.x pre-1.0 + 12개월 deprecation) / §16.2 deprecation policy 12개월 + dual endpoint support (M-v0.3.0 release 시 /api/v0-2/ + /api/v0-3/ 동시 운영 + 6개월 warning + 12개월 제거 + client migration SOP 5 step) / §16.3 API gateway deprecation header (Sunset RFC 8594 + Deprecation + Link successor-version) / §16.4 monitoring 2개 버전 동시 운영 4 metrics (per endpoint request count + error rate + client identification + migration progress) + §11.3 monitoring 5 + 4 = 9 metrics / §16.5 breaking change 정의 5 종 (a) path 변경 / (b) method 변경 / (c) schema 변경 / (d) auth 변경 / (e) default 변경 + release notes 정합 §14.7 + §14.3 / §16.6 §3.1 API 매트릭스 versioning 영향 + future deprecation timing (M-v0.2.0~v0.3.0 / M-v0.3.0+ deprecation / M-v0.3.0+ 제거) + 운영 runbook 영향 (§11.1 incident + §11.3 monitoring + §11.4 on-call role API curator 6번째). OKF 정합: §16.5 의 breaking change 정의 5 종 (a)~(e) 가 OKF `type` field 변경 ❌ (path/method/schema 변경 = OKF bundle 디렉터리 구조 변경 + cross-link 변경, M-v0.3.0+ 정합) + §16.6 §3.1 API 매트릭스 의 모든 endpoint 의 version lifecycle 표 14 row + Graph endpoints 4 row 정합. cross-section 정합 fix 5 위치: §13.4 정합 검증 row 추가 (API versioning) / §3.1 API 매트릭스 future deprecation timing / §14.7 release notes template breaking change / §15.5 deprecation policy 12개월 1:1 정합 / §11.3 monitoring 9 metrics (5+4).
- **§3.5.8 Pi LLM resolution false positive rollback 정공법 (M-v0.2.3+ 부터, 2026-06-18 신규 — §3.5.7.4 auto-apply safety net + §3.5.7.5 M3 능동적 강화)** — §3.5.7.5 의 M3 (false positive rate ≤ 5%) 의 **능동적 강화** + §3.5.7.4 auto-apply 의 **safety net** 으로 M-v0.2.3+ 부터 활성화. 4 subsection: §3.5.8.1 false positive 정의 + 4 종 시나리오 (typo 매칭 / renamed target / self-reference / cycle) + 발생 빈도 target ≤ 5% (M-v0.2.3+ PoC 예상 ~3.6%, production ~5.6% estimated) / §3.5.8.2 rollback trigger 3 종 (operator manual undo `POST /api/v0-2/concepts/{id}/resolve-links?undo=true&applied_at=...` / 5분 내 impact analysis detection / 24시간 monitoring flag) + 5분 pending undo 상태 4 종 (`x_devhub_status`: `applied_pending` / `applied` / `rolled_back` / `rolled_back_late`) + audit log `pi_link_resolve.rolled_back` event + impact analysis snapshot / §3.5.8.3 operator notification 4 channel (Slack `#backend-knowledge` / email `backend-knowledge-alerts@example.com` / dashboard banner 10분간 / §11.3 monitoring alert) + 4 stage timing (T+0분 banner / T+5분 확정 / T+1시간 email / T+24시간 flag) + alert routing M3 threshold (≤ 5% ✅ / > 5% info / > 10% warning / > 20% critical + auto-apply 일시 정지) / §3.5.8.4 recovery workflow 5 step ≤ 5분 (`cli/revert_unresolved.py {concept_path} {applied_at}` + `POST /api/v0-2/graph/reindex` + `POST /api/v0-2/bundles/{bundle}/rebuild` + viz.html + audit log) + §3.5.6.4 stale handling 3 strategy 정합 (tolerate / warn / auto-fix safety net). OKF 정합: §3.5.8.2 의 audit log 의 `pi_link_resolve.rolled_back` event 가 OKF 형식 (event / timestamp / operator_id / impact analysis snapshot) + §3.5.8.4 recovery workflow 의 5 step 이 §3.5.6.2 reverse index + §3.5.6.4 archive 정책 정합. cross-section 정합 fix 4 위치: §3.5.7.4 auto-apply row + "상세 정공법: §3.5.8 false positive rollback 정공법" cross-reference / §3.5.7.5 M3 row + "상세 정공법: §3.5.8 false positive rollback 정공법" cross-reference / §13.4 정합 검증 row 추가 (false positive rollback) / §11.3 monitoring 10 metrics (5 + §3.5.7.5 5) + §12 frontend page admin dashboard banner + ADR-0035 §3.5 운영 환경 standalone 정합.

## 5. 후속 작업

- M-v0.2.0 sprint 진입 시 OKF SPEC.md 1차 정독 (vendor-neutral 정책 + frontmatter 정확한 spec 확인) — release_v0-2_roadmap.md §5.3 checklist 5
- v0.3.0+ 에서 OKF spec 변경 시 ADR 갱신
- `curate/enricher.py` 의 rule-based enricher 구현 (M-v0.2.0) + Pi LLM enrich (M-v0.2.3+)

## 6. Supersession / 변경 이력 (2026-06-18)

| 일자 | 변경 | 사유 |
| --- | --- | --- |
| 2026-06-17 | Initial ADR (status: Accepted) | 사용자 결정 + Google Cloud OKF v0.1 발표 (1차 출처: `GoogleCloudPlatform/knowledge-catalog/okf/SPEC.md`, Apache 2.0) |
| 2026-06-18 | §4.3 영향 18 row 추가 (umbrella doc cross-section 정합) | 17 commit 의 §3.2.1 / §3.3 / §3.5 / §3.6 / §3.7 / §3.8 / §3.9 / §6.5~§6.7 / §10 / §11 / §12 / §13 / §1.1 / §1.3 / §3.5.6 / §2.4 / §14 / §8 영향 row 추가 |

**M-v0.2.3+ 부터 supersession 가능** (2026-06-18 신규 결정, §15 ADR supersession 정공법 정합): 본 ADR-0034 가 후속 ADR (e.g., ADR-0036 OKF v0.2 채택 / OKF v0.1 의 type enum 8종 → 12종 / vendor-neutral 정책 변경) 에 의해 supersede 될 경우, [`release_v0-2_roadmap.md` §15](../planning/release_v0-2_roadmap.md) 의 5 step 정공법 (New ADR 작성 → 본 ADR frontmatter `superseded-by` 추가 → 본 ADR §6 row 추가 → cross-reference 4~5 file 갱신 → state.json `adrs` field 갱신) + 12개월 deprecation policy + release notes 정합. supersession 발생 시 본 §6 의 table 에 supersession 결정 row 가 추가됨 (형식: `| YYYY-MM-DD | **superseded-by ADR-NNNN** | [사유 1-2 문장] |`).

