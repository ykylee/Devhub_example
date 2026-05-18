import { apiClient } from "./api-client";
import type {
  CreateIntegrationProviderInput,
  IntegrationProvider,
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
}

export const integrationService = new IntegrationService();
