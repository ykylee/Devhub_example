"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Loader2, AlertCircle, RefreshCcw, TestTubes, GitCommit, Building2 } from "lucide-react";
import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from "recharts";
import { fetchPlatformTestResults } from "../service/platform-tests.service";
import {
  DEFAULT_PLATFORM_TEST_RESULTS_LIMIT,
  DEFAULT_PLATFORM_TEST_RESULTS_WINDOW,
  PlatformTestResultsWindow,
  PlatformWeightedTestResults,
} from "../schema/platform-tests.types";
import { toUserErrorMessage } from "@/shared/utils/error-message";
import { KpiTestErrorState } from "@/shared/ui-foundation/components/KpiTestErrorState";

// PlatformTestsSection — Sprint C (kpi-tests-per-domain-scope.md §2.3
// follow-up + §6.3)
//
// Platform 단위 sub-project rollup test results sub-section. N개 sub-project 의
// build_runs status 종합 + sub-project equal average pass rate + multi-project
// recent (project_full_name + repository_full_name). Sprint A (Repository raw)
// + Sprint B-Tests (Project 가중치) 의 정공법과 분리 (sub-project 균등).

interface PlatformTestsSectionProps {
  platformId: string;
}

const WINDOW_OPTIONS: PlatformTestResultsWindow[] = ["7d", "30d", "90d", "1y"];

const STATUS_COLORS: Record<string, string> = {
  success: "#10b981", // emerald-500
  failed: "#ef4444", // red-500
  running: "#3b82f6", // blue-500
  cancelled: "#94a3b8", // slate-400
  skipped: "#f59e0b", // amber-500
  queued: "#8b5cf6", // violet-500
  unknown: "#cbd5e1", // slate-300
};

const STATUS_LABEL_KO: Record<string, string> = {
  success: "성공",
  failed: "실패",
  running: "진행중",
  cancelled: "취소",
  skipped: "건너뜀",
  queued: "대기",
  unknown: "알 수 없음",
};

