# backend-knowledge v0.2.0+ — Entry Point

- 문서 목적: `backend-knowledge` v0.2.0+ standalone backend 의 진입점. architecture + tech-stack + operation 진입 doc.
- 범위: ADR-0035 정합 (Python 3.13+ / FastAPI / OKF / Pi / 완전 standalone). 5종 PoC source plugin (Gitea 4 + homelab_mock).
- 대상 독자: M-v0.2.0+ sprint 진입자, 후속 contributor, 운영자, reviewer.
- 상태: **accepted** (2026-06-19, M-v0.2.0 PoC release post-impl retrospective, retro-design recovery)
- 최종 수정일: 2026-06-19
- 관련 문서: [`architecture.md`](./architecture.md) (main design) / [`tech-stack.md`](./tech-stack.md) / [`docs/planning/release_v0-2_roadmap.md`](../planning/release_v0-2_roadmap.md) / [ADR-0035 backend-knowledge 신설](../adr/0035-backend-knowledge-creation.md) / [ADR-0034 OKF v0.1 채택](../adr/0034-okf-adoption.md)

## 1. 개요

`backend-knowledge` 는 **v0.2.0 신규 백엔드** (umbrella doc §1.2 G7 + ADR-0035 §3):
- **완전 standalone**: 다른 backend (backend-core / 다른 시스템) 의 어떤 layer 든 import / 호출 / 공유 ❌
- **외부 시스템 7종 source 만 단방향** (M-v0.2.0 PoC = Gitea 4 sub-plugin + homelab_mock 5종, M-v0.2.3 운영 기준 = + metrics + hrdb = 7종)
- **Path Y caller-provided user context** (§3.6.1): backend-knowledge 는 auth 자체 안 함, caller 가 `X-DevHub-User-Context` header 로 user/org/project/roles 7 field 검증
- **OKF v0.1 concept format** (ADR-0034): 1 concept = 1 .md + YAML frontmatter (12 field, `type` 1개 필수)
- **봉투 암호화 v2** (per-raw DEK + KEK wrap, Codex P2 review fix)

## 2. 진입 doc

| doc | 내용 |
|---|---|
| [`architecture.md`](./architecture.md) | **main design doc** — 15 section: layer 격리 / component diagram / module dependency / data flow / sequence diagram / storage / API 30 endpoint matrix / design decisions / operation / test strategy / known violations / forward-looking |
| [`tech-stack.md`](./tech-stack.md) | Python backend + SvelteKit frontend runtime + test 의존성 + 선택 근거 + 호환성 매트릭스 (M-v0.2.0~M-v0.3.0+) |
| [`frontend-design.md`](./frontend-design.md) | **SvelteKit frontend design** — 9 section: tech stack / 3 layer (routes/lib/components) / 5 page scope / Path Y dev fixture / API client / build+dev / CI / forward-looking |

## 3. PoC 운영 위치

| 파일 | 설명 |
|---|---|
| `src/backend_knowledge/` | source code (22 file) |
| `var/raw/{source}/{id}.{bin|json}` | 봉투 암호화 v2 raw data (per-source) |
| `var/raw/{source}/{id}.meta.json` | raw metadata sidecar (Codex P1 fix) |
| `var/bundles/{bundle}/{type}/{slug}.md` | OKF v0.1 concept (1 .md = 1 concept) |
| `var/bundles/{bundle}/{type}/{slug}.meta.json` | concept metadata sidecar |
| `var/bundles/{bundle}/.index/reverse_index.json` | reverse in-link index |
| `var/bundles/{bundle}/.index/index.md` | bundle index (per concept summary) |
| `var/bundles/{bundle}/.index/viz.html` | Cytoscape.js v3.30.2 self-contained viewer |
| `var/audit/audit-YYYY-MM-DD.jsonl` | JSON Lines audit log (daily rotation) |
| `docs/operations/runbooks/11.1.{1..6}-*.md` | 6 incident runbook |

## 4. M-v0.2.0 PoC status (2026-06-19 release)

| 항목 | 값 |
|---|---|
| **Implementation PR** | #654 (PR 1, 8 endpoint) / #655 (PR 2, 14 endpoint) / #656 (PR 3, 8 endpoint) — **3 PR MERGED** |
| **Total endpoint** | 30 (8 + 14 + 8) |
| **Total UT** | 166 (PR 1: 65 + PR 2: 50 + PR 3: 51) |
| **Total E2E step** | 11 (`tests/e2e/test_smoke.py`) |
| **Audit event** | 7 (per §3.6.6.1) |
| **Metric** | 18 PoC + 10 stub (M-v0.2.3+/M-v0.3.0+) |
| **Runbook doc** | 6 (§11.1.1~6) |
| **Test** | 166/166 pytest pass |
| **Release PR** | #657 (release notes + traceability IMPL/UT/TC row) |
| **Tag** | `v0.2.0` (annotated, force-push, points to release commit `952005ba`) |
| **Tier** | **사외** (single env, mock + standalone + GitHub main push, 2026-06-19 결정) |

## 5. 후속 작업 (M-v0.2.1+ sprint)

- **§13 known violations 4 row refactor** (architecture.md §13)
  - 13.1 API cross-router call (P1): `api/_common.py` 로 helper 추출
  - 13.2 FastAPI `on_event` deprecation (P2): `lifespan` context manager
  - 13.3 Private helper cross-router (P1): `api/curate.py` private helper public 화
  - 13.4 No lint (P3, optional): ruff + mypy
- **10 stub metric 활성화** (Pi LLM 5 + API versioning 4 + false positive 1)
- **on-call rotation + 4 role 권한** (§11.4)
- **archive/publish status field 실제 변경** (§3.9.4)
- **PostgreSQL option** (§10.1, M-v0.2.3+)
- **Pi (pi.dev) SDK** (§3.5.7, M-v0.2.3+)
- **HMAC signature verification** (§1.1 한계 4, M-v0.2.3+ 검토)

## 6. Cross-reference

| doc | role |
|---|---|
| [`docs/planning/release_v0-2_roadmap.md`](../planning/release_v0-2_roadmap.md) | umbrella doc (18 main section + 80+ subsection, M-v0.2.0 PoC 정합) |
| [ADR-0034 OKF v0.1 채택](../adr/0034-okf-adoption.md) | format 결정 (1 concept = 1 .md, 8 type enum) |
| [ADR-0035 backend-knowledge 신설](../adr/0035-backend-knowledge-creation.md) | 신규 백엔드 정당화 + layer 격리 정공법 |
| [`docs/traceability/report.md`](../traceability/report.md) v0.2.0 PoC row | 7-level chain (REQ → FR → NFR → UC → IMPL → UT → TC) 정합 |
| [`docs/operations/runbooks/`](../operations/runbooks/README.md) | 6 incident runbook + index |
| [`docs/requirements/v0.2.0-*.md`](../requirements/) | §1 도메인 분류 / §2 FR / §3 NFR / §4 UC / §5 TM (PR #653) |

## 7. 변경 이력

| 일자 | 변경 |
|---|---|
| 2026-06-19 | 1차 작성 (M-v0.2.0 PoC release post-impl retrospective, retro-design recovery, PR #657 follow-up). architecture.md + tech-stack.md + README.md 3 file 신규 (3 doc = 1 logical concept). |
