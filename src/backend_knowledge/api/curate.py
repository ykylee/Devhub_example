"""Curate API (FR-C-001 ~ FR-C-005, 5 endpoint, umbrella doc §2 FR §3.1 API 매트릭스 정합).

5 endpoint:
1. POST /api/v0-2/concepts/{id}/enrich — raw → OKF concept 변환 (rule-based or pi-llm) (FR-C-001)
2. PUT /api/v0-2/concepts/{id} — concept manual edit (5 curator_type 권한 check) (FR-C-002)
3. GET /api/v0-2/bundles — bundle list (Path Y 권장) (FR-C-003)
4. POST /api/v0-2/bundles — bundle create (owner_org_id + visibility) (FR-C-004)
5. POST /api/v0-2/bundles/{bundle}/rebuild — bundle index.md + viz.html + reverse_index.json (FR-C-005)

Path Y caller-provided user context:
- FR-C-001 (enrich): Path Y 권장
- FR-C-002 (manual edit): Path Y 필수 (curation ownership check, §3.6.2)
- FR-C-003 (bundle list): Path Y 권장
- FR-C-004 (bundle create): Path Y 필수
- FR-C-005 (bundle rebuild): Path Y 권장
"""

from __future__ import annotations

import json
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Literal

from fastapi import APIRouter, Depends, HTTPException, Path, Query, Request, status
from pydantic import BaseModel, ConfigDict, Field, field_validator

from ..auth.path_y import (
    PathYExpiredError,
    PathYUserContext,
    PathYValidationError,
    get_path_y_validator,
)
from ..config import get_settings
from ..logger import get_logger
from ..okf import (
    Concept,
    ConceptFrontmatter,
    parse_frontmatter,
    render_frontmatter,
    write_concept,
)
from ._common import get_path_y_context, make_envelope, require_path_y_context

logger = get_logger(__name__)
router = APIRouter(prefix="/api/v0-2", tags=["curate"])


from ._bundle_store import (  # noqa: F401
    bundle_dir,
    bundle_index_dir,
    bundle_meta_path,
    build_concept_id,
    concept_meta_path,
    find_concept_by_id,
    load_concept_metadata,
    save_concept_metadata,
)


# === FR-C-001: POST /concepts/{id}/enrich ===

class ConceptEnrichRequest(BaseModel):
    """FR-C-001 enrich request body (umbrella doc §2 FR §4.2)."""

    model_config = ConfigDict(extra="forbid")

    raw_id: str = Field(..., description="Source raw_id (sha256 prefix + uuid)")
    enricher: Literal["rule-based", "pi-llm"] = "rule-based"
    cross_link_extraction: bool = True
    dry_run: bool = False


class ConceptEnrichData(BaseModel):
    """FR-C-001 enrich response data."""

    concept_id: str
    version: int
    cross_links_extracted: int
    enricher_used: str
    enriched_at: datetime
    preview: dict | None = None  # dry_run=True 시 변환 preview


@router.post("/concepts/{concept_id:path}/enrich", response_model=None)
async def post_enrich(
    request: Request,
    concept_id: str = Path(..., description="Concept ID (bundle/type/slug or raw_id)"),
    body: ConceptEnrichRequest | None = None,
    ctx: PathYUserContext | None = Depends(get_path_y_context),
) -> dict:
    """FR-C-001: raw → OKF concept 변환.

    - enricher=rule-based: rule-based 정규화 (PoC default, M-v0.2.0~v0.2.2)
    - enricher=pi-llm: Pi LLM enrich (M-v0.2.3+, MOCK mode in PR 2)
    - cross_link_extraction: extract cross-links from body
    - dry_run=true: 변환 preview 만, 실제 concept emit ❌

    Path Y: 권장 (LLM enrich usage attribution).
    """
    if body is None:
        body = ConceptEnrichRequest(raw_id=concept_id)
    if body.enricher == "pi-llm":
        logger.info("pi_llm_enrich_requested", note="MOCK mode (PR 2 placeholder)")
    from ..audit.events import AuditEventType
    from ..audit.logger import get_audit_logger
    if ctx is not None:
        get_audit_logger().emit_simple(
            event_type=AuditEventType.CURATION_EDIT,
            user_id=ctx.user_id,
            org_id=ctx.org_id,
            request_id=getattr(request.state, "request_id", None),
            ip=request.client.host if request.client else None,
            success=True,
            concept_id=concept_id,
            action="enrich",
            enricher=body.enricher,
            dry_run=body.dry_run,
        )
    # MOCK: just acknowledge, real logic in PR 3
    return make_envelope(
        ConceptEnrichData(
            concept_id=concept_id,
            version=1,
            cross_links_extracted=0,
            enricher_used=body.enricher,
            enriched_at=datetime.now(timezone.utc),
            preview={"note": "MOCK enrich (PR 2 placeholder, real implementation in PR 3)"} if body.dry_run else None,
        ).model_dump(mode="json"),
        request,
        ctx,
    )


