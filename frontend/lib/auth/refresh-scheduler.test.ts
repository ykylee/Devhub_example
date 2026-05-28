import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { tokenStore } from "./token-store";
import { initRefreshScheduler, _resetRefreshScheduler } from "./refresh-scheduler";
import { _resetSessionExpiredGuard } from "./session-death";
import type { RefreshOutcome } from "./refresh";

describe("refresh-scheduler", () => {
  let assignMock: ReturnType<typeof vi.fn>;
  let originalLocation: Location;

  beforeEach(() => {
    sessionStorage.clear();
    tokenStore.clear();
    _resetSessionExpiredGuard();
    _resetRefreshScheduler();
    vi.useFakeTimers();
    assignMock = vi.fn();
    originalLocation = window.location;
    Object.defineProperty(window, "location", {
      value: { ...originalLocation, assign: assignMock, pathname: "/admin/catalog" },
      writable: true,
      configurable: true,
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    Object.defineProperty(window, "location", {
      value: originalLocation,
      writable: true,
      configurable: true,
    });
  });

  it("schedules refresh and calls performer near MIN_SCHEDULE_MS when expires_in is short", async () => {
    const performer = vi.fn<() => Promise<RefreshOutcome>>().mockResolvedValue({ kind: "ok" });
    initRefreshScheduler(performer);

    tokenStore.save({
      access_token: "a",
      refresh_token: "r",
      expires_in: 2,
      token_type: "Bearer",
    });
    expect(performer).not.toHaveBeenCalled();

    await vi.advanceTimersByTimeAsync(1500);
    expect(performer).toHaveBeenCalledTimes(1);
  });

  it("auth_failed → triggerSessionExpired (/login redirect + clear)", async () => {
    const performer = vi.fn<() => Promise<RefreshOutcome>>()
      .mockResolvedValue({ kind: "auth_failed", reason: "http_400" });
    initRefreshScheduler(performer);

    tokenStore.save({
      access_token: "a",
      refresh_token: "r",
      expires_in: 2,
      token_type: "Bearer",
    });
    await vi.advanceTimersByTimeAsync(1500);
    await Promise.resolve();
    await Promise.resolve();

    expect(performer).toHaveBeenCalledTimes(1);
    expect(assignMock).toHaveBeenCalled();
    expect(String(assignMock.mock.calls[0][0])).toContain("/login?error=session_expired");
  });

  it("transient_failed → 세션 사망 처리 안 함 (assign 호출 없음)", async () => {
    const performer = vi.fn<() => Promise<RefreshOutcome>>()
      .mockResolvedValue({ kind: "transient_failed", reason: "network_error" });
    initRefreshScheduler(performer);

    tokenStore.save({
      access_token: "a",
      refresh_token: "r",
      expires_in: 2,
      token_type: "Bearer",
    });
    await vi.advanceTimersByTimeAsync(1500);
    await Promise.resolve();
    await Promise.resolve();

    expect(performer).toHaveBeenCalledTimes(1);
    // transient 는 세션을 죽이지 않음 — assign 호출 없음, tokenStore 도 그대로.
    expect(assignMock).not.toHaveBeenCalled();
    expect(tokenStore.getAccessToken()).toBe("a");
  });

  it("performer 가 throw 하면 transient 로 분류 (세션 사망 안 함)", async () => {
    const performer = vi.fn<() => Promise<RefreshOutcome>>()
      .mockRejectedValue(new Error("unexpected throw"));
    initRefreshScheduler(performer);

    tokenStore.save({
      access_token: "a",
      refresh_token: "r",
      expires_in: 2,
      token_type: "Bearer",
    });
    await vi.advanceTimersByTimeAsync(1500);
    await Promise.resolve();
    await Promise.resolve();

    expect(performer).toHaveBeenCalledTimes(1);
    expect(assignMock).not.toHaveBeenCalled();
  });

  it("clear() cancels the scheduled timer", async () => {
    const performer = vi.fn<() => Promise<RefreshOutcome>>().mockResolvedValue({ kind: "ok" });
    initRefreshScheduler(performer);

    tokenStore.save({
      access_token: "a",
      refresh_token: "r",
      expires_in: 2,
      token_type: "Bearer",
    });
    tokenStore.clear();
    await vi.advanceTimersByTimeAsync(5000);
    expect(performer).not.toHaveBeenCalled();
  });

  it("save twice re-schedules (only the latest fires)", async () => {
    const performer = vi.fn<() => Promise<RefreshOutcome>>().mockResolvedValue({ kind: "ok" });
    initRefreshScheduler(performer);

    tokenStore.save({
      access_token: "a",
      refresh_token: "r",
      expires_in: 2,
      token_type: "Bearer",
    });
    tokenStore.save({
      access_token: "a2",
      refresh_token: "r2",
      expires_in: 3600,
      token_type: "Bearer",
    });

    await vi.advanceTimersByTimeAsync(1500);
    expect(performer).not.toHaveBeenCalled();
  });

  it("does nothing when no performer is registered (init not called)", async () => {
    _resetRefreshScheduler();
    tokenStore.save({
      access_token: "a",
      refresh_token: "r",
      expires_in: 2,
      token_type: "Bearer",
    });
    await vi.advanceTimersByTimeAsync(2000);
    expect(true).toBe(true);
  });
});
