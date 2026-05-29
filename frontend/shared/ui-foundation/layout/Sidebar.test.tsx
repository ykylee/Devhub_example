import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const pathnameRef = { current: "/" };
vi.mock("next/navigation", () => ({
  usePathname: () => pathnameRef.current,
}));

vi.mock("next/link", () => {
  const React = require("react");
  return {
    default: ({ children, href, ...rest }: { children: React.ReactNode; href: string }) =>
      React.createElement("a", { href, ...rest }, children),
  };
});

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

import { useStore, type AuthenticatedActor } from "@/lib/store";
import { Sidebar } from "./Sidebar";

function resetStore(overrides: Partial<{
  actor: AuthenticatedActor | null;
  isSidebarOpen: boolean;
  isSidebarCollapsed: boolean;
}> = {}) {
  act(() => {
    useStore.setState({
      actor: overrides.actor ?? null,
      isSidebarOpen: overrides.isSidebarOpen ?? false,
      isSidebarCollapsed: overrides.isSidebarCollapsed ?? false,
    });
  });
}

const devActor: AuthenticatedActor = {
  login: "alice",
  role: "Developer",
};
const adminActor: AuthenticatedActor = {
  login: "admin",
  role: "System Admin",
};

describe("Sidebar (F-1)", () => {
  beforeEach(() => {
    pathnameRef.current = "/";
    resetStore();
  });

  it("기본 메뉴 (Applications/Repositories/Projects) 가 렌더된다", () => {
    render(<Sidebar />);
    expect(screen.getByLabelText("Applications")).toBeInTheDocument();
    expect(screen.getByLabelText("Repositories")).toBeInTheDocument();
    expect(screen.getByLabelText("Projects")).toBeInTheDocument();
  });

  it("Developer role 일 때 System (Admin only) 섹션이 노출되지 않는다", () => {
    resetStore({ actor: devActor });
    render(<Sidebar />);
    expect(screen.queryByText(/system \(admin only\)/i)).toBeNull();
    expect(screen.queryByLabelText("Admin Catalog")).toBeNull();
    expect(screen.queryByLabelText("Sys Admin Settings")).toBeNull();
  });

  it("System Admin role 일 때 admin menu (Admin Catalog, Sys Admin Settings) 가 노출된다", () => {
    resetStore({ actor: adminActor });
    render(<Sidebar />);
    expect(screen.getByText(/system \(admin only\)/i)).toBeInTheDocument();
    expect(screen.getByLabelText("Admin Catalog")).toBeInTheDocument();
    expect(screen.getByLabelText("Sys Admin Settings")).toBeInTheDocument();
  });

  it("pathname 이 메뉴 href 와 일치하면 해당 항목이 active style 을 가진다", () => {
    pathnameRef.current = "/applications";
    render(<Sidebar />);
    const link = screen.getByLabelText("Applications");
    const inner = link.querySelector("div");
    expect(inner?.className).toMatch(/text-primary/);
  });

  it("pathname 이 메뉴 href 의 sub-path 인 경우도 active 로 처리된다", () => {
    pathnameRef.current = "/applications/abc-123";
    render(<Sidebar />);
    const link = screen.getByLabelText("Applications");
    const inner = link.querySelector("div");
    expect(inner?.className).toMatch(/text-primary/);
  });

  it("Account 링크가 항상 렌더된다", () => {
    render(<Sidebar />);
    expect(screen.getByLabelText("Account Settings")).toBeInTheDocument();
  });

  it("mobile 닫기 (X) 버튼 클릭 시 setSidebarOpen(false) 가 호출된다", async () => {
    resetStore({ isSidebarOpen: true });
    const user = userEvent.setup();
    render(<Sidebar />);
    const close = screen.getByLabelText("Close sidebar");
    await user.click(close);
    expect(useStore.getState().isSidebarOpen).toBe(false);
  });

  it("collapse toggle 클릭 시 isSidebarCollapsed 가 토글된다", async () => {
    const user = userEvent.setup();
    render(<Sidebar />);
    // mounted=true → isSidebarCollapsed=false 일 때 'Collapse sidebar' aria-label.
    const toggle = screen.getByLabelText("Collapse sidebar");
    await user.click(toggle);
    expect(useStore.getState().isSidebarCollapsed).toBe(true);
  });

  it("collapsed=true 일 때 toggle aria-label 이 'Expand sidebar' 로 변경되고 클릭 시 false 로 돌린다", async () => {
    resetStore({ isSidebarCollapsed: true });
    const user = userEvent.setup();
    render(<Sidebar />);
    const toggle = screen.getByLabelText("Expand sidebar");
    await user.click(toggle);
    expect(useStore.getState().isSidebarCollapsed).toBe(false);
  });

  it("isSidebarOpen=true 면 backdrop 이 렌더되고 클릭 시 sidebar 닫힘", async () => {
    resetStore({ isSidebarOpen: true });
    const user = userEvent.setup();
    const { container } = render(<Sidebar />);
    const backdrop = container.querySelector("div.fixed.inset-0.bg-background\\/80") as HTMLElement;
    expect(backdrop).not.toBeNull();
    await user.click(backdrop);
    expect(useStore.getState().isSidebarOpen).toBe(false);
  });

  it("메뉴 항목 클릭 시 (Link) setSidebarOpen(false) 가 호출된다 (모바일 자동 닫힘)", async () => {
    resetStore({ isSidebarOpen: true });
    const user = userEvent.setup();
    render(<Sidebar />);
    const link = screen.getByLabelText("Applications");
    await user.click(link);
    expect(useStore.getState().isSidebarOpen).toBe(false);
  });
});
