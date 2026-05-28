"use client";

import { tokenStore } from "@/lib/auth/token-store";
import { BASE_PATH } from "@/lib/config/endpoints";

// 세션 사망 처리의 단일 진입점.
// - apiClient.doRefresh 실패 (refresh token 만료/거부) → 호출
// - refresh-scheduler 의 proactive refresh 실패 → 호출
// - 기타 인증 복구 불가 상태 → 호출
//
// 호출 시: tokenStore 비움 + /login?error=session_expired 로 hard nav.
// idempotent: 동일 페이지 라이프사이클 내 다중 호출 무시 (location.assign 중복 회피).

let triggered = false;

export function triggerSessionExpired(reason: string = "session_expired"): void {
  if (typeof window === "undefined") return;
  if (triggered) return;
  triggered = true;

  try {
    tokenStore.clear();
  } catch (err) {
    console.warn("[session-death] tokenStore.clear failed", err);
  }

  // 이미 /login 이면 다시 안 보냄 (redirect loop 방지)
  const pathname = window.location.pathname;
  const loginPath = `${BASE_PATH}/login`;
  if (pathname === loginPath || pathname.startsWith(`${loginPath}/`) || pathname.startsWith(`${loginPath}?`)) {
    return;
  }

  // window.location.assign 으로 hard nav — React 트리/라우터 transition 상태 초기화 효과.
  // (next/navigation router.replace 는 비-React 컨텍스트에서 호출 불가.)
  const target = `${BASE_PATH}/login?error=${encodeURIComponent(reason)}`;
  window.location.assign(target);
}

// 테스트용 — 모듈 단위 idempotent 가드 리셋.
export function _resetSessionExpiredGuard(): void {
  triggered = false;
}
