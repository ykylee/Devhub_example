import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";

const MODE = (process.env.DEVHUB_PROJECT_MODEL ?? "hybrid").trim().toLowerCase();

test.describe("project model route mode gating", () => {
  test("TC-PROJ-MODE-01 — legacy/v2 route availability follows DEVHUB_PROJECT_MODEL", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto(appPath("/admin/settings/platforms"));
    const currentPath = new URL(page.url()).pathname;
    const inferredBasePath = currentPath.startsWith("/devhub/") ? "/devhub" : "";
    const apiBasePath = appPath("/").replace(/\/$/, "") || inferredBasePath;

    const result = await page.evaluate(async ({ mode, apiBasePath }) => {
      const token = sessionStorage.getItem("devhub_access_token");
      if (!token) throw new Error("missing access token");
      const headers = { Authorization: `Bearer ${token}` };

      const legacy = await fetch(`${apiBasePath}/api/v1/repositories/1/projects`, { headers });
      const v2 = await fetch(`${apiBasePath}/api/v1/platforms/00000000-0000-0000-0000-000000000000/projects`, { headers });

      return {
        mode,
        legacyStatus: legacy.status,
        v2Status: v2.status,
      };
    }, { mode: MODE, apiBasePath });

    if (MODE === "legacy") {
      expect(result.legacyStatus).not.toBe(410);
      expect(result.v2Status).toBe(410);
      return;
    }
    if (MODE === "v2") {
      expect(result.legacyStatus).toBe(410);
      expect(result.v2Status).not.toBe(410);
      return;
    }
    // hybrid(default)
    expect(result.legacyStatus).not.toBe(410);
    expect(result.v2Status).not.toBe(410);
  });
});
