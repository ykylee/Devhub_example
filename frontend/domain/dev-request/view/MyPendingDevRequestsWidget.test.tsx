import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";

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

vi.mock("next/link", () => ({
  default: ({ children, href }: { children: React.ReactNode; href: string }) => {
    const React = require("react");
    return React.createElement("a", { href }, children);
  },
}));

const getMyPending = vi.fn();
vi.mock("@/domain/dev-request/service/dev_request.service", () => ({
  devRequestService: {
    getMyPending: (...args: unknown[]) => getMyPending(...args),
  },
}));

import { MyPendingDevRequestsWidget } from "./MyPendingDevRequestsWidget";
import type { DevRequest } from "@/domain/dev-request/schema/dev_request.types";

function makeRequest(overrides: Partial<DevRequest> = {}): DevRequest {
  return {
    id: "dr-1",
    title: "Sample req",
    details: "",
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
  getMyPending.mockReset();
});

describe("MyPendingDevRequestsWidget", () => {
  it("renders loading spinner before service resolves", async () => {
    let resolveFn: (v: { data: DevRequest[]; total: number }) => void = () => {};
    getMyPending.mockReturnValueOnce(
      new Promise<{ data: DevRequest[]; total: number }>((res) => {
        resolveFn = res;
      }),
    );

    const { container } = render(<MyPendingDevRequestsWidget />);
    // spinner div uses animate-spin
    expect(container.querySelector(".animate-spin")).toBeTruthy();

    resolveFn({ data: [], total: 0 });
    await waitFor(() => {
      expect(container.querySelector(".animate-spin")).toBeFalsy();
    });
  });

  it("renders error message when service rejects with Error", async () => {
    getMyPending.mockRejectedValueOnce(new Error("boom"));
    render(<MyPendingDevRequestsWidget />);
    await waitFor(() => {
      expect(screen.getByText("boom")).toBeInTheDocument();
    });
  });

  it("renders fallback error message when rejection is not Error instance", async () => {
    getMyPending.mockRejectedValueOnce("string-rejection");
    render(<MyPendingDevRequestsWidget />);
    await waitFor(() => {
      expect(screen.getByText("Failed to load dev requests")).toBeInTheDocument();
    });
  });

  it("renders empty state when no items returned", async () => {
    getMyPending.mockResolvedValueOnce({ data: [], total: 0 });
    render(<MyPendingDevRequestsWidget />);
    await waitFor(() => {
      expect(screen.getByText("대기 중인 의뢰가 없습니다.")).toBeInTheDocument();
    });
    expect(screen.getByText("0 pending / in_review")).toBeInTheDocument();
  });

  it("renders up to 5 items + total count + link to /dev-requests", async () => {
    const items: DevRequest[] = Array.from({ length: 7 }, (_, i) =>
      makeRequest({ id: `dr-${i}`, title: `Req-${i}` }),
    );
    getMyPending.mockResolvedValueOnce({ data: items, total: 12 });
    render(<MyPendingDevRequestsWidget />);
    await waitFor(() => {
      expect(screen.getByText("Req-0")).toBeInTheDocument();
    });
    // Only 5 items rendered
    expect(screen.queryByText("Req-5")).not.toBeInTheDocument();
    expect(screen.queryByText("Req-6")).not.toBeInTheDocument();
    // Total counter reflects backend total
    expect(screen.getByText("12 pending / in_review")).toBeInTheDocument();
    // 전체 보기 link
    const links = screen.getAllByRole("link");
    const overview = links.find((l) => l.getAttribute("href") === "/dev-requests");
    expect(overview).toBeTruthy();
  });

  it("renders received_at fallback dash when value is empty", async () => {
    getMyPending.mockResolvedValueOnce({
      data: [makeRequest({ received_at: "" })],
      total: 1,
    });
    render(<MyPendingDevRequestsWidget />);
    await waitFor(() => {
      expect(screen.getByText(/jira ·/)).toBeInTheDocument();
    });
  });
});
