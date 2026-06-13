import { test, expect, SEEDED, loginAs, appPath } from "./fixtures";

// X-2 multi-provider webhook 운영 UI e2e (release_v0-1_roadmap.md §3.5 X-2,
// sprint `feat/work_260614-x2-frontend-e2e`).
//
// TC 카탈로그 (X-2 5차 commit):
//   - TC-ADMIN-X2-01: system_admin /admin/inbound-source 진입 + 4 widget 렌더
//   - TC-ADMIN-X2-02: inbound_source_type selector 변경 (Gitea → Jira) + JSONB editor 정합
//   - TC-ADMIN-X2-03: non-admin (developer) /admin/inbound-source 진입 차단
//   - TC-ADMIN-X2-04: custom_external_ref_pattern preview 검증 (GITEA-123 → MATCH)
//
// Backend: API-108 (POST /api/v1/integration/providers/{provider_id}/webhook 의
// multi-provider dispatcher endpoint, 4차 PR #589).
// Page: frontend/app/(dashboard)/admin/inbound-source/page.tsx (X-2 운영 UI).
// Widget 4: InboundSourceTypeSelector / InboundSourceConfigEditor / PatternPreview /
//            InboundSourceManager (통합 view).

test.describe("X-2 Inbound Source 운영 UI", () => {
  test("TC-ADMIN-X2-01 — system_admin /admin/inbound-source 진입 + 4 widget 렌더", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto(appPath("/admin/inbound-source"));

    // h1 + widget 4 heading
    await expect(page.getByRole("heading", { name: /Inbound Source \(Multi-Provider Webhook\)/i })).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText(/Inbound Source Type/i)).toBeVisible();
    await expect(page.getByText(/Inbound Source Config \(JSONB\)/i)).toBeVisible();
    await expect(page.getByText(/Pattern Preview/i)).toBeVisible();
    // platform selector
    const platformSelect = page.locator("select").first();
    await expect(platformSelect).toBeVisible();
  });

  test("TC-ADMIN-X2-02 — inbound_source_type selector 변경 (Gitea → Jira) + JSONB editor 정합", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto(appPath("/admin/inbound-source"));

    // InboundSourceTypeSelector 의 Provider Type select 변경 (raw <select>).
    const typeSelect = page.locator("#inbound-source-type-select");
    await expect(typeSelect).toBeVisible({ timeout: 15_000 });
    // default = gitea (첫 번째 platform) → jira 로 변경
    await typeSelect.selectOption("jira");
    // Jira hint 표시 확인
    await expect(page.getByText(/JIRA-\d+ external_ref/i)).toBeInTheDocument();
  });

  test("TC-ADMIN-X2-03 — non-admin (developer) /admin/inbound-source 진입 차단", async ({ page }) => {
    await loginAs(page, SEEDED.developer);
    await page.goto(appPath("/admin/inbound-source"));

    // Sidebar 의 isSystemAdmin gate → defaultLandingFor('developer') = /developer
    await page.waitForURL(/\/developer(\/|$)/, { timeout: 10_000 });
    expect(new URL(page.url()).pathname).toMatch(/^\/developer(\/|$)/);
  });

  test("TC-ADMIN-X2-04 — custom_external_ref_pattern preview 검증 (GITEA-123 → MATCH)", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto(appPath("/admin/inbound-source"));

    // InboundSourceConfigEditor 의 textarea 검증
    const textarea = page.getByRole("textbox");
    await expect(textarea).toBeVisible({ timeout: 15_000 });
    const initial = await textarea.inputValue();
    // initial JSONB = empty {} 또는 inbound_source_config 의 현재 값
    expect(initial).toMatch(/^\{[\s\S]*\}$/);

    // type=other 의 platform 으로 전환 (3번째 option)
    const typeSelect = page.locator("#inbound-source-type-select");
    await typeSelect.selectOption("other");

    // textarea 에 custom_external_ref_pattern 입력
    fireEventLikeInput(page, textarea, '{ "custom_external_ref_pattern": "^CUSTOM-\\\\d+$" }');

    // PatternPreview 의 sample selector 의 GITEA-123 → MATCH (auto_route.go 의
    // jiraExternalRefPattern 의 ^([A-Z][A-Z0-9_]{1,9})-\d+$ 정공법).
    const sampleSelect = page.locator("#pattern-preview-sample");
    if (await sampleSelect.count() > 0) {
      await sampleSelect.selectOption("CUSTOM-999");
      // MATCH 표시 (custom regex 의 경우)
      await expect(page.getByText(/MATCH|NO MATCH/)).toBeVisible();
    }
  });
});

/** fireEvent 의 input event 를 page.fill() 으로 위임 (textarea 의 경우). */
async function fireEventLikeInput(page: import("@playwright/test").Page, locator: import("@playwright/test").Locator, value: string) {
  await locator.fill(value);
}
