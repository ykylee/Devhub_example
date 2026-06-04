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

const updateBinding = vi.fn();
vi.mock("@/domain/integration-registry/service/integration.service", () => ({
  integrationService: {
    updateBinding: (...args: unknown[]) => updateBinding(...args),
  },
}));

import { ApiError } from "@/shared/api/api-client";
import { EditBindingModal } from "./EditBindingModal";
import type { IntegrationBinding, IntegrationProvider } from "@/domain/integration-registry/schema/integration.types";

const providers: IntegrationProvider[] = [
  {
    provider_id: "prov-1",
    provider_key: "gitea-main",
    provider_type: "scm",
    display_name: "Gitea Main",
    enabled: true,
    auth_mode: "token",
    credentials_ref: "hmac",
    capabilities: ["pull", "push"],
    sync_status: "ok",
    last_sync_at: null,
    last_error_code: null,
    base_url: null,
    api_token_set: true,
    auth_username: null,
    auth_client_id: null,
    auth_token_url: null,
    auth_secret_set: false,
    created_at: "",
    updated_at: "",
  },
  {
    provider_id: "prov-2",
    provider_key: "jira-prod",
    provider_type: "alm",
    display_name: "Jira Prod",
    enabled: true,
    auth_mode: "basic",
    credentials_ref: "hmac",
    capabilities: ["pull"],
    sync_status: "ok",
    last_sync_at: null,
    last_error_code: null,
    base_url: null,
    api_token_set: false,
    auth_username: "u",
    auth_client_id: null,
    auth_token_url: null,
    auth_secret_set: true,
    created_at: "",
    updated_at: "",
  },
];

const binding: IntegrationBinding = {
  binding_id: "b1",
  scope_type: "platform",
  scope_id: "app-1",
  provider_id: "prov-1",
  external_key: "org/repo",
  policy: "summary_only",
  enabled: true,
  created_at: "",
  updated_at: "",
};

beforeEach(() => {
  updateBinding.mockReset();
});

describe("EditBindingModal", () => {
  it("prefills initial values from binding", () => {
    render(
      <EditBindingModal
        binding={binding}
        providers={providers}
        onClose={vi.fn()}
        onUpdated={vi.fn()}
      />,
    );
    const providerSelect = screen.getByLabelText(/Provider/) as HTMLSelectElement;
    expect(providerSelect.value).toBe("prov-1");
    const externalKey = screen.getByLabelText(/External Key/) as HTMLInputElement;
    expect(externalKey.value).toBe("org/repo");
    const policy = screen.getByLabelText(/Policy/) as HTMLSelectElement;
    expect(policy.value).toBe("summary_only");
    expect(screen.getByText("Enabled")).toBeInTheDocument();
    // Scope is shown as read-only
    expect(screen.getByText("app-1")).toBeInTheDocument();
  });

  it("toggles enabled/disabled status button", async () => {
    const user = userEvent.setup();
    render(
      <EditBindingModal
        binding={binding}
        providers={providers}
        onClose={vi.fn()}
        onUpdated={vi.fn()}
      />,
    );
    expect(screen.getByText("Enabled")).toBeInTheDocument();
    await user.click(screen.getByText("Enabled"));
    expect(screen.getByText("Disabled")).toBeInTheDocument();
  });

  it("submits with updated payload and calls onUpdated + onClose", async () => {
    const user = userEvent.setup();
    const onUpdated = vi.fn();
    const onClose = vi.fn();
    const updated: IntegrationBinding = { ...binding, policy: "execution_system" };
    updateBinding.mockResolvedValueOnce(updated);

    render(
      <EditBindingModal
        binding={binding}
        providers={providers}
        onClose={onClose}
        onUpdated={onUpdated}
      />,
    );
    const policy = screen.getByLabelText(/Policy/) as HTMLSelectElement;
    fireEvent.change(policy, { target: { value: "execution_system" } });
    await user.click(screen.getByRole("button", { name: /Save Changes/i }));

    await waitFor(() => {
      expect(updateBinding).toHaveBeenCalledWith("b1", {
        provider_id: "prov-1",
        external_key: "org/repo",
        policy: "execution_system",
        enabled: true,
      });
    });
    expect(onUpdated).toHaveBeenCalledWith(updated);
    expect(onClose).toHaveBeenCalled();
  });

  it("shows validation error when external_key is empty", async () => {
    const { container } = render(
      <EditBindingModal
        binding={binding}
        providers={providers}
        onClose={vi.fn()}
        onUpdated={vi.fn()}
      />,
    );
    const externalKey = screen.getByLabelText(/External Key/) as HTMLInputElement;
    fireEvent.change(externalKey, { target: { value: "   " } });
    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);
    await waitFor(() => {
      expect(screen.getByText(/external_key 는 필수입니다/)).toBeInTheDocument();
    });
  });

  it("shows 409 conflict error in Korean message", async () => {
    const user = userEvent.setup();
    const apiErr = new ApiError(409, { code: "integration_binding_conflict" }, "conflict");
    updateBinding.mockRejectedValueOnce(apiErr);
    render(
      <EditBindingModal
        binding={binding}
        providers={providers}
        onClose={vi.fn()}
        onUpdated={vi.fn()}
      />,
    );
    await user.click(screen.getByRole("button", { name: /Save Changes/i }));
    await waitFor(() => {
      expect(screen.getByText(/이미 동일한/)).toBeInTheDocument();
    });
  });

  it("shows generic error message for non-API errors", async () => {
    const user = userEvent.setup();
    updateBinding.mockRejectedValueOnce(new Error("network"));
    render(
      <EditBindingModal
        binding={binding}
        providers={providers}
        onClose={vi.fn()}
        onUpdated={vi.fn()}
      />,
    );
    await user.click(screen.getByRole("button", { name: /Save Changes/i }));
    await waitFor(() => {
      expect(screen.getByText("network")).toBeInTheDocument();
    });
  });

  it("calls onClose when Cancel is clicked", async () => {
    const onClose = vi.fn();
    render(
      <EditBindingModal
        binding={binding}
        providers={providers}
        onClose={onClose}
        onUpdated={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Cancel/i }));
    expect(onClose).toHaveBeenCalled();
  });

  it("calls onClose when X close button is clicked", async () => {
    const onClose = vi.fn();
    render(
      <EditBindingModal
        binding={binding}
        providers={providers}
        onClose={onClose}
        onUpdated={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Close" }));
    expect(onClose).toHaveBeenCalled();
  });
});
