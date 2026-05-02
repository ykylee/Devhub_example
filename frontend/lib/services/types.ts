export type UserRole = "Developer" | "Manager" | "System Admin";

export interface Metric {
  label: string;
  value: string;
  trend: string;
  color: string;
}

export interface Risk {
  title: string;
  reason: string;
  impact: string;
  status: string;
  owner: string;
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
}