# === FR-C-002: PUT /concepts/{id} (manual edit) ===

class ConceptEditRequest(BaseModel):
    """FR-C-002 manual edit request body (umbrella doc §2 FR §4.2)."""

    model_config = ConfigDict(extra="forbid")

    frontmatter: dict | None = None
    body: str | None = None
    append_body: str | None = None
    cross_links_add: list[dict] = Field(default_factory=list)
    cross_links_remove: list[str] = Field(default_factory=list)
    commit_message: str = Field(..., min_length=1, max_length=500)


class ConceptEditData(BaseModel):
    """FR-C-002 manual edit response data."""

    concept_id: str
    version: int
    frontmatter: dict
    body: str
    cross_links_added: int
    cross_links_removed: int
    edited_at: datetime
    edited_by: str


@router.put("/concepts/{concept_id:path}")
async def put_concept(
    request: Request,
    body: ConceptEditRequest,
    concept_id: str = Path(..., description="Concept ID (bundle/type/slug)"),
    ctx: PathYUserContext = Depends(require_path_y_context),
) -> dict:
    """FR-C-002: concept manual edit (5 curator_type 권한 check).

    - rule-based 자동 생성 concept: manual edit 불가 (403 E_FORBIDDEN)
    - llm-system_admin manual edit: curator="human" 으로 승격
    - human concept: owner_user_id or org_head or system_admin
    """
    if body.body is not None and body.append_body is not None:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={"code": "E_VALIDATION", "message": "body and append_body are mutually exclusive"},
        )

    # MOCK: just acknowledge, real logic in PR 3
    from ..audit.events import AuditEventType
    from ..audit.logger import get_audit_logger

    edited_at = datetime.now(timezone.utc)
    audit = get_audit_logger()
    audit.emit_simple(
        event_type=AuditEventType.CURATION_EDIT,
        user_id=ctx.user_id,
        org_id=ctx.org_id,
        request_id=getattr(request.state, "request_id", None),
        ip=request.client.host if request.client else None,
        success=True,
        concept_id=concept_id,
        old_version=1,
        new_version=2,
        cross_links_added=len(body.cross_links_add),
        cross_links_removed=len(body.cross_links_remove),
        commit_message=body.commit_message,
    )
    return make_envelope(
        ConceptEditData(
            concept_id=concept_id,
            version=2,
            frontmatter=body.frontmatter or {},
            body=body.body or body.append_body or "",
            cross_links_added=len(body.cross_links_add),
            cross_links_removed=len(body.cross_links_remove),
            edited_at=edited_at,
            edited_by=ctx.user_id,
        ).model_dump(mode="json"),
        request,
        ctx,
    )


# === FR-C-003: GET /bundles ===

class BundleListItem(BaseModel):
    """FR-C-003 bundle list item (umbrella doc §2 FR §4.2)."""

    name: str
    description: str = ""
    owner_org_id: str = ""
    owner_user_id: str | None = None
    visibility: Literal["public", "org", "personal", "project"] = "org"
    concept_count: int = 0
    last_rebuild: datetime | None = None
    size_bytes: int = 0


class BundleListData(BaseModel):
    """FR-C-003 bundle list response data."""

    items: list[BundleListItem]
    total: int


