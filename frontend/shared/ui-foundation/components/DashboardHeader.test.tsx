import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { Zap } from "lucide-react";

vi.mock("framer-motion", () => {
  const React = require("react");
  type AnyProps = { children?: unknown; [k: string]: unknown };
  const motion = new Proxy(
    {},
    {
      get: (_target, tag) =>
        ({ children, ...props }: AnyProps) =>
          React.createElement(tag as string, props, children),
    },
  );
  return {
    motion,
    AnimatePresence: ({ children }: AnyProps) =>
      React.createElement(React.Fragment, null, children),
  };
});

import { DashboardHeader } from "./DashboardHeader";

describe("DashboardHeader (F-1)", () => {
  it("titlePrefix + titleGradient 가 한 h1 에 렌더된다", () => {
    render(
      <DashboardHeader
        titlePrefix="Welcome to"
        titleGradient="DevHub"
        subtitle="An internal portal"
      />,
    );
    const heading = screen.getByRole("heading", { level: 1 });
    expect(heading.textContent).toContain("Welcome to");
    expect(heading.textContent).toContain("DevHub");
  });

  it("subtitle 문자열이 렌더된다", () => {
    render(
      <DashboardHeader
        titlePrefix="A"
        titleGradient="B"
        subtitle="my subtitle"
      />,
    );
    expect(screen.getByText("my subtitle")).toBeInTheDocument();
  });

  it("subtitle 가 ReactNode 인 경우에도 렌더된다", () => {
    render(
      <DashboardHeader
        titlePrefix="A"
        titleGradient="B"
        subtitle={<span data-testid="sub">complex</span>}
      />,
    );
    expect(screen.getByTestId("sub")).toBeInTheDocument();
  });

  it("icon prop 이 제공되면 svg 가 렌더된다", () => {
    const { container } = render(
      <DashboardHeader
        titlePrefix="A"
        titleGradient="B"
        subtitle="sub"
        icon={Zap}
      />,
    );
    const svg = container.querySelector("svg");
    expect(svg).not.toBeNull();
  });

  it("icon prop 미지정 시에는 subtitle 영역에 svg 가 없다", () => {
    const { container } = render(
      <DashboardHeader
        titlePrefix="A"
        titleGradient="B"
        subtitle="sub"
      />,
    );
    // subtitle 영역 div (text-muted-foreground 클래스 보유)
    const subDiv = container.querySelector(".text-muted-foreground");
    expect(subDiv?.querySelector("svg")).toBeNull();
  });

  it("actions slot 의 React 노드가 렌더된다", () => {
    render(
      <DashboardHeader
        titlePrefix="A"
        titleGradient="B"
        subtitle="sub"
        actions={<button data-testid="action-btn">Add</button>}
      />,
    );
    expect(screen.getByTestId("action-btn")).toBeInTheDocument();
  });
});
