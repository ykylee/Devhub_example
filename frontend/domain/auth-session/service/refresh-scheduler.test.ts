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

  // P0 회귀 가드 (router-idle-timeout-analysis 2026-05-28 + codex P2 #405 backoff).
  // transient_failed 후 후속 타이머 누락이 root cause 였고, 단순 reschedule() 호출은
  // tight 1초 retry loop 부작용 (codex P2) → 지수 backoff (5s/10s/20s/40s/60s/...) 로
  // 정정. 본 test 는 backoff 타이머가 설정되어 재시도되고 1초 단위 loop 가 아님을 검증.
  it("transient_failed → backoff retry (5s 후 재시도, tight 1초 loop 아님)", async () => {
    let callCount = 0;
    const performer = vi.fn<() => Promise<RefreshOutcome>>().mockImplementation(async () => {
      callCount += 1;
      // 첫 호출 transient, 두 번째 호출 ok (회복).
      return callCount === 1
        ? { kind: "transient_failed" as const, reason: "network_error" }
        : { kind: "ok" as const };
    });
    initRefreshScheduler(performer);

    // expires_in=120s → REFRESH_BUFFER 60s 전 (60s 후) 첫 호출.
    tokenStore.save({
      access_token: "a",
      refresh_token: "r",
      expires_in: 120,
      token_type: "Bearer",
    });
    await vi.advanceTimersByTimeAsync(60_500);
    await Promise.resolve();
    await Promise.resolve();
    expect(performer).toHaveBeenCalledTimes(1);

    // codex P2 (#405) 정합 — 1.5s 후엔 아직 호출 안 됨 (backoff base = 5s).
    await vi.advanceTimersByTimeAsync(1_500);
    await Promise.resolve();
    await Promise.resolve();
    expect(performer).toHaveBeenCalledTimes(1);

    // 추가 4s 진행 → 누적 5.5s, backoff 5s 경과 → 두 번째 호출.
    await vi.advanceTimersByTimeAsync(4_000);
    await Promise.resolve();
    await Promise.resolve();
    expect(performer).toHaveBeenCalledTimes(2);
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

  // initRefreshScheduler 의 idempotent 재호출 — 기존 unsubscribe + 기존 timer 모두 정리.
  // 첫 init 의 unsubscribe 가 호출되고 첫 timer 가 cleared 되어 동일 토큰 변동에도 첫
  // performer 는 호출되지 않아야 한다.
  it("re-init replaces previous unsubscribe + cancels previous timer", async () => {
    const first = vi.fn<() => Promise<RefreshOutcome>>().mockResolvedValue({ kind: "ok" });
    const second = vi.fn<() => Promise<RefreshOutcome>>().mockResolvedValue({ kind: "ok" });

    initRefreshScheduler(first);
    tokenStore.save({ access_token: "a", refresh_token: "r", expires_in: 2, token_type: "Bearer" });

    // Re-init 직후엔 첫 timer 가 clear 되고 첫 unsubscribe 도 해제되어야.
    initRefreshScheduler(second);

    // 두 번째 init 가 reschedule 했으므로 second 가 만료 직전에 호출, first 는 호출 안 됨.
    await vi.advanceTimersByTimeAsync(1500);
    expect(first).not.toHaveBeenCalled();
    expect(second).toHaveBeenCalledTimes(1);
  });

  // _resetRefreshScheduler — timer 가 존재하는 상태에서 clearTimeout 분기 cover.
  it("_resetRefreshScheduler cancels active timer", async () => {
    const performer = vi.fn<() => Promise<RefreshOutcome>>().mockResolvedValue({ kind: "ok" });
    initRefreshScheduler(performer);
    tokenStore.save({ access_token: "a", refresh_token: "r", expires_in: 60, token_type: "Bearer" });

    // 타이머 활성 상태에서 reset 호출 → clearTimeout 분기.
    _resetRefreshScheduler();
    await vi.advanceTimersByTimeAsync(120_000);
    expect(performer).not.toHaveBeenCalled();
  });
});
