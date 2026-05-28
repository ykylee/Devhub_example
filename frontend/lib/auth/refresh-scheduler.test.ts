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

  // P0 회귀 가드 (router-idle-timeout-analysis 2026-05-28) — transient_failed 후
  // reschedule() 누락이 root cause 였음. 본 test 는 transient 후에도 다음 만료
  // 시점에 performer 가 다시 호출됨을 검증.
  it("transient_failed → reschedule() 호출 → 다음 만료 직전 재시도", async () => {
    let callCount = 0;
    const performer = vi.fn<() => Promise<RefreshOutcome>>().mockImplementation(async () => {
      callCount += 1;
      // 첫 호출은 transient, 두 번째 호출은 ok (회복 시뮬레이션).
      return callCount === 1
        ? { kind: "transient_failed" as const, reason: "network_error" }
        : { kind: "ok" as const };
    });
    initRefreshScheduler(performer);

    // expires_in=120s 토큰 → REFRESH_BUFFER 60s 전 (즉 60s 후) 에 첫 호출.
    tokenStore.save({
      access_token: "a",
      refresh_token: "r",
      expires_in: 120,
      token_type: "Bearer",
    });
    // 첫 타이머 발화 (60s 후 + tick 여유).
    await vi.advanceTimersByTimeAsync(60_500);
    await Promise.resolve();
    await Promise.resolve();
    expect(performer).toHaveBeenCalledTimes(1);

    // P0 fix 적용 전엔 다음 타이머 미설정 → 추가 시간 진행해도 performer 가 다시
    // 호출 안 됨. fix 적용 후엔 reschedule() 호출되어 다음 타이머 (남은 만료까지)
    // 설정 → 추가 진행 시 두 번째 호출 발생.
    await vi.advanceTimersByTimeAsync(60_500);
    await Promise.resolve();
    await Promise.resolve();
    expect(performer.mock.calls.length).toBeGreaterThanOrEqual(2);
    expect(assignMock).not.toHaveBeenCalled();
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
