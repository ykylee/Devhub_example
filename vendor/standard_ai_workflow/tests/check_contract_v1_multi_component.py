#!/usr/bin/env python3
"""Contract v1 §4.2 / §5.2 multi-component fan-out / fan-in regression check.

Verifies that workflow_kit.contract_v1 (v0.5.7) correctly handles:
- delegator.choose_roles: batch delegation for `task.sub_tasks` (fan-out)
  - parent decision first, then per-sub decision with `sub_id` + parent-prefix delegation_id
  - schema validation for sub_tasks
  - §6.3 rejection in any sub propagates correctly
- output_validator.validate_fanin_output: fan-in payload validation
  - sub_results required fields + delegation_id prefix match
  - aggregated status consistency (any failed→failed, any partial→partial, else ok)
  - parent_delegation_id field
- baseline: 1 parent + 3 sub fan-out/in pilot walkthrough passes
- violation: parent §6.3 violation propagates, sub schema invalid raises

Reference: workflow-source/core/orchestrator_subagent_contract_v1.md §4.2, §5.2
"""

from __future__ import annotations

import sys
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.contract_v1 import (  # noqa: E402
    choose_roles,
    DelegationRejected,
    enforce_fanin_response,
    validate_fanin_output,
)


def _ok_payload(
    delegation_id: str = "del-2026-06-08-fanin-pilot",
    parent_id: str = "del-2026-06-08-fanout-pilot",
    sub_statuses: list[str] | None = None,
) -> dict[str, object]:
    """Build a valid §5.2 fan-in payload for testing."""
    if sub_statuses is None:
        sub_statuses = ["ok", "ok", "ok"]
    sub_results: list[dict[str, object]] = []
    for i, st in enumerate(sub_statuses, start=1):
        sub_results.append({
            "sub_id": f"st-{i}",
            "delegation_id": f"{parent_id}-st-{i}",
            "status": st,
            "summary": f"sub {i} done",
            "written_paths": [f"file_{i}.py"],
        })
    # aggregated status
    if any(s == "failed" for s in sub_statuses):
        agg = "failed"
    elif any(s == "partial" for s in sub_statuses):
        agg = "partial"
    else:
        agg = "ok"
    return {
        "contract_version": "1.0",
        "delegation_id": delegation_id,
        "completed_at": "2026-06-08T17:30:00+09:00",
        "parent_delegation_id": parent_id,
        "worker": {
            "session_id": "mvs_parent_test",
            "role": "code-worker",
            "model_tier": "main",
        },
        "result": {
            "status": agg,
            "summary": f"{len(sub_statuses)} sub-component 처리, aggregated={agg}",
            "artifacts": [
                {"path": "backend-core/handlers/auth_logout.go", "kind": "code", "action": "created", "added": 1, "removed": 0, "total": 1},
                {"path": "backend-core/handlers/ci_runs.go", "kind": "code", "action": "modified", "added": 5, "removed": 1, "total": 42},
            ],
            "written_paths": ["backend-core/handlers/auth_logout.go", "backend-core/handlers/ci_runs.go"],
            "sub_results": sub_results,
            "validation_result": {"ran": True, "status": "pass", "details": "all sub ok"},
            "next_step": "orchestrator 가 PR 본문 작성 + cross-link 갱신",
        },
    }


def _fanout_task() -> dict[str, object]:
    """A 3-sub fan-out task mirroring the v0.5.7 spec example.

    sub_tasks 의 parent_delegation_id 는 비워둠 — choose_roles 이 자동 채움
    (delegator 가 §4.2 parent delegation_id 발급 권한 = orchestrator).
    """
    return {
        "task_type": "code_change",
        "brief": "Devhub N+1 endpoint 1차 구현 + 4 시나리오 동시 진행",
        "expected_outputs": {
            "primary_artifact": "backend-core/handlers/",
            "artifact_kind": "code",
            "must_include": ["3 sub-component 모두 PASS", "fan-in aggregation 1 report"],
        },
        "sub_tasks": [
            {
                "sub_id": "st-1",
                "task_type": "code_change",
                "brief": "POST /auth/logout 구현 (OIDC + 204)",
                "primary_artifact": "backend-core/handlers/auth_logout.go",
                "artifact_kind": "code",
            },
            {
                "sub_id": "st-2",
                "task_type": "code_change",
                "brief": "POST /ci-runs 구현 (repository_id + 7 enum + 409 idempotency)",
                "primary_artifact": "backend-core/handlers/ci_runs.go",
                "artifact_kind": "code",
            },
            {
                "sub_id": "st-3",
                "task_type": "validation_run",
                "brief": "4 endpoint 통합 test suite 실행",
                "primary_artifact": "backend-core/tests/contract_v1_integration.py",
                "artifact_kind": "code",
            },
        ],
    }


