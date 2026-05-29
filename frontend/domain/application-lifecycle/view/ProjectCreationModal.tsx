"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { X, FolderKanban, GitBranch, Plus, Trash2, Users, Loader2 } from "lucide-react";
import {
  ApplicationRepository,
  Project,
  ApplicationVisibility,
  ProjectStatus,
  Application,
  SCMProvider,
  ProjectMemberRole,
} from "@/lib/services/project.types";
import { Repository } from "@/domain/repository-integration/service/repository.service";
import { projectService } from "@/domain/application-lifecycle/service/project.service";
import { identityService } from "@/domain/organization-management/service/identity.service";
import { ComboBox } from "@/shared/ui-foundation/components/ComboBox";
import { cn } from "@/shared/utils";
import { useStore } from "@/lib/store";

interface ProjectCreationModalProps {
  applicationId?: string;
  repositories: (ApplicationRepository | Repository)[];
  onClose: () => void;
  onCreated: (project: Project) => void;
  initialData?: Partial<Project>;
}

export function ProjectCreationModal({ applicationId, repositories, onClose, onCreated, initialData }: ProjectCreationModalProps) {
  const actor = useStore((s) => s.actor);
  // 기본 owner/leader = 현재 사용자의 canonical user_id. leaderOptions 가 user_id 로
  // 키잉되고 (identity.service.ts: `id: u.user_id`), login 은 user_id 와 다를 수 있어
  // (backend `/me` 가 login·user_id 를 별도 반환) 기본값으로 login 을 쓰면 ComboBox 매칭
  // 실패 + 잘못된 owner 저장. user_id 우선, 미해석 시 login fallback.
  const actorDefaultOwnerId = actor?.user_id || actor?.login || "";

  const [applications, setApplications] = useState<Application[]>([]);
  const [scmProviders, setScmProviders] = useState<SCMProvider[]>([]);
  type MemberRole = "leader" | "developer" | "reviewer" | "tester";
  type ProjectMemberDraft = { user_id: string; project_role: MemberRole };

  const numericRepositories = repositories
    .map((r) => {
      const isAppRepo = "repo_provider" in r;
      const repository_id = isAppRepo ? r.repository_id : r.id;
      const repo_full_name = isAppRepo ? r.repo_full_name : (r.full_name ?? r.name ?? "");
      const repo_provider = isAppRepo ? r.repo_provider : "github";
      return {
        ...r,
        repository_id,
        repo_full_name,
        repo_provider,
      };
    })
    .filter((r) => typeof r.repository_id === "number" && r.repository_id > 0);

  const initialRepositoryId = initialData?.repository_id || numericRepositories[0]?.repository_id || 0;

  const [formData, setFormData] = useState({
    key: initialData?.key || "",
    name: initialData?.name || "",
    description: initialData?.description || "",
    owner_user_id: initialData?.owner_user_id || actorDefaultOwnerId || "",
    visibility: (initialData?.visibility || "internal") as ApplicationVisibility,
    status: (initialData?.status || "planning") as ProjectStatus,
    start_date: initialData?.start_date || "",
    due_date: initialData?.due_date || "",
    repository_id: initialData?.repository_id || initialRepositoryId || 0,
    repository_ids: initialData?.repository_ids || (initialRepositoryId ? [initialRepositoryId] : ([] as number[])),
    application_id: initialData?.application_id || applicationId || "",
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [leaderOptions, setLeaderOptions] = useState<Array<{ label: string; value: string; description?: string }>>([]);
  const [repositoryLinks, setRepositoryLinks] = useState<number[]>(
    initialData?.repository_ids?.length
      ? initialData.repository_ids.filter((id): id is number => typeof id === "number" && id > 0)
      : initialRepositoryId
        ? [initialRepositoryId]
        : [],
  );
  const [projectMembers, setProjectMembers] = useState<ProjectMemberDraft[]>([
    { user_id: initialData?.owner_user_id || actorDefaultOwnerId || "", project_role: "leader" },
  ]);
  const [createRepository, setCreateRepository] = useState(false);
  const [repositoryCreate, setRepositoryCreate] = useState({
    key: "",
    slug: "",
    scm_provider: "",
  });

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onClose]);

  useEffect(() => {
    projectService.getApplications().then(setApplications).catch(() => setApplications([]));
    projectService
      .getSCMProviders()
      .then((providers) => {
        setScmProviders(providers);
        const enabled = providers.find((p) => p.enabled);
        if (enabled) {
          setRepositoryCreate((prev) => ({ ...prev, scm_provider: enabled.provider_key }));
        }
      })
      .catch(() => setScmProviders([]));
    identityService
      .getUsers()
      .then((users) => {
        setLeaderOptions(
          users.map((u) => ({
            label: u.name || u.id,
            value: u.id,
            description: u.email,
          })),
        );
      })
      .catch(() => setLeaderOptions([]));
  }, []);

  const isEdit = !!initialData?.id;

  const normalizedMembers = projectMembers.filter((m) => m.user_id.trim());
  const hasLeadMember = normalizedMembers.some((m) => m.project_role === "leader");
  const selectedRepositoryIDs = Array.from(new Set(repositoryLinks.filter((id) => id > 0)));
  const isCreateRepositoryPayloadValid =
    !createRepository ||
    (repositoryCreate.key.trim().length > 0 &&
      repositoryCreate.slug.trim().length > 0 &&
      repositoryCreate.scm_provider.trim().length > 0);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSubmitting(true);

    try {
      let result: Project;
      if (isEdit && initialData?.id) {
        const patchPayload: Partial<typeof formData> = { ...formData };
        delete patchPayload.key;
        // P1-#5 정정 — repository_id=0 fallback 은 backend createProjectRequest.RepositoryID
        // (int64 FK to repositories.id) 가 0 을 받으면 FK violation 또는 silent ignore.
        // 선택된 repo 가 있을 때만 명시 set, 빈 배열이면 두 필드 모두 미전송 (Partial 의도
        // 살림 — backend updateProjectRequest 가 미포함 필드는 갱신 skip).
        if (selectedRepositoryIDs.length > 0) {
          patchPayload.repository_ids = selectedRepositoryIDs;
          patchPayload.repository_id = selectedRepositoryIDs[0];
        } else {
          delete patchPayload.repository_ids;
          delete patchPayload.repository_id;
        }
        result = await projectService.updateProject(initialData.id, patchPayload);

        // N:M project-repositories sync during edit mode
        const existingLinks = await projectService.getProjectRepositories(initialData.id);
        const existingIDs = existingLinks.map((l) => l.repository_id);
        
        // Find links to remove
        const toRemove = existingIDs.filter((id) => !selectedRepositoryIDs.includes(id));
        // Find links to add
        const toAdd = selectedRepositoryIDs.filter((id) => !existingIDs.includes(id));

        await Promise.all([
          ...toRemove.map((id) => projectService.unlinkProjectRepository(initialData.id!, id)),
          ...toAdd.map((id) => {
            const role = selectedRepositoryIDs.indexOf(id) === 0 ? "primary" : "linked";
            return projectService.linkProjectRepository(initialData.id!, id, role);
          })
        ]);
      } else {
        const leader = normalizedMembers.find((m) => m.project_role === "leader")?.user_id.trim() || formData.owner_user_id.trim();
        const selectedApplicationId = formData.application_id || applicationId || "";
        const project_members: Array<{ user_id: string; project_role: ProjectMemberRole }> = normalizedMembers.map((m) => ({
          user_id: m.user_id.trim(),
          project_role: m.project_role === "leader" ? "lead" : "contributor",
        }));

        if (selectedApplicationId) {
          result = await projectService.createApplicationProject(selectedApplicationId, {
            ...formData,
            owner_user_id: leader,
            application_id: selectedApplicationId,
            repository_ids: selectedRepositoryIDs,
            repository_id: selectedRepositoryIDs[0] || 0,
            project_members,
            repository_create_payload: createRepository ? repositoryCreate : undefined,
          });
        } else {
          result = await projectService.createProjectStandalone({
            ...formData,
            owner_user_id: leader,
            repository_ids: selectedRepositoryIDs,
            repository_id: selectedRepositoryIDs[0] || 0,
            application_id: "",
            project_members,
            repository_create_payload: createRepository ? repositoryCreate : undefined,
          });
        }
      }
      onCreated(result);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save project");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-6">
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        onClick={onClose}
        className="absolute inset-0 bg-background/80 backdrop-blur-sm"
      />

      <motion.div
        initial={{ opacity: 0, scale: 0.95, y: 20 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.95, y: 20 }}
        role="dialog"
        aria-modal="true"
        className="relative w-full max-w-2xl glass border-border rounded-3xl shadow-2xl overflow-hidden"
      >
        <div className="p-8 border-b border-border flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-indigo-500/20 rounded-xl flex items-center justify-center">
              <FolderKanban className="w-5 h-5 text-indigo-400" />
            </div>
            <div>
              <h2 className="text-xl font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">
                {isEdit ? "Edit" : "Create"} <span className="text-indigo-400">Project</span>
              </h2>
              <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
                Independent unit; repository/application are optional
              </p>
            </div>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-muted/30 rounded-xl text-muted-foreground transition-colors">
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-8 space-y-6 max-h-[75vh] overflow-y-auto custom-scrollbar">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Project Key</label>
              <input
                required
                disabled={isEdit}
                value={formData.key}
                onChange={(e) => setFormData({ ...formData, key: e.target.value.toUpperCase() })}
                placeholder="E.G. API-V1"
                className={cn(
                  "w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm font-mono text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-indigo-400/50 uppercase",
                  isEdit && "opacity-50 cursor-not-allowed",
                )}
              />
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Display Name</label>
              <input
                required
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="e.g. Backend Refactoring"
                className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-indigo-400/50"
              />
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Application (Optional)</label>
            <select
              value={formData.application_id}
              onChange={(e) => setFormData({ ...formData, application_id: e.target.value })}
              className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-indigo-400/50 appearance-none"
            >
              <option value="">No application (independent)</option>
              {applications.map((app) => (
                <option key={app.id} value={app.id}>
                  {app.name} ({app.key})
                </option>
              ))}
            </select>
          </div>

          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Repositories</label>
              <button
                type="button"
                onClick={() => setRepositoryLinks((prev) => [...prev, 0])}
                className="inline-flex items-center gap-1 rounded-lg border border-border px-2 py-1 text-[10px] font-black uppercase tracking-widest hover:bg-muted/30"
              >
                <Plus className="w-3 h-3" /> Add
              </button>
            </div>
            {repositoryLinks.length === 0 && (
              <p className="text-[11px] text-muted-foreground">No repository selected. You can create the project first and connect later.</p>
            )}
            <div className="space-y-2">
              {repositoryLinks.map((repoID, idx) => (
                <div key={`repo-link-row-${idx}`} className="flex items-center gap-2">
                  <div className="relative flex-1">
                    <GitBranch className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground/40" />
                    <select
                      value={repoID}
                      onChange={(e) => {
                        const next = [...repositoryLinks];
                        next[idx] = Number(e.target.value);
                        setRepositoryLinks(next);
                      }}
                      className="w-full bg-muted/30 border border-border rounded-xl pl-10 pr-3 py-2 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-indigo-400/50 appearance-none"
                    >
                      <option value={0}>Select repository</option>
                      {numericRepositories
                        .filter((repo) => {
                          const repositoryID = repo.repository_id ?? 0;
                          return !repositoryLinks.includes(repositoryID) || repositoryID === repoID;
                        })
                        .map((repo) => {
                          const repositoryID = repo.repository_id ?? 0;
                          return (
                            <option key={`${idx}-${repo.repo_provider}/${repo.repo_full_name}`} value={repositoryID} className="bg-slate-900">
                              {repo.repo_full_name} ({repo.repo_provider})
                            </option>
                          );
                        })}
                    </select>
                  </div>
                  <button
                    type="button"
                    onClick={() => setRepositoryLinks((prev) => prev.filter((_, i) => i !== idx))}
                    className="h-9 w-9 rounded-lg border border-border text-muted-foreground hover:text-destructive hover:border-destructive/40 flex items-center justify-center"
                    aria-label="Remove repository"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              ))}
            </div>
            <p className="text-[9px] text-accent/60 px-1 italic">첫 번째 repository가 primary로 저장됩니다.</p>
          </div>

          {!isEdit && (
            <div className="space-y-3 rounded-2xl border border-border/60 bg-muted/10 p-4">
              <label className="flex items-center gap-2 text-xs text-foreground dark:text-primary-foreground">
                <input
                  type="checkbox"
                  checked={createRepository}
                  onChange={(e) => setCreateRepository(e.target.checked)}
                  className="h-4 w-4 rounded border-border"
                />
                <span>Create and link repository on project creation</span>
              </label>
              {createRepository && (
                <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                  <input
                    required
                    value={repositoryCreate.key}
                    onChange={(e) => setRepositoryCreate({ ...repositoryCreate, key: e.target.value.toUpperCase() })}
                    placeholder="Repo Key"
                    className="w-full bg-muted/30 border border-border rounded-xl px-3 py-2 text-xs text-foreground dark:text-primary-foreground"
                  />
                  <input
                    required
                    value={repositoryCreate.slug}
                    onChange={(e) => setRepositoryCreate({ ...repositoryCreate, slug: e.target.value })}
                    placeholder="org/repo-slug"
                    className="w-full bg-muted/30 border border-border rounded-xl px-3 py-2 text-xs text-foreground dark:text-primary-foreground"
                  />
                  <select
                    required
                    value={repositoryCreate.scm_provider}
                    onChange={(e) => setRepositoryCreate({ ...repositoryCreate, scm_provider: e.target.value })}
                    className="w-full bg-muted/30 border border-border rounded-xl px-3 py-2 text-xs text-foreground dark:text-primary-foreground"
                  >
                    {scmProviders
                      .filter((p) => p.enabled)
                      .map((p) => (
                        <option key={p.provider_key} value={p.provider_key}>
                          {p.display_name} ({p.provider_key})
                        </option>
                      ))}
                  </select>
                </div>
              )}
            </div>
          )}

          <div className="space-y-2">
            <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Description</label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder="Scope and deliverables..."
              rows={3}
              className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-indigo-400/50 resize-none"
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Project Leader</label>
              {leaderOptions.length > 0 ? (
                <ComboBox
                  options={leaderOptions}
                  value={formData.owner_user_id}
                  onChange={(nextLeader) => {
                    setFormData({ ...formData, owner_user_id: nextLeader });
                    setProjectMembers((prev) => {
                      const next = [...prev];
                      const leadIndex = next.findIndex((m) => m.project_role === "leader");
                      if (leadIndex >= 0) next[leadIndex] = { ...next[leadIndex], user_id: nextLeader };
                      else next.unshift({ user_id: nextLeader, project_role: "leader" });
                      return next;
                    });
                  }}
                  placeholder="Search leader by name/email/user_id"
                  emptyText="No matching users."
                  className="w-full"
                />
              ) : (
                <input
                  required
                  value={formData.owner_user_id}
                  onChange={(e) => {
                    const nextLeader = e.target.value;
                    setFormData({ ...formData, owner_user_id: nextLeader });
                    setProjectMembers((prev) => {
                      const next = [...prev];
                      const leadIndex = next.findIndex((m) => m.project_role === "leader");
                      if (leadIndex >= 0) next[leadIndex] = { ...next[leadIndex], user_id: nextLeader };
                      else next.unshift({ user_id: nextLeader, project_role: "leader" });
                      return next;
                    });
                  }}
                  placeholder="User ID..."
                  className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-indigo-400/50"
                />
              )}
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Status</label>
              <select
                value={formData.status}
                onChange={(e) => setFormData({ ...formData, status: e.target.value as ProjectStatus })}
                className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-indigo-400/50 appearance-none"
              >
                <option value="planning">Planning</option>
                <option value="active">Active</option>
                <option value="on_hold">On Hold</option>
                <option value="closed">Closed</option>
                <option value="archived">Archived</option>
              </select>
            </div>
          </div>

          <div className="space-y-3 rounded-2xl border border-border/60 bg-muted/10 p-4">
            <div className="flex items-center justify-between">
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest flex items-center gap-2">
                <Users className="w-3.5 h-3.5" /> Project Members
              </p>
              <button
                type="button"
                onClick={() => setProjectMembers((prev) => [...prev, { user_id: "", project_role: "developer" }])}
                className="inline-flex items-center gap-1 rounded-lg border border-border px-2 py-1 text-[10px] font-black uppercase tracking-widest hover:bg-muted/30"
              >
                <Plus className="w-3 h-3" /> Add
              </button>
            </div>
            <div className="space-y-2">
              {projectMembers.map((member, idx) => (
                <div key={`member-row-${idx}`} className="grid grid-cols-12 gap-2 items-center">
                  <div className="col-span-7">
                    {leaderOptions.length > 0 ? (
                      <ComboBox
                        options={leaderOptions}
                        value={member.user_id}
                        onChange={(value) => {
                          const next = [...projectMembers];
                          next[idx] = { ...next[idx], user_id: value };
                          setProjectMembers(next);
                          if (next[idx].project_role === "leader") {
                            setFormData((prev) => ({ ...prev, owner_user_id: value }));
                          }
                        }}
                        placeholder="Search member by name/email/user_id"
                        emptyText="No matching users."
                        className="w-full"
                      />
                    ) : (
                      <input
                        value={member.user_id}
                        onChange={(e) => {
                          const next = [...projectMembers];
                          next[idx] = { ...next[idx], user_id: e.target.value };
                          setProjectMembers(next);
                          if (next[idx].project_role === "leader") {
                            setFormData((prev) => ({ ...prev, owner_user_id: e.target.value }));
                          }
                        }}
                        placeholder="user id"
                        className="w-full bg-muted/30 border border-border rounded-xl px-3 py-2 text-xs text-foreground dark:text-primary-foreground"
                      />
                    )}
                  </div>
                  <select
                    value={member.project_role}
                    onChange={(e) => {
                      const role = e.target.value as MemberRole;
                      let next = [...projectMembers];
                      if (role === "leader") {
                        next = next.map((m, i) => ({
                          ...m,
                          project_role: i === idx ? "leader" : m.project_role === "leader" ? "developer" : m.project_role,
                        }));
                      } else {
                        next[idx] = { ...next[idx], project_role: role };
                      }
                      setProjectMembers(next);
                      if (role === "leader") {
                        setFormData((prev) => ({ ...prev, owner_user_id: next[idx].user_id }));
                      }
                    }}
                    className="col-span-4 bg-muted/30 border border-border rounded-xl px-2 py-2 text-xs text-foreground dark:text-primary-foreground"
                  >
                    <option value="leader">Leader</option>
                    <option value="developer">Developer</option>
                    <option value="reviewer">Reviewer</option>
                    <option value="tester">Tester</option>
                  </select>
                  <button
                    type="button"
                    onClick={() => {
                      if (projectMembers.length === 1) return;
                      const target = projectMembers[idx];
                      let next = projectMembers.filter((_, i) => i !== idx);
                      if (target.project_role === "leader") {
                        next = next.map((m, i) => ({ ...m, project_role: i === 0 ? "leader" : m.project_role }));
                        setFormData((prev) => ({ ...prev, owner_user_id: next[0]?.user_id ?? "" }));
                      }
                      setProjectMembers(next);
                    }}
                    disabled={projectMembers.length === 1}
                    className="col-span-1 h-8 w-8 rounded-lg border border-border text-muted-foreground hover:text-destructive hover:border-destructive/40 flex items-center justify-center"
                    aria-label="Remove member"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </button>
                </div>
              ))}
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Visibility</label>
              <div className="grid grid-cols-3 gap-2">
                {(["public", "internal", "restricted"] as ApplicationVisibility[]).map((v) => (
                  <button
                    key={v}
                    type="button"
                    onClick={() => setFormData({ ...formData, visibility: v })}
                    className={cn(
                      "py-2.5 rounded-xl border text-[10px] font-black uppercase tracking-widest transition-all flex flex-col items-center gap-1",
                      formData.visibility === v
                        ? "bg-indigo-500/10 border-indigo-500/40 text-indigo-400 shadow-lg shadow-indigo-500/5"
                        : "bg-muted/20 border-border/60 text-muted-foreground hover:bg-muted/40",
                    )}
                  >
                    {v}
                  </button>
                ))}
              </div>
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Period (Optional)</label>
              <div className="flex items-center gap-2">
                <input
                  type="date"
                  value={formData.start_date}
                  onChange={(e) => setFormData({ ...formData, start_date: e.target.value })}
                  className="w-full bg-muted/20 border border-border/40 rounded-xl px-3 py-2 text-xs text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-indigo-400/50"
                />
                <span className="text-muted-foreground/40">→</span>
                <input
                  type="date"
                  value={formData.due_date}
                  onChange={(e) => setFormData({ ...formData, due_date: e.target.value })}
                  className="w-full bg-muted/20 border border-border/40 rounded-xl px-3 py-2 text-xs text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-indigo-400/50"
                />
              </div>
            </div>
          </div>

          {error && <div className="p-4 bg-accent/10 border border-accent/20 rounded-2xl text-[11px] text-accent font-medium">{error}</div>}

          {!hasLeadMember && (
            <div className="p-3 bg-warning/10 border border-warning/20 rounded-2xl text-[11px] text-warning font-medium">
              Project Leader를 선택해야 Create Project 버튼이 활성화됩니다.
            </div>
          )}

          <div className="flex gap-4 pt-4 border-t border-border/60">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 glass border-border text-foreground dark:text-primary-foreground font-bold py-4 rounded-2xl hover:bg-muted/30 transition-all uppercase tracking-widest text-[10px]"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting || !hasLeadMember || !isCreateRepositoryPayloadValid}
              className="flex-1 bg-primary text-primary-foreground font-black py-4 px-8 rounded-2xl hover:scale-[1.02] active:scale-[0.98] transition-all shadow-xl disabled:opacity-50 uppercase tracking-widest text-[10px] flex items-center justify-center gap-2"
            >
              {submitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <>{isEdit ? "Save Changes" : "Create Project"}</>}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  );
}
