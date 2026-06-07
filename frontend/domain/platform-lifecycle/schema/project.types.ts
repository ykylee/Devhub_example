/**
 * Project Management Domain Types
 * Based on backend-core/internal/domain/platform.go and concept docs.
 */

export type PlatformStatus = 'planning' | 'active' | 'on_hold' | 'closed' | 'archived';
export type ProjectStatus = PlatformStatus;

export type PlatformVisibility = 'public' | 'internal' | 'restricted';

export type PlatformRepositoryRole = 'primary' | 'sub' | 'shared';

export type PlatformRepositorySyncStatus = 'requested' | 'verifying' | 'active' | 'degraded' | 'disconnected';

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

export type IntegrationScope = 'platform' | 'project';

export type IntegrationType = 'jira' | 'confluence';

export type IntegrationPolicy = 'summary_only' | 'execution_system';

export interface Platform {
  id: string;
  key: string;
  name: string;
  description: string;
  status: PlatformStatus;
  visibility: PlatformVisibility;
  owner_user_id: string;
  leader_user_id: string;
  development_unit_id: string;
  start_date?: string;
  due_date?: string;
  archived_at?: string;
  created_at: string;
  updated_at: string;
}

export interface PlatformRepository {
  platform_id: string;
  repo_provider: string;
  repo_full_name: string;
  external_repo_id?: string;
  role: PlatformRepositoryRole;
  sync_status: PlatformRepositorySyncStatus;
  sync_error_code?: SyncErrorCode;
  sync_error_retryable?: boolean;
  sync_error_at?: string;
  last_sync_at?: string;
  linked_at: string;
  link_source: string;
}

export interface Project {
  id: string;
  platform_id?: string;
  repository_id?: number;
  repository_ids?: number[];
  key: string;
  name: string;
  description: string;
  status: ProjectStatus;
  visibility: PlatformVisibility;
  owner_user_id: string;
  project_members?: Array<{
    user_id: string;
    project_role: ProjectMemberRole;
  }>;
  start_date?: string;
  due_date?: string;
  archived_at?: string;
  created_at: string;
  updated_at: string;
}

export interface ProjectRepositoryCreatePayload {
  key: string;
  slug: string;
  scm_provider: string;
}

export interface ProjectRepositoryLink {
  project_id: string;
  repository_id: number;
  role: "primary" | "linked" | "shared";
  contribution_weight?: number;
  linked_at: string;
}

export interface ProjectActivityItem {
  id: string;
  user: string;
  action: string;
  target: string;
  occurred_at: string;
}

export type ProjectTaskStatus = "todo" | "in_progress" | "review" | "done";

export interface ProjectTaskItem {
  id: string;
  title: string;
  priority: "low" | "medium" | "high" | "critical";
  status: ProjectTaskStatus;
  due_date?: string;
  comment_count?: number;
  attachment_count?: number;
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
  platform_id?: string;
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
