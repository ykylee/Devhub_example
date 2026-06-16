import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import React from "react";

// DomainPicker 단위 테스트 (Sprint D — kpi-tests-per-domain-scope.md §6.4).
//
// 검증 범위:
// 1. scope tab (Platform/Project/Repository) 클릭 시 entity list 가 해당 scope
//    의 entity 로 교체 + active tab 시각화 (aria-selected)
// 2. entity link 의 href 가 scope 별로 /platforms/[id] /projects/[id]
//    /repositories/[id] 매핑
// 3. ready: true 인 scope (Repository) 만 entity 가 link 로, 미준비 scope
//    (Platform/Project) 는 'sub-section 예정' 배지 표시
// 4. loading/error/empty 3 상태별 메시지
// 5. 정공법: page 가 entity fetch 후 props 주입 → picker 는 순수 UI
//    (외부 fetch 호출 0)

vi.mock("framer-motion", () => {
  // motion.div / motion.a 가 실제 element 를 렌더하도록 — href 등 props 가 통과되게.
  const renderTag = (tag: string) => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const Cmp = ({ children, ...props }: any) =>
      React.createElement(tag, props, children);
    Cmp.displayName = `motion.${tag}`;
    return Cmp;
  };
  return {
    motion: {
      div: renderTag("div"),
      a: renderTag("a"),
    },
    AnimatePresence: ({ children }: { children?: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
  };
});

import { DomainPicker } from "../DomainPicker";

describe("DomainPicker", () => {
  it("renders 3 scope tabs with Repository default-active", () => {
    render(
      <DomainPicker
        defaultScope="repository"
        repositories={[{ id: "1", name: "org/repo-a" }]}
      />,
    );

    const platformTab = screen.getByTestId("domain-picker-scope-platform");
    const projectTab = screen.getByTestId("domain-picker-scope-project");
    const repoTab = screen.getByTestId("domain-picker-scope-repository");

    expect(platformTab).toHaveAttribute("aria-selected", "false");
    expect(projectTab).toHaveAttribute("aria-selected", "false");
    expect(repoTab).toHaveAttribute("aria-selected", "true");

    // header text
    expect(screen.getByRole("heading", { name: /Repository Analytics/i })).toBeInTheDocument();
  });

  it("renders entity link with /repositories/[id] href (ready=true → link, no '예정' badge)", () => {
    render(
      <DomainPicker
        defaultScope="repository"
        repositories={[
          { id: "1", name: "org/repo-a" },
          { id: "2", name: "org/repo-b" },
        ]}
      />,
    );

    const link1 = screen.getByTestId("domain-picker-entity-1");
    expect(link1).toHaveAttribute("href", "/repositories/1");

    const link2 = screen.getByTestId("domain-picker-entity-2");
    expect(link2).toHaveAttribute("href", "/repositories/2");

    // Repository scope 는 ready → 'sub-section 예정' 배지 미노출
    expect(screen.queryByText(/sub-section 예정/i)).not.toBeInTheDocument();
  });

  it("switches to Platform scope → 'sub-section 예정' badge + entity link with /platforms/[id] href", async () => {
    const user = userEvent.setup();
    render(
      <DomainPicker
        defaultScope="repository"
        repositories={[{ id: "1", name: "org/repo-a" }]}
        platforms={[{ id: "p1", name: "Core API Platform" }]}
      />,
    );

    await user.click(screen.getByTestId("domain-picker-scope-platform"));

    // Platform scope header
    expect(screen.getByRole("heading", { name: /Platform Analytics/i })).toBeInTheDocument();

    const link = screen.getByTestId("domain-picker-entity-p1");
    expect(link).toHaveAttribute("href", "/platforms/p1");

    // 미준비 scope → 'sub-section 예정' 배지
    expect(screen.getByText(/sub-section 예정/i)).toBeInTheDocument();

    // entity list 가 Repository → Platform 으로 교체 (repo entity 안 보임)
    expect(screen.queryByTestId("domain-picker-entity-1")).not.toBeInTheDocument();
  });

  it("switches to Project scope → entity link with /projects/[id] href", async () => {
    const user = userEvent.setup();
    render(
      <DomainPicker
        defaultScope="repository"
        repositories={[{ id: "1", name: "org/repo-a" }]}
        projects={[{ id: "prj1", name: "Q2 Refactoring" }]}
      />,
    );

    await user.click(screen.getByTestId("domain-picker-scope-project"));

    expect(screen.getByRole("heading", { name: /Project Analytics/i })).toBeInTheDocument();

    const link = screen.getByTestId("domain-picker-entity-prj1");
    expect(link).toHaveAttribute("href", "/projects/prj1");
  });

  it("renders 'Loading ...' when loading=true and entities empty", () => {
    render(<DomainPicker defaultScope="repository" repositories={[]} loading />);

    expect(screen.getByText(/Loading repository entities/i)).toBeInTheDocument();
  });

  it("renders error message when error prop is set", () => {
    render(
      <DomainPicker
        defaultScope="repository"
        repositories={[]}
        error="HTTP 500 — internal"
      />,
    );

    expect(screen.getByText(/HTTP 500 — internal/)).toBeInTheDocument();
  });

  it("renders 'No ... available' when not loading + no error + empty list", () => {
    render(<DomainPicker defaultScope="repository" repositories={[]} />);

    expect(screen.getByText(/No repository entities available/i)).toBeInTheDocument();
  });

  it("does NOT call any external API — pure UI / props-driven", async () => {
    // vitest fetch spy — DomainPicker 자체는 fetch 를 호출하면 안 됨 (page 가
    // fetch 한 결과를 props 로 주입하는 정공법).
    const fetchSpy = vi.spyOn(globalThis, "fetch");
    const user = userEvent.setup();
    render(
      <DomainPicker
        defaultScope="repository"
        repositories={[{ id: "1", name: "org/repo-a" }]}
      />,
    );
    await user.click(screen.getByTestId("domain-picker-scope-platform"));

    await waitFor(() => {
      expect(screen.getByTestId("domain-picker-scope-platform")).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });
    expect(fetchSpy).not.toHaveBeenCalled();
  });
});
