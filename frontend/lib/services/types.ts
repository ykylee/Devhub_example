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

export interface WSEvent<T = any> {
  schema_version: string;
  type: string;
  event_id: string;
  occurred_at: string;
  data: T;
}

export type WSEventHandler<T = any> = (event: WSEvent<T>) => void;
