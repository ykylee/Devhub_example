"""Homelab mock source (M-v0.2.0 PoC default).

In-memory mock — no external system call. Returns 3 sample dataset/metric/runbook
concepts for testing and demo. M-v0.2.0+ 시 homelab 실제 instance 로 교체.
"""

from __future__ import annotations

from datetime import datetime, timezone
from typing import Any

from ..config import get_settings
from ..logger import get_logger
from ._base import ConceptDict, SourcePlugin, SourcePluginError
from .registry import register_source

logger = get_logger(__name__)


# In-memory mock data (PoC default)
_MOCK_CONCEPTS: list[dict[str, Any]] = [
    {
        "id": 1,
        "name": "homelab-dataset-cpu-metrics",
        "title": "Homelab CPU Metrics Dataset",
        "type": "dataset",
        "body": "# Homelab CPU Metrics\n\nSample dataset for v0.2.0 PoC mock.\n\n- Source: homelab\n- Metric: cpu_usage_percent\n- Period: 2026-06-01 to 2026-06-19",
        "tags": ["homelab", "cpu", "metrics"],
        "timestamp": "2026-06-19T00:00:00+00:00",
    },
    {
        "id": 2,
        "name": "homelab-metric-cpu-usage",
        "title": "Homelab CPU Usage Metric",
        "type": "metric",
        "body": "# Homelab CPU Usage\n\nPrometheus-style metric: `node_cpu_usage_percent{host=\"homelab-01\"} 0.42`\n\n- Threshold: warning 70%, critical 90%",
        "tags": ["homelab", "cpu", "prometheus"],
        "timestamp": "2026-06-19T00:00:00+00:00",
    },
    {
        "id": 3,
        "name": "homelab-runbook-cpu-high",
        "title": "Homelab High CPU Runbook",
        "type": "runbook",
        "body": "# Homelab High CPU\n\n## Detection\n- Alert: `cpu_usage > 90%` for 5 min\n- Source: [[homelab-metric-cpu-usage]]\n\n## Triage\n1. SSH to host\n2. `top` / `htop`\n3. Check recent processes\n\n## Mitigation\n- Kill runaway processes\n- Scale up via [[homelab-dataset-cpu-metrics]]",
        "tags": ["homelab", "cpu", "runbook"],
        "timestamp": "2026-06-19T00:00:00+00:00",
    },
]


@register_source
class HomelabMockSource(SourcePlugin):
    """In-memory mock source (PoC default, no external call).

    Use case: v0.2.0 PoC 운영 + M-v0.2.0+ homelab 실제 instance 교체.
    """

    name: str = "homelab_mock"

    def __init__(self) -> None:
        self._connected: bool = False
        self._last_error: dict | None = None
        self.settings = get_settings()

    async def connect(self, credential: dict) -> None:
        """In-memory connect (no-op)."""
        self._connected = True
        self._last_error = None
        logger.info("homelab_mock_connected", credential_keys=list(credential.keys()))

    async def fetch(self, since: datetime | None) -> list[dict]:
        """Return all mock concepts (filter by `since` if provided)."""
        if not self._connected:
            await self.connect({})

        results: list[dict] = []
        for concept in _MOCK_CONCEPTS:
            ts = datetime.fromisoformat(concept["timestamp"].replace("Z", "+00:00"))
            if since is None or ts > since:
                results.append(concept)
        logger.info("homelab_mock_fetched", since=since.isoformat() if since else None, count=len(results))
        return results

    async def normalize(self, raw: dict) -> ConceptDict:
        """Convert mock dict to ConceptDict."""
        if "id" not in raw or "name" not in raw:
            raise SourcePluginError(f"invalid raw dict: missing id/name: {raw}")
        return ConceptDict(
            source=self.name,
            type=raw.get("type", "reference"),
            name=raw["name"],
            title=raw.get("title", raw["name"]),
            body=raw.get("body", ""),
            frontmatter={
                "tags": raw.get("tags", []),
                "description": raw.get("title", ""),
            },
            raw_refs=[],
            timestamp=raw.get("timestamp", datetime.now(timezone.utc).isoformat()),
            bundle="homelab-mock",
        )

    async def health_check(self) -> dict:
        """Mock source 는 항상 healthy."""
        return {"healthy": True, "last_error": None}
