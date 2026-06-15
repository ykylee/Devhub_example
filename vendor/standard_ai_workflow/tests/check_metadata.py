"""workflow_kit.common.metadata 의 smoke test (v0.7.6+).

[tool.workflow-doctor] section 의 pyproject.toml metadata loader 검증.
- load_config: section 부재 / file 부재 / invalid TOML 모두 default fallback
- DoctorConfig: 5 field (partial_rules / opt_in / thresholds / excluded_paths / fail_on)
- should_fail: severity 비교 (non_compliant > advisory > compliant > not_applicable)

Test 구성 (8 test):
1. load_config: section 부재 시 default fallback
2. load_config: [tool.workflow-doctor] 정상 parse
3. load_config: invalid TOML (truncated) 시 default fallback
4. load_config: partial type mismatch (int list) 시 default fallback
5. DoctorConfig.to_dict: JSON-serializable
6. should_fail: non_compliant + fail_on=non_compliant → True
7. should_fail: advisory + fail_on=non_compliant → False
8. should_fail: not_applicable → False (severity 0)

Reference:
- workflow_kit/common/metadata.py 본체
- workflow-source/pyproject.toml 의 [tool.workflow-doctor] section (v0.7.6+)
- workflow_kit.common.contracts.baselines (rule spec)
- workflow_kit.cli.doctor (v0.7.4+)
"""

from __future__ import annotations

import json
import re
import subprocess
import sys
import tempfile
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
KIT = SOURCE_ROOT / "workflow_kit"


def _import_metadata():
    """workflow_kit.common.metadata 를 importlib 로 로드."""
    import importlib.util
    spec = importlib.util.spec_from_file_location("metadata", str(KIT / "common/metadata.py"))
    mod = importlib.util.module_from_spec(spec)
    import sys
    sys.modules["metadata"] = mod
    spec.loader.exec_module(mod)
    return mod


# --- Test 1: default fallback ---


def test_load_config_no_section() -> None:
    """[tool.workflow-doctor] section 부재 시 default."""
    mod = _import_metadata()
    with tempfile.TemporaryDirectory() as tmp:
        # pyproject.toml 자체 부재
        config = mod.load_config(tmp)
        assert config.partial_rules == {}
        assert config.opt_in == {}
        assert config.thresholds == {"score_alert": 0.3, "memory_alert_mb": 100.0}
        assert config.fail_on == "non_compliant"
        # excluded_paths default
        assert "build/*" in config.excluded_paths


def test_load_config_no_section_but_file() -> None:
    """pyproject.toml 은 있지만 [tool.workflow-doctor] 없으면 default."""
    mod = _import_metadata()
    with tempfile.TemporaryDirectory() as tmp:
        p = Path(tmp) / "pyproject.toml"
        p.write_text("[project]\nname = 'x'\nversion = '0.0.1'\n")
        config = mod.load_config(tmp)
        assert config.partial_rules == {}
        assert config.fail_on == "non_compliant"


# --- Test 2: 정상 parse ---


def test_load_config_full_section() -> None:
    """[tool.workflow-doctor] 5 field 모두 parse."""
    mod = _import_metadata()
    with tempfile.TemporaryDirectory() as tmp:
        p = Path(tmp) / "pyproject.toml"
        p.write_text("""[tool.workflow-doctor]
partial_rules = { resiliency = ["RES-WF-01", "RES-WF-02"] }
opt_in = { "security-auth" = ["SEC-AUTH-04"] }
thresholds = { score_alert = 0.5, memory_alert_mb = 200.0 }
excluded_paths = ["build/*", ".venv/*", "tests/fixtures/*"]
fail_on = "advisory"
""")
        config = mod.load_config(tmp)
        assert config.partial_rules == {"resiliency": ["RES-WF-01", "RES-WF-02"]}
        assert config.opt_in == {"security-auth": ["SEC-AUTH-04"]}
        assert config.thresholds == {"score_alert": 0.5, "memory_alert_mb": 200.0}
        assert config.excluded_paths == ["build/*", ".venv/*", "tests/fixtures/*"]
        assert config.fail_on == "advisory"


