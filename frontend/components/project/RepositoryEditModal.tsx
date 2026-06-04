"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { X, FolderGit2, Link2, KeyRound, Loader2 } from "lucide-react";
import { repositoryService, Repository } from "@/domain/repository-integration/service/repository.service";
import { integrationService } from "@/domain/integration-registry/service/integration.service";
import { IntegrationProvider } from "@/domain/integration-registry/schema/integration.types";

interface RepositoryEditModalProps {
  repository: Repository;
  onClose: () => void;
  onUpdated: (repository: Repository) => void;
}

export function RepositoryEditModal({ repository, onClose, onUpdated }: RepositoryEditModalProps) {
  const [formData, setFormData] = useState({
    key: repository.name ?? "",
    slug: repository.full_name ?? "",
    provider_key: repository.provider_key ?? "",
  });
  const [scmProviders, setScmProviders] = useState<IntegrationProvider[]>([]);
  const [loadingProviders, setLoadingProviders] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onClose]);

  useEffect(() => {
    let cancelled = false;
    // backend `updateRepositoryDraft` 가 `GetIntegrationProviderByKey` 로
    // `integration_providers` 에서 resolve 하므로 dropdown source 도 일치 (PR #470).
    integrationService
      .listProviders({ provider_type: "scm", enabled: true })
      .then(({ data }) => {
        if (cancelled) return;
        setScmProviders(data);
      })
      .catch((err) => {
        if (cancelled) return;
        console.error("Failed to load SCM providers", err);
        setScmProviders([]);
      })
      .finally(() => {
        if (cancelled) return;
        setLoadingProviders(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSubmitting(true);

    try {
      // provider_key 변경 여부 추적 — 변경 없으면 PATCH body 에서 제외 (nil=unchanged).
      const originalProviderKey = repository.provider_key ?? "";
      const providerChanged = formData.provider_key !== originalProviderKey;
      const updated = await repositoryService.updateRepository(repository.id, {
        key: formData.key.trim() !== repository.name ? formData.key.trim() : undefined,
        slug: formData.slug.trim() !== repository.full_name ? formData.slug.trim() : undefined,
        provider_key: providerChanged ? formData.provider_key : undefined,
      });
      onUpdated(updated);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update repository draft");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-6">
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        onClick={onClose}
        className="absolute inset-0 bg-background/80 backdrop-blur-sm"
      />

      <motion.div
        initial={{ opacity: 0, scale: 0.95, y: 20 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.95, y: 20 }}
        role="dialog"
        aria-modal="true"
        className="relative w-full max-w-2xl glass border-border rounded-3xl shadow-2xl overflow-hidden"
      >
        <div className="p-8 border-b border-border flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-cyan-500/20 rounded-xl flex items-center justify-center">
              <FolderGit2 className="w-5 h-5 text-cyan-400" />
            </div>
            <div>
              <h2 className="text-xl font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">
                Edit <span className="text-cyan-400">Repository Draft</span>
              </h2>
              <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
                Draft 상태만 수정 가능 (publish 후 SCM 이 source of truth)
              </p>
            </div>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-muted/30 rounded-xl text-muted-foreground transition-colors">
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-8 space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <label htmlFor="editRepoKey" className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">
                Repository Key
              </label>
              <div className="relative group">
                <KeyRound className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-foreground/35 dark:text-primary-foreground/20 group-focus-within:text-cyan-400 transition-colors" />
                <input
                  id="editRepoKey"
                  required
                  value={formData.key}
                  onChange={(e) => setFormData({ ...formData, key: e.target.value.toUpperCase() })}
                  placeholder="E.G. DEVHUBAPI"
                  maxLength={32}
                  className="w-full bg-muted/30 border border-border rounded-2xl pl-12 pr-4 py-3 text-sm font-mono text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-cyan-400/50 uppercase"
                />
              </div>
            </div>
            <div className="space-y-2">
              <label htmlFor="editRepoProviderKey" className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">
                SCM Provider (Optional)
              </label>
              {loadingProviders ? (
                <div className="h-12 bg-muted/20 animate-pulse rounded-2xl" />
              ) : (
                <div className="relative group">
                  <Link2 className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-foreground/35 dark:text-primary-foreground/20 group-focus-within:text-cyan-400 transition-colors" />
                  <select
                    id="editRepoProviderKey"
                    value={formData.provider_key}
                    onChange={(e) => setFormData({ ...formData, provider_key: e.target.value })}
                    disabled={submitting}
                    className="w-full bg-muted/30 border border-border rounded-2xl pl-12 pr-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-cyan-400/50 appearance-none disabled:opacity-60"
                  >
                    <option value="">No SCM link (unlink)</option>
                    {scmProviders.map((p) => (
                      <option key={p.provider_key} value={p.provider_key} className="bg-slate-900">
                        {p.display_name} ({p.provider_key})
                      </option>
                    ))}
                    {scmProviders.length === 0 && (
                      <option value="" disabled>
                        No enabled SCM providers registered
                      </option>
                    )}
                  </select>
                </div>
              )}
            </div>
          </div>

          <div className="space-y-2">
            <label htmlFor="editRepoSlug" className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">
              Repository Slug
            </label>
            <input
              id="editRepoSlug"
              required
              value={formData.slug}
              onChange={(e) => setFormData({ ...formData, slug: e.target.value })}
              placeholder="e.g. devhub/devhub-api"
              className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm font-mono text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-cyan-400/50"
            />
            <p className="text-[9px] text-muted-foreground px-1 italic">형식 예시: owner/repository-name</p>
          </div>

          {error && <p className="text-xs text-destructive font-semibold">{error}</p>}

          <div className="flex items-center justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              disabled={submitting}
              className="rounded-xl border border-border px-4 py-2 text-xs font-black uppercase tracking-widest text-muted-foreground hover:bg-muted/30"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="rounded-xl bg-cyan-500/90 px-4 py-2 text-xs font-black uppercase tracking-widest text-cyan-950 hover:bg-cyan-400 disabled:opacity-60"
            >
              {submitting ? (
                <span className="inline-flex items-center gap-2">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Saving...
                </span>
              ) : (
                "Save Changes"
              )}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  );
}
