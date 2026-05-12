import { test, expect, loginAs, SEEDED } from "./fixtures";

// rbac-routes.spec — F6 권한 매트릭스 sub-routes.
// TC-RBAC-SUB-01, TC-RBAC-MGR-01. 기존 auth.spec 의 단일 path gating
// (/admin/settings) 을 sub-routes 와 manager → /admin 으로 확장.
//
// Source-of-truth: pathRequiresSystemAdmin + isSystemAdmin in
// frontend/lib/auth/role-routing.ts + AuthGuard.tsx 의 redirect 로직.

test.describe("RBAC route matrix", () => {
  test("TC-RBAC-SUB-01 — developer is bounced off every /admin/settings sub-route", async ({ page }) => {
    await loginAs(page, SEEDED.developer);

    const subRoutes = [
      "/admin/settings/users",
      "/admin/settings/permissions",
      "/admin/settings/audit",
      "/admin/settings/organization",
    ];

    for (const route of subRoutes) {
      await page.goto(route);
      // pathRequiresSystemAdmin(<any admin path>) === true,
      // isSystemAdmin('developer') === false → router.replace('/developer').
      await expect(page).toHaveURL(/\/developer(\/|$)/, { timeout: 10_000 });
    }
  });

  test("TC-RBAC-MGR-01 — manager is bounced off /admin", async ({ page }) => {
    await loginAs(page, SEEDED.manager);

    await page.goto("/admin");
    // pathRequiresSystemAdmin('/admin') === true,
    // isSystemAdmin('manager') === false → /manager.
    await expect(page).toHaveURL(/\/manager(\/|$)/, { timeout: 10_000 });
  });
});
