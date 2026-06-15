---
type: entity
status: active
last_ingested_from: workflow-source/harnesses/codex/ + dist/harnesses/codex/v0.6.3-beta/
related_pages: [entities/standard-ai-workflow, concepts/harness-distribution, concepts/agent-topology, entities/mcp-read-only-bundle]
created: 2026-06-12
updated: 2026-06-12
---

# harness-overlay-codex

## Role

OpenAI Codex CLI harness overlay (package `standard-ai-workflow-codex`, version `v0.6.3-beta`). One of 6 harness overlays produced from `workflow-source/harnesses/codex/`. Targets Codex, which reads project-root `AGENTS.md` and merges user-global `~/.codex/config.toml`.

| Aspect | Value |
|---|---|
| Package name | `standard-ai-workflow-codex` |
| Version | `v0.6.3-beta` |
| Harness key | `codex` |
| Export profile | `agent_runtime_minimal` |
| Release focus | `workflow_skill_onboarding` |
| Deferred | `official_mcp_server_default_adoption`, `harness_mcp_activation` |
| Excluded by default | `developer_source_docs`, `global_snippet_examples`, `draft_mcp_reference_assets` |

## Entry Files

| Path | Role |
|---|---|
| `AGENTS.md` (root) | Codex project-instructions entry point. Always reads `ai-workflow/memory/active/{state.json, session_handoff.md, work_backlog.md, PROJECT_PROFILE.md}` first; references `ai-workflow/wiki/index.md` (R4 anchor) for agent query. |
| `.codex/config.toml.example` | Additive MCP snippet for `~/.codex/config.toml`. Contains only the `[mcp_servers.openaiDeveloperDocs]` block (URL + `default_tools_approval_mode = "approve"`). Does not override global model/provider defaults. |
| `workflow-source/harnesses/codex/README.md` | Source-side package guide (review points, bootstrap example, post-adopt checklist). |
| `workflow-source/harnesses/codex/apply_guide.md` | Source-side apply procedure (global/local layering, `--enable-mcp`, post-adopt checklist). |

Codex does not support project-local agent permission splits — the main/worker separation lives in `AGENTS.md` prose, not in config.

## Agent Topology

In-document main/worker pattern (less granular than [[concepts/agent-topology|OpenCode's orchestrator + doc/code/validation worker split]]). Codex has no project-local agent definitions; role separation is enforced by `AGENTS.md` operating principles, not by file boundaries.

| Role | Model (recommended) | Bounded scope |
|---|---|---|
| Main agent (orchestrator/integrator) | `main` | Hard judgment + integration; delegates below. |
| Worker — read | `small` | `ai-workflow/memory/active/` and project doc reads. |
| Worker — write | `small` | Bounded scope edits, draft generation. |
| Worker — verify | `small` | Test/validation log collection, command execution. |

Rules:

- Worker handoff: pass `responsible_files` + `exit_condition`; main agent receives core facts + result only.
- `main`/`small` split is a convention, not a config — document it in `AGENTS.md` and rely on Codex's runtime dispatch.
- First session: read `onboarding_summary.recommended_next_steps → warnings → orchestration_plan → validation_plan → code_index_update → session_start → repository_assessment.summary` (existing-project mode).

## MCP Config

Codex MCP config: TOML `[mcp_servers.<alias>]` section. Two surfaces:

| Surface | Path | Status | Notes |
|---|---|---|---|
| Bundle example | `bundle/.codex/config.toml.example` | shipped (v0.6.3-beta) | OpenAI Developer Docs MCP only (`url = "https://developers.openai.com/mcp"`). |
| Bootstrap emit | `<root>/.codex/mcp.toml` (via `--enable-mcp`) | opt-in | `bootstrap_workflow_kit.py --harness codex --enable-mcp`. |
| Global merge | `~/.codex/config.toml` | manual | Additive only; never override `model`/`provider` defaults. |
| Read-only transport | `read_only_harness_mcp_examples.json` → `harness_examples.codex` | `manual_review_only` | `transport_ready=false`; see [[entities/mcp-read-only-bundle]]. |

Bridge selection: `--mcp-bridge jsonrpc-bridge` (stable, default) or `stdio-sdk` (experimental, known connection-closed regression). Confirmed in `apply_guide.md` §9.1.

## Bundle Artifacts

`dist/harnesses/codex/v0.6.3-beta/` — 5 top-level files, 1 zip:

| File | Size class | Purpose |
|---|---|---|
| `PACKAGE_CONTENTS.md` | 2.4 KB | Layer map + recommended entry points (read first). |
| `APPLY_GUIDE.md` | 3.3 KB | Manual apply procedure + post-apply edit checklist. |
| `manifest.json` | 2.4 KB | Machine-readable: 15 included files, 0 global snippets, 2 deferred items, 3 excluded categories. |
| `standard-ai-workflow-codex-v0.6.3-beta.zip` | 25 KB | Distributable bundle (consumers use this). |
| `bundle/` (dir) | — | Mirror tree: `AGENTS.md`, `.codex/config.toml.example`, `ai-workflow/{README.md, VERSION, core/*, memory/active/*}`. |

Manifest excerpt (`manifest.json`):

```json
{
  "harness": "codex",
  "package_name": "standard-ai-workflow-codex",
  "package_version": "v0.6.3-beta",
  "optimization_profile": "agent_runtime_minimal",
  "deferred_release_items": ["official_mcp_server_default_adoption", "harness_mcp_activation"]
}
```

Apply flow (from `APPLY_GUIDE.md` §2):

1. Unzip → `bundle/` is source.
2. Copy `bundle/AGENTS.md`, `bundle/.codex/config.toml.example`, `bundle/ai-workflow/` → repo root.
3. Re-run `apply_workflow_upgrade.py` when upgrading; it handles version diff + `.gitignore`.
4. First session: read `AGENTS.md → state.json → session_handoff.md → work_backlog.md → PROJECT_PROFILE.md`.

## Related

- [[entities/standard-ai-workflow]] — parent project.
- [[concepts/harness-distribution]] — overlay packaging contract (6-harness registry).
- [[concepts/agent-topology]] — orchestrator/worker pattern Codex approximates via in-doc prose.
- [[entities/mcp-read-only-bundle]] — `manual_review_only` MCP draft backing the `[mcp_servers.openaiDeveloperDocs]` example.
- `workflow-source/harnesses/codex/README.md`, `apply_guide.md` — source-side guides.
- Sibling overlays: opencode, minimax-code, gemini-cli, antigravity, pi-dev.
