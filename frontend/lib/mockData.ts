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
        { label: "Build Success", value: "98%", trend: "+2%", color: "text-emerald-500" },
        { label: "Code Review", value: "2", trend: "Pending", color: "text-amber-500" },
      ];
    case "Manager":
      return [
        { label: "Completion", value: "72%", trend: "+5%", color: "text-indigo-500" },
        { label: "Team Velocity", value: "48", trend: "+12%", color: "text-purple-500" },
        { label: "Open Risks", value: "2", trend: "High", color: "text-rose-500" },
        { label: "Avg Cycle Time", value: "4.2d", trend: "-0.5d", color: "text-emerald-500" },
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
  { name: "Mon", velocity: 45, load: 60 },
  { name: "Tue", velocity: 52, load: 65 },
  { name: "Wed", velocity: 48, load: 70 },
  { name: "Thu", velocity: 61, load: 68 },
  { name: "Fri", velocity: 55, load: 72 },
  { name: "Sat", velocity: 42, load: 40 },
  { name: "Sun", velocity: 38, load: 35 },
];
