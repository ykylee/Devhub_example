"""Curate API test (FR-C-001 ~ FR-C-005, 5 endpoint, umbrella doc §2 FR §3.1 정합)."""

from __future__ import annotations

import pytest
from fastapi.testclient import TestClient

from backend_knowledge.main import app


@pytest.fixture
def client(temp_var_dir) -> TestClient:
    """FastAPI test client with isolated var_dir."""
    return TestClient(app)


def _path_y_with_org(org_id: str = "ou_test_dept_a", roles: list[str] | None = None) -> str:
    """Build X-DevHub-User-Context header with given org + roles."""
    import base64
    import json
    from datetime import datetime, timezone
    payload = {
        "version": "v0",
        "user_id": "u_test_001",
        "org_id": org_id,
        "org_unit_ids": [org_id],
        "project_ids": ["prj_test_x"],
        "roles": roles or ["developer", "project_leader:prj_test_x"],
        "request_id": "req_test_20260619_curate",
        "issued_at": datetime.now(timezone.utc).isoformat(),
    }
    return base64.urlsafe_b64encode(
        json.dumps(payload).encode("utf-8")
    ).decode("ascii").rstrip("=")


class TestEnrich:
    """FR-C-001: POST /concepts/{id}/enrich."""

    def test_enrich_dry_run_returns_preview(self, client: TestClient) -> None:
        """Enrich with dry_run=true returns preview payload."""
        resp = client.post(
            "/api/v0-2/concepts/devhub-gitea/dataset/test-ds/enrich",
            json={"raw_id": "test-raw-001", "enricher": "rule-based", "dry_run": True},
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["concept_id"] == "devhub-gitea/dataset/test-ds"
        assert data["version"] == 1
        assert data["enricher_used"] == "rule-based"
        assert data["preview"] is not None
        assert "MOCK" in data["preview"]["note"]

    def test_enrich_rule_based_no_preview(self, client: TestClient) -> None:
        """Enrich with dry_run=false returns no preview."""
        resp = client.post(
            "/api/v0-2/concepts/devhub-gitea/dataset/test-ds/enrich",
            json={"raw_id": "test-raw-001", "enricher": "rule-based", "dry_run": False},
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["preview"] is None
        assert data["cross_links_extracted"] == 0

    def test_enrich_pi_llm_logged_as_mock(self, client: TestClient) -> None:
        """Enrich with pi-llm enricher is logged as MOCK (PR 2 placeholder)."""
        resp = client.post(
            "/api/v0-2/concepts/devhub-gitea/dataset/test-ds/enrich",
            json={"raw_id": "test-raw-001", "enricher": "pi-llm"},
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["enricher_used"] == "pi-llm"

    def test_enrich_invalid_enricher_rejected(self, client: TestClient) -> None:
        """Enrich with invalid enricher value returns 422."""
        resp = client.post(
            "/api/v0-2/concepts/x/y/enrich",
            json={"raw_id": "r1", "enricher": "gpt-4"},  # not in Literal
        )
        assert resp.status_code == 422


class TestConceptEdit:
    """FR-C-002: PUT /concepts/{id} (manual edit, Path Y 필수)."""

    def test_edit_without_path_y_fails(self, client: TestClient) -> None:
        """Manual edit without Path Y returns 400 E_VALIDATION."""
        resp = client.put(
            "/api/v0-2/concepts/devhub-gitea/dataset/test-ds",
            json={"body": "new body", "commit_message": "fix typo"},
        )
        assert resp.status_code == 400
        assert resp.json()["detail"]["code"] == "E_VALIDATION"

    def test_edit_with_path_y_succeeds(self, client: TestClient, path_y_header_value: str) -> None:
        """Manual edit with Path Y returns 200 with version + edited_by."""
        resp = client.put(
            "/api/v0-2/concepts/devhub-gitea/dataset/test-ds",
            headers={"X-DevHub-User-Context": path_y_header_value},
            json={"body": "# Updated\n\nNew body content", "commit_message": "fix typo"},
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["concept_id"] == "devhub-gitea/dataset/test-ds"
        assert data["version"] == 2
        assert data["edited_by"] == "u_test_001"
        assert "edited_at" in data

    def test_edit_body_xor_append_body_rejected(
        self, client: TestClient, path_y_header_value: str
    ) -> None:
        """Setting both body and append_body returns 400 (mutually exclusive)."""
        resp = client.put(
            "/api/v0-2/concepts/x/y",
            headers={"X-DevHub-User-Context": path_y_header_value},
            json={
                "body": "new",
                "append_body": "more",
                "commit_message": "x",
            },
        )
        assert resp.status_code == 400
        assert resp.json()["detail"]["code"] == "E_VALIDATION"

    def test_edit_missing_commit_message_rejected(
        self, client: TestClient, path_y_header_value: str
    ) -> None:
        """Edit without commit_message returns 422."""
        resp = client.put(
            "/api/v0-2/concepts/x/y",
            headers={"X-DevHub-User-Context": path_y_header_value},
            json={"body": "x"},
        )
        assert resp.status_code == 422

    def test_edit_cross_links_add_removed_counted(
        self, client: TestClient, path_y_header_value: str
    ) -> None:
        """cross_links_add / cross_links_remove counts are reflected in response."""
        resp = client.put(
            "/api/v0-2/concepts/x/y",
            headers={"X-DevHub-User-Context": path_y_header_value},
            json={
                "append_body": "extra",
                "commit_message": "add refs",
                "cross_links_add": [{"target": "a/b/c", "type": "explicit"}],
                "cross_links_remove": ["a/b/d"],
            },
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["cross_links_added"] == 1
        assert data["cross_links_removed"] == 1


class TestBundleList:
    """FR-C-003: GET /bundles."""

    def test_list_empty(self, client: TestClient) -> None:
        """Empty bundle list returns 0 items."""
        resp = client.get("/api/v0-2/bundles")
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["total"] == 0
        assert data["items"] == []

    def test_list_after_create(self, client: TestClient, path_y_header_value: str) -> None:
        """After creating a bundle, GET /bundles returns it."""
        # create
        create = client.post(
            "/api/v0-2/bundles",
            headers={"X-DevHub-User-Context": path_y_header_value},
            json={
                "name": "test-bundle-1",
                "description": "Test",
                "owner_org_id": "ou_test_dept_a",
                "visibility": "org",
            },
        )
        assert create.status_code == 201
        # list
        resp = client.get("/api/v0-2/bundles")
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["total"] == 1
        assert data["items"][0]["name"] == "test-bundle-1"
        assert data["items"][0]["visibility"] == "org"
        assert data["items"][0]["concept_count"] == 0


class TestBundleCreate:
    """FR-C-004: POST /bundles."""

    def test_create_without_path_y_fails(self, client: TestClient) -> None:
        """Bundle create without Path Y returns 400."""
        resp = client.post(
            "/api/v0-2/bundles",
            json={"name": "x", "owner_org_id": "ou_test_dept_a"},
        )
        assert resp.status_code == 400

    def test_create_duplicate_returns_409(
        self, client: TestClient, path_y_header_value: str
    ) -> None:
        """Duplicate bundle name returns 409 E_CONFLICT."""
        body = {
            "name": "dup-bundle",
            "owner_org_id": "ou_test_dept_a",
            "visibility": "org",
        }
        headers = {"X-DevHub-User-Context": path_y_header_value}
        first = client.post("/api/v0-2/bundles", headers=headers, json=body)
        assert first.status_code == 201
        second = client.post("/api/v0-2/bundles", headers=headers, json=body)
        assert second.status_code == 409
        assert second.json()["detail"]["code"] == "E_CONFLICT"

    def test_create_org_mismatch_returns_403(self, client: TestClient) -> None:
        """caller.org_id != owner_org_id returns 403 (not system_admin)."""
        resp = client.post(
            "/api/v0-2/bundles",
            headers={"X-DevHub-User-Context": _path_y_with_org("ou_test_dept_b")},
            json={"name": "x", "owner_org_id": "ou_test_dept_a", "visibility": "org"},
        )
        assert resp.status_code == 403
        assert resp.json()["detail"]["code"] == "E_FORBIDDEN"

    def test_create_system_admin_bypass(self, client: TestClient) -> None:
        """system_admin role can create bundle for any org."""
        resp = client.post(
            "/api/v0-2/bundles",
            headers={"X-DevHub-User-Context": _path_y_with_org(roles=["system_admin"])},
            json={"name": "admin-bundle", "owner_org_id": "ou_other", "visibility": "public"},
        )
        assert resp.status_code == 201
        data = resp.json()["data"]
        assert data["created_by"] == "u_test_001"
        assert data["visibility"] == "public"

    def test_create_invalid_name_pattern_rejected(
        self, client: TestClient, path_y_header_value: str
    ) -> None:
        """Invalid bundle name (uppercase) returns 422 (pattern violation)."""
        resp = client.post(
            "/api/v0-2/bundles",
            headers={"X-DevHub-User-Context": path_y_header_value},
            json={"name": "InvalidName", "owner_org_id": "ou_test_dept_a", "visibility": "org"},
        )
        assert resp.status_code == 422


class TestBundleRebuild:
    """FR-C-005: POST /bundles/{bundle}/rebuild."""

    def _make_bundle_with_concept(
        self, client: TestClient, path_y_header_value: str, bundle: str, type_: str, slug: str, body_md: str
    ) -> None:
        """Helper: create bundle + write concept .md + meta sidecar."""
        client.post(
            "/api/v0-2/bundles",
            headers={"X-DevHub-User-Context": path_y_header_value},
            json={"name": bundle, "owner_org_id": "ou_test_dept_a", "visibility": "org"},
        )
        from backend_knowledge.config import get_settings
        settings = get_settings()
        md_dir = settings.var_dir / "bundles" / bundle / type_
        md_dir.mkdir(parents=True, exist_ok=True)
        if not body_md.startswith("---"):
            frontmatter_str = (
                f"---\n"
                f"type: {type_}\n"
                f"x_devhub_name: {slug}\n"
                f"x_devhub_visibility: org\n"
                f"x_devhub_version: 1\n"
                f"title: {slug}\n"
                f"---\n\n"
            )
            body_md = frontmatter_str + body_md
        else:
            parts = body_md.split("---", 2)
            if len(parts) >= 3:
                existing_yaml = parts[1]
                if "type:" not in existing_yaml:
                    injected = f"type: {type_}\n" + existing_yaml
                    body_md = f"---{injected}---{parts[2]}"
        (md_dir / f"{slug}.md").write_text(body_md, encoding="utf-8")
        import json
        meta = {
            "bundle": bundle,
            "type": type_,
            "name": slug,
            "sha256": "fake_sha_" + slug,
            "source": "homelab_mock",
            "raw_id": None,
            "registered_by": "u_test_001",
            "visibility": "org",
            "frontmatter": {"title": slug, "description": f"Test {slug}"},
            "registered_at": "2026-06-19T00:00:00+00:00",
        }
        (md_dir / f"{slug}.meta.json").write_text(json.dumps(meta), encoding="utf-8")

    def test_rebuild_nonexistent_bundle_returns_404(
        self, client: TestClient
    ) -> None:
        """Rebuild on non-existent bundle returns 404."""
        resp = client.post("/api/v0-2/bundles/nonexistent-bundle/rebuild")
        assert resp.status_code == 404

    def test_rebuild_dry_run_no_artifacts(
        self, client: TestClient, path_y_header_value: str
    ) -> None:
        """dry_run=true rebuild returns stats but does not write artifacts."""
        self._make_bundle_with_concept(
            client, path_y_header_value,
            bundle="dryrun-bundle", type_="dataset", slug="alpha",
            body_md="---\ntitle: Alpha\ndescription: Test\n---\n\n# Alpha\n\nNo links.\n",
        )
        resp = client.post(
            "/api/v0-2/bundles/dryrun-bundle/rebuild",
            json={"dry_run": True},
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["concept_count"] == 1
        assert data["link_count"] == 0
        assert data["reverse_index_generated"] is False
        assert data["index_md_generated"] is False
        assert data["viz_html_generated"] is False
        # artifacts not created
        from backend_knowledge.config import get_settings
        settings = get_settings()
        index_dir = settings.var_dir / "bundles" / "dryrun-bundle" / ".index"
        assert not (index_dir / "index.md").exists()
        assert not (index_dir / "viz.html").exists()
        assert not (index_dir / "reverse_index.json").exists()

    def test_rebuild_full_generates_artifacts(
        self, client: TestClient, path_y_header_value: str
    ) -> None:
        """Full rebuild creates reverse_index.json + index.md + viz.html."""
        self._make_bundle_with_concept(
            client, path_y_header_value,
            bundle="full-bundle", type_="dataset", slug="alpha",
            body_md="---\ntitle: Alpha\ndescription: First\n---\n\n# Alpha\n\nSee [Beta](beta).\n",
        )
        self._make_bundle_with_concept(
            client, path_y_header_value,
            bundle="full-bundle", type_="dataset", slug="beta",
            body_md="---\ntitle: Beta\ndescription: Second\n---\n\n# Beta\n\nNo links.\n",
        )
        resp = client.post(
            "/api/v0-2/bundles/full-bundle/rebuild",
            json={"dry_run": False},
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["concept_count"] == 2
        assert data["link_count"] == 1  # alpha → beta
        assert data["reverse_index_generated"] is True
        assert data["index_md_generated"] is True
        assert data["viz_html_generated"] is True
        assert data["duration_ms"] >= 0

        from backend_knowledge.config import get_settings
        settings = get_settings()
        index_dir = settings.var_dir / "bundles" / "full-bundle" / ".index"
        assert (index_dir / "index.md").exists()
        assert (index_dir / "viz.html").exists()
        assert (index_dir / "reverse_index.json").exists()
        # index.md has both concepts
        index_md = (index_dir / "index.md").read_text(encoding="utf-8")
        assert "Alpha" in index_md
        assert "Beta" in index_md
        # reverse index has 1 entry for beta
        import json
        ri = json.loads((index_dir / "reverse_index.json").read_text(encoding="utf-8"))
        assert "full-bundle/dataset/beta" in ri
        assert len(ri["full-bundle/dataset/beta"]) == 1
        # viz.html has Cytoscape CDN
        viz = (index_dir / "viz.html").read_text(encoding="utf-8")
        assert "cytoscape" in viz.lower()
        assert "Alpha" in viz
        assert "Beta" in viz
