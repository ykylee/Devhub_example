import { apiClient } from "@/shared/api/api-client";
import {
  DEFAULT_KPI_WINDOW_DAYS,
  KPIWindowDays,
  RepositoryKPI,
  RepositoryKPIResponse,
} from "../schema/repository-kpi.types";

// Repository KPI service — Sprint A (kpi-tests-per-domain-scope.md §2.1)
//
// 정공법: apiClient<T> (자동 token refresh + session death). 기존 repository.service.ts
// 와 동일 pattern. 단일 endpoint getRepositoryKPI — window/limit query option.

const BASE_PATH = (id: number) => `/api/v1/repositories/${id}/kpi`;

export interface FetchRepositoryKPIOptions {
  /** Window days (default 30d). 7|30|90|365. */
  windowDays?: KPIWindowDays;
  /** Custom window start (RFC3339). takes precedence over windowDays when both set. */
  from?: string;
  /** Custom window end (RFC3339). */
  to?: string;
}

/**
 * Fetches the raw (weight=1) KPI summary for a single repository. Returns null
 * when the underlying repository or window has no data (404 / empty).
 */
export async function fetchRepositoryKPI(
  repositoryId: number,
  opts: FetchRepositoryKPIOptions = {},
): Promise<RepositoryKPI | null> {
  const params = new URLSearchParams();
  if (opts.from && opts.to) {
    params.set("from", opts.from);
    params.set("to", opts.to);
  } else {
    params.set(
      "window",
      `${opts.windowDays ?? DEFAULT_KPI_WINDOW_DAYS}d`,
    );
  }
  const url = `${BASE_PATH(repositoryId)}?${params.toString()}`;
  const res = await apiClient.get<RepositoryKPIResponse>(url);
  return res.data ?? null;
}
