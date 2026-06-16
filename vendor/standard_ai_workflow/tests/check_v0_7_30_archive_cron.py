#!/usr/bin/env python3
"""
v0.7.30: TASK-V0728-001 (archive_stale_memory --install-cron) smoke test (5/5 PASS)

Test cases:
1. test_install_cron_subprocess — --install-cron 이 'mavis cron create' 호출
2. test_uninstall_cron_subprocess — --uninstall-cron 이 'mavis cron disable' 호출
3. test_show_cron_subprocess — --show-cron 이 'mavis cron list' 호출
4. test_cron_name_flag — --cron-name=<custom> 적용
5. test_cron_interval_flag — --cron-interval=<custom> 적용
"""
from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path
from unittest.mock import MagicMock, patch

REPO_ROOT = Path(__file__).resolve().parent.parent.parent
TOOL = REPO_ROOT / "workflow-source" / "tools" / "archive_stale_memory.py"
SOURCE_ROOT = REPO_ROOT / "workflow-source"
PYTHON = sys.executable

if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

# In-process module load
import importlib.util
_spec = importlib.util.spec_from_file_location("archive_stale_memory", str(TOOL))
_am = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(_am)
cmd_install_cron = _am.cmd_install_cron
cmd_uninstall_cron = _am.cmd_uninstall_cron
cmd_show_cron = _am.cmd_show_cron


def _make_fake_proc(returncode: int = 0, stdout: str = "", stderr: str = ""):
    proc = MagicMock()
    proc.returncode = returncode
    proc.stdout = stdout
    proc.stderr = stderr
    return proc


def _make_args(**kwargs):
    """Build argparse.Namespace with cron flags + memory dir flags."""
    import argparse
    defaults = {
        "older_than": 30,
        "list": False,
        "dry_run": False,
        "apply": False,
        "cleanup": False,
        "repo_root": str(REPO_ROOT),
        "install_cron": False,
        "uninstall_cron": False,
        "show_cron": False,
        "cron_name": "archive-memory",
        "cron_interval": "7d",
        "agent": "mavis",
    }
    defaults.update(kwargs)
    return argparse.Namespace(**defaults)


# === Test 1: install_cron subprocess ===
def test_install_cron_subprocess() -> bool:
    """--install-cron 이 'mavis cron create' 호출 (correct args).

    v0.7.31 의 TASK-V0730-001 fix: install_cron 가 *idempotency* check (info) 추가 → 2 calls
    (info + create). info fail → create 만. → 2 calls expected.
    """
    args = _make_args(install_cron=True)
    with patch("subprocess.run") as mock_run:
        # v0.7.31: info fail (returncode 1) → create OK (returncode 0)
        mock_run.side_effect = [
            _make_fake_proc(1, "", "Error: cron not found\n"),  # info fail
            _make_fake_proc(0, "created\n"),  # create OK
        ]
        result = cmd_install_cron(args)
        if not result.get("ok"):
            print(f"  FAIL: ok=False. Got: {result}")
            return False
        if mock_run.call_count != 2:
            print(f"  FAIL: expected 2 subprocess calls (info + create), got {mock_run.call_count}")
            return False
        # verify 2nd call (create)
        call_args = mock_run.call_args_list[1][0][0]
        # verify 'mavis cron create' + agent + cronName + --schedule
        if "mavis" not in call_args or "cron" not in call_args or "create" not in call_args:
            print(f"  FAIL: expected 'mavis cron create', got {call_args}")
            return False
        if "mavis" not in call_args:
            print(f"  FAIL: agent 'mavis' not in args: {call_args}")
            return False
        if "archive-memory" not in call_args:
            print(f"  FAIL: cron name 'archive-memory' not in args: {call_args}")
            return False
        if not any("--schedule=7d" in str(a) for a in call_args):
            print(f"  FAIL: --schedule=7d not in args: {call_args}")
            return False
    print("  PASS: --install-cron → 'mavis cron create' with correct args")
    return True


