import { test, expect, deleteKratosIdentityByEmail } from "./fixtures";

// signup.spec — F7 회원가입.
// TC-SIGNUP-01..04. /auth/signup 페이지 + HRDB mock (yklee/akim/sjones)
// round-trip + HR 매칭 실패 + password mismatch. Sign Up 자체는 M3 트랙
// 이지만 코드가 main 에 있어 운영 진입 직전 회귀로 본 sprint 게이트에
// 포함.
//
// HRDB Mock (backend-core/internal/hrdb/mock.go):
//   - YK Lee  / yklee  / 1001 → yklee@example.com  / Engineering
//   - Alex Kim / akim  / 1002 → akim@example.com   / Product
//   - Sam Jones / sjones / 1003 → sjones@example.com / Infrastructure

const NEW_USER = {
  name: "YK Lee",
  systemId: "yklee",
  employeeId: "1001",
  email: "yklee@example.com",
  password: "Signup-12345!",
};

test.describe("/auth/signup", () => {
  test("TC-SIGNUP-01 — page loads with form fields and Sign In link", async ({ page }) => {
    await page.goto("/auth/signup");
    await expect(page).toHaveURL(/\/auth\/signup/);

    // Page heading + key form labels
    await expect(page.getByRole("heading", { name: /join devhub/i })).toBeVisible({ timeout: 10_000 });
    for (const label of [/full name/i, /system id/i, /employee id/i, /^password$/i, /confirm password/i]) {
      await expect(page.getByLabel(label)).toBeVisible();
    }

    // Sign In link → /login
    const signInLink = page.getByRole("link", { name: /sign in/i });
    await expect(signInLink).toBeVisible();
    await signInLink.click();
    await expect(page).toHaveURL(/\/login(\?|$|\/)/, { timeout: 10_000 });
  });

  test("TC-SIGNUP-02 — HR mock match registers a new identity and redirects to /login", async ({ page }) => {
    // Pre-condition: yklee must not exist yet. Best-effort cleanup before
    // we start so a left-over identity from a previous failed run does
    // not 409 the POST.
    await deleteKratosIdentityByEmail(NEW_USER.email).catch((err) => {
      console.warn("[signup.spec] pre-cleanup warning:", err);
    });

    let identityCreated = false;
    try {
      await page.goto("/auth/signup");

      await page.getByLabel(/full name/i).fill(NEW_USER.name);
      await page.getByLabel(/system id/i).fill(NEW_USER.systemId);
      await page.getByLabel(/employee id/i).fill(NEW_USER.employeeId);
      await page.getByLabel(/^password$/i).fill(NEW_USER.password);
      await page.getByLabel(/confirm password/i).fill(NEW_USER.password);
      await page.getByRole("button", { name: /register account/i }).click();

      // Success state — handler renders the verification card and sets
      // a 3-second timer to push to /login.
      await expect(page.getByText(/identity verified/i)).toBeVisible({ timeout: 15_000 });
      identityCreated = true;

      // After the setTimeout the URL changes to /login. Wait generously.
      await page.waitForURL(/\/login(\?|$|\/)/, { timeout: 10_000 });
    } finally {
      if (!identityCreated) return;
      // Cleanup: delete the Kratos identity so the next run starts clean.
      // DevHub `users` row is left in place — globalSetup of subsequent
      // runs is idempotent (002_seed_e2e_users.sql) and the stray row
      // does not affect other specs (it is not in SEEDED).
      try {
        await deleteKratosIdentityByEmail(NEW_USER.email);
      } catch (err) {
        console.warn("[signup.spec] post-cleanup failed; manual `kratos delete identity` may be required:", err);
      }
    }
  });

  test("TC-SIGNUP-03 — HR mismatch is rejected with a clear error", async ({ page }) => {
    await page.goto("/auth/signup");

    // Triplet that nothing in the mock matches.
    await page.getByLabel(/full name/i).fill("Nobody");
    await page.getByLabel(/system id/i).fill("unknown-id");
    await page.getByLabel(/employee id/i).fill("9999");
    await page.getByLabel(/^password$/i).fill(NEW_USER.password);
    await page.getByLabel(/confirm password/i).fill(NEW_USER.password);
    await page.getByRole("button", { name: /register account/i }).click();

    // Backend returns 403 with code=hr_lookup_failed. Frontend surfaces
    // `data.details || data.error` — "identity verification failed".
    await expect(page.getByText(/identity verification failed|hr_lookup_failed/i)).toBeVisible({ timeout: 15_000 });
    // We must NOT have advanced to /login.
    await expect(page).toHaveURL(/\/auth\/signup/);
  });

  test("TC-SIGNUP-04 — password mismatch is caught client-side", async ({ page }) => {
    await page.goto("/auth/signup");

    await page.getByLabel(/full name/i).fill(NEW_USER.name);
    await page.getByLabel(/system id/i).fill(NEW_USER.systemId);
    await page.getByLabel(/employee id/i).fill(NEW_USER.employeeId);
    await page.getByLabel(/^password$/i).fill("Foo-12345!");
    await page.getByLabel(/confirm password/i).fill("Bar-67890!");
    await page.getByRole("button", { name: /register account/i }).click();

    await expect(page.getByText(/passwords do not match/i)).toBeVisible({ timeout: 5_000 });
    // No backend call, URL unchanged.
    await expect(page).toHaveURL(/\/auth\/signup/);
  });
});
