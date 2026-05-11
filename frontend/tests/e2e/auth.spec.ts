import { test, expect, loginAs, SEEDED } from "./fixtures";

// auth.spec — login + role-based landing (PR-S1) + system route gating.
// Source-of-truth: defaultLandingFor + pathRequiresSystemAdmin in
// frontend/lib/auth/role-routing.ts.

test.describe("role-based landing", () => {
  test("developer lands on /developer", async ({ page }) => {
    await loginAs(page, SEEDED.developer);
    await expect(page).toHaveURL(/\/developer(\/|$)/);
  });

  test("manager lands on /manager", async ({ page }) => {
    await loginAs(page, SEEDED.manager);
    await expect(page).toHaveURL(/\/manager(\/|$)/);
  });

  test("system_admin lands on /admin", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await expect(page).toHaveURL(/\/admin(\/|$)/);
  });
});

test.describe("system route gating", () => {
  test("developer cannot reach /admin/settings — AuthGuard bounces to default landing", async ({ page }) => {
    await loginAs(page, SEEDED.developer);
    await page.goto("/admin/settings");
    // pathRequiresSystemAdmin + isSystemAdmin guard in AuthGuard.tsx
    // redirects to defaultLandingFor(actor.role) = /developer.
    await expect(page).toHaveURL(/\/developer(\/|$)/, { timeout: 10_000 });
  });
});
