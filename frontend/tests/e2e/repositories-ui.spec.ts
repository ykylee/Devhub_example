import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";

test.describe("/repositories — repository list/detail UI", () => {
  const repoALink = (page: import("@playwright/test").Page) =>
    page.getByRole("link", { name: "e2e-repo-a", exact: true });
  const repoBLink = (page: import("@playwright/test").Page) =>
    page.getByRole("link", { name: "e2e-repo-b", exact: true });

  test.beforeEach(async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto(appPath("/repositories"));
    await expect(page).toHaveURL(new RegExp(`${appPath("/repositories")}(\\/|$)`), { timeout: 20_000 });
  });

  test("TC-REPO-UI-01 — 저장소 목록 진입 + fixture repository 노출", async ({ page }) => {
    await expect(repoALink(page)).toBeVisible({ timeout: 20_000 });
    await expect(repoBLink(page)).toBeVisible({ timeout: 20_000 });
  });

  test("TC-REPO-SEARCH-01 — 저장소명 검색으로 목록 필터링", async ({ page }) => {
    const searchInput = () => page.getByPlaceholder("Search repositories by name or owner...");

    await expect(searchInput()).toBeVisible({ timeout: 15_000 });
    await searchInput().fill("e2e-repo-a");
    await expect(repoALink(page)).toBeVisible({ timeout: 15_000 });
    await expect(repoBLink(page)).toBeHidden({ timeout: 15_000 });

    await expect(searchInput()).toBeVisible({ timeout: 15_000 });
    await searchInput().fill("devhub");
    await expect(repoALink(page)).toBeVisible({ timeout: 15_000 });
    await expect(repoBLink(page)).toBeVisible({ timeout: 15_000 });
  });

  test("TC-REPO-DETAIL-01 — 저장소 상세 진입 + 핵심 활동 카드 노출", async ({ page }) => {
    await page.getByRole("link", { name: /e2e-repo-a/i }).first().click();
    await expect(page).toHaveURL(new RegExp(`${appPath("/repositories")}/\\d+$`), { timeout: 20_000 });

    // ---- Repository header (preserved from legacy page) ----
    await expect(page.getByRole("heading", { name: /e2e-repo-a/i })).toBeVisible({ timeout: 15_000 });
    // full_name text renders slightly after the heading; default 5s is too tight.
    // CI race 환경에서 15s 도 intermittent fail — 20s 로 늘리고 regex selector 로 robust.
    await expect(page.getByText(/devhub\/e2e-repo-a/i).first()).toBeVisible({ timeout: 20_000 });
    await expect(page.getByRole("link", { name: /view on scm/i })).toBeVisible();

    // ---- New dashboard subheader (PR #482 revamp) ----
    await expect(page.getByText("Dashboard Perspective", { exact: false })).toBeVisible();
    await expect(page.getByText(/Viewing as/i)).toBeVisible();

    // Role-toggle tabs
    await expect(page.getByRole("button", { name: /^Developer$/i })).toBeVisible();
    await expect(page.getByRole("button", { name: /Manager & Governance/i })).toBeVisible();

    // Default view for systemAdmin is Manager (RepositoryDashboardView.tsx L51 useEffect).
    // Verify Manager view focus cards are rendered.
    await expect(page.getByText("Team Manager Focus")).toBeVisible();
    await expect(page.getByText("Organization Admin Focus")).toBeVisible();
    await expect(page.getByText("System Admin Focus")).toBeVisible();
    await expect(page.getByText("Repository Activity Trend")).toBeVisible();
    await expect(page.getByText("Contributor Distribution")).toBeVisible();
  });
});
