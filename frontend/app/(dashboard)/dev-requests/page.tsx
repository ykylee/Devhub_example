"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { Inbox } from "lucide-react";
import { devRequestService } from "@/lib/services/dev_request.service";
import { DevRequest, DevRequestStatus } from "@/lib/services/dev_request.types";
import { DevRequestTable } from "@/components/dev-request/DevRequestTable";
import { DevRequestDetailModal } from "@/components/dev-request/DevRequestDetailModal";
import { useToast } from "@/components/ui/Toast";
import { useStore } from "@/lib/store";
import { isSystemAdmin } from "@/lib/auth/role-routing";
import { cn } from "@/lib/utils";

/**
 * 일반 사용자용 Dev Requests 페이지 (codex PR #125 P1 hotfix, sprint claude/work_260515-k).
 *
 * /admin/settings/dev-requests 가 system_admin only gate 라서 담당자 본인이 진입할 수 없는
 * 문제를 해결한다. 본 페이지는 admin layout 외에 있어 developer/manager 도 진입 가능하며,
 * backend 의 row-level filter (assignee == actor.login) 가 본인 의뢰만 노출한다.
 *
 * system_admin 도 본 페이지 진입 시 본인 의뢰만 보이는 게 자연스러움. 전체 관리는
 * /admin/settings/dev-requests 유지.
 */

const STATUS_FILTERS: { label: string; value: DevRequestStatus | "all" }[] = [
  { label: "All", value: "all" },
  { label: "Pending", value: "pending" },
  { label: "In Review", value: "in_review" },
  { label: "Registered", value: "registered" },
  { label: "Rejected", value: "rejected" },
];

export default function MyDevRequestsPage() {
  const [items, setItems] = useState<DevRequest[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState<DevRequestStatus | "all">("all");
  const [selected, setSelected] = useState<DevRequest | null>(null);
  const { toast } = useToast();
  const actor = useStore((s) => s.actor);
  const allowSystemAdmin = isSystemAdmin(actor?.role);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setIsLoading(true);
      try {
        const result = await devRequestService.list({
          status: statusFilter === "all" ? undefined : statusFilter,
          // assignee_user_id 는 보내지 않아도 backend 가 actor.login 으로 row filter.
        });
        if (!cancelled) setItems(result.data);
      } catch (error) {
        if (!cancelled) {
          console.error("[dev-requests] load failed:", error);
        }
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [statusFilter]);

  const handleChanged = (updated: DevRequest) => {
    setItems((prev) => prev.map((r) => (r.id === updated.id ? updated : r)));
    toast(`의뢰 상태가 ${updated.status} 로 변경되었습니다.`, "success");
  };

  return (
    <div className="space-y-10 pb-20">
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        className="flex flex-col md:flex-row md:items-end justify-between gap-6"
      >
        <div className="flex items-center gap-4">
          <div className="w-12 h-12 bg-orange-500/10 border border-orange-500/30 rounded-2xl flex items-center justify-center">
            <Inbox className="w-6 h-6 text-orange-400" />
          </div>
          <div>
            <h1 className="text-3xl font-black text-foreground dark:text-primary-foreground tracking-tighter uppercase">
              내 <span className="text-orange-400">개발 의뢰</span>
            </h1>
            <p className="text-muted-foreground font-bold text-xs uppercase tracking-widest mt-2">
              본인 담당 의뢰만 표시 · 외부 시스템 의뢰 처리
            </p>
          </div>
        </div>
      </motion.div>

      <div className="flex flex-wrap gap-2">
        {STATUS_FILTERS.map((f) => (
          <button
            key={f.value}
            type="button"
            onClick={() => setStatusFilter(f.value)}
            className={cn(
              "px-4 py-2 rounded-xl border text-[10px] font-black uppercase tracking-widest transition-all",
              statusFilter === f.value
                ? "bg-orange-500/10 border-orange-500/40 text-orange-400"
                : "bg-muted/20 border-border/60 text-muted-foreground hover:bg-muted/40",
            )}
          >
            {f.label}
          </button>
        ))}
      </div>

      {isLoading ? (
        <div className="flex flex-col items-center justify-center py-32 gap-4">
          <div className="w-12 h-12 border-4 border-orange-500/20 border-t-orange-500 rounded-full animate-spin" />
          <p className="text-muted-foreground font-bold animate-pulse uppercase tracking-[0.3em] text-[10px]">
            Loading Dev Requests...
          </p>
        </div>
      ) : (
        <DevRequestTable items={items} onSelect={setSelected} />
      )}

      {selected && (
        <DevRequestDetailModal
          request={selected}
          isSystemAdmin={allowSystemAdmin}
          onClose={() => setSelected(null)}
          onChanged={handleChanged}
        />
      )}
    </div>
  );
}
