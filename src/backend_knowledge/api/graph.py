"""Graph API (FR-G-001 ~ FR-G-004, 4 endpoint, umbrella doc §2 FR §3.1 API 매트릭스 정합).

4 endpoint:
1. GET /api/v0-2/graph/reverse/{concept_path} — reverse in-link list (FR-G-001)
2. GET /api/v0-2/graph/impact/{concept_path} — impact analysis (in + out + rank) (FR-G-002)
3. POST /api/v0-2/graph/reindex — reverse index 수동 rebuild (FR-G-003)
4. POST /api/v0-2/concepts/{id}/resolve-links — Pi LLM link resolve (MOCK for PR 2) (FR-G-004)

Path Y caller-provided user context:
- FR-G-001 (reverse): Path Y 권장
- FR-G-002 (impact): Path Y 권장
- FR-G-003 (reindex): Path Y 권장
- FR-G-004 (resolve-links): Path Y 권장 + audit log 필수
"""

from __future__ import annotations

import json
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Literal

from fastapi import APIRouter, Depends, HTTPException, Path, Request, status
from pydantic import BaseModel, ConfigDict, Field

from ..auth.path_y import (
    PathYUserContext,
    get_path_y_validator,
)
from ..config import get_settings
from ..logger import get_logger
from ..okf import extract_cross_links, parse_frontmatter
from ._bundle_store import (
    bundle_index_dir,
    load_concept_metadata,
    build_concept_id,
)
from ._common import get_path_y_context, make_envelope, require_path_y_context

logger = get_logger(__name__)
router = APIRouter(prefix="/api/v0-2", tags=["graph"])


# === Helper: read reverse index ===

def _read_reverse_index(bundle: str) -> dict:
    """Read reverse_index.json for a bundle. Returns empty dict if not exists."""
    reverse_index_path = bundle_index_dir(bundle) / "reverse_index.json"
    if not reverse_index_path.exists():
        return {}
    try:
        return json.loads(reverse_index_path.read_text(encoding="utf-8"))
    except json.JSONDecodeError:
        return {}


# === FR-G-001: GET /graph/reverse/{concept_path} ===

class ReverseLinkItem(BaseModel):
    """FR-G-001 reverse link item (umbrella doc §2 FR §4.4)."""

    source_concept: str
    type: Literal["explicit", "implicit", "tag", "wikilink"]
    section: str | None = None
    context: str
    created_at: datetime | None = None
    created_by: Literal["rule-based", "human", "pi-llm"]


class ReverseLinkData(BaseModel):
    """FR-G-001 reverse link response data (umbrella doc §2 FR §4.4)."""

    concept_path: str
    inlinks: list[ReverseLinkItem]
    count: int
    is_orphan: bool


@router.get("/graph/reverse/{concept_path:path}")
async def get_reverse(
    request: Request,
    concept_path: str = Path(..., description="Concept path (bundle/type/slug)"),
    ctx: PathYUserContext | None = Depends(get_path_y_context),
) -> dict:
    """FR-G-001: reverse in-link list 조회 (Path Y 권장)."""
    # Parse concept_path
    parts = concept_path.split("/")
    if len(parts) < 3:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={"code": "E_VALIDATION", "message": f"concept_path must be bundle/type/slug, got: {concept_path}"},
        )
    bundle = parts[0]
    reverse_index = _read_reverse_index(bundle)
    inlinks_raw = reverse_index.get(concept_path, [])
    inlinks: list[ReverseLinkItem] = []
    for rev in inlinks_raw:
        inlinks.append(
            ReverseLinkItem(
                source_concept=rev.get("source_concept", ""),
                type=rev.get("type", "explicit"),
                section=rev.get("section"),
                context=rev.get("context", ""),
                created_at=None,
                created_by="rule-based",  # PR 2 MOCK
            )
        )
    return make_envelope(
        ReverseLinkData(
            concept_path=concept_path,
            inlinks=inlinks,
            count=len(inlinks),
            is_orphan=len(inlinks) == 0,
        ).model_dump(mode="json"),
        request,
        ctx,
    )


# === FR-G-002: GET /graph/impact/{concept_path} ===

class ImpactForwardLinkItem(BaseModel):
    """FR-G-002 forward link item."""

    target_concept: str
    type: Literal["explicit", "implicit", "tag", "wikilink"]
    section: str | None = None
    context: str
    resolved: bool


class ImpactData(BaseModel):
    """FR-G-002 impact analysis response data (umbrella doc §2 FR §4.4)."""

    concept_path: str
    inlinks: list[ReverseLinkItem]
    outlinks: list[ImpactForwardLinkItem]
    is_orphan: bool
    inlink_count: int
    outlink_count: int
    rank_score: float
    last_impact_analysis: datetime


