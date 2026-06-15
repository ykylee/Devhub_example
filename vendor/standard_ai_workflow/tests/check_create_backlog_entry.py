#!/usr/bin/env python3
"""Smoke test the create_backlog_entry MCP prototype."""

from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
SCRIPT_PATH = SOURCE_ROOT / "mcp_servers" / "create-backlog-entry" / "scripts" / "run_create_backlog_entry.py"

if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.common.output_contracts import validate_output_payload


def main() -> int:
    completed = subprocess.run(
        [
            sys.executable,
            str(SCRIPT_PATH),
            "--task-id",
            "TASK-099",
            "--task-name",
            "출력 샘플 정리",
            "--request-date",
            "2026-04-20",
        ],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
        check=True,
    )
    payload = json.loads(completed.stdout)
    output_errors = validate_output_payload(payload, family="create_backlog_entry")
    if output_errors:
        raise AssertionError(f"create_backlog_entry payload violated output contract: {output_errors}")
    if not payload["draft_entry"]:
        raise AssertionError("Expected non-empty draft_entry lines.")
    if not payload["draft_entry"][0].startswith("## TASK-099"):
        raise AssertionError("Expected backlog heading to include the provided task id.")

    print("Create-backlog-entry smoke check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
