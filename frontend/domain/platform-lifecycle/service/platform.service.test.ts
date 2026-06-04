import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("PlatformService", () => {
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

  describe("listPlatforms", () => {
    it("fetches platforms without options", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: [
          { id: "app-1", key: "APP1", name: "App One", status: "active" },
        ],
        meta: { total: 1 },
      });

      const { platformService } = await import("./platform.service");
      const apps = await platformService.listPlatforms();

      expect(apps).toHaveLength(1);
      expect(apps[0].id).toBe("app-1");
      expect(apiClientMock).toHaveBeenCalledWith("GET", expect.stringContaining("/api/v1/platforms"));
    });

    it("passes query parameters", async () => {
      apiClientMock.mockResolvedValue({ status: "ok", data: [], meta: { total: 0 } });

      const { platformService } = await import("./platform.service");
      await platformService.listPlatforms({ status: "active", q: "devhub", include_archived: true });

      expect(apiClientMock).toHaveBeenCalledWith(
        "GET",
        expect.stringContaining("status=active&q=devhub&include_archived=true"),
      );
    });

    it("returns empty array when no platforms", async () => {
      apiClientMock.mockResolvedValue({ status: "ok", data: [], meta: { total: 0 } });

      const { platformService } = await import("./platform.service");
      const apps = await platformService.listPlatforms();

      expect(apps).toEqual([]);
    });
  });

  describe("getPlatform", () => {
    it("fetches a single platform by id", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: { id: "app-42", key: "FORTYTWO", name: "Answer", status: "active" },
      });

      const { platformService } = await import("./platform.service");
      const app = await platformService.getPlatform("app-42");

      expect(app.id).toBe("app-42");
      expect(app.name).toBe("Answer");
      expect(apiClientMock).toHaveBeenCalledWith("GET", expect.stringContaining("/api/v1/platforms/app-42"));
    });
  });

  describe("getPlatformRollup", () => {
    it("returns rollup metrics", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: {
          pull_request_distribution: { open: 5, merged: 12 },
          build_success_rate: 0.92,
          build_avg_duration_seconds: 180,
          quality_score: 85,
          quality_gate_failed_count: 1,
          critical_warning_count: 0,
          target_branch_build_status: "healthy",
        },
      });

      const { platformService } = await import("./platform.service");
      const rollup = await platformService.getPlatformRollup("app-1");

      expect(rollup.build_success_rate).toBe(0.92);
      expect(rollup.target_branch_build_status).toBe("healthy");
      expect(rollup.pull_request_distribution).toEqual({ open: 5, merged: 12 });
      expect(apiClientMock).toHaveBeenCalledWith("GET", expect.stringContaining("/api/v1/platforms/app-1/rollup"));
    });
  });

  describe("listPlatforms param branches", () => {
    it("omits query string entirely when options object is empty", async () => {
      apiClientMock.mockResolvedValue({ status: "ok", data: [], meta: { total: 0 } });

      const { platformService } = await import("./platform.service");
      await platformService.listPlatforms();

      const [, url] = apiClientMock.mock.calls[0];
      expect(url).not.toContain("?");
    });

    it("passes only status when q and include_archived are omitted", async () => {
      apiClientMock.mockResolvedValue({ status: "ok", data: [], meta: { total: 0 } });

      const { platformService } = await import("./platform.service");
      await platformService.listPlatforms({ status: "archived" });

      const [, url] = apiClientMock.mock.calls[0];
      expect(url).toContain("?status=archived");
      expect(url).not.toContain("q=");
      expect(url).not.toContain("include_archived=");
    });

    it("passes include_archived=false explicitly (not skipped as falsy)", async () => {
      apiClientMock.mockResolvedValue({ status: "ok", data: [], meta: { total: 0 } });

      const { platformService } = await import("./platform.service");
      await platformService.listPlatforms({ include_archived: false });

      const [, url] = apiClientMock.mock.calls[0];
      expect(url).toContain("include_archived=false");
    });

    it("propagates apiClient errors", async () => {
      apiClientMock.mockRejectedValue(new Error("network"));

      const { platformService } = await import("./platform.service");
      await expect(platformService.listPlatforms()).rejects.toThrow("network");
    });
  });

  describe("getPlatform error", () => {
    it("propagates apiClient errors", async () => {
      apiClientMock.mockRejectedValue(new Error("404"));

      const { platformService } = await import("./platform.service");
      await expect(platformService.getPlatform("missing")).rejects.toThrow("404");
    });
  });

  describe("getPlatformRollup error", () => {
    it("propagates apiClient errors", async () => {
      apiClientMock.mockRejectedValue(new Error("500"));

      const { platformService } = await import("./platform.service");
      await expect(platformService.getPlatformRollup("app-1")).rejects.toThrow("500");
    });
  });

  describe("getPlatformDashboard", () => {
    it("returns full dashboard data", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: {
          platform_id: "app-1",
          key: "APP1",
          name: "App One",
          status: "active",
          visibility: "internal",
          leader: "Alice",
          development_unit: "Platform",
          updated_at: "2026-05-28T12:00:00Z",
          metrics_overview: {
            target_branch_build_status: "healthy",
            avg_build_duration_seconds: 180,
            quality_score: 85,
            critical_warning_count: 0,
          },
          build_failures: [],
          quality_metrics: {
            normalized_score: 85,
            unresolved_issues: { blocker: 0, critical: 1, major: 3 },
            comment: "Good",
          },
          projects_progress: [{
            project_id: "proj-1",
            key: "PROJ1",
            name: "Project One",
            progress_percent: 75,
            status: "active",
            due_date: "2026-06-15",
            d_day: 18,
            risk_level: "low",
            risk_badge_color: "text-emerald-500",
          }],
          linked_dev_requests: [],
          history_trend: [{
            date: "2026-05-01",
            avg_duration_seconds: 200,
            build_success_rate: 0.85,
            quality_score: 80,
          }],
        },
      });

      const { platformService } = await import("./platform.service");
      const dashboard = await platformService.getPlatformDashboard("app-1");

      expect(dashboard.platform_id).toBe("app-1");
      expect(dashboard.metrics_overview.target_branch_build_status).toBe("healthy");
      expect(dashboard.projects_progress).toHaveLength(1);
      expect(dashboard.projects_progress[0].risk_level).toBe("low");
      expect(apiClientMock).toHaveBeenCalledWith("GET", expect.stringContaining("/api/v1/platforms/app-1/dashboard"));
    });
  });
});
