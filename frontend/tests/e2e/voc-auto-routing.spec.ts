import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";

test.describe("N-13 — Voc auto-routing (TC-INBOUND-SRC-01)", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
  });

  test("TC-INBOUND-SRC-01 — PATCH inbound_source → POST voc → auto_routed 검증", async ({ page }) => {
    const apiBasePath = appPath("/").replace(/\/$/, "") || "";
    const suffix = Date.now().toString().slice(-6);

    // 1. PATCH inbound_source on existing seed platform (global-setup 의 DevHub Simulation App)
    const seedPlatformId = "e8a9bc11-a89c-4cb1-8071-8890ab2345ef";
    const patchResp = await page.request.patch(`${apiBasePath}/api/v1/platforms/${seedPlatformId}`, {
      data: {
        inbound_source_type: "gitea",
        inbound_source_config: '{"external_ref_pattern":"^GITEA-([0-9]+)$"}',
      },
    });
    expect(patchResp.ok()).toBeTruthy();

    // 2. POST voc with GITEA- external_ref + source_system=gitea
    const testReqTitle = `E2E AutoRoute Test ${suffix}`;
    const vocResp = await page.request.post(
      `${apiBasePath}/api/v1/dev-requests/GITEA-${Date.now()}`,
      {
        data: {
          title: testReqTitle,
          details: "Auto-routing E2E test",
          source_system: "gitea",
        },
      }
    );
    expect(vocResp.ok()).toBeTruthy();
    const vocBody = await vocResp.json();

    // 3. 응답에 auto_routed 필드 검증
    const data = vocBody.data || vocBody;
    expect(data.status).toBe("routed");
  });

  test("TC-INBOUND-SRC-01-NEG — PATCH inbound_source + POST voc with no match → voc stage 유지", async ({ page }) => {
    const apiBasePath = appPath("/").replace(/\/$/, "") || "";
    const suffix = Date.now().toString().slice(-6);

    // 1. PATCH inbound_source_type=gitea on seed platform
    const seedPlatformId = "e8a9bc11-a89c-4cb1-8071-8890ab2345ef";
    const patchResp = await page.request.patch(`${apiBasePath}/api/v1/platforms/${seedPlatformId}`, {
      data: {
        inbound_source_type: "gitea",
        inbound_source_config: '{"external_ref_pattern":"^GITEA-([0-9]+)$"}',
      },
    });
    expect(patchResp.ok()).toBeTruthy();

    // 2. POST voc with non-matching external_ref (RANDOM-{ts})
    const vocResp = await page.request.post(
      `${apiBasePath}/api/v1/dev-requests/RANDOM-${Date.now()}`,
      {
        data: {
          title: "No match test",
          details: "Should stay in received state",
          source_system: "manual",
        },
      }
    );
    expect(vocResp.ok()).toBeTruthy();
    const vocBody = await vocResp.json();

    const data = vocBody.data || vocBody;
    expect(data.status).toBe("received");
  });
});
