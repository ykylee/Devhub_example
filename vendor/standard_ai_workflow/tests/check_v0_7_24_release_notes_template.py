#!/usr/bin/env python3
"""v0.7.24+: cmd_release --notes-template flag smoke test.

`Beta-v<X>.<Y>.<Z>.md` 의 default 외에, simple (1-line summary) / changelog
(Keep-a-Changelog 1.1.0) / custom:<path> 4가지 template 지원. GH release notes
format 자유도.

Test 구성 (5 test):
1. test_notes_template_default_argparse: --notes-template=default argparse error 부재
2. test_notes_template_simple: simple template 의 1-line summary 자동 generate
3. test_notes_template_changelog: changelog template 가 CHANGELOG.md 가리킴
4. test_notes_template_custom: custom:<path> 의 임의 path 지원
5. test_notes_template_unknown: unknown value 시 명확한 error message

Reference:
- workflow-source/tools/release_pipeline.py (v0.7.24 본 release, --notes-template + _resolve_notes_file)
- v0.7.14 release note (changelog-gen subcommand, 본 release 의 'changelog' template 의 1차 출처)
"""

from __future__ import annotations

import importlib.util
import sys
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
TOOL = SOURCE_ROOT / "tools" / "release_pipeline.py"


def _import_tool():
    spec = importlib.util.spec_from_file_location("release_pipeline", str(TOOL))
    mod = importlib.util.module_from_spec(spec)
    sys.modules["release_pipeline"] = mod
    spec.loader.exec_module(mod)
    return mod


# --- Test 1: --notes-template argparse error 부재 ---


def test_notes_template_default_argparse() -> None:
    """release --notes-template=default 가 argparse error 없이 받아들여짐."""
    import subprocess
    proc = subprocess.run(
        [sys.executable, str(TOOL), "release", "--notes-template=default", "--skip-validate", "--dry-run", "--json"],
        capture_output=True, text=True, timeout=30,
    )
    assert "unrecognized arguments" not in proc.stderr, f"argparse error: {proc.stderr}"
    # output 이 valid JSON (dist 부재 error 가능)
    import json
    try:
        out = json.loads(proc.stdout)
    except json.JSONDecodeError:
        return  # graceful
    assert "error" in out or "notes_template" in out or "version_source" in out


# --- Test 2: simple template 자동 generate ---


def test_notes_template_simple() -> None:
    """simple template 가 default notes 의 1st paragraph 자동 generate."""
    mod = _import_tool()
    # 테스트용 default notes 생성
    test_version = "9.9.9-test"
    test_notes = mod.RELEASES_DIR / f"Beta-v{test_version}.md"
    test_notes.parent.mkdir(parents=True, exist_ok=True)
    try:
        test_notes.write_text(
            "# Beta v9.9.9-test — Test Release\n\n"
            "## 핵심 추가\n\n"
            "이것은 본 release 의 *1st paragraph* 입니다.\n\n"
            "## 다음 섹션\n\n"
            "이건 안 포함.\n"
        )
        # simple template 호출
        result = mod._resolve_notes_file(test_version, "simple")
        assert "error" not in result or result.get("error") is None
        notes_file = result["notes_file"]
        assert result["source"] == "simple"
        # simple file 이 default notes 의 *1st # + 1st ## + 1st paragraph* 포함
        if notes_file.exists():
            content = notes_file.read_text(encoding="utf-8")
            # 1st # 헤더 + 1st ## 헤더 + 1st paragraph 본문 포함
            assert "Beta v9.9.9-test" in content
            assert "핵심 추가" in content
            assert "1st paragraph" in content
            # 2nd ## 헤더 ("다음 섹션") 와 그 paragraph 는 포함 안 됨
            assert "다음 섹션" not in content
            assert "이건 안 포함" not in content
    finally:
        # cleanup
        if test_notes.exists():
            test_notes.unlink()
        simple_file = mod.RELEASES_DIR / f"Beta-v{test_version}-simple.md"
        if simple_file.exists():
            simple_file.unlink()


# --- Test 3: changelog template ---


def test_notes_template_changelog() -> None:
    """changelog template 가 workflow-source/CHANGELOG.md 가리킴."""
    mod = _import_tool()
    result = mod._resolve_notes_file("0.7.24", "changelog")
    assert result.get("error") is None
    assert result["source"] == "changelog"
    notes_file = result["notes_file"]
    assert notes_file.name == "CHANGELOG.md"
    assert "workflow-source" in str(notes_file)


# --- Test 4: custom:<path> ---


def test_notes_template_custom() -> None:
    """custom:<path> 의 임의 path 지원 (in-repo + absolute)."""
    mod = _import_tool()
    # in-repo custom path
    custom_rel = "workflow-source/releases/Beta-v0.7.24.md"
    result = mod._resolve_notes_file("0.7.24", f"custom:{custom_rel}")
    assert result.get("error") is None
    assert result["source"] == f"custom:{custom_rel}"
    notes_file = result["notes_file"]
    assert notes_file.name == "Beta-v0.7.24.md"
    assert notes_file.is_absolute(), (
        f"custom path 가 absolute 가 아님: {notes_file}"
    )

    # absolute path
    abs_path = "/tmp/test-notes.md"
    result_abs = mod._resolve_notes_file("0.7.24", f"custom:{abs_path}")
    assert result_abs.get("error") is None
    assert str(result_abs["notes_file"]) == abs_path


# --- Test 5: unknown template ---


def test_notes_template_unknown() -> None:
    """unknown --notes-template value 시 명확한 error message."""
    mod = _import_tool()
    result = mod._resolve_notes_file("0.7.24", "invalid-template-name")
    assert result.get("error") is not None
    assert "unknown" in result["error"]
    assert "--notes-template" in result["error"]
    # 사용 가능한 option 들이 error message 에 명시
    for opt in ("default", "detailed", "simple", "changelog", "custom"):
        assert opt in result["error"], f"{opt} not in error message: {result['error']}"


# --- 메인 실행 ---


def main() -> int:
    test_funcs = [
        test_notes_template_default_argparse,
        test_notes_template_simple,
        test_notes_template_changelog,
        test_notes_template_custom,
        test_notes_template_unknown,
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
