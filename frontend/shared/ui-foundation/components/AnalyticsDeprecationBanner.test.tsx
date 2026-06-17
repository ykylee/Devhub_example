import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { AnalyticsDeprecationBanner } from "./AnalyticsDeprecationBanner";

describe("AnalyticsDeprecationBanner (Sprint E)", () => {
  it("data-testid=analytics-deprecation-banner element 가 렌더된다", () => {
    render(<AnalyticsDeprecationBanner />);
    expect(screen.getByTestId("analytics-deprecation-banner")).toBeInTheDocument();
  });

  it("role=status 와 aria-live=polite 속성을 가진다 (a11y)", () => {
    render(<AnalyticsDeprecationBanner />);
    const banner = screen.getByTestId("analytics-deprecation-banner");
    expect(banner).toHaveAttribute("role", "status");
    expect(banner).toHaveAttribute("aria-live", "polite");
  });

  it("cross-reference picker 안내 문구가 표시된다", () => {
    render(<AnalyticsDeprecationBanner />);
    expect(
      screen.getByText(/Cross-reference picker/i),
    ).toBeInTheDocument();
    expect(
      screen.getByText(/도메인 상세 페이지의 KPI\/Tests sub-section 이 1차 진입점/i),
    ).toBeInTheDocument();
  });

  it("legacy 위젯 deprecate 안내 문구가 표시된다", () => {
    render(<AnalyticsDeprecationBanner />);
    expect(screen.getByText(/legacy 위젯/i)).toBeInTheDocument();
    expect(screen.getByText(/v0\.1\.0 이전 구현/i)).toBeInTheDocument();
  });
});
