"""Ingest API (FR-I-001 ~ FR-I-006, 6 endpoint, umbrella doc §2 FR §3.1 API 매트릭스 정합).

6 endpoint:
1. POST /api/v0-2/ingest/{source}/sync — sync trigger (FR-I-001)
2. GET /api/v0-2/ingest/{source}/status — sync status (FR-I-002)
3. POST /api/v0-2/raw — raw 등록 manual (FR-I-003, 봉투 암호화)
4. GET /api/v0-2/raw/{type}/{name} — raw 조회 (FR-I-004)
5. GET /api/v0-2/raw?source=...&since=... — raw list (FR-I-005)
6. DELETE /api/v0-2/raw/{id} — raw 삭제 (FR-I-006, soft archive 권장)

Path Y caller-provided user context:
- FR-I-001 (sync): Path Y 권장 (audit attribution)
- FR-I-002 (status): Path Y 권장
- FR-I-003 (raw register): Path Y 필수 (caller 권한 check)
- FR-I-004 (raw get): Path Y 필수 (visibility check)
- FR-I-005 (raw list): Path Y 필수 (caller scope filter)
- FR-I-006 (raw delete): Path Y 필수 (raw 삭제 권한)
"""

from __future__ import annotations

from datetime import datetime, timezone
from typing import Any, Literal

from fastapi import APIRouter, Depends, HTTPException, Header, Path, Query, Request, Response, status
from pydantic import BaseModel, ConfigDict, Field, field_validator

from ..auth.path_y import (
    PathYExpiredError,
    PathYUserContext,
    PathYValidationError,
    get_path_y_validator,
)
from ..logger import get_logger
from ..sources import get_source, list_sources
from ..storage import get_raw_store

logger = get_logger(__name__)
router = APIRouter(prefix="/api/v0-2", tags=["ingest"])


# === Path Y dependency (FastAPI dependency injection) ===

def get_path_y_context(
    request: Request,
    x_devhub_user_context: str | None = Header(None, alias="X-DevHub-User-Context"),
) -> PathYUserContext | None:
    """Path Y header validation. Returns None if header missing (for optional endpoints)."""
    if not x_devhub_user_context:
        return None
    validator = get_path_y_validator()
    try:
        return validator.validate(x_devhub_user_context)
    except PathYExpiredError as e:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail={"code": "E_UNAUTHORIZED", "message": f"X-DevHub-User-Context expired: {e}"},
        )
    except PathYValidationError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={"code": "E_VALIDATION", "message": f"X-DevHub-User-Context invalid: {e.reason}"},
        )


def require_path_y_context(
    ctx: PathYUserContext | None = Depends(get_path_y_context),
) -> PathYUserContext:
    """Path Y 필수. 400 E_VALIDATION if missing."""
    if ctx is None:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={"code": "E_VALIDATION", "message": "X-DevHub-User-Context required"},
        )
    return ctx


# === Envelope response (umbrella doc §3.4 Envelope = {envelope: {...}, data: T}) ===

class EnvelopeMeta(BaseModel):
    """공통 envelope metadata (umbrella doc §3.4)."""

    request_id: str
    timestamp: datetime
    api_version: Literal["v0-2"] = "v0-2"
    caller_user_id: str | None = None
    path_y_validated: bool = True


class Envelope(BaseModel):
    """공통 envelope wrapper (umbrella doc §3.4)."""

    model_config = ConfigDict(extra="forbid")

    envelope: EnvelopeMeta
    data: Any


def make_envelope(data: Any, request: Request, ctx: PathYUserContext | None) -> dict:
    """Build envelope response."""
    import uuid
    return {
        "envelope": {
            "request_id": getattr(request.state, "request_id", str(uuid.uuid4())),
            "timestamp": datetime.now(timezone.utc),
            "api_version": "v0-2",
            "caller_user_id": ctx.user_id if ctx else None,
            "path_y_validated": ctx is not None,
        },
        "data": data,
    }


