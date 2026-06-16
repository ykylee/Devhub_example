"use client";

import { useEffect, useState } from "react";
import { Loader2, AlertCircle, RefreshCcw, TrendingUp, GitPullRequest, Users, Activity, Link2 } from "lucide-react";
import { fetchProjectKPI } from "../service/project-kpi.service";
import {
  DEFAULT_PROJECT_KPI_WINDOW_DAYS,
  ProjectKPIWindowDays,
  ProjectWeightedKPI,
} from "../schema/project-kpi.types";
import { toUserErrorMessage } from "@/shared/utils/error-message";

// ProjectKPISection — Sprint B (kpi-tests-per-domain-scope.md §2.2 + §6.2)
//
// Project 단위 가중치 적용 KPI 종합 sub-section. contribution_weight 가중 평균.
// Sprint A (Repository raw) 의 RepositoryKPISection 와 정합 (4 card + window selector).
//
// 표시: Weighted Quality Score (큰 카드) + Weighted Build Success Rate +
// Weighted Open/Merged PR Count + Active Contributors + Linked Repository Count.
// Window selector (7d/30d/90d/1y) + "(weighted)" 라벨.

interface ProjectKPISectionProps {
  projectId: string;
}

const WINDOW_OPTIONS: { label: string; days: ProjectKPIWindowDays }[] = [
  { label: "7d", days: 7 },
  { label: "30d", days: 30 },
  { label: "90d", days: 90 },
  { label: "1y", days: 365 },
];

export function ProjectKPISection({ projectId }: ProjectKPISectionProps) {
  const [kpi, setKpi] = useState<ProjectWeightedKPI | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [windowDays, setWindowDays] = useState<ProjectKPIWindowDays>(DEFAULT_PROJECT_KPI_WINDOW_DAYS);

  const loadKpi = async (days: ProjectKPIWindowDays) => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchProjectKPI(projectId, { windowDays: days });
      setKpi(data);
    } catch (err) {
      setError(toUserErrorMessage(err, "Failed to load project KPI"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    loadKpi(windowDays);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [projectId, windowDays]);

  if (loading && !kpi) {
    return (
      <div className="glass border border-border rounded-2xl p-8 flex items-center justify-center gap-2 text-muted-foreground">
        <Loader2 className="w-4 h-4 animate-spin" />
        <span>Loading project KPI...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="glass border border-red-300 dark:border-red-700 rounded-2xl p-6 flex items-start gap-2 text-red-600 dark:text-red-300">
        <AlertCircle className="w-4 h-4 mt-0.5" />
        <div>
          <div className="font-semibold">Failed to load project KPI</div>
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
      aria-label="Project KPI summary (weighted)"
      className="space-y-4"
      data-testid="project-kpi-section"
    >
      <header className="flex items-center justify-between gap-2 flex-wrap">
        <h2 className="text-lg font-semibold flex items-center gap-2">
          <Activity className="w-5 h-5" />
          Project KPI <span className="text-xs text-muted-foreground font-normal">(weighted)</span>
        </h2>
        <div className="flex items-center gap-2">
          <select
            aria-label="Window"
            value={windowDays}
            onChange={(e) => setWindowDays(Number(e.target.value) as ProjectKPIWindowDays)}
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
            data-testid="project-kpi-quality-score"
            className={`text-3xl font-bold mt-2 ${qualityScoreColor}`}
          >
            {kpi.weighted_quality_score.toFixed(1)}
          </div>
          <div className="text-xs text-muted-foreground mt-1">
            가중평균 across {kpi.linked_repository_count} linked repository
          </div>
        </div>

        <div className="glass border border-border rounded-2xl p-5">
          <div className="text-xs uppercase tracking-wide text-muted-foreground">
            Weighted Build Success Rate
          </div>
          <div
            data-testid="project-kpi-build-success-rate"
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
            <GitPullRequest className="w-3 h-3" /> Weighted Pull Requests
          </div>
          <div className="text-3xl font-bold mt-2 flex items-baseline gap-3">
            <span data-testid="project-kpi-open-pr" className="text-violet-600 dark:text-violet-300">
              {kpi.weighted_open_pr_count}
            </span>
            <span className="text-sm text-muted-foreground">open</span>
            <span className="text-muted-foreground">/</span>
            <span data-testid="project-kpi-merged-pr" className="text-emerald-600 dark:text-emerald-300">
              {kpi.weighted_merged_pr_count}
            </span>
            <span className="text-sm text-muted-foreground">merged</span>
          </div>
        </div>

        <div className="glass border border-border rounded-2xl p-5">
          <div className="text-xs uppercase tracking-wide text-muted-foreground flex items-center gap-1">
            <Users className="w-3 h-3" /> Active Contributors
          </div>
          <div
            data-testid="project-kpi-active-contributors"
            className="text-3xl font-bold mt-2"
          >
            {kpi.active_contributor_count}
          </div>
          <div className="text-xs text-muted-foreground mt-1">distinct authors in window</div>
        </div>

        <div className="glass border border-border rounded-2xl p-5">
          <div className="text-xs uppercase tracking-wide text-muted-foreground flex items-center gap-1">
            <Link2 className="w-3 h-3" /> Linked Repositories
          </div>
          <div
            data-testid="project-kpi-linked-repos"
            className="text-3xl font-bold mt-2"
          >
            {kpi.linked_repository_count}
          </div>
          <div className="text-xs text-muted-foreground mt-1">with contribution_weight</div>
        </div>
      </div>
    </section>
  );
}
