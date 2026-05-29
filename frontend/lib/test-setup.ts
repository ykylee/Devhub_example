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
