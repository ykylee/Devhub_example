"""Query API test (FR-Q-001 ~ FR-Q-005, 5 endpoint, umbrella doc §2 FR §3.1 정합)."""

from __future__ import annotations

import json
import pytest
from fastapi.testclient import TestClient

from backend_knowledge.main import app


@pytest.fixture
def client(temp_var_dir) -> TestClient:
    """FastAPI test client with isolated var_dir."""
    return TestClient(app)


def _write_concept_with_meta(
    var_dir,
    bundle: str,
    type_: str,
    slug: str,
    title: str = "",
    description: str = "",
    body: str = "",
    visibility: str = "org",
    registered_by: str = "u_test_001",
) -> None:
    """Helper: write a concept .md + meta sidecar directly to disk."""
    title = title or slug
    md_dir = var_dir / "bundles" / bundle / type_
    md_dir.mkdir(parents=True, exist_ok=True)
    md_text = (
        f"---\n"
        f"type: {type_}\n"
        f"x_devhub_name: {slug}\n"
        f"x_devhub_visibility: {visibility}\n"
        f"x_devhub_version: 1\n"
        f"title: {title}\n"
        f"description: {description}\n"
        f"---\n\n"
        f"# {title}\n\n{body}\n"
    )
    (md_dir / f"{slug}.md").write_text(md_text, encoding="utf-8")
    meta = {
        "bundle": bundle,
        "type": type_,
        "name": slug,
        "sha256": "fake_sha_" + slug,
        "source": "homelab_mock",
        "raw_id": None,
        "registered_by": registered_by,
        "visibility": visibility,
        "frontmatter": {"title": title, "description": description},
        "registered_at": "2026-06-19T00:00:00+00:00",
    }
    (md_dir / f"{slug}.meta.json").write_text(json.dumps(meta), encoding="utf-8")


