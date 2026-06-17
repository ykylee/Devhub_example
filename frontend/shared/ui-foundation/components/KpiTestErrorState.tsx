"use client";

import { AlertCircle, RefreshCcw } from "lucide-react";

// KpiTestErrorState — Platform/Project/Repository 의 KPI + Tests sub-section 의
// 정공법 error state component (2026-06-17 fix). backend `platformStoreOrUnavailable`
// helper (router.go) 가 503 + code: "platform_store_unavailable" 반환 시 frontend
// `toUserErrorMessage` 가 machine-readable code 매핑 → 본 component 가 명확한
// 안내 + retry button 으로 사용자가 즉시 재시도 가능. 모든 5xx/4xx/네트워크 에러
// 도 동일 UI 표시 (`message` prop 으로 구분).
//
// 5 KPI/test endpoint (repositories/:id/kpi + /test-results, projects/:id/kpi +
// /test-results, platforms/:id/kpi + /test-results) 가 backend 의 단일
// `platformStoreOrUnavailable` helper 사용 → 본 component 도 단일 component 로
// 정공법 (DRY).

interface KpiTestErrorStateProps {
  /** error 메시지 (toUserErrorMessage 결과). "Backend store is not initialized." 등. */
  message: string;
  /** section 제목. 예: "Failed to load platform KPI" */
  title: string;
  /** retry handler — useEffect 의 loadKpi 재호출 */
  onRetry: () => void;
  /** data-testid prefix. 예: "platform-kpi" → data-testid="{prefix}-error" + "{prefix}-retry" */
  testIdPrefix: string;
}

export function KpiTestErrorState({ message, title, onRetry, testIdPrefix }: KpiTestErrorStateProps) {
  return (
    <div
      data-testid={`${testIdPrefix}-error`}
      className="glass border border-red-300 dark:border-red-700 rounded-2xl p-6 flex flex-col gap-3 text-red-600 dark:text-red-300"
    >
      <div className="flex items-start gap-2">
        <AlertCircle className="w-4 h-4 mt-0.5" />
        <div>
          <div className="font-semibold">{title}</div>
          <div className="text-sm">{message}</div>
        </div>
      </div>
      <button
        type="button"
        onClick={onRetry}
        data-testid={`${testIdPrefix}-retry`}
        className="self-end flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-red-500/10 hover:bg-red-500/20 border border-red-300 dark:border-red-700 text-xs font-medium transition-colors"
      >
        <RefreshCcw className="w-3.5 h-3.5" />
        Retry
      </button>
    </div>
  );
}
