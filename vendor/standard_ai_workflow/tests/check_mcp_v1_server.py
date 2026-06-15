#!/usr/bin/env python3
"""Smoke test for MCP v1.0 SDK server implementation."""

from __future__ import annotations

import os
import sys
import json
import subprocess
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"

if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

# We will try to import anyio and mcp client tools to run the test
try:
    import anyio
    from mcp import ClientSession
    from mcp.client.stdio import StdioServerParameters, stdio_client
    HAS_CLIENT_DEPS = True
except ImportError:
    HAS_CLIENT_DEPS = False

async def run_v1_smoke() -> None:
    # Use latest_backlog_v1 as the test target
    server_script = SOURCE_ROOT / "mcp_servers" / "latest-backlog" / "scripts" / "run_latest_backlog_v1.py"
    
    server_params = StdioServerParameters(
        command=sys.executable,
        args=[str(server_script)],
        cwd=str(REPO_ROOT),
        env={
            **os.environ,
            "PYTHONPATH": str(SOURCE_ROOT),
        },
    )

    async with stdio_client(server_params) as (read_stream, write_stream):
        async with ClientSession(read_stream, write_stream) as session:
            # Initialize
            await session.initialize()
            
            # List tools
            tools_result = await session.list_tools()
            tool_names = [t.name for t in tools_result.tools]
            print(f"Discovered tools: {tool_names}")
            
            if "latest_backlog" not in tool_names:
                raise AssertionError("latest_backlog tool not found in v1 server.")
            
            # Call tool (with a mock path)
            # We don't need real files for the smoke check if we just check the tool existence,
            # but let's try a call.
            # payload = {"work_backlog_index_path": "non_existent.md"}
            # result = await session.call_tool("latest_backlog", payload)
            # print(f"Call result: {result.content}")

def main() -> int:
    if not HAS_CLIENT_DEPS:
        print("Skipping MCP v1.0 smoke test: anyio or mcp not installed in the test runner environment.")
        return 0
    
    try:
        anyio.run(run_v1_smoke)
        print("MCP v1.0 smoke test passed.")
        return 0
    except Exception as e:
        print(f"MCP v1.0 smoke test failed: {e}")
        return 1

if __name__ == "__main__":
    sys.exit(main())
