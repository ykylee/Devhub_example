#!/usr/bin/env python3
"""Verify the optional official MCP SDK candidate module stays aligned with runtime state."""

from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
SDK_CANDIDATE = SOURCE_ROOT / "workflow_kit" / "server" / "read_only_mcp_sdk.py"

if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.server.read_only_mcp_sdk import build_lowlevel_server, sdk_runtime_status
from workflow_kit.server.read_only_registry import READ_ONLY_SERVER_NAME, build_transport_tool_descriptors


def main() -> int:
    runtime = sdk_runtime_status()
    descriptors = build_transport_tool_descriptors()
    if runtime["status"] != "ok":
        raise AssertionError("Expected sdk runtime status to be ok.")
    if runtime["server_name"] != READ_ONLY_SERVER_NAME:
        raise AssertionError("Expected sdk runtime status to use read-only bundle server name.")
    if runtime["transport_ready"] is not False:
        raise AssertionError("Expected optional sdk candidate to remain transport_ready=false.")
    if runtime["sdk_candidate_phase"] != "official_sdk_optional_candidate":
        raise AssertionError("Expected official sdk optional candidate phase.")
    if runtime["descriptor_target"] != descriptors["descriptor_target"]:
        raise AssertionError("Expected sdk runtime status to reuse descriptor target.")
    if runtime["tool_count"] != descriptors["tool_count"]:
        raise AssertionError("Expected sdk runtime status to reuse descriptor tool count.")

    completed = subprocess.run(
        [sys.executable, str(SDK_CANDIDATE), "--print-sdk-runtime"],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
        check=True,
    )
    cli_runtime = json.loads(completed.stdout)
    if cli_runtime != runtime:
        raise AssertionError("Expected sdk runtime CLI output to match module helper output.")

    if runtime["sdk_available"]:
        server = build_lowlevel_server()
        if server is None:
            raise AssertionError("Expected low-level server to be buildable when the SDK is available.")
    else:
        run_completed = subprocess.run(
            [sys.executable, str(SDK_CANDIDATE), "--stdio-sdk"],
            cwd=REPO_ROOT,
            capture_output=True,
            text=True,
            check=False,
        )
        if run_completed.returncode == 0:
            raise AssertionError("Expected missing official sdk run to fail.")
        failure = json.loads(run_completed.stdout)
        if failure["error_code"] != "missing_official_mcp_sdk":
            raise AssertionError("Expected missing official sdk error code.")
        if failure["source_context"]["sdk_available"] is not False:
            raise AssertionError("Expected missing official sdk source_context to preserve sdk_available=false.")

    print("Read-only MCP SDK candidate check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
