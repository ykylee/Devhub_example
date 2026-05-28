import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { tokenStore } from "./token-store";
import { initRefreshScheduler, _resetRefreshScheduler } from "./refresh-scheduler";
import { _resetSessionExpiredGuard } from "./session-death";

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

  it("schedules refresh and calls refreshFn around MIN_SCHEDULE_MS when expires_in is short", async () => {
    const refreshFn = vi.fn().mockResolvedValue(undefined);
    initRefreshScheduler(refreshFn);

    tokenStore.save({
      access_token: "a",
      refresh_token: "r",
      expires_in: 2, // 2초 — refreshAt 이 과거 → delay = MIN_SCHEDULE_MS (1000ms)
      token_type: "Bearer",
    });
    expect(refreshFn).not.toHaveBeenCalled();

    await vi.advanceTimersByTimeAsync(1500);
    expect(refreshFn).toHaveBeenCalledTimes(1);
  });

  it("triggers session-expired (clear + redirect) when refreshFn rejects", async () => {
    const refreshFn = vi.fn().mockRejectedValue(new Error("refresh failed"));
    initRefreshScheduler(refreshFn);

    tokenStore.save({
      access_token: "a",
      refresh_token: "r",
      expires_in: 2,
      token_type: "Bearer",
    });
    await vi.advanceTimersByTimeAsync(1500);
    // refreshFn rejection 의 async 마이크로태스크 flush 를 위해 한 번 더 awaiting.
    await Promise.resolve();
    await Promise.resolve();

    expect(refreshFn).toHaveBeenCalledTimes(1);
    expect(assignMock).toHaveBeenCalled();
    expect(String(assignMock.mock.calls[0][0])).toContain("/login?error=session_expired");
  });

  it("clear() cancels the scheduled timer", async () => {
    const refreshFn = vi.fn().mockResolvedValue(undefined);
    initRefreshScheduler(refreshFn);

    tokenStore.save({
      access_token: "a",
      refresh_token: "r",
      expires_in: 2,
      token_type: "Bearer",
    });
    tokenStore.clear();
    await vi.advanceTimersByTimeAsync(5000);
    expect(refreshFn).not.toHaveBeenCalled();
  });

  it("save twice re-schedules (only the latest fires)", async () => {
    const refreshFn = vi.fn().mockResolvedValue(undefined);
    initRefreshScheduler(refreshFn);

    tokenStore.save({
      access_token: "a",
      refresh_token: "r",
      expires_in: 2,
      token_type: "Bearer",
    });
    // 두 번째 save (e.g., refresh 결과로 새 토큰) — 기존 타이머 취소 + 새로 스케줄.
    tokenStore.save({
      access_token: "a2",
      refresh_token: "r2",
      expires_in: 3600, // 60분 — refreshAt 미래(~59분), MIN_SCHEDULE_MS 아님
      token_type: "Bearer",
    });

    // 1.5s 후엔 첫 타이머가 이미 취소됐어야 함.
    await vi.advanceTimersByTimeAsync(1500);
    expect(refreshFn).not.toHaveBeenCalled();
  });

  it("does nothing when no refreshFn is registered (init not called)", async () => {
    _resetRefreshScheduler();
    // init 안 한 상태에서 save 해도 listener 없음 → no-op.
    tokenStore.save({
      access_token: "a",
      refresh_token: "r",
      expires_in: 2,
      token_type: "Bearer",
    });
    await vi.advanceTimersByTimeAsync(2000);
    // 별도 검증 대상 없음 — 그저 throw 없이 통과 확인.
    expect(true).toBe(true);
  });
});
