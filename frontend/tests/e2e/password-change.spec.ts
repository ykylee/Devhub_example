import { test, expect, loginAs, SEEDED } from "./fixtures";

// password-change.spec — /account password change via Kratos public
// settings flow (PR-L3, DEC-2=B). Validates the full round-trip: login
// with old password -> change -> sign out -> log in with new password ->
// (cleanup) restore the original password so other specs in the same
// `npm run e2e` (signout.spec reuses alice) keep their seeded login.
//
// The cleanup is best-effort: if the rollback throws (page closed mid-
// run, form not rendered, network blip), globalSetup will force-reset
// the seeded password on the next invocation (PR-T3.5 hardening,
// work_260512-f), so a single failed cleanup never leaves the suite
// permanently broken.

test.describe("/account password change end-to-end", () => {
  // Use the developer seed; switching passwords for a single role keeps
  // the matrix small and lets the cleanup step restore the seed.
  const original = SEEDED.developer.password;
  const rotated = `Rotated-${Date.now()}!`;

  // Unskipped in PR-T3.5 once PR-L4 introduced the backend proxy
  // `POST /api/v1/account/password`. The proxy verifies current_password
  // by running a fresh Kratos api-mode login on the user's behalf, opens
  // the settings flow with the resulting session_token, and submits the
  // new password — all server-side, so the api-mode login path stays
  // cookie-free.
  test("changes password, signs out, and re-authenticates with the new password", async ({ page }) => {
    let passwordRotated = false;
    try {
      // 1) log in with the seeded password
      await loginAs(page, SEEDED.developer);

      // 2) go to /account and submit the password change form
      await page.goto("/account");
      await page.getByLabel(/current password/i).fill(original);
      await page.getByLabel(/^new password$/i).fill(rotated);
      await page.getByLabel(/confirm new password/i).fill(rotated);
      await page.getByRole("button", { name: /save changes/i }).click();
      await expect(page.getByText(/password updated successfully/i)).toBeVisible({ timeout: 15_000 });
      passwordRotated = true;

      // 3) sign out — Hydra session must end so /login asks for credentials
      await page.getByText(SEEDED.developer.user_id, { exact: false }).first().click();
      await page.getByRole("button", { name: /sign out/i }).click();

      // 4) log in with the new password
      await loginAs(page, { ...SEEDED.developer, password: rotated });
    } finally {
      if (!passwordRotated) return;
      // Best-effort rollback so other specs in the same run reuse the seed.
      // globalSetup (PR-T3.5 hardening) force-resets on the next invocation,
      // so a thrown rollback is logged and swallowed rather than failing the
      // test on top of whatever failed inside the try block.
      try {
        await page.goto("/account");
        await page.getByLabel(/current password/i).fill(rotated);
        await page.getByLabel(/^new password$/i).fill(original);
        await page.getByLabel(/confirm new password/i).fill(original);
        await page.getByRole("button", { name: /save changes/i }).click();
        await expect(page.getByText(/password updated successfully/i)).toBeVisible({ timeout: 15_000 });
      } catch (err) {
        console.warn("[password-change] best-effort rollback failed (globalSetup will recover on next run):", err);
      }
    }
  });
});
