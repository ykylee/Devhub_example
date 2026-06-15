#!/usr/bin/env python3
"""Stage Gate Runtime (v0.6.5) regression check.

Verifies that workflow_kit.common.contracts.stage_gate_runtime correctly:
- build_stage_completion() returns dict with 8 field
- merge_into_result() preserves existing fields (status, warnings, source_context)
- merge_into_result() is idempotent (artifacts 합집합)
- merge_into_result() overwrites when overwrite=True
- is_stage_completion_present() detects valid 8-field dict
- get_stage_status_from_result() prefers stage_completion over legacy 'status'
- emit_and_log() outputs 2-option message + optional audit log
- 기존 52 smoke test 호환: stage_completion 없는 result 도 valid (optional field)
Reference: workflow-source/core/stage_gate_runtime_migration.md (TBD)
"""

from __future__ import annotations

import sys
import tempfile
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.common.contracts.stage_gate_runtime import (  # noqa: E402
    build_stage_completion,
    emit_and_log,
    get_stage_status_from_result,
    is_stage_completion_present,
    merge_into_result,
)


def test_build_stage_completion_basic() -> None:
    """build_stage_completion() returns 8-field dict."""
    sc = build_stage_completion(
        stage_name="session-start",
        stage_status="ok",
        artifacts=["ai-workflow/memory/active/state.json"],
        next_stage=None,
        notes=["session restored"],
    )
    assert isinstance(sc, dict)
    assert sc["stage_name"] == "session-start"
    assert sc["stage_status"] == "ok"
    assert sc["next_stage"] is None
    assert sc["artifacts"] == ["ai-workflow/memory/active/state.json"]
    assert sc["notes"] == ["session restored"]
    # 미승인 상태
    assert sc["approval_timestamp"] is None
    assert sc["approval_actor"] is None
    assert sc["requested_changes"] == []


def test_build_stage_completion_with_approval() -> None:
    """build_stage_completion() with approval fields."""
    sc = build_stage_completion(
        stage_name="doc-sync",
        stage_status="ok",
        artifacts=["doc.md"],
        next_stage="validation-plan",
        approval_timestamp="2026-06-12T22:30:15Z",
        approval_actor="user",
    )
    assert sc["approval_timestamp"] == "2026-06-12T22:30:15Z"
    assert sc["approval_actor"] == "user"
    assert sc["next_stage"] == "validation-plan"


def test_merge_into_result_preserves_existing() -> None:
    """merge_into_result() 보존 기존 field."""
    original = {
        "status": "ok",
        "tool_version": "0.6.5",
        "warnings": ["stale handoff"],
        "source_context": {"path": "x.md"},
    }
    sc = build_stage_completion(
        stage_name="session-start",
        stage_status="ok",
        artifacts=["new.md"],
    )
    merged = merge_into_result(original, sc)
    # 기존 field 보존
    assert merged["status"] == "ok"
    assert merged["tool_version"] == "0.6.5"
    assert merged["warnings"] == ["stale handoff"]
    assert merged["source_context"] == {"path": "x.md"}
    # stage_completion 추가
    assert "stage_completion" in merged
    assert merged["stage_completion"]["stage_name"] == "session-start"


def test_merge_into_result_idempotent() -> None:
    """merge_into_result() idempotent — 기존 artifacts 와 새 artifacts 합집합."""
    sc1 = build_stage_completion(
        stage_name="test",
        stage_status="ok",
        artifacts=["a.md", "b.md"],
    )
    sc2 = build_stage_completion(
        stage_name="test",
        stage_status="ok",
        artifacts=["b.md", "c.md"],
    )
    base = {"status": "ok"}
    after_first = merge_into_result(base, sc1)
    after_second = merge_into_result(after_first, sc2)
    # 합집합 (a, b, c)
    assert set(after_second["stage_completion"]["artifacts"]) == {"a.md", "b.md", "c.md"}


def test_merge_into_result_overwrite() -> None:
    """overwrite=True 이면 기존 stage_completion 완전 교체."""
    sc1 = build_stage_completion(
        stage_name="first",
        stage_status="ok",
        artifacts=["a.md"],
    )
    sc2 = build_stage_completion(
        stage_name="second",
        stage_status="error",
        artifacts=["b.md"],
    )
    base = {"status": "ok"}
    after_first = merge_into_result(base, sc1)
    after_second = merge_into_result(after_first, sc2, overwrite=True)
    # overwrite=True → 두 번째가 완전 교체
    assert after_second["stage_completion"]["stage_name"] == "second"
    assert after_second["stage_completion"]["stage_status"] == "error"
    assert after_second["stage_completion"]["artifacts"] == ["b.md"]


def test_is_stage_completion_present_valid() -> None:
    """8 field 모두 있는 stage_completion → True."""
    result = {
        "status": "ok",
        "stage_completion": build_stage_completion(
            stage_name="test",
            stage_status="ok",
        ),
    }
    assert is_stage_completion_present(result) is True


def test_is_stage_completion_present_missing() -> None:
    """stage_completion 없거나 field 부족 → False."""
    # 없음
    assert is_stage_completion_present({"status": "ok"}) is False
    # 부족
    assert is_stage_completion_present({
        "status": "ok",
        "stage_completion": {"stage_name": "test"},  # 7 field 부족
    }) is False


