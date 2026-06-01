"use client";

import { useCallback, useEffect, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  Activity,
  ArrowLeft,
  Globe,
  ShieldCheck,
  Zap,
  Clock,
  RefreshCcw,
  Settings,
  GitBranch,
  AlertTriangle,
  Play,
  Briefcase,
  Layers,
  Sparkles,
  Rocket
} from "lucide-react";
import { useRouter, useParams } from "next/navigation";
import { Badge } from "@/shared/ui-foundation/components/Badge";
import { cn } from "@/shared/utils";
import { useStore } from "@/lib/store";
import { applicationService, ApplicationDashboard, Application } from "@/domain/application-lifecycle/service/application.service";
import { ApplicationRepository } from "@/domain/application-lifecycle/schema/project.types";
import { projectService } from "@/domain/application-lifecycle/service/project.service";
import type { ApplicationRepository } from "@/domain/application-lifecycle/schema/project.types";
import { ApplicationCreationModal } from "@/domain/application-lifecycle/view/ApplicationCreationModal";
import { useToast } from "@/shared/ui-foundation/components/Toast";
import { toUserErrorMessage } from "@/shared/utils/error-message";
import { lifecycleStatusBadgeVariant } from "@/shared/utils/lifecycle-status";
import { applicationBuildStatusView } from "@/shared/utils/last-build";
import { PageError, PageLoading } from "@/shared/ui-foundation/components/PageState";
import { apiClient } from "@/shared/api/api-client";
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip
} from "recharts";

