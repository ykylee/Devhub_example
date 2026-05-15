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
    const externalRef = `E2E-REQ-${Date.now()}`;
    const intakeResponse = await request.post("/api/v1/dev-requests", {
      headers: {
        Authorization: `Bearer ${plainToken?.trim()}`,
      },
      data: {
        title: "E2E Provisioning Request",
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
    await expect(devPage.getByRole("heading", { name: /my dev requests/i })).toBeVisible();
    
    // Click on the specific request in the widget
    const reqLink = devPage.getByText("E2E Provisioning Request");
    await expect(reqLink).toBeVisible();
    await reqLink.click();
    
    // Verify DevRequest Detail Modal opens
    const detailModal = devPage.getByRole("dialog");
    await expect(detailModal).toBeVisible();
    await expect(detailModal.getByText("E2E Provisioning Request")).toBeVisible();
    
    // 4. Promote to Project
    // Select Promote action
    await detailModal.getByRole("button", { name: /promote/i }).click();
    
    // It should open ProjectCreationModal (because it's the default or we might need to select)
    // Wait, let's assume it opens the target selection or defaults to Project/App
    // If it opens a form, fill it:
    // This depends on the UI implementation of DevRequestDetailModal...
    // Let's assume there's a button to "Create Application" or "Create Project"
    const promoteAppBtn = devPage.getByRole("button", { name: /새 어플리케이션/i });
    if (await promoteAppBtn.isVisible()) {
       await promoteAppBtn.click();
       // Fill application form
       await devPage.getByLabel(/application key/i).fill(`E2EAPP${Math.floor(Math.random() * 10000)}`);
       await devPage.getByLabel(/application name/i).fill("E2E Promoted App");
       await devPage.getByLabel(/owner/i).fill(SEEDED.developer.user_id);
       await devPage.getByLabel(/leader/i).fill(SEEDED.developer.user_id);
       await devPage.getByRole("button", { name: /create application/i }).click();
       // Wait for completion
       await expect(devPage.getByText(/registered/i).first()).toBeVisible({ timeout: 10000 });
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
