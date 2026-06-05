import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import React from "react";

// Mock framer-motion
vi.mock("framer-motion", () => {
  type AnyProps = { children?: unknown; [k: string]: unknown };
  return {
    motion: new Proxy(
      {},
      {
        get: (_target, tag) =>
          ({ children, ...props }: AnyProps) =>
            React.createElement(tag as string, props, children),
      },
    ),
    AnimatePresence: ({ children }: AnyProps) =>
      React.createElement(React.Fragment, null, children),
  };
});

// Mock next/navigation
const mockBack = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({
    back: mockBack,
    push: vi.fn(),
  }),
}));

// Mock recharts to avoid rendering layout errors in JSDOM/HappyDOM
vi.mock("recharts", () => {
  return {
    ResponsiveContainer: ({ children }: { children: React.ReactNode }) =>
      React.createElement("div", { className: "responsive-container-mock" }, children),
    AreaChart: ({ children, data }: any) =>
      React.createElement("div", { "data-testid": "area-chart", "data-data": JSON.stringify(data) }, children),
    Area: () => null,
    XAxis: () => null,
    YAxis: () => null,
    CartesianGrid: () => null,
    Tooltip: () => null,
    PieChart: ({ children }: any) =>
      React.createElement("div", { "data-testid": "pie-chart" }, children),
    Pie: ({ data }: any) =>
      React.createElement("div", { "data-testid": "pie-data", "data-data": JSON.stringify(data) }),
    Cell: () => null,
  };
});

import { RepositoryDashboardView } from "./RepositoryDashboardView";
import { repositoryService, Repository, RepositoryActivity, RepositoryDashboardData, RepositoryBuildRun } from "@/domain/repository-integration/service/repository.service";
import { useStore } from "@/lib/store";

const mockRepo: Repository = {
  id: 1,
  full_name: "test-owner/test-repo",
  owner_login: "test-owner",
  name: "test-repo",
  clone_url: "https://scm.devhub/test-owner/test-repo.git",
  html_url: "https://scm.devhub/test-owner/test-repo",
  default_branch: "main",
  private: true,
  status: "active",
  provider_id: "github-p1",
  provider_key: "github",
  updated_at: "2026-06-05T00:00:00Z",
  linked_applications_count: 1,
  linked_projects_count: 1,
};

const mockActivity: RepositoryActivity = {
  repository_id: 1,
  window_from: "2026-05-01T00:00:00Z",
  window_to: "2026-06-05T00:00:00Z",
  pr_event_count: 5,
  active_contributors: ["alice", "bob"],
  build_run_count: 10,
  build_success_rate: 0.8,
  last_build_status: "success",
  last_build_at: "2026-06-05T10:00:00Z",
};

const mockDashboardData: RepositoryDashboardData = {
  repository_id: 1,
  quality: {
    coverage: 85.5,
    duplication: 1.2,
    quality_gate: "passed",
    issues: { blocker: 0, critical: 1, major: 3 },
  },
  security: {
    security_gate: "passed",
    secrets_detected: 0,
    vulnerabilities: { high: 0, medium: 2, low: 5 },
  },
  productivity: {
    avg_pr_lead_time_hours: 4.5,
    weekly_commits: [{ week: "05/01", count: 10 }],
    weekly_prs: [{ week: "05/01", count: 2 }],
  },
  linkage: {
    linked_platforms: [{ id: "p1", name: "Core API Platform", status: "active" }],
    linked_projects: [{ id: "prj1", name: "Q2 Refactoring", status: "active" }],
  },
};

const mockBuildRuns: RepositoryBuildRun[] = [
  {
    id: 45,
    repository_id: 1,
    run_external_id: "ext-45",
    branch: "main",
    commit_sha: "abcdef1234567890",
    status: "success",
    duration_seconds: 120,
    started_at: "2026-06-05T10:00:00Z",
    finished_at: "2026-06-05T10:02:00Z",
  },
  {
    id: 46,
    repository_id: 1,
    run_external_id: "ext-46",
    branch: "feat/some",
    commit_sha: "7890abcdef123456",
    status: "failed",
    duration_seconds: 45,
    started_at: "2026-06-05T10:10:00Z",
    finished_at: "2026-06-05T10:10:45Z",
  },
];

