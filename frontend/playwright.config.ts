import { defineConfig, devices } from "@playwright/test";

// Playwright config (PR-T3, work_26_05_11-d sprint).
//
// DEC-3=A: the e2e suite runs against the real Hydra/Kratos/backend/
// frontend stack on the operator's machine, not against mocks. The
// operator brings up the 5 processes (PostgreSQL + Hydra + Kratos +
// backend + frontend) following docs/setup/e2e-test-guide.md, then
// invokes `npm run e2e`.
//
// `webServer` is intentionally not set — starting all 5 processes from
// Playwright would obscure failures (which one is misbehaving?) and
// re-creates the deploy-guide's manual sequencing in less reviewable
// form. The trade-off is that operators have to remember the prereqs;
// e2e-test-guide is the canonical checklist.
export default defineConfig({
  testDir: "./tests/e2e",
  fullyParallel: false, // Hydra/Kratos sessions are global per-browser-context
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1, // single worker keeps Kratos session state predictable
  reporter: process.env.CI ? "github" : "html",
  use: {
    baseURL: process.env.PLAYWRIGHT_BASE_URL ?? "http://localhost:3000",
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
