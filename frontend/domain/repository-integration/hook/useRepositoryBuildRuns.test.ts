import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { useRepositoryBuildRuns, type RepositoryBuildRunStatus } from "./useRepositoryBuildRuns";

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
    getRepositoryBuildRunsWithMeta: vi.fn(),
  },
}));

// import the mocked module (must come after vi.mock)
import { repositoryService } from "../service/repository.service";

const mockBuildRun = (overrides: Record<string, unknown> = {}) => ({
  id: 100,
  repository_id: 1,
  run_external_id: 'ext-100',
  branch: 'main',
  commit_sha: 'feedface1234567',
  status: 'success',
  duration_seconds: 60,
  started_at: '2026-06-15T01:00:00Z',
  finished_at: '2026-06-15T01:01:00Z',
  ...overrides,
});

// build-runs list 를 ListBuildRunsResult envelope 으로 wrap (service.getRepositoryBuildRunsWithMeta mock 용).
const mockListResult = (items: ReturnType<typeof mockBuildRun>[], total?: number) => ({
  status: "ok",
  data: items,
  meta: { total: total ?? items.length },
});
describe("useRepositoryBuildRuns (N-9 residual)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("TC-N9-HOOK-01 — initial fetch 시 1 page 의 build run list 를 반환한다", async () => {
    const items = Array.from({ length: 20 }, (_, i) =>
      mockBuildRun({ id: 100 + i, status: "success" }),
    );
    (
      repositoryService.getRepositoryBuildRunsWithMeta as ReturnType<typeof vi.fn>
    ).mockResolvedValueOnce(mockListResult(items, 47));

    const { result } = renderHook(() => useRepositoryBuildRuns(1, { pageSize: 20 }));

    expect(result.current.loading).toBe(true);
    expect(result.current.items).toEqual([]);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.items).toHaveLength(20);
    expect(result.current.items[0].id).toBe(100);
    // P2-1 fix 정공법: backend meta.total 이 47 이면 hasMore=true.
    // 첫 page (20) < 47 → Load more 가능.
    expect(result.current.total).toBe(47);
    expect(result.current.hasMore).toBe(true);
    expect(repositoryService.getRepositoryBuildRunsWithMeta).toHaveBeenCalledTimes(1);
    expect(repositoryService.getRepositoryBuildRunsWithMeta).toHaveBeenCalledWith(
      1,
      expect.objectContaining({ limit: 20, offset: 0, status: undefined }),
    );
  });

  it("TC-N9-HOOK-02 — status filter 변경 시 refetch (new offset=0)", async () => {
    (
      repositoryService.getRepositoryBuildRunsWithMeta as ReturnType<typeof vi.fn>
    )
      .mockResolvedValueOnce(
        mockListResult(
          Array.from({ length: 20 }, (_, i) => mockBuildRun({ id: 100 + i, status: "success" })),
          20,
        ),
      )
      .mockResolvedValueOnce(mockListResult([mockBuildRun({ id: 999, status: "failed" })], 1));

    const { result, rerender } = renderHook(
      ({ filter }: { filter: RepositoryBuildRunStatus | null }) =>
        useRepositoryBuildRuns(1, { statusFilter: filter, pageSize: 20 }),
      { initialProps: { filter: null as RepositoryBuildRunStatus | null } },
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.items).toHaveLength(20);

    rerender({ filter: "failed" as "failed" | null });

    await waitFor(() => expect(result.current.items).toHaveLength(1));
    expect(result.current.items[0].status).toBe("failed");
    expect(repositoryService.getRepositoryBuildRunsWithMeta).toHaveBeenCalledTimes(2);
    expect(repositoryService.getRepositoryBuildRunsWithMeta).toHaveBeenLastCalledWith(
      1,
      expect.objectContaining({ limit: 20, offset: 0, status: "failed" }),
    );
  });

  it("TC-N9-HOOK-03 — repository 404 (not_found) 에러 정규화", async () => {
    const notFoundError = { code: "repository_not_found", message: "repository not found", status: 404 };
    (
      repositoryService.getRepositoryBuildRunsWithMeta as ReturnType<typeof vi.fn>
    ).mockRejectedValueOnce(notFoundError);

    const { result } = renderHook(() => useRepositoryBuildRuns(999, { pageSize: 20 }));

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.error).not.toBeNull();
    expect(result.current.error?.code).toBe("not_found");
    expect(result.current.error?.message).toBe("Repository not found");
    expect(result.current.items).toEqual([]);
  });

  it("TC-N9-HOOK-04 — network error (no status) 정규화", async () => {
    const networkError = { message: "Network Error" };
    (
      repositoryService.getRepositoryBuildRunsWithMeta as ReturnType<typeof vi.fn>
    ).mockRejectedValueOnce(networkError);

    const { result } = renderHook(() => useRepositoryBuildRuns(1, { pageSize: 20 }));

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.error?.code).toBe("network");
  });

  it("TC-N9-HOOK-05 — 401/403 (unauthorized) 정규화", async () => {
    const unauthorizedError = { status: 403, message: "Forbidden" };
    (
      repositoryService.getRepositoryBuildRunsWithMeta as ReturnType<typeof vi.fn>
    ).mockRejectedValueOnce(unauthorizedError);

    const { result } = renderHook(() => useRepositoryBuildRuns(1, { pageSize: 20 }));

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.error?.code).toBe("unauthorized");
  });

  it("TC-N9-HOOK-06 — abort 시 stale 결과가 setState 되지 않는다 (P2-2 fix 정공법 검증)", async () => {
    // 첫 요청: 100ms 후 "success 20건" 으로 resolve. 두 번째: 1ms 후 "failed 1건".
    // rerender(filter="failed") 가 첫 요청을 abort → 두 번째 결과만 표시되어야 한다.
    (
      repositoryService.getRepositoryBuildRunsWithMeta as ReturnType<typeof vi.fn>
    ).mockImplementation((_id: number, opts: { signal?: AbortSignal } = {}) => {
      return new Promise((resolve, reject) => {
        const isFirst = (repositoryService.getRepositoryBuildRunsWithMeta as ReturnType<typeof vi.fn>).mock
          .results.length === 0;
        if (isFirst) {
          // 첫 요청은 abort signal 발생 시 reject. 안하면 stale 결과로 표시될 위험.
          opts.signal?.addEventListener("abort", () => {
            reject(new DOMException("Aborted", "AbortError"));
          });
          setTimeout(
            () =>
              resolve(
                mockListResult(
                  Array.from({ length: 20 }, (_, i) => mockBuildRun({ id: 100 + i, status: "success" })),
                  20,
                ),
              ),
            50,
          );
        } else {
          setTimeout(
            () => resolve(mockListResult([mockBuildRun({ id: 999, status: "failed" })], 1)),
            5,
          );
        }
      });
    });

    const { result, rerender } = renderHook(
      ({ filter }: { filter: RepositoryBuildRunStatus | null }) =>
        useRepositoryBuildRuns(1, { statusFilter: filter, pageSize: 20 }),
      { initialProps: { filter: null as RepositoryBuildRunStatus | null } },
    );

    // filter 변경 → 첫 요청 abort + 두 번째 요청 fire
    rerender({ filter: "failed" as "failed" | null });

    await waitFor(() => expect(result.current.items.at(-1)?.status).toBe("failed"));
    // 첫 요청 결과 (20 success) 가 setState 되면 안 됨.
    expect(result.current.items).toHaveLength(1);
    expect(result.current.items[0].id).toBe(999);
  });

  it("TC-N9-HOOK-07 — repositoryId 가 string 이면 Number 로 변환", async () => {
    (
      repositoryService.getRepositoryBuildRunsWithMeta as ReturnType<typeof vi.fn>
    ).mockResolvedValueOnce(mockListResult([]));

    const { result } = renderHook(() => useRepositoryBuildRuns("42", { pageSize: 20 }));

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(repositoryService.getRepositoryBuildRunsWithMeta).toHaveBeenCalledWith(
      42,
      expect.objectContaining({ limit: 20, offset: 0, status: undefined }),
    );
  });

  it("TC-N9-HOOK-08 — enabled=false 면 fetch 안함", async () => {
    (
      repositoryService.getRepositoryBuildRunsWithMeta as ReturnType<typeof vi.fn>
    ).mockResolvedValueOnce(mockListResult([]));

    const { result } = renderHook(() => useRepositoryBuildRuns(1, { enabled: false, pageSize: 20 }));

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.items).toEqual([]);
    expect(repositoryService.getRepositoryBuildRunsWithMeta).not.toHaveBeenCalled();
  });

  it("TC-N9-HOOK-09 — signal 옵션이 service 로 plumb 된다 (P2-2 fix contract)", async () => {
    (
      repositoryService.getRepositoryBuildRunsWithMeta as ReturnType<typeof vi.fn>
    ).mockResolvedValueOnce(mockListResult([]));

    renderHook(() => useRepositoryBuildRuns(1, { pageSize: 20 }));

    await waitFor(() => {
      expect(repositoryService.getRepositoryBuildRunsWithMeta).toHaveBeenCalled();
    });
    const opts = (repositoryService.getRepositoryBuildRunsWithMeta as ReturnType<typeof vi.fn>).mock
      .calls[0][1];
    expect(opts.signal).toBeInstanceOf(AbortSignal);
  });
});
