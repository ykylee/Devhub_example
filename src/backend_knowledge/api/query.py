"""Query API (FR-Q-001 ~ FR-Q-005, 5 endpoint, umbrella doc §2 FR §3.1 API 매트릭스 정합).

5 endpoint:
1. POST /api/v0-2/query — 자연어 query → context + answer (FR-Q-001)
2. GET /api/v0-2/concepts/{type}/{name} — concept 직접 조회 (FR-Q-002)
3. GET /api/v0-2/search?q=... — full-text search (FR-Q-003)
4. GET /api/v0-2/bundles/{bundle}/index.md — bundle index (FR-Q-004)
5. GET /api/v0-2/bundles/{bundle}/viz.html — viz.html viewer (FR-Q-005)

Path Y caller-provided user context:
- FR-Q-001 (query): Path Y 필수
- FR-Q-002 (concept): Path Y 필수
- FR-Q-003 (search): Path Y 필수
- FR-Q-004 (bundle index): Path Y 권장
- FR-Q-005 (viz.html): Path Y 권장
"""

from __future__ import annotations

import json
import re
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Literal

from fastapi import APIRouter, Depends, HTTPException, Path, Query, Request, status
from fastapi.responses import HTMLResponse, PlainTextResponse
from pydantic import BaseModel, ConfigDict, Field

from ..auth.path_y import (
    PathYUserContext,
    PathYExpiredError,
    PathYValidationError,
    get_path_y_validator,
)
from ..config import get_settings
from ..logger import get_logger
from ..okf import parse_frontmatter
from .curate import (
    _bundle_dir,
    _bundle_index_dir,
    _concept_meta_path,
    _find_concept_by_id,
    _load_concept_metadata,
    _build_concept_id,
)
from ._common import get_path_y_context, make_envelope, require_path_y_context

logger = get_logger(__name__)
router = APIRouter(prefix="/api/v0-2", tags=["query"])


# === Helper: search across all bundles ===

def _search_all_concepts(
    query: str | None = None,
    bundle: str | None = None,
    type_: str | None = None,
    limit: int = 50,
) -> list[dict]:
    """Full-text search across all bundles' concept metadata (PR 2 MOCK: simple substring match)."""
    settings = get_settings()
    bundles_dir = settings.var_dir / "bundles"
    results: list[dict] = []
    if not bundles_dir.exists():
        return results

    bundle_filter = [bundle] if bundle else None
    for bundle_dir in bundles_dir.iterdir():
        if not bundle_dir.is_dir() or bundle_dir.name.startswith("."):
            continue
        if bundle_filter and bundle_dir.name not in bundle_filter:
            continue
        for meta_path in bundle_dir.glob("*/*.meta.json"):
            try:
                meta = json.loads(meta_path.read_text(encoding="utf-8"))
            except json.JSONDecodeError:
                continue
            type_meta = meta.get("type", "")
            if type_ and type_ != type_meta:
                continue
            # Simple substring match (MOCK; PR 3 will add BM25)
            if query:
                searchable = (
                    meta.get("name", "")
                    + " "
                    + meta.get("frontmatter", {}).get("title", "")
                    + " "
                    + meta.get("frontmatter", {}).get("description", "")
                ).lower()
                if query.lower() not in searchable:
                    continue
            results.append(meta)
            if len(results) >= limit:
                return results
    return results


# === FR-Q-001: POST /query ===

class QueryRequest(BaseModel):
    """FR-Q-001 query request body (umbrella doc §2 FR §4.3)."""

    model_config = ConfigDict(extra="forbid")

    query: str = Field(..., min_length=1, max_length=2000)
    bundle: str | None = None
    type: list[str] | None = None
    source: list[str] | None = None
    top_k: int = Field(10, ge=1, le=50)
    include_raw: bool = False
    llm_synthesis: bool = False  # M-v0.2.3+ LLM 합성 (PR 2 MOCK: false only)


class QueryContext(BaseModel):
    """FR-Q-001 query context item (umbrella doc §2 FR §4.3)."""

    concept_id: str
    type: str
    title: str
    score: float
    snippet: str


class QueryData(BaseModel):
    """FR-Q-001 query response data (umbrella doc §2 FR §4.3)."""

    answer: str | None = None
    contexts: list[QueryContext]
    raw_contexts: list[dict] = Field(default_factory=list)
    query_metadata: dict = Field(default_factory=dict)


