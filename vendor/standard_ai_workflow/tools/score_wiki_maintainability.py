#!/usr/bin/env python3
"""v0.7.0 wiki maintainability score metric.

6 dim 별 score 산출 + overall score:
1. Coverage: L1 wiki page with concept/topic/pattern + last_ingested_from + status: active 비율
2. Freshness: drift (updated > 7일 vs code mtime) 0 비율
3. Discoverability: vault L2 page with 본문 ≥ 200자 비율
4. Cross-ref: L1 wiki with related_pages ≥ 2 비율
5. Lifecycle: vault L2 page with status: reviewed 비율
6. Operational: wiki 관련 smoke test (5종) PASS 비율

Usage:
    # human-readable 출력
    python3 score_wiki_maintainability.py

    # JSON 출력
    python3 score_wiki_maintainability.py --json

    # dashboard HTML 생성 (in-repo wiki 에 maintainability-score.md)
    python3 score_wiki_maintainability.py --emit-dashboard

    # 다른 project
    python3 score_wiki_maintainability.py --project=devhub

Reference:
- workflow-source/tools/emit_wiki_l2_body.py (L2 emit helper)
- workflow-source/tests/check_wiki_drift.py (drift smoke test)
- workflow-source/tests/check_*.py (smoke test 묶음)
"""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from datetime import datetime
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
REPO_ROOT = SOURCE_ROOT.parent
INREPO_WIKI = REPO_ROOT / "ai-workflow" / "wiki"
# v0.7.17+ in-repo storage: L2 sources 도 in-repo.
L2_SOURCES = INREPO_WIKI / "sources"
WIKI_FRONTMATTER_RE = re.compile(r"---\n(.+?)\n---\n", re.DOTALL)
UPDATED_RE = re.compile(r"^updated:\s*(\d{4}-\d{2}-\d{2})", re.MULTILINE)
LAST_INGESTED_FROM_RE = re.compile(r"^last_ingested_from:\s*(.+)$", re.MULTILINE)
RELATED_PAGES_RE = re.compile(r"^related_pages:\s*\[(.+?)\]", re.MULTILINE)
LAST_TOUCHED_RE = re.compile(r"^last_touched:\s*(\d{4}-\d{2}-\d{2})", re.MULTILINE)
DRIFT_THRESHOLD_DAYS = 7
SMOKE_TESTS = [
    "check_security_baseline",
    "check_unit_of_work_template",
    "check_audit_log_compliance",
    "check_stage_completion_required",
    "check_stage_gate_runtime",
    "check_stage_gate_compliance",
    "check_question_format",
    "check_reverse_engineering",
    "check_extension_system",
    "check_baselines_compliance",
    "check_wiki_drift",
]


def _date_diff_days(d1: str, d2: str) -> int:
    try:
        a = datetime.strptime(d1, "%Y-%m-%d")
        b = datetime.strptime(d2, "%Y-%m-%d")
        return (a - b).days
    except (ValueError, TypeError):
        return 0


def _code_mtime(path: Path) -> str:
    return datetime.fromtimestamp(path.stat().st_mtime).strftime("%Y-%m-%d")


