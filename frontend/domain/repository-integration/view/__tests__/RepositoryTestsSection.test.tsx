import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import React from "react";

// RepositoryTestsSection 컴포넌트 unit test (Sprint A follow-up, TC-REPO-TESTS-UI-01).
//
// 정공법: RepositoryTestsSection 가 import 하는 repository-tests.service
// 의 fetchRepositoryTestResults 만 vi.mock 으로 stub.

vi.mock("../../service/repository-tests.service", () => ({
  fetchRepositoryTestResults: vi.fn(),
}));

vi.mock("recharts", () => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const passthrough = (_props: any) => null;
  return {
    ResponsiveContainer: ({ children }: { children: React.ReactNode }) =>
      React.createElement("div", { className: "responsive-container-mock" }, children),
    PieChart: ({ children }: { children?: React.ReactNode }) =>
      React.createElement("div", { "data-testid": "pie-chart" }, children),
    Pie: ({ data }: { data?: unknown }) =>
      React.createElement("div", { "data-testid": "pie-data", "data-data": JSON.stringify(data) }),
    Cell: passthrough,
    Tooltip: passthrough,
  };
});

import { RepositoryTestsSection } from "../RepositoryTestsSection";
import { fetchRepositoryTestResults } from "../../service/repository-tests.service";

const baseResults = {
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
    { id: 100, run_external_id: "ext-100", commit_sha: "feedface1234", status: "success", branch: "main", started_at: "2026-06-15T01:00:00Z", finished_at: "2026-06-15T01:02:00Z" },
    { id: 99, run_external_id: "ext-99", commit_sha: "badfeed5678", status: "failed", branch: "feat/x", started_at: "2026-06-15T00:30:00Z", finished_at: "2026-06-15T00:31:00Z" },
  ],
};

describe("RepositoryTestsSection", () => {
  beforeEach(() => {
    vi.mocked(fetchRepositoryTestResults).mockReset();
  });

  it("renders pass rate, status distribution, and recent runs table", async () => {
    vi.mocked(fetchRepositoryTestResults).mockResolvedValue(baseResults);

    render(<RepositoryTestsSection repoId={1} />);

    // pass_rate = 0.933 → "93.3%"
    await waitFor(() => {
      expect(screen.getByTestId("tests-pass-rate")).toHaveTextContent("93.3%");
    });

    // status 분포 — 각 status 별 data-testid 가 있음
    expect(screen.getByTestId("tests-status-success")).toHaveTextContent("42");
    expect(screen.getByTestId("tests-status-failed")).toHaveTextContent("3");
    expect(screen.getByTestId("tests-status-running")).toHaveTextContent("1");
    expect(screen.getByTestId("tests-status-cancelled")).toHaveTextContent("1");
    expect(screen.getByTestId("tests-status-skipped")).toHaveTextContent("0");

    // recent table — first row commit SHA short form (7 chars)
    const table = screen.getByTestId("tests-recent-table");
    expect(table).toBeInTheDocument();
    expect(table).toHaveTextContent("feedfac"); // "feedface1234".slice(0,7)
    expect(table).toHaveTextContent("badfeed"); // "badfeed5678".slice(0,7)
    expect(table).toHaveTextContent("main");
    expect(table).toHaveTextContent("feat/x");

    // pie chart 가 pieData 로 mount 됨
    expect(screen.getByTestId("pie-chart")).toBeInTheDocument();
    const pieDataEl = screen.getByTestId("pie-data");
    const pieData = JSON.parse(pieDataEl.getAttribute("data-data") ?? "[]");
    expect(pieData).toHaveLength(4); // count > 0 만 — skipped/queued/unknown = 0 제외
  });

  it("pass_rate is null when no success+failed (only running/cancelled) — muted color", async () => {
    vi.mocked(fetchRepositoryTestResults).mockReset();
    vi.mocked(fetchRepositoryTestResults).mockResolvedValue({
      ...baseResults,
      totals: { success: 0, failed: 0, running: 1, cancelled: 1, skipped: 0, queued: 0, unknown: 0 },
      pass_rate: null,
      recent: [],
    });

    render(<RepositoryTestsSection repoId={1} />);

    await waitFor(() => {
      expect(screen.getByTestId("tests-pass-rate").textContent).toContain("—");
    });
    expect(screen.getByTestId("tests-pass-rate").className).toContain("text-muted-foreground");
  });

  it("colors pass_rate emerald / amber / red based on threshold", async () => {
    const cases: Array<{ rate: number; expectedClass: string }> = [
      { rate: 0.95, expectedClass: "text-emerald-600" },
      { rate: 0.75, expectedClass: "text-amber-600" },
      { rate: 0.5, expectedClass: "text-red-600" },
    ];

    for (const c of cases) {
      vi.mocked(fetchRepositoryTestResults).mockReset();
      vi.mocked(fetchRepositoryTestResults).mockResolvedValue({
        ...baseResults,
        pass_rate: c.rate,
      });
      const { unmount } = render(<RepositoryTestsSection repoId={1} />);
      await waitFor(() => {
        expect(screen.getByTestId("tests-pass-rate").className).toContain(c.expectedClass);
      });
      unmount();
    }
  });

  it("exposes 4 window options (7d / 30d / 90d / 1y) and refetches on change", async () => {
    const spy = vi.mocked(fetchRepositoryTestResults);
    spy.mockResolvedValue(baseResults);

    render(<RepositoryTestsSection repoId={1} />);

    await waitFor(() => {
      expect(screen.getByTestId("tests-pass-rate")).toHaveTextContent("93.3%");
    });

    const select = screen.getByLabelText("Window") as HTMLSelectElement;
    expect(select.value).toBe("30d");

    const user = userEvent.setup();
    await user.selectOptions(select, "7d");

    await waitFor(() => {
      const calls = spy.mock.calls.filter((c) => c[0] === 1);
      const lastCall = calls[calls.length - 1];
      expect(lastCall?.[1]).toEqual({ window: "7d", limit: 20 });
    });
  });

  it("renders 'No recent runs' when recent list is empty", async () => {
    vi.mocked(fetchRepositoryTestResults).mockReset();
    vi.mocked(fetchRepositoryTestResults).mockResolvedValue({
      ...baseResults,
      recent: [],
    });

    render(<RepositoryTestsSection repoId={1} />);

    await waitFor(() => {
      expect(screen.getByText(/No recent runs/i)).toBeInTheDocument();
    });
  });

  it("renders error fallback when fetch fails", async () => {
    vi.mocked(fetchRepositoryTestResults).mockReset();
    vi.mocked(fetchRepositoryTestResults).mockRejectedValue(new Error("network down"));

    render(<RepositoryTestsSection repoId={1} />);

    await waitFor(() => {
      expect(screen.getByText(/Failed to load repository test results/i)).toBeInTheDocument();
    });
    expect(screen.getByText(/network down/i)).toBeInTheDocument();
  });
});
