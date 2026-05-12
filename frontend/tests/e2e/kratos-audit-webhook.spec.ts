import { test, expect, loginAs, SEEDED, getKratosIdentityIdByEmail } from "./fixtures";

// kratos-audit-webhook.spec — F4 감사 (Kratos audit webhook).
// TC-AUD-01..02. PR-M2-AUDIT 의 Kratos settings/password/after web_hook
// 이 실제로 발화되어 audit_logs 에 source_type=kratos 행을 적었는지
// e2e 단에서 확인. Kratos `response.ignore: true` 라 webhook 실패해도
// password 변경 자체는 success — false-positive 회피를 위해 source_type
// 까지 detail 단에서 검증한다.
//
// 환경 전제 (docs/setup/test-server-deployment.md §3.4):
//   - DEVHUB_KRATOS_WEBHOOK_TOKEN 양쪽 (backend-core + Kratos) export
//   - Kratos 재기동으로 새 hook 활성
//
// alice 의 password 를 임시 변경하므로 account.spec.ts (TC-ACC-02) 와
// alice 시드를 공유한다. Playwright `fullyParallel: false, workers: 1`
// + 파일명 알파벳 순으로 account.spec (a) → kratos-audit-webhook.spec (k)
// 순 실행되며 양쪽 finally rollback 이 안전망. globalSetup 이 force-reset
// 으로 최종 복구.

test.describe("Kratos password webhook → audit_logs", () => {
  const original = SEEDED.developer.password;
  const rotated = `WebhookProbe-${Date.now()}!`;

  test("TC-AUD-01/02 — password change emits an audit row with source_type=kratos and target_id matching alice's Kratos identity", async ({ page }) => {
    // 0) precondition — alice 의 Kratos identity_id 를 사전 확인.
    const aliceIdentityID = await getKratosIdentityIdByEmail(SEEDED.developer.email);
    expect(aliceIdentityID, "alice Kratos identity must exist (globalSetup seeds it)").not.toBeNull();

    let passwordRotated = false;
    try {
      // 1) alice 로 로그인 + /account 진입 + password 변경
      await loginAs(page, SEEDED.developer);
      await page.goto("/account");
      await page.getByLabel(/current password/i).fill(original);
      await page.getByLabel(/^new password$/i).fill(rotated);
      await page.getByLabel(/confirm new password/i).fill(rotated);
      await page.getByRole("button", { name: /save changes/i }).click();
      await expect(page.getByText(/password updated successfully/i)).toBeVisible({ timeout: 15_000 });
      passwordRotated = true;

      // 2) Sign Out, charlie (system_admin) 로 재로그인
      await page.getByText(SEEDED.developer.user_id, { exact: false }).first().click();
      await page.getByRole("button", { name: /sign out/i }).click();
      await loginAs(page, SEEDED.systemAdmin);

      // 3) /admin/settings/audit 진입 + account.password_changed 행 노출
      await page.goto("/admin/settings/audit");
      await expect(page.getByText(/Audit Log Filters/i)).toBeVisible({ timeout: 15_000 });

      // 4) action=account.password_changed entry — audit.go 의
      //    auditLogResponse 가 action 을 mono 텍스트로 노출. 최신 entry
      //    가 prepend 되므로 첫 번째 매칭이 우리 행.
      const passwordEntry = page.locator('button:has-text("account.password_changed")').first();
      await expect(passwordEntry).toBeVisible({ timeout: 15_000 });

      // 5) entry 클릭 → detail 영역에서 source_type=kratos, target_id 매칭.
      await passwordEntry.click();
      // audit page 의 Detail 컴포넌트는 "source_type" + 값을 같은 컨테이너
      // 안에 노출. 정확한 마크업은 page.tsx:238 의 Detail label/value 구조.
      await expect(page.getByText("source_type")).toBeVisible({ timeout: 5_000 });
      await expect(page.getByText("kratos", { exact: true })).toBeVisible();
      // target_id 매칭 — Kratos UUID 가 detail 에 노출되어야 함.
      await expect(page.getByText(aliceIdentityID as string, { exact: false })).toBeVisible();
    } finally {
      if (!passwordRotated) return;
      // best-effort rollback (account.spec.ts 와 동일 패턴) —
      // globalSetup 이 다음 run 에서 force-reset 으로 최종 복구.
      try {
        // 시스템 관리자 로그인 상태에서 alice 로 갈아타려면 한 번 더
        // sign-out 후 alice 로 로그인. 일부 케이스에서 charlie 가 이미
        // sign-out 상태일 수 있으므로 try 로 감싼다.
        try {
          await page.getByText(SEEDED.systemAdmin.user_id, { exact: false }).first().click({ timeout: 3_000 });
          await page.getByRole("button", { name: /sign out/i }).click({ timeout: 3_000 });
        } catch {
          // already signed out — fine
        }
        await loginAs(page, { ...SEEDED.developer, password: rotated });
        await page.goto("/account");
        await page.getByLabel(/current password/i).fill(rotated);
        await page.getByLabel(/^new password$/i).fill(original);
        await page.getByLabel(/confirm new password/i).fill(original);
        await page.getByRole("button", { name: /save changes/i }).click();
        await expect(page.getByText(/password updated successfully/i)).toBeVisible({ timeout: 15_000 });
      } catch (err) {
        console.warn("[kratos-audit-webhook] best-effort rollback failed (globalSetup will recover on next run):", err);
      }
    }
  });
});
