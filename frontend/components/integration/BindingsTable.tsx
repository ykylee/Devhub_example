"use client";

import { Link2 } from "lucide-react";
import { format, parseISO } from "date-fns";
import { motion, AnimatePresence } from "framer-motion";
import { Badge } from "@/components/ui/Badge";
import type { IntegrationBinding, IntegrationProvider } from "@/lib/services/integration.types";

interface BindingsTableProps {
  items: IntegrationBinding[];
  providersByID: Record<string, IntegrationProvider>;
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

function scopeTypeVariant(t: IntegrationBinding["scope_type"]): BadgeVariant {
  return t === "application" ? "primary" : "accent";
}

function policyVariant(p: IntegrationBinding["policy"]): BadgeVariant {
  return p === "execution_system" ? "success" : "secondary";
}

export function BindingsTable({ items, providersByID }: BindingsTableProps) {
  if (items.length === 0) {
    return (
      <div className="glass border-border rounded-3xl py-20 flex flex-col items-center justify-center gap-3">
        <Link2 className="w-12 h-12 text-muted-foreground/30" />
        <p className="text-xs font-bold text-muted-foreground uppercase tracking-widest">
          등록된 binding 이 없습니다
        </p>
        <p className="text-[10px] text-muted-foreground/60">
          상단의 Create Binding 버튼으로 첫 매핑을 추가하세요.
        </p>
      </div>
    );
  }

  return (
    <div className="glass border-border rounded-3xl overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="border-b border-border/60 bg-muted/20">
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Scope</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Scope ID</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Provider</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">External Key</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Policy</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest text-center">Enabled</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Created</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/40">
            <AnimatePresence mode="popLayout">
              {items.map((b) => {
                const provider = providersByID[b.provider_id];
                const providerLabel = provider?.display_name ?? b.provider_id;
                const providerKey = provider?.provider_key;
                return (
                  <motion.tr
                    key={b.binding_id}
                    layout
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    className="group hover:bg-muted/30 transition-colors"
                  >
                    <td className="px-6 py-5">
                      <Badge variant={scopeTypeVariant(b.scope_type)}>{b.scope_type}</Badge>
                    </td>
                    <td className="px-6 py-5">
                      <span className="text-xs font-mono text-foreground dark:text-primary-foreground">{b.scope_id}</span>
                    </td>
                    <td className="px-6 py-5">
                      <div className="text-xs font-black text-foreground dark:text-primary-foreground tracking-tight truncate max-w-[240px]">
                        {providerLabel}
                      </div>
                      {providerKey && (
                        <p className="text-[10px] text-muted-foreground mt-1 truncate max-w-[240px] opacity-60">
                          {providerKey}
                        </p>
                      )}
                    </td>
                    <td className="px-6 py-5">
                      <span className="text-xs font-mono text-muted-foreground">{b.external_key}</span>
                    </td>
                    <td className="px-6 py-5">
                      <Badge variant={policyVariant(b.policy)}>{b.policy}</Badge>
                    </td>
                    <td className="px-6 py-5 text-center">
                      <Badge variant={b.enabled ? "success" : "secondary"}>{b.enabled ? "Yes" : "No"}</Badge>
                    </td>
                    <td className="px-6 py-5">
                      <span className="text-[10px] text-muted-foreground font-mono">{safeFormat(b.created_at)}</span>
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
