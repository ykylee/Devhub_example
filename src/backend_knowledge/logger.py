"""Logger (structlog).

umbrella doc §11.3 monitoring 정공법 + §3.6.6 audit log 분리:
- 본 logger: application log (var/log/backend-knowledge.jsonl)
- audit log: 별도 모듈 (§3.6.6.1, var/audit/audit-YYYY-MM-DD.jsonl)
"""

from __future__ import annotations

import logging
import sys

import structlog

from .config import get_settings


def configure_logging() -> None:
    """structlog 설정. main.py 시작 시 1회 호출."""
    settings = get_settings()
    level_name = settings.log_level.upper()
    log_format = settings.log_format.lower()

    # logging module standard level (INFO=20, DEBUG=10, WARNING=30, ERROR=40)
    level_int = getattr(logging, level_name, logging.INFO)

    if log_format == "json":
        processors = [
            structlog.contextvars.merge_contextvars,
            structlog.processors.add_log_level,
            structlog.processors.TimeStamper(fmt="iso", utc=True),
            structlog.processors.StackInfoRenderer(),
            structlog.processors.format_exc_info,
            structlog.processors.JSONRenderer(),
        ]
    else:
        processors = [
            structlog.contextvars.merge_contextvars,
            structlog.processors.add_log_level,
            structlog.processors.TimeStamper(fmt="%Y-%m-%d %H:%M:%S", utc=False),
            structlog.dev.ConsoleRenderer(colors=True),
        ]

    structlog.configure(
        processors=processors,
        wrapper_class=structlog.make_filtering_bound_logger(level_int),
        context_class=dict,
        logger_factory=structlog.PrintLoggerFactory(file=sys.stdout),
        cache_logger_on_first_use=True,
    )


def get_logger(name: str | None = None) -> structlog.stdlib.BoundLogger:
    """structlog logger accessor."""
    if name:
        return structlog.get_logger(name)
    return structlog.get_logger("backend_knowledge")
