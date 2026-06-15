#!/usr/bin/env python3
"""Standardized runner for the robust-patcher skill."""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[4]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit import __version__ as TOOL_VERSION
from workflow_kit.common.errors import build_error_result
from workflow_kit.common.patching import apply_robust_patch


def main() -> int:
    parser = argparse.ArgumentParser(description="Run the robust-patcher standardized prototype.")
    parser.add_argument("--file", required=True, help="Target file to patch.")
    parser.add_argument("--patch-file", required=True, help="File containing SEARCH/REPLACE blocks.")
    args = parser.parse_args()

    source_context = {
        "file": args.file,
        "patch_file": args.patch_file,
    }

    try:
        file_path = Path(args.file).expanduser().resolve()
        patch_file_path = Path(args.patch_file).expanduser().resolve()
        
        if not patch_file_path.exists():
            result = build_error_result(
                tool_version=TOOL_VERSION,
                error="패치 파일을 찾을 수 없다.",
                error_code="patch_file_missing",
                warnings=[],
                source_context=source_context | {"missing_path": str(patch_file_path)},
            )
            print(json.dumps(result, ensure_ascii=False, indent=2))
            return 1

        patch_content = patch_file_path.read_text(encoding="utf-8")
        success, message = apply_robust_patch(file_path, patch_content)

        if not success:
            result = build_error_result(
                tool_version=TOOL_VERSION,
                error=message,
                error_code="patch_application_failed",
                warnings=[message],
                source_context=source_context,
            )
            print(json.dumps(result, ensure_ascii=False, indent=2))
            return 1

        result = {
            "status": "ok",
            "tool_version": TOOL_VERSION,
            "message": message,
            "file_path": str(file_path),
            "warnings": [],
        }
        print(json.dumps(result, ensure_ascii=False, indent=2))
        return 0

    except Exception as exc:
        result = build_error_result(
            tool_version=TOOL_VERSION,
            error=f"패치 적용 중 예기치 않은 오류 발생: {exc}",
            error_code="patch_engine_runtime_error",
            warnings=[],
            source_context=source_context | {"exception_type": type(exc).__name__},
        )
        print(json.dumps(result, ensure_ascii=False, indent=2))
        return 1


if __name__ == "__main__":
    sys.exit(main())
