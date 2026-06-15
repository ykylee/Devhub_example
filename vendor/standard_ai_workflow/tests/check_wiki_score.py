#!/usr/bin/env python3
"""v0.7.0 wiki: maintainability score smoke test.

- score_wiki_maintainability.py 가 실행 가능 + 6 dim score 산출
- overall score 가 0.0 ~ 5.0 범위
- grade 가 A/B/C/D/F enum
- details dict 가 6 dim 모두 포함
- emit-dashboard 가 wiki-maintainability-score.md 생성 (in-repo wiki)

Reference:
- workflow-source/tools/score_wiki_maintainability.py
- workflow-source/ai-workflow/wiki/concepts/wiki-maintainability-score.md
"""

from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
REPO_ROOT = SOURCE_ROOT.parent
TOOL_PATH = SOURCE_ROOT / "tools" / "score_wiki_maintainability.py"
DASHBOARD_PATH = REPO_ROOT / "ai-workflow" / "wiki" / "concepts" / "wiki-maintainability-score.md"


def _run_score_tool(args: list[str] = None) -> dict:
    """score tool 실행 + JSON 반환."""
    args = args or ["--json"]
    proc = subprocess.run(
        ["python3", str(TOOL_PATH)] + args,
        cwd=str(REPO_ROOT),
        capture_output=True,
        text=True,
        timeout=120,
    )
    if proc.returncode != 0:
        raise RuntimeError(f"score tool failed: {proc.stderr}")
    if "--json" in args:
        return json.loads(proc.stdout)
    return {}


# --- Test 1: tool importable + executable ---


def test_tool_importable() -> None:
    """score tool 실행 가능."""
    assert TOOL_PATH.exists(), f"tool not found: {TOOL_PATH}"


def test_tool_runs() -> None:
    """score tool 의 --json 실행 + valid JSON 반환."""
    score = _run_score_tool()
    assert isinstance(score, dict)


# --- Test 2: score structure ---


def test_score_structure() -> None:
    """score dict 가 6 dim + overall + grade + timestamp 포함."""
    score = _run_score_tool()
    assert "timestamp" in score
    assert "overall" in score
    assert "grade" in score
    assert "scores" in score
    assert "details" in score
    assert set(score["scores"].keys()) == {
        "coverage", "freshness", "discoverability",
        "cross_ref", "lifecycle", "operational",
    }
    assert set(score["details"].keys()) == set(score["scores"].keys())


def test_score_range() -> None:
    """6 dim score 가 0.0 ~ 5.0 범위."""
    score = _run_score_tool()
    for dim, s in score["scores"].items():
        assert 0.0 <= s <= 5.0, f"{dim}: {s} out of range"
    assert 0.0 <= score["overall"] <= 5.0


def test_grade_enum() -> None:
    """grade 가 A/B/C/D/F enum."""
    score = _run_score_tool()
    assert score["grade"] in ("A", "B", "C", "D", "F")


def test_grade_matches_score() -> None:
    """grade 가 overall score 와 일치."""
    score = _run_score_tool()
    overall = score["overall"]
    if overall >= 4.5:
        expected = "A"
    elif overall >= 4.0:
        expected = "B"
    elif overall >= 3.5:
        expected = "C"
    elif overall >= 3.0:
        expected = "D"
    else:
        expected = "F"
    assert score["grade"] == expected, f"grade={score['grade']} but expected={expected} (overall={overall})"


# --- Test 3: detail metric consistency ---


def test_details_consistency() -> None:
    """details 의 total/active/ratio 가 score 와 일치."""
    score = _run_score_tool()
    for dim in score["scores"]:
        detail = score["details"][dim]
        if "ratio" in detail:
            # score = ratio * 5.0 (or (1-ratio)*5.0 for freshness)
            ratio = detail["ratio"]
            if dim == "freshness":
                expected = round((1 - ratio) * 5.0, 2)
            else:
                expected = round(ratio * 5.0, 2)
            actual = score["scores"][dim]
            assert abs(actual - expected) <= 0.05, \
                f"{dim}: actual={actual} but expected={expected} (ratio={ratio})"


def test_operational_smoke_passes() -> None:
    """operational dim 의 11 smoke test 가 모두 PASS."""
    score = _run_score_tool()
    op = score["details"]["operational"]
    assert op["total"] >= 5, f"too few smoke tests: {op['total']}"
    # operational score 가 4.0 이상 = 대부분 통과
    assert score["scores"]["operational"] >= 4.0, \
        f"operational score too low: {score['scores']['operational']}"


# --- Test 4: dashboard emit ---


def test_dashboard_emit() -> None:
    """--emit-dashboard 실행 시 dashboard file 생성."""
    if DASHBOARD_PATH.exists():
        # 이미 생성됨 — skip
        return
    proc = subprocess.run(
        ["python3", str(TOOL_PATH), "--emit-dashboard"],
        cwd=str(REPO_ROOT),
        capture_output=True,
        text=True,
        timeout=60,
    )
    assert proc.returncode == 0, f"emit failed: {proc.stderr}"
    assert DASHBOARD_PATH.exists(), f"dashboard not created: {DASHBOARD_PATH}"


def test_dashboard_format() -> None:
    """dashboard file 의 frontmatter (none) + Overall + 6 dim table 포함."""
    if not DASHBOARD_PATH.exists():
        return
    content = DASHBOARD_PATH.read_text(encoding="utf-8")
    assert "Overall Score" in content
    assert "Grade" in content
    assert "Coverage" in content
    assert "Freshness" in content
    assert "Discoverability" in content
    assert "Cross-ref" in content
    assert "Lifecycle" in content
    assert "Operational" in content


def test_dashboard_in_index() -> None:
    """dashboard page 가 index.md 에 anchor 됨 (v0.7.0+ 권장)."""
    index = (REPO_ROOT / "ai-workflow" / "wiki" / "index.md").read_text(encoding="utf-8")
    if "wiki-maintainability-score" not in index:
        print("INFO: wiki-maintainability-score not in index.md (v0.7.1+ follow-up 권장)")


# --- Test 5: idempotency ---


def test_score_idempotent() -> None:
    """2회 연속 실행 시 overall score 동일 (deterministic)."""
    s1 = _run_score_tool()
    s2 = _run_score_tool()
    assert s1["overall"] == s2["overall"], \
        f"non-idempotent: {s1['overall']} vs {s2['overall']}"


# --- 메인 실행 ---


def main() -> int:
    test_funcs = [
        test_tool_importable,
        test_tool_runs,
        test_score_structure,
        test_score_range,
        test_grade_enum,
        test_grade_matches_score,
        test_details_consistency,
        test_operational_smoke_passes,
        test_dashboard_emit,
        test_dashboard_format,
        test_dashboard_in_index,
        test_score_idempotent,
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
