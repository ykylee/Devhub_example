"""Audit event types and schema (umbrella doc §3.6.6.1).

7 audit event type enum + base event model. Each event shares a common shape
(timestamp, user_id, request_id, ip) plus event-specific fields.
"""

from __future__ import annotations

from datetime import datetime, timezone
from enum import Enum
from typing import Any, Literal

from pydantic import BaseModel, ConfigDict, Field


class AuditEventType(str, Enum):
    """7 audit event types per §3.6.6.1."""

    USER_LOGIN = "audit.user.login"
    CONCEPT_ACCESS = "audit.concept.access"
    CURATION_EDIT = "audit.curation.edit"
    QUERY = "audit.query"
    CONCEPT_ARCHIVE = "audit.concept.archive"
    CONCEPT_PUBLISH = "audit.concept.publish"
    CONFIG_CHANGE = "audit.config.change"


class AuditEvent(BaseModel):
    """Base audit event (common fields, all 7 event types inherit this shape)."""

    model_config = ConfigDict(extra="allow")

    event: str
    timestamp: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))
    user_id: str | None = None
    org_id: str | None = None
    request_id: str | None = None
    ip: str | None = None
    success: bool = True
    # event-specific fields stored as extra="allow" (dict passthrough)


AuditSeverity = Literal["info", "warning", "critical"]


def build_audit_event(
    event_type: AuditEventType,
    user_id: str | None = None,
    org_id: str | None = None,
    request_id: str | None = None,
    ip: str | None = None,
    success: bool = True,
    **extra: Any,
) -> AuditEvent:
    """Construct an audit event with common fields + event-specific extras."""
    payload: dict[str, Any] = {
        "event": event_type.value,
        "user_id": user_id,
        "org_id": org_id,
        "request_id": request_id,
        "ip": ip,
        "success": success,
    }
    payload.update(extra)
    return AuditEvent(**payload)
