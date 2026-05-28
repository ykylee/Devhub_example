import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("RbacService", () => {
  const apiClientMock = vi.fn();

  // Mock ApiError class matching the real one in api-client.ts
  class MockApiError extends Error {
    status: number;
    payload: unknown;
    constructor(status: number, payload: unknown, message: string) {
      super(message);
      this.name = "ApiError";
      this.status = status;
      this.payload = payload;
    }
  }

  beforeEach(() => {
    vi.resetModules();
    apiClientMock.mockReset();
    vi.doMock("@/lib/services/api-client", () => ({
      apiClient: apiClientMock,
      ApiError: MockApiError,
    }));
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("listPolicies", () => {
    it("returns roles and meta on success", async () => {
      const mockResponse = {
        status: "ok",
        data: [
          { id: "admin", name: "Admin", description: "Full access", permissions: ["*:*"] },
          { id: "developer", name: "Developer", description: "Dev scope", permissions: ["app:read", "app:write"] },
        ],
        meta: { total: 2, page: 1, page_size: 50 },
      };
      apiClientMock.mockResolvedValue(mockResponse);

      const { rbacService } = await import("./rbac.service");
      const result = await rbacService.listPolicies();

      expect(result.roles).toHaveLength(2);
      expect(result.roles[0].id).toBe("admin");
      expect(result.meta.total).toBe(2);
      expect(apiClientMock).toHaveBeenCalledWith("GET", expect.stringContaining("/api/v1/rbac/policies"));
    });

    it("throws RbacError on API failure", async () => {
      apiClientMock.mockRejectedValue(new MockApiError(500, { code: "internal_error" }, "Server error"));

      const { rbacService } = await import("./rbac.service");
      await expect(rbacService.listPolicies()).rejects.toMatchObject({
        status: 500,
        code: "internal_error",
      });
    });
  });

  describe("createPolicy", () => {
    it("sends POST and returns created role", async () => {
      const newRole = {
        id: "custom-reviewer",
        name: "Reviewer",
        description: "Can review PRs",
        permissions: ["pr:read", "pr:approve"],
      };
      apiClientMock.mockResolvedValue({
        status: "created",
        data: { ...newRole, is_system: false },
      });

      const { rbacService } = await import("./rbac.service");
      const result = await rbacService.createPolicy(newRole);

      expect(result.id).toBe("custom-reviewer");
      expect(result.is_system).toBe(false);
      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        expect.stringContaining("/api/v1/rbac/policies"),
        newRole,
      );
    });

    it("throws RbacError on conflict", async () => {
      apiClientMock.mockRejectedValue(
        new MockApiError(409, { code: "role_exists" }, "Role already exists"),
      );

      const { rbacService } = await import("./rbac.service");
      await expect(rbacService.createPolicy({ id: "dup", name: "Dup", description: "", permissions: [] })).rejects.toMatchObject({
        status: 409,
        code: "role_exists",
      });
    });
  });

  describe("updatePolicies", () => {
    it("sends PUT with roles array and returns updated result", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: [
          { id: "developer", name: "Developer", permissions: ["app:read", "app:write", "app:delete"] },
        ],
        meta: { total: 1, page: 1, page_size: 50 },
      });

      const { rbacService } = await import("./rbac.service");
      const result = await rbacService.updatePolicies([
        { id: "developer", permissions: ["app:read", "app:write", "app:delete"] },
      ]);

      expect(result.roles).toHaveLength(1);
      expect(result.roles[0].permissions).toContain("app:delete");
      expect(apiClientMock).toHaveBeenCalledWith(
        "PUT",
        expect.stringContaining("/api/v1/rbac/policies"),
        { roles: [{ id: "developer", permissions: ["app:read", "app:write", "app:delete"] }] },
      );
    });
  });

  describe("deletePolicy", () => {
    it("sends DELETE and resolves void on status 'deleted'", async () => {
      apiClientMock.mockResolvedValue({ status: "deleted", data: { role_id: "custom-foo" } });

      const { rbacService } = await import("./rbac.service");
      await expect(rbacService.deletePolicy("custom-foo")).resolves.toBeUndefined();

      expect(apiClientMock).toHaveBeenCalledWith("DELETE", expect.stringContaining("/api/v1/rbac/policies/custom-foo"));
    });

    it("throws on unexpected status", async () => {
      apiClientMock.mockResolvedValue({ status: "conflict", data: { role_id: "custom-foo" } });

      const { rbacService } = await import("./rbac.service");
      await expect(rbacService.deletePolicy("custom-foo")).rejects.toThrow("deletePolicy unexpected status: conflict");
    });

    it("throws RbacError on API error (system role undeletable)", async () => {
      apiClientMock.mockRejectedValue(
        new MockApiError(422, { code: "system_role_not_deletable" }, "System roles cannot be deleted"),
      );

      const { rbacService } = await import("./rbac.service");
      await expect(rbacService.deletePolicy("admin")).rejects.toMatchObject({
        status: 422,
        code: "system_role_not_deletable",
      });
    });
  });
});
