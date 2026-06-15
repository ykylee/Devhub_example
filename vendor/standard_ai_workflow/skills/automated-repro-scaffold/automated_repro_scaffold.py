#!/usr/bin/env python3
import argparse
import sys
import os
import json
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[3]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit import __version__ as TOOL_VERSION
from workflow_kit.common.errors import build_error_result
# v0.6.5 stage_completion integration (runtime migration pilot)
from workflow_kit.common.contracts.stage_gate_runtime import (
    build_stage_completion, merge_into_result,
)

# v0.6.5 spec 보강 (commit 5b16517) — Stage Name 매핑
STAGE_NAME = "automated-repro-scaffold"
NEXT_STAGE = "validation-plan"  # repro script → validation-plan 으로 진행
ARTIFACTS_TEMPLATE = ["<repro_script_path>"]  # runtime 에서 args.output 으로 채움

def main():
    parser = argparse.ArgumentParser(description="Automated Bug Reproduction Scaffolder (Prototype)")
    parser.add_argument("--report", required=True, help="Path to the bug report file")
    parser.add_argument("--output", required=True, help="Path to save the generated reproduction script")

    args = parser.parse_args()
    source_context = {"report_path": args.report, "output_path": args.output}

    if not os.path.exists(args.report):
        result = build_error_result(
            tool_version=TOOL_VERSION,
            error=f"리포트 파일을 찾을 수 없다: {args.report}",
            error_code="report_file_not_found",
            warnings=[f"report file not found: {args.report}"],
            source_context=source_context,
        )
        # v0.6.5: stage_completion merge (status: error, gate 정지)
        result = merge_into_result(
            result,
            build_stage_completion(
                stage_name=STAGE_NAME,
                stage_status="error",
                artifacts=[args.output],
                next_stage=None,  # error 시 다음 stage 없음
                notes=[f"report file not found: {args.report}"],
            ),
        )
        print(json.dumps(result, ensure_ascii=False, indent=2))
        sys.exit(1)

    with open(args.report, "r", encoding="utf-8") as f:
        report_content = f.read()

    # Prototype logic: Simple pattern matching or just wrapping in a template
    # In a real scenario, this would involve more complex parsing or LLM calls

    repro_template = f"""import unittest
import sys

# Bug Report Context:
# {report_content.strip().replace('#', '-')}

class TestReproduction(unittest.TestCase):
    def test_reproduce_issue(self):
        \"\"\"
        This test is auto-generated to reproduce the reported issue.
        Modify this section to include the actual logic that triggers the bug.
        \"\"\"
        # TODO: Implement the actual reproduction logic based on the report
        print("\\n[INFO] Attempting to reproduce issue...")

        # Example of a failing assertion (expected bug state)
        # self.assertEqual(actual_result, expected_result, "Bug detected: results do not match")

        # For prototype demonstration, we just show it's working
        self.assertTrue(True)

if __name__ == "__main__":
    print("Running auto-generated reproduction script...")
    unittest.main()
"""

    try:
        os.makedirs(os.path.dirname(os.path.abspath(args.output)), exist_ok=True)
        with open(args.output, "w", encoding="utf-8") as f:
            f.write(repro_template)

        result = {
            "status": "success",
            "repro_script_path": args.output,
            "execution_command": f"python3 {args.output}",
            "tool_version": TOOL_VERSION,
            "warnings": ["이것은 자동 생성된 프로토타입입니다. 수동 조정이 필요할 수 있습니다."],
            "source_context": source_context,
        }
        # v0.6.5: stage_completion merge (status: ok, validation-plan 으로 진행)
        result = merge_into_result(
            result,
            build_stage_completion(
                stage_name=STAGE_NAME,
                stage_status="ok",  # legacy "success" → "ok" 매핑
                artifacts=[args.output],
                next_stage=NEXT_STAGE,
                notes=[f"repro script generated: {args.output}"],
            ),
        )
        print(json.dumps(result, indent=2, ensure_ascii=False))

    except Exception as e:
        result = build_error_result(
            tool_version=TOOL_VERSION,
            error=f"재현 스크립트 작성 중 오류가 발생했다: {str(e)}",
            error_code="repro_write_failed",
            warnings=[f"repro write failed: {str(e)[:200]}"],
            source_context=source_context,
        )
        # v0.6.5: stage_completion merge (status: error)
        result = merge_into_result(
            result,
            build_stage_completion(
                stage_name=STAGE_NAME,
                stage_status="error",
                artifacts=[args.output],
                next_stage=None,
                notes=[f"repro write failed: {str(e)[:100]}"],
            ),
        )
        print(json.dumps(result, ensure_ascii=False, indent=2))
        sys.exit(1)

if __name__ == "__main__":
    main()
