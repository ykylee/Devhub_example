import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("DevRequestTokenService", () => {
  const apiClientMock = vi.fn();

  beforeEach(() => {
    vi.resetModules();
    apiClientMock.mockReset();
    vi.doMock("@/shared/api/api-client", () => ({ apiClient: apiClientMock }));
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("list", () => {
    it("issues GET to /api/v1/dev-request-tokens and unwraps data + meta.total", async () => {
      apiClientMock.mockResolvedValue({
        data: [
          {
            token_id: "tok-1",
            client_label: "Jira webhook",
            source_system: "jira",
            allowed_ips: ["192.0.2.1"],
            created_at: "2026-05-28T00:00:00Z",
            created_by: "alice",
            last_used_at: null,
            revoked_at: null,
            expires_at: null,
          },
        ],
        meta: { total: 1 },
      });
      const { devRequestTokenService } = await import("./dev_request_token.service");

      const result = await devRequestTokenService.list();

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/dev-request-tokens");
      expect(result.data).toHaveLength(1);
      expect(result.total).toBe(1);
    });

    it("falls back to data.length when meta.total is missing", async () => {
      apiClientMock.mockResolvedValue({
        data: [{ token_id: "a" }, { token_id: "b" }],
      });
      const { devRequestTokenService } = await import("./dev_request_token.service");

      const result = await devRequestTokenService.list();

      expect(result.total).toBe(2);
    });
  });

  describe("issue", () => {
    it("issues POST with payload and unwraps data envelope (returning plain_token once)", async () => {
      apiClientMock.mockResolvedValue({
        data: {
          token_id: "tok-2",
          client_label: "Gitea hook",
          source_system: "gitea",
          allowed_ips: ["10.0.0.1"],
          created_at: "2026-05-28T01:00:00Z",
          created_by: "bob",
          last_used_at: null,
          revoked_at: null,
          expires_at: "2026-12-31T00:00:00Z",
          plain_token: "secret-once-only-xxx",
        },
      });
      const { devRequestTokenService } = await import("./dev_request_token.service");

      const result = await devRequestTokenService.issue({
        client_label: "Gitea hook",
        source_system: "gitea",
        allowed_ips: ["10.0.0.1"],
        expires_at: "2026-12-31T00:00:00Z",
      });

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        "/api/v1/dev-request-tokens",
        expect.objectContaining({
          client_label: "Gitea hook",
          source_system: "gitea",
          allowed_ips: ["10.0.0.1"],
          expires_at: "2026-12-31T00:00:00Z",
        }),
      );
      expect(result.plain_token).toBe("secret-once-only-xxx");
      expect(result.token_id).toBe("tok-2");
    });
  });

  describe("revoke", () => {
    it("issues DELETE with token id segment and returns the revoked token row", async () => {
      apiClientMock.mockResolvedValue({
        data: {
          token_id: "tok-1",
          client_label: "x",
          source_system: "x",
          allowed_ips: [],
          created_at: "2026-05-28T00:00:00Z",
          created_by: "alice",
          last_used_at: null,
          revoked_at: "2026-05-28T02:00:00Z",
          expires_at: null,
        },
      });
      const { devRequestTokenService } = await import("./dev_request_token.service");

      const result = await devRequestTokenService.revoke("tok-1");

      expect(apiClientMock).toHaveBeenCalledWith(
        "DELETE",
        "/api/v1/dev-request-tokens/tok-1",
      );
      expect(result.revoked_at).toBe("2026-05-28T02:00:00Z");
    });
  });

  describe("update", () => {
    it("issues PATCH with allowed_ips and expires_at payload", async () => {
      apiClientMock.mockResolvedValue({
        data: {
          token_id: "tok-1",
          client_label: "x",
          source_system: "x",
          allowed_ips: ["10.0.0.2"],
          created_at: "2026-05-28T00:00:00Z",
          created_by: "alice",
          last_used_at: null,
          revoked_at: null,
          expires_at: null,
        },
      });
      const { devRequestTokenService } = await import("./dev_request_token.service");

      const result = await devRequestTokenService.update("tok-1", {
        allowed_ips: ["10.0.0.2"],
        expires_at: null,
      });

      expect(apiClientMock).toHaveBeenCalledWith(
        "PATCH",
        "/api/v1/dev-request-tokens/tok-1",
        { allowed_ips: ["10.0.0.2"], expires_at: null },
      );
      expect(result.allowed_ips).toEqual(["10.0.0.2"]);
    });
  });

  describe("updateIPs", () => {
    it("is a thin wrapper around update — same PATCH and payload", async () => {
      apiClientMock.mockResolvedValue({
        data: {
          token_id: "tok-1",
          client_label: "x",
          source_system: "x",
          allowed_ips: ["172.16.0.1"],
          created_at: "2026-05-28T00:00:00Z",
          created_by: "alice",
          last_used_at: null,
          revoked_at: null,
          expires_at: null,
        },
      });
      const { devRequestTokenService } = await import("./dev_request_token.service");

      await devRequestTokenService.updateIPs("tok-1", { allowed_ips: ["172.16.0.1"] });

      expect(apiClientMock).toHaveBeenCalledWith(
        "PATCH",
        "/api/v1/dev-request-tokens/tok-1",
        { allowed_ips: ["172.16.0.1"] },
      );
    });
  });
});