# --- Test 3: invalid TOML ---


def test_load_config_invalid_toml() -> None:
    """truncated / broken TOML 시 default fallback (실패 ❌)."""
    mod = _import_metadata()
    with tempfile.TemporaryDirectory() as tmp:
        p = Path(tmp) / "pyproject.toml"
        p.write_text("[tool.workflow-doctor\npartial_rules = { broken")  # missing close
        config = mod.load_config(tmp)
        # invalid 이면 default
        assert config.partial_rules == {}
        assert config.fail_on == "non_compliant"


# --- Test 4: type mismatch ---


def test_load_config_type_mismatch() -> None:
    """partial_rules 가 dict 가 아니면 default fallback."""
    mod = _import_metadata()
    with tempfile.TemporaryDirectory() as tmp:
        p = Path(tmp) / "pyproject.toml"
        p.write_text("""[tool.workflow-doctor]
partial_rules = "not_a_dict"
fail_on = "unknown_value"
""")
        config = mod.load_config(tmp)
        assert config.partial_rules == {}  # type mismatch → default
        assert config.fail_on == "non_compliant"  # invalid enum → default


# --- Test 5: DoctorConfig.to_dict ---


def test_doctor_config_to_dict() -> None:
    """to_dict() 가 JSON-serializable."""
    mod = _import_metadata()
    config = mod.DoctorConfig(
        partial_rules={"security": ["SEC-WF-01"]},
        opt_in={"security-auth": ["SEC-AUTH-04"]},
        thresholds={"score_alert": 0.3},
        excluded_paths=["build/*"],
        fail_on="advisory",
    )
    d = config.to_dict()
    # JSON serialize 가능 검증
    s = json.dumps(d, ensure_ascii=False)
    parsed = json.loads(s)
    assert parsed["partial_rules"] == {"security": ["SEC-WF-01"]}
    assert parsed["fail_on"] == "advisory"


# --- Test 6/7/8: should_fail severity ---


def test_should_fail_non_compliant() -> None:
    """non_compliant + fail_on=non_compliant → True."""
    mod = _import_metadata()
    config = mod.DoctorConfig(fail_on="non_compliant")
    assert mod.should_fail("non_compliant", config) is True


def test_should_fail_advisory_below_threshold() -> None:
    """advisory + fail_on=non_compliant → False (advisory 는 non_compliant 미만)."""
    mod = _import_metadata()
    config = mod.DoctorConfig(fail_on="non_compliant")
    assert mod.should_fail("advisory", config) is False


def test_should_fail_advisory_above_threshold() -> None:
    """advisory + fail_on=advisory → True."""
    mod = _import_metadata()
    config = mod.DoctorConfig(fail_on="advisory")
    assert mod.should_fail("advisory", config) is True
    assert mod.should_fail("non_compliant", config) is True
    # compliant 는 advisory 미만 → False
    assert mod.should_fail("compliant", config) is False


def test_should_fail_not_applicable() -> None:
    """not_applicable 은 severity 0, 어떤 fail_on 이든 False (skip 의미)."""
    mod = _import_metadata()
    for fail_on in ("compliant", "advisory", "non_compliant"):
        config = mod.DoctorConfig(fail_on=fail_on)
        assert mod.should_fail("not_applicable", config) is False, f"fail_on={fail_on}"


# --- 메인 실행 ---


def main() -> int:
    test_funcs = [
        test_load_config_no_section,
        test_load_config_no_section_but_file,
        test_load_config_full_section,
        test_load_config_invalid_toml,
        test_load_config_type_mismatch,
        test_doctor_config_to_dict,
        test_should_fail_non_compliant,
        test_should_fail_advisory_below_threshold,
        test_should_fail_advisory_above_threshold,
        test_should_fail_not_applicable,
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
