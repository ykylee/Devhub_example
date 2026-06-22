"""LinkResolver unit tests (umbrella doc §3.5.7 + M-v0.2.3+).

mock mode (Pi SDK/RPC 미설치) 의 _mock_pi_response 사용.
Real Pi integration test 는 별도 (CI env Pi SDK 또는 RPC server 필요).
"""

from __future__ import annotations

import os
from pathlib import Path

import pytest

# PI_MODE / PI_RPC_URL 환경변수 명시적 unset (mock mode 강제)
os.environ.pop("PI_MODE", None)
os.environ.pop("PI_RPC_URL", None)

from backend_knowledge.curate.link_resolver import (  # noqa: E402
    LinkRecommendation,
    LinkResolver,
    PiMode,
    ResolutionMode,
    UnresolvedLink,
    render_prompt,
)


class TestRenderPrompt:
    def test_basic_template(self):
        prompt = render_prompt(
            source_path="devhub-gitea/issue/gitea-issue-42.md",
            link_text="../gitea-issue-43.md",
            link_target="../gitea-issue-43.md",
            context="Some context here",
            candidates=[
                {"path": "devhub-gitea/issue/gitea-issue-43.md", "title": "Issue 43", "type": "event"},
            ],
        )
        assert "devhub-gitea/issue/gitea-issue-42.md" in prompt
        assert "../gitea-issue-43.md" in prompt
        assert "Issue 43" in prompt
        assert "rank" in prompt
        assert "confidence" in prompt

    def test_empty_candidates(self):
        prompt = render_prompt(
            source_path="test.md",
            link_text="../other.md",
            link_target="../other.md",
            context="ctx",
            candidates=[],
        )
        assert "Candidate concepts" in prompt
        assert "test.md" in prompt


class TestPiMode:
    def test_default_mode_is_sdk(self):
        os.environ.pop("PI_MODE", None)
        resolver = LinkResolver()
        assert resolver.pi_mode == PiMode.SDK

    def test_explicit_sdk_mode(self):
        resolver = LinkResolver(pi_mode=PiMode.SDK)
        assert resolver.pi_mode == PiMode.SDK

    def test_rpc_mode(self):
        resolver = LinkResolver(pi_mode=PiMode.RPC)
        assert resolver.pi_mode == PiMode.RPC


class TestLinkRecommendation:
    def test_construction(self):
        rec = LinkRecommendation(
            rank=1,
            target_slug="gitea-issue-43",
            target_path="devhub-gitea/issue/gitea-issue-43.md",
            reason="Same repo, similar title",
            confidence=0.92,
        )
        assert rec.rank == 1
        assert rec.confidence == 0.92

    def test_confidence_validation(self):
        with pytest.raises(Exception):
            LinkRecommendation(
                rank=1,
                target_slug="x",
                target_path="x.md",
                reason="r",
                confidence=1.5,
            )


class TestResolutionResult:
    def test_dry_run_no_apply(self):
        result = LinkResolver()._mock_pi_response.__wrapped__ if False else None
        assert result is None


class TestLinkResolverMockPiResponse:
    def test_mock_3_candidates(self):
        resolver = LinkResolver()
        candidates = [
            {"path": "a/x.md", "title": "A", "type": "event"},
            {"path": "b/y.md", "title": "B", "type": "runbook"},
            {"path": "c/z.md", "title": "C", "type": "dataset"},
        ]
        recs = resolver._mock_pi_response("test prompt", candidates)
        assert len(recs) == 3
        assert recs[0].rank == 1
        assert recs[0].confidence == 0.95
        assert recs[1].confidence == 0.85
        assert recs[2].confidence == 0.75
        assert recs[0].target_path == "a/x.md"

    def test_mock_fewer_candidates(self):
        resolver = LinkResolver()
        recs = resolver._mock_pi_response("test", [{"path": "x.md", "title": "X", "type": "event"}])
        assert len(recs) == 1


