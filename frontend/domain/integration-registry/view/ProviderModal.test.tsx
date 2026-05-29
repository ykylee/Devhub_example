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

const createProvider = vi.fn();
const updateProvider = vi.fn();
const testConnection = vi.fn();
vi.mock("@/lib/services/integration.service", () => ({
  integrationService: {
    createProvider: (...args: unknown[]) => createProvider(...args),
    updateProvider: (...args: unknown[]) => updateProvider(...args),
    testConnection: (...args: unknown[]) => testConnection(...args),
  },
}));

import { ProviderModal } from "./ProviderModal";
import type { IntegrationProvider } from "@/lib/services/integration.types";

const existing: IntegrationProvider = {
  provider_id: "p1",
  provider_key: "gitea-main",
  provider_type: "scm",
  display_name: "Gitea Main",
  enabled: true,
  auth_mode: "token",
  credentials_ref: "hmac_sha256:set",
  capabilities: ["pull", "webhook"],
  sync_status: "ok",
  last_sync_at: null,
  last_error_code: null,
  base_url: "https://gitea.example.com",
  api_token_set: true,
  auth_username: null,
  auth_client_id: null,
  auth_token_url: null,
  auth_secret_set: false,
  created_at: "",
  updated_at: "",
};

beforeEach(() => {
  createProvider.mockReset();
  updateProvider.mockReset();
  testConnection.mockReset();
});

