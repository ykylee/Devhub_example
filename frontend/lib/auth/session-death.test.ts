import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { tokenStore } from "./token-store";
import { triggerSessionExpired, _resetSessionExpiredGuard } from "./session-death";

describe("session-death.triggerSessionExpired", () => {
  let assignMock: ReturnType<typeof vi.fn>;
  let originalLocation: Location;

  function stubLocation(pathname: string) {
    assignMock = vi.fn();
    Object.defineProperty(window, "location", {
      value: { ...originalLocation, assign: assignMock, pathname },
      writable: true,
      configurable: true,
    });
  }

  beforeEach(() => {
    sessionStorage.clear();
    tokenStore.clear();
    _resetSessionExpiredGuard();
    originalLocation = window.location;
    stubLocation("/admin/catalog");
  });

  afterEach(() => {
    Object.defineProperty(window, "location", {
      value: originalLocation,
      writable: true,
      configurable: true,
    });
  });

  it("clears tokens and redirects to /login?error=session_expired", () => {
    tokenStore.save({
      access_token: "a",
      refresh_token: "r",
      expires_in: 300,
      token_type: "Bearer",
    });
    triggerSessionExpired();
    expect(tokenStore.getAccessToken()).toBeNull();
    expect(tokenStore.getRefreshToken()).toBeNull();
    expect(assignMock).toHaveBeenCalledTimes(1);
    expect(String(assignMock.mock.calls[0][0])).toContain("/login?error=session_expired");
  });

  it("is idempotent: multiple calls = single redirect", () => {
    triggerSessionExpired();
    triggerSessionExpired();
    triggerSessionExpired();
    expect(assignMock).toHaveBeenCalledTimes(1);
  });

  it("does not redirect when already on /login", () => {
    stubLocation("/login");
    triggerSessionExpired();
    expect(assignMock).not.toHaveBeenCalled();
  });

  it("does not redirect when path starts with /login (query / sub path)", () => {
    stubLocation("/login/");
    triggerSessionExpired();
    expect(assignMock).not.toHaveBeenCalled();
  });

  it("accepts custom reason and url-encodes it", () => {
    triggerSessionExpired("refresh_failed");
    expect(String(assignMock.mock.calls[0][0])).toContain("error=refresh_failed");
  });
});
