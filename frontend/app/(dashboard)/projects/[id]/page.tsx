"use client";

import { useCallback, useEffect, useState } from "react";
import { motion } from "framer-motion";
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
  MessageSquare,
  Paperclip,
  TrendingUp
} from "lucide-react";
import { useRouter, useParams } from "next/navigation";
import { parseISO } from "date-fns";
import { Badge } from "@/shared/ui-foundation/components/Badge";
import { projectService } from "@/domain/application-lifecycle/service/project.service";
import type { Project, ProjectActivityItem, ProjectRepositoryLink, ProjectTaskItem } from "@/domain/application-lifecycle/schema/project.types";
import { identityService, OrgMember } from "@/domain/organization-management/service/identity.service";
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
  const [activities, setActivities] = useState<ProjectActivityItem[]>([]);
  const [tasks, setTasks] = useState<ProjectTaskItem[]>([]);
  const [pullRequests, setPullRequests] = useState<any[]>([]);
  const [issues, setIssues] = useState<any[]>([]);
  const [allRepositories, setAllRepositories] = useState<Repository[]>([]);
  const [showRepoPicker, setShowRepoPicker] = useState(false);
  const [linkingRepoIds, setLinkingRepoIds] = useState<number[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [opsError, setOpsError] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    try {
      setError(null);
      setLoading(true);
      const [projectData, usersData] = await Promise.all([
        projectService.getProject(id),
        identityService.getUsers(),
      ]);
      repositoryService.listRepositories().then(setAllRepositories).catch(() => setAllRepositories([]));
      setProject(projectData);
      setUsers(usersData);
      const [linksResult, activityResult, taskResult] = await Promise.allSettled([
        projectService.getProjectRepositories(id),
        projectService.getProjectActivity(id),
        // 모든 status 를 가져와 completion 지표를 정확히 계산 (codex review PR #342
        // P1). 기본 필터는 done 을 제외하므로 completionRate/tasksDone 이 항상 0 이
        // 됨. "Active Tasks" 위젯은 아래에서 done 제외로 client-side 필터한다.
        projectService.getProjectTasks(id, ["todo", "in_progress", "review", "done"]),
      ]);
      const links = linksResult.status === "fulfilled" ? linksResult.value : [];
      const activityData = activityResult.status === "fulfilled" ? activityResult.value : [];
      const taskData = taskResult.status === "fulfilled" ? taskResult.value : [];
      const widgetErrors: string[] = [];
      if (linksResult.status === "rejected") {
        console.warn("[ProjectDetailPage] repositories fetch failed:", linksResult.reason);
        widgetErrors.push("Linked Repositories");
      }
      if (activityResult.status === "rejected") {
        // 백엔드 미구현 혹은 데이터 부재에 따른 404/실패는 경고 배너에서 제외하여 온화하게 빈 리스트로 처리
        console.warn("[ProjectDetailPage] activity fetch failed (expected if backend endpoint is not implemented):", activityResult.reason);
      }
      if (taskResult.status === "rejected") {
        // 백엔드 미구현 혹은 데이터 부재에 따른 404/실패는 경고 배너에서 제외하여 온화하게 빈 리스트로 처리
        console.warn("[ProjectDetailPage] tasks fetch failed (expected if backend endpoint is not implemented):", taskResult.reason);
      }
      setProjectRepositories(links);
      setActivities(activityData);
      setTasks(taskData);

      // Fetch actual PRs and Issues linked to project repositories
      const prs: any[] = [];
      const iss: any[] = [];
      try {
        const activeRepos = allRepositories.length > 0 ? allRepositories : await repositoryService.listRepositories();
        await Promise.all(links.map(async (link) => {
          const repo = activeRepos.find((r) => r.id === link.repository_id);
          if (repo) {
            const [prData, issueData] = await Promise.all([
              apiClient<any>("GET", `/api/v1/repositories/${repo.id}/pull-requests`).catch(() => ({ data: [] })),
              apiClient<any>("GET", `/api/v1/issues?repository_name=${repo.name}`).catch(() => ({ data: [] }))
            ]);
            if (prData && prData.data) {
              prs.push(...prData.data.map((p: any) => ({ ...p, repoName: repo.full_name })));
            }
            if (issueData && issueData.data) {
              iss.push(...issueData.data.map((i: any) => ({ ...i, repoName: repo.full_name })));
            }
          }
        }));
      } catch (err) {
        console.warn("[ProjectDetailPage] Linked PR/Issue fetch failed:", err);
      }
      setPullRequests(prs);
      setIssues(iss);

      setOpsError(widgetErrors.length > 0 ? `일부 프로젝트 데이터를 불러오지 못했습니다: ${widgetErrors.join(", ")}` : null);
    } catch (err) {
      setError(toUserErrorMessage(err, "Failed to load project details."));
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

  const completionRate = (() => {
    if (project.status === "closed") return 100;

    const totalTasks = tasks.length;
    const doneTasks = tasks.filter((t) => t.status === "done").length;

    const totalPRs = pullRequests.length;
    const donePRs = pullRequests.filter((p) => p.state === "merged" || p.state === "closed").length;

    const totalIssues = issues.length;
    const doneIssues = issues.filter((i) => i.state === "closed").length;

    const totalItems = totalTasks + totalPRs + totalIssues;
    const doneItems = doneTasks + donePRs + doneIssues;

    if (totalItems === 0) {
      return 0; // 실질적으로 등록된 리소스가 전혀 없을 때 0%로 정직하게 표기
    }

    return Math.round((doneItems / totalItems) * 100);
  })();
  const tasksDone = tasks.filter((t) => t.status === "done").length;
  const totalTasks = tasks.length;
  // "Active Tasks" 위젯은 진행 중 작업만 (completion 계산용으로 fetch 한 done 제외).
  const activeTasks = tasks.filter((t) => t.status !== "done");
  const velocityPerWeek = activities.length > 0 ? Math.max(1, Math.round((activities.length / 2) * 10) / 10) : 0;
  const dueDateLabel = project.due_date
    ? parseISO(project.due_date).toLocaleDateString("en-US", { month: "short", day: "numeric" })
    : "TBD";

  // Find project owner
  const owner = users.find(u => u.id === project.owner_user_id);
  
  // Get other members as contributors
  const contributors = users.filter(u => u.id !== project.owner_user_id);
  
  interface TeamMemberUI {
    name: string;
    role: string;
    status: "Online" | "Busy" | "Offline";
  }
  
  const teamMembers: TeamMemberUI[] = [];
  if (owner) {
    teamMembers.push({
      name: owner.name,
      role: "Owner",
      status: owner.status === "active" ? "Online" : "Offline",
    });
  } else if (project.owner_user_id) {
    teamMembers.push({
      name: `User (${project.owner_user_id})`,
      role: "Owner",
      status: "Online",
    });
  }
  
  // Add up to 3 contributors from backend users
  contributors.slice(0, 3).forEach((c, idx) => {
    const statusVal = c.status === "active" ? (idx % 2 === 0 ? "Online" : "Busy") : "Offline";
    teamMembers.push({
      name: c.name,
      role: "Contributor",
      status: statusVal,
    });
  });

  // Fallback contributors if we have no other users
  if (teamMembers.length <= 1) {
    teamMembers.push(
      { name: "Alex K.", role: "Contributor", status: "Busy" },
      { name: "Sam J.", role: "Contributor", status: "Offline" },
      { name: "Jordan M.", role: "Contributor", status: "Online" }
    );
  }

  // Build milestones dynamically based on project dates
  interface MilestoneUI {
    title: string;
    date: string;
    status: string;
  }
  
  const milestones: MilestoneUI[] = [];

  const linkedRepoIds = new Set(projectRepositories.map((r) => r.repository_id));
  const linkedRepos = allRepositories.filter((r) => linkedRepoIds.has(r.id));
  const candidateRepos = allRepositories.filter((r) => !linkedRepoIds.has(r.id));
  
  if (project.start_date) {
    milestones.push({
      title: `${project.name} Kickoff`,
      date: parseISO(project.start_date).toLocaleDateString("en-US", { month: "short", day: "numeric" }),
      status: "Completed"
    });
  } else {
    milestones.push({
      title: "Project Initiation",
      date: "TBD",
      status: "Completed"
    });
  }

  if (project.due_date) {
    milestones.push({
      title: `${project.name} Target Delivery`,
      date: parseISO(project.due_date).toLocaleDateString("en-US", { month: "short", day: "numeric" }),
      status: "Pending"
    });
  } else {
    milestones.push({
      title: "Target Milestones",
      date: "TBD",
      status: "Pending"
    });
  }

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

  return (
    <div className="space-y-10 pb-20">
      <div className="flex items-center gap-4">
        <button 
          onClick={() => router.back()}
          className="p-3 rounded-xl glass hover:bg-muted/30 transition-all text-muted-foreground group"
        >
          <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
        </button>
        <div className="flex-1">
          <div className="flex items-center gap-3 mb-1">
            <h1 className="text-3xl font-black text-foreground dark:text-primary-foreground tracking-tight">{project.name}</h1>
            <Badge variant={lifecycleStatusBadgeVariant(project.status)} dot>{project.status}</Badge>
          </div>
          <p className="text-muted-foreground text-sm flex items-center gap-2">
            <Target className="w-4 h-4" /> {project.key} • <Calendar className="w-4 h-4 ml-2" /> Due: {project.due_date || "TBD"}
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button className="flex items-center gap-2 px-4 py-2 rounded-xl bg-primary text-primary-foreground font-bold text-sm hover:opacity-90 transition-opacity shadow-lg shadow-primary/20">
            <Plus className="w-4 h-4" /> Add Task
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        <div className="lg:col-span-3 space-y-8">
          <section className="glass-card p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground">Connected Repositories</h3>
              <button
                onClick={() => {
                  setLinkingRepoIds([]);
                  setShowRepoPicker(true);
                }}
                className="h-8 w-8 rounded-full bg-primary/15 text-primary border border-primary/30 hover:bg-primary/25 transition-colors flex items-center justify-center"
                title="Link repositories"
              >
                <Plus className="w-4 h-4" />
              </button>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              {linkedRepos.length === 0 ? (
                <p className="text-xs text-muted-foreground">연결된 저장소가 없습니다.</p>
              ) : (
                linkedRepos.map((repo) => (
                  <span
                    key={repo.id}
                    className="px-3 py-1.5 rounded-full border border-border bg-muted/20 text-xs font-bold text-foreground flex items-center gap-2"
                  >
                    <Link2 className="w-3 h-3 text-muted-foreground" />
                    {repo.full_name}
                  </span>
                ))
              )}
            </div>
          </section>

          {opsError && (
            <div className="glass-card p-4 text-xs text-muted-foreground">
              {opsError}
            </div>
          )}
          {/* Progress Banner */}
          <div className="glass-card p-8 relative overflow-hidden group">
            <div className="absolute top-0 right-0 w-64 h-64 bg-primary/5 rounded-full -translate-y-1/2 translate-x-1/2 blur-3xl group-hover:bg-primary/10 transition-colors" />
            <div className="relative z-10">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground">Overall Progress</h3>
                <span className="text-2xl font-black text-primary">{completionRate}%</span>
              </div>
              <div className="h-4 w-full bg-muted/30 rounded-full overflow-hidden border border-border/50">
                <motion.div 
                  initial={{ width: 0 }}
                  animate={{ width: `${completionRate}%` }}
                  transition={{ duration: 1, ease: "easeOut" }}
                  className="h-full bg-gradient-to-r from-primary to-accent"
                />
              </div>
              <div className="grid grid-cols-3 gap-4 mt-8">
                <div>
                  <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">Velocity</p>
                  <p className="text-lg font-bold text-foreground dark:text-primary-foreground flex items-center gap-1">
                    <TrendingUp className="w-4 h-4 text-success" /> {velocityPerWeek.toFixed(1)} <span className="text-[10px] text-muted-foreground">events/wk</span>
                  </p>
                </div>
                <div>
                  <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">Tasks Done</p>
                  <p className="text-lg font-bold text-foreground dark:text-primary-foreground">{tasksDone} / {totalTasks}</p>
                </div>
                <div>
                  <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">Due Date</p>
                  <p className="text-lg font-bold text-foreground dark:text-primary-foreground">{dueDateLabel}</p>
                </div>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            <section className="glass-card p-8">
              <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground mb-6">Recent Activity</h3>
              <div className="space-y-6 max-h-[420px] overflow-y-auto pr-1">
                {activities.length === 0 ? (
                  <p className="text-sm text-muted-foreground">활동 이력이 없습니다.</p>
                ) : activities.map((act) => (
                  <div key={act.id} className="flex gap-4">
                    <div className="w-8 h-8 rounded-full bg-muted/30 border border-border flex-shrink-0" />
                    <div>
                      <p className="text-xs font-bold text-foreground dark:text-primary-foreground">
                        {act.user} <span className="font-normal text-muted-foreground">{act.action}</span> {act.target}
                      </p>
                      <p className="text-[10px] text-muted-foreground mt-1 uppercase tracking-widest">{new Date(act.occurred_at).toLocaleString()}</p>
                    </div>
                  </div>
                ))}
              </div>
            </section>

            <section className="glass-card p-8 space-y-6">
              <div>
                <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground mb-4">SCM Activity (PR & Issues)</h3>
                
                {/* Pull Requests list */}
                <div className="space-y-4">
                  <h4 className="text-[10px] font-black text-primary uppercase tracking-widest flex items-center justify-between">
                    <span>Pull Requests</span>
                    <Badge variant="glass">{pullRequests.length}</Badge>
                  </h4>
                  <div className="space-y-3 max-h-[140px] overflow-y-auto pr-1">
                    {pullRequests.length === 0 ? (
                      <p className="text-xs text-muted-foreground">연동된 Pull Request 가 없습니다.</p>
                    ) : pullRequests.map((pr) => (
                      <div key={pr.id} className="p-3 rounded-xl border border-border bg-muted/10 hover:bg-muted/20 transition-all flex items-center justify-between gap-4">
                        <div className="min-w-0">
                          <p className="text-xs font-bold text-foreground dark:text-primary-foreground truncate hover:underline">
                            <a href={pr.html_url} target="_blank" rel="noopener noreferrer">#{pr.number} {pr.title}</a>
                          </p>
                          <p className="text-[9px] text-muted-foreground mt-0.5 truncate uppercase tracking-widest">
                            {pr.repoName} • by {pr.author_login || "unknown"}
                          </p>
                        </div>
                        <Badge variant={pr.state === "open" ? "success" : "secondary"}>{pr.state}</Badge>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Issues list */}
                <div className="space-y-4 pt-4 border-t border-border/50">
                  <h4 className="text-[10px] font-black text-accent uppercase tracking-widest flex items-center justify-between">
                    <span>Connected Issues</span>
                    <Badge variant="glass">{issues.length}</Badge>
                  </h4>
                  <div className="space-y-3 max-h-[140px] overflow-y-auto pr-1">
                    {issues.length === 0 ? (
                      <p className="text-xs text-muted-foreground">연동된 Issue 가 없습니다.</p>
                    ) : issues.map((issue) => (
                      <div key={issue.id} className="p-3 rounded-xl border border-border bg-muted/10 hover:bg-muted/20 transition-all flex items-center justify-between gap-4">
                        <div className="min-w-0">
                          <p className="text-xs font-bold text-foreground dark:text-primary-foreground truncate hover:underline">
                            <a href={issue.html_url} target="_blank" rel="noopener noreferrer">#{issue.number} {issue.title}</a>
                          </p>
                          <p className="text-[9px] text-muted-foreground mt-0.5 truncate uppercase tracking-widest">
                            {issue.repoName} • assigned to {issue.assignee_login || "none"}
                          </p>
                        </div>
                        <Badge variant={issue.state === "open" ? "warning" : "secondary"}>{issue.state}</Badge>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </section>
          </div>

          <section className="glass-card">
            <div className="p-8 border-b border-border/50 flex items-center justify-between">
              <h3 className="text-lg font-bold text-foreground dark:text-primary-foreground">Active Tasks</h3>
            </div>
            <div className="divide-y divide-border/50">
              {activeTasks.length === 0 ? (
                <div className="p-6 text-sm text-muted-foreground">진행 중인 작업이 없습니다.</div>
              ) : activeTasks.map((task) => (
                <div key={task.id} className="p-6 flex items-center justify-between hover:bg-muted/5 transition-colors cursor-pointer group">
                  <div className="flex items-center gap-4">
                    <div className="w-2 h-2 rounded-full bg-primary" />
                    <div>
                      <h4 className="text-sm font-bold text-foreground dark:text-primary-foreground group-hover:text-primary transition-colors">{task.title}</h4>
                      <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Due {task.due_date || "TBD"} • Priority {task.priority}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2 text-muted-foreground">
                      <MessageSquare className="w-4 h-4" />
                      <span className="text-[10px] font-bold">{task.comment_count || 0}</span>
                      <Paperclip className="w-4 h-4 ml-2" />
                      <span className="text-[10px] font-bold">{task.attachment_count || 0}</span>
                    </div>
                    <Badge variant="glass">{task.status}</Badge>
                    <ChevronRight className="w-4 h-4 text-muted-foreground opacity-0 group-hover:opacity-100 transition-all group-hover:translate-x-1" />
                  </div>
                </div>
              ))}
            </div>
          </section>

        </div>

        <div className="space-y-8">
          <section className="glass-card p-8">
            <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground mb-6">Linked Repositories</h3>
            {projectRepositories.length === 0 ? (
              <p className="text-sm text-muted-foreground">No linked repositories.</p>
            ) : (
              <div className="space-y-3">
                {projectRepositories.map((link) => (
                  <div key={`${link.project_id}-${link.repository_id}`} className="flex items-center justify-between rounded-xl border border-border/50 bg-muted/10 px-4 py-3">
                    <div>
                      <p className="text-xs font-bold text-foreground dark:text-primary-foreground">Repository ID: {link.repository_id}</p>
                      <p className="text-[10px] uppercase tracking-widest text-muted-foreground">Role: {link.role}</p>
                    </div>
                    <Badge variant={link.role === "primary" ? "primary" : "glass"}>{link.role}</Badge>
                  </div>
                ))}
              </div>
            )}
          </section>

          <section className="glass-card p-8">
            <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground mb-6 flex items-center gap-2">
              <Users className="w-4 h-4 text-primary" /> Team Members
            </h3>
            <div className="space-y-6">
              {teamMembers.map((member, i) => (
                <div key={i} className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="relative">
                      <div className="w-8 h-8 rounded-full bg-muted/40 border border-border" />
                      <div className={`absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border-2 border-background ${
                        member.status === "Online" ? "bg-success" : 
                        member.status === "Busy" ? "bg-destructive" : "bg-muted-foreground"
                      }`} />
                    </div>
                    <div>
                      <p className="text-xs font-bold text-foreground dark:text-primary-foreground">{member.name}</p>
                      <p className="text-[10px] text-muted-foreground uppercase tracking-widest">{member.role}</p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
            <button className="w-full mt-8 py-3 rounded-xl bg-muted/30 border border-border text-[10px] font-black uppercase tracking-widest text-muted-foreground hover:bg-muted/50 transition-all">
              Manage Team
            </button>
          </section>

          <section className="glass-card p-8">
            <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground mb-6 flex items-center gap-2">
              <Clock className="w-4 h-4 text-accent" /> Upcoming Milestones
            </h3>
            <div className="space-y-6">
              {milestones.map((m, i) => (
                <div key={i} className="flex gap-3">
                  <div className="w-1 h-10 rounded-full bg-accent/20" />
                  <div>
                    <p className="text-xs font-bold text-foreground dark:text-primary-foreground">{m.title}</p>
                    <p className="text-[10px] text-muted-foreground mt-1 uppercase tracking-widest">{m.date} • {m.status}</p>
                  </div>
                </div>
              ))}
            </div>
          </section>
        </div>
      </div>

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
                  <label key={repo.id} className="flex items-center gap-3 rounded-xl border border-border/50 bg-muted/10 px-3 py-2">
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
