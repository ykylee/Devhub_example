import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";

import { OrgUnitTable } from "./OrgUnitTable";
import type {
  OrgMember,
  OrgNode,
} from "@/domain/organization-management/service/identity.service";

const nodes: OrgNode[] = [
  {
    id: "unit-eng",
    type: "division",
    data: { label: "Engineering", type: "division", leader_id: "u1" },
    position: { x: 0, y: 0 },
  },
  {
    id: "unit-ops",
    type: "team",
    data: { label: "Operations", type: "team" },
    position: { x: 0, y: 0 },
  },
  {
    id: "unit-bio",
    type: "group",
    // Leader id present but with no matching member in members list -> falls
    // back to leader_id string display.
    data: { label: "Biotech", type: "group", leader_id: "ghost-id" },
    position: { x: 0, y: 0 },
  },
];

const members: OrgMember[] = [
  {
    id: "u1",
    name: "Alice",
    email: "alice@example.com",
    primary_dept_id: "unit-eng",
    current_dept_id: "unit-eng",
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
    primary_dept_id: "unit-eng",
    current_dept_id: "unit-eng",
    is_seconded: false,
    role: "Manager",
    status: "active",
    appointments: [],
    joined_at: "",
  },
];

function setup(overrides: Partial<React.ComponentProps<typeof OrgUnitTable>> = {}) {
  const onManage = vi.fn();
  const onEdit = vi.fn();
  const onDelete = vi.fn();
  const baseProps: React.ComponentProps<typeof OrgUnitTable> = {
    nodes,
    members,
    unitLeaders: {},
    unitMembers: { "unit-eng": ["u1", "u2"], "unit-ops": [] },
    onManage,
    onEdit,
    onDelete,
    ...overrides,
  };
  const utils = render(<OrgUnitTable {...baseProps} />);
  return { ...utils, onManage, onEdit, onDelete };
}

describe("OrgUnitTable", () => {
  it("renders all unit rows + count badge", () => {
    setup();
    expect(screen.getByText("Engineering")).toBeInTheDocument();
    expect(screen.getByText("Operations")).toBeInTheDocument();
    expect(screen.getByText("Biotech")).toBeInTheDocument();
    expect(screen.getByText(/3 units found/i)).toBeInTheDocument();
  });

  it("renders unit id under label + type chip", () => {
    setup();
    expect(screen.getByText("unit-eng")).toBeInTheDocument();
    // Type column renders three occurrences of types
    expect(screen.getByText("division")).toBeInTheDocument();
    expect(screen.getByText("team")).toBeInTheDocument();
    expect(screen.getByText("group")).toBeInTheDocument();
  });

  it("prefers unitLeaders override for leader name resolution", () => {
    // unitLeaders override beats node.data.leader_id
    setup({ unitLeaders: { "unit-eng": "u2" } });
    expect(screen.getByText("Bob")).toBeInTheDocument();
    expect(screen.queryByText("Alice")).not.toBeInTheDocument();
  });

  it("falls back to node.data.leader_id when unitLeaders is missing", () => {
    setup();
    expect(screen.getByText("Alice")).toBeInTheDocument();
  });

  it("shows leader id literal when no member with that id exists", () => {
    setup();
    expect(screen.getByText("ghost-id")).toBeInTheDocument();
  });

  it("renders 'No Leader' placeholder when leader is empty", () => {
    setup();
    // unit-ops has no leader_id and no override -> placeholder
    expect(screen.getByText("No Leader")).toBeInTheDocument();
  });

  it("renders member count from unitMembers map", () => {
    setup();
    // unit-eng has 2 members
    expect(screen.getByText("2")).toBeInTheDocument();
    // unit-ops has 0 members (empty array)
    const zeros = screen.getAllByText("0");
    expect(zeros.length).toBeGreaterThanOrEqual(1);
  });

  it("renders 0 member count when unitMembers entry is missing", () => {
    // unit-bio has no entry in unitMembers -> 0 via `?.length || 0`
    setup();
    const zeros = screen.getAllByText("0");
    expect(zeros.length).toBeGreaterThanOrEqual(2);
  });

  it("filters nodes by label substring (case insensitive)", () => {
    setup();
    const input = screen.getByPlaceholderText(/Search units/i);
    fireEvent.change(input, { target: { value: "ENG" } });
    expect(screen.getByText("Engineering")).toBeInTheDocument();
    expect(screen.queryByText("Operations")).not.toBeInTheDocument();
    expect(screen.queryByText("Biotech")).not.toBeInTheDocument();
    expect(screen.getByText(/1 units found/i)).toBeInTheDocument();
  });

  it("filters nodes by type substring (case insensitive)", () => {
    setup();
    const input = screen.getByPlaceholderText(/Search units/i);
    fireEvent.change(input, { target: { value: "team" } });
    expect(screen.getByText("Operations")).toBeInTheDocument();
    expect(screen.queryByText("Engineering")).not.toBeInTheDocument();
  });

  it("renders empty-state row when no rows match the filter", () => {
    setup();
    const input = screen.getByPlaceholderText(/Search units/i);
    fireEvent.change(input, { target: { value: "no-match-xyz" } });
    expect(screen.getByText(/No matching units found/i)).toBeInTheDocument();
    expect(screen.getByText(/0 units found/i)).toBeInTheDocument();
  });

  it("calls onManage when the Members action button is clicked", () => {
    const { onManage } = setup();
    const buttons = screen.getAllByRole("button", { name: /Members/i });
    fireEvent.click(buttons[0]);
    expect(onManage).toHaveBeenCalledWith("unit-eng");
  });

  it("calls onEdit when the Edit action button is clicked", () => {
    const { onEdit } = setup();
    fireEvent.click(screen.getByLabelText("Edit Engineering"));
    expect(onEdit).toHaveBeenCalledWith("unit-eng");
  });

  it("calls onDelete when the Delete action button is clicked", () => {
    const { onDelete } = setup();
    fireEvent.click(screen.getByLabelText("Delete Operations"));
    expect(onDelete).toHaveBeenCalledWith("unit-ops");
  });

  it("renders leader avatar initial uppercase", () => {
    setup({ unitLeaders: { "unit-eng": "u1" } });
    // Alice -> A. Avatar div contains a single uppercase letter.
    const letters = screen.getAllByText("A");
    expect(letters.length).toBeGreaterThanOrEqual(1);
  });
});
