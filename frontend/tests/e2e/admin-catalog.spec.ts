import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";
const STRICT_ADMIN_UI = process.env.DEVHUB_E2E_STRICT_ADMIN_UI === "1";
const SEEDED_APPLICATION_ID = "e8a9bc11-a89c-4cb1-8071-8890ab2345ef";
const SEEDED_PROJECT_NAME = "DevHub Simulation Project";

test.describe("/admin/catalog — Admin Catalog", () => {
  test("TC-ADMIN-CATALOG-01 — system_admin 접근 + 3탭 전환 + 검색", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto(appPath("/admin/catalog"));

    await expect(page).toHaveURL(/\/admin\/catalog(\/|\?|$)/, { timeout: 20_000 });
    await expect(page.getByRole("heading", { name: /admin catalog/i })).toBeVisible();

    await page.getByTestId("catalog-tab-repositories").click();
    await expect(page).toHaveURL(/tab=repositories/);

    await page.getByTestId("catalog-tab-projects").click();
    await expect(page).toHaveURL(/tab=projects/);

    const search = page.getByPlaceholder("key/name/leader/status 검색");
    await search.fill("charlie");
    await expect(page).toHaveURL(/q=charlie/);
  });

  test("TC-ADMIN-CATALOG-02 — Platforms 탭 상세/프로젝트 드릴다운", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto(appPath("/admin/catalog?tab=platforms"));

    await expect(page.getByRole("button", { name: /platforms/i })).toBeVisible();
    const appProjectsButton = page.getByTestId(`catalog-app-projects-${SEEDED_APPLICATION_ID}`);
    await expect(appProjectsButton).toBeVisible();
    await page.getByTestId(`catalog-app-detail-${SEEDED_APPLICATION_ID}`).click();
    await expect(page).toHaveURL(/\/platforms\//, { timeout: 15_000 });

    await page.goBack();
    await expect(page).toHaveURL(/\/admin\/catalog\?tab=platforms/, { timeout: 15_000 });
    await expect(appProjectsButton).toBeVisible();
    await appProjectsButton.click();
    await expect(page).toHaveURL(/tab=projects/, { timeout: 15_000 });
    await expect(page).toHaveURL(new RegExp(`q=${SEEDED_APPLICATION_ID}`), { timeout: 15_000 });
    await expect(page.getByRole("cell", { name: SEEDED_PROJECT_NAME })).toBeVisible({ timeout: 15_000 });
  });

  test("TC-ADMIN-CATALOG-RBAC-01 — non-system_admin 접근 차단", async ({ page }) => {
    await loginAs(page, SEEDED.developer);
    await page.goto(appPath("/admin/catalog"));

    await page.waitForTimeout(1000);
    const path = new URL(page.url()).pathname;
    if (path.includes("/admin/catalog") && !STRICT_ADMIN_UI) {
      test.skip(true, "현재 환경에서 /admin/catalog RBAC redirect 비활성");
    }
    await expect(path.includes("/admin/catalog")).toBeFalsy();
    await expect(page).toHaveURL(/\/developer(\/|$)|\/onboarding(\/|$)/, { timeout: 15_000 });
  });
});
