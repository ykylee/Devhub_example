#!/usr/bin/env python3
"""atomic_write helper smoke test (DevHub-specific, v0.7.15 follow-up D).

scripts/atomic_write.py 자체 검증 (POSIX only, macOS/Linux). 12-15 test 로
atomic_write 의 핵심 invariant + DevHub 적용 시나리오를 cover.

Test 구성 (16 test):
1.  test_normal_text_write              — 기본 text write
2.  test_normal_json_write              — JSON write (indent=2)
3.  test_json_round_trip                — write → load 동일
4.  test_unicode_korean                 — 한글 round-trip
5.  test_json_ensure_ascii_false        — Unicode 보존 (escape ❌)
6.  test_subdir_mkdir_parents           — parent dir 자동 생성
7.  test_overwrite_idempotent           — 기존 file 덮어쓰기
8.  test_no_temp_leftover               — *.tmp.*.atomic file 0
9.  test_partial_failure_temp_cleanup  — write 실패 시 temp file cleanup
10. test_json_indent_custom             — indent=4 등 custom
11. test_empty_string_write             — 빈 문자열 write
12. test_multiline_write                — multi-line text
13. test_log_md_append_pattern          — DevHub log.md append 패턴 (read+append+write)
14. test_state_json_round_trip          — state.json 패턴 (JSON dict 큰 케이스)
15. test_trailing_newline_convention    — JSON trailing newline 보장
16. test_atomic_append_text_concurrent  — P1 race regression (PR #607, 8 thread × 25 line)

Reference:
- scripts/atomic_write.py (v0.7.15 follow-up D)
- POSIX rename(2) atomicity guarantee
- vendor/standard_ai_workflow/v0.7.15+ workflow_kit.common.atomic_write (pattern 동일, 독립)

Wiki: ai-workflow/wiki/patterns/in-repo-redirect.md
      ai-workflow/wiki/decisions/v0.7.17-import.md
"""

from __future__ import annotations

import json
import os
import shutil
import subprocess
import sys
import tempfile
import threading
import uuid
from pathlib import Path

# ----- paths -----
SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent
ATOMIC_WRITE_PY = REPO_ROOT / "scripts" / "atomic_write.py"

# ----- POSIX guard -----
if sys.platform == "win32":
    print(f"[check-atomic-write] SKIP: POSIX only (current platform: {sys.platform})")
    print(f"[check-atomic-write] RESULT: 0/0 SKIP")
    sys.exit(0)


# ----- result tracker -----
PASS = 0
FAIL = 0
FAILURES: list[str] = []


def _ok(name: str) -> None:
    global PASS
    PASS += 1
    print(f"  [PASS] {name}")


def _fail(name: str, reason: str) -> None:
    global FAIL
    FAIL += 1
    FAILURES.append(f"{name}: {reason}")
    print(f"  [FAIL] {name}: {reason}")


def _import_atomic_write():
    """scripts/atomic_write.py 동적 import."""
    sys.path.insert(0, str(REPO_ROOT / "scripts"))
    import atomic_write  # type: ignore[import-not-found]
    return atomic_write


# ----- 1. normal text write -----
def test_normal_text_write():
    name = "test_normal_text_write"
    try:
        aw = _import_atomic_write()
        with tempfile.TemporaryDirectory() as d:
            p = Path(d) / "test.txt"
            aw.atomic_write_text(p, "hello world\n")
            if p.read_text(encoding="utf-8") != "hello world\n":
                return _fail(name, "text mismatch")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 2. normal json write -----
def test_normal_json_write():
    name = "test_normal_json_write"
    try:
        aw = _import_atomic_write()
        with tempfile.TemporaryDirectory() as d:
            p = Path(d) / "test.json"
            aw.atomic_write_json(p, {"a": 1, "b": [1, 2, 3]})
            content = p.read_text(encoding="utf-8")
            if '"a": 1' not in content or '"b"' not in content:
                return _fail(name, f"unexpected JSON: {content!r}")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 3. json round-trip -----
def test_json_round_trip():
    name = "test_json_round_trip"
    try:
        aw = _import_atomic_write()
        with tempfile.TemporaryDirectory() as d:
            p = Path(d) / "rt.json"
            data = {"key": "value", "list": [1, 2, 3], "nested": {"a": True}}
            aw.atomic_write_json(p, data)
            loaded = json.loads(p.read_text(encoding="utf-8"))
            if loaded != data:
                return _fail(name, f"round-trip mismatch: {loaded} != {data}")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 4. Unicode (한글) -----
def test_unicode_korean():
    name = "test_unicode_korean"
    try:
        aw = _import_atomic_write()
        with tempfile.TemporaryDirectory() as d:
            p = Path(d) / "kr.txt"
            kr_text = "한글 테스트 — atomic write Unicode 보존"
            aw.atomic_write_text(p, kr_text)
            if p.read_text(encoding="utf-8") != kr_text:
                return _fail(name, "Korean text mismatch")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 5. JSON ensure_ascii=False (default) -----
