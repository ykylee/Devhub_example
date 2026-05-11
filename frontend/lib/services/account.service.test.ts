import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { accountService, SettingsFlowError } from "./account.service";

// account.service drives the Kratos public settings flow (DEC-2=B, PR-L3) +
// surfaces redirect/validation/auth errors via the SettingsFlowError code.
// These tests pin the branches that PR-L3 added and the PR-L3 fix-up
// (Codex P1 — redirect:'manual' so a 303 to /self-service/login doesn't
// become a false success).

describe("accountService.updateMyPassword", () => {
  const flowID = "flow-1";
  const flowAction = "http://localhost:4433/self-service/settings?flow=flow-1";
  let fetchSpy: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchSpy = vi.fn();
    vi.stubGlobal("fetch", fetchSpy);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  function flowResponse(state?: string) {
    return {
      ok: true,
      status: 200,
      type: "basic",
      json: async () => ({
        id: flowID,
        state,
        ui: {
          action: flowAction,
          method: "POST",
          nodes: [{ attributes: { name: "csrf_token", value: "csrf-1" } }],
        },
      }),
    };
  }

  it("happy path: GET settings/browser → POST settings → success state resolves", async () => {
    fetchSpy
      .mockResolvedValueOnce(flowResponse()) // GET /self-service/settings/browser
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        type: "basic",
        json: async () => ({ id: flowID, state: "success", ui: { nodes: [] } }),
      });

    await expect(accountService.updateMyPassword("any-current", "NewLongerPassword123!"))
      .resolves.toBeUndefined();
    expect(fetchSpy).toHaveBeenCalledTimes(2);
  });

  it("REAUTH_REQUIRED when GET settings/browser returns 401 with redirect_browser_to", async () => {
    fetchSpy.mockResolvedValueOnce({
      ok: false,
      status: 401,
      type: "basic",
      json: async () => ({ redirect_browser_to: "http://localhost:4433/self-service/login?refresh=true" }),
    });

    try {
      await accountService.updateMyPassword("c", "n");
      throw new Error("should have thrown");
    } catch (err) {
      expect(err).toBeInstanceOf(SettingsFlowError);
      const sfe = err as SettingsFlowError;
      expect(sfe.code).toBe("REAUTH_REQUIRED");
      expect(sfe.redirectURL).toBe("http://localhost:4433/self-service/login?refresh=true");
    }
  });

  it("VALIDATION when POST settings returns 400 with ui.messages", async () => {
    fetchSpy
      .mockResolvedValueOnce(flowResponse())
      .mockResolvedValueOnce({
        ok: false,
        status: 400,
        type: "basic",
        json: async () => ({
          ui: {
            messages: [{ text: "password is too short" }],
            nodes: [],
          },
        }),
      });

    try {
      await accountService.updateMyPassword("c", "short");
      throw new Error("should have thrown");
    } catch (err) {
      expect(err).toBeInstanceOf(SettingsFlowError);
      const sfe = err as SettingsFlowError;
      expect(sfe.code).toBe("VALIDATION");
      expect(sfe.message).toContain("password is too short");
    }
  });

  it("REAUTH_REQUIRED when POST settings is 3xx (opaqueredirect path, Codex P1 fix-up)", async () => {
    fetchSpy
      .mockResolvedValueOnce(flowResponse())
      .mockResolvedValueOnce({
        ok: false,
        status: 0,
        type: "opaqueredirect",
        json: async () => ({}),
      });

    try {
      await accountService.updateMyPassword("c", "n");
      throw new Error("should have thrown");
    } catch (err) {
      expect(err).toBeInstanceOf(SettingsFlowError);
      expect((err as SettingsFlowError).code).toBe("REAUTH_REQUIRED");
    }
  });

  it("SUBMIT_FAILED when POST settings returns 200 without state==='success' (Codex P1 fix-up)", async () => {
    fetchSpy
      .mockResolvedValueOnce(flowResponse())
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        type: "basic",
        json: async () => ({ id: flowID, state: "show_form", ui: { nodes: [] } }),
      });

    try {
      await accountService.updateMyPassword("c", "n");
      throw new Error("should have thrown");
    } catch (err) {
      expect(err).toBeInstanceOf(SettingsFlowError);
      expect((err as SettingsFlowError).code).toBe("SUBMIT_FAILED");
    }
  });

  it("REAUTH_REQUIRED when POST settings returns 403 session_refresh_required", async () => {
    fetchSpy
      .mockResolvedValueOnce(flowResponse())
      .mockResolvedValueOnce({
        ok: false,
        status: 403,
        type: "basic",
        json: async () => ({
          error: { id: "session_refresh_required" },
          redirect_browser_to: "http://localhost:4433/self-service/login?refresh=true",
        }),
      });

    try {
      await accountService.updateMyPassword("c", "n");
      throw new Error("should have thrown");
    } catch (err) {
      expect((err as SettingsFlowError).code).toBe("REAUTH_REQUIRED");
    }
  });
});
