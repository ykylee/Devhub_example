// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import React from "react";

// service mock — page 가 호출하는 fetchPlatformKPI 격리
vi.mock("@/domain/platform-lifecycle/service/platform-kpi.service", () => ({
  fetchPlatformKPI: vi.fn(),
}));

import { fetchPlatformKPI } from "@/domain/platform-lifecycle/service/platform-kpi.service";
import { KpiTestDetailPage } from "@/shared/ui-foundation/components/KpiTestDetailPage";
import { PlatformKPISection } from "@/domain/platform-lifecycle/view/PlatformKPISection";

// React 19 `use(params)` 가 vitest 의 happy-dom 에서 suspend → test 에서 sync wrapper 로 검증.
// page 자체는 RSC/server-component pattern (params: Promise) 이므로 production runtime 에서만 동작.
function PageUnderTest() {
  return (
    <KpiTestDetailPage entityType="platform" kind="kpi" entityId="pl-1">
      <PlatformKPISection platformId="pl-1" />
    </KpiTestDetailPage>
  );
}

describe("PlatformKpiPage (2026-06-17 fix — issue 3 detail page)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders back link + header + PlatformKPISection (sync wrapper)", async () => {
    vi.mocked(fetchPlatformKPI).mockResolvedValue({
      platform_id: "pl-1",
      window_from: "2026-05-18T00:00:00Z",
      window_to: "2026-06-17T00:00:00Z",
      weighted_at: "2026-06-17T00:00:00Z",
      weighted_quality_score: 0,
      weighted_build_success_rate: 0,
      open_pr_count: 0,
      merged_pr_count: 0,
      active_contributor_count: 0,
      linked_project_count: 0,
    });

    render(<PageUnderTest />);

    expect(screen.getByTestId("kpi-test-detail-back-platform-kpi")).toHaveAttribute("href", "/platforms/pl-1");
    await waitFor(() => {
      expect(screen.getByRole("heading", { name: /KPI/ })).toBeInTheDocument();
    });
    await waitFor(() => {
      expect(screen.getByTestId("platform-kpi-section")).toBeInTheDocument();
    });
  });
});
