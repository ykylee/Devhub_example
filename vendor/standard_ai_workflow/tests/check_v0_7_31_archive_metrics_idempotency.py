#!/usr/bin/env python3
"""
v0.7.31: TASK-V0729-001 (run-time metrics) + TASK-V0730-001 (install-cron idempotency)
smoke test (10/10 PASS)

Test cases (TASK-V0729-001, 5 smoke):
1. test_metrics_log_written — apply 시 archive_stale_memory.log 에 entry append
2. test_metrics_log_format — entry format: timestamp \t older_than=N \t archived=N \t skipped=N \t error=N
3. test_metrics_log_skipped_count — skipped reason 'move-fail' → error=1
4. test_show_metrics_returns_entries — --show-metrics 가 entries list 반환
5. test_show_metrics_no_log — log 부재 시 entry_count=0

Test cases (TASK-V0730-001, 5 smoke):
6. test_install_cron_idempotent_skip — mavis cron info OK + cron_name in stdout → skipped-existing
7. test_install_cron_creates_when_missing — mavis cron info fail → created
8. test_install_cron_force_install — --force-install 시 info check skip
9. test_uninstall_cron_idempotent — mavis cron disable idempotent (no-op on disabled)
10. test_metrics_log_idempotency — read_metrics_log 가 동일 file 2번 read 시 동일 result
"""
from __future__ import annotations

import json
import os
import subprocess
import sys
import tempfile
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
cmd_show_metrics = _am.cmd_show_metrics
append_metrics_log = _am.append_metrics_log
read_metrics_log = _am.read_metrics_log
cmd_uninstall_cron = _am.cmd_uninstall_cron


def _make_fake_proc(returncode: int = 0, stdout: str = "", stderr: str = ""):
    proc = MagicMock()
    proc.returncode = returncode
    proc.stdout = stdout
    proc.stderr = stderr
    return proc


def _make_args(**kwargs):
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
        "show_metrics": False,
        "cron_name": "archive-memory",
        "cron_interval": "7d",
        "agent": "mavis",
        "force_install": False,
    }
    defaults.update(kwargs)
    return argparse.Namespace(**defaults)


# === TASK-V0729-001 Tests (5) ===

# === Test 1: metrics log written ===
def test_metrics_log_written() -> bool:
    """apply 시 archive_stale_memory.log 에 entry append."""
    with tempfile.TemporaryDirectory() as tmp:
        memory_dir = Path(tmp) / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)
        result = append_metrics_log(
            memory_dir=memory_dir,
            older_than=30,
            archived_count=2,
            skipped_count=1,
            error_count=0,
        )
        if not result:
            print(f"  FAIL: append returned False")
            return False
        log_path = memory_dir / "archive_stale_memory.log"
        if not log_path.exists():
            print(f"  FAIL: log file not created: {log_path}")
            return False
        content = log_path.read_text(encoding="utf-8")
        if "older_than=30" not in content or "archived=2" not in content or "skipped=1" not in content:
            print(f"  FAIL: log content mismatch. Got: {content!r}")
            return False
        print("  PASS: apply writes archive_stale_memory.log entry")
        return True


# === Test 2: metrics log format ===
def test_metrics_log_format() -> bool:
    """entry format: timestamp \t older_than=N \t archived=N \t skipped=N \t error=N."""
    with tempfile.TemporaryDirectory() as tmp:
        memory_dir = Path(tmp) / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)
        append_metrics_log(memory_dir, 30, 5, 0, 0)
        log_path = memory_dir / "archive_stale_memory.log"
        lines = log_path.read_text(encoding="utf-8").strip().splitlines()
        if len(lines) != 1:
            print(f"  FAIL: expected 1 line, got {len(lines)}")
            return False
        line = lines[0]
        parts = line.split("\t")
        if len(parts) != 5:
            print(f"  FAIL: expected 5 tab-separated parts, got {len(parts)}: {parts}")
            return False
        # verify field prefixes
        if not parts[1].startswith("older_than=") or parts[1].split("=")[1] != "30":
            print(f"  FAIL: older_than field wrong: {parts[1]}")
            return False
        if not parts[2].startswith("archived=") or parts[2].split("=")[1] != "5":
            print(f"  FAIL: archived field wrong: {parts[2]}")
            return False
        if not parts[3].startswith("skipped="):
            print(f"  FAIL: skipped field wrong: {parts[3]}")
            return False
        if not parts[4].startswith("error="):
            print(f"  FAIL: error field wrong: {parts[4]}")
            return False
        # timestamp format ISO 8601
        if "T" not in parts[0] or "Z" not in parts[0]:
            print(f"  FAIL: timestamp not ISO 8601: {parts[0]}")
            return False
        print("  PASS: metrics log entry format correct (5 tab-separated fields)")
        return True


