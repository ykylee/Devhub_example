#!/usr/bin/env python3
import sys
import tempfile
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.common.project_docs import parse_backlog, parse_handoff


def check_branch_memory_format() -> None:
    with tempfile.TemporaryDirectory() as tmpdir:
        root = Path(tmpdir)
        handoff = root / "session_handoff.md"
        backlog = root / "backlog" / "2026-05-01.md"
        task = backlog.parent / "tasks" / "2026-05-01_TASK-042.md"
        task.parent.mkdir(parents=True)

        handoff.write_text(
            "\n".join(
                [
                    "# Session Handoff",
                    "",
                    "## Current Focus",
                    "",
                    "- beta 0.4.0 workflow parser alignment",
                    "",
                    "## Work Status",
                    "",
                    "- TASK-041 Previous task: done",
                    "- TASK-042 Parser alignment: in_progress",
                    "- WF-042-06 Branch parser compatibility: in_progress",
                    "- TASK-043 Waiting on access: blocked",
                    "",
                    "## Next Actions",
                    "",
                    "- [ ] Continue parser work",
                ]
            )
            + "\n",
            encoding="utf-8",
        )
        backlog.write_text(
            "\n".join(
                [
                    "# 2026-05-01 작업 백로그",
                    "",
                    "## 작업 목록",
                    "",
                    "- [TASK-042 Parser alignment](./tasks/2026-05-01_TASK-042.md)",
                ]
            )
            + "\n",
            encoding="utf-8",
        )
        task.write_text(
            "\n".join(
                [
                    "# TASK-042 Parser alignment",
                    "",
                    "- 상태: in_progress",
                    "- 우선순위: high",
                ]
            )
            + "\n",
            encoding="utf-8",
        )

        handoff_data = parse_handoff(handoff)
        if handoff_data["current_baseline"] != "beta 0.4.0 workflow parser alignment":
            raise AssertionError(f"unexpected current baseline: {handoff_data['current_baseline']}")
        if handoff_data["in_progress_items"] != [
            "TASK-042 Parser alignment",
            "WF-042-06 Branch parser compatibility",
        ]:
            raise AssertionError(f"unexpected in-progress items: {handoff_data['in_progress_items']}")
        if handoff_data["blocked_items"] != ["TASK-043 Waiting on access"]:
            raise AssertionError(f"unexpected blocked items: {handoff_data['blocked_items']}")
        if handoff_data["recent_done_items"] != ["TASK-041 Previous task"]:
            raise AssertionError(f"unexpected done items: {handoff_data['recent_done_items']}")

        backlog_data = parse_backlog(backlog)
        if backlog_data["in_progress_items"] != ["TASK-042 Parser alignment"]:
            raise AssertionError(f"unexpected backlog in-progress items: {backlog_data['in_progress_items']}")
        if len(backlog_data["tasks"]) != 1:
            raise AssertionError(f"expected one linked task, got {backlog_data['tasks']}")


def check_legacy_format() -> None:
    example_root = SOURCE_ROOT / "examples" / "research_eval_hub"
    handoff_data = parse_handoff(example_root / "session_handoff.md")
    backlog_data = parse_backlog(example_root / "backlog" / "2026-04-19.md")
    if not handoff_data["in_progress_items"]:
        raise AssertionError("expected legacy handoff in-progress items")
    if not backlog_data["in_progress_items"]:
        raise AssertionError("expected legacy backlog in-progress items")


def main() -> int:
    check_branch_memory_format()
    check_legacy_format()
    print("Project docs branch memory parser check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