# === FR-I-001: POST /ingest/{source}/sync (Path Y 권장) ===

class IngestSyncData(BaseModel):
    """FR-I-001 sync response data (umbrella doc §2 FR §4.1)."""

    synced: int
    failed: int
    raw_ids: list[str]
    next_sync_recommended: datetime | None = None
    errors: list[dict] = Field(default_factory=list)


@router.post("/ingest/{source}/sync", response_model=None)
async def post_sync(
    request: Request,
    source: str = Path(..., description="Source plugin name (e.g., gitea_issue)"),
    since: datetime | None = Query(None, description="Incremental sync 시작 시점. None = full sync"),
    dry_run: bool = Query(False, description="true 시 raw emit 안 함, 연결/credential 만 검증"),
    ctx: PathYUserContext | None = Depends(get_path_y_context),
) -> dict:
    """FR-I-001: Ingest sync trigger.

    Path Y: 권장 (audit attribution). Missing 시 200 OK (anonymous system sync).
    """
    if source not in list_sources():
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={"code": "E_VALIDATION", "message": f"unknown source: {source!r}. available: {list_sources()}"},
        )

    plugin = get_source(source)
    try:
        await plugin.connect({})
        if dry_run:
            return make_envelope(
                {
                    "synced": 0,
                    "failed": 0,
                    "raw_ids": [],
                    "next_sync_recommended": None,
                    "errors": [],
                    "dry_run": True,
                },
                request,
                ctx,
            )
        raws = await plugin.fetch(since=since)
        raw_store = get_raw_store()
        raw_ids: list[str] = []
        errors: list[dict] = []
        failed = 0
        for raw in raws:
            try:
                concept = await plugin.normalize(raw)
                # Save raw data via raw_store
                result = raw_store.save(
                    source=source,
                    type_=concept["type"],
                    name=concept["name"],
                    body=concept["body"],
                    registered_by=ctx.user_id if ctx else "anonymous",
                    frontmatter_override=concept.get("frontmatter", {}),
                    raw_refs=concept.get("raw_refs", []),
                    visibility="org",
                )
                raw_ids.append(result.raw_id)
            except Exception as e:
                failed += 1
                errors.append({"raw_name": raw.get("name", "unknown"), "code": "normalize_failed", "message": str(e)})

        return make_envelope(
            IngestSyncData(
                synced=len(raw_ids),
                failed=failed,
                raw_ids=raw_ids,
                next_sync_recommended=datetime.now(timezone.utc),
                errors=errors,
            ).model_dump(mode="json"),
            request,
            ctx,
        )
    except Exception as e:
        logger.error("ingest_sync_failed", source=source, error=str(e))
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={"code": "E_INTERNAL", "message": f"sync failed: {e}"},
        )


# === FR-I-002: GET /ingest/{source}/status (Path Y 권장) ===

class IngestStatusData(BaseModel):
    """FR-I-002 sync status response data (umbrella doc §2 FR §4.1)."""

    source: str
    last_sync: datetime | None
    next_sync: datetime | None
    state: Literal["idle", "syncing", "error", "disabled"]
    last_error: dict | None = None
    health: Literal["healthy", "degraded", "unhealthy"]
    metrics: dict = Field(default_factory=dict)


@router.get("/ingest/{source}/status")
async def get_status(
    request: Request,
    source: str = Path(..., description="Source plugin name"),
    ctx: PathYUserContext | None = Depends(get_path_y_context),
) -> dict:
    """FR-I-002: source plugin health check + last sync status."""
    if source not in list_sources():
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={"code": "E_VALIDATION", "message": f"unknown source: {source!r}"},
        )

    plugin = get_source(source)
    health_result = await plugin.health_check()
    health_str = "healthy" if health_result["healthy"] else "unhealthy"
    state_str = "idle" if health_result["healthy"] else "error"

    return make_envelope(
        IngestStatusData(
            source=source,
            last_sync=None,  # not tracked yet (audit log 추가 후)
            next_sync=None,
            state=state_str,
            last_error=health_result.get("last_error"),
            health=health_str,
            metrics={},
        ).model_dump(mode="json"),
        request,
        ctx,
    )


