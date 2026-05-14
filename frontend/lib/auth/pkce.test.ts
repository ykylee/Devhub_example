import { describe, it, expect, beforeEach, vi } from "vitest";
import {
  createPkceState,
  consumeVerifier,
  challengeFromVerifier,
  sha256FallbackBase64Url,
} from "./pkce";

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

describe("challengeFromVerifier — PKCE spec conformance", () => {
  // RFC 7636 Appendix B
  const rfcVerifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk";
  const rfcChallenge = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM";

  it("matches RFC 7636 Appendix B vector via the subtle path", async () => {
    const got = await challengeFromVerifier(rfcVerifier);
    expect(got).toBe(rfcChallenge);
  });

  it("sha256Fallback produces the same RFC 7636 Appendix B output", () => {
    expect(sha256FallbackBase64Url(rfcVerifier)).toBe(rfcChallenge);
  });

  it("sha256Fallback agrees with crypto.subtle.digest over 50 random verifiers", async () => {
    for (let i = 0; i < 50; i++) {
      const bytes = new Uint8Array(32);
      crypto.getRandomValues(bytes);
      // Use the same base64url-encoded verifier shape as randomVerifier().
      const verifier = btoa(String.fromCharCode(...bytes))
        .replace(/\+/g, "-")
        .replace(/\//g, "_")
        .replace(/=/g, "");
      const subtleHash = await crypto.subtle.digest(
        "SHA-256",
        new TextEncoder().encode(verifier),
      );
      const subtleB64 = btoa(String.fromCharCode(...new Uint8Array(subtleHash)))
        .replace(/\+/g, "-")
        .replace(/\//g, "_")
        .replace(/=/g, "");
      expect(sha256FallbackBase64Url(verifier)).toBe(subtleB64);
    }
  });
});