export function PlatformTestsSection({ platformId }: PlatformTestsSectionProps) {
  const [results, setResults] = useState<PlatformWeightedTestResults | null>(null);
  const isOnDetailPage = usePathname()?.endsWith(`/platforms/${platformId}/tests`);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [window, setWindow] = useState<PlatformTestResultsWindow>(DEFAULT_PLATFORM_TEST_RESULTS_WINDOW);

  const loadResults = async (w: PlatformTestResultsWindow) => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchPlatformTestResults(platformId, {
        window: w,
        limit: DEFAULT_PLATFORM_TEST_RESULTS_LIMIT,
      });
      setResults(data);
    } catch (err) {
      setError(toUserErrorMessage(err, "Failed to load platform test results"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    loadResults(window);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [platformId, window]);

  if (loading && !results) {
    return (
      <div className="glass border border-border rounded-2xl p-8 flex items-center justify-center gap-2 text-muted-foreground">
        <Loader2 className="w-4 h-4 animate-spin" />
        <span>Loading platform test results...</span>
      </div>
    );
  }

  if (error) {
    return (
      <KpiTestErrorState
        title="Failed to load platform test results"
        message={error}
        onRetry={() => loadResults(window)}
        testIdPrefix="platform-tests"
      />
    );
  }

  if (!results) {
    return (
      <div className="glass border border-border rounded-2xl p-6 text-center text-muted-foreground">
        No test results available.
      </div>
    );
  }

  const passRatePct =
    results.weighted_pass_rate === null
      ? "—"
      : `${(results.weighted_pass_rate * 100).toFixed(1)}%`;

  const totalsEntries = Object.entries(results.totals).filter(([, n]) => n > 0);
  const totalRuns = Object.values(results.totals).reduce((a, b) => a + b, 0);
  const pieData = totalsEntries.map(([status, count]) => ({
    name: STATUS_LABEL_KO[status] ?? status,
    status,
    value: count,
  }));

  return (
    <section
      aria-label="Platform Test Results (sub-project avg)"
      className="space-y-4"
      data-testid="platform-tests-section"
    >
      <header className="flex items-center justify-between gap-2 flex-wrap">
        <h2 className="text-lg font-semibold flex items-center gap-2">
          <TestTubes className="w-5 h-5" />
          Platform Tests <span className="text-xs text-muted-foreground font-normal">(sub-project avg)</span>
        </h2>
        <div className="flex items-center gap-2">
          <select
            aria-label="Window"
            value={window}
            onChange={(e) => setWindow(e.target.value as PlatformTestResultsWindow)}
            className="text-sm bg-muted/30 border border-border rounded-md px-2 py-1"
          >
            {WINDOW_OPTIONS.map((opt) => (
              <option key={opt} value={opt}>
                {opt}
              </option>
            ))}
          </select>
          <button
            type="button"
            onClick={() => loadResults(window)}
            aria-label="Refresh"
            className="p-1 rounded-md hover:bg-muted/40 transition-colors"
            disabled={loading}
          >
            <RefreshCcw className={`w-4 h-4 ${loading ? "animate-spin" : ""}`} />
          </button>
        </div>
          {!isOnDetailPage && (
            <Link
              href={`/platforms/${platformId}/tests`}
              data-testid="platforms-tests-drill-down"
              className="text-xs text-muted-foreground hover:text-foreground transition-colors underline underline-offset-2"
            >
              자세히 보기
            </Link>
          )}

      </header>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="glass border border-border rounded-2xl p-5">
          <div className="text-xs uppercase tracking-wide text-muted-foreground">
            Weighted Pass Rate
          </div>
          <div
            data-testid="platform-tests-pass-rate"
            className="text-3xl font-bold mt-2"
          >
            {passRatePct}
          </div>
          <div className="text-xs text-muted-foreground mt-1">
            sub-project avg across multi-project window
          </div>
        </div>

        <div className="glass border border-border rounded-2xl p-5">
          <div className="text-xs uppercase tracking-wide text-muted-foreground">
            Total Runs
          </div>
          <div
            data-testid="platform-tests-total"
            className="text-3xl font-bold mt-2"
          >
            {totalRuns}
          </div>
          <div className="text-xs text-muted-foreground mt-1">
            {Object.keys(results.totals).length} status counter
          </div>
        </div>
      </div>

      {pieData.length > 0 && (
        <div className="glass border border-border rounded-2xl p-5">
          <div className="text-xs uppercase tracking-wide text-muted-foreground mb-2">
            Build Status Distribution
          </div>
          <div className="h-56">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={pieData}
                  dataKey="value"
                  nameKey="name"
                  outerRadius={80}
                  innerRadius={50}
                  stroke="rgba(0,0,0,0.1)"
                >
                  {pieData.map((d) => (
                    <Cell key={d.status} fill={STATUS_COLORS[d.status] ?? "#cbd5e1"} />
                  ))}
                </Pie>
                <Tooltip
                  contentStyle={{
                    backgroundColor: "rgba(0,0,0,0.85)",
                    border: "1px solid rgba(255,255,255,0.1)",
                    borderRadius: 8,
                    color: "#fff",
                  }}
                />
              </PieChart>
            </ResponsiveContainer>
          </div>
        </div>
      )}

      <div className="glass border border-border rounded-2xl p-5">
        <div className="text-xs uppercase tracking-wide text-muted-foreground mb-3 flex items-center gap-1">
          <GitCommit className="w-3 h-3" /> Recent Runs (multi-project)
        </div>
        {results.recent.length === 0 ? (
          <div className="text-sm text-muted-foreground">No recent runs in window.</div>
        ) : (
          <ul className="space-y-2" data-testid="platform-tests-recent">
            {results.recent.map((run) => (
              <li
                key={run.id}
                className="flex items-center justify-between gap-2 text-sm border-b border-white/5 last:border-b-0 py-1.5"
              >
                <div className="flex items-center gap-2 min-w-0 flex-1">
                  <span
                    className="inline-block w-2 h-2 rounded-full shrink-0"
                    style={{ backgroundColor: STATUS_COLORS[run.status] ?? "#cbd5e1" }}
                    aria-label={STATUS_LABEL_KO[run.status] ?? run.status}
                  />
                  <Building2 className="w-3 h-3 text-muted-foreground shrink-0" />
                  <span className="font-medium truncate" data-testid={`platform-tests-recent-project-${run.id}`}>
                    {run.project_full_name}
                  </span>
                  <span className="text-muted-foreground shrink-0">/</span>
                  <span className="text-muted-foreground truncate">{run.repository_full_name}</span>
                </div>
                <div className="flex items-center gap-2 shrink-0">
                  <span className="text-xs font-mono text-muted-foreground truncate max-w-[100px]">
                    {run.branch}
                  </span>
                  <span className="text-xs text-muted-foreground">{run.commit_sha.slice(0, 7)}</span>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </section>
  );
}
