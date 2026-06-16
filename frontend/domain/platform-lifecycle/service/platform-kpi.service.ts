import { apiClient } from "@/shared/api/api-client";
import {
  DEFAULT_PLATFORM_KPI_WINDOW_DAYS,
  PlatformKPIResponse,
  PlatformKPIWindowDays,
  PlatformWeightedKPI,
} from "../schema/platform-kpi.types";

// Platform KPI service — Sprint C (kpi-tests-per-domain-scope.md §2.3 + §6.3).
//
// 정공법: apiClient<T> (자동 token refresh + session death). 기존
// project-kpi.service 와 동일 pattern. 단일 endpoint getPlatformKPI —
// sub-project equal average (Sprint B 의 sub-repo × contribution_weight 와 정공법 분리).
// platformTestResults 는 follow-up 정공법 (별도 service).

const BASE_PATH = (id: string) => `/api/v1/platforms/${id}/kpi`;

export interface FetchPlatformKPIOptions {
  /** Window days (default 30d). 7|30|90|365. */
  windowDays?: PlatformKPIWindowDays;
  /** Custom window start (RFC3339). takes precedence over windowDays when both set. */
  from?: string;
  /** Custom window end (RFC3339). */
  to?: string;
}

/**
 * Fetches the sub-project equal average PlatformWeightedKPI for a single
 * platform. Returns null when the underlying platform or window has no data
 * (404 / empty).
 */
export async function fetchPlatformKPI(
  platformId: string,
  opts: FetchPlatformKPIOptions = {},
): Promise<PlatformWeightedKPI | null> {
  const params = new URLSearchParams();
  if (opts.from && opts.to) {
    params.set("from", opts.from);
    params.set("to", opts.to);
  } else {
    params.set(
      "window",
      `${opts.windowDays ?? DEFAULT_PLATFORM_KPI_WINDOW_DAYS}d`,
    );
  }
  const url = `${BASE_PATH(platformId)}?${params.toString()}`;
  const res = await apiClient<PlatformKPIResponse>("GET", url);
  return res.data ?? null;
}
