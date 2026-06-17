"use client";

import { Info } from "lucide-react";

// AnalyticsDeprecationBanner — Sprint E (kpi-tests-per-domain-scope.md §6.5)
//
// /kpis, /tests 글로벌 페이지 상단에 표시. 옵션 B (Cross-Reference) 결정의
// user-facing 안내. 도메인 상세 페이지의 sub-section 이 1차 진입점이고,
// 본 페이지의 legacy 본문 (KPIItem Python / 캘린더 / 결과 조율 큐) 은
// v0.1.0 이전 구현의 잔재이며 추후 deprecate 예정임을 명시.
//
// Tier: 사외 (frontend only, 사내 한정 정보 미포함).

export function AnalyticsDeprecationBanner() {
  return (
    <div
      role="status"
      aria-live="polite"
      data-testid="analytics-deprecation-banner"
      className="flex items-start gap-3 px-4 py-3 rounded-xl border border-amber-500/30 bg-amber-50/40 dark:bg-amber-950/20 text-sm"
    >
      <Info className="w-4 h-4 mt-0.5 shrink-0 text-amber-700 dark:text-amber-400" aria-hidden />
      <div className="space-y-0.5 text-foreground/80">
        <p className="font-medium text-amber-900 dark:text-amber-200">
          Cross-reference picker — 도메인 상세 페이지의 KPI/Tests sub-section 이 1차 진입점입니다.
        </p>
        <p className="text-xs text-muted-foreground">
          본 페이지의 legacy 위젯 (Python 스크립트 / 캘린더 / 결과 조율 큐) 은
          v0.1.0 이전 구현의 잔재이며 추후 deprecate 예정입니다. 각 entity 의
          상세 페이지 (Repository / Project / Platform) 의 KPI/Tests sub-section
          을 우선 이용해 주세요.
        </p>
      </div>
    </div>
  );
}
