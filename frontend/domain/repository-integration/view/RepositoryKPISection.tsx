"use client";

import { useEffect, useState } from "react";
import { Loader2, AlertCircle, RefreshCcw, TrendingUp, GitPullRequest, Users, Activity } from "lucide-react";
import { fetchRepositoryKPI } from "../service/repository-kpi.service";
import {
  DEFAULT_KPI_WINDOW_DAYS,
  KPIWindowDays,
  RepositoryKPI,
} from "../schema/repository-kpi.types";

import { KpiTestErrorState } from "../../../shared/ui-foundation/components/KpiTestErrorState";
import { toUserErrorMessage } from "../../../shared/utils/error-message";

// RepositoryKPISection — Sprint A (kpi-tests-per-domain-scope.md §2.1)
//
// Repository 단위 raw KPI 종합 sub-section. 가중치 미적용 (single repo = weight=1).
// Sprint B (Project 가중치) + Sprint C (Platform sub-rollup) 의 기반.
//
// 표시: Quality Score (큰 카드) + Build Success Rate + Open PR + Merged PR +
// Active Contributors. Window selector (7d/30d/90d).

interface RepositoryKPISectionProps {
  repoId: number;
}

const WINDOW_OPTIONS: { label: string; days: KPIWindowDays }[] = [
  { label: "7d", days: 7 },
  { label: "30d", days: 30 },
  { label: "90d", days: 90 },
  { label: "1y", days: 365 },
];

export function RepositoryKPISection({ repoId }: RepositoryKPISectionProps) {
  const [kpi, setKpi] = useState<RepositoryKPI | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [windowDays, setWindowDays] = useState<KPIWindowDays>(DEFAULT_KPI_WINDOW_DAYS);

  const loadKpi = async (days: KPIWindowDays) => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchRepositoryKPI(repoId, { windowDays: days });
      setKpi(data);
    } catch (err) {
      setError(toUserErrorMessage(err, "Failed to load repository KPI"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadKpi(windowDays);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [repoId, windowDays]);

  if (loading && !kpi) {
    return (
      <div className="glass border border-border rounded-2xl p-8 flex items-center justify-center gap-2 text-muted-foreground">
        <Loader2 className="w-4 h-4 animate-spin" />
        <span>Loading repository KPI...</span>
      </div>
    );
  }

  if (error) {
    return (
      <KpiTestErrorState
        title="Failed to load repository KPI"
        message={error}
        onRetry={() => loadKpi(windowDays)}
        testIdPrefix="repository-kpi"
      />
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
    kpi.quality_score === null
      ? "text-muted-foreground"
      : kpi.quality_score >= 80
        ? "text-emerald-600 dark:text-emerald-300"
        : kpi.quality_score >= 60
          ? "text-amber-600 dark:text-amber-300"
          : "text-red-600 dark:text-red-300";

  const bsrPct = (kpi.build_success_rate * 100).toFixed(1);
  const bsrColor =
    kpi.build_success_rate >= 0.9
      ? "text-emerald-600 dark:text-emerald-300"
      : kpi.build_success_rate >= 0.7
        ? "text-amber-600 dark:text-amber-300"
        : "text-red-600 dark:text-red-300";

  return (
    <section
      aria-label="Repository KPI summary"
      className="space-y-4"
      data-testid="repository-kpi-section"
    >
      <header className="flex items-center justify-between gap-2 flex-wrap">
        <h2 className="text-lg font-semibold flex items-center gap-2">
          <Activity className="w-5 h-5" />
          Repository KPI
        </h2>
        <div className="flex items-center gap-2">
          <select
            aria-label="Window"
            value={windowDays}
            onChange={(e) => setWindowDays(Number(e.target.value) as KPIWindowDays)}
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
            <TrendingUp className="w-3 h-3" /> Quality Score
          </div>
          <div
            data-testid="kpi-quality-score"
            className={`text-3xl font-bold mt-2 ${qualityScoreColor}`}
          >
            {kpi.quality_score !== null ? kpi.quality_score.toFixed(1) : "—"}
          </div>
          <div className="text-xs text-muted-foreground mt-1">
            {kpi.quality_score_measured_at
              ? `Measured ${new Date(kpi.quality_score_measured_at).toLocaleString()}`
              : "No recent measurement"}
          </div>
        </div>

        <div className="glass border border-border rounded-2xl p-5">
          <div className="text-xs uppercase tracking-wide text-muted-foreground">Build Success Rate</div>
          <div
            data-testid="kpi-build-success-rate"
            className={`text-3xl font-bold mt-2 ${bsrColor}`}
          >
            {bsrPct}%
          </div>
          <div className="text-xs text-muted-foreground mt-1">
            {kpi.build_run_count} build run{kpi.build_run_count === 1 ? "" : "s"} in window
          </div>
        </div>

        <div className="glass border border-border rounded-2xl p-5">
          <div className="text-xs uppercase tracking-wide text-muted-foreground flex items-center gap-1">
            <GitPullRequest className="w-3 h-3" /> Pull Requests
          </div>
          <div className="text-3xl font-bold mt-2 flex items-baseline gap-3">
            <span data-testid="kpi-open-pr" className="text-violet-600 dark:text-violet-300">
              {kpi.open_pr_count}
            </span>
            <span className="text-sm text-muted-foreground">open</span>
            <span className="text-muted-foreground">/</span>
            <span data-testid="kpi-merged-pr" className="text-emerald-600 dark:text-emerald-300">
              {kpi.merged_pr_count}
            </span>
            <span className="text-sm text-muted-foreground">merged</span>
          </div>
        </div>

        <div className="glass border border-border rounded-2xl p-5">
          <div className="text-xs uppercase tracking-wide text-muted-foreground flex items-center gap-1">
            <Users className="w-3 h-3" /> Active Contributors
          </div>
          <div
            data-testid="kpi-active-contributors"
            className="text-3xl font-bold mt-2"
          >
            {kpi.active_contributor_count}
          </div>
          <div className="text-xs text-muted-foreground mt-1">distinct authors in window</div>
        </div>
      </div>
    </section>
  );
}
