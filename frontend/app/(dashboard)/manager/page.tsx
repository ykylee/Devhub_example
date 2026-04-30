"use client";

import { AlertTriangle, BarChart3, CheckCircle2, Clock, GitPullRequest, Target, TrendingUp, Users } from "lucide-react";

export default function ManagerDashboard() {
  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Project Overview</h1>
          <p className="text-muted-foreground mt-1">DevHub Milestone v1.0 Progress and Risks.</p>
        </div>
        <div className="flex items-center gap-3">
          <select className="bg-card border border-border text-sm rounded-lg px-3 py-2 focus:outline-none focus:ring-1 focus:ring-ring">
            <option>Milestone v1.0</option>
            <option>Milestone v2.0</option>
          </select>
          <button className="bg-primary text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium hover:bg-primary/90 transition-colors shadow-sm">
            Generate Report
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        {[
          { label: "Overall Progress", value: "68%", icon: Target, color: "text-blue-400" },
          { label: "Active Issues", value: "24", icon: GitPullRequest, color: "text-emerald-400" },
          { label: "Team Velocity", value: "42 pts", icon: TrendingUp, color: "text-purple-400" },
          { label: "Blocked / At Risk", value: "3", icon: AlertTriangle, color: "text-rose-400" },
        ].map((stat, i) => (
          <div key={i} className="bg-card rounded-xl border border-border p-6 shadow-sm">
            <div className="flex items-center justify-between">
              <p className="text-sm font-medium text-muted-foreground">{stat.label}</p>
              <stat.icon className={`w-4 h-4 ${stat.color}`} />
            </div>
            <h3 className="text-2xl font-bold mt-2">{stat.value}</h3>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Risk Alerts */}
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-card rounded-xl border border-rose-500/20 overflow-hidden shadow-sm relative">
            <div className="absolute top-0 left-0 w-1 h-full bg-rose-500"></div>
            <div className="px-6 py-4 border-b border-border bg-rose-500/5">
              <h2 className="font-semibold flex items-center gap-2 text-rose-500">
                <AlertTriangle className="w-4 h-4" /> Action Required: Risks Detected
              </h2>
            </div>
            <div className="divide-y divide-border">
              {[
                { title: "Database Migration Script", delay: "8 days delayed", assignee: "Alex", priority: "High" },
                { title: "OAuth Integration", delay: "Pending dependency", assignee: "Sam", priority: "Medium" }
              ].map((risk, i) => (
                <div key={i} className="p-6 hover:bg-accent/50 transition-colors">
                  <div className="flex items-start justify-between">
                    <div>
                      <h3 className="font-medium">{risk.title}</h3>
                      <p className="text-sm text-rose-400/80 mt-1 flex items-center gap-2">
                        <Clock className="w-3 h-3" /> {risk.delay}
                      </p>
                    </div>
                    <div className="flex items-center gap-3 text-sm">
                      <span className="flex items-center gap-1 text-muted-foreground"><Users className="w-3 h-3"/> {risk.assignee}</span>
                      <span className="px-2 py-1 bg-rose-500/10 text-rose-400 rounded text-xs font-medium">{risk.priority}</span>
                      <button className="px-3 py-1 bg-accent text-accent-foreground rounded hover:bg-accent/80 transition-colors">
                        Review
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Progress Chart Placeholder */}
          <div className="bg-card rounded-xl border border-border p-6 shadow-sm h-72 flex flex-col items-center justify-center text-center">
            <BarChart3 className="w-12 h-12 text-muted-foreground/30 mb-4" />
            <p className="font-medium text-muted-foreground">Burndown Chart</p>
            <p className="text-sm text-muted-foreground/60 mt-1">Chart visualization will be implemented with Recharts.</p>
          </div>
        </div>

        {/* Team Capacity */}
        <div className="space-y-6">
          <div className="bg-card rounded-xl border border-border overflow-hidden shadow-sm">
            <div className="px-6 py-4 border-b border-border bg-card/50">
              <h2 className="font-semibold flex items-center gap-2">
                <Users className="w-4 h-4 text-blue-400" /> Team Workload
              </h2>
            </div>
            <div className="p-6 space-y-5">
              {[
                { name: "YK Lee", load: 85, tasks: 4 },
                { name: "Alex K", load: 40, tasks: 2 },
                { name: "Sam J", load: 95, tasks: 6 }
              ].map((member, i) => (
                <div key={i}>
                  <div className="flex justify-between text-sm mb-2">
                    <span className="font-medium">{member.name}</span>
                    <span className="text-muted-foreground">{member.tasks} tasks</span>
                  </div>
                  <div className="h-2 bg-background rounded-full overflow-hidden border border-border">
                    <div 
                      className={`h-full rounded-full ${member.load > 90 ? 'bg-rose-500' : 'bg-blue-500'}`} 
                      style={{ width: `${member.load}%` }}
                    ></div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
