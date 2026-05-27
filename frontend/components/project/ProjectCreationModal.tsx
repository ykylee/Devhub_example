"use client";

import { useState, useEffect } from "react";
import { motion } from "framer-motion";
import { X, FolderKanban, Loader2, GitBranch } from "lucide-react";
import { ApplicationRepository, Project, ApplicationVisibility, ProjectStatus, Application, SCMProvider } from "@/lib/services/project.types";
import { Repository } from "@/lib/services/repository.service";
import { projectService } from "@/lib/services/project.service";
import { cn } from "@/lib/utils";

interface ProjectCreationModalProps {
  applicationId?: string;
  repositories: (ApplicationRepository | Repository)[];
  onClose: () => void;
  onCreated: (project: Project) => void;
  initialData?: Partial<Project>;
}

export function ProjectCreationModal({ applicationId, repositories, onClose, onCreated, initialData }: ProjectCreationModalProps) {
  const [applications, setApplications] = useState<Application[]>([]);
  const [scmProviders, setScmProviders] = useState<SCMProvider[]>([]);
  const numericRepositories = repositories.map((r) => {
    // ApplicationRepository.repository_id 가 optional 이라 `"repository_id" in r`
    // 로는 narrow 불가. ApplicationRepository required field `repo_provider` 로
    // discriminate (Repository 에는 없음).
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
  }).filter((r) => typeof r.repository_id === "number" && r.repository_id > 0);

  const initialRepositoryId =
    initialData?.repository_id || numericRepositories[0]?.repository_id || 0;

  const [formData, setFormData] = useState({
    key: initialData?.key || "",
    name: initialData?.name || "",
    description: initialData?.description || "",
    owner_user_id: initialData?.owner_user_id || "",
    visibility: initialData?.visibility || "internal" as ApplicationVisibility,
    status: initialData?.status || "planning" as ProjectStatus,
    start_date: initialData?.start_date || "",
    due_date: initialData?.due_date || "",
    repository_id: initialData?.repository_id || initialRepositoryId || 0,
    repository_ids: initialData?.repository_ids || (initialRepositoryId ? [initialRepositoryId] : ([] as number[])),
    application_id: initialData?.application_id || applicationId || "",
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
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
    projectService.getSCMProviders().then((providers) => {
      setScmProviders(providers);
      const enabled = providers.find((p) => p.enabled);
      if (enabled) {
        setRepositoryCreate((prev) => ({ ...prev, scm_provider: enabled.provider_key }));
      }
    }).catch(() => setScmProviders([]));
  }, []);

  const isEdit = !!initialData?.id;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSubmitting(true);

    try {
      let result: Project;
      if (isEdit && initialData.id) {
        // PATCH 시 `key` 는 백엔드(updateProject)에서 project_key_immutable 로 reject 되므로
        // payload 에서 제외한다. (codex PR #114 review P1, Application 과 동일 정합)
        const patchPayload: Partial<typeof formData> = { ...formData };
        delete patchPayload.key;
        result = await projectService.updateProject(initialData.id, patchPayload);
      } else {
        const selected = new Set<number>();
        if (formData.repository_id) selected.add(formData.repository_id);
        for (const id of formData.repository_ids) selected.add(id);
        const repository_ids = Array.from(selected);
        const selectedApplicationId = formData.application_id || applicationId || "";

        if (selectedApplicationId) {
          const payload = {
            ...formData,
            application_id: selectedApplicationId,
            repository_ids,
            repository_create_payload: createRepository ? repositoryCreate : undefined,
          };
          result = await projectService.createApplicationProject(selectedApplicationId, payload);
        } else {
          result = await projectService.createProjectStandalone({
            ...formData,
            repository_ids,
            repository_id: repository_ids[0] || 0,
            application_id: "",
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
                onChange={e => setFormData({ ...formData, key: e.target.value.toUpperCase() })}
                placeholder="E.G. API-V1"
                className={cn(
                  "w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm font-mono text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-indigo-400/50 uppercase",
                  isEdit && "opacity-50 cursor-not-allowed"
                )}
              />
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Display Name</label>
              <input
                required
                value={formData.name}
                onChange={e => setFormData({ ...formData, name: e.target.value })}
                placeholder="e.g. Backend Refactoring"
                className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-indigo-400/50"
              />
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Application (Optional)</label>
            <select
              disabled={isEdit}
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

          <div className="space-y-2">
            <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Target Repository</label>
            <div className="relative group">
              <GitBranch className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground/40" />
              <select
                disabled={isEdit}
                value={formData.repository_id}
                onChange={e => {
                  const nextPrimary = Number(e.target.value);
                  const existing = formData.repository_ids.filter((id) => id !== nextPrimary);
                  setFormData({
                    ...formData,
                    repository_id: nextPrimary,
                    repository_ids: [nextPrimary, ...existing],
                  });
                }}
                className="w-full bg-muted/30 border border-border rounded-2xl pl-12 pr-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-indigo-400/50 appearance-none"
              >
                <option value={0}>No repository (connect later)</option>
                {numericRepositories.map((repo) => (
                    <option
                      key={`${repo.repo_provider}/${repo.repo_full_name}`}
                      value={repo.repository_id}
                      className="bg-slate-900"
                    >
                      {repo.repo_full_name} ({repo.repo_provider})
                    </option>
                  ))}
              </select>
            </div>
            <p className="text-[9px] text-accent/60 px-1 italic">선택 시 primary repository로 저장됩니다. 없어도 생성 가능합니다.</p>
            {!isEdit && numericRepositories.length > 1 && (
              <div className="mt-3 p-3 rounded-xl border border-border/60 bg-muted/20">
                <p className="text-[9px] font-black uppercase tracking-widest text-muted-foreground mb-2">
                  Additional Linked Repositories (N:M)
                </p>
                <div className="space-y-2 max-h-36 overflow-y-auto pr-1">
                  {numericRepositories.map((repo) => {
                    const repoId = repo.repository_id as number;
                    const checked = formData.repository_ids.includes(repoId);
                    return (
                      <label
                        key={`multi-${repo.repo_provider}-${repo.repo_full_name}`}
                        className="flex items-center gap-2 text-xs text-foreground dark:text-primary-foreground"
                      >
                        <input
                          type="checkbox"
                          checked={checked}
                          onChange={(e) => {
                            if (e.target.checked) {
                              const merged = Array.from(new Set([...formData.repository_ids, repoId]));
                              setFormData({
                                ...formData,
                                repository_ids: merged,
                                repository_id: formData.repository_id || repoId,
                              });
                              return;
                            }
                            const next = formData.repository_ids.filter((id) => id !== repoId);
                            setFormData({
                              ...formData,
                              repository_ids: next,
                              repository_id:
                                formData.repository_id === repoId
                                  ? (next[0] ?? 0)
                                  : formData.repository_id,
                            });
                          }}
                          className="accent-indigo-400"
                        />
                        <span>
                          {repo.repo_full_name} ({repo.repo_provider})
                        </span>
                      </label>
                    );
                  })}
                </div>
              </div>
            )}
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
                    {scmProviders.filter((p) => p.enabled).map((p) => (
                      <option key={p.provider_key} value={p.provider_key}>
                        {p.display_name} ({p.provider_key})
                      </option>
                    ))}
                  </select>
                </div>
              )}
            </div>
          )}

          {!isEdit && (
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Additional Repositories (N:M)</label>
              <div className="grid grid-cols-1 gap-2 max-h-44 overflow-y-auto rounded-2xl border border-border/50 bg-muted/10 p-3">
                {numericRepositories.map((repo) => {
                  const id = repo.repository_id as number;
                  const checked = formData.repository_ids.includes(id);
                  return (
                    <label key={`repo-link-${id}`} className="flex items-center gap-3 text-xs text-foreground dark:text-primary-foreground">
                      <input
                        type="checkbox"
                        checked={checked}
                        onChange={(e) => {
                          const next = new Set(formData.repository_ids);
                          if (e.target.checked) next.add(id);
                          else next.delete(id);
                          setFormData({ ...formData, repository_ids: Array.from(next) });
                        }}
                        className="h-4 w-4 rounded border-border"
                      />
                      <span>{repo.repo_full_name} ({repo.repo_provider})</span>
                    </label>
                  );
                })}
              </div>
            </div>
          )}

          <div className="space-y-2">
            <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Description</label>
            <textarea
              value={formData.description}
              onChange={e => setFormData({ ...formData, description: e.target.value })}
              placeholder="Scope and deliverables..."
              rows={3}
              className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-indigo-400/50 resize-none"
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Owner</label>
              <input
                required
                value={formData.owner_user_id}
                onChange={e => setFormData({ ...formData, owner_user_id: e.target.value })}
                placeholder="User ID..."
                className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-indigo-400/50"
              />
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Status</label>
              <select
                value={formData.status}
                onChange={e => setFormData({ ...formData, status: e.target.value as ProjectStatus })}
                className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-indigo-400/50 appearance-none"
              >
                <option value="planning" className="bg-slate-900">Planning</option>
                <option value="active" className="bg-slate-900">Active</option>
                <option value="on_hold" className="bg-slate-900">On Hold</option>
              </select>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Visibility</label>
              <div className="grid grid-cols-3 gap-2">
                {(['public', 'internal', 'restricted'] as ApplicationVisibility[]).map((v) => (
                  <button
                    key={v}
                    type="button"
                    onClick={() => setFormData({ ...formData, visibility: v })}
                    className={cn(
                      "py-2.5 rounded-xl border text-[10px] font-black uppercase tracking-widest transition-all flex flex-col items-center gap-1",
                      formData.visibility === v
                        ? "bg-indigo-500/10 border-indigo-500/40 text-indigo-400 shadow-lg shadow-indigo-500/5"
                        : "bg-muted/20 border-border/60 text-muted-foreground hover:bg-muted/40"
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
                    onChange={e => setFormData({ ...formData, start_date: e.target.value })}
                    className="w-full bg-muted/20 border border-border/40 rounded-xl px-3 py-2 text-xs text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-indigo-400/50"
                  />
                  <span className="text-muted-foreground/40">→</span>
                  <input
                    type="date"
                    value={formData.due_date}
                    onChange={e => setFormData({ ...formData, due_date: e.target.value })}
                    className="w-full bg-muted/20 border border-border/40 rounded-xl px-3 py-2 text-xs text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-indigo-400/50"
                  />
               </div>
            </div>
          </div>

          {error && (
            <div className="p-4 bg-accent/10 border border-accent/20 rounded-2xl text-[11px] text-accent font-medium">
              {error}
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
              disabled={submitting}
              className="flex-1 bg-primary text-primary-foreground font-black py-4 px-8 rounded-2xl hover:scale-[1.02] active:scale-[0.98] transition-all shadow-xl disabled:opacity-50 uppercase tracking-widest text-[10px] flex items-center justify-center gap-2"
            >
              {submitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <>{isEdit ? 'Save Changes' : 'Create Project'}</>}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  );
}
