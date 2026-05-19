import { test, expect, waitForSignInForm } from "./fixtures";

// Self-signup is disabled during Keycloak migration.
// This spec asserts the placeholder/guard behavior.

test.describe("/auth/signup", () => {
  test("shows migration notice and sign-in link", async ({ page }) => {
    await page.goto("/auth/signup");
    await expect(page).toHaveURL(/\/auth\/signup/);

    await expect(page.getByRole("heading", { name: /sign up unavailable/i })).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText(/self-signup is temporarily disabled/i)).toBeVisible();

    // Sign In link → /login entrypoint, then OIDC flow may immediately
    // redirect to provider sign-in form depending on runtime config.
    const signInLink = page.getByRole("link", { name: /sign in/i });
    await expect(signInLink).toBeVisible();
    await signInLink.click();
    await waitForSignInForm(page);
  });
});
