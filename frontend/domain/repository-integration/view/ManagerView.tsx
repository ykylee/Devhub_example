"use client";

import { useState } from "react";
import { 
  BarChart, 
  Bar, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer, 
  PieChart, 
  Pie, 
  Cell, 
  AreaChart, 
  Area 
} from "recharts";
import { 
  Users, 
  Layers, 
  Clock, 
  Eye, 
  EyeOff, 
  CheckCircle, 
  GitFork, 
  AlertTriangle, 
  Cpu, 
import { Repository, RepositoryActivity, RepositoryDashboardData } from "@/domain/repository-integration/service/repository.service";
import { Badge } from "@/shared/ui-foundation/components/Badge";

import { RepositoryKPISection } from "./RepositoryKPISection";
import { RepositoryTestsSection } from "./RepositoryTestsSection";
interface ManagerViewProps {
  repo: Repository;
  activity: RepositoryActivity | null;
  dashboardData: RepositoryDashboardData;
}

const COLORS = ["#3b82f6", "#10b981", "#8b5cf6", "#f59e0b", "#94a3b8"];

export function ManagerView({ repo, activity, dashboardData }: ManagerViewProps) {
  const [showContributors, setShowContributors] = useState(true);

  // Chart Data preparation
  const trendData = dashboardData.productivity.weekly_commits.map((commit, idx) => ({
    week: commit.week,
    commits: commit.count,
    prs: dashboardData.productivity.weekly_prs[idx]?.count || 0
  }));

  const contributorData = [
    { name: "alice", value: 45 },
    { name: "bob", value: 30 },
    { name: "charlie", value: 15 },
    { name: "david", value: 8 },
    { name: "Others", value: 7 }
  ];

  // System & Organization status derived
  const syncStatus = repo.provider_id ? "synced" : "unlinked";

  return (
    <div className="space-y-8">
      {/* 3-Part Manager Focus Cards (Team, Org, Sys) */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        
        {/* Team Manager Focus Card */}
        <div className="glass-card p-6 border-t-4 border-t-primary/70">
          <div className="flex items-center gap-3 mb-4">
            <div className="p-2 rounded-lg bg-primary/10 text-primary">
              <Clock className="w-5 h-5" />
            </div>
            <div>
              <h4 className="text-sm font-black text-foreground dark:text-primary-foreground">Team Manager Focus</h4>
              <p className="text-[9px] text-muted-foreground uppercase tracking-widest">Velocity & Lead Time</p>
            </div>
          </div>
          <div className="space-y-4">
            <div>
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-wider mb-1">Avg PR Lead Time</p>
              <h3 className="text-2xl font-black text-foreground dark:text-primary-foreground font-mono">
                {dashboardData.productivity.avg_pr_lead_time_hours} hrs
              </h3>
              <p className="text-[9px] text-muted-foreground mt-0.5">Time taken from PR creation to merge.</p>
            </div>
            <div className="pt-2 border-t border-border/40">
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-wider mb-1">Build Success Ratio</p>
              <h3 className="text-xl font-black text-success font-mono">
                {(activity?.build_success_rate ? activity.build_success_rate * 100 : 100).toFixed(0)}%
              </h3>
            </div>
          </div>
        </div>

        {/* Organization Admin Focus Card */}
        <div className="glass-card p-6 border-t-4 border-t-success/70">
          <div className="flex items-center gap-3 mb-4">
            <div className="p-2 rounded-lg bg-success/10 text-success">
              <Users className="w-5 h-5" />
            </div>
            <div>
              <h4 className="text-sm font-black text-foreground dark:text-primary-foreground">Organization Admin Focus</h4>
              <p className="text-[9px] text-muted-foreground uppercase tracking-widest">Collaborators & Rollup</p>
            </div>
          </div>
          <div className="space-y-4">
            <div>
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-wider mb-1">Active Contributors</p>
              <h3 className="text-2xl font-black text-foreground dark:text-primary-foreground font-mono">
                {activity?.active_contributors.length || 0} devs
              </h3>
              <p className="text-[9px] text-muted-foreground mt-0.5">Unique active code committers in window.</p>
            </div>
            <div className="pt-2 border-t border-border/40 flex items-center justify-between">
              <div>
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-wider mb-0.5">Code Governance</p>
                <Badge variant={dashboardData.quality.quality_gate === "passed" ? "success" : "danger"}>
                  Gate {dashboardData.quality.quality_gate === "passed" ? "PASS" : "FAIL"}
                </Badge>
              </div>
              <div>
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-wider mb-0.5">Review Bottleneck</p>
                <Badge variant={dashboardData.quality.issues.blocker > 0 ? "warning" : "glass"}>
                  {dashboardData.quality.issues.blocker > 0 ? "High" : "Low"}
                </Badge>
              </div>
            </div>
          </div>
        </div>

        {/* System Admin Focus Card */}
        <div className="glass-card p-6 border-t-4 border-t-purple-500/75">
          <div className="flex items-center gap-3 mb-4">
            <div className="p-2 rounded-lg bg-purple-500/10 text-purple-400">
              <Cpu className="w-5 h-5" />
            </div>
            <div>
              <h4 className="text-sm font-black text-foreground dark:text-primary-foreground">System Admin Focus</h4>
              <p className="text-[9px] text-muted-foreground uppercase tracking-widest">SCM Sync & Resource</p>
            </div>
          </div>
          <div className="space-y-4">
            <div>
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-wider mb-1">SCM Provider Integration</p>
              <Badge variant={syncStatus === "synced" ? "success" : "warning"} className="uppercase font-black font-mono">
                {syncStatus}
              </Badge>
              <p className="text-[9px] text-muted-foreground mt-1.5 font-mono">Key: {repo.provider_key || "none"}</p>
            </div>
            <div className="pt-2 border-t border-border/40 flex items-center justify-between text-xs">
              <span className="text-muted-foreground font-semibold">Visibility:</span>
              <span className="font-bold text-foreground dark:text-primary-foreground">{repo.private ? "Private" : "Public"}</span>
            </div>
          </div>
        </div>

      </div>

      {/* Main Charts area */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        
        {/* Left Area (2-Spans): Activity & Productivity Trend */}
        <section className="lg:col-span-2 glass-card p-6">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h3 className="text-base font-bold text-foreground dark:text-primary-foreground flex items-center gap-2">
                <GitFork className="w-5 h-5 text-primary" /> Repository Activity Trend
              </h3>
              <p className="text-xs text-muted-foreground">Weekly code changes and PR integration frequency.</p>
            </div>
          </div>
          
          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={trendData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                <defs>
                  <linearGradient id="colorCommits" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.2}/>
                    <stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/>
                  </linearGradient>
                  <linearGradient id="colorPRs" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#10b981" stopOpacity={0.2}/>
                    <stop offset="95%" stopColor="#10b981" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" className="stroke-border/40" />
                <XAxis dataKey="week" stroke="#94a3b8" fontSize={10} tickLine={false} />
                <YAxis stroke="#94a3b8" fontSize={10} tickLine={false} />
                <Tooltip 
                  contentStyle={{ backgroundColor: "rgba(15, 23, 42, 0.9)", border: "1px solid rgba(255, 255, 255, 0.1)", borderRadius: "12px" }}
                  labelStyle={{ color: "#94a3b8", fontSize: "10px", fontWeight: "bold" }}
                />
                <Area type="monotone" dataKey="commits" name="Commits" stroke="#3b82f6" strokeWidth={2} fillOpacity={1} fill="url(#colorCommits)" />
                <Area type="monotone" dataKey="prs" name="Merged PRs" stroke="#10b981" strokeWidth={2} fillOpacity={1} fill="url(#colorPRs)" />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </section>

        {/* Right Area (1-Span): Contributor Distribution (Toggleable) */}
        <section className="glass-card p-6 flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground flex items-center gap-2">
                <Users className="w-4 h-4 text-primary" /> Contributor Distribution
              </h3>
              <button
                onClick={() => setShowContributors(!showContributors)}
                className="p-1.5 hover:bg-muted/30 rounded-lg text-muted-foreground transition-all"
                title={showContributors ? "Hide distribution chart" : "Show distribution chart"}
              >
                {showContributors ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>

            {showContributors ? (
              <div className="space-y-6">
                <div className="h-40 w-full flex items-center justify-center">
                  <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                      <Pie
                        data={contributorData}
                        cx="50%"
                        cy="50%"
                        innerRadius={45}
                        outerRadius={65}
                        paddingAngle={3}
                        dataKey="value"
                      >
                        {contributorData.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip
                        contentStyle={{ backgroundColor: "rgba(15, 23, 42, 0.9)", border: "none", borderRadius: "8px" }}
                      />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
                <div className="grid grid-cols-2 gap-2 text-[10px]">
                  {contributorData.map((item, idx) => (
                    <div key={item.name} className="flex items-center gap-2 font-bold">
                      <span className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: COLORS[idx % COLORS.length] }} />
                      <span className="text-muted-foreground">{item.name}</span>
                      <span className="text-foreground dark:text-primary-foreground font-mono">{item.value}%</span>
                    </div>
                  ))}
                </div>
              </div>
            ) : (
              <div className="h-56 flex flex-col items-center justify-center bg-muted/5 border border-dashed border-border/80 rounded-2xl p-4 text-center">
                <EyeOff className="w-8 h-8 text-muted-foreground/60 mb-2 animate-pulse" />
                <p className="text-xs text-muted-foreground font-bold mb-1">Contributor chart is hidden</p>
                <button
                  onClick={() => setShowContributors(true)}
                  className="mt-2 px-3 py-1 bg-primary text-primary-foreground text-[10px] font-black uppercase tracking-wider rounded-lg hover:scale-105 active:scale-95 transition-all"
                >
                  Unhide
                </button>
              </div>
            )}
          </div>
        </section>

      </div>

      {/* Down Area: Platform/Project Linkages impact */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        
        {/* Project Linkage Details */}
        <section className="lg:col-span-2 glass-card p-6">
          <h3 className="text-base font-bold text-foreground dark:text-primary-foreground flex items-center gap-2 mb-6">
            <Layers className="w-5 h-5 text-primary" /> Integrated Platform & Project Links
          </h3>

          <div className="space-y-4">
            {/* Platforms Link */}
            <div>
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2">Linked Platforms</p>
              {dashboardData.linkage.linked_platforms.map((platform) => (
                <div key={platform.id} className="p-3 border border-border/60 bg-muted/5 rounded-xl flex items-center justify-between text-xs">
                  <span className="font-bold text-foreground dark:text-primary-foreground">{platform.name}</span>
                  <Badge variant="success" className="uppercase font-black text-[9px] tracking-wider">{platform.status}</Badge>
                </div>
              ))}
            </div>

            {/* Projects Link */}
            <div>
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2">Linked Projects</p>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {dashboardData.linkage.linked_projects.map((project) => (
                  <div key={project.id} className="p-3 border border-border/60 bg-muted/5 rounded-xl flex items-center justify-between text-xs">
                    <span className="font-bold text-foreground dark:text-primary-foreground">{project.name}</span>
                    <Badge variant={project.status === "active" ? "success" : "glass"} className="uppercase font-black text-[9px] tracking-wider">
                      {project.status}
                    </Badge>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </section>

        {/* SCM Sync Errors / Alerts log */}
        <section className="glass-card p-6 border-l-4 border-l-destructive">
          <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground mb-6 flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 text-destructive" /> SCM Connection Log
          </h3>
          <div className="space-y-4 text-xs">
            <div className="p-3 bg-muted/10 border border-border rounded-xl">
              <p className="text-[9px] font-black text-muted-foreground uppercase tracking-wider mb-1">Status</p>
              <div className="flex items-center gap-2 font-bold text-foreground dark:text-primary-foreground">
                <CheckCircle className="w-4 h-4 text-success" /> Synced & Active
              </div>
            </div>
            <div className="space-y-2">
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-wider">Operational Notes</p>
              <p className="text-muted-foreground text-[10px] leading-relaxed">
                SCM sync is running on default 30-minute interval pull. Push webhooks are configured and verified.
              </p>
            </div>
          </div>
        </section>

        {/* Sprint A — Repository KPI / Tests sub-section (kpi-tests-per-domain-scope.md §2.1) */}
        <RepositoryKPISection repoId={repo.id} />
        <RepositoryTestsSection repoId={repo.id} />

      </div>
    </div>

  );
}
