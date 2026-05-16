import { UserRole } from "./store";

export interface SystemNode {
  id: string;
  label: string;
  status: "Healthy" | "Warning" | "Down";
  load: number;
}

export interface Metric {
  label: string;
  value: string;
  trend: string;
  color: string;
}

export const getMockMetrics = (role: UserRole): Metric[] => {
  switch (role) {
    case "Developer":
      return [
        { label: "Active Tasks", value: "3", trend: "On Track", color: "text-blue-500" },
        { label: "Open PRs", value: "2", trend: "Reviewing", color: "text-indigo-500" },
        { label: "Weekly Velocity", value: "42 pts", trend: "+15%", color: "text-emerald-500" },
        { label: "Assigned DREQs", value: "5", trend: "+1 new", color: "text-purple-500" },
      ];
    case "Manager":
      return [
        { label: "SLA Compliance", value: "96%", trend: "+1.2%", color: "text-emerald-500" },
        { label: "Critical Vulnerabilities", value: "0", trend: "Stable", color: "text-green-500" },
        { label: "Code Coverage", value: "84.2%", trend: "+2.1%", color: "text-indigo-500" },
        { label: "Security Score", value: "A+", trend: "Top Tier", color: "text-purple-500" },
      ];
    case "System Admin":
      return [
        { label: "Availability", value: "99.99%", trend: "Stable", color: "text-emerald-500" },
        { label: "Active Runners", value: "12/12", trend: "Full", color: "text-primary" },
        { label: "AI Engine Load", value: "24%", trend: "Low", color: "text-purple-500" },
        { label: "Storage", value: "1.2TB", trend: "82%", color: "text-amber-500" },
      ];
  }
};

export const mockBuildLogs = [
  { id: "101", title: "Build #101 for main", time: "2m 14s", status: "Passed" },
  { id: "102", title: "Build #102 for feat/auth", time: "1m 45s", status: "Passed" },
  { id: "103", title: "Build #103 for fix/deadlock", time: "3m 10s", status: "Failed" },
];

export const mockVelocityData = [
  { name: "Mon", quality: 45, security: 60 },
  { name: "Tue", quality: 52, security: 65 },
  { name: "Wed", quality: 48, security: 70 },
  { name: "Thu", quality: 61, security: 68 },
  { name: "Fri", quality: 55, security: 72 },
  { name: "Sat", quality: 42, security: 40 },
  { name: "Sun", quality: 38, security: 35 },
];
