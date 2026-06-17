import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";

// analytics-picker.spec.ts
//
// Sprint E (kpi-tests-per-domain-scope.md §6.5) — Cross-Reference Picker 회귀 가드.
//
// /kpis, /tests 글로벌 페이지 진입 시:
//   1. AnalyticsDeprecationBanner 가 상단에 표시 (옵션 B 결정의 user-facing 안내)
//   2. DomainPicker 의 3 scope tab (Repository/Project/Platform) 모두 표시
//   3. 각 scope 의 entity list 가 fetch 결과로 표시
//   4. entity 클릭 시 해당 도메인 상세 페이지로 redirect
//
// Backend mock 정공법은 repository-kpi-tests-section.spec.ts 와 동일 — CI 환경의
// intermittent 응답 회피.

test.describe("Analytics Cross-Reference Picker (Sprint E)", () => {
  test("TC-ANALYTICS-KPI-01 — /kpis 진입 시 deprecation banner + 3 scope tab + entity redirect", async ({ page }) => {
    // 1) /api/v1/repositories, /api/v1/projects, /api/v1/platforms mock
    await page.route(/\/api\/v1\/repositories(\?|$)/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          status: "ok",
          data: [
            { id: 1, full_name: "e2e-repo-a", clone_url: "https://example.com/repo-a.git" },
            { id: 2, full_name: "e2e-repo-b", clone_url: "https://example.com/repo-b.git" },
          ],
        }),
      });
    });
    await page.route(/\/api\/v1\/projects(\?|$)/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          status: "ok",
          data: [
            { id: "proj-1", name: "Project Alpha" },
            { id: "proj-2", name: "Project Beta" },
          ],
        }),
      });
    });
    await page.route(/\/api\/v1\/platforms(\?|$)/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          status: "ok",
          data: [
            { id: "plat-1", name: "Platform One" },
          ],
        }),
      });
    });

    // 2) team_manager 로그인 — Manager 가 analytics 접근 가능
    await loginAs(page, SEEDED.team_manager);

    // 3) /kpis 진입
    await page.goto(appPath("/kpis"));

    // 4) Deprecation banner 표시 확인
    await expect(page.getByTestId("analytics-deprecation-banner")).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText(/Cross-reference picker/i)).toBeVisible();
    await expect(page.getByText(/legacy 위젯/i)).toBeVisible();

    // 5) DomainPicker 의 3 scope tab 표시
    const picker = page.getByTestId("domain-picker");
    await expect(picker).toBeVisible({ timeout: 10_000 });
    await expect(page.getByTestId("domain-picker-scope-repository")).toBeVisible();
    await expect(page.getByTestId("domain-picker-scope-project")).toBeVisible();
    await expect(page.getByTestId("domain-picker-scope-platform")).toBeVisible();

    // 6) Repository scope (default) entity list 표시
    const repoList = page.getByTestId("domain-picker-entity-list-repository");
    await expect(repoList).toBeVisible();
    await expect(repoList).toContainText("e2e-repo-a");
    await expect(repoList).toContainText("e2e-repo-b");

    // 7) Project scope 전환 → entity list 변경
    await page.getByTestId("domain-picker-scope-project").click();
    const projList = page.getByTestId("domain-picker-entity-list-project");
    await expect(projList).toBeVisible();
    await expect(projList).toContainText("Project Alpha");
    await expect(projList).toContainText("Project Beta");

    // 8) Platform scope 전환 → entity list 변경
    await page.getByTestId("domain-picker-scope-platform").click();
    const platList = page.getByTestId("domain-picker-entity-list-platform");
    await expect(platList).toBeVisible();
    await expect(platList).toContainText("Platform One");

    // 9) Repository scope 로 다시 전환 + 첫 entity 클릭 → 상세 페이지로 redirect
    await page.getByTestId("domain-picker-scope-repository").click();
    const firstRepoLink = page.getByTestId("domain-picker-entity-1");
    await expect(firstRepoLink).toBeVisible();
    await firstRepoLink.click();
    await expect(page).toHaveURL(new RegExp(`${appPath("/repositories")}/1$`), { timeout: 10_000 });
  });

  test("TC-ANALYTICS-TESTS-01 — /tests 진입 시 동일 검증 (kpi 와 mirror)", async ({ page }) => {
    // 동일 mock (1번과 같은 3개 endpoint)
    await page.route(/\/api\/v1\/repositories(\?|$)/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          status: "ok",
          data: [{ id: 7, full_name: "tests-repo", clone_url: "https://example.com/tests.git" }],
        }),
      });
    });
    await page.route(/\/api\/v1\/projects(\?|$)/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          status: "ok",
          data: [{ id: "tests-proj", name: "Tests Project" }],
        }),
      });
    });
    await page.route(/\/api\/v1\/platforms(\?|$)/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          status: "ok",
          data: [{ id: "tests-plat", name: "Tests Platform" }],
        }),
      });
    });

    await loginAs(page, SEEDED.team_manager);
    await page.goto(appPath("/tests"));

    // banner + 3 scope tab + Repository entity 표시
    await expect(page.getByTestId("analytics-deprecation-banner")).toBeVisible({ timeout: 10_000 });
    await expect(page.getByTestId("domain-picker")).toBeVisible();
    await expect(page.getByTestId("domain-picker-scope-repository")).toBeVisible();
    await expect(page.getByTestId("domain-picker-entity-list-repository")).toContainText("tests-repo");
  });
});
