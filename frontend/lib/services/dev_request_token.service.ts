import { apiClient } from "./api-client";
import type {
  DevRequestIntakeToken,
  IssueDevRequestIntakeTokenInput,
  IssuedDevRequestIntakeToken,
} from "./dev_request_token.types";

class DevRequestTokenService {
  async list(): Promise<{ data: DevRequestIntakeToken[]; total: number }> {
    const resp = await apiClient<{ data: DevRequestIntakeToken[]; meta?: { total: number } }>(
      "GET",
      "/api/v1/dev-request-tokens",
    );
    return { data: resp.data, total: resp.meta?.total ?? resp.data.length };
  }

  async issue(input: IssueDevRequestIntakeTokenInput): Promise<IssuedDevRequestIntakeToken> {
    const resp = await apiClient<{ data: IssuedDevRequestIntakeToken }>(
      "POST",
      "/api/v1/dev-request-tokens",
      input,
    );
    return resp.data;
  }

  async revoke(tokenID: string): Promise<DevRequestIntakeToken> {
    const resp = await apiClient<{ data: DevRequestIntakeToken }>(
      "DELETE",
      `/api/v1/dev-request-tokens/${tokenID}`,
    );
    return resp.data;
  }
}

export const devRequestTokenService = new DevRequestTokenService();
