import { expect, loginAs, SEEDED, test } from "./fixtures";

test.describe("/admin/settings/applications — CRUD UI smoke", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto("/admin/settings/applications");
    await expect(page).toHaveURL(/\/admin\/settings\/applications/, { timeout: 15_000 });
  });

  test("TC-APP-UI-01 — Applications 탭 진입 + New Application 버튼 노출", async ({ page }) => {
    await expect(page.getByRole("button", { name: /new application/i })).toBeVisible();
  });

  test("TC-APP-UI-02 — New Application 모달 open/close", async ({ page }) => {
    await page.getByRole("button", { name: /new application/i }).click();
    const dialog = page.getByRole("dialog");
    await expect(dialog).toBeVisible();
    await page.keyboard.press("Escape");
    await expect(dialog).toBeHidden();
  });

  test("TC-APP-UI-03 — 필수값 없이 submit 시 브라우저 검증으로 제출 차단", async ({ page }) => {
    await page.getByRole("button", { name: /new application/i }).click();
    const dialog = page.getByRole("dialog");
    await expect(dialog).toBeVisible();

    await dialog.getByRole("button", { name: /create application/i }).click();

    // required input validation 유지 시 모달이 그대로 열려 있어야 한다.
    await expect(dialog).toBeVisible();
  });

  test("TC-APP-SEARCH-01 — leader/dev unit 키워드로 검색 가능", async ({ page }) => {
    const unique = Date.now().toString().slice(-6);
    const appKey = `A${unique}BCDE`;
    const appName = `E2E Search ${unique}`;
    const leader = "charlie";
    const devUnit = "dept-eng";

    await page.getByRole("button", { name: /new application/i }).click();
    const dialog = page.getByRole("dialog");
    await expect(dialog).toBeVisible();

    await dialog.getByPlaceholder("E.G. PLATFORM26").fill(appKey);
    await dialog.getByPlaceholder("e.g. DevHub Platform 2026").fill(appName);
    await dialog.getByPlaceholder("Strategic goals, KPI, and scope summary...").fill("e2e search coverage");
    await dialog.getByPlaceholder("e.g. charlie").first().fill(leader);
    await dialog.getByPlaceholder("e.g. dept-eng").fill(devUnit);
    await dialog.getByPlaceholder("e.g. charlie").nth(1).fill(leader);
    await dialog.getByRole("button", { name: /create application/i }).click();

    await expect(dialog).toBeHidden({ timeout: 10_000 });
    await expect(page.getByRole("row").filter({ hasText: appName })).toBeVisible({ timeout: 15_000 });

    const searchInput = page.getByPlaceholder("Search by name, key, or owner...");
    await searchInput.fill(leader);
    await expect(page.getByRole("row").filter({ hasText: appName })).toBeVisible({ timeout: 15_000 });

    await searchInput.fill(devUnit);
    await expect(page.getByRole("row").filter({ hasText: appName })).toBeVisible({ timeout: 15_000 });
  });
});
