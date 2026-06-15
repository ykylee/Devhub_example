#!/usr/bin/env python3
"""Generate draft harness MCP config examples for the read-only descriptor bundle."""

from __future__ import annotations

import json
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.server.read_only_registry import build_transport_tool_descriptors


SOURCE_DESCRIPTOR_PATH = "workflow-source/schemas/read_only_transport_descriptors.json"
SERVER_ALIAS = "standardAiWorkflowReadOnly"


def tool_names_from_descriptor(descriptor_bundle: dict[str, object]) -> list[str]:
    tools = descriptor_bundle["tools"]
    if not isinstance(tools, list):
        raise TypeError("Expected descriptor tools to be a list.")
    return [str(tool["name"]) for tool in tools if isinstance(tool, dict)]


def codex_toml_example(tool_names: list[str]) -> str:
    tools = ", ".join(tool_names)
    return "\n".join(
        [
            "# Draft only: generated from schemas/read_only_transport_descriptors.json.",
            "# transport_ready=false; do not paste this as an active server until an MCP SDK server loop exists.",
            f"# Tools described: {tools}",
            f"# [mcp_servers.{SERVER_ALIAS}]",
            '# command = "python3"',
            '# args = ["-m", "workflow_kit.server.read_only_jsonrpc", "--stdio-lines"]',
            "# NOTE: current bridge is a JSON-RPC draft fixture, not a full MCP SDK server.",
        ]
    )


def opencode_jsonc_example(tool_names: list[str]) -> str:
    tools = ", ".join(tool_names)
    return "\n".join(
        [
            "{",
            "  // Draft only: generated from schemas/read_only_transport_descriptors.json.",
            "  // transport_ready=false; do not enable until an MCP SDK server loop exists.",
            f"  // Tools described: {tools}",
            '  "mcp_servers": {',
            f'    // "{SERVER_ALIAS}": {{',
            '    //   "type": "local",',
            '    //   "command": "python3",',
            '    //   "args": ["-m", "workflow_kit.server.read_only_jsonrpc", "--stdio-lines"]',
            "    // }",
            "  }",
            "}",
        ]
    )


def build_harness_mcp_examples() -> dict[str, object]:
    descriptor_bundle = build_transport_tool_descriptors()
    tool_names = tool_names_from_descriptor(descriptor_bundle)
    return {
        "status": "ok",
        "tool_version": descriptor_bundle["tool_version"],
        "source_descriptor_path": SOURCE_DESCRIPTOR_PATH,
        "descriptor_target": descriptor_bundle["descriptor_target"],
        "transport_ready": descriptor_bundle["transport_ready"],
        "tool_count": descriptor_bundle["tool_count"],
        "tool_names": tool_names,
        "harness_examples": {
            "codex": {
                "format": "toml_snippet_draft",
                "target_path": "~/.codex/config.toml",
                "apply_mode": "manual_review_only",
                "bridge_entrypoint": "workflow_kit.server.read_only_jsonrpc",
                "server_alias": SERVER_ALIAS,
                "content": codex_toml_example(tool_names),
            },
            "opencode": {
                "format": "jsonc_snippet_draft",
                "target_path": "opencode.json",
                "apply_mode": "manual_review_only",
                "bridge_entrypoint": "workflow_kit.server.read_only_jsonrpc",
                "server_alias": SERVER_ALIAS,
                "content": opencode_jsonc_example(tool_names),
            },
        },
    }


def main() -> int:
    print(json.dumps(build_harness_mcp_examples(), ensure_ascii=False, indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
