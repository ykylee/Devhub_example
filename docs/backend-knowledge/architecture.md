# backend-knowledge v0.2.0+ — Architecture Design

- 문서 목적: v0.2.0 PoC `backend-knowledge` standalone backend 의 main design doc. layer 격리 / component diagram / module dependency / data flow / sequence diagram / storage layout / API surface / design decisions / known technical debt 정공법.
- 범위: §3 layer (API / Sources / Storage / OKF / Auth / Audit / Monitoring) + §4 module dependency + §5 data flow 3 종 + §6 sequence 3 종 + §7 storage + §8 API surface (30 endpoint) + §9 design decisions 8 종 + §10 operation + §11 test strategy + §12 forward-looking (M-v0.2.1+ scope) + §13 known violations (technical debt, refactor target).
- 대상 독자: M-v0.2.0+ sprint 진입자, PR reviewer, 운영자, 신규 contributor.
- 상태: **accepted** (2026-06-19, M-v0.2.0 PoC release post-impl retrospective, retro-design recovery)
- 최종 수정일: 2026-06-19
- 관련 문서: [`tech-stack.md`](./tech-stack.md) / [`README.md`](./README.md) / [`docs/planning/release_v0-2_roadmap.md`](../planning/release_v0-2_roadmap.md) / [ADR-0034 OKF](../adr/0034-okf-adoption.md) / [ADR-0035 backend-knowledge](../adr/0035-backend-knowledge-creation.md) / [`docs/traceability/report.md`](../traceability/report.md) v0.2.0 PoC row

## 1. 개요

### 1.1 컨셉
- **§1.2 G7 standalone 정공법** (umbrella doc): backend-knowledge 는 **완전 standalone backend**. 다른 backend (`backend-core` / 다른 시스템) 연결 ❌, OIDC ❌, **외부 시스템 7종 source 만 단방향** (M-v0.2.0 PoC = Gitea 4 sub-plugin + homelab_mock 5종)
- **§3.6.1 Path Y caller-provided user context**: backend-knowledge 는 auth 자체 안 함, caller (gateway / 별도 agent) 가 `X-DevHub-User-Context` header 로 user/org/project/roles 7 field 전달 시 format 검증 (JSON parse + schema check + 만료 5분) + filter/curation ownership check 만 수행
- **ADR-0035 §3** (신설 정당화): 기존 `backend-core/internal/integrations/adapters/` + `infrastructure/` 의 외부 연동 코드를 신규 백엔드로 이전. source plugin **7종** (M-v0.2.3 운영 기준) 의 외부 시스템 API 만 참조, 기존 Go adapter 0 line 참조

### 1.2 위치
- **layer**: 7 module group (API / Sources / Storage / OKF / Auth / Audit / Monitoring) + 2 utility (Config / Logger)
- **module count**: 22 file in `src/backend_knowledge/` (excluding `__init__.py`)
- **endpoint count**: 30 (8 + 14 + 8, PR 1+2+3 MERGED)

### 1.3 Tier
- **사외** (2026-06-19 결정, umbrella doc §1.1 한계 4~7 의 M-v0.2.0 PoC trade-off)
- 사외 한정 정보 0 row (DEVHUB_KEYCLOAK_* / GITEA_URL / HR_EXPORT_CMD / internal-registry.example.com / kc.internal.example.com / devhub.example.com / 172.16.0.0/12 등 pattern 0)
- 사내 한정 경로 변경 0 (`infrastructure/` / `infra/idp/` / `scripts/setup-keycloak.sh` / `docker-compose.{local,test,deploy,colima}.yml` 0 row)
- 사내 env var 추가 0 (`.env.deploy` / `.env.test` / `frontend/.env.example` 0 row)

## 2. 기술 스택 요약

자세한 내용: [`tech-stack.md`](./tech-stack.md).

- **Python 3.13+** / **FastAPI 0.115.6** / **Pydantic 2.9.2** / **uvicorn 0.32.1** / **structlog 24.4.0** / **cryptography 43.0.0** / **httpx** (transitive) / **PyYAML** (OKF frontmatter)
- **test**: pytest 8.3.3 / pytest-asyncio / anyio / FastAPI TestClient

## 3. Layer 격리 (Layer Isolation Rules)

### 3.1 7 layer 정공법

