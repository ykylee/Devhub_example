"""Source plugin ABC (umbrella doc §3.8.1 + ADR-0035 §3.2).

ConceptDict = raw fetch 결과 (raw JSON dict) 의 normalized form.
SourcePlugin = 외부 시스템 5종 의 ABC.
"""

from __future__ import annotations

from abc import ABC, abstractmethod
from datetime import datetime
from typing import Any, TypedDict


class ConceptDict(TypedDict, total=False):
    """Normalized concept dict (raw fetch 결과).

    - source: source plugin name (e.g., "gitea_issue")
    - type: 8종 enum (dataset / metric / api_endpoint / runbook / integration / event / reference / decision)
    - name: concept identifier (slug, e.g., "gitea-issue-42")
    - title: human-readable title
    - body: Markdown body (cross-link 포함 가능)
    - frontmatter: additional frontmatter (tags, description, etc)
    - raw_refs: list of raw_id this concept references
    - timestamp: source-side timestamp (ISO 8601)
    - bundle: bundle name (default = "devhub-gitea")
    """

    source: str
    type: str  # 8종 enum
    name: str
    title: str
    body: str
    frontmatter: dict[str, Any]
    raw_refs: list[str]
    timestamp: str  # ISO 8601
    bundle: str


class SourcePluginError(Exception):
    """Source plugin errors (connection, fetch, normalize)."""


class SourcePlugin(ABC):
    """External system source plugin ABC.

    Lifecycle:
    1. instantiate plugin
    2. await connect(credential) — Gitea: API token 검증 / homelab_mock: in-memory setup
    3. await fetch(since) — incremental fetch (since=None → full sync)
    4. await normalize(raw) — raw dict → ConceptDict (1:1 per external item)
    5. (caller) ConceptDict → OKF frontmatter + body + cross-link

    Per umbrella doc §3.8.1:
    - name: source plugin identifier
    - health_check(): connectivity check
    - last_error: dict | None (for status endpoint)
    """

    name: str

    @abstractmethod
    async def connect(self, credential: dict) -> None:
        """Connect to external system. May raise SourcePluginError on failure."""
        ...

    @abstractmethod
    async def fetch(self, since: datetime | None) -> list[dict]:
        """Fetch raw data from external system.

        Args:
            since: last sync timestamp (None = full sync)

        Returns: list of raw dicts (one per external item)
        """
        ...

    @abstractmethod
    async def normalize(self, raw: dict) -> ConceptDict:
        """Normalize raw dict to ConceptDict (1:1).

        Must produce 8 field: source / type / name / title / body / frontmatter / raw_refs / timestamp
        """
        ...

    async def health_check(self) -> dict:
        """Connectivity check. Returns {"healthy": bool, "last_error": dict | None}."""
        try:
            await self.connect({})
            return {"healthy": True, "last_error": None}
        except Exception as e:
            return {"healthy": False, "last_error": {"code": "connect_failed", "message": str(e)}}
