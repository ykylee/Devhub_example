---
type: topic
status: active
last_ingested_from: ai-workflow/wiki/INGEST_GUIDE.md + ai-workflow/wiki/SCHEMA.md §2-§4
related_pages: [concepts/wiki-source-rule-r9, concepts/memory-3-state-lifecycle, patterns/r4-anchor-index, patterns/frozen-archive-immutability, patterns/stale-90day-lint, decisions/adr-004-wiki-layer, decisions/adr-005-r9-wiki-source-rule, INGEST_GUIDE]
created: 2026-06-12
updated: 2026-06-12
r9_skip: true
---

# Wiki Ingest Lifecycle — Raw → Wiki → Lint → Query

## TL;DR

runtime wiki 의 전체 ingest lifecycle 은 **두 갈래**로 운영된다 — (1) `memory/active/` → freeze → `archive/` → wiki 의 **R8+R9 강제 경로** (memory snapshot), (2) `workflow-source/` + `docs/` + `ai-workflow/wiki/` → wiki 의 **R9 면제 경로** (codebase self-describe). 양쪽 모두 [[SCHEMA]] §2 의 ingest workflow + §3 query + §4 lint 사이클로 수렴.

| Path | Source | R9 적용 | Trigger | Owner |
|---|---|---|---|---|
| **Memory snapshot** | `memory/archive/YYYY-MM-DD/` | YES (강제) | session-end freeze → next-session ingest | freeze → wiki-ingest skill |
| **Codebase self-describe** | `workflow-source/`, `docs/`, `ai-workflow/wiki/` | NO (면제) | Phase 1~7 (INGEST_GUIDE) | 수동 curator + 반자동 sync |
| **Release snapshot** | `memory/release/v0.X.Y/` | YES (archive 동급) | release 절차 deep-freeze | release manager |

## Ingest Phases

[[INGEST_GUIDE]] §2 의 Phase 1~7. codebase self-ingest 의 master timeline. 1 phase = 1 `wiki-ingest:` commit (5~15 page 동시 갱신, [[SCHEMA]] §2.1 + [[INGEST_GUIDE]] §6).

| Phase | Action | Output | 자동화 |
|---|---|---|---|
| **P1. 스켈레톤** | L1 wiki schema 확장 + index anchor + INGEST_GUIDE + L3 raw mirror + L2 index placeholder | 4 file + 1 sync script + 1 log entry | 반자동 |
| **P2. Concepts** | 6 concept page (3-layer, R9, memory 3-state, contract v1, agent topology, harness distribution) | 5 new + 1 extend | 수동 |
| **P3. Decisions** | 4 ADR mirror (001-003) + 1 신규 (ADR-005 R9) | 4-5 page | 수동 |
| **P4. Patterns** | 4 pattern (R4 anchor + frozen archive + wiki-stub emit + stale-90day) | 3 new + 1 extend | 수동 |
| **P4b. Entities** | 11 entity (standard_ai_workflow + workflow_kit + workflow_source + ai-workflow-runtime + 6 harness + bootstrap-wiki-py + mcp-read-only-bundle + skill-catalog) | 10 new + 1 extend | 수동 |
| **P5. Topics** | 3 cross-cutting topic (architecture-2026, harness-distribution-model, wiki-ingest-lifecycle) | 3 page | 수동 |
| **P6. L2 외부 wiki** | `~/wiki/wiki/projects/standard-ai-workflow/` 셋업 + `wiki-source-sync` 으로 mass source stub (588 → ~N stub) | 2-3 page + N stub | 반자동 |
| **P7. 검증** | L1 index anchor 매칭 + log append + L2 lint + cross-ref 양방향 | 2 log entries | 수동 |

commit prefix `wiki-ingest:` (예: `wiki-ingest: bootstrap + 2 concept pages`). push 직전 `git fetch && git rebase origin/main` (R3). `.ingest_lock` 으로 동시 ingest 직렬화.

## Memory Snapshot Path (R8+R9)

session state 가 wiki page 가 되기까지의 전 단계. R9 강제 ([[concepts/wiki-source-rule-r9]] + [[decisions/adr-005-r9-wiki-source-rule]]). [[INGEST_GUIDE]] §1 의 "Memory snapshot" 행.