| Layer | Module | Import 가능 | Import 금지 |
|---|---|---|---|
| **API** (FastAPI routers) | `api/{ingest, curate, query, graph, lifecycle, audit, monitoring, health}.py` | storage / sources / okf / auth / audit / monitoring / config / logger | **다른 API router (cross-router call ❌)** |
| **Sources** (plugin) | `sources/{_base, _gitea_base, gitea_repo_pull, gitea_issue, gitea_wiki, gitea_action, homelab_mock, registry}.py` | okf / config / logger / storage (read) | 다른 source plugin (cross-source ❌) |
| **Storage** | `storage/{raw_store}.py` | config / logger / okf (frontmatter 만) | 다른 backend module (단독) |
| **OKF** (format library) | `okf/{frontmatter, concept, cross_link}.py` | config / logger | (독립 library, 다른 backend-knowledge module 호출 ❌) |
| **Auth (Path Y)** | `auth/path_y.py` | config / logger | (단독, 다른 module 호출 ❌) |
| **Audit** | `audit/{events, logger, api}.py` | logger | (다른 layer 호출 ❌, **단 logger 만**) |
| **Monitoring** | `monitoring/{metrics, prometheus, api}.py` | audit / config / logger | (다른 layer 호출 ❌, **단 audit/config/logger 만**) |
| **Config / Logger** (utility) | `config.py` / `logger.py` | (utility 끼리 호출만) | (다른 layer 호출 ❌) |

### 3.2 격리 정당화
- **API cross-router call ❌**: router 는 HTTP endpoint → 다른 router 호출 시 cycle 가능 + test 격리 약화. 대신 **공통 helper** 는 `api/_common.py` 또는 별도 `bundle_store.py` 등의 utility module 로 추출.
- **Source cross-source ❌**: 5 plugin 의 `SourcePlugin` ABC + `registry` 가 dispatch, 개별 plugin 끼리는 무관.
- **OKF 독립 ❌**: `okf/` 는 cross-link / frontmatter / concept 만, `frontmatter.py` 만 storage 에서 import 허용 (frontmatter model 단독 사용 시).
- **Audit cross-layer ❌**: audit log 자체는 Logger 만 의존. 다른 layer 호출 시 audit log 의 순수성 (raw I/O 만) 위배.

## 4. Component diagram

```
┌──────────────────────────────────────────────────────────────────┐
│ HTTP Client (gateway / external agent / developer)               │
│   ↓ X-DevHub-User-Context header (Path Y)                        │
└──────────────────────────────────────────────────────────────────┘
                            ↓
┌──────────────────────────────────────────────────────────────────┐
│ FastAPI app (main.py)                                            │
│   8 router: ingest / curate / query / graph /                   │
│            lifecycle / audit / monitoring / health              │
│   30 endpoint registered via app.include_router()                │
└──────────────────────────────────────────────────────────────────┘
       ↓                ↓                  ↓                ↓
┌──────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ Sources  │  │   Storage    │  │  Audit log   │  │  Monitoring  │
│ 5 plugin │  │ raw_store    │  │ JSON Lines   │  │ 18 metric    │
│ Gitea ×4 │  │ AES-256-GCM  │  │ 7 event      │  │ Prometheus   │
│ + mock   │  │ v2 envelope  │  │ 4 viewer     │  │ 3-tier alert │
└─────┬────┘  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘
      │             │                │                │
      ↓             ↓                ↓                ↓
┌──────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ OKF      │  │ var/         │  │ var/audit/   │  │ structlog    │
│ frontmat │  │ raw/         │  │ audit-YYYY-  │  │ JSON line    │
│ concept  │  │ {source}/    │  │ MM-DD.jsonl  │  │ + Prometheus │
│ cross_lk │  │ {id}.bin|json│  │ (rotation)   │  │ exposition   │
└──────────┘  │ + .meta.json │  └──────────────┘  └──────────────┘
              └──────────────┘
                    ↓
┌──────────────────────────────────────────────────────────────────┐
│ Config (pydantic-settings, 11 env var) + Logger (structlog)      │
│   - PATH_Y_MAX_AGE_SECONDS / VAR_DIR / GITEA_URL / RAW_ENCRYPTION_KEY
│   - AUDIT_LOG_RETENTION_DAYS / ENABLE_METRICS / LOG_LEVEL        │
└──────────────────────────────────────────────────────────────────┘
                            ↓
┌──────────────────────────────────────────────────────────────────┐
│ Auth (Path Y caller-provided user context)                       │
│   - 8 field schema + 5분 만료 + 4 scope filter                 │
│   - get_path_y_context / require_path_y_context (FastAPI dep)   │
└──────────────────────────────────────────────────────────────────┘
                            ↑
┌──────────────────────────────────────────────────────────────────┐
│ Frontend (backend-knowledge/web/, SvelteKit 2 + Svelte 5 + TS 5)│
│   - 5 page (§12.1~§12.5): dashboard / concepts / ingest /       │
│     bundles / raw + audit                                        │
│   - Path Y dev fixture (사외 standalone, gateway 없이)         │
│   - native fetch + src/lib/api.ts wrapper                       │
│   - 사외 (devhub frontend 와 완전 분리)                          │
│   - 자세한 design: [frontend-design.md](./frontend-design.md)   │
└──────────────────────────────────────────────────────────────────┘
```