def test_json_ensure_ascii_false():
    name = "test_json_ensure_ascii_false"
    try:
        aw = _import_atomic_write()
        with tempfile.TemporaryDirectory() as d:
            p = Path(d) / "kr.json"
            aw.atomic_write_json(p, {"name": "한글", "desc": "テスト"})
            content = p.read_text(encoding="utf-8")
            # ensure_ascii=False 이면 한글 직접 저장
            if "\\u" in content:
                return _fail(name, f"Unicode escape found (should not): {content!r}")
            if "한글" not in content:
                return _fail(name, f"한글 missing: {content!r}")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 6. sub-dir mkdir parents -----
def test_subdir_mkdir_parents():
    name = "test_subdir_mkdir_parents"
    try:
        aw = _import_atomic_write()
        with tempfile.TemporaryDirectory() as d:
            p = Path(d) / "deep" / "nested" / "sub" / "test.txt"
            aw.atomic_write_text(p, "deep")
            if not p.exists():
                return _fail(name, "file not created in nested dir")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 7. overwrite idempotent -----
def test_overwrite_idempotent():
    name = "test_overwrite_idempotent"
    try:
        aw = _import_atomic_write()
        with tempfile.TemporaryDirectory() as d:
            p = Path(d) / "ow.txt"
            aw.atomic_write_text(p, "first")
            aw.atomic_write_text(p, "second")
            aw.atomic_write_text(p, "third")
            if p.read_text(encoding="utf-8") != "third":
                return _fail(name, "overwrite failed")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 8. no temp leftover -----
def test_no_temp_leftover():
    name = "test_no_temp_leftover"
    try:
        aw = _import_atomic_write()
        with tempfile.TemporaryDirectory() as d:
            for i in range(5):
                p = Path(d) / f"f{i}.txt"
                aw.atomic_write_text(p, f"content {i}")
            leftovers = list(Path(d).glob("*.tmp.*.atomic"))
            if leftovers:
                return _fail(name, f"temp leftovers: {leftovers}")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 9. partial failure temp cleanup -----
def test_partial_failure_temp_cleanup():
    name = "test_partial_failure_temp_cleanup"
    try:
        aw = _import_atomic_write()
        with tempfile.TemporaryDirectory() as d:
            # invalid path: write to a path under a non-directory
            blocker = Path(d) / "blocker"
            blocker.write_text("I am a file, not a dir")
            bad = blocker / "child" / "file.txt"  # parent is not a dir
            try:
                aw.atomic_write_text(bad, "should fail")
                # 일부 FS 는 file-as-parent 도 accept → OSError 안 날 수 있음
                # 그 경우 test 통과 (FS-dependent)
            except (OSError, NotADirectoryError, FileExistsError):
                pass  # expected
            # temp leftover check
            leftovers = list(Path(d).rglob("*.tmp.*.atomic"))
            if leftovers:
                return _fail(name, f"temp leftovers after failure: {leftovers}")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 10. json indent custom -----
def test_json_indent_custom():
    name = "test_json_indent_custom"
    try:
        aw = _import_atomic_write()
        with tempfile.TemporaryDirectory() as d:
            p = Path(d) / "ind.json"
            aw.atomic_write_json(p, {"x": 1, "y": 2}, indent=4)
            content = p.read_text(encoding="utf-8")
            if "    \"x\"" not in content:
                return _fail(name, f"indent=4 not applied: {content!r}")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 11. empty string write -----
def test_empty_string_write():
    name = "test_empty_string_write"
    try:
        aw = _import_atomic_write()
        with tempfile.TemporaryDirectory() as d:
            p = Path(d) / "empty.txt"
            aw.atomic_write_text(p, "")
            if not p.exists():
                return _fail(name, "empty file not created")
            if p.read_text(encoding="utf-8") != "":
                return _fail(name, "empty file has content")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 12. multiline write -----
def test_multiline_write():
    name = "test_multiline_write"
    try:
        aw = _import_atomic_write()
        with tempfile.TemporaryDirectory() as d:
            p = Path(d) / "ml.txt"
            content = "line 1\nline 2\nline 3\n"
            aw.atomic_write_text(p, content)
            if p.read_text(encoding="utf-8") != content:
                return _fail(name, "multiline mismatch")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 13. log.md append pattern (DevHub usage) -----
def test_log_md_append_pattern():
    name = "test_log_md_append_pattern"
    try:
        aw = _import_atomic_write()
        with tempfile.TemporaryDirectory() as d:
            log = Path(d) / "log.md"
            # initial 1 line
            aw.atomic_write_text(log, "[2026-06-15] query | initial | results=0\n")
            # append 2 more lines (read+append+write pattern, DevHub wiki-query.sh 동일)
            for i in range(1, 3):
                existing = log.read_text(encoding="utf-8")
                new_line = f"[2026-06-15] query | line{i} | results={i}\n"
                aw.atomic_write_text(log, existing + new_line)
            # verify
            final = log.read_text(encoding="utf-8")
            lines = [l for l in final.split("\n") if l]
            if len(lines) != 3:
                return _fail(name, f"expected 3 lines, got {len(lines)}: {lines}")
            if "line1" not in final or "line2" not in final:
                return _fail(name, f"appended lines missing: {final!r}")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 14. state.json round-trip (large nested) -----
