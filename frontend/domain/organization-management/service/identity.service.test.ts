import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("IdentityService", () => {
  const apiClientMock = vi.fn();

  // Mock ApiError for tests
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

  describe("whoAmI", () => {
    it("returns resolved actor with UI role mapping", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: {
          login: "yklee",
          user_id: "u1",
          role: "system_admin",
          display_name: "YK Lee",
          email: "yklee@example.com",
          onboarding_required: false,
        },
      });

      const { identityService } = await import("./identity.service");
      const actor = await identityService.whoAmI();

      expect(actor.login).toBe("yklee");
      expect(actor.role).toBe("System Admin");
      expect(actor.display_name).toBe("YK Lee");
      expect(actor.onboarding_required).toBe(false);
    });

    it("throws ApiError when data is missing", async () => {
      apiClientMock.mockResolvedValue({ status: "error" });

      const { identityService } = await import("./identity.service");
      await expect(identityService.whoAmI()).rejects.toThrow();
    });
  });

  describe("getUsers", () => {
    it("returns mapped OrgMember array", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: [
          { user_id: "u1", display_name: "Alice", email: "alice@test.com", role: "manager", status: "active" },
          { user_id: "u2", display_name: "Bob", email: "bob@test.com", role: "developer", status: "active", appointments: [{ unit_id: "dept-1", appointment_role: "member" }] },
        ],
      });

      const { identityService } = await import("./identity.service");
      const users = await identityService.getUsers();

      expect(users).toHaveLength(2);
      expect(users[0].role).toBe("Manager");
      expect(users[1].role).toBe("Developer");
      expect(users[1].appointments).toHaveLength(1);
      expect(users[1].appointments[0].dept_id).toBe("dept-1");
    });

    it("returns empty array when no data", async () => {
      apiClientMock.mockResolvedValue({ status: "ok" });

      const { identityService } = await import("./identity.service");
      const users = await identityService.getUsers();

      expect(users).toEqual([]);
    });
  });

  describe("createUser", () => {
    it("maps UI role to backend role in request body", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: { user_id: "new-u", display_name: "New User", email: "new@test.com", role: "developer", status: "active" },
      });

      const { identityService } = await import("./identity.service");
      await identityService.createUser({
        user_id: "new-u",
        email: "new@test.com",
        display_name: "New User",
        role: "Developer",
        status: "active",
        type: "human",
      });

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        expect.stringContaining("/api/v1/users"),
        expect.objectContaining({ role: "developer" }),
      );
    });
  });

  describe("getOrgHierarchy", () => {
    it("transforms backend units and edges to OrgNode/OrgEdge format", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: {
          units: [
            { unit_id: "root", unit_type: "division", label: "Root", position_x: 100, position_y: 200 },
            { unit_id: "child", parent_unit_id: "root", unit_type: "team", label: "Child" },
          ],
          edges: [
            { source_unit_id: "root", target_unit_id: "child" },
          ],
        },
      });

      const { identityService } = await import("./identity.service");
      const hierarchy = await identityService.getOrgHierarchy();

      expect(hierarchy.nodes).toHaveLength(2);
      expect(hierarchy.nodes[0].data.label).toBe("Root");
      expect(hierarchy.nodes[0].position).toEqual({ x: 100, y: 200 });
      expect(hierarchy.edges).toHaveLength(1);
      expect(hierarchy.edges[0].source).toBe("root");
    });
  });

  describe("calculatePrimaryDept", () => {
    it("returns current_dept_id when only one appointment", async () => {
      const { identityService } = await import("./identity.service");
      const member = {
        id: "u1", name: "Test", email: "t@t.com", role: "Developer" as const,
        status: "active" as const, primary_dept_id: "dept-a", current_dept_id: "dept-a",
        is_seconded: false, appointments: [{ dept_id: "team-1", role: "member" as const }],
        joined_at: "2026-01-01",
      };

      const result = identityService.calculatePrimaryDept(member, []);
      expect(result).toBe("dept-a");
    });

    it("prefers leader appointment in higher-priority dept type", async () => {
      const { identityService } = await import("./identity.service");
      const member = {
        id: "u1", name: "Test", email: "t@t.com", role: "Developer" as const,
        status: "active" as const, primary_dept_id: "", current_dept_id: "team-x",
        is_seconded: false,
        appointments: [
          { dept_id: "div-1", role: "leader" as const },
          { dept_id: "team-1", role: "leader" as const },
        ],
        joined_at: "2026-01-01",
      };
      const nodes = [
        { id: "div-1", data: { label: "Division", type: "division" }, position: { x: 0, y: 0 } },
        { id: "team-1", data: { label: "Team", type: "team" }, position: { x: 0, y: 0 } },
      ];

      const result = identityService.calculatePrimaryDept(member, nodes);
      // Division (priority 4) > Team (priority 3)
      expect(result).toBe("div-1");
    });
  });

  describe("deleteUser", () => {
    it("sends DELETE request", async () => {
      apiClientMock.mockResolvedValue({ status: "ok" });

      const { identityService } = await import("./identity.service");
      await identityService.deleteUser("user-1");

      expect(apiClientMock).toHaveBeenCalledWith("DELETE", expect.stringContaining("/api/v1/users/user-1"));
    });
  });
});