## 5. Module dependency graph

```
config.py ← logger.py ← (모든 module)
                          ↓
                ┌─────────┴─────────┐
                ↓                   ↓
        auth/path_y.py        okf/frontmatter.py
                ↓                   ↓ (개별 module)
                ↓              okf/concept.py
                ↓              okf/cross_link.py
                ↓                   ↓
                ↓           storage/raw_store.py
                ↓                   ↓
                ↓              api/ingest.py
                ↓                   ↓
                ↓            sources/_base.py ← sources/registry.py
                ↓                   ↓
                ↓         sources/{gitea_*, homelab_mock}.py
                ↓                   ↓
                ↓         sources/_gitea_base.py
                ↓                   ↓
                └─────────→ api/ingest.py ←────┐
                                                ↓
                            api/{curate, query, graph, lifecycle, audit, monitoring}.py
                                                ↓
                                            main.py (FastAPI app)
```

**violation (refactor target)**: 현재 구현에서 `api/{query, graph, lifecycle, audit, monitoring}.py` 가 `api/ingest.py` 에서 `get_path_y_context` / `make_envelope` / `require_path_y_context` 직접 import. **§3.1 API cross-router call ❌** 위배. **§13 known violations (technical debt)** 참조.

## 6. Data flow (3 가지)

### 6.1 Ingest (raw → 봉투 암호화 v2 저장)

```
[POST /api/v0-2/ingest/homelab_mock/sync]
   ↓
api/ingest.py::post_sync(ctx: PathY optional)
   ↓ source not in list_sources → 400
   ↓ plugin = get_source(source)
   ↓ plugin.connect({})  ← 4 Gitea + homelab_mock
   ↓ plugin.fetch(since)  ← in-memory mock or httpx async
   ↓
loop raw in raws:
   ↓ concept = plugin.normalize(raw)
   ↓ raw_store.save(source, type_, name, body, registered_by, owner_org_id, ...)
       ↓ validate source name ^[a-z0-9][a-z0-9_-]{0,62}$
       ↓ sha256(body) → raw_id
       ↓ DEK random 32-byte + KEK wrap + body encrypt (envelope v2)
       ↓ .bin write + .meta.json sidecar (type/name/owner/visibility/frontmatter_override)
       ↓ audit.raw.received event (PoC: 미구현, M-v0.2.3+ scope)
   ↓
return IngestSyncData(synced, failed, raw_ids, ...)
```

### 6.2 Curate (raw → concept + bundle rebuild)

```
[POST /api/v0-2/bundles/{bundle}/rebuild]
   ↓
api/curate.py::rebuild_bundle(body, bundle, ctx optional)
   ↓ bundle_dir not exists → 404
   ↓ full_scan=True (default) → bundle_dir.glob("*/*.md")
   ↓
loop md_file in md_files:
   ↓ text = md_file.read_text()
   ↓ frontmatter, body_md = parse_frontmatter(text)  ← okf/
   ↓ cross_links = extract_cross_links(body_md, base_path=md_file)  ← okf/
   ↓ for cl in cross_links:
       ↓ target_id = f"{bundle}/{type}/{cl.target.split('/')[-1]}"
       ↓ reverse_index[target_id].append({source_concept, type, section, context})
   ↓
write var/bundles/{bundle}/.index/reverse_index.json
write var/bundles/{bundle}/.index/index.md (per concept: title + desc + path + in-link count)
generate var/bundles/{bundle}/.index/viz.html (Cytoscape.js v3.30.2 self-contained)
   ↓ audit.curation.edit event (PoC: 미구현, PR 2 스텁)
   ↓
return BundleRebuildData(concept_count, link_count, ...)
```

### 6.3 Query (concept get / search / viz.html)

