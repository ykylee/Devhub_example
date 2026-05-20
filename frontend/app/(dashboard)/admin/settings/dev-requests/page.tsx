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
import { FilterBar } from "@/components/ui/FilterBar";

const STATUS_OPTIONS = [
  { label: "All Requests", value: "all" },
  { label: "Pending", value: "pending" },
  { label: "In Review", value: "in_review" },
  { label: "Registered", value: "registered" },
  { label: "Rejected", value: "rejected" },
  { label: "Closed", value: "closed" },
];

export default function AdminSettingsDevRequestsPage() {
  const [items, setItems] = useState<DevRequest[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState<DevRequestStatus | "all">("all");
  const [sourceSystem, setSourceSystem] = useState("");
  const [selected, setSelected] = useState<DevRequest | null>(null);
  const { toast } = useToast();
  const actor = useStore((s) => s.actor);
  const allowSystemAdmin = isSystemAdmin(actor?.role);

  const refresh = async () => {
    setIsLoading(true);
    try {
      const result = await devRequestService.list({
        status: statusFilter === "all" ? undefined : statusFilter,
        source_system: sourceSystem.trim() || undefined,
      });
      setItems(result.data);
    } catch (error) {
      console.error("[admin/settings/dev-requests] load failed:", error);
      if ((error as { status?: number })?.status === 501) {
        toast("Backend dev_requests API not implemented yet (501). Showing empty list.", "warning");
      } else {
        toast("의뢰 목록을 불러오지 못했습니다.", "error");
      }
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setIsLoading(true);
      try {
        const result = await devRequestService.list({
          status: statusFilter === "all" ? undefined : statusFilter,
          source_system: sourceSystem.trim() || undefined,
        });
        if (!cancelled) setItems(result.data);
      } catch (error) {
        if (!cancelled) {
          console.error("[admin/settings/dev-requests] load failed:", error);
        }
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [statusFilter, sourceSystem]);

  const handleChanged = (updated: DevRequest) => {
    setItems((prev) => prev.map((r) => (r.id === updated.id ? updated : r)));
    toast(`의뢰 상태가 ${updated.status} 로 변경되었습니다.`, "success");
    void refresh();
  };

  return (
    <div className="space-y-8">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="flex items-center gap-3">
          <div className="w-12 h-12 bg-accent/10 border border-accent/30 rounded-2xl flex items-center justify-center">
            <Inbox className="w-6 h-6 text-accent" />
          </div>
          <div>
            <h2 className="text-xl font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">
              Dev <span className="text-accent">Requests</span>
            </h2>
            <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest mt-1">
              외부 시스템에서 들어온 개발 의뢰 — 등록/거절/재할당
            </p>
          </div>
        </motion.div>
      </div>

      <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }}>
        <FilterBar 
          onSearch={setSourceSystem}
          onFilterChange={(val) => setStatusFilter(val as DevRequestStatus | "all")}
          activeFilter={statusFilter}
          filterOptions={STATUS_OPTIONS}
          placeholder="Search by source system name..."
          searchLabel="Search dev requests"
        />
      </motion.div>

      {isLoading ? (
        <div className="flex flex-col items-center justify-center py-32 gap-4">
          <div className="w-12 h-12 border-4 border-accent/20 border-t-accent rounded-full animate-spin" />
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
