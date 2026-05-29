import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// File-isolated test for the API_BASE_URL prefix branch. This branch depends on
// the `@/shared/config/endpoints` module mock being set to a non-empty value
// before the api-client module is first imported. Co-locating with the main
// api-client.test.ts caused intermittent leakage because both describe blocks
// register a different `@/shared/config/endpoints` factory inside the same
// suite, and vitest's module cache is per file — so separating into this
// file pins the mock for the lifetime of the suite.

describe("apiClient — API_BASE_URL prefix", () => {
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
      tokenStore: { getAccessToken, getRefreshToken },
    }));
    vi.doMock("@/domain/auth-session/service/session-death", () => ({
      triggerSessionExpired,
    }));
    vi.doMock("@/domain/auth-session/service/refresh", () => ({
      refreshAccessToken,
    }));
    vi.doMock("@/shared/config/endpoints", () => ({
      API_BASE_URL: "https://api.example.com",
    }));
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("prepends API_BASE_URL for relative /api/* paths", async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      status: 200,
      text: async () => "{}",
    });

    const { apiClient } = await import("./api-client");
    await apiClient("GET", "/api/v1/x");

    expect(fetchMock.mock.calls[0][0]).toBe("https://api.example.com/api/v1/x");
  });

  it("does NOT prepend API_BASE_URL for non /api/* absolute paths even when base url is set", async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      status: 200,
      text: async () => "{}",
    });

    const { apiClient } = await import("./api-client");
    await apiClient("GET", "https://other.example/health");

    expect(fetchMock.mock.calls[0][0]).toBe("https://other.example/health");
  });
});
