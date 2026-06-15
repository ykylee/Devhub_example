---
type: pattern
status: active
used_in: [ai-workflow/wiki/index.md, ai-workflow/memory/archive/*/work_backlog.md]
related_pages: [decisions/adr-004-wiki-layer, concepts/project-architecture]
---

# R4 Anchor-Based Index Pattern

## Problem

A knowledge index that multiple agents (or sessions) write to needs a merge-safe structure. Free-form prose in an index file causes permanent merge conflicts when two agents add entries.

## Solution

Use structured anchor-based entries:

```markdown
### [[path/to/page]] {#unique-anchor-id}
```

Each entry is a single line with `[[link]]` and `{#anchor-id}`. Merge surface is limited to the anchor level. New entries are append-only (zero conflicts).

## When to Use

- Master knowledge catalogs (`wiki/index.md`)
- Work backlog indices (`memory/active/work_backlog.md`)
- Any file that accumulates cross-references across sessions

## When Not to Use

- Deep prose documents (use additive merge R5 instead)
- Files with single-owner write patterns (append-only is unnecessary overhead)

## Related

- R4 rule from distributed rules plan (`.omo/plans/v0.5.11-plus-llm-wiki-distributed-rules.md`)
- V-4 validator (`check_wiki_index_structure.py`) — enforces `### [[path]] {#id}` format
- T2 work_backlog anchor migration (v0.6.2)