```
[GET /api/v0-2/concepts/{type}/{name}?bundle=kb]
   ↓
api/query.py::get_concept(type, name, bundle, ctx required)
   ↓
meta = _load_concept_metadata(bundle, type, name)  ← .meta.json sidecar
   ↓ meta not exists → 404
   ↓ visibility check (personal ≠ caller.user_id → 403)
   ↓ md_path = var/bundles/{bundle}/{type}/{name}.md
   ↓ text = md_path.read_text()
   ↓ frontmatter, body = parse_frontmatter(text)
   ↓ audit.concept.access event (with concept_id, bundle, type, visibility)
   ↓
return ConceptGetData(concept_id, type, name, bundle, frontmatter, body, ...)

[GET /api/v0-2/bundles/{bundle}/viz.html]
   ↓
api/query.py::get_bundle_viz(bundle, ctx optional)
   ↓ viz_path = var/bundles/{bundle}/.index/viz.html
   ↓ viz_path not exists → 404
   ↓
return HTMLResponse(content=viz_path.read_text(), media_type="text/html")
```

## 7. Sequence diagram (3 종)

### 7.1 raw 등록 (POST /raw)

```
Client                FastAPI             raw_store           Audit
  │  POST /raw         │                    │                   │
  │  X-DevHub-User-    │                    │                   │
  │   Context: ...     │                    │                   │
  ├───────────────────>│                    │                   │
  │                    │ get_path_y_context │                   │
  │                    │ (Path Y 검증)      │                   │
  │                    ├────────────────────┼──────────────────>│
  │                    │                    │   audit.user.login│
  │                    │                    │   (success)        │
  │                    │<───────────────────┼───────────────────┤
  │                    │                    │                   │
  │                    │ raw_store.save     │                   │
  │                    │ (source, type,     │                   │
  │                    │  name, body,       │                   │
  │                    │  registered_by,    │                   │
  │                    │  owner_org_id,     │                   │
  │                    │  owner_project_ids)│                   │
  │                    ├───────────────────>│                   │
  │                    │                    │ _validate_source_ │
  │                    │                    │  name (whitelist) │
  │                    │                    │ sha256(body)      │
  │                    │                    │ DEK random 32B    │
  │                    │                    │ envelope v2:      │
  │                    │                    │  [v2][kek_n][wd48] │
  │                    │                    │  [dek_n][ct+tag]  │
  │                    │                    │ write .bin        │
  │                    │                    │ write .meta.json  │
  │                    │<───────────────────┤                   │
  │                    │                    │                   │
  │  201 + raw_id      │                    │                   │
  │<───────────────────┤                    │                   │
```

### 7.2 Concept manual edit (PUT /concepts/{id})

```
Client              FastAPI              Audit            (storage)
  │  PUT /concepts/   │                    │                   │
  │   {id} (Path Y)   │                    │                   │
  ├──────────────────>│                    │                   │
  │                   │ require_path_y_    │                   │
  │                   │  context (400 if  │                   │
  │                   │  missing)          │                   │
  │                   │ body XOR append_   │                   │
  │                   │  body (400)        │                   │
  │                   ├────────────────────┼──────────────────>│
  │                   │                    │ audit.curation.   │
  │                   │                    │  edit (concept_id,│
  │                   │                    │  old=1, new=2,    │
  │                   │                    │  cross_links,    │
  │                   │                    │  commit_message)  │
  │                   │<───────────────────┼───────────────────┤
  │  200 + version=2  │                    │                   │
  │<──────────────────┤                    │                   │
  │  (PoC: stub, PR 3  │                    │                   │
  │   real logic in   │                    │                   │
  │   M-v0.2.3+)       │                    │                   │
```

### 7.3 Audit log viewer (GET /audit?from=...&to=...)

```
Client             FastAPI          Audit logger        Audit file
  │  GET /audit     │                    │                   │
  │  X-DevHub-User- │                    │                   │
  │   Context       │                    │                   │
  ├────────────────>│                    │                   │
  │                 │ get_path_y_context │                   │
  │                 │   (audit.user.     │                   │
  │                 │    login success)  │                   │
  │                 │ get_audit_logger() │                   │
  │                 ├───────────────────>│                   │
  │                 │                    │ read_range(       │
  │                 │                    │  from_date,       │
  │                 │                    │  to_date,         │
  │                 │                    │  event_type,      │
  │                 │                    │  user_id,         │
  │                 │                    │  limit)           │
  │                 │                    │ glob audit-*.jsonl│
  │                 │                    │ parse JSONL       │
  │                 │                    │ filter            │
  │                 │                    │ sort by timestamp │
  │                 │                    │ desc, take limit  │
  │                 │                    ├──────────────────>│
  │                 │                    │<──────────────────┤
  │                 │                    │ return list[dict] │
  │                 │<───────────────────┤                   │
  │  200 + items[]  │                    │                   │
  │<────────────────┤                    │                   │
```

## 8. Storage layout

