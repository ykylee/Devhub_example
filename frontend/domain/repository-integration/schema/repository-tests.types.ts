// Repository Test Results types — Sprint A (kpi-tests-per-domain-scope.md §2.1)
//
// Backend: GET /api/v1/repositories/:id/test-results?window=30d&limit=20
// Response: { status: "ok", data: RepositoryTestResults, meta: { total, limit } }
//
// 정공법: build_runs.status 분포 (별도 repository_tests table 미존재). Sprint A
// 한정 표현. 별도 table 도입은 후속 sprint (Sprint D 의 follow-up).

export interface RepositoryTestResultsTotals {
  success: number;
  failed: number;
  running: number;
  cancelled: number;
  skipped: number;
  queued: number;
  unknown: number;
}

export interface RepositoryTestResultRun {
  id: number;
  run_external_id: string;
  commit_sha: string;
  status: string; // build_runs.status 그대로
  branch: string;
  started_at: string; // RFC3339
  finished_at: string | null; // RFC3339
}

export interface RepositoryTestResults {
  repository_id: number;
  window_from: string; // RFC3339
  window_to: string; // RFC3339
  totals: RepositoryTestResultsTotals;
  pass_rate: number | null; // 0.0~1.0, success/(success+failed) — 그 외 0
  recent: RepositoryTestResultRun[];
}

export interface RepositoryTestResultsResponse {
  status: "ok";
  data: RepositoryTestResults;
  meta: {
    total: number;
    limit: number;
  };
}

export type TestResultsWindow = "7d" | "30d" | "90d" | "1y";

export const DEFAULT_TEST_RESULTS_WINDOW: TestResultsWindow = "30d";
export const DEFAULT_TEST_RESULTS_LIMIT = 20;
