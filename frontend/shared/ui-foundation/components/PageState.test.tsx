import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { PageLoading, PageError, PageEmpty } from "./PageState";

describe("PageLoading (F-1)", () => {
  it("기본 label='Loading...' 로 렌더된다", () => {
    render(<PageLoading />);
    expect(screen.getByText("Loading...")).toBeInTheDocument();
  });

  it("label override 가 반영된다", () => {
    render(<PageLoading label="Fetching applications..." />);
    expect(screen.getByText("Fetching applications...")).toBeInTheDocument();
  });
});

describe("PageError (F-1)", () => {
  it("message 가 렌더된다", () => {
    render(<PageError message="Network down" />);
    expect(screen.getByText("Network down")).toBeInTheDocument();
  });

  it("onRetry 미지정 시 Retry 버튼이 렌더되지 않는다", () => {
    render(<PageError message="err" />);
    expect(screen.queryByRole("button", { name: /retry/i })).toBeNull();
  });

  it("onRetry 지정 시 Retry 버튼이 렌더되고 클릭 시 호출된다", async () => {
    const onRetry = vi.fn();
    const user = userEvent.setup();
    render(<PageError message="err" onRetry={onRetry} />);
    const btn = screen.getByRole("button", { name: /retry/i });
    expect(btn).toBeInTheDocument();
    await user.click(btn);
    expect(onRetry).toHaveBeenCalledTimes(1);
  });
});

describe("PageEmpty (F-1)", () => {
  it("message 가 렌더된다", () => {
    render(<PageEmpty message="No records" />);
    expect(screen.getByText("No records")).toBeInTheDocument();
  });
});
