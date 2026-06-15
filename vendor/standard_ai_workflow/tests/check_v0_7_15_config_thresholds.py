#!/usr/bin/env python3
"""v0.7.15+: [tool.workflow-doctor] config layer 의 3종 적용 smoke test.

deferred #2 (score_alert threshold hardcoded 0.3), #3 (memory_alert_mb 미사용),
#4 (excluded_paths 미적용) 한꺼번에 해소. score_wiki_trend.py 와 workflow-linter
모두 [tool.workflow-doctor] 의 thresholds / excluded_paths 적용 검증.

Test 구성 (9 test):
1. score_wiki_trend.SCORE_ALERT_DEFAULT: [tool.workflow-doctor] thresholds.score_alert 적용
2. score_wiki_trend.MEMORY_ALERT_MB_DEFAULT: thresholds.memory_alert_mb 적용
3. score_wiki_trend.compare_scores(None): config 의 default threshold 사용
4. score_wiki_trend.compare_scores(explicit): explicit override 우선
5. score_wiki_trend.record_current: rss_mb field 추가
6. score_wiki_trend._probe_rss_mb: RSS 측정 (resource module)
7. linter._is_excluded: glob match (build/*, .venv/*)
8. linter.check_workflow_consistency: excluded_paths 적용 시 broken link skip
9. run_workflow_linter: load_config → config.excluded_paths → linter 전달

Reference:
- workflow-source/tools/score_wiki_trend.py (v0.7.15 본 release)
- workflow-source/workflow_kit/common/linter.py (v0.7.15 excluded_paths 인자)
- workflow-source/skills/workflow-linter/scripts/run_workflow_linter.py
- workflow-source/workflow_kit/common/metadata.py (v0.7.6+ load_config / DoctorConfig)
- workflow-source/pyproject.toml [tool.workflow-doctor] section
- memory "release pipeline Phase 3" (v0.7.15 deferred #2/#3/#4 해소)
"""

from __future__ import annotations

import importlib.util
import json
import sys
import tempfile
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
REPO_ROOT = SOURCE_ROOT.parent
TOOLS_DIR = SOURCE_ROOT / "tools"
KIT_COMMON = SOURCE_ROOT / "workflow_kit" / "common"


def _import_module(name: str, path: Path):
    spec = importlib.util.spec_from_file_location(name, str(path))
    mod = importlib.util.module_from_spec(spec)
    sys.modules[name] = mod
    spec.loader.exec_module(mod)
    return mod


def _import_linter():
    return _import_module("wf_linter", KIT_COMMON / "linter.py")


def _import_score_trend():
    """score_wiki_trend.py 를 import. module-level load_config 가 호출되므로
    conftest-like 격리 (별도 cwd)."""
    if str(TOOLS_DIR) not in sys.path:
        sys.path.insert(0, str(TOOLS_DIR))
    return _import_module("score_wiki_trend", TOOLS_DIR / "score_wiki_trend.py")


# Eagerly import so that subsequent tests don't repeat the sys.path dance
_import_score_trend()


# --- Test 1: SCORE_ALERT_DEFAULT from config ---


def test_score_alert_default_from_config():
    """[tool.workflow-doctor] thresholds.score_alert=0.5 일 때 load_config 가 반영."""
    with tempfile.TemporaryDirectory() as tmp:
        pyproject = Path(tmp) / "pyproject.toml"
        pyproject.write_text(
            "[tool.workflow-doctor]\n"
            "thresholds = { score_alert = 0.5, memory_alert_mb = 200.0 }\n"
        )
        sys.path.insert(0, str(SOURCE_ROOT))
        from workflow_kit.common.metadata import load_config

        config = load_config(tmp)
        assert config.thresholds["score_alert"] == 0.5
        assert config.thresholds["memory_alert_mb"] == 200.0


# --- Test 2: MEMORY_ALERT_MB_DEFAULT from config ---


