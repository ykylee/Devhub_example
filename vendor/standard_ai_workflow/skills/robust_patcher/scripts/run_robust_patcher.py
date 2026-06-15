#!/usr/bin/env python3
"""v0.6.6 신규 — robust-patcher skill 의 runtime entry point.

pilot + batch 와 동일한 stage_completion integration.
patch_engine.py 의 library function 사용.
"""
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[3]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit import __version__ as TOOL_VERSION
from workflow_kit.common.errors import build_error_result
# v0.6.6 follow-up: stage_completion integration
from workflow_kit.common.contracts.stage_gate_runtime import build_stage_completion, merge_into_result

# v0.6.5 spec §5.2 매핑
STAGE_NAME = "robust-patcher"
NEXT_STAGE = "validation-plan"  # patch 적용 후 validation-plan 으로 진행
ARTIFACTS_TEMPLATE = ["<patched_file>"]  # runtime 에서 args.target 으로 채움


def main() -> int:
    parser = argparse.ArgumentParser(description="Robust Patcher (v0.6.6 신규 runtime)")
    parser.add_argument("--target", required=True, help="Path to the file to patch")
    parser.add_argument("--search", required=True, help="Search pattern (text or regex)")
    parser.add_argument("--replace", required=True, help="Replacement text")
    parser.add_argument("--regex", action="store_true", help="Treat --search as regex")
    args = parser.parse_args()

    source_context = {
        "target_path": args.target,
        "search": args.search[:100],
        "replace": args.replace[:100],
        "regex": args.regex,
    }

    try:
        # patch_engine.py 의 library function 호출 (v0.6.6 stub — 실제 구현은 patch_engine.py 참조)
        from workflow_kit.skills.robust_patcher.scripts.patch_engine import apply_patch

        patched = apply_patch(
            target_path=args.target,
            search=args.search,
            replace=args.replace,
            regex=args.regex,
        )

        result = {
            "status": "ok",
            "summary": f"Patched {args.target} successfully",
            "patched_path": args.target,
            "applied": patched.get("applied", True),
            "tool_version": TOOL_VERSION,
            "warnings": [],
            "source_context": source_context,
        }

        # v0.6.6: stage_completion merge
        result = merge_into_result(
            result,
            build_stage_completion(
                stage_name=STAGE_NAME,
                stage_status="ok",
                artifacts=[args.target],
                next_stage=NEXT_STAGE,
                notes=[result["summary"][:200]],
            ),
        )
        print(json.dumps(result, ensure_ascii=False, indent=2))
        return 0

    except Exception as e:
        result = build_error_result(
            tool_version=TOOL_VERSION,
            error=f"patch 적용 중 오류: {str(e)}",
            error_code="patch_failed",
            warnings=[f"patch failed: {str(e)[:200]}"],
            source_context=source_context,
        )
        # v0.6.6: stage_completion merge (status: error)
        result = merge_into_result(
            result,
            build_stage_completion(
                stage_name=STAGE_NAME,
                stage_status="error",
                artifacts=[args.target],
                next_stage=None,
                notes=[f"patch failed: {str(e)[:100]}"],
            ),
        )
        print(json.dumps(result, ensure_ascii=False, indent=2))
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
