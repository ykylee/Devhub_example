import { test, expect, loginAs, SEEDED } from "./fixtures";

/**
 * admin-permissions.spec.ts
 * F8 권한 편집을 검증하는 E2E 테스트.
 * 매핑 TC: TC-PERMISSIONS-SMOKE-01
 */

test.describe("/admin/settings/permissions — PermissionEditor smoke", () => {
  test("TC-PERMISSIONS-SMOKE-01 — 진입 + role 선택 + matrix 노출", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto("/admin/settings/permissions");

    const devCard = page.getByRole("heading", { name: /^developer$/i });
    await expect(devCard).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole("heading", { name: /^manager$/i })).toBeVisible();
    await expect(page.getByRole("heading", { name: /^system admin$/i })).toBeVisible();

    await devCard.click();

    await expect(page.getByRole("heading", { name: /developer matrix/i })).toBeVisible({ timeout: 5_000 });

    const resourceLabels = [
      /Infrastructure & Topology/i,
      /CI\/CD Pipelines/i,
      /Organization & Members/i,
      /Risk & Security/i,
      /Audit Logs & History/i,
    ];
    for (const label of resourceLabels) {
      await expect(page.getByText(label)).toBeVisible();
    }

    for (const action of ["View", "Create", "Edit", "Delete"]) {
      await expect(page.getByRole("columnheader", { name: action })).toBeVisible();
    }
  });
});
