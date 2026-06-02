import { test, expect, loginAs, SEEDED, appPath } from "./fixtures";

// account.spec — F2 내 계정 / Keycloak Account Console redirect.
//
// ADR-0019 / sprint claude/work_260519-ad: self-service 비밀번호 변경
// 흐름은 Keycloak Account Console 로 위임됐다. 기존 TC-ACC-01..03 의
// password form 시나리오는 dead UI element 의존이라 삭제. profile info
// 노출 (TC-ACC-PROFILE-01) 만 유지하고 console redirect 버튼 검증을
// TC-ACC-KEYCLOAK-CONSOLE-01 로 신규 추가.

test.describe("/account — profile + Keycloak console", () => {
  test("TC-ACC-PROFILE-01 — Profile Info 에 alice 의 login/email/role 만 노출", async ({ page }) => {
    await loginAs(page, SEEDED.developer);
    await page.goto(appPath("/account"));

    await expect(page.getByText(SEEDED.developer.user_id, { exact: true }).first()).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText(`${SEEDED.developer.user_id}@example.com`)).toBeVisible();
    await expect(page.getByText("Developer", { exact: true }).first()).toBeVisible();

    await expect(page.getByText(SEEDED.team_manager.user_id, { exact: true })).toHaveCount(0);
    await expect(page.getByText(SEEDED.systemAdmin.user_id, { exact: true })).toHaveCount(0);
  });

  test("TC-ACC-KEYCLOAK-CONSOLE-01 — Open Keycloak Console 버튼이 외부 issuer/account/ 로 향한다", async ({ page }) => {
    await loginAs(page, SEEDED.developer);
    await page.goto(appPath("/account"));

    const consoleLink = page.getByRole("link", { name: /open keycloak console/i });
    await expect(consoleLink).toBeVisible({ timeout: 10_000 });

    const href = await consoleLink.getAttribute("href");
    expect(href, "Keycloak Console link should resolve to an issuer/account/ URL").toBeTruthy();
    expect(href).toMatch(/\/account\/?$/);
    expect(await consoleLink.getAttribute("target")).toBe("_blank");
    expect(await consoleLink.getAttribute("rel") ?? "").toMatch(/noopener/);
  });
});
