import { test, expect, loginAs, SEEDED } from "./fixtures";

// infra-topology.spec — TC-INFRA-RENDER-01 (sprint claude/work_260513-j, C2 1차).
// Scope: 정적 렌더 surface 검증만 — \"system_admin 으로 /admin 진입 시 infra
// topology 캔버스가 마운트된다\". 노드/엣지 개수 + 메타데이터 검증은 seed 의존도가
// 높아 carve out. TC-INFRA-NODE-CLICK-01 / TC-INFRA-GROUP-TOGGLE-01 등 인터랙션
// TC 도 carve out (test_cases_m3_command_infra.md §4 의 spec ts 후보).

test.describe("/admin — infra topology render (TC-INFRA-RENDER-01)", () => {
  test("system_admin reaches /admin and the topology view mounts", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);

    await page.goto("/admin");
    await expect(page).toHaveURL(/\/admin/);

    // Heading from admin/page.tsx — public marker that the topology view
    // rendered. Node count + edge metadata is seed-dependent and carved
    // out for the follow-up sprint.
    await expect(page.getByText(/System.*Infrastructure/i).first()).toBeVisible({ timeout: 20_000 });
  });
});
