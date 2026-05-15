"use client";

import { useState, useEffect } from "react";
import { motion } from "framer-motion";
import { X, GitBranch, Loader2, Link } from "lucide-react";
import { ApplicationRepository, ApplicationRepositoryRole, SCMProvider } from "@/lib/services/project.types";
import { projectService } from "@/lib/services/project.service";
import { cn } from "@/lib/utils";

interface RepositoryLinkModalProps {
  applicationId: string;
  onClose: () => void;
  onLinked: (repo: ApplicationRepository) => void;
}

export function RepositoryLinkModal({ applicationId, onClose, onLinked }: RepositoryLinkModalProps) {
  const [formData, setFormData] = useState({
    repo_provider: "",
    repo_full_name: "",
    role: "sub" as ApplicationRepositoryRole,
  });
  const [providers, setProviders] = useState<SCMProvider[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const [loadingProviders, setLoadingProviders] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadProviders = async () => {
      try {
        const data = await projectService.getSCMProviders();
        setProviders(data);
        if (data.length > 0) {
          setFormData(prev => ({ ...prev, repo_provider: data[0].provider_key }));
        }
      } catch (err) {
        console.error("Failed to load SCM providers", err);
      } finally {
        setLoadingProviders(false);
      }
    };
    loadProviders();
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSubmitting(true);

    try {
      const result = await projectService.connectRepository(applicationId, formData);
      onLinked(result);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to link repository");
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
        className="relative w-full max-w-lg glass border-border rounded-3xl shadow-2xl overflow-hidden"
      >
        <div className="p-8 border-b border-border flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-pink-500/20 rounded-xl flex items-center justify-center">
              <Link className="w-5 h-5 text-pink-400" />
            </div>
            <div>
              <h2 className="text-xl font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">
                Link <span className="text-pink-400">Repository</span>
              </h2>
              <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
                Connect external code source to this app
              </p>
            </div>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-muted/30 rounded-xl text-muted-foreground transition-colors">
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-8 space-y-6">
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">SCM Provider</label>
              {loadingProviders ? (
                <div className="h-12 bg-muted/20 animate-pulse rounded-2xl" />
              ) : (
                <select
                  value={formData.repo_provider}
                  onChange={e => setFormData({ ...formData, repo_provider: e.target.value })}
                  className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-pink-400/50 appearance-none"
                >
                  {providers.map(p => (
                    <option key={p.provider_key} value={p.provider_key} className="bg-slate-900">
                      {p.display_name}
                    </option>
                  ))}
                  {providers.length === 0 && <option value="github" className="bg-slate-900">GitHub (Default)</option>}
                </select>
              )}
            </div>

            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Repository Full Name</label>
              <div className="relative group">
                <GitBranch className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-primary-foreground/20 group-focus-within:text-pink-400 transition-colors" />
                <input
                  required
                  value={formData.repo_full_name}
                  onChange={e => setFormData({ ...formData, repo_full_name: e.target.value })}
                  placeholder="e.g. devhub/backend-core"
                  className="w-full bg-muted/30 border border-border rounded-2xl pl-12 pr-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-pink-400/50"
                />
              </div>
              <p className="text-[9px] text-muted-foreground px-1 italic">Format: org/repo or user/repo</p>
            </div>

            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Role in Application</label>
              <div className="grid grid-cols-3 gap-2">
                {(['primary', 'sub', 'shared'] as ApplicationRepositoryRole[]).map((r) => (
                  <button
                    key={r}
                    type="button"
                    onClick={() => setFormData({ ...formData, role: r })}
                    className={cn(
                      "py-2.5 rounded-xl border text-[10px] font-black uppercase tracking-widest transition-all",
                      formData.role === r
                        ? "bg-pink-500/10 border-pink-500/40 text-pink-400 shadow-lg shadow-pink-500/5"
                        : "bg-muted/20 border-border/60 text-muted-foreground hover:bg-muted/40"
                    )}
                  >
                    {r}
                  </button>
                ))}
              </div>
            </div>
          </div>

          {error && (
            <div className="p-4 bg-orange-500/10 border border-orange-500/20 rounded-2xl text-[11px] text-orange-400 font-medium">
              {error}
            </div>
          )}

          <div className="flex gap-4 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 glass border-border text-foreground dark:text-primary-foreground font-bold py-4 rounded-2xl hover:bg-muted/30 transition-all uppercase tracking-widest text-[10px]"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="flex-1 bg-pink-600 text-white font-black py-4 px-8 rounded-2xl hover:scale-[1.02] active:scale-[0.98] transition-all shadow-xl shadow-pink-500/20 disabled:opacity-50 uppercase tracking-widest text-[10px] flex items-center justify-center gap-2"
            >
              {submitting ? <Loader2 className="w-4 h-4 animate-spin" /> : "Link Repository"}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  );
}
