import { describe, expect, it } from "vitest";
import { ApiError } from "@/shared/api/api-client";
import { toUserErrorMessage } from "@/shared/utils/error-message";

describe("toUserErrorMessage", () => {
  it("returns the 5xx-aware Korean message for ApiError with status >= 500", () => {
    const error = new ApiError(500, { status: "rejected" }, "HTTP 500");

    expect(toUserErrorMessage(error, "fallback")).toBe(
      "서버 오류가 발생했습니다. 잠시 후 다시 시도해주세요.",
    );
  });

  it("returns the 5xx Korean message for any status >= 500 (e.g. 503)", () => {
    const error = new ApiError(503, null, "HTTP 503");

    expect(toUserErrorMessage(error, "fallback")).toBe(
      "서버 오류가 발생했습니다. 잠시 후 다시 시도해주세요.",
    );
  });

  // 2026-06-17 fix: backend platformStoreOrUnavailable (router.go) 가 503 + body
  // {code: "platform_store_unavailable"} 반환. toUserErrorMessage 가 payload.code
  // 매핑 → "Backend store is not initialized" 안내. generic 5xx 메시지보다 우선.
  it("returns Backend-store-not-initialized message for 503 + code: platform_store_unavailable", () => {
    const error = new ApiError(503, { code: "platform_store_unavailable", error: "..." }, "HTTP 503");

    expect(toUserErrorMessage(error, "fallback")).toContain("Backend store is not initialized");
  });

  it("returns generic 5xx message for 503 without code field (fallback)", () => {
    const error = new ApiError(503, null, "HTTP 503");

    expect(toUserErrorMessage(error, "fallback")).toBe(
      "서버 오류가 발생했습니다. 잠시 후 다시 시도해주세요.",
    );
  });

  it("returns the not-found Korean message for ApiError 404", () => {
    const error = new ApiError(404, { status: "rejected" }, "HTTP 404");

    expect(toUserErrorMessage(error, "fallback")).toBe(
      "요청한 데이터를 찾을 수 없습니다.",
    );
  });

  it("returns the auth-denied Korean message for ApiError 401", () => {
    const error = new ApiError(401, null, "HTTP 401");

    expect(toUserErrorMessage(error, "fallback")).toBe(
      "접근 권한이 없습니다. 다시 로그인 후 시도해주세요.",
    );
  });

  it("returns the auth-denied Korean message for ApiError 403", () => {
    const error = new ApiError(403, null, "HTTP 403");

    expect(toUserErrorMessage(error, "fallback")).toBe(
      "접근 권한이 없습니다. 다시 로그인 후 시도해주세요.",
    );
  });

  it("returns the ApiError.message when it's not the generic 'HTTP {status}' string", () => {
    const error = new ApiError(409, { status: "rejected", error: "duplicate key" }, "duplicate key");

    expect(toUserErrorMessage(error, "fallback")).toBe("duplicate key");
  });

  it("falls through ApiError-specific path and returns Error.message when message is the generic HTTP placeholder (4xx but not 401/403/404)", () => {
    // ApiError extends Error, so when the ApiError-specific branch decides the
    // message is just the generic placeholder it skips returning it, then the
    // subsequent `error instanceof Error && error.message` branch picks it up.
    // This documents the existing behavior of error-message.ts.
    const error = new ApiError(418, null, "HTTP 418");

    expect(toUserErrorMessage(error, "teapot fallback")).toBe("HTTP 418");
  });

  it("returns Error.message when error is a plain Error (non-ApiError) with a message", () => {
    const error = new Error("network timeout");

    expect(toUserErrorMessage(error, "fallback")).toBe("network timeout");
  });

  it("returns fallback when Error has empty message", () => {
    const error = new Error("");

    expect(toUserErrorMessage(error, "default-fallback")).toBe("default-fallback");
  });

  it("returns fallback for non-Error values (string, null, undefined, object)", () => {
    expect(toUserErrorMessage("oops", "default")).toBe("default");
    expect(toUserErrorMessage(null, "default")).toBe("default");
    expect(toUserErrorMessage(undefined, "default")).toBe("default");
    expect(toUserErrorMessage({ foo: "bar" }, "default")).toBe("default");
  });

  it("returns the auth-denied message before falling through to ApiError.message (401/403 short-circuit)", () => {
    // Ensures 401 path wins even when message is custom-set
    const error = new ApiError(401, null, "Custom auth message");

    expect(toUserErrorMessage(error, "fallback")).toBe(
      "접근 권한이 없습니다. 다시 로그인 후 시도해주세요.",
    );
  });
});
