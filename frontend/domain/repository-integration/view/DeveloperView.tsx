"use client";

import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { 
  Activity, 
  GitPullRequest, 
  ShieldAlert, 
  FileCode2, 
  CheckCircle, 
  XCircle, 
  Terminal,
  ExternalLink,
  ChevronRight,
  GitBranch,
  ShieldCheck,
  AlertTriangle
} from "lucide-react";
import { 
  repositoryService, 
  Repository, 
  RepositoryActivity, 
  RepositoryBuildRun, 
  RepositoryDashboardData 
} from "@/domain/repository-integration/service/repository.service";
import { Badge } from "@/shared/ui-foundation/components/Badge";
import { BuildLogModal } from "./BuildLogModal";

interface DeveloperViewProps {
  repo: Repository;
  activity: RepositoryActivity | null;
  dashboardData: RepositoryDashboardData;
}

export function DeveloperView({ repo, activity, dashboardData }: DeveloperViewProps) {
  const [buildRuns, setBuildRuns] = useState<RepositoryBuildRun[]>([]);
  const [loadingBuilds, setLoadingBuilds] = useState(true);
  const [activeLogRun, setActiveLogRun] = useState<{ id: number; externalId: string } | null>(null);

  useEffect(() => {
    const fetchBuilds = async () => {
      try {
        setLoadingBuilds(true);
        const data = await repositoryService.getRepositoryBuildRuns(repo.id, { limit: 10 });
        setBuildRuns(data);
      } catch (err) {
        console.error("Failed to load build runs", err);
      } finally {
        setLoadingBuilds(false);
      }
    };
    void fetchBuilds();
  }, [repo.id]);

  const formatDuration = (seconds?: number | null) => {
    if (!seconds) return "-";
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return mins > 0 ? `${mins}m ${secs}s` : `${secs}s`;
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "success": return "text-success bg-success/10 border-success/20";
      case "failed": return "text-destructive bg-destructive/10 border-destructive/20";
      case "running": return "text-primary bg-primary/10 border-primary/20";
      case "queued": return "text-warning bg-warning/10 border-warning/20";
      default: return "text-muted-foreground bg-muted/20 border-border/80";
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "success": return CheckCircle;
      case "failed": return XCircle;
      case "running": return Activity;
      default: return Terminal;
    }
  };

  // Mock Active PR Stream Data
  const activePRs = [
    { id: 101, title: "feat: add oauth2 device authorization flow support", branch: "feat/device-auth", author: "alice", reviews: "Approved", state: "open", age: "2 hours ago" },
    { id: 104, title: "fix(core): resolve concurrency deadlocks in database operations", branch: "fix/deadlocks", author: "charlie", reviews: "Changes Requested", state: "open", age: "1 day ago" }
  ];

  return (
    <div className="space-y-8">
      {/* 2-Column Dashboard Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        
        {/* Left Column (2-Span): Build Runs & PRs */}
        <div className="lg:col-span-2 space-y-8">
          
          {/* Build History console stream */}
          <section className="glass-card p-6">
            <div className="flex items-center justify-between mb-6">
              <div>
                <h3 className="text-lg font-bold text-foreground dark:text-primary-foreground flex items-center gap-2">
                  <Activity className="w-5 h-5 text-primary" /> Build & Integration Runs
                </h3>
                <p className="text-xs text-muted-foreground">Recent CI pipelines from integrated SCM triggers.</p>
              </div>
              <Badge variant="glass">{buildRuns.length} runs loaded</Badge>
            </div>

            {loadingBuilds ? (
              <div className="space-y-4">
                {[...Array(3)].map((_, idx) => (
                  <div key={idx} className="h-16 bg-muted/10 animate-pulse rounded-2xl border border-border/60" />
                ))}
              </div>
            ) : buildRuns.length === 0 ? (
              <div className="h-32 flex flex-col items-center justify-center border border-dashed border-border/80 rounded-2xl bg-muted/5">
                <Terminal className="w-8 h-8 text-muted-foreground/60 mb-2 animate-bounce" />
                <p className="text-xs text-muted-foreground font-bold">No CI pipelines recorded for this repository.</p>
              </div>
            ) : (
              <div className="divide-y divide-border/40">
                {buildRuns.map((run) => {
                  const Icon = getStatusIcon(run.status);
                  return (
                    <div key={run.id} className="py-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 group">
                      <div className="flex items-center gap-3">
                        <div className={`p-2 rounded-xl border ${getStatusColor(run.status)}`}>
                          <Icon className="w-4 h-4" />
                        </div>
                        <div>
                          <div className="flex items-center gap-2 mb-1">
                            <span className="text-xs font-mono font-bold text-foreground dark:text-primary-foreground">
                              #{run.id}
                            </span>
                            <span className="text-xs font-semibold text-muted-foreground">
                              {run.commit_sha.substring(0, 7)}
                            </span>
                            <Badge variant="glass" className="text-[10px] py-0 px-2 font-mono flex items-center gap-1">
                              <GitBranch className="w-2.5 h-2.5" /> {run.branch}
                            </Badge>
                          </div>
                          <p className="text-[10px] text-muted-foreground">
                            Started: {new Date(run.started_at).toLocaleString()}
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center justify-between sm:justify-end gap-6 text-right">
                        <div>
                          <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-0.5">Duration</p>
                          <p className="text-xs font-mono font-bold text-foreground dark:text-primary-foreground">{formatDuration(run.duration_seconds)}</p>
                        </div>
                        <div className="flex gap-2">
                          {run.status === "failed" && (
                            <button
                              onClick={() => setActiveLogRun({ id: run.id, externalId: run.run_external_id })}
                              className="px-3 py-1.5 rounded-lg bg-destructive/15 border border-destructive/30 hover:bg-destructive/25 text-[10px] font-black uppercase text-destructive tracking-widest transition-all"
                            >
                              View Logs
                            </button>
                          )}
                          <a
                            href={repo.html_url + "/actions/runs/" + run.run_external_id}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="p-1.5 rounded-lg hover:bg-muted/30 border border-transparent hover:border-border transition-all text-muted-foreground hover:text-primary"
                          >
                            <ExternalLink className="w-4 h-4" />
                          </a>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </section>

          {/* Active Pull Requests */}
          <section className="glass-card p-6">
            <div className="flex items-center justify-between mb-6">
              <div>
                <h3 className="text-lg font-bold text-foreground dark:text-primary-foreground flex items-center gap-2">
                  <GitPullRequest className="w-5 h-5 text-primary" /> Active Pull Requests
                </h3>
                <p className="text-xs text-muted-foreground">Active code reviews and code integration streams.</p>
              </div>
              <Badge variant="glass">{activePRs.length} active PRs</Badge>
            </div>

            <div className="space-y-4">
              {activePRs.map((pr) => (
                <div key={pr.id} className="p-4 border border-border/80 bg-muted/10 rounded-2xl hover:bg-muted/20 transition-all flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-black text-primary">#{pr.id}</span>
                      <h4 className="text-xs font-bold text-foreground dark:text-primary-foreground hover:underline cursor-pointer">{pr.title}</h4>
                    </div>
                    <div className="flex items-center gap-3 text-[10px] text-muted-foreground font-semibold">
                      <span>By @{pr.author}</span>
                      <span>•</span>
                      <span className="flex items-center gap-0.5"><GitBranch className="w-3 h-3" /> {pr.branch}</span>
                      <span>•</span>
                      <span>{pr.age}</span>
                    </div>
                  </div>
                  <div className="flex items-center justify-between sm:justify-end gap-3">
                    <Badge 
                      variant={
                        pr.reviews === "Approved" ? "success" : 
                        pr.reviews === "Changes Requested" ? "danger" : 
                        "warning"
                      }
                      className="text-[9px] uppercase tracking-wider font-black"
                    >
                      {pr.reviews}
                    </Badge>
                    <ChevronRight className="w-4 h-4 text-muted-foreground/60" />
                  </div>
                </div>
              ))}
            </div>
          </section>

        </div>

        {/* Right Column (1-Span): Static Analysis & Security */}
        <div className="space-y-8">
          
          {/* Static Analysis (SonarQube) */}
          <section className="glass-card p-6">
            <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground mb-6 flex items-center gap-2">
              <FileCode2 className="w-4 h-4 text-primary" /> Static Analysis (SonarQube)
            </h3>
            
            <div className="space-y-6">
              {/* Quality Gate Status */}
              <div className="p-4 border border-border bg-muted/10 rounded-2xl flex items-center justify-between">
                <div>
                  <h4 className="text-xs font-bold text-foreground dark:text-primary-foreground">Quality Gate</h4>
                  <p className="text-[10px] text-muted-foreground">Standard compliance state</p>
                </div>
                <Badge 
                  variant={dashboardData.quality.quality_gate === "passed" ? "success" : "danger"}
                  className="font-black uppercase tracking-wider text-[10px] py-1 px-3"
                >
                  {dashboardData.quality.quality_gate === "passed" ? "Passed" : "Failed"}
                </Badge>
              </div>

              {/* Core Quality Metrics */}
              <div className="grid grid-cols-2 gap-4">
                <div className="p-4 border border-border bg-muted/5 rounded-2xl text-center">
                  <p className="text-[9px] font-black text-muted-foreground uppercase tracking-widest mb-1">Coverage</p>
                  <p className="text-xl font-mono font-black text-foreground dark:text-primary-foreground">{dashboardData.quality.coverage}%</p>
                  <div className="mt-2 w-full bg-muted/30 rounded-full h-1.5 overflow-hidden">
                    <div className="bg-success h-1.5 rounded-full" style={{ width: `${dashboardData.quality.coverage}%` }} />
                  </div>
                </div>
                <div className="p-4 border border-border bg-muted/5 rounded-2xl text-center">
                  <p className="text-[9px] font-black text-muted-foreground uppercase tracking-widest mb-1">Duplication</p>
                  <p className="text-xl font-mono font-black text-foreground dark:text-primary-foreground">{dashboardData.quality.duplication}%</p>
                  <div className="mt-2 w-full bg-muted/30 rounded-full h-1.5 overflow-hidden">
                    <div className="bg-warning h-1.5 rounded-full" style={{ width: `${dashboardData.quality.duplication * 10}%` }} />
                  </div>
                </div>
              </div>

              {/* Issue Breakdown */}
              <div className="space-y-3">
                <h4 className="text-[10px] font-black text-muted-foreground uppercase tracking-wider">Unresolved Issues</h4>
                {[
                  { label: "Blocker", count: dashboardData.quality.issues.blocker, color: "bg-destructive text-destructive-foreground" },
                  { label: "Critical", count: dashboardData.quality.issues.critical, color: "bg-warning text-warning-foreground" },
                  { label: "Major", count: dashboardData.quality.issues.major, color: "bg-primary text-primary-foreground" }
                ].map((item) => (
                  <div key={item.label} className="flex items-center justify-between text-xs py-1">
                    <span className="font-semibold text-muted-foreground">{item.label}</span>
                    <span className={`px-2 py-0.5 rounded-md font-mono font-black text-[10px] ${item.color}`}>
                      {item.count}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </section>

          {/* Security & Vulnerability Scan */}
          <section className="glass-card p-6 border-l-4 border-l-warning">
            <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground mb-6 flex items-center gap-2">
              <ShieldAlert className="w-4 h-4 text-warning" /> Security & Vulnerability
            </h3>

            <div className="space-y-6">
              {/* Hardcoded Secrets Detect */}
              <div className="p-4 border border-border bg-muted/10 rounded-2xl flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <AlertTriangle className={`w-4 h-4 ${dashboardData.security.secrets_detected > 0 ? "text-destructive animate-pulse" : "text-muted-foreground"}`} />
                  <div>
                    <h4 className="text-xs font-bold text-foreground dark:text-primary-foreground">Hardcoded Secrets</h4>
                    <p className="text-[10px] text-muted-foreground">Tokens, passwords in commit diffs</p>
                  </div>
                </div>
                <Badge variant={dashboardData.security.secrets_detected > 0 ? "danger" : "glass"} className="font-black font-mono">
                  {dashboardData.security.secrets_detected}
                </Badge>
              </div>

              {/* Package Dependency Vulnerabilities */}
              <div className="space-y-3">
                <h4 className="text-[10px] font-black text-muted-foreground uppercase tracking-wider">Dependency Vulnerabilities</h4>
                {[
                  { label: "High Risk", count: dashboardData.security.vulnerabilities.high, color: "text-destructive" },
                  { label: "Medium Risk", count: dashboardData.security.vulnerabilities.medium, color: "text-warning" },
                  { label: "Low Risk", count: dashboardData.security.vulnerabilities.low, color: "text-muted-foreground" }
                ].map((item) => (
                  <div key={item.label} className="flex items-center justify-between text-xs py-1 font-semibold">
                    <span className="text-muted-foreground">{item.label}</span>
                    <span className={`font-mono font-black ${item.color}`}>{item.count}</span>
                  </div>
                ))}
              </div>

              {/* Security Shield badge */}
              <div className="pt-2 flex items-center gap-2 text-[10px] text-muted-foreground font-semibold">
                <ShieldCheck className="w-4 h-4 text-success" />
                <span>Last full scanner sync: 1 hour ago</span>
              </div>
            </div>
          </section>

        </div>

      </div>

      {/* AnimatePresence for BuildLogModal */}
      <AnimatePresence>
        {activeLogRun && (
          <BuildLogModal
            repositoryId={repo.id}
            runId={activeLogRun.id}
            runExternalId={activeLogRun.externalId}
            onClose={() => setActiveLogRun(null)}
          />
        )}
      </AnimatePresence>
    </div>
  );
}
