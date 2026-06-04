import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import * as React from "react";

vi.mock("framer-motion", () => {
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

const mocks = vi.hoisted(() => ({
  listAllProjects: vi.fn(),
  getProjectTasks: vi.fn(),
  archiveProject: vi.fn(),
}));
vi.mock("@/domain/application-lifecycle/service/project.service", () => ({
  projectService: {
    listAllProjects: (...a: unknown[]) => mocks.listAllProjects(...a),
    getProjectTasks: (...a: unknown[]) => mocks.getProjectTasks(...a),
    archiveProject: (...a: unknown[]) => mocks.archiveProject(...a),
  },
}));

import ProjectsStatusPage from "./page";
import type { Project } from "@/domain/application-lifecycle/schema/project.types";

const sampleProjects: Project[] = [
  {
    id: "p-1",
    key: "PROJ-A",
    name: "Project Alpha",
    description: "First project",
    status: "active",
    visibility: "internal",
    owner_user_id: "u-1",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-06-01T00:00:00Z",
  },
  {
    id: "p-2",
    key: "PROJ-B",
    name: "Project Beta",
    description: "Second project",
    status: "planning",
    visibility: "internal",
    owner_user_id: "u-2",
    created_at: "2026-02-01T00:00:00Z",
    updated_at: "2026-06-02T00:00:00Z",
  },
];

beforeEach(() => {
  mocks.listAllProjects.mockReset();
  mocks.getProjectTasks.mockReset();
  mocks.archiveProject.mockReset();
  mocks.listAllProjects.mockResolvedValue(sampleProjects);
  // Default: no tasks for any project
  mocks.getProjectTasks.mockResolvedValue([]);
});

describe("ProjectsStatusPage", () => {
  it("renders loading state initially", () => {
    render(<ProjectsStatusPage />);
    expect(screen.getByText(/Loading projects/)).toBeInTheDocument();
  });

  it("renders project list with summary cards", async () => {
    render(<ProjectsStatusPage />);
    await waitFor(() => {
      expect(screen.getByText("Project Alpha")).toBeInTheDocument();
    });
    expect(screen.getByText("Project Beta")).toBeInTheDocument();
    // Summary cards
    expect(screen.getByText("Total Projects")).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument(); // total = 2
  });

  it("shows 'No tasks' when project has 0 tasks", async () => {
    render(<ProjectsStatusPage />);
    await waitFor(() => {
      expect(screen.getByText("Project Alpha")).toBeInTheDocument();
    });
    await waitFor(() => {
      // Wait for task fetch to complete
      expect(mocks.getProjectTasks).toHaveBeenCalled();
    });
    expect(screen.getAllByText("No tasks").length).toBeGreaterThan(0);
  });

  it("shows actual progress % when tasks exist", async () => {
    mocks.getProjectTasks.mockImplementation(async (projectId: string) => {
      if (projectId === "p-1") {
        return [
          { id: "t1", title: "Done 1", priority: "medium", status: "done" },
          { id: "t2", title: "Done 2", priority: "medium", status: "done" },
          { id: "t3", title: "Todo 1", priority: "medium", status: "todo" },
        ];
      }
      return [];
    });
    render(<ProjectsStatusPage />);
    await waitFor(() => {
      expect(screen.getByText("Project Alpha")).toBeInTheDocument();
    });
    await waitFor(() => {
      expect(screen.getByText("67% tasks done")).toBeInTheDocument();
    });
  });

  it("filters projects by status (active only)", async () => {
    render(<ProjectsStatusPage />);
    await waitFor(() => {
      expect(screen.getByText("Project Alpha")).toBeInTheDocument();
      expect(screen.getByText("Project Beta")).toBeInTheDocument();
    });
    // Click "Active" filter — FilterBar component is shared, test interaction via testid
    const activeFilter = screen.getByText("Active", { selector: "button" });
    fireEvent.click(activeFilter);
    await waitFor(() => {
      expect(mocks.listAllProjects).toHaveBeenCalledWith({ include_archived: false });
    });
  });

  it("renders error state when listAllProjects fails", async () => {
    mocks.listAllProjects.mockReset();
    mocks.listAllProjects.mockRejectedValue(new Error("network"));
    render(<ProjectsStatusPage />);
    await waitFor(() => {
      expect(screen.getByText(/Failed to load projects data/)).toBeInTheDocument();
    });
  });

  it("renders empty state when no projects", async () => {
    mocks.listAllProjects.mockReset();
    mocks.listAllProjects.mockResolvedValue([]);
    render(<ProjectsStatusPage />);
    await waitFor(() => {
      expect(screen.getByText(/No projects matching/)).toBeInTheDocument();
    });
  });
});
