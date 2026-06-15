#!/usr/bin/env python3
"""Smoke test the session-start prototype."""

from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
SCRIPT_PATH = SOURCE_ROOT / "skills" / "session-start" / "scripts" / "run_session_start.py"

if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.common.output_contracts import validate_output_payload


def run_session_start(*, expect_success: bool, args: list[str]) -> tuple[int, dict[str, object]]:
    completed = subprocess.run(
        [sys.executable, str(SCRIPT_PATH), *args],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if expect_success and completed.returncode != 0:
        raise AssertionError(f"Expected session-start success but got {completed.returncode}: {completed.stderr}")
    if not expect_success and completed.returncode == 0:
        raise AssertionError("Expected session-start failure path but command succeeded.")
    return completed.returncode, json.loads(completed.stdout)


def main() -> int:
    example_root = SOURCE_ROOT / "examples" / "acme_delivery_platform"
    latest_backlog = sorted((example_root / "backlog").glob("*.md"))[-1]

    _, payload = run_session_start(
        expect_success=True,
        args=[
            "--session-handoff-path",
            str(example_root / "session_handoff.md"),
            "--work-backlog-index-path",
            str(example_root / "work_backlog.md"),
            "--project-profile-path",
            str(example_root / "PROJECT_PROFILE.md"),
            "--latest-backlog-path",
            str(latest_backlog),
        ],
    )
    output_errors = validate_output_payload(payload, family="session_start")
    if output_errors:
        raise AssertionError(f"Session-start success payload violated output contract: {output_errors}")
    if payload["status"] != "ok":
        raise AssertionError("Expected ok status for session-start success path.")
    if payload["latest_backlog_path"] != str(latest_backlog):
        raise AssertionError("Expected latest_backlog_path to preserve the explicit backlog path.")
    if not payload["summary"]:
        raise AssertionError("Expected non-empty session summary.")
    if not payload["recommended_next_action"]:
        raise AssertionError("Expected recommended next action.")
    if not payload["next_documents"]:
        raise AssertionError("Expected next_documents to include review targets.")

    failure_code, failure_payload = run_session_start(
        expect_success=False,
        args=[
            "--session-handoff-path",
            "/tmp/missing-session-handoff.md",
            "--work-backlog-index-path",
            str(example_root / "work_backlog.md"),
            "--project-profile-path",
            str(example_root / "PROJECT_PROFILE.md"),
        ],
    )
    if failure_code == 0:
        raise AssertionError("Expected session-start failure for missing required document.")
    output_errors = validate_output_payload(failure_payload, family="session_start")
    if output_errors:
        raise AssertionError(f"Session-start error payload violated output contract: {output_errors}")
    if failure_payload["error_code"] != "missing_required_document":
        raise AssertionError("Expected missing_required_document error code.")

    print("Session-start smoke check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