```
var/
├── raw/                          # 봉투 암호화 v2 + .meta.json
│   ├── homelab_mock/
│   │   ├── abc1234-def5678.bin    # envelope v2: [v2][kek_n12][wd48][dek_n12][ct+tag]
│   │   ├── abc1234-def5678.json   # plaintext mode (KEK 미설정)
│   │   ├── abc1234-def5678.meta.json  # sidecar (Codex P1 fix)
│   │   └── ...
│   ├── gitea_repo_pull/
│   └── ...
├── bundles/                      # OKF v0.1 concept + viz.html
│   ├── devhub-gitea/
│   │   ├── dataset/
│   │   │   ├── users.md          # frontmatter (12 field) + body
│   │   │   ├── users.meta.json
│   │   │   └── ...
│   │   ├── metric/
│   │   ├── api_endpoint/
│   │   └── .index/
│   │       ├── reverse_index.json  # {target_concept: [inlink, ...]}
│   │       ├── index.md            # per concept: title + desc + path + in-link count
│   │       └── viz.html            # Cytoscape.js v3.30.2 self-contained
│   └── devhub-homelab/
├── audit/                         # JSON Lines, daily rotation
│   ├── audit-2026-06-19.jsonl
│   ├── audit-2026-06-18.jsonl
│   └── ...
└── log/                           # application log (M-v0.2.3+ scope)
```

**file mode only** (M-v0.2.0 PoC). M-v0.2.3+ production 시 PostgreSQL option (`§10.1`): `raw_records` table 14 field + `bundle_index` table + `audit_log` table.

## 9. API surface (30 endpoint matrix)

| Module | count | Endpoint | FR | Path Y | 봉투 암호화 | Audit event |
|---|---|---|---|---|---|---|
| **ingest** | 6 | POST /ingest/{source}/sync | FR-I-001 | 권장 | (raw emit) | audit.raw.received (M-v0.2.3+) |
| | | GET /ingest/{source}/status | FR-I-002 | 권장 | - | - |
| | | POST /raw | FR-I-003 | 필수 | ✅ (per-raw DEK) | - |
| | | GET /raw/{type}/{name} | FR-I-004 | 필수 | - | audit.raw.access (M-v0.2.3+) |
| | | GET /raw?source=&since= | FR-I-005 | 필수 | - | - |
| | | DELETE /raw/{id} | FR-I-006 | 필수 | - | audit.raw.deleted (PoC: log 만) |
| **curate** | 5 | POST /concepts/{id:path}/enrich | FR-C-001 | 권장 | - | audit.curation.edit (action=enrich) |
| | | PUT /concepts/{id:path} | FR-C-002 | 필수 | - | audit.curation.edit (old→new) |
| | | GET /bundles | FR-C-003 | 권장 | - | - |
| | | POST /bundles | FR-C-004 | 필수 | - | audit.bundle.create (PoC: 미구현) |
| | | POST /bundles/{bundle}/rebuild | FR-C-005 | 권장 | - | audit.bundle.rebuild (PoC: 미구현) |
| **query** | 5 | POST /query | FR-Q-001 | 필수 | - | audit.query |
| | | GET /concepts/{type}/{name} | FR-Q-002 | 필수 | - | audit.concept.access |
| | | GET /search?q= | FR-Q-003 | 필수 | - | - |
| | | GET /bundles/{bundle}/index.md | FR-Q-004 | 권장 | - | - |
| | | GET /bundles/{bundle}/viz.html | FR-Q-005 | 권장 | - | - |
| **graph** | 4 | GET /graph/reverse/{concept_path:path} | FR-G-001 | 권장 | - | - |
| | | GET /graph/impact/{concept_path:path} | FR-G-002 | 권장 | - | - |
| | | POST /graph/reindex | FR-G-003 | 권장 | - | - |
| | | POST /concepts/{id:path}/resolve-links | FR-G-004 | 필수 | - | audit.pi_link_resolve.* (M-v0.2.3+) |
| **lifecycle** | 2 | POST /concepts/{id:path}/archive | §3.9.4 | 필수 | - | audit.concept.archive |
| | | POST /concepts/{id:path}/publish | §3.9.4 | 필수 | - | audit.concept.publish |
| **audit** | 4 | GET /audit | §3.6.6.3 | 권장 | - | - |
| | | GET /audit/concept/{concept_path:path} | §3.6.6.3 | 권장 | - | - |
| | | GET /audit/user/{user_id} | §3.6.6.3 | self / system_admin | - | - |
| | | GET /audit/org/{org_id} | §3.6.6.3 | same_org / system_admin | - | - |
| **monitoring** | 2 | GET /metrics (Prometheus v0.0.4) | §11.3 | (none) | - | - |
| | | GET /monitoring/alerts | §11.3 | 권장 | - | - |
| **health** | 2 | GET /health | §3.1 | (none) | - | - |
| | | GET /health/protected | §3.6.1 | 필수 | - | audit.user.login (success) |

