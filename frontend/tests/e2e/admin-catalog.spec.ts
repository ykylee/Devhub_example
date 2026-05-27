import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";

test.describe("/admin/catalog — Admin Catalog", () => {
  test("TC-ADMIN-CATALOG-01 — system_admin 접근 + 3탭 전환 + 검색", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto(appPath("/admin/catalog"));

    await expect(page).toHaveURL(/\/admin\/catalog(\/|\?|$)/, { timeout: 20_000 });
    await expect(page.getByRole("heading", { name: /admin catalog/i })).toBeVisible();

    await page.getByRole("button", { name: /repositories/i }).click();
    await expect(page).toHaveURL(/tab=repositories/);

    await page.getByRole("button", { name: /projects/i }).click();
    await expect(page).toHaveURL(/tab=projects/);

    const search = page.getByPlaceholder("key/name/owner/status 검색");
    await search.fill("charlie");
    await expect(page).toHaveURL(/q=charlie/);
  });

  test("TC-ADMIN-CATALOG-02 — Applications 탭 상세/프로젝트 드릴다운", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto(appPath("/admin/catalog?tab=applications"));

    await expect(page.getByRole("button", { name: /applications/i })).toBeVisible();

    const rows = page.locator("tbody tr");
    const rowCount = await rows.count();
    test.skip(rowCount === 0, "applications 데이터가 없어 드릴다운 검증 생략");

    const firstRow = rows.first();
    await firstRow.getByRole("link", { name: /detail/i }).click();
    await expect(page).toHaveURL(/\/applications\//, { timeout: 15_000 });

    await page.goto(appPath("/admin/catalog?tab=applications"));
    await firstRow.getByRole("button", { name: /projects/i }).click();
    await expect(page).toHaveURL(/tab=projects/, { timeout: 15_000 });
  });

  test("TC-ADMIN-CATALOG-RBAC-01 — non-system_admin 접근 차단", async ({ page }) => {
    await loginAs(page, SEEDED.developer);
    await page.goto(appPath("/admin/catalog"));

    const path = new URL(page.url()).pathname;
    expect(path.includes("/admin/catalog")).toBeFalsy();
    await expect(page).toHaveURL(/\/developer(\/|$)/, { timeout: 15_000 });
  });
});
