import { test, expect, loginAs, SEEDED } from "./fixtures";

// audit.spec — /admin/settings/audit smoke. Pins:
//   1) the page is reachable by a system_admin actor (AuthGuard + sub-tab),
//   2) the list renders entries from GET /api/v1/audit-logs (so the service
//      contract aligns with audit.go's auditLogResponse — earlier draft had
//      `id`/`event_id`/`occurred_at` field names that did not exist).

test.describe("/admin/settings/audit", () => {
  test("system_admin can open the audit tab and see log entries", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);

    await page.goto("/admin/settings/audit");
    await expect(page).toHaveURL(/\/admin\/settings\/audit/);

    // Filter card heading + first action column come from page.tsx.
    await expect(page.getByText(/Audit Log Filters/i)).toBeVisible({ timeout: 10_000 });

    // Backend has audit rows from every prior login (auth.login.succeeded),
    // so the first page should contain at least one entry. The page renders
    // each entry inside a button; assert at least one is present.
    const entries = page.locator('button:has-text("auth.login")');
    await expect(entries.first()).toBeVisible({ timeout: 10_000 });
  });
});
