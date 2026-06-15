#!/usr/bin/env python3
"""v0.7.1+: wiki maintainability score trend over time (commit 별 추적).

각 commit 의 score 를 jsonl history 에 append + ASCII chart 시각화.
회귀 감지: dim 별 점수 하락 ≥ 0.3 시 alert.

Usage:
    # 현재 HEAD 의 score 기록
    python3 score_wiki_trend.py --record-current

    # 최근 10 commit 의 score 재기록 (backfill)
    python3 score_wiki_trend.py --record-range 10

    # trend 시각화 (ASCII chart)
    python3 score_wiki_trend.py --show

    # JSON 출력
    python3 score_wiki_trend.py --json

Reference:
- tools/score_wiki_maintainability.py (단일 시점 score)
- tools/.score_history.jsonl (jsonl history)
- workflow-source/concepts/wiki-maintainability-score.md (dashboard)
"""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
REPO_ROOT = SOURCE_ROOT.parent
HISTORY_PATH = SOURCE_ROOT / "tools" / ".score_history.jsonl"

# trend 시각화용 dim
DIMS = ["coverage", "freshness", "discoverability", "cross_ref", "lifecycle", "operational"]

# v0.7.15: alert threshold 의 *config layer* 적용. v0.7.7 deferred #2 + #3 해소.
# hardcoded 0.3 / 100.0 → [tool.workflow-doctor] 의 thresholds dict 에서 load.
# - score_alert: dim 별 하락 alert 임계값 (default 0.3, v0.7.7 default)
# - memory_alert_mb: runtime RSS alert 임계값 (default 100.0, profiling 용)
def _load_config_thresholds() -> dict:
    """pyproject.toml [tool.workflow-doctor] thresholds load. 실패 시 default."""
    try:
        sys.path.insert(0, str(SOURCE_ROOT))
        from workflow_kit.common.metadata import load_config

        config = load_config(REPO_ROOT / "workflow-source")
        return dict(config.thresholds) if config.thresholds else {}
    except Exception:
        return {"score_alert": 0.3, "memory_alert_mb": 100.0}


THRESHOLDS = _load_config_thresholds()
SCORE_ALERT_DEFAULT = THRESHOLDS.get("score_alert", 0.3)
MEMORY_ALERT_MB_DEFAULT = THRESHOLDS.get("memory_alert_mb", 100.0)


def get_git_commits(limit: int = 10) -> list[tuple[str, str]]:
    """git log 의 최근 N commit (hash, subject) 반환."""
    proc = subprocess.run(
        ["git", "log", f"-n{limit}", "--pretty=format:%H %s"],
        cwd=str(REPO_ROOT),
        capture_output=True,
        text=True,
        timeout=30,
    )
    if proc.returncode != 0:
        return []
    out = []
    for line in proc.stdout.strip().split("\n"):
        if not line:
            continue
        parts = line.split(" ", 1)
        if len(parts) == 2:
            out.append((parts[0], parts[1]))
        else:
            out.append((parts[0], ""))
    return out


def compute_score_at_commit(commit: str) -> dict:
    """특정 commit 의 score 산출 (git checkout 없이 working tree 기준)."""
    # score tool 실행 (현재 working tree 기준)
    proc = subprocess.run(
        ["python3", str(SOURCE_ROOT / "tools" / "score_wiki_maintainability.py"), "--json"],
        cwd=str(REPO_ROOT),
        capture_output=True,
        text=True,
        timeout=120,
    )
    if proc.returncode != 0:
        return {"error": proc.stderr}
    try:
        return json.loads(proc.stdout)
    except json.JSONDecodeError:
        return {"error": "invalid JSON"}


def record_current(commit: str) -> dict:
    """현재 commit 의 score 기록 (append to jsonl)."""
    score = compute_score_at_commit(commit)
    # v0.7.15+: runtime RSS probe (profiling memory threshold, deferred #3 해소)
    rss_mb = _probe_rss_mb()
    record = {
        "timestamp": datetime.now().isoformat(timespec="seconds"),
        "commit": commit[:7],
        "scores": score.get("scores", {}),
        "overall": score.get("overall", 0),
        "grade": score.get("grade", "F"),
        "rss_mb": rss_mb,
    }
    HISTORY_PATH.parent.mkdir(parents=True, exist_ok=True)
    with HISTORY_PATH.open("a", encoding="utf-8") as f:
        f.write(json.dumps(record, ensure_ascii=False) + "\n")
    if rss_mb is not None and rss_mb > MEMORY_ALERT_MB_DEFAULT:
        print(
            f"⚠️  RSS {rss_mb:.1f}MB > memory_alert_mb {MEMORY_ALERT_MB_DEFAULT:.1f}MB "
            f"(config.thresholds['memory_alert_mb'])",
            file=sys.stderr,
        )
    return record


