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

    const matrixTable = page.locator("table").filter({ has: page.getByRole("columnheader", { name: "View" }) }).first();
    const resourceLabels = [
      /Infrastructure & Topology/i,
      /CI\/CD Pipelines/i,
      /Organization & Members/i,
      /Risk & Security/i,
      /Audit Logs & History/i,
      /Applications/i,
      /Application Repositories/i,
      /Projects/i,
      /SCM Providers/i,
    ];
    for (const label of resourceLabels) {
      await expect(matrixTable.getByRole("cell", { name: label })).toBeVisible();
    }

    for (const action of ["View", "Create", "Edit", "Delete"]) {
      await expect(page.getByRole("columnheader", { name: action })).toBeVisible();
    }
  });

  test("TC-PERMISSIONS-EDIT-01 — Custom role 생성 + 권한 편집 + 저장", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto("/admin/settings/permissions");

    // 1. Create Role
    await page.getByRole("button", { name: /create role/i }).click();
    await expect(page.getByRole("heading", { name: /new custom role matrix/i })).toBeVisible();

    // 2. Edit Permissions (e.g. grant view to Applications)
    // Find the row for Applications and click the View button (the first one)
    const matrixTable = page.locator("table").filter({ has: page.getByRole("columnheader", { name: "View" }) }).first();
    const appRow = matrixTable.locator("tr").filter({ has: page.getByRole("cell", { name: /^Applications$/i }) });
    const viewBtn = appRow.locator("button").first();
    
    // Check if it's currently not granted (has X icon or specific class)
    // The component uses <X className="w-4 h-4" /> when not granted.
    await viewBtn.click();
    
    // 3. Save
    const saveBtn = page.getByRole("button", { name: /save permissions/i });
    await expect(saveBtn).toBeEnabled();
    await saveBtn.click();

    // 4. Verify Success
    // "Saving…" text may be too brief to observe reliably in CI/local.
    // Stable post-condition: dirty flag clears and save button is disabled.
    await expect(saveBtn).toBeDisabled({ timeout: 10_000 });
  });
});
