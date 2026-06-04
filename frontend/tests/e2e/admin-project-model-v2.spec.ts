import { appPath, expect, loginAs, SEEDED, test } from "./fixtures";

test.describe("project model v2/hybrid API flow", () => {
  test("TC-PROJ-V2-01 — platform -> project -> project_repositories(N:M)", async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto(appPath("/admin/settings/platforms"));

    const unique = Date.now().toString().slice(-8);
    const appKey = `A${unique}Z`; // 10 chars
    const currentPath = new URL(page.url()).pathname;
    const inferredBasePath = currentPath.startsWith("/devhub/") ? "/devhub" : "";
    const apiBasePath = appPath("/").replace(/\/$/, "") || inferredBasePath;
    const repoCount = await page.evaluate(async ({ apiBasePath }) => {
      const token = sessionStorage.getItem("devhub_access_token");
      if (!token) return -1;
      const resp = await fetch(`${apiBasePath}/api/v1/repositories`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!resp.ok) return -1;
      const body = (await resp.json().catch(() => null)) as { data?: unknown[] } | null;
      return Array.isArray(body?.data) ? body.data.length : -1;
    }, { apiBasePath });
    test.skip(repoCount === 0, "requires at least one repository fixture (sync or seed) in environment");

    const result = await page.evaluate(async ({ appKey, unique, apiBasePath }) => {
      const token = sessionStorage.getItem("devhub_access_token");
      if (!token) throw new Error("missing access token");
      const readBody = async (resp: Response) => {
        const raw = await resp.text();
        try {
          return JSON.parse(raw);
        } catch {
          return { raw };
        }
      };

      const headers = {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      };

      const providersResp = await fetch(`${apiBasePath}/api/v1/scm/providers`, { method: "GET", headers });
      const providersBody = await readBody(providersResp);
      if (!providersResp.ok) throw new Error(`list scm providers failed: ${providersResp.status} ${JSON.stringify(providersBody)}`);
      const enabledProvider = Array.isArray(providersBody?.data)
        ? providersBody.data.find((p: { provider_key?: string; enabled?: boolean }) => p?.enabled && p?.provider_key)
        : null;
      if (!enabledProvider?.provider_key) throw new Error("no enabled scm provider available for test");

      const repoListResp = await fetch(`${apiBasePath}/api/v1/repositories`, { method: "GET", headers });
      const repoListBody = await readBody(repoListResp);
      if (!repoListResp.ok) throw new Error(`list repositories failed: ${repoListResp.status} ${JSON.stringify(repoListBody)}`);
      const firstRepo = Array.isArray(repoListBody?.data)
        ? repoListBody.data.find((r: { id?: number; full_name?: string }) => typeof r?.id === "number" && typeof r?.full_name === "string")
        : null;
      if (!firstRepo?.id || !firstRepo?.full_name) throw new Error("no repository(id/full_name) available for test");

      const appResp = await fetch(`${apiBasePath}/api/v1/platforms`, {
        method: "POST",
        headers,
        body: JSON.stringify({
          key: appKey,
          name: `E2E Proj V2 ${unique}`,
          description: "e2e v2 flow",
          owner_user_id: "charlie",
          leader_user_id: "charlie",
          development_unit_id: "dept-eng",
          visibility: "internal",
          status: "planning",
        }),
      });
      const appBody = await readBody(appResp);
      if (!appResp.ok) throw new Error(`create platform failed: ${appResp.status} ${JSON.stringify(appBody)}`);
      const appID = appBody?.data?.id;

      const linkResp = await fetch(`${apiBasePath}/api/v1/platforms/${appID}/repositories`, {
        method: "POST",
        headers,
        body: JSON.stringify({
          repo_provider: enabledProvider.provider_key,
          repo_full_name: firstRepo.full_name,
          role: "primary",
        }),
      });
      const linkBody = await readBody(linkResp);
      if (!linkResp.ok) throw new Error(`link app repository failed: ${linkResp.status} ${JSON.stringify(linkBody)}`);

      const projectResp = await fetch(`${apiBasePath}/api/v1/platforms/${appID}/projects`, {
        method: "POST",
        headers,
        body: JSON.stringify({
          key: `PROJ-${unique.slice(-4)}`,
          name: `E2E Project ${unique}`,
          description: "project via v2 endpoint",
          owner_user_id: "charlie",
          visibility: "internal",
          status: "planning",
          repository_id: firstRepo.id,
          repository_ids: [firstRepo.id],
        }),
      });
      const projectBody = await readBody(projectResp);
      if (!projectResp.ok) throw new Error(`create project failed: ${projectResp.status} ${JSON.stringify(projectBody)}`);
      const projectID = projectBody?.data?.id;

      const linksResp = await fetch(`${apiBasePath}/api/v1/projects/${projectID}/repositories`, { method: "GET", headers });
      const linksBody = await readBody(linksResp);
      if (!linksResp.ok) throw new Error(`list project repos failed: ${linksResp.status} ${JSON.stringify(linksBody)}`);

      await fetch(`${apiBasePath}/api/v1/platforms/${appID}`, { method: "DELETE", headers }).catch(() => undefined);

      return {
        appID,
        projectID,
        linkCount: Array.isArray(linksBody?.data) ? linksBody.data.length : 0,
      };
    }, { appKey, unique, apiBasePath });

    expect(result.appID).toBeTruthy();
    expect(result.projectID).toBeTruthy();
    expect(result.linkCount).toBeGreaterThan(0);
  });
});
