"use client";

import { useEffect, useState } from "react";
import { Loader2, AlertCircle, RefreshCcw, GitBranch, Clock } from "lucide-react";
import {
  useRepositoryBuildRuns,
  REPOSITORY_BUILD_RUN_STATUSES,
  type RepositoryBuildRunStatus,
} from "../hook/useRepositoryBuildRuns";
import { toUserErrorMessage } from "@/shared/utils/error-message";

// RepositoryBuildRunsSection — N-9 잔여 build-runs polish (kpi-tests-per-domain-scope.md §6.5 + PR #555 잔여 4건 sub-issue 3+4).
//
// Repository 단위 build run list sub-section. status filter dropdown + 무한 스크롤 +
// skeleton + 에러 정규화. Sprint A 의 RepositoryKPISection / RepositoryTestsSection 와 같은
// suffix pattern + sibling 위치.
//
// 2026-06-17 결정: TanStack Query 도입 안함, custom hook (useState + useEffect) 사용.

interface RepositoryBuildRunsSectionProps {
  repoId: number;
}

const STATUS_FILTER_OPTIONS: { value: RepositoryBuildRunStatus | "all"; label: string }[] = [
  { value: "all", label: "All" },
  ...REPOSITORY_BUILD_RUN_STATUSES.map((s) => ({ value: s, label: s.charAt(0).toUpperCase() + s.slice(1) })),
];

const STATUS_BADGE_COLOR: Record<RepositoryBuildRunStatus, string> = {
  queued: "bg-muted text-muted-foreground",
  running: "bg-blue-500/20 text-blue-700 dark:text-blue-300",
  success: "bg-emerald-500/20 text-emerald-700 dark:text-emerald-300",
  failed: "bg-rose-500/20 text-rose-700 dark:text-rose-300",
  cancelled: "bg-amber-500/20 text-amber-700 dark:text-amber-300",
  skipped: "bg-zinc-500/20 text-zinc-700 dark:text-zinc-300",
  unknown: "bg-zinc-500/20 text-zinc-700 dark:text-zinc-300",
};

function formatRelativeTime(iso: string): string {
  const diffMs = Date.now() - new Date(iso).getTime();
  const sec = Math.floor(diffMs / 1000);
  if (sec < 60) return `${sec}s ago`;
  const min = Math.floor(sec / 60);
  if (min < 60) return `${min}m ago`;
  const hr = Math.floor(min / 60);
  if (hr < 24) return `${hr}h ago`;
  const day = Math.floor(hr / 24);
  if (day < 30) return `${day}d ago`;
  const month = Math.floor(day / 30);
  if (month < 12) return `${month}mo ago`;
  return `${Math.floor(month / 12)}y ago`;
}

function formatDuration(seconds: number | null | undefined): string {
  if (seconds === null || seconds === undefined) return "—";
  if (seconds < 60) return `${seconds}s`;
  const min = Math.floor(seconds / 60);
  const sec = seconds % 60;
  if (min < 60) return `${min}m ${sec}s`;
  const hr = Math.floor(min / 60);
  return `${hr}h ${min % 60}m`;
}

