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
    const map = JSON.parse(sessionStorage.getItem("oidc_pkce_map") ?? "{}") as Record<string, string>;
    expect(map[state]).toBeTruthy();
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

  it("supports overlapping OIDC starts and consumes by matching state", async () => {
    const first = await createPkceState();
    const second = await createPkceState();
    const mapBefore = JSON.parse(sessionStorage.getItem("oidc_pkce_map") ?? "{}") as Record<string, string>;
    expect(Object.keys(mapBefore).length).toBe(2);

    const secondVerifier = consumeVerifier(second.state);
    expect(secondVerifier).toBe(mapBefore[second.state]);
    const mapAfter = JSON.parse(sessionStorage.getItem("oidc_pkce_map") ?? "{}") as Record<string, string>;
    expect(mapAfter[first.state]).toBeTruthy();
    expect(mapAfter[second.state]).toBeUndefined();
  });
});

describe("randomState fallback (crypto.randomUUID absent)", () => {
  beforeEach(() => {
    sessionStorage.clear();
  });

  it("falls back to manual UUID assembly when crypto.randomUUID is undefined", async () => {
    const originalRandomUUID = crypto.randomUUID;
    // Reflect-defineProperty so we can revert; some envs forbid plain delete.
    Object.defineProperty(crypto, "randomUUID", { value: undefined, configurable: true, writable: true });
    try {
      const { state } = await createPkceState();
      // RFC 4122 v4 shape: 8-4-4-4-12, lowercase hex.
      expect(state).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/);
    } finally {
      Object.defineProperty(crypto, "randomUUID", { value: originalRandomUUID, configurable: true, writable: true });
    }
  });
});

describe("challengeFromVerifier — sha256Fallback path", () => {
  beforeEach(() => {
    sessionStorage.clear();
  });

  it("uses sha256Fallback when crypto.subtle.digest is unavailable", async () => {
    const originalSubtle = crypto.subtle;
    // happy-dom 의 crypto.subtle 를 일시 제거 — challengeFromVerifier 의 fallback 분기 트리거.
    Object.defineProperty(crypto, "subtle", { value: undefined, configurable: true, writable: true });
    try {
      const rfcVerifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk";
      const rfcChallenge = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM";
      const got = await challengeFromVerifier(rfcVerifier);
      expect(got).toBe(rfcChallenge);
    } finally {
      Object.defineProperty(crypto, "subtle", { value: originalSubtle, configurable: true, writable: true });
    }
  });
});

describe("consumeVerifier — legacy fallback paths", () => {
  beforeEach(() => {
    sessionStorage.clear();
  });

  it("legacy: when only oidc_state/oidc_verifier exist (no pkce map), valid match returns verifier", () => {
    sessionStorage.setItem("oidc_state", "legacy-state");
    sessionStorage.setItem("oidc_verifier", "legacy-verifier");
    const v = consumeVerifier("legacy-state");
    expect(v).toBe("legacy-verifier");
    expect(sessionStorage.getItem("oidc_state")).toBeNull();
    expect(sessionStorage.getItem("oidc_verifier")).toBeNull();
  });

  it("legacy: CSRF mismatch when saved state differs", () => {
    sessionStorage.setItem("oidc_state", "saved");
    sessionStorage.setItem("oidc_verifier", "v");
    expect(() => consumeVerifier("bogus")).toThrow(/CSRF/);
  });

  it("legacy: missing verifier even with matching state", () => {
    sessionStorage.setItem("oidc_state", "matching");
    // no oidc_verifier
    expect(() => consumeVerifier("matching")).toThrow(/Missing code verifier/);
  });

  it("pkce_map: consuming the only entry removes the map key entirely", async () => {
    const { state } = await createPkceState();
    expect(sessionStorage.getItem("oidc_pkce_map")).toBeTruthy();
    consumeVerifier(state);
    // Single-entry map → after consume → map key removed entirely.
    expect(sessionStorage.getItem("oidc_pkce_map")).toBeNull();
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
