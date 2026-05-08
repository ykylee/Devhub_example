"use client";

import { AuthenticatedActor, useStore } from "../store";
import { identityService } from "./identity.service";

const OIDC_AUTH_URL = process.env.NEXT_PUBLIC_OIDC_AUTH_URL ?? "http://127.0.0.1:4444/oauth2/auth";
const OIDC_TOKEN_URL = process.env.NEXT_PUBLIC_OIDC_TOKEN_URL ?? "http://127.0.0.1:4444/oauth2/token";
const OIDC_CLIENT_ID = process.env.NEXT_PUBLIC_OIDC_CLIENT_ID ?? "devhub-frontend";
const OIDC_REDIRECT_URI = process.env.NEXT_PUBLIC_OIDC_REDIRECT_URI ?? "http://127.0.0.1:3000/auth/callback";
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
    const state = crypto.randomUUID();
    const verifier = this.generateCodeVerifier();
    const challenge = await this.generateCodeChallenge(verifier);

    localStorage.setItem("oidc_state", state);
    localStorage.setItem("oidc_verifier", verifier);

    const url = new URL(OIDC_AUTH_URL);
    url.searchParams.set("client_id", OIDC_CLIENT_ID);
    url.searchParams.set("response_type", "code");
    url.searchParams.set("redirect_uri", OIDC_REDIRECT_URI);
    url.searchParams.set("scope", OIDC_SCOPE);
    url.searchParams.set("state", state);
    url.searchParams.set("code_challenge", challenge);
    url.searchParams.set("code_challenge_method", "S256");

    return url.toString();
  }

  /**
   * Exchanges authorization code for tokens
   */
  public async exchangeCode(code: string, state: string): Promise<TokenResponse> {
    const savedState = localStorage.getItem("oidc_state");
    const verifier = localStorage.getItem("oidc_verifier");

    if (state !== savedState) {
      throw new Error("Invalid state (CSRF protection failed)");
    }
    if (!verifier) {
      throw new Error("Missing code verifier");
    }

    const body = new URLSearchParams({
      grant_type: "authorization_code",
      code,
      redirect_uri: OIDC_REDIRECT_URI,
      client_id: OIDC_CLIENT_ID,
      code_verifier: verifier,
    });

    const response = await fetch(OIDC_TOKEN_URL, {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: body.toString(),
    });

    if (!response.ok) {
      const err = await response.json().catch(() => ({}));
      throw new Error(err.error_description || err.error || "Token exchange failed");
    }

    const tokens: TokenResponse = await response.json();
    this.saveTokens(tokens);

    // Cleanup OIDC temp state
    localStorage.removeItem("oidc_state");
    localStorage.removeItem("oidc_verifier");

    return tokens;
  }

  /**
   * Clears session and redirects to logout if needed
   */
  public logout() {
    localStorage.removeItem("devhub_access_token");
    localStorage.removeItem("devhub_refresh_token");
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

  private saveTokens(tokens: TokenResponse) {
    localStorage.setItem("devhub_access_token", tokens.access_token);
    if (tokens.refresh_token) {
      localStorage.setItem("devhub_refresh_token", tokens.refresh_token);
    }
  }

  private generateCodeVerifier(): string {
    const array = new Uint8Array(32);
    crypto.getRandomValues(array);
    return btoa(String.fromCharCode(...array))
      .replace(/\+/g, "-")
      .replace(/\//g, "_")
      .replace(/=/g, "");
  }

  private async generateCodeChallenge(verifier: string): Promise<string> {
    const encoder = new TextEncoder();
    const data = encoder.encode(verifier);
    const hash = await crypto.subtle.digest("SHA-256", data);
    return btoa(String.fromCharCode(...new Uint8Array(hash)))
      .replace(/\+/g, "-")
      .replace(/\//g, "_")
      .replace(/=/g, "");
  }

  public getAccessToken(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem("devhub_access_token");
  }
}

export const authService = AuthService.getInstance();
