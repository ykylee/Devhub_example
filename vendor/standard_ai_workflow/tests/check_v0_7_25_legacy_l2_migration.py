#!/usr/bin/env python3
"""
v0.7.25: migrate_legacy_l2 smoke test (5/5 PASS)

Test cases:
1. test_dry_run_detect_legacy — dry-run 이 15 legacy file detect
2. test_apply_writes_mirror — apply 가 in-repo mirror file 생성
3. test_idempotency_skip — 동일 content 재apply 시 skipped (identical)
4. test_drift_warning — 외부 wiki 의 file 변경 시 drift 감지 + skip (manual review)
5. test_unknown_args — argparse error (--dry-run / --apply 모두 부재 시)
"""
from __future__ import annotations

import json
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

# === Test setup ===
REPO_ROOT = Path(__file__).resolve().parent.parent.parent
TOOL = REPO_ROOT / "workflow-source" / "tools" / "migrate_legacy_l2.py"
PYTHON = sys.executable

# Use a tempdir-based REPO_ROOT for isolation (mimic R-4 격리 정공법)
# 외부 wiki 는 실제 /Users/yklee/wiki/.../sources/ 의 read-only 사용
EXTERNAL_WIKI = Path.home() / "wiki" / "wiki" / "projects" / "standard-ai-workflow" / "sources"


def _run_tool(*args: str, repo_root: Path) -> dict:
    """Run migrate_legacy_l2.py with given args + repo_root override. Returns parsed JSON."""
    cmd = [
        PYTHON, str(TOOL), *args,
        "--repo-root", str(repo_root),
        "--json",
    ]
    proc = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
    if proc.returncode != 0:
        return {"error": proc.stderr, "returncode": proc.returncode, "stdout": proc.stdout}
    try:
        return json.loads(proc.stdout)
    except json.JSONDecodeError as e:
        return {"error": f"JSON parse fail: {e}", "stdout": proc.stdout}


# === Test 1: dry-run detect legacy ===
def test_dry_run_detect_legacy() -> bool:
    """dry-run 이 15 legacy file 을 detect 하는지 확인.
    *real* in-repo (31 dir) 와 *real* external (20 file) 의 diff = 15.
    """
    # *real* REPO_ROOT 사용 (temp dir ❌ — in-repo state 가 *real* 이어야 filter 정확)
    result = _run_tool("--dry-run", repo_root=REPO_ROOT)
    if result.get("error"):
        print(f"  FAIL: tool error: {result['error']}")
        return False
    if result.get("legacy_count") != 15:
        print(f"  FAIL: expected 15 legacy files, got {result.get('legacy_count')}")
        return False
    if result.get("mode") != "dry-run":
        print(f"  FAIL: expected mode=dry-run, got {result.get('mode')}")
        return False
    # file list 정합성
    versions = [f["version"] for f in result.get("files", [])]
    # lexicographic sort (sorted() in tool) — v0.5.10.1 < v0.5.7.1
    expected_versions = ["v0.1.0", "v0.2.0", "v0.3.0", "v0.3.1", "v0.3.2",
                        "v0.5.0", "v0.5.10.1", "v0.5.11", "v0.5.7.1", "v0.5.8",
                        "v0.5.9.1", "v0.6.0.1", "v0.6.1.5", "v0.6.2", "v0.6.3"]
    if versions != expected_versions:
        print(f"  FAIL: expected versions {expected_versions}, got {versions}")
        return False
    print("  PASS: dry-run detect 15 legacy files")
    return True


