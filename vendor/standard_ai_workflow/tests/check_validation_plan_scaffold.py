#!/usr/bin/env python3
"""Smoke test for the validation-plan skill's --scaffold mode."""

from __future__ import annotations

import json
import subprocess
import sys
import tempfile
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
SCRIPT_PATH = SOURCE_ROOT / "skills" / "validation-plan" / "scripts" / "run_validation_plan.py"

if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))


def main() -> int:
    with tempfile.TemporaryDirectory() as tmpdir:
        tmp_path = Path(tmpdir)
        project_root = tmp_path / "project"
        project_root.mkdir()

        # 1. 테스트용 프로젝트 문서 준비
        profile_path = project_root / "PROJECT_PROFILE.md"
        profile_path.write_text(
            "# Project Profile\n\n- 프로젝트명: Test Project\n- 빠른 테스트: python3 -m unittest discover tests\n",
            encoding="utf-8"
        )

        handoff_path = project_root / "session_handoff.md"
        handoff_path.write_text(
            "# 세션 인계 문서\n\n## 1. 현재 작업 요약\n- 최종 수정일: 2026-01-01\n\n## 현재 세션 운영 메모\n- 기존 메모\n\n## 다음에 읽을 문서\n- [README.md](./README.md)\n",
            encoding="utf-8"
        )

        # 2. validation-plan --scaffold 실행
        task_id = "TASK-123"
        completed = subprocess.run(
            [
                sys.executable,
                str(SCRIPT_PATH),
                "--project-profile-path", str(profile_path),
                "--session-handoff-path", str(handoff_path),
                "--changed-file", "app/main.py",
                "--change-summary", "신규 로직 추가",
                "--scaffold",
                "--task-id", task_id
            ],
            capture_output=True,
            text=True,
            check=False
        )

        if completed.returncode != 0:
            print(f"validation-plan failed with return code {completed.returncode}")
            print(f"stdout: {completed.stdout}")
            print(f"stderr: {completed.stderr}")
            return 1

        payload = json.loads(completed.stdout)

        # 3. 결과 검증
        if payload.get("scaffold_status") != "created":
            print(f"Expected scaffold_status 'created', got {payload.get('scaffold_status')}")
            return 1

        scaffold_path = Path(payload.get("scaffold_path"))
        if not scaffold_path.exists():
            print(f"Scaffold file not created at {scaffold_path}")
            return 1

        if f"repro_{task_id.lower().replace('-', '_')}.py" not in scaffold_path.name:
            print(f"Unexpected scaffold filename: {scaffold_path.name}")
            return 1

        # 파일 내용 검증
        scaffold_content = scaffold_path.read_text(encoding="utf-8")
        if f"Validation script for {task_id}" not in scaffold_content:
            print("Task ID not found in scaffold content.")
            return 1
        if "python3 -m unittest discover tests" not in scaffold_content:
            print("Recommended command not found in scaffold content.")
            return 1

        # handoff 갱신 검증
        handoff_content = handoff_path.read_text(encoding="utf-8")
        if scaffold_path.name not in handoff_content:
            print(f"Scaffold path link not found in handoff: {scaffold_path.name}")
            return 1
        if "[validation-plan]" not in handoff_content:
            print("validation-plan memo not found in handoff.")
            return 1

        print("Validation-plan --scaffold smoke check passed.")
        return 0

if __name__ == "__main__":
    raise SystemExit(main())
