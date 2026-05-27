"use client";

import { useEffect, useState } from "react";
import { FolderDown, X, Lock, Globe, Check } from "lucide-react";
import { motion } from "framer-motion";
import { integrationService } from "@/lib/services/integration.service";
import type { IntegrationProvider, ScmRepository } from "@/lib/services/integration.types";

interface ImportRepositoriesModalProps {
  provider: IntegrationProvider;
  onClose: () => void;
  /** import 성공 시 호출 (목록 refresh 용). 개수를 전달. */
  onImported: (count: number) => void;
}

// SCM(provider)의 원격 repository 를 조회·선택해 시스템으로 import 하는 모달 (API-88/89).
export function ImportRepositoriesModal({ provider, onClose, onImported }: ImportRepositoriesModalProps) {
  const [repos, setRepos] = useState<ScmRepository[]>([]);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(true);
  const [importing, setImporting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // loading 은 useState(true) 로 시작 — effect 본문에서 동기 setState 하지 않고
    // 모든 상태 전이를 async 콜백(then/catch/finally)에서 처리한다 (set-state-in-effect).
    let active = true;
    integrationService
      .listScmRepositories(provider.provider_id)
      .then((list) => {
        if (active) {
          setRepos(list);
          setError(null);
        }
      })
      .catch((err) => {
        if (active) setError(err instanceof Error ? err.message : "원격 저장소 조회에 실패했습니다.");
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [provider.provider_id]);

  const toggle = (fullName: string) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(fullName)) next.delete(fullName);
      else next.add(fullName);
      return next;
    });
  };

  const importable = repos.filter((r) => !r.imported);
  const allSelected = importable.length > 0 && importable.every((r) => selected.has(r.full_name));
  const toggleAll = () => {
    if (allSelected) {
      setSelected(new Set());
    } else {
      setSelected(new Set(importable.map((r) => r.full_name)));
    }
  };

  const handleImport = async () => {
    if (selected.size === 0) return;
    setImporting(true);
    setError(null);
    try {
      const result = await integrationService.importScmRepositories(provider.provider_id, Array.from(selected));
      onImported(result.imported);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "import 에 실패했습니다.");
    } finally {
      setImporting(false);
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
        className="glass border-border rounded-3xl w-full max-w-2xl max-h-[90vh] overflow-y-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between p-6 border-b border-border/60">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-accent/10 flex items-center justify-center border border-accent/20">
              <FolderDown className="w-5 h-5 text-accent" />
            </div>
            <div>
              <h3 className="text-lg font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">
                Import Repositories
              </h3>
              <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-widest mt-0.5">
                {provider.display_name} ({provider.provider_key})
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

        <div className="p-6 space-y-4">
          {loading ? (
            <p className="text-sm text-muted-foreground py-8 text-center">원격 저장소를 불러오는 중…</p>
          ) : error ? (
            <div className="p-3 rounded-xl bg-destructive/10 border border-destructive/30 text-destructive text-xs font-bold">
              {error}
            </div>
          ) : repos.length === 0 ? (
            <p className="text-sm text-muted-foreground py-8 text-center">연동 가능한 원격 저장소가 없습니다.</p>
          ) : (
            <>
              <div className="flex items-center justify-between">
                <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-widest">
                  {repos.length} repositories · {importable.length} importable
                </p>
                {importable.length > 0 && (
                  <button
                    type="button"
                    onClick={toggleAll}
                    className="text-[10px] font-black text-accent uppercase tracking-widest hover:underline"
                  >
                    {allSelected ? "Clear" : "Select all"}
                  </button>
                )}
              </div>
              <ul className="space-y-2">
                {repos.map((r) => {
                  const isSelected = selected.has(r.full_name);
                  return (
                    <li key={r.full_name}>
                      <label
                        className={`flex items-center gap-3 px-4 py-3 rounded-xl border cursor-pointer transition-colors ${
                          r.imported
                            ? "border-border/40 bg-muted/10 cursor-default"
                            : isSelected
                              ? "border-accent bg-accent/10"
                              : "border-border bg-muted/20 hover:bg-muted/30"
                        }`}
                      >
                        {r.imported ? (
                          <Check className="w-4 h-4 text-emerald-500 shrink-0" />
                        ) : (
                          <input
                            type="checkbox"
                            checked={isSelected}
                            onChange={() => toggle(r.full_name)}
                            className="w-4 h-4 accent-accent shrink-0"
                          />
                        )}
                        <div className="min-w-0 flex-1">
                          <p className="text-sm font-mono text-foreground dark:text-primary-foreground truncate">
                            {r.full_name}
                          </p>
                          <p className="text-[10px] text-muted-foreground">
                            {r.default_branch || "—"}
                          </p>
                        </div>
                        {r.private ? (
                          <Lock className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
                        ) : (
                          <Globe className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
                        )}
                        {r.imported && (
                          <span className="text-[10px] font-black text-emerald-500 uppercase tracking-widest shrink-0">
                            Imported
                          </span>
                        )}
                      </label>
                    </li>
                  );
                })}
              </ul>
            </>
          )}
        </div>

        <div className="flex gap-3 p-6 border-t border-border/60">
          <button
            type="button"
            onClick={onClose}
            className="flex-1 px-6 py-3 rounded-2xl border border-border text-foreground dark:text-primary-foreground font-black uppercase tracking-widest text-[10px] hover:bg-muted/30 transition-colors"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={() => void handleImport()}
            disabled={importing || selected.size === 0}
            className="flex-1 px-6 py-3 rounded-2xl bg-primary text-primary-foreground font-black uppercase tracking-widest text-[10px] shadow-xl hover:scale-[1.02] active:scale-[0.98] transition-all disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {importing ? "Importing…" : `Import${selected.size > 0 ? ` (${selected.size})` : ""}`}
          </button>
        </div>
      </motion.div>
    </div>
  );
}