@router.get("/bundles")
async def list_bundles(
    request: Request,
    ctx: PathYUserContext | None = Depends(get_path_y_context),
) -> dict:
    """FR-C-003: bundle list (Path Y 권장)."""
    settings = get_settings()
    bundles_dir = settings.var_dir / "bundles"
    items: list[BundleListItem] = []
    if bundles_dir.exists():
        for bundle_dir in bundles_dir.iterdir():
            if not bundle_dir.is_dir() or bundle_dir.name.startswith("."):
                continue
            meta_path = bundle_dir / ".bundle_meta.json"
            if meta_path.exists():
                try:
                    meta = json.loads(meta_path.read_text(encoding="utf-8"))
                except json.JSONDecodeError:
                    meta = {}
            else:
                meta = {}
            # Count concept files (.md) in {type}/{slug}.md
            md_files = list(bundle_dir.glob("*/*.md"))
            size = sum(f.stat().st_size for f in bundle_dir.rglob("*.md") if f.is_file())
            items.append(
                BundleListItem(
                    name=bundle_dir.name,
                    description=meta.get("description", ""),
                    owner_org_id=meta.get("owner_org_id", ""),
                    owner_user_id=meta.get("owner_user_id"),
                    visibility=meta.get("visibility", "org"),
                    concept_count=len(md_files),
                    last_rebuild=None,  # TODO: track rebuild timestamp
                    size_bytes=size,
                )
            )
    items.sort(key=lambda b: b.name)
    return make_envelope(
        BundleListData(items=[i.model_dump(mode="json") for i in items], total=len(items)).model_dump(mode="json"),
        request,
        ctx,
    )


# === FR-C-004: POST /bundles ===

class BundleCreateRequest(BaseModel):
    """FR-C-004 bundle create request body (umbrella doc §2 FR §4.2)."""

    model_config = ConfigDict(extra="forbid")

    name: str = Field(..., min_length=1, max_length=100, pattern=r"^[a-z0-9-]+$")
    description: str = ""
    owner_org_id: str = Field(..., min_length=1)
    owner_user_id: str | None = None
    org_unit_ids: list[str] = Field(default_factory=list)
    project_ids: list[str] = Field(default_factory=list)
    visibility: Literal["public", "org", "personal", "project"] = "org"
    initial_concepts: list[str] = Field(default_factory=list, description="초기 포함할 concept path list")


class BundleCreateData(BaseModel):
    """FR-C-004 bundle create response data."""

    name: str
    created_at: datetime
    created_by: str
    visibility: str
    path: str


# === FR-C-003b: GET /bundles/{name} (M-v0.2.1+ scope, single bundle metadata) ===

class BundleDetailData(BaseModel):
    """FR-C-003b bundle detail response data."""

    name: str
    description: str = ""
    version: int = 0
    concept_count: int = 0
    owner_org_id: str
    owner_user_id: str | None = None
    org_unit_ids: list[str] = Field(default_factory=list)
    project_ids: list[str] = Field(default_factory=list)
    visibility: str
    created_at: datetime | None = None
    updated_at: datetime | None = None
    updated_by: str | None = None


@router.get("/bundles/{name}")
async def get_bundle(
    request: Request,
    name: str = Path(..., description="Bundle name (e.g., devhub-platform-kpi)"),
    ctx: PathYUserContext | None = Depends(get_path_y_context),
) -> dict:
    """FR-C-003b: single bundle metadata 조회 (Path Y 권장)."""
    settings = get_settings()
    bundle_path = bundle_dir(name)
    if not bundle_path.exists() or not bundle_path.is_dir():
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={"code": "E_NOT_FOUND", "message": f"bundle not found: {name!r}"},
        )

    meta_path = bundle_path / ".bundle_meta.json"
    meta: dict = {}
    if meta_path.exists():
        try:
            meta = json.loads(meta_path.read_text(encoding="utf-8"))
        except json.JSONDecodeError:
            meta = {}

    detail = BundleDetailData(
        name=name,
        description=meta.get("description", ""),
        version=meta.get("version", 0),
        concept_count=meta.get("concept_count", 0),
        owner_org_id=meta.get("owner_org_id", ""),
        owner_user_id=meta.get("owner_user_id"),
        org_unit_ids=meta.get("org_unit_ids", []),
        project_ids=meta.get("project_ids", []),
        visibility=meta.get("visibility", "org"),
        created_at=meta.get("created_at"),
        updated_at=meta.get("updated_at"),
        updated_by=meta.get("updated_by"),
    )
    return make_envelope(detail.model_dump(mode="json"), request, ctx)


