"use client";

import { useEffect, useMemo, useState } from "react";
import { motion } from "framer-motion";
import { Link2, Plus } from "lucide-react";
import { integrationService } from "@/lib/services/integration.service";
import type { IntegrationBinding, IntegrationProvider, IntegrationScopeType } from "@/lib/services/integration.types";
import { BindingsTable } from "@/components/integration/BindingsTable";
import { CreateBindingModal } from "@/components/integration/CreateBindingModal";
import { useToast } from "@/components/ui/Toast";

export default function AdminSettingsIntegrationBindingsPage() {
  const [bindings, setBindings] = useState<IntegrationBinding[]>([]);
  const [providers, setProviders] = useState<IntegrationProvider[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [scopeFilter, setScopeFilter] = useState<"" | IntegrationScopeType>("");
  const { toast } = useToast();

  // codex hotfix #6 P1 #1 reference (PR #148): toast 는 effect dep 에서 제외.
  // 진입 시 1회 로드 + scopeFilter 변경 시 재로드. providers 는 modal dropdown
  // 옵션 + table 의 display_name 매핑에 동시에 사용되므로 함께 fetch.
  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(() => {
    let cancelled = false;
    (async () => {
      setIsLoading(true);
      try {
        const [bindingsResp, providersResp] = await Promise.all([
          integrationService.listBindings(scopeFilter ? { scope_type: scopeFilter } : {}),
          integrationService.listProviders(),
        ]);
        if (!cancelled) {
          setBindings(bindingsResp.data);
          setProviders(providersResp.data);
        }
      } catch (error) {
        if (!cancelled) {
          console.error("[admin/settings/integration-bindings] load failed:", error);
          toast("binding 또는 provider 목록을 불러오지 못했습니다.", "error");
        }
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [scopeFilter]);

  const providersByID = useMemo(() => {
    const map: Record<string, IntegrationProvider> = {};
    for (const p of providers) map[p.provider_id] = p;
    return map;
  }, [providers]);

  const handleCreated = (binding: IntegrationBinding) => {
    const provider = providersByID[binding.provider_id];
    const label = provider?.display_name ?? binding.provider_id;
    // codex hotfix #8 P2 #1 — 활성 scopeFilter 와 다른 scope_type 의 신규 binding 은
    // 현재 view 에 포함되지 않으므로 table prepend 안 함. UX 일관성 (사용자가
    // "필터된 view 안에 갑자기 다른 scope row 가 나타나는" 혼란 방지).
    if (scopeFilter && binding.scope_type !== scopeFilter) {
      toast(
        `Binding 생성됨 — 현재 '${scopeFilter}' 필터에 가려져 있습니다. All scopes 로 전환하면 노출됩니다.`,
        "info",
      );
      return;
    }
    setBindings((prev) => [binding, ...prev]);
    toast(`Binding '${binding.scope_id} → ${label}' 이(가) 생성되었습니다.`, "success");
  };

  return (
    <div className="space-y-8">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="flex items-center gap-3">
          <div className="w-12 h-12 bg-orange-500/10 border border-orange-500/30 rounded-2xl flex items-center justify-center">
            <Link2 className="w-6 h-6 text-orange-400" />
          </div>
          <div>
            <h2 className="text-xl font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">
              Integration <span className="text-orange-400">Bindings</span>
            </h2>
            <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest mt-1">
              provider 와 application/project scope 매핑
            </p>
          </div>
        </motion.div>

        <div className="flex items-center gap-3">
          <select
            aria-label="Scope filter"
            value={scopeFilter}
            onChange={(e) => setScopeFilter(e.target.value as "" | IntegrationScopeType)}
            className="px-4 py-3 rounded-xl bg-muted/30 border border-border text-foreground dark:text-primary-foreground text-xs font-bold uppercase tracking-widest focus:outline-none focus:border-orange-400"
          >
            <option value="">All scopes</option>
            <option value="application">application</option>
            <option value="project">project</option>
          </select>
          <button
            type="button"
            onClick={() => setShowCreate(true)}
            disabled={providers.length === 0}
            className="inline-flex items-center gap-2 px-6 py-3 rounded-2xl bg-orange-500 text-white font-black uppercase tracking-widest text-[10px] shadow-xl shadow-orange-500/20 hover:scale-[1.02] active:scale-[0.98] transition-all disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Plus className="w-4 h-4" />
            Create Binding
          </button>
        </div>
      </div>

      {providers.length === 0 && !isLoading && (
        <div className="glass border border-yellow-500/30 bg-yellow-500/5 rounded-2xl p-4 text-xs text-yellow-300 font-bold">
          등록된 provider 가 없어 binding 을 생성할 수 없습니다. 먼저{" "}
          <a href="/admin/settings/integrations" className="underline decoration-orange-400 hover:text-orange-300">
            Integrations
          </a>{" "}
          탭에서 provider 를 등록하세요.
        </div>
      )}

      {isLoading ? (
        <div className="flex flex-col items-center justify-center py-32 gap-4">
          <div className="w-12 h-12 border-4 border-orange-500/20 border-t-orange-500 rounded-full animate-spin" />
          <p className="text-muted-foreground font-bold animate-pulse uppercase tracking-[0.3em] text-[10px]">
            Loading Bindings...
          </p>
        </div>
      ) : (
        <BindingsTable items={bindings} providersByID={providersByID} />
      )}

      {showCreate && (
        <CreateBindingModal
          providers={providers}
          onClose={() => setShowCreate(false)}
          onCreated={handleCreated}
        />
      )}
    </div>
  );
}
