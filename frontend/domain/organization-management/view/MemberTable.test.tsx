import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";

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

const deleteUser = vi.fn();
vi.mock("@/domain/organization-management/service/identity.service", async () => {
  const actual = await vi.importActual<typeof import("@/domain/organization-management/service/identity.service")>(
    "@/domain/organization-management/service/identity.service",
  );
  return {
    ...actual,
    identityService: {
      ...actual.identityService,
      deleteUser: (...args: unknown[]) => deleteUser(...args),
    },
  };
});

const toastFn = vi.fn();
vi.mock("@/shared/ui-foundation/components/Toast", () => ({
  useToast: () => ({ toast: toastFn }),
}));

// UserEditModal — simple noise mock
vi.mock("@/components/organization/UserEditModal", () => ({
  UserEditModal: () => null,
}));

// UserCreationModal — noise mock (we are not testing the inner modal)
vi.mock("./UserCreationModal", () => ({
  UserCreationModal: ({ onClose }: { onClose: () => void }) => {
    const React = require("react");
    return React.createElement(
      "div",
      { "data-testid": "user-creation-modal" },
      React.createElement("button", { onClick: onClose, type: "button" }, "mock-close"),
    );
  },
}));

import { MemberTable } from "./MemberTable";
import type { OrgMember } from "@/domain/organization-management/service/identity.service";
import type { Role } from "@/lib/services/rbac.types";

const roles: Role[] = [
  { id: 1, name: "Developer", description: "" },
  { id: 2, name: "Manager", description: "" },
  { id: 3, name: "System Admin", description: "" },
];

const members: OrgMember[] = [
  {
    id: "u1",
    name: "Alice",
    email: "alice@example.com",
    primary_dept_id: "d1",
    current_dept_id: "d1",
    is_seconded: false,
    role: "Developer",
    status: "active",
    appointments: [{ dept_id: "d1", role: "leader" }],
    joined_at: "",
  },
  {
    id: "u2",
    name: "Bob",
    email: "bob@example.com",
    primary_dept_id: "d1",
    current_dept_id: "d2",
    is_seconded: true,
    role: "Manager",
    status: "pending",
    appointments: [{ dept_id: "d1", role: "leader" }, { dept_id: "d2", role: "leader" }],
    joined_at: "",
    type: "human",
  },
  {
    id: "bot-1",
    name: "Bot Garden",
    email: "bot@example.com",
    primary_dept_id: "d1",
    current_dept_id: "d1",
    is_seconded: false,
    role: "Developer",
    status: "deactivated",
    appointments: [],
    joined_at: "",
    type: "system",
  },
];

beforeEach(() => {
  deleteUser.mockReset();
  toastFn.mockClear();
});

describe("MemberTable", () => {
  it("renders member rows + email + status badges", () => {
    render(
      <MemberTable
        members={members}
        roles={roles}
        unitNames={{ d1: "Engineering", d2: "Ops" }}
        onUpdateMemberRole={vi.fn()}
      />,
    );
    expect(screen.getByText("Alice")).toBeInTheDocument();
    expect(screen.getByText("Bob")).toBeInTheDocument();
    expect(screen.getByText("Bot Garden")).toBeInTheDocument();
    expect(screen.getByText("alice@example.com")).toBeInTheDocument();
    expect(screen.getByText("active")).toBeInTheDocument();
    expect(screen.getByText("pending")).toBeInTheDocument();
    expect(screen.getByText("deactivated")).toBeInTheDocument();
  });

  it("renders Dual badge for member with multiple leader appointments", () => {
    render(
      <MemberTable
        members={members}
        roles={roles}
        unitNames={{ d1: "Engineering", d2: "Ops" }}
        onUpdateMemberRole={vi.fn()}
      />,
    );
    expect(screen.getByText("Dual")).toBeInTheDocument();
  });

  it("renders System / AI badge for system type member", () => {
    render(
      <MemberTable
        members={members}
        roles={roles}
        unitNames={{ d1: "Engineering" }}
        onUpdateMemberRole={vi.fn()}
      />,
    );
    expect(screen.getByText(/System \/ AI/)).toBeInTheDocument();
  });

  it("renders Seconded badge for is_seconded member", () => {
    render(
      <MemberTable
        members={members}
        roles={roles}
        unitNames={{ d1: "Engineering", d2: "Ops" }}
        onUpdateMemberRole={vi.fn()}
      />,
    );
    expect(screen.getByText("Seconded")).toBeInTheDocument();
    expect(screen.getByText(/Original: Engineering/)).toBeInTheDocument();
  });

  it("calls onUpdateMemberRole when role select changes", () => {
    const onUpdateMemberRole = vi.fn();
    render(
      <MemberTable
        members={members}
        roles={roles}
        unitNames={{ d1: "Engineering" }}
        onUpdateMemberRole={onUpdateMemberRole}
      />,
    );
    const selects = screen.getAllByRole("combobox");
    fireEvent.change(selects[0], { target: { value: "Manager" } });
    expect(onUpdateMemberRole).toHaveBeenCalledWith("u1", "Manager");
  });

  it("opens UserCreationModal when Invite Member button is clicked", () => {
    render(
      <MemberTable
        members={members}
        roles={roles}
        unitNames={{ d1: "Engineering" }}
        onUpdateMemberRole={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Invite Member/i }));
    expect(screen.getByTestId("user-creation-modal")).toBeInTheDocument();
  });

  it("opens DestructiveConfirmModal when delete button is clicked", () => {
    render(
      <MemberTable
        members={members}
        roles={roles}
        unitNames={{ d1: "Engineering" }}
        onUpdateMemberRole={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByLabelText("Delete Alice"));
    expect(screen.getByText(/되돌릴 수 없습니다/)).toBeInTheDocument();
  });

  it("calls deleteUser and onMemberDeleted on confirm delete", async () => {
    deleteUser.mockResolvedValueOnce(undefined);
    const onMemberDeleted = vi.fn();
    render(
      <MemberTable
        members={members}
        roles={roles}
        unitNames={{ d1: "Engineering" }}
        onUpdateMemberRole={vi.fn()}
        onMemberDeleted={onMemberDeleted}
      />,
    );
    fireEvent.click(screen.getByLabelText("Delete Alice"));
    const confirmBtn = screen.getByRole("button", { name: /Delete Member/i });
    fireEvent.click(confirmBtn);

    await waitFor(() => {
      expect(deleteUser).toHaveBeenCalledWith("u1");
    });
    expect(onMemberDeleted).toHaveBeenCalledWith("u1");
  });

  it("toasts error when deleteUser rejects", async () => {
    deleteUser.mockRejectedValueOnce(new Error("api"));
    const errSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    render(
      <MemberTable
        members={members}
        roles={roles}
        unitNames={{ d1: "Engineering" }}
        onUpdateMemberRole={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByLabelText("Delete Alice"));
    fireEvent.click(screen.getByRole("button", { name: /Delete Member/i }));
    await waitFor(() => {
      expect(toastFn).toHaveBeenCalledWith("Failed to delete member", "error");
    });
    errSpy.mockRestore();
  });

  it("renders dept name from unitNames map or falls back to id", () => {
    render(
      <MemberTable
        members={members}
        roles={roles}
        unitNames={{ d1: "Engineering" }}
        onUpdateMemberRole={vi.fn()}
      />,
    );
    // d1 maps to Engineering — Alice's current_dept = d1 → 'Engineering'
    expect(screen.getAllByText("Engineering").length).toBeGreaterThan(0);
    // d2 has no mapping → falls back to "d2" id for Bob
    expect(screen.getByText("d2")).toBeInTheDocument();
  });
});
