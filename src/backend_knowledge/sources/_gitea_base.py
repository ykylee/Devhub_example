"""Gitea base source (공통 로직).

4 Gitea sub-plugin (repo_pull / issue / wiki / action) 의 공통:
- httpx.AsyncClient + Bearer token
- mock mode (GITEA_URL or GITEA_TOKEN missing)
- health_check / connect / fetch / normalize
"""

from __future__ import annotations

from datetime import datetime
from typing import Any

import httpx

from ..config import get_settings
from ..logger import get_logger
from ._base import ConceptDict, SourcePlugin, SourcePluginError

logger = get_logger(__name__)


class GiteaBaseSource(SourcePlugin):
    """Gitea 4 sub-plugin 의 공통 base.

    - name: 서브클래스에서 정의 (e.g., "gitea_issue")
    - api_path: API endpoint path (e.g., "/api/v1/repos/{owner}/{repo}/issues")
    - normalize_one(raw): 서브클래스에서 정의 (raw → ConceptDict 1:1)
    """

    name: str = ""  # 서브클래스에서 override
    api_path: str = ""  # 서브클래스에서 override

    def __init__(self) -> None:
        self.settings = get_settings()
        self._client: httpx.AsyncClient | None = None
        self._connected: bool = False
        self._mock_mode: bool = False
        self._last_error: dict | None = None

    @property
    def is_mock_mode(self) -> bool:
        """GITEA_URL or GITEA_TOKEN 미설정 시 mock mode."""
        return self._mock_mode

    async def connect(self, credential: dict) -> None:
        """Gitea API 연결 (mock mode 자동 fallback)."""
        url = self.settings.gitea_url
        token = self.settings.gitea_token
        if not url or not token:
            self._mock_mode = True
            self._connected = True
            self._last_error = None
            logger.info(
                f"{self.name}_connected_mock",
                reason="GITEA_URL or GITEA_TOKEN missing → mock mode",
            )
            return
        self._client = httpx.AsyncClient(
            base_url=url,
            headers={"Authorization": f"Bearer {token}"},
            timeout=self.settings.gitea_timeout_seconds,
        )
        # Health check
        try:
            resp = await self._client.get("/api/v1/user")
            resp.raise_for_status()
            self._connected = True
            self._mock_mode = False
            self._last_error = None
            logger.info(f"{self.name}_connected", url=url)
        except httpx.HTTPError as e:
            self._connected = False
            self._last_error = {"code": "connect_failed", "message": str(e)}
            raise SourcePluginError(f"Gitea connect failed: {e}") from e

    async def fetch(self, since: datetime | None) -> list[dict]:
        """Fetch raw data from Gitea API (or mock)."""
        if not self._connected:
            await self.connect({})

        if self._mock_mode:
            return self._mock_fetch(since)

        if self._client is None:
            raise SourcePluginError("not connected (client is None)")

        owner = self.settings.gitea_default_owner
        repo = self.settings.gitea_default_repo
        path = self.api_path.format(owner=owner, repo=repo)

        try:
            params: dict[str, Any] = {}
            if since:
                # Gitea API 는 ISO 8601 string 받음
                params["since"] = since.isoformat()

            resp = await self._client.get(path, params=params)
            resp.raise_for_status()
            items = resp.json()
            if not isinstance(items, list):
                raise SourcePluginError(f"expected list, got {type(items).__name__}")
            logger.info(f"{self.name}_fetched", count=len(items), since=since.isoformat() if since else None)
            return items
        except httpx.HTTPError as e:
            self._last_error = {"code": "fetch_failed", "message": str(e)}
            raise SourcePluginError(f"Gitea fetch failed: {e}") from e

    def _mock_fetch(self, since: datetime | None) -> list[dict]:
        """Mock mode: return 1 sample raw dict per source.

        PoC: 1 sample per source (PoC default)
        """
        now = datetime.now()
        if self.name == "gitea_repo_pull":
            return [
                {
                    "id": 1001,
                    "number": 1,
                    "title": "Sample PR (mock)",
                    "body": "Mock pull request for v0.2.0 PoC.",
                    "state": "open",
                    "user": {"login": "devhub-bot"},
                    "created_at": now.isoformat(),
                    "updated_at": now.isoformat(),
                    "html_url": "https://gitea.example.com/devhub/example/pulls/1",
                }
            ]
        if self.name == "gitea_issue":
            return [
                {
                    "id": 2001,
                    "number": 1,
                    "title": "Sample issue (mock)",
                    "body": "Mock issue for v0.2.0 PoC.",
                    "state": "open",
                    "user": {"login": "devhub-bot"},
                    "created_at": now.isoformat(),
                    "updated_at": now.isoformat(),
                    "html_url": "https://gitea.example.com/devhub/example/issues/1",
                    "labels": [{"name": "backend-knowledge"}],
                }
            ]
        if self.name == "gitea_wiki":
            return [
                {
                    "title": "Sample-Wiki-Page-mock",
                    "content": "# Sample Wiki Page\n\nMock wiki page for v0.2.0 PoC.",
                    "last_commit": {"author": {"login": "devhub-bot"}, "created": now.isoformat()},
                    "html_url": "https://gitea.example.com/devhub/example/wiki/Sample-Wiki-Page-mock",
                }
            ]
        if self.name == "gitea_action":
            return [
                {
                    "id": 4001,
                    "name": "ci-mock",
                    "status": "success",
                    "conclusion": "success",
                    "head_branch": "main",
                    "event": "push",
                    "created_at": now.isoformat(),
                    "updated_at": now.isoformat(),
                    "html_url": "https://gitea.example.com/devhub/example/actions/runs/4001",
                }
            ]
        return []

    async def normalize(self, raw: dict) -> ConceptDict:
        """서브클래스에서 override (각 source 별 normalize 로직)."""
        raise NotImplementedError(f"{self.__class__.__name__} must implement normalize()")
