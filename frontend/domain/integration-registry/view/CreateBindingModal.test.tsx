import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";

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

const createBinding = vi.fn();
vi.mock("@/domain/integration-registry/service/integration.service", () => ({
  integrationService: {
    createBinding: (...args: unknown[]) => createBinding(...args),
  },
}));

const getApplications = vi.fn();
vi.mock("@/domain/application-lifecycle/service/project.service", () => ({
  projectService: {
    getApplications: (...args: unknown[]) => getApplications(...args),
  },
}));

import { ApiError } from "@/shared/api/api-client";
import { CreateBindingModal } from "./CreateBindingModal";
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
];

function getForm(container: HTMLElement): HTMLFormElement {
  const form = container.querySelector("form");
  if (!form) throw new Error("form not found");
  return form;
}

beforeEach(() => {
  createBinding.mockReset();
  getApplications.mockReset();
  getApplications.mockResolvedValue([
    { id: "app-1", key: "APP-1", name: "App One" },
    { id: "app-2", key: "APP-2", name: "App Two" },
  ]);
});

describe("CreateBindingModal", () => {
  it("renders scope_type select with application as default", async () => {
    render(
      <CreateBindingModal providers={providers} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    const select = screen.getByLabelText(/Scope Type/) as HTMLSelectElement;
    expect(select.value).toBe("application");
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
  });

  it("renders project plain input when scope_type is project", async () => {
    render(
      <CreateBindingModal providers={providers} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
    const select = screen.getByLabelText(/Scope Type/) as HTMLSelectElement;
    fireEvent.change(select, { target: { value: "project" } });
    await waitFor(() => {
      expect(screen.getByPlaceholderText("PROJ-001")).toBeInTheDocument();
    });
  });

  it("shows scope_id validation error when empty (project plain input)", async () => {
    const { container } = render(
      <CreateBindingModal providers={providers} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    const select = screen.getByLabelText(/Scope Type/) as HTMLSelectElement;
    fireEvent.change(select, { target: { value: "project" } });
    const externalKey = screen.getByLabelText(/External Key/) as HTMLInputElement;
    fireEvent.change(externalKey, { target: { value: "key1" } });
    fireEvent.submit(getForm(container));
    await waitFor(() => {
      expect(screen.getByText(/scope_id 는 필수입니다/)).toBeInTheDocument();
    });
  });

  it("shows external_key validation error when empty", async () => {
    const { container } = render(
      <CreateBindingModal providers={providers} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    const select = screen.getByLabelText(/Scope Type/) as HTMLSelectElement;
    fireEvent.change(select, { target: { value: "project" } });
    const scopeID = screen.getByPlaceholderText("PROJ-001") as HTMLInputElement;
    fireEvent.change(scopeID, { target: { value: "P1" } });
    fireEvent.submit(getForm(container));
    await waitFor(() => {
      expect(screen.getByText(/external_key 는 필수입니다/)).toBeInTheDocument();
    });
  });

  it("submits successfully with project scope and calls onCreated + onClose", async () => {
    const onCreated = vi.fn();
    const onClose = vi.fn();
    const created: IntegrationBinding = {
      binding_id: "b-new",
      scope_type: "project",
      scope_id: "P1",
      provider_id: "prov-1",
      external_key: "ext-key",
      policy: "execution_system",
      enabled: true,
      created_at: "",
      updated_at: "",
    };
    createBinding.mockResolvedValueOnce(created);

    const { container } = render(
      <CreateBindingModal providers={providers} onClose={onClose} onCreated={onCreated} />,
    );
    const select = screen.getByLabelText(/Scope Type/) as HTMLSelectElement;
    fireEvent.change(select, { target: { value: "project" } });
    fireEvent.change(screen.getByPlaceholderText("PROJ-001"), {
      target: { value: "P1" },
    });
    fireEvent.change(screen.getByLabelText(/External Key/), {
      target: { value: "ext-key" },
    });
    fireEvent.submit(getForm(container));

    await waitFor(() => {
      expect(createBinding).toHaveBeenCalledWith({
        scope_type: "project",
        scope_id: "P1",
        provider_id: "prov-1",
        external_key: "ext-key",
        policy: "execution_system",
      });
    });
    expect(onCreated).toHaveBeenCalledWith(created);
    expect(onClose).toHaveBeenCalled();
  });

  it("shows 409 binding_conflict error in Korean", async () => {
    const apiErr = new ApiError(
      409,
      { code: "integration_binding_conflict" },
      "conflict",
    );
    createBinding.mockRejectedValueOnce(apiErr);
    const { container } = render(
      <CreateBindingModal providers={providers} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    const select = screen.getByLabelText(/Scope Type/) as HTMLSelectElement;
    fireEvent.change(select, { target: { value: "project" } });
    fireEvent.change(screen.getByPlaceholderText("PROJ-001"), {
      target: { value: "P1" },
    });
    fireEvent.change(screen.getByLabelText(/External Key/), {
      target: { value: "ext-key" },
    });
    fireEvent.submit(getForm(container));
    await waitFor(() => {
      expect(screen.getByText(/이미 동일한/)).toBeInTheDocument();
    });
  });

  it("shows 422 policy_violation error in Korean", async () => {
    const apiErr = new ApiError(
      422,
      { code: "integration_policy_violation" },
      "policy",
    );
    createBinding.mockRejectedValueOnce(apiErr);
    const { container } = render(
      <CreateBindingModal providers={providers} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    const select = screen.getByLabelText(/Scope Type/) as HTMLSelectElement;
    fireEvent.change(select, { target: { value: "project" } });
    fireEvent.change(screen.getByPlaceholderText("PROJ-001"), {
      target: { value: "P1" },
    });
    fireEvent.change(screen.getByLabelText(/External Key/), {
      target: { value: "ext-key" },
    });
    fireEvent.submit(getForm(container));
    await waitFor(() => {
      expect(screen.getByText(/지원되지 않는 policy/)).toBeInTheDocument();
    });
  });

  it("shows fallback error message for non-API errors", async () => {
    createBinding.mockRejectedValueOnce(new Error("net down"));
    const { container } = render(
      <CreateBindingModal providers={providers} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    const select = screen.getByLabelText(/Scope Type/) as HTMLSelectElement;
    fireEvent.change(select, { target: { value: "project" } });
    fireEvent.change(screen.getByPlaceholderText("PROJ-001"), {
      target: { value: "P1" },
    });
    fireEvent.change(screen.getByLabelText(/External Key/), {
      target: { value: "k" },
    });
    fireEvent.submit(getForm(container));
    await waitFor(() => {
      expect(screen.getByText("net down")).toBeInTheDocument();
    });
  });

  it("calls onClose when Cancel is clicked", async () => {
    const onClose = vi.fn();
    render(
      <CreateBindingModal providers={providers} onClose={onClose} onCreated={vi.fn()} />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Cancel/i }));
    expect(onClose).toHaveBeenCalled();
  });

  it("disables Create button when providers list is empty", () => {
    render(
      <CreateBindingModal providers={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    const btn = screen.getByRole("button", { name: /Create/i });
    expect(btn).toBeDisabled();
    expect(screen.getByText(/등록된 provider 가 없습니다/)).toBeInTheDocument();
  });

  it("logs error and continues when getApplications fails", async () => {
    const errSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    getApplications.mockRejectedValueOnce(new Error("load fail"));
    render(
      <CreateBindingModal providers={providers} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(errSpy).toHaveBeenCalled();
    });
    errSpy.mockRestore();
  });
});
