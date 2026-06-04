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

import { PlatformTable } from "./PlatformTable";
import type { Platform } from "@/domain/platform-lifecycle/schema/project.types";

function makeApp(overrides: Partial<Platform> = {}): Platform {
  return {
    id: "app-1",
    key: "APP-1",
    name: "App One",
    description: "First app",
    status: "active",
    visibility: "internal",
    owner_user_id: "alice",
    leader_user_id: "bob",
    development_unit_id: "u-1",
    start_date: "2026-01-01",
    due_date: "2026-12-31",
    created_at: "",
    updated_at: "",
    ...overrides,
  };
}

describe("PlatformTable", () => {
  it("renders empty state when no platforms", () => {
    render(
      <PlatformTable
        platforms={[]}
        onEdit={vi.fn()}
        onArchive={vi.fn()}
        onViewRepositories={vi.fn()}
      />,
    );
    expect(screen.getByText("No Platforms Found")).toBeInTheDocument();
  });

  it("renders platform name, key, description", () => {
    render(
      <PlatformTable
        platforms={[makeApp()]}
        onEdit={vi.fn()}
        onArchive={vi.fn()}
        onViewRepositories={vi.fn()}
      />,
    );
    expect(screen.getByText("App One")).toBeInTheDocument();
    expect(screen.getByText("APP-1")).toBeInTheDocument();
    expect(screen.getByText("First app")).toBeInTheDocument();
  });

  it("renders status badges for all known statuses", () => {
    const apps: Platform[] = [
      makeApp({ id: "a1", key: "K1", name: "n1", status: "active" }),
      makeApp({ id: "a2", key: "K2", name: "n2", status: "planning" }),
      makeApp({ id: "a3", key: "K3", name: "n3", status: "on_hold" }),
      makeApp({ id: "a4", key: "K4", name: "n4", status: "closed" }),
      makeApp({ id: "a5", key: "K5", name: "n5", status: "archived" }),
    ];
    render(
      <PlatformTable
        platforms={apps}
        onEdit={vi.fn()}
        onArchive={vi.fn()}
        onViewRepositories={vi.fn()}
      />,
    );
    expect(screen.getByText("Active")).toBeInTheDocument();
    expect(screen.getByText("Planning")).toBeInTheDocument();
    expect(screen.getByText("On Hold")).toBeInTheDocument();
    expect(screen.getByText("Closed")).toBeInTheDocument();
    expect(screen.getByText("Archived")).toBeInTheDocument();
  });

  it("renders visibility badges (public, internal, restricted)", () => {
    const apps: Platform[] = [
      makeApp({ id: "a1", key: "K1", name: "n1", visibility: "public" }),
      makeApp({ id: "a2", key: "K2", name: "n2", visibility: "internal" }),
      makeApp({ id: "a3", key: "K3", name: "n3", visibility: "restricted" }),
    ];
    render(
      <PlatformTable
        platforms={apps}
        onEdit={vi.fn()}
        onArchive={vi.fn()}
        onViewRepositories={vi.fn()}
      />,
    );
    expect(screen.getByText("Public")).toBeInTheDocument();
    expect(screen.getByText("Internal")).toBeInTheDocument();
    expect(screen.getByText("Restricted")).toBeInTheDocument();
  });

  it("renders fallback for empty description", () => {
    render(
      <PlatformTable
        platforms={[makeApp({ description: "" })]}
        onEdit={vi.fn()}
        onArchive={vi.fn()}
        onViewRepositories={vi.fn()}
      />,
    );
    expect(screen.getByText("No description provided.")).toBeInTheDocument();
  });

  it("calls onViewRepositories when 'View Repositories' is clicked", () => {
    const onViewRepositories = vi.fn();
    const app = makeApp();
    render(
      <PlatformTable
        platforms={[app]}
        onEdit={vi.fn()}
        onArchive={vi.fn()}
        onViewRepositories={onViewRepositories}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Actions/i }));
    fireEvent.click(screen.getByText("View Repositories"));
    expect(onViewRepositories).toHaveBeenCalledWith(app);
  });

  it("calls onEdit when 'Edit Meta' is clicked", () => {
    const onEdit = vi.fn();
    const app = makeApp();
    render(
      <PlatformTable
        platforms={[app]}
        onEdit={onEdit}
        onArchive={vi.fn()}
        onViewRepositories={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Actions/i }));
    fireEvent.click(screen.getByText("Edit Meta"));
    expect(onEdit).toHaveBeenCalledWith(app);
  });

  it("calls onArchive when 'Archive' is clicked", () => {
    const onArchive = vi.fn();
    const app = makeApp();
    render(
      <PlatformTable
        platforms={[app]}
        onEdit={vi.fn()}
        onArchive={onArchive}
        onViewRepositories={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Actions/i }));
    fireEvent.click(screen.getByText("Archive"));
    expect(onArchive).toHaveBeenCalledWith(app);
  });

  it("renders TBD for missing dates", () => {
    render(
      <PlatformTable
        platforms={[makeApp({ start_date: undefined, due_date: undefined })]}
        onEdit={vi.fn()}
        onArchive={vi.fn()}
        onViewRepositories={vi.fn()}
      />,
    );
    expect(screen.getAllByText(/TBD/).length).toBeGreaterThan(0);
  });
});
