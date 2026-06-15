"""Smoke test: v0.7.4 follow-up — CLI wrapper / decorator / optional dep 3종.

Test count: 9 (3 per category)
"""

from __future__ import annotations

import json
import os
import signal
import subprocess
import sys
import time
from pathlib import Path

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
SOURCE_ROOT = PROJECT_ROOT / "workflow-source"
sys.path.insert(0, str(SOURCE_ROOT))


# --- CLI wrapper (workflow doctor) ---


def test_doctor_pretty_default():
    """workflow doctor — pretty table default."""
    result = subprocess.run(
        ["python3", "-m", "workflow_kit.cli.doctor", "--project-root", str(PROJECT_ROOT)],
        cwd=str(SOURCE_ROOT),
        capture_output=True,
        text=True,
        timeout=30,
    )
    assert result.returncode == 0, f"exit {result.returncode}: {result.stderr}"
    assert "Workflow Doctor" in result.stdout
    assert "Summary:" in result.stdout


def test_doctor_json_output():
    """workflow doctor --json → JSON dict."""
    result = subprocess.run(
        [
            "python3",
            "-m",
            "workflow_kit.cli.doctor",
            "--project-root",
            str(PROJECT_ROOT),
            "--json",
        ],
        cwd=str(SOURCE_ROOT),
        capture_output=True,
        text=True,
        timeout=30,
    )
    assert result.returncode == 0
    data = json.loads(result.stdout)
    # 7 baseline
    expected = {
        "security",
        "testing",
        "performance",
        "security-auth",
        "testing-property-based",
        "performance-memory",
        "resiliency",
    }
    assert set(data.keys()) == expected


def test_doctor_single_baseline():
    """workflow doctor --baseline=resiliency → 단일 baseline."""
    result = subprocess.run(
        [
            "python3",
            "-m",
            "workflow_kit.cli.doctor",
            "--project-root",
            str(PROJECT_ROOT),
            "--baseline",
            "resiliency",
            "--json",
        ],
        cwd=str(SOURCE_ROOT),
        capture_output=True,
        text=True,
        timeout=30,
    )
    assert result.returncode == 0
    data = json.loads(result.stdout)
    assert "resiliency" in data
    assert len(data["resiliency"]["results"]) == 8


# --- @graceful_shutdown decorator ---


def test_graceful_shutdown_decorator_basic():
    """decorator 가 callable 을 wrap 하고 __graceful_shutdown__ marker 가 있는지."""
    from workflow_kit.common.decorators import graceful_shutdown

    @graceful_shutdown
    def my_func():
        return 42

    assert hasattr(my_func, "__graceful_shutdown__")
    assert my_func.__graceful_shutdown__ is True
    # function is still callable
    assert my_func() == 42


def test_graceful_shutdown_with_cleanup():
    """decorator with cleanup callback."""
    from workflow_kit.common.decorators import graceful_shutdown

    calls = []

    @graceful_shutdown(cleanup=lambda: calls.append("cleanup"))
    def my_func():
        calls.append("body")

    my_func()
    assert calls == ["body"]
    # cleanup not called when no signal
    assert "cleanup" not in calls


def test_graceful_shutdown_preserves_metadata():
    """@functools.wraps 가 메타데이터 보존."""
    from workflow_kit.common.decorators import graceful_shutdown

    @graceful_shutdown
    def my_documented_func():
        """test docstring."""
        return 1

    assert my_documented_func.__name__ == "my_documented_func"
    assert my_documented_func.__doc__ == "test docstring."


# --- Optional dependency (hypothesis / objgraph) ---


def test_hypothesis_optional_fallback():
    """testing.py 가 hypothesis 미설치 시 fallback 정상 동작."""
    from workflow_kit.common.testing import check_generator_present
    result = check_generator_present([])
    # empty test_files → N/A
    assert result.status in ("advisory", "not_applicable", "compliant")


def test_objgraph_optional_fallback():
    """profiling.py 가 objgraph 미설치 시 fallback 정상 동작."""
    from workflow_kit.common.profiling import check_reference_cycle
    result = check_reference_cycle(lambda: sum(range(100)))
    # 정상 callable → compliant
    assert result.status in ("compliant", "advisory", "not_applicable")


def test_pyproject_optional_deps_declared():
    """pyproject.toml 에 pbt/profiling optional-deps 가 선언됨."""
    import tomllib

    with open(SOURCE_ROOT / "pyproject.toml", "rb") as f:
        data = tomllib.load(f)
    optional = data["project"]["optional-dependencies"]
    assert "pbt" in optional, "missing pbt optional"
    assert "profiling" in optional, "missing profiling optional"
    # hypothesis 는 pbt 안에
    assert any("hypothesis" in d for d in optional["pbt"])
    # objgraph 는 profiling 안에
    assert any("objgraph" in d for d in optional["profiling"])


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
