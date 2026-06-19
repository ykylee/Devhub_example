"""Pytest configuration and fixtures (umbrella doc §3.6 + §3.8 정합)."""

from __future__ import annotations

import base64
import json
import tempfile
from datetime import datetime, timezone
from pathlib import Path

import pytest

from backend_knowledge.config import get_settings
from backend_knowledge.storage import get_raw_store


@pytest.fixture
def temp_var_dir(monkeypatch, tmp_path: Path) -> Path:
    """Override VAR_DIR to tmp_path for test isolation."""
    var_dir = tmp_path / "var"
    var_dir.mkdir(parents=True, exist_ok=True)
    settings = get_settings()
    monkeypatch.setattr(settings, "var_dir", var_dir)
    # Reset singleton raw_store so it picks up new VAR_DIR
    from backend_knowledge.storage import raw_store as raw_store_module
    raw_store_module._raw_store = None
    return var_dir


@pytest.fixture
def path_y_header_value() -> str:
    """Build a valid X-DevHub-User-Context header value (base64url(json))."""
    payload = {
        "version": "v0",
        "user_id": "u_test_001",
        "org_id": "ou_test_dept_a",
        "org_unit_ids": ["ou_test_dept_a", "ou_test_dept_b1"],
        "project_ids": ["prj_test_x"],
        "roles": ["developer", "project_leader:prj_test_x"],
        "request_id": "req_test_20260619_001",
        "issued_at": datetime.now(timezone.utc).isoformat(),
    }
    json_str = json.dumps(payload)
    return base64.urlsafe_b64encode(json_str.encode("utf-8")).decode("ascii").rstrip("=")


@pytest.fixture
def path_y_expired_header_value() -> str:
    """Build an expired X-DevHub-User-Context header (issued_at > 5분 전)."""
    payload = {
        "version": "v0",
        "user_id": "u_test_001",
        "org_id": "ou_test_dept_a",
        "org_unit_ids": ["ou_test_dept_a"],
        "project_ids": [],
        "roles": ["developer"],
        "request_id": "req_test_expired",
        "issued_at": "2020-01-01T00:00:00+00:00",  # 6년 전
    }
    json_str = json.dumps(payload)
    return base64.urlsafe_b64encode(json_str.encode("utf-8")).decode("ascii").rstrip("=")


@pytest.fixture
def path_y_invalid_header_value() -> str:
    """Build an invalid X-DevHub-User-Context header (base64 깨짐)."""
    return "this-is-not-base64url-json!!!"
