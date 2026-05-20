import { apiClient } from "./api-client";
import type {
  CreateIntegrationBindingInput,
  CreateIntegrationProviderInput,
  IntegrationBinding,
  IntegrationProvider,
  ListIntegrationBindingsOptions,
  ListIntegrationProvidersOptions,
  UpdateIntegrationProviderInput,
} from "./integration.types";

class IntegrationService {
  async listProviders(opts: ListIntegrationProvidersOptions = {}): Promise<{ data: IntegrationProvider[]; total: number }> {
    const params = new URLSearchParams();
    if (opts.provider_type) params.set("provider_type", opts.provider_type);
    if (opts.enabled !== undefined) params.set("enabled", String(opts.enabled));
    if (opts.limit) params.set("limit", String(opts.limit));
    const qs = params.toString();
    const path = qs ? `/api/v1/integration/providers?${qs}` : "/api/v1/integration/providers";
    const resp = await apiClient<{ data: IntegrationProvider[]; meta?: { total: number } }>("GET", path);
    return { data: resp.data, total: resp.meta?.total ?? resp.data.length };
  }

  async createProvider(input: CreateIntegrationProviderInput): Promise<IntegrationProvider> {
    const resp = await apiClient<{ data: IntegrationProvider }>("POST", "/api/v1/integration/providers", input);
    return resp.data;
  }

  async updateProvider(providerID: string, input: UpdateIntegrationProviderInput): Promise<IntegrationProvider> {
    const resp = await apiClient<{ data: IntegrationProvider }>(
      "PATCH",
      `/api/v1/integration/providers/${providerID}`,
      input,
    );
    return resp.data;
  }

  /** API-72 — trigger out-of-band sync. backend 가 sync_status 를 비동기 갱신.
   *  응답은 `{status: "accepted", job_id: ...}` 형식이며 provider envelope 가
   *  아니다 (codex hotfix #6 P1 #2, PR #148). caller 는 sync_status 갱신을
   *  보려면 listProviders() 로 refresh 하거나 optimistic update 한다. */
  async syncProvider(providerID: string): Promise<{ status: string; job_id: string }> {
    return await apiClient<{ status: string; job_id: string }>(
      "POST",
      `/api/v1/integration/providers/${providerID}/sync`,
    );
  }

  /** API-80 — Provider 삭제 (sprint claude/work_260518-j). FK guard:
   *  active binding 존재 시 backend 가 409 `integration_provider_has_bindings`
   *  반환 — caller 는 ApiError.payload.code 로 분기 처리. */
  async deleteProvider(providerID: string): Promise<void> {
    await apiClient<{ status: string }>(
      "DELETE",
      `/api/v1/integration/providers/${providerID}`,
    );
  }

  /** API-74 — provider ↔ application/project binding 목록. RBAC view.
   *  sprint claude/work_260518-m. */
  async listBindings(opts: ListIntegrationBindingsOptions = {}): Promise<{ data: IntegrationBinding[]; total: number }> {
    const params = new URLSearchParams();
    if (opts.scope_type) params.set("scope_type", opts.scope_type);
    if (opts.scope_id) params.set("scope_id", opts.scope_id);
    if (opts.provider_type) params.set("provider_type", opts.provider_type);
    if (opts.enabled !== undefined) params.set("enabled", String(opts.enabled));
    if (opts.limit) params.set("limit", String(opts.limit));
    if (opts.offset) params.set("offset", String(opts.offset));
    const qs = params.toString();
    const path = qs ? `/api/v1/integration/bindings?${qs}` : "/api/v1/integration/bindings";
    const resp = await apiClient<{ data: IntegrationBinding[]; meta?: { total: number } }>("GET", path);
    return { data: resp.data, total: resp.meta?.total ?? resp.data.length };
  }

  /** API-75 — binding 생성. RBAC edit (system_admin only). 409
   *  `integration_binding_conflict` (중복 또는 provider 미존재), 422
   *  `integration_policy_violation` 분기 가능. */
  async createBinding(input: CreateIntegrationBindingInput): Promise<IntegrationBinding> {
    const resp = await apiClient<{ data: IntegrationBinding }>("POST", "/api/v1/integration/bindings", input);
    return resp.data;
  }

  async updateBinding(bindingID: string, input: Partial<CreateIntegrationBindingInput> & { enabled?: boolean }): Promise<IntegrationBinding> {
    const resp = await apiClient<{ data: IntegrationBinding }>(
      "PATCH",
      `/api/v1/integration/bindings/${bindingID}`,
      input,
    );
    return resp.data;
  }

  async deleteBinding(bindingID: string): Promise<void> {
    await apiClient<{ status: string }>(
      "DELETE",
      `/api/v1/integration/bindings/${bindingID}`,
    );
  }
}

export const integrationService = new IntegrationService();
