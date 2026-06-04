import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("DashboardService", () => {
  const apiClientMock = vi.fn();

  beforeEach(() => {
    vi.resetModules();
    apiClientMock.mockReset();
    vi.doMock("@/shared/api/api-client", () => ({ apiClient: apiClientMock }));
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("getDeveloperStream", () => {
    it("issues GET to developer stream endpoint", async () => {
      apiClientMock.mockResolvedValue([]);
      const { dashboardService } = await import("./dashboard.service");

      await dashboardService.getDeveloperStream();

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/dashboard/developer/stream");
    });

    it("maps array response items to DeveloperStreamItem with primary fields", async () => {
      apiClientMock.mockResolvedValue([
        {
          id: "task-1",
          title: "Implement feature",
          repo: "core",
          status: "in_progress",
          progress: 60,
          updated_at: "2026-05-28T10:00:00Z",
        },
      ]);
      const { dashboardService } = await import("./dashboard.service");

      const result = await dashboardService.getDeveloperStream();

      expect(result).toEqual([
        {
          id: "task-1",
          title: "Implement feature",
          repo: "core",
          status: "in_progress",
          progress: 60,
          updated_at: "2026-05-28T10:00:00Z",
        },
      ]);
    });

    it("falls back to task_id / repository / unknown when canonical fields missing", async () => {
      apiClientMock.mockResolvedValue([
        { task_id: "alt-1", name: "Alt", repository: "alt-repo" },
      ]);
      const { dashboardService } = await import("./dashboard.service");

      const result = await dashboardService.getDeveloperStream();

      expect(result[0]).toEqual({
        id: "alt-1",
        title: "Alt",
        repo: "alt-repo",
        status: "unknown",
        progress: undefined,
        updated_at: undefined,
      });
    });

    it("unwraps { data: [...] } envelope wrapper", async () => {
      apiClientMock.mockResolvedValue({ data: [{ id: "t-2", title: "Wrapped" }] });
      const { dashboardService } = await import("./dashboard.service");

      const result = await dashboardService.getDeveloperStream();

      expect(result).toHaveLength(1);
      expect(result[0].id).toBe("t-2");
      expect(result[0].title).toBe("Wrapped");
    });

    it("assigns index-based fallback id when no id/task_id/key", async () => {
      apiClientMock.mockResolvedValue([{ title: "No-ID" }]);
      const { dashboardService } = await import("./dashboard.service");

      const result = await dashboardService.getDeveloperStream();

      expect(result[0].id).toBe("stream-0");
      expect(result[0].repo).toBe("-");
    });

    it("returns empty array when response is neither array nor data-wrapped", async () => {
      apiClientMock.mockResolvedValue({ unrelated: "x" });
      const { dashboardService } = await import("./dashboard.service");

      expect(await dashboardService.getDeveloperStream()).toEqual([]);
    });
  });

  describe("getDeveloperBuilds", () => {
    it("issues GET to developer builds endpoint", async () => {
      apiClientMock.mockResolvedValue([]);
      const { dashboardService } = await import("./dashboard.service");

      await dashboardService.getDeveloperBuilds();

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/dashboard/developer/builds");
    });

    it("maps canonical build fields", async () => {
      apiClientMock.mockResolvedValue([
        {
          id: "b-1",
          title: "Build 1",
          status: "success",
          duration: "2m",
          commit_sha: "abc123",
          finished_at: "2026-05-28T11:00:00Z",
        },
      ]);
      const { dashboardService } = await import("./dashboard.service");

      const result = await dashboardService.getDeveloperBuilds();

      expect(result[0]).toEqual({
        id: "b-1",
        title: "Build 1",
        status: "success",
        duration: "2m",
        commit_sha: "abc123",
        finished_at: "2026-05-28T11:00:00Z",
      });
    });

    it("falls back to build_id and sha", async () => {
      apiClientMock.mockResolvedValue([{ build_id: "alt", sha: "def" }]);
      const { dashboardService } = await import("./dashboard.service");

      const result = await dashboardService.getDeveloperBuilds();

      expect(result[0].id).toBe("alt");
      expect(result[0].title).toBe("Build");
      expect(result[0].commit_sha).toBe("def");
    });

    it("assigns index-based id when no build_id", async () => {
      apiClientMock.mockResolvedValue([{ title: "X" }]);
      const { dashboardService } = await import("./dashboard.service");

      const result = await dashboardService.getDeveloperBuilds();

      expect(result[0].id).toBe("build-0");
    });
  });

  describe("getManagerVelocity", () => {
    it("issues GET to manager velocity endpoint", async () => {
      apiClientMock.mockResolvedValue([]);
      const { dashboardService } = await import("./dashboard.service");

      await dashboardService.getManagerVelocity();

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/dashboard/manager/velocity");
    });

    it("maps canonical velocity fields with numeric coercion", async () => {
      apiClientMock.mockResolvedValue([
        { name: "Mon", quality: 85, security: 90 },
      ]);
      const { dashboardService } = await import("./dashboard.service");

      const result = await dashboardService.getManagerVelocity();

      expect(result[0]).toEqual({ name: "Mon", quality: 85, security: 90 });
    });

    it("falls back to date/quality_score/security_score", async () => {
      apiClientMock.mockResolvedValue([
        { date: "2026-05-28", quality_score: "70", security_score: "75" },
      ]);
      const { dashboardService } = await import("./dashboard.service");

      const result = await dashboardService.getManagerVelocity();

      expect(result[0]).toEqual({ name: "2026-05-28", quality: 70, security: 75 });
    });

    it("uses default '-' name and 0 numbers when fields missing", async () => {
      apiClientMock.mockResolvedValue([{}]);
      const { dashboardService } = await import("./dashboard.service");

      const result = await dashboardService.getManagerVelocity();

      expect(result[0]).toEqual({ name: "-", quality: 0, security: 0 });
    });
  });

  describe("getManagerTeamLoad", () => {
    it("issues GET to manager team-load endpoint", async () => {
      apiClientMock.mockResolvedValue([]);
      const { dashboardService } = await import("./dashboard.service");

      await dashboardService.getManagerTeamLoad();

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/dashboard/manager/team-load");
    });

    it("maps canonical team load fields with numeric coercion", async () => {
      apiClientMock.mockResolvedValue([
        { name: "Alice", load: 75, status: "busy" },
      ]);
      const { dashboardService } = await import("./dashboard.service");

      const result = await dashboardService.getManagerTeamLoad();

      expect(result[0]).toEqual({ name: "Alice", load: 75, status: "busy" });
    });

    it("falls back to user_name and load_percent", async () => {
      apiClientMock.mockResolvedValue([
        { user_name: "Bob", load_percent: 50, status: "available" },
      ]);
      const { dashboardService } = await import("./dashboard.service");

      const result = await dashboardService.getManagerTeamLoad();

      expect(result[0]).toEqual({ name: "Bob", load: 50, status: "available" });
    });
  });

  describe("getManagerDecisions", () => {
    it("issues GET to manager decisions endpoint", async () => {
      apiClientMock.mockResolvedValue([]);
      const { dashboardService } = await import("./dashboard.service");

      await dashboardService.getManagerDecisions();

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/dashboard/manager/decisions");
    });

    it("maps canonical decision fields", async () => {
      apiClientMock.mockResolvedValue([
        { id: "d-1", title: "Decision 1", type: "approval", occurred_at: "2026-05-28T12:00:00Z" },
      ]);
      const { dashboardService } = await import("./dashboard.service");

      const result = await dashboardService.getManagerDecisions();

      expect(result[0]).toEqual({
        id: "d-1",
        title: "Decision 1",
        type: "approval",
        occurred_at: "2026-05-28T12:00:00Z",
      });
    });

    it("falls back to decision_id / summary / category / created_at", async () => {
      apiClientMock.mockResolvedValue([
        { decision_id: "alt", summary: "Sum", category: "policy", created_at: "2026-05-28T13:00:00Z" },
      ]);
      const { dashboardService } = await import("./dashboard.service");

      const result = await dashboardService.getManagerDecisions();

      expect(result[0]).toEqual({
        id: "alt",
        title: "Sum",
        type: "policy",
        occurred_at: "2026-05-28T13:00:00Z",
      });
    });

    it("uses index-based id + ISO occurred_at fallback when missing", async () => {
      apiClientMock.mockResolvedValue([{}]);
      const { dashboardService } = await import("./dashboard.service");

      const result = await dashboardService.getManagerDecisions();

      expect(result[0].id).toBe("decision-0");
      expect(result[0].title).toBe("Decision");
      expect(result[0].type).toBe("general");
      // occurred_at must be a valid ISO date string (the fallback uses new Date().toISOString())
      expect(Number.isNaN(Date.parse(result[0].occurred_at))).toBe(false);
    });
  });
});
