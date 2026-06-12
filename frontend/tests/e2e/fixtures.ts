import { test as base, expect, type Page } from "@playwright/test";

// Test fixtures + helpers for the e2e suite (PR-T3, work_26_05_11-d).
//
// Seeded users — these must already exist in IdP identity store (with
// metadata_public.user_id matching) and in DevHub `users`. The e2e guide
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
export const SEEDED: Record<"developer" | "team_manager" | "systemAdmin", SeededUser> = {
  developer: {
    user_id: "alice",
    email: "alice@example.com",
    password: "ChangeMe-12345!",
    role: "developer",
    landing: "/developer",
  },
  team_manager: {
    user_id: "bob",
    email: "bob@example.com",
    password: "ChangeMe-12345!",
    role: "team_manager",
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

const E2E_BASE_PATH = (process.env.PLAYWRIGHT_BASE_PATH ?? "").trim().replace(/\/+$/, "");
const KC_BASE_URL = (
  process.env.DEVHUB_E2E_KEYCLOAK_ADMIN_URL
  ?? process.env.DEVHUB_KEYCLOAK_ADMIN_URL
  ?? process.env.KEYCLOAK_URL
  ?? "http://localhost:8180/devhub/auth/keycloak"
).replace(/\/+$/, "");
const KC_REALM = (process.env.DEVHUB_KEYCLOAK_ADMIN_REALM ?? "devhub").trim();
const OIDC_CLIENT_ID = (
  process.env.DEVHUB_E2E_OIDC_CLIENT_ID
  ?? process.env.NEXT_PUBLIC_OIDC_CLIENT_ID
  ?? process.env.DEVHUB_OIDC_CLIENT_ID
  ?? "devhub-frontend"
).trim();

export function appPath(path: string): string {
  const normalized = path.startsWith("/") ? path : `/${path}`;
  return E2E_BASE_PATH ? `${E2E_BASE_PATH}${normalized}` : normalized;
}

export function apiPath(path: string): string {
  return appPath(path);
}

async function firstVisibleLocator(page: Page, selectors: string[]): Promise<import("@playwright/test").Locator> {
  for (const selector of selectors) {
    const locator = page.locator(selector).first();
    if ((await locator.count()) > 0 && (await locator.isVisible().catch(() => false))) {
      return locator;
    }
  }
  throw new Error(`none of selectors are visible: ${selectors.join(", ")}`);
}

function isAppLoginURL(rawURL: string): boolean {
  try {
    const url = new URL(rawURL);
    const loginPath = appPath("/login");
    return url.pathname === loginPath || url.pathname === `${loginPath}/`;
  } catch {
    return false;
  }
}

type WaitForSignInFormOptions = {
  restartOIDCOnAppLogin?: boolean;
  /** Deadline in milliseconds. Default = 30_000 (CI race 환경에서 intermittent 30s 도 fail 가능 → 45_000 권장). */
  timeoutMs?: number;
};

async function restartOIDCFromLoginPage(page: Page): Promise<"clicked" | "reloaded" | "noop"> {
  const continueButton = page.getByRole("button", { name: /continue to sign in|redirecting/i }).first();
  if ((await continueButton.count()) === 0 || !(await continueButton.isVisible().catch(() => false))) {
    return "noop";
  }

  const isDisabled = await continueButton.isDisabled().catch(() => false);
  if (!isDisabled) {
    await continueButton.click().catch((err) => {
      const msg = err instanceof Error ? err.message : String(err);
      if (
        msg.includes("Target page, context or browser has been closed")
        || msg.includes("Execution context was destroyed")
      ) {
        return;
      }
      throw err;
    });
    return "clicked";
  }

  await page.reload({ waitUntil: "domcontentloaded" }).catch(() => {});
  return "reloaded";
}

export async function waitForSignInForm(page: Page, options: WaitForSignInFormOptions = {}): Promise<void> {
  const deadline = Date.now() + (options.timeoutMs ?? (process.env.CI ? 45_000 : 30_000));
  let restartCount = 0;
  const maxRestarts = options.restartOIDCOnAppLogin ? 3 : 0;
  let stuckLoginLoops = 0;
  while (Date.now() < deadline) {
    const userVisible = await page.locator('input#username, input[name="username"], input#identifier, input[name="identifier"]').first().isVisible().catch(() => false);
    const passVisible = await page.locator('input#password, input[name="password"]').first().isVisible().catch(() => false);
    if (userVisible && passVisible) return;

    const backToSignInButton = page.getByRole("button", { name: /back to sign in/i }).first();
    if ((await backToSignInButton.count()) > 0 && (await backToSignInButton.isVisible().catch(() => false))) {
      await backToSignInButton.click();
      await page.waitForTimeout(250);
      continue;
    }

    if (isAppLoginURL(page.url())) {
      const action = await restartOIDCFromLoginPage(page);
      if (action !== "noop") {
        stuckLoginLoops = 0;
        await page.waitForTimeout(250);
        continue;
      }
    }

    if (restartCount < maxRestarts && isAppLoginURL(page.url())) {
      // Sign-out flows can occasionally land back on the app's /login page
      // without resuming the redirect chain. Restart through the page's own
      // login CTA / page handler so the app generates a fresh PKCE state
      // instead of mutating sessionStorage behind its back.
      restartCount += 1;
      stuckLoginLoops = 0;
      await restartOIDCFromLoginPage(page);
      await page.waitForTimeout(250);
      continue;
    }

    if (isAppLoginURL(page.url())) {
      stuckLoginLoops += 1;
      // If the app login page stays visible without rendering the IdP form or
      // progressing the redirect chain, periodically hard-refresh to clear a
      // stale "Redirecting..." client state.
      if (stuckLoginLoops >= 4) {
        stuckLoginLoops = 0;
        await page.reload({ waitUntil: "domcontentloaded" }).catch(() => {});
        await page.waitForTimeout(250);
        continue;
      }
    }
    await page.waitForTimeout(500);
  }
  throw new Error("sign-in form was not shown within timeout");
}

export async function submitSignInForm(page: Page, userID: string, password: string): Promise<void> {
  const userInput = await firstVisibleLocator(page, [
    'input#identifier',
    'input[name="identifier"]',
    'input#username',
    'input[name="username"]',
  ]);
  await userInput.fill(userID);

  const passwordInput = await firstVisibleLocator(page, ['input#password', 'input[name="password"]']);
  await passwordInput.fill(password);

  const keycloakSubmit = page.locator('input#kc-login, input[type="submit"]#kc-login').first();
  if ((await keycloakSubmit.count()) > 0 && (await keycloakSubmit.isVisible().catch(() => false))) {
    await keycloakSubmit.click();
    return;
  }
  await page.getByRole("button", { name: /sign in|log in|login/i }).first().click();
}

export async function completeKeycloakRequiredActionsIfPresent(page: Page): Promise<void> {
  const updateHeading = page.getByRole("heading", { name: /update account information/i }).first();
  if ((await updateHeading.count()) === 0 || !(await updateHeading.isVisible().catch(() => false))) {
    return;
  }

  const firstName = page.getByRole("textbox", { name: /first name/i }).first();
  const lastName = page.getByRole("textbox", { name: /last name/i }).first();

  if ((await firstName.count()) > 0) {
    const value = await firstName.inputValue().catch(() => "");
    if (!value.trim()) {
      await firstName.fill("E2E");
    }
  }
  if ((await lastName.count()) > 0) {
    const value = await lastName.inputValue().catch(() => "");
    if (!value.trim()) {
      await lastName.fill("User");
    }
  }

  const submit = page.getByRole("button", { name: /submit/i }).first();
  if ((await submit.count()) > 0 && (await submit.isVisible().catch(() => false))) {
    await submit.click();
  }
}

export async function loginAs(page: Page, user: SeededUser) {
  await page.goto(appPath("/login")).catch((err) => {
    const msg = err instanceof Error ? err.message : String(err);
    if (!msg.includes("ERR_ABORTED")) throw err;
  });
  await waitForSignInForm(page, { restartOIDCOnAppLogin: true });
  await submitSignInForm(page, user.email, user.password);
  await completeKeycloakRequiredActionsIfPresent(page);
  await page.waitForURL(new RegExp(`${user.landing}(/|$)`), { timeout: process.env.CI ? 60_000 : 30_000 });
}

export async function openHeaderUserMenu(page: Page, user: SeededUser): Promise<void> {
  const byAria = page.getByRole("button", { name: /user menu/i }).first();
  if ((await byAria.count()) > 0 && (await byAria.isVisible().catch(() => false))) {
    await byAria.click();
    return;
  }

  const byClickableLogin = page
    .locator("header [class*='cursor-pointer']")
    .filter({ hasText: user.user_id })
    .first();
  if ((await byClickableLogin.count()) > 0 && (await byClickableLogin.isVisible().catch(() => false))) {
    await byClickableLogin.click();
    return;
  }

  const byBannerText = page.getByRole("banner").getByText(user.user_id, { exact: false }).first();
  if ((await byBannerText.count()) > 0 && (await byBannerText.isVisible().catch(() => false))) {
    await byBannerText.click();
    return;
  }

  throw new Error(`failed to open header user menu for user_id=${user.user_id}`);
}

/**
 * Asserts the current header shows the supplied user as the active actor.
 * The Header (frontend/components/layout/Header.tsx) renders actor.login
 * inside the avatar block.
 */
export async function expectActorIs(page: Page, user: SeededUser) {
  await expect(page.getByText(user.user_id, { exact: false }).first()).toBeVisible({ timeout: 10_000 });
}

const KC_ADMIN_CLIENT_ID = (process.env.DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID ?? "devhub-backend").trim();
const KC_ADMIN_CLIENT_SECRET = (process.env.DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET ?? "secret-change-me-backend").trim();

interface KeycloakUserLite {
  id: string;
  email?: string;
}

async function fetchKeycloakAdminToken(): Promise<string> {
  if (!KC_ADMIN_CLIENT_SECRET) {
    throw new Error("DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET is required for keycloak e2e helpers");
  }
  const tokenURL = `${KC_BASE_URL}/realms/${encodeURIComponent(KC_REALM)}/protocol/openid-connect/token`;
  const resp = await fetch(tokenURL, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      grant_type: "client_credentials",
      client_id: KC_ADMIN_CLIENT_ID,
      client_secret: KC_ADMIN_CLIENT_SECRET,
    }),
  });
  if (!resp.ok) {
    throw new Error(`Keycloak token grant failed ${resp.status}: ${await resp.text()}`);
  }
  const body = (await resp.json()) as { access_token?: string };
  if (!body.access_token) throw new Error("Keycloak token response missing access_token");
  return body.access_token;
}

