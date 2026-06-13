import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import React from "react";

// framer-motion Proxy pattern (catalog page.test.tsx 와 동일)
vi.mock("framer-motion", () => {
  type AnyProps = { children?: React.ReactNode; [k: string]: unknown };
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

// next/navigation + next/link 최소 mock
vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: vi.fn(), push: vi.fn() }),
  usePathname: () => "/admin",
  useSearchParams: () => new URLSearchParams(),
}));
vi.mock("next/link", () => ({
  default: ({ children, href }: { children: React.ReactNode; href: string }) =>
    React.createElement("a", { href }, children),
}));

// adminX1Service mock (per-test)
const adminX1Mock = {
  listSyncJobs: vi.fn(),
  getSyncJob: vi.fn(),
  getStatusSummary: vi.fn(),
};
vi.mock("@/domain/integration-registry/service/admin-x1.service", () => ({
  adminX1Service: adminX1Mock,
}));

// ---------------------------------------------------------------------------
// SyncJobQueueWidget
// ---------------------------------------------------------------------------
describe("SyncJobQueueWidget", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders queue rows from queued + running responses", async () => {
    adminX1Mock.listSyncJobs.mockImplementation(async (opts: { status?: string }) => {
      if (opts.status === "queued") {
        return {
          items: [
            { job_id: "job-q-1", provider_id: "p1", requested_by: "admin", status: "queued", created_at: new Date().toISOString() },
          ],
          total: 1,
          limit: 10,
          offset: 0,
        };
      }
      if (opts.status === "running") {
        return {
          items: [
            { job_id: "job-r-1", provider_id: "p1", requested_by: "system", status: "running", created_at: new Date().toISOString() },
          ],
          total: 1,
          limit: 10,
          offset: 0,
        };
      }
      return { items: [], total: 0, limit: 10, offset: 0 };
    });

    const { SyncJobQueueWidget } = await import("./SyncJobQueueWidget");
    render(React.createElement(SyncJobQueueWidget));
    await waitFor(() => {
      expect(screen.getByText(/Sync Job Queue/i)).toBeInTheDocument();
    });
    // job_id prefix 렌더 확인 (8자)
    expect(screen.getByText("job-q-1")).toBeInTheDocument();
    expect(screen.getByText("job-r-1")).toBeInTheDocument();
    // status label 렌더 확인
    expect(screen.getAllByText(/Queued|Running/).length).toBeGreaterThanOrEqual(2);
  });

  it("renders empty state when queue has no items", async () => {
    adminX1Mock.listSyncJobs.mockResolvedValue({ items: [], total: 0, limit: 10, offset: 0 });
    const { SyncJobQueueWidget } = await import("./SyncJobQueueWidget");
    render(React.createElement(SyncJobQueueWidget));
    await waitFor(() => {
      expect(screen.getByText(/큐가 비어 있습니다/)).toBeInTheDocument();
    });
  });
});

// ---------------------------------------------------------------------------
// SyncJobStatusWidget
// ---------------------------------------------------------------------------
describe("SyncJobStatusWidget", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders 4 status counts from getStatusSummary", async () => {
    adminX1Mock.getStatusSummary.mockResolvedValue({
      sync_job_status_counts: { queued: 2, running: 1, succeeded: 10, failed: 3 },
    });
    const { SyncJobStatusWidget } = await import("./SyncJobStatusWidget");
    render(React.createElement(SyncJobStatusWidget));
    await waitFor(() => {
      expect(screen.getByText(/Sync Job Status/i)).toBeInTheDocument();
    });
    expect(screen.getByText("2")).toBeInTheDocument();
    expect(screen.getByText("1")).toBeInTheDocument();
    expect(screen.getByText("10")).toBeInTheDocument();
    expect(screen.getByText("3")).toBeInTheDocument();
  });
});

// ---------------------------------------------------------------------------
// DashboardSummaryWidget
// ---------------------------------------------------------------------------
describe("DashboardSummaryWidget", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("computes total/queue/failed/successRate from counts", async () => {
    adminX1Mock.getStatusSummary.mockResolvedValue({
      sync_job_status_counts: { queued: 1, running: 2, succeeded: 7, failed: 1 },
    });
    const { DashboardSummaryWidget } = await import("./DashboardSummaryWidget");
    render(React.createElement(DashboardSummaryWidget));
    await waitFor(() => {
      expect(screen.getByText(/Dashboard Summary/i)).toBeInTheDocument();
    });
    // totalJobs = 1+2+7+1 = 11
    expect(screen.getByText("11")).toBeInTheDocument();
    // queueDepth = 1+2 = 3
    expect(screen.getByText("3")).toBeInTheDocument();
    // failed = 1
    expect(screen.getByText("1")).toBeInTheDocument();
    // successRate = 7/(7+1) = 87.5%
    expect(screen.getByText("87.5%")).toBeInTheDocument();
  });
});

// ---------------------------------------------------------------------------
// ProviderHealthWidget — placeholder (endpoint 미구현)
// ---------------------------------------------------------------------------
describe("ProviderHealthWidget", () => {
  it("renders placeholder when endpoint not ready", async () => {
    const { ProviderHealthWidget } = await import("./ProviderHealthWidget");
    render(React.createElement(ProviderHealthWidget));
    await waitFor(() => {
      expect(screen.getByText(/Provider Health/i)).toBeInTheDocument();
      expect(screen.getByText(/Provider health endpoint 미구현/)).toBeInTheDocument();
      expect(screen.getByText(/ADR-0032 §3 carve/)).toBeInTheDocument();
    });
  });
});
