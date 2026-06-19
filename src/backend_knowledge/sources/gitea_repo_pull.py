"""Gitea repo pull source (umbrella doc §3.8).

Gitea API: GET /api/v1/repos/{owner}/{repo}/pulls
Mock mode: GITEA_URL or GITEA_TOKEN missing 시 자동 fallback.
"""

from __future__ import annotations

from ._base import ConceptDict
from ._gitea_base import GiteaBaseSource
from .registry import register_source


@register_source
class GiteaRepoPullSource(GiteaBaseSource):
    """Gitea pull request source (4 Gitea sub-plugin 중 repo_pull)."""

    name: str = "gitea_repo_pull"
    api_path: str = "/api/v1/repos/{owner}/{repo}/pulls"

    async def normalize(self, raw: dict) -> ConceptDict:
        """PR raw → concept (api_endpoint or integration)."""
        number = raw.get("number", 0)
        title = raw.get("title", "")
        body = raw.get("body", "")
        state = raw.get("state", "open")
        user = raw.get("user", {}).get("login", "")
        html_url = raw.get("html_url", "")

        # body markdown + cross-link + metadata
        md = f"# {title}\n\n"
        if body:
            md += body + "\n\n"
        md += f"**PR #{number}** ({state}) by @{user}\n\n"
        md += f"- URL: {html_url}\n"

        return ConceptDict(
            source=self.name,
            type="api_endpoint",  # PR = API endpoint for review/CI
            name=f"pr-{number}",
            title=title,
            body=md,
            frontmatter={
                "tags": ["gitea", "pull-request", state],
                "description": f"Gitea PR #{number}: {title}",
            },
            raw_refs=[],
            timestamp=raw.get("updated_at", ""),
            bundle="devhub-gitea",
        )
