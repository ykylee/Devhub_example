#!/usr/bin/env python3
"""Verify the read-only MCP transport promotion spec tracks runtime fixture names."""

from __future__ import annotations

import json
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
SPEC_PATH = SOURCE_ROOT / "core" / "read_only_mcp_transport_promotion.md"
FIXTURE_PATH = SOURCE_ROOT / "schemas" / "read_only_jsonrpc_fixtures.json"


def main() -> int:
    spec_text = SPEC_PATH.read_text(encoding="utf-8")
    fixtures = json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))

    for key_phrase in (
        "transport_ready=false",
        "manual_review_only",
        "annotations.readOnlyHint: true",
        "inputSchema",
        "outputSchema",
    ):
        if key_phrase not in spec_text:
            raise AssertionError(f"Promotion spec should mention `{key_phrase}`.")

    for fixture in fixtures["fixtures"]:
        fixture_name = fixture["name"]
        if f"`{fixture_name}`" not in spec_text:
            raise AssertionError(f"Promotion spec should mention fixture `{fixture_name}`.")
    for fixture in fixtures.get("session_fixtures", []):
        fixture_name = fixture["name"]
        if f"`{fixture_name}`" not in spec_text:
            raise AssertionError(f"Promotion spec should mention session fixture `{fixture_name}`.")

    allowed_envelope_terms = (
        "JSON-RPC top-level `id`",
        "tool call result 의 `content` wrapper",
        "tool call error 의 JSON-RPC error code",
        "stdio framing",
    )
    for term in allowed_envelope_terms:
        if term not in spec_text:
            raise AssertionError(f"Promotion spec should document envelope term: {term}")

    print("Read-only MCP transport promotion spec check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
