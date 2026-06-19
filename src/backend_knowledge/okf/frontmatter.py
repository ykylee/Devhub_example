"""OKF frontmatter model (Pydantic v2, umbrella doc §3.2 + §3.3 + §3.6.4 정합).

OKF spec (ADR-0034) 정합:
- `type` 1개 필수 (8종 enum)
- 권장 field: title, description, resource, tags, timestamp (모두 옵션)

DevHub 확장 (umbrella doc §3.6.4):
- `x_devhub_source`: source plugin name (e.g., gitea_issue)
- `x_devhub_bundle`: bundle name (e.g., devhub-gitea)
- `x_devhub_version`: concept version (1부터 increment)
- `x_devhub_curator`: 5 curator_type (rule-based / llm-system_admin / human-self-user / human-org-head / human-system-admin)
- 5 governance field (§3.6.4):
  - x_devhub_owner_org_id (string, FK → org_units.unit_id)
  - x_devhub_owner_user_id (string, FK → users.user_id, nullable)
  - x_devhub_owner_org_unit_ids (array of string, recursive subtree)
  - x_devhub_owner_project_ids (array of string)
  - x_devhub_visibility (enum: public / org / personal / project)
"""

from __future__ import annotations

from enum import Enum
from typing import Any, Literal

import yaml
from pydantic import BaseModel, ConfigDict, Field, field_validator

# 8 OKF concept type (umbrella doc §3.2 + ADR-0034 §3.2)
ConceptType = Literal[
    "dataset",
    "metric",
    "api_endpoint",
    "runbook",
    "integration",
    "event",
    "reference",
    "decision",
]

# 4 visibility enum (umbrella doc §3.6.4)
Visibility = Literal["public", "org", "personal", "project"]

# 5 curator_type (umbrella doc §3.6.2)
XDevHubCurator = Literal[
    "rule-based",
    "llm-system_admin",
    "human-self-user",
    "human-org-head",
    "human-system-admin",
]


class ConceptFrontmatter(BaseModel):
    """OKF concept frontmatter (Pydantic v2 model).

    umbrella doc §3.6.4 + ADR-0034 §3.2 정합:
    - `type` 1개 필수 (8종 enum)
    - 권장: title, description, resource, tags, timestamp (모두 옵션)
    - DevHub 확장: x_devhub_source / x_devhub_bundle / x_devhub_version / x_devhub_curator
    - 5 governance field: x_devhub_owner_org_id / x_devhub_owner_user_id / x_devhub_owner_org_unit_ids / x_devhub_owner_project_ids / x_devhub_visibility
    """

    model_config = ConfigDict(
        extra="allow",  # OKF spec "extra keys 자유" 원칙
        populate_by_name=True,
        str_strip_whitespace=True,
    )

    # OKF spec required
    type: ConceptType

    # OKF spec recommended (all optional)
    title: str | None = None
    description: str | None = None
    resource: str | None = None
    tags: list[str] | None = None
    timestamp: str | None = None  # ISO 8601

    # DevHub extension: source attribution
    x_devhub_source: str | None = None  # "gitea_issue" | "gitea_wiki" | ...
    x_devhub_bundle: str | None = None  # "devhub-gitea" | ...
    x_devhub_version: int = Field(default=1, ge=1)
    x_devhub_curator: XDevHubCurator = "rule-based"

    # 5 governance field (umbrella doc §3.6.4)
    x_devhub_owner_org_id: str | None = None  # FK → org_units.unit_id
    x_devhub_owner_user_id: str | None = None  # FK → users.user_id, nullable
    x_devhub_owner_org_unit_ids: list[str] = Field(default_factory=list)
    x_devhub_owner_project_ids: list[str] = Field(default_factory=list)
    x_devhub_visibility: Visibility = "org"  # default = source plugin auto emit

    @field_validator("x_devhub_visibility", mode="before")
    @classmethod
    def _validate_visibility(cls, v: Any) -> str:
        """빈 문자열 / None → 'org' default."""
        if v is None or v == "":
            return "org"
        if v not in ("public", "org", "personal", "project"):
            raise ValueError(f"invalid visibility: {v}")
        return v


# Field alias (for type-safe exports)
XDevHubOwnerOrgId = str
XDevHubOwnerUserId = str | None
XDevHubOwnerOrgUnitIds = list[str]
XDevHubOwnerProjectIds = list[str]
XDevHubVisibility = Visibility


def parse_frontmatter(text: str) -> tuple[ConceptFrontmatter, str]:
    """Parse YAML frontmatter (between `---` markers) from Markdown.

    Returns: (frontmatter model, remaining Markdown body)
    """
    if not text.startswith("---"):
        # No frontmatter
        return ConceptFrontmatter(type="reference"), text  # type default

    # Find closing `---`
    parts = text.split("---", 2)
    if len(parts) < 3:
        # Malformed: only opening but no closing
        return ConceptFrontmatter(type="reference"), text

    yaml_block = parts[1].strip()
    body = parts[2].lstrip("\n")

    data: dict[str, Any] = {}
    if yaml_block:
        loaded = yaml.safe_load(yaml_block)
        if isinstance(loaded, dict):
            data = loaded

    return ConceptFrontmatter(**data), body


def render_frontmatter(frontmatter: ConceptFrontmatter) -> str:
    """Render frontmatter to YAML between `---` markers."""
    data = frontmatter.model_dump(exclude_none=False)
    # strip None values for cleaner output (but keep explicit defaults like x_devhub_visibility='org')
    clean = {k: v for k, v in data.items() if v is not None}
    yaml_str = yaml.safe_dump(clean, allow_unicode=True, sort_keys=False, default_flow_style=False)
    return f"---\n{yaml_str}---\n"
