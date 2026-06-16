"use client";

import { useEffect, useState } from "react";
import { Loader2, AlertCircle, RefreshCcw, TrendingUp, GitPullRequest, Users, Activity, Building2 } from "lucide-react";
import { fetchPlatformKPI } from "../service/platform-kpi.service";
import {
  DEFAULT_PLATFORM_KPI_WINDOW_DAYS,
  PlatformKPIWindowDays,
  PlatformWeightedKPI,
} from "../schema/platform-kpi.types";
import { toUserErrorMessage } from "@/shared/utils/error-message";

// PlatformKPISection — Sprint C (kpi-tests-per-domain-scope.md §2.3 + §6.3)
//
// Platform 단위 sub-project rollup KPI 종합 sub-section. N개 sub-project 의
// raw metric 을 sub-project equal average 로 종합. Sprint A (Repository raw)
// + Sprint B (Project 가중치) 의 정공법과 분리 (sub-project 균등).
//
// 표시: Weighted Quality Score (큰 카드) + Weighted Build Success Rate +
// Open/Merged PR Count (Σ 단순 합산) + Active Contributors (Σ) + Linked
// Project Count. Window selector (7d/30d/90d/1y) + "(sub-project avg)" 라벨.

interface PlatformKPISectionProps {
  platformId: string;
}

const WINDOW_OPTIONS: { label: string; days: PlatformKPIWindowDays }[] = [
  { label: "7d", days: 7 },
  { label: "30d", days: 30 },
  { label: "90d", days: 90 },
  { label: "1y", days: 365 },
];

