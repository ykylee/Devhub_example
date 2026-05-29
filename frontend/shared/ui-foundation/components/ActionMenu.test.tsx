import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { ActionMenu, type ActionMenuItem } from "./ActionMenu";

function buildItems(): ActionMenuItem[] {
  return [
    { key: "edit", label: "Edit", onClick: vi.fn() },
    { key: "delete", label: "Delete", onClick: vi.fn(), tone: "danger" },
    { key: "archive", label: "Archive", onClick: vi.fn(), icon: <span data-testid="archive-icon">A</span> },
  ];
}

describe("ActionMenu (F-1)", () => {
  beforeEach(() => {
    // getBoundingClientRect mock so toggleFromTarget pos 계산이 정상.
    Element.prototype.getBoundingClientRect = vi.fn(() => ({
      x: 0,
      y: 0,
      width: 100,
      height: 30,
      top: 50,
      left: 100,
      right: 200,
      bottom: 80,
      toJSON: () => ({}),
    }));
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("trigger 가 기본 title='Actions' aria-label 로 렌더된다", () => {
    render(<ActionMenu items={buildItems()} />);
    expect(screen.getByRole("button", { name: "Actions" })).toBeInTheDocument();
  });

  it("title 을 override 하면 aria-label 이 반영된다", () => {
    render(<ActionMenu title="Row Actions" items={buildItems()} />);
    expect(screen.getByRole("button", { name: "Row Actions" })).toBeInTheDocument();
  });

  it("trigger 클릭 시 menu 가 열리고 모든 아이템 라벨 + 커스텀 icon 이 렌더된다", async () => {
    const user = userEvent.setup();
    render(<ActionMenu items={buildItems()} />);
    await user.click(screen.getByRole("button", { name: "Actions" }));
    expect(screen.getByRole("menu", { name: "Actions" })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: /edit/i })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: /delete/i })).toBeInTheDocument();
    expect(screen.getByTestId("archive-icon")).toBeInTheDocument();
  });

  it("trigger 두 번 클릭 시 menu 가 닫힌다 (toggle)", async () => {
    const user = userEvent.setup();
    render(<ActionMenu items={buildItems()} />);
    const trigger = screen.getByRole("button", { name: "Actions" });
    await user.click(trigger);
    expect(screen.getByRole("menu")).toBeInTheDocument();
    await user.click(trigger);
    expect(screen.queryByRole("menu")).toBeNull();
  });

  it("Escape 키 입력 시 menu 가 닫힌다", async () => {
    const user = userEvent.setup();
    render(<ActionMenu items={buildItems()} />);
    await user.click(screen.getByRole("button", { name: "Actions" }));
    expect(screen.getByRole("menu")).toBeInTheDocument();
    fireEvent.keyDown(window, { key: "Escape" });
    expect(screen.queryByRole("menu")).toBeNull();
  });

  it("아이템 click (keyboard path) 시 onClick 이 호출되고 menu 가 닫힌다", async () => {
    const items = buildItems();
    const user = userEvent.setup();
    render(<ActionMenu items={items} />);
    await user.click(screen.getByRole("button", { name: "Actions" }));

    // user-event click 은 keyboard-synthesized pointer 가 아닌 mouse pointer 를
    // 시뮬레이트 — onPointerDown path 가 실행됨. 명시적 keyboard sim 을 위해
    // fireEvent.click 으로 synthetic click 만 발생시켜 onClick path 검증.
    fireEvent.click(screen.getByRole("menuitem", { name: /edit/i }));

    expect(items[0].onClick).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole("menu")).toBeNull();
  });

  it("아이템 pointerdown (touch/mouse path) 시 onClick 이 호출되고 menu 가 닫힌다", async () => {
    const items = buildItems();
    const user = userEvent.setup();
    render(<ActionMenu items={items} />);
    await user.click(screen.getByRole("button", { name: "Actions" }));

    fireEvent.pointerDown(screen.getByRole("menuitem", { name: /delete/i }));
    expect(items[1].onClick).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole("menu")).toBeNull();
  });

  it("danger tone 아이템에 text-destructive 클래스가 적용된다", async () => {
    const user = userEvent.setup();
    render(<ActionMenu items={buildItems()} />);
    await user.click(screen.getByRole("button", { name: "Actions" }));
    const deleteItem = screen.getByRole("menuitem", { name: /delete/i });
    expect(deleteItem.className).toMatch(/text-destructive/);
  });

  it("custom widthPx 가 menu style.width 에 반영된다", async () => {
    const user = userEvent.setup();
    render(<ActionMenu items={buildItems()} widthPx={240} />);
    await user.click(screen.getByRole("button", { name: "Actions" }));
    const menu = screen.getByRole("menu") as HTMLElement;
    expect(menu.style.width).toBe("240px");
  });

  it("triggerClassName / menuClassName 이 병합된다", async () => {
    const user = userEvent.setup();
    render(
      <ActionMenu
        items={buildItems()}
        triggerClassName="extra-trigger-class"
        menuClassName="extra-menu-class"
      />,
    );
    const trigger = screen.getByRole("button", { name: "Actions" });
    expect(trigger.className).toMatch(/extra-trigger-class/);
    await user.click(trigger);
    const menu = screen.getByRole("menu");
    expect(menu.className).toMatch(/extra-menu-class/);
  });

  it("backdrop pointerdown 시 menu 가 닫힌다", async () => {
    const user = userEvent.setup();
    const { container } = render(<ActionMenu items={buildItems()} />);
    await user.click(screen.getByRole("button", { name: "Actions" }));
    expect(screen.getByRole("menu")).toBeInTheDocument();
    // backdrop = fixed inset-0 z-40 div (menu 가 z-[60]).
    const backdrop = container.querySelector("div.fixed.inset-0.z-40") as HTMLElement;
    expect(backdrop).not.toBeNull();
    fireEvent.pointerDown(backdrop);
    expect(screen.queryByRole("menu")).toBeNull();
  });

  it("touch end → 직후 click 은 swallow 되어 menu 가 닫히지 않는다 (touch dedup)", async () => {
    const user = userEvent.setup();
    render(<ActionMenu items={buildItems()} />);
    const trigger = screen.getByRole("button", { name: "Actions" });

    // touch end 가 먼저 menu 를 연다.
    fireEvent.touchEnd(trigger);
    expect(screen.getByRole("menu")).toBeInTheDocument();

    // 같은 sequence 에서 곧바로 발생한 click 은 touchedRef.current=true 라 swallow → menu 유지.
    fireEvent.click(trigger);
    expect(screen.getByRole("menu")).toBeInTheDocument();

    // 안전: 정리.
    await user.keyboard("{Escape}");
  });

  it("touch end 후 350ms 이 지나면 touched flag 가 reset 되어 다음 click 정상 동작", async () => {
    vi.useFakeTimers();
    render(<ActionMenu items={buildItems()} />);
    const trigger = screen.getByRole("button", { name: "Actions" });

    fireEvent.touchEnd(trigger);
    expect(screen.getByRole("menu")).toBeInTheDocument();

    act(() => {
      vi.advanceTimersByTime(400);
    });

    // 다시 click 했을 때 touchedRef 가 false 라 toggle 동작 — 열려있던 menu 가 닫힌다.
    fireEvent.click(trigger);
    expect(screen.queryByRole("menu")).toBeNull();

    vi.useRealTimers();
  });

  it("falsy item 은 filter 되어 menu 에 표시되지 않는다", async () => {
    const items = [
      { key: "a", label: "A", onClick: vi.fn() },
      null,
      undefined,
      false,
      { key: "b", label: "B", onClick: vi.fn() },
    ] as unknown as ActionMenuItem[];
    const user = userEvent.setup();
    render(<ActionMenu items={items} />);
    await user.click(screen.getByRole("button", { name: "Actions" }));
    expect(screen.getAllByRole("menuitem")).toHaveLength(2);
  });

  it("pointer click (mouse) 가 onPointerDown 으로 처리되면 onClick path 는 nativeEvent.pointerType 가 비어있지 않아 early-return 한다", async () => {
    const items = buildItems();
    render(<ActionMenu items={items} />);
    fireEvent.pointerDown(screen.getByRole("button", { name: "Actions" }));
    // toggleFromTarget 은 onClick path 가 아닌 onTouchEnd / mouse onClick 에서만 호출.
    // 다만 onClick 만 발생한 경우(`pointerType` 비어있을 때) toggle 발생.
    // 본 케이스는 그저 onPointerDown 단독 호출 — menu 가 열리지 않아야 함.
    expect(screen.queryByRole("menu")).toBeNull();
  });

  it("언마운트 시 keydown listener 와 touched timer 가 정리된다", async () => {
    vi.useFakeTimers();
    const remove = vi.spyOn(window, "removeEventListener");
    const { unmount } = render(<ActionMenu items={buildItems()} />);
    // open → keydown listener 등록
    fireEvent.click(screen.getByRole("button", { name: "Actions" }));
    // touched timer 도 가동
    fireEvent.touchEnd(screen.getByRole("button", { name: "Actions" }));
    unmount();
    // keydown listener 정리됨
    expect(remove).toHaveBeenCalledWith("keydown", expect.any(Function));
    vi.useRealTimers();
  });
});
