import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// fetchProjectKPI service unit test (Sprint B follow-up, TC-PROJ-KPI-SVC-01).
// Sprint A 의 repository-kpi.service.test.ts 와 동일 pattern.

class MockApiError extends Error {
  constructor(public status: number, public payload: unknown, message: string) {
    super(message);
    this.name = "ApiError";
  }
}

describe("fetchProjectKPI", () => {
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
        project_id: "p-001",
        window_from: "2026-05-16T00:00:00Z",
        window_to: "2026-06-15T00:00:00Z",
        weighted_quality_score: 87.3,
        weighted_build_success_rate: 0.94,
        total_build_run_count: 156,
        weighted_open_pr_count: 7,
        weighted_merged_pr_count: 23,
        active_contributor_count: 12,
        linked_repository_count: 3,
        weighted_at: "2026-06-15T00:00:00Z",
      },
    });

    const { fetchProjectKPI } = await import("../project-kpi.service");
    const data = await fetchProjectKPI("p-001");

    expect(data).not.toBeNull();
    expect(data?.weighted_quality_score).toBe(87.3);
    expect(data?.weighted_build_success_rate).toBe(0.94);
    expect(data?.linked_repository_count).toBe(3);

    const [method, url] = apiClientMock.mock.calls[0];
    expect(method).toBe("GET");
    expect(url).toContain("/api/v1/projects/p-001/kpi");
    expect(url).toContain("window=30d");
  });

  it("from/to take precedence over windowDays", async () => {
    apiClientMock.mockResolvedValue({
      status: "ok",
      data: {
        project_id: "p-001",
        window_from: "2026-01-01T00:00:00Z",
        window_to: "2026-01-31T00:00:00Z",
        weighted_quality_score: 70.0,
        weighted_build_success_rate: 0.5,
        total_build_run_count: 10,
        weighted_open_pr_count: 0,
        weighted_merged_pr_count: 1,
        active_contributor_count: 1,
        linked_repository_count: 1,
        weighted_at: "2026-02-01T00:00:00Z",
      },
    });

    const { fetchProjectKPI } = await import("../project-kpi.service");
    await fetchProjectKPI("p-001", {
      windowDays: 7,
      from: "2026-01-01T00:00:00Z",
      to: "2026-01-31T00:00:00Z",
    });

    const [, url] = apiClientMock.mock.calls[0];
    expect(url).toContain("from=");
    expect(url).toContain("to=");
    expect(url).not.toContain("window=");
  });

  it("returns null when response data is missing (no linked repos)", async () => {
    apiClientMock.mockResolvedValue({ status: "ok", data: null });

    const { fetchProjectKPI } = await import("../project-kpi.service");
    const data = await fetchProjectKPI("p-001");

    expect(data).toBeNull();
  });

  it("propagates 403 auth.policy_unmapped as thrown ApiError", async () => {
    apiClientMock.mockRejectedValue(
      new MockApiError(403, { error: "auth.policy_unmapped" }, "auth.policy_unmapped"),
    );

    const { fetchProjectKPI } = await import("../project-kpi.service");
    await expect(fetchProjectKPI("p-001", { windowDays: 30 })).rejects.toThrow(
      "auth.policy_unmapped",
    );
  });
});
