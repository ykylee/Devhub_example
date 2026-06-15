"""tools/release_pipeline.py release --version flag smoke test (v0.7.13+).

v0.7.13 follow-up: `cmd_release --version=<X.Y.Z>` flag 추가.
backfill 시 staging 용도 (pyproject.toml 일시 patch 불필요).
3 test PASS 기준.

Test list:
1. test_version_argparse_recognized: --version=<X.Y.Z> argparse error 없음 + version_source=cli-flag
2. test_version_override_pyproject: --version=0.7.5 일 때 tag=v0.7.5-beta + notes_file=Beta-v0.7.5.md (staging area 가용 시)
3. test_version_default_pyproject: --version 미지정 시 version_source=pyproject.toml (default) + tag=v{HEAD}-beta 동적
"""

from __future__ import annotations

import importlib.util
import json
import subprocess
import sys
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
TOOL = SOURCE_ROOT / "tools" / "release_pipeline.py"


# --- Test 1: --version argparse 인식 + cli-flag source ---


def test_version_argparse_recognized() -> None:
    """release --version=<X.Y.Z> 가 argparse error 없이 받아들여짐 + version_source=cli-flag."""
    proc = subprocess.run(
        [sys.executable, str(TOOL), "release", "--version=0.7.5", "--skip-validate", "--dry-run", "--json"],
        capture_output=True, text=True, timeout=30,
    )
    # argparse 가 --version 을 인식하면 stderr 에 'unrecognized arguments' 없어야 함
    assert "unrecognized arguments" not in proc.stderr, f"argparse error: {proc.stderr}"
    out = json.loads(proc.stdout)
    # version_source: cli-flag (override 적용)
    assert out["version_source"] == "cli-flag", \
        f"expected cli-flag, got {out.get('version_source')}"
    # tag 도 override 적용 (dist 부재 error 시 tag 부재 가능)
    if "tag" in out:
        assert out["tag"] == "v0.7.5-beta", f"expected v0.7.5-beta, got {out['tag']}"
    else:
        # dist 부재 → error 에서 version 0.7.5 가 반영
        assert "0.7.5" in out.get("error", ""), f"expected version 0.7.5 in error, got {out.get('error')}"


# --- Test 2: --version=0.7.5 의 tag / notes_file / assets 검증 ---


def test_version_override_pyproject() -> None:
    """--version=0.7.5 일 때 tag=v0.7.5-beta + notes_file=Beta-v0.7.5.md.

    staging area (/tmp/dist_v_0.7.5/) 가 없으면 dry-run 결과 (version_source=cli-flag) 만 검증.
    """
    import shutil

    src = Path("/tmp/dist_v_0.7.5")
    whl = src / "standard_ai_workflow-0.7.5-py3-none-any.whl"
    tar = src / "standard_ai_workflow-0.7.5.tar.gz"

    has_staging = whl.exists() and tar.exists()

    if has_staging:
        # v0.7.11 dist 보존: pre-stage 에서 보장 (있으면 backup, 없으면 staging area 에서 copy)
        v11_whl = SOURCE_ROOT / "dist" / "standard_ai_workflow-0.7.11-py3-none-any.whl"
        v11_tar = SOURCE_ROOT / "dist" / "standard_ai_workflow-0.7.11.tar.gz"
        v11_staging = Path("/tmp/dist_v0_7_11")
        backup_dir = Path("/tmp/dist_v0_7_11_version_flag_test")
        backup_dir.mkdir(parents=True, exist_ok=True)

        # pre-stage: v0.7.11 dist 보장
        if not v11_whl.exists() and (v11_staging / v11_whl.name).exists():
            shutil.copy(v11_staging / v11_whl.name, v11_whl)
        if not v11_tar.exists() and (v11_staging / v11_tar.name).exists():
            shutil.copy(v11_staging / v11_tar.name, v11_tar)

        # backup (있을 때만)
        if v11_whl.exists():
            shutil.copy(v11_whl, backup_dir / v11_whl.name)
        if v11_tar.exists():
            shutil.copy(v11_tar, backup_dir / v11_tar.name)

        # 0.7.5 staging
        shutil.copy(whl, SOURCE_ROOT / "dist" / whl.name)
        shutil.copy(tar, SOURCE_ROOT / "dist" / tar.name)

    try:
        proc = subprocess.run(
            [sys.executable, str(TOOL), "release", "--version=0.7.5", "--skip-validate", "--dry-run", "--json"],
            capture_output=True, text=True, timeout=30,
        )
        out = json.loads(proc.stdout)
        assert out["version_source"] == "cli-flag", \
            f"expected cli-flag, got {out.get('version_source')}"
        if has_staging:
            assert out["tag"] == "v0.7.5-beta", f"expected v0.7.5-beta, got {out.get('tag')}"
            assert out["notes_file"].endswith("Beta-v0.7.5.md"), \
                f"unexpected notes_file: {out.get('notes_file')}"
            assert any("0.7.5" in a for a in out["assets"]), \
                f"no 0.7.5 asset: {out.get('assets')}"
        else:
            # staging 부재 → error return (no dist files). version 만 검증.
            assert "0.7.5" in out.get("error", ""), \
                f"expected 0.7.5 in error, got {out.get('error')}"
    finally:
        if has_staging:
            # staging cleanup
            for f in (whl, tar):
                (SOURCE_ROOT / "dist" / f.name).unlink(missing_ok=True)
            # v0.7.11 dist restore
            backup_dir = Path("/tmp/dist_v0_7_11_version_flag_test")
            v11_whl = SOURCE_ROOT / "dist" / "standard_ai_workflow-0.7.11-py3-none-any.whl"
            v11_tar = SOURCE_ROOT / "dist" / "standard_ai_workflow-0.7.11.tar.gz"
            if (backup_dir / "standard_ai_workflow-0.7.11-py3-none-any.whl").exists():
                shutil.copy(
                    backup_dir / "standard_ai_workflow-0.7.11-py3-none-any.whl",
                    v11_whl,
                )
            if (backup_dir / "standard_ai_workflow-0.7.11.tar.gz").exists():
                shutil.copy(
                    backup_dir / "standard_ai_workflow-0.7.11.tar.gz",
                    v11_tar,
                )


# --- Test 3: --version 미지정 시 default (pyproject.toml) ---


def test_version_default_pyproject() -> None:
    """--version 미지정 시 version_source=pyproject.toml (default)."""
    spec = importlib.util.spec_from_file_location("release_pipeline", str(TOOL))
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    current = mod.read_version()

    proc = subprocess.run(
        [sys.executable, str(TOOL), "release", "--skip-validate", "--dry-run", "--json"],
        capture_output=True, text=True, timeout=30,
    )
    out = json.loads(proc.stdout)
    assert out["version_source"] == "pyproject.toml", \
        f"expected pyproject.toml, got {out.get('version_source')}"
    # default version 은 pyproject 의 current (HEAD 정합, release 마다 갱신)
    assert out["tag"] == f"v{current}-beta", f"expected v{current}-beta, got {out.get('tag')}"


# --- 메인 실행 ---


def main() -> int:
    test_funcs = [
        test_version_argparse_recognized,
        test_version_override_pyproject,
        test_version_default_pyproject,
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
