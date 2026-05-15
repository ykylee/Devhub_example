"use client";

import { AuthenticatedActor, useStore } from "../store";
import { identityService } from "./identity.service";
import { tokenStore } from "@/lib/auth/token-store";
import { consumeVerifier, createPkceState } from "@/lib/auth/pkce";
import { killKratosSession, performKratosBrowserLogout } from "@/lib/auth/kratos-logout";
import { apiClient } from "./api-client";

// SignUpPayload mirrors backend_api_contract.md §11.5.2 (POST /api/v1/auth/signup).
// All four fields are required; HR DB validation runs server-side.
export interface SignUpPayload {
  name: string;
  system_id: string;
  employee_id: string;
  password: string;
}

export interface SignUpResponse {
  status: "created";
  data: {
    user_id: string;
    kratos_id: string;
    department: string;
    message: string;
  };
}

import { OIDC_AUTH_URL, HYDRA_PUBLIC_BASE, OIDC_REDIRECT_URI as OIDC_REDIRECT_URI_DEFAULT } from "../config/endpoints";

const OIDC_CLIENT_ID = process.env.NEXT_PUBLIC_OIDC_CLIENT_ID ?? "devhub-frontend";
const OIDC_REDIRECT_URI = typeof window !== "undefined"
  ? `${window.location.origin}/auth/callback`
  : OIDC_REDIRECT_URI_DEFAULT;
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
    const { state, codeChallenge, codeChallengeMethod } = await createPkceState();

    const url = new URL(OIDC_AUTH_URL);
    url.searchParams.set("client_id", OIDC_CLIENT_ID);
    url.searchParams.set("response_type", "code");
    url.searchParams.set("redirect_uri", OIDC_REDIRECT_URI);
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
   * Header Sign Out flow.
   *
   * Codex review (PR #46) showed the prior path left Hydra's SSO cookie
   * intact. Since auth_login.go fast-paths hydraReq.Skip=true, the next
   * /login could silently re-authenticate without a Kratos credential,
   * making Sign Out cosmetic. We now drive Hydra RP-initiated logout
   * (id_token_hint -> /oauth2/sessions/logout) which lands the browser at
   * /auth/logout?logout_challenge=... where completeRPInitiatedLogout
   * finishes both Hydra accept and Kratos cookie kill.
   *
   * Fallback (no id_token persisted, e.g., legacy session): same Kratos
   * browser logout as before so the user is still bounced to /.
   */
  public async logout(): Promise<void> {
    const refreshToken = tokenStore.getRefreshToken();
    const idToken = tokenStore.getIdToken();

    // Best-effort backend revoke. We do not await: the Hydra navigation
    // below must happen even if revoke is slow or fails. The RP-initiated
    // path also revokes (with the same token) as a second defence layer.
    if (refreshToken) {
      void fetch("/api/v1/auth/logout", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          refresh_token: refreshToken,
          client_id: OIDC_CLIENT_ID,
        }),
      }).catch((err) => {
        console.warn("[AuthService] backend logout call failed (continuing)", err);
      });
    }

    if (idToken) {
      const url = new URL(`${HYDRA_PUBLIC_BASE}/oauth2/sessions/logout`);
      url.searchParams.set("id_token_hint", idToken);
      url.searchParams.set("post_logout_redirect_uri", `${window.location.origin}/`);
      // Clear local state before navigating; completeRPInitiatedLogout will
      // re-clear when /auth/logout loads, but doing it here keeps any
      // intermediate state (back button, devtools) clean.
      tokenStore.clear();
      useStore.getState().clearActor();
      window.location.assign(url.toString());
      return;
    }

    // Fallback: no id_token to drive Hydra logout. Kill Kratos cookie via
    // navigation so at least one half of the SSO state is gone.
    tokenStore.clear();
    useStore.getState().clearActor();
    await performKratosBrowserLogout("/");
  }

  /**
   * RP-initiated logout entry point (Hydra urls.logout target /auth/logout).
   * Hydra has redirected the browser here with a logout_challenge; backend
   * accepts the challenge (and revokes the refresh token if we still hold
   * one — Codex review P2), Kratos cookie is killed via fetch (best-effort,
   * cannot navigate twice), then we follow Hydra's redirect_to.
   */
  public async completeRPInitiatedLogout(challenge: string): Promise<void> {
    const refreshToken = tokenStore.getRefreshToken();

    let redirectTo = "/";
    try {
      const body: Record<string, string> = { logout_challenge: challenge };
      if (refreshToken) {
        body.refresh_token = refreshToken;
        body.client_id = OIDC_CLIENT_ID;
      }
      const res = await fetch("/api/v1/auth/logout", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
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

    // Best-effort Kratos cookie kill before following Hydra's redirect.
    await killKratosSession();

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

  /**
   * Self-service Sign Up (RM-M3-01). POSTs to /api/v1/auth/signup; the
   * backend verifies the (name, system_id, employee_id) triple against the
   * HR DB and creates Kratos identity + DevHub user. See
   * backend_api_contract.md §11.5.2 for the response/error matrix.
   */
  public async signup(payload: SignUpPayload): Promise<SignUpResponse> {
    return apiClient<SignUpResponse>("POST", "/api/v1/auth/signup", payload);
  }
}

export const authService = AuthService.getInstance();
