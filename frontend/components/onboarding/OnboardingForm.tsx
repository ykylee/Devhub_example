"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Loader2, AlertCircle, CheckCircle2 } from "lucide-react";
import { onboardingService } from "@/lib/services/onboarding.service";
import { useStore } from "@/lib/store";
import { OrganizationPicker } from "./OrganizationPicker";
import { ApiError } from "@/lib/services/api-client";
import { markOnboardingSkipped } from "@/lib/storage/onboardingSkip";
import { defaultLandingFor } from "@/lib/auth/role-routing";

interface Props {
  initialDisplayName?: string;
  initialUnitId?: string;
  email?: string;
  fromAdmin?: boolean;
}

export function OnboardingForm({ initialDisplayName = "", initialUnitId = "", email, fromAdmin }: Props) {
  const router = useRouter();
  const { actor, setActor, addToast } = useStore();
  const [displayName, setDisplayName] = useState(initialDisplayName);
  const [unitId, setUnitId] = useState(initialUnitId);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const valid = displayName.trim().length >= 1 && displayName.trim().length <= 100 && unitId.trim().length > 0;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!valid || submitting) return;
    setSubmitting(true);
    setError(null);
    try {
      const me = await onboardingService.submit({
        display_name: displayName.trim(),
        primary_unit_id: unitId.trim(),
      });
      if (actor) {
        setActor({
          ...actor,
          display_name: me.display_name,
          primary_unit_id: me.primary_unit_id ?? null,
          onboarding_required: false,
          onboarding_completed_at: me.onboarding_completed_at ?? null,
          review_status: (me.review_status as "pending_review" | "reviewed") ?? "pending_review",
        });
      }
      addToast("프로필 등록이 완료되었습니다. 관리자 검토 후 활성화됩니다.", "success");
      router.replace(defaultLandingFor(actor?.role ?? "Developer"));
    } catch (err) {
      if (err instanceof ApiError) {
        const code = (err.payload as { error?: { code?: string } } | undefined)?.error?.code;
        if (code === "unit_not_found") setError("선택한 조직을 찾을 수 없습니다.");
        else if (code === "invalid_payload") setError("입력값이 올바르지 않습니다.");
        else if (code === "onboarding_already_completed") setError("이미 등록이 완료되었습니다.");
        else setError(err.message || "제출 중 오류가 발생했습니다.");
      } else {
        setError(err instanceof Error ? err.message : "제출 중 오류가 발생했습니다.");
      }
    } finally {
      setSubmitting(false);
    }
  }

  function handleSkip() {
    markOnboardingSkipped();
    addToast("나중에 등록할 수 있습니다. 일부 기능이 제한됩니다.", "info");
    router.replace(defaultLandingFor(actor?.role ?? "Developer"));
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6" data-testid="onboarding-form">
      {fromAdmin && (
        <div className="flex items-start gap-3 bg-primary/10 border border-primary/30 rounded p-3 text-sm">
          <CheckCircle2 className="w-4 h-4 text-primary mt-0.5 flex-shrink-0" />
          <div>
            관리자가 사전 등록한 정보가 채워져 있습니다. 확인 후 제출해주세요.
          </div>
        </div>
      )}

      <div className="space-y-2">
        <label className="text-xs font-bold uppercase tracking-widest text-muted-foreground" htmlFor="onboarding-email">
          이메일
        </label>
        <input
          id="onboarding-email"
          type="email"
          value={email ?? ""}
          disabled
          className="w-full px-3 py-2 bg-card/40 border border-border rounded text-sm text-muted-foreground"
          data-testid="onboarding-email"
        />
        <p className="text-xs text-muted-foreground">Keycloak SSO 토큰에서 가져온 값으로 변경할 수 없습니다.</p>
      </div>

      <div className="space-y-2">
        <label className="text-xs font-bold uppercase tracking-widest text-muted-foreground" htmlFor="onboarding-display-name">
          이름 <span className="text-red-500">*</span>
        </label>
        <input
          id="onboarding-display-name"
          type="text"
          value={displayName}
          onChange={(e) => setDisplayName(e.target.value)}
          maxLength={100}
          required
          disabled={submitting}
          className="w-full px-3 py-2 bg-card border border-border rounded text-sm focus:outline-none focus:ring-2 focus:ring-primary"
          data-testid="onboarding-display-name"
        />
      </div>

      <div className="space-y-2">
        <label className="text-xs font-bold uppercase tracking-widest text-muted-foreground">
          소속 조직 <span className="text-red-500">*</span>
        </label>
        <OrganizationPicker
          value={unitId}
          onChange={(id) => setUnitId(id)}
          disabled={submitting}
          allowTree={false}
          data-testid="onboarding-unit-picker"
        />
      </div>

      {error && (
        <div className="flex items-start gap-2 bg-red-500/10 border border-red-500/30 rounded p-3 text-sm text-red-400" role="alert" data-testid="onboarding-error">
          <AlertCircle className="w-4 h-4 mt-0.5 flex-shrink-0" />
          <span>{error}</span>
        </div>
      )}

      <div className="flex flex-col-reverse sm:flex-row gap-3 pt-2">
        <button
          type="button"
          onClick={handleSkip}
          disabled={submitting}
          className="flex-1 px-4 py-2.5 bg-card hover:bg-card/70 border border-border rounded text-sm font-bold uppercase tracking-wider disabled:opacity-50"
          data-testid="onboarding-skip"
        >
          나중에 하기
        </button>
        <button
          type="submit"
          disabled={!valid || submitting}
          className="flex-1 px-4 py-2.5 bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50 rounded text-sm font-bold uppercase tracking-wider flex items-center justify-center gap-2"
          data-testid="onboarding-submit"
        >
          {submitting && <Loader2 className="w-4 h-4 animate-spin" />}
          제출하기
        </button>
      </div>
    </form>
  );
}
