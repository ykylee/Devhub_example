import { test, expect, loginAs, SEEDED } from "./fixtures";

// signout.spec — Sign Out drives Hydra /oauth2/sessions/logout via
// id_token_hint (PR-L2) and the next /login attempt must prompt for the
// password again (Hydra session terminated).

test.describe("Sign Out terminates Hydra session", () => {
  test("after Sign Out, /login flow asks for credentials again", async ({ page }) => {
    await loginAs(page, SEEDED.developer);

    // Open the header dropdown and click Sign Out
    await page.getByText(SEEDED.developer.user_id, { exact: false }).first().click();
    await page.getByRole("button", { name: /sign out/i }).click();

    // After redirects we should be back at / (post_logout_redirect_uri)
    // or /login. Either way, /login should kick off a fresh OIDC dance
    // that lands at the password form again — not silent re-auth.
    //
    // The OIDC redirect chain (client-side window.location.assign on the
    // /login page) aborts the original goto navigation under Next 16
    // turbopack dev mode; the page actually arrives at the form, but
    // page.goto rejects with ERR_ABORTED. We swallow that specific abort
    // and let the subsequent waitForURL prove the redirect landed.
    await page.goto("/login").catch((err) => {
      const msg = err instanceof Error ? err.message : String(err);
      if (!msg.includes("ERR_ABORTED")) throw err;
    });
    await page.waitForURL(/\/auth\/login\?login_challenge=/, { timeout: 15_000 });

    // The password field must be empty — no auto-completion of identity
    await expect(page.getByLabel(/password/i)).toHaveValue("");
  });
});
