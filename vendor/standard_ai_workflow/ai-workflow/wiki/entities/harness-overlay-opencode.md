---
type: entity
status: active
last_ingested_from: workflow-source/harnesses/opencode/ + dist/harnesses/opencode/v0.6.3-beta/
related_pages: [entities/standard-ai-workflow, concepts/harness-distribution, concepts/agent-topology, entities/mcp-read-only-bundle]
created: 2026-06-12
updated: 2026-06-12
---

# harness-overlay-opencode

## Role

OpenCode harness overlay for [[entities/standard-ai-workflow]] at `v0.6.3-beta`. Emits a full 5-agent topology (1 primary orchestrator + 4 subagent workers) under `.opencode/agents/`, plus an `AGENTS.md` top-level entry, a `opencode.json` harness config, and a project-local `standard-ai-workflow` skill. Overlay contract defined in `workflow-source/harnesses/opencode/README.md` and `apply_guide.md`; rendered output at `dist/harnesses/opencode/v0.6.3-beta/bundle/`.

Bootstrap invocation:

```
python3 scripts/bootstrap_workflow_kit.py \
  --target-root <project_root> --harness opencode --copy-core-docs
```

## Entry Files

| File | Role |
|---|---|
| `AGENTS.md` (bundle root) | Top-level workflow entry. Lists session-restoration reads, Korean-report rule, TODO project defaults. |
| `opencode.json` (bundle root) | OpenCode harness config. Sets `$schema`, `instructions` (5 files), `agent` map (5 agents), `mcp_servers.openaiDeveloperDocs` (remote). |
| `.opencode/agents/workflow-orchestrator.md` | Primary coordinator. `mode: primary`. `edit/bash/webfetch: deny`. |
| `.opencode/agents/workflow-worker.md` | Generic bounded worker. `mode: subagent`. |
| `.opencode/agents/workflow-doc-worker.md` | Document read/compare/draft worker. `mode: subagent`. |
| `.opencode/agents/workflow-code-worker.md` | Implementation/build worker. `mode: subagent`. |
| `.opencode/agents/workflow-validation-worker.md` | Check/log/evidence worker. `mode: subagent`. |
| `.opencode/skills/standard-ai-workflow/SKILL.md` | Project-local skill for session start, backlog update, doc sync, handoff. |

`opencode.json` keeps the harness config minimal — only `instructions`, `agent`, `mcp_servers` — and does not overwrite user `provider`/`model` defaults (per [[concepts/harness-distribution]] policy).

## Agent Topology

[[concepts/agent-topology]] for OpenCode: 1 primary + 4 subagents. Orchestrator is read-mostly, task-only; workers are bounded-scope executors.

| Agent | Mode | Permissions | Scope |
|---|---|---|---|
| `workflow-orchestrator` | `primary` | `edit: deny`, `bash: deny`, `webfetch: deny`; `task: deny *`, `allow workflow-*` | Coordination, prioritization, integration, user-facing report. Delegates all reads/writes/checks. |
| `workflow-worker` | `subagent` | `edit/bash/webfetch: allow` | Generic bounded task; preferred when scope does not fit a specialized worker. |
| `workflow-doc-worker` | `subagent` | `edit/bash/webfetch: allow` | Large document reads, comparisons, draft updates. |
| `workflow-code-worker` | `subagent` | `edit/bash/webfetch: allow` | Bounded code/config edits, build-oriented checks. |
| `workflow-validation-worker` | `subagent` | `edit/bash/webfetch: allow` | Checks, log inspection, evidence collection. |

Delegation rule: orchestrator gives each worker a bounded scope, clear output, and concise completion contract. Per-harness model selection — main model for orchestrator, smaller model for workers — is recommended when supported.

## MCP Config

OpenCode has no project-local MCP config in the `v0.6.3-beta` bundle. The `opencode.json` ships only a single remote MCP entry (`openaiDeveloperDocs` → `https://developers.openai.com/mcp`) for documentation lookups.

The read-only local MCP draft config lives outside the bundle: `workflow-source/examples/mcp_config_examples/opencode-mcp.json` (key `standardAiWorkflowReadOnly`, transport `local`, module `workflow_kit.server.read_only_jsonrpc`). Apply to project as `mcp.opencode.json` only after reviewing the [[entities/mcp-read-only-bundle]] `transport_ready=false` status and `manual_review_only` flag.

## Bundle Artifacts

Source: `dist/harnesses/opencode/v0.6.3-beta/bundle/`

| Path | Notes |
|---|---|
| `AGENTS.md` | Workflow entry, kit-bom marker `v0.6.3-beta`. |
| `opencode.json` | 5-agent config + OpenAI docs MCP. |
| `ai-workflow/` | Mirrored workflow state docs layer (memory, wiki, core, scripts). |
| `.opencode/agents/` | 5 persona files (see Agent Topology). |
| `.opencode/skills/standard-ai-workflow/SKILL.md` | Project-local workflow skill. |

`export_harness_package.py` produces this bundle; `tools/check_packaging.py` runs the packaging smoke gate.

## Related

- [[entities/standard-ai-workflow]] — parent project entity; this overlay is one of 6 harness overlays.
- [[concepts/harness-distribution]] — overlay-vs-policy contract; minimal-key rule for `opencode.json`.
- [[concepts/agent-topology]] — orchestrator/worker delegation pattern this overlay instantiates.
- [[entities/mcp-read-only-bundle]] — read-only MCP draft that an OpenCode project can opt into via `mcp.opencode.json`.
- Source: `workflow-source/harnesses/opencode/README.md`, `workflow-source/harnesses/opencode/apply_guide.md`.
- Rendered: `dist/harnesses/opencode/v0.6.3-beta/bundle/`.
- MCP draft: `workflow-source/examples/mcp_config_examples/opencode-mcp.json`.
