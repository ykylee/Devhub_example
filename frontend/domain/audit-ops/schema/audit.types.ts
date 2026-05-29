// audit.types.ts — wire shape for GET /api/v1/audit-logs.
// Source of truth: backend-core/internal/httpapi/audit.go (auditLogResponse +
// listAuditLogs handler). Field names follow the JSON tags emitted there.
//
// The earlier draft assumed names like `id`/`event_id`/`occurred_at`/`actor_id`
// that did not exist on the backend; consumers must use these field names so
// the JSON envelope round-trips correctly.

export interface AuditLogEntry {
  audit_id: string;
  actor_login: string;
  action: string;
  target_type: string;
  target_id: string;
  command_id?: string;
  payload: Record<string, unknown>;
  source_ip?: string;
  request_id?: string;
  source_type?: string;
  created_at: string; // ISO-8601 UTC
}

// AuditLogFilters — query parameters supported by listAuditLogs.
// Only the fields below are wired on the backend; `since`/`until` are not
// supported yet (track with the backend handler before adding them here).
export interface AuditLogFilters {
  actor_login?: string;
  action?: string;
  target_type?: string;
  target_id?: string;
  command_id?: string;
  limit?: number;
  offset?: number;
}

export interface AuditLogListMeta {
  limit?: number;
  offset?: number;
  count?: number;
}
