"""tools/release_pipeline.py changelog-gen subcommand smoke test (v0.7.14+).

v0.7.14 follow-up: cmd_changelog_gen — multi-release git log → CHANGELOG.md (Keep-a-Changelog 형식).
v0.7.15 follow-up: --from-tag/--to-tag range filter 추가.
6 test PASS 기준.

Test list:
1. test_changelog_gen_argparse: --output / --unreleased-label / --dry-run / --json / --from-tag / --to-tag argparse error 없음
2. test_changelog_gen_dry_run: dry-run mode 에서 file 변경 없음 + result 정합
3. test_changelog_gen_apply: apply mode 에서 CHANGELOG.md 작성 + Keep-a-Changelog 형식 검증
4. test_changelog_gen_section_categorization: commit subject prefix → section mapping 검증
5. test_changelog_gen_range_filter (v0.7.15+): --from-tag/--to-tag range scan
6. test_changelog_gen_out_of_range_graceful (v0.7.15+): invalid --from-tag 시 graceful fail
"""

from __future__ import annotations

import importlib.util
import json
import os
import re
import subprocess
import sys
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
TOOL = SOURCE_ROOT / "tools" / "release_pipeline.py"
DEFAULT_OUTPUT = SOURCE_ROOT / "CHANGELOG.md"

# workflow_kit.common.atomic_write import 위해 (v0.7.15+ release_pipeline.py 의 의존)
sys.path.insert(0, str(SOURCE_ROOT))


def _import_tool():
    """release_pipeline.py 를 importlib 로 로드."""
    spec = importlib.util.spec_from_file_location("release_pipeline", str(TOOL))
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


# --- Test 1: argparse 인식 ---


def test_changelog_gen_argparse() -> None:
    """changelog-gen 의 --output / --unreleased-label / --dry-run / --json / --from-tag / --to-tag argparse error 없음."""
    for args in [
        ["--dry-run", "--json"],
        ["--output=/tmp/test_changelog.md", "--dry-run"],
        ["--unreleased-label=Pending", "--dry-run"],
        ["--from-tag=v0.7.0-beta", "--to-tag=v0.7.10-beta", "--dry-run", "--json"],
        ["--from-tag=v0.7.5-beta", "--dry-run"],
    ]:
        proc = subprocess.run(
            [sys.executable, str(TOOL), "changelog-gen"] + args,
            capture_output=True, text=True, timeout=30,
            env={**os.environ, "PYTHONPATH": str(SOURCE_ROOT)},
        )
        assert "unrecognized arguments" not in proc.stderr, \
            f"args={args} → argparse error: {proc.stderr}"


# --- Test 2: dry-run mode 에서 file 변경 없음 + result 정합 ---


def test_changelog_gen_dry_run() -> None:
    """dry-run mode 에서 file 변경 없음 + mode=dry-run + commits/versions 정합."""
    pre_exists = DEFAULT_OUTPUT.exists()
    pre_content = DEFAULT_OUTPUT.read_text() if pre_exists else ""

    proc = subprocess.run(
        [sys.executable, str(TOOL), "changelog-gen", "--dry-run", "--json"],
        capture_output=True, text=True, timeout=60,
        env={**os.environ, "PYTHONPATH": str(SOURCE_ROOT)},
    )
    assert proc.returncode == 0, f"exit {proc.returncode}: {proc.stderr}"
    out = json.loads(proc.stdout)
    assert out["mode"] == "dry-run"
    assert "output_path" in out
    assert out["commits"] > 0
    assert out["versions"] > 0

    # file 변경 없음
    post_exists = DEFAULT_OUTPUT.exists()
    assert pre_exists == post_exists, "dry-run 에서 file 부재 변경"
    if pre_exists:
        assert DEFAULT_OUTPUT.read_text() == pre_content, "dry-run 에서 file 내용 변경"


# --- Test 3: apply mode 에서 CHANGELOG.md 작성 + Keep-a-Changelog 형식 ---


def test_changelog_gen_apply() -> None:
    """apply mode 에서 CHANGELOG.md 작성 + Keep-a-Changelog 형식 검증.

    임시 output file 사용. 본 repo 의 CHANGELOG.md 는 건드리지 않음.
    """
    import tempfile
    with tempfile.TemporaryDirectory() as tmpdir:
        out_path = Path(tmpdir) / "CHANGELOG.md"
        proc = subprocess.run(
            [sys.executable, str(TOOL), "changelog-gen", "--output", str(out_path), "--apply", "--json"],
            capture_output=True, text=True, timeout=60,
            env={**os.environ, "PYTHONPATH": str(SOURCE_ROOT)},
        )
        assert proc.returncode == 0, f"exit {proc.returncode}: {proc.stderr}"
        out = json.loads(proc.stdout)
        assert out["mode"] == "applied"
        assert out["output_path"] == str(out_path)
        assert out["commits"] > 0
        assert out["versions"] > 0

        # file 존재
        assert out_path.exists()
        body = out_path.read_text()

        # Keep-a-Changelog 형식 검증
        assert body.startswith("# Changelog\n")
        assert "## [Unreleased]" in body or re.search(r"## \[\d+\.\d+\.\d+\]", body), \
            "no version section"
        # Section header
        assert "### Added" in body, "no Added section"
        # commit 형식: `<subject> (<short>)`
        assert re.search(r"\([0-9a-f]{7}\)", body), "no commit hash format"


