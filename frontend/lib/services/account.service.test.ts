import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { accountService, SettingsFlowError } from "./account.service";

// account.service.updateMyPassword now drives the backend proxy
// POST /api/v1/account/password (PR-L4, DEC-A=A). The handler translates
// the backend response's `code` field onto SettingsFlowErrorCode so the
// /account page can stick to the existing REAUTH_REQUIRED / VALIDATION /
// CURRENT_PASSWORD_INVALID / SUBMIT_FAILED branches.

describe("accountService.updateMyPassword", () => {
  let fetchSpy: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchSpy = vi.fn();
    vi.stubGlobal("fetch", fetchSpy);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  function backendResponse(status: number, body: unknown) {
    return {
      ok: status >= 200 && status < 300,
      status,
      type: "basic",
      text: async () => JSON.stringify(body),
    };
  }

  it("happy path: POSTs /api/v1/account/password with both fields, resolves on 200", async () => {
    fetchSpy.mockResolvedValueOnce(backendResponse(200, { status: "ok", data: { user_id: "alice" } }));

    await expect(accountService.updateMyPassword("OldPass-1!", "NewPass-2!")).resolves.toBeUndefined();

    expect(fetchSpy).toHaveBeenCalledTimes(1);
    const [url, init] = fetchSpy.mock.calls[0];
    expect(url).toBe("/api/v1/account/password");
    expect(init.method).toBe("POST");
    expect(JSON.parse(init.body)).toEqual({ current_password: "OldPass-1!", new_password: "NewPass-2!" });
  });

  it("CURRENT_PASSWORD_INVALID when backend returns 401 + code=current_password_invalid", async () => {
    fetchSpy.mockResolvedValueOnce(
      backendResponse(401, { status: "unauthenticated", error: "current password is incorrect", code: "current_password_invalid" }),
    );

    try {
      await accountService.updateMyPassword("wrong", "NewPass-2!");
      throw new Error("should have thrown");
    } catch (err) {
      expect(err).toBeInstanceOf(SettingsFlowError);
      expect((err as SettingsFlowError).code).toBe("CURRENT_PASSWORD_INVALID");
      expect((err as SettingsFlowError).message).toContain("current password");
    }
  });

  it("REAUTH_REQUIRED when backend returns 401 + code=reauth_required", async () => {
    fetchSpy.mockResolvedValueOnce(
      backendResponse(401, { status: "unauthenticated", error: "re-authentication required", code: "reauth_required" }),
    );

    try {
      await accountService.updateMyPassword("c", "n");
      throw new Error("should have thrown");
    } catch (err) {
      expect((err as SettingsFlowError).code).toBe("REAUTH_REQUIRED");
    }
  });

  it("VALIDATION when backend returns 400 + code=validation (weak new password)", async () => {
    fetchSpy.mockResolvedValueOnce(
      backendResponse(400, { status: "rejected", error: "password too short", code: "validation" }),
    );

    try {
      await accountService.updateMyPassword("c", "shrt");
      throw new Error("should have thrown");
    } catch (err) {
      expect((err as SettingsFlowError).code).toBe("VALIDATION");
      expect((err as SettingsFlowError).message).toContain("password too short");
    }
  });

  it("FLOW_INIT_FAILED when backend returns 410 + code=flow_expired", async () => {
    fetchSpy.mockResolvedValueOnce(backendResponse(410, { status: "gone", error: "flow expired", code: "flow_expired" }));

    try {
      await accountService.updateMyPassword("c", "n");
      throw new Error("should have thrown");
    } catch (err) {
      expect((err as SettingsFlowError).code).toBe("FLOW_INIT_FAILED");
    }
  });

  it("SUBMIT_FAILED for unknown error codes (defense in depth)", async () => {
    fetchSpy.mockResolvedValueOnce(backendResponse(500, { status: "failed", error: "internal error" }));

    try {
      await accountService.updateMyPassword("c", "n");
      throw new Error("should have thrown");
    } catch (err) {
      expect((err as SettingsFlowError).code).toBe("SUBMIT_FAILED");
    }
  });

  it("503 backend unavailable also surfaces as SUBMIT_FAILED", async () => {
    fetchSpy.mockResolvedValueOnce(
      backendResponse(503, { status: "unavailable", error: "account password proxy requires KratosLogin + OrganizationStore" }),
    );

    try {
      await accountService.updateMyPassword("c", "n");
      throw new Error("should have thrown");
    } catch (err) {
      expect((err as SettingsFlowError).code).toBe("SUBMIT_FAILED");
    }
  });
});
