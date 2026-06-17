// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import React from "react";

vi.mock("@/domain/platform-lifecycle/service/project-kpi.service", () => ({
  fetchProjectKPI: vi.fn(),
}));

import { fetchProjectKPI } from "@/domain/platform-lifecycle/service/project-kpi.service";
import { KpiTestDetailPage } from "@/shared/ui-foundation/components/KpiTestDetailPage";
import { ProjectKPISection } from "@/domain/platform-lifecycle/view/ProjectKPISection";

function PageUnderTest() {
  return (
    <KpiTestDetailPage entityType="project" kind="kpi" entityId="pr-1">
      <ProjectKPISection projectId="pr-1" />
    </KpiTestDetailPage>
  );
}

describe("ProjectKpiPage (2026-06-17 fix — issue 3 detail page)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders back link + header + ProjectKPISection (sync wrapper)", async () => {
    vi.mocked(fetchProjectKPI).mockResolvedValue({
      project_id: "pr-1",
      window_from: "2026-05-18T00:00:00Z",
      window_to: "2026-06-17T00:00:00Z",
      weighted_at: "2026-06-17T00:00:00Z",
      weighted_quality_score: 0,
      weighted_build_success_rate: 0,
      open_pr_count: 0,
      merged_pr_count: 0,
      active_contributor_count: 0,
      repository_count: 0,
      repositories: [],
    });

    render(<PageUnderTest />);

    expect(screen.getByTestId("kpi-test-detail-back-project-kpi")).toHaveAttribute("href", "/projects/pr-1");
    await waitFor(() => {
      expect(screen.getByRole("heading", { name: /KPI/ })).toBeInTheDocument();
    });
  });
});
