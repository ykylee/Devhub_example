"""tools/release_pipeline.py cmd_version_bump auto-sync __version__ smoke test (v0.7.14+).

v0.7.14 follow-up: cmd_version_bump 가 pyproject.toml + workflow_kit/__init__.py 둘 다 자동 patch.
v0.7.11~v0.7.13 release 동안 manual sync 누락 패턴 해소.

4 test PASS 기준.

Test list:
1. test_version_bump_argparse: --no-init / --dry-run / --json argparse error 없음
2. test_version_bump_dry_run_no_change: dry-run mode 에서 file 변경 없음
3. test_version_bump_apply_sync_both: apply mode 에서 pyproject + __init__.py 둘 다 갱신 (dynamic pre/current)
4. test_version_bump_no_init_skips: --no-init 시 pyproject 만 갱신 + __init__.py 보존
"""

from __future__ import annotations

import json
import re
import subprocess
import sys
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
TOOL = SOURCE_ROOT / "tools" / "release_pipeline.py"
PYPROJECT = SOURCE_ROOT / "pyproject.toml"
WORKFLOW_KIT_INIT = SOURCE_ROOT / "workflow_kit" / "__init__.py"


def _read_pyproject_version() -> str:
    """pyproject.toml [project] version 읽기."""
    text = PYPROJECT.read_text()
    m = re.search(r'version\s*=\s*"([^"]+)"', text)
    assert m, f"version not found in {PYPROJECT}"
    return m.group(1)


def _read_init_version() -> str:
    """__init__.py __version__ 읽기."""
    text = WORKFLOW_KIT_INIT.read_text()
    m = re.search(r'__version__\s*=\s*"([^"]+)"', text)
    assert m, f"__version__ not found in {WORKFLOW_KIT_INIT}"
    return m.group(1)


def _bump_patch(version: str) -> str:
    """version 'X.Y.Z' → 'X.Y.(Z+1)'."""
    parts = version.split(".")
    if len(parts) != 3:
        raise ValueError(f"invalid version format: {version}")
    return f"{parts[0]}.{parts[1]}.{int(parts[2]) + 1}"


# --- Test 1: argparse 인식 ---


def test_version_bump_argparse() -> None:
    """version-bump 의 --no-init / --dry-run / --json argparse error 없음."""
    for args in [
        ["--to=0.7.99", "--dry-run", "--json"],
        ["--to=0.7.99", "--no-init", "--dry-run"],
        ["--patch", "--dry-run", "--json"],
        ["--minor", "--dry-run"],
    ]:
        proc = subprocess.run(
            [sys.executable, str(TOOL), "version-bump"] + args,
            capture_output=True, text=True, timeout=30,
        )
        assert "unrecognized arguments" not in proc.stderr, \
            f"args={args} → argparse error: {proc.stderr}"


# --- Test 2: dry-run mode 에서 file 변경 없음 ---


def test_version_bump_dry_run_no_change() -> None:
    """dry-run mode 에서 pyproject + __init__.py 모두 변경 없음."""
    pre_py = _read_pyproject_version()
    pre_init = _read_init_version()

    proc = subprocess.run(
        [sys.executable, str(TOOL), "version-bump", "--to=0.7.99", "--dry-run", "--json"],
        capture_output=True, text=True, timeout=30,
    )
    assert proc.returncode == 0, f"exit {proc.returncode}: {proc.stderr}"
    out = json.loads(proc.stdout)
    assert out["mode"] == "dry-run"
    assert out["next_pyproject"] == "0.7.99"

    # post-state 동일
    post_py = _read_pyproject_version()
    post_init = _read_init_version()
    assert pre_py == post_py, f"pyproject 변경됨: {pre_py} → {post_py}"
    assert pre_init == post_init, f"__init__.py 변경됨: {pre_init} → {post_init}"


# --- Test 3: apply mode 에서 pyproject + __init__.py 둘 다 갱신 ---


def test_version_bump_apply_sync_both() -> None:
    """apply mode 에서 pyproject + __init__.py 둘 다 patch. dynamic pre/current (release 마다 갱신)."""
    pre_py = _read_pyproject_version()
    pre_init = _read_init_version()
    target = _bump_patch(pre_py)
    target_init = f"v{target}-beta"

    # apply
    proc = subprocess.run(
        [sys.executable, str(TOOL), "version-bump", "--to", target, "--apply", "--json"],
        capture_output=True, text=True, timeout=30,
    )
    assert proc.returncode == 0, f"exit {proc.returncode}: {proc.stderr}"
    out = json.loads(proc.stdout)
    assert out["mode"] == "applied"
    assert out["previous_pyproject"] == pre_py
    assert out["current_pyproject"] == target
    assert out["previous_workflow_kit"] == pre_init
    assert out["current_workflow_kit"] == target_init

    # post-state
    post_py = _read_pyproject_version()
    post_init = _read_init_version()
    assert post_py == target, f"pyproject not updated: {post_py}"
    assert post_init == target_init, f"__init__ not updated: {post_init}"

    # restore
    proc2 = subprocess.run(
        [sys.executable, str(TOOL), "version-bump", "--to", pre_py, "--apply"],
        capture_output=True, text=True, timeout=30,
    )
    assert proc2.returncode == 0, f"restore exit {proc2.returncode}: {proc2.stderr}"
    assert _read_pyproject_version() == pre_py
    assert _read_init_version() == pre_init


# --- Test 4: --no-init 시 pyproject 만 갱신, __init__.py 보존 ---


def test_version_bump_no_init_skips() -> None:
    """--no-init 시 pyproject 만 patch, __init__.py 보존. dynamic pre (release 마다 갱신)."""
    pre_py = _read_pyproject_version()
    pre_init = _read_init_version()
    target = _bump_patch(pre_py)

    proc = subprocess.run(
        [sys.executable, str(TOOL), "version-bump", "--to", target, "--no-init", "--apply"],
        capture_output=True, text=True, timeout=30,
    )
    assert proc.returncode == 0, f"exit {proc.returncode}: {proc.stderr}"

    # pyproject 갱신, __init__ 보존
    post_py = _read_pyproject_version()
    post_init = _read_init_version()
    assert post_py == target, f"pyproject not updated: {post_py}"
    assert post_init == pre_init, f"__init__ 잘못 갱신됨: {post_init}"

    # restore
    subprocess.run(
        [sys.executable, str(TOOL), "version-bump", "--to", pre_py, "--apply"],
        capture_output=True, text=True, timeout=30,
        check=True,
    )
    assert _read_pyproject_version() == pre_py


# --- 메인 실행 ---


def main() -> int:
    test_funcs = [
        test_version_bump_argparse,
        test_version_bump_dry_run_no_change,
        test_version_bump_apply_sync_both,
        test_version_bump_no_init_skips,
    ]

    failed: list[str] = []
    for fn in test_funcs:
        name = fn.__name__
        try:
            fn()
            print(f"  PASS  {name}")
        except Exception as e:
            print(f"  FAIL  {name}: {type(e).__name__}: {e}")
            failed.append(name)

    total = len(test_funcs)
    passed = total - len(failed)
    print(f"\n{passed}/{total} tests passed.")
    if failed:
        print(f"\n{len(failed)} tests failed:")
        for name in failed:
            print(f"  - {name}")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