| Step | 위치 | 동작 | Rule / Pattern |
|---|---|---|---|
| 1. session write | `memory/active/` | session 중 hot-write (handoff, backlog, state.json, log.md) | mutable. [[concepts/memory-3-state-lifecycle]] §2.1 |
| 2. session-end freeze | `memory/active/` → `memory/archive/YYYY-MM-DD/` | `memory-freeze` skill 의 COPY (NOT move) + `.frozen` YAML marker | R8 + [[patterns/frozen-archive-immutability]] |
| 3. archive immutability | `memory/archive/YYYY-MM-DD/` | `.frozen` marker = read-only. V-R10 lint 가 5 파일 무결성 검증 (R10) | R10 (post-freeze lint) |
| 4. wiki-ingest source gate | wiki layer | `last_ingested_from` = `archive/...` 만 허용. V-R9 가 `active/` 언급 차단 | R9 + V-R9 ([[concepts/wiki-source-rule-r9]] §Lint Enforcement) |
| 5. multi-page update | `ai-workflow/wiki/{concepts,decisions,patterns,entities,topics}/` | [[SCHEMA]] §2 4-step (read sources → identify pages → update → log append) | R2 + R5 |
| 6. release deep-freeze (선택) | `memory/release/v0.X.Y/` | release 시점 `archive/` 전체 deep-freeze. R9 준거 (archive 동급 immutability) | [[decisions/adr-005-r9-wiki-source-rule]] §Decision |

`active/` 는 step 4 에서 명시 차단 — mutable state 가 immutable wiki 의 source 가 되는 immutability boundary ([[concepts/memory-3-state-lifecycle]] §3 의 mutable → immutable 전이). R8 freeze 의 산출물인 `archive/` 만이 R9 의 자격을 얻는다. step 5 의 4-step 은 [[SCHEMA]] §2 의 master workflow 그대로.

## Codebase Path (R9-exempt)

`standard_ai_workflow` repo 자체를 wiki 로 ingest 하는 별도 경로. R9 scope 밖 ([[concepts/wiki-source-rule-r9]] §Exceptions, [[INGEST_GUIDE]] §1). 면제 사유: 모든 source 가 git-tracked + PR-reviewed 또는 manual curation. `active/` 미사용. ingest 산출물인 wiki page 가 다음 ingest 의 raw source 역할 (loop).

| Step | Phase | 위치 | Action | Output |
|---|---|---|---|---|
| 1. scaffold | P1 | `ai-workflow/wiki/` | L1 schema + L2 index placeholder + L3 raw mirror + INGEST_GUIDE | 4 file + 1 sync script |
| 2. raw mirror | P1 (L3) | `ai-workflow/wiki/_raw/` | `workflow-source/`, `docs/`, `ai-workflow/` snapshot (git-tracked) | full source copy |
| 3. concept | P2 | `concepts/<slug>.md` | 6 concept page (project-architecture, R9, memory 3-state, contract v1, agent topology, harness distribution) | 5 new + 1 extend |
| 4. decision | P3 | `decisions/ADR-NNN-<slug>.md` | 4 ADR mirror (001-003) + 1 신규 (ADR-005) | 4-5 page |
| 5. pattern | P4 | `patterns/<slug>.md` | R4 anchor, frozen archive, wiki-stub emit, stale-90day | 3 new + 1 extend |
| 6. entity | P4b | `entities/<group>/<slug>.md` | 11 entity (1 repo + workflow_kit + workflow_source + ai-workflow-runtime + 6 harness + bootstrap-wiki-py + mcp-read-only-bundle + skill-catalog) | 10 new + 1 extend |
| 7. topic | P5 | `topics/<slug>.md` | 3 cross-cutting topic (architecture-2026, harness-distribution-model, wiki-ingest-lifecycle) | 3 page |
| 8. L2 외부 wiki | P6 | `~/wiki/wiki/projects/standard-ai-workflow/` | `wiki-source-sync` 으로 mass source stub (588 → ~N stub) | 2-3 page + N stub |
| 9. lint + verify | P7 | `ai-workflow/wiki/` | L1 anchor 매칭 + log append + L2 lint + cross-ref 양방향 | 2 log entries |

[[INGEST_GUIDE]] §3 의 page type assignment 가 step 3~7 의 routing 결정. source 별 매핑:

| Codebase source | Page type | 예시 |
|---|---|---|
| `workflow-source/MEMORY_GOVERNANCE.md` §X | `concept` | R9, memory 3-state |
| `docs/architecture/ADR-NNN-*.md` | `decision` | ADR-001~004, ADR-005 |
| `workflow-source/core/*` 패턴 문서 | `pattern` | R4 anchor, frozen archive |
| `workflow-source/scripts/bootstrap_lib/*` | `entity` | bootstrap_lib/wiki.py, mcp.py |
| `workflow-source/harnesses/<harness>/` | `entity` (6개) | codex, opencode, gemini-cli, antigravity, minimax-code, pi-dev |

## Lint & Query Cycle

### 5 Lint Checks ([[SCHEMA]] §4)

