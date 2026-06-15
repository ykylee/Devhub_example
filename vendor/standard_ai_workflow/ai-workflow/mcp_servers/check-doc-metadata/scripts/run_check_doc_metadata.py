#!/usr/bin/env python3
"""Prototype runner for check_doc_metadata MCP."""

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
from workflow_kit.common.read_only_bundle import check_doc_metadata_payload


def main() -> int:
    parser = argparse.ArgumentParser(description="Run check_doc_metadata MCP prototype.")
    parser.add_argument("--doc-dir-path", required=True)
    args = parser.parse_args()

    print(
        json.dumps(
            check_doc_metadata_payload(doc_dir_path=args.doc_dir_path, tool_version=TOOL_VERSION),
            ensure_ascii=False,
            indent=2,
        )
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
