"use client";

const OIDC_STATE_KEY = "oidc_state";
const OIDC_VERIFIER_KEY = "oidc_verifier";

function base64Url(input: Uint8Array): string {
  return btoa(String.fromCharCode(...input))
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=/g, "");
}

function randomVerifier(): string {
  const bytes = new Uint8Array(32);
  crypto.getRandomValues(bytes);
  return base64Url(bytes);
}

async function challengeFromVerifier(verifier: string): Promise<string> {
  const hash = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(verifier));
  return base64Url(new Uint8Array(hash));
}

export async function createPkceState(): Promise<{ state: string; codeChallenge: string }> {
  const state = crypto.randomUUID();
  const verifier = randomVerifier();
  const codeChallenge = await challengeFromVerifier(verifier);
  sessionStorage.setItem(OIDC_STATE_KEY, state);
  sessionStorage.setItem(OIDC_VERIFIER_KEY, verifier);
  return { state, codeChallenge };
}

export function consumeVerifier(expectedState: string): string {
  const savedState = sessionStorage.getItem(OIDC_STATE_KEY);
  const verifier = sessionStorage.getItem(OIDC_VERIFIER_KEY);
  sessionStorage.removeItem(OIDC_STATE_KEY);
  sessionStorage.removeItem(OIDC_VERIFIER_KEY);
  if (!savedState || savedState !== expectedState) {
    throw new Error("Invalid state (CSRF protection failed)");
  }
  if (!verifier) {
    throw new Error("Missing code verifier");
  }
  return verifier;
}
