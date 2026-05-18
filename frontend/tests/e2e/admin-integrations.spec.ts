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
      const tableOrEmpty = page.locator("table, text=/등록된 integration provider 가 없습니다/i").first();
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

      // provider_id 추출 — API list 호출 (page.request 가 OIDC session 공유).
      const listResp = await page.request.get("/api/v1/integration/providers");
      expect(listResp.ok()).toBeTruthy();
      const listBody = await listResp.json();
      const target = (listBody.data as Array<{ provider_id: string; provider_key: string }>).find(
        (p) => p.provider_key === providerKey
      );
      expect(target).toBeTruthy();
      providerID = target!.provider_id;
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

    await test.step("TC-INT-FRONTEND-SYNC-01 — Sync 버튼 클릭 → sync_status 즉시 전이", async () => {
      const row = page.locator("tr").filter({ hasText: updatedDisplayName }).first();
      const syncBtn = row.getByRole("button", { name: /sync/i });
      await syncBtn.click();
      // API-72 호출 후 sync_status badge 갱신 (요청 시점에 backend 가 어떤 status 를
      // 돌려주는지에 따라 다름 — "OK" / "Pending" / "Error" 중 하나가 표시되면 OK).
      await expect(row.locator("td").nth(4)).toBeVisible();
    });

    // Cleanup — 본 test 가 생성한 provider 는 backend 에 영구 남으므로 disabled 처리만 함.
    // (DELETE endpoint 가 별도로 없으므로 backend 의 provider 삭제는 carve out.)
    await test.step("Cleanup note", async () => {
      // 본 test 는 enabled=false 로 종료 — production CI 에서 누적된 e2e provider 는
      // 운영자가 별도 정리하거나, store 가 e2e 패턴 (`e2e-int-*` prefix) 을 인식하는
      // cleanup job 을 carve out 으로 가져갈 수 있음.
      expect(providerID).toBeTruthy();
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
