---
type: entity
status: active
last_ingested_from: ai-workflow/README.md
related_pages: [entities/standard-ai-workflow, entities/workflow-source, concepts/memory-3-state-lifecycle, concepts/wiki-source-rule-r9]
created: 2026-06-12
updated: 2026-06-12
r9_skip: true
---

# ai-workflow/

## Role

Runtime layer directory. Bootstrap output target — files here are produced by `python3 -m bootstrap_lib` (or `workflow-source/scripts/bootstrap_workflow_kit.py`) from canonical sources under `workflow-source/`.

- **R1 / D1 / D2 boundary**: `workflow-source/` is the source-of-truth; `ai-workflow/` is the per-project runtime mirror consumed by the AI harness at session time.
- **Non-project scope**: contents are workflow state, not project code/docs. `WORKFLOW_INDEX.md` §2 mandates exclude `ai-workflow/` from code base analysis and semantic search.
- **Mixed git tracking**: bootstrap output is partially gitignored (scripts, skills, harness overlays); `core/`, `mcp_servers/`, `wiki/`, `memory/` scaffolding is checked in, session state inside `memory/active/` is local-only.
- **Three writer classes**: (1) bootstrap emit, (2) session runtime write, (3) wiki-ingest.

## Subdirectories

| Path | Tracking | Purpose |
|---|---|---|
| `core/` | checked-in | Bootstrap-synced copies of `workflow-source/core/` — read-only runtime reference. |
| `mcp_servers/` | checked-in | Runtime copies of MCP server specs (jsonrpc-bridge stable, stdio-sdk experimental). |
| `memory/` | mixed | Session state root — `active/`, `archive/`, `release/`, harness subtrees (`codex/`, `gemini/`). |
| `wiki/` | checked-in (v0.6.0+) | LLM wiki layer emitted by `--enable-wiki` flag; contains `entities/`, `concepts/`, `decisions/`, `index.md`, `log.md`. |
| `README.md` | checked-in | Bootstrap result guide for the local project. |
| `WORKFLOW_INDEX.md` | checked-in | Runtime-only ops index; warns against using this path for project understanding. |

## Boundary Rules (R1, D1, D2)

| Rule | Direction | Constraint |
|---|---|---|
| R1 | `workflow-source/` → `ai-workflow/` | Only the bootstrap pipeline may write into `ai-workflow/{core,mcp_servers,wiki}`. Manual edits are overwritten on re-bootstrap. |
| D1 | `ai-workflow/` ↛ `workflow-source/` | Runtime must never promote changes back upstream; source edits live in a separate PR to `workflow-source/`. |
| D2 | `docs/` ↛ `ai-workflow/memory/` | Project operational docs (runbook, handoff, release notes) stay under `docs/`; workflow state docs stay under `ai-workflow/memory/active/`. |

## Runtime Lifecycle

Five-phase flow from source to wiki ingestion:

| Phase | Actor | Action |
|---|---|---|
| 1. Bootstrap | `bootstrap_lib` | Copies `workflow-source/core/`, `workflow-source/mcp_servers/`, `workflow-source/templates/` into `ai-workflow/{core,mcp_servers,memory/active}`; emits harness overlay (e.g. `ANTIGRAVITY.md`). |
| 2. Session write | harness + agent | Writes `memory/active/<branch>/session_handoff.md`, `work_backlog.md`, `state.json`, `backlog/YYYY-MM-DD.md`. |
| 3. Freeze (R8) | agent | On session close, snapshots `memory/active/<branch>/` → `memory/archive/<branch>/` with `release/` ledger. |
| 4. Wiki-ingest (R9) | wiki sync | Mirrors finalized state docs and source changes into `ai-workflow/wiki/{entities,concepts,decisions}/` per [[concepts/wiki-source-rule-r9]]. |
| 5. Reset | bootstrap rerun | Next session can re-bootstrap core/mcp layers without touching `memory/archive/` or `wiki/`. |

> [!NOTE]
> Phase boundaries are soft: R9 wiki-ingest may run alongside active sessions; R8 freeze is mandatory before archive, not before wiki-ingest.

## Writers & Consumers

| Path | Writer | Consumer |
|---|---|---|
| `core/` | `bootstrap_lib` | Harness (read-only at session start). |
| `mcp_servers/` | `bootstrap_lib` | MCP runtime (jsonrpc-bridge or stdio-sdk). |
| `memory/active/<branch>/` | agent during session | Next session via `session_handoff.md` + `state.json`. |
| `memory/archive/<branch>/` | freeze tool (R8) | Release audit, postmortem. |
| `wiki/entities/`, `wiki/concepts/`, `wiki/decisions/` | wiki sync (R9) | LLM context primer, query layer. |

## Related

- [[entities/standard-ai-workflow]] — parent project entity.
- [[entities/workflow-source]] — canonical source-of-truth mirrored into this runtime.
- [[concepts/memory-3-state-lifecycle]] — `active` → `archive` → `release` transition rules.
- [[concepts/wiki-source-rule-r9]] — ingestion rules for phase 4 above.
- `entities/bootstrap-wiki-py` — emitter of the `wiki/` subtree (v0.6.0+).
