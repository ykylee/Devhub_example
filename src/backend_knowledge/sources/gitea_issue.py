"""Gitea issue source (umbrella doc §3.8).

Gitea API: GET /api/v1/repos/{owner}/{repo}/issues
Mock mode: GITEA_URL or GITEA_TOKEN missing 시 자동 fallback.
"""

from __future__ import annotations

from ._base import ConceptDict
from ._gitea_base import GiteaBaseSource
from .registry import register_source


@register_source
class GiteaIssueSource(GiteaBaseSource):
    """Gitea issue source (4 Gitea sub-plugin 중 issue)."""

    name: str = "gitea_issue"
    api_path: str = "/api/v1/repos/{owner}/{repo}/issues"

    async def normalize(self, raw: dict) -> ConceptDict:
        """Issue raw → concept (event or runbook)."""
        number = raw.get("number", 0)
        title = raw.get("title", "")
        body = raw.get("body", "")
        state = raw.get("state", "open")
        user = raw.get("user", {}).get("login", "")
        html_url = raw.get("html_url", "")
        labels = [label.get("name", "") for label in raw.get("labels", [])]

        md = f"# {title}\n\n"
        if body:
            md += body + "\n\n"
        md += f"**Issue #{number}** ({state}) by @{user}\n\n"
        md += f"- URL: {html_url}\n"
        if labels:
            md += f"- Labels: {', '.join(labels)}\n"

        return ConceptDict(
            source=self.name,
            type="event",  # issue = event
            name=f"issue-{number}",
            title=title,
            body=md,
            frontmatter={
                "tags": ["gitea", "issue", state] + labels,
                "description": f"Gitea Issue #{number}: {title}",
            },
            raw_refs=[],
            timestamp=raw.get("updated_at", ""),
            bundle="devhub-gitea",
        )