@router.post("/query", response_model=None)
async def post_query(
    request: Request,
    body: QueryRequest,
    ctx: PathYUserContext = Depends(require_path_y_context),
) -> dict:
    """FR-Q-001: 자연어 query → context + answer (MOCK: simple substring match, M-v0.2.0 PoC)."""
    import time
    from ..audit.events import AuditEventType
    from ..audit.logger import get_audit_logger

    query_start = time.time()
    if body.llm_synthesis:
        # MOCK: LLM 합성 시 answer = None (M-v0.2.3+ Pi LLM enrich 에서 추가)
        answer = None
    else:
        answer = None  # M-v0.2.0~v0.2.2: 단순 retrieval, answer 없음

    # Simple search (PR 2 MOCK: substring, no scope filter)
    matched = _search_all_concepts(query=body.query, bundle=body.bundle, type_=body.type[0] if body.type else None, limit=body.top_k)
    contexts: list[QueryContext] = []
    for i, meta in enumerate(matched):
        # Simple score (0.5~1.0) — PR 3 will add BM25
        score = 1.0 - (i * 0.05) if i < 20 else 0.0
        contexts.append(
            QueryContext(
                concept_id=_build_concept_id(meta.get("bundle", ""), meta.get("type", ""), meta.get("name", "")),
                type=meta.get("type", ""),
                title=meta.get("frontmatter", {}).get("title", meta.get("name", "")),
                score=score,
                snippet=meta.get("frontmatter", {}).get("description", ""),
            )
        )

    return make_envelope(
        QueryData(
            answer=answer,
            contexts=contexts,
            raw_contexts=[],
            query_metadata={
                "retrieval_method": "substring-mock",
                "duration_ms": 0.5,
                "scope_filter": "none (PR 2 MOCK)",
                "top_k": body.top_k,
            },
        ).model_dump(mode="json"),
        request,
        ctx,
    )
    get_audit_logger().emit_simple(
        event_type=AuditEventType.QUERY,
        user_id=ctx.user_id,
        org_id=ctx.org_id,
        request_id=getattr(request.state, "request_id", None),
        ip=request.client.host if request.client else None,
        success=True,
        query_text=body.query[:200],
        result_count=len(contexts),
        response_time_ms=round((time.time() - query_start) * 1000, 2),
        top_k=body.top_k,
    )


# === FR-Q-002: GET /concepts/{type}/{name} ===

class ConceptGetData(BaseModel):
    """FR-Q-002 concept get response data (umbrella doc §2 FR §4.3)."""

    concept_id: str
    type: str
    name: str
    bundle: str
    frontmatter: dict
    body: str
    cross_links_out: list[dict] = Field(default_factory=list)
    cross_links_in: list[dict] = Field(default_factory=list)
    version: int = 1
    created_at: datetime | None = None
    updated_at: datetime | None = None
    updated_by: str = ""
    visibility: str = "org"


@router.get("/concepts/{type}/{name}")
async def get_concept(
    request: Request,
    type: str = Path(..., description="Concept type (8종 enum)"),
    name: str = Path(..., description="Concept name (slug)"),
    bundle: str = Query("homelab-mock", description="Bundle name (default: homelab-mock)"),
    ctx: PathYUserContext = Depends(require_path_y_context),
) -> dict:
    """FR-Q-002: concept 직접 조회 (Path Y 필수 + visibility check)."""
    from ..audit.events import AuditEventType
    from ..audit.logger import get_audit_logger

    # Load concept metadata
    meta = _load_concept_metadata(bundle, type, name)
    if not meta:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={"code": "E_NOT_FOUND", "message": f"concept not found: {bundle}/{type}/{name}"},
        )

    # Visibility check (per §3.6.3 4-tier scope priority)
    visibility = meta.get("visibility", "org")
    if visibility == "personal" and meta.get("registered_by") != ctx.user_id:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail={"code": "E_FORBIDDEN", "message": "personal concept: caller user_id mismatch"},
        )
    # org / project / public: simplified check (MOCK; PR 3 will add full scope filter)

    # Read concept file
    settings = get_settings()
    md_path = settings.var_dir / "bundles" / bundle / type / f"{name}.md"
    if not md_path.exists():
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={"code": "E_NOT_FOUND", "message": f"concept file not found: {md_path}"},
        )
    text = md_path.read_text(encoding="utf-8")
    frontmatter, body = parse_frontmatter(text)

    return make_envelope(
        ConceptGetData(
            concept_id=_build_concept_id(bundle, type, name),
            type=type,
            name=name,
            bundle=bundle,
            frontmatter=frontmatter.model_dump(exclude_none=True),
            body=body,
            cross_links_out=[],  # PR 3 will populate
            cross_links_in=[],  # PR 3 will populate
            version=frontmatter.x_devhub_version,
            created_at=None,  # PR 3 will add
            updated_at=None,
            updated_by=meta.get("registered_by", ""),
            visibility=visibility,
        ).model_dump(mode="json"),
        request,
        ctx,
    )
    get_audit_logger().emit_simple(
        event_type=AuditEventType.CONCEPT_ACCESS,
        user_id=ctx.user_id,
        org_id=ctx.org_id,
        request_id=getattr(request.state, "request_id", None),
        ip=request.client.host if request.client else None,
        success=True,
        concept_id=f"{bundle}/{type}/{name}",
        bundle=bundle,
        type=type,
        visibility=visibility,
    )