export function PlatformKPISection({ platformId }: PlatformKPISectionProps) {
  const [kpi, setKpi] = useState<PlatformWeightedKPI | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [windowDays, setWindowDays] = useState<PlatformKPIWindowDays>(DEFAULT_PLATFORM_KPI_WINDOW_DAYS);

  const loadKpi = async (days: PlatformKPIWindowDays) => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchPlatformKPI(platformId, { windowDays: days });
      setKpi(data);
    } catch (err) {
      setError(toUserErrorMessage(err, "Failed to load platform KPI"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    loadKpi(windowDays);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [platformId, windowDays]);

  if (loading && !kpi) {
    return (
      <div className="glass border border-border rounded-2xl p-8 flex items-center justify-center gap-2 text-muted-foreground">
        <Loader2 className="w-4 h-4 animate-spin" />
        <span>Loading platform KPI...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="glass border border-red-300 dark:border-red-700 rounded-2xl p-6 flex items-start gap-2 text-red-600 dark:text-red-300">
        <AlertCircle className="w-4 h-4 mt-0.5" />
        <div>
          <div className="font-semibold">Failed to load platform KPI</div>
          <div className="text-sm">{error}</div>
        </div>
      </div>
    );
  }

  if (!kpi) {
    return (
      <div className="glass border border-border rounded-2xl p-6 text-center text-muted-foreground">
        No KPI data available.
      </div>
    );
  }

  const qualityScoreColor =
    kpi.weighted_quality_score >= 80
      ? "text-emerald-600 dark:text-emerald-300"
      : kpi.weighted_quality_score >= 60
        ? "text-amber-600 dark:text-amber-300"
        : "text-red-600 dark:text-red-300";

  const bsrPct = (kpi.weighted_build_success_rate * 100).toFixed(1);
  const bsrColor =
    kpi.weighted_build_success_rate >= 0.9
      ? "text-emerald-600 dark:text-emerald-300"
      : kpi.weighted_build_success_rate >= 0.7
        ? "text-amber-600 dark:text-amber-300"
        : "text-red-600 dark:text-red-300";

  return (
    <section
      aria-label="Platform KPI summary (sub-project avg)"
      className="space-y-4"
      data-testid="platform-kpi-section"
    >
      <header className="flex items-center justify-between gap-2 flex-wrap">
        <h2 className="text-lg font-semibold flex items-center gap-2">
          <Activity className="w-5 h-5" />
          Platform KPI <span className="text-xs text-muted-foreground font-normal">(sub-project avg)</span>
        </h2>
        <div className="flex items-center gap-2">
          <select
            aria-label="Window"
            value={windowDays}
            onChange={(e) => setWindowDays(Number(e.target.value) as PlatformKPIWindowDays)}
            className="text-sm bg-muted/30 border border-border rounded-md px-2 py-1"
          >
            {WINDOW_OPTIONS.map((opt) => (
              <option key={opt.days} value={opt.days}>
                {opt.label}
              </option>
            ))}
          </select>
          <button
            type="button"
            onClick={() => loadKpi(windowDays)}
            aria-label="Refresh"
            className="p-1 rounded-md hover:bg-muted/40 transition-colors"
            disabled={loading}
          >
            <RefreshCcw className={`w-4 h-4 ${loading ? "animate-spin" : ""}`} />
          </button>
        </div>
      </header>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div className="glass border border-border rounded-2xl p-5 lg:col-span-1">
          <div className="text-xs uppercase tracking-wide text-muted-foreground flex items-center gap-1">
            <TrendingUp className="w-3 h-3" /> Weighted Quality Score
          </div>
          <div
            data-testid="platform-kpi-quality-score"
            className={`text-3xl font-bold mt-2 ${qualityScoreColor}`}
          >
            {kpi.weighted_quality_score.toFixed(1)}
          </div>
          <div className="text-xs text-muted-foreground mt-1">
            sub-project avg across {kpi.linked_project_count} project{kpi.linked_project_count === 1 ? "" : "s"}
          </div>
        </div>

        <div className="glass border border-border rounded-2xl p-5">
          <div className="text-xs uppercase tracking-wide text-muted-foreground">
            Weighted Build Success Rate
          </div>
          <div
            data-testid="platform-kpi-build-success-rate"
            className={`text-3xl font-bold mt-2 ${bsrColor}`}
          >
            {bsrPct}%
          </div>
          <div className="text-xs text-muted-foreground mt-1">
            {kpi.total_build_run_count} build run{kpi.total_build_run_count === 1 ? "" : "s"} in window
          </div>
        </div>

        <div className="glass border border-border rounded-2xl p-5">
          <div className="text-xs uppercase tracking-wide text-muted-foreground flex items-center gap-1">
            <GitPullRequest className="w-3 h-3" /> Pull Requests
          </div>
          <div className="text-3xl font-bold mt-2 flex items-baseline gap-3">
            <span data-testid="platform-kpi-open-pr" className="text-violet-600 dark:text-violet-300">
              {kpi.open_pr_count}
            </span>
            <span className="text-sm text-muted-foreground">open</span>
            <span className="text-muted-foreground">/</span>
            <span data-testid="platform-kpi-merged-pr" className="text-emerald-600 dark:text-emerald-300">
              {kpi.merged_pr_count}
            </span>
            <span className="text-sm text-muted-foreground">merged</span>
          </div>
          <div className="text-xs text-muted-foreground mt-1">Σ across {kpi.linked_project_count} project</div>
        </div>

        <div className="glass border border-border rounded-2xl p-5">
          <div className="text-xs uppercase tracking-wide text-muted-foreground flex items-center gap-1">
            <Users className="w-3 h-3" /> Active Contributors
          </div>
          <div
            data-testid="platform-kpi-active-contributors"
            className="text-3xl font-bold mt-2"
          >
            {kpi.active_contributor_count}
          </div>
          <div className="text-xs text-muted-foreground mt-1">distinct authors in window (Σ)</div>
        </div>

        <div className="glass border border-border rounded-2xl p-5">
          <div className="text-xs uppercase tracking-wide text-muted-foreground flex items-center gap-1">
            <Building2 className="w-3 h-3" /> Linked Projects
          </div>
          <div
            data-testid="platform-kpi-linked-projects"
            className="text-3xl font-bold mt-2"
          >
            {kpi.linked_project_count}
          </div>
          <div className="text-xs text-muted-foreground mt-1">sub-projects on this platform</div>
        </div>
      </div>
    </section>
  );
}