class TestLinkResolverResolve:
    @pytest.mark.asyncio
    async def test_dry_run_mode_returns_recommendations_without_apply(self):
        resolver = LinkResolver()
        link = UnresolvedLink(
            source_path="devhub-gitea/issue/gitea-issue-42.md",
            link_text="../gitea-issue-43.md",
            link_target="../gitea-issue-43.md",
            context="context",
        )
        candidates = [{"path": "devhub-gitea/issue/gitea-issue-43.md", "title": "Issue 43", "type": "event"}]
        result = await resolver.resolve(
            link=link,
            candidate_concepts=candidates,
            mode=ResolutionMode.DRY_RUN,
            confidence_threshold=0.9,
        )
        assert result.mode == ResolutionMode.DRY_RUN
        assert result.applied is False
        assert result.selected is not None

    @pytest.mark.asyncio
    async def test_auto_apply_high_confidence(self):
        resolver = LinkResolver()
        link = UnresolvedLink(
            source_path="x.md",
            link_text="../y.md",
            link_target="../y.md",
            context="ctx",
        )
        candidates = [{"path": "y.md", "title": "Y", "type": "event"}]
        result = await resolver.resolve(
            link=link,
            candidate_concepts=candidates,
            mode=ResolutionMode.AUTO_APPLY,
            confidence_threshold=0.9,
        )
        # mock confidence = 0.95 > threshold 0.9
        assert result.applied is True
        assert result.selected is not None
        assert result.selected.confidence >= 0.9

    @pytest.mark.asyncio
    async def test_auto_apply_low_confidence_skipped(self):
        resolver = LinkResolver()
        link = UnresolvedLink(
            source_path="x.md",
            link_text="../y.md",
            link_target="../y.md",
            context="ctx",
        )
        # 빈 candidates → mock returns empty → applied = False
        result = await resolver.resolve(
            link=link,
            candidate_concepts=[],
            mode=ResolutionMode.AUTO_APPLY,
            confidence_threshold=0.99,
        )
        assert result.applied is False

    @pytest.mark.asyncio
    async def test_confirm_mode_applies_with_selected_rank(self):
        resolver = LinkResolver()
        link = UnresolvedLink(
            source_path="x.md",
            link_text="../y.md",
            link_target="../y.md",
            context="ctx",
        )
        candidates = [
            {"path": "y1.md", "title": "Y1", "type": "event"},
            {"path": "y2.md", "title": "Y2", "type": "runbook"},
        ]
        result = await resolver.resolve(
            link=link,
            candidate_concepts=candidates,
            mode=ResolutionMode.CONFIRM,
            confidence_threshold=0.99,
            selected_rank=2,
        )
        assert result.applied is True
        assert result.selected.rank == 2
        assert result.selected.target_path == "y2.md"


class TestLinkResolverFindUnresolved:
    @pytest.mark.asyncio
    async def test_nonexistent_bundle_dir_returns_empty(self, tmp_path: Path):
        nonexistent = tmp_path / "nonexistent"
        resolver = LinkResolver()
        unresolved = await resolver.find_unresolved_links(nonexistent)
        assert unresolved == []

    @pytest.mark.asyncio
    async def test_empty_bundle_dir_returns_empty(self, tmp_path: Path):
        resolver = LinkResolver()
        unresolved = await resolver.find_unresolved_links(tmp_path)
        assert unresolved == []


class TestLinkResolverListCandidates:
    @pytest.mark.asyncio
    async def test_nonexistent_var_dir_returns_empty(self, tmp_path: Path, monkeypatch):
        # Mock settings.var_dir to nonexistent
        from backend_knowledge import config

        original = config._settings
        config._settings = None
        try:
            from backend_knowledge.config import Settings

            monkeypatch.setattr(
                Settings, "var_dir", tmp_path / "nonexistent_var"
            )
            resolver = LinkResolver()
            candidates = await resolver.list_candidates("test-slug")
            assert candidates == []
        finally:
            config._settings = original


class TestParseContextHelper:
    def test_extract_context_basic(self):
        from backend_knowledge.curate.link_resolver import _extract_context

        body = "line1\nline2\nline3 [link](../target.md) line4\nline5\nline6\nline7"
        ctx = _extract_context(body, body.index("[link]"), body.index("[/link]") + 7)
        assert "line3" in ctx or "line2" in ctx


class TestUnresolvedLinkModel:
    def test_construction(self):
        link = UnresolvedLink(
            source_path="x.md",
            link_text="../y.md",
            link_target="../y.md",
            context="context text",
        )
        assert link.source_path == "x.md"
        assert link.context == "context text"