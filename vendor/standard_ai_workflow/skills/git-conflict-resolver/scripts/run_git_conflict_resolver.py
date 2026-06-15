#!/usr/bin/env python3
"""Enhanced runner for the git-conflict-resolver skill with contextual analysis."""

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
from workflow_kit.common.project_docs import parse_handoff
from workflow_kit.common.schemas import GitConflictResolverOutput, Status, ConflictPoint, ResolutionStrategy


def find_conflicts_in_file(file_path: Path) -> list[ConflictPoint]:
    if not file_path.exists():
        return []
    
    content = file_path.read_text(encoding="utf-8")
    pattern = re.compile(r"<<<<<<< .*?\n(.*?)\n=======\n(.*?)\n>>>>>>> .*?\n?", re.DOTALL)
    
    conflicts = []
    for match in pattern.finditer(content):
        ours, theirs = match.groups()
        conflicts.append(ConflictPoint(
            file_path=str(file_path),
            our_content=ours,
            their_content=theirs,
            resolution_strategy=ResolutionStrategy.MANUAL,
            resolution_note="Initial detection."
        ))
    return conflicts


def resolve_conflict_contextually(conflict: ConflictPoint, context_keywords: list[str]) -> ConflictPoint:
    """Attempt to resolve a conflict by matching content against session context keywords."""
    our_matches = any(kw.lower() in conflict.our_content.lower() for kw in context_keywords if kw)
    their_matches = any(kw.lower() in conflict.their_content.lower() for kw in context_keywords if kw)

    if our_matches and not their_matches:
        conflict.resolution_strategy = ResolutionStrategy.OURS
        conflict.resolution_note = f"Our change contains context keywords: {', '.join([kw for kw in context_keywords if kw.lower() in conflict.our_content.lower()])}"
    elif their_matches and not our_matches:
        conflict.resolution_strategy = ResolutionStrategy.THEIRS
        conflict.resolution_note = f"Incoming change contains context keywords: {', '.join([kw for kw in context_keywords if kw.lower() in conflict.their_content.lower()])}"
    elif our_matches and their_matches:
        conflict.resolution_strategy = ResolutionStrategy.MERGE
        conflict.resolution_note = "Both sides contain context keywords. Intelligent merge required."
    else:
        # Default to manual if no keywords match
        conflict.resolution_strategy = ResolutionStrategy.MANUAL
        conflict.resolution_note = "No context keywords matched. Requires manual intervention."
    
    return conflict


def main() -> int:
    parser = argparse.ArgumentParser(description="Enhanced git merge conflict resolver.")
    parser.add_argument("--file", action="append", dest="files", help="Files to check for conflicts")
    parser.add_argument("--handoff-path", help="Path to session_handoff.md for context")
    parser.add_argument("--apply", action="store_true", help="Apply resolutions (placeholder)")
    args = parser.parse_args()

    # 1. Extract context keywords from handoff
    context_keywords = []
    if args.handoff_path:
        handoff_path = Path(args.handoff_path).resolve()
        if handoff_path.exists():
            try:
                handoff_data = parse_handoff(handoff_path)
                # Extract keywords from 'current_axis' and 'in_progress_items'
                axis = handoff_data.get("current_axis", "")
                if axis:
                    context_keywords.extend(str(axis).split())
                
                for item in handoff_data.get("in_progress_items", []):
                    # Clean up TASK-xxx prefix and split
                    clean_item = re.sub(r"TASK-\d+", "", str(item)).strip()
                    context_keywords.extend(clean_item.split())
            except Exception as e:
                print(f"Warning: Failed to parse handoff for context: {e}", file=sys.stderr)

    # 2. Find and resolve conflicts
    all_conflicts = []
    if args.files:
        for f in args.files:
            file_path = Path(f).resolve()
            conflicts = find_conflicts_in_file(file_path)
            for c in conflicts:
                resolved_c = resolve_conflict_contextually(c, context_keywords)
                all_conflicts.append(resolved_c)

    resolved_count = len([c for c in all_conflicts if c.resolution_strategy != ResolutionStrategy.MANUAL])

    output = GitConflictResolverOutput(
        status=Status.OK,
        tool_version=TOOL_VERSION,
        conflict_count=len(all_conflicts),
        resolved_count=resolved_count,
        resolution_summary=f"Processed {len(all_conflicts)} conflicts. Automatically resolved {resolved_count} based on handoff context.",
        conflicts=all_conflicts,
        source_context={
            "checked_files": str(args.files),
            "handoff_path": str(args.handoff_path) if args.handoff_path else "N/A",
            "context_keywords": str(context_keywords)
        }
    )

    result = output.model_dump()
        # v0.6.6 follow-up: stage_completion merge (pilot template)
        result = merge_into_result(
            result,
            build_stage_completion(
                stage_name="git-conflict-resolver",
                stage_status="ok" if result.get("status") in ("ok", "success") else "warning" if result.get("status") == "warning" else "error",
                artifacts=["(resolved_conflicts)"],
                next_stage=None,
                notes=[result.get("summary", "")[:200]] if result.get("summary") else [],
            ),
        )
    print(json.dumps(output.model_dump(), ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
