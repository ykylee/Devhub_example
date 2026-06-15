---
type: pattern
status: active
used_in: [workflow-source/scripts/bootstrap_lib/wiki.py, workflow-source/scripts/bootstrap_workflow_kit.py]
related_components: [entities/bootstrap-wiki-py, decisions/adr-004-wiki-layer, concepts/project-architecture]
created: 2026-06-12
updated: 2026-06-12
---

# Wiki Skeleton Emit Pattern

## TL;DR

| Aspect | Value |
|---|---|
| Trigger | `python3 -m bootstrap_lib --enable-wiki ...` (or shim) |
| Inputs | `argparse.Namespace` (`--force` honored), `Paths`, harness list |
| Outputs | 4 files at `<target>/ai-workflow/wiki/` |
| Entry point | `write_wiki_files(args, paths, harnesses) -> dict[str, str]` |
| Coverage | Harness-agnostic — every harness maps to the same renderer |

## Problem

A fresh checkout of any project bootstrapped by the kit ships with
`ai-workflow/memory/`, `core/`, harness overlays — but **no wiki layer**.
Downstream agents that rely on the LLM Wiki
(`ai-workflow/wiki/`) hit a missing directory on first read.

Compounding factors:

- **Multi-harness parallel bootstrap**: 6 harnesses (codex, opencode,
  gemini-cli, antigravity, MiniMax-code, pi-dev) can be targeted in one
  run. Wiki emit must be idempotent across all of them.
- **Source-bundled vs wheel install**: real projects read the source
  files from `ai-workflow/wiki/` in the standard-ai-workflow repo; wheel
  installs may not bundle data. Both paths must produce a usable
  skeleton, not an empty directory.
- **Idempotency + force**: re-runs must not clobber local edits unless
  `--force` is explicit.

## Solution

Wire a single opt-in flag — `--enable-wiki` — that delegates to
`bootstrap_lib.wiki.write_wiki_files()`. The function:

1. Resolves the wiki source root via `_wiki_source_root()` (looks up
   `<repo_root>/ai-workflow/wiki/` in the source repo).
2. Calls four pure renderers (`render_wiki_schema`, `render_wiki_index`,
   `render_wiki_log`, `render_wiki_gitignore`). Each renderer tries the
   bundled source first, falls back to an inline `_INLINE_STUB_*`
   constant when source is absent.
3. Materializes `<target>/ai-workflow/wiki/` and writes all 4 files
   through the shared `write_text()` helper, which respects
   `args.force`.
4. Returns a `{overlay_key: file_path}` map for manifest reporting.

The `WIKI_CONFIG_RENDERERS` dispatch table funnels all 6 harnesses to
the single `render_wiki_schema` renderer — the wiki layer is
deliberately harness-agnostic.

## Emitted Files

| File | Source | Renderer | Fallback |
|---|---|---|---|
| `wiki/SCHEMA.md` | `<repo>/ai-workflow/wiki/SCHEMA.md` | `render_wiki_schema` | `_INLINE_STUB_SCHEMA` (R1~R7/A1~A4/V-1~V-8/P1~P4 stub table) |
| `wiki/index.md` | `<repo>/ai-workflow/wiki/index.md` | `render_wiki_index` | `_INLINE_STUB_INDEX` (empty Concepts anchor template) |
| `wiki/log.md` | `<repo>/ai-workflow/wiki/log.md` | `render_wiki_log` | `_INLINE_STUB_LOG` (append-only event template) |
| `wiki/.gitignore` | `<repo>/ai-workflow/wiki/.gitignore` | `render_wiki_gitignore` | `_INLINE_STUB_GITIGNORE` (`.ingest_lock`) |

All 4 filenames are enumerated in `WIKI_SOURCE_FILES: tuple[str, ...]`
at module top, so the loader and the writer cannot drift.

## When to Use

- New project bootstrap (`--adoption-mode new` + `--enable-wiki`)
- Harness migration of an existing project to a new harness overlay
- Self-dogfooding the standard-ai-workflow repo itself (see README §6)
- Wheel install smoke tests where the wiki must exist (even as stubs)
  to keep downstream validators from erroring

## When Not to Use

- **Existing wiki present**: `write_wiki_files()` overwrites by default.
  Skip the flag when `ai-workflow/wiki/` already holds non-trivial
  content; use `merge-doc-reconcile` or manual edits instead.
- **Test fixtures**: tests that exercise the wiki layer should mount a
  tmp wiki directory, not depend on bootstrap emit.
- **Wheel-only installs with intent to customize**: the inline stubs
  are placeholders. Commit real source files to the source repo before
  shipping the wheel.
- **Harness-specific wiki content**: there is none — the pattern
  intentionally does not branch per harness. If you need per-harness
  schema, write a sibling overlay pattern.

## Related

- [[entities/bootstrap-wiki-py]] — entity page for the emitter module
- [[decisions/adr-004-wiki-layer]] — ADR that mandates the wiki Runtime layer
- [[concepts/project-architecture]] — R1~R7 / D1~D2 rule set the stubs
  fall back to
- Mirror pattern: `bootstrap_lib.mcp.write_mcp_config_files`
  (`--enable-mcp`)
- V-1~V-8 validators and `check_wiki_index_structure.py` consume the
  emitted skeleton
