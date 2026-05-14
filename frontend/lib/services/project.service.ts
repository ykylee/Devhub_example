import { apiClient } from './api-client';
import {
  Application,
  ApplicationRepository,
  SCMProvider,
  Project,
} from './project.types';

class ProjectService {
  // SCM Providers
  async getSCMProviders(): Promise<SCMProvider[]> {
    const resp = await apiClient.get<{ data: SCMProvider[] }>('/api/v1/scm/providers');
    return resp.data.data;
  }

  // Applications
  async getApplications(params?: { status?: string; include_archived?: boolean; q?: string }): Promise<Application[]> {
    const resp = await apiClient.get<{ data: Application[] }>('/api/v1/applications', { params });
    return resp.data.data;
  }

  async createApplication(data: Partial<Application>): Promise<Application> {
    const resp = await apiClient.post<{ data: Application }>('/api/v1/applications', data);
    return resp.data.data;
  }

  async getApplication(id: string): Promise<Application> {
    const resp = await apiClient.get<{ data: Application }>(`/api/v1/applications/${id}`);
    return resp.data.data;
  }

  async updateApplication(id: string, data: Partial<Application> & { hold_reason?: string; resume_reason?: string; archived_reason?: string }): Promise<Application> {
    const resp = await apiClient.patch<{ data: Application }>(`/api/v1/applications/${id}`, data);
    return resp.data.data;
  }

  async archiveApplication(id: string): Promise<void> {
    await apiClient.delete(`/api/v1/applications/${id}`);
  }

  // Application-Repository Links
  async getApplicationRepositories(applicationId: string): Promise<ApplicationRepository[]> {
    const resp = await apiClient.get<{ data: ApplicationRepository[] }>(`/api/v1/applications/${applicationId}/repositories`);
    return resp.data.data;
  }

  async connectRepository(applicationId: string, data: { repo_provider: string; repo_full_name: string; role: string }): Promise<ApplicationRepository> {
    const resp = await apiClient.post<{ data: ApplicationRepository }>(`/api/v1/applications/${applicationId}/repositories`, data);
    return resp.data.data;
  }

  async disconnectRepository(applicationId: string, repoProvider: string, repoFullName: string): Promise<void> {
    // Note: API-50 uses repo_key in path, but usually it's provider/fullname. 
    // Backend contract says /repositories/{repo_key}. We might need to encode it or use params.
    const repoKey = `${repoProvider}/${repoFullName}`;
    await apiClient.delete(`/api/v1/applications/${applicationId}/repositories/${encodeURIComponent(repoKey)}`);
  }

  // Projects (Hosted under Repositories)
  async getRepositoryProjects(repositoryId: number): Promise<Project[]> {
    const resp = await apiClient.get<{ data: Project[] }>(`/api/v1/repositories/${repositoryId}/projects`);
    return resp.data.data;
  }

  async createProject(repositoryId: number, data: Partial<Project>): Promise<Project> {
    const resp = await apiClient.post<{ data: Project }>(`/api/v1/repositories/${repositoryId}/projects`, data);
    return resp.data.data;
  }

  async getProject(id: string): Promise<Project> {
    const resp = await apiClient.get<{ data: Project }>(`/api/v1/projects/${id}`);
    return resp.data.data;
  }

  async updateProject(id: string, data: Partial<Project>): Promise<Project> {
    const resp = await apiClient.patch<{ data: Project }>(`/api/v1/projects/${id}`, data);
    return resp.data.data;
  }

  async archiveProject(id: string): Promise<void> {
    await apiClient.delete(`/api/v1/projects/${id}`);
  }
}

export const projectService = new ProjectService();
