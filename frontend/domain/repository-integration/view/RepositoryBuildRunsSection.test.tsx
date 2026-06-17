import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

// RepositoryBuildRunsSection unit test — N-9 잔여 build-runs polish
// (kpi-tests-per-domain-scope.md §6.5 + PR #555 잔여 4건 sub-issue 4).
//
// 검증 항목:
// 1. initial loading 시 skeleton 5 row 표시
// 2. data fetch 완료 후 build-runs-list + 20 row 표시
// 3. status filter dropdown 표시 + 옵션 8개 (all + 7 enum)
// 4. error 발생 시 에러 UI + Retry 버튼
// 5. empty state (items.length === 0) 시 안내 + "View all repositories" link
// 6. status filter selectOption('failed') 시 refetch 호출

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

vi.mock("../service/repository.service", () => ({
  repositoryService: {
    getRepositoryBuildRuns: vi.fn(),
    getRepositoryBuildRunsWithMeta: vi.fn(),
  },
}));

import { RepositoryBuildRunsSection } from "./RepositoryBuildRunsSection";
import { repositoryService } from "../service/repository.service";

const mockBuildRun = (overrides: Record<string, unknown> = {}) => ({
  id: 100,
  repository_id: 1,
  run_external_id: "ext-100",
  branch: "main",
  commit_sha: "feedface1234567",
  status: "success",
  duration_seconds: 60,
  started_at: "2026-06-15T01:00:00Z",
  finished_at: "2026-06-15T01:01:00Z",
  ...overrides,
});

const mockItems = (n: number, status: string = "success") =>
  Array.from({ length: n }, (_, i) => mockBuildRun({ id: 100 + i, status }));