# === Test 2: uninstall_cron subprocess ===
def test_uninstall_cron_subprocess() -> bool:
    """--uninstall-cron 이 'mavis cron disable' 호출."""
    args = _make_args(uninstall_cron=True)
    with patch("subprocess.run") as mock_run:
        mock_run.return_value = _make_fake_proc(0, "disabled\n")
        result = cmd_uninstall_cron(args)
        if not result.get("ok"):
            print(f"  FAIL: ok=False. Got: {result}")
            return False
        if mock_run.call_count != 1:
            print(f"  FAIL: expected 1 subprocess call, got {mock_run.call_count}")
            return False
        call_args = mock_run.call_args[0][0]
        if "cron" not in call_args or "disable" not in call_args:
            print(f"  FAIL: expected 'mavis cron disable', got {call_args}")
            return False
        if "mavis" not in call_args:
            print(f"  FAIL: agent 'mavis' not in args: {call_args}")
            return False
        if "archive-memory" not in call_args:
            print(f"  FAIL: cron name not in args: {call_args}")
            return False
    print("  PASS: --uninstall-cron → 'mavis cron disable' with correct args")
    return True


# === Test 3: show_cron subprocess ===
def test_show_cron_subprocess() -> bool:
    """--show-cron 이 'mavis cron list' 호출 + found=True/False 정확."""
    args = _make_args(show_cron=True)
    with patch("subprocess.run") as mock_run:
        # mavis cron list 의 output 에 cronName 'archive-memory' 포함
        mock_run.return_value = _make_fake_proc(0, '{"tasks": [{"cronName": "archive-memory"}]}\n')
        result = cmd_show_cron(args)
        if not result.get("ok"):
            print(f"  FAIL: ok=False. Got: {result}")
            return False
        if not result.get("found"):
            print(f"  FAIL: expected found=True, got {result.get('found')}")
            return False
        call_args = mock_run.call_args[0][0]
        if "cron" not in call_args or "list" not in call_args:
            print(f"  FAIL: expected 'mavis cron list', got {call_args}")
            return False
        if "mavis" not in call_args:
            print(f"  FAIL: agent 'mavis' not in args: {call_args}")
            return False
    # found=False case
    with patch("subprocess.run") as mock_run:
        mock_run.return_value = _make_fake_proc(0, '{"tasks": [{"cronName": "other-cron"}]}\n')
        result = cmd_show_cron(args)
        if result.get("found"):
            print(f"  FAIL: expected found=False, got {result.get('found')}")
            return False
    print("  PASS: --show-cron → 'mavis cron list' + found=True/False correct")
    return True


# === Test 4: --cron-name flag ===
def test_cron_name_flag() -> bool:
    """--cron-name=<custom> 적용."""
    args = _make_args(install_cron=True, cron_name="my-archive-job")
    with patch("subprocess.run") as mock_run:
        mock_run.return_value = _make_fake_proc(0, "created\n")
        result = cmd_install_cron(args)
        if not result.get("ok"):
            print(f"  FAIL: ok=False. Got: {result}")
            return False
        if result.get("cron_name") != "my-archive-job":
            print(f"  FAIL: cron_name != 'my-archive-job'. Got: {result.get('cron_name')}")
            return False
        call_args = mock_run.call_args[0][0]
        if "my-archive-job" not in call_args:
            print(f"  FAIL: custom cron name not in args: {call_args}")
            return False
    print("  PASS: --cron-name=<custom> applied to subprocess args")
    return True


# === Test 5: --cron-interval flag ===
def test_cron_interval_flag() -> bool:
    """--cron-interval=<custom> 적용."""
    args = _make_args(install_cron=True, cron_interval="1d")
    with patch("subprocess.run") as mock_run:
        mock_run.return_value = _make_fake_proc(0, "created\n")
        result = cmd_install_cron(args)
        if not result.get("ok"):
            print(f"  FAIL: ok=False. Got: {result}")
            return False
        if result.get("cron_interval") != "1d":
            print(f"  FAIL: cron_interval != '1d'. Got: {result.get('cron_interval')}")
            return False
        call_args = mock_run.call_args[0][0]
        if not any("--schedule=1d" in str(a) for a in call_args):
            print(f"  FAIL: --schedule=1d not in args: {call_args}")
            return False
    print("  PASS: --cron-interval=<custom> applied to subprocess args")
    return True


# === Main ===
def main() -> int:
    print("=" * 60)
    print("TASK-V0728-001 (archive_stale_memory --install-cron) smoke test (v0.7.30)")
    print("=" * 60)

    tests = [
        test_install_cron_subprocess,
        test_uninstall_cron_subprocess,
        test_show_cron_subprocess,
        test_cron_name_flag,
        test_cron_interval_flag,
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
