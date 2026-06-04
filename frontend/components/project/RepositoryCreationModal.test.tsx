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
  createRepositoryDraft: vi.fn(),
  listProviders: vi.fn(),
}));
vi.mock("@/domain/repository-integration/service/repository.service", () => ({
  repositoryService: {
    createRepositoryDraft: (...a: unknown[]) => mocks.createRepositoryDraft(...a),
  },
}));
vi.mock("@/domain/integration-registry/service/integration.service", () => ({
  integrationService: {
    listProviders: (...a: unknown[]) => mocks.listProviders(...a),
  },
}));

import { RepositoryCreationModal } from "./RepositoryCreationModal";
import type { Repository } from "@/domain/repository-integration/service/repository.service";

const sampleProviders = [
  { provider_id: "p1", provider_key: "gitea", provider_type: "scm", display_name: "Local Gitea", enabled: true, auth_mode: "token", credentials_ref: "", capabilities: [], sync_status: "active", last_sync_at: null, last_error_code: null, base_url: null, api_token_set: false, auth_username: null, auth_client_id: null, auth_token_url: null, auth_secret_set: false, created_at: "", updated_at: "" },
  { provider_id: "p2", provider_key: "github", provider_type: "scm", display_name: "GitHub", enabled: true, auth_mode: "token", credentials_ref: "", capabilities: [], sync_status: "active", last_sync_at: null, last_error_code: null, base_url: null, api_token_set: false, auth_username: null, auth_client_id: null, auth_token_url: null, auth_secret_set: false, created_at: "", updated_at: "" },
];

const createdRepo = {
  id: 1, full_name: "devhub/new-repo", name: "NEW-REPO", owner_login: "devhub",
  owner_login_field: "devhub", owner_login_unused: "devhub",
  clone_url: "", html_url: "", default_branch: "main", private: false, status: "draft",
  provider_id: "p1", provider_key: "gitea", updated_at: "2026-06-04T00:00:00Z",
  linked_applications_count: 0, linked_projects_count: 0,
} as unknown as Repository;

beforeEach(() => {
  mocks.createRepositoryDraft.mockReset();
  mocks.listProviders.mockReset();
  mocks.listProviders.mockResolvedValue({ data: sampleProviders, total: 2 });
});

describe("RepositoryCreationModal", () => {
  it("loads SCM providers on mount", async () => {
    render(
      <RepositoryCreationModal onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(mocks.listProviders).toHaveBeenCalledWith({ provider_type: "scm", enabled: true });
    });
  });

  it("submits POST with key, slug and provider_key (gitea selected)", async () => {
    const onCreated = vi.fn();
    const onClose = vi.fn();
    mocks.createRepositoryDraft.mockResolvedValueOnce(createdRepo);
    render(
      <RepositoryCreationModal onClose={onClose} onCreated={onCreated} />,
    );
    await waitFor(() => {
      expect(mocks.listProviders).toHaveBeenCalled();
    });

    fireEvent.change(screen.getByLabelText(/SCM Provider Key/), {
      target: { value: "gitea" },
    });
    fireEvent.change(screen.getByPlaceholderText("E.G. DEVHUBAPI"), {
      target: { value: "new-repo" },
    });
    fireEvent.change(screen.getByPlaceholderText("e.g. devhub/devhub-api"), {
      target: { value: "devhub/new-repo" },
    });
    fireEvent.click(screen.getByRole("button", { name: /create repository/i }));

    await waitFor(() => {
      expect(mocks.createRepositoryDraft).toHaveBeenCalledWith({
        key: "NEW-REPO",
        slug: "devhub/new-repo",
        provider_key: "gitea",
      });
    });
    expect(onCreated).toHaveBeenCalledWith(createdRepo);
    expect(onClose).toHaveBeenCalled();
  });

  it("submits POST with empty provider_key when 'No SCM link' selected (default)", async () => {
    const onCreated = vi.fn();
    const onClose = vi.fn();
    mocks.createRepositoryDraft.mockResolvedValueOnce(createdRepo);
    render(
      <RepositoryCreationModal onClose={onClose} onCreated={onCreated} />,
    );
    await waitFor(() => {
      expect(mocks.listProviders).toHaveBeenCalled();
    });

    fireEvent.change(screen.getByPlaceholderText("E.G. DEVHUBAPI"), {
      target: { value: "new-repo" },
    });
    fireEvent.change(screen.getByPlaceholderText("e.g. devhub/devhub-api"), {
      target: { value: "devhub/new-repo" },
    });
    fireEvent.click(screen.getByRole("button", { name: /create repository/i }));

    await waitFor(() => {
      expect(mocks.createRepositoryDraft).toHaveBeenCalledWith({
        key: "NEW-REPO",
        slug: "devhub/new-repo",
        provider_key: undefined,
      });
    });
  });

  it("shows 'No enabled SCM providers' when list returns empty", async () => {
    mocks.listProviders.mockReset();
    mocks.listProviders.mockResolvedValue({ data: [], total: 0 });
    render(
      <RepositoryCreationModal onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(screen.getByText(/No enabled SCM providers/)).toBeInTheDocument();
    });
  });

  it("shows error message on submit failure", async () => {
    mocks.createRepositoryDraft.mockRejectedValueOnce(new Error("409 conflict"));
    render(
      <RepositoryCreationModal onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(mocks.listProviders).toHaveBeenCalled();
    });

    fireEvent.change(screen.getByPlaceholderText("E.G. DEVHUBAPI"), {
      target: { value: "new-repo" },
    });
    fireEvent.change(screen.getByPlaceholderText("e.g. devhub/devhub-api"), {
      target: { value: "devhub/new-repo" },
    });
    fireEvent.click(screen.getByRole("button", { name: /create repository/i }));

    await waitFor(() => {
      expect(screen.getByText("409 conflict")).toBeInTheDocument();
    });
  });

  it("calls onClose when ESC is pressed", () => {
    const onClose = vi.fn();
    render(
      <RepositoryCreationModal onClose={onClose} onCreated={vi.fn()} />,
    );
    fireEvent.keyDown(window, { key: "Escape" });
    expect(onClose).toHaveBeenCalled();
  });
});
