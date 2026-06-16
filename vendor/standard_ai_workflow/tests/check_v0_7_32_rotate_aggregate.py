#!/usr/bin/env python3
"""
v0.7.32: TASK-V0731-001 (log rotation) + TASK-V0732-001 (metrics aggregation)
smoke test (10/10 PASS)

Test cases (TASK-V0731-001, 5 smoke):
1. test_rotate_log_under_threshold — line < threshold → no rotate
2. test_rotate_log_at_threshold — line = threshold → no rotate (strict <)
3. test_rotate_log_over_threshold — line > threshold → gzip + truncate
4. test_rotate_log_creates_gz_archive — archive path = log.2026-06-15T13-00-00Z.gz
5. test_rotate_log_read_rotated — read_rotated_logs 가 gz 의 entries read

Test cases (TASK-V0732-001, 5 smoke):
6. test_aggregate_weekly — 7 entries (1 week) → 1 bucket
7. test_aggregate_monthly — 30 entries (1 month) → 1 bucket
8. test_aggregate_multiple_weeks — 14 entries (2 weeks) → 2 buckets
9. test_aggregate_include_rotated — --include-rotated 시 rotated log entries 포함
10. test_aggregate_invalid_period — invalid period (e.g. 'yearly') → ok=False, error
"""
from __future__ import annotations

import datetime as dt
import gzip
import json
import os
import subprocess
import sys
import tempfile
from pathlib import Path
from unittest.mock import patch

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
cmd_rotate_logs = _am.cmd_rotate_logs
cmd_aggregate_metrics = _am.cmd_aggregate_metrics
rotate_log_if_needed = _am.rotate_log_if_needed
aggregate_metrics = _am.aggregate_metrics
read_rotated_logs = _am.read_rotated_logs
append_metrics_log = _am.append_metrics_log
read_metrics_log = _am.read_metrics_log
LOG_ROTATE_LINE_THRESHOLD = _am.LOG_ROTATE_LINE_THRESHOLD


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
        "rotate_logs": False,
        "rotate_threshold": LOG_ROTATE_LINE_THRESHOLD,
        "aggregate": None,
        "include_rotated": False,
        "cron_name": "archive-memory",
        "cron_interval": "7d",
        "agent": "mavis",
        "force_install": False,
    }
    defaults.update(kwargs)
    return argparse.Namespace(**defaults)


def _write_n_entries(memory_dir: Path, n: int) -> None:
    """n entries 를 archive_stale_memory.log 에 write."""
    log_path = memory_dir / "archive_stale_memory.log"
    for i in range(n):
        append_metrics_log(memory_dir, 30, archived_count=i, skipped_count=i + 1, error_count=0)


def _make_repo_root(tmp: str) -> Path:
    """cmd_aggregate_metrics 의 *real* path: repo_root/ai-workflow/memory.

    test 의 `memory_dir` 가 tool 의 *real* path 와 일치하도록.
    """
    repo_root = Path(tmp) / "fake_repo"
    (repo_root / "ai-workflow" / "memory").mkdir(parents=True, exist_ok=True)
    return repo_root


# === TASK-V0731-001 Tests (5) ===

# === Test 1: rotate_log under threshold ===
def test_rotate_log_under_threshold() -> bool:
    """line < threshold → no rotate."""
    with tempfile.TemporaryDirectory() as tmp:
        memory_dir = Path(tmp) / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)
        _write_n_entries(memory_dir, 100)  # 100 line
        result = rotate_log_if_needed(memory_dir, line_threshold=10000)
        if result["rotated"]:
            print(f"  FAIL: should not rotate (100 < 10000). Got: {result}")
            return False
        if result["line_count"] != 100:
            print(f"  FAIL: expected line_count=100, got {result['line_count']}")
            return False
        if result["archive_path"] is not None:
            print(f"  FAIL: archive_path should be None, got {result['archive_path']}")
            return False
    print("  PASS: line < threshold → no rotate")
    return True