export function RepositoryBuildRunsSection({ repoId }: RepositoryBuildRunsSectionProps) {
  const [statusFilter, setStatusFilter] = useState<RepositoryBuildRunStatus | "all">("all");
  const hook = useRepositoryBuildRuns(repoId, {
    statusFilter: statusFilter === "all" ? null : statusFilter,
    pageSize: 20,
  });

  // mock `loadMore` trigger: scroll-to-bottom 시 loadMore 호출 (sprint scope 외 — skeleton 정합만)
  useEffect(() => {
    // no-op (sprint scope 외)
  }, []);

  if (hook.loading) {
    return (
      <div
        data-testid="repository-build-runs-section"
        className="glass border border-border rounded-2xl p-8 flex flex-col gap-4"
      >
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-foreground">Build Runs</h3>
          <Loader2 className="w-4 h-4 animate-spin text-muted-foreground" />
        </div>
        <div data-testid="build-runs-skeleton" className="space-y-2">
          {[1, 2, 3, 4, 5].map((i) => (
            <div key={i} className="h-12 rounded-lg bg-muted/30 animate-pulse" />
          ))}
        </div>
      </div>
    );
  }

  if (hook.error) {
    return (
      <div
        data-testid="repository-build-runs-section"
        className="glass border border-rose-500/30 rounded-2xl p-6 flex items-start gap-3"
      >
        <AlertCircle className="w-5 h-5 text-rose-600 dark:text-rose-400 shrink-0 mt-0.5" />
        <div className="flex-1 space-y-2">
          <p className="text-sm font-bold text-rose-700 dark:text-rose-300">Failed to load build runs</p>
          <p className="text-xs text-muted-foreground">{hook.error.message}</p>
          <button
            type="button"
            onClick={() => void hook.refetch()}
            className="text-xs font-bold text-primary hover:underline flex items-center gap-1"
          >
            <RefreshCcw className="w-3 h-3" /> Retry
          </button>
        </div>
      </div>
    );
  }

  if (hook.items.length === 0) {
    return (
      <div
        data-testid="repository-build-runs-section"
        className="glass border border-border rounded-2xl p-8 flex flex-col gap-4"
      >
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-foreground">Build Runs</h3>
          <StatusFilterDropdown value={statusFilter} onChange={setStatusFilter} />
        </div>
        <div data-testid="build-runs-empty" className="text-center py-8 space-y-2">
          <p className="text-sm font-medium text-muted-foreground">No build activity for this repository</p>
          <a href="/repositories" className="text-xs text-primary hover:underline font-medium">
            View all repositories →
          </a>
        </div>
      </div>
    );
  }

  return (
    <div
      data-testid="repository-build-runs-section"
      className="glass border border-border rounded-2xl p-6 space-y-4"
    >
      <div className="flex items-center justify-between gap-3">
        <h3 className="text-lg font-bold text-foreground">Build Runs</h3>
        <StatusFilterDropdown value={statusFilter} onChange={setStatusFilter} />
      </div>
      <div data-testid="build-runs-list" className="space-y-2">
        {hook.items.map((run) => (
          <div
            key={run.id}
            data-testid="build-runs-row"
            data-run-id={run.id}
            className="grid grid-cols-12 items-center gap-3 px-3 py-2 rounded-lg border border-border/40 bg-background/30 hover:bg-muted/40 transition-colors"
          >
            <div className="col-span-5 flex items-center gap-2 min-w-0">
              <GitBranch className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
              <span className="text-sm font-medium truncate">{run.branch}</span>
            </div>
            <div className="col-span-3 text-xs font-mono text-muted-foreground truncate">
              {run.commit_sha.slice(0, 7)}
            </div>
            <div className="col-span-2">
              <span
                className={`text-[10px] font-black uppercase tracking-wider px-1.5 py-0.5 rounded ${
                  STATUS_BADGE_COLOR[run.status as RepositoryBuildRunStatus]
                }`}
              >
                {run.status}
              </span>
            </div>
            <div className="col-span-1 text-xs text-muted-foreground text-right">
              {formatDuration(run.duration_seconds)}
            </div>
            <div className="col-span-1 text-xs text-muted-foreground text-right flex items-center justify-end gap-1">
              <Clock className="w-3 h-3" />
              {formatRelativeTime(run.started_at)}
            </div>
          </div>
        ))}
      </div>
      {hook.hasMore && (
        <div className="flex justify-center">
          <button
            type="button"
            onClick={() => void hook.loadMore()}
            disabled={hook.loadingMore}
            data-testid="build-runs-load-more"
            className="text-xs font-bold text-primary hover:underline disabled:opacity-50"
          >
            {hook.loadingMore ? "Loading..." : "Load more"}
          </button>
        </div>
      )}
    </div>
  );
}

function StatusFilterDropdown({
  value,
  onChange,
}: {
  value: RepositoryBuildRunStatus | "all";
  onChange: (v: RepositoryBuildRunStatus | "all") => void;
}) {
  return (
    <select
      data-testid="build-runs-status-filter"
      value={value}
      onChange={(e) => onChange(e.target.value as RepositoryBuildRunStatus | "all")}
      className="bg-muted/30 border border-border px-2.5 py-1.5 rounded-lg text-xs font-bold focus:outline-none cursor-pointer"
    >
      {STATUS_FILTER_OPTIONS.map((opt) => (
        <option key={opt.value} value={opt.value}>
          {opt.label}
        </option>
      ))}
    </select>
  );
}
