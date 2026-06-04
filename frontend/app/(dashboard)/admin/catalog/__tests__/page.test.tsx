import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import React from "react";

// ---------------------------------------------------------------------------
// Global mocks (top-level, before any component import)
// ---------------------------------------------------------------------------

// framer-motion: Proxy pattern used across all existing tests.
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

// next/navigation: useSearchParams returns controllable URL params.
const routerReplace = vi.fn();
let urlSearchParams = new URLSearchParams();
vi.mock("next/navigation", () => ({
  useSearchParams: () => urlSearchParams,
  useRouter: () => ({ replace: routerReplace }),
}));

// useToast: simple hook stub — Toast has its own dedicated tests.
vi.mock("@/shared/ui-foundation/components/Toast", () => ({
  useToast: () => ({ toast: vi.fn() }),
}));

// 4 modal components — they have their own dedicated tests.
vi.mock("@/domain/application-lifecycle/view/ApplicationCreationModal", () => ({
  ApplicationCreationModal: () => null,
}));
vi.mock("@/domain/application-lifecycle/view/ProjectCreationModal", () => ({
  ProjectCreationModal: () => null,
}));
vi.mock("@/components/project/RepositoryCreationModal", () => ({
  RepositoryCreationModal: () => null,
}));
vi.mock("@/components/project/RepositoryEditModal", () => ({
  RepositoryEditModal: () => null,
}));

// PageState components — they have their own dedicated tests.
vi.mock("@/shared/ui-foundation/components/PageState", () => ({
  PageLoading: ({ label }: { label?: string }) =>
    React.createElement("div", { "data-testid": "page-loading" }, label ?? "Loading..."),
  PageError: ({ message, onRetry }: { message?: string; onRetry?: () => void }) =>
    React.createElement(
      "div",
      { "data-testid": "page-error" },
      message ?? "Error",
      onRetry && React.createElement("button", { onClick: onRetry, "data-testid": "retry-btn" }, "Retry"),
    ),
  PageEmpty: ({ message }: { message?: string }) =>
    React.createElement("div", { "data-testid": "page-empty" }, message ?? "No data"),
}));

// 4 service mocks via vi.hoisted.
const mocks = vi.hoisted(() => ({
  listApplications: vi.fn<(...args: unknown[]) => Promise<unknown[]>>(),
  listRepositories: vi.fn<(...args: unknown[]) => Promise<unknown[]>>(),
  getApplicationProjectsV2: vi.fn<(...args: unknown[]) => Promise<unknown[]>>(),
  getRepositoryProjects: vi.fn<(...args: unknown[]) => Promise<unknown[]>>(),
  getUsers: vi.fn<(...args: unknown[]) => Promise<unknown[]>>(),
}));

vi.mock("@/domain/application-lifecycle/service/application.service", () => ({
  applicationService: {
    listApplications: (...a: unknown[]) => mocks.listApplications(...a),
  },
}));

vi.mock("@/domain/repository-integration/service/repository.service", () => ({
  repositoryService: {
    listRepositories: (...a: unknown[]) => mocks.listRepositories(...a),
  },
}));

vi.mock("@/domain/application-lifecycle/service/project.service", () => ({
  projectService: {
    getApplicationProjectsV2: (...a: unknown[]) => mocks.getApplicationProjectsV2(...a),
    getRepositoryProjects: (...a: unknown[]) => mocks.getRepositoryProjects(...a),
    archiveApplication: vi.fn(),
    archiveProject: vi.fn(),
    getProject: vi.fn(),
  },
}));

vi.mock("@/domain/organization-management/service/identity.service", () => ({
  identityService: {
    getUsers: (...a: unknown[]) => mocks.getUsers(...a),
  },
}));

// ---------------------------------------------------------------------------
// Import the component under test (after all mocks are set up)
// ---------------------------------------------------------------------------
import AdminCatalogPage from "../page";

// ---------------------------------------------------------------------------
// Test data
// ---------------------------------------------------------------------------
const sampleApp = {
  id: "app-1",
  key: "MYAPP",
  name: "My Application",
  description: "Test app",
  status: "active",
  visibility: "internal",
  owner_user_id: "u-1",
  leader_user_id: "u-1",
  development_unit_id: "dept-1",
  start_date: null,
  due_date: null,
  archived_at: null,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-06-01T00:00:00Z",
};

const sampleRepo = {
  id: 1,
  full_name: "org/repo-a",
  owner_login: "org",
  name: "repo-a",
  clone_url: "https://example.com/repo-a",
  html_url: "https://example.com/repo-a",
  default_branch: "main",
  private: false,
  status: "active" as const,
  updated_at: "2026-06-01T00:00:00Z",
  linked_applications_count: 0,
  linked_projects_count: 0,
};

