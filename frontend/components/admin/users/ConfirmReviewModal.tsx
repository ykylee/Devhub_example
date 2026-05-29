"use client";

import { useState } from "react";
import { Loader2, AlertCircle, CheckCircle2, X } from "lucide-react";
import { onboardingService } from "@/domain/onboarding/service/onboarding.service";
import type { OrgMember } from "@/domain/organization-management/service/identity.service";
import { ApiError } from "@/lib/services/api-client";

interface Props {
  member: OrgMember;
  onClose: () => void;
  onConfirmed: (userId: string, reviewedAt: string, reviewedBy: string) => void;
}

export function ConfirmReviewModal({ member, onClose, onConfirmed }: Props) {
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleConfirm() {
    if (submitting) return;
    setSubmitting(true);
    setError(null);
    try {
      const result = await onboardingService.confirmUserReview(member.id);
      onConfirmed(member.id, result.reviewed_at, result.reviewed_by);
      onClose();
    } catch (err) {
      if (err instanceof ApiError) {
        const code = (err.payload as { error?: { code?: string } } | undefined)?.error?.code;
        if (code === "user_not_found") setError("사용자를 찾을 수 없습니다.");
        else if (code === "review_already_confirmed") setError("이미 검토 완료된 사용자입니다.");
        else if (code === "onboarding_not_completed") setError("사용자가 아직 onboarding 을 완료하지 않았습니다.");
        else setError(err.message || "검토 확정 실패");
      } else {
        setError(err instanceof Error ? err.message : "검토 확정 실패");
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4" role="dialog" aria-modal="true" data-testid="confirm-review-modal">
      <div className="bg-background border border-border rounded-2xl max-w-md w-full shadow-2xl">
        <div className="flex items-center justify-between px-6 py-4 border-b border-border">
          <h2 className="text-lg font-bold tracking-tight">검토 확정</h2>
          <button type="button" onClick={onClose} className="p-1 hover:bg-card/50 rounded" aria-label="닫기">
            <X className="w-4 h-4" />
          </button>
        </div>

        <div className="px-6 py-5 space-y-4">
          <p className="text-sm text-muted-foreground">
            아래 사용자의 onboarding 정보를 검토하고 확정합니다. 확정 후에는 사용자가 정상 활성화됩니다.
          </p>

          <div className="bg-card/50 border border-border rounded p-4 space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-muted-foreground">사용자 ID</span>
              <span className="font-mono">{member.id}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">이름</span>
              <span className="font-medium">{member.name}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">이메일</span>
              <span className="font-mono text-xs">{member.email}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">소속</span>
              <span className="font-mono text-xs">{member.primary_dept_id || "(미지정)"}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">역할</span>
              <span className="font-bold">{member.role}</span>
            </div>
            {member.onboarding_completed_at && (
              <div className="flex justify-between">
                <span className="text-muted-foreground">제출일시</span>
                <span className="font-mono text-xs">{new Date(member.onboarding_completed_at).toLocaleString("ko-KR")}</span>
              </div>
            )}
          </div>

          {error && (
            <div className="flex items-start gap-2 bg-red-500/10 border border-red-500/30 rounded p-3 text-sm text-red-400" role="alert">
              <AlertCircle className="w-4 h-4 mt-0.5 flex-shrink-0" />
              <span>{error}</span>
            </div>
          )}
        </div>

        <div className="flex gap-3 px-6 py-4 border-t border-border">
          <button
            type="button"
            onClick={onClose}
            disabled={submitting}
            className="flex-1 px-4 py-2 bg-card hover:bg-card/70 border border-border rounded text-sm font-bold uppercase tracking-wider disabled:opacity-50"
          >
            취소
          </button>
          <button
            type="button"
            onClick={handleConfirm}
            disabled={submitting}
            className="flex-1 px-4 py-2 bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50 rounded text-sm font-bold uppercase tracking-wider flex items-center justify-center gap-2"
            data-testid="confirm-review-submit"
          >
            {submitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <CheckCircle2 className="w-4 h-4" />}
            확정
          </button>
        </div>
      </div>
    </div>
  );
}
