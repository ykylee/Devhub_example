import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, act } from "@testing-library/react";

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

import { useStore } from "@/lib/store";
import { LogoutOverlay } from "./LogoutOverlay";

describe("LogoutOverlay (F-1)", () => {
  beforeEach(() => {
    act(() => {
      useStore.setState({ isLoggingOut: false });
    });
  });

  it("isLoggingOut=false 면 overlay 가 렌더되지 않는다", () => {
    render(<LogoutOverlay />);
    expect(screen.queryByText(/logging out/i)).toBeNull();
  });

  it("isLoggingOut=true 가 되면 overlay 와 메시지가 노출된다", () => {
    act(() => {
      useStore.setState({ isLoggingOut: true });
    });
    render(<LogoutOverlay />);
    expect(screen.getByText(/logging out/i)).toBeInTheDocument();
    expect(screen.getByText(/securing your session/i)).toBeInTheDocument();
  });
});
