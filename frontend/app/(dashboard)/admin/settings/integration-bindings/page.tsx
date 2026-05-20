"use client";

import { useEffect, useMemo, useState } from "react";
import { motion } from "framer-motion";
import { Link2, Plus, ChevronLeft, ChevronRight } from "lucide-react";
import { integrationService } from "@/lib/services/integration.service";
import type { IntegrationBinding, IntegrationProvider, IntegrationScopeType } from "@/lib/services/integration.types";
import { BindingsTable } from "@/components/integration/BindingsTable";
import { CreateBindingModal } from "@/components/integration/CreateBindingModal";
import { EditBindingModal } from "@/components/integration/EditBindingModal";
import { DestructiveConfirmModal } from "@/components/ui/DestructiveConfirmModal";
import { useToast } from "@/components/ui/Toast";

export default function AdminSettingsIntegrationBindingsPage() {
  const [bindings, setBindings] = useState<IntegrationBinding[]>([]);
  const [providers, setProviders] = useState<IntegrationProvider[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [editingBinding, setEditingBinding] = useState<IntegrationBinding | null>(null);
  const [deletingBinding, setDeletingBinding] = useState<IntegrationBinding | null>(null);
  
  const [scopeFilter, setScopeFilter] = useState<"" | IntegrationScopeType>("");
  const [offset, setOffset] = useState(0);
  const [total, setTotal] = useState(0);
  const LIMIT = 20;

  const { toast } = useToast();

  const load = async () => {
    setIsLoading(true);
    try {
      const [bindingsResp, providersResp] = await Promise.all([
        integrationService.listBindings({ 
          scope_type: scopeFilter || undefined,
          limit: LIMIT,
          offset,
        }),
        integrationService.listProviders(),
      ]);
      setBindings(bindingsResp.data);
      setTotal(bindingsResp.total);
      setProviders(providersResp.data);
    } catch (error) {
      console.error("[admin/settings/integration-bindings] load failed:", error);
      toast("binding 또는 provider 목록을 불러오지 못했습니다.", "error");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, [scopeFilter, offset]);

  const providersByID = useMemo(() => {
    const map: Record<string, IntegrationProvider> = {};
    for (const p of providers) map[p.provider_id] = p;
    return map;
  }, [providers]);

  const handleCreated = (binding: IntegrationBinding) => {
    load(); // Refresh to get correct total and order
    const provider = providersByID[binding.provider_id];
    const label = provider?.display_name ?? binding.provider_id;
    toast(`Binding '${binding.scope_id} → ${label}' 이(가) 생성되었습니다.`, "success");
  };

  const handleUpdated = (updated: IntegrationBinding) => {
    setBindings(prev => prev.map(b => b.binding_id === updated.binding_id ? updated : b));
    const provider = providersByID[updated.provider_id];
    const label = provider?.display_name ?? updated.provider_id;
    toast(`Binding '${updated.scope_id} → ${label}' 정보가 수정되었습니다.`, "success");
  };

  const handleDeleted = async () => {
    if (!deletingBinding) return;
    try {
      await integrationService.deleteBinding(deletingBinding.binding_id);
      toast(`Binding 을 삭제했습니다.`, "success");
      // codex P2 (PR #251 review): pagination offset clamp 후 reload.
      // 마지막 페이지에서 1건 삭제 시 total < offset 가 되면 빈 페이지 표시 위험.
      // setOffset trigger 시 useEffect 가 load() 자동 호출하므로 명시 load 생략.
      const newTotal = Math.max(0, total - 1);
      if (offset > 0 && offset >= newTotal) {
        const clampedOffset = Math.max(0, Math.floor((newTotal - 1) / LIMIT) * LIMIT);
        setOffset(clampedOffset);
      } else {
        load();
      }
    } catch (err) {
      toast("Binding 삭제에 실패했습니다.", "error");
    } finally {
      setDeletingBinding(null);
    }
  };

  return (
    <div className="space-y-8">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="flex items-center gap-3">
          <div className="w-12 h-12 bg-accent/10 border border-accent/30 rounded-2xl flex items-center justify-center">
            <Link2 className="w-6 h-6 text-accent" />
          </div>
          <div>
            <h2 className="text-xl font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">
              Integration <span className="text-accent">Bindings</span>
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
            onChange={(e) => {
              setScopeFilter(e.target.value as "" | IntegrationScopeType);
              setOffset(0);
            }}
            className="px-4 py-3 rounded-xl bg-muted/30 border border-border text-foreground dark:text-primary-foreground text-xs font-bold uppercase tracking-widest focus:outline-none focus:border-accent"
          >
            <option value="">All scopes</option>
            <option value="application">application</option>
            <option value="project">project</option>
          </select>
          <button
            type="button"
            onClick={() => setShowCreate(true)}
            disabled={providers.length === 0}
            className="inline-flex items-center gap-2 px-6 py-3 rounded-2xl bg-accent text-white font-black uppercase tracking-widest text-[10px] shadow-xl shadow-accent/20 hover:scale-[1.02] active:scale-[0.98] transition-all disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Plus className="w-4 h-4" />
            Create Binding
          </button>
        </div>
      </div>

      {providers.length === 0 && !isLoading && (
        <div className="glass border border-warning/30 bg-warning/5 rounded-2xl p-4 text-xs text-yellow-300 font-bold">
          등록된 provider 가 없어 binding 을 생성할 수 없습니다. 먼저{" "}
          <a href="/admin/settings/integrations" className="underline decoration-accent hover:text-orange-300">
            Integrations
          </a>{" "}
          탭에서 provider 를 등록하세요.
        </div>
      )}

      {isLoading ? (
        <div className="flex flex-col items-center justify-center py-32 gap-4">
          <div className="w-12 h-12 border-4 border-accent/20 border-t-accent rounded-full animate-spin" />
          <p className="text-muted-foreground font-bold animate-pulse uppercase tracking-[0.3em] text-[10px]">
            Loading Bindings...
          </p>
        </div>
      ) : (
        <div className="space-y-6">
          <BindingsTable 
            items={bindings} 
            providersByID={providersByID} 
            onEdit={setEditingBinding}
            onDelete={setDeletingBinding}
          />

          {total > LIMIT && (
            <div className="flex items-center justify-between px-2">
              <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
                Showing {offset + 1} - {Math.min(offset + LIMIT, total)} of {total} bindings
              </p>
              <div className="flex items-center gap-2">
                <button
                  aria-label="Previous page"
                  onClick={() => setOffset(Math.max(0, offset - LIMIT))}
                  disabled={offset === 0}
                  className="p-2 rounded-xl glass border border-border hover:bg-muted/50 disabled:opacity-30 disabled:cursor-not-allowed transition-all"
                >
                  <ChevronLeft className="w-4 h-4" />
                </button>
                <button
                  aria-label="Next page"
                  onClick={() => setOffset(offset + LIMIT)}
                  disabled={offset + LIMIT >= total}
                  className="p-2 rounded-xl glass border border-border hover:bg-muted/50 disabled:opacity-30 disabled:cursor-not-allowed transition-all"
                >
                  <ChevronRight className="w-4 h-4" />
                </button>
              </div>
            </div>
          )}
        </div>
      )}

      {showCreate && (
        <CreateBindingModal
          providers={providers}
          onClose={() => setShowCreate(false)}
          onCreated={handleCreated}
        />
      )}

      {editingBinding && (
        <EditBindingModal
          binding={editingBinding}
          providers={providers}
          onClose={() => setEditingBinding(null)}
          onUpdated={handleUpdated}
        />
      )}

      {deletingBinding && (
        <DestructiveConfirmModal
          isOpen={!!deletingBinding}
          onClose={() => setDeletingBinding(null)}
          onConfirm={handleDeleted}
          title="Delete Binding"
          description={`Are you sure you want to delete the binding for '${deletingBinding.scope_id}'? This will stop integration data flow for this scope.`}
          confirmText="Delete Binding"
        />
      )}
    </div>
  );
}
