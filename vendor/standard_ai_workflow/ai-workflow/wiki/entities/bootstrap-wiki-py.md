---
type: entity
status: active
last_ingested_from: workflow-source/scripts/bootstrap_lib/wiki.py
related_pages: [concepts/project-architecture, decisions/adr-004-wiki-layer]
---

# bootstrap_lib wiki.py

## Role

Emits the `ai-workflow/wiki/` skeleton when `--enable-wiki` flag is used. Modeled on `mcp.py` (--enable-mcp pattern).

## Interface

```
write_wiki_files(args, paths, harnesses) -> dict[str, str]
```

Returns `{wiki_schema: path, wiki_index: path, wiki_log: path, wiki_gitignore: path}`.

## Emitted Files

| File | Content Source |
|---|---|
| `wiki/SCHEMA.md` | Copy from `ai-workflow/wiki/SCHEMA.md` |
| `wiki/index.md` | Copy from `ai-workflow/wiki/index.md` |
| `wiki/log.md` | Copy from `ai-workflow/wiki/log.md` |
| `wiki/.gitignore` | Static content: `.ingest_lock` |

## Location

`workflow-source/scripts/bootstrap_lib/wiki.py` (208 lines)
