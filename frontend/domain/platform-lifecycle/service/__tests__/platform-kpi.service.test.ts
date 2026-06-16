import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// fetchPlatformKPI service unit test (Sprint C, TC-PLAT-KPI-SVC-01).
// project-kpi.service.test.ts 와 동일 pattern.

class MockApiError extends Error {
  constructor(public status: number, public payload: unknown, message: string) {
    super(message);
    this.name = "ApiError";
  }
}

describe("fetchPlatformKPI", () => {
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

  it("default window is 30d when no options provided", async () => {
    apiClientMock.mockResolvedValue({
      status: "ok",
      data: {
        platform_id: "pl-1",
        window_from: "2026-05-16T00:00:00Z",
        window_to: "2026-06-15T00:00:00Z",
        weighted_quality_score: 85.0,
        weighted_build_success_rate: 0.9,
        total_build_run_count: 100,
        open_pr_count: 5,
        merged_pr_count: 15,
        active_contributor_count: 8,
        linked_project_count: 3,
        weighted_at: "2026-06-15T00:00:00Z",
      },
    });

    const { fetchPlatformKPI } = await import("../platform-kpi.service");
    const result = await fetchPlatformKPI("pl-1");
    expect(result?.weighted_quality_score).toBe(85.0);
    expect(apiClientMock).toHaveBeenCalledWith(
      "GET",
      "/api/v1/platforms/pl-1/kpi?window=30d",
    );
  });

  it("uses custom windowDays when provided", async () => {
    apiClientMock.mockResolvedValue({
      status: "ok",
      data: {
        platform_id: "pl-1",
        window_from: "2026-06-10T00:00:00Z",
        window_to: "2026-06-15T00:00:00Z",
        weighted_quality_score: 90.0,
        weighted_build_success_rate: 0.95,
        total_build_run_count: 50,
        open_pr_count: 2,
        merged_pr_count: 10,
        active_contributor_count: 5,
        linked_project_count: 3,
        weighted_at: "2026-06-15T00:00:00Z",
      },
    });

    const { fetchPlatformKPI } = await import("../platform-kpi.service");
    await fetchPlatformKPI("pl-1", { windowDays: 7 });
    expect(apiClientMock).toHaveBeenCalledWith(
      "GET",
      "/api/v1/platforms/pl-1/kpi?window=7d",
    );
  });

  it("from/to takes precedence over windowDays when both set", async () => {
    apiClientMock.mockResolvedValue({
      status: "ok",
      data: {
        platform_id: "pl-1",
        window_from: "2026-01-01T00:00:00Z",
        window_to: "2026-06-01T00:00:00Z",
        weighted_quality_score: 0,
        weighted_build_success_rate: 0,
        total_build_run_count: 0,
        open_pr_count: 0,
        merged_pr_count: 0,
        active_contributor_count: 0,
        linked_project_count: 0,
        weighted_at: "2026-06-01T00:00:00Z",
      },
    });

    const { fetchPlatformKPI } = await import("../platform-kpi.service");
    await fetchPlatformKPI("pl-1", {
      windowDays: 30,
      from: "2026-01-01T00:00:00Z",
      to: "2026-06-01T00:00:00Z",
    });
    expect(apiClientMock).toHaveBeenCalledWith(
      "GET",
      "/api/v1/platforms/pl-1/kpi?from=2026-01-01T00%3A00%3A00Z&to=2026-06-01T00%3A00%3A00Z",
    );
  });

  it("returns null when data is missing", async () => {
    apiClientMock.mockResolvedValue({ status: "ok" });
    const { fetchPlatformKPI } = await import("../platform-kpi.service");
    const result = await fetchPlatformKPI("pl-unknown");
    expect(result).toBeNull();
  });

  it("propagates apiClient error", async () => {
    apiClientMock.mockRejectedValue(new MockApiError(503, {}, "server error"));
    const { fetchPlatformKPI } = await import("../platform-kpi.service");
    await expect(fetchPlatformKPI("pl-1")).rejects.toThrow("server error");
  });
});
