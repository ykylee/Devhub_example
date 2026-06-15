"""v0.7.8+ state-aware evaluate_compliance + config integration smoke test.

baselines.py 의 7 _eval_* / 2 public evaluate_compliance / evaluate_all 에 `state` keyword
arg 추가. doctor.py 의 config.partial_rules / opt_in *actual apply* 검증.

Test 구성 (8 test):
1. evaluate_compliance(state=...) — keyword arg 정상 (state=None = backward compat)
2. evaluate_all(state=...) — keyword arg 정상
3. _eval_resiliency_baseline(state=...) — merged state 의 partial_rules 적용
4. doctor.py actual apply: config.partial_rules → state merge → evaluate_compliance
5. doctor.py actual apply: config.opt_in → state[f"{baseline}_baseline"]["status"]="enabled"
6. backward compat: state=None 으로 호출 시 _read_state_json 자동 호출
7. signature 변경의 breaking change 0: positional caller (fn, baseline_path) 정상
8. config.partial_rules[baseline] + opt_in 동시 적용 — union

Reference:
- workflow_kit/common/contracts/baselines.py (v0.7.8 state-aware variant)
- workflow_kit/cli/doctor.py (v0.7.8 config actual apply)
- v0.7.6 metadata (load_config / DoctorConfig) 의 1차 source
"""

from __future__ import annotations

import json
import re
import subprocess
import sys
import tempfile
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]


def _import_baselines():
    """workflow_kit.common.contracts.baselines 를 importlib 로 로드."""
    import importlib.util
    import sys
    if str(SOURCE_ROOT) not in sys.path:
        sys.path.insert(0, str(SOURCE_ROOT))
    spec = importlib.util.spec_from_file_location(
        "baselines", str(SOURCE_ROOT / "workflow_kit/common/contracts/baselines.py")
    )
    mod = importlib.util.module_from_spec(spec)
    sys.modules["baselines"] = mod  # 3.14 dataclass 호환
    spec.loader.exec_module(mod)
    return mod


def _import_metadata():
    import importlib.util
    import sys
    if str(SOURCE_ROOT) not in sys.path:
        sys.path.insert(0, str(SOURCE_ROOT))
    spec = importlib.util.spec_from_file_location(
        "metadata", str(SOURCE_ROOT / "workflow_kit/common/metadata.py")
    )
    mod = importlib.util.module_from_spec(spec)
    sys.modules["metadata"] = mod
    spec.loader.exec_module(mod)
    return mod


# --- Test 1: evaluate_compliance state keyword arg ---


def test_evaluate_compliance_state_kwarg() -> None:
    """evaluate_compliance(state=...) keyword arg 정상."""
    bl = _import_baselines()
    # state 없으면 _read_state_json 자동 호출 (backward compat)
    cs_no_state = bl.evaluate_compliance(SOURCE_ROOT, "resiliency")
    assert cs_no_state.baseline == "resiliency"

    # state 명시 — partial_rules 가 state 에 있으면 반영
    explicit_state = {"resiliency_baseline": {"partial_rules": ["RES-WF-01"]}}
    cs_with_state = bl.evaluate_compliance(
        SOURCE_ROOT, "resiliency", state=explicit_state,
    )
    assert cs_with_state.partial_rules == ["RES-WF-01"]


# --- Test 2: evaluate_all state keyword arg ---


def test_evaluate_all_state_kwarg() -> None:
    """evaluate_all(state=...) keyword arg 정상."""
    bl = _import_baselines()
    explicit_state = {
        "security_baseline": {"partial_rules": ["SEC-WF-01"]},
        "testing_baseline": {"partial_rules": []},
    }
    summaries = bl.evaluate_all(SOURCE_ROOT, state=explicit_state)
    assert len(summaries) == 7
    assert summaries["security"].partial_rules == ["SEC-WF-01"]


# --- Test 3: _eval_resiliency_baseline state override ---


def test_eval_resiliency_state_override() -> None:
    """_eval_resiliency_baseline(state=...) 의 partial_rules 반영."""
    bl = _import_baselines()
    explicit_state = {"resiliency_baseline": {"partial_rules": ["RES-WF-01", "RES-WF-02"]}}
    cs = bl._eval_resiliency_baseline(SOURCE_ROOT, state=explicit_state)
    assert cs.partial_rules == ["RES-WF-01", "RES-WF-02"]


# --- Test 4: backward compat (state=None) ---