def _probe_rss_mb() -> float | None:
    """현재 process 의 RSS (resident set size) MB 측정. 실패 시 None.

    macOS / Linux 에서 동작 (resource module). Windows 는 psutil fallback 가능하나
    본 tool 의 target 은 macOS / Linux.
    """
    try:
        import resource  # type: ignore[import-not-found]

        usage = resource.getrusage(resource.RUSAGE_SELF)
        # ru_maxrss: macOS = bytes, Linux = KB. platform 분기.
        rss = usage.ru_maxrss
        if sys.platform == "darwin":
            return rss / (1024 * 1024)
        return rss / 1024  # Linux KB
    except Exception:
        return None


def record_range(limit: int) -> list[dict]:
    """최근 N commit 의 score 재기록 (backfill). 현재 working tree 기준.

    주의: git checkout 없이 현재 state 의 score 를 N 회 기록. commit 별
    *정확한* score 는 git checkout 필요 (시간 소요). 본 도구는 *추세* 가
    목적이며, *정확도* 는 smoke 의 reference 만.
    """
    commits = get_git_commits(limit)
    if not commits:
        print("No commits found")
        return []

    # 기존 history 의 해당 commit entry 제거 (overwrite)
    existing = load_history()
    existing_keys = {r["commit"] for r in existing}
    records = []
    for commit_hash, subject in commits:
        # score tool 실행
        score = compute_score_at_commit(commit_hash)
        record = {
            "timestamp": datetime.now().isoformat(timespec="seconds"),
            "commit": commit_hash[:7],
            "subject": subject,
            "scores": score.get("scores", {}),
            "overall": score.get("overall", 0),
            "grade": score.get("grade", "F"),
        }
        records.append(record)

    # 기존 + 신규 (중복 제거)
    new_commits = {r["commit"] for r in records}
    filtered = [r for r in existing if r["commit"] not in new_commits]
    combined = filtered + records

    # 다시 쓰기
    HISTORY_PATH.parent.mkdir(parents=True, exist_ok=True)
    with HISTORY_PATH.open("w", encoding="utf-8") as f:
        for r in combined:
            f.write(json.dumps(r, ensure_ascii=False) + "\n")
    return records


def load_history() -> list[dict]:
    """jsonl history 로드."""
    if not HISTORY_PATH.exists():
        return []
    records = []
    for line in HISTORY_PATH.read_text(encoding="utf-8", errors="ignore").splitlines():
        if not line.strip():
            continue
        try:
            records.append(json.loads(line))
        except json.JSONDecodeError:
            continue
    return records


def ascii_chart(records: list[dict], dim: str = "overall") -> str:
    """ASCII chart: x=commit, y=score."""
    if not records:
        return "(no records)"
    if dim not in DIMS and dim != "overall":
        return f"(unknown dim: {dim})"

    # commit prefix → score
    series = []
    for r in records:
        if dim == "overall":
            v = r.get("overall", 0)
        else:
            v = r.get("scores", {}).get(dim, 0)
        series.append((r.get("commit", "?"), v))

    # 0~5 scale
    rows = []
    for commit, v in series:
        bar_len = int(v / 5.0 * 30)
        bar = "█" * bar_len + "░" * (30 - bar_len)
        rows.append(f"  {commit}  {v:5.2f}  {bar}")

    return "\n".join(rows)


@dataclass
class DimAlert:
    """dim 별 alert 결과."""

    dim: str
    baseline: float
    current: float
    delta: float  # current - baseline
    severity: str  # "alert" | "info" | "ok"


def compare_scores(
    baseline: dict,
    current: dict,
    alert_threshold: float | None = None,
) -> list[DimAlert]:
    """baseline vs current 의 dim 별 비교. 하락 ≥ alert_threshold → alert.

    Args:
        baseline: 이전 commit 의 score record
        current: 현재 commit 의 score record
        alert_threshold: dim 별 하락 alert 임계값. None 이면 ``config.thresholds['score_alert']``
            (default 0.3). explicit override 가능.

    Returns:
        DimAlert list (dim 별 1개)
    """
    if alert_threshold is None:
        alert_threshold = SCORE_ALERT_DEFAULT
    alerts = []
    for dim in DIMS:
        b = baseline.get("scores", {}).get(dim, 0.0)
        c = current.get("scores", {}).get(dim, 0.0)
        delta = c - b
        if delta <= -alert_threshold:
            severity = "alert"
        elif delta >= alert_threshold:
            severity = "info"
        else:
            severity = "ok"
        alerts.append(DimAlert(dim=dim, baseline=b, current=c, delta=delta, severity=severity))
    return alerts


