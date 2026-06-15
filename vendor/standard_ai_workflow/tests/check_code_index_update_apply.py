#!/usr/bin/env python3
"""Smoke test for the code-index-update skill's --apply mode."""

from __future__ import annotations

import json
import os
import subprocess
import sys
import tempfile
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
SCRIPT_PATH = SOURCE_ROOT / "skills" / "code-index-update" / "scripts" / "run_code_index_update.py"

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
            "# Project Profile\n\n- 프로젝트명: Test Project\n- 문서 위키 홈: README.md\n- 운영 문서 위치: docs/ops.md\n",
            encoding="utf-8"
        )

        # README.md 생성 (추천 후보로 등장하게 함)
        readme_path = project_root / "README.md"
        readme_path.write_text("# Root README", encoding="utf-8")

        handoff_path = project_root / "session_handoff.md"
        handoff_path.write_text(
            "# 세션 인계 문서\n\n## 1. 현재 작업 요약\n- 최종 수정일: 2026-01-01\n\n## 현재 세션 운영 메모\n- 기존 메모\n\n## 다음에 읽을 문서\n- [Old Doc](./old.md)\n",
            encoding="utf-8"
        )

        work_backlog_index_path = project_root / "work_backlog.md"
        work_backlog_index_path.write_text("# Work Backlog Index", encoding="utf-8")

        # 2. code-index-update --apply 실행
        env = os.environ.copy()
        env["CODEX_WORKFLOW_BRANCH"] = "test/apply"
        completed = subprocess.run(
            [
                sys.executable,
                str(SCRIPT_PATH),
                "--project-profile-path", str(profile_path),
                "--session-handoff-path", str(handoff_path),
                "--work-backlog-index-path", str(work_backlog_index_path),
                "--changed-file", "docs/operations/runbooks/new-guide.md",
                "--change-summary", "신규 가이드 추가",
                "--apply"
            ],
            capture_output=True,
            text=True,
            check=False,
            env=env,
        )

        if completed.returncode != 0:
            print(f"code-index-update failed with return code {completed.returncode}")
            print(f"stdout: {completed.stdout}")
            print(f"stderr: {completed.stderr}")
            return 1

        payload = json.loads(completed.stdout)

        # 3. 결과 검증
        if payload.get("apply_status") != "applied":
            print(f"Expected apply_status 'applied', got {payload.get('apply_status')}")
            return 1

        # handoff 갱신 검증
        handoff_content = handoff_path.read_text(encoding="utf-8")
        if "[README.md]" not in handoff_content:
            print("README.md link not found in updated handoff.")
            return 1
        if "[code-index-update]" not in handoff_content:
            print("code-index-update memo not found in handoff.")
            return 1

        # state.json 생성 검증
        state_path = project_root / "test" / "apply" / "state.json"
        if not state_path.exists():
            print("state.json not created after code-index-update --apply.")
            return 1

        print("Code-index-update --apply smoke check passed.")
        return 0

if __name__ == "__main__":
    raise SystemExit(main())
