"use client";

import { Key, Globe, ShieldOff, Clock } from "lucide-react";
import { format, parseISO } from "date-fns";
import { motion, AnimatePresence } from "framer-motion";
import { Badge } from "@/components/ui/Badge";
import type { DevRequestIntakeToken } from "@/lib/services/dev_request_token.types";

interface IntakeTokenTableProps {
  items: DevRequestIntakeToken[];
  onRevoke: (token: DevRequestIntakeToken) => void;
  revokingTokenID: string | null;
}

function safeFormat(iso: string | null | undefined): string {
  if (!iso) return "—";
  try {
    return format(parseISO(iso), "yyyy-MM-dd HH:mm");
  } catch {
    return iso;
  }
}

export function IntakeTokenTable({ items, onRevoke, revokingTokenID }: IntakeTokenTableProps) {
  if (items.length === 0) {
    return (
      <div className="glass border-border rounded-3xl py-20 flex flex-col items-center justify-center gap-3">
        <Key className="w-12 h-12 text-muted-foreground/30" />
        <p className="text-xs font-bold text-muted-foreground uppercase tracking-widest">
          발급된 intake token 이 없습니다
        </p>
        <p className="text-[10px] text-muted-foreground/60">상단의 Issue Token 버튼으로 첫 토큰을 발급하세요.</p>
      </div>
    );
  }

  return (
    <div className="glass border-border rounded-3xl overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="border-b border-border/60 bg-muted/20">
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Client / Source</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Allowed IPs</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest text-center">Status</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Expires</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Created</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Last Used</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/40">
            <AnimatePresence mode="popLayout">
              {items.map((tok) => {
                const isRevoked = Boolean(tok.revoked_at);
                const isExpired = tok.expires_at && new Date(tok.expires_at) < new Date();
                const isBusy = revokingTokenID === tok.token_id;
                return (
                  <motion.tr
                    key={tok.token_id}
                    layout
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    className="group hover:bg-muted/30 transition-colors"
                  >
                    <td className="px-6 py-5">
                      <div className="flex items-center gap-4">
                        <div className="w-10 h-10 rounded-xl bg-orange-500/10 flex items-center justify-center border border-orange-500/20">
                          <Key className="w-5 h-5 text-orange-400" />
                        </div>
                        <div className="min-w-0">
                          <div className="text-xs font-black text-foreground dark:text-primary-foreground tracking-tight truncate max-w-[280px]">
                            {tok.client_label}
                          </div>
                          <p className="text-[10px] text-muted-foreground mt-1 truncate max-w-[280px] opacity-60">
                            source: {tok.source_system}
                          </p>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-5">
                      <div className="flex flex-wrap gap-1.5 max-w-[280px]">
                        {tok.allowed_ips.map((ip) => (
                          <span
                            key={ip}
                            className="px-2 py-1 rounded-md bg-muted/40 border border-border/60 text-[10px] font-mono text-muted-foreground flex items-center gap-1.5"
                          >
                            <Globe className="w-3 h-3 opacity-50" />
                            {ip}
                          </span>
                        ))}
                      </div>
                    </td>
                    <td className="px-6 py-5 text-center">
                      {isRevoked ? (
                        <Badge variant="glass" className="bg-red-500/10 text-red-400 border-red-500/30">
                          Revoked
                        </Badge>
                      ) : isExpired ? (
                        <Badge variant="glass" className="bg-amber-500/10 text-amber-400 border-amber-500/30">
                          Expired
                        </Badge>
                      ) : (
                        <Badge variant="glass" className="bg-emerald-500/10 text-emerald-400 border-emerald-500/30">
                          Active
                        </Badge>
                      )}
                    </td>
                    <td className="px-6 py-5">
                      <div className={`text-[11px] flex items-center gap-1.5 ${isExpired ? "text-amber-400 font-bold" : "text-muted-foreground"}`}>
                        {tok.expires_at ? safeFormat(tok.expires_at) : "무기한"}
                      </div>
                    </td>
                    <td className="px-6 py-5">
                      <div className="text-[11px] text-muted-foreground flex items-center gap-1.5">
                        <Clock className="w-3 h-3 opacity-50" />
                        {safeFormat(tok.created_at)}
                      </div>
                      <p className="text-[10px] text-muted-foreground/60 mt-0.5">by {tok.created_by}</p>
                    </td>
                    <td className="px-6 py-5 text-[11px] text-muted-foreground">{safeFormat(tok.last_used_at)}</td>
                    <td className="px-6 py-5 text-right">
                      {!isRevoked && (
                        <button
                          type="button"
                          onClick={() => onRevoke(tok)}
                          disabled={isBusy}
                          className="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 hover:bg-red-500/20 disabled:opacity-40 transition-all text-[10px] font-black uppercase tracking-widest"
                        >
                          <ShieldOff className="w-3.5 h-3.5" />
                          {isBusy ? "Revoking…" : "Revoke"}
                        </button>
                      )}
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
