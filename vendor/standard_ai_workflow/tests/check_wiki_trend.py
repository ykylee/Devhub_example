#!/usr/bin/env python3
"""v0.7.1+: wiki maintainability score trend smoke test.

- tools/score_wiki_trend.py 실행 가능
- .score_history.jsonl load + parse
- ASCII chart format 검증
- dashboard 가 trend section 포함 (emit-dashboard 후)

Reference:
- workflow-source/tools/score_wiki_trend.py
- workflow-source/tools/.score_history.jsonl
- workflow-source/concepts/wiki-maintainability-score.md
"""

from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
REPO_ROOT = SOURCE_ROOT.parent
TOOL_PATH = SOURCE_ROOT / "tools" / "score_wiki_trend.py"
HISTORY_PATH = SOURCE_ROOT / "tools" / ".score_history.jsonl"
DASHBOARD_PATH = REPO_ROOT / "ai-workflow" / "wiki" / "concepts" / "wiki-maintainability-score.md"


def _run(args: list[str], timeout: int = 60) -> tuple[int, str, str]:
    proc = subprocess.run(
        ["python3", str(TOOL_PATH)] + args,
        cwd=str(REPO_ROOT),
        capture_output=True,
        text=True,
        timeout=timeout,
    )
    return proc.returncode, proc.stdout, proc.stderr


# --- Test 1: tool importable + executable ---


def test_tool_importable() -> None:
    """trend tool 실행 가능."""
    assert TOOL_PATH.exists(), f"tool not found: {TOOL_PATH}"


def test_tool_show_runs() -> None:
    """--show 실행 + 출력에 trend 시각화 포함."""
    rc, out, err = _run(["--show"])
    assert rc == 0, f"tool failed: {err}"
    assert "Wiki Maintainability Score Trend" in out
    assert "Overall (5.0 max)" in out
    assert "coverage" in out
    assert "freshness" in out
    assert "discoverability" in out


# --- Test 2: history jsonl ---


def test_history_file_exists() -> None:
    """.score_history.jsonl 존재 (v0.7.1+ 의 trend 누적)."""
    assert HISTORY_PATH.exists(), f"history not found: {HISTORY_PATH}"


def test_history_valid_jsonl() -> None:
    """history 의 각 줄이 valid JSON."""
    if not HISTORY_PATH.exists():
        return
    for i, line in enumerate(HISTORY_PATH.read_text(encoding="utf-8", errors="ignore").splitlines()):
        if not line.strip():
            continue
        try:
            rec = json.loads(line)
        except json.JSONDecodeError as e:
            raise AssertionError(f"line {i+1} invalid JSON: {e}\n  {line[:100]}")
        # required fields
        for k in ("timestamp", "commit", "scores", "overall", "grade"):
            assert k in rec, f"line {i+1} missing field: {k}"


def test_history_score_in_range() -> None:
    """history 의 overall / scores 가 0.0~5.0 범위."""
    if not HISTORY_PATH.exists():
        return
    for i, line in enumerate(HISTORY_PATH.read_text(encoding="utf-8", errors="ignore").splitlines()):
        if not line.strip():
            continue
        rec = json.loads(line)
        assert 0.0 <= rec["overall"] <= 5.0, f"line {i+1}: overall {rec['overall']} out of range"
        for dim, s in rec.get("scores", {}).items():
            assert 0.0 <= s <= 5.0, f"line {i+1}: {dim} {s} out of range"


# --- Test 3: chart format ---


