import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";
import {
  createDogfoodScmProvider,
  deleteDogfoodProvider,
  listDogfoodScmRepositories,
  syncDogfoodProvider,
} from "./dogfood-helpers";

test.describe("dogfood gitea integration admin flow", () => {
  test("system_admin registers a Gitea provider, syncs it, and verifies remote repositories", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto(appPath("/admin/settings/integrations"));
    await expect(page.getByRole("heading", { name: /integration providers/i })).toBeVisible({ timeout: 15_000 });

    let providerID = "";
    let providerKey = "";
    let displayName = "";

    try {
      await test.step("register dogfood gitea provider", async () => {
        const provider = await createDogfoodScmProvider(page);
        providerID = provider.providerID;
        providerKey = provider.providerKey;
        displayName = provider.displayName;

        await page.reload();
        const row = page.locator(`tr[data-provider-id="${providerID}"]`).first();
        await expect(row).toBeVisible({ timeout: 15_000 });
        await expect(row.getByText(providerKey)).toBeVisible();
        await expect(row.getByText(displayName)).toBeVisible();
      });

      await test.step("request sync job", async () => {
        const sync = await syncDogfoodProvider(page, providerID);
        expect(sync.jobID).toBeTruthy();

        const row = page.locator(`tr[data-provider-id="${providerID}"]`).first();
        await page.reload();
        await expect(row).toBeVisible({ timeout: 15_000 });
      });

      await test.step("verify Gitea repositories are reachable through provider API", async () => {
        let repositories: Array<{ full_name?: string; imported?: boolean }> = [];
        await expect
          .poll(async () => {
            repositories = await listDogfoodScmRepositories(page, providerID);
            return repositories.length;
          }, { timeout: 20_000, intervals: [1_000, 2_000, 3_000] })
          .toBeGreaterThan(0);

        const fullNames = repositories.map((repo) => repo.full_name).filter(Boolean);
        expect(fullNames).toContain("yklee/devhub-simulation");
      });
    } finally {
      if (providerID) {
        await deleteDogfoodProvider(page, providerID);
        await page.reload();
        await expect(page.locator(`tr[data-provider-id="${providerID}"]`)).toHaveCount(0, { timeout: 15_000 });
      }
    }
  });
});
