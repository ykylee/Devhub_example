import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";

test.describe("N-13 — Voc auto-routing (TC-INBOUND-SRC-01)", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
  });

  test("TC-INBOUND-SRC-01 — PATCH inbound_source → POST voc → auto_routed 검증", async ({ page }) => {
    const apiBasePath = appPath("/").replace(/\/$/, "");
    const suffix = Date.now().toString().slice(-6);
    const platformKey = `AGITEA${suffix}`;

    // 1. Create a platform (PATCH requires existing platform)
    const createResp = await page.evaluate(async ({ apiBasePath, platformKey, suffix }) => {
      const token = sessionStorage.getItem("devhub_access_token");
      const res = await fetch(`${apiBasePath}/api/v1/platforms`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        credentials: "include",
        body: JSON.stringify({
          key: platformKey,
          name: `E2E AutoRoute ${suffix}`,
          status: "active",
          visibility: "internal",
        }),
      });
      const body = await res.json();
      return { ok: res.ok, status: res.status, platformId: body?.data?.id };
    }, { apiBasePath, platformKey, suffix });

    expect(createResp.ok).toBeTruthy();
    expect(createResp.platformId).toBeTruthy();

    // 2. PATCH inbound_source_type=gitea + config
    const patchResp = await page.evaluate(async ({ apiBasePath, platformId }) => {
      const token = sessionStorage.getItem("devhub_access_token");
      const res = await fetch(`${apiBasePath}/api/v1/platforms/${platformId}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        credentials: "include",
        body: JSON.stringify({
          inbound_source_type: "gitea",
          inbound_source_config: '{"external_ref_pattern":"^GITEA-([0-9]+)$"}',
        }),
      });
      return { ok: res.ok, status: res.status };
    }, { apiBasePath, platformId: createResp.platformId });

    expect(patchResp.ok).toBeTruthy();

    // 3. POST voc with GITEA- external_ref + source_system=gitea
    const testReqTitle = `E2E AutoRoute Test ${suffix}`;
    const vocResp = await page.evaluate(async ({ apiBasePath, testReqTitle }) => {
      const token = sessionStorage.getItem("devhub_access_token");
      const res = await fetch(`${apiBasePath}/api/v1/dev-requests/GITEA-${Date.now()}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        credentials: "include",
        body: JSON.stringify({
          title: testReqTitle,
          details: "Auto-routing E2E test",
          source_system: "gitea",
        }),
      });
      const body = await res.json();
      return { ok: res.ok, status: res.status, body };
    }, { apiBasePath, testReqTitle });

    expect(vocResp.ok).toBeTruthy();
    expect(vocResp.status).toBe(201);

    // 4. 응답에 auto_routed 필드 검증
    const data = vocResp.body.data || vocResp.body;
    expect(data.status).toBe("routed");
  });

  test("TC-INBOUND-SRC-01-NEG — PATCH inbound_source + POST voc with no match → voc stage 유지", async ({ page }) => {
    const apiBasePath = appPath("/").replace(/\/$/, "");
    const suffix = Date.now().toString().slice(-6);
    const platformKey = `BNEG${suffix}`;

    // 1. Create platform with gitea inbound source
    const createResp = await page.evaluate(async ({ apiBasePath, platformKey, suffix }) => {
      const token = sessionStorage.getItem("devhub_access_token");
      const res = await fetch(`${apiBasePath}/api/v1/platforms`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        credentials: "include",
        body: JSON.stringify({
          key: platformKey,
          name: `E2E Neg ${suffix}`,
          status: "active",
          visibility: "internal",
        }),
      });
      const body = await res.json();
      return { ok: res.ok, platformId: body?.data?.id };
    }, { apiBasePath, platformKey, suffix });

    expect(createResp.ok).toBeTruthy();

    // 2. PATCH inbound_source_type=gitea
    await page.evaluate(async ({ apiBasePath, platformId }) => {
      const token = sessionStorage.getItem("devhub_access_token");
      await fetch(`${apiBasePath}/api/v1/platforms/${platformId}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        credentials: "include",
        body: JSON.stringify({
          inbound_source_type: "gitea",
          inbound_source_config: '{"external_ref_pattern":"^GITEA-([0-9]+)$"}',
        }),
      });
    }, { apiBasePath, platformId: createResp.platformId });

    // 3. POST voc with non-matching external_ref (RANDOM-123)
    const vocResp = await page.evaluate(async ({ apiBasePath }) => {
      const token = sessionStorage.getItem("devhub_access_token");
      const res = await fetch(`${apiBasePath}/api/v1/dev-requests/RANDOM-${Date.now()}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        credentials: "include",
        body: JSON.stringify({
          title: "No match test",
          details: "Should stay in received state",
          source_system: "manual",
        }),
      });
      const body = await res.json();
      return { ok: res.ok, status: res.status, body };
    }, { apiBasePath });

    expect(vocResp.ok).toBeTruthy();
    expect(vocResp.status).toBe(201);

    const data = vocResp.body.data || vocResp.body;
    expect(data.status).toBe("received");
  });
});
