#!/usr/bin/env python3
"""Example of multi-agent delegation flow (Orchestration Demo)."""

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
    print("=== Phase 10: Multi-Agent Orchestration Demo ===\n")

    # 1. Orchestrator identifies a task
    print("[Orchestrator] Task identified: Audit Phase 9 documentation integrity.")
    
    # 2. Orchestrator prepares a delegation payload
    delegation = WorkerTask(
        worker_id="doc-worker",
        task_description="Verify that all markdown files in 'ai-workflow/core/' have valid links and no stale Phase 8 references.",
        input_files=["ai-workflow/core/read_only_mcp_transport_promotion.md", "ai-workflow/core/output_schema_guide.md"],
        constraints=["Do not modify files directly; suggest changes in the summary."],
        context_summary="Transitioning from Phase 9 to Phase 10 requires clean documentation baselines."
    )
    
    print(f"[Orchestrator] Delegating to {delegation.worker_id}...")
    print(json.dumps(delegation.model_dump(), indent=2, ensure_ascii=False))
    print("\n" + "-"*40 + "\n")

    # 3. Simulate Worker execution
    print("[Worker: doc-worker] Executing audit...")
    print("[Worker: doc-worker] Scanning files...")
    
    # 4. Worker returns a response
    response = WorkerResponse(
        status=Status.OK,
        tool_version="1.0.0",
        summary="Audit complete. Found 2 stale links in 'output_schema_guide.md' and 1 missing cross-reference in 'read_only_mcp_transport_promotion.md'.",
        produced_artifacts=[],
        risks_identified=["Legacy references to 'mcp/' (pre-migration) still exist in internal comments."],
        suggested_follow_up=["Update the 'mcp/' to 'mcp_servers/' in internal code comments via a regex tool."]
    )
    
    print("[Orchestrator] Received response from worker:")
    print(json.dumps(response.model_dump(), indent=2, ensure_ascii=False))
    
    print("\n[Orchestrator] Decision: Task successfully delegated and results integrated.")
    print("=== Demo Complete ===")

if __name__ == "__main__":
    main()
