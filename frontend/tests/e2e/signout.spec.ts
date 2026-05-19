import { test, expect, loginAs, SEEDED, submitSignInForm, waitForSignInForm } from "./fixtures";

// signout.spec — Sign Out drives IdP end-session endpoint via id_token_hint
// and the next /login attempt must prompt for credentials again.
//
// 2026-05-12 (claude/login_usermanagement_finish): added
// TC-AUTH-SIGNOUT-REDIR-01 (post-signout AuthGuard redirect) and
// TC-USER-SWITCH-01 (clean user switch across signout boundary).

async function waitForSessionCleared(page: import("@playwright/test").Page): Promise<void> {
  await expect
    .poll(
      async () =>
        page.evaluate(async () => {
          const resp = await fetch("/api/v1/me", { credentials: "include", cache: "no-store" });
          return resp.status;
        }),
      { timeout: 20_000, intervals: [250, 500, 1000] }
    )
    .toBe(401);
}

test.describe("Sign Out terminates IdP session", () => {
  test("after Sign Out, /login flow asks for credentials again", async ({ page }) => {
    await loginAs(page, SEEDED.developer);

    // Open the header dropdown and click Sign Out
    await page.getByText(SEEDED.developer.user_id, { exact: false }).first().click();
    await page.getByRole("button", { name: /sign out/i }).click();
    await waitForSessionCleared(page);

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
    // Redirect chain timing can differ by environment; the important
    // assertion is that the credential form is shown again (no silent auth).
    await waitForSignInForm(page);

    // The password field must be empty — no auto-completion of identity
    await expect(page.locator("input#password, input[name='password']").first()).toHaveValue("");
  });

  test("TC-AUTH-SIGNOUT-REDIR-01 — direct navigation to a protected route after Sign Out bounces to /login", async ({ page }) => {
    await loginAs(page, SEEDED.developer);

    // Sign Out via header dropdown
    await page.getByText(SEEDED.developer.user_id, { exact: false }).first().click();
    await page.getByRole("button", { name: /sign out/i }).click();
    await waitForSessionCleared(page);

    // After the post_logout_redirect resolves, try going back into a
    // guarded route directly. AuthGuard's whoAmI() must see no session
    // and route to /login → OIDC dance → sign-in form. ERR_ABORTED
    // tolerated for the same reason as auth.spec / signout.spec above:
    // /login does a client-side window.location.assign to OIDC authorize.
    await page.goto("/developer").catch((err) => {
      const msg = err instanceof Error ? err.message : String(err);
      if (!msg.includes("ERR_ABORTED")) throw err;
    });
    await waitForSignInForm(page);
  });
});

test.describe("user switch across Sign Out", () => {
  test("TC-USER-SWITCH-01 — Sign Out from alice and Sign In as bob shows bob's profile, never alice's", async ({ page }) => {
    test.setTimeout(90_000);

    // 1) alice 로 로그인 후 /account 의 actor.login 이 alice
    await loginAs(page, SEEDED.developer);
    await page.goto("/account");
    await expect(page.getByText(SEEDED.developer.user_id, { exact: true }).first()).toBeVisible({ timeout: 10_000 });

    // 2) Sign Out
    await page.getByText(SEEDED.developer.user_id, { exact: false }).first().click();
    await page.getByRole("button", { name: /sign out/i }).click();
    await waitForSessionCleared(page);

    // 3) bob (manager) 로 로그인 — signout 직후 loginAs 의 goto('/login') 에서
    // ERR_ABORTED 가 발생할 수 있으므로, 직접 OIDC dance 를 수행한다.
    await page.goto("/login").catch((err) => {
      const msg = err instanceof Error ? err.message : String(err);
      if (!msg.includes("ERR_ABORTED")) throw err;
    });
    await waitForSignInForm(page);
    await submitSignInForm(page, SEEDED.manager.email, SEEDED.manager.password);
    await page.waitForURL(/\/(manager|admin|developer)/, { timeout: 30_000 });

    // 4) /account 의 사용자 정보가 bob, alice 의 잔재 없음
    await page.goto("/account");
    await expect(page.getByText(SEEDED.manager.user_id, { exact: true }).first()).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText(`${SEEDED.manager.user_id}@example.com`)).toBeVisible();
    // alice 의 user_id 가 어떤 곳에도 노출되지 않아야 한다 — actor 가
    // bob 인데 alice 가 보이면 store/UI 가 깨끗히 리셋되지 않은 증거.
    await expect(page.getByText(SEEDED.developer.user_id, { exact: true })).toHaveCount(0);
  });
});
