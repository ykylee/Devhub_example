import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";

const SERVER_ERROR_TEXT = "서버 오류가 발생했습니다. 잠시 후 다시 시도해주세요.";

test.describe("platform/project detail negative states", () => {
  test("TC-APP-DETAIL-ERR-01 — dashboard 조회 실패 시 에러와 retry 노출", async ({ page }) => {
    // PR #389 가 platform detail page 의 데이터 소스를 rollup → dashboard
    // 단일 엔드포인트로 통합. TC ID 는 추적성 매트릭스 보존 위해 유지하고
    // 라우트 stub 만 `/dashboard` 로 이전한다.
    await loginAs(page, SEEDED.systemAdmin);

    let dashboardRequestCount = 0;
    await page.route(/\/api\/v1\/platforms\/1\/repositories$/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ status: "ok", data: [] }),
      });
    });
    await page.route(/\/api\/v1\/platforms\/1\/dashboard$/, async (route) => {
      dashboardRequestCount += 1;
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ error: "forced platform dashboard failure" }),
      });
    });

    await page.goto(appPath("/platforms/1"));
    await expect(page.getByText(SERVER_ERROR_TEXT, { exact: true })).toBeVisible({ timeout: 15_000 });

    const retry = page.getByRole("button", { name: /retry/i }).first();
    await expect(retry).toBeVisible();
    await retry.click();

    await expect(page.getByText(SERVER_ERROR_TEXT, { exact: true })).toBeVisible({ timeout: 15_000 });
    expect(dashboardRequestCount).toBeGreaterThanOrEqual(2);
  });

  test("TC-PROJ-DETAIL-WARN-01 — repositories 실패 시 경고 배너 노출 및 activity 실패 시 온화한 fallback", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);

    await page.route(/\/api\/v1\/projects\/1$/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          status: "ok",
          data: {
            id: "1",
            platform_id: "1",
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
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ error: "forced repositories failure" }),
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
    // 1. repositories 실패 시 "Linked Repositories" 경고 배너 노출 검증
    await expect(page.getByText("일부 프로젝트 데이터를 불러오지 못했습니다: Linked Repositories", { exact: true })).toBeVisible({ timeout: 15_000 });
    // 2. activity 실패 시 온화하게 대체된 텍스트 확인 (Recent Activity 에러 배너는 없고 "활동 이력이 없습니다"가 노출)
    await expect(page.getByText("활동 이력이 없습니다.", { exact: true })).toBeVisible();
    await expect(page.getByRole("heading", { name: /stub project/i })).toBeVisible();
  });
});
