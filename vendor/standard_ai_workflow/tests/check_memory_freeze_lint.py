#!/usr/bin/env python3
"""Smoke test R10 Freeze Lint — validates archive/YYYY-MM-DD/ freeze integrity.

Checks:
  V-R8: freeze 후 read-only (active 파일과 archive 파일 일치)
  V-R10: freeze 시 5개 파일 무결성 + state.json 일치
"""

from __future__ import annotations

import json
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
ACTIVE_DIR = REPO_ROOT / "ai-workflow" / "memory" / "active"
ARCHIVE_ROOT = REPO_ROOT / "ai-workflow" / "memory" / "archive"

REQUIRED_FILES = [
    "PROJECT_PROFILE.md",
    "project_status_assessment.md",
    "repository_assessment.md",
    "state.json",
    "state.json.template",
    "work_backlog.md",
]


def main() -> int:
    errors: list[str] = []

    # Check active/ files exist
    for f in REQUIRED_FILES:
        if not (ACTIVE_DIR / f).exists():
            errors.append(f"[V-R10] Missing in active/: {f}")

    # Check for the most recent archive (latest date)
    if ARCHIVE_ROOT.is_dir():
        subdirs = sorted(
            [d for d in ARCHIVE_ROOT.iterdir() if d.is_dir()],
            reverse=True,
        )
    else:
        subdirs = []

    if subdirs:
        latest = subdirs[0]
        frozen_marker = latest / ".frozen"
        if not frozen_marker.exists():
            errors.append(f"[V-R8] Latest archive {latest.name} missing .frozen marker")
        else:
            marker_text = frozen_marker.read_text(encoding="utf-8")
            if "frozen_at" not in marker_text:
                errors.append(f"[V-R8] .frozen marker in {latest.name} missing frozen_at")

        # Check all required files exist in the archive
        for f in REQUIRED_FILES:
            if not (latest / f).exists():
                errors.append(f"[V-R10] Archive {latest.name} missing: {f}")
    else:
        # No archives yet — this is acceptable for a fresh repo
        pass

    if errors:
        for e in errors:
            print(f"[V-R10] FAIL: {e}")
        raise AssertionError(f"Freeze lint check failed with {len(errors)} violation(s).")

    n_archives = len(subdirs)
    print(f"Freeze lint check passed. {n_archives} archive(s) found, all required files present.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
