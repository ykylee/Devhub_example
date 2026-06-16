#!/usr/bin/env python3
"""
v0.7.27: TASK-V0726-003 (sync_release_hash post-step 자동 호출) smoke test (5/5 PASS)

Test cases:
1. test_apply_auto_calls_sync_hash — version-bump --apply 가 sync_release_hash 자동 호출
2. test_skip_sync_hash_flag — --skip-sync-hash flag 시 sync_hash 호출 안 함
3. test_dry_run_no_sync — --dry-run 시 sync_hash 호출 안 함
4. test_sync_hash_failure_graceful — sync_hash 가 fail 해도 version-bump 는 성공
5. test_post_step_in_result — result dict 에 sync_hash_result 포함
"""
from __future__ import annotations

import json
import os
import subprocess
import sys
import tempfile
from pathlib import Path
from unittest.mock import patch

REPO_ROOT = Path(__file__).resolve().parent.parent.parent
TOOL = REPO_ROOT / "workflow-source" / "tools" / "release_pipeline.py"
SOURCE_ROOT = REPO_ROOT / "workflow-source"
PYTHON = sys.executable

# Ensure 'tools' import path for unittest.mock.patch
if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

# In-process module load (so unittest.mock.patch can target it)
import importlib.util
_spec = importlib.util.spec_from_file_location("release_pipeline", str(TOOL))
_rp = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(_rp)
cmd_version_bump = _rp.cmd_version_bump
_run_post_step_sync_hash = _rp._run_post_step_sync_hash
parse_version = _rp.parse_version


def _run_inproc(args: list[str], repo_root: Path) -> dict:
    """In-process invocation of cmd_version_bump (mock-friendly)."""
    import argparse
    # Strip "version-bump" subcommand positional (release_pipeline argparse 에서 subparser)
    if args and args[0] == "version-bump":
        args = args[1:]
    parser = argparse.ArgumentParser()
    parser.add_argument("--patch", action="store_true")
    parser.add_argument("--minor", action="store_true")
    parser.add_argument("--major", action="store_true")
    parser.add_argument("--to", default=None)
    parser.add_argument("--no-init", action="store_true", dest="no_init")
    parser.add_argument("--dry-run", action="store_true", dest="dry_run")
    parser.add_argument("--apply", dest="apply", action="store_true", default=True)
    parser.add_argument("--skip-sync-hash", action="store_true", dest="skip_sync_hash")
    parser.add_argument("--json", action="store_true")

    # Mock REPO_ROOT (release_pipeline 의 module-level REPO_ROOT)
    import unittest.mock
    with unittest.mock.patch.object(_rp, "REPO_ROOT", repo_root):
        with unittest.mock.patch.object(_rp, "read_version", return_value="0.7.0"):
            with unittest.mock.patch.object(_rp, "read_workflow_kit_version", return_value="v0.7.0-beta"):
                with unittest.mock.patch.object(_rp, "write_version") as mock_wv:
                    with unittest.mock.patch.object(_rp, "write_workflow_kit_version", return_value="v0.7.0-beta") as mock_wk:
                        # argparse parse from args directly (not sys.argv)
                        ns = parser.parse_args(args)
                        # defaults
                        if not ns.dry_run and not ns.apply:
                            ns.apply = True
                        try:
                            result = cmd_version_bump(ns)
                            return result
                        except SystemExit as e:
                            return {"error": f"SystemExit: {e}"}


def _run_tool(*args: str, repo_root: Path) -> dict:
    """Run release_pipeline.py with given args. Returns parsed JSON (subprocess)."""
    cmd = [PYTHON, str(TOOL), *args, "--json"]
    proc = subprocess.run(cmd, capture_output=True, text=True, timeout=30, cwd=str(repo_root))
    if proc.returncode != 0:
        return {"error": proc.stderr, "returncode": proc.returncode, "stdout": proc.stdout}
    try:
        return json.loads(proc.stdout)
    except json.JSONDecodeError as e:
        return {"error": f"JSON parse fail: {e}", "stdout": proc.stdout}


