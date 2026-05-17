import { test, expect, loginAs, SEEDED } from "./fixtures";

/**
 * admin-users-search.spec.ts
 * F-USR-SEARCH 를 검증하는 E2E 테스트.
 * 매핑 TC: TC-USR-01, TC-USR-02, TC-USR-03, TC-USR-04, TC-USR-05, TC-USR-06
 */

test.describe("/admin/settings/users — 검색 필터", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto("/admin/settings/users");
    // MemberTable rows hydrate after the GET /api/v1/users round-trip.
    // Wait for alice row as a stable anchor that the page reached the
    // populated state (charlie also visible, but alice has the lowest
    // alphabetical bias in dependent assertions).
    await expect(page.getByRole("row").filter({ hasText: "alice" })).toBeVisible({ timeout: 15_000 });
  });

  // Stable selector for a MemberTable user row. We scope to row-role
  // elements that contain the seed user_id so search-result assertions
  // are not confused by the table header or empty rows.
  const userRow = (page: import("@playwright/test").Page, userId: string) =>
    page.getByRole("row").filter({ hasText: userId });

  test("TC-USR-01 — name 부분일치 + 빈 검색어 복귀", async ({ page }) => {
    // Step 2: 최소 alice + bob 두 행이 보여야 함 (charlie 도 자기 자신)
    await expect(userRow(page, "alice")).toBeVisible();
    await expect(userRow(page, "bob")).toBeVisible();

    // Step 3-4: alice 입력 → alice 만
    const search = page.getByLabel("Search users");
    await search.fill("alice");
    await expect(userRow(page, "alice")).toBeVisible();
    await expect(userRow(page, "bob")).toHaveCount(0);

    // Step 5-6: 비우면 전체 복귀
    await search.fill("");
    await expect(userRow(page, "alice")).toBeVisible();
    await expect(userRow(page, "bob")).toBeVisible();
  });

  test("TC-USR-02 — email 부분일치", async ({ page }) => {
    const search = page.getByLabel("Search users");
    await search.fill("bob@");
    await expect(userRow(page, "bob")).toBeVisible();
    await expect(userRow(page, "alice")).toHaveCount(0);
  });

  test("TC-USR-03 — role 부분일치", async ({ page }) => {
    const search = page.getByLabel("Search users");
    // MemberTable renders role display name "Manager" — wire format is
    // "manager", display is "Manager". The filter compares case-insensitive
    // substring so either works; using lowercase asserts that the
    // toLowerCase() path actually matches the display strings too.
    await search.fill("manager");
    await expect(userRow(page, "bob")).toBeVisible();
    await expect(userRow(page, "alice")).toHaveCount(0);
    await expect(userRow(page, "charlie")).toHaveCount(0);
  });

  test("TC-USR-04 — Filter 버튼 disabled", async ({ page }) => {
    const filterBtn = page.getByRole("button", { name: /advanced filters/i });
    await expect(filterBtn).toBeDisabled();
    await expect(filterBtn).toHaveAttribute("title", "Advanced filters coming soon");
  });

  test("TC-USR-05 — case-insensitive 매칭", async ({ page }) => {
    const search = page.getByLabel("Search users");
    await search.fill("ALICE");
    await expect(userRow(page, "alice")).toBeVisible();
  });

  test("TC-USR-06 — 매칭 0건 (empty result)", async ({ page }) => {
    const search = page.getByLabel("Search users");
    await search.fill("zzzz-no-match");

    // No MemberTable user rows. The header <tr> still exists, so we
    // assert specifically on rows that carry a seed user_id — that's the
    // population we filtered against.
    await expect(userRow(page, "alice")).toHaveCount(0);
    await expect(userRow(page, "bob")).toHaveCount(0);
    await expect(userRow(page, "charlie")).toHaveCount(0);

    // Page itself stays interactive: clearing the query restores rows.
    await search.fill("");
    await expect(userRow(page, "alice")).toBeVisible();
  });
});
