"use client";

import { AuthenticatedActor, useStore } from "../store";
import { identityService } from "./identity.service";
import { tokenStore } from "@/lib/auth/token-store";
import { consumeVerifier, createPkceState } from "@/lib/auth/pkce";

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
   * Clears session and redirects to logout if needed
   */
  public logout() {
    tokenStore.clear();
    useStore.getState().clearActor();
    window.location.assign("/");
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
