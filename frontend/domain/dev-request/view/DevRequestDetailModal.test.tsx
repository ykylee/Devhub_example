import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
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

const register = vi.fn();
const reject = vi.fn();
const reassign = vi.fn();
vi.mock("@/domain/dev-request/service/dev_request.service", () => ({
  devRequestService: {
    register: (...args: unknown[]) => register(...args),
    reject: (...args: unknown[]) => reject(...args),
    reassign: (...args: unknown[]) => reassign(...args),
  },
}));

import { DevRequestDetailModal } from "./DevRequestDetailModal";
import type { DevRequest } from "@/domain/dev-request/schema/dev_request.types";

function makeRequest(overrides: Partial<DevRequest> = {}): DevRequest {
  return {
    id: "dr-1",
    title: "Sample req",
    details: "body content",
    requester: "alice",
    assignee_user_id: "bob",
    source_system: "jira",
    external_ref: "JIRA-1",
    status: "pending",
    received_at: "2026-05-29T10:00:00Z",
    created_at: "2026-05-29T10:00:00Z",
    updated_at: "2026-05-29T10:00:00Z",
    ...overrides,
  };
}

beforeEach(() => {
  register.mockReset();
  reject.mockReset();
  reassign.mockReset();
});

describe("DevRequestDetailModal", () => {
  it("renders title + details + requester/assignee + external_ref", () => {
    render(
      <DevRequestDetailModal
        request={makeRequest()}
        isSystemAdmin={false}
        onClose={vi.fn()}
        onChanged={vi.fn()}
      />,
    );
    expect(screen.getByText("Sample req")).toBeInTheDocument();
    expect(screen.getByText("body content")).toBeInTheDocument();
    expect(screen.getByText("alice")).toBeInTheDocument();
    expect(screen.getByText("bob")).toBeInTheDocument();
    expect(screen.getByText("JIRA-1")).toBeInTheDocument();
  });

  it("does not render action buttons when isSystemAdmin is false", () => {
    render(
      <DevRequestDetailModal
        request={makeRequest()}
        isSystemAdmin={false}
        onClose={vi.fn()}
        onChanged={vi.fn()}
      />,
    );
    expect(screen.queryByText("Register as Application")).not.toBeInTheDocument();
    expect(screen.queryByText("Register as Project")).not.toBeInTheDocument();
    expect(screen.queryByText("Reject")).not.toBeInTheDocument();
    expect(screen.queryByText("Reassign")).not.toBeInTheDocument();
  });

  it("renders all admin action buttons when isSystemAdmin + status pending", () => {
    render(
      <DevRequestDetailModal
        request={makeRequest({ status: "pending" })}
        isSystemAdmin
        onClose={vi.fn()}
        onChanged={vi.fn()}
      />,
    );
    expect(screen.getByText("Register as Application")).toBeInTheDocument();
    expect(screen.getByText("Register as Project")).toBeInTheDocument();
    expect(screen.getByText("Reject")).toBeInTheDocument();
    expect(screen.getByText("Reassign")).toBeInTheDocument();
    expect(screen.getByText("Promote to Project")).toBeInTheDocument();
  });

  it("renders registered banner when status is registered", () => {
    render(
      <DevRequestDetailModal
        request={makeRequest({
          status: "registered",
          registered_target_type: "application",
          registered_target_id: "app-9",
        })}
        isSystemAdmin
        onClose={vi.fn()}
        onChanged={vi.fn()}
      />,
    );
    expect(screen.getByText(/Registered as/)).toBeInTheDocument();
    expect(screen.getByText("app-9")).toBeInTheDocument();
  });

  it("renders rejected banner with reason when status is rejected", () => {
    render(
      <DevRequestDetailModal
        request={makeRequest({ status: "rejected", rejected_reason: "duplicate" })}
        isSystemAdmin
        onClose={vi.fn()}
        onChanged={vi.fn()}
      />,
    );
    expect(screen.getByText(/Rejected — duplicate/)).toBeInTheDocument();
  });

  it("calls onPromote when 'Promote to Project' is clicked", async () => {
    const user = userEvent.setup();
    const onPromote = vi.fn();
    const req = makeRequest({ status: "pending" });
    render(
      <DevRequestDetailModal
        request={req}
        isSystemAdmin
        onClose={vi.fn()}
        onChanged={vi.fn()}
        onPromote={onPromote}
      />,
    );
    await user.click(screen.getByText("Promote to Project"));
    expect(onPromote).toHaveBeenCalledWith(req);
  });

  it("switches to register mode and submits success", async () => {
    const user = userEvent.setup();
    const onChanged = vi.fn();
    const onClose = vi.fn();
    const updated = makeRequest({ status: "registered" });
    register.mockResolvedValueOnce(updated);

    render(
      <DevRequestDetailModal
        request={makeRequest({ status: "pending" })}
        isSystemAdmin
        onClose={onClose}
        onChanged={onChanged}
      />,
    );

    await user.click(screen.getByText("Register as Application"));
    const input = screen.getByPlaceholderText("application id (uuid)");
    fireEvent.change(input, { target: { value: "app-99" } });
    await user.click(screen.getByText(/Confirm/));

    await waitFor(() => {
      expect(register).toHaveBeenCalledWith("dr-1", {
        target_type: "application",
        target_id: "app-99",
      });
    });
    expect(onChanged).toHaveBeenCalledWith(updated);
    expect(onClose).toHaveBeenCalled();
  });

  it("shows error when register target_id is empty", async () => {
    const user = userEvent.setup();
    render(
      <DevRequestDetailModal
        request={makeRequest({ status: "pending" })}
        isSystemAdmin
        onClose={vi.fn()}
        onChanged={vi.fn()}
      />,
    );
    await user.click(screen.getByText("Register as Application"));
    await user.click(screen.getByText(/Confirm/));
    expect(screen.getByText("target_id is required")).toBeInTheDocument();
  });

  it("displays error when register service rejects", async () => {
    const user = userEvent.setup();
    register.mockRejectedValueOnce(new Error("conflict"));
    render(
      <DevRequestDetailModal
        request={makeRequest({ status: "pending" })}
        isSystemAdmin
        onClose={vi.fn()}
        onChanged={vi.fn()}
      />,
    );
    await user.click(screen.getByText("Register as Application"));
    const input = screen.getByPlaceholderText("application id (uuid)");
    fireEvent.change(input, { target: { value: "app-1" } });
    await user.click(screen.getByText(/Confirm/));
    await waitFor(() => {
      expect(screen.getByText("conflict")).toBeInTheDocument();
    });
  });

  it("rejects with reason and calls onChanged + onClose", async () => {
    const user = userEvent.setup();
    const onChanged = vi.fn();
    const onClose = vi.fn();
    const updated = makeRequest({ status: "rejected" });
    reject.mockResolvedValueOnce(updated);

    render(
      <DevRequestDetailModal
        request={makeRequest({ status: "pending" })}
        isSystemAdmin
        onClose={onClose}
        onChanged={onChanged}
      />,
    );
    await user.click(screen.getByText("Reject"));
    const ta = screen.getByPlaceholderText(/중복 의뢰/);
    fireEvent.change(ta, { target: { value: "duplicate" } });
    await user.click(screen.getByRole("button", { name: "Reject" }));

    await waitFor(() => {
      expect(reject).toHaveBeenCalledWith("dr-1", "duplicate");
    });
    expect(onChanged).toHaveBeenCalledWith(updated);
    expect(onClose).toHaveBeenCalled();
  });

  it("shows error when reject reason is empty", async () => {
    const user = userEvent.setup();
    render(
      <DevRequestDetailModal
        request={makeRequest({ status: "pending" })}
        isSystemAdmin
        onClose={vi.fn()}
        onChanged={vi.fn()}
      />,
    );
    await user.click(screen.getByText("Reject"));
    await user.click(screen.getByRole("button", { name: "Reject" }));
    expect(screen.getByText("rejected_reason is required")).toBeInTheDocument();
  });

  it("reassigns to a new assignee and calls service", async () => {
    const user = userEvent.setup();
    const updated = makeRequest({ assignee_user_id: "charlie" });
    reassign.mockResolvedValueOnce(updated);

    render(
      <DevRequestDetailModal
        request={makeRequest({ status: "pending", assignee_user_id: "bob" })}
        isSystemAdmin
        onClose={vi.fn()}
        onChanged={vi.fn()}
      />,
    );
    await user.click(screen.getByText("Reassign"));
    const input = screen.getByPlaceholderText("new assignee user_id");
    fireEvent.change(input, { target: { value: "charlie" } });
    await user.click(screen.getByRole("button", { name: "Reassign" }));

    await waitFor(() => {
      expect(reassign).toHaveBeenCalledWith("dr-1", "charlie");
    });
  });

  it("shows error when reassign target is same as current assignee", async () => {
    const user = userEvent.setup();
    render(
      <DevRequestDetailModal
        request={makeRequest({ status: "pending", assignee_user_id: "bob" })}
        isSystemAdmin
        onClose={vi.fn()}
        onChanged={vi.fn()}
      />,
    );
    await user.click(screen.getByText("Reassign"));
    // input pre-fills with current assignee — click Reassign w/o change
    await user.click(screen.getByRole("button", { name: "Reassign" }));
    expect(screen.getByText("new assignee_user_id is required")).toBeInTheDocument();
  });

  it("calls onClose when Escape key is pressed", () => {
    const onClose = vi.fn();
    render(
      <DevRequestDetailModal
        request={makeRequest()}
        isSystemAdmin
        onClose={onClose}
        onChanged={vi.fn()}
      />,
    );
    fireEvent.keyDown(window, { key: "Escape" });
    expect(onClose).toHaveBeenCalled();
  });

  it("calls onClose when X header button is clicked", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    render(
      <DevRequestDetailModal
        request={makeRequest()}
        isSystemAdmin
        onClose={onClose}
        onChanged={vi.fn()}
      />,
    );
    const buttons = screen.getAllByRole("button");
    await user.click(buttons[0]);
    expect(onClose).toHaveBeenCalled();
  });
});
