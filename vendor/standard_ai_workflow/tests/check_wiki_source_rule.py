#!/usr/bin/env python3
"""Check V-R9: wiki-ingest source = archive/ only, NOT active/.

R9 (Raw Source of Truth) — wiki-ingest/memory-ingest must source from
archive/YYYY-MM-DD/, never from active/. This prevents the agent from
ingesting mutable/frozen-incomplete state into the wiki.

v0.1.x naive grep 의 한계: R9 rule 자체를 정의/설명하는 page (R9 spec 자체,
R8 freeze protocol, memory 3-state lifecycle 등) 가 본문에 의도적으로
`active/` mention → 16-17 false-positive.

v0.2.0 보강: frontmatter `r9_skip: true` 가 있으면 V-R9 skip. body 미수정
(wiki 운영 R-4 정책 — frontmatter only marker).
"""

from __future__ import annotations

import re
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
WIKI_DIR = REPO_ROOT / "ai-workflow" / "wiki"
MEMORY_ACTIVE = REPO_ROOT / "ai-workflow" / "memory" / "active"
MEMORY_ARCHIVE = REPO_ROOT / "ai-workflow" / "memory" / "archive"


def has_r9_skip_marker(text: str) -> bool:
    """frontmatter 의 r9_skip: true/1/yes 면 True.

    R-4 정책 (frontmatter only) 따라 r9_skip marker 는 YAML frontmatter 내.
    body 의 `r9_skip:` 언급은 무시 (frontmatter 시작 --- 부재 시).
    """
    if not text.startswith("---\n"):
        return False
    end = text.find("\n---\n", 4)
    if end < 0:
        return False
    fm = text[4:end]
    m = re.search(r"^r9_skip:\s*(true|1|yes)\s*$", fm, re.MULTILINE | re.IGNORECASE)
    return m is not None


def check_wiki_specs_no_active_ref() -> list[str]:
    """Check that no wiki spec/docs reference active/ as ingest source."""
    errors: list[str] = []
    if not WIKI_DIR.is_dir():
        return errors
    for md_file in WIKI_DIR.rglob("*.md"):
        text = md_file.read_text(encoding="utf-8")
        if has_r9_skip_marker(text):
            continue  # R-4 정책: frontmatter r9_skip: true 면 skip
        # Look for patterns like "ingest from active/" or "source=active/"
        if re.search(r"ingest.*active[/\s]|active/.*ingest|source.*active/", text, re.IGNORECASE):
            rel = md_file.relative_to(REPO_ROOT)
            errors.append(f"[V-R9] {rel}: references active/ as ingest source (must use archive/ only)")
    return errors


def check_active_files_not_in_wiki() -> list[str]:
    """Check that no wiki page uses memory/active/ as an ingest source reference."""
    errors: list[str] = []
    if not WIKI_DIR.is_dir() or not MEMORY_ACTIVE.is_dir():
        return errors
    for wiki_file in WIKI_DIR.rglob("*.md"):
        text = wiki_file.read_text(encoding="utf-8")
        if has_r9_skip_marker(text):
            continue
        # Only flag when active/ appears in ingest or source context
        if re.search(r"ingest.*memory/active|memory/active.*ingest|source.*memory/active|memory/active.*source", text, re.IGNORECASE):
            rel = wiki_file.relative_to(REPO_ROOT)
            errors.append(f"[V-R9] {rel}: references memory/active/ as ingest source (must use archive/ only)")
    return errors


def main() -> int:
    errors: list[str] = []
    errors.extend(check_wiki_specs_no_active_ref())
    errors.extend(check_active_files_not_in_wiki())

    if errors:
        for e in errors:
            print(f"FAIL: {e}")
        raise AssertionError(f"V-R9 check failed with {len(errors)} violation(s).")
    print("V-R9 check passed: no active/ references as wiki-ingest source.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
