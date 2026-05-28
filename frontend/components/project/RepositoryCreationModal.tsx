"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { X, FolderGit2, Link2, KeyRound, Loader2 } from "lucide-react";
import { repositoryService, Repository } from "@/lib/services/repository.service";

interface RepositoryCreationModalProps {
  onClose: () => void;
  onCreated: (repository: Repository) => void;
}

export function RepositoryCreationModal({ onClose, onCreated }: RepositoryCreationModalProps) {
  const [formData, setFormData] = useState({
    key: "",
    slug: "",
    provider_key: "",
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onClose]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSubmitting(true);

    try {
      const repository = await repositoryService.createRepositoryDraft({
        key: formData.key.trim(),
        slug: formData.slug.trim(),
        provider_key: formData.provider_key.trim() || undefined,
      });
      onCreated(repository);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create repository draft");
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
                Create <span className="text-cyan-400">Repository</span>
              </h2>
              <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
                Register as draft before publish to SCM
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
              <label htmlFor="repoKey" className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">
                Repository Key
              </label>
              <div className="relative group">
                <KeyRound className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-foreground/35 dark:text-primary-foreground/20 group-focus-within:text-cyan-400 transition-colors" />
                <input
                  id="repoKey"
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
              <label htmlFor="repoProviderKey" className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">
                SCM Provider Key (Optional)
              </label>
              <div className="relative group">
                <Link2 className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-foreground/35 dark:text-primary-foreground/20 group-focus-within:text-cyan-400 transition-colors" />
                <input
                  id="repoProviderKey"
                  value={formData.provider_key}
                  onChange={(e) => setFormData({ ...formData, provider_key: e.target.value })}
                  placeholder="e.g. gitea-main"
                  className="w-full bg-muted/30 border border-border rounded-2xl pl-12 pr-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-cyan-400/50"
                />
              </div>
            </div>
          </div>

          <div className="space-y-2">
            <label htmlFor="repoSlug" className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">
              Repository Slug
            </label>
            <input
              id="repoSlug"
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
                  Creating...
                </span>
              ) : (
                "Create Repository"
              )}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  );
}
