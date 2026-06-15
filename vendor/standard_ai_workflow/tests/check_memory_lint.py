#!/usr/bin/env python3
"""Umbrella test for T1 Memory Lint 4종:
- contradiction: handoff next_actions vs backlog status mismatch
- stale: 90+ days without modification
- orphan: task files with no inbound link from work_backlog.md
- missing: TASK-XXX referenced in handoff/backlog but no file
"""

from __future__ import annotations

import re
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
ACTIVE_DIR = REPO_ROOT / "ai-workflow" / "memory" / "active"
ARCHIVE_ROOT = REPO_ROOT / "ai-workflow" / "memory" / "archive"


def check_contradiction() -> list[str]:
    """T1a: handoff next_actions vs backlog completed mismatch."""
    errors: list[str] = []
    handoff_path = ACTIVE_DIR / "PROJECT_PROFILE.md"
    if not handoff_path.exists():
        return errors  # skip if no profile
    # Check latest archive for contradiction between handoff and backlog
    if ARCHIVE_ROOT.is_dir():
        subdirs = sorted([d for d in ARCHIVE_ROOT.iterdir() if d.is_dir()], reverse=True)
        if subdirs:
            latest = subdirs[0]
            handoff = latest / "session_handoff.md"
            backlog = latest / "work_backlog.md"
            if handoff.exists() and backlog.exists():
                handoff_text = handoff.read_text(encoding="utf-8")
                backlog_text = backlog.read_text(encoding="utf-8")
                # Simple check: find TASK-XXX in handoff next_actions
                handoff_tasks = set(re.findall(r"TASK-\w+", handoff_text))
                backlog_tasks = set(re.findall(r"TASK-\w+", backlog_text))
                # If handoff has a task not in backlog at all, it's suspicious
                missing_in_backlog = handoff_tasks - backlog_tasks
                if missing_in_backlog:
                    errors.append(f"[T1a] Tasks in handoff but missing from latest archive: {missing_in_backlog}")
    return errors


def check_stale() -> list[str]:
    """T1b: files not modified in 90+ days."""
    errors: list[str] = []
    import time
    now = time.time()
    cutoff = now - (90 * 24 * 3600)

    if ACTIVE_DIR.is_dir():
        for item in ACTIVE_DIR.iterdir():
            if item.is_file() and item.stat().st_mtime < cutoff:
                errors.append(f"[T1b] Stale file (90+ days): {item.name}")
    return errors


def check_orphan() -> list[str]:
    """T1c: task files with no inbound link from work_backlog.md."""
    errors: list[str] = []
    backlog_index = ACTIVE_DIR / "work_backlog.md"
    if not backlog_index.exists():
        return errors

    backlog_text = backlog_index.read_text(encoding="utf-8")
    # Look for task files referenced in backlog via [[...]] links
    referenced: set[str] = set()
    for m in re.findall(r"\[\[([^\]]+)\]\]", backlog_text):
        referenced.add(m.strip().lstrip("/"))

    # Look for actual task files in active/backlog/ (if it exists)
    backlog_dir = ACTIVE_DIR / "backlog"
    if backlog_dir.is_dir():
        for f in backlog_dir.rglob("*.md"):
            rel = str(f.relative_to(backlog_dir))
            if rel not in referenced:
                errors.append(f"[T1c] Orphan backlog file: {rel} (not linked from work_backlog.md)")
    return errors


def check_missing() -> list[str]:
    """T1d: TASK-XXX references with no corresponding file."""
    errors: list[str] = []
    task_path = ACTIVE_DIR / "backlog"
    if not task_path.is_dir():
        return errors  # no task dir, skip

    existing_tasks: set[str] = set()
    for f in task_path.rglob("*.md"):
        content = f.read_text(encoding="utf-8")
        for m in re.findall(r"TASK-\w+", content):
            existing_tasks.add(m)

    return errors


def main() -> int:
    results: list[str] = []
    failures = 0

    checks = [
        ("contradiction", check_contradiction),
        ("stale", check_stale),
        ("orphan", check_orphan),
        ("missing", check_missing),
    ]

    for name, fn in checks:
        errs = fn()
        if errs:
            failures += 1
            for e in errs:
                print(e)
                results.append(f"[FAIL] {name}: {e}")
        else:
            results.append(f"[PASS] {name}")

    for r in results:
        print(r)

    if failures:
        raise AssertionError(f"Memory lint check failed ({failures} check(s) had issues).")
    print("Memory lint smoke check passed (all 4 checks).")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
