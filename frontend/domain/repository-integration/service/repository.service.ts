import { apiClient } from "@/shared/api/api-client";
import { API_BASE_URL } from "@/shared/config/endpoints";

export interface Repository {
  id: number;
  full_name: string;
  owner_login: string;
  name: string;
  clone_url: string;
  html_url: string;
  default_branch: string;
  private: boolean;
  status: "draft" | "active";
  // 연동 SCM provider — provider_id(FK) 단일 출처 (migration 000045, 구 scm_provider 통합).
  // provider_key 는 백엔드 join derive (표시용 read-only).
  provider_id?: string;
  provider_key?: string;
  publish_requested_at?: string | null;
  published_at?: string | null;
  updated_at: string;
  // linked classification (Task B, 2026-05-28) — UI 가 linked vs unlinked 분기 표기.
  // 합산 = 0 이면 외부 SCM mirror 만 존재 (orphan), > 0 이면 시스템 platform/project 연결.
  linked_applications_count: number;
  linked_projects_count: number;
}

export interface RepositoryActivity {
  repository_id: number;
  window_from: string;
  window_to: string;
  pr_event_count: number;
  active_contributors: string[];
  build_run_count: number;
  build_success_rate: number;
  // REQ-FR-APPDASH-001 — 단순 % 보다 broken/red 상태 즉시 표기.
  // backend `domain.RepositoryActivity.LastBuildStatus` — window 무관, build_runs 최신 1건.
  last_build_status: "queued" | "running" | "success" | "failed" | "cancelled" | "skipped" | "unknown";
  last_build_at: string | null;
}

export interface RepositoryBuildRun {
  id: number;
  repository_id: number;
  run_external_id: string;
  branch: string;
  commit_sha: string;
  status: "queued" | "running" | "success" | "failed" | "cancelled" | "skipped" | "unknown";
  duration_seconds?: number | null;
  started_at: string;
  finished_at?: string | null;
}

export interface ListRepositoriesResult {
  status: string;
  data: Repository[];
}

export interface ActivityResult {
  status: string;
  data: RepositoryActivity;
}

export interface ListBuildRunsResult {
  status: string;
  data: RepositoryBuildRun[];
  meta?: {
    total?: number;
  };
}

function normalizeRepository(repo: Partial<Repository>): Repository {
  return {
    id: repo.id ?? 0,
    full_name: repo.full_name ?? "",
    owner_login: repo.owner_login ?? "",
    name: repo.name ?? "",
    clone_url: repo.clone_url ?? "",
    html_url: repo.html_url ?? "",
    default_branch: repo.default_branch ?? "",
    private: repo.private ?? false,
    status: repo.status ?? "draft",
    provider_id: repo.provider_id,
    provider_key: repo.provider_key,
    publish_requested_at: repo.publish_requested_at ?? null,
    published_at: repo.published_at ?? null,
    updated_at: repo.updated_at ?? "",
    linked_applications_count: repo.linked_applications_count ?? 0,
    linked_projects_count: repo.linked_projects_count ?? 0,
  };
}

class RepositoryService {
  private baseUrl = API_BASE_URL;

  async listRepositories(): Promise<Repository[]> {
    const url = `${this.baseUrl}/api/v1/repositories`;
    const body = await apiClient<ListRepositoriesResult>("GET", url);
    return body.data.map(normalizeRepository);
  }

  async getRepository(repositoryId: number): Promise<Repository | undefined> {
    const repos = await this.listRepositories();
    return repos.find(r => r.id === repositoryId);
  }

  async createRepositoryDraft(input: { key: string; slug: string; provider_key?: string }): Promise<Repository> {
    const url = `${this.baseUrl}/api/v1/repositories`;
    const body = await apiClient<{ status: string; data: Repository }>("POST", url, input);
    return normalizeRepository(body.data);
  }

  async updateRepository(
    repositoryId: number,
    input: { key?: string; slug?: string; provider_key?: string | null },
  ): Promise<Repository> {
    const url = `${this.baseUrl}/api/v1/repositories/${repositoryId}`;
    const body = await apiClient<{ status: string; data: Repository }>("PATCH", url, input);
    return normalizeRepository(body.data);
  }

  async deleteRepository(repositoryId: number): Promise<void> {
    const url = `${this.baseUrl}/api/v1/repositories/${repositoryId}`;
    await apiClient<{ status: string }>("DELETE", url);
  }

  async requestRepositoryPublish(repositoryId: number): Promise<Repository> {
    const url = `${this.baseUrl}/api/v1/repositories/${repositoryId}/publish`;
    const body = await apiClient<{ status: string; data: Repository }>("POST", url, {});
    return normalizeRepository(body.data);
  }

  async getRepositoryActivity(repositoryId: number): Promise<RepositoryActivity> {
    const url = `${this.baseUrl}/api/v1/repositories/${repositoryId}/activity`;
    const body = await apiClient<ActivityResult>("GET", url);
    return body.data;
  }

