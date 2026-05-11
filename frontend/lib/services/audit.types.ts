export interface AuditLogEntry {
  id: string;
  event_id: string;
  action: string;
  target_type: string;
  target_id: string;
  actor_id: string;
  actor_login: string;
  occurred_at: string;
  payload: Record<string, unknown>;
  request_id?: string;
}

export interface AuditLogFilters {
  action?: string;
  target_type?: string;
  actor_id?: string;
  since?: string;
  until?: string;
  limit?: number;
  offset?: number;
}
