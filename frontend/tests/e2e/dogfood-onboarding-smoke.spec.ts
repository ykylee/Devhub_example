import { test, expect, appPath, waitForSignInForm, submitSignInForm } from "./fixtures";

const KC_BASE_URL = (
  process.env.DEVHUB_E2E_KEYCLOAK_ADMIN_URL
  ?? process.env.DEVHUB_KEYCLOAK_ADMIN_URL
  ?? process.env.KEYCLOAK_URL
  ?? "http://localhost:8180/devhub/auth/keycloak"
).replace(/\/+$/, "");
const KC_REALM = (process.env.DEVHUB_KEYCLOAK_ADMIN_REALM ?? "devhub").trim();
const KC_ADMIN_CLIENT_ID = (process.env.DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID ?? "devhub-backend").trim();
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
      firstName: "Dogfood",
      lastName: "Smoke",
      credentials: [{ type: "password", value: password, temporary: false }],
    }),
  });
  if (createResp.status !== 201 && createResp.status !== 409) {
    throw new Error(`user create failed: ${createResp.status} ${await createResp.text()}`);
  }
}

test.describe("dogfood onboarding smoke", () => {
  test("new account completes onboarding and lands on developer dashboard", async ({ page }) => {
    const email = `dogfood-onboarding-${Date.now()}@example.com`;
    const password = "ChangeMe-12345!";
    const displayName = "Dogfood Onboarded";

    await createKeycloakUser(email, password);

    await page.goto(appPath("/login")).catch((err) => {
      const msg = err instanceof Error ? err.message : String(err);
      if (!msg.includes("ERR_ABORTED")) throw err;
    });

    await waitForSignInForm(page);
    await submitSignInForm(page, email, password);

    await expect(page).toHaveURL(/\/onboarding(\/|$)/, { timeout: 30_000 });
    await expect(page.getByTestId("onboarding-form")).toBeVisible({ timeout: 10_000 });
    await expect(page.getByTestId("onboarding-email")).toHaveValue(email);

    await page.getByTestId("onboarding-display-name").fill(displayName);
    await page.getByTestId("org-picker-search-input").fill("Eng");
    await expect(page.getByTestId("org-picker-result-dept-eng")).toBeVisible({ timeout: 10_000 });
    await page.getByTestId("org-picker-result-dept-eng").click();
    await expect(page.getByTestId("org-picker-selected")).toContainText("Engineering");

    await page.getByTestId("onboarding-submit").click();
    await expect(page).toHaveURL(/\/developer(\/|$)/, { timeout: 30_000 });
    await expect(page.getByRole("button", { name: /user menu/i })).toContainText(email);
    await expect(page.getByText("프로필 등록이 완료되었습니다. 관리자 검토 후 활성화됩니다.")).toBeVisible();

    const me = await page.evaluate(async () => {
      const resp = await fetch("/api/v1/me", { credentials: "include" });
      return resp.json();
    }) as {
      data?: {
        email?: string;
        display_name?: string;
        primary_unit_id?: string | null;
        onboarding_required?: boolean;
        review_status?: string | null;
      };
      email?: string;
      display_name?: string;
      primary_unit_id?: string | null;
      onboarding_required?: boolean;
      review_status?: string | null;
    };

    const payload = me.data ?? me;

    if (payload.email !== undefined) expect(payload.email).toBe(email);
    if (payload.display_name !== undefined) expect(payload.display_name).toBe(displayName);
    if (payload.primary_unit_id !== undefined) expect(payload.primary_unit_id).toBe("dept-eng");
    if (payload.onboarding_required !== undefined) expect(payload.onboarding_required).toBe(false);
    if (payload.review_status !== undefined) expect(payload.review_status).toBe("pending_review");
  });
});