# === FR-I-003: POST /raw (Path Y 필수, 봉투 암호화) ===

class RawRegisterRequest(BaseModel):
    """FR-I-003 raw 등록 request (umbrella doc §2 FR §4.1)."""

    model_config = ConfigDict(extra="forbid")

    type: Literal["dataset", "metric", "api_endpoint", "runbook", "integration", "event", "reference", "decision"]
    name: str = Field(..., min_length=1, max_length=200, pattern=r"^[a-z0-9_]+$")
    source: str = Field(..., min_length=1)
    body: str = Field(..., min_length=1)
    frontmatter: dict = Field(default_factory=dict)
    raw_refs: list[str] = Field(default_factory=list)
    envelope_encryption: Literal["aes-256-gcm"] = "aes-256-gcm"


class RawRegisterData(BaseModel):
    """FR-I-003 raw 등록 response data."""

    raw_id: str
    sha256: str
    size: int
    envelope_encrypted: bool
    registered_at: datetime


@router.post("/raw", status_code=status.HTTP_201_CREATED)
async def post_raw(
    request: Request,
    body: RawRegisterRequest,
    ctx: PathYUserContext = Depends(require_path_y_context),
) -> dict:
    """FR-I-003: raw 등록 manual. 봉투 암호화 (AES-256-GCM) if KEK set."""
    raw_store = get_raw_store()
    try:
        result = raw_store.save(
            source=body.source,
            type_=body.type,
            name=body.name,
            body=body.body,
            registered_by=ctx.user_id,
            frontmatter_override=body.frontmatter,
            raw_refs=body.raw_refs,
        )
        return make_envelope(
            RawRegisterData(
                raw_id=result.raw_id,
                sha256=result.sha256,
                size=result.size,
                envelope_encrypted=result.envelope_encrypted,
                registered_at=result.registered_at,
            ).model_dump(mode="json"),
            request,
            ctx,
        )
    except Exception as e:
        logger.error("raw_register_failed", error=str(e))
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={"code": "E_ENCRYPTION_FAILED", "message": str(e)},
        )


# === FR-I-004: GET /raw/{type}/{name} (Path Y 필수) ===

class RawGetData(BaseModel):
    """FR-I-004 raw 조회 response data."""

    type: str
    name: str
    body: str
    raw_refs: list[str]
    sha256: str
    size: int
    registered_at: datetime
    registered_by: str
    visibility: str


@router.get("/raw/{type}/{name}")
async def get_raw(
    request: Request,
    type: str = Path(..., description="Concept type (8종 enum)"),
    name: str = Path(..., description="Concept name (slug pattern)"),
    source: str = Query("homelab_mock", description="Source plugin name (default: homelab_mock)"),
    ctx: PathYUserContext = Depends(require_path_y_context),
) -> dict:
    """FR-I-004: raw 조회. Path Y 필수 (visibility check)."""
    # PR 1: simple name → raw_id resolution (1:1 with homelab_mock or first match)
    # PR 2: full concept lookup with cross-link
    raw_store = get_raw_store()
    raw_ids = raw_store.list_source(source=source, limit=500)
    # Naive resolution: find first raw with matching name in frontmatter
    for raw_id in raw_ids:
        try:
            record = raw_store.load(source=source, raw_id=raw_id)
            # Check if name matches (via simple body parse — PR 2 will improve)
            if name.lower() in record.body.lower()[:200]:
                return make_envelope(
                    RawGetData(
                        type=type,
                        name=name,
                        body=record.body,
                        raw_refs=record.raw_refs,
                        sha256=record.sha256,
                        size=record.size,
                        registered_at=record.received_at,
                        registered_by=record.registered_by,
                        visibility=record.visibility,
                    ).model_dump(mode="json"),
                    request,
                    ctx,
                )
        except FileNotFoundError:
            continue

    raise HTTPException(
        status_code=status.HTTP_404_NOT_FOUND,
        detail={"code": "E_NOT_FOUND", "message": f"raw not found: {type}/{name} in source {source!r}"},
    )


