"""Graph API test (FR-G-001 ~ FR-G-004, 4 endpoint, umbrella doc §2 FR §3.1 정합)."""

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
) -> None:
    """Helper: write a concept .md + meta sidecar directly to disk."""
    title = title or slug
    md_dir = var_dir / "bundles" / bundle / type_
    md_dir.mkdir(parents=True, exist_ok=True)
    md_text = (
        f"---\n"
        f"type: {type_}\n"
        f"x_devhub_name: {slug}\n"
        f"x_devhub_visibility: org\n"
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
        "registered_by": "u_test_001",
        "visibility": "org",
        "frontmatter": {"title": title, "description": description},
        "registered_at": "2026-06-19T00:00:00+00:00",
    }
    (md_dir / f"{slug}.meta.json").write_text(json.dumps(meta), encoding="utf-8")


def _write_reverse_index(var_dir, bundle: str, reverse_index: dict[str, list[dict]]) -> None:
    """Helper: write reverse_index.json directly to disk."""
    index_dir = var_dir / "bundles" / bundle / ".index"
    index_dir.mkdir(parents=True, exist_ok=True)
    (index_dir / "reverse_index.json").write_text(
        json.dumps(reverse_index, ensure_ascii=False), encoding="utf-8"
    )


class TestReverseLinks:
    """FR-G-001: GET /graph/reverse/{concept_path}."""

    def test_reverse_short_path_returns_400(self, client: TestClient) -> None:
        """concept_path with < 3 segments returns 400."""
        resp = client.get("/api/v0-2/graph/reverse/only-bundle-type")
        assert resp.status_code == 400
        assert resp.json()["detail"]["code"] == "E_VALIDATION"

    def test_reverse_no_inlinks_is_orphan(
        self, client: TestClient, temp_var_dir
    ) -> None:
        """Concept with no reverse index entry is reported as orphan."""
        _write_reverse_index(temp_var_dir, "rb", {})
        resp = client.get("/api/v0-2/graph/reverse/rb/dataset/alpha")
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["concept_path"] == "rb/dataset/alpha"
        assert data["count"] == 0
        assert data["is_orphan"] is True
        assert data["inlinks"] == []

    def test_reverse_with_inlinks(
        self, client: TestClient, temp_var_dir
    ) -> None:
        """Concept with reverse index entry returns inlinks list."""
        reverse_index = {
            "rb/dataset/alpha": [
                {
                    "source_concept": "rb/dataset/beta",
                    "type": "explicit",
                    "section": None,
                    "context": "see [alpha](alpha.md)",
                },
                {
                    "source_concept": "rb/dataset/gamma",
                    "type": "wikilink",
                    "section": None,
                    "context": "[[alpha]]",
                },
            ]
        }
        _write_reverse_index(temp_var_dir, "rb", reverse_index)
        resp = client.get("/api/v0-2/graph/reverse/rb/dataset/alpha")
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["count"] == 2
        assert data["is_orphan"] is False
        sources = {item["source_concept"] for item in data["inlinks"]}
        assert "rb/dataset/beta" in sources
        assert "rb/dataset/gamma" in sources
        # All inlinks have created_by = "rule-based" (PR 2 MOCK)
        assert all(item["created_by"] == "rule-based" for item in data["inlinks"])


class TestImpact:
    """FR-G-002: GET /graph/impact/{concept_path}."""

    def test_impact_short_path_returns_400(self, client: TestClient) -> None:
        """concept_path with < 3 segments returns 400."""
        resp = client.get("/api/v0-2/graph/impact/x/y")
        assert resp.status_code == 400

    def test_impact_orphan_no_outlinks(
        self, client: TestClient, temp_var_dir
    ) -> None:
        """Orphan concept has 0 inlinks, 0 outlinks, rank_score=0."""
        _write_reverse_index(temp_var_dir, "ib", {})
        resp = client.get("/api/v0-2/graph/impact/ib/dataset/orphan")
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["inlink_count"] == 0
        assert data["outlink_count"] == 0
        assert data["is_orphan"] is True
        assert data["rank_score"] == 0.0

    def test_impact_with_inlinks_outlinks(
        self, client: TestClient, temp_var_dir
    ) -> None:
        """Concept with inlinks + outlinks returns both lists."""
        # Source: beta (inlink to alpha)
        # Target: gamma (outlink from alpha)
        _write_concept_with_meta(
            temp_var_dir, "ib", "dataset", "alpha",
            title="Alpha", description="central",
            body="See [Gamma](gamma).\n",
        )
        _write_concept_with_meta(
            temp_var_dir, "ib", "dataset", "beta",
            title="Beta", description="depends on alpha",
            body="See [Alpha](alpha).\n",
        )
        _write_concept_with_meta(
            temp_var_dir, "ib", "dataset", "gamma",
            title="Gamma", description="target",
        )
        reverse_index = {
            "ib/dataset/alpha": [
                {
                    "source_concept": "ib/dataset/beta",
                    "type": "explicit",
                    "section": None,
                    "context": "see [alpha](alpha.md)",
                }
            ]
        }
        _write_reverse_index(temp_var_dir, "ib", reverse_index)

        resp = client.get("/api/v0-2/graph/impact/ib/dataset/alpha")
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["inlink_count"] == 1
        assert data["outlink_count"] == 1
        assert data["is_orphan"] is False
        assert data["rank_score"] == 1.0  # 1 inlink / max(1,1) = 1.0
        assert data["inlinks"][0]["source_concept"] == "ib/dataset/beta"
        assert data["outlinks"][0]["target_concept"] == "ib/dataset/gamma"
        assert data["outlinks"][0]["resolved"] is True


