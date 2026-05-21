"use client";

import { useMemo, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { CheckCircle2, Clock, UserCheck } from "lucide-react";
import type { OrgMember } from "@/lib/services/identity.service";
import { ConfirmReviewModal } from "./ConfirmReviewModal";

interface Props {
  members: OrgMember[];
  onReviewed: (userId: string) => void;
}

export function PendingReviewPanel({ members, onReviewed }: Props) {
  const [picked, setPicked] = useState<OrgMember | null>(null);

  const pending = useMemo(
    () => members.filter((m) => m.review_status === "pending_review"),
    [members],
  );

  if (pending.length === 0) {
    return (
      <div className="glass border-border rounded-2xl p-5 flex items-center gap-3" data-testid="pending-review-empty">
        <CheckCircle2 className="w-5 h-5 text-success" />
        <div>
          <p className="text-sm font-bold">검토 대기 사용자 없음</p>
          <p className="text-xs text-muted-foreground">모든 신규 사용자가 검토 완료된 상태입니다.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3">
        <div className="p-2 bg-amber-500/15 rounded-lg">
          <Clock className="w-4 h-4 text-amber-500" />
        </div>
        <div className="flex-1">
          <h3 className="text-sm font-black uppercase tracking-tight">검토 대기 ({pending.length})</h3>
          <p className="text-xs text-muted-foreground">Onboarding 을 제출한 신규 사용자가 검토를 기다리고 있습니다.</p>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-3" data-testid="pending-review-list">
        {pending.map((member) => (
          <motion.div
            key={member.id}
            initial={{ opacity: 0, y: 6 }}
            animate={{ opacity: 1, y: 0 }}
            className="glass border-border rounded-xl p-4 flex items-center gap-3"
            data-testid={`pending-review-row-${member.id}`}
          >
            <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-primary/20 to-accent/20 flex items-center justify-center border border-border">
              <span className="font-black">{member.name.charAt(0)}</span>
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-bold truncate">{member.name}</p>
              <p className="text-xs text-muted-foreground truncate">{member.email}</p>
              <p className="text-[10px] text-muted-foreground font-mono mt-0.5">{member.primary_dept_id || "(미지정)"}</p>
            </div>
            <button
              type="button"
              onClick={() => setPicked(member)}
              className="px-3 py-1.5 bg-primary text-primary-foreground rounded text-[10px] font-bold uppercase tracking-wider flex items-center gap-1.5"
              data-testid={`pending-review-confirm-${member.id}`}
            >
              <UserCheck className="w-3 h-3" />
              확정
            </button>
          </motion.div>
        ))}
      </div>

      <AnimatePresence>
        {picked && (
          <ConfirmReviewModal
            member={picked}
            onClose={() => setPicked(null)}
            onConfirmed={(userId) => onReviewed(userId)}
          />
        )}
      </AnimatePresence>
    </div>
  );
}