def test_backward_compat_state_none() -> None:
    """state=None (default) — _read_state_json 자동 호출, 기존 caller 영향 0."""
    bl = _import_baselines()
    # positional caller (state 명시 안 함) — keyword-only arg default None
    cs = bl.evaluate_compliance(SOURCE_ROOT, "resiliency")
    assert cs.baseline == "resiliency"
    # state.json 부재 시 partial_rules = [] (default empty state)
    # workflow-source/ 에는 state.json 부재 — _read_state_json 이 {} 반환 → partial = []


# --- Test 5: positional caller (fn, baseline_path) 정상 ---


def test_positional_caller_breaking_change_0() -> None:
    """evaluate_compliance 의 positional caller (fn, baseline_path) 영향 0.

    v0.7.8+ 가 state 를 keyword-only 로 추가했어도 positional caller (fn=..., baseline_path=...)
    는 정상 작동. breaking change 0.
    """
    bl = _import_baselines()
    # positional: project_root, baseline 만. fn / baseline_path / state 모두 default
    cs = bl.evaluate_compliance(SOURCE_ROOT, "resiliency")
    assert cs.baseline == "resiliency"
    # performance-memory: fn, baseline_path positional
    cs_pm = bl.evaluate_compliance(SOURCE_ROOT, "performance-memory")
    assert cs_pm.baseline == "performance-memory"


# --- Test 6: doctor.py actual apply (config.partial_rules) ---


def test_doctor_config_partial_rules_applied() -> None:
    """doctor.py 의 config.partial_rules 가 state 에 merge → evaluate_compliance 에 actual apply."""
    with tempfile.TemporaryDirectory() as tmp:
        Path(tmp, "pyproject.toml").write_text(
            '[tool.workflow-doctor]\n'
            'partial_rules = { resiliency = ["RES-WF-01", "RES-WF-02"] }\n'
        )
        proc = subprocess.run(
            [sys.executable, "-m", "workflow_kit.cli.doctor",
             "--project-root", tmp, "--baseline=resiliency", "--json"],
            capture_output=True, text=True, timeout=60,
        )
        assert proc.returncode == 0, f"exit {proc.returncode}: {proc.stderr}"
        out = json.loads(proc.stdout)
        # actual apply — cs.partial_rules 에 config.partial_rules 반영
        assert out["results"]["resiliency"]["partial_rules"] == ["RES-WF-01", "RES-WF-02"]


# --- Test 7: doctor.py actual apply (config.opt_in) ---


def test_doctor_config_opt_in_applied() -> None:
    """doctor.py 의 config.opt_in 가 state[baseline_baseline].status="enabled" 로 merge."""
    with tempfile.TemporaryDirectory() as tmp:
        Path(tmp, "pyproject.toml").write_text(
            '[tool.workflow-doctor]\n'
            'opt_in = { "security-auth" = ["SEC-AUTH-04"] }\n'
        )
        proc = subprocess.run(
            [sys.executable, "-m", "workflow_kit.cli.doctor",
             "--project-root", tmp, "--baseline=security-auth", "--json"],
            capture_output=True, text=True, timeout=60,
        )
        assert proc.returncode == 0
        out = json.loads(proc.stdout)
        # opt_in rule 도 partial_rules 에 union (state partial mode hard constraint)
        partial = out["results"]["security-auth"]["partial_rules"]
        assert "SEC-AUTH-04" in partial


# --- Test 8: config.partial_rules + opt_in 동시 (union) ---


def test_doctor_config_partial_rules_and_opt_in_union() -> None:
    """config.partial_rules + opt_in 동시 — partial_rules union (dedup)."""
    with tempfile.TemporaryDirectory() as tmp:
        Path(tmp, "pyproject.toml").write_text(
            '[tool.workflow-doctor]\n'
            'partial_rules = { resiliency = ["RES-WF-01"] }\n'
            'opt_in = { resiliency = ["RES-WF-02"] }\n'
        )
        proc = subprocess.run(
            [sys.executable, "-m", "workflow_kit.cli.doctor",
             "--project-root", tmp, "--baseline=resiliency", "--json"],
            capture_output=True, text=True, timeout=60,
        )
        assert proc.returncode == 0
        out = json.loads(proc.stdout)
        # union (order: partial_rules 먼저, opt_in 추가)
        partial = out["results"]["resiliency"]["partial_rules"]
        assert "RES-WF-01" in partial
        assert "RES-WF-02" in partial


# --- 메인 실행 ---


def main() -> int:
    test_funcs = [
        test_evaluate_compliance_state_kwarg,
        test_evaluate_all_state_kwarg,
        test_eval_resiliency_state_override,
        test_backward_compat_state_none,
        test_positional_caller_breaking_change_0,
        test_doctor_config_partial_rules_applied,
        test_doctor_config_opt_in_applied,
        test_doctor_config_partial_rules_and_opt_in_union,
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
