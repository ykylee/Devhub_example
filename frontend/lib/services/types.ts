/**
 * UI domain types — shapes the React layer renders.
 *
 * Convention (PR-B2, work_26_05_11-b):
 *   - UI labels (e.g. UserRole "Developer"|"Manager"|"System Admin") live here
 *   - service modules convert wire shapes (./wire) to these on the way in
 *   - UI components import from this file, NOT from ./wire
 *   - For backwards compatibility this file re-exports the envelope types
 *     (ApiResponse, ApiUserRole, ApiMetric, WSEvent, WSEventHandler) so
 *     existing service code can keep its single import. New service code is
 *     encouraged to import wire shapes from "./wire" directly to keep
 *     intent clear.
 */

export type {
  ApiResponse,
  ApiUserRole,
  ApiMetric,
  WSEvent,
  WSEventHandler,
} from "./wire";

export type UserRole = "Developer" | "Manager" | "System Admin";

export interface Metric {
  label: string;
  value: string;
  trend: string;
  color: string;
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

export type RbacPermission = "none" | "read" | "write" | "admin";

import type { ApiUserRole } from "./wire";

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
