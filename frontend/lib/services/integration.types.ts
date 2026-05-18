/**
 * External Integration provider types.
 * Mirrors backend-core/internal/httpapi/integration_registry.go (API-69..75).
 * Sprint claude/work_260518-g (External Integration frontend 진입점 1차).
 * ADR-INT-* + ARCH-INT-01..06.
 *
 * scope: provider list / create / update / sync. binding 관리 + infra topology v2
 * 는 후속 carve out.
 */

export type IntegrationProviderType = "alm" | "scm" | "ci_cd" | "doc" | "infra";

export type IntegrationAuthMode = "token" | "basic" | "oauth2" | "app_password" | "agent";

export interface IntegrationProvider {
  provider_id: string;
  provider_key: string;
  provider_type: IntegrationProviderType;
  display_name: string;
  enabled: boolean;
  auth_mode: IntegrationAuthMode;
  credentials_ref: string;
  capabilities: string[];
  sync_status: string;
  last_sync_at: string | null;
  last_error_code: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateIntegrationProviderInput {
  provider_key: string;
  provider_type: IntegrationProviderType;
  display_name: string;
  auth_mode: IntegrationAuthMode;
  credentials_ref: string;
  capabilities?: string[];
}

export interface UpdateIntegrationProviderInput {
  enabled?: boolean;
  display_name?: string;
  credentials_ref?: string;
  capabilities?: string[];
}

export interface ListIntegrationProvidersOptions {
  provider_type?: IntegrationProviderType;
  enabled?: boolean;
  limit?: number;
}
