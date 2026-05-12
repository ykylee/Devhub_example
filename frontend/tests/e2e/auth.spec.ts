import { test, expect, loginAs, SEEDED } from "./fixtures";

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

  test("manager lands on /manager", async ({ page }) => {
    await loginAs(page, SEEDED.manager);
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
    await page.goto("/admin/settings");
    // pathRequiresSystemAdmin + isSystemAdmin guard in AuthGuard.tsx
    // redirects to defaultLandingFor(actor.role) = /developer.
    await expect(page).toHaveURL(/\/developer(\/|$)/, { timeout: 10_000 });
  });
});

test.describe("login failure + auth guard (2026-05-12)", () => {
  test("TC-AUTH-NEG-01 — wrong password keeps the user on the Kratos login form", async ({ page }) => {
    await page.goto("/login");
    await page.waitForURL(/\/auth\/login\?login_challenge=/, { timeout: 15_000 });

    await page.getByLabel(/system id/i).fill(SEEDED.developer.user_id);
    await page.getByLabel(/^password$/i).fill("wrong-password-not-real");
    await page.getByRole("button", { name: /sign in/i }).click();

    // Kratos echoes a credential-invalid message via the Kratos
    // self-service flow. Exact wording is version dependent, so we
    // assert on a loose substring and confirm the URL stays on the
    // login form (login_challenge unchanged → no advance).
    await expect(page).toHaveURL(/\/auth\/login\?login_challenge=/, { timeout: 10_000 });
    // The frontend renders Kratos' returned `ui.messages` / per-node
    // messages — at least one error indicator must appear.
    await expect(
      page.getByText(/(invalid|incorrect|credentials)/i).first()
    ).toBeVisible({ timeout: 10_000 });
  });

  test("TC-AUTH-NOAUTH-01 — unauthenticated request to a protected route bounces to /login", async ({ page }) => {
    // No login. Direct navigation to a guarded landing page.
    // AuthGuard's whoAmI() returns 401 → router.replace("/login") →
    // /login bootstraps the OIDC dance → page lands on the Kratos
    // login form with a login_challenge token.
    //
    // The /login route triggers a client-side window.location.assign
    // to the Hydra authorize endpoint; under Next 16 turbopack dev
    // mode that aborts the original page.goto with ERR_ABORTED. Same
    // shape as signout.spec — we swallow the specific abort and rely
    // on the subsequent waitForURL to assert the redirect chain
    // landed on the Kratos form.
    await page.goto("/developer").catch((err) => {
      const msg = err instanceof Error ? err.message : String(err);
      if (!msg.includes("ERR_ABORTED")) throw err;
    });
    await page.waitForURL(/\/auth\/login\?login_challenge=/, { timeout: 20_000 });
  });
});
