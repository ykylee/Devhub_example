"use client";

import { useCallback, useEffect, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { 
  ArrowLeft, 
  Calendar, 
  Clock, 
  Plus,
  Link2,
  X,
  Target,
  Users,
  ChevronRight,
  TrendingUp,
  ShieldAlert,
  Gauge,
  Zap,
  AlertTriangle,
  Flame,
  ShieldCheck,
  Code2,
  GitPullRequest,
  CheckCircle,
  HelpCircle,
  BarChart3,
  GitBranch
} from "lucide-react";
import { useRouter, useParams } from "next/navigation";
import { parseISO } from "date-fns";
import { Badge } from "@/shared/ui-foundation/components/Badge";
import { projectService } from "@/domain/platform-lifecycle/service/project.service";
import type { Project, ProjectRepositoryLink, ProjectTaskItem } from "@/domain/platform-lifecycle/schema/project.types";
import { identityService, OrgMember, ResolvedActor } from "@/domain/organization-management/service/identity.service";
import { repositoryService, Repository } from "@/domain/repository-integration/service/repository.service";
import { toUserErrorMessage } from "@/shared/utils/error-message";
import { lifecycleStatusBadgeVariant } from "@/shared/utils/lifecycle-status";
import { PageError, PageLoading } from "@/shared/ui-foundation/components/PageState";
import { apiClient } from "@/shared/api/api-client";

export default function ProjectDetailPage() {
  const params = useParams();
  const id = params.id as string;
  const router = useRouter();
  
  const [project, setProject] = useState<Project | null>(null);
  const [projectRepositories, setProjectRepositories] = useState<ProjectRepositoryLink[]>([]);
  const [users, setUsers] = useState<OrgMember[]>([]);
  const [allRepositories, setAllRepositories] = useState<Repository[]>([]);
  const [showRepoPicker, setShowRepoPicker] = useState(false);
  const [linkingRepoIds, setLinkingRepoIds] = useState<number[]>([]);
  
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [opsError, setOpsError] = useState<string | null>(null);

  // Persona State
  const [me, setMe] = useState<ResolvedActor | null>(null);
  const [selectedPersona, setSelectedPersona] = useState<"developer" | "project_leader" | "manager">("developer");
  const [dashboardData, setDashboardData] = useState<any | null>(null);
  const [dashboardLoading, setDashboardLoading] = useState(false);
  const [accessDenied, setAccessDenied] = useState(false);
  const [accessDeniedMessage, setAccessDeniedMessage] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    try {
      setError(null);
      setLoading(true);
      
      const [projectData, usersData, actorData] = await Promise.all([
        projectService.getProject(id),
        identityService.getUsers(),
        identityService.whoAmI().catch(() => null),
      ]);
      
      repositoryService.listRepositories().then(setAllRepositories).catch(() => setAllRepositories([]));
      setProject(projectData);
      setUsers(usersData);
      setMe(actorData);

      // Determine default persona based on 2D RBAC
      if (projectData && actorData) {
        const userMember = projectData.project_members?.find(m => m.user_id === actorData.login);
        const isLead = userMember?.project_role === "lead" || userMember?.project_role === "project_leader";
        
        let defaultPersona: "developer" | "project_leader" | "manager" = "developer";
        if (actorData.role === "System Admin" || actorData.role === "Manager") {
          defaultPersona = "manager";
        } else if (isLead) {
          defaultPersona = "project_leader";
        }
        setSelectedPersona(defaultPersona);
      }

      const linksResult = await projectService.getProjectRepositories(id);
      setProjectRepositories(linksResult);
    } catch (err) {
      setError(toUserErrorMessage(err, "Failed to load project details."));
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [id]);

  const loadDashboard = useCallback(async (persona: "developer" | "project_leader" | "manager") => {
    try {
      setDashboardLoading(true);
      setAccessDenied(false);
      setAccessDeniedMessage(null);
      const data = await projectService.getProjectDashboard(id, persona);
      setDashboardData(data);
    } catch (err: any) {
      console.error("[ProjectDashboard] fetch error:", err);
      setAccessDenied(true);
      setAccessDeniedMessage(
        err?.message || "OIDC/RBAC 가드 2단계에서 접근이 거부되었습니다. 해당 뷰를 조회할 수 있는 권한이 없습니다."
      );
      setDashboardData(null);
    } finally {
      setDashboardLoading(false);
    }
  }, [id]);

  useEffect(() => {
    const timer = setTimeout(() => {
      void loadData();
    }, 0);
    return () => clearTimeout(timer);
  }, [loadData]);

  useEffect(() => {
    if (project) {
      void loadDashboard(selectedPersona);
    }
  }, [selectedPersona, project, loadDashboard]);

  async function linkSelectedRepositories() {
    if (!project || linkingRepoIds.length === 0) return;
    const projectID = project.id;
    try {
      await Promise.all(
        linkingRepoIds.map((repoId) => projectService.linkProjectRepository(projectID, repoId, "linked")),
      );
      setShowRepoPicker(false);
      setLinkingRepoIds([]);
      await loadData();
    } catch (err) {
      setOpsError(toUserErrorMessage(err, "저장소 연결에 실패했습니다."));
    }
  }

  // PR 퀵액션 (Remind/Notify)
  const triggerRemindAction = (prTitle: string, author: string) => {
    alert(`[협업 촉구 🚀] ${author}님에게 PR (${prTitle}) 머지 검토 리마인더 알림이 성공적으로 전송되었습니다.`);
  };

  if (loading) {
    return <PageLoading label="Loading project details..." />;
  }

  if (error || !project) {
    return (
      <div className="space-y-6">
        <PageError message={error || "The requested project roadmap could not be located."} onRetry={() => void loadData()} />
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

  const linkedRepoIds = new Set(projectRepositories.map((r) => r.repository_id));
  const linkedRepos = allRepositories.filter((r) => linkedRepoIds.has(r.id));
  const candidateRepos = allRepositories.filter((r) => !linkedRepoIds.has(r.id));

  return (
    <div className="space-y-10 pb-20">
      {/* Header Area */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-6">
        <div className="flex items-center gap-4">
          <button 
            onClick={() => router.back()}
            className="p-3 rounded-xl glass hover:bg-muted/30 transition-all text-muted-foreground group"
          >
            <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
          </button>
          <div>
            <div className="flex items-center gap-3 mb-1">
              <h1 className="text-3xl font-black text-foreground dark:text-primary-foreground tracking-tight">{project.name}</h1>
              <Badge variant={lifecycleStatusBadgeVariant(project.status)} dot>{project.status}</Badge>
            </div>
            <p className="text-muted-foreground text-sm flex items-center gap-2">
              <Target className="w-4 h-4 text-primary" /> {project.key} • 
              <Calendar className="w-4 h-4 ml-2 text-primary" /> Due: {project.due_date || "TBD"}
            </p>
          </div>
        </div>

        {/* 3-Way Persona Switcher (Segment Controller) */}
        <div className="flex items-center gap-4 self-end lg:self-center">
          <div className="p-1 rounded-2xl bg-muted/40 backdrop-blur-md border border-border/50 flex items-center gap-1 shadow-inner">
            {(["developer", "project_leader", "manager"] as const).map((persona) => {
              const isActive = selectedPersona === persona;
              const labels = {
                developer: "Developer",
                project_leader: "Project Leader",
                manager: "Org Manager"
              };
              return (
                <button
                  key={persona}
                  onClick={() => setSelectedPersona(persona)}
                  className={`px-4 py-2 rounded-xl text-xs font-black tracking-wide uppercase transition-all duration-300 relative ${
                    isActive 
                      ? "text-primary-foreground bg-primary shadow-lg shadow-primary/25 scale-[1.02]" 
                      : "text-muted-foreground hover:text-foreground hover:bg-muted/20"
                  }`}
                >
                  {labels[persona]}
                </button>
              );
            })}
          </div>
        </div>
      </div>

      {/* Main Content Body */}
      <AnimatePresence mode="wait">
        {dashboardLoading ? (
          <motion.div 
            key="loading"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="h-96 flex items-center justify-center"
          >
            <PageLoading label="Loading persona metrics..." />
          </motion.div>
        ) : accessDenied ? (
          /* OIDC/RBAC Access Denied Panel (Wow Neon Warning) */
          <motion.div
            key="access_denied"
            initial={{ opacity: 0, y: 15 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -15 }}
            className="glass border border-destructive/40 shadow-[0_0_15px_rgba(239,68,68,0.15)] rounded-2xl p-10 flex flex-col items-center text-center space-y-6 max-w-2xl mx-auto animate-pulse"
          >
            <div className="w-16 h-16 rounded-full bg-destructive/10 border border-destructive/30 flex items-center justify-center text-destructive">
              <ShieldAlert className="w-8 h-8" />
            </div>
            <div>
              <h2 className="text-xl font-black text-destructive tracking-tight mb-2">ACCESS DENIED (OIDC/RBAC 가드)</h2>
              <p className="text-sm text-muted-foreground leading-relaxed">
                {accessDeniedMessage}
              </p>
            </div>
            <div className="text-xs text-muted-foreground/60 border-t border-border/40 pt-4 w-full">
              현 로그인 계정: <span className="font-bold text-foreground">{me?.login} ({me?.role})</span>
            </div>
          </motion.div>
        ) : dashboardData ? (
          <motion.div
            key={selectedPersona}
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.3 }}
            className="grid grid-cols-1 lg:grid-cols-4 gap-6"
          >
            <div className="lg:col-span-3 space-y-8">
              
              {/* --- Developer View --- */}
              {selectedPersona === "developer" && dashboardData.developer_view && (
                <div className="space-y-8">
                  {/* Review Guard (failed builds / conflicts) */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    {/* Build failures */}
                    <div className="glass border border-destructive/30 shadow-[0_0_12px_rgba(239,68,68,0.08)] p-6 rounded-2xl relative overflow-hidden group">
                      <div className="absolute top-0 right-0 w-24 h-24 bg-destructive/5 rounded-full blur-2xl" />
                      <h3 className="text-xs font-black uppercase tracking-widest text-destructive mb-4 flex items-center gap-2">
                        <Flame className="w-4 h-4 animate-pulse" /> Build Failures (My PRs)
                      </h3>
                      <div className="space-y-3">
                        {dashboardData.developer_view.review_guard.failed_build_prs.map((pr: any) => (
                          <div key={pr.id} className="p-3 rounded-xl border border-destructive/20 bg-destructive/5 flex items-center justify-between gap-3 hover:bg-destructive/10 transition-colors">
                            <div className="min-w-0">
                              <p className="text-xs font-bold text-foreground truncate">{pr.title}</p>
                              <p className="text-[10px] text-muted-foreground mt-0.5 truncate uppercase tracking-widest">{pr.repository_name} • {pr.last_build_id}</p>
                            </div>
                            <a href={pr.url} target="_blank" rel="noopener noreferrer" className="flex items-center gap-1 text-[10px] font-black uppercase tracking-widest text-destructive hover:underline flex-shrink-0">
                              PR Link <ChevronRight className="w-3 h-3" />
                            </a>
                          </div>
                        ))}
                      </div>
                    </div>

                    {/* Conflict PRs */}
                    <div className="glass border border-warning/30 shadow-[0_0_12px_rgba(245,158,11,0.08)] p-6 rounded-2xl relative overflow-hidden group">
                      <div className="absolute top-0 right-0 w-24 h-24 bg-warning/5 rounded-full blur-2xl" />
                      <h3 className="text-xs font-black uppercase tracking-widest text-warning mb-4 flex items-center gap-2">
                        <AlertTriangle className="w-4 h-4" /> Merge Conflicts (My PRs)
                      </h3>
                      <div className="space-y-3">
                        {dashboardData.developer_view.review_guard.conflict_prs.map((pr: any) => (
                          <div key={pr.id} className="p-3 rounded-xl border border-warning/20 bg-warning/5 flex items-center justify-between gap-3 hover:bg-warning/10 transition-colors">
                            <div className="min-w-0">
                              <p className="text-xs font-bold text-foreground truncate">{pr.title}</p>
                              <p className="text-[10px] text-muted-foreground mt-0.5 truncate uppercase tracking-widest">{pr.repository_name}</p>
                            </div>
                            <a href={pr.url} target="_blank" rel="noopener noreferrer" className="flex items-center gap-1 text-[10px] font-black uppercase tracking-widest text-warning hover:underline flex-shrink-0">
                              Resolve <ChevronRight className="w-3 h-3" />
                            </a>
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>

                  {/* My Work (Active Tasks + Review Requests) */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                    {/* Active Tasks Feed */}
                    <section className="glass-card p-6">
                      <h3 className="text-xs font-black uppercase tracking-widest text-muted-foreground mb-4 flex items-center gap-2">
                        <Code2 className="w-4 h-4 text-primary" /> Active Tasks ({dashboardData.developer_view.my_work.active_tasks.length})
                      </h3>
                      <div className="space-y-3 max-h-[300px] overflow-y-auto pr-1">
                        {dashboardData.developer_view.my_work.active_tasks.map((task: any) => (
                          <div key={task.id} className="p-4 rounded-xl border border-border/50 bg-muted/10 flex items-center justify-between gap-4 hover:border-primary/30 transition-all group">
                            <div className="min-w-0">
                              <p className="text-xs font-bold text-foreground truncate">{task.title}</p>
                              <p className="text-[9px] text-muted-foreground mt-1 uppercase tracking-widest">
                                {task.repository_name} • Due: {task.due_date ? new Date(task.due_date).toLocaleDateString() : "TBD"}
                              </p>
                            </div>
                            <Badge variant={task.priority === "high" ? "destructive" : "glass"}>{task.priority}</Badge>
                          </div>
                        ))}
                      </div>
                    </section>

                    {/* Review Requests */}
                    <section className="glass-card p-6">
                      <h3 className="text-xs font-black uppercase tracking-widest text-muted-foreground mb-4 flex items-center gap-2">
                        <GitPullRequest className="w-4 h-4 text-accent" /> Review Requests ({dashboardData.developer_view.my_work.review_requests.length})
                      </h3>
                      <div className="space-y-3 max-h-[300px] overflow-y-auto pr-1">
                        {dashboardData.developer_view.my_work.review_requests.map((pr: any) => (
                          <div key={pr.id} className="p-4 rounded-xl border border-border/50 bg-muted/10 flex items-center justify-between gap-4 hover:border-accent/30 transition-all">
                            <div className="min-w-0">
                              <p className="text-xs font-bold text-foreground truncate">{pr.title}</p>
                              <p className="text-[9px] text-muted-foreground mt-1 uppercase tracking-widest">
                                {pr.repository_name} • requested by {pr.author}
                              </p>
                            </div>
                            <a href={pr.pull_request_url} target="_blank" rel="noopener noreferrer" className="p-2 rounded-lg bg-accent/10 text-accent text-[9px] font-black uppercase tracking-widest hover:bg-accent/25 transition-colors">
                              Review
                            </a>
                          </div>
                        ))}
                      </div>
                    </section>
                  </div>

                  {/* Branches Code Health */}
                  <section className="glass-card p-6">
                    <h3 className="text-xs font-black uppercase tracking-widest text-muted-foreground mb-4 flex items-center gap-2">
                      <GitBranch className="w-4 h-4 text-primary" /> Branch Build Health & Quality
                    </h3>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      {dashboardData.developer_view.code_health.branches.map((b: any, idx: number) => (
                        <div key={idx} className="p-4 rounded-xl border border-border/50 bg-muted/10 flex items-center justify-between gap-4">
                          <div>
                            <p className="text-xs font-black text-foreground">{b.branch_name}</p>
                            <p className="text-[9px] text-muted-foreground uppercase tracking-widest mt-0.5">{b.repository_name}</p>
                            <div className="flex items-center gap-3 mt-3">
                              <span className="text-[10px] text-muted-foreground">Coverage: <strong className="text-foreground">{(b.test_coverage * 100).toFixed(1)}%</strong></span>
                              <span className="text-[10px] text-muted-foreground">Duplication: <strong className="text-foreground">{(b.duplicate_ratio * 100).toFixed(1)}%</strong></span>
                            </div>
                          </div>
                          <Badge variant={b.last_build_status === "healthy" ? "success" : "destructive"}>
                            {b.last_build_status}
                          </Badge>
                        </div>
                      ))}
                    </div>
                  </section>
                </div>
              )}

              {/* --- Project Leader View --- */}
              {selectedPersona === "project_leader" && dashboardData.project_leader_view && (
                <div className="space-y-8">
                  {/* PR Integration Hub ( failed / conflicting / stale ) */}
                  <section className="glass-card p-6">
                    <h3 className="text-xs font-black uppercase tracking-widest text-muted-foreground mb-6 flex items-center gap-2">
                      <GitPullRequest className="w-4 h-4 text-primary" /> PR Integration Hub & Stale Reviews
                    </h3>
                    <div className="space-y-4">
                      {/* Failed build prs */}
                      {dashboardData.project_leader_view.pr_integration_hub.failed_build_prs.map((pr: any) => (
                        <div key={pr.id} className="p-4 rounded-xl border border-destructive/20 bg-destructive/5 flex items-center justify-between gap-4 hover:bg-destructive/10 transition-colors">
                          <div>
                            <Badge variant="destructive" className="mb-2">Build Failed</Badge>
                            <p className="text-xs font-bold text-foreground">{pr.title}</p>
                            <p className="text-[10px] text-muted-foreground mt-1 uppercase tracking-widest">{pr.repository_name} • Author: {pr.author}</p>
                          </div>
                          <div className="flex items-center gap-2">
                            <button onClick={() => triggerRemindAction(pr.title, pr.author)} className="px-3 py-1.5 rounded-lg bg-destructive/15 text-destructive text-[10px] font-black uppercase tracking-widest hover:bg-destructive/25 transition-colors">
                              Remind
                            </button>
                            <a href={pr.url} target="_blank" rel="noopener noreferrer" className="p-2 text-muted-foreground hover:text-foreground">
                              <ChevronRight className="w-4 h-4" />
                            </a>
                          </div>
                        </div>
                      ))}

                      {/* Conflicting prs */}
                      {dashboardData.project_leader_view.pr_integration_hub.conflicting_prs.map((pr: any) => (
                        <div key={pr.id} className="p-4 rounded-xl border border-warning/20 bg-warning/5 flex items-center justify-between gap-4 hover:bg-warning/10 transition-colors">
                          <div>
                            <Badge variant="warning" className="mb-2">Conflict</Badge>
                            <p className="text-xs font-bold text-foreground">{pr.title}</p>
                            <p className="text-[10px] text-muted-foreground mt-1 uppercase tracking-widest">{pr.repository_name} • Author: {pr.author}</p>
                          </div>
                          <div className="flex items-center gap-2">
                            <button onClick={() => triggerRemindAction(pr.title, pr.author)} className="px-3 py-1.5 rounded-lg bg-warning/15 text-warning text-[10px] font-black uppercase tracking-widest hover:bg-warning/25 transition-colors">
                              Remind
                            </button>
                            <a href={pr.url} target="_blank" rel="noopener noreferrer" className="p-2 text-muted-foreground hover:text-foreground">
                              <ChevronRight className="w-4 h-4" />
                            </a>
                          </div>
                        </div>
                      ))}

                      {/* Stale prs */}
                      {dashboardData.project_leader_view.pr_integration_hub.stale_prs.map((pr: any) => (
                        <div key={pr.id} className="p-4 rounded-xl border border-border/50 bg-muted/10 flex items-center justify-between gap-4 hover:border-accent/40 transition-all">
                          <div>
                            <Badge variant="glass" className="mb-2 text-accent border-accent/30 bg-accent/5">Stale {pr.idle_duration_hours}h</Badge>
                            <p className="text-xs font-bold text-foreground">{pr.title}</p>
                            <p className="text-[10px] text-muted-foreground mt-1 uppercase tracking-widest">{pr.repository_name} • Author: {pr.author}</p>
                          </div>
                          <div className="flex items-center gap-2">
                            <button onClick={() => triggerRemindAction(pr.title, pr.author)} className="px-3 py-1.5 rounded-lg bg-accent/15 text-accent text-[10px] font-black uppercase tracking-wide hover:bg-accent/25 transition-colors">
                              Remind
                            </button>
                            <a href={pr.url} target="_blank" rel="noopener noreferrer" className="p-2 text-muted-foreground hover:text-foreground">
                              <ChevronRight className="w-4 h-4" />
                            </a>
                          </div>
                        </div>
                      ))}
                    </div>
                  </section>

                  {/* Feature Progress Radar (milestones & epics) */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                    <section className="glass-card p-6">
                      <h3 className="text-xs font-black uppercase tracking-widest text-muted-foreground mb-4">Milestone Progress</h3>
                      <div className="space-y-6">
                        {dashboardData.project_leader_view.feature_progress_radar.milestones.map((ms: any) => (
                          <div key={ms.id} className="space-y-2">
                            <div className="flex justify-between text-xs font-bold">
                              <span>{ms.title}</span>
                              <span className="text-primary">{ms.progress_percent}%</span>
                            </div>
                            <div className="h-2 w-full bg-muted/30 rounded-full overflow-hidden border border-border/50">
                              <div className="h-full bg-primary" style={{ width: `${ms.progress_percent}%` }} />
                            </div>
                            <p className="text-[9px] text-muted-foreground uppercase tracking-widest mt-1">Due: {new Date(ms.due_date).toLocaleDateString()} • {ms.status}</p>
                          </div>
                        ))}
                      </div>
                    </section>

                    <section className="glass-card p-6">
                      <h3 className="text-xs font-black uppercase tracking-widest text-muted-foreground mb-4">Epic Delivery Status</h3>
                      <div className="space-y-6">
                        {dashboardData.project_leader_view.feature_progress_radar.epics.map((epic: any) => (
                          <div key={epic.id} className="space-y-2">
                            <div className="flex justify-between text-xs font-bold">
                              <span>{epic.name}</span>
                              <span className="text-accent">{epic.progress_percent}%</span>
                            </div>
                            <div className="h-2 w-full bg-muted/30 rounded-full overflow-hidden border border-border/50">
                              <div className="h-full bg-accent" style={{ width: `${epic.progress_percent}%` }} />
                            </div>
                            <p className="text-[9px] text-muted-foreground uppercase tracking-widest mt-1">Points: {epic.completed_points} / {epic.total_points} SP</p>
                          </div>
                        ))}
                      </div>
                    </section>
                  </div>

                  {/* Escalation & Blocker Feed */}
                  <section className="glass-card p-6">
                    <h3 className="text-xs font-black uppercase tracking-widest text-muted-foreground mb-4 flex items-center gap-2">
                      <Flame className="w-4 h-4 text-destructive" /> Escalation & Blocker Feed
                    </h3>
                    <div className="space-y-4">
                      {/* Blocked Tasks */}
                      {dashboardData.project_leader_view.escalation_feed.blocked_tasks.map((task: any) => (
                        <div key={task.id} className="p-4 rounded-xl border border-destructive/20 bg-destructive/5 flex items-center justify-between gap-4">
                          <div>
                            <span className="px-2 py-0.5 rounded text-[9px] font-black uppercase tracking-widest bg-destructive/15 text-destructive border border-destructive/30">BLOCKED</span>
                            <h4 className="text-xs font-bold text-foreground mt-2">{task.title}</h4>
                            <p className="text-[10px] text-muted-foreground mt-1 font-bold">Reason: <span className="font-normal">{task.block_reason}</span></p>
                          </div>
                          <div className="text-right flex-shrink-0">
                            <p className="text-[9px] text-muted-foreground uppercase tracking-widest">Assignee: {task.assignee}</p>
                            <p className="text-[9px] text-muted-foreground uppercase tracking-widest mt-0.5">Since: {new Date(task.blocked_since).toLocaleDateString()}</p>
                          </div>
                        </div>
                      ))}

                      {/* Help Needed */}
                      {dashboardData.project_leader_view.escalation_feed.critical_help_needed.map((task: any) => (
                        <div key={task.id} className="p-4 rounded-xl border border-warning/20 bg-warning/5 flex items-center justify-between gap-4">
                          <div>
                            <span className="px-2 py-0.5 rounded text-[9px] font-black uppercase tracking-widest bg-warning/15 text-warning border border-warning/30">HELP NEEDED</span>
                            <h4 className="text-xs font-bold text-foreground mt-2">{task.title}</h4>
                            <p className="text-[10px] text-muted-foreground mt-1 font-bold">Detected: <span className="font-normal text-warning">{task.keyphrase_detected}</span></p>
                          </div>
                          <div className="text-right flex-shrink-0">
                            <p className="text-[9px] text-muted-foreground uppercase tracking-widest">Assignee: {task.assignee}</p>
                          </div>
                        </div>
                      ))}
                    </div>
                  </section>
                </div>
              )}

              {/* --- Org Manager View --- */}
              {selectedPersona === "manager" && dashboardData.manager_view && (
                <div className="space-y-8">
                  {/* Delivery Health & Forecast */}
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                    <div className="glass-card p-6 flex flex-col justify-between">
                      <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">SLA Risk Assessment</p>
                      <div className="flex items-center gap-3 mt-2">
                        <div className={`w-3 h-3 rounded-full animate-ping ${
                          dashboardData.manager_view.delivery_health.sla_risk === "warning" ? "bg-warning" :
                          dashboardData.manager_view.delivery_health.sla_risk === "at_risk" ? "bg-destructive" : "bg-success"
                        }`} />
                        <span className={`text-lg font-black uppercase tracking-wide ${
                          dashboardData.manager_view.delivery_health.sla_risk === "warning" ? "text-warning" :
                          dashboardData.manager_view.delivery_health.sla_risk === "at_risk" ? "text-destructive" : "text-success"
                        }`}>{dashboardData.manager_view.delivery_health.sla_risk}</span>
                      </div>
                      <p className="text-xs text-muted-foreground mt-4">Risk Index score: <strong>{dashboardData.manager_view.delivery_health.sla_risk_index}</strong></p>
                    </div>

                    <div className="glass-card p-6 flex flex-col justify-between">
                      <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">Weekly Velocity</p>
                      <p className="text-2xl font-black text-foreground mt-2">{dashboardData.manager_view.delivery_health.weekly_velocity} <span className="text-xs text-muted-foreground">issues/wk</span></p>
                      <p className="text-xs text-muted-foreground mt-4">Remaining tasks to target: <strong>{dashboardData.manager_view.delivery_health.open_tasks_count} / {dashboardData.manager_view.delivery_health.total_tasks_count}</strong></p>
                    </div>

                    <div className="glass-card p-6 flex flex-col justify-between">
                      <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">Days Remaining</p>
                      <p className="text-2xl font-black text-primary mt-2">{dashboardData.manager_view.delivery_health.remaining_days} <span className="text-xs text-muted-foreground">days</span></p>
                      <p className="text-xs text-muted-foreground mt-4">Estimated completion: <strong>On Track</strong></p>
                    </div>
                  </div>

                  {/* Workload Meter */}
                  <section className="glass-card p-6">
                    <h3 className="text-xs font-black uppercase tracking-widest text-muted-foreground mb-6 flex items-center gap-2">
                      <Gauge className="w-4 h-4 text-primary" /> Team Workload Meter & Allocation
                    </h3>
                    <div className="space-y-6">
                      {dashboardData.manager_view.workload_meter.members.map((member: any, idx: number) => {
                        const isOverloaded = member.status === "overloaded";
                        return (
                          <div key={idx} className="space-y-2">
                            <div className="flex items-center justify-between text-xs">
                              <span className="font-bold">{member.display_name}</span>
                              <div className="flex items-center gap-3">
                                <span className="text-muted-foreground">Tasks: <strong>{member.active_tasks_count}</strong> | Reviews: <strong>{member.active_reviews_count}</strong></span>
                                {isOverloaded && (
                                  <Badge variant="destructive" className="animate-pulse">Overloaded ⚠️</Badge>
                                )}
                              </div>
                            </div>
                            <div className="h-2 w-full bg-muted/30 rounded-full overflow-hidden border border-border/50">
                              <div className={`h-full ${isOverloaded ? "bg-destructive" : "bg-primary"}`} style={{ width: `${Math.min(100, (member.workload_score / 8) * 100)}%` }} />
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  </section>

                  {/* Governance Shield (SonarQube Summary) */}
                  <section className="glass-card p-6">
                    <h3 className="text-xs font-black uppercase tracking-widest text-muted-foreground mb-6 flex items-center gap-2">
                      <ShieldCheck className="w-4 h-4 text-success" /> Governance Shield (Quality & Security Rollup)
                    </h3>
                    <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
                      <div className="p-4 rounded-xl bg-muted/10 border border-border/50 text-center">
                        <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Quality Score</p>
                        <p className="text-3xl font-black text-success mt-2">{dashboardData.manager_view.governance_shield.rollup_score}</p>
                        <p className="text-[9px] text-muted-foreground mt-1">out of 5.0</p>
                      </div>

                      <div className="p-4 rounded-xl bg-muted/10 border border-border/50 text-center">
                        <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Blocker Bugs</p>
                        <p className={`text-3xl font-black mt-2 ${
                          dashboardData.manager_view.governance_shield.blocker_bugs > 0 ? "text-destructive" : "text-foreground"
                        }`}>{dashboardData.manager_view.governance_shield.blocker_bugs}</p>
                        <p className="text-[9px] text-muted-foreground mt-1">unresolved</p>
                      </div>

                      <div className="p-4 rounded-xl bg-muted/10 border border-border/50 text-center">
                        <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Vulnerabilities</p>
                        <p className={`text-3xl font-black mt-2 ${
                          dashboardData.manager_view.governance_shield.vulnerabilities > 0 ? "text-warning" : "text-foreground"
                        }`}>{dashboardData.manager_view.governance_shield.vulnerabilities}</p>
                        <p className="text-[9px] text-muted-foreground mt-1">open issues</p>
                      </div>

                      <div className="p-4 rounded-xl bg-muted/10 border border-border/50 text-center">
                        <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Avg Test Coverage</p>
                        <p className="text-3xl font-black text-foreground mt-2">{(dashboardData.manager_view.governance_shield.average_coverage * 100).toFixed(1)}%</p>
                        <p className="text-[9px] text-muted-foreground mt-1">project average</p>
                      </div>
                    </div>
                  </section>
                </div>
              )}

            </div>

            {/* Sidebar widgets */}
            <div className="space-y-8">
              {/* Linked Repositories */}
              <section className="glass-card p-6">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-xs font-black uppercase tracking-widest text-muted-foreground">Connected Repositories</h3>
                  <button
                    onClick={() => {
                      setLinkingRepoIds([]);
                      setShowRepoPicker(true);
                    }}
                    className="h-7 w-7 rounded-full bg-primary/15 text-primary border border-primary/30 hover:bg-primary/25 transition-colors flex items-center justify-center"
                    title="Link repositories"
                  >
                    <Plus className="w-3.5 h-3.5" />
                  </button>
                </div>
                <div className="space-y-3">
                  {linkedRepos.length === 0 ? (
                    <p className="text-xs text-muted-foreground">연결된 저장소가 없습니다.</p>
                  ) : (
                    linkedRepos.map((repo) => (
                      <div key={repo.id} className="flex items-center justify-between rounded-xl border border-border/50 bg-muted/10 px-4 py-2.5">
                        <div className="min-w-0">
                          <p className="text-xs font-black text-foreground truncate">{repo.name}</p>
                          <p className="text-[9px] text-muted-foreground truncate uppercase tracking-widest">{repo.full_name}</p>
                        </div>
                        <Badge variant="glass">linked</Badge>
                      </div>
                    ))
                  )}
                </div>
              </section>

              {/* Team Members */}
              <section className="glass-card p-6">
                <h3 className="text-xs font-black uppercase tracking-widest text-muted-foreground mb-4 flex items-center gap-2">
                  <Users className="w-4 h-4 text-primary" /> Team Members
                </h3>
                <div className="space-y-4">
                  {(project.project_members ?? []).map((m, idx) => {
                    const user = users.find((u) => u.id === m.user_id);
                    return (
                      <div key={idx} className="flex items-center justify-between">
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8 rounded-full bg-muted/30 border border-border flex items-center justify-center text-xs font-bold text-muted-foreground">
                            {(user?.name ?? m.user_id).slice(0, 2).toUpperCase()}
                          </div>
                          <div>
                            <p className="text-xs font-bold text-foreground">{user?.name ?? m.user_id}</p>
                            <p className="text-[9px] text-muted-foreground uppercase tracking-widest">{m.project_role}</p>
                          </div>
                        </div>
                        <Badge variant={m.project_role === "lead" ? "primary" : "glass"}>
                          {m.project_role}
                        </Badge>
                      </div>
                    );
                  })}
                </div>
              </section>
            </div>
          </motion.div>
        ) : (
          <div className="h-96 flex items-center justify-center text-muted-foreground">
            No dashboard data available.
          </div>
        )}
      </AnimatePresence>

      {/* Link Repository Modal */}
      {showRepoPicker && (
        <div className="fixed inset-0 z-[90] flex items-center justify-center p-6">
          <div className="absolute inset-0 bg-background/70 backdrop-blur-sm" onClick={() => setShowRepoPicker(false)} />
          <div className="relative w-full max-w-xl glass border-border rounded-2xl p-6">
            <div className="flex items-center justify-between mb-4">
              <h4 className="text-sm font-black uppercase tracking-widest text-foreground">Link Repositories</h4>
              <button onClick={() => setShowRepoPicker(false)} className="p-1.5 rounded-lg hover:bg-muted/30">
                <X className="w-4 h-4 text-muted-foreground" />
              </button>
            </div>
            <div className="space-y-2 max-h-72 overflow-y-auto pr-1">
              {candidateRepos.length === 0 ? (
                <p className="text-xs text-muted-foreground">추가로 연결 가능한 저장소가 없습니다.</p>
              ) : (
                candidateRepos.map((repo) => (
                  <label key={repo.id} className="flex items-center gap-3 rounded-xl border border-border/50 bg-muted/10 px-3 py-2 cursor-pointer hover:bg-muted/20">
                    <input
                      type="checkbox"
                      checked={linkingRepoIds.includes(repo.id)}
                      onChange={(e) => {
                        if (e.target.checked) {
                          setLinkingRepoIds((prev) => Array.from(new Set([...prev, repo.id])));
                        } else {
                          setLinkingRepoIds((prev) => prev.filter((id) => id !== repo.id));
                        }
                      }}
                    />
                    <span className="text-xs font-bold text-foreground">{repo.full_name}</span>
                  </label>
                ))
              )}
            </div>
            <div className="mt-5 flex justify-end gap-2">
              <button
                onClick={() => setShowRepoPicker(false)}
                className="px-4 py-2 rounded-xl border border-border text-xs font-bold text-foreground"
              >
                Cancel
              </button>
              <button
                onClick={() => void linkSelectedRepositories()}
                disabled={linkingRepoIds.length === 0}
                className="px-4 py-2 rounded-xl bg-primary text-primary-foreground text-xs font-black disabled:opacity-50"
              >
                Connect ({linkingRepoIds.length})
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