def test_state_json_round_trip():
    name = "test_state_json_round_trip"
    try:
        aw = _import_atomic_write()
        with tempfile.TemporaryDirectory() as d:
            p = Path(d) / "state.json"
            # state.json pattern: nested dict with list + dict
            data = {
                "branch": "feat/atomic-write-devhub",
                "agent": "coder",
                "status": "in_progress",
                "todos": [
                    {"id": 1, "content": "explore", "status": "done"},
                    {"id": 2, "content": "implement", "status": "in_progress"},
                ],
                "metadata": {
                    "version": "v0.7.15",
                    "follow_up": "D",
                },
            }
            aw.atomic_write_json(p, data)
            loaded = json.loads(p.read_text(encoding="utf-8"))
            if loaded != data:
                return _fail(name, "state.json round-trip mismatch")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 16. trailing newline convention -----
def test_trailing_newline_convention():
    name = "test_trailing_newline_convention"
    try:
        aw = _import_atomic_write()
        with tempfile.TemporaryDirectory() as d:
            p = Path(d) / "tn.json"
            # write dict without trailing newline
            aw.atomic_write_json(p, {"k": "v"})
            content = p.read_text(encoding="utf-8")
            if not content.endswith("\n"):
                return _fail(name, f"missing trailing newline: {content!r}")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 16. atomic_append_text — concurrent append (P1 race regression, PR #607) -----
def test_atomic_append_text_concurrent():
    name = "test_atomic_append_text_concurrent"
    try:
        aw = _import_atomic_write()
        if not hasattr(aw, "atomic_append_text"):
            return _fail(name, "atomic_append_text helper missing")
        with tempfile.TemporaryDirectory() as d:
            log = Path(d) / "concurrent.log"
            n_threads = 8
            n_lines_per_thread = 25

            errors: list[str] = []

            def worker(thread_id: int) -> None:
                try:
                    for i in range(n_lines_per_thread):
                        aw.atomic_append_text(
                            log, f"t{thread_id:02d}-i{i:02d}\n"
                        )
                except Exception as e:  # noqa: BLE001
                    errors.append(f"t{thread_id}: {e}")

            threads = [threading.Thread(target=worker, args=(i,)) for i in range(n_threads)]
            for t in threads:
                t.start()
            for t in threads:
                t.join()

            if errors:
                return _fail(name, f"thread errors: {errors}")

            content = log.read_text(encoding="utf-8")
            lines = [l for l in content.split("\n") if l]
            expected = n_threads * n_lines_per_thread
            if len(lines) != expected:
                return _fail(
                    name,
                    f"expected {expected} lines, got {len(lines)} (lost updates: {expected - len(lines)})",
                )
            seen: set[str] = set()
            for line in lines:
                if line in seen:
                    return _fail(name, f"duplicate line: {line}")
                seen.add(line)
            for t in range(n_threads):
                for i in range(n_lines_per_thread):
                    tag = f"t{t:02d}-i{i:02d}"
                    if tag not in seen:
                        return _fail(name, f"missing line: {tag}")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")

# ----- main -----
def main() -> int:
    print(f"[check-atomic-write] === atomic_write helper smoke (DevHub, v0.7.15 follow-up D) ===")
    print(f"[check-atomic-write]   module: {ATOMIC_WRITE_PY}")
    print(f"[check-atomic-write]   python: {sys.version.split()[0]}")
    print(f"[check-atomic-write]   platform: {sys.platform}")
    print("")

    if not ATOMIC_WRITE_PY.exists():
        print(f"[check-atomic-write] FAIL: module not found: {ATOMIC_WRITE_PY}")
        return 1

    # run all tests
    test_normal_text_write()
    test_normal_json_write()
    test_json_round_trip()
    test_unicode_korean()
    test_json_ensure_ascii_false()
    test_subdir_mkdir_parents()
    test_overwrite_idempotent()
    test_no_temp_leftover()
    test_partial_failure_temp_cleanup()
    test_json_indent_custom()
    test_empty_string_write()
    test_multiline_write()
    test_log_md_append_pattern()
    test_state_json_round_trip()
    test_trailing_newline_convention()
    test_atomic_append_text_concurrent()

    print("")
    print(f"[check-atomic-write] === summary ===")
    print(f"[check-atomic-write]   PASS: {PASS}")
    print(f"[check-atomic-write]   FAIL: {FAIL}")
    print(f"[check-atomic-write]   total: {PASS}/{PASS + FAIL}")

    if FAIL > 0:
        print("")
        print(f"[check-atomic-write] FAILURES:")
        for f in FAILURES:
            print(f"  - {f}")
        return 1

    print(f"[check-atomic-write] RESULT: {PASS}/{PASS} PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