# === FR-Q-003: GET /search ===

class SearchData(BaseModel):
    """FR-Q-003 search response data (umbrella doc §2 FR §4.3)."""

    hits: list[dict]
    total: int
    next_offset: int | None = None
    query_metadata: dict = Field(default_factory=dict)


@router.get("/search")
async def search(
    request: Request,
    q: str = Query(..., min_length=1, max_length=500, description="Search query"),
    bundle: str | None = Query(None, description="Filter by bundle"),
    type: list[str] | None = Query(None, description="Filter by type"),
    limit: int = Query(20, ge=1, le=100),
    offset: int = Query(0, ge=0),
    ctx: PathYUserContext = Depends(require_path_y_context),
) -> dict:
    """FR-Q-003: full-text search (Path Y 필수)."""
    matched = _search_all_concepts(query=q, bundle=bundle, type_=type[0] if type else None, limit=1000)
    paged = matched[offset : offset + limit]
    next_offset = offset + limit if offset + limit < len(matched) else None
    return make_envelope(
        SearchData(
            hits=[
                {
                    "concept_id": _build_concept_id(m.get("bundle", ""), m.get("type", ""), m.get("name", "")),
                    "type": m.get("type", ""),
                    "title": m.get("frontmatter", {}).get("title", m.get("name", "")),
                    "snippet": m.get("frontmatter", {}).get("description", ""),
                    "score": 0.0,  # PR 3 will add BM25 score
                    "bundle": m.get("bundle", ""),
                    "source": m.get("source", ""),
                }
                for m in paged
            ],
            total=len(matched),
            next_offset=next_offset,
            query_metadata={"search_method": "substring-mock", "duration_ms": 0.5},
        ).model_dump(mode="json"),
        request,
        ctx,
    )


# === FR-Q-004: GET /bundles/{bundle}/index.md ===

@router.get("/bundles/{bundle}/index.md", response_class=PlainTextResponse)
async def get_bundle_index(
    request: Request,
    bundle: str = Path(..., min_length=1, max_length=100, pattern=r"^[a-z0-9-]+$"),
    ctx: PathYUserContext | None = Depends(get_path_y_context),
) -> PlainTextResponse:
    """FR-Q-004: bundle index.md (Path Y 권장). Returns raw Markdown."""
    index_path = _bundle_index_dir(bundle) / "index.md"
    if not index_path.exists():
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={"code": "E_NOT_FOUND", "message": f"bundle index.md not found: {bundle}. Run POST /bundles/{bundle}/rebuild first."},
        )
    return PlainTextResponse(content=index_path.read_text(encoding="utf-8"), media_type="text/markdown; charset=utf-8")


# === FR-Q-005: GET /bundles/{bundle}/viz.html ===

@router.get("/bundles/{bundle}/viz.html", response_class=HTMLResponse)
async def get_bundle_viz(
    request: Request,
    bundle: str = Path(..., min_length=1, max_length=100, pattern=r"^[a-z0-9-]+$"),
    ctx: PathYUserContext | None = Depends(get_path_y_context),
) -> HTMLResponse:
    """FR-Q-005: bundle viz.html (Cytoscape.js v3.x + marked.js v5.x CDN, self-contained)."""
    viz_path = _bundle_index_dir(bundle) / "viz.html"
    if not viz_path.exists():
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={"code": "E_NOT_FOUND", "message": f"bundle viz.html not found: {bundle}. Run POST /bundles/{bundle}/rebuild first."},
        )
    return HTMLResponse(content=viz_path.read_text(encoding="utf-8"))