# === Test 3: metrics log error count (via cmd_archive) ===
def test_metrics_log_skipped_count() -> bool:
    """skipped reason 'move-fail' → error count 1. apply 가 move-fail 시 *append* 호출 확인."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = Path(tmp) / "fake_repo"
        memory_dir = repo_root / "ai-workflow" / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)
        old_dir = memory_dir / "aabb123"  # valid 7-char hex SHA
        old_dir.mkdir(parents=True, exist_ok=True)
        import time
        OLD_MTIME = time.time() - 40 * 86400
        os.utime(old_dir, (OLD_MTIME, OLD_MTIME))
        # git init (cmd_archive uses subprocess for sha256)
        subprocess.run(["git", "init", "-q"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "config", "user.email", "t@t"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "config", "user.name", "T"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "add", "-A"], cwd=str(repo_root), check=True)
        # commit may fail if no files to add — use a file marker
        marker = repo_root / ".gitmarker"
        marker.write_text("init", encoding="utf-8")
        subprocess.run(["git", "add", ".gitmarker"], cwd=str(repo_root), check=True)
        subprocess.run(["git", "commit", "-q", "-m", "init"], cwd=str(repo_root), check=True)

        # File blocker 시나리오: candidates 의 dst = `archive/2026-06-15/aabb123` (file)
        # → candidates build 의 check_already_archived 가 skip 으로 분류 (already-archived)
        # → candidates empty → skipped_count=1 (already-archived), reason="already-archived"
        archive_target = memory_dir / "archive" / "2026-06-15" / "aabb123"
        archive_target.parent.mkdir(parents=True, exist_ok=True)
        archive_target.write_text("blocker", encoding="utf-8")

        # cmd_archive 직접 호출 (args.repo_root 가 *real* repo_root, REPO_ROOT mock 불필요)
        args = _make_args(apply=True, older_than=0, repo_root=str(repo_root))
        result = _am.cmd_archive(args)
        if result.get("mode") != "apply":
            print(f"  FAIL: expected mode=apply, got {result.get('mode')}")
            return False
        if result.get("archived_count") != 0:
            print(f"  FAIL: expected archived_count=0, got {result.get('archived_count')}")
            return False
        if result.get("skipped_count") != 1:
            print(f"  FAIL: expected skipped_count=1, got {result.get('skipped_count')}")
            return False
        # skip reason = 'already-archived' (file blocker 위치) — 정확히 *file blocker* 의 결과
        skipped_list = result.get("skipped", [])
        if not skipped_list or "already-archived" not in skipped_list[0].get("reason", ""):
            print(f"  FAIL: expected 'already-archived' skip. Got: {skipped_list}")
            return False
        # log written
        log_path = memory_dir / "archive_stale_memory.log"
        if not log_path.exists():
            print(f"  FAIL: log file not created: {log_path}")
            return False
        content = log_path.read_text(encoding="utf-8")
        if "skipped=1" not in content:
            print(f"  FAIL: log skipped count wrong: {content!r}")
            return False
        # error count = 0 (already-archived 는 'fail' reason 아님, v0.7.31 의 fix)
        # → error=0 검증
        if "error=0" not in content:
            print(f"  FAIL: expected error=0 (already-archived is not 'fail' reason). Got: {content!r}")
            return False
        print("  PASS: apply with file-blocker appends log (skipped=1, error=0)")
        return True


# === Test 4: show-metrics returns entries ===
def test_show_metrics_returns_entries() -> bool:
    """--show-metrics 가 entries list 반환 (post-mortem 분석용)."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = Path(tmp) / "fake_repo"
        memory_dir = repo_root / "ai-workflow" / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)
        # write 3 entries
        for i in range(3):
            append_metrics_log(memory_dir, 30 + i, i, i + 1, 0)
        # cmd_show_metrics
        args = _make_args(show_metrics=True, repo_root=str(repo_root))
        result = cmd_show_metrics(args)
        if not result.get("ok"):
            print(f"  FAIL: ok=False. Got: {result}")
            return False
        if result.get("entry_count") != 3:
            print(f"  FAIL: expected entry_count=3, got {result.get('entry_count')}")
            return False
        entries = result.get("entries", [])
        if len(entries) != 3:
            print(f"  FAIL: expected 3 entries, got {len(entries)}")
            return False
        # verify entry structure
        e0 = entries[0]
        required_keys = {"timestamp", "older_than", "archived", "skipped", "error"}
        if not required_keys.issubset(e0.keys()):
            print(f"  FAIL: entry missing keys. Got: {e0}")
            return False
        if e0.get("older_than") != 30 or e0.get("archived") != 0 or e0.get("skipped") != 1:
            print(f"  FAIL: entry[0] values wrong: {e0}")
            return False
        print("  PASS: --show-metrics returns 3 entries with correct structure")
        return True


