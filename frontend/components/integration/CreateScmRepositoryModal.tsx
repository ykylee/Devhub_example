"use client";

import { useState, FormEvent } from "react";
import { FolderPlus, X } from "lucide-react";
import { motion } from "framer-motion";
import { integrationService } from "@/lib/services/integration.service";
import type { IntegrationProvider } from "@/lib/services/integration.types";

interface CreateScmRepositoryModalProps {
  provider: IntegrationProvider;
  onClose: () => void;
  /** 생성 성공 시 호출 (full_name 전달). */
  onCreated: (fullName: string) => void;
}

const inputCls =
  "w-full px-4 py-3 rounded-xl bg-muted/30 border border-border text-foreground dark:text-primary-foreground text-sm focus:outline-none focus:border-accent";
const labelCls = "block text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2";

// Phase C — 선택 SCM(provider)에 실제 저장소를 생성하는 모달 (API-90).
export function CreateScmRepositoryModal({ provider, onClose, onCreated }: CreateScmRepositoryModalProps) {
  const [name, setName] = useState("");
  const [owner, setOwner] = useState("");
  const [description, setDescription] = useState("");
  const [isPrivate, setIsPrivate] = useState(false);
  const [autoInit, setAutoInit] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      setError("repository 이름은 필수입니다.");
      return;
    }
    setSubmitting(true);
    setError(null);
    try {
      const result = await integrationService.createScmRepository(provider.provider_id, {
        name: name.trim(),
        owner: owner.trim() || undefined,
        description: description.trim() || undefined,
        private: isPrivate,
        auto_init: autoInit,
      });
      onCreated(result.repository.full_name);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "저장소 생성에 실패했습니다.");
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
        className="glass border-border rounded-3xl w-full max-w-lg max-h-[90vh] overflow-y-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between p-6 border-b border-border/60">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-accent/10 flex items-center justify-center border border-accent/20">
              <FolderPlus className="w-5 h-5 text-accent" />
            </div>
            <div>
              <h3 className="text-lg font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">
                Create Repository
              </h3>
              <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-widest mt-0.5">
                {provider.display_name} 에 새 저장소 생성
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
          <div>
            <label htmlFor="repo_name" className={labelCls}>Name *</label>
            <input
              id="repo_name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="my-service"
              className={`${inputCls} font-mono`}
              required
            />
          </div>
          <div>
            <label htmlFor="repo_owner" className={labelCls}>Owner (organization)</label>
            <input
              id="repo_owner"
              type="text"
              value={owner}
              onChange={(e) => setOwner(e.target.value)}
              placeholder="비우면 인증 계정 하위에 생성"
              className={`${inputCls} font-mono`}
            />
          </div>
          <div>
            <label htmlFor="repo_desc" className={labelCls}>Description</label>
            <input
              id="repo_desc"
              type="text"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className={inputCls}
            />
          </div>
          <div className="flex gap-4">
            <label className="flex items-center gap-2 text-xs font-bold text-foreground dark:text-primary-foreground cursor-pointer">
              <input type="checkbox" checked={isPrivate} onChange={(e) => setIsPrivate(e.target.checked)} className="w-4 h-4 accent-accent" />
              Private
            </label>
            <label className="flex items-center gap-2 text-xs font-bold text-foreground dark:text-primary-foreground cursor-pointer">
              <input type="checkbox" checked={autoInit} onChange={(e) => setAutoInit(e.target.checked)} className="w-4 h-4 accent-accent" />
              Initialize (README)
            </label>
          </div>

          {error && (
            <div className="p-3 rounded-xl bg-destructive/10 border border-destructive/30 text-destructive text-xs font-bold">
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
              className="flex-1 px-6 py-3 rounded-2xl bg-primary text-primary-foreground font-black uppercase tracking-widest text-[10px] shadow-xl hover:scale-[1.02] active:scale-[0.98] transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {submitting ? "Creating…" : "Create"}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  );
}
