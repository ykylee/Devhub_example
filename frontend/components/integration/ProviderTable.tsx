"use client";

import { Plug, Settings, RefreshCw, Trash2 } from "lucide-react";
import { format, parseISO } from "date-fns";
import { motion, AnimatePresence } from "framer-motion";
import { Badge } from "@/components/ui/Badge";
import type { IntegrationProvider } from "@/lib/services/integration.types";

interface ProviderTableProps {
  items: IntegrationProvider[];
  onEdit: (provider: IntegrationProvider) => void;
  onSync: (provider: IntegrationProvider) => void;
  onDelete: (provider: IntegrationProvider) => void;
  syncingProviderID: string | null;
  deletingProviderID: string | null;
}

function safeFormat(iso: string | null | undefined): string {
  if (!iso) return "—";
  try {
    return format(parseISO(iso), "yyyy-MM-dd HH:mm");
  } catch {
    return iso;
  }
}

type BadgeVariant = "primary" | "secondary" | "accent" | "success" | "warning" | "danger" | "glass";

function providerTypeVariant(t: IntegrationProvider["provider_type"]): BadgeVariant {
  switch (t) {
    case "scm": return "primary";
    case "ci_cd": return "success";
    case "alm": return "accent";
    case "doc": return "warning";
    case "infra": return "accent";
  }
}

function syncStatusBadge(s: string): { variant: BadgeVariant; label: string } {
  const norm = s.toLowerCase();
  if (norm === "ok" || norm === "success") return { variant: "success", label: "OK" };
  if (norm === "requested" || norm === "pending") return { variant: "primary", label: "Pending" };
  if (norm === "degraded" || norm === "warning") return { variant: "warning", label: "Degraded" };
  if (norm === "error" || norm === "failed") return { variant: "danger", label: "Error" };
  return { variant: "secondary", label: s || "—" };
}

export function ProviderTable({ items, onEdit, onSync, onDelete, syncingProviderID, deletingProviderID }: ProviderTableProps) {
  if (items.length === 0) {
    return (
      <div className="glass border-border rounded-3xl py-20 flex flex-col items-center justify-center gap-3">
        <Plug className="w-12 h-12 text-muted-foreground/30" />
        <p className="text-xs font-bold text-muted-foreground uppercase tracking-widest">
          등록된 integration provider 가 없습니다
        </p>
        <p className="text-[10px] text-muted-foreground/60">상단의 Register Provider 버튼으로 첫 provider 를 추가하세요.</p>
      </div>
    );
  }

  return (
    <div className="glass border-border rounded-3xl overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="border-b border-border/60 bg-muted/20">
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Provider</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Type</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Auth</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest text-center">Enabled</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest text-center">Sync Status</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Last Sync</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/40">
            <AnimatePresence mode="popLayout">
              {items.map((p) => {
                const isSyncing = syncingProviderID === p.provider_id;
                const isDeleting = deletingProviderID === p.provider_id;
                const sync = syncStatusBadge(p.sync_status);
                return (
                  <motion.tr
                    key={p.provider_id}
                    layout
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    className="group hover:bg-muted/30 transition-colors"
                  >
                    <td className="px-6 py-5">
                      <div className="flex items-center gap-4">
                        <div className="w-10 h-10 rounded-xl bg-orange-500/10 flex items-center justify-center border border-orange-500/20">
                          <Plug className="w-5 h-5 text-orange-400" />
                        </div>
                        <div className="min-w-0">
                          <div className="text-xs font-black text-foreground dark:text-primary-foreground tracking-tight truncate max-w-[280px]">
                            {p.display_name}
                          </div>
                          <p className="text-[10px] text-muted-foreground mt-1 truncate max-w-[280px] opacity-60">
                            {p.provider_key}
                          </p>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-5">
                      <Badge variant={providerTypeVariant(p.provider_type)}>{p.provider_type}</Badge>
                    </td>
                    <td className="px-6 py-5">
                      <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">{p.auth_mode}</span>
                    </td>
                    <td className="px-6 py-5 text-center">
                      <Badge variant={p.enabled ? "success" : "secondary"}>{p.enabled ? "Yes" : "No"}</Badge>
                    </td>
                    <td className="px-6 py-5 text-center">
                      <Badge variant={sync.variant}>{sync.label}</Badge>
                      {p.last_error_code && (
                        <p className="text-[9px] text-red-400 mt-1 font-mono">{p.last_error_code}</p>
                      )}
                    </td>
                    <td className="px-6 py-5">
                      <span className="text-[10px] text-muted-foreground font-mono">{safeFormat(p.last_sync_at)}</span>
                    </td>
                    <td className="px-6 py-5 text-right">
                      <div className="inline-flex gap-2">
                        <button
                          type="button"
                          onClick={() => onSync(p)}
                          disabled={isSyncing}
                          className="inline-flex items-center gap-1 px-3 py-1.5 rounded-lg bg-muted/50 hover:bg-muted text-[10px] font-bold uppercase tracking-widest text-foreground dark:text-primary-foreground transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                          aria-label={`Sync ${p.display_name}`}
                        >
                          <RefreshCw className={"w-3 h-3 " + (isSyncing ? "animate-spin" : "")} />
                          {isSyncing ? "Syncing" : "Sync"}
                        </button>
                        <button
                          type="button"
                          onClick={() => onEdit(p)}
                          className="inline-flex items-center gap-1 px-3 py-1.5 rounded-lg bg-muted/50 hover:bg-muted text-[10px] font-bold uppercase tracking-widest text-foreground dark:text-primary-foreground transition-colors"
                          aria-label={`Edit ${p.display_name}`}
                        >
                          <Settings className="w-3 h-3" />
                          Edit
                        </button>
                        <button
                          type="button"
                          onClick={() => onDelete(p)}
                          disabled={isDeleting}
                          className="inline-flex items-center gap-1 px-3 py-1.5 rounded-lg bg-red-500/10 hover:bg-red-500/20 text-[10px] font-bold uppercase tracking-widest text-red-400 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                          aria-label={`Delete ${p.display_name}`}
                        >
                          <Trash2 className="w-3 h-3" />
                          {isDeleting ? "Deleting" : "Delete"}
                        </button>
                      </div>
                    </td>
                  </motion.tr>
                );
              })}
            </AnimatePresence>
          </tbody>
        </table>
      </div>
    </div>
  );
}
