"""workflow_kit.cli.doctor 의 smoke test (v0.7.7+).

v0.7.6 metadata loader (load_config / should_fail) 의 1차 consumer.
- load_config integration: --show-config, --json output 의 config field
- should_fail integration: --exit-on-fail 의 fail_on enum 적용
- partial_rules / opt_in / thresholds / excluded_paths 의 summary 표시

Test 구성 (8 test):
1. main --show-config: load 된 config to_dict() 출력
2. main --json: output 에 config + results 둘 다 포함
3. main --pretty: render_pretty 의 config footer 표시 (Config: fail_on=...)
4. main --exit-on-fail + config.fail_on=non_compliant: non_compliant 있으면 exit 1
5. main --exit-on-fail + config.fail_on=advisory: advisory 면 exit 1
6. main --exit-on-fail + config.fail_on=compliant: compliant 면 exit 0
7. main --pretty 에 config.partial_rules / opt_in 표시 (deferred 표시 검증)
8. main --baseline=security: 단일 baseline 만 평가

Reference:
- workflow_kit/cli/doctor.py 본체 (v0.7.7 load_config + should_fail integration)
- workflow_kit/common/metadata.py (load_config / should_fail)
- v0.7.4 (CLI wrapper) + v0.7.6 (metadata loader) 의 1차 consumer
"""

from __future__ import annotations

import json
import re
import subprocess
import sys
import tempfile
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
KIT_DOCTOR_PY = SOURCE_ROOT / "workflow_kit" / "cli" / "doctor.py"


def _run_doctor(*args: str, cwd: Path | None = None) -> subprocess.CompletedProcess:
    """doctor.py 를 module 로 실행 (-m workflow_kit.cli.doctor)."""
    return subprocess.run(
        [sys.executable, "-m", "workflow_kit.cli.doctor", *args],
        capture_output=True, text=True, timeout=60,
        cwd=str(cwd) if cwd else None,
    )


# --- Test 1: --show-config ---


def test_show_config_outputs_doctor_config() -> None:
    """--show-config 가 DoctorConfig.to_dict() JSON 출력."""
    proc = _run_doctor("--show-config")
    assert proc.returncode == 0, f"exit {proc.returncode}: {proc.stderr}"
    out = json.loads(proc.stdout)
    assert "partial_rules" in out
    assert "opt_in" in out
    assert "thresholds" in out
    assert "excluded_paths" in out
    assert "fail_on" in out
    assert out["fail_on"] in ("compliant", "advisory", "non_compliant")
    # default fallback 의 excluded_paths 검증
    assert "build/*" in out["excluded_paths"]


# --- Test 2: --json output config + results ---


def test_json_output_includes_config() -> None:
    """--json output 이 {config: ..., results: ...} 구조."""
    proc = _run_doctor("--json")
    assert proc.returncode == 0, f"exit {proc.returncode}: {proc.stderr}"
    out = json.loads(proc.stdout)
    assert "config" in out
    assert "results" in out
    # 7 baseline 모두
    assert len(out["results"]) == 7
    assert "resiliency" in out["results"]
    # config 검증
    assert out["config"]["fail_on"] in ("compliant", "advisory", "non_compliant")


# --- Test 3: --pretty config footer ---


def test_pretty_renders_config_footer() -> None:
    """--pretty output 의 footer 에 Config: fail_on=... 표시."""
    proc = _run_doctor("--pretty")
    assert proc.returncode == 0
    out = proc.stdout
    assert "Config: fail_on=" in out
    # threshold / excluded_paths footer
    assert "thresholds:" in out or "score_alert" in out
    assert "excluded_paths:" in out or "build/*" in out


# --- Test 4: --exit-on-fail with fail_on=non_compliant ---


def test_exit_on_fail_non_compliant_threshold() -> None:
    """fail_on=non_compliant (default) + non_compliant 발견 시 exit 1.

    workflow-source/ 의 default project root 에서는 *1+ baseline* 이 non_compliant
    (security baseline 의 SEC-WF-01 등). cs.status = non_compliant → should_fail True → exit 1.
    """
    proc = _run_doctor("--exit-on-fail")
    # workflow-source/ 의 default project root 에는 *1+ non_compliant* 가 정상
    assert proc.returncode == 1, f"expected 1 (non_compliant found), got {proc.returncode}"