export default function ApplicationDetailPage() {
  const params = useParams();
  const id = params.id as string;
  const router = useRouter();
  const actor = useStore((s) => s.actor);
  const { toast } = useToast();

  const [dashboard, setDashboard] = useState<ApplicationDashboard | null>(null);
  const [repositories, setRepositories] = useState<ApplicationRepository[]>([]);
  const [application, setApplication] = useState<Application | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Edit Modal state
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);

  // Promote Modal states
  const [isPromoteOpen, setIsPromoteOpen] = useState(false);
  const [selectedDreq, setSelectedDreq] = useState<{ dreq_id: string; title: string } | null>(null);
  const [projectKey, setProjectKey] = useState("");
  const [projectName, setProjectName] = useState("");
  const [projectLeader, setProjectLeader] = useState("");
  const [promoting, setPromoting] = useState(false);

  const loadData = useCallback(async () => {
    try {
      setError(null);
      setLoading(true);
      const [dashData, reposData, appData] = await Promise.all([
        applicationService.getApplicationDashboard(id),
        projectService.getApplicationRepositories(id),
        applicationService.getApplication(id),
      ]);
      setDashboard(dashData);
      setRepositories(reposData);
      setApplication(appData);
    } catch (err) {
      setError(toUserErrorMessage(err, "Failed to load application details."));
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    const timer = setTimeout(() => {
      void loadData();
    }, 0);
    return () => clearTimeout(timer);
  }, [loadData]);

  const handlePromoteClick = (dreq_id: string, title: string) => {
    setSelectedDreq({ dreq_id, title });
    setProjectName(title);
    setProjectKey("PROJ-" + dreq_id.slice(-4).toUpperCase());
    // Leader 입력을 현재 사용자(actor)의 user_id 로 prefill (수정 가능) — 사용자가 매번
    // 타이핑할 필요 없이 본인 promote 시 즉시 진행 가능 (codex 3a 후속).
    setProjectLeader(actor?.user_id || actor?.login || "");
    setIsPromoteOpen(true);
  };

  const handlePromoteSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedDreq) return;

    // 검증 1: 본 application 에 연결된 repository 가 있어야 promote 가능.
    // Backend 가 (repo_provider, repo_full_name) → repository_id resolve 지원 (P1 fix).
    const primaryRepo = repositories.find((r) => r.role === "primary") ?? repositories[0];
    if (!primaryRepo) {
      alert("Cannot promote: this application has no linked repository. Please link a repository first.");
      return;
    }

    // 검증 2: 폼의 Leader User ID 입력값을 owner_user_id 로 사용. 빈 값이면 현재
    // 사용자(actor)로 fallback (handlePromoteClick 에서 prefill 한 값과 동일).
    const ownerUserID = projectLeader.trim() || actor?.user_id || actor?.login;
    if (!ownerUserID) {
      alert("Cannot promote: leader user id is empty and current user identity is missing.");
      return;
    }

    try {
      setPromoting(true);
      // Promote DREQ API — backend resolves repository_id from (repo_provider, repo_full_name)
      await apiClient("POST", `/api/v1/dev-requests/${selectedDreq.dreq_id}/register`, {
        target_type: "project",
        project_payload: {
          application_id: id,
          repo_provider: primaryRepo.repo_provider,
          repo_full_name: primaryRepo.repo_full_name,
          key: projectKey,
          name: projectName,
          owner_user_id: ownerUserID,
          visibility: "internal",
          status: "active",
          start_date: new Date().toISOString().split("T")[0],
          due_date: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split("T")[0], // D+30
        }
      });
      setIsPromoteOpen(false);
      void loadData();
    } catch (err) {
      console.error(err);
      alert("Promotion failed: " + toUserErrorMessage(err, "Promotion failed."));
    } finally {
      setPromoting(false);
    }
  };

  if (loading) {
    return <PageLoading label="Loading application dashboard..." />;
  }

  if (error || !dashboard) {
    return (
      <div className="space-y-6">
        <PageError message={error || "The requested application dashboard could not be loaded."} onRetry={() => void loadData()} />
        <div>
          <button
            onClick={() => router.back()}
            className="px-6 py-2 rounded-xl bg-primary text-primary-foreground font-bold text-sm"
          >
            Go Back
          </button>
        </div>
      </div>
    );
  }

  const qualityScore = dashboard.metrics_overview.quality_score.toFixed(1);
  const criticalWarnings = dashboard.metrics_overview.critical_warning_count;
  const gateFailures = dashboard.quality_metrics.unresolved_issues.blocker;
  const buildStatus = dashboard.metrics_overview.target_branch_build_status;
  // REQ-FR-APPDASH-001 — 단순 % 보다 broken/red 상태 즉시 표기. buildStatus 는
  // backend dashboard 응답의 target_branch_build_status ("healthy"|"broken"|"unknown").
  const lastBuildView = applicationBuildStatusView(buildStatus);

  return (
    <div className="space-y-8 pb-20 px-4 md:px-8">
      {/* Header glass panel */}
      <div className="relative p-6 md:p-8 rounded-3xl glass border border-white/10 dark:border-white/5 flex flex-col md:flex-row md:items-center justify-between gap-6 overflow-hidden">
        <div className="absolute top-0 right-0 w-96 h-96 bg-primary/10 rounded-full blur-3xl -z-10" />
        
        <div className="flex items-center gap-4">
          <button 
            onClick={() => router.back()}
            className="p-3 rounded-2xl glass hover:bg-white/10 transition-all text-muted-foreground group"
          >
            <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
          </button>
          <div>
            <div className="flex flex-wrap items-center gap-3 mb-1">
              <h1 className="text-3xl font-black text-foreground tracking-tight">{dashboard.name}</h1>
              <Badge variant={lifecycleStatusBadgeVariant(dashboard.status)} dot>{dashboard.status}</Badge>
              <Badge variant="secondary" className="bg-white/5 backdrop-blur-md border border-white/10">{dashboard.visibility}</Badge>
            </div>
            <p className="text-muted-foreground text-sm flex flex-wrap items-center gap-x-2 gap-y-1">
              <Clock className="w-4 h-4" /> Updated {new Date(dashboard.updated_at).toLocaleDateString()} • {dashboard.key}
              <span className="hidden md:inline">•</span>
              <span>Leader: {dashboard.leader}</span>
            </p>
          </div>
        </div>
        <div className="flex items-center gap-3 self-end md:self-auto">
          <button onClick={() => void loadData()} className="p-3 rounded-2xl glass border border-white/10 hover:bg-white/5 text-muted-foreground hover:text-foreground transition-all">
            <RefreshCcw className="w-5 h-5" />
          </button>
          <button 
            onClick={() => setIsEditModalOpen(true)}
            className="p-3 rounded-2xl glass border border-white/10 hover:bg-white/5 text-muted-foreground hover:text-foreground transition-all"
            title="Edit Application Settings"
          >
            <Settings className="w-5 h-5" />
          </button>
        </div>
      </div>

      {/* Overview stats layout */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
        {[
          { label: "Last Build", value: lastBuildView.label, icon: Activity, color: lastBuildView.tone === "negative" ? "text-rose-500" : lastBuildView.tone === "positive" ? "text-emerald-500" : "text-muted-foreground", trend: "Latest run", bg: lastBuildView.tone === "negative" ? "bg-rose-500/10" : "bg-emerald-500/10" },
          { label: "Quality Score", value: `${qualityScore} / 5.0`, icon: ShieldCheck, color: "text-blue-500", trend: "Standard A+", bg: "bg-blue-500/10" },
          { label: "Critical Warnings", value: String(criticalWarnings), icon: Zap, color: criticalWarnings > 0 ? "text-amber-500" : "text-emerald-500", trend: "Governance", bg: "bg-amber-500/10" },
          { label: "Gate Failures", value: String(gateFailures), icon: Globe, color: gateFailures > 0 ? "text-rose-500" : "text-emerald-500", trend: "Quality Gate", bg: "bg-rose-500/10" },
        ].map((stat, i) => (
          <motion.div 
            key={stat.label}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.05 }}
            className="glass-card p-6 flex flex-col justify-between relative overflow-hidden group hover:border-primary/20 transition-all duration-300"
          >
            <div className="absolute top-0 right-0 w-24 h-24 bg-white/5 rounded-full blur-2xl -z-10 group-hover:bg-primary/5 transition-colors" />
            <div className="flex items-center justify-between mb-4">
              <div className={cn("p-3 rounded-2xl border border-white/5", stat.bg, stat.color)}>
                <stat.icon className="w-5 h-5" />
              </div>
              <span className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
                {stat.trend}
              </span>
            </div>
            <div>
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">{stat.label}</p>
              <h3 className="text-3xl font-black text-foreground tracking-tight">{stat.value}</h3>
            </div>
          </motion.div>
        ))}
      </div>

      {/* Main dashboard body */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        
        {/* Left Column: Build Stability & Quality Analysis */}
        <div className="lg:col-span-2 space-y-8">
          
          {/* Target Branch Build Status */}
          <section className="glass-card p-8 relative overflow-hidden">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-lg font-bold text-foreground flex items-center gap-2">
                <GitBranch className={cn("w-5 h-5", buildStatus === "broken" ? "text-rose-500 animate-pulse" : buildStatus === "healthy" ? "text-emerald-500" : "text-muted-foreground")} /> 
                Target Branch Build Status
              </h3>
              <Badge variant={buildStatus === "broken" ? "danger" : buildStatus === "healthy" ? "success" : "secondary"}>
                {buildStatus === "broken" ? "Broken" : buildStatus === "healthy" ? "Healthy" : "없음"}
              </Badge>
            </div>

            {buildStatus === "broken" ? (
              <div className="space-y-4">
                <div className="p-4 rounded-2xl bg-rose-500/10 border border-rose-500/20 text-rose-600 dark:text-rose-400 flex items-start gap-3">
                  <AlertTriangle className="w-5 h-5 shrink-0 mt-0.5" />
                  <div>
                    <h4 className="text-sm font-bold">Target branches are broken!</h4>
                    <p className="text-xs text-muted-foreground mt-1">Please inspect failed build runs below and fix the issues immediately to resume deployment pipeline.</p>
                  </div>
                </div>
                <div className="divide-y divide-white/5 border border-white/10 dark:border-white/5 rounded-2xl overflow-hidden bg-white/5">
                  {dashboard.build_failures.map((fail, index) => (
                    <div key={index} className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4 hover:bg-white/5 transition-colors">
                      <div>
                        <div className="flex items-center gap-2 mb-1">
                          <span className="text-xs font-bold text-foreground">{fail.repo_slug}</span>
                          <Badge variant="secondary" className="text-[10px] scale-90">{fail.branch}</Badge>
                        </div>
                        <p className="text-xs text-muted-foreground font-mono truncate max-w-md">{fail.error_snippet}</p>
                      </div>
                      <div className="flex items-center gap-3 self-end sm:self-auto">
                        <span className="text-[10px] text-muted-foreground font-mono">#{fail.build_number}</span>
                        <a 
                          href={fail.log_url}
                          className="px-3 py-1.5 rounded-xl bg-rose-500/20 hover:bg-rose-500 text-rose-500 hover:text-white font-bold text-xs flex items-center gap-1.5 transition-all"
                        >
                          <Play className="w-3.5 h-3.5 fill-current" /> Log URL
                        </a>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ) : (
              <div className="py-10 text-center flex flex-col items-center justify-center gap-3">
                <div className="w-16 h-16 rounded-full bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-500">
                  <Sparkles className="w-8 h-8" />
                </div>
                <h4 className="text-md font-bold text-foreground">All Branches Healthy 🟢</h4>
                <p className="text-xs text-muted-foreground max-w-sm">There are no currently failing build runs on primary SCM integration lines.</p>
              </div>
            )}
          </section>

          {/* SCM History & Build Trend (Recharts Area Chart) */}
          <section className="glass-card p-8">
            <h3 className="text-md font-bold text-foreground mb-6 flex items-center gap-2">
              <Activity className="w-4 h-4 text-primary" /> Build & Quality 7-Day Trend
            </h3>
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={dashboard.history_trend}>
                  <defs>
                    <linearGradient id="colorAvg" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="var(--primary)" stopOpacity={0.3}/>
                      <stop offset="95%" stopColor="var(--primary)" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <XAxis dataKey="date" axisLine={false} tickLine={false} tick={{ fill: "var(--muted-foreground)", fontSize: 10, fontWeight: 700 }} />
                  <YAxis axisLine={false} tickLine={false} tick={{ fill: "var(--muted-foreground)", fontSize: 10, fontWeight: 700 }} />
                  <Tooltip 
                    contentStyle={{ backgroundColor: 'var(--card)', borderRadius: '16px', border: '1px solid var(--border)' }}
                  />
                  <Area type="monotone" dataKey="avg_duration_seconds" stroke="var(--primary)" fillOpacity={1} fill="url(#colorAvg)" strokeWidth={3} />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </section>

          {/* Linked Projects (Milestones) Progress */}
          <section className="glass-card p-8">
            <h3 className="text-lg font-bold text-foreground mb-6 flex items-center gap-2">
              <Layers className="w-5 h-5 text-muted-foreground" /> Linked Projects Roadmap
            </h3>
            <div className="space-y-6">
              {dashboard.projects_progress.map((project) => (
                <div key={project.project_id} className="p-5 rounded-2xl border border-white/10 dark:border-white/5 bg-white/5 backdrop-blur-md space-y-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <h4 className="text-sm font-bold text-foreground flex items-center gap-2">
                        {project.name}
                        <Badge variant="secondary" className="scale-90 font-mono">{project.key}</Badge>
                      </h4>
                      <p className="text-[10px] text-muted-foreground mt-0.5">Due: {project.due_date ? new Date(project.due_date).toLocaleDateString() : "N/A"}</p>
                    </div>
                    <Badge
                      className="text-xs"
                      variant={project.risk_level === "At Risk" ? "danger" : project.risk_level === "Warning" ? "warning" : "success"}
                    >
                      {project.risk_level} (D-{project.d_day})
                    </Badge>
                  </div>
                  {/* Custom Story Point Progress Bar */}
                  <div>
                    <div className="flex justify-between text-[10px] font-black text-muted-foreground uppercase mb-1.5 tracking-wider">
                      <span>Milestone Progress</span>
                      <span>{project.progress_percent}%</span>
                    </div>
                    <div className="w-full h-3 bg-white/10 dark:bg-black/20 rounded-full overflow-hidden">
                      <div 
                        className="h-full bg-primary rounded-full transition-all duration-500" 
                        style={{ width: `${project.progress_percent}%` }}
                      />
                    </div>
                  </div>
                </div>
              ))}
              {dashboard.projects_progress.length === 0 && (
                <div className="py-12 text-center text-muted-foreground text-sm">
                  No active projects currently linked.
                </div>
              )}
            </div>
          </section>

        </div>

        {/* Right Column: Quality Metrics & Dev Requests */}
        <div className="space-y-8">
          
          {/* Quality Analysis Details */}
          <section className="glass-card p-8">
            <h3 className="text-md font-bold text-foreground mb-6 flex items-center gap-2">
              <ShieldCheck className="w-4 h-4 text-primary" /> Static Analysis Quality Gates
            </h3>
            <div className="space-y-6">
              <div className="p-6 rounded-2xl bg-white/5 border border-white/10 dark:border-white/5 text-center">
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">Normalized score</p>
                <h4 className="text-4xl font-black text-foreground">{qualityScore} <span className="text-lg text-muted-foreground">/ 5.0</span></h4>
                <p className="text-xs text-emerald-500 mt-2 font-bold">Stable Quality Grade</p>
              </div>
              <div className="grid grid-cols-3 gap-4 text-center">
                {[
                  { label: "Blocker", value: gateFailures, color: "text-rose-500" },
                  { label: "Critical", value: 0, color: "text-amber-500" },
                  { label: "Major", value: criticalWarnings, color: "text-blue-500" }
                ].map((gate) => (
                  <div key={gate.label} className="p-3 rounded-xl bg-white/5 border border-white/5">
                    <p className="text-[10px] font-black text-muted-foreground uppercase mb-1">{gate.label}</p>
                    <span className={cn("text-lg font-black", gate.color)}>{gate.value}</span>
                  </div>
                ))}
              </div>
              <p className="text-[10px] text-muted-foreground text-center italic">{dashboard.quality_metrics.comment}</p>
            </div>
          </section>

          {/* Linked DREQs & Promotion Actions */}
          <section className="glass-card p-8">
            <h3 className="text-md font-bold text-foreground mb-6 flex items-center gap-2">
              <Briefcase className="w-4 h-4 text-primary" /> Dev Requests (DREQ)
            </h3>
            <div className="space-y-4">
              {dashboard.linked_dev_requests.map((dreq) => (
                <div key={dreq.dreq_id} className="p-4 rounded-xl border border-white/5 bg-white/5 backdrop-blur-md flex items-center justify-between gap-4">
                  <div className="min-w-0">
                    <h4 className="text-xs font-bold text-foreground truncate">{dreq.title}</h4>
                    <p className="text-[10px] text-muted-foreground mt-0.5 font-mono">{dreq.dreq_id}</p>
                  </div>
                  <div className="flex items-center gap-3">
                    <Badge variant={dreq.status === "pending" ? "warning" : "primary"}>{dreq.status}</Badge>
                    {(dreq.status === "pending" || dreq.status === "in_review") && (
                      <button 
                        onClick={() => handlePromoteClick(dreq.dreq_id, dreq.title)}
                        className="p-1.5 rounded-lg bg-primary/20 hover:bg-primary text-primary hover:text-white transition-all group"
                      >
                        <Rocket className="w-3.5 h-3.5 group-hover:scale-110 transition-transform" />
                      </button>
                    )}
                  </div>
                </div>
              ))}
              {dashboard.linked_dev_requests.length === 0 && (
                <div className="py-8 text-center text-muted-foreground text-xs">
                  No development requests mapped to this application.
                </div>
              )}
            </div>
          </section>

          {/* Linked SCM Repositories list */}
          <section className="glass-card p-8">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-md font-bold text-foreground flex items-center gap-2">
                <GitBranch className="w-4 h-4 text-muted-foreground" /> Repositories
              </h3>
              <Badge variant="secondary">{repositories.length} Linked</Badge>
            </div>
            <div className="space-y-3">
              {repositories.map((repo, i) => (
                <div key={i} className="p-4 rounded-xl border border-white/5 hover:border-white/10 bg-white/5 hover:bg-white/10 transition-all flex items-center justify-between">
                  <div>
                    <h4 className="text-xs font-bold text-foreground">{repo.repo_full_name}</h4>
                    <p className="text-[9px] text-muted-foreground uppercase tracking-widest mt-0.5">{repo.repo_provider} • {repo.role}</p>
                  </div>
                  <Badge variant={repo.sync_status === "active" ? "success" : "warning"}>{repo.sync_status}</Badge>
                </div>
              ))}
            </div>
          </section>

        </div>

      </div>

      {/* DREQ Promotion Modal (AnimatePresence glass panel) */}
      <AnimatePresence>
        {isPromoteOpen && (
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
            <motion.div 
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="absolute inset-0 bg-black/60 backdrop-blur-md"
              onClick={() => setIsPromoteOpen(false)}
            />
            <motion.div 
              initial={{ opacity: 0, scale: 0.95, y: 20 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.95, y: 20 }}
              className="relative w-full max-w-lg p-8 rounded-3xl glass border border-white/10 bg-card dark:bg-zinc-900/90 shadow-2xl z-10 space-y-6"
            >
              <div className="flex items-center justify-between">
                <h3 className="text-xl font-black text-foreground flex items-center gap-2">
                  <Rocket className="w-5 h-5 text-primary" /> Promote to Project
                </h3>
                <button 
                  onClick={() => setIsPromoteOpen(false)}
                  className="p-2 rounded-xl hover:bg-white/10 text-muted-foreground transition-all"
                >
                  ✕
                </button>
              </div>

              <form onSubmit={handlePromoteSubmit} className="space-y-4">
                <div>
                  <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest block mb-1">DREQ Title (Inherited)</label>
                  <input 
                    type="text"
                    value={projectName}
                    onChange={(e) => setProjectName(e.target.value)}
                    required
                    className="w-full px-4 py-3 rounded-xl border border-white/10 bg-white/5 text-foreground text-sm focus:outline-none focus:border-primary transition-all"
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest block mb-1">Project Key</label>
                    <input 
                      type="text"
                      value={projectKey}
                      onChange={(e) => setProjectKey(e.target.value)}
                      required
                      className="w-full px-4 py-3 rounded-xl border border-white/10 bg-white/5 text-foreground text-sm focus:outline-none focus:border-primary transition-all"
                    />
                  </div>
                  <div>
                    <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest block mb-1">Leader User ID</label>
                    <input 
                      type="text"
                      value={projectLeader}
                      onChange={(e) => setProjectLeader(e.target.value)}
                      placeholder="e.g. u1"
                      required
                      className="w-full px-4 py-3 rounded-xl border border-white/10 bg-white/5 text-foreground text-sm focus:outline-none focus:border-primary transition-all"
                    />
                  </div>
                </div>
                
                <div className="flex justify-end gap-3 pt-4">
                  <button 
                    type="button" 
                    onClick={() => setIsPromoteOpen(false)}
                    className="px-5 py-2.5 rounded-xl border border-white/10 text-sm text-foreground hover:bg-white/5 transition-all"
                  >
                    Cancel
                  </button>
                  <button 
                    type="submit" 
                    disabled={promoting}
                    className="px-5 py-2.5 rounded-xl bg-primary text-primary-foreground font-bold text-sm flex items-center gap-1.5 hover:opacity-90 disabled:opacity-50 transition-all"
                  >
                    {promoting ? "Promoting..." : "Promote Project 🚀"}
                  </button>
                </div>
              </form>
            </motion.div>
          </div>
        )}
        {isEditModalOpen && application && (
          <ApplicationCreationModal
            initialData={{
              ...application,
              start_date: application.start_date ?? undefined,
              due_date: application.due_date ?? undefined,
              archived_at: application.archived_at ?? undefined,
            }}
            onClose={() => setIsEditModalOpen(false)}
            onCreated={(newApp) => {
              toast(`Application ${newApp.name} updated successfully`, "success");
              setIsEditModalOpen(false);
              void loadData();
            }}
          />
        )}
      </AnimatePresence>
    </div>
  );
}
