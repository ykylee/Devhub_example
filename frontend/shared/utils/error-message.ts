import { ApiError } from "@/shared/api/api-client";

export function toUserErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof ApiError) {
    if (error.status >= 500) return "서버 오류가 발생했습니다. 잠시 후 다시 시도해주세요.";
    if (error.status === 404) return "요청한 데이터를 찾을 수 없습니다.";
    if (error.status === 401 || error.status === 403) return "접근 권한이 없습니다. 다시 로그인 후 시도해주세요.";
    if (error.message && error.message !== `HTTP ${error.status}`) return error.message;
  }
  if (error instanceof Error && error.message) return error.message;
  return fallback;
}
