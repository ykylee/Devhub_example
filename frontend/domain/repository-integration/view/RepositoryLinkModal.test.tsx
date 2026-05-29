import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

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

const getSCMProviders = vi.fn();
const connectRepository = vi.fn();

vi.mock("@/domain/application-lifecycle/service/project.service", () => ({
  projectService: {
    getSCMProviders: (...args: unknown[]) => getSCMProviders(...args),
    connectRepository: (...args: unknown[]) => connectRepository(...args),
  },
}));

import { RepositoryLinkModal } from "./RepositoryLinkModal";

const baseProps = {
  applicationId: "app-1",
  onClose: vi.fn(),
  onLinked: vi.fn(),
};

beforeEach(() => {
  getSCMProviders.mockReset();
  connectRepository.mockReset();
  baseProps.onClose = vi.fn();
  baseProps.onLinked = vi.fn();
});

describe("RepositoryLinkModal", () => {
  it("renders skeleton while providers loading then provider options", async () => {
    let resolveFn: (v: unknown) => void = () => {};
    getSCMProviders.mockReturnValueOnce(new Promise((res) => { resolveFn = res; }));

    render(<RepositoryLinkModal {...baseProps} />);

    expect(screen.getByText("SCM Provider")).toBeInTheDocument();
    // loading skeleton — no select rendered yet
    expect(screen.queryByRole("combobox")).not.toBeInTheDocument();

    resolveFn([
      { provider_key: "github", display_name: "GitHub", enabled: true, adapter_version: "v1", created_at: "", updated_at: "" },
      { provider_key: "gitea", display_name: "Gitea", enabled: true, adapter_version: "v1", created_at: "", updated_at: "" },
    ]);

    await waitFor(() => {
      expect(screen.getByRole("combobox")).toBeInTheDocument();
    });
    expect(screen.getByRole("option", { name: "GitHub" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Gitea" })).toBeInTheDocument();
  });

  it("renders fallback option when getSCMProviders fails", async () => {
    const errSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    getSCMProviders.mockRejectedValueOnce(new Error("boom"));
    render(<RepositoryLinkModal {...baseProps} />);
    await waitFor(() => {
      expect(screen.getByRole("combobox")).toBeInTheDocument();
    });
    // No providers in state → default fallback option "GitHub (Default)"
    expect(screen.getByRole("option", { name: /GitHub \(Default\)/ })).toBeInTheDocument();
    errSpy.mockRestore();
  });

  it("submits with full_name + selected role and calls onLinked/onClose", async () => {
    const user = userEvent.setup();
    getSCMProviders.mockResolvedValueOnce([
      { provider_key: "github", display_name: "GitHub", enabled: true, adapter_version: "v1", created_at: "", updated_at: "" },
    ]);
    const linkedRepo = {
      application_id: "app-1",
      repo_provider: "github",
      repo_full_name: "x/y",
      role: "primary",
      sync_status: "requested",
      linked_at: "2026-05-29T00:00:00Z",
    };
    connectRepository.mockResolvedValueOnce(linkedRepo);

    render(<RepositoryLinkModal {...baseProps} />);
    await waitFor(() => {
      expect(screen.getByRole("combobox")).toBeInTheDocument();
    });

    // Type repo full_name (use fireEvent.change to avoid `/` keyboard mapping issue)
    const input = screen.getByPlaceholderText(/e\.g\. devhub\/backend-core/i);
    fireEvent.change(input, { target: { value: "x/y" } });

    // Click "primary" role
    await user.click(screen.getByRole("button", { name: "primary" }));

    // Submit
    await user.click(screen.getByRole("button", { name: /Link Repository/i }));

    await waitFor(() => {
      expect(connectRepository).toHaveBeenCalledWith("app-1", {
        repo_provider: "github",
        repo_full_name: "x/y",
        role: "primary",
      });
    });
    expect(baseProps.onLinked).toHaveBeenCalledWith(linkedRepo);
    expect(baseProps.onClose).toHaveBeenCalled();
  });

  it("displays inline error when connectRepository fails", async () => {
    const user = userEvent.setup();
    getSCMProviders.mockResolvedValueOnce([
      { provider_key: "github", display_name: "GitHub", enabled: true, adapter_version: "v1", created_at: "", updated_at: "" },
    ]);
    connectRepository.mockRejectedValueOnce(new Error("repo already linked"));

    render(<RepositoryLinkModal {...baseProps} />);
    await waitFor(() => {
      expect(screen.getByRole("combobox")).toBeInTheDocument();
    });

    const input = screen.getByPlaceholderText(/e\.g\. devhub\/backend-core/i);
    fireEvent.change(input, { target: { value: "x/y" } });
    await user.click(screen.getByRole("button", { name: /Link Repository/i }));

    await waitFor(() => {
      expect(screen.getByText(/repo already linked/i)).toBeInTheDocument();
    });
    expect(baseProps.onLinked).not.toHaveBeenCalled();
  });

  it("calls onClose when Cancel button is clicked", async () => {
    const user = userEvent.setup();
    getSCMProviders.mockResolvedValueOnce([]);
    render(<RepositoryLinkModal {...baseProps} />);
    await waitFor(() => {
      expect(screen.getByRole("combobox")).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: "Cancel" }));
    expect(baseProps.onClose).toHaveBeenCalled();
  });

  it("calls onClose when close icon (X) is clicked", async () => {
    const user = userEvent.setup();
    getSCMProviders.mockResolvedValueOnce([]);
    render(<RepositoryLinkModal {...baseProps} />);
    await waitFor(() => {
      expect(screen.getByRole("combobox")).toBeInTheDocument();
    });

    // The X close button has no accessible name — find by lucide X icon container button
    const buttons = screen.getAllByRole("button");
    // Find the close header button — it precedes the form
    // We can rely on it being the FIRST button (header X) since role buttons (primary/sub/shared) come later
    await user.click(buttons[0]);
    expect(baseProps.onClose).toHaveBeenCalled();
  });
});
