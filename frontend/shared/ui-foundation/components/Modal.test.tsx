import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

// framer-motion Proxy mock (BindingsTable.test.tsx 패턴).
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

import { Modal } from "./Modal";

describe("Modal (F-1)", () => {
  let onClose: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    onClose = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("isOpen=false 면 child 가 렌더되지 않는다", () => {
    render(
      <Modal isOpen={false} onClose={onClose}>
        <p data-testid="body">hidden content</p>
      </Modal>,
    );
    expect(screen.queryByTestId("body")).toBeNull();
  });

  it("isOpen=true 면 child 와 title 을 렌더한다", () => {
    render(
      <Modal isOpen onClose={onClose} title="Delete project">
        <p data-testid="body">visible content</p>
      </Modal>,
    );
    expect(screen.getByTestId("body")).toHaveTextContent("visible content");
    expect(screen.getByRole("heading", { name: /delete project/i })).toBeInTheDocument();
  });

  it("title 미지정 시 heading 이 렌더되지 않는다", () => {
    render(
      <Modal isOpen onClose={onClose}>
        <span>body</span>
      </Modal>,
    );
    expect(screen.queryByRole("heading")).toBeNull();
  });

  it("close 버튼 (X) 클릭 시 onClose 가 호출된다", async () => {
    const user = userEvent.setup();
    render(
      <Modal isOpen onClose={onClose} title="Test">
        <span>body</span>
      </Modal>,
    );
    // close 버튼은 X 아이콘 button — title 헤더 옆 단 하나의 button
    const buttons = screen.getAllByRole("button");
    await user.click(buttons[0]);
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("backdrop 클릭 시 onClose 가 호출된다", () => {
    const { container } = render(
      <Modal isOpen onClose={onClose}>
        <span>body</span>
      </Modal>,
    );
    // backdrop = 첫번째 fixed motion.div (z-[100]).
    const backdrop = container.querySelector("div.fixed.inset-0.z-\\[100\\]");
    expect(backdrop).not.toBeNull();
    fireEvent.click(backdrop as HTMLElement);
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("Escape 키 입력 시 onClose 가 호출된다", () => {
    render(
      <Modal isOpen onClose={onClose}>
        <span>body</span>
      </Modal>,
    );
    fireEvent.keyDown(window, { key: "Escape" });
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("Escape 외의 키는 onClose 를 호출하지 않는다", () => {
    render(
      <Modal isOpen onClose={onClose}>
        <span>body</span>
      </Modal>,
    );
    fireEvent.keyDown(window, { key: "Enter" });
    fireEvent.keyDown(window, { key: "a" });
    expect(onClose).not.toHaveBeenCalled();
  });

  it("size prop 별 sizeClasses 가 반영된다 (sm/md/lg/xl/full)", () => {
    const sizes: Array<{ size: "sm" | "md" | "lg" | "xl" | "full"; cls: string }> = [
      { size: "sm", cls: "max-w-md" },
      { size: "md", cls: "max-w-lg" },
      { size: "lg", cls: "max-w-2xl" },
      { size: "xl", cls: "max-w-4xl" },
      { size: "full", cls: "max-w-[95vw]" },
    ];
    for (const { size, cls } of sizes) {
      const { container, unmount } = render(
        <Modal isOpen onClose={onClose} size={size}>
          <span>body</span>
        </Modal>,
      );
      // glass-card 가 size class 를 받는 motion.div
      const card = container.querySelector(`.${CSS.escape(cls)}`);
      expect(card).not.toBeNull();
      unmount();
    }
  });

  it("className 이 컨텐트 컨테이너에 병합된다", () => {
    const { container } = render(
      <Modal isOpen onClose={onClose} className="custom-modal-class">
        <span>body</span>
      </Modal>,
    );
    expect(container.querySelector(".custom-modal-class")).not.toBeNull();
  });

  it("언마운트 시 keydown listener 가 제거된다", () => {
    const remove = vi.spyOn(window, "removeEventListener");
    const { unmount } = render(
      <Modal isOpen onClose={onClose}>
        <span>body</span>
      </Modal>,
    );
    unmount();
    expect(remove).toHaveBeenCalledWith("keydown", expect.any(Function));
  });
});
