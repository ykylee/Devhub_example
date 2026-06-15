#!/usr/bin/env python3
"""Smoke test the backlog-update prototype."""

from __future__ import annotations

import json
import re
import subprocess
import sys
import tempfile
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
SCRIPT_PATH = SOURCE_ROOT / "skills" / "backlog-update" / "scripts" / "run_backlog_update.py"

if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.common.output_contracts import validate_output_payload
from workflow_kit.common.paths import get_current_branch


def run_backlog_update(*, expect_success: bool, args: list[str]) -> tuple[int, dict[str, object]]:
    completed = subprocess.run(
        [sys.executable, str(SCRIPT_PATH), *args],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if expect_success and completed.returncode != 0:
        raise AssertionError(f"Expected backlog-update success but got {completed.returncode}: {completed.stderr}")
    if not expect_success and completed.returncode == 0:
        raise AssertionError("Expected backlog-update failure path but command succeeded.")
    return completed.returncode, json.loads(completed.stdout)


def main() -> int:
    example_root = SOURCE_ROOT / "examples" / "acme_delivery_platform"
    backlog_path = sorted((example_root / "backlog").glob("*.md"))[-1]

    # Case 1: Standard update without --apply
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_root = Path(temp_dir).resolve()
        temp_project_root = (temp_root / "project").resolve()
        temp_project_root.mkdir()

        current_branch = get_current_branch()
        temp_branch_root = (temp_project_root / current_branch).resolve()
        temp_branch_root.mkdir(parents=True)

        for relative_path in ("PROJECT_PROFILE.md", "work_backlog.md"):
            source_path = example_root / relative_path
            target_path = temp_project_root / relative_path
            target_path.write_text(source_path.read_text(encoding="utf-8"), encoding="utf-8")

        source_path = example_root / "session_handoff.md"
        target_path = temp_branch_root / "session_handoff.md"
        target_path.write_text(source_path.read_text(encoding="utf-8"), encoding="utf-8")

        temp_backlog_dir = (temp_branch_root / "backlog").resolve()
        temp_backlog_dir.mkdir()
        temp_backlog_path = (temp_backlog_dir / backlog_path.name).resolve()
        temp_backlog_path.write_text(backlog_path.read_text(encoding="utf-8"), encoding="utf-8")

        _, payload = run_backlog_update(
            expect_success=True,
            args=[
                "--project-profile-path",
                str(temp_project_root / "PROJECT_PROFILE.md"),
                "--daily-backlog-path",
                str(temp_backlog_path),
                "--task-name",
                "배송 상태 동기화 실패 대응 절차 문서 정리",
                "--task-brief",
                "runbook 및 handoff 반영 상태를 점검했다.",
                "--task-id",
                "TASK-021",
                "--mode",
                "update",
            ],
        )

        output_errors = validate_output_payload(payload, family="backlog_update")
        if output_errors:
            raise AssertionError(f"Backlog-update success payload violated output contract: {output_errors}")
        if payload["operation_type"] != "update_entry":
            raise AssertionError("Expected update_entry operation type.")
        if payload["status_recommendation"]["value"] != "in_progress":
            raise AssertionError("Expected conservative in_progress status recommendation.")
        if not payload["draft_entry"]:
            raise AssertionError("Expected non-empty backlog draft entry.")

        if "state.json" not in payload["state_cache_update_note"]:
            raise AssertionError("Expected backlog-update to include a state cache refresh note.")
        if "generate_workflow_state.py" not in payload["state_cache_refresh_command"]:
            raise AssertionError("Expected backlog-update to include a state cache refresh command.")
        if payload["state_cache_status"] != "refreshed":
            raise AssertionError(f"Expected backlog-update to refresh state.json automatically. Got: {payload['state_cache_status']}")

        state_path = (temp_branch_root / "state.json").resolve()
        if not state_path.exists():
            raise AssertionError(f"Expected state.json to exist at {state_path}, but it was not found.")

        state_payload = json.loads(state_path.read_text(encoding="utf-8"))
        expected_profile_path = str((temp_project_root / "PROJECT_PROFILE.md").resolve())
        # The path in state.json is now relative to workspace root (temp_project_root)
        actual_profile_rel = state_payload["source_of_truth"]["project_profile_path"]
        actual_profile_path = str((temp_project_root / actual_profile_rel).resolve())
        if actual_profile_path != expected_profile_path:
            raise AssertionError(f"Expected state.json to be refreshed from {expected_profile_path}, but got {actual_profile_path} (rel: {actual_profile_rel})")

    # Case 2: Update with --apply
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_root = Path(temp_dir).resolve()
        temp_project_root = (temp_root / "project").resolve()
        temp_project_root.mkdir()

        current_branch = get_current_branch()
        temp_branch_root = (temp_project_root / current_branch).resolve()
        temp_branch_root.mkdir(parents=True)

        for relative_path in ("PROJECT_PROFILE.md", "work_backlog.md"):
            source_path = example_root / relative_path
            target_path = temp_project_root / relative_path
            target_path.write_text(source_path.read_text(encoding="utf-8"), encoding="utf-8")

        source_path = example_root / "session_handoff.md"
        target_path = temp_branch_root / "session_handoff.md"
        target_path.write_text(source_path.read_text(encoding="utf-8"), encoding="utf-8")

        temp_backlog_dir = (temp_branch_root / "backlog").resolve()
        temp_backlog_dir.mkdir()
        temp_backlog_path = (temp_backlog_dir / backlog_path.name).resolve()
        temp_backlog_path.write_text(backlog_path.read_text(encoding="utf-8"), encoding="utf-8")

        _, apply_payload = run_backlog_update(
            expect_success=True,
            args=[
                "--project-profile-path",
                str(temp_project_root / "PROJECT_PROFILE.md"),
                "--daily-backlog-path",
                str(temp_backlog_path),
                "--work-backlog-index-path",
                str(temp_project_root / "work_backlog.md"),
                "--session-handoff-path",
                str(temp_project_root / "session_handoff.md"),
                "--task-name",
                "배송 상태 동기화 실패 대응 절차 문서 정리",
                "--task-brief",
                "검증 대기 상태라 차단으로 되돌렸다.",
                "--task-id",
                "TASK-021",
                "--mode",
                "update",
                "--status",
                "blocked",
                "--apply",
            ],
        )
        if apply_payload["apply_status"] != "applied":
            raise AssertionError("Expected backlog-update apply mode to report applied status.")
        backlog_text = temp_backlog_path.read_text(encoding="utf-8")
        if "- 상태: blocked" not in backlog_text:
            raise AssertionError("Expected apply mode to update the backlog task status in the target file.")
        handoff_text = (temp_branch_root / "session_handoff.md").read_text(encoding="utf-8")
        if "TASK-021 배송 상태 동기화 실패 대응 절차 문서 정리" not in handoff_text:
            raise AssertionError("Expected apply mode to keep the task visible in handoff.")
        blocked_section = handoff_text.split("- 현재 `blocked` 작업:", 1)[1]
        if "TASK-021 배송 상태 동기화 실패 대응 절차 문서 정리" not in blocked_section:
            raise AssertionError("Expected apply mode to move the task into the blocked handoff section.")

        resolved_written_paths = [str(Path(p).resolve()) for p in apply_payload["written_paths"]]
        if str(temp_backlog_path.resolve()) not in resolved_written_paths:
            raise AssertionError(f"Expected apply mode to report the written backlog path. Got: {resolved_written_paths}")

        work_backlog_text = (temp_project_root / "work_backlog.md").read_text(encoding="utf-8")
        backlog_link_count = sum(
            1
            for line in work_backlog_text.splitlines()
            if line.strip() == f"- [{backlog_path.stem} 작업 백로그](./backlog/{backlog_path.name})"
        )
        if backlog_link_count != 1:
            raise AssertionError("Expected apply mode to keep only one backlog index link per daily backlog file.")

    # Case 3: Duplicate link cleanup with --apply
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_root = Path(temp_dir).resolve()
        temp_project_root = (temp_root / "project").resolve()
        temp_project_root.mkdir()

        current_branch = get_current_branch()
        temp_branch_root = (temp_project_root / current_branch).resolve()
        temp_branch_root.mkdir(parents=True)

        for relative_path in ("PROJECT_PROFILE.md", "work_backlog.md"):
            source_path = example_root / relative_path
            target_path = temp_project_root / relative_path
            target_path.write_text(source_path.read_text(encoding="utf-8"), encoding="utf-8")

        source_path = example_root / "session_handoff.md"
        target_path = temp_branch_root / "session_handoff.md"
        target_path.write_text(source_path.read_text(encoding="utf-8"), encoding="utf-8")

        temp_backlog_dir = (temp_branch_root / "backlog").resolve()
        temp_backlog_dir.mkdir()
        temp_backlog_path = (temp_backlog_dir / backlog_path.name).resolve()
        temp_backlog_path.write_text(backlog_path.read_text(encoding="utf-8"), encoding="utf-8")
        index_path = (temp_project_root / "work_backlog.md").resolve()
        current_branch = get_current_branch()
        canonical_rel_path = f"./{current_branch}/backlog/{backlog_path.name}"
        index_path = (temp_project_root / "work_backlog.md").resolve()
        index_text = index_path.read_text(encoding="utf-8")

        # Replace the existing link (which might be in the old format) with a duplicate in the new format
        # First, find what's actually in the index
        old_link_pattern = rf"- \[{backlog_path.stem} 작업 백로그\]\(\.\/backlog\/{backlog_path.name}\)"
        new_link = f"- [{backlog_path.stem} 작업 백로그]({canonical_rel_path})"

        index_text = re.sub(old_link_pattern, f"{new_link}\n{new_link}", index_text)
        index_path.write_text(index_text, encoding="utf-8")

        _, apply_payload = run_backlog_update(
            expect_success=True,
            args=[
                "--project-profile-path",
                str(temp_project_root / "PROJECT_PROFILE.md"),
                "--daily-backlog-path",
                str(temp_backlog_path),
                "--work-backlog-index-path",
                str(index_path),
                "--session-handoff-path",
                str(temp_branch_root / "session_handoff.md"),
                "--task-name",
                "배송 상태 동기화 실패 대응 절차 문서 정리",
                "--task-brief",
                "중복 링크를 자동 정리하는지 확인한다.",
                "--task-id",
                "TASK-021",
                "--mode",
                "update",
                "--apply",
            ],
        )
        if apply_payload["apply_status"] != "applied":
            raise AssertionError("Expected apply mode to succeed while cleaning duplicate backlog links.")
        normalized_lines = [
            line.strip()
            for line in index_path.read_text(encoding="utf-8").splitlines()
            if line.strip().startswith("- [") and backlog_path.name in line
        ]
        if len(normalized_lines) != 1:
            raise AssertionError("Expected duplicate backlog links to be collapsed to one canonical entry.")

    # Case 4: Failure path (missing profile)
    failure_code, failure_payload = run_backlog_update(
        expect_success=False,
        args=[
            "--project-profile-path",
            "/tmp/missing-profile.md",
            "--task-name",
            "운영 허브 링크 무결성 재점검",
            "--task-brief",
            "새 runbook 링크 반영 여부를 확인한다.",
        ],
    )
    if failure_code == 0:
        raise AssertionError("Expected backlog-update failure for missing profile path.")
    output_errors = validate_output_payload(failure_payload, family="backlog_update")
    if output_errors:
        raise AssertionError(f"Backlog-update error payload violated output contract: {output_errors}")
    if failure_payload["error_code"] != "missing_required_document":
        raise AssertionError("Expected missing_required_document error code.")

    print("Backlog-update smoke check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
