import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("ApplicationService", () => {
  const apiClientMock = vi.fn();

  beforeEach(() => {
    vi.resetModules();
    apiClientMock.mockReset();
    vi.doMock("@/lib/services/api-client", () => ({
      apiClient: apiClientMock,
    }));
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("listApplications", () => {
    it("fetches applications without options", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: [
          { id: "app-1", key: "APP1", name: "App One", status: "active" },
        ],
        meta: { total: 1 },
      });

      const { applicationService } = await import("./application.service");
      const apps = await applicationService.listApplications();

      expect(apps).toHaveLength(1);
      expect(apps[0].id).toBe("app-1");
      expect(apiClientMock).toHaveBeenCalledWith("GET", expect.stringContaining("/api/v1/applications"));
    });

    it("passes query parameters", async () => {
      apiClientMock.mockResolvedValue({ status: "ok", data: [], meta: { total: 0 } });

      const { applicationService } = await import("./application.service");
      await applicationService.listApplications({ status: "active", q: "devhub", include_archived: true });

      expect(apiClientMock).toHaveBeenCalledWith(
        "GET",
        expect.stringContaining("status=active&q=devhub&include_archived=true"),
      );
    });

    it("returns empty array when no applications", async () => {
      apiClientMock.mockResolvedValue({ status: "ok", data: [], meta: { total: 0 } });

      const { applicationService } = await import("./application.service");
      const apps = await applicationService.listApplications();

      expect(apps).toEqual([]);
    });
  });

  describe("getApplication", () => {
    it("fetches a single application by id", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: { id: "app-42", key: "FORTYTWO", name: "Answer", status: "active" },
      });

      const { applicationService } = await import("./application.service");
      const app = await applicationService.getApplication("app-42");

      expect(app.id).toBe("app-42");
      expect(app.name).toBe("Answer");
      expect(apiClientMock).toHaveBeenCalledWith("GET", expect.stringContaining("/api/v1/applications/app-42"));
    });
  });

  describe("getApplicationRollup", () => {
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

      const { applicationService } = await import("./application.service");
      const rollup = await applicationService.getApplicationRollup("app-1");

      expect(rollup.build_success_rate).toBe(0.92);
      expect(rollup.target_branch_build_status).toBe("healthy");
      expect(rollup.pull_request_distribution).toEqual({ open: 5, merged: 12 });
      expect(apiClientMock).toHaveBeenCalledWith("GET", expect.stringContaining("/api/v1/applications/app-1/rollup"));
    });
  });

  describe("listApplications param branches", () => {
    it("omits query string entirely when options object is empty", async () => {
      apiClientMock.mockResolvedValue({ status: "ok", data: [], meta: { total: 0 } });

      const { applicationService } = await import("./application.service");
      await applicationService.listApplications();

      const [, url] = apiClientMock.mock.calls[0];
      expect(url).not.toContain("?");
    });

    it("passes only status when q and include_archived are omitted", async () => {
      apiClientMock.mockResolvedValue({ status: "ok", data: [], meta: { total: 0 } });

      const { applicationService } = await import("./application.service");
      await applicationService.listApplications({ status: "archived" });

      const [, url] = apiClientMock.mock.calls[0];
      expect(url).toContain("?status=archived");
      expect(url).not.toContain("q=");
      expect(url).not.toContain("include_archived=");
    });

    it("passes include_archived=false explicitly (not skipped as falsy)", async () => {
      apiClientMock.mockResolvedValue({ status: "ok", data: [], meta: { total: 0 } });

      const { applicationService } = await import("./application.service");
      await applicationService.listApplications({ include_archived: false });

      const [, url] = apiClientMock.mock.calls[0];
      expect(url).toContain("include_archived=false");
    });

    it("propagates apiClient errors", async () => {
      apiClientMock.mockRejectedValue(new Error("network"));

      const { applicationService } = await import("./application.service");
      await expect(applicationService.listApplications()).rejects.toThrow("network");
    });
  });

  describe("getApplication error", () => {
    it("propagates apiClient errors", async () => {
      apiClientMock.mockRejectedValue(new Error("404"));

      const { applicationService } = await import("./application.service");
      await expect(applicationService.getApplication("missing")).rejects.toThrow("404");
    });
  });

  describe("getApplicationRollup error", () => {
    it("propagates apiClient errors", async () => {
      apiClientMock.mockRejectedValue(new Error("500"));

      const { applicationService } = await import("./application.service");
      await expect(applicationService.getApplicationRollup("app-1")).rejects.toThrow("500");
    });
  });

  describe("getApplicationDashboard", () => {
    it("returns full dashboard data", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: {
          application_id: "app-1",
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

      const { applicationService } = await import("./application.service");
      const dashboard = await applicationService.getApplicationDashboard("app-1");

      expect(dashboard.application_id).toBe("app-1");
      expect(dashboard.metrics_overview.target_branch_build_status).toBe("healthy");
      expect(dashboard.projects_progress).toHaveLength(1);
      expect(dashboard.projects_progress[0].risk_level).toBe("low");
      expect(apiClientMock).toHaveBeenCalledWith("GET", expect.stringContaining("/api/v1/applications/app-1/dashboard"));
    });
  });
});