# === Test 2: rotate_log at threshold (strict <) ===
def test_rotate_log_at_threshold() -> bool:
    """line = threshold → no rotate (strict < 만 trigger)."""
    with tempfile.TemporaryDirectory() as tmp:
        memory_dir = Path(tmp) / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)
        _write_n_entries(memory_dir, 100)  # 100 line
        result = rotate_log_if_needed(memory_dir, line_threshold=100)
        # 100 < 100 = False → no rotate
        if result["rotated"]:
            print(f"  FAIL: should NOT rotate at threshold (strict <). Got: {result}")
            return False
    print("  PASS: line = threshold (100) → no rotate (strict < only)")
    return True


# === Test 3: rotate_log over threshold ===
def test_rotate_log_over_threshold() -> bool:
    """line > threshold → gzip + truncate (log file = 0 byte)."""
    with tempfile.TemporaryDirectory() as tmp:
        memory_dir = Path(tmp) / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)
        _write_n_entries(memory_dir, 105)  # 105 line
        log_path = memory_dir / "archive_stale_memory.log"
        if log_path.stat().st_size == 0:
            print(f"  FAIL: log file empty before rotate")
            return False
        result = rotate_log_if_needed(memory_dir, line_threshold=100)
        if not result["rotated"]:
            print(f"  FAIL: should rotate (105 > 100). Got: {result}")
            return False
        if result["line_count"] != 105:
            print(f"  FAIL: expected line_count=105, got {result['line_count']}")
            return False
        # log file should be truncated to 0 byte
        if log_path.stat().st_size != 0:
            print(f"  FAIL: log not truncated. size={log_path.stat().st_size}")
            return False
        # archive should exist
        archive_path = Path(result["archive_path"])
        if not archive_path.exists():
            print(f"  FAIL: archive not created: {archive_path}")
            return False
        # archive should be .gz
        if not str(archive_path).endswith(".gz"):
            print(f"  FAIL: archive not .gz: {archive_path}")
            return False
    print("  PASS: line > threshold → gzip + truncate")
    return True


# === Test 4: rotate_log creates gz archive ===
def test_rotate_log_creates_gz_archive() -> bool:
    """archive path = log.2026-06-15T13-00-00Z.gz (timestamp 포함)."""
    with tempfile.TemporaryDirectory() as tmp:
        memory_dir = Path(tmp) / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)
        _write_n_entries(memory_dir, 105)
        result = rotate_log_if_needed(memory_dir, line_threshold=100)
        archive_path = result["archive_path"]
        if archive_path is None:
            print(f"  FAIL: archive_path is None")
            return False
        # verify format: archive_stale_memory.log.YYYY-MM-DDTHH-MM-SSZ.gz
        import re
        pattern = r"archive_stale_memory\.log\.\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}Z\.gz$"
        if not re.search(pattern, archive_path):
            print(f"  FAIL: archive path doesn't match pattern. Got: {archive_path}")
            return False
        # archive_size_bytes > 0
        if result["archive_size_bytes"] is None or result["archive_size_bytes"] <= 0:
            print(f"  FAIL: archive_size_bytes invalid. Got: {result['archive_size_bytes']}")
            return False
    print("  PASS: rotate creates timestamped .gz archive")
    return True


# === Test 5: rotate_log + read rotated ===
def test_rotate_log_read_rotated() -> bool:
    """read_rotated_logs 가 gz 의 entries 를 read + source field 포함."""
    with tempfile.TemporaryDirectory() as tmp:
        memory_dir = Path(tmp) / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)
        _write_n_entries(memory_dir, 105)  # trigger rotate
        result = rotate_log_if_needed(memory_dir, line_threshold=100)
        if not result["rotated"]:
            print(f"  FAIL: rotate failed")
            return False
        # read rotated
        rotated_entries = read_rotated_logs(memory_dir)
        # rotated entries 에 105 entry (정확) — main log 는 0 이고 *모든* entry 가 rotated 에
        if len(rotated_entries) != 105:
            print(f"  FAIL: expected 105 entries in rotated, got {len(rotated_entries)}")
            return False
        # source field 포함
        for e in rotated_entries:
            if "source" not in e or not str(e["source"]).endswith(".gz"):
                print(f"  FAIL: entry missing 'source' field. Got: {e}")
                return False
                break
    print("  PASS: read_rotated_logs reads gz entries with source field")
    return True


