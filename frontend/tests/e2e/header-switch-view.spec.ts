import { test, expect, loginAs, SEEDED } from "./fixtures";

// header-switch-view.spec — F3 헤더 / 네비게이션.
// TC-NAV-01..03 + TC-NAV-SIM-01. PR-UX3 (안내 1줄 추가) + 기존 dropdown
// 동작 회귀 + 시뮬레이션 우회 불가 확인.

test.describe("Header Switch View", () => {
  test("TC-NAV-01 — Switch View 한계 안내 노출", async ({ page }) => {
    await loginAs(page, SEEDED.developer);

    // Header dropdown trigger: user 영역 (avatar + login). Header 코드에
    // motion.div 로 감싸져 있어 role="button" 이 아니라 click 가능한
    // 텍스트 영역. SEEDED.developer.user_id 가 첫 번째 가시 요소.
    await page.getByText(SEEDED.developer.user_id, { exact: false }).first().click();

    // PR-UX3 가 추가한 안내 텍스트.
    await expect(page.getByText(/Menu preview only — actual permissions follow server actor\.role/i)).toBeVisible();
    // 그 위의 "Switch View" 헤더도 같이 보여야 함.
    await expect(page.getByText(/^Switch View$/)).toBeVisible();
  });

  test("TC-NAV-02 — Switch View role 전환 회귀 (Developer → Manager)", async ({ page }) => {
    await loginAs(page, SEEDED.developer);

    await page.getByText(SEEDED.developer.user_id, { exact: false }).first().click();

    // dropdown 의 "Manager" 버튼 클릭. (다른 페이지의 manager 라벨과 충돌
    // 없도록 dropdown context 안에서 찾되, exact match 로 좁힘)
    await page.getByRole("button", { name: "Manager" }).click();

    // store role 가 바뀌고 router.push("/manager") 호출.
    await page.waitForURL(/\/manager(\/|$)/, { timeout: 10_000 });

    // Wait for the dropdown exit animation to finish before asserting on the
    // header role badge. While the dropdown is still on its way out, both the
    // badge <span> and the "Manager" <button> match `getByText("Manager")` and
    // trigger a strict-mode violation. The "Switch View" caption only exists
    // inside the dropdown, so its disappearance is a reliable close signal —
    // observed flaky locally, deterministic in CI (PR #86 review pass).
    await expect(page.getByText(/^Switch View$/)).not.toBeVisible({ timeout: 5_000 });

    // Header 의 role 표시 (line 100 의 <span> with role) 가 "Manager".
    // exact 텍스트 매칭으로 다른 페이지 콘텐츠와 분리.
    await expect(page.locator("header").getByText("Manager", { exact: true })).toBeVisible();
  });

  test("TC-NAV-03 — Account Settings 메뉴 → /account 이동", async ({ page }) => {
    await loginAs(page, SEEDED.developer);

    await page.getByText(SEEDED.developer.user_id, { exact: false }).first().click();
    await page.getByRole("button", { name: /account settings/i }).click();

    await page.waitForURL(/\/account(\/|$)/, { timeout: 10_000 });
  });

  test("TC-NAV-SIM-01 — Switch View 시뮬레이션은 서버 actor.role 우회 불가", async ({ page }) => {
    // charlie (system_admin) 로 로그인.
    await loginAs(page, SEEDED.systemAdmin);

    // Switch View 로 Developer 시뮬레이션.
    await page.getByText(SEEDED.systemAdmin.user_id, { exact: false }).first().click();
    await page.getByRole("button", { name: "Developer" }).click();
    await page.waitForURL(/\/developer(\/|$)/, { timeout: 10_000 });

    // URL 로 직접 admin 영역 진입 — actor.role 가 system_admin 이므로
    // AuthGuard 통과. (시뮬레이션은 store role 만 바꾸고 actor.role 은
    // 그대로 system_admin.)
    await page.goto("/admin/settings/users");
    await expect(page).toHaveURL(/\/admin\/settings\/users/, { timeout: 10_000 });
    // 페이지가 정상 렌더되는지 — 사용자 목록 헤더가 안정 anchor.
    await expect(page.getByText(/Organization/i).first()).toBeVisible({ timeout: 10_000 });
  });
});
