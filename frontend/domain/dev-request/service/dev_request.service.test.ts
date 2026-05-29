import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { DevRequest, DevRequestRegisterPayload } from "@/domain/dev-request/schema/dev_request.types";

describe("DevRequestService", () => {
  const apiClientMock = vi.fn();

  beforeEach(() => {
    vi.resetModules();
    apiClientMock.mockReset();
    vi.doMock("@/lib/services/api-client", () => ({ apiClient: apiClientMock }));
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  const sample: DevRequest = {
    id: "dr-1",
    title: "Example",
  } as DevRequest;

  describe("list", () => {
    it("issues GET without query when no params given", async () => {
      apiClientMock.mockResolvedValue({ data: [sample], meta: { total: 1 } });

      const { devRequestService } = await import("./dev_request.service");
      const result = await devRequestService.list();

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/dev-requests");
      expect(result.data).toEqual([sample]);
      expect(result.total).toBe(1);
    });

    it("encodes single status string in query", async () => {
      apiClientMock.mockResolvedValue({ data: [], meta: { total: 0 } });

      const { devRequestService } = await import("./dev_request.service");
      await devRequestService.list({ status: "pending" });

      expect(apiClientMock).toHaveBeenCalledWith(
        "GET",
        "/api/v1/dev-requests?status=pending",
      );
    });

    it("joins status array with comma", async () => {
      apiClientMock.mockResolvedValue({ data: [], meta: { total: 0 } });

      const { devRequestService } = await import("./dev_request.service");
      await devRequestService.list({ status: ["pending", "in_review"] });

      expect(apiClientMock).toHaveBeenCalledWith(
        "GET",
        expect.stringContaining("status=pending%2Cin_review"),
      );
    });

    it("encodes all supported filter params", async () => {
      apiClientMock.mockResolvedValue({ data: [], meta: { total: 0 } });

      const { devRequestService } = await import("./dev_request.service");
      await devRequestService.list({
        status: "in_review",
        assignee_user_id: "alice",
        source_system: "jira",
        limit: 25,
        offset: 50,
      });

      const [, path] = apiClientMock.mock.calls[0];
      const url = new URL(path as string, "http://example");
      expect(url.searchParams.get("status")).toBe("in_review");
      expect(url.searchParams.get("assignee_user_id")).toBe("alice");
      expect(url.searchParams.get("source_system")).toBe("jira");
      expect(url.searchParams.get("limit")).toBe("25");
      expect(url.searchParams.get("offset")).toBe("50");
    });

    it("skips undefined and empty-string filter values", async () => {
      apiClientMock.mockResolvedValue({ data: [], meta: { total: 0 } });

      const { devRequestService } = await import("./dev_request.service");
      await devRequestService.list({ assignee_user_id: "", source_system: undefined });

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/dev-requests");
    });

    it("falls back total = data.length when meta is absent", async () => {
      apiClientMock.mockResolvedValue({ data: [sample, sample] });

      const { devRequestService } = await import("./dev_request.service");
      const result = await devRequestService.list();

      expect(result.total).toBe(2);
    });
  });

  describe("get", () => {
    it("issues GET /api/v1/dev-requests/:id and returns data", async () => {
      apiClientMock.mockResolvedValue({ data: sample });

      const { devRequestService } = await import("./dev_request.service");
      const result = await devRequestService.get("dr-1");

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/dev-requests/dr-1");
      expect(result).toEqual(sample);
    });
  });

  describe("register", () => {
    it("POSTs payload to /:id/register and returns dev_request from envelope", async () => {
      const payload: DevRequestRegisterPayload = {
        target_type: "application",
        target_id: "app-1",
      };
      apiClientMock.mockResolvedValue({ data: { dev_request: sample } });

      const { devRequestService } = await import("./dev_request.service");
      const result = await devRequestService.register("dr-1", payload);

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        "/api/v1/dev-requests/dr-1/register",
        payload,
      );
      expect(result).toEqual(sample);
    });
  });

  describe("reject", () => {
    it("POSTs rejected_reason to /:id/reject", async () => {
      apiClientMock.mockResolvedValue({ data: sample });

      const { devRequestService } = await import("./dev_request.service");
      await devRequestService.reject("dr-1", "duplicate");

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        "/api/v1/dev-requests/dr-1/reject",
        { rejected_reason: "duplicate" },
      );
    });
  });

  describe("reassign", () => {
    it("PATCHes assignee_user_id", async () => {
      apiClientMock.mockResolvedValue({ data: sample });

      const { devRequestService } = await import("./dev_request.service");
      await devRequestService.reassign("dr-1", "bob");

      expect(apiClientMock).toHaveBeenCalledWith(
        "PATCH",
        "/api/v1/dev-requests/dr-1",
        { assignee_user_id: "bob" },
      );
    });
  });

  describe("close", () => {
    it("DELETEs /:id and returns data", async () => {
      apiClientMock.mockResolvedValue({ data: sample });

      const { devRequestService } = await import("./dev_request.service");
      const result = await devRequestService.close("dr-1");

      expect(apiClientMock).toHaveBeenCalledWith("DELETE", "/api/v1/dev-requests/dr-1");
      expect(result).toEqual(sample);
    });
  });

  describe("getMyPending", () => {
    it("delegates to list with status pending+in_review", async () => {
      apiClientMock.mockResolvedValue({ data: [sample], meta: { total: 1 } });

      const { devRequestService } = await import("./dev_request.service");
      await devRequestService.getMyPending();

      expect(apiClientMock).toHaveBeenCalledWith(
        "GET",
        expect.stringContaining("status=pending%2Cin_review"),
      );
    });
  });
});
