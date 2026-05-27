/**
 * External Integration provider + binding types.
 * Mirrors backend-core/internal/httpapi/integration_registry.go (API-69..75, 80).
 * Provider scope: sprint claude/work_260518-g.
 * Binding scope: sprint claude/work_260518-m (bindings 관리 UI).
 * ADR-INT-* + ARCH-INT-01..06.
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
  base_url: string | null;
  /** api_token 은 write-only — 응답엔 raw 미노출, 설정 여부만. */
  api_token_set: boolean;
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
  base_url?: string;
  api_token?: string;
}

export interface UpdateIntegrationProviderInput {
  enabled?: boolean;
  display_name?: string;
  credentials_ref?: string;
  capabilities?: string[];
  base_url?: string;
  api_token?: string;
}

export interface ListIntegrationProvidersOptions {
  provider_type?: IntegrationProviderType;
  enabled?: boolean;
  limit?: number;
}

// 등록 UX 고도화 #5 — test-connection 응답.
export interface TestConnectionResult {
  status: string;
  reachable: boolean;
  status_code?: number;
  latency_ms?: number;
  error?: string;
}

// Bindings — sprint claude/work_260518-m.
// 1 provider 가 N 개의 application/project scope 에 매핑되어 외부 시스템과의
// 구체 연결 (Jira PROJ key, Gitea repo path 등) 을 표현. backend §15.2 API-74/75.

export type IntegrationScopeType = "application" | "project";

export type IntegrationPolicy = "summary_only" | "execution_system";

export interface IntegrationBinding {
  binding_id: string;
  scope_type: IntegrationScopeType;
  scope_id: string;
  provider_id: string;
  external_key: string;
  policy: IntegrationPolicy;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateIntegrationBindingInput {
  scope_type: IntegrationScopeType;
  scope_id: string;
  provider_id: string;
  external_key: string;
  policy: IntegrationPolicy;
}

export interface ListIntegrationBindingsOptions {
  scope_type?: IntegrationScopeType;
  scope_id?: string;
  provider_type?: IntegrationProviderType;
  enabled?: boolean;
  limit?: number;
  offset?: number;
}
