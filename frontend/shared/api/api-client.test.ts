import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("apiClient", () => {
  let fetchMock: ReturnType<typeof vi.fn>;
  let getAccessToken: ReturnType<typeof vi.fn>;
  let getRefreshToken: ReturnType<typeof vi.fn>;
  let refreshAccessToken: ReturnType<typeof vi.fn>;
  let triggerSessionExpired: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    vi.resetModules();
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    getAccessToken = vi.fn(() => null);
    getRefreshToken = vi.fn(() => null);
    refreshAccessToken = vi.fn();
    triggerSessionExpired = vi.fn();

    vi.doMock("@/domain/auth-session/service/token-store", () => ({
      tokenStore: {
        getAccessToken,
        getRefreshToken,
      },
    }));
    vi.doMock("@/domain/auth-session/service/session-death", () => ({
      triggerSessionExpired,
    }));
    vi.doMock("@/domain/auth-session/service/refresh", () => ({
      refreshAccessToken,
    }));
    vi.doMock("@/shared/config/endpoints", () => ({
      API_BASE_URL: "",
    }));
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  describe("basic request", () => {
    it("resolves with parsed JSON body on 200", async () => {
      fetchMock.mockResolvedValue({
        ok: true,
        status: 200,
        text: async () => '{"data":[{"id":1}]}',
      });

      const { apiClient } = await import("./api-client");
      const result = await apiClient("GET", "/api/v1/things");

      expect(fetchMock).toHaveBeenCalledTimes(1);
      const [url, init] = fetchMock.mock.calls[0];
      expect(url).toBe("/api/v1/things");
      expect(init.method).toBe("GET");
      expect(init.body).toBeUndefined();
      expect(init.headers).not.toHaveProperty("Content-Type");
      expect(init.headers).not.toHaveProperty("Authorization");
      expect(result).toEqual({ data: [{ id: 1 }] });
    });

    it("sends Content-Type and JSON-stringified body when body is provided", async () => {
      fetchMock.mockResolvedValue({
        ok: true,
        status: 200,
        text: async () => '{"status":"created"}',
      });

      const { apiClient } = await import("./api-client");
      await apiClient("POST", "/api/v1/things", { name: "foo" });

      const init = fetchMock.mock.calls[0][1];
      expect(init.headers["Content-Type"]).toBe("application/json");
      expect(init.body).toBe('{"name":"foo"}');
    });

    it("injects Authorization Bearer when an access token is present", async () => {
      getAccessToken.mockReturnValue("token-xyz");
      fetchMock.mockResolvedValue({
        ok: true,
        status: 200,
        text: async () => '{"ok":true}',
      });

      const { apiClient } = await import("./api-client");
      await apiClient("GET", "/api/v1/me");

      const init = fetchMock.mock.calls[0][1];
      expect(init.headers["Authorization"]).toBe("Bearer token-xyz");
    });

    it("returns null when response body is empty", async () => {
      fetchMock.mockResolvedValue({
        ok: true,
        status: 204,
        text: async () => "",
      });

      const { apiClient } = await import("./api-client");
      const result = await apiClient<null>("DELETE", "/api/v1/things/1");

      expect(result).toBeNull();
    });

    it("returns { raw: <text> } when body is non-JSON text", async () => {
      fetchMock.mockResolvedValue({
        ok: true,
        status: 200,
        text: async () => "plain-text-payload",
      });

      const { apiClient } = await import("./api-client");
      const result = await apiClient<{ raw: string }>("GET", "/api/v1/raw");

      expect(result).toEqual({ raw: "plain-text-payload" });
    });

    it("does NOT prepend API_BASE_URL for non /api/* paths", async () => {
      vi.resetModules();
      vi.doMock("@/shared/config/endpoints", () => ({ API_BASE_URL: "https://api.example.com" }));
      vi.doMock("@/domain/auth-session/service/token-store", () => ({
        tokenStore: { getAccessToken, getRefreshToken },
      }));
      vi.doMock("@/domain/auth-session/service/session-death", () => ({ triggerSessionExpired }));
      vi.doMock("@/domain/auth-session/service/refresh", () => ({ refreshAccessToken }));

      fetchMock.mockResolvedValue({
        ok: true,
        status: 200,
        text: async () => "{}",
      });

      const { apiClient } = await import("./api-client");
      await apiClient("GET", "https://other.example/health");

      expect(fetchMock.mock.calls[0][0]).toBe("https://other.example/health");
    });

    it("prepends API_BASE_URL for relative /api/* paths", async () => {
      vi.resetModules();
      vi.doMock("@/shared/config/endpoints", () => ({ API_BASE_URL: "https://api.example.com" }));
      vi.doMock("@/domain/auth-session/service/token-store", () => ({
        tokenStore: { getAccessToken, getRefreshToken },
      }));
      vi.doMock("@/domain/auth-session/service/session-death", () => ({ triggerSessionExpired }));
      vi.doMock("@/domain/auth-session/service/refresh", () => ({ refreshAccessToken }));

      fetchMock.mockResolvedValue({
        ok: true,
        status: 200,
        text: async () => "{}",
      });

      const { apiClient } = await import("./api-client");
      await apiClient("GET", "/api/v1/x");

      expect(fetchMock.mock.calls[0][0]).toBe("https://api.example.com/api/v1/x");
    });
  });

  describe("error handling (non-2xx)", () => {
    it("throws ApiError with the error string from JSON body", async () => {
      fetchMock.mockResolvedValue({
        ok: false,
        status: 422,
        text: async () => '{"status":"rejected","error":"invalid_payload"}',
      });

      const { apiClient, ApiError } = await import("./api-client");

      await expect(apiClient("POST", "/api/v1/x", {})).rejects.toBeInstanceOf(ApiError);
      try {
        await apiClient("POST", "/api/v1/x", {});
      } catch (e) {
        const err = e as InstanceType<typeof ApiError>;
        expect(err.status).toBe(422);
        expect(err.message).toBe("invalid_payload");
        expect(err.payload).toEqual({ status: "rejected", error: "invalid_payload" });
      }
    });

    it("throws ApiError with default 'HTTP {status}' message when body has no error field", async () => {
      fetchMock.mockResolvedValue({
        ok: false,
        status: 500,
        text: async () => "{}",
      });

      const { apiClient, ApiError } = await import("./api-client");

      try {
        await apiClient("GET", "/api/v1/x");
        throw new Error("should not reach");
      } catch (e) {
        const err = e as InstanceType<typeof ApiError>;
        expect(err).toBeInstanceOf(ApiError);
        expect(err.status).toBe(500);
        expect(err.message).toBe("HTTP 500");
      }
    });
  });

  describe("401 + refresh flow", () => {
    it("does NOT attempt refresh when neither access nor refresh token are present (anon endpoint 401)", async () => {
      getAccessToken.mockReturnValue(null);
      getRefreshToken.mockReturnValue(null);
      fetchMock.mockResolvedValue({
        ok: false,
        status: 401,
        text: async () => '{"error":"unauthorized"}',
      });

      const { apiClient } = await import("./api-client");

      await expect(apiClient("GET", "/api/v1/things")).rejects.toThrow("unauthorized");
      expect(refreshAccessToken).not.toHaveBeenCalled();
      expect(triggerSessionExpired).not.toHaveBeenCalled();
    });

    it("retries with new Bearer when refresh succeeds (kind: 'ok')", async () => {
      getAccessToken
        .mockReturnValueOnce("expired-token") // initial
        .mockReturnValue("fresh-token"); // after refresh
      getRefreshToken.mockReturnValue("rt-1");
      refreshAccessToken.mockResolvedValue({ kind: "ok" });

      fetchMock
        .mockResolvedValueOnce({
          ok: false,
          status: 401,
          text: async () => '{"error":"expired"}',
        })
        .mockResolvedValueOnce({
          ok: true,
          status: 200,
          text: async () => '{"ok":true}',
        });

      const { apiClient } = await import("./api-client");
      const result = await apiClient<{ ok: boolean }>("GET", "/api/v1/me");

      expect(refreshAccessToken).toHaveBeenCalledTimes(1);
      expect(fetchMock).toHaveBeenCalledTimes(2);
      // Second call should carry the fresh Bearer
      const secondInit = fetchMock.mock.calls[1][1];
      expect(secondInit.headers["Authorization"]).toBe("Bearer fresh-token");
      expect(result).toEqual({ ok: true });
    });

    it("calls triggerSessionExpired and surfaces 401 when refresh kind is 'auth_failed'", async () => {
      getAccessToken.mockReturnValue("expired-token");
      getRefreshToken.mockReturnValue("rt-1");
      refreshAccessToken.mockResolvedValue({ kind: "auth_failed" });

      fetchMock.mockResolvedValue({
        ok: false,
        status: 401,
        text: async () => '{"error":"refresh_rejected"}',
      });

      const { apiClient } = await import("./api-client");

      await expect(apiClient("GET", "/api/v1/me")).rejects.toThrow("refresh_rejected");
      expect(refreshAccessToken).toHaveBeenCalledTimes(1);
      expect(triggerSessionExpired).toHaveBeenCalledTimes(1);
      // Only the initial request fires; no retry on auth_failed.
      expect(fetchMock).toHaveBeenCalledTimes(1);
    });

    it("does NOT triggerSessionExpired on transient_failed; surfaces original 401", async () => {
      getAccessToken.mockReturnValue("expired-token");
      getRefreshToken.mockReturnValue("rt-1");
      refreshAccessToken.mockResolvedValue({ kind: "transient_failed" });

      fetchMock.mockResolvedValue({
        ok: false,
        status: 401,
        text: async () => '{"error":"transient"}',
      });

      const { apiClient } = await import("./api-client");

      await expect(apiClient("GET", "/api/v1/me")).rejects.toThrow("transient");
      expect(refreshAccessToken).toHaveBeenCalledTimes(1);
      expect(triggerSessionExpired).not.toHaveBeenCalled();
      expect(fetchMock).toHaveBeenCalledTimes(1);
    });

    it("attempts refresh when only a refresh token is present (access token cleared earlier)", async () => {
      getAccessToken
        .mockReturnValueOnce(null) // initial — no access token attached on header
        .mockReturnValue("fresh-token");
      getRefreshToken.mockReturnValue("rt-only");
      refreshAccessToken.mockResolvedValue({ kind: "ok" });

      fetchMock
        .mockResolvedValueOnce({
          ok: false,
          status: 401,
          text: async () => '{"error":"unauthorized"}',
        })
        .mockResolvedValueOnce({
          ok: true,
          status: 200,
          text: async () => '{"ok":true}',
        });

      const { apiClient } = await import("./api-client");
      await apiClient("GET", "/api/v1/me");

      expect(refreshAccessToken).toHaveBeenCalledTimes(1);
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
  });

  describe("ApiError", () => {
    it("preserves status / payload / message and has name 'ApiError'", async () => {
      const { ApiError } = await import("./api-client");

      const error = new ApiError(409, { code: "conflict" }, "duplicate");
      expect(error.status).toBe(409);
      expect(error.payload).toEqual({ code: "conflict" });
      expect(error.message).toBe("duplicate");
      expect(error.name).toBe("ApiError");
      expect(error).toBeInstanceOf(Error);
    });
  });
});
