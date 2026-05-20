"use client";

import { useState, FormEvent, useEffect, useMemo } from "react";
import { Link2, X, AlertTriangle } from "lucide-react";
import { motion } from "framer-motion";
import { cn } from "@/lib/utils";
import { integrationService } from "@/lib/services/integration.service";
import { ApiError } from "@/lib/services/api-client";
import type {
  IntegrationBinding,
  IntegrationPolicy,
  IntegrationProvider,
} from "@/lib/services/integration.types";

interface EditBindingModalProps {
  binding: IntegrationBinding;
  providers: IntegrationProvider[];
  onClose: () => void;
  onUpdated: (binding: IntegrationBinding) => void;
}

const policyOptions: IntegrationPolicy[] = ["summary_only", "execution_system"];

export function EditBindingModal({ binding, providers, onClose, onUpdated }: EditBindingModalProps) {
  const [providerID, setProviderID] = useState(binding.provider_id);
  const [externalKey, setExternalKey] = useState(binding.external_key);
  const [policy, setPolicy] = useState<IntegrationPolicy>(binding.policy);
  const [enabled, setEnabled] = useState(binding.enabled);

  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
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
      const updated = await integrationService.updateBinding(binding.binding_id, {
        provider_id: providerID,
        external_key: externalKey.trim(),
        policy,
        enabled,
      });
      onUpdated(updated);
      onClose();
    } catch (err) {
      if (err instanceof ApiError) {
        const payload = err.payload as { code?: string; error?: string } | null;
        if (err.status === 409) {
          setError("이미 동일한 (scope, provider, external_key) binding 이 존재합니다.");
          return;
        }
      }
      const msg = err instanceof Error ? err.message : "binding 수정에 실패했습니다.";
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
            <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center border border-primary/20">
              <Link2 className="w-5 h-5 text-primary" />
            </div>
            <div>
              <h3 className="text-lg font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">
                Edit Binding
              </h3>
              <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-widest mt-0.5">
                {binding.scope_type}: {binding.scope_id} 매핑 수정
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
          <div className="p-4 bg-muted/20 border border-border rounded-2xl">
            <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">Target Scope (Read-only)</p>
            <div className="flex items-center gap-2">
              <span className="text-xs font-bold px-2 py-0.5 bg-primary/10 text-primary rounded uppercase">{binding.scope_type}</span>
              <span className="text-sm font-mono font-bold text-foreground">{binding.scope_id}</span>
            </div>
            <p className="text-[10px] text-muted-foreground mt-2 italic">Scope ID 는 변경할 수 없습니다. 대상을 바꾸려면 삭제 후 새로 생성하세요.</p>
          </div>

          <div>
            <label htmlFor="provider_id" className="block text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2">
              Provider *
            </label>
            <select
              id="provider_id"
              value={providerID}
              onChange={(e) => setProviderID(e.target.value)}
              className="w-full px-4 py-3 rounded-xl bg-muted/30 border border-border text-foreground dark:text-primary-foreground text-sm focus:outline-none focus:border-primary"
              required
            >
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
              className="w-full px-4 py-3 rounded-xl bg-muted/30 border border-border text-foreground dark:text-primary-foreground font-mono text-sm focus:outline-none focus:border-primary"
              required
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label htmlFor="policy" className="block text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2">
                Policy *
              </label>
              <select
                id="policy"
                value={policy}
                onChange={(e) => setPolicy(e.target.value as IntegrationPolicy)}
                className="w-full px-4 py-3 rounded-xl bg-muted/30 border border-border text-foreground dark:text-primary-foreground text-sm focus:outline-none focus:border-primary"
              >
                {policyOptions.map((p) => (
                  <option key={p} value={p}>{p}</option>
                ))}
              </select>
            </div>
            <div className="flex flex-col">
              <label className="block text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2">
                Status
              </label>
              <button
                type="button"
                onClick={() => setEnabled(!enabled)}
                className={cn(
                  "flex-1 flex items-center justify-center gap-2 px-4 py-2.5 rounded-xl border transition-all text-xs font-bold uppercase tracking-widest",
                  enabled 
                    ? "bg-success/10 border-success/30 text-success" 
                    : "bg-muted border-border text-muted-foreground"
                )}
              >
                {enabled ? "Enabled" : "Disabled"}
              </button>
            </div>
          </div>

          {error && (
            <div className="p-3 rounded-xl bg-destructive/10 border border-destructive/30 text-destructive text-xs font-bold flex items-center gap-2">
              <AlertTriangle className="w-4 h-4" />
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
              disabled={submitting}
              className="flex-1 px-6 py-3 rounded-2xl bg-primary text-primary-foreground font-black uppercase tracking-widest text-[10px] shadow-xl hover:scale-[1.02] active:scale-[0.98] transition-all disabled:opacity-50"
            >
              {submitting ? "Saving..." : "Save Changes"}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  );
}
