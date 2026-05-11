import { test as base, expect, type Page } from "@playwright/test";

// Test fixtures + helpers for the e2e suite (PR-T3, work_26_05_11-d).
//
// Seeded users — these must already exist in Kratos (with metadata_public.
// user_id matching) and in DevHub `users`. The e2e guide
// (docs/setup/e2e-test-guide.md) walks operators through the seeding step.

export const SEEDED = {
  developer: {
    user_id: "alice",
    email: "alice@example.com",
    password: "ChangeMe-12345!",
    role: "developer",
    landing: "/developer",
  },
  manager: {
    user_id: "bob",
    email: "bob@example.com",
    password: "ChangeMe-12345!",
    role: "manager",
    landing: "/manager",
  },
  systemAdmin: {
    user_id: "charlie",
    email: "charlie@example.com",
    password: "ChangeMe-12345!",
    role: "system_admin",
    landing: "/admin",
  },
} as const;

export type SeededUser = (typeof SEEDED)[keyof typeof SEEDED];

/**
 * Drives the login form at /auth/login. The caller is responsible for
 * starting from /login (which redirects through Hydra and lands the
 * browser on /auth/login?login_challenge=...).
 */
export async function loginAs(page: Page, user: SeededUser) {
  await page.goto("/login");
  // The /login route auto-triggers the OIDC dance; we end up at the
  // Kratos-backed /auth/login form.
  await page.waitForURL(/\/auth\/login\?login_challenge=/, { timeout: 15_000 });

  await page.getByLabel(/email/i).fill(user.email);
  await page.getByLabel(/password/i).fill(user.password);
  await page.getByRole("button", { name: /sign in/i }).click();

  // After successful login the OIDC callback fires and AuthGuard +
  // role-routing land the user on their default page.
  await page.waitForURL(new RegExp(`${user.landing}(/|$)`), { timeout: 15_000 });
}

/**
 * Asserts the current header shows the supplied user as the active actor.
 * The Header (frontend/components/layout/Header.tsx) renders actor.login
 * inside the avatar block.
 */
export async function expectActorIs(page: Page, user: SeededUser) {
  await expect(page.getByText(user.user_id, { exact: false }).first()).toBeVisible({ timeout: 10_000 });
}

export const test = base.extend({});
export { expect };
