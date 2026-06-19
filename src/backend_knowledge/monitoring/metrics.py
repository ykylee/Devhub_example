"""18 PoC metrics — registry + computation (umbrella doc §11.3 + §3.6.6.3).

5 base metrics (§11.3):
1. bk_sync_success_rate (per source, 24h sliding)
2. bk_query_p95_latency_ms
3. bk_integrity_violation_rate (per day)
4. bk_pi_ingest_success_rate (1h sliding) [stub: M-v0.2.3+]
5. bk_archive_trigger_failures (per day) [stub: M-v0.2.3+]

13 governance metrics (§3.6.6.3, 4 layer × ~3 metric):
per user (4): active_users, curation_count, query_count, access_count
per org   (4): active_users_per_org, curation_count_per_org, query_count_per_org, access_count_per_org
per project (4): same structure, project_id dimension
per event type (1): audit_event_count_by_type (7 type × count)

Total: 18 metrics computed from audit log JSONL files.
"""

from __future__ import annotations

from collections import Counter, defaultdict
from datetime import datetime, timedelta, timezone
from pathlib import Path
from typing import Any

from ..audit.logger import get_audit_logger


def _read_audit_in_range(from_dt: datetime, to_dt: datetime) -> list[dict]:
    """Read all audit events in [from_dt, to_dt] range."""
    audit = get_audit_logger()
    return audit.read_range(from_date=from_dt, to_date=to_dt, limit=100_000)


def _calc_sync_success_rate(source: str, window_hours: int = 24) -> float:
    """M1: Source plugin sync 성공률 (§11.3 metric 1)."""
    from_dt = datetime.now(timezone.utc) - timedelta(hours=window_hours)
    events = _read_audit_in_range(from_dt, datetime.now(timezone.utc))
    # MOCK: PR 2 ingest does not emit per-source success events yet. Returns neutral 1.0.
    # PR 4 (M-v0.2.3+) will implement actual per-source sync event aggregation.
    relevant = [e for e in events if e.get("source") == source]
    if not relevant:
        return 1.0
    success = sum(1 for e in relevant if e.get("success", True))
    return round(success / len(relevant), 4)


def _calc_query_p95_latency_ms() -> float:
    """M2: Query API p95 latency (§11.3 metric 2)."""
    from_dt = datetime.now(timezone.utc) - timedelta(hours=1)
    events = _read_audit_in_range(from_dt, datetime.now(timezone.utc))
    query_events = [e for e in events if e.get("event") == "audit.query"]
    if not query_events:
        return 0.0
    latencies = sorted(float(e.get("response_time_ms", 0)) for e in query_events)
    p95_idx = int(len(latencies) * 0.95)
    return round(latencies[min(p95_idx, len(latencies) - 1)], 2)


def _calc_integrity_violation_rate(window_hours: int = 24) -> float:
    """M3: Raw 정합성 violation rate (§11.3 metric 3)."""
    from_dt = datetime.now(timezone.utc) - timedelta(hours=window_hours)
    events = _read_audit_in_range(from_dt, datetime.now(timezone.utc))
    total = sum(1 for e in events if e.get("event") in ("audit.raw.received", "audit.curation.edit"))
    violations = sum(1 for e in events if e.get("event") == "audit.raw.integrity_violation")
    if total == 0:
        return 0.0
    return round(violations / total, 6)


def _calc_pi_ingest_success_rate(window_hours: int = 1) -> float:
    """M4: Pi ingest pipeline success rate (§11.3 metric 4) — STUB for M-v0.2.3+."""
    return 1.0


def _calc_archive_trigger_failures() -> int:
    """M5: Concept archive trigger failures (§11.3 metric 5) — STUB for M-v0.2.3+."""
    return 0


