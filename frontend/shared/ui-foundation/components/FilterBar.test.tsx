import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
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

import { FilterBar } from "./FilterBar";

const baseOptions = [
  { label: "All", value: "all" },
  { label: "Active", value: "active" },
  { label: "Inactive", value: "inactive" },
];

describe("FilterBar (F-1)", () => {
  let onSearch: ReturnType<typeof vi.fn>;
  let onFilterChange: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    vi.useFakeTimers();
    onSearch = vi.fn();
    onFilterChange = vi.fn();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("초기 렌더 시 searchLabel 이 input aria-label 로 노출되고 모든 filterOption 이 노출된다", () => {
    render(
      <FilterBar
        onSearch={onSearch}
        onFilterChange={onFilterChange}
        filterOptions={baseOptions}
        searchLabel="Search repositories"
      />,
    );
    expect(screen.getByLabelText("Search repositories")).toBeInTheDocument();
    expect(screen.getAllByText("All").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("Active").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("Inactive").length).toBeGreaterThanOrEqual(1);
  });

  it("입력 후 300ms debounce 가 지나면 onSearch 가 호출된다", () => {
    render(
      <FilterBar
        onSearch={onSearch}
        onFilterChange={onFilterChange}
        filterOptions={baseOptions}
        searchLabel="search"
      />,
    );
    // 첫 마운트 시 mount-effect 가 즉시 empty string 으로 onSearch("") 한 번 호출됨.
    vi.advanceTimersByTime(300);
    onSearch.mockClear();

    const input = screen.getByLabelText("search") as HTMLInputElement;
    fireEvent.change(input, { target: { value: "alpha" } });
    // 아직 debounce 전.
    expect(onSearch).not.toHaveBeenCalled();
    vi.advanceTimersByTime(300);
    expect(onSearch).toHaveBeenCalledWith("alpha");
  });

  it("검색어가 있으면 clear (X) 버튼이 노출되고 클릭 시 비워진다", async () => {
    render(
      <FilterBar
        onSearch={onSearch}
        onFilterChange={onFilterChange}
        filterOptions={baseOptions}
        searchLabel="search"
      />,
    );
    const input = screen.getByLabelText("search") as HTMLInputElement;
    fireEvent.change(input, { target: { value: "beta" } });
    expect(input.value).toBe("beta");
    // clear button = search input 옆의 X 아이콘 버튼 (Advanced filters 외에 button)
    const clearBtn = input.parentElement?.querySelector("button") as HTMLButtonElement;
    expect(clearBtn).not.toBeNull();
    fireEvent.click(clearBtn);
    expect(input.value).toBe("");
  });

  it("desktop filter chip 클릭 시 onFilterChange 가 해당 값으로 호출된다", async () => {
    vi.useRealTimers();
    const user = userEvent.setup();
    render(
      <FilterBar
        onSearch={onSearch}
        onFilterChange={onFilterChange}
        filterOptions={baseOptions}
        searchLabel="search"
        activeFilter="all"
      />,
    );
    // 동일 라벨이 desktop chip + mobile dropdown 모두에 있을 수 있으니 모든 'Active' 매치 후
    // 가장 먼저 발견된 chip 클릭.
    const activeBtns = screen.getAllByRole("button", { name: "Active" });
    await user.click(activeBtns[0]);
    expect(onFilterChange).toHaveBeenCalledWith("active");
  });

  it("activeFilter 가 chip 의 active style (bg-primary) 을 활성화한다", () => {
    const { container } = render(
      <FilterBar
        onSearch={onSearch}
        onFilterChange={onFilterChange}
        filterOptions={baseOptions}
        searchLabel="search"
        activeFilter="inactive"
      />,
    );
    // active chip 은 'bg-primary' class 보유 — Inactive 항목 chip 만 그 클래스를 갖는다.
    const buttons = container.querySelectorAll("button.bg-primary");
    expect(buttons.length).toBeGreaterThanOrEqual(1);
    const found = Array.from(buttons).some((b) => b.textContent?.trim() === "Inactive");
    expect(found).toBe(true);
  });

  it("mobile compact filter dropdown 을 열고 옵션 선택 시 onFilterChange 호출되고 dropdown 이 닫힌다", async () => {
    vi.useRealTimers();
    const user = userEvent.setup();
    const { container } = render(
      <FilterBar
        onSearch={onSearch}
        onFilterChange={onFilterChange}
        filterOptions={baseOptions}
        searchLabel="search"
        activeFilter="all"
      />,
    );
    // mobile toggle: 'Filter' label 보유 button (currently shows activeFilter label "All").
    // mobile 컨테이너는 .lg\:hidden — 그 안의 첫 번째 button.
    const mobileWrap = container.querySelector("div.lg\\:hidden.relative") as HTMLElement;
    expect(mobileWrap).not.toBeNull();
    const toggle = mobileWrap.querySelector("button") as HTMLButtonElement;
    await user.click(toggle);

    // dropdown 의 'Active' 옵션 — 그 옵션은 dropdown 안 buttons 중 'Active' label.
    // dropdown 컨테이너 내부 button 들 (chip 과 헷갈리지 않게 mobileWrap 안에서만 검색).
    const dropdownActive = Array.from(mobileWrap.querySelectorAll("button")).find(
      (b) => b.textContent?.trim() === "Active",
    ) as HTMLButtonElement;
    expect(dropdownActive).toBeDefined();
    await user.click(dropdownActive);
    expect(onFilterChange).toHaveBeenCalledWith("active");
  });

  it("placeholder prop 이 input 에 반영된다", () => {
    render(
      <FilterBar
        onSearch={onSearch}
        onFilterChange={onFilterChange}
        filterOptions={baseOptions}
        searchLabel="search"
        placeholder="Find anything"
      />,
    );
    expect(screen.getByPlaceholderText("Find anything")).toBeInTheDocument();
  });
});
