#!/usr/bin/env python3
"""Smoke test for the doc-sync skill's --apply mode."""

from __future__ import annotations

import json
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
SCRIPT_PATH = SOURCE_ROOT / "skills" / "doc-sync" / "scripts" / "run_doc_sync.py"

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

        handoff_path = project_root / "session_handoff.md"
        handoff_path.write_text(
            "# 세션 인계 문서\n\n## 1. 현재 작업 요약\n- 최종 수정일: 2026-01-01\n\n## 현재 세션 운영 메모\n- 기존 메모\n\n## 다음에 읽을 문서\n- [Old Doc](./old.md)\n",
            encoding="utf-8"
        )

        # 영향받는 문서로 추천될 파일 생성
        ops_path = project_root / "docs" / "ops.md"
        ops_path.parent.mkdir(parents=True)
        ops_path.write_text("# Ops Runbook", encoding="utf-8")

        # 2. doc-sync --apply 실행
        completed = subprocess.run(
            [
                sys.executable,
                str(SCRIPT_PATH),
                "--project-profile-path", str(profile_path),
                "--session-handoff-path", str(handoff_path),
                "--changed-file", "app/main.py",
                "--change-summary", "설정 변경 및 운영 가이드 확인 필요",
                "--apply"
            ],
            capture_output=True,
            text=True,
            check=False
        )

        if completed.returncode != 0:
            print(f"doc-sync failed with return code {completed.returncode}")
            print(f"stdout: {completed.stdout}")
            print(f"stderr: {completed.stderr}")
            return 1

        payload = json.loads(completed.stdout)

        # 3. 결과 검증
        if payload.get("apply_status") != "applied":
            print(f"Expected apply_status 'applied', got {payload.get('apply_status')}")
            return 1

        if str(handoff_path.resolve()) not in [str(Path(p).resolve()) for p in payload.get("written_paths", [])]:
            print(f"handoff_path not in written_paths: {payload.get('written_paths')}")
            return 1

        # 파일 내용 검증
        updated_content = handoff_path.read_text(encoding="utf-8")

        # '다음에 읽을 문서' 섹션이 갱신되었는지 확인 (ops.md 링크 포함 여부)
        if "[ops.md](docs/ops.md)" not in updated_content:
            print("Link to ops.md not found in updated handoff.")
            print(f"Content:\n{updated_content}")
            return 1

        # 기존 링크는 제거되었는지 확인 (교체 모드이므로)
        if "[Old Doc]" in updated_content:
            print("Old link still present in updated handoff (should be replaced).")
            return 1

        # '현재 세션 운영 메모'에 follow_up_actions가 추가되었는지 확인
        if "[doc-sync]" not in updated_content:
            print("doc-sync follow-up action not found in updated handoff.")
            return 1

        if "기존 메모" not in updated_content:
            print("Existing memo lost in updated handoff.")
            return 1

        print("Doc-sync --apply smoke check passed.")
        return 0

if __name__ == "__main__":
    raise SystemExit(main())