export async function getKeycloakUserIdByEmail(email: string): Promise<string | null> {
  const token = await fetchKeycloakAdminToken();
  const url = `${KC_BASE_URL}/admin/realms/${encodeURIComponent(KC_REALM)}/users?email=${encodeURIComponent(email)}&exact=true`;
  const resp = await fetch(url, {
    headers: { Authorization: `Bearer ${token}`, Accept: "application/json" },
  });
  if (!resp.ok) {
    throw new Error(`Keycloak user lookup ${email} failed ${resp.status}: ${await resp.text()}`);
  }
  const users = (await resp.json()) as KeycloakUserLite[];
  return users[0]?.id ?? null;
}

export async function deleteKeycloakUserByEmail(email: string): Promise<void> {
  const token = await fetchKeycloakAdminToken();
  const id = await getKeycloakUserIdByEmail(email);
  if (!id) return;
  const resp = await fetch(`${KC_BASE_URL}/admin/realms/${encodeURIComponent(KC_REALM)}/users/${encodeURIComponent(id)}`, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${token}`, Accept: "application/json" },
  });
  if (!resp.ok && resp.status !== 404) {
    throw new Error(`Keycloak user delete ${email} failed ${resp.status}: ${await resp.text()}`);
  }
}

export const test = base.extend({});
export { expect };
