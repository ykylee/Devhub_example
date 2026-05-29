import { describe, it, expect, vi } from "vitest";
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

import { MemberManagementModal } from "./MemberManagementModal";
import type { OrgMember } from "@/domain/organization-management/service/identity.service";

const allMembers: OrgMember[] = [
  {
    id: "u1",
    name: "Alice",
    email: "alice@example.com",
    primary_dept_id: "d1",
    current_dept_id: "d1",
    is_seconded: false,
    role: "Developer",
    status: "active",
    appointments: [],
    joined_at: "",
  },
  {
    id: "u2",
    name: "Bob",
    email: "bob@example.com",
    primary_dept_id: "d1",
    current_dept_id: "d1",
    is_seconded: false,
    role: "Manager",
    status: "active",
    appointments: [],
    joined_at: "",
  },
  {
    id: "u3",
    name: "Charlie",
    email: "charlie@example.com",
    primary_dept_id: "d1",
    current_dept_id: "d1",
    is_seconded: false,
    role: "Developer",
    status: "active",
    appointments: [],
    joined_at: "",
  },
];

describe("MemberManagementModal", () => {
  it("renders unit name + available/current panels", () => {
    render(
      <MemberManagementModal
        unitId="u-1"
        unitName="Engineering"
        allMembers={allMembers}
        currentMemberIds={["u1"]}
        currentLeaderId={null}
        onClose={vi.fn()}
        onSave={vi.fn()}
      />,
    );
    expect(screen.getByText("Engineering")).toBeInTheDocument();
    expect(screen.getByText("Available Personnel")).toBeInTheDocument();
    expect(screen.getByText("Unit Roster")).toBeInTheDocument();
  });

  it("renders current members in roster + available in panel", () => {
    render(
      <MemberManagementModal
        unitId="u-1"
        unitName="Engineering"
        allMembers={allMembers}
        currentMemberIds={["u1"]}
        currentLeaderId={null}
        onClose={vi.fn()}
        onSave={vi.fn()}
      />,
    );
    expect(screen.getByText("Alice")).toBeInTheDocument();
    expect(screen.getByText("Bob")).toBeInTheDocument();
    expect(screen.getByText("Charlie")).toBeInTheDocument();
  });

  it("filters available list via search input", () => {
    render(
      <MemberManagementModal
        unitId="u-1"
        unitName="Engineering"
        allMembers={allMembers}
        currentMemberIds={[]}
        currentLeaderId={null}
        onClose={vi.fn()}
        onSave={vi.fn()}
      />,
    );
    const search = screen.getByPlaceholderText("Search available...");
    fireEvent.change(search, { target: { value: "bob" } });
    expect(screen.getByText("Bob")).toBeInTheDocument();
    expect(screen.queryByText("Charlie")).not.toBeInTheDocument();
  });

  it("toggles a member from available -> roster on click", () => {
    render(
      <MemberManagementModal
        unitId="u-1"
        unitName="Engineering"
        allMembers={allMembers}
        currentMemberIds={[]}
        currentLeaderId={null}
        onClose={vi.fn()}
        onSave={vi.fn()}
      />,
    );
    // Click Bob in available panel — the available row has ChevronRight + onclick on outer
    const bobRow = screen.getByText("Bob").closest("div[class*='cursor-pointer']");
    if (!bobRow) throw new Error("bob row not found");
    fireEvent.click(bobRow);
    // After toggle, the roster count badge should be 1
    expect(screen.getByText("1")).toBeInTheDocument();
  });

  it("sets leader when 'Set lead' button is clicked", () => {
    render(
      <MemberManagementModal
        unitId="u-1"
        unitName="Engineering"
        allMembers={allMembers}
        currentMemberIds={["u1"]}
        currentLeaderId={null}
        onClose={vi.fn()}
        onSave={vi.fn()}
      />,
    );
    const setLeadBtn = screen.getByTitle("Set as leader");
    fireEvent.click(setLeadBtn);
    expect(screen.getAllByText("Leader").length).toBeGreaterThan(0);
  });

  it("resets leader when leader is removed from roster (toggleMember)", () => {
    render(
      <MemberManagementModal
        unitId="u-1"
        unitName="Engineering"
        allMembers={allMembers}
        currentMemberIds={["u1"]}
        currentLeaderId="u1"
        onClose={vi.fn()}
        onSave={vi.fn()}
      />,
    );
    // Remove Alice from roster
    const removeBtn = screen.getByTitle("Remove from unit");
    fireEvent.click(removeBtn);
    // leader resets to (none)
    expect(screen.getByText(/none/)).toBeInTheDocument();
  });

  it("calls onSave with selected ids + leader id when Save is clicked", async () => {
    const onSave = vi.fn().mockResolvedValue(undefined);
    render(
      <MemberManagementModal
        unitId="u-1"
        unitName="Engineering"
        allMembers={allMembers}
        currentMemberIds={["u1", "u2"]}
        currentLeaderId="u1"
        onClose={vi.fn()}
        onSave={onSave}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Save Configuration/ }));
    await waitFor(() => {
      expect(onSave).toHaveBeenCalled();
    });
    const [ids, leader] = onSave.mock.calls[0];
    expect(ids).toEqual(expect.arrayContaining(["u1", "u2"]));
    expect(leader).toBe("u1");
  });

  it("renders saveError when provided", () => {
    render(
      <MemberManagementModal
        unitId="u-1"
        unitName="Engineering"
        allMembers={allMembers}
        currentMemberIds={[]}
        currentLeaderId={null}
        onClose={vi.fn()}
        onSave={vi.fn()}
        saveError="Backend rejected"
      />,
    );
    expect(screen.getByText("Backend rejected")).toBeInTheDocument();
  });

  it("calls onClose when Escape key is pressed", () => {
    const onClose = vi.fn();
    render(
      <MemberManagementModal
        unitId="u-1"
        unitName="Engineering"
        allMembers={allMembers}
        currentMemberIds={[]}
        currentLeaderId={null}
        onClose={onClose}
        onSave={vi.fn()}
      />,
    );
    fireEvent.keyDown(window, { key: "Escape" });
    expect(onClose).toHaveBeenCalled();
  });

  it("calls onClose when Cancel button is clicked", () => {
    const onClose = vi.fn();
    render(
      <MemberManagementModal
        unitId="u-1"
        unitName="Engineering"
        allMembers={allMembers}
        currentMemberIds={[]}
        currentLeaderId={null}
        onClose={onClose}
        onSave={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Cancel/i }));
    expect(onClose).toHaveBeenCalled();
  });

  it("calls onClose when Close (X) button is clicked", () => {
    const onClose = vi.fn();
    render(
      <MemberManagementModal
        unitId="u-1"
        unitName="Engineering"
        allMembers={allMembers}
        currentMemberIds={[]}
        currentLeaderId={null}
        onClose={onClose}
        onSave={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByLabelText("Close"));
    expect(onClose).toHaveBeenCalled();
  });

  it("toggles leader off when same leader button is clicked again", () => {
    render(
      <MemberManagementModal
        unitId="u-1"
        unitName="Engineering"
        allMembers={allMembers}
        currentMemberIds={["u1"]}
        currentLeaderId="u1"
        onClose={vi.fn()}
        onSave={vi.fn()}
      />,
    );
    // Already leader → click "Remove leader"
    const btn = screen.getByTitle("Remove leader");
    fireEvent.click(btn);
    expect(screen.getByText(/none/)).toBeInTheDocument();
  });
});
