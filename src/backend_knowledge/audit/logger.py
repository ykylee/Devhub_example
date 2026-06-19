"""Audit log JSON Lines writer with daily rotation (umbrella doc §3.6.6.1).

Format: var/audit/audit-YYYY-MM-DD.jsonl (one event per line).
Daily rotation by date in filename.
Retention cleanup deletes files older than AUDIT_LOG_RETENTION_DAYS (default 7).
"""

from __future__ import annotations

import json
import threading
from datetime import datetime, timezone
from pathlib import Path

from ..config import get_settings
from ..logger import get_logger
from .events import AuditEvent, AuditEventType, build_audit_event

logger = get_logger(__name__)


class AuditLogger:
    """Thread-safe JSON Lines audit logger with daily file rotation."""

    def __init__(self, base_dir: Path | None = None, retention_days: int | None = None) -> None:
        settings = get_settings()
        self.base_dir = base_dir or (settings.var_dir / "audit")
        self.base_dir.mkdir(parents=True, exist_ok=True)
        self.retention_days = retention_days or settings.audit_log_retention_days
        self._lock = threading.Lock()
        self._current_date: str | None = None

    def _file_path(self, date: datetime) -> Path:
        return self.base_dir / f"audit-{date.strftime('%Y-%m-%d')}.jsonl"

    def _current_file(self) -> Path:
        today = datetime.now(timezone.utc).strftime("%Y-%m-%d")
        if self._current_date != today:
            self._current_date = today
        return self.base_dir / f"audit-{today}.jsonl"

    def emit(self, event: AuditEvent) -> None:
        """Append a single audit event to today's JSON Lines file (thread-safe)."""
        line = event.model_dump_json() + "\n"
        with self._lock:
            path = self._current_file()
            try:
                with path.open("a", encoding="utf-8") as f:
                    f.write(line)
            except OSError as e:
                logger.error("audit_log_write_failed", path=str(path), error=str(e))

    def emit_simple(
        self,
        event_type: AuditEventType,
        user_id: str | None = None,
        org_id: str | None = None,
        request_id: str | None = None,
        ip: str | None = None,
        success: bool = True,
        **extra: Any,
    ) -> None:
        """Convenience method: build + emit in one call."""
        event = build_audit_event(
            event_type=event_type,
            user_id=user_id,
            org_id=org_id,
            request_id=request_id,
            ip=ip,
            success=success,
            **extra,
        )
        self.emit(event)

    def read_range(
        self,
        from_date: datetime | None = None,
        to_date: datetime | None = None,
        event_type: AuditEventType | None = None,
        user_id: str | None = None,
        limit: int = 100,
    ) -> list[dict]:
        """Read audit events from JSONL files, with optional filters.

        Returns list of event dicts (parsed JSON), newest first, capped at `limit`.
        """
        results: list[dict] = []
        files = sorted(self.base_dir.glob("audit-*.jsonl"), reverse=True)
        for path in files:
            if from_date or to_date:
                try:
                    file_date = datetime.strptime(path.stem.replace("audit-", ""), "%Y-%m-%d").replace(tzinfo=timezone.utc)
                except ValueError:
                    continue
                if from_date and file_date < from_date:
                    continue
                if to_date and file_date > to_date:
                    continue
            try:
                with path.open("r", encoding="utf-8") as f:
                    for line in f:
                        line = line.strip()
                        if not line:
                            continue
                        try:
                            entry = json.loads(line)
                        except json.JSONDecodeError:
                            continue
                        if event_type and entry.get("event") != event_type.value:
                            continue
                        if user_id and entry.get("user_id") != user_id:
                            continue
                        results.append(entry)
                        if len(results) >= limit:
                            return results
            except OSError:
                continue
        results.sort(key=lambda e: e.get("timestamp", ""), reverse=True)
        if len(results) > limit:
            results = results[:limit]
        return results

    def cleanup_old(self) -> int:
        """Delete audit log files older than retention_days. Returns number deleted."""
        from datetime import timedelta
        cutoff = datetime.now(timezone.utc) - timedelta(days=self.retention_days)
        deleted = 0
        for path in self.base_dir.glob("audit-*.jsonl"):
            try:
                file_date = datetime.strptime(path.stem.replace("audit-", ""), "%Y-%m-%d").replace(tzinfo=timezone.utc)
            except ValueError:
                continue
            if file_date < cutoff:
                try:
                    path.unlink()
                    deleted += 1
                except OSError:
                    pass
        return deleted


_audit_logger: AuditLogger | None = None


def get_audit_logger() -> AuditLogger:
    """Singleton audit logger accessor."""
    global _audit_logger
    if _audit_logger is None:
        _audit_logger = AuditLogger()
    return _audit_logger
