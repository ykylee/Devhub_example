#!/usr/bin/env python3
"""
v0.7.28: TASK-V0726-004 (archive_stale_memory) smoke test (5/5 PASS)

Test cases:
1. test_short_sha_dir_detected — 7-char hex dir name → short SHA dir 로 detect
2. test_age_filter — N day 이전 dir 만 archive 후보 (too-recent skip)
3. test_already_archived_skip — archive/ 하위에 *이미* 있는 dir skip
4. test_dry_run_no_move — dry-run 시 file 이동 0
5. test_apply_archives_old_dir — apply 시 옛 dir 이 archive/<date>/<sha>/ 로 move
"""
from __future__ import annotations

import json
import os
import shutil
import subprocess
import sys
import tempfile
import time
from datetime import datetime, timedelta, timezone
from pathlib import Path


# KST = UTC+9. test 의 archive date 가 KST 정합.
def _today_kst() -> str:
    """KST (UTC+9) 기준 오늘 날짜. archive target date 와 정합."""
    kst = timezone(timedelta(hours=9))
    return datetime.now(tz=kst).strftime("%Y-%m-%d")

REPO_ROOT = Path(__file__).resolve().parent.parent.parent
TOOL = REPO_ROOT / "workflow-source" / "tools" / "archive_stale_memory.py"
PYTHON = sys.executable


def _run_tool(*args: str, repo_root: Path) -> dict:
    """Run archive_stale_memory.py with given args. Returns parsed JSON."""
    cmd = [PYTHON, str(TOOL), *args, "--repo-root", str(repo_root)]
    proc = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
    if proc.returncode != 0:
        return {"error": proc.stderr, "returncode": proc.returncode, "stdout": proc.stdout}
    try:
        return json.loads(proc.stdout)
    except json.JSONDecodeError as e:
        return {"error": f"JSON parse fail: {e}", "stdout": proc.stdout}


# === Test 1: short SHA dir detected ===
def test_short_sha_dir_detected() -> bool:
    """7-char hex dir name → short SHA dir 로 detect. named branch (e.g. 'main', 'feat/x') 는 *불일치*."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = Path(tmp) / "fake_repo"
        memory_dir = repo_root / "ai-workflow" / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)
        # short SHA dir (3개)
        for sha in ["abc1234", "deadbee", "1234abc"]:
            (memory_dir / sha).mkdir(parents=True, exist_ok=True)
            (memory_dir / sha / "state.md").write_text("test", encoding="utf-8")
        # named branch dir (불일치)
        for name in ["main", "feat/x", "release-v1.0.0"]:
            (memory_dir / name).mkdir(parents=True, exist_ok=True)
        # file (not dir)
        (memory_dir / "abc1234.md").write_text("test", encoding="utf-8")

        result = _run_tool("--list", "--dry-run", repo_root=repo_root)
        if result.get("error"):
            print(f"  FAIL: tool error: {result['error']}")
            return False
        items = result.get("items", [])
        shas = [item["sha"] for item in items]
        if sorted(shas) != ["1234abc", "abc1234", "deadbee"]:
            print(f"  FAIL: expected 3 short SHA dirs, got {shas}")
            return False
        # named branch dir 가 *불포함* 확인
        for name in ["main", "feat/x", "release-v1.0.0"]:
            if name in shas:
                print(f"  FAIL: named branch dir {name!r} incorrectly included")
                return False
        print("  PASS: short SHA dir detected (named branch excluded)")
        return True


# === Test 2: age filter ===
def test_age_filter() -> bool:
    """N day 이전 dir 만 archive 후보. too-recent skip."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = Path(tmp) / "fake_repo"
        memory_dir = repo_root / "ai-workflow" / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)
        # old dir (40 day)
        old_dir = memory_dir / "ff00ff0"
        old_dir.mkdir(parents=True, exist_ok=True)
        (old_dir / "state.md").write_text("old", encoding="utf-8")
        old_mtime = time.time() - (40 * 86400)
        os.utime(old_dir, (old_mtime, old_mtime))
        # new dir (5 day) — different name
        new_dir = memory_dir / "ff00ff1"
        new_dir.mkdir(parents=True, exist_ok=True)
        (new_dir / "state.md").write_text("new", encoding="utf-8")
        # new mtime = now (default)

        result = _run_tool("--older-than=30", "--dry-run", repo_root=repo_root)
        if result.get("error"):
            print(f"  FAIL: tool error: {result['error']}")
            return False
        candidates = result.get("candidates", [])
        candidate_shas = [c["sha"] for c in candidates]
        skipped = result.get("skipped", [])
        skipped_shas = [s["sha"] for s in skipped]
        if "ff00ff0" not in candidate_shas:
            print(f"  FAIL: old dir not in candidates. candidates: {candidate_shas}, skipped: {skipped_shas}")
            return False
        if "ff00ff1" not in skipped_shas:
            print(f"  FAIL: new dir not in skipped. skipped: {skipped_shas}")
            return False
        # verify reason
        new_skipped = next(s for s in skipped if s["sha"] == "ff00ff1")
        if new_skipped.get("reason") != "too-recent":
            print(f"  FAIL: expected reason='too-recent', got {new_skipped}")
            return False
        print("  PASS: age filter (old in candidates, new skipped)")
        return True


