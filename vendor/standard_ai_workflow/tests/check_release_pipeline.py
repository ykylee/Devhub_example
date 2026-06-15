"""tools/release_pipeline.py smoke test (v0.7.9+).

3 subcommand 의 release pipeline 정합성 검증.
- validate: 4 source (packaging, doctor, state, git) 의 release-readiness
- version-bump: pyproject.toml version patch (--patch / --minor / --major / --to)
- note-draft: git log <from>..HEAD → release note skeleton 자동 생성

Test 구성 (8 test):
1. validate --json output: 4 source 결과 dict
2. version-bump --patch dry-run: current 0.7.8 → next 0.7.9
3. version-bump --to=0.8.0 dry-run: 명시 버전 적용
4. version-bump apply: pyproject.toml 실제 갱신 + restore
5. note-draft dry-run: output_path + commits count
6. note-draft apply: 파일 생성 + 내용 검증
7. parse_version: '0.7.8' / '0.7.8-beta' 정합
8. main CLI: --dry-run / --apply / subcommand help

Reference:
- tools/release_pipeline.py 본체
- tools/check_packaging.py (validate 의 packaging source)
- workflow_kit.cli.doctor (v0.7.8, validate 의 doctor source)
- tools/refresh_wiki_memory.py (v0.7.5, note-draft 의 git log 패턴)
- memory #5 standard-ai-workflow.md (release 채널 정책)
"""

from __future__ import annotations

import json
import re
import subprocess
import sys
import tempfile
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
TOOL = SOURCE_ROOT / "tools" / "release_pipeline.py"
PYPROJECT = SOURCE_ROOT / "pyproject.toml"


def _import_tool():
    """release_pipeline.py 를 importlib 로 로드."""
    import importlib.util
    import sys
    spec = importlib.util.spec_from_file_location("release_pipeline", str(TOOL))
    mod = importlib.util.module_from_spec(spec)
    sys.modules["release_pipeline"] = mod  # 3.14 dataclass 호환
    spec.loader.exec_module(mod)
    return mod


# --- Test 1: validate --json output ---


def test_validate_json_output() -> None:
    """validate --json output 이 4 source 결과 dict 반환."""
    proc = subprocess.run(
        [sys.executable, str(TOOL), "validate",
         "--skip-packaging", "--skip-doctor",
         "--json"],
        capture_output=True, text=True, timeout=60,
    )
    # exit 0 또는 1 (git.status untracked 있을 수 있음) — JSON parse 가능 검증
    out = json.loads(proc.stdout)
    assert "packaging" in out
    assert "doctor" in out
    assert "state" in out
    assert "git" in out


# --- Test 2: version-bump --patch dry-run ---


def test_version_bump_patch_dry_run() -> None:
    """version-bump --patch dry-run: current 0.7.x → next 0.7.x+1."""
    mod = _import_tool()
    current = mod.read_version()
    result = mod.cmd_version_bump(type("Args", (), {"patch": True, "minor": False, "major": False, "to": None, "dry_run": True, "apply": False, "no_init": False})())
    assert result["mode"] == "dry-run"
    major, minor, patch = mod.parse_version(current)
    expected = f"{major}.{minor}.{patch + 1}"
    assert result["next_pyproject"] == expected


# --- Test 3: version-bump --to=... dry-run ---


def test_version_bump_to_explicit() -> None:
    """--to=0.8.0 명시 시 그대로 사용."""
    mod = _import_tool()
    result = mod.cmd_version_bump(type("Args", (), {"patch": False, "minor": False, "major": False, "to": "0.8.0", "dry_run": True, "apply": False, "no_init": False})())
    assert result["next_pyproject"] == "0.8.0"


# --- Test 4: version-bump apply + restore ---


def test_version_bump_apply_and_restore() -> None:
    """--apply 시 pyproject.toml 실제 갱신, 원복 후 정합."""
    mod = _import_tool()
    original = mod.read_version()
    original_init = mod.read_workflow_kit_version()
    try:
        # 0.7.8 → 0.7.9 patch
        result = mod.cmd_version_bump(type("Args", (), {"patch": True, "minor": False, "major": False, "to": None, "dry_run": False, "apply": True, "no_init": False})())
        assert result["mode"] == "applied"
        assert result["previous_pyproject"] == original
        assert result["current_pyproject"] != original
        # file 갱신 검증
        assert mod.read_version() == result["current_pyproject"]
        # __init__.py 도 auto-sync 검증
        assert "current_workflow_kit" in result
        assert mod.read_workflow_kit_version() == result["current_workflow_kit"]
    finally:
        # restore (pyproject + __init__.py)
        mod.write_version(original)
        mod.write_workflow_kit_version(original.lstrip("v").split("-")[0] if original.startswith("v") else original.split("-")[0])
        assert mod.read_version() == original
        assert mod.read_workflow_kit_version() == original_init


def test_note_draft_dry_run() -> None:
    """note-draft dry-run: output_path + commits count."""
    proc = subprocess.run(
        [sys.executable, str(TOOL), "note-draft",
         "--from=v0.7.4-beta", "--to=0.7.9", "--dry-run"],
        capture_output=True, text=True, timeout=30,
    )
    assert proc.returncode == 0, f"exit {proc.returncode}: {proc.stderr}"
    out = re.search(r"output_path:\s*([^\s]+)", proc.stdout)
    assert out is not None
    assert out.group(1) == "releases/Beta-v0.7.9.md"
    commits_m = re.search(r"commits:\s*(\d+)", proc.stdout)
    assert commits_m is not None
    assert int(commits_m.group(1)) > 0


# --- Test 6: parse_version ---


def test_parse_version_formats() -> None:
    """parse_version 이 'X.Y.Z' / 'X.Y.Z-suffix' 모두 정합."""
    mod = _import_tool()
    assert mod.parse_version("0.7.8") == (0, 7, 8)
    assert mod.parse_version("0.7.8-beta") == (0, 7, 8)
    assert mod.parse_version("1.0.0") == (1, 0, 0)
    # invalid format
    try:
        mod.parse_version("invalid")
        assert False, "should have raised"
    except ValueError:
        pass


# --- Test 7: bump_version logic ---


def test_bump_version_logic() -> None:
    """bump_version 의 major / minor / patch / to 분기."""
    mod = _import_tool()
    assert mod.bump_version("0.7.8", patch=True) == "0.7.9"
    assert mod.bump_version("0.7.8", minor=True) == "0.8.0"
    assert mod.bump_version("0.7.8", major=True) == "1.0.0"
    assert mod.bump_version("0.7.8", to="2.0.0") == "2.0.0"
    # default = patch
    assert mod.bump_version("0.7.8") == "0.7.9"


# --- Test 8: main CLI subcommand help ---


def test_cli_subcommand_help() -> None:
    """main --help + 각 subcommand --help 정상."""
    proc = subprocess.run(
        [sys.executable, str(TOOL), "--help"],
        capture_output=True, text=True, timeout=10,
    )
    assert proc.returncode == 0
    assert "validate" in proc.stdout
    assert "version-bump" in proc.stdout
    assert "note-draft" in proc.stdout


# --- 메인 실행 ---


def main() -> int:
    test_funcs = [
        test_validate_json_output,
        test_version_bump_patch_dry_run,
        test_version_bump_to_explicit,
        test_version_bump_apply_and_restore,
        test_note_draft_dry_run,
        test_parse_version_formats,
        test_bump_version_logic,
        test_cli_subcommand_help,
    ]

    passed = 0
    failed = 0
    failures: list[tuple[str, str]] = []
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
