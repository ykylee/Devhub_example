"""v0.2.0 PoC FastAPI application entry.

umbrella doc §1.2 G7 + §3.3 + ADR-0035 정합:
- Python 3.13+ / FastAPI / Pydantic v2
- 완전 standalone (다른 backend 연결 ❌)
- Path Y caller-provided user context
- 5 source plugin (Gitea 4 + homelab_mock)
- 24 endpoint (PR 1: Ingest 6, PR 2: Curate/Query/Graph 14, PR 3: Audit/Monitoring/Operational)
"""

from __future__ import annotations

import time
from contextlib import asynccontextmanager

from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse

from . import __version__
from .api.curate import router as curate_router
from .api.graph import router as graph_router
from .api.health import router as health_router
from .api.ingest import router as ingest_router
from .api.lifecycle import router as lifecycle_router
from .api.query import router as query_router
from .audit.api import router as audit_router
from .config import get_settings
from .logger import configure_logging, get_logger
from .monitoring.api import router as monitoring_router

# Initialize logging
configure_logging()
logger = get_logger(__name__)

# Track process start time (for uptime)
_process_start = time.time()

# Initialize settings
settings = get_settings()


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Replace @app.on_event("startup") / @app.on_event("shutdown") (architecture.md §13.2 refactor)."""
    logger.info(
        "backend_knowledge_startup",
        version=__version__,
        var_dir=str(settings.var_dir),
        path_y_max_age_seconds=settings.path_y_max_age_seconds,
        gitea_mock_mode=not (settings.gitea_url and settings.gitea_token),
        raw_encryption_enabled=settings.raw_encryption_key is not None,
        enable_metrics=settings.enable_metrics,
    )
    for subdir in ("raw", "bundles", "audit", "log"):
        (settings.var_dir / subdir).mkdir(parents=True, exist_ok=True)
    yield
    logger.info("backend_knowledge_shutdown", uptime_seconds=time.time() - _process_start)


app = FastAPI(
    title="backend-knowledge",
    description="v0.2.0 PoC standalone backend knowledge tool (umbrella doc §1.2 G7 + ADR-0035)",
    version=__version__,
    docs_url="/docs",
    redoc_url="/redoc",
    openapi_url="/openapi.json",
    lifespan=lifespan,
)


# === Middleware: request_id, timing, access log ===

@app.middleware("http")
async def request_context_middleware(request: Request, call_next):
    """Per-request: generate request_id, log access, measure timing."""
    import uuid
    request_id = request.headers.get("X-Request-Id", str(uuid.uuid4()))
    request.state.request_id = request_id

    start = time.time()
    response = await call_next(request)
    duration_ms = (time.time() - start) * 1000.0

    logger.info(
        "http_request",
        request_id=request_id,
        method=request.method,
        path=request.url.path,
        status_code=response.status_code,
        duration_ms=round(duration_ms, 2),
    )

    response.headers["X-Request-Id"] = request_id
    return response


# === Exception handlers ===

@app.exception_handler(Exception)
async def unhandled_exception_handler(request: Request, exc: Exception) -> JSONResponse:
    """Catch-all exception handler (return JSON envelope)."""
    logger.error(
        "unhandled_exception",
        request_id=getattr(request.state, "request_id", "unknown"),
        path=request.url.path,
        error=str(exc),
        error_type=type(exc).__name__,
    )
    return JSONResponse(
        status_code=500,
        content={
            "envelope": {
                "request_id": getattr(request.state, "request_id", "unknown"),
                "timestamp": time.time(),
                "api_version": "v0-2",
            },
            "error": {
                "code": "E_INTERNAL",
                "message": f"unhandled exception: {type(exc).__name__}",
            },
        },
    )


# === Routers ===

# Health (no prefix)
app.include_router(health_router)

# Ingest (6 endpoint, prefix /api/v0-2)
app.include_router(ingest_router)

# Curate (5 endpoint, prefix /api/v0-2)
app.include_router(curate_router)

# Query (5 endpoint, prefix /api/v0-2, includes /bundles/{}/index.md + /viz.html)
app.include_router(query_router)

# Graph (4 endpoint, prefix /api/v0-2)
app.include_router(graph_router)

# Lifecycle (2 endpoint, prefix /api/v0-2)
app.include_router(lifecycle_router)

# Audit (4 endpoint, prefix /api/v0-2)
app.include_router(audit_router)

# Monitoring (2 endpoint, includes /metrics)
app.include_router(monitoring_router)


# === Entry point ===

def main() -> None:
    """CLI entry point: uvicorn backend_knowledge.main:app."""
    import uvicorn
    uvicorn.run(
        "backend_knowledge.main:app",
        host="0.0.0.0",
        port=8000,
        reload=False,
        log_level=settings.log_level.lower(),
    )


if __name__ == "__main__":
    main()
