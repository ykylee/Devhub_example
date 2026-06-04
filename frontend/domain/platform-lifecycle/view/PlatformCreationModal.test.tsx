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

const createPlatform = vi.fn();
const updatePlatform = vi.fn();
const getPlatformRepositories = vi.fn();
const getPlatformProjectsV2 = vi.fn();
const listStandaloneProjects = vi.fn();
const connectRepository = vi.fn();
const disconnectRepository = vi.fn();
const updateProject = vi.fn();
vi.mock("@/domain/platform-lifecycle/service/project.service", () => ({
  projectService: {
    createPlatform: (...a: unknown[]) => createPlatform(...a),
    updatePlatform: (...a: unknown[]) => updatePlatform(...a),
    getPlatformRepositories: (...a: unknown[]) => getPlatformRepositories(...a),
    getPlatformProjectsV2: (...a: unknown[]) => getPlatformProjectsV2(...a),
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

import { PlatformCreationModal } from "./PlatformCreationModal";
import type { Platform } from "@/domain/platform-lifecycle/schema/project.types";

beforeEach(() => {
  createPlatform.mockReset();
  updatePlatform.mockReset();
  getPlatformRepositories.mockReset();
  getPlatformProjectsV2.mockReset();
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
  getPlatformRepositories.mockResolvedValue([]);
  getPlatformProjectsV2.mockResolvedValue([]);
  listStandaloneProjects.mockResolvedValue([]);
});

describe("PlatformCreationModal", () => {
  it("renders Create header in create mode", () => {
    render(<PlatformCreationModal onClose={vi.fn()} onCreated={vi.fn()} />);
    expect(screen.getByText("Create")).toBeInTheDocument();
    expect(screen.getByLabelText("Platform Key")).toBeInTheDocument();
  });

  it("renders Edit header in edit mode and disables key", async () => {
    render(
      <PlatformCreationModal
        onClose={vi.fn()}
        onCreated={vi.fn()}
        initialData={{ id: "app-1", key: "APP1", name: "App 1" }}
      />,
    );
    expect(screen.getByText("Edit")).toBeInTheDocument();
    const keyInput = screen.getByLabelText("Platform Key") as HTMLInputElement;
    expect(keyInput.disabled).toBe(true);
    expect(keyInput.value).toBe("APP1");
  });

  it("calls getUsers + getOrgHierarchy + listRepositories on mount", async () => {
    render(<PlatformCreationModal onClose={vi.fn()} onCreated={vi.fn()} />);
    await waitFor(() => {
      expect(getUsers).toHaveBeenCalled();
    });
    expect(getOrgHierarchy).toHaveBeenCalled();
    expect(listRepositories).toHaveBeenCalled();
  });

  it("shows error when leader is empty on submit", async () => {
    const { container } = render(
      <PlatformCreationModal onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getUsers).toHaveBeenCalled();
    });
    fireEvent.change(screen.getByLabelText("Platform Key"), {
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

  it("submits createPlatform and calls onCreated + onClose", async () => {
    const created: Platform = {
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
    createPlatform.mockResolvedValueOnce(created);

    // Provide a fallback render with no leader options so plain inputs appear.
    getUsers.mockReset();
    getUsers.mockResolvedValue([]);
    getOrgHierarchy.mockReset();
    getOrgHierarchy.mockResolvedValue({ nodes: [], edges: [] });

    const onCreated = vi.fn();
    const onClose = vi.fn();
    const { container } = render(
      <PlatformCreationModal onClose={onClose} onCreated={onCreated} />,
    );

    await waitFor(() => {
      expect(getUsers).toHaveBeenCalled();
    });
    fireEvent.change(screen.getByLabelText("Platform Key"), {
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
      expect(createPlatform).toHaveBeenCalled();
    });
    expect(onCreated).toHaveBeenCalledWith(created);
    expect(onClose).toHaveBeenCalled();
  });

  it("shows 409 conflict error message", async () => {
    getUsers.mockReset();
    getUsers.mockResolvedValue([]);
    getOrgHierarchy.mockReset();
    getOrgHierarchy.mockResolvedValue({ nodes: [], edges: [] });

    createPlatform.mockRejectedValueOnce({ status: 409 });
    const { container } = render(
      <PlatformCreationModal onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getUsers).toHaveBeenCalled();
    });
    fireEvent.change(screen.getByLabelText("Platform Key"), {
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
        screen.getByText("Platform key already exists."),
      ).toBeInTheDocument();
    });
  });

  it("switches visibility to public when public button is clicked", async () => {
    render(<PlatformCreationModal onClose={vi.fn()} onCreated={vi.fn()} />);
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
    render(<PlatformCreationModal onClose={onClose} onCreated={vi.fn()} />);
    fireEvent.keyDown(window, { key: "Escape" });
    expect(onClose).toHaveBeenCalled();
  });

  it("calls onClose when Cancel button is clicked", () => {
    const onClose = vi.fn();
    render(<PlatformCreationModal onClose={onClose} onCreated={vi.fn()} />);
    fireEvent.click(screen.getByRole("button", { name: /Cancel/i }));
    expect(onClose).toHaveBeenCalled();
  });

  it("loads connected repositories in edit mode", async () => {
    getPlatformRepositories.mockResolvedValueOnce([
      {
        platform_id: "app-1",
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
      <PlatformCreationModal
        onClose={vi.fn()}
        onCreated={vi.fn()}
        initialData={{ id: "app-1", key: "APP1", name: "App 1" }}
      />,
    );
    await waitFor(() => {
      expect(getPlatformRepositories).toHaveBeenCalledWith("app-1");
    });
    await waitFor(() => {
      expect(screen.getByText(/org\/repo \(github\)/)).toBeInTheDocument();
    });
  });

  it("displays 'No repositories available' when list is empty in edit mode", async () => {
    listRepositories.mockResolvedValueOnce([]);
    render(
      <PlatformCreationModal
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

  it("falls back to empty option arrays when identity service rejects", async () => {
    getUsers.mockReset();
    getOrgHierarchy.mockReset();
    getUsers.mockRejectedValueOnce(new Error("boom"));
    getOrgHierarchy.mockResolvedValue({ nodes: [], edges: [] });
    render(<PlatformCreationModal onClose={vi.fn()} onCreated={vi.fn()} />);
    // catch branch sets leader/unit options to empty arrays → plain inputs visible
    await waitFor(() => {
      expect(screen.getByPlaceholderText("e.g. charlie")).toBeInTheDocument();
    });
    expect(screen.getByPlaceholderText("e.g. dept-eng")).toBeInTheDocument();
  });

  it("shows error when development department is empty on submit", async () => {
    getUsers.mockReset();
    getUsers.mockResolvedValue([]);
    getOrgHierarchy.mockReset();
    getOrgHierarchy.mockResolvedValue({ nodes: [], edges: [] });
    const { container } = render(
      <PlatformCreationModal onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getUsers).toHaveBeenCalled();
    });
    fireEvent.change(screen.getByLabelText("Platform Key"), {
      target: { value: "APP1" },
    });
    fireEvent.change(screen.getByLabelText("Display Name"), {
      target: { value: "Test App" },
    });
    fireEvent.change(screen.getByPlaceholderText("e.g. charlie"), {
      target: { value: "alice" },
    });
    // dept left empty
    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);
    await waitFor(() => {
      expect(screen.getByText(/Development department is required/)).toBeInTheDocument();
    });
  });

  it("shows generic error message when create rejects with non-409", async () => {
    getUsers.mockReset();
    getUsers.mockResolvedValue([]);
    getOrgHierarchy.mockReset();
    getOrgHierarchy.mockResolvedValue({ nodes: [], edges: [] });
    createPlatform.mockRejectedValueOnce(new Error("backend down"));
    const { container } = render(
      <PlatformCreationModal onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getUsers).toHaveBeenCalled();
    });
    fireEvent.change(screen.getByLabelText("Platform Key"), {
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
      expect(screen.getByText("backend down")).toBeInTheDocument();
    });
  });

  it("shows fallback error message when create rejects with non-Error value", async () => {
    getUsers.mockReset();
    getUsers.mockResolvedValue([]);
    getOrgHierarchy.mockReset();
    getOrgHierarchy.mockResolvedValue({ nodes: [], edges: [] });
    createPlatform.mockRejectedValueOnce("boom-string");
    const { container } = render(
      <PlatformCreationModal onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getUsers).toHaveBeenCalled();
    });
    fireEvent.change(screen.getByLabelText("Platform Key"), {
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
      expect(screen.getByText("Failed to save platform")).toBeInTheDocument();
    });
  });

  it("loads connected + standalone projects in edit mode and merges (dedup)", async () => {
    getPlatformProjectsV2.mockResolvedValueOnce([
      {
        id: "p-1",
        key: "P1",
        name: "Project 1",
        description: "",
        status: "active",
        visibility: "internal",
        owner_user_id: "alice",
        platform_id: "app-1",
        created_at: "",
        updated_at: "",
      },
    ]);
    listStandaloneProjects.mockResolvedValueOnce([
      {
        id: "p-1", // duplicate — should be deduped
        key: "P1",
        name: "Project 1",
        description: "",
        status: "active",
        visibility: "internal",
        owner_user_id: "alice",
        platform_id: "",
        created_at: "",
        updated_at: "",
      },
      {
        id: "p-2",
        key: "P2",
        name: "Project Two",
        description: "",
        status: "planning",
        visibility: "internal",
        owner_user_id: "bob",
        platform_id: "",
        created_at: "",
        updated_at: "",
      },
    ]);
    render(
      <PlatformCreationModal
        onClose={vi.fn()}
        onCreated={vi.fn()}
        initialData={{ id: "app-1", key: "APP1", name: "App 1" }}
      />,
    );
    await waitFor(() => {
      expect(getPlatformProjectsV2).toHaveBeenCalledWith("app-1");
    });
    await waitFor(() => {
      expect(screen.getByText(/Project 1 \(P1\)/)).toBeInTheDocument();
    });
    expect(screen.getByText(/Project Two \(P2\)/)).toBeInTheDocument();
    // already-connected p-1 → checkbox pre-checked
    const checkboxes = screen
      .getAllByRole("checkbox")
      .filter((c) => (c as HTMLInputElement).type === "checkbox");
    const p1Checkbox = checkboxes.find((c) => {
      const label = c.closest("label");
      return label?.textContent?.includes("Project 1");
    }) as HTMLInputElement | undefined;
    expect(p1Checkbox?.checked).toBe(true);
  });

  it("shows 'No projects available to connect' message in edit when both lists empty", async () => {
    getPlatformProjectsV2.mockResolvedValueOnce([]);
    listStandaloneProjects.mockResolvedValueOnce([]);
    render(
      <PlatformCreationModal
        onClose={vi.fn()}
        onCreated={vi.fn()}
        initialData={{ id: "app-1", key: "APP1", name: "App 1" }}
      />,
    );
    await waitFor(() => {
      expect(getPlatformProjectsV2).toHaveBeenCalled();
    });
    await waitFor(() => {
      expect(
        screen.getByText("No projects available to connect."),
      ).toBeInTheDocument();
    });
  });

  it("submits update with repo + project add/remove sync in edit mode", async () => {
    const updated: Platform = {
      id: "app-1",
      key: "APP1",
      name: "App 1 Updated",
      description: "",
      status: "active",
      visibility: "internal",
      owner_user_id: "alice",
      leader_user_id: "alice",
      development_unit_id: "u-1",
      created_at: "",
      updated_at: "",
    };
    // mount-time fetches
    listRepositories.mockResolvedValueOnce([
      { id: 10, full_name: "org/keep", provider_key: "github", owner_login: "org", name: "keep", clone_url: "", html_url: "", default_branch: "main", private: false, status: "active", updated_at: "", linked_applications_count: 1, linked_projects_count: 0 },
      { id: 11, full_name: "org/add", provider_key: "github", owner_login: "org", name: "add", clone_url: "", html_url: "", default_branch: "main", private: false, status: "active", updated_at: "", linked_applications_count: 0, linked_projects_count: 0 },
    ]);
    getPlatformRepositories.mockResolvedValueOnce([
      { platform_id: "app-1", repo_provider: "github", repo_full_name: "org/keep", role: "primary", sync_status: "active", linked_at: "", link_source: "direct" },
      { platform_id: "app-1", repo_provider: "github", repo_full_name: "org/remove", role: "sub", sync_status: "active", linked_at: "", link_source: "direct" },
    ]);
    getPlatformProjectsV2.mockResolvedValueOnce([
      // connected; we'll remove this one
      { id: "p-remove", key: "RM", name: "Removed", description: "", status: "active", visibility: "internal", owner_user_id: "alice", platform_id: "app-1", created_at: "", updated_at: "" },
    ]);
    listStandaloneProjects.mockResolvedValueOnce([
      // candidate to be added
      { id: "p-add", key: "AD", name: "ToAdd", description: "", status: "planning", visibility: "internal", owner_user_id: "bob", platform_id: "", created_at: "", updated_at: "" },
    ]);

    // submit-time re-fetches (current state)
    getPlatformRepositories.mockResolvedValueOnce([
      { platform_id: "app-1", repo_provider: "github", repo_full_name: "org/keep", role: "primary", sync_status: "active", linked_at: "", link_source: "direct" },
      { platform_id: "app-1", repo_provider: "github", repo_full_name: "org/remove", role: "sub", sync_status: "active", linked_at: "", link_source: "direct" },
    ]);
    getPlatformProjectsV2.mockResolvedValueOnce([
      { id: "p-remove", key: "RM", name: "Removed", description: "", status: "active", visibility: "internal", owner_user_id: "alice", platform_id: "app-1", created_at: "", updated_at: "" },
    ]);

    updatePlatform.mockResolvedValueOnce(updated);
    connectRepository.mockResolvedValue(undefined);
    disconnectRepository.mockResolvedValue(undefined);
    updateProject.mockResolvedValue(undefined);

    const onCreated = vi.fn();
    const onClose = vi.fn();
    const { container } = render(
      <PlatformCreationModal
        onClose={onClose}
        onCreated={onCreated}
        initialData={{
          id: "app-1",
          key: "APP1",
          name: "App 1",
          leader_user_id: "alice",
          development_unit_id: "u-1",
        }}
      />,
    );

    // Wait for initial loads.
    await waitFor(() => {
      expect(screen.getByText(/org\/keep \(github\)/)).toBeInTheDocument();
    });
    await waitFor(() => {
      expect(screen.getByText(/ToAdd \(AD\)/)).toBeInTheDocument();
    });

    // Toggle repo: add org/add. The remove path is exercised because
    // org/remove was pre-selected via getPlatformRepositories but is NOT
    // present in listRepositories — so on submit, selectedRepoKeys still
    // contains it (no UI row to uncheck) and the diff between current
    // (re-fetched) and selected leaves it intact. So we add org/add and
    // also explicitly disconnect org/remove by removing its key from
    // selectedRepoKeys via direct toggling of listed repos only:
    //   selected initially = ["github/org/keep", "github/org/remove"]
    //   listed rows        = ["github/org/keep", "github/org/add"]
    // After clicking "org/add" → selected adds "github/org/add".
    // To exercise the disconnect path, also uncheck "org/keep"
    // (currently checked because pre-fetched) — that triggers
    // disconnectRepository for "github/org/keep" instead of "github/org/remove".
    const addRepoLabel = screen.getByText(/org\/add \(github\)/).closest("label");
    if (!addRepoLabel) throw new Error("add repo label not found");
    const addRepoCheckbox = addRepoLabel.querySelector("input[type='checkbox']") as HTMLInputElement;
    fireEvent.click(addRepoCheckbox);
    const keepRepoLabel = screen.getByText(/org\/keep \(github\)/).closest("label");
    if (!keepRepoLabel) throw new Error("keep repo label not found");
    const keepRepoCheckbox = keepRepoLabel.querySelector("input[type='checkbox']") as HTMLInputElement;
    fireEvent.click(keepRepoCheckbox);

    // Toggle project: add p-add, remove p-remove
    const addProjLabel = screen.getByText(/ToAdd \(AD\)/).closest("label");
    if (!addProjLabel) throw new Error("add proj label not found");
    const addProjCheckbox = addProjLabel.querySelector("input[type='checkbox']") as HTMLInputElement;
    fireEvent.click(addProjCheckbox);
    const removeProjLabel = screen.getByText(/Removed \(RM\)/).closest("label");
    if (!removeProjLabel) throw new Error("remove proj label not found");
    const removeProjCheckbox = removeProjLabel.querySelector("input[type='checkbox']") as HTMLInputElement;
    fireEvent.click(removeProjCheckbox);

    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);

    await waitFor(() => {
      expect(updatePlatform).toHaveBeenCalledWith(
        "app-1",
        expect.objectContaining({ owner_user_id: "alice" }),
      );
    });
    // connect org/add and disconnect org/keep (org/keep was checked → unchecked)
    await waitFor(() => {
      expect(connectRepository).toHaveBeenCalledWith(
        "app-1",
        expect.objectContaining({ repo_provider: "github", repo_full_name: "org/add", role: "sub" }),
      );
    });
    expect(disconnectRepository).toHaveBeenCalledWith("app-1", "github", "org/keep");
    // project add → updateProject(p-add, { platform_id: "app-1" })
    expect(updateProject).toHaveBeenCalledWith("p-add", { platform_id: "app-1" });
    // project remove → updateProject(p-remove, { platform_id: "" })
    expect(updateProject).toHaveBeenCalledWith("p-remove", { platform_id: "" });
    expect(onCreated).toHaveBeenCalledWith(updated);
    expect(onClose).toHaveBeenCalled();
  });

  it("submits update without repo/project changes when nothing toggled", async () => {
    const updated: Platform = {
      id: "app-1",
      key: "APP1",
      name: "App 1",
      description: "",
      status: "active",
      visibility: "internal",
      owner_user_id: "alice",
      leader_user_id: "alice",
      development_unit_id: "u-1",
      created_at: "",
      updated_at: "",
    };
    listRepositories.mockResolvedValue([]);
    getPlatformRepositories.mockResolvedValue([]);
    getPlatformProjectsV2.mockResolvedValue([]);
    listStandaloneProjects.mockResolvedValue([]);
    updatePlatform.mockResolvedValueOnce(updated);

    const onCreated = vi.fn();
    const onClose = vi.fn();
    const { container } = render(
      <PlatformCreationModal
        onClose={onClose}
        onCreated={onCreated}
        initialData={{
          id: "app-1",
          key: "APP1",
          name: "App 1",
          leader_user_id: "alice",
          development_unit_id: "u-1",
        }}
      />,
    );

    await waitFor(() => {
      expect(getPlatformProjectsV2).toHaveBeenCalled();
    });

    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);

    await waitFor(() => {
      expect(updatePlatform).toHaveBeenCalled();
    });
    expect(connectRepository).not.toHaveBeenCalled();
    expect(disconnectRepository).not.toHaveBeenCalled();
    expect(updateProject).not.toHaveBeenCalled();
    expect(onCreated).toHaveBeenCalledWith(updated);
    expect(onClose).toHaveBeenCalled();
  });

  it("changes status select option and includes it in submit payload", async () => {
    getUsers.mockReset();
    getUsers.mockResolvedValue([]);
    getOrgHierarchy.mockReset();
    getOrgHierarchy.mockResolvedValue({ nodes: [], edges: [] });
    const created: Platform = {
      id: "n",
      key: "APPX",
      name: "App X",
      description: "",
      status: "active",
      visibility: "internal",
      owner_user_id: "alice",
      leader_user_id: "alice",
      development_unit_id: "u-1",
      created_at: "",
      updated_at: "",
    };
    createPlatform.mockResolvedValueOnce(created);

    const { container } = render(
      <PlatformCreationModal onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getUsers).toHaveBeenCalled();
    });
    const statusSelect = screen.getByDisplayValue("Planning") as HTMLSelectElement;
    fireEvent.change(statusSelect, { target: { value: "active" } });
    fireEvent.change(screen.getByLabelText("Platform Key"), {
      target: { value: "appx" }, // lowercase → key normalized to upper by createPlatform branch
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
      expect(createPlatform).toHaveBeenCalled();
    });
    const callArg = createPlatform.mock.calls[0][0];
    expect(callArg.status).toBe("active");
    // input onChange uppercases as user types (auto), but normalized again in submit
    expect(callArg.key).toBe("APPX");
  });

  it("renders 'Save Changes' label in edit mode submit button", async () => {
    render(
      <PlatformCreationModal
        onClose={vi.fn()}
        onCreated={vi.fn()}
        initialData={{ id: "app-1", key: "APP1", name: "App 1" }}
      />,
    );
    await waitFor(() => {
      expect(getUsers).toHaveBeenCalled();
    });
    expect(
      screen.getByRole("button", { name: /Save Changes/i }),
    ).toBeInTheDocument();
  });

  it("logs error when edit-mode project fetch rejects", async () => {
    const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => undefined);
    getPlatformProjectsV2.mockRejectedValueOnce(new Error("api down"));
    listStandaloneProjects.mockRejectedValueOnce(new Error("api down"));
    render(
      <PlatformCreationModal
        onClose={vi.fn()}
        onCreated={vi.fn()}
        initialData={{ id: "app-1", key: "APP1", name: "App 1" }}
      />,
    );
    await waitFor(() => {
      expect(consoleSpy).toHaveBeenCalled();
    });
    consoleSpy.mockRestore();
  });
});
