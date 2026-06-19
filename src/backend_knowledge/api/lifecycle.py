"""Concept lifecycle endpoints (umbrella doc §3.9.4 — archive / publish).

2 endpoint (audit event source for events 5 + 6):
1. POST /api/v0-2/concepts/{concept_id}/archive — audit.concept.archive
2. POST /api/v0-2/concepts/{concept_id}/publish — audit.concept.publish

PoC: stub implementation emits audit event without actual lifecycle state change.
Real archive/publish semantics (status field, viz.html update) is M-v0.2.3+ scope.
"""

from __future__ import annotations

from datetime import datetime, timezone
from typing import Literal

from fastapi import APIRouter, Depends, Path, Request, status
from pydantic import BaseModel, ConfigDict, Field

from ..audit.events import AuditEventType
from ..audit.logger import get_audit_logger
from ..auth.path_y import PathYUserContext
from ._common import get_path_y_context, make_envelope, require_path_y_context

router = APIRouter(prefix="/api/v0-2", tags=["lifecycle"])


class ArchiveRequest(BaseModel):
    model_config = ConfigDict(extra="forbid")
    reason: Literal["operator-manual", "superseded", "orphan"] = "operator-manual"
    note: str = Field(default="", max_length=500)


class ArchiveData(BaseModel):
    concept_id: str
    archived_at: datetime
    archived_by: str
    reason: str
    new_status: Literal["archived"] = "archived"


@router.post("/concepts/{concept_id:path}/archive", status_code=status.HTTP_200_OK)
async def post_archive(
    request: Request,
    body: ArchiveRequest,
    concept_id: str = Path(..., description="Concept ID (bundle/type/slug)"),
    ctx: PathYUserContext = Depends(require_path_y_context),
) -> dict:
    """FR: §3.9.4 archive a concept (PoC: audit event only, no state change)."""
    audit = get_audit_logger()
    audit.emit_simple(
        event_type=AuditEventType.CONCEPT_ARCHIVE,
        user_id=ctx.user_id,
        org_id=ctx.org_id,
        request_id=getattr(request.state, "request_id", None),
        ip=request.client.host if request.client else None,
        success=True,
        concept_id=concept_id,
        reason=body.reason,
        note=body.note,
    )
    return make_envelope(
        ArchiveData(
            concept_id=concept_id,
            archived_at=datetime.now(timezone.utc),
            archived_by=ctx.user_id,
            reason=body.reason,
        ).model_dump(mode="json"),
        request,
        ctx,
    )


class PublishRequest(BaseModel):
    model_config = ConfigDict(extra="forbid")
    version: int = Field(default=1, ge=1)
    note: str = Field(default="", max_length=500)


class PublishData(BaseModel):
    concept_id: str
    version: int
    published_at: datetime
    published_by: str
    new_status: Literal["published"] = "published"


@router.post("/concepts/{concept_id:path}/publish", status_code=status.HTTP_200_OK)
async def post_publish(
    request: Request,
    body: PublishRequest,
    concept_id: str = Path(..., description="Concept ID (bundle/type/slug)"),
    ctx: PathYUserContext = Depends(require_path_y_context),
) -> dict:
    """FR: §3.9.4 publish a concept (PoC: audit event only, no state change)."""
    audit = get_audit_logger()
    audit.emit_simple(
        event_type=AuditEventType.CONCEPT_PUBLISH,
        user_id=ctx.user_id,
        org_id=ctx.org_id,
        request_id=getattr(request.state, "request_id", None),
        ip=request.client.host if request.client else None,
        success=True,
        concept_id=concept_id,
        version=body.version,
        note=body.note,
    )
    return make_envelope(
        PublishData(
            concept_id=concept_id,
            version=body.version,
            published_at=datetime.now(timezone.utc),
            published_by=ctx.user_id,
        ).model_dump(mode="json"),
        request,
        ctx,
    )