# --- Test 4: section categorization (commit subject prefix → section) ---


def test_changelog_gen_section_categorization() -> None:
    """commit subject prefix → Keep-a-Changelog section mapping.

    categorize_by_section() 직접 호출 + 각 prefix 의 section 검증.
    """
    mod = _import_tool()
    # standard prefix
    assert mod.categorize_by_section("feat(v0.7.0): ...") == "Added"
    assert mod.categorize_by_section("fix(v0.7.0): ...") == "Fixed"
    assert mod.categorize_by_section("docs(v0.7.0): ...") == "Changed"
    assert mod.categorize_by_section("chore(v0.7.0): ...") == "Changed"
    # unknown prefix
    assert mod.categorize_by_section("random subject") == "Changed"
    # known: refactor, perf, test, build, ci
    assert mod.categorize_by_section("refactor(v0.7.0): ...") == "Changed"
    assert mod.categorize_by_section("perf(v0.7.0): ...") == "Changed"
    assert mod.categorize_by_section("test(v0.7.0): ...") == "Changed"


# --- Test 5: --from-tag/--to-tag range filter (v0.7.15+) ---


def test_changelog_gen_range_filter() -> None:
    """--from-tag/--to-tag 으로 range scan. dry-run mode.

    v0.7.0-beta..v0.7.10-beta range 의 commit count 가 full history 보다 적어야 함.
    """
    # full history
    proc_full = subprocess.run(
        [sys.executable, str(TOOL), "changelog-gen", "--dry-run", "--json"],
        capture_output=True, text=True, timeout=60,
        env={**os.environ, "PYTHONPATH": str(SOURCE_ROOT)},
    )
    assert proc_full.returncode == 0
    full = json.loads(proc_full.stdout)
    full_commits = full["commits"]
    assert full_commits > 50, f"full history expected 50+, got {full_commits}"

    # range v0.7.0..v0.7.10
    proc_range = subprocess.run(
        [sys.executable, str(TOOL), "changelog-gen",
         "--from-tag=v0.7.0-beta", "--to-tag=v0.7.10-beta",
         "--dry-run", "--json"],
        capture_output=True, text=True, timeout=60,
        env={**os.environ, "PYTHONPATH": str(SOURCE_ROOT)},
    )
    assert proc_range.returncode == 0
    rng = json.loads(proc_range.stdout)
    assert rng["from_tag"] == "v0.7.0-beta"
    assert rng["to_tag"] == "v0.7.10-beta"
    assert rng["commits"] < full_commits, \
        f"range commits ({rng['commits']}) should be < full ({full_commits})"
    assert rng["commits"] > 0

    # range v0.7.5..v0.7.8 (smaller)
    proc_small = subprocess.run(
        [sys.executable, str(TOOL), "changelog-gen",
         "--from-tag=v0.7.5-beta", "--to-tag=v0.7.8-beta",
         "--dry-run", "--json"],
        capture_output=True, text=True, timeout=60,
        env={**os.environ, "PYTHONPATH": str(SOURCE_ROOT)},
    )
    assert proc_small.returncode == 0
    small = json.loads(proc_small.stdout)
    assert small["commits"] < rng["commits"], \
        f"smaller range ({small['commits']}) should be < larger ({rng['commits']})"


# --- Test 6: out-of-range (invalid tag) graceful fail ---


def test_changelog_gen_out_of_range_graceful() -> None:
    """invalid --from-tag 시 graceful fail (mode=error, error 메시지).

    v9.9.9-beta 등 존재하지 않는 tag. `git log` 가 exit != 0 → empty commits → error.
    """
    proc = subprocess.run(
        [sys.executable, str(TOOL), "changelog-gen",
         "--from-tag=v9.9.9-beta", "--dry-run", "--json"],
        capture_output=True, text=True, timeout=30,
        env={**os.environ, "PYTHONPATH": str(SOURCE_ROOT)},
    )
    out = json.loads(proc.stdout) if proc.stdout else {}
    if "mode" not in out and "error" in out:
        # error dict (no mode key) — convert to {mode: "error", error: ...}
        out["mode"] = "error"
    assert out.get("mode") == "error", f"expected error mode, got {out.get('mode')}"
    err_msg = out.get("error", "")
    assert "v9.9.9-beta" in err_msg or "no commits" in err_msg, \
        f"expected error message, got {err_msg}"


# --- 메인 실행 ---


def main() -> int:
    test_funcs = [
        test_changelog_gen_argparse,
        test_changelog_gen_dry_run,
        test_changelog_gen_apply,
        test_changelog_gen_section_categorization,
        test_changelog_gen_range_filter,
        test_changelog_gen_out_of_range_graceful,
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
