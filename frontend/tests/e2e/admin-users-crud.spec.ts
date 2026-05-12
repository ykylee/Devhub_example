import { test, expect, loginAs, SEEDED } from "./fixtures";

// admin-users-crud.spec — F1 사용자 관리 / CRUD UI smoke.
// TC-USR-CRUD-01..03. backend mutation 회피 — UserCreationModal 열림 /
// Role select 옵션 노출 / Action 메뉴 admin 액션 노출만 확인. 실제
// round-trip 은 backend Go test (PR #54 10 cases) 가 커버.

test.describe("/admin/settings/users — CRUD UI smoke", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto("/admin/settings/users");
    await expect(page.getByRole("row").filter({ hasText: "alice" })).toBeVisible({ timeout: 15_000 });
  });

  test("TC-USR-CRUD-01 — Invite Member 모달 열림 + 닫기", async ({ page }) => {
    await page.getByRole("button", { name: /invite member/i }).click();

    // UserCreationModal renders form fields — the modal title 'Create
    // User' (or similar Modal title prop) is the closest stable anchor.
    // Probe a few common modal anchors; pick the first that resolves.
    const modal = page.getByRole("dialog");
    await expect(modal).toBeVisible({ timeout: 5_000 });

    // Close via the ESC key — Modal component handles onClose. Avoid
    // form submit so no backend mutation fires.
    await page.keyboard.press("Escape");
    await expect(modal).toBeHidden({ timeout: 5_000 });
  });

  test("TC-USR-CRUD-02 — Role select dropdown 옵션 노출", async ({ page }) => {
    // The first MemberTable row's Role <select> is the smoke target.
    // We use the alice row as the anchor (always present in the seed).
    const aliceRow = page.getByRole("row").filter({ hasText: "alice" });
    const roleSelect = aliceRow.locator("select").first();

    await expect(roleSelect).toBeVisible();

    // Roles come from rbacService.listPolicies() with a defaultRoles
    // fallback — three roles (Developer / Manager / System Admin) are
    // guaranteed by defaultRoles even if the policy call fails.
    const options = await roleSelect.locator("option").allTextContents();
    expect(options.length).toBeGreaterThanOrEqual(3);
    expect(options).toEqual(expect.arrayContaining(["Developer", "Manager", "System Admin"]));
  });

  test("TC-USR-CRUD-03 — Action 메뉴 (...) 에 system_admin 액션 3종 노출", async ({ page }) => {
    const aliceRow = page.getByRole("row").filter({ hasText: "alice" });

    // The MoreHorizontal trigger does not have an accessible label;
    // it's the only button inside the action <td>. The action <td> is
    // the last cell; targeting its button works without coupling to
    // implementation detail.
    await aliceRow.locator("td").last().locator("button").click();

    // Dropdown panel contains the three admin actions when
    // currentUserRole === "System Admin". They render as <button> with
    // distinctive text.
    await expect(page.getByRole("button", { name: /issue account/i })).toBeVisible();
    await expect(page.getByRole("button", { name: /force reset password/i })).toBeVisible();
    await expect(page.getByRole("button", { name: /revoke account/i })).toBeVisible();

    // Close menu via overlay click (the menu fixed-inset overlay)
    // — pressing Escape doesn't close it because the component listens
    // for backdrop onClick. Click outside the menu area instead.
    await page.mouse.click(10, 10);
    await expect(page.getByRole("button", { name: /issue account/i })).toBeHidden();
  });
});
