// Project KPI types — Sprint B (kpi-tests-per-domain-scope.md §2.2 + §6.2).
//
// Backend: GET /api/v1/projects/:id/kpi?window=30d (or from/to RFC3339)
// Response: { status: "ok", data: ProjectWeightedKPI, meta?: ... }
//
// 정공법: snake_case wire type 그대로 노출 (변환 adapter 미사용 — 기존
// RepositoryKPI 와 동일). UI 표시는 component 내 변환.
// 가중치 정공법: linked repository 의 raw metric × contribution_weight 가중평균.

export interface ProjectWeightedKPI {
  project_id: string; // UUID
  window_from: string; // RFC3339
  window_to: string; // RFC3339
  weighted_quality_score: number; // 0~100, 가중평균
  weighted_build_success_rate: number; // 0.0~1.0, 가중평균
  total_build_run_count: number; // Σ (단순 합산)
  weighted_open_pr_count: number; // 가중치 적용 정수 반올림
  weighted_merged_pr_count: number; // 가중치 적용 정수 반올림
  active_contributor_count: number; // Σ (단순 합산)
  linked_repository_count: number;
  weighted_at: string; // RFC3339
}

export interface ProjectKPIResponse {
  status: "ok";
  data: ProjectWeightedKPI;
}

export type ProjectKPIWindowDays = 7 | 30 | 90 | 365;

export const DEFAULT_PROJECT_KPI_WINDOW_DAYS: ProjectKPIWindowDays = 30;
