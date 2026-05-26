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

async function firstVisibleLocator(page: Page, selectors: string[]): Promise<import("@playwright/test").Locator> {
  for (const selector of selectors) {
    const locator = page.locator(selector).first();
    if ((await locator.count()) > 0 && (await locator.isVisible().catch(() => false))) {
      return locator;
    }
  }
  throw new Error(`none of selectors are visible: ${selectors.join(", ")}`);
}

async function forceStartOIDCFlow(page: Page): Promise<void> {
  const keycloakAdminURL = KC_BASE_URL;
  const keycloakRealm = KC_REALM;
  const oidcClientID = OIDC_CLIENT_ID;
  const runtimeConfigPath = appPath("/api/runtime-config");
  await page.evaluate(async ({ keycloakAdminURL, keycloakRealm, runtimeConfigPath, oidcClientID }) => {
    const authURL = `${keycloakAdminURL}/realms/${encodeURIComponent(keycloakRealm)}/protocol/openid-connect/auth`;
    let redirectURI = `${window.location.origin}/auth/callback`;
    try {
      const runtimeResp = await fetch(runtimeConfigPath, { cache: "no-store" });
      if (runtimeResp.ok) {
        const runtimeBody = (await runtimeResp.json()) as { oidc_redirect_uri?: string };
        if (runtimeBody.oidc_redirect_uri?.trim()) {
          redirectURI = runtimeBody.oidc_redirect_uri.trim();
        }
      }
    } catch {
      // Keep default fallback when runtime-config is unavailable.
    }

    const randomBytes = (n: number) => {
      const b = new Uint8Array(n);
      crypto.getRandomValues(b);
      return b;
    };
    const b64u = (arr: Uint8Array) =>
      btoa(String.fromCharCode(...arr)).replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, "");
    const verifier = b64u(randomBytes(32));
    const state = crypto.randomUUID ? crypto.randomUUID() : b64u(randomBytes(16));
    const hash = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(verifier));
    const challenge = b64u(new Uint8Array(hash));

    sessionStorage.setItem("oidc_state", state);
    sessionStorage.setItem("oidc_verifier", verifier);

    const url = new URL(authURL);
    url.searchParams.set("client_id", oidcClientID);
    url.searchParams.set("response_type", "code");
    url.searchParams.set("redirect_uri", redirectURI);
    url.searchParams.set("scope", "openid offline_access email profile");
    url.searchParams.set("state", state);
    url.searchParams.set("code_challenge", challenge);
    url.searchParams.set("code_challenge_method", "S256");
    window.location.assign(url.toString());
  }, { keycloakAdminURL, keycloakRealm, runtimeConfigPath, oidcClientID });
}

export async function waitForSignInForm(page: Page): Promise<void> {
  const deadline = Date.now() + 30_000;
  let forcedOIDC = false;
  while (Date.now() < deadline) {
    const userVisible = await page.locator('input#username, input[name="username"], input#identifier, input[name="identifier"]').first().isVisible().catch(() => false);
    const passVisible = await page.locator('input#password, input[name="password"]').first().isVisible().catch(() => false);
    if (userVisible && passVisible) return;

    const continueButton = page.getByRole("button", { name: /continue to sign in/i }).first();
    if ((await continueButton.count()) > 0 && (await continueButton.isVisible().catch(() => false))) {
      await continueButton.click();
      if (!forcedOIDC) {
        forcedOIDC = true;
        await forceStartOIDCFlow(page).catch(() => {});
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

async function completeKeycloakRequiredActionsIfPresent(page: Page): Promise<void> {
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
  await waitForSignInForm(page);
  await submitSignInForm(page, user.email, user.password);
  await completeKeycloakRequiredActionsIfPresent(page);
  await page.waitForURL(new RegExp(`${user.landing}(/|$)`), { timeout: 30_000 });
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
