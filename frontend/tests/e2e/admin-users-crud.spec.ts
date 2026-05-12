import { test, expect, loginAs, SEEDED } from "./fixtures";

/**
 * admin-users-crud.spec.ts
 * F-USR-CRUD 를 검증하는 E2E UI smoke 테스트.
 * 매핑 TC: TC-USR-CRUD-01, TC-USR-CRUD-02, TC-USR-CRUD-03
 */

test.describe("/admin/settings/users — CRUD UI smoke", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto("/admin/settings/users");
    
    // Wait for the page to stabilize
    await page.waitForLoadState("networkidle");
    
    // Wait for loading to finish
    await expect(page.getByText(/loading users/i)).toBeHidden({ timeout: 20_000 });
    
    // Ensure Alice is visible as a smoke check for data loading
    await expect(page.getByRole("row").filter({ hasText: "alice" })).toBeVisible({ timeout: 15_000 });
  });

  test("TC-USR-CRUD-01 — Invite Member 모달 열림 + 닫기", async ({ page }) => {
    await page.getByRole("button", { name: /invite member/i }).click();
    
    // Probe a few common modal anchors; pick the first that resolves.
    const modal = page.getByRole("dialog");
    await expect(modal).toBeVisible({ timeout: 5_000 });

    // Close via the ESC key — Modal component handles onClose. Avoid
    // form submit so no backend mutation fires.
    await page.keyboard.press("Escape");
    await expect(modal).toBeHidden();
  });

  test("TC-USR-CRUD-02 — Role select dropdown 옵션 노출", async ({ page }) => {
    // alice 행의 role select box 확인
    const aliceRow = page.getByRole("row").filter({ hasText: "alice" });
    const roleSelect = aliceRow.getByRole("combobox");
    
    await expect(roleSelect).toBeVisible();
    
    // Check if the options exist within the select (use toBeAttached for native select options)
    await expect(roleSelect.locator("option", { hasText: "System Admin" })).toBeAttached();
    await expect(roleSelect.locator("option", { hasText: "Developer" })).toBeAttached();
    await expect(roleSelect.locator("option", { hasText: "Manager" })).toBeAttached();
  });

  test("TC-USR-CRUD-03 — Action 메뉴 (...) 에 system_admin 액션 3종 노출", async ({ page }) => {
    const aliceRow = page.getByRole("row").filter({ hasText: "alice" });
    
    // (...) 버튼 클릭
    await aliceRow.getByRole("button").filter({ has: page.locator("svg") }).click();
    
    // 3가지 액션 버튼 노출 확인
    await expect(page.getByRole("button", { name: /issue account/i })).toBeVisible();
    await expect(page.getByRole("button", { name: /force reset password/i })).toBeVisible();
    await expect(page.getByRole("button", { name: /revoke account/i })).toBeVisible();
    
    // Close the menu
    await page.keyboard.press("Escape");
    await expect(page.getByRole("button", { name: /issue account/i })).toBeHidden();
  });
});
