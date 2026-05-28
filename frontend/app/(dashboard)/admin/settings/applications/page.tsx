"use client";

import { useEffect, useState, useMemo, useCallback } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Plus } from "lucide-react";
import { useRouter } from "next/navigation";
import { projectService } from "@/lib/services/project.service";
import { Application, ApplicationRepository, Project } from "@/lib/services/project.types";
import { ApplicationTable } from "@/components/project/ApplicationTable";
import { ApplicationCreationModal } from "@/components/project/ApplicationCreationModal";
import { RepositoryTable } from "@/components/project/RepositoryTable";
import { ProjectTable } from "@/components/project/ProjectTable";
import { RepositoryLinkModal } from "@/components/project/RepositoryLinkModal";
import { ProjectCreationModal } from "@/components/project/ProjectCreationModal";
import { FilterBar } from "@/shared/ui-foundation/components/FilterBar";
import { useToast } from "@/shared/ui-foundation/components/Toast";

const STATUS_OPTIONS = [
  { label: "All Status", value: "all" },
  { label: "Active", value: "active" },
  { label: "Planning", value: "planning" },
  { label: "On Hold", value: "on_hold" },
  { label: "Closed", value: "closed" },
  { label: "Archived", value: "archived" },
];

export default function AdminSettingsApplicationsPage() {
  const router = useRouter();
  const [applications, setApplications] = useState<Application[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [query, setQuery] = useState("");
  const [activeStatus, setActiveStatus] = useState("all");
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingApp, setEditingApp] = useState<Application | null>(null);
  const [selectedApp, setSelectedApp] = useState<Application | null>(null);
  const [appRepos, setAppRepos] = useState<ApplicationRepository[]>([]);
  const [appProjects, setAppProjects] = useState<Project[]>([]);
  const [showRepoLinkModal, setShowRepoLinkModal] = useState(false);
  const [showProjectModal, setShowProjectModal] = useState(false);
  const [editingProject, setEditingProject] = useState<Project | null>(null);
  const { toast } = useToast();

  const filteredApplications = useMemo(() => {
    const q = query.trim().toLowerCase();
    return applications.filter((app) => {
      const matchesQuery = !q || 
        app.name.toLowerCase().includes(q) ||
        app.key.toLowerCase().includes(q) ||
        app.description.toLowerCase().includes(q);
      
      const matchesStatus = activeStatus === "all" || app.status === activeStatus;
      
      return matchesQuery && matchesStatus;
    });
  }, [applications, query, activeStatus]);

  const refresh = useCallback(async (showLoading = true) => {
    if (showLoading) setIsLoading(true);
    try {
      const data = await projectService.getApplications();
      setApplications(data);
    } catch (error) {
      console.error("[admin/settings/applications] load failed:", error);
      if ((error as { status?: number })?.status === 501) {
        toast("Backend API not implemented yet (501). Showing empty list.", "warning");
      }
    } finally {
      setIsLoading(false);
    }
  }, [toast]);

  useEffect(() => {
    const timer = setTimeout(() => {
      void refresh(false);
    }, 0);
    return () => clearTimeout(timer);
  }, [refresh]);

  const handleCreate = () => {
    setEditingApp(null);
    setShowCreateModal(true);
  };

  const handleEdit = (app: Application) => {
    setEditingApp(app);
    setShowCreateModal(true);
  };

  // archived 상태 → hard-delete (permanent), 그 외 → archive (soft-delete).
  const handleArchive = async (app: Application) => {
    const isHard = app.status === "archived";
    const msg = isHard
      ? `Permanently delete archived application "${app.name}"? This cannot be undone.`
      : `Are you sure you want to archive ${app.name}?`;
    if (!confirm(msg)) return;
    try {
      await projectService.archiveApplication(app.id, isHard);
      toast(`Application ${app.name} ${isHard ? "permanently deleted" : "archived"}`, "success");
      refresh();
    } catch {
      toast(isHard ? "Failed to delete application" : "Failed to archive application", "error");
    }
  };

  const loadAppChildren = useCallback(async (app: Application) => {
    setSelectedApp(app);
    // codex review (#312 P1-2) — Promise.all 통합 시 v2 projects fail (legacy
    // mode 410 등) 이 repositories load 도 같이 reject. 두 호출 독립 처리:
    // repositories 는 legacy/v2 무관, projects v2 fail 시 빈 array fallback.
    const reposPromise = projectService.getApplicationRepositories(app.id).catch((err) => {
      console.error("[admin/settings/applications] repositories load failed:", err);
      toast("Failed to load repositories for selected application", "error");
      return [] as ApplicationRepository[];
    });
    const projectsPromise = projectService.getApplicationProjectsV2(app.id).catch((err) => {
      // DEVHUB_PROJECT_MODEL=legacy 시 v2 route 410 gone. 빈 array fallback.
      console.warn("[admin/settings/applications] v2 projects load skipped (likely legacy mode):", err);
      return [] as Project[];
    });
    const [repos, projects] = await Promise.all([reposPromise, projectsPromise]);
    setAppRepos(repos);
    setAppProjects(projects);
  }, [toast]);

  const handleDisconnectRepo = useCallback(async (repo: ApplicationRepository) => {
    if (!selectedApp) return;
    try {
      await projectService.disconnectRepository(selectedApp.id, repo.repo_provider, repo.repo_full_name);
      toast(`Disconnected ${repo.repo_full_name}`, "success");
      await loadAppChildren(selectedApp);
    } catch {
      toast("Failed to disconnect repository", "error");
    }
  }, [selectedApp, toast, loadAppChildren]);

  return (
    <div className="space-y-8">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="flex-1">
          <FilterBar 
            onSearch={setQuery}
            onFilterChange={setActiveStatus}
            activeFilter={activeStatus}
            filterOptions={STATUS_OPTIONS}
            placeholder="Search by name, key, or description..."
          />
        </motion.div>

        <motion.button
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          whileHover={{ scale: 1.02 }}
          whileTap={{ scale: 0.98 }}
          onClick={handleCreate}
          className="bg-primary text-primary-foreground px-6 py-3.5 rounded-2xl font-black uppercase tracking-widest text-[10px] flex items-center gap-2 shadow-xl shadow-primary/20 transition-all border border-primary/20"
        >
          <Plus className="w-4 h-4" />
          New Application
        </motion.button>
      </div>

      {isLoading ? (
        <div className="flex flex-col items-center justify-center py-32 gap-4">
          <div className="w-12 h-12 border-4 border-primary/20 border-t-primary rounded-full animate-spin" />
          <p className="text-muted-foreground font-bold animate-pulse uppercase tracking-[0.3em] text-[10px]">Loading Applications...</p>
        </div>
      ) : (
        <ApplicationTable
          applications={filteredApplications}
          onEdit={handleEdit}
          onArchive={handleArchive}
          onViewRepositories={(app) => {
            void loadAppChildren(app);
          }}
        />
      )}

      {selectedApp && (
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground">
              Application Scope: {selectedApp.key}
            </h3>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setShowRepoLinkModal(true)}
                className="px-4 py-2 rounded-xl border border-border bg-muted/20 text-foreground dark:text-primary-foreground text-[10px] font-black uppercase tracking-widest hover:bg-muted/40"
              >
                Link Repository
              </button>
              <button
                onClick={() => {
                  setEditingProject(null);
                  setShowProjectModal(true);
                }}
                className="px-4 py-2 rounded-xl bg-primary text-primary-foreground text-[10px] font-black uppercase tracking-widest hover:opacity-90"
              >
                New Project
              </button>
            </div>
          </div>

          <div className="space-y-4">
            <h4 className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Repositories</h4>
            <RepositoryTable
              repositories={appRepos}
              onDisconnect={handleDisconnectRepo}
              onViewRepository={(repo) => {
                if (typeof repo.repository_id !== "number" || repo.repository_id <= 0) {
                  toast("Repository detail is unavailable for this link.", "warning");
                  return;
                }
                router.push(`/repositories/${repo.repository_id}`);
              }}
              onViewRepositoryMetrics={(repo) => {
                if (typeof repo.repository_id !== "number" || repo.repository_id <= 0) {
                  toast("Repository metrics are unavailable for this link.", "warning");
                  return;
                }
                router.push(`/repositories/${repo.repository_id}`);
              }}
            />
          </div>

          <div className="space-y-4">
            <h4 className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Projects</h4>
            <ProjectTable
              projects={appProjects}
              onViewDetails={(project) => {
                router.push(`/projects/${project.id}`);
              }}
              onEditProject={(project) => {
                setEditingProject(project);
                setShowProjectModal(true);
              }}
            />
          </div>
        </div>
      )}

      <AnimatePresence>
        {showCreateModal && (
          <ApplicationCreationModal
            initialData={editingApp || undefined}
            onClose={() => setShowCreateModal(false)}
            onCreated={(newApp) => {
              toast(`Application ${newApp.name} ${editingApp ? 'updated' : 'created'}`, "success");
              refresh();
            }}
          />
        )}
        {showRepoLinkModal && selectedApp && (
          <RepositoryLinkModal
            applicationId={selectedApp.id}
            onClose={() => setShowRepoLinkModal(false)}
            onLinked={() => {
              setShowRepoLinkModal(false);
              void loadAppChildren(selectedApp);
            }}
          />
        )}
        {showProjectModal && selectedApp && (
          <ProjectCreationModal
            applicationId={selectedApp.id}
            repositories={appRepos}
            initialData={editingProject ?? undefined}
            onClose={() => {
              setShowProjectModal(false);
              setEditingProject(null);
            }}
            onCreated={(project) => {
              toast(`Project ${project.name} ${editingProject ? "updated" : "created"}`, "success");
              setShowProjectModal(false);
              setEditingProject(null);
              void loadAppChildren(selectedApp);
            }}
          />
        )}
      </AnimatePresence>
    </div>
  );
}
