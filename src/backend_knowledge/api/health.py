"""Health check endpoints (umbrella doc §3.1 API 매트릭스 row 1 + §3.6.1).

- GET /health — public health check (no Path Y required)
- GET /health/protected — Path Y required (returns caller user_id + roles)
"""

from __future__ import annotations

from datetime import datetime, timezone

from fastapi import APIRouter, Depends, Header, Request
from pydantic import BaseModel

from .. import __version__
from ..auth.path_y import PathYUserContext, PathYExpiredError, PathYValidationError, get_path_y_validator
from fastapi import HTTPException, status
from ._common import make_envelope

router = APIRouter(tags=["health"])


class HealthData(BaseModel):
    """Health check response data."""

    status: str
    version: str
    timestamp: datetime
    uptime_seconds: float


@router.get("/health")
async def health(request: Request) -> dict:
    """Public health check (no Path Y)."""
    return make_envelope(
        {
            "status": "ok",
            "version": __version__,
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "uptime_seconds": 0.0,  # TODO: track process start time
            "path_y_validated": False,  # public health, no Path Y
        },
        request,
        ctx=None,
    )


class ProtectedHealthData(BaseModel):
    """Path Y validated health check response."""

    status: str
    user_id: str
    org_id: str
    roles: list[str]
    project_ids: list[str]


@router.get("/health/protected")
async def health_protected(
    request: Request,
    x_devhub_user_context: str | None = Header(None, alias="X-DevHub-User-Context"),
) -> dict:
    """Path Y required health check (returns caller user_id + roles)."""
    if not x_devhub_user_context:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={"code": "E_VALIDATION", "message": "X-DevHub-User-Context required"},
        )
    validator = get_path_y_validator()
    try:
        ctx = validator.validate(x_devhub_user_context)
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

    return make_envelope(
        ProtectedHealthData(
            status="ok",
            user_id=ctx.user_id,
            org_id=ctx.org_id,
            roles=ctx.roles,
            project_ids=ctx.project_ids,
        ).model_dump(mode="json"),
        request,
        ctx,
    )
