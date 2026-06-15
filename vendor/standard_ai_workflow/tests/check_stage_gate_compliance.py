#!/usr/bin/env python3
"""Stage Gate Compliance (v0.6.4) regression check.

Verifies that workflow_kit.common.contracts.stage_gate correctly:
- validates StageCompletion 8-field shape
- enforces 2-option only (no 3-option, no 4-option)
- distinguishes auto-approval environments (ci/cron/p0_hotfix only)
- blocks auto-approval for production code / state doc / release
- appends audit log in ISO 8601 format (append-only)
- emits 2-option completion message in canonical format
Reference: workflow-source/core/stage_gate_pattern.md
"""

from __future__ import annotations

import re
import sys
import tempfile
from pathlib import Path
from re import compile as re_compile

SOURCE_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.common.contracts.stage_gate import (  # noqa: E402
    ALLOWED_OPTIONS,
    AUTO_APPROVAL_ENVS,
    ApprovalActor,
    StageCompletion,
    StageStatus,
    append_audit_log,
    emit_completion_message,
    normalize_option_label,
    require_explicit_approval,
    validate_completion,
)


# --- Test cases ---


def test_stage_completion_happy_path() -> None:
    """user 가 명시적 approval → gate 통과."""
    sc = StageCompletion(
        stage_name="code-generation",
        stage_status="ok",
        next_stage="build-and-test",
        artifacts=["src/services/user.py"],
        approval_timestamp="2026-06-12T22:30:15Z",
        approval_actor="user",
        notes=["3-method implementation, all unit tests pass"],
    )
    assert sc.is_approved() is True
    result = validate_completion(sc)
    assert result.is_valid, f"expected valid, got errors: {result.errors}"


def test_stage_completion_pending() -> None:
    """approval_timestamp None → gate 미통과 (is_approved False)."""
    sc = StageCompletion(
        stage_name="requirements-analysis",
        stage_status="ok",
        next_stage="user-stories",
        artifacts=["docs/requirements.md"],
        approval_timestamp=None,
        approval_actor=None,
    )
    assert sc.is_approved() is False


def test_invalid_stage_status() -> None:
    """stage_status 가 enum 외 값 → validation error."""
    # type: ignore[arg-type]  # intentional bad input
    sc = StageCompletion(
        stage_name="test",
        stage_status="invalid_status",  # type: ignore[arg-type]
    )
    result = validate_completion(sc)
    assert not result.is_valid
    assert any(e.field == "stage_status" for e in result.errors)


def test_invalid_iso_timestamp() -> None:
    """approval_timestamp 가 ISO 8601 형식 아님 → validation error."""
    sc = StageCompletion(
        stage_name="test",
        stage_status="ok",
        approval_timestamp="not-iso-8601",
        approval_actor="user",
    )
    result = validate_completion(sc)
    assert not result.is_valid
    assert any(e.field == "approval_timestamp" for e in result.errors)


def test_2_option_only() -> None:
    """ALLOWED_OPTIONS 가 정확히 2개 (Request Changes, Continue) — 3-option/4-option ❌."""
    assert len(ALLOWED_OPTIONS) == 2, f"ALLOWED_OPTIONS must be 2-option, got {len(ALLOWED_OPTIONS)}"
    assert "request_changes" in ALLOWED_OPTIONS
    assert "continue" in ALLOWED_OPTIONS


def test_auto_approval_in_ci_env() -> None:
    """CI 환경에서 auto-approval → 허용."""
    sc = StageCompletion(
        stage_name="ci-test",
        stage_status="ok",
        approval_timestamp="2026-06-12T22:30:15Z",
        approval_actor="auto",
        notes=["auto-approved: CI timeout (5min gate)"],
    )
    allowed, reason = require_explicit_approval(sc, env="ci")
    assert allowed is True
    assert "auto" in reason.lower() or "ci" in reason.lower()


def test_auto_approval_in_dev_env_blocked() -> None:
    """dev 환경에서 auto-approval → 차단 (user explicit approval mandatory)."""
    sc = StageCompletion(
        stage_name="local-edit",
        stage_status="ok",
        approval_timestamp="2026-06-12T22:30:15Z",
        approval_actor="auto",
        notes=["auto-approved: 편의상"],
    )
    allowed, reason = require_explicit_approval(sc, env="dev")
    assert allowed is False
    assert "env" in reason.lower()


def test_auto_approval_production_code_blocked() -> None:
    """production code 변경에 auto-approval → 차단."""
    sc = StageCompletion(
        stage_name="prod-deploy",
        stage_status="ok",
        approval_timestamp="2026-06-12T22:30:15Z",
        approval_actor="auto",
        notes=["auto-approved: CI timeout"],
    )
    allowed, reason = require_explicit_approval(
        sc, env="ci", is_production_code=True
    )
    assert allowed is False
    assert "production" in reason.lower()


def test_auto_approval_state_doc_blocked() -> None:
    """state 문서 갱신에 auto-approval → 차단."""
    sc = StageCompletion(
        stage_name="state-update",
        stage_status="ok",
        approval_timestamp="2026-06-12T22:30:15Z",
        approval_actor="auto",
        notes=["auto-approved: CI timeout"],
    )
    allowed, reason = require_explicit_approval(
        sc, env="ci", is_state_doc=True
    )
    assert allowed is False
    assert "state" in reason.lower()


