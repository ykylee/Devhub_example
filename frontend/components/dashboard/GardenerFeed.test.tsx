import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";

// Mock framer-motion -> render children as plain DOM.
vi.mock("framer-motion", () => {
  const React = require("react");
  type AnyProps = { children?: unknown; [k: string]: unknown };
  const motion = new Proxy(
    {},
    {
      get: (_target, tag) =>
        ({ children, ...props }: AnyProps) => {
          // Drop framer-motion-only props that React would warn about.
          const safeProps: Record<string, unknown> = {};
          for (const key of Object.keys(props)) {
            if (["initial", "animate", "exit", "transition", "layout", "whileHover", "whileTap"].includes(key)) continue;
            safeProps[key] = (props as Record<string, unknown>)[key];
          }
          return React.createElement(tag as string, safeProps, children);
        },
    },
  );
  return {
    motion,
    AnimatePresence: ({ children }: AnyProps) => React.createElement(React.Fragment, null, children),
  };
});

const getSuggestions = vi.fn();
const applySuggestion = vi.fn();

vi.mock("@/domain/platform-lifecycle/service/gardener.service", () => ({
  gardenerService: {
    getSuggestions: (...args: unknown[]) => getSuggestions(...args),
    applySuggestion: (...args: unknown[]) => applySuggestion(...args),
  },
}));

const addToast = vi.fn();
vi.mock("@/lib/store", () => ({
  useStore: () => ({ addToast }),
}));

import { GardenerFeed } from "./GardenerFeed";
import type { Suggestion } from "@/domain/platform-lifecycle/service/gardener.service";

function makeSuggestion(overrides: Partial<Suggestion> = {}): Suggestion {
  return {
    id: "sug-1",
    title: "Suggestion One",
    description: "Description one body.",
    type: "optimization",
    impact: "low",
    auto_fixable: true,
    created_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

beforeEach(() => {
  getSuggestions.mockReset();
  applySuggestion.mockReset();
  addToast.mockReset();
});

describe("GardenerFeed", () => {
  it("renders skeleton placeholders during initial load", () => {
    // Never-resolving promise -> stays in loading state.
    getSuggestions.mockReturnValue(new Promise(() => {}));
    const { container } = render(<GardenerFeed />);
    const skeletons = container.querySelectorAll(".animate-pulse");
    expect(skeletons.length).toBeGreaterThan(0);
    expect(screen.queryByText(/AI Gardener/)).not.toBeInTheDocument();
  });

  it("renders header + empty-state when service returns no suggestions", async () => {
    getSuggestions.mockResolvedValueOnce([]);
    render(<GardenerFeed />);
    await waitFor(() => {
      expect(screen.getByText(/AI Gardener/)).toBeInTheDocument();
    });
    expect(screen.getByText(/All systems are optimized/i)).toBeInTheDocument();
    expect(screen.getByText(/0 Active Insights/i)).toBeInTheDocument();
  });

  it("renders suggestion list + active count when service returns rows", async () => {
    getSuggestions.mockResolvedValueOnce([
      makeSuggestion({ id: "s1", title: "Optimize CPU", type: "optimization", impact: "low" }),
      makeSuggestion({ id: "s2", title: "Patch CVE", type: "security", impact: "high" }),
      makeSuggestion({ id: "s3", title: "Auto-scale", type: "scaling", impact: "medium" }),
    ]);
    render(<GardenerFeed />);
    await waitFor(() => {
      expect(screen.getByText("Optimize CPU")).toBeInTheDocument();
    });
    expect(screen.getByText("Patch CVE")).toBeInTheDocument();
    expect(screen.getByText("Auto-scale")).toBeInTheDocument();
    expect(screen.getByText(/3 Active Insights/i)).toBeInTheDocument();
  });

  it("renders the impact label per suggestion (low/medium/high)", async () => {
    getSuggestions.mockResolvedValueOnce([
      makeSuggestion({ id: "s1", impact: "low" }),
      makeSuggestion({ id: "s2", impact: "medium" }),
      makeSuggestion({ id: "s3", impact: "high" }),
    ]);
    render(<GardenerFeed />);
    await waitFor(() => {
      expect(screen.getByText(/low Impact/i)).toBeInTheDocument();
    });
    expect(screen.getByText(/medium Impact/i)).toBeInTheDocument();
    expect(screen.getByText(/high Impact/i)).toBeInTheDocument();
  });

  it("logs an error + still leaves loading state when getSuggestions rejects", async () => {
    const errSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    getSuggestions.mockRejectedValueOnce(new Error("boom"));
    render(<GardenerFeed />);
    await waitFor(() => {
      // After rejection, finally block clears loading -> header renders.
      expect(screen.getByText(/AI Gardener/)).toBeInTheDocument();
    });
    // No suggestions -> empty state.
    expect(screen.getByText(/All systems are optimized/i)).toBeInTheDocument();
    expect(errSpy).toHaveBeenCalledWith("Failed to fetch suggestions");
    errSpy.mockRestore();
  });

  it("removes the suggestion + toasts success when Apply succeeds", async () => {
    getSuggestions.mockResolvedValueOnce([
      makeSuggestion({ id: "s1", title: "Optimize CPU" }),
      makeSuggestion({ id: "s2", title: "Patch CVE" }),
    ]);
    applySuggestion.mockResolvedValueOnce({ command_id: "cmd-1", status: "queued" });
    render(<GardenerFeed />);
    await waitFor(() => {
      expect(screen.getByText("Optimize CPU")).toBeInTheDocument();
    });

    const buttons = screen.getAllByRole("button", { name: /Apply Optimization/i });
    fireEvent.click(buttons[0]);

    await waitFor(() => {
      expect(applySuggestion).toHaveBeenCalledWith("s1");
    });
    await waitFor(() => {
      expect(screen.queryByText("Optimize CPU")).not.toBeInTheDocument();
    });
    expect(addToast).toHaveBeenCalledWith("Applying AI suggestion: Optimize CPU", "info");
    expect(addToast).toHaveBeenCalledWith("Suggestion applied successfully.", "success");
    // Other suggestion still rendered.
    expect(screen.getByText("Patch CVE")).toBeInTheDocument();
  });

  it("toasts error + keeps the suggestion when Apply rejects", async () => {
    getSuggestions.mockResolvedValueOnce([makeSuggestion({ id: "s1", title: "Optimize CPU" })]);
    applySuggestion.mockRejectedValueOnce(new Error("api"));
    render(<GardenerFeed />);
    await waitFor(() => {
      expect(screen.getByText("Optimize CPU")).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole("button", { name: /Apply Optimization/i }));
    await waitFor(() => {
      expect(addToast).toHaveBeenCalledWith("Failed to apply suggestion.", "error");
    });
    // Suggestion remains because filter only happens on success.
    expect(screen.getByText("Optimize CPU")).toBeInTheDocument();
  });

  it("renders one Apply button per suggestion", async () => {
    getSuggestions.mockResolvedValueOnce([
      makeSuggestion({ id: "s1" }),
      makeSuggestion({ id: "s2", title: "Second" }),
      makeSuggestion({ id: "s3", title: "Third" }),
    ]);
    render(<GardenerFeed />);
    await waitFor(() => {
      expect(screen.getByText("Second")).toBeInTheDocument();
    });
    const buttons = screen.getAllByRole("button", { name: /Apply Optimization/i });
    expect(buttons).toHaveLength(3);
  });
});
