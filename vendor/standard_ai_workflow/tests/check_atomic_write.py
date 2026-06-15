"""workflow_kit.common.atomic_write helper smoke test (v0.7.15+).

POSIX 의 os.replace atomic guarantee 기반 JSON / text atomic write. 3 test PASS.

Test list:
1. test_atomic_write_json_creates_file: 부재 path → file 생성 + 내용 정합
2. test_atomic_write_json_replaces_existing: 기존 file → 새 내용으로 replace
3. test_atomic_write_json_atomicity: tmp file 패턴 + os.replace 사용 검증 (POSIX)
"""

from __future__ import annotations

import json
import os
import re
import subprocess
import sys
import tempfile
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
ATOMIC_WRITE = SOURCE_ROOT / "workflow_kit" / "common" / "atomic_write.py"


def _import_atomic():
    """atomic_write 모듈 importlib 로 load."""
    import importlib.util
    spec = importlib.util.spec_from_file_location("atomic_write", str(ATOMIC_WRITE))
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


# --- Test 1: 부재 path → file 생성 + 내용 정합 ---


def test_atomic_write_json_creates_file() -> None:
    """부재 dir + 부재 file → atomic_write_json 으로 file 생성 + 내용 검증."""
    with tempfile.TemporaryDirectory() as tmpdir:
        target = Path(tmpdir) / "subdir" / "state.json"
        data = {"session": {"in_progress": [], "recent_done": ["v0.7.14"]}, "memory": {"last_freeze": "2026-06-15"}}
        mod = _import_atomic()
        mod.atomic_write_json(target, data)

        assert target.exists(), "file not created"
        assert target.parent.exists(), "parent dir not created"
        body = target.read_text(encoding="utf-8")
        loaded = json.loads(body)
        assert loaded == data, f"content mismatch: {loaded} != {data}"


# --- Test 2: 기존 file → 새 내용으로 replace ---


def test_atomic_write_json_replaces_existing() -> None:
    """기존 file 의 내용을 새 dict 로 replace. 이전 내용 부재 확인."""
    with tempfile.TemporaryDirectory() as tmpdir:
        target = Path(tmpdir) / "state.json"
        # initial content
        target.write_text(json.dumps({"v": 1}), encoding="utf-8")
        pre = target.read_text(encoding="utf-8")
        assert "v\": 1" in pre

        # replace
        mod = _import_atomic()
        new_data = {"session": {"in_progress": [], "recent_done": []}, "memory": {"version": "0.7.15"}}
        mod.atomic_write_json(target, new_data)

        post = target.read_text(encoding="utf-8")
        loaded = json.loads(post)
        assert loaded == new_data
        # old content absent
        assert '"v": 1' not in post, "old content not removed"


# --- Test 3: atomicity (tmp file + os.replace 패턴 검증) ---


def test_atomic_write_json_atomicity() -> None:
    """atomic_write_json 가 tempfile + os.replace 패턴 사용. (POSIX atomic guarantee)

    source code level 검증: 'tempfile' + 'os.replace' + 'mkstemp' 의 import / 사용 존재.
    """
    text = ATOMIC_WRITE.read_text()
    # import 검증
    assert "import tempfile" in text, "tempfile import missing"
    assert "import os" in text, "os import missing"
    # 핵심 atomic 패턴
    assert "os.replace" in text, "os.replace call missing (POSIX atomic rename)"
    assert "mkstemp" in text, "mkstemp call missing (unique tmp file)"
    assert "os.fsync" in text, "os.fsync call missing (durability before rename)"
    # tmp file prefix (target dir)
    assert "tmp" in text, "tmp file marker missing"


# --- 메인 실행 ---


def main() -> int:
    test_funcs = [
        test_atomic_write_json_creates_file,
        test_atomic_write_json_replaces_existing,
        test_atomic_write_json_atomicity,
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
