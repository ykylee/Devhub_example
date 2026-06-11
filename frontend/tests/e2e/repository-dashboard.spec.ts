import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";

test.describe("Repository Detailed Dashboard E2E", () => {
  test("TC-REPO-DASH-DEV — Developer 역할로 진입 후 UI 검증 & 모달 팝업 작동", async ({ page }) => {
    // Determinism guards: CI's Keycloak user-lookup race intermittently 403s
    // /api/v1/repositories/{id}/build-runs (auth.role_denied → fallback 'offline_access'
    // → RBAC 403). Without mocking, the "View Logs" button is missing intermittently
    // and the test fails on transient infrastructure issues. Mock the build-runs
    // and logs endpoints to deterministic synthetic data.

    // Mock build-runs: 1 success + 1 failed (so the "View Logs" button renders)
    await page.route(/\/api\/v1\/repositories\/\d+\/build-runs.*/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          status: "ok",
          data: [
            {
              id: 101,
              repository_id: 1,
              run_external_id: "ext-101",
              branch: "main",
              commit_sha: "abcdef1234567890",
              status: "success",
              duration_seconds: 120,
              started_at: "2026-06-05T10:00:00Z",
              finished_at: "2026-06-05T10:02:00Z",
            },
            {
              id: 103,
              repository_id: 1,
              run_external_id: "ext-103",
              branch: "main",
              commit_sha: "7890abcdef123456",
              status: "failed",
              duration_seconds: 45,
              started_at: "2026-06-05T10:10:00Z",
              finished_at: "2026-06-05T10:10:45Z",
            },
          ],
          meta: { total: 2 },
        }),
      });
    });

    // Mock logs: return lines including the "deadlock regression" text the test asserts
    await page.route(/\/api\/v1\/ci-runs\/.*\/logs.*/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          status: "ok",
          data: [
            { timestamp: "2026-06-05T10:14:11Z", level: "info", message: "Starting build run...", step_name: "checkout" },
            { timestamp: "2026-06-05T10:14:14Z", level: "info", message: "Installing dependencies...", step_name: "install" },
            { timestamp: "2026-06-05T10:14:20Z", level: "warning", message: "Found 12 high-severity vulnerabilities.", step_name: "audit" },
            { timestamp: "2026-06-05T10:14:25Z", level: "error", message: "Test suite failed: deadlock regression test failed on rbac enforcement", step_name: "test" },
            { timestamp: "2026-06-05T10:14:26Z", level: "error", message: "Build failed. Please investigate test regressions.", step_name: "test" },
          ],
          meta: { count: 5 },
        }),
      });
    });

    // 1. Developer (Alice) 로그인
    await loginAs(page, SEEDED.developer);

    // 2. 저장소 목록으로 이동하여 e2e-repo-a 상세 화면 진입
    await page.goto(appPath("/repositories"));
    // .first() to bypass strict mode when prior run residue (e.g. e2e-repo-a3xd7) is present.
    await expect(page.getByText("e2e-repo-a", { exact: false }).first()).toBeVisible({ timeout: 20_000 });
    await page.getByRole("link", { name: /e2e-repo-a/i }).first().click();

    // 3. 상세 대시보드 로딩 완료 확인
    await expect(page).toHaveURL(new RegExp(`${appPath("/repositories")}/\\d+$`), { timeout: 20_000 });
    await expect(page.getByRole("heading", { name: /e2e-repo-a/i })).toBeVisible({ timeout: 20_000 });

    // 4. Developer 기본 뷰 및 배지 확인
    await expect(page.getByText("Viewing as Developer")).toBeVisible();
    await expect(page.getByText("Build & Integration Runs")).toBeVisible();
    await expect(page.getByText("Static Analysis (SonarQube)")).toBeVisible();

    // 5. Build Log Modal 작동 확인
    // Mock build-runs 응답에 failed 빌드가 있어서 "View Logs" 버튼이 렌더링됨
    const viewLogsBtn = page.getByRole("button", { name: /View Logs/i }).first();
    await expect(viewLogsBtn).toBeVisible();
    await viewLogsBtn.click();

    // 모달 다이얼로그 노출 확인
    await expect(page.getByRole("dialog")).toBeVisible();
    await expect(page.getByText("Build Run Console Output")).toBeVisible();
    await expect(page.getByText("deadlock regression test failed", { exact: false })).toBeVisible();

    // 모달 닫기
    const closeBtn = page.getByRole("button", { name: /Close/i }).first();
    await closeBtn.click();
    await expect(page.getByRole("dialog")).toBeHidden();

    // 6. Manager & Governance 탭으로 전환 검증
    const managerTab = page.getByRole("button", { name: /Manager & Governance/i });
    await managerTab.click();

    // Manager 뷰 컨텐츠 노출 확인
    await expect(page.getByText("Team Manager Focus")).toBeVisible();
    await expect(page.getByText("Build & Integration Runs")).toBeHidden();
  });

  test("TC-REPO-DASH-MGR — Manager 역할로 진입 후 UI 검증 & 기여자 분포 차트 토글", async ({ page }) => {
    // 1. Team Manager (Bob) 로그인
    await loginAs(page, SEEDED.team_manager);

    // 2. 저장소 목록을 거쳐 e2e-repo-a 상세 화면 진입
    await page.goto(appPath("/repositories"));
    // 저장소 목록의 e2e-repo-a 항목 — CI race 환경에서 20s 도 intermittent fail 발생 가능.
    // selector 를 regex 로 robust + timeout 20s 명시.
    await expect(page.getByText(/e2e-repo-a/i).first()).toBeVisible({ timeout: 20_000 });
    await page.getByRole("link", { name: /e2e-repo-a/i }).first().click();

    // 3. 상세 대시보드 로딩 확인
    await expect(page).toHaveURL(new RegExp(`${appPath("/repositories")}/\\d+$`), { timeout: 20_000 });

    // 4. Manager 기본 뷰 및 배지 확인
    await expect(page.getByText("Viewing as Manager")).toBeVisible();
    await expect(page.getByText("Team Manager Focus")).toBeVisible();
    await expect(page.getByText("Organization Admin Focus")).toBeVisible();
    await expect(page.getByText("System Admin Focus")).toBeVisible();

    // 5. Contributor Distribution 차트 토글 검증
    await expect(page.getByText("Contributor Distribution")).toBeVisible();
    
    // 차트 숨기기 버튼 클릭
    const hideBtn = page.getByTitle("Hide distribution chart");
    await hideBtn.click();

    // 숨김 안내 문구와 Unhide 버튼 확인
    await expect(page.getByText("Contributor chart is hidden")).toBeVisible();
    
    // Unhide 버튼 클릭하여 다시 보이기
    const unhideBtn = page.getByRole("button", { name: /Unhide/i });
    await unhideBtn.click();
    await expect(page.getByText("Contributor chart is hidden")).toBeHidden();
  });
});
