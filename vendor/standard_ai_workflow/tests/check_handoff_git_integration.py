#!/usr/bin/env python3
import subprocess
import json
from pathlib import Path

def test_handoff_git_integration():
    print("Testing handoff-git integration...")

    repo_root = Path(__file__).resolve().parents[2]
    latest_backlog = repo_root / "workflow-source" / "examples" / "acme_delivery_platform" / "backlog" / "2026-04-18.md"

    cmd = [
        "python3", str(repo_root / "workflow-source" / "mcp_servers" / "create-session-handoff-draft" / "scripts" / "run_create_session_handoff_draft.py"),
        "--latest-backlog-path", str(latest_backlog),
        "--git-range", "HEAD~3..HEAD"
    ]

    result = subprocess.run(cmd, capture_output=True, text=True, check=True, cwd=repo_root)
    payload = json.loads(result.stdout)

    assert payload["status"] == "ok"
    assert "draft_handoff" in payload

    draft_text = "\n".join(payload["draft_handoff"])
    assert "## Git Summary" in draft_text
    assert "Git 작업 요약 (HEAD~3..HEAD)" in draft_text

    print("  [OK] Handoff draft contains git summary")
    print("All tests passed for handoff-git integration!")

if __name__ == "__main__":
    try:
        test_handoff_git_integration()
    except Exception as e:
        print(f"Test failed: {e}")
        exit(1)
