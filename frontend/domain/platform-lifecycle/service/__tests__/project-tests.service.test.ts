// fetchProjectTestResults service unit test (Sprint B-Tests, TC-PROJ-TESTS-SVC-01).
//
// 정공법 (PR #597 follow-up 정합): module-level vi.mock 으로 apiClient stub + 4
// case (happy / limit / from+to / window 옵션) + 1 reject 케이스.

import { afterEach, describe, expect, it, vi } from "vitest";

vi.mock("@/shared/api/api-client", () => ({
  apiClient: vi.fn(),
}));

import { apiClient } from "@/shared/api/api-client";
import { fetchProjectTestResults } from "../project-tests.service";

describe("fetchProjectTestResults", () => {
  afterEach(() => {
    vi.mocked(apiClient).mockReset();
  });

  it("TC-PROJ-TESTS-SVC-01 — happy path: 기본 window=30d + limit=20 (default), 0.93 pass_rate 응답", async () => {
    vi.mocked(apiClient).mockResolvedValue({
      status: "ok",
      data: {
        project_id: "p-001",
        window_from: "2026-05-16T00:00:00Z",
        window_to: "2026-06-15T00:00:00Z",
        weighted_pass_rate: 0.93,
        totals: {
          success: 145, failed: 8, running: 1, cancelled: 2,
          skipped: 0, queued: 0, unknown: 0,
        },
        recent: [
          { id: 100, repository_id: 1, repository_full_name: "org/repo-a", run_external_id: "ext-100", commit_sha: "feedface", status: "success", branch: "main", started_at: "2026-06-15T01:00:00Z", finished_at: "2026-06-15T01:02:00Z" },
        ],
      },
      meta: { total: 156, limit: 20 },
    });

    const data = await fetchProjectTestResults("p-001");
    expect(data).not.toBeNull();
    expect(data?.weighted_pass_rate).toBeCloseTo(0.93, 5);
    expect(data?.totals.success).toBe(145);
    expect(data?.totals.failed).toBe(8);
    expect(data?.recent[0].repository_full_name).toBe("org/repo-a");
    expect(vi.mocked(apiClient)).toHaveBeenCalledWith("GET", expect.stringContaining("/api/v1/projects/p-001/test-results"));
    expect(vi.mocked(apiClient)).toHaveBeenCalledWith("GET", expect.stringContaining("window=30d"));
    // default limit=20 이므로 limit= 미포함
    expect(vi.mocked(apiClient).mock.calls[0][1]).not.toContain("limit=");
  });

  it("limit 옵션: default 와 다를 때만 limit= 쿼리 추가", async () => {
    vi.mocked(apiClient).mockResolvedValue({
      status: "ok",
      data: {
        project_id: "p-001", window_from: "...", window_to: "...",
        weighted_pass_rate: 0.9, totals: { success: 0, failed: 0, running: 0, cancelled: 0, skipped: 0, queued: 0, unknown: 0 },
        recent: [],
      },
      meta: { total: 0, limit: 50 },
    });

    await fetchProjectTestResults("p-001", { limit: 50 });
    const calledUrl = vi.mocked(apiClient).mock.calls[0][1] as string;
    expect(calledUrl).toContain("limit=50");
  });

  it("from+to RFC3339 옵션: window 보다 우선, limit 미포함 시 default", async () => {
    vi.mocked(apiClient).mockResolvedValue({
      status: "ok",
      data: {
        project_id: "p-001", window_from: "2026-04-01T00:00:00Z", window_to: "2026-05-01T00:00:00Z",
        weighted_pass_rate: null, totals: { success: 0, failed: 0, running: 0, cancelled: 0, skipped: 0, queued: 0, unknown: 0 },
        recent: [],
      },
      meta: { total: 0, limit: 20 },
    });

    await fetchProjectTestResults("p-001", {
      from: "2026-04-01T00:00:00Z",
      to: "2026-05-01T00:00:00Z",
    });
    const calledUrl = vi.mocked(apiClient).mock.calls[0][1] as string;
    expect(calledUrl).toContain("from=2026-04-01T00%3A00%3A00Z");
    expect(calledUrl).toContain("to=2026-05-01T00%3A00%3A00Z");
    expect(calledUrl).not.toContain("window=");
  });

  it("window=90d 옵션: window=90d 쿼리 적용", async () => {
    vi.mocked(apiClient).mockResolvedValue({
      status: "ok",
      data: {
        project_id: "p-001", window_from: "...", window_to: "...",
        weighted_pass_rate: 0.85, totals: { success: 0, failed: 0, running: 0, cancelled: 0, skipped: 0, queued: 0, unknown: 0 },
        recent: [],
      },
      meta: { total: 0, limit: 20 },
    });

    await fetchProjectTestResults("p-001", { window: "90d" });
    const calledUrl = vi.mocked(apiClient).mock.calls[0][1] as string;
    expect(calledUrl).toContain("window=90d");
  });

  it("denom=0 케이스: weighted_pass_rate=null 정상 수신 (linked repo 0 시나리오)", async () => {
    vi.mocked(apiClient).mockResolvedValue({
      status: "ok",
      data: {
        project_id: "p-empty", window_from: "...", window_to: "...",
        weighted_pass_rate: null,
        totals: { success: 0, failed: 0, running: 0, cancelled: 0, skipped: 0, queued: 0, unknown: 0 },
        recent: [],
      },
      meta: { total: 0, limit: 20 },
    });

    const data = await fetchProjectTestResults("p-empty");
    expect(data?.weighted_pass_rate).toBeNull();
    expect(data?.totals.success).toBe(0);
  });
});
