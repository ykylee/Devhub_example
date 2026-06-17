import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { KpiTestErrorState } from "../KpiTestErrorState";

// 2026-06-17 fix: backend platformStoreOrUnavailable (router.go) 가 503 + body
// {code: "platform_store_unavailable"} 반환. frontend 가 본 component 로
// 명확한 안내 ("Backend store is not initialized") + retry button 제공.
// 5 KPI/test endpoint 공통 사용.

describe("KpiTestErrorState", () => {
  it("renders title + message + retry button with data-testid prefix", () => {
    const onRetry = vi.fn();
    render(
      <KpiTestErrorState
        title="Failed to load platform KPI"
        message="Backend store is not initialized."
        onRetry={onRetry}
        testIdPrefix="platform-kpi"
      />,
    );
    expect(screen.getByText("Failed to load platform KPI")).toBeInTheDocument();
    expect(screen.getByText("Backend store is not initialized.")).toBeInTheDocument();
    expect(screen.getByTestId("platform-kpi-error")).toBeInTheDocument();
    expect(screen.getByTestId("platform-kpi-retry")).toBeInTheDocument();
  });

  it("invokes onRetry when retry button clicked", async () => {
    const onRetry = vi.fn();
    const user = userEvent.setup();
    render(
      <KpiTestErrorState
        title="Failed to load test results"
        message="Some error"
        onRetry={onRetry}
        testIdPrefix="repository-tests"
      />,
    );
    await user.click(screen.getByTestId("repository-tests-retry"));
    expect(onRetry).toHaveBeenCalledTimes(1);
  });

  it("renders different testIdPrefix for different sections (Platform vs Project vs Repository)", () => {
    const { rerender } = render(
      <KpiTestErrorState
        title="x"
        message="y"
        onRetry={vi.fn()}
        testIdPrefix="project-kpi"
      />,
    );
    expect(screen.getByTestId("project-kpi-error")).toBeInTheDocument();
    rerender(
      <KpiTestErrorState
        title="x"
        message="y"
        onRetry={vi.fn()}
        testIdPrefix="repository-tests"
      />,
    );
    expect(screen.getByTestId("repository-tests-error")).toBeInTheDocument();
  });
});
