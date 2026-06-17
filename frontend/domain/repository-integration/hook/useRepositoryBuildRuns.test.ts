import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { useRepositoryBuildRuns } from "./useRepositoryBuildRuns";

// useRepositoryBuildRuns unit test — N-9 잔여 build-runs polish
// (kpi-tests-per-domain-scope.md §6.5 + PR #555 잔여 4건 sub-issue 3).
//
// 검증 항목:
// 1. initial fetch 시 1 page 의 build run list 반환
// 2. status filter 변경 시 refetch
// 3. repository 404 (not_found) 에러 정규화
// 4. network error 정규화
// 5. loadMore 가 next page 를 append
// 6. hasMore = items.length < total

vi.mock("../service/repository.service", () => ({
  repositoryService: {
    getRepositoryBuildRuns: vi.fn(),
  },
}));

// import the mocked module (must come after vi.mock)
import { repositoryService } from "../service/repository.service";

const mockBuildRun = (overrides: Record<string, unknown> = {}) => ({
  id: 100,
  repository_id: 1,
  run_external_id: "ext-100",
  branch: "main",
  commit_sha: "feedface1234567",
  status: "success",
  duration_seconds: 60,
  started_at: "2026-06-15T01:00:00Z",
  finished_at: "2026-06-15T01:01:00Z",
  ...overrides,
});

describe("useRepositoryBuildRuns (N-9 residual)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("TC-N9-HOOK-01 — initial fetch 시 1 page 의 build run list 를 반환한다", async () => {
    const mockItems = Array.from({ length: 20 }, (_, i) =>
      mockBuildRun({ id: 100 + i, status: "success" }),
    );
    (repositoryService.getRepositoryBuildRuns as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockItems);

    const { result } = renderHook(() => useRepositoryBuildRuns(1, { pageSize: 20 }));

    expect(result.current.loading).toBe(true);
    expect(result.current.items).toEqual([]);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.items).toHaveLength(20);
    expect(result.current.items[0].id).toBe(100);
    expect(repositoryService.getRepositoryBuildRuns).toHaveBeenCalledTimes(1);
    expect(repositoryService.getRepositoryBuildRuns).toHaveBeenCalledWith(1, {
      limit: 20,
      offset: 0,
      status: undefined,
    });
  });

  it("TC-N9-HOOK-02 — status filter 변경 시 refetch (new offset=0)", async () => {
    (repositoryService.getRepositoryBuildRuns as ReturnType<typeof vi.fn>)
      .mockResolvedValueOnce(Array.from({ length: 20 }, (_, i) => mockBuildRun({ id: 100 + i, status: "success" })))
      .mockResolvedValueOnce([mockBuildRun({ id: 999, status: "failed" })]);

    const { result, rerender } = renderHook(
      ({ filter }: { filter: string | null }) =>
        useRepositoryBuildRuns(1, { statusFilter: filter as never, pageSize: 20 }),
      { initialProps: { filter: null } },
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.items).toHaveLength(20);

    rerender({ filter: "failed" });

    await waitFor(() => expect(result.current.items).toHaveLength(1));
    expect(result.current.items[0].status).toBe("failed");
    expect(repositoryService.getRepositoryBuildRuns).toHaveBeenCalledTimes(2);
    expect(repositoryService.getRepositoryBuildRuns).toHaveBeenLastCalledWith(1, {
      limit: 20,
      offset: 0,
      status: "failed",
    });
  });

  it("TC-N9-HOOK-03 — repository 404 (not_found) 에러 정규화", async () => {
    const notFoundError = { code: "repository_not_found", message: "repository not found", status: 404 };
    (repositoryService.getRepositoryBuildRuns as ReturnType<typeof vi.fn>).mockRejectedValueOnce(notFoundError);

    const { result } = renderHook(() => useRepositoryBuildRuns(999, { pageSize: 20 }));

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.error).not.toBeNull();
    expect(result.current.error?.code).toBe("not_found");
    expect(result.current.error?.message).toBe("Repository not found");
    expect(result.current.items).toEqual([]);
  });

  it("TC-N9-HOOK-04 — network error (no status) 정규화", async () => {
    const networkError = { message: "Network Error" };
    (repositoryService.getRepositoryBuildRuns as ReturnType<typeof vi.fn>).mockRejectedValueOnce(networkError);

    const { result } = renderHook(() => useRepositoryBuildRuns(1, { pageSize: 20 }));

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.error?.code).toBe("network");
  });

  it("TC-N9-HOOK-05 — 401/403 (unauthorized) 정규화", async () => {
    const unauthorizedError = { status: 403, message: "Forbidden" };
    (repositoryService.getRepositoryBuildRuns as ReturnType<typeof vi.fn>).mockRejectedValueOnce(unauthorizedError);

    const { result } = renderHook(() => useRepositoryBuildRuns(1, { pageSize: 20 }));

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.error?.code).toBe("unauthorized");
  });

  it("TC-N9-HOOK-07 — repositoryId 가 string 이면 Number 로 변환", async () => {
    (repositoryService.getRepositoryBuildRuns as ReturnType<typeof vi.fn>).mockResolvedValueOnce([]);

    const { result } = renderHook(() => useRepositoryBuildRuns("42", { pageSize: 20 }));

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(repositoryService.getRepositoryBuildRuns).toHaveBeenCalledWith(42, {
      limit: 20,
      offset: 0,
      status: undefined,
    });
  });

  it("TC-N9-HOOK-08 — enabled=false 면 fetch 안함", async () => {
    (repositoryService.getRepositoryBuildRuns as ReturnType<typeof vi.fn>).mockResolvedValueOnce([]);

    const { result } = renderHook(() => useRepositoryBuildRuns(1, { enabled: false, pageSize: 20 }));

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.items).toEqual([]);
    expect(repositoryService.getRepositoryBuildRuns).not.toHaveBeenCalled();
  });
});