def check_choose_roles_single_no_subtasks() -> None:
    """sub_tasks 없으면 단일 결과 (backward compatible)."""
    task = {
        "task_type": "doc_draft",
        "brief": "draft spec",
        "expected_outputs": {"primary_artifact": "spec.md"},
    }
    decisions = choose_roles(task)
    if len(decisions) != 1:
        raise AssertionError(f"single task should yield 1 decision, got {len(decisions)}")
    if decisions[0].role != "doc-worker":
        raise AssertionError(f"role mismatch: {decisions[0].role!r}")
    if decisions[0].sub_id is not None:
        raise AssertionError(f"single task should have sub_id=None, got {decisions[0].sub_id!r}")


def check_choose_roles_fanout_three_subs() -> None:
    """3 sub-task fan-out: parent + 3 sub decisions."""
    task = _fanout_task()
    decisions = choose_roles(task)
    if len(decisions) != 4:
        raise AssertionError(f"3-sub fan-out should yield 4 decisions (parent + 3), got {len(decisions)}")
    parent = decisions[0]
    if parent.role != "code-worker":
        raise AssertionError(f"parent role should be code-worker, got {parent.role!r}")
    if parent.delegation_id is None:
        raise AssertionError("parent delegation_id should be set")
    # sub decisions
    expected_sub_roles = ["code-worker", "code-worker", "validation-worker"]
    for i, sub_decision in enumerate(decisions[1:], start=0):
        if sub_decision.role != expected_sub_roles[i]:
            raise AssertionError(
                f"sub {i} role should be {expected_sub_roles[i]!r}, got {sub_decision.role!r}"
            )
        if sub_decision.sub_id != f"st-{i + 1}":
            raise AssertionError(
                f"sub {i} sub_id should be st-{i + 1}, got {sub_decision.sub_id!r}"
            )
        if sub_decision.delegation_id is None:
            raise AssertionError(f"sub {i} delegation_id should be set")
        if f"st-{i + 1}" not in sub_decision.delegation_id:
            raise AssertionError(
                f"sub {i} delegation_id {sub_decision.delegation_id!r} should contain sub_id st-{i + 1}"
            )
        # v0.5.10 spec 정합: sub delegation_id 는 반드시 parent delegation_id
        # 의 prefix 여야 한다 (validate_fanin_output 의 prefix check 와 정합).
        # v0.5.7.1 까지는 _generate_delegation_id_with_suffix 가 random parent
        # prefix 를 발급해 fanin validator 가 거절했음. v0.5.10 fix 후
        # choose_roles 가 직접 f"{parent_id}-st-{i}" 형식으로 발급.
        if not sub_decision.delegation_id.startswith(parent.delegation_id):
            raise AssertionError(
                f"sub {i} delegation_id {sub_decision.delegation_id!r} should "
                f"start with parent delegation_id {parent.delegation_id!r} "
                f"(v0.5.10 spec 정합 — fanin validator prefix check)"
            )
        # 서브 delegation_id 들이 서로 유니크
        sibling_ids = [d.delegation_id for d in decisions[1:] if d.delegation_id is not None]
        if len(sibling_ids) != len(set(sibling_ids)):
            raise AssertionError(
                f"sub delegation_ids should be unique within parent fan-out, got {sibling_ids}"
            )


