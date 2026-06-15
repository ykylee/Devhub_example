// Repository KPI types — Sprint A (kpi-tests-per-domain-scope.md §2.1)
//
// Backend: GET /api/v1/repositories/:id/kpi?window=30d (or from/to RFC3339)
// Response: { status: "ok", data: RepositoryKPI, meta?: ... }
//
// 정공법: snake_case wire type 그대로 노출 (변환 adapter 미사용 — 기존
// RepositoryActivity 와 동일). UI 표시는 component 내 변환.

export interface RepositoryKPI {
  repository_id: number;
  window_from: string; // RFC3339
  window_to: string; // RFC3339
  quality_score: number | null; // 0~100 (last quality_snapshots.score)
  quality_score_measured_at: string | null; // RFC3339
  build_success_rate: number; // 0.0~1.0 (window 내 가중평균)
  build_run_count: number;
  open_pr_count: number;
  merged_pr_count: number;
  active_contributor_count: number;
}

export interface RepositoryKPIResponse {
  status: "ok";
  data: RepositoryKPI;
}

export type KPIWindowDays = 7 | 30 | 90 | 365;

export const DEFAULT_KPI_WINDOW_DAYS: KPIWindowDays = 30;
