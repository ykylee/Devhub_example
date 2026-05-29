"use client";

import { tokenStore } from "@/domain/auth-session/service/token-store";
import { API_BASE_URL, OIDC_ISSUER_URL } from "@/shared/config/endpoints";

// access token 갱신의 **단일 진입점**.
// - apiClient (reactive 401 path) / refresh-scheduler (proactive timer) /
//   realtime.fetchTicket (WS handshake 401) 가 모두 본 함수를 거친다.
// - 단일 single-flight 가드(`inflight`) 로 동시 호출 시 같은 promise 공유 →
//   Keycloak `Refresh Token Max Reuse = 0` 환경에서 동일 refresh_token 이 두 번
//   사용돼 invalid_grant 가 터지던 race (#388 codex P1) 차단.
// - 결과를 `ok` / `auth_failed`(refresh_token 거부, 세션 사망 신호) /
//   `transient_failed`(네트워크 / IdP 5xx / parse 오류 등 재시도 가능) 로 분리 →
//   호출 측이 transient 일 때 세션을 죽이지 않도록 결정 (#388 codex P2).

export type RefreshOutcome =
  | { kind: "ok" }
  | { kind: "auth_failed"; reason: string }
  | { kind: "transient_failed"; reason: string };

const OIDC_CLIENT_ID = process.env.NEXT_PUBLIC_OIDC_CLIENT_ID ?? "devhub-frontend";

// 토큰 엔드포인트 해석은 1회 캐시. (Docker 배포 시 runtime-config 우회로 issuer 가
// 늦게 알려지는 케이스 호환 — 기존 api-client.ts 의 동작과 동일.)
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

let inflight: Promise<RefreshOutcome> | null = null;

export async function refreshAccessToken(): Promise<RefreshOutcome> {
  if (inflight) return inflight;
  inflight = doRefresh().finally(() => {
    inflight = null;
  });
  return inflight;
}

async function doRefresh(): Promise<RefreshOutcome> {
  const refreshToken = tokenStore.getRefreshToken();
  if (!refreshToken) {
    // 갱신 수단 부재 — 인증 종료.
    return { kind: "auth_failed", reason: "no_refresh_token" };
  }
  const tokenEndpoint = await resolveTokenEndpoint();
  if (!tokenEndpoint) {
    // OIDC issuer 미해석 — 일시 (runtime-config fetch 실패 등).
    return { kind: "transient_failed", reason: "no_token_endpoint" };
  }

  let response: Response;
  try {
    response = await fetch(tokenEndpoint, {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({
        grant_type: "refresh_token",
        refresh_token: refreshToken,
        client_id: OIDC_CLIENT_ID,
      }).toString(),
    });
  } catch (err) {
    console.warn("[refresh] network error", err);
    return { kind: "transient_failed", reason: "network_error" };
  }

  if (!response.ok) {
    const reason = `http_${response.status}`;
    console.warn("[refresh] token endpoint %s", reason);
    // 4xx = refresh_token 자체 거부 (invalid_grant, expired 등) → 세션 사망 신호.
    // 5xx = IdP 일시 장애 → 재시도 가능.
    return response.status >= 500
      ? { kind: "transient_failed", reason }
      : { kind: "auth_failed", reason };
  }

  try {
    const tokens = (await response.json()) as {
      access_token: string;
      refresh_token?: string;
      id_token?: string;
      expires_in: number;
      token_type: string;
    };
    tokenStore.save(tokens);
    return { kind: "ok" };
  } catch (err) {
    console.warn("[refresh] response parse error", err);
    return { kind: "transient_failed", reason: "parse_error" };
  }
}

/** 테스트용 — 캐시된 token endpoint 와 in-flight promise 리셋. */
export function _resetRefreshCoordinator(): void {
  cachedTokenEndpoint = undefined;
  inflight = null;
}