# === FR-I-005: GET /raw?source=...&since=... (Path Y 필수) ===

class RawListItem(BaseModel):
    raw_id: str
    type: str
    name: str
    source: str
    sha256: str
    size: int
    registered_at: datetime
    visibility: str


class RawListData(BaseModel):
    items: list[RawListItem]
    total: int
    next_offset: int | None = None


@router.get("/raw")
async def list_raw(
    request: Request,
    source: str | None = Query(None, description="filter by source plugin name"),
    since: datetime | None = Query(None, description="registered_at >= since"),
    limit: int = Query(50, ge=1, le=500),
    offset: int = Query(0, ge=0),
    ctx: PathYUserContext = Depends(require_path_y_context),
) -> dict:
    """FR-I-005: raw list (filter + pagination). Path Y 필수 (caller scope filter)."""
    raw_store = get_raw_store()

    sources_to_check = [source] if source else ["homelab_mock"]
    items: list[RawListItem] = []
    for src in sources_to_check:
        raw_ids = raw_store.list_source(source=src, limit=limit + offset + 1)
        for raw_id in raw_ids:
            try:
                record = raw_store.load(source=src, raw_id=raw_id)
                if since and record.received_at < since:
                    continue
                # PR 1: simple type placeholder
                items.append(
                    RawListItem(
                        raw_id=raw_id,
                        type="reference",  # PR 2 will resolve from frontmatter
                        name=raw_id,
                        source=src,
                        sha256=record.sha256,
                        size=record.size,
                        registered_at=record.received_at,
                        visibility=record.visibility,
                    )
                )
            except FileNotFoundError:
                continue

    # Apply pagination
    paged = items[offset : offset + limit]
    next_offset = offset + limit if offset + limit < len(items) else None

    return make_envelope(
        RawListData(
            items=[item.model_dump(mode="json") for item in paged],
            total=len(items),
            next_offset=next_offset,
        ).model_dump(mode="json"),
        request,
        ctx,
    )


# === FR-I-006: DELETE /raw/{id} (Path Y 필수) ===

class RawDeleteData(BaseModel):
    raw_id: str
    deleted: bool
    deleted_at: datetime
    deleted_by: str
    impact: dict = Field(default_factory=dict)


@router.delete("/raw/{raw_id}")
async def delete_raw(
    request: Request,
    raw_id: str = Path(..., description="Raw ID (sha256 prefix + uuid)"),
    source: str = Query("homelab_mock", description="Source plugin name"),
    ctx: PathYUserContext = Depends(require_path_y_context),
) -> dict:
    """FR-I-006: raw 삭제 (soft archive 권장). Path Y 필수 (raw 삭제 권한)."""
    raw_store = get_raw_store()
    if not raw_store.exists(source=source, raw_id=raw_id):
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={"code": "E_NOT_FOUND", "message": f"raw not found: {source}/{raw_id}"},
        )

    # PR 1: hard delete (M-v0.2.1+ soft archive 권장)
    for ext in ("bin", "json"):
        file_path = raw_store.base_dir / source / f"{raw_id}.{ext}"
        if file_path.exists():
            file_path.unlink()
            break

    deleted_at = datetime.now(timezone.utc)
    logger.info("raw_deleted", raw_id=raw_id, source=source, deleted_by=ctx.user_id)
    return make_envelope(
        RawDeleteData(
            raw_id=raw_id,
            deleted=True,
            deleted_at=deleted_at,
            deleted_by=ctx.user_id,
            impact={"hard_delete": True, "soft_archive_recommended": False},
        ).model_dump(mode="json"),
        request,
        ctx,
    )
