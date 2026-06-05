import { apiClient } from "@/shared/api/api-client";
import { API_BASE_URL } from "@/shared/config/endpoints";

export interface Platform {
  id: string;
  key: string;
  name: string;
  description: string;
  status: string;
  visibility: string;
  owner_user_id: string;
  leader_user_id: string;
  development_unit_id: string;
  start_date: string | null;
  due_date: string | null;
  archived_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface PlatformRollup {
  pull_request_distribution: Record<string, number>;
  build_success_rate: number;
  build_avg_duration_seconds: number;
  quality_score: number;
  quality_gate_failed_count: number;
  critical_warning_count: number;
  // REQ-FR-APPDASH-001 — 단순 % 보다 broken/red 상태 즉시 표기.
  // backend `domain.PlatformRollup.TargetBranchBuildStatus` derive: 연결된 repository 들의
  // 마지막 빌드 결과 종합. healthy | broken | unknown.
  target_branch_build_status: "healthy" | "broken" | "unknown";
}

export interface ListPlatformsResult {
  status: string;
  data: Platform[];
  meta: {
    total: number;
  };
}

export interface RollupResult {
  status: string;
  data: PlatformRollup;
  meta: {
    [key: string]: unknown;
  };
}

export interface GetPlatformResult {
  status: string;
  data: Platform;
  meta?: {
    [key: string]: unknown;
  };
}

export interface PlatformDashboard {
  platform_id: string;
  key: string;
  name: string;
  status: string;
  visibility: string;
  leader: string;
  development_unit: string;
  updated_at: string;
  metrics_overview: {
    target_branch_build_status: string;
    avg_build_duration_seconds: number;
    quality_score: number;
    critical_warning_count: number;
  };
  build_failures: Array<{
    repo_provider: string;
    repo_slug: string;
    branch: string;
    build_number: number;
    failed_at: string;
    error_snippet: string;
    log_url: string;
  }>;
  quality_metrics: {
    normalized_score: number;
    unresolved_issues: {
      blocker: number;
      critical: number;
      major: number;
    };
    comment: string;
  };
  projects_progress: Array<{
    project_id: string;
    key: string;
    name: string;
    progress_percent: number;
    status: string;
    due_date: string | null;
    d_day: number;
    risk_level: string;
    risk_badge_color: string;
  }>;
  linked_dev_requests: Array<{
    dreq_id: string;
    title: string;
    status: string;
    assignee_display_name: string;
    created_at: string;
  }>;
  history_trend: Array<{
    date: string;
    avg_duration_seconds: number;
    build_success_rate: number;
    quality_score: number;
  }>;
}

export interface DashboardResult {
  status: string;
  data: PlatformDashboard;
  meta?: Record<string, unknown>;
}

function normalizePlatformDashboard(dashboard: PlatformDashboard): PlatformDashboard {
  return {
    ...dashboard,
    build_failures: Array.isArray(dashboard.build_failures) ? dashboard.build_failures : [],
    projects_progress: Array.isArray(dashboard.projects_progress) ? dashboard.projects_progress : [],
    linked_dev_requests: Array.isArray(dashboard.linked_dev_requests) ? dashboard.linked_dev_requests : [],
    history_trend: Array.isArray(dashboard.history_trend) ? dashboard.history_trend : [],
  };
}

class PlatformService {
  private baseUrl = API_BASE_URL;

  async listPlatforms(options: { status?: string; q?: string; include_archived?: boolean } = {}): Promise<Platform[]> {
    const params = new URLSearchParams();
    if (options.status) params.append("status", options.status);
    if (options.q) params.append("q", options.q);
    if (options.include_archived !== undefined) params.append("include_archived", String(options.include_archived));

    const query = params.toString();
    const url = `${this.baseUrl}/api/v1/platforms${query ? `?${query}` : ""}`;

    const body = await apiClient<ListPlatformsResult>("GET", url);
    return body.data;
  }

  async getPlatform(platformId: string): Promise<Platform> {
    const url = `${this.baseUrl}/api/v1/platforms/${platformId}`;
    const body = await apiClient<GetPlatformResult>("GET", url);
    return body.data;
  }

  async getPlatformRollup(platformId: string): Promise<PlatformRollup> {
    const url = `${this.baseUrl}/api/v1/platforms/${platformId}/rollup`;
    const body = await apiClient<RollupResult>("GET", url);
    return body.data;
  }

  async getPlatformDashboard(platformId: string): Promise<PlatformDashboard> {
    const url = `${this.baseUrl}/api/v1/platforms/${platformId}/dashboard`;
    const body = await apiClient<DashboardResult>("GET", url);
    return normalizePlatformDashboard(body.data);
  }
}

export const platformService = new PlatformService();
