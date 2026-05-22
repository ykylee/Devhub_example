import { describe, expect, it } from "vitest";

import { ERROR_MESSAGES, resolveErrorMessage } from "./page";

// ADR-0020 sub-carve F (sprint claude/work_260522-adr-0020-subcarve-f-login)
// — `?error=<code>` query 가 ERROR_MESSAGES 매핑에 맞춰 사용자 메시지로
// 변환되는지 회귀 가드. AuthGuard 의 401 fallback (`session_expired`) + 기타
// fallback (`login_failed`) + 사내 anchor (`unauthorized`) 3 케이스 + provider
// `error_description` propagate + 미매핑 fallback 까지 cover.

describe("resolveErrorMessage", () => {
  it("returns null when no code nor description is provided", () => {
    expect(resolveErrorMessage(null, null)).toBeNull();
  });

  it("prefers the explicit error_description from the provider", () => {
    expect(resolveErrorMessage("login_failed", "kid_mismatch_retry_required")).toBe(
      "kid_mismatch_retry_required",
    );
  });

  it("maps session_expired to the user-facing message", () => {
    expect(resolveErrorMessage("session_expired", null)).toBe(
      ERROR_MESSAGES.session_expired,
    );
  });

  it("maps login_failed to the user-facing message", () => {
    expect(resolveErrorMessage("login_failed", null)).toBe(
      ERROR_MESSAGES.login_failed,
    );
  });

  it("maps unauthorized to the user-facing message", () => {
    expect(resolveErrorMessage("unauthorized", null)).toBe(
      ERROR_MESSAGES.unauthorized,
    );
  });

  it("falls back to the raw code when no mapping exists and no description was provided", () => {
    expect(resolveErrorMessage("provider_specific_code", null)).toBe(
      "provider_specific_code",
    );
  });
});