describe("ProviderModal", () => {
  it("renders Register title and provider_key field in create mode", () => {
    render(<ProviderModal onClose={vi.fn()} onSaved={vi.fn()} />);
    expect(screen.getByText(/Register Provider/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Provider Key/)).toBeInTheDocument();
  });

  it("renders Edit title and hides provider_key field in edit mode", () => {
    render(
      <ProviderModal initial={existing} onClose={vi.fn()} onSaved={vi.fn()} />,
    );
    expect(screen.getByText(/Edit Provider/i)).toBeInTheDocument();
    expect(screen.queryByLabelText(/Provider Key/)).not.toBeInTheDocument();
    expect(screen.getByText(/gitea-main/)).toBeInTheDocument();
  });

  it("applies vendor preset and hides type/auth select in create mode (known vendor)", async () => {
    render(<ProviderModal onClose={vi.fn()} onSaved={vi.fn()} />);
    // Initially custom — type/auth select rendered
    expect(screen.getByLabelText(/Type/)).toBeInTheDocument();
    // Select 'Gitea' preset
    const presetSelect = screen.getByLabelText(/Vendor Template/) as HTMLSelectElement;
    fireEvent.change(presetSelect, { target: { value: "gitea" } });
    // type/auth select is hidden when known vendor
    expect(screen.queryByLabelText("Type *")).not.toBeInTheDocument();
    expect(screen.getByText(/Template Defaults/)).toBeInTheDocument();
  });

  it("validates required provider_key on submit (create)", async () => {
    const { container } = render(<ProviderModal onClose={vi.fn()} onSaved={vi.fn()} />);
    // Bypass HTML5 required by submitting the form directly.
    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);
    await waitFor(() => {
      expect(screen.getByText(/provider_key 는 필수입니다/)).toBeInTheDocument();
    });
  });

  it("validates required display_name on submit (create)", async () => {
    const { container } = render(<ProviderModal onClose={vi.fn()} onSaved={vi.fn()} />);
    fireEvent.change(screen.getByLabelText(/Provider Key/), {
      target: { value: "my-prov" },
    });
    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);
    await waitFor(() => {
      expect(screen.getByText(/display_name 은 필수입니다/)).toBeInTheDocument();
    });
  });

  it("validates required secret on submit (create)", async () => {
    const { container } = render(<ProviderModal onClose={vi.fn()} onSaved={vi.fn()} />);
    fireEvent.change(screen.getByLabelText(/Provider Key/), {
      target: { value: "my-prov" },
    });
    fireEvent.change(screen.getByLabelText(/Display Name/), {
      target: { value: "Prov One" },
    });
    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);
    await waitFor(() => {
      expect(screen.getByText(/secret 은 필수입니다/)).toBeInTheDocument();
    });
  });

  it("submits createProvider successfully and calls onSaved + onClose", async () => {
    const user = userEvent.setup();
    const onSaved = vi.fn();
    const onClose = vi.fn();
    const saved: IntegrationProvider = { ...existing, provider_id: "new-id" };
    createProvider.mockResolvedValueOnce(saved);

    render(<ProviderModal onClose={onClose} onSaved={onSaved} />);
    fireEvent.change(screen.getByLabelText(/Provider Key/), {
      target: { value: "my-prov" },
    });
    fireEvent.change(screen.getByLabelText(/Display Name/), {
      target: { value: "Prov One" },
    });
    fireEvent.change(screen.getByLabelText(/Secret \*/), {
      target: { value: "topsecret" },
    });
    await user.click(screen.getByRole("button", { name: /Register/ }));

    await waitFor(() => {
      expect(createProvider).toHaveBeenCalled();
    });
    expect(onSaved).toHaveBeenCalledWith(saved);
    expect(onClose).toHaveBeenCalled();
  });

  it("toggles secret visibility (Eye/EyeOff button)", async () => {
    render(<ProviderModal onClose={vi.fn()} onSaved={vi.fn()} />);
    const secretInput = document.getElementById("secret") as HTMLInputElement;
    expect(secretInput.type).toBe("password");
    // Show secret toggle button is sibling to the input within .relative wrapper.
    const wrapper = secretInput.parentElement;
    const toggle = wrapper?.querySelector("button");
    if (!toggle) throw new Error("toggle button not found");
    fireEvent.click(toggle);
    await waitFor(() => {
      const refreshed = document.getElementById("secret") as HTMLInputElement;
      expect(refreshed.type).toBe("text");
    });
  });

  it("toggles capabilities via checkbox", async () => {
    const user = userEvent.setup();
    render(<ProviderModal onClose={vi.fn()} onSaved={vi.fn()} />);
    // capability vocabulary includes webhook/pull/push/snapshot/sync
    const pullCheckbox = screen.getByLabelText("pull") as HTMLInputElement;
    expect(pullCheckbox.checked).toBe(false);
    await user.click(pullCheckbox);
    expect(pullCheckbox.checked).toBe(true);
    await user.click(pullCheckbox);
    expect(pullCheckbox.checked).toBe(false);
  });

  it("runs test connection successfully when reachable", async () => {
    const user = userEvent.setup();
    testConnection.mockResolvedValueOnce({
      reachable: true,
      status_code: 200,
      latency_ms: 42,
    });
    render(<ProviderModal onClose={vi.fn()} onSaved={vi.fn()} />);
    fireEvent.change(screen.getByLabelText(/Base URL/), {
      target: { value: "https://example.com" },
    });
    await user.click(screen.getByRole("button", { name: /^Test/ }));
    await waitFor(() => {
      expect(screen.getByText(/Reachable/)).toBeInTheDocument();
    });
  });

  it("shows unreachable in test connection result", async () => {
    const user = userEvent.setup();
    testConnection.mockResolvedValueOnce({
      reachable: false,
      error: "dns",
    });
    render(<ProviderModal onClose={vi.fn()} onSaved={vi.fn()} />);
    fireEvent.change(screen.getByLabelText(/Base URL/), {
      target: { value: "https://example.com" },
    });
    await user.click(screen.getByRole("button", { name: /^Test/ }));
    await waitFor(() => {
      expect(screen.getByText(/Unreachable: dns/)).toBeInTheDocument();
    });
  });

  it("shows test connection error message on exception", async () => {
    const user = userEvent.setup();
    testConnection.mockRejectedValueOnce(new Error("network"));
    render(<ProviderModal onClose={vi.fn()} onSaved={vi.fn()} />);
    fireEvent.change(screen.getByLabelText(/Base URL/), {
      target: { value: "https://example.com" },
    });
    await user.click(screen.getByRole("button", { name: /^Test/ }));
    await waitFor(() => {
      expect(screen.getByText(/network/)).toBeInTheDocument();
    });
  });

  it("disables Test button when base_url is empty", () => {
    render(<ProviderModal onClose={vi.fn()} onSaved={vi.fn()} />);
    const btn = screen.getByRole("button", { name: /^Test/ });
    expect(btn).toBeDisabled();
  });

  it("toggles enabled checkbox in edit mode", async () => {
    const user = userEvent.setup();
    render(
      <ProviderModal initial={existing} onClose={vi.fn()} onSaved={vi.fn()} />,
    );
    const enabledCb = screen.getByLabelText(/Enabled/) as HTMLInputElement;
    expect(enabledCb.checked).toBe(true);
    await user.click(enabledCb);
    expect(enabledCb.checked).toBe(false);
  });

  it("renders basic auth username + password fields when auth_mode is basic", () => {
    render(<ProviderModal onClose={vi.fn()} onSaved={vi.fn()} />);
    const authMode = screen.getByLabelText(/Auth Mode/) as HTMLSelectElement;
    fireEvent.change(authMode, { target: { value: "basic" } });
    expect(screen.getByLabelText(/Username/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Password/)).toBeInTheDocument();
  });

  it("renders oauth2 client_id + token_url fields when auth_mode is oauth2", () => {
    render(<ProviderModal onClose={vi.fn()} onSaved={vi.fn()} />);
    const authMode = screen.getByLabelText(/Auth Mode/) as HTMLSelectElement;
    fireEvent.change(authMode, { target: { value: "oauth2" } });
    expect(screen.getByLabelText(/Client ID/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Token URL/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Client Secret/)).toBeInTheDocument();
  });

  it("renders agent identifier field when auth_mode is agent", () => {
    render(<ProviderModal onClose={vi.fn()} onSaved={vi.fn()} />);
    const authMode = screen.getByLabelText(/Auth Mode/) as HTMLSelectElement;
    fireEvent.change(authMode, { target: { value: "agent" } });
    expect(screen.getByLabelText(/Agent Identifier/)).toBeInTheDocument();
  });

  it("submits updateProvider successfully in edit mode (secret blank = keep)", async () => {
    const user = userEvent.setup();
    const onSaved = vi.fn();
    const updated: IntegrationProvider = { ...existing, display_name: "Renamed" };
    updateProvider.mockResolvedValueOnce(updated);

    render(<ProviderModal initial={existing} onClose={vi.fn()} onSaved={onSaved} />);
    fireEvent.change(screen.getByLabelText(/Display Name/), {
      target: { value: "Renamed" },
    });
    await user.click(screen.getByRole("button", { name: /^Save$/ }));
    await waitFor(() => {
      expect(updateProvider).toHaveBeenCalledWith("p1", expect.objectContaining({
        display_name: "Renamed",
        credentials_ref: undefined, // blank = keep
      }));
    });
    expect(onSaved).toHaveBeenCalledWith(updated);
  });

  it("shows error when createProvider rejects", async () => {
    const user = userEvent.setup();
    createProvider.mockRejectedValueOnce(new Error("api error"));
    render(<ProviderModal onClose={vi.fn()} onSaved={vi.fn()} />);
    fireEvent.change(screen.getByLabelText(/Provider Key/), {
      target: { value: "p" },
    });
    fireEvent.change(screen.getByLabelText(/Display Name/), {
      target: { value: "x" },
    });
    fireEvent.change(screen.getByLabelText(/Secret \*/), {
      target: { value: "s" },
    });
    await user.click(screen.getByRole("button", { name: /Register/ }));
    await waitFor(() => {
      expect(screen.getByText("api error")).toBeInTheDocument();
    });
  });

  it("calls onClose when Cancel is clicked", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    render(<ProviderModal onClose={onClose} onSaved={vi.fn()} />);
    await user.click(screen.getByRole("button", { name: /Cancel/i }));
    expect(onClose).toHaveBeenCalled();
  });
});