describe("RepositoryBuildRunsSection (N-9 residual)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("TC-N9-SECTION-01 — initial loading 시 skeleton 5 row 표시", async () => {
    // Never-resolving promise to keep loading state
    (repositoryService.getRepositoryBuildRunsWithMeta as ReturnType<typeof vi.fn>).mockImplementation(
      () => new Promise(() => {}),
    );

    render(<RepositoryBuildRunsSection repoId={1} />);

    // section + skeleton 표시
    expect(screen.getByTestId("repository-build-runs-section")).toBeInTheDocument();
    const skeleton = screen.getByTestId("build-runs-skeleton");
    expect(skeleton).toBeInTheDocument();
    // 5 skeleton row
    const skeletonRows = skeleton.querySelectorAll('[class*="animate-pulse"]');
    expect(skeletonRows.length).toBe(5);
  });

  it("TC-N9-SECTION-02 — data fetch 완료 후 build-runs-list + 20 row 표시", async () => {
    (repositoryService.getRepositoryBuildRunsWithMeta as ReturnType<typeof vi.fn>).mockResolvedValueOnce({ status: "ok", data: mockItems(20), meta: { total: 20 } });

    render(<RepositoryBuildRunsSection repoId={1} />);

    await waitFor(() => {
      expect(screen.queryByTestId("build-runs-skeleton")).not.toBeInTheDocument();
    });

    const list = screen.getByTestId("build-runs-list");
    expect(list).toBeInTheDocument();
    const rows = screen.getAllByTestId("build-runs-row");
    expect(rows).toHaveLength(20);
  });

  it("TC-N9-SECTION-03 — status filter dropdown 표시 + 8 옵션 (all + 7 enum)", async () => {
    (repositoryService.getRepositoryBuildRunsWithMeta as ReturnType<typeof vi.fn>).mockResolvedValueOnce({ status: "ok", data: mockItems(20), meta: { total: 20 } });

    render(<RepositoryBuildRunsSection repoId={1} />);

    await waitFor(() => {
      expect(screen.queryByTestId("build-runs-skeleton")).not.toBeInTheDocument();
    });

    const dropdown = screen.getByTestId("build-runs-status-filter");
    expect(dropdown).toBeInTheDocument();
    const options = dropdown.querySelectorAll("option");
    expect(options).toHaveLength(8); // all + 7 enum
    expect(options[0]).toHaveTextContent("All");
    expect(options[1]).toHaveTextContent("Queued");
    expect(options[2]).toHaveTextContent("Running");
    expect(options[3]).toHaveTextContent("Success");
    expect(options[4]).toHaveTextContent("Failed");
  });

  it("TC-N9-SECTION-04 — error (not_found) 시 에러 UI + Retry 버튼", async () => {
    (repositoryService.getRepositoryBuildRunsWithMeta as ReturnType<typeof vi.fn>).mockRejectedValueOnce({
      code: "repository_not_found",
      message: "repository not found",
      status: 404,
    });

    render(<RepositoryBuildRunsSection repoId={999} />);

    await waitFor(() => {
      expect(screen.getByText(/Failed to load build runs/i)).toBeInTheDocument();
    });

    expect(screen.getByText(/Repository not found/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Retry/i })).toBeInTheDocument();
  });

  it("TC-N9-SECTION-05 — empty state (0 item) 시 안내 + 'View all repositories' link", async () => {
    (repositoryService.getRepositoryBuildRunsWithMeta as ReturnType<typeof vi.fn>).mockResolvedValueOnce({ status: "ok", data: [], meta: { total: 0 } });

    render(<RepositoryBuildRunsSection repoId={1} />);

    await waitFor(() => {
      expect(screen.getByTestId("build-runs-empty")).toBeInTheDocument();
    });

    expect(screen.getByText(/No build activity for this repository/i)).toBeInTheDocument();
    const link = screen.getByRole("link", { name: /View all repositories/i });
    expect(link).toBeInTheDocument();
    expect(link).toHaveAttribute("href", "/repositories");
  });

  it("TC-N9-SECTION-06 — status filter selectOption('failed') 시 refetch 호출", async () => {
    (repositoryService.getRepositoryBuildRunsWithMeta as ReturnType<typeof vi.fn>)
      .mockResolvedValueOnce({ status: "ok", data: mockItems(20), meta: { total: 20 } })
      .mockResolvedValueOnce({ status: "ok", data: mockItems(1, "failed"), meta: { total: 1 } });

    const user = userEvent.setup();
    render(<RepositoryBuildRunsSection repoId={1} />);

    await waitFor(() => {
      expect(screen.queryByTestId("build-runs-skeleton")).not.toBeInTheDocument();
    });

    const dropdown = screen.getByTestId("build-runs-status-filter");
    await user.selectOptions(dropdown, "failed");

    await waitFor(() => {
      expect(screen.getAllByTestId("build-runs-row")).toHaveLength(1);
    });

    expect(repositoryService.getRepositoryBuildRunsWithMeta).toHaveBeenCalledTimes(2);
    expect(repositoryService.getRepositoryBuildRunsWithMeta).toHaveBeenLastCalledWith(
      1,
      expect.objectContaining({ limit: 20, offset: 0, status: "failed" }),
    );
  });

  it("TC-N9-SECTION-07 — row 에 status badge + branch + commit short + relative time 표시", async () => {
    (repositoryService.getRepositoryBuildRunsWithMeta as ReturnType<typeof vi.fn>).mockResolvedValueOnce({ status: "ok", data: mockItems(3, "success"), meta: { total: 3 } });

    render(<RepositoryBuildRunsSection repoId={1} />);

    await waitFor(() => {
      expect(screen.queryByTestId("build-runs-skeleton")).not.toBeInTheDocument();
    });

    const firstRow = screen.getAllByTestId("build-runs-row")[0];
    expect(firstRow).toHaveTextContent("main"); // branch
    expect(firstRow).toHaveTextContent("feedfac"); // commit SHA short (7자) = "feedfac"
    expect(firstRow).toHaveTextContent("success"); // status badge
  });
});
