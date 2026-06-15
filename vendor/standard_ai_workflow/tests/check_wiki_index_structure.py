#!/usr/bin/env python3
"""Smoke test the wiki index structure validator (V-4, R4)."""

from __future__ import annotations

import re
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
CANONICAL_WIKI_ROOT = REPO_ROOT / "ai-workflow" / "wiki"
CANONICAL_INDEX = CANONICAL_WIKI_ROOT / "index.md"

# R4 schema: every index entry MUST match
#     ### [[<path>]] {#<anchor-id>}
ENTRY_RE = re.compile(r"^### \[\[([^\]]+)\]\] \{#([^}]+)\}$")
H1_RE = re.compile(r"^# (.+)$")
H2_RE = re.compile(r"^## (.+)$")
H3_PREFIX_RE = re.compile(r"^### ")


def _validate(lines: list[str], wiki_root: Path) -> tuple[list[str], int]:
    errors: list[str] = []
    if not lines or not H1_RE.match(lines[0]):
        first = repr(lines[0]) if lines else "<empty file>"
        errors.append(f"First line must be a top-level '# Title' heading, got: {first}")
        return errors, 0

    seen: dict[str, int] = {}
    section: str | None = None
    section_has_entry = False
    saw_h2 = False
    total = 0

    for lineno, raw in enumerate(lines, start=1):
        line = raw.rstrip()
        if not line:
            continue
        if H1_RE.match(line):
            if lineno != 1:
                errors.append(
                    f"Line {lineno}: extra '#' top-level heading '{line}' "
                    "(R4 allows a single H1 per index.md)."
                )
            continue
        if H2_RE.match(line):
            if section is not None and not section_has_entry:
                errors.append(
                    f"Line {lineno}: ## section '{section}' has no ### entries "
                    "(R4 forbids empty sections)."
                )
            section = H2_RE.match(line).group(1).strip()
            section_has_entry = False
            saw_h2 = True
            continue
        if H3_PREFIX_RE.match(line):
            match = ENTRY_RE.match(line)
            if not match:
                errors.append(
                    f"Line {lineno}: malformed ### entry '{line}' "
                    "(R4 requires '### [[<path>]] {#<anchor-id>}' format)."
                )
                continue
            path, anchor = match.group(1).strip(), match.group(2).strip()
            if not path:
                errors.append(f"Line {lineno}: empty path in entry '{line}'.")
                continue
            if not anchor:
                errors.append(f"Line {lineno}: empty anchor-id in entry '{line}'.")
                continue
            if anchor in seen:
                errors.append(
                    f"Line {lineno}: duplicate anchor-id '{anchor}' "
                    f"(first seen on line {seen[anchor]})."
                )
            else:
                seen[anchor] = lineno
            target = wiki_root / f"{path}.md"
            if not target.exists():
                errors.append(
                    f"Line {lineno}: entry path '{path}' does not resolve to a file "
                    f"(expected: {target})."
                )
            section_has_entry = True
            total += 1
            continue
        if section is None and not saw_h2:
            errors.append(
                f"Line {lineno}: orphan content '{line[:60]}' outside any '##' section "
                "(R4 requires anchor-based structure)."
            )

    if section is not None and not section_has_entry:
        errors.append(
            f"## section '{section}' has no ### entries (R4 forbids empty sections)."
        )
    if not saw_h2:
        errors.append("index.md has no '##' sections (R4 requires at least one).")
    return errors, total


def main() -> int:
    if not CANONICAL_INDEX.exists():
        print(f"[V-4] FAIL: canonical index missing: {CANONICAL_INDEX.relative_to(REPO_ROOT)}")
        raise AssertionError("Wiki index structure check cannot run: index.md not found.")

    lines = CANONICAL_INDEX.read_text(encoding="utf-8").splitlines()
    errors, total = _validate(lines, CANONICAL_WIKI_ROOT)
    if errors:
        for line in errors:
            print(f"[V-4] FAIL: {line}")
        raise AssertionError(
            f"Wiki index structure check failed with {len(errors)} violation(s)."
        )
    print(f"Wiki index structure check passed (V-4, {total} entries validated).")
    return 0


if __name__ == "__main__":
    sys.exit(main())
