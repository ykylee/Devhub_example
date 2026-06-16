// Platform Test Results types — Sprint C (kpi-tests-per-domain-scope.md §2.3
// follow-up + §6.3).
//
// Backend: GET /api/v1/platforms/:id/test-results?window=30d&limit=20
// Response: { status: "ok", data: PlatformWeightedTestResults, meta: { total, limit } }
//
// 정공법: snake_case wire type 그대로 노출 (변환 adapter 미사용 — 기존
// ProjectWeightedTestResults 와 동일 pattern). Platform sub-project rollup:
//   - weighted_pass_rate = AVG(per_project_weighted_pass_rate) (sub-project 균등).
//     각 sub-project 별 denom=0 → null (linked repo 0 또는 모든 build 가
//     queued/running/cancelled/skipped/unknown).
//   - totals = N개 sub-project 의 build_runs status 합산 (sub-project 무관, 단순 count).
//   - recent = 모든 sub-project 의 build_runs 최신순 limit (multi-project, project_full_name
//     + repository_full_name 표시 — ProjectBuildRun 의 상위 확장).

export interface PlatformTestResultsTotals {
  success: number;
  failed: number;
  running: number;
  cancelled: number;
  skipped: number;
  queued: number;
  unknown: number;
}

export interface PlatformBuildRun {
  id: number;
  project_id: string; // multi-project 표기용 추가
  project_full_name: string;
  repository_id: number;
  repository_full_name: string;
  run_external_id: string;
  commit_sha: string;
  status: string;
  branch: string;
  started_at: string; // RFC3339
  finished_at: string | null; // RFC3339
  // Sprint C — duration_seconds optional (PR #597 의 projectBuildRun 와 정합)
  duration_seconds?: number | null;
}

export interface PlatformWeightedTestResults {
  platform_id: string; // UUID
  window_from: string; // RFC3339
  window_to: string; // RFC3339
  weighted_pass_rate: number | null; // 0.0~1.0, sub-project avg
  totals: PlatformTestResultsTotals; // 7 status
  recent: PlatformBuildRun[]; // multi-project
}

export interface PlatformTestResultsResponse {
  status: "ok";
  data: PlatformWeightedTestResults;
  meta: {
    total: number;
    limit: number;
  };
}

export type PlatformTestResultsWindow = "7d" | "30d" | "90d" | "1y";

export const DEFAULT_PLATFORM_TEST_RESULTS_WINDOW: PlatformTestResultsWindow = "30d";
export const DEFAULT_PLATFORM_TEST_RESULTS_LIMIT = 20;
