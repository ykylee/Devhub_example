#!/usr/bin/env python3
"""Prototype runner for suggest_impacted_docs MCP."""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[4]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit import __version__ as TOOL_VERSION
from workflow_kit.common.read_only_bundle import suggest_impacted_docs_payload


def main() -> int:
    parser = argparse.ArgumentParser(description="Run suggest_impacted_docs MCP prototype.")
    parser.add_argument("--changed-file", action="append", dest="changed_files", default=[])
    parser.add_argument("--session-handoff-path")
    parser.add_argument("--latest-backlog-path")
    parser.add_argument("--work-backlog-index-path")
    args = parser.parse_args()

    if not args.changed_files:
        raise SystemExit("at least one --changed-file is required")

    print(
        json.dumps(
            suggest_impacted_docs_payload(
                changed_files=args.changed_files,
                session_handoff_path=args.session_handoff_path,
                latest_backlog_path=args.latest_backlog_path,
                work_backlog_index_path=args.work_backlog_index_path,
                tool_version=TOOL_VERSION,
            ),
            ensure_ascii=False,
            indent=2,
        )
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
