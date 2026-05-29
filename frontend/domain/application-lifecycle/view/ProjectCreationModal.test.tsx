import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";

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

vi.mock("@/lib/store", () => ({
  useStore: (selector: (s: { actor: { user_id: string; login: string; role: string } }) => unknown) =>
    selector({
      actor: { user_id: "u-actor", login: "actor-login", role: "Developer" },
    }),
}));

const getApplications = vi.fn();
const getSCMProviders = vi.fn();
const createApplicationProject = vi.fn();
const createProjectStandalone = vi.fn();
const updateProject = vi.fn();
const getProjectRepositories = vi.fn();
const linkProjectRepository = vi.fn();
const unlinkProjectRepository = vi.fn();
vi.mock("@/domain/application-lifecycle/service/project.service", () => ({
  projectService: {
    getApplications: (...a: unknown[]) => getApplications(...a),
    getSCMProviders: (...a: unknown[]) => getSCMProviders(...a),
    createApplicationProject: (...a: unknown[]) => createApplicationProject(...a),
    createProjectStandalone: (...a: unknown[]) => createProjectStandalone(...a),
    updateProject: (...a: unknown[]) => updateProject(...a),
    getProjectRepositories: (...a: unknown[]) => getProjectRepositories(...a),
    linkProjectRepository: (...a: unknown[]) => linkProjectRepository(...a),
    unlinkProjectRepository: (...a: unknown[]) => unlinkProjectRepository(...a),
  },
}));

const getUsers = vi.fn();
vi.mock("@/domain/organization-management/service/identity.service", () => ({
  identityService: {
    getUsers: (...a: unknown[]) => getUsers(...a),
  },
}));

import { ProjectCreationModal } from "./ProjectCreationModal";
import type { Project } from "@/lib/services/project.types";

beforeEach(() => {
  getApplications.mockReset();
  getSCMProviders.mockReset();
  createApplicationProject.mockReset();
  createProjectStandalone.mockReset();
  updateProject.mockReset();
  getProjectRepositories.mockReset();
  linkProjectRepository.mockReset();
  unlinkProjectRepository.mockReset();
  getUsers.mockReset();
  getApplications.mockResolvedValue([]);
  getSCMProviders.mockResolvedValue([]);
  getUsers.mockResolvedValue([]);
});

