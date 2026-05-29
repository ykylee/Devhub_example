import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";

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

import { ProjectTable } from "./ProjectTable";
import type { Project } from "@/domain/application-lifecycle/schema/project.types";

function makeProject(overrides: Partial<Project> = {}): Project {
  return {
    id: "proj-1",
    key: "PROJ-1",
    name: "Project Alpha",
    description: "",
    status: "active",
    visibility: "internal",
    owner_user_id: "alice",
    repository_id: 42,
    start_date: "2026-05-01",
    due_date: "2026-09-01",
    created_at: "",
    updated_at: "",
    ...overrides,
  };
}

describe("ProjectTable", () => {
  it("renders project name + key + repository id", () => {
    render(<ProjectTable projects={[makeProject()]} />);
    expect(screen.getByText("Project Alpha")).toBeInTheDocument();
    expect(screen.getByText("PROJ-1")).toBeInTheDocument();
    expect(screen.getByText(/Repo ID: 42/)).toBeInTheDocument();
  });

  it("renders link to /projects/:id", () => {
    render(<ProjectTable projects={[makeProject()]} />);
    const link = screen.getByText("Project Alpha").closest("a");
    expect(link).toHaveAttribute("href", "/projects/proj-1");
  });

  it("renders status badges for all known statuses", () => {
    const projects: Project[] = [
      makeProject({ id: "p1", key: "K1", name: "n1", status: "active" }),
      makeProject({ id: "p2", key: "K2", name: "n2", status: "planning" }),
      makeProject({ id: "p3", key: "K3", name: "n3", status: "on_hold" }),
      makeProject({ id: "p4", key: "K4", name: "n4", status: "closed" }),
      makeProject({ id: "p5", key: "K5", name: "n5", status: "archived" }),
    ];
    render(<ProjectTable projects={projects} />);
    expect(screen.getByText("Active")).toBeInTheDocument();
    expect(screen.getByText("Planning")).toBeInTheDocument();
    expect(screen.getByText("On Hold")).toBeInTheDocument();
    expect(screen.getByText("Closed")).toBeInTheDocument();
    expect(screen.getByText("Archived")).toBeInTheDocument();
  });

  it("calls onViewDetails when 'View Details' is clicked", () => {
    const onViewDetails = vi.fn();
    const p = makeProject();
    render(
      <ProjectTable projects={[p]} onViewDetails={onViewDetails} />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Actions/i }));
    fireEvent.click(screen.getByText("View Details"));
    expect(onViewDetails).toHaveBeenCalledWith(p);
  });

  it("does not render 'Edit Project' option when onEditProject is omitted", () => {
    render(<ProjectTable projects={[makeProject()]} />);
    fireEvent.click(screen.getByRole("button", { name: /Actions/i }));
    expect(screen.queryByText("Edit Project")).not.toBeInTheDocument();
  });

  it("calls onEditProject when 'Edit Project' is clicked", () => {
    const onEditProject = vi.fn();
    const p = makeProject();
    render(
      <ProjectTable projects={[p]} onEditProject={onEditProject} />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Actions/i }));
    fireEvent.click(screen.getByText("Edit Project"));
    expect(onEditProject).toHaveBeenCalledWith(p);
  });

  it("renders TBD for missing dates", () => {
    render(
      <ProjectTable
        projects={[makeProject({ start_date: undefined, due_date: undefined })]}
      />,
    );
    expect(screen.getByText(/TBD → TBD/)).toBeInTheDocument();
  });
});