**총 30 endpoint** (PR 1: 8 + PR 2: 14 + PR 3: 8). FR 1:1 정합.

## 10. Design decisions / trade-offs (8 종)

### 10.1 봉투 암호화 v2 envelope format
- **v1 (1 commit)**: `[version 1][nonce 12][ciphertext + tag 16]` — KEK 직접 사용 (Codex P2 review 위배)
- **v2 (2 commit, `f28a2973` review fix)**: `[version 2][kek_nonce 12][wrapped_dek 48][dek_nonce 12][ciphertext + tag 16+]` — per-raw DEK random + KEK wrap
- **선택**: DEK per raw 로 per-record compartmentalization + KEK rotation 시 DEK 만 re-wrap 가능
- **트레이드오프**: 73 byte overhead (v1: 13 byte) vs per-raw key isolation

### 10.2 Path Y caller-provided user context
- **선택**: backend-knowledge 가 auth 자체 안 함, caller 가 `X-DevHub-User-Context` header (base64url(json)) 로 user/org/project/roles 7 field 전달
- **trade-off**: gateway 신뢰 가정 → §1.1 한계 4 (M-v0.2.3+ HMAC signature)
- **format 검증**: JSON parse + 8 field schema (extra=forbid) + issued_at 만료 5분 (`PATH_Y_MAX_AGE_SECONDS=300`)

### 10.3 Source name whitelist (path traversal defense)
- **선택**: `^[a-z0-9][a-z0-9_-]{0,62}$` regex whitelist
- **reject**: `../tmp`, `/etc`, `..`, `''`, uppercase, spaces, slashes, backslashes, leading non-alphanumeric
- **적용 위치**: `RawStore.save/load/exists/list_source/delete` (5 method) + API layer translate to 400 E_VALIDATION
- **Codex P2 review fix** (commit `f28a2973`)

### 10.4 OKF v0.1 format
- **1 concept = 1 .md** + YAML frontmatter (12 field, `type` 1개 필수 + 8종 enum) + body (Markdown)
- **선택**: git-pushable + wiki review flow + viz.html 자가 viewer 가능
- **Pydantic v2 model**: `ConceptFrontmatter` (extra=allow, populate_by_name, str_strip_whitespace)
- **cross-link 4 type**: explicit / implicit / tag / wikilink (per §3.5.6)

### 10.5 Audit log JSON Lines + daily rotation
- **format**: `{"event": "audit.user.login", "timestamp": "...", "user_id": "...", "success": true, ...}\n`
- **storage**: `var/audit/audit-YYYY-MM-DD.jsonl` (per day, thread-safe Lock)
- **retention**: 7일 (configurable via `AUDIT_LOG_RETENTION_DAYS`)
- **read_range**: filter by event_type / user_id / from_date / to_date / limit, sort by timestamp desc

### 10.6 3-tier alert (info / warning / critical)
- **threshold 기반**: `_calc_sync_success_rate` < 99% warn / < 95% critical, `_calc_query_p95_latency_ms` > 500ms warn / > 1s critical
- **severity label**: Prometheus exposition 에 `severity="ok|warning|critical"` per metric
- **escalation**: info = Slack 일 digest, warning = Slack + on-call 1시간, critical = Slack + page 15분

### 10.7 5 source plugin (4 Gitea + homelab_mock)
- **Gitea 1 instance** (4 sub-plugin): `gitea_repo_pull` / `gitea_issue` / `gitea_wiki` / `gitea_action` (httpx.AsyncClient + Bearer token)
- **homelab_mock** (PoC): in-memory 3 sample (dataset / metric / runbook)
- **mock mode**: `GITEA_URL`/`GITEA_TOKEN` 미설정 시 자동 fallback → in-memory

### 10.8 E2E smoke approach (FastAPI TestClient)
- **선택**: pytest + `fastapi.testclient.TestClient` (in-process, deterministic, no network)
- **11 step**: health → sync → raw → bundle → enrich → rebuild → get → search → graph → lifecycle → audit/metric
- **trade-off**: real Gitea instance 통합 test 는 M-v0.2.1+ scope (§3.8.5 3 tier 검증의 unit + e2e)

## 11. Operation (audit + monitoring + alert + runbook)

