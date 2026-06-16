import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// fetchPlatformTestResults service unit test (Sprint C, TC-PLAT-TESTS-SVC-01).
// project-tests.service.test.ts 와 동일 pattern.

class MockApiError extends Error {
  constructor(public status: number, public payload: unknown, message: string) {
    super(message);
    this.name = "ApiError";
  }
}

describe("fetchPlatformTestResults", () => {
  const apiClientMock = vi.fn();

  beforeEach(() => {
    vi.resetModules();
    apiClientMock.mockReset();
    vi.doMock("@/shared/api/api-client", () => ({
      apiClient: apiClientMock,
      ApiError: MockApiError,
    }));
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("default window is 30d with limit 20 when no options provided", async () => {
    apiClientMock.mockResolvedValue({
      status: "ok",
      data: {
        platform_id: "pl-1",
        window_from: "2026-05-16T00:00:00Z",
        window_to: "2026-06-15T00:00:00Z",
        weighted_pass_rate: 0.93,
        totals: { success: 145, failed: 8, running: 1, cancelled: 2, skipped: 0, queued: 0, unknown: 0 },
        recent: [],
      },
      meta: { total: 156, limit: 20 },
    });

    const { fetchPlatformTestResults } = await import("../platform-tests.service");
    const result = await fetchPlatformTestResults("pl-1");
    expect(result?.weighted_pass_rate).toBe(0.93);
    expect(apiClientMock).toHaveBeenCalledWith(
      "GET",
      "/api/v1/platforms/pl-1/test-results?window=30d",
    );
  });

  it("uses custom window + custom limit when provided", async () => {
    apiClientMock.mockResolvedValue({
      status: "ok",
      data: {
        platform_id: "pl-1",
        window_from: "2026-06-10T00:00:00Z",
        window_to: "2026-06-15T00:00:00Z",
        weighted_pass_rate: null,
        totals: { success: 0, failed: 0, running: 0, cancelled: 0, skipped: 0, queued: 0, unknown: 0 },
        recent: [],
      },
      meta: { total: 0, limit: 10 },
    });

    const { fetchPlatformTestResults } = await import("../platform-tests.service");
    await fetchPlatformTestResults("pl-1", { window: "7d", limit: 10 });
    expect(apiClientMock).toHaveBeenCalledWith(
      "GET",
      "/api/v1/platforms/pl-1/test-results?window=7d&limit=10",
    );
  });

  it("from/to takes precedence over window when both set", async () => {
    apiClientMock.mockResolvedValue({
      status: "ok",
      data: {
        platform_id: "pl-1",
        window_from: "2026-01-01T00:00:00Z",
        window_to: "2026-06-01T00:00:00Z",
        weighted_pass_rate: 0.85,
        totals: { success: 100, failed: 5, running: 0, cancelled: 1, skipped: 0, queued: 0, unknown: 0 },
        recent: [],
      },
      meta: { total: 106, limit: 20 },
    });

    const { fetchPlatformTestResults } = await import("../platform-tests.service");
    await fetchPlatformTestResults("pl-1", {
      window: "7d",
      from: "2026-01-01T00:00:00Z",
      to: "2026-06-01T00:00:00Z",
    });
    expect(apiClientMock).toHaveBeenCalledWith(
      "GET",
      "/api/v1/platforms/pl-1/test-results?from=2026-01-01T00%3A00%3A00Z&to=2026-06-01T00%3A00%3A00Z",
    );
  });

  it("does not append limit param when limit equals default 20", async () => {
    apiClientMock.mockResolvedValue({
      status: "ok",
      data: {
        platform_id: "pl-1",
        window_from: "2026-05-16T00:00:00Z",
        window_to: "2026-06-15T00:00:00Z",
        weighted_pass_rate: 0.9,
        totals: { success: 0, failed: 0, running: 0, cancelled: 0, skipped: 0, queued: 0, unknown: 0 },
        recent: [],
      },
      meta: { total: 0, limit: 20 },
    });

    const { fetchPlatformTestResults } = await import("../platform-tests.service");
    await fetchPlatformTestResults("pl-1", { limit: 20 });
    expect(apiClientMock).toHaveBeenCalledWith(
      "GET",
      "/api/v1/platforms/pl-1/test-results?window=30d",
    );
  });

  it("returns null when data is missing", async () => {
    apiClientMock.mockResolvedValue({ status: "ok" });
    const { fetchPlatformTestResults } = await import("../platform-tests.service");
    const result = await fetchPlatformTestResults("pl-unknown");
    expect(result).toBeNull();
  });

  it("propagates apiClient error", async () => {
    apiClientMock.mockRejectedValue(new MockApiError(400, {}, "invalid window"));
    const { fetchPlatformTestResults } = await import("../platform-tests.service");
    // @ts-expect-error — intentionally invalid window for error path test
    await expect(fetchPlatformTestResults("pl-1", { window: "bogus" })).rejects.toThrow("invalid window");
  });
});
