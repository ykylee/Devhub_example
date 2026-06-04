import { ApiError, apiClient } from "@/shared/api/api-client";
import {
  Platform,
  PlatformRepository,
  Project,
  ProjectActivityItem,
  ProjectRepositoryCreatePayload,
  ProjectRepositoryLink,
  ProjectTaskItem,
  SCMProvider,
} from "@/domain/platform-lifecycle/schema/project.types";

type PlatformQuery = { status?: string; include_archived?: boolean; q?: string };
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

  async getPlatforms(params?: PlatformQuery): Promise<Platform[]> {
    const path = withQuery("/api/v1/platforms", params);
    const resp = await apiClient<{ data: Platform[] }>("GET", path);
    return resp.data;
  }

  async createPlatform(data: Partial<Platform>): Promise<Platform> {
    const resp = await apiClient<{ data: Platform }>("POST", "/api/v1/platforms", data);
    return resp.data;
  }

  async getPlatform(id: string): Promise<Platform> {
    const resp = await apiClient<{ data: Platform }>("GET", `/api/v1/platforms/${id}`);
    return resp.data;
  }

  async updatePlatform(
    id: string,
    data: Partial<Platform> & { hold_reason?: string; resume_reason?: string; archived_reason?: string },
  ): Promise<Platform> {
    const resp = await apiClient<{ data: Platform }>("PATCH", `/api/v1/platforms/${id}`, data);
    return resp.data;
  }

  async archivePlatform(id: string, hard?: boolean): Promise<void> {
    const path = hard ? `/api/v1/platforms/${id}?hard=true` : `/api/v1/platforms/${id}`;
    await apiClient("DELETE", path);
  }

  async getPlatformRepositories(platformId: string): Promise<PlatformRepository[]> {
    const resp = await apiClient<{ data: PlatformRepository[] }>(
      "GET",
      `/api/v1/platforms/${platformId}/repositories`,
    );
    return resp.data;
  }

  async connectRepository(
    platformId: string,
    data: { repo_provider: string; repo_full_name: string; role: string },
  ): Promise<PlatformRepository> {
    const resp = await apiClient<{ data: PlatformRepository }>(
      "POST",
      `/api/v1/platforms/${platformId}/repositories`,
      data,
    );
    return resp.data;
  }

  async disconnectRepository(platformId: string, repoProvider: string, repoFullName: string): Promise<void> {
    const repoKey = `${repoProvider}/${repoFullName}`;
    await apiClient("DELETE", `/api/v1/platforms/${platformId}/repositories/${encodeURIComponent(repoKey)}`);
  }

  async getRepositoryProjects(repositoryId: number, params?: ProjectQuery): Promise<Project[]> {
    const path = withQuery(`/api/v1/repositories/${repositoryId}/projects`, params);
    const resp = await apiClient<{ data: Project[] }>("GET", path);
    return resp.data;
  }

  async getPlatformProjects(platformId: string, params?: ProjectQuery): Promise<Project[]> {
    return this.getPlatformProjectsV2(platformId, params);
  }

  async createProject(repositoryId: number, data: Partial<Project>): Promise<Project> {
    const resp = await apiClient<{ data: Project }>("POST", `/api/v1/repositories/${repositoryId}/projects`, data);
    return resp.data;
  }

  async createProjectStandalone(data: Partial<Project> & { repository_ids?: number[]; repository_create_payload?: ProjectRepositoryCreatePayload }): Promise<Project> {
    const resp = await apiClient<{ data: Project }>("POST", "/api/v1/projects", data);
    return resp.data;
  }

  async getPlatformProjectsV2(platformId: string, params?: ProjectQuery): Promise<Project[]> {
    const path = withQuery(`/api/v1/platforms/${platformId}/projects`, params);
    const resp = await apiClient<{ data: Project[] }>("GET", path);
    return resp.data;
  }

  async listStandaloneProjects(params?: ProjectQuery): Promise<Project[]> {
    const path = withQuery(`/api/v1/projects/standalone`, params);
    const resp = await apiClient<{ data: Project[] }>("GET", path);
    return resp.data;
  }

  async createPlatformProject(
    platformId: string,
    data: Partial<Project> & { repository_ids?: number[]; repository_create_payload?: ProjectRepositoryCreatePayload },
  ): Promise<Project> {
    try {
      const resp = await apiClient<{ data: Project }>("POST", `/api/v1/platforms/${platformId}/projects`, data);
      return resp.data;
    } catch (err) {
      if (err instanceof ApiError && (err.status === 404 || err.status === 405)) {
        const repoID =
          (data.repository_ids && data.repository_ids[0]) ||
          (data as Partial<Project> & { repository_id?: number }).repository_id;
        if (repoID) {
          return this.createProject(repoID, { ...data, platform_id: platformId });
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

  async getProjectDashboard(id: string, persona: string): Promise<any> {
    const resp = await apiClient<{ data: any }>("GET", `/api/v1/projects/${id}/dashboard?persona=${persona}`);
    return resp.data;
  }

  async updateProject(id: string, data: Partial<Project>): Promise<Project> {
    const resp = await apiClient<{ data: Project }>("PATCH", `/api/v1/projects/${id}`, data);
    return resp.data;
  }

  async archiveProject(id: string, hard?: boolean): Promise<void> {
    const path = hard ? `/api/v1/projects/${id}?hard=true` : `/api/v1/projects/${id}`;
    await apiClient("DELETE", path);
  }

  async listAllProjects(params?: ProjectQuery): Promise<Project[]> {
    const seen = new Set<string>();
    const allProjects: Project[] = [];
    const pushUnique = (projects: Project[]) => {
      for (const p of projects) {
        if (p && p.id && !seen.has(p.id)) {
          seen.add(p.id);
          allProjects.push(p);
        }
      }
    };

    try {
      const standalone = await this.listStandaloneProjects(params);
      pushUnique(standalone);
    } catch (err) {
      console.warn("[projectService.listAllProjects] standalone fetch failed:", err);
    }

    try {
      const apps = await this.getPlatforms({
        include_archived: params?.include_archived,
      });
      const perAppResults = await Promise.allSettled(
        apps.map((app) => this.getPlatformProjectsV2(app.id, params)),
      );
      for (const result of perAppResults) {
        if (result.status === "fulfilled") {
          pushUnique(result.value);
        } else {
          console.warn(
            "[projectService.listAllProjects] per-platform fetch failed:",
            result.reason,
          );
        }
      }
    } catch (err) {
      console.warn(
        "[projectService.listAllProjects] platforms list failed:",
        err,
      );
    }

    return allProjects;
  }
}

export const projectService = new ProjectService();
