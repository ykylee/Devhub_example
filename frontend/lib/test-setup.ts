// Vitest setup file (PR-T2, work_26_05_11-d sprint).
// Runs once per worker before any test file. Wires jest-dom matchers into
// vitest's expect, and resets the DOM between tests to keep RTL renders
// isolated.

import "@testing-library/jest-dom/vitest";
import { afterEach } from "vitest";
import { cleanup } from "@testing-library/react";

afterEach(() => {
  cleanup();
});