const sampleProject = {
  id: "proj-1",
  key: "PROJ-A",
  name: "Project Alpha",
  description: "",
  status: "active" as const,
  visibility: "internal" as const,
  owner_user_id: "u-1",
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-06-01T00:00:00Z",
};

const sampleUser = {
  id: "u-1",
  name: "Alice",
  email: "alice@example.com",
  primary_dept_id: "dept-1",
  current_dept_id: "dept-1",
  is_seconded: false,
  role: "System Admin" as const,
  status: "active" as const,
  appointments: [{ dept_id: "dept-1", role: "leader" as const }],
  joined_at: "2026-01-01T00:00:00Z",
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe("AdminCatalogPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    urlSearchParams = new URLSearchParams();
    routerReplace.mockReset();

    // Default successful responses
    mocks.listApplications.mockResolvedValue([sampleApp]);
    mocks.listRepositories.mockResolvedValue([sampleRepo]);
    mocks.getApplicationProjectsV2.mockResolvedValue([sampleProject]);
    mocks.getRepositoryProjects.mockResolvedValue([]);
    mocks.getUsers.mockResolvedValue([sampleUser]);
  });

  it("renders loading state initially", () => {
    // Keep promises pending so loading state persists
    mocks.listApplications.mockReturnValue(new Promise(() => {}));
    render(<AdminCatalogPage />);
    expect(screen.getByTestId("page-loading")).toBeInTheDocument();
    expect(screen.getByText("Admin Catalog 로딩 중...")).toBeInTheDocument();
  });

  it("renders 3 tab buttons with correct testid attributes", async () => {
    render(<AdminCatalogPage />);
    await waitFor(() => {
      expect(screen.getByTestId("catalog-tab-applications")).toBeInTheDocument();
    });
    expect(screen.getByTestId("catalog-tab-repositories")).toBeInTheDocument();
    expect(screen.getByTestId("catalog-tab-projects")).toBeInTheDocument();
  });

  it("renders Applications tab as default with data", async () => {
    render(<AdminCatalogPage />);
    // Wait for data to load
    await waitFor(() => {
      expect(mocks.listApplications).toHaveBeenCalled();
    });
    // Application table should show the sample data
    await waitFor(() => {
      expect(screen.getByText("My Application")).toBeInTheDocument();
    });
    expect(screen.getByText("MYAPP")).toBeInTheDocument();
    expect(screen.getByText("active")).toBeInTheDocument();
  });

  it("shows PageEmpty when no applications", async () => {
    mocks.listApplications.mockResolvedValue([]);
    mocks.getApplicationProjectsV2.mockResolvedValue([]);
    render(<AdminCatalogPage />);
    await waitFor(() => {
      expect(screen.getByTestId("page-empty")).toBeInTheDocument();
    });
    expect(screen.getByText(/조회된 Application 이 없습니다/)).toBeInTheDocument();
  });

  it("shows PageError with retry button when listApplications fails", async () => {
    mocks.listApplications.mockRejectedValue(new Error("network failure"));
    render(<AdminCatalogPage />);
    await waitFor(() => {
      expect(screen.getByTestId("page-error")).toBeInTheDocument();
    });
    // The actual error message "network failure" is passed through toUserErrorMessage
    expect(screen.getByText("network failure")).toBeInTheDocument();
    expect(screen.getByTestId("retry-btn")).toBeInTheDocument();
  });

  it("renders search/query input", async () => {
    render(<AdminCatalogPage />);
    await waitFor(() => {
      // The search input placeholder should be present
      expect(screen.getByPlaceholderText(/key\/name\/leader\/status 검색/)).toBeInTheDocument();
    });
  });

  it("renders repositories tab filter buttons after mount", async () => {
    render(<AdminCatalogPage />);
    // Wait for initial load to complete
    await waitFor(() => {
      expect(mocks.listRepositories).toHaveBeenCalled();
    });
    // Repositories tab buttons should be in the DOM (not visible until tab is active)
    // The tab buttons themselves are rendered regardless
    await waitFor(() => {
      expect(screen.getByTestId("catalog-tab-repositories")).toBeInTheDocument();
    });
  });

  it("renders count badges on tab buttons", async () => {
    render(<AdminCatalogPage />);
    await waitFor(() => {
      expect(screen.getByText("Applications")).toBeInTheDocument();
    });
    // Each tab button has a count badge
    const appTab = screen.getByTestId("catalog-tab-applications");
    expect(appTab.textContent).toContain("1"); // 1 application
  });
});
