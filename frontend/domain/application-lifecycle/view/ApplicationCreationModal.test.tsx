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

const createApplication = vi.fn();
const updateApplication = vi.fn();
const getApplicationRepositories = vi.fn();
const getApplicationProjectsV2 = vi.fn();
const listStandaloneProjects = vi.fn();
const connectRepository = vi.fn();
const disconnectRepository = vi.fn();
const updateProject = vi.fn();
vi.mock("@/domain/application-lifecycle/service/project.service", () => ({
  projectService: {
    createApplication: (...a: unknown[]) => createApplication(...a),
    updateApplication: (...a: unknown[]) => updateApplication(...a),
    getApplicationRepositories: (...a: unknown[]) => getApplicationRepositories(...a),
    getApplicationProjectsV2: (...a: unknown[]) => getApplicationProjectsV2(...a),
    listStandaloneProjects: (...a: unknown[]) => listStandaloneProjects(...a),
    connectRepository: (...a: unknown[]) => connectRepository(...a),
    disconnectRepository: (...a: unknown[]) => disconnectRepository(...a),
    updateProject: (...a: unknown[]) => updateProject(...a),
  },
}));

const getUsers = vi.fn();
const getOrgHierarchy = vi.fn();
vi.mock("@/domain/organization-management/service/identity.service", () => ({
  identityService: {
    getUsers: (...a: unknown[]) => getUsers(...a),
    getOrgHierarchy: (...a: unknown[]) => getOrgHierarchy(...a),
  },
}));

const listRepositories = vi.fn();
vi.mock("@/domain/repository-integration/service/repository.service", () => ({
  repositoryService: {
    listRepositories: (...a: unknown[]) => listRepositories(...a),
  },
}));

import { ApplicationCreationModal } from "./ApplicationCreationModal";
import type { Application } from "@/lib/services/project.types";

beforeEach(() => {
  createApplication.mockReset();
  updateApplication.mockReset();
  getApplicationRepositories.mockReset();
  getApplicationProjectsV2.mockReset();
  listStandaloneProjects.mockReset();
  connectRepository.mockReset();
  disconnectRepository.mockReset();
  updateProject.mockReset();
  getUsers.mockReset();
  getOrgHierarchy.mockReset();
  listRepositories.mockReset();
  // Default mocks for resource loading
  getUsers.mockResolvedValue([
    { id: "alice", name: "Alice", email: "alice@example.com" },
    { id: "bob", name: "Bob", email: "bob@example.com" },
  ]);
  getOrgHierarchy.mockResolvedValue({
    nodes: [
      { id: "u-1", data: { label: "Engineering", type: "division" } },
    ],
    edges: [],
  });
  listRepositories.mockResolvedValue([]);
  getApplicationRepositories.mockResolvedValue([]);
  getApplicationProjectsV2.mockResolvedValue([]);
  listStandaloneProjects.mockResolvedValue([]);
});

