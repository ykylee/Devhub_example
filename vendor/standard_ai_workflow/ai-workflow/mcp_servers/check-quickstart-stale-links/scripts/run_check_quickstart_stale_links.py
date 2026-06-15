#!/usr/bin/env python3
"""Prototype runner for check_quickstart_stale_links MCP."""

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
from workflow_kit.common.read_only_bundle import check_quickstart_stale_links_payload


def main() -> int:
    parser = argparse.ArgumentParser(description="Run check_quickstart_stale_links MCP prototype.")
    parser.add_argument("--quickstart-path", action="append", dest="quickstart_paths", required=True)
    parser.add_argument("--project-profile-path")
    parser.add_argument("--session-handoff-path")
    parser.add_argument("--work-backlog-index-path")
    parser.add_argument("--agents-path")
    args = parser.parse_args()

    print(
        json.dumps(
            check_quickstart_stale_links_payload(
                quickstart_paths=args.quickstart_paths,
                project_profile_path=args.project_profile_path,
                session_handoff_path=args.session_handoff_path,
                work_backlog_index_path=args.work_backlog_index_path,
                agents_path=args.agents_path,
                tool_version=TOOL_VERSION,
            ),
            ensure_ascii=False,
            indent=2,
        )
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
