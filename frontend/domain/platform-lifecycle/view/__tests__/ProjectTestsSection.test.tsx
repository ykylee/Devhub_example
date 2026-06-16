// ProjectTestsSection component unit test (Sprint B-Tests, TC-PROJ-TESTS-COMP-01).
//
// 정공법: RepositoryTestsSection 가 import 하는 project-tests.service 의
// fetchProjectTestResults 만 vi.mock 으로 stub. PR #597 follow-up 회귀 가드와
// 동일 패턴 (PR #625 의 RepositoryKPISection/TestsSection.test 정합).

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

vi.mock("../../service/project-tests.service", () => ({
  fetchProjectTestResults: vi.fn(),
}));

import { fetchProjectTestResults } from "../../service/project-tests.service";
import { ProjectTestsSection } from "../ProjectTestsSection";

const baseResults = {
  project_id: "p-001",
  window_from: "2026-05-16T00:00:00Z",
  window_to: "2026-06-15T00:00:00Z",
  weighted_pass_rate: 0.93,
  totals: {
    success: 145, failed: 8, running: 1, cancelled: 2,
    skipped: 0, queued: 0, unknown: 0,
  },
  recent: [
    { id: 100, repository_id: 1, repository_full_name: "org/repo-a", run_external_id: "ext-100", commit_sha: "feedface1234", status: "success", branch: "main", started_at: "2026-06-15T01:00:00Z", finished_at: "2026-06-15T01:02:00Z" },
    { id: 99, repository_id: 2, repository_full_name: "org/repo-b", run_external_id: "ext-99", commit_sha: "badfeed5678", status: "failed", branch: "feat/x", started_at: "2026-06-15T00:30:00Z", finished_at: "2026-06-15T00:31:00Z" },
  ],
};

describe("ProjectTestsSection", () => {
  beforeEach(() => {
    vi.mocked(fetchProjectTestResults).mockReset();
  });

  afterEach(() => {
    cleanup();
  });

  it("TC-PROJ-TESTS-COMP-01 — happy path: 가중치 pass rate + status 분포 + recent runs 표시 + (weighted) 라벨", async () => {
    vi.mocked(fetchProjectTestResults).mockResolvedValue(baseResults);
    render(<ProjectTestsSection projectId="p-001" />);

    await waitFor(() => {
      expect(screen.getByTestId("project-tests-section")).toBeInTheDocument();
    });
    expect(screen.getByTestId("project-tests-pass-rate").textContent).toContain("93.0%");
    expect(screen.getByTestId("project-tests-status-success").textContent).toContain("145");
    expect(screen.getByTestId("project-tests-status-failed").textContent).toContain("8");
    // (weighted) 라벨 확인
    expect(screen.getByTestId("project-tests-section").textContent).toContain("(weighted)");
    // recent table — repository_full_name 표시
    const table = screen.getByTestId("project-tests-recent-table");
    expect(table).toBeInTheDocument();
    expect(table.textContent).toContain("org/repo-a");
    expect(table.textContent).toContain("org/repo-b");
    expect(table.textContent).toContain("feedfac"); // feedface1234.slice(0, 7)
    expect(table.textContent).toContain("badfeed");
    // multi-repo repo selector — repositoryFullName 노출 (data-testid 기반)
    expect(screen.getByTestId("project-tests-recent-repo-100").textContent).toContain("org/repo-a");
    expect(screen.getByTestId("project-tests-recent-repo-99").textContent).toContain("org/repo-b");
  });

  it("denom=0 케이스: weighted_pass_rate=null → '—' 표시 + status 0", async () => {
    vi.mocked(fetchProjectTestResults).mockResolvedValue({
      ...baseResults,
      weighted_pass_rate: null,
      totals: { success: 0, failed: 0, running: 0, cancelled: 0, skipped: 0, queued: 0, unknown: 0 },
      recent: [],
    });
    render(<ProjectTestsSection projectId="p-empty" />);
    await waitFor(() => {
      expect(screen.getByTestId("project-tests-section")).toBeInTheDocument();
    });
    expect(screen.getByTestId("project-tests-pass-rate").textContent).toContain("—");
    expect(screen.getByTestId("project-tests-status-success").textContent).toContain("0");
  });

  it("window selector 전환 시 refetch 검증 (7d 전환 → 새 fetch 호출)", async () => {
    vi.mocked(fetchProjectTestResults).mockResolvedValue(baseResults);
    const user = userEvent.setup();
    render(<ProjectTestsSection projectId="p-001" />);

    await waitFor(() => {
      expect(screen.getByTestId("project-tests-section")).toBeInTheDocument();
    });

    const windowSelect = screen.getByLabelText("Window");
    expect(windowSelect).toBeInTheDocument();
    const initialCalls = vi.mocked(fetchProjectTestResults).mock.calls.length;

    await user.selectOptions(windowSelect, "7d");
    await waitFor(() => {
      expect(vi.mocked(fetchProjectTestResults).mock.calls.length).toBeGreaterThan(initialCalls);
    });
    const lastCallArgs = vi.mocked(fetchProjectTestResults).mock.calls.at(-1)?.[1] as { window: string };
    expect(lastCallArgs?.window).toBe("7d");
  });

  it("error 케이스: service throw → 에러 메시지 표시 (data-testid 없음, message 로 검증)", async () => {
    vi.mocked(fetchProjectTestResults).mockRejectedValue(new Error("network down"));
    render(<ProjectTestsSection projectId="p-001" />);
    await waitFor(() => {
      expect(screen.getByText(/Failed to load project test results/i)).toBeInTheDocument();
    });
    expect(screen.getByText(/network down/)).toBeInTheDocument();
  });

  it("loading 상태: 첫 render 시 spinner 표시, results 도착 후 section 노출", async () => {
    let resolveFn: (v: typeof baseResults) => void = () => undefined;
    vi.mocked(fetchProjectTestResults).mockImplementation(
      () => new Promise((resolve) => { resolveFn = resolve; }),
    );
    render(<ProjectTestsSection projectId="p-001" />);
    // 첫 render 시 loading
    expect(screen.getByText(/Loading project test results/i)).toBeInTheDocument();
    // resolve 후 section 노출
    resolveFn(baseResults);
    await waitFor(() => {
      expect(screen.getByTestId("project-tests-section")).toBeInTheDocument();
    });
  });
});
