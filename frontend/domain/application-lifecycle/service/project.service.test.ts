import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("projectService", () => {
  const apiClientMock = vi.fn();

  // ApiError mock — production ApiError 와 동일한 shape (status / payload).
  class MockApiError extends Error {
    status: number;
    payload: unknown;
    constructor(status: number, payload: unknown, message: string) {
      super(message);
      this.name = "ApiError";
      this.status = status;
      this.payload = payload;
    }
  }

  beforeEach(() => {
    vi.resetModules();
    apiClientMock.mockReset();
    vi.doMock("@/lib/services/api-client", () => ({
      apiClient: apiClientMock,
      ApiError: MockApiError,
    }));
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("withQuery / getApplications", () => {
    it("serializes query params for getApplications", async () => {
      apiClientMock.mockResolvedValue({ data: [] });
      const { projectService } = await import("./project.service");

      await projectService.getApplications({ status: "active", include_archived: true, q: "devhub" });

      expect(apiClientMock).toHaveBeenCalledWith(
        "GET",
        "/api/v1/applications?status=active&include_archived=true&q=devhub",
      );
    });

    it("returns plain path when no params (undefined)", async () => {
      apiClientMock.mockResolvedValue({ data: [] });
      const { projectService } = await import("./project.service");

      await projectService.getApplications();

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/applications");
    });

    it("skips undefined fields in query but keeps defined ones", async () => {
      apiClientMock.mockResolvedValue({ data: [] });
      const { projectService } = await import("./project.service");

      await projectService.getApplications({ status: "active", include_archived: undefined });

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/applications?status=active");
    });

    it("returns path without ?qs when all params are undefined (raw empty)", async () => {
      apiClientMock.mockResolvedValue({ data: [] });
      const { projectService } = await import("./project.service");

      await projectService.getApplications({ status: undefined, include_archived: undefined });

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/applications");
    });
  });

  describe("getSCMProviders", () => {
    it("GETs /api/v1/scm/providers and unwraps data", async () => {
      apiClientMock.mockResolvedValue({ data: [{ provider_key: "gitea" }] });
      const { projectService } = await import("./project.service");

      const out = await projectService.getSCMProviders();

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/scm/providers");
      expect(out).toEqual([{ provider_key: "gitea" }]);
    });
  });

  describe("Application CRUD", () => {
    it("createApplication POSTs body and returns data", async () => {
      apiClientMock.mockResolvedValue({ data: { id: "a1", name: "App" } });
      const { projectService } = await import("./project.service");

      const out = await projectService.createApplication({ name: "App" });

      expect(apiClientMock).toHaveBeenCalledWith("POST", "/api/v1/applications", { name: "App" });
      expect(out).toEqual({ id: "a1", name: "App" });
    });

    it("getApplication GETs /api/v1/applications/:id", async () => {
      apiClientMock.mockResolvedValue({ data: { id: "a1" } });
      const { projectService } = await import("./project.service");

      const out = await projectService.getApplication("a1");

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/applications/a1");
      expect(out).toEqual({ id: "a1" });
    });

    it("updateApplication PATCHes /api/v1/applications/:id with body", async () => {
      apiClientMock.mockResolvedValue({ data: { id: "a1", status: "on_hold" } });
      const { projectService } = await import("./project.service");

      const out = await projectService.updateApplication("a1", { status: "on_hold", hold_reason: "freeze" });

      expect(apiClientMock).toHaveBeenCalledWith(
        "PATCH",
        "/api/v1/applications/a1",
        { status: "on_hold", hold_reason: "freeze" },
      );
      expect(out).toEqual({ id: "a1", status: "on_hold" });
    });

    it("archiveApplication DELETEs plain path when hard is falsy", async () => {
      apiClientMock.mockResolvedValue(undefined);
      const { projectService } = await import("./project.service");

      await projectService.archiveApplication("a1");
      await projectService.archiveApplication("a2", false);

      expect(apiClientMock).toHaveBeenNthCalledWith(1, "DELETE", "/api/v1/applications/a1");
      expect(apiClientMock).toHaveBeenNthCalledWith(2, "DELETE", "/api/v1/applications/a2");
    });

    it("archiveApplication DELETEs with ?hard=true when hard is true", async () => {
      apiClientMock.mockResolvedValue(undefined);
      const { projectService } = await import("./project.service");

      await projectService.archiveApplication("a1", true);

      expect(apiClientMock).toHaveBeenCalledWith("DELETE", "/api/v1/applications/a1?hard=true");
    });
  });

  describe("Repository linking", () => {
    it("getApplicationRepositories GETs /api/v1/applications/:id/repositories", async () => {
      apiClientMock.mockResolvedValue({ data: [{ application_id: "a1" }] });
      const { projectService } = await import("./project.service");

      const out = await projectService.getApplicationRepositories("a1");

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/applications/a1/repositories");
      expect(out).toHaveLength(1);
    });

    it("connectRepository POSTs role+repo data", async () => {
      apiClientMock.mockResolvedValue({ data: { application_id: "a1" } });
      const { projectService } = await import("./project.service");

      await projectService.connectRepository("a1", {
        repo_provider: "gitea",
        repo_full_name: "team/repo",
        role: "primary",
      });

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        "/api/v1/applications/a1/repositories",
        { repo_provider: "gitea", repo_full_name: "team/repo", role: "primary" },
      );
    });

    it("disconnectRepository URL-encodes the provider/repo composite key", async () => {
      apiClientMock.mockResolvedValue(undefined);
      const { projectService } = await import("./project.service");

      await projectService.disconnectRepository("a1", "gitea", "team/repo");

      expect(apiClientMock).toHaveBeenCalledWith(
        "DELETE",
        "/api/v1/applications/a1/repositories/gitea%2Fteam%2Frepo",
      );
    });
  });

  describe("getRepositoryProjects / getApplicationProjects", () => {
    it("getRepositoryProjects builds query when params present", async () => {
      apiClientMock.mockResolvedValue({ data: [] });
      const { projectService } = await import("./project.service");

      await projectService.getRepositoryProjects(42, { status: "active", include_archived: true });

      expect(apiClientMock).toHaveBeenCalledWith(
        "GET",
        "/api/v1/repositories/42/projects?status=active&include_archived=true",
      );
    });

    it("getRepositoryProjects omits qs when no params", async () => {
      apiClientMock.mockResolvedValue({ data: [] });
      const { projectService } = await import("./project.service");

      await projectService.getRepositoryProjects(42);

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/repositories/42/projects");
    });

    it("fetches application projects by repository ids and flattens result", async () => {
      const { projectService } = await import("./project.service");
      apiClientMock
        .mockResolvedValueOnce({
          data: [
            { repository_id: 101, repo_provider: "gitea", repo_full_name: "a/b" },
            { repository_id: 202, repo_provider: "gitea", repo_full_name: "a/c" },
            { repo_provider: "gitea", repo_full_name: "a/no-id" },
          ],
        })
        .mockResolvedValueOnce({ data: [{ id: "p1", name: "A" }] })
        .mockResolvedValueOnce({ data: [{ id: "p2", name: "B" }] });

      const projects = await projectService.getApplicationProjects("app-1");

      expect(projects).toHaveLength(2);
      expect(apiClientMock).toHaveBeenNthCalledWith(1, "GET", "/api/v1/applications/app-1/repositories");
      expect(apiClientMock).toHaveBeenNthCalledWith(2, "GET", "/api/v1/repositories/101/projects");
      expect(apiClientMock).toHaveBeenNthCalledWith(3, "GET", "/api/v1/repositories/202/projects");
    });

    it("swallows per-repository project lookup errors", async () => {
      const { projectService } = await import("./project.service");
      apiClientMock
        .mockResolvedValueOnce({
          data: [
            { repository_id: 10, repo_provider: "gitea", repo_full_name: "x/a" },
            { repository_id: 20, repo_provider: "gitea", repo_full_name: "x/b" },
          ],
        })
        .mockRejectedValueOnce(new Error("repo 10 fail"))
        .mockResolvedValueOnce({ data: [{ id: "p20", name: "from-20" }] });

      const projects = await projectService.getApplicationProjects("app-x");

      expect(projects).toEqual([{ id: "p20", name: "from-20" }]);
    });

    it("filters out NaN / non-finite repository_ids", async () => {
      const { projectService } = await import("./project.service");
      apiClientMock.mockResolvedValueOnce({
        data: [
          { repository_id: 7, repo_provider: "gitea", repo_full_name: "x/y" },
          { repository_id: NaN, repo_provider: "gitea", repo_full_name: "bad/nan" },
          { repository_id: "10", repo_provider: "gitea", repo_full_name: "string/id" },
        ],
      });
      apiClientMock.mockResolvedValueOnce({ data: [{ id: "p7" }] });

      const projects = await projectService.getApplicationProjects("app-1");

      expect(projects).toEqual([{ id: "p7" }]);
      // 2 calls only: 1 repos + 1 valid repo (7) — NaN filtered, string excluded.
      expect(apiClientMock).toHaveBeenCalledTimes(2);
    });
  });

  describe("Project create flow", () => {
    it("createProject POSTs under repository", async () => {
      apiClientMock.mockResolvedValue({ data: { id: "p1" } });
      const { projectService } = await import("./project.service");

      await projectService.createProject(42, { name: "P", key: "PROJ" });

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        "/api/v1/repositories/42/projects",
        { name: "P", key: "PROJ" },
      );
    });

    it("createProjectStandalone POSTs to /api/v1/projects", async () => {
      apiClientMock.mockResolvedValue({ data: { id: "p1" } });
      const { projectService } = await import("./project.service");

      await projectService.createProjectStandalone({
        name: "S",
        key: "STD",
        repository_ids: [1, 2],
        repository_create_payload: { key: "REPO", slug: "repo", scm_provider: "gitea" },
      });

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        "/api/v1/projects",
        expect.objectContaining({ name: "S", repository_ids: [1, 2] }),
      );
    });

    it("getApplicationProjectsV2 with params", async () => {
      apiClientMock.mockResolvedValue({ data: [] });
      const { projectService } = await import("./project.service");

      await projectService.getApplicationProjectsV2("a1", { status: "active" });

      expect(apiClientMock).toHaveBeenCalledWith(
        "GET",
        "/api/v1/applications/a1/projects?status=active",
      );
    });

    it("getApplicationProjectsV2 without params", async () => {
      apiClientMock.mockResolvedValue({ data: [] });
      const { projectService } = await import("./project.service");

      await projectService.getApplicationProjectsV2("a1");

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/applications/a1/projects");
    });

    it("listStandaloneProjects builds query when present", async () => {
      apiClientMock.mockResolvedValue({ data: [] });
      const { projectService } = await import("./project.service");

      await projectService.listStandaloneProjects({ status: "active", include_archived: false });

      expect(apiClientMock).toHaveBeenCalledWith(
        "GET",
        "/api/v1/projects/standalone?status=active&include_archived=false",
      );
    });

    it("listStandaloneProjects without params", async () => {
      apiClientMock.mockResolvedValue({ data: [] });
      const { projectService } = await import("./project.service");

      await projectService.listStandaloneProjects();

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/projects/standalone");
    });
  });

  describe("createApplicationProject — ApiError fallback (404/405)", () => {
    it("happy path POSTs to /api/v1/applications/:id/projects", async () => {
      apiClientMock.mockResolvedValue({ data: { id: "p1" } });
      const { projectService } = await import("./project.service");

      const out = await projectService.createApplicationProject("a1", { name: "P" });

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        "/api/v1/applications/a1/projects",
        { name: "P" },
      );
      expect(out).toEqual({ id: "p1" });
    });

    it("falls back to createProject when ApiError 404 + repository_ids present", async () => {
      const { projectService } = await import("./project.service");
      apiClientMock
        .mockRejectedValueOnce(new MockApiError(404, null, "not found"))
        .mockResolvedValueOnce({ data: { id: "fallback" } });

      const out = await projectService.createApplicationProject("a1", {
        name: "P",
        repository_ids: [11],
      });

      expect(apiClientMock).toHaveBeenNthCalledWith(
        2,
        "POST",
        "/api/v1/repositories/11/projects",
        expect.objectContaining({ name: "P", application_id: "a1", repository_ids: [11] }),
      );
      expect(out).toEqual({ id: "fallback" });
    });

    it("falls back to createProject when ApiError 405 + repository_id (singular)", async () => {
      const { projectService } = await import("./project.service");
      apiClientMock
        .mockRejectedValueOnce(new MockApiError(405, null, "method not allowed"))
        .mockResolvedValueOnce({ data: { id: "via-repo-id" } });

      const out = await projectService.createApplicationProject("a1", {
        name: "P",
        repository_id: 99,
      } as Partial<{ repository_id: number; name: string }>);

      expect(apiClientMock).toHaveBeenNthCalledWith(
        2,
        "POST",
        "/api/v1/repositories/99/projects",
        expect.objectContaining({ application_id: "a1" }),
      );
      expect(out).toEqual({ id: "via-repo-id" });
    });

    it("re-throws ApiError 404 when no repository hint available", async () => {
      const { projectService } = await import("./project.service");
      const err = new MockApiError(404, null, "not found");
      apiClientMock.mockRejectedValueOnce(err);

      await expect(
        projectService.createApplicationProject("a1", { name: "P" }),
      ).rejects.toBe(err);
    });

    it("re-throws ApiError with other status (e.g., 500) — does not fallback", async () => {
      const { projectService } = await import("./project.service");
      const err = new MockApiError(500, null, "server error");
      apiClientMock.mockRejectedValueOnce(err);

      await expect(
        projectService.createApplicationProject("a1", { name: "P", repository_ids: [1] }),
      ).rejects.toBe(err);
      expect(apiClientMock).toHaveBeenCalledTimes(1); // no fallback
    });

    it("re-throws non-ApiError errors verbatim", async () => {
      const { projectService } = await import("./project.service");
      const err = new Error("network down");
      apiClientMock.mockRejectedValueOnce(err);

      await expect(
        projectService.createApplicationProject("a1", { name: "P", repository_ids: [1] }),
      ).rejects.toBe(err);
    });
  });

  describe("Project read endpoints", () => {
    it("getProjectRepositories", async () => {
      apiClientMock.mockResolvedValue({ data: [{ project_id: "p1", repository_id: 1, role: "primary" }] });
      const { projectService } = await import("./project.service");

      const out = await projectService.getProjectRepositories("p1");

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/projects/p1/repositories");
      expect(out).toHaveLength(1);
    });

    it("getProject", async () => {
      apiClientMock.mockResolvedValue({ data: { id: "p1" } });
      const { projectService } = await import("./project.service");

      const out = await projectService.getProject("p1");

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/projects/p1");
      expect(out).toEqual({ id: "p1" });
    });

    it("updateProject", async () => {
      apiClientMock.mockResolvedValue({ data: { id: "p1", name: "renamed" } });
      const { projectService } = await import("./project.service");

      await projectService.updateProject("p1", { name: "renamed" });

      expect(apiClientMock).toHaveBeenCalledWith("PATCH", "/api/v1/projects/p1", { name: "renamed" });
    });

    it("archiveProject without hard", async () => {
      apiClientMock.mockResolvedValue(undefined);
      const { projectService } = await import("./project.service");

      await projectService.archiveProject("p1");

      expect(apiClientMock).toHaveBeenCalledWith("DELETE", "/api/v1/projects/p1");
    });

    it("archiveProject with hard=true", async () => {
      apiClientMock.mockResolvedValue(undefined);
      const { projectService } = await import("./project.service");

      await projectService.archiveProject("p1", true);

      expect(apiClientMock).toHaveBeenCalledWith("DELETE", "/api/v1/projects/p1?hard=true");
    });
  });

  describe("getProjectActivity — asList + item normalization", () => {
    it("returns [] when response is not array and lacks data", async () => {
      apiClientMock.mockResolvedValue({ unrelated: "field" });
      const { projectService } = await import("./project.service");

      const out = await projectService.getProjectActivity("p1");

      expect(out).toEqual([]);
    });

    it("returns [] when response is null", async () => {
      apiClientMock.mockResolvedValue(null);
      const { projectService } = await import("./project.service");

      const out = await projectService.getProjectActivity("p1");

      expect(out).toEqual([]);
    });

    it("returns [] when response is a string (primitive)", async () => {
      apiClientMock.mockResolvedValue("string-payload");
      const { projectService } = await import("./project.service");

      const out = await projectService.getProjectActivity("p1");

      expect(out).toEqual([]);
    });

    it("returns [] when data wrapper has a non-array data field", async () => {
      apiClientMock.mockResolvedValue({ data: { not: "array" } });
      const { projectService } = await import("./project.service");

      const out = await projectService.getProjectActivity("p1");

      expect(out).toEqual([]);
    });

    it("normalizes raw array with falsy fields → fallback strings", async () => {
      apiClientMock.mockResolvedValue([{}]);
      const { projectService } = await import("./project.service");

      const out = await projectService.getProjectActivity("p1");

      expect(out).toHaveLength(1);
      expect(out[0].id).toBe("activity-0");
      expect(out[0].user).toBe("-");
      expect(out[0].action).toBe("updated");
      expect(out[0].target).toBe("-");
      // occurred_at fallback = new Date().toISOString() — must look like ISO 8601.
      expect(out[0].occurred_at).toMatch(/^\d{4}-\d{2}-\d{2}T/);
    });

    it("respects activity_id / actor / target_name / created_at fallback aliases", async () => {
      apiClientMock.mockResolvedValue([
        {
          activity_id: "act-1",
          actor: "alice",
          action: "approved",
          target_name: "dreq-7",
          created_at: "2026-05-28T11:00:00Z",
        },
      ]);
      const { projectService } = await import("./project.service");

      const out = await projectService.getProjectActivity("p1");

      expect(out[0]).toEqual({
        id: "act-1",
        user: "alice",
        action: "approved",
        target: "dreq-7",
        occurred_at: "2026-05-28T11:00:00Z",
      });
    });

    it("prefers id over activity_id, user over actor, target over target_name", async () => {
      apiClientMock.mockResolvedValue([
        {
          id: "id-1",
          activity_id: "alt-1",
          user: "u1",
          actor: "alt-actor",
          action: "merged",
          target: "tgt-1",
          target_name: "alt-tgt",
          occurred_at: "2026-05-28T12:00:00Z",
        },
      ]);
      const { projectService } = await import("./project.service");

      const out = await projectService.getProjectActivity("p1");

      expect(out[0].id).toBe("id-1");
      expect(out[0].user).toBe("u1");
      expect(out[0].target).toBe("tgt-1");
    });

    it("falls back user → user_name when actor missing", async () => {
      apiClientMock.mockResolvedValue([{ user_name: "u-name", action: "x" }]);
      const { projectService } = await import("./project.service");

      const out = await projectService.getProjectActivity("p1");

      expect(out[0].user).toBe("u-name");
    });

    it("unwraps {data: [...]} envelope", async () => {
      apiClientMock.mockResolvedValue({ data: [{ id: "a-1", user: "u1", action: "ok", target: "x", occurred_at: "t" }] });
      const { projectService } = await import("./project.service");

      const out = await projectService.getProjectActivity("p1");

      expect(out).toHaveLength(1);
      expect(out[0].id).toBe("a-1");
    });
  });

  describe("getProjectTasks — normalize + status filter", () => {
    it("joins default status array into query", async () => {
      apiClientMock.mockResolvedValue({ data: [] });
      const { projectService } = await import("./project.service");

      await projectService.getProjectTasks("p1");

      const [, path] = apiClientMock.mock.calls[0];
      const url = new URL(path as string, "http://example");
      expect(url.searchParams.get("status")).toBe("todo,in_progress,review");
    });

    it("respects custom statuses array", async () => {
      apiClientMock.mockResolvedValue({ data: [] });
      const { projectService } = await import("./project.service");

      await projectService.getProjectTasks("p1", ["done"]);

      const [, path] = apiClientMock.mock.calls[0];
      const url = new URL(path as string, "http://example");
      expect(url.searchParams.get("status")).toBe("done");
    });

    it("normalizes priority — accepted values pass through", async () => {
      apiClientMock.mockResolvedValue([
        { id: "t1", title: "low task", priority: "LOW", status: "todo" },
        { id: "t2", title: "medium task", priority: "Medium", status: "in_progress" },
        { id: "t3", title: "high task", priority: "high", status: "review" },
        { id: "t4", title: "crit task", priority: "CRITICAL", status: "done" },
      ]);
      const { projectService } = await import("./project.service");

      const out = await projectService.getProjectTasks("p1");

      expect(out[0].priority).toBe("low");
      expect(out[1].priority).toBe("medium");
      expect(out[2].priority).toBe("high");
      expect(out[3].priority).toBe("critical");
    });

    it("normalizePriority returns 'medium' for unknown / null / undefined", async () => {
      apiClientMock.mockResolvedValue([
        { id: "t1", title: "weird", priority: "bogus", status: "todo" },
        { id: "t2", title: "null-prio", priority: null, status: "todo" },
        { id: "t3", title: "undef-prio", status: "todo" }, // priority undefined
        { id: "t4", title: "num-prio", priority: 42, status: "todo" },
      ]);
      const { projectService } = await import("./project.service");

      const out = await projectService.getProjectTasks("p1");

      out.forEach((t) => expect(t.priority).toBe("medium"));
    });

    it("normalizes status — known values + fallback to 'todo'", async () => {
      apiClientMock.mockResolvedValue([
        { id: "t1", title: "x", priority: "low", status: "DONE" },
        { id: "t2", title: "y", priority: "low", status: "Review" },
        { id: "t3", title: "z", priority: "low", status: "IN_PROGRESS" },
        { id: "t4", title: "w", priority: "low", status: "unknown" },
        { id: "t5", title: "v", priority: "low", status: null },
        { id: "t6", title: "u", priority: "low" }, // status missing
      ]);
      const { projectService } = await import("./project.service");

      const out = await projectService.getProjectTasks("p1");

      expect(out[0].status).toBe("done");
      expect(out[1].status).toBe("review");
      expect(out[2].status).toBe("in_progress");
      expect(out[3].status).toBe("todo");
      expect(out[4].status).toBe("todo");
      expect(out[5].status).toBe("todo");
    });

    it("preserves due_date string but drops non-string", async () => {
      apiClientMock.mockResolvedValue([
        { id: "t1", title: "x", priority: "low", status: "todo", due_date: "2026-05-30" },
        { id: "t2", title: "y", priority: "low", status: "todo", due_date: 12345 },
      ]);
      const { projectService } = await import("./project.service");

      const out = await projectService.getProjectTasks("p1");

      expect(out[0].due_date).toBe("2026-05-30");
      expect(out[1].due_date).toBeUndefined();
    });

    it("preserves comment_count / attachment_count when number; otherwise undefined", async () => {
      apiClientMock.mockResolvedValue([
        { id: "t1", title: "x", priority: "low", status: "todo", comment_count: 3, attachment_count: 0 },
        { id: "t2", title: "y", priority: "low", status: "todo", comment_count: "5", attachment_count: null },
      ]);
      const { projectService } = await import("./project.service");

      const out = await projectService.getProjectTasks("p1");

      expect(out[0].comment_count).toBe(3);
      expect(out[0].attachment_count).toBe(0);
      expect(out[1].comment_count).toBeUndefined();
      expect(out[1].attachment_count).toBeUndefined();
    });

    it("normalize id/title fallbacks: task_id / name / Untitled / index", async () => {
      apiClientMock.mockResolvedValue([
        { task_id: "tid-1", name: "by-name", priority: "low", status: "todo" },
        { priority: "low", status: "todo" }, // no id, no title
      ]);
      const { projectService } = await import("./project.service");

      const out = await projectService.getProjectTasks("p1");

      expect(out[0].id).toBe("tid-1");
      expect(out[0].title).toBe("by-name");
      expect(out[1].id).toBe("task-1");
      expect(out[1].title).toBe("Untitled Task");
    });

    it("unwraps {data: [...]} envelope for tasks", async () => {
      apiClientMock.mockResolvedValue({ data: [{ id: "t1", title: "from-data", priority: "low", status: "todo" }] });
      const { projectService } = await import("./project.service");

      const out = await projectService.getProjectTasks("p1");

      expect(out).toHaveLength(1);
      expect(out[0].title).toBe("from-data");
    });
  });

  describe("linkProjectRepository / unlinkProjectRepository", () => {
    it("link with default role 'linked'", async () => {
      apiClientMock.mockResolvedValue({ data: { project_id: "p1", repository_id: 5, role: "linked" } });
      const { projectService } = await import("./project.service");

      await projectService.linkProjectRepository("p1", 5);

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        "/api/v1/projects/p1/repositories",
        { repository_id: 5, role: "linked" },
      );
    });

    it("link with role 'primary'", async () => {
      apiClientMock.mockResolvedValue({ data: {} });
      const { projectService } = await import("./project.service");

      await projectService.linkProjectRepository("p1", 5, "primary");

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        "/api/v1/projects/p1/repositories",
        { repository_id: 5, role: "primary" },
      );
    });

    it("link with role 'shared'", async () => {
      apiClientMock.mockResolvedValue({ data: {} });
      const { projectService } = await import("./project.service");

      await projectService.linkProjectRepository("p1", 5, "shared");

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        "/api/v1/projects/p1/repositories",
        { repository_id: 5, role: "shared" },
      );
    });

    it("unlinkProjectRepository DELETEs nested path", async () => {
      apiClientMock.mockResolvedValue(undefined);
      const { projectService } = await import("./project.service");

      await projectService.unlinkProjectRepository("p1", 5);

      expect(apiClientMock).toHaveBeenCalledWith("DELETE", "/api/v1/projects/p1/repositories/5");
    });
  });

  describe("listAllProjects — concatenates + swallows per-repo errors", () => {
    it("aggregates projects across repository ids", async () => {
      const { projectService } = await import("./project.service");
      apiClientMock
        .mockResolvedValueOnce({ data: [{ id: "p-a" }] })
        .mockResolvedValueOnce({ data: [{ id: "p-b" }, { id: "p-c" }] });

      const out = await projectService.listAllProjects([1, 2]);

      expect(out).toHaveLength(3);
      expect(out.map((p) => p.id)).toEqual(["p-a", "p-b", "p-c"]);
    });

    it("logs and swallows error for a single repo and continues to next", async () => {
      const errSpy = vi.spyOn(console, "error").mockImplementation(() => undefined);
      const { projectService } = await import("./project.service");
      apiClientMock
        .mockRejectedValueOnce(new Error("repo-1 fail"))
        .mockResolvedValueOnce({ data: [{ id: "p-2" }] });

      const out = await projectService.listAllProjects([1, 2]);

      expect(out).toEqual([{ id: "p-2" }]);
      expect(errSpy).toHaveBeenCalled();
    });

    it("forwards params to underlying getRepositoryProjects", async () => {
      const { projectService } = await import("./project.service");
      apiClientMock.mockResolvedValueOnce({ data: [] });

      await projectService.listAllProjects([7], { status: "closed" });

      expect(apiClientMock).toHaveBeenCalledWith(
        "GET",
        "/api/v1/repositories/7/projects?status=closed",
      );
    });

    it("returns [] when no repository ids", async () => {
      const { projectService } = await import("./project.service");
      const out = await projectService.listAllProjects([]);

      expect(out).toEqual([]);
      expect(apiClientMock).not.toHaveBeenCalled();
    });
  });
});
