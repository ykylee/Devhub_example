import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";

const SERVER_ERROR_TEXT = "서버 오류가 발생했습니다. 잠시 후 다시 시도해주세요.";

test.describe("application/project detail negative states", () => {
  test("TC-APP-DETAIL-ERR-01 — rollup 조회 실패 시 에러와 retry 노출", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);

    let rollupRequestCount = 0;
    await page.route(/\/api\/v1\/applications\/1$/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          status: "ok",
          data: {
            id: "1",
            key: "APP-1",
            name: "Stub Application",
            description: "stub",
            status: "active",
            visibility: "internal",
            owner_user_id: "charlie",
            leader_user_id: "charlie",
            development_unit_id: "dept-eng",
            start_date: null,
            due_date: null,
            archived_at: null,
            created_at: "2026-01-01T00:00:00Z",
            updated_at: "2026-01-01T00:00:00Z"
          }
        }),
      });
    });
    await page.route(/\/api\/v1\/applications\/1\/repositories$/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ status: "ok", data: [] }),
      });
    });
    await page.route(/\/api\/v1\/applications\/1\/rollup$/, async (route) => {
      rollupRequestCount += 1;
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ error: "forced application rollup failure" }),
      });
    });

    await page.goto(appPath("/applications/1"));
    await expect(page.getByText(SERVER_ERROR_TEXT, { exact: true })).toBeVisible({ timeout: 15_000 });

    const retry = page.getByRole("button", { name: /retry/i }).first();
    await expect(retry).toBeVisible();
    await retry.click();

    await expect(page.getByText(SERVER_ERROR_TEXT, { exact: true })).toBeVisible({ timeout: 15_000 });
    expect(rollupRequestCount).toBeGreaterThanOrEqual(2);
  });

  test("TC-PROJ-DETAIL-WARN-01 — activity 일부 실패 시 경고 배너 노출", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);

    await page.route(/\/api\/v1\/projects\/1$/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          status: "ok",
          data: {
            id: "1",
            application_id: "1",
            repository_id: 1,
            repository_ids: [1],
            key: "PROJ-1",
            name: "Stub Project",
            description: "stub",
            status: "active",
            visibility: "internal",
            owner_user_id: "charlie",
            start_date: "2026-01-01",
            due_date: "2026-12-31",
            archived_at: null,
            created_at: "2026-01-01T00:00:00Z",
            updated_at: "2026-01-01T00:00:00Z"
          }
        }),
      });
    });
    await page.route(/\/api\/v1\/users$/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          status: "ok",
          data: [
            {
              user_id: "charlie",
              display_name: "Charlie",
              email: "charlie@example.com",
              role: "system_admin",
              status: "active",
              primary_unit_id: "dept-eng",
              current_unit_id: "dept-eng",
              is_seconded: false,
              appointments: [],
              joined_at: "2026-01-01"
            }
          ]
        }),
      });
    });
    await page.route(/\/api\/v1\/projects\/1\/repositories$/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ status: "ok", data: [] }),
      });
    });
    await page.route(/\/api\/v1\/projects\/1\/tasks(\?.*)?$/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ status: "ok", data: [] }),
      });
    });
    await page.route(/\/api\/v1\/projects\/1\/activity$/, async (route) => {
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ error: "forced project activity failure" }),
      });
    });

    await page.goto(appPath("/projects/1"));
    await expect(page.getByText("일부 프로젝트 데이터를 불러오지 못했습니다: Recent Activity", { exact: true })).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole("heading", { name: /stub project/i })).toBeVisible();
  });
});
