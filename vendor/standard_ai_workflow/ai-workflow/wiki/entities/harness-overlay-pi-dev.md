---
type: entity
status: active
last_ingested_from: workflow-source/harnesses/pi-dev/
related_pages: [entities/standard-ai-workflow, concepts/harness-distribution, concepts/agent-topology]
created: 2026-06-12
updated: 2026-06-12
---

# harness-overlay-pi-dev

## Role

Pi Coding Agent (pi.dev) harness overlay for [[entities/standard-ai-workflow]] (v0.6.3-beta). One of 6 harness overlays produced from `workflow-source/harnesses/pi-dev/`. Targets Pi — a "Primitives, not features" minimal agent — and binds it to the project's `state.json` / `session_handoff.md` strict source-of-truth discipline.

| Aspect | Value |
|---|---|
| Harness key | `pi-dev` |
| Target agent | Pi Coding Agent (pi.dev) |
| Source dir | `workflow-source/harnesses/pi-dev/` (3 files) |
| Bootstrap emit | `<project_root>/AGENTS.md` (sole overlay file) |
| Persona file | `SYSTEM.md` (declared in README, prototype — not emitted by bootstrap; manual write) |
| MCP | None — no MCP config snippet shipped; Pi invokes `ai-workflow/scripts/*` tools directly per `AGENTS.md` §4 |

## Entry Files

| Path | Role | Emitted? |
|---|---|---|
| `AGENTS.md` (project root) | Pi project-instructions entry point. Mandates the Standard AI Workflow; priority docs are `state.json` → `session_handoff.md` → `work_backlog.md`. Enforces session-start routine (read `current_focus` + `next_documents`), `Research → Strategy → Execution` loop, and `backlog/YYYY-MM-DD.md` updates. References `ai-workflow/wiki/index.md` (R4 anchor) for agent query. | yes (bootstrap) |
| `SYSTEM.md` | Persona / operating-principles supplement to `AGENTS.md`. | no (prototype; manual) |
| `workflow-source/harnesses/pi-dev/README.md` | Source-side package guide (scope, status, included files). | source only |
| `workflow-source/harnesses/pi-dev/apply_guide.md` | Source-side apply procedure (precheck, overlay copy, first-session verify, troubleshooting, MCP section). | source only |

`apply_guide.md` §2 explicitly notes `SYSTEM.md` is declared as an included file but is not yet emitted by `bootstrap_workflow_kit.py`; adopters write it manually if a persona layer is needed.

## Agent Topology

Persona-based, no separate workers dir. Pi has no project-local agent definitions; role separation lives in `AGENTS.md` prose and the strict `state.json` contract — not in file boundaries. Aligns with [[concepts/agent-topology]] via in-doc convention rather than config.

| Layer | Mechanism | Source of truth |
|---|---|---|
| Persona | `SYSTEM.md` (when present) + `AGENTS.md` §5 language guide | manual pair file |
| Main agent | `AGENTS.md` §1 session-start routine | `ai-workflow/memory/active/state.json` |
| Tool use | `AGENTS.md` §4 — `python3 ai-workflow/scripts/*` direct invocation; structured JSON preferred | workflow-source scripts |
| State handoff | `AGENTS.md` §3 — daily `backlog/YYYY-MM-DD.md` update + end-of-session `state.json` + `session_handoff.md` refresh | `ai-workflow/memory/active/` |

Rules:

- `state.json` is the single source of truth — Pi must read it before any other doc.
- Backlog updates are mandatory on state change, not optional.
- `AGENTS.md` §4 routes complex workflow control through `workflow-source/scripts/` rather than MCP — this is Pi's substitute for the MCP surface other harnesses get.

## MCP Config

None. Pi-dev ships no MCP config snippet and bootstrap does not emit one.

| Surface | Status | Notes |
|---|---|---|
| `<root>/.mcp.json` / `<root>/pi.mcp.json` | not emitted | apply_guide.md §6.1: "별도 MCP config 파일을 생성하지 않음" |
| `AGENTS.md` §4 tool-use | substitute | `python3 ai-workflow/scripts/*` direct calls, structured JSON |
| Read-only transport | n/a | `apply_guide.md` §6.2 lists bridge options (`jsonrpc-bridge` stable, `stdio-sdk` experimental) but only for direct script invocation, not as an MCP server config |

`--enable-mcp` is accepted by `bootstrap_workflow_kit.py --harness pi-dev` for symmetry with other harnesses, but produces no MCP config file — see `apply_guide.md` §6.1.

## Related

- [[entities/standard-ai-workflow]] — parent project.
- [[concepts/harness-distribution]] — 6-harness registry contract (Codex, OpenCode, MiniMax Code, Gemini CLI, Antigravity, pi-dev).
- [[concepts/agent-topology]] — persona/in-doc role split model pi-dev approximates.
- `workflow-source/harnesses/pi-dev/README.md`, `apply_guide.md` — source-side guides.
- Sibling overlays: codex, opencode, minimax-code, gemini-cli, antigravity.