describe("ApplicationCreationModal", () => {
  it("renders Create header in create mode", () => {
    render(<ApplicationCreationModal onClose={vi.fn()} onCreated={vi.fn()} />);
    expect(screen.getByText("Create")).toBeInTheDocument();
    expect(screen.getByLabelText("Application Key")).toBeInTheDocument();
  });

  it("renders Edit header in edit mode and disables key", async () => {
    render(
      <ApplicationCreationModal
        onClose={vi.fn()}
        onCreated={vi.fn()}
        initialData={{ id: "app-1", key: "APP1", name: "App 1" }}
      />,
    );
    expect(screen.getByText("Edit")).toBeInTheDocument();
    const keyInput = screen.getByLabelText("Application Key") as HTMLInputElement;
    expect(keyInput.disabled).toBe(true);
    expect(keyInput.value).toBe("APP1");
  });

  it("calls getUsers + getOrgHierarchy + listRepositories on mount", async () => {
    render(<ApplicationCreationModal onClose={vi.fn()} onCreated={vi.fn()} />);
    await waitFor(() => {
      expect(getUsers).toHaveBeenCalled();
    });
    expect(getOrgHierarchy).toHaveBeenCalled();
    expect(listRepositories).toHaveBeenCalled();
  });

  it("shows error when leader is empty on submit", async () => {
    const { container } = render(
      <ApplicationCreationModal onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getUsers).toHaveBeenCalled();
    });
    fireEvent.change(screen.getByLabelText("Application Key"), {
      target: { value: "APP1" },
    });
    fireEvent.change(screen.getByLabelText("Display Name"), {
      target: { value: "Test App" },
    });
    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);
    await waitFor(() => {
      expect(screen.getByText(/leader is required/)).toBeInTheDocument();
    });
  });

  it("submits createApplication and calls onCreated + onClose", async () => {
    const created: Application = {
      id: "app-new",
      key: "APPX",
      name: "App X",
      description: "",
      status: "planning",
      visibility: "internal",
      owner_user_id: "alice",
      leader_user_id: "alice",
      development_unit_id: "u-1",
      created_at: "",
      updated_at: "",
    };
    createApplication.mockResolvedValueOnce(created);

    // Provide a fallback render with no leader options so plain inputs appear.
    getUsers.mockReset();
    getUsers.mockResolvedValue([]);
    getOrgHierarchy.mockReset();
    getOrgHierarchy.mockResolvedValue({ nodes: [], edges: [] });

    const onCreated = vi.fn();
    const onClose = vi.fn();
    const { container } = render(
      <ApplicationCreationModal onClose={onClose} onCreated={onCreated} />,
    );

    await waitFor(() => {
      expect(getUsers).toHaveBeenCalled();
    });
    fireEvent.change(screen.getByLabelText("Application Key"), {
      target: { value: "APPX" },
    });
    fireEvent.change(screen.getByLabelText("Display Name"), {
      target: { value: "App X" },
    });
    // Plain inputs replace ComboBox when empty options
    fireEvent.change(screen.getByPlaceholderText("e.g. charlie"), {
      target: { value: "alice" },
    });
    fireEvent.change(screen.getByPlaceholderText("e.g. dept-eng"), {
      target: { value: "u-1" },
    });
    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);

    await waitFor(() => {
      expect(createApplication).toHaveBeenCalled();
    });
    expect(onCreated).toHaveBeenCalledWith(created);
    expect(onClose).toHaveBeenCalled();
  });

  it("shows 409 conflict error message", async () => {
    getUsers.mockReset();
    getUsers.mockResolvedValue([]);
    getOrgHierarchy.mockReset();
    getOrgHierarchy.mockResolvedValue({ nodes: [], edges: [] });

    createApplication.mockRejectedValueOnce({ status: 409 });
    const { container } = render(
      <ApplicationCreationModal onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getUsers).toHaveBeenCalled();
    });
    fireEvent.change(screen.getByLabelText("Application Key"), {
      target: { value: "APPX" },
    });
    fireEvent.change(screen.getByLabelText("Display Name"), {
      target: { value: "App X" },
    });
    fireEvent.change(screen.getByPlaceholderText("e.g. charlie"), {
      target: { value: "alice" },
    });
    fireEvent.change(screen.getByPlaceholderText("e.g. dept-eng"), {
      target: { value: "u-1" },
    });
    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);
    await waitFor(() => {
      expect(
        screen.getByText("Application key already exists."),
      ).toBeInTheDocument();
    });
  });

  it("switches visibility to public when public button is clicked", async () => {
    render(<ApplicationCreationModal onClose={vi.fn()} onCreated={vi.fn()} />);
    await waitFor(() => {
      expect(getUsers).toHaveBeenCalled();
    });
    fireEvent.click(screen.getByRole("button", { name: /public/i }));
    // selected state — class includes purple-500
    const btn = screen.getByRole("button", { name: /public/i });
    expect(btn.className).toContain("purple-500");
  });

  it("calls onClose when Escape is pressed", () => {
    const onClose = vi.fn();
    render(<ApplicationCreationModal onClose={onClose} onCreated={vi.fn()} />);
    fireEvent.keyDown(window, { key: "Escape" });
    expect(onClose).toHaveBeenCalled();
  });

  it("calls onClose when Cancel button is clicked", () => {
    const onClose = vi.fn();
    render(<ApplicationCreationModal onClose={onClose} onCreated={vi.fn()} />);
    fireEvent.click(screen.getByRole("button", { name: /Cancel/i }));
    expect(onClose).toHaveBeenCalled();
  });

  it("loads connected repositories in edit mode", async () => {
    getApplicationRepositories.mockResolvedValueOnce([
      {
        application_id: "app-1",
        repo_provider: "github",
        repo_full_name: "org/repo",
        role: "primary",
        sync_status: "active",
        linked_at: "",
      },
    ]);
    listRepositories.mockResolvedValueOnce([
      {
        repository_id: 1,
        provider_key: "github",
        full_name: "org/repo",
      },
    ]);
    render(
      <ApplicationCreationModal
        onClose={vi.fn()}
        onCreated={vi.fn()}
        initialData={{ id: "app-1", key: "APP1", name: "App 1" }}
      />,
    );
    await waitFor(() => {
      expect(getApplicationRepositories).toHaveBeenCalledWith("app-1");
    });
    await waitFor(() => {
      expect(screen.getByText(/org\/repo \(github\)/)).toBeInTheDocument();
    });
  });

  it("displays 'No repositories available' when list is empty in edit mode", async () => {
    listRepositories.mockResolvedValueOnce([]);
    render(
      <ApplicationCreationModal
        onClose={vi.fn()}
        onCreated={vi.fn()}
        initialData={{ id: "app-1", key: "APP1", name: "App 1" }}
      />,
    );
    await waitFor(() => {
      expect(listRepositories).toHaveBeenCalled();
    });
    await waitFor(() => {
      expect(
        screen.getByText("No repositories available."),
      ).toBeInTheDocument();
    });
  });
});
