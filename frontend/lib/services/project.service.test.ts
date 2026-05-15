import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("projectService", () => {
  const apiClientMock = vi.fn();

  beforeEach(() => {
    vi.resetModules();
    apiClientMock.mockReset();
    vi.doMock("./api-client", () => ({
      apiClient: apiClientMock,
    }));
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("serializes query params for getApplications", async () => {
    const payload = { data: [] };
    apiClientMock.mockResolvedValue(payload);
    const { projectService } = await import("./project.service");

    await projectService.getApplications({ status: "active", include_archived: true, q: "devhub" });

    expect(apiClientMock).toHaveBeenCalledWith(
      "GET",
      "/api/v1/applications?status=active&include_archived=true&q=devhub",
    );
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
});
