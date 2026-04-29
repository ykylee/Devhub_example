#!/usr/bin/env python3
"""Smoke test the check_quickstart_stale_links MCP prototype."""

from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[1]
SCRIPT_PATH = REPO_ROOT / "mcp" / "check-quickstart-stale-links" / "scripts" / "run_check_quickstart_stale_links.py"


def main() -> int:
    completed = subprocess.run(
        [
            sys.executable,
            str(SCRIPT_PATH),
            "--quickstart-path",
            str(REPO_ROOT / "examples" / "bootstrap_output_samples.md"),
            "--project-profile-path",
            str(REPO_ROOT / "templates" / "project_workflow_profile_template.md"),
            "--session-handoff-path",
            str(REPO_ROOT / "examples" / "README.md"),
        ],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
        check=True,
    )
    payload = json.loads(completed.stdout)
    if not payload["checked_files"]:
        raise AssertionError("Expected checked quickstart files.")
    if payload["broken_links"]:
        raise AssertionError("Expected no broken links in bootstrap_output_samples.md.")
    if not payload["missing_expected_links"]:
        raise AssertionError("Expected at least one missing expected link warning.")
    if not payload["stale_link_warnings"]:
        raise AssertionError("Expected stale link warnings.")
    print("Quickstart stale link smoke check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
