#!/usr/bin/env python3
"""Smoke test T3 ingest session atomicity + T2 work_backlog anchor structure.

T3: ingest_session_atomic() lock/atomicity.
T2: work_backlog anchor format (### [[path]] {#id}).
"""

from __future__ import annotations

import json
import sys
import tempfile
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.common.ingest import ingest_session_atomic, IngestLockedError


def test_atomic_write() -> list[str]:
    errors: list[str] = []
    with tempfile.TemporaryDirectory() as tmp:
        active_dir = Path(tmp) / "active"
        active_dir.mkdir()

        result = ingest_session_atomic(
            memory_active=active_dir,
            handoff_content="# Session Handoff\ntest content",
            backlog_content="# 2026-06-12 Backlog\ntest content",
            state_content={"session": {"active": True}},
            backlog_index_content="# Work Backlog Index\ntest",
        )

        if "handoff" not in result:
            errors.append("[T3] handoff not in result")
        if "backlog" not in result:
            errors.append("[T3] backlog not in result")
        if "state" not in result:
            errors.append("[T3] state not in result")
        if "worklog" not in result:
            errors.append("[T3] worklog not in result")

        # Verify files exist
        if not (active_dir / "session_handoff.md").exists():
            errors.append("[T3] session_handoff.md not written")
        if not (active_dir / "backlog" / "2026-06-12.md").exists():
            errors.append("[T3] daily backlog not written")

        # Verify state.json content
        state = json.loads((active_dir / "state.json").read_text())
        if not state.get("session", {}).get("active"):
            errors.append("[T3] state.json content mismatch")

        # Verify lock released
        if (active_dir / ".ingest_lock").exists():
            errors.append("[T3] lock not released after ingest")

    return errors


def test_lock_contention() -> list[str]:
    errors: list[str] = []
    with tempfile.TemporaryDirectory() as tmp:
        active_dir = Path(tmp) / "active"
        active_dir.mkdir()
        (active_dir / ".ingest_lock").touch()  # simulate concurrent ingest

        try:
            ingest_session_atomic(memory_active=active_dir)
            errors.append("[T3] should have raised IngestLockedError")
        except IngestLockedError:
            pass  # expected

    return errors


def test_worklog_anchor_structure() -> list[str]:
    errors: list[str] = []
    worklog_path = REPO_ROOT / "ai-workflow" / "memory" / "active" / "work_backlog.md"
    if not worklog_path.exists():
        return errors

    text = worklog_path.read_text(encoding="utf-8")
    lines = text.splitlines()
    for i, line in enumerate(lines, 1):
        stripped = line.strip()
        # Check that every ### has [[path]] {#anchor} format
        if stripped.startswith("### ") and "최근 작업" not in stripped:
            if "[[" in stripped:
                # This is an entry line — must have [[path]] AND {#id}
                if "{#" not in stripped:
                    errors.append(f"[T2] Line {i}: ### entry missing {{#anchor-id}} format: {stripped}")
            else:
                # This is a non-entry divider header — must still have {#id}
                if "{#" not in stripped:
                    errors.append(f"[T2] Line {i}: ### header missing {{#anchor-id}}: {stripped}")

    return errors


def main() -> int:
    all_errors: list[str] = []
    all_errors.extend(test_atomic_write())
    all_errors.extend(test_lock_contention())
    all_errors.extend(test_worklog_anchor_structure())

    if all_errors:
        for e in all_errors:
            print(f"FAIL: {e}")
        raise AssertionError(f"T2+T3 smoke check failed with {len(all_errors)} violation(s).")
    print("T2+T3 smoke check passed (atomic write, lock contention, work_log anchor structure).")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
