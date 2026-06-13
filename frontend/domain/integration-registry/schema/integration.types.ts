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
  // 구조화 outbound auth 자격증명 (auth_mode 별). 비밀 외 필드는 노출,
  // auth_secret 은 write-only (auth_secret_set bool 만).
  auth_username: string | null;
  auth_client_id: string | null;
  auth_token_url: string | null;
  auth_secret_set: boolean;
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
  // auth_mode 별 outbound 자격증명. auth_secret 은 write-only.
  auth_username?: string;
  auth_client_id?: string;
  auth_token_url?: string;
  auth_secret?: string;
}

export interface UpdateIntegrationProviderInput {
  enabled?: boolean;
  display_name?: string;
  credentials_ref?: string;
  capabilities?: string[];
  base_url?: string;
  api_token?: string;
  auth_username?: string;
  auth_client_id?: string;
  auth_token_url?: string;
  auth_secret?: string;
}

export interface ListIntegrationProvidersOptions {
  provider_type?: IntegrationProviderType;
  enabled?: boolean;
  limit?: number;
}

// SCM repository import (API-88/89, sprint scm-repo-sync).
// SCM 으로부터 조회한 원격 repository 1건. imported 는 시스템에 이미 연동(import)됐는지.
export interface ScmRepository {
  full_name: string;
  name: string;
  clone_url: string;
  html_url: string;
  default_branch: string;
  private: boolean;
  imported: boolean;
}

export interface ImportScmRepositoriesResult {
  status: string;
  imported: number;
  repositories: { full_name: string; name: string }[];
  not_found: string[];
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
// 1 provider 가 N 개의 platform/project scope 에 매핑되어 외부 시스템과의
// 구체 연결 (Jira PROJ key, Gitea repo path 등) 을 표현. backend §15.2 API-74/75.

export type IntegrationScopeType = "platform" | "project";

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

// ============================================================================
// X-1 System Admin 운영 대시보드 (release_v0-1_roadmap.md line 198, RM-M4-07).
// Backend API-104/105/106 — sync job 큐/상태 + dashboard summary.
// ============================================================================

export type IntegrationSyncJobStatus = "queued" | "running" | "succeeded" | "failed";

export interface IntegrationSyncJob {
  job_id: string;
  provider_id: string;
  requested_by: string | null;
  status: IntegrationSyncJobStatus;
  created_at: string;
}

export interface IntegrationSyncJobStatusCounts {
  queued: number;
  running: number;
  succeeded: number;
  failed: number;
}

export interface ListIntegrationSyncJobsOptions {
  status?: IntegrationSyncJobStatus;
  limit?: number;
  offset?: number;
}

export interface ListIntegrationSyncJobsResponse {
  items: IntegrationSyncJob[];
  total: number;
  limit: number;
  offset: number;
}

export interface IntegrationSyncJobStatusSummaryResponse {
  sync_job_status_counts: IntegrationSyncJobStatusCounts;
}

// ============================================================================
// X-2 inbound webhook multi-provider 정공법 (release_v0-1_roadmap.md §3.5 X-2,
// sprint `feat/work_260614-x2-frontend-e2e`).
// ============================================================================

/** IntegrationProviderType — integration-registry 의 6종 provider type. */
export type IntegrationProviderType =
  | "alm"
  | "scm"
  | "ci_cd"
  | "doc"
  | "infra"
  | "task_tracker"
  | "other";

/** InboundSourceType — N-13 backend foundation (migration 000007). */
export type InboundSourceType = "gitea" | "jira" | "other" | "";

/** WebhookProviderHint — X-2 multi-provider webhook 의 provider 식별 hint. */
export type WebhookProviderHint =
  | "gitea"
  | "jira"
  | "github"
  | "gitlab"
  | "other_custom"
  | "custom"
  | "";

/** AutoRouteReason — X-2 1차 PR #586 의 auto_route.go 의 정공법. */
export type AutoRouteReason =
  | "external_ref_pattern"
  | "requester_email"
  | "req_department"
  | "no_match";

/** AutoRouteDecision — X-2 1차 PR 의 backend Go struct 와 1:1 매핑. */
export interface AutoRouteDecision {
  matched: boolean;
  platform_id: string;
  dev_request_id: string;
  reason: AutoRouteReason;
  provider_hint: WebhookProviderHint;
}

/** InboundSourceRoutingConfig — backend 1차 PR #586 의 Go struct 와 1:1 매핑.
 *  모든 field optional — 미설정 시 backend 가 provider-default pattern 사용. */
export interface InboundSourceRoutingConfig {
  custom_external_ref_pattern?: string;
  custom_requester_pattern?: string;
  custom_department_pattern?: string;
}

/** WebhookEvent — X-2 2~3차 PR 의 WebhookAdapter 가 추출하는 normalized event. */
export interface WebhookEvent {
  provider_type: IntegrationProviderType;
  provider_key: string;
  event_type: string;
  delivery_id?: string;
  external_ref: string;
  actor_login?: string;
  payload_hash: string;
  raw_payload: number[]; // byte array (base64 또는 raw)
}

/** Platform inbound_source 통합 view — frontend 운영 UI 의 SSOT. */
export interface PlatformInboundSourceView {
  platform_id: string;
  platform_key: string;
  platform_name: string;
  inbound_source_type: InboundSourceType;
  inbound_source_config: InboundSourceRoutingConfig;
  updated_at: string;
}
