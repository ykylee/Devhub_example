"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { AnimatePresence } from "framer-motion";
import { useSearchParams, useRouter } from "next/navigation";
import { Box, FolderGit2, FolderKanban } from "lucide-react";
import Link from "next/link";
import { PageEmpty, PageError, PageLoading } from "@/components/ui/PageState";
import { toUserErrorMessage } from "@/lib/services/error-message";
import { applicationService, Application as AdminApplication } from "@/lib/services/application.service";
import { repositoryService, Repository } from "@/lib/services/repository.service";
import { projectService } from "@/lib/services/project.service";
import { ApplicationStatus, ApplicationVisibility, Project } from "@/lib/services/project.types";
import { ApplicationCreationModal } from "@/components/project/ApplicationCreationModal";
import { ProjectCreationModal } from "@/components/project/ProjectCreationModal";
import { useToast } from "@/components/ui/Toast";
import { cn } from "@/lib/utils";

type CatalogTab = "applications" | "repositories" | "projects";

function parseTab(raw: string | null): CatalogTab {
  if (raw === "applications" || raw === "repositories" || raw === "projects") return raw;
  return "applications";
}

export default function AdminCatalogPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const activeTab = parseTab(searchParams.get("tab"));
  const initialQuery = searchParams.get("q") ?? "";

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [query, setQuery] = useState(initialQuery);

  const [applications, setApplications] = useState<AdminApplication[]>([]);
  const [repositories, setRepositories] = useState<Repository[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [showApplicationModal, setShowApplicationModal] = useState(false);
  const [showProjectModal, setShowProjectModal] = useState(false);
  const [editingApplication, setEditingApplication] = useState<AdminApplication | null>(null);
  const [editingProject, setEditingProject] = useState<Project | null>(null);
  const [projectSeed, setProjectSeed] = useState<Partial<Project> | null>(null);
  const [creatingRepository, setCreatingRepository] = useState(false);
  const [publishingRepositoryID, setPublishingRepositoryID] = useState<number | null>(null);
  const { toast } = useToast();

  const loadAll = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const apps = await applicationService.listApplications();
      const repos = await repositoryService.listRepositories();

      const appScopedProjects = await Promise.all(
        apps.map((app) => projectService.getApplicationProjectsV2(app.id).catch(() => [])),
      );
      const repoScopedProjects = await Promise.all(
        repos.map((repo) => projectService.getRepositoryProjects(repo.id).catch(() => [])),
      );
      const allProjects = [...appScopedProjects.flat(), ...repoScopedProjects.flat()];
      const dedupProjects = Array.from(new Map(allProjects.map((p) => [p.id, p])).values());

      setApplications(apps);
      setRepositories(repos);
      setProjects(dedupProjects);
    } catch (err) {
      setError(toUserErrorMessage(err, "Admin Catalog 데이터를 불러오지 못했습니다."));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    // Initial load on mount. This effect intentionally kicks off async fetch.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void loadAll();
  }, [loadAll]);

  const filteredApplications = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return applications;
    return applications.filter((a) =>
      [a.key, a.name, a.owner_user_id, a.status].some((v) => (v ?? "").toLowerCase().includes(q)),
    );
  }, [applications, query]);

  const filteredRepositories = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return repositories;
    return repositories.filter((r) =>
      [r.full_name, r.owner_login, r.name, r.status, r.scm_provider ?? ""].some((v) => (v ?? "").toLowerCase().includes(q)),
    );
  }, [repositories, query]);

  const filteredProjects = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return projects;
    const numericQuery = Number(q);
    const hasNumericQuery = Number.isFinite(numericQuery) && q !== "";
    return projects.filter((p) =>
      [p.key, p.name, p.owner_user_id, p.status, p.application_id ?? ""].some((v) => String(v).toLowerCase().includes(q)) ||
      (hasNumericQuery && p.repository_id === numericQuery),
    );
  }, [projects, query]);

  const tabs: Array<{ key: CatalogTab; label: string; icon: React.ComponentType<{ className?: string }>; count: number }> = [
    { key: "applications", label: "Applications", icon: Box, count: filteredApplications.length },
    { key: "repositories", label: "Repositories", icon: FolderGit2, count: filteredRepositories.length },
    { key: "projects", label: "Projects", icon: FolderKanban, count: filteredProjects.length },
  ];

  const setTab = (tab: CatalogTab) => {
    const params = new URLSearchParams(searchParams.toString());
    params.set("tab", tab);
    // 탭 이동 시 이전 탭 검색어(q)로 인해 "데이터 없음"으로 보이는 혼선을 줄인다.
    params.delete("q");
    setQuery("");
    router.replace(`/admin/catalog?${params.toString()}`);
  };

  const appProjectCount = useMemo(() => {
    const counts = new Map<string, number>();
    for (const p of projects) {
      if (!p.application_id) continue;
      counts.set(p.application_id, (counts.get(p.application_id) ?? 0) + 1);
    }
    return counts;
  }, [projects]);

  const repoProjectCount = useMemo(() => {
    const counts = new Map<number, number>();
    for (const p of projects) {
      if (!p.repository_id) continue;
      counts.set(p.repository_id, (counts.get(p.repository_id) ?? 0) + 1);
    }
    return counts;
  }, [projects]);

  const applicationNameByID = useMemo(
    () => new Map(applications.map((app) => [app.id, app.name])),
    [applications],
  );

  const openProjectTabByApplication = (applicationID: string) => {
    const params = new URLSearchParams(searchParams.toString());
    params.set("tab", "projects");
    params.set("q", applicationID);
    setQuery(applicationID);
    router.replace(`/admin/catalog?${params.toString()}`);
  };

  const openProjectTabByRepository = (repositoryID: number) => {
    const q = String(repositoryID);
    const params = new URLSearchParams(searchParams.toString());
    params.set("tab", "projects");
    params.set("q", q);
    setQuery(q);
    router.replace(`/admin/catalog?${params.toString()}`);
  };

  const handleArchiveApplication = async (app: AdminApplication) => {
    if (!confirm(`Archive application ${app.name}?`)) return;
    try {
      await projectService.archiveApplication(app.id);
      toast(`Application ${app.name} archived`, "success");
      await loadAll();
    } catch (err) {
      toast(toUserErrorMessage(err, "Application 삭제에 실패했습니다."), "error");
    }
  };

  const handleArchiveProject = async (project: Project) => {
    if (!confirm(`Archive project ${project.name}?`)) return;
    try {
      await projectService.archiveProject(project.id);
      toast(`Project ${project.name} archived`, "success");
      await loadAll();
    } catch (err) {
      toast(toUserErrorMessage(err, "Project 삭제에 실패했습니다."), "error");
    }
  };

  const handleCreateRepositoryDraft = async () => {
    const key = prompt("Repository key를 입력하세요 (예: DEVHUB-API)");
    if (!key || !key.trim()) return;
    const slug = prompt("Repository slug(full_name)를 입력하세요 (예: devhub/devhub-api)");
    if (!slug || !slug.trim()) return;
    const scmProvider = prompt("SCM provider key를 입력하세요 (선택, 예: gitea)", "") ?? "";
    setCreatingRepository(true);
    try {
      await repositoryService.createRepositoryDraft({
        key: key.trim(),
        slug: slug.trim(),
        scm_provider: scmProvider.trim() || undefined,
      });
      toast(`Repository draft ${slug.trim()} 생성 완료`, "success");
      await loadAll();
    } catch (err) {
      toast(toUserErrorMessage(err, "Repository draft 생성에 실패했습니다."), "error");
    } finally {
      setCreatingRepository(false);
    }
  };

  const handleRequestPublish = async (repo: Repository) => {
    if (repo.status !== "draft") return;
    if (!confirm(`${repo.full_name} draft를 외부 SCM 연동 대상으로 전송할까요?`)) return;
    setPublishingRepositoryID(repo.id);
    try {
      await repositoryService.requestRepositoryPublish(repo.id);
      toast(`Publish 요청이 접수되었습니다: ${repo.full_name}`, "success");
      await loadAll();
    } catch (err) {
      toast(toUserErrorMessage(err, "Repository publish 요청에 실패했습니다."), "error");
    } finally {
      setPublishingRepositoryID(null);
    }
  };

  const editingApplicationInitialData = editingApplication
    ? {
        id: editingApplication.id,
        key: editingApplication.key,
        name: editingApplication.name,
        description: editingApplication.description,
        owner_user_id: editingApplication.owner_user_id,
        leader_user_id: editingApplication.leader_user_id,
        development_unit_id: editingApplication.development_unit_id,
        status: editingApplication.status as ApplicationStatus,
        visibility: editingApplication.visibility as ApplicationVisibility,
        start_date: editingApplication.start_date ?? undefined,
        due_date: editingApplication.due_date ?? undefined,
        archived_at: editingApplication.archived_at ?? undefined,
        created_at: editingApplication.created_at,
        updated_at: editingApplication.updated_at,
      }
    : undefined;

  if (loading) return <PageLoading label="Admin Catalog 로딩 중..." />;

  return (
    <div className="space-y-6">
      <div className="flex items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-black uppercase tracking-tight text-foreground dark:text-primary-foreground">
            Admin Catalog
          </h1>
          <p className="text-xs font-bold uppercase tracking-widest text-muted-foreground">
            system_admin 전용 통합 자산 관리
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => void loadAll()}
            className="rounded-xl border border-border px-3 py-2 text-xs font-bold uppercase tracking-widest hover:bg-muted/30"
          >
            Refresh
          </button>
          {activeTab === "applications" && (
            <button
              onClick={() => {
                setEditingApplication(null);
                setShowApplicationModal(true);
              }}
              className="rounded-xl bg-primary px-3 py-2 text-xs font-bold uppercase tracking-widest text-primary-foreground hover:opacity-90"
            >
              New Application
            </button>
          )}
          {activeTab === "projects" && (
            <button
              onClick={() => {
                setEditingProject(null);
                setProjectSeed(null);
                setShowProjectModal(true);
              }}
              className="rounded-xl bg-primary px-3 py-2 text-xs font-bold uppercase tracking-widest text-primary-foreground hover:opacity-90"
            >
              New Project
            </button>
          )}
          {activeTab === "repositories" && (
            <button
              onClick={() => void handleCreateRepositoryDraft()}
              disabled={creatingRepository}
              className="rounded-xl bg-primary px-3 py-2 text-xs font-bold uppercase tracking-widest text-primary-foreground hover:opacity-90"
            >
              New Repository
            </button>
          )}
        </div>
      </div>

      <div className="glass-card p-4 space-y-4">
        <div className="flex flex-wrap gap-2">
          {tabs.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setTab(tab.key)}
              data-testid={`catalog-tab-${tab.key}`}
              className={cn(
                "inline-flex items-center gap-2 rounded-xl border px-3 py-2 text-xs font-black uppercase tracking-widest",
                activeTab === tab.key
                  ? "border-primary/40 bg-primary/10 text-primary"
                  : "border-border text-muted-foreground hover:bg-muted/30",
              )}
            >
              <tab.icon className="w-4 h-4" />
              <span>{tab.label}</span>
              <span className="rounded-md bg-muted/50 px-1.5 py-0.5 text-[10px]">{tab.count}</span>
            </button>
          ))}
        </div>

        <input
          value={query}
          onChange={(e) => {
            const next = e.target.value;
            setQuery(next);
            const params = new URLSearchParams(searchParams.toString());
            // 검색어 변경 시 현재 탭을 항상 유지 (탭이 applications로 튀는 현상 방지)
            params.set("tab", activeTab);
            if (next.trim()) params.set("q", next.trim());
            else params.delete("q");
            router.replace(`/admin/catalog?${params.toString()}`);
          }}
          placeholder="key/name/leader/status 검색"
          className="w-full rounded-xl border border-border bg-muted/20 px-3 py-2 text-sm text-foreground"
        />
      </div>

      {error && <PageError message={error} onRetry={() => void loadAll()} />}

      {!error && activeTab === "applications" && (
        filteredApplications.length === 0 ? (
          <PageEmpty message="조회된 Application 이 없습니다." />
        ) : (
          <div className="glass-card overflow-hidden border border-border/60">
            <table className="w-full text-sm">
              <thead className="bg-muted/30 text-xs uppercase tracking-widest text-muted-foreground">
                <tr>
                  <th className="px-4 py-3 text-left">Key</th>
                  <th className="px-4 py-3 text-left">Name</th>
                  <th className="px-4 py-3 text-left">Status</th>
                  <th className="px-4 py-3 text-left">Leader</th>
                  <th className="px-4 py-3 text-left">Projects</th>
                  <th className="px-4 py-3 text-left">Updated</th>
                  <th className="px-4 py-3 text-left">Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredApplications.map((a) => (
                  <tr key={a.id} className="border-t border-border/40">
                    <td className="px-4 py-3 font-mono">{a.key}</td>
                    <td className="px-4 py-3">{a.name}</td>
                    <td className="px-4 py-3">{a.status}</td>
                    <td className="px-4 py-3">{a.owner_user_id || "-"}</td>
                    <td className="px-4 py-3">{appProjectCount.get(a.id) ?? 0}</td>
                    <td className="px-4 py-3">{new Date(a.updated_at).toLocaleString()}</td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <Link
                          href={`/applications/${a.id}`}
                          data-testid={`catalog-app-detail-${a.id}`}
                          className="rounded-lg border border-border px-2 py-1 text-[10px] font-bold uppercase tracking-widest hover:bg-muted/30"
                        >
                          Detail
                        </Link>
                        <button
                          data-testid={`catalog-app-projects-${a.id}`}
                          onClick={() => openProjectTabByApplication(a.id)}
                          className="rounded-lg border border-border px-2 py-1 text-[10px] font-bold uppercase tracking-widest hover:bg-muted/30"
                        >
                          Projects
                        </button>
                        <button
                          onClick={() => {
                            setEditingApplication(a);
                            setShowApplicationModal(true);
                          }}
                          className="rounded-lg border border-border px-2 py-1 text-[10px] font-bold uppercase tracking-widest hover:bg-muted/30"
                        >
                          Edit
                        </button>
                        <button
                          onClick={() => void handleArchiveApplication(a)}
                          className="rounded-lg border border-destructive/40 px-2 py-1 text-[10px] font-bold uppercase tracking-widest text-destructive hover:bg-destructive/10"
                        >
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )
      )}

      {!error && activeTab === "repositories" && (
        filteredRepositories.length === 0 ? (
          <PageEmpty message="조회된 Repository 가 없습니다." />
        ) : (
          <div className="glass-card overflow-hidden border border-border/60">
            <table className="w-full text-sm">
              <thead className="bg-muted/30 text-xs uppercase tracking-widest text-muted-foreground">
                <tr>
                  <th className="px-4 py-3 text-left">Name</th>
                  <th className="px-4 py-3 text-left">Leader</th>
                  <th className="px-4 py-3 text-left">Status</th>
                  <th className="px-4 py-3 text-left">SCM</th>
                  <th className="px-4 py-3 text-left">Private</th>
                  <th className="px-4 py-3 text-left">Projects</th>
                  <th className="px-4 py-3 text-left">Updated</th>
                  <th className="px-4 py-3 text-left">Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredRepositories.map((r) => (
                  <tr key={r.id} className="border-t border-border/40">
                    <td className="px-4 py-3 font-mono">{r.full_name}</td>
                    <td className="px-4 py-3">{r.owner_login || "-"}</td>
                    <td className="px-4 py-3">{r.status}</td>
                    <td className="px-4 py-3">{r.scm_provider || "-"}</td>
                    <td className="px-4 py-3">{r.private ? "yes" : "no"}</td>
                    <td className="px-4 py-3">{repoProjectCount.get(r.id) ?? 0}</td>
                    <td className="px-4 py-3">{new Date(r.updated_at).toLocaleString()}</td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <Link
                          href={`/repositories/${r.id}`}
                          data-testid={`catalog-repo-detail-${r.id}`}
                          className="rounded-lg border border-border px-2 py-1 text-[10px] font-bold uppercase tracking-widest hover:bg-muted/30"
                        >
                          Detail
                        </Link>
                        <button
                          data-testid={`catalog-repo-projects-${r.id}`}
                          onClick={() => openProjectTabByRepository(r.id)}
                          className="rounded-lg border border-border px-2 py-1 text-[10px] font-bold uppercase tracking-widest hover:bg-muted/30"
                        >
                          Projects
                        </button>
                        <button
                          onClick={() => {
                            setEditingProject(null);
                            setProjectSeed({
                              repository_id: r.id,
                              repository_ids: [r.id],
                            });
                            setShowProjectModal(true);
                          }}
                          className="rounded-lg border border-border px-2 py-1 text-[10px] font-bold uppercase tracking-widest hover:bg-muted/30"
                        >
                          Add Project
                        </button>
                        <button
                          onClick={() => void handleRequestPublish(r)}
                          disabled={r.status !== "draft" || publishingRepositoryID === r.id}
                          className="rounded-lg border border-border px-2 py-1 text-[10px] font-bold uppercase tracking-widest hover:bg-muted/30"
                        >
                          Publish
                        </button>
                        <button
                          onClick={() => toast("Repository 삭제는 연결된 SCM에서 관리됩니다.", "warning")}
                          className="rounded-lg border border-border px-2 py-1 text-[10px] font-bold uppercase tracking-widest hover:bg-muted/30"
                        >
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )
      )}

      {!error && activeTab === "projects" && (
        filteredProjects.length === 0 ? (
          <PageEmpty message="조회된 Project 가 없습니다." />
        ) : (
          <div className="glass-card overflow-hidden border border-border/60">
            <table className="w-full text-sm">
              <thead className="bg-muted/30 text-xs uppercase tracking-widest text-muted-foreground">
                <tr>
                  <th className="px-4 py-3 text-left">Key</th>
                  <th className="px-4 py-3 text-left">Name</th>
                  <th className="px-4 py-3 text-left">Application</th>
                  <th className="px-4 py-3 text-left">Status</th>
                  <th className="px-4 py-3 text-left">Leader</th>
                  <th className="px-4 py-3 text-left">Updated</th>
                  <th className="px-4 py-3 text-left">Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredProjects.map((p) => (
                  <tr key={p.id} className="border-t border-border/40">
                    <td className="px-4 py-3 font-mono">{p.key}</td>
                    <td className="px-4 py-3">{p.name}</td>
                    <td className="px-4 py-3">{(p.application_id && applicationNameByID.get(p.application_id)) || "-"}</td>
                    <td className="px-4 py-3">{p.status}</td>
                    <td className="px-4 py-3">{p.owner_user_id || "-"}</td>
                    <td className="px-4 py-3">{new Date(p.updated_at).toLocaleString()}</td>
                    <td className="px-4 py-3">
                      <Link
                        href={`/projects/${p.id}`}
                        data-testid={`catalog-project-detail-${p.id}`}
                        className="rounded-lg border border-border px-2 py-1 text-[10px] font-bold uppercase tracking-widest hover:bg-muted/30"
                      >
                        Detail
                      </Link>
                      <button
                        onClick={() => {
                          setEditingProject(p);
                          setProjectSeed(null);
                          setShowProjectModal(true);
                        }}
                        className="ml-2 rounded-lg border border-border px-2 py-1 text-[10px] font-bold uppercase tracking-widest hover:bg-muted/30"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => void handleArchiveProject(p)}
                        className="ml-2 rounded-lg border border-destructive/40 px-2 py-1 text-[10px] font-bold uppercase tracking-widest text-destructive hover:bg-destructive/10"
                      >
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )
      )}
      <AnimatePresence>
        {showApplicationModal && (
          <ApplicationCreationModal
            initialData={editingApplicationInitialData}
            onClose={() => {
              setShowApplicationModal(false);
              setEditingApplication(null);
            }}
            onCreated={(app) => {
              toast(`Application ${app.name} ${editingApplication ? "updated" : "created"}`, "success");
              setShowApplicationModal(false);
              setEditingApplication(null);
              void loadAll();
            }}
          />
        )}
        {showProjectModal && (
          <ProjectCreationModal
            repositories={repositories}
            initialData={(editingProject ?? projectSeed ?? undefined) as Partial<Project> | undefined}
            onClose={() => {
              setShowProjectModal(false);
              setEditingProject(null);
              setProjectSeed(null);
            }}
            onCreated={(project) => {
              toast(`Project ${project.name} ${editingProject ? "updated" : "created"}`, "success");
              setShowProjectModal(false);
              setEditingProject(null);
              setProjectSeed(null);
              void loadAll();
            }}
          />
        )}
      </AnimatePresence>
    </div>
  );
}