def test_get_stage_status_prefers_stage_completion() -> None:
    """stage_completion.stage_status 우선."""
    result = {
        "status": "ok",  # legacy
        "stage_completion": {
            "stage_name": "test",
            "stage_status": "warning",
        },
    }
    assert get_stage_status_from_result(result) == "warning"


def test_get_stage_status_falls_back_to_legacy() -> None:
    """stage_completion 없으면 legacy 'status' 사용."""
    result = {"status": "error"}
    assert get_stage_status_from_result(result) == "error"


def test_emit_and_log_basic() -> None:
    """emit_and_log() returns 2-option message without audit log."""
    msg = emit_and_log(
        stage_name="code-generation",
        artifacts=["src/services/user.py"],
        next_stage="build-and-test",
        notes=["3-method implementation"],
    )
    assert "code-generation" in msg
    assert "Request Changes" in msg
    assert "Continue" in msg


def test_emit_and_log_with_audit_log() -> None:
    """emit_and_log() with audit_log_path → audit log entry 추가."""
    audit_path = Path(tempfile.gettempdir()) / "test-runtime-audit.md"
    try:
        emit_and_log(
            stage_name="runtime-test",
            artifacts=["out.md"],
            next_stage=None,
            stage_status="ok",
            notes=["runtime helper test"],
            audit_log_path=audit_path,
            approval_timestamp="2026-06-12T22:30:15Z",
            approval_actor="user",
        )
        content = audit_path.read_text(encoding="utf-8")
        assert "runtime-test" in content
        assert "2026-06-12T22:30:15Z" in content
        assert "approved" in content.lower()
    finally:
        audit_path.unlink(missing_ok=True)


def test_emit_and_log_audit_log_pending() -> None:
    """audit_log_path + approval 없으면 'pending' entry."""
    audit_path = Path(tempfile.gettempdir()) / "test-runtime-pending.md"
    try:
        emit_and_log(
            stage_name="pending-test",
            artifacts=["out.md"],
            next_stage=None,
            stage_status="warning",
            notes=["no approval yet"],
            audit_log_path=audit_path,
        )
        content = audit_path.read_text(encoding="utf-8")
        assert "pending-test" in content
        assert "pending" in content.lower()
    finally:
        audit_path.unlink(missing_ok=True)


def test_existing_result_without_stage_completion_still_valid() -> None:
    """기존 52 smoke test 호환: stage_completion 없는 result 도 valid."""
    # legacy result (52 smoke test 의 baseline format)
    legacy_result = {
        "status": "ok",
        "tool_version": "0.6.3",
        "warnings": [],
        "source_context": {"path": "x.md"},
    }
    # is_stage_completion_present → False (정상)
    assert is_stage_completion_present(legacy_result) is False
    # get_stage_status → legacy 'status' (정상)
    assert get_stage_status_from_result(legacy_result) == "ok"
    # merge_into_result 로 stage_completion 추가 가능
    sc = build_stage_completion(stage_name="test", stage_status="ok")
    merged = merge_into_result(legacy_result, sc)
    # 기존 field 보존 + stage_completion 추가
    assert merged["status"] == "ok"
    assert merged["tool_version"] == "0.6.3"
    assert "stage_completion" in merged


# --- Main ---

def main() -> int:
    tests = [
        ("build_stage_completion_basic", test_build_stage_completion_basic),
        ("build_stage_completion_with_approval", test_build_stage_completion_with_approval),
        ("merge_into_result_preserves_existing", test_merge_into_result_preserves_existing),
        ("merge_into_result_idempotent", test_merge_into_result_idempotent),
        ("merge_into_result_overwrite", test_merge_into_result_overwrite),
        ("is_stage_completion_present_valid", test_is_stage_completion_present_valid),
        ("is_stage_completion_present_missing", test_is_stage_completion_present_missing),
        ("get_stage_status_prefers_stage_completion", test_get_stage_status_prefers_stage_completion),
        ("get_stage_status_falls_back_to_legacy", test_get_stage_status_falls_back_to_legacy),
        ("emit_and_log_basic", test_emit_and_log_basic),
        ("emit_and_log_with_audit_log", test_emit_and_log_with_audit_log),
        ("emit_and_log_audit_log_pending", test_emit_and_log_audit_log_pending),
        ("existing_result_without_stage_completion_still_valid", test_existing_result_without_stage_completion_still_valid),
    ]
    failed: list[str] = []
    for name, fn in tests:
        try:
            fn()
            print(f"  PASS  {name}")
        except AssertionError as e:
            failed.append(f"  FAIL  {name}: {e}")
            print(f"  FAIL  {name}: {e}")
        except Exception as e:  # noqa: BLE001
            failed.append(f"  ERROR {name}: {type(e).__name__}: {e}")
            print(f"  ERROR {name}: {type(e).__name__}: {e}")

    print()
    if failed:
        print(f"{len(failed)}/{len(tests)} tests failed:")
        for f in failed:
            print(f)
        return 1
    print(f"All {len(tests)} tests passed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
