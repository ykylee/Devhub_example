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
  // 합산 = 0 이면 외부 SCM mirror 만 존재 (orphan), > 0 이면 시스템 application/project 연결.
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

class RepositoryService {
  private baseUrl = API_BASE_URL;

  async listRepositories(): Promise<Repository[]> {
    const url = `${this.baseUrl}/api/v1/repositories`;
    const body = await apiClient<ListRepositoriesResult>("GET", url);
    return body.data;
  }

  async getRepository(repositoryId: number): Promise<Repository | undefined> {
    const repos = await this.listRepositories();
    return repos.find(r => r.id === repositoryId);
  }

  async createRepositoryDraft(input: { key: string; slug: string; provider_key?: string }): Promise<Repository> {
    const url = `${this.baseUrl}/api/v1/repositories`;
    const body = await apiClient<{ status: string; data: Repository }>("POST", url, input);
    return body.data;
  }

  async updateRepository(
    repositoryId: number,
    input: { key?: string; slug?: string; provider_key?: string | null },
  ): Promise<Repository> {
    const url = `${this.baseUrl}/api/v1/repositories/${repositoryId}`;
    const body = await apiClient<{ status: string; data: Repository }>("PATCH", url, input);
    return body.data;
  }

  async deleteRepository(repositoryId: number): Promise<void> {
    const url = `${this.baseUrl}/api/v1/repositories/${repositoryId}`;
    await apiClient<{ status: string }>("DELETE", url);
  }

  async requestRepositoryPublish(repositoryId: number): Promise<Repository> {
    const url = `${this.baseUrl}/api/v1/repositories/${repositoryId}/publish`;
    const body = await apiClient<{ status: string; data: Repository }>("POST", url, {});
    return body.data;
  }

  async getRepositoryActivity(repositoryId: number): Promise<RepositoryActivity> {
    const url = `${this.baseUrl}/api/v1/repositories/${repositoryId}/activity`;
    const body = await apiClient<ActivityResult>("GET", url);
    return body.data;
  }

  async getRepositoryBuildRuns(repositoryId: number, options: { limit?: number; offset?: number; status?: string; branch?: string } = {}): Promise<RepositoryBuildRun[]> {
    const params = new URLSearchParams();
    if (typeof options.limit === "number") params.set("limit", String(options.limit));
    if (typeof options.offset === "number") params.set("offset", String(options.offset));
    if (options.status) params.set("status", options.status);
    if (options.branch) params.set("branch", options.branch);
    const query = params.toString();
    const url = `${this.baseUrl}/api/v1/repositories/${repositoryId}/build-runs${query ? `?${query}` : ""}`;
    const body = await apiClient<ListBuildRunsResult>("GET", url);
    return body.data;
  }
}

export const repositoryService = new RepositoryService();
