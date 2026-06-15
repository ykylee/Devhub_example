---
type: entity
status: active
last_ingested_from: workflow-source/harnesses/gemini-cli/
related_pages: [entities/standard-ai-workflow, concepts/harness-distribution, concepts/agent-topology, entities/mcp-read-only-bundle]
created: 2026-06-12
updated: 2026-06-12
---

# harness overlay — Gemini CLI

## Role

Gemini CLI harness overlay (v0.6.3-beta). One of six [[concepts/harness-distribution|harness distribution]] targets (`codex`, `opencode`, `gemini-cli`, `antigravity`, `minimax-code`, `pi-dev`). Emitted by `bootstrap_lib` when `--harness gemini-cli` is passed. Source: `workflow-source/harnesses/gemini-cli/` (README.md, apply_guide.md).

| Property | Value |
|---|---|
| Harness target id | `gemini-cli` |
| Source dir | `workflow-source/harnesses/gemini-cli/` |
| Bootstrap flag | `--harness gemini-cli` |
| Source status | draft (last updated 2026-04-25) |
| Overlay style | Single-file entry point + optional MCP json |
| Granularity vs. OpenCode/Codex | Lower — main+worker split only, no per-role agent dir |

## Entry Files

Single entry point: `GEMINI.md` at project root. No second-level config file is emitted by bootstrap.

| Path | Role | Priority |
|---|---|---|
| `GEMINI.md` | Project instruction entry point for Gemini CLI | **Overrides system prompt** — strongest directive in the harness chain |
| `ai-workflow/memory/active/PROJECT_PROFILE.md` | Project-specific rules (linked from `GEMINI.md`) | Read first per `GEMINI.md` instruction |
| `ai-workflow/memory/active/session_handoff.md` | Last session's handoff (linked from `GEMINI.md`) | Read first per `GEMINI.md` instruction |
| `ai-workflow/memory/active/work_backlog.md` | Backlog index (linked from `GEMINI.md`) | Read first per `GEMINI.md` instruction |
| `ai-workflow/memory/active/repository_assessment.md` | Existing-project analysis | Read on first session of `existing` mode |

`GEMINI.md` priority rule (from `apply_guide.md` §2): "Gemini CLI 에서는 `GEMINI.md` 가 시스템 프롬프트보다 우선하는 강력한 지침이므로, 핵심 운영 원칙을 여기에 명시한다." This is a sharper contract than Codex (where `AGENTS.md` is advisory) or OpenCode (where `opencode.json` shares authority).

## Agent Topology

`invoke_agent`-based main+worker split. Less granular than [[concepts/agent-topology|OpenCode orchestrator+doc/code/validation worker]] or [[concepts/orchestrator-subagent-pattern|Codex task-only orchestrator]].

| Layer | Mechanism | Granularity |
|---|---|---|
| Main coordinator | Primary Gemini CLI session reads `GEMINI.md` + active memory | Single |
| Worker sub-agents | `invoke_agent` calls with bounded scope (read/write/verify) | One tier, role-by-task |
| Contract | Implicit (no `contract_v1` enforcement; orchestrator ↔ sub-agent enforcement only applies to OpenCode/Codex paths) | None |

`apply_guide.md` §2.2 recommends: "메인 에이전트는 조정/통합에 집중하고, bounded scope 읽기/쓰기/검증은 `invoke_agent` 를 통해 서브 에이전트로 분리". No per-role agent directory is generated (cf. OpenCode's `.opencode/agents/` or MiniMax Code's `.minimax/agents/`).

## MCP Config

Optional, read-only. Path differs by scope (project-local vs. global).

| Scope | Path | Emitted by |
|---|---|---|
| Project-local | `.gemini/mcp.json` | `bootstrap_workflow_kit.py --enable-mcp` |
| Global | `~/.gemini/settings.json` → `mcpServers` block | Manual merge |
| Transport | jsonrpc-bridge (stable) or stdio-sdk (experimental) | `--mcp-bridge` flag |
| Server module | `workflow_kit.server.read_only_jsonrpc` | `[[entities/mcp-read-only-bundle]]` |
| Tools exposed | `latest_backlog`, `check_doc_metadata`, `check_doc_links`, `suggest_impacted_docs` | See [[entities/mcp-read-only-bundle]] |

Gemini CLI MCP keys differ from the other five harnesses: the field is `mcpServers` (plural, object) with `trust: true` + `includeTools` array per server. No mcp bundle is emitted unless `--enable-mcp` is passed — unlike core docs, MCP is opt-in for this overlay.

## Related

- [[entities/standard-ai-workflow]] — parent project entity.
- [[concepts/harness-distribution]] — six-harness overlay pattern, of which this is one target.
- [[concepts/agent-topology]] — main+worker vs. orchestrator+specialized-worker comparison.
- [[entities/mcp-read-only-bundle]] — read-only MCP descriptor exposed via `.gemini/mcp.json`.
- Source files: `workflow-source/harnesses/gemini-cli/README.md`, `workflow-source/harnesses/gemini-cli/apply_guide.md`.
- Sibling overlays: `codex/`, `opencode/`, `antigravity/`, `minimax-code/`, `pi-dev/` (see [[concepts/harness-distribution]]).