# === TASK-V0732-001 Tests (5) ===

# === Test 6: aggregate_weekly ===
def test_aggregate_weekly() -> bool:
    """7 entries (1 week) → 1 bucket. weekly aggregation."""
    with tempfile.TemporaryDirectory() as tmp:
        memory_dir = Path(tmp) / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)
        # 7 entries 같은 주 (2026-W25, KST 2026-06-15 ~ 21)
        for i in range(7):
            log_path = memory_dir / "archive_stale_memory.log"
            ts = f"2026-06-{15 + i:02d}T00:00:00Z"
            line = f"{ts}\tolder_than=30\tarchived={i}\tskipped={i+1}\terror=0\n"
            with log_path.open("a", encoding="utf-8") as f:
                f.write(line)
        entries = read_metrics_log(memory_dir)
        result = aggregate_metrics(entries, period="weekly")
        if result["period"] != "weekly":
            print(f"  FAIL: period != 'weekly'. Got: {result['period']}")
            return False
        if len(result["buckets"]) != 1:
            print(f"  FAIL: expected 1 weekly bucket, got {len(result['buckets'])}. Buckets: {result['buckets']}")
            return False
        b = result["buckets"][0]
        if b["count"] != 7:
            print(f"  FAIL: expected count=7 in 1 bucket, got {b['count']}")
            return False
        if b["archived"] != sum(range(7)):  # 0+1+2+...+6 = 21
            print(f"  FAIL: expected archived=21, got {b['archived']}")
            return False
    print("  PASS: 7 entries (1 week) → 1 weekly bucket")
    return True


# === Test 7: aggregate_monthly ===
def test_aggregate_monthly() -> bool:
    """30 entries (1 month) → 1 bucket."""
    with tempfile.TemporaryDirectory() as tmp:
        memory_dir = Path(tmp) / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)
        # 30 entries 같은 달
        for i in range(1, 31):
            ts = f"2026-06-{i:02d}T00:00:00Z"
            line = f"{ts}\tolder_than=30\tarchived={i}\tskipped=0\terror=0\n"
            log_path = memory_dir / "archive_stale_memory.log"
            with log_path.open("a", encoding="utf-8") as f:
                f.write(line)
        entries = read_metrics_log(memory_dir)
        result = aggregate_metrics(entries, period="monthly")
        if result["period"] != "monthly":
            print(f"  FAIL: period != 'monthly'. Got: {result['period']}")
            return False
        if len(result["buckets"]) != 1:
            print(f"  FAIL: expected 1 monthly bucket, got {len(result['buckets'])}")
            return False
        b = result["buckets"][0]
        if b["count"] != 30:
            print(f"  FAIL: expected count=30, got {b['count']}")
            return False
    print("  PASS: 30 entries (1 month) → 1 monthly bucket")
    return True


# === Test 8: aggregate multiple weeks ===
def test_aggregate_multiple_weeks() -> bool:
    """14 entries (2 weeks) → 2 buckets."""
    with tempfile.TemporaryDirectory() as tmp:
        memory_dir = Path(tmp) / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)
        # Week 1: 2026-06-15 ~ 21 (7 days)
        for i in range(7):
            ts = f"2026-06-{15 + i:02d}T00:00:00Z"
            line = f"{ts}\tolder_than=30\tarchived=1\tskipped=0\terror=0\n"
            with (memory_dir / "archive_stale_memory.log").open("a", encoding="utf-8") as f:
                f.write(line)
        # Week 2: 2026-06-22 ~ 28 (7 days, ISO W26)
        for i in range(7):
            ts = f"2026-06-{22 + i:02d}T00:00:00Z"
            line = f"{ts}\tolder_than=30\tarchived=2\tskipped=0\terror=0\n"
            with (memory_dir / "archive_stale_memory.log").open("a", encoding="utf-8") as f:
                f.write(line)
        entries = read_metrics_log(memory_dir)
        result = aggregate_metrics(entries, period="weekly")
        if len(result["buckets"]) != 2:
            print(f"  FAIL: expected 2 weekly buckets, got {len(result['buckets'])}. Buckets: {result['buckets']}")
            return False
        # buckets[0] = W25 (7 days, archived=1 each → total 7)
        # buckets[1] = W26 (7 days, archived=2 each → total 14)
        w1 = result["buckets"][0]
        w2 = result["buckets"][1]
        if w1["archived"] != 7 or w2["archived"] != 14:
            print(f"  FAIL: week 1 archived={w1['archived']} (expected 7), week 2 archived={w2['archived']} (expected 14)")
            return False
    print("  PASS: 14 entries (2 weeks) → 2 weekly buckets")
    return True


