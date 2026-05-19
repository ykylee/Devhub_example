"use client";

import { AuthenticatedActor, useStore } from "../store";
import { identityService } from "./identity.service";
import { tokenStore } from "@/lib/auth/token-store";
import { consumeVerifier, createPkceState } from "@/lib/auth/pkce";

import { BASE_PATH, OIDC_AUTH_URL, OIDC_ISSUER_URL, OIDC_REDIRECT_URI as OIDC_REDIRECT_URI_DEFAULT } from "../config/endpoints";

const OIDC_CLIENT_ID = process.env.NEXT_PUBLIC_OIDC_CLIENT_ID ?? "devhub-frontend";
// sprint -s (PR #187) — basePath 포함 정합 (sprint -j codex review #9 #4 backend 확장 carve #3).
// ADR-0018 basePath /devhub 환경에서 origin/auth/callback (basePath 없음) → origin/devhub/auth/callback 정합.
// Keycloak whitelist 의 Valid redirect URIs 가 /devhub/auth/callback 으로 정합 (keycloak_operations §3.1).
const OIDC_REDIRECT_URI = typeof window !== "undefined"
  ? `${window.location.origin}${BASE_PATH}/auth/callback`
  : OIDC_REDIRECT_URI_DEFAULT;
const OIDC_SCOPE = process.env.NEXT_PUBLIC_OIDC_SCOPE ?? "openid offline_access email profile";

export interface TokenResponse {
  access_token: string;
  refresh_token?: string;
  id_token?: string;
  expires_in: number;
  token_type: string;
}

interface RuntimeConfigResponse {
  oidc_auth_url?: string;
  oidc_redirect_uri?: string;
  oidc_issuer_url?: string;
}

interface OIDCDiscoveryDocument {
  authorization_endpoint: string;
  token_endpoint: string;
  end_session_endpoint?: string;
}

class AuthService {
  private static instance: AuthService;
  private runtimeConfig?: { oidcAuthURL: string; oidcRedirectURI: string; oidcIssuerURL: string };
  private discoveryDoc?: OIDCDiscoveryDocument;

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
    const { state, codeChallenge, codeChallengeMethod } = await createPkceState();
    const discovery = await this.getDiscovery();
    const runtimeConfig = await this.getRuntimeOIDCConfig();

    const url = new URL(discovery.authorization_endpoint || runtimeConfig.oidcAuthURL);
    url.searchParams.set("client_id", OIDC_CLIENT_ID);
    url.searchParams.set("response_type", "code");
    url.searchParams.set("redirect_uri", runtimeConfig.oidcRedirectURI);
    url.searchParams.set("scope", OIDC_SCOPE);
    url.searchParams.set("state", state);
    url.searchParams.set("code_challenge", codeChallenge);
    url.searchParams.set("code_challenge_method", codeChallengeMethod);