def test_chart_bar_chars() -> None:
    """ASCII bar chart 가 █ / ░ chars 만 사용."""
    rc, out, _ = _run(["--show"])
    for line in out.splitlines():
        if "█" in line or "░" in line:
            # bar line
            for ch in line:
                if ch not in ("█", "░", " ", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9", ".", "-", "│", "└", "┌"):
                    pass  # ignore others
            # bar line 은 █ / ░ / space / digits 만
            non_bar = [c for c in line if c not in ("█", "░", " ", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9", ".")]
            # non-bar chars 는 prefix (commit, score) 만
            assert all(c.isalnum() or c in (".", " ", "│") for c in non_bar), \
                f"unexpected char in bar line: {line!r}"


def test_json_output() -> None:
    """--json 실행 + valid JSON."""
    rc, out, err = _run(["--json"])
    assert rc == 0, f"json failed: {err}"
    data = json.loads(out)
    assert isinstance(data, list)
    assert len(data) >= 1


# --- Test 4: dashboard integration ---


def test_dashboard_trend_section() -> None:
    """dashboard page 가 trend section 포함 (Trend Over Time heading)."""
    if not DASHBOARD_PATH.exists():
        return
    content = DASHBOARD_PATH.read_text(encoding="utf-8")
    assert "## Trend Over Time" in content, "dashboard missing Trend Over Time section"
    # 1 row 이상 (commit hash + overall)
    assert "0052da1" in content or "bad14d8" in content, \
        "dashboard trend table missing sample commit"


def test_dashboard_includes_trend_tool_reference() -> None:
    """dashboard 가 trend tool 의 path 명시."""
    if not DASHBOARD_PATH.exists():
        return
    content = DASHBOARD_PATH.read_text(encoding="utf-8")
    assert "score_wiki_trend" in content, "dashboard missing trend tool reference"


# --- Test 5: idempotency ---


def test_show_idempotent() -> None:
    """--show 2회 연속 실행 시 결과 동일."""
    _, out1, _ = _run(["--show"])
    _, out2, _ = _run(["--show"])
    assert out1 == out2, "show not idempotent"


# --- Test 6: alert 분기 ---


def test_compare_scores_no_alert() -> None:
    """baseline < current → alert 0, exit 0."""
    import sys
    sys.path.insert(0, str(SOURCE_ROOT / "tools"))
    import score_wiki_trend as swt

    baseline = {"scores": {d: 4.0 for d in swt.DIMS}}
    current = {"scores": {d: 4.5 for d in swt.DIMS}}  # +0.5 (info)
    alerts = swt.compare_scores(baseline, current, alert_threshold=0.3)
    assert all(a.severity in ("info", "ok") for a in alerts), "expected no alert"
    assert any(a.severity == "info" for a in alerts), "expected info"


def test_compare_scores_alert() -> None:
    """dim 별 -0.5 → alert 1, exit 1."""
    import sys
    sys.path.insert(0, str(SOURCE_ROOT / "tools"))
    import score_wiki_trend as swt

    baseline = {"scores": {d: 5.0 for d in swt.DIMS}}
    current = {"scores": dict(baseline["scores"])}
    current["scores"]["freshness"] = 4.5  # -0.5

    alerts = swt.compare_scores(baseline, current, alert_threshold=0.3)
    alert_dims = [a.dim for a in alerts if a.severity == "alert"]
    assert "freshness" in alert_dims
    assert len(alert_dims) == 1


def test_alert_cli_no_alert() -> None:
    """CLI --alert 실행 + exit 0 (real baseline 7a4dbae vs current)."""
    rc, out, err = _run(["--alert", "--baseline=7a4dbae"], timeout=120)
    assert rc in (0, 1), f"unexpected exit code: {rc}"
    assert "Dim alerts" in out
    assert "alert(s)" in out
    # baseline 7a4dbae (D 3.70) vs current (A 4.67) → 0 alert
    assert rc == 0, f"expected 0 alert, got exit {rc}: {out[-200:]}"


def test_alert_cli_missing_baseline() -> None:
    """--alert without --baseline → exit 2."""
    rc, out, _ = _run(["--alert"], timeout=30)
    assert rc == 2
    assert "ERROR" in out
    assert "--baseline" in out


# --- 메인 실행 ---


def main() -> int:
    test_funcs = [
        test_tool_importable,
        test_tool_show_runs,
        test_history_file_exists,
        test_history_valid_jsonl,
        test_history_score_in_range,
        test_chart_bar_chars,
        test_json_output,
        test_dashboard_trend_section,
        test_dashboard_includes_trend_tool_reference,
        test_show_idempotent,
        test_compare_scores_no_alert,
        test_compare_scores_alert,
        test_alert_cli_no_alert,
        test_alert_cli_missing_baseline,
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
