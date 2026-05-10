"use client";

import { useEffect, useState } from "react";
import { useRouter, usePathname } from "next/navigation";
import { useStore } from "@/lib/store";
import { Loader2 } from "lucide-react";
import { websocketService, WsMessage } from "@/lib/services/websocket.service";
import { identityService, IdentityServiceError } from "@/lib/services/identity.service";

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
        });
        setIsAuthorized(true);
      } catch (err) {
        if (cancelled) return;
        clearActor();
        setIsAuthorized(false);
        if (err instanceof IdentityServiceError && err.status === 401) {
          router.replace("/login");
          return;
        }
        console.error("[AuthGuard] whoAmI failed", err);
        router.replace("/login");
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
      <div className="flex items-center justify-center h-screen bg-[#030014]">
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
