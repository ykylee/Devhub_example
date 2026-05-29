import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("GardenerService", () => {
  const apiClientMock = vi.fn();

  beforeEach(() => {
    vi.resetModules();
    apiClientMock.mockReset();
    vi.spyOn(console, "error").mockImplementation(() => {});
    vi.doMock("@/shared/api/api-client", () => ({ apiClient: apiClientMock }));
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("getInstance", () => {
    it("returns a stable singleton across multiple calls", async () => {
      const { GardenerService } = await import("./gardener.service");

      const a = GardenerService.getInstance();
      const b = GardenerService.getInstance();

      expect(a).toBe(b);
    });

    it("exports a default gardenerService instance backed by the singleton", async () => {
      const { gardenerService, GardenerService } = await import("./gardener.service");

      expect(gardenerService).toBe(GardenerService.getInstance());
    });
  });

  describe("getSuggestions", () => {
    it("issues GET to suggestions endpoint and unwraps data field", async () => {
      const suggestions = [
        {
          id: "sug-a",
          title: "Scale up",
          description: "Add capacity",
          type: "scaling",
          impact: "high",
          auto_fixable: true,
          created_at: "2026-05-28T00:00:00Z",
        },
      ];
      apiClientMock.mockResolvedValue({ data: suggestions });

      const { gardenerService } = await import("./gardener.service");
      const result = await gardenerService.getSuggestions();

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/gardener/suggestions");
      expect(result).toEqual(suggestions);
    });

    it("returns mock fallback array when apiClient throws", async () => {
      apiClientMock.mockRejectedValue(new Error("network"));

      const { gardenerService } = await import("./gardener.service");
      const result = await gardenerService.getSuggestions();

      // Two stable mock items per the implementation (Phase 4 prototyping fallback).
      expect(result).toHaveLength(2);
      expect(result[0].id).toBe("sug-001");
      expect(result[1].id).toBe("sug-002");
      expect(result[0].type).toBe("scaling");
      expect(result[1].type).toBe("optimization");
    });

    it("logs the error before falling back when apiClient rejects", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      apiClientMock.mockRejectedValue(new Error("boom"));

      const { gardenerService } = await import("./gardener.service");
      await gardenerService.getSuggestions();

      expect(errorSpy).toHaveBeenCalled();
      expect(String(errorSpy.mock.calls[0][0])).toContain("GardenerService");
    });
  });

  describe("applySuggestion", () => {
    it("issues POST to apply endpoint and maps command_status to status", async () => {
      apiClientMock.mockResolvedValue({
        data: { command_id: "cmd-1", command_status: "queued" },
      });

      const { gardenerService } = await import("./gardener.service");
      const result = await gardenerService.applySuggestion("sug-1");

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        "/api/v1/gardener/suggestions/sug-1/apply",
      );
      expect(result).toEqual({ command_id: "cmd-1", status: "queued" });
    });

    it("propagates error when apiClient rejects (no fallback for write path)", async () => {
      apiClientMock.mockRejectedValue(new Error("denied"));

      const { gardenerService } = await import("./gardener.service");

      await expect(gardenerService.applySuggestion("sug-x")).rejects.toThrow("denied");
    });

    it("URL-encodes the suggestion id segment literally (no escaping inside template)", async () => {
      apiClientMock.mockResolvedValue({
        data: { command_id: "cmd-2", command_status: "accepted" },
      });

      const { gardenerService } = await import("./gardener.service");
      await gardenerService.applySuggestion("sug with space");

      const [, path] = apiClientMock.mock.calls[0];
      expect(path).toBe("/api/v1/gardener/suggestions/sug with space/apply");
    });
  });
});
