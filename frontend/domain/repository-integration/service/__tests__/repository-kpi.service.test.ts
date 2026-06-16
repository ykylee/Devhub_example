import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// fetchRepositoryKPI service unit test (Sprint A follow-up, TC-REPO-KPI-SVC-01).
//
// 정공법: apiClient<T> 자동 token refresh + session death. 기존
// repository.service.test.ts 의 `apiClientMock` + `vi.doMock` 패턴 그대로.
//
// 404 케이스 정합: apiClient 는 !response.ok 일 때 ApiError throw. service 가
// null 을 반환하는 정공법은 404 도 catch 후 null 로 변환하는 경우 (현재 코드
// 정공법과 불일치) — 따라서 404 는 `rejects.toThrow` 로 검증.

// mock 에서도 실제 ApiError class 와 동일 signature 를 사용 — test 가 그 class
// instance 를 throw 해도 service 가 catch 없이 propagate 하는지 검증 가능.
class MockApiError extends Error {
  constructor(public status: number, public payload: unknown, message: string) {
    super(message);
    this.name = "ApiError";
  }
}

describe("fetchRepositoryKPI", () => {
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
        repository_id: 1,
        window_from: "2026-05-16T00:00:00Z",
        window_to: "2026-06-15T00:00:00Z",
        quality_score: 87.3,
        quality_score_measured_at: "2026-06-14T22:00:00Z",
        build_success_rate: 0.94,
        build_run_count: 47,
        open_pr_count: 3,
        merged_pr_count: 12,
        active_contributor_count: 4,
      },
    });

    const { fetchRepositoryKPI } = await import("../repository-kpi.service");
    const data = await fetchRepositoryKPI(1);

    expect(data).not.toBeNull();
    expect(data?.quality_score).toBe(87.3);
    expect(data?.build_success_rate).toBe(0.94);

    const [method, url] = apiClientMock.mock.calls[0];
    expect(method).toBe("GET");
    expect(url).toContain("/api/v1/repositories/1/kpi");
    expect(url).toContain("window=30d");
  });

  it("uses explicit windowDays option (7d / 30d / 90d / 365d)", async () => {
    apiClientMock.mockResolvedValue({
      status: "ok",
      data: {
        repository_id: 1,
        window_from: "2026-06-08T00:00:00Z",
        window_to: "2026-06-15T00:00:00Z",
        quality_score: 90.0,
        quality_score_measured_at: null,
        build_success_rate: 1.0,
        build_run_count: 5,
        open_pr_count: 1,
        merged_pr_count: 0,
        active_contributor_count: 2,
      },
    });

    const { fetchRepositoryKPI } = await import("../repository-kpi.service");
    const data = await fetchRepositoryKPI(1, { windowDays: 7 });

    expect(data?.quality_score).toBe(90.0);
    const [, url] = apiClientMock.mock.calls[0];
    expect(url).toContain("window=7d");
  });

  it("from/to take precedence over windowDays when both set", async () => {
    apiClientMock.mockResolvedValue({
      status: "ok",
      data: {
        repository_id: 1,
        window_from: "2026-01-01T00:00:00Z",
        window_to: "2026-01-31T00:00:00Z",
        quality_score: 70.0,
        quality_score_measured_at: null,
        build_success_rate: 0.5,
        build_run_count: 10,
        open_pr_count: 0,
        merged_pr_count: 1,
        active_contributor_count: 1,
      },
    });

    const { fetchRepositoryKPI } = await import("../repository-kpi.service");
    await fetchRepositoryKPI(1, {
      windowDays: 7,
      from: "2026-01-01T00:00:00Z",
      to: "2026-01-31T00:00:00Z",
    });

    const [, url] = apiClientMock.mock.calls[0];
    expect(url).toContain("from=");
    expect(url).toContain("to=");
    // from/to 가 우선 → window 쿼리는 미포함
    expect(url).not.toContain("window=");
  });

  it("returns null when response data is missing (empty repository)", async () => {
    apiClientMock.mockResolvedValue({ status: "ok", data: null });

    const { fetchRepositoryKPI } = await import("../repository-kpi.service");
    const data = await fetchRepositoryKPI(1, { windowDays: 30 });

    expect(data).toBeNull();
  });

  it("propagates API errors (non-2xx) as thrown ApiError", async () => {
    // 4xx/5xx — apiClient 가 ApiError throw. service 가 swallow 하지 않음.
    apiClientMock.mockRejectedValue(new MockApiError(500, { error: "internal" }, "HTTP 500"));

    const { fetchRepositoryKPI } = await import("../repository-kpi.service");
    await expect(fetchRepositoryKPI(1, { windowDays: 30 })).rejects.toThrow("HTTP 500");
  });

  it("propagates 404 as thrown ApiError (service does not swallow)", async () => {
    // sprint plan §1.1.1 의 404 정합: backend 가 status=ok + data=null 로 응답
    // (실제 404 안 던짐). 따라서 404 케이스는 service 가 throw 하지 않고 null 반환.
    // — 이 케이스는 backend 가 404 를 *던지는* 미래 변경에 대한 회귀 가드.
    apiClientMock.mockRejectedValue(new MockApiError(404, { error: "not_found" }, "HTTP 404"));

    const { fetchRepositoryKPI } = await import("../repository-kpi.service");
    await expect(fetchRepositoryKPI(1, { windowDays: 30 })).rejects.toThrow("HTTP 404");
  });
});
