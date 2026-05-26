import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";

test.describe("/repositories — error and empty states", () => {
  test("TC-REPO-ERR-01 — 목록 조회 실패 시 에러와 retry 노출", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);

    let requestCount = 0;
    await page.route(/\/api\/v1\/repositories$/, async (route) => {
      requestCount += 1;
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ error: "forced repository failure" }),
      });
    });

    await page.goto(appPath("/repositories"));
    await expect(page.getByText("Failed to load repositories data.", { exact: true })).toBeVisible({ timeout: 15_000 });

    const retry = page.getByRole("button", { name: /retry/i }).first();
    await expect(retry).toBeVisible();
    await retry.click();

    await expect(page.getByText("Failed to load repositories data.", { exact: true })).toBeVisible({ timeout: 15_000 });
    expect(requestCount).toBeGreaterThanOrEqual(2);
  });

  test("TC-REPO-EMPTY-01 — 빈 목록 응답 시 empty state 노출", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);

    await page.route(/\/api\/v1\/repositories$/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          status: "ok",
          data: [],
        }),
      });
    });

    await page.goto(appPath("/repositories"));
    await expect(page.getByText("No repositories matching your filters", { exact: true })).toBeVisible({ timeout: 15_000 });
  });
});
