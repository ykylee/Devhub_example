#!/usr/bin/env python3
"""Smoke test the code-index-update prototype."""

from __future__ import annotations

import json
import subprocess
import sys
import tempfile
from pathlib import Path

if str(Path(__file__).resolve().parents[1]) not in sys.path:
    sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from workflow_kit.common.output_contracts import validate_output_payload


REPO_ROOT = Path(__file__).resolve().parents[1]
SCRIPT_PATH = REPO_ROOT / "skills" / "code-index-update" / "scripts" / "run_code_index_update.py"


def run_index_update(example_name: str, changed_files: list[str], change_summary: str) -> dict[str, object]:
    example_root = REPO_ROOT / "examples" / example_name
    completed = subprocess.run(
        [
            sys.executable,
            str(SCRIPT_PATH),
            "--project-profile-path",
            str(example_root / "project_workflow_profile.md"),
            "--work-backlog-index-path",
            str(example_root / "work_backlog.md"),
            "--session-handoff-path",
            str(example_root / "session_handoff.md"),
            *sum([["--changed-file", item] for item in changed_files], []),
            "--change-summary",
            change_summary,
        ],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
        check=True,
    )
    return json.loads(completed.stdout)


def main() -> int:
    acme_payload = run_index_update(
        "acme_delivery_platform",
        [
            "app/jobs/delivery_sync.py",
            "docs/operations/runbooks/delivery-sync.md",
        ],
        "delivery sync 재시도 로직과 운영 runbook 동시 수정",
    )
    output_errors = validate_output_payload(acme_payload, family="code_index_update")
    if output_errors:
        raise AssertionError(f"Acme code-index-update payload violated output contract: {output_errors}")
    acme_priorities = set(acme_payload["priority_index_candidates"])
    if not any(item.endswith("docs/README.md") for item in acme_priorities):
        raise AssertionError("Acme example should elevate the declared document home as a priority index candidate.")
    if not any(item.endswith("docs/operations/README.md") for item in acme_priorities):
        raise AssertionError("Acme example should elevate the operations README candidate.")
    if not acme_payload["stale_index_warnings"]:
        raise AssertionError("Acme example should produce at least one stale index warning.")

    research_payload = run_index_update(
        "research_eval_hub",
        [
            "evals/pipelines/report_builder.py",
            "docs/evals/reports/release-report-v2.md",
        ],
        "evaluation report builder 로직과 release report 동시 수정",
    )
    output_errors = validate_output_payload(research_payload, family="code_index_update")
    if output_errors:
        raise AssertionError(f"Research code-index-update payload violated output contract: {output_errors}")
    research_candidates = set(research_payload["index_update_candidates"])
    if not any(item.endswith("docs/evals/README.md") for item in research_candidates):
        raise AssertionError("Research example should include the evals README candidate.")
    if not research_payload["document_structure_signals"]:
        raise AssertionError("Research example should include document structure signals.")

    workflow_meta_payload = run_index_update(
        "acme_delivery_platform",
        [
            "ai-workflow/project/session_handoff.md",
        ],
        "workflow 상태 문서만 수정",
    )
    if workflow_meta_payload["source_context"]["changed_files"]:
        raise AssertionError("Code-index-update should ignore ai-workflow metadata paths in changed_files.")

    with tempfile.TemporaryDirectory() as tmpdir:
        repo_root = Path(tmpdir) / "repo"
        workflow_project = repo_root / "ai-workflow" / "project"
        workflow_project.mkdir(parents=True, exist_ok=True)
        (repo_root / "docs" / "operations").mkdir(parents=True, exist_ok=True)
        (repo_root / "README.md").write_text("# Repo\n", encoding="utf-8")
        (repo_root / "docs" / "README.md").write_text("# Docs\n", encoding="utf-8")
        (repo_root / "docs" / "operations" / "README.md").write_text("# Operations\n", encoding="utf-8")
        (workflow_project / "project_workflow_profile.md").write_text(
            (
                "# Project Workflow Profile\n\n"
                "## 1. 프로젝트 개요\n\n"
                "- 프로젝트명:\n- `Temp Repo`\n\n"
                "## 2. 문서 구조\n\n"
                "- 문서 위키 홈:\n- `docs/README.md`\n"
                "- 운영 문서 위치:\n- `docs/operations/`\n"
                "- 백로그 위치:\n- `docs/operations/backlog/`\n"
                "- 세션 인계 문서 위치:\n- `docs/operations/session_handoff.md`\n"
                "- 환경 기록 위치:\n- `docs/operations/environments/`\n"
            ),
            encoding="utf-8",
        )
        completed = subprocess.run(
            [
                sys.executable,
                str(SCRIPT_PATH),
                "--project-profile-path",
                str(workflow_project / "project_workflow_profile.md"),
                "--changed-file",
                "docs/operations/runbooks/foo.md",
                "--change-summary",
                "runbook 수정",
            ],
            cwd=REPO_ROOT,
            capture_output=True,
            text=True,
            check=True,
        )
        payload = json.loads(completed.stdout)
        if not any(item.endswith("/docs/README.md") for item in payload["priority_index_candidates"]):
            raise AssertionError("Project docs should resolve from repository root, not ai-workflow/project.")
        if any("/ai-workflow/project/docs/" in item for item in payload["index_update_candidates"]):
            raise AssertionError("Project doc candidates should not resolve under ai-workflow/project/docs.")

    print("Code-index-update smoke check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
