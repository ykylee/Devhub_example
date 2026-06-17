import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

// RepositoryKPISection 컴포넌트 unit test (Sprint A follow-up, TC-REPO-KPI-UI-01).
//
// 정공법:
// - repository-integration/service/repository-kpi.service 를 vi.mock 으로
//   격리 (실제 apiClient 호출 방지). service 의 fetchRepositoryKPI 만 stub.
// - 실제 component 가 service 의 fetchRepositoryKPI 를 import 해서 사용.
// - happy-dom 환경 + Recharts mock (RepositoryDashboardView.test.tsx 패턴).

vi.mock("../../service/repository-kpi.service", () => ({
  fetchRepositoryKPI: vi.fn(),
}));

vi.mock("recharts", () => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const passthrough = (_props: any) => null;
  return {
    ResponsiveContainer: ({ children }: { children: React.ReactNode }) =>
      React.createElement("div", { className: "responsive-container-mock" }, children),
    AreaChart: ({ children, data }: { children?: React.ReactNode; data?: unknown }) =>
      React.createElement("div", { "data-testid": "area-chart", "data-data": JSON.stringify(data) }, children),
    Area: passthrough,
    XAxis: passthrough,
    YAxis: passthrough,
    CartesianGrid: passthrough,
    Tooltip: passthrough,
    PieChart: ({ children }: { children?: React.ReactNode }) =>
      React.createElement("div", { "data-testid": "pie-chart" }, children),
    Pie: ({ data }: { data?: unknown }) =>
      React.createElement("div", { "data-testid": "pie-data", "data-data": JSON.stringify(data) }),
    Cell: passthrough,
  };
});

import { RepositoryKPISection } from "../RepositoryKPISection";
import { fetchRepositoryKPI } from "../../service/repository-kpi.service";

const baseKpi = {
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
};

describe("RepositoryKPISection", () => {
  beforeEach(() => {
    vi.mocked(fetchRepositoryKPI).mockReset();
  });

  it("renders loader initially, then shows quality/build/pr/contributor metrics", async () => {
    const spy = vi.mocked(fetchRepositoryKPI);
    spy.mockResolvedValue(baseKpi);

    render(<RepositoryKPISection repoId={1} />);

    // quality_score = 87.3 → emerald
    await waitFor(() => {
      expect(screen.getByTestId("kpi-quality-score")).toHaveTextContent("87.3");
    });
    // build_success_rate = 0.94 → "94.0%"
    expect(screen.getByTestId("kpi-build-success-rate")).toHaveTextContent("94.0%");
    // open_pr / merged_pr
    expect(screen.getByTestId("kpi-open-pr")).toHaveTextContent("3");
    expect(screen.getByTestId("kpi-merged-pr")).toHaveTextContent("12");
    // active contributors
    expect(screen.getByTestId("kpi-active-contributors")).toHaveTextContent("4");

    expect(spy).toHaveBeenCalledWith(1, { windowDays: 30 });
  });

  it("colors quality_score emerald / amber / red based on threshold", async () => {
    const cases: Array<{ score: number | null; expectedClass: string }> = [
      { score: 85, expectedClass: "text-emerald-600" },
      { score: 70, expectedClass: "text-amber-600" },
      { score: 50, expectedClass: "text-red-600" },
      { score: null, expectedClass: "text-muted-foreground" },
    ];

    for (const c of cases) {
      vi.mocked(fetchRepositoryKPI).mockReset();
      vi.mocked(fetchRepositoryKPI).mockResolvedValue({
        ...baseKpi,
        quality_score: c.score,
      });
      const { unmount } = render(<RepositoryKPISection repoId={1} />);
      await waitFor(() => {
        expect(screen.getByTestId("kpi-quality-score").className).toContain(c.expectedClass);
      });
      unmount();
    }
  });

  it("exposes 4 window options (7d / 30d / 90d / 1y) and refetches on change", async () => {
    const spy = vi.mocked(fetchRepositoryKPI);
    spy.mockResolvedValue(baseKpi);

    render(<RepositoryKPISection repoId={1} />);

    await waitFor(() => {
      expect(screen.getByTestId("kpi-quality-score")).toHaveTextContent("87.3");
    });

    const select = screen.getByLabelText("Window") as HTMLSelectElement;
    expect(select.value).toBe("30");

    const user = userEvent.setup();
    await user.selectOptions(select, "7");

    await waitFor(() => {
      const calls = spy.mock.calls.filter((c) => c[0] === 1);
      const lastCall = calls[calls.length - 1];
      expect(lastCall?.[1]).toEqual({ windowDays: 7 });
    });
  });

  it("renders error fallback when fetch fails", async () => {
    vi.mocked(fetchRepositoryKPI).mockReset();
    vi.mocked(fetchRepositoryKPI).mockRejectedValue(new Error("boom"));

    render(<RepositoryKPISection repoId={1} />);

    await waitFor(() => {
      expect(screen.getByText(/Failed to load repository KPI/i)).toBeInTheDocument();
    });
    // error 메시지 본문 노출
    expect(screen.getByText(/boom/i)).toBeInTheDocument();
  });

  it("renders empty state when apiClient returns data=null (no data)", async () => {
    vi.mocked(fetchRepositoryKPI).mockReset();
    vi.mocked(fetchRepositoryKPI).mockResolvedValue(null);

    render(<RepositoryKPISection repoId={1} />);

    await waitFor(() => {
      expect(screen.getByText(/No KPI data available/i)).toBeInTheDocument();
    });
  });
});
