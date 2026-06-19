"""Cross-link extractor (umbrella doc §3.5.6 + ADR-0034).

Markdown cross-link:
- Inline link: `[text](path)` or `[text](path#anchor)` or `[text](path "title")`
- Reference link: `[text][ref]` + `[ref]: path`
- Wiki link: `[[slug]]` (Obsidian-style, optional)

Cross-link types (umbrella doc §3.5.6.4):
- explicit: standard markdown link `[text](path)`
- implicit: not currently used (placeholder)
- tag: frontmatter tag reference (e.g., `tags: [gitea, cicd]`)
- wikilink: `[[slug]]`
"""

from __future__ import annotations

import re
from dataclasses import dataclass
from pathlib import Path
from typing import Literal

# Markdown inline link: [text](path) or [text](path#anchor) or [text](path "title")
INLINE_LINK_RE = re.compile(
    r"\[([^\]]+)\]\(([^\s\)]+)(?:\s+\"([^\"]*)\")?\)",
)

# Wiki link: [[slug]] or [[slug|alias]]
WIKI_LINK_RE = re.compile(r"\[\[([^\]|]+)(?:\|([^\]]+))?\]\]")

# Reference link: [text][ref]
REFERENCE_LINK_RE = re.compile(r"\[([^\]]+)\]\[([^\]]+)\]")

# Reference definition: [ref]: path "title"
REFERENCE_DEF_RE = re.compile(
    r"^\s*\[([^\]]+)\]:\s+(\S+)(?:\s+\"([^\"]*)\")?",
    re.MULTILINE,
)

CrossLinkType = Literal["explicit", "implicit", "tag", "wikilink"]


@dataclass
class CrossLink:
    """Cross-link extracted from concept body."""

    type: CrossLinkType
    target: str  # path / slug / ref_id
    section: str | None = None  # #anchor
    context: str  # surrounding text (~50 char)


def _extract_context(body: str, match_start: int, match_end: int, context_chars: int = 50) -> str:
    """Extract surrounding context for the match."""
    start = max(0, match_start - context_chars)
    end = min(len(body), match_end + context_chars)
    context = body[start:end]
    if start > 0:
        context = "..." + context
    if end < len(body):
        context = context + "..."
    return context.replace("\n", " ").strip()


def extract_cross_links(body: str, base_path: Path | None = None) -> list[CrossLink]:
    """Extract all cross-links from Markdown body.

    Args:
        body: Markdown body
        base_path: concept file path (for relative link resolution, optional)

    Returns: list of CrossLink (deduplicated, source-relative)
    """
    links: list[CrossLink] = []
    seen: set[tuple[str, str, str | None]] = set()

    # 1. Reference link definitions: [ref]: path "title"
    ref_defs: dict[str, str] = {}
    for match in REFERENCE_DEF_RE.finditer(body):
        ref_id, path, _ = match.groups()
        ref_defs[ref_id.strip()] = path.strip()

    # 2. Inline links: [text](path)
    for match in INLINE_LINK_RE.finditer(body):
        text, target, _ = match.groups()
        target = target.strip()
        section = None
        if "#" in target:
            target, section = target.split("#", 1)
        # Skip external links (http://, https://, mailto:, etc.)
        if "://" in target or target.startswith("mailto:"):
            continue
        # Skip anchors-only links (#xxx)
        if not target:
            continue
        key = ("explicit", target, section)
        if key not in seen:
            seen.add(key)
            links.append(
                CrossLink(
                    type="explicit",
                    target=target,
                    section=section,
                    context=_extract_context(body, match.start(), match.end()),
                )
            )

    # 3. Reference links: [text][ref] (referencing ref_defs)
    for match in REFERENCE_LINK_RE.finditer(body):
        text, ref_id = match.groups()
        if ref_id in ref_defs:
            target = ref_defs[ref_id]
            section = None
            if "#" in target:
                target, section = target.split("#", 1)
            if "://" in target or target.startswith("mailto:"):
                continue
            key = ("explicit", target, section)
            if key not in seen:
                seen.add(key)
                links.append(
                    CrossLink(
                        type="explicit",
                        target=target,
                        section=section,
                        context=_extract_context(body, match.start(), match.end()),
                    )
                )

    # 4. Wiki links: [[slug]] or [[slug|alias]]
    for match in WIKI_LINK_RE.finditer(body):
        slug, alias = match.groups()
        slug = slug.strip()
        key = ("wikilink", slug, None)
        if key not in seen:
            seen.add(key)
            links.append(
                CrossLink(
                    type="wikilink",
                    target=slug,
                    section=None,
                    context=_extract_context(body, match.start(), match.end()),
                )
            )

    return links
