import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";

// repository-kpi-tests-section.spec.ts
//
// Sprint A follow-up E2E (PR #597 follow-up, TC-REPO-KPI-TESTS-01).
// 시드 repo (e2e-repo-a) 의 Repository 상세 화면에 진입 → Manager 뷰의 KPI/Tests
// sub-section 이 정상 표시되고 window selector 전환 시 data refetch 가 일어나는지 검증.
//
// Determinism guards: 백엔드 / Keycloak 응답이 CI 환경에서 intermittent 으로
// 흔들리므로 KPI/test-results endpoint 를 page.route() 로 mock. 기존
// repository-dashboard.spec.ts 의 정공법 그대로.

test.describe("Repository KPI/Tests sub-section", () => {
  test("TC-REPO-KPI-TESTS-01 — Manager 진입 후 KPI/Tests sub-section metric 표시 + window selector refetch", async ({ page }) => {
    // KPI endpoint mock — 87.3 quality + 0.94 build_success + 3 open / 12 merged / 4 contributors
    await page.route(/\/api\/v1\/repositories\/\d+\/kpi.*/, async (route) => {
      const url = new URL(route.request().url());
      const window = url.searchParams.get("window") ?? "30d";
      // window 값에 따라 약간 변형 — refetch 검증용
      const pass = window === "7d" ? 0.99 : window === "90d" ? 0.88 : 0.94;
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
            build_success_rate: pass,
            build_run_count: 47,
            open_pr_count: 3,
            merged_pr_count: 12,
            active_contributor_count: 4,
          },
        }),
      });
    });

    // test-results endpoint mock — 42 success / 3 failed / 1 running / 1 cancelled
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
            totals: {
              success: 42,
              failed: 3,
              running: 1,
              cancelled: 1,
              skipped: 0,
              queued: 0,
              unknown: 0,
            },
            pass_rate: 42 / 45,
            recent: [
              { id: 100, run_external_id: "ext-100", commit_sha: "feedface1234", status: "success", branch: "main", started_at: "2026-06-15T01:00:00Z", finished_at: "2026-06-15T01:02:00Z" },
              { id: 99, run_external_id: "ext-99", commit_sha: "badfeed5678", status: "failed", branch: "feat/x", started_at: "2026-06-15T00:30:00Z", finished_at: "2026-06-15T00:31:00Z" },
            ],
          },
          meta: { total: 47, limit: 20 },
        }),
      });
    });

    // 1. Team Manager (Bob) 로그인 — Manager 가 KPI/Tests sub-section 노출
    await loginAs(page, SEEDED.team_manager);

    // 2. e2e-repo-a 상세 화면 진입 (저장소 목록 → link click)
    await page.goto(appPath("/repositories"));
    await expect(page.getByText("e2e-repo-a", { exact: false }).first()).toBeVisible({ timeout: 20_000 });
    await page.getByRole("link", { name: /e2e-repo-a/i }).first().click();
    await expect(page).toHaveURL(new RegExp(`${appPath("/repositories")}/\\d+$`), { timeout: 20_000 });
    await expect(page.getByRole("heading", { name: /e2e-repo-a/i })).toBeVisible({ timeout: 20_000 });

    // 3. KPI sub-section metric 표시 확인 (data-testid 기반)
    await expect(page.getByTestId("repository-kpi-section")).toBeVisible({ timeout: 10_000 });
    await expect(page.getByTestId("kpi-quality-score")).toHaveText("87.3", { timeout: 10_000 });
    await expect(page.getByTestId("kpi-build-success-rate")).toContainText("94.0%");
    await expect(page.getByTestId("kpi-open-pr")).toHaveText("3");
    await expect(page.getByTestId("kpi-merged-pr")).toHaveText("12");
    await expect(page.getByTestId("kpi-active-contributors")).toHaveText("4");

    // 4. Tests sub-section metric 표시 확인
    await expect(page.getByTestId("repository-tests-section")).toBeVisible();
    await expect(page.getByTestId("tests-pass-rate")).toContainText("93.3%");
    await expect(page.getByTestId("tests-status-success")).toContainText("42");
    await expect(page.getByTestId("tests-status-failed")).toContainText("3");
    // recent runs table 의 commit SHA short form
    const recentTable = page.getByTestId("tests-recent-table");
    await expect(recentTable).toBeVisible();
    await expect(recentTable).toContainText("feedfac"); // feedface1234.slice(0, 7)
    await expect(recentTable).toContainText("badfeed"); // badfeed5678.slice(0, 7)

    // 5. window selector 7d 전환 → KPI refetch 검증
    // 7d 일 때 build_success_rate = 0.99 → "99.0%" 노출
    const kpiWindowSelect = page.getByTestId("repository-kpi-section").getByLabel("Window");
    await kpiWindowSelect.selectOption("7");
    await expect(page.getByTestId("kpi-build-success-rate")).toContainText("99.0%", { timeout: 10_000 });

    // 6. window selector 90d 전환 → 0.88 → "88.0%"
    await kpiWindowSelect.selectOption("90");
    await expect(page.getByTestId("kpi-build-success-rate")).toContainText("88.0%", { timeout: 10_000 });
  });
});
