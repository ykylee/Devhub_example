import React from "react";
// Vitest setup file (PR-T2, work_26_05_11-d sprint).
// Runs once per worker before any test file. Wires jest-dom matchers into
// vitest's expect, and resets the DOM between tests to keep RTL renders
// isolated.
//
// postinstall.js patches @testing-library/react/dist/act-compat.js for React 19.

import "@testing-library/jest-dom/vitest";
import { afterEach, vi } from "vitest";
import { cleanup } from "@testing-library/react";


afterEach(() => {
  cleanup();
});

// --- Happy-DOM localStorage Mock --------------------------------------
class LocalStorageMock implements Storage {
  private store: Record<string, string> = {};

  clear() {
    this.store = {};
  }

  getItem(key: string): string | null {
    return this.store[key] || null;
  }

  setItem(key: string, value: string): void {
    this.store[key] = String(value);
  }

  removeItem(key: string): void {
    delete this.store[key];
  }

  get length(): number {
    return Object.keys(this.store).length;
  }

  key(index: number): string | null {
    const keys = Object.keys(this.store);
    return keys[index] || null;
  }
}

const localStorageMock = new LocalStorageMock();

Object.defineProperty(window, "localStorage", {
  value: localStorageMock,
  writable: true,
});

// --- next/link + next/navigation mock (2026-06-17 issue 3 detail page test 정합) ---
// KpiTestDetailPage wrapper + 4 component (Platform/Project/Repository × KPI/Tests) 가
// \`next/link\` 의 <Link> + \`next/navigation\` 의 usePathname 사용. vitest runtime 에서
// next.js router 가 없으므로 <a> + location 모킹으로 정합. 실제 navigation 동작은
// production runtime 에서 next.js 가 처리 (client-side prefetch + push).
vi.mock("next/link", () => ({
  default: ({ children, href, ...rest }: { children: React.ReactNode; href: string; [k: string]: unknown }) =>
    React.createElement("a", { href, ...rest }, children),
}));
vi.mock("next/navigation", () => ({
  usePathname: () => "/",
  useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
  useParams: () => ({}),
  useSearchParams: () => new URLSearchParams(),
}));