class TestPostQuery:
    """FR-Q-001: POST /query."""

    def test_query_without_path_y_fails(self, client: TestClient) -> None:
        """Query without Path Y returns 400."""
        resp = client.post("/api/v0-2/query", json={"query": "test"})
        assert resp.status_code == 400

    def test_query_empty_results(
        self, client: TestClient, path_y_header_value: str
    ) -> None:
        """Query against empty store returns 0 contexts."""
        resp = client.post(
            "/api/v0-2/query",
            headers={"X-DevHub-User-Context": path_y_header_value},
            json={"query": "kubernetes"},
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["contexts"] == []
        assert data["answer"] is None
        assert data["query_metadata"]["retrieval_method"] == "substring-mock"

    def test_query_substring_match(
        self, client: TestClient, path_y_header_value: str, temp_var_dir
    ) -> None:
        """Query with substring match returns matching concepts."""
        _write_concept_with_meta(
            temp_var_dir, "gitea-bundle", "dataset", "users",
            title="Users Dataset", description="All registered users",
        )
        _write_concept_with_meta(
            temp_var_dir, "gitea-bundle", "dataset", "orders",
            title="Orders Dataset", description="All customer orders",
        )
        resp = client.post(
            "/api/v0-2/query",
            headers={"X-DevHub-User-Context": path_y_header_value},
            json={"query": "users", "top_k": 10},
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert len(data["contexts"]) == 1
        assert data["contexts"][0]["title"] == "Users Dataset"
        assert data["contexts"][0]["type"] == "dataset"

    def test_query_with_bundle_filter(
        self, client: TestClient, path_y_header_value: str, temp_var_dir
    ) -> None:
        """Query with bundle filter only returns concepts from that bundle."""
        _write_concept_with_meta(
            temp_var_dir, "bundle-a", "dataset", "shared",
            title="Shared in A", description="A only",
        )
        _write_concept_with_meta(
            temp_var_dir, "bundle-b", "dataset", "shared",
            title="Shared in B", description="B only",
        )
        resp = client.post(
            "/api/v0-2/query",
            headers={"X-DevHub-User-Context": path_y_header_value},
            json={"query": "Shared", "bundle": "bundle-a"},
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert len(data["contexts"]) == 1
        assert data["contexts"][0]["title"] == "Shared in A"

    def test_query_top_k_limits_results(
        self, client: TestClient, path_y_header_value: str, temp_var_dir
    ) -> None:
        """Query with top_k=1 limits to 1 result even if 5 match."""
        for i in range(5):
            _write_concept_with_meta(
                temp_var_dir, "kb", "dataset", f"item-{i}",
                title=f"Common Item {i}", description="test",
            )
        resp = client.post(
            "/api/v0-2/query",
            headers={"X-DevHub-User-Context": path_y_header_value},
            json={"query": "Common", "top_k": 1},
        )
        assert resp.status_code == 200
        assert len(resp.json()["data"]["contexts"]) == 1


class TestGetConcept:
    """FR-Q-002: GET /concepts/{type}/{name}."""

    def test_concept_without_path_y_fails(
        self, client: TestClient, temp_var_dir
    ) -> None:
        """Concept get without Path Y returns 400."""
        _write_concept_with_meta(
            temp_var_dir, "kb", "dataset", "alpha",
        )
        resp = client.get("/api/v0-2/concepts/dataset/alpha")
        assert resp.status_code == 400

    def test_concept_not_found(
        self, client: TestClient, path_y_header_value: str
    ) -> None:
        """Concept get for non-existent slug returns 404."""
        resp = client.get(
            "/api/v0-2/concepts/dataset/nonexistent",
            headers={"X-DevHub-User-Context": path_y_header_value},
        )
        assert resp.status_code == 404

    def test_concept_get_succeeds(
        self, client: TestClient, path_y_header_value: str, temp_var_dir
    ) -> None:
        """Concept get returns frontmatter + body + version + bundle."""
        _write_concept_with_meta(
            temp_var_dir, "kb", "dataset", "alpha",
            title="Alpha", description="First test",
            body="This is the alpha body.",
        )
        resp = client.get(
            "/api/v0-2/concepts/dataset/alpha",
            params={"bundle": "kb"},
            headers={"X-DevHub-User-Context": path_y_header_value},
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["concept_id"] == "kb/dataset/alpha"
        assert data["type"] == "dataset"
        assert data["name"] == "alpha"
        assert data["bundle"] == "kb"
        assert "title" in data["frontmatter"]
        assert "alpha body" in data["body"]
        assert data["version"] == 1

    def test_concept_personal_visibility_other_user_403(
        self, client: TestClient, temp_var_dir
    ) -> None:
        """Personal concept visible only to registered_by user (other user gets 403)."""
        import base64
        from datetime import datetime, timezone
        # Other user
        other_payload = {
            "version": "v0",
            "user_id": "u_other",
            "org_id": "ou_test_dept_a",
            "org_unit_ids": ["ou_test_dept_a"],
            "project_ids": [],
            "roles": ["developer"],
            "request_id": "req_other",
            "issued_at": datetime.now(timezone.utc).isoformat(),
        }
        other_header = base64.urlsafe_b64encode(
            json.dumps(other_payload).encode("utf-8")
        ).decode("ascii").rstrip("=")

        _write_concept_with_meta(
            temp_var_dir, "kb", "dataset", "private",
            title="Private", visibility="personal", registered_by="u_test_001",
        )
        resp = client.get(
            "/api/v0-2/concepts/dataset/private",
            params={"bundle": "kb"},
            headers={"X-DevHub-User-Context": other_header},
        )
        assert resp.status_code == 403
        assert resp.json()["detail"]["code"] == "E_FORBIDDEN"


class TestSearch:
    """FR-Q-003: GET /search."""

    def test_search_without_path_y_fails(self, client: TestClient) -> None:
        """Search without Path Y returns 400."""
        resp = client.get("/api/v0-2/search?q=test")
        assert resp.status_code == 400

    def test_search_missing_q_returns_422(
        self, client: TestClient, path_y_header_value: str
    ) -> None:
        """Search without q param returns 422."""
        resp = client.get(
            "/api/v0-2/search",
            headers={"X-DevHub-User-Context": path_y_header_value},
        )
        assert resp.status_code == 422

    def test_search_substring_match(
        self, client: TestClient, path_y_header_value: str, temp_var_dir
    ) -> None:
        """Search by title substring returns matching concepts."""
        _write_concept_with_meta(
            temp_var_dir, "kb", "dataset", "users-2026",
            title="Users 2026", description="Annual snapshot",
        )
        _write_concept_with_meta(
            temp_var_dir, "kb", "dataset", "orders-2026",
            title="Orders 2026", description="Annual orders",
        )
        resp = client.get(
            "/api/v0-2/search",
            params={"q": "Users"},
            headers={"X-DevHub-User-Context": path_y_header_value},
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["total"] == 1
        assert data["hits"][0]["title"] == "Users 2026"

    def test_search_with_type_filter(
        self, client: TestClient, path_y_header_value: str, temp_var_dir
    ) -> None:
        """Search with type filter narrows results to that type only."""
        _write_concept_with_meta(
            temp_var_dir, "kb", "dataset", "users",
            title="Users", description="all users",
        )
        _write_concept_with_meta(
            temp_var_dir, "kb", "metric", "user-count",
            title="User Count", description="metric about users",
        )
        resp = client.get(
            "/api/v0-2/search",
            params={"q": "user", "type": ["dataset"]},
            headers={"X-DevHub-User-Context": path_y_header_value},
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["total"] == 1
        assert data["hits"][0]["type"] == "dataset"

    def test_search_pagination(
        self, client: TestClient, path_y_header_value: str, temp_var_dir
    ) -> None:
        """Search pagination: limit=2 returns 2 hits + next_offset."""
        for i in range(5):
            _write_concept_with_meta(
                temp_var_dir, "kb", "dataset", f"item-{i}",
                title=f"Common {i}", description="x",
            )
        resp = client.get(
            "/api/v0-2/search",
            params={"q": "Common", "limit": 2, "offset": 0},
            headers={"X-DevHub-User-Context": path_y_header_value},
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert len(data["hits"]) == 2
        assert data["total"] == 5
        assert data["next_offset"] == 2


class TestBundleIndex:
    """FR-Q-004: GET /bundles/{bundle}/index.md."""

    def test_index_md_not_found(
        self, client: TestClient
    ) -> None:
        """Bundle with no index.md returns 404."""
        resp = client.get("/api/v0-2/bundles/missing-bundle/index.md")
        assert resp.status_code == 404

    def test_index_md_after_rebuild(
        self, client: TestClient, path_y_header_value: str, temp_var_dir
    ) -> None:
        """After rebuild, GET index.md returns Markdown text."""
        _write_concept_with_meta(
            temp_var_dir, "idx-bundle", "dataset", "alpha",
            title="Alpha", description="First",
        )
        # rebuild first
        client.post(
            "/api/v0-2/bundles/idx-bundle/rebuild",
            json={"dry_run": False},
        )
        resp = client.get("/api/v0-2/bundles/idx-bundle/index.md")
        assert resp.status_code == 200
        assert "text/markdown" in resp.headers["content-type"]
        assert "Alpha" in resp.text
        assert "dataset/alpha.md" in resp.text


class TestBundleViz:
    """FR-Q-005: GET /bundles/{bundle}/viz.html."""

    def test_viz_html_not_found(self, client: TestClient) -> None:
        """Bundle with no viz.html returns 404."""
        resp = client.get("/api/v0-2/bundles/missing-bundle/viz.html")
        assert resp.status_code == 404

    def test_viz_html_after_rebuild(
        self, client: TestClient, path_y_header_value: str, temp_var_dir
    ) -> None:
        """After rebuild, GET viz.html returns self-contained HTML with Cytoscape."""
        _write_concept_with_meta(
            temp_var_dir, "viz-bundle", "dataset", "alpha",
            title="Alpha", description="First",
        )
        client.post(
            "/api/v0-2/bundles/viz-bundle/rebuild",
            json={"dry_run": False},
        )
        resp = client.get("/api/v0-2/bundles/viz-bundle/viz.html")
        assert resp.status_code == 200
        assert "text/html" in resp.headers["content-type"]
        body = resp.text
        assert "cytoscape" in body.lower()
        assert "Alpha" in body
        # Self-contained: no external CSS / JS files except Cytoscape CDN
        assert "<link" not in body  # no external CSS
