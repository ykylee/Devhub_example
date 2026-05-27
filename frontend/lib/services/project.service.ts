import { ApiError, apiClient } from "./api-client";
import {
  Application,
  ApplicationRepository,
  Project,
  ProjectActivityItem,
  ProjectRepositoryCreatePayload,
  ProjectRepositoryLink,
  ProjectTaskItem,
  SCMProvider,
} from "./project.types";

type ApplicationQuery = { status?: string; include_archived?: boolean; q?: string };
type ProjectQuery = { status?: string; include_archived?: boolean };

function withQuery(path: string, params?: Record<string, string | number | boolean | undefined>): string {
  if (!params) return path;
  const qs = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined) continue;
    qs.set(key, String(value));
  }
  const raw = qs.toString();
  return raw ? `${path}?${raw}` : path;
}

class ProjectService {
  private asList<T>(value: unknown): T[] {
    if (Array.isArray(value)) return value as T[];
    if (value && typeof value === "object" && "data" in (value as Record<string, unknown>)) {
      const inner = (value as { data?: unknown }).data;
      return Array.isArray(inner) ? (inner as T[]) : [];
    }
    return [];
  }

  private normalizePriority(value: unknown): ProjectTaskItem["priority"] {
    const v = String(value ?? "medium").toLowerCase();
    if (v === "low" || v === "medium" || v === "high" || v === "critical") return v;
    return "medium";
  }

  private normalizeStatus(value: unknown): ProjectTaskItem["status"] {
    const v = String(value ?? "todo").toLowerCase();
    if (v === "todo" || v === "in_progress" || v === "review" || v === "done") return v;
    return "todo";
  }
  async getSCMProviders(): Promise<SCMProvider[]> {
    const resp = await apiClient<{ data: SCMProvider[] }>("GET", "/api/v1/scm/providers");
    return resp.data;
  }

  async getApplications(params?: ApplicationQuery): Promise<Application[]> {
    const path = withQuery("/api/v1/applications", params);
    const resp = await apiClient<{ data: Application[] }>("GET", path);
    return resp.data;
  }

  async createApplication(data: Partial<Application>): Promise<Application> {
    const resp = await apiClient<{ data: Application }>("POST", "/api/v1/applications", data);
    return resp.data;
  }

  async getApplication(id: string): Promise<Application> {
    const resp = await apiClient<{ data: Application }>("GET", `/api/v1/applications/${id}`);
    return resp.data;
  }

  async updateApplication(
    id: string,
    data: Partial<Application> & { hold_reason?: string; resume_reason?: string; archived_reason?: string },
  ): Promise<Application> {
    const resp = await apiClient<{ data: Application }>("PATCH", `/api/v1/applications/${id}`, data);
    return resp.data;
  }

  async archiveApplication(id: string): Promise<void> {
    await apiClient("DELETE", `/api/v1/applications/${id}`);
  }

  async getApplicationRepositories(applicationId: string): Promise<ApplicationRepository[]> {
    const resp = await apiClient<{ data: ApplicationRepository[] }>(
      "GET",
      `/api/v1/applications/${applicationId}/repositories`,
    );
    return resp.data;
  }

  async connectRepository(
    applicationId: string,
    data: { repo_provider: string; repo_full_name: string; role: string },
  ): Promise<ApplicationRepository> {
    const resp = await apiClient<{ data: ApplicationRepository }>(
      "POST",
      `/api/v1/applications/${applicationId}/repositories`,
      data,
    );
    return resp.data;
  }

  async disconnectRepository(applicationId: string, repoProvider: string, repoFullName: string): Promise<void> {
    const repoKey = `${repoProvider}/${repoFullName}`;
    await apiClient("DELETE", `/api/v1/applications/${applicationId}/repositories/${encodeURIComponent(repoKey)}`);
  }

  async getRepositoryProjects(repositoryId: number, params?: ProjectQuery): Promise<Project[]> {
    const path = withQuery(`/api/v1/repositories/${repositoryId}/projects`, params);
    const resp = await apiClient<{ data: Project[] }>("GET", path);
    return resp.data;
  }

  async getApplicationProjects(applicationId: string): Promise<Project[]> {
    const repos = await this.getApplicationRepositories(applicationId);
    const repoIDs = repos
      .map((repo) => repo.repository_id)
      .filter((id): id is number => typeof id === "number" && Number.isFinite(id));
    const nested = await Promise.all(repoIDs.map((id) => this.getRepositoryProjects(id).catch(() => [])));
    return nested.flat();
  }

  async createProject(repositoryId: number, data: Partial<Project>): Promise<Project> {
    const resp = await apiClient<{ data: Project }>("POST", `/api/v1/repositories/${repositoryId}/projects`, data);
    return resp.data;
  }

