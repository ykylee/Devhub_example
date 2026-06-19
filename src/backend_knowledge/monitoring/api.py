"""Monitoring API (umbrella doc §11.3 + §3.6.6.3).

2 endpoint:
1. GET /metrics — Prometheus text exposition format (gated by ENABLE_METRICS)
2. GET /api/v0-2/monitoring/alerts — current fired alerts (3-tier)
"""

from __future__ import annotations

from fastapi import APIRouter, Depends, Request, Response
from fastapi.responses import PlainTextResponse

from ..api._common import get_path_y_context, make_envelope
from ..auth.path_y import PathYUserContext
from ..config import get_settings
from .prometheus import collect_alerts, render_prometheus_exposition

router = APIRouter(tags=["monitoring"])


@router.get("/metrics", response_class=PlainTextResponse, include_in_schema=False)
async def prometheus_metrics() -> Response:
    """Prometheus /metrics endpoint (gated by ENABLE_METRICS config)."""
    settings = get_settings()
    if not settings.enable_metrics:
        return PlainTextResponse(
            content="# metrics endpoint disabled (set ENABLE_METRICS=true to enable)\n",
            status_code=200,
            media_type="text/plain; version=0.0.4; charset=utf-8",
        )
    return PlainTextResponse(
        content=render_prometheus_exposition(),
        media_type="text/plain; version=0.0.4; charset=utf-8",
    )


@router.get("/api/v0-2/monitoring/alerts")
async def get_alerts(request: Request, ctx: PathYUserContext | None = Depends(get_path_y_context)) -> dict:
    """Current fired alerts (3-tier severity from §11.3)."""
    alerts = collect_alerts()
    return make_envelope(
        {
            "alerts": alerts,
            "total": len(alerts),
            "by_severity": {
                "info": sum(1 for a in alerts if a["severity"] == "info"),
                "warning": sum(1 for a in alerts if a["severity"] == "warning"),
                "critical": sum(1 for a in alerts if a["severity"] == "critical"),
            },
        },
        request,
        ctx,
    )
