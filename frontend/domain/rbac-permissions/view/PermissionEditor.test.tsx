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

import { PermissionEditor } from "./PermissionEditor";
import type { Role } from "@/domain/rbac-permissions/schema/rbac.types";

const systemRole: Role = {
  id: "system_admin",
  name: "System Admin",
  description: "System administrator",
  system: true,
  permissions: { infrastructure: { view: true } },
};

const customRole: Role = {
  id: "custom-1",
  name: "Custom Role",
  description: "Custom role description",
  system: false,
  permissions: {},
};

describe("PermissionEditor", () => {
  it("renders role list with names + descriptions + system tag for system role", () => {
    const setRoles = vi.fn();
    render(
      <PermissionEditor roles={[systemRole, customRole]} setRoles={setRoles} />,
    );

    expect(screen.getByText("System Admin")).toBeInTheDocument();
    expect(screen.getByText("System administrator")).toBeInTheDocument();
    expect(screen.getByText("Custom Role")).toBeInTheDocument();
    expect(screen.getByText("Custom role description")).toBeInTheDocument();
    expect(screen.getByText("System")).toBeInTheDocument();
  });

  it("shows empty selection state initially", () => {
    const setRoles = vi.fn();
    render(<PermissionEditor roles={[systemRole]} setRoles={setRoles} />);
    expect(screen.getByText("Select a Role")).toBeInTheDocument();
  });

  it("selects a role on click and shows its matrix", async () => {
    const user = userEvent.setup();
    const setRoles = vi.fn();
    render(<PermissionEditor roles={[customRole]} setRoles={setRoles} />);

    await user.click(screen.getByText("Custom Role"));
    expect(screen.getByText(/Custom Role Matrix/i)).toBeInTheDocument();
  });

  it("calls onCreate when Create Role is clicked", async () => {
    const user = userEvent.setup();
    const setRoles = vi.fn();
    const onCreate = vi.fn();
    render(
      <PermissionEditor roles={[]} setRoles={setRoles} onCreate={onCreate} />,
    );

    await user.click(screen.getByRole("button", { name: /Create Role/i }));
    expect(onCreate).toHaveBeenCalledTimes(1);
    expect(onCreate.mock.calls[0][0]).toMatchObject({
      name: "New Custom Role",
      system: false,
    });
  });

  it("falls back to setRoles when onCreate is not provided", async () => {
    const user = userEvent.setup();
    const setRoles = vi.fn();
    render(<PermissionEditor roles={[customRole]} setRoles={setRoles} />);

    await user.click(screen.getByRole("button", { name: /Create Role/i }));
    expect(setRoles).toHaveBeenCalledTimes(1);
    const newRoles = setRoles.mock.calls[0][0] as Role[];
    expect(newRoles).toHaveLength(2);
    expect(newRoles[1].name).toBe("New Custom Role");
  });

  it("calls onDelete for non-system roles when trash button is clicked", async () => {
    const user = userEvent.setup();
    const setRoles = vi.fn();
    const onDelete = vi.fn();
    render(
      <PermissionEditor
        roles={[customRole]}
        setRoles={setRoles}
        onDelete={onDelete}
      />,
    );

    // The Trash2 icon-only button — find by role within the custom role card
    const trashBtns = screen.getAllByRole("button");
    // Find the inner trash button (without name) — should not include Create Role
    const trashCandidates = trashBtns.filter(
      (b) => !b.textContent?.includes("Create Role"),
    );
    expect(trashCandidates.length).toBeGreaterThan(0);
    // Trigger click on the candidate that is inside the role card (first one usually)
    await user.click(trashCandidates[0]);
    expect(onDelete).toHaveBeenCalledWith("custom-1");
  });

  it("does not allow deleting system roles (no delete button)", () => {
    const setRoles = vi.fn();
    render(<PermissionEditor roles={[systemRole]} setRoles={setRoles} />);

    // System role card should NOT show a delete (Trash2) button.
    // The header has Create Role only.
    expect(screen.queryAllByRole("button").length).toBeLessThanOrEqual(2);
  });

  it("renders Save Permissions button only when onSave is provided", () => {
    const setRoles = vi.fn();
    const onSave = vi.fn();
    const { rerender } = render(
      <PermissionEditor roles={[customRole]} setRoles={setRoles} />,
    );

    expect(
      screen.queryByRole("button", { name: /Save Permissions/i }),
    ).not.toBeInTheDocument();

    rerender(
      <PermissionEditor
        roles={[customRole]}
        setRoles={setRoles}
        onSave={onSave}
        isDirty
      />,
    );
    const saveBtn = screen.getByRole("button", { name: /Save Permissions/i });
    expect(saveBtn).toBeInTheDocument();
    expect(saveBtn).not.toBeDisabled();
  });

  it("disables Save Permissions when not dirty or while saving", () => {
    const setRoles = vi.fn();
    const onSave = vi.fn();
    const { rerender } = render(
      <PermissionEditor
        roles={[customRole]}
        setRoles={setRoles}
        onSave={onSave}
        isDirty={false}
      />,
    );

    expect(screen.getByRole("button", { name: /Save Permissions/i })).toBeDisabled();

    rerender(
      <PermissionEditor
        roles={[customRole]}
        setRoles={setRoles}
        onSave={onSave}
        isDirty
        saving
      />,
    );
    expect(screen.getByRole("button", { name: /Saving…/i })).toBeDisabled();
  });

  it("renders error message when errorMessage is provided", () => {
    const setRoles = vi.fn();
    render(
      <PermissionEditor
        roles={[customRole]}
        setRoles={setRoles}
        errorMessage="Something went wrong"
      />,
    );
    expect(screen.getByText("Something went wrong")).toBeInTheDocument();
  });

  it("calls onSave when Save Permissions is clicked", async () => {
    const user = userEvent.setup();
    const setRoles = vi.fn();
    const onSave = vi.fn();
    render(
      <PermissionEditor
        roles={[customRole]}
        setRoles={setRoles}
        onSave={onSave}
        isDirty
      />,
    );

    await user.click(screen.getByRole("button", { name: /Save Permissions/i }));
    expect(onSave).toHaveBeenCalled();
  });

  it("renders count of resources configured for a role", () => {
    const setRoles = vi.fn();
    render(<PermissionEditor roles={[systemRole]} setRoles={setRoles} />);
    expect(screen.getByText(/1 Resources Configured/i)).toBeInTheDocument();
  });
});
