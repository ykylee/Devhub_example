"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { X, Inbox, ArrowRight, FolderKanban, Box, Loader2, UserCog } from "lucide-react";
import { DevRequest, DevRequestTargetType } from "@/lib/services/dev_request.types";
import { devRequestService } from "@/lib/services/dev_request.service";
import { cn } from "@/lib/utils";

interface DevRequestDetailModalProps {
  request: DevRequest;
  // system_admin 만 reassign / close 액션이 활성화된다. 페이지가 actor.role 을 전달.
  isSystemAdmin: boolean;
  onClose: () => void;
  onChanged: (updated: DevRequest) => void;
}

type Mode = "view" | "register" | "reject" | "reassign";

export function DevRequestDetailModal({
  request,
  isSystemAdmin,
  onClose,
  onChanged,
}: DevRequestDetailModalProps) {
  const [mode, setMode] = useState<Mode>("view");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [targetType, setTargetType] = useState<DevRequestTargetType>("application");
  const [targetID, setTargetID] = useState("");
  const [rejectedReason, setRejectedReason] = useState("");
  const [newAssignee, setNewAssignee] = useState(request.assignee_user_id);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onClose]);

  const canRegister = request.status === "pending" || request.status === "in_review";
  const canReject = canRegister; // 동일 상태에서 가능
  const canReassign = isSystemAdmin;

  const handleRegister = async () => {
    if (!targetID.trim()) {
      setError("target_id is required");
      return;
    }
    setError(null);
    setSubmitting(true);
    try {
      const updated = await devRequestService.register(request.id, {
        target_type: targetType,
        target_id: targetID.trim(),
      });
      onChanged(updated);
      onClose();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to register dev_request");
    } finally {
      setSubmitting(false);
    }
  };

  const handleReject = async () => {
    if (!rejectedReason.trim()) {
      setError("rejected_reason is required");
      return;
    }
    setError(null);
    setSubmitting(true);
    try {
      const updated = await devRequestService.reject(request.id, rejectedReason.trim());
      onChanged(updated);
      onClose();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to reject");
    } finally {
      setSubmitting(false);
    }
  };

  const handleReassign = async () => {
    if (!newAssignee.trim() || newAssignee === request.assignee_user_id) {
      setError("new assignee_user_id is required");
      return;
    }
    setError(null);
    setSubmitting(true);
    try {
      const updated = await devRequestService.reassign(request.id, newAssignee.trim());
      onChanged(updated);
      onClose();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to reassign");
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
        className="relative w-full max-w-2xl glass border-border rounded-3xl shadow-2xl overflow-hidden"
        role="dialog"
        aria-modal="true"
      >
        <div className="p-8 border-b border-border flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-orange-500/20 rounded-xl flex items-center justify-center">
              <Inbox className="w-5 h-5 text-orange-400" />
            </div>
            <div>
              <h2 className="text-xl font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">
                Dev <span className="text-orange-400">Request</span>
              </h2>
              <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
                {request.source_system} · {request.status}
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-2 hover:bg-muted/30 rounded-xl text-muted-foreground transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="p-8 space-y-6 max-h-[75vh] overflow-y-auto custom-scrollbar">
          <div className="space-y-2">
            <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Title</p>
            <p className="text-base font-bold text-foreground dark:text-primary-foreground">{request.title}</p>
          </div>

          {request.details && (
            <div className="space-y-2">
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Details</p>
              <pre className="text-xs text-foreground dark:text-primary-foreground whitespace-pre-wrap bg-muted/20 border border-border/40 rounded-2xl p-4">{request.details}</pre>
            </div>
          )}

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <Field label="Requester" value={request.requester} />
            <Field label="Assignee" value={request.assignee_user_id} />
            <Field label="External Ref" value={request.external_ref || "—"} mono />
            <Field label="Received" value={request.received_at} mono />
          </div>

          {request.status === "registered" && (
            <div className="p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-2xl text-[11px] text-emerald-400">
              Registered as <span className="font-bold">{request.registered_target_type}</span>{" "}
              <span className="font-mono">{request.registered_target_id}</span>
            </div>
          )}

          {request.status === "rejected" && request.rejected_reason && (
            <div className="p-4 bg-rose-500/10 border border-rose-500/20 rounded-2xl text-[11px] text-rose-400">
              Rejected — {request.rejected_reason}
            </div>
          )}

          {/* 액션 영역 */}
          {mode === "view" && (
            <div className="flex flex-wrap gap-3 pt-4 border-t border-border/60">
              {canRegister && (
                <>
                  <ActionButton
                    icon={<Box className="w-4 h-4" />}
                    label="Register as Application"
                    onClick={() => {
                      setMode("register");
                      setTargetType("application");
                    }}
                    color="purple"
                  />
                  <ActionButton
                    icon={<FolderKanban className="w-4 h-4" />}
                    label="Register as Project"
                    onClick={() => {
                      setMode("register");
                      setTargetType("project");
                    }}
                    color="indigo"
                  />
                </>
              )}
              {canReject && (
                <ActionButton
                  icon={<X className="w-4 h-4" />}
                  label="Reject"
                  onClick={() => setMode("reject")}
                  color="rose"
                />
              )}
              {canReassign && (
                <ActionButton
                  icon={<UserCog className="w-4 h-4" />}
                  label="Reassign"
                  onClick={() => setMode("reassign")}
                  color="amber"
                />
              )}
            </div>
          )}

          {mode === "register" && (
            <div className="space-y-4 pt-4 border-t border-border/60">
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">
                Register as {targetType}
              </p>
              <p className="text-[11px] text-muted-foreground">
                기존 {targetType} 의 ID 를 입력해 의뢰를 매핑합니다. 신규 {targetType} 생성과의 단일 트랜잭션은 후속 sprint 에서 도입.
              </p>
              <input
                value={targetID}
                onChange={(e) => setTargetID(e.target.value)}
                placeholder={targetType === "application" ? "application id (uuid)" : "project id (uuid)"}
                className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm font-mono text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-orange-400/50"
              />
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => setMode("view")}
                  className="flex-1 glass border-border text-foreground dark:text-primary-foreground font-bold py-3 rounded-2xl hover:bg-muted/30 transition-all uppercase tracking-widest text-[10px]"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={handleRegister}
                  disabled={submitting}
                  className="flex-1 bg-orange-600 text-white font-black py-3 rounded-2xl hover:scale-[1.02] active:scale-[0.98] transition-all uppercase tracking-widest text-[10px] disabled:opacity-50 flex items-center justify-center gap-2"
                >
                  {submitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <><ArrowRight className="w-4 h-4" /> Confirm</>}
                </button>
              </div>
            </div>
          )}

          {mode === "reject" && (
            <div className="space-y-4 pt-4 border-t border-border/60">
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">
                Reject reason (required)
              </p>
              <textarea
                value={rejectedReason}
                onChange={(e) => setRejectedReason(e.target.value)}
                rows={3}
                placeholder="중복 의뢰 / 범위 외 / 정보 부족 ..."
                className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-rose-400/50 resize-none"
              />
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => setMode("view")}
                  className="flex-1 glass border-border text-foreground dark:text-primary-foreground font-bold py-3 rounded-2xl hover:bg-muted/30 transition-all uppercase tracking-widest text-[10px]"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={handleReject}
                  disabled={submitting}
                  className="flex-1 bg-rose-600 text-white font-black py-3 rounded-2xl hover:scale-[1.02] active:scale-[0.98] transition-all uppercase tracking-widest text-[10px] disabled:opacity-50"
                >
                  {submitting ? <Loader2 className="w-4 h-4 animate-spin mx-auto" /> : "Reject"}
                </button>
              </div>
            </div>
          )}

          {mode === "reassign" && (
            <div className="space-y-4 pt-4 border-t border-border/60">
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">
                Reassign to (system_admin only)
              </p>
              <input
                value={newAssignee}
                onChange={(e) => setNewAssignee(e.target.value)}
                placeholder="new assignee user_id"
                className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-amber-400/50"
              />
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => setMode("view")}
                  className="flex-1 glass border-border text-foreground dark:text-primary-foreground font-bold py-3 rounded-2xl hover:bg-muted/30 transition-all uppercase tracking-widest text-[10px]"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={handleReassign}
                  disabled={submitting}
                  className="flex-1 bg-amber-600 text-white font-black py-3 rounded-2xl hover:scale-[1.02] active:scale-[0.98] transition-all uppercase tracking-widest text-[10px] disabled:opacity-50"
                >
                  {submitting ? <Loader2 className="w-4 h-4 animate-spin mx-auto" /> : "Reassign"}
                </button>
              </div>
            </div>
          )}

          {error && (
            <div className="p-4 bg-rose-500/10 border border-rose-500/20 rounded-2xl text-[11px] text-rose-400 font-medium">
              {error}
            </div>
          )}
        </div>
      </motion.div>
    </div>
  );
}

function Field({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="space-y-1">
      <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">{label}</p>
      <p className={cn("text-sm text-foreground dark:text-primary-foreground", mono && "font-mono text-xs")}>{value}</p>
    </div>
  );
}

function ActionButton({
  icon,
  label,
  onClick,
  color,
}: {
  icon: React.ReactNode;
  label: string;
  onClick: () => void;
  color: "purple" | "indigo" | "rose" | "amber";
}) {
  const palette = {
    purple: "bg-purple-500/10 border-purple-500/30 text-purple-400 hover:bg-purple-500/20",
    indigo: "bg-indigo-500/10 border-indigo-500/30 text-indigo-400 hover:bg-indigo-500/20",
    rose: "bg-rose-500/10 border-rose-500/30 text-rose-400 hover:bg-rose-500/20",
    amber: "bg-amber-500/10 border-amber-500/30 text-amber-400 hover:bg-amber-500/20",
  }[color];
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "flex items-center gap-2 px-4 py-2.5 border rounded-2xl text-[10px] font-black uppercase tracking-widest transition-all",
        palette,
      )}
    >
      {icon}
      {label}
    </button>
  );
}
