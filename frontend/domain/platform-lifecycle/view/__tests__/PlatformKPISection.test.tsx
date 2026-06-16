// PlatformKPISection component unit test (Sprint C, TC-PLAT-KPI-COMP-01).
//
// 정공법: ProjectKPISection 가 import 하는 platform-kpi.service 의
// fetchPlatformKPI 만 vi.mock 으로 stub. ProjectKPISection.test.tsx 와 동일 패턴.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

vi.mock("../../service/platform-kpi.service", () => ({
  fetchPlatformKPI: vi.fn(),
}));

import { fetchPlatformKPI } from "../../service/platform-kpi.service";
import { PlatformKPISection } from "../PlatformKPISection";

const baseKPI = {
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
};

describe("PlatformKPISection", () => {
  beforeEach(() => {
    vi.mocked(fetchPlatformKPI).mockReset();
  });

  afterEach(() => {
    cleanup();
  });

  it("renders the loading state initially", () => {
    vi.mocked(fetchPlatformKPI).mockImplementation(() => new Promise(() => {}));
    render(<PlatformKPISection platformId="pl-1" />);
    expect(screen.getByText(/Loading platform KPI/i)).toBeDefined();
  });

  it("renders KPI cards on happy path", async () => {
    vi.mocked(fetchPlatformKPI).mockResolvedValue(baseKPI);
    render(<PlatformKPISection platformId="pl-1" />);
    await waitFor(() =>
      expect(screen.getByTestId("platform-kpi-quality-score").textContent).toContain("85.0"),
    );
    expect(screen.getByTestId("platform-kpi-build-success-rate").textContent).toContain("90.0%");
    expect(screen.getByTestId("platform-kpi-open-pr").textContent).toContain("5");
    expect(screen.getByTestId("platform-kpi-merged-pr").textContent).toContain("15");
    expect(screen.getByTestId("platform-kpi-active-contributors").textContent).toContain("8");
    expect(screen.getByTestId("platform-kpi-linked-projects").textContent).toContain("3");
  });

  it("shows 'no data' when fetch returns null", async () => {
    vi.mocked(fetchPlatformKPI).mockResolvedValue(null);
    render(<PlatformKPISection platformId="pl-empty" />);
    await waitFor(() => expect(screen.getByText(/No KPI data available/i)).toBeDefined());
  });

  it("shows error message on fetch failure", async () => {
    vi.mocked(fetchPlatformKPI).mockRejectedValue(new Error("network down"));
    render(<PlatformKPISection platformId="pl-1" />);
    await waitFor(() => expect(screen.getByText(/Failed to load platform KPI/i)).toBeDefined());
    expect(screen.getByText(/network down/i)).toBeDefined();
  });

  it("refetches when window selector changes", async () => {
    vi.mocked(fetchPlatformKPI).mockResolvedValue(baseKPI);
    const user = userEvent.setup();
    render(<PlatformKPISection platformId="pl-1" />);
    await waitFor(() =>
      expect(screen.getByTestId("platform-kpi-quality-score").textContent).toContain("85.0"),
    );
    expect(fetchPlatformKPI).toHaveBeenCalledWith("pl-1", { windowDays: 30 });

    await user.selectOptions(screen.getByLabelText(/Window/i), "7");
    await waitFor(() => expect(fetchPlatformKPI).toHaveBeenCalledWith("pl-1", { windowDays: 7 }));
  });

  it("applies correct color tier for low quality score (< 60)", async () => {
    vi.mocked(fetchPlatformKPI).mockResolvedValue({
      ...baseKPI,
      weighted_quality_score: 45.0,
    });
    render(<PlatformKPISection platformId="pl-1" />);
    await waitFor(() => {
      const node = screen.getByTestId("platform-kpi-quality-score");
      expect(node.className).toMatch(/text-red/);
    });
  });
});
