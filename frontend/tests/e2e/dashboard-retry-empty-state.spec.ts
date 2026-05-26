import { test, expect, SEEDED, loginAs, appPath } from "./fixtures";
const METRICS_ERROR_TEXT = "서버 오류가 발생했습니다. 잠시 후 다시 시도해주세요.";

async function stubMetrics500(page: import("@playwright/test").Page, roleQuery: string): Promise<{ getCount: () => number }> {
  let metricsRequestCount = 0;
  await page.route(/\/api\/v1\/dashboard\/metrics\?role=.*/, async (route) => {
    const url = route.request().url();
    if (!url.includes(`role=${encodeURIComponent(roleQuery)}`)) {
      await route.continue();
      return;
    }
    metricsRequestCount += 1;
    await route.fulfill({
      status: 500,
      contentType: "application/json",
      body: JSON.stringify({ error: "forced test failure" }),
    });
  });
  return { getCount: () => metricsRequestCount };
}

test.describe("Dashboard retry/empty-state", () => {
  test("developer dashboard shows retry on metrics failure", async ({ page }) => {
    await loginAs(page, SEEDED.developer);
    const metrics = await stubMetrics500(page, "developer");

    await page.goto(appPath("/developer"));
    await expect(page.getByText(METRICS_ERROR_TEXT, { exact: true })).toBeVisible({ timeout: 15_000 });
    const retry = page.getByRole("button", { name: /retry/i }).first();
    await expect(retry).toBeVisible();
    await retry.click();
    await expect(page.getByText(METRICS_ERROR_TEXT, { exact: true })).toBeVisible();
    expect(metrics.getCount()).toBeGreaterThanOrEqual(2);
  });

  test("manager dashboard shows retry on metrics failure", async ({ page }) => {
    await loginAs(page, SEEDED.manager);
    const metrics = await stubMetrics500(page, "manager");

    await page.goto(appPath("/manager"));
    await expect(page.getByText(METRICS_ERROR_TEXT, { exact: true })).toBeVisible({ timeout: 15_000 });
    const retry = page.getByRole("button", { name: /retry/i }).first();
    await expect(retry).toBeVisible();
    await retry.click();
    await expect(page.getByText(METRICS_ERROR_TEXT, { exact: true })).toBeVisible();
    expect(metrics.getCount()).toBeGreaterThanOrEqual(2);
  });

  test("admin dashboard shows retry on metrics failure", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    const metrics = await stubMetrics500(page, "system_admin");

    await page.goto(appPath("/admin"));
    await expect(page.getByText(METRICS_ERROR_TEXT, { exact: true })).toBeVisible({ timeout: 15_000 });
    const retry = page.getByRole("button", { name: /retry/i }).first();
    await expect(retry).toBeVisible();
    await retry.click();
    await expect(page.getByText(METRICS_ERROR_TEXT, { exact: true })).toBeVisible();
    expect(metrics.getCount()).toBeGreaterThanOrEqual(2);
  });
});
