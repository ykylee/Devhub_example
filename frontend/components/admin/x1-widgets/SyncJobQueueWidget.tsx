"use client";

import { useEffect, useState } from "react";
import { adminX1Service } from "@/domain/integration-registry/service/admin-x1.service";
import type { IntegrationSyncJob, IntegrationSyncJobStatus } from "@/domain/integration-registry/schema/integration.types";
import { ListTodo, Loader2, AlertTriangle, RefreshCw } from "lucide-react";
import { cn } from "@/shared/utils";

const STATUS_DOT: Record<IntegrationSyncJobStatus, string> = {
  queued: "bg-amber-500",
  running: "bg-sky-500 animate-pulse",
  succeeded: "bg-emerald-500",
  failed: "bg-rose-500",
};

const STATUS_LABEL: Record<IntegrationSyncJobStatus, string> = {
  queued: "Queued",
  running: "Running",
  succeeded: "Succeeded",
  failed: "Failed",
};

/**
 * SyncJobQueueWidget — X-1 dashboard 의 sync job 큐 (queued/running) +
 * 최근 10개 row 위젯. status filter 없이 queued+running 만 표시.
 * API-104 의 listSyncJobs() 백엔드 (status=queued|running, limit=10).
 */
export function SyncJobQueueWidget() {
  const [jobs, setJobs] = useState<IntegrationSyncJob[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchQueue = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const [queued, running] = await Promise.all([
        adminX1Service.listSyncJobs({ status: "queued", limit: 10 }),
        adminX1Service.listSyncJobs({ status: "running", limit: 10 }),
      ]);
      const merged = [...running.items, ...queued.items].sort(
        (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
      );
      setJobs(merged.slice(0, 10));
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load sync job queue");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchQueue();
  }, []);

  return (
    <div className="glass border-border rounded-2xl p-5 space-y-3 shadow-sm">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <ListTodo className="w-4 h-4 text-primary" />
          <h3 className="text-sm font-bold uppercase tracking-wider text-foreground">
            Sync Job Queue
          </h3>
        </div>
        <button
          type="button"
          aria-label="Refresh sync job queue"
          onClick={fetchQueue}
          className="text-muted-foreground hover:text-primary transition-colors"
        >
          <RefreshCw className="w-3.5 h-3.5" />
        </button>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center py-8 text-muted-foreground">
          <Loader2 className="w-5 h-5 animate-spin mr-2" />
          <span className="text-xs">Loading queue...</span>
        </div>
      ) : error ? (
        <div className="flex items-center gap-2 py-4 text-xs text-destructive">
          <AlertTriangle className="w-4 h-4" />
          <span>{error}</span>
        </div>
      ) : jobs.length === 0 ? (
        <div className="text-xs text-muted-foreground py-4 text-center">
          큐가 비어 있습니다. 모든 sync job 이 처리되었습니다.
        </div>
      ) : (
        <ul className="space-y-1.5">
          {jobs.map((j) => (
            <li
              key={j.job_id}
              className="flex items-center justify-between gap-3 px-2 py-1.5 rounded-md hover:bg-muted/30 transition-colors"
            >
              <div className="flex items-center gap-2 min-w-0 flex-1">
                <span className={cn("w-1.5 h-1.5 rounded-full shrink-0", STATUS_DOT[j.status])} />
                <span className="text-[10px] font-mono text-muted-foreground truncate">
                  {j.job_id.slice(0, 8)}
                </span>
                <span className="text-xs text-foreground/80 truncate">
                  {j.requested_by ?? "(system)"}
                </span>
              </div>
              <span className="text-[10px] text-muted-foreground shrink-0">
                {STATUS_LABEL[j.status]}
              </span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
