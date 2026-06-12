import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";

// seed platform id (global-setup 의 DevHub Simulation App fixture).
// sprint plan v2 정공법: POST platform 단계 제거 (PR #576 commit 90256ec 정합) —
// global-setup 의 seed platform 'DevHub Simulation App' 만 사용.
const SEED_PLATFORM_ID = "e8a9bc11-a89c-4cb1-8071-8890ab2345ef";

test.describe("N-13 — Voc auto-routing (TC-INBOUND-SRC-01)", () => {
  // PATCH inbound_source 를 beforeAll hook 으로 이동 (Keycloak startup race 회피 정공법).
  //
  // PR #579 (1차 commit) 의 E2E shard 3/3 fail 2건 정공법:
  // - shard 3/3 의 Keycloak container 가 initial start-up race 로 늦게 ready
  //   (GitHub Actions runner 의 transient network issue, curl: (56) Recv failure 9 회)
  // - shard 1/2 에서는 beforeEach 의 loginAs 후 즉시 PATCH 가 PASS
  // - shard 3/3 에서는 loginAs 가 systemAdmin token 받기 전에 backend 401 → patchResp 4xx
  //
  // 정공법:
  // - beforeAll 에서 1회만 PATCH (Keycloak ready 보장)
  // - retry 3회 with backoff (5s + 10s + 15s)
  // - 2 test case 의 PATCH 단계 제거 (정합 검증만)
  test.beforeAll(async ({ browser }) => {
    const context = await browser.newContext();
    const page = await context.newPage();
    const apiBasePath = appPath("/").replace(/\/$/, "") || "";

    // loginAs 로 systemAdmin token 획득 (cookie + actor set).
    await loginAs(page, SEEDED.systemAdmin);

    // PATCH inbound_source on seed platform (retry 3 회 with backoff).
    const delays = [0, 5000, 10000, 15000]; // 0, 5s, 10s, 15s (4 attempts max)
    let lastErr: Error | undefined;
    for (const delay of delays) {
      if (delay > 0) {
        await page.waitForTimeout(delay);
      }
      try {
        const patchResp = await page.request.patch(
          `${apiBasePath}/api/v1/platforms/${SEED_PLATFORM_ID}`,
          {
            data: {
              inbound_source_type: "gitea",
              inbound_source_config: '{"external_ref_pattern":"^GITEA-([0-9]+)$"}',
            },
          }
        );
        if (patchResp.ok()) {
          await context.close();
          return; // success — early exit
        }
        lastErr = new Error(`PATCH inbound_source status=${patchResp.status()} body=${await patchResp.text()}`);
      } catch (err) {
        lastErr = err instanceof Error ? err : new Error(String(err));
      }
    }
    await context.close();
    throw new Error(
      `beforeAll PATCH inbound_source failed after ${delays.length} attempts (Keycloak startup race). last error: ${lastErr?.message ?? "unknown"}`
    );
  });

  test.beforeEach(async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
  });

  test("TC-INBOUND-SRC-01 — POST voc with external_ref pattern → auto_routed 검증", async ({ page }) => {
    const apiBasePath = appPath("/").replace(/\/$/, "") || "";
    const suffix = Date.now().toString().slice(-6);

    // POST voc with GITEA- external_ref + source_system=gitea
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

    // 응답에 auto_routed 필드 검증
    const data = vocBody.data || vocBody;
    expect(data.status).toBe("routed");
  });

  test("TC-INBOUND-SRC-01-NEG — POST voc with no match → voc stage 유지", async ({ page }) => {
    const apiBasePath = appPath("/").replace(/\/$/, "") || "";
    const suffix = Date.now().toString().slice(-6);

    // POST voc with non-matching external_ref (RANDOM-{ts})
    const vocResp = await page.request.post(
      `${apiBasePath}/api/v1/dev-requests/RANDOM-${Date.now()}`,
      {
        data: {
          title: `No match test ${suffix}`,
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
