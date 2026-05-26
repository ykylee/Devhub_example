export const legacyMockBuildLogs = [
  { id: "101", title: "Build #101 for main", time: "2m 14s", status: "Passed" },
  { id: "102", title: "Build #102 for feat/auth", time: "1m 45s", status: "Passed" },
  { id: "103", title: "Build #103 for fix/deadlock", time: "3m 10s", status: "Failed" },
];

export const legacyMockManagerVelocity = [
  { name: "Mon", quality: 45, security: 60 },
  { name: "Tue", quality: 52, security: 65 },
  { name: "Wed", quality: 48, security: 70 },
  { name: "Thu", quality: 61, security: 68 },
  { name: "Fri", quality: 55, security: 72 },
  { name: "Sat", quality: 42, security: 40 },
  { name: "Sun", quality: 38, security: 35 },
];

export const legacyMockManagerLoad = [
  { name: "YK Lee", load: 85, status: "Critical", color: "bg-destructive" },
  { name: "Alex K.", load: 45, status: "Optimal", color: "bg-success" },
  { name: "Sam J.", load: 92, status: "Overloaded", color: "bg-destructive" },
  { name: "Jordan M.", load: 60, status: "Optimal", color: "bg-success" },
];

export const legacyMockManagerDecisions = [
  { title: "Branch Protection Policy", date: "2 days ago", type: "Security" },
  { title: "gRPC IDL Specification", date: "4 days ago", type: "Architecture" },
];

export const legacyMockProjectActivity = [
  { user: "YK Lee", action: "Completed task", target: "API Authentication", time: "2h ago" },
  { user: "Alex K.", action: "Commented on", target: "UI Redesign", time: "4h ago" },
  { user: "Sam J.", action: "Added attachment", target: "Workflow Specs", time: "Yesterday" },
];

export const legacyMockProjectTasks = [
  { title: "Implement RBAC Persistence", priority: "High", status: "In Progress", due: "May 20" },
  { title: "Dashboard Responsive Audit", priority: "Medium", status: "Review", due: "May 18" },
  { title: "Legacy Cleanup", priority: "Low", status: "To Do", due: "May 25" },
];
