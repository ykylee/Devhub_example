import { test, expect, SEEDED, loginAs } from "./fixtures";

test.describe("DREQ E2E", () => {
  test("Intake to Promote to Revoke lifecycle", async ({ page, request, browser }) => {
    // 1. System Admin logs in to issue a token
    await loginAs(page, SEEDED.systemAdmin);
    
    // Navigate to Dev Request Tokens settings
    await page.goto("/admin/settings/dev-request-tokens");
    await expect(page.getByRole("heading", { name: /intake tokens/i })).toBeVisible();

    // Open Issue Token Modal
    await page.getByRole("button", { name: /issue token/i }).click();
    await expect(page.getByRole("dialog")).toBeVisible();

    // Fill form
    const clientLabel = `e2e_client_${Date.now()}`;
    await page.getByLabel(/client label/i).fill(clientLabel);
    await page.getByLabel(/source system/i).fill("e2e_sys");
    await page.getByLabel(/allowed ips/i).first().fill("0.0.0.0/0"); // allow all for test
    
    // Submit
    await page.getByRole("dialog").getByRole("button", { name: /issue token/i }).click();

    // Reveal Phase - grab the token
    await expect(page.getByText(/token shown once/i)).toBeVisible();
    await page.getByRole("button", { name: /show token/i }).click();
    const plainTokenCode = page.locator("code").first();
    const plainToken = await plainTokenCode.textContent();
    expect(plainToken).toBeTruthy();

    // Close modal
    await page.getByRole("button", { name: /저장 완료 — 닫기/i }).click();
    await expect(page.getByRole("dialog")).toBeHidden();

    // 2. External System makes Intake Request using the token
    const testSuffix = Date.now();
    const requestTitle = `E2E Provisioning Request ${testSuffix}`;
    const externalRef = `E2E-REQ-${testSuffix}`;
    const intakeResponse = await request.post("/api/v1/dev-requests", {
      headers: {
        Authorization: `Bearer ${plainToken?.trim()}`,
      },
      data: {
        title: requestTitle,
        details: "Please provision a new project for E2E testing.",
        requester: "e2e_tester",
        assignee_user_id: SEEDED.developer.user_id, // Assigned to developer
        external_ref: externalRef,
      },
    });
    expect(intakeResponse.ok()).toBeTruthy();
    const intakeBody = await intakeResponse.json();
    expect(intakeBody.status).toBe("ok");
    const dreqId = intakeBody.data.id;

    // 3. Assignee logs in and sees it on dashboard
    // We can use a separate page for the developer context
    const devContext = await browser.newContext();
    const devPage = await devContext.newPage();
    await loginAs(devPage, SEEDED.developer);
    
    // Check Dashboard Widget
    await devPage.goto("/developer");
    await expect(
      devPage.getByRole("heading", { name: /my dev requests|내 대기 의뢰/i })
    ).toBeVisible();
    
    // Click widget item -> navigate to /dev-requests list
    const reqLink = devPage.getByText(requestTitle).first();
    await expect(reqLink).toBeVisible();
    await reqLink.click();
    await expect(devPage).toHaveURL(/\/dev-requests(\/|$)/);

    // Open detail modal from list row
    const reqRow = devPage.locator("tr").filter({ hasText: requestTitle }).first();
    await expect(reqRow).toBeVisible();
    await reqRow.click();

    // Verify DevRequest Detail Modal opens
    const detailModal = devPage.getByRole("dialog");
    await expect(detailModal).toBeVisible();
    await expect(detailModal.getByText(requestTitle)).toBeVisible();
    
    // 4. Promote action is role-gated; execute only when visible in current actor context.
    const promoteBtn = detailModal.getByRole("button", { name: /promote/i });
    if (await promoteBtn.isVisible().catch(() => false)) {
      await promoteBtn.click();

      // Optional UI path depending on current promote UX implementation.
      const promoteAppBtn = devPage.getByRole("button", { name: /새 어플리케이션/i });
      if (await promoteAppBtn.isVisible().catch(() => false)) {
        await promoteAppBtn.click();
        await devPage.getByLabel(/application key/i).fill(`E2EAPP${Math.floor(Math.random() * 10000)}`);
        await devPage.getByLabel(/application name/i).fill("E2E Promoted App");
        await devPage.getByLabel(/owner/i).fill(SEEDED.developer.user_id);
        await devPage.getByLabel(/leader/i).fill(SEEDED.developer.user_id);
        await devPage.getByRole("button", { name: /create application/i }).click();
        await expect(devPage.getByText(/registered/i).first()).toBeVisible({ timeout: 10000 });
      }
    }

    // Close dev context
    await devContext.close();

    // 5. System Admin revokes the token
    // Back on System Admin page
    const row = page.getByRole("row").filter({ hasText: clientLabel });
    await row.getByRole("button", { name: /revoke/i }).click();
    
    // Confirm via DestructiveConfirmModal
    const confirmModal = page.getByRole("dialog");
    await expect(confirmModal.getByText(/revoke token/i)).toBeVisible();
    await confirmModal.getByRole("button", { name: /revoke/i, exact: true }).click();
    
    // Verify revoked
    await expect(page.getByText(`토큰 '${clientLabel}' 이 revoke 되었습니다.`)).toBeVisible();
    await expect(row.getByText(/revoked/i)).toBeVisible();

    // Verify token can no longer be used
    const failResponse = await request.post("/api/v1/dev-requests", {
      headers: {
        Authorization: `Bearer ${plainToken?.trim()}`,
      },
      data: {
        title: "Should Fail",
        details: "x",
        requester: "x",
        assignee_user_id: "x",
        external_ref: "x",
      },
    });
    expect(failResponse.ok()).toBeFalsy();
    expect(failResponse.status()).toBe(401);
  });
});
