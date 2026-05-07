#!/usr/bin/env python3
"""Simulation of Scenario B: Multi-Agent Feedback Loop."""

from __future__ import annotations

import json
import sys
from pathlib import Path

# Standard AI Workflow Library Discovery
SOURCE_ROOT = Path(__file__).resolve().parent
while not (SOURCE_ROOT / "workflow_kit").is_dir() and SOURCE_ROOT != SOURCE_ROOT.parent:
    SOURCE_ROOT = SOURCE_ROOT.parent
REPO_ROOT = SOURCE_ROOT.parent
if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.common.schemas import WorkerTask, WorkerResponse, Status


def main():
    print("=== Scenario B: Multi-Agent Feedback Loop Simulation ===\n")

    # 1. Orchestrator starts a task
    print("[Orchestrator] Step 1: Delegating feature implementation to code-worker.")
    task1 = WorkerTask(
        worker_id="code-worker",
        task_description="Implement Pydantic models for git conflict resolution.",
        input_files=["workflow-source/workflow_kit/common/schemas/git.py"],
        constraints=["Follow PEP 8."]
    )
    print(f"Delegation 1: {task1.worker_id} -> '{task1.task_description}'")

    # 2. Worker returns with a risk
    print("\n[Worker: code-worker] Working...")
    response1 = WorkerResponse(
        status=Status.WARNING,
        tool_version="1.0.0",
        summary="Implementation complete, but found a pre-existing lint error in the target file.",
        produced_artifacts=["workflow-source/workflow_kit/common/schemas/git.py"],
        risks_identified=["lint_failure: E501 line too long at line 42"],
        suggested_follow_up=["Run workflow-linter or a formatting tool."]
    )
    print(f"Response 1: Status={response1.status}, Risks={response1.risks_identified}")

    # 3. Orchestrator detects the risk and decides to pivot
    print("\n[Orchestrator] Step 2: Risk detected! Delegating lint fix to validation-worker.")
<<<<<<< HEAD
    
=======

>>>>>>> origin/codex/service-action-command
    if any("lint_failure" in r for r in response1.risks_identified):
        task2 = WorkerTask(
            worker_id="validation-worker",
            task_description="Fix the E501 lint error reported by code-worker.",
            input_files=response1.produced_artifacts,
            context_summary=f"Previous worker reported: {response1.risks_identified[0]}"
        )
        print(f"Delegation 2: {task2.worker_id} -> '{task2.task_description}'")
<<<<<<< HEAD
        
=======

>>>>>>> origin/codex/service-action-command
        # 4. Second worker resolves the risk
        print("\n[Worker: validation-worker] Running auto-formatter...")
        response2 = WorkerResponse(
            status=Status.OK,
            tool_version="1.0.0",
            summary="Lint error E501 resolved successfully.",
            produced_artifacts=response1.produced_artifacts
        )
        print(f"Response 2: Status={response2.status}")
<<<<<<< HEAD
    
=======

>>>>>>> origin/codex/service-action-command
    # 5. Final Orchestration Result
    print("\n[Orchestrator] Step 3: All tasks completed and risks mitigated.")
    print("=== Scenario B Simulation Complete ===")


if __name__ == "__main__":
    main()
