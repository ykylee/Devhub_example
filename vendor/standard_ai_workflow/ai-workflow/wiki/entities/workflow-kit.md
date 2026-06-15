---
type: entity
status: active
last_ingested_from: workflow-source/workflow_kit/README.md
related_pages: [entities/standard-ai-workflow, entities/workflow-source, concepts/contract-v1-output-validation, concepts/mcp-transport]
created: 2026-06-12
updated: 2026-06-12
---

# workflow_kit

## Role

Reusable Python package (`workflow-source/workflow_kit/`) installed editable via `pip install -e`. Consolidates logic previously scattered across scripts and prototypes into importable modules. Currently homes the read-only MCP utility layer, output-contract runtime maps, and the v0.5.4+ orchestrator↔sub-agent contract enforcement helpers. Not a runtime service — every consumer (skill, MCP server, orchestration script) imports its submodules directly.

## Modules

Top-level layout: `common/`, `contract_v1/`, `server/`, plus `constants.py`, `gitignore.py`, `harness/`, `pyproject.toml`, `upgrade_diff.py`.

### `common/`

30+ submodules. Core shared helpers consumed by skills, MCP servers, and runners.

| Submodule | Responsibility |
| --- | --- |
| `paths` | Repo path resolution |
| `runner` | JSON subprocess exec, step failure structuring, repeat-flag assembly, top-level warnings/orchestration payload |
| `errors` | Runner top-level step failure wrapping, error JSON payload |
| `contracts` / `schemas` / `output_contracts` | Output contract runtime map (single source for generated JSON Schema + sample smoke) |
| `state` | Workflow state doc assembly |
| `markdown` / `docs` / `project_docs` | MD link parsing, broken-link detect, metadata gap check, profile/handoff/backlog parsing |
| `text` / `normalize` | String normalization, common dedupe |
| `change_types` | Changed-file classification, validation change-type inference |
| `reconcile` | Handoff/backlog diff explainer |
| `planning` | Validation-level calc, conservative task-status decision |
| `doc_sync` | Doc-sync document candidate assembly |
| `session_outputs` | Session summary / recommended actions / merge reconcile note |
| `scaffold` | Latest-backlog step assembly, optional path flag |
| `read_only_bundle` | Common callable layer for the 5 read-only MCP tools |
| `doc_transformer`, `exploration`, `exploration_scope`, `git`, `ingest`, `linter`, `milestones`, `modes`, `patching`, `rotation`, `workflow_state`, `workflow_writes`, `writing_bundle` | Auxiliary helpers (newer / narrower scope) |

### `contract_v1/`

Pydantic v2 enforcement for the external [[concepts/contract-v1-output-validation|orchestrator↔sub-agent contract v1]].

| Submodule | Role |
| --- | --- |
| `output_validator` | Validates sub-agent response against contract v1 schema; rejects 7 MUST-NOT-delegate patterns (v0.5.6 P0) |
| `delegator` | `choose_role` (single) + `choose_roles` (multi-sub fan-out, v0.5.7), `recommend_model_tier` |

### `server/`

Read-only MCP bridge + official SDK candidate. See [[concepts/mcp-transport|mcp-transport]] for transport details.

| Submodule | Role |
| --- | --- |
| `read_only_registry` | Draft tool registry, generated JSON Schema exposure |
| `read_only_tools` | Direct-call adapter |
| `read_only_entrypoint` | Schema-validated entrypoint |
| `read_only_jsonrpc` | JSON-RPC draft bridge (default stable transport, v0.5.7+) |
| `read_only_mcp_sdk` | Official MCP Python SDK low-level stdio server candidate (experimental, v1.0) |
| `mcp_v1_server` | Newer v1.0 SDK server entrypoint |

## Versioning

Semver string exported as `workflow_kit.__version__`. Current value: `v0.6.3-beta`. Code paths read `tool_version` from this single attribute — no hardcoded duplicates. Version bumps drive wheel packaging (`tools/check_packaging.py` smoke, v0.5.8+) and harness export bundle filenames. Package is shipped as part of the editable install; the `v0.5.7.1` wheel previously missed `common.state`/`contracts`/`schemas` — fixed in subsequent releases.

## Related

- [[entities/standard-ai-workflow]] — parent repo entity
- [[entities/workflow-source]] — source tree where this package lives
- [[concepts/contract-v1-output-validation]] — contract enforced by `contract_v1/`
- [[concepts/mcp-transport]] — transport context for `server/`
