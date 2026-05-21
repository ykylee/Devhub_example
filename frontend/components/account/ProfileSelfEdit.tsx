"use client";

import { useState } from "react";
import { Loader2, Save, AlertCircle, CheckCircle2 } from "lucide-react";
import { useStore } from "@/lib/store";
import { onboardingService } from "@/lib/services/onboarding.service";
import { ApiError } from "@/lib/services/api-client";
import { OrganizationPicker } from "@/components/onboarding/OrganizationPicker";

export function ProfileSelfEdit() {
  const { actor, setActor, addToast } = useStore();
  const [displayName, setDisplayName] = useState(actor?.display_name ?? "");
  const [unitId, setUnitId] = useState(actor?.primary_unit_id ?? "");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const initialDisplayName = actor?.display_name ?? "";
  const initialUnitId = actor?.primary_unit_id ?? "";
  const changed = displayName.trim() !== initialDisplayName.trim() || (unitId ?? "") !== (initialUnitId ?? "");
  const valid = displayName.trim().length >= 1 && displayName.trim().length <= 100;

  async function handleSave() {
    if (!actor || !changed || !valid || saving) return;
    setSaving(true);
    setError(null);
    setSuccess(null);
    try {
      const payload: { display_name?: string; primary_unit_id?: string } = {};
      if (displayName.trim() !== initialDisplayName.trim()) payload.display_name = displayName.trim();
      if ((unitId ?? "") !== (initialUnitId ?? "")) payload.primary_unit_id = unitId;
      const me = await onboardingService.patchMe(payload);
      setActor({
        ...actor,
        display_name: me.display_name,
        primary_unit_id: me.primary_unit_id ?? null,
        review_status: (me.review_status as "pending_review" | "reviewed") ?? actor.review_status,
      });
      const newStatus = me.review_status;
      if (newStatus === "pending_review" && actor.review_status === "reviewed") {
        setSuccess("프로필이 갱신되었습니다. 소속 변경으로 인해 관리자 검토가 다시 필요합니다.");
      } else {
        setSuccess("프로필이 갱신되었습니다.");
      }
      addToast("프로필을 저장했습니다.", "success");
    } catch (err) {
      if (err instanceof ApiError) {
        const code = (err.payload as { error?: { code?: string } } | undefined)?.error?.code;
        if (code === "unit_not_found") setError("선택한 조직을 찾을 수 없습니다.");
        else if (code === "invalid_payload") setError("입력값이 올바르지 않습니다.");
        else setError(err.message || "저장 실패");
      } else {
        setError(err instanceof Error ? err.message : "저장 실패");
      }
    } finally {
      setSaving(false);
    }
  }

  if (!actor) return null;
  if (actor.onboarding_required) {
    // skip 사용자는 AuthGuard 가 차단. pending_review 사용자는 actor.onboarding_required=true 이나
    // onboarding_completed_at 이 set 되어 있어 본 컴포넌트 표시는 안전 (PATCH /me 도 onboardingGate allowlist 외).
    if (!actor.onboarding_completed_at) return null;
  }

  return (
    <div className="glass border-border rounded-3xl p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-bold tracking-tight">프로필</h3>
          <p className="text-xs text-muted-foreground">이름과 소속 조직을 변경할 수 있습니다.</p>
        </div>
        {actor.review_status && (
          <span
            className={`text-[10px] font-bold uppercase tracking-widest px-2 py-1 rounded ${
              actor.review_status === "reviewed"
                ? "bg-success/20 text-success"
                : "bg-amber-500/20 text-amber-500"
            }`}
            data-testid="account-review-status"
          >
            {actor.review_status === "reviewed" ? "검토 완료" : "검토 대기"}
          </span>
        )}
      </div>

      <div className="space-y-2">
        <label className="text-xs font-bold uppercase tracking-widest text-muted-foreground" htmlFor="account-display-name">
          이름
        </label>
        <input
          id="account-display-name"
          type="text"
          value={displayName}
          onChange={(e) => setDisplayName(e.target.value)}
          maxLength={100}
          disabled={saving}
          className="w-full px-3 py-2 bg-card border border-border rounded text-sm focus:outline-none focus:ring-2 focus:ring-primary"
          data-testid="account-display-name"
        />
      </div>

      <div className="space-y-2">
        <label className="text-xs font-bold uppercase tracking-widest text-muted-foreground">
          소속 조직
        </label>
        <OrganizationPicker
          value={unitId}
          onChange={(id) => setUnitId(id)}
          disabled={saving}
          data-testid="account-unit-picker"
        />
        <p className="text-xs text-muted-foreground">
          소속을 변경하면 관리자 재검토가 필요합니다.
        </p>
      </div>

      {error && (
        <div className="flex items-start gap-2 bg-red-500/10 border border-red-500/30 rounded p-3 text-sm text-red-400" role="alert" data-testid="account-error">
          <AlertCircle className="w-4 h-4 mt-0.5 flex-shrink-0" />
          <span>{error}</span>
        </div>
      )}
      {success && (
        <div className="flex items-start gap-2 bg-success/10 border border-success/30 rounded p-3 text-sm text-success" role="status" data-testid="account-success">
          <CheckCircle2 className="w-4 h-4 mt-0.5 flex-shrink-0" />
          <span>{success}</span>
        </div>
      )}

      <div className="flex justify-end">
        <button
          type="button"
          onClick={handleSave}
          disabled={!changed || !valid || saving}
          className="px-4 py-2 bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50 rounded text-sm font-bold uppercase tracking-wider flex items-center gap-2"
          data-testid="account-save"
        >
          {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
          저장
        </button>
      </div>
    </div>
  );
}
