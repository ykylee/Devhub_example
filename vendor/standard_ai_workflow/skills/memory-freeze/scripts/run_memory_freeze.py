#!/usr/bin/env python3
"""Memory freeze: copy active/ state to archive/YYYY-MM-DD/ with .frozen marker.

R8 (Memory Raw Freeze) implementation.
See SKILL.md for protocol details.
"""

from __future__ import annotations

import argparse
import json
import shutil
import sys
from datetime import date
from pathlib import Path


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Freeze active/ state to archive/")
    parser.add_argument("--active-root", default="ai-workflow/memory/active/",
                        help="Path to the active mutable state directory")
    parser.add_argument("--archive-root", default="ai-workflow/memory/archive/",
                        help="Path to the archive root directory")
    parser.add_argument("--freeze-date", default=date.today().isoformat(),
                        help="Freeze date (YYYY-MM-DD)")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    repo_root = Path.cwd()
    active_dir = (repo_root / args.active_root).resolve()
    archive_root = (repo_root / args.archive_root).resolve()
    freeze_date = args.freeze_date
    archive_dir = archive_root / freeze_date

    # Validate active dir
    if not active_dir.is_dir():
        payload = {
            "status": "error",
            "error": f"Active directory not found: {active_dir}",
            "error_code": "ACTIVE_DIR_MISSING",
        }
        print(json.dumps(payload, ensure_ascii=False, indent=2))
        return 1

    # Skip if already frozen (immutability)
    frozen_marker = archive_dir / ".frozen"
    if frozen_marker.exists():
        payload = {
            "status": "skipped",
            "archive_path": str(archive_dir),
            "reason": f"Already frozen on {freeze_date}",
        }
        print(json.dumps(payload, ensure_ascii=False, indent=2))
        return 0

    # Collect files to freeze
    frozen_files: list[str] = []
    for item in sorted(active_dir.iterdir()):
        if item.is_file() and (item.suffix in (".md", ".json", ".toml", ".txt", ".yaml", ".yml") or item.name.endswith(".template")):
            rel = item.relative_to(active_dir)
            frozen_files.append(str(rel))

    if not frozen_files:
        payload = {
            "status": "error",
            "error": f"No freezeable files in {active_dir}",
            "error_code": "NO_FILES",
        }
        print(json.dumps(payload, ensure_ascii=False, indent=2))
        return 1

    # Create archive dir and copy files
    archive_dir.mkdir(parents=True, exist_ok=True)
    for rel in frozen_files:
        src = active_dir / rel
        dst = archive_dir / rel
        shutil.copy2(src, dst)

    # Write .frozen marker
    frozen_meta = {
        "frozen_at": freeze_date,
        "source": str(active_dir),
        "files": frozen_files,
    }
    frozen_marker.write_text(
        "frozen_at: {frozen_at}\nsource: {source}\nfiles:\n{files}".format(
            frozen_at=frozen_meta["frozen_at"],
            source=frozen_meta["source"],
            files="\n".join(f"  - {f}" for f in frozen_meta["files"]),
        ),
        encoding="utf-8",
    )

    payload = {
        "status": "success",
        "archive_path": str(archive_dir),
        "frozen_files": frozen_files,
        "file_count": len(frozen_files),
    }
        # v0.6.6 follow-up: stage_completion merge (pilot template)
        memory-freeze_completion = build_stage_completion(
            stage_name="memory-freeze",
            stage_status="ok" if result.get("status") in ("ok", "success") else "warning" if result.get("status") == "warning" else "error",
            artifacts=["ai-workflow/memory/archive/<target_date>/"],
            next_stage=None,
            notes=[result.get("summary", "")[:200]] if result.get("summary") else [],
        )
        result = merge_into_result(result, memory-freeze_completion)
    print(json.dumps(payload, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
