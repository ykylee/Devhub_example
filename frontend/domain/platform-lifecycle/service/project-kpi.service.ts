import { apiClient } from "@/shared/api/api-client";
import {
  DEFAULT_PROJECT_KPI_WINDOW_DAYS,
  ProjectKPIResponse,
  ProjectKPIWindowDays,
  ProjectWeightedKPI,
} from "../schema/project-kpi.types";

// Project KPI service — Sprint B (kpi-tests-per-domain-scope.md §2.2 + §6.2).
//
// 정공법: apiClient<T> (자동 token refresh + session death). 기존
// repository-kpi.service 와 동일 pattern. 단일 endpoint getProjectKPI —
// 가중치 적용 rollup (weighted_quality_score, weighted_build_success_rate).
// projectTestResults 는 follow-up PR (Sprint B-Tests) 에서 구현.

const BASE_PATH = (id: string) => `/api/v1/projects/${id}/kpi`;

export interface FetchProjectKPIOptions {
  /** Window days (default 30d). 7|30|90|365. */
  windowDays?: ProjectKPIWindowDays;
  /** Custom window start (RFC3339). takes precedence over windowDays when both set. */
  from?: string;
  /** Custom window end (RFC3339). */
  to?: string;
}

/**
 * Fetches the contribution_weight 가중치 적용 ProjectWeightedKPI for a single
 * project. Returns null when the underlying project or window has no data
 * (404 / empty).
 */
export async function fetchProjectKPI(
  projectId: string,
  opts: FetchProjectKPIOptions = {},
): Promise<ProjectWeightedKPI | null> {
  const params = new URLSearchParams();
  if (opts.from && opts.to) {
    params.set("from", opts.from);
    params.set("to", opts.to);
  } else {
    params.set(
      "window",
      `${opts.windowDays ?? DEFAULT_PROJECT_KPI_WINDOW_DAYS}d`,
    );
  }
  const url = `${BASE_PATH(projectId)}?${params.toString()}`;
  const res = await apiClient<ProjectKPIResponse>("GET", url);
  return res.data ?? null;
}
