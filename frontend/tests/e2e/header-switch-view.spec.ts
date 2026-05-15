import { test, expect, loginAs, SEEDED } from "./fixtures";

// header-switch-view.spec — F3 헤더 / 네비게이션.
//
// PR #115 (light theme + dropdown refactor) 에서 dropdown 의 "Switch View"
// (Developer/Manager/System Admin) 섹션이 의도적으로 제거되었다. 이에 따라
// 기존 TC-NAV-01 (한계 안내 노출), TC-NAV-02 (role 전환 회귀),
// TC-NAV-SIM-01 (시뮬레이션 우회 불가) 은 검증 대상 UI 가 사라져 삭제.
//
// 남은 시나리오:
//   - TC-NAV-03: Account Profile 항목 → /account 이동 (dropdown 의 동일한
//     entry point 가 유지됨, label 만 "Account Settings" → "Account Profile").
//
// traceability(report.md) 의 TC-NAV-01/02/SIM-01 row 정리는 후속 PR 에서.

test.describe("Header dropdown", () => {
  test("TC-NAV-03 — Account Profile 메뉴 → /account 이동", async ({ page }) => {
    await loginAs(page, SEEDED.developer);

    await page.getByText(SEEDED.developer.user_id, { exact: false }).first().click();
    await page.getByRole("button", { name: /account profile/i }).click();

    await page.waitForURL(/\/account(\/|$)/, { timeout: 10_000 });
  });
});