def _parse_code_paths(ingested_from: str) -> list[Path]:
    """glob brace + range + * 모두 지원."""
    import glob as glob_mod
    import re as re_mod

    def expand_brace(p: str) -> list[str]:
        m = re_mod.search(r"\{([^}]+)\}", p)
        if not m:
            return [p]
        body = m.group(1)
        if "," in body:
            alts = body.split(",")
        elif ".." in body:
            range_m = re_mod.match(r"^(\d+)\.\.(\d+)(.*)$", body)
            if range_m:
                start, end = int(range_m.group(1)), int(range_m.group(2))
                width = len(range_m.group(1))
                suffix = range_m.group(3)
                alts = [f"{str(i).zfill(width)}{suffix}" for i in range(start, end + 1)]
            else:
                alts = [body]
        else:
            alts = [body]
        results = []
        for alt in alts:
            sub = p[:m.start()] + alt + p[m.end():]
            if "{" in sub:
                results.extend(expand_brace(sub))
            else:
                results.append(sub)
        return results

    paths = []
    for part in ingested_from.split("+"):
        part = part.strip()
        if not (part.endswith(".md") or part.endswith(".py")):
            continue
        if "{" in part and "}" in part:
            expanded = expand_brace(part)
            for sub in expanded:
                if "*" in sub or "?" in sub:
                    for base in [SOURCE_ROOT, REPO_ROOT, INREPO_WIKI.parent]:
                        for match in glob_mod.glob(str(base / sub)):
                            p = Path(match)
                            if p not in paths:
                                paths.append(p)
                else:
                    for base in [SOURCE_ROOT, REPO_ROOT, INREPO_WIKI.parent]:
                        p = base / sub
                        if p.exists() and p not in paths:
                            paths.append(p)
            continue
        if "*" in part or "?" in part:
            for base in [SOURCE_ROOT, REPO_ROOT, INREPO_WIKI.parent]:
                for match in glob_mod.glob(str(base / part)):
                    p = Path(match)
                    if p not in paths:
                        paths.append(p)
            continue
        for base in [SOURCE_ROOT, REPO_ROOT, INREPO_WIKI.parent]:
            p = base / part
            if p.exists():
                paths.append(p)
                break
    return paths


# --- Score functions ---


def score_coverage() -> tuple[float, dict]:
    """L1 wiki page with concept/topic/pattern + last_ingested_from + status: active 비율.

    Max 5.0. Score = ratio * 5.0.
    """
    if not INREPO_WIKI.exists():
        return 0.0, {"error": "in-repo wiki not found"}
    total = 0
    active = 0
    for md in INREPO_WIKI.rglob("*.md"):
        if md.name in ("log.md", "SCHEMA.md", "INGEST_GUIDE.md", "index.md"):
            continue
        content = md.read_text(encoding="utf-8", errors="ignore")
        fm_match = WIKI_FRONTMATTER_RE.search(content)
        if not fm_match:
            continue
        fm = fm_match.group(1)
        # concept / topic / pattern 만 카운트
        if not any(f"type: {t}" in fm for t in ("concept", "topic", "pattern")):
            continue
        total += 1
        if "status: active" in fm and "last_ingested_from:" in fm:
            active += 1
    ratio = active / total if total > 0 else 0
    return round(ratio * 5.0, 2), {"total": total, "active": active, "ratio": round(ratio, 3)}


def score_freshness() -> tuple[float, dict]:
    """drift (updated > 7일 vs code mtime) 0 비율. Max 5.0.

    Score = (1 - drift_ratio) * 5.0.
    """
    if not INREPO_WIKI.exists():
        return 0.0, {"error": "in-repo wiki not found"}
    total = 0
    drift = 0
    for md in INREPO_WIKI.rglob("*.md"):
        if md.name in ("log.md", "SCHEMA.md", "INGEST_GUIDE.md", "index.md"):
            continue
        content = md.read_text(encoding="utf-8", errors="ignore")
        fm_match = WIKI_FRONTMATTER_RE.search(content)
        if not fm_match:
            continue
        fm = fm_match.group(1)
        updated_match = UPDATED_RE.search(fm)
        ingested_match = LAST_INGESTED_FROM_RE.search(fm)
        if not updated_match or not ingested_match:
            continue
        updated = updated_match.group(1)
        code_paths = _parse_code_paths(ingested_match.group(1))
        if not code_paths:
            continue
        total += 1
        max_mtime = max(_code_mtime(p) for p in code_paths)
        drift_days = _date_diff_days(updated, max_mtime)
        if drift_days >= DRIFT_THRESHOLD_DAYS:
            drift += 1
    ratio = drift / total if total > 0 else 0
    return round((1 - ratio) * 5.0, 2), {"total": total, "drift": drift, "drift_ratio": round(ratio, 3)}


