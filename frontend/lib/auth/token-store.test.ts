import { describe, it, expect, beforeEach } from "vitest";
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
});
