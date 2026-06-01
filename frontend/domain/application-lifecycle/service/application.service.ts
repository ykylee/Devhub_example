import { apiClient } from "@/shared/api/api-client";
import { API_BASE_URL } from "@/shared/config/endpoints";
import type { ApplicationStatus, ApplicationVisibility } from "@/domain/application-lifecycle/schema/project.types";

export interface Application {
  id: string;
  key: string;
  name: string;
  description: string;
  status: ApplicationStatus;
  visibility: ApplicationVisibility;
  owner_user_id: string;
  leader_user_id: string;
  development_unit_id: string;
  start_date: string | null;
  due_date: string | null;
  archived_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface ApplicationRollup {
  pull_request_distribution: Record<string, number>;
  build_success_rate: number;
  build_avg_duration_seconds: number;
  quality_score: number;
  quality_gate_failed_count: number;
  critical_warning_count: number;
  // REQ-FR-APPDASH-001 — 단순 % 보다 broken/red 상태 즉시 표기.
  // backend `domain.ApplicationRollup.TargetBranchBuildStatus` derive: 연결된 repository 들의
  // 마지막 빌드 결과 종합. healthy | broken | unknown.
  target_branch_build_status: "healthy" | "broken" | "unknown";
}

export interface ListApplicationsResult {
  status: string;
  data: Application[];
  meta: {
    total: number;
  };
}

export interface RollupResult {
  status: string;
  data: ApplicationRollup;
  meta: {
    [key: string]: unknown;
  };
}

export interface GetApplicationResult {
  status: string;
  data: Application;
  meta?: {
    [key: string]: unknown;
  };
}

export interface ApplicationDashboard {
  application_id: string;
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
  data: ApplicationDashboard;
  meta?: Record<string, unknown>;
}

class ApplicationService {
  private baseUrl = API_BASE_URL;

  async listApplications(options: { status?: string; q?: string; include_archived?: boolean } = {}): Promise<Application[]> {
    const params = new URLSearchParams();
    if (options.status) params.append("status", options.status);
    if (options.q) params.append("q", options.q);
    if (options.include_archived !== undefined) params.append("include_archived", String(options.include_archived));

    const query = params.toString();
    const url = `${this.baseUrl}/api/v1/applications${query ? `?${query}` : ""}`;

    const body = await apiClient<ListApplicationsResult>("GET", url);
    return body.data;
  }

  async getApplication(applicationId: string): Promise<Application> {
    const url = `${this.baseUrl}/api/v1/applications/${applicationId}`;
    const body = await apiClient<GetApplicationResult>("GET", url);
    return body.data;
  }

  async getApplicationRollup(applicationId: string): Promise<ApplicationRollup> {
    const url = `${this.baseUrl}/api/v1/applications/${applicationId}/rollup`;
    const body = await apiClient<RollupResult>("GET", url);
    return body.data;
  }

  async getApplicationDashboard(applicationId: string): Promise<ApplicationDashboard> {
    const url = `${this.baseUrl}/api/v1/applications/${applicationId}/dashboard`;
    const body = await apiClient<DashboardResult>("GET", url);
    return body.data;
  }
}

export const applicationService = new ApplicationService();
