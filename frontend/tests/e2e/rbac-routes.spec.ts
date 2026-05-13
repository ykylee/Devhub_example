import { test, expect, loginAs, SEEDED } from "./fixtures";

/**
 * rbac-routes.spec.ts
 * F6 권한 매트릭스(서브 경로 차단)를 검증하는 E2E 테스트.
 * 매핑 TC: TC-RBAC-SUB-01, TC-RBAC-MGR-01
 */

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
