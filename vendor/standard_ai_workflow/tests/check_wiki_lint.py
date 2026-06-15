#!/usr/bin/env python3
"""Smoke test the wiki lint skill prototype (umbrella: V-1 + V-4)."""

from __future__ import annotations

import sys
from pathlib import Path
from types import ModuleType


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"

# Make sibling check_wiki_* modules importable when this script is launched
# directly (e.g. ``python3 workflow-source/tests/check_wiki_lint.py``).
_TESTS_DIR = Path(__file__).resolve().parent
if str(_TESTS_DIR) not in sys.path:
    sys.path.insert(0, str(_TESTS_DIR))

import check_wiki_index_structure  # noqa: E402  (sys.path tweak above)
import check_wiki_location  # noqa: E402


Check = ModuleType


def _run(name: str, module: Check) -> tuple[bool, str]:
    """Run ``module.main()`` and return (passed, label)."""
    try:
        rc = module.main()
    except AssertionError as exc:
        print(f"[umbrella] {name} FAILED: {exc}")
        return False, name
    except Exception as exc:  # noqa: BLE001
        print(f"[umbrella] {name} raised unexpected {type(exc).__name__}: {exc}")
        return False, name
    if rc != 0:
        print(f"[umbrella] {name} returned non-zero exit ({rc}).")
        return False, name
    return True, name


def main() -> int:
    results: list[tuple[str, bool]] = []
    for label, mod in (
        ("V-1 (wiki location)", check_wiki_location),
        ("V-4 (index structure)", check_wiki_index_structure),
    ):
        ok, _ = _run(label, mod)
        results.append((label, ok))

    passed = [label for label, ok in results if ok]
    failed = [label for label, ok in results if not ok]

    if not failed:
        print(
            "Wiki lint smoke check passed "
            f"(V-1, V-4). Validators: {', '.join(passed)}."
        )
        return 0

    print("Wiki lint smoke check FAILED:")
    for label in failed:
        print(f"  - {label}")
    return 1


if __name__ == "__main__":
    raise SystemExit(main())
