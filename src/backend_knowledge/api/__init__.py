"""API routers (umbrella doc §3.1 API 매트릭스 정합).

24 endpoint = Ingest 6 + Curate 5 + Query 5 + Graph 4 + Audit 4.

PR 1: Ingest 6 endpoint (FR-I-001 ~ FR-I-006).
PR 2 (이 파일): Curate 5 + Query 5 + Graph 4 + viz.html.
PR 3: Audit + Monitoring + Operational.
"""

from .ingest import router as ingest_router
from .curate import router as curate_router
from .query import router as query_router
from .graph import router as graph_router

__all__ = ["ingest_router", "curate_router", "query_router", "graph_router"]
