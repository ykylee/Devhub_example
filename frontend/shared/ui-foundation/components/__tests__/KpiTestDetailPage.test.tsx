// @vitest-environment happy-dom
import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { KpiTestDetailPage } from "../KpiTestDetailPage";

vi.mock("next/navigation", () => ({
  usePathname: vi.fn(() => "/platforms/pl-1"),
}));

describe("KpiTestDetailPage (2026-06-17 fix — issue 3 detail page wrapper)", () => {
  it("renders header with kind label (KPI)", () => {
    render(
      <KpiTestDetailPage entityType="platform" kind="kpi" entityId="pl-1">
        <div data-testid="child">child content</div>
      </KpiTestDetailPage>,
    );
    expect(screen.getByRole("heading", { name: /KPI/ })).toBeInTheDocument();
    expect(screen.getByTestId("child")).toBeInTheDocument();
  });

  it("renders header with kind label (Test Results)", () => {
    render(
      <KpiTestDetailPage entityType="project" kind="tests" entityId="pr-1" entityName="MyProject">
        <div data-testid="child-2">content</div>
      </KpiTestDetailPage>,
    );
    expect(screen.getByText("MyProject — Test Results")).toBeInTheDocument();
  });

  it("renders back link with correct base path for each entity type", () => {
    const { rerender } = render(
      <KpiTestDetailPage entityType="platform" kind="kpi" entityId="pl-1">
        <span />
      </KpiTestDetailPage>,
    );
    expect(screen.getByTestId("kpi-test-detail-back-platform-kpi")).toHaveAttribute("href", "/platforms/pl-1");

    rerender(
      <KpiTestDetailPage entityType="project" kind="tests" entityId="pr-1">
        <span />
      </KpiTestDetailPage>,
    );
    expect(screen.getByTestId("kpi-test-detail-back-project-tests")).toHaveAttribute("href", "/projects/pr-1");

    rerender(
      <KpiTestDetailPage entityType="repository" kind="kpi" entityId="42">
        <span />
      </KpiTestDetailPage>,
    );
    expect(screen.getByTestId("kpi-test-detail-back-repository-kpi")).toHaveAttribute("href", "/repositories/42");
  });
});
