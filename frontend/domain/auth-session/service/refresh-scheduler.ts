"use client";

import { tokenStore } from "@/domain/auth-session/service/token-store";
import { triggerSessionExpired } from "@/domain/auth-session/service/session-death";
import type { RefreshOutcome } from "@/domain/auth-session/service/refresh";

// proactive token refresh — access token 만료 ~60s 전에 자동 갱신 시도.
// 만료 자체를 사전 예방 (reactive 401 → refresh 패턴의 사각지대 해소).
//
// **단일 mutex 공유 (#388 codex P1)**: refresh 실 호출은 `lib/auth/refresh.ts` 의
// `refreshAccessToken` 을 통해 — apiClient(reactive 401) / realtime.fetchTicket 등
// 다른 경로와 single-flight 가드를 공유한다. AuthGuard mount 에서 그 함수를
// performer 로 등록.
//
// **transient vs auth_failed 구분 (#388 codex P2)**: performer 가 RefreshOutcome 을
// 반환. auth_failed 만 `triggerSessionExpired()` 로 세션 사망 처리하고, transient 는
// 다음 reactive 401 또는 만료 후 재시도로 자연 회복하도록 둔다.

// access token 만료 X 초 전에 갱신 시도. Keycloak default 5분 / 60s 버퍼 = ~80% 시점.
const REFRESH_BUFFER_SECONDS = 60;
// 최소 1초는 두고 스케줄링 (즉시 호출로 인한 동기 loop 회피).
const MIN_SCHEDULE_MS = 1000;
// transient_failed 지수 backoff (codex P2 #405 정합) — 정상 시점 (만료 60s 전) 에
// transient 발생 시 단순 reschedule() 은 이미 과거 refreshAt 으로 MIN_SCHEDULE_MS
// fallback → 매초 retry loop 였음. base 5s × 2^n → max 60s (5/10/20/40/60/60/...).
const TRANSIENT_RETRY_BASE_MS = 5_000;
const TRANSIENT_RETRY_MAX_MS = 60_000;

type Performer = () => Promise<RefreshOutcome>;

let timer: ReturnType<typeof setTimeout> | null = null;
let performer: Performer | null = null;
let unsubscribe: (() => void) | null = null;
// transient 연속 실패 카운터 — ok 분기에서 reset (tokenStore.save → subscribeExpiryChange
// → reschedule 자동 호출되며 그 시점에 0 으로 초기화).
let transientFailureCount = 0;

/**
 * 앱 부트 시 1회 호출. fn 은 `refreshAccessToken` (lib/auth/refresh.ts) — single-flight
 * 가드 + RefreshOutcome 반환을 보장한다.
 * (idempotent: 다시 호출하면 기존 구독 해제 후 재설정.)
 */
export function initRefreshScheduler(fn: Performer): void {
  if (typeof window === "undefined") return;
  if (unsubscribe) {
    unsubscribe();
    unsubscribe = null;
  }
  if (timer !== null) {
    clearTimeout(timer);
    timer = null;
  }
  performer = fn;
  unsubscribe = tokenStore.subscribeExpiryChange(() => reschedule());
  // 초기 상태 (F5 직후 로딩된 토큰) 도 즉시 스케줄.
  reschedule();
}

/**
 * 현재 tokenStore 상태 기준으로 타이머 재계산. ok 분기 (token refresh 성공) 또는
 * initRefreshScheduler 등 정상 진입점에서 호출되며 transient backoff counter 도 reset.
 */
export function reschedule(): void {
  if (typeof window === "undefined") return;
  if (timer !== null) {
    clearTimeout(timer);
    timer = null;
  }
  if (!performer) return;
  const expiresAt = tokenStore.getExpiresAt();
  if (expiresAt === null) {
    // 토큰 없음 → 스케줄 불필요.
    transientFailureCount = 0;
    return;
  }
  // tokenStore 변동 = 새 token → backoff reset.
  transientFailureCount = 0;
  const refreshAt = expiresAt - REFRESH_BUFFER_SECONDS * 1000;
  const delay = Math.max(MIN_SCHEDULE_MS, refreshAt - Date.now());
  timer = setTimeout(() => {
    timer = null;
    void runRefresh();
  }, delay);
}

async function runRefresh(): Promise<void> {
  if (!performer) return;
  // refresh 호출 시점에 토큰 없으면 (clear 됨) 스킵.
  if (tokenStore.getAccessToken() === null) return;
  let outcome: RefreshOutcome;
  try {
    outcome = await performer();
  } catch (err) {
    // performer 는 보통 throw 하지 않으나, 예외적 경우엔 transient 로 처리해 세션을
    // 죽이지 않는다 (#388 codex P2 정신 — 모호한 에러는 회복 가능 쪽으로 분류).
    console.warn("[refresh-scheduler] performer threw (treated as transient)", err);
    outcome = { kind: "transient_failed", reason: "performer_threw" };
  }
  if (outcome.kind === "auth_failed") {
    console.warn("[refresh-scheduler] proactive refresh auth_failed:", outcome.reason);
    triggerSessionExpired();
  } else if (outcome.kind === "transient_failed") {
    // P0 fix (router-idle-timeout-analysis 2026-05-28) — transient 시 후속 타이머 미설정이
    // root cause 였음. 단순 `reschedule()` 호출은 codex P2 (#405) 지적대로 정상 시점
    // (만료 60s 전) 발생 시 refreshAt 이 이미 과거 → MIN_SCHEDULE_MS (1초) fallback →
    // **매초 retry loop**. 지수 backoff 로 정정: base 5s × 2^n → max 60s. transient
    // 연속 실패해도 IdP/network outage 동안 tab 당 분당 ~1회 부담만. ok 분기에서
    // counter reset (reschedule 안에서). 분석 SoT: `docs/analysis/2026-05-28-router-
    // idle-timeout-analysis.md` Fix 6.1 + 본 PR codex P2.
    transientFailureCount += 1;
    const backoffMs = Math.min(
      TRANSIENT_RETRY_MAX_MS,
      TRANSIENT_RETRY_BASE_MS * Math.pow(2, transientFailureCount - 1),
    );
    console.warn(
      `[refresh-scheduler] transient_failed (attempt ${transientFailureCount}):`,
      outcome.reason,
      `— retrying in ${backoffMs}ms`,
    );
    timer = setTimeout(() => {
      timer = null;
      void runRefresh();
    }, backoffMs);
  }
  // ok: tokenStore.save 가 자동으로 subscribeExpiryChange → reschedule 호출 (backoff reset).
}

/** 테스트용 — 스케줄러 완전 정지(타이머/구독/backoff counter 모두 reset). */
export function _resetRefreshScheduler(): void {
  if (timer !== null) {
    clearTimeout(timer);
  }
  if (unsubscribe) {
    unsubscribe();
  }
  timer = null;
  performer = null;
  unsubscribe = null;
  transientFailureCount = 0;
}
