import { test, expect, SEEDED, loginAs } from "./fixtures";

// DREQ E2E — TC 카탈로그는 docs/tests/test_cases_m5_dreq.md 참조.
// 본 spec 은 카탈로그의 E2E 영역 (TC-DREQ-ADMIN-TOKEN-01 / INTAKE-AUTH-01 /
// WIDGET-FLOW-01 / PROMOTE-TX-01 / ADMIN-TOKEN-REVOKE-01 / INTAKE-AUTH-NEG-03)
// 을 mega lifecycle test 한 케이스로 묶고, negative bearer (INTAKE-AUTH-NEG-01)
// 와 PATCH allowed_ips (ADMIN-TOKEN-PATCH-01) 는 별도 test 로 분리.

test.describe("DREQ E2E", () => {
  test("Intake to Promote to Revoke lifecycle", async ({ page, request, browser }) => {
    let plainToken: string | undefined;
    const clientLabel = `e2e_client_${Date.now()}`;
    const testSuffix = Date.now();
    const requestTitle = `E2E Provisioning Request ${testSuffix}`;
    const externalRef = `E2E-REQ-${testSuffix}`;

    await test.step("TC-DREQ-ADMIN-TOKEN-01 — issue intake token (plain 1회 노출)", async () => {
      await loginAs(page, SEEDED.systemAdmin);
      await page.goto("/admin/settings/dev-request-tokens");
      await expect(page.getByRole("heading", { name: /intake tokens/i })).toBeVisible();

      await page.getByRole("button", { name: /issue token/i }).click();
      await expect(page.getByRole("dialog")).toBeVisible();

      await page.getByLabel(/client label/i).fill(clientLabel);
      await page.getByLabel(/source system/i).fill("e2e_sys");
      // CI runner 는 IPv6 loopback (::1) 으로 intake 를 칠 수 있어 IPv4/IPv6 모두 허용.
      const allowedIpInputs = page.getByPlaceholder(/10\.0\.0\.0/i);
      await allowedIpInputs.first().fill("0.0.0.0/0");
      await page.getByRole("button", { name: /add ip\s*\/\s*cidr/i }).click();
      await expect(allowedIpInputs).toHaveCount(2);
      await allowedIpInputs.nth(1).fill("::1");

      await page.getByRole("dialog").getByRole("button", { name: /issue token/i }).click();

      const tokenModal = page.getByRole("dialog");
      await expect(tokenModal.getByText(/token shown once/i)).toBeVisible();
      await tokenModal.getByRole("button", { name: /show token/i }).click();
      const plainTokenCode = tokenModal.locator("code").first();
      // Token format 은 opaque (구현 정의) — masking 해제 확인 + 빈 값 아님 확인.
      await expect(plainTokenCode).not.toContainText("•");
      plainToken = (await plainTokenCode.textContent())?.trim();
      expect(plainToken).toBeTruthy();

      await page.getByRole("button", { name: /저장 완료 — 닫기/i }).click();
      await expect(page.getByRole("dialog")).toBeHidden();
    });

    await test.step("TC-DREQ-INTAKE-AUTH-01 — external POST with Bearer succeeds", async () => {
      const intakeResponse = await request.post("/api/v1/dev-requests", {
        headers: {
          Authorization: `Bearer ${plainToken}`,
        },
        data: {
          title: requestTitle,
          details: "Please provision a new project for E2E testing.",
          requester: "e2e_tester",
          assignee_user_id: SEEDED.developer.user_id,
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
      expect(intakeBody.data.id).toBeTruthy();
    });

    await test.step("TC-DREQ-WIDGET-FLOW-01 — assignee widget → list → detail modal", async () => {
      const devContext = await browser.newContext();
      const devPage = await devContext.newPage();
      try {
        await loginAs(devPage, SEEDED.developer);

        await devPage.goto("/developer");
        await expect(
          devPage.getByRole("heading", { name: /my dev requests|내 대기 의뢰/i })
        ).toBeVisible();

        const reqLink = devPage.getByText(requestTitle).first();
        await expect(reqLink).toBeVisible();
        await reqLink.click();
        await expect(devPage).toHaveURL(/\/dev-requests(\/|$)/);

        const reqRow = devPage.locator("tr").filter({ hasText: requestTitle }).first();
        await expect(reqRow).toBeVisible();
        await reqRow.click();

        const detailModal = devPage.getByRole("dialog");
        await expect(detailModal).toBeVisible();
        await expect(detailModal.getByText(requestTitle)).toBeVisible();
      } finally {
        await devContext.close();
      }
    });

    await test.step("TC-DREQ-PROMOTE-TX-01 — system_admin registers as application", async () => {
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
    });

    await test.step("TC-DREQ-ADMIN-TOKEN-REVOKE-01 — revoke via DestructiveConfirmModal", async () => {
      await page.goto("/admin/settings/dev-request-tokens");
      const row = page.getByRole("row").filter({ hasText: clientLabel });
      await expect(row).toBeVisible();
      await row.getByRole("button", { name: /revoke/i }).click();

      const confirmModal = page.getByRole("dialog");
      await expect(confirmModal.getByText(/revoke token/i)).toBeVisible();
      await confirmModal.getByRole("button", { name: /revoke/i, exact: true }).click();

      await expect(page.getByText(`토큰 '${clientLabel}' 이 revoke 되었습니다.`)).toBeVisible();
      await expect(row.getByText(/revoked/i)).toBeVisible();
    });

    await test.step("TC-DREQ-INTAKE-AUTH-NEG-03 — revoked token rejected with 401", async () => {
      const failResponse = await request.post("/api/v1/dev-requests", {
        headers: {
          Authorization: `Bearer ${plainToken}`,
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

  test("TC-DREQ-INTAKE-AUTH-NEG-01 — invalid bearer rejected with 401", async ({ request }) => {
    // 발급된 token 의 hash 와 충돌하지 않는 임의 base64url 문자열 (43 chars 길이 정합).
    const fakeBearer = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA";
    const response = await request.post("/api/v1/dev-requests", {
      headers: { Authorization: `Bearer ${fakeBearer}` },
      data: {
        title: "neg-01",
        details: "neg",
        requester: "neg",
        assignee_user_id: SEEDED.developer.user_id,
        external_ref: `neg-${Date.now()}`,
      },
    });
    expect(response.ok()).toBeFalsy();
    expect(response.status()).toBe(401);
  });

  test("TC-DREQ-ADMIN-TOKEN-PATCH-01 — PATCH allowed_ips updates token", async ({ page }) => {
    // 별도 token 발급 → PATCH 로 allowed_ips 갱신 → 검증.
    // admin endpoint 는 OIDC system_admin session 이 필요하므로 page.request 사용
    // (fixture-level `request` 는 storage state 를 공유하지 않음).
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto("/admin/settings/dev-request-tokens");

    const clientLabel = `patch_e2e_${Date.now()}`;

    // UI 로 token 발급 (modal flow 가 backend create handler 를 wire 하는 정상 경로).
    await page.getByRole("button", { name: /issue token/i }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
    await page.getByLabel(/client label/i).fill(clientLabel);
    await page.getByLabel(/source system/i).fill("patch_sys");
    const allowedIpInputs = page.getByPlaceholder(/10\.0\.0\.0/i);
    await allowedIpInputs.first().fill("10.0.0.0/24");
    await page.getByRole("dialog").getByRole("button", { name: /issue token/i }).click();
    await expect(page.getByText(/token shown once/i)).toBeVisible();
    await page.getByRole("button", { name: /저장 완료 — 닫기/i }).click();

    // token_id 추출 — IntakeTokenTable 의 data-token-id attribute 에서 직접
    // 읽음. page.request.get 의 OIDC session propagation 이 CI 에서 flaky
    // (codex hotfix 회피 패턴 — sprint claude/work_260518-m).
    const tokenRow = page.locator(`tr[data-token-id]`).filter({ hasText: clientLabel }).first();
    await expect(tokenRow).toBeVisible({ timeout: 10_000 });
    const tokenId = (await tokenRow.getAttribute("data-token-id")) ?? "";
    expect(tokenId).toBeTruthy();

    // PATCH allowed_ips & expires_at — page.request 의 OIDC session propagation 이 CI 에서
    // flaky. page.evaluate 의 browser context fetch 가 cookies + sessionStorage
    // 의 Bearer token (apiClient 와 동일 구조) 을 사용해 modal form submit 와
    // 같은 인증 보장 (sprint claude/work_260518-m hotfix #4+5).
    const futureDate = new Date(Date.now() + 86400000).toISOString(); // 1 day in the future
    const patchResult = await page.evaluate(async ({ id, expiresAt }) => {
      const accessToken = sessionStorage.getItem("devhub_access_token");
      const resp = await fetch(`/api/v1/dev-request-tokens/${id}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
        },
        body: JSON.stringify({
          allowed_ips: ["192.0.2.0/24", "203.0.113.5"],
          expires_at: expiresAt,
        }),
        credentials: "include",
      });
      const body = await resp.json();
      return { ok: resp.ok, status: resp.status, body };
    }, { id: tokenId, expiresAt: futureDate });
    expect(patchResult.ok).toBeTruthy();
    expect(patchResult.body.status).toBe("ok");
    // backend 가 dedup / canonicalize 할 수 있어 정확한 순서·길이 의존하지 않고
    // sorted equality 로 검증.
    expect([...patchResult.body.data.allowed_ips].sort()).toEqual(
      ["192.0.2.0/24", "203.0.113.5"].sort()
    );
    expect(patchResult.body.data.expires_at).toBeTruthy();
    const gotTime = new Date(patchResult.body.data.expires_at).getTime();
    const wantTime = new Date(futureDate).getTime();
    expect(Math.abs(gotTime - wantTime)).toBeLessThan(5000);

    // 정리 — 본 test 가 생성한 token revoke (cleanup, 동일 패턴).
    await page.evaluate(async (id: string) => {
      const accessToken = sessionStorage.getItem("devhub_access_token");
      await fetch(`/api/v1/dev-request-tokens/${id}`, {
        method: "DELETE",
        headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : {},
        credentials: "include",
      });
    }, tokenId);
  });
});
