import { tokenStore } from "@/domain/auth-session/service/token-store";
import { triggerSessionExpired } from "@/domain/auth-session/service/session-death";
import { refreshAccessToken } from "@/domain/auth-session/service/refresh";
import { API_BASE_URL } from "@/shared/config/endpoints";

export class ApiError extends Error {
  constructor(public status: number, public payload: unknown, message: string) {
    super(message);
    this.name = "ApiError";
  }
}

type JsonObject = Record<string, unknown>;

function isJsonObject(value: unknown): value is JsonObject {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

// ADR-0024 §6 carve 3 extension:
// access token 갱신은 `lib/auth/refresh.ts` 의 `refreshAccessToken()` 단일 진입점을
// 통해 처리한다 (proactive scheduler / realtime fetchTicket 과 같은 single-flight
// mutex 공유 — #388 codex P1 정합). 결과 enum 으로 transient / auth_failed 를 구분해
// 세션 사망 결정도 명확히 한다 (#388 codex P2).

export async function apiClient<T>(
  method: string,
  path: string,
  body?: unknown,
  options?: { headers?: Record<string, string> }
): Promise<T> {
  const headers: Record<string, string> = {};
  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
  }

  // Inject Bearer token if available
  const token = tokenStore.getAccessToken();
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  // Inject custom headers if provided
  if (options?.headers) {
    Object.assign(headers, options.headers);
  }

  // Prepend same-origin API_BASE_URL (basePath prefix) if path is relative /api/v1/*
  const resolvedUrl = path.startsWith("/api/") ? `${API_BASE_URL}${path}` : path;

  let response = await fetch(resolvedUrl, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  // 401 핸들링: 인증 시도가 있었던(token 또는 refresh token 보유) 경우에만 refresh 시도.
  // (anon endpoint 호출의 401 은 회복 대상 아님.)
  if (response.status === 401 && (token || tokenStore.getRefreshToken())) {
    const outcome = await refreshAccessToken();
    if (outcome.kind === "ok") {
      const refreshedAccessToken = tokenStore.getAccessToken();
      if (refreshedAccessToken) {
        headers["Authorization"] = `Bearer ${refreshedAccessToken}`;
        response = await fetch(resolvedUrl, {
          method,
          headers,
          body: body !== undefined ? JSON.stringify(body) : undefined,
        });
      }
    } else if (outcome.kind === "auth_failed") {
      // refresh_token 거부 / 부재 — 세션 사망. /login 으로 hard nav (idempotent).
      // 본 호출은 그대로 401 throw — 진행 중인 caller 는 일시적으로 ApiError 받으나
      // location.assign 이 비동기로 화면을 전환하므로 무해.
      triggerSessionExpired();
    }
    // transient_failed: 세션을 죽이지 않음. 401 그대로 propagate → caller 가 결정
    // (다음 호출 시점에 다시 refresh 시도되어 자연 회복 가능).
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
