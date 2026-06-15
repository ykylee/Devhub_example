#!/usr/bin/env python3
"""Verify checked-in JSON-RPC draft bridge fixtures stay aligned with runtime output."""

from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
SCRIPT_PATH = SOURCE_ROOT / "scripts" / "generate_read_only_jsonrpc_fixtures.py"
FIXTURE_PATH = SOURCE_ROOT / "schemas" / "read_only_jsonrpc_fixtures.json"

if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from scripts.generate_read_only_jsonrpc_fixtures import build_jsonrpc_fixtures


def main() -> int:
    checked_in = json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))
    runtime = build_jsonrpc_fixtures()
    if checked_in != runtime:
        raise AssertionError("Checked-in read_only_jsonrpc_fixtures.json is out of date.")

    completed = subprocess.run(
        [sys.executable, str(SCRIPT_PATH)],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
        check=True,
    )
    script_output = json.loads(completed.stdout)
    if script_output != runtime:
        raise AssertionError("generate_read_only_jsonrpc_fixtures.py output does not match runtime output.")

    if checked_in["fixture_phase"] != "jsonrpc_draft_fixture":
        raise AssertionError("Expected JSON-RPC draft fixture phase.")
    if checked_in["transport_ready"] is not False:
        raise AssertionError("Expected fixtures to remain transport_ready=false.")
    if checked_in["session_fixture_count"] != len(checked_in["session_fixtures"]):
        raise AssertionError("Expected session fixture count to match the checked-in session fixture list.")
    names = [fixture["name"] for fixture in checked_in["fixtures"]]
    expected_names = [
        "initialize",
        "initialize_with_supported_capabilities",
        "initialize_invalid_capabilities",
        "initialize_invalid_tools_list_changed",
        "initialize_invalid_roots_capability",
        "notification_initialized_no_response",
        "notification_unknown_no_response",
        "notification_with_id_invalid_request",
        "tools_list",
        "latest_backlog_call_success",
        "check_doc_metadata_call_schema_error",
        "unknown_method",
        "invalid_boolean_id",
        "malformed_json_parse_error",
        "non_object_invalid_request",
    ]
    if names != expected_names:
        raise AssertionError("Unexpected JSON-RPC fixture ordering.")

    initialize_rich = checked_in["fixtures"][1]["response"]["result"]
    if initialize_rich["serverInfo"]["name"] != "workflow_read_only_bundle":
        raise AssertionError("Expected richer initialize fixture to succeed.")
    initialize_invalid = checked_in["fixtures"][2]["response"]["error"]
    if initialize_invalid["code"] != -32602:
        raise AssertionError("Expected invalid initialize fixture to use Invalid params.")
    initialize_invalid_tools = checked_in["fixtures"][3]["response"]["error"]
    if initialize_invalid_tools["data"]["reason"] != "initialize params.capabilities.tools.listChanged must be a boolean":
        raise AssertionError("Expected tools.listChanged fixture to preserve boolean validation reason.")
    initialize_invalid_roots = checked_in["fixtures"][4]["response"]["error"]
    if initialize_invalid_roots["data"]["reason"] != "initialize params.capabilities.roots must be an object":
        raise AssertionError("Expected roots capability fixture to preserve object validation reason.")
    if checked_in["fixtures"][5]["response"] is not None:
        raise AssertionError("Expected initialized notification fixture to have no response.")
    if checked_in["fixtures"][6]["response"] is not None:
        raise AssertionError("Expected unknown notification fixture to have no response.")
    notification_with_id = checked_in["fixtures"][7]["response"]["error"]
    if notification_with_id["data"]["reason"] != "notifications must not include id":
        raise AssertionError("Expected notification-with-id fixture to preserve lifecycle validation reason.")
    success = checked_in["fixtures"][9]["response"]["result"]
    if success["structuredContent"]["status"] != "ok":
        raise AssertionError("Expected successful call fixture to contain structuredContent.")
    if checked_in["fixtures"][9]["request"]["params"]["arguments"]["work_backlog_index_path"] != "workflow-source/examples/acme_delivery_platform/work_backlog.md":
        raise AssertionError("Expected fixture request paths to remain repo-relative for portable comparisons.")
    latest_backlog_path = success["structuredContent"]["latest_backlog_path"]
    if latest_backlog_path != "workflow-source/examples/acme_delivery_platform/backlog/2026-04-18.md":
        raise AssertionError("Expected fixture response paths to remain repo-relative for portable comparisons.")
    schema_error = checked_in["fixtures"][10]["response"]["error"]
    if schema_error["code"] != -32000:
        raise AssertionError("Expected schema failure fixture to use JSON-RPC server error code.")
    if schema_error["data"]["error_code"] != "invalid_tool_payload_schema":
        raise AssertionError("Expected schema failure fixture to preserve entrypoint error code.")
    invalid_boolean_id = checked_in["fixtures"][12]["response"]["error"]
    if invalid_boolean_id["data"]["reason"] != "id must be a string, number, or null":
        raise AssertionError("Expected invalid id fixture to preserve id validation reason.")
    malformed = checked_in["fixtures"][13]["response"]["error"]
    if malformed["code"] != -32700:
        raise AssertionError("Expected malformed JSON fixture to preserve Parse error.")
    non_object = checked_in["fixtures"][14]["response"]["error"]
    if non_object["code"] != -32600:
        raise AssertionError("Expected non-object request fixture to preserve Invalid Request.")

    session_names = [fixture["name"] for fixture in checked_in["session_fixtures"]]
    expected_session_names = [
        "stdio_session_requires_initialize",
        "stdio_session_rejects_second_initialize",
    ]
    if session_names != expected_session_names:
        raise AssertionError("Unexpected JSON-RPC session fixture ordering.")

    requires_initialize = checked_in["session_fixtures"][0]["responses"]
    if requires_initialize[0]["error"]["code"] != -32002:
        raise AssertionError("Expected stdio session fixture to reject tools/list before initialize.")
    if requires_initialize[1]["result"]["serverInfo"]["name"] != "workflow_read_only_bundle":
        raise AssertionError("Expected stdio session fixture initialize to succeed.")
    if requires_initialize[2] is not None:
        raise AssertionError("Expected stdio notification fixture entry to produce no response.")
    if requires_initialize[3]["result"]["_meta"]["tool_count"] < 5:
        raise AssertionError("Expected stdio session fixture to allow tools/list after initialize.")

    duplicate_initialize = checked_in["session_fixtures"][1]["responses"][1]["error"]
    if duplicate_initialize["data"]["reason"] != "initialize may only be called once per stdio session":
        raise AssertionError("Expected duplicate initialize session fixture to preserve session validation reason.")

    print("Read-only JSON-RPC fixture generation check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
