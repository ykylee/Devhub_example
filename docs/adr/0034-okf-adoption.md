# ADR-0034: Open Knowledge Format (OKF) v0.1 채택

## 1. 상태

- **상태**: Accepted
- **작성일**: 2026-06-17
- **수정일**: 2026-06-18 (umbrella doc §3.5~§3.9 + §10 + §11 신규 + path pattern 정합 fix 에 따른 §4.3 영향 section 갱신 — §3.2.1 5 카테고리 결정 + §3.3 `x_devhub_category` + §3.5 concept organization + §3.6 data governance & query scoping + §3.7 data normalization pipeline + §3.8 source plugin 작성 정공법 + §3.9 OKF concept 운영 lifecycle + §10 DB-based raw + Pi periodic ingest pipeline + **§11 운영 runbook (day-2 운영 정공법)** 9 row 추가)
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
- §6.4 source plugin 작성 (외부 시스템 API spec 만 참조, OKF 형 concept emit)

## 5. 후속 작업

- M-v0.2.0 sprint 진입 시 OKF SPEC.md 1차 정독 (vendor-neutral 정책 + frontmatter 정확한 spec 확인) — release_v0-2_roadmap.md §5.3 checklist 5
- v0.3.0+ 에서 OKF spec 변경 시 ADR 갱신
- `curate/enricher.py` 의 rule-based enricher 구현 (M-v0.2.0) + Pi LLM enrich (M-v0.2.3+)
