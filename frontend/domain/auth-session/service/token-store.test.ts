import { describe, it, expect, beforeEach, vi } from "vitest";
import { tokenStore } from "./token-store";

describe("tokenStore", () => {
  beforeEach(() => {
    sessionStorage.clear();
    tokenStore.clear(); // also resets in-memory cache
  });

  it("save persists access/refresh/id and getters return them", () => {
    tokenStore.save({
      access_token: "access-1",
      refresh_token: "refresh-1",
      id_token: "id-1",
      expires_in: 1800,
      token_type: "Bearer",
    });

    expect(tokenStore.getAccessToken()).toBe("access-1");
    expect(tokenStore.getRefreshToken()).toBe("refresh-1");
    expect(tokenStore.getIdToken()).toBe("id-1");
    expect(sessionStorage.getItem("devhub_access_token")).toBe("access-1");
    expect(sessionStorage.getItem("devhub_refresh_token")).toBe("refresh-1");
    expect(sessionStorage.getItem("devhub_id_token")).toBe("id-1");
  });

  it("save without refresh_token / id_token clears the corresponding storage key", () => {
    tokenStore.save({
      access_token: "access-2",
      expires_in: 1800,
      token_type: "Bearer",
    });
    expect(tokenStore.getRefreshToken()).toBeNull();
    expect(tokenStore.getIdToken()).toBeNull();
    expect(sessionStorage.getItem("devhub_refresh_token")).toBeNull();
    expect(sessionStorage.getItem("devhub_id_token")).toBeNull();
  });

  it("clear wipes both in-memory cache and sessionStorage", () => {
    tokenStore.save({
      access_token: "a",
      refresh_token: "r",
      id_token: "i",
      expires_in: 1800,
      token_type: "Bearer",
    });
    tokenStore.clear();
    expect(tokenStore.getAccessToken()).toBeNull();
    expect(tokenStore.getRefreshToken()).toBeNull();
    expect(tokenStore.getIdToken()).toBeNull();
    expect(sessionStorage.getItem("devhub_access_token")).toBeNull();
  });

  it("getters fall through to sessionStorage when the in-memory cache is cold", () => {
    sessionStorage.setItem("devhub_access_token", "from-storage");
    sessionStorage.setItem("devhub_refresh_token", "refresh-from-storage");
    sessionStorage.setItem("devhub_id_token", "id-from-storage");
    // tokenStore.clear in beforeEach reset the in-memory cache; getters
    // should re-hydrate from sessionStorage on the next read.
    expect(tokenStore.getAccessToken()).toBe("from-storage");
    expect(tokenStore.getRefreshToken()).toBe("refresh-from-storage");
    expect(tokenStore.getIdToken()).toBe("id-from-storage");
  });

  it("save computes expires_at from expires_in (ms epoch, persisted)", () => {
    const before = Date.now();
    tokenStore.save({
      access_token: "a",
      expires_in: 300, // 5분
      token_type: "Bearer",
    });
    const exp = tokenStore.getExpiresAt();
    expect(exp).not.toBeNull();
    // expires_at ≈ now + 300_000 ms
    expect(exp!).toBeGreaterThanOrEqual(before + 299_000);
    expect(exp!).toBeLessThanOrEqual(Date.now() + 301_000);
    expect(Number(sessionStorage.getItem("devhub_access_token_expires_at"))).toBe(exp);
  });

  it("save with expires_in=0 leaves expires_at null", () => {
    tokenStore.save({
      access_token: "a",
      expires_in: 0,
      token_type: "Bearer",
    });
    expect(tokenStore.getExpiresAt()).toBeNull();
    expect(sessionStorage.getItem("devhub_access_token_expires_at")).toBeNull();
  });

  it("clear wipes expires_at + storage", () => {
    tokenStore.save({ access_token: "a", expires_in: 300, token_type: "Bearer" });
    expect(tokenStore.getExpiresAt()).not.toBeNull();
    tokenStore.clear();
    expect(tokenStore.getExpiresAt()).toBeNull();
    expect(sessionStorage.getItem("devhub_access_token_expires_at")).toBeNull();
  });

  it("subscribeExpiryChange fires on save and clear; unsubscribe stops it", () => {
    const events: Array<number | null> = [];
    const unsub = tokenStore.subscribeExpiryChange((exp) => events.push(exp));

    tokenStore.save({ access_token: "a", expires_in: 600, token_type: "Bearer" });
    tokenStore.clear();

    expect(events).toHaveLength(2);
    expect(events[0]).not.toBeNull(); // save → expires_at 값
    expect(events[1]).toBeNull();     // clear → null

    unsub();
    tokenStore.save({ access_token: "b", expires_in: 600, token_type: "Bearer" });
    expect(events).toHaveLength(2); // unsubscribe 후 호출 안 됨
  });

  it("save with non-finite expires_in (NaN) leaves expires_at null and removes storage key", () => {
    // covers the Number.isFinite guard branch (false path).
    tokenStore.save({
      access_token: "a",
      expires_in: Number.NaN,
      token_type: "Bearer",
    });
    expect(tokenStore.getExpiresAt()).toBeNull();
    expect(sessionStorage.getItem("devhub_access_token_expires_at")).toBeNull();
  });

  it("save with negative expires_in falls through to null (defensive)", () => {
    tokenStore.save({
      access_token: "a",
      expires_in: -1,
      token_type: "Bearer",
    });
    expect(tokenStore.getExpiresAt()).toBeNull();
  });

  it("listener throw is caught and warn-logged; other listeners still notified", () => {
    const warnSpy = vi.spyOn(console, "warn").mockImplementation(() => undefined);
    const ok: Array<number | null> = [];
    const unsubBad = tokenStore.subscribeExpiryChange(() => {
      throw new Error("listener boom");
    });
    const unsubGood = tokenStore.subscribeExpiryChange((exp) => ok.push(exp));

    tokenStore.save({ access_token: "a", expires_in: 60, token_type: "Bearer" });

    // bad listener throw 가 catch 되어 good listener 도 호출되어야 한다.
    expect(ok.length).toBeGreaterThanOrEqual(1);
    expect(warnSpy).toHaveBeenCalled();

    unsubBad();
    unsubGood();
    warnSpy.mockRestore();
  });
});
