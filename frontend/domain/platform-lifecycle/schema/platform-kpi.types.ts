// Platform KPI types — Sprint C (kpi-tests-per-domain-scope.md §2.3 + §6.3).
//
// Backend: GET /api/v1/platforms/:id/kpi?window=30d (or from/to RFC3339)
// Response: { status: "ok", data: PlatformWeightedKPI }
//
// 정공법: snake_case wire type 그대로 노출 (변환 adapter 미사용 — 기존
// ProjectWeightedKPI 와 동일 pattern). Platform sub-project rollup 정공법:
// platform 의 N개 sub-project 의 raw metric 을 **sub-project 단위 equal average**.
// Sprint B (projectKPI = sub-repo × contribution_weight) 와 정공법 분리.
//
//   - per_project_X = sub-project 의 N개 linked repository 의 raw metric 종합
//   - platform weighted_X = AVG(per_project_X) (sub-project 균등)
//   - total_build_run_count / open_pr_count / merged_pr_count / active_contributor_count
//     = Σ (sub-project 무관, 단순 합산)
//   - linked_project_count = 0 인 경우 모든 가중치 metric 0.0

export interface PlatformWeightedKPI {
  platform_id: string; // UUID
  window_from: string; // RFC3339
  window_to: string; // RFC3339
  weighted_quality_score: number; // 0~100, sub-project avg
  weighted_build_success_rate: number; // 0.0~1.0, sub-project avg
  total_build_run_count: number; // Σ (단순 합산)
  open_pr_count: number; // Σ (단순 합산)
  merged_pr_count: number; // Σ (단순 합산)
  active_contributor_count: number; // Σ (단순 합산)
  linked_project_count: number; // projects.platform_id 의 project 수
  weighted_at: string; // RFC3339
}

export interface PlatformKPIResponse {
  status: "ok";
  data: PlatformWeightedKPI;
}

export type PlatformKPIWindowDays = 7 | 30 | 90 | 365;

export const DEFAULT_PLATFORM_KPI_WINDOW_DAYS: PlatformKPIWindowDays = 30;