| # | Check | Severity | Automation | 근거 |
|---|---|---|---|---|
| 1 | **contradiction** | error | `check_wiki_antipatterns.py` | `[CONTRADICTION]` 태그 미해결 = 0건 유지 (R5) |
| 2 | **stale** | warning | `check_wiki_antipatterns.py` | `updated` 가 90일 이전 (5% 이상) — [[patterns/stale-90day-lint]] |
| 3 | **orphan** | error | `check_wiki_index_structure.py` | inbound link 0 페이지. 신규는 1개+ 강제 |
| 4 | **missing** | error | `check_wiki_index_structure.py` | `related_pages` 참조 vs filesystem 1:1 매핑 |
| 5 | **broken backlinks** | error | `check_wiki_antipatterns.py` | wikilink `[[path]]` 또는 `[text](path)` 의 target 부재 |

추가 lint: **V-R9** ([[concepts/wiki-source-rule-r9]] §Lint Enforcement) — wiki layer 가 `active/` 를 ingest/source 로 언급하면 error (R9 강제). scope: `ai-workflow/wiki/**/*.md` 전체. violation 시 `AssertionError` + exit 1.

### 4-Step Query Workflow ([[SCHEMA]] §3)

| Step | Action | Output |
|---|---|---|
| 1. **index load** | `wiki/index.md` 로드, anchor 구조 카탈로그 확인 ([[patterns/r4-anchor-index]] R4) | 후보 anchor 목록 |
| 2. **identify relevant** | anchor 별 1줄 요약 + `related_pages` + `last_ingested_from` 보고 3~7개 선정 | 관련 페이지 집합 |
| 3. **synthesize** | 선정 페이지 본문 + 표 + 코드 인용 합성. 페이지 citation 필수 | 답변 (user 보고) |
| 4. **file-back decision** | 30줄 초과 또는 durable knowledge → `queries/<date>-<slug>.md` 생성. 1회성 lookup 은 메모리만 | queries/ 페이지 또는 휘발 답변 |

file-back 트리거 3종 ([[SCHEMA]] §3.1): 답변 30줄 초과 / 합성 결과 재사용 가능 / 모순 발견 (양쪽 페이지 `[CONTRADICTION]` 명시). queries/ 페이지는 [[patterns/stale-90day-lint]] 의 lint 제외 대상 (ephemeral).

### Cycle 결합

| Phase | Ingest | Lint | Query |
|---|---|---|---|
| 1 commit 시점 | 5~15 page 갱신 (R2) | n/a | n/a |
| 주 1회 / release 전 | n/a | 5항목 + V-R9 실행 | n/a |
| 사용자 질문 시 | n/a | n/a | 4-step |

## Ingest Cadence

| Path | Cadence | Trigger signal | Latency |
|---|---|---|---|
| Memory snapshot | 매 session | D8 session-end → R8 freeze → 다음 session 시작 시 ingest | ~1 session (실시간 X) |
| Codebase self-describe | 1~2회 / release | Phase 1~7 수동 curate + P6 sync | release-bound |
| Release snapshot | release 시점 | release 절차 deep-freeze → wiki 동시 ingest | release-day |

memory path 의 latency 는 ADR-005 §Consequences Negative #1 의 trade-off. session 진행 중 wiki 검색이 필요하면 handoff 의 `open questions` 섹션 (volatile) 를 우선 source 로 사용 ([[decisions/adr-005-r9-wiki-source-rule]] §Consequences #2 완화책).

## Related

### Concept pages

- [[concepts/wiki-source-rule-r9]] — R9 정의, V-R9 lint, 면제 scope. memory path 의 source gate
- [[concepts/memory-3-state-lifecycle]] — active → archive → release 3-state + R8 freeze mechanism
- [[concepts/project-architecture]] — 3-Layer + LLM Wiki 의 source/raw/wiki 분리 정신 (ADR-001)
- [[concepts/agent-topology]] — Phase P2 의 concept 1 (agent topology concept page)
- [[concepts/harness-distribution]] — Phase P4b 의 6 harness entity 의 routing 기준

### Pattern pages

- [[patterns/r4-anchor-index]] — `index.md` anchor 기반 (R4). query step 1 의 카탈로그 구조
- [[patterns/frozen-archive-immutability]] — R8 freeze 의 immutability 보장 (memory path step 2~3)
- [[patterns/stale-90day-lint]] — lint check #2 의 90일 임계치 (L05)
- [[patterns/wiki-stub-emit]] — Phase P6 외부 wiki mass source stub emit

### Decision pages

- [[decisions/adr-004-wiki-layer]] — wiki layer 도입 (R1~R7, 7 rules). 본 lifecycle 의 상위 spec
- [[decisions/adr-005-r9-wiki-source-rule]] — R9 정식 채택 ADR (proposed, v0.6.1.5)

### Procedure

- [[INGEST_GUIDE]] — codebase self-ingest 의 canonical 절차 (Phase 1~7). 본 topic 의 master source
- [[SCHEMA]] §2 ingest / §3 query / §4 lint — 양 path 의 master workflow 정의. 본 topic 의 canonical spec
