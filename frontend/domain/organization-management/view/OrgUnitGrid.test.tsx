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

import { OrgUnitGrid } from "./OrgUnitGrid";
import type { OrgNode } from "@/domain/organization-management/service/identity.service";

const nodes: OrgNode[] = [
  {
    id: "u1",
    type: "division",
    data: { label: "Engineering", type: "division", leader_id: "alice" },
    position: { x: 0, y: 0 },
  },
  {
    id: "u2",
    type: "team",
    data: { label: "Backend Team", type: "team" },
    position: { x: 0, y: 0 },
  },
  {
    id: "u3",
    type: "group",
    data: { label: "Infra Group", type: "group", leader_id: "bob" },
    position: { x: 0, y: 0 },
  },
];

describe("OrgUnitGrid", () => {
  it("renders unit cards with label + type", () => {
    render(<OrgUnitGrid nodes={nodes} unitMembers={{}} onManage={vi.fn()} />);
    expect(screen.getByText("Engineering")).toBeInTheDocument();
    expect(screen.getByText("Backend Team")).toBeInTheDocument();
    expect(screen.getByText("Infra Group")).toBeInTheDocument();
  });

  it("renders leader badge when leader_id is set", () => {
    render(<OrgUnitGrid nodes={nodes} unitMembers={{}} onManage={vi.fn()} />);
    expect(screen.getByText("Leader: alice")).toBeInTheDocument();
    expect(screen.getByText("Leader: bob")).toBeInTheDocument();
  });

  it("renders 0 members text when no current members and shows placeholder avatar", () => {
    render(<OrgUnitGrid nodes={[nodes[0]]} unitMembers={{}} onManage={vi.fn()} />);
    expect(screen.getByText("0 members")).toBeInTheDocument();
  });

  it("renders member count from unitMembers map", () => {
    render(
      <OrgUnitGrid
        nodes={[nodes[0]]}
        unitMembers={{ u1: ["m1", "m2", "m3", "m4"] }}
        onManage={vi.fn()}
      />,
    );
    expect(screen.getByText("4 members")).toBeInTheDocument();
  });

  it("calls onManage with unit id when Manage is clicked", () => {
    const onManage = vi.fn();
    render(
      <OrgUnitGrid
        nodes={[nodes[0]]}
        unitMembers={{}}
        onManage={onManage}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Manage/i }));
    expect(onManage).toHaveBeenCalledWith("u1");
  });

  it("renders fallback icon for unknown types without crashing", () => {
    const odd: OrgNode[] = [
      {
        id: "u-odd",
        type: undefined,
        data: { label: "Odd Unit", type: "unknown_type" },
        position: { x: 0, y: 0 },
      },
    ];
    render(<OrgUnitGrid nodes={odd} unitMembers={{}} onManage={vi.fn()} />);
    expect(screen.getByText("Odd Unit")).toBeInTheDocument();
  });
});
