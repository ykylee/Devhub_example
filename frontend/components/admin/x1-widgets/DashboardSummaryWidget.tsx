"use client";

import { useEffect, useState } from "react";
import { adminX1Service } from "@/domain/integration-registry/service/admin-x1.service";
import type { IntegrationSyncJobStatusCounts } from "@/domain/integration-registry/schema/integration.types";
import { Gauge, Loader2, AlertTriangle } from "lucide-react";

interface Summary {
  totalJobs: number;
  queueDepth: number;
  failedCount: number;
  successRate: number;
}

/**
 * DashboardSummaryWidget — X-1 dashboard 의 종합 summary 위젯.
 * sync job 4 status 별 count → totalJobs / queueDepth / failedCount / successRate.
 * API-106 의 getStatusSummary() 백엔드.
 */
export function DashboardSummaryWidget() {
  const [summary, setSummary] = useState<Summary | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchSummary = async () => {
      setIsLoading(true);
      setError(null);
      try {
        const resp = await adminX1Service.getStatusSummary();
        const c: IntegrationSyncJobStatusCounts = resp.sync_job_status_counts;
        const total = c.queued + c.running + c.succeeded + c.failed;
        const completed = c.succeeded + c.failed;
        const successRate = completed > 0 ? c.succeeded / completed : 0;
        setSummary({
          totalJobs: total,
          queueDepth: c.queued + c.running,
          failedCount: c.failed,
          successRate: Math.round(successRate * 1000) / 10,
        });
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to load dashboard summary");
      } finally {
        setIsLoading(false);
      }
    };
    fetchSummary();
  }, []);

  return (
    <div className="glass border-border rounded-2xl p-5 space-y-3 shadow-sm">
      <div className="flex items-center gap-2">
        <Gauge className="w-4 h-4 text-primary" />
        <h3 className="text-sm font-bold uppercase tracking-wider text-foreground">
          Dashboard Summary
        </h3>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center py-8 text-muted-foreground">
          <Loader2 className="w-5 h-5 animate-spin mr-2" />
          <span className="text-xs">Loading summary...</span>
        </div>
      ) : error ? (
        <div className="flex items-center gap-2 py-4 text-xs text-destructive">
          <AlertTriangle className="w-4 h-4" />
          <span>{error}</span>
        </div>
      ) : summary ? (
        <div className="grid grid-cols-2 gap-3">
          <Stat label="Total Jobs" value={summary.totalJobs} />
          <Stat label="Queue Depth" value={summary.queueDepth} />
          <Stat label="Failed" value={summary.failedCount} tone={summary.failedCount > 0 ? "danger" : "neutral"} />
          <Stat label="Success Rate" value={`${summary.successRate}%`} tone={summary.successRate >= 95 ? "good" : "warn"} />
        </div>
      ) : null}
    </div>
  );
}

function Stat({ label, value, tone = "neutral" }: { label: string; value: number | string; tone?: "neutral" | "good" | "warn" | "danger" }) {
  const toneClass =
    tone === "good"
      ? "text-emerald-600 dark:text-emerald-400"
      : tone === "warn"
        ? "text-amber-600 dark:text-amber-400"
        : tone === "danger"
          ? "text-rose-600 dark:text-rose-400"
          : "text-foreground";
  return (
    <div className="space-y-1">
      <div className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">
        {label}
      </div>
      <div className={`text-2xl font-black tabular-nums ${toneClass}`}>{value}</div>
    </div>
  );
}