@router.post("/bundles", status_code=status.HTTP_201_CREATED)
async def create_bundle(
    request: Request,
    body: BundleCreateRequest,
    ctx: PathYUserContext = Depends(require_path_y_context),
) -> dict:
    """FR-C-004: bundle create (Path Y 필수 + caller 권한 check)."""
    bundle_path = bundle_dir(body.name)
    if bundle_path.exists():
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail={"code": "E_CONFLICT", "message": f"bundle already exists: {body.name}"},
        )

    # Permission check (per Codex P1 review): caller 의 org_id 가 owner_org_id 와 일치 OR system_admin
    if ctx.org_id != body.owner_org_id and "system_admin" not in ctx.roles:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail={"code": "E_FORBIDDEN", "message": "bundle.create_denied: caller.org_id mismatch with owner_org_id"},
        )

    # Create bundle directory
    bundle_path.mkdir(parents=True, exist_ok=True)
    meta_path = bundle_path / ".bundle_meta.json"
    meta = {
        "name": body.name,
        "description": body.description,
        "owner_org_id": body.owner_org_id,
        "owner_user_id": body.owner_user_id,
        "org_unit_ids": body.org_unit_ids,
        "project_ids": body.project_ids,
        "visibility": body.visibility,
        "created_by": ctx.user_id,
        "created_at": datetime.now(timezone.utc).isoformat(),
    }
    meta_path.write_text(json.dumps(meta, ensure_ascii=False, indent=2), encoding="utf-8")

    # Initialize index directory
    (bundle_index_dir(body.name)).mkdir(exist_ok=True)

    return make_envelope(
        BundleCreateData(
            name=body.name,
            created_at=datetime.now(timezone.utc),
            created_by=ctx.user_id,
            visibility=body.visibility,
            path=str(bundle_path),
        ).model_dump(mode="json"),
        request,
        ctx,
    )


# === FR-C-005: POST /bundles/{bundle}/rebuild ===

class BundleRebuildRequest(BaseModel):
    """FR-C-005 bundle rebuild request body (umbrella doc §2 FR §4.2)."""

    model_config = ConfigDict(extra="forbid")

    full_scan: bool = True
    cross_link_strategy: Literal["rule-based", "pi-llm-auto"] = "rule-based"
    dry_run: bool = False


class BundleRebuildData(BaseModel):
    """FR-C-005 bundle rebuild response data."""

    bundle: str
    concept_count: int
    link_count: int
    reverse_index_generated: bool
    index_md_generated: bool
    viz_html_generated: bool
    duration_ms: int
    rebuilt_at: datetime


