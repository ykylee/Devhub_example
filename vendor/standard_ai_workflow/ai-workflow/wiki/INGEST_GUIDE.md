---
type: concept
status: active
last_ingested_from: .omo/plans/llm-wiki-convergence-design.md
related_pages: [concepts/project-architecture, decisions/adr-004-wiki-layer, patterns/r4-anchor-index, entities/bootstrap-wiki-py]
created: 2026-06-12
updated: 2026-06-12
r9_skip: true
backfilled_at: 2026-06-12
backfill_note: |
  v0.6.3 P7 검증 중 INGEST_GUIDE frontmatter 의 last_ingested_from 이
  .omo/plans/v0.6.1-plus-codebase-ingest-design.md 였으나 해당 plan 파일은
  소스/raw mirror 어느 쪽에도 존재하지 않음. 가장 의미적으로 가까운
  master plan 인 .omo/plans/llm-wiki-convergence-design.md 로 정정.
  (INGEST_GUIDE 본문 §7 의 "다음에 읽을 문서" 도 동일 plan 을 가리킴)
---

# INGEST_GUIDE — Codebase Ingest into Runtime Wiki (Phase 1-7)

## TL;DR

본 가이드는 **`standard_ai_workflow` repo 자체**를 `ai-workflow/wiki/` runtime layer 에 ingest 하는 절차. 메모리 snapshot 의 wiki-ingest 와는 별개 (R9 rule 미적용 — codebase 자체가 SSOT).

| Layer | Ingest source | 적용 R9? | 자동화 |
|---|---|---|---|
| **Memory snapshot** | `archive/YYYY-MM-DD/` | YES (R9 강제) | `wiki-ingest` skill |
| **Codebase self-description** | `workflow-source/`, `docs/`, `ai-workflow/` | NO (R9 면제) | 본 가이드 (Phase 1-7 수동) |

## §1 왜 R9 가 본 ingest 에는 적용되지 않는가  {#s1-r9-scope}

R9 (v0.6.1.5) 정의: **"wiki-ingest 는 `archive/` 만 source 로 사용"**. 적용 대상:
- `active/` 의 mutable state → `archive/` 로 freeze → wiki-ingest
- 목적: mutable raw 가 wiki 의 immutable source 가 되는 것을 방지 (immutability boundary)

본 codebase ingest 는 R9 의 의도와 무관:
- `workflow-source/`, `docs/`, `ai-workflow/wiki/` 모두 git-tracked + PR-reviewed (또는 manual curation)
- `active/` 미사용 — session state 가 wiki 로 직행하지 않음
- ingest 산출물인 wiki page 가 *새로운* raw source 역할을 수행 (다음 ingest 의 input)

→ **결론**: 본 ingest 는 R9 scope 외. 별도 phase 로 운영.

## §2 Ingest Phase 계획 (Phase 1-7)  {#s2-phases}

| Phase | 작업 | 산출물 | 자동화 |
|---|---|---|---|
| **P1. 스켈레톤** | L1 wiki schema 확장 + index anchor + INGEST_GUIDE + L3 raw mirror + L2 index placeholder | 4 file + 1 sync script + 1 log entry | 반자동 |
| **P2. Concepts** | 6 concept page (3-layer, R9, memory 3-state, contract v1, agent topology, harness distribution) | 5 new + 1 extend | 수동 |
| **P3. Decisions** | 4 ADR mirror (001-003) + 1 신규 (ADR-005 R9) | 4-5 page | 수동 |
| **P4. Patterns** | 4 pattern (R4 anchor + frozen archive + wiki-stub emit + stale-90day) | 3 new + 1 extend | 수동 |
| **P4b. Entities** | 11 entity (standard_ai_workflow + workflow_kit + workflow_source + ai-workflow-runtime + 6 harness + bootstrap-wiki-py + mcp-read-only-bundle + skill-catalog) | 10 new + 1 extend | 수동 |
| **P5. Topics** | 3 cross-cutting topic (architecture-2026, harness-distribution-model, wiki-ingest-lifecycle) | 3 page | 수동 |
| **P6. L2 외부 wiki** | `~/wiki/wiki/projects/standard-ai-workflow/` 셋업 + `wiki-source-sync` 으로 mass source stub (588 → ~N stub) | 2-3 page + N stub | 반자동 |
| **P7. 검증** | L1 index anchor 매칭 + log append + L2 lint + cross-ref 양방향 | 2 log entries | 수동 |

