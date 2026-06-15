---
type: entity
status: active
last_ingested_from: workflow-source/harnesses/minimax-code/
related_pages: [entities/standard-ai-workflow, concepts/harness-distribution, concepts/agent-topology, entities/mcp-read-only-bundle]
created: 2026-06-12
updated: 2026-06-12
---

# harness-overlay-minimax-code

## Role

MiniMax Code harness overlay for [[entities/standard-ai-workflow]] at `v0.6.3-beta`. One of 6 harness overlays. Targets the MiniMax Code CLI, which reads project-root `AGENTS.md` (shared with Codex/OpenCode) plus a harness-specific `MiniMax.md` orchestrator entry. Same 4-worker topology as [[entities/harness-overlay-opencode|opencode]] and [[entities/harness-overlay-codex|codex]] (in-doc separation), but MiniMax Code is the only harness that pairs `AGENTS.md` with a dedicated `<harness>.md` orchestrator doc.

| Aspect | Value |
|---|---|
| Harness key | `minimax-code` |
| Version | `v0.6.3-beta` |
| Source dir | `workflow-source/harnesses/minimax-code/` |
| Source files | `README.md`, `apply_guide.md` (no `_template` clone — uses inline overlay pattern) |
| Dist bundle | **not produced** — see [Note](#note) |
| Priority doc | `MiniMax.md` (over `AGENTS.md` if they conflict) |

## Entry Files

| Path | Role |
|---|---|
| `AGENTS.md` (root) | Shared with Codex/OpenCode. Workflow rules summary, session-restoration reads, wiki index pointer. |
| `MiniMax.md` (root) | MiniMax Code-only orchestrator entry. Operating principles, Korean-report rule, orchestrator/worker contract. Takes precedence over `AGENTS.md` on conflict (per `README.md` §1). |
| `MiniMax_config.example.json` (root) | Source-of-truth config template. Copy to `.minimax/config.json` and fill `project_name`, `agents.*.file`, `mcp_servers.*.command`. |
| `.MiniMax/agents/workflow-orchestrator.md` | Main orchestrator persona (task decomposition, worker delegation, handoff/state sync). |
| `.MiniMax/agents/workflow-worker.md` | Generic worker contract (`WorkerTask` / `WorkerResponse` schemas). |
| `.MiniMax/agents/workflow-doc-worker.md` | Doc reconciliation / catalog sync (`workflow-source/skills/doc-sync`). |
| `.MiniMax/agents/workflow-code-worker.md` | Code implementation / refactor (`robust_patcher`, `code-index-update`). |
| `.MiniMax/agents/workflow-validation-worker.md` | Test/smoke execution + log capture (`validation-plan`, `check_*.py`). |
| `workflow-source/harnesses/minimax-code/README.md` | Source-side package guide. |
| `workflow-source/harnesses/minimax-code/apply_guide.md` | Source-side apply procedure (config init, first-session verification, troubleshooting). |

Note the path casing: overlay lives at `.MiniMax/agents/` (capital M) per the bootstrap registry, even though the README uses `.minimax/` lowercase. Both casings appear in source docs.

## Agent Topology

Mirrors the [[concepts/agent-topology|OpenCode 1+4 orchestrator/worker pattern]]: 1 primary orchestrator + 4 subagent workers, all defined as project-local agent files. Identical role set to OpenCode (worker / doc / code / validation).

| Agent | Mode | Scope |
|---|---|---|
| `workflow-orchestrator` | primary (read-mostly) | Coordination, prioritization, integration, Korean report. Delegates all reads/writes/checks. |
| `workflow-worker` | subagent | Generic bounded task. Preferred when scope does not fit a specialized worker. |
| `workflow-doc-worker` | subagent | Document reads, comparisons, draft updates, catalog sync. |
| `workflow-code-worker` | subagent | Bounded code/config edits, build-oriented checks. |
| `workflow-validation-worker` | subagent | Test/smoke execution, log inspection, evidence collection. |

Per `README.md` §2, the orchestrator persona pairs with `workflow-source/prompts/code_worker_prompt.md`; worker contract uses `workflow_kit.common.schemas.worker.*`. No `mode: deny` perms are declared in `MiniMax.md` (unlike OpenCode's `opencode.json` perms) — separation is enforced by prompt contract, not config.

## MCP Config

`MiniMax_config.example.json` ships with an `mcp_servers.standardAiWorkflowReadOnly` block pointing at the [[entities/mcp-read-only-bundle|read-only MCP bundle]]. MCP key in MiniMax config is `mcp_servers` (TOML-style JSON), not `mcpServers`/`mcp` as in some other harnesses.

| Surface | Path | Status | Notes |
|---|---|---|---|
| Template example | `MiniMax_config.example.json` → `.minimax/config.json` | shipped | `mcp_servers.standardAiWorkflowReadOnly` + `env.PYTHONPATH=workflow-source`. |
| Bootstrap emit (per-harness) | `<root>/.MiniMax/mcp.json` (via `--enable-mcp`) | opt-in | `bootstrap_workflow_kit.py --harness minimax-code --enable-mcp`. |
| Global merge | `~/.minimax/config.json` `mcp_servers` block | manual | Add `standardAiWorkflowReadOnly` entry; absolute `PYTHONPATH` + `STANDARD_AI_WORKFLOW_ROOT`. |
| Read-only transport | `read_only_harness_mcp_examples.json` → `harness_examples.minimax-code` | `manual_review_only` | `transport_ready=false`; see [[entities/mcp-read-only-bundle]]. |

Secrets (user auth tokens, API keys) **must** be injected via env vars or external secret manager — never inlined into `config.json`. Bridge selection: `jsonrpc-bridge` (default, stable) or `stdio-sdk` (experimental, known connection-closed regression). MCP example files: `workflow-source/examples/mcp_config_examples/minimax-code-mcp.json` (+ `…stdio-sdk.json` variant).

## Note

**MiniMax Code is not in `export_harness_package.py → SUPPORTED_HARNESSES`.** Current registry (`codex`, `opencode`, `gemini-cli`, `pi-dev`, `antigravity`) has 5 entries; `minimax-code` is the 6th overlay, served by `bootstrap_lib.harnesses` but skipped by the export pipeline. As a result:

- No `dist/harnesses/minimax-code/<version>/` bundle is emitted.
- No `manifest.json` / `PACKAGE_CONTENTS.md` / `APPLY_GUIDE.md` is generated.
- Consumers must bootstrap directly: `python3 workflow-source/scripts/bootstrap_workflow_kit.py --harness minimax-code --adoption-mode existing --copy-core-docs [--enable-mcp]`.

To add MiniMax Code to the export pipeline, register a builder in `export_harness_package.py` and append `"minimax-code"` to `SUPPORTED_HARNESSES` (matching the [[concepts/harness-distribution|6-harness registry]]).

## Related

- [[entities/standard-ai-workflow]] — parent project entity; this overlay is one of 6 harness overlays.
- [[concepts/harness-distribution]] — 6-harness overlay model; explains why MiniMax Code gets a paired `AGENTS.md` + `MiniMax.md` entry.
- [[concepts/agent-topology]] — orchestrator/worker pattern this overlay instantiates.
- [[entities/mcp-read-only-bundle]] — `mcp_servers.standardAiWorkflowReadOnly` source backing the MiniMax Code example.
- Source: `workflow-source/harnesses/minimax-code/README.md`, `workflow-source/harnesses/minimax-code/apply_guide.md`.
- MCP examples: `workflow-source/examples/mcp_config_examples/minimax-code-mcp.json`, `…stdio-sdk.json`.
- Sibling overlays: [[entities/harness-overlay-codex|codex]], [[entities/harness-overlay-opencode|opencode]], gemini-cli, antigravity, pi-dev.
