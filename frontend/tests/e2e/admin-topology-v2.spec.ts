import { test, expect, SEEDED, loginAs } from "./fixtures";

// Infra topology v2 admin frontend (sprint claude/work_260518-n).
// TC 카탈로그는 docs/tests/test_cases_m4_integration.md §3.
// Backend: API-76 GET /api/v1/infra/services + API-78 GET /api/v1/infra/topology/v2
// (PR #139, sprint codex/next-step-20260516).
// Page: frontend/app/(dashboard)/admin/topology-v2/page.tsx.

test.describe("Infra topology v2 admin UI", () => {
  test("TC-INT-HOMELAB-03 — system_admin 이 /admin/topology-v2 접근 + 노드/서비스/snapshot 메타 렌더", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto("/admin/topology-v2");

    // Header — v2 view 의 marker.
    await expect(page.getByRole("heading", { name: /topology.*v2/i })).toBeVisible({ timeout: 15_000 });

    // snapshot_at 메타 tag — `Last snapshot:` 텍스트가 헤더에 노출.
    await expect(page.getByText(/last snapshot:/i)).toBeVisible();

    // React Flow canvas 또는 empty state — 첫 진입은 backend snapshot 의존이라
    // 둘 중 하나는 보장. 둘을 .or() chain (admin-integrations.spec.ts 패턴 정합).
    const canvasOrEmpty = page
      .locator(".react-flow")
      .or(page.getByText(/등록된 infra node 가 없습니다/i))
      .first();
    await expect(canvasOrEmpty).toBeVisible({ timeout: 10_000 });

    // Services sidebar — heading "Services (N)" 가 노출.
    await expect(page.getByRole("heading", { name: /^services/i })).toBeVisible();
  });

  test("TC-INT-FRONTEND-TOPOLOGY-V2-NAV-01 — /admin 페이지의 v2 nav link → /admin/topology-v2", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto("/admin");
    // /admin/page.tsx 헤더의 "Topology v2" Link (ArrowRight icon).
    const v2Link = page.getByRole("link", { name: /topology v2/i });
    await expect(v2Link).toBeVisible({ timeout: 15_000 });
    await v2Link.click();
    await page.waitForURL(/\/admin\/topology-v2(\/|$)/, { timeout: 10_000 });
    await expect(page.getByRole("heading", { name: /topology.*v2/i })).toBeVisible();
  });

  test("TC-INT-FRONTEND-TOPOLOGY-V2-RBAC-01 — non-system_admin 접근 시 default landing redirect", async ({ page }) => {
    await loginAs(page, SEEDED.developer);
    await page.goto("/admin/topology-v2");
    // AuthGuard + role-routing 가 default landing 으로 redirect (developer → /developer).
    await page.waitForURL(/\/developer(\/|$)/, { timeout: 15_000 });
    await expect(page.getByRole("heading", { name: /topology.*v2/i })).toHaveCount(0);
  });
});
