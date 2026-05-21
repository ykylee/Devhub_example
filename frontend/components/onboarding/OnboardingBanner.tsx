"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { AlertTriangle, X, UserPlus } from "lucide-react";
import { useStore } from "@/lib/store";

// REQ-FR-ONBOARD-010 의 dismissible 배너. §16.2 spec 상 `onboarding_required: true`
// 는 항상 `onboarding_completed_at IS NULL` 이므로 배너는 limited-mode (skip 단계 또는
// admin pre-seed 후 첫 로그인 미제출) 만 노출. pending_review 단계는 `onboarding_required:
// false` 라 배너 자체가 안 보임 (검토 대기 상태 안내는 /account 의 ProfileSelfEdit 에서 별도 표시).
export function OnboardingBanner() {
  const router = useRouter();
  const { actor } = useStore();
  const [dismissed, setDismissed] = useState(false);

  if (!actor?.onboarding_required) return null;
  if (dismissed) return null;

  return (
    <div
      className="bg-primary/10 border-b border-primary/30 px-4 py-2 flex items-center gap-3 text-sm"
      data-testid="onboarding-banner"
    >
      {actor.primary_unit_id ? (
        <AlertTriangle className="w-4 h-4 text-amber-500 flex-shrink-0" />
      ) : (
        <UserPlus className="w-4 h-4 text-primary flex-shrink-0" />
      )}
      <span className="flex-1">
        프로필 등록이 완료되지 않았습니다. 일부 기능이 제한됩니다.
      </span>
      <button
        type="button"
        onClick={() => router.push("/onboarding")}
        className="px-3 py-1 bg-primary text-primary-foreground rounded text-xs font-bold uppercase tracking-wider"
        data-testid="onboarding-banner-resume"
      >
        지금 등록
      </button>
      <button
        type="button"
        onClick={() => setDismissed(true)}
        className="p-1 hover:bg-card/50 rounded"
        aria-label="배너 닫기"
        data-testid="onboarding-banner-dismiss"
      >
        <X className="w-3 h-3" />
      </button>
    </div>
  );
}
