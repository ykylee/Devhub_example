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
    // CI runner can hit intake via IPv6 loopback(::1), so include both IPv4/IPv6 allow entries.
    const allowedIpInputs = page.getByPlaceholder(/10\.0\.0\.0/i);
    await allowedIpInputs.first().fill("0.0.0.0/0");
    await page.getByRole("button", { name: /add ip\s*\/\s*cidr/i }).click();
    await expect(allowedIpInputs).toHaveCount(2);
    await allowedIpInputs.nth(1).fill("::1");
    
    // Submit
    await page.getByRole("dialog").getByRole("button", { name: /issue token/i }).click();

    // Reveal phase: extract the issued token from the modal in a deterministic way.
    const tokenModal = page.getByRole("dialog");
    await expect(tokenModal.getByText(/token shown once/i)).toBeVisible();
    await tokenModal.getByRole("button", { name: /show token/i }).click();
    const plainTokenCode = tokenModal.locator("code").first();
    // Token format is opaque (implementation-defined), so assert visibility/non-masked value.
    await expect(plainTokenCode).not.toContainText("•");
    const plainToken = (await plainTokenCode.textContent())?.trim();
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
    if (!intakeResponse.ok()) {
      const intakeErrorBody = await intakeResponse.text();
      throw new Error(
        `intake failed: status=${intakeResponse.status()} body=${intakeErrorBody}`
      );
    }
    const intakeBody = await intakeResponse.json();
    expect(intakeBody.status).toBe("ok");
    const dreqId = intakeBody.data.id;

    // 3. Assignee (developer) logs in and verifies visibility on dashboard/list
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
    
    // Developer context validation complete.
    await devContext.close();

    // 4. Register step (successor of old "promote") is mandatory and system_admin-only.
    await page.goto("/admin/settings/dev-requests");
    const adminRow = page.locator("tr").filter({ hasText: requestTitle }).first();
    await expect(adminRow).toBeVisible();
    await adminRow.click();
    const adminDetailModal = page.getByRole("dialog");
    await expect(adminDetailModal).toBeVisible();

    const registerAppBtn = adminDetailModal.getByRole("button", { name: /register as application/i });
    await expect(registerAppBtn).toBeVisible();
    await registerAppBtn.click();
    await page.getByPlaceholder(/application id \(uuid\)/i).fill(`app-e2e-${Date.now()}`);
    await page.getByRole("button", { name: /confirm/i }).click();

    // 5. System Admin revokes the token
    // Return to token management page before revoke
    await page.goto("/admin/settings/dev-request-tokens");
    const row = page.getByRole("row").filter({ hasText: clientLabel });
    await expect(row).toBeVisible();
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
