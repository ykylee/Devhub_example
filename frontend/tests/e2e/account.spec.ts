import { test, expect, loginAs, SEEDED } from "./fixtures";

// account.spec — F2 내 계정 / 비밀번호.
// TC-ACC-01..03 + TC-ACC-PROFILE-01. PR-UX2 (Kratos privileged session
// 안내 + aria-describedby) 검증 + 기존 password-change.spec 흡수.
//
// password-change.spec.ts 는 본 파일이 흡수하므로 삭제됨. globalSetup
// (PR-T3.5 hardening) 이 매 run 마다 시드 비밀번호를 force-reset 하므로
// finally rollback 이 실패해도 다음 invocation 에서 자동 복구된다.

test.describe("/account — UX 안내 + 사용자 정보", () => {
  test("TC-ACC-01 — Current Password 안내 텍스트 + aria-describedby", async ({ page }) => {
    await loginAs(page, SEEDED.developer);
    await page.goto("/account");

    const help = page.locator("#current-password-help");
    await expect(help).toBeVisible();
    await expect(help).toContainText(/Ory Kratos/);
    await expect(help).toContainText(/Sign In Again/i);

    // ARIA 연결 — input 에 aria-describedby=current-password-help.
    const currentInput = page.locator("#current-password");
    await expect(currentInput).toHaveAttribute("aria-describedby", "current-password-help");
  });

  test("TC-ACC-03 — New ≠ Confirm 클라이언트 측 검증", async ({ page }) => {
    await loginAs(page, SEEDED.developer);
    await page.goto("/account");

    // Current Password 값은 form 의 클라이언트 측 short-circuit 이
    // 먼저 동작하므로 임의의 non-empty 값으로 충분 — backend 호출
    // 발생하지 않음.
    await page.getByLabel(/current password/i).fill("anything");
    await page.getByLabel(/^new password$/i).fill("Foo-12345!");
    await page.getByLabel(/confirm new password/i).fill("Bar-67890!");
    await page.getByRole("button", { name: /save changes/i }).click();

    await expect(page.getByText(/New passwords do not match/i)).toBeVisible({ timeout: 5_000 });
  });

  test("TC-ACC-PROFILE-01 — Profile Info 에 alice 의 login/email/role 만 노출", async ({ page }) => {
    await loginAs(page, SEEDED.developer);
    await page.goto("/account");

    // Profile Info 카드 — alice 의 user_id, role, email 이 모두 등장.
    await expect(page.getByText(SEEDED.developer.user_id, { exact: true }).first()).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText(`${SEEDED.developer.user_id}@example.com`)).toBeVisible();
    // role display name — frontend store maps wire 'developer' → UI 'Developer'.
    await expect(page.getByText("Developer", { exact: true }).first()).toBeVisible();

    // 다른 시드 사용자 정보가 어디에도 노출되지 않아야 함.
    await expect(page.getByText(SEEDED.manager.user_id, { exact: true })).toHaveCount(0);
    await expect(page.getByText(SEEDED.systemAdmin.user_id, { exact: true })).toHaveCount(0);
  });
});

test.describe("/account — password 변경 round-trip (옛 password-change.spec 흡수)", () => {
  const original = SEEDED.developer.password;
  const rotated = `Rotated-${Date.now()}!`;

  test("TC-ACC-02 — change password, sign out, re-authenticate with the new password", async ({ page }) => {
    let passwordRotated = false;
    try {
      // 1) log in with the seeded password
      await loginAs(page, SEEDED.developer);

      // 2) /account 폼으로 비밀번호 변경
      await page.goto("/account");
      await page.getByLabel(/current password/i).fill(original);
      await page.getByLabel(/^new password$/i).fill(rotated);
      await page.getByLabel(/confirm new password/i).fill(rotated);
      await page.getByRole("button", { name: /save changes/i }).click();
      await expect(page.getByText(/password updated successfully/i)).toBeVisible({ timeout: 15_000 });
      passwordRotated = true;

      // 3) Sign Out — Hydra 세션 종료
      await page.getByText(SEEDED.developer.user_id, { exact: false }).first().click();
      await page.getByRole("button", { name: /sign out/i }).click();

      // 4) 새 password 로 로그인
      await loginAs(page, { ...SEEDED.developer, password: rotated });
    } finally {
      if (!passwordRotated) return;
      // best-effort rollback. globalSetup (PR-T3.5 hardening) 이 다음
      // run 에서 시드 비밀번호를 force-reset 하므로 여기 실패해도 suite
      // 가 다음 invocation 에서 자동 복구된다.
      try {
        await page.goto("/account");
        await page.getByLabel(/current password/i).fill(rotated);
        await page.getByLabel(/^new password$/i).fill(original);
        await page.getByLabel(/confirm new password/i).fill(original);
        await page.getByRole("button", { name: /save changes/i }).click();
        await expect(page.getByText(/password updated successfully/i)).toBeVisible({ timeout: 15_000 });
      } catch (err) {
        console.warn("[account.spec] best-effort rollback failed (globalSetup will recover on next run):", err);
      }
    }
  });
});
