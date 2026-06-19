"""Monitoring + alert unit test (umbrella doc §11.3 + §3.6.6.3)."""

from __future__ import annotations

import pytest
from fastapi.testclient import TestClient

from backend_knowledge.monitoring.metrics import collect_all_metrics
from backend_knowledge.monitoring.prometheus import (
    THRESHOLDS,
    collect_alerts,
    evaluate_alert,
    render_prometheus_exposition,
)


class TestMetricsCollection:
    """18 PoC metrics (5 base + 13 governance)."""

    def test_collect_returns_base_metrics(self, temp_var_dir) -> None:
        snapshot = collect_all_metrics()
        assert "base" in snapshot
        base = snapshot["base"]
        assert "bk_sync_success_rate_homelab_mock_24h" in base
        assert "bk_query_p95_latency_ms_1h" in base
        assert "bk_integrity_violation_rate_24h" in base
        assert "bk_pi_ingest_success_rate_1h" in base
        assert "bk_archive_trigger_failures_24h" in base

    def test_collect_returns_governance_metrics(self, temp_var_dir) -> None:
        snapshot = collect_all_metrics()
        assert "governance_user" in snapshot
        assert "governance_org" in snapshot
        assert "governance_event_type" in snapshot

    def test_collect_stubs_m_v0_2_3_plus(self, temp_var_dir) -> None:
        snapshot = collect_all_metrics()
        assert "stubbed_for_m_v0_2_3_plus" in snapshot
        stubs = snapshot["stubbed_for_m_v0_2_3_plus"]
        assert len(stubs) == 10

    def test_governance_event_type_empty_when_no_events(self, temp_var_dir) -> None:
        snapshot = collect_all_metrics()
        assert snapshot["governance_event_type"] == {}
        assert snapshot["governance_user"] == {}


class TestAlertEvaluation:
    """3-tier alert threshold evaluation."""

    def test_thresholds_defined_for_base_metrics(self) -> None:
        assert "bk_sync_success_rate_homelab_mock_24h" in THRESHOLDS
        assert "bk_query_p95_latency_ms_1h" in THRESHOLDS
        assert "bk_integrity_violation_rate_24h" in THRESHOLDS
        assert "bk_archive_trigger_failures_24h" in THRESHOLDS

    def test_sync_success_rate_critical_below_95(self) -> None:
        assert evaluate_alert("bk_sync_success_rate_homelab_mock_24h", 0.90) == "critical"
        assert evaluate_alert("bk_sync_success_rate_homelab_mock_24h", 0.94) == "critical"

    def test_sync_success_rate_warning_between_95_and_99(self) -> None:
        assert evaluate_alert("bk_sync_success_rate_homelab_mock_24h", 0.97) == "warning"
        assert evaluate_alert("bk_sync_success_rate_homelab_mock_24h", 0.98) == "warning"

    def test_sync_success_rate_ok_above_99(self) -> None:
        assert evaluate_alert("bk_sync_success_rate_homelab_mock_24h", 0.995) == "ok"
        assert evaluate_alert("bk_sync_success_rate_homelab_mock_24h", 1.0) == "ok"

    def test_query_p95_latency_critical_above_1000(self) -> None:
        assert evaluate_alert("bk_query_p95_latency_ms_1h", 1500.0) == "critical"

    def test_query_p95_latency_warning_between_500_and_1000(self) -> None:
        assert evaluate_alert("bk_query_p95_latency_ms_1h", 750.0) == "warning"

    def test_query_p95_latency_ok_below_500(self) -> None:
        assert evaluate_alert("bk_query_p95_latency_ms_1h", 250.0) == "ok"

    def test_archive_trigger_critical_above_5(self) -> None:
        assert evaluate_alert("bk_archive_trigger_failures_24h", 10) == "critical"

    def test_archive_trigger_warning_at_1(self) -> None:
        assert evaluate_alert("bk_archive_trigger_failures_24h", 1) == "warning"

    def test_archive_trigger_ok_at_0(self) -> None:
        assert evaluate_alert("bk_archive_trigger_failures_24h", 0) == "ok"

    def test_collect_alerts_empty_on_clean_metrics(self, temp_var_dir) -> None:
        from unittest.mock import patch
        with patch("backend_knowledge.monitoring.metrics._calc_sync_success_rate", return_value=1.0), \
             patch("backend_knowledge.monitoring.metrics._calc_query_p95_latency_ms", return_value=10.0), \
             patch("backend_knowledge.monitoring.metrics._calc_integrity_violation_rate", return_value=0.0), \
             patch("backend_knowledge.monitoring.metrics._calc_archive_trigger_failures", return_value=0):
            alerts = collect_alerts()
        assert alerts == []


class TestPrometheusExposition:
    """Prometheus text format v0.0.4."""

    @pytest.fixture
    def client(self, temp_var_dir) -> TestClient:
        from fastapi.testclient import TestClient
        from backend_knowledge.main import app
        return TestClient(app)

    def test_render_returns_text_with_help_lines(self, temp_var_dir) -> None:
        from unittest.mock import patch
        with patch("backend_knowledge.monitoring.metrics._calc_sync_success_rate", return_value=1.0), \
             patch("backend_knowledge.monitoring.metrics._calc_query_p95_latency_ms", return_value=10.0), \
             patch("backend_knowledge.monitoring.metrics._calc_integrity_violation_rate", return_value=0.0), \
             patch("backend_knowledge.monitoring.metrics._calc_archive_trigger_failures", return_value=0):
            output = render_prometheus_exposition()
        assert "# HELP" in output
        assert "# TYPE" in output
        assert "backend_knowledge_info" in output
        assert "bk_sync_success_rate_homelab_mock_24h" in output

    def test_render_includes_severity_label(self, temp_var_dir) -> None:
        from unittest.mock import patch
        with patch("backend_knowledge.monitoring.metrics._calc_sync_success_rate", return_value=0.90), \
             patch("backend_knowledge.monitoring.metrics._calc_query_p95_latency_ms", return_value=10.0), \
             patch("backend_knowledge.monitoring.metrics._calc_integrity_violation_rate", return_value=0.0), \
             patch("backend_knowledge.monitoring.metrics._calc_archive_trigger_failures", return_value=0):
            output = render_prometheus_exposition()
        assert 'severity="critical"' in output

    def test_metrics_endpoint_disabled_by_default(self, client) -> None:
        resp = client.get("/metrics")
        assert resp.status_code == 200
        assert "disabled" in resp.text.lower()

    def test_alerts_endpoint_returns_envelope(self, client) -> None:
        resp = client.get("/api/v0-2/monitoring/alerts")
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert "alerts" in data
        assert "by_severity" in data
