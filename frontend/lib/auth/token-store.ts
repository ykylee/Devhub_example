"use client";

import type { TokenResponse } from "@/lib/services/auth.service";

const ACCESS_TOKEN_KEY = "devhub_access_token";
const REFRESH_TOKEN_KEY = "devhub_refresh_token";

class TokenStore {
  private accessToken: string | null = null;
  private refreshToken: string | null = null;

  private ensureLoaded() {
    if (typeof window === "undefined") return;
    if (this.accessToken === null) {
      this.accessToken = sessionStorage.getItem(ACCESS_TOKEN_KEY);
    }
    if (this.refreshToken === null) {
      this.refreshToken = sessionStorage.getItem(REFRESH_TOKEN_KEY);
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

  save(tokens: TokenResponse) {
    if (typeof window === "undefined") return;
    this.accessToken = tokens.access_token;
    this.refreshToken = tokens.refresh_token ?? null;
    sessionStorage.setItem(ACCESS_TOKEN_KEY, tokens.access_token);
    if (tokens.refresh_token) {
      sessionStorage.setItem(REFRESH_TOKEN_KEY, tokens.refresh_token);
    } else {
      sessionStorage.removeItem(REFRESH_TOKEN_KEY);
    }
  }

  clear() {
    if (typeof window === "undefined") return;
    this.accessToken = null;
    this.refreshToken = null;
    sessionStorage.removeItem(ACCESS_TOKEN_KEY);
    sessionStorage.removeItem(REFRESH_TOKEN_KEY);
  }
}

export const tokenStore = new TokenStore();
