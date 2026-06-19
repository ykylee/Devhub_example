"""Ingest API test (umbrella doc §2 FR §3.1 — FR-I-001 ~ FR-I-006)."""

from __future__ import annotations

import pytest
from fastapi.testclient import TestClient

from backend_knowledge.main import app


@pytest.fixture
def client(temp_var_dir) -> TestClient:
    """FastAPI test client with isolated var_dir."""
    return TestClient(app)


class TestIngestSync:
    """FR-I-001: POST /ingest/{source}/sync."""

    def test_sync_unknown_source_fails(self, client: TestClient) -> None:
        """Sync with unknown source should return 400 E_VALIDATION."""
        resp = client.post("/api/v0-2/ingest/unknown_source/sync")
        assert resp.status_code == 400
        body = resp.json()
        assert body["detail"]["code"] == "E_VALIDATION"
        assert "unknown source" in body["detail"]["message"]

    def test_sync_homelab_mock_dry_run(self, client: TestClient) -> None:
        """Sync homelab_mock with dry_run=true should return 0 synced (no emit)."""
        resp = client.post(
            "/api/v0-2/ingest/homelab_mock/sync",
            params={"dry_run": True},
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["dry_run"] is True
        assert data["synced"] == 0
        assert data["failed"] == 0

    def test_sync_homelab_mock_full(
        self, client: TestClient, path_y_header_value: str
    ) -> None:
        """Sync homelab_mock (full) should return 3 synced raw_ids."""
        resp = client.post(
            "/api/v0-2/ingest/homelab_mock/sync",
            headers={"X-DevHub-User-Context": path_y_header_value},
        )
        assert resp.status_code == 200
        body = resp.json()
        data = body["data"]
        assert data["synced"] == 3
        assert data["failed"] == 0
        assert len(data["raw_ids"]) == 3
        assert body["envelope"]["caller_user_id"] == "u_test_001"


class TestIngestStatus:
    """FR-I-002: GET /ingest/{source}/status."""

    def test_status_healthy(self, client: TestClient) -> None:
        """Status of homelab_mock should be healthy (mock always)."""
        resp = client.get("/api/v0-2/ingest/homelab_mock/status")
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["source"] == "homelab_mock"
        assert data["health"] == "healthy"

    def test_status_unknown_source_fails(self, client: TestClient) -> None:
        """Status of unknown source should return 400."""
        resp = client.get("/api/v0-2/ingest/unknown_source/status")
        assert resp.status_code == 400


class TestRawRegister:
    """FR-I-003: POST /raw (Path Y 필수)."""

    def test_register_without_path_y_fails(self, client: TestClient) -> None:
        """Register without Path Y should return 400."""
        resp = client.post(
            "/api/v0-2/raw",
            json={
                "type": "dataset",
                "name": "test_dataset",
                "source": "homelab_mock",
                "body": "# Test\n\nBody content",
            },
        )
        assert resp.status_code == 400
        body = resp.json()
        assert body["detail"]["code"] == "E_VALIDATION"

    def test_register_with_path_y(
        self, client: TestClient, path_y_header_value: str
    ) -> None:
        """Register with Path Y should return 201 with raw_id."""
        resp = client.post(
            "/api/v0-2/raw",
            headers={"X-DevHub-User-Context": path_y_header_value},
            json={
                "type": "dataset",
                "name": "test_dataset_001",
                "source": "homelab_mock",
                "body": "# Test Dataset\n\nTest body content",
            },
        )
        assert resp.status_code == 201
        data = resp.json()["data"]
        assert "raw_id" in data
        assert data["sha256"]
        assert data["envelope_encrypted"] is False  # KEK not set → plaintext mode
        assert data["size"] > 0
        assert "registered_at" in data

    def test_register_invalid_name_pattern(
        self, client: TestClient, path_y_header_value: str
    ) -> None:
        """Register with invalid name (uppercase) should return 422."""
        resp = client.post(
            "/api/v0-2/raw",
            headers={"X-DevHub-User-Context": path_y_header_value},
            json={
                "type": "dataset",
                "name": "InvalidNameWithUppercase",  # pattern violation
                "source": "homelab_mock",
                "body": "# Test",
            },
        )
        # Pydantic validation error → 422
        assert resp.status_code == 422


class TestRawList:
    """FR-I-005: GET /raw (Path Y 필수)."""

    def test_list_without_path_y_fails(self, client: TestClient) -> None:
        """List without Path Y should return 400."""
        resp = client.get("/api/v0-2/raw")
        assert resp.status_code == 400

    def test_list_with_path_y_empty(
        self, client: TestClient, path_y_header_value: str
    ) -> None:
        """List with Path Y (empty var/raw) should return 0 items."""
        resp = client.get(
            "/api/v0-2/raw",
            headers={"X-DevHub-User-Context": path_y_header_value},
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["total"] == 0
        assert data["items"] == []


class TestRawDelete:
    """FR-I-006: DELETE /raw/{id} (Path Y 필수)."""

    def test_delete_nonexistent_returns_404(
        self, client: TestClient, path_y_header_value: str
    ) -> None:
        """Delete non-existent raw should return 404."""
        resp = client.delete(
            "/api/v0-2/raw/nonexistent-raw-id",
            headers={"X-DevHub-User-Context": path_y_header_value},
        )
        assert resp.status_code == 404

    def test_delete_without_path_y_fails(self, client: TestClient) -> None:
        """Delete without Path Y should return 400."""
        resp = client.delete("/api/v0-2/raw/some-id")
        assert resp.status_code == 400
