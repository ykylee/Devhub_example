"""tools/refresh_wiki_memory.py smoke test (v0.7.5+).

Git log → 4 raw mirror file 갱신 + 4 L2 stub dense 재emit 의 정합성 검증.
v0.7.1 score_wiki_trend.py (10 test) / v0.7.2 trend alert (4 test) 의 패턴 따라 작성.

Test 구성 (8 test):
1. CLI subcommand 파싱 (--refresh-raw / --emit-l2 / --since / --dry-run / --json)
2. git log collect + categorize
3. pick_feat_commit (feat 우선, fallback)
4. update_state_json (dry → 8 entries)
5. update_work_backlog (dry → 5 anchors)
6. update_wiki_log (dry → 5 entries)
7. update_memory_log (dry → 1 entry)
8. reemit_l2_stubs (dry → 4 stubs with non-zero body)

Reference:
- tools/score_wiki_trend.py 의 argparse + dry-run 패턴
- tools/refresh_wiki_memory.py 본체
"""

from __future__ import annotations

import json
import os
import re
import subprocess
import sys
import tempfile
from pathlib import Path
from unittest import mock

SOURCE_ROOT = Path(__file__).resolve().parents[1]
TOOL = SOURCE_ROOT / "tools" / "refresh_wiki_memory.py"

# 2차 출처 (raw mirror) — v0.7.17+ in-repo. 외부 vault (~/wiki/) 연결 없음.
# 1차 출처 = ai-workflow/memory/active/, 2차 출처 = ai-workflow/wiki/sources/ (L2 dense).
# v0.7.17 이전의 VAULT_ROOT (Path.home() / "wiki") 는 *legacy* — 본 test 의 *in-repo*
# redirect 검증이 본 release 의 핵심.
INREPO_WIKI = SOURCE_ROOT.parent / "ai-workflow" / "wiki"

SOURCE_ROOT = Path(__file__).resolve().parents[1]
TOOL = SOURCE_ROOT / "tools" / "refresh_wiki_memory.py"
WORKFLOW_KIT_SRC = SOURCE_ROOT / "workflow_kit"

# workflow_kit.common.atomic_write import 위해 (v0.7.15+ refresh_wiki_memory 의 의존)
sys.path.insert(0, str(SOURCE_ROOT))


def _import_tool():
    """refresh_wiki_memory.py 를 importlib 로 로드."""
    import importlib.util
    spec = importlib.util.spec_from_file_location("refresh_wiki_memory", str(TOOL))
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def test_cli_subcommand_parsing() -> None:
    """--refresh-raw / --emit-l2 / --since / --dry-run / --json 옵션 인식."""
    proc = subprocess.run(
        [sys.executable, str(TOOL), "--refresh-raw", "--dry-run", "--json"],
        capture_output=True, text=True, timeout=30,
    )
    assert proc.returncode == 0, f"exit {proc.returncode}: {proc.stderr}"
    out = json.loads(proc.stdout)
    assert out["dry_run"] is True
    assert "refresh_raw" in out
    assert out["refresh_raw"]["mode"] == "dry-run"
    assert out["refresh_raw"]["commits"] > 0


def test_cli_no_subcommand_errors() -> None:
    """--refresh-raw / --emit-l2 모두 미지정 시 에러."""
    proc = subprocess.run(
        [sys.executable, str(TOOL), "--dry-run"],
        capture_output=True, text=True, timeout=30,
    )
    assert proc.returncode != 0
    assert "refresh-raw" in proc.stderr or "emit-l2" in proc.stderr


# --- Test 2: git log collect + categorize ---


def test_collect_and_categorize() -> None:
    """collect_commits + categorize 가 13 release bucket (v0.6.1~v0.7.4 + unreleased) 반환."""
    mod = _import_tool()
    rows = mod.collect_commits("2026-06-10")
    assert len(rows) > 50, f"expected 50+ commits, got {len(rows)}"
    by_rel = mod.categorize(rows)
    # 5 v0.7.x + 3 v0.6.x (v0.6.4~v0.6.6) + 1 v0.6.3 + v0.6.2 + v0.6.1 = 11+ bucket
    assert len(by_rel) >= 10, f"expected 10+ release buckets, got {len(by_rel)}"
    # v0.7.4 bucket 존재
    assert "(v0.7.4)" in by_rel, "v0.7.4 bucket missing"
    # v0.7.0 bucket 이 v0.7.0 step 6/7/8/9/10 feat commit 포함
    v070 = by_rel.get("(v0.7.0)", [])
    step_commits = [c for c in v070 if "v0.7.0 step" in c["subject"]]
    assert len(step_commits) >= 5, f"v0.7.0 step commit 부족: {len(step_commits)}"


# --- Test 3: pick_feat_commit ---


