import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
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

import { ComboBox } from "./ComboBox";

const sampleOptions = [
  { label: "Alice", value: "alice", description: "frontend lead" },
  { label: "Bob", value: "bob", description: "backend lead" },
  { label: "Carol", value: "carol" },
];

describe("ComboBox (F-1)", () => {
  it("기본 placeholder 가 노출되고 trigger 가 닫힌 상태로 렌더된다", () => {
    render(<ComboBox options={sampleOptions} value="" onChange={vi.fn()} />);
    expect(screen.getByText("Select option...")).toBeInTheDocument();
    const trigger = screen.getByRole("button");
    expect(trigger).toHaveAttribute("aria-expanded", "false");
  });

  it("placeholder override 가 반영된다", () => {
    render(
      <ComboBox
        options={sampleOptions}
        value=""
        onChange={vi.fn()}
        placeholder="Pick someone"
      />,
    );
    expect(screen.getByText("Pick someone")).toBeInTheDocument();
  });

  it("value 가 옵션과 일치하면 selectedOption.label 이 노출된다", () => {
    render(<ComboBox options={sampleOptions} value="bob" onChange={vi.fn()} />);
    expect(screen.getByText("Bob")).toBeInTheDocument();
  });

  it("trigger 클릭 시 listbox 가 열리고 모든 옵션이 렌더된다", async () => {
    const user = userEvent.setup();
    render(<ComboBox options={sampleOptions} value="" onChange={vi.fn()} />);
    await user.click(screen.getByRole("button"));
    expect(screen.getByRole("button")).toHaveAttribute("aria-expanded", "true");
    expect(screen.getAllByRole("option")).toHaveLength(3);
  });

  it("옵션 클릭 시 onChange 가 호출되고 listbox 가 닫힌다", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();
    render(<ComboBox options={sampleOptions} value="" onChange={onChange} />);
    await user.click(screen.getByRole("button"));
    await user.click(screen.getByRole("option", { name: /alice/i }));
    expect(onChange).toHaveBeenCalledWith("alice");
    expect(screen.getByRole("button")).toHaveAttribute("aria-expanded", "false");
  });

  it("검색 input 에 입력 시 label/value/description 매칭 결과만 남는다", async () => {
    const user = userEvent.setup();
    render(<ComboBox options={sampleOptions} value="" onChange={vi.fn()} />);
    await user.click(screen.getByRole("button"));
    const searchInput = screen.getByPlaceholderText("Search...") as HTMLInputElement;
    await user.type(searchInput, "back");
    // description 'backend lead' 만 매치 → Bob 만 남아야 함.
    expect(screen.getByRole("option", { name: /bob/i })).toBeInTheDocument();
    expect(screen.queryByRole("option", { name: /alice/i })).toBeNull();
    expect(screen.queryByRole("option", { name: /carol/i })).toBeNull();
  });

  it("검색 결과가 없으면 emptyText 가 노출된다", async () => {
    const user = userEvent.setup();
    render(
      <ComboBox
        options={sampleOptions}
        value=""
        onChange={vi.fn()}
        emptyText="아무도 없음"
      />,
    );
    await user.click(screen.getByRole("button"));
    await user.type(screen.getByPlaceholderText("Search..."), "zzz");
    expect(screen.getByText("아무도 없음")).toBeInTheDocument();
  });

  it("검색 후 Enter 키 입력 시 첫 번째 결과가 선택된다", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();
    render(<ComboBox options={sampleOptions} value="" onChange={onChange} />);
    await user.click(screen.getByRole("button"));
    const searchInput = screen.getByPlaceholderText("Search...") as HTMLInputElement;
    await user.type(searchInput, "car"); // matches Carol
    await user.keyboard("{Enter}");
    expect(onChange).toHaveBeenCalledWith("carol");
  });

  it("검색 결과가 비어있을 때 Enter 는 무동작", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();
    render(<ComboBox options={sampleOptions} value="" onChange={onChange} />);
    await user.click(screen.getByRole("button"));
    const searchInput = screen.getByPlaceholderText("Search...") as HTMLInputElement;
    await user.type(searchInput, "qq");
    await user.keyboard("{Enter}");
    expect(onChange).not.toHaveBeenCalled();
  });

  it("검색 input 의 Escape 키 입력 시 listbox 가 닫힌다", async () => {
    const user = userEvent.setup();
    render(<ComboBox options={sampleOptions} value="" onChange={vi.fn()} />);
    await user.click(screen.getByRole("button"));
    const searchInput = screen.getByPlaceholderText("Search...");
    fireEvent.keyDown(searchInput, { key: "Escape" });
    expect(screen.getByRole("button")).toHaveAttribute("aria-expanded", "false");
  });

  it("search clear 버튼 클릭 시 검색어 state 가 비워진다 (filtering 결과로 검증)", async () => {
    // happy-dom 의 controlled input 은 React state 갱신 후 dom .value 가 갱신되지 않는
    // 버그가 있어, 실제 검색 효과 (filteredOptions 의 렌더 결과) 와 clear 버튼의 출현/소멸
    // 을 통해 search state 흐름을 검증한다.
    const user = userEvent.setup();
    render(<ComboBox options={sampleOptions} value="" onChange={vi.fn()} />);
    await user.click(screen.getByRole("button"));
    const searchInput = screen.getByPlaceholderText("Search...");
    await user.type(searchInput, "back");
    // 'back' → Bob 만 매치.
    await waitFor(() => {
      expect(screen.queryByRole("option", { name: /alice/i })).toBeNull();
    });
    const clearBtn = screen.getByRole("button", { name: "Clear search" });
    await user.click(clearBtn);
    // clear 후 검색 state 가 "" 로 → 모든 옵션 복귀.
    await waitFor(() => expect(screen.getAllByRole("option")).toHaveLength(3));
  });

  it("disabled=true 면 trigger 가 disabled 이고 클릭해도 열리지 않는다", async () => {
    const user = userEvent.setup();
    render(<ComboBox options={sampleOptions} value="" onChange={vi.fn()} disabled />);
    const trigger = screen.getByRole("button");
    expect(trigger).toBeDisabled();
    await user.click(trigger);
    expect(trigger).toHaveAttribute("aria-expanded", "false");
  });

  it("외부 클릭 시 listbox 가 닫힌다", async () => {
    const user = userEvent.setup();
    render(
      <div>
        <ComboBox options={sampleOptions} value="" onChange={vi.fn()} />
        <button data-testid="outside">outside</button>
      </div>,
    );
    await user.click(screen.getByRole("button", { name: /select option/i }));
    expect(screen.getAllByRole("option").length).toBeGreaterThan(0);
    // mousedown 외부 → handleClickOutside.
    fireEvent.mouseDown(screen.getByTestId("outside"));
    expect(screen.queryByRole("option")).toBeNull();
  });

  it("className 이 컨테이너에 병합된다", () => {
    const { container } = render(
      <ComboBox options={sampleOptions} value="" onChange={vi.fn()} className="combo-extra" />,
    );
    expect(container.querySelector(".combo-extra")).not.toBeNull();
  });

  it("선택된 옵션에는 Check 아이콘이 렌더된다", async () => {
    const user = userEvent.setup();
    const { container } = render(
      <ComboBox options={sampleOptions} value="alice" onChange={vi.fn()} />,
    );
    await user.click(screen.getByRole("button", { name: /alice/i }));
    // alice 의 aria-selected="true" 옵션 안에 Check svg 가 있어야 함.
    const aliceOption = screen.getByRole("option", { selected: true });
    expect(aliceOption.querySelector("svg")).not.toBeNull();
    expect(container).toBeDefined();
  });

  it("opening 시 search 가 reset 된다 (filtering 결과로 검증)", async () => {
    // 위와 동일 — happy-dom 의 controlled input dom value 한계로 filtering 효과로 검증.
    const user = userEvent.setup();
    render(<ComboBox options={sampleOptions} value="" onChange={vi.fn()} />);
    const trigger = screen.getByRole("button");
    await user.click(trigger);
    const searchInput = screen.getByPlaceholderText("Search...");
    await user.type(searchInput, "back");
    await waitFor(() => {
      expect(screen.queryByRole("option", { name: /alice/i })).toBeNull();
    });
    // 닫고 다시 열기.
    await user.click(trigger);
    await user.click(trigger);
    // 재오픈 후 search 가 reset 되어 모든 옵션 복귀.
    await waitFor(() => expect(screen.getAllByRole("option")).toHaveLength(3));
  });
});
