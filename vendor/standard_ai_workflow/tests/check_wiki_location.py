#!/usr/bin/env python3
"""Smoke test the wiki location validator (V-1, R1)."""

from __future__ import annotations

import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"

CANONICAL_WIKI_ROOT = REPO_ROOT / "ai-workflow" / "wiki"
CANONICAL_INDEX = CANONICAL_WIKI_ROOT / "index.md"

# R1 explicitly forbids these alternative locations. Even a single file inside
# any of them is treated as a violation because the rule is "location 단일성".
FORBIDDEN_WIKI_LOCATIONS: tuple[tuple[str, Path], ...] = (
    ("docs/wiki/", REPO_ROOT / "docs" / "wiki"),
    (".wiki/", REPO_ROOT / ".wiki"),
    ("workflow-source/wiki/", SOURCE_ROOT / "wiki"),
    ("root wiki/", REPO_ROOT / "wiki"),
)


def _find_first_file(root: Path) -> Path | None:
    """Return the first file under ``root`` or ``None`` if absent.

    ``rglob('*')`` would descend into nested directories, but we only need
    to know whether *any* file lives under the forbidden location. A single
    match is enough to fail the rule.
    """
    if not root.exists():
        return None
    for entry in sorted(root.rglob("*")):
        if entry.is_file():
            return entry
    return None


def _check_canonical_location() -> list[str]:
    errors: list[str] = []
    if not CANONICAL_WIKI_ROOT.exists():
        errors.append(
            f"Canonical wiki directory missing: {CANONICAL_WIKI_ROOT} "
            "(R1 requires ai-workflow/wiki/)"
        )
    if not CANONICAL_INDEX.exists():
        errors.append(
            f"Canonical wiki index missing: {CANONICAL_INDEX} "
            "(R1 requires ai-workflow/wiki/index.md)"
        )
    return errors


def _check_no_duplicates() -> list[str]:
    errors: list[str] = []
    for label, path in FORBIDDEN_WIKI_LOCATIONS:
        offender = _find_first_file(path)
        if offender is not None:
            errors.append(
                f"Forbidden wiki location '{label}' exists "
                f"(found: {offender.relative_to(REPO_ROOT)}). "
                "R1 forbids all locations other than ai-workflow/wiki/."
            )
    return errors


def main() -> int:
    errors = _check_canonical_location() + _check_no_duplicates()
    if errors:
        for line in errors:
            print(f"[V-1] FAIL: {line}")
        raise AssertionError(
            f"Wiki location check failed with {len(errors)} violation(s)."
        )
    print("Wiki location check passed (V-1).")
    return 0


if __name__ == "__main__":
    sys.exit(main())
