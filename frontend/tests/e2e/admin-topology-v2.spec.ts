import { test, expect, SEEDED, loginAs, appPath } from "./fixtures";

// Infra topology v2 admin frontend (sprint claude/work_260518-n).
// TC 카탈로그는 docs/tests/test_cases_m4_integration.md §3.
// Backend: API-76 GET /api/v1/infra/services + API-78 GET /api/v1/infra/topology/v2
// (PR #139, sprint codex/next-step-20260516).
// Page: frontend/app/(dashboard)/admin/topology-v2/page.tsx.

test.describe("Infra topology v2 admin UI", () => {
  test("TC-INT-HOMELAB-03 — system_admin 이 /admin/topology-v2 접근 + 노드/서비스/snapshot 메타 렌더", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);

    // codex hotfix #8 P2 #2 (PR #155) — 이전 assertion 은 `.react-flow` shell 이
    // 항상 마운트되어 API-78 fetch 실패에도 PASS 되는 false positive 였음.
    // page.waitForResponse 로 API-78 호출 자체를 검증 + 응답 body 의 meta/data
    // 형식 검증 + node/service row 또는 명시 empty state 로 fetch 결과 반영 검증.
    const topologyResponsePromise = page.waitForResponse(
      (resp) =>
        resp.url().includes("/api/v1/infra/topology/v2") &&
        resp.request().method() === "GET",
      { timeout: 20_000 },
    );

    await page.goto(appPath("/admin/topology-v2"));

    // Header — v2 view 의 marker.
    await expect(page.getByRole("heading", { name: /topology.*v2/i })).toBeVisible({ timeout: 15_000 });

    // snapshot_at 메타 tag — `Last snapshot:` 텍스트가 헤더에 노출.
    await expect(page.getByText(/last snapshot:/i)).toBeVisible();

    // API-78 호출 검증 — fetch 실패 시 즉시 fail.
    const topologyResponse = await topologyResponsePromise;
    expect(topologyResponse.ok()).toBeTruthy();
    const topologyBody = await topologyResponse.json();
    expect(topologyBody.meta).toBeDefined();
    expect(typeof topologyBody.meta.snapshot_at).toBe("string");
    expect(topologyBody.data).toBeDefined();
    expect(Array.isArray(topologyBody.data.nodes)).toBeTruthy();
    expect(Array.isArray(topologyBody.data.services)).toBeTruthy();

    // fetch 결과를 UI 가 정확히 반영했는지 — node 가 0 건이면 명시 empty state,
    // 1건 이상이면 sidebar list 에 service row 노출 (또는 sidebar heading 의 count 가 양수).
    const nodeCount = topologyBody.data.nodes.length;
    if (nodeCount === 0) {
      await expect(page.getByText(/등록된 infra node 가 없습니다/i)).toBeVisible({ timeout: 5_000 });
    } else {
      // 1+ node — sidebar heading "Services (N)" 또는 empty 도 가능 (services 가 비어도 OK).
      await expect(page.getByRole("heading", { name: /^services/i })).toBeVisible();
    }
  });

  test("TC-INT-FRONTEND-TOPOLOGY-V2-NAV-01 — /admin 페이지의 v2 nav link → /admin/topology-v2", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto(appPath("/admin"));
    // /admin/page.tsx 헤더의 "Topology v2" Link (ArrowRight icon).
    const v2Link = page.getByRole("link", { name: /topology v2/i });
    await expect(v2Link).toBeVisible({ timeout: 15_000 });
    await v2Link.click();
    await page.waitForURL(/\/admin\/topology-v2(\/|$)/, { timeout: 10_000 });
    await expect(page.getByRole("heading", { name: /topology.*v2/i })).toBeVisible();
  });

  test("TC-INT-FRONTEND-TOPOLOGY-V2-RBAC-01 — non-system_admin 접근 시 default landing redirect", async ({ page }) => {
    await loginAs(page, SEEDED.developer);
    await page.goto(appPath("/admin/topology-v2"));
    // AuthGuard + role-routing 가 default landing 으로 redirect (developer → /developer).
    await page.waitForURL(/\/developer(\/|$)/, { timeout: 15_000 });
    await expect(page.getByRole("heading", { name: /topology.*v2/i })).toHaveCount(0);
  });
});
