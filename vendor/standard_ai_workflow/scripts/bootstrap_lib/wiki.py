"""Wiki layer renderer for the generated kit.

The bootstrap script can optionally emit an LLM Wiki layer
(``--enable-wiki``). Each generated project receives the same wiki
skeleton (SCHEMA.md, index.md, log.md, .gitignore) so downstream agents
have a stable ``ai-workflow/wiki/`` entry point. Mirrors
:mod:`bootstrap_lib.mcp` and the ``--enable-mcp`` pattern.
"""

from __future__ import annotations

import argparse
from pathlib import Path
from typing import Callable

from bootstrap_lib.paths import Paths
from bootstrap_lib.writes import write_text


# Source loading — the wiki skeleton lives at <repo>/ai-workflow/wiki/ in
# the standard-ai-workflow source repo. We read those files at emit time
# and write them under <target>/ai-workflow/wiki/. When the source files
# are missing (e.g. a wheel install without bundled data) the bootstrap
# falls back to inline minimal stubs.

WIKI_SOURCE_FILES: tuple[str, ...] = ("SCHEMA.md", "index.md", "log.md", ".gitignore")


def _wiki_source_root() -> Path | None:
    """Return the directory containing the wiki/ source files, if any."""
    # Local import to avoid circular import (__main__ imports wiki).
    from bootstrap_lib import __main__ as _bm

    repo_root = _bm.REPO_ROOT
    if repo_root is None:
        return None
    candidate = repo_root / "ai-workflow" / "wiki"
    return candidate if candidate.is_dir() else None


def _load_wiki_source(filename: str) -> str | None:
    """Return the contents of ``ai-workflow/wiki/<filename>`` from the source repo, or None."""
    src_root = _wiki_source_root()
    if src_root is None:
        return None
    path = src_root / filename
    if not path.is_file():
        return None
    try:
        return path.read_text(encoding="utf-8")
    except OSError:
        return None


# Inline minimal stubs — used when the wiki source is not bundled (e.g. a
# wheel install). Real projects should commit the source-of-truth wiki
# content to the standard-ai-workflow repo so the bootstrap can copy
# it verbatim; the stubs only keep the wiki layer from being empty.

_INLINE_STUB_SCHEMA = """# Wiki 운영 헌법 (Operating Constitution)

> STUB: source files not bundled (wheel install?). Replace with the real
> ``ai-workflow/wiki/SCHEMA.md`` from the standard-ai-workflow source
# repo, or write your own R1~R7 / A1~A4 / V-1~V-8 / P1~P4 set per
# ``.omo/plans/v0.5.11-plus-llm-wiki-distributed-rules.md``.

| # | 항목 | 값 |
|---|---|---|
| 1 | 위치 | `ai-workflow/wiki/` (Runtime layer, R1) |
| 2 | 추적 | git 추적 (memory/ 와 분리, D2) |
| 3 | 페이지 타입 | 5종 (entities, concepts, decisions, patterns, queries) |
| 4 | primary record | 페이지 atomic (R2, 1 commit = 1 ingest) |
| 5 | 인덱스 | anchor 기반 (R4) |
| 6 | 머지 | additive (R5) + wiki 확장 (R7) |
| 7 | lint | 모순·스테일·고아·누락·깨진 backlink (5항목) |
"""

_INLINE_STUB_INDEX = """# Master Knowledge Index

## Concepts

<!-- 신규 concept 페이지를 위 한 줄 + 한 줄 설명으로 추가한다. -->
<!-- 형식: ### [[concepts/<slug>]] {#<slug>} -->
"""

_INLINE_STUB_LOG = """# Wiki Ingest/Query Log

- 문서 목적: 모든 ingest / query / lint 이벤트의 append-only 작업 로그. 시간 순 보존, 편집 금지 (append only).
- 갱신 규칙: ingest 종료 시 또는 query/lint 실행 시 한 줄 추가. 형식 `## [YYYY-MM-DD] <event> | <summary>`.
- 최종 갱신일: YYYY-MM-DD

## [YYYY-MM-DD] bootstrap | wiki layer P1 prototype (stub)

- SCHEMA.md, index.md, log.md seeded from stub (source files not bundled)
- target: ai-workflow/wiki/ Runtime layer (R1, D1, D2)
"""

