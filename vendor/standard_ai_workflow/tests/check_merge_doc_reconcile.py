#!/usr/bin/env python3
"""Smoke test the merge-doc-reconcile prototype."""

from __future__ import annotations

import json
import subprocess
import sys
import tempfile
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
SCRIPT_PATH = SOURCE_ROOT / "skills" / "merge-doc-reconcile" / "scripts" / "run_merge_doc_reconcile.py"

if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.common.output_contracts import validate_output_payload


def run_merge_doc_reconcile(*, expect_success: bool, args: list[str]) -> tuple[int, dict[str, object]]:
    completed = subprocess.run(
        [sys.executable, str(SCRIPT_PATH), *args],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if expect_success and completed.returncode != 0:
        raise AssertionError(
            f"Expected merge-doc-reconcile success but got {completed.returncode}: {completed.stderr}"
        )
    if not expect_success and completed.returncode == 0:
        raise AssertionError("Expected merge-doc-reconcile failure path but command succeeded.")
    return completed.returncode, json.loads(completed.stdout)


def main() -> int:
    example_root = SOURCE_ROOT / "examples" / "acme_delivery_platform"
    latest_backlog = sorted((example_root / "backlog").glob("*.md"))[-1]
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_root = Path(temp_dir)
        from workflow_kit.common.paths import workflow_branch_dir, get_current_branch

        temp_project_root = temp_root / "project"
        temp_project_root.mkdir()

        # In the new architecture, profile is in docs/ or memory/ root,
        # but session_handoff and backlog are in branch-specific dirs.
        profile_path = temp_project_root / "PROJECT_PROFILE.md"
        profile_path.write_text((example_root / "PROJECT_PROFILE.md").read_text(encoding="utf-8"), encoding="utf-8")

        temp_work_backlog = temp_project_root / "work_backlog.md"
        temp_work_backlog.write_text((example_root / "work_backlog.md").read_text(encoding="utf-8"), encoding="utf-8")

        # Get branch dir using the same logic as the tool
        branch_name = get_current_branch()
        branch_dir = temp_project_root / branch_name
        branch_dir.mkdir(parents=True)

        temp_handoff = branch_dir / "session_handoff.md"
        temp_handoff.write_text((example_root / "session_handoff.md").read_text(encoding="utf-8"), encoding="utf-8")

        temp_backlog_dir = branch_dir / "backlog"
        temp_backlog_dir.mkdir()
        temp_latest_backlog = temp_backlog_dir / latest_backlog.name
        temp_latest_backlog.write_text(latest_backlog.read_text(encoding="utf-8"), encoding="utf-8")

        _, payload = run_merge_doc_reconcile(
            expect_success=True,
            args=[
                "--project-profile-path",
                str(profile_path),
                "--merge-result-summary",
                "delivery sync 관련 문서와 코드 변경이 병합됐다.",
                "--session-handoff-path",
                str(temp_handoff),
                "--work-backlog-index-path",
                str(temp_work_backlog),
                "--latest-backlog-path",
                str(temp_latest_backlog),
                "--changed-file",
                "docs/operations/runbooks/delivery-sync.md",
                "--changed-file",
                "app/jobs/delivery_sync.py",
            ],
        )
    output_errors = validate_output_payload(payload, family="merge_doc_reconcile")
    if output_errors:
        raise AssertionError(f"Merge-doc-reconcile success payload violated output contract: {output_errors}")
    if payload["status"] != "ok":
        raise AssertionError("Expected ok status for merge-doc-reconcile success path.")
    if not payload["reconcile_targets"]:
        raise AssertionError("Expected reconcile targets.")
    if not payload["reconfirmation_points"]:
        raise AssertionError("Expected reconfirmation points.")
    if not payload["draft_reconcile_notes"]:
        raise AssertionError("Expected draft reconcile notes.")
    if not payload["validation_follow_up"]:
        raise AssertionError("Expected validation follow-up note.")
    if "state.json" not in payload["state_cache_update_note"]:
        raise AssertionError("Expected merge-doc-reconcile to include a state cache refresh note.")
        if "generate_workflow_state.py" not in payload["state_cache_refresh_command"]:
            raise AssertionError("Expected merge-doc-reconcile to include a state cache refresh command.")
        if payload["state_cache_status"] != "refreshed":
            raise AssertionError("Expected merge-doc-reconcile to refresh state.json automatically.")
        state_path = Path(payload["state_cache_refresh_command"].split("--output-path ", 1)[1].split(" --", 1)[0])
        if not state_path.exists():
            raise AssertionError("Expected merge-doc-reconcile automatic state.json refresh to write the state cache file.")
        state_payload = json.loads(state_path.read_text(encoding="utf-8"))
        if state_payload["source_of_truth"]["latest_backlog_path"] != str(temp_latest_backlog):
            raise AssertionError("Expected state.json to track the refreshed latest backlog path.")

        _, apply_payload = run_merge_doc_reconcile(
            expect_success=True,
            args=[
                "--project-profile-path",
                str(profile_path),
                "--merge-result-summary",
                "delivery sync 관련 문서와 코드 변경이 병합됐다.",
                "--session-handoff-path",
                str(temp_handoff),
                "--work-backlog-index-path",
                str(temp_work_backlog),
                "--latest-backlog-path",
                str(temp_latest_backlog),
                "--changed-file",
                "docs/operations/runbooks/delivery-sync.md",
                "--changed-file",
                "app/jobs/delivery_sync.py",
                "--apply",
            ],
        )
        if apply_payload["apply_status"] != "applied":
            raise AssertionError("Expected merge-doc-reconcile apply mode to report applied status.")
        handoff_text = temp_handoff.read_text(encoding="utf-8")
        if "[merge-doc-reconcile]" not in handoff_text:
            raise AssertionError("Expected merge-doc-reconcile apply mode to append reconcile notes to handoff.")
        if str(temp_handoff) not in apply_payload["written_paths"]:
            raise AssertionError("Expected merge-doc-reconcile apply mode to report the written handoff path.")

        _, workflow_meta_payload = run_merge_doc_reconcile(
            expect_success=True,
            args=[
                "--project-profile-path",
                str(profile_path),
                "--merge-result-summary",
                "workflow 상태 문서만 병합됐다.",
                "--session-handoff-path",
                str(temp_handoff),
                "--work-backlog-index-path",
                str(temp_work_backlog),
                "--latest-backlog-path",
                str(temp_latest_backlog),
                "--changed-file",
                f"ai-workflow/memory/{branch_name}/session_handoff.md",
            ],
        )
        if workflow_meta_payload["source_context"]["changed_files"]:
            raise AssertionError("Merge-doc-reconcile should ignore ai-workflow metadata paths in changed_files.")

    failure_code, failure_payload = run_merge_doc_reconcile(
        expect_success=False,
        args=[
            "--project-profile-path",
            "/tmp/missing-project-profile.md",
            "--merge-result-summary",
            "병합 후 문서 상태를 재정리한다.",
        ],
    )
    if failure_code == 0:
        raise AssertionError("Expected merge-doc-reconcile failure for missing required document.")
    output_errors = validate_output_payload(failure_payload, family="merge_doc_reconcile")
    if output_errors:
        raise AssertionError(f"Merge-doc-reconcile error payload violated output contract: {output_errors}")
    if failure_payload["error_code"] != "missing_required_document":
        raise AssertionError("Expected missing_required_document error code.")

    print("Merge-doc-reconcile smoke check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
