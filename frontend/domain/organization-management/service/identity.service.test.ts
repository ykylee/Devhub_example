import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { OrgMember, OrgNode } from "./identity.service";

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
    vi.doMock("@/shared/api/api-client", () => ({
      apiClient: apiClientMock,
      ApiError: MockApiError,
    }));
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("singleton", () => {
    it("getInstance returns the same instance on repeated calls", async () => {
      const mod = await import("./identity.service");
      expect(mod.identityService).toBe(mod.identityService);
      // Re-import via dynamic should also return same singleton inside a single resetModules scope.
      expect(mod.IdentityService.getInstance()).toBe(mod.identityService);
    });
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

    it("defaults role to Developer when backend role unknown", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: { login: "x", role: "unknown_role" },
      });
      const { identityService } = await import("./identity.service");
      const actor = await identityService.whoAmI();
      expect(actor.role).toBe("Developer");
    });

    it("defaults role to Developer when backend role is missing entirely", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: { login: "no-role" },
      });
      const { identityService } = await import("./identity.service");
      const actor = await identityService.whoAmI();
      expect(actor.role).toBe("Developer");
    });

    it("sets onboarding_required default false when missing", async () => {
      apiClientMock.mockResolvedValue({ status: "ok", data: { login: "a", role: "developer" } });
      const { identityService } = await import("./identity.service");
      const actor = await identityService.whoAmI();
      expect(actor.onboarding_required).toBe(false);
    });

    it("propagates onboarding_completed_at / review_status / primary_unit_id", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: {
          login: "alice",
          role: "developer",
          onboarding_completed_at: "2026-05-28T11:00:00Z",
          review_status: "reviewed",
          primary_unit_id: "unit-7",
        },
      });
      const { identityService } = await import("./identity.service");
      const actor = await identityService.whoAmI();
      expect(actor.onboarding_completed_at).toBe("2026-05-28T11:00:00Z");
      expect(actor.review_status).toBe("reviewed");
      expect(actor.primary_unit_id).toBe("unit-7");
    });

    it("null primary_unit_id is preserved as null", async () => {
      apiClientMock.mockResolvedValue({ status: "ok", data: { login: "x", role: "developer", primary_unit_id: null } });
      const { identityService } = await import("./identity.service");
      const actor = await identityService.whoAmI();
      expect(actor.primary_unit_id).toBeNull();
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
          { user_id: "u1", display_name: "Alice", email: "alice@test.com", role: "team_manager", status: "active" },
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

    it("maps system_admin role → 'System Admin' label", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: [{ user_id: "u1", display_name: "Admin", email: "a@a.com", role: "system_admin", status: "active" }],
      });
      const { identityService } = await import("./identity.service");
      const users = await identityService.getUsers();
      expect(users[0].role).toBe("System Admin");
    });

    it("defaults unknown role to Developer", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: [{ user_id: "u1", display_name: "X", email: "x@x.com", role: "weirdo", status: "active" }],
      });
      const { identityService } = await import("./identity.service");
      const users = await identityService.getUsers();
      expect(users[0].role).toBe("Developer");
    });

    it("slices joined_at to YYYY-MM-DD when string is full ISO", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: [{ user_id: "u1", display_name: "Z", email: "z@z.com", role: "developer", status: "active", joined_at: "2026-05-28T01:23:45Z" }],
      });
      const { identityService } = await import("./identity.service");
      const users = await identityService.getUsers();
      expect(users[0].joined_at).toBe("2026-05-28");
    });

    it("falls back to '' when joined_at is null/undefined", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: [{ user_id: "u1", display_name: "Z", email: "z@z.com", role: "developer", status: "active" }],
      });
      const { identityService } = await import("./identity.service");
      const users = await identityService.getUsers();
      expect(users[0].joined_at).toBe("");
    });

    it("defaults primary_dept_id and current_dept_id to '' when absent", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: [{ user_id: "u1", display_name: "Z", email: "z@z.com", role: "developer", status: "active" }],
      });
      const { identityService } = await import("./identity.service");
      const users = await identityService.getUsers();
      expect(users[0].primary_dept_id).toBe("");
      expect(users[0].current_dept_id).toBe("");
      expect(users[0].is_seconded).toBe(false);
      expect(users[0].appointments).toEqual([]);
      expect(users[0].onboarding_completed_at).toBeNull();
      expect(users[0].review_status).toBeNull();
    });

    it("preserves onboarding_completed_at / review_status when present", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: [{
          user_id: "u1", display_name: "Z", email: "z@z.com", role: "developer", status: "active",
          onboarding_completed_at: "2026-01-01T00:00:00Z", review_status: "pending_review",
        }],
      });
      const { identityService } = await import("./identity.service");
      const users = await identityService.getUsers();
      expect(users[0].onboarding_completed_at).toBe("2026-01-01T00:00:00Z");
      expect(users[0].review_status).toBe("pending_review");
    });
  });

  describe("updateOrgHierarchy", () => {
    it("PUTs nodes + edges payload", async () => {
      apiClientMock.mockResolvedValue({ status: "ok" });
      const { identityService } = await import("./identity.service");
      const nodes: OrgNode[] = [{ id: "root", data: { label: "Root", type: "division" }, position: { x: 0, y: 0 } }];
      const edges = [{ source: "root", target: "team-a" }];

      await identityService.updateOrgHierarchy(nodes, edges);

      expect(apiClientMock).toHaveBeenCalledWith(
        "PUT",
        "/api/v1/organization/hierarchy",
        { nodes, edges },
      );
    });
  });

  describe("lookupHR", () => {
    it("returns payload data on success", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: { email: "hr@test.com", user_id: "u-hr", department: "Eng" },
      });
      const { identityService } = await import("./identity.service");
      const res = await identityService.lookupHR("emp-1");

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/hr/lookup?system_id=emp-1");
      expect(res.email).toBe("hr@test.com");
      expect(res.user_id).toBe("u-hr");
      expect(res.department).toBe("Eng");
    });

    it("url-encodes system_id", async () => {
      apiClientMock.mockResolvedValue({ status: "ok", data: { email: "e", user_id: "u", department: "d" } });
      const { identityService } = await import("./identity.service");
      await identityService.lookupHR("a/b c");
      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/hr/lookup?system_id=a%2Fb%20c");
    });

    it("throws ApiError when data is missing", async () => {
      apiClientMock.mockResolvedValue({ status: "error" });
      const { identityService } = await import("./identity.service");
      await expect(identityService.lookupHR("emp-1")).rejects.toThrow(/hr lookup/);
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
          edges: [{ source_unit_id: "root", target_unit_id: "child" }],
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

    it("adds type:'input' for org-root node only", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: {
          units: [
            { unit_id: "org-root", unit_type: "division", label: "Top", leader_user_id: "u1" },
            { unit_id: "team-a", unit_type: "team", label: "A" },
          ],
          edges: [],
        },
      });
      const { identityService } = await import("./identity.service");
      const h = await identityService.getOrgHierarchy();

      expect(h.nodes[0].type).toBe("input");
      expect(h.nodes[1].type).toBeUndefined();
      expect(h.nodes[0].data.leader_id).toBe("u1");
    });

    it("animates edges originating from org-root only", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: {
          units: [],
          edges: [
            { source_unit_id: "org-root", target_unit_id: "child-1" },
            { source_unit_id: "child-1", target_unit_id: "child-2" },
          ],
        },
      });
      const { identityService } = await import("./identity.service");
      const h = await identityService.getOrgHierarchy();
      expect(h.edges[0].animated).toBe(true);
      expect(h.edges[1].animated).toBe(false);
      expect(h.edges[0].id).toBe("e-org-root-child-1");
    });

    it("defaults position to 0,0 when missing", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: {
          units: [{ unit_id: "t1", unit_type: "team", label: "T1" }],
          edges: [],
        },
      });
      const { identityService } = await import("./identity.service");
      const h = await identityService.getOrgHierarchy();
      expect(h.nodes[0].position).toEqual({ x: 0, y: 0 });
      expect(h.nodes[0].data.leader_id).toBeUndefined();
    });

    it("returns empty nodes/edges when response.data is missing", async () => {
      apiClientMock.mockResolvedValue({ status: "ok" });
      const { identityService } = await import("./identity.service");
      const h = await identityService.getOrgHierarchy();
      expect(h.nodes).toEqual([]);
      expect(h.edges).toEqual([]);
    });

    it("returns empty arrays when data has no units / edges field", async () => {
      apiClientMock.mockResolvedValue({ status: "ok", data: {} });
      const { identityService } = await import("./identity.service");
      const h = await identityService.getOrgHierarchy();
      expect(h.nodes).toEqual([]);
      expect(h.edges).toEqual([]);
    });
  });

  describe("getUnitMembers", () => {
    it("url-encodes unit id and maps backend users", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: [{ user_id: "u1", display_name: "Z", email: "z@z.com", role: "developer", status: "active" }],
      });
      const { identityService } = await import("./identity.service");
      const members = await identityService.getUnitMembers("unit/with space");
      expect(apiClientMock).toHaveBeenCalledWith(
        "GET",
        "/api/v1/organization/units/unit%2Fwith%20space/members",
      );
      expect(members).toHaveLength(1);
    });

    it("returns [] when no data", async () => {
      apiClientMock.mockResolvedValue({ status: "ok" });
      const { identityService } = await import("./identity.service");
      const members = await identityService.getUnitMembers("u-1");
      expect(members).toEqual([]);
    });
  });

  describe("createUser", () => {
    it("maps UI role to backend role + defaults optional fields", async () => {
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

      const [, , body] = apiClientMock.mock.calls[0];
      expect(body).toMatchObject({
        user_id: "new-u",
        role: "developer",
        primary_unit_id: "",
        current_unit_id: "",
        is_seconded: false,
        joined_at: "",
      });
    });

    it("forwards provided dept ids / joined_at / is_seconded", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: { user_id: "u1", display_name: "X", email: "x@x.com", role: "team_manager", status: "active" },
      });
      const { identityService } = await import("./identity.service");
      await identityService.createUser({
        user_id: "u1",
        email: "x@x.com",
        display_name: "X",
        role: "Manager",
        status: "active",
        type: "human",
        primary_dept_id: "dept-a",
        current_dept_id: "dept-b",
        is_seconded: true,
        joined_at: "2026-05-28",
      });

      const [, , body] = apiClientMock.mock.calls[0];
      expect(body).toMatchObject({
        primary_unit_id: "dept-a",
        current_unit_id: "dept-b",
        is_seconded: true,
        joined_at: "2026-05-28",
        role: "team_manager",
      });
    });

    it("throws ApiError when no data returned", async () => {
      apiClientMock.mockResolvedValue({ status: "error" });
      const { identityService } = await import("./identity.service");
      await expect(
        identityService.createUser({
          user_id: "u1", email: "e", display_name: "d", role: "Developer", status: "active", type: "human",
        }),
      ).rejects.toThrow(/user payload/);
    });
  });

  describe("updateUser", () => {
    it("includes only specified payload fields in PATCH body", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: { user_id: "u1", display_name: "Updated", email: "e@e.com", role: "developer", status: "active" },
      });
      const { identityService } = await import("./identity.service");

      await identityService.updateUser("u1", { display_name: "Updated", role: "Manager", status: "deactivated" });

      const [method, path, body] = apiClientMock.mock.calls[0];
      expect(method).toBe("PATCH");
      expect(path).toBe("/api/v1/users/u1");
      expect(body).toEqual({ display_name: "Updated", role: "team_manager", status: "deactivated" });
    });

    it("forwards all optional fields when provided", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: { user_id: "u1", display_name: "x", email: "x", role: "developer", status: "active" },
      });
      const { identityService } = await import("./identity.service");
      await identityService.updateUser("u1", {
        email: "e",
        display_name: "d",
        role: "Developer",
        status: "active",
        primary_dept_id: "dept-a",
        current_dept_id: "dept-b",
        is_seconded: false,
        joined_at: "2026-01-01",
      });
      const [, , body] = apiClientMock.mock.calls[0];
      expect(body).toEqual({
        email: "e",
        display_name: "d",
        role: "developer",
        status: "active",
        primary_unit_id: "dept-a",
        current_unit_id: "dept-b",
        is_seconded: false,
        joined_at: "2026-01-01",
      });
    });

    it("sends empty body when no fields are provided", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: { user_id: "u1", display_name: "x", email: "x", role: "developer", status: "active" },
      });
      const { identityService } = await import("./identity.service");
      await identityService.updateUser("u1", {});

      const [, , body] = apiClientMock.mock.calls[0];
      expect(body).toEqual({});
    });

    it("url-encodes user id (slashes / spaces)", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: { user_id: "u/1", display_name: "x", email: "x", role: "developer", status: "active" },
      });
      const { identityService } = await import("./identity.service");
      await identityService.updateUser("u/1", { display_name: "n" });
      const [, path] = apiClientMock.mock.calls[0];
      expect(path).toBe("/api/v1/users/u%2F1");
    });

    it("throws when data missing", async () => {
      apiClientMock.mockResolvedValue({ status: "error" });
      const { identityService } = await import("./identity.service");
      await expect(identityService.updateUser("u1", { display_name: "x" })).rejects.toThrow(/user payload/);
    });
  });

  describe("deleteUser", () => {
    it("sends DELETE request and url-encodes id", async () => {
      apiClientMock.mockResolvedValue({ status: "ok" });
      const { identityService } = await import("./identity.service");
      await identityService.deleteUser("user/1");

      expect(apiClientMock).toHaveBeenCalledWith("DELETE", "/api/v1/users/user%2F1");
    });
  });

  describe("getUnit", () => {
    it("GETs encoded unit and maps response", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: { unit_id: "u-1", unit_type: "team", label: "T", parent_unit_id: "root", leader_user_id: "u-leader", position_x: 5, position_y: 10, direct_count: 3, total_count: 7 },
      });
      const { identityService } = await import("./identity.service");
      const unit = await identityService.getUnit("u 1");

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/organization/units/u%201");
      expect(unit.unit_id).toBe("u-1");
      expect(unit.parent_unit_id).toBe("root");
      expect(unit.leader_user_id).toBe("u-leader");
      expect(unit.position_x).toBe(5);
      expect(unit.position_y).toBe(10);
      expect(unit.direct_count).toBe(3);
      expect(unit.total_count).toBe(7);
    });

    it("defaults parent_unit_id / leader_user_id / positions when missing", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: { unit_id: "u-1", unit_type: "team", label: "T" },
      });
      const { identityService } = await import("./identity.service");
      const unit = await identityService.getUnit("u-1");
      expect(unit.parent_unit_id).toBe("");
      expect(unit.leader_user_id).toBe("");
      expect(unit.position_x).toBe(0);
      expect(unit.position_y).toBe(0);
    });

    it("throws when data missing", async () => {
      apiClientMock.mockResolvedValue({ status: "error" });
      const { identityService } = await import("./identity.service");
      await expect(identityService.getUnit("u-1")).rejects.toThrow(/unit payload/);
    });
  });

  describe("createUnit", () => {
    it("POSTs payload with optional defaults", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: { unit_id: "u-2", unit_type: "team", label: "T2" },
      });
      const { identityService } = await import("./identity.service");
      const unit = await identityService.createUnit({
        unit_id: "u-2",
        unit_type: "team",
        label: "T2",
      });

      const [method, path, body] = apiClientMock.mock.calls[0];
      expect(method).toBe("POST");
      expect(path).toBe("/api/v1/organization/units");
      expect(body).toEqual({
        unit_id: "u-2",
        parent_unit_id: "",
        unit_type: "team",
        label: "T2",
        leader_user_id: "",
        position_x: 0,
        position_y: 0,
      });
      expect(unit.unit_id).toBe("u-2");
    });

    it("forwards provided optional fields", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: { unit_id: "u-3", unit_type: "group", label: "G" },
      });
      const { identityService } = await import("./identity.service");
      await identityService.createUnit({
        unit_id: "u-3",
        unit_type: "group",
        label: "G",
        parent_unit_id: "p",
        leader_user_id: "leader",
        position_x: 100,
        position_y: 200,
      });
      const [, , body] = apiClientMock.mock.calls[0];
      expect(body).toMatchObject({
        parent_unit_id: "p",
        leader_user_id: "leader",
        position_x: 100,
        position_y: 200,
      });
    });

    it("throws when data missing", async () => {
      apiClientMock.mockResolvedValue({ status: "error" });
      const { identityService } = await import("./identity.service");
      await expect(
        identityService.createUnit({ unit_id: "x", unit_type: "team", label: "x" }),
      ).rejects.toThrow(/unit payload/);
    });
  });

  describe("updateUnit", () => {
    it("includes only specified fields in PATCH body", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: { unit_id: "u-1", unit_type: "team", label: "Renamed" },
      });
      const { identityService } = await import("./identity.service");
      await identityService.updateUnit("u-1", { label: "Renamed", position_x: 50 });

      const [, path, body] = apiClientMock.mock.calls[0];
      expect(path).toBe("/api/v1/organization/units/u-1");
      expect(body).toEqual({ label: "Renamed", position_x: 50 });
    });

    it("forwards all optional fields when provided", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: { unit_id: "u-1", unit_type: "team", label: "T" },
      });
      const { identityService } = await import("./identity.service");
      await identityService.updateUnit("u-1", {
        parent_unit_id: "p",
        unit_type: "team",
        label: "T",
        leader_user_id: "leader",
        position_x: 1,
        position_y: 2,
      });
      const [, , body] = apiClientMock.mock.calls[0];
      expect(body).toEqual({
        parent_unit_id: "p",
        unit_type: "team",
        label: "T",
        leader_user_id: "leader",
        position_x: 1,
        position_y: 2,
      });
    });

    it("url-encodes unit id and sends empty body when no fields", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: { unit_id: "x/1", unit_type: "team", label: "T" },
      });
      const { identityService } = await import("./identity.service");
      await identityService.updateUnit("x/1", {});

      const [, path, body] = apiClientMock.mock.calls[0];
      expect(path).toBe("/api/v1/organization/units/x%2F1");
      expect(body).toEqual({});
    });

    it("throws when data missing", async () => {
      apiClientMock.mockResolvedValue({ status: "error" });
      const { identityService } = await import("./identity.service");
      await expect(identityService.updateUnit("u-1", { label: "x" })).rejects.toThrow(/unit payload/);
    });
  });

  describe("deleteUnit", () => {
    it("DELETEs encoded path", async () => {
      apiClientMock.mockResolvedValue({ status: "ok" });
      const { identityService } = await import("./identity.service");
      await identityService.deleteUnit("u 1");
      expect(apiClientMock).toHaveBeenCalledWith("DELETE", "/api/v1/organization/units/u%201");
    });
  });

  describe("replaceUnitMembers", () => {
    it("PUTs the user_ids list and maps response", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        data: [
          { user_id: "u1", display_name: "A", email: "a@a.com", role: "developer", status: "active" },
        ],
      });
      const { identityService } = await import("./identity.service");
      const members = await identityService.replaceUnitMembers("unit-1", ["u1", "u2"]);

      expect(apiClientMock).toHaveBeenCalledWith(
        "PUT",
        "/api/v1/organization/units/unit-1/members",
        { user_ids: ["u1", "u2"] },
      );
      expect(members).toHaveLength(1);
    });

    it("returns [] when no data", async () => {
      apiClientMock.mockResolvedValue({ status: "ok" });
      const { identityService } = await import("./identity.service");
      const members = await identityService.replaceUnitMembers("unit-1", []);
      expect(members).toEqual([]);
    });
  });

  describe("calculatePrimaryDept", () => {
    function member(
      overrides: Partial<OrgMember> & { appointments: OrgMember["appointments"] },
    ): OrgMember {
      return {
        id: "u1",
        name: "Test",
        email: "t@t.com",
        role: "Developer",
        status: "active",
        primary_dept_id: "",
        current_dept_id: "current-x",
        is_seconded: false,
        joined_at: "2026-01-01",
        ...overrides,
      };
    }

    it("returns current_dept_id when zero appointments", async () => {
      const { identityService } = await import("./identity.service");
      const result = identityService.calculatePrimaryDept(member({ appointments: [] }), []);
      expect(result).toBe("current-x");
    });

    it("returns current_dept_id when only one appointment (regardless of role)", async () => {
      const { identityService } = await import("./identity.service");
      const m = member({ appointments: [{ dept_id: "team-1", role: "member" }] });
      const result = identityService.calculatePrimaryDept(m, []);
      expect(result).toBe("current-x");
    });

    it("returns current_dept_id when 2+ appointments but no leader role", async () => {
      const { identityService } = await import("./identity.service");
      const m = member({
        appointments: [
          { dept_id: "team-1", role: "member" },
          { dept_id: "team-2", role: "member" },
        ],
      });
      expect(identityService.calculatePrimaryDept(m, [])).toBe("current-x");
    });

    it("returns current_dept_id when leader role exists but node missing in nodes list", async () => {
      const { identityService } = await import("./identity.service");
      const m = member({
        appointments: [
          { dept_id: "team-1", role: "leader" },
          { dept_id: "team-2", role: "leader" },
        ],
      });
      const result = identityService.calculatePrimaryDept(m, []);
      expect(result).toBe("current-x");
    });

    it("prefers leader appointment in higher-priority dept type (division > team)", async () => {
      const { identityService } = await import("./identity.service");
      const m = member({
        appointments: [
          { dept_id: "div-1", role: "leader" },
          { dept_id: "team-1", role: "leader" },
        ],
      });
      const nodes = [
        { id: "div-1", data: { label: "Division", type: "division" }, position: { x: 0, y: 0 } },
        { id: "team-1", data: { label: "Team", type: "team" }, position: { x: 0, y: 0 } },
      ];
      const result = identityService.calculatePrimaryDept(m, nodes);
      expect(result).toBe("div-1");
    });

    it("returns first when priorities equal (stable order by sort)", async () => {
      const { identityService } = await import("./identity.service");
      const m = member({
        appointments: [
          { dept_id: "team-a", role: "leader" },
          { dept_id: "team-b", role: "leader" },
        ],
      });
      const nodes = [
        { id: "team-a", data: { label: "A", type: "team" }, position: { x: 0, y: 0 } },
        { id: "team-b", data: { label: "B", type: "team" }, position: { x: 0, y: 0 } },
      ];
      const result = identityService.calculatePrimaryDept(m, nodes);
      // both priority 3 → sort returns 0 (stable) → first leader appointment wins.
      expect(result).toBe("team-a");
    });

    it("handles unknown dept type (priority = 0) as lowest", async () => {
      const { identityService } = await import("./identity.service");
      const m = member({
        appointments: [
          { dept_id: "weird", role: "leader" },
          { dept_id: "team-1", role: "leader" },
        ],
      });
      const nodes = [
        { id: "weird", data: { label: "X", type: "unknown" }, position: { x: 0, y: 0 } },
        { id: "team-1", data: { label: "T", type: "team" }, position: { x: 0, y: 0 } },
      ];
      expect(identityService.calculatePrimaryDept(m, nodes)).toBe("team-1");
    });

    it("company > division priority", async () => {
      const { identityService } = await import("./identity.service");
      const m = member({
        appointments: [
          { dept_id: "co-1", role: "leader" },
          { dept_id: "div-1", role: "leader" },
        ],
      });
      const nodes = [
        { id: "co-1", data: { label: "C", type: "company" }, position: { x: 0, y: 0 } },
        { id: "div-1", data: { label: "D", type: "division" }, position: { x: 0, y: 0 } },
      ];
      expect(identityService.calculatePrimaryDept(m, nodes)).toBe("co-1");
    });
  });

});
