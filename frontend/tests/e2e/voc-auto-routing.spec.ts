import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";

test.describe("N-13 — Voc auto-routing (TC-INBOUND-SRC-01)", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
  });

  test("TC-INBOUND-SRC-01 — PATCH inbound_source → POST voc → auto_routed 검증", async ({ page }) => {
    const apiBasePath = appPath("/").replace(/\/$/, "") || "";
    const suffix = Date.now().toString().slice(-6);
    const platformKey = `AGITEA${suffix}`;

    // 1. Create a platform (PATCH requires existing platform) via page.request (base URL 자동 정합)
    const createResp = await page.request.post(`${apiBasePath}/api/v1/platforms`, {
      data: {
        key: platformKey,
        name: `E2E AutoRoute ${suffix}`,
        status: "active",
        visibility: "internal",
      },
    });
    const createBody = await createResp.json();
    expect(createResp.ok()).toBeTruthy();
    const platformId = createBody?.data?.id;
    expect(platformId).toBeTruthy();

    // 2. PATCH inbound_source_type=gitea + config
    const patchResp = await page.request.patch(`${apiBasePath}/api/v1/platforms/${platformId}`, {
      data: {
        inbound_source_type: "gitea",
        inbound_source_config: '{"external_ref_pattern":"^GITEA-([0-9]+)$"}',
      },
    });
    expect(patchResp.ok()).toBeTruthy();

    // 3. POST voc with GITEA- external_ref + source_system=gitea
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

    // 4. 응답에 auto_routed 필드 검증
    const data = vocBody.data || vocBody;
    expect(data.status).toBe("routed");
  });

  test("TC-INBOUND-SRC-01-NEG — PATCH inbound_source + POST voc with no match → voc stage 유지", async ({ page }) => {
    const apiBasePath = appPath("/").replace(/\/$/, "") || "";
    const suffix = Date.now().toString().slice(-6);
    const platformKey = `BNEG${suffix}`;

    // 1. Create platform with gitea inbound source
    const createResp = await page.request.post(`${apiBasePath}/api/v1/platforms`, {
      data: {
        key: platformKey,
        name: `E2E Neg ${suffix}`,
        status: "active",
        visibility: "internal",
      },
    });
    expect(createResp.ok()).toBeTruthy();
    const createBody = await createResp.json();
    const platformId = createBody?.data?.id;
    expect(platformId).toBeTruthy();

    // 2. PATCH inbound_source_type=gitea
    const patchResp = await page.request.patch(`${apiBasePath}/api/v1/platforms/${platformId}`, {
      data: {
        inbound_source_type: "gitea",
        inbound_source_config: '{"external_ref_pattern":"^GITEA-([0-9]+)$"}',
      },
    });
    expect(patchResp.ok()).toBeTruthy();

    // 3. POST voc with non-matching external_ref (RANDOM-123)
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
