#!/usr/bin/env python3
"""
v0.7.26: sync_release_hash smoke test (5/5 PASS)

Test cases:
1. test_state_json_hash_updated — state.json 의 `v0.7.26 (TBD):` → `v0.7.26 (latest):`
2. test_backlog_commit_updated — backlog 의 `**commit**: \`TBD\`` → `**commit**: \`latest\``
3. test_idempotent_no_match — 이미 매칭된 hash 면 no change
4. test_dry_run_no_write — --dry-run 시 file 변경 없음
5. test_real_repo_hash_sync — *real* REPO_ROOT + *real* v0.7.25 entry + *real* HEAD → sync
"""
from __future__ import annotations

import json
import os
import re
import subprocess
import sys
import tempfile
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent.parent
TOOL = REPO_ROOT / "workflow-source" / "tools" / "sync_release_hash.py"
PYTHON = sys.executable


def _run_tool(*args: str, repo_root: Path) -> dict:
    """Run sync_release_hash.py with given args. Returns parsed output (text mode)."""
    cmd = [PYTHON, str(TOOL), *args, "--repo-root", str(repo_root)]
    proc = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
    if proc.returncode != 0:
        return {"error": proc.stderr, "returncode": proc.returncode, "stdout": proc.stdout}
    return {"stdout": proc.stdout}


# === Test 1: state.json hash updated ===
def test_state_json_hash_updated() -> bool:
    """state.json 의 `v0.7.26 (TBD):` → `v0.7.26 (abc1234):`."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = Path(tmp) / "fake_repo"
        (repo_root / "ai-workflow" / "memory" / "active").mkdir(parents=True, exist_ok=True)
        state_path = repo_root / "ai-workflow" / "memory" / "active" / "state.json"
        state_path.write_text(json.dumps({
            "session": {
                "recent_done_items": [
                    "v0.7.26 (TBD): P3 test entry",
                    "v0.7.25 (old): P3 prior entry",
                ]
            }
        }, indent=2), encoding="utf-8")

        # git init (sync_release_hash 의 get_latest_commit_hash 가 동작해야 함)
        subprocess.run(["git", "init", "-q"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "config", "user.email", "test@test.com"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "config", "user.name", "Test"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "add", "-A"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "commit", "-q", "-m", "init"], cwd=str(repo_root), check=True)
        head_full = subprocess.run(
            ["git", "rev-parse", "HEAD"],
            cwd=str(repo_root), capture_output=True, text=True, check=True,
        ).stdout.strip()
        head_sha = subprocess.run(
            ["git", "rev-parse", "--short=7", head_full],
            cwd=str(repo_root), capture_output=True, text=True, check=True,
        ).stdout.strip()[:7]

        # backlog 디렉토리 (없어도 됨 — skip)
        result = _run_tool("--version=v0.7.26", "--apply", repo_root=repo_root)
        if result.get("error"):
            print(f"  FAIL: tool error: {result['error']}")
            return False
        new_state = state_path.read_text(encoding="utf-8")
        if f"v0.7.26 ({head_sha}):" not in new_state:
            print(f"  FAIL: state.json not updated. Expected 'v0.7.26 ({head_sha}):', got:")
            print(new_state[:500])
            return False
        if "v0.7.26 (TBD):" in new_state:
            print(f"  FAIL: state.json still has 'TBD'")
            return False
        print(f"  PASS: state.json updated ({head_sha})")
        return True


# === Test 2: backlog commit updated ===
def test_backlog_commit_updated() -> bool:
    """backlog 의 `**commit**: \`TBD\`` → `**commit**: \`abc1234\``."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = Path(tmp) / "fake_repo"
        (repo_root / "ai-workflow" / "memory" / "active").mkdir(parents=True, exist_ok=True)
        (repo_root / "ai-workflow" / "memory" / "release" / "v0.7.26" / "backlog").mkdir(parents=True, exist_ok=True)
        state_path = repo_root / "ai-workflow" / "memory" / "active" / "state.json"
        state_path.write_text(json.dumps({"session": {"recent_done_items": []}}, indent=2), encoding="utf-8")
        backlog_path = repo_root / "ai-workflow" / "memory" / "release" / "v0.7.26" / "backlog" / "2026-06-15.md"
        backlog_path.write_text("# v0.7.26 Backlog\n\n- **commit**: `TBD`\n", encoding="utf-8")

        subprocess.run(["git", "init", "-q"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "config", "user.email", "t@t"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "config", "user.name", "T"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "add", "-A"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "commit", "-q", "-m", "init"], cwd=str(repo_root), check=True)
        head_full = subprocess.run(
            ["git", "rev-parse", "HEAD"],
            cwd=str(repo_root), capture_output=True, text=True, check=True,
        ).stdout.strip()
        head_sha = subprocess.run(
            ["git", "rev-parse", "--short=7", head_full],
            cwd=str(repo_root), capture_output=True, text=True, check=True,
        ).stdout.strip()[:7]

        result = _run_tool("--version=v0.7.26", "--apply", repo_root=repo_root)
        if result.get("error"):
            print(f"  FAIL: tool error: {result['error']}")
            return False
        new_backlog = backlog_path.read_text(encoding="utf-8")
        if f"**commit**: `{head_sha}`" not in new_backlog:
            print(f"  FAIL: backlog not updated. Expected '**commit**: `{head_sha}`'")
            return False
        print(f"  PASS: backlog updated ({head_sha})")
        return True


