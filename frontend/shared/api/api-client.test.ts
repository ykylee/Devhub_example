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

    it("uses path as-is for non /api/* absolute URLs (when API_BASE_URL is empty)", async () => {
      fetchMock.mockResolvedValue({
        ok: true,
        status: 200,
        text: async () => "{}",
      });

      const { apiClient } = await import("./api-client");
      await apiClient("GET", "https://other.example/health");

      // API_BASE_URL is "" in this suite, so absolute URLs are passed through
      // unchanged. The prepend branch is tested in api-client.basurl.test.ts
      // (file-isolated to keep the `@/shared/config/endpoints` mock pinned).
      expect(fetchMock.mock.calls[0][0]).toBe("https://other.example/health");
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
        if (!(e instanceof ApiError)) throw e;
        expect(e.status).toBe(422);
        expect(e.message).toBe("invalid_payload");
        expect(e.payload).toEqual({ status: "rejected", error: "invalid_payload" });
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
        if (!(e instanceof ApiError)) throw e;
        expect(e.status).toBe(500);
        expect(e.message).toBe("HTTP 500");
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

    it("does NOT retry the request when refresh kind='ok' but no fresh access token is available", async () => {
      // Edge case: refreshAccessToken resolves 'ok' but tokenStore returns null
      // afterwards (e.g. the refresh was applied but tokenStore was cleared
      // by a concurrent session-death event). The original 401 is surfaced
      // and no second fetch is issued.
      getAccessToken
        .mockReturnValueOnce("expired-token") // initial
        .mockReturnValue(null); // after refresh
      getRefreshToken.mockReturnValue("rt-1");
      refreshAccessToken.mockResolvedValue({ kind: "ok" });

      fetchMock.mockResolvedValue({
        ok: false,
        status: 401,
        text: async () => '{"error":"expired"}',
      });

      const { apiClient } = await import("./api-client");

      await expect(apiClient("GET", "/api/v1/me")).rejects.toThrow("expired");
      expect(refreshAccessToken).toHaveBeenCalledTimes(1);
      expect(fetchMock).toHaveBeenCalledTimes(1);
      expect(triggerSessionExpired).not.toHaveBeenCalled();
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

  describe("AbortSignal plumb-through (N-9 residual)", () => {
    it("rejects with DOMException AbortError when signal is already aborted", async () => {
      const { apiClient } = await import("./api-client");
      const controller = new AbortController();
      controller.abort();

      await expect(
        apiClient("GET", "/api/v1/build-runs", undefined, { signal: controller.signal }),
      ).rejects.toMatchObject({ name: "AbortError" });
      // abort 시점 즉시 reject — fetch 자체가 호출되지 않는다.
      expect(fetchMock).not.toHaveBeenCalled();
    });

    it("forwards AbortSignal to fetch options.signal", async () => {
      fetchMock.mockResolvedValue({
        ok: true,
        status: 200,
        text: async () => '{"data":[]}',
      });

      const { apiClient } = await import("./api-client");
      const controller = new AbortController();
      await apiClient("GET", "/api/v1/build-runs", undefined, { signal: controller.signal });

      expect(fetchMock).toHaveBeenCalledTimes(1);
      const init = fetchMock.mock.calls[0][1];
      expect(init.signal).toBe(controller.signal);
    });
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
