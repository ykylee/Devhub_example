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

import { DevRequestTable } from "./DevRequestTable";
import type { DevRequest, DevRequestStatus } from "@/domain/dev-request/schema/dev_request.types";

function makeRequest(overrides: Partial<DevRequest> = {}): DevRequest {
  return {
    id: "dr-1",
    title: "Sample request",
    details: "details body",
    requester: "alice",
    assignee_user_id: "bob",
    source_system: "jira",
    external_ref: "JIRA-123",
    status: "pending",
    received_at: "2026-05-29T10:00:00Z",
    created_at: "2026-05-29T10:00:00Z",
    updated_at: "2026-05-29T10:00:00Z",
    ...overrides,
  };
}

describe("DevRequestTable", () => {
  it("renders empty state when items is empty", () => {
    render(<DevRequestTable items={[]} onSelect={vi.fn()} />);
    expect(screen.getByText("No Dev Requests")).toBeInTheDocument();
  });

  it("renders title + requester + source + assignee + external_ref", () => {
    const items = [makeRequest()];
    render(<DevRequestTable items={items} onSelect={vi.fn()} />);
    expect(screen.getByText("Sample request")).toBeInTheDocument();
    expect(screen.getByText(/requested by alice/)).toBeInTheDocument();
    expect(screen.getByText("jira")).toBeInTheDocument();
    expect(screen.getByText("bob")).toBeInTheDocument();
    expect(screen.getByText("JIRA-123")).toBeInTheDocument();
  });

  it("omits external_ref span when external_ref is empty", () => {
    const items = [makeRequest({ external_ref: "" })];
    render(<DevRequestTable items={items} onSelect={vi.fn()} />);
    expect(screen.queryByText(/· ref/)).not.toBeInTheDocument();
  });

  it("renders StatusBadge per status (pending, in_review, registered, rejected, closed, received)", () => {
    const statuses: DevRequestStatus[] = [
      "pending",
      "in_review",
      "registered",
      "rejected",
      "closed",
      "received",
    ];
    const items = statuses.map((status, idx) =>
      makeRequest({ id: `dr-${idx}`, status, title: `Req-${status}` }),
    );
    render(<DevRequestTable items={items} onSelect={vi.fn()} />);
    expect(screen.getByText("Pending")).toBeInTheDocument();
    expect(screen.getByText("In Review")).toBeInTheDocument();
    expect(screen.getByText("Registered")).toBeInTheDocument();
    expect(screen.getByText("Rejected")).toBeInTheDocument();
    expect(screen.getByText("Closed")).toBeInTheDocument();
    // "Received" appears both as column header and as a status badge — at least 2
    expect(screen.getAllByText("Received").length).toBeGreaterThanOrEqual(2);
  });

  it("calls onSelect with the row's DevRequest when a row is clicked", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    const items = [makeRequest()];
    render(<DevRequestTable items={items} onSelect={onSelect} />);

    await user.click(screen.getByText("Sample request"));
    expect(onSelect).toHaveBeenCalledWith(items[0]);
  });

  it("renders received_at via formatReceivedAt fallback dash when empty", () => {
    const items = [makeRequest({ received_at: "" })];
    render(<DevRequestTable items={items} onSelect={vi.fn()} />);
    expect(screen.getByText("—")).toBeInTheDocument();
  });

  it("renders raw value when received_at is unparseable", () => {
    const items = [makeRequest({ received_at: "not-a-date" })];
    render(<DevRequestTable items={items} onSelect={vi.fn()} />);
    expect(screen.getByText("not-a-date")).toBeInTheDocument();
  });

  it("renders multiple rows", () => {
    const items = [
      makeRequest({ id: "dr-1", title: "Req 1" }),
      makeRequest({ id: "dr-2", title: "Req 2" }),
      makeRequest({ id: "dr-3", title: "Req 3" }),
    ];
    render(<DevRequestTable items={items} onSelect={vi.fn()} />);
    expect(screen.getByText("Req 1")).toBeInTheDocument();
    expect(screen.getByText("Req 2")).toBeInTheDocument();
    expect(screen.getByText("Req 3")).toBeInTheDocument();
  });
});
