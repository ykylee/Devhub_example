import { tokenStore } from "@/lib/auth/token-store";
import { triggerSessionExpired } from "@/lib/auth/session-death";
import { API_BASE_URL, OIDC_ISSUER_URL } from "@/lib/config/endpoints";

export class ApiError extends Error {
  constructor(public status: number, public payload: unknown, message: string) {
    super(message);
    this.name = "ApiError";
  }
}

type JsonObject = Record<string, unknown>;
interface TokenResponse {
  access_token: string;
  refresh_token?: string;
  id_token?: string;
  expires_in: number;
  token_type: string;
}

function isJsonObject(value: unknown): value is JsonObject {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

// ADR-0024 §6 carve 3 extension: resolve the OIDC token endpoint.
// Prefers the build-time NEXT_PUBLIC_OIDC_ISSUER_URL; falls back to the
// server-side /api/runtime-config for Docker deployments where OIDC issuer
// is injected at container runtime.
const OIDC_CLIENT_ID = process.env.NEXT_PUBLIC_OIDC_CLIENT_ID ?? "devhub-frontend";

// Cached token endpoint URL. undefined = not yet resolved, null = resolution
// failed, string = usable URL. Cache avoids a runtime-config fetch on every
// 401 when OIDC_ISSUER_URL is not baked at build time (Docker deployments).
let cachedTokenEndpoint: string | null | undefined;

async function resolveTokenEndpoint(): Promise<string | null> {
  if (cachedTokenEndpoint !== undefined) return cachedTokenEndpoint;

  const issuer = OIDC_ISSUER_URL;
  if (issuer) {
    cachedTokenEndpoint = `${issuer}/protocol/openid-connect/token`;
    return cachedTokenEndpoint;
  }

  try {
    const resp = await fetch(`${API_BASE_URL}/api/runtime-config`, { cache: "no-store" });
    if (!resp.ok) {
      cachedTokenEndpoint = null;
      return null;
    }
    const body = (await resp.json()) as { oidc_issuer_url?: string };
    const url = body.oidc_issuer_url?.trim();
    if (!url) {
      cachedTokenEndpoint = null;
      return null;
    }
    cachedTokenEndpoint = `${url}/protocol/openid-connect/token`;
    return cachedTokenEndpoint;
  } catch {
    cachedTokenEndpoint = null;
    return null;
  }
}

// ADR-0024 §6 carve 3 extension: exchange the stored refresh_token for a new
// access_token. Used by apiClient's 401 interceptor to transparently recover
// without forcing the user to re-authenticate.
// Guard against parallel refresh storms — if N requests all get 401 at the
// same time, only one refresh call hits the IdP; the rest join the in-flight
// promise and return the same result.
let inflightRefresh: Promise<boolean> | null = null;

async function attemptTokenRefresh(): Promise<boolean> {
  const refreshToken = tokenStore.getRefreshToken();
  if (!refreshToken) {
    // 호출 측이 인증을 시도했고 401 인데 refresh token 도 없다 = 회복 수단 부재.
    // 좀비 상태 방지를 위해 세션 종료(/login 으로 hard nav).
    triggerSessionExpired();
    return false;
  }

  if (inflightRefresh) return inflightRefresh;

  inflightRefresh = doRefresh(refreshToken);
  const result = await inflightRefresh;
  inflightRefresh = null;
  return result;
}

async function doRefresh(refreshToken: string): Promise<boolean> {
  const tokenEndpoint = await resolveTokenEndpoint();
  if (!tokenEndpoint) return false;

  try {
    const response = await fetch(tokenEndpoint, {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({
        grant_type: "refresh_token",
        refresh_token: refreshToken,
        client_id: OIDC_CLIENT_ID,
      }).toString(),
    });
    if (!response.ok) {
      console.warn("[apiClient] token refresh failed (HTTP %d); session_expired", response.status);
      // triggerSessionExpired 내부에서 tokenStore.clear + /login 리다이렉트.
      triggerSessionExpired();
      return false;
    }
    const tokens = (await response.json()) as TokenResponse;
    tokenStore.save(tokens);
    return true;
  } catch (err) {
    // 네트워크 오류는 transient (DNS / 일시 단절 등) — 세션 사망으로 단정하지 않는다.
    // 호출 측이 401 throw 하면 다음 호출에서 다시 refresh 시도.
    console.warn("[apiClient] token refresh network error", err);
    return false;
  }
}

export async function apiClient<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = {};
  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
  }

  // Inject Bearer token if available
  const token = tokenStore.getAccessToken();
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  // Prepend same-origin API_BASE_URL (basePath prefix) if path is relative /api/v1/*
  const resolvedUrl = path.startsWith("/api/") ? `${API_BASE_URL}${path}` : path;

  let response = await fetch(resolvedUrl, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  // ADR-0024 §6 carve 3 extension:
  // - 기존: access token 이 실린 요청에서만 refresh-then-retry.
  // - 보강: hard refresh 직후처럼 access token header 가 비어도 refresh token 이
  //   남아 있으면 1회 복구 시도.
  if (response.status === 401 && (token || tokenStore.getRefreshToken())) {
    const refreshed = await attemptTokenRefresh();
    if (refreshed) {
      const refreshedAccessToken = tokenStore.getAccessToken();
      if (refreshedAccessToken) {
        headers["Authorization"] = `Bearer ${refreshedAccessToken}`;
        response = await fetch(resolvedUrl, {
          method,
          headers,
          body: body !== undefined ? JSON.stringify(body) : undefined,
        });
      }
    }
  }

  let parsed: unknown = null;
  const text = await response.text();
  if (text.length > 0) {
    try {
      parsed = JSON.parse(text);
    } catch {
      parsed = { raw: text };
    }
  }

  if (!response.ok) {
    const errMessage = isJsonObject(parsed) && typeof parsed.error === "string" 
      ? parsed.error 
      : `HTTP ${response.status}`;
    throw new ApiError(response.status, parsed, errMessage);
  }

  return parsed as T;
}
