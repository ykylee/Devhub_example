// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import React from "react";

vi.mock("@/domain/repository-integration/service/repository-kpi.service", () => ({
  fetchRepositoryKPI: vi.fn(),
}));

import { fetchRepositoryKPI } from "@/domain/repository-integration/service/repository-kpi.service";
import { KpiTestDetailPage } from "@/shared/ui-foundation/components/KpiTestDetailPage";
import { RepositoryKPISection } from "@/domain/repository-integration/view/RepositoryKPISection";

function PageUnderTest() {
  return (
    <KpiTestDetailPage entityType="repository" kind="kpi" entityId="42">
      <RepositoryKPISection repoId={42} />
    </KpiTestDetailPage>
  );
}

describe("RepositoryKpiPage (2026-06-17 fix — issue 3 detail page)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders back link + header + RepositoryKPISection (sync wrapper)", async () => {
    vi.mocked(fetchRepositoryKPI).mockResolvedValue({
      repository_id: 42,
      window_from: "2026-05-18T00:00:00Z",
      window_to: "2026-06-17T00:00:00Z",
      quality_score: 0,
      build_success_rate: 0,
      open_pr_count: 0,
      merged_pr_count: 0,
      active_contributor_count: 0,
      history_trend: [],
    });

    render(<PageUnderTest />);

    expect(screen.getByTestId("kpi-test-detail-back-repository-kpi")).toHaveAttribute("href", "/repositories/42");
    await waitFor(() => {
      expect(screen.getByRole("heading", { name: /KPI/ })).toBeInTheDocument();
    });
  });
});
