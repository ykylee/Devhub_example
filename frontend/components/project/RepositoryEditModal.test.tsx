import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import * as React from "react";

vi.mock("framer-motion", () => {
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

const mocks = vi.hoisted(() => ({
  updateRepository: vi.fn(),
  listProviders: vi.fn(),
}));
vi.mock("@/domain/repository-integration/service/repository.service", () => ({
  repositoryService: {
    updateRepository: (...a: unknown[]) => mocks.updateRepository(...a),
  },
}));
vi.mock("@/domain/integration-registry/service/integration.service", () => ({
  integrationService: {
    listProviders: (...a: unknown[]) => mocks.listProviders(...a),
  },
}));

import { RepositoryEditModal } from "./RepositoryEditModal";
import type { Repository } from "@/domain/repository-integration/service/repository.service";

const sampleRepo: Repository = {
  id: 42,
  full_name: "devhub/sample-repo",
  name: "SAMPLE-REPO",
  owner_login: "devhub",
  owner_login_unused: "devhub",
  owner_login_field: "devhub",
  clone_url: "",
  html_url: "",
  default_branch: "main",
  private: false,
  status: "draft",
  provider_id: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
  provider_key: "gitea",
  updated_at: "2026-06-04T00:00:00Z",
  linked_applications_count: 0,
  linked_projects_count: 0,
} as unknown as Repository;

const sampleProviders = [
  { provider_id: "p1", provider_key: "gitea", provider_type: "scm", display_name: "Local Gitea", enabled: true, auth_mode: "token", credentials_ref: "", capabilities: [], sync_status: "active", last_sync_at: null, last_error_code: null, base_url: null, api_token_set: false, auth_username: null, auth_client_id: null, auth_token_url: null, auth_secret_set: false, created_at: "", updated_at: "" },
  { provider_id: "p2", provider_key: "github", provider_type: "scm", display_name: "GitHub", enabled: true, auth_mode: "token", credentials_ref: "", capabilities: [], sync_status: "active", last_sync_at: null, last_error_code: null, base_url: null, api_token_set: false, auth_username: null, auth_client_id: null, auth_token_url: null, auth_secret_set: false, created_at: "", updated_at: "" },
];

beforeEach(() => {
  mocks.updateRepository.mockReset();
  mocks.listProviders.mockReset();
  mocks.listProviders.mockResolvedValue({ data: sampleProviders, total: 2 });
});

describe("RepositoryEditModal", () => {
  it("pre-fills key, slug and provider_key from repository prop", async () => {
    render(
      <RepositoryEditModal repository={sampleRepo} onClose={vi.fn()} onUpdated={vi.fn()} />,
    );
    expect((screen.getByDisplayValue("SAMPLE-REPO") as HTMLInputElement).value).toBe("SAMPLE-REPO");
    expect((screen.getByDisplayValue("devhub/sample-repo") as HTMLInputElement).value).toBe("devhub/sample-repo");
    await waitFor(() => {
      expect((screen.getByLabelText(/SCM Provider/) as HTMLSelectElement).value).toBe("gitea");
    });
  });

  it("loads SCM providers from integrationService.listProviders on mount", async () => {
    render(
      <RepositoryEditModal repository={sampleRepo} onClose={vi.fn()} onUpdated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(mocks.listProviders).toHaveBeenCalledWith({ provider_type: "scm", enabled: true });
    });
  });

  it("shows 'No enabled SCM providers' when list returns empty", async () => {
    mocks.listProviders.mockReset();
    mocks.listProviders.mockResolvedValue({ data: [], total: 0 });
    render(
      <RepositoryEditModal repository={sampleRepo} onClose={vi.fn()} onUpdated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(screen.getByText(/No enabled SCM providers/)).toBeInTheDocument();
    });
  });

  it("shows loading skeleton while providers are loading", async () => {
    let resolveProviders!: (v: unknown) => void;
    mocks.listProviders.mockReset();
    mocks.listProviders.mockReturnValue(new Promise((r) => { resolveProviders = r; }));
    render(
      <RepositoryEditModal repository={sampleRepo} onClose={vi.fn()} onUpdated={vi.fn()} />,
    );
    expect(document.querySelector(".animate-pulse")).toBeInTheDocument();
    resolveProviders({ data: sampleProviders, total: 2 });
  });

  it("submits PATCH with only changed fields (key + slug)", async () => {
    const onUpdated = vi.fn();
    const onClose = vi.fn();
    mocks.updateRepository.mockResolvedValueOnce({ ...sampleRepo, name: "SAMPLE-REPO-2", full_name: "devhub/sample-repo-2" });
    render(
      <RepositoryEditModal repository={sampleRepo} onClose={onClose} onUpdated={onUpdated} />,
    );
    await waitFor(() => {
      expect((screen.getByLabelText(/SCM Provider/) as HTMLSelectElement).value).toBe("gitea");
    });

    fireEvent.change(screen.getByDisplayValue("SAMPLE-REPO"), {
      target: { value: "SAMPLE-REPO-2" },
    });
    fireEvent.change(screen.getByDisplayValue("devhub/sample-repo"), {
      target: { value: "devhub/sample-repo-2" },
    });
    // provider_key unchanged → should not be in PATCH body
    fireEvent.click(screen.getByRole("button", { name: /save changes/i }));

    await waitFor(() => {
      expect(mocks.updateRepository).toHaveBeenCalledWith(42, {
        key: "SAMPLE-REPO-2",
        slug: "devhub/sample-repo-2",
        // provider_key omitted (unchanged)
      });
    });
    expect(onUpdated).toHaveBeenCalled();
    expect(onClose).toHaveBeenCalled();
  });

  it("submits PATCH with provider_key when provider changes", async () => {
    mocks.updateRepository.mockResolvedValueOnce(sampleRepo);
    render(
      <RepositoryEditModal repository={sampleRepo} onClose={vi.fn()} onUpdated={vi.fn()} />,
    );
    await waitFor(() => {
      expect((screen.getByLabelText(/SCM Provider/) as HTMLSelectElement).value).toBe("gitea");
    });

    fireEvent.change(screen.getByLabelText(/SCM Provider/), {
      target: { value: "github" },
    });
    fireEvent.click(screen.getByRole("button", { name: /save changes/i }));

    await waitFor(() => {
      expect(mocks.updateRepository).toHaveBeenCalledWith(42, { provider_key: "github" });
    });
  });

  it("shows error message on submit failure", async () => {
    mocks.updateRepository.mockRejectedValueOnce(new Error("network down"));
    render(
      <RepositoryEditModal repository={sampleRepo} onClose={vi.fn()} onUpdated={vi.fn()} />,
    );
    await waitFor(() => {
      expect((screen.getByLabelText(/SCM Provider/) as HTMLSelectElement).value).toBe("gitea");
    });

    fireEvent.click(screen.getByRole("button", { name: /save changes/i }));

    await waitFor(() => {
      expect(screen.getByText("network down")).toBeInTheDocument();
    });
  });

  it("calls onClose when ESC is pressed", () => {
    const onClose = vi.fn();
    render(
      <RepositoryEditModal repository={sampleRepo} onClose={onClose} onUpdated={vi.fn()} />,
    );
    fireEvent.keyDown(window, { key: "Escape" });
    expect(onClose).toHaveBeenCalled();
  });
});
