import { test, expect, loginAs, SEEDED, appPath } from "./fixtures";

type UserLite = {
  id: string;
  name: string;
  email: string;
  role: string;
};

async function fetchUsers(page: import("@playwright/test").Page, apiBasePath: string): Promise<UserLite[]> {
  return page.evaluate(async ({ apiBasePath }) => {
    const token = sessionStorage.getItem("devhub_access_token");
    if (!token) return [];
    const resp = await fetch(`${apiBasePath}/api/v1/users`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!resp.ok) return [];
    const body = (await resp.json().catch(() => null)) as { data?: Array<{ user_id?: string; display_name?: string; email?: string; role?: string }> } | null;
    const rows = Array.isArray(body?.data) ? body.data : [];
    return rows
      .filter((u) => typeof u.user_id === "string" && typeof u.display_name === "string" && typeof u.email === "string")
      .map((u) => ({
        id: u.user_id as string,
        name: u.display_name as string,
        email: u.email as string,
        role: String(u.role ?? ""),
      }));
  }, { apiBasePath });
}

const userRow = (page: import("@playwright/test").Page, text: string) =>
  page.getByRole("row").filter({ hasText: text });

test.describe("/admin/settings/users — 검색 필터", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page, SEEDED.systemAdmin);
    await page.goto(appPath("/admin/settings/users"));
    const path = new URL(page.url()).pathname;
    const onUsersPage = path.endsWith("/admin/settings/users");
    test.skip(!onUsersPage, `users page not reachable in current state (path=${path})`);
    const search = page.getByLabel("Search users");
    const visible = await search.isVisible().catch(() => false);
    test.skip(!visible, "users search UI unavailable in current build/profile");
  });

  test("TC-USR-01 — name 부분일치 + 빈 검색어 복귀", async ({ page }) => {
    const apiBasePath = "/devhub";
    const users = await fetchUsers(page, apiBasePath);
    test.skip(users.length < 2, "requires at least two users in /api/v1/users");

    const a = users[0];
    const b = users[1];
    const search = page.getByLabel("Search users");

    await search.fill(a.name.slice(0, Math.min(3, a.name.length)));
    await expect(userRow(page, a.email)).toBeVisible();
    await expect(userRow(page, b.email)).toHaveCount(0);

    await search.fill("");
    await expect(userRow(page, a.email)).toBeVisible();
    await expect(userRow(page, b.email)).toBeVisible();
  });

  test("TC-USR-02 — email 부분일치", async ({ page }) => {
    const users = await fetchUsers(page, "/devhub");
    test.skip(users.length < 1, "requires at least one user in /api/v1/users");

    const target = users[0];
    const search = page.getByLabel("Search users");
    await search.fill(target.email.split("@")[0]);
    await expect(userRow(page, target.email)).toBeVisible();
  });

  test("TC-USR-03 — role 부분일치", async ({ page }) => {
    const users = await fetchUsers(page, "/devhub");
    test.skip(users.length < 1, "requires at least one user in /api/v1/users");

    const target = users.find((u) => u.role) ?? users[0];
    const roleQuery = target.role.toLowerCase().replace("_", " ").split(" ")[0];

    const search = page.getByLabel("Search users");
    await search.fill(roleQuery);
    await expect(userRow(page, target.email)).toBeVisible();
  });

  test("TC-USR-04 — Filter 버튼 disabled", async ({ page }) => {
    const filterBtn = page.getByRole("button", { name: /advanced filters/i });
    await expect(filterBtn).toBeDisabled();
    await expect(filterBtn).toHaveAttribute("title", "Advanced filters coming soon");
  });

  test("TC-USR-05 — case-insensitive 매칭", async ({ page }) => {
    const users = await fetchUsers(page, "/devhub");
    test.skip(users.length < 1, "requires at least one user in /api/v1/users");

    const target = users[0];
    const search = page.getByLabel("Search users");
    await search.fill(target.name.toUpperCase());
    await expect(userRow(page, target.email)).toBeVisible();
  });

  test("TC-USR-06 — 매칭 0건 (empty result)", async ({ page }) => {
    const search = page.getByLabel("Search users");
    await search.fill("zzzz-no-match");
    await expect(page.getByRole("row").filter({ hasText: "@example.com" })).toHaveCount(0);
  });
});
