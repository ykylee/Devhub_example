#!/usr/bin/env python3
"""Official MCP v1.0 runner for latest_backlog."""

import sys
from pathlib import Path

# Add lib to path for common_utils
LIB_PATH = Path(__file__).resolve().parents[2] / "lib"
if str(LIB_PATH) not in sys.path:
    sys.path.insert(0, str(LIB_PATH))

from common_utils import inject_workflow_source
inject_workflow_source()

from workflow_kit import __version__ as TOOL_VERSION
from workflow_kit.server.mcp_v1_server import create_v1_server
from workflow_kit.common.read_only_bundle import latest_backlog_payload

server = create_v1_server("latest-backlog-v1", version=TOOL_VERSION)

@server.tool()
def latest_backlog(
    backlog_dir_path: str | None = None,
    work_backlog_index_path: str | None = None
) -> dict:
    """Locate the latest dated backlog document from an index or backlog directory."""
    return latest_backlog_payload(
        backlog_dir_path=backlog_dir_path,
        work_backlog_index_path=work_backlog_index_path,
        tool_version=TOOL_VERSION
    )

if __name__ == "__main__":
    server.run()
