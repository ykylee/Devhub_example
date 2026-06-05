import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";
import { createSelfDogfoodWorkspace } from "./dogfood-helpers";

test.describe("dogfood self-dogfood dashboard flow", () => {
  test("system_admin verifies the self-dogfood platform and project dashboards", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    const workspace = await createSelfDogfoodWorkspace(page);

    await test.step("platform dashboard renders linked project and repository widgets", async () => {
      await page.goto(appPath(`/platforms/${workspace.platformID}`));
      await expect(page).toHaveURL(new RegExp(`/platforms/${workspace.platformID}(/|$)`), { timeout: 20_000 });
      await expect(page.getByRole("heading", { name: workspace.platformName })).toBeVisible({ timeout: 15_000 });
      await expect(page.getByText(/build & quality 7-day trend/i)).toBeVisible();
      await expect(page.getByText(/linked projects roadmap/i)).toBeVisible();
      await expect(page.getByRole("heading", { name: /^repositories$/i })).toBeVisible();
      await expect(page.getByText(workspace.projectName)).toBeVisible({ timeout: 15_000 });
      await expect(page.getByText(workspace.repositorySlug).first()).toBeVisible({ timeout: 15_000 });
    });

    await test.step("project dashboard renders repository, activity, and task widgets", async () => {
      await page.goto(appPath(`/projects/${workspace.projectID}`));
      await expect(page).toHaveURL(new RegExp(`/projects/${workspace.projectID}(/|$)`), { timeout: 20_000 });
      await expect(page.getByRole("heading", { name: workspace.projectName })).toBeVisible({ timeout: 15_000 });
      await expect(page.getByText(/connected repositories/i)).toBeVisible();
      await expect(page.getByText(/recent activity/i)).toBeVisible();
      await expect(page.getByText(/active tasks/i)).toBeVisible();
      await expect(page.getByText(/linked repositories/i)).toBeVisible();
      await expect(page.getByText(workspace.repositorySlug).first()).toBeVisible({ timeout: 15_000 });
    });
  });
});