@router.get("/graph/impact/{concept_path:path}")
async def get_impact(
    request: Request,
    concept_path: str = Path(..., description="Concept path (bundle/type/slug)"),
    ctx: PathYUserContext | None = Depends(get_path_y_context),
) -> dict:
    """FR-G-002: impact analysis (in-link + out-link + rank score)."""
    parts = concept_path.split("/")
    if len(parts) < 3:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={"code": "E_VALIDATION", "message": f"concept_path must be bundle/type/slug, got: {concept_path}"},
        )
    bundle, type_, slug = parts[0], parts[1], parts[2]

    # In-links from reverse_index
    reverse_index = _read_reverse_index(bundle)
    inlinks_raw = reverse_index.get(concept_path, [])
    inlinks: list[ReverseLinkItem] = [
        ReverseLinkItem(
            source_concept=rev.get("source_concept", ""),
            type=rev.get("type", "explicit"),
            section=rev.get("section"),
            context=rev.get("context", ""),
            created_at=None,
            created_by="rule-based",
        )
        for rev in inlinks_raw
    ]

    # Out-links from concept body
    settings = get_settings()
    md_path = settings.var_dir / "bundles" / bundle / type_ / f"{slug}.md"
    outlinks: list[ImpactForwardLinkItem] = []
    if md_path.exists():
        try:
            text = md_path.read_text(encoding="utf-8")
            _, body = parse_frontmatter(text)
            cross_links = extract_cross_links(body, base_path=md_path)
            for cl in cross_links:
                target_slug = cl.target.split("#")[0].split("/")[-1]
                target_id = f"{bundle}/{type_}/{target_slug}"
                outlinks.append(
                    ImpactForwardLinkItem(
                        target_concept=target_id,
                        type=cl.type,
                        section=cl.section,
                        context=cl.context,
                        resolved=True,  # naive: assume same bundle
                    )
                )
        except Exception as e:
            logger.warning("impact_extract_outlinks_failed", path=str(md_path), error=str(e))

    # Rank score: simple inlink_count / max_inlink_count
    inlink_count = len(inlinks)
    outlink_count = len(outlinks)
    max_inlink = max(inlink_count, 1)  # avoid div by zero
    rank_score = min(inlink_count / max_inlink, 1.0) if inlink_count > 0 else 0.0

    return make_envelope(
        ImpactData(
            concept_path=concept_path,
            inlinks=inlinks,
            outlinks=outlinks,
            is_orphan=inlink_count == 0,
            inlink_count=inlink_count,
            outlink_count=outlink_count,
            rank_score=rank_score,
            last_impact_analysis=datetime.now(timezone.utc),
        ).model_dump(mode="json"),
        request,
        ctx,
    )


# === FR-G-003: POST /graph/reindex ===

class ReindexRequest(BaseModel):
    """FR-G-003 reindex request body (umbrella doc §2 FR §4.4)."""

    model_config = ConfigDict(extra="forbid")

    full_scan: bool = True
    bundle: str | None = None


class ReindexStats(BaseModel):
    """FR-G-003 reindex statistics (umbrella doc §2 FR §4.4)."""

    bundles_scanned: int
    concepts_scanned: int
    links_extracted: int
    reverse_index_entries: int
    orphans_detected: int
    unresolved_links: int
    duration_ms: int


class ReindexData(BaseModel):
    """FR-G-003 reindex response data (umbrella doc §2 FR §4.4)."""

    status: Literal["completed", "partial", "failed"]
    generated_at: datetime
    stats: ReindexStats
    errors: list[dict] = Field(default_factory=list)


