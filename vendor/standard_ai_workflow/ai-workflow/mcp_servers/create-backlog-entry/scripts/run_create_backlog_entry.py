#!/usr/bin/env python3
"""Prototype runner for create_backlog_entry MCP."""

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


def main() -> int:
    parser = argparse.ArgumentParser(description="Run create_backlog_entry MCP prototype.")
    parser.add_argument("--task-id", required=True)
    parser.add_argument("--task-name", required=True)
    parser.add_argument("--request-date", required=True)
    parser.add_argument("--status", default="planned")
    parser.add_argument("--priority", default="high")
    args = parser.parse_args()

    draft_entry = [
        f"## {args.task_id} {args.task_name}",
        "",
        f"- 상태: {args.status}",
        f"- 우선순위: {args.priority}",
        f"- 요청일: {args.request_date}",
        "- 완료일:",
        "- 담당:",
        "- 호스트명:",
        "- 호스트 IP:",
        "- 영향 문서:",
        "- 작업 내용:",
        "- 진행 현황:",
        "- 완료 기준:",
        "- 작업 결과:",
        "- 다음 세션 시작 포인트:",
        "- 남은 리스크:",
        "- 후속 작업:",
    ]
    print(
        json.dumps(
            {
                "status": "ok",
                "tool_version": TOOL_VERSION,
                "draft_entry": draft_entry,
                "warnings": [],
            },
            ensure_ascii=False,
            indent=2,
        )
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
