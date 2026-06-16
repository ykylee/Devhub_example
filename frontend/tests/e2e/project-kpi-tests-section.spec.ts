import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";

// project-kpi-tests-section.spec.ts
//
// Sprint B-Tests E2E (PR #627 follow-up, TC-PROJ-KPI-TESTS-01).
// 시드 project (SEEDED_PROJECT_ID, e.g. "DevHub Simulation Project") 의 Project
// 상세 화면에 진입 → Manager 뷰의 KPI/Tests sub-section 이 정상 표시되고
// window selector 전환 시 data refetch 가 일어나는지 검증.
//
// 정공법: PR #597 의 repository-kpi-tests-section.spec.ts 와 동일 — backend /
// Keycloak 응답이 CI 환경에서 intermittent 으로 흔들리므로 KPI/test-results
// endpoint 를 page.route() 로 mock. projectKPI + projectTestResults 2 endpoint
// 모두 mock (Sprint B 1차 + Sprint B-Tests 2 PR 의 회귀 가드).

test.describe("Project KPI/Tests sub-section", () => {
  test("TC-PROJ-KPI-TESTS-01 — Manager 진입 후 KPI/Tests sub-section metric 표시 + window selector refetch", async ({ page }) => {
    // projectKPI endpoint mock — 87.3 quality + 0.94 build_success + 3 linked repo
    await page.route(/\/api\/v1\/projects\/[^/]+\/kpi.*/, async (route) => {
      const url = new URL(route.request().url());
      const window = url.searchParams.get("window") ?? "30d";
      const pass = window === "7d" ? 0.99 : window === "90d" ? 0.88 : 0.94;
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          status: "ok",
          data: {
            project_id: "31b9e2cb-b1b0-466a-bb10-ea00ee1234a1",
            window_from: "2026-05-16T00:00:00Z",
            window_to: "2026-06-15T00:00:00Z",
            weighted_quality_score: 87.3,
            weighted_build_success_rate: pass,
            total_build_run_count: 156,
            weighted_open_pr_count: 7,
            weighted_merged_pr_count: 23,
            active_contributor_count: 12,
            linked_repository_count: 3,
            weighted_at: "2026-06-15T00:00:00Z",
          },
        }),
      });
    });

    // projectTestResults endpoint mock — 0.93 pass_rate + 145 success / 8 failed / 2 status / recent 2 row (multi-repo)
    await page.route(/\/api\/v1\/projects\/[^/]+\/test-results.*/, async (route) => {
      const url = new URL(route.request().url());
      const window = url.searchParams.get("window") ?? "30d";
      const pass = window === "7d" ? 0.97 : window === "90d" ? 0.88 : 0.93;
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          status: "ok",
          data: {
            project_id: "31b9e2cb-b1b0-466a-bb10-ea00ee1234a1",
            window_from: "2026-05-16T00:00:00Z",
            window_to: "2026-06-15T00:00:00Z",
            weighted_pass_rate: pass,
            totals: {
              success: 145, failed: 8, running: 1, cancelled: 2,
              skipped: 0, queued: 0, unknown: 0,
            },
            recent: [
              { id: 100, repository_id: 1, repository_full_name: "org/repo-a", run_external_id: "ext-100", commit_sha: "feedface1234", status: "success", branch: "main", started_at: "2026-06-15T01:00:00Z", finished_at: "2026-06-15T01:02:00Z" },
              { id: 99, repository_id: 2, repository_full_name: "org/repo-b", run_external_id: "ext-99", commit_sha: "badfeed5678", status: "failed", branch: "feat/x", started_at: "2026-06-15T00:30:00Z", finished_at: "2026-06-15T00:31:00Z" },
            ],
          },
          meta: { total: 156, limit: 20 },
        }),
      });
    });

    // 1. Team Manager (Bob) 로그인
    await loginAs(page, SEEDED.team_manager);

    // 2. project 상세 화면 진입 (URL 직접, 시드 project UUID)
    const SEEDED_PROJECT_ID = "31b9e2cb-b1b0-466a-bb10-ea00ee1234a1";
    await page.goto(appPath(`/projects/${SEEDED_PROJECT_ID}`));
    await expect(page).toHaveURL(new RegExp(`${appPath("/projects")}/${SEEDED_PROJECT_ID}$`), { timeout: 20_000 });

    // 3. Project KPI sub-section 표시 (Sprint B 1차, weighted)
    await expect(page.getByTestId("project-kpi-section")).toBeVisible({ timeout: 10_000 });
    await expect(page.getByTestId("project-kpi-quality-score")).toHaveText("87.3", { timeout: 10_000 });
    await expect(page.getByTestId("project-kpi-build-success-rate")).toContainText("94.0%");
    await expect(page.getByTestId("project-kpi-open-pr")).toHaveText("7");
    await expect(page.getByTestId("project-kpi-merged-pr")).toHaveText("23");
    await expect(page.getByTestId("project-kpi-linked-repos")).toHaveText("3");

    // 4. Project Tests sub-section 표시 (Sprint B-Tests, weighted + multi-repo)
    await expect(page.getByTestId("project-tests-section")).toBeVisible();
    await expect(page.getByTestId("project-tests-pass-rate")).toContainText("93.0%");
    await expect(page.getByTestId("project-tests-status-success")).toContainText("145");
    await expect(page.getByTestId("project-tests-status-failed")).toContainText("8");
    await expect(page.getByTestId("project-tests-status-running")).toContainText("1");
    // recent runs table — multi-repo repository_full_name 표시
    const recentTable = page.getByTestId("project-tests-recent-table");
    await expect(recentTable).toBeVisible();
    await expect(recentTable).toContainText("org/repo-a");
    await expect(recentTable).toContainText("org/repo-b");
    await expect(recentTable).toContainText("feedfac");
    await expect(recentTable).toContainText("badfeed");
    // repository selector — repositoryFullName 노출 (data-testid 기반)
    await expect(page.getByTestId("project-tests-recent-repo-100")).toContainText("org/repo-a");
    await expect(page.getByTestId("project-tests-recent-repo-99")).toContainText("org/repo-b");

    // 5. Tests sub-section window selector 7d 전환 → projectTestResults refetch (0.97 → "97.0%")
    const testsWindowSelect = page.getByTestId("project-tests-section").getByLabel("Window");
    await testsWindowSelect.selectOption("7d");
    await expect(page.getByTestId("project-tests-pass-rate")).toContainText("97.0%", { timeout: 10_000 });
  });
});
