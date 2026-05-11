import { describe, it, expect, beforeEach, vi } from "vitest";
import { createPkceState, consumeVerifier } from "./pkce";

describe("createPkceState + consumeVerifier", () => {
  beforeEach(() => {
    sessionStorage.clear();
    vi.restoreAllMocks();
  });

  it("creates a state + code_challenge and stores verifier+state in sessionStorage", async () => {
    const { state, codeChallenge } = await createPkceState();

    expect(state).toMatch(/^[0-9a-f-]{36}$/i); // UUID
    expect(codeChallenge).toMatch(/^[A-Za-z0-9_-]+$/); // base64url, no padding
    expect(sessionStorage.getItem("oidc_state")).toBe(state);
    expect(sessionStorage.getItem("oidc_verifier")).toBeTruthy();
  });

  it("consumeVerifier returns the stored verifier when state matches and clears both keys", async () => {
    const { state } = await createPkceState();
    const storedVerifier = sessionStorage.getItem("oidc_verifier");

    const verifier = consumeVerifier(state);
    expect(verifier).toBe(storedVerifier);
    expect(sessionStorage.getItem("oidc_state")).toBeNull();
    expect(sessionStorage.getItem("oidc_verifier")).toBeNull();
  });

  it("throws when state does not match (CSRF protection)", async () => {
    await createPkceState();
    expect(() => consumeVerifier("tampered-state")).toThrow(/CSRF/);
  });

  it("throws when verifier is missing", async () => {
    sessionStorage.setItem("oidc_state", "abc");
    expect(() => consumeVerifier("abc")).toThrow(/Missing code verifier/);
  });
});
