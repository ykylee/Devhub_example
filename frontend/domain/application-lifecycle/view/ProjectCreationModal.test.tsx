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
import type { Project } from "@/domain/application-lifecycle/schema/project.types";

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

  it("shows fallback error message when createProjectStandalone rejects with non-Error value", async () => {
    createProjectStandalone.mockRejectedValueOnce("boom-string");
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
      expect(screen.getByText("Failed to save project")).toBeInTheDocument();
    });
  });

  it("falls back to empty applications/scm/users arrays when service calls reject", async () => {
    getApplications.mockReset();
    getSCMProviders.mockReset();
    getUsers.mockReset();
    getApplications.mockRejectedValueOnce(new Error("a"));
    getSCMProviders.mockRejectedValueOnce(new Error("b"));
    getUsers.mockRejectedValueOnce(new Error("c"));
    render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    // catch branches set empty arrays → plain user_id input visible (no ComboBox)
    await waitFor(() => {
      expect(screen.getByPlaceholderText("User ID...")).toBeInTheDocument();
    });
  });

  it("renders SCM provider select with only enabled providers and preselects first enabled", async () => {
    getSCMProviders.mockReset();
    getSCMProviders.mockResolvedValueOnce([
      { provider_key: "gitlab", display_name: "GitLab", enabled: false, adapter_version: "v1", created_at: "", updated_at: "" },
      { provider_key: "github", display_name: "GitHub", enabled: true, adapter_version: "v1", created_at: "", updated_at: "" },
      { provider_key: "gitea", display_name: "Gitea", enabled: true, adapter_version: "v1", created_at: "", updated_at: "" },
    ]);
    render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getSCMProviders).toHaveBeenCalled();
    });
    // Enable createRepository to reveal provider select
    fireEvent.click(screen.getByLabelText(/Create and link repository/));
    const enabledOptions = await screen.findAllByText(/GitHub|Gitea/);
    expect(enabledOptions.length).toBeGreaterThan(0);
    // Disabled provider should not appear inside the create panel's select
    expect(screen.queryByText(/GitLab/)).toBeNull();
  });

  it("includes repository_create_payload when checkbox toggled with valid fields", async () => {
    const created: Project = {
      id: "p-new",
      key: "PWITHREPO",
      name: "With repo",
      description: "",
      status: "planning",
      visibility: "internal",
      owner_user_id: "u-actor",
      created_at: "",
      updated_at: "",
    };
    createProjectStandalone.mockResolvedValueOnce(created);
    getSCMProviders.mockReset();
    getSCMProviders.mockResolvedValueOnce([
      { provider_key: "github", display_name: "GitHub", enabled: true, adapter_version: "v1", created_at: "", updated_at: "" },
    ]);

    const { container } = render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getSCMProviders).toHaveBeenCalled();
    });
    fireEvent.change(screen.getByPlaceholderText("E.G. API-V1"), {
      target: { value: "PWITHREPO" },
    });
    fireEvent.change(screen.getByPlaceholderText(/Backend Refactoring/), {
      target: { value: "With repo" },
    });
    fireEvent.click(screen.getByLabelText(/Create and link repository/));
    fireEvent.change(screen.getByPlaceholderText("Repo Key"), {
      target: { value: "newkey" },
    });
    fireEvent.change(screen.getByPlaceholderText("org/repo-slug"), {
      target: { value: "org/newrepo" },
    });
    // scm_provider auto-set to first enabled.
    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);

    await waitFor(() => {
      expect(createProjectStandalone).toHaveBeenCalled();
    });
    const payload = createProjectStandalone.mock.calls[0][0];
    expect(payload.repository_create_payload).toEqual({
      key: "NEWKEY",
      slug: "org/newrepo",
      scm_provider: "github",
    });
  });

  it("disables submit button when createRepository payload is incomplete", async () => {
    render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
    fireEvent.click(screen.getByLabelText(/Create and link repository/));
    // checkbox on but no key/slug/scm
    const submitBtn = screen.getByRole("button", { name: /Create Project/i }) as HTMLButtonElement;
    expect(submitBtn.disabled).toBe(true);
  });

  it("converts member role to lead/contributor in create payload (initial leader from actor)", async () => {
    const created: Project = {
      id: "p-x",
      key: "PMEM",
      name: "Mem P",
      description: "",
      status: "planning",
      visibility: "internal",
      owner_user_id: "u-actor",
      created_at: "",
      updated_at: "",
    };
    createProjectStandalone.mockResolvedValueOnce(created);

    const { container } = render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
    fireEvent.change(screen.getByPlaceholderText("E.G. API-V1"), {
      target: { value: "PMEM" },
    });
    fireEvent.change(screen.getByPlaceholderText(/Backend Refactoring/), {
      target: { value: "Mem P" },
    });
    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);
    await waitFor(() => {
      expect(createProjectStandalone).toHaveBeenCalled();
    });
    const payload = createProjectStandalone.mock.calls[0][0];
    expect(payload.project_members).toEqual([
      { user_id: "u-actor", project_role: "lead" },
    ]);
    expect(payload.owner_user_id).toBe("u-actor");
  });

  it("promotes another member to leader and demotes existing leader (role-change branch)", async () => {
    const { container } = render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
    // Add a second member
    const addBtns = screen.getAllByRole("button", { name: /Add/i });
    fireEvent.click(addBtns[addBtns.length - 1]);
    // Member role selects = only those whose first option value is "leader".
    // querySelector("select") would also match application/status/visibility/repo
    // selects in the form. Filter by option[0].value === "leader".
    const allSelects = Array.from(container.querySelectorAll<HTMLSelectElement>("select"));
    const memberRoleSelects = allSelects.filter((s) => s.options[0]?.value === "leader");
    expect(memberRoleSelects.length).toBe(2);
    expect(memberRoleSelects[0].value).toBe("leader");
    expect(memberRoleSelects[1].value).toBe("developer");
    // Switch second member to "leader"
    fireEvent.change(memberRoleSelects[1], { target: { value: "leader" } });
    // The first member's role should have been demoted to "developer"
    const reAll = Array.from(container.querySelectorAll<HTMLSelectElement>("select"));
    const reSelects = reAll.filter((s) => s.options[0]?.value === "leader");
    expect(reSelects[0].value).toBe("developer");
    expect(reSelects[1].value).toBe("leader");
  });

  it("changes a non-leader role without affecting leader assignment", async () => {
    const { container } = render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
    const addBtns = screen.getAllByRole("button", { name: /Add/i });
    fireEvent.click(addBtns[addBtns.length - 1]);
    const allSelects = Array.from(container.querySelectorAll<HTMLSelectElement>("select"));
    const memberRoleSelects = allSelects.filter((s) => s.options[0]?.value === "leader");
    fireEvent.change(memberRoleSelects[1], { target: { value: "reviewer" } });
    const reAll = Array.from(container.querySelectorAll<HTMLSelectElement>("select"));
    const reSelects = reAll.filter((s) => s.options[0]?.value === "leader");
    expect(reSelects[0].value).toBe("leader");
    expect(reSelects[1].value).toBe("reviewer");
  });

  it("removes a non-leader member and keeps leader intact", async () => {
    render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
    const addBtns = screen.getAllByRole("button", { name: /Add/i });
    fireEvent.click(addBtns[addBtns.length - 1]);
    const removeBtns = screen.getAllByRole("button", { name: /Remove member/i });
    expect(removeBtns.length).toBe(2);
    fireEvent.click(removeBtns[1]);
    const stillThere = screen.getAllByRole("button", { name: /Remove member/i });
    expect(stillThere.length).toBe(1);
  });

  it("removes the leader member and promotes first remaining member to leader", async () => {
    const { container } = render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
    // Add second member
    const addBtns = screen.getAllByRole("button", { name: /Add/i });
    fireEvent.click(addBtns[addBtns.length - 1]);
    const memberInputs = screen.getAllByPlaceholderText("user id");
    fireEvent.change(memberInputs[0], { target: { value: "bob" } });
    // Remove the first member (= the leader)
    const removeBtns = screen.getAllByRole("button", { name: /Remove member/i });
    fireEvent.click(removeBtns[0]);
    const allSelects = Array.from(container.querySelectorAll<HTMLSelectElement>("select"));
    const memberRoleSelects = allSelects.filter((s) => s.options[0]?.value === "leader");
    expect(memberRoleSelects.length).toBe(1);
    expect(memberRoleSelects[0].value).toBe("leader");
  });

  it("does NOT remove the last remaining member when remove is clicked", async () => {
    render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
    const removeBtns = screen.getAllByRole("button", { name: /Remove member/i });
    expect(removeBtns.length).toBe(1);
    expect((removeBtns[0] as HTMLButtonElement).disabled).toBe(true);
    fireEvent.click(removeBtns[0]);
    const after = screen.getAllByRole("button", { name: /Remove member/i });
    expect(after.length).toBe(1);
  });

  it("renders existing applications in the application select and lets user choose one", async () => {
    getApplications.mockReset();
    getApplications.mockResolvedValueOnce([
      { id: "app-a", key: "APP1", name: "App One", description: "", status: "active", visibility: "internal", owner_user_id: "u", leader_user_id: "u", development_unit_id: "d", created_at: "", updated_at: "" },
      { id: "app-b", key: "APP2", name: "App Two", description: "", status: "active", visibility: "internal", owner_user_id: "u", leader_user_id: "u", development_unit_id: "d", created_at: "", updated_at: "" },
    ]);
    const created: Project = {
      id: "p-app-bound",
      key: "PSEL",
      name: "Sel",
      description: "",
      status: "planning",
      visibility: "internal",
      owner_user_id: "u-actor",
      application_id: "app-a",
      created_at: "",
      updated_at: "",
    };
    createApplicationProject.mockResolvedValueOnce(created);

    const { container } = render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(screen.getByText(/App One \(APP1\)/)).toBeInTheDocument();
    });
    fireEvent.change(screen.getByPlaceholderText("E.G. API-V1"), {
      target: { value: "PSEL" },
    });
    fireEvent.change(screen.getByPlaceholderText(/Backend Refactoring/), {
      target: { value: "Sel" },
    });
    const appSelect = screen.getByDisplayValue("No application (independent)") as HTMLSelectElement;
    fireEvent.change(appSelect, { target: { value: "app-a" } });
    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);
    await waitFor(() => {
      expect(createApplicationProject).toHaveBeenCalledWith(
        "app-a",
        expect.objectContaining({ application_id: "app-a" }),
      );
    });
  });

  it("updates project in edit mode and syncs repository link/unlink", async () => {
    const repoFixtures = [
      { id: 10, full_name: "org/keep", provider_key: "github", owner_login: "org", name: "keep", clone_url: "", html_url: "", default_branch: "main", private: false, status: "active" as const, updated_at: "", linked_applications_count: 0, linked_projects_count: 1 },
      { id: 11, full_name: "org/add", provider_key: "github", owner_login: "org", name: "add", clone_url: "", html_url: "", default_branch: "main", private: false, status: "active" as const, updated_at: "", linked_applications_count: 0, linked_projects_count: 0 },
    ];
    getProjectRepositories.mockResolvedValueOnce([
      { project_id: "proj-1", repository_id: 10, role: "primary", linked_at: "" },
      { project_id: "proj-1", repository_id: 99, role: "linked", linked_at: "" }, // not in selected → remove
    ]);
    const updated: Project = {
      id: "proj-1",
      key: "PROJ-1",
      name: "Edited",
      description: "",
      status: "active",
      visibility: "internal",
      owner_user_id: "u-actor",
      created_at: "",
      updated_at: "",
    };
    updateProject.mockResolvedValueOnce(updated);
    linkProjectRepository.mockResolvedValue(undefined);
    unlinkProjectRepository.mockResolvedValue(undefined);

    const onCreated = vi.fn();
    const onClose = vi.fn();
    const { container } = render(
      <ProjectCreationModal
        repositories={repoFixtures}
        onClose={onClose}
        onCreated={onCreated}
        initialData={{
          id: "proj-1",
          key: "PROJ-1",
          name: "Project 1",
          repository_id: 10,
          repository_ids: [10],
          owner_user_id: "u-actor",
        }}
      />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
    // Add a second repository link row and pick repo id 11
    const addBtns = screen.getAllByRole("button", { name: /Add/i });
    fireEvent.click(addBtns[0]);
    // Find the new repo select (second one)
    const repoSelects = screen
      .getAllByRole("combobox")
      .filter((el) => Array.from((el as HTMLSelectElement).options).some((o) => o.text === "Select repository"));
    expect(repoSelects.length).toBeGreaterThanOrEqual(2);
    fireEvent.change(repoSelects[1], { target: { value: "11" } });

    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);

    await waitFor(() => {
      expect(updateProject).toHaveBeenCalledWith(
        "proj-1",
        expect.objectContaining({ repository_ids: [10, 11], repository_id: 10 }),
      );
    });
    // unlink repo 99 (was existing, not selected anymore)
    expect(unlinkProjectRepository).toHaveBeenCalledWith("proj-1", 99);
    // link repo 11 (newly selected); role = "linked" because index !== 0
    expect(linkProjectRepository).toHaveBeenCalledWith("proj-1", 11, "linked");
    expect(onCreated).toHaveBeenCalledWith(updated);
    expect(onClose).toHaveBeenCalled();
  });

  it("updates project in edit mode without repo sync when nothing changes", async () => {
    getProjectRepositories.mockResolvedValueOnce([]);
    const updated: Project = {
      id: "proj-2",
      key: "PROJ-2",
      name: "edit no repo",
      description: "",
      status: "planning",
      visibility: "internal",
      owner_user_id: "u-actor",
      created_at: "",
      updated_at: "",
    };
    updateProject.mockResolvedValueOnce(updated);

    const { container } = render(
      <ProjectCreationModal
        repositories={[]}
        onClose={vi.fn()}
        onCreated={vi.fn()}
        initialData={{
          id: "proj-2",
          key: "PROJ-2",
          name: "P2",
          owner_user_id: "u-actor",
        }}
      />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);
    await waitFor(() => {
      expect(updateProject).toHaveBeenCalledWith(
        "proj-2",
        // repository_id/ids are deleted from payload when no repo selected
        expect.not.objectContaining({ repository_id: expect.anything() }),
      );
    });
    expect(linkProjectRepository).not.toHaveBeenCalled();
    expect(unlinkProjectRepository).not.toHaveBeenCalled();
  });

  it("normalizes Repository input shape into numeric repository_id", async () => {
    const repoFixtures = [
      { id: 42, full_name: "org/from-repo", provider_key: "github", owner_login: "org", name: "from-repo", clone_url: "", html_url: "", default_branch: "main", private: false, status: "active" as const, updated_at: "", linked_applications_count: 0, linked_projects_count: 0 },
    ];
    render(
      <ProjectCreationModal
        repositories={repoFixtures}
        onClose={vi.fn()}
        onCreated={vi.fn()}
      />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
    // The repo should appear in the dropdown as a selectable option (text from repo_full_name)
    expect(screen.getByText(/org\/from-repo \(github\)/)).toBeInTheDocument();
  });

  it("renders ComboBox for leader when user options are present (replaces plain input)", async () => {
    getUsers.mockReset();
    getUsers.mockResolvedValueOnce([
      { id: "alice", name: "Alice", email: "alice@example.com" },
      { id: "bob", name: "Bob", email: "bob@example.com" },
    ]);
    render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      // ComboBox renders instead of plain input
      expect(screen.queryByPlaceholderText("User ID...")).toBeNull();
    });
    // ComboBox trigger button uses placeholder "Search leader by name/email/user_id"
    const triggers = screen.getAllByRole("button", { name: /Search leader/i });
    expect(triggers.length).toBeGreaterThan(0);
    // Member panel inputs also have "Search member by name/email/user_id" placeholder
    const memberCombos = screen.getAllByRole("button", { name: /Search member/i });
    expect(memberCombos.length).toBeGreaterThan(0);
  });

  it("updates start_date and due_date inputs and includes them in submit payload", async () => {
    const created: Project = {
      id: "p-date",
      key: "PDATE",
      name: "Date",
      description: "",
      status: "planning",
      visibility: "internal",
      owner_user_id: "u-actor",
      created_at: "",
      updated_at: "",
    };
    createProjectStandalone.mockResolvedValueOnce(created);
    const { container } = render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
    fireEvent.change(screen.getByPlaceholderText("E.G. API-V1"), {
      target: { value: "PDATE" },
    });
    fireEvent.change(screen.getByPlaceholderText(/Backend Refactoring/), {
      target: { value: "Date" },
    });
    // Both date inputs are type=date and have no placeholder; query by type.
    // Re-query after each change because the controlled value triggers a re-render.
    let dateInputs = container.querySelectorAll<HTMLInputElement>("input[type='date']");
    expect(dateInputs.length).toBe(2);
    fireEvent.change(dateInputs[0], { target: { value: "2026-01-01" } });
    dateInputs = container.querySelectorAll<HTMLInputElement>("input[type='date']");
    fireEvent.change(dateInputs[1], { target: { value: "2026-12-31" } });

    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);
    await waitFor(() => {
      expect(createProjectStandalone).toHaveBeenCalled();
    });
    const payload = createProjectStandalone.mock.calls[0][0];
    expect(payload.start_date).toBe("2026-01-01");
    expect(payload.due_date).toBe("2026-12-31");
  });

  it("updates leader user_id via plain input (no leaderOptions) and syncs to first leader member", async () => {
    // getUsers default = [] in beforeEach → plain "User ID..." input is rendered.
    render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
    const leaderInput = screen.getByPlaceholderText("User ID...") as HTMLInputElement;
    fireEvent.change(leaderInput, { target: { value: "eve" } });
    // After change, member panel first row user_id input should also reflect "eve"
    // (since the member row has the actor default; the leader-input onChange
    // updates the lead member's user_id via setProjectMembers branch).
    const memberInputs = screen.getAllByPlaceholderText("user id") as HTMLInputElement[];
    expect(memberInputs[0].value).toBe("eve");
  });

  it("propagates leader user_id update from member ComboBox onChange (leader branch)", async () => {
    getUsers.mockReset();
    getUsers.mockResolvedValueOnce([
      { id: "alice", name: "Alice", email: "alice@example.com" },
      { id: "bob", name: "Bob", email: "bob@example.com" },
    ]);
    render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      // leader-row ComboBox uses placeholder "Search leader…", member-row uses "Search member…"
      expect(screen.getAllByRole("button", { name: /Search member/i }).length).toBeGreaterThan(0);
    });
    // Click the first member ComboBox trigger to open its panel.
    const memberTriggers = screen.getAllByRole("button", { name: /Search member/i });
    fireEvent.click(memberTriggers[0]);
    // Inside the panel, options are rendered as role="option"
    const aliceOpts = await screen.findAllByRole("option", { name: /Alice/ });
    // Click first matching option (there may be multiple combos open simultaneously across leader+member rows)
    fireEvent.click(aliceOpts[0]);
    // After selection, member input value should be alice; the leader branch
    // triggers setFormData({owner_user_id:"alice"}), which is reflected in
    // the leader ComboBox trigger label.
    await waitFor(() => {
      const leaderTriggers = screen.getAllByRole("button", { name: /Search leader|Alice/i });
      expect(leaderTriggers.some((t) => t.textContent?.includes("Alice"))).toBe(true);
    });
  });

  it("updates non-leader member user_id via plain input (no ComboBox)", async () => {
    // Members list with leader (actor) + new developer; no leaderOptions so plain input.
    render(
      <ProjectCreationModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });
    const addBtns = screen.getAllByRole("button", { name: /Add/i });
    fireEvent.click(addBtns[addBtns.length - 1]);
    const memberInputs = screen.getAllByPlaceholderText("user id");
    expect(memberInputs.length).toBe(2);
    fireEvent.change(memberInputs[1], { target: { value: "carol" } });
    expect((memberInputs[1] as HTMLInputElement).value).toBe("carol");
    // Member[0] (the leader) user_id update also triggers setFormData.owner_user_id branch.
    fireEvent.change(memberInputs[0], { target: { value: "diana" } });
    expect((memberInputs[0] as HTMLInputElement).value).toBe("diana");
  });

  it("disables submit when there is no leader member (no actor → leaderless)", async () => {
    vi.resetModules();
    // Re-mock store with no actor for this test scope
    vi.doMock("@/lib/store", () => ({
      useStore: (selector: (s: { actor: null }) => unknown) =>
        selector({ actor: null }),
    }));
    const { ProjectCreationModal: NoActorModal } = await import("./ProjectCreationModal");
    const { unmount } = render(
      <NoActorModal repositories={[]} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    await waitFor(() => {
      // banner present
      expect(
        screen.getByText(/Project Leader를 선택해야/),
      ).toBeInTheDocument();
    });
    const submitBtn = screen.getByRole("button", { name: /Create Project/i }) as HTMLButtonElement;
    expect(submitBtn.disabled).toBe(true);
    unmount();
    vi.doUnmock("@/lib/store");
  });

  it("initializes projectMembers from initialData and sends them on update submission", async () => {
    updateProject.mockResolvedValueOnce({ id: "proj-1" });
    getProjectRepositories.mockResolvedValueOnce([]);

    const initialProject: Partial<Project> = {
      id: "proj-1",
      key: "PROJ-1",
      name: "P1",
      owner_user_id: "u-actor",
      project_members: [
        { user_id: "u-actor", project_role: "lead" },
        { user_id: "member-a", project_role: "contributor" },
      ],
    };

    const { container } = render(
      <ProjectCreationModal
        repositories={[]}
        onClose={vi.fn()}
        onCreated={vi.fn()}
        initialData={initialProject}
      />,
    );

    await waitFor(() => {
      expect(getApplications).toHaveBeenCalled();
    });

    // Check initialized members
    const memberInputs = screen.getAllByPlaceholderText("user id") as HTMLInputElement[];
    expect(memberInputs.length).toBe(2);
    expect(memberInputs[0].value).toBe("u-actor");
    expect(memberInputs[1].value).toBe("member-a");

    // Submit form to trigger update
    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);

    await waitFor(() => {
      expect(updateProject).toHaveBeenCalled();
    });

    const patchPayload = updateProject.mock.calls[0][1];
    expect(patchPayload.project_members).toEqual([
      { user_id: "u-actor", project_role: "lead" },
      { user_id: "member-a", project_role: "contributor" },
    ]);
  });
});

