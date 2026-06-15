#!/usr/bin/env python3
import subprocess
import json
import os
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]


def test_git_history_summarizer():
    print("Testing git-history-summarizer MCP...")

    env = os.environ.copy()
    existing_pythonpath = env.get("PYTHONPATH", "")
    env["PYTHONPATH"] = (
        f"{REPO_ROOT / 'workflow-source'}"
        + (os.pathsep + existing_pythonpath if existing_pythonpath else "")
    )

    # Test Markdown format
    cmd_md = [sys.executable, str(REPO_ROOT / "workflow-source/mcp_servers/git-history-summarizer/scripts/run_git_history_summarizer.py"), "--range", "HEAD~1..HEAD"]
    result_md = subprocess.run(cmd_md, capture_output=True, text=True, check=True, cwd=str(REPO_ROOT), env=env)
    assert "Git 작업 요약" in result_md.stdout
    print("  [OK] Markdown output format")

    # Test JSON format
    cmd_json = [sys.executable, str(REPO_ROOT / "workflow-source/mcp_servers/git-history-summarizer/scripts/run_git_history_summarizer.py"), "--range", "HEAD~1..HEAD", "--json"]
    result_json = subprocess.run(cmd_json, capture_output=True, text=True, check=True, cwd=str(REPO_ROOT), env=env)
    data = json.loads(result_json.stdout)
    assert data["status"] == "ok"
    assert "entries" in data
    assert len(data["entries"]) >= 1
    print("  [OK] JSON output format")

    print("All tests passed for git-history-summarizer!")

if __name__ == "__main__":
    try:
        test_git_history_summarizer()
    except Exception as e:
        print(f"Test failed: {e}")
        exit(1)