def check_choose_roles_parent_must_not_delegate_propagates() -> None:
    """parent brief 가 §6.3 위반이면 fan-out 전체 거부 (non-strict)."""
    task = _fanout_task()
    task["brief"] = "handoff 갱신 + 3 sub fan-out"  # §6.3 marker
    decisions = choose_roles(task, strict=False)
    if not decisions[0].must_not_delegate:
        raise AssertionError("parent with §6.3 marker should be must_not_delegate=True")
    if decisions[0].rejected_reason is None or "handoff" not in decisions[0].rejected_reason:
        raise AssertionError(
            f"rejected_reason should mention handoff, got {decisions[0].rejected_reason!r}"
        )


def check_choose_roles_parent_must_not_delegate_strict_raises() -> None:
    """parent §6.3 위반 + strict=True 면 DelegationRejected raise."""
    task = _fanout_task()
    task["brief"] = "PR 본문 작성 + 3 sub"
    try:
        choose_roles(task, strict=True)
    except DelegationRejected as exc:
        if not exc.decision.must_not_delegate:
            raise AssertionError("DelegationRejected but must_not_delegate=False")
        return
    raise AssertionError("strict=True on parent §6.3 should raise DelegationRejected")


def check_choose_roles_sub_must_not_delegate_propagates() -> None:
    """sub brief 가 §6.3 위반이면 §6.3 marker 가 haystack 에 들어가
    parent 부터 reject 됨 (delegator 가 sub brief 까지 검사, v0.5.7 spec §6.3).
    """
    task = _fanout_task()
    task["sub_tasks"][1]["brief"] = "backlog update 처리"  # st-2 위반
    decisions = choose_roles(task, strict=False)
    # v0.5.7 spec: sub 의 §6.3 marker 도 parent fan-out 전체 거부 (정합성)
    # orchestrator 가 sub brief 까지 보고 fan-out 결정을 내리는 것이 안전
    if not decisions[0].must_not_delegate:
        raise AssertionError(
            "v0.5.7 spec §6.3: sub brief 의 marker 도 parent reject 야 함 "
            "(sub 가 fan-out 의 일부이므로)"
        )
    if decisions[0].rejected_reason is None or "backlog" not in decisions[0].rejected_reason:
        raise AssertionError(
            f"rejected_reason should mention backlog, got {decisions[0].rejected_reason!r}"
        )


def check_choose_roles_sub_schema_invalid_raises() -> None:
    """sub_task schema 가 잘못되면 ValueError."""
    task = _fanout_task()
    task["sub_tasks"][0] = {"sub_id": "st-1"}  # 필수 필드 누락
    try:
        choose_roles(task)
    except ValueError as exc:
        if "missing required field" not in str(exc):
            raise AssertionError(f"ValueError should mention 'missing required field', got: {exc}")
        return
    raise AssertionError("sub_task with missing fields should raise ValueError")


def check_choose_roles_sub_task_type_invalid_raises() -> None:
    """sub_task 의 task_type 이 enum 밖이면 ValueError."""
    task = _fanout_task()
    task["sub_tasks"][0]["task_type"] = "magic_type"  # 알 수 없음
    try:
        choose_roles(task)
    except ValueError as exc:
        if "task_type" not in str(exc):
            raise AssertionError(f"ValueError should mention 'task_type', got: {exc}")
        return
    raise AssertionError("invalid sub_task_type should raise ValueError")


def check_validate_fanin_ok() -> None:
    """baseline: ok sub_results → validate PASS."""
    payload = _ok_payload()
    result = validate_fanin_output(payload, expected_parent_delegation_id="del-2026-06-08-fanout-pilot")
    if not result.is_valid:
        raise AssertionError(f"baseline ok should validate, errors: {result.errors}")


def check_validate_fanin_partial_aggregation() -> None:
    """partial sub 1개 → aggregated status 'partial'."""
    payload = _ok_payload(sub_statuses=["ok", "partial", "ok"])
    result = validate_fanin_output(payload, expected_parent_delegation_id="del-2026-06-08-fanout-pilot")
    if not result.is_valid:
        raise AssertionError(f"partial aggregation should validate, errors: {result.errors}")