  async getRepositoryBuildRuns(
    repositoryId: number,
    options: { limit?: number; offset?: number; status?: string; branch?: string } = {},
  ): Promise<RepositoryBuildRun[]> {
    const body = await this.getRepositoryBuildRunsWithMeta(repositoryId, options);
    return body.data;
  }

  /**
   * build-runs list + meta.total 함께 반환 (cursor pagination 정합).
   * N-9 잔여 polish (codex P2 review 2026-06-17): backend store.ListRepositoryBuildRuns
   * 가 (runs, total, err) 3-tuple 을 반환하므로 frontend 도 meta.total 을 노출해야
   * hasMore 가 정확히 계산된다. 기존 getRepositoryBuildRuns(data only) caller 보존.
   */
  async getRepositoryBuildRunsWithMeta(
    repositoryId: number,
    options: { limit?: number; offset?: number; status?: string; branch?: string; signal?: AbortSignal } = {},
  ): Promise<ListBuildRunsResult> {
    const params = new URLSearchParams();
    if (typeof options.limit === "number") params.set("limit", String(options.limit));
    if (typeof options.offset === "number") params.set("offset", String(options.offset));
    if (options.status) params.set("status", options.status);
    if (options.branch) params.set("branch", options.branch);
    const query = params.toString();
    const url = `${this.baseUrl}/api/v1/repositories/${repositoryId}/build-runs${query ? `?${query}` : ""}`;
    return apiClient<ListBuildRunsResult>("GET", url, undefined, { signal: options.signal });
  }



  async getRepositoryDashboardData(repositoryId: number): Promise<RepositoryDashboardData> {
    return {
      repository_id: repositoryId,
      quality: {
        coverage: 78.5,
        duplication: 2.4,
        quality_gate: repositoryId % 2 === 0 ? "passed" : "failed",
        issues: {
          blocker: repositoryId % 2 === 0 ? 0 : 1,
          critical: repositoryId % 2 === 0 ? 2 : 5,
          major: repositoryId % 2 === 0 ? 8 : 14,
        },
      },
      security: {
        security_gate: repositoryId % 2 === 0 ? "passed" : "failed",
        secrets_detected: repositoryId % 2 === 0 ? 0 : 2,
        vulnerabilities: {
          high: repositoryId % 2 === 0 ? 0 : 1,
          medium: repositoryId % 2 === 0 ? 3 : 7,
          low: repositoryId % 2 === 0 ? 10 : 24,
        },
      },
      productivity: {
        avg_pr_lead_time_hours: repositoryId % 2 === 0 ? 4.2 : 18.5,
        weekly_commits: [
          { week: "05/01", count: 12 },
          { week: "05/08", count: 19 },
          { week: "05/15", count: 15 },
          { week: "05/22", count: 24 },
          { week: "05/29", count: 18 },
        ],
        weekly_prs: [
          { week: "05/01", count: 3 },
          { week: "05/08", count: 5 },
          { week: "05/15", count: 4 },
          { week: "05/22", count: 8 },
          { week: "05/29", count: 6 },
        ],
      },
      linkage: {
        linked_platforms: [
          { id: "p1", name: "Core API Platform", status: "active" },
        ],
        linked_projects: [
          { id: "prj1", name: "Q2 Refactoring", status: "active" },
          { id: "prj2", name: "OIDC Client Integration", status: "closed" },
        ],
      },
    };
  }

  async getBuildRunLog(repositoryId: number, runExternalId: string): Promise<string> {
    const repo = await this.getRepository(repositoryId);
    if (!repo) {
      throw new Error(`Repository with ID ${repositoryId} not found`);
    }
    const [owner, repoName] = repo.full_name.split("/");
    const url = `${this.baseUrl}/api/v1/ci-runs/${runExternalId}/logs?owner=${owner}&repo=${repoName}`;
    
    interface LogLineResponse {
      timestamp: string;
      level: string;
      message: string;
      step_name: string;
    }
    
    interface GetLogsResult {
      status: string;
      data: LogLineResponse[];
    }
    
    const body = await apiClient<GetLogsResult>("GET", url);
    return body.data.map(line => {
      const levelStr = line.level ? `[${line.level.toUpperCase()}] ` : "";
      return `${levelStr}${line.message}`;
    }).join("\n");
  }
}

export interface RepositoryDashboardData {
  repository_id: number;
  quality: {
    coverage: number;
    duplication: number;
    quality_gate: "passed" | "failed";
    issues: {
      blocker: number;
      critical: number;
      major: number;
    };
  };
  security: {
    security_gate: "passed" | "failed";
    secrets_detected: number;
    vulnerabilities: {
      high: number;
      medium: number;
      low: number;
    };
  };
  productivity: {
    avg_pr_lead_time_hours: number;
    weekly_commits: { week: string; count: number }[];
    weekly_prs: { week: string; count: number }[];
  };
  linkage: {
    linked_platforms: { id: string; name: string; status: string }[];
    linked_projects: { id: string; name: string; status: string }[];
  };
}

export const repositoryService = new RepositoryService();
