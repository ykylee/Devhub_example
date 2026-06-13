"use client";

import { useEffect, useState } from "react";
import { adminX1Service } from "@/domain/integration-registry/service/admin-x1.service";
import type { IntegrationSyncJobStatusCounts } from "@/domain/integration-registry/schema/integration.types";
import { Activity, Loader2, AlertTriangle } from "lucide-react";

const STATUS_LABEL: Record<keyof IntegrationSyncJobStatusCounts, string> = {
  queued: "Queued",
  running: "Running",
  succeeded: "Succeeded",
  failed: "Failed",
};

const STATUS_COLOR: Record<keyof IntegrationSyncJobStatusCounts, string> = {
  queued: "bg-amber-500/20 text-amber-700 dark:text-amber-300 border-amber-500/30",
  running: "bg-sky-500/20 text-sky-700 dark:text-sky-300 border-sky-500/30",
  succeeded: "bg-emerald-500/20 text-emerald-700 dark:text-emerald-300 border-emerald-500/30",
  failed: "bg-rose-500/20 text-rose-700 dark:text-rose-300 border-rose-500/30",
};

/**
 * SyncJobStatusWidget — X-1 dashboard 의 sync job 4 status 별 count 위젯.
 * API-106 의 getStatusSummary() 백엔드. 한눈에 큐 depth / 진행 중 / 성공 / 실패 분포.
 */
export function SyncJobStatusWidget() {
  const [counts, setCounts] = useState<IntegrationSyncJobStatusCounts | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchCounts = async () => {
      setIsLoading(true);
      setError(null);
      try {
        const resp = await adminX1Service.getStatusSummary();
        setCounts(resp.sync_job_status_counts);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to load sync job status counts");
      } finally {
        setIsLoading(false);
      }
    };
    fetchCounts();
  }, []);

  return (
    <div className="glass border-border rounded-2xl p-5 space-y-3 shadow-sm">
      <div className="flex items-center gap-2">
        <Activity className="w-4 h-4 text-primary" />
        <h3 className="text-sm font-bold uppercase tracking-wider text-foreground">
          Sync Job Status
        </h3>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center py-8 text-muted-foreground">
          <Loader2 className="w-5 h-5 animate-spin mr-2" />
          <span className="text-xs">Loading counts...</span>
        </div>
      ) : error ? (
        <div className="flex items-center gap-2 py-4 text-xs text-destructive">
          <AlertTriangle className="w-4 h-4" />
          <span>{error}</span>
        </div>
      ) : counts ? (
        <div className="grid grid-cols-2 gap-2">
          {(Object.keys(STATUS_LABEL) as Array<keyof IntegrationSyncJobStatusCounts>).map((key) => (
            <div
              key={key}
              className={`flex items-center justify-between px-3 py-2 rounded-xl border ${STATUS_COLOR[key]}`}
            >
              <span className="text-[10px] font-bold uppercase tracking-wider">
                {STATUS_LABEL[key]}
              </span>
              <span className="text-lg font-black tabular-nums">{counts[key]}</span>
            </div>
          ))}
        </div>
      ) : null}
    </div>
  );
}
