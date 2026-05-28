"use client";

import type { TokenResponse } from "@/domain/auth-session/service/auth.service";

const ACCESS_TOKEN_KEY = "devhub_access_token";
const REFRESH_TOKEN_KEY = "devhub_refresh_token";
const ID_TOKEN_KEY = "devhub_id_token";
// access token 의 절대 만료 시각(ms since epoch). proactive refresh scheduler 가
// `expires_in` 으로부터 계산해 저장하고, 페이지 로드(F5) 이후 일정 스케줄링에 사용.
const EXPIRES_AT_KEY = "devhub_access_token_expires_at";

// 동일 탭 라이프사이클 내 'expires_at 변경' 을 listener 에 알림 (refresh scheduler 가 구독).
type ExpiryListener = (expiresAt: number | null) => void;

class TokenStore {
  private accessToken: string | null = null;
  private refreshToken: string | null = null;
  private idToken: string | null = null;
  private expiresAt: number | null = null;
  private listeners = new Set<ExpiryListener>();

  // null-sentinel 패턴: 각 필드가 null 이면 sessionStorage 에서 1회 hydrate.
  // clear() 가 모두 null 로 리셋하므로 그 뒤 외부에서 sessionStorage 를 직접 설정한
  // 시나리오(테스트)도 자연스럽게 반영된다.
  private ensureLoaded() {
    if (typeof window === "undefined") return;
    if (this.accessToken === null) {
      this.accessToken = sessionStorage.getItem(ACCESS_TOKEN_KEY);
    }
    if (this.refreshToken === null) {
      this.refreshToken = sessionStorage.getItem(REFRESH_TOKEN_KEY);
    }
    if (this.idToken === null) {
      this.idToken = sessionStorage.getItem(ID_TOKEN_KEY);
    }
    if (this.expiresAt === null) {
      const raw = sessionStorage.getItem(EXPIRES_AT_KEY);
      const parsed = raw ? Number(raw) : null;
      this.expiresAt = parsed !== null && Number.isFinite(parsed) ? parsed : null;
    }
  }

  getAccessToken(): string | null {
    this.ensureLoaded();
    return this.accessToken;
  }

  getRefreshToken(): string | null {
    this.ensureLoaded();
    return this.refreshToken;
  }

  // id_token is held so RP-initiated logout can pass it as id_token_hint to
  // Hydra /oauth2/sessions/logout. Without it Hydra cannot identify which
  // login session to terminate and the SSO cookie remains valid.
  getIdToken(): string | null {
    this.ensureLoaded();
    return this.idToken;
  }

  save(tokens: TokenResponse) {
    if (typeof window === "undefined") return;
    this.accessToken = tokens.access_token;
    this.refreshToken = tokens.refresh_token ?? null;
    this.idToken = tokens.id_token ?? null;
    // expires_in 은 access token 의 남은 수명(초). 절대 시각(ms) 으로 변환해 저장 —
    // F5 이후에도 refresh scheduler 가 남은 시간을 계산할 수 있게 한다.
    this.expiresAt = Number.isFinite(tokens.expires_in) && tokens.expires_in > 0
      ? Date.now() + tokens.expires_in * 1000
      : null;
    sessionStorage.setItem(ACCESS_TOKEN_KEY, tokens.access_token);
    if (tokens.refresh_token) {
      sessionStorage.setItem(REFRESH_TOKEN_KEY, tokens.refresh_token);
    } else {
      sessionStorage.removeItem(REFRESH_TOKEN_KEY);
    }
    if (tokens.id_token) {
      sessionStorage.setItem(ID_TOKEN_KEY, tokens.id_token);
    } else {
      sessionStorage.removeItem(ID_TOKEN_KEY);
    }
    if (this.expiresAt !== null) {
      sessionStorage.setItem(EXPIRES_AT_KEY, String(this.expiresAt));
    } else {
      sessionStorage.removeItem(EXPIRES_AT_KEY);
    }
    this.notifyExpiryListeners();
  }

  clear() {
    if (typeof window === "undefined") return;
    this.accessToken = null;
    this.refreshToken = null;
    this.idToken = null;
    this.expiresAt = null;
    sessionStorage.removeItem(ACCESS_TOKEN_KEY);
    sessionStorage.removeItem(REFRESH_TOKEN_KEY);
    sessionStorage.removeItem(ID_TOKEN_KEY);
    sessionStorage.removeItem(EXPIRES_AT_KEY);
    this.notifyExpiryListeners();
  }

  // access token 의 절대 만료 시각(ms since epoch). 없으면 null.
  getExpiresAt(): number | null {
    this.ensureLoaded();
    return this.expiresAt;
  }

  // expires_at 변경 시 호출되는 listener 등록 — refresh scheduler 가 재스케줄링용.
  // 반환된 unsubscribe 함수로 해제.
  subscribeExpiryChange(listener: ExpiryListener): () => void {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  }

  private notifyExpiryListeners() {
    for (const listener of this.listeners) {
      try {
        listener(this.expiresAt);
      } catch (err) {
        console.warn("[tokenStore] expiry listener error", err);
      }
    }
  }
}

export const tokenStore = new TokenStore();
