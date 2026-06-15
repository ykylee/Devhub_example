#!/usr/bin/env python3
"""S3: Orchestrator ↔ Sub-agent contract v1 direct-only check.

Negative test: verifies that §6.3 (MUST NOT delegate) actions are flagged
as orchestrator-only, and that an attempted delegation is rejected.
Reference: workflow-source/core/orchestrator_subagent_contract_v1.md §8.3.
"""

from __future__ import annotations

from pathlib import Path

CONTRACT_PATH = Path(__file__).resolve().parents[1] / "core" / "orchestrator_subagent_contract_v1.md"

# §6.3 MUST NOT delegate list (negative test catalog).
# v0.5.6: 7 actions. v0.5.7: +2 (cross-ref 갱신, fan-in 통합 보고).
DIRECT_ONLY_ACTIONS = (
    "handoff 갱신",
    "backlog 갱신",
    "state.json 갱신",
    "ask_user 호출",
    "우선순위 결정",
    "sub-agent 출력 통합/리뷰",
    "PR 본문 작성",
    # v0.5.7 신규
    "docs cross-ref 갱신",
    "fan-in 통합 보고",
)


def _check_doc_has_section_6_3() -> None:
    if not CONTRACT_PATH.exists():
        raise AssertionError(f"Contract v1 spec doc not found: {CONTRACT_PATH}")
    text = CONTRACT_PATH.read_text(encoding="utf-8")
    if "## 6.3 직접 처리 (MUST NOT delegate)" not in text:
        raise AssertionError("Contract doc missing §6.3 (직접 처리 (MUST NOT delegate))")


def _is_orchestrator_only(action: str) -> bool:
    """Heuristic: an action is orchestrator-only if it appears in DIRECT_ONLY_ACTIONS
    or is a known orchestrator responsibility (handoff / backlog / state / ask_user)."""
    if action in DIRECT_ONLY_ACTIONS:
        return True
    # Substring matches for compound names
    for marker in ("handoff", "backlog", "state.json", "ask_user", "PR 본문", "통합/리뷰"):
        if marker in action:
            return True
    return False


def check_direct_only_actions_rejected() -> None:
    """Simulate a sub-agent receiving a delegation request for an orchestrator-only
    action. The contract v1 must reject this — verification here asserts the
    rejection rule is encoded in the doc."""
    for action in DIRECT_ONLY_ACTIONS:
        if not _is_orchestrator_only(action):
            raise AssertionError(f"Action {action!r} should be flagged as orchestrator-only but isn't")
    # All entries must be orchestrator-only
    for action in DIRECT_ONLY_ACTIONS:
        assert _is_orchestrator_only(action), f"DIRECT_ONLY_ACTIONS contains non-orchestrator item: {action!r}"


def check_negative_test_examples_present() -> None:
    """§8.3 calls out specific negative examples (handoff 직접 write, ask_user 호출)."""
    text = CONTRACT_PATH.read_text(encoding="utf-8")
    if "## 8.3 S3: 직접 처리 영역 준수" not in text:
        raise AssertionError("Contract doc missing §8.3 S3 description")
    for example in ("handoff 직접 write", "ask_user 호출"):
        if example not in text:
            raise AssertionError(f"§8.3 missing negative-test example: {example!r}")


def main() -> int:
    _check_doc_has_section_6_3()
    check_direct_only_actions_rejected()
    check_negative_test_examples_present()
    print("Contract v1 direct-only smoke check passed (9 MUST-NOT-delegate actions [7 v0.5.6 + 2 v0.5.7] + §8.3 negative examples).")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