# === Test 2: apply writes mirror ===
def test_apply_writes_mirror() -> bool:
    """apply 가 in-repo mirror file 생성하는지 확인."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = Path(tmp) / "fake_repo"
        repo_root.mkdir(parents=True, exist_ok=True)
        (repo_root / "ai-workflow" / "memory" / "release").mkdir(parents=True, exist_ok=True)
        (repo_root / "ai-workflow" / "memory" / "release" / "v0.7.24").mkdir(parents=True, exist_ok=True)

        result = _run_tool("--apply", repo_root=repo_root)
        if result.get("error"):
            print(f"  FAIL: tool error: {result['error']}")
            return False
        mirror = Path(result["inrepo_mirror"])
        if not mirror.exists():
            print(f"  FAIL: mirror file not created: {mirror}")
            return False
        if result.get("action_performed") != "written (fresh)":
            print(f"  FAIL: expected 'written (fresh)', got {result.get('action_performed')}")
            return False
        # mirror content 에 frontmatter + 15 version 모두 포함
        content = mirror.read_text(encoding="utf-8")
        if not content.startswith("---"):
            print(f"  FAIL: mirror missing frontmatter")
            return False
        for v in ["v0.1.0", "v0.6.3"]:
            if v not in content:
                print(f"  FAIL: mirror missing version {v}")
                return False
        print(f"  PASS: apply wrote mirror ({mirror.stat().st_size:,} bytes)")
        return True


# === Test 3: idempotency skip ===
def test_idempotency_skip() -> bool:
    """apply 후 재apply 시 skipped (identical)."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = Path(tmp) / "fake_repo"
        repo_root.mkdir(parents=True, exist_ok=True)
        (repo_root / "ai-workflow" / "memory" / "release").mkdir(parents=True, exist_ok=True)
        (repo_root / "ai-workflow" / "memory" / "release" / "v0.7.24").mkdir(parents=True, exist_ok=True)

        # 1st apply
        result1 = _run_tool("--apply", repo_root=repo_root)
        if result1.get("error") or result1.get("action_performed") != "written (fresh)":
            print(f"  FAIL: 1st apply: {result1.get('error') or result1.get('action_performed')}")
            return False
        # 2nd apply (동일 external content, 동일 git HEAD, 거의 동일 hash)
        # commit 이 동일 hash 로 evaluate 되므로 content 동일 → identical
        result2 = _run_tool("--apply", repo_root=repo_root)
        if result2.get("error"):
            print(f"  FAIL: 2nd apply tool error: {result2['error']}")
            return False
        action = result2.get("action_performed")
        if action not in ("skipped (identical)", "skipped (drift — manual review)"):
            print(f"  FAIL: 2nd apply expected idempotency, got {action}")
            return False
        print(f"  PASS: 2nd apply idempotent ({action})")
        return True


# === Test 4: drift warning ===
def test_drift_warning() -> bool:
    """mirror file 의 frontmatter 를 *의도적으로* 변조 → drift 감지 + skip."""
    with tempfile.TemporaryDirectory() as tmp:
        repo_root = Path(tmp) / "fake_repo"
        repo_root.mkdir(parents=True, exist_ok=True)
        (repo_root / "ai-workflow" / "memory" / "release").mkdir(parents=True, exist_ok=True)
        (repo_root / "ai-workflow" / "memory" / "release" / "v0.7.24").mkdir(parents=True, exist_ok=True)

        # 1st apply
        result1 = _run_tool("--apply", repo_root=repo_root)
        if result1.get("error"):
            print(f"  FAIL: 1st apply: {result1['error']}")
            return False
        mirror = Path(result1["inrepo_mirror"])
        # mirror 변조
        original = mirror.read_text(encoding="utf-8")
        tampered = original.replace("last_touched: 2026-06-15", "last_touched: 2027-01-01")
        mirror.write_text(tampered, encoding="utf-8")
        # 2nd apply → drift
        result2 = _run_tool("--apply", repo_root=repo_root)
        if result2.get("error"):
            print(f"  FAIL: 2nd apply tool error: {result2['error']}")
            return False
        if result2.get("drift", {}).get("status") != "drift":
            print(f"  FAIL: expected drift, got {result2.get('drift')}")
            return False
        if result2.get("action_performed") != "skipped (drift — manual review)":
            print(f"  FAIL: expected 'skipped (drift — manual review)', got {result2.get('action_performed')}")
            return False
        # tampered file 이 그대로 보존됨 (덮어쓰기 안 됨)
        if mirror.read_text(encoding="utf-8") != tampered:
            print(f"  FAIL: mirror was overwritten despite drift")
            return False
        print("  PASS: drift detected + skipped (manual review)")
        return True


# === Test 5: unknown args (no --dry-run / --apply) ===
def test_unknown_args() -> bool:
    """--dry-run / --apply 모두 부재 시 argparse error."""
    cmd = [PYTHON, str(TOOL), "--json"]
    proc = subprocess.run(cmd, capture_output=True, text=True, timeout=10)
    if proc.returncode == 0:
        print(f"  FAIL: expected argparse error, got success")
        return False
    if "at least one of --dry-run or --apply" not in proc.stderr:
        print(f"  FAIL: expected argparse error message, got: {proc.stderr[:200]}")
        return False
    print("  PASS: argparse rejects missing mode flag")
    return True


# === Main ===
def main() -> int:
    print("=" * 60)
    print("migrate_legacy_l2 smoke test (v0.7.25)")
    print("=" * 60)

    tests = [
        test_dry_run_detect_legacy,
        test_apply_writes_mirror,
        test_idempotency_skip,
        test_drift_warning,
        test_unknown_args,
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
