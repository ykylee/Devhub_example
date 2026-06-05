import { apiPath, appPath, expect, loginAs, SEEDED, test } from "./fixtures";

test.describe("dogfood organization admin flow", () => {
  test("system_admin manages an organization unit end-to-end", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto(appPath("/admin/settings/organization"));
    await page.waitForLoadState("networkidle");
    await expect(page.getByText(/initializing organization data/i)).toBeHidden({ timeout: 20_000 });

    const unique = Date.now().toString().slice(-6);
    const unitName = `Dogfood Org Unit ${unique}`;
    const updatedUnitName = `${unitName} Updated`;
    let unitID = "";

    const searchInput = page.getByPlaceholder(/search units by name or type/i);
    const getUnitByLabel = async (label: string) =>
      await page.evaluate(async ({ label, hierarchyPath }) => {
        const token = sessionStorage.getItem("devhub_access_token");
        if (!token) throw new Error("missing access token");
        const resp = await fetch(hierarchyPath, {
          headers: { Authorization: `Bearer ${token}` },
        });
        const body = await resp.json();
        return (body?.data?.units ?? []).find((unit: { unit_id: string; label: string; leader_user_id?: string }) => unit.label === label) ?? null;
      }, { label, hierarchyPath: apiPath("/api/v1/organization/hierarchy") });
    const getUnitMembers = async (resolvedUnitID: string) =>
      await page.evaluate(async ({ membersPath }) => {
        const token = sessionStorage.getItem("devhub_access_token");
        if (!token) throw new Error("missing access token");
        const resp = await fetch(membersPath, {
          headers: { Authorization: `Bearer ${token}` },
        });
        const body = await resp.json();
        return Array.isArray(body?.data) ? body.data : [];
      }, { membersPath: apiPath(`/api/v1/organization/units/${encodeURIComponent(resolvedUnitID)}/members`) });

    await test.step("create a new organization unit", async () => {
      await page.getByRole("button", { name: /create unit/i }).click();
      const modal = page.getByRole("dialog");
      await expect(modal.getByText(/create new unit/i)).toBeVisible();

      await modal.getByPlaceholder(/infrastructure team/i).fill(unitName);
      await modal.locator("select").nth(0).selectOption("team");
      await modal.locator("select").nth(1).selectOption({ label: "Engineering" });
      await modal.locator("select").nth(2).selectOption("charlie");
      await modal.getByRole("button", { name: /create unit/i }).click();

      await expect(modal).toBeHidden({ timeout: 15_000 });
      await searchInput.fill(unitName);
      const row = page.getByRole("row").filter({ hasText: unitName }).first();
      await expect(row).toBeVisible({ timeout: 15_000 });
      await expect(row.getByText(/charlie/i)).toBeVisible();
      const createdUnit = await getUnitByLabel(unitName);
      expect(createdUnit?.unit_id).toBeTruthy();
      unitID = createdUnit.unit_id;
    });

    await test.step("edit unit name and change leader", async () => {
      const row = page.getByRole("row").filter({ hasText: unitName }).first();
      await row.getByRole("button", { name: new RegExp(`Edit ${unitName}`, "i") }).click();

      const modal = page.getByRole("dialog");
      await expect(modal.getByText(/edit unit details/i)).toBeVisible();
      const nameInput = modal.getByPlaceholder(/infrastructure team/i);
      await nameInput.fill(updatedUnitName);
      await modal.locator("select").nth(2).selectOption("bob");
      await modal.getByRole("button", { name: /save changes/i }).click();

      await expect(modal).toBeHidden({ timeout: 15_000 });
      await searchInput.fill(updatedUnitName);
      const updatedRow = page.getByRole("row").filter({ hasText: updatedUnitName }).first();
      await expect(updatedRow).toBeVisible({ timeout: 15_000 });
      await expect(updatedRow.getByText(/bob/i)).toBeVisible();
      await expect.poll(async () => (await getUnitByLabel(updatedUnitName))?.leader_user_id ?? "").toBe("bob");
    });

    await test.step("add members and change leader from member management", async () => {
      const row = page.getByRole("row").filter({ hasText: updatedUnitName }).first();
      await row.getByRole("button", { name: /members/i }).click();

      const modal = page.getByRole("dialog");
      await expect(modal.getByText(/manage members/i)).toBeVisible();

      const availableSearch = modal.getByPlaceholder(/search available/i);
      await availableSearch.fill("alice@example.com");
      await modal.getByText(/alice@example.com/i).first().click();

      await availableSearch.fill("bob@example.com");
      await modal.getByText(/bob@example.com/i).first().click();

      await expect(modal.getByText(/alice@example.com/i).first()).toBeVisible({ timeout: 15_000 });
      await modal.getByRole("button", { name: /set lead/i }).first().click();

      await modal.getByRole("button", { name: /save configuration/i }).click();
      await expect(modal).toBeHidden({ timeout: 15_000 });

      const updatedRow = page.getByRole("row").filter({ hasText: updatedUnitName }).first();
      await expect(updatedRow.getByText(/alice/i)).toBeVisible({ timeout: 15_000 });
      await expect.poll(async () => (await getUnitMembers(unitID)).map((member) => member.user_id).sort().join(",")).toContain("alice");
      await expect.poll(async () => (await getUnitMembers(unitID)).map((member) => member.user_id).sort().join(",")).toContain("bob");
      await expect.poll(async () => (await getUnitByLabel(updatedUnitName))?.leader_user_id ?? "").toBe("alice");
    });

    await test.step("modify members by removing one roster entry", async () => {
      const row = page.getByRole("row").filter({ hasText: updatedUnitName }).first();
      await row.getByRole("button", { name: /members/i }).click();

      const modal = page.getByRole("dialog");
      await expect(modal.getByText(/manage members/i)).toBeVisible();

      await modal.getByPlaceholder(/search current roster/i).fill("bob@example.com");
      const bobRow = modal.locator("div").filter({ hasText: /bob@example.com/i }).first();
      await bobRow.getByTitle(/remove from unit/i).click();
      await modal.getByRole("button", { name: /save configuration/i }).click();

      await expect(modal).toBeHidden({ timeout: 15_000 });
      const updatedRow = page.getByRole("row").filter({ hasText: updatedUnitName }).first();
      await expect(updatedRow.getByText(/alice/i)).toBeVisible();
      await expect.poll(async () => (await getUnitMembers(unitID)).map((member) => member.user_id).sort().join(",")).not.toContain("bob");
    });

    await test.step("delete the organization unit", async () => {
      page.once("dialog", async (dialog) => {
        await dialog.accept();
      });

      const row = page.getByRole("row").filter({ hasText: updatedUnitName }).first();
      await row.getByRole("button", { name: new RegExp(`Delete ${updatedUnitName}`, "i") }).click();

      await expect(page.getByRole("row").filter({ hasText: updatedUnitName })).toHaveCount(0, { timeout: 15_000 });
      await searchInput.fill(updatedUnitName);
      await expect(page.getByText(/no matching units found/i)).toBeVisible({ timeout: 15_000 });
      await expect.poll(async () => await getUnitByLabel(updatedUnitName)).toBeNull();
    });
  });
});
