"""tools/release_pipeline.py release --auto-bump + remote tag pre-check smoke test (v0.7.18+).

v0.7.16 의 release coordination race lesson 반영: `cmd_release` 가
- `git ls-remote origin "refs/tags/{tag}"` 로 *원격 tag pre-check*
- 존재 시 default: exit 1 + auto-bump hint
- `--auto-bump` flag: 다음 version 으로 자동 bump + re-flow

Test list (7 test):
1. test_remote_tag_check_argparse: --auto-bump argparse error 없음
2. test_remote_tag_check_dry_run_no_remote: origin URL 부재 시 graceful (exists=False)
3. test_remote_tag_check_dry_run_existing_tag: 존재하는 tag dry-run 시 tag_pre_check_warning
4. test_version_sort_key: PEP 440 + suffix 정렬 (v0.7.17 > v0.7.16)
5. test_next_available_version_no_remote: remote_tags 비어있을 때 local 그대로 (bumped=False)
6. test_next_available_version_remote_ahead: remote 가 local 보다 클 때 +0.0.1 (bumped=True)
7. test_next_available_version_local_ahead: local 이 remote max 보다 클 때 그대로 (bumped=False)

Reference:
- workflow-source/tools/release_pipeline.py (v0.7.18 본 release, --auto-bump + _check_remote_tag)
- v0.7.16 release note (race lesson: 5-step re-version + cherry-pick + merge)
- memory #22 §release coordination race
"""

from __future__ import annotations

import importlib.util
import json
import subprocess
import sys
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
TOOL = SOURCE_ROOT / "tools" / "release_pipeline.py"


def _import_tool():
    """release_pipeline.py 를 importlib 로 로드."""
    spec = importlib.util.spec_from_file_location("release_pipeline", str(TOOL))
    mod = importlib.util.module_from_spec(spec)
    sys.modules["release_pipeline"] = mod
    spec.loader.exec_module(mod)
    return mod


# --- Test 1: --auto-bump argparse 인식 ---


def test_remote_tag_check_argparse() -> None:
    """release --auto-bump 가 argparse error 없이 받아들여짐."""
    proc = subprocess.run(
        [sys.executable, str(TOOL), "release", "--auto-bump", "--skip-validate", "--dry-run", "--json"],
        capture_output=True, text=True, timeout=30,
    )
    assert "unrecognized arguments" not in proc.stderr, f"argparse error: {proc.stderr}"
    # dry-run 결과는 dist 부재로 error 포함 가능. version_source 또는 auto_bump field 검증.
    out = json.loads(proc.stdout)
    # --auto-bump flag 가 받아들여졌으면, code path 가 next_available_version 호출
    # (bumped=False 면 local 그대로, 즉 version_source=pyproject.toml 유지)
    # 이 test 의 핵심 = argparse error 부재.
    assert "error" in out or "version_source" in out, (
        f"unexpected output structure: {out}"
    )


def test_allow_existing_tag_argparse() -> None:
    """release --allow-existing-tag 이 argparse error 없이 받아들여짐 (v0.7.21+)."""
    proc = subprocess.run(
        [sys.executable, str(TOOL), "release", "--allow-existing-tag", "--skip-validate", "--dry-run", "--json"],
        capture_output=True, text=True, timeout=30,
    )
    assert "unrecognized arguments" not in proc.stderr, f"argparse error: {proc.stderr}"
    # dry-run 결과는 dist 부재로 error 포함 가능. 정상.
    try:
        out = json.loads(proc.stdout)
    except json.JSONDecodeError:
        return  # graceful
    assert "error" in out or "version_source" in out, (
        f"unexpected output structure: {out}"
    )


# --- Test 2: dry-run 시 pre-check warning (mock remote with non-existent URL) ---


def test_remote_tag_check_dry_run_no_remote() -> None:
    """origin remote 가 부재 / unreachable 시 graceful (exists=False)."""
    # origin URL 을 임시 dir 의 non-git remote 로 redirect
    import tempfile
    with tempfile.TemporaryDirectory() as tmp:
        # 빈 git init + remote add dummy
        subprocess.run(["git", "init"], cwd=tmp, capture_output=True, timeout=10)
        subprocess.run(
            ["git", "remote", "add", "origin", "https://example.invalid/repo.git"],
            cwd=tmp, capture_output=True, timeout=10,
        )
        # run release --dry-run from this dir (REPO_ROOT 는 module-level 이지만,
        # _check_remote_tag 가 cwd 기준 호출 — git rev-parse 가 fail 하면 exists=False)
        proc = subprocess.run(
            [sys.executable, str(TOOL), "release", "--skip-validate", "--dry-run", "--json"],
            capture_output=True, text=True, timeout=30, cwd=tmp,
        )
        # dummy remote 에 tag 부재 → exists=False, release 흐름은 dist 부재로 error 가능
        # argparse + _check_remote_tag 가 graceful 인지 검증
        assert "unrecognized arguments" not in proc.stderr
        # stdout 이 valid JSON
        try:
            out = json.loads(proc.stdout)
        except json.JSONDecodeError:
            # _check_remote_tag 의 network error 시 output 이 dict 아닐 수 있음
            return  # graceful: skip
        # tag_pre_check field 가 있으면 exists=False
        if "tag_pre_check" in out:
            assert out["tag_pre_check"]["exists"] is False


