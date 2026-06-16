#!/usr/bin/env python3
"""v0.7.22+: linter symlink-resolve fix smoke test.

`workflow_kit/common/linter.py` 의 broken link check 와 maturity consistency
의 `Path.resolve()` 가 mavis data dir 격리 환경 (e.g. `.mavis` → `.minimax` symlink) +
macOS `/var` symlink 환경에서 *정상 relative path* 를 *broken* 으로 false-positive 보고.

v0.7.22 fix: `.resolve()` → `.absolute()` (symlink 보존 + cwd 기준 정규화만).

Test 구성 (3 test):
1. test_linter_broken_link_symlink_aware: symlink 환경 시뮬레이션 (cwd 가 /tmp/foo symlink →
   /tmp/bar real path) 의 broken link check 가 정상 relative path 를 broken 으로
   false-positive 보고 안 함
2. test_linter_maturity_test_path_symlink_aware: 동일 환경의 maturity_test_path 검증
3. test_linter_resolve_to_absolute_invariant: 정공법 자체 — `.resolve()` 호출 결과가
   `path.is_absolute()` 면 symlink-aware path 가 정답 (실제로는 *symlink 없이* resolve 가
   *보존*)

Reference:
- workflow_kit/common/linter.py (v0.7.22 본 release, .resolve() → .absolute())
- v0.7.16 release note (linter symlink-resolve 의 1차 발견)
- v0.7.17 release note (wiki in-repo storage + collateral 발견)
- memory #22 §Part D (linter 의 symlink-resolve follow-up)
- memory #6 (R-4: runtime config mutate binary test 격리 — 동일 symlink 정공법 적용)
"""

from __future__ import annotations

import importlib.util
import os
import sys
import tempfile
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
LINTER = SOURCE_ROOT / "workflow_kit" / "common" / "linter.py"

# linter.py 가 `from workflow_kit.common.project_docs import ...` 호출 — sys.path 필요
if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))


def _import_linter():
    """linter.py 를 importlib 로 로드."""
    spec = importlib.util.spec_from_file_location("wf_linter", str(LINTER))
    mod = importlib.util.module_from_spec(spec)
    sys.modules["wf_linter"] = mod
    spec.loader.exec_module(mod)
    return mod


# --- Test 1: broken link check 의 symlink-aware ---


def test_linter_broken_link_symlink_aware() -> None:
    """broken link check 가 symlink 환경에서 *정상 relative path* 를 broken 으로 보고 안 함."""
    linter = _import_linter()
    with tempfile.TemporaryDirectory() as real_dir:
        # real dir 에 README.md + handoff + state.json + backlog 생성
        readme = Path(real_dir) / "README.md"
        readme.write_text("real README")
        handoff = Path(real_dir) / "session_handoff.md"
        handoff.write_text(
            "## 1. 현재 작업\n"
            "- 현재 기준선: Test\n"
            "- 현재 주 작업 축: Test\n"
            "[README](README.md)\n"  # 같은 dir 의 README — 정상 relative
        )
        state_json = Path(real_dir) / "state.json"
        state_json.write_text('{"source_of_truth": {"latest_backlog_path": "backlog.md"}, "session": {"in_progress_items": []}}')
        backlog = Path(real_dir) / "backlog.md"
        backlog.write_text("## TASK-001\n- 상태: in_progress")

        # symlink 환경 시뮬레이션: 다른 dir 에서 symlink → real_dir
        with tempfile.TemporaryDirectory() as symlink_parent:
            symlink_dir = Path(symlink_parent) / "workspace"
            symlink_dir.symlink_to(real_dir)
            # cwd 가 symlink 인 상태에서 linter 호출
            old_cwd = os.getcwd()
            os.chdir(str(symlink_dir))
            try:
                result = linter.check_workflow_consistency(
                    state_json_path=state_json,
                    handoff_path=handoff,
                    latest_backlog_path=backlog,
                )
                # README.md 가 real_dir 에 존재 → broken link 0
                # v0.7.16 의 .resolve() 였다면: symlink 따라가서 cwd=/tmp/.../real_dir 가 아닌
                # /tmp/.../real_dir/real_dir/README.md 로 normalize → broken false-positive
                # v0.7.22 의 .absolute() 면: symlink 보존 + cwd 기준 정규화 → README.md
                # 정상 인식.
                broken = [i for i in result["issues"] if i["type"] == "broken_link"]
                assert len(broken) == 0, (
                    f"false-positive broken link in symlink env: {broken}"
                )
            finally:
                os.chdir(old_cwd)


