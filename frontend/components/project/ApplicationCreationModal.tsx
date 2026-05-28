"use client";

import { useState, useEffect } from "react";
import { motion } from "framer-motion";
import { X, Box, Info, Globe, Eye, Lock, Loader2, Calendar, GitBranch, FolderKanban } from "lucide-react";
import { Application, ApplicationStatus, ApplicationVisibility, Project } from "@/lib/services/project.types";
import { projectService } from "@/lib/services/project.service";
import { identityService } from "@/lib/services/identity.service";
import { repositoryService, Repository } from "@/lib/services/repository.service";
import { ComboBox } from "@/shared/ui-foundation/components/ComboBox";
import { cn } from "@/shared/utils";

interface ApplicationCreationModalProps {
  onClose: () => void;
  onCreated: (app: Application) => void;
  initialData?: Partial<Application>;
}

export function ApplicationCreationModal({ onClose, onCreated, initialData }: ApplicationCreationModalProps) {
  const [formData, setFormData] = useState({
    key: initialData?.key || "",
    name: initialData?.name || "",
    description: initialData?.description || "",
    leader_user_id: initialData?.leader_user_id || "",
    development_unit_id: initialData?.development_unit_id || "",
    visibility: initialData?.visibility || "internal" as ApplicationVisibility,
    status: initialData?.status || "planning" as ApplicationStatus,
    start_date: initialData?.start_date || "",
    due_date: initialData?.due_date || "",
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [leaderOptions, setLeaderOptions] = useState<Array<{ label: string; value: string; description?: string }>>([]);
  const [unitOptions, setUnitOptions] = useState<Array<{ label: string; value: string; description?: string }>>([]);

  // Projects + Repositories connection management states.
  // 후속 carve (post-#395/#396) — backend updateProject 가 `application_id` nullable
  // PATCH 를 지원하면서 양쪽 모두 편집 가능. allProjects = 본 application 에 이미 연결된
  // 것 + standalone (application_id IS NULL). selectedProjectIDs 는 체크된 set.
  const [allProjects, setAllProjects] = useState<Project[]>([]);
  const [selectedProjectIDs, setSelectedProjectIDs] = useState<string[]>([]);
  const [allRepositories, setAllRepositories] = useState<Repository[]>([]);
  const [selectedRepoKeys, setSelectedRepoKeys] = useState<string[]>([]); // repo_provider/repo_full_name

  const isEdit = !!initialData?.id;

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onClose]);

  useEffect(() => {
    let alive = true;
    (async () => {
      try {
        const [users, hierarchy, repos] = await Promise.all([
          identityService.getUsers(),
          identityService.getOrgHierarchy(),
          repositoryService.listRepositories()
        ]);
        if (!alive) return;
        setLeaderOptions(
          users.map((u) => ({
            label: u.name || u.id,
            value: u.id,
            description: u.email,
          })),
        );
        setUnitOptions(
          hierarchy.nodes.map((n) => ({
            label: n.data.label,
            value: n.id,
            description: n.data.type,
          })),
        );
        setAllRepositories(repos);

        // Fetch application repositories if in edit mode
        if (isEdit && initialData.id) {
          const appRepos = await projectService.getApplicationRepositories(initialData.id);
          if (alive) {
            setSelectedRepoKeys(appRepos.map(r => `${r.repo_provider}/${r.repo_full_name}`));
          }
        }
      } catch {
        if (!alive) return;
        setLeaderOptions([]);
        setUnitOptions([]);
      }
    })();
    return () => {
      alive = false;
    };
  }, [isEdit, initialData]);

  // edit 모드 — 본 application 에 이미 연결된 + standalone (application_id IS NULL)
  // 프로젝트를 합쳐 표시. backend updateProject 의 `application_id` nullable PATCH
  // 가 지원되므로 (post-#395/#396 carve) 체크박스로 연결/해제 가능.
  useEffect(() => {
    if (!isEdit || !initialData.id) {
      return;
    }
    let alive = true;
    (async () => {
      try {
        // codex P2 (#397 hotfix) — 기존 `listAllProjects([])` 는 빈 repository ID list
        // 라 빈 array 반환 → standalone candidates 가 0 이었음. 신규 `/projects/standalone`
        // endpoint (StandaloneOnly=true, application_id IS NULL) 로 정공법 fetch.
        const [appProjects, standaloneProjects] = await Promise.all([
          projectService.getApplicationProjectsV2(initialData.id!),
          projectService.listStandaloneProjects(),
        ]);
        if (!alive) return;
        // appProjects (이미 연결) + standaloneProjects (연결 가능) 합쳐 dedup.
        const seen = new Set<string>();
        const merged: Project[] = [];
        for (const p of [...appProjects, ...standaloneProjects]) {
          if (!seen.has(p.id)) {
            seen.add(p.id);
            merged.push(p);
          }
        }
        setAllProjects(merged);
        setSelectedProjectIDs(appProjects.map((p) => p.id));
      } catch (err) {
        if (alive) console.error(err);
      }
    })();
    return () => {
      alive = false;
    };
  }, [isEdit, initialData]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSubmitting(true);

    try {
      let result: Application;
      if (!formData.leader_user_id.trim()) {
        throw new Error("Application leader is required.");
      }
      if (!formData.development_unit_id.trim()) {
        throw new Error("Development department is required.");
      }
      if (isEdit && initialData.id) {
        // PATCH key is immutable
        const patchPayload: Partial<typeof formData> & { owner_user_id?: string } = { ...formData };
        delete patchPayload.key;
        patchPayload.owner_user_id = formData.leader_user_id;
        result = await projectService.updateApplication(initialData.id, patchPayload);

        // Synchronize connected repositories
        const currentAppRepos = await projectService.getApplicationRepositories(initialData.id);
        const currentRepoKeys = currentAppRepos.map(r => `${r.repo_provider}/${r.repo_full_name}`);

        const toAddRepos = selectedRepoKeys.filter(k => !currentRepoKeys.includes(k));
        const toRemoveRepos = currentRepoKeys.filter(k => !selectedRepoKeys.includes(k));

        // codex P2 정합 (#395) — repo full_name 이 `owner/repo` 형태 ("/" 다수 포함)
        // 일 때 `key.split("/")` destructure 가 두 번째 segment 만 fullName 으로 받아
        // 나머지를 drop 했음 (e.g. `github/acme/web` → provider="github", fullName="acme",
        // "web" 손실). 첫 `/` 위치 기준으로 정확히 둘로 split.
        const splitRepoKey = (key: string): [string, string] => {
          const i = key.indexOf("/");
          return i < 0 ? [key, ""] : [key.slice(0, i), key.slice(i + 1)];
        };
        await Promise.all([
          ...toAddRepos.map(key => {
            const [provider, fullName] = splitRepoKey(key);
            return projectService.connectRepository(initialData.id!, {
              repo_provider: provider,
              repo_full_name: fullName,
              role: "sub"
            });
          }),
          ...toRemoveRepos.map(key => {
            const [provider, fullName] = splitRepoKey(key);
            return projectService.disconnectRepository(initialData.id!, provider, fullName);
          })
        ]);

        // Connected Projects 동기화 — backend updateProject 가 `application_id` nullable
        // PATCH 를 지원하므로 (post-#395/#396 carve) toAdd/toRemove 를 PATCH 로 반영.
        // toAdd: 빈/다른 application_id 였던 project → 본 application 으로 이전
        // toRemove: 본 application 에서 해제 → application_id = "" (NULL)
        const connectedProjects = await projectService.getApplicationProjectsV2(initialData.id);
        const connectedIDs = connectedProjects.map((p) => p.id);
        const toAddProjects = selectedProjectIDs.filter((pid) => !connectedIDs.includes(pid));
        const toRemoveProjects = connectedIDs.filter((pid) => !selectedProjectIDs.includes(pid));
        await Promise.all([
          ...toAddProjects.map((pid) =>
            projectService.updateProject(pid, { application_id: initialData.id }),
          ),
          ...toRemoveProjects.map((pid) =>
            projectService.updateProject(pid, { application_id: "" }),
          ),
        ]);
      } else {
        const normalizedKey = formData.key.trim().toUpperCase();
        result = await projectService.createApplication({
          ...formData,
          key: normalizedKey,
          owner_user_id: formData.leader_user_id,
        });
      }
      onCreated(result);
      onClose();
    } catch (err) {
      if (
        typeof err === "object" &&
        err !== null &&
        "status" in err &&
        (err as { status?: number }).status === 409
      ) {
        setError("Application key already exists.");
      } else {
        setError(err instanceof Error ? err.message : "Failed to save application");
      }
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
        role="dialog"
        aria-modal="true"
      >
        <div className="p-8 border-b border-border flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-purple-500/20 rounded-xl flex items-center justify-center">
              <Box className="w-5 h-5 text-purple-400" />
            </div>
            <div>
              <h2 className="text-xl font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">
                {isEdit ? "Edit" : "Create"} <span className="text-purple-400">Application</span>
              </h2>
              <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
                Governance container for repositories & projects
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
              <label htmlFor="appKey" className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Application Key</label>
              <div className="relative group">
                <Info className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-foreground/35 dark:text-primary-foreground/20 group-focus-within:text-purple-400 transition-colors" />
                <input
                  id="appKey"
                  required
                  disabled={isEdit}
                  value={formData.key}
                  onChange={e => setFormData({ ...formData, key: e.target.value.toUpperCase() })}
                  placeholder="E.G. PLATFORM26"
                  maxLength={10}
                  pattern="^[A-Za-z0-9]{1,10}$"
                  className={cn(
                    "w-full bg-muted/30 border border-border rounded-2xl pl-12 pr-4 py-3 text-sm font-mono text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-purple-400/50 uppercase",
                    isEdit && "opacity-50 cursor-not-allowed"
                  )}
                />
              </div>
              {!isEdit && <p className="text-[9px] text-muted-foreground px-1 italic">Immutable 1-10 char alphanumeric ID.</p>}
            </div>
            <div className="space-y-2">
              <label htmlFor="appName" className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Display Name</label>
              <input
                id="appName"
                required
                value={formData.name}
                onChange={e => setFormData({ ...formData, name: e.target.value })}
                placeholder="e.g. DevHub Platform 2026"
                className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-purple-400/50"
              />
            </div>
          </div>

          <div className="space-y-2">
            <label htmlFor="appDesc" className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Description</label>
            <textarea
              id="appDesc"
              value={formData.description}
              onChange={e => setFormData({ ...formData, description: e.target.value })}
              placeholder="Strategic goals, KPI, and scope summary..."
              rows={3}
              className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-purple-400/50 resize-none"
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Application Leader</label>
              {leaderOptions.length > 0 ? (
                <ComboBox
                  options={leaderOptions}
                  value={formData.leader_user_id}
                  onChange={(value) => setFormData({ ...formData, leader_user_id: value })}
                  placeholder="Search leader by name/email/user_id"
                  emptyText="No matching leaders."
                  className="w-full"
                />
              ) : (
                <input
                  required
                  value={formData.leader_user_id}
                  onChange={(e) => setFormData({ ...formData, leader_user_id: e.target.value })}
                  placeholder="e.g. charlie"
                  className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-purple-400/50"
                />
              )}
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Development Department</label>
              {unitOptions.length > 0 ? (
                <ComboBox
                  options={unitOptions}
                  value={formData.development_unit_id}
                  onChange={(value) => setFormData({ ...formData, development_unit_id: value })}
                  placeholder="Search department/unit by name or id"
                  emptyText="No matching departments."
                  className="w-full"
                />
              ) : (
                <input
                  required
                  value={formData.development_unit_id}
                  onChange={(e) => setFormData({ ...formData, development_unit_id: e.target.value })}
                  placeholder="e.g. dept-eng"
                  className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-purple-400/50"
                />
              )}
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-1 gap-6">
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
                        ? "bg-purple-500/10 border-purple-500/40 text-purple-400 shadow-lg shadow-purple-500/5"
                        : "bg-muted/20 border-border/60 text-muted-foreground hover:bg-muted/40"
                    )}
                  >
                    {v === 'public' && <Globe className="w-3 h-3" />}
                    {v === 'internal' && <Eye className="w-3 h-3" />}
                    {v === 'restricted' && <Lock className="w-3 h-3" />}
                    {v}
                  </button>
                ))}
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Operating Status</label>
              <select
                value={formData.status}
                onChange={e => setFormData({ ...formData, status: e.target.value as ApplicationStatus })}
                className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-purple-400/50 appearance-none"
              >
                <option value="planning">Planning</option>
                <option value="active">Active</option>
                <option value="on_hold">On Hold</option>
                <option value="closed">Closed</option>
                <option value="archived">Archived</option>
              </select>
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Period (Optional)</label>
              <div className="flex items-center gap-2">
                <div className="relative flex-1">
                  <Calendar className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground/40" />
                  <input
                    type="date"
                    value={formData.start_date}
                    onChange={e => setFormData({ ...formData, start_date: e.target.value })}
                    className="w-full bg-muted/20 border border-border/40 rounded-xl pl-10 pr-3 py-2 text-xs text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-purple-400/50"
                  />
                </div>
                <span className="text-muted-foreground/40">→</span>
                <div className="relative flex-1">
                  <Calendar className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground/40" />
                  <input
                    type="date"
                    value={formData.due_date}
                    onChange={e => setFormData({ ...formData, due_date: e.target.value })}
                    className="w-full bg-muted/20 border border-border/40 rounded-xl pl-10 pr-3 py-2 text-xs text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-purple-400/50"
                  />
                </div>
              </div>
            </div>
          </div>

          {isEdit && (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 border-t border-border/60 pt-6">
              <div className="space-y-3">
                <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1 flex items-center gap-1.5">
                  <GitBranch className="w-3.5 h-3.5 text-purple-400" /> Connected Repositories
                </label>
                <div className="max-h-[160px] overflow-y-auto border border-border rounded-2xl p-4 bg-muted/5 space-y-2 custom-scrollbar">
                  {allRepositories.map(repo => {
                    const key = `${repo.provider_key}/${repo.full_name}`;
                    const isChecked = selectedRepoKeys.includes(key);
                    return (
                      <label key={key} className="flex items-center gap-3 text-xs text-foreground dark:text-primary-foreground cursor-pointer select-none">
                        <input
                          type="checkbox"
                          checked={isChecked}
                          onChange={(e) => {
                            if (e.target.checked) setSelectedRepoKeys([...selectedRepoKeys, key]);
                            else setSelectedRepoKeys(selectedRepoKeys.filter(k => k !== key));
                          }}
                          className="h-4 w-4 rounded border-border"
                        />
                        <span>{repo.full_name} ({repo.provider_key})</span>
                      </label>
                    );
                  })}
                  {allRepositories.length === 0 && (
                    <p className="text-[10px] text-muted-foreground italic">No repositories available.</p>
                  )}
                </div>
              </div>

              <div className="space-y-3">
                <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1 flex items-center gap-1.5">
                  <FolderKanban className="w-3.5 h-3.5 text-purple-400" /> Connected Projects
                </label>
                <div className="max-h-[160px] overflow-y-auto border border-border rounded-2xl p-4 bg-muted/5 space-y-2 custom-scrollbar">
                  {allProjects.map((proj) => {
                    const isChecked = selectedProjectIDs.includes(proj.id);
                    return (
                      <label key={proj.id} className="flex items-center gap-3 text-xs text-foreground dark:text-primary-foreground cursor-pointer select-none">
                        <input
                          type="checkbox"
                          checked={isChecked}
                          onChange={(e) => {
                            if (e.target.checked) setSelectedProjectIDs([...selectedProjectIDs, proj.id]);
                            else setSelectedProjectIDs(selectedProjectIDs.filter((id) => id !== proj.id));
                          }}
                          className="h-4 w-4 rounded border-border"
                        />
                        <span>{proj.name} ({proj.key})</span>
                      </label>
                    );
                  })}
                  {allProjects.length === 0 && (
                    <p className="text-[10px] text-muted-foreground italic">No projects available to connect.</p>
                  )}
                </div>
                <p className="text-[10px] text-muted-foreground italic px-1">
                  Standalone (unattached) and already-connected projects only — projects under other applications must be moved via their own edit modal.
                </p>
              </div>
            </div>
          )}

          {error && (
            <motion.div 
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              className="p-4 bg-accent/10 border border-accent/20 rounded-2xl text-[11px] text-accent font-medium"
            >
              {error}
            </motion.div>
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
              {submitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <>{isEdit ? 'Save Changes' : 'Create Application'}</>}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  );
}
