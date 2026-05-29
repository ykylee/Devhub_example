import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
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

import { DestructiveConfirmModal } from "./DestructiveConfirmModal";

describe("DestructiveConfirmModal (F-1)", () => {
  it("isOpen=false 면 null 을 반환한다", () => {
    const { container } = render(
      <DestructiveConfirmModal
        isOpen={false}
        onClose={vi.fn()}
        onConfirm={vi.fn()}
        title="Delete"
        description="permanent"
      />,
    );
    expect(container.firstChild).toBeNull();
  });

  it("isOpen=true 일 때 title/description/Cancel/Confirm 이 렌더된다", () => {
    render(
      <DestructiveConfirmModal
        isOpen
        onClose={vi.fn()}
        onConfirm={vi.fn()}
        title="Delete project"
        description="This will permanently delete data"
      />,
    );
    expect(screen.getByRole("heading", { name: /delete project/i })).toBeInTheDocument();
    expect(screen.getByText("This will permanently delete data")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /cancel/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /confirm/i })).toBeInTheDocument();
  });

  it("confirmText 가 지정되면 confirm 버튼 라벨이 교체된다", () => {
    render(
      <DestructiveConfirmModal
        isOpen
        onClose={vi.fn()}
        onConfirm={vi.fn()}
        title="t"
        description="d"
        confirmText="Delete forever"
      />,
    );
    expect(screen.getByRole("button", { name: /delete forever/i })).toBeInTheDocument();
  });

  it("Cancel 버튼 클릭 시 onClose 가 호출된다", async () => {
    const onClose = vi.fn();
    const user = userEvent.setup();
    render(
      <DestructiveConfirmModal
        isOpen
        onClose={onClose}
        onConfirm={vi.fn()}
        title="t"
        description="d"
      />,
    );
    await user.click(screen.getByRole("button", { name: /cancel/i }));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("Confirm 비동기 onConfirm 처리 → submitting 동안 spinner 가 노출되고 종료 후 onClose 가 호출된다", async () => {
    let resolveFn: () => void = () => {};
    const onConfirm = vi.fn(
      () =>
        new Promise<void>((resolve) => {
          resolveFn = resolve;
        }),
    );
    const onClose = vi.fn();
    const user = userEvent.setup();

    const { container } = render(
      <DestructiveConfirmModal
        isOpen
        onClose={onClose}
        onConfirm={onConfirm}
        title="t"
        description="d"
      />,
    );

    await user.click(screen.getByRole("button", { name: /confirm/i }));
    expect(onConfirm).toHaveBeenCalledTimes(1);

    // submitting=true → spinner svg.animate-spin 등장 + close (X) 버튼 hidden
    await waitFor(() => {
      expect(container.querySelector("svg.animate-spin")).not.toBeNull();
    });

    // promise resolve → finally 에서 setSubmitting(false) + onClose
    resolveFn();
    await waitFor(() => {
      expect(onClose).toHaveBeenCalledTimes(1);
    });
  });

  it("동기 onConfirm 도 호출 후 onClose 가 트리거된다", async () => {
    const onConfirm = vi.fn();
    const onClose = vi.fn();
    const user = userEvent.setup();
    render(
      <DestructiveConfirmModal
        isOpen
        onClose={onClose}
        onConfirm={onConfirm}
        title="t"
        description="d"
      />,
    );
    await user.click(screen.getByRole("button", { name: /confirm/i }));
    await waitFor(() => {
      expect(onConfirm).toHaveBeenCalledTimes(1);
      expect(onClose).toHaveBeenCalledTimes(1);
    });
  });

  it("close (X) 아이콘 버튼 클릭 시 onClose 가 호출된다", async () => {
    const onClose = vi.fn();
    const user = userEvent.setup();
    const { container } = render(
      <DestructiveConfirmModal
        isOpen
        onClose={onClose}
        onConfirm={vi.fn()}
        title="t"
        description="d"
      />,
    );
    // X 버튼은 title 옆 첫 번째 transition-colors hover:bg-muted/30 패턴.
    const closeIcon = container.querySelector("button.hover\\:bg-muted\\/30") as HTMLButtonElement;
    expect(closeIcon).not.toBeNull();
    await user.click(closeIcon);
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("backdrop 클릭 시 onClose 가 호출된다 (submitting=false)", async () => {
    const onClose = vi.fn();
    const user = userEvent.setup();
    const { container } = render(
      <DestructiveConfirmModal
        isOpen
        onClose={onClose}
        onConfirm={vi.fn()}
        title="t"
        description="d"
      />,
    );
    const backdrop = container.querySelector("div.absolute.inset-0.bg-background\\/80") as HTMLElement;
    expect(backdrop).not.toBeNull();
    await user.click(backdrop);
    expect(onClose).toHaveBeenCalledTimes(1);
  });
});
