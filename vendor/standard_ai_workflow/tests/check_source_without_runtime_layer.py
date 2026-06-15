#!/usr/bin/env python3
"""Verify source development checks do not require the applied ai-workflow runtime layer."""

from __future__ import annotations

import subprocess
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
RUNTIME_DIR = REPO_ROOT / "ai-workflow"
HIDDEN_RUNTIME_DIR = REPO_ROOT / ".tmp-ai-workflow-runtime-hidden"


def run_check(script_name: str) -> None:
    script_path = REPO_ROOT / "workflow-source" / "tests" / script_name
    subprocess.run([sys.executable, str(script_path)], cwd=REPO_ROOT, check=True)


def main() -> int:
    if HIDDEN_RUNTIME_DIR.exists():
        raise AssertionError(f"temporary runtime hide path already exists: {HIDDEN_RUNTIME_DIR}")

    runtime_was_present = RUNTIME_DIR.exists()
    if runtime_was_present:
        RUNTIME_DIR.rename(HIDDEN_RUNTIME_DIR)

    try:
        for script_name in (
            "check_docs.py",
            "check_bootstrap.py",
            "check_export_harness_package.py",
            "check_read_only_jsonrpc_fixtures.py",
            "check_read_only_harness_mcp_examples.py",
            "check_workflow_state_generator.py",
            "check_handoff_git_integration.py",
            "check_paths.py",
        ):
            run_check(script_name)
    finally:
        if HIDDEN_RUNTIME_DIR.exists():
            HIDDEN_RUNTIME_DIR.rename(RUNTIME_DIR)

    print("Source development checks passed without the applied ai-workflow runtime layer.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
