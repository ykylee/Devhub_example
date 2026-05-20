import { test } from "@playwright/test";
import { SEEDED, loginAs } from "./fixtures";

// screenshots.spec.ts — UI 가시화 자산 (sprint -h, issue #211, P0-3).
//
// 18 페이지 fullPage screenshot 캡처. CI artifact 로 업로드되어 Gemini
// (UI polish) 와 사용자 (디자인 review) 의 review source 가 된다.
//
// release_v1_roadmap.md §8 의 plan 따름. UI polish (P0-2, #210) sprint
// 진입 전 baseline 자산 확보.
//
// role 별 분담:
// - system_admin (charlie) — /admin/* 11 페이지 + dashboard admin
// - manager (bob) — /manager dashboard + (system_admin 미접근) carve out
// - developer (alice) — /developer dashboard + /applications/repositories/projects/dev-requests/account 7 페이지
//
// screenshot 파일 경로: `test-results/screenshots/<name>.png` (fullPage).
//
// 본 spec 은 visual regression 검증이 아니라 design review source 캡처가
// 목적이므로 expect/snapshot assertion 을 두지 않는다. 페이지 load + 일정
// 대기 후 캡처만.

const ADMIN_PAGES = [
  { path: "/admin", name: "01-admin-dashboard" },
  { path: "/admin/settings", name: "02-admin-settings-shell" },
  { path: "/admin/settings/users", name: "03-admin-users" },
  { path: "/admin/settings/organization", name: "04-admin-organization" },
  { path: "/admin/settings/permissions", name: "05-admin-permissions" },
  { path: "/admin/settings/audit", name: "06-admin-audit" },
  { path: "/admin/settings/dev-requests", name: "07-admin-dev-requests" },
  { path: "/admin/settings/dev-request-tokens", name: "08-admin-dev-request-tokens" },
  { path: "/admin/settings/integrations", name: "09-admin-integrations" },
  { path: "/admin/settings/integration-bindings", name: "10-admin-integration-bindings" },
  { path: "/admin/settings/applications", name: "11-admin-applications" },
  { path: "/admin/topology-v2", name: "12-admin-topology-v2" },
];

const USER_PAGES = [
  { path: "/applications", name: "13-user-applications" },
  { path: "/repositories", name: "14-user-repositories" },
  { path: "/projects", name: "15-user-projects" },
  { path: "/dev-requests", name: "16-user-dev-requests" },
  { path: "/account", name: "17-user-account" },
  { path: "/developer", name: "18-user-developer-dashboard" },
];

const CAPTURE_DIR = "test-results/screenshots";

// Quiet ahead-of-render delay so any client-side fetch + animation settles
// before fullPage capture. 1.5s is a heuristic — bump if pages keep being
// captured mid-skeleton.
const SETTLE_MS = 1500;

test.describe("UI screenshot capture (P0-3, design review source)", () => {
  test("admin pages (system_admin) — 12 captures", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    for (const { path, name } of ADMIN_PAGES) {
      await page.goto(path);
      await page.waitForLoadState("networkidle").catch(() => undefined);
      await page.waitForTimeout(SETTLE_MS);
      await page.screenshot({ path: `${CAPTURE_DIR}/${name}.png`, fullPage: true });
    }
  });

  test("login page — unauthenticated capture", async ({ page }) => {
    await page.goto("/login");
    await page.waitForLoadState("networkidle").catch(() => undefined);
    await page.waitForTimeout(SETTLE_MS);
    await page.screenshot({ path: `${CAPTURE_DIR}/00-login.png`, fullPage: true });
  });

  test("user pages (developer) — 6 captures", async ({ page }) => {
    await loginAs(page, SEEDED.developer);
    for (const { path, name } of USER_PAGES) {
      await page.goto(path);
      await page.waitForLoadState("networkidle").catch(() => undefined);
      await page.waitForTimeout(SETTLE_MS);
      await page.screenshot({ path: `${CAPTURE_DIR}/${name}.png`, fullPage: true });
    }
  });

  test("manager dashboard (manager) — 1 capture", async ({ page }) => {
    await loginAs(page, SEEDED.manager);
    await page.goto("/manager");
    await page.waitForLoadState("networkidle").catch(() => undefined);
    await page.waitForTimeout(SETTLE_MS);
    await page.screenshot({ path: `${CAPTURE_DIR}/19-manager-dashboard.png`, fullPage: true });
  });
});
