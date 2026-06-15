#!/usr/bin/env python3
"""Smoke test the official MCP SDK stdio candidate through a real MCP client session."""

from __future__ import annotations

import os
import sys
from pathlib import Path

try:
    import anyio
    from mcp import ClientSession
    from mcp.client.stdio import StdioServerParameters, stdio_client
    HAS_MCP_SDK = True
except ImportError:
    HAS_MCP_SDK = False


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"


async def run_stdio_smoke() -> None:
    server = StdioServerParameters(
        command=sys.executable,
        args=["-m", "workflow_kit.server.read_only_mcp_sdk", "--stdio-sdk"],
        cwd=str(REPO_ROOT),
        env={
            **os.environ,
            "PYTHONPATH": str(SOURCE_ROOT),
        },
    )

    async with stdio_client(server) as (read_stream, write_stream):
        async with ClientSession(read_stream, write_stream) as session:
            initialize = await session.initialize()
            if initialize.serverInfo.name != "workflow_read_only_bundle":
                raise AssertionError("Expected stdio SDK server to expose read-only bundle server name.")

            tools_result = await session.list_tools()
            if len(tools_result.tools) < 6:
                raise AssertionError(f"Expected at least six read-only tools over stdio SDK session, got {len(tools_result.tools)}.")
            latest_backlog = next((tool for tool in tools_result.tools if tool.name == "latest_backlog"), None)
            if latest_backlog is None:
                raise AssertionError("Expected latest_backlog tool over stdio SDK session.")
            if latest_backlog.annotations is None or latest_backlog.annotations.readOnlyHint is not True:
                raise AssertionError("Expected latest_backlog tool to remain read-only annotated.")

            create_backlog_entry = next((tool for tool in tools_result.tools if tool.name == "create_backlog_entry"), None)
            if create_backlog_entry is None:
                raise AssertionError("Expected create_backlog_entry tool over stdio SDK session.")

            latest_backlog_payload = {
                "work_backlog_index_path": str(SOURCE_ROOT / "examples" / "acme_delivery_platform" / "work_backlog.md")
            }
            call_result = await session.call_tool("latest_backlog", latest_backlog_payload)
            if call_result.isError:
                raise AssertionError("Expected latest_backlog stdio SDK call to succeed.")
            if call_result.structuredContent is None or call_result.structuredContent["status"] != "ok":
                raise AssertionError("Expected stdio SDK call to preserve structuredContent.")

            schema_error = await session.call_tool("check_doc_metadata", {})
            if not schema_error.isError:
                raise AssertionError("Expected invalid check_doc_metadata payload to surface as tool error.")
            if schema_error.structuredContent is None:
                raise AssertionError("Expected failing stdio SDK call to preserve structuredContent.")
            if schema_error.structuredContent["error_code"] != "invalid_tool_payload_schema":
                raise AssertionError("Expected schema-invalid stdio SDK call to preserve entrypoint error code.")


def main() -> int:
    if not HAS_MCP_SDK:
        print("Skipping Read-only MCP SDK stdio smoke check: mcp or anyio not installed.")
        return 0
    anyio.run(run_stdio_smoke)
    print("Read-only MCP SDK stdio smoke check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