def check_validate_fanin_failed_aggregation() -> None:
    """failed sub 1개 → aggregated status 'failed'."""
    payload = _ok_payload(sub_statuses=["ok", "failed", "ok"])
    result = validate_fanin_output(payload, expected_parent_delegation_id="del-2026-06-08-fanout-pilot")
    if not result.is_valid:
        raise AssertionError(f"failed aggregation should validate, errors: {result.errors}")


def check_validate_fanin_status_mismatch() -> None:
    """declared status 가 aggregated 와 불일치 → error."""
    payload = _ok_payload(sub_statuses=["ok", "failed", "ok"])
    payload["result"]["status"] = "ok"  # 잘못된 declared status
    result = validate_fanin_output(payload, expected_parent_delegation_id="del-2026-06-08-fanout-pilot")
    if result.is_valid:
        raise AssertionError("status mismatch should produce error")
    if not any("declared" in e.reason and "aggregate" in e.reason for e in result.errors):
        raise AssertionError(f"expected aggregation mismatch error, got: {result.errors}")


def check_validate_fanin_parent_mismatch() -> None:
    """expected_parent_delegation_id 와 parent_delegation_id 불일치 → error."""
    payload = _ok_payload()
    result = validate_fanin_output(payload, expected_parent_delegation_id="del-WRONG")
    if result.is_valid:
        raise AssertionError("parent mismatch should produce error")
    if not any(e.field == "parent_delegation_id" for e in result.errors):
        raise AssertionError(f"expected parent_delegation_id error, got: {result.errors}")


def check_validate_fanin_sub_delegation_id_prefix_mismatch() -> None:
    """sub_result delegation_id 가 parent prefix 와 불일치 → error."""
    payload = _ok_payload()
    payload["result"]["sub_results"][0]["delegation_id"] = "del-WRONG-st-1"
    result = validate_fanin_output(payload, expected_parent_delegation_id="del-2026-06-08-fanout-pilot")
    if result.is_valid:
        raise AssertionError("sub delegation_id prefix mismatch should produce error")


def check_validate_fanin_missing_sub_field() -> None:
    """sub_result 필수 필드 누락 → error."""
    payload = _ok_payload()
    del payload["result"]["sub_results"][0]["summary"]
    result = validate_fanin_output(payload, expected_parent_delegation_id="del-2026-06-08-fanout-pilot")
    if result.is_valid:
        raise AssertionError("missing sub_result field should produce error")


def check_validate_fanin_missing_parent_field() -> None:
    """sub_results 가 있는데 parent_delegation_id 없으면 error."""
    payload = _ok_payload()
    del payload["parent_delegation_id"]
    result = validate_fanin_output(payload, expected_parent_delegation_id="del-2026-06-08-fanout-pilot")
    if result.is_valid:
        raise AssertionError("missing parent_delegation_id with sub_results should produce error")


def check_validate_fanin_artifact_action_and_stats() -> None:
    """artifacts[].action + added/removed/total 검증 (§5 + v0.5.7 P2)."""
    payload = _ok_payload()
    result = validate_fanin_output(payload)
    if not result.is_valid:
        raise AssertionError(f"baseline with action/stats should validate, errors: {result.errors}")


def check_validate_fanin_invalid_action() -> None:
    """artifacts[].action 이 enum 밖이면 error."""
    payload = _ok_payload()
    payload["result"]["artifacts"][0]["action"] = "magic"
    result = validate_fanin_output(payload)
    if result.is_valid:
        raise AssertionError("invalid action enum should produce error")


def check_validate_fanin_invalid_stat_type() -> None:
    """artifacts[].added 가 int 아니면 error."""
    payload = _ok_payload()
    payload["result"]["artifacts"][0]["added"] = "three"  # str
    result = validate_fanin_output(payload)
    if result.is_valid:
        raise AssertionError("non-int added should produce error")


