"use client";

import { useState, FormEvent } from "react";
import { Link2, X } from "lucide-react";
import { motion } from "framer-motion";
import { integrationService } from "@/lib/services/integration.service";
import { ApiError } from "@/lib/services/api-client";
import type {
  IntegrationBinding,
  IntegrationPolicy,
  IntegrationProvider,
  IntegrationScopeType,
} from "@/lib/services/integration.types";

interface CreateBindingModalProps {
  providers: IntegrationProvider[];
  onClose: () => void;
  onCreated: (binding: IntegrationBinding) => void;
}

const scopeTypeOptions: IntegrationScopeType[] = ["application", "project"];
const policyOptions: IntegrationPolicy[] = ["summary_only", "execution_system"];

export function CreateBindingModal({ providers, onClose, onCreated }: CreateBindingModalProps) {
  const [scopeType, setScopeType] = useState<IntegrationScopeType>("application");
  const [scopeID, setScopeID] = useState("");
  const [providerID, setProviderID] = useState(providers[0]?.provider_id ?? "");
  const [externalKey, setExternalKey] = useState("");
  const [policy, setPolicy] = useState<IntegrationPolicy>("execution_system");

  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    if (!scopeID.trim()) {
      setError("scope_id 는 필수입니다.");
      return;
    }
    if (!providerID) {
      setError("provider 를 선택하세요.");
      return;
    }
    if (!externalKey.trim()) {
      setError("external_key 는 필수입니다.");
      return;
    }
    setSubmitting(true);
    try {
      const created = await integrationService.createBinding({
        scope_type: scopeType,
        scope_id: scopeID.trim(),
        provider_id: providerID,
        external_key: externalKey.trim(),
        policy,
      });
      onCreated(created);
      onClose();
    } catch (err) {
      // backend §15.2 — 409 integration_binding_conflict / 422 integration_policy_violation
      if (err instanceof ApiError) {
        const payload = err.payload as { code?: string; error?: string } | null;
        if (err.status === 409 && payload?.code === "integration_binding_conflict") {
          setError("이미 동일한 (scope, provider, external_key) binding 이 있거나 provider 가 존재하지 않습니다.");
          return;
        }
        if (err.status === 422 && payload?.code === "integration_policy_violation") {
          setError("지원되지 않는 policy 값입니다.");
          return;
        }
      }
      const msg = err instanceof Error ? err.message : "binding 생성에 실패했습니다.";
      setError(msg);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
    >
      <motion.div
        initial={{ opacity: 0, y: 20, scale: 0.98 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        exit={{ opacity: 0, y: 20, scale: 0.98 }}
        className="glass border-border rounded-3xl w-full max-w-2xl max-h-[90vh] overflow-y-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between p-6 border-b border-border/60">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-orange-500/10 flex items-center justify-center border border-orange-500/20">
              <Link2 className="w-5 h-5 text-orange-400" />
            </div>
            <div>
              <h3 className="text-lg font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">
                Create Binding
              </h3>
              <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-widest mt-0.5">
                provider 와 application/project scope 를 매핑합니다
              </p>
            </div>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="p-2 rounded-lg hover:bg-muted/50 transition-colors"
            aria-label="Close"
          >
            <X className="w-4 h-4 text-muted-foreground" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-5">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label htmlFor="scope_type" className="block text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2">
                Scope Type *
              </label>
              <select
                id="scope_type"
                value={scopeType}
                onChange={(e) => setScopeType(e.target.value as IntegrationScopeType)}
                className="w-full px-4 py-3 rounded-xl bg-muted/30 border border-border text-foreground dark:text-primary-foreground text-sm focus:outline-none focus:border-orange-400"
              >
                {scopeTypeOptions.map((s) => (
                  <option key={s} value={s}>{s}</option>
                ))}
              </select>
            </div>
            <div>
              <label htmlFor="scope_id" className="block text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2">
                Scope ID *
              </label>
              <input
                id="scope_id"
                type="text"
                value={scopeID}
                onChange={(e) => setScopeID(e.target.value)}
                placeholder="APP-001 또는 PROJ-001"
                className="w-full px-4 py-3 rounded-xl bg-muted/30 border border-border text-foreground dark:text-primary-foreground font-mono text-sm focus:outline-none focus:border-orange-400"
                required
              />
            </div>
          </div>

          <div>
            <label htmlFor="provider_id" className="block text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2">
              Provider *
            </label>
            <select
              id="provider_id"
              value={providerID}
              onChange={(e) => setProviderID(e.target.value)}
              className="w-full px-4 py-3 rounded-xl bg-muted/30 border border-border text-foreground dark:text-primary-foreground text-sm focus:outline-none focus:border-orange-400"
              required
            >
              {providers.length === 0 && <option value="">등록된 provider 가 없습니다</option>}
              {providers.map((p) => (
                <option key={p.provider_id} value={p.provider_id}>
                  {p.display_name} ({p.provider_key} · {p.provider_type})
                </option>
              ))}
            </select>
          </div>

          <div>
            <label htmlFor="external_key" className="block text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2">
              External Key *
            </label>
            <input
              id="external_key"
              type="text"
              value={externalKey}
              onChange={(e) => setExternalKey(e.target.value)}
              placeholder="JIRA: PROJ / Gitea: org/repo"
              className="w-full px-4 py-3 rounded-xl bg-muted/30 border border-border text-foreground dark:text-primary-foreground font-mono text-sm focus:outline-none focus:border-orange-400"
              required
            />
            <p className="text-[10px] text-muted-foreground mt-1.5">
              provider 측 식별자 (Jira project key, Gitea repo path 등).
            </p>
          </div>

          <div>
            <label htmlFor="policy" className="block text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2">
              Policy *
            </label>
            <select
              id="policy"
              value={policy}
              onChange={(e) => setPolicy(e.target.value as IntegrationPolicy)}
              className="w-full px-4 py-3 rounded-xl bg-muted/30 border border-border text-foreground dark:text-primary-foreground text-sm focus:outline-none focus:border-orange-400"
            >
              {policyOptions.map((p) => (
                <option key={p} value={p}>{p}</option>
              ))}
            </select>
            <p className="text-[10px] text-muted-foreground mt-1.5">
              <code>execution_system</code>: 실행/명령 연동. <code>summary_only</code>: 요약/상태만 수집.
            </p>
          </div>

          {error && (
            <div className="p-3 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-xs font-bold">
              {error}
            </div>
          )}

          <div className="flex gap-3 pt-4 border-t border-border/60">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-6 py-3 rounded-2xl border border-border text-foreground dark:text-primary-foreground font-black uppercase tracking-widest text-[10px] hover:bg-muted/30 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting || providers.length === 0}
              className="flex-1 px-6 py-3 rounded-2xl bg-orange-500 text-white font-black uppercase tracking-widest text-[10px] shadow-xl shadow-orange-500/20 hover:scale-[1.02] active:scale-[0.98] transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {submitting ? "Creating..." : "Create"}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  );
}
