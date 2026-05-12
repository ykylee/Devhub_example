import { test, expect, loginAs, SEEDED } from "./fixtures";

// admin-permissions.spec — F8 권한 편집 smoke.
// TC-PERMISSIONS-SMOKE-01. M1 RBAC track (PR-G6) 으로 머지된
// PermissionEditor 가 e2e 검증 0건이었던 갭을 본 sprint 게이트에서 1건 보강.
// backend mutation 회피 — role 선택 후 PermissionMatrix 의 5 resource ×
// 4 action 매트릭스가 노출되는지만 확인.

test.describe("/admin/settings/permissions — PermissionEditor smoke", () => {
  test("TC-PERMISSIONS-SMOKE-01 — 진입 + role 선택 + matrix 노출", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto("/admin/settings/permissions");

    // Role 카드들이 hydrate 되어야 함. defaultRoles fallback 으로 인해
    // listPolicies() 가 실패해도 3개 카드는 보장.
    const devCard = page.getByRole("heading", { name: /^developer$/i });
    await expect(devCard).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole("heading", { name: /^manager$/i })).toBeVisible();
    await expect(page.getByRole("heading", { name: /^system admin$/i })).toBeVisible();

    // role 카드 클릭 (제목을 클릭하면 부모 motion.div 의 onClick 으로
    // setSelectedRoleId 가 발화).
    await devCard.click();

    // 오른쪽 패널이 "Developer Matrix" 헤더로 전환.
    await expect(page.getByRole("heading", { name: /developer matrix/i })).toBeVisible({ timeout: 5_000 });

    // PermissionMatrix 의 5 resource label.
    const resourceLabels = [
      /Infrastructure & Topology/i,
      /CI\/CD Pipelines/i,
      /Organization & Members/i,
      /Risk & Security/i,
      /Audit Logs & History/i,
    ];
    for (const label of resourceLabels) {
      await expect(page.getByText(label)).toBeVisible();
    }

    // 4 action 컬럼 헤더 — View / Create / Edit / Delete. role 카드의
    // Trash2 ("Delete role") 버튼 텍스트와 구분되도록 column 헤더는
    // 명시적으로 cell role 로 검증.
    for (const action of ["View", "Create", "Edit", "Delete"]) {
      await expect(page.getByRole("columnheader", { name: action })).toBeVisible();
    }

    // mutation 회피 — Save Permissions 버튼은 클릭하지 않는다. 카드 클릭
    // 자체는 setSelectedRoleId 만 호출하므로 backend 호출 없음.
  });
});