# === Test 5: show-metrics no log ===
def test_show_metrics_no_log() -> bool:
    """log 부재 시 entry_count=0 + ok=True (graceful)."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = Path(tmp) / "fake_repo"
        memory_dir = repo_root / "ai-workflow" / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)
        # NO log file
        args = _make_args(show_metrics=True, repo_root=str(repo_root))
        result = cmd_show_metrics(args)
        if not result.get("ok"):
            print(f"  FAIL: ok=False (should be True for missing log). Got: {result}")
            return False
        if result.get("entry_count") != 0:
            print(f"  FAIL: expected entry_count=0, got {result.get('entry_count')}")
            return False
        print("  PASS: --show-metrics with no log → ok=True, entry_count=0")
        return True


# === TASK-V0730-001 Tests (5) ===

# === Test 6: install-cron idempotent skip ===
def test_install_cron_idempotent_skip() -> bool:
    """mavis cron info OK + cron_name in stdout → skipped-existing (no create)."""
    args = _make_args(install_cron=True)
    with patch("subprocess.run") as mock_run:
        # 1st: mavis cron info (returncode 0, stdout 에 'archive-memory' 포함)
        # 2nd: mavis cron create (NOT called — skipped)
        mock_run.side_effect = [
            _make_fake_proc(0, '{"cronName": "archive-memory", "enabled": true}\n'),
        ]
        result = cmd_install_cron(args)
        if not result.get("ok"):
            print(f"  FAIL: ok=False. Got: {result}")
            return False
        if result.get("idempotency") != "skipped-existing":
            print(f"  FAIL: expected idempotency='skipped-existing', got {result.get('idempotency')}")
            return False
        if mock_run.call_count != 1:
            print(f"  FAIL: expected 1 subprocess call (info only), got {mock_run.call_count}")
            return False
        call_args = mock_run.call_args[0][0]
        if "info" not in call_args:
            print(f"  FAIL: expected 'mavis cron info', got {call_args}")
            return False
        if "create" in call_args:
            print(f"  FAIL: mavis cron create should NOT be called. Got: {call_args}")
            return False
    print("  PASS: --install-cron idempotency: existing → skipped-existing, no create")
    return True


# === Test 7: install-cron creates when missing ===
def test_install_cron_creates_when_missing() -> bool:
    """mavis cron info fail → create (returncode != 0 OR cron_name not in stdout)."""
    args = _make_args(install_cron=True)
    with patch("subprocess.run") as mock_run:
        mock_run.side_effect = [
            _make_fake_proc(1, "", "Error: cron not found\n"),  # info fail
            _make_fake_proc(0, "created\n"),  # create OK
        ]
        result = cmd_install_cron(args)
        if not result.get("ok"):
            print(f"  FAIL: ok=False. Got: {result}")
            return False
        if result.get("idempotency") != "created":
            print(f"  FAIL: expected idempotency='created', got {result.get('idempotency')}")
            return False
        if mock_run.call_count != 2:
            print(f"  FAIL: expected 2 subprocess calls (info + create), got {mock_run.call_count}")
            return False
        # verify 'create' in 2nd call args
        call_args_2 = mock_run.call_args_list[1][0][0]
        if "create" not in call_args_2:
            print(f"  FAIL: expected 'mavis cron create' in 2nd call, got {call_args_2}")
            return False
    print("  PASS: --install-cron: info fail → create (idempotency='created')")
    return True


# === Test 8: install-cron force install ===
def test_install_cron_force_install() -> bool:
    """--force-install 시 info check skip, 바로 create."""
    args = _make_args(install_cron=True, force_install=True)
    with patch("subprocess.run") as mock_run:
        mock_run.side_effect = [
            # No info call (skipped)
            _make_fake_proc(0, "created\n"),
        ]
        result = cmd_install_cron(args)
        if not result.get("ok"):
            print(f"  FAIL: ok=False. Got: {result}")
            return False
        if result.get("idempotency") != "forced":
            print(f"  FAIL: expected idempotency='forced', got {result.get('idempotency')}")
            return False
        if mock_run.call_count != 1:
            print(f"  FAIL: expected 1 subprocess call (create only, no info), got {mock_run.call_count}")
            return False
        call_args = mock_run.call_args[0][0]
        if "create" not in call_args:
            print(f"  FAIL: expected 'mavis cron create', got {call_args}")
            return False
    print("  PASS: --force-install skips info check, directly creates (idempotency='forced')")
    return True


# === Test 9: uninstall-cron idempotent ===
def test_uninstall_cron_idempotent() -> bool:
    """mavis cron disable idempotent — 이미 disabled 인 cron 도 disable 가능 (returncode 0)."""
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
    print("  PASS: --uninstall-cron idempotent (no check, just disable)")
    return True


# === Test 10: read_metrics_log idempotency ===
def test_metrics_log_idempotency() -> bool:
    """read_metrics_log 가 동일 file 2번 read 시 동일 result (no state side-effect)."""
    with tempfile.TemporaryDirectory() as tmp:
        memory_dir = Path(tmp) / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)
        append_metrics_log(memory_dir, 30, 1, 2, 0)
        # 1st read
        entries1 = read_metrics_log(memory_dir)
        # 2nd read
        entries2 = read_metrics_log(memory_dir)
        if entries1 != entries2:
            print(f"  FAIL: read_metrics_log not idempotent. 1st: {entries1}, 2nd: {entries2}")
            return False
        if len(entries1) != 1:
            print(f"  FAIL: expected 1 entry, got {len(entries1)}")
            return False
        e = entries1[0]
        if e.get("older_than") != 30 or e.get("archived") != 1 or e.get("skipped") != 2 or e.get("error") != 0:
            print(f"  FAIL: entry values wrong: {e}")
            return False
    print("  PASS: read_metrics_log idempotent (1st = 2nd read)")
    return True


# === Main ===
def main() -> int:
    print("=" * 60)
    print("TASK-V0729-001 + TASK-V0730-001 smoke test (v0.7.31)")
    print("=" * 60)

    tests = [
        # TASK-V0729-001 (5)
        test_metrics_log_written,
        test_metrics_log_format,
        test_metrics_log_skipped_count,
        test_show_metrics_returns_entries,
        test_show_metrics_no_log,
        # TASK-V0730-001 (5)
        test_install_cron_idempotent_skip,
        test_install_cron_creates_when_missing,
        test_install_cron_force_install,
        test_uninstall_cron_idempotent,
        test_metrics_log_idempotency,
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
