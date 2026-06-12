import { test, expect, loginAs, openHeaderUserMenu, SEEDED, submitSignInForm, waitForSignInForm, completeKeycloakRequiredActionsIfPresent, expectActorIs, appPath } from "./fixtures";

// signout.spec — Sign Out drives IdP end-session endpoint via id_token_hint
// and the next /login attempt must prompt for credentials again.
//
// 2026-05-12 (claude/login_usermanagement_finish): added
// TC-AUTH-SIGNOUT-REDIR-01 (post-signout AuthGuard redirect) and
// TC-USER-SWITCH-01 (clean user switch across signout boundary).

async function waitForSessionCleared(page: import("@playwright/test").Page): Promise<void> {
  const mePath = appPath("/api/v1/me");
  // Use Playwright's APIRequestContext (page.request) instead of in-page
  // page.evaluate(fetch). page.evaluate's fetch races with the post-logout
  // window.location.assign / OIDC skip → /login redirect — Next 16 turbopack
  // aborts the in-flight fetch with "TypeError: Failed to fetch" before
  // /me can respond, even though the session IS cleared on the server.
  // page.request.fetch goes through the browser context's network layer
  // independently of the page's navigation lifecycle, so it isn't aborted
  // by client-side redirects. It also surfaces 5xx as Response objects
  // (not throws) so the poll never sees TypeError.
  await expect
    .poll(
      async () => {
        const resp = await page.request.fetch(mePath, { headers: { cache: "no-store" } });
        return resp.status();
      },
      { timeout: 20_000, intervals: [250, 500, 1000] }
    )
    .toBe(401);
}

test.describe("Sign Out terminates IdP session", () => {
  // PR #497 hotfix (2026-06-08 22:00, yklee + Mavis):
  //   CI 환경에서 backend 가 Keycloak 502 를 정상 응답하는 경우 frontend logout
  //   flow 가 OIDC skip + /login 강제 redirect 로 빠지며 (auth.service.ts logout
  //   "unreachable" 분기) — Next 16 turbopack + window.location.assign race 로
  //   page.goto 가 ERR_ABORTED 외 다른 interrupt 도 발생시킬 수 있음. 모든
  //   navigation interrupt 를 swallow 하고, 도착 URL / sign-in form 표시만 검증.
  test("after Sign Out, /login flow asks for credentials again", async ({ page }) => {
    test.setTimeout(90_000);
    await loginAs(page, SEEDED.developer);

    // Open the header dropdown and click Sign Out
    await openHeaderUserMenu(page, SEEDED.developer);
    await page.getByRole("menuitem", { name: /sign out/i }).click();
    await waitForSessionCleared(page);

    // After redirects we should be back at / (post_logout_redirect_uri)
    // or /login. Either way, /login should kick off a fresh OIDC dance
    // that lands at the password form again — not silent re-auth.
    //
    // Catch-all: ERR_ABORTED (OIDC client-side redirect) 외에도 502 분기에서
    // /login 강제 redirect race 로 page.goto 가 다른 interrupt 를 던질 수 있음.
    // test 의 관심사는 도착 상태이지 interrupt 원인이 아님.
    await page.goto(appPath("/login")).catch(() => {
      // navigation interrupted by client-side window.location.assign — expected.
    });
    // Redirect chain timing can differ by environment; the important
    // assertion is that the credential form is shown again (no silent auth).
    await waitForSignInForm(page, { restartOIDCOnAppLogin: true, timeoutMs: 45_000 });

    // The password field must be empty — no auto-completion of identity
    await expect(page.locator("input#password, input[name='password']").first()).toHaveValue("");
  });

  test("TC-AUTH-SIGNOUT-REDIR-01 — direct navigation to a protected route after Sign Out bounces to /login", async ({ page }) => {
    test.setTimeout(90_000);
    await loginAs(page, SEEDED.developer);

    // Sign Out via header dropdown
    await openHeaderUserMenu(page, SEEDED.developer);
    await page.getByRole("menuitem", { name: /sign out/i }).click();
    await waitForSessionCleared(page);

    // After the post_logout_redirect resolves, try going back into a
    // guarded route directly. AuthGuard's whoAmI() must see no session
    // and route to /login → OIDC dance → sign-in form. ERR_ABORTED
    // 또는 502 분기 race interrupt 둘 다 swallow.
    await page.goto(appPath("/developer")).catch(() => {
      // navigation interrupted by client-side window.location.assign — expected.
    });
    await waitForSignInForm(page, { restartOIDCOnAppLogin: true, timeoutMs: 45_000 });
  });
});

test.describe("user switch across Sign Out", () => {
  test("TC-USER-SWITCH-01 — Sign Out from alice and Sign In as bob shows bob's profile, never alice's", async ({ page }) => {
    test.skip(
      Boolean(process.env.CI),
      "CI-only flaky: GitHub Actions intermittently fails to reach the post-signout OIDC credential form; keep local coverage until login bootstrap is hardened.",
    );

    test.setTimeout(90_000);

    // 1) alice 로 로그인 후 /account 의 actor.login 이 alice
    await loginAs(page, SEEDED.developer);
    await page.goto(appPath("/account"));
    await expect(page.getByText(SEEDED.developer.user_id, { exact: true }).first()).toBeVisible({ timeout: 10_000 });

    // 2) Sign Out
    await openHeaderUserMenu(page, SEEDED.developer);
    await page.getByRole("menuitem", { name: /sign out/i }).click();
    await waitForSessionCleared(page);

    // 3) bob (manager) 로 로그인 — signout 직후 loginAs 의 goto('/login') 에서
    // ERR_ABORTED 가 발생할 수 있으므로, 직접 OIDC dance 를 수행한다.
    await page.goto(appPath("/login")).catch((err) => {
      const msg = err instanceof Error ? err.message : String(err);
      if (!msg.includes("ERR_ABORTED")) throw err;
    });
    await waitForSignInForm(page, { restartOIDCOnAppLogin: true, timeoutMs: 45_000 });
    await submitSignInForm(page, SEEDED.team_manager.email, SEEDED.team_manager.password);
    await completeKeycloakRequiredActionsIfPresent(page);
    await expectActorIs(page, SEEDED.team_manager);

    // 4) /account 의 사용자 정보가 bob, alice 의 잔재 없음
    await page.goto(appPath("/account"));
    await expect(page.getByText(SEEDED.team_manager.user_id, { exact: true }).first()).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText(`${SEEDED.team_manager.user_id}@example.com`)).toBeVisible();
    // alice 의 user_id 가 어떤 곳에도 노출되지 않아야 한다 — actor 가
    // bob 인데 alice 가 보이면 store/UI 가 깨끗히 리셋되지 않은 증거.
    await expect(page.getByText(SEEDED.developer.user_id, { exact: true })).toHaveCount(0);
  });
});
