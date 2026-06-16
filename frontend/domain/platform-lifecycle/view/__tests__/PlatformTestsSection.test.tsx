// PlatformTestsSection component unit test (Sprint C, TC-PLAT-TESTS-COMP-01).
//
// 정공법: ProjectTestsSection 가 import 하는 platform-tests.service 의
// fetchPlatformTestResults 만 vi.mock 으로 stub. ProjectTestsSection.test.tsx 와 동일 패턴.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

vi.mock("../../service/platform-tests.service", () => ({
  fetchPlatformTestResults: vi.fn(),
}));

import { fetchPlatformTestResults } from "../../service/platform-tests.service";
import { PlatformTestsSection } from "../PlatformTestsSection";

const baseResults = {
  platform_id: "pl-1",
  window_from: "2026-05-16T00:00:00Z",
  window_to: "2026-06-15T00:00:00Z",
  weighted_pass_rate: 0.93,
  totals: {
    success: 145, failed: 8, running: 1, cancelled: 2,
    skipped: 0, queued: 0, unknown: 0,
  },
  recent: [
    { id: 100, project_id: "p-001", project_full_name: "API Modernization", repository_id: 1, repository_full_name: "org/repo-a", run_external_id: "ext-100", commit_sha: "feedface1234", status: "success", branch: "main", started_at: "2026-06-15T01:00:00Z", finished_at: "2026-06-15T01:02:00Z" },
    { id: 99, project_id: "p-002", project_full_name: "UI Refresh", repository_id: 2, repository_full_name: "org/repo-b", run_external_id: "ext-99", commit_sha: "badfeed5678", status: "failed", branch: "feat/x", started_at: "2026-06-15T00:30:00Z", finished_at: "2026-06-15T00:31:00Z" },
  ],
};

describe("PlatformTestsSection", () => {
  beforeEach(() => {
    vi.mocked(fetchPlatformTestResults).mockReset();
  });

  afterEach(() => {
    cleanup();
  });

  it("renders loading state initially", () => {
    vi.mocked(fetchPlatformTestResults).mockImplementation(() => new Promise(() => {}));
    render(<PlatformTestsSection platformId="pl-1" />);
    expect(screen.getByText(/Loading platform test results/i)).toBeDefined();
  });

  it("renders pass rate + total + recent list on happy path", async () => {
    vi.mocked(fetchPlatformTestResults).mockResolvedValue(baseResults);
    render(<PlatformTestsSection platformId="pl-1" />);
    await waitFor(() =>
      expect(screen.getByTestId("platform-tests-pass-rate").textContent).toContain("93.0%"),
    );
    expect(screen.getByTestId("platform-tests-total").textContent).toContain("156");
    const recent = screen.getByTestId("platform-tests-recent");
    expect(recent.textContent).toContain("API Modernization");
    expect(recent.textContent).toContain("org/repo-a");
    expect(recent.textContent).toContain("UI Refresh");
  });

  it("renders em-dash when weighted_pass_rate is null", async () => {
    vi.mocked(fetchPlatformTestResults).mockResolvedValue({
      ...baseResults,
      weighted_pass_rate: null,
      recent: [],
    });
    render(<PlatformTestsSection platformId="pl-empty" />);
    await waitFor(() => {
      expect(screen.getByTestId("platform-tests-pass-rate").textContent).toContain("—");
    });
  });

  it("shows error on fetch failure", async () => {
    vi.mocked(fetchPlatformTestResults).mockRejectedValue(new Error("api 500"));
    render(<PlatformTestsSection platformId="pl-1" />);
    await waitFor(() => expect(screen.getByText(/Failed to load platform test results/i)).toBeDefined());
    expect(screen.getByText(/api 500/i)).toBeDefined();
  });

  it("refetches when window selector changes", async () => {
    vi.mocked(fetchPlatformTestResults).mockResolvedValue(baseResults);
    const user = userEvent.setup();
    render(<PlatformTestsSection platformId="pl-1" />);
    await waitFor(() =>
      expect(screen.getByTestId("platform-tests-pass-rate").textContent).toContain("93.0%"),
    );
    expect(fetchPlatformTestResults).toHaveBeenCalledWith("pl-1", {
      window: "30d",
      limit: 20,
    });

    await user.selectOptions(screen.getByLabelText(/Window/i), "7d");
    await waitFor(() =>
      expect(fetchPlatformTestResults).toHaveBeenCalledWith("pl-1", {
        window: "7d",
        limit: 20,
      }),
    );
  });
});