def test_auto_approval_release_blocked() -> None:
    """release / version tag 에 auto-approval → 차단."""
    sc = StageCompletion(
        stage_name="release",
        stage_status="ok",
        approval_timestamp="2026-06-12T22:30:15Z",
        approval_actor="auto",
        notes=["auto-approved: CI timeout"],
    )
    allowed, reason = require_explicit_approval(
        sc, env="ci", is_release=True
    )
    assert allowed is False
    assert "release" in reason.lower()


def test_p0_hotfix_auto_approval_allowed() -> None:
    """P0 hotfix + notes 에 사유 → auto-approval 허용."""
    sc = StageCompletion(
        stage_name="hotfix",
        stage_status="ok",
        approval_timestamp="2026-06-12T22:30:15Z",
        approval_actor="auto",
        notes=["P0 hotfix: skip gate, reason: prod outage"],
    )
    allowed, reason = require_explicit_approval(sc, env="p0_hotfix")
    assert allowed is True


def test_audit_log_append_only() -> None:
    """audit log 에 append-only 기록 (overwrite ❌)."""
    sc = StageCompletion(
        stage_name="first-stage",
        stage_status="ok",
        next_stage="second-stage",
        artifacts=["output.md"],
        approval_timestamp="2026-06-12T22:30:15Z",
        approval_actor="user",
        notes=["test entry 1"],
    )
    audit_path = Path(tempfile.gettempdir()) / "test-audit.md"
    try:
        append_audit_log(audit_path, sc)
        content_first = audit_path.read_text(encoding="utf-8")
        assert "first-stage" in content_first
        assert "2026-06-12T22:30:15Z" in content_first

        # 두 번째 entry — 기존 내용 보존되어야 함
        sc2 = StageCompletion(
            stage_name="second-stage",
            stage_status="warning",
            next_stage="third-stage",
            artifacts=["output2.md"],
            approval_timestamp="2026-06-12T22:35:00Z",
            approval_actor="user",
            notes=["test entry 2"],
        )
        append_audit_log(audit_path, sc2)
        content_second = audit_path.read_text(encoding="utf-8")

        # 첫 번째 entry 보존
        assert "first-stage" in content_second, "first entry must be preserved (append-only)"
        assert "second-stage" in content_second
        # 두 timestamp 모두 존재
        assert "22:30:15Z" in content_second
        assert "22:35:00Z" in content_second
    finally:
        audit_path.unlink(missing_ok=True)


def test_emit_completion_message_2_option() -> None:
    """emit_completion_message 가 정확히 2-option 만 emit (3-option ❌)."""
    msg = emit_completion_message(
        stage_name="code-generation",
        artifacts=["src/services/user.py", "tests/test_user.py"],
        next_stage="build-and-test",
        notes=["3-method implementation", "all unit tests pass"],
        stage_status="ok",
    )
    # 2-option 만 있는지 검증 — line-by-line scan
    option_lines: list[str] = []
    for line in msg.splitlines():
        # 🔧 또는 ✅ 로 시작하는 markdown 옵션 line
        if line.strip().startswith("🔧") or line.strip().startswith("✅"):
            option_lines.append(line.strip())
    assert len(option_lines) == 2, f"expected 2 options, got {len(option_lines)}: {option_lines}"
    # 정확한 옵션 이름
    option_text = " ".join(option_lines).lower()
    assert "request changes" in option_text
    assert "continue" in option_text


def test_normalize_option_label() -> None:
    """label 정규화 — AIDLC 의 'Approve & Continue' 도 'continue' 로."""
    assert normalize_option_label("Request Changes") == "request_changes"
    assert normalize_option_label("Continue to Next Stage") == "continue"
    assert normalize_option_label("✅ Approve & Continue") == "continue"
    assert normalize_option_label("🔧 Request Changes") == "request_changes"
    assert normalize_option_label("invalid option") == ""


def test_no_approval_recorded() -> None:
    """approval_actor=None → user approval mandatory (default)."""
    sc = StageCompletion(
        stage_name="test",
        stage_status="ok",
    )
    allowed, reason = require_explicit_approval(sc, env="dev")
    assert allowed is False
    assert "no approval" in reason.lower() or "user" in reason.lower()


# --- Main ---

def main() -> int:
    tests = [
        ("stage_completion_happy_path", test_stage_completion_happy_path),
        ("stage_completion_pending", test_stage_completion_pending),
        ("invalid_stage_status", test_invalid_stage_status),
        ("invalid_iso_timestamp", test_invalid_iso_timestamp),
        ("2_option_only", test_2_option_only),
        ("auto_approval_in_ci_env", test_auto_approval_in_ci_env),
        ("auto_approval_in_dev_env_blocked", test_auto_approval_in_dev_env_blocked),
        ("auto_approval_production_code_blocked", test_auto_approval_production_code_blocked),
        ("auto_approval_state_doc_blocked", test_auto_approval_state_doc_blocked),
        ("auto_approval_release_blocked", test_auto_approval_release_blocked),
        ("p0_hotfix_auto_approval_allowed", test_p0_hotfix_auto_approval_allowed),
        ("audit_log_append_only", test_audit_log_append_only),
        ("emit_completion_message_2_option", test_emit_completion_message_2_option),
        ("normalize_option_label", test_normalize_option_label),
        ("no_approval_recorded", test_no_approval_recorded),
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