def print_alerts(alerts: list[DimAlert], baseline_label: str, current_label: str, alert_threshold: float) -> tuple[int, int]:
    """alert 출력 + (alert_count, info_count) 반환.

    CI 통합용: exit code 는 별도 caller 가 결정.
    """
    alert_count = 0
    info_count = 0
    print(f"Dim alerts ({baseline_label} → {current_label}, threshold={alert_threshold}):")
    print()
    for a in alerts:
        sign = "+" if a.delta >= 0 else ""
        bar = ""
        if a.severity == "alert":
            color = "🔴"
            alert_count += 1
        elif a.severity == "info":
            color = "🟢"
            info_count += 1
        else:
            color = "⚪"
        print(f"  {color} {a.dim:18s} {a.baseline:5.2f} → {a.current:5.2f} ({sign}{a.delta:+.2f})")
    print()
    print(f"Total: {alert_count} alert(s), {info_count} improvement(s)")
    return alert_count, info_count


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.split("\n\n")[0])
    parser.add_argument("--record-current", action="store_true", help="현재 HEAD 의 score 기록")
    parser.add_argument("--record-range", type=int, default=0, help="최근 N commit 의 score 재기록")
    parser.add_argument("--show", action="store_true", help="trend 시각화 (ASCII chart)")
    parser.add_argument("--json", action="store_true", help="JSON 출력")
    parser.add_argument("--alert", action="store_true", help="dim 별 하락 alert (≥ threshold)")
    parser.add_argument("--baseline", help="baseline commit (alert 시 비교 대상)")
    parser.add_argument(
        "--threshold",
        type=float,
        default=SCORE_ALERT_DEFAULT,
        help=f"alert 임계값 (default: {SCORE_ALERT_DEFAULT} from [tool.workflow-doctor] config)",
    )
    args = parser.parse_args()

    # 현재 HEAD
    head_proc = subprocess.run(
        ["git", "rev-parse", "HEAD"],
        cwd=str(REPO_ROOT),
        capture_output=True,
        text=True,
    )
    head = head_proc.stdout.strip()

    if args.record_current:
        record = record_current(head)
        print(f"Recorded: {record['commit']} overall={record['overall']} grade={record['grade']}")
        return 0

    if args.record_range > 0:
        records = record_range(args.record_range)
        print(f"Recorded {len(records)} commits (range={args.record_range})")
        for r in records:
            print(f"  {r['commit']}: overall={r['overall']} grade={r['grade']}")
        return 0

    if args.alert:
        # baseline commit 의 score record 찾기
        if not args.baseline:
            print("ERROR: --alert requires --baseline=<commit>")
            return 2

        records = load_history()
        baseline_rec = next((r for r in records if r.get("commit") == args.baseline[:7]), None)
        if not baseline_rec:
            print(f"ERROR: baseline commit {args.baseline[:7]} not found in history")
            return 2

        # 현재 score 측정
        current_score = compute_score_at_commit(head)
        if "error" in current_score:
            print(f"ERROR: cannot compute current score: {current_score['error']}")
            return 2

        # baseline + current 를 동일한 형식으로 normalize
        baseline = {
            "commit": baseline_rec["commit"],
            "scores": baseline_rec.get("scores", {}),
            "overall": baseline_rec.get("overall", 0),
        }
        current = {
            "commit": head[:7],
            "scores": current_score.get("scores", {}),
            "overall": current_score.get("overall", 0),
        }

        alerts = compare_scores(baseline, current, alert_threshold=args.threshold)
        alert_count, info_count = print_alerts(
            alerts,
            f"baseline={baseline_rec['commit']}",
            f"current={head[:7]}",
            args.threshold,
        )

        if alert_count > 0:
            print()
            print(f"❌ ALERT: {alert_count} dim(s) dropped ≥ {args.threshold} since baseline")
            return 1
        else:
            print()
            print(f"✅ OK: no dim dropped ≥ {args.threshold} since baseline")
            return 0

    records = load_history()

    if args.json:
        print(json.dumps(records, ensure_ascii=False, indent=2))
        return 0

    # human-readable
    if not records:
        print("(empty history)")
        print()
        print("Usage:")
        print("  --record-current: 현재 HEAD 의 score 기록")
        print("  --record-range N: 최근 N commit 의 score 재기록")
        print("  --show: trend 시각화")
        return 0

    print(f"Wiki Maintainability Score Trend (n={len(records)})")
    print()

    # overall trend
    print("## Overall (5.0 max)")
    print(ascii_chart(records, "overall"))
    print()

    # dim 별 trend
    for dim in DIMS:
        print(f"## {dim} (5.0 max)")
        print(ascii_chart(records, dim))
        print()

    return 0


if __name__ == "__main__":
    sys.exit(main())