class TestReindex:
    """FR-G-003: POST /graph/reindex."""

    def test_reindex_no_bundles(
        self, client: TestClient
    ) -> None:
        """Reindex with no bundles directory returns status=failed + 0 stats."""
        resp = client.post("/api/v0-2/graph/reindex", json={"full_scan": True})
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["status"] == "failed"
        assert data["stats"]["bundles_scanned"] == 0
        assert data["errors"][0]["code"] == "E_NOT_FOUND"

    def test_reindex_with_concepts(
        self, client: TestClient, temp_var_dir
    ) -> None:
        """Reindex scans all bundles' concepts and writes reverse_index.json."""
        _write_concept_with_meta(
            temp_var_dir, "rb-1", "dataset", "alpha",
            title="Alpha", body="[Beta](beta)\n",
        )
        _write_concept_with_meta(
            temp_var_dir, "rb-1", "dataset", "beta",
            title="Beta", body="No links.\n",
        )
        _write_concept_with_meta(
            temp_var_dir, "rb-2", "dataset", "solo",
            title="Solo", body="No links.\n",
        )
        resp = client.post(
            "/api/v0-2/graph/reindex",
            json={"full_scan": True},
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["status"] == "completed"
        assert data["stats"]["bundles_scanned"] == 2
        assert data["stats"]["concepts_scanned"] == 3
        assert data["stats"]["links_extracted"] == 1
        assert data["stats"]["orphans_detected"] == 2  # beta + solo
        # reverse_index.json written to each bundle
        assert (temp_var_dir / "bundles" / "rb-1" / ".index" / "reverse_index.json").exists()
        assert (temp_var_dir / "bundles" / "rb-2" / ".index" / "reverse_index.json").exists()
        ri = json.loads(
            (temp_var_dir / "bundles" / "rb-1" / ".index" / "reverse_index.json").read_text(
                encoding="utf-8"
            )
        )
        assert "rb-1/dataset/beta" in ri
        assert ri["rb-1/dataset/beta"][0]["source_concept"] == "rb-1/dataset/alpha"

    def test_reindex_bundle_filter(
        self, client: TestClient, temp_var_dir
    ) -> None:
        """Reindex with bundle filter only scans that bundle."""
        _write_concept_with_meta(
            temp_var_dir, "fb-a", "dataset", "x", title="X",
        )
        _write_concept_with_meta(
            temp_var_dir, "fb-b", "dataset", "y", title="Y",
        )
        resp = client.post(
            "/api/v0-2/graph/reindex",
            json={"full_scan": True, "bundle": "fb-a"},
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["stats"]["bundles_scanned"] == 1
        assert data["stats"]["concepts_scanned"] == 1
        # fb-b NOT scanned → no reverse_index.json
        assert not (
            temp_var_dir / "bundles" / "fb-b" / ".index" / "reverse_index.json"
        ).exists()


class TestResolveLinks:
    """FR-G-004: POST /concepts/{id}/resolve-links (Pi LLM MOCK)."""

    def test_resolve_without_path_y_fails(self, client: TestClient) -> None:
        """resolve-links without Path Y returns 400."""
        resp = client.post(
            "/api/v0-2/concepts/x/y/resolve-links",
            json={"mode": "dry-run"},
        )
        assert resp.status_code == 400

    def test_resolve_dry_run_returns_1_mock_candidate(
        self, client: TestClient, path_y_header_value: str
    ) -> None:
        """dry-run mode returns 1 MOCK candidate with confidence 0.5."""
        resp = client.post(
            "/api/v0-2/concepts/kb/dataset/alpha/resolve-links",
            headers={"X-DevHub-User-Context": path_y_header_value},
            json={"mode": "dry-run"},
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["concept_id"] == "kb/dataset/alpha"
        assert data["mode"] == "dry-run"
        assert data["applied"] is False
        assert data["applied_at"] is None
        assert len(data["candidates"]) == 1
        assert data["candidates"][0]["rank"] == 1
        assert data["candidates"][0]["target_concept"] == "MOCK_TARGET"
        assert data["candidates"][0]["confidence"] == 0.5
        assert "MOCK" in data["candidates"][0]["reasoning"]

    def test_resolve_auto_apply_sets_applied_true(
        self, client: TestClient, path_y_header_value: str
    ) -> None:
        """auto-apply mode returns applied=True + applied_at timestamp."""
        resp = client.post(
            "/api/v0-2/concepts/kb/dataset/alpha/resolve-links",
            headers={"X-DevHub-User-Context": path_y_header_value},
            json={"mode": "auto-apply", "selected_rank": 1, "confidence_threshold": 0.9},
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["mode"] == "auto-apply"
        assert data["applied"] is True
        assert data["applied_at"] is not None

    def test_resolve_confidence_threshold_out_of_range_rejected(
        self, client: TestClient, path_y_header_value: str
    ) -> None:
        """confidence_threshold > 1.0 returns 422 (Pydantic le=1.0 violation)."""
        resp = client.post(
            "/api/v0-2/concepts/kb/dataset/alpha/resolve-links",
            headers={"X-DevHub-User-Context": path_y_header_value},
            json={"mode": "dry-run", "confidence_threshold": 1.5},
        )
        assert resp.status_code == 422