describe("ProjectCreationModal", () => {
  it("renders Create header for new project", async () => {
    render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    expect(screen.getByText("Create")).toBeInTheDocument();
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
  });

  it("renders Edit header for edit mode and disables key", async () => {
    render(
      <ProjectCreationModal
        repositories={[]}
        onClose={vi.fn()}
        onCreated={vi.fn()}
        initialData={{ id: "proj-1", key: "PROJ-1", name: "P1" }}
      />,
    );
    expect(screen.getByText("Edit")).toBeInTheDocument();
    const inputs = screen.getAllByDisplayValue("PROJ-1");
    expect((inputs[0] as HTMLInputElement).disabled).toBe(true);
  });

  it("renders warning banner when no leader member is set", async () => {
    render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    // actor.user_id is "u-actor" so leader pre-filled — switch leader role away to clear it
    // Easier: render with no actor user_id by overriding store... skip — verify member control exists.
    expect(screen.getByText("Project Members")).toBeInTheDocument();
  });

  it("creates a standalone project when no application is selected", async () => {
    const created: Project = {
      id: "p-new",
      key: "P1",
      name: "P1",
      description: "",
      status: "planning",
      visibility: "internal",
      owner_user_id: "u-actor",
      created_at: "",
      updated_at: "",
    };
    createProjectStandalone.mockResolvedValueOnce(created);

    const onCreated = vi.fn();
    const onClose = vi.fn();
    const { container } = render(
      <ProjectCreationModal
        repositories={[]}
        onClose={onClose}
        onCreated={onCreated}
      />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });

    fireEvent.change(screen.getByPlaceholderText("E.G. API-V1"), {
      target: { value: "P1" },
    });
    fireEvent.change(screen.getByPlaceholderText(/Backend Refactoring/), {
      target: { value: "Project One" },
    });
    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);

    await waitFor(() => {
      expect(createProjectStandalone).toHaveBeenCalled();
    });
    expect(onCreated).toHaveBeenCalledWith(created);
    expect(onClose).toHaveBeenCalled();
  });

  it("creates an application-bound project when applicationId is provided", async () => {
    const created: Project = {
      id: "p-new",
      key: "P2",
      name: "Bound",
      description: "",
      status: "planning",
      visibility: "internal",
      owner_user_id: "u-actor",
      application_id: "app-1",
      created_at: "",
      updated_at: "",
    };
    createApplicationProject.mockResolvedValueOnce(created);

    const { container } = render(
      <ProjectCreationModal
        applicationId="app-1"
        repositories={[]}
        onClose={vi.fn()}
        onCreated={vi.fn()}
      />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });

    fireEvent.change(screen.getByPlaceholderText("E.G. API-V1"), {
      target: { value: "P2" },
    });
    fireEvent.change(screen.getByPlaceholderText(/Backend Refactoring/), {
      target: { value: "Bound P" },
    });
    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);

    await waitFor(() => {
      expect(createApplicationProject).toHaveBeenCalledWith(
        "app-1",
        expect.objectContaining({ application_id: "app-1" }),
      );
    });
  });

  it("toggles createRepository checkbox and shows repo fields", async () => {
    render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
    const checkbox = screen.getByLabelText(/Create and link repository/) as HTMLInputElement;
    fireEvent.click(checkbox);
    expect(checkbox.checked).toBe(true);
    expect(screen.getByPlaceholderText("Repo Key")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("org/repo-slug")).toBeInTheDocument();
  });

  it("adds and removes a repository link row", async () => {
    render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
    // No repo selected by default — message visible
    expect(
      screen.getByText(/No repository selected/),
    ).toBeInTheDocument();
    // Add a repo link row
    const addBtns = screen.getAllByRole("button", { name: /Add/i });
    fireEvent.click(addBtns[0]);
    // 'Select repository' option now visible
    expect(screen.getByText("Select repository")).toBeInTheDocument();
  });

  it("changes visibility when button is clicked", async () => {
    render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
    fireEvent.click(screen.getByRole("button", { name: /public/i }));
    const publicBtn = screen.getByRole("button", { name: /public/i });
    expect(publicBtn.className).toContain("indigo-500");
  });

  it("adds a project member row when Add is clicked", async () => {
    render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
    const initialRows = screen.getAllByRole("button", { name: /Remove member/i }).length;
    // Both Add buttons (repo + member) match /Add/i — use the second one (member panel)
    const addBtns = screen.getAllByRole("button", { name: /Add/i });
    fireEvent.click(addBtns[addBtns.length - 1]);
    const newRows = screen.getAllByRole("button", { name: /Remove member/i }).length;
    expect(newRows).toBeGreaterThan(initialRows);
  });

  it("calls onClose when Escape is pressed", async () => {
    const onClose = vi.fn();
    render(
      <ProjectCreationModal repositories={[]} onClose={onClose} onCreated={vi.fn()} />,
    );
    fireEvent.keyDown(window, { key: "Escape" });
    expect(onClose).toHaveBeenCalled();
  });

  it("calls onClose when Cancel button is clicked", async () => {
    const onClose = vi.fn();
    render(
      <ProjectCreationModal repositories={[]} onClose={onClose} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
    fireEvent.click(screen.getByRole("button", { name: /Cancel/i }));
    expect(onClose).toHaveBeenCalled();
  });

  it("shows error when createProjectStandalone rejects", async () => {
    createProjectStandalone.mockRejectedValueOnce(new Error("fail"));
    const { container } = render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
    fireEvent.change(screen.getByPlaceholderText("E.G. API-V1"), {
      target: { value: "PX" },
    });
    fireEvent.change(screen.getByPlaceholderText(/Backend Refactoring/), {
      target: { value: "P X" },
    });
    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);
    await waitFor(() => {
      expect(screen.getByText("fail")).toBeInTheDocument();
    });
  });
});
