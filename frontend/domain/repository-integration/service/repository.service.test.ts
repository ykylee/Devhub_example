import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("RepositoryService", () => {
  const apiClientMock = vi.fn();

  beforeEach(() => {
    vi.resetModules();
    apiClientMock.mockReset();
    vi.doMock("@/shared/api/api-client", () => ({
      apiClient: apiClientMock,
    }));
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("listRepositories", () => {
    it("fetches all repositories", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: [
          { id: 1, full_name: "org/repo-a", name: "repo-a", status: "active", private: false, linked_applications_count: 1, linked_projects_count: 0 },
          { id: 2, full_name: "org/repo-b", name: "repo-b", status: "draft", private: true, linked_applications_count: 0, linked_projects_count: 0 },
        ],
      });

      const { repositoryService } = await import("./repository.service");
      const repos = await repositoryService.listRepositories();

      expect(repos).toHaveLength(2);
      expect(repos[0].full_name).toBe("org/repo-a");
      expect(repos[1].status).toBe("draft");
      expect(apiClientMock).toHaveBeenCalledWith("GET", expect.stringContaining("/api/v1/repositories"));
    });

    it("returns empty array when no repositories", async () => {
      apiClientMock.mockResolvedValue({ status: "ok", data: [] });

      const { repositoryService } = await import("./repository.service");
      const repos = await repositoryService.listRepositories();

      expect(repos).toEqual([]);
    });
  });

  describe("getRepository", () => {
    it("finds repository by id from list", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: [
          { id: 1, full_name: "org/repo-a", name: "repo-a", status: "active", private: false, linked_applications_count: 0, linked_projects_count: 0 },
          { id: 2, full_name: "org/repo-b", name: "repo-b", status: "active", private: false, linked_applications_count: 0, linked_projects_count: 0 },
        ],
      });

      const { repositoryService } = await import("./repository.service");
      const repo = await repositoryService.getRepository(2);

      expect(repo).toBeDefined();
      expect(repo!.id).toBe(2);
      expect(repo!.name).toBe("repo-b");
    });

    it("returns undefined when id not found", async () => {
      apiClientMock.mockResolvedValue({ status: "ok", data: [] });

      const { repositoryService } = await import("./repository.service");
      const repo = await repositoryService.getRepository(999);

      expect(repo).toBeUndefined();
    });
  });

  describe("createRepositoryDraft", () => {
    it("sends POST with key, slug and provider_key", async () => {
      apiClientMock.mockResolvedValue({
        status: "created",
        data: { id: 3, full_name: "org/new-repo", name: "new-repo", status: "draft", private: false, linked_applications_count: 0, linked_projects_count: 0 },
      });

      const { repositoryService } = await import("./repository.service");
      const repo = await repositoryService.createRepositoryDraft({
        key: "NEW",
        slug: "new-repo",
        provider_key: "gitea-main",
      });

      expect(repo.id).toBe(3);
      expect(repo.status).toBe("draft");
      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        expect.stringContaining("/api/v1/repositories"),
        { key: "NEW", slug: "new-repo", provider_key: "gitea-main" },
      );
    });
  });

  describe("requestRepositoryPublish", () => {
    it("sends POST to publish endpoint", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: { id: 1, full_name: "org/repo-a", name: "repo-a", status: "active", private: false, linked_applications_count: 1, linked_projects_count: 0 },
      });

      const { repositoryService } = await import("./repository.service");
      const repo = await repositoryService.requestRepositoryPublish(1);

      expect(repo.status).toBe("active");
      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        expect.stringContaining("/api/v1/repositories/1/publish"),
        {},
      );
    });
  });

  describe("getRepositoryActivity", () => {
    it("returns activity metrics", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: {
          repository_id: 1,
          window_from: "2026-05-01",
          window_to: "2026-05-28",
          pr_event_count: 12,
          active_contributors: ["alice", "bob"],
          build_run_count: 45,
          build_success_rate: 0.88,
          last_build_status: "success",
          last_build_at: "2026-05-28T10:00:00Z",
        },
      });

      const { repositoryService } = await import("./repository.service");
      const activity = await repositoryService.getRepositoryActivity(1);

      expect(activity.pr_event_count).toBe(12);
      expect(activity.build_success_rate).toBe(0.88);
      expect(activity.last_build_status).toBe("success");
      expect(apiClientMock).toHaveBeenCalledWith("GET", expect.stringContaining("/api/v1/repositories/1/activity"));
    });
  });

  describe("getRepositoryBuildRuns", () => {
    it("fetches build runs with default options", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: [
          { id: 101, repository_id: 1, branch: "main", status: "success", started_at: "2026-05-28T09:00:00Z" },
        ],
      });

      const { repositoryService } = await import("./repository.service");
      const runs = await repositoryService.getRepositoryBuildRuns(1);

      expect(runs).toHaveLength(1);
      expect(runs[0].status).toBe("success");
      expect(apiClientMock).toHaveBeenCalledWith("GET", expect.stringContaining("/api/v1/repositories/1/build-runs"));
    });

    it("passes query parameters for filtering", async () => {
      apiClientMock.mockResolvedValue({ status: "ok", data: [] });

      const { repositoryService } = await import("./repository.service");
      await repositoryService.getRepositoryBuildRuns(1, { limit: 10, status: "failed", branch: "main" });

      const callUrl = apiClientMock.mock.calls[0][1];
      expect(callUrl).toContain("limit=10");
      expect(callUrl).toContain("status=failed");
      expect(callUrl).toContain("branch=main");
    });
  });
});