_INLINE_STUB_GITIGNORE = """.ingest_lock
"""


# Renderers — one per wiki file. Pure functions, no side effects.


def render_wiki_schema(args: argparse.Namespace, paths: Paths) -> str:
    """Return the content for ``ai-workflow/wiki/SCHEMA.md``."""
    return _load_wiki_source("SCHEMA.md") or _INLINE_STUB_SCHEMA


def render_wiki_index(args: argparse.Namespace, paths: Paths) -> str:
    """Return the content for ``ai-workflow/wiki/index.md``."""
    return _load_wiki_source("index.md") or _INLINE_STUB_INDEX


def render_wiki_log(args: argparse.Namespace, paths: Paths) -> str:
    """Return the content for ``ai-workflow/wiki/log.md``."""
    return _load_wiki_source("log.md") or _INLINE_STUB_LOG


def render_wiki_gitignore(args: argparse.Namespace, paths: Paths) -> str:
    """Return the content for ``ai-workflow/wiki/.gitignore``."""
    return _load_wiki_source(".gitignore") or _INLINE_STUB_GITIGNORE


# Writer — emits the wiki/ directory to <target>/ai-workflow/wiki/ when
# ``--enable-wiki`` is set. Mirrors :func:`bootstrap_lib.mcp.write_mcp_config_files`.


def render_wiki_overlay(
    args: argparse.Namespace,
    paths: Paths,
    harnesses: list[str],
) -> dict[str, str]:
    """Return a {overlay_key: file_path} map for the wiki layer (no writes).

    The wiki layer is harness-agnostic: every harness overlay pointing
    at the same target gets the same wiki skeleton.
    """
    wiki_root = paths.kit_root / "wiki"
    return {
        "wiki_schema": str(wiki_root / "SCHEMA.md"),
        "wiki_index": str(wiki_root / "index.md"),
        "wiki_log": str(wiki_root / "log.md"),
        "wiki_gitignore": str(wiki_root / ".gitignore"),
    }


def write_wiki_files(
    args: argparse.Namespace,
    paths: Paths,
    harnesses: list[str],
) -> dict[str, str]:
    """Emit the wiki/ directory when ``--enable-wiki`` is set."""
    overlay = render_wiki_overlay(args, paths, harnesses)
    wiki_root = paths.kit_root / "wiki"
    wiki_root.mkdir(parents=True, exist_ok=True)

    write_text(
        Path(overlay["wiki_schema"]),
        render_wiki_schema(args, paths),
        force=args.force,
        rel_to=paths.target_root,
    )
    write_text(
        Path(overlay["wiki_index"]),
        render_wiki_index(args, paths),
        force=args.force,
        rel_to=paths.target_root,
    )
    write_text(
        Path(overlay["wiki_log"]),
        render_wiki_log(args, paths),
        force=args.force,
        rel_to=paths.target_root,
    )
    write_text(
        Path(overlay["wiki_gitignore"]),
        render_wiki_gitignore(args, paths),
        force=args.force,
        rel_to=paths.target_root,
    )
    return overlay


#: Dispatch table from harness name to its wiki renderer. The wiki layer
#: is currently harness-agnostic (single skeleton for every project), so
#: the dispatch points every harness at the same shared schema renderer.
WIKI_CONFIG_RENDERERS: dict[str, Callable[[argparse.Namespace, Paths], str]] = {
    "codex": render_wiki_schema,
    "opencode": render_wiki_schema,
    "gemini-cli": render_wiki_schema,
    "antigravity": render_wiki_schema,
    "minimax-code": render_wiki_schema,
    "pi-dev": render_wiki_schema,
}


__all__ = [
    "WIKI_CONFIG_RENDERERS",
    "WIKI_SOURCE_FILES",
    "render_wiki_gitignore",
    "render_wiki_index",
    "render_wiki_log",
    "render_wiki_overlay",
    "render_wiki_schema",
    "write_wiki_files",
]

