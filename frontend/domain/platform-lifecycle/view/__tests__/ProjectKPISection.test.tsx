import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import React from "react";

// ProjectKPISection 컴포넌트 unit test (Sprint B follow-up, TC-PROJ-KPI-UI-01).
// 정공법: service 만 vi.mock 으로 stub. Recharts 미사용 (KPI 카드만).

vi.mock("../../service/project-kpi.service", () => ({
  fetchProjectKPI: vi.fn(),
}));

import { ProjectKPISection } from "../ProjectKPISection";
import { fetchProjectKPI } from "../../service/project-kpi.service";

const baseKpi = {
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
};

describe("ProjectKPISection", () => {
  beforeEach(() => {
    vi.mocked(fetchProjectKPI).mockReset();
  });

  it("renders loader initially, then weighted quality/build/pr/contributors/linkedRepos", async () => {
    vi.mocked(fetchProjectKPI).mockResolvedValue(baseKpi);

    render(<ProjectKPISection projectId="p-001" />);

    await waitFor(() => {
      expect(screen.getByTestId("project-kpi-quality-score")).toHaveTextContent("87.3");
    });
    expect(screen.getByTestId("project-kpi-build-success-rate")).toHaveTextContent("94.0%");
    expect(screen.getByTestId("project-kpi-open-pr")).toHaveTextContent("7");
    expect(screen.getByTestId("project-kpi-merged-pr")).toHaveTextContent("23");
    expect(screen.getByTestId("project-kpi-active-contributors")).toHaveTextContent("12");
    expect(screen.getByTestId("project-kpi-linked-repos")).toHaveTextContent("3");

    expect(vi.mocked(fetchProjectKPI)).toHaveBeenCalledWith("p-001", { windowDays: 30 });
  });

  it("colors weighted_quality_score emerald / amber / red", async () => {
    const cases: Array<{ score: number; expectedClass: string }> = [
      { score: 85, expectedClass: "text-emerald-600" },
      { score: 70, expectedClass: "text-amber-600" },
      { score: 50, expectedClass: "text-red-600" },
    ];

    for (const c of cases) {
      vi.mocked(fetchProjectKPI).mockReset();
      vi.mocked(fetchProjectKPI).mockResolvedValue({ ...baseKpi, weighted_quality_score: c.score });
      const { unmount } = render(<ProjectKPISection projectId="p-001" />);
      await waitFor(() => {
        expect(screen.getByTestId("project-kpi-quality-score").className).toContain(
          c.expectedClass,
        );
      });
      unmount();
    }
  });

  it("exposes 4 window options and refetches on change", async () => {
    const spy = vi.mocked(fetchProjectKPI);
    spy.mockResolvedValue(baseKpi);

    render(<ProjectKPISection projectId="p-001" />);

    await waitFor(() => {
      expect(screen.getByTestId("project-kpi-quality-score")).toHaveTextContent("87.3");
    });

    const select = screen.getByLabelText("Window") as HTMLSelectElement;
    expect(select.value).toBe("30");

    const user = userEvent.setup();
    await user.selectOptions(select, "7");

    await waitFor(() => {
      const calls = spy.mock.calls.filter((c) => c[0] === "p-001");
      const lastCall = calls[calls.length - 1];
      expect(lastCall?.[1]).toEqual({ windowDays: 7 });
    });
  });

  it("renders error fallback when fetch fails", async () => {
    vi.mocked(fetchProjectKPI).mockRejectedValue(new Error("boom"));

    render(<ProjectKPISection projectId="p-001" />);

    await waitFor(() => {
      expect(screen.getByText(/Failed to load project KPI/i)).toBeInTheDocument();
    });
    expect(screen.getByText(/boom/i)).toBeInTheDocument();
  });

  it("renders empty state when apiClient returns data=null", async () => {
    vi.mocked(fetchProjectKPI).mockResolvedValue(null);

    render(<ProjectKPISection projectId="p-001" />);

    await waitFor(() => {
      expect(screen.getByText(/No KPI data available/i)).toBeInTheDocument();
    });
  });
});
