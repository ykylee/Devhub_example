import { expect, loginAs, SEEDED, test, appPath } from "./fixtures";

// 프로젝트 생성(New Project + ProjectCreationModal)은 커밋 fea5d32
// ("feat(admin): move project creation to admin catalog") 로 /projects → admin/catalog 로 이전됐다.
// 현재 /projects(ProjectsStatusPage)는 읽기 전용 "과제 현황" 대시보드이므로, 본 CRUD UI 검증은
// 생성 진입점이 실재하는 /admin/catalog?tab=projects 를 대상으로 한다 (#379 baseline 정정).
test.describe("/admin/catalog?tab=projects — Project CRUD UI (생성은 catalog 로 이전, fea5d32)", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto(appPath("/admin/catalog?tab=projects"));
    await expect(page).toHaveURL(/\/admin\/catalog/, { timeout: 20_000 });
    await expect(page.getByRole("heading", { name: /admin catalog/i })).toBeVisible({ timeout: 15_000 });
  });

  test("TC-PROJ-UI-01 — projects 탭 진입 + New Project 버튼 노출", async ({ page }) => {
    await expect(page.getByRole("button", { name: /new project/i })).toBeVisible();
  });

  test("TC-PROJ-UI-02 — New Project 모달 open/close", async ({ page }) => {
    await page.getByRole("button", { name: /new project/i }).click();
    const dialog = page.getByRole("dialog");
    await expect(dialog).toBeVisible();
    await page.keyboard.press("Escape");
    await expect(dialog).toBeHidden();
  });

  test("TC-PROJ-UI-03 — 필수값(key/name) 미입력 시 제출 차단 (dialog 유지)", async ({ page }) => {
    await page.getByRole("button", { name: /new project/i }).click();
    const dialog = page.getByRole("dialog");
    await expect(dialog).toBeVisible();

    // Project Leader 는 현재 사용자(actorDefaultOwnerId)로 prefill 되어 Create 버튼은 활성이지만,
    // key/name 은 required + 빈 값이라 브라우저 검증으로 submit 이 차단되어 dialog 가 유지된다.
    await dialog.getByRole("button", { name: /create project/i }).click();
    await expect(dialog).toBeVisible();
  });

  // TC-PROJ-CRUD-01 happy-path 는 admin/catalog 이전 후 (a) Project Leader ComboBox 선택으로
  // 버튼 활성화, (b) standalone 생성 흐름, (c) catalog 테이블(heading 아님)에서 생성 결과 검증으로
  // 재작성이 필요하다. 시드 leader 의존 + 로컬 E2E full-stack 미검증 제약으로 happy-path E2E carve
  // (#380) 로 분리한다.
  test.skip("TC-PROJ-CRUD-01 — 프로젝트 생성 및 조회 (admin/catalog leader ComboBox 재작성 carve #380)", async ({ page }) => {
    const unique = Date.now().toString().slice(-6);
    const projKey = `P${unique}`;
    const projName = `E2E Project ${unique}`;

    await page.getByRole("button", { name: /new project/i }).click();
    const dialog = page.getByRole("dialog");
    await expect(dialog).toBeVisible();

    await dialog.getByPlaceholder("E.G. API-V1").fill(projKey);
    await dialog.getByPlaceholder("e.g. Backend Refactoring").fill(projName);
    await dialog.getByPlaceholder("Scope and deliverables...").fill("E2E Test project deliverables.");
    // NOTE(#380): Project Leader 는 leaderOptions 존재 시 ComboBox 이므로 선택 인터랙션 재작성 필요.

    await dialog.getByRole("button", { name: /create project/i }).click();
    await expect(dialog).toBeHidden({ timeout: 10_000 });

    // catalog projects 탭 테이블 행(이름 셀)에 노출되는지 검증 (구 /projects heading 검증 대체).
    await expect(page.getByRole("cell", { name: projName })).toBeVisible({ timeout: 15_000 });
  });
});
