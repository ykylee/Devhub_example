#!/usr/bin/env python3
"""v0.7.1: Extension Baseline Compliance Evaluator 검증.

- workflow_kit.common.contracts.baselines.evaluate_compliance() 3종 baseline
- evaluate_all() 통합
- ComplianceSummary dataclass + to_dict()
- Security/Testing/Performance 6 rule × 3 = 18 RuleResult
- _aggregate_status 분기: compliant / non_compliant / advisory
- state.json partial_rules 적용
- N/A 처리: import 실패 또는 spec 없음

Reference: workflow-source/extensions/SCHEMA.md §6 Helper Contract
           workflow-source/workflow_kit/common/contracts/baselines.py
"""

from __future__ import annotations

import sys
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SOURCE_ROOT))

BASELINES_MODULE = "workflow_kit.common.contracts.baselines"


# --- Test 1: module import ---


def test_baselines_module_importable() -> None:
    """workflow_kit.common.contracts.baselines import 가능."""
    import importlib
    mod = importlib.import_module(BASELINES_MODULE)
    assert hasattr(mod, "evaluate_compliance"), "evaluate_compliance missing"
    assert hasattr(mod, "evaluate_all"), "evaluate_all missing"


def test_dataclass_classes() -> None:
    """RuleResult + ComplianceSummary dataclass 정의."""
    from workflow_kit.common.contracts.baselines import RuleResult, ComplianceSummary
    # dataclass 인스턴스 생성 가능
    rr = RuleResult(rule_id="TEST-01", title="Test", status="compliant")
    cs = ComplianceSummary(baseline="test", status="compliant")
    assert rr.rule_id == "TEST-01"
    assert cs.baseline == "test"


# --- Test 2: 3종 baseline 평가 ---


def test_evaluate_security_baseline() -> None:
    """evaluate_compliance(baseline='security') → ComplianceSummary."""
    from workflow_kit.common.contracts.baselines import evaluate_compliance
    cs = evaluate_compliance(SOURCE_ROOT, "security")
    assert cs.baseline == "security"
    assert len(cs.results) == 6, f"expected 6 rules, got {len(cs.results)}"
    rule_ids = {r.rule_id for r in cs.results}
    expected = {f"SEC-WF-{i:02d}" for i in range(1, 7)}
    assert rule_ids == expected, f"rule ID mismatch: {rule_ids} != {expected}"


def test_evaluate_testing_baseline() -> None:
    """evaluate_compliance(baseline='testing') → ComplianceSummary."""
    from workflow_kit.common.contracts.baselines import evaluate_compliance
    cs = evaluate_compliance(SOURCE_ROOT, "testing")
    assert cs.baseline == "testing"
    assert len(cs.results) == 6, f"expected 6 rules, got {len(cs.results)}"
    rule_ids = {r.rule_id for r in cs.results}
    expected = {f"TST-WF-{i:02d}" for i in range(1, 7)}
    assert rule_ids == expected, f"rule ID mismatch: {rule_ids} != {expected}"


def test_evaluate_performance_baseline() -> None:
    """evaluate_compliance(baseline='performance') → ComplianceSummary."""
    from workflow_kit.common.contracts.baselines import evaluate_compliance
    cs = evaluate_compliance(SOURCE_ROOT, "performance")
    assert cs.baseline == "performance"
    assert len(cs.results) == 6, f"expected 6 rules, got {len(cs.results)}"
    rule_ids = {r.rule_id for r in cs.results}
    expected = {f"PERF-WF-{i:02d}" for i in range(1, 7)}
    assert rule_ids == expected, f"rule ID mismatch: {rule_ids} != {expected}"


# --- Test 2.1: 4 신규 sub-cat dispatcher (v0.7.3+, v0.7.5 audit 보강) ---


def test_evaluate_security_auth_baseline() -> None:
    """evaluate_compliance(baseline='security-auth') → 6 SEC-AUTH rule.

    v0.7.5 audit: v0.7.3 의 4 신규 sub-cat dispatcher 중 security-auth 의
    runtime test 누락 보강. SEC-AUTH-01~06 (keyring chmod 600, token rotation,
    OAuth scope, 2FA, entropy, audit log) 검증.
    """
    from workflow_kit.common.contracts.baselines import evaluate_compliance
    cs = evaluate_compliance(SOURCE_ROOT, "security-auth")
    assert cs.baseline == "security-auth"
    assert len(cs.results) == 6, f"expected 6 rules, got {len(cs.results)}"
    rule_ids = {r.rule_id for r in cs.results}
    expected = {f"SEC-AUTH-{i:02d}" for i in range(1, 7)}
    assert rule_ids == expected, f"rule ID mismatch: {rule_ids} != {expected}"


