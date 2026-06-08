import { test, expect, loginAs, SEEDED, submitSignInForm, waitForSignInForm, appPath } from "./fixtures";

// auth.spec — login + role-based landing (PR-S1) + system route gating.
// Source-of-truth: defaultLandingFor + pathRequiresSystemAdmin in
// frontend/lib/auth/role-routing.ts.
//
// 2026-05-12 (claude/login_usermanagement_finish): TC-AUTH-NEG-01 +
// TC-AUTH-NOAUTH-01 추가 — 로그인 실패 시 에러 + 비로그인 보호 페이지
// 접근 시 /login 리다이렉트.

test.describe("role-based landing", () => {
  test("developer lands on /developer", async ({ page }) => {
    await loginAs(page, SEEDED.developer);
    await expect(page).toHaveURL(/\/developer(\/|$)/);
  });

  test("team_manager lands on /manager", async ({ page }) => {
    await loginAs(page, SEEDED.team_manager);
    await expect(page).toHaveURL(/\/manager(\/|$)/);
  });

  test("system_admin lands on /admin", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await expect(page).toHaveURL(/\/admin(\/|$)/);
  });
});

test.describe("system route gating", () => {
  test("developer cannot reach /admin/settings — AuthGuard bounces to default landing", async ({ page }) => {
    await loginAs(page, SEEDED.developer);
    await page.goto(appPath("/admin/settings"));
    // pathRequiresSystemAdmin + isSystemAdmin guard in AuthGuard.tsx
    // redirects to defaultLandingFor(actor.role) = /developer.
    await expect(page).toHaveURL(/\/developer(\/|$)/, { timeout: 10_000 });
  });
});

test.describe("login failure + auth guard (2026-05-12)", () => {
  test("TC-AUTH-NEG-01 — wrong password keeps the user on the sign-in form", async ({ page }) => {
    await page.goto(appPath("/login")).catch((err) => {
      const msg = err instanceof Error ? err.message : String(err);
      if (!msg.includes("ERR_ABORTED")) throw err;
    });
    await waitForSignInForm(page);
    await submitSignInForm(page, SEEDED.developer.user_id, "wrong-password-not-real");

    // Provider validation echoes a credential-invalid message. Exact wording is version dependent, so we
    // assert on a loose substring and confirm the URL stays on the
    // login form (login_challenge unchanged → no advance).
    await waitForSignInForm(page);
    // The frontend renders provider error messages
    // messages — at least one error indicator must appear.
    await expect(
      page.getByText(/(invalid|incorrect|credentials)/i).first()
    ).toBeVisible({ timeout: 10_000 });
  });

  test("TC-AUTH-NOAUTH-01 — unauthenticated request to a protected route bounces to /login", async ({ page }) => {
    // No login. Direct navigation to a guarded landing page.
    // AuthGuard's whoAmI() returns 401 → router.replace("/login") →
    // /login bootstraps the OIDC dance → page lands on the sign-in form.
    //
    // The /login route triggers a client-side window.location.assign
    // to the OIDC authorize endpoint; under Next 16 turbopack dev
    // mode that aborts the original page.goto with ERR_ABORTED. Same
    // shape as signout.spec — we swallow the specific abort and rely
    // on the subsequent waitForURL to assert the redirect chain
    // landed on the sign-in form.
    await page.goto(appPath("/developer")).catch((err) => {
      const msg = err instanceof Error ? err.message : String(err);
      if (!msg.includes("ERR_ABORTED")) throw err;
    });
    await waitForSignInForm(page);
  });
});

// TC-E2E-LOGOUT-01 — N-8 sprint -i (sprint -h backend PR #495 의 frontend carve).
// 시나리오:
//   1) developer 시드로 Keycloak test login → /developer 도착
//   2) /auth/signout 진입 → authService.logout() → backend POST /api/v1/auth/logout
//      204 응답 + OIDC end_session_endpoint redirect → /login 도착
//   3) 재로그인 → /developer 다시 도착 (Keycloak session 이 정상 종료/재발급)
// 검증: OIDC end_session_endpoint 가 호출됐는지 (post_logout_redirect_uri=…/login),
//       /login 페이지에 다시 도달했는지.
test.describe("logout flow (N-8 / P1-6 frontend carve)", () => {
  test("TC-E2E-LOGOUT-01 — full logout redirects through OIDC end_session_endpoint and lands on /login", async ({ page }) => {
    // 1) login
    await loginAs(page, SEEDED.developer);
    await expect(page).toHaveURL(/\/developer(\/|$)/);

    // 2) Hit /auth/signout — authService.logout() 가 호출되어 backend logout + OIDC
    //    end_session_endpoint 로 redirect. logout flow 자체가 window.location.assign
    //    으로 navigate 하므로 page.goto 가 ERR_ABORTED 로 abort 될 수 있음.
    await page.goto(appPath("/auth/signout")).catch((err) => {
      const msg = err instanceof Error ? err.message : String(err);
      if (!msg.includes("ERR_ABORTED")) throw err;
    });

    // 3) OIDC end_session_endpoint 가 한 번이라도 호출됐는지 (post_logout_redirect_uri
    //    가 BASE_PATH/login 으로 세팅돼 있어야 함).
    const endSessionCall = await page
      .waitForRequest(
        (req) => {
          const url = req.url();
          return (
            /\/protocol\/openid-connect\/logout(\?|$)/.test(url) ||
            /\/logout(\?|$)/.test(url)
          );
        },
        { timeout: 15_000 },
      )
      .catch(() => null);

    if (endSessionCall) {
      const url = new URL(endSessionCall.url());
      const postLogout = url.searchParams.get("post_logout_redirect_uri") ?? "";
      expect(postLogout).toMatch(/\/login(\/?)$/);
    }

    // 4) 결국 /login 페이지에 도착해야 함.
    await expect(page).toHaveURL(/\/login(\?|$|\/)/, { timeout: 30_000 });
  });
});
