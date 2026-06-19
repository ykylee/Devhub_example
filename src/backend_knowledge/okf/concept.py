"""OKF concept reader/writer (Markdown + frontmatter).

umbrella doc §3.3 + §3.9 + ADR-0034 정합:
- 1 concept = 1 .md file (Markdown + YAML frontmatter)
- File path: `var/bundles/{bundle}/{type}/{slug}.md` (e.g., `devhub-gitea/dataset/dataset_gitea_issue_42.md`)
- Body: Markdown (cross-link: `[text](path)` or `[text](path#anchor)`)
"""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path

from .cross_link import CrossLink, extract_cross_links
from .frontmatter import ConceptFrontmatter, render_frontmatter, parse_frontmatter


@dataclass
class Concept:
    """In-memory concept (frontmatter + body + cross-links).

    umbrella doc §3.9 OKF concept 운영 lifecycle 정합:
    - Created → Reviewed → Published → Active → Archived
    - x_devhub_status (not in frontmatter, separate state file or memory)
    """

    frontmatter: ConceptFrontmatter
    body: str
    cross_links: list[CrossLink]
    source_path: Path | None = None  # 원본 file path (read 시에만)


@dataclass
class ConceptReadResult:
    """Read result (success or fallback)."""

    concept: Concept
    parse_warnings: list[str]


def read_concept(path: Path) -> ConceptReadResult:
    """Read concept from file path.

    Args:
        path: bundle concept path (var/bundles/{bundle}/{type}/{slug}.md)

    Returns: ConceptReadResult (concept + parse_warnings)
    """
    text = path.read_text(encoding="utf-8")
    frontmatter, body = parse_frontmatter(text)
    cross_links = extract_cross_links(body, base_path=path)
    warnings: list[str] = []
    if not text.startswith("---"):
        warnings.append("no_frontmatter")
    if not body.strip():
        warnings.append("empty_body")
    return ConceptReadResult(
        concept=Concept(frontmatter=frontmatter, body=body, cross_links=cross_links, source_path=path),
        parse_warnings=warnings,
    )


def write_concept(concept: Concept, path: Path) -> None:
    """Write concept to file path.

    Args:
        concept: Concept (frontmatter + body)
        path: bundle concept path (var/bundles/{bundle}/{type}/{slug}.md)
    """
    path.parent.mkdir(parents=True, exist_ok=True)
    frontmatter_text = render_frontmatter(concept.frontmatter)
    body_text = concept.body if concept.body.endswith("\n") else concept.body + "\n"
    path.write_text(frontmatter_text + body_text, encoding="utf-8")
