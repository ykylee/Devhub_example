"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { Plug, Plus } from "lucide-react";
import { integrationService } from "@/lib/services/integration.service";
import type { IntegrationProvider } from "@/lib/services/integration.types";
import { ProviderTable } from "@/components/integration/ProviderTable";
import { ProviderModal } from "@/components/integration/ProviderModal";
import { useToast } from "@/components/ui/Toast";

export default function AdminSettingsIntegrationsPage() {
  const [providers, setProviders] = useState<IntegrationProvider[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [editingProvider, setEditingProvider] = useState<IntegrationProvider | null>(null);
  const [syncingProviderID, setSyncingProviderID] = useState<string | null>(null);
  const { toast } = useToast();

  // codex hotfix #6 P1 #1 (PR #148): `useToast()` 가 매 render 마다 새 toast
  // callback 을 반환해서 `[toast]` dep 이 매번 변경 → effect 무한 재실행 → request
  // spam. 첫 mount 만 실행하는 게 의도이므로 dep 을 비우고 ESLint 를 명시적으로
  // suppress. toast 의 stale closure 는 빈 페이지의 1회 호출에서만 사용되므로
  // 실 영향 없음.
  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(() => {
    let cancelled = false;
    (async () => {
      setIsLoading(true);
      try {
        const result = await integrationService.listProviders();
        if (!cancelled) setProviders(result.data);
      } catch (error) {
        if (!cancelled) {
          console.error("[admin/settings/integrations] load failed:", error);
          toast("integration provider 목록을 불러오지 못했습니다.", "error");
        }
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const handleSaved = (saved: IntegrationProvider) => {
    setProviders((prev) => {
      const idx = prev.findIndex((p) => p.provider_id === saved.provider_id);
      if (idx >= 0) {
        const next = [...prev];
        next[idx] = saved;
        return next;
      }
      return [saved, ...prev];
    });
    toast(`Provider '${saved.display_name}' 이(가) 저장되었습니다.`, "success");
  };

  const handleSync = async (provider: IntegrationProvider) => {
    setSyncingProviderID(provider.provider_id);
    try {
      // API-72 는 {status: "accepted", job_id} 만 반환 (provider envelope 없음).
      // 실 sync_status 는 backend job 결과 — UI 는 즉시 "requested" 로 optimistic
      // update + 후속 list refresh 로 정합 (codex hotfix #6 P1 #2, PR #148).
      const result = await integrationService.syncProvider(provider.provider_id);
      setProviders((prev) =>
        prev.map((p) =>
          p.provider_id === provider.provider_id ? { ...p, sync_status: "requested" } : p,
        ),
      );
      toast(`Sync triggered: ${provider.display_name} (job ${result.job_id})`, "success");
    } catch (error) {
      console.error("[admin/settings/integrations] sync failed:", error);
      toast("sync 호출에 실패했습니다.", "error");
    } finally {
      setSyncingProviderID(null);
    }
  };

  return (
    <div className="space-y-8">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="flex items-center gap-3">
          <div className="w-12 h-12 bg-orange-500/10 border border-orange-500/30 rounded-2xl flex items-center justify-center">
            <Plug className="w-6 h-6 text-orange-400" />
          </div>
          <div>
            <h2 className="text-xl font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">
              Integration <span className="text-orange-400">Providers</span>
            </h2>
            <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest mt-1">
              외부 시스템 연동 — webhook / pull / agent 어댑터 등록
            </p>
          </div>
        </motion.div>

        <button
          type="button"
          onClick={() => setShowCreate(true)}
          className="inline-flex items-center gap-2 px-6 py-3 rounded-2xl bg-orange-500 text-white font-black uppercase tracking-widest text-[10px] shadow-xl shadow-orange-500/20 hover:scale-[1.02] active:scale-[0.98] transition-all"
        >
          <Plus className="w-4 h-4" />
          Register Provider
        </button>
      </div>

      {isLoading ? (
        <div className="flex flex-col items-center justify-center py-32 gap-4">
          <div className="w-12 h-12 border-4 border-orange-500/20 border-t-orange-500 rounded-full animate-spin" />
          <p className="text-muted-foreground font-bold animate-pulse uppercase tracking-[0.3em] text-[10px]">
            Loading Providers...
          </p>
        </div>
      ) : (
        <ProviderTable
          items={providers}
          onEdit={setEditingProvider}
          onSync={handleSync}
          syncingProviderID={syncingProviderID}
        />
      )}

      {showCreate && (
        <ProviderModal
          onClose={() => setShowCreate(false)}
          onSaved={handleSaved}
        />
      )}

      {editingProvider && (
        <ProviderModal
          initial={editingProvider}
          onClose={() => setEditingProvider(null)}
          onSaved={handleSaved}
        />
      )}
    </div>
  );
}
