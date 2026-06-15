#!/usr/bin/env python3
"""Official MCP v1.0 runner for check_doc_metadata."""

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
from workflow_kit.common.read_only_bundle import check_doc_metadata_payload

server = create_v1_server("check-doc-metadata-v1", version=TOOL_VERSION)

@server.tool()
def check_doc_metadata(doc_dir_path: str) -> dict:
    """Inspect markdown files and report missing required metadata fields."""
    return check_doc_metadata_payload(
        doc_dir_path=doc_dir_path,
        tool_version=TOOL_VERSION
    )

if __name__ == "__main__":
    server.run()
