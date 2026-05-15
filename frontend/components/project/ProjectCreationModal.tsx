"use client";

import { useState } from "react";
import { motion } from "framer-motion";
import { X, FolderKanban, Loader2, GitBranch } from "lucide-react";
import { ApplicationRepository, Project, ApplicationVisibility, ProjectStatus } from "@/lib/services/project.types";
import { projectService } from "@/lib/services/project.service";
import { cn } from "@/lib/utils";

interface ProjectCreationModalProps {
  applicationId: string;
  repositories: ApplicationRepository[];
  onClose: () => void;
  onCreated: (project: Project) => void;
  initialData?: Partial<Project>;
}

export function ProjectCreationModal({ applicationId, repositories, onClose, onCreated, initialData }: ProjectCreationModalProps) {
  const [formData, setFormData] = useState({
    key: initialData?.key || "",
    name: initialData?.name || "",
    description: initialData?.description || "",
    owner_user_id: initialData?.owner_user_id || "",
    visibility: initialData?.visibility || "internal" as ApplicationVisibility,
    status: initialData?.status || "planning" as ProjectStatus,
    start_date: initialData?.start_date || "",
    due_date: initialData?.due_date || "",
    repository_id: initialData?.repository_id || repositories.find((r) => typeof r.repository_id === "number")?.repository_id || 0,
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

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
        const { key: _omitKey, ...patchPayload } = formData;
        result = await projectService.updateProject(initialData.id, patchPayload);
      } else {
        // We need repository_id to create a project.
        if (!formData.repository_id) {
           throw new Error("A repository must be selected for the project.");
        }
        // Inject application_id if present
        const payload = { ...formData, application_id: applicationId };
        result = await projectService.createProject(formData.repository_id, payload);
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
                Operational unit under a repository
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
            <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Target Repository</label>
            <div className="relative group">
              <GitBranch className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground/40" />
              <select
                disabled={isEdit}
                value={formData.repository_id}
                onChange={e => setFormData({ ...formData, repository_id: Number(e.target.value) })}
                className="w-full bg-muted/30 border border-border rounded-2xl pl-12 pr-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-indigo-400/50 appearance-none"
              >
                <option value={0} disabled>Select a repository...</option>
                {repositories
                  .filter((repo) => typeof repo.repository_id === "number")
                  .map((repo) => (
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
            <p className="text-[9px] text-orange-400/60 px-1 italic">
              Note: only repositories with numeric `repository_id` can create projects.
            </p>
          </div>

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
            <div className="p-4 bg-orange-500/10 border border-orange-500/20 rounded-2xl text-[11px] text-orange-400 font-medium">
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
              className="flex-1 bg-indigo-600 text-white font-black py-4 px-8 rounded-2xl hover:scale-[1.02] active:scale-[0.98] transition-all shadow-xl shadow-indigo-500/20 disabled:opacity-50 uppercase tracking-widest text-[10px] flex items-center justify-center gap-2"
            >
              {submitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <>{isEdit ? 'Save Changes' : 'Create Project'}</>}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  );
}
