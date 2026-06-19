"""Gitea wiki source (umbrella doc §3.8).

Gitea API: GET /api/v1/repos/{owner}/{repo}/wiki/pages
Mock mode: GITEA_URL or GITEA_TOKEN missing 시 자동 fallback.
"""

from __future__ import annotations

from ._base import ConceptDict
from ._gitea_base import GiteaBaseSource
from .registry import register_source


@register_source
class GiteaWikiSource(GiteaBaseSource):
    """Gitea wiki page source (4 Gitea sub-plugin 중 wiki)."""

    name: str = "gitea_wiki"
    api_path: str = "/api/v1/repos/{owner}/{repo}/wiki/pages"

    async def normalize(self, raw: dict) -> ConceptDict:
        """Wiki page raw → concept (reference)."""
        title = raw.get("title", "")
        content = raw.get("content", "")
        last_commit = raw.get("last_commit", {})
        author = last_commit.get("author", {}).get("login", "")
        created = last_commit.get("created", "")
        html_url = raw.get("html_url", "")

        # content 가 이미 markdown
        body = content if content else f"# {title}\n\n(Mock wiki page)"

        return ConceptDict(
            source=self.name,
            type="reference",  # wiki = reference
            name=title.lower().replace(" ", "-") if title else "wiki-page",
            title=title,
            body=body,
            frontmatter={
                "tags": ["gitea", "wiki"],
                "description": f"Gitea Wiki: {title}",
            },
            raw_refs=[],
            timestamp=created,
            bundle="devhub-gitea",
        )
