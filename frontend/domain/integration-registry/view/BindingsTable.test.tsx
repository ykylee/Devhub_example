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

import { BindingsTable } from "./BindingsTable";

import type { IntegrationBinding, IntegrationProvider } from "@/domain/integration-registry/schema/integration.types";

const mockProviders: Record<string, IntegrationProvider> = {
  "prov-1": {
    provider_id: "prov-1",
    provider_key: "gitea-main",
    provider_type: "scm",
    display_name: "Gitea Main",
    enabled: true,
    auth_mode: "token",
    credentials_ref: "hmac_sha256:abc",
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
};

const mockBindings: IntegrationBinding[] = [
  {
    binding_id: "b1",
    scope_type: "platform",
    scope_id: "app-123",
    provider_id: "prov-1",
    external_key: "my-repo",
    policy: "summary_only",
    enabled: true,
    created_at: "2026-05-28T10:00:00Z",
    updated_at: "2026-05-28T10:00:00Z",
  },
  {
    binding_id: "b2",
    scope_type: "project",
    scope_id: "proj-456",
    provider_id: "prov-1",
    external_key: "docs",
    policy: "execution_system",
    enabled: false,
    created_at: "2026-05-27T00:00:00Z",
    updated_at: "2026-05-27T00:00:00Z",
  },
];

describe("BindingsTable", () => {
  const baseProps = {
    items: [] as IntegrationBinding[],
    providersByID: {} as Record<string, IntegrationProvider>,
    onEdit: vi.fn(),
    onDelete: vi.fn(),
  };

  it("renders empty state correctly", () => {
    render(<BindingsTable {...baseProps} />);
    expect(screen.getByText("등록된 binding 이 없습니다")).toBeInTheDocument();
  });

  describe("with bindings", () => {
    const user = userEvent.setup();
    const props = {
      items: mockBindings,
      providersByID: mockProviders,
      onEdit: vi.fn(),
      onDelete: vi.fn(),
    };

    it("renders binding rows with scope, provider, external key, policy, enabled", () => {
      render(<BindingsTable {...props} />);

      // Scope type badges
      expect(screen.getByText("platform")).toBeInTheDocument();
      expect(screen.getByText("project")).toBeInTheDocument();

      // Scope IDs
      expect(screen.getByText("app-123")).toBeInTheDocument();
      expect(screen.getByText("proj-456")).toBeInTheDocument();

      // Provider display name
      expect(screen.getAllByText("Gitea Main")).toHaveLength(2);
      expect(screen.getAllByText("gitea-main")).toHaveLength(2);

      // External keys
      expect(screen.getByText("my-repo")).toBeInTheDocument();
      expect(screen.getByText("docs")).toBeInTheDocument();

      // Policy badges
      expect(screen.getByText("summary_only")).toBeInTheDocument();
      expect(screen.getByText("execution_system")).toBeInTheDocument();

      // Enabled badges
      expect(screen.getAllByText("Yes").length).toBeGreaterThan(0);
      expect(screen.getByText("No")).toBeInTheDocument();
    });

    it("calls onEdit when ActionMenu Edit is clicked", async () => {
      const onEdit = vi.fn();
      render(<BindingsTable {...props} onEdit={onEdit} />);

      // ActionMenu button
      const actionBtns = screen.getAllByRole("button", { name: /Binding Actions/i });
      await user.click(actionBtns[0]);

      // Click Edit in the dropdown
      const editBtn = screen.getByText("Edit Binding");
      await user.click(editBtn);

      expect(onEdit).toHaveBeenCalledWith(mockBindings[0]);
    });

    it("calls onDelete when ActionMenu Delete is clicked", async () => {
      const onDelete = vi.fn();
      render(<BindingsTable {...props} onDelete={onDelete} />);

      const actionBtns = screen.getAllByRole("button", { name: /Binding Actions/i });
      await user.click(actionBtns[0]);

      const deleteBtn = screen.getByText("Delete Binding");
      await user.click(deleteBtn);

      expect(onDelete).toHaveBeenCalledWith(mockBindings[0]);
    });

    it("shows provider key fallback when provider display name is unavailable", () => {
      const unknownProviderBindings: IntegrationBinding[] = [
        {
          binding_id: "b3",
          scope_type: "platform",
          scope_id: "app-999",
          provider_id: "unknown-prov",
          external_key: "ext-key",
          policy: "summary_only",
          enabled: true,
          created_at: "2026-05-28T10:00:00Z",
          updated_at: "2026-05-28T10:00:00Z",
        },
      ];

      render(
        <BindingsTable
          items={unknownProviderBindings}
          providersByID={{}}
          onEdit={vi.fn()}
          onDelete={vi.fn()}
        />,
      );

      // Falls back to provider_id when not found in providersByID
      expect(screen.getByText("unknown-prov")).toBeInTheDocument();
    });

    it("renders raw created_at when value cannot be parsed (safeFormat catch)", () => {
      const malformed: IntegrationBinding[] = [
        {
          binding_id: "b4",
          scope_type: "platform",
          scope_id: "app-xyz",
          provider_id: "prov-1",
          external_key: "k",
          policy: "summary_only",
          enabled: true,
          created_at: "not-a-date",
          updated_at: "not-a-date",
        },
      ];
      render(
        <BindingsTable
          items={malformed}
          providersByID={mockProviders}
          onEdit={vi.fn()}
          onDelete={vi.fn()}
        />,
      );
      expect(screen.getByText("not-a-date")).toBeInTheDocument();
    });

    it("renders dash when created_at is missing (safeFormat null branch)", () => {
      const noDate: IntegrationBinding[] = [
        {
          binding_id: "b5",
          scope_type: "project",
          scope_id: "proj-x",
          provider_id: "prov-1",
          external_key: "k2",
          policy: "summary_only",
          enabled: true,
          created_at: "",
          updated_at: "",
        },
      ];
      render(
        <BindingsTable
          items={noDate}
          providersByID={mockProviders}
          onEdit={vi.fn()}
          onDelete={vi.fn()}
        />,
      );
      expect(screen.getByText("—")).toBeInTheDocument();
    });
  });
});