def test_evaluate_testing_pbt_baseline() -> None:
    """evaluate_compliance(baseline='testing-property-based') → 6 PBT-WF rule.

    v0.7.5 audit: testing sub-cat (property-based) 의 runtime dispatcher 검증.
    PBT-WF-01~06 (test_prop_, round_trip, invariant, idempotency, generator, shrink).
    """
    from workflow_kit.common.contracts.baselines import evaluate_compliance
    cs = evaluate_compliance(SOURCE_ROOT, "testing-property-based")
    assert cs.baseline == "testing-property-based"
    assert len(cs.results) == 6, f"expected 6 rules, got {len(cs.results)}"
    rule_ids = {r.rule_id for r in cs.results}
    expected = {f"PBT-WF-{i:02d}" for i in range(1, 7)}
    assert rule_ids == expected, f"rule ID mismatch: {rule_ids} != {expected}"


def test_evaluate_performance_memory_baseline() -> None:
    """evaluate_compliance(baseline='performance-memory') → 6 PERF-MEM rule.

    v0.7.5 audit: performance sub-cat (memory) 의 runtime dispatcher 검증.
    PERF-MEM-01~06 (tracemalloc+RSS, leak, GC pause, ref cycle, cProfile, regression).
    """
    from workflow_kit.common.contracts.baselines import evaluate_compliance
    cs = evaluate_compliance(SOURCE_ROOT, "performance-memory")
    assert cs.baseline == "performance-memory"
    assert len(cs.results) == 6, f"expected 6 rules, got {len(cs.results)}"
    rule_ids = {r.rule_id for r in cs.results}
    expected = {f"PERF-MEM-{i:02d}" for i in range(1, 7)}
    assert rule_ids == expected, f"rule ID mismatch: {rule_ids} != {expected}"


def test_evaluate_resiliency_baseline() -> None:
    """evaluate_compliance(baseline='resiliency') → 8 RES-WF rule.

    v0.7.5 audit: resiliency sub-cat 의 runtime dispatcher 검증.
    RES-WF-01~08 (doctor, structured log, metrics, run_id, 5-tuple error,
    SIGTERM, max iter, health snapshot).
    """
    from workflow_kit.common.contracts.baselines import evaluate_compliance
    cs = evaluate_compliance(SOURCE_ROOT, "resiliency")
    assert cs.baseline == "resiliency"
    assert len(cs.results) == 8, f"expected 8 rules, got {len(cs.results)}"
    rule_ids = {r.rule_id for r in cs.results}
    expected = {f"RES-WF-{i:02d}" for i in range(1, 9)}
    assert rule_ids == expected, f"rule ID mismatch: {rule_ids} != {expected}"


def test_evaluate_all_baselines() -> None:
    """evaluate_all() → 7 baseline 모두 (v0.7.3+ 4 dispatcher 추가)."""
    from workflow_kit.common.contracts.baselines import evaluate_all
    result = evaluate_all(SOURCE_ROOT)
    expected = {
        "security",
        "testing",
        "performance",
        "security-auth",
        "testing-property-based",
        "performance-memory",
        "resiliency",
    }
    assert set(result.keys()) == expected, f"missing: {expected - set(result.keys())}"
    for baseline_name, cs in result.items():
        assert cs.baseline == baseline_name
        # 6 rule (security/testing/performance/security-auth/testing-property-based/performance-memory)
        # 8 rule (resiliency)
        if baseline_name == "resiliency":
            assert len(cs.results) == 8, f"{baseline_name}: {len(cs.results)} rule (expected 8)"
        else:
            assert len(cs.results) == 6, f"{baseline_name}: {len(cs.results)} rule (expected 6)"


# --- Test 3: ComplianceSummary.to_dict() ---


def test_compliance_summary_to_dict() -> None:
    """to_dict() 가 JSON-serializable dict 반환."""
    from workflow_kit.common.contracts.baselines import evaluate_compliance
    cs = evaluate_compliance(SOURCE_ROOT, "security")
    d = cs.to_dict()
    assert isinstance(d, dict)
    assert d["baseline"] == "security"
    assert "results" in d
    assert len(d["results"]) == 6
    for r in d["results"]:
        assert "rule_id" in r
        assert "title" in r
        assert "status" in r
        assert r["status"] in ("compliant", "non_compliant", "not_applicable", "advisory")


# --- Test 4: status enum ---


def test_status_values_valid() -> None:
    """status 가 4 value (compliant/non_compliant/not_applicable/advisory) 중 1."""
    from workflow_kit.common.contracts.baselines import evaluate_all
    valid = {"compliant", "non_compliant", "not_applicable", "advisory"}
    for cs in evaluate_all(SOURCE_ROOT).values():
        for r in cs.results:
            assert r.status in valid, f"{r.rule_id}: invalid status {r.status}"


