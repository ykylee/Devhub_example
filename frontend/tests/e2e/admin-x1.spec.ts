import { test, expect, SEEDED, loginAs, appPath } from "./fixtures";

// X-1 System Admin 운영 대시보드 e2e (release_v0-1_roadmap.md line 198, RM-M4-07,
// sprint `feat/work_260614-x1-frontend-e2e`).
//
// TC 카탈로그 (5차 commit):
//   - TC-ADMIN-X1-01: system_admin /admin 진입 + X-1 widget 4 표시
//   - TC-ADMIN-X1-02: sync job status API-106 호출 검증 + DashboardSummary 위젯의
//     success rate 계산 검증
//   - TC-ADMIN-X1-03: non-admin (developer) /admin 진입 차단 → /developer landing
//     redirect (defaultLandingFor(role))
//
// Backend: API-104/105/106 (PR #583, sprint `feat/work_260614-x1-system-admin-dashboard`).
// Page: frontend/app/(dashboard)/admin/page.tsx (X-1 widget 4 + 2x2 grid).

test.describe("X-1 System Admin 운영 대시보드", () => {
  test("TC-ADMIN-X1-01 — system_admin /admin 진입 + X-1 widget 4 렌더", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);

    // API-106 (summary) 호출 capture — page 가 dashboard summary 위젯을
    // 위해 status summary 를 fetch 한다.
    const summaryResponsePromise = page.waitForResponse(
      (resp) =>
        resp.url().includes("/api/v1/admin/integrations/summary") &&
        resp.request().method() === "GET",
      { timeout: 15_000 },
    );

    await page.goto(appPath("/admin"));

    // 4 widget heading 모두 visible — page.tsx 의 2x2 grid 배치.
    await expect(page.getByRole("heading", { name: /sync job queue/i })).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole("heading", { name: /sync job status/i })).toBeVisible();
    await expect(page.getByRole("heading", { name: /provider health/i })).toBeVisible();
    await expect(page.getByRole("heading", { name: /dashboard summary/i })).toBeVisible();

    // 운영 도구 link (Topology v2, Settings) — admin page 의 '운영 도구' 섹션.
    await expect(page.getByRole("link", { name: /topology v2/i })).toBeVisible();
    await expect(page.getByRole("link", { name: /settings/i })).toBeVisible();

    // API-106 fetch 검증 (summary 위젯 마운트 후 호출).
    const summaryResponse = await summaryResponsePromise;
    expect(summaryResponse.ok()).toBeTruthy();
    const summaryBody = await summaryResponse.json();
    expect(summaryBody.sync_job_status_counts).toBeDefined();
    expect(typeof summaryBody.sync_job_status_counts.queued).toBe("number");
    expect(typeof summaryBody.sync_job_status_counts.running).toBe("number");
    expect(typeof summaryBody.sync_job_status_counts.succeeded).toBe("number");
    expect(typeof summaryBody.sync_job_status_counts.failed).toBe("number");
  });

  test("TC-ADMIN-X1-02 — sync job status API-106 응답 → DashboardSummary 위젯의 totalJobs/queueDepth/failed/successRate 표시", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);

    // API mock — known counts 로 success rate 계산 검증.
    const fixtureCounts = { queued: 1, running: 2, succeeded: 7, failed: 1 };
    const totalJobs = fixtureCounts.queued + fixtureCounts.running + fixtureCounts.succeeded + fixtureCounts.failed;
    const queueDepth = fixtureCounts.queued + fixtureCounts.running;
    const completed = fixtureCounts.succeeded + fixtureCounts.failed;
    const expectedSuccessRate = Math.round((fixtureCounts.succeeded / completed) * 1000) / 10;

    await page.route("**/api/v1/admin/integrations/summary", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ sync_job_status_counts: fixtureCounts }),
      });
    });

    await page.goto(appPath("/admin"));

    // DashboardSummary 위젯의 stat label + 값 매칭.
    const summaryCard = page.locator("text=Dashboard Summary").locator("..");
    await expect(summaryCard).toBeVisible({ timeout: 15_000 });
    await expect(summaryCard.getByText(String(totalJobs))).toBeVisible(); // Total Jobs
    await expect(summaryCard.getByText(String(queueDepth))).toBeVisible(); // Queue Depth
    await expect(summaryCard.getByText(String(fixtureCounts.failed))).toBeVisible(); // Failed
    await expect(summaryCard.getByText(`${expectedSuccessRate}%`)).toBeVisible(); // Success Rate
  });

  test("TC-ADMIN-X1-03 — non-admin (developer) /admin 진입 차단 → /developer landing redirect", async ({ page }) => {
    await loginAs(page, SEEDED.developer);

    await page.goto(appPath("/admin"));

    // defaultLandingFor(role) — developer 의 landing = /developer
    await page.waitForURL(/\/developer(\/|$)/, { timeout: 10_000 });
    expect(new URL(page.url()).pathname).toMatch(/^\/developer(\/|$)/);
  });
});