### 11.1 Audit log 7 event
- `audit.user.login` / `audit.concept.access` / `audit.curation.edit` / `audit.query`
- `audit.concept.archive` / `audit.concept.publish` / `audit.config.change`

### 11.2 Monitoring 18 PoC metric
- **5 base**: `bk_sync_success_rate_homelab_mock_24h` / `bk_query_p95_latency_ms_1h` / `bk_integrity_violation_rate_24h` / `bk_pi_ingest_success_rate_1h` (stub) / `bk_archive_trigger_failures_24h` (stub)
- **13 governance**: 4 per user (active_logins / curation_count / query_count / access_count) + 4 per org + 4 per project + 1 per event type
- **10 stub** (M-v0.2.3+/M-v0.3.0+): Pi LLM 5 + API versioning 4 + false positive 1

### 11.3 Alert 3-tier
- `info` / `warning` / `critical` severity
- Slack channel routing (M-v0.2.1+)
- on-call page (PagerDuty / Opsgenie, M-v0.2.1+)
- 5분 deduplication window

### 11.4 6 runbook doc
- `docs/operations/runbooks/11.1.1-source-plugin-sync-failure.md`
- `docs/operations/runbooks/11.1.2-credential-expired.md`
- `docs/operations/runbooks/11.1.3-pi-ingest-timeout.md`
- `docs/operations/runbooks/11.1.4-retention-cron-failure.md`
- `docs/operations/runbooks/11.1.5-integrity-violation.md`
- `docs/operations/runbooks/11.1.6-archive-trigger-failure.md`
- `docs/operations/runbooks/README.md` (index)
- Each: Trigger / Detection / Triage / Mitigation / Recovery / MTTR target / Related

### 11.5 E2E 11 step
- `tests/e2e/test_smoke.py` — health / sync / raw / bundle / enrich / rebuild / get / search / graph / lifecycle / audit+metric

## 12. Test strategy

### 12.1 Unit test (10 file, 166 test)
- `tests/test_path_y.py` (8) — 8 field schema + format + 만료
- `tests/test_health.py` (5) — public + protected + 4 ctx 검증
- `tests/test_source_plugins.py` (13) — 5 source + registry + mock mode + normalize
- `tests/test_ingest.py` (15) — 6 Ingest endpoint + DELETE auth (Codex P1 fix)
- `tests/test_storage.py` (16) — plaintext + encrypted + per-raw DEK + source name whitelist (Codex P2 fix)
- `tests/test_curate.py` (15) — 5 Curate endpoint + rebuild + viz.html
- `tests/test_query.py` (17) — 5 Query endpoint + pagination + visibility
- `tests/test_graph.py` (15) — 4 Graph endpoint + orphan + reindex
- `tests/test_audit.py` (18) — 7 event enum + JSONL writer + 4 viewer endpoint
- `tests/test_monitoring.py` (16) — metrics collection + alert evaluation + Prometheus

### 12.2 E2E test (1 file, 11 step)
- `tests/e2e/test_smoke.py` — 11 step happy path (§11.5)

### 12.3 Coverage
- 166/166 pytest pass
- 4 warning (FastAPI `on_event` deprecation — M-v0.2.3+ 에서 lifespan 으로 전환)
- 0 lint (ruff 미설정, M-v0.2.1+ 도입 검토)

## 13. Known violations (technical debt, refactor target)

### 13.1 API cross-router call (refactor P1) — ✅ **resolved (commit `45039918`)**
- **이전 상태**: 7 file (`api/{curate, query, graph, lifecycle, health}.py` + `audit/api.py` + `monitoring/api.py`) 가 `api/ingest.py` 에서 `get_path_y_context` / `make_envelope` / `require_path_y_context` / `EnvelopMeta` / `Envelope` 직접 import
- **violation**: §3.1 API cross-router call ❌ 규칙 위배
- **refactor 정공법 (적용)**: 공통 helper 를 `api/_common.py` 로 추출 + ingest.py 에서 import 제거
- **변경**:
  - `api/_common.py` (NEW, 105 line) — 5 helper + audit log emission
  - `api/ingest.py` (MOD) — 5 helper 정의 본문 제거 (~100 line)
  - 6 caller file (MOD) — `from ._common import ...` 또는 `from ..api._common import ...`
- **expected effort**: 1 commit, 9 file 변경, 0 test 변경
- **테스트**: 166/166 pass

