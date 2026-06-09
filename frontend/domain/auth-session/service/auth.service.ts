"use client";

import { AuthenticatedActor, useStore } from "@/lib/store";
import { identityService } from "@/domain/organization-management/service/identity.service";
import { tokenStore } from "@/domain/auth-session/service/token-store";
import { consumeVerifier, createPkceState } from "@/domain/auth-session/service/pkce";

import { BASE_PATH, OIDC_AUTH_URL, OIDC_ISSUER_URL, OIDC_REDIRECT_URI as OIDC_REDIRECT_URI_DEFAULT } from "@/shared/config/endpoints";

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
    const runtimeConfig = await this.getRuntimeOIDCConfig();
    const discovery = await this.getDiscovery();

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

  /**
   * Exchanges the stored refresh_token for a new access_token (and rotated
   * refresh_token if Keycloak issues one). ADR-0024 §6 carve 3 — used by
   * realtime service on WS handshake 401 to recover without forcing logout.
   * Throws on no refresh_token / network / 4xx from token endpoint.
   */
  public async refreshTokens(): Promise<TokenResponse> {
    const refreshToken = tokenStore.getRefreshToken();
    if (!refreshToken) throw new Error("no refresh_token available");
    const discovery = await this.getDiscovery();
    const response = await fetch(discovery.token_endpoint, {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({
        grant_type: "refresh_token",
        refresh_token: refreshToken,
        client_id: OIDC_CLIENT_ID,
      }).toString(),
    });
    if (!response.ok) {
      const err = await response.json().catch(() => ({} as Record<string, string>));
      throw new Error(err.error || `Token refresh failed (HTTP ${response.status})`);
    }
    const tokens = await response.json() as TokenResponse;
    tokenStore.save(tokens);
    return tokens;
  }

  // 백엔드 POST /api/v1/auth/logout 응답 status 에 따른 분기.
  //   204 No Content  : 정상 — OIDC end_session_endpoint 로 redirect (RP-initiated logout).
  //   401 Unauthorized : idempotent — 토큰이 이미 만료/무효. 동일 cleanup + OIDC redirect.
  //   502 Bad Gateway : backend 가 Keycloak 에 도달 못함 (N-8 hotfix 4차 이전 동작).
  //                     hotfix 4차 (issue #501, 2026-06-09) 으로 backend 가 502 대신
  //                     204 + audit revoke_status=unreachable 로 정합 변경됨. 본 분기
  //                     (line 155) 는 dead code (defensive guard) 로 유지. 미래에
  //                     backend 가 다시 502 를 반환할 경우 fallback 으로 동작.
  //   그 외 4xx/5xx   : defensive — toast (warning) + OIDC redirect (Keycloak 단독 logout 시도).
  // network throw     : fetch 자체가 실패 (CORS / network) — toast (warning) + OIDC redirect.
  // 본 분기는 issue #488 spec §"Frontend (Gemini — 후속, sprint -i)" 의 결정 권장값을
  // 그대로 구현 (stale-while-error 미적용, 정합 우선, 502 즉시 /login 강제).
  private async postBackendLogout(accessToken: string | null, refreshToken: string | null, idToken: string | null): Promise<"ok" | "expired" | "unreachable" | "error"> {
    try {
      const response = await fetch(`${BASE_PATH}/api/v1/auth/logout`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
        },
        body: JSON.stringify({
          refresh_token: refreshToken ?? undefined,
          id_token: idToken ?? undefined,
        }),
      });
      if (response.status === 204) return "ok";
      if (response.status === 401) return "expired";
      if (response.status === 502) return "unreachable";
      return "error";
    } catch (error) {
      console.warn("[AuthService] logout API call failed (network)", error);
      return "error";
    }
  }

  // cleanup + OIDC end_session_endpoint redirect. backend 가 502 를 반환한 경우
  // (즉 Keycloak 이 unreachable) 에는 OIDC 단계도 실패할 가능성이 높으므로 호출하지
  // 않고 `/login` 직접 redirect 로 fall back.
  private redirectToOIDCEndSession(
    discovery: OIDCDiscoveryDocument,
    runtimeConfig: { oidcIssuerURL: string },
    idToken: string | null,
  ): void {
    const endSessionEndpoint =
      discovery.end_session_endpoint ||
      `${runtimeConfig.oidcIssuerURL}/protocol/openid-connect/logout`;

    try {
      const url = new URL(endSessionEndpoint);
      url.searchParams.set("client_id", OIDC_CLIENT_ID);
      // post_logout_redirect 는 `/login` (BASE_PATH 포함). 이전엔 `/` 였는데, `/` 가
      // 대시보드 루트(보호 경로)라 AuthGuard whoAmI 가 401 을 받아 `/login?error=
      // session_expired` 로 다시 보내 "깨끗한 로그아웃" 대신 에러 화면이 표시되던
      // UX 버그(#387 ④) 정정. /login 직접 도착으로 clean logout 경로 확립.
      // (Keycloak Valid post-logout redirect URI 에 /devhub/login 등록 필요.)
      url.searchParams.set("post_logout_redirect_uri", `${window.location.origin}${BASE_PATH}/login`);
      if (idToken) {
        url.searchParams.set("id_token_hint", idToken);
      }
      window.location.assign(url.toString());
      return;
    } catch (error) {
      console.warn("[AuthService] logout redirect build failed", error);
      window.location.assign(`${BASE_PATH}/login`);
    }
  }

  public async logout(): Promise<void> {
    const accessToken = tokenStore.getAccessToken();
    const refreshToken = tokenStore.getRefreshToken();
    const idToken = tokenStore.getIdToken();

    // backend 가 access token revoke 를 시도하는 동안 frontend 도 token store 를 비우면
    // Bearer 헤더가 빈 채로 가서 middleware 가 401 로 단락시킨다. 따라서 cleanup 은
    // backend 응답을 받은 "이후" 로 미루고, status 분기는 backend 응답 기반으로 결정.
    const outcome = await this.postBackendLogout(accessToken, refreshToken, idToken);

    tokenStore.clear();
    useStore.getState().setIsLoggingOut(true);
    useStore.getState().clearActor();

    if (outcome === "unreachable") {
      // backend 가 Keycloak 에 도달 못함 → OIDC 도 unreachable 가능성. OIDC 단계 건너뛰고
      // toast 로 알린 뒤 /login 강제 redirect. 정합 우선 (#488 spec 권장).
      useStore.getState().addToast(
        "Sign-out service is temporarily unreachable. Your session has been cleared locally.",
        "error",
      );
      window.location.assign(`${BASE_PATH}/login`);
      return;
    }

    if (outcome === "error") {
      // 5xx/기타 4xx 또는 network throw — Keycloak logout 단독은 시도해보고, 결과에 따라
      // /login fallback. UX 우선.
      useStore.getState().addToast(
        "Sign-out encountered a non-fatal error. Continuing local cleanup.",
        "warning",
      );
    }
    // outcome === "ok" | "expired" : 204/401 — 일반 흐름. OIDC end_session_endpoint 로
    // RP-initiated logout 시도. (idempotent 401 도 OIDC logout 은 안전하게 시도 가능.)
    const runtimeConfig = await this.getRuntimeOIDCConfig();
    const discovery = await this.getDiscovery();
    this.redirectToOIDCEndSession(discovery, runtimeConfig, idToken);
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
      const response = await fetch(`${BASE_PATH}/api/runtime-config`, {
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
    try {
      const response = await fetch(discoveryURL, { method: "GET", cache: "no-store" });
      if (!response.ok) {
        throw new Error(`OIDC discovery failed: ${response.status}`);
      }
      this.discoveryDoc = await response.json() as OIDCDiscoveryDocument;
      return this.discoveryDoc;
    } catch (error) {
      // Some environments block cross-origin discovery fetch in the browser.
      // Fall back to issuer-derived endpoints so login can still begin.
      console.warn("[AuthService] discovery fetch failed, using issuer-derived endpoints", error);
      const issuer = runtimeConfig.oidcIssuerURL.replace(/\/+$/, "");
      this.discoveryDoc = {
        authorization_endpoint: runtimeConfig.oidcAuthURL || `${issuer}/protocol/openid-connect/auth`,
        token_endpoint: `${issuer}/protocol/openid-connect/token`,
        end_session_endpoint: `${issuer}/protocol/openid-connect/logout`,
      };
    }
    return this.discoveryDoc;
  }

}

export const authService = AuthService.getInstance();