    return url.toString();
  }

  /**
   * Exchanges authorization code for tokens
   */
  public async exchangeCode(code: string, state: string): Promise<TokenResponse> {
    const verifier = consumeVerifier(state);
    const runtimeConfig = await this.getRuntimeOIDCConfig();
    const discovery = await this.getDiscovery();
    const response = await fetch(discovery.token_endpoint, {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({
        grant_type: "authorization_code",
        code,
        code_verifier: verifier,
        redirect_uri: runtimeConfig.oidcRedirectURI,
        client_id: OIDC_CLIENT_ID,
      }).toString(),
    });

    if (!response.ok) {
      const err = await response.json().catch(() => ({} as Record<string, string>));
      throw new Error(err.error || "Token exchange failed");
    }

    const tokens = await response.json() as TokenResponse;
    tokenStore.save(tokens);

    return tokens;
  }

  public async logout(): Promise<void> {
    const idToken = tokenStore.getIdToken();
    const runtimeConfig = await this.getRuntimeOIDCConfig();
    const discovery = await this.getDiscovery();

    tokenStore.clear();
    useStore.getState().setIsLoggingOut(true);
    useStore.getState().clearActor();

    const endSessionEndpoint =
      discovery.end_session_endpoint ||
      `${runtimeConfig.oidcIssuerURL}/protocol/openid-connect/logout`;

    try {
      const url = new URL(endSessionEndpoint);
      url.searchParams.set("client_id", OIDC_CLIENT_ID);
      // sprint -s (PR #187) basePath 포함 정합 — ADR-0018 basePath /devhub 환경에서
      // logout 후 사용자가 DevHub 진입점 (/devhub/) 으로 redirect. sprint -j codex review #9 #4
      // 의 basePath 미포함 표기는 sprint -s 에서 backend 확장 carve 로 정합.
      url.searchParams.set("post_logout_redirect_uri", `${window.location.origin}${BASE_PATH}/`);
      if (idToken) {
        url.searchParams.set("id_token_hint", idToken);
      }
      window.location.assign(url.toString());
      return;
    } catch (error) {
      console.warn("[AuthService] logout redirect build failed", error);
      window.location.assign("/");
    }
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

  private async getRuntimeOIDCConfig(): Promise<{ oidcAuthURL: string; oidcRedirectURI: string; oidcIssuerURL: string }> {
    if (this.runtimeConfig) {
      return this.runtimeConfig;
    }

    const fallback = {
      oidcAuthURL: OIDC_AUTH_URL,
      oidcRedirectURI: OIDC_REDIRECT_URI,
      oidcIssuerURL: OIDC_ISSUER_URL,
    };

    try {
      const response = await fetch("/api/runtime-config", {
        method: "GET",
        cache: "no-store",
      });
      if (!response.ok) {
        return fallback;
      }

      const body = (await response.json()) as RuntimeConfigResponse;
      const oidcAuthURL = body.oidc_auth_url?.trim() || fallback.oidcAuthURL;
      const oidcRedirectURI = body.oidc_redirect_uri?.trim() || fallback.oidcRedirectURI;
      const oidcIssuerURL = body.oidc_issuer_url?.trim() || fallback.oidcIssuerURL;
      this.runtimeConfig = { oidcAuthURL, oidcRedirectURI, oidcIssuerURL };
      return this.runtimeConfig;
    } catch (error) {
      console.warn("[AuthService] runtime OIDC config fetch failed, using fallback", error);
      return fallback;
    }
  }

  // getAccountConsoleURL returns the Keycloak Account Console URL
  // (`${issuer}/account/`) using the same runtime OIDC config path that login
  // / logout flows use. This keeps the link working in deployments that
  // surface the issuer via /api/runtime-config server env (rather than baking
  // NEXT_PUBLIC_OIDC_ISSUER_URL into the browser bundle).
  // sprint claude/work_260519-ad Stage 3 — codex P1 (runtime config 정합) +
  // self-review P1-2 (e2e env 미설정 시 link hidden) 통합 fix.
  async getAccountConsoleURL(): Promise<string> {
    const config = await this.getRuntimeOIDCConfig();
    const issuer = config.oidcIssuerURL?.trim() ?? "";
    if (!issuer) {
      return "";
    }
    return `${issuer.replace(/\/$/, "")}/account/`;
  }

  private async getDiscovery(): Promise<OIDCDiscoveryDocument> {
    if (this.discoveryDoc) {
      return this.discoveryDoc;
    }

    const runtimeConfig = await this.getRuntimeOIDCConfig();
    const discoveryURL = `${runtimeConfig.oidcIssuerURL}/.well-known/openid-configuration`;
    const response = await fetch(discoveryURL, { method: "GET", cache: "no-store" });
    if (!response.ok) {
      throw new Error("OIDC discovery failed");
    }
    this.discoveryDoc = await response.json() as OIDCDiscoveryDocument;
    return this.discoveryDoc;
  }

}

export const authService = AuthService.getInstance();