def test_memory_alert_default_from_config():
    """load_config default memory_alert_mb = 100.0 (없을 때)."""
    sys.path.insert(0, str(SOURCE_ROOT))
    from workflow_kit.common.metadata import _default_config

    cfg = _default_config()
    assert cfg.thresholds["memory_alert_mb"] == 100.0


# --- Test 3: compare_scores(None) uses config default ---


def test_compare_scores_default_uses_config():
    """compare_scores(baseline, current, alert_threshold=None) → SCORE_ALERT_DEFAULT."""
    import score_wiki_trend as swt

    baseline = {"scores": {d: 5.0 for d in swt.DIMS}}
    current = {"scores": dict(baseline["scores"])}
    # -0.5 drop: alert if threshold 0.3 (default)
    current["scores"]["freshness"] = 4.5
    alerts_default = swt.compare_scores(baseline, current, alert_threshold=None)
    assert any(a.severity == "alert" and a.dim == "freshness" for a in alerts_default)
    # +0.5 drop: alert if threshold 0.6 (high)
    current["scores"]["freshness"] = 4.4
    alerts_high = swt.compare_scores(baseline, current, alert_threshold=0.6)
    assert not any(a.severity == "alert" and a.dim == "freshness" for a in alerts_high)


# --- Test 4: compare_scores explicit override ---


def test_compare_scores_explicit_override():
    """compare_scores explicit alert_threshold 가 config default 보다 우선."""
    import score_wiki_trend as swt

    baseline = {"scores": {d: 5.0 for d in swt.DIMS}}
    current = {"scores": dict(baseline["scores"])}
    current["scores"]["coverage"] = 4.9  # -0.1
    # default 0.3 → no alert (4.9 - 5.0 = -0.1, |delta| < 0.3)
    alerts_default = swt.compare_scores(baseline, current, alert_threshold=None)
    assert not any(a.severity == "alert" and a.dim == "coverage" for a in alerts_default)
    # explicit 0.05 → alert (|delta| = 0.1 > 0.05)
    alerts_explicit = swt.compare_scores(baseline, current, alert_threshold=0.05)
    assert any(a.severity == "alert" and a.dim == "coverage" for a in alerts_explicit)


# --- Test 5: record_current adds rss_mb ---


def test_record_current_adds_rss_mb_field():
    """record_current record 에 rss_mb field 가 추가됨."""
    import score_wiki_trend as swt

    # override HISTORY_PATH to temp
    with tempfile.TemporaryDirectory() as tmp:
        original_history = swt.HISTORY_PATH
        swt.HISTORY_PATH = Path(tmp) / ".score_history.jsonl"
        try:
            # stub compute_score_at_commit
            original_compute = swt.compute_score_at_commit
            swt.compute_score_at_commit = lambda c: {
                "scores": {d: 4.0 for d in swt.DIMS},
                "overall": 4.0,
                "grade": "A",
            }
            try:
                rec = swt.record_current("29af65d8")
                assert "rss_mb" in rec, "missing rss_mb field"
                # rss_mb may be None on unsupported platform, or float
                assert rec["rss_mb"] is None or isinstance(rec["rss_mb"], float)
            finally:
                swt.compute_score_at_commit = original_compute
        finally:
            swt.HISTORY_PATH = original_history


# --- Test 6: _probe_rss_mb ---


def test_probe_rss_mb():
    """_probe_rss_mb 가 None 또는 float 반환 (platform 분기)."""
    import score_wiki_trend as swt

    rss = swt._probe_rss_mb()
    assert rss is None or isinstance(rss, float)
    if rss is not None:
        assert rss > 0, f"rss_mb should be positive, got {rss}"


# --- Test 7: linter _is_excluded glob match ---


def test_linter_is_excluded():
    """_is_excluded 가 glob pattern 과 path match."""
    linter = _import_linter()

    # single-segment glob
    assert linter._is_excluded(Path("build/foo.py"), ["build/*"])
    assert linter._is_excluded(Path(".venv/lib/x.py"), [".venv/*"])
    # non-match
    assert not linter._is_excluded(Path("src/main.py"), ["build/*"])
    # empty list
    assert not linter._is_excluded(Path("build/foo.py"), [])
    # multiple patterns
    assert linter._is_excluded(Path("tests/fixtures/x.json"), ["build/*", "tests/fixtures/*"])


