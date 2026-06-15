#!/usr/bin/env python3
"""Smoke test the draft JSON-RPC bridge for the read-only bundle."""

from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
BRIDGE = SOURCE_ROOT / "workflow_kit" / "server" / "read_only_jsonrpc.py"


def run_request(request: dict[str, object]) -> dict[str, object]:
    completed = subprocess.run(
        [sys.executable, str(BRIDGE), "--request-json", json.dumps(request, ensure_ascii=False)],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
        check=True,
    )
    if not completed.stdout.strip():
        raise AssertionError("Expected JSON-RPC response on stdout.")
    return json.loads(completed.stdout)


def run_raw_request(raw_request: str, *, expect_success: bool = True) -> tuple[int, dict[str, object]]:
    completed = subprocess.run(
        [sys.executable, str(BRIDGE), "--request-json", raw_request],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if expect_success and completed.returncode != 0:
        raise AssertionError(f"Expected raw request to succeed: {completed.stderr}\n{completed.stdout}")
    if not completed.stdout.strip():
        raise AssertionError("Expected JSON-RPC response on stdout.")
    return completed.returncode, json.loads(completed.stdout)


def run_stdio_lines(lines: list[str]) -> list[dict[str, object]]:
    completed = subprocess.run(
        [sys.executable, str(BRIDGE), "--stdio-lines"],
        input="\n".join(lines) + "\n",
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
        check=True,
    )
    return [json.loads(line) for line in completed.stdout.splitlines() if line.strip()]


def main() -> int:
    initialize = run_request({"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}})
    if initialize["result"]["serverInfo"]["name"] != "workflow_read_only_bundle":
        raise AssertionError("Expected read-only bundle serverInfo name.")
    if initialize["result"]["_meta"]["transport_ready"] is not False:
        raise AssertionError("Expected draft bridge to remain transport_ready=false.")
    if "protocolVersion" not in initialize["result"]:
        raise AssertionError(
            "OpenCode / 2025-03-26 client 가 initialize response 에서 "
            "protocolVersion 을 필수로 기대. 부재 시 '4 of 5 requests failed: "
            "unexpected error' 발생 (MCP 2025-03-26 spec 준수)."
        )

    initialize_rich = run_request(
        {
            "jsonrpc": "2.0",
            "id": 10,
            "method": "initialize",
            "params": {
                "protocolVersion": "2025-03-26",
                "clientInfo": {"name": "fixture-client", "version": "0.1.0"},
                "capabilities": {
                    "tools": {"listChanged": True},
                    "roots": {"listChanged": False},
                    "sampling": {},
                    "elicitation": {},
                    "experimental": {},
                },
            },
        }
    )
    if initialize_rich["result"]["serverInfo"]["name"] != "workflow_read_only_bundle":
        raise AssertionError("Expected valid richer initialize request to succeed.")

    initialize_invalid = run_request(
        {"jsonrpc": "2.0", "id": 11, "method": "initialize", "params": {"capabilities": []}}
    )
    if initialize_invalid["error"]["code"] != -32602:
        raise AssertionError("Expected invalid initialize capabilities to return Invalid params.")
    initialize_invalid_tools = run_request(
        {
            "jsonrpc": "2.0",
            "id": 12,
            "method": "initialize",
            "params": {"capabilities": {"tools": {"listChanged": "yes"}}},
        }
    )
    if initialize_invalid_tools["error"]["code"] != -32602:
        raise AssertionError("Expected invalid tools.listChanged to return Invalid params.")
    initialize_invalid_roots = run_request(
        {"jsonrpc": "2.0", "id": 13, "method": "initialize", "params": {"capabilities": {"roots": []}}}
    )
    if initialize_invalid_roots["error"]["code"] != -32602:
        raise AssertionError("Expected invalid roots capability to return Invalid params.")
    notification_with_id = run_request(
        {"jsonrpc": "2.0", "id": 14, "method": "notifications/initialized", "params": {}}
    )
    if notification_with_id["error"]["code"] != -32600:
        raise AssertionError("Expected notification with id to return Invalid Request.")

    tools_list = run_request({"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}})
    tools = tools_list["result"]["tools"]
    if len(tools) < 5:
        raise AssertionError("Expected at least five read-only tools.")
    latest_backlog = next((tool for tool in tools if tool["name"] == "latest_backlog"), None)
    if latest_backlog is None:
        raise AssertionError("Expected latest_backlog tool descriptor.")
    if latest_backlog["annotations"]["readOnlyHint"] is not True:
        raise AssertionError("Expected readOnlyHint annotation.")

    call_payload = {
        "work_backlog_index_path": str(SOURCE_ROOT / "examples" / "acme_delivery_platform" / "work_backlog.md")
    }
    tool_call = run_request(
        {
            "jsonrpc": "2.0",
            "id": 3,
            "method": "tools/call",
            "params": {"name": "latest_backlog", "arguments": call_payload},
        }
    )
    result = tool_call["result"]
    if result["structuredContent"]["status"] != "ok":
        raise AssertionError("Expected latest_backlog JSON-RPC tool call to succeed.")
    if result["content"][0]["type"] != "text":
        raise AssertionError("Expected text content wrapper.")
    if result["_meta"]["bridge_phase"] != "jsonrpc_draft_fixture":
        raise AssertionError("Expected JSON-RPC draft bridge phase metadata.")

    error_call = run_request(
        {
            "jsonrpc": "2.0",
            "id": 4,
            "method": "tools/call",
            "params": {"name": "check_doc_metadata", "arguments": {}},
        }
    )
    if error_call["error"]["code"] != -32000:
        raise AssertionError("Expected tool-call failure to map to JSON-RPC server error.")
    if error_call["error"]["data"]["error_code"] != "invalid_tool_payload_schema":
        raise AssertionError("Expected entrypoint error payload in JSON-RPC error data.")

    missing_method = run_request({"jsonrpc": "2.0", "id": 5, "method": "unknown/method", "params": {}})
    if missing_method["error"]["code"] != -32601:
        raise AssertionError("Expected unknown methods to return Method not found.")

    invalid_boolean_id = run_request({"jsonrpc": "2.0", "id": True, "method": "tools/list", "params": {}})
    if invalid_boolean_id["error"]["code"] != -32600:
        raise AssertionError("Expected boolean id to return Invalid Request.")

    parse_error_code, parse_error = run_raw_request("{not-json", expect_success=False)
    if parse_error_code == 0:
        raise AssertionError("Expected malformed --request-json to return non-zero.")
    if parse_error["error"]["code"] != -32700:
        raise AssertionError("Expected malformed JSON to return JSON-RPC Parse error.")

    invalid_request_code, invalid_request = run_raw_request("[]", expect_success=False)
    if invalid_request_code == 0:
        raise AssertionError("Expected non-object --request-json to return non-zero.")
    if invalid_request["error"]["code"] != -32600:
        raise AssertionError("Expected non-object JSON to return Invalid Request.")

    stdio_responses = run_stdio_lines(
        [
            "{not-json",
            json.dumps({"jsonrpc": "2.0", "method": "notifications/initialized", "params": {}}, ensure_ascii=False),
            json.dumps({"jsonrpc": "2.0", "method": "notifications/progress", "params": {"step": "draft"}}, ensure_ascii=False),
            json.dumps({"jsonrpc": "2.0", "id": 6, "method": "tools/list", "params": {}}, ensure_ascii=False),
            json.dumps({"jsonrpc": "2.0", "id": 7, "method": "initialize", "params": {}}, ensure_ascii=False),
            json.dumps({"jsonrpc": "2.0", "method": "notifications/initialized", "params": {}}, ensure_ascii=False),
            json.dumps({"jsonrpc": "2.0", "id": 8, "method": "tools/list", "params": {}}, ensure_ascii=False),
            json.dumps({"jsonrpc": "2.0", "id": 9, "method": "initialize", "params": {}}, ensure_ascii=False),
        ]
    )
    if [response.get("error", {}).get("code") for response in stdio_responses[:1]] != [-32700]:
        raise AssertionError("Expected stdio malformed line to return Parse error and continue.")
    if len(stdio_responses) != 5:
        raise AssertionError("Expected notifications to be ignored while stdio requests still emit responses.")
    if stdio_responses[1]["error"]["code"] != -32002:
        raise AssertionError("Expected pre-initialize stdio tools/list to return Server not initialized.")
    if stdio_responses[2]["result"]["serverInfo"]["name"] != "workflow_read_only_bundle":
        raise AssertionError("Expected stdio initialize to succeed once session state is created.")
    if stdio_responses[3]["result"]["_meta"]["tool_count"] < 5:
        raise AssertionError("Expected stdio bridge to allow tools/list after initialize.")
    if stdio_responses[4]["error"]["data"]["reason"] != "initialize may only be called once per stdio session":
        raise AssertionError("Expected stdio bridge to reject a second initialize call.")

    print("Read-only JSON-RPC bridge smoke check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
