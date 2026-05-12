import { test, expect, loginAs, SEEDED } from "./fixtures";

/**
 * admin-org-crud.spec.ts
 * F-ORG-LIST, F-ORG-CRUD, F-ORG-MEM, F-ORG-CHART 를 검증하는 E2E 테스트.
 * 매핑 TC: TC-ORG-LIST-01, TC-ORG-LIST-02, TC-ORG-UNIT-01, TC-ORG-MEM-01, TC-ORG-CHART-01
 */

test.describe("/admin/settings/organization — Org Management CRUD", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto("/admin/settings/organization");
    
    // Wait for the page to stabilize
    await page.waitForLoadState("networkidle");
    
    // Wait for loading to finish
    await expect(page.getByText(/initializing organization data/i)).toBeHidden({ timeout: 20_000 });
    
    // Wait for the actual content (Search placeholder)
    await expect(page.getByPlaceholder(/search units by name or type/i)).toBeVisible({ timeout: 15_000 });
  });

  test("TC-ORG-LIST-01 — List View 기본 노출 및 컬럼 확인", async ({ page }) => {
    await expect(page.getByRole("table")).toBeVisible();
    await expect(page.getByRole("columnheader", { name: /unit name/i })).toBeVisible();
    await expect(page.getByRole("columnheader", { name: /type/i })).toBeVisible();
    await expect(page.getByRole("columnheader", { name: /leader/i })).toBeVisible();
    await expect(page.getByRole("columnheader", { name: /members/i })).toBeVisible();
  });

  test("TC-ORG-LIST-02 — 조직 단위 검색 (Search)", async ({ page }) => {
    const searchInput = page.getByPlaceholder(/search units by name or type/i);
    await searchInput.fill("Engineering");
    
    // Check if only Engineering related units are shown
    const rows = page.getByRole("row");
    await expect(rows.filter({ hasText: "Engineering" }).first()).toBeVisible();
    
    // Search for something non-existent
    await searchInput.fill("NonExistentUnitXYZ");
    await expect(page.getByText(/no matching units found/i)).toBeVisible();
  });

  test("TC-ORG-UNIT-01 — 신규 부서 생성 모달 열림", async ({ page }) => {
    await page.getByRole("button", { name: /create unit/i }).click();
    
    const modal = page.getByRole("dialog");
    await expect(modal).toBeVisible();
    await expect(modal.getByText(/create new unit/i)).toBeVisible();
    
    // Close modal via aria-label
    await modal.getByRole("button", { name: /close/i }).click();
    await expect(modal).toBeHidden();
  });

  test("TC-ORG-MEM-01 — 멤버 관리 모달 진입", async ({ page }) => {
    // Find the first unit and click 'Members' button
    const firstRow = page.getByRole("row").nth(1); // Row 0 is header
    await firstRow.hover(); // Actions are hidden until hover
    
    const manageButton = firstRow.getByRole("button", { name: /members/i });
    await manageButton.click();
    
    const modal = page.getByRole("dialog");
    await expect(modal).toBeVisible();
    await expect(modal.getByText(/manage members/i)).toBeVisible();
    
    // Close modal via aria-label
    await modal.getByRole("button", { name: /close/i }).click();
    await expect(modal).toBeHidden();
  });

  test("TC-ORG-CHART-01 — Org Chart 뷰 전환 확인", async ({ page }) => {
    await page.getByRole("button", { name: /org chart/i }).click();
    
    // Wait for chart rendering spinner to finish (if any)
    const renderingText = page.getByText(/rendering hierarchy/i);
    if (await renderingText.isVisible()) {
      await expect(renderingText).toBeHidden({ timeout: 15_000 });
    }
    
    // ReactFlow container should be visible
    await expect(page.locator(".react-flow")).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText(/scope filter/i)).toBeVisible();
  });
});
