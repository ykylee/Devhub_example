"use client";

import { useEffect, useState } from "react";
import { useRouter, usePathname } from "next/navigation";
import { useStore } from "@/lib/store";
import { Loader2 } from "lucide-react";
import { websocketService, WsMessage } from "@/lib/services/websocket.service";
import { identityService } from "@/lib/services/identity.service";
import { ApiError } from "@/lib/services/api-client";
import { defaultLandingFor, isSystemAdmin, pathRequiresSystemAdmin } from "@/lib/auth/role-routing";
import { isOnboardingSkipped } from "@/lib/storage/onboardingSkip";

// Routes that limited-mode (onboarding incomplete) users can still access
// after dismissing onboarding. Account self-service editing requires a
// completed profile, so /account is intentionally NOT in this list.
const LIMITED_MODE_ALLOWED_PREFIXES = ["/onboarding", "/auth/"] as const;

function pathAllowedInLimitedMode(path: string): boolean {
  return LIMITED_MODE_ALLOWED_PREFIXES.some((p) => path.startsWith(p));
}

type NotificationPayload = { message?: string };

function messageOf(msg: WsMessage): string | undefined {
  const data = msg.data as NotificationPayload | undefined;
  return data?.message;
}

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const { actor, setActor, clearActor, addToast, incrementNotifications } = useStore();
  const [isAuthorized, setIsAuthorized] = useState(false);

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
        // 3-branch gating (REQ-FR-ONBOARD-010):
        //   1) skip 액션 미실행 + 미완료 → /onboarding 강제 redirect
        //   2) skip 액션 + 미완료 → 정상 진입 + banner (단 /account 같은 보호 경로는 차단)
        //   3) 완료 → 정상 진입
        if (resolved.onboarding_required && pathname !== "/onboarding") {
          const skipped = isOnboardingSkipped();
          if (!skipped) {
            router.replace("/onboarding");
            return;
          }
          if (!pathAllowedInLimitedMode(pathname)) {
            const isPendingReview = resolved.onboarding_completed_at !== null;
            // pending_review 단계는 limited mode 보다 넓은 접근 (할당 리소스 query 가능).
            // skip 단계 (DB row 없음) 만 limited 보호 경로 차단.
            if (!isPendingReview) {
              router.replace("/onboarding");
              return;
            }
          }
        }
        // System routes (/admin, /admin/settings/*, /organization) must be
        // gated on actor.role — the source-of-truth for actual permissions —
        // not the zustand `role` field which Header's Switch View can
        // simulate. Non-admins get bounced to their default landing page.
        if (pathRequiresSystemAdmin(pathname) && !isSystemAdmin(resolved.role)) {
          router.replace(defaultLandingFor(resolved.role));
          return;
        }
        setIsAuthorized(true);
      } catch (err) {
        if (cancelled) return;
        clearActor();
        setIsAuthorized(false);
        if (err instanceof ApiError && err.status === 401) {
          router.replace("/auth/login");
          return;
        }
        console.error("[AuthGuard] whoAmI failed", err);
        router.replace("/auth/login");
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [pathname, router, setActor, clearActor]);

  useEffect(() => {
    if (!isAuthorized) return;

    websocketService.connect();

    const handleNotification = (msg: WsMessage) => {
      incrementNotifications();
      addToast(messageOf(msg) || "New Notification", "info");
    };

    const handleCriticalRisk = (msg: WsMessage) => {
      addToast(`CRITICAL: ${messageOf(msg) || "Risk Detected"}`, "error");
    };

    websocketService.subscribe("notification.created", handleNotification);
    websocketService.subscribe("risk.critical.created", handleCriticalRisk);

    return () => {
      websocketService.unsubscribe("notification.created", handleNotification);
      websocketService.unsubscribe("risk.critical.created", handleCriticalRisk);
      websocketService.disconnect();
    };
  }, [isAuthorized, incrementNotifications, addToast]);

  if (!isAuthorized || !actor) {
    return (
      <div className="flex items-center justify-center h-screen bg-background">
        <div className="flex flex-col items-center gap-4">
          <Loader2 className="w-8 h-8 text-primary animate-spin" />
          <p className="text-xs font-bold text-muted-foreground uppercase tracking-widest">
            Verifying Identity...
          </p>
        </div>
      </div>
    );
  }

  return <>{children}</>;
}
