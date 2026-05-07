"use client";

import { useEffect, useState } from "react";
import { useRouter, usePathname } from "next/navigation";
import { useStore } from "@/lib/store";
import { Loader2 } from "lucide-react";
import { websocketService, WsMessage } from "@/lib/services/websocket.service";

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const { role, addToast, incrementNotifications } = useStore();
  const [isAuthorized, setIsAuthorized] = useState(false);

  useEffect(() => {
    // Basic Mock Auth Check
    // If there is no role selected, redirect to login
    if (!role) {
      router.replace("/login");
    } else {
      setIsAuthorized(true);
    }
  }, [role, router, pathname]);

  useEffect(() => {
    if (!isAuthorized) return;

    // Connect WebSocket when authorized
    websocketService.connect();

    // Global Event Handlers
    const handleNotification = (msg: WsMessage) => {
      incrementNotifications();
      addToast(msg.data?.message || "New Notification", "info");
    };

    const handleCriticalRisk = (msg: WsMessage) => {
      addToast(`CRITICAL: ${msg.data?.message || "Risk Detected"}`, "error");
    };

    websocketService.subscribe("notification.created", handleNotification);
    websocketService.subscribe("risk.critical.created", handleCriticalRisk);

    return () => {
      websocketService.unsubscribe("notification.created", handleNotification);
      websocketService.unsubscribe("risk.critical.created", handleCriticalRisk);
      websocketService.disconnect();
    };
  }, [isAuthorized, incrementNotifications, addToast]);

  if (!isAuthorized) {
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
