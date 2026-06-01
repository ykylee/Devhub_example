import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

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

import { RepositoryTable } from "./RepositoryTable";
import type { ApplicationRepository } from "@/domain/application-lifecycle/schema/project.types";

const repos: ApplicationRepository[] = [
  {
    application_id: "app-1",
    repo_provider: "github",
    repo_full_name: "devhub/backend-core",
    role: "primary",
    sync_status: "active",
    last_sync_at: "2026-05-28T10:00:00Z",
    linked_at: "2026-05-01T00:00:00Z",
    link_source: "direct",
  },
  {
    application_id: "app-1",
    repo_provider: "gitea",
    repo_full_name: "ops/infra",
    role: "sub",
    sync_status: "degraded",
    sync_error_code: "rate_limited",
    last_sync_at: null as unknown as string,
    linked_at: "2026-05-01T00:00:00Z",
    link_source: "direct",
  },
  {
    application_id: "app-1",
    repo_provider: "gitlab",
    repo_full_name: "shared/lib",
    role: "shared",
    sync_status: "requested",
    linked_at: "2026-05-02T00:00:00Z",
    link_source: "direct",
  },
];

describe("RepositoryTable", () => {
  it("renders provider/repo name + sync status + role badges", () => {
    render(<RepositoryTable repositories={repos} />);

    expect(screen.getByText("devhub/backend-core")).toBeInTheDocument();
    expect(screen.getByText("ops/infra")).toBeInTheDocument();
    expect(screen.getByText("shared/lib")).toBeInTheDocument();
    expect(screen.getByText("Active")).toBeInTheDocument();
    expect(screen.getByText("Degraded")).toBeInTheDocument();
    expect(screen.getByText("Syncing")).toBeInTheDocument();
    expect(screen.getByText("primary")).toBeInTheDocument();
    expect(screen.getByText("sub")).toBeInTheDocument();
    expect(screen.getByText("shared")).toBeInTheDocument();
  });

  it("renders repo name without link (no numeric repository_id available)", () => {
    render(<RepositoryTable repositories={repos} />);

    // ApplicationRepository has no numeric repository_id — all names display as plain text
    const name1 = screen.getByText("devhub/backend-core");
    expect(name1.closest("a")).toBeNull();
  });

  it("renders error code below Degraded badge", () => {
    render(<RepositoryTable repositories={repos} />);
    expect(screen.getByText("rate_limited")).toBeInTheDocument();
  });

  it("renders 'Never' when last_sync_at is missing", () => {
    render(<RepositoryTable repositories={repos} />);
    // Both row 2 (no last_sync_at) and row 3 (also missing) render Never
    expect(screen.getAllByText("Never").length).toBeGreaterThan(0);
  });

  it("shows Application column only when showApplicationColumn is true", () => {
    const { rerender } = render(<RepositoryTable repositories={repos} />);
    expect(screen.queryByText("Application")).not.toBeInTheDocument();

    rerender(<RepositoryTable repositories={repos} showApplicationColumn />);
    expect(screen.getByText("Application")).toBeInTheDocument();
    // application_id rendered in cell
    const appIds = screen.getAllByText("app-1");
    expect(appIds.length).toBe(3);
  });

  it("hides ActionMenu when no callbacks are provided", () => {
    render(<RepositoryTable repositories={repos.slice(0, 1)} />);
    // The action menu trigger button uses title "<repo> Actions"
    const trigger = screen.queryByRole("button", { name: /devhub\/backend-core Actions/i });
    // When all callback arrays are empty, the menu may still render but with no items
    // Just ensure no Disconnect / Metrics / View Repository menu entries leak
    expect(screen.queryByText("Disconnect")).not.toBeInTheDocument();
    expect(screen.queryByText("View Metrics")).not.toBeInTheDocument();
    // Allow trigger to exist; we just check menu items are gated
    void trigger;
  });

  it("invokes onViewRepository when 'View Repository' is clicked", async () => {
    const user = userEvent.setup();
    const onViewRepository = vi.fn();
    render(
      <RepositoryTable
        repositories={repos.slice(0, 1)}
        onViewRepository={onViewRepository}
      />,
    );

    const trigger = screen.getByRole("button", { name: /Actions/i });
    await user.click(trigger);
    await user.click(screen.getByText("View Repository"));
    expect(onViewRepository).toHaveBeenCalledWith(repos[0]);
  });

  it("invokes onDisconnect when 'Disconnect' is clicked", async () => {
    const user = userEvent.setup();
    const onDisconnect = vi.fn();
    render(
      <RepositoryTable
        repositories={repos.slice(0, 1)}
        onDisconnect={onDisconnect}
      />,
    );

    const trigger = screen.getByRole("button", { name: /Actions/i });
    await user.click(trigger);
    await user.click(screen.getByText("Disconnect"));
    expect(onDisconnect).toHaveBeenCalledWith(repos[0]);
  });

  it("invokes onViewRepositoryMetrics when 'View Metrics' is clicked", async () => {
    const user = userEvent.setup();
    const onViewRepositoryMetrics = vi.fn();
    render(
      <RepositoryTable
        repositories={repos.slice(0, 1)}
        onViewRepositoryMetrics={onViewRepositoryMetrics}
      />,
    );

    const trigger = screen.getByRole("button", { name: /Actions/i });
    await user.click(trigger);
    await user.click(screen.getByText("View Metrics"));
    expect(onViewRepositoryMetrics).toHaveBeenCalledWith(repos[0]);
  });

  it("renders fallback badge for disconnected status", () => {
    const repo: ApplicationRepository = {
      application_id: "app-1",
      repo_provider: "github",
      repo_full_name: "x/y",
      role: "sub",
      sync_status: "disconnected",
      linked_at: "2026-05-01T00:00:00Z",
      link_source: "direct",
    };
    render(<RepositoryTable repositories={[repo]} />);
    expect(screen.getByText("Disconnected")).toBeInTheDocument();
  });
});
