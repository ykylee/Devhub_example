/**
 * Project Management Domain Types
 * Based on backend-core/internal/domain/application.go and concept docs.
 */

export type ApplicationStatus = 'planning' | 'active' | 'on_hold' | 'closed' | 'archived';
export type ProjectStatus = ApplicationStatus;

export type ApplicationVisibility = 'public' | 'internal' | 'restricted';

export type ApplicationRepositoryRole = 'primary' | 'sub' | 'shared';

export type ApplicationRepositorySyncStatus = 'requested' | 'verifying' | 'active' | 'degraded' | 'disconnected';

export type SyncErrorCode =
  | 'provider_unreachable'
  | 'auth_invalid'
  | 'permission_denied'
  | 'rate_limited'
  | 'webhook_signature_invalid'
  | 'payload_schema_mismatch'
  | 'resource_not_found'
  | 'internal_adapter_error';

export type ProjectMemberRole = 'lead' | 'contributor' | 'observer';

export type IntegrationScope = 'application' | 'project';

export type IntegrationType = 'jira' | 'confluence';

export type IntegrationPolicy = 'summary_only' | 'execution_system';

export interface Application {
  id: string;
  key: string;
  name: string;
  description: string;
  status: ApplicationStatus;
  visibility: ApplicationVisibility;
  owner_user_id: string;
  start_date?: string;
  due_date?: string;
  archived_at?: string;
  created_at: string;
  updated_at: string;
}

export interface ApplicationRepository {
  application_id: string;
  repo_provider: string;
  repo_full_name: string;
  external_repo_id?: string;
  role: ApplicationRepositoryRole;
  sync_status: ApplicationRepositorySyncStatus;
  sync_error_code?: SyncErrorCode;
  sync_error_retryable?: boolean;
  sync_error_at?: string;
  last_sync_at?: string;
  linked_at: string;
}

export interface Project {
  id: string;
  application_id?: string;
  repository_id: number;
  key: string;
  name: string;
  description: string;
  status: ProjectStatus;
  visibility: ApplicationVisibility;
  owner_user_id: string;
  start_date?: string;
  due_date?: string;
  archived_at?: string;
  created_at: string;
  updated_at: string;
}

export interface ProjectMember {
  project_id: string;
  user_id: string;
  project_role: ProjectMemberRole;
  joined_at: string;
}

export interface ProjectIntegration {
  id: string;
  scope: IntegrationScope;
  project_id?: string;
  application_id?: string;
  integration_type: IntegrationType;
  external_key: string;
  url: string;
  policy: IntegrationPolicy;
  created_at: string;
  updated_at: string;
}

export interface SCMProvider {
  provider_key: string;
  display_name: string;
  enabled: boolean;
  adapter_version: string;
  created_at: string;
  updated_at: string;
}
