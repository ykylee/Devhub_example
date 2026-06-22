#!/usr/bin/env python3
"""Incident runbook tuning analyzer (umbrella doc §13.2 row 3 + §11.1 + §11.3).

M-v0.2.0 PoC 운영 +1주 manual SOP (ETA 2026-06-26~27):
- var/audit/audit-YYYY-MM-DD.jsonl 분석
- 7 incident type (§11.1) 별 frequency + false positive/negative 통계
- §11.3 monitoring 5 지표 threshold 자동 추천

Usage:
    python scripts/analyze_incidents.py                    # last 7 days
    python scripts/analyze_incidents.py --days 14          # last 14 days
    python scripts/analyze_incidents.py --audit-dir custom  # custom audit dir

Output: structured JSON report to stdout + human-readable summary.
"""
from __future__ import annotations

import argparse
import json
import sys
from collections import Counter, defaultdict
from datetime import datetime, timedelta, timezone
from pathlib import Path

# --- §11.1 incident type mapping (§11.1.1~§11.1.7) ---

INCIDENT_TYPES = {
    "sync_failure": {
        "runbook": "§11.1.1",
        "audit_events": ["audit.raw.received"],
        "success_field": "success",
        "current_threshold_warning": 0.99,
        "current_threshold_critical": 0.95,
    },
    "credential_expired": {
        "runbook": "§11.1.2",
        "audit_events": ["audit.raw.received"],
        "match_condition": lambda e: e.get("success") is False
        and e.get("last_error", "").startswith(("401", "403")),
        "current_threshold_warning": 0,
        "current_threshold_critical": 1,
    },
    "pi_ingest_timeout": {
        "runbook": "§11.1.3",
        "audit_events": ["audit.pi_ingest"],
        "match_condition": lambda e: e.get("degraded") is True
        or e.get("timeout_seconds", 0) > 30,
        "current_threshold_warning": 0.95,
        "current_threshold_critical": 0.80,
    },
    "retention_cron_failure": {
        "runbook": "§11.1.4",
        "audit_events": ["audit.retention.deleted", "audit.retention.failed"],
        "match_condition": lambda e: e.get("event") == "audit.retention.failed",
        "current_threshold_warning": 1,
        "current_threshold_critical": 5,
    },
    "integrity_violation": {
        "runbook": "§11.1.5",
        "audit_events": ["audit.raw.integrity_violation"],
        "current_threshold_warning": 0.0001,
        "current_threshold_critical": 0.001,
    },
    "archive_trigger_failure": {
        "runbook": "§11.1.6",
        "audit_events": ["audit.concept.archive"],
        "match_condition": lambda e: e.get("success") is False,
        "current_threshold_warning": 1,
        "current_threshold_critical": 5,
    },
    "stale_link_detected": {
        "runbook": "§11.1.7",
        "audit_events": ["audit.pi_link_resolve"],
        "match_condition": lambda e: e.get("stale") is True,
        "current_threshold_warning": 5,
        "current_threshold_critical": 20,
    },
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        prog="analyze_incidents.py",
        description="Incident runbook tuning analyzer (§13.2 row 3)",
    )
    parser.add_argument(
        "--days",
        type=int,
        default=7,
        help="분석 기간 (default: 7일, M-v0.2.0 PoC 운영 +1주)",
    )
    parser.add_argument(
        "--audit-dir",
        type=Path,
        default=None,
        help="var/audit/ 디렉터리 경로 (default: settings.var_dir / audit)",
    )
    return parser.parse_args()


def load_audit_events(audit_dir: Path, from_dt: datetime, to_dt: datetime) -> list[dict]:
    """Load all audit events from JSONL files in [from_dt, to_dt] range."""
    if not audit_dir.exists():
        return []
    events: list[dict] = []
    current = from_dt.date()
    end = to_dt.date()
    while current <= end:
        path = audit_dir / f"audit-{current.strftime('%Y-%m-%d')}.jsonl"
        if path.exists():
            try:
                with path.open(encoding="utf-8") as f:
                    for line in f:
                        line = line.strip()
                        if not line:
                            continue
                        try:
                            events.append(json.loads(line))
                        except json.JSONDecodeError:
                            pass
            except OSError:
                pass
        current += timedelta(days=1)
    return events


def count_by_type(events: list[dict]) -> dict[str, dict[str, int]]:
    """Count events per incident type."""
    result: dict[str, dict[str, int]] = {}
    for type_name, config in INCIDENT_TYPES.items():
        match = config.get("match_condition")
        count = 0
        fail_count = 0
        for e in events:
            if e.get("event") not in config["audit_events"]:
                continue
            if match is not None:
                if match(e):
                    count += 1
                continue
            if e.get("success") is False:
                fail_count += 1
            count += 1
        result[type_name] = {
            "total_events": count,
            "failures": fail_count,
            "incidents": count if match is not None else fail_count,
        }
    return result


