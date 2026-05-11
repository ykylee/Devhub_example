import { test as base, expect, type Page } from "@playwright/test";

// Test fixtures + helpers for the e2e suite (PR-T3, work_26_05_11-d).
//
// Seeded users — these must already exist in Kratos (with metadata_public.
// user_id matching) and in DevHub `users`. The e2e guide
// (docs/setup/e2e-test-guide.md) walks operators through the seeding step.

export type SeededUser = {
  user_id: string;
  email: string;
  password: string;
  role: string;
  landing: string;
};

// SEEDED is intentionally not `as const` — password-change.spec rotates the
// password through `{ ...SEEDED.developer, password: rotated }`, which needs
// password to be a widened `string` rather than the seeded literal.
export const SEEDED: Record<"developer" | "manager" | "systemAdmin", SeededUser> = {
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
};

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

  // /auth/login asks for the System ID (DevHub users.user_id), not the
  // email — Kratos resolves credentials via metadata_public.user_id. The
  // label is "System ID" + the input is wired with htmlFor=identifier.
  await page.getByLabel(/system id/i).fill(user.user_id);
  await page.getByLabel(/^password$/i).fill(user.password);
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
