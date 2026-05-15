"use client";

import { Inbox, Clock, User, ExternalLink } from "lucide-react";
import { format, parseISO } from "date-fns";
import { motion, AnimatePresence } from "framer-motion";
import { DevRequest, DevRequestStatus } from "@/lib/services/dev_request.types";
import { Badge } from "@/components/ui/Badge";

interface DevRequestTableProps {
  items: DevRequest[];
  onSelect: (dr: DevRequest) => void;
}

export function DevRequestTable({ items, onSelect }: DevRequestTableProps) {
  return (
    <div className="glass border-border rounded-3xl overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="border-b border-border/60 bg-muted/20">
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Title & Requester</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest text-center">Status</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Source</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Assignee</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Received</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/40">
            <AnimatePresence mode="popLayout">
              {items.map((dr) => (
                <motion.tr
                  key={dr.id}
                  layout
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  className="group hover:bg-muted/30 transition-colors cursor-pointer"
                  onClick={() => onSelect(dr)}
                >
                  <td className="px-6 py-5">
                    <div className="flex items-center gap-4">
                      <div className="w-10 h-10 rounded-xl bg-orange-500/10 flex items-center justify-center border border-orange-500/20">
                        <Inbox className="w-5 h-5 text-orange-400" />
                      </div>
                      <div className="min-w-0">
                        <div className="text-xs font-black text-foreground dark:text-primary-foreground tracking-tight truncate max-w-[420px]">
                          {dr.title}
                        </div>
                        <p className="text-[10px] text-muted-foreground mt-1 truncate max-w-[420px] opacity-60">
                          requested by {dr.requester}
                          {dr.external_ref ? <> · ref <span className="font-mono">{dr.external_ref}</span></> : null}
                        </p>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-5 text-center">
                    <StatusBadge status={dr.status} />
                  </td>
                  <td className="px-6 py-5 text-[11px] font-bold text-foreground/80">
                    <div className="flex items-center gap-1.5">
                      <ExternalLink className="w-3 h-3 opacity-40" />
                      <span>{dr.source_system}</span>
                    </div>
                  </td>
                  <td className="px-6 py-5">
                    <div className="flex items-center gap-2">
                      <User className="w-3.5 h-3.5 text-muted-foreground/40" />
                      <span className="text-[11px] font-bold text-foreground/80">{dr.assignee_user_id}</span>
                    </div>
                  </td>
                  <td className="px-6 py-5 text-[10px] font-medium text-muted-foreground">
                    <div className="flex items-center gap-1.5">
                      <Clock className="w-3 h-3 opacity-40" />
                      <span>{formatReceivedAt(dr.received_at)}</span>
                    </div>
                  </td>
                </motion.tr>
              ))}
            </AnimatePresence>
          </tbody>
        </table>
      </div>
      {items.length === 0 && (
        <div className="py-24 flex flex-col items-center justify-center gap-4 text-center">
          <div className="w-16 h-16 rounded-3xl bg-muted/20 flex items-center justify-center border border-border/40">
            <Inbox className="w-8 h-8 text-muted-foreground/30" />
          </div>
          <div>
            <p className="text-sm font-black text-foreground/60 uppercase tracking-widest">No Dev Requests</p>
            <p className="text-[10px] text-muted-foreground font-bold mt-1 uppercase tracking-widest opacity-40">
              외부 시스템에서 의뢰가 도착하면 여기 표시됩니다
            </p>
          </div>
        </div>
      )}
    </div>
  );
}

function formatReceivedAt(value: string): string {
  if (!value) return "—";
  try {
    return format(parseISO(value), "MMM d, HH:mm");
  } catch {
    return value;
  }
}

function StatusBadge({ status }: { status: DevRequestStatus }) {
  switch (status) {
    case "pending":
      return <Badge variant="warning" dot>Pending</Badge>;
    case "in_review":
      return <Badge variant="primary" dot>In Review</Badge>;
    case "registered":
      return <Badge variant="success" dot>Registered</Badge>;
    case "rejected":
      return <Badge variant="danger" dot>Rejected</Badge>;
    case "closed":
      return <Badge variant="secondary" dot>Closed</Badge>;
    case "received":
      return <Badge variant="glass" dot>Received</Badge>;
    default:
      return <Badge variant="glass">{status}</Badge>;
  }
}
