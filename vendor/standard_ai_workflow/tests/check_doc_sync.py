#!/usr/bin/env python3
"""Smoke test the doc-sync prototype."""

from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
SCRIPT_PATH = SOURCE_ROOT / "skills" / "doc-sync" / "scripts" / "run_doc_sync.py"

if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.common.output_contracts import validate_output_payload


def run_doc_sync(*, expect_success: bool, args: list[str]) -> tuple[int, dict[str, object]]:
    completed = subprocess.run(
        [sys.executable, str(SCRIPT_PATH), *args],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if expect_success and completed.returncode != 0:
        raise AssertionError(f"Expected doc-sync success but got {completed.returncode}: {completed.stderr}")
    if not expect_success and completed.returncode == 0:
        raise AssertionError("Expected doc-sync failure path but command succeeded.")
    return completed.returncode, json.loads(completed.stdout)


def main() -> int:
    example_root = SOURCE_ROOT / "examples" / "acme_delivery_platform"
    latest_backlog = sorted((example_root / "backlog").glob("*.md"))[-1]

    _, payload = run_doc_sync(
        expect_success=True,
        args=[
            "--project-profile-path",
            str(example_root / "PROJECT_PROFILE.md"),
            "--session-handoff-path",
            str(example_root / "session_handoff.md"),
            "--work-backlog-index-path",
            str(example_root / "work_backlog.md"),
            "--latest-backlog-path",
            str(latest_backlog),
            "--changed-file",
            "docs/operations/runbooks/delivery-sync.md",
            "--changed-file",
            "app/jobs/delivery_sync.py",
            "--change-summary",
            "delivery sync 재시도 로직과 운영 runbook 동시 수정",
        ],
    )
    output_errors = validate_output_payload(payload, family="doc_sync")
    if output_errors:
        raise AssertionError(f"Doc-sync success payload violated output contract: {output_errors}")
    if payload["status"] != "ok":
        raise AssertionError("Expected ok status for doc-sync success path.")
    if not payload["impacted_documents"]:
        raise AssertionError("Expected impacted document candidates.")
    if not payload["recommended_review_order"]:
        raise AssertionError("Expected recommended review order.")
    if not payload["follow_up_actions"]:
        raise AssertionError("Expected follow-up actions.")
    if not any("runbook" in item or "operations" in item for item in payload["impacted_documents"]):
        raise AssertionError("Expected runbook/operations document to appear in impacted documents.")

    _, workflow_meta_payload = run_doc_sync(
        expect_success=True,
        args=[
            "--project-profile-path",
            str(example_root / "PROJECT_PROFILE.md"),
            "--session-handoff-path",
            str(example_root / "session_handoff.md"),
            "--work-backlog-index-path",
            str(example_root / "work_backlog.md"),
            "--latest-backlog-path",
            str(latest_backlog),
            "--changed-file",
            "ai-workflow/memory/active/session_handoff.md",
            "--change-summary",
            "workflow 상태 문서만 수정",
        ],
    )
    if any("ai-workflow/" in item for item in workflow_meta_payload["impacted_documents"]):
        raise AssertionError("Workflow metadata paths should be ignored in impacted project documents.")
    if any("ai-workflow/" in item for item in workflow_meta_payload["hub_update_candidates"]):
        raise AssertionError("Workflow metadata paths should be ignored in project hub candidates.")

    failure_code, failure_payload = run_doc_sync(
        expect_success=False,
        args=[
            "--project-profile-path",
            str(example_root / "PROJECT_PROFILE.md"),
        ],
    )
    if failure_code == 0:
        raise AssertionError("Expected doc-sync failure for missing change input.")
    output_errors = validate_output_payload(failure_payload, family="doc_sync")
    if output_errors:
        raise AssertionError(f"Doc-sync error payload violated output contract: {output_errors}")
    if failure_payload["error_code"] != "missing_change_input":
        raise AssertionError("Expected missing_change_input error code.")

    print("Doc-sync smoke check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
