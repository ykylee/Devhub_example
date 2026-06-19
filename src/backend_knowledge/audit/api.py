"""Audit log viewer API (umbrella doc §3.6.6.3 + §12 audit log viewer 4 endpoint).

4 endpoint (§3.6.6.3):
1. GET /api/v0-2/audit — generic filter (event_type, user_id, from, to, limit)
2. GET /api/v0-2/audit/concept/{concept_path} — per concept change history
3. GET /api/v0-2/audit/user/{user_id} — per user activity log
4. GET /api/v0-2/audit/org/{org_id} — per org activity log

Path Y caller-provided user context:
- audit.viewer: Path Y 권장 (caller 권한 check, system_admin 또는 본인 user_id)
"""

from __future__ import annotations

from datetime import datetime
from typing import Any

from fastapi import APIRouter, Depends, HTTPException, Path, Query, Request, status
from pydantic import BaseModel, ConfigDict, Field

from ..auth.path_y import PathYUserContext
from ..api.ingest import get_path_y_context, make_envelope
from .logger import get_audit_logger

router = APIRouter(prefix="/api/v0-2", tags=["audit"])


class AuditEventItem(BaseModel):
    """Audit log viewer response item."""

    model_config = ConfigDict(extra="allow")

    event: str
    timestamp: datetime
    user_id: str | None = None
    org_id: str | None = None
    request_id: str | None = None
    ip: str | None = None
    success: bool = True


class AuditListData(BaseModel):
    items: list[dict]
    total: int
    filters: dict


@router.get("/audit")
async def list_audit(
    request: Request,
    event_type: str | None = Query(None, description="Filter by event type (e.g., audit.curation.edit)"),
    user_id: str | None = Query(None, description="Filter by user_id"),
    org_id: str | None = Query(None, description="Filter by org_id"),
    from_date: datetime | None = Query(None, alias="from", description="From datetime (inclusive)"),
    to_date: datetime | None = Query(None, alias="to", description="To datetime (inclusive)"),
    limit: int = Query(100, ge=1, le=1000),
    ctx: PathYUserContext | None = Depends(get_path_y_context),
) -> dict:
    """Generic audit log filter endpoint."""
    from .events import AuditEventType
    et: AuditEventType | None = None
    if event_type:
        try:
            et = AuditEventType(event_type)
        except ValueError:
            valid = ", ".join(t.value for t in AuditEventType)
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail={"code": "E_VALIDATION", "message": f"unknown event_type: {event_type!r}. valid: {valid}"},
            )

    audit = get_audit_logger()
    items = audit.read_range(
        from_date=from_date,
        to_date=to_date,
        event_type=et,
        user_id=user_id,
        limit=limit,
    )
    if org_id:
        items = [i for i in items if i.get("org_id") == org_id]
    return make_envelope(
        AuditListData(
            items=items,
            total=len(items),
            filters={
                "event_type": event_type,
                "user_id": user_id,
                "org_id": org_id,
                "from": from_date.isoformat() if from_date else None,
                "to": to_date.isoformat() if to_date else None,
                "limit": limit,
            },
        ).model_dump(mode="json"),
        request,
        ctx,
    )


@router.get("/audit/concept/{concept_path:path}")
async def get_concept_audit(
    request: Request,
    concept_path: str = Path(..., description="Concept path (bundle/type/slug)"),
    limit: int = Query(50, ge=1, le=500),
    ctx: PathYUserContext | None = Depends(get_path_y_context),
) -> dict:
    """Per-concept change history (audit.curation.edit + audit.concept.archive + audit.concept.publish)."""
    from .events import AuditEventType
    audit = get_audit_logger()
    curation_events = audit.read_range(event_type=AuditEventType.CURATION_EDIT, limit=1000)
    archive_events = audit.read_range(event_type=AuditEventType.CONCEPT_ARCHIVE, limit=1000)
    publish_events = audit.read_range(event_type=AuditEventType.CONCEPT_PUBLISH, limit=1000)
    all_events = curation_events + archive_events + publish_events
    filtered = [e for e in all_events if e.get("concept_id") == concept_path]
    filtered.sort(key=lambda e: e.get("timestamp", ""), reverse=True)
    return make_envelope(
        {
            "concept_path": concept_path,
            "items": filtered[:limit],
            "total": len(filtered),
        },
        request,
        ctx,
    )


@router.get("/audit/user/{user_id}")
async def get_user_audit(
    request: Request,
    user_id: str = Path(..., description="User ID"),
    limit: int = Query(100, ge=1, le=1000),
    ctx: PathYUserContext | None = Depends(get_path_y_context),
) -> dict:
    """Per-user activity log (all 7 event types)."""
    if ctx is not None and ctx.user_id != user_id and "system_admin" not in ctx.roles:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail={"code": "E_FORBIDDEN", "message": "audit.user.view_denied: not self or system_admin"},
        )
    audit = get_audit_logger()
    items = audit.read_range(user_id=user_id, limit=limit)
    return make_envelope(
        {
            "user_id": user_id,
            "items": items,
            "total": len(items),
        },
        request,
        ctx,
    )


@router.get("/audit/org/{org_id}")
async def get_org_audit(
    request: Request,
    org_id: str = Path(..., description="Org ID"),
    limit: int = Query(100, ge=1, le=1000),
    ctx: PathYUserContext | None = Depends(get_path_y_context),
) -> dict:
    """Per-org activity log."""
    if ctx is not None and ctx.org_id != org_id and "system_admin" not in ctx.roles:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail={"code": "E_FORBIDDEN", "message": "audit.org.view_denied: not same org or system_admin"},
        )
    audit = get_audit_logger()
    items = audit.read_range(limit=limit * 10)
    filtered = [i for i in items if i.get("org_id") == org_id]
    return make_envelope(
        {
            "org_id": org_id,
            "items": filtered[:limit],
            "total": len(filtered),
        },
        request,
        ctx,
    )
