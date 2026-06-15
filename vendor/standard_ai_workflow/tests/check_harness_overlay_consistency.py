#!/usr/bin/env python3
"""P4: harness overlay consistency — entry files reference active/ directory + wiki/."""

from __future__ import annotations
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
HARNESSES_DIR = REPO_ROOT / "workflow-source" / "harnesses"

def main() -> int:
    errors: list[str] = []
    harnesses = sorted([d for d in HARNESSES_DIR.iterdir() if d.is_dir() and not d.name.startswith("_")])
    for h in harnesses:
        for md_file in h.rglob("*.md"):
            text = md_file.read_text(encoding="utf-8")
            rel = md_file.relative_to(REPO_ROOT)
            fname = md_file.name
            if fname in ("overlay_spec.md", "AGENTS.md"):
                if "ai-workflow/memory/active/" not in text:
                    errors.append(f"{rel}: missing memory/active/ reference")
                if "ai-workflow/wiki/index.md" not in text:
                    errors.append(f"{rel}: missing wiki/index.md reference")
    if errors:
        for e in errors:
            print(f"FAIL: {e}")
        raise AssertionError(f"{len(errors)} violation(s).")
    print(f"Harness overlay check passed ({len(harnesses)} harnesses).")
    return 0

if __name__ == "__main__":
    raise SystemExit(main())
