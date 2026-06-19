"""Source plugin registry (umbrella doc §3.8.1 정공법).

5 source plugin 등록 + name 기반 dispatch.
PoC default: homelab_mock 만 사용 가능 (Gitea 4 sub-plugin 은 GITEA_URL + GITEA_TOKEN 설정 시 활성화).
"""

from __future__ import annotations

from ._base import SourcePlugin

# Registry: name → SourcePlugin class
SOURCES: dict[str, type[SourcePlugin]] = {}


def register_source(cls: type[SourcePlugin]) -> type[SourcePlugin]:
    """Decorator: source plugin class 등록.

    Usage:
        @register_source
        class GiteaIssueSource(SourcePlugin):
            name = "gitea_issue"
            ...
    """
    if not cls.name:
        raise ValueError(f"SourcePlugin class {cls.__name__} must define `name` class variable")
    if cls.name in SOURCES:
        raise ValueError(f"duplicate source name: {cls.name}")
    SOURCES[cls.name] = cls
    return cls


def get_source(name: str) -> SourcePlugin:
    """Get source plugin instance by name.

    Raises KeyError if name not registered.
    """
    if name not in SOURCES:
        available = ", ".join(sorted(SOURCES.keys()))
        raise KeyError(f"unknown source: {name!r}. available: {available or '(none)'}")
    return SOURCES[name]()


def list_sources() -> list[str]:
    """List registered source names (sorted)."""
    return sorted(SOURCES.keys())


def clear_sources() -> None:
    """Clear all registered sources (for test)."""
    SOURCES.clear()