def _calc_governance_per_user(from_dt: datetime) -> dict[str, dict[str, int]]:
    """13 governance metrics (§3.6.6.3) per user."""
    events = _read_audit_in_range(from_dt, datetime.now(timezone.utc))
    metrics: dict[str, dict[str, int]] = defaultdict(lambda: {
        "active_logins": 0,
        "curation_count": 0,
        "query_count": 0,
        "access_count": 0,
    })
    for e in events:
        uid = e.get("user_id")
        if not uid:
            continue
        evt = e.get("event", "")
        if evt == "audit.user.login":
            metrics[uid]["active_logins"] += 1
        elif evt == "audit.curation.edit":
            metrics[uid]["curation_count"] += 1
        elif evt == "audit.query":
            metrics[uid]["query_count"] += 1
        elif evt == "audit.concept.access":
            metrics[uid]["access_count"] += 1
    return dict(metrics)


def _calc_governance_per_org(from_dt: datetime) -> dict[str, dict[str, int]]:
    """Per-org metrics."""
    events = _read_audit_in_range(from_dt, datetime.now(timezone.utc))
    metrics: dict[str, dict[str, int]] = defaultdict(lambda: {
        "active_logins": 0,
        "curation_count": 0,
        "query_count": 0,
        "access_count": 0,
    })
    seen: dict[tuple[str, str], bool] = set()
    for e in events:
        oid = e.get("org_id")
        uid = e.get("user_id")
        if not oid or not uid:
            continue
        evt = e.get("event", "")
        if evt == "audit.user.login" and (uid, oid) not in seen:
            seen[(uid, oid)] = True
            metrics[oid]["active_logins"] += 1
        elif evt == "audit.curation.edit":
            metrics[oid]["curation_count"] += 1
        elif evt == "audit.query":
            metrics[oid]["query_count"] += 1
        elif evt == "audit.concept.access":
            metrics[oid]["access_count"] += 1
    return dict(metrics)


def _calc_event_type_counts(from_dt: datetime) -> dict[str, int]:
    """Per event type count."""
    events = _read_audit_in_range(from_dt, datetime.now(timezone.utc))
    counter: Counter = Counter()
    for e in events:
        counter[e.get("event", "unknown")] += 1
    return dict(counter)


def collect_all_metrics() -> dict[str, Any]:
    """Collect all 18 PoC metrics in a single snapshot.

    Returns:
        {
            "base": {5 base metrics},
            "governance_user": {user_id: {4 metrics}},
            "governance_org": {org_id: {4 metrics}},
            "governance_event_type": {event_name: count},
            "generated_at": iso8601 timestamp,
            "metrics_version": "v0.2.0",
        }
    """
    day_ago = datetime.now(timezone.utc) - timedelta(hours=24)
    return {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "metrics_version": "v0.2.0-poc",
        "base": {
            "bk_sync_success_rate_homelab_mock_24h": _calc_sync_success_rate("homelab_mock"),
            "bk_query_p95_latency_ms_1h": _calc_query_p95_latency_ms(),
            "bk_integrity_violation_rate_24h": _calc_integrity_violation_rate(),
            "bk_pi_ingest_success_rate_1h": _calc_pi_ingest_success_rate(),
            "bk_archive_trigger_failures_24h": _calc_archive_trigger_failures(),
        },
        "governance_user": _calc_governance_per_user(day_ago),
        "governance_org": _calc_governance_per_org(day_ago),
        "governance_event_type": _calc_event_type_counts(day_ago),
        "stubbed_for_m_v0_2_3_plus": {
            "bk_pi_link_resolve_mttr_seconds": "M-v0.2.3+",
            "bk_pi_link_resolve_accuracy": "M-v0.2.3+",
            "bk_pi_link_false_positive_rate": "M-v0.2.3+",
            "bk_pi_sdk_timeout_rate": "M-v0.2.3+",
            "bk_pi_llm_recommendation_count_daily": "M-v0.2.3+",
            "bk_pi_link_false_positive_24h": "M-v0.2.3+",
            "bk_api_v0_3_request_count": "M-v0.3.0+",
            "bk_api_v0_3_error_rate": "M-v0.3.0+",
            "bk_api_v0_3_client_identification": "M-v0.3.0+",
            "bk_api_v0_3_migration_progress": "M-v0.3.0+",
        },
    }
