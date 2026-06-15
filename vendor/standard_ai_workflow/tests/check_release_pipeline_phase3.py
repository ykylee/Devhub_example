"""tools/release_pipeline.py Phase 3 subcommand (dist) smoke test.

v0.7.11: dist subcommand (`python3 -m build` wheel + sdist 자동 빌드).
8 test PASS 기준.

Test list:
1. test_dist_dry_run_no_build_module: build module 부재 시 graceful error + exit 1
2. test_dist_dry_run_with_build_module: build module 가용 시 command plan 출력
3. test_dist_apply_with_build_module: build module 가용 + apply → wheel + sdist glob
4. test_dist_skip_existing: --skip-existing + 기존 dist 파일 존재 시 mode=skip
5. test_build_command_sdist_wheel_only: --sdist-only / --wheel-only flag 반영
6. test_expected_dist_pattern_pep440: PEP 440 normalize (0.7.10-beta → 0.7.10)
7. test_dist_subcommand_help: main --help + 7 subcommand 노출
8. test_dist_skip_existing_argparse: --skip-existing argparse error 없음
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
    spec = importlib.util.spec_from_file_location("release_pipeline", TOOL)
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def _has_build_module() -> bool:
    """`build` module 가용 여부."""
    try:
        import build  # type: ignore[import-not-found]  # noqa: F401

        return True
    except ImportError:
        return False


# --- Test 1: dist --dry-run (build module 부재 시 graceful error) ---


def test_dist_dry_run_no_build_module() -> None:
    """dist --dry-run — build module 부재 시 graceful error + exit 1.

    이 환경 (system python, build module 부재) 에서 실행 시 exit 1 + error + hint 검증.
    build module 가용 환경에서는 skip (test 2 가 검증).
    """
    if _has_build_module():
        return

    proc = subprocess.run(
        [sys.executable, str(TOOL), "dist", "--dry-run", "--json"],
        capture_output=True, text=True, timeout=30,
    )
    assert proc.returncode == 1, f"expected exit 1, got {proc.returncode}: {proc.stderr}"
    out = json.loads(proc.stdout)
    assert "error" in out
    assert "build module not installed" in out["error"]
    assert out["build_module"]["available"] is False
    assert "pip install build" in out["build_module"]["hint"]


# --- Test 2: dist --dry-run (build module 가용 시 command plan) ---


def test_dist_dry_run_with_build_module() -> None:
    """dist --dry-run — build module 가용 시 command + version + expected_pattern 출력.

    build module 부재 환경에서는 skip (이 환경 test 1 이 graceful error 검증).
    """
    if not _has_build_module():
        return

    proc = subprocess.run(
        [sys.executable, str(TOOL), "dist", "--dry-run", "--json"],
        capture_output=True, text=True, timeout=30,
    )
    assert proc.returncode == 0, f"expected exit 0, got {proc.returncode}: {proc.stderr}"
    out = json.loads(proc.stdout)
    assert out["mode"] == "dry-run"
    assert out["build_module"]["available"] is True
    assert "version" in out
    assert "command" in out
    assert "expected_pattern" in out
    assert out["expected_pattern"].startswith("standard_ai_workflow-")


# --- Test 3: dist --apply (build module 가용 시 wheel + sdist glob) ---


def test_dist_apply_with_build_module() -> None:
    """dist --apply — build module 가용 + apply → wheel + sdist glob 결과 report.

    build module 부재 시 skip (apply 불가). 가용 시에만 dist/ 의 wheel + sdist 검증.
    """
    if not _has_build_module():
        return

    proc = subprocess.run(
        [sys.executable, str(TOOL), "dist", "--apply", "--timeout=120", "--json"],
        capture_output=True, text=True, timeout=180,
        cwd=str(SOURCE_ROOT),
    )
    assert proc.returncode == 0, f"expected exit 0, got {proc.returncode}: {proc.stderr}"
    out = json.loads(proc.stdout)
    assert out["mode"] == "apply"
    assert out["ok"] is True
    assert "built" in out
    built = out["built"]
    assert any(f.endswith(".whl") for f in built), f"no .whl in built: {built}"
    assert any(f.endswith(".tar.gz") for f in built), f"no .tar.gz in built: {built}"


# --- Test 4: dist --skip-existing (build module 가용 시 기존 파일 skip) ---


def test_dist_skip_existing() -> None:
    """dist --skip-existing + dist/ 의 current-version 파일 존재 시 mode=skip.

    build module 부재 시 skip. 가용 시 mode=skip 또는 mode=dry-run (기존 파일 없으면).
    """
    if not _has_build_module():
        return

    proc = subprocess.run(
        [sys.executable, str(TOOL), "dist", "--skip-existing", "--dry-run", "--json"],
        capture_output=True, text=True, timeout=30,
        cwd=str(SOURCE_ROOT),
    )
    assert proc.returncode == 0
    out = json.loads(proc.stdout)
    # mode 는 skip 또는 dry-run (기존 파일이 있는 경우 skip, 없으면 dry-run)
    assert out["mode"] in ("skip", "dry-run")
    if out["mode"] == "skip":
        assert out["skipped"] is True
        assert "existing" in out
        assert all(f.startswith("standard_ai_workflow-") for f in out["existing"])


# --- Test 5: _build_command (sdist-only / wheel-only) ---


def test_build_command_sdist_wheel_only() -> None:
    """_build_command 가 --sdist-only / --wheel-only flag 반영."""
    mod = _import_tool()
    out_dir = Path("/tmp/dist_test")

    cmd_default = mod._build_command(out_dir)
    assert "--sdist" not in cmd_default
    assert "--wheel" not in cmd_default
    assert "--outdir" in cmd_default
    assert str(out_dir) in cmd_default

    cmd_sdist = mod._build_command(out_dir, sdist_only=True)
    assert "--sdist" in cmd_sdist
    assert "--wheel" not in cmd_sdist

    cmd_wheel = mod._build_command(out_dir, wheel_only=True)
    assert "--wheel" in cmd_wheel
    assert "--sdist" not in cmd_wheel


# --- Test 6: _expected_dist_pattern (PEP 440 normalize) ---


def test_expected_dist_pattern_pep440() -> None:
    """_expected_dist_pattern 가 PEP 440 suffix 처리.

    '0.7.10' / '0.7.10-beta' 모두 '0.7.10' (split('-')[0]) 으로 normalize.
    """
    mod = _import_tool()
    assert mod._expected_dist_pattern("0.7.10") == "0.7.10"
    assert mod._expected_dist_pattern("0.7.10-beta") == "0.7.10"
    assert mod._expected_dist_pattern("1.0.0-rc1") == "1.0.0"


# --- Test 7: main --help (dist subcommand 노출) ---


def test_dist_subcommand_help() -> None:
    """main --help + 7 subcommand (validate / version-bump / note-draft / release / verify / rollback / dist) 노출."""
    proc = subprocess.run(
        [sys.executable, str(TOOL), "--help"],
        capture_output=True, text=True, timeout=10,
    )
    assert proc.returncode == 0, f"expected exit 0, got {proc.returncode}: {proc.stderr}"
    for cmd in ("validate", "version-bump", "note-draft", "release", "verify", "rollback", "dist"):
        assert cmd in proc.stdout, f"missing subcommand {cmd!r} in help"


# --- Test 8: dist --skip-existing argparse 검증 ---


def test_dist_skip_existing_argparse() -> None:
    """dist --skip-existing 가 argparse error 없이 받아들여짐."""
    proc = subprocess.run(
        [sys.executable, str(TOOL), "dist", "--skip-existing", "--dry-run", "--json"],
        capture_output=True, text=True, timeout=30,
    )
    # build module 부재 시 exit 1 이지만 argparse error 는 없어야 함
    assert "unrecognized arguments" not in proc.stderr
    assert "unrecognized arguments" not in proc.stdout
    # 정상 실행 결과 (성공 또는 build module 부재 error)
    out = json.loads(proc.stdout)
    assert "mode" in out


# --- 메인 실행 ---


def main() -> int:
    test_funcs = [
        test_dist_dry_run_no_build_module,
        test_dist_dry_run_with_build_module,
        test_dist_apply_with_build_module,
        test_dist_skip_existing,
        test_build_command_sdist_wheel_only,
        test_expected_dist_pattern_pep440,
        test_dist_subcommand_help,
        test_dist_skip_existing_argparse,
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