describe("RepositoryDashboardView", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    useStore.setState({ role: null });

    // Mock service calls
    vi.spyOn(repositoryService, "getRepository").mockResolvedValue(mockRepo);
    vi.spyOn(repositoryService, "getRepositoryActivity").mockResolvedValue(mockActivity);
    vi.spyOn(repositoryService, "getRepositoryDashboardData").mockResolvedValue(mockDashboardData);
    vi.spyOn(repositoryService, "getRepositoryBuildRuns").mockResolvedValue(mockBuildRuns);
  });

  it("renders loader initially", async () => {
    render(<RepositoryDashboardView repoId={1} />);
    expect(screen.getByText(/Aggregating workspace telemetry/i)).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.queryByText(/Aggregating workspace telemetry/i)).not.toBeInTheDocument();
    });
  });

  it("renders error view when repository data fetch fails", async () => {
    vi.spyOn(repositoryService, "getRepository").mockRejectedValue(new Error("Network Error"));

    render(<RepositoryDashboardView repoId={1} />);
    await waitFor(() => {
      expect(screen.getByText(/Failed to fetch repository dashboard metrics/i)).toBeInTheDocument();
    });

    const retryBtn = screen.getByRole("button", { name: /Retry/i });
    expect(retryBtn).toBeInTheDocument();

    // Setup success resolve on retry
    vi.spyOn(repositoryService, "getRepository").mockResolvedValue(mockRepo);
    const user = userEvent.setup();
    await user.click(retryBtn);

    await waitFor(() => {
      expect(screen.getByText("test-repo")).toBeInTheDocument();
    });
  });

  it("renders Developer view by default if user role is Developer", async () => {
    useStore.setState({ role: "Developer" });
    render(<RepositoryDashboardView repoId={1} />);

    await waitFor(() => {
      expect(screen.getByText("test-repo")).toBeInTheDocument();
    });

    // Check header states
    expect(screen.getByText(/Viewing as/i)).toHaveTextContent("Viewing as Developer");

    // Developer View content should be visible
    expect(screen.getByText("Build & Integration Runs")).toBeInTheDocument();
    expect(screen.getByText("Active Pull Requests")).toBeInTheDocument();
    expect(screen.getByText("Static Analysis (SonarQube)")).toBeInTheDocument();
    expect(screen.queryByText("Team Manager Focus")).not.toBeInTheDocument();
  });

  it("renders Manager view by default if user role is Manager or System Admin", async () => {
    useStore.setState({ role: "Manager" });
    render(<RepositoryDashboardView repoId={1} />);

    await waitFor(() => {
      expect(screen.getByText("test-repo")).toBeInTheDocument();
    });

    expect(screen.getByText(/Viewing as/i)).toHaveTextContent("Viewing as Manager");

    // Manager View content should be visible
    expect(screen.getByText("Team Manager Focus")).toBeInTheDocument();
    expect(screen.getByText("Organization Admin Focus")).toBeInTheDocument();
    expect(screen.getByText("System Admin Focus")).toBeInTheDocument();
    expect(screen.queryByText("Build & Integration Runs")).not.toBeInTheDocument();
  });

  it("toggles view between Developer and Manager modes via subheader tabs", async () => {
    useStore.setState({ role: "Developer" });
    render(<RepositoryDashboardView repoId={1} />);

    await waitFor(() => {
      expect(screen.getByText("Build & Integration Runs")).toBeInTheDocument();
    });

    const user = userEvent.setup();
    const managerTab = screen.getByRole("button", { name: /Manager & Governance/i });
    await user.click(managerTab);

    // Should switch to Manager View
    expect(screen.getByText("Team Manager Focus")).toBeInTheDocument();
    expect(screen.queryByText("Build & Integration Runs")).not.toBeInTheDocument();

    const devTab = screen.getByRole("button", { name: /Developer/i });
    await user.click(devTab);

    // Should switch back to Developer View
    expect(screen.getByText("Build & Integration Runs")).toBeInTheDocument();
  });

  it("toggles contributor chart visibility in ManagerView", async () => {
    useStore.setState({ role: "Manager" });
    render(<RepositoryDashboardView repoId={1} />);

    await waitFor(() => {
      expect(screen.getByText("Team Manager Focus")).toBeInTheDocument();
    });

    // Contributor Distribution should be visible initially
    expect(screen.getByText("Contributor Distribution")).toBeInTheDocument();
    
    // Select eye toggle button. It uses lucide-react EyeOff initially to hide it, with title
    const toggleBtn = screen.getByTitle("Hide distribution chart");
    expect(toggleBtn).toBeInTheDocument();

    const user = userEvent.setup();
    await user.click(toggleBtn);

    // Should display hidden state details
    expect(screen.getByText("Contributor chart is hidden")).toBeInTheDocument();

    // Click Unhide
    const unhideBtn = screen.getByRole("button", { name: /Unhide/i });
    await user.click(unhideBtn);

    // Should bring the chart back
    expect(screen.queryByText("Contributor chart is hidden")).not.toBeInTheDocument();
  });

  it("triggers router back when back button is clicked", async () => {
    useStore.setState({ role: "Developer" });
    render(<RepositoryDashboardView repoId={1} />);

    await waitFor(() => {
      expect(screen.getByText("test-repo")).toBeInTheDocument();
    });

    // The back button has an ArrowLeft icon, click on it
    const backBtn = screen.getByRole("button", { name: "" }); // icon-only button, let's find it or query by structure
    const buttons = screen.getAllByRole("button");
    // The first button in our structure is the back button
    const backButton = buttons[0];

    const user = userEvent.setup();
    await user.click(backButton);
    expect(mockBack).toHaveBeenCalled();
  });
});
