import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";

test.describe("SCM Repository Draft & Publish Happy Path", () => {
  test("TC-REPO-PUBLISH-01 — repository draft 생성 및 publish 요청 성공 흐름", async ({ page }) => {
    // 1. system_admin 으로 OIDC 로그인
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto(appPath("/admin/catalog"));
    await expect(page).toHaveURL(/\/admin\/catalog(\/|\?|$)/, { timeout: 20_000 });

    // window.confirm 다이얼로그 승인 리스너 미리 설정 (레이스 방지)
    page.on("dialog", async (dialog) => {
      expect(dialog.message()).toContain("draft를 외부 SCM 연동 대상으로 전송");
      await dialog.accept();
    });

    // 2. Repositories 탭으로 전환
    await page.getByTestId("catalog-tab-repositories").click();
    await expect(page).toHaveURL(/tab=repositories/);

    // 3. New Repository 모달 트리거
    await page.getByRole("button", { name: /new repository/i }).click();
    const modal = page.locator("role=dialog");
    await expect(modal).toBeVisible();

    // 4. Repository draft 데이터 기입
    const randomSuffix = Math.random().toString(36).substring(2, 7).toUpperCase();
    const repoKey = `E2EREPO${randomSuffix}`;
    const repoSlug = `e2e-org/e2e-repo-${randomSuffix.toLowerCase()}`;

    await modal.locator("#repoKey").fill(repoKey);
    await modal.locator("#repoSlug").fill(repoSlug);
    // SCM Provider는 시드로 제공되는 'gitea' 선택
    await modal.locator("#repoProviderKey").selectOption("gitea");

    // 5. 생성 제출 및 모달 닫힘 검증
    await modal.getByRole("button", { name: /create repository/i }).click();
    await expect(modal).not.toBeVisible({ timeout: 10_000 });

    // 6. 테이블에 생성된 draft 노출 검증 (필터를 통해 검색)
    const search = page.getByPlaceholder("key/name/leader/status 검색");
    await search.fill(repoSlug);
    await expect(page).toHaveURL(new RegExp(`q=${encodeURIComponent(repoSlug)}`));

    // repoSlug 텍스트가 확실하게 담겨있는 tr을 타겟팅하여 병렬 레이스 컨디션을 예방
    const row = page.locator("tbody tr", { hasText: repoSlug }).first();
    await expect(row).toBeVisible({ timeout: 10_000 });
    await expect(row).toContainText("draft");

    // 8. Publish 버튼 클릭
    const publishBtn = row.getByRole("button", { name: /publish/i });
    await expect(publishBtn).toBeEnabled();
    await publishBtn.click();

    // 9. 성공 토스트 메시지 수신 및 상태 전이 검증
    const toast = page.locator("text=Publish 요청이 접수되었습니다");
    await expect(toast).toBeVisible({ timeout: 15_000 });
  });
});
