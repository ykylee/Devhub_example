"use client";

import { useEffect, useState, useMemo } from "react";
import { motion } from "framer-motion";
import { devRequestService } from "@/lib/services/dev_request.service";
import { DevRequest, DevRequestStatus } from "@/lib/services/dev_request.types";
import { DevRequestTable } from "@/components/dev-request/DevRequestTable";
import { DevRequestDetailModal } from "@/components/dev-request/DevRequestDetailModal";
import { useToast } from "@/components/ui/Toast";
import { useStore } from "@/lib/store";
import { isSystemAdmin } from "@/lib/auth/role-routing";
import { DashboardHeader } from "@/components/ui/DashboardHeader";
import { FilterBar } from "@/components/ui/FilterBar";

const STATUS_OPTIONS: { label: string; value: DevRequestStatus | "all" }[] = [
  { label: "All Requests", value: "all" },
  { label: "Pending", value: "pending" },
  { label: "In Review", value: "in_review" },
  { label: "Registered", value: "registered" },
  { label: "Rejected", value: "rejected" },
];

export default function MyDevRequestsPage() {
  const [items, setItems] = useState<DevRequest[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState<DevRequestStatus | "all">("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [selected, setSelected] = useState<DevRequest | null>(null);
  const { toast } = useToast();
  const actor = useStore((s) => s.actor);
  const allowSystemAdmin = isSystemAdmin(actor?.role);

  const filteredItems = useMemo(() => {
    const q = searchQuery.toLowerCase().trim();
    return items.filter(item => {
      const matchesQuery = !q || 
        item.title.toLowerCase().includes(q) ||
        item.details?.toLowerCase().includes(q) ||
        item.id.toLowerCase().includes(q);
      const matchesStatus = statusFilter === "all" || item.status === statusFilter;
      return matchesQuery && matchesStatus;
    });
  }, [items, searchQuery, statusFilter]);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setIsLoading(true);
      try {
        const result = await devRequestService.list({
          // Backend might support filtering, but client-side filter is fine for personal list
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
  }, []);

  const handleChanged = (updated: DevRequest) => {
    setItems((prev) => prev.map((r) => (r.id === updated.id ? updated : r)));
    toast(`의뢰 상태가 ${updated.status} 로 변경되었습니다.`, "success");
  };

  return (
    <div className="space-y-10 pb-20">
      <DashboardHeader 
        titlePrefix="내"
        titleGradient="개발 의뢰"
        subtitle="본인 담당 의뢰만 표시 · 외부 시스템 의뢰 처리"
      />

      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
      >
        <FilterBar 
          onSearch={setSearchQuery}
          onFilterChange={(val) => setStatusFilter(val as DevRequestStatus | "all")}
          activeFilter={statusFilter}
          filterOptions={STATUS_OPTIONS}
          placeholder="Search by ID, title, or description..."
        />
      </motion.div>

      {isLoading ? (
        <div className="flex flex-col items-center justify-center py-32 gap-4">
          <div className="w-12 h-12 border-4 border-primary/20 border-t-primary rounded-full animate-spin" />
          <p className="text-muted-foreground font-bold animate-pulse uppercase tracking-[0.3em] text-[10px]">
            Loading Dev Requests...
          </p>
        </div>
      ) : (
        <DevRequestTable items={filteredItems} onSelect={setSelected} />
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
