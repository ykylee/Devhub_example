import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";

test.describe("/repositories — repository list/detail UI", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto(appPath("/repositories"));
    await expect(page).toHaveURL(new RegExp(`${appPath("/repositories")}(\\/|$)`), { timeout: 20_000 });
  });

  test("TC-REPO-UI-01 — 저장소 목록 진입 + fixture repository 노출", async ({ page }) => {
    await expect(page.getByText("e2e-repo-a", { exact: false })).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText("e2e-repo-b", { exact: false })).toBeVisible({ timeout: 20_000 });
  });

  test("TC-REPO-SEARCH-01 — 저장소명 검색으로 목록 필터링", async ({ page }) => {
    const searchInput = page.getByPlaceholder("Search repositories by name or owner...");

    await searchInput.fill("e2e-repo-a");
    await expect(page.getByText("e2e-repo-a", { exact: false })).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText("e2e-repo-b", { exact: false })).toBeHidden({ timeout: 15_000 });

    await searchInput.fill("devhub");
    await expect(page.getByText("e2e-repo-a", { exact: false })).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText("e2e-repo-b", { exact: false })).toBeVisible({ timeout: 15_000 });
  });

  test("TC-REPO-DETAIL-01 — 저장소 상세 진입 + 핵심 활동 카드 노출", async ({ page }) => {
    await page.getByRole("link", { name: /e2e-repo-a/i }).first().click();
    await expect(page).toHaveURL(new RegExp(`${appPath("/repositories")}/\\d+$`), { timeout: 20_000 });

    await expect(page.getByRole("heading", { name: /e2e-repo-a/i })).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText("devhub/e2e-repo-a", { exact: false })).toBeVisible();
    await expect(page.getByText("PR Events", { exact: false }).first()).toBeVisible();
    await expect(page.getByText("Build Runs", { exact: false }).first()).toBeVisible();
    // PR #396 (REQ-FR-APPDASH-001) — "Build Success: %" → "Last Build: status".
    await expect(page.getByText("Last Build", { exact: false }).first()).toBeVisible();
    await expect(page.getByText("Contributors", { exact: false }).first()).toBeVisible();
    await expect(page.getByRole("link", { name: /view on scm/i })).toBeVisible();
  });
});
