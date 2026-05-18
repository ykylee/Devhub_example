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

  /** API-72 — trigger out-of-band sync. backend 가 sync_status 를 비동기 갱신. */
  async syncProvider(providerID: string): Promise<IntegrationProvider> {
    const resp = await apiClient<{ data: IntegrationProvider }>(
      "POST",
      `/api/v1/integration/providers/${providerID}/sync`,
    );
    return resp.data;
  }
}

export const integrationService = new IntegrationService();
