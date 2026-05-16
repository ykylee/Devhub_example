import { apiClient } from "./api-client";
import { API_BASE_URL } from "../config/endpoints";

export interface Repository {
  id: number;
  full_name: string;
  owner_login: string;
  name: string;
  clone_url: string;
  html_url: string;
  default_branch: string;
  private: boolean;
  updated_at: string;
}

export interface RepositoryActivity {
  repository_id: number;
  window_from: string;
  window_to: string;
  pr_event_count: number;
  active_contributors: string[];
  build_run_count: number;
  build_success_rate: number;
}

export interface ListRepositoriesResult {
  status: string;
  data: Repository[];
}

export interface ActivityResult {
  status: string;
  data: RepositoryActivity;
}

class RepositoryService {
  private baseUrl = API_BASE_URL;

  async listRepositories(): Promise<Repository[]> {
    const url = `${this.baseUrl}/api/v1/repositories`;
    const body = await apiClient<ListRepositoriesResult>("GET", url);
    return body.data;
  }

  async getRepositoryActivity(repositoryId: number): Promise<RepositoryActivity> {
    const url = `${this.baseUrl}/api/v1/repositories/${repositoryId}/activity`;
    const body = await apiClient<ActivityResult>("GET", url);
    return body.data;
  }
}

export const repositoryService = new RepositoryService();
