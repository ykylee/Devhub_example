#!/usr/bin/env python3
"""Set up .gitignore for the standard AI workflow kit.

This is a thin CLI wrapper around ``workflow_kit.gitignore``.
All pattern definitions and idempotent update logic live in the shared
library so that ``apply_workflow_upgrade.py`` uses the same code path.
"""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

# Ensure workflow_kit is importable
REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.gitignore import ensure_gitignore_patterns


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Set up .gitignore for the standard AI workflow."
    )
    parser.add_argument(
        "--target-root", default=".", help="Target repository root."
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Show changes without applying them.",
    )
    args = parser.parse_args()

    target_root = Path(args.target_root).resolve()
    print(f"--- Setting up .gitignore in {target_root} ---")

    changes = ensure_gitignore_patterns(target_root, args.dry_run)

    if not changes:
        print("No changes needed. .gitignore is already up to date.")
    else:
        for change in changes:
            print(f"- {change}")
        if args.dry_run:
            print("\n[DRY RUN] No changes were written.")
        else:
            print("\n.gitignore updated successfully.")


if __name__ == "__main__":
    main()