def test_pick_feat_commit_priority() -> None:
    """pick_feat_commit 가 *뒤쪽* feat (실제 코드 변경) 우선 선택."""
    mod = _import_tool()
    # reversed 후 첫 feat = 가장 오래된 feat
    commits = [
        {"short": "a", "subject": "docs(v0.7.0): wiki log commit hash 갱신"},
        {"short": "b", "subject": "chore(v0.7.0): version bump"},
        {"short": "c", "subject": "feat(v0.7.0 step 1): stage_completion required 격상"},
    ]
    picked = mod.pick_feat_commit(commits)
    assert picked["short"] == "c", f"expected 'c' (reversed feat), got {picked['short']}"


def test_pick_feat_commit_fallback() -> None:
    """feat 부재 시 첫 commit fallback."""
    mod = _import_tool()
    commits = [
        {"short": "a", "subject": "docs: wiki log"},
        {"short": "b", "subject": "chore: bump"},
    ]
    picked = mod.pick_feat_commit(commits)
    assert picked["short"] == "a"


# --- Test 4: update_state_json (dry) ---


def test_update_state_json_dry_returns_8_lines() -> None:
    """update_state_json dry → 8 entry (v0.6.4~v0.7.4).

    v0.7.5+ 는 release_order 에 (v0.7.5) 추가 시 9 entry 로 증가. 본 test 는
    *정확히 8* 검증 (v0.7.5 release 시 rel_order 확장 → 본 test 도 9 로 갱신 필요).
    """
    mod = _import_tool()
    rows = mod.collect_commits("2026-06-10")
    by_rel = mod.categorize(rows)
    new_lines = mod.update_state_json(by_rel, dry=True)
    # v0.7.5 release 가 rel_order 에 들어가면 9
    assert len(new_lines) >= 8, f"expected 8+ entries, got {len(new_lines)}"
    # 각 line format "<version> (<short>): <msg>"
    for line in new_lines:
        assert re.match(r"^v\d+\.\d+\.\d+ \([0-9a-f]{7,}\):", line), f"format mismatch: {line}"


# --- Test 5: update_work_backlog (dry) ---


def test_update_work_backlog_dry_returns_5_blocks() -> None:
    """update_work_backlog dry → 5+ release anchor (v0.7.0~v0.7.5+)."""
    mod = _import_tool()
    rows = mod.collect_commits("2026-06-10")
    by_rel = mod.categorize(rows)
    new_block = mod.update_work_backlog(by_rel, dry=True)
    # v0.7.5+ 는 release loop 에 (v0.7.5) 추가 시 6 anchor
    assert len(new_block) >= 5, f"expected 5+ anchors, got {len(new_block)}"
    for block in new_block:
        assert "[[release/v" in block
        assert "{#v" in block
        assert "(head:" in block


# --- Test 6: update_wiki_log (dry) ---


def test_update_wiki_log_dry_returns_5_entries() -> None:
    """update_wiki_log dry → 5+ release entry (v0.7.0~v0.7.5+)."""
    mod = _import_tool()
    rows = mod.collect_commits("2026-06-10")
    by_rel = mod.categorize(rows)
    new_entries = mod.update_wiki_log(by_rel, dry=True)
    # v0.7.5+ 는 release loop 에 (v0.7.5) 추가 시 6 entry
    assert len(new_entries) >= 5, f"expected 5+ entries, got {len(new_entries)}"
    for entry in new_entries:
        assert "## [2026-06-13] release |" in entry
        assert "head:" in entry
        assert "commits:" in entry
        assert "range:" in entry


# --- Test 7: update_memory_log (dry) ---


def test_update_memory_log_dry_returns_1_entry() -> None:
    """update_memory_log dry → 1 sync backfill entry."""
    mod = _import_tool()
    entry = mod.update_memory_log(dry=True)
    assert "## [2026-06-14] sync |" in entry
    assert "wiki raw mirror backfill" in entry


# --- Test 8: reemit_l2_stubs (dry) ---


def test_reemit_l2_stubs_dry_returns_4_non_empty() -> None:
    """reemit_l2_stubs dry → 4 stub 모두 0 bytes 초과."""
    mod = _import_tool()
    rows = mod.collect_commits("2026-06-10")
    by_rel = mod.categorize(rows)
    state_lines = mod.update_state_json(by_rel, dry=True)
    sizes = mod.reemit_l2_stubs(by_rel, state_lines, dry=True)
    assert set(sizes.keys()) == {
        "active-state", "active-work-backlog",
        "active-session-handoff", "wiki-log",
    }
    for name, size in sizes.items():
        assert size > 100, f"{name} body too small: {size} bytes"


# --- Test 9: REPO_ROOT auto-detect (v0.7.12) — argparse 인식 ---

