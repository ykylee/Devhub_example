// toUserErrorMessage — 5xx/4xx/network 에러를 사용자 친화적 한국어 메시지로 변환.
// (2026-06-17 fix) backend 의 platformStoreOrUnavailable 가 503 + code: "platform_store_unavailable"
// 반환 시 machine-readable code 매핑 → "Backend store not initialized." 안내.
// @/shared/api/api-client (ApiError) import cycle 회피 — duck-typed shape (status + payload) 만 검사.

interface ApiErrorLike {
  status: number;
  payload?: unknown;
  message?: string;
}

function asApiErrorLike(value: unknown): ApiErrorLike | null {
  if (typeof value !== "object" || value === null) return null;
  const v = value as Record<string, unknown>;
  if (typeof v.status !== "number") return null;
  return v as unknown as ApiErrorLike;
}

function isJsonObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

export function toUserErrorMessage(error: unknown, fallback: string): string {
  const apiErr = asApiErrorLike(error);
  if (apiErr) {
    if (apiErr.status === 503 && isJsonObject(apiErr.payload) && apiErr.payload.code === "platform_store_unavailable") {
      return "Backend store is not initialized. 잠시 후 다시 시도하거나 관리자에게 문의하세요.";
    }
    if (apiErr.status >= 500) return "서버 오류가 발생했습니다. 잠시 후 다시 시도해주세요.";
    if (apiErr.status === 404) return "요청한 데이터를 찾을 수 없습니다.";
    if (apiErr.status === 401 || apiErr.status === 403) return "접근 권한이 없습니다. 다시 로그인 후 시도해주세요.";
    if (apiErr.message && apiErr.message !== `HTTP ${apiErr.status}`) return apiErr.message;
  }
  if (error instanceof Error && error.message) return error.message;
  return fallback;
}
