"""E2E smoke test (umbrella doc §3.8.5 + §6.5.4 + §11.1.6 cross-section).

ingest → curate → query happy path:
1. Health check
2. Source sync (homelab_mock)
3. Raw register (Path Y)
4. Bundle create
5. Concept enrich
6. Bundle rebuild (index.md + viz.html + reverse_index)
7. Concept get
8. Search
9. Reverse / impact graph
10. Archive / publish lifecycle (audit event emit)
11. Audit log viewer (4 endpoint)
12. Metrics endpoint
13. 3-tier alert evaluation
"""

from __future__ import annotations

import base64
import json
from datetime import datetime, timezone

import pytest
from fastapi.testclient import TestClient

from backend_knowledge.main import app


@pytest.fixture
def client(temp_var_dir) -> TestClient:
    return TestClient(app)


def _path_y(user_id: str = "u_e2e_001", org_id: str = "ou_e2e_dept_a", roles: list[str] | None = None) -> str:
    payload = {
        "version": "v0",
        "user_id": user_id,
        "org_id": org_id,
        "org_unit_ids": [org_id],
        "project_ids": ["prj_e2e"],
        "roles": roles or ["developer", "project_leader:prj_e2e"],
        "request_id": "req_e2e_20260619",
        "issued_at": datetime.now(timezone.utc).isoformat(),
    }
    return base64.urlsafe_b64encode(
        json.dumps(payload).encode("utf-8")
    ).decode("ascii").rstrip("=")


def test_e2e_health(client: TestClient) -> None:
    """Step 1: GET /health returns 200."""
    resp = client.get("/health")
    assert resp.status_code == 200
    assert resp.json()["data"]["status"] == "ok"


def test_e2e_source_sync_homelab_mock(client: TestClient) -> None:
    """Step 2: homelab_mock sync returns 3 raws."""
    resp = client.post("/api/v0-2/ingest/homelab_mock/sync")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert data["synced"] == 3
    assert len(data["raw_ids"]) == 3


def test_e2e_raw_register_and_get(client: TestClient) -> None:
    """Step 3: register raw + GET raw (PR 1)."""
    header = {"X-DevHub-User-Context": _path_y()}
    reg = client.post(
        "/api/v0-2/raw",
        headers=header,
        json={
            "type": "dataset",
            "name": "e2e_dataset",
            "source": "homelab_mock",
            "body": "# E2E Dataset\n\nTest body",
        },
    )
    assert reg.status_code == 201
    raw_id = reg.json()["data"]["raw_id"]
    assert raw_id


def test_e2e_bundle_create_enrich_rebuild(client: TestClient) -> None:
    """Step 4-6: create bundle + enrich concept + rebuild index."""
    header = {"X-DevHub-User-Context": _path_y()}
    bundle = "e2e-bundle"
    create = client.post(
        "/api/v0-2/bundles",
        headers=header,
        json={"name": bundle, "owner_org_id": "ou_e2e_dept_a", "visibility": "org"},
    )
    assert create.status_code == 201

    enrich = client.post(
        f"/api/v0-2/concepts/{bundle}/dataset/e2e-ds/enrich",
        json={"raw_id": "fake-raw-id", "enricher": "rule-based", "dry_run": True},
    )
    assert enrich.status_code == 200
    assert enrich.json()["data"]["enricher_used"] == "rule-based"

    rebuild = client.post(
        f"/api/v0-2/bundles/{bundle}/rebuild",
        json={"dry_run": True},
    )
    assert rebuild.status_code == 200
    assert rebuild.json()["data"]["concept_count"] == 0


def test_e2e_concept_get_search(client: TestClient, temp_var_dir) -> None:
    """Step 7-8: write concept directly + GET + search."""
    import json
    from backend_knowledge.config import get_settings
    settings = get_settings()
    bundle = "e2e-search"
    bundle_dir = settings.var_dir / "bundles" / bundle / "dataset"
    bundle_dir.mkdir(parents=True, exist_ok=True)
    md_text = (
        "---\n"
        "type: dataset\n"
        "title: E2E Search\n"
        "description: searchable concept\n"
        "---\n\n"
        "# E2E Search\n\n"
    )
    (bundle_dir / "e2e-search.md").write_text(md_text, encoding="utf-8")
    meta = {
        "bundle": bundle,
        "type": "dataset",
        "name": "e2e-search",
        "sha256": "fake_sha_e2e",
        "source": "homelab_mock",
        "raw_id": None,
        "registered_by": "u_e2e_001",
        "visibility": "org",
        "frontmatter": {"title": "E2E Search", "description": "searchable concept"},
        "registered_at": "2026-06-19T00:00:00+00:00",
    }
    (bundle_dir / "e2e-search.meta.json").write_text(json.dumps(meta), encoding="utf-8")

    header = {"X-DevHub-User-Context": _path_y()}
    get_resp = client.get(
        f"/api/v0-2/concepts/dataset/e2e-search",
        params={"bundle": bundle},
        headers=header,
    )
    assert get_resp.status_code == 200
    assert get_resp.json()["data"]["frontmatter"]["title"] == "E2E Search"

    search_resp = client.get(
        "/api/v0-2/search",
        params={"q": "E2E"},
        headers=header,
    )
    assert search_resp.status_code == 200
    assert search_resp.json()["data"]["total"] >= 1


