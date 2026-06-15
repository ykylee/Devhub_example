#!/usr/bin/env python3
"""Verify checked-in read-only harness MCP examples stay aligned with descriptors."""

from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
SCRIPT_PATH = SOURCE_ROOT / "scripts" / "generate_read_only_harness_mcp_examples.py"
EXAMPLES_PATH = SOURCE_ROOT / "schemas" / "read_only_harness_mcp_examples.json"

if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from scripts.generate_read_only_harness_mcp_examples import build_harness_mcp_examples
from workflow_kit.server.read_only_registry import build_transport_tool_descriptors


def main() -> int:
    checked_in = json.loads(EXAMPLES_PATH.read_text(encoding="utf-8"))
    runtime = build_harness_mcp_examples()
    if checked_in != runtime:
        raise AssertionError("Checked-in read_only_harness_mcp_examples.json is out of date.")

    completed = subprocess.run(
        [sys.executable, str(SCRIPT_PATH)],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
        check=True,
    )
    script_output = json.loads(completed.stdout)
    if script_output != runtime:
        raise AssertionError("generate_read_only_harness_mcp_examples.py output does not match runtime output.")

    descriptors = build_transport_tool_descriptors()
    if checked_in["descriptor_target"] != descriptors["descriptor_target"]:
        raise AssertionError("Harness examples should preserve descriptor target.")
    if checked_in["transport_ready"] is not False:
        raise AssertionError("Harness examples should remain transport_ready=false.")
    if checked_in["tool_names"] != [tool["name"] for tool in descriptors["tools"]]:
        raise AssertionError("Harness examples should list descriptor tool names in registry order.")

    examples = checked_in["harness_examples"]
    for harness in ("codex", "opencode"):
        example = examples[harness]
        if example["apply_mode"] != "manual_review_only":
            raise AssertionError(f"{harness} example should be manual-review-only.")
        content = example["content"]
        if example["bridge_entrypoint"] != "workflow_kit.server.read_only_jsonrpc":
            raise AssertionError(f"{harness} example should name the JSON-RPC draft bridge.")
        if "transport_ready=false" not in content:
            raise AssertionError(f"{harness} example should expose transport_ready=false.")
        if "read_only_transport_descriptors.json" not in content:
            raise AssertionError(f"{harness} example should point back to the descriptor file.")
        if "workflow_kit.server.read_only_jsonrpc" not in content:
            raise AssertionError(f"{harness} example should name the current JSON-RPC draft bridge.")

    print("Read-only harness MCP example generation check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
