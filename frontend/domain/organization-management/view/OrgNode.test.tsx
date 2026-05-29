import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";

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

vi.mock("@xyflow/react", () => ({
  Handle: () => null,
  Position: { Top: "top", Bottom: "bottom" },
}));

import { OrgNode } from "./OrgNode";
import type { OrgMember } from "@/domain/organization-management/service/identity.service";

const users: OrgMember[] = [
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
];

function renderNode(dataOverrides: Record<string, unknown> = {}, selected = false) {
  const props = {
    id: "node-1",
    data: {
      label: "Engineering",
      type: "division",
      direct_count: 5,
      total_count: 20,
      users,
      ...dataOverrides,
    },
    selected,
    // Other NodeProps unused fields
    type: "orgNode",
    dragging: false,
    isConnectable: false,
    xPos: 0,
    yPos: 0,
    zIndex: 0,
  };
  // The component memo signature accepts NodeProps; here we render and cast props.
  return render(
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (<OrgNode {...(props as any)} />),
  );
}

describe("OrgNode", () => {
  it("renders label + type in view mode", () => {
    renderNode();
    expect(screen.getByText("Engineering")).toBeInTheDocument();
    expect(screen.getAllByText("division").length).toBeGreaterThan(0);
  });

  it("renders leader name when leader_id matches a user in data.users", () => {
    renderNode({ leader_id: "u1" });
    expect(screen.getByText("Alice")).toBeInTheDocument();
  });

  it("renders leader id when no matching user is found", () => {
    renderNode({ leader_id: "unknown" });
    expect(screen.getByText("unknown")).toBeInTheDocument();
  });

  it("renders 'No Leader' when leader_id is empty", () => {
    renderNode();
    expect(screen.getByText("No Leader")).toBeInTheDocument();
  });

  it("renders direct/total count", () => {
    renderNode();
    expect(screen.getByText("5")).toBeInTheDocument();
    expect(screen.getByText("20")).toBeInTheDocument();
  });

  it("calls onDelete when delete button is clicked (view mode)", () => {
    const onDelete = vi.fn();
    renderNode({ onDelete }, true);
    const delBtn = screen.getByTitle("Delete");
    fireEvent.click(delBtn);
    expect(onDelete).toHaveBeenCalledWith("node-1");
  });

  it("calls onAddChild when Add Child button is clicked (view mode)", () => {
    const onAddChild = vi.fn();
    renderNode({ onAddChild }, true);
    const addBtn = screen.getByTitle("Add Child");
    fireEvent.click(addBtn);
    expect(onAddChild).toHaveBeenCalledWith("node-1");
  });

  it("enters edit mode when Edit button is clicked", () => {
    renderNode({}, true);
    const editBtn = screen.getByTitle("Edit");
    fireEvent.click(editBtn);
    // edit mode shows Type / Name labels
    expect(screen.getByText("Name")).toBeInTheDocument();
    expect(screen.getByText("Leader")).toBeInTheDocument();
  });

  it("calls onUpdate with new label when Save is clicked in edit mode", () => {
    const onUpdate = vi.fn();
    renderNode({ onUpdate, isInitialEditing: true });
    // Already in edit mode (initial)
    expect(screen.getByText("Name")).toBeInTheDocument();
    // Find the autoFocus input and change value
    const inputs = screen.getAllByRole("textbox");
    const nameInput = inputs.find(
      (i) => (i as HTMLInputElement).value === "Engineering",
    ) as HTMLInputElement;
    fireEvent.change(nameInput, { target: { value: "Engineering Renamed" } });
    fireEvent.click(screen.getByText(/Save/i));
    expect(onUpdate).toHaveBeenCalledWith(
      "node-1",
      expect.objectContaining({ label: "Engineering Renamed", isInitialEditing: false }),
    );
  });

  it("calls onDelete when Cancel is clicked in initial editing mode", () => {
    const onDelete = vi.fn();
    renderNode({ onDelete, isInitialEditing: true });
    fireEvent.click(screen.getByText(/Discard/i));
    expect(onDelete).toHaveBeenCalledWith("node-1");
  });

  it("exits edit mode without onDelete when Cancel is clicked after entering Edit mode", () => {
    const onDelete = vi.fn();
    renderNode({ onDelete }, true);
    fireEvent.click(screen.getByTitle("Edit"));
    fireEvent.click(screen.getByText(/Cancel/i));
    expect(onDelete).not.toHaveBeenCalled();
    // back to view mode — label shown
    expect(screen.getByText("Engineering")).toBeInTheDocument();
  });

  it("calls onToggleExpand when hasChildren is true and toggle button is clicked", () => {
    const onToggleExpand = vi.fn();
    renderNode({ onToggleExpand, hasChildren: true, isExpanded: false }, true);
    fireEvent.click(screen.getByTitle(/Expand|Collapse/));
    expect(onToggleExpand).toHaveBeenCalledWith("node-1");
  });
});