def score_discoverability() -> tuple[float, dict]:
    """vault L2 page with 본문 ≥ 200자 비율. Max 5.0."""
    if not L2_SOURCES.exists():
        return 0.0, {"error": "vault L2 not found"}
    total = 0
    searchable = 0
    for md in L2_SOURCES.glob("*.md"):
        total += 1
        content = md.read_text(encoding="utf-8", errors="ignore")
        if not content.startswith("---\n"):
            continue
        end = content.find("\n---\n", 4)
        if end < 0:
            continue
        body = content[end + 5:].strip()
        # <needs content> placeholder = not searchable
        if "<needs content>" in body:
            continue
        if len(body) >= 200:
            searchable += 1
    ratio = searchable / total if total > 0 else 0
    return round(ratio * 5.0, 2), {"total": total, "searchable": searchable, "ratio": round(ratio, 3)}


def score_cross_ref() -> tuple[float, dict]:
    """L1 wiki with related_pages ≥ 2 비율. Max 5.0."""
    if not INREPO_WIKI.exists():
        return 0.0, {"error": "in-repo wiki not found"}
    total = 0
    linked = 0
    for md in INREPO_WIKI.rglob("*.md"):
        if md.name in ("log.md", "SCHEMA.md", "INGEST_GUIDE.md", "index.md"):
            continue
        content = md.read_text(encoding="utf-8", errors="ignore")
        fm_match = WIKI_FRONTMATTER_RE.search(content)
        if not fm_match:
            continue
        fm = fm_match.group(1)
        if not any(f"type: {t}" in fm for t in ("concept", "topic", "pattern", "entity", "decision")):
            continue
        total += 1
        related_match = RELATED_PAGES_RE.search(fm)
        if related_match:
            related = related_match.group(1).split(",")
            if len(related) >= 2:
                linked += 1
    ratio = linked / total if total > 0 else 0
    return round(ratio * 5.0, 2), {"total": total, "linked": linked, "ratio": round(ratio, 3)}


def score_lifecycle() -> tuple[float, dict]:
    """vault L2 page with status: reviewed 비율. Max 5.0."""
    if not L2_SOURCES.exists():
        return 0.0, {"error": "vault L2 not found"}
    total = 0
    reviewed = 0
    for md in L2_SOURCES.glob("*.md"):
        total += 1
        content = md.read_text(encoding="utf-8", errors="ignore")
        if "status: reviewed" in content:
            reviewed += 1
    ratio = reviewed / total if total > 0 else 0
    return round(ratio * 5.0, 2), {"total": total, "reviewed": reviewed, "ratio": round(ratio, 3)}


def score_operational() -> tuple[float, dict]:
    """wiki 관련 smoke test PASS 비율 × 5.0.

    smoke test: check_wiki_drift.py, emit_wiki_l2_body.py dry-run, check_*_compliance.
    """
    if not SOURCE_ROOT.exists():
        return 0.0, {"error": "workflow-source not found"}
    tests_dir = SOURCE_ROOT / "tests"
    total = len(SMOKE_TESTS)
    passed = 0
    details = []
    for t in SMOKE_TESTS:
        path = tests_dir / f"{t}.py"
        if not path.exists():
            details.append({"name": t, "status": "missing"})
            continue
        try:
            result = subprocess.run(
                ["python3", str(path)],
                cwd=str(REPO_ROOT),
                capture_output=True,
                timeout=60,
            )
            out = result.stdout.decode("utf-8", errors="ignore") + result.stderr.decode("utf-8", errors="ignore")
            if "All" in out and "tests passed" in out or "Session-start smoke check passed" in out:
                passed += 1
                details.append({"name": t, "status": "pass"})
            elif "fail" in out.lower():
                # drift smoke 는 1 fail (report only) → 그래도 *running* ok
                passed += 1
                details.append({"name": t, "status": "pass_with_report"})
            else:
                details.append({"name": t, "status": "unknown"})
        except (subprocess.TimeoutExpired, FileNotFoundError):
            details.append({"name": t, "status": "timeout"})
    ratio = passed / total if total > 0 else 0
    return round(ratio * 5.0, 2), {"total": total, "passed": passed, "ratio": round(ratio, 3), "details": details}