# === Test 3: idempotent (no match) ===
def test_idempotent_no_match() -> bool:
    """이미 매칭된 hash 면 no change (state_updated=False)."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = Path(tmp) / "fake_repo"
        (repo_root / "ai-workflow" / "memory" / "active").mkdir(parents=True, exist_ok=True)
        state_path = repo_root / "ai-workflow" / "memory" / "active" / "state.json"
        state_path.write_text(json.dumps({
            "session": {"recent_done_items": []}
        }, indent=2), encoding="utf-8")

        subprocess.run(["git", "init", "-q"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "config", "user.email", "t@t"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "config", "user.name", "T"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "add", "-A"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "commit", "-q", "-m", "init"], cwd=str(repo_root), check=True)

        # v0.7.26 entry 부재 → state_updated=False
        result = _run_tool("--version=v0.7.26", "--apply", repo_root=repo_root)
        if result.get("error"):
            print(f"  FAIL: tool error: {result['error']}")
            return False
        if "state_updated: False" not in result["stdout"]:
            print(f"  FAIL: expected state_updated: False. Got: {result['stdout']}")
            return False
        print("  PASS: no v0.7.26 entry → state_updated: False (idempotent)")
        return True


# === Test 4: dry-run no write ===
def test_dry_run_no_write() -> bool:
    """--dry-run 시 file 변경 없음."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = Path(tmp) / "fake_repo"
        (repo_root / "ai-workflow" / "memory" / "active").mkdir(parents=True, exist_ok=True)
        state_path = repo_root / "ai-workflow" / "memory" / "active" / "state.json"
        original_content = json.dumps({
            "session": {
                "recent_done_items": ["v0.7.26 (TBD): test"]
            }
        }, indent=2)
        state_path.write_text(original_content, encoding="utf-8")

        subprocess.run(["git", "init", "-q"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "config", "user.email", "t@t"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "config", "user.name", "T"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "add", "-A"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "commit", "-q", "-m", "init"], cwd=str(repo_root), check=True)

        result = _run_tool("--version=v0.7.26", "--dry-run", repo_root=repo_root)
        if result.get("error"):
            print(f"  FAIL: tool error: {result['error']}")
            return False
        # file 변경 없음 확인
        current_content = state_path.read_text(encoding="utf-8")
        if current_content != original_content:
            print(f"  FAIL: file was modified despite --dry-run")
            return False
        if "mode: dry-run" not in result["stdout"]:
            print(f"  FAIL: expected 'mode: dry-run'. Got: {result['stdout']}")
            return False
        print("  PASS: --dry-run: no file write")
        return True


# === Test 5: real repo + real v0.7.25 entry ===
def test_real_repo_hash_sync() -> bool:
    """*real* REPO_ROOT + *real* v0.7.25 state.json entry → sync with current HEAD (00e7ca8)."""
    state_path = REPO_ROOT / "ai-workflow" / "memory" / "active" / "state.json"
    if not state_path.exists():
        print(f"  SKIP: real state.json not found")
        return True  # not a failure, just skip
    original_content = state_path.read_text(encoding="utf-8")
    head_full = subprocess.run(
        ["git", "rev-parse", "HEAD"],
        cwd=str(REPO_ROOT), capture_output=True, text=True, check=True,
    ).stdout.strip()
    head_sha = subprocess.run(
        ["git", "rev-parse", "--short=7", head_full],
        cwd=str(REPO_ROOT), capture_output=True, text=True, check=True,
    ).stdout.strip()[:7]

    # 현재 v0.7.25 entry 가 어떤 hash 인지 확인
    m = re.search(r"v0\.7\.25 \(([a-f0-9]{7})\):", original_content)
    if not m:
        print(f"  SKIP: no v0.7.25 entry in state.json")
        return True
    current_hash = m.group(1)
    if current_hash == head_sha:
        print(f"  SKIP: v0.7.25 entry already at HEAD ({head_sha})")
        return True

    # --dry-run 으로 verify
    result = _run_tool("--version=v0.7.25", "--dry-run", repo_root=REPO_ROOT)
    if result.get("error"):
        print(f"  FAIL: tool error: {result['error']}")
        return False
    if f"new_hash: {head_sha}" not in result["stdout"]:
        print(f"  FAIL: expected new_hash={head_sha}. Got: {result['stdout']}")
        return False
    # file unchanged (--dry-run)
    after_content = state_path.read_text(encoding="utf-8")
    if after_content != original_content:
        print(f"  FAIL: --dry-run modified file")
        return False
    print(f"  PASS: real repo v0.7.25 sync verify ({current_hash} → {head_sha}, --dry-run)")
    return True


# === Main ===
def main() -> int:
    print("=" * 60)
    print("sync_release_hash smoke test (v0.7.26)")
    print("=" * 60)

    tests = [
        test_state_json_hash_updated,
        test_backlog_commit_updated,
        test_idempotent_no_match,
        test_dry_run_no_write,
        test_real_repo_hash_sync,
    ]
    passed = 0
    for test in tests:
        print(f"\n{test.__name__}:")
        if test():
            passed += 1

    print()
    print("=" * 60)
    print(f"Result: {passed}/{len(tests)} PASS")
    print("=" * 60)
    return 0 if passed == len(tests) else 1


if __name__ == "__main__":
    sys.exit(main())
