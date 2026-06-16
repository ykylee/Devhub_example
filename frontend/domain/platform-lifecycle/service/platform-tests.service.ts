import { apiClient } from "@/shared/api/api-client";
import {
  DEFAULT_PLATFORM_TEST_RESULTS_LIMIT,
  DEFAULT_PLATFORM_TEST_RESULTS_WINDOW,
  PlatformTestResultsResponse,
  PlatformTestResultsWindow,
  PlatformWeightedTestResults,
} from "../schema/platform-tests.types";

// Platform Test Results service — Sprint C (kpi-tests-per-domain-scope.md §2.3
// follow-up + §6.3).
//
// 정공법: apiClient<T> (자동 token refresh + session death). 기존
// project-tests.service 와 동일 pattern. sub-project equal average pass rate
// + multi-project recent (project_full_name + repository_full_name). PR #628
// 의 projectTestResults 정공법 차용 (sub-project 단위만 다름).

const BASE_PATH = (id: string) => `/api/v1/platforms/${id}/test-results`;

export interface FetchPlatformTestResultsOptions {
  /** Window days (default 30d). "7d" | "30d" | "90d" | "1y". */
  window?: PlatformTestResultsWindow;
  /** Custom window start (RFC3339). takes precedence over window when both set. */
  from?: string;
  /** Custom window end (RFC3339). */
  to?: string;
  /** Recent runs limit (default 20, max 50). */
  limit?: number;
}

/**
 * Fetches the sub-project equal average PlatformWeightedTestResults for a
 * single platform. Returns null when the underlying platform or window has no
 * data (404 / empty).
 */
export async function fetchPlatformTestResults(
  platformId: string,
  opts: FetchPlatformTestResultsOptions = {},
): Promise<PlatformWeightedTestResults | null> {
  const params = new URLSearchParams();
  if (opts.from && opts.to) {
    params.set("from", opts.from);
    params.set("to", opts.to);
  } else {
    params.set("window", opts.window ?? DEFAULT_PLATFORM_TEST_RESULTS_WINDOW);
  }
  if (opts.limit && opts.limit !== DEFAULT_PLATFORM_TEST_RESULTS_LIMIT) {
    params.set("limit", String(opts.limit));
  }
  const url = `${BASE_PATH(platformId)}?${params.toString()}`;
  const res = await apiClient<PlatformTestResultsResponse>("GET", url);
  return res.data ?? null;
}
