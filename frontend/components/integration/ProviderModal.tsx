"use client";

import { useState, FormEvent } from "react";
import { Plug, X } from "lucide-react";
import { motion } from "framer-motion";
import { integrationService } from "@/lib/services/integration.service";
import type {
  IntegrationProvider,
  IntegrationProviderType,
  IntegrationAuthMode,
} from "@/lib/services/integration.types";

interface ProviderModalProps {
  /** edit 모드 시 기존 provider, create 모드 시 undefined */
  initial?: IntegrationProvider;
  onClose: () => void;
  onSaved: (provider: IntegrationProvider) => void;
}

const providerTypeOptions: IntegrationProviderType[] = ["alm", "scm", "ci_cd", "doc", "infra"];
const authModeOptions: IntegrationAuthMode[] = ["token", "basic", "oauth2", "app_password", "agent"];

export function ProviderModal({ initial, onClose, onSaved }: ProviderModalProps) {
  const isEdit = Boolean(initial);

  const [providerKey, setProviderKey] = useState(initial?.provider_key ?? "");
  const [providerType, setProviderType] = useState<IntegrationProviderType>(initial?.provider_type ?? "scm");
  const [displayName, setDisplayName] = useState(initial?.display_name ?? "");
  const [authMode, setAuthMode] = useState<IntegrationAuthMode>(initial?.auth_mode ?? "token");
  const [credentialsRef, setCredentialsRef] = useState(initial?.credentials_ref ?? "");
  const [enabled, setEnabled] = useState(initial?.enabled ?? true);
  const [capabilitiesRaw, setCapabilitiesRaw] = useState((initial?.capabilities ?? []).join(", "));

  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const parseCapabilities = (): string[] =>
    capabilitiesRaw
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      let saved: IntegrationProvider;
      if (isEdit && initial) {
        saved = await integrationService.updateProvider(initial.provider_id, {
          enabled,
          display_name: displayName.trim() || undefined,
          credentials_ref: credentialsRef.trim() || undefined,
          capabilities: parseCapabilities(),
        });
      } else {
        if (!providerKey.trim()) {
          setError("provider_key 는 필수입니다.");
          return;
        }
        if (!displayName.trim()) {
          setError("display_name 은 필수입니다.");
          return;
        }
        if (!credentialsRef.trim()) {
          setError("credentials_ref 는 필수입니다.");
          return;
        }
        saved = await integrationService.createProvider({
          provider_key: providerKey.trim(),
          provider_type: providerType,
          display_name: displayName.trim(),
          auth_mode: authMode,
          credentials_ref: credentialsRef.trim(),
          capabilities: parseCapabilities(),
        });
      }
      onSaved(saved);
      onClose();
    } catch (err) {
      const msg = err instanceof Error ? err.message : "저장에 실패했습니다.";
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
              <Plug className="w-5 h-5 text-orange-400" />
            </div>
            <div>
              <h3 className="text-lg font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">
                {isEdit ? "Edit Provider" : "Register Provider"}
              </h3>
              <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-widest mt-0.5">
                {isEdit ? `provider_key: ${initial?.provider_key}` : "신규 외부 시스템 연동 등록"}
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
          {!isEdit && (
            <div>
              <label htmlFor="provider_key" className="block text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2">
                Provider Key *
              </label>
              <input
                id="provider_key"
                type="text"
                value={providerKey}
                onChange={(e) => setProviderKey(e.target.value)}
                placeholder="gitea_main / jenkins_prod"
                className="w-full px-4 py-3 rounded-xl bg-muted/30 border border-border text-foreground dark:text-primary-foreground font-mono text-sm focus:outline-none focus:border-orange-400"
                required
              />
              <p className="text-[10px] text-muted-foreground mt-1.5">
                URL-safe 식별자. 발급 후 변경 불가.
              </p>
            </div>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label htmlFor="provider_type" className="block text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2">
                Type *
              </label>
              <select
                id="provider_type"
                value={providerType}
                onChange={(e) => setProviderType(e.target.value as IntegrationProviderType)}
                disabled={isEdit}
                className="w-full px-4 py-3 rounded-xl bg-muted/30 border border-border text-foreground dark:text-primary-foreground text-sm focus:outline-none focus:border-orange-400 disabled:opacity-60"
              >
                {providerTypeOptions.map((t) => (
                  <option key={t} value={t}>{t}</option>
                ))}
              </select>
            </div>
            <div>
              <label htmlFor="auth_mode" className="block text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2">
                Auth Mode *
              </label>
              <select
                id="auth_mode"
                value={authMode}
                onChange={(e) => setAuthMode(e.target.value as IntegrationAuthMode)}
                disabled={isEdit}
                className="w-full px-4 py-3 rounded-xl bg-muted/30 border border-border text-foreground dark:text-primary-foreground text-sm focus:outline-none focus:border-orange-400 disabled:opacity-60"
              >
                {authModeOptions.map((m) => (
                  <option key={m} value={m}>{m}</option>
                ))}
              </select>
            </div>
          </div>

          <div>
            <label htmlFor="display_name" className="block text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2">
              Display Name *
            </label>
            <input
              id="display_name"
              type="text"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="Gitea Main / Jenkins Production"
              className="w-full px-4 py-3 rounded-xl bg-muted/30 border border-border text-foreground dark:text-primary-foreground text-sm focus:outline-none focus:border-orange-400"
            />
          </div>

          <div>
            <label htmlFor="credentials_ref" className="block text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2">
              Credentials Ref *
            </label>
            <input
              id="credentials_ref"
              type="text"
              value={credentialsRef}
              onChange={(e) => setCredentialsRef(e.target.value)}
              placeholder="hmac_sha256:<secret> 또는 provider_sdk:<vendor>:<secret>"
              className="w-full px-4 py-3 rounded-xl bg-muted/30 border border-border text-foreground dark:text-primary-foreground text-sm font-mono focus:outline-none focus:border-orange-400"
            />
            <p className="text-[10px] text-muted-foreground mt-1.5">
              webhook signature 검증 전략. 예: <code>hmac_sha256:abc...</code>, <code>provider_sdk:github:def...</code>
            </p>
          </div>

          <div>
            <label htmlFor="capabilities" className="block text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2">
              Capabilities (comma-separated)
            </label>
            <input
              id="capabilities"
              type="text"
              value={capabilitiesRaw}
              onChange={(e) => setCapabilitiesRaw(e.target.value)}
              placeholder="webhook, pull, snapshot"
              className="w-full px-4 py-3 rounded-xl bg-muted/30 border border-border text-foreground dark:text-primary-foreground text-sm font-mono focus:outline-none focus:border-orange-400"
            />
          </div>

          {isEdit && (
            <div className="flex items-center gap-3 p-4 rounded-xl bg-muted/20 border border-border/60">
              <input
                id="enabled"
                type="checkbox"
                checked={enabled}
                onChange={(e) => setEnabled(e.target.checked)}
                className="w-4 h-4 accent-orange-400"
              />
              <label htmlFor="enabled" className="text-xs font-bold text-foreground dark:text-primary-foreground cursor-pointer">
                Enabled (수신/sync 활성)
              </label>
            </div>
          )}

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
              disabled={submitting}
              className="flex-1 px-6 py-3 rounded-2xl bg-orange-500 text-white font-black uppercase tracking-widest text-[10px] shadow-xl shadow-orange-500/20 hover:scale-[1.02] active:scale-[0.98] transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {submitting ? "Saving..." : isEdit ? "Save" : "Register"}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  );
}
