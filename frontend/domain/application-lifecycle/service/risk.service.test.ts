import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("RiskService", () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    vi.resetModules();
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    vi.spyOn(console, "error").mockImplementation(() => {});
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  describe("getInstance", () => {
    it("returns a stable singleton across multiple calls", async () => {
      const { RiskService } = await import("./risk.service");

      const a = RiskService.getInstance();
      const b = RiskService.getInstance();

      expect(a).toBe(b);
    });

    it("exports a default riskService instance backed by the singleton", async () => {
      const { riskService, RiskService } = await import("./risk.service");

      expect(riskService).toBe(RiskService.getInstance());
    });
  });

  describe("getCriticalRisks", () => {
    it("issues GET to /api/v1/risks/critical and maps API response", async () => {
      fetchMock.mockResolvedValue({
        ok: true,
        json: async () => ({
          status: "ok",
          data: [
            {
              id: "r-1",
              title: "Pipeline failure",
              reason: "Build flaky",
              impact: "High",
              status: "Open",
              owner_login: "alice",
              suggested_actions: ["retry", "investigate"],
              created_at: "2026-05-28T00:00:00Z",
            },
          ],
        }),
      });

      const { riskService } = await import("./risk.service");
      const result = await riskService.getCriticalRisks();

      expect(fetchMock).toHaveBeenCalledWith("/api/v1/risks/critical");
      expect(result).toEqual([
        {
          id: "r-1",
          title: "Pipeline failure",
          reason: "Build flaky",
          impact: "High",
          status: "Open",
          owner: "alice",
          owner_login: "alice",
          suggested_actions: ["retry", "investigate"],
          created_at: "2026-05-28T00:00:00Z",
        },
      ]);
    });

    it("returns mock fallback array when response.ok is false", async () => {
      fetchMock.mockResolvedValue({
        ok: false,
        json: async () => ({ status: "rejected", error: "boom" }),
      });

      const { riskService } = await import("./risk.service");
      const result = await riskService.getCriticalRisks();

      expect(result).toHaveLength(2);
      expect(result[0].title).toBe("Gitea Migration Blocked");
      expect(result[1].title).toBe("Frontend CI Pipeline Delay");
    });

    it("returns mock fallback when fetch itself rejects (network error)", async () => {
      fetchMock.mockRejectedValue(new Error("ENETDOWN"));

      const { riskService } = await import("./risk.service");
      const result = await riskService.getCriticalRisks();

      expect(result).toHaveLength(2);
      expect(result[0].owner).toBe("Alex K.");
      expect(result[1].owner).toBe("YK Lee");
    });

    it("logs the error before falling back", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      fetchMock.mockResolvedValue({
        ok: false,
        json: async () => ({ status: "rejected" }),
      });

      const { riskService } = await import("./risk.service");
      await riskService.getCriticalRisks();

      expect(errorSpy).toHaveBeenCalled();
      expect(String(errorSpy.mock.calls[0][0])).toContain("RiskService");
    });
  });

  describe("applyMitigation", () => {
    it("issues POST with X-Devhub-Actor header + JSON body and maps command_status to status", async () => {
      fetchMock.mockResolvedValue({
        ok: true,
        json: async () => ({
          status: "ok",
          data: {
            command_id: "cmd-99",
            command_status: "queued",
            requires_approval: false,
            audit_log_id: "log-1",
            idempotent_replay: false,
            created_at: "2026-05-28T01:00:00Z",
          },
        }),
      });

      const { riskService } = await import("./risk.service");
      const result = await riskService.applyMitigation("r-1", "retry", "yklee");

      expect(fetchMock).toHaveBeenCalledTimes(1);
      const [url, init] = fetchMock.mock.calls[0];
      expect(url).toBe("/api/v1/risks/r-1/mitigations");
      expect(init.method).toBe("POST");
      expect(init.headers["Content-Type"]).toBe("application/json");
      expect(init.headers["X-Devhub-Actor"]).toBe("yklee");

      const body = JSON.parse(init.body as string);
      expect(body.action_type).toBe("retry");
      expect(body.reason).toBe("Manual mitigation from Manager Dashboard");
      expect(body.idempotency_key).toContain("mitigation-r-1-retry-");

      expect(result).toEqual({ command_id: "cmd-99", status: "queued" });
    });

    it("throws when response.ok is false (no fallback for write path)", async () => {
      fetchMock.mockResolvedValue({
        ok: false,
        json: async () => ({ status: "rejected", error: "forbidden" }),
      });

      const { riskService } = await import("./risk.service");

      await expect(riskService.applyMitigation("r-1", "retry")).rejects.toThrow(
        "Failed to apply mitigation",
      );
    });

    it("throws when fetch rejects (network error)", async () => {
      fetchMock.mockRejectedValue(new Error("ENETDOWN"));

      const { riskService } = await import("./risk.service");

      await expect(riskService.applyMitigation("r-2", "investigate")).rejects.toThrow(
        "ENETDOWN",
      );
    });

    it("generates a unique idempotency key per call (Date.now() suffix)", async () => {
      fetchMock.mockResolvedValue({
        ok: true,
        json: async () => ({
          status: "ok",
          data: {
            command_id: "cmd-1",
            command_status: "queued",
            requires_approval: false,
            audit_log_id: "log",
            idempotent_replay: false,
            created_at: "2026-05-28T00:00:00Z",
          },
        }),
      });

      const { riskService } = await import("./risk.service");
      await riskService.applyMitigation("r-3", "retry");
      // Advance time to ensure Date.now() differs
      await new Promise((resolve) => setTimeout(resolve, 5));
      await riskService.applyMitigation("r-3", "retry");

      const body1 = JSON.parse(fetchMock.mock.calls[0][1].body as string);
      const body2 = JSON.parse(fetchMock.mock.calls[1][1].body as string);
      expect(body1.idempotency_key).not.toBe(body2.idempotency_key);
    });
  });
});
