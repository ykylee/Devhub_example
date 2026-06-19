"""Health endpoint test (umbrella doc §3.1 API 매트릭스 row 1)."""

from __future__ import annotations

import pytest
from fastapi.testclient import TestClient

from backend_knowledge.main import app


@pytest.fixture
def client() -> TestClient:
    """FastAPI test client."""
    return TestClient(app)


class TestHealth:
    """Health endpoint (/health, /health/protected)."""

    def test_health_public(self, client: TestClient) -> None:
        """GET /health should return 200 OK (no Path Y required)."""
        resp = client.get("/health")
        assert resp.status_code == 200
        body = resp.json()
        assert body["envelope"]["api_version"] == "v0-2"
        assert body["data"]["status"] == "ok"
        assert body["data"]["version"] == "0.2.0"
        assert body["data"]["path_y_validated"] is False

    def test_health_protected_without_path_y_fails(self, client: TestClient) -> None:
        """GET /health/protected without Path Y should return 400."""
        resp = client.get("/health/protected")
        assert resp.status_code == 400
        body = resp.json()
        assert body["detail"]["code"] == "E_VALIDATION"

    def test_health_protected_with_valid_path_y(
        self, client: TestClient, path_y_header_value: str
    ) -> None:
        """GET /health/protected with valid Path Y should return 200 with caller info."""
        resp = client.get(
            "/health/protected",
            headers={"X-DevHub-User-Context": path_y_header_value},
        )
        assert resp.status_code == 200
        body = resp.json()
        assert body["data"]["status"] == "ok"
        assert body["data"]["user_id"] == "u_test_001"
        assert body["data"]["org_id"] == "ou_test_dept_a"
        assert "developer" in body["data"]["roles"]
        assert body["envelope"]["path_y_validated"] is True
        assert body["envelope"]["caller_user_id"] == "u_test_001"

    def test_health_protected_with_expired_path_y(
        self, client: TestClient, path_y_expired_header_value: str
    ) -> None:
        """GET /health/protected with expired Path Y should return 401."""
        resp = client.get(
            "/health/protected",
            headers={"X-DevHub-User-Context": path_y_expired_header_value},
        )
        assert resp.status_code == 401
        body = resp.json()
        assert body["detail"]["code"] == "E_UNAUTHORIZED"

    def test_health_protected_with_invalid_path_y(
        self, client: TestClient, path_y_invalid_header_value: str
    ) -> None:
        """GET /health/protected with invalid Path Y should return 400."""
        resp = client.get(
            "/health/protected",
            headers={"X-DevHub-User-Context": path_y_invalid_header_value},
        )
        assert resp.status_code == 400
        body = resp.json()
        assert body["detail"]["code"] == "E_VALIDATION"