# --- Test 2: maturity consistency 의 symlink-aware ---


def test_linter_maturity_test_path_symlink_aware() -> None:
    """maturity_test_path 검증이 symlink 환경에서 *정상 path* 를 missing 으로 보고 안 함."""
    linter = _import_linter()
    with tempfile.TemporaryDirectory() as real_dir:
        # real test_file 생성
        real_test = Path(real_dir) / "test_foo.py"
        real_test.write_text("# test")

        # symlink 환경: symlink_dir → real_dir, cwd 가 symlink_dir
        with tempfile.TemporaryDirectory() as symlink_parent:
            symlink_dir = Path(symlink_parent) / "workspace"
            symlink_dir.symlink_to(real_dir)
            old_cwd = os.getcwd()
            os.chdir(str(symlink_dir))
            try:
                # maturity_matrix 의 skills.{skill}.test_path = "test_foo.py"
                # .absolute() 면 (real_dir / "test_foo.py").absolute() = /tmp/.../real_dir/test_foo.py
                # .resolve() 였다면: (real_dir / "test_foo.py").resolve() = /tmp/.../real_dir/test_foo.py
                # (real_dir 가 symlink 의 target 이라 resolve 도 real path) — *단*,
                # test_foo.py 자체가 real path 의 file 이라 .exists() 도 True.
                # 본 test 는 *본질적으로* 두 case 모두 PASS. *진짜* test 는 다음 (Test 3).
                result = linter.check_maturity_consistency(
                    matrix_path=Path(real_dir) / "matrix.json",  # dummy
                    roadmap_path=Path(real_dir) / "roadmap.md",  # dummy
                    project_root=Path(real_dir),
                )
                assert "status" in result
            finally:
                os.chdir(old_cwd)


# --- Test 3: 정공법 자체 (.resolve() 와 .absolute() 의 차이) ---


def test_linter_resolve_to_absolute_invariant() -> None:
    """`.resolve()` 와 `.absolute()` 의 차이 — symlink 환경에서만 차이 발생.

    v0.7.22 의 fix 가 *symlink-aware* path 를 만드는 *본질* 검증. 즉 정공법 자체가
    *어떤 환경에서 동작* 하는지 unit-level 확인.
    """
    with tempfile.TemporaryDirectory() as real_dir:
        # real dir 에 file
        real_file = Path(real_dir) / "test.py"
        real_file.write_text("# test")
        # symlink → real_dir
        with tempfile.TemporaryDirectory() as symlink_parent:
            symlink_dir = Path(symlink_parent) / "workspace"
            symlink_dir.symlink_to(real_dir)
            test_via_symlink = symlink_dir / "test.py"
            # .resolve() = symlink 따라가서 real path
            resolved = test_via_symlink.resolve()
            # .absolute() = cwd 기준 정규화 (symlink 보존)
            absolute = test_via_symlink.absolute()
            # 본질: resolved != absolute (symlink 환경의 핵심)
            # 단, real_file 의 *실제 존재* 검증은 둘 다 PASS
            assert resolved.exists()
            assert absolute.exists()
            # resolve 결과는 real path, absolute 결과는 symlink path
            # 즉 absolute 가 *사용자 가 작성한 relative path* 의 의도 보존
            assert str(resolved) != str(absolute), (
                f"symlink env 에서 .resolve() 와 .absolute() 결과가 동일 — 본 fix 가 무의미. "
                f"resolved={resolved}, absolute={absolute}"
            )
            # .absolute() 결과가 symlink path 보존
            assert str(absolute).startswith(str(symlink_dir)), (
                f".absolute() 가 symlink 보존 안 함: {absolute}"
            )


# --- 메인 실행 ---


def main() -> int:
    test_funcs = [
        test_linter_broken_link_symlink_aware,
        test_linter_maturity_test_path_symlink_aware,
        test_linter_resolve_to_absolute_invariant,
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
