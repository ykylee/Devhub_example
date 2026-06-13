import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import React from "react";

vi.mock("framer-motion", () => {
  type AnyProps = { children?: React.ReactNode; [k: string]: unknown };
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

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: vi.fn(), push: vi.fn(), refresh: vi.fn() }),
  usePathname: () => "/admin/inbound-source",
  useSearchParams: () => new URLSearchParams(),
}));

// ============================================================================
// X-2 multi-provider 운영 UI widget 4종 unit test (Vitest)
// ============================================================================

describe("InboundSourceTypeSelector", () => {
  beforeEach(() => vi.clearAllMocks());

  it("renders 4 options (Disabled / Gitea / Jira / Other) and Disabled hint when value is empty", async () => {
    const { InboundSourceTypeSelector } = await import("./InboundSourceTypeSelector");
    render(React.createElement(InboundSourceTypeSelector, { value: "", onChange: () => {} }));
    expect(screen.getByText(/Inbound Source Type/i)).toBeInTheDocument();
    expect(screen.getByText(/Disabled \(manual routing\)/i)).toBeInTheDocument();
  });

  it("shows Gitea hint when value=gitea", async () => {
    const { InboundSourceTypeSelector } = await import("./InboundSourceTypeSelector");
    render(React.createElement(InboundSourceTypeSelector, { value: "gitea", onChange: () => {} }));
    expect(screen.getByText(/GITEA-\\\d\+ external_ref \+ X-Gitea-\* header \+ provider_type='scm'/i)).toBeInTheDocument();
  });

  it("calls onChange when select changes", async () => {
    const onChange = vi.fn();
    const { InboundSourceTypeSelector } = await import("./InboundSourceTypeSelector");
    render(React.createElement(InboundSourceTypeSelector, { value: "", onChange }));
    fireEvent.change(screen.getByLabelText(/Provider Type/i), { target: { value: "jira" } });
    expect(onChange).toHaveBeenCalledWith("jira");
  });
});

describe("InboundSourceConfigEditor", () => {
  beforeEach(() => vi.clearAllMocks());

  it("renders config as JSONB text", async () => {
    const { InboundSourceConfigEditor } = await import("./InboundSourceConfigEditor");
    const config = { custom_external_ref_pattern: "^CUSTOM-\\d+$" };
    render(React.createElement(InboundSourceConfigEditor, { config, onSave: () => {} }));
    const textarea = screen.getByRole("textbox") as HTMLTextAreaElement;
    expect(textarea.value).toContain("CUSTOM-");
  });

  it("shows parse error for invalid JSON", async () => {
    const { InboundSourceConfigEditor } = await import("./InboundSourceConfigEditor");
    render(React.createElement(InboundSourceConfigEditor, { config: {}, onSave: () => {} }));
    const textarea = screen.getByRole("textbox");
    fireEvent.change(textarea, { target: { value: "{invalid json" } });
    expect(screen.getByText(/invalid JSON/i)).toBeInTheDocument();
  });

  it("disables Save button when parse error", async () => {
    const { InboundSourceConfigEditor } = await import("./InboundSourceConfigEditor");
    render(React.createElement(InboundSourceConfigEditor, { config: {}, onSave: () => {} }));
    const textarea = screen.getByRole("textbox");
    fireEvent.change(textarea, { target: { value: "{not json" } });
    const saveBtn = screen.getByRole("button", { name: /Save Config/i });
    expect(saveBtn).toBeDisabled();
  });
});

describe("PatternPreview", () => {
  beforeEach(() => vi.clearAllMocks());

  it("renders disabled message when type is empty", async () => {
    const { PatternPreview } = await import("./PatternPreview");
    render(React.createElement(PatternPreview, { type: "", customPattern: "" }));
    expect(screen.getByText(/pattern preview 미표시/i)).toBeInTheDocument();
  });

  it("shows provider-default pattern for gitea", async () => {
    const { PatternPreview } = await import("./PatternPreview");
    render(React.createElement(PatternPreview, { type: "gitea", customPattern: "" }));
    expect(screen.getByText(/pattern: \^GITEA-\\d\+\$/i)).toBeInTheDocument();
  });

  it("shows custom pattern for other", async () => {
    const { PatternPreview } = await import("./PatternPreview");
    render(
      React.createElement(PatternPreview, {
        type: "other",
        customPattern: "^CUSTOM-\\d+$",
      }),
    );
    expect(screen.getByText(/pattern: \^CUSTOM-\\d\+\$/i)).toBeInTheDocument();
  });
});

describe("InboundSourceManager", () => {
  beforeEach(() => vi.clearAllMocks());

  it("renders 4 widgets in 2x2 grid", async () => {
    const { InboundSourceManager } = await import("./InboundSourceManager");
    const platforms = [
      {
        platform_id: "p1",
        platform_key: "gitea-main",
        platform_name: "Gitea Main",
        inbound_source_type: "gitea" as const,
        inbound_source_config: {},
        updated_at: "2026-06-13T10:00:00Z",
      },
    ];
    render(React.createElement(InboundSourceManager, { platforms, onSave: vi.fn() }));
    expect(screen.getByText(/Inbound Source Type/i)).toBeInTheDocument();
    expect(screen.getByText(/Inbound Source Config \(JSONB\)/i)).toBeInTheDocument();
    expect(screen.getByText(/Pattern Preview/i)).toBeInTheDocument();
  });

  it("calls onSave with updated config", async () => {
    const onSave = vi.fn().mockResolvedValue(undefined);
    const { InboundSourceManager } = await import("./InboundSourceManager");
    const platforms = [
      {
        platform_id: "p1",
        platform_key: "gitea-main",
        platform_name: "Gitea Main",
        inbound_source_type: "gitea" as const,
        inbound_source_config: {},
        updated_at: "2026-06-13T10:00:00Z",
      },
    ];
    render(React.createElement(InboundSourceManager, { platforms, onSave }));
    const saveBtn = screen.getByRole("button", { name: /Save Inbound Source/i });
    fireEvent.click(saveBtn);
    await waitFor(() => {
      expect(onSave).toHaveBeenCalled();
    });
  });
});
