import { test, expect, SEEDED, loginAs } from "./fixtures";

// External Integration bindings admin frontend (sprint claude/work_260518-m).
// TC 카탈로그는 docs/tests/test_cases_m4_integration.md §3.
// Backend: API-74 GET + API-75 POST (sprint codex/next-step-20260516, PR #139).
// Page: frontend/app/(dashboard)/admin/settings/integration-bindings/page.tsx
// (sprint claude/work_260518-m).

test.describe("External Integration bindings admin UI", () => {
  test("TC-INT-FRONTEND-BIND-LIST-01 + CREATE-01 — bindings lifecycle", async ({ page }) => {
    const stamp = Date.now();
    const providerKey = `e2e-bind-${stamp}`;
    const providerDisplay = `E2E Bind Provider ${stamp}`;
    const scopeID = `APP-E2E-${stamp}`;
    const externalKey = `EXT-${stamp}`;
    let providerID = "";

    await test.step("seed — system_admin 으로 provider 1건 생성 (binding 생성 사전 조건)", async () => {
      await loginAs(page, SEEDED.systemAdmin);
      // page.request 가 OIDC session 을 공유해 API-70 호출.
      const createResp = await page.request.post("/api/v1/integration/providers", {
        data: {
          provider_key: providerKey,
          provider_type: "scm",
          display_name: providerDisplay,
          auth_mode: "token",
          credentials_ref: "hmac_sha256:e2e-bind-secret",
          capabilities: ["webhook"],
        },
      });
      expect(createResp.ok()).toBeTruthy();
      const body = await createResp.json();
      providerID = body.data.provider_id as string;
      expect(providerID).toBeTruthy();
    });

    await test.step("TC-INT-FRONTEND-BIND-LIST-01 — system_admin 이 /admin/settings/integration-bindings 접근", async () => {
      await page.goto("/admin/settings/integration-bindings");
      await expect(page.getByRole("heading", { name: /integration bindings/i })).toBeVisible();
      // 페이지 로드 후 BindingsTable 또는 empty state 렌더 완료.
      const tableOrEmpty = page
        .locator("table, text=/등록된 binding 이 없습니다/i")
        .first();
      await expect(tableOrEmpty).toBeVisible();
    });

    await test.step("TC-INT-FRONTEND-BIND-CREATE-01 — Create Binding 모달 등록 + table 반영", async () => {
      // codex hotfix #6 P2 (PR #149) 패턴: API-75 호출을 waitForResponse 로
      // 직접 검증해 false positive 차단.
      const createResponsePromise = page.waitForResponse(
        (resp) =>
          resp.url().includes("/api/v1/integration/bindings") &&
          resp.request().method() === "POST",
      );

      await page.getByRole("button", { name: /create binding/i }).click();
      const modal = page.getByRole("dialog");
      await expect(modal).toBeVisible();
      await expect(modal.getByRole("heading", { name: /create binding/i })).toBeVisible();

      await modal.getByLabel(/scope type/i).selectOption("application");
      await modal.getByLabel(/scope id/i).fill(scopeID);
      // provider dropdown — display_name (provider_key · type) 패턴이 option label.
      await modal.getByLabel(/^provider/i).selectOption(providerID);
      await modal.getByLabel(/external key/i).fill(externalKey);
      await modal.getByLabel(/^policy/i).selectOption("execution_system");

      await modal.getByRole("button", { name: /^create$/i }).click();

      const createResponse = await createResponsePromise;
      expect(createResponse.ok()).toBeTruthy();
      const createBody = await createResponse.json();
      expect(createBody.data.scope_id).toBe(scopeID);
      expect(createBody.data.external_key).toBe(externalKey);

      // 모달 close + table row 추가.
      await expect(page.getByRole("dialog")).toBeHidden({ timeout: 10_000 });
      const row = page.locator("tr").filter({ hasText: scopeID }).first();
      await expect(row).toBeVisible();
      await expect(row.getByText(externalKey)).toBeVisible();
      await expect(row.getByText(providerDisplay)).toBeVisible();
    });
  });

  test("TC-INT-FRONTEND-BIND-RBAC-01 — non-system_admin 접근 시 default landing 으로 redirect", async ({ page }) => {
    // developer 로 로그인 → /admin/settings/integration-bindings 직접 접근 →
    // /developer redirect (AuthGuard + layout.tsx 의 isSystemAdmin 가드).
    await loginAs(page, SEEDED.developer);
    await page.goto("/admin/settings/integration-bindings");
    await page.waitForURL(/\/developer(\/|$)/, { timeout: 15_000 });
    // Bindings page 본문 ("Integration Bindings" heading) 는 노출 안 됨.
    await expect(page.getByRole("heading", { name: /integration bindings/i })).toHaveCount(0);
  });
});
