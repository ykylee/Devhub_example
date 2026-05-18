import { test, expect, SEEDED, loginAs } from "./fixtures";

// External Integration admin frontend (sprint claude/work_260518-h).
// TC 카탈로그는 docs/tests/test_cases_m4_integration.md §3.
// Backend: API-69~72 (sprint codex/next-step-20260516, PR #139).
// Page: frontend/app/(dashboard)/admin/settings/integrations/page.tsx
// (sprint claude/work_260518-g, PR #148).

test.describe("External Integration admin UI", () => {
  test("TC-INT-FRONTEND-LIST-01 + CREATE-01 + EDIT-01 + SYNC-01 — provider lifecycle", async ({ page }) => {
    const providerKey = `e2e-int-${Date.now()}`;
    const displayName = `E2E Integration ${Date.now()}`;
    const updatedDisplayName = `${displayName} (updated)`;
    let providerID = "";

    await test.step("TC-INT-FRONTEND-LIST-01 — system_admin 이 /admin/settings/integrations 접근", async () => {
      await loginAs(page, SEEDED.systemAdmin);
      await page.goto("/admin/settings/integrations");
      await expect(page.getByRole("heading", { name: /integration providers/i })).toBeVisible();
      // 페이지 로드 후 ProviderTable 또는 empty state 렌더 완료.
      // Playwright 의 CSS 셀렉터는 list 안에 text=/regex/ engine selector 를
      // 직접 못 둠 → invalid CSS 파싱 에러. locator.or() 패턴으로 두 후보를
      // chain (codex hotfix #7 P1, sprint claude/work_260518-m).
      const tableOrEmpty = page
        .locator("table")
        .or(page.getByText(/등록된 integration provider 가 없습니다/i))
        .first();
      await expect(tableOrEmpty).toBeVisible();
    });

    await test.step("TC-INT-FRONTEND-CREATE-01 — Register Provider 모달 등록", async () => {
      await page.getByRole("button", { name: /register provider/i }).click();
      const modal = page.getByRole("dialog");
      await expect(modal).toBeVisible();
      await expect(modal.getByRole("heading", { name: /register provider/i })).toBeVisible();

      await modal.getByLabel(/provider key/i).fill(providerKey);
      await modal.getByLabel(/^type/i).selectOption("scm");
      await modal.getByLabel(/auth mode/i).selectOption("token");
      await modal.getByLabel(/display name/i).fill(displayName);
      await modal.getByLabel(/credentials ref/i).fill("hmac_sha256:e2e-test-secret");
      await modal.getByLabel(/capabilities/i).fill("webhook, pull");

      await modal.getByRole("button", { name: /^register$/i }).click();

      // 모달 close + table row 추가.
      await expect(page.getByRole("dialog")).toBeHidden({ timeout: 10_000 });
      const row = page.locator("tr").filter({ hasText: displayName }).first();
      await expect(row).toBeVisible();
      await expect(row.getByText(providerKey)).toBeVisible();

      // provider_id 추출 — page.request.get 의 OIDC session propagation 이
      // CI 에서 flaky (codex hotfix #7+, sprint claude/work_260518-m). 대신
      // ProviderTable row 의 data-provider-id attribute 에서 직접 읽어
      // 결정적 매핑 보장.
      const newRow = page.locator(`tr[data-provider-id]`).filter({ hasText: providerKey }).first();
      await expect(newRow).toBeVisible({ timeout: 10_000 });
      providerID = (await newRow.getAttribute("data-provider-id")) ?? "";
      expect(providerID).toBeTruthy();
    });

    await test.step("TC-INT-FRONTEND-EDIT-01 — Edit 모달에서 display_name + capabilities + enabled 갱신", async () => {
      const row = page.locator("tr").filter({ hasText: providerKey }).first();
      await row.getByRole("button", { name: /edit/i }).click();
      const modal = page.getByRole("dialog");
      await expect(modal).toBeVisible();
      await expect(modal.getByRole("heading", { name: /edit provider/i })).toBeVisible();

      // provider_key 입력 필드는 edit 모드에서 노출 안 됨 (immutable).
      await expect(modal.getByLabel(/provider key/i)).toHaveCount(0);
      // provider_type / auth_mode 는 disabled.
      await expect(modal.getByLabel(/^type/i)).toBeDisabled();
      await expect(modal.getByLabel(/auth mode/i)).toBeDisabled();

      // display_name 갱신.
      const nameInput = modal.getByLabel(/display name/i);
      await nameInput.fill(updatedDisplayName);
      // enabled checkbox toggle (default true → false).
      const enabledCheckbox = modal.getByLabel(/enabled/i);
      await enabledCheckbox.uncheck();

      await modal.getByRole("button", { name: /^save$/i }).click();
      await expect(page.getByRole("dialog")).toBeHidden({ timeout: 10_000 });

      // table row 갱신 — updated display name.
      const updatedRow = page.locator("tr").filter({ hasText: updatedDisplayName }).first();
      await expect(updatedRow).toBeVisible();
      // enabled badge = "No" (disabled).
      await expect(updatedRow.getByText(/^no$/i)).toBeVisible();
    });

    await test.step("TC-INT-FRONTEND-SYNC-01 — Sync 버튼 클릭 → API-72 호출 + sync_status 전이", async () => {
      const row = page.locator("tr").filter({ hasText: updatedDisplayName }).first();
      const syncBtn = row.getByRole("button", { name: /sync/i });

      // codex hotfix #6 P2 (PR #149): cell visibility 만 검증하면 click 전후
      // 동일하게 PASS 되어 false positive. API-72 호출 자체를 network 으로 검증
      // + sync_status badge 가 "Pending" 으로 전이됐는지 확인 (page.tsx 의
      // optimistic update — handleSync 가 즉시 sync_status="requested" 로
      // setProviders, badge 매핑이 "Pending" label).
      const syncResponsePromise = page.waitForResponse(
        (resp) =>
          resp.url().includes(`/api/v1/integration/providers/${providerID}/sync`) &&
          resp.request().method() === "POST",
      );
      await syncBtn.click();
      const syncResponse = await syncResponsePromise;
      expect(syncResponse.ok()).toBeTruthy();
      const syncBody = await syncResponse.json();
      expect(syncBody.status).toBe("accepted");
      expect(syncBody.job_id).toBeTruthy();

      // optimistic update — badge 가 "Pending" 으로 즉시 전이.
      await expect(row.getByText(/pending/i)).toBeVisible({ timeout: 5_000 });
    });

    await test.step("TC-INT-FRONTEND-DELETE-01 — Delete 버튼 → DestructiveConfirmModal → API-80", async () => {
      // sprint claude/work_260518-j: 명시 DELETE endpoint (API-80) 도입으로
      // mega test 의 cleanup 이 실제 row 제거로 전환. binding 없는 provider 이므로
      // 200 OK 예상.
      const row = page.locator("tr").filter({ hasText: updatedDisplayName }).first();
      const deleteResponsePromise = page.waitForResponse(
        (resp) =>
          resp.url().includes(`/api/v1/integration/providers/${providerID}`) &&
          resp.request().method() === "DELETE",
      );
      await row.getByRole("button", { name: /delete/i }).click();

      // DestructiveConfirmModal 의 confirm.
      const confirmModal = page.getByRole("dialog");
      await expect(confirmModal.getByText(/delete provider/i)).toBeVisible();
      await confirmModal.getByRole("button", { name: /delete/i, exact: true }).click();

      const deleteResponse = await deleteResponsePromise;
      expect(deleteResponse.ok()).toBeTruthy();

      // table row 가 사라졌는지 — display name 으로 찾아 0 카운트 확인.
      await expect(page.locator("tr").filter({ hasText: updatedDisplayName })).toHaveCount(0, {
        timeout: 10_000,
      });
    });
  });

  test("TC-INT-FRONTEND-RBAC-01 — non-system_admin 접근 시 default landing 으로 redirect", async ({ page }) => {
    // developer 로 로그인 → /admin/settings/integrations 직접 접근 → /developer redirect.
    await loginAs(page, SEEDED.developer);
    await page.goto("/admin/settings/integrations");
    // AuthGuard / layout.tsx 의 isSystemAdmin 가드가 default landing 으로 redirect.
    await page.waitForURL(/\/developer(\/|$)/, { timeout: 15_000 });
    // Integration page 본문 ("Integration Providers" heading) 는 노출 안 됨.
    await expect(page.getByRole("heading", { name: /integration providers/i })).toHaveCount(0);
  });
});