## §3 Page Type Assignment  {#s3-page-types}

codebase content 를 L1 wiki 의 5 page type 중 어디에 배치할지:

| Codebase source | Page type | 예시 |
|---|---|---|
| `workflow-source/MEMORY_GOVERNANCE.md` §X | `concept` | R9, memory 3-state |
| `docs/architecture/ADR-NNN-*.md` | `decision` | ADR-001~004, ADR-005 (R9) |
| `workflow-source/core/*` 패턴 문서 | `pattern` | R4 anchor, frozen archive, harness overlay |
| `workflow-source/scripts/bootstrap_lib/*` | `entity` | bootstrap_lib/wiki.py, mcp.py |
| `workflow-source/skills/<name>/SKILL.md` | `entity` (11개) | session-start, backlog-update, ... |
| `workflow-source/mcp_servers/<name>/SKILL.md` | `entity` (8+개) | latest-backlog, check-doc-links, ... |
| `workflow-source/harnesses/<harness>/` | `entity` (6개) | codex, opencode, gemini-cli, antigravity, minimax-code, pi-dev |
| `docs/architecture/README.md`, INSTALLATION | `topic` | architecture-2026, distribution model |
| `releases/Beta-v0.X.Y.md` | `topic` | ingest lifecycle, distribution model |

## §4 Body Template  {#s4-body-template}

기존 page 가 따르는 4-section 구조 (project-architecture.md, r4-anchor-index.md 참조):

```markdown
## TL;DR
(1 단락 또는 요약 table)

## <Section 1>
(상세 내용, table 위주)

## <Section 2>
(상세 내용)

## Related Decisions
또는 ## Related (page-type 별)
(cross-ref wikilinks)

## References
(외부/내부 raw path 들)
```

한국어 prose + 영어 technical term 혼용. Code/paths 는 원문.

## §5 Frontmatter  {#s5-frontmatter}

SCHEMA.md §1.1 따른 5 type. 필수 공통: `type`, `status`, `last_ingested_from`, `related_pages`, `created`, `updated`.

추천 frontmatter 보강:
- `last_ingested_from`: ingest source path (workflow-source/, docs/, ...)
- `confidence`: high | medium | low (codebase 직접 확인 vs 추론)

## §6 Ingest Commit 규칙  {#s6-commit}

SCHEMA.md §2.1 적용:
- prefix: `wiki-ingest:`
- 1 commit = 1 phase (5~15 page 동시 갱신)
- push 직전 `git fetch && git rebase origin/main` (R3)
- `.ingest_lock` 직렬화

## §7 다음에 읽을 문서  {#s7-next}

- [§1 5종 페이지 타입](./SCHEMA.md#s1-page-types) — frontmatter 스키마
- [§2 Ingest Workflow](./SCHEMA.md#s2-ingest) — multi-page 갱신 절차
- [§4 Lint Checklist](./SCHEMA.md#s4-lint) — 5항목 검사
- [`./index.md`](./index.md) — 페이지 카탈로그
- [`./log.md`](./log.md) — ingest/query/lint 이벤트
- 상위 plan: [`../.omo/plans/llm-wiki-convergence-design.md`](../.omo/plans/llm-wiki-convergence-design.md)
- R9 원본: [`../workflow-source/MEMORY_GOVERNANCE.md`](../workflow-source/MEMORY_GOVERNANCE.md) §4
