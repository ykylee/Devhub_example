"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { motion } from "framer-motion";
import { Inbox, ArrowRight, AlertCircle } from "lucide-react";
import { format, parseISO } from "date-fns";
import { devRequestService } from "@/lib/services/dev_request.service";
import { DevRequest } from "@/lib/services/dev_request.types";

const MAX_PREVIEW = 5;

/**
 * 담당자 dashboard 의 "내 대기 의뢰" 위젯.
 * Backend 가 일반 role (developer/manager) 호출 시 row-level filter 로 assignee == actor.login 만 반환하므로
 * 추가 client-side filter 는 불필요. system_admin 은 전체를 보지만 본 위젯은 "내 대기" 의도라 그대로 표시.
 */
export function MyPendingDevRequestsWidget() {
  const [items, setItems] = useState<DevRequest[]>([]);
  const [total, setTotal] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const result = await devRequestService.getMyPending();
        if (cancelled) return;
        setItems(result.data.slice(0, MAX_PREVIEW));
        setTotal(result.total);
      } catch (e) {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : "Failed to load dev requests");
        }
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      className="glass border-border rounded-3xl p-6 space-y-4"
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-orange-500/10 border border-orange-500/30 rounded-2xl flex items-center justify-center">
            <Inbox className="w-5 h-5 text-orange-400" />
          </div>
          <div>
            <h3 className="text-sm font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">
              내 대기 의뢰
            </h3>
            <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest mt-0.5">
              {total} pending / in_review
            </p>
          </div>
        </div>
        <Link
          href="/admin/settings/dev-requests"
          className="flex items-center gap-1 text-[10px] font-black text-orange-400 hover:text-orange-300 uppercase tracking-widest"
        >
          전체 보기 <ArrowRight className="w-3 h-3" />
        </Link>
      </div>

      {isLoading && (
        <div className="py-8 flex items-center justify-center">
          <div className="w-6 h-6 border-2 border-orange-500/20 border-t-orange-500 rounded-full animate-spin" />
        </div>
      )}

      {!isLoading && error && (
        <div className="flex items-center gap-2 p-3 bg-rose-500/10 border border-rose-500/20 rounded-2xl text-[11px] text-rose-400">
          <AlertCircle className="w-4 h-4" />
          {error}
        </div>
      )}

      {!isLoading && !error && items.length === 0 && (
        <div className="py-6 text-center">
          <p className="text-[11px] text-muted-foreground">대기 중인 의뢰가 없습니다.</p>
        </div>
      )}

      {!isLoading && !error && items.length > 0 && (
        <ul className="space-y-2">
          {items.map((dr) => (
            <li key={dr.id}>
              <Link
                href="/admin/settings/dev-requests"
                className="flex items-center gap-3 p-3 rounded-2xl bg-muted/20 border border-border/40 hover:bg-muted/30 hover:border-orange-500/30 transition-all group"
              >
                <div className="flex-1 min-w-0">
                  <p className="text-xs font-bold text-foreground dark:text-primary-foreground truncate group-hover:text-orange-300">
                    {dr.title}
                  </p>
                  <p className="text-[10px] text-muted-foreground truncate">
                    {dr.source_system} · {formatTime(dr.received_at)}
                  </p>
                </div>
                <ArrowRight className="w-3.5 h-3.5 text-muted-foreground/40 group-hover:text-orange-400 transition-colors" />
              </Link>
            </li>
          ))}
        </ul>
      )}
    </motion.div>
  );
}

function formatTime(value: string): string {
  if (!value) return "—";
  try {
    return format(parseISO(value), "MMM d, HH:mm");
  } catch {
    return value;
  }
}
