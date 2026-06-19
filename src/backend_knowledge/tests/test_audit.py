"""Audit log unit test (umbrella doc §3.6.6.1)."""

from __future__ import annotations

import json

import pytest
from fastapi.testclient import TestClient

from backend_knowledge.audit.events import AuditEvent, AuditEventType, build_audit_event
from backend_knowledge.audit.logger import AuditLogger, get_audit_logger


class TestAuditEvent:
    """7 audit event types per §3.6.6.1."""

    def test_seven_event_types_defined(self) -> None:
        assert len(AuditEventType) == 7

    @pytest.mark.parametrize("event_name", [
        "audit.user.login",
        "audit.concept.access",
        "audit.curation.edit",
        "audit.query",
        "audit.concept.archive",
        "audit.concept.publish",
        "audit.config.change",
    ])
    def test_all_seven_event_names(self, event_name: str) -> None:
        assert AuditEventType(event_name).value == event_name

    def test_build_audit_event_common_fields(self) -> None:
        event = build_audit_event(
            event_type=AuditEventType.USER_LOGIN,
            user_id="u_test",
            org_id="ou_test",
            request_id="req_test",
            ip="10.0.0.1",
            success=True,
            roles=["developer"],
        )
        assert event.event == "audit.user.login"
        assert event.user_id == "u_test"
        assert event.org_id == "ou_test"
        assert event.success is True
        assert event.roles == ["developer"]


class TestAuditLogger:
    """JSON Lines writer with daily rotation."""

    def test_emit_creates_jsonl_file(self, temp_var_dir) -> None:
        logger = AuditLogger(base_dir=temp_var_dir / "audit")
        logger.emit_simple(event_type=AuditEventType.USER_LOGIN, user_id="u_1", success=True)
        files = list((temp_var_dir / "audit").glob("audit-*.jsonl"))
        assert len(files) == 1
        content = files[0].read_text(encoding="utf-8")
        entry = json.loads(content.strip())
        assert entry["event"] == "audit.user.login"
        assert entry["user_id"] == "u_1"

    def test_emit_multiple_events_appends(self, temp_var_dir) -> None:
        logger = AuditLogger(base_dir=temp_var_dir / "audit")
        for i in range(5):
            logger.emit_simple(event_type=AuditEventType.QUERY, user_id=f"u_{i}")
        files = list((temp_var_dir / "audit").glob("audit-*.jsonl"))
        content = files[0].read_text(encoding="utf-8").strip().split("\n")
        assert len(content) == 5

    def test_read_range_returns_newest_first(self, temp_var_dir) -> None:
        logger = AuditLogger(base_dir=temp_var_dir / "audit")
        for i in range(3):
            logger.emit_simple(event_type=AuditEventType.QUERY, user_id=f"u_{i}")
        items = logger.read_range(limit=10)
        assert len(items) == 3
        assert items[0]["user_id"] == "u_2"
        assert items[-1]["user_id"] == "u_0"

    def test_read_range_filter_by_event_type(self, temp_var_dir) -> None:
        logger = AuditLogger(base_dir=temp_var_dir / "audit")
        logger.emit_simple(event_type=AuditEventType.USER_LOGIN, user_id="u_1")
        logger.emit_simple(event_type=AuditEventType.QUERY, user_id="u_1")
        logger.emit_simple(event_type=AuditEventType.QUERY, user_id="u_2")
        items = logger.read_range(event_type=AuditEventType.QUERY, limit=10)
        assert len(items) == 2
        assert all(i["event"] == "audit.query" for i in items)

    def test_read_range_filter_by_user_id(self, temp_var_dir) -> None:
        logger = AuditLogger(base_dir=temp_var_dir / "audit")
        logger.emit_simple(event_type=AuditEventType.QUERY, user_id="u_alice")
        logger.emit_simple(event_type=AuditEventType.QUERY, user_id="u_bob")
        items = logger.read_range(user_id="u_alice", limit=10)
        assert len(items) == 1
        assert items[0]["user_id"] == "u_alice"

    def test_read_range_limit(self, temp_var_dir) -> None:
        logger = AuditLogger(base_dir=temp_var_dir / "audit")
        for i in range(20):
            logger.emit_simple(event_type=AuditEventType.QUERY, user_id=f"u_{i}")
        items = logger.read_range(limit=5)
        assert len(items) == 5

    def test_cleanup_old_removes_files(self, temp_var_dir) -> None:
        import os
        from datetime import datetime, timedelta, timezone
        old_date = (datetime.now(timezone.utc) - timedelta(days=10)).strftime("%Y-%m-%d")
        audit_dir = temp_var_dir / "audit"
        audit_dir.mkdir(parents=True, exist_ok=True)
        old_file = audit_dir / f"audit-{old_date}.jsonl"
        old_file.write_text('{"event":"audit.user.login","timestamp":"2020-01-01T00:00:00Z"}', encoding="utf-8")
        logger = AuditLogger(base_dir=temp_var_dir / "audit", retention_days=7)
        deleted = logger.cleanup_old()
        assert deleted == 1
        assert not old_file.exists()

    def test_singleton_returns_same_instance(self, temp_var_dir) -> None:
        from backend_knowledge.audit.logger import _audit_logger
        first = get_audit_logger()
        second = get_audit_logger()
        assert first is second


class TestAuditApi:
    """4 audit log viewer endpoints."""

    @pytest.fixture
    def client(self, temp_var_dir) -> TestClient:
        from fastapi.testclient import TestClient
        from backend_knowledge.main import app
        return TestClient(app)

    def test_list_audit(self, client) -> None:
        resp = client.get("/api/v0-2/audit")
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert "items" in data

    def test_list_audit_invalid_event_type(self, client) -> None:
        resp = client.get("/api/v0-2/audit", params={"event_type": "invalid.event"})
        assert resp.status_code == 400

    def test_user_audit_self_view(self, client, path_y_header_value) -> None:
        resp = client.get(
            "/api/v0-2/audit/user/u_test_001",
            headers={"X-DevHub-User-Context": path_y_header_value},
        )
        assert resp.status_code == 200

    def test_user_audit_other_user_denied(self, client, path_y_header_value) -> None:
        resp = client.get(
            "/api/v0-2/audit/user/u_someone_else",
            headers={"X-DevHub-User-Context": path_y_header_value},
        )
        assert resp.status_code == 403