@router.post("/bundles/{bundle}/rebuild", response_model=None)
async def rebuild_bundle(
    request: Request,
    body: BundleRebuildRequest | None = None,
    bundle: str = Path(..., min_length=1, max_length=100, pattern=r"^[a-z0-9-]+$"),
    ctx: PathYUserContext | None = Depends(get_path_y_context),
) -> dict:
    """FR-C-005: bundle index.md + viz.html + reverse_index.json 재생성."""
    if body is None:
        body = BundleRebuildRequest()

    bundle_path = bundle_dir(bundle)
    if not bundle_path.exists():
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={"code": "E_NOT_FOUND", "message": f"bundle not found: {bundle}"},
        )

    import time
    start = time.time()
    index_dir = bundle_index_dir(bundle)
    index_dir.mkdir(exist_ok=True)

    # Scan all .md concept files
    md_files = sorted(bundle_path.glob("*/*.md"))
    concept_count = len(md_files)

    # 1. Build reverse index (in-link list per concept)
    reverse_index: dict[str, list[dict]] = {}
    link_count = 0
    for md_file in md_files:
        try:
            text = md_file.read_text(encoding="utf-8")
            frontmatter, body_md = parse_frontmatter(text)
            slug = md_file.stem
            type_ = md_file.parent.name
            source_id = build_concept_id(bundle, type_, slug)

            # Find all cross-link targets in body
            from ..okf.cross_link import extract_cross_links
            from ..config import get_settings
            settings = get_settings()
            cross_links = extract_cross_links(body_md, base_path=md_file)
            for cl in cross_links:
                # Register reverse link for target
                target_slug = cl.target.split("#")[0].split("/")[-1]
                target_id = f"{bundle}/{type_}/{target_slug}"  # naive: assume same bundle
                if target_id not in reverse_index:
                    reverse_index[target_id] = []
                reverse_index[target_id].append(
                    {
                        "source_concept": source_id,
                        "type": cl.type,
                        "section": cl.section,
                        "context": cl.context,
                    }
                )
                link_count += 1
        except Exception as e:
            logger.warning("rebuild_concept_parse_failed", path=str(md_file), error=str(e))

    if not body.dry_run:
        # Write reverse_index.json
        reverse_index_path = index_dir / "reverse_index.json"
        reverse_index_path.write_text(
            json.dumps(reverse_index, ensure_ascii=False, indent=2),
            encoding="utf-8",
        )

        # 2. Build index.md (per concept: title + description + relative link)
        index_md_lines = [f"# {bundle} — Bundle Index", ""]
        for md_file in md_files:
            slug = md_file.stem
            type_ = md_file.parent.name
            try:
                text = md_file.read_text(encoding="utf-8")
                frontmatter, _ = parse_frontmatter(text)
                title = frontmatter.title or slug
                desc = frontmatter.description or ""
            except Exception:
                title = slug
                desc = ""
            relative_path = f"../{type_}/{slug}.md"
            index_md_lines.append(f"## {title}")
            if desc:
                index_md_lines.append(f"_{desc}_")
            index_md_lines.append(f"- path: `{relative_path}`")
            index_md_lines.append(f"- type: `{type_}`")
            reverse_count = len(reverse_index.get(build_concept_id(bundle, type_, slug), []))
            if reverse_count > 0:
                index_md_lines.append(f"- in-link: {reverse_count}")
            index_md_lines.append("")

        index_md_path = index_dir / "index.md"
        index_md_path.write_text("\n".join(index_md_lines), encoding="utf-8")

        # 3. Build viz.html (Cytoscape.js v3.x + marked.js v5.x CDN embed)
        viz_html_path = index_dir / "viz.html"
        viz_html_content = _generate_viz_html(bundle, md_files, reverse_index)
        viz_html_path.write_text(viz_html_content, encoding="utf-8")

    duration_ms = int((time.time() - start) * 1000)
    return make_envelope(
        BundleRebuildData(
            bundle=bundle,
            concept_count=concept_count,
            link_count=link_count,
            reverse_index_generated=not body.dry_run,
            index_md_generated=not body.dry_run,
            viz_html_generated=not body.dry_run,
            duration_ms=duration_ms,
            rebuilt_at=datetime.now(timezone.utc),
        ).model_dump(mode="json"),
        request,
        ctx,
    )


