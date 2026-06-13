"use client";

import { useEffect, useState } from "react";
import { Loader2, AlertTriangle, HeartPulse } from "lucide-react";
import { cn } from "@/shared/utils";

interface ProviderHealth {
  provider_id: string;
  provider_key: string;
  display_name: string;
  provider_type: string;
  healthy: boolean;
  last_checked_at: string;
}

/**
 * ProviderHealthWidget — X-1 dashboard 의 integration provider health 위젯.
 * backend 의 `RuntimeSnapshotProvider` 또는 `/admin/integrations/provider-health`
 * 의 향후 endpoint 가 추가될 자리 placeholder. 현재는 빈 상태 (provider health
 * endpoint 가 별도 carve).
 *
 * ADR-0032 §3 의 정공법 — provider health endpoint 는 v0.1.1 milestone 의 별도
 * carve (현 sprint X-1 의 optional API-107). 본 widget 은 placeholder rendering
 * 만 제공하여 frontend 정합을 유지한다.
 */
export function ProviderHealthWidget() {
  const [providers, setProviders] = useState<ProviderHealth[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error] = useState<string | null>(null);
  const [endpointReady] = useState(false);

  useEffect(() => {
    // Provider health endpoint (API-107) 는 v0.1.1 의 별도 carve.
    // 본 widget 은 placeholder 정합만 유지 (loading false, empty state).
    setIsLoading(false);
    setProviders([]);
  }, []);

  return (
    <div className="glass border-border rounded-2xl p-5 space-y-3 shadow-sm">
      <div className="flex items-center gap-2">
        <HeartPulse className="w-4 h-4 text-primary" />
        <h3 className="text-sm font-bold uppercase tracking-wider text-foreground">
          Provider Health
        </h3>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center py-8 text-muted-foreground">
          <Loader2 className="w-5 h-5 animate-spin mr-2" />
          <span className="text-xs">Loading provider health...</span>
        </div>
      ) : error ? (
        <div className="flex items-center gap-2 py-4 text-xs text-destructive">
          <AlertTriangle className="w-4 h-4" />
          <span>{error}</span>
        </div>
      ) : !endpointReady ? (
        <div className="text-xs text-muted-foreground py-4 text-center">
          <p>Provider health endpoint 미구현.</p>
          <p className="mt-1 text-[10px]">
            ADR-0032 §3 carve — v0.1.1 의 후속 sprint 에서 API-107 추가 예정.
          </p>
        </div>
      ) : providers.length === 0 ? (
        <div className="text-xs text-muted-foreground py-4 text-center">
          등록된 provider 가 없습니다.
        </div>
      ) : (
        <ul className="space-y-1.5">
          {providers.map((p) => (
            <li
              key={p.provider_id}
              className="flex items-center justify-between gap-3 px-2 py-1.5 rounded-md hover:bg-muted/30 transition-colors"
            >
              <div className="flex items-center gap-2 min-w-0 flex-1">
                <span
                  className={cn(
                    "w-1.5 h-1.5 rounded-full shrink-0",
                    p.healthy ? "bg-emerald-500" : "bg-rose-500",
                  )}
                />
                <span className="text-xs text-foreground/80 truncate">{p.display_name}</span>
              </div>
              <span className="text-[10px] text-muted-foreground shrink-0">
                {p.provider_type}
              </span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
