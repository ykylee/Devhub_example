"""scripts/analyze_incidents.py unit tests (umbrella doc §13.2 row 3).

mock audit log JSONL 파일을 생성하여 tuning 추천 로직 검증.
"""
from __future__ import annotations

import json
import sys
from datetime import datetime, timedelta, timezone
from pathlib import Path

# scripts/ 디렉터리를 sys.path 에 추가
sys.path.insert(0, str(Path(__file__).parent.parent / "scripts"))

from analyze_incidents import (  # noqa: E402
    INCIDENT_TYPES,
    count_by_type,
    generate_report,
    load_audit_events,
    recommend_thresholds,
)


def write_audit_file(audit_dir: Path, date: datetime, events: list[dict]) -> Path:
    """mock audit log JSONL 파일 생성."""
    path = audit_dir / f"audit-{date.strftime('%Y-%m-%d')}.jsonl"
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as f:
        for e in events:
            f.write(json.dumps(e) + "\n")
    return path


def make_event(event_type: str, success: bool = True, **kwargs) -> dict:
    """mock audit event 생성."""
    return {
        "event": event_type,
        "success": success,
        "timestamp": datetime.now(timezone.utc).isoformat(),
        **kwargs,
    }


class TestLoadAuditEvents:
    def test_empty_dir_returns_empty(self, tmp_path: Path):
        events = load_audit_events(tmp_path, datetime.now(timezone.utc) - timedelta(days=1), datetime.now(timezone.utc))
        assert events == []

    def test_single_file_loaded(self, tmp_path: Path):
        today = datetime.now(timezone.utc)
        write_audit_file(tmp_path, today, [make_event("audit.user.login")])
        events = load_audit_events(tmp_path, today, today)
        assert len(events) == 1
        assert events[0]["event"] == "audit.user.login"

    def test_skips_corrupted_jsonl(self, tmp_path: Path):
        today = datetime.now(timezone.utc)
        path = write_audit_file(tmp_path, today, [make_event("audit.user.login")])
        with path.open("a", encoding="utf-8") as f:
            f.write("not-json{\n")
            f.write("\n")
        events = load_audit_events(tmp_path, today, today)
        assert len(events) == 1

    def test_filters_by_date_range(self, tmp_path: Path):
        today = datetime.now(timezone.utc)
        old = today - timedelta(days=10)
        write_audit_file(tmp_path, today, [make_event("audit.user.login")])
        write_audit_file(tmp_path, old, [make_event("audit.user.login"), make_event("audit.user.login")])
        events = load_audit_events(tmp_path, today - timedelta(days=1), today)
        assert len(events) == 1


class TestCountByType:
    def test_empty_events(self):
        result = count_by_type([])
        assert result["sync_failure"]["incidents"] == 0

    def test_sync_failure_count(self):
        events = [
            make_event("audit.raw.received", success=False),
            make_event("audit.raw.received", success=True),
        ]
        result = count_by_type(events)
        assert result["sync_failure"]["total_events"] == 2
        assert result["sync_failure"]["failures"] == 1
        assert result["sync_failure"]["incidents"] == 1

    def test_credential_expired_401(self):
        events = [
            make_event("audit.raw.received", success=False, last_error="401 Unauthorized"),
            make_event("audit.raw.received", success=False, last_error="500 Server Error"),
        ]
        result = count_by_type(events)
        assert result["credential_expired"]["incidents"] == 1

    def test_integrity_violation(self):
        events = [make_event("audit.raw.integrity_violation")]
        result = count_by_type(events)
        assert result["integrity_violation"]["total_events"] == 1


class TestRecommendThresholds:
    def test_sync_failure_maintain(self):
        stats = {"incidents": 1, "total_events": 100, "failures": 1}
        rec = recommend_thresholds("sync_failure", stats, window_days=7)
        assert rec["recommendation"] == "maintain"
        assert rec["current_warning"] == 0.99

    def test_sync_failure_tighten_high_frequency(self):
        stats = {"incidents": 30, "total_events": 100, "failures": 30}
        rec = recommend_thresholds("sync_failure", stats, window_days=7)
        assert rec["recommendation"] == "tighten"
        assert rec["recommended_warning"] < 0.99

    def test_sync_failure_loosen_low_frequency(self):
        stats = {"incidents": 1, "total_events": 1000, "failures": 1}
        rec = recommend_thresholds("sync_failure", stats, window_days=14)
        assert rec["recommendation"] == "loosen"
        assert rec["recommended_warning"] > 0.99

    def test_credential_expired_tighten(self):
        stats = {"incidents": 20, "total_events": 100, "failures": 20}
        rec = recommend_thresholds("credential_expired", stats, window_days=7)
        assert rec["recommendation"] == "tighten"
        assert rec["recommended_warning"] == 0

    def test_archive_trigger_maintain(self):
        stats = {"incidents": 0, "total_events": 100, "failures": 0}
        rec = recommend_thresholds("archive_trigger_failure", stats, window_days=7)
        assert rec["recommendation"] == "maintain"

    def test_integrity_violation_tighten(self):
        stats = {"incidents": 5, "total_events": 100, "failures": 5}
        rec = recommend_thresholds("integrity_violation", stats, window_days=7)
        assert rec["recommendation"] == "tighten"


class TestGenerateReport:
    def test_empty_report(self):
        to_dt = datetime.now(timezone.utc)
        from_dt = to_dt - timedelta(days=7)
        report = generate_report([], from_dt, to_dt, 7)
        assert report["total_events"] == 0
        assert report["summary"]["total_incidents"] == 0
        assert report["summary"]["types_total"] == 7

    def test_report_with_incidents(self):
        to_dt = datetime.now(timezone.utc)
        from_dt = to_dt - timedelta(days=7)
        events = [
            make_event("audit.raw.received", success=False),
            make_event("audit.raw.integrity_violation"),
            make_event("audit.concept.archive", success=False),
        ]
        report = generate_report(events, from_dt, to_dt, 7)
        assert report["total_events"] == 3
        assert report["summary"]["total_incidents"] >= 3

    def test_report_recommendations_per_type(self):
        to_dt = datetime.now(timezone.utc)
        from_dt = to_dt - timedelta(days=7)
        events = [make_event("audit.raw.received", success=False)]
        report = generate_report(events, from_dt, to_dt, 7)
        assert "sync_failure" in report["recommendations"]
        assert "current_warning" in report["recommendations"]["sync_failure"]


class TestIncidentTypesConfig:
    def test_all_7_incident_types_defined(self):
        expected = {
            "sync_failure",
            "credential_expired",
            "pi_ingest_timeout",
            "retention_cron_failure",
            "integrity_violation",
            "archive_trigger_failure",
            "stale_link_detected",
        }
        assert set(INCIDENT_TYPES.keys()) == expected

    def test_all_have_runbook_section(self):
        for type_name, config in INCIDENT_TYPES.items():
            assert config["runbook"].startswith("§11.1.")
            assert len(config["audit_events"]) > 0
            assert config["current_threshold_warning"] is not None
            assert config["current_threshold_critical"] is not None