import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";
import { createSelfDogfoodWorkspace } from "./dogfood-helpers";

test.describe("dogfood self-dogfood admin flow", () => {
  test("system_admin creates platform, repository draft, and project for this workspace", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    const result = await createSelfDogfoodWorkspace(page);

    expect(result.platformID).toBeTruthy();
    expect(result.projectID).toBeTruthy();
    expect(result.repositoryID).toBeTruthy();
    expect(result.repositorySlug).toBeTruthy();
    expect(result.platformRepoCount).toBeGreaterThan(0);
    expect(result.projectRepoCount).toBeGreaterThan(0);

    await page.goto(appPath(`/platforms/${result.platformID}`));
    await expect(page).toHaveURL(new RegExp(`/platforms/${result.platformID}(/|$)`), { timeout: 20_000 });
    await expect(page.getByText(result.platformName)).toBeVisible({ timeout: 15_000 });

    await page.goto(appPath(`/projects/${result.projectID}`));
    await expect(page).toHaveURL(new RegExp(`/projects/${result.projectID}(/|$)`), { timeout: 20_000 });
    await expect(page.getByText(result.projectName)).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText(result.repositorySlug)).toBeVisible({ timeout: 15_000 });
  });
});
