import { tokenStore } from "@/lib/auth/token-store";
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

async function resolveTokenEndpoint(): Promise<string | null> {
  const issuer = OIDC_ISSUER_URL;
  if (issuer) return `${issuer}/protocol/openid-connect/token`;

  try {
    const resp = await fetch(`${API_BASE_URL}/api/runtime-config`, { cache: "no-store" });
    if (!resp.ok) return null;
    const body = (await resp.json()) as { oidc_issuer_url?: string };
    const url = body.oidc_issuer_url?.trim();
    if (!url) return null;
    return `${url}/protocol/openid-connect/token`;
  } catch {
    return null;
  }
}

// ADR-0024 §6 carve 3 extension: exchange the stored refresh_token for a new
// access_token. Used by apiClient's 401 interceptor to transparently recover
// without forcing the user to re-authenticate.
async function attemptTokenRefresh(): Promise<boolean> {
  const refreshToken = tokenStore.getRefreshToken();
  if (!refreshToken) return false;

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
      tokenStore.clear();
      return false;
    }
    const tokens = (await response.json()) as TokenResponse;
    tokenStore.save(tokens);
    return true;
  } catch {
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

  // ADR-0024 §6 carve 3 extension: 401 + we had a token → attempt refresh + retry once.
  // Without this, an expired access_token causes an immediate redirect to /login?error=session_expired
  // even when a valid refresh_token is available in sessionStorage.
  if (response.status === 401 && token) {
    const refreshed = await attemptTokenRefresh();
    if (refreshed) {
      headers["Authorization"] = `Bearer ${tokenStore.getAccessToken()}`;
      response = await fetch(resolvedUrl, {
        method,
        headers,
        body: body !== undefined ? JSON.stringify(body) : undefined,
      });
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
