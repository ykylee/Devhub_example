#!/usr/bin/env python3
"""Verify generated JSON Schema drafts stay aligned with runtime contracts."""

from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
SCRIPT_PATH = SOURCE_ROOT / "scripts" / "generate_output_json_schema.py"
SCHEMA_PATH = SOURCE_ROOT / "schemas" / "generated_output_schemas.json"

if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.common.output_contracts import output_json_schema_bundle


def main() -> int:
    checked_in = json.loads(SCHEMA_PATH.read_text(encoding="utf-8"))
    runtime = output_json_schema_bundle()
    if checked_in != runtime:
        raise AssertionError("Checked-in generated_output_schemas.json is out of date with runtime contracts.")

    completed = subprocess.run(
        [sys.executable, str(SCRIPT_PATH)],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
        check=True,
    )
    script_output = json.loads(completed.stdout)
    if script_output != runtime:
        raise AssertionError("generate_output_json_schema.py output does not match runtime contracts.")

    if "validation_plan" not in checked_in.get("outputs", {}):
        raise AssertionError("Expected validation_plan family in generated JSON Schema bundle outputs.")
    if "validation_plan" not in checked_in.get("errors", {}):
        raise AssertionError("Expected validation_plan family in generated JSON Schema bundle errors.")

    print("Output JSON Schema generation check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
