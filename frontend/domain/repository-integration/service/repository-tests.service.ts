import { apiClient } from "@/shared/api/api-client";
import {
  DEFAULT_TEST_RESULTS_LIMIT,
  DEFAULT_TEST_RESULTS_WINDOW,
  RepositoryTestResults,
  RepositoryTestResultsResponse,
  TestResultsWindow,
} from "../schema/repository-tests.types";

// Repository Test Results service — Sprint A (kpi-tests-per-domain-scope.md §2.1)
//
// 정공법: build_runs 기반 분포. 별도 repository_tests table 미존재 (Sprint A 한정).
// 별도 table 도입은 후속.

const BASE_PATH = (id: number) => `/api/v1/repositories/${id}/test-results`;

export interface FetchRepositoryTestResultsOptions {
  window?: TestResultsWindow;
  from?: string;
  to?: string;
  limit?: number;
}

export async function fetchRepositoryTestResults(
  repositoryId: number,
  opts: FetchRepositoryTestResultsOptions = {},
): Promise<RepositoryTestResults | null> {
  const params = new URLSearchParams();
  if (opts.from && opts.to) {
    params.set("from", opts.from);
    params.set("to", opts.to);
  } else {
    params.set("window", opts.window ?? DEFAULT_TEST_RESULTS_WINDOW);
  }
  if (opts.limit && opts.limit !== DEFAULT_TEST_RESULTS_LIMIT) {
    params.set("limit", String(opts.limit));
  }
  const url = `${BASE_PATH(repositoryId)}?${params.toString()}`;
  const res = await apiClient.get<RepositoryTestResultsResponse>(url);
  return res.data ?? null;
}
