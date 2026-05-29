import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("InfraService", () => {
  const apiClientMock = vi.fn();

  beforeEach(() => {
    vi.resetModules();
    apiClientMock.mockReset();
    vi.spyOn(console, "error").mockImplementation(() => {});
    vi.doMock("@/lib/services/api-client", () => ({
      apiClient: apiClientMock,
    }));
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("getMetrics", () => {
    it("converts ApiMetric response to Metric[] with correct color mapping", async () => {
      apiClientMock.mockResolvedValue({
        data: [
          { label: "Active Users", value: "42", trend: "+5%", trend_direction: "up" },
          { label: "Error Rate", value: "0.5%", trend: "-0.1%", trend_direction: "down" },
          { label: "Uptime", value: "99.9%", trend: "0%", trend_direction: "flat" },
        ],
      });

      const { infraService } = await import("./infra.service");
      const metrics = await infraService.getMetrics("Developer");

      expect(metrics).toEqual([
        { label: "Active Users", value: "42", trend: "+5%", color: "text-emerald-500" },
        { label: "Error Rate", value: "0.5%", trend: "-0.1%", color: "text-rose-500" },
        { label: "Uptime", value: "99.9%", trend: "0%", color: "text-rose-500" },
      ]);
    });

    it("encodes role query parameter", async () => {
      apiClientMock.mockResolvedValue({ data: [] });

      const { infraService } = await import("./infra.service");
      await infraService.getMetrics("System Admin");

      expect(apiClientMock).toHaveBeenCalledWith(
        "GET",
        expect.stringContaining("/api/v1/dashboard/metrics?role=system_admin"),
      );
    });
  });

  describe("getNodes", () => {
    it("maps API response to ServiceNode[] with formatted cpu and memory", async () => {
      apiClientMock.mockResolvedValue({
        data: [
          {
            id: "n1",
            label: "Node 1",
            status: "stable",
            cpu_percent: 45.3,
            memory_bytes: 2_147_483_648,
            kind: "server",
            region: "us-east",
            updated_at: "2026-05-28T12:00:00Z",
          },
          {
            id: "n2",
            label: "Node 2",
            status: "down",
            cpu_percent: 0,
            memory_bytes: 0,
          },
        ],
      });

      const { infraService } = await import("./infra.service");
      const nodes = await infraService.getNodes();

      expect(nodes).toHaveLength(2);
      expect(nodes[0]).toMatchObject({
        id: "n1",
        label: "Node 1",
        status: "stable",
        cpu: "45.3%",
        memory: "2 GB",
        kind: "server",
        region: "us-east",
      });
      expect(nodes[1]).toMatchObject({
        id: "n2",
        label: "Node 2",
        status: "down",
        cpu: "0%",
        memory: "0 B",
      });
    });

    it("returns fallback data when API call fails", async () => {
      apiClientMock.mockRejectedValue(new Error("Network error"));

      const { infraService } = await import("./infra.service");
      const nodes = await infraService.getNodes();

      expect(nodes).toHaveLength(4);
      expect(nodes[0]).toMatchObject({ id: "1", label: "Go Core Service", status: "stable" });
      expect(nodes[2]).toMatchObject({ id: "3", label: "Python AI Engine", status: "warning" });
    });

    it("handles missing optional fields gracefully", async () => {
      apiClientMock.mockResolvedValue({
        data: [{ id: "n1", label: "Minimal Node", status: "stable" }],
      });

      const { infraService } = await import("./infra.service");
      const nodes = await infraService.getNodes();

      expect(nodes[0]).toMatchObject({ cpu: "0%", memory: "0 B" });
    });
  });

  describe("getTopology", () => {
    it("maps API response to nodes and edges", async () => {
      apiClientMock.mockResolvedValue({
        data: {
          nodes: [
            { id: "n1", label: "Web", status: "stable", cpu_percent: 12.5, memory_bytes: 536_870_912 },
            { id: "n2", label: "DB", status: "warning", cpu_percent: 78.1, memory_bytes: 4_294_967_296 },
          ],
          edges: [{ id: "e1", source_id: "n1", target_id: "n2", label: "connects" }],
        },
      });

      const { infraService } = await import("./infra.service");
      const topology = await infraService.getTopology();

      expect(topology.nodes).toHaveLength(2);
      expect(topology.edges).toHaveLength(1);
      expect(topology.nodes[0].cpu).toBe("12.5%");
      expect(topology.nodes[1].memory).toBe("4 GB");
      expect(topology.edges[0]).toMatchObject({ id: "e1", source_id: "n1", target_id: "n2" });
    });

    it("returns empty topology on API error", async () => {
      apiClientMock.mockRejectedValue(new Error("API unavailable"));

      const { infraService } = await import("./infra.service");
      const topology = await infraService.getTopology();

      expect(topology).toEqual({ nodes: [], edges: [] });
    });
  });

  describe("getTopologyV2", () => {
    it("returns parsed v2 topology response", async () => {
      const mockResponse = {
        data: {
          nodes: [{ node_id: "v2-1", hostname: "host-a", status: "healthy" }],
          edges: [{ id: "e1", source_id: "v2-1", target_id: "v2-2" }],
          services: [{ service_id: "s1", node_id: "v2-1", name: "nginx", health_status: "ok" }],
        },
        meta: { snapshot_at: "2026-05-28T12:00:00Z", degraded_providers: [] },
      };
      apiClientMock.mockResolvedValue(mockResponse);

      const { infraService } = await import("./infra.service");
      const result = await infraService.getTopologyV2();

      expect(result.nodes).toHaveLength(1);
      expect(result.nodes[0].hostname).toBe("host-a");
      expect(result.services).toHaveLength(1);
      expect(result.meta.snapshot_at).toBe("2026-05-28T12:00:00Z");
    });
  });

  describe("controlService", () => {
    it("returns true when command_status is pending", async () => {
      apiClientMock.mockResolvedValue({
        data: { command_status: "pending" },
      });

      const { infraService } = await import("./infra.service");
      const result = await infraService.controlService("svc-1", "restart");

      expect(result).toBe(true);
      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        expect.stringContaining("/api/v1/admin/service-actions"),
        expect.objectContaining({
          service_id: "svc-1",
          action_type: "restart",
          dry_run: true,
        }),
      );
    });

    it("returns false when command_status is not pending", async () => {
      apiClientMock.mockResolvedValue({
        data: { command_status: "rejected" },
      });

      const { infraService } = await import("./infra.service");
      const result = await infraService.controlService("svc-2", "stop");

      expect(result).toBe(false);
    });
  });
});
