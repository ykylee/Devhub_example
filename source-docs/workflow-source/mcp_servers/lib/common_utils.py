"""Common utilities for MCP server scripts."""

import sys
import json
import argparse
from pathlib import Path
from typing import Any, Callable, Optional, Union

def inject_workflow_source():
    """Inject workflow_kit root into sys.path."""
    source_root = Path(__file__).resolve().parent
    while not (source_root / "workflow_kit").is_dir() and source_root != source_root.parent:
        source_root = source_root.parent
    if str(source_root) not in sys.path:
        sys.path.insert(0, str(source_root))

def mcp_main(
    description: str,
    arg_builder: Callable[[argparse.ArgumentParser], None],
    payload_func: Callable[..., dict[str, Any]],
    output_handler: Optional[Callable[[dict[str, Any]], None]] = None
):
    """Generic main function for MCP scripts."""
    parser = argparse.ArgumentParser(description=description)
    arg_builder(parser)
    # Add global --json flag for all MCP runners
    parser.add_argument("--json", action="store_true", help="Output raw JSON instead of formatted text.")
    args = parser.parse_args()

    # We expect tool_version to be injected by the caller or retrieved here
    from workflow_kit import __version__ as TOOL_VERSION

    # Extract --json before passing args to payload_func
    arg_vars = vars(args)
    use_json = arg_vars.pop("json", False)

    result = payload_func(**arg_vars, tool_version=TOOL_VERSION)

    if use_json or not output_handler:
        print(json.dumps(result, ensure_ascii=False, indent=2))
        if result.get("status") != "ok":
            sys.exit(1)
    else:
        output_handler(result)

def standard_output_handler(result: dict[str, Any], content_key: str):
    """Standard handler that prints content or JSON error."""
    if result.get("status") == "ok":
        content = result.get(content_key)
        if isinstance(content, list):
            print("\n\n".join(content))
        else:
            print(content)
    else:
        print(json.dumps(result, ensure_ascii=False, indent=2))
        sys.exit(1)
