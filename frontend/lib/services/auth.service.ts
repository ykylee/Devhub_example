"use client";

import { AuthenticatedActor, useStore } from "../store";
import { identityService } from "./identity.service";
import { tokenStore } from "@/lib/auth/token-store";
import { consumeVerifier, createPkceState } from "@/lib/auth/pkce";
import { performKratosBrowserLogout } from "@/lib/auth/kratos-logout";

const OIDC_AUTH_URL = process.env.NEXT_PUBLIC_OIDC_AUTH_URL ?? "http://localhost:4444/oauth2/auth";
const OIDC_CLIENT_ID = process.env.NEXT_PUBLIC_OIDC_CLIENT_ID ?? "devhub-frontend";
const OIDC_REDIRECT_URI = typeof window !== "undefined" 
  ? `${window.location.origin}/auth/callback`
  : (process.env.NEXT_PUBLIC_OIDC_REDIRECT_URI ?? "http://localhost:3000/auth/callback");
const OIDC_SCOPE = process.env.NEXT_PUBLIC_OIDC_SCOPE ?? "openid offline_access email profile";

export interface TokenResponse {
  access_token: string;
  refresh_token?: string;
  id_token?: string;
  expires_in: number;
  token_type: string;
}

class AuthService {
  private static instance: AuthService;

  private constructor() {}

  public static getInstance(): AuthService {
    if (!AuthService.instance) {
      AuthService.instance = new AuthService();
    }
    return AuthService.instance;
  }

  /**
   * Generates OIDC authorization URL with PKCE
   */
  public async getAuthorizeURL(): Promise<string> {
    const { state, codeChallenge } = await createPkceState();

    const url = new URL(OIDC_AUTH_URL);
    url.searchParams.set("client_id", OIDC_CLIENT_ID);
    url.searchParams.set("response_type", "code");
    url.searchParams.set("redirect_uri", OIDC_REDIRECT_URI);
    url.searchParams.set("scope", OIDC_SCOPE);
    url.searchParams.set("state", state);
    url.searchParams.set("code_challenge", codeChallenge);
    url.searchParams.set("code_challenge_method", "S256");

    return url.toString();
  }

  /**
   * Exchanges authorization code for tokens
   */
  public async exchangeCode(code: string, state: string): Promise<TokenResponse> {
    const verifier = consumeVerifier(state);
    const response = await fetch("/api/v1/auth/token", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
      code,
      code_verifier: verifier,
      redirect_uri: OIDC_REDIRECT_URI,
      client_id: OIDC_CLIENT_ID,
      }),
    });

    if (!response.ok) {
      const err = await response.json().catch(() => ({} as Record<string, string>));
      throw new Error(err.error || "Token exchange failed");
    }

    const payload = await response.json() as { data: TokenResponse };
    const tokens = payload.data;
    tokenStore.save(tokens);

    return tokens;
  }

  /**
   * Header Sign Out flow (no Hydra logout_challenge in scope). DEC-1=B:
   *   1) backend revokes the refresh token via Hydra public /oauth2/revoke
   *   2) local token + actor state cleared
   *   3) Kratos browser logout navigates to / via Kratos return URL
   *
   * Hydra session is not actively terminated here — the next /login attempt
   * has no Kratos credential, so Hydra cannot accept and the user is forced
   * back through the password flow.
   */
  public async logout(): Promise<void> {
    const refreshToken = tokenStore.getRefreshToken();
    if (refreshToken) {
      try {
        await fetch("/api/v1/auth/logout", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            refresh_token: refreshToken,
            client_id: OIDC_CLIENT_ID,
          }),
        });
      } catch (err) {
        console.warn("[AuthService] backend logout call failed (continuing)", err);
      }
    }
    tokenStore.clear();
    useStore.getState().clearActor();
    await performKratosBrowserLogout("/");
  }

  /**
   * RP-initiated logout entry point (Hydra urls.logout target /auth/logout).
   * Hydra has redirected the browser here with a logout_challenge; we hand
   * that to the backend, clear local state, and navigate to Hydra's
   * redirect_to which finishes the OIDC logout cleanup.
   */
  public async completeRPInitiatedLogout(challenge: string): Promise<void> {
    let redirectTo = "/";
    try {
      const res = await fetch("/api/v1/auth/logout", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ logout_challenge: challenge }),
      });
      if (res.ok) {
        const payload = (await res.json()) as { data?: { redirect_to?: string } };
        if (payload.data?.redirect_to) {
          redirectTo = payload.data.redirect_to;
        }
      } else {
        console.warn("[AuthService] backend logout accept failed", res.status);
      }
    } catch (err) {
      console.warn("[AuthService] backend logout accept call failed", err);
    }
    tokenStore.clear();
    useStore.getState().clearActor();
    window.location.assign(redirectTo);
  }

  /**
   * Resolves the current user identity using the access token
   */
  public async resolveIdentity(): Promise<AuthenticatedActor> {
    try {
      const actor = await identityService.whoAmI();
      useStore.getState().setActor(actor);
      return actor;
    } catch (error) {
      console.error("[AuthService] resolveIdentity failed:", error);
      this.logout();
      throw error;
    }
  }

  public getAccessToken(): string | null {
    return tokenStore.getAccessToken();
  }
}

export const authService = AuthService.getInstance();