### 13.2 FastAPI `on_event` deprecation (refactor P2) — ✅ **resolved (commit `a58ed29c`)**
- **이전 상태**: `main.py` 의 `@app.on_event("startup")` / `@app.on_event("shutdown")` 4 DeprecationWarning
- **violation**: FastAPI 0.115+ deprecation
- **refactor 정공법 (적용)**: `@asynccontextmanager` `lifespan` context manager 로 전환
- **변경**:
  - `from contextlib import asynccontextmanager` 추가
  - `@asynccontextmanager async def lifespan(app): ...` 정의
  - `app = FastAPI(..., lifespan=lifespan)`
  - 2 `@app.on_event` decorator 본문 제거
- **expected effort**: 1 file 변경, 0 test 변경
- **테스트**: 166/166 pass + 0 DeprecationWarning (이전 4 warning 제거)

### 13.3 Private helper cross-router (refactor P1) — ✅ **resolved (commit `87d9006e`)**
- **이전 상태**: `api/query.py` + `api/graph.py` 가 `api/curate.py` 의 private helper 6 종 (`_bundle_dir` / `_bundle_index_dir` / `_concept_meta_path` / `_load_concept_metadata` / `_find_concept_by_id` / `_build_concept_id`) 직접 import
- **violation**: §3.1 API cross-router call ❌ 규칙 위배
- **refactor 정공법 (적용)**: `api/_bundle_store.py` 에 8 public helper (underscore prefix 제거) 추출
- **변경**:
  - `api/_bundle_store.py` (NEW, 130 line) — 8 public helper (bundle_dir / bundle_index_dir / bundle_meta_path / concept_meta_path / save_concept_metadata / load_concept_metadata / find_concept_by_id / build_concept_id)
  - `api/curate.py` (MOD) — 8 helper 정의 본문 제거 (~100 line) + 8 import + 7 internal call site rename + 4 local var name collision 회피 (bundle_path)
  - `api/query.py` (MOD) — `from .curate import _xxx` → `from ._bundle_store import xxx` + 3 internal call site rename
  - `api/graph.py` (MOD) — `from .curate import _xxx` → `from ._bundle_store import xxx` + 2 internal call site rename
- **expected effort**: 1 commit, 4 file 변경, 0 test 변경
- **테스트**: 166/166 pass

### 13.4 No lint (refactor P3, optional) — ⏳ **deferred (M-v0.2.1+ scope)**
- **현재 상태**: ruff / mypy / black 미설정
- **M-v0.2.1+ 도입 검토**: `pyproject.toml` 에 `[tool.ruff]` + `[tool.mypy]` 추가 + CI pre-merge

## 14. Forward-looking (M-v0.2.1+ scope)

| Milestone | scope | 영향 |
|---|---|---|
| **M-v0.2.1** | 1차 완성 + Gitea 정식 wire + 5 page frontend + e2e smoke | 4 Gitea sub-plugin real Gitea instance + frontend standalone (`backend-knowledge/web/`, M-v0.2.1 §12) |
| **M-v0.2.2** | 외부 시스템 6종 source wire (Gitea 4 + homelab + metrics) + `backend-ai/` 폐기 | metrics source plugin + `backend-ai/` 디렉터리 + Dockerfile + docs 일괄 정리 — ✅ **done 2026-06-22** (PR #663 + #664) |
| **M-v0.2.3** | 외부 시스템 7종 source wire (+ hrdb) + Pi LLM enrich 활성화 + PostgreSQL option + cross-link 자동 resolution | hrdb source + Pi (pi.dev) SDK + DB mode (sqlite/PostgreSQL) + §3.5.7 auto-apply |
| **M-v0.3.0+** | multi-vendor LLM + 풀 RAG + transactional backup + CI contract test + HMAC signature | 임베딩 vendor + chunking/embedding/retrieval/reranking + §1.1 한계 4~7 능동적 강화 |

**M-v0.2.0 PoC 정합**: 5 source plugin (Gitea 4 + homelab_mock) + 30 endpoint + 166 UT + 11 E2E step + 7 audit event + 18 PoC metric + 6 runbook doc. **한계 4~7 (Path Y trust / dual mode / backup DR / frontend lifecycle)** 의 M-v0.2.0 = 사외 환경 mock 처리 (사용자 2026-06-19 결정).

## 15. 변경 이력

| 일자 | 변경 |
|---|---|
| 2026-06-19 | 1차 작성 (M-v0.2.0 PoC release post-impl retrospective, retro-design recovery, PR #657 follow-up). §3 layer 격리 7 layer + §4 component + §5 module dependency + §6 data flow 3 종 + §7 sequence 3 종 + §8 storage + §9 API 30 matrix + §10 design 8 종 + §11 operation + §12 test strategy + §13 known violation 4 row (technical debt, refactor target) + §14 forward-looking. |
