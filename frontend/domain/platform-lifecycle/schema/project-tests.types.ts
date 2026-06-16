// Project Test Results types — Sprint B-Tests (kpi-tests-per-domain-scope.md §2.2
// follow-up).
//
// Backend: GET /api/v1/projects/:id/test-results?window=30d&limit=20
// Response: { status: "ok", data: ProjectWeightedTestResults, meta: { total, limit } }
//
// 정공법: snake_case wire type 그대로 노출 (변환 adapter 미사용 — 기존
// RepositoryTestResults 와 동일). 가중치 정공법: linked repository 의 build_runs
// status × contribution_weight 가중평균 (pass rate). multi-repo recent 표시 위해
// repository_full_name 필드 추가.

export interface ProjectTestResultsTotals {
  success: number;
  failed: number;
  running: number;
  cancelled: number;
  skipped: number;
  queued: number;
  unknown: number;
}

export interface ProjectBuildRun {
  id: number;
  repository_id: number;
  repository_full_name: string;
  run_external_id: string;
  commit_sha: string;
  status: string; // build_runs.status 그대로
  branch: string;
  started_at: string; // RFC3339
  finished_at: string | null; // RFC3339
}

export interface ProjectWeightedTestResults {
  project_id: string; // UUID
  window_from: string; // RFC3339
  window_to: string; // RFC3339
  weighted_pass_rate: number | null; // 0.0~1.0, denom=0 → null
  totals: ProjectTestResultsTotals; // 7 status
  recent: ProjectBuildRun[]; // multi-repo, repository_full_name 포함
}

export interface ProjectTestResultsResponse {
  status: "ok";
  data: ProjectWeightedTestResults;
  meta: {
    total: number;
    limit: number;
  };
}

export type ProjectTestResultsWindow = "7d" | "30d" | "90d" | "1y";

export const DEFAULT_PROJECT_TEST_RESULTS_WINDOW: ProjectTestResultsWindow = "30d";
export const DEFAULT_PROJECT_TEST_RESULTS_LIMIT = 20;
