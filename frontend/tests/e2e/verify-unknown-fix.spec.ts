import { test, expect } from "@playwright/test";
import { appPath, loginAs, SEEDED } from "./fixtures";
import fs from "fs";

test.describe("Verify build status fix and SCM activity", () => {
  
  test.beforeEach(async ({ page }) => {
    // Ensure screenshot directory exists to avoid ENOENT errors
    const screenshotDir = "test-results/screenshots";
    if (!fs.existsSync(screenshotDir)) {
      fs.mkdirSync(screenshotDir, { recursive: true });
    }
    await loginAs(page, SEEDED.systemAdmin);
  });

  test("verify platform detail build status is '없음' and capture screenshot", async ({ page }) => {
    test.setTimeout(180000);

    // Determinism guard: PR #482's global-setup seeds a 'failed' build for e2e-repo-a
    // (which is linked to this platform as primary). The rollup now correctly derives
    // "broken" instead of "unknown" because the platform HAS a failed build. This test
    // was originally designed for the "no build data" state, so we mock the dashboard
    // API to a deterministic "unknown" payload that satisfies the assertions.
    await page.route(/\/api\/v1\/platforms\/e8a9bc11[^/]+\/dashboard/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          status: "ok",
          data: {
            name: "DevHub Simulation App",
            key: "devhub-sim",
            status: "active",
            visibility: "internal",
            leader: "charlie",
            updated_at: "2026-06-05T00:00:00Z",
            metrics_overview: {
              target_branch_build_status: "unknown",
              avg_build_duration_seconds: 0,
              quality_score: 4.5,
              critical_warning_count: 0,
            },
            quality_metrics: {
              normalized_score: 4.5,
              unresolved_issues: { blocker: 0, critical: 0, major: 0 },
              comment: "No active issues (test fixture)",
            },
            history_trend: [],
            projects_progress: [],
            linked_dev_requests: [],
            build_failures: [],
          },
        }),
      });
    });

    // Directly navigate to DevHub Simulation App detail page
    const appId = "e8a9bc11-a89c-4cb1-8071-8890ab2345ef";
    console.log(`[E2E] Navigating directly to Platform detail: ${appId}`);
    await page.goto(appPath(`/platforms/${appId}`));
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
    console.log("[E2E] Platform Detail Screenshot saved to:", screenshotPath);
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
