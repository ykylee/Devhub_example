"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { motion } from "framer-motion";
import { Loader2, UserPlus } from "lucide-react";
import { identityService, type ResolvedActor } from "@/domain/organization-management/service/identity.service";
import { ApiError } from "@/lib/services/api-client";
import { useStore } from "@/lib/store";
import { OnboardingForm } from "@/components/onboarding/OnboardingForm";
import { defaultLandingFor } from "@/lib/auth/role-routing";

export default function OnboardingPage() {
  const router = useRouter();
  const { setActor } = useStore();
  const [me, setMe] = useState<ResolvedActor | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const resolved = await identityService.whoAmI();
        if (cancelled) return;
        setActor({
          login: resolved.login,
          subject: resolved.subject,
          role: resolved.role,
          source: resolved.source,
          display_name: resolved.display_name,
          email: resolved.email,
          primary_unit_id: resolved.primary_unit_id ?? null,
          onboarding_required: resolved.onboarding_required,
          onboarding_completed_at: resolved.onboarding_completed_at ?? null,
          review_status: resolved.review_status ?? null,
        });
        if (!resolved.onboarding_required) {
          router.replace(defaultLandingFor(resolved.role));
          return;
        }
        setMe(resolved);
      } catch (err) {
        if (cancelled) return;
        if (err instanceof ApiError && err.status === 401) {
          router.replace("/login?error=session_expired");
          return;
        }
        setError(err instanceof Error ? err.message : "사용자 정보를 가져올 수 없습니다.");
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [router, setActor]);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-background">
        <div className="flex flex-col items-center gap-4">
          <Loader2 className="w-8 h-8 text-primary animate-spin" />
          <p className="text-xs font-bold text-muted-foreground uppercase tracking-widest">Loading…</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-background">
        <div className="bg-card border border-red-500/30 rounded p-6 max-w-md text-center space-y-2">
          <p className="text-red-400 font-bold">{error}</p>
          <button
            type="button"
            onClick={() => window.location.reload()}
            className="text-xs text-primary underline"
          >
            새로고침
          </button>
        </div>
      </div>
    );
  }

  if (!me) return null;

  const fromAdmin = !!me.display_name && !!me.primary_unit_id;

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-6">
      <motion.div
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        className="w-full max-w-xl space-y-6"
      >
        <div className="text-center space-y-3">
          <div className="w-14 h-14 mx-auto rounded-2xl bg-gradient-to-br from-primary to-accent flex items-center justify-center shadow-lg">
            <UserPlus className="w-7 h-7 text-primary-foreground" />
          </div>
          <h1 className="text-2xl font-black tracking-tighter uppercase">
            DevHub에 <span className="text-primary">환영합니다</span>
          </h1>
          <p className="text-sm text-muted-foreground">
            서비스 이용을 위해 프로필을 등록해주세요. 관리자 검토 후 활성화됩니다.
          </p>
        </div>

        <div className="glass border border-border rounded-3xl p-8">
          <OnboardingForm
            initialDisplayName={me.display_name ?? ""}
            initialUnitId={me.primary_unit_id ?? ""}
            email={me.email}
            fromAdmin={fromAdmin}
          />
        </div>

        <p className="text-center text-xs text-muted-foreground" data-testid="onboarding-login-hint">
          로그인 ID: <span className="font-mono">{me.login}</span>
        </p>
      </motion.div>
    </div>
  );
}
