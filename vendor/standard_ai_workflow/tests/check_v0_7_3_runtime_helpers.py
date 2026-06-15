"""Smoke test: v0.7.3 runtime helpers — auth / testing / profiling / resiliency 4종 dispatcher.

각 helper 의 evaluate_compliance() 가 6+6+6+8 = 26 RuleResult 를 반환하는지 검증.

Test count: 12 (3 per helper × 4 helper)
"""

from __future__ import annotations

import sys
from pathlib import Path

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT / "workflow-source"))


# --- Auth (6 SEC-AUTH rule) ---


def test_auth_returns_6_rules():
    from workflow_kit.common.auth import evaluate_compliance
    result = evaluate_compliance(project_root=PROJECT_ROOT)
    assert result["baseline"] == "security-auth"
    assert len(result["results"]) == 6, f"expected 6, got {len(result['results'])}"


def test_auth_rule_id_format():
    from workflow_kit.common.auth import evaluate_compliance
    result = evaluate_compliance(project_root=PROJECT_ROOT)
    for r in result["results"]:
        assert r["rule_id"].startswith("SEC-AUTH-"), f"bad rule_id: {r['rule_id']}"
        assert len(r["rule_id"]) == len("SEC-AUTH-NN"), f"rule_id format: {r['rule_id']}"


def test_auth_dispatcher_via_baselines():
    from workflow_kit.common.contracts.baselines import evaluate_compliance
    summary = evaluate_compliance(PROJECT_ROOT, "security-auth")
    assert summary.baseline == "security-auth"
    assert len(summary.results) == 6
    assert summary.status in ("compliant", "non_compliant", "advisory")


# --- Testing PBT (6 PBT-WF rule) ---


def test_pbt_returns_6_rules():
    from workflow_kit.common.testing import evaluate_compliance
    result = evaluate_compliance(project_root=PROJECT_ROOT)
    assert result["baseline"] == "testing-property-based"
    assert len(result["results"]) == 6, f"expected 6, got {len(result['results'])}"


def test_pbt_rule_id_format():
    from workflow_kit.common.testing import evaluate_compliance
    result = evaluate_compliance(project_root=PROJECT_ROOT)
    for r in result["results"]:
        assert r["rule_id"].startswith("PBT-WF-"), f"bad rule_id: {r['rule_id']}"
        assert len(r["rule_id"]) == len("PBT-WF-NN"), f"rule_id format: {r['rule_id']}"


def test_pbt_dispatcher_via_baselines():
    from workflow_kit.common.contracts.baselines import evaluate_compliance
    summary = evaluate_compliance(PROJECT_ROOT, "testing-property-based")
    assert summary.baseline == "testing-property-based"
    assert len(summary.results) == 6


# --- Profiling (6 PERF-MEM rule) ---


def test_profiling_returns_6_rules():
    from workflow_kit.common.profiling import evaluate_compliance
    result = evaluate_compliance(fn=None)  # N/A 모드
    assert result["baseline"] == "performance-memory"
    assert len(result["results"]) == 6


def test_profiling_with_callable():
    from workflow_kit.common.profiling import evaluate_compliance
    result = evaluate_compliance(fn=lambda: sum(range(1000)))
    assert len(result["results"]) == 6
    # peak memory rule 은 callable 이면 compliant 또는 advisory
    peak = [r for r in result["results"] if r["rule_id"] == "PERF-MEM-01"][0]
    assert peak["status"] in ("compliant", "advisory", "non_compliant")


def test_profiling_dispatcher_via_baselines():
    from workflow_kit.common.contracts.baselines import evaluate_compliance
    summary = evaluate_compliance(PROJECT_ROOT, "performance-memory")
    assert summary.baseline == "performance-memory"
    assert len(summary.results) == 6


# --- Resiliency (8 RES-WF rule) ---


def test_resiliency_returns_8_rules():
    from workflow_kit.common.resiliency import evaluate_compliance
    result = evaluate_compliance(project_root=PROJECT_ROOT)
    assert result["baseline"] == "resiliency"
    assert len(result["results"]) == 8, f"expected 8, got {len(result['results'])}"


def test_resiliency_rule_id_format():
    from workflow_kit.common.resiliency import evaluate_compliance
    result = evaluate_compliance(project_root=PROJECT_ROOT)
    for r in result["results"]:
        assert r["rule_id"].startswith("RES-WF-"), f"bad rule_id: {r['rule_id']}"


def test_resiliency_dispatcher_via_baselines():
    from workflow_kit.common.contracts.baselines import evaluate_compliance
    summary = evaluate_compliance(PROJECT_ROOT, "resiliency")
    assert summary.baseline == "resiliency"
    assert len(summary.results) == 8
    # overall status: 1+ advisory → advisory 가 일반적
    assert summary.status in ("compliant", "advisory", "non_compliant")


if __name__ == "__main__":
    tests = [v for k, v in globals().items() if k.startswith("test_") and callable(v)]
    print(f"running {len(tests)} tests")
    passed = 0
    for t in tests:
        try:
            t()
            print(f"  PASS  {t.__name__}")
            passed += 1
        except Exception as e:
            print(f"  FAIL  {t.__name__}: {e}")
    print(f"\n{passed}/{len(tests)} tests passed")
    sys.exit(0 if passed == len(tests) else 1)
