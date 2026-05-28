"use client";

import { tokenStore } from "@/lib/auth/token-store";
import { triggerSessionExpired } from "@/lib/auth/session-death";

// proactive token refresh — access token 만료 ~60s 전에 자동 `refreshTokens()`.
// 만료 자체를 사전 예방 (reactive 401 → refresh 패턴의 사각지대 해소).
//
// 사용:
//   - 앱 부트 시 1회 `initRefreshScheduler(refreshFn)` 호출 (AuthGuard mount).
//   - 이후 tokenStore.save (로그인/갱신 시) 가 자동으로 스케줄 재계산.
//   - tokenStore.clear (세션 종료) 가 자동으로 타이머 취소.

// access token 만료 X 초 전에 갱신 시도. Keycloak default 5분 / 60s 버퍼 = ~80% 시점.
const REFRESH_BUFFER_SECONDS = 60;
// 최소 1초는 두고 스케줄링 (즉시 호출로 인한 동기 loop 회피).
const MIN_SCHEDULE_MS = 1000;

let timer: number | null = null;
let refreshFn: (() => Promise<void>) | null = null;
let unsubscribe: (() => void) | null = null;

/**
 * 앱 부트 시 1회 호출. refreshFn 은 authService.refreshTokens 를 감싼 콜백.
 * (idempotent: 다시 호출하면 기존 구독 해제 후 재설정.)
 */
export function initRefreshScheduler(fn: () => Promise<void>): void {
  if (typeof window === "undefined") return;
  // 재호출 시 기존 구독·타이머 정리.
  if (unsubscribe) {
    unsubscribe();
    unsubscribe = null;
  }
  if (timer !== null) {
    window.clearTimeout(timer);
    timer = null;
  }
  refreshFn = fn;
  // tokenStore 의 expires_at 변경(save/clear) 마다 재스케줄링.
  unsubscribe = tokenStore.subscribeExpiryChange(() => reschedule());
  // 초기 상태 (F5 직후 로딩된 토큰) 도 즉시 스케줄.
  reschedule();
}

/**
 * 현재 tokenStore 상태 기준으로 타이머 재계산. 외부에서 직접 호출할 일 없으나
 * 테스트용/디버깅용 export.
 */
export function reschedule(): void {
  if (typeof window === "undefined") return;
  if (timer !== null) {
    window.clearTimeout(timer);
    timer = null;
  }
  if (!refreshFn) return;
  const expiresAt = tokenStore.getExpiresAt();
  if (expiresAt === null) {
    // 토큰 없음 → 스케줄 불필요.
    return;
  }
  const refreshAt = expiresAt - REFRESH_BUFFER_SECONDS * 1000;
  const delay = Math.max(MIN_SCHEDULE_MS, refreshAt - Date.now());
  timer = window.setTimeout(() => {
    timer = null;
    void runRefresh();
  }, delay);
}

async function runRefresh(): Promise<void> {
  if (!refreshFn) return;
  // refresh 호출 시점에 토큰 없으면 (clear 됨) 스킵.
  if (tokenStore.getAccessToken() === null) return;
  try {
    await refreshFn();
    // refreshFn 성공 시 tokenStore.save 가 호출되어 subscribeExpiryChange →
    // reschedule 가 자동 실행됨. 별도 호출 불필요.
  } catch (err) {
    console.warn("[refresh-scheduler] proactive refresh failed", err);
    triggerSessionExpired();
  }
}

/** 테스트용 — 스케줄러 완전 정지(타이머/구독 모두 해제). */
export function _resetRefreshScheduler(): void {
  if (timer !== null && typeof window !== "undefined") {
    window.clearTimeout(timer);
  }
  if (unsubscribe) {
    unsubscribe();
  }
  timer = null;
  refreshFn = null;
  unsubscribe = null;
}
