import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// fetchRepositoryTestResults service unit test (Sprint A follow-up, TC-REPO-TESTS-SVC-01).
//
// 정공법: build_runs.status 분포 (별도 repository_tests table 미존재). limit
// 1~50 강제 (1..50 outside → backend 400), default 20.

class MockApiError extends Error {
  constructor(public status: number, public payload: unknown, message: string) {
    super(message);
    this.name = "ApiError";
  }
}

describe("fetchRepositoryTestResults", () => {
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

  it("default window is 30d and default limit is 20 (limit omitted in URL)", async () => {
    apiClientMock.mockResolvedValue({
      status: "ok",
      data: {
        repository_id: 1,
        window_from: "2026-05-16T00:00:00Z",
        window_to: "2026-06-15T00:00:00Z",
        totals: {
          success: 42,
          failed: 3,
          running: 1,
          cancelled: 1,
          skipped: 0,
          queued: 0,
          unknown: 0,
        },
        pass_rate: 0.933,
        recent: [
          { id: 100, run_external_id: "ext-100", commit_sha: "feedface", status: "success", branch: "main", started_at: "2026-06-15T01:00:00Z", finished_at: "2026-06-15T01:02:00Z" },
        ],
      },
      meta: { total: 47, limit: 20 },
    });

    const { fetchRepositoryTestResults } = await import("../repository-tests.service");
    const data = await fetchRepositoryTestResults(1);

    expect(data).not.toBeNull();
    expect(data?.totals.success).toBe(42);
    expect(data?.pass_rate).toBeCloseTo(0.933, 3);
    expect(data?.recent).toHaveLength(1);

    const [method, url] = apiClientMock.mock.calls[0];
    expect(method).toBe("GET");
    expect(url).toContain("/api/v1/repositories/1/test-results");
    expect(url).toContain("window=30d");
    // default limit = 20 → URL 에 limit 쿼리 미포함 (service 가 default 와 같으면 미전송)
    expect(url).not.toContain("limit=");
  });

  it("passes explicit limit when different from default", async () => {
    apiClientMock.mockResolvedValue({
      status: "ok",
      data: {
        repository_id: 1,
        window_from: "2026-05-16T00:00:00Z",
        window_to: "2026-06-15T00:00:00Z",
        totals: { success: 0, failed: 0, running: 0, cancelled: 0, skipped: 0, queued: 0, unknown: 0 },
        pass_rate: null,
        recent: [],
      },
      meta: { total: 0, limit: 50 },
    });

    const { fetchRepositoryTestResults } = await import("../repository-tests.service");
    await fetchRepositoryTestResults(1, { limit: 50 });

    const [, url] = apiClientMock.mock.calls[0];
    expect(url).toContain("limit=50");
  });

  it("from/to take precedence over window when both set", async () => {
    apiClientMock.mockResolvedValue({
      status: "ok",
      data: {
        repository_id: 1,
        window_from: "2026-01-01T00:00:00Z",
        window_to: "2026-01-31T00:00:00Z",
        totals: { success: 5, failed: 1, running: 0, cancelled: 0, skipped: 0, queued: 0, unknown: 0 },
        pass_rate: 5 / 6,
        recent: [],
      },
      meta: { total: 6, limit: 20 },
    });

    const { fetchRepositoryTestResults } = await import("../repository-tests.service");
    await fetchRepositoryTestResults(1, {
      window: "7d",
      from: "2026-01-01T00:00:00Z",
      to: "2026-01-31T00:00:00Z",
      limit: 20,
    });

    const [, url] = apiClientMock.mock.calls[0];
    expect(url).toContain("from=");
    expect(url).toContain("to=");
    expect(url).not.toContain("window=");
  });

  it("passes explicit window short string (7d / 90d / 1y)", async () => {
    apiClientMock.mockResolvedValue({
      status: "ok",
      data: {
        repository_id: 1,
        window_from: "2026-03-17T00:00:00Z",
        window_to: "2026-06-15T00:00:00Z",
        totals: { success: 1, failed: 0, running: 0, cancelled: 0, skipped: 0, queued: 0, unknown: 0 },
        pass_rate: 1.0,
        recent: [],
      },
      meta: { total: 1, limit: 20 },
    });

    const { fetchRepositoryTestResults } = await import("../repository-tests.service");
    await fetchRepositoryTestResults(1, { window: "90d" });

    const [, url] = apiClientMock.mock.calls[0];
    expect(url).toContain("window=90d");
  });

  it("returns null when response data is missing (no runs in window)", async () => {
    apiClientMock.mockResolvedValue({ status: "ok", data: null });

    const { fetchRepositoryTestResults } = await import("../repository-tests.service");
    const data = await fetchRepositoryTestResults(1);

    expect(data).toBeNull();
  });

  it("propagates 400 (invalid limit) as thrown ApiError", async () => {
    apiClientMock.mockRejectedValue(new MockApiError(400, { error: "limit must be 1..50" }, "limit must be 1..50"));

    const { fetchRepositoryTestResults } = await import("../repository-tests.service");
    await expect(
      fetchRepositoryTestResults(1, { limit: 200 }),
    ).rejects.toThrow("limit must be 1..50");
  });
});
