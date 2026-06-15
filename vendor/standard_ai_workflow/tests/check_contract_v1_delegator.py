#!/usr/bin/env python3
"""Contract v1 §6.1 / §6.3 delegator regression check.

Verifies that workflow_kit.contract_v1.delegator.choose_role correctly:
- maps the 4 task_types to roles per §6.1
- rejects all 9 §6.3 MUST-NOT-delegate markers (v0.5.6: 7 + v0.5.7: cross-ref / fan-in 통합)
- raises on unknown task_type
- generates valid §4 delegation_id format
- (v0.5.7) recommends model_tier based on task keywords

Reference: workflow-source/core/orchestrator_subagent_contract_v1.md §6
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.contract_v1 import (  # noqa: E402
    choose_role,
    DelegationRejected,
    enforce_subagent_response,
    recommend_model_tier,
)

DELEGATION_ID_RE = re.compile(r"^del-\d{4}-\d{2}-\d{2}-[0-9a-f]{8}$")


def _task(task_type: str, brief: str = "draft spec", primary_artifact: str = "spec.md") -> dict[str, object]:
    return {
        "task_type": task_type,
        "brief": brief,
        "expected_outputs": {"primary_artifact": primary_artifact},
    }


def check_task_type_to_role_mapping() -> None:
    """§6.1: each task_type maps to its expected role."""
    cases = [
        ("doc_draft", "doc-worker"),
        ("code_change", "code-worker"),
        ("validation_run", "validation-worker"),
        ("bounded_research", "doc-worker"),
    ]
    for task_type, expected_role in cases:
        decision = choose_role(_task(task_type))
        if decision.role != expected_role:
            raise AssertionError(
                f"task_type={task_type!r} should map to {expected_role!r}, got {decision.role!r}"
            )
        if decision.must_not_delegate:
            raise AssertionError(f"task_type={task_type!r} should NOT be §6.3 rejected")
        if not DELEGATION_ID_RE.match(decision.delegation_id or ""):
            raise AssertionError(
                f"delegation_id format wrong: {decision.delegation_id!r} (expected del-YYYY-MM-DD-xxxxxxxx)"
            )


def check_must_not_delegate_rejection() -> None:
    """§6.3: all 9 markers must be rejected with must_not_delegate=True.

    v0.5.6: 7 markers (handoff, backlog, state.json, ask_user, 우선순위, 통합/리뷰, PR 본문)
    v0.5.7: +2 (cross-ref, fan-in 통합)
    """
    cases = [
        ("handoff 갱신 및 정리", "handoff 갱신은 orchestrator 직접 처리 (contract v1 §6.3)"),
        ("backlog update", "backlog 갱신은 orchestrator 직접 처리 (contract v1 §6.3)"),
        ("state.json 동기화", "state.json 갱신은 orchestrator 직접 처리 (contract v1 §6.3)"),
        ("ask_user 호출", "ask_user 호출은 orchestrator 직접 처리 (contract v1 §6.3)"),
        ("우선순위 결정", "우선순위 결정은 orchestrator 직접 처리 (contract v1 §6.3)"),
        ("sub-agent 출력 통합/리뷰", "sub-agent 출력 통합/리뷰는 orchestrator 직접 처리 (contract v1 §6.3)"),
        ("PR 본문 작성", "PR 본문 작성은 orchestrator 직접 처리 (contract v1 §6.3)"),
        # v0.5.7 신규
        (
            "docs cross-ref 갱신 (3 doc-link + 2 memory link)",
            "cross-ref 갱신은 orchestrator 직접 처리 (contract v1 §6.3, v0.5.7)",
        ),
        (
            "fan-in 통합 보고서 작성",
            "fan-in 통합 보고는 orchestrator 직접 처리 (contract v1 §6.3, v0.5.7)",
        ),
    ]
    for brief, expected_reason in cases:
        decision = choose_role(_task("doc_draft", brief=brief))
        if not decision.must_not_delegate:
            raise AssertionError(f"brief={brief!r} should be §6.3 rejected, but was not")
        if decision.rejected_reason != expected_reason:
            raise AssertionError(
                f"brief={brief!r} reason mismatch:\n  expected: {expected_reason}\n  got: {decision.rejected_reason}"
            )
        if decision.role is not None:
            raise AssertionError(f"brief={brief!r} should have role=None, got {decision.role!r}")


def check_strict_mode_raises() -> None:
    """strict=True must raise DelegationRejected on §6.3 violation."""
    try:
        choose_role(_task("doc_draft", brief="handoff 갱신"), strict=True)
    except DelegationRejected as exc:
        if not exc.decision.must_not_delegate:
            raise AssertionError("DelegationRejected but decision.must_not_delegate is False")
        return
    raise AssertionError("strict=True on §6.3 should raise DelegationRejected")


def check_unknown_task_type_raises() -> None:
    """Unknown task_type must raise ValueError."""
    try:
        choose_role(_task("unknown_type_xyz"))
    except ValueError as exc:
        if "task_type" not in str(exc):
            raise AssertionError(f"ValueError should mention 'task_type', got: {exc}")
        return
    raise AssertionError("Unknown task_type should raise ValueError")


def check_marker_in_primary_artifact_also_rejected() -> None:
    """§6.3 markers in primary_artifact (not just brief) must also be rejected."""
    decision = choose_role(_task(
        "doc_draft",
        brief="draft something",
        primary_artifact="handoff-update.md",  # marker in path
    ))
    if not decision.must_not_delegate:
        raise AssertionError(
            "Marker in primary_artifact should also trigger §6.3 rejection"
        )


def check_non_must_not_delegate_brief_passes() -> None:
    """Brief that contains unrelated substring should NOT be rejected (regression
    against overly broad substring matching)."""
    # "비교" contains "교" which is in "통합/리뷰" — should NOT match
    decision = choose_role(_task(
        "doc_draft",
        brief="v0.5.3 release notes 와 비교하여 변경점 정리",
    ))
    if decision.must_not_delegate:
        raise AssertionError(
            f"비교 should NOT trigger §6.3 rejection, but got: {decision.rejected_reason}"
        )


def check_recommend_model_tier_main_keywords() -> None:
    """v0.5.7: model_tier 자동 결정 — main 후보 keyword 매치 시 'main'."""
    main_cases = [
        {"brief": "아키텍처 재설계 — 모듈 경계 변경", "task_type": "code_change"},
        {"brief": "정책 문구 추가 (보안 정책 v2)", "task_type": "doc_draft"},
        {"brief": "5+ 파일 cross-cutting refactor", "task_type": "code_change"},
        {"brief": "bounded_research 5+ source 종합", "task_type": "bounded_research"},
    ]
    for task in main_cases:
        tier = recommend_model_tier(task)
        if tier != "main":
            raise AssertionError(
                f"task brief={task['brief']!r} should recommend 'main', got {tier!r}"
            )


def check_recommend_model_tier_default_small() -> None:
    """v0.5.7: main keyword 가 없으면 'small'."""
    small_cases = [
        {"brief": "단일 파일 200줄 read", "task_type": "doc_draft"},
        {"brief": "build / lint / test 실행", "task_type": "validation_run"},
        {"brief": "small fix 1 file", "task_type": "code_change"},
    ]
    for task in small_cases:
        tier = recommend_model_tier(task)
        if tier != "small":
            raise AssertionError(
                f"task brief={task['brief']!r} should recommend 'small', got {tier!r}"
            )


def check_recommend_model_tier_explicit_override() -> None:
    """v0.5.7: required_model_tier 명시 시 override."""
    # 명시적으로 small 인데 brief 가 main keyword 있어도 small
    tier = recommend_model_tier({
        "brief": "아키텍처 재설계",
        "task_type": "code_change",
        "required_model_tier": "small",
    })
    if tier != "small":
        raise AssertionError(f"explicit small should override brief keyword, got {tier!r}")
    # 명시적으로 main 인데 brief 가 일반이어도 main
    tier = recommend_model_tier({
        "brief": "small fix",
        "task_type": "code_change",
        "required_model_tier": "main",
    })
    if tier != "main":
        raise AssertionError(f"explicit main should override default, got {tier!r}")


def check_choose_role_returns_recommended_tier() -> None:
    """v0.5.7: choose_role 결과에 recommended_model_tier 박혀 있어야."""
    decision = choose_role(_task("code_change", brief="아키텍처 변경"))
    if decision.recommended_model_tier != "main":
        raise AssertionError(
            f"choose_role should propagate recommended_model_tier=main, got {decision.recommended_model_tier!r}"
        )


def check_enforce_subagent_response_happy_path() -> None:
    """v0.5.11 §6.5: enforce_subagent_response 가 정상 응답을 통과시킨다."""
    valid_response = {
        "contract_version": "1.0",
        "delegation_id": "del-2026-06-09-aaaaaaaa",
        "completed_at": "2026-06-09T12:00:00Z",
        "worker": {"session_id": "mvs_test", "role": "doc-worker", "model_tier": "small"},
        "result": {
            "status": "ok",
            "summary": "Test response",
            "artifacts": [],
            "written_paths": [],
            "next_step": "continue",
        },
    }
    enforce_subagent_response(
        valid_response, expected_delegation_id="del-2026-06-09-aaaaaaaa"
    )


def check_enforce_subagent_response_violation_raises_value_error() -> None:
    """v0.5.11 §6.5: expected_delegation_id 불일치 시 ValueError raise."""
    bad_response = {
        "contract_version": "1.0",
        "delegation_id": "del-2026-06-09-bbbbbbbb",
        "completed_at": "2026-06-09T12:00:00Z",
        "worker": {"session_id": "mvs_test", "role": "doc-worker", "model_tier": "small"},
        "result": {
            "status": "ok",
            "summary": "Test",
            "artifacts": [],
            "written_paths": [],
            "next_step": "continue",
        },
    }
    try:
        enforce_subagent_response(
            bad_response, expected_delegation_id="del-2026-06-09-aaaaaaaa"
        )
    except ValueError as exc:
        if "delegation_id" not in str(exc):
            raise AssertionError(
                f"ValueError should mention 'delegation_id', got: {exc}"
            )
        return
    raise AssertionError(
        "expected_delegation_id mismatch should raise ValueError"
    )


def main() -> int:
    check_task_type_to_role_mapping()
    check_must_not_delegate_rejection()
    check_strict_mode_raises()
    check_unknown_task_type_raises()
    check_marker_in_primary_artifact_also_rejected()
    check_non_must_not_delegate_brief_passes()
    # v0.5.7 신규
    check_recommend_model_tier_main_keywords()
    check_recommend_model_tier_default_small()
    check_recommend_model_tier_explicit_override()
    check_choose_role_returns_recommended_tier()
    # v0.5.11 신규 — §6.5 Mavis engine hook
    check_enforce_subagent_response_happy_path()
    check_enforce_subagent_response_violation_raises_value_error()
    print(
        "Contract v1 §6.1/§6.3 delegator smoke check passed "
        "(4 task_type mappings, 9 must-not-delegate rejections [7 v0.5.6 + 2 v0.5.7], "
        "strict mode, unknown type, primary_artifact marker, non-match baseline, "
        "model_tier main keywords, model_tier default small, model_tier explicit "
        "override, choose_role propagates tier, enforce_subagent_response happy + violation)."
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
