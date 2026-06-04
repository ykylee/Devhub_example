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
