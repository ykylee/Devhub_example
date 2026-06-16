import { apiClient } from "@/shared/api/api-client";
import {
  DEFAULT_PROJECT_TEST_RESULTS_LIMIT,
  DEFAULT_PROJECT_TEST_RESULTS_WINDOW,
  ProjectTestResultsResponse,
  ProjectTestResultsWindow,
  ProjectWeightedTestResults,
} from "../schema/project-tests.types";

// Project Test Results service — Sprint B-Tests (kpi-tests-per-domain-scope.md
// §2.2 follow-up).
//
// 정공법: apiClient<T> (자동 token refresh + session death). 기존
// repository-tests.service 와 동일 pattern. 가중치 적용 rollup (weighted_pass_rate)
// + multi-repo recent. PR #597 의 RepositoryTestsSection 정공법 차용.

const BASE_PATH = (id: string) => `/api/v1/projects/${id}/test-results`;

export interface FetchProjectTestResultsOptions {
  /** Window days (default 30d). "7d" | "30d" | "90d" | "1y". */
  window?: ProjectTestResultsWindow;
  /** Custom window start (RFC3339). takes precedence over window when both set. */
  from?: string;
  /** Custom window end (RFC3339). */
  to?: string;
  /** Recent runs limit (default 20, max 50). */
  limit?: number;
}

/**
 * Fetches the contribution_weight 가중치 적용 ProjectWeightedTestResults for a
 * single project. Returns null when the underlying project or window has no data
 * (404 / empty).
 */
export async function fetchProjectTestResults(
  projectId: string,
  opts: FetchProjectTestResultsOptions = {},
): Promise<ProjectWeightedTestResults | null> {
  const params = new URLSearchParams();
  if (opts.from && opts.to) {
    params.set("from", opts.from);
    params.set("to", opts.to);
  } else {
    params.set("window", opts.window ?? DEFAULT_PROJECT_TEST_RESULTS_WINDOW);
  }
  if (opts.limit && opts.limit !== DEFAULT_PROJECT_TEST_RESULTS_LIMIT) {
    params.set("limit", String(opts.limit));
  }
  const url = `${BASE_PATH(projectId)}?${params.toString()}`;
  const res = await apiClient<ProjectTestResultsResponse>("GET", url);
  return res.data ?? null;
}