  async createProjectStandalone(data: Partial<Project> & { repository_ids?: number[]; repository_create_payload?: ProjectRepositoryCreatePayload }): Promise<Project> {
    const resp = await apiClient<{ data: Project }>("POST", "/api/v1/projects", data);
    return resp.data;
  }

  async getApplicationProjectsV2(applicationId: string, params?: ProjectQuery): Promise<Project[]> {
    const path = withQuery(`/api/v1/applications/${applicationId}/projects`, params);
    const resp = await apiClient<{ data: Project[] }>("GET", path);
    return resp.data;
  }

  // Hybrid project creation. v2 endpoint primary, legacy createProject fallback
  // on 404/405 (when DEVHUB_PROJECT_MODEL=legacy or backend route disabled).
  async createApplicationProject(
    applicationId: string,
    data: Partial<Project> & { repository_ids?: number[]; repository_create_payload?: ProjectRepositoryCreatePayload },
  ): Promise<Project> {
    try {
      const resp = await apiClient<{ data: Project }>("POST", `/api/v1/applications/${applicationId}/projects`, data);
      return resp.data;
    } catch (err) {
      if (err instanceof ApiError && (err.status === 404 || err.status === 405)) {
        const repoID =
          (data.repository_ids && data.repository_ids[0]) ||
          (data as Partial<Project> & { repository_id?: number }).repository_id;
        if (repoID) {
          return this.createProject(repoID, { ...data, application_id: applicationId });
        }
      }
      throw err;
    }
  }

  async getProjectRepositories(projectId: string): Promise<ProjectRepositoryLink[]> {
    const resp = await apiClient<{ data: ProjectRepositoryLink[] }>("GET", `/api/v1/projects/${projectId}/repositories`);
    return resp.data;
  }

  async getProjectActivity(projectId: string): Promise<ProjectActivityItem[]> {
    const resp = await apiClient<unknown>("GET", `/api/v1/projects/${projectId}/activity`);
    return this.asList<Record<string, unknown>>(resp).map((item, idx) => ({
      id: String(item.id ?? item.activity_id ?? `activity-${idx}`),
      user: String(item.user ?? item.actor ?? item.user_name ?? "-"),
      action: String(item.action ?? "updated"),
      target: String(item.target ?? item.target_name ?? "-"),
      occurred_at: String(item.occurred_at ?? item.created_at ?? new Date().toISOString()),
    }));
  }

  async getProjectTasks(projectId: string, statuses: string[] = ["todo", "in_progress", "review"]): Promise<ProjectTaskItem[]> {
    const path = withQuery(`/api/v1/projects/${projectId}/tasks`, { status: statuses.join(",") });
    const resp = await apiClient<unknown>("GET", path);
    return this.asList<Record<string, unknown>>(resp).map((item, idx) => ({
      id: String(item.id ?? item.task_id ?? `task-${idx}`),
      title: String(item.title ?? item.name ?? "Untitled Task"),
      priority: this.normalizePriority(item.priority),
      status: this.normalizeStatus(item.status),
      due_date: typeof item.due_date === "string" ? item.due_date : undefined,
      comment_count: typeof item.comment_count === "number" ? item.comment_count : undefined,
      attachment_count: typeof item.attachment_count === "number" ? item.attachment_count : undefined,
    }));
  }

  async linkProjectRepository(projectId: string, repositoryId: number, role: "primary" | "linked" | "shared" = "linked"): Promise<ProjectRepositoryLink> {
    const resp = await apiClient<{ data: ProjectRepositoryLink }>("POST", `/api/v1/projects/${projectId}/repositories`, {
      repository_id: repositoryId,
      role,
    });
    return resp.data;
  }

  async unlinkProjectRepository(projectId: string, repositoryId: number): Promise<void> {
    await apiClient("DELETE", `/api/v1/projects/${projectId}/repositories/${repositoryId}`);
  }

  async getProject(id: string): Promise<Project> {
    const resp = await apiClient<{ data: Project }>("GET", `/api/v1/projects/${id}`);
    return resp.data;
  }

  async updateProject(id: string, data: Partial<Project>): Promise<Project> {
    const resp = await apiClient<{ data: Project }>("PATCH", `/api/v1/projects/${id}`, data);
    return resp.data;
  }

  async archiveProject(id: string): Promise<void> {
    await apiClient("DELETE", `/api/v1/projects/${id}`);
  }

  // Helper to fetch all projects across all repositories
  async listAllProjects(repositoryIds: number[]): Promise<Project[]> {
    const allProjects: Project[] = [];
    for (const repoId of repositoryIds) {
      try {
        const projects = await this.getRepositoryProjects(repoId);
        allProjects.push(...projects);
      } catch (err) {
        console.error(`Failed to fetch projects for repo ${repoId}:`, err);
      }
    }
    return allProjects;
  }
}

export const projectService = new ProjectService();
