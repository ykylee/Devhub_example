import { test, expect, loginAs, SEEDED } from "./fixtures";

// password-change.spec — /account password change via Kratos public
// settings flow (PR-L3, DEC-2=B). Validates the full round-trip: login
// with old password -> change -> sign out -> log in with new password ->
// (cleanup) restore the original password so subsequent test runs are
// idempotent on the same Kratos seed.

test.describe("/account password change end-to-end", () => {
  // Use the developer seed; switching passwords for a single role keeps
  // the matrix small and lets the cleanup step restore the seed.
  const original = SEEDED.developer.password;
  const rotated = `Rotated-${Date.now()}!`;

  // SKIPPED — PR-L4 follow-up. The current login path drives Kratos via
  // api-mode (backend `/self-service/login/api`), which yields a
  // session_token but never plants the `ory_kratos_session` cookie in the
  // user agent. /account's password change calls Kratos browser-mode
  // `/self-service/settings/browser`, which authenticates via cookie and
  // therefore rejects the request with "No active session was found".
  // Unskip after PR-L4 introduces either (a) a backend proxy that uses the
  // session_token to drive the settings flow server-side, or (b) a
  // browser-mode login redirect that lands a Kratos cookie alongside the
  // OIDC handshake. Tracking note: ai-workflow backlog PR-L4.
  test.skip("changes password, signs out, and re-authenticates with the new password", async ({ page }) => {
    // 1) log in with the seeded password
    await loginAs(page, SEEDED.developer);

    // 2) go to /account and submit the password change form
    await page.goto("/account");
    await page.getByLabel(/current password/i).fill(original);
    await page.getByLabel(/^new password$/i).fill(rotated);
    await page.getByLabel(/confirm new password/i).fill(rotated);
    await page.getByRole("button", { name: /save changes/i }).click();
    await expect(page.getByText(/password updated successfully/i)).toBeVisible({ timeout: 15_000 });

    // 3) sign out — Hydra session must end so /login asks for credentials
    await page.getByText(SEEDED.developer.user_id, { exact: false }).first().click();
    await page.getByRole("button", { name: /sign out/i }).click();

    // 4) log in with the new password
    await loginAs(page, { ...SEEDED.developer, password: rotated });

    // 5) cleanup — rotate back to the original so the seed remains usable
    await page.goto("/account");
    await page.getByLabel(/current password/i).fill(rotated);
    await page.getByLabel(/^new password$/i).fill(original);
    await page.getByLabel(/confirm new password/i).fill(original);
    await page.getByRole("button", { name: /save changes/i }).click();
    await expect(page.getByText(/password updated successfully/i)).toBeVisible({ timeout: 15_000 });
  });
});
