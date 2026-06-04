import { test, expect, loginAs, SEEDED, appPath } from "./fixtures";

/**
 * rbac-data-scope.spec.ts
 * release_v1_roadmap §3.5 N-10 P1 follow-up — RBAC E2E (frontend) follow-up.
 *
 * 매핑 TC (docs/domain/rbac-permissions/test_cases.md):
 *   - TC-RBAC-LOGOUT-02 — `/auth/signout`이 logout API orchestration route로 동작 (REQ-RBAC-012A)
 *   - TC-RBAC-ROW-READ-01 — List read scope 필터 (REQ-RBAC-013,015)
 *   - TC-RBAC-ROW-READ-02 — Get read scope 차단 (REQ-RBAC-013,015)
 *   - TC-RBAC-CODE-01 — 거부 코드 표준화 (REQ-RBAC-014)
 *
 * 잔여 (이 spec scope out):
 *   - TC-RBAC-LOGOUT-01 — backend IT (Claude 별도 sprint)
 *   - TC-RBAC-ROLE-DRIFT-01 — backend IT (Keycloak drift 환경 의존)
 *   - TC-RBAC-TRACE-01 — process/review (이 spec 의 TC ID 가 spec 문서와 1:1 매핑됨을 본 header 주석으로 입증)
 */

const SEEDED_PROJECT_ID = "31b9e2cb-b1b0-466a-bb10-ea00ee1234a1";
const SEEDED_PROJECT_NAME = "DevHub Simulation Project";

test.describe("RBAC data scope + logout (N-10 P1 follow-up)", () => {
  test("TC-RBAC-LOGOUT-02 — /auth/signout triggers POST /api/v1/auth/logout", async ({ page }) => {
    await loginAs(page, SEEDED.developer);

    const logoutRequests: string[] = [];
    await page.route(/\/api\/v1\/auth\/logout$/, async (route) => {
      logoutRequests.push(route.request().method());
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ status: "ok", data: { revoked: true } }),
      });
    });

    await page.goto(appPath("/auth/signout"));

    await expect
      .poll(() => logoutRequests.length, { timeout: 15_000, intervals: [200, 500] })
      .toBeGreaterThanOrEqual(1);

    const methods = logoutRequests.join(",");
    expect(methods).toContain("POST");
  });

  test("TC-RBAC-ROW-READ-01 — developer sees no projects they are not a member of", async ({ page }) => {
    await loginAs(page, SEEDED.developer);

    await page.goto(appPath("/projects"));

    await expect(page).toHaveURL(/\/projects/, { timeout: 10_000 });

    const seededProjectCard = page.getByText(SEEDED_PROJECT_NAME, { exact: true });
    await expect(seededProjectCard).toHaveCount(0);
  });

  test("TC-RBAC-ROW-READ-02 — developer cannot view project detail they are not a member of", async ({ page }) => {
    await loginAs(page, SEEDED.developer);

    const responsePromise = page.waitForResponse(
      (resp) =>
        resp.url().includes(`/api/v1/projects/${SEEDED_PROJECT_ID}`) &&
        resp.request().method() === "GET",
      { timeout: 15_000 },
    );

    await page.goto(appPath(`/projects/${SEEDED_PROJECT_ID}`));

    const response = await responsePromise;
    expect(response.status()).toBeGreaterThanOrEqual(400);
  });

  test("TC-RBAC-CODE-01 — denial response body uses standardized auth.* code", async ({ page }) => {
    await loginAs(page, SEEDED.developer);

    const responsePromise = page.waitForResponse(
      (resp) =>
        resp.url().includes(`/api/v1/projects/${SEEDED_PROJECT_ID}`) &&
        resp.request().method() === "GET" &&
        !resp.url().includes("/repositories") &&
        !resp.url().includes("/activity"),
      { timeout: 15_000 },
    );

    await page.goto(appPath(`/projects/${SEEDED_PROJECT_ID}`));

    const response = await responsePromise;
    expect(response.status()).toBe(403);

    const body = (await response.json()) as { code?: string; status?: string; error?: { code?: string } };
    const code = body?.code ?? body?.error?.code;
    expect(code, `expected standardized auth.* code, got body=${JSON.stringify(body)}`).toMatch(/^auth_/);
  });
});