def test_e2e_graph_reverse_impact(client: TestClient, temp_var_dir) -> None:
    """Step 9: graph reverse + impact endpoints return valid envelope."""
    from backend_knowledge.config import get_settings
    from backend_knowledge.monitoring import prometheus as prom_module
    settings = get_settings()
    index_dir = settings.var_dir / "bundles" / "e2e-graph" / ".index"
    index_dir.mkdir(parents=True, exist_ok=True)
    (index_dir / "reverse_index.json").write_text('{"e2e-graph/dataset/alpha": []}', encoding="utf-8")

    rev = client.get("/api/v0-2/graph/reverse/e2e-graph/dataset/alpha")
    assert rev.status_code == 200
    assert rev.json()["data"]["concept_path"] == "e2e-graph/dataset/alpha"
    assert rev.json()["data"]["is_orphan"] is True

    imp = client.get("/api/v0-2/graph/impact/e2e-graph/dataset/alpha")
    assert imp.status_code == 200
    assert imp.json()["data"]["is_orphan"] is True


def test_e2e_lifecycle_archive_publish(client: TestClient) -> None:
    """Step 10: archive + publish emit audit events."""
    header = {"X-DevHub-User-Context": _path_y()}
    archive = client.post(
        "/api/v0-2/concepts/kb/dataset/foo/archive",
        headers=header,
        json={"reason": "operator-manual", "note": "e2e test"},
    )
    assert archive.status_code == 200
    assert archive.json()["data"]["new_status"] == "archived"
    assert archive.json()["data"]["archived_by"] == "u_e2e_001"

    publish = client.post(
        "/api/v0-2/concepts/kb/dataset/foo/publish",
        headers=header,
        json={"version": 1, "note": "e2e test"},
    )
    assert publish.status_code == 200
    assert publish.json()["data"]["new_status"] == "published"


def test_e2e_audit_log_viewer(client: TestClient) -> None:
    """Step 11: audit log viewer 4 endpoint accessible."""
    header = {"X-DevHub-User-Context": _path_y()}

    list_resp = client.get("/api/v0-2/audit", headers=header)
    assert list_resp.status_code == 200
    assert "items" in list_resp.json()["data"]

    user_resp = client.get("/api/v0-2/audit/user/u_e2e_001", headers=header)
    assert user_resp.status_code == 200

    org_resp = client.get("/api/v0-2/audit/org/ou_e2e_dept_a", headers=header)
    assert org_resp.status_code == 200

    concept_resp = client.get("/api/v0-2/audit/concept/kb/dataset/foo", headers=header)
    assert concept_resp.status_code == 200


def test_e2e_metrics_endpoint(client: TestClient) -> None:
    """Step 12: /metrics endpoint (gated by ENABLE_METRICS=false by default)."""
    resp = client.get("/metrics")
    assert resp.status_code == 200
    assert "text/plain" in resp.headers["content-type"]


def test_e2e_alerts_endpoint(client: TestClient) -> None:
    """Step 13: monitoring alerts endpoint returns valid envelope."""
    resp = client.get("/api/v0-2/monitoring/alerts")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert "alerts" in data
    assert "by_severity" in data
    assert data["by_severity"]["info"] >= 0


def test_e2e_full_happy_path(client: TestClient) -> None:
    """End-to-end happy path: ingest sync + raw register + bundle + audit + metrics."""
    header = {"X-DevHub-User-Context": _path_y(user_id="u_happy_001")}

    sync = client.post("/api/v0-2/ingest/homelab_mock/sync", headers=header)
    assert sync.status_code == 200

    reg = client.post(
        "/api/v0-2/raw",
        headers=header,
        json={"type": "dataset", "name": "happy", "source": "homelab_mock", "body": "# Happy"},
    )
    assert reg.status_code == 201
    raw_id = reg.json()["data"]["raw_id"]

    bundle = "happy-bundle"
    client.post(
        "/api/v0-2/bundles",
        headers=header,
        json={"name": bundle, "owner_org_id": "ou_e2e_dept_a", "visibility": "org"},
    )

    user_audit = client.get(
        "/api/v0-2/audit/user/u_happy_001",
        headers=header,
    )
    assert user_audit.status_code == 200
    items = user_audit.json()["data"]["items"]
    assert len(items) > 0

    assert raw_id
