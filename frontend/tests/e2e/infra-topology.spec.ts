import { test, expect, loginAs, SEEDED } from "./fixtures";

// infra-topology.spec — TC-INFRA-RENDER-01 (sprint claude/work_260513-j, C2 1차).
// Scope: 정적 데이터 렌더 검증만. TC-INFRA-NODE-CLICK-01 / TC-INFRA-GROUP-TOGGLE-01
// 등 인터랙션 TC 는 carve out (test_cases_m3_command_infra.md §4 의 spec ts 후보).
//
// 핵심 검증:
//   1. system_admin 으로 /admin 진입 시 "System Infrastructure" heading 노출
//   2. /api/v1/infra/topology 응답 데이터 기반 React Flow 노드가 렌더됨

test.describe("/admin — infra topology render (TC-INFRA-RENDER-01)", () => {
  test("system_admin sees the topology canvas with at least one node", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);

    await page.goto("/admin");
    await expect(page).toHaveURL(/\/admin/);

    // Heading from admin/page.tsx — public marker that the topology view rendered.
    await expect(page.getByText(/System.*Infrastructure/i).first()).toBeVisible({ timeout: 15_000 });

    // React Flow renders each node as div.react-flow__node. The seeded
    // topology contract (GET /api/v1/infra/topology, API-07) returns at
    // least one node; if it returns empty the assertion fails.
    const nodes = page.locator(".react-flow__node");
    await expect(nodes.first()).toBeVisible({ timeout: 15_000 });
  });
});
