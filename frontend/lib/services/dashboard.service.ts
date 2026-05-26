import { apiClient } from "./api-client";

export interface DeveloperStreamItem {
  id: string;
  title: string;
  repo: string;
  status: string;
  progress?: number;
  updated_at?: string;
}

export interface DeveloperBuildItem {
  id: string;
  title: string;
  status: string;
  duration?: string;
  commit_sha?: string;
  finished_at?: string;
}

export interface ManagerVelocityPoint {
  name: string;
  quality: number;
  security: number;
}

export interface ManagerTeamLoadItem {
  name: string;
  load: number;
  status: string;
}

export interface ManagerDecisionItem {
  id: string;
  title: string;
  type: string;
  occurred_at: string;
}

class DashboardService {
  private asList<T>(value: unknown): T[] {
    if (Array.isArray(value)) return value as T[];
    if (value && typeof value === "object" && "data" in (value as Record<string, unknown>)) {
      const inner = (value as { data?: unknown }).data;
      return Array.isArray(inner) ? (inner as T[]) : [];
    }
    return [];
  }

  private normalizeDeveloperStream(raw: unknown): DeveloperStreamItem[] {
    return this.asList<Record<string, unknown>>(raw).map((item, idx) => ({
      id: String(item.id ?? item.task_id ?? item.key ?? `stream-${idx}`),
      title: String(item.title ?? item.name ?? "Untitled"),
      repo: String(item.repo ?? item.repository ?? item.repo_name ?? "-"),
      status: String(item.status ?? "unknown"),
      progress: typeof item.progress === "number" ? item.progress : undefined,
      updated_at: typeof item.updated_at === "string" ? item.updated_at : undefined,
    }));
  }

  private normalizeDeveloperBuilds(raw: unknown): DeveloperBuildItem[] {
    return this.asList<Record<string, unknown>>(raw).map((item, idx) => ({
      id: String(item.id ?? item.build_id ?? `build-${idx}`),
      title: String(item.title ?? item.name ?? "Build"),
      status: String(item.status ?? "unknown"),
      duration: typeof item.duration === "string" ? item.duration : undefined,
      commit_sha: typeof item.commit_sha === "string" ? item.commit_sha : (typeof item.sha === "string" ? item.sha : undefined),
      finished_at: typeof item.finished_at === "string" ? item.finished_at : undefined,
    }));
  }

  private normalizeManagerVelocity(raw: unknown): ManagerVelocityPoint[] {
    return this.asList<Record<string, unknown>>(raw).map((item) => ({
      name: String(item.name ?? item.date ?? item.day ?? "-"),
      quality: Number(item.quality ?? item.quality_score ?? 0),
      security: Number(item.security ?? item.security_score ?? 0),
    }));
  }

  private normalizeManagerTeamLoad(raw: unknown): ManagerTeamLoadItem[] {
    return this.asList<Record<string, unknown>>(raw).map((item) => ({
      name: String(item.name ?? item.user ?? item.user_name ?? "-"),
      load: Number(item.load ?? item.load_percent ?? 0),
      status: String(item.status ?? "unknown"),
    }));
  }

  private normalizeManagerDecisions(raw: unknown): ManagerDecisionItem[] {
    return this.asList<Record<string, unknown>>(raw).map((item, idx) => ({
      id: String(item.id ?? item.decision_id ?? `decision-${idx}`),
      title: String(item.title ?? item.summary ?? "Decision"),
      type: String(item.type ?? item.category ?? "general"),
      occurred_at: String(item.occurred_at ?? item.created_at ?? new Date().toISOString()),
    }));
  }

  async getDeveloperStream(): Promise<DeveloperStreamItem[]> {
    const resp = await apiClient<unknown>("GET", "/api/v1/dashboard/developer/stream");
    return this.normalizeDeveloperStream(resp);
  }

  async getDeveloperBuilds(): Promise<DeveloperBuildItem[]> {
    const resp = await apiClient<unknown>("GET", "/api/v1/dashboard/developer/builds");
    return this.normalizeDeveloperBuilds(resp);
  }

  async getManagerVelocity(): Promise<ManagerVelocityPoint[]> {
    const resp = await apiClient<unknown>("GET", "/api/v1/dashboard/manager/velocity");
    return this.normalizeManagerVelocity(resp);
  }

  async getManagerTeamLoad(): Promise<ManagerTeamLoadItem[]> {
    const resp = await apiClient<unknown>("GET", "/api/v1/dashboard/manager/team-load");
    return this.normalizeManagerTeamLoad(resp);
  }

  async getManagerDecisions(): Promise<ManagerDecisionItem[]> {
    const resp = await apiClient<unknown>("GET", "/api/v1/dashboard/manager/decisions");
    return this.normalizeManagerDecisions(resp);
  }
}

export const dashboardService = new DashboardService();