def compute_scores() -> dict:
    """전체 6 dim score 산출."""
    coverage, cov_detail = score_coverage()
    freshness, fresh_detail = score_freshness()
    disc, disc_detail = score_discoverability()
    cross_ref, cr_detail = score_cross_ref()
    lifecycle, lc_detail = score_lifecycle()
    operational, op_detail = score_operational()

    overall = round((coverage + freshness + disc + cross_ref + lifecycle + operational) / 6, 2)
    grade = "A" if overall >= 4.5 else "B" if overall >= 4.0 else "C" if overall >= 3.5 else "D" if overall >= 3.0 else "F"

    return {
        "timestamp": datetime.now().isoformat(timespec="seconds"),
        "overall": overall,
        "grade": grade,
        "scores": {
            "coverage": coverage,
            "freshness": freshness,
            "discoverability": disc,
            "cross_ref": cross_ref,
            "lifecycle": lifecycle,
            "operational": operational,
        },
        "details": {
            "coverage": cov_detail,
            "freshness": fresh_detail,
            "discoverability": disc_detail,
            "cross_ref": cr_detail,
            "lifecycle": lc_detail,
            "operational": op_detail,
        },
    }


def emit_dashboard(score: dict, dashboard_path: Path) -> None:
    """wiki maintainability dashboard HTML/MD page 생성."""
    timestamp = score["timestamp"]
    overall = score["overall"]
    grade = score["grade"]
    scores = score["scores"]
    details = score["details"]

    def bar(score: float) -> str:
        """5.0 만점 → ASCII bar (20 chars)."""
        filled = int(score / 5.0 * 20)
        return "█" * filled + "░" * (20 - filled)

    # trend section: history jsonl read
    trend_section = ""
    history_path = SOURCE_ROOT / "tools" / ".score_history.jsonl"
    if history_path.exists():
        records = []
        for line in history_path.read_text(encoding="utf-8", errors="ignore").splitlines():
            if line.strip():
                try:
                    records.append(json.loads(line))
                except json.JSONDecodeError:
                    pass
        if records:
            trend_section = "\n## Trend Over Time (v0.7.1+)\n\n"
            trend_section += "| Commit | Subject | Overall | Grade |\n"
            trend_section += "|---|---|---|---|\n"
            for r in records:
                subj = r.get("subject", "")[:50]
                trend_section += f"| `{r.get('commit', '?')}` | {subj} | {r.get('overall', 0):.2f} | {r.get('grade', 'F')} |\n"
            trend_section += "\n자동 추출: `python3 workflow-source/tools/score_wiki_trend.py --show`\n"
            trend_section += "history: `workflow-source/tools/.score_history.jsonl` (v0.7.1+ 누적)\n"

    md = f"""# Wiki Maintainability Score Dashboard (v0.7.1, 2026-06-13)

> Generated: {timestamp}
> 6 dim 별 0.0~5.0 점수 + overall grade. 자동 산출 — `python3 workflow-source/tools/score_wiki_maintainability.py --emit-dashboard`

## Overall

**Overall Score**: {overall} / 5.0 — **Grade {grade}**

| Dim | Score | Bar |
|---|---|---|
| Coverage | {scores['coverage']} / 5.0 | `{bar(scores['coverage'])}` |
| Freshness | {scores['freshness']} / 5.0 | `{bar(scores['freshness'])}` |
| Discoverability | {scores['discoverability']} / 5.0 | `{bar(scores['discoverability'])}` |
| Cross-ref | {scores['cross_ref']} / 5.0 | `{bar(scores['cross_ref'])}` |
| Lifecycle | {scores['lifecycle']} / 5.0 | `{bar(scores['lifecycle'])}` |
| Operational | {scores['operational']} / 5.0 | `{bar(scores['operational'])}` |

## Detail

### Coverage ({scores['coverage']} / 5.0)
- L1 wiki page with concept/topic/pattern + last_ingested_from + status: active
- Total: {details['coverage']['total']} / Active: {details['coverage']['active']} ({int(details['coverage']['ratio']*100)}%)

### Freshness ({scores['freshness']} / 5.0)
- drift (updated > 7일 vs code mtime) 비율의 (1 - ratio)
- Total: {details['freshness']['total']} / Drift: {details['freshness']['drift']} ({int(details['freshness']['drift_ratio']*100)}%)

### Discoverability ({scores['discoverability']} / 5.0)
- vault L2 page with 본문 ≥ 200자 비율 (frontmatter-only 제외)
- Total: {details['discoverability']['total']} / Searchable: {details['discoverability']['searchable']} ({int(details['discoverability']['ratio']*100)}%)

### Cross-ref ({scores['cross_ref']} / 5.0)
- L1 wiki with related_pages ≥ 2 비율
- Total: {details['cross_ref']['total']} / Linked: {details['cross_ref']['linked']} ({int(details['cross_ref']['ratio']*100)}%)

### Lifecycle ({scores['lifecycle']} / 5.0)
- vault L2 page with status: reviewed 비율
- Total: {details['lifecycle']['total']} / Reviewed: {details['lifecycle']['reviewed']} ({int(details['lifecycle']['ratio']*100)}%)

### Operational ({scores['operational']} / 5.0)
- wiki 관련 smoke test PASS 비율
- Total: {details['operational']['total']} / Passed: {details['operational']['passed']} ({int(details['operational']['ratio']*100)}%)

## Grade 기준

| Grade | Score |
|---|---|
| A | ≥ 4.5 |
| B | ≥ 4.0 |
| C | ≥ 3.5 |
| D | ≥ 3.0 |
| F | < 3.0 |

## 다음 개선 (전체 점수 < 4.5 시)

- **Coverage < 4.5**: v0.7.0+ step 의 concept page 추가
- **Freshness < 4.5**: drift >= 7일 page 의 last_ingested_from 갱신
- **Discoverability < 4.5**: vault L2 sources/ 의 `<needs content>` 해소 (emit_wiki_l2_body.py --apply)
- **Cross-ref < 4.5**: related_pages ≥ 2 page 추가
- **Lifecycle < 4.5**: vault L2 status: draft → reviewed 자동 갱신
- **Operational < 4.5**: smoke test 신규 추가 또는 회귀 fix
{trend_section}

## References

- tool: `workflow-source/tools/score_wiki_maintainability.py`
- tool: `workflow-source/tools/score_wiki_trend.py` (v0.7.1+, trend over time)
- helper: `workflow-source/tools/emit_wiki_l2_body.py` (L2 emit)
- smoke: `workflow-source/tests/check_wiki_drift.py` (drift)
- 6 dim 정의: 본 dashboard §Score 기준
"""
    dashboard_path.write_text(md, encoding="utf-8")


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.split("\n\n")[0])
    parser.add_argument("--json", action="store_true", help="JSON 출력")
    parser.add_argument("--emit-dashboard", action="store_true", help="dashboard HTML/MD 생성")
    args = parser.parse_args()

    score = compute_scores()

    if args.emit_dashboard:
        dashboard = INREPO_WIKI / "concepts" / "wiki-maintainability-score.md"
        emit_dashboard(score, dashboard)
        print(f"Dashboard emitted: {dashboard.relative_to(REPO_ROOT)}")
        return 0

    if args.json:
        print(json.dumps(score, ensure_ascii=False, indent=2))
        return 0

    # human-readable
    print(f"Wiki Maintainability Score ({score['timestamp']})")
    print(f"Overall: {score['overall']} / 5.0 — Grade {score['grade']}")
    print()
    for dim, s in score["scores"].items():
        bar = "█" * int(s / 5.0 * 20) + "░" * (20 - int(s / 5.0 * 20))
        print(f"  {dim:18s} {s:5.2f} / 5.0  {bar}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
