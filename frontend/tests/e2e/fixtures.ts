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
  await page.waitForURL(/\/auth\/login\?login_challenge=/, { timeout: 30_000 });

  // /auth/login asks for the System ID (DevHub users.user_id), not the
  // email — Kratos resolves credentials via metadata_public.user_id. The
  // label is "System ID" + the input is wired with htmlFor=identifier.
  await page.getByLabel(/system id/i).fill(user.user_id);
  await page.getByLabel(/^password$/i).fill(user.password);
  await page.getByRole("button", { name: /sign in/i }).click();

  // After successful login the OIDC callback fires and AuthGuard +
  // role-routing land the user on their default page.
  await page.waitForURL(new RegExp(`${user.landing}(/|$)`), { timeout: 30_000 });
}

/**
 * Asserts the current header shows the supplied user as the active actor.
 * The Header (frontend/components/layout/Header.tsx) renders actor.login
 * inside the avatar block.
 */
export async function expectActorIs(page: Page, user: SeededUser) {
  await expect(page.getByText(user.user_id, { exact: false }).first()).toBeVisible({ timeout: 10_000 });
}

// Kratos admin API helpers — used by spec files that need to drive
// identity lifecycle outside the normal browser flow (signup cleanup,
// audit target_id matching). KRATOS_ADMIN_URL env mirrors the one used
// by global-setup.ts; defaults to localhost:4434.
const KRATOS_ADMIN_URL = (process.env.KRATOS_ADMIN_URL ?? "http://localhost:4434").replace(/\/$/, "");

interface KratosIdentityLite {
  id: string;
  traits?: { email?: string; [k: string]: unknown };
}

/** Walks Kratos /admin/identities pagination to find the identity whose
 *  traits.email matches (case-insensitive). Returns the identity.id or
 *  null when no match exists. */
export async function getKratosIdentityIdByEmail(email: string): Promise<string | null> {
  const needle = email.trim().toLowerCase();
  let pageNo = 0;
  const perPage = 250;
  while (pageNo < 40) {
    const url = `${KRATOS_ADMIN_URL}/admin/identities?page=${pageNo}&per_page=${perPage}`;
    const resp = await fetch(url, { headers: { Accept: "application/json" } });
    if (!resp.ok) {
      throw new Error(`Kratos admin list identities ${resp.status}: ${await resp.text()}`);
    }
    const batch = (await resp.json()) as KratosIdentityLite[];
    for (const ident of batch) {
      if (ident.traits?.email?.toLowerCase() === needle && ident.id) {
        return ident.id;
      }
    }
    if (batch.length < perPage) break;
    pageNo += 1;
  }
  return null;
}

/** Best-effort cleanup helper for spec files that create a new Kratos
 *  identity (currently only signup.spec). Silent no-op when the identity
 *  cannot be found — the next spec run will fail loudly if the leak
 *  actually matters. 404 from the DELETE is also tolerated for the same
 *  reason (already cleaned by a previous attempt). */
export async function deleteKratosIdentityByEmail(email: string): Promise<void> {
  const id = await getKratosIdentityIdByEmail(email);
  if (!id) return;
  const resp = await fetch(`${KRATOS_ADMIN_URL}/admin/identities/${id}`, {
    method: "DELETE",
    headers: { Accept: "application/json" },
  });
  if (!resp.ok && resp.status !== 404) {
    throw new Error(`Kratos admin delete identity ${id} (${email}) → ${resp.status}: ${await resp.text()}`);
  }
}

export const test = base.extend({});
export { expect };
