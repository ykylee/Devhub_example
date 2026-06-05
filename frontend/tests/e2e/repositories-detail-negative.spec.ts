import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";

const SERVER_ERROR_TEXT = "서버 오류가 발생했습니다. 잠시 후 다시 시도해주세요.";

test.describe("/repositories/[id] — detail error states", () => {
  test("TC-REPO-DETAIL-ERR-01 — activity 조회 실패 시 graceful degradation (non-fatal)", async ({ page }) => {
    // PR #482 P2 fix: RepositoryDashboardView splits Promise.all — activity failure
    // is non-fatal. Page should render successfully with empty activity data, NOT
    // enter the legacy error state. This is a deliberate behavior change from the
    // legacy page which used Promise.all and surfaced activity 500s as page errors.
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

    // Page renders with h1 (repo data loaded) — activity failure does NOT block
    await expect(page.getByRole("heading", { name: /e2e-repo-a/i })).toBeVisible({ timeout: 15_000 });

    // Legacy error state text should NOT be visible
    await expect(page.getByText(SERVER_ERROR_TEXT, { exact: true })).not.toBeVisible();

    // Activity endpoint was hit at least once
    expect(requestCount).toBeGreaterThanOrEqual(1);
  });
});