# === Test 9: aggregate include_rotated ===
def test_aggregate_include_rotated() -> bool:
    """--include-rotated 시 rotated log entries 포함 (main + rotated 모두)."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = _make_repo_root(tmp)
        memory_dir = repo_root / "ai-workflow" / "memory"
        # write 105 entries (trigger rotate later)
        _write_n_entries(memory_dir, 105)
        # rotate
        result = rotate_log_if_needed(memory_dir, line_threshold=100)
        if not result["rotated"]:
            print(f"  FAIL: rotate failed")
            return False
        # main log empty (0 entries)
        main_entries = read_metrics_log(memory_dir)
        rotated_entries = read_rotated_logs(memory_dir)
        if len(main_entries) != 0:
            print(f"  FAIL: main log should be empty after rotate, got {len(main_entries)}")
            return False
        if len(rotated_entries) != 105:
            print(f"  FAIL: rotated should have 105 entries, got {len(rotated_entries)}")
            return False
        # cmd_aggregate_metrics with --include-rotated
        args = _make_args(aggregate="weekly", include_rotated=True, repo_root=str(repo_root))
        result = cmd_aggregate_metrics(args)
        if not result["ok"]:
            print(f"  FAIL: cmd_aggregate_metrics.ok=False. Got: {result}")
            return False
        if result["entry_count"] != 105:
            print(f"  FAIL: expected entry_count=105, got {result['entry_count']}")
            return False
        if result["included_rotated"] != 105:
            print(f"  FAIL: expected included_rotated=105, got {result['included_rotated']}")
            return False
        # total.archived = 0+1+2+...+104 = 104*105/2 = 5460
        expected_archived = 104 * 105 // 2
        if result["total"]["archived"] != expected_archived:
            print(f"  FAIL: expected total.archived={expected_archived}, got {result['total']['archived']}")
            return False
    print("  PASS: --include-rotated aggregates main + rotated entries")
    return True


# === Test 10: aggregate invalid period ===
def test_aggregate_invalid_period() -> bool:
    """invalid --aggregate (e.g. 'yearly') → ok=False, error message."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = _make_repo_root(tmp)
        args = _make_args(aggregate="yearly", repo_root=str(repo_root))
        result = cmd_aggregate_metrics(args)
        if result.get("ok") is not False:
            print(f"  FAIL: expected ok=False, got {result}")
            return False
        if "invalid" not in result.get("error", "").lower() and "expected" not in result.get("error", "").lower():
            print(f"  FAIL: error message missing. Got: {result.get('error')!r}")
            return False
    print("  PASS: invalid --aggregate → ok=False, error message")
    return True


# === Main ===
def main() -> int:
    print("=" * 60)
    print("TASK-V0731-001 + TASK-V0732-001 smoke test (v0.7.32)")
    print("=" * 60)

    tests = [
        # TASK-V0731-001 (5)
        test_rotate_log_under_threshold,
        test_rotate_log_at_threshold,
        test_rotate_log_over_threshold,
        test_rotate_log_creates_gz_archive,
        test_rotate_log_read_rotated,
        # TASK-V0732-001 (5)
        test_aggregate_weekly,
        test_aggregate_monthly,
        test_aggregate_multiple_weeks,
        test_aggregate_include_rotated,
        test_aggregate_invalid_period,
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
