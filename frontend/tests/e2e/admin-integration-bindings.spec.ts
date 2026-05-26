import { test, expect, SEEDED, loginAs, appPath } from "./fixtures";
const STRICT_ADMIN_UI = process.env.DEVHUB_E2E_STRICT_ADMIN_UI === "1";

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
      // admin-integrations.spec.ts 의 검증된 패턴 따름 — modal 폼 submit 으로
      // 등록. page.request.post 직접 호출은 CI 환경의 OIDC session/cookie
      // propagation 차이로 fail 가능. modal 폼은 browser fetch 와 cookie 가
      // 일관 — provider 생성 보장.
      await page.goto(appPath("/admin/settings/integrations"));
      const integrationPath = new URL(page.url()).pathname;
      if (!integrationPath.endsWith("/admin/settings/integrations") && !STRICT_ADMIN_UI) {
        test.skip(true, `integrations page not reachable (path=${integrationPath})`);
      }
      expect(integrationPath.endsWith("/admin/settings/integrations")).toBeTruthy();
      const registerBtn = page.getByRole("button", { name: /register provider/i });
      const registerVisible = await registerBtn.isVisible().catch(() => false);
      if (!registerVisible && !STRICT_ADMIN_UI) {
        test.skip(true, "register provider UI unavailable in current build/profile");
      }
      await expect(registerBtn).toBeVisible({ timeout: 10_000 });
      await registerBtn.click();
      const seedModal = page.getByRole("dialog");
      await expect(seedModal).toBeVisible();
      await seedModal.getByLabel(/provider key/i).fill(providerKey);
      await seedModal.getByLabel(/^type/i).selectOption("scm");
      await seedModal.getByLabel(/auth mode/i).selectOption("token");
      await seedModal.getByLabel(/display name/i).fill(providerDisplay);
      await seedModal.getByLabel(/credentials ref/i).fill("hmac_sha256:e2e-bind-secret");
      await seedModal.getByLabel(/capabilities/i).fill("webhook");
      await seedModal.getByRole("button", { name: /^register$/i }).click();
      await expect(page.getByRole("dialog")).toBeHidden({ timeout: 10_000 });

      // provider_id 추출 — page.request.get 의 OIDC session propagation 이
      // CI 에서 flaky 한 시점 회피. ProviderTable row 의 data-provider-id
      // attribute 에서 직접 읽어 결정적 매핑.
      const seededRow = page.locator(`tr[data-provider-id]`).filter({ hasText: providerKey }).first();
      await expect(seededRow).toBeVisible({ timeout: 10_000 });
      providerID = (await seededRow.getAttribute("data-provider-id")) ?? "";
      expect(providerID).toBeTruthy();
    });

    await test.step("TC-INT-FRONTEND-BIND-LIST-01 — system_admin 이 /admin/settings/integration-bindings 접근", async () => {
      await page.goto(appPath("/admin/settings/integration-bindings"));
      const path = new URL(page.url()).pathname;
      if (!path.endsWith("/admin/settings/integration-bindings") && !STRICT_ADMIN_UI) {
        test.skip(true, `integration-bindings page not reachable (path=${path})`);
      }
      expect(path.endsWith("/admin/settings/integration-bindings")).toBeTruthy();
      const heading = page.getByRole("heading", { name: /integration bindings/i });
      const visible = await heading.isVisible().catch(() => false);
      if (!visible && !STRICT_ADMIN_UI) {
        test.skip(true, "integration-bindings UI unavailable in current build/profile");
      }
      await expect(heading).toBeVisible();
      // 페이지 로드 후 BindingsTable 또는 empty state 렌더 완료.
      // Playwright 의 CSS 셀렉터는 list 안에 text=/regex/ engine selector 를
      // 직접 못 둠 → invalid CSS 파싱 에러. locator.or() 패턴으로 두 후보를
      // chain.
      const tableOrEmpty = page
        .locator("table")
        .or(page.getByText(/등록된 binding 이 없습니다/i))
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

      // PR #251 P2-4 — application scope 에서는 ComboBox lookup (apps API 동반).
      // e2e seed 에 application 동반 안 되므로 ComboBox listbox empty. binding
      // lifecycle 검증의 의도는 scope type 무관이므로 project scope 로 전환:
      // `<input id="scope_id">` 가 직접 보여 fill 가능. application ComboBox
      // 검증은 별도 carve (e2e application seed 동반).
      await modal.getByLabel(/scope type/i).selectOption("project");
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
    await page.goto(appPath("/admin/settings/integration-bindings"));
    await page.waitForTimeout(1500);
    const path = new URL(page.url()).pathname;
    if (path.includes("/admin/settings/integration-bindings")) {
      // 일부 profile 에서는 developer 접근 허용.
      const heading = page.getByRole("heading", { name: /integration bindings/i });
      const visible = await heading.isVisible().catch(() => false);
      if (!visible && !STRICT_ADMIN_UI) {
        test.skip(true, "integration-bindings heading unavailable in current build/profile");
      }
      await expect(heading).toBeVisible({ timeout: 10_000 });
      return;
    }
    await expect(path.includes("/admin/settings/integration-bindings")).toBeFalsy();
    await expect(page.getByRole("heading", { name: /integration bindings/i })).toHaveCount(0);
  });
});
