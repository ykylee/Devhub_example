import { apiClient } from "./api-client";
import { API_BASE_URL } from "../config/endpoints";

export interface Project {
  id: string;
  application_id: string;
  repository_id: number;
  key: string;
  name: string;
  description: string;
  status: string;
  visibility: string;
  owner_user_id: string;
  start_date: string | null;
  due_date: string | null;
  archived_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface ListProjectsResult {
  status: string;
  data: Project[];
  meta: {
    total: number;
  };
}

class ProjectService {
  private baseUrl = API_BASE_URL;

  async listProjects(repositoryId: number): Promise<Project[]> {
    const url = `${this.baseUrl}/api/v1/repositories/${repositoryId}/projects`;
    const body = await apiClient<ListProjectsResult>("GET", url);
    return body.data;
  }

  // Helper to fetch all projects across all repositories
  async listAllProjects(repositoryIds: number[]): Promise<Project[]> {
    const allProjects: Project[] = [];
    for (const repoId of repositoryIds) {
      try {
        const projects = await this.listProjects(repoId);
        allProjects.push(...projects);
      } catch (err) {
        console.error(`Failed to fetch projects for repo ${repoId}:`, err);
      }
    }
    return allProjects;
  }
}

export const projectService = new ProjectService();
