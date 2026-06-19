"""Source plugins (umbrella doc §3.8 + ADR-0035 §3.2).

5 source plugin (M-v0.2.0 PoC):
1. gitea_repo_pull — Gitea pull request sync
2. gitea_issue — Gitea issue sync
3. gitea_wiki — Gitea wiki page sync
4. gitea_action — Gitea Actions / CI run sync
5. homelab_mock — In-memory mock (PoC default, no external call)

각 source:
- name: source plugin 식별자 (e.g., "gitea_issue")
- connect(credential): 외부 시스템 연결 (Gitea) / mock 확인 (homelab)
- fetch(since): 마지막 sync 시각 이후 변경분 fetch
- normalize(raw): raw → OKF concept (frontmatter + body)
"""

from ._base import ConceptDict, SourcePlugin, SourcePluginError
from .gitea_action import GiteaActionSource
from .gitea_issue import GiteaIssueSource
from .gitea_repo_pull import GiteaRepoPullSource
from .gitea_wiki import GiteaWikiSource
from .homelab_mock import HomelabMockSource
from .registry import (
    SOURCES,
    get_source,
    list_sources,
    register_source,
)

__all__ = [
    "ConceptDict",
    "SourcePlugin",
    "SourcePluginError",
    "GiteaActionSource",
    "GiteaIssueSource",
    "GiteaRepoPullSource",
    "GiteaWikiSource",
    "HomelabMockSource",
    "SOURCES",
    "get_source",
    "list_sources",
    "register_source",
]
