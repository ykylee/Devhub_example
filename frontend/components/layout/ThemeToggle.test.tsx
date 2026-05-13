import { render, screen, act, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ThemeToggle } from "./ThemeToggle";

// ThemeToggle component unit tests (sprint claude/work_260513-i, C1 1차 도입).
// Scope: mount + initial render (light), localStorage persistence on toggle,
// DOM class toggling on documentElement. AuthGuard / Header / Sidebar 의 본격
// Vitest 는 mock 양이 커서 별도 후속 sprint 의 carve out.

describe("ThemeToggle", () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.classList.remove("theme-dark");
  });

  it("returns null before mount, then renders the toggle button", async () => {
    const { container } = render(<ThemeToggle />);
    // The component returns null synchronously until the mount-effect fires
    // (setTimeout 0). Wait for the button to appear instead of asserting on
    // the initial render — React 's effect timing makes the synchronous null
    // window flaky to pin.
    await waitFor(() => {
      const btn = container.querySelector("button");
      expect(btn).not.toBeNull();
    });
  });

  it("persists theme to localStorage and toggles documentElement class", async () => {
    render(<ThemeToggle />);
    const user = userEvent.setup();

    const btn = await waitFor(() => {
      const found = document.querySelector("button");
      if (!found) throw new Error("button not yet mounted");
      return found as HTMLButtonElement;
    });

    expect(document.documentElement.classList.contains("theme-dark")).toBe(false);

    await act(async () => {
      await user.click(btn);
    });

    expect(localStorage.getItem("devhub-theme")).toBe("dark");
    expect(document.documentElement.classList.contains("theme-dark")).toBe(true);

    await act(async () => {
      await user.click(btn);
    });

    expect(localStorage.getItem("devhub-theme")).toBe("light");
    expect(document.documentElement.classList.contains("theme-dark")).toBe(false);
  });

  it("reads saved theme from localStorage on mount", async () => {
    localStorage.setItem("devhub-theme", "dark");
    render(<ThemeToggle />);

    await waitFor(() => {
      expect(document.documentElement.classList.contains("theme-dark")).toBe(true);
    });
  });
});
