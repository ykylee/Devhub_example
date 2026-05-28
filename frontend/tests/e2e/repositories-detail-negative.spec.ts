import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";

const SERVER_ERROR_TEXT = "서버 오류가 발생했습니다. 잠시 후 다시 시도해주세요.";

test.describe("/repositories/[id] — detail error states", () => {
  test("TC-REPO-DETAIL-ERR-01 — activity 조회 실패 시 에러와 retry 노출", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);

    let requestCount = 0;
    await page.route(/\/api\/v1\/repositories\/1\/activity$/, async (route) => {
      requestCount += 1;
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ error: "forced repository activity failure" }),
      });
    });

    await page.goto(appPath("/repositories/1"));
    await expect(page.getByText(SERVER_ERROR_TEXT, { exact: true })).toBeVisible({ timeout: 15_000 });

    const retry = page.getByRole("button", { name: /retry/i }).first();
    await expect(retry).toBeVisible();
    await retry.click();

    await expect(page.getByText(SERVER_ERROR_TEXT, { exact: true })).toBeVisible({ timeout: 15_000 });
    expect(requestCount).toBeGreaterThanOrEqual(2);
  });
});
