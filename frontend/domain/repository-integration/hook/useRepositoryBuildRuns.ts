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
  // Per-request token. effect/loadMore/refetch 가 호출될 때마다 새 AbortController 를
  // 생성해 abortControllerRef 에 보관하고 cleanup 시 abort. await 후 setState 단계
  // 진입 직전에 controllerRef.current 와 일치 여부 + signal.aborted 로 stale 결과를
  // 차단 (codex P2 review 2026-06-17).
  const abortControllerRef = useRef<AbortController | null>(null);

  const fetchPage = useCallback(
    async (offset: number): Promise<{ result: ListBuildRunsResult | null; controller: AbortController }> => {
      const controller = new AbortController();
      abortControllerRef.current = controller;
      try {
        const result = await repositoryService.getRepositoryBuildRunsWithMeta(
          typeof repositoryId === "string" ? Number(repositoryId) : repositoryId,
          {
            limit: pageSize,
            offset,
            status: statusFilter ?? undefined,
            signal: controller.signal,
          },
        );
        return { result, controller };
      } catch (e: unknown) {
        // abort 는 정상 cancellation — error UI 표시 안 함. signal.aborted 만 검사.
        if (controller.signal.aborted) {
          return { result: null, controller };
        }
        setError(normalizeError(e));
        return { result: null, controller };
      }
    },
    [repositoryId, pageSize, statusFilter],
  );

  const isCurrent = (controller: AbortController): boolean =>
    abortControllerRef.current === controller && !controller.signal.aborted;

  // initial fetch + status filter 변경 시 refetch
  useEffect(() => {
    if (!enabled) {
      abortControllerRef.current?.abort();
      abortControllerRef.current = null;
      setItems([]);
      setTotal(null);
      setError(null);
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    offsetRef.current = 0;
    (async () => {
      const { result, controller } = await fetchPage(0);
      if (!isCurrent(controller)) return;
      if (result) {
        setItems(result.data);
        setTotal(result.meta?.total ?? null);
        offsetRef.current = result.data.length;
      }
      setLoading(false);
    })();
    return () => {
      abortControllerRef.current?.abort();
    };
  }, [fetchPage, enabled, statusFilter]);

  const loadMore = useCallback(async () => {
    if (loading || loadingMore || !enabled) return;
    setLoadingMore(true);
    const { result, controller } = await fetchPage(offsetRef.current);
    if (!isCurrent(controller)) {
      // PR #635/#636/#637 codex P2 review 정공법 (2026-06-17) — stale controller 에서
      // setLoadingMore(false) 누락 시 button 영구 disabled. effect 의 cleanup 이
      // 새 controller 로 교체 시 stale return path 가 setLoadingMore(false) 호출
      // 없이 skip → 다음 effect 의 loading state 가 stuck true.
      setLoadingMore(false);
      return;
    }
    if (result) {
      setItems((prev) => [...prev, ...result.data]);
      offsetRef.current += result.data.length;
    }
    setLoadingMore(false);
  }, [fetchPage, loading, loadingMore, enabled]);

  const refetch = useCallback(async () => {
    if (!enabled) return;
    abortControllerRef.current?.abort();
    setLoading(true);
    setError(null);
    offsetRef.current = 0;
    const { result, controller } = await fetchPage(0);
    if (!isCurrent(controller)) return;
    if (result) {
      setItems(result.data);
      setTotal(result.meta?.total ?? null);
      offsetRef.current = result.data.length;
    }
    setLoading(false);
  }, [fetchPage, enabled]);

  // hasMore: backend meta.total 우선, 미노출 시 pageSize 신호로 추정 (fallback).
  const hasMore = total != null ? items.length < total : items.length >= pageSize;

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