# === Test 3: already-archived skip ===
def test_already_archived_skip() -> bool:
    """sha 가 archive/ 하위에 *이미* 있으면 skip (idempotency)."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = Path(tmp) / "fake_repo"
        memory_dir = repo_root / "ai-workflow" / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)
        # old dir (40 day)
        old_dir = memory_dir / "ff00ff2"
        old_dir.mkdir(parents=True, exist_ok=True)
        old_mtime = time.time() - (40 * 86400)
        os.utime(old_dir, (old_mtime, old_mtime))
        # already-archived (under archive/2026-06-01/ff00ff2/)
        archive_target = memory_dir / "archive" / "2026-06-01" / "ff00ff2"
        archive_target.mkdir(parents=True, exist_ok=True)

        result = _run_tool("--older-than=30", "--dry-run", repo_root=repo_root)
        if result.get("error"):
            print(f"  FAIL: tool error: {result['error']}")
            return False
        candidates = result.get("candidates", [])
        skipped = result.get("skipped", [])
        if any(c["sha"] == "ff00ff2" for c in candidates):
            print(f"  FAIL: ff00ff2 should NOT be in candidates (already archived)")
            return False
        skipped_old = next((s for s in skipped if s["sha"] == "ff00ff2"), None)
        if not skipped_old or skipped_old.get("reason") != "already-archived":
            print(f"  FAIL: expected 'already-archived' skip. Got: {skipped}")
            return False
        print("  PASS: already-archived skip (idempotency)")
        return True


# === Test 4: dry-run no move ===
def test_dry_run_no_move() -> bool:
    """--dry-run 시 file 이동 0. memory dir 의 dir 수 변경 없음."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = Path(tmp) / "fake_repo"
        memory_dir = repo_root / "ai-workflow" / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)
        old_dir = memory_dir / "ff00ff0"
        old_dir.mkdir(parents=True, exist_ok=True)
        old_mtime = time.time() - (40 * 86400)
        os.utime(old_dir, (old_mtime, old_mtime))

        # dir 수 (top-level) 측정
        before_dirs = {p.name for p in memory_dir.iterdir() if p.is_dir()}
        result = _run_tool("--older-than=30", "--dry-run", repo_root=repo_root)
        after_dirs = {p.name for p in memory_dir.iterdir() if p.is_dir()}
        if result.get("error"):
            print(f"  FAIL: tool error: {result['error']}")
            return False
        if before_dirs != after_dirs:
            print(f"  FAIL: dirs changed during dry-run. before={before_dirs}, after={after_dirs}")
            return False
        # archive dir 가 생성되지 *않음* 확인
        if (memory_dir / "archive").exists():
            print(f"  FAIL: archive dir created during dry-run")
            return False
        print("  PASS: --dry-run: no file move, no archive dir created")
        return True


# === Test 5: apply archives old dir ===
def test_apply_archives_old_dir() -> bool:
    """--apply 시 옛 dir 이 archive/<date>/<sha>/ 로 move. SHA256 동일성 확인."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = Path(tmp) / "fake_repo"
        memory_dir = repo_root / "ai-workflow" / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)
        old_dir = memory_dir / "ff00ff3"
        old_dir.mkdir(parents=True, exist_ok=True)
        test_content = "important state data"
        (old_dir / "state.md").write_text(test_content, encoding="utf-8")
        old_mtime = time.time() - (40 * 86400)
        os.utime(old_dir, (old_mtime, old_mtime))

        result = _run_tool("--older-than=30", "--apply", repo_root=repo_root)
        if result.get("error"):
            print(f"  FAIL: tool error: {result['error']}")
            return False
        # old dir 가 move 됨
        if old_dir.exists():
            print(f"  FAIL: old dir still exists after apply: {old_dir}")
            return False
        # archive/<date>/old1234/state.md 가 존재
        archive_base = memory_dir / "archive"
        if not archive_base.exists():
            print(f"  FAIL: archive dir not created")
            return False
        today = datetime.now(tz=timezone.utc).strftime("%Y-%m-%d")
        target = archive_base / _today_kst() / "ff00ff3"
        if not target.exists():
            print(f"  FAIL: target dir not created: {target}")
            return False
        archived_state = target / "state.md"
        if not archived_state.exists():
            print(f"  FAIL: archived state.md not found: {archived_state}")
            return False
        if archived_state.read_text(encoding="utf-8") != test_content:
            print(f"  FAIL: archived content mismatch")
            return False
        # result dict 에 archived_count
        if result.get("archived_count") != 1:
            print(f"  FAIL: expected archived_count=1, got {result.get('archived_count')}")
            return False
        if result.get("mode") != "apply":
            print(f"  FAIL: expected mode='apply', got {result.get('mode')}")
            return False
        print("  PASS: --apply moves old dir to archive/<date>/<sha>/ (content preserved)")
        return True


# === Main ===
def main() -> int:
    print("=" * 60)
    print("TASK-V0726-004 (archive_stale_memory) smoke test (v0.7.28)")
    print("=" * 60)

    tests = [
        test_short_sha_dir_detected,
        test_age_filter,
        test_already_archived_skip,
        test_dry_run_no_move,
        test_apply_archives_old_dir,
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
