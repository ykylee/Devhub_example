#!/usr/bin/env python3
"""Runner for the workflow-linter skill, updated to use Pydantic contracts."""

from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parents[4]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit import __version__ as TOOL_VERSION
from workflow_kit.common.errors import build_error_result
from workflow_kit.common.contracts.stage_gate_runtime import build_stage_completion, merge_into_result
from workflow_kit.common.paths import resolve_existing_path, workflow_branch_dir, workflow_memory_dir
from workflow_kit.common.linter import check_workflow_consistency, check_maturity_consistency
from workflow_kit.common.metadata import load_config  # v0.7.15+: excluded_paths from config
from workflow_kit.common.schemas import WorkflowLinterOutput, Status


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Lint workflow documents for consistency.")
    parser.add_argument("--project-profile-path", required=True)
    parser.add_argument("--state-json-path")
    parser.add_argument("--session-handoff-path")
    parser.add_argument("--latest-backlog-path")
    parser.add_argument("--maturity", action="store_true", help="Check project maturity matrix")
    parser.add_argument("--apply", action="store_true", help="Attempt to auto-fix some issues")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    
    try:
        project_profile_path = resolve_existing_path(args.project_profile_path)
        project_root = project_profile_path.parent.parent.parent # Assuming standard structure: <root>/ai-workflow/memory/active/PROJECT_PROFILE.md
        
        branch_dir = workflow_branch_dir(project_profile_path)
        
        state_json_path = (
            resolve_existing_path(args.state_json_path) if args.state_json_path
            else (workflow_memory_dir(project_profile_path) / "state.json").resolve()
        )
        session_handoff_path = (
            resolve_existing_path(args.session_handoff_path) if args.session_handoff_path
            else (branch_dir / "session_handoff.md").resolve()
        )
        
        # Resolve latest backlog path
        latest_backlog_path = None
        if args.latest_backlog_path:
            latest_backlog_path = resolve_existing_path(args.latest_backlog_path)
        elif state_json_path.exists():
            try:
                state_data = json.loads(state_json_path.read_text(encoding="utf-8"))
                raw_backlog = state_data["source_of_truth"]["latest_backlog_path"]
                # Resolve relative to branch_dir
                latest_backlog_path = (branch_dir / raw_backlog).resolve()
            except Exception:
                pass

        if not latest_backlog_path:
            # Fallback: newest in backlog/
            backlog_dir = workflow_branch_dir(project_profile_path) / "backlog"
            backlog_files = sorted(backlog_dir.glob("*.md"), reverse=True)
            if backlog_files:
                latest_backlog_path = backlog_files[0]
            else:
                latest_backlog_path = (backlog_dir / "tasks").resolve() # placeholder

        source_context = {
            "project_profile_path": str(project_profile_path),
            "state_json_path": str(state_json_path),
            "session_handoff_path": str(session_handoff_path),
            "latest_backlog_path": str(latest_backlog_path),
        }

        # v0.7.15+: [tool.workflow-doctor] excluded_paths load → check_workflow_consistency
        config = load_config(project_root)
        excluded_paths = config.excluded_paths

        # 1. Workflow Consistency (Docs)
        linter_result = check_workflow_consistency(
            state_json_path=state_json_path,
            handoff_path=session_handoff_path,
            latest_backlog_path=latest_backlog_path,
            excluded_paths=excluded_paths,
        )

        if linter_result.get("status") == "error":
             result = build_error_result(
                tool_version=TOOL_VERSION,
                error=linter_result.get("description", "Unknown linter error"),
                error_code=linter_result.get("error_code", "linter_failed"),
                warnings=linter_result.get("warnings", []),
                source_context=source_context
            )
             print(json.dumps(result, ensure_ascii=False, indent=2))
             return 1

        # 2. Maturity Consistency (Optional)
        if args.maturity:
            matrix_path = project_root / "ai-workflow/core/maturity_matrix.json"
            roadmap_path = project_root / "ai-workflow/core/workflow_kit_roadmap.md"
            maturity_result = check_maturity_consistency(matrix_path, roadmap_path, project_root)
            if maturity_result.get("status") == "issues_found":
                linter_result["issues"].extend(maturity_result["issues"])
                linter_result["warnings"].extend(maturity_result["warnings"])
                # Update summary
                linter_result["summary"]["total_issues"] = len(linter_result["issues"])

        # 3. Auto-fix (Optional)
        written_paths = []
        if args.apply and linter_result["issues"]:
            # Basic auto-fix for task status mismatch in state.json
            modified_state = False
            if state_json_path.exists():
                state_data = json.loads(state_json_path.read_text(encoding="utf-8"))
                for issue in linter_result["issues"]:
                    if issue["code"] == "task_status_mismatch" and "state.json" in issue["description"]:
                        task_match = re.search(r"Task (TASK-\d+)", issue["description"])
                        if task_match:
                            task_id = task_match.group(1)
                            if "session" not in state_data: state_data["session"] = {}
                            if "in_progress_items" not in state_data["session"]: state_data["session"]["in_progress_items"] = []
                            
                            if task_id not in [t.split()[0] for t in state_data["session"]["in_progress_items"]]:
                                state_data["session"]["in_progress_items"].append(task_id)
                                modified_state = True
                
                if modified_state:
                    state_json_path.write_text(json.dumps(state_data, indent=2, ensure_ascii=False), encoding="utf-8")
                    written_paths.append(str(state_json_path))

        # Final Output
        output_model = WorkflowLinterOutput(
            status=Status.OK if not linter_result["issues"] else Status.WARNING,
            tool_version=TOOL_VERSION,
            issues=linter_result["issues"],
            warnings=linter_result["warnings"],
            summary=linter_result["summary"],
            source_context=source_context
        )
        
        result = output_model.model_dump()
        # v0.6.6 follow-up: stage_completion merge (pilot template)
        # v0.7.15 fix: result["summary"] 가 dict (linter summary) 면 str() 변환 후 [:200]
        summary_value = result.get("summary", "")
        summary_str = str(summary_value)[:200] if summary_value else ""
        result = merge_into_result(
            result,
            build_stage_completion(
                stage_name="workflow-linter",
                stage_status="ok" if result.get("status") in ("ok", "success") else "warning" if result.get("status") == "warning" else "error",
                artifacts=["(workflow_linter_report)"],
                next_stage=None,
                notes=[summary_str] if summary_str else [],
            ),
        )
        print(json.dumps(result, ensure_ascii=False, indent=2))
        return 0

    except Exception as exc:
        result = build_error_result(
            tool_version=TOOL_VERSION,
            error=str(exc),
            error_code="workflow_linter_runtime_error",
            warnings=[],
            source_context={"exception_type": type(exc).__name__}
        )
        print(json.dumps(result, ensure_ascii=False, indent=2))
        return 1


if __name__ == "__main__":
    sys.exit(main())