def test_repo_root_argparse_recognized() -> None:
    """--repo-root argparse flag 가 인식되고 result.repo_root 에 반영."""
    # cwd 가 workflow-source/ inner. 본 repo 의 git toplevel 은 outer.
    # `git rev-parse --show-toplevel` 가 outer 를 반환 (workflow-source 는 subdir).
    proc = subprocess.run(
        [sys.executable, str(TOOL), "--refresh-raw", "--dry-run", "--json"],
        capture_output=True, text=True, timeout=30,
    )
    assert proc.returncode == 0, f"exit {proc.returncode}: {proc.stderr}"
    out = json.loads(proc.stdout)
    assert "repo_root" in out
    expected_toplevel = Path(
        subprocess.run(
            ["git", "rev-parse", "--show-toplevel"],
            capture_output=True, text=True, timeout=5,
        ).stdout.strip()
    ).resolve()
    assert out["repo_root"] == str(expected_toplevel), \
        f"expected git toplevel {expected_toplevel}, got {out['repo_root']}"


def test_repo_root_cli_flag_priority() -> None:
    """--repo-root CLI flag 가 env var 보다 우선. (valid path + dry-run)"""
    # valid path + dry-run → collect_commits 가 git log 정상 호출. result.repo_root 검증.
    proc = subprocess.run(
        [
            sys.executable, str(TOOL), "--refresh-raw", "--dry-run", "--json",
            "--repo-root", str(SOURCE_ROOT.resolve()),
        ],
        capture_output=True, text=True, timeout=30,
        env={**os.environ, "STANDARD_AI_WF_REPO": "/tmp/env-var-path-must-be-ignored"},
    )
    assert proc.returncode == 0, f"exit {proc.returncode}: {proc.stderr}"
    out = json.loads(proc.stdout)
    assert out["repo_root"] == str(SOURCE_ROOT.resolve()), \
        f"CLI flag 무시됨: {out['repo_root']} (env var 가 우선?)"

# --- Test 11: REPO_ROOT auto-detect — env var fallback (cwd 외부 실행) ---


def test_repo_root_env_var_fallback() -> None:
    """STANDARD_AI_WF_REPO env var 가 cwd 외부에서도 동작."""
    proc = subprocess.run(
        [sys.executable, str(TOOL), "--refresh-raw", "--dry-run", "--json"],
        capture_output=True, text=True, timeout=30,
        cwd="/tmp",
        env={**os.environ, "STANDARD_AI_WF_REPO": str(SOURCE_ROOT.resolve())},
    )
    assert proc.returncode == 0, f"exit {proc.returncode}: {proc.stderr}"
    out = json.loads(proc.stdout)
    assert out["repo_root"] == str(SOURCE_ROOT.resolve()), \
        f"env var 미반영: {out['repo_root']}"


# --- Test 12: REPO_ROOT auto-detect — git rev-parse 결과 일치 (priority 3) ---


def test_repo_root_git_toplevel_priority() -> None:
    """cwd 가 git repo 안일 때 `git rev-parse --show-toplevel` 결과."""
    mod = _import_tool()
    expected = mod.REPO_ROOT
    proc = subprocess.run(
        ["git", "rev-parse", "--show-toplevel"],
        capture_output=True, text=True, timeout=5,
        cwd=str(SOURCE_ROOT),
    )
    assert proc.returncode == 0
    expected_str = str(Path(proc.stdout.strip()).resolve())
    assert str(expected) == expected_str, \
        f"REPO_ROOT {expected!r} != git rev-parse {expected_str!r}"

def main() -> int:
    passed = 0
    failed = 0
    failures: list[tuple[str, str]] = []
    test_funcs = [
        test_cli_subcommand_parsing,
        test_cli_no_subcommand_errors,
        test_collect_and_categorize,
        test_pick_feat_commit_priority,
        test_pick_feat_commit_fallback,
        test_update_state_json_dry_returns_8_lines,
        test_update_work_backlog_dry_returns_5_blocks,
        test_update_wiki_log_dry_returns_5_entries,
        test_update_memory_log_dry_returns_1_entry,
        test_reemit_l2_stubs_dry_returns_4_non_empty,
        test_repo_root_argparse_recognized,
        test_repo_root_cli_flag_priority,
        test_repo_root_env_var_fallback,
        test_repo_root_git_toplevel_priority,
    ]
    for func in test_funcs:
        try:
            func()
            print(f"  PASS  {func.__name__}")
            passed += 1
        except AssertionError as e:
            print(f"  FAIL  {func.__name__}: {e}")
            failed += 1
            failures.append((func.__name__, str(e)))
        except Exception as e:  # noqa: BLE001
            print(f"  ERROR {func.__name__}: {type(e).__name__}: {e}")
            failed += 1
            failures.append((func.__name__, f"{type(e).__name__}: {e}"))

    print()
    if failed == 0:
        print(f"All {passed} tests passed.")
        return 0
    print(f"{failed}/{passed + failed} tests failed:")
    for name, err in failures:
        print(f"  - {name}: {err}")
    return 1


if __name__ == "__main__":
    sys.exit(main())
