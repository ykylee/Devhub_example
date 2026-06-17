"use client";

// useRepositoryBuildRuns — N-9 잔여 build-runs polish (kpi-tests-per-domain-scope.md §6.5 + PR #555 잔여 4건 sub-issue).
//
// Custom hook for repository build runs fetch + status filter + cursor pagination.
// 2026-06-17 결정: TanStack Query 도입 안함, 기존 useState + useEffect + 직접 fetch 패턴
// 유지 (architectural sprint 분리). skeleton + status filter dropdown + 무한 스크롤 + 에러
// 정규화는 sub-issue 원본 그대로 제공.

import { useCallback, useEffect, useRef, useState } from "react";
import {
  repositoryService,
  type ListBuildRunsResult,
  type RepositoryBuildRun,
} from "../service/repository.service";

// 7 status enum (PR #486 housekeeping comment take 2 spec).
export const REPOSITORY_BUILD_RUN_STATUSES = [
  "queued",
  "running",
  "success",
  "failed",
  "cancelled",
  "skipped",
  "unknown",
] as const;
export type RepositoryBuildRunStatus = (typeof REPOSITORY_BUILD_RUN_STATUSES)[number];

export interface UseRepositoryBuildRunsOptions {
  statusFilter?: RepositoryBuildRunStatus | null;
  pageSize?: number;
  enabled?: boolean;
}

export interface UseRepositoryBuildRunsError {
  code: "not_found" | "unauthorized" | "network" | "unknown";
  message: string;
}

export interface UseRepositoryBuildRunsState {
  items: RepositoryBuildRun[];
  total: number | null;
  loading: boolean;
  loadingMore: boolean;
  error: UseRepositoryBuildRunsError | null;
  hasMore: boolean;
  loadMore: () => Promise<void>;
  refetch: () => Promise<void>;
}

/**
 * 단일 repository 의 build runs list fetch. status filter + cursor pagination
 * (offset 기반) + skeleton-free loading.
 *
 * 무한 스크롤: `loadMore` 를 IntersectionObserver 가 호출. `pageSize` 기본 20.
 */
export function useRepositoryBuildRuns(
  repositoryId: string | number,
  options: UseRepositoryBuildRunsOptions = {},
): UseRepositoryBuildRunsState {
  const { statusFilter = null, pageSize = 20, enabled = true } = options;
  const [items, setItems] = useState<RepositoryBuildRun[]>([]);
  const [total, setTotal] = useState<number | null>(null);
  const [loading, setLoading] = useState<boolean>(enabled);
  const [loadingMore, setLoadingMore] = useState<boolean>(false);
  const [error, setError] = useState<UseRepositoryBuildRunsError | null>(null);
  const offsetRef = useRef<number>(0);
  const cancelledRef = useRef<boolean>(false);

  const fetchPage = useCallback(
    async (offset: number, append: boolean): Promise<{ next: boolean; result: ListBuildRunsResult | null }> => {
      try {
        const data = await repositoryService.getRepositoryBuildRuns(
          typeof repositoryId === "string" ? Number(repositoryId) : repositoryId,
          {
            limit: pageSize,
            offset,
            status: statusFilter ?? undefined,
          },
        );
        if (cancelledRef.current) return { next: false, result: null };
        const next = data.length === pageSize;
        return { next, result: { status: "ok", data, meta: { total: data.length } } };
      } catch (e: unknown) {
        if (cancelledRef.current) return { next: false, result: null };
        const err = normalizeError(e);
        setError(err);
        return { next: false, result: null };
      }
    },
    [repositoryId, pageSize, statusFilter],
  );

  // initial fetch + status filter 변경 시 refetch
  useEffect(() => {
    cancelledRef.current = false;
    if (!enabled) {
      setItems([]);
      setTotal(null);
      setLoading(false);
      return () => {
        cancelledRef.current = true;
      };
    }
    setLoading(true);
    setError(null);
    offsetRef.current = 0;
    (async () => {
      const { next, result } = await fetchPage(0, false);
      if (cancelledRef.current) return;
      if (result) {
        setItems(result.data);
        setTotal(result.meta?.total ?? null);
        offsetRef.current = result.data.length;
      }
      setLoading(false);
      // hasMore 는 caller 가 next 로 추정
      void next;
    })();
    return () => {
      cancelledRef.current = true;
    };
  }, [fetchPage, enabled, statusFilter]);

  const loadMore = useCallback(async () => {
    if (loading || loadingMore || !enabled) return;
    setLoadingMore(true);
    const { next, result } = await fetchPage(offsetRef.current, true);
    if (cancelledRef.current) return;
    if (result) {
      setItems((prev) => [...prev, ...result.data]);
      offsetRef.current += result.data.length;
    }
    setLoadingMore(false);
  }, [fetchPage, loading, loadingMore, enabled]);

  const refetch = useCallback(async () => {
    if (!enabled) return;
    setLoading(true);
    setError(null);
    offsetRef.current = 0;
    const { result } = await fetchPage(0, false);
    if (cancelledRef.current) return;
    if (result) {
      setItems(result.data);
      setTotal(result.meta?.total ?? null);
      offsetRef.current = result.data.length;
    }
    setLoading(false);
  }, [fetchPage, enabled]);

  const hasMore = items.length < (total ?? Infinity);

  return {
    items,
    total,
    loading,
    loadingMore,
    error,
    hasMore,
    loadMore,
    refetch,
  };
}

function normalizeError(e: unknown): UseRepositoryBuildRunsError {
  if (typeof e === "object" && e !== null) {
    const obj = e as { code?: string; message?: string; status?: number };
    if (obj.status === 404 || obj.code === "repository_not_found") {
      return { code: "not_found", message: "Repository not found" };
    }
    if (obj.status === 401 || obj.status === 403) {
      return { code: "unauthorized", message: "Authentication required or insufficient permission" };
    }
    if (obj.message === "Network Error" || obj.status === undefined) {
      return { code: "network", message: obj.message ?? "Network error" };
    }
  }
  return {
    code: "unknown",
    message: e instanceof Error ? e.message : "Unknown error",
  };
}
