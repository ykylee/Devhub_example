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
//      backend status 분기에 따라 두 갈래:
//        a) 204/401    — OIDC end_session_endpoint (post_logout_redirect_uri=/login)
//        b) 502/기타/network — OIDC skip, toast, 강제 /login redirect (#488 spec "정합 우선")
//      두 경로 모두 최종적으로 /login 페이지에 도착해야 한다.
// 검증:
//   (필수) logout 후 /login URL 도달 — spec 의 모든 backend status 분기에서 정합.
//   (조건부) 204/401 분기에서는 추가로 OIDC end_session_endpoint 가 post_logout_redirect_uri
//            를 /login 으로 설정해 호출됐는지 확인 — 502/network 분기는 OIDC 자체가
//            unreachable 할 수 있어 호출이 일어나지 않을 수 있으므로 검증 skip.
//
// PR #497 hotfix (2026-06-08 22:00, yklee + Mavis):
//   CI 에서 backend 가 Keycloak 502 를 정상 응답해 (E2E 환경 Keycloak flake) test 가
//   30s 안에 /login 도달하지 못해 flake. test 의 "최종 도착" 만 검증하도록 단순화.
//   ERR_ABORTED 가 아닌 다른 navigation interrupt (예: 502 분기에서 곧장 /login
//   으로 가는 case) 도 swallow 하도록 catch-all. 단, goto 자체가 resolve 한 경우엔
//   (즉 /login 도착 후) swallow 불필요.
test.describe("logout flow (N-8 / P1-6 frontend carve)", () => {
  test("TC-E2E-LOGOUT-01 — full logout eventually lands on /login (OIDC redirect or 502 fallback)", async ({ page }) => {
    // 1) login
    await loginAs(page, SEEDED.developer);
    await expect(page).toHaveURL(/\/developer(\/|$)/);

    // 2) Hit /auth/signout — logout flow 는 backend 응답 후 window.location.assign 으로
    //    navigate. ERR_ABORTED (Next 16 turbopack + OIDC authorize client-side redirect)
    //    또는 OIDC skip + /login 강제 redirect 로 goto 가 interrupt 됨. 모든 navigation
    //    interrupt 를 swallow — test 의 관심사는 "최종 URL" 이지 중간 abort 가 아님.
    //
    //    page.goto 가 throw 하지 않고 resolve 하는 경우도 있음 (예: backend 502 분기에서
    //    곧장 /login 으로 가는 경우, page.url() == /login 이라 goto 자체는 success). 그
    //    경우엔 catch 가 호출되지 않음 — OK.
    await page.goto(appPath("/auth/signout")).catch(() => {
      // navigation interrupted by client-side window.location.assign — expected.
    });

    // 3) (선택) OIDC end_session_endpoint 가 호출됐는지 — 204/401 분기에서만 발생.
    //    502/network 분기에서는 OIDC skip 되므로 null 가능. test 통과에는 영향 없음.
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

    // 4) (필수) logout 후 /login 페이지에 도착. backend status 4종 (204/401/502/network)
    //    모두에서 동일하게 /login 으로 도착해야 spec 정합 (#488 §"Frontend").
    //    CI 환경 Keycloak flake 시 502 가 정상 응답될 수 있어 timeout 여유 30s 유지.
    await expect(page).toHaveURL(/\/login(\?|$|\/)/, { timeout: 30_000 });
  });
});
