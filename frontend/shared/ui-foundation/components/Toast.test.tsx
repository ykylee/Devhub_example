import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

// framer-motion Proxy mock.
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

import { useStore, type ToastType } from "@/lib/store";
import { ToastContainer, useToast } from "./Toast";

function resetStore() {
  act(() => {
    useStore.setState({ toasts: [] });
  });
}

describe("Toast ToastContainer (F-1)", () => {
  beforeEach(() => {
    resetStore();
  });

  it("toasts 가 비면 컨테이너만 렌더되고 항목은 없다", () => {
    const { container } = render(<ToastContainer />);
    // 외곽 컨테이너 div 1개 + 빈 AnimatePresence
    expect(container.querySelectorAll("div.glass").length).toBe(0);
  });

  it("4 종류 toast (info/success/warning/error) 가 모두 렌더된다", () => {
    const types: ToastType[] = ["info", "success", "warning", "error"];
    act(() => {
      useStore.setState({
        toasts: types.map((t, i) => ({ id: `id-${i}`, message: `msg-${t}`, type: t })),
      });
    });
    render(<ToastContainer />);
    expect(screen.getByText("msg-info")).toBeInTheDocument();
    expect(screen.getByText("msg-success")).toBeInTheDocument();
    expect(screen.getByText("msg-warning")).toBeInTheDocument();
    expect(screen.getByText("msg-error")).toBeInTheDocument();
  });

  it("close 버튼 클릭 시 store.removeToast 가 호출되어 항목이 제거된다", async () => {
    act(() => {
      useStore.setState({
        toasts: [
          { id: "a", message: "alpha", type: "info" },
          { id: "b", message: "beta", type: "success" },
        ],
      });
    });
    const user = userEvent.setup();
    render(<ToastContainer />);
    expect(screen.getByText("alpha")).toBeInTheDocument();

    // X 버튼은 각 toast 의 유일한 button — alpha 옆 button 클릭
    const alphaToast = screen.getByText("alpha").closest("div.glass") as HTMLElement;
    const closeBtn = alphaToast.querySelector("button") as HTMLButtonElement;
    await user.click(closeBtn);

    expect(useStore.getState().toasts.map((t) => t.id)).toEqual(["b"]);
  });
});

// `useToast` 는 hook 이라 component 안에서만 호출 가능 — wrapper 컴포넌트로 검증.
function ToastTrigger({ kind }: { kind: "toast" | "info" | "success" | "warning" | "error" }) {
  const t = useToast();
  return (
    <button
      onClick={() => {
        if (kind === "toast") t.toast("plain");
        else t[kind](`msg-${kind}`);
      }}
    >
      fire
    </button>
  );
}

describe("Toast useToast (F-1)", () => {
  beforeEach(() => {
    resetStore();
  });

  it("toast(message) 기본 호출 시 type=info 로 store 에 추가된다", async () => {
    const user = userEvent.setup();
    render(<ToastTrigger kind="toast" />);
    await user.click(screen.getByRole("button", { name: /fire/i }));
    const toasts = useStore.getState().toasts;
    expect(toasts).toHaveLength(1);
    expect(toasts[0].message).toBe("plain");
    expect(toasts[0].type).toBe("info");
  });

  it.each(["info", "success", "warning", "error"] as const)(
    "%s helper 호출 시 동일 type 으로 store 에 추가된다",
    async (kind) => {
      const user = userEvent.setup();
      render(<ToastTrigger kind={kind} />);
      await user.click(screen.getByRole("button", { name: /fire/i }));
      const toasts = useStore.getState().toasts;
      expect(toasts).toHaveLength(1);
      expect(toasts[0].type).toBe(kind);
      expect(toasts[0].message).toBe(`msg-${kind}`);
    },
  );
});