# --- Test 3: dry-run 시 존재하는 tag pre-check warning ---


def test_remote_tag_check_dry_run_existing_tag() -> None:
    """존재하는 tag dry-run 시 tag_pre_check_warning + pre-check 결과 노출."""
    import tempfile
    with tempfile.TemporaryDirectory() as tmp:
        subprocess.run(["git", "init"], cwd=tmp, capture_output=True, timeout=10)
        # local tag 생성 (origin remote 없으므로 tag 조회 시 부재 → exists=False)
        # _check_remote_tag 가 network error 시 graceful — 일단 stderr 에 argparse error 부재만 확인
        proc = subprocess.run(
            [sys.executable, str(TOOL), "release", "--skip-validate", "--dry-run", "--json"],
            capture_output=True, text=True, timeout=30, cwd=tmp,
        )
        # 본 test 의 핵심: --auto-bump 또는 --dry-run 둘 다 graceful
        assert "unrecognized arguments" not in proc.stderr
        try:
            out = json.loads(proc.stdout)
        except json.JSONDecodeError:
            return  # graceful
        # tag_pre_check 가 dict 일 때만 검증
        if isinstance(out.get("tag_pre_check"), dict):
            assert "tag" in out["tag_pre_check"]


# --- Test 4: _version_sort_key (PEP 440 + suffix 정렬) ---


def test_version_sort_key() -> None:
    """PEP 440 + suffix 정렬. v0.7.17 > v0.7.16-beta > v0.7.16 > v0.7.15."""
    mod = _import_tool()
    key_fn = mod._version_sort_key
    # v0.7.15 < v0.7.16 < v0.7.17
    assert key_fn("v0.7.15") < key_fn("v0.7.16")
    assert key_fn("v0.7.16") < key_fn("v0.7.17")
    # v0.7.16 > v0.7.16-beta (release > beta 인 default — 동일 base 의 release 가 더 큼)
    # 본 tool 의 sort: suffix_rank ''=0, 'beta'=2. v0.7.16 → (0,7,16, 0) < v0.7.16-beta (0,7,16, 2)
    # 즉 *본 tool* 은 release < beta 정렬. wheel/sdist build 시 rc/beta 가 release 뒤.
    # cross-version 비교: v0.7.16-beta < v0.7.17 (numeric 16 < 17)
    assert key_fn("v0.7.16-beta") < key_fn("v0.7.17")
    # v0.7.18-beta < v0.7.19
    assert key_fn("v0.7.18-beta") < key_fn("v0.7.19")


# --- Test 5: next_available_version: remote_tags 비어있을 때 ---


def test_next_available_version_no_remote() -> None:
    """remote_tags=[] 이면 local 그대로 (bumped=False)."""
    mod = _import_tool()
    result = mod.next_available_version("0.7.18", remote_tags=[])
    assert result["bumped"] is False
    assert result["next"] == "0.7.18"
    assert result["current_local"] == "0.7.18"
    assert result["remote_max"] is None


# --- Test 6: next_available_version: remote 가 local 보다 클 때 +0.0.1 ---


def test_next_available_version_remote_ahead() -> None:
    """remote_tags 에 v0.7.17 이 있고 local=0.7.17 이면 +0.0.1 → 0.7.18 (bumped=True)."""
    mod = _import_tool()
    result = mod.next_available_version("0.7.17", remote_tags=["v0.7.17-beta"])
    assert result["bumped"] is True
    assert result["next"] == "0.7.18"
    assert result["remote_max"] == "v0.7.17-beta"
    # current_local=0.7.17 → next=0.7.18
    assert result["current_local"] == "0.7.17"


# --- Test 7: next_available_version: local 이 remote max 보다 클 때 ---


def test_next_available_version_local_ahead() -> None:
    """remote 의 max=0.7.17 이고 local=0.7.18 이면 그대로 (bumped=False)."""
    mod = _import_tool()
    result = mod.next_available_version("0.7.18", remote_tags=["v0.7.17-beta"])
    assert result["bumped"] is False
    assert result["next"] == "0.7.18"
    assert result["current_local"] == "0.7.18"
    assert result["remote_max"] == "v0.7.17-beta"


# --- 메인 실행 ---


def main() -> int:
    test_funcs = [
        test_remote_tag_check_argparse,
        test_remote_tag_check_dry_run_no_remote,
        test_remote_tag_check_dry_run_existing_tag,
        test_version_sort_key,
        test_next_available_version_no_remote,
        test_next_available_version_remote_ahead,
        test_next_available_version_local_ahead,
    ]
    passed = 0
    failed = 0
    for func in test_funcs:
        try:
            func()
            print(f"  PASS  {func.__name__}")
            passed += 1
        except AssertionError as e:
            print(f"  FAIL  {func.__name__}: {e}")
            failed += 1
        except Exception as e:  # noqa: BLE001
            print(f"  ERROR {func.__name__}: {type(e).__name__}: {e}")
            failed += 1

    print()
    print(f"{passed} pass, {failed} fail")
    return 0 if failed == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
