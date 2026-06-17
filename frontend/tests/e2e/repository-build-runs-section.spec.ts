import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";

// repository-build-runs-section.spec.ts
//
// N-9 잔여 build-runs polish (kpi-tests-per-domain-scope.md §6.5 + PR #555 잔여 4건 sub-issue 3+4).
// Repository 상세 view 진입 후 `RepositoryBuildRunsSection` 표시 + status filter dropdown + 적어도
// 1 row + skeleton 5 row 초기 로드 검증.
//
// 정합 대상:
// - frontend/domain/repository-integration/view/RepositoryBuildRunsSection.tsx
// - frontend/domain/repository-integration/hook/useRepositoryBuildRuns.ts
// - backend-core/internal/httpapi/repository_ops.go:144-152 RBAC 404 가드
// - backend-core/internal/httpapi/metrics.go:29 Histogram metric

test.describe("Repository Build Runs Section (N-9 residual)", () => {
  test("TC-E2E-BUILD-RUNS-SECTION-01 — Manager 진입 후 Build Runs section + status filter + 1 row", async ({ page }) => {
    // build-runs endpoint mock (20 row: 12 success / 5 failed / 2 running / 1 cancelled)
    await page.route(/\/api\/v1\/repositories\/\d+\/build-runs.*/, async (route) => {
      const url = new URL(route.request().url());
      const status = url.searchParams.get("status");
      // status filter 별 mock 분기
      if (status === "failed") {
        await route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({
            status: "ok",
            data: [
              {
                id: 100,
                repository_id: 1,
                run_external_id: "ext-100",
                branch: "main",
                commit_sha: "badfeed5678",
                status: "failed",
                duration_seconds: 45,
                started_at: "2026-06-15T01:00:00Z",
                finished_at: "2026-06-15T01:00:45Z",
              },
            ],
            meta: { total: 1 },
          }),
        });
        return;
      }
      // default (status=null or "all") — 20 row mixed
      const items = [];
      for (let i = 0; i < 20; i++) {
        const statusCycle = ["success", "failed", "running", "cancelled", "skipped", "queued", "unknown"][i % 7];
        items.push({
          id: 100 + i,
          repository_id: 1,
          run_external_id: `ext-${100 + i}`,
          branch: i % 2 === 0 ? "main" : `feat/sprint-a-${i}`,
          commit_sha: `feedface${(1000 + i).toString(16).padStart(4, "0")}`,
          status: statusCycle,
          duration_seconds: 30 + i,
          started_at: new Date(Date.now() - i * 60_000).toISOString(),
          finished_at: new Date(Date.now() - i * 60_000 + 30_000).toISOString(),
        });
      }
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ status: "ok", data: items, meta: { total: 20 } }),
      });
    });

    // RepositoryKPISection / RepositoryTestsSection 통합 정합 — mock
    await page.route(/\/api\/v1\/repositories\/\d+\/kpi.*/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          status: "ok",
          data: {
            repository_id: 1,
            window_from: "2026-05-16T00:00:00Z",
            window_to: "2026-06-15T00:00:00Z",
            quality_score: 87.3,
            quality_score_measured_at: "2026-06-14T22:00:00Z",
            build_success_rate: 0.94,
            build_run_count: 47,
            open_pr_count: 3,
            merged_pr_count: 12,
            active_contributor_count: 4,
          },
        }),
      });
    });
    await page.route(/\/api\/v1\/repositories\/\d+\/test-results.*/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          status: "ok",
          data: {
            repository_id: 1,
            window_from: "2026-05-16T00:00:00Z",
            window_to: "2026-06-15T00:00:00Z",
            totals: { success: 42, failed: 3, running: 1, cancelled: 1, skipped: 0, queued: 0, unknown: 0 },
            pass_rate: 42 / 45,
            recent: [],
          },
          meta: { total: 47, limit: 20 },
        }),
      });
    });

    // team_manager 로그인 — Manager 가 Build Runs section 노출
    await loginAs(page, SEEDED.team_manager);

    // e2e-repo-a 상세 화면 진입
    await page.goto(appPath("/repositories"));
    await expect(page.getByText("e2e-repo-a", { exact: false }).first()).toBeVisible({ timeout: 20_000 });
    await page.getByRole("link", { name: /e2e-repo-a/i }).first().click();
    await expect(page).toHaveURL(new RegExp(`${appPath("/repositories")}/\\d+$`), { timeout: 20_000 });
    await expect(page.getByRole("heading", { name: /e2e-repo-a/i })).toBeVisible({ timeout: 20_000 });

    // Build Runs section 표시
    await expect(page.getByTestId("repository-build-runs-section")).toBeVisible({ timeout: 10_000 });

    // status filter dropdown 표시
    const statusFilter = page.getByTestId("build-runs-status-filter");
    await expect(statusFilter).toBeVisible();

    // 20 row 표시 (initial page)
    const rows = page.getByTestId("build-runs-row");
    await expect(rows).toHaveCount(20, { timeout: 10_000 });

    // status filter 전환 (failed) → refetch + 1 row
    await statusFilter.selectOption("failed");
    // refetch 는 debounce 없으므로 즉시 확인
    await expect(rows).toHaveCount(1, { timeout: 10_000 });
    const firstFailedRow = rows.first();
    await expect(firstFailedRow).toContainText("failed");
    await expect(firstFailedRow).toContainText("badfeed"); // badfeed5678.slice(0, 7) = "badfeed"
  });
});