def test_aggregate_status_branches() -> None:
    """_aggregate_status 의 3 분기 (compliant/non_compliant/advisory)."""
    from workflow_kit.common.contracts.baselines import (
        RuleResult, _aggregate_status, _eval_security_baseline,
    )

    # all compliant → compliant
    r1 = [RuleResult("X-01", "t", "compliant"), RuleResult("X-02", "t", "compliant")]
    assert _aggregate_status(r1, []) == "compliant"

    # 1 non_compliant (in non-partial) → non_compliant
    r2 = [RuleResult("X-01", "t", "compliant"), RuleResult("X-02", "t", "non_compliant")]
    assert _aggregate_status(r2, []) == "non_compliant"

    # partial_rules 가 있어 non_compliant 가 partial list 외 → advisory
    r3 = [RuleResult("X-01", "t", "compliant"), RuleResult("X-02", "t", "non_compliant")]
    assert _aggregate_status(r3, ["X-01"]) == "advisory"


# --- Test 5: state.json partial_rules ---


def test_partial_rules_applied() -> None:
    """state.json 의 partial_rules 가 적용됨."""
    from workflow_kit.common.contracts.baselines import _get_partial_rules, _is_enabled
    state = {
        "security_baseline": {"status": "partial", "partial_rules": ["SEC-WF-01", "SEC-WF-02"]},
        "testing_baseline": {"status": "enabled", "partial_rules": []},
        "performance_baseline": {"status": "disabled", "partial_rules": []},
    }
    assert _get_partial_rules(state, "security") == ["SEC-WF-01", "SEC-WF-02"]
    assert _is_enabled(state, "testing") is True
    assert _is_enabled(state, "performance") is False


# --- Test 6: 6 rule × 7 baseline = 44 RuleResult (v0.7.3+ 7 baseline dispatcher) ---


def test_total_rule_results() -> None:
    """7 baseline × 6+6+6+6+6+6+8 = 44 RuleResult.

    v0.7.1: 3 baseline (security 6 + testing 6 + performance 6) = 18
    v0.7.3: + 4 baseline dispatcher (security-auth 6 + testing-property-based 6 +
                                performance-memory 6 + resiliency 8) = +26
    Total: 18 + 26 = 44
    """
    from workflow_kit.common.contracts.baselines import evaluate_all
    result = evaluate_all(SOURCE_ROOT)
    total = sum(len(cs.results) for cs in result.values())
    assert total == 44, f"expected 44 total (v0.7.3: 3 + 4 baseline dispatcher), got {total}"


def test_seven_baseline_dispatcher_v0_7_3() -> None:
    """v0.7.3+ 7 baseline dispatcher."""
    from workflow_kit.common.contracts.baselines import evaluate_all
    result = evaluate_all(SOURCE_ROOT)
    expected = {
        "security",
        "testing",
        "performance",
        "security-auth",
        "testing-property-based",
        "performance-memory",
        "resiliency",
    }
    assert set(result.keys()) == expected, f"missing: {expected - set(result.keys())}"


# --- Test 7: error handling ---


def test_unknown_baseline_raises() -> None:
    """unknown baseline → ValueError."""
    from workflow_kit.common.contracts.baselines import evaluate_compliance
    try:
        evaluate_compliance(SOURCE_ROOT, "unknown_baseline")
        assert False, "should have raised"
    except ValueError:
        pass


# --- 메인 실행 ---


def main() -> int:
    """모든 test 실행 후 결과 출력."""
    test_funcs = [
        test_baselines_module_importable,
        test_dataclass_classes,
        test_evaluate_security_baseline,
        test_evaluate_testing_baseline,
        test_evaluate_performance_baseline,
        test_evaluate_security_auth_baseline,
        test_evaluate_testing_pbt_baseline,
        test_evaluate_performance_memory_baseline,
        test_evaluate_resiliency_baseline,
        test_evaluate_all_baselines,
        test_compliance_summary_to_dict,
        test_status_values_valid,
        test_aggregate_status_branches,
        test_partial_rules_applied,
        test_total_rule_results,
        test_unknown_baseline_raises,
    ]

    passed = 0
    failed = 0
    failures: list[tuple[str, str]] = []
    for func in test_funcs:
        try:
            func()
            print(f"  PASS  {func.__name__}")
            passed += 1
        except AssertionError as e:
            print(f"  FAIL  {func.__name__}: {e}")
            failed += 1
            failures.append((func.__name__, str(e)))
        except Exception as e:  # noqa: BLE001
            print(f"  ERROR {func.__name__}: {type(e).__name__}: {e}")
            failed += 1
            failures.append((func.__name__, f"{type(e).__name__}: {e}"))

    print()
    if failed == 0:
        print(f"All {passed} tests passed.")
        return 0
    print(f"{failed}/{passed + failed} tests failed:")
    for name, err in failures:
        print(f"  - {name}: {err}")
    return 1


if __name__ == "__main__":
    sys.exit(main())