def recommend_thresholds(
    type_name: str,
    stats: dict[str, int],
    window_days: int,
) -> dict[str, float]:
    """Recommend adjusted thresholds based on observed incident frequency."""
    config = INCIDENT_TYPES[type_name]
    incidents = stats["incidents"]
    incidents_per_day = incidents / max(window_days, 1)
    warn = config["current_threshold_warning"]
    crit = config["current_threshold_critical"]
    if type_name in ("sync_failure", "pi_ingest_timeout"):
        # success rate metrics: lower threshold = more sensitive
        if incidents_per_day > 3:
            warn_new = round(warn - 0.02, 4)
            crit_new = round(crit - 0.02, 4)
            recommended = "tighten"
        elif incidents_per_day < 0.5:
            warn_new = round(warn + 0.01, 4)
            crit_new = round(crit + 0.01, 4)
            recommended = "loosen"
        else:
            warn_new = warn
            crit_new = crit
            recommended = "maintain"
    elif type_name in ("credential_expired", "archive_trigger_failure", "retention_cron_failure", "stale_link_detected"):
        # count metrics: higher threshold = more sensitive
        if incidents_per_day > 2:
            warn_new = max(warn - 1, 0)
            crit_new = max(crit - 1, 1)
            recommended = "tighten"
        elif incidents_per_day < 0.2:
            warn_new = warn + 1
            crit_new = crit + 2
            recommended = "loosen"
        else:
            warn_new = warn
            crit_new = crit
            recommended = "maintain"
    elif type_name == "integrity_violation":
        if incidents_per_day > 0.5:
            warn_new = max(warn / 2, 0.0001)
            crit_new = max(crit / 2, 0.001)
            recommended = "tighten"
        else:
            warn_new = warn
            crit_new = crit
            recommended = "maintain"
    else:
        warn_new = warn
        crit_new = crit
        recommended = "maintain"
    return {
        "current_warning": warn,
        "current_critical": crit,
        "recommended_warning": warn_new,
        "recommended_critical": crit_new,
        "recommendation": recommended,
        "incidents_per_day": round(incidents_per_day, 3),
    }


def generate_report(
    events: list[dict],
    from_dt: datetime,
    to_dt: datetime,
    window_days: int,
) -> dict:
    """Generate structured tuning report."""
    stats = count_by_type(events)
    recommendations = {}
    for type_name, type_stats in stats.items():
        recommendations[type_name] = recommend_thresholds(type_name, type_stats, window_days)
    return {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "window": {
            "from": from_dt.isoformat(),
            "to": to_dt.isoformat(),
            "days": window_days,
        },
        "total_events": len(events),
        "by_type": stats,
        "recommendations": recommendations,
        "summary": {
            "total_incidents": sum(s["incidents"] for s in stats.values()),
            "types_with_incidents": sum(1 for s in stats.values() if s["incidents"] > 0),
            "types_total": len(stats),
        },
    }


def print_human_report(report: dict) -> None:
    """Print human-readable summary to stderr."""
    print(f"\n=== Incident Runbook Tuning Report ===", file=sys.stderr)
    print(f"Window: {report['window']['from']} ~ {report['window']['to']} ({report['window']['days']}d)", file=sys.stderr)
    print(f"Total events: {report['total_events']}", file=sys.stderr)
    print(f"Total incidents: {report['summary']['total_incidents']}", file=sys.stderr)
    print(f"Types with incidents: {report['summary']['types_with_incidents']}/{report['summary']['types_total']}", file=sys.stderr)
    if report["summary"]["total_incidents"] == 0:
        print("\nNo incidents observed in window. Thresholds maintained.", file=sys.stderr)
        print("(PoC 운영 +1주 manual SOP, §13.2 row 3 ETA 2026-06-26~27)", file=sys.stderr)
        return
    print("\nRecommendations:", file=sys.stderr)
    for type_name, rec in report["recommendations"].items():
        if rec["incidents_per_day"] == 0:
            continue
        runbook = INCIDENT_TYPES[type_name]["runbook"]
        print(f"  {type_name} ({runbook}):", file=sys.stderr)
        print(f"    incidents/day: {rec['incidents_per_day']}", file=sys.stderr)
        print(f"    current: warning={rec['current_warning']} critical={rec['current_critical']}", file=sys.stderr)
        print(f"    recommended: warning={rec['recommended_warning']} critical={rec['recommended_critical']} ({rec['recommendation']})", file=sys.stderr)


def main() -> int:
    args = parse_args()
    to_dt = datetime.now(timezone.utc)
    from_dt = to_dt - timedelta(days=args.days)
    audit_dir = args.audit_dir
    if audit_dir is None:
        from backend_knowledge.config import get_settings

        audit_dir = get_settings().var_dir / "audit"
    events = load_audit_events(audit_dir, from_dt, to_dt)
    report = generate_report(events, from_dt, to_dt, args.days)
    print(json.dumps(report, indent=2, ensure_ascii=False))
    print_human_report(report)
    return 0


if __name__ == "__main__":
    sys.exit(main())