export type UserRole = "Developer" | "Manager" | "System Admin";
export type ApiUserRole = "developer" | "manager" | "system_admin";

export interface Metric {
  label: string;
  value: string;
  trend: string;
  color: string;
}

export interface ApiResponse<T> {
  status: string;
  data: T;
}

export interface ApiMetric {
  label: string;
  value: string;
  trend: string;
  trend_direction: "up" | "down" | "flat";
  numeric_value?: number;
  unit?: string;
}

export interface Risk {
  id?: string;
  title: string;
  reason: string;
  impact: string;
  status: string;
  owner?: string;
  owner_login?: string;
  suggested_actions?: string[];
  created_at?: string;
}

export interface BuildLog {
  id: string;
  repo: string;
  status: "success" | "failed" | "running";
  time: string;
  duration: string;
}

export interface ServiceNode {
  id: string;
  label: string;
  status: "stable" | "warning" | "down";
  cpu: string;
  memory: string;
  cpu_percent?: number;
  memory_bytes?: number;
  kind?: string;
  region?: string;
  updated_at?: string;
}

export interface ServiceEdge {
  id: string;
  source_id: string;
  target_id: string;
  label: string;
  status: "stable" | "warning" | "down";
  latency_ms?: number;
  updated_at?: string;
}

export interface ServiceActionCommand {
  command_id: string;
  command_status: string;
  requires_approval: boolean;
  audit_log_id: string;
  idempotent_replay: boolean;
  created_at: string;
}

export interface WSEvent<T = unknown> {
  schema_version: string;
  type: string;
  event_id: string;
  occurred_at: string;
  data: T;
}

export type WSEventHandler<T = unknown> = (event: WSEvent<T>) => void;

export type RbacPermission = "none" | "read" | "write" | "admin";

export interface RbacRole {
  role: ApiUserRole;
  label: UserRole;
  description: string;
}

export interface RbacResource {
  resource: string;
  label: string;
  description: string;
}

export interface RbacPermissionLevel {
  permission: RbacPermission;
  label: string;
  rank: number;
  description: string;
}

export interface RbacPolicy {
  roles: RbacRole[];
  resources: RbacResource[];
  permissions: RbacPermissionLevel[];
  matrix: Record<string, Record<string, RbacPermission>>;
}
