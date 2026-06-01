import { test, expect } from "@playwright/test";
import { appPath } from "./fixtures";

test.describe("Verify build status fix and SCM activity (Ultra-stable mock-auth E2E)", () => {
  
  test.beforeEach(async ({ page }) => {
    // Intercept identity service whoAmI API to bypass OIDC login entirely
    await page.route("**/api/v1/me", async (route) => {
      console.log("[Mock OIDC] Intercepted /api/v1/me request, returning pre-seeded Charlie (System Admin)");
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          status: "ok",
          data: {
            login: "charlie",
            user_id: "charlie",
            subject: "55e9ed4c-9530-4a62-95b0-aa178afac019",
            role: "System Admin",
            display_name: "Charlie",
            email: "charlie@example.com",
            primary_unit_id: "root",
            onboarding_required: false,
            onboarding_completed_at: "2026-05-19T00:00:00Z",
            review_status: "reviewed"
          }
        })
      });
    });

    // Seed Zustand local storage to match the mock actor so frontend store state is consistent
    await page.addInitScript(() => {
      window.localStorage.setItem(
        "devhub-storage",
        JSON.stringify({
          state: {
            role: "System Admin",
            actor: {
              login: "charlie",
              user_id: "charlie",
              subject: "55e9ed4c-9530-4a62-95b0-aa178afac019",
              role: "System Admin",
              display_name: "Charlie",
              email: "charlie@example.com",
              primary_unit_id: "root",
              onboarding_required: false,
              onboarding_completed_at: "2026-05-19T00:00:00Z",
              review_status: "reviewed"
            },
            isDeepFocus: false,
            notifications: 3,
            isSidebarCollapsed: false
          },
          version: 0
        })
      );
      // Keep session storage empty (clear token) to force api-client to use dev_fallback (no Authorization header)
      window.sessionStorage.clear();
    });
  });

  test("verify application detail build status is '없음' and capture screenshot", async ({ page }) => {
    test.setTimeout(180000);

    // Directly navigate to DevHub Simulation App detail page
    const appId = "e8a9bc11-a89c-4cb1-8071-8890ab2345ef";
    console.log(`[E2E] Navigating directly to Application detail: ${appId}`);
    await page.goto(appPath(`/applications/${appId}`));
    await page.waitForLoadState("networkidle", { timeout: 20000 }).catch(() => undefined);
    await page.waitForTimeout(2000); // Settle delay for Recharts dynamic animations

    // Verify "Last Build" card value is "없음"
    const lastBuildCard = page.locator("text=Last Build").locator("xpath=..").locator("h3");
    await expect(lastBuildCard).toBeVisible();
    const lastBuildVal = await lastBuildCard.innerText();
    console.log("[E2E] Last Build Card Value:", lastBuildVal);
    expect(lastBuildVal).toBe("없음");

    // Verify "Target Branch Build Status" badge value is "없음" by querying target text inside section
    const targetBranchSection = page.locator("section:has-text('Target Branch Build Status')");
    await expect(targetBranchSection).toBeVisible();
    const badgeText = targetBranchSection.locator("text=없음").first();
    await expect(badgeText).toBeVisible();
    console.log("[E2E] Target Branch Badge Value: 없음 (Verified!)");

    // Take screenshot and save to artifacts directory
    const screenshotPath = "test-results/screenshots/application_detail_verified.png";
    await page.screenshot({ path: screenshotPath, fullPage: true });
    console.log("[E2E] Application Detail Screenshot saved to:", screenshotPath);
  });

  test("verify project detail SCM Activity and capture screenshot", async ({ page }) => {
    test.setTimeout(180000);

    // Directly navigate to DevHub Simulation Project detail page
    const projectId = "31b9e2cb-b1b0-466a-bb10-ea00ee1234a1";
    console.log(`[E2E] Navigating directly to Project detail: ${projectId}`);
    await page.goto(appPath(`/projects/${projectId}`));
    await page.waitForLoadState("networkidle", { timeout: 20000 }).catch(() => undefined);
    await page.waitForTimeout(2000); // Settle delay for charts/activity rendering

    // Verify "SCM Activity (PR & Issues)" widget presence
    const scmActivityTitle = page.locator("text=SCM Activity");
    await expect(scmActivityTitle).toBeVisible();

    // Take screenshot and save to artifacts directory
    const screenshotPath = "test-results/screenshots/project_detail_verified.png";
    await page.screenshot({ path: screenshotPath, fullPage: true });
    console.log("[E2E] Project Detail Screenshot saved to:", screenshotPath);
  });
});
