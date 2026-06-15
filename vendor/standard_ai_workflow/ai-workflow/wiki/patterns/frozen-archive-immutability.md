---
type: pattern
status: active
used_in: [workflow-source/skills/memory-freeze, ai-workflow/memory/archive]
related_components: [concepts/memory-3-state-lifecycle, concepts/wiki-source-rule-r9, decisions/adr-004-wiki-layer, decisions/adr-005-r9-wiki-source-rule]
created: 2026-06-12
updated: 2026-06-12
r9_skip: true
---

# R8 Frozen-Archive Immutability Pattern

## TL;DR

| Field | Value |
|---|---|
| **Trigger** | session end (D8 auto) · release time · manual `memory-freeze` skill fire |
| **Action** | COPY `active/` → `archive/YYYY-MM-DD/` + write `.frozen` YAML marker |
| **Output** | `archive/YYYY-MM-DD/{*.md,*.json}` + `.frozen` (`frozen_at`, `source`, `files`) |
| **Immutability boundary** | `archive/YYYY-MM-DD/` is read-only raw source after freeze. NEVER mutate. R10 lint verifies. |
| **Same-date re-fire** | SKIP. status = `skipped`. Existing freeze is preserved (immutability invariant). |
| **Reverse direction** | FORBIDDEN. `archive/ → active/` rollback not allowed. Fix in next session's `active/`. |

## Problem

Mutable `active/` state leaking into wiki/RAG ingest causes three failure modes:

| Failure | Symptom | Root cause |
|---|---|---|
| **Drift** | wiki snapshot reflects in-progress half-edited file | wiki-ingest reads `active/` directly (R9 violation) |
| **Partial freeze** | archive/ snapshot missing files because move was used | freeze = move breaks next-session cold start (D9) |
| **Hot-write conflict** | agent writes `active/` while freeze is in flight; archive captures torn state | no atomic boundary between write and freeze copy |

Wiki needs a **stable, time-anchored raw source**. Mutable `active/` cannot be it. A dedicated `archive/` snapshot — written once, read many — is the boundary.

## Solution

Three primitives enforce the freeze-then-immutable contract.

### 1. COPY (NOT MOVE)

```
active/  ──── copy ────►  archive/YYYY-MM-DD/
   │                            │
   │ keep                       │ immutable
   ▼                            ▼
next session                wiki-ingest source
cold start                  (R9)
```

- `active/` files are never deleted by freeze.
- Next session reuses `active/` as the mutable workspace (per-session granularity, D9).
- Reverse direction (`archive/ → active/`) is structurally impossible — no rollback code path.

### 2. `.frozen` YAML Marker

Writes a single YAML file at `archive/YYYY-MM-DD/.frozen`:

```yaml
frozen_at: 2026-06-12T18:42:00+09:00
source: ai-workflow/memory/active/
files: [session_handoff.md, work_backlog.md, state.json, backlog/2026-06-12.md]
```

| Key | Purpose |
|---|---|
| `frozen_at` | ISO timestamp. Audit + same-date dedup. |
| `source` | absolute path to the `active/` root that was frozen. |
| `files` | explicit list. R10 lint cross-checks against archive directory contents. |

Marker is the **single source of truth** for "this archive is sealed". Presence of `.frozen` = immutable. Absence = not yet frozen (or pre-R8).

### 3. Same-Date Skip

| Condition | Behavior | Why |
|---|---|---|
| `archive/YYYY-MM-DD/.frozen` exists | skip. status = `skipped`. no files touched. | Immutability invariant: existing freeze MUST NOT be re-written or merged. |
| `archive/YYYY-MM-DD/` exists but no `.frozen` | abort with `error_code: partial_freeze_detected`. | Partial state from a crashed prior run. Manual recovery required. |
| `archive/YYYY-MM-DD/` absent | proceed with full copy + marker write. | Normal case. |

### 4. R10 Freeze Lint (post-condition)

`workflow-source/tests/check_memory_freeze_lint.py` runs after every freeze and verifies:

- `.frozen` marker is valid YAML with all 3 required keys.
- `files:` list matches archive directory contents exactly (no extra, no missing).
- `frozen_at` is parseable ISO timestamp.
- No file in archive/ has mtime later than `.frozen` mtime (no late writes).

Lint failure = freeze is not accepted. Rerun required.

## When to Use

| Trigger | How | Output |
|---|---|---|
| **Session end (D8 auto)** | `python3 scripts/run_memory_freeze.py` on session-end signal | `archive/<today>/` snapshot |
| **Release time** | release procedure deep-freezes `archive/` entire span → `release/v0.X.Y/` | release snapshot |
| **Manual R8 fire** | operator runs `memory-freeze` skill explicitly (incident recovery) | ad-hoc archive |
| **Cross-machine handoff** | freeze on machine A, resume on machine B from `archive/<date>/` | portable state boundary |

## When Not to Use

| Anti-pattern | Why not | Use instead |
|---|---|---|
| **In-session hot write** | freeze is a one-shot per-session boundary, not a write coordinator. Calling mid-session breaks the active workspace. | Direct `active/` write (R3 hot path) |
| **Tests / fixtures** | freeze + `.frozen` is a real-git operation. Test runs would pollute `archive/` with synthetic dates. | Mock at the script boundary; do not invoke real `run_memory_freeze.py` in unit tests |
| **Wiki-ingest of mutable state** | R9 explicitly forbids `active/` as wiki source. Freeze must run first. | Wait for freeze, then ingest from `archive/YYYY-MM-DD/` |
| **Cross-session mutation of frozen archive** | immutability invariant is hard, not soft. "Just one fix" in archive/ breaks audit chain. | Apply fix in next session's `active/`; let next freeze supersede. |
| **Re-encoding / reformatting on freeze** | freeze is byte-stable. Reformatting = re-writing. Forbidden. | Keep raw bytes exact. Reformat only in `active/` before freeze. |

## Related

| Ref | Relationship |
|---|---|
| [[concepts/memory-3-state-lifecycle]] | defines `active` ↔ `archive` ↔ `release` 3-state model that this pattern implements |
| [[concepts/wiki-source-rule-r9]] | R9 wiki-ingest = `archive/...` only; this pattern produces the only legal ingest source |
| [[decisions/adr-005-r9-wiki-source-rule]] | ADR-005 (Memory as Raw Layer, proposed) — formalizes the immutability boundary as policy |
| [memory-freeze skill](../../../../workflow-source/skills/memory-freeze/SKILL.md) | R8 implementation: `scripts/run_memory_freeze.py` + `.frozen` marker writer |
| [R10 freeze lint](../../../../workflow-source/tests/check_memory_freeze_lint.py) | post-freeze integrity check; enforces `files:` list match + mtime invariant |
| [MEMORY_GOVERNANCE §4](../../../../workflow-source/MEMORY_GOVERNANCE.md) | freeze rule source-of-truth: COPY not move, same-date skip, archive read-only |
