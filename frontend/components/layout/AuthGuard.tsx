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

// limited-mode (skip 단계) 사용자가 client-side 에서 차단되는 경로.
// TC-ONBOARD-SKIP-PROTECTED-01 (P0) 가 `/account` 진입 시 hard redirect 를 요구한다 —
// `/account` self-service 편집은 완료된 프로필을 전제로 한다.
// /admin 계열은 별도 `pathRequiresSystemAdmin` 가드로 차단되므로 본 목록에 포함하지 않는다.
const LIMITED_MODE_BLOCKED_PREFIXES = ["/account"] as const;

function pathBlockedInLimitedMode(path: string): boolean {
  return LIMITED_MODE_BLOCKED_PREFIXES.some((p) => path.startsWith(p));
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
        // REQ-FR-ONBOARD-010 3-branch gating:
        //   1) skip 액션 미실행 + 미완료 → /onboarding 강제 redirect (첫 진입)
        //   2) skip 액션 + 미완료 + 일반 페이지 → 정상 진입 + banner 노출
        //   3) skip 액션 + 미완료 + /account → /onboarding hard redirect (TC-ONBOARD-SKIP-PROTECTED-01)
        // 이전 구현은 whitelist (`["/onboarding", "/auth/"]`) 였는데 default landing
        // (/developer, /manager) 까지 막아 skip 직후 무한 redirect 루프가 발생했다.
        // blocklist 방식으로 전환 — /account 만 명시 차단하고 나머지는 banner 와 함께 통과.
        if (resolved.onboarding_required && pathname !== "/onboarding") {
          if (!isOnboardingSkipped() || pathBlockedInLimitedMode(pathname)) {
            router.replace("/onboarding");
            return;
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