# === Test 1: apply auto-calls sync_hash ===
def test_apply_auto_calls_sync_hash() -> bool:
    """version-bump --apply 가 sync_release_hash 자동 호출 (post-step)."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = Path(tmp) / "fake_repo"
        repo_root.mkdir(parents=True, exist_ok=True)
        # minimal pyproject.toml
        (repo_root / "pyproject.toml").write_text('[project]\nname = "test"\nversion = "0.0.0"\n', encoding="utf-8")
        # workflow_kit dir
        (repo_root / "workflow_kit" / "__init__.py").parent.mkdir(parents=True, exist_ok=True)
        (repo_root / "workflow_kit" / "__init__.py").write_text('__version__ = "v0.0.0-beta"\n', encoding="utf-8")
        # workflow-source dir with sync_release_hash.py mock
        (repo_root / "workflow-source" / "tools").mkdir(parents=True, exist_ok=True)
        sync_tool = repo_root / "workflow-source" / "tools" / "sync_release_hash.py"
        sync_tool.write_text("# mock sync_release_hash.py\nimport sys\nsys.exit(0)\n", encoding="utf-8")

        # git init (REPO_ROOT detect)
        subprocess.run(["git", "init", "-q"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "config", "user.email", "t@t"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "config", "user.name", "T"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "add", "-A"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "commit", "-q", "-m", "init"], cwd=str(repo_root), check=True)

        # release_pipeline.py 의 _run_post_step_sync_hash 를 mock (sync_release_hash 호출 capture)
        with patch.object(_rp, "_run_post_step_sync_hash", return_value={"ok": True, "stdout": "mocked", "stderr": "", "returncode": 0}) as mock_sync:
            result = _run_inproc(["version-bump", "--to=0.7.27", "--apply"], repo_root=repo_root)
            if result.get("error"):
                print(f"  FAIL: tool error: {result['error']}")
                return False
            if mock_sync.call_count == 0:
                print(f"  FAIL: sync_hash not called (count=0)")
                return False
            # call_args[0][0] 가 '0.7.27' 이어야
            called_version = mock_sync.call_args[0][0]
            if called_version != "0.7.27":
                print(f"  FAIL: expected version '0.7.27', got {called_version!r}")
                return False
        # result 에 sync_hash_result 포함
        if "sync_hash_result" not in result:
            print(f"  FAIL: result missing 'sync_hash_result'. Got: {result}")
            return False
        print("  PASS: version-bump --apply auto-calls sync_release_hash (post-step)")
        return True


# === Test 2: --skip-sync-hash flag ===
def test_skip_sync_hash_flag() -> bool:
    """--skip-sync-hash flag 시 sync_hash 호출 안 함."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = Path(tmp) / "fake_repo"
        repo_root.mkdir(parents=True, exist_ok=True)
        (repo_root / "pyproject.toml").write_text('[project]\nname = "test"\nversion = "0.0.0"\n', encoding="utf-8")
        (repo_root / "workflow_kit" / "__init__.py").parent.mkdir(parents=True, exist_ok=True)
        (repo_root / "workflow_kit" / "__init__.py").write_text('__version__ = "v0.0.0-beta"\n', encoding="utf-8")

        subprocess.run(["git", "init", "-q"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "config", "user.email", "t@t"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "config", "user.name", "T"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "add", "-A"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "commit", "-q", "-m", "init"], cwd=str(repo_root), check=True)

        with patch.object(_rp, "_run_post_step_sync_hash", return_value={"ok": True, "stdout": "mocked", "stderr": "", "returncode": 0}) as mock_sync:
            result = _run_inproc(["version-bump", "--to=0.7.27", "--apply", "--skip-sync-hash"], repo_root=repo_root)
            if result.get("error"):
                print(f"  FAIL: tool error: {result['error']}")
                return False
            if mock_sync.call_count != 0:
                print(f"  FAIL: sync_hash called despite --skip-sync-hash (count={mock_sync.call_count})")
                return False
        if "sync_hash_result" in result:
            print(f"  FAIL: result should NOT have 'sync_hash_result' when --skip-sync-hash")
            return False
        print("  PASS: --skip-sync-hash flag bypasses post-step sync")
        return True


