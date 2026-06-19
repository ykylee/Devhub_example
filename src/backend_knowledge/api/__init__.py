"""API routers (umbrella doc §3.1 API 매트릭스 정합).

20 endpoint = Ingest 6 + Curate 5 + Query 5 + Graph 4.

PR 1 (이 파일): Ingest 6 endpoint (FR-I-001 ~ FR-I-006).
PR 2: Curate 5 + Query 5 + Graph 4 + viz.html.
PR 3: Audit + Monitoring + Operational.
"""

from .ingest import router as ingest_router

__all__ = ["ingest_router"]
