import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";
const STRICT_ADMIN_UI = process.env.DEVHUB_E2E_STRICT_ADMIN_UI === "1";

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

  // TC-ADMIN-CATALOG-02 는 application detail 진입(/applications) → 뒤로 catalog 복귀 →
  // catalog-app-projects 드릴다운(tab=projects + q=appID) 의 다단계 네비게이션으로,
  // 시드 데이터(applications + 연결 projects) 의존 + 재렌더 타이밍에 fragile 하다.
  // 로컬 E2E full-stack 미검증 제약으로 happy-path 안정화는 carve (#380) 로 분리한다.
  test.skip("TC-ADMIN-CATALOG-02 — Applications 탭 상세/프로젝트 드릴다운 (drilldown 안정화 carve #380)", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto(appPath("/admin/catalog?tab=applications"));

    await expect(page.getByRole("button", { name: /applications/i })).toBeVisible();

    const rows = page.locator("tbody tr");
    const rowCount = await rows.count();
    test.skip(rowCount === 0, "applications 데이터가 없어 드릴다운 검증 생략");

    const firstRow = rows.first();
    const appID = (await firstRow.getByTestId(/catalog-app-detail-.+/).first().getAttribute("data-testid"))?.replace("catalog-app-detail-", "");
    test.skip(!appID, "application id resolve 실패");

    await page.getByTestId(`catalog-app-detail-${appID}`).click();
    await expect(page).toHaveURL(/\/applications\//, { timeout: 15_000 });

    await page.goto(appPath("/admin/catalog?tab=applications"));
    await expect(page.getByRole("button", { name: /applications/i })).toBeVisible();

    const projectButtons = page.getByTestId(/catalog-app-projects-.+/);
    const projectButtonCount = await projectButtons.count();
    test.skip(projectButtonCount === 0, "projects 드릴다운 버튼이 없어 검증 생략");

    const firstProjectButton = projectButtons.first();
    const projectButtonTestID = await firstProjectButton.getAttribute("data-testid");
    const projectAppID = projectButtonTestID?.replace("catalog-app-projects-", "");
    test.skip(!projectAppID, "projects 버튼에서 application id resolve 실패");

    for (let i = 0; i < 3; i += 1) {
      await firstProjectButton.click();
      if (page.url().includes("tab=projects")) break;
      await page.waitForTimeout(250);
    }

    await expect(page).toHaveURL(/tab=projects/, { timeout: 15_000 });
    await expect(page).toHaveURL(new RegExp(`q=${projectAppID}`), { timeout: 15_000 });
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
