"""Gitea Actions / CI source (umbrella doc §3.8).

Gitea API: GET /api/v1/repos/{owner}/{repo}/actions/runs
Mock mode: GITEA_URL or GITEA_TOKEN missing 시 자동 fallback.
"""

from __future__ import annotations

from ._base import ConceptDict
from ._gitea_base import GiteaBaseSource
from .registry import register_source


@register_source
class GiteaActionSource(GiteaBaseSource):
    """Gitea Actions / CI run source (4 Gitea sub-plugin 중 action)."""

    name: str = "gitea_action"
    api_path: str = "/api/v1/repos/{owner}/{repo}/actions/runs"

    async def normalize(self, raw: dict) -> ConceptDict:
        """Action run raw → concept (event or metric)."""
        run_id = raw.get("id", 0)
        name = raw.get("name", "")
        status = raw.get("status", "")
        conclusion = raw.get("conclusion", "")
        head_branch = raw.get("head_branch", "")
        event = raw.get("event", "")
        html_url = raw.get("html_url", "")

        md = f"# {name} (#{run_id})\n\n"
        md += f"**Status**: {status} ({conclusion})\n"
        md += f"**Branch**: {head_branch}\n"
        md += f"**Event**: {event}\n"
        md += f"- URL: {html_url}\n"

        return ConceptDict(
            source=self.name,
            type="event",  # CI run = event
            name=f"action-{run_id}",
            title=f"{name} (#{run_id})",
            body=md,
            frontmatter={
                "tags": ["gitea", "actions", "ci", status, conclusion],
                "description": f"Gitea Actions run #{run_id} for {name}",
            },
            raw_refs=[],
            timestamp=raw.get("updated_at", ""),
            bundle="devhub-gitea",
        )
