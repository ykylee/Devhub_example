"""Prometheus exposition format for /metrics endpoint (umbrella doc §11.3 + §2.4 item 8).

3 tier alert thresholds:
- info: Slack #backend-knowledge-info (M-v0.2.1+)
- warning: Slack #backend-knowledge-alerts + on-call responder
- critical: Slack #backend-knowledge-critical + on-call page (15 min)
"""

from __future__ import annotations

from datetime import datetime, timezone
from typing import Any

from .metrics import collect_all_metrics


# 3-tier alert thresholds (umbrella doc §11.3)
THRESHOLDS: dict[str, dict[str, tuple[float | None, float | None]]] = {
    "bk_sync_success_rate_homelab_mock_24h": {
        "warning": (0.99, None),
        "critical": (0.95, None),
    },
    "bk_query_p95_latency_ms_1h": {
        "warning": (None, 500.0),
        "critical": (None, 1000.0),
    },
    "bk_integrity_violation_rate_24h": {
        "warning": (None, 0.0001),
        "critical": (None, 0.001),
    },
    "bk_archive_trigger_failures_24h": {
        "warning": (None, 0),
        "critical": (None, 5),
    },
}


def evaluate_alert(metric_name: str, value: float) -> str:
    """Return severity tier for a given metric value.

    Returns: "ok" | "info" | "warning" | "critical"
    """
    if metric_name not in THRESHOLDS:
        return "ok"
    thresholds = THRESHOLDS[metric_name]
    crit_lower, crit_upper = thresholds.get("critical", (None, None))
    warn_lower, warn_upper = thresholds.get("warning", (None, None))

    if crit_lower is not None and value < crit_lower:
        return "critical"
    if crit_upper is not None and value > crit_upper:
        return "critical"
    if warn_lower is not None and value < warn_lower:
        return "warning"
    if warn_upper is not None and value > warn_upper:
        return "warning"
    return "ok"


def collect_alerts() -> list[dict[str, Any]]:
    """Evaluate all 5 base metrics against 3-tier thresholds, return fired alerts.

    Returns list of: {metric, value, severity, threshold, message, evaluated_at}
    """
    snapshot = collect_all_metrics()
    base = snapshot["base"]
    alerts: list[dict[str, Any]] = []
    now = datetime.now(timezone.utc).isoformat()
    for metric_name, value in base.items():
        severity = evaluate_alert(metric_name, value)
        if severity == "ok":
            continue
        threshold = THRESHOLDS.get(metric_name, {}).get(severity, (None, None))
        alerts.append({
            "metric": metric_name,
            "value": value,
            "severity": severity,
            "threshold": list(threshold) if threshold != (None, None) else None,
            "message": f"{metric_name} = {value} (severity: {severity}, threshold: {threshold})",
            "evaluated_at": now,
        })
    return alerts


def render_prometheus_exposition(snapshot: dict[str, Any] | None = None) -> str:
    """Render metrics snapshot in Prometheus text exposition format (v0.0.4).

    Returns plain-text response body for /metrics endpoint.
    """
    if snapshot is None:
        snapshot = collect_all_metrics()
    base = snapshot["base"]
    lines: list[str] = []
    lines.append("# HELP backend_knowledge_info backend-knowledge metadata")
    lines.append("# TYPE backend_knowledge_info gauge")
    lines.append(f'backend_knowledge_info{{version="v0.2.0-poc"}} 1')
    lines.append("")

    for metric_name, value in base.items():
        if isinstance(value, (int, float)):
            severity = evaluate_alert(metric_name, float(value))
            labels = f'{{severity="{severity}"}}'
            lines.append(f"# HELP {metric_name} PoC base metric")
            lines.append(f"# TYPE {metric_name} gauge")
            lines.append(f"{metric_name}{labels} {value}")
            lines.append("")

    governance_event = snapshot.get("governance_event_type", {})
    if governance_event:
        lines.append("# HELP bk_audit_event_count_total Audit log event count by type (24h)")
        lines.append("# TYPE bk_audit_event_count_total counter")
        for event_name, count in sorted(governance_event.items()):
            lines.append(f'bk_audit_event_count_total{{event="{event_name}"}} {count}')
        lines.append("")

    governance_user = snapshot.get("governance_user", {})
    if governance_user:
        lines.append("# HELP bk_user_activity_count Total user activity events (24h)")
        lines.append("# TYPE bk_user_activity_count counter")
        for user_id, counts in sorted(governance_user.items()):
            for kind, count in counts.items():
                if count > 0:
                    lines.append(f'bk_user_activity_count{{user_id="{user_id}",kind="{kind}"}} {count}')
        lines.append("")

    governance_org = snapshot.get("governance_org", {})
    if governance_org:
        lines.append("# HELP bk_org_activity_count Total org activity events (24h)")
        lines.append("# TYPE bk_org_activity_count counter")
        for org_id, counts in sorted(governance_org.items()):
            for kind, count in counts.items():
                if count > 0:
                    lines.append(f'bk_org_activity_count{{org_id="{org_id}",kind="{kind}"}} {count}')

    return "\n".join(lines) + "\n"
