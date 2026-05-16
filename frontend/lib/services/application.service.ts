import { apiClient } from "./api-client";
import { API_BASE_URL } from "../config/endpoints";

export interface Application {
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

export interface ApplicationRollup {
  pull_request_distribution: Record<string, number>;
  build_success_rate: number;
  build_avg_duration_seconds: number;
  quality_score: number;
  quality_gate_failed_count: number;
  critical_warning_count: number;
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
  meta: any;
}

class ApplicationService {
  private baseUrl = API_BASE_URL;

  async listApplications(options: { status?: string; q?: string } = {}): Promise<Application[]> {
    const params = new URLSearchParams();
    if (options.status) params.append("status", options.status);
    if (options.q) params.append("q", options.q);

    const query = params.toString();
    const url = `${this.baseUrl}/api/v1/applications${query ? `?${query}` : ""}`;

    const body = await apiClient<ListApplicationsResult>("GET", url);
    return body.data;
  }

  async getApplicationRollup(applicationId: string): Promise<ApplicationRollup> {
    const url = `${this.baseUrl}/api/v1/applications/${applicationId}/rollup`;
    const body = await apiClient<RollupResult>("GET", url);
    return body.data;
  }
}

export const applicationService = new ApplicationService();
