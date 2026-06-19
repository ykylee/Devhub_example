"""Source plugin test (umbrella doc §3.8 + ADR-0035)."""

from __future__ import annotations

from datetime import datetime, timezone

import pytest

from backend_knowledge.sources import (
    GiteaActionSource,
    GiteaIssueSource,
    GiteaRepoPullSource,
    GiteaWikiSource,
    HomelabMockSource,
    get_source,
    list_sources,
)


class TestSourceRegistry:
    """Test source plugin registry."""

    def test_5_sources_registered(self) -> None:
        """All 5 PoC source plugin should be registered."""
        sources = list_sources()
        assert len(sources) == 5
        assert "homelab_mock" in sources
        assert "gitea_repo_pull" in sources
        assert "gitea_issue" in sources
        assert "gitea_wiki" in sources
        assert "gitea_action" in sources

    def test_get_source_returns_instance(self) -> None:
        """get_source should return SourcePlugin instance."""
        for name in ("homelab_mock", "gitea_repo_pull", "gitea_issue", "gitea_wiki", "gitea_action"):
            plugin = get_source(name)
            assert plugin.name == name

    def test_unknown_source_raises(self) -> None:
        """get_source with unknown name should raise KeyError."""
        with pytest.raises(KeyError, match="unknown source"):
            get_source("nonexistent_source")


class TestHomelabMockSource:
    """Test homelab_mock source (in-memory, no external call)."""

    @pytest.mark.asyncio
    async def test_connect(self) -> None:
        """Connect should be no-op for mock."""
        source = HomelabMockSource()
        await source.connect({})
        health = await source.health_check()
        assert health["healthy"] is True

    @pytest.mark.asyncio
    async def test_fetch_full_sync(self) -> None:
        """Full sync (since=None) should return all 3 mock concepts."""
        source = HomelabMockSource()
        await source.connect({})
        raws = await source.fetch(since=None)
        assert len(raws) == 3
        names = {raw["name"] for raw in raws}
        assert "homelab-dataset-cpu-metrics" in names
        assert "homelab-metric-cpu-usage" in names
        assert "homelab-runbook-cpu-high" in names

    @pytest.mark.asyncio
    async def test_fetch_incremental(self) -> None:
        """Incremental sync (since=future) should return 0 concepts."""
        source = HomelabMockSource()
        await source.connect({})
        future = datetime.now(timezone.utc).replace(year=2030)
        raws = await source.fetch(since=future)
        assert raws == []

    @pytest.mark.asyncio
    async def test_normalize(self) -> None:
        """Normalize should produce ConceptDict with 8 fields."""
        source = HomelabMockSource()
        await source.connect({})
        raws = await source.fetch(since=None)
        for raw in raws:
            concept = await source.normalize(raw)
            assert concept["source"] == "homelab_mock"
            assert concept["type"] in ("dataset", "metric", "api_endpoint", "runbook", "integration", "event", "reference", "decision")
            assert concept["name"]
            assert concept["title"]
            assert concept["body"]
            assert "tags" in concept["frontmatter"]


class TestGiteaMockSources:
    """Test 4 Gitea sub-plugin in mock mode (GITEA_URL/T 미설정)."""

    @pytest.mark.asyncio
    async def test_all_4_gitea_sources_use_mock_mode(self) -> None:
        """All 4 Gitea sources should fall back to mock mode without env vars."""
        for cls in (GiteaRepoPullSource, GiteaIssueSource, GiteaWikiSource, GiteaActionSource):
            source = cls()
            await source.connect({})
            assert source.is_mock_mode is True
            raws = await source.fetch(since=None)
            assert len(raws) == 1  # 1 sample per source
            health = await source.health_check()
            assert health["healthy"] is True

    @pytest.mark.asyncio
    async def test_gitea_repo_pull_normalize(self) -> None:
        """Gitea repo pull should normalize as api_endpoint concept."""
        source = GiteaRepoPullSource()
        await source.connect({})
        raws = await source.fetch(since=None)
        concept = await source.normalize(raws[0])
        assert concept["source"] == "gitea_repo_pull"
        assert concept["type"] == "api_endpoint"
        assert concept["name"].startswith("pr-")
        assert "PR" in concept["body"]

    @pytest.mark.asyncio
    async def test_gitea_issue_normalize(self) -> None:
        """Gitea issue should normalize as event concept."""
        source = GiteaIssueSource()
        await source.connect({})
        raws = await source.fetch(since=None)
        concept = await source.normalize(raws[0])
        assert concept["source"] == "gitea_issue"
        assert concept["type"] == "event"
        assert concept["name"].startswith("issue-")

    @pytest.mark.asyncio
    async def test_gitea_wiki_normalize(self) -> None:
        """Gitea wiki should normalize as reference concept."""
        source = GiteaWikiSource()
        await source.connect({})
        raws = await source.fetch(since=None)
        concept = await source.normalize(raws[0])
        assert concept["source"] == "gitea_wiki"
        assert concept["type"] == "reference"

    @pytest.mark.asyncio
    async def test_gitea_action_normalize(self) -> None:
        """Gitea action should normalize as event concept."""
        source = GiteaActionSource()
        await source.connect({})
        raws = await source.fetch(since=None)
        concept = await source.normalize(raws[0])
        assert concept["source"] == "gitea_action"
        assert concept["type"] == "event"
        assert concept["name"].startswith("action-")
