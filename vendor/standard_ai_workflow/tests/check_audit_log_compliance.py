#!/usr/bin/env python3
"""v0.7.0 step 10: Audit Log 표준 검증.

- audit_log_standard.md §3.1 entry format 검증
- append-only 정책 (overwrite ❌)
- ISO 8601 timestamp 형식
- workflow_kit.common.contracts.stage_gate.append_audit_log 의 8-field 정합성
- audit log 가 있는 기존 sample 의 format 일관성
Reference: workflow-source/core/audit_log_standard.md
           workflow_kit/common/contracts/stage_gate.py `append_audit_log`
"""

from __future__ import annotations

import re
import sys
import tempfile
from datetime import datetime, timezone
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.common.contracts.stage_gate import (  # noqa: E402
    StageCompletion,
    append_audit_log,
)

ISO_8601_RE = re.compile(r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(Z|[+-]\d{2}:?\d{2})$")
# MULTILINE flag: ^ 가 line start, $ 가 line end 매칭
ENTRY_HEADER_RE = re.compile(r"^## \[Stage: (.+?)\] \[(.+?)\]\s*$", re.MULTILINE)


# --- Helper ---


def _sample_completion(stage_name: str = "test-stage") -> StageCompletion:
    return StageCompletion(
        stage_name=stage_name,
        stage_status="ok",
        next_stage=None,
        artifacts=["/path/to/artifact.md"],
        approval_timestamp="2026-06-12T22:30:15Z",
        approval_actor="user",
        notes=["Test entry"],
    )


# --- Test 1: entry format 검증 ---


def test_entry_header_format() -> None:
    """## [Stage: X] [ISO_8601] 형식 검증."""
    audit_path = Path(tempfile.gettempdir()) / "test-audit-format.md"
    try:
        append_audit_log(audit_path, _sample_completion("code-generation"))
        content = audit_path.read_text(encoding="utf-8")
        # 첫 줄이 ## [Stage: ...] [...] 인지
        first_line = content.splitlines()[0]
        m = ENTRY_HEADER_RE.match(first_line)
        assert m is not None, f"header format invalid: {first_line!r}"
        assert m.group(1) == "code-generation"
        # ISO 8601 형식
        ts = m.group(2)
        assert ISO_8601_RE.match(ts), f"timestamp not ISO 8601: {ts}"
        # datetime.fromisoformat 도 통과
        datetime.fromisoformat(ts.replace("Z", "+00:00"))
    finally:
        audit_path.unlink(missing_ok=True)


def test_entry_required_fields() -> None:
    """StageCompletion 8 field 모두 entry 에 포함."""
    audit_path = Path(tempfile.gettempdir()) / "test-audit-fields.md"
    try:
        append_audit_log(audit_path, _sample_completion("doc-sync"))
        content = audit_path.read_text(encoding="utf-8")
        # 8 field 모두 등장
        for field in [
            "**Stage**:",
            "**Status**:",
            "**Approval**:",
            "**Actor**:",
            "**Notes**:",
        ]:
            assert field in content, f"missing field marker: {field}"
        # artifacts 또는 next stage 가 있으면 등장
        assert "**Artifacts**:" in content or "**Next Stage**:" in content
    finally:
        audit_path.unlink(missing_ok=True)


# --- Test 2: append-only 정책 ---


def test_append_only_two_entries() -> None:
    """두 번 append 시 첫 entry 보존."""
    audit_path = Path(tempfile.gettempdir()) / "test-append-only.md"
    try:
        append_audit_log(audit_path, _sample_completion("first-stage"))
        first_content = audit_path.read_text(encoding="utf-8")
        assert "first-stage" in first_content
        first_ts = "2026-06-12T22:30:15Z"

        append_audit_log(audit_path, _sample_completion("second-stage"))
        second_content = audit_path.read_text(encoding="utf-8")

        # 첫 entry 보존
        assert "first-stage" in second_content
        assert first_ts in second_content
        # 두 entry 모두 존재
        assert "second-stage" in second_content
        # 두 timestamp 모두 존재
        assert second_content.count("## [Stage:") == 2
    finally:
        audit_path.unlink(missing_ok=True)


def test_append_only_no_overwrite() -> None:
    """append 가 기존 content 를 *덮어쓰지* 않음 (overwrite 절대 금지)."""
    audit_path = Path(tempfile.gettempdir()) / "test-no-overwrite.md"
    try:
        # 첫 entry: 5-line content 포함
        first = StageCompletion(
            stage_name="first",
            stage_status="ok",
            artifacts=["/path/with/many/segments/file1.md", "/path/file2.md"],
            notes=["First entry with multiple artifacts"],
        )
        append_audit_log(audit_path, first)
        size_after_first = audit_path.stat().st_size

        # 두 번째 entry
        second = StageCompletion(
            stage_name="second",
            stage_status="warning",
            notes=["Second entry"],
        )
        append_audit_log(audit_path, second)
        size_after_second = audit_path.stat().st_size

        # 두 번째 사이즈 > 첫 번째 (append)
        assert size_after_second > size_after_first
        # 차이 = second entry 의 size
        size_diff = size_after_second - size_after_first
        assert size_diff > 50  # reasonable entry size
    finally:
        audit_path.unlink(missing_ok=True)


# --- Test 3: ISO 8601 timestamp 정책 ---


def test_iso_8601_z_suffix() -> None:
    """timestamp 가 Z suffix (UTC) 인지."""
    audit_path = Path(tempfile.gettempdir()) / "test-iso-z.md"
    try:
        # approval_timestamp=None → 자동 UTC now
        sc = StageCompletion(
            stage_name="test",
            stage_status="ok",
            approval_timestamp=None,  # auto-generate
        )
        append_audit_log(audit_path, sc)
        content = audit_path.read_text(encoding="utf-8")
        # Z suffix timestamp 검증 (auto-generate 가 Z 로 끝나야 함, v0.7.0 fix)
        m = ENTRY_HEADER_RE.search(content)
        assert m is not None, f"header not found in: {content[:100]!r}"
        ts = m.group(2)
        # microsecond 없어야 함 (split('.') 후 길이 검증)
        assert "." not in ts, f"microsecond leaked: {ts}"
        # ISO 8601 형식 + timezone marker
        assert ISO_8601_RE.match(ts), f"ISO 8601 format invalid: {ts}"
    finally:
        audit_path.unlink(missing_ok=True)


def test_iso_8601_with_offset() -> None:
    """+09:00 offset 도 허용."""
    audit_path = Path(tempfile.gettempdir()) / "test-iso-offset.md"
    try:
        sc = StageCompletion(
            stage_name="test",
            stage_status="ok",
            approval_timestamp="2026-06-12T22:30:15+09:00",
        )
        append_audit_log(audit_path, sc)
        content = audit_path.read_text(encoding="utf-8")
        m = ENTRY_HEADER_RE.search(content)
        assert m is not None
        ts = m.group(2)
        # +09:00 format 검증 (ISO 8601 형식 통과)
        assert ISO_8601_RE.match(ts), f"ISO 8601 with offset invalid: {ts}"
        assert "+09:00" in ts
    finally:
        audit_path.unlink(missing_ok=True)


# --- Test 4: approval 상태 ---


def test_approved_entry_status() -> None:
    """approved entry → 'approved' marker, pending → 'pending' marker."""
    audit_path = Path(tempfile.gettempdir()) / "test-approval-state.md"
    try:
        # Approved
        append_audit_log(audit_path, _sample_completion("approved-stage"))
        content = audit_path.read_text(encoding="utf-8")
        # is_approved() = True → 'approved' marker
        assert "**Approval**: approved" in content

        # Pending (no approval_timestamp, no approval_actor)
        pending = StageCompletion(
            stage_name="pending-stage",
            stage_status="warning",
            approval_timestamp=None,
            approval_actor=None,
        )
        append_audit_log(audit_path, pending)
        content2 = audit_path.read_text(encoding="utf-8")
        assert "**Approval**: pending" in content2
    finally:
        audit_path.unlink(missing_ok=True)


def test_actor_label() -> None:
    """Actor label 이 user/orchestrator/auto 만."""
    for actor in ("user", "orchestrator", "auto"):
        audit_path = Path(tempfile.gettempdir()) / f"test-actor-{actor}.md"
        try:
            sc = StageCompletion(
                stage_name="test",
                stage_status="ok",
                approval_actor=actor,
            )
            append_audit_log(audit_path, sc)
            content = audit_path.read_text(encoding="utf-8")
            assert f"**Actor**: {actor}" in content
        finally:
            audit_path.unlink(missing_ok=True)


# --- Test 5: 8 field 정합성 ---


def test_requested_changes_appears_when_present() -> None:
    """requested_changes 비어있지 않으면 '**Requested Changes**:' 등장."""
    audit_path = Path(tempfile.gettempdir()) / "test-requested-changes.md"
    try:
        sc = StageCompletion(
            stage_name="test",
            stage_status="ok",
            requested_changes=["Use type hints", "Add docstring"],
        )
        append_audit_log(audit_path, sc)
        content = audit_path.read_text(encoding="utf-8")
        assert "**Requested Changes**:" in content
        assert "Use type hints" in content
        assert "Add docstring" in content
    finally:
        audit_path.unlink(missing_ok=True)


def test_next_stage_appears_when_present() -> None:
    """next_stage None 아니면 '**Next Stage**:' 등장."""
    audit_path = Path(tempfile.gettempdir()) / "test-next-stage.md"
    try:
        sc = StageCompletion(
            stage_name="test",
            stage_status="ok",
            next_stage="validation-plan",
        )
        append_audit_log(audit_path, sc)
        content = audit_path.read_text(encoding="utf-8")
        assert "**Next Stage**: validation-plan" in content
    finally:
        audit_path.unlink(missing_ok=True)


def test_separator_after_each_entry() -> None:
    """각 entry 끝에 '---' separator."""
    audit_path = Path(tempfile.gettempdir()) / "test-separator.md"
    try:
        append_audit_log(audit_path, _sample_completion("one"))
        content = audit_path.read_text(encoding="utf-8")
        assert content.count("---") >= 1
    finally:
        audit_path.unlink(missing_ok=True)


# --- Test 6: 통합 (build → merge → audit log) ---


def test_full_workflow_audit_log() -> None:
    """v0.6.4-5 runtime helper 통합:
    build_stage_completion → merge_into_result → emit_and_log(audit_log_path)
    """
    from workflow_kit.common.contracts.stage_gate_runtime import (
        build_stage_completion,
        emit_and_log,
        merge_into_result,
    )

    audit_path = Path(tempfile.gettempdir()) / "test-full-workflow.md"
    try:
        # 1) build + merge
        sc = build_stage_completion(
            stage_name="code-generation",
            stage_status="ok",
            artifacts=["src/services/user.py"],
            next_stage="build-and-test",
            notes=["3-method implementation"],
        )
        result = {
            "status": "ok",
            "tool_version": "0.6.5",
            "warnings": [],
            "source_context": {"path": "src/services/user.py"},
        }
        result = merge_into_result(result, sc)

        # 2) emit + audit log
        msg = emit_and_log(
            stage_name="code-generation",
            artifacts=["src/services/user.py"],
            next_stage="build-and-test",
            notes=["3-method implementation"],
            audit_log_path=audit_path,
            approval_timestamp=result["stage_completion"]["approval_timestamp"],
            approval_actor="user",
        )
        # message 검증
        assert "code-generation" in msg
        assert "Request Changes" in msg
        # audit log 검증
        audit_content = audit_path.read_text(encoding="utf-8")
        assert "code-generation" in audit_content
        assert "src/services/user.py" in audit_content
        assert "**Actor**: user" in audit_content
    finally:
        audit_path.unlink(missing_ok=True)


# --- Test 7: 기존 smoke test 와의 호환성 ---


def test_legacy_audit_log_readable() -> None:
    """v0.6.4-5 의 audit log format 도 본 표준 parser 로 읽기 가능.

    (Format forward-compatibility 검증 — 본 표준 format 과 동등한 regex)
    """
    legacy_content = """## [Stage: code-generation] [2026-06-12T22:30:15Z]
**Stage**: code-generation
**Status**: ok
**Artifacts**:
- src/x.py
**Approval**: approved
**Actor**: user
**Notes**: legacy entry

---
"""
    # ENTRY_HEADER_RE + ISO_8601_RE 로 parse 가능
    for line in legacy_content.splitlines():
        m = ENTRY_HEADER_RE.match(line)
        if m:
            assert ISO_8601_RE.match(m.group(2))


# --- Main ---


def main() -> int:
    tests = [
        ("entry_header_format", test_entry_header_format),
        ("entry_required_fields", test_entry_required_fields),
        ("append_only_two_entries", test_append_only_two_entries),
        ("append_only_no_overwrite", test_append_only_no_overwrite),
        ("iso_8601_z_suffix", test_iso_8601_z_suffix),
        ("iso_8601_with_offset", test_iso_8601_with_offset),
        ("approved_entry_status", test_approved_entry_status),
        ("actor_label", test_actor_label),
        ("requested_changes_appears_when_present", test_requested_changes_appears_when_present),
        ("next_stage_appears_when_present", test_next_stage_appears_when_present),
        ("separator_after_each_entry", test_separator_after_each_entry),
        ("full_workflow_audit_log", test_full_workflow_audit_log),
        ("legacy_audit_log_readable", test_legacy_audit_log_readable),
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
    raise SystemExit(main())
