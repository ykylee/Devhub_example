import { test, expect, appPath, waitForSignInForm, submitSignInForm, deleteKeycloakUserByEmail } from "./fixtures";

const KC_BASE_URL = (process.env.DEVHUB_KEYCLOAK_ADMIN_URL ?? "http://localhost:8180/devhub/auth/keycloak").replace(/\/+$/, "");
const KC_REALM = (process.env.DEVHUB_KEYCLOAK_ADMIN_REALM ?? "devhub").trim();
const KC_ADMIN_CLIENT_ID = (process.env.DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID ?? "devhub-e2e-seeder").trim();
const KC_ADMIN_CLIENT_SECRET = (process.env.DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET ?? "secret-change-me-backend").trim();

async function fetchAdminToken(): Promise<string> {
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
    throw new Error(`token grant failed: ${resp.status} ${await resp.text()}`);
  }
  const body = (await resp.json()) as { access_token?: string };
  if (!body.access_token) throw new Error("missing access_token");
  return body.access_token;
}

async function createKeycloakUser(email: string, password: string): Promise<void> {
  const token = await fetchAdminToken();
  const createResp = await fetch(`${KC_BASE_URL}/admin/realms/${encodeURIComponent(KC_REALM)}/users`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
      Accept: "application/json",
    },
    body: JSON.stringify({
      username: email,
      email,
      enabled: true,
      emailVerified: true,
      firstName: "E2E",
      lastName: "Onboarding",
      credentials: [{ type: "password", value: password, temporary: false }],
    }),
  });
  if (createResp.status !== 201 && createResp.status !== 409) {
    throw new Error(`user create failed: ${createResp.status} ${await createResp.text()}`);
  }
}

test.describe("first login onboarding", () => {
  test("token-only actor logs in via keycloak and lands on /onboarding", async ({ page }) => {
    const email = `e2e-new-${Date.now()}@example.com`;
    const password = "ChangeMe-12345!";

    await createKeycloakUser(email, password);
    try {
      await page.goto(appPath("/login")).catch((err) => {
        const msg = err instanceof Error ? err.message : String(err);
        if (!msg.includes("ERR_ABORTED")) throw err;
      });

      await waitForSignInForm(page);
      await submitSignInForm(page, email, password);

      await expect(page).toHaveURL(/\/onboarding(\/|$)/, { timeout: 30_000 });
      await expect(page.getByTestId("onboarding-form")).toBeVisible({ timeout: 10_000 });
    } finally {
      await deleteKeycloakUserByEmail(email);
    }
  });
});