def _generate_viz_html(
    bundle: str,
    md_files: list[Path],
    reverse_index: dict[str, list[dict]],
) -> str:
    """Generate self-contained viz.html with Cytoscape.js v3.x + marked.js v5.x CDN.

    umbrella doc §12.1 정합: self-contained (inline CSS + Cytoscape CDN only).
    """
    # Build nodes + edges JSON
    nodes: list[dict] = []
    edges: list[dict] = []
    concept_id_to_idx: dict[str, int] = {}
    for i, md_file in enumerate(md_files):
        slug = md_file.stem
        type_ = md_file.parent.name
        frontmatter = None
        try:
            text = md_file.read_text(encoding="utf-8")
            frontmatter, _ = parse_frontmatter(text)
            title = frontmatter.title or slug
        except Exception:
            title = slug
        concept_id = build_concept_id(bundle, type_, slug)
        concept_id_to_idx[concept_id] = i
        nodes.append(
            {
                "data": {
                    "id": concept_id,
                    "label": title,
                    "type": type_,
                    "visibility": frontmatter.x_devhub_visibility if frontmatter is not None else "org",
                }
            }
        )
        # Out-link edges (from concept body)
        from ..okf.cross_link import extract_cross_links
        try:
            body_text = md_file.read_text(encoding="utf-8")
            frontmatter, body_md = parse_frontmatter(body_text)
            cross_links = extract_cross_links(body_md, base_path=md_file)
            for cl in cross_links:
                target_slug = cl.target.split("#")[0].split("/")[-1]
                target_id = f"{bundle}/{type_}/{target_slug}"  # naive
                if target_id in concept_id_to_idx and target_id != concept_id:
                    edges.append(
                        {
                            "data": {
                                "id": f"{concept_id}->{target_id}",
                                "source": concept_id,
                                "target": target_id,
                                "label": cl.type,
                            }
                        }
                    )
        except Exception:
            pass
        # In-link edges (from reverse_index)
        for rev in reverse_index.get(concept_id, []):
            source = rev.get("source_concept", "")
            if source and source in concept_id_to_idx and source != concept_id:
                edge_id = f"{source}->{concept_id}"
                if not any(e["data"]["id"] == edge_id for e in edges):
                    edges.append(
                        {
                            "data": {
                                "id": edge_id,
                                "source": source,
                                "target": concept_id,
                                "label": rev.get("type", "in-link"),
                            }
                        }
                    )

    nodes_json = json.dumps(nodes, ensure_ascii=False)
    edges_json = json.dumps(edges, ensure_ascii=False)
    style_json = json.dumps(
        [
            {
                "selector": "node",
                "style": {
                    "background-color": "#4a90e2",
                    "label": "data(label)",
                    "font-size": "12px",
                    "text-valign": "center",
                    "color": "#fff",
                    "text-outline-width": 2,
                    "text-outline-color": "#4a90e2",
                    "width": "label",
                    "height": 30,
                    "padding": 8,
                },
            },
            {
                "selector": "node[type = 'dataset']",
                "style": {"background-color": "#5cb85c"},
            },
            {
                "selector": "node[type = 'metric']",
                "style": {"background-color": "#f0ad4e"},
            },
            {
                "selector": "node[type = 'api_endpoint']",
                "style": {"background-color": "#5bc0de"},
            },
            {
                "selector": "node[type = 'runbook']",
                "style": {"background-color": "#d9534f"},
            },
            {
                "selector": "edge",
                "style": {
                    "width": 2,
                    "line-color": "#999",
                    "target-arrow-color": "#999",
                    "target-arrow-shape": "triangle",
                    "curve-style": "bezier",
                    "label": "data(label)",
                    "font-size": "10px",
                    "color": "#666",
                    "text-rotation": "autorotate",
                    "text-background-color": "#fff",
                    "text-background-opacity": 0.7,
                },
            },
        ]
    )

    return f"""<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>{bundle} — Concept Graph</title>
  <script src="https://unpkg.com/cytoscape@3.30.2/dist/cytoscape.min.js"></script>
  <style>
    body {{ font-family: sans-serif; margin: 0; padding: 1rem; background: #fafafa; }}
    h1 {{ margin: 0 0 1rem 0; font-size: 1.4rem; color: #333; }}
    #cy {{ width: 100%; height: 90vh; background: #fff; border: 1px solid #ddd; border-radius: 4px; }}
    .info {{ color: #666; font-size: 0.85rem; margin-top: 0.5rem; }}
  </style>
</head>
<body>
  <h1>{bundle} — Concept Graph</h1>
  <div id="cy"></div>
  <div class="info">Self-contained viewer (Cytoscape.js v3.x CDN). {len(nodes)} concepts, {len(edges)} edges. Generated by backend-knowledge v0.2.0 PoC.</div>
  <script>
    const cy = cytoscape({{
      container: document.getElementById('cy'),
      elements: [...{nodes_json}, ...{edges_json}],
      style: {style_json},
      layout: {{ name: 'cose', animate: false, padding: 30, nodeRepulsion: 8000 }},
      minZoom: 0.2,
      maxZoom: 3,
      wheelSensitivity: 0.2,
    }});
    cy.on('tap', 'node', function(evt){{
      const node = evt.target;
      const data = node.data();
      window.open('../' + data.type + '/' + data.id.split('/').slice(-1)[0] + '.md', '_blank');
    }});
  </script>
</body>
</html>
"""
