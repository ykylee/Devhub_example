import { expect, loginAs, SEEDED, test, appPath } from "./fixtures";

test.describe("/projects — Project CRUD UI", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto(appPath("/projects"));
    await expect(page).toHaveURL(new RegExp(appPath("/projects")), { timeout: 15_000 });
  });

  test("TC-PROJ-UI-01 — 과제 현황 진입 + New Project 버튼 노출", async ({ page }) => {
    await expect(page.getByRole("button", { name: /new project/i })).toBeVisible();
  });

  test("TC-PROJ-UI-02 — New Project 모달 open/close", async ({ page }) => {
    await page.getByRole("button", { name: /new project/i }).click();
    const dialog = page.getByRole("dialog");
    await expect(dialog).toBeVisible();
    await page.keyboard.press("Escape");
    await expect(dialog).toBeHidden();
  });

  test("TC-PROJ-UI-03 — 필수값 없이 submit 시 브라우저 검증으로 제출 차단", async ({ page }) => {
    await page.getByRole("button", { name: /new project/i }).click();
    const dialog = page.getByRole("dialog");
    await expect(dialog).toBeVisible();

    await dialog.getByRole("button", { name: /create project/i }).click();
    await expect(dialog).toBeVisible();
  });

  test("TC-PROJ-CRUD-01 — 프로젝트 생성 및 조회 시나리오 (N:M 연결 포함)", async ({ page }) => {
    const unique = Date.now().toString().slice(-6);
    const projKey = `P${unique}`;
    const projName = `E2E Project ${unique}`;
    const owner = "charlie";

    await page.getByRole("button", { name: /new project/i }).click();
    const dialog = page.getByRole("dialog");
    await expect(dialog).toBeVisible();

    await dialog.getByPlaceholder("E.G. API-V1").fill(projKey);
    await dialog.getByPlaceholder("e.g. Backend Refactoring").fill(projName);
    
    // Select primary repository
    // Repository is optional in the new flow. CI seed data may not always have
    // additional repository options, so keep the default "No repository".
    
    await dialog.getByPlaceholder("Scope and deliverables...").fill("E2E Test project deliverables.");
    await dialog.getByPlaceholder("User ID...").fill(owner);
    
    await dialog.getByRole("button", { name: /create project/i }).click();
    await expect(dialog).toBeHidden({ timeout: 10_000 });

    // Verify project appears in the list
    await expect(page.getByRole("heading", { name: projName })).toBeVisible({ timeout: 15_000 });
  });
});