@router.post("/graph/reindex", response_model=None)
async def post_reindex(
    request: Request,
    body: ReindexRequest,
    ctx: PathYUserContext | None = Depends(get_path_y_context),
) -> dict:
    """FR-G-003: reverse index 수동 rebuild (full scan). Path Y 권장."""
    import time
    start = time.time()
    settings = get_settings()
    bundles_dir = settings.var_dir / "bundles"

    bundle_filter = [body.bundle] if body.bundle else None
    bundles_scanned = 0
    concepts_scanned = 0
    links_extracted = 0
    reverse_index_entries = 0
    orphans_detected = 0
    errors: list[dict] = []

    if not bundles_dir.exists():
        return make_envelope(
            ReindexData(
                status="failed",
                generated_at=datetime.now(timezone.utc),
                stats=ReindexStats(
                    bundles_scanned=0, concepts_scanned=0, links_extracted=0,
                    reverse_index_entries=0, orphans_detected=0, unresolved_links=0,
                    duration_ms=int((time.time() - start) * 1000),
                ),
                errors=[{"code": "E_NOT_FOUND", "message": "no bundles directory"}],
            ).model_dump(mode="json"),
            request,
            ctx,
        )

    for bundle_dir in bundles_dir.iterdir():
        if not bundle_dir.is_dir() or bundle_dir.name.startswith("."):
            continue
        if bundle_filter and bundle_dir.name not in bundle_filter:
            continue
        bundles_scanned += 1
        # Reuse rebuild logic from curate.py
        reverse_index: dict[str, list[dict]] = {}
        md_files = sorted(bundle_dir.glob("*/*.md"))
        for md_file in md_files:
            concepts_scanned += 1
            slug = md_file.stem
            type_ = md_file.parent.name
            try:
                text = md_file.read_text(encoding="utf-8")
                _, body_md = parse_frontmatter(text)
                cross_links = extract_cross_links(body_md, base_path=md_file)
                for cl in cross_links:
                    target_slug = cl.target.split("#")[0].split("/")[-1]
                    target_id = f"{bundle_dir.name}/{type_}/{target_slug}"
                    if target_id not in reverse_index:
                        reverse_index[target_id] = []
                    reverse_index[target_id].append(
                        {
                            "source_concept": f"{bundle_dir.name}/{type_}/{slug}",
                            "type": cl.type,
                            "section": cl.section,
                            "context": cl.context,
                        }
                    )
                    links_extracted += 1
            except Exception as e:
                errors.append({"bundle": bundle_dir.name, "file": str(md_file), "error": str(e)})

        # Count orphans
        for slug_type in [(md.parent.name, md.stem) for md in md_files]:
            type_, slug = slug_type
            cid = f"{bundle_dir.name}/{type_}/{slug}"
            if cid not in reverse_index:
                orphans_detected += 1
            else:
                reverse_index_entries += len(reverse_index[cid])

        # Write reverse_index.json
        if body.full_scan:
            index_dir = bundle_index_dir(bundle_dir.name)
            index_dir.mkdir(exist_ok=True)
            (index_dir / "reverse_index.json").write_text(
                json.dumps(reverse_index, ensure_ascii=False, indent=2),
                encoding="utf-8",
            )

    duration_ms = int((time.time() - start) * 1000)
    return make_envelope(
        ReindexData(
            status="completed" if not errors else "partial",
            generated_at=datetime.now(timezone.utc),
            stats=ReindexStats(
                bundles_scanned=bundles_scanned,
                concepts_scanned=concepts_scanned,
                links_extracted=links_extracted,
                reverse_index_entries=reverse_index_entries,
                orphans_detected=orphans_detected,
                unresolved_links=0,  # PR 3 will add unresolved tracking
                duration_ms=duration_ms,
            ),
            errors=errors,
        ).model_dump(mode="json"),
        request,
        ctx,
    )


# === FR-G-004: POST /concepts/{id}/resolve-links (Pi LLM MOCK) ===

class ResolveLinksRequest(BaseModel):
    """FR-G-004 resolve-links request body (umbrella doc §2 FR §4.4, §3.5.7 Pi LLM confirm)."""

    model_config = ConfigDict(extra="forbid")

    mode: Literal["dry-run", "confirm", "auto-apply"] = "dry-run"
    selected_rank: int = Field(1, ge=1, le=3)
    confidence_threshold: float = Field(0.9, ge=0.0, le=1.0)
    max_candidates: int = Field(10, ge=1, le=50)


class ResolveLinksCandidate(BaseModel):
    """FR-G-004 Pi LLM 추천 candidate (umbrella doc §2 FR §4.4)."""

    rank: int
    target_concept: str
    confidence: float
    reasoning: str


class ResolveLinksData(BaseModel):
    """FR-G-004 resolve-links response data (umbrella doc §2 FR §4.4)."""

    concept_id: str
    mode: str
    candidates: list[ResolveLinksCandidate] = Field(default_factory=list)
    applied: bool = False
    applied_at: datetime | None = None
    error: str | None = None


@router.post("/concepts/{concept_id:path}/resolve-links", response_model=None)
async def post_resolve_links(
    request: Request,
    body: ResolveLinksRequest,
    concept_id: str = Path(..., description="Concept ID (bundle/type/slug)"),
    ctx: PathYUserContext = Depends(require_path_y_context),
) -> dict:
    """FR-G-004: Pi LLM link resolve (MOCK for PR 2, M-v0.2.3+).

    PR 2 MOCK: returns 1 MOCK candidate with confidence 0.5.
    PR 3: implement real Pi LLM (or rule-based fallback).
    """
    # MOCK: 1 candidate with 0.5 confidence
    candidates = [
        ResolveLinksCandidate(
            rank=1,
            target_concept="MOCK_TARGET",
            confidence=0.5,
            reasoning="PR 2 MOCK: real Pi LLM integration in M-v0.2.3+",
        )
    ]
    applied = body.mode == "auto-apply"  # MOCK: auto-apply returns applied=True
    return make_envelope(
        ResolveLinksData(
            concept_id=concept_id,
            mode=body.mode,
            candidates=candidates,
            applied=applied,
            applied_at=datetime.now(timezone.utc) if applied else None,
            error=None,
        ).model_dump(mode="json"),
        request,
        ctx,
    )
