"""API 공통 helper (architecture.md §3 layer 격리 정공법).

7 file 의 cross-router import 정공법 회복 (§13.1 refactor):
- api/curate.py / api/query.py / api/graph.py / api/lifecycle.py
- api/health.py / api/audit/api.py / api/monitoring/api.py

원래 ingest.py 에 정의되어 있던 5 helper 를 본 module 로 추출:
- get_path_y_context (FastAPI dependency)
- require_path_y_context (FastAPI dependency)
- EnvelopeMeta / Envelope (Pydantic model, umbrella doc §3.4)
- make_envelope (utility)
"""

from __future__ import annotations

import uuid
from datetime import datetime, timezone
from typing import Any, Literal

from fastapi import Depends, Header, HTTPException, Request, status
from pydantic import BaseModel, ConfigDict

from ..audit.events import AuditEventType
from ..audit.logger import get_audit_logger
from ..auth.path_y import (
    PathYExpiredError,
    PathYUserContext,
    PathYValidationError,
    get_path_y_validator,
)


def get_path_y_context(
    request: Request,
    x_devhub_user_context: str | None = Header(None, alias="X-DevHub-User-Context"),
) -> PathYUserContext | None:
    """Path Y header validation. Returns None if header missing (for optional endpoints)."""
    if not x_devhub_user_context:
        return None
    validator = get_path_y_validator()
    try:
        ctx = validator.validate(x_devhub_user_context)
        get_audit_logger().emit_simple(
            event_type=AuditEventType.USER_LOGIN,
            user_id=ctx.user_id,
            org_id=ctx.org_id,
            request_id=getattr(request.state, "request_id", None),
            ip=request.client.host if request.client else None,
            success=True,
            roles=ctx.roles,
            issued_at=ctx.issued_at,
        )
        return ctx
    except PathYExpiredError as e:
        get_audit_logger().emit_simple(
            event_type=AuditEventType.USER_LOGIN,
            request_id=getattr(request.state, "request_id", None),
            ip=request.client.host if request.client else None,
            success=False,
            error_reason="expired",
            error=str(e),
        )
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail={"code": "E_UNAUTHORIZED", "message": f"X-DevHub-User-Context expired: {e}"},
        )
    except PathYValidationError as e:
        get_audit_logger().emit_simple(
            event_type=AuditEventType.USER_LOGIN,
            request_id=getattr(request.state, "request_id", None),
            ip=request.client.host if request.client else None,
            success=False,
            error_reason="invalid",
            error=e.reason,
        )
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


class EnvelopeMeta(BaseModel):
    request_id: str
    timestamp: datetime
    api_version: Literal["v0-2"] = "v0-2"
    caller_user_id: str | None = None
    path_y_validated: bool = True


class Envelope(BaseModel):
    model_config = ConfigDict(extra="forbid")

    envelope: EnvelopeMeta
    data: Any


def make_envelope(data: Any, request: Request, ctx: PathYUserContext | None) -> dict:
    """Build envelope response (umbrella doc §3.4)."""
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