def check_enforce_fanin_sub_delegation_id_prefix_mismatch_raises_value_error() -> None:
    """v0.5.11 §6.5: enforce_fanin_response 가 sub.delegation_id prefix mismatch 시 ValueError raise."""
    payload = _ok_payload()
    payload["result"]["sub_results"][0]["delegation_id"] = "del-WRONG-PREFIX-aaaaaaaa"
    try:
        enforce_fanin_response(
            payload, expected_parent_delegation_id="del-2026-06-08-fanout-pilot"
        )
    except ValueError as exc:
        if "delegation_id" not in str(exc) and "parent" not in str(exc):
            raise AssertionError(
                f"ValueError should mention delegation_id/parent, got: {exc}"
            )
        return
    raise AssertionError(
        "sub.delegation_id prefix mismatch should raise ValueError"
    )


def check_validate_fanin_parent_delegation_id_missing_raises() -> None:
    """v0.5.11 §2 회귀: fan-in payload 에 parent_delegation_id 누락 시 ValueError raise."""
    payload = _ok_payload()
    del payload["parent_delegation_id"]
    result = validate_fanin_output(payload)
    if result.is_valid:
        raise AssertionError(
            "missing parent_delegation_id should produce error"
        )


def check_validate_fanin_sub_id_duplicate_with_status_inconsistency_detected() -> None:
    """v0.5.11 §2 회귀: 동일 sub_id 가 status 불일치 동반 시 aggregated mismatch 로 검출.

    Known limitation (시나리오 B): caller 가 동일 sub_id + 동일 status 의
    중복을 보내면 aggregated check 가 통과함. 옵션 (b) — choose_roles
    dedup check — 는 v0.5.12+ 별도 plan.
    """
    payload = _ok_payload()
    payload["result"]["sub_results"][0]["sub_id"] = "st-dup"
    payload["result"]["sub_results"][1]["sub_id"] = "st-dup"
    payload["result"]["sub_results"][0]["status"] = "ok"
    payload["result"]["sub_results"][1]["status"] = "failed"
    result = validate_fanin_output(payload)
    if result.is_valid:
        raise AssertionError(
            "duplicate sub_id with status inconsistency should be detected"
        )


def main() -> int:
    # choose_roles
    check_choose_roles_single_no_subtasks()
    check_choose_roles_fanout_three_subs()
    check_choose_roles_parent_must_not_delegate_propagates()
    check_choose_roles_parent_must_not_delegate_strict_raises()
    check_choose_roles_sub_must_not_delegate_propagates()
    check_choose_roles_sub_schema_invalid_raises()
    check_choose_roles_sub_task_type_invalid_raises()
    # validate_fanin_output
    check_validate_fanin_ok()
    check_validate_fanin_partial_aggregation()
    check_validate_fanin_failed_aggregation()
    check_validate_fanin_status_mismatch()
    check_validate_fanin_parent_mismatch()
    check_validate_fanin_sub_delegation_id_prefix_mismatch()
    check_validate_fanin_missing_sub_field()
    check_validate_fanin_missing_parent_field()
    check_validate_fanin_artifact_action_and_stats()
    check_validate_fanin_invalid_action()
    check_validate_fanin_invalid_stat_type()
    # v0.5.11 신규 — §6.5 Mavis engine hook (fan-in variant)
    check_enforce_fanin_sub_delegation_id_prefix_mismatch_raises_value_error()
    # v0.5.11 신규 — §2 회귀 test 강화
    check_validate_fanin_parent_delegation_id_missing_raises()
    check_validate_fanin_sub_id_duplicate_with_status_inconsistency_detected()
    print(
        "Contract v1 §4.2/§5.2 multi-component smoke check passed "
        "(choose_roles: single/3-sub/parent-reject/strict/sub-reject/schema-invalid/type-invalid; "
        "validate_fanin_output: ok/partial/failed aggregation, status mismatch, parent mismatch, "
        "sub delegation_id prefix mismatch, missing sub field, missing parent, "
        "artifact action+stats, invalid action, invalid stat type, "
        "enforce_fanin_response prefix mismatch, "
        "parent_delegation_id missing raise, sub_id duplicate status-inconsistency)."
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
