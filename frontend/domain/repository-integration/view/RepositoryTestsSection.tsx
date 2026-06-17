"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Loader2, AlertCircle, RefreshCcw, TestTubes, GitCommit } from "lucide-react";
import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from "recharts";
import { fetchRepositoryTestResults } from "../service/repository-tests.service";
import {
  DEFAULT_TEST_RESULTS_LIMIT,
  DEFAULT_TEST_RESULTS_WINDOW,
  RepositoryTestResults,
  TestResultsWindow,
} from "../schema/repository-tests.types";

import { KpiTestErrorState } from "../../../shared/ui-foundation/components/KpiTestErrorState";
import { toUserErrorMessage } from "../../../shared/utils/error-message";

// RepositoryTestsSection — Sprint A (kpi-tests-per-domain-scope.md §2.1)
//
// Repository 단위 test results sub-section. build_runs.status 분포로 표현
// (별도 repository_tests table 미존재). 도넛 차트 + recent run list + window selector.

interface RepositoryTestsSectionProps {
  repoId: number;
}

const WINDOW_OPTIONS: TestResultsWindow[] = ["7d", "30d", "90d", "1y"];

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

export function RepositoryTestsSection({ repoId }: RepositoryTestsSectionProps) {
  const [results, setResults] = useState<RepositoryTestResults | null>(null);
  const isOnDetailPage = usePathname()?.endsWith(`/repositories/${repoId}/tests`);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [window, setWindow] = useState<TestResultsWindow>(DEFAULT_TEST_RESULTS_WINDOW);

  const loadResults = async (w: TestResultsWindow) => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchRepositoryTestResults(repoId, { window: w, limit: DEFAULT_TEST_RESULTS_LIMIT });
      setResults(data);
    } catch (err) {
      setError(toUserErrorMessage(err, "Failed to load repository test results"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadResults(window);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [repoId, window]);

  if (loading && !results) {
    return (
      <div className="glass border border-border rounded-2xl p-8 flex items-center justify-center gap-2 text-muted-foreground">
        <Loader2 className="w-4 h-4 animate-spin" />
        <span>Loading repository test results...</span>
      </div>
    );
  }

  if (error) {
    return (
      <KpiTestErrorState
        title="Failed to load repository test results"
        message={error}
        onRetry={() => loadResults(window)}
        testIdPrefix="repository-tests"
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

  const total = Object.values(results.totals).reduce((s, n) => s + n, 0);
  const pieData = Object.entries(results.totals)
    .filter(([, count]) => count > 0)
    .map(([status, count]) => ({
      name: STATUS_LABEL_KO[status] ?? status,
      status,
      value: count,
    }));

  const passRatePct =
    results.pass_rate !== null ? (results.pass_rate * 100).toFixed(1) : "—";
  const passRateColor =
    results.pass_rate === null
      ? "text-muted-foreground"
      : results.pass_rate >= 0.9
        ? "text-emerald-600 dark:text-emerald-300"
        : results.pass_rate >= 0.7
          ? "text-amber-600 dark:text-amber-300"
          : "text-red-600 dark:text-red-300";

  return (
    <section
      aria-label="Repository test results"
      className="space-y-4"
      data-testid="repository-tests-section"
    >
      <header className="flex items-center justify-between gap-2 flex-wrap">
        <h2 className="text-lg font-semibold flex items-center gap-2">
          <TestTubes className="w-5 h-5" />
          Test Results
        </h2>
        <div className="flex items-center gap-2">
          <select
            aria-label="Window"
            value={window}
            onChange={(e) => setWindow(e.target.value as TestResultsWindow)}
            className="text-sm bg-muted/30 border border-border rounded-md px-2 py-1"
          >
            {WINDOW_OPTIONS.map((w) => (
              <option key={w} value={w}>
                {w}
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
              href={`/repositories/${repoId}/tests`}
              data-testid="repositories-tests-drill-down"
              className="text-xs text-muted-foreground hover:text-foreground transition-colors underline underline-offset-2"
            >
              자세히 보기
            </Link>
          )}

      </header>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="glass border border-border rounded-2xl p-5">
          <div className="text-xs uppercase tracking-wide text-muted-foreground">Pass Rate</div>
          <div
            data-testid="tests-pass-rate"
            className={`text-3xl font-bold mt-2 ${passRateColor}`}
          >
            {passRatePct}%
          </div>
          <div className="text-xs text-muted-foreground mt-1">
            {total} build run{total === 1 ? "" : "s"} in window
          </div>
          {pieData.length > 0 ? (
            <div className="h-40 mt-4" data-testid="tests-pie-chart">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={pieData}
                    dataKey="value"
                    nameKey="name"
                    cx="50%"
                    cy="50%"
                    innerRadius={30}
                    outerRadius={60}
                    paddingAngle={2}
                  >
                    {pieData.map((entry, idx) => (
                      <Cell
                        key={idx}
                        fill={STATUS_COLORS[entry.status] ?? STATUS_COLORS.unknown}
                      />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            </div>
          ) : null}
        </div>

        <div className="glass border border-border rounded-2xl p-5">
          <div className="text-xs uppercase tracking-wide text-muted-foreground mb-3">
            Status Distribution
          </div>
          <ul className="space-y-1.5">
            {Object.entries(results.totals).map(([status, count]) => (
              <li
                key={status}
                className="flex items-center justify-between text-sm"
                data-testid={`tests-status-${status}`}
              >
                <span className="flex items-center gap-2">
                  <span
                    className="inline-block w-3 h-3 rounded-full"
                    style={{ background: STATUS_COLORS[status] ?? STATUS_COLORS.unknown }}
                    aria-hidden
                  />
                  {STATUS_LABEL_KO[status] ?? status}
                </span>
                <span className="font-mono tabular-nums text-muted-foreground">{count}</span>
              </li>
            ))}
          </ul>
        </div>
      </div>

      <div className="glass border border-border rounded-2xl p-5">
        <div className="text-xs uppercase tracking-wide text-muted-foreground mb-3">
          Recent Runs
        </div>
        {results.recent.length === 0 ? (
          <div className="text-sm text-muted-foreground text-center py-4">No recent runs</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm" data-testid="tests-recent-table">
              <thead>
                <tr className="text-left text-muted-foreground">
                  <th className="pb-2 font-medium">Status</th>
                  <th className="pb-2 font-medium">Branch</th>
                  <th className="pb-2 font-medium">Commit</th>
                  <th className="pb-2 font-medium">Started</th>
                </tr>
              </thead>
              <tbody>
                {results.recent.map((run) => (
                  <tr key={run.id} className="border-t border-border/40">
                    <td className="py-2">
                      <span
                        className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs"
                        style={{
                          background: (STATUS_COLORS[run.status] ?? STATUS_COLORS.unknown) + "22",
                          color: STATUS_COLORS[run.status] ?? STATUS_COLORS.unknown,
                        }}
                      >
                        <span
                          className="w-1.5 h-1.5 rounded-full"
                          style={{ background: STATUS_COLORS[run.status] ?? STATUS_COLORS.unknown }}
                          aria-hidden
                        />
                        {STATUS_LABEL_KO[run.status] ?? run.status}
                      </span>
                    </td>
                    <td className="py-2 font-mono text-xs">{run.branch}</td>
                    <td className="py-2 font-mono text-xs flex items-center gap-1">
                      <GitCommit className="w-3 h-3" />
                      {run.commit_sha.slice(0, 7)}
                    </td>
                    <td className="py-2 text-xs text-muted-foreground">
                      {new Date(run.started_at).toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </section>
  );
}
