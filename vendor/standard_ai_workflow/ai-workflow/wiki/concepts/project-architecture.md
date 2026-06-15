---
type: concept
status: active
last_ingested_from: docs/architecture/ADR-001 + ADR-004 + .omo/plans/v0.6.1-plus-memory-raw-ops-design.md
related_pages: [concepts/mcp-transport, concepts/orchestrator-subagent-pattern]
created: 2026-06-12
updated: 2026-06-12
---

# Project Architecture: 3-Layer + LLM Wiki + Memory 3-State

## TL;DR

| 계층 | 위치 | 추적 | 목적 |
|---|---|---|---|
| **Source** | `workflow-source/` | git | Template definitions, governance rules, skill specs |
| **Runtime** | `ai-workflow/` | mixed | `memory/active/` (gitignore) + `wiki/` (git-tracked) |
| **Project Docs** | `docs/`, root | git | Actual project documentation, runbooks |

Wiki layer (v0.6.0+): Karpathy LLM Wiki 패턴 적용. `ai-workflow/wiki/` 에 6개 파일.
Memory layer (v0.6.1+): 3-state lifecycle (active/archive/release). R8 Freeze 로 archive.

## 3-Layer Separation (ADR-001, v0.5.2)

| Layer | 위치 | 갱신 정책 | 예시 |
|---|---|---|---|
| Source | `workflow-source/` | PR + review | `core/`, `templates/`, `skills/` |
| Runtime | `ai-workflow/` | session-write | `memory/`, `wiki/` |
| Project Docs | `docs/`, root | human + AI | `PROJECT_PROFILE.md`, runbooks |

## LLM Wiki Layer (ADR-004, v0.6.0)

- **Karpathy LLM Wiki 패턴** (2026-04 gist) 을 workflow state 관리에 통합
- `ai-workflow/wiki/`: git-tracked, anchor-based index (`### [[path]] {#id}`)
- 5 page types: entities/concepts/decisions/patterns/queries
- 3 workflows: ingest/query/lint
- 7 rules (R1~R7): Location, Atomicity, Pull-Before-Push, Index, Additive Merge, Topic-Branch, Merge-Resolution

## Memory 3-State Lifecycle (v0.6.1+)

| State | 위치 | Mutability | Lifecycle |
|---|---|---|---|
| **Active** | `memory/active/` _(archive after freeze)_ | mutable (session write) | session start → end |
| **Archive** | `memory/archive/YYYY-MM-DD/` | immutable (R8 freeze) | session end → freeze |
| **Release** | `memory/release/v0.5.X/` | immutable (release snapshot) | release time → deep freeze |

Freeze mechanism (R8):
- COPY (NOT move) — active/ preserved for next session
- `.frozen` YAML marker (frozen_at, source, files)
- Same-date freeze = skip (immutability)

## Related Decisions

- ADR-001: 3-Layer separation (Source/Runtime/Project Docs)
- ADR-004: LLM Wiki layer (Karpathy pattern, accepted v0.6.0)
- ADR-005 (proposed): Memory as Raw Layer (v0.6.1+)

## References

- [Karpathy LLM Wiki gist](https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f) (2026-04)
- [Plans: LLM Wiki Convergence](../../.omo/plans/llm-wiki-convergence-design.md)
- [Plans: Distributed Rules](../../.omo/plans/v0.5.11-plus-llm-wiki-distributed-rules.md)
- [Plans: Memory Raw Ops](../../.omo/plans/v0.6.1-plus-memory-raw-ops-design.md)