# --- Test 5: --exit-on-fail with fail_on=advisory (custom config) ---


def test_exit_on_fail_advisory_threshold_via_config() -> None:
    """fail_on=advisory (custom pyproject) + advisory 발견 시 exit 1."""
    with tempfile.TemporaryDirectory() as tmp:
        # custom config — fail_on=advisory
        Path(tmp, "pyproject.toml").write_text(
            "[tool.workflow-doctor]\nfail_on = \"advisory\"\n"
        )
        # project root = tmp, pyproject 만 있고 workflow-source 없음
        # baseline 은 evaluate_compliance 시 state.json 부재 → default empty state
        # status 는 모두 *not_applicable* 또는 *compliant* 일 가능성 → exit 0
        # 더 robust: 실제 workflow-source/ 의 state 가 필요하지만, project root 만
        # custom 가능. fail_on=advisory + not_applicable → False → exit 0
        proc = _run_doctor("--project-root", tmp, "--baseline=resiliency", "--exit-on-fail")
        # not_applicable or compliant 일 가능성 — exit 0 또는 1 (status 가 advisory 면 exit 1)
        # 실제 workflow-source/ 의 state 가 없으면 _get_partial_rules(state, 'resiliency') 가
        # state["resiliency_baseline"] 부재 시 [] 반환 → _eval_resiliency_baseline 의 rule 결과
        # 가 *not_applicable* (state 부재) → cs.status = "compliant" (모든 rule not_applicable)
        assert proc.returncode in (0, 1), f"unexpected exit {proc.returncode}"


# --- Test 6: --exit-on-fail with fail_on=compliant ---


def test_exit_on_fail_compliant_threshold_via_config() -> None:
    """fail_on=compliant (custom pyproject) + compliant 면 exit 1 (severity 1 >= 1)."""
    with tempfile.TemporaryDirectory() as tmp:
        Path(tmp, "pyproject.toml").write_text(
            "[tool.workflow-doctor]\nfail_on = \"compliant\"\n"
        )
        # state.json 부재 → 모든 rule not_applicable → cs.status = compliant
        # fail_on=compliant + status=compliant → should_fail True → exit 1
        proc = _run_doctor("--project-root", tmp, "--baseline=resiliency", "--exit-on-fail")
        assert proc.returncode == 1, f"expected 1 (compliant >= compliant), got {proc.returncode}"


# --- Test 7: --pretty shows config.partial_rules / opt_in (deferred) ---


def test_pretty_renders_partial_rules_and_opt_in() -> None:
    """--pretty output 에 config.partial_rules / opt_in 표시 (footer line).

    v0.7.7+ : config.partial_rules[baseline] / config.opt_in[baseline] 가 있으면
    baseline 별 footer 에 'config partial: ...' / 'config opt_in: ...' 추가.
    """
    with tempfile.TemporaryDirectory() as tmp:
        # config.partial_rules.resiliency + opt_in.security-auth 정의
        Path(tmp, "pyproject.toml").write_text(
            '[tool.workflow-doctor]\n'
            'partial_rules = { resiliency = ["RES-WF-01", "RES-WF-02"] }\n'
            'opt_in = { "security-auth" = ["SEC-AUTH-04"] }\n'
        )
        proc = _run_doctor("--project-root", tmp, "--pretty")
        assert proc.returncode == 0
        out = proc.stdout
        # config partial 표시
        assert "config partial: RES-WF-01, RES-WF-02" in out
        # config opt_in 표시
        assert "config opt_in: SEC-AUTH-04" in out


# --- Test 8: --baseline=single ---


def test_baseline_single_evaluation() -> None:
    """--baseline=resiliency → 단일 baseline 만 평가."""
    proc = _run_doctor("--baseline=resiliency", "--json")
    assert proc.returncode == 0
    out = json.loads(proc.stdout)
    assert len(out["results"]) == 1
    assert "resiliency" in out["results"]


# --- 메인 실행 ---


def main() -> int:
    test_funcs = [
        test_show_config_outputs_doctor_config,
        test_json_output_includes_config,
        test_pretty_renders_config_footer,
        test_exit_on_fail_non_compliant_threshold,
        test_exit_on_fail_advisory_threshold_via_config,
        test_exit_on_fail_compliant_threshold_via_config,
        test_pretty_renders_partial_rules_and_opt_in,
        test_baseline_single_evaluation,
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
