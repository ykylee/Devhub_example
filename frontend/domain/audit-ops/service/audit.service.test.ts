import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { AuditLogEntry, AuditLogListMeta } from "@/domain/audit-ops/schema/audit.types";

describe("AuditService", () => {
  const apiClientMock = vi.fn();

  beforeEach(() => {
    vi.resetModules();
    apiClientMock.mockReset();
    vi.doMock("@/shared/api/api-client", () => ({ apiClient: apiClientMock }));
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("getLogs", () => {
    it("issues GET without query string when no filters", async () => {
      apiClientMock.mockResolvedValue({ data: [], meta: {} });
      const { auditService } = await import("./audit.service");

      const result = await auditService.getLogs();

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/audit/logs");
      expect(result.entries).toEqual([]);
      expect(result.meta).toEqual({});
    });

    it("encodes all supported filter fields as query parameters", async () => {
      apiClientMock.mockResolvedValue({ data: [], meta: {} });
      const { auditService } = await import("./audit.service");

      await auditService.getLogs({
        actor_login: "alice",
        action: "command.execute",
        target_type: "command",
        target_id: "cmd-1",
        command_id: "cmd-1",
        limit: 25,
        offset: 50,
      });

      const [, path] = apiClientMock.mock.calls[0];
      const url = new URL(path as string, "http://example");
      expect(url.pathname).toBe("/api/v1/audit/logs");
      expect(url.searchParams.get("actor_login")).toBe("alice");
      expect(url.searchParams.get("action")).toBe("command.execute");
      expect(url.searchParams.get("target_type")).toBe("command");
      expect(url.searchParams.get("target_id")).toBe("cmd-1");
      expect(url.searchParams.get("command_id")).toBe("cmd-1");
      expect(url.searchParams.get("limit")).toBe("25");
      expect(url.searchParams.get("offset")).toBe("50");
    });

    it("omits undefined filter fields entirely (only sets explicit values)", async () => {
      apiClientMock.mockResolvedValue({ data: [], meta: {} });
      const { auditService } = await import("./audit.service");

      await auditService.getLogs({ actor_login: "bob" });

      const [, path] = apiClientMock.mock.calls[0];
      expect(path).toBe("/api/v1/audit/logs?actor_login=bob");
    });

    it("treats limit/offset = 0 as explicit (not skipped)", async () => {
      apiClientMock.mockResolvedValue({ data: [], meta: {} });
      const { auditService } = await import("./audit.service");

      await auditService.getLogs({ limit: 0, offset: 0 });

      const [, path] = apiClientMock.mock.calls[0];
      const url = new URL(path as string, "http://example");
      expect(url.searchParams.get("limit")).toBe("0");
      expect(url.searchParams.get("offset")).toBe("0");
    });

    it("returns entries + meta from API response", async () => {
      const entries: AuditLogEntry[] = [
        {
          audit_id: "a-1",
          actor_login: "alice",
          action: "command.execute",
          target_type: "command",
          target_id: "cmd-1",
          payload: { result: "ok" },
          created_at: "2026-05-28T00:00:00Z",
        },
      ];
      const meta: AuditLogListMeta = { limit: 50, offset: 0, count: 1 };
      apiClientMock.mockResolvedValue({ data: entries, meta });

      const { auditService } = await import("./audit.service");
      const result = await auditService.getLogs();

      expect(result.entries).toEqual(entries);
      expect(result.meta).toEqual(meta);
    });

    it("defaults entries to [] when API returns null data", async () => {
      apiClientMock.mockResolvedValue({ data: null, meta: undefined });

      const { auditService } = await import("./audit.service");
      const result = await auditService.getLogs();

      expect(result.entries).toEqual([]);
      expect(result.meta).toEqual({});
    });

    it("propagates apiClient errors without wrapping", async () => {
      const err = new Error("403 forbidden");
      apiClientMock.mockRejectedValue(err);

      const { auditService } = await import("./audit.service");
      await expect(auditService.getLogs()).rejects.toBe(err);
    });
  });

  describe("singleton", () => {
    it("getInstance returns the same instance on repeated calls", async () => {
      apiClientMock.mockResolvedValue({ data: [], meta: {} });
      const mod = await import("./audit.service");
      const a = mod.auditService;
      const b = mod.auditService;
      expect(a).toBe(b);
    });
  });
});
