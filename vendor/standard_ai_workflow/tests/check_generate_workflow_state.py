#!/usr/bin/env python3
"""Smoke test the workflow state cache generator."""

from __future__ import annotations

import json
import subprocess
import sys
import tempfile
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
SCRIPT_PATH = SOURCE_ROOT / "scripts" / "generate_workflow_state.py"


def main() -> int:
    example_root = SOURCE_ROOT / "examples" / "acme_delivery_platform"
    latest_backlog = sorted((example_root / "backlog").glob("*.md"))[-1]

    with tempfile.TemporaryDirectory() as tmpdir:
        output_path = Path(tmpdir) / "state.json"
        completed = subprocess.run(
            [
                sys.executable,
                str(SCRIPT_PATH),
                "--project-profile-path",
                str(example_root / "PROJECT_PROFILE.md"),
                "--session-handoff-path",
                str(example_root / "session_handoff.md"),
                "--work-backlog-index-path",
                str(example_root / "work_backlog.md"),
                "--latest-backlog-path",
                str(latest_backlog),
                "--output-path",
                str(output_path),
                "--generated-at",
                "2026-04-24",
            ],
            cwd=REPO_ROOT,
            capture_output=True,
            text=True,
            check=True,
        )
        payload = json.loads(completed.stdout)
        if payload["status"] != "ok":
            raise AssertionError("Expected ok status from workflow state generator.")
        state = json.loads(output_path.read_text(encoding="utf-8"))
        if state["schema_version"] != "1":
            raise AssertionError("Expected workflow state schema version 1.")
        if state["generated_at"] != "2026-04-24":
            raise AssertionError("Expected workflow state generated_at to preserve the provided value.")
        if not state["session"]["current_focus"]:
            raise AssertionError("Expected workflow state current_focus to be populated.")
        if not state["next_documents"]:
            raise AssertionError("Expected workflow state next_documents to be populated.")

    print("Workflow state generator smoke check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
