import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
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

import { ProviderTable } from "./ProviderTable";

import type { IntegrationProvider } from "@/lib/services/integration.types";

const mockProviders: IntegrationProvider[] = [
  {
    provider_id: "p1",
    provider_key: "gitea-main",
    provider_type: "scm",
    display_name: "Gitea Main",
    enabled: true,
    auth_mode: "token",
    credentials_ref: "hmac_sha256:abc123",
    capabilities: ["pull", "push", "webhook"],
    sync_status: "ok",
    last_sync_at: "2026-05-28T10:00:00Z",
    last_error_code: null,
    base_url: "https://gitea.example.com",
    api_token_set: true,
    auth_username: null,
    auth_client_id: null,
    auth_token_url: null,
    auth_secret_set: false,
    created_at: "2026-05-01T00:00:00Z",
    updated_at: "2026-05-28T10:00:00Z",
  },
  {
    provider_id: "p2",
    provider_key: "jenkins-prod",
    provider_type: "ci_cd",
    display_name: "Jenkins Production",
    enabled: false,
    auth_mode: "basic",
    credentials_ref: "hmac_sha256:def456",
    capabilities: ["webhook"],
    sync_status: "error",
    last_sync_at: null,
    last_error_code: "ECONNREFUSED",
    base_url: null,
    api_token_set: false,
    auth_username: "svc-jenkins",
    auth_client_id: null,
    auth_token_url: null,
    auth_secret_set: true,
    created_at: "2026-05-10T00:00:00Z",
    updated_at: "2026-05-27T00:00:00Z",
  },
];

describe("ProviderTable", () => {
  const baseProps = {
    items: [] as IntegrationProvider[],
    onEdit: vi.fn(),
    onSync: vi.fn(),
    onDelete: vi.fn(),
    syncingProviderID: null as string | null,
    deletingProviderID: null as string | null,
  };

  it("renders empty state correctly", () => {
    render(<ProviderTable {...baseProps} />);
    expect(screen.getByText("등록된 integration provider 가 없습니다")).toBeInTheDocument();
  });

  describe("with providers", () => {
    const user = userEvent.setup();

    it("renders provider rows with display name, type badge, auth mode, and enabled badge", () => {
      render(<ProviderTable {...baseProps} items={mockProviders} />);

      expect(screen.getByText("Gitea Main")).toBeInTheDocument();
      expect(screen.getByText("gitea-main")).toBeInTheDocument();
      expect(screen.getByText("Jenkins Production")).toBeInTheDocument();
      expect(screen.getByText("jenkins-prod")).toBeInTheDocument();

      // Type badges
      expect(screen.getByText("scm")).toBeInTheDocument();
      expect(screen.getByText("ci_cd")).toBeInTheDocument();

      // Auth modes
      expect(screen.getByText("token")).toBeInTheDocument();
      expect(screen.getByText("basic")).toBeInTheDocument();

      // Sync status
      expect(screen.getByText("OK")).toBeInTheDocument();
      expect(screen.getByText("Error")).toBeInTheDocument();

      // Error code
      expect(screen.getByText("ECONNREFUSED")).toBeInTheDocument();
    });

    it("calls onSync when Sync button is clicked", async () => {
      const onSync = vi.fn();
      render(<ProviderTable {...baseProps} items={mockProviders} onSync={onSync} />);

      const syncBtn = screen.getAllByRole("button", { name: /Sync/i });
      await user.click(syncBtn[0]);

      expect(onSync).toHaveBeenCalledWith(mockProviders[0]);
    });

    it("calls onEdit when Edit button is clicked", async () => {
      const onEdit = vi.fn();
      render(<ProviderTable {...baseProps} items={mockProviders} onEdit={onEdit} />);

      const editBtns = screen.getAllByRole("button", { name: /Edit/i });
      await user.click(editBtns[0]);

      expect(onEdit).toHaveBeenCalledWith(mockProviders[0]);
    });

    it("calls onDelete when Delete button is clicked", async () => {
      const onDelete = vi.fn();
      render(<ProviderTable {...baseProps} items={mockProviders} onDelete={onDelete} />);

      const deleteBtns = screen.getAllByRole("button", { name: /Delete/i });
      await user.click(deleteBtns[0]);

      expect(onDelete).toHaveBeenCalledWith(mockProviders[0]);
    });

    it("shows syncing state and disables Sync button", () => {
      render(<ProviderTable {...baseProps} items={mockProviders} syncingProviderID="p1" />);

      const syncingBtns = screen.getAllByRole("button", { name: /^Sync .+/i });
      // In syncing state, the first Sync button should be disabled (spinner)
      expect(syncingBtns[0]).toBeDisabled();
    });

    it("shows deleting state and disables Delete button", () => {
      render(<ProviderTable {...baseProps} items={mockProviders} deletingProviderID="p1" />);

      const deleteBtns = screen.getAllByRole("button", { name: /Delete/i });
      expect(deleteBtns[0]).toBeDisabled();
    });

    it("shows Import button for scm + pull capability when onImport is provided", () => {
      const onImport = vi.fn();
      render(<ProviderTable {...baseProps} items={mockProviders} onImport={onImport} />);

      // p1 is scm with pull capability
      expect(screen.getByRole("button", { name: /Import/i })).toBeInTheDocument();
    });

    it("shows New Repo button for scm + push capability when onCreateRepo is provided", () => {
      const onCreateRepo = vi.fn();
      render(<ProviderTable {...baseProps} items={mockProviders} onCreateRepo={onCreateRepo} />);

      // p1 is scm with push capability
      const newRepoBtns = screen.getAllByText("New Repo");
      expect(newRepoBtns.length).toBe(1);
    });

    it("does not show Import/NewRepo buttons when callbacks are omitted", () => {
      render(<ProviderTable {...baseProps} items={mockProviders} />);

      expect(screen.queryByRole("button", { name: /Import/i })).not.toBeInTheDocument();
      expect(screen.queryByRole("button", { name: /New Repo/i })).not.toBeInTheDocument();
    });

    it("hides Import button for non-scm providers", () => {
      const onImport = vi.fn();
      // Only non-scm providers
      const nonScmItems = [mockProviders[1]]; // jenkins-prod = ci_cd
      render(<ProviderTable {...baseProps} items={nonScmItems} onImport={onImport} />);

      expect(screen.queryByRole("button", { name: /Import/i })).not.toBeInTheDocument();
    });
  });
});
