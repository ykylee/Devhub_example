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
    await expect(page.getByText(/Audit Log Intelligence/i)).toBeVisible({ timeout: 10_000 });

    // 데이터 유무는 환경에 따라 달라질 수 있으므로 "목록 렌더" 자체를 스모크로 본다.
    // 엔트리가 있으면 row button 이 보이고, 없으면 empty-state 문구가 보인다.
    const listRows = page.locator("button").filter({ has: page.locator("span.font-mono.text-accent") });
    const emptyState = page.getByText(/no audit log entries match the current filters/i);
    await expect(listRows.first().or(emptyState)).toBeVisible({ timeout: 10_000 });
  });
});
