import { test, expect, loginAs, SEEDED, appPath } from "./fixtures";

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
      await page.goto(appPath(route));
      // pathRequiresSystemAdmin(<any admin path>) === true,
      // isSystemAdmin('developer') === false → router.replace('/developer').
      await expect(page).toHaveURL(/\/developer(\/|$)/, { timeout: 10_000 });
    }
  });

  test("TC-RBAC-MGR-01 — manager is bounced off /admin", async ({ page }) => {
    await loginAs(page, SEEDED.team_manager);

    await page.goto(appPath("/admin"));
    // pathRequiresSystemAdmin('/admin') === true,
    // isSystemAdmin('team_manager') === false → /manager.
    await expect(page).toHaveURL(/\/manager(\/|$)/, { timeout: 10_000 });
  });

  // 6-P1: developer has platforms:view + projects:view
  test("TC-RBAC-DEV-VIEW-01 — developer can access /platforms list", async ({ page }) => {
    await loginAs(page, SEEDED.developer);
    await page.goto(appPath("/platforms"));
    // DefaultPermissionMatrix grants developer platforms:view.
    // /platforms is not under /admin so pathRequiresSystemAdmin is false.
    await expect(page).toHaveURL(/\/platforms(\/|$)/, { timeout: 10_000 });
  });

  test("TC-RBAC-DEV-VIEW-02 — developer can access /projects list", async ({ page }) => {
    await loginAs(page, SEEDED.developer);
    await page.goto(appPath("/projects"));
    // DefaultPermissionMatrix grants developer projects:view.
    await expect(page).toHaveURL(/\/projects(\/|$)/, { timeout: 10_000 });
  });

  // 6-P3: team_manager is also bounced from /admin (pathRequiresSystemAdmin)
  test("TC-RBAC-MGR-DENY-ORG-01 — team_manager is bounced from /admin/settings/organization", async ({ page }) => {
    await loginAs(page, SEEDED.team_manager);
    await page.goto(appPath("/admin/settings/organization"));
    // pathRequiresSystemAdmin('/admin/settings/organization') === true → /manager.
    await expect(page).toHaveURL(/\/manager(\/|$)/, { timeout: 10_000 });
  });

  // developer is bounced from team_manager-only endpoints
  test("TC-RBAC-DEV-DENY-01 — developer is bounced from organization settings", async ({ page }) => {
    await loginAs(page, SEEDED.developer);
    await page.goto(appPath("/admin/settings/organization"));
    // pathRequiresSystemAdmin('/admin/settings/organization') === true →
    // developer redirected to /developer.
    await expect(page).toHaveURL(/\/developer(\/|$)/, { timeout: 10_000 });
  });
});