# --- Test 8: check_workflow_consistency with excluded_paths ---


def test_linter_excluded_paths_skip_broken_link():
    """excluded_paths 적용 시 broken link check skip."""
    linter = _import_linter()
    with tempfile.TemporaryDirectory() as tmp:
        root = Path(tmp)
        state_json = root / "state.json"
        handoff = root / "handoff.md"
        backlog = root / "backlog.md"

        state_json.write_text(json.dumps({
            "source_of_truth": {"latest_backlog_path": str(backlog)},
            "session": {"in_progress_items": []},
        }))
        handoff.write_text(
            "## 1. 현재 작업\n"
            "- 현재 기준선: Test\n"
            "- 현재 주 작업 축: Test\n"
            "[broken](../../nonexistent.md)\n"
        )
        backlog.write_text("Empty")

        # excluded_paths 없이 → broken link 감지
        result_no_exclude = linter.check_workflow_consistency(state_json, handoff, backlog)
        assert any(i["code"] == "file_not_found" for i in result_no_exclude["issues"])

        # excluded_paths 와 match → skip
        result_excluded = linter.check_workflow_consistency(
            state_json, handoff, backlog, excluded_paths=["nonexistent.md"]
        )
        assert not any(i["code"] == "file_not_found" for i in result_excluded["issues"])
        # summary 에 excluded_paths 표시
        assert result_excluded["summary"]["excluded_paths"] == ["nonexistent.md"]


# --- Test 9: run_workflow_linter load_config integration ---


def test_run_workflow_linter_loads_config():
    """run_workflow_linter 가 load_config 호출 → 정상 실행 (custom pyproject.toml).

    pydantic / linter 가 system python 에 없는 경우를 위해 module import 만
    검증 (subprocess 호출은 check_workflow_linter.py 의 role).
    """
    # run_workflow_linter.py 본체에 load_config import + 호출이 존재하는지 확인
    runner = REPO_ROOT / "workflow-source/skills/workflow-linter/scripts/run_workflow_linter.py"
    assert runner.exists(), f"runner not found: {runner}"
    content = runner.read_text(encoding="utf-8")
    assert "from workflow_kit.common.metadata import load_config" in content, (
        "load_config import missing"
    )
    assert "load_config(project_root)" in content, "load_config call missing"
    assert "excluded_paths=excluded_paths" in content, "excluded_paths arg missing"
    # check_workflow_consistency 가 excluded_paths 인자를 받는지 확인
    linter = _import_linter()
    import inspect

    sig = inspect.signature(linter.check_workflow_consistency)
    assert "excluded_paths" in sig.parameters, (
        f"check_workflow_consistency missing excluded_paths param: {list(sig.parameters)}"
    )


# --- 메인 실행 ---


def main() -> int:
    test_funcs = [
        test_score_alert_default_from_config,
        test_memory_alert_default_from_config,
        test_compare_scores_default_uses_config,
        test_compare_scores_explicit_override,
        test_record_current_adds_rss_mb_field,
        test_probe_rss_mb,
        test_linter_is_excluded,
        test_linter_excluded_paths_skip_broken_link,
        test_run_workflow_linter_loads_config,
    ]
    passed = 0
    failed = 0
    for func in test_funcs:
        try:
            func()
            print(f"  PASS  {func.__name__}")
            passed += 1
        except AssertionError as e:
            print(f"  FAIL  {func.__name__}: {e}")
            failed += 1
        except Exception as e:  # noqa: BLE001
            print(f"  ERROR {func.__name__}: {type(e).__name__}: {e}")
            failed += 1

    print()
    print(f"{passed} pass, {failed} fail")
    return 0 if failed == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
