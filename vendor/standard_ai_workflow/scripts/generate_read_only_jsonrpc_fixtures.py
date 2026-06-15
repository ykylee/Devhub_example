#!/usr/bin/env python3
"""Generate checked-in JSON-RPC draft bridge fixtures."""

from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.server.read_only_jsonrpc import JsonRpcSessionState, handle_jsonrpc_request, parse_request_json
from workflow_kit.server.read_only_registry import build_transport_tool_descriptors


def normalize_fixture_value(value: Any) -> Any:
    if isinstance(value, dict):
        return {key: normalize_fixture_value(item) for key, item in value.items()}
    if isinstance(value, list):
        return [normalize_fixture_value(item) for item in value]
    if isinstance(value, str):
        stripped = value.strip()
        if stripped.startswith("{") or stripped.startswith("["):
            try:
                parsed = json.loads(value)
            except json.JSONDecodeError:
                pass
            else:
                normalized = normalize_fixture_value(parsed)
                return json.dumps(normalized, ensure_ascii=False, sort_keys=True)
        try:
            path = Path(value)
        except OSError:
            return value
        if path.is_absolute():
            try:
                return path.relative_to(REPO_ROOT).as_posix()
            except ValueError:
                return value
    return value


def request_response_pair(name: str, request: dict[str, Any]) -> dict[str, Any]:
    response = handle_jsonrpc_request(request)
    if response is None:
        raise ValueError(f"Fixture request produced no response: {name}")
    return {
        "name": name,
        "request": normalize_fixture_value(request),
        "response": normalize_fixture_value(response),
    }


def raw_request_response_pair(name: str, raw_request: str) -> dict[str, Any]:
    request, response = parse_request_json(raw_request)
    if response is None:
        if request is None:
            raise ValueError(f"Raw fixture request produced neither request nor response: {name}")
        response = handle_jsonrpc_request(request)
    if response is None:
        raise ValueError(f"Raw fixture request produced no response: {name}")
    return {
        "name": name,
        "request": raw_request,
        "response": normalize_fixture_value(response),
    }


def request_no_response_pair(name: str, request: dict[str, Any]) -> dict[str, Any]:
    response = handle_jsonrpc_request(request)
    if response is not None:
        raise ValueError(f"Fixture request should not produce a response: {name}")
    return {
        "name": name,
        "request": normalize_fixture_value(request),
        "response": None,
    }


def session_request_response_pair(name: str, requests: list[dict[str, Any]]) -> dict[str, Any]:
    session_state = JsonRpcSessionState()
    responses = [handle_jsonrpc_request(request, session_state) for request in requests]
    return {
        "name": name,
        "requests": normalize_fixture_value(requests),
        "responses": normalize_fixture_value(responses),
    }


def build_jsonrpc_fixtures() -> dict[str, Any]:
    descriptors = build_transport_tool_descriptors()
    latest_backlog_payload = {
        "work_backlog_index_path": str((REPO_ROOT / "workflow-source" / "examples" / "acme_delivery_platform" / "work_backlog.md").resolve())
    }
    fixtures = [
        request_response_pair(
            "initialize",
            {"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}},
        ),
        request_response_pair(
            "initialize_with_supported_capabilities",
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
            },
        ),
        request_response_pair(
            "initialize_invalid_capabilities",
            {"jsonrpc": "2.0", "id": 11, "method": "initialize", "params": {"capabilities": []}},
        ),
        request_response_pair(
            "initialize_invalid_tools_list_changed",
            {
                "jsonrpc": "2.0",
                "id": 12,
                "method": "initialize",
                "params": {"capabilities": {"tools": {"listChanged": "yes"}}},
            },
        ),
        request_response_pair(
            "initialize_invalid_roots_capability",
            {
                "jsonrpc": "2.0",
                "id": 13,
                "method": "initialize",
                "params": {"capabilities": {"roots": []}},
            },
        ),
        request_no_response_pair(
            "notification_initialized_no_response",
            {"jsonrpc": "2.0", "method": "notifications/initialized", "params": {}},
        ),
        request_no_response_pair(
            "notification_unknown_no_response",
            {"jsonrpc": "2.0", "method": "notifications/progress", "params": {"step": "draft"}},
        ),
        request_response_pair(
            "notification_with_id_invalid_request",
            {"jsonrpc": "2.0", "id": 14, "method": "notifications/initialized", "params": {}},
        ),
        request_response_pair(
            "tools_list",
            {"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}},
        ),
        request_response_pair(
            "latest_backlog_call_success",
            {
                "jsonrpc": "2.0",
                "id": 3,
                "method": "tools/call",
                "params": {"name": "latest_backlog", "arguments": latest_backlog_payload},
            },
        ),
        request_response_pair(
            "check_doc_metadata_call_schema_error",
            {
                "jsonrpc": "2.0",
                "id": 4,
                "method": "tools/call",
                "params": {"name": "check_doc_metadata", "arguments": {}},
            },
        ),
        request_response_pair(
            "unknown_method",
            {"jsonrpc": "2.0", "id": 5, "method": "unknown/method", "params": {}},
        ),
        request_response_pair(
            "invalid_boolean_id",
            {"jsonrpc": "2.0", "id": True, "method": "tools/list", "params": {}},
        ),
        raw_request_response_pair("malformed_json_parse_error", "{not-json"),
        raw_request_response_pair("non_object_invalid_request", "[]"),
    ]
    session_fixtures = [
        session_request_response_pair(
            "stdio_session_requires_initialize",
            [
                {"jsonrpc": "2.0", "id": 21, "method": "tools/list", "params": {}},
                {"jsonrpc": "2.0", "id": 22, "method": "initialize", "params": {}},
                {"jsonrpc": "2.0", "method": "notifications/initialized", "params": {}},
                {"jsonrpc": "2.0", "id": 23, "method": "tools/list", "params": {}},
            ],
        ),
        session_request_response_pair(
            "stdio_session_rejects_second_initialize",
            [
                {"jsonrpc": "2.0", "id": 24, "method": "initialize", "params": {}},
                {"jsonrpc": "2.0", "id": 25, "method": "initialize", "params": {}},
            ],
        ),
    ]
    return {
        "status": "ok",
        "fixture_phase": "jsonrpc_draft_fixture",
        "transport_ready": False,
        "descriptor_target": descriptors["descriptor_target"],
        "tool_version": descriptors["tool_version"],
        "fixture_count": len(fixtures),
        "session_fixture_count": len(session_fixtures),
        "fixtures": fixtures,
        "session_fixtures": session_fixtures,
    }


def main() -> int:
    print(json.dumps(build_jsonrpc_fixtures(), ensure_ascii=False, indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