# === Test 3: --dry-run no sync ===
def test_dry_run_no_sync() -> bool:
    """--dry-run 시 sync_hash 호출 안 함."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = Path(tmp) / "fake_repo"
        repo_root.mkdir(parents=True, exist_ok=True)
        (repo_root / "pyproject.toml").write_text('[project]\nname = "test"\nversion = "0.7.0"\n', encoding="utf-8")

        subprocess.run(["git", "init", "-q"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "config", "user.email", "t@t"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "config", "user.name", "T"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "add", "-A"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "commit", "-q", "-m", "init"], cwd=str(repo_root), check=True)

        with patch.object(_rp, "_run_post_step_sync_hash", return_value={"ok": True, "stdout": "mocked", "stderr": "", "returncode": 0}) as mock_sync:
            result = _run_inproc(["version-bump", "--patch", "--dry-run"], repo_root=repo_root)
            if result.get("error"):
                print(f"  FAIL: tool error: {result['error']}")
                return False
            if mock_sync.call_count != 0:
                print(f"  FAIL: sync_hash called despite --dry-run (count={mock_sync.call_count})")
                return False
        if result.get("mode") != "dry-run":
            print(f"  FAIL: expected mode='dry-run', got {result.get('mode')!r}")
            return False
        print("  PASS: --dry-run does NOT call sync_hash")
        return True


# === Test 4: sync_hash failure is graceful ===
def test_sync_hash_failure_graceful() -> bool:
    """sync_hash 가 fail (returncode != 0) 해도 version-bump 결과는 정상 반환."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = Path(tmp) / "fake_repo"
        repo_root.mkdir(parents=True, exist_ok=True)
        (repo_root / "pyproject.toml").write_text('[project]\nname = "test"\nversion = "0.0.0"\n', encoding="utf-8")
        (repo_root / "workflow_kit" / "__init__.py").parent.mkdir(parents=True, exist_ok=True)
        (repo_root / "workflow_kit" / "__init__.py").write_text('__version__ = "v0.0.0-beta"\n', encoding="utf-8")

        subprocess.run(["git", "init", "-q"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "config", "user.email", "t@t"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "config", "user.name", "T"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "add", "-A"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "commit", "-q", "-m", "init"], cwd=str(repo_root), check=True)

        with patch.object(_rp, "_run_post_step_sync_hash", return_value={"ok": False, "stdout": "", "stderr": "mock failure", "returncode": 1}) as mock_sync:
            result = _run_inproc(["version-bump", "--to=0.7.27", "--apply"], repo_root=repo_root)
            if result.get("error"):
                print(f"  FAIL: tool error: {result['error']}")
                return False
            # version-bump 자체는 성공 (pyproject.toml + workflow_kit 변경됨)
            if result.get("mode") != "applied":
                print(f"  FAIL: expected mode='applied', got {result.get('mode')!r}")
                return False
            if result.get("current_pyproject") != "0.7.27":
                print(f"  FAIL: pyproject not bumped: {result.get('current_pyproject')!r}")
                return False
            # sync_hash_result.ok = False 포함
            sync_result = result.get("sync_hash_result", {})
            if sync_result.get("ok") is not False:
                print(f"  FAIL: sync_hash_result.ok should be False. Got: {sync_result}")
                return False
        print("  PASS: sync_hash failure is graceful (version-bump succeeds)")
        return True


# === Test 5: post-step in result dict ===
def test_post_step_in_result() -> bool:
    """apply 시 result dict 에 sync_hash_result field 가 *있어야* 함 (caller 가 확인 가능)."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = Path(tmp) / "fake_repo"
        repo_root.mkdir(parents=True, exist_ok=True)
        (repo_root / "pyproject.toml").write_text('[project]\nname = "test"\nversion = "0.0.0"\n', encoding="utf-8")
        (repo_root / "workflow_kit" / "__init__.py").parent.mkdir(parents=True, exist_ok=True)
        (repo_root / "workflow_kit" / "__init__.py").write_text('__version__ = "v0.0.0-beta"\n', encoding="utf-8")

        subprocess.run(["git", "init", "-q"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "config", "user.email", "t@t"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "config", "user.name", "T"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "add", "-A"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "commit", "-q", "-m", "init"], cwd=str(repo_root), check=True)

        with patch.object(_rp, "_run_post_step_sync_hash", return_value={"ok": True, "stdout": "synced", "stderr": "", "returncode": 0}) as mock_sync:
            result = _run_inproc(["version-bump", "--to=0.7.27", "--apply"], repo_root=repo_root)
            if result.get("error"):
                print(f"  FAIL: tool error: {result['error']}")
                return False
            sync_result = result.get("sync_hash_result", {})
            required_keys = {"ok", "stdout", "stderr", "returncode"}
            if not required_keys.issubset(sync_result.keys()):
                missing = required_keys - sync_result.keys()
                print(f"  FAIL: sync_hash_result missing keys: {missing}. Got: {sync_result}")
                return False
        print("  PASS: sync_hash_result in result dict with required keys")
        return True


# === Main ===
def main() -> int:
    print("=" * 60)
    print("TASK-V0726-003 (sync_release_hash post-step) smoke test (v0.7.27)")
    print("=" * 60)

    tests = [
        test_apply_auto_calls_sync_hash,
        test_skip_sync_hash_flag,
        test_dry_run_no_sync,
        test_sync_hash_failure_graceful,
        test_post_step_in_result,
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
