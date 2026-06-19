"""OKF (Open Knowledge Format) v0.1 library (umbrella doc §1.3 + ADR-0034).

ADR-0034 정합:
- 1 concept = 1 .md (Markdown + frontmatter)
- frontmatter `type` 1개 필수 (8종 enum: dataset / metric / api_endpoint / runbook / integration / event / reference / decision)
- DevHub 확장: `x_devhub_*` prefix 5 governance field (§3.6.4)

Bundle: 1 bundle = N concept .md (Markdown + frontmatter) in `var/bundles/{bundle}/{type}/{slug}.md`
Raw: 봉투 암호화 후 `var/raw/{source}/{sha256_prefix}.bin` (file mode)
"""

from .concept import Concept, ConceptReadResult, read_concept, write_concept
from .cross_link import CrossLink, extract_cross_links
from .frontmatter import (
    ConceptFrontmatter,
    ConceptType,
    Visibility,
    XDevHubCurator,
    XDevHubOwnerOrgId,
    XDevHubOwnerOrgUnitIds,
    XDevHubOwnerProjectIds,
    XDevHubOwnerUserId,
    XDevHubVisibility,
    parse_frontmatter,
    render_frontmatter,
)

__all__ = [
    "Concept",
    "ConceptReadResult",
    "ConceptType",
    "ConceptFrontmatter",
    "Visibility",
    "XDevHubCurator",
    "XDevHubOwnerOrgId",
    "XDevHubOwnerOrgUnitIds",
    "XDevHubOwnerProjectIds",
    "XDevHubOwnerUserId",
    "XDevHubVisibility",
    "CrossLink",
    "extract_cross_links",
    "parse_frontmatter",
    "render_frontmatter",
    "read_concept",
    "write_concept",
]
