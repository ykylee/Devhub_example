import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { PermissionState } from "@/domain/rbac-permissions/view/PermissionMatrix";
import type { Role, RbacPolicyMeta } from "@/domain/rbac-permissions/schema/rbac.types";

describe("RbacService", () => {
  const apiClientMock = vi.fn();

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

  const sampleMeta: RbacPolicyMeta = {
    policy_version: "2026.05.28-1",
    source: "db",
    editable: true,
    system_roles: ["developer", "manager", "system_admin", "pmo_manager"],
  };

  const adminPermissions: PermissionState = {
    infrastructure: { view: true, create: true, edit: true, delete: true },
    audit: { view: true, create: false, edit: false, delete: false },
  };

  const devPermissions: PermissionState = {
    applications: { view: true, edit: true, delete: true },
  };

  beforeEach(() => {
    vi.resetModules();
    apiClientMock.mockReset();
    vi.doMock("@/shared/api/api-client", () => ({
      apiClient: apiClientMock,
      ApiError: MockApiError,
    }));
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("listPolicies", () => {
    it("returns roles and meta on success", async () => {
      const mockRoles: Role[] = [
        { id: "system_admin", name: "Admin", description: "Full", system: true, permissions: adminPermissions },
        { id: "developer", name: "Dev", description: "Dev scope", system: true, permissions: devPermissions },
      ];
      apiClientMock.mockResolvedValue({ status: "ok", data: mockRoles, meta: sampleMeta });

      const { rbacService } = await import("./rbac.service");
      const result = await rbacService.listPolicies();

      expect(result.roles).toHaveLength(2);
      expect(result.roles[0].id).toBe("system_admin");
      expect(result.roles[0].permissions.infrastructure?.view).toBe(true);
      expect(result.meta.policy_version).toBe("2026.05.28-1");
      expect(result.meta.editable).toBe(true);
      expect(apiClientMock).toHaveBeenCalledWith("GET", expect.stringContaining("/api/v1/rbac/policies"));
    });

    it("propagates RbacError with code on 500 ApiError", async () => {
      apiClientMock.mockRejectedValue(new MockApiError(500, { code: "internal_error" }, "Server error"));

      const { rbacService } = await import("./rbac.service");
      await expect(rbacService.listPolicies()).rejects.toMatchObject({
        status: 500,
        code: "internal_error",
      });
    });

    it("falls back to code 'unknown' when ApiError payload has no code", async () => {
      apiClientMock.mockRejectedValue(new MockApiError(503, { message: "down" }, "service unavailable"));

      const { rbacService } = await import("./rbac.service");
      await expect(rbacService.listPolicies()).rejects.toMatchObject({
        status: 503,
        code: "unknown",
      });
    });

    it("wraps non-ApiError as RbacError with status 500 and prefix", async () => {
      apiClientMock.mockRejectedValue(new TypeError("network down"));

      const { rbacService } = await import("./rbac.service");
      await expect(rbacService.listPolicies()).rejects.toMatchObject({
        status: 500,
        code: "unknown",
        message: expect.stringContaining("listPolicies failed: network down"),
      });
    });
  });

  describe("createPolicy", () => {
    it("sends POST and returns created role", async () => {
      const reviewerPerms: PermissionState = {
        applications: { view: true, edit: true },
      };
      const newRole: Pick<Role, "id" | "name" | "description" | "permissions"> = {
        id: "custom-reviewer",
        name: "Reviewer",
        description: "Can review PRs",
        permissions: reviewerPerms,
      };
      const createdRole: Role = { ...newRole, system: false };
      apiClientMock.mockResolvedValue({ status: "created", data: createdRole });

      const { rbacService } = await import("./rbac.service");
      const result = await rbacService.createPolicy(newRole);

      expect(result.id).toBe("custom-reviewer");
      expect(result.system).toBe(false);
      expect(result.permissions.applications?.edit).toBe(true);
      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        expect.stringContaining("/api/v1/rbac/policies"),
        newRole,
      );
    });

    it("throws RbacError with code 'role_exists' on 409 conflict", async () => {
      apiClientMock.mockRejectedValue(
        new MockApiError(409, { code: "role_exists" }, "Role already exists"),
      );

      const { rbacService } = await import("./rbac.service");
      const dup: Pick<Role, "id" | "name" | "description" | "permissions"> = {
        id: "dup", name: "Dup", description: "", permissions: {},
      };
      await expect(rbacService.createPolicy(dup)).rejects.toMatchObject({
        status: 409,
        code: "role_exists",
      });
    });

    it("rejects empty permissions object as createPolicy payload (caller responsibility)", async () => {
      // empty permissions is accepted by service (no schema enforcement); test
      // verifies service does not transform/strip the payload.
      const emptyRole: Pick<Role, "id" | "name" | "description" | "permissions"> = {
        id: "custom-empty", name: "Empty", description: "", permissions: {},
      };
      apiClientMock.mockResolvedValue({ status: "created", data: { ...emptyRole, system: false } });

      const { rbacService } = await import("./rbac.service");
      const result = await rbacService.createPolicy(emptyRole);

      expect(result.id).toBe("custom-empty");
      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        expect.any(String),
        emptyRole,
      );
    });
  });

  describe("updatePolicies", () => {
    it("sends PUT with roles array and returns updated result", async () => {
      const updatedDev: Role = {
        id: "developer", name: "Developer", description: "", system: true, permissions: devPermissions,
      };
      apiClientMock.mockResolvedValue({ status: "ok", data: [updatedDev], meta: sampleMeta });

      const { rbacService } = await import("./rbac.service");
      const result = await rbacService.updatePolicies([
        { id: "developer", permissions: devPermissions },
      ]);

      expect(result.roles).toHaveLength(1);
      expect(result.roles[0].permissions.applications?.delete).toBe(true);
      expect(result.meta.policy_version).toBe("2026.05.28-1");
      expect(apiClientMock).toHaveBeenCalledWith(
        "PUT",
        expect.stringContaining("/api/v1/rbac/policies"),
        { roles: [{ id: "developer", permissions: devPermissions }] },
      );
    });

    it("propagates 422 system_role_immutable error", async () => {
      apiClientMock.mockRejectedValue(
        new MockApiError(422, { code: "system_role_immutable" }, "Cannot rename system role"),
      );

      const { rbacService } = await import("./rbac.service");
      await expect(
        rbacService.updatePolicies([{ id: "system_admin", name: "Renamed" }]),
      ).rejects.toMatchObject({ status: 422, code: "system_role_immutable" });
    });
  });

  describe("deletePolicy", () => {
    it("sends DELETE and resolves void on status 'deleted'", async () => {
      apiClientMock.mockResolvedValue({ status: "deleted", data: { role_id: "custom-foo" } });

      const { rbacService } = await import("./rbac.service");
      await expect(rbacService.deletePolicy("custom-foo")).resolves.toBeUndefined();

      expect(apiClientMock).toHaveBeenCalledWith(
        "DELETE",
        expect.stringContaining("/api/v1/rbac/policies/custom-foo"),
      );
    });

    it("URL-encodes role_id with special characters", async () => {
      apiClientMock.mockResolvedValue({ status: "deleted", data: { role_id: "custom/odd id" } });

      const { rbacService } = await import("./rbac.service");
      await rbacService.deletePolicy("custom/odd id");

      expect(apiClientMock).toHaveBeenCalledWith(
        "DELETE",
        expect.stringContaining("/api/v1/rbac/policies/custom%2Fodd%20id"),
      );
    });

    it("throws on unexpected response status", async () => {
      apiClientMock.mockResolvedValue({ status: "conflict", data: { role_id: "custom-foo" } });

      const { rbacService } = await import("./rbac.service");
      await expect(rbacService.deletePolicy("custom-foo")).rejects.toThrow(
        "deletePolicy unexpected status: conflict",
      );
    });

    it("throws RbacError with code on 422 system_role_not_deletable", async () => {
      apiClientMock.mockRejectedValue(
        new MockApiError(422, { code: "system_role_not_deletable" }, "System roles cannot be deleted"),
      );

      const { rbacService } = await import("./rbac.service");
      await expect(rbacService.deletePolicy("admin")).rejects.toMatchObject({
        status: 422,
        code: "system_role_not_deletable",
      });
    });

    it("throws RbacError with code on 422 role_in_use", async () => {
      apiClientMock.mockRejectedValue(
        new MockApiError(422, { code: "role_in_use" }, "Role still assigned"),
      );

      const { rbacService } = await import("./rbac.service");
      await expect(rbacService.deletePolicy("custom-foo")).rejects.toMatchObject({
        status: 422,
        code: "role_in_use",
      });
    });
  });
});
